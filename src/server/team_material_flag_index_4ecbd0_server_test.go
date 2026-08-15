package server

import (
	"testing"
	"unsafe"
)

func TestTeamMaterialObjectIndex4ECBD0NativeLayout(t *testing.T) {
	wantObjectClass := uintptr(8)
	wantObjectInitData := uintptr(692)
	wantInitDataSize := uintptr(20)
	wantSecondModifier := uintptr(4)
	wantModifierSize := uintptr(144)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectClass = 12
		wantObjectInitData = 760
		wantInitDataSize = 40
		wantSecondModifier = 8
		wantModifierSize = 208
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantObjectClass},
		{"Object.InitData", unsafe.Offsetof(Object{}.InitData), wantObjectInitData},
		{"ModifierInitData size", unsafe.Sizeof(ModifierInitData{}), wantInitDataSize},
		{"ModifierInitData.Modifiers", unsafe.Offsetof(ModifierInitData{}.Modifiers), 0},
		{"ModifierInitData.Modifiers[1]", unsafe.Sizeof((*ModifierEff)(nil)), wantSecondModifier},
		{"ModifierEff size", unsafe.Sizeof(ModifierEff{}), wantModifierSize},
		{"ModifierEff.name0", unsafe.Offsetof(ModifierEff{}.name0), 0},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}
