package csv

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type FloatConverter struct {
	colName     string
	conditions  []string
	targetVals  []float64
	convertVals []string
}

func NewFloatConverter(tag string) (FloatConverter, error) {
	var c FloatConverter
	cons := strings.Split(tag, ";")
	c.colName = cons[0]
	if len(cons) >= 2 && c.colName != "-" {
		cons = cons[1:]
		index := 0
		for _, con := range cons {
			mapCon := strings.Split(con, ":")
			if len(mapCon) == 3 {
				if targetVal, err := strconv.ParseFloat(mapCon[1], 10); err != nil {
					return c, errors.New(fmt.Sprintf("cant parse %s to float", mapCon[1]))
				} else {
					c.conditions = append(c.conditions, mapCon[0])
					c.targetVals = append(c.targetVals, targetVal)
					c.convertVals = append(c.convertVals, mapCon[2])
					index++
				}
			} else if len(mapCon) == 2 {
				if mapCon[0] == "$default" {
					c.conditions = append(c.conditions, mapCon[0])
					c.targetVals = append(c.targetVals, 0.0)
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

func (c FloatConverter) Covert(fv reflect.Value) ([]string, error) {
	if c.colName == "-" {
		return []string{}, nil
	}
	rawVal := fv.Float()
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
		case "$gte":
			if rawVal >= c.targetVals[i] {
				return []string{c.convertVals[i]}, nil
			}
		case "$lte":
			if rawVal <= c.targetVals[i] {
				return []string{c.convertVals[i]}, nil
			}
		case "$default":
			return []string{c.convertVals[i]}, nil
		default:
			continue
		}
	}
	if fv.Kind() == reflect.Float32 {
		return []string{strconv.FormatFloat(rawVal, 'g', -1, 32)}, nil
	}
	return []string{strconv.FormatFloat(rawVal, 'g', -1, 64)}, nil
}
