package csv

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"reflect"
	"strconv"
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
		if data.Kind() == reflect.Ptr {
			data = data.Elem()
			if data.Kind() != reflect.Slice {
				return []byte{}, errors.New("only slices can be marshalled")
			}
		} else {
			return []byte{}, errors.New("only slices can be marshalled")
		}
	}

	el := data.Index(0)
	enc, err := newEncoder(el)

	if err != nil {
		return []byte{}, err
	}

	err = enc.encodeAll(data)

	if err != nil {
		return []byte{}, err
	}

	enc.Flush()
	return enc.buffer.Bytes(), nil
}

func newEncoder(el reflect.Value) (*encoder, error) {
	b := bytes.NewBuffer([]byte{})

	enc := &encoder{
		buffer: b,
		Writer: csv.NewWriter(b),
	}

	err := enc.Write(colNames(el))

	return enc, err
}

// colNames takes a struct and returns the computed columns names for each
// field.
func colNames(v reflect.Value) (out []string) {
	l := v.Type().NumField()

	for x := 0; x < l; x++ {
		f := v.Field(x)
		t := v.Type().Field(x)
		h, ok := fieldHeaderName(t, f)
		if ok {
			out = append(out, h...)
		}
	}

	return
}

// fieldHeaderName returns the header name to use for the given StructField
// This can be a user defined name (via the Tag) or a default name.
func fieldHeaderName(t reflect.StructField, f reflect.Value) ([]string, bool) {
	tag := t.Tag.Get("csv")
	if tag == "-" {
		return []string{}, false
	}

	// If there is no tag set, use a default name
	if tag == "" {
		if f.Kind() == reflect.Struct || f.Kind() == reflect.Interface {
			return colNames(f), true
		} else if f.Kind() == reflect.Ptr {
			f = f.Elem()
			return colNames(f), true
		}
		return []string{t.Name}, true
	}

	return []string{tag}, true
}

// encodeAll iterates over each item in data, encoder it then writes it
func (enc *encoder) encodeAll(data reflect.Value) error {
	n := data.Len()
	for c := 0; c < n; c++ {
		row, err := enc.encodeRow(data.Index(c))
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
func (enc *encoder) encodeRow(v reflect.Value) ([]string, error) {

	var row []string
	// TODO env.columns should map to a cfield
	// iterate over each cfield and encode with it
	l := v.Type().NumField()

	for x := 0; x < l; x++ {
		fv := v.Field(x)
		st := v.Type().Field(x).Tag

		if st.Get("csv") == "-" {
			continue
		}
		o, err := enc.encodeCol(fv, st)
		if err != nil {
			return []string{}, err
		}
		row = append(row, o...)
	}

	return row, nil
}

// Returns the string representation of the field value
func (enc *encoder) encodeCol(fv reflect.Value, st reflect.StructTag) ([]string, error) {
	switch fv.Kind() {
	case reflect.String:
		return []string{fv.String()}, nil
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
		return []string{fmt.Sprintf("%v", fv.Int())}, nil
	case reflect.Float32:
		return []string{encodeFloat(32, fv)}, nil
	case reflect.Float64:
		return []string{encodeFloat(64, fv)}, nil
	case reflect.Bool:
		return []string{encodeBool(fv.Bool(), st)}, nil
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		return []string{fmt.Sprintf("%v", fv.Uint())}, nil
	case reflect.Complex64, reflect.Complex128:
		return []string{fmt.Sprintf("%+.3g", fv.Complex())}, nil
	case reflect.Interface:
		return encodeInterface(enc, fv)
	case reflect.Struct:
		return encodeInterface(enc, fv)
	case reflect.Ptr:
		fv = fv.Elem()
		fmt.Println("encode col ptr")
		return enc.encodeCol(fv, st)
	default:
		return []string{}, errors.New(fmt.Sprintf("Unsupported type %s", fv.Kind()))
	}

	return []string{}, nil
}

func encodeFloat(bits int, f reflect.Value) string {
	return strconv.FormatFloat(f.Float(), 'g', -1, bits)
}

func encodeBool(b bool, st reflect.StructTag) string {
	v := strconv.FormatBool(b)
	tv := st.Get(v)

	if tv != "" {
		return tv
	}
	return v
}

func encodeInterface(enc *encoder, fv reflect.Value) ([]string, error) {
	marshalerType := reflect.TypeOf(new(Marshaler)).Elem()

	if fv.Type().Implements(marshalerType) {
		m := fv.Interface().(Marshaler)
		return m.MarshalCSV()
	}

	return enc.encodeRow(fv)
}
