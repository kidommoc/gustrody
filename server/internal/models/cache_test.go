package models

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/kidommoc/gustrody/internal/test"
)

func TestCacheString(t *testing.T) {
	logger := test.NewMockingLogger(t)
	client := initRedis(modelscfg, logger, redisOpt{
		Addr:    "localhost:6738",
		Db:      0,
		MaxConn: 10,
	})
	cacheDb := &CacheDb{logger, client}

	table := map[string]string{
		"foo": "hello",
		"bar": "world",
	}

	t.Cleanup(func() {
		for k := range table {
			client.GetDel(defaultCtx, k)
		}
	})

	t.Run("Test Set", func(t *testing.T) {
		for k, v := range table {
			err := cacheDb.CacheSetString(k, v)
			test.AssertNoError(t, err, "when set string: %s")
			exp, err := client.ExpireTime(defaultCtx, k).Result()
			test.AssertNoError(t, err, "when get expire time: %s")
			if exp < cacheExpires-time.Minute {
				t.Fatalf("wrong expire time: %s", exp)
			}
		}
	})

	t.Run("Test Query", func(t *testing.T) {
		input := []string{}
		for k := range table {
			input = append(input, k)
		}
		got, err := cacheDb.CacheQueryString(input)
		test.AssertNoError(t, err)
		for k, want := range table {
			test.AssertEqual(t, want, got[k])
		}
	})
}

func TestCachePage(t *testing.T) {
	logger := test.NewMockingLogger(t)
	client := initRedis(modelscfg, logger, redisOpt{
		Addr:    "localhost:6738",
		Db:      0,
		MaxConn: 10,
	})
	cacheDb := &CacheDb{logger, client}

	key := "pages"
	d := time.Now().UTC()
	table := [][]PageItem{
		{PageItem{"1-1", d}, PageItem{"1-2", d}, PageItem{"1-3", d}},
		{PageItem{"2-1", d}, PageItem{"2-2", d}, PageItem{"2-3", d}},
		{PageItem{"3-1", d}, PageItem{"3-2", d}, PageItem{"3-3", d}},
	}

	t.Cleanup(func() { client.LTrim(defaultCtx, key, 1, 0) })

	t.Run("Test push page", func(t *testing.T) {
		want := make([]string, 0, len(table))
		for _, v := range table {
			err := cacheDb.CachePushPage(key, v)
			test.AssertNoError(t, err, "when push page: %s")
			s, _ := json.Marshal(v)
			want = append(want, string(s))
		}
		got, _ := client.LRange(defaultCtx, key, 0, -1).Result()
		test.AssertEqual(t, want, got)
	})

	t.Run("Test query page", func(t *testing.T) {
		for i, v := range table {
			got, err := cacheDb.CacheQueryPageByNum(key, i)
			test.AssertNoError(t, err, fmt.Sprintf("when query page %d: ", i)+"%s")
			test.AssertEqual(t, v, got)
		}
	})

	t.Run("Test pop page", func(t *testing.T) {
		got, err := cacheDb.CachePopPage(key)
		test.AssertNoError(t, err, "when pop page 1: %s")
		want := table[0]
		test.AssertEqual(t, want, got)
	})

	t.Run("Test remove item", func(t *testing.T) {
		err := cacheDb.CacheRemoveItem(key, "2-1")
		test.AssertNoError(t, err, "when remove item 2-1: %s")
		got, err := cacheDb.CachePopPage(key)
		test.AssertNoError(t, err, "when pop page 2: %s")
		want := table[1][1:]
		test.AssertEqual(t, want, got)

		// remove non-existent item
		err = cacheDb.CacheRemoveItem(key, "3-4")
		test.AssertNoError(t, err, "when remove item 3-4: %s")
		got, err = cacheDb.CachePopPage(key)
		test.AssertNoError(t, err, "when pop page 3: %s")
		want = table[2]
		test.AssertEqual(t, want, got)
	})
}
