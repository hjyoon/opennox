package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestObjectSetModifierAttrsMatchesGAMEEXEContract(t *testing.T) {
	mods := [4]ModifierEff{}
	full := ModifierInitData{
		Modifiers: [4]*ModifierEff{&mods[0], nil, &mods[2], &mods[3]},
		Field16:   0x89ABCDEF,
	}
	empty := ModifierInitData{Field16: 0x13579BDF}
	tests := []struct {
		name     string
		class    object.Class
		subClass uint32
		typeInd  uint16
		teamBase uint32
		attrs    ModifierInitData
		applied  bool
	}{
		{name: "empty ordinary item is ignored", class: object.ClassWeapon, attrs: empty},
		{name: "forced wand accepts empty attributes", class: object.ClassWand, subClass: 0x00010000, attrs: empty, applied: true},
		{name: "ineligible object is ignored", class: object.ClassMissile, typeInd: 7, teamBase: 8, attrs: full},
		{name: "ordinary item", class: object.ClassArmor, attrs: full, applied: true},
		{name: "ordinary TeamBase", class: object.ClassMissile, typeInd: 9, teamBase: 9, attrs: full, applied: true},
		{name: "special TeamBase", class: object.ClassPlayer, typeInd: 10, teamBase: 10, attrs: full, applied: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dst := ModifierInitData{
				Modifiers: [4]*ModifierEff{nil, &mods[1], nil, &mods[0]},
				Field16:   0x2468ACE0,
			}
			obj := &Object{
				TypeInd:     tc.typeInd,
				ObjClass:    tc.class,
				ObjSubClass: object.SubClass(tc.subClass),
				ObjFlags:    object.FlagActive,
				Field37:     0x80000015,
				Field38:     0x2468ACE0,
				InitData:    unsafe.Pointer(&dst),
			}
			for i := range obj.Field140 {
				obj.Field140[i] = 0xA5000000 | 0x02000000 | 0x200 | uint32(i<<12)
			}
			before := *obj
			beforeAttrs := dst
			if tc.applied && !tc.class.HasAny(object.ClassPlayer|object.ClassImmobile|object.ClassClientPersist) {
				bindModifierTestType(t, obj, tc.typeInd)
				before.serverHandle = obj.serverHandle
				before.TypeInd = obj.TypeInd
			}

			got := obj.SetModifierAttrs(&tc.attrs, tc.teamBase)
			if got != tc.applied {
				t.Fatalf("applied: got %v, want %v", got, tc.applied)
			}
			if !tc.applied {
				if *obj != before || dst != beforeAttrs {
					t.Fatal("rejected modifier update changed object state")
				}
				return
			}
			if dst != tc.attrs {
				t.Errorf("attributes: got %#v, want %#v", dst, tc.attrs)
			}
			if obj.Field38 != math.MaxUint32 {
				t.Errorf("sync field: got %#08x, want %#08x", obj.Field38, uint32(math.MaxUint32))
			}
			wantSlots := before.Field140
			applyGAMEEXEModifierSlots(&before, wantSlots[:])
			if obj.Field140 != wantSlots {
				t.Errorf("Field140 differs: got %#v, want %#v", obj.Field140, wantSlots)
			}
			if obj.Field37 != before.Field37 || obj.ObjClass != before.ObjClass || obj.ObjSubClass != before.ObjSubClass || obj.ObjFlags != before.ObjFlags || obj.TypeInd != before.TypeInd || obj.InitData != before.InitData {
				t.Fatal("modifier update overwrote state outside the original function contract")
			}
		})
	}
}

func bindModifierTestType(t *testing.T, obj *Object, typeInd uint16) {
	t.Helper()
	bindStatusTestType(t, obj, 0)
	if typeInd != 0 {
		obj.TypeInd = typeInd
		s := obj.Server()
		for len(s.Types.byInd) <= int(typeInd) {
			s.Types.byInd = append(s.Types.byInd, nil)
		}
		s.Types.byInd[typeInd] = &ObjectType{}
	}
}

func applyGAMEEXEModifierSlots(obj *Object, slots []uint32) {
	if obj.Class().HasAny(object.ClassPlayer | object.ClassImmobile | object.ClassClientPersist) {
		for i := range slots {
			slots[i] = slots[i]&0xFFFFF000 | 0x02000000
		}
		return
	}
	changed := obj.Class().HasAny(object.ClassFlag | object.ClassWeapon | object.ClassArmor | object.ClassWand)
	for i := range slots {
		if changed {
			slots[i] |= 0x200
		} else {
			slots[i] &^= 0x200
		}
		if obj.Field37&uint32(1<<i) != 0 {
			slots[i] |= 0x02000000
		} else if slots[i]&0x200 == 0 {
			slots[i] &^= 0x02000000
		}
	}
}
