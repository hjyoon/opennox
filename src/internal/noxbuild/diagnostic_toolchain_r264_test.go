package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Every diagnostic entry point must agree with the repository's one text pin;
// passing package unit tests alone does not execute a command's main guard.
func TestDiagnosticToolchainPolicyR264(t *testing.T) {
	data, err := os.ReadFile("../../../toolchain/go-version.txt")
	if err != nil {
		t.Fatal(err)
	}
	want := strings.TrimSpace(string(data))
	if want == "" {
		t.Fatal("empty repository toolchain policy")
	}
	for _, tool := range []string{"cgoabi", "layoutaudit", "noxbuild"} {
		t.Run(tool, func(t *testing.T) {
			path := filepath.Join("..", tool, "main.go")
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			var values []string
			for _, declaration := range file.Decls {
				group, ok := declaration.(*ast.GenDecl)
				if !ok || group.Tok != token.CONST {
					continue
				}
				for _, spec := range group.Specs {
					value := spec.(*ast.ValueSpec)
					for index, name := range value.Names {
						if name.Name != "requiredGoVersion" {
							continue
						}
						if len(value.Values) != len(value.Names) {
							t.Fatal("nonliteral version policy")
						}
						literal, ok := value.Values[index].(*ast.BasicLit)
						if !ok || literal.Kind != token.STRING {
							t.Fatal("non-string version policy")
						}
						version, err := strconv.Unquote(literal.Value)
						if err != nil {
							t.Fatal(err)
						}
						values = append(values, version)
					}
				}
			}
			if len(values) != 1 || values[0] != want {
				t.Fatalf("%s requires %q; repository policy is %q", tool, values, want)
			}
		})
	}
}
