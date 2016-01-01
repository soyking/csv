package csv

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const MAX_ORDER = 99999

// Marshaler is an interface for objects which can Marshal themselves into CSV.
type Marshaler interface {
	MarshalCSV() ([]string, error)
}

type encoder struct {
	*csv.Writer
	buffer *bytes.Buffer
}

// Marshal returns the CSV encoding of d, which must be a slice of struct or pointer to a slice of struct.
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
// When a field is a string/int/uint/bool/float, you could use tag to add conditions to it, such as:
//
//   Married   bool    `csv:"married;true:yes;false:false"`
//
// You could view samples.The condition is something like:
//
//   $lt less than; $gt greater than; $lte less than or equal; $gte greater than or equal
//   $eq equal; $ne not equal
//   $default default val
//
// Please watch out that when comparing float values, sometimes it wouldn't give you correct answer
//
// When a field is a struct or a pointer to struct type, it will be flatten
func Marshal(d interface{}) ([]byte, error) {
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
	orders := getOrders(col)

	enc, err := newEncoder(col, orders)
	if err != nil {
		return []byte{}, err
	}

	var converters []Converter
	converters, err = getConverters(col)
	if err != nil {
		return []byte{}, err
	}

	err = enc.encodeAll(data, converters, orders)
	if err != nil {
		return []byte{}, err
	}

	enc.Flush()
	return enc.buffer.Bytes(), nil
}

func newEncoder(col reflect.Value, orders []int) (*encoder, error) {
	b := bytes.NewBuffer([]byte{})

	enc := &encoder{
		buffer: b,
		Writer: csv.NewWriter(b),
	}

	header, err := colNames(col, orders)
	if err != nil {
		return enc, err
	}

	err = enc.Write(header)
	return enc, err
}

// colNames takes a struct and returns the computed columns names for each
// field.
func colNames(col reflect.Value, orders []int) (out []string, err error) {
	l := col.Type().NumField()
	for i := 0; i < l; i++ {
		f := col.Field(i)
		t := col.Type().Field(i)
		h, err := getColName(t, f, orders)
		if err != nil {
			return []string{}, err
		}
		out = append(out, h...)
	}
	if orders != nil {
		out, err = reOrderRow(out, orders)
	}

	return
}

// getColName returns the header name to use for the given StructField
// This can be a user defined name (via the Tag) or a default name.
func getColName(t reflect.StructField, f reflect.Value, orders []int) ([]string, error) {
	tag := strings.Split(t.Tag.Get("csv"), ";")[0]
	if tag == "-" {
		return []string{}, nil
	}

	// If there is no tag set, use a default name
	if tag == "" {
		if f.Kind() == reflect.Struct {
			if subColNames, err := colNames(f, nil); err != nil {
				return []string{}, err
			} else {
				return subColNames, nil
			}
		} else {
			for f.Kind() == reflect.Ptr {
				f = f.Elem()
			}
			if f.Kind() == reflect.Struct {
				if subColNames, err := colNames(f, nil); err != nil {
					return []string{}, err
				} else {
					return subColNames, nil
				}
			}
		}
		return []string{t.Name}, nil
	}

	return []string{tag}, nil
}

func getConverters(col reflect.Value) ([]Converter, error) {
	var c []Converter
	l := col.Type().NumField()
	for i := 0; i < l; i++ {
		f := col.Field(i)
		tag := col.Type().Field(i).Tag.Get("csv")

		if converter, err := genConverter(f, tag); err != nil {
			return c, err
		} else {
			c = append(c, converter)
		}
	}

	return c, nil
}

func genConverter(f reflect.Value, tag string) (Converter, error) {
	var converter Converter
	var err error

	if tag == "-" {
		converter, err = NewEmptyConverter()
	} else {
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
	}

	return converter, err
}

func getOrders(col reflect.Value) []int {
	var orders []int
	l := col.Type().NumField()
	for i := 0; i < l; i++ {
		t := col.Type().Field(i)
		if t.Tag.Get("csv") == "-" {
			continue
		}

		tag := t.Tag.Get("csv-order")
		if tag != "" {
			if order, err := strconv.ParseInt(tag, 10, 0); err != nil {
				orders = append(orders, MAX_ORDER)
			} else {
				orders = append(orders, int(order))
			}
		} else {
			f := col.Field(i)
			if f.Kind() == reflect.Struct {
				orders = append(orders, getOrders(f)...)
			} else {
				for f.Kind() == reflect.Ptr {
					f = f.Elem()
				}
				if f.Kind() == reflect.Struct {
					orders = append(orders, getOrders(f)...)
				} else {
					orders = append(orders, MAX_ORDER)
				}
			}
		}
	}

	return orders
}

// encodeAll iterates over each item in data, encoder it then writes it
func (enc *encoder) encodeAll(data reflect.Value, converters []Converter, orders []int) error {
	n := data.Len()
	for i := 0; i < n; i++ {
		row, err := encodeRow(data.Index(i), converters)
		if err != nil {
			return err
		}

		newRow, err := reOrderRow(row, orders)
		if err != nil {
			return err
		}

		err = enc.Write(newRow)
		if err != nil {
			return err
		}
	}

	return nil
}

// reorder the row
func reOrderRow(row []string, orders []int) ([]string, error) {
	newRow := make([]string, len(row))
	tail := len(row) - 1
	for i := len(row) - 1; i >= 0; i-- {
		if orders[i] == MAX_ORDER {
			if tail < 0 || newRow[tail] != "" {
				return newRow, errors.New("order conflict, please check your order in tag")
			}
			newRow[tail] = row[i]
			tail -= 1
		} else {
			order := orders[i] - 1
			if order < 0 || order >= len(row) || newRow[order] != "" {
				return newRow, errors.New("order conflict, please check your order in tag")
			}
			newRow[order] = row[i]
		}
	}

	return newRow, nil
}

// encodes a struct into a CSV row
func encodeRow(v reflect.Value, converters []Converter) ([]string, error) {
	var row []string
	l := v.Type().NumField()

	for i := 0; i < l; i++ {
		f := v.Field(i)

		st := v.Type().Field(i).Tag
		if st.Get("csv") == "-" {
			continue
		}

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
