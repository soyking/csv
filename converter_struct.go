package csv

import (
	"reflect"
	"strings"
)

type StructCoverter struct {
	colName    string
	converters []Converter
}

func NewStructConverter(tag string, fv reflect.Value) (StructCoverter, error) {
	var c StructCoverter

	cons := strings.Split(tag, ";")
	c.colName = cons[0]
	if c.colName != "-" {
		var err error
		if c.converters, err = getConverters(fv); err != nil {
			return c, err
		}
	}

	return c, nil
}

func (c StructCoverter) Covert(fv reflect.Value) ([]string, error) {
	if c.colName == "-" {
		return []string{}, nil
	}

	marshalerType := reflect.TypeOf(new(Marshaler)).Elem()
	if fv.Type().Implements(marshalerType) {
		m := fv.Interface().(Marshaler)
		return m.MarshalCSV()
	}

	return encodeRow(fv, c.converters)
}
