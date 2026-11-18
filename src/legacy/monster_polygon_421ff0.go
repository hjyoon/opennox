package legacy

/*
#include "GAME1_1.h"
*/
import "C"

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

const monsterPolygonUninitialized421FF0 = uint32(0xdeadface)

type MonsterPolygonRuntime421FF0 struct {
	Find     func([2]int32, uint32) *Nox_player_polygon_check_data
	ByID     func(uint32) *Nox_player_polygon_check_data
	Callback func(*Nox_player_polygon_check_data, int, *server.Object, server.ScriptEventType)
}

func polygonCoordinateEqual421FF0(a, b float32) bool {
	// The x87 C3 flag is set for equal and unordered comparisons.
	return !(a < b || a > b)
}

func MonsterPolygonTransition421FF0(unit *server.Object, r MonsterPolygonRuntime421FF0) {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) {
		return
	}
	update := unit.UpdateDataMonster()
	current := update.Field0
	pos, previous := unit.PosVec, unit.PrevPos
	if current != monsterPolygonUninitialized421FF0 &&
		polygonCoordinateEqual421FF0(pos.X, previous.X) && polygonCoordinateEqual421FF0(pos.Y, previous.Y) {
		return
	}
	point := [2]int32{server.X87TruncateSignedDWORD419A70(pos.X), server.X87TruncateSignedDWORD419A70(pos.Y)}
	next := r.Find(point, current)
	current = update.Field0
	if next != nil {
		if current == next.Field_0[20] {
			return
		}
		if current != monsterPolygonUninitialized421FF0 {
			if current != 0 {
				if old := r.ByID(current); old != nil && old.Field_0[31] != ^uint32(0) {
					r.Callback(old, 30, unit, server.NoxEventPolygonEnterZZZ)
				}
			}
			if next.Field_0[29] != ^uint32(0) {
				r.Callback(next, 28, unit, server.NoxEventPolygonEnterYYY)
			}
		}
		update.Field0 = next.Field_0[20]
		return
	}
	if current == 0 || current == monsterPolygonUninitialized421FF0 {
		return
	}
	if old := r.ByID(current); old != nil && old.Field_0[21] != 0 && old.Field_0[31] != ^uint32(0) {
		r.Callback(old, 30, unit, server.NoxEventPolygonEnterXXX)
	}
	update.Field0 = 0
}

//export nox_xxx_monsterPolygonEnter_421FF0
func nox_xxx_monsterPolygonEnter_421FF0(unit *nox_object_t) {
	Nox_xxx_monsterPolygonEnter_421FF0(asObjectS(unit))
}
