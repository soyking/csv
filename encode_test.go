package csv

import (
	"fmt"
	"testing"
)

type TypeA struct {
	A    string  `csv:"-"`
	AA   string  `csv:"aa;abc:def"`
	AAA  string  `csv:"aaa;abc:def;$default:字符串"`
	B    int     `csv:"-"`
	BB   int     `csv:"bb;$eq:1:男"`
	BBB  int     `csv:"bbb;$gt:1:女"`
	BBBB int     `csv:"bbbb;$default:整数"`
	C    float32 `csv:"-"`
	CC   float32 `csv:"cc;$eq:1.1:man"`
	CCC  float32 `csv:"ccc;$lt:6.6:woman"`
	CCCC float32 `csv:"cccc;$default:浮点数"`
	D    bool    `csv:"-"`
	DD   bool    `csv:"dd;true:是的呢;false:不是呢"`
	DDD  bool    ``
	E    uint    `csv:"-"`
	EE   uint    `csv:"ee;$eq:1:男"`
	EEE  uint    `csv:"eee;$gt:1:女"`
	EEEE uint    `csv:"eeee;$default:短整数"`
	F    TypeB   `csv:"-"`
	FF   TypeB
	FFF  *TypeB
	G    TypeC
}

type TypeB struct {
	s string `csv:"ss;abc:def"`
	i int    `csv:"ii;$eq:2:is2"`
}

type TypeC struct {
	s string `csv:"string"`
	i int    `csv:"int"`
}

func (c TypeC) MarshalCSV() ([]string, error) {
	return []string{c.s, fmt.Sprintf("is %d", c.i)}, nil
}

func TestConverter(t *testing.T) {
	aSlice := []TypeA{
		TypeA{
			A:   "not show",
			AA:  "abc",
			AAA: "ab",
			B:   0,
			BB:  1,
			BBB: 2,
			C:   0.0,
			CC:  1.1,
			CCC: 5.0,
			D:   true,
			DD:  false,
			DDD: true,
			E:   0,
			EE:  1,
			EEE: 0,
			F:   TypeB{s: "abc", i: 2},
			FF:  TypeB{s: "abc", i: 2},
			FFF: &TypeB{s: "abc", i: 2},
			G:   TypeC{"fun", 123},
		},
	}

	r, err := Marshal(aSlice)
	if err != nil {
		t.Error(err)
	} else {
		fmt.Println(string(r))
	}
}
