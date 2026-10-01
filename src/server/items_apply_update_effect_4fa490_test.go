package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestItemsApplyUpdateEffect4FA490LiveTraversalAndItemArgument(t *testing.T) {
	var token byte
	fn := unsafe.Pointer(&token)
	mods := [4]*ModifierEff{}
	for i := range mods {
		mods[i] = &ModifierEff{Update100: ModifierEffFnc{Fnc: fn}}
	}
	data := &ModifierInitData{Modifiers: [4]*ModifierEff{mods[0], nil, mods[2], nil}}
	first := &Object{ObjClass: object.ClassWeapon, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(data)}
	oldNext := &Object{ObjClass: object.ClassWeapon, ObjFlags: object.FlagEquipped}
	last := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, nil, mods[3]}})}
	unequipped := &Object{ObjClass: object.ClassWeapon, InvNextItem: first}
	wrongClass := &Object{ObjClass: object.ClassPlayer, ObjFlags: object.FlagEquipped, InvNextItem: unequipped}
	owner := &Object{InvFirstItem: wrongClass}
	first.InvNextItem = oldNext
	var calls []*ModifierEff
	var items []*Object
	itemsApplyUpdateEffect4FA490(owner, func(gotfn unsafe.Pointer, mod *ModifierEff, item *Object) {
		if gotfn != fn || item == owner {
			t.Fatal("wrong function or owner substituted for item")
		}
		calls = append(calls, mod)
		items = append(items, item)
		if mod == mods[0] {
			first.InitData = nil
			data.Modifiers[1] = mods[1]
			mods[2].Update100.Fnc = nil
			first.InvNextItem = last
			owner.InvFirstItem = nil
		}
	})
	if !reflect.DeepEqual(calls, []*ModifierEff{mods[0], mods[1], mods[3]}) || !reflect.DeepEqual(items, []*Object{first, first, last}) {
		t.Fatalf("calls/items=%v/%v", calls, items)
	}
}

func TestItemsApplyUpdateEffect4FA490RequiredInputs(t *testing.T) {
	for _, tc := range []struct {
		name  string
		owner *Object
	}{
		{"owner", nil},
		{"equipped modifier array", &Object{InvFirstItem: &Object{ObjClass: object.ClassWeapon, ObjFlags: object.FlagEquipped}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("required input silently skipped")
				}
			}()
			itemsApplyUpdateEffect4FA490(tc.owner, nil)
		})
	}
	itemsApplyUpdateEffect4FA490(&Object{}, nil)
}
