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

type PqConn interface {
	Conn
	Query(q string, args ...any) (rows *sql.Rows, err error)
	QueryOne(q string, args ...any) *sql.Row
	Exec(q string, args ...any) (affected int64, err error)
	BeginTx() (tx Tx, err error)
}

type pqConn struct {
	absConn[PqConn]
	client *sql.DB
}

type Tx interface {
	Query(q string, args ...any) (rows *sql.Rows, err error)
	QueryOne(q string, args ...any) *sql.Row
	Exec(q string, args ...any) (affected int64, err error)
	Commit() error
}

type tx struct {
	tx *sql.Tx
}

func newPqConn(client interface{}, pool *connPool[PqConn]) (c PqConn, ok bool) {
	if client, ok := client.(*sql.DB); ok {
		return &pqConn{
			absConn: absConn[PqConn]{
				pool: pool,
			},
			client: client,
		}, true
	}
	return nil, false
}

func (c *pqConn) Close() {
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

func (c *pqConn) Query(q string, args ...any) (rows *sql.Rows, err error) {
	logger := logging.Get()
	rows, err = c.client.Query(q, args...)
	if err != nil {
		logger.Error("[Db] Cannot query", err)
		return nil, ErrDbInternal
	}
	return
}

func (c *pqConn) QueryOne(q string, args ...any) *sql.Row {
	return c.client.QueryRow(q, args...)
}

func (c *pqConn) Exec(q string, args ...any) (affected int64, err error) {
	logger := logging.Get()
	affected, err = exec(c.client, q, args...)
	if err != nil {
		logger.Error("[Db] Cannot execute", err)
		return 0, ErrDbInternal
	}
	return
}

func (c *pqConn) BeginTx() (t Tx, err error) {
	logger := logging.Get()
	tt, e := c.client.Begin()
	if e != nil {
		logger.Error("[Db] Cannot start transaction", err)
		return nil, ErrDbInternal
	}
	return &tx{tt}, nil
}

// CLOSE ROWS!
func (t *tx) Query(q string, args ...any) (rows *sql.Rows, err error) {
	logger := logging.Get()
	rows, err = t.tx.Query(q, args...)
	if err != nil {
		logger.Error("[Db] Cannot query", err)
		return nil, ErrDbInternal
	}
	return
}

func (t *tx) QueryOne(q string, args ...any) *sql.Row {
	return t.tx.QueryRow(q, args...)
}

func (t *tx) Exec(q string, args ...any) (affected int64, err error) {
	logger := logging.Get()
	affected, err = exec(t.tx, q, args...)
	if err != nil {
		logger.Error("[Db] Cannot execute", err)
		return 0, ErrDbInternal
	}
	return
}

func (t *tx) Commit() error {
	logger := logging.Get()
	err := t.tx.Commit()
	if err != nil {
		logger.Error("[Db] Cannot commit transaction", err)
		return ErrDbInternal
	}
	return nil
}

func newPqConnPool(cfg config.Config) *connPool[PqConn] {
	p := connPool[PqConn]{
		capacity: main_conn,
		using:    0,
		newConn:  newPqConn,
	}
	logger := logging.Get()

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
