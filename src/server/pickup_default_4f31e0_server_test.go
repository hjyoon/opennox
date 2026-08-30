package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func defaultPickupDefaultNativeDeps4F31E0() pickupDefaultNativeDeps4F31E0 {
	return pickupDefaultNativeDeps4F31E0{
		gameFlagsCheck:    func(uint32) int32 { return 0 },
		findTeam:          func(uint8) *Team { return nil },
		informTeam:        func(uint8, uint8, uint32) {},
		primaryMessage:    func(*Object, string, uint8) {},
		deleteWorldObject: func(*Object) {},
		inventoryPut:      func(*Object, *Object, int32) {},
	}
}

func TestPickupDefault4F31E0NativeLayout(t *testing.T) {
	checks32 := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.size", unsafe.Sizeof(Object{}), 780},
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), 4},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), 8},
		{"Object.TeamVal", unsafe.Offsetof(Object{}.TeamVal), 48},
		{"Object.Weight", unsafe.Offsetof(Object{}.Weight), 488},
		{"Object.CarryCapacity", unsafe.Offsetof(Object{}.CarryCapacity), 490},
		{"Object.InvHolder", unsafe.Offsetof(Object{}.InvHolder), 492},
		{"Object.InvNextItem", unsafe.Offsetof(Object{}.InvNextItem), 496},
		{"Object.InvFirstItem", unsafe.Offsetof(Object{}.InvFirstItem), 504},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), 748},
		{"ObjectTeam.size", unsafe.Sizeof(ObjectTeam{}), 8},
		{"ObjectTeam.ID", unsafe.Offsetof(ObjectTeam{}.ID), 4},
		{"Team.size", unsafe.Sizeof(Team{}), 80},
		{"Team.ColorInd", unsafe.Offsetof(Team{}.ColorInd), 56},
		{"PlayerUpdateData.size", unsafe.Sizeof(PlayerUpdateData{}), 556},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), 276},
		{"Player.size", unsafe.Sizeof(Player{}), 4828},
		{"Player.PlayerInd", unsafe.Offsetof(Player{}.PlayerInd), 2064},
	}
	checks64 := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.size", unsafe.Sizeof(Object{}), 928},
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), 8},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), 12},
		{"Object.TeamVal", unsafe.Offsetof(Object{}.TeamVal), 52},
		{"Object.Weight", unsafe.Offsetof(Object{}.Weight), 516},
		{"Object.CarryCapacity", unsafe.Offsetof(Object{}.CarryCapacity), 518},
		{"Object.InvHolder", unsafe.Offsetof(Object{}.InvHolder), 520},
		{"Object.InvNextItem", unsafe.Offsetof(Object{}.InvNextItem), 528},
		{"Object.InvFirstItem", unsafe.Offsetof(Object{}.InvFirstItem), 544},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), 872},
		{"ObjectTeam.size", unsafe.Sizeof(ObjectTeam{}), 8},
		{"ObjectTeam.ID", unsafe.Offsetof(ObjectTeam{}.ID), 4},
		{"Team.size", unsafe.Sizeof(Team{}), 88},
		{"Team.ColorInd", unsafe.Offsetof(Team{}.ColorInd), 56},
		{"PlayerUpdateData.size", unsafe.Sizeof(PlayerUpdateData{}), 656},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), 336},
		{"Player.size", unsafe.Sizeof(Player{}), 6160},
		{"Player.PlayerInd", unsafe.Offsetof(Player{}.PlayerInd), 2068},
	}
	checks := checks64
	if unsafe.Sizeof(uintptr(0)) == 4 {
		checks = checks32
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestPickupDefaultNative4F31E0PreservesPointersAndFields(t *testing.T) {
	player := &Player{PlayerInd: 0xe1}
	update := &PlayerUpdateData{Player: player}
	owner := &Object{
		ObjClass:      object.ClassPlayer | object.Class(0x80000000),
		TeamVal:       ObjectTeam{ID: 3},
		CarryCapacity: 50,
		UpdateData:    unsafe.Pointer(update),
	}
	first := &Object{Weight: 11}
	second := &Object{Weight: 17}
	first.InvNextItem = second
	owner.InvFirstItem = first
	item := &Object{
		TypeInd:  0xf123,
		ObjClass: object.ClassFood | object.Class(0x80000000),
		ObjFlags: object.FlagActive,
		TeamVal:  ObjectTeam{ID: 9},
		Weight:   20,
	}
	team := &Team{ColorInd: TeamOrange}

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for index, pointer := range []unsafe.Pointer{
			unsafe.Pointer(owner), unsafe.Pointer(item), unsafe.Pointer(first),
			unsafe.Pointer(second), unsafe.Pointer(team), unsafe.Pointer(update),
			unsafe.Pointer(player),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("pointer %d = %p, want native high address", index, pointer)
			}
		}
	}

	events := make([]string, 0, 8)
	deps := defaultPickupDefaultNativeDeps4F31E0()
	deps.gameFlagsCheck = func(mask uint32) int32 {
		events = append(events, "flags")
		if mask == pickupDefaultQuestFlag4F31E0 {
			return 1
		}
		if mask != pickupDefaultQuestCoopFlags4F31E0 {
			t.Fatalf("mask = %#x", mask)
		}
		return 0
	}
	deps.findTeam = func(uint8) *Team {
		t.Fatal("Quest path resolved a team")
		return nil
	}
	deps.primaryMessage = func(*Object, string, uint8) {
		t.Fatal("successful pickup emitted a primary message")
	}
	deps.deleteWorldObject = func(got *Object) {
		if got != item {
			t.Fatalf("deleted = %p, want %p", got, item)
		}
		events = append(events, "delete")
		item.ObjFlags &^= object.FlagActive
	}
	deps.inventoryPut = func(gotOwner, gotItem *Object, report int32) {
		if gotOwner != owner || gotItem != item || report != -17 {
			t.Fatalf("put = (%p,%p,%d), want (%p,%p,-17)", gotOwner, gotItem, report, owner, item)
		}
		if item.Flags().Has(object.FlagActive) {
			t.Fatal("inventory insertion ran before world deletion")
		}
		events = append(events, "put")
	}

	if got := pickupDefaultNative4F31E0(owner, item, -17, math.MaxInt32, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if want := []string{"flags", "flags", "delete", "put"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestPickupDefaultNative4F31E0TeamInformUsesLiveNativeChain(t *testing.T) {
	initialPlayer := &Player{PlayerInd: 4}
	replacementPlayer := &Player{PlayerInd: 7}
	update := &PlayerUpdateData{Player: initialPlayer}
	owner := &Object{
		ObjClass:   object.ClassPlayer,
		TeamVal:    ObjectTeam{ID: 2},
		UpdateData: unsafe.Pointer(update),
	}
	item := &Object{TeamVal: ObjectTeam{ID: 9}}
	team := &Team{ColorInd: TeamViolet}
	deps := defaultPickupDefaultNativeDeps4F31E0()
	deps.findTeam = func(id uint8) *Team {
		if id != 9 {
			t.Fatalf("team id = %d, want 9", id)
		}
		update.Player = replacementPlayer
		return team
	}
	deps.informTeam = func(index, code uint8, color uint32) {
		if index != replacementPlayer.PlayerInd || code != 16 || color != uint32(TeamViolet) {
			t.Fatalf("inform = (%d,%d,%d)", index, code, color)
		}
		team.ColorInd = TeamOrange
	}

	if got := pickupDefaultNative4F31E0(owner, item, 1, 2, deps); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if team.ColorInd != TeamOrange {
		t.Fatal("team inform callback was not reached")
	}
}
