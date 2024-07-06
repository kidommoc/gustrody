package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/logging"

	_ "github.com/lib/pq"
)

var defaultCtx = context.Background()

// connections

type Conn interface {
	Close()
	Closed() bool
}

type absConn[C Conn] struct {
	pool *connPool[C]
}

func (c *absConn[C]) close() {
	c.pool.mutex.Lock()
	logger := logging.Get()
	logger.Debug(fmt.Sprintf("closed, using: %d", c.pool.using))
	if c.pool.using > 0 {
		c.pool.using -= 1
	}
	if len(c.pool.listener) != 0 {
		sig := c.pool.listener[0]
		c.pool.listener = c.pool.listener[1:]
		sig <- true
	} else {
		c.pool.mutex.Unlock()
	}
	c.pool = nil
}

func (c *absConn[C]) Closed() bool {
	return c.pool == nil
}

// connection pool

type ConnPool[C Conn] interface {
	Open() (conn C, err error)
}

type connPool[C Conn] struct {
	mutex    sync.Mutex
	listener []chan bool
	capacity int
	using    int
	client   interface{}
	newConn  func(interface{}, *connPool[C]) (C, bool)
}

// should be async
func (p *connPool[C]) Open() (conn C, err error) {
	p.mutex.Lock()
	if p.using >= p.capacity {
		sig := make(chan bool)
		p.listener = append(p.listener, sig)
		p.mutex.Unlock()
		<-sig
	}
	c, ok := p.newConn(p.client, p)
	if !ok {
		p.mutex.Unlock()
		return c, ErrNoConn
	}
	p.using += 1
	p.mutex.Unlock()
	return c, nil
}

// auth pool

var authPoolIns *connPool[RdConn] = nil

func AuthPool(cfg *config.Config) ConnPool[RdConn] {
	if authPoolIns != nil {
		return authPoolIns
	}
	if cfg == nil {
		return nil
	}
	authPoolIns = newRdConnPool(*cfg, redis_auth)
	return authPoolIns
}

// main pool

var mainPoolIns *connPool[PqConn] = nil

func MainPool(cfg *config.Config) ConnPool[PqConn] {
	if mainPoolIns != nil {
		return mainPoolIns
	}
	if cfg == nil {
		return nil
	}
	mainPoolIns = newPqConnPool(*cfg)
	return mainPoolIns
}

func Init() {
	cfg := config.Get()
	logger := logging.Get()
	AuthPool(&cfg)
	logger.Info("[Db]Initailized AuthPool")
	MainPool(&cfg)
	logger.Info("[Db]Initailized MainPool")
}
