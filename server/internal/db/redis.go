package db

import (
	"fmt"

	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/utils"
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

func newRdConn(client interface{}, pool *ConnPool[*RdConn]) (c *RdConn, ok bool) {
	if client, ok := client.(*redis.Client); ok {
		return &RdConn{
			absConn: absConn[*RdConn]{
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

func (c *RdConn) Get(key string) (result string, err utils.Error) {
	logger := logging.Get()
	if c.client == nil {
		err = newErr(ErrConnClosed)
		logger.Error("[Model.Redis] Connection is closed.", err)
		return "", err
	}
	result, e := c.client.Get(defaultCtx, key).Result()
	if e == redis.Nil {
		err = newErr(ErrNotFound, e.Error())
		logger.Error("[Model.Redis] Cannot find key", err)
		return "", err
	}
	if e != nil {
		err = newErr(ErrDbInternal, e.Error())
		logger.Error("[Model.Redis] Cannot get value", err)
		return "", err
	}
	return result, nil
}

func (c *RdConn) SetString(key string, value string) utils.Error {
	logger := logging.Get()
	if c.client == nil {
		err := newErr(ErrConnClosed)
		logger.Error("[Model.Redis] Connection is closed.", err)
		return err
	}
	if e := c.client.Set(defaultCtx, key, value, 0).Err(); e != nil {
		err := newErr(ErrDbInternal, e.Error())
		logger.Error("[Model.Redis] Cannot set value", err)
		return err
	}
	return nil
}

func newRdConnPool(cfg config.Config, db int) *ConnPool[*RdConn] {
	p := ConnPool[*RdConn]{
		capacity: redis_conn[db],
		using:    0,
		newConn:  newRdConn,
	}

	client := redis.NewClient(&redis.Options{
		Addr: func(db int) string {
			return fmt.Sprintf("%s:%d", redis_addr, redis_port[db])
		}(db),
		Password: cfg.RdSecret,
		DB:       db,
		PoolSize: redis_conn[db],
	})
	if client == nil {
		logger := logging.Get()
		logger.Error("[Db.Redis] Cannot create Redis client", nil)
		panic("Cannot create Redis client")
	}
	p.client = client
	return &p
}
