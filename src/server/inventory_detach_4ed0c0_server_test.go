package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestInventoryDetach4ED0C0NativeLayout(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) == 4 {
		for _, tc := range []struct {
			name string
			got  uintptr
			want uintptr
		}{
			{"Object.size", unsafe.Sizeof(Object{}), 780},
			{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), 8},
			{"Object.ObjSubClass", unsafe.Offsetof(Object{}.ObjSubClass), 12},
			{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), 16},
			{"Object.Weight", unsafe.Offsetof(Object{}.Weight), 488},
			{"Object.CarryCapacity", unsafe.Offsetof(Object{}.CarryCapacity), 490},
			{"Object.InvHolder", unsafe.Offsetof(Object{}.InvHolder), 492},
			{"Object.InvNextItem", unsafe.Offsetof(Object{}.InvNextItem), 496},
			{"Object.Field125", unsafe.Offsetof(Object{}.Field125), 500},
			{"Object.InvFirstItem", unsafe.Offsetof(Object{}.InvFirstItem), 504},
			{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), 748},
			{"PlayerUpdateData.size", unsafe.Sizeof(PlayerUpdateData{}), 556},
			{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), 276},
			{"Player.size", unsafe.Sizeof(Player{}), 4828},
			{"Player.WeaponEquip", unsafe.Offsetof(Player{}.WeaponEquip), 4},
			{"Player.PlayerInd", unsafe.Offsetof(Player{}.PlayerInd), 2064},
			{"Player.Field3656", unsafe.Offsetof(Player{}.Field3656), 3656},
			{"Player.Prot4632", unsafe.Offsetof(Player{}.Prot4632), 4632},
		} {
			if tc.got != tc.want {
				t.Errorf("%s = %d, want %d", tc.name, tc.got, tc.want)
			}
		}
		return
	}
	for _, tc := range []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.size", unsafe.Sizeof(Object{}), 928},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), 12},
		{"Object.ObjSubClass", unsafe.Offsetof(Object{}.ObjSubClass), 16},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), 20},
		{"Object.Weight", unsafe.Offsetof(Object{}.Weight), 516},
		{"Object.CarryCapacity", unsafe.Offsetof(Object{}.CarryCapacity), 518},
		{"Object.InvHolder", unsafe.Offsetof(Object{}.InvHolder), 520},
		{"Object.InvNextItem", unsafe.Offsetof(Object{}.InvNextItem), 528},
		{"Object.Field125", unsafe.Offsetof(Object{}.Field125), 536},
		{"Object.InvFirstItem", unsafe.Offsetof(Object{}.InvFirstItem), 544},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), 872},
		{"PlayerUpdateData.size", unsafe.Sizeof(PlayerUpdateData{}), 640},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), 320},
		{"Player.size", unsafe.Sizeof(Player{}), 6160},
		{"Player.WeaponEquip", unsafe.Offsetof(Player{}.WeaponEquip), 4},
		{"Player.PlayerInd", unsafe.Offsetof(Player{}.PlayerInd), 2068},
		{"Player.Field3656", unsafe.Offsetof(Player{}.Field3656), 4952},
		{"Player.Prot4632", unsafe.Offsetof(Player{}.Prot4632), 5936},
	} {
		if tc.got != tc.want {
			t.Errorf("%s = %d, want %d", tc.name, tc.got, tc.want)
		}
	}
}

func TestInventoryDetachNative4ED0C0UsesNativePointersAndLivePostClearState(t *testing.T) {
	initialPlayer := &Player{WeaponEquip: 7, PlayerInd: 3, Prot4632: 91}
	initialUpdate := &PlayerUpdateData{Player: initialPlayer}
	postPlayer := &Player{}
	postUpdate := &PlayerUpdateData{Player: postPlayer}
	remaining := &Object{Weight: 7}
	owner := &Object{
		ObjClass:      object.ClassPlayer,
		CarryCapacity: 6,
		UpdateData:    unsafe.Pointer(initialUpdate),
	}
	previous := &Object{}
	next := &Object{Field125: previous}
	item := &Object{
		ObjClass:    object.ClassFlag,
		InvHolder:   owner,
		InvNextItem: next,
		Field125:    previous,
	}
	previous.InvNextItem = item
	owner.InvFirstItem = previous

	var events []string
	detachInventoryNative4ED0C0(owner, item, inventoryDetachNativeDeps4ED0C0{
		gameFlag: func(flag uint32) uint32 {
			events = append(events, "flag:"+inventoryDetachUint4ED0C0(flag))
			if flag == inventoryDetachOnlineFlag4ED0C0 {
				return 1
			}
			return 0
		},
		netReportDequip: func(index uint8, got *Object) {
			events = append(events, "dequip:"+inventoryDetachUint4ED0C0(uint32(index)))
			if got != item {
				t.Fatalf("dequip item = %p, want %p", got, item)
			}
		},
		dequipArmor: func(gotOwner, gotItem *Object, mode, report int32) {
			events = append(events, "armor:"+inventoryDetachInt4ED0C0(mode)+":"+inventoryDetachInt4ED0C0(report))
			if gotOwner != owner || gotItem != item {
				t.Fatalf("armor pointers = %p/%p, want %p/%p", gotOwner, gotItem, owner, item)
			}
		},
		dequipWeapon: func(gotOwner, gotItem *Object, mode, report int32) {
			events = append(events, "weapon:"+inventoryDetachInt4ED0C0(mode)+":"+inventoryDetachInt4ED0C0(report))
		},
		netReportDrop: func(index uint8, got *Object) {
			events = append(events, "drop:"+inventoryDetachUint4ED0C0(uint32(index)))
		},
		protectItem: func(value uint32, got *Object) {
			events = append(events, "protect:"+inventoryDetachUint4ED0C0(value))
		},
		npcSetItemEquip: func(*Object, *Object, int32) {
			t.Fatal("Player path called NPC equip")
		},
		clearOwner: func(got *Object) {
			events = append(events, "clear")
			if got != item {
				t.Fatalf("clear item = %p, want %p", got, item)
			}
			owner.InvFirstItem = remaining
			owner.UpdateData = unsafe.Pointer(postUpdate)
		},
	})

	wantEvents := []string{
		"flag:4096", "flag:32", "dequip:255", "armor:0:1", "weapon:0:1",
		"drop:3", "protect:91", "clear",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if initialPlayer.WeaponEquip != 6 {
		t.Fatalf("initial WeaponEquip = %#x, want 6", initialPlayer.WeaponEquip)
	}
	if previous.InvNextItem != next || next.Field125 != previous {
		t.Fatalf("native unlink = next %p previous %p, want %p/%p", previous.InvNextItem, next.Field125, next, previous)
	}
	if item.InvHolder != nil || item.InvNextItem != next || item.Field125 != previous {
		t.Fatalf("detached item fields = holder %p next %p previous %p", item.InvHolder, item.InvNextItem, item.Field125)
	}
	if postPlayer.Field3656 != 1 {
		t.Fatalf("post-clear overweight = %d, want 1", postPlayer.Field3656)
	}
}
