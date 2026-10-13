package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// Competitive player initialization reads EngineAdmin to award the normal
// multiplayer spells and abilities. The standalone server bypasses startServer;
// it must establish the same permissions, without enabling invulnerability.
func TestDedicatedServerUsesMultiplayerSpellPermissions(t *testing.T) {
	for _, entry := range []struct{ file, name, boundary string }{
		{"game.go", "startServer", "nox_xxx_parseGamedataBinPre_4D1630"},
		{"gameloop.go", "MainLoop", "newSession"},
	} {
		t.Run(entry.name, func(t *testing.T) {
			file, err := parser.ParseFile(token.NewFileSet(), entry.file, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			var body *ast.BlockStmt
			for _, decl := range file.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == entry.name {
					body = fn.Body
				}
			}
			if body == nil {
				t.Fatal("server entry missing")
			}
			var admin, god, boundary token.Pos
			ast.Inspect(body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				if fn, ok := call.Fun.(*ast.Ident); ok && fn.Name == entry.boundary {
					boundary = call.Pos()
				}
				fn, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				if fn.Sel.Name == entry.boundary {
					boundary = call.Pos()
				}
				pkg, ok := fn.X.(*ast.Ident)
				if !ok || pkg.Name != "noxflags" || len(call.Args) != 1 {
					return true
				}
				flag, ok := call.Args[0].(*ast.SelectorExpr)
				if !ok {
					return true
				}
				flagPkg, ok := flag.X.(*ast.Ident)
				if !ok || flagPkg.Name != "noxflags" {
					return true
				}
				if fn.Sel.Name == "SetEngine" && flag.Sel.Name == "EngineAdmin" {
					admin = call.Pos()
				}
				if fn.Sel.Name == "UnsetEngine" && flag.Sel.Name == "EngineGodMode" {
					god = call.Pos()
				}
				return true
			})
			if admin == token.NoPos || god == token.NoPos {
				t.Fatal("server entry must enable ordinary multiplayer spell permissions and clear god mode")
			}
			if boundary == token.NoPos || admin >= boundary || god >= boundary {
				t.Fatal("multiplayer permissions must precede player initialization")
			}
		})
	}
}
