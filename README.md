CSV Marshaller UnMarshaller for Go
==================================

This is a work in progress.  The API should not change, but until this is used
in active production I make no promises.

[Docs](https://godoc.org/github.com/jweir/csv)

-

在以上项目基础上，只保留了 `Marshal` 这个函数。

添加的功能：

- `csv`的 tag 中可以写条件，实现对结果的转化，这其中每个不同类型的字段（`int` `float32` 等）用不同的`Converter`进行转换
- 在`struct`中存在`struct`字段时会将数据平铺出来，而不是单独一个条目；`struct`会递归生成，如果是指针会访问实际内容
- 添加`csv-order`的 tag ，控制 csv 输出的条目顺序，顺序从1开始，不填会按照字段正常顺序输出；顺序值有冲突会返回错误

例子：

```
package csv

import (
	"fmt"
	"testing"
)

type TypeA struct {
	A    string  `csv:"-"`
	AA   string  `csv:"aa;abc:def" csv-order:"2"`
	AAA  string  `csv:"aaa;abc:def;$default:字符串"`
	B    int     `csv:"-"`
	BB   int     `csv:"bb;$eq:1:男" csv-order:"1"`
	BBB  int     `csv:"bbb;$gt:1:女"`
	BBBB int     `csv:"bbbb;$default:整数" csv-order:"3"`
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
	G    TypeC
}

type TypeB struct {
	s string `csv:"ss;abc:def" csv-order:"4"`
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

```
输出：（作了适当调整）

```
bb, aa,  bbbb, ss,  aaa,   bbb, cc,  ccc,   cccc,  dd,     DDD,  ee, eee, eeee,  ii,  string, int
男, def, 整数,  def, 字符串, 女,  1.1,  woman, 浮点数, 不是呢, true, 男,  0,   短整数, is2, fun,    is 123
```