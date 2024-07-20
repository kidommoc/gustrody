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

func (db *CacheDb) CacheSetString(key string, value string) error {
	logger := db.lg
	_, err := db.client.Set(defaultCtx, key, value, cacheExpires).Result()
	if err != nil {
		msg := fmt.Sprintf(`[Models.CacheQueryString] Cannot set string: "%s":"%s"`, key, value)
		logger.Error(msg, err)
		return ErrDbInternal
	}
	return nil
}

func (db *CacheDb) warnJsonMarshal(loc, varname string, err error) error {
	msg := fmt.Sprintf(`[Models.%s] Failed to marshal "%s" to json.`, loc, varname)
	db.lg.Warning(msg, "error", err)
	return ErrSyntax
}

func (db *CacheDb) warnJsonUnmarshal(loc, varname string, err error) error {
	msg := fmt.Sprintf(`[Models.%s] Failed to unmarshal "%s" to json.`, loc, varname)
	db.lg.Warning(msg, "error", err)
	return ErrSyntax
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
			return ErrSyntax
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
				return ErrSyntax
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
				logger.Debug("")
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
//   - bool: old = new. Note that bool field MUST NOT be omitempty.
//   - slice: old = new. Note that slice field MUST NOT be omitempty and MUST NOT be nil.
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
			return ErrSyntax
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
		return ErrSyntax
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
				return ErrSyntax
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

// n: indexed from 1
func (db *CacheDb) CacheQueryPage(key string, n int) (page []PageItem, err error) {
	logger := db.lg
	s, err := db.client.LIndex(defaultCtx, key, int64(n-1)).Result()
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
		return nil, ErrSyntax
	}
	return page, nil
}

func (db *CacheDb) CacheRemoveItem(key string, id string) error {
	logger := db.lg
	for i := 0; i < redisMaxRetries; i += 1 {
		if err := db.client.Watch(defaultCtx, func(tx *redis.Tx) error {
			ss, err := tx.LRange(defaultCtx, key, 0, -1).Result()
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
					return ErrSyntax
				}
				for i, item := range page {
					if item.ID == id {
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
		}, key); err == redis.TxFailedErr {
			continue
		} else {
			return err
		}
	}
	return ErrMaxRetries
}

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
		return nil, ErrSyntax
	}
	return page, nil
}

func (db *CacheDb) CachePushPage(key string, page []PageItem) error {
	logger := db.lg
	s, _ := json.Marshal(page)
	_, err := db.client.RPush(defaultCtx, key, s).Result()
	if err != nil {
		logger.Warning("[Models.CachePushPage] Failed to push page.",
			"key", key, "error", err,
		)
		return err
	}
	return nil
}
