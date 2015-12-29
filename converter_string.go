package csv

import (
	"errors"
	"reflect"
	"strings"
)

// ------ convert string type to set in tag of `csv`
type StringCoverter struct {
	colName    string
	conditions map[string]string
}

func NewStringConverter(tag string) (StringCoverter, error) {
	var c StringCoverter
	c.conditions = make(map[string]string)
	cons := strings.Split(tag, ";")
	c.colName = cons[0]
	if len(cons) >= 2 && c.colName != "-" {
		cons = cons[1:]
		for _, con := range cons {
			mapCon := strings.Split(con, ":")
			if len(mapCon) != 2 {
				return c, errors.New("condition should be 'condition:target'")
			} else {
				c.conditions[mapCon[0]] = mapCon[1]
			}
		}
	}
	return c, nil
}

func (c StringCoverter) Covert(fv reflect.Value) ([]string, error) {
	if c.colName == "-" {
		return []string{}, nil
	}
	rawVal := fv.String()
	if convertVal, ok := c.conditions[rawVal]; ok {
		return []string{convertVal}, nil
	}
	if convertVal, ok := c.conditions["$default"]; ok {
		return []string{convertVal}, nil
	}
	return []string{rawVal}, nil
}
