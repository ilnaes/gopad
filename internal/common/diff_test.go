package common

import (
	"fmt"
	"testing"
)

func stringApply(s1, s2 string) (string, []Op) {
	b1 := []byte(s1)
	op := diff(b1, []byte(s2))

	res := append([]byte{}, b1...)
	res = apply(res, op)

	return string(res), op
}

func stringXform(s1, s2, s3 string) (string, [][]Op) {
	b1 := []byte(s1)
	op1 := diff(b1, []byte(s2))
	op2 := diff(b1, []byte(s3))

	res := append([]byte{}, b1...)

	res = apply(res, op1)
	op3 := xform(op1, op2)
	res = apply(res, op3)

	ops := append([][]Op{}, op1)
	ops = append(ops, op2)
	ops = append(ops, op3)

	return string(res), ops
}

func TestXform(t *testing.T) {
	cases := [][]string{
		{"sad", "esad", "ade", "eade"},
		{"abcdef", "adf", "acf", "af"},
		{"abcdef", "acf", "adf", "af"},
		{"abc123456def", "abc456def", "abc12def", "abcdef"},
		{"abc123456def", "abc6def", "abcdef", "abcdef"},
		{"start: long string :end", "start: :end", "start: long text", "start:ext"},
	}

	for _, c := range cases {
		if res, _ := stringApply(c[0], c[1]); res != c[1] {
			t.Error(fmt.Sprintf("APPLY: %s should be %s", res, c[1]))
		}
		if res, _ := stringApply(c[0], c[2]); res != c[2] {
			t.Error(fmt.Sprintf("APPLY: %s should be %s", res, c[2]))
		}
		if res, ops := stringXform(c[0], c[1], c[2]); res != c[3] {
			t.Error(fmt.Sprintf("XFORM: %s should be %s\nOP1: %v\nOP2: %v\nOP3: %v", res, c[3], ops[0], ops[1], ops[2]))
		}
	}
}
