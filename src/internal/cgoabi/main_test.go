package main

import (
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestRewriteFunc(t *testing.T) {
	const src = `package legacy

import "unsafe"

//export scalar
func scalar(a, b int, p *uint) int {
	if a == 0 { return b }
	return a + int(*p)
}

//export named
func named(a uint) (out int) {
	out = int(a)
	return
}

//export noResult
func noResult(a int) {}
`
	want := `package legacy

import "unsafe"

//export scalar
func scalar(a_cgo, b_cgo int32, p_cgo *uint32) int32 {
	a := int(a_cgo)
	b := int(b_cgo)
	p := (*uint)(unsafe.Pointer(p_cgo))
	if a == 0 {
		return int32(b)
	}
	return int32(a + int(*p))
}

//export named
func named(a_cgo uint32) (out_cgo int32) {
	a := uint(a_cgo)
	var out int
	out = int(a)
	return int32(out)
}

//export noResult
func noResult(a_cgo int32) {}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || !isCgoExport(fn) {
			continue
		}
		if _, _, err := rewriteFunc(fset, "test.go", fn, true); err != nil {
			t.Fatal(err)
		}
	}
	var out strings.Builder
	if err := format.Node(&out, fset, file); err != nil {
		t.Fatal(err)
	}
	if out.String() != want {
		t.Fatalf("rewrite mismatch\n--- got ---\n%s\n--- want ---\n%s", out.String(), want)
	}
}

func TestInspectABIType(t *testing.T) {
	tests := []struct {
		src  string
		want string
		ok   bool
	}{
		{"int", "int", true},
		{"uint", "uint", true},
		{"**int", "**int", true},
		{"int32", "", false},
		{"C.int", "", false},
	}
	for _, tc := range tests {
		expr, err := parser.ParseExpr(tc.src)
		if err != nil {
			t.Fatal(err)
		}
		got, ok := inspectABIType(expr)
		if ok != tc.ok || ok && got.String() != tc.want {
			t.Errorf("inspectABIType(%s) = (%q, %v), want (%q, %v)", tc.src, got.String(), ok, tc.want, tc.ok)
		}
	}
}
