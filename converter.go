package csv

import (
	"reflect"
)

type Converter interface {
	Covert(reflect.Value) ([]string, error)
}
