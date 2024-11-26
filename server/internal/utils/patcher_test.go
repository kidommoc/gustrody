package utils

import (
	"sync"
	"testing"
)

type myPatcher struct {
	*Patcher[int, int]
}

func newMyPatcher() *myPatcher {
	ch := make(chan struct {
		p []int
		t []chan<- (map[int]int)
	})
	p := NewPatcher(200, ch)
	mp := myPatcher{p}
	go mp.handler(ch)
	return &mp
}

// make a request. simply make channel and add.
func (mp *myPatcher) Request(n int) int {
	ch := make(chan map[int]int)
	mp.add(n, ch)
	// Note that patcher will NOT map any single request to its result.
	// It only return a big map, and the exact work remains to be done.
	return (<-ch)[n]
}

func (mp *myPatcher) handler(ch <-chan struct {
	p []int
	t []chan<- (map[int]int)
}) {
	for {
		in := <-ch
		go func(p []int, t []chan<- (map[int]int)) {
			m := make(map[int]int)
			for n := range p {
				m[n] = n + 1
			}
			for i := 0; i < len(t); i++ {
				t[i] <- m
			}
		}(in.p, in.t)
	}
}
func TestPatcher(t *testing.T) {
	mp := newMyPatcher()
	// assume mp is not nil
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int, t *testing.T) {
			result := mp.Request(n)
			if result != n+1 {
				t.Errorf("error!")
			}
			t.Logf("req: %d, res: %d\n", n, result)
			wg.Done()
		}(i, t)
	}
	wg.Wait()
}
