package opennox

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestServerQuitPendingAckCount(t *testing.T) {
	for _, mode := range []struct {
		name      string
		dedicated bool
	}{{"listen_host", false}, {"dedicated", true}} {
		t.Run(mode.name, func(t *testing.T) {
			for players := 0; players <= 32; players++ {
				t.Run(fmt.Sprintf("players_%02d", players), func(t *testing.T) {
					want := players
					if !mode.dedicated {
						// Preserve the original local-host subtraction, including
						// its empty-player-list result.
						want--
					}
					if got := serverQuitPendingAckCount(players, mode.dedicated); got != want {
						t.Fatalf("pending ACKs=%d, want %d for %d players, dedicated=%t",
							got, want, players, mode.dedicated)
					}
				})
			}
		})
	}
}

func TestServerShutdownUsesLivePlayerCountAndBuildRole(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "gui_quit.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var body *ast.BlockStmt
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "sub_446190" {
			body = fn.Body
		}
	}
	if body == nil {
		t.Fatal("normal shutdown entry missing")
	}
	found := 0
	ast.Inspect(body, func(node ast.Node) bool {
		assign, ok := node.(*ast.AssignStmt)
		if !ok || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
			return true
		}
		dst, ok := assign.Lhs[0].(*ast.Ident)
		if !ok || dst.Name != "dword_5d4594_825752" {
			return true
		}
		call, ok := assign.Rhs[0].(*ast.CallExpr)
		if !ok {
			return true
		}
		fn, ok := call.Fun.(*ast.Ident)
		if !ok || fn.Name != "serverQuitPendingAckCount" {
			t.Fatal("shutdown count bypasses the acknowledgement policy")
		}
		if len(call.Args) != 2 {
			t.Fatal("shutdown count must receive the live player count and build role")
		}
		count, ok := call.Args[0].(*ast.CallExpr)
		if !ok || len(count.Args) != 0 {
			t.Fatal("shutdown count does not sample Players.Count()")
		}
		method, ok := count.Fun.(*ast.SelectorExpr)
		if !ok || method.Sel.Name != "Count" {
			t.Fatal("shutdown count does not sample Players.Count()")
		}
		players, ok := method.X.(*ast.SelectorExpr)
		if !ok || players.Sel.Name != "Players" {
			t.Fatal("shutdown count does not use server Players")
		}
		server, ok := players.X.(*ast.Ident)
		if !ok || server.Name != "s" {
			t.Fatal("shutdown count does not use the live server")
		}
		dedicated, ok := call.Args[1].(*ast.Ident)
		if !ok || dedicated.Name != "isDedicatedServer" {
			t.Fatal("shutdown count does not use the actual build role")
		}
		found++
		return true
	})
	if found != 1 {
		t.Fatalf("shutdown acknowledgement policy calls=%d, want one", found)
	}
}
