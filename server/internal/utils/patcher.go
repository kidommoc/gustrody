package utils

import (
	"time"
)

/*
Patcher is used to patch multiple P->T requests to a single []P->map[P]T request.
In gustrody, it could be used to query users or posts from db.
Patcher would patch as much requests as possible in a given time,
and send the request([]P) to the handler (a channel to the customer handler function).
It's necessary to inherit Patcher struct and implement the specified one.
Please see patcher_test.go for an example.
*/

type Patcher[P comparable, T any] struct {
	tm   int // timer duration
	cllt chan (struct {
		p P
		t chan<- (map[P]T)
	}) // collector channel
}

// tm: timer duration in milliseconds
func NewPatcher[P comparable, T any](tm int, hdl chan<- struct {
	p []P
	t []chan<- (map[P]T)
}) *Patcher[P, T] {
	p := Patcher[P, T]{tm, make(chan struct {
		p P
		t chan<- map[P]T
	})}
	go p.collector(hdl)
	return &p
}

func (p *Patcher[P, T]) collector(hdl chan<- struct {
	p []P
	t []chan<- (map[P]T)
}) {
	var timer *time.Timer
	// timer := time.NewTimer(time.Duration(p.tm) * time.Millisecond)
	prmt := make([]P, 0)
	chnn := make([]chan<- (map[P]T), 0)
	for {
		if timer == nil {
			pt := <-p.cllt
			prmt = append(prmt, pt.p)
			chnn = append(chnn, pt.t)
			timer = time.NewTimer(time.Duration(p.tm) * time.Millisecond)
		} else {
			select {
			case pt := <-p.cllt:
				prmt = append(prmt, pt.p)
				chnn = append(chnn, pt.t)
			case <-timer.C:
				timer = nil
				hdl <- struct {
					p []P
					t []chan<- map[P]T
				}{prmt, chnn}
				prmt = make([]P, 0)
				chnn = make([]chan<- (map[P]T), 0)
			}
		}
	}
}

func (p *Patcher[P, T]) add(prmt P, ret chan<- (map[P]T)) {
	p.cllt <- struct {
		p P
		t chan<- (map[P]T)
	}{prmt, ret}
}
