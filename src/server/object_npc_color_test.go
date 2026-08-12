package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestObjectSetNPCColorMatchesGAMEEXEContract(t *testing.T) {
	if got := unsafe.Sizeof(Color3{}); got != 3 {
		t.Fatalf("Color3 size: got %d, want 3", got)
	}
	tests := []struct {
		name     string
		class    object.Class
		subClass object.SubClass
		index    byte
		changed  bool
	}{
		{name: "client persist", class: object.ClassMonster | object.ClassClientPersist, index: 0},
		{name: "immobile", class: object.ClassMonster | object.ClassImmobile, index: 5},
		{name: "player", class: object.ClassMonster | object.ClassPlayer, index: 2},
		{name: "ordinary NPC", class: object.ClassMonster, subClass: object.SubClass(object.MonsterNPC), index: 1, changed: true},
		{name: "ordinary female NPC", class: object.ClassMonster, subClass: object.SubClass(object.MonsterFemaleNPC), index: 3, changed: true},
		{name: "ordinary monster", class: object.ClassMonster, index: 4},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ud := &MonsterUpdateData{
				Field518:   0x13579BDF,
				Field523_2: 0x24,
				Field523_3: 0x68,
			}
			for i := range ud.Color {
				ud.Color[i] = Color3{R: byte(0x10 + i), G: byte(0x20 + i), B: byte(0x30 + i)}
			}
			got := &Object{
				ObjClass:    tc.class,
				ObjSubClass: tc.subClass,
				ObjFlags:    object.FlagActive,
				Field37:     0x80000015,
				Field38:     0x2468ACE0,
				UpdateData:  unsafe.Pointer(ud),
			}
			for i := range got.Field140 {
				got.Field140[i] = 0xA5000000 | 0x04000000 | 0x400 | uint32(i<<12)
			}
			if !tc.class.HasAny(object.ClassClientPersist | object.ClassImmobile | object.ClassPlayer) {
				bindStatusTestType(t, got, 0)
			}
			before := *got
			wantColors := ud.Color
			color := Color3{R: 0xAB, G: 0xCD, B: 0xEF}
			wantColors[tc.index] = color
			wantSlots := before.Field140
			applyGAMEEXENPCColorSlots(&before, wantSlots[:], tc.changed)

			got.Nox_xxx_setNPCColor_4E4A90(tc.index, &color)

			if ud.Color != wantColors {
				t.Errorf("colors: got %#v, want %#v", ud.Color, wantColors)
			}
			if ud.Field518 != 0x13579BDF || ud.Field523_2 != 0x24 || ud.Field523_3 != 0x68 {
				t.Fatal("color copy overwrote adjacent NPC update data")
			}
			if got.Field38 != math.MaxUint32 {
				t.Errorf("sync field: got %#08x, want %#08x", got.Field38, uint32(math.MaxUint32))
			}
			if got.Field140 != wantSlots {
				t.Errorf("Field140 differs: got %#v, want %#v", got.Field140, wantSlots)
			}
			if got.Field37 != before.Field37 || got.ObjClass != before.ObjClass || got.ObjSubClass != before.ObjSubClass ||
				got.ObjFlags != before.ObjFlags || got.TypeInd != before.TypeInd || got.UpdateData != before.UpdateData {
				t.Fatal("NPC color update overwrote state outside the original function contract")
			}
		})
	}
}

func applyGAMEEXENPCColorSlots(obj *Object, slots []uint32, changed bool) {
	if obj.Class().HasAny(object.ClassPlayer | object.ClassImmobile | object.ClassClientPersist) {
		for i := range slots {
			slots[i] = slots[i]&0xFFFFF000 | 0x04000000
		}
		return
	}
	for i := range slots {
		if changed {
			slots[i] |= 0x400
		} else {
			slots[i] &^= 0x400
		}
		if obj.Field37&uint32(1<<i) != 0 {
			slots[i] |= 0x04000000
		} else if slots[i]&0x400 == 0 {
			slots[i] &^= 0x04000000
		}
	}
}
