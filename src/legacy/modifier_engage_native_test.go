package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestEnchantItemInventoryUsesNativeObjectAndModifierLayouts(t *testing.T) {
	owner, freeOwner := alloc.New(server.Object{})
	defer freeOwner()
	ignored, freeIgnored := alloc.New(server.Object{})
	defer freeIgnored()
	equipped, freeEquipped := alloc.New(server.Object{})
	defer freeEquipped()
	attrs, freeAttrs := alloc.New(server.ModifierInitData{})
	defer freeAttrs()
	modifier, freeModifier := alloc.New(server.ModifierEff{})
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
