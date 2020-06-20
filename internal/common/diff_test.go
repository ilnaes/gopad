package common

import (
	"fmt"
	"testing"
)

func TestApply(t *testing.T) {
	s1 := []byte("caeqwhdoqi")
	s2 := []byte("scqoid")

	ops := diff(s1, s2)

	s3 := make([]byte, len(s1))
	copy(s3, s1)

	s3 = apply(s3, ops)

	if string(s3) != string(s2) {
		t.Error(fmt.Sprintf("%s should be %s", s3, s1))
	}
}

func TestXform(t *testing.T) {
	s1 := []byte("sad")
	s2 := []byte("esad")
	s3 := []byte("ade")

	op1 := diff(s1, s2)
	op2 := diff(s1, s3)

	op2 = xform(op1, op2)

	s4 := append([]byte{}, s1...)
	s4 = apply(s4, op1)
	s4 = apply(s4, op2)

	if string(s4) != "eade" {
		t.Error(fmt.Sprintf("%s should be eade", s4))
	}
}
