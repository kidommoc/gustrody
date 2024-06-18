package db

import (
	"context"

	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/utils"

	_ "github.com/lib/pq"
)

var defaultCtx = context.Background()

// connections

type Conn interface {
	Close()
	Closed() bool
}

type absConn[C Conn] struct {
	pool *ConnPool[C]
}

func (c *absConn[C]) close() {
	c.pool.returnConn()
	c.pool = nil
}

func (c *absConn[C]) Closed() bool {
	return c.pool == nil
}

// connection pool

type ConnPool[C Conn] struct {
	capacity int
	using    int
	client   interface{}
	newConn  func(interface{}, *ConnPool[C]) (C, bool)
}

// should be async
func (p *ConnPool[C]) Open() (c C, err utils.Error) {
	if p.using >= p.capacity {
		return c, newErr(ErrNoConn)
	}
	c, ok := p.newConn(p.client, p)
	if !ok {
		// handle error
		return c, newErr(100)
	}
	p.using += 1
	return c, nil
}

func (p *ConnPool[C]) returnConn() {
	if p.using > 0 {
		p.using -= 1
	}
}

// auth pool

var authPoolIns *ConnPool[*RdConn] = nil

func AuthPool() *ConnPool[*RdConn] {
	if authPoolIns != nil {
		return authPoolIns
	}
	cfg := config.Get()
	authPoolIns = newRdConnPool(cfg, redis_auth)
	return authPoolIns
}

// main pool

var mainPoolIns *ConnPool[*PqConn] = nil

func MainPool() *ConnPool[*PqConn] {
	if mainPoolIns != nil {
		return mainPoolIns
	}
	cfg := config.Get()
	mainPoolIns = newPqConnPool(cfg)
	return mainPoolIns
}

func Init() {
	logger := logging.Get()
	AuthPool()
	logger.Info("[Db]Initailized AuthPool")
	MainPool()
	logger.Info("[Db]Initailized MainPool")
}
