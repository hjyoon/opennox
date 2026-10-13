package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// RunServerArgs enters MainLoop directly, without the client menu's game-data
// initialization. Keep the dedicated entry wired to the same balance, ability
// and monster-data service before newSession can start accepting players.
// The separate real-clock headless diagnostic exercises the actual network
// join, player initialization and movement without changing game state.
func TestDedicatedServerInitializesGameDataBeforeAcceptingPlayers(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "gameloop.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var body *ast.BlockStmt
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "MainLoop" {
			body = fn.Body
		}
	}
	if body == nil {
		t.Fatal("dedicated server entry missing")
	}
	var initGuard *ast.IfStmt
	var session token.Pos
	ast.Inspect(body, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			if fn, ok := call.Fun.(*ast.SelectorExpr); ok && fn.Sel.Name == "newSession" {
				session = call.Pos()
			}
		}
		guard, ok := node.(*ast.IfStmt)
		if !ok {
			return true
		}
		cond, ok := guard.Cond.(*ast.BinaryExpr)
		if !ok || cond.Op != token.EQL {
			return true
		}
		call, ok := cond.X.(*ast.CallExpr)
		if !ok {
			return true
		}
		fn, ok := call.Fun.(*ast.Ident)
		zero, isZero := cond.Y.(*ast.BasicLit)
		if ok && fn.Name == "nox_xxx_parseGamedataBinPre_4D1630" && isZero && zero.Value == "0" {
			initGuard = guard
		}
		return true
	})
	if initGuard == nil {
		t.Fatal("dedicated entry skips game data: new players retain zero health, mana and speed")
	}
	if session == token.NoPos || initGuard.End() >= session {
		t.Fatal("game data must be initialized before the server accepts players")
	}
	if len(initGuard.Body.List) != 1 {
		t.Fatal("game-data failure must abort startup")
	}
	ret, ok := initGuard.Body.List[0].(*ast.ReturnStmt)
	if !ok || len(ret.Results) != 1 {
		t.Fatal("game-data failure must return an error")
	}
	call, ok := ret.Results[0].(*ast.CallExpr)
	if !ok {
		t.Fatal("game-data failure must not return nil")
	}
	fn, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || fn.Sel.Name != "New" {
		t.Fatal("game-data failure must return an explicit error")
	}
	pkg, ok := fn.X.(*ast.Ident)
	if !ok || pkg.Name != "errors" {
		t.Fatal("game-data failure must return an explicit error")
	}
}
