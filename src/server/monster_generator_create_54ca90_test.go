package server

import (
	"math"
	"testing"
	"unsafe"
)

func TestMonsterGeneratorCreateNative54CA90UsesNativeLayout(t *testing.T) {
	first := new(Object)
	last := new(Object)
	healthSamples := [32]uint16{0: 0x9696, 1: 0x9696, 31: 0x9f9f}
	update := &MonsterGenUpdateData{
		Field0:          [12]*Object{0: first, 11: last},
		Field48:         0x48484848,
		FuncInd52:       0x52525252,
		Field56:         0x56565656,
		FuncInd60:       0x60606060,
		Field64:         0x64646464,
		FuncInd68:       0x68686868,
		ScriptCollision: ScriptCallback{Flags: 0x72727272, Func: 0x76767676},
		Field92:         0x92929292,
		HealthSamples:   healthSamples,
	}
	obj := &Object{UpdateData: unsafe.Pointer(update)}

	MonsterGeneratorCreateNative54CA90(obj)

	if update.Field92 != 2 || update.FuncInd52 != math.MaxUint32 ||
		update.FuncInd60 != math.MaxUint32 || update.FuncInd68 != math.MaxUint32 ||
		update.ScriptCollision.Func != -1 {
		t.Fatalf("initialized fields = Field92:%d Func52:%#x Func60:%#x Func68:%#x Collision:%d",
			update.Field92, update.FuncInd52, update.FuncInd60, update.FuncInd68, update.ScriptCollision.Func)
	}
	if update.Field0[0] != first || update.Field0[11] != last ||
		update.Field48 != 0x48484848 || update.Field56 != 0x56565656 ||
		update.Field64 != 0x64646464 || update.ScriptCollision.Flags != 0x72727272 ||
		update.HealthSamples != healthSamples {
		t.Fatalf("neighboring native-width fields were modified: %+v", update)
	}
}

func TestRegisterObjectCreateGoDispatchesNativeObject(t *testing.T) {
	token := new(byte)
	callback := unsafe.Pointer(token)
	const name = "TestNativeCreate54CA90"
	want := new(Object)
	var got *Object
	RegisterObjectCreateGo(name, callback, func(obj *Object) {
		got = obj
	})

	CallObjectCreate(callback, want)
	if got != want {
		t.Fatalf("create object = %p, want %p", got, want)
	}
}
