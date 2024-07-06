package models

import (
	"database/sql"
	"fmt"

	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/db"
	"github.com/kidommoc/gustrody/internal/logging"
)

type mockingPqConn struct {
	client *sql.DB
}

func (c *mockingPqConn) Close() {}

func (c *mockingPqConn) Closed() bool { return false }

func (c *mockingPqConn) Query(q string, args ...any) (rows *sql.Rows, err error) {
	return c.client.Query(q, args...)
}

func (c *mockingPqConn) QueryOne(q string, args ...any) *sql.Row {
	return c.client.QueryRow(q, args...)
}

func (c *mockingPqConn) Exec(q string, args ...any) (affected int64, err error) {
	r, err := c.client.Exec(q, args...)
	if err != nil {
		return 0, err
	}
	return r.RowsAffected()
}

func (c *mockingPqConn) BeginTx() (tx db.Tx, err error) {
	tt, err := c.client.Begin()
	if err != nil {
		return nil, err
	}
	return &mockingTx{tt}, nil
}

type mockingTx struct {
	tx *sql.Tx
}

func (t *mockingTx) Query(q string, args ...any) (rows *sql.Rows, err error) {
	return t.tx.Query(q, args...)
}

func (t *mockingTx) QueryOne(q string, args ...any) *sql.Row {
	return t.tx.QueryRow(q, args...)
}

func (t *mockingTx) Exec(q string, args ...any) (affected int64, err error) {
	r, err := t.tx.Exec(q, args...)
	if err != nil {
		return 0, err
	}
	return r.RowsAffected()
}

func (t *mockingTx) Commit() error {
	return t.tx.Commit()
}

type mockingPqPool[C db.PqConn] struct {
	lg     logging.Logger
	client *sql.DB
}

func newMockingPqPool(cfg config.Config, lg logging.Logger) db.ConnPool[db.PqConn] {
	logger := lg
	connStr := fmt.Sprintf(
		"postgresql://%s:%s@%s/austrody?sslmode=disable", // db name: austrody
		cfg.PqUser, cfg.PqSecret,
		"localhost:5432",
	)
	db, e := sql.Open("postgres", connStr)
	if e != nil {
		logger.Error("[Db.Postgres] Cannot create Postgres client", nil)
		panic("Cannot create Postgres client")
	}
	p := mockingPqPool[*mockingPqConn]{
		lg:     lg,
		client: db,
	}
	return &p
}

func (p *mockingPqPool[C]) Open() (conn db.PqConn, err error) {
	return &mockingPqConn{p.client}, nil
}
