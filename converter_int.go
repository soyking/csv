package csv

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type IntConverter struct {
	colName     string
	conditions  []string
	targetVals  []int64
	convertVals []string
}

func NewIntConverter(tag string) (IntConverter, error) {
	var c IntConverter
	cons := strings.Split(tag, ";")
	c.colName = cons[0]
	if len(cons) >= 2 && c.colName != "-" {
		cons = cons[1:]
		index := 0
		for _, con := range cons {
			mapCon := strings.Split(con, ":")
			if len(mapCon) == 3 {
				if targetVal, err := strconv.ParseInt(mapCon[1], 10, 0); err != nil {
					return c, errors.New(fmt.Sprintf("cant parse %s to int", mapCon[1]))
				} else {
					c.conditions = append(c.conditions, mapCon[0])
					c.targetVals = append(c.targetVals, targetVal)
					c.convertVals = append(c.convertVals, mapCon[2])
					index++
				}
			} else if len(mapCon) == 2 {
				if mapCon[0] == "$default" {
					c.conditions = append(c.conditions, mapCon[0])
					c.targetVals = append(c.targetVals, 0)
					c.convertVals = append(c.convertVals, mapCon[1])
					index++
				} else {
					return c, errors.New("only support 'condition:target:convert' or '$default:val'")
				}
			} else {
				return c, errors.New("only support 'condition:target:convert' or '$default:val'")
			}
		}
	}
	return c, nil
}

func (c IntConverter) Covert(fv reflect.Value) ([]string, error) {
	if c.colName == "-" {
		return []string{}, nil
	}
	rawVal := fv.Int()
	for i := range c.conditions {
		switch c.conditions[i] {
		case "$eq":
			if rawVal == c.targetVals[i] {
				return []string{c.convertVals[i]}, nil
			}
		case "$ne":
			if rawVal != c.targetVals[i] {
				return []string{c.convertVals[i]}, nil
			}
		case "$gt":
			if rawVal > c.targetVals[i] {
				return []string{c.convertVals[i]}, nil
			}
		case "$lt":
			if rawVal < c.targetVals[i] {
				return []string{c.convertVals[i]}, nil
			}
		case "$default":
			return []string{c.convertVals[i]}, nil
		default:
			continue
		}
	}
	return []string{strconv.FormatInt(rawVal, 10)}, nil
}
