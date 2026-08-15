package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

func TestNoxXxxRespawnAdd4EC5E0NativeRecord(t *testing.T) {
	handles.Init()
	t.Cleanup(handles.Release)

	oldAllow := respawnAddLoadAllow4EC5E0()
	oldAllocator := respawnAddLoadAllocator4EC5E0()
	oldHead := respawnAddLoadHead4EC5E0()
	oldWeaponFlags := objectNPCWeaponEquipFlags

	class := alloc.NewClass("RespawnAdd4EC5E0Test", unsafe.Sizeof(respawnRecord4EC5E0{}), 3)
	defer func() {
		respawnAddStoreHead4EC5E0(oldHead)
		respawnAddStoreAllocator4EC5E0(oldAllocator)
		respawnAddStoreAllow4EC5E0(oldAllow)
		objectNPCWeaponEquipFlags = oldWeaponFlags
		class.Free()
	}()

	obj, freeObj := alloc.New(server.Object{})
	defer freeObj()
	modifier, freeModifier := alloc.New(server.ModifierEff{})
	defer freeModifier()
	attrs, freeAttrs := alloc.New(server.ModifierInitData{})
	defer freeAttrs()
	useData, freeUseData := alloc.New([2]uint8{})
	defer freeUseData()

	attrs.Modifiers[0] = modifier
	attrs.Field16 = 0x89abcdef
	useData[0] = 0x34
	useData[1] = 0x56
	obj.TypeInd = 0xfedc
	obj.ObjClass = 0x01001000
	obj.PosVec.X = math.Float32frombits(0x7fa54321)
	obj.PosVec.Y = math.Float32frombits(0x80000000)
	obj.Direction1 = server.Dir16(0xf123)
	obj.InitData = unsafe.Pointer(attrs)
	obj.UseData.Ptr = unsafe.Pointer(useData)

	previous := (*respawnRecord4EC5E0)(class.NewObject())
	previous.TypeInd = 0x12345678
	respawnAddStoreAllow4EC5E0(1)
	respawnAddStoreAllocator4EC5E0(class.UPtr())
	respawnAddStoreHead4EC5E0(previous)

	weaponCalls := 0
	objectNPCWeaponEquipFlags = func(got *server.Object) uint32 {
		weaponCalls++
		if got != obj {
			t.Fatalf("weapon object = %p, want %p", got, obj)
		}
		return 0xffff0082
	}

	Nox_xxx_respawnAdd_4EC5E0(obj)

	rec := respawnAddLoadHead4EC5E0()
	if rec == nil || rec == previous {
		t.Fatalf("new head = %p, previous = %p", rec, previous)
	}
	if weaponCalls != 1 {
		t.Fatalf("weapon calls = %d, want 1", weaponCalls)
	}
	if rec.TypeInd != uint32(obj.TypeInd) || rec.Object != obj {
		t.Fatalf("record identity = (%#x, %p), want (%#x, %p)", rec.TypeInd, rec.Object, obj.TypeInd, obj)
	}
	if got := math.Float32bits(rec.X); got != 0x7fa54321 {
		t.Fatalf("x bits = %#x, want %#x", got, uint32(0x7fa54321))
	}
	if got := math.Float32bits(rec.Y); got != 0x80000000 {
		t.Fatalf("y bits = %#x, want %#x", got, uint32(0x80000000))
	}
	if rec.Direction != uint16(obj.Direction1) {
		t.Fatalf("direction = %#x, want %#x", rec.Direction, obj.Direction1)
	}
	if rec.Reserved18 != 0 || rec.RespawnAt != 0 || rec.Pending != 0 || rec.Reserved50 != 0 {
		t.Fatalf("zero fields = (%#x, %#x, %#x, %#x)", rec.Reserved18, rec.RespawnAt, rec.Pending, rec.Reserved50)
	}
	if rec.Attrs.Modifiers != attrs.Modifiers || rec.Attrs.Field16 != attrs.Field16 {
		t.Fatalf("attrs = %#v/%#x, want %#v/%#x", rec.Attrs.Modifiers, rec.Attrs.Field16, attrs.Modifiers, attrs.Field16)
	}
	if rec.Charge1 != useData[1] || rec.Charge0 != useData[0] {
		t.Fatalf("charges = (%#x, %#x), want (%#x, %#x)", rec.Charge1, rec.Charge0, useData[1], useData[0])
	}
	if rec.Prev != nil || rec.Next != previous || previous.Prev != rec {
		t.Fatalf("links = prev:%p next:%p previous.prev:%p", rec.Prev, rec.Next, previous.Prev)
	}
}
