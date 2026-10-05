package opennox

import (
	"math"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

// aiDependencyObjectAtVisibleLocation546E0A preserves condition 48 of
// GAME.EXE 00546A70. The cached slot, rather than the entry target or a new
// update record, supplies the position after CanInteract. The ray always
// runs, and only a clear ray without interaction rejects the dependency.
func aiDependencyObjectAtVisibleLocation546E0A(
	unit *server.Object,
	slot *server.AIStackItem,
	canInteract func(*server.Object, *server.Object, int) bool,
	traceRay func(types.Pointf, types.Pointf, server.MapTraceFlags) bool,
) bool {
	target := slot.ArgObj(2)
	visible := false
	if target != nil && canInteract(unit, target, 0) {
		visible = true
		pos := slot.ArgObj(2).PosVec
		slot.Args[0] = uintptr(math.Float32bits(pos.X))
		slot.Args[1] = uintptr(math.Float32bits(pos.Y))
	}
	return !traceRay(unit.PosVec, slot.ArgPos(0), server.MapTraceFlag1) || visible
}
