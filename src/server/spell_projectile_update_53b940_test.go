package server

import (
	"math"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestSpellProjectileUpdate53B940TracksNativeTarget(t *testing.T) {
	owner := new(Object)
	source := new(Object)
	target := &Object{PosVec: types.Ptf(13, 14)}
	update := &SpellProjectileUpdateData{Field0: owner, Target: target, Field8: source, Spell12: 42}
	missile := &Object{PosVec: types.Ptf(10, 10), VelVec: types.Ptf(-1, 0), SpeedCur: 10, Field32: 10, UpdateData: unsafe.Pointer(update)}
	lifetimeCalls := 0
	SpellProjectileUpdate53B940(missile, SpellProjectileUpdateRuntime53B940{
		Frame:    func() uint32 { return 12 },
		TickRate: func() uint32 { t.Fatal("unexpected tick-rate read"); return 0 },
		Lifetime: func(key string) float64 {
			if key != "TargetedSpellLifetime" {
				t.Fatalf("lifetime key = %q", key)
			}
			lifetimeCalls++
			return 2.5
		},
		SearchTarget: func(*Object, *Object, uint32) *Object { t.Fatal("unexpected search"); return nil },
		SendPointFX:  func(types.Pointf) { t.Fatal("unexpected point FX") },
		Expire:       func(*Object) { t.Fatal("unexpected expiry") },
	})
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(target)) <= 0xffffffff {
		t.Fatal("target pointer unexpectedly below 4 GiB")
	}
	if lifetimeCalls != 1 || update.Field0 != owner || update.Field8 != source || update.Target != target {
		t.Fatalf("native pointers or lifetime changed: calls=%d owner=%p source=%p target=%p", lifetimeCalls, update.Field0, update.Field8, update.Target)
	}
	want := types.Ptf(3*10/5.1, 4*10/5.1)
	if math.Abs(float64(missile.ForceVec.X-want.X)) > 1e-5 || math.Abs(float64(missile.ForceVec.Y-want.Y)) > 1e-5 {
		t.Fatalf("force = %v, want %v", missile.ForceVec, want)
	}
	if got := math.Float32bits(missile.Float28); got != 1063675494 {
		t.Fatalf("drag bits = %#x", got)
	}
}

func TestSpellProjectileUpdate53B940ClearsDestroyedPointersAndRetargets(t *testing.T) {
	owner := &Object{ObjFlags: object.FlagDestroyed}
	source := &Object{ObjFlags: object.FlagDestroyed}
	deadTarget := &Object{ObjFlags: object.FlagDead}
	newTarget := &Object{PosVec: types.Ptf(0, 10)}
	update := &SpellProjectileUpdateData{Field0: owner, Target: deadTarget, Field8: source, Spell12: 31}
	missile := &Object{VelVec: types.Ptf(3, 4), SpeedCur: 10, Field32: 10, Field34: 15, UpdateData: unsafe.Pointer(update)}
	frame := uint32(20)
	searches := 0
	lifetimeKeys := []string{}
	runtime := SpellProjectileUpdateRuntime53B940{
		Frame:    func() uint32 { return frame },
		TickRate: func() uint32 { return 30 },
		Lifetime: func(key string) float64 {
			lifetimeKeys = append(lifetimeKeys, key)
			return 100
		},
		SearchTarget: func(gotMissile, gotOwner *Object, spellID uint32) *Object {
			if gotMissile != missile || gotOwner != nil || spellID != 31 {
				t.Fatalf("search args = %p/%p/%d", gotMissile, gotOwner, spellID)
			}
			searches++
			return newTarget
		},
		SendPointFX: func(types.Pointf) { t.Fatal("unexpected point FX") },
		Expire:      func(*Object) { t.Fatal("unexpected expiry") },
	}
	SpellProjectileUpdate53B940(missile, runtime)
	if update.Field0 != nil || update.Field8 != nil || update.Target != nil || searches != 0 {
		t.Fatalf("destroyed pointers/searches = %p/%p/%p/%d", update.Field0, update.Field8, update.Target, searches)
	}
	if len(lifetimeKeys) != 1 || lifetimeKeys[0] != "TargetedSpellLifetime" {
		t.Fatalf("first lifetime key = %v", lifetimeKeys)
	}
	if want := types.Ptf(3*10/5.1, 4*10/5.1); math.Abs(float64(missile.ForceVec.X-want.X)) > 1e-5 || math.Abs(float64(missile.ForceVec.Y-want.Y)) > 1e-5 {
		t.Fatalf("untargeted force = %v, want %v", missile.ForceVec, want)
	}
	frame = 23
	SpellProjectileUpdate53B940(missile, runtime)
	if len(lifetimeKeys) != 2 || lifetimeKeys[1] != "UnTargetedSpellLifetime" {
		t.Fatalf("second lifetime key = %v", lifetimeKeys)
	}
	if searches != 1 || update.Target != newTarget || missile.Field34 != 23 || missile.ForceVec.Y < 9.8 {
		t.Fatalf("retarget state = searches=%d target=%p frame=%d force=%v", searches, update.Target, missile.Field34, missile.ForceVec)
	}
}

func TestSpellProjectileUpdate53B940ExpiresAfterRoundedUntargetedLifetime(t *testing.T) {
	update := &SpellProjectileUpdateData{}
	missile := &Object{PosVec: types.Ptf(8, 9), Field32: 10, Field34: 12, UpdateData: unsafe.Pointer(update)}
	frame := uint32(12)
	var calls []string
	runtime := SpellProjectileUpdateRuntime53B940{
		Frame:    func() uint32 { return frame },
		TickRate: func() uint32 { return 30 },
		Lifetime: func(key string) float64 {
			if key != "UnTargetedSpellLifetime" {
				t.Fatalf("lifetime key = %q", key)
			}
			return 2.5 // x87 nearest-even rounds this to two frames.
		},
		SearchTarget: func(*Object, *Object, uint32) *Object { t.Fatal("unexpected search"); return nil },
		SendPointFX: func(pos types.Pointf) {
			if pos != missile.PosVec {
				t.Fatalf("point FX position = %v", pos)
			}
			calls = append(calls, "point FX")
		},
		Expire: func(got *Object) {
			if got != missile {
				t.Fatalf("expired object = %p", got)
			}
			calls = append(calls, "expire")
		},
	}
	SpellProjectileUpdate53B940(missile, runtime)
	if len(calls) != 0 {
		t.Fatalf("expired on lifetime boundary: %v", calls)
	}
	frame = 13
	SpellProjectileUpdate53B940(missile, runtime)
	if got := strings.Join(calls, ","); got != "point FX,expire" {
		t.Fatalf("expiry call order = %q", got)
	}
}
