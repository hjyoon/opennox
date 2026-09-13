package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestMonsterActionCast5413B0ObjectRecoilPreservesNativeTarget(t *testing.T) {
	update := &MonsterUpdateData{
		AIStackInd: 0,
		Field120_1: 3,
		Field330:   0.5,
		MonsterDef: &MonsterDef{MissileAttackFrame216: 3},
	}
	unit := &Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(update), PosVec: types.Ptf(4, 5)}
	target := &Object{PosVec: types.Ptf(50, 80), VelVec: types.Ptf(2, -3)}
	update.AIStack[0].Args[0] = 42
	update.AIStack[0].Args[2] = uintptr(unsafe.Pointer(target))
	var randomCalls [][2]float32
	var gotArg SpellAcceptArg
	var gotID int32
	(&Server{}).MonsterActionCast5413B0(unit, 0, MonsterActionCastRuntime5413B0{
		RandomFloat: func(min, max float32) float64 {
			randomCalls = append(randomCalls, [2]float32{min, max})
			return 0.25
		},
		CastSpell: func(id int32, caster *Object, arg *SpellAcceptArg) {
			if caster != unit {
				t.Fatalf("caster = %p, want %p", caster, unit)
			}
			gotID, gotArg = id, *arg
		},
	})
	if gotID != 42 || gotArg.Obj != target {
		t.Fatalf("cast id/target = %d/%p, want 42/%p", gotID, gotArg.Obj, target)
	}
	if len(randomCalls) != 3 || randomCalls[0] != [2]float32{0.5, 1.5} ||
		randomCalls[1] != [2]float32{-60, 60} || randomCalls[2] != [2]float32{-60, 60} {
		t.Fatalf("random calls = %v", randomCalls)
	}
	if gotArg.Pos != (types.Ptf(47.15, 84.65)) {
		t.Fatalf("recoil position = %v", gotArg.Pos)
	}
	if unit.Direction2 != DirFromVec(gotArg.Pos.Sub(unit.PosVec)) {
		t.Fatalf("direction = %d", unit.Direction2)
	}
}

func TestMonsterActionCast5413B0LocationAndGates(t *testing.T) {
	update := &MonsterUpdateData{
		AIStackInd: 0,
		Field120_1: 2,
		MonsterDef: &MonsterDef{MissileAttackFrame216: 2},
	}
	update.AIStack[0].SetArgs(uint32(18), uint32(0), types.Ptf(20, 30))
	unit := &Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(update)}
	var calls int
	runtime := MonsterActionCastRuntime5413B0{CastSpell: func(id int32, caster *Object, arg *SpellAcceptArg) {
		calls++
		if id != 18 || caster != unit || arg.Obj != nil || arg.Pos != (types.Ptf(20, 30)) {
			t.Fatalf("cast = %d/%p/%+v", id, caster, arg)
		}
	}}
	s := &Server{}
	s.MonsterActionCast5413B0(unit, 1, runtime)
	if calls != 1 {
		t.Fatalf("casts = %d, want 1", calls)
	}
	update.Field120_2 = 1
	s.MonsterActionCast5413B0(unit, 1, runtime)
	update.Field120_2 = 0
	unit.Buffs = 1 << ENCHANT_ANTI_MAGIC
	s.MonsterActionCast5413B0(unit, 1, runtime)
	if calls != 1 {
		t.Fatalf("gated casts = %d, want 1", calls)
	}
}

func TestMonsterActionCast5413B0EarlyFrameSound(t *testing.T) {
	update := &MonsterUpdateData{
		AIStackInd: 0,
		Field120_1: 1,
		MonsterDef: &MonsterDef{MissileAttackFrame216: 4},
	}
	var soundSet [16]uint32
	soundSet[14] = 913
	update.SoundSet122 = unsafe.Pointer(&soundSet)
	unit := &Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(update)}
	var got uint32
	(&Server{}).MonsterActionCast5413B0(unit, 0, MonsterActionCastRuntime5413B0{
		AudioEvent: func(id uint32, obj *Object) {
			if obj != unit {
				t.Fatalf("audio object = %p, want %p", obj, unit)
			}
			got = id
		},
	})
	if got != 913 {
		t.Fatalf("sound = %d, want 913", got)
	}
}
