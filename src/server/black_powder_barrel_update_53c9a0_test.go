package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestBlackPowderBarrelUpdate53C9A0SchedulesFuseWithNativePointer(t *testing.T) {
	s := new(Server)
	s.SetFrame(math.MaxUint32 - 2)
	source := &Object{Field32: 77, Field34: 77}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(source)) <= math.MaxUint32 {
		t.Fatalf("source pointer = %p, want address above the ABI32 range", source)
	}
	s.BlackPowderBarrelUpdate53C9A0(source, BlackPowderBarrelUpdateRuntime53C9A0{
		RandomInt: func(minimum, maximum int) int {
			if minimum != 1 || maximum != 5 {
				t.Fatalf("random range = %d..%d, want 1..5", minimum, maximum)
			}
			return 5
		},
	})
	if source.Field34 != 2 {
		t.Fatalf("fuse frame = %d, want 2", source.Field34)
	}
}

func TestBlackPowderBarrelUpdate53C9A0ExplodesAndCreatesFlames(t *testing.T) {
	s := new(Server)
	s.SetFrame(105)
	s.SetTickRate(30)
	source := &Object{Field32: 100, Field34: 105, PosVec: types.Ptf(10, 20)}

	var damaged, pushed bool
	var typeNames []string
	var positions []types.Pointf
	var decays []uint32
	flames := [blackPowderBarrelFlameCount53C9A0]Object{}
	flameIndex := 0
	s.BlackPowderBarrelUpdate53C9A0(source, BlackPowderBarrelUpdateRuntime53C9A0{
		DamageUnitsAround: func(position types.Pointf, outer, inner float32, damage int, damageType object.DamageType, got *Object, excluded Obj) {
			damaged = true
			if position != source.PosVec || outer != 100 || inner != 30 || damage != 30 ||
				damageType != object.DamageExplosion || got != source || excluded != nil {
				t.Fatalf("damage args = %+v/%v/%v/%d/%v/%p/%v", position, outer, inner, damage, damageType, got, excluded)
			}
		},
		PushUnitsAround: func(position types.Pointf, outer, inner, force float32) {
			pushed = true
			if position != source.PosVec || outer != 100 || inner != 30 || force != 60 {
				t.Fatalf("push args = %+v/%v/%v/%v", position, outer, inner, force)
			}
		},
		RandomInt: func(minimum, maximum int) int {
			switch {
			case minimum == 0 && maximum == 1:
				return flameIndex & 1
			case minimum == 0 && maximum == 255:
				return 0
			case minimum == 5 && maximum == 20:
				return 7
			default:
				t.Fatalf("unexpected random range %d..%d", minimum, maximum)
				return 0
			}
		},
		RandomFloat: func(minimum, maximum float32) float32 {
			if minimum != 0 || maximum != 15 {
				t.Fatalf("random float range = %v..%v, want 0..15", minimum, maximum)
			}
			return 2
		},
		TraceRay: func(from, to types.Pointf) bool {
			if from != source.PosVec || to != (types.Ptf(22, 20)) {
				t.Fatalf("trace = %+v -> %+v, want %+v -> {22 20}", from, to, source.PosVec)
			}
			return true
		},
		NewObjectByTypeID: func(typeID string) *Object {
			typeNames = append(typeNames, typeID)
			return &flames[flameIndex]
		},
		CreateAt: func(flame, owner *Object, position types.Pointf) {
			if flame != &flames[flameIndex] || owner != nil {
				t.Fatalf("create object/owner = %p/%p, want %p/nil", flame, owner, &flames[flameIndex])
			}
			positions = append(positions, position)
		},
		SetDecayTime: func(flame *Object, frames uint32) {
			if flame != &flames[flameIndex] {
				t.Fatalf("decay object = %p, want %p", flame, &flames[flameIndex])
			}
			decays = append(decays, frames)
			flameIndex++
		},
		DelayedDelete: func(*Object) { t.Fatal("explosion tick deleted barrel") },
	})
	if !damaged || !pushed || flameIndex != blackPowderBarrelFlameCount53C9A0 {
		t.Fatalf("effects = damaged %t, pushed %t, flames %d", damaged, pushed, flameIndex)
	}
	wantTypes := []string{"SmallFlame", "MediumFlame", "SmallFlame", "MediumFlame"}
	for i := range wantTypes {
		if typeNames[i] != wantTypes[i] || positions[i] != (types.Ptf(22, 20)) || decays[i] != 210 {
			t.Fatalf("flame %d = %q/%+v/%d, want %q/{22 20}/210", i, typeNames[i], positions[i], decays[i], wantTypes[i])
		}
	}
}

func TestBlackPowderBarrelUpdate53C9A0LifetimeBoundaryAndWrap(t *testing.T) {
	s := new(Server)
	s.SetTickRate(30)
	source := &Object{Field32: 100, Field34: 101}
	deleted := 0
	runtime := BlackPowderBarrelUpdateRuntime53C9A0{
		DelayedDelete: func(got *Object) {
			if got != source {
				t.Fatalf("deleted = %p, want %p", got, source)
			}
			deleted++
		},
	}
	s.SetFrame(129)
	s.BlackPowderBarrelUpdate53C9A0(source, runtime)
	if deleted != 0 {
		t.Fatalf("deleted before boundary = %d", deleted)
	}
	s.SetFrame(130)
	s.BlackPowderBarrelUpdate53C9A0(source, runtime)
	if deleted != 1 {
		t.Fatalf("deleted at boundary = %d, want 1", deleted)
	}
	source.Field32 = math.MaxUint32 - 10
	source.Field34 = math.MaxUint32 - 9
	s.SetFrame(20)
	s.BlackPowderBarrelUpdate53C9A0(source, runtime)
	if deleted != 2 {
		t.Fatalf("deleted after frame wrap = %d, want 2", deleted)
	}
	s.BlackPowderBarrelUpdate53C9A0(nil, runtime)
}
