package server

import (
	"encoding/binary"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestDefaultDrop4ED290NativeLayout(t *testing.T) {
	wants := []struct {
		name string
		got  uintptr
		v32  uintptr
		v64  uintptr
	}{
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), 4, 8},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), 8, 12},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), 16, 20},
		{"Object.NetCode", unsafe.Offsetof(Object{}.NetCode), 36, 40},
		{"Object.TeamVal.ID", unsafe.Offsetof(Object{}.TeamVal) + unsafe.Offsetof(ObjectTeam{}.ID), 52, 56},
		{"Object.InvHolder", unsafe.Offsetof(Object{}.InvHolder), 492, 520},
		{"Object.InvNextItem", unsafe.Offsetof(Object{}.InvNextItem), 496, 528},
		{"Object.InvFirstItem", unsafe.Offsetof(Object{}.InvFirstItem), 504, 544},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), 748, 872},
		{"equip update byte +2", unsafe.Offsetof(DefaultDropEquipUpdatePrefix4ED290{}.Field2), 2, 2},
		{"Flag update frame +8", unsafe.Offsetof(FlagUpdateData4EA490{}.State), 8, 8},
		{"Monster AI action", unsafe.Offsetof(MonsterUpdateData{}.AIAction340), 1360, 2036},
		{"Monster status", unsafe.Offsetof(MonsterUpdateData{}.StatusFlags), 1440, 2116},
	}
	for _, tc := range wants {
		want := tc.v64
		if unsafe.Sizeof(uintptr(0)) == 4 {
			want = tc.v32
		}
		if tc.got != want {
			t.Errorf("%s = %d, want %d", tc.name, tc.got, want)
		}
	}
}

func defaultDefaultDropNativeDeps4ED290() defaultDropNativeDeps4ED290 {
	glyph := uint32(70)
	torch := uint32(71)
	lantern := uint32(72)
	return defaultDropNativeDeps4ED290{
		itemIsDroppable: func(*Object) int32 { return 0 },
		itemDropMask:    func(*Object, uint32) int32 { return 0 },
		primaryMessage:  func(*Object, string, uint8) {},
		playAudio:       func(uint32, *Object, int32, uint32) {},
		detachInventory: func(*Object, *Object) {},
		createAt:        func(*Object, *Object, types.Pointf) {},
		weaponFlags:     func(*Object) uint32 { return 0 },
		delayedDelete:   func(*Object) {},
		materialIndex:   func(*Object) uint32 { return 0 },
		informFlagDrop:  func(uint8, uint32, uint32) {},
		markMinimap:     func(*Object, uint32) {},
		loadFrame:       func() uint32 { return 0 },
		teamFlagStatus:  func(uint8, uint8, uint8, uint16) int32 { return 0 },
		loadGlyphCache:  func() uint32 { return glyph },
		storeGlyphCache: func(value uint32) { glyph = value },
		loadTorchCache:  func() uint32 { return torch },
		storeTorchCache: func(value uint32) { torch = value },
		loadLanternCache: func() uint32 {
			return lantern
		},
		storeLanternCache: func(value uint32) {
			lantern = value
		},
		lookupType:   func(string) uint32 { return 0 },
		gameFlag:     func(uint32) uint32 { return 1 },
		loadGameFPS:  func() uint32 { return 30 },
		setDecayTime: func(*Object, uint32) {},
		raise:        func(*Object, float32) {},
		buffOff:      func(*Object, uint32) {},
	}
}

func TestDefaultDropNative4ED290WeaponReplacementUsesNativeLinks(t *testing.T) {
	owner := &Object{}
	item := &Object{ObjClass: object.ClassWeapon, InvHolder: owner, TypeInd: 10}
	prefix := &DefaultDropEquipUpdatePrefix4ED290{Field2: 1}
	candidate := &Object{ObjClass: object.ClassWeapon, UpdateData: unsafe.Pointer(prefix)}
	owner.InvFirstItem = candidate
	point := &types.Pointf{X: 3.5, Y: -9.25}
	deps := defaultDefaultDropNativeDeps4ED290()
	var events []string
	deps.weaponFlags = func(obj *Object) uint32 {
		switch obj {
		case item:
			return 4
		case candidate:
			return 2
		default:
			return 0
		}
	}
	deps.detachInventory = func(gotOwner, gotItem *Object) {
		if gotOwner != owner {
			t.Fatalf("owner = %p, want %p", gotOwner, owner)
		}
		if gotItem == item {
			events = append(events, "drop")
		} else if gotItem == candidate {
			events = append(events, "replace")
		} else {
			t.Fatalf("unexpected item %p", gotItem)
		}
	}
	deps.createAt = func(gotItem, gotOwner *Object, gotPoint types.Pointf) {
		if gotItem != item || gotOwner != nil || gotPoint != *point {
			t.Fatalf("create = (%p, %p, %+v)", gotItem, gotOwner, gotPoint)
		}
		events = append(events, "create")
	}
	deps.delayedDelete = func(got *Object) {
		if got != candidate {
			t.Fatalf("deleted = %p, want %p", got, candidate)
		}
		events = append(events, "delete")
	}

	if got := defaultDropNative4ED290(owner, item, point, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if want := []string{"drop", "create", "replace", "delete"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
}

func TestDefaultDropNative4ED290FlagUsesNativeUpdateAndFullPacketValues(t *testing.T) {
	owner := &Object{NetCode: 0x11223344}
	update := &FlagUpdateData4EA490{}
	item := &Object{
		ObjClass:   object.ClassFlag,
		InvHolder:  owner,
		TeamVal:    ObjectTeam{ID: 9},
		TypeInd:    10,
		UpdateData: unsafe.Pointer(update),
	}
	point := &types.Pointf{}
	deps := defaultDefaultDropNativeDeps4ED290()
	deps.materialIndex = func(got *Object) uint32 {
		if got != item {
			t.Fatalf("material object = %p, want %p", got, item)
		}
		return 0xa1b2c303
	}
	var inform [3]uint32
	deps.informFlagDrop = func(code uint8, netCode, material uint32) {
		inform = [3]uint32{uint32(code), netCode, material}
	}
	marked := false
	deps.markMinimap = func(got *Object, flags uint32) {
		marked = got == item && flags == 1
	}
	deps.loadFrame = func() uint32 { return 0x55667788 }
	var status [4]uint32
	deps.teamFlagStatus = func(team, state, material uint8, carrier uint16) int32 {
		status = [4]uint32{uint32(team), uint32(state), uint32(material), uint32(carrier)}
		return -1
	}

	if got := defaultDropNative4ED290(owner, item, point, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if inform != [3]uint32{7, 0x11223344, 0xa1b2c303} {
		t.Fatalf("inform = %#v", inform)
	}
	if !marked || update.State != 0x55667788 {
		t.Fatalf("marked = %v, frame = %#x", marked, update.State)
	}
	if status != [4]uint32{9, 2, 3, 0} {
		t.Fatalf("status = %#v", status)
	}
}

func TestDefaultDropNative4ED290MonsterUsesNativeUpdateFields(t *testing.T) {
	owner := &Object{}
	update := &MonsterUpdateData{StatusFlags: object.MonsterStatus(0x4000)}
	item := &Object{
		ObjClass:   object.ClassMonster | object.Class(defaultDropAudioClass4ED290),
		InvHolder:  owner,
		TypeInd:    10,
		UpdateData: unsafe.Pointer(update),
	}
	deps := defaultDefaultDropNativeDeps4ED290()
	var audio uint32
	deps.playAudio = func(id uint32, got *Object, kind int32, code uint32) {
		if got != item || kind != 0 || code != 0 {
			t.Fatalf("audio = (%p, %d, %d)", got, kind, code)
		}
		audio = id
	}

	if got := defaultDropNative4ED290(owner, item, &types.Pointf{}, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if audio != 821 {
		t.Fatalf("audio = %d, want 821", audio)
	}
	if update.AIAction340 != 15 || uint32(update.StatusFlags) != 0x4100 {
		t.Fatalf("monster update = action %d, status %#x", update.AIAction340, update.StatusFlags)
	}
}

func TestDefaultDropFlagPacket4ED290ExactLayout(t *testing.T) {
	packet := defaultDropFlagPacket4ED290(7, 0x11223344, 0xa1b2c3d4)
	want := [10]byte{
		byte(netmsg.MSG_INFORM), 7,
		0x44, 0x33, 0x22, 0x11,
		0xd4, 0xc3, 0xb2, 0xa1,
	}
	if packet != want {
		t.Fatalf("packet = % x, want % x", packet, want)
	}
	if binary.LittleEndian.Uint32(packet[2:6]) != 0x11223344 ||
		binary.LittleEndian.Uint32(packet[6:10]) != 0xa1b2c3d4 {
		t.Fatalf("packet words = % x", packet)
	}
}
