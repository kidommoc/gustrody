package models

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/lib/pq"
)

// magic
type SV[P driver.Valuer] interface {
	sql.Scanner
	*P
}

type Array[P driver.Valuer, T SV[P]] struct {
	lg   logging.Logger
	data []P
}

func NewArray[P driver.Valuer, T SV[P]](arr []P, lgs ...logging.Logger) *Array[P, T] {
	if len(lgs) == 0 {
		return &Array[P, T]{data: arr, lg: logging.Get()}
	} else {
		return &Array[P, T]{data: arr, lg: lgs[0]}
	}
}

func (arr *Array[P, T]) Data() []P {
	return arr.data
}

func (arr *Array[P, T]) Append(item P) {
	arr.data = append(arr.data, item)
}

func (arr Array[P, T]) Value() (driver.Value, error) {
	logger := arr.lg
	items := make([]string, 0, len(arr.data))
	for _, v := range arr.data {
		value, err := v.Value()
		if err != nil {
			msg := fmt.Sprint("[Models] Failed to convert ", v, " to driver.Value")
			logger.Error(msg, err)
		}
		items = append(items, fmt.Sprint(value))
	}
	return fmt.Sprintf("{%s}", strings.Join(items, ",")), nil
}

func (arr *Array[P, T]) ToPqArray() interface {
	driver.Valuer
	sql.Scanner
} {
	return pq.Array(&arr.data)
}
