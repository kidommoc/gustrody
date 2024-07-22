package models

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/utils"
	"github.com/redis/go-redis/v9"
)

var cacheExpires = time.Hour * 24 * 14

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

func (db *CacheDb) CacheQueryString(key []string) (kv map[string]string, err error) {
	logger := db.lg
	kv = make(map[string]string, len(key))
	result, err := db.client.MGet(defaultCtx, key...).Result()
	switch err {
	case redis.Nil:
		return nil, ErrNotFound
	case nil:
	default:
		msg := fmt.Sprintf("[Models.CacheQueryString] Cannot query strings: %v", key)
		logger.Error(msg, err)
		return nil, ErrDbInternal
	}
	for i, v := range result {
		s, ok := v.(string)
		if ok {
			kv[key[i]] = s
		}
	}
	return kv, nil
}

// nx: only set when not exist
func (db *CacheDb) CacheSetString(key string, value string, nx bool) error {
	logger := db.lg
	var err error
	if nx {
		_, err = db.client.SetNX(defaultCtx, key, value, cacheExpires).Result()
	} else {
		_, err = db.client.Set(defaultCtx, key, value, cacheExpires).Result()
	}
	if err != nil {
		msg := fmt.Sprintf(`[Models.CacheSetString] Cannot set string: "%s"`, key)
		logger.Warning(msg, err)
		return ErrDbInternal
	}
	return nil
}

func (db *CacheDb) CacheRemoveString(key string) error {
	logger := db.lg
	_, err := db.client.GetDel(defaultCtx, key).Result()
	if err != nil {
		msg := fmt.Sprintf(`[Models.CacheRemoveString] Cannot remove string: "%s"`, key)
		logger.Warning(msg, err)
		return ErrDbInternal
	}
	return nil
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
func (db *CacheDb) CacheUpdateJson(key string, value interface{}) error {
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

	for i := 0; i < redisMaxRetries; i += 1 {
		if err := db.client.Watch(defaultCtx, func(tx *redis.Tx) error {
			var om map[string]interface{}
			oldJson, err := tx.Get(defaultCtx, key).Result()
			switch err {
			case redis.Nil:
				logger.Warning("[Models.CacheUpdateJson] Query not found",
					"key", key, "error", err,
				)
				return ErrNotFound
			case nil:
			default:
				logger.Warning("[Models.CacheUpdateJson] Query failed.",
					"key", key, "error", err,
				)
				return ErrDbInternal
			}
			if err = json.Unmarshal([]byte(oldJson), &om); err != nil {
				return db.warnJsonUnmarshal("CacheUpdateJson", "oldJson", err)
			}

			if err = db.updateJson(&om, &nm); err != nil {
				logger.Warning("[Models.CacheUpdateJson] Json struct not match.",
					"key", key, "error", err,
				)
				return ErrFormat
			}
			s, _ := json.Marshal(om)
			tx.Set(defaultCtx, key, string(s), redis.KeepTTL)
			return nil
		}, key); err == redis.TxFailedErr {
			continue
		} else {
			return err
		}
	}
	return ErrMaxRetries
}

type PageItem struct {
	ID   string    `json:"id"`
	Date time.Time `json:"date"`
}

// n: indexed from 1. negative n indicates the n-th page counted from the tail.
func (db *CacheDb) CacheQueryPage(key string, n int) (page []PageItem, err error) {
	logger := db.lg
	s, err := db.client.LIndex(defaultCtx, key, int64(n)).Result()
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
func (db *CacheDb) CacheQueryPageByDate(key string, date time.Time, acse bool) (page []PageItem, err error) {
	logger := db.lg
	for i := 0; i < redisMaxRetries; i += 1 {
		if err := db.client.Watch(defaultCtx, func(tx *redis.Tx) error {
			// get index page
			s, err := tx.LIndex(defaultCtx, key, 0).Result()
			switch err {
			case nil:
			case redis.Nil:
				msg := fmt.Sprintf("[Models.CacheQueryPageByDate] Pages not found: %s", key)
				logger.Error(msg, err)
				return ErrDbInternal
			default:
				msg := fmt.Sprintf("[Models.CacheQueryPageByDate] Failed to query index page: %s", key)
				logger.Error(msg, err)
				return ErrDbInternal
			}

			var idx []time.Time
			if err := json.Unmarshal([]byte(s), &idx); err != nil {
				return db.warnJsonUnmarshal("CacheQueryPageByDate", "idx", err)
			}

			// get pages (1 or 2)
			pos := 0
			du := date.Unix()
			for ; pos < len(idx); pos += 1 {
				if acse && idx[pos].Unix() > du || !acse && idx[pos].Unix() < du {
					break
				}
			}
			ss, err := tx.LRange(defaultCtx, key, int64(pos), int64(pos+1)).Result()
			if err != nil {
				msg := fmt.Sprintf("[Models.CacheQueryPageByDate] Failed to query pages of %s: p%d to p%d", key, pos, pos+1)
				logger.Error(msg, err)
				return ErrDbInternal
			}
			if len(ss) == 0 {
				// cache corrupted. clear cache
				logger.Warning("[Models.CacheQueryPageByDate] Cache corrupted.", "key", key)
				tx.LTrim(defaultCtx, key, 1, 0)
				return ErrNotFound
			}

			var page1 []PageItem
			var page2 []PageItem
			if err := json.Unmarshal([]byte(ss[0]), &page1); err != nil {
				return db.warnJsonUnmarshal("CacheQueryPageByDate", "page1", err)
			}
			err = nil
			if len(ss) < 2 {
				err = ErrNoEnoughPages
			} else {
				if err := json.Unmarshal([]byte(ss[1]), &page2); err != nil {
					return db.warnJsonUnmarshal("CacheQueryPageByDate", "page2", err)
				}
			}

			// find where target page starts
			start := 0
			for ; start < len(page1); start += 1 {
				if acse && page1[start].Date.Unix() > du || !acse && page1[start].Date.Unix() < du {
					break
				}
			}
			page = page1[start:]
			if len(page) < redisPageSize/2 {
				page = append(page, page2...)
			}
			return err
		}, key); err == nil {
			return page, nil
		} else if err == redis.TxFailedErr {
			continue
		} else {
			return page, err
		}
	}
	return nil, ErrMaxRetries
}

/*
func (db *CacheDb) CacheRemoveItems(key string, ids []string) error {
	logger := db.lg
	for i := 0; i < redisMaxRetries; i += 1 {
		if err := db.client.Watch(defaultCtx, func(tx *redis.Tx) error {
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
						var after []PageItem
						if i < len(page) {
							after = page[i+1:]
						}
						page = append(page[0:i], after...)
						s, _ := json.Marshal(page)
						tx.LSet(defaultCtx, key, int64(idx), s)
						return nil
					}
				}
			}
			return nil
		}, key); err == nil {
			return nil
		} else if err == redis.TxFailedErr {
			continue
		} else {
			return err
		}
	}
	return ErrMaxRetries
}
*/

/*
func (db *CacheDb) CachePopPage(key string) (page []PageItem, err error) {
	logger := db.lg
	s, err := db.client.LPop(defaultCtx, key).Result()
	switch err {
	case redis.Nil:
		return nil, ErrNotFound
	case nil:
	default:
		msg := fmt.Sprintf(`[Models.CachePopPage] Cannot query "%s".`, key)
		logger.Error(msg, err)
		return nil, ErrDbInternal
	}
	err = json.Unmarshal([]byte(s), &page)
	if err != nil {
		msg := fmt.Sprintf("[Models.CachePopPage] Not page: %s", key)
		logger.Error(msg, err)
		return nil, ErrFormat
	}
	return page, nil
}
*/

// will not check whether items in pages are sorted acsending / decsending
func (db *CacheDb) CachePushPages(key string, pages [][]PageItem, acse bool) error {
	logger := db.lg
	if len(pages) == 0 || len(pages[0]) == 0 {
		return nil
	}

	for i := 0; i < redisMaxRetries; i += 1 {
		if err := db.client.Watch(defaultCtx, func(tx *redis.Tx) error {
			// get and set index page
			s, err := tx.LIndex(defaultCtx, key, 0).Result()
			switch err {
			case nil:
				var idx []time.Time
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
				s, _ := json.Marshal(idx)
				_, err = tx.LSet(defaultCtx, key, 0, s).Result()
				if err != nil {
					logger.Warning("[Models.CachePushPage] Failed to set index page.",
						"key", key, "error", err,
					)
					return err
				}
			case redis.Nil:
				// index page not found (key-value inexsitent yet)
				idx := make([]time.Time, 0, len(pages))
				for _, page := range pages {
					idx = append(idx, page[0].Date)
				}
				s, _ := json.Marshal(idx)
				_, err = tx.RPush(defaultCtx, key, s).Result()
				if err != nil {
					logger.Warning("[Models.CachePushPage] Failed to push index page.",
						"key", key, "error", err,
					)
					return err
				}
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
			_, err = tx.RPush(defaultCtx, key, ps...).Result()
			if err != nil {
				logger.Warning("[Models.CachePushPage] Failed to push page.",
					"key", key, "error", err,
				)
				return err
			}
			return nil
		}, key); err == nil {
			return nil
		} else if err == redis.TxFailedErr {
			continue
		} else {
			return err
		}
	}
	return ErrMaxRetries
}
