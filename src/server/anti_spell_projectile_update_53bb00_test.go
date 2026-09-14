package server

import (
	"math"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestAntiSpellProjectileUpdate53BB00RetargetsNativePointers(t *testing.T) {
	owner := new(Object)
	update := &MissileUpdateData{}
	missile := &Object{PosVec: types.Ptf(10, 10), SpeedCur: 6, Field32: 1, UpdateData: unsafe.Pointer(update), ObjOwner: owner}
	tooFar := &Object{PosVec: types.Ptf(100, 10), ObjClass: object.ClassMissile, ObjSubClass: 2}
	near := &Object{PosVec: types.Ptf(30, 10), ObjClass: object.ClassMissile, ObjSubClass: 2}
	ownerChild := &Object{PosVec: types.Ptf(15, 10), ObjOwner: owner}
	destroyed := &Object{PosVec: types.Ptf(11, 10), ObjFlags: object.FlagDestroyed}
	wrongSubclass := &Object{PosVec: types.Ptf(12, 10), ObjClass: object.ClassMissile}
	blocked := &Object{PosVec: types.Ptf(13, 10)}
	var searches, sparks int
	runtime := AntiSpellProjectileRuntime53BB00{
		Frame: func() uint32 { return 10 }, FPS: func() uint32 { return 30 },
		EachMissile: func(pos types.Pointf, radius float32, visit func(*Object) bool) {
			if pos != missile.PosVec || radius != 600 {
				t.Fatalf("search area = %v/%g", pos, radius)
			}
			searches++
			for _, candidate := range []*Object{missile, destroyed, wrongSubclass, ownerChild, blocked, tooFar, near} {
				if !visit(candidate) {
					t.Fatal("search stopped early")
				}
			}
		},
		CanInteract:   func(_, target *Object) bool { return target != blocked },
		DelayedDelete: func(*Object) { t.Fatal("unexpected deletion") },
		Audio:         func(*Object) { t.Fatal("unexpected audio") },
		RandomFloat:   func(float32, float32) float32 { return 0 },
		RandomInt:     func(int, int) int { return 20 },
		CreateSpark: func(pos types.Pointf, kind, lifetime int, velocity types.Pointf, z float32, gotOwner *Object) {
			if pos != missile.PosVec || kind != 3 || lifetime != 20 || velocity != (types.Pointf{}) || z != 0 || gotOwner != missile {
				t.Fatalf("spark = %v/%d/%d/%v/%g/%p", pos, kind, lifetime, velocity, z, gotOwner)
			}
			sparks++
		},
	}
	AntiSpellProjectileUpdate53BB00(missile, runtime)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(near)) <= 0xffffffff {
		t.Fatal("target pointer unexpectedly below 4 GiB")
	}
	if update.Target != near || missile.Field34 != 10 || searches != 1 || sparks != 1 || missile.ForceVec.X < 5.9 || missile.ForceVec.Y != 0 {
		t.Fatalf("target/frame/searches/sparks/force = %p/%d/%d/%d/%v", update.Target, missile.Field34, searches, sparks, missile.ForceVec)
	}
	AntiSpellProjectileUpdate53BB00(missile, runtime)
	if searches != 1 || sparks != 2 {
		t.Fatalf("unnecessary retarget = searches %d, sparks %d", searches, sparks)
	}
}

func TestAntiSpellProjectileUpdate53BB00HitAndSparkRandomOrder(t *testing.T) {
	target := &Object{PosVec: types.Ptf(3, 4)}
	update := &MissileUpdateData{Target: target}
	missile := &Object{SpeedCur: 10, UpdateData: unsafe.Pointer(update)}
	var calls []string
	randomFloats := []float32{-1.5, 0.5, 2, -3}
	runtime := AntiSpellProjectileRuntime53BB00{
		Frame: func() uint32 { return 1 }, FPS: func() uint32 { return 30 },
		EachMissile: func(types.Pointf, float32, func(*Object) bool) { t.Fatal("unexpected search") },
		CanInteract: func(*Object, *Object) bool { t.Fatal("unexpected interaction"); return false },
		DelayedDelete: func(obj *Object) {
			if obj == missile {
				calls = append(calls, "delete missile")
			} else if obj == target {
				calls = append(calls, "delete target")
			} else {
				t.Fatalf("deleted %p", obj)
			}
		},
		Audio: func(obj *Object) {
			if obj != missile {
				t.Fatalf("audio object %p", obj)
			}
			calls = append(calls, "audio")
		},
		RandomFloat: func(min, max float32) float32 {
			if len(randomFloats) == 0 {
				t.Fatal("extra random float")
			}
			value := randomFloats[0]
			randomFloats = randomFloats[1:]
			if len(randomFloats) >= 2 && (min != -2 || max != 2) || len(randomFloats) < 2 && (min != -4 || max != 4) {
				t.Fatalf("random float range %g..%g", min, max)
			}
			calls = append(calls, "float")
			return value
		},
		RandomInt: func(min, max int) int {
			if min != 15 || max != 30 {
				t.Fatalf("random int range %d..%d", min, max)
			}
			calls = append(calls, "int")
			return 22
		},
		CreateSpark: func(pos types.Pointf, kind, lifetime int, velocity types.Pointf, z float32, owner *Object) {
			calls = append(calls, "spark")
			if pos != (types.Ptf(-3, 2)) || kind != 3 || lifetime != 22 || velocity != (types.Ptf(0.5, -1.5)) || z != 0 || owner != missile {
				t.Fatalf("spark = %v/%d/%d/%v/%g/%p", pos, kind, lifetime, velocity, z, owner)
			}
		},
	}
	AntiSpellProjectileUpdate53BB00(missile, runtime)
	if got, want := missile.ForceVec, (types.Ptf(3*10/5.1, 4*10/5.1)); math.Abs(float64(got.X-want.X)) > 1e-5 || math.Abs(float64(got.Y-want.Y)) > 1e-5 {
		t.Fatalf("homing force = %v, want %v", got, want)
	}
	if got := strings.Join(calls, ","); got != "audio,delete missile,delete target,float,float,int,float,float,spark" {
		t.Fatalf("call order = %s", got)
	}
}

func TestAntiSpellProjectileUpdate53BB00NoTargetAndExpiry(t *testing.T) {
	update := &MissileUpdateData{Target: &Object{ObjFlags: object.FlagDestroyed}}
	missile := &Object{VelVec: types.Ptf(3, 4), SpeedCur: 10, Field32: 1, Field34: 5, UpdateData: unsafe.Pointer(update)}
	frame, searches, deletions, sparks := uint32(10), 0, 0, 0
	runtime := AntiSpellProjectileRuntime53BB00{
		Frame: func() uint32 { return frame }, FPS: func() uint32 { return 30 },
		EachMissile:   func(types.Pointf, float32, func(*Object) bool) { searches++ },
		CanInteract:   func(*Object, *Object) bool { return true },
		DelayedDelete: func(*Object) { deletions++ },
		Audio:         func(*Object) { t.Fatal("unexpected audio") },
		RandomFloat:   func(float32, float32) float32 { return 0 },
		RandomInt:     func(int, int) int { return 15 },
		CreateSpark:   func(types.Pointf, int, int, types.Pointf, float32, *Object) { sparks++ },
	}
	AntiSpellProjectileUpdate53BB00(missile, runtime)
	if update.Target != nil || searches != 0 || sparks != 1 || math.Abs(float64(missile.ForceVec.X-3*10/5.1)) > 1e-5 {
		t.Fatalf("idle state target/searches/sparks/force = %p/%d/%d/%v", update.Target, searches, sparks, missile.ForceVec)
	}
	frame = 13
	AntiSpellProjectileUpdate53BB00(missile, runtime)
	if searches != 1 || missile.Field34 != 13 {
		t.Fatalf("retarget timing = %d/%d", searches, missile.Field34)
	}
	frame = 152
	AntiSpellProjectileUpdate53BB00(missile, runtime)
	if deletions != 1 || sparks != 2 {
		t.Fatalf("expiry = deletions %d, sparks %d", deletions, sparks)
	}
}
