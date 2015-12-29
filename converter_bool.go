package csv

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type BoolCoverter struct {
	colName  string
	trueVal  string
	falseVal string
}

func NewBoolConverter(tag string) (BoolCoverter, error) {
	var c BoolCoverter
	c.trueVal = "-"
	c.falseVal = "-"
	cons := strings.Split(tag, ";")
	c.colName = cons[0]
	if len(cons) >= 2 && c.colName != "-" {
		cons = cons[1:]
		for _, con := range cons {
			mapCon := strings.Split(con, ":")
			if len(mapCon) != 2 {
				return c, errors.New("condition should be 'true/false:target'")
			} else {
				if strings.EqualFold(mapCon[0], "true") {
					c.trueVal = mapCon[1]
				} else if strings.EqualFold(mapCon[0], "false") {
					c.falseVal = mapCon[1]
				} else {
					return c, errors.New("condition should be 'true/false:target'")
				}
			}
		}
	}
	return c, nil
}

func (c BoolCoverter) Covert(fv reflect.Value) ([]string, error) {
	if c.colName == "-" {
		return []string{}, nil
	}
	rawVal := fv.Bool()
	if rawVal && c.trueVal != "-" {
		return []string{c.trueVal}, nil
	}
	if !rawVal && c.falseVal != "-" {
		return []string{c.falseVal}, nil
	}
	return []string{fmt.Sprintf("%t", rawVal)}, nil
}
