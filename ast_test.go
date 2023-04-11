package ast

import (
	"testing"
)

var filePath = "./test/person.go"

func TestAstHelper_GetFileDesc(t *testing.T) {

	h := NewAstHelper(filePath)
	pk, err := h.GetFileDesc()
	if err != nil {
		t.Logf("err:%v", err)
		return
	}
	t.Logf("pk:%v", pk)
}
func TestAstHelper_GetPackage(t *testing.T) {

	h := NewAstHelper(filePath)
	pk, err := h.getPackage()
	if err != nil {
		t.Logf("err:%v", err)
		return
	}
	t.Logf("pk:%v", pk)
}
func TestAstHelper_GetImports(t *testing.T) {

	h := NewAstHelper(filePath)
	res, err := h.getImports()
	if err != nil {
		t.Logf("err:%v", err)
		return
	}
	t.Logf("res:%v", res)
}
func TestAstHelper_GetStructDescs(t *testing.T) {

	h := NewAstHelper(filePath)
	res, err := h.getStructDescs()
	if err != nil {
		t.Logf("err:%v", err)
		return
	}
	t.Logf("res:%v", res)
}
