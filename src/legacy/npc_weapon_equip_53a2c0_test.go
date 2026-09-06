package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/server"
)

func TestNPCWeaponEquip53A2C0KeepsNativePointersThroughPlayerEquip(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("native-width routing regression applies to 64-bit builds")
	}

	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	oldGetServer := GetServer
	GetServer = func() Server { return &playerAttackLegacyServer538960{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })

	oldWeaponFlags := objectNPCWeaponEquipFlags
	objectNPCWeaponEquipFlags = func(item *server.Object) uint32 {
		if item == nil {
			t.Fatal("NPC equip weapon lookup received nil")
		}
		if !item.Class().Has(object.ClassWeapon) {
			t.Fatalf("NPC equip weapon lookup = %p class %#x", item, item.ObjClass)
		}
		return uint32(object.WeaponBow)
	}
	t.Cleanup(func() { objectNPCWeaponEquipFlags = oldWeaponFlags })

	update := &server.MonsterUpdateData{
		Field516: 0xec213490,
		Field517: 0x11223344,
	}
	owner := &server.Object{
		ObjClass:    object.ClassMonster,
		ObjSubClass: object.SubClass(object.MonsterNPC),
		UpdateData:  unsafe.Pointer(update),
	}
	attrs := &server.ModifierInitData{}
	item := &server.Object{
		TypeInd:     0x3401,
		ObjClass:    object.ClassWeapon,
		ObjSubClass: object.SubClass(object.WeaponBow),
		InvHolder:   owner,
		InitData:    unsafe.Pointer(attrs),
	}
	owner.InvFirstItem = item

	var pin runtime.Pinner
	for index, pointer := range []unsafe.Pointer{
		unsafe.Pointer(owner), unsafe.Pointer(update), unsafe.Pointer(item), unsafe.Pointer(attrs),
	} {
		pin.Pin(pointer)
		if uintptr(pointer) <= math.MaxUint32 {
			t.Fatalf("NPC equip pointer %d = %p, want address above ABI32", index, pointer)
		}
	}
	defer pin.Unpin()

	// This enters the original public C equip routine before crossing the new
	// typed callback. The former `(int)(intptr_t)owner` handoff truncated this
	// exact high address before GAME.EXE read owner +748.
	if got := playerEquipWeaponNativeCall53A420(owner, item); got != 1 {
		t.Fatalf("NPC weapon equip result = %d, want 1", got)
	}
	if !item.Flags().Has(object.FlagEquipped) {
		t.Fatal("NPC bow is not equipped")
	}
	if update.WeaponEquipFlags != uint32(object.WeaponBow) {
		t.Fatalf("NPC weapon appearance = %#x, want %#x",
			update.WeaponEquipFlags, uint32(object.WeaponBow))
	}
	if update.Field516 != 0 {
		t.Fatalf("NPC compatibility weapon word = %#x, want zero on 64-bit", update.Field516)
	}
	if update.Field517 != 0x11223300 {
		t.Fatalf("NPC animation state = %#x, want %#x", update.Field517, uint32(0x11223300))
	}
	if owner.Field38 != math.MaxUint32 {
		t.Fatalf("NPC sync field = %#x, want MaxUint32", owner.Field38)
	}
	runtime.KeepAlive(owner)
	runtime.KeepAlive(update)
	runtime.KeepAlive(item)
	runtime.KeepAlive(attrs)
}
