package db

import (
	"context"
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
	lg   logging.Logger
	pool *ConnPool[C]
}

func (c *absConn[C]) close() {
	c.pool.mu.Lock()
	if c.pool.using > 0 {
		c.pool.using -= 1
	}
	if len(c.pool.listener) != 0 {
		sig := c.pool.listener[0]
		c.pool.listener = c.pool.listener[1:]
		sig <- true
	} else {
		c.pool.mu.Unlock()
	}
	c.pool = nil
}

func (c *absConn[C]) Closed() bool {
	return c.pool == nil
}

// connection pool

type ConnPool[C Conn] struct {
	mu       sync.Mutex
	listener []chan bool
	lg       logging.Logger
	capacity int
	using    int
	client   interface{}
	newConn  func(interface{}, logging.Logger, *ConnPool[C]) (C, bool)
}

// should be async
func (p *ConnPool[C]) Open() (c C, err error) {
	p.mu.Lock()
	if p.using >= p.capacity {
		sig := make(chan bool)
		p.listener = append(p.listener, sig)
		p.mu.Unlock()
		<-sig
	}
	c, ok := p.newConn(p.client, p.lg, p)
	if !ok {
		p.mu.Unlock()
		return c, ErrNoConn
	}
	p.using += 1
	p.mu.Unlock()
	return c, nil
}

// auth pool

var authPoolIns *ConnPool[*RdConn] = nil

func AuthPool(cfg *config.Config, lg logging.Logger) *ConnPool[*RdConn] {
	if authPoolIns != nil {
		return authPoolIns
	}
	if cfg == nil || lg == nil {
		return nil
	}
	authPoolIns = newRdConnPool(*cfg, lg, redis_auth)
	return authPoolIns
}

// main pool

var mainPoolIns *ConnPool[*PqConn] = nil

func MainPool(cfg *config.Config, lg logging.Logger) *ConnPool[*PqConn] {
	if mainPoolIns != nil {
		return mainPoolIns
	}
	if cfg == nil || lg == nil {
		return nil
	}
	mainPoolIns = newPqConnPool(*cfg, lg)
	return mainPoolIns
}

func Init() {
	cfg := config.Get()
	logger := logging.Get()
	AuthPool(&cfg, logger)
	logger.Info("[Db]Initailized AuthPool")
	MainPool(&cfg, logger)
	logger.Info("[Db]Initailized MainPool")
}
