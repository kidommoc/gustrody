package models

import (
	"encoding/json"
	"fmt"
	"strings"
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
			err := cacheDb.SetString(k, v, false)
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
		got, err := cacheDb.QueryString(input)
		test.AssertNoError(t, err)
		for k, want := range table {
			test.AssertEqual(t, want, got[k])
		}
	})
}

func TestCacheJson(t *testing.T) {
	logger := test.NewMockingLogger(t)
	client := initRedis(modelscfg, logger, redisOpt{
		Addr:    "localhost:6738",
		Db:      0,
		MaxConn: 10,
	})
	cacheDb := &CacheDb{logger, client}

	type jsonS struct {
		B bool    `json:"b"`
		S string  `json:"s,omitempty"`
		F float64 `json:"f,omitempty"`
		A []int   `json:"a"`
		J *jsonS  `json:"j,omitempty"`
	}

	type ti struct {
		old   map[string]interface{}
		input interface{}
		want  map[string]interface{}
	}

	table := map[string]ti{
		"update1": {
			old:   map[string]interface{}{"b": true, "s": "abcd", "f": 10.0, "a": []int{1, 2}, "j": jsonS{F: 20.0, A: []int{}, B: false}},
			input: map[string]interface{}{"s": stringClearFlag, "f": 1, "a": []int{1, 2, 3}, "j": jsonS{F: -1, A: []int{}, B: false}},
			want:  map[string]interface{}{"b": true, "s": "", "f": 11.0, "a": []int{1, 2, 3}, "j": map[string]interface{}{"f": 19.0, "a": []interface{}{}, "b": false}},
		},
		"cache1": {
			old:   map[string]interface{}{"j": jsonS{B: true, S: "abcd", F: 10.0, A: []int{7, 11}, J: &jsonS{S: "efgh", A: []int{}, B: false}}},
			input: jsonS{B: false, S: "foo", F: 11, A: []int{}, J: &jsonS{S: "bar", A: []int{}, B: false}},
			want:  map[string]interface{}{"b": false, "s": "foo", "f": 21.0, "a": []interface{}{}, "j": map[string]interface{}{"s": "bar", "a": []interface{}{}, "b": false}},
			// note that json.Unmarshal will convert [] to []interface{}
		},
		"cache2": {
			old:   map[string]interface{}{"j": jsonS{B: true, A: []int{}}},
			input: jsonS{B: true, S: "foo", F: 11, A: []int{}, J: &jsonS{S: "bar", A: []int{}}},
			want:  map[string]interface{}{"b": true, "a": []interface{}{}},
		},
	}

	t.Cleanup(func() {
		for k := range table {
			client.GetDel(defaultCtx, k)
		}
	})

	t.Run("Test update json", func(t *testing.T) {
		for k, v := range table {
			if strings.Contains(k, "update") {
				got := v.old
				input, _ := v.input.(map[string]interface{})
				err := cacheDb.updateJson(&got, &input)
				test.AssertNoError(t, err)
				test.AssertEqual(t, v.want, got)
			}
		}
	})

	t.Run("Test update json cache", func(t *testing.T) {
		for k, v := range table {
			if strings.Contains(k, "cache") {
				s, _ := json.Marshal(v.old["j"])
				err := cacheDb.SetString(k, string(s), false)
				test.AssertNoError(t, err, "when set string: %s")
				err = cacheDb.UpdateJson(k, v.input)
				test.AssertNoError(t, err, "when update json: %s")
				ss, err := cacheDb.QueryString([]string{k})
				test.AssertNoError(t, err, "when query string: %s")
				var got map[string]interface{}
				err = json.Unmarshal([]byte(ss[k]), &got)
				test.AssertNoError(t, err, "when unmarshal got string: %s")
				test.AssertEqual(t, v.want, got)
			}
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

	d := time.Now().UTC()
	acseKey := "acse"
	acseTable := [][]PageItem{}
	for i := 0; i < 3; i++ {
		page := make([]PageItem, 0, cachePageSize)
		for j := 0; j < cachePageSize; j++ {
			page = append(page, PageItem{
				ID:   fmt.Sprintf("%d-%d", i+1, j+1),
				Date: d.Add(time.Duration(i*cachePageSize+j) * time.Minute),
			})
		}
		acseTable = append(acseTable, page)
	}
	decsKey := "decs"
	decsTable := [][]PageItem{}
	for i := 0; i < 3; i++ {
		page := make([]PageItem, 0, cachePageSize)
		for j := 0; j < cachePageSize; j++ {
			page = append(page, PageItem{
				ID:   fmt.Sprintf("%d-%d", i+1, j+1),
				Date: d.Add(-1 * time.Duration(i*cachePageSize+j) * time.Minute),
			})
		}
		decsTable = append(decsTable, page)
	}

	t.Cleanup(func() {
		client.LTrim(defaultCtx, acseKey, 1, 0)
		client.LTrim(defaultCtx, decsKey, 1, 0)
	})

	t.Run("Test push page acse", func(t *testing.T) {
		want := make([]string, 0, len(acseTable))
		err := cacheDb.PushPages(nil, acseKey, acseTable, false)
		test.AssertNoError(t, err, "when push pages: %s")
		idx := make([]time.Time, 0, len(acseTable))
		for _, v := range acseTable {
			s, _ := json.Marshal(v)
			want = append(want, string(s))
			idx = append(idx, v[0].Date)
		}
		got, _ := client.LRange(defaultCtx, acseKey, 1, -1).Result()
		test.AssertEqual(t, want, got)

		wantIdx, _ := json.Marshal(idx)
		gotIdx, _ := client.LIndex(defaultCtx, acseKey, 0).Result()
		test.AssertEqual(t, string(wantIdx), gotIdx)
	})

	t.Run("Test push page decs", func(t *testing.T) {
		want := make([]string, 0, len(decsTable))
		err := cacheDb.PushPages(nil, decsKey, decsTable, false)
		test.AssertNoError(t, err, "when push pages: %s")
		idx := make([]time.Time, 0, len(decsTable))
		for _, v := range decsTable {
			s, _ := json.Marshal(v)
			want = append(want, string(s))
			idx = append(idx, v[0].Date)
		}
		got, _ := client.LRange(defaultCtx, decsKey, 1, -1).Result()
		test.AssertEqual(t, want, got)

		wantIdx, _ := json.Marshal(idx)
		gotIdx, _ := client.LIndex(defaultCtx, decsKey, 0).Result()
		test.AssertEqual(t, string(wantIdx), gotIdx)
	})

	t.Run("Test query page acse", func(t *testing.T) {
		for i, v := range acseTable {
			got, err := cacheDb.QueryPage(nil, acseKey, i+1)
			test.AssertNoError(t, err, fmt.Sprintf("when query page %d: ", i)+"%s")
			test.AssertEqual(t, v, got)
		}
	})

	t.Run("Test query page decs", func(t *testing.T) {
		for i, v := range decsTable {
			got, err := cacheDb.QueryPage(nil, decsKey, i+1)
			test.AssertNoError(t, err, fmt.Sprintf("when query page %d: ", i)+"%s")
			test.AssertEqual(t, v, got)
		}
	})

	t.Run("Test query page by date acse", func(t *testing.T) {
		d1 := d.Add(3 * time.Minute) // from 5th item of 1st page
		got, err := cacheDb.QueryPageByDate(acseKey, d1, true)
		test.AssertNoError(t, err, "when query 1: %s")
		want := acseTable[0][4:]
		test.AssertEqual(t, want, got)

		d2 := d.Add(time.Duration(1*cachePageSize+12) * time.Minute) // from 14th item of 2nd page
		got, err = cacheDb.QueryPageByDate(acseKey, d2, true)
		test.AssertEqual(t, err, ErrNoEnoughPages)
		want = append(acseTable[1][13:], acseTable[2]...)
		want = append(want, PageItem{ID: ":last", Date: acseTable[2][19].Date})
		test.AssertEqual(t, want, got)

		d3 := d.Add(time.Duration(2*cachePageSize+1) * time.Minute) // from 3rd item of 3rd page
		got, err = cacheDb.QueryPageByDate(acseKey, d3, true)
		test.AssertEqual(t, ErrNoEnoughPages, err)
		want = append(acseTable[2][2:], PageItem{ID: ":last", Date: acseTable[2][19].Date})
		test.AssertEqual(t, want, got)

		d4 := d.Add(time.Duration(4*cachePageSize) * time.Minute) // very late
		got, err = cacheDb.QueryPageByDate(acseKey, d4, true)
		test.AssertEqual(t, ErrNotFound, err)
		test.AssertEqual(t, []PageItem{{ID: ":last", Date: acseTable[2][19].Date}}, got)
	})

	t.Run("Test query page by date decs", func(t *testing.T) {
		d1 := d.Add(-3 * time.Minute) // from 5th item of 1st page
		got, err := cacheDb.QueryPageByDate(decsKey, d1, false)
		test.AssertNoError(t, err, "when query 1: %s")
		want := decsTable[0][4:]
		test.AssertEqual(t, want, got)

		d2 := d.Add(-1 * time.Duration(1*cachePageSize+12) * time.Minute) // from 14th item of 2nd page
		got, err = cacheDb.QueryPageByDate(decsKey, d2, false)
		test.AssertEqual(t, err, ErrNoEnoughPages)
		want = append(decsTable[1][13:], decsTable[2]...)
		want = append(want, PageItem{ID: ":last", Date: decsTable[2][19].Date})
		test.AssertEqual(t, want, got)

		d3 := d.Add(-1 * time.Duration(2*cachePageSize+1) * time.Minute) // from 3rd item of 3rd page
		got, err = cacheDb.QueryPageByDate(decsKey, d3, false)
		test.AssertEqual(t, ErrNoEnoughPages, err)
		want = append(decsTable[2][2:], PageItem{ID: ":last", Date: decsTable[2][19].Date})
		test.AssertEqual(t, want, got)

		d4 := d.Add(-1 * time.Duration(4*cachePageSize) * time.Minute) // very late
		got, err = cacheDb.QueryPageByDate(decsKey, d4, false)
		test.AssertEqual(t, ErrNotFound, err)
		test.AssertEqual(t, []PageItem{{ID: ":last", Date: decsTable[2][19].Date}}, got)
	})

	t.Run("Test remove item", func(t *testing.T) {
		/*
			err := cacheDb.RemoveItem(key, "2-1")
			test.AssertNoError(t, err, "when remove item 2-1: %s")
			got, err := cacheDb.CachePopPage(key)
			test.AssertNoError(t, err, "when pop page 2: %s")
			want := table[1][1:]
			test.AssertEqual(t, want, got)

			// remove non-existent item
			err = cacheDb.RemoveItem(key, "3-4")
			test.AssertNoError(t, err, "when remove item 3-4: %s")
			got, err = cacheDb.CachePopPage(key)
			test.AssertNoError(t, err, "when pop page 3: %s")
			want = table[2]
			test.AssertEqual(t, want, got)
		*/
	})
}
