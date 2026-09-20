package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestDeathBallFragmentUpdateNative53D220UsesNativeObjectAndOriginalOrder(t *testing.T) {
	source := &Object{Field32: 100, PosVec: types.Ptf(12, 34)}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(source)) <= 0xffffffff {
		t.Fatalf("fragment pointer = %p, want above 4 GiB", source)
	}
	events := make([]string, 0, 4)
	deathBallFragmentUpdateNative53D220(source, deathBallFragmentUpdateNativeDeps53D220{
		frame: func() uint32 {
			events = append(events, "frame")
			return 161
		},
		fps: func() uint32 {
			events = append(events, "fps")
			return 30
		},
		damageUnitsAround: func(pos types.Pointf, outerRadius, innerRadius float32, damage int, damageType object.DamageType, got *Object, excluded Obj) {
			events = append(events, "damage")
			if pos != source.PosVec || outerRadius != 25 || innerRadius != 0 || damage != 20 ||
				damageType != object.DamageCrush || got != source || excluded != nil {
				t.Fatalf("damage args = %v/%v/%v/%v/%v/%p/%v", pos, outerRadius, innerRadius, damage, damageType, got, excluded)
			}
		},
		delayedDelete: func(got *Object) {
			events = append(events, "delete")
			if got != source {
				t.Fatalf("deleted fragment = %p, want %p", got, source)
			}
		},
	})
	if want := []string{"frame", "damage", "fps", "delete"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestDeathBallFragmentUpdateNative53D220BoundariesAndMixedSignedness(t *testing.T) {
	tests := []struct {
		name          string
		frame         uint32
		creationFrame uint32
		fps           uint32
		damage        bool
		deleted       bool
	}{
		{name: "damage one tick early", frame: 110, creationFrame: 100, fps: 30},
		{name: "damage starts strictly after ten", frame: 111, creationFrame: 100, fps: 30, damage: true},
		{name: "lifetime exact boundary", frame: 160, creationFrame: 100, fps: 30, damage: true},
		{name: "lifetime expires strictly after boundary", frame: 161, creationFrame: 100, fps: 30, damage: true, deleted: true},
		{name: "future creation is signed for damage and unsigned for expiry", frame: 10, creationFrame: 20, fps: 30, deleted: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			source := &Object{Field32: tc.creationFrame}
			damageCalls := 0
			deleteCalls := 0
			deathBallFragmentUpdateNative53D220(source, deathBallFragmentUpdateNativeDeps53D220{
				frame: func() uint32 { return tc.frame },
				fps:   func() uint32 { return tc.fps },
				damageUnitsAround: func(types.Pointf, float32, float32, int, object.DamageType, *Object, Obj) {
					damageCalls++
				},
				delayedDelete: func(*Object) { deleteCalls++ },
			})
			if got := damageCalls != 0; got != tc.damage {
				t.Fatalf("damage calls = %d, want damage=%t", damageCalls, tc.damage)
			}
			if got := deleteCalls != 0; got != tc.deleted {
				t.Fatalf("delete calls = %d, want deleted=%t", deleteCalls, tc.deleted)
			}
		})
	}
}
