package db

import (
	"fmt"

	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/redis/go-redis/v9"
)

const (
	redis_addr string = "localhost"
	redis_auth int    = 0
)

var redis_conn = []int{2}
var redis_port = []int{6739}

type RdConn struct {
	absConn[*RdConn]
	client *redis.Client
}

func newRdConn(client interface{}, lg logging.Logger, pool *ConnPool[*RdConn]) (c *RdConn, ok bool) {
	if client, ok := client.(*redis.Client); ok {
		return &RdConn{
			absConn: absConn[*RdConn]{
				lg:   lg,
				pool: pool,
			},
			client: client,
		}, true
	}
	return nil, false
}

func (c *RdConn) Close() {
	if c.client == nil {
		return
	}
	c.client = nil
	c.absConn.close()
}

func (c *RdConn) Get(key string) (result string, err error) {
	logger := c.lg
	if c.client == nil {
		logger.Error("[Model.Redis] Connection is closed.", nil)
		return "", ErrConnClosed
	}
	result, e := c.client.Get(defaultCtx, key).Result()
	if e == redis.Nil {
		logger.Error("[Model.Redis] Cannot find key", e)
		return "", ErrNotFound
	}
	if e != nil {
		logger.Error("[Model.Redis] Cannot get value", e)
		return "", ErrDbInternal
	}
	return result, nil
}

func (c *RdConn) SetString(key string, value string) error {
	logger := c.lg
	if c.client == nil {
		logger.Error("[Model.Redis] Connection is closed.", nil)
		return ErrConnClosed
	}
	if e := c.client.Set(defaultCtx, key, value, 0).Err(); e != nil {
		logger.Error("[Model.Redis] Cannot set value", e)
		return ErrDbInternal
	}
	return nil
}

func newRdConnPool(cfg config.Config, lg logging.Logger, db int) *ConnPool[*RdConn] {
	p := ConnPool[*RdConn]{
		lg:       lg,
		capacity: redis_conn[db],
		using:    0,
		newConn:  newRdConn,
	}
	logger := lg

	client := redis.NewClient(&redis.Options{
		Addr: func(db int) string {
			return fmt.Sprintf("%s:%d", redis_addr, redis_port[db])
		}(db),
		Password: cfg.RdSecret,
		DB:       db,
		PoolSize: 2 * redis_conn[db],
	})
	if client == nil {
		logger.Error("[Db.Redis] Cannot create Redis client", nil)
		panic("Cannot create Redis client")
	}
	p.client = client
	return &p
}
