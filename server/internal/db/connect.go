package db

import (
	"context"

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
	c.pool.returnConn()
	c.pool = nil
}

func (c *absConn[C]) Closed() bool {
	return c.pool == nil
}

// connection pool

type ConnPool[C Conn] struct {
	lg       logging.Logger
	capacity int
	using    int
	client   interface{}
	newConn  func(interface{}, logging.Logger, *ConnPool[C]) (C, bool)
}

// should be async
func (p *ConnPool[C]) Open() (c C, err error) {
	if p.using >= p.capacity {
		return c, ErrNoConn
	}
	c, ok := p.newConn(p.client, p.lg, p)
	if !ok {
		return c, ErrNoConn
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
