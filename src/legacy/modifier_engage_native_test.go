package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

func TestEnchantItemInventoryUsesNativeObjectAndModifierLayouts(t *testing.T) {
	owner, freeOwner := alloc.NewHigh(server.Object{})
	defer freeOwner()
	ignored, freeIgnored := alloc.NewHigh(server.Object{})
	defer freeIgnored()
	equipped, freeEquipped := alloc.NewHigh(server.Object{})
	defer freeEquipped()
	attrs, freeAttrs := alloc.NewHigh(server.ModifierInitData{})
	defer freeAttrs()
	modifier, freeModifier := alloc.NewHigh(server.ModifierEff{})
	defer freeModifier()

	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(owner)) <= uintptr(^uint32(0)) {
		t.Fatalf("owner pointer %#x did not exercise the native 64-bit range", uintptr(unsafe.Pointer(owner)))
	}
	owner.InvFirstItem = ignored
	ignored.InvNextItem = equipped
	equipped.ObjClass = object.ClassArmor
	equipped.ObjFlags = object.FlagEquipped
	equipped.InitData = unsafe.Pointer(attrs)
	attrs.Modifiers[2] = modifier
	modifier.Engage112 = modifierEngagePointerNative4DFBB0(8)

	if !enchantItemTestInventoryNative4DFBB0(owner, 8) {
		t.Fatal("equipped native modifier was not found")
	}
	if enchantItemTestInventoryNative4DFBB0(owner, 4) {
		t.Fatal("different engage effect matched")
	}
	equipped.ObjFlags &^= object.FlagEquipped
	if enchantItemTestInventoryNative4DFBB0(owner, 8) {
		t.Fatal("unequipped modifier matched")
	}
	if enchantItemTestInventoryNative4DFBB0(nil, 8) || enchantItemTestInventoryNative4DFBB0(owner, 0) {
		t.Fatal("nil owner or unknown flag matched")
	}
}

func TestEquipmentProtectionDisengageKeepsRemainingModifiers(t *testing.T) {
	objectStatusTestServer(t, 1)
	owner, freeOwner := alloc.NewHigh(server.Object{})
	t.Cleanup(freeOwner)
	item, freeItem := alloc.NewHigh(server.Object{})
	t.Cleanup(freeItem)
	attrs, freeAttrs := alloc.NewHigh(server.ModifierInitData{})
	t.Cleanup(freeAttrs)
	remaining, freeRemaining := alloc.NewHigh(server.ModifierEff{})
	t.Cleanup(freeRemaining)
	*owner = server.Object{ObjClass: object.ClassPlayer, InvFirstItem: item}
	*item = server.Object{ObjClass: object.ClassWand, InitData: unsafe.Pointer(attrs)}
	attrs.Modifiers[3] = remaining

	const otherFlags = 0x40000500
	for _, flag := range []byte{1, 2, 4, 8, 32} {
		on := modifierEngagePointerNative4DFBB0(flag)
		off := modifierDisengagePointerNative4DFBB0(flag)
		remaining.Engage112 = on
		owner.Field110 = otherFlags
		ccall.CallVoidPtr3(on, nil, owner.CObj(), item.CObj())
		wantSet := otherFlags | uint32(flag)
		if owner.Field110 != wantSet {
			t.Fatalf("engaging %#x produced %#x, want %#x", flag, owner.Field110, wantSet)
		}
		item.ObjFlags = object.FlagEquipped
		ccall.CallVoidPtr3(off, nil, owner.CObj(), item.CObj())
		if owner.Field110 != wantSet {
			t.Fatalf("disengaging %#x with an equipped modifier produced %#x, want %#x", flag, owner.Field110, wantSet)
		}
		item.ObjFlags = 0
		ccall.CallVoidPtr3(off, nil, owner.CObj(), item.CObj())
		if owner.Field110 != otherFlags {
			t.Fatalf("disengaging the last %#x modifier produced %#x, want %#x", flag, owner.Field110, otherFlags)
		}
	}
	owner.Field110 = 1 | 32
	ccall.CallVoidPtr3(modifierDisengagePointerNative4DFBB0(1), nil, owner.CObj(), nil)
	owner.ObjClass = object.ClassMonster
	ccall.CallVoidPtr3(modifierDisengagePointerNative4DFBB0(32), nil, owner.CObj(), item.CObj())
	if owner.Field110 != 1|32 {
		t.Fatalf("guarded fire and regeneration callbacks wrote flags %#x", owner.Field110)
	}
}

func TestEquipmentProtectionGoHeapEngageAndDisengage(t *testing.T) {
	objectStatusTestServer(t, 1)
	const flag = 4
	effect := &server.ModifierEff{
		Engage112:    modifierEngagePointerNative4DFBB0(flag),
		Disengage116: modifierDisengagePointerNative4DFBB0(flag),
	}
	attrs := &server.ModifierInitData{Modifiers: [4]*server.ModifierEff{2: effect}}
	second := &server.Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(attrs)}
	first := &server.Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(attrs), InvNextItem: second}
	owner := &server.Object{ObjClass: object.ClassPlayer, InvFirstItem: first, Field110: 0x40000000}

	server.ItemApplyEngageEffect4F2FF0(first, owner)
	if owner.Field110 != 0x40000000|flag {
		t.Fatalf("engage flags=%#x", owner.Field110)
	}
	first.ObjFlags &^= object.FlagEquipped
	server.ItemApplyDisengageEffect4F3030(first, owner)
	if owner.Field110 != 0x40000000|flag {
		t.Fatalf("remaining item flags=%#x", owner.Field110)
	}
	second.ObjFlags &^= object.FlagEquipped
	server.ItemApplyDisengageEffect4F3030(second, owner)
	if owner.Field110 != 0x40000000 {
		t.Fatalf("final disengage flags=%#x", owner.Field110)
	}
}
