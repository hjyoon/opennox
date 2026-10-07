package opennox

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"testing"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func e2eLockLayoutOverlaps(door *server.Object, rect types.Rectf) bool {
	return door.CollideP1.X < rect.Max.X && door.CollideP2.X > rect.Min.X &&
		door.CollideP1.Y < rect.Max.Y && door.CollideP2.Y > rect.Min.Y
}

func TestE2ELockLayoutSeparatesNativeDoorFootprint(t *testing.T) {
	for _, offset := range []types.Pointf{
		types.Ptf(160, 0), types.Ptf(-160, 0), types.Ptf(0, 160), types.Ptf(0, -160),
		types.Ptf(114, 114), types.Ptf(-114, 114), types.Ptf(114, -114), types.Ptf(-114, -114),
	} {
		t.Run(fmt.Sprint(offset), func(t *testing.T) {
			origin, direction := types.Ptf(1011, 1065), offset.Normalize()
			positions := e2eLockLayout(origin, direction)
			update := &server.DoorUpdateData{
				TileX: e2eDoorTileCoordinate4F4CB0(server.DoorDirectionX(16), positions[0].X),
				TileY: e2eDoorTileCoordinate4F4CB0(server.DoorDirectionY(16), positions[0].Y),
			}
			bounds := e2eLockGroupBounds(update)
			for index, position := range positions {
				// This conservative radius includes the stock door footprint;
				// use the actual production collider, not center-only geometry.
				door := &server.Object{PosVec: position, Shape: server.Shape{Kind: server.ShapeKindCircle}}
				door.Shape.Circle.R = 40
				door.Nox_xxx_objectUnkUpdateCoords_4E7290()
				if got := e2eLockLayoutOverlaps(door, bounds); got != (index < 2) {
					t.Fatalf("Door %d footprint %v..%v overlaps=%t bounds=%v", index, door.CollideP1, door.CollideP2, got, bounds)
				}
			}
			for _, caster := range []types.Pointf{origin, origin.Sub(direction.Mul(24))} {
				selected := positions[0].Sub(caster)
				for _, position := range positions[1:] {
					delta := position.Sub(caster)
					if math.Hypot(float64(delta.X), float64(delta.Y)) <= math.Hypot(float64(selected.X), float64(selected.Y)) {
						t.Fatal("fixture changed nearest-door identity")
					}
				}
			}
		})
	}
}

func TestE2ELockFormerDiagonalOutsideDoorActuallyOverlaps(t *testing.T) {
	direction := types.Ptf(114, 114).Normalize()
	selected := types.Ptf(1012, 1058)
	update := &server.DoorUpdateData{
		TileX: e2eDoorTileCoordinate4F4CB0(server.DoorDirectionX(16), selected.X),
		TileY: e2eDoorTileCoordinate4F4CB0(server.DoorDirectionY(16), selected.Y),
	}
	door := &server.Object{PosVec: selected.Add(direction.Mul(70)), Shape: server.Shape{Kind: server.ShapeKindCircle}}
	door.Shape.Circle.R = 32
	door.Nox_xxx_objectUnkUpdateCoords_4E7290()
	if !e2eLockLayoutOverlaps(door, e2eLockGroupBounds(update)) {
		t.Fatal("former 70-unit diagonal fixture no longer demonstrates footprint overlap")
	}
}

func TestE2ELockPreparationUsesAndChecksStockLayout(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_lock_spell.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var prepare *ast.FuncDecl
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "prepare" {
			prepare = fn
		}
	}
	if prepare == nil {
		t.Fatal("actual Lock preparation missing")
	}
	calls := make(map[string]int)
	ast.Inspect(prepare.Body, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			switch fn := call.Fun.(type) {
			case *ast.Ident:
				calls[fn.Name]++
			case *ast.SelectorExpr:
				calls[fn.Sel.Name]++
			}
		}
		return true
	})
	for _, name := range []string{"e2eLockLayout", "e2eLockGroupBounds", "EachObjInRect"} {
		if calls[name] != 1 {
			t.Errorf("actual prepare must call %s exactly once; got=%d", name, calls[name])
		}
	}
}
