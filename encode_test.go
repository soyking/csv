package csv

import (
	"fmt"
	"testing"
)

type TypeA struct {
	Name    string `csv:"Na"`
	Phone   int    `csv:"ph"`
	Age     *int   `csv:"aGe"`
	Married bool   `csv:"ma" true:"yes" false:"no"`
	B       *TypeB
}

type TypeB struct {
	House string `csv:"-"`
	Tele  string
}

// func (b TypeB) MarshalCSV() ([]string, error) {
// 	return []string{b.House, b.Tele}, nil
// }

func TestMarshal(t *testing.T) {
	var aSlice = []TypeA{TypeA{Name: "soy", Phone: 123, B: &TypeB{"west", "010-56"}}}
	a := 9
	var age = &a
	aSlice[0].Age = age
	r, err := Marshal(aSlice)
	if err != nil {
		t.Error(err)
	} else {
		fmt.Println(string(r))
	}
}
