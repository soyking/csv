package csv

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Marshaler is an interface for objects which can Marshal themselves into CSV.
type Marshaler interface {
	MarshalCSV() ([]string, error)
}

type encoder struct {
	*csv.Writer
	buffer *bytes.Buffer
}

// Marshal returns the CSV encoding of i, which must be a slice of struct types.
//
// Marshal traverses the slice and encodes the primative values.
//
// The first row of the CSV output is a header row. The column names are based
// on the field name.  If a different name is required a struct tag can be used
// to define a new name.
//
//   Field string `csv:"Column Name"`
//
// To skip encoding a field use the "-" as the tag value.
//
//   Field string `csv:"-"`
//
// Boolean fields can use string values to define true or false.
//   Bool bool `true:"Yes" false:"No"`
func Marshal(d interface{}) ([]byte, error) {
	// validate the interface
	// create a new encoder
	//   assing the cfields
	// get the headers
	// encoder each row

	data := reflect.ValueOf(d)
	if data.Kind() != reflect.Slice {
		for data.Kind() == reflect.Ptr {
			data = data.Elem()
		}
		if data.Kind() != reflect.Slice {
			return []byte{}, errors.New("only slices and pointer to slices can be marshalled")
		}

	}

	if data.Len() == 0 {
		return []byte{}, nil
	}
	col := data.Index(0)
	enc, err := newEncoder(col)
	if err != nil {
		return []byte{}, err
	}

	var converters []Converter
	converters, err = getConverters(col)
	if err != nil {
		return []byte{}, err
	}

	err = enc.encodeAll(data, converters)
	if err != nil {
		return []byte{}, err
	}

	enc.Flush()
	return enc.buffer.Bytes(), nil
}

func newEncoder(col reflect.Value) (*encoder, error) {
	b := bytes.NewBuffer([]byte{})

	enc := &encoder{
		buffer: b,
		Writer: csv.NewWriter(b),
	}

	err := enc.Write(colNames(col))

	return enc, err
}

// colNames takes a struct and returns the computed columns names for each
// field.
func colNames(col reflect.Value) (out []string) {
	l := col.Type().NumField()
	for i := 0; i < l; i++ {
		f := col.Field(i)
		t := col.Type().Field(i)
		h, ok := getColName(t, f)
		if ok {
			out = append(out, h...)
		}
	}

	return
}

// getColName returns the header name to use for the given StructField
// This can be a user defined name (via the Tag) or a default name.
func getColName(t reflect.StructField, f reflect.Value) ([]string, bool) {
	tag := strings.Split(t.Tag.Get("csv"), ";")[0]
	if tag == "-" {
		return []string{}, false
	}

	// If there is no tag set, use a default name
	if tag == "" {
		if f.Kind() == reflect.Struct {
			return colNames(f), true
		} else {
			for f.Kind() == reflect.Ptr {
				f = f.Elem()
			}
			if f.Kind() == reflect.Struct {
				return colNames(f), true
			}
		}
		return []string{t.Name}, true
	}

	return []string{tag}, true
}

func getConverters(col reflect.Value) ([]Converter, error) {
	var c []Converter
	l := col.Type().NumField()
	for i := 0; i < l; i++ {
		f := col.Field(i)
		tag := col.Type().Field(i).Tag.Get("csv")

		if strings.Split(tag, ";")[0] != "=" {
			if converter, err := genConverter(f, tag); err != nil {
				return c, err
			} else {
				c = append(c, converter)
			}
		}
	}

	return c, nil
}

func genConverter(f reflect.Value, tag string) (Converter, error) {
	var converter Converter
	var err error

	switch f.Kind() {
	case reflect.String:
		converter, err = NewStringConverter(tag)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		converter, err = NewIntConverter(tag)
	case reflect.Float32, reflect.Float64:
		converter, err = NewFloatConverter(tag)
	case reflect.Bool:
		converter, err = NewBoolConverter(tag)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		converter, err = NewUintConverter(tag)
	case reflect.Struct:
		converter, err = NewStructConverter(tag, f)
	case reflect.Ptr:
		for f.Kind() == reflect.Ptr {
			f = f.Elem()
		}
		return genConverter(f, tag)
	default:
		if strings.Split(tag, ";")[0] == "-" {
			converter, err = NewEmptyConverter()
		} else {
			return converter, errors.New(fmt.Sprintf("not support this type:%s", f.Kind()))
		}
	}

	return converter, err
}

// encodeAll iterates over each item in data, encoder it then writes it
func (enc *encoder) encodeAll(data reflect.Value, converters []Converter) error {
	n := data.Len()
	for i := 0; i < n; i++ {
		row, err := encodeRow(data.Index(i), converters)
		if err != nil {
			return err
		}

		err = enc.Write(row)
		if err != nil {
			return err
		}
	}

	return nil
}

// encodes a struct into a CSV row
func encodeRow(v reflect.Value, converters []Converter) ([]string, error) {
	var row []string
	l := v.Type().NumField()
	for i := 0; i < l; i++ {
		f := v.Field(i)
		for f.Kind() == reflect.Ptr {
			f = f.Elem()
		}
		o, err := converters[i].Covert(f)
		if err != nil {
			return []string{}, err
		}
		row = append(row, o...)
	}

	return row, nil
}
