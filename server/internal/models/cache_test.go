package models

import (
	"encoding/json"
	"fmt"
	"sync"
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

type jsonableStruct struct {
	atomic bool
	B      bool    `json:"b"`
	S      *string `json:"s,omitempty"`
	F      float64 `json:"f,omitempty"`
	A      *[]int  `json:"a,omitempty"`
}

func (j *jsonableStruct) Json() (map[string]interface{}, error) {
	m, err := structToMap(j)
	if err != nil {
		return nil, err
	}
	m["__type"] = "::jsonable"
	return m, nil
}

func (j *jsonableStruct) Atomic() bool {
	return j.atomic
}

func TestCacheJson(t *testing.T) {
	logger := test.NewMockingLogger(t)
	client := initRedis(modelscfg, logger, redisOpt{
		Addr:    "localhost:6738",
		Db:      0,
		MaxConn: 10,
	})
	cacheDb := &CacheDb{logger, client}

	cacheKey := "testjsoncache"
	s1 := "foo"
	s2 := ""
	b, err := json.Marshal(map[string]interface{}{"__type": "::jsonable", "b": true, "s": "abcd", "a": []float64{1, 2}})
	test.AssertNoError(t, err, "when set cache: %s")
	cacheDb.SetString(cacheKey, string(b), false)

	t.Cleanup(func() {
		client.GetDel(defaultCtx, cacheKey)
	})

	table := []struct {
		input jsonableStruct
		want  map[string]interface{}
	}{{
		input: jsonableStruct{B: true, S: nil, F: 1},
		want:  map[string]interface{}{"__type": "::jsonable", "b": true, "s": "abcd", "f": 1.0, "a": []interface{}{1.0, 2.0}},
	}, {
		input: jsonableStruct{S: &s1, F: -1, A: &[]int{}},
		want:  map[string]interface{}{"__type": "::jsonable", "b": false, "s": "foo", "f": 0.0, "a": []interface{}{}},
	}, {
		input: jsonableStruct{B: true, S: &s2, A: &[]int{2, 1}},
		want:  map[string]interface{}{"__type": "::jsonable", "b": true, "s": "", "f": 0.0, "a": []interface{}{2.0, 1.0}},
	}}

	for i, v := range table {
		err = cacheDb.UpdateJson(cacheKey, &v.input)
		test.AssertNoError(t, err, fmt.Sprintf("when %d-th update: ", i)+"%s")
		ss, err := cacheDb.QueryString([]string{cacheKey})
		test.AssertNoError(t, err, fmt.Sprintf("when %d-th query: ", i)+"%s")
		var got map[string]interface{}
		err = json.Unmarshal([]byte(ss[cacheKey]), &got)
		test.AssertNoError(t, err, fmt.Sprintf("when %d-th unmarshal: ", i)+"%s")
		test.AssertEqual(t, v.want, got)
	}

	// ATOMIC UPDATE
	var wg sync.WaitGroup
	wg.Add(10)
	routine := func() {
		defer wg.Done()
		cacheDb.UpdateJson(cacheKey, &jsonableStruct{F: 1, atomic: true})
	}
	for i := 0; i < 10; i++ {
		go routine()
	}

	wg.Wait()
	ss, err := cacheDb.QueryString([]string{cacheKey})
	test.AssertNoError(t, err, "when query after concurrency: %s")
	var got map[string]interface{}
	err = json.Unmarshal([]byte(ss[cacheKey]), &got)
	test.AssertNoError(t, err, "when unmarshal after concurrency: %s")
	want := map[string]interface{}{"__type": "::jsonable", "b": false, "s": "", "f": 10.0, "a": []interface{}{2.0, 1.0}}
	test.AssertEqual(t, want, got)
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
