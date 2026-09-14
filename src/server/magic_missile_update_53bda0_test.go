package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestMagicMissileUpdate53BDA0NativeLayoutAndHoming(t *testing.T) {
	ptrSize := unsafe.Sizeof(uintptr(0))
	if got, want := unsafe.Sizeof(MissileUpdateData{}), uintptr(28)+3*(ptrSize-4); got != want {
		t.Fatalf("MissileUpdateData size = %d, want %d", got, want)
	}
	if got, want := unsafe.Offsetof(MissileUpdateData{}.Target), ptrSize; got != want {
		t.Fatalf("Target offset = %d, want %d", got, want)
	}
	if got, want := unsafe.Offsetof(MissileUpdateData{}.SpellID), 2*ptrSize+4; got != want {
		t.Fatalf("SpellID offset = %d, want %d", got, want)
	}

	owner := new(Object)
	target := &Object{PosVec: types.Ptf(10, 20)}
	update := &MissileUpdateData{Owner: owner, Target: target, SpellID: 50}
	missile := &Object{
		PosVec:     types.Ptf(10, 10),
		Direction1: 77,
		Direction2: 250,
		Field32:    100,
		SpeedCur:   8,
		UpdateData: unsafe.Pointer(update),
	}
	if ptrSize == 8 && (uintptr(unsafe.Pointer(update)) <= 0xffffffff || uintptr(unsafe.Pointer(target)) <= 0xffffffff) {
		t.Fatal("expected missile and target pointers above 4 GiB")
	}
	runtime := MagicMissileUpdateRuntime53BDA0{
		Frame: func() uint32 { return 101 },
		FPS:   func() uint32 { return 30 },
		SearchTarget: func(*Object, *Object) *Object {
			t.Fatal("existing target should not be searched")
			return nil
		},
		Collide: func(*Object) { t.Fatal("live missile should not collide") },
	}
	MagicMissileUpdate53BDA0(missile, runtime)
	if update.Owner != owner || update.Target != target || update.SpellID != 50 {
		t.Fatalf("update pointers/spell = %p/%p/%d", update.Owner, update.Target, update.SpellID)
	}
	if missile.Direction1 != 77 || missile.Direction2 != 36 {
		t.Fatalf("directions = %d/%d, want 77/36", missile.Direction1, missile.Direction2)
	}
	force := missile.Direction2.Vec().Mul(missile.SpeedCur)
	if missile.ForceVec != force || math.Float32bits(missile.Float28) != 1061997773 {
		t.Fatalf("force/drag = %v/%#x, want %v/%#x", missile.ForceVec, math.Float32bits(missile.Float28), force, uint32(1061997773))
	}

	missile.Direction2 = 10
	target.PosVec = types.Ptf(10, 0)
	MagicMissileUpdate53BDA0(missile, runtime)
	if missile.Direction2 != 224 {
		t.Fatalf("clockwise wrap = %d, want 224", missile.Direction2)
	}
}

func TestMagicMissileUpdate53BDA0RetargetAndExpiry(t *testing.T) {
	owner := new(Object)
	oldTarget := &Object{ObjFlags: object.FlagDead}
	newTarget := &Object{PosVec: types.Ptf(20, 0)}
	update := &MissileUpdateData{Owner: owner, Target: oldTarget}
	missile := &Object{UpdateData: unsafe.Pointer(update), Field32: 100, SpeedCur: 5}
	frame, searches, collisions := uint32(103), 0, 0
	runtime := MagicMissileUpdateRuntime53BDA0{
		Frame: func() uint32 { return frame },
		FPS:   func() uint32 { return 30 },
		SearchTarget: func(gotMissile, gotOwner *Object) *Object {
			if gotMissile != missile || gotOwner != owner {
				t.Fatalf("search arguments = %p/%p", gotMissile, gotOwner)
			}
			searches++
			return newTarget
		},
		Collide: func(got *Object) {
			if got != missile {
				t.Fatalf("collision object = %p, want %p", got, missile)
			}
			collisions++
		},
	}
	MagicMissileUpdate53BDA0(missile, runtime)
	if update.Target != nil || searches != 0 {
		t.Fatalf("dead target before search tick = %p, searches %d", update.Target, searches)
	}
	frame = 104
	MagicMissileUpdate53BDA0(missile, runtime)
	if update.Target != newTarget || searches != 1 {
		t.Fatalf("retarget = %p, searches %d; want %p/1", update.Target, searches, newTarget)
	}

	update.Target = nil
	missile.ObjOwner = newTarget
	frame = 112
	MagicMissileUpdate53BDA0(missile, runtime)
	if update.Target != nil || searches != 2 {
		t.Fatalf("owner-chain target = %p, searches %d; want nil/2", update.Target, searches)
	}

	frame = 190
	MagicMissileUpdate53BDA0(missile, runtime)
	if collisions != 0 {
		t.Fatalf("missile collided on its last live frame: %d", collisions)
	}
	frame = 191
	force := missile.ForceVec
	MagicMissileUpdate53BDA0(missile, runtime)
	if collisions != 1 || missile.ForceVec != force {
		t.Fatalf("expired missile collisions/force = %d/%v, want 1/%v", collisions, missile.ForceVec, force)
	}
	frame = 100
	owner.ObjFlags = object.FlagDestroyed
	MagicMissileUpdate53BDA0(missile, runtime)
	if collisions != 2 {
		t.Fatalf("destroyed owner collisions = %d, want 2", collisions)
	}
	update.Owner = nil
	MagicMissileUpdate53BDA0(missile, runtime)
	if collisions != 3 {
		t.Fatalf("missing owner collisions = %d, want 3", collisions)
	}
}
