package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestBlackPowderBurnUpdateNative53CCB0InitializesFuseOnNativeObject(t *testing.T) {
	source := &Object{Field32: 100, Field34: 100}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(source)) <= math.MaxUint32 {
		t.Fatalf("source pointer = %p, want address above the ABI32 range", source)
	}
	damageCalls := 0
	deleteCalls := 0
	blackPowderBurnUpdateNative53CCB0(source, blackPowderBurnUpdateDeps53CCB0{
		frame: func() uint32 { return math.MaxUint32 - 1 },
		fps: func() uint32 {
			t.Fatal("fuse initialization read FPS")
			return 0
		},
		damageUnitsAround: func(types.Pointf, float32, float32, int, object.DamageType, *Object, Obj) {
			damageCalls++
		},
		delayedDelete: func(*Object) { deleteCalls++ },
	})
	if source.Field34 != 1 || damageCalls != 0 || deleteCalls != 0 {
		t.Fatalf("fuse/damage/delete = %d/%d/%d, want 1/0/0", source.Field34, damageCalls, deleteCalls)
	}
}

func TestBlackPowderBurnUpdateNative53CCB0DamagePrecedesExpiry(t *testing.T) {
	source := &Object{Field32: 10, Field34: 100, PosVec: types.Ptf(12, 34)}
	damageCalls := 0
	deleteCalls := 0
	blackPowderBurnUpdateNative53CCB0(source, blackPowderBurnUpdateDeps53CCB0{
		frame: func() uint32 { return 100 },
		fps: func() uint32 {
			t.Fatal("scheduled damage read FPS")
			return 0
		},
		damageUnitsAround: func(pos types.Pointf, outer, inner float32, damage int, damageType object.DamageType, got *Object, excluded Obj) {
			damageCalls++
			if pos != source.PosVec || outer != 15 || inner != 15 || damage != 1 ||
				damageType != object.DamageFlame || got != source || excluded != nil {
				t.Fatalf("damage args = %v/%v/%v/%d/%v/%p/%v", pos, outer, inner, damage, damageType, got, excluded)
			}
		},
		delayedDelete: func(*Object) { deleteCalls++ },
	})
	if damageCalls != 1 || deleteCalls != 0 || source.Field34 != 100 {
		t.Fatalf("damage/delete/fuse = %d/%d/%d, want 1/0/100", damageCalls, deleteCalls, source.Field34)
	}
}

func TestBlackPowderBurnUpdateNative53CCB0ExpiryBoundaryAndWraparound(t *testing.T) {
	tests := []struct {
		name          string
		frame         uint32
		creationFrame uint32
		fps           uint32
		deleted       bool
	}{
		{name: "before boundary", frame: 159, creationFrame: 100, fps: 30},
		{name: "exact boundary", frame: 160, creationFrame: 100, fps: 30, deleted: true},
		{name: "after boundary", frame: 161, creationFrame: 100, fps: 30, deleted: true},
		{name: "frame wrap", frame: 1, creationFrame: math.MaxUint32 - 1, fps: 1, deleted: true},
		{name: "wrapped lifetime product", frame: 1, creationFrame: 0, fps: 0x80000000, deleted: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			source := &Object{Field32: tc.creationFrame, Field34: 77}
			deleteCalls := 0
			blackPowderBurnUpdateNative53CCB0(source, blackPowderBurnUpdateDeps53CCB0{
				frame: func() uint32 { return tc.frame },
				fps:   func() uint32 { return tc.fps },
				damageUnitsAround: func(types.Pointf, float32, float32, int, object.DamageType, *Object, Obj) {
					t.Fatal("unexpected damage")
				},
				delayedDelete: func(got *Object) {
					deleteCalls++
					if got != source {
						t.Fatalf("deleted source = %p, want %p", got, source)
					}
				},
			})
			if got := deleteCalls != 0; got != tc.deleted {
				t.Fatalf("delete calls = %d, want deleted=%t", deleteCalls, tc.deleted)
			}
		})
	}
}
