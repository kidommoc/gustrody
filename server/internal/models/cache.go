package models

import (
	"encoding/json"
	"time"

	"github.com/kidommoc/gustrody/internal/logging"
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
	kv = make(map[string]string, len(key))
	result, err := db.client.MGet(defaultCtx, key...).Result()
	if err != nil {
		// handle error
		return nil, err
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
	_, err := db.client.Set(defaultCtx, key, value, cacheExpires).Result()
	if err != nil {
		// handle error
		return err
	}
	return nil
}

type PageItem struct {
	ID   string    `json:"id"`
	Date time.Time `json:"date"`
}

func (db *CacheDb) CacheQueryPageByNum(key string, n int) (page []PageItem, err error) {
	s, err := db.client.LIndex(defaultCtx, key, int64(n)).Result()
	if err != nil {
		// handle error
		return nil, err
	}
	err = json.Unmarshal([]byte(s), &page)
	if err != nil {
		// handle error
		return nil, err
	}
	return page, nil
}

func (db *CacheDb) CacheRemoveItem(key string, id string) (e error) {
	db.client.Watch(defaultCtx, func(tx *redis.Tx) error {
		ss, err := tx.LRange(defaultCtx, key, 0, -1).Result()
		if err != nil {
			// handle error
			e = err
			return err
		}
		for idx, s := range ss {
			var page []PageItem
			err = json.Unmarshal([]byte(s), &page)
			if err != nil {
				// handle error
				e = err
				return err
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
	}, key)
	return
}

func (db *CacheDb) CachePopPage(key string) (page []PageItem, err error) {
	s, err := db.client.LPop(defaultCtx, key).Result()
	if err != nil {
		// handle error
		return nil, err
	}
	err = json.Unmarshal([]byte(s), &page)
	if err != nil {
		// handle error
		return nil, err
	}
	return page, nil
}

func (db *CacheDb) CachePushPage(key string, page []PageItem) error {
	s, err := json.Marshal(page)
	if err != nil {
		// handle error
		return err
	}
	_, err = db.client.RPush(defaultCtx, key, s).Result()
	if err != nil {
		// handle error
		return err
	}
	return nil
}
