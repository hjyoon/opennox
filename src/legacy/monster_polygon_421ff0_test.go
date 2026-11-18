package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

type monsterPolygonCallback421FF0 struct {
	polygon *Nox_player_polygon_check_data
	word    int
	event   server.ScriptEventType
}

func monsterPolygonTestUnit421FF0(current uint32) *server.Object {
	update := &server.MonsterUpdateData{Field0: current}
	return &server.Object{
		ObjClass:   object.ClassMonster,
		UpdateData: unsafe.Pointer(update),
	}
}

func TestMonsterPolygonEnterNative421FF0Initial(t *testing.T) {
	unit := monsterPolygonTestUnit421FF0(monsterPolygonUninitialized421FF0)
	polygon := &Nox_player_polygon_check_data{}
	polygon.Field_0[20] = 7
	var callbacks []monsterPolygonCallback421FF0
	MonsterPolygonTransition421FF0(unit, MonsterPolygonRuntime421FF0{
		Find: func(_ [2]int32, previous uint32) *Nox_player_polygon_check_data {
			if previous != monsterPolygonUninitialized421FF0 {
				t.Fatalf("previous = %#x", previous)
			}
			return polygon
		},
		ByID: func(uint32) *Nox_player_polygon_check_data { return nil },
		Callback: func(p *Nox_player_polygon_check_data, word int, _ *server.Object, event server.ScriptEventType) {
			callbacks = append(callbacks, monsterPolygonCallback421FF0{polygon: p, word: word, event: event})
		},
	})
	if got := unit.UpdateDataMonster().Field0; got != 7 {
		t.Fatalf("polygon = %d, want 7", got)
	}
	if len(callbacks) != 0 {
		t.Fatalf("initial callbacks = %d, want 0", len(callbacks))
	}
}

func TestMonsterPolygonEnterNative421FF0Transition(t *testing.T) {
	unit := monsterPolygonTestUnit421FF0(2)
	unit.PosVec.X = 2
	old := &Nox_player_polygon_check_data{}
	old.Field_0[31] = 1
	next := &Nox_player_polygon_check_data{}
	next.Field_0[20] = 3
	next.Field_0[29] = 1
	var callbacks []monsterPolygonCallback421FF0
	MonsterPolygonTransition421FF0(unit, MonsterPolygonRuntime421FF0{
		Find: func([2]int32, uint32) *Nox_player_polygon_check_data { return next },
		ByID: func(id uint32) *Nox_player_polygon_check_data {
			if id == 2 {
				return old
			}
			return nil
		},
		Callback: func(p *Nox_player_polygon_check_data, word int, _ *server.Object, event server.ScriptEventType) {
			callbacks = append(callbacks, monsterPolygonCallback421FF0{polygon: p, word: word, event: event})
		},
	})
	if got := unit.UpdateDataMonster().Field0; got != 3 {
		t.Fatalf("polygon = %d, want 3", got)
	}
	want := []monsterPolygonCallback421FF0{
		{polygon: old, word: 30, event: server.NoxEventPolygonEnterZZZ},
		{polygon: next, word: 28, event: server.NoxEventPolygonEnterYYY},
	}
	if len(callbacks) != len(want) {
		t.Fatalf("callbacks = %#v, want %#v", callbacks, want)
	}
	for i := range want {
		if callbacks[i] != want[i] {
			t.Errorf("callback %d = %#v, want %#v", i, callbacks[i], want[i])
		}
	}
}

func TestMonsterPolygonEnterNative421FF0Exit(t *testing.T) {
	unit := monsterPolygonTestUnit421FF0(4)
	unit.PosVec.Y = 1
	old := &Nox_player_polygon_check_data{}
	old.Field_0[21] = 1
	old.Field_0[31] = 1
	var callbacks []monsterPolygonCallback421FF0
	MonsterPolygonTransition421FF0(unit, MonsterPolygonRuntime421FF0{
		Find: func([2]int32, uint32) *Nox_player_polygon_check_data { return nil },
		ByID: func(uint32) *Nox_player_polygon_check_data { return old },
		Callback: func(p *Nox_player_polygon_check_data, word int, _ *server.Object, event server.ScriptEventType) {
			callbacks = append(callbacks, monsterPolygonCallback421FF0{polygon: p, word: word, event: event})
		},
	})
	if got := unit.UpdateDataMonster().Field0; got != 0 {
		t.Fatalf("polygon = %d, want 0", got)
	}
	want := monsterPolygonCallback421FF0{polygon: old, word: 30, event: server.NoxEventPolygonEnterXXX}
	if len(callbacks) != 1 || callbacks[0] != want {
		t.Fatalf("callbacks = %#v, want %#v", callbacks, []monsterPolygonCallback421FF0{want})
	}
}
