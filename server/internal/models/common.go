package models

import (
	"database/sql"
	"database/sql/driver"

	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/lib/pq"
)

const redisMaxRetries = 20

// magic
type SV[P driver.Valuer] interface {
	sql.Scanner
	*P
}

type Array[P driver.Valuer, T SV[P]] struct {
	lg   logging.Logger
	data []P
}

func NewArray[P driver.Valuer, T SV[P]](arr []P) *Array[P, T] {
	return &Array[P, T]{data: arr}
}

func (arr *Array[P, T]) Data() []P {
	return arr.data
}

func (arr *Array[P, T]) Append(item P) {
	arr.data = append(arr.data, item)
}

func (arr *Array[P, T]) ScanArray() interface {
	driver.Valuer
	sql.Scanner
} {
	return pq.Array(&arr.data)
}

func (arr *Array[P, T]) ValueArray() interface {
	driver.Valuer
	sql.Scanner
} {
	return pq.Array(arr.data)
}
