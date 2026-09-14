package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/types"
)

func TestSpellTurnUndeadCreate531310NativePointersAndProjectiles(t *testing.T) {
	caster := &Object{PosVec: types.Pointf{X: 100, Y: 200}}
	anchor := &Object{PosVec: types.Pointf{X: 300, Y: 400}}
	for _, tc := range []struct {
		name     string
		flag     uint32
		caster   *Object
		position types.Pointf
	}{
		{"caster", 0, caster, caster.PosVec},
		{"anchor", 1, caster, anchor.PosVec},
		{"glyph anchor without caster", 1, nil, anchor.PosVec},
	} {
		t.Run(tc.name, func(t *testing.T) {
			record := &DurSpell{Level: 2, Caster16: tc.caster, Flag20: tc.flag, Obj24: anchor,
				Pos: types.Pointf{X: math.Float32frombits(0x3fdccccc)}}
			var objects []*Object
			var events []string
			runtime := SpellTurnUndeadRuntime531310{
				KillPoints: func(index uint32) float32 {
					if index != 1 {
						t.Errorf("kill-point index = %d, want 1", index)
					}
					return 2.5 // FISTP rounds ties to even.
				},
				NewObject: func(id string) *Object {
					if id != "UndeadKiller" {
						t.Errorf("object type = %q", id)
					}
					data := &UndeadKillerCollideData{}
					obj := &Object{CollideData: unsafe.Pointer(data), Float28: 7}
					objects = append(objects, obj)
					return obj
				},
				CreateAt: func(obj, owner *Object, position types.Pointf) {
					events = append(events, "create")
					if owner != tc.caster || position != tc.position ||
						(*UndeadKillerCollideData)(obj.CollideData).Spell != record {
						t.Errorf("create = %p/%p/%v", obj, owner, position)
					}
				},
				SendPointFX: func(effect netmsg.Op, position types.Pointf) {
					events = append(events, "fx")
					if effect != netmsg.MSG_FX_TURN_UNDEAD || position != tc.position {
						t.Errorf("FX = %d/%v", effect, position)
					}
				},
			}
			if got := SpellTurnUndeadCreate531310(record, runtime); got != 0 {
				t.Fatalf("create = %d, want 0", got)
			}
			if record.Field72 != 2 || len(objects) != 43 || len(events) != 44 || events[43] != "fx" {
				t.Fatalf("budget/objects/events = %d/%d/%v", record.Field72, len(objects), events)
			}
			for i, obj := range objects {
				dir := Dir16(i * 6)
				cos, sin := SinCosDir(byte(dir))
				if obj.Direction1 != dir || obj.Direction2 != dir ||
					obj.VelVec != (types.Pointf{X: cos * 4, Y: sin * 4}) || obj.Float28 != 0 {
					t.Fatalf("projectile %d = dir %d/%d, velocity %v, damping %g",
						i, obj.Direction1, obj.Direction2, obj.VelVec, obj.Float28)
				}
			}
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(record)) <= uintptr(^uint32(0)) {
				t.Fatalf("record pointer %p did not exercise native width", record)
			}
		})
	}
}

func TestSpellTurnUndeadCreate531310MissingObjectsAndSource(t *testing.T) {
	for _, tc := range []struct {
		name   string
		level  uint32
		caster *Object
		want   int32
	}{
		{"missing source", 1, nil, 1},
		{"missing projectile type", 0, &Object{PosVec: types.Pointf{X: 5}}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			record := &DurSpell{Level: tc.level, Caster16: tc.caster}
			var indices []uint32
			var newCount, fxCount int
			runtime := SpellTurnUndeadRuntime531310{
				KillPoints: func(index uint32) float32 {
					indices = append(indices, index)
					return 9
				},
				NewObject:   func(string) *Object { newCount++; return nil },
				SendPointFX: func(netmsg.Op, types.Pointf) { fxCount++ },
			}
			if got := SpellTurnUndeadCreate531310(record, runtime); got != tc.want {
				t.Fatalf("create = %d, want %d", got, tc.want)
			}
			if len(indices) != 1 || indices[0] != tc.level-1 || record.Field72 != 9 {
				t.Fatalf("budget/index = %d/%v", record.Field72, indices)
			}
			if tc.caster == nil && (newCount != 0 || fxCount != 0) {
				t.Fatalf("missing source created %d objects and %d FX", newCount, fxCount)
			}
			if tc.caster != nil && (newCount != 43 || fxCount != 1) {
				t.Fatalf("missing type attempts/FX = %d/%d", newCount, fxCount)
			}
		})
	}
}

func TestSpellTurnUndeadUpdate531410NoOp(t *testing.T) {
	record := &DurSpell{Field72: 15, Flags88: 1}
	if got := SpellTurnUndeadUpdate531410(record); got != 0 || record.Field72 != 15 || record.Flags88 != 1 {
		t.Fatalf("update = %d, record = %+v", got, record)
	}
}

func TestSpellTurnUndeadDestroy531420MatchesSpellAndType(t *testing.T) {
	record := &DurSpell{}
	other := &DurSpell{}
	matchingA := &Object{TypeInd: 19, CollideData: unsafe.Pointer(&UndeadKillerCollideData{Spell: record})}
	wrongType := &Object{TypeInd: 20, CollideData: unsafe.Pointer(&UndeadKillerCollideData{Spell: record})}
	wrongSpell := &Object{TypeInd: 19, CollideData: unsafe.Pointer(&UndeadKillerCollideData{Spell: other})}
	missingData := &Object{TypeInd: 19}
	matchingB := &Object{TypeInd: 19, CollideData: unsafe.Pointer(&UndeadKillerCollideData{Spell: record})}
	matchingA.ObjNext = wrongType
	wrongType.ObjNext = wrongSpell
	wrongSpell.ObjNext = missingData
	missingData.ObjNext = matchingB
	var deleted []*Object
	SpellTurnUndeadDestroy531420(record, SpellTurnUndeadRuntime531310{
		TypeInd: func(id string) int {
			if id != "UndeadKiller" {
				t.Errorf("type ID = %q", id)
			}
			return 19
		},
		FirstObject: func() *Object { return matchingA },
		DelayedDelete: func(obj *Object) {
			deleted = append(deleted, obj)
			obj.ObjNext = nil // The traversal must cache Next before deleting.
		},
	})
	if len(deleted) != 2 || deleted[0] != matchingA || deleted[1] != matchingB {
		t.Fatalf("deleted = %v, want both matching killers", deleted)
	}
}
