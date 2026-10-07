package opennox

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/opennox/libs/types"
)

func TestE2EMarkNPCLayoutClearance(t *testing.T) {
	for index, direction := range []types.Pointf{
		types.Ptf(1, 0), types.Ptf(-1, 0), types.Ptf(0, 1), types.Ptf(0, -1),
		types.Ptf(1, 1).Normalize(), types.Ptf(-1, 1).Normalize(), types.Ptf(1, -1).Normalize(), types.Ptf(-1, -1).Normalize(),
	} {
		t.Run(fmt.Sprint(index), func(t *testing.T) {
			origin := types.Ptf(1000, 2000)
			for _, radii := range [][2]float32{{11, 11}, {32, 64}} {
				bound := radii[0] + radii[1] + 4
				position := e2eMarkNPCPosition(origin, direction, radii[0], radii[1])
				for ordinal := 0; ordinal < 5; ordinal++ {
					end := origin.Add(direction.Mul(float32(12 * ordinal)))
					if !e2eWarriorLaneMissesCircle(origin, end, position, bound) {
						t.Fatalf("idle NPC overlaps player input lane: radii=%v ordinal=%d point=%v", radii, ordinal, position)
					}
				}
			}
			if e2eWarriorLaneMissesCircle(origin, origin.Add(direction.Mul(48)), origin.Sub(direction.Mul(24)), 26) {
				t.Fatal("original 24-unit NPC placement did not reproduce the conservative overlap")
			}
		})
	}
}

func TestE2EMarkStockNPCLayoutBinding(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_mark_spell.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	bound := false
	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "prepare" {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) != 3 {
				return true
			}
			method, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || method.Sel.Name != "CreateObjectAt" {
				return true
			}
			position, ok := call.Args[2].(*ast.CallExpr)
			if !ok || len(position.Args) != 4 {
				return true
			}
			name, ok := position.Fun.(*ast.Ident)
			bound = ok && name.Name == "e2eMarkNPCPosition"
			return true
		})
	}
	if !bound {
		t.Fatal("Mark stock NPC still uses the overlapping fixed 24-unit input instead of the radius-aware verified forward lane")
	}
}
