// Command cgoabi audits and rewrites architecture-sized integer types in
// cgo-exported function signatures. C int and unsigned int stay 32-bit on all
// supported targets, while Go int and uint follow the target pointer width.
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

const requiredGoVersion = "go1.26.5"

var writeFiles = flag.Bool("w", false, "rewrite files in place")

type intKind uint8

const (
	intNone intKind = iota
	intSigned
	intUnsigned
)

type abiType struct {
	kind  intKind
	stars int
}

type finding struct {
	file     string
	line     int
	fn       string
	part     string
	typeName string
}

func main() {
	flag.Parse()
	if runtime.Version() != requiredGoVersion {
		fmt.Fprintf(os.Stderr, "cgoabi must be built with exactly %s (got %q)\n", requiredGoVersion, runtime.Version())
		os.Exit(2)
	}
	paths := flag.Args()
	if len(paths) == 0 {
		paths = []string{"./legacy"}
	}
	findings, changed, err := run(paths, *writeFiles)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, f := range findings {
		fmt.Printf("%s:%d\t%s\t%s\t%s\n", f.file, f.line, f.fn, f.part, f.typeName)
	}
	if *writeFiles {
		fmt.Printf("rewrote %d file(s), %d ABI occurrence(s)\n", changed, len(findings))
	} else {
		fmt.Printf("found %d ABI occurrence(s)\n", len(findings))
	}
}

func run(paths []string, write bool) ([]finding, int, error) {
	files, err := goFiles(paths)
	if err != nil {
		return nil, 0, err
	}
	fset := token.NewFileSet()
	var all []finding
	changed := 0
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, changed, err
		}
		file, err := parser.ParseFile(fset, path, data, parser.ParseComments)
		if err != nil {
			return nil, changed, err
		}
		var fileFindings []finding
		needUnsafe := false
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil || !isCgoExport(fn) {
				continue
			}
			found, usesUnsafe, err := rewriteFunc(fset, path, fn, write)
			if err != nil {
				return nil, changed, err
			}
			fileFindings = append(fileFindings, found...)
			needUnsafe = needUnsafe || usesUnsafe
		}
		all = append(all, fileFindings...)
		if !write || len(fileFindings) == 0 {
			continue
		}
		_ = needUnsafe
		out, err := format.Source(renderFile(fset, file))
		if err != nil {
			return nil, changed, fmt.Errorf("format %s: %w", path, err)
		}
		if string(out) == string(data) {
			continue
		}
		if err := os.WriteFile(path, out, 0o644); err != nil {
			return nil, changed, err
		}
		changed++
	}
	return all, changed, nil
}

func renderFile(fset *token.FileSet, file *ast.File) []byte {
	var b strings.Builder
	if err := format.Node(&b, fset, file); err != nil {
		panic(err)
	}
	return []byte(b.String())
}

func goFiles(paths []string) ([]string, error) {
	set := make(map[string]bool)
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
				set[filepath.Clean(path)] = true
			}
			continue
		}
		err = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if p != path && (d.Name() == "vendor" || strings.HasPrefix(d.Name(), ".")) {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(p, ".go") && !strings.HasSuffix(p, "_test.go") {
				set[filepath.Clean(p)] = true
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	files := make([]string, 0, len(set))
	for path := range set {
		files = append(files, path)
	}
	sort.Strings(files)
	return files, nil
}

func isCgoExport(fn *ast.FuncDecl) bool {
	if fn.Doc == nil {
		return false
	}
	want := "//export " + fn.Name.Name
	for _, c := range fn.Doc.List {
		if strings.TrimSpace(c.Text) == want {
			return true
		}
	}
	return false
}

func rewriteFunc(fset *token.FileSet, path string, fn *ast.FuncDecl, write bool) ([]finding, bool, error) {
	var out []finding
	usesUnsafe := false
	var prefix []ast.Stmt
	argIndex := 0
	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			abi, ok := inspectABIType(field.Type)
			if !ok {
				argIndex += max(1, len(field.Names))
				continue
			}
			pos := fset.Position(field.Type.Pos())
			out = append(out, finding{file: path, line: pos.Line, fn: fn.Name.Name, part: "parameter", typeName: abi.String()})
			if !write {
				argIndex += max(1, len(field.Names))
				continue
			}
			if abi.stars > 1 {
				return nil, false, fmt.Errorf("%s:%d: %s: pointer depth %d is not supported", path, pos.Line, fn.Name.Name, abi.stars)
			}
			if len(field.Names) == 0 {
				field.Names = []*ast.Ident{identAt(fmt.Sprintf("arg%d_cgo", argIndex), field.Type.Pos())}
				field.Type = abi.fixedExprAt(field.Type.Pos())
				argIndex++
				continue
			}
			field.Type = abi.fixedExprAt(field.Type.Pos())
			for _, name := range field.Names {
				if name.Name == "_" {
					continue
				}
				orig := name.Name
				name.Name = uniqueParamName(fn, orig+"_cgo")
				if identUsed(fn.Body, orig) {
					prefix = append(prefix, abi.paramConversions(orig, name.Name, fn.Body.Lbrace+1)...)
				}
			}
			usesUnsafe = usesUnsafe || abi.stars != 0
			argIndex += len(field.Names)
		}
	}

	results := flattenFields(fn.Type.Results)
	resultABI := make([]*abiType, len(results))
	for i, field := range results {
		abi, ok := inspectABIType(field.Type)
		if !ok {
			continue
		}
		copyABI := abi
		resultABI[i] = &copyABI
		pos := fset.Position(field.Type.Pos())
		out = append(out, finding{file: path, line: pos.Line, fn: fn.Name.Name, part: "result", typeName: abi.String()})
		usesUnsafe = usesUnsafe || abi.stars != 0
		if write && abi.stars != 0 {
			return nil, false, fmt.Errorf("%s:%d: %s: pointer results require a function-specific ownership adapter", path, pos.Line, fn.Name.Name)
		}
	}
	if !write || len(out) == 0 {
		return out, usesUnsafe, nil
	}

	if fn.Type.Results != nil {
		for _, field := range fn.Type.Results.List {
			abi, ok := inspectABIType(field.Type)
			if !ok {
				continue
			}
			field.Type = abi.fixedExprAt(field.Type.Pos())
			for _, name := range field.Names {
				if name.Name == "_" {
					continue
				}
				orig := name.Name
				name.Name = uniqueParamName(fn, orig+"_cgo")
				prefix = append(prefix, abi.namedResultDecl(orig, fn.Body.Lbrace+1))
			}
		}
	}
	if err := rewriteReturns(fn.Body, results, resultABI); err != nil {
		return nil, false, fmt.Errorf("%s:%d: %s: %w", path, fset.Position(fn.Pos()).Line, fn.Name.Name, err)
	}
	if len(prefix) != 0 {
		fn.Body.List = append(prefix, fn.Body.List...)
	}
	return out, usesUnsafe, nil
}

func flattenFields(list *ast.FieldList) []*ast.Field {
	if list == nil {
		return nil
	}
	var out []*ast.Field
	for _, field := range list.List {
		n := len(field.Names)
		if n == 0 {
			n = 1
		}
		for range n {
			out = append(out, field)
		}
	}
	return out
}

func rewriteReturns(body *ast.BlockStmt, results []*ast.Field, kinds []*abiType) error {
	var rewriteErr error
	ast.Inspect(body, func(node ast.Node) bool {
		if rewriteErr != nil {
			return false
		}
		if _, ok := node.(*ast.FuncLit); ok {
			return false
		}
		ret, ok := node.(*ast.ReturnStmt)
		if !ok {
			return true
		}
		if len(ret.Results) == 0 {
			var exprs []ast.Expr
			for i, field := range results {
				if len(field.Names) == 0 {
					rewriteErr = fmt.Errorf("naked return with unnamed result")
					return false
				}
				name := field.Names[0].Name
				if kinds[i] != nil {
					exprs = append(exprs, kinds[i].resultConversion(identAt(strings.TrimSuffix(name, "_cgo"), ret.Return)))
				} else {
					exprs = append(exprs, identAt(name, ret.Return))
				}
			}
			ret.Results = exprs
			return true
		}
		if len(ret.Results) != len(results) {
			rewriteErr = fmt.Errorf("return has %d expression(s) for %d result(s)", len(ret.Results), len(results))
			return false
		}
		for i, abi := range kinds {
			if abi != nil {
				ret.Results[i] = abi.resultConversion(ret.Results[i])
			}
		}
		return true
	})
	return rewriteErr
}

func inspectABIType(expr ast.Expr) (abiType, bool) {
	stars := 0
	for {
		star, ok := expr.(*ast.StarExpr)
		if !ok {
			break
		}
		stars++
		expr = star.X
	}
	id, ok := expr.(*ast.Ident)
	if !ok {
		return abiType{}, false
	}
	switch id.Name {
	case "int":
		return abiType{kind: intSigned, stars: stars}, true
	case "uint":
		return abiType{kind: intUnsigned, stars: stars}, true
	default:
		return abiType{}, false
	}
}

func (t abiType) String() string {
	base := "int"
	if t.kind == intUnsigned {
		base = "uint"
	}
	return strings.Repeat("*", t.stars) + base
}

func (t abiType) fixedExprAt(pos token.Pos) ast.Expr {
	base := "int32"
	if t.kind == intUnsigned {
		base = "uint32"
	}
	var out ast.Expr = identAt(base, pos)
	for range t.stars {
		out = &ast.StarExpr{Star: pos, X: out}
	}
	return out
}

func (t abiType) nativeExprAt(pos token.Pos) ast.Expr {
	base := "int"
	if t.kind == intUnsigned {
		base = "uint"
	}
	var out ast.Expr = identAt(base, pos)
	for range t.stars {
		out = &ast.StarExpr{Star: pos, X: out}
	}
	return out
}

func (t abiType) paramConversions(dst, src string, pos token.Pos) []ast.Stmt {
	if t.stars == 0 {
		rhs := &ast.CallExpr{Fun: t.nativeExprAt(pos), Lparen: pos, Args: []ast.Expr{identAt(src, pos)}, Rparen: pos}
		return []ast.Stmt{&ast.AssignStmt{
			Lhs: []ast.Expr{identAt(dst, pos)}, TokPos: pos, Tok: token.DEFINE, Rhs: []ast.Expr{rhs},
		}}
	}
	finish := dst + "_cgo_finish"
	helper := "cgoABIIntPtr"
	if t.kind == intUnsigned {
		helper = "cgoABIUintPtr"
	}
	return []ast.Stmt{
		&ast.AssignStmt{
			Lhs:    []ast.Expr{identAt(dst, pos), identAt(finish, pos)},
			TokPos: pos,
			Tok:    token.DEFINE,
			Rhs: []ast.Expr{&ast.CallExpr{
				Fun:    identAt(helper, pos),
				Lparen: pos,
				Args:   []ast.Expr{identAt(src, pos)},
				Rparen: pos,
			}},
		},
		&ast.DeferStmt{
			Defer: pos,
			Call: &ast.CallExpr{
				Fun:    identAt(finish, pos),
				Lparen: pos,
				Rparen: pos,
			},
		},
	}
}

func (t abiType) namedResultDecl(name string, pos token.Pos) ast.Stmt {
	return &ast.DeclStmt{Decl: &ast.GenDecl{
		TokPos: pos,
		Tok:    token.VAR,
		Specs:  []ast.Spec{&ast.ValueSpec{Names: []*ast.Ident{identAt(name, pos)}, Type: t.nativeExprAt(pos)}},
	}}
}

func (t abiType) resultConversion(expr ast.Expr) ast.Expr {
	pos := expr.Pos()
	if t.stars == 0 {
		return &ast.CallExpr{Fun: t.fixedExprAt(pos), Lparen: pos, Args: []ast.Expr{expr}, Rparen: pos}
	}
	return &ast.CallExpr{
		Fun:    t.fixedExprAt(pos),
		Lparen: pos,
		Args: []ast.Expr{&ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   identAt("unsafe", pos),
				Sel: identAt("Pointer", pos),
			},
			Lparen: pos,
			Args:   []ast.Expr{expr},
			Rparen: pos,
		}},
		Rparen: pos,
	}
}

func identAt(name string, pos token.Pos) *ast.Ident {
	return &ast.Ident{NamePos: pos, Name: name}
}

func identUsed(node ast.Node, name string) bool {
	used := false
	ast.Inspect(node, func(node ast.Node) bool {
		if id, ok := node.(*ast.Ident); ok && id.Name == name {
			used = true
			return false
		}
		return !used
	})
	return used
}

func uniqueParamName(fn *ast.FuncDecl, base string) string {
	used := make(map[string]bool)
	ast.Inspect(fn, func(node ast.Node) bool {
		if id, ok := node.(*ast.Ident); ok {
			used[id.Name] = true
		}
		return true
	})
	if !used[base] {
		return base
	}
	for i := 2; ; i++ {
		name := base + strconv.Itoa(i)
		if !used[name] {
			return name
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
