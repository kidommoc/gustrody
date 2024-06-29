package db

import (
	"database/sql"
	"fmt"

	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/logging"
)

const (
	postgres_addr string = "localhost:5432"
	// the next commented lines are the addresses
	//   after docker compose applied
	// postgres_addr string = "db:5432"
	main_conn int = 12
)

type PqConn struct {
	absConn[*PqConn]
	client *sql.DB
}

type Tx struct {
	lg logging.Logger
	tx *sql.Tx
}

func newPqConn(client interface{}, lg logging.Logger, pool *ConnPool[*PqConn]) (c *PqConn, ok bool) {
	if client, ok := client.(*sql.DB); ok {
		return &PqConn{
			absConn: absConn[*PqConn]{
				lg:   lg,
				pool: pool,
			},
			client: client,
		}, true
	}
	return nil, false
}

func (c *PqConn) Close() {
	if c.client == nil {
		return
	}
	c.client = nil
	c.absConn.close()
}

// just wrapping for now

type X interface {
	Query(string, ...any) (*sql.Rows, error)
	QueryRow(string, ...any) *sql.Row
	Exec(string, ...any) (sql.Result, error)
}

func exec(x X, q string, args ...any) (affected int64, err error) {
	r, e := x.Exec(q, args...)
	if e != nil {
		return 0, e
	}
	return r.RowsAffected()
}

func (c *PqConn) Query(q string, args ...any) (rows *sql.Rows, err error) {
	logger := c.lg
	rows, err = c.client.Query(q, args...)
	if err != nil {
		logger.Error("[Db] Cannot query", err)
		return nil, ErrDbInternal
	}
	return
}

func (c *PqConn) QueryOne(q string, args ...any) *sql.Row {
	return c.client.QueryRow(q, args...)
}

func (c *PqConn) Exec(q string, args ...any) (affected int64, err error) {
	logger := c.lg
	affected, err = exec(c.client, q, args...)
	if err != nil {
		logger.Error("[Db] Cannot execute", err)
		return 0, ErrDbInternal
	}
	return
}

func (c *PqConn) BeginTx() (tx *Tx, err error) {
	logger := c.lg
	t, e := c.client.Begin()
	if e != nil {
		logger.Error("[Db] Cannot start transaction", err)
		return nil, ErrDbInternal
	}
	return &Tx{lg: c.lg, tx: t}, nil
}

// CLOSE ROWS!
func (t *Tx) Query(q string, args ...any) (rows *sql.Rows, err error) {
	logger := t.lg
	rows, err = t.tx.Query(q, args...)
	if err != nil {
		logger.Error("[Db] Cannot query", err)
		return nil, ErrDbInternal
	}
	return
}

func (t *Tx) QueryOne(q string, args ...any) *sql.Row {
	return t.tx.QueryRow(q, args...)
}

func (t *Tx) Exec(q string, args ...any) (affected int64, err error) {
	logger := t.lg
	affected, err = exec(t.tx, q, args...)
	if err != nil {
		logger.Error("[Db] Cannot execute", err)
		return 0, ErrDbInternal
	}
	return
}

func (t *Tx) Commit() error {
	logger := t.lg
	err := t.tx.Commit()
	if err != nil {
		logger.Error("[Db] Cannot commit transaction", err)
		return ErrDbInternal
	}
	return nil
}

func newPqConnPool(cfg config.Config, lg logging.Logger) *ConnPool[*PqConn] {
	p := ConnPool[*PqConn]{
		lg:       lg,
		capacity: main_conn,
		using:    0,
		newConn:  newPqConn,
	}
	logger := lg

	connStr := fmt.Sprintf(
		"postgresql://%s:%s@%s/austrody?sslmode=disable", // db name: austrody
		cfg.PqUser, cfg.PqSecret,
		postgres_addr,
	)
	db, e := sql.Open("postgres", connStr)
	if e != nil {
		logger.Error("[Db.Postgres] Cannot create Postgres client", nil)
		panic("Cannot create Postgres client")
	}
	db.SetMaxOpenConns(2 * main_conn)
	p.client = db
	return &p
}
