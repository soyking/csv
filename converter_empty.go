package csv

import (
	"reflect"
)

// ------ convert string type to set in tag of `csv`
type EmptyConverter struct {
}

func NewEmptyConverter() (EmptyConverter, error) {
	var c EmptyConverter
	return c, nil
}

func (c EmptyConverter) Covert(fv reflect.Value) ([]string, error) {
	return []string{}, nil
}
