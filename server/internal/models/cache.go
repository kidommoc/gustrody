package models

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"time"

	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/utils"
	"github.com/redis/go-redis/v9"
)

const cachePageSize = 20
const cacheExpires = time.Hour * 24 * 14
const cacheExpiresShort = time.Hour * 24

type ICacheDb interface {
}

type CacheDb struct {
	lg     logging.Logger
	client *redis.Client
}

var cacheIns *CacheDb = nil

func CacheInstance(lg logging.Logger, cl *redis.Client) *CacheDb {
	if cacheIns == nil {
		cacheIns = &CacheDb{
			lg:     lg,
			client: cl,
		}
	}
	return cacheIns
}

func (db *CacheDb) warnRedisInternal(loc, action string, err error, attach ...any) error {
	msg := fmt.Sprintf(`[Models.%s] Redis error when %s.`, loc, action)
	db.lg.Warning(msg, "error", err, attach)
	return ErrDbInternal
}

func (db *CacheDb) warnJsonMarshal(loc, varname string, err error) error {
	msg := fmt.Sprintf(`[Models.%s] Failed to marshal "%s" to json.`, loc, varname)
	db.lg.Warning(msg, "error", err)
	return ErrFormat
}

func (db *CacheDb) warnJsonUnmarshal(loc, varname string, err error) error {
	msg := fmt.Sprintf(`[Models.%s] Failed to unmarshal "%s" to json.`, loc, varname)
	db.lg.Warning(msg, "error", err)
	return ErrFormat
}

func (db *CacheDb) QueryString(key []string) (kv map[string]string, err error) {
	kv = make(map[string]string, len(key))
	result, err := db.client.MGet(defaultCtx, key...).Result()
	if err != nil {
		return nil, db.warnRedisInternal("CacheQueryString", "get strings", err, "keys", key)
	}
	for i, v := range result {
		if v == nil {
			continue
		}
		s, ok := v.(string)
		if ok {
			kv[key[i]] = s
		}
	}
	return kv, nil
}

// nx: only set when not exist
func (db *CacheDb) SetString(key string, value string, nx bool) error {
	var err error
	if nx {
		_, err = db.client.SetNX(defaultCtx, key, value, cacheExpires).Result()
	} else {
		_, err = db.client.Set(defaultCtx, key, value, cacheExpires).Result()
	}
	if err != nil {
		return db.warnRedisInternal("CacheSetString", "set string", err, "key", key)
	}
	return nil
}

func (db *CacheDb) RemoveString(key string) error {
	_, err := db.client.GetDel(defaultCtx, key).Result()
	if err != nil {
		return db.warnRedisInternal("CacheRemoveString", "remove string", err, "key", key)
	}
	return nil
}

// use a string of 16 chars to mark a field to clear.
// empty string will be omitted
var stringClearFlag = utils.GenerateRamdonHexString(16)

func (db *CacheDb) updateJson(om, nm *map[string]interface{}) error {
	logger := db.lg
	for k, o := range *om {
		n := (*nm)[k]
		if n == nil {
			continue
		}
		t := reflect.TypeOf(o)
		nt := reflect.TypeOf(n)
		if !nt.ConvertibleTo(t) {
			msg := fmt.Sprintf("[Models.updateJson] %s: %s is not convertable to %s", k, nt.Name(), t.Name())
			logger.Warning(msg)
			return ErrFormat
		}
		switch t.Kind() {
		case reflect.Slice:
			(*om)[k] = n
		case reflect.Struct:
			s, err := json.Marshal(o)
			if err != nil {
				return db.warnJsonMarshal("updateJson", "om."+k, err)
			}
			var oo map[string]interface{}
			json.Unmarshal(s, &oo)

			s, err = json.Marshal(n)
			if err != nil {
				return db.warnJsonMarshal("updateJson", "nm."+k, err)
			}
			var nn map[string]interface{}
			json.Unmarshal(s, &nn)
			if err := db.updateJson(&oo, &nn); err != nil {
				return err
			}
			(*om)[k] = oo
		case reflect.Map:
			if t.Key().Kind() != reflect.String {
				logger.Warning(`[Models.CacheUpdateJson] Unsupported map as "om.%s"`, k)
				return ErrFormat
			}
			oo, _ := o.(map[string]interface{})
			nn, _ := n.(map[string]interface{})
			if err := db.updateJson(&oo, &nn); err != nil {
				return err
			}
			(*om)[k] = oo
		default:
			switch o.(type) {
			case string:
				nn := reflect.ValueOf(n).String()
				if nn == stringClearFlag {
					(*om)[k] = ""
				} else {
					(*om)[k] = nn
				}
			case float64:
				oo, _ := o.(float64)
				nn := reflect.ValueOf(n).Convert(t).Float()
				(*om)[k] = oo + nn
			case bool:
				(*om)[k] = n
			}
		}
	}
	return nil
}

// value must be a map[string] or a struct. update rule:
//
//   - number: old + new
//   - bool: old = new. Bool field MUST NOT be omitempty, otherwise false value will be omitted.
//   - slice: old = new. Slice field MUST NOT be omitempty (empty slice will be omitted) and MUST NOT be nil.
//   - string: if new == stringClearFlag old = "" else old = new
//   - struct: recursively update. As struct filed, it could be a pointer. But as map entry it MUST NOT be a pointer.
func (db *CacheDb) UpdateJson(key string, value interface{}) error {
	logger := db.lg
	var nm map[string]interface{}
	t := reflect.TypeOf(value)
	switch t.Kind() {
	case reflect.Map:
		n, ok := value.(map[string]interface{})
		if !ok {
			logger.Warning(`[Models.CacheUpdateJson] Unsupported map as "value": %s`, reflect.TypeOf(value).Name())
			return ErrFormat
		}
		nm = n
	case reflect.Struct:
		s, err := json.Marshal(value)
		if err != nil {
			return db.warnJsonMarshal("CacheUpdateJson", `"value"`, err)
		}
		json.Unmarshal(s, &nm)
	default:
		logger.Warning(`[Models.CacheUpdateJson] Unsupported type as "value": %s`, reflect.TypeOf(value).Name())
		return ErrFormat
	}

	if err := db.client.Watch(defaultCtx, func(tx *redis.Tx) error {
		var om map[string]interface{}
		oldJson, err := tx.Get(defaultCtx, key).Result()
		switch err {
		case redis.Nil:
			logger.Warning("[Models.CacheUpdateJson] Query not found",
				"error", err, "key", key,
			)
			return ErrNotFound
		case nil:
		default:
			return db.warnRedisInternal("CacheUpdateJson", "get json string", err, "key", key)
		}
		if err = json.Unmarshal([]byte(oldJson), &om); err != nil {
			return db.warnJsonUnmarshal("CacheUpdateJson", "oldJson", err)
		}

		if err = db.updateJson(&om, &nm); err != nil {
			logger.Warning("[Models.CacheUpdateJson] Json struct not match.",
				"error", err, "key", key,
			)
			return ErrFormat
		}
		s, _ := json.Marshal(om)
		_, err = tx.Pipelined(defaultCtx, func(p redis.Pipeliner) error {
			p.Set(defaultCtx, key, string(s), redis.KeepTTL)
			return nil
		})
		return err
	}, key); err == redis.TxFailedErr {
		// corrupted
		db.RemoveString(key)
		return nil
	} else {
		return err
	}
}

type PageItem struct {
	ID   string    `json:"i"`
	Date time.Time `json:"d"`
}

func (db *CacheDb) clearPage(key string) {
	if _, err := db.client.LTrim(defaultCtx, key, 1, 0).Result(); err != nil {
		db.warnRedisInternal("clearPage", "clear page", err, "key", key)
	}
}

// n: indexed from 1. negative n indicates the n-th page counted from the tail.
func (db *CacheDb) QueryPage(tx *redis.Tx, key string, n int) (page []PageItem, err error) {
	var r redis.Cmdable = tx
	if tx == nil {
		r = db.client
	}
	logger := db.lg
	s, err := r.LIndex(defaultCtx, key, int64(n)).Result()
	switch err {
	case redis.Nil:
		return nil, ErrNotFound
	case nil:
	default:
		msg := fmt.Sprintf(`[Models.CacheQueryPage] Cannot query "%s"[%d].`, key, n)
		logger.Error(msg, err)
		return nil, ErrDbInternal
	}
	err = json.Unmarshal([]byte(s), &page)
	if err != nil {
		msg := fmt.Sprintf("[Models.CachePopPage] Not page: %s", key)
		logger.Error(msg, nil)
		return nil, ErrFormat
	}
	return page, nil
}

// acse == true: assume that items are ordered by date ascendingly and query a page by min date (not included)
//
// acse == false: assume that items are ordered by date decsendingly and query a page by max date (not included)
//
// Errors:
//
//   - ErrNoEnoughPages: the returned page is the last page in cache.
//   - ErrNotFound: when page is not nil, there are 2 conditions: 1) page contains only "first", target date is before first; 2) page contains only "last": target date is after last.
func (db *CacheDb) QueryPageByDate(key string, date time.Time, acse bool) (page []PageItem, err error) {
	ss, err := db.client.LRange(defaultCtx, key, 0, -1).Result()
	if len(ss) < 2 {
		return nil, ErrNotFound
	}

	// get index page
	var idx []time.Time
	if err := json.Unmarshal([]byte(ss[0]), &idx); err != nil {
		db.warnJsonUnmarshal("CacheQueryPageByDate", "idx", err)
		return nil, ErrNotFound
	}
	if len(idx) == 0 {
		// corrupted
		db.clearPage(key)
		return nil, ErrNotFound
	}
	du := date.Unix()

	// find in index page
	if acse && idx[0].Unix() > du || !acse && idx[0].Unix() < du {
		return nil, ErrNotFound
	}
	pos := 1
	for ; pos < len(idx); pos += 1 {
		if acse && idx[pos].Unix() > du || !acse && idx[pos].Unix() < du {
			break
		}
	}
	if pos == len(ss)-1 {
		ss = ss[pos : pos+1]
	} else if pos < len(ss) {
		ss = ss[pos : pos+2]
	} else {
		// corrupted
		db.clearPage(key)
		return nil, ErrNotFound
	}

	// get pages (1 or 2)
	var page1 []PageItem
	var page2 []PageItem
	if err := json.Unmarshal([]byte(ss[0]), &page1); err != nil {
		db.warnJsonUnmarshal("CacheQueryPageByDate", "page1", err)
		return nil, ErrNotFound
	}
	last := page1[len(page1)-1:]
	if len(ss) >= 2 {
		if err := json.Unmarshal([]byte(ss[1]), &page2); err != nil {
			db.warnJsonUnmarshal("CacheQueryPageByDate", "page2", err)
			return nil, ErrNotFound
		}
		last = page2[len(page2)-1:]
	}
	if last[0].ID == ":end" {
		last = nil
	} else {
		last = []PageItem{{ID: ":last", Date: last[0].Date}}
	}

	// find where is the first item
	start := 0
	for ; start < len(page1); start += 1 {
		if acse && page1[start].Date.Unix() > du || !acse && page1[start].Date.Unix() < du {
			break
		}
	}
	page = page1[start:]
	var e error = nil
	if len(page) < cachePageSize/2 {
		page = append(page, page2...)
	}
	if pos+1 >= len(idx) {
		e = ErrNoEnoughPages
		page = append(page, last...)
	}
	if len(page) == 1 && page[0].ID == ":last" {
		e = ErrNotFound
	}
	return page, e
}

func (db *CacheDb) RemoveItems(tx *redis.Tx, key string, ids []string) error {
	logger := db.lg
	f := func(tx *redis.Tx) error {
		ss, err := tx.LRange(defaultCtx, key, 1, -1).Result()
		if err != nil {
			logger.Warning("[Models.CachePopPage] Cannot query.",
				"key", key, "error", err,
			)
			return err
		}
		if len(ss) == 0 {
			logger.Warning("[Models.CachePopPage] Pages inexist.", "key", key)
			return ErrNotFound
		}
		for idx, s := range ss {
			var page []PageItem
			err = json.Unmarshal([]byte(s), &page)
			if err != nil {
				logger.Warning("[Models.CachePopPage] Not a page.",
					"key", key, "error", err,
				)
				return ErrFormat
			}
			for i, item := range page {
				if slices.Contains(ids, item.ID) {
					page = slices.Delete(page, i, i+1)

					if len(page) == 0 {
					}

					s, _ := json.Marshal(page)
					tx.LSet(defaultCtx, key, int64(idx), s)
				}
			}
			return nil
		}
		return nil
	}

	if tx != nil {
		return f(tx)
	} else {
		for i := 0; i < redisMaxRetries; i += 1 {
			if err := db.client.Watch(defaultCtx, f, key); err == redis.TxFailedErr {
				continue
			} else {
				return err
			}
		}
		return ErrMaxRetries
	}
}

// will not check whether items in pages are sorted acsending / decsending
func (db *CacheDb) PushPages(tx *redis.Tx, key string, pages [][]PageItem, acse bool) error {
	logger := db.lg
	if len(pages) == 0 || len(pages[0]) == 0 {
		return nil
	}

	f := func(tx *redis.Tx) error {
		// get and set index page
		s, err := tx.LIndex(defaultCtx, key, 0).Result()
		var idx []time.Time
		var exists bool
		switch err {
		case nil:
			if err = json.Unmarshal([]byte(s), &idx); err != nil {
				return db.warnJsonUnmarshal("CachePushPage", "idx", err)
			}

			// check index page to ensure acsending or desending
			du := pages[0][0].Date.Unix()
			if acse && idx[len(idx)-1].Unix() >= du || !acse && idx[len(idx)-1].Unix() <= du {
				logger.Warning("[CachePushPage] Order inconsistent when push pages.", "key", key)
				return ErrInconsistent
			}

			for _, page := range pages {
				idx = append(idx, page[0].Date)
			}
			exists = true
		case redis.Nil:
			// index page not found (key-value inexsitent yet)
			idx = make([]time.Time, 0, len(pages))
			for _, page := range pages {
				idx = append(idx, page[0].Date)
			}
			exists = false
		default:
			logger.Warning("[Models.CachePushPage] Failed to query index page.",
				"key", key, "error", err,
			)
			return err
		}

		ps := make([]interface{}, 0, len(pages))
		for _, page := range pages {
			if len(page) == 0 {
				continue
			}
			s, _ := json.Marshal(page)
			ps = append(ps, string(s))
		}
		_, err = tx.TxPipelined(defaultCtx, func(p redis.Pipeliner) error {
			s, _ := json.Marshal(idx)
			if exists {
				err = tx.LSet(defaultCtx, key, 0, s).Err()
			} else {
				err = tx.RPush(defaultCtx, key, s).Err()
			}
			if err != nil {
				db.warnRedisInternal("CachePushPage", "set index page", err, "key", key)
				return err
			}

			if err := tx.RPush(defaultCtx, key, ps...).Err(); err != nil {
				db.warnRedisInternal("CachePushPage", "push page", err, "key", key)
				return err
			}
			if err := tx.ExpireGT(defaultCtx, key, cacheExpiresShort).Err(); err != nil {
				db.warnRedisInternal("CachePushPage", "set pages expires", err, "key", key)
				return err
			}
			return nil
		})
		return err
	}

	if tx != nil {
		return f(tx)
	} else {
		for i := 0; i < redisMaxRetries; i += 1 {
			if err := db.client.Watch(defaultCtx, f, key); err == redis.TxFailedErr {
				continue
			} else {
				return err
			}
		}
		return ErrMaxRetries
	}
}
