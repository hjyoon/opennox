package server

import (
	"encoding/binary"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/types"
)

func defaultCrownDropNativeDeps4ED5E0() crownDropNativeDeps4ED5E0 {
	return crownDropNativeDeps4ED5E0{
		gameFlag:       func(uint32) int32 { return 0 },
		gameplayFlag:   func(uint32) int32 { return 0 },
		loadFrame:      func() uint32 { return 0 },
		firstPlayer:    func() *Object { return nil },
		nextPlayer:     func(*Object) *Object { return nil },
		defaultDrop:    func(*Object, *Object, *types.Pointf) int32 { return 0 },
		teamContains:   func(*ObjectTeam, TeamID) int32 { return 0 },
		clearOwner:     func(*Object) {},
		buffOff:        func(*Object, EnchantID) {},
		informDrop:     func(uint8, uint32, uint32) {},
		markMinimapAll: func(*Object, uint32) {},
	}
}

func TestCrownDrop4ED5E0NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantNetCode := uintptr(36)
	wantTeam := uintptr(48)
	wantObjectUpdate := uintptr(748)
	wantPlayerFrame := uintptr(264)
	wantCrownUpdateSize := uintptr(12)
	wantCrownPending := uintptr(4)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantNetCode = 40
		wantTeam = 52
		wantObjectUpdate = 872
		wantPlayerFrame = 324
		wantCrownUpdateSize = 24
		wantCrownPending = 8
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.NetCode", unsafe.Offsetof(Object{}.NetCode), wantNetCode},
		{"Object.TeamVal", unsafe.Offsetof(Object{}.TeamVal), wantTeam},
		{"ObjectTeam.ID", unsafe.Offsetof(ObjectTeam{}.ID), 4},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantObjectUpdate},
		{"PlayerUpdateData.Field66", unsafe.Offsetof(PlayerUpdateData{}.Field66), wantPlayerFrame},
		{"CrownUpdateData size", unsafe.Sizeof(CrownUpdateData{}), wantCrownUpdateSize},
		{"CrownUpdateData.PickupTarget", unsafe.Offsetof(CrownUpdateData{}.PickupTarget), wantCrownPending},
		{"Pointf size", unsafe.Sizeof(types.Pointf{}), 8},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestCrownDropNative4ED5E0BindsNativeFieldsAndServices(t *testing.T) {
	owner := &Object{NetCode: 0xf1234567, TeamVal: ObjectTeam{ID: 5}}
	update := &CrownUpdateData{}
	crown := &Object{TeamVal: ObjectTeam{ID: 7}, UpdateData: unsafe.Pointer(update)}
	firstData := &PlayerUpdateData{Field66: 80}
	secondData := &PlayerUpdateData{Field66: 90}
	first := &Object{TeamVal: ObjectTeam{ID: 7}, UpdateData: unsafe.Pointer(firstData)}
	second := &Object{TeamVal: ObjectTeam{ID: 7}, UpdateData: unsafe.Pointer(secondData)}
	point := &types.Pointf{X: 3.5, Y: -8.25}
	events := make([]string, 0, 8)

	deps := defaultCrownDropNativeDeps4ED5E0()
	deps.gameFlag = func(flag uint32) int32 {
		if flag != 16 {
			t.Fatalf("game flag = %d", flag)
		}
		return -1
	}
	deps.gameplayFlag = func(flag uint32) int32 {
		if flag != 4 {
			t.Fatalf("gameplay flag = %d", flag)
		}
		return 1
	}
	deps.loadFrame = func() uint32 { return 100 }
	deps.firstPlayer = func() *Object { return first }
	deps.nextPlayer = func(obj *Object) *Object {
		switch obj {
		case first:
			return second
		case second:
			return nil
		default:
			t.Fatalf("next player = %p", obj)
			return nil
		}
	}
	deps.teamContains = func(team *ObjectTeam, id TeamID) int32 {
		if id != 7 || (team != &first.TeamVal && team != &second.TeamVal) {
			t.Fatalf("team membership = %p/%d", team, id)
		}
		return -1
	}
	deps.defaultDrop = func(gotOwner, gotCrown *Object, gotPoint *types.Pointf) int32 {
		events = append(events, "default")
		if gotOwner != owner || gotCrown != crown || gotPoint != point || update.PickupTarget != first {
			t.Fatalf("default args/target = %p/%p/%p/%p", gotOwner, gotCrown, gotPoint, update.PickupTarget)
		}
		return -1
	}
	deps.clearOwner = func(obj *Object) {
		events = append(events, "clear")
		if obj != crown {
			t.Fatalf("clear = %p, want crown %p", obj, crown)
		}
	}
	deps.buffOff = func(obj *Object, enchant EnchantID) {
		events = append(events, "buff")
		if obj != owner || enchant != ENCHANT_CROWN {
			t.Fatalf("buff = %p/%v", obj, enchant)
		}
	}
	deps.informDrop = func(code uint8, netCode, teamID uint32) {
		events = append(events, "inform")
		if code != 11 || netCode != 0xf1234567 || teamID != 5 {
			t.Fatalf("inform = %d/%#x/%#x", code, netCode, teamID)
		}
	}
	deps.markMinimapAll = func(obj *Object, flags uint32) {
		events = append(events, "minimap")
		if obj != crown || flags != 1 {
			t.Fatalf("minimap = %p/%d", obj, flags)
		}
	}

	if got := crownDropNative4ED5E0(owner, crown, point, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(events, []string{"default", "clear", "buff", "inform", "minimap"}) {
		t.Fatalf("events = %#v", events)
	}
}

func TestCrownDropNative4ED5E0CachesOriginalCrownUpdate(t *testing.T) {
	owner := &Object{}
	oldUpdate := &CrownUpdateData{}
	newUpdate := &CrownUpdateData{}
	crown := &Object{TeamVal: ObjectTeam{ID: 3}, UpdateData: unsafe.Pointer(oldUpdate)}
	playerData := &PlayerUpdateData{Field66: 1}
	player := &Object{TeamVal: ObjectTeam{ID: 3}, UpdateData: unsafe.Pointer(playerData)}
	point := &types.Pointf{}
	deps := defaultCrownDropNativeDeps4ED5E0()
	deps.gameFlag = func(uint32) int32 { return 1 }
	deps.gameplayFlag = func(uint32) int32 { return 1 }
	deps.loadFrame = func() uint32 { return 2 }
	deps.firstPlayer = func() *Object {
		crown.UpdateData = unsafe.Pointer(newUpdate)
		return player
	}
	deps.nextPlayer = func(*Object) *Object { return nil }
	deps.teamContains = func(*ObjectTeam, TeamID) int32 { return 1 }
	deps.defaultDrop = func(*Object, *Object, *types.Pointf) int32 { return 0 }

	if got := crownDropNative4ED5E0(owner, crown, point, deps); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if oldUpdate.PickupTarget != player || newUpdate.PickupTarget != nil {
		t.Fatalf("targets = old:%p new:%p, want %p/nil", oldUpdate.PickupTarget, newUpdate.PickupTarget, player)
	}
}

func TestCrownDropNative4ED5E0DisabledSelectionRejectsNilWithoutDereference(t *testing.T) {
	deps := defaultCrownDropNativeDeps4ED5E0()
	deps.defaultDrop = func(owner, crown *Object, point *types.Pointf) int32 {
		if owner != nil || crown != nil || point != nil {
			t.Fatalf("default args = %p/%p/%p", owner, crown, point)
		}
		return 0
	}
	if got := crownDropNative4ED5E0(nil, nil, nil, deps); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
}

func TestCrownDropInformPacket4ED5E0ExactLayout(t *testing.T) {
	packet := crownDropInformPacket4ED5E0(11, 0xf1234567, 0x000000ab)
	want := [10]byte{
		byte(netmsg.MSG_INFORM), 11,
		0x67, 0x45, 0x23, 0xf1,
		0xab, 0x00, 0x00, 0x00,
	}
	if packet != want {
		t.Fatalf("packet = %#v, want %#v", packet, want)
	}
	if got := binary.LittleEndian.Uint32(packet[2:6]); got != 0xf1234567 {
		t.Fatalf("net code = %#x", got)
	}
	if got := binary.LittleEndian.Uint32(packet[6:10]); got != 0xab {
		t.Fatalf("team ID = %#x", got)
	}
}
