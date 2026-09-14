package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestChannelLife52F460NativeBindingUsesWidePointersAndFloatBits(t *testing.T) {
	s := &Server{}
	targetData := &PlayerUpdateData{ManaCur: 4, ManaMax: 10}
	target := &Object{ObjClass: object.Class(4), UpdateData: unsafe.Pointer(targetData)}
	caster := &Object{ObjClass: object.Class(4), HealthData: &HealthData{Cur: 3}}
	r := &DurSpell{Target48: target, Caster16: caster, Level: 2, Field72: int32(math.Float32bits(0.25))}
	r.Pos.X = math.Float32frombits(0x3fdccccc) // Bad PE32 target pointer in the crash.
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(target)) <= 0xffffffff {
		t.Fatal("expected target above 4 GiB")
	}
	var addCalls, clearCalls int
	runtime := SpellChannelLifeRuntime52F460{
		AddMana: func(got *Object, amount int16) {
			addCalls++
			if got != target || amount != 1 {
				t.Fatalf("add = %p/%d", got, amount)
			}
		},
		ClearDamage: func(got *Object, amount int32) {
			clearCalls++
			if got != caster || amount != 1 {
				t.Fatalf("clear = %p/%d", got, amount)
			}
		},
		Coefficient: func(index uint32) float64 {
			if index != 1 {
				t.Fatalf("coefficient index = %d", index)
			}
			return 0.5
		},
	}
	if got := s.SpellChannelLifeUpdate52F460(r, runtime); got != 0 {
		t.Fatalf("update = %d, want 0", got)
	}
	if addCalls != 1 || clearCalls != 1 || math.Float32frombits(uint32(r.Field72)) != -0.25 {
		t.Fatalf("effects = %d/%d, fraction = %g", addCalls, clearCalls, math.Float32frombits(uint32(r.Field72)))
	}
}
