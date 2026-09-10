package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestSummonUnitNativeLayout5016C0(t *testing.T) {
	wantDirection1 := uintptr(124)
	wantDirection2 := uintptr(126)
	wantClass := uintptr(8)
	wantSubclass := uintptr(12)
	wantNetCode := uintptr(36)
	wantTeamID := uintptr(52)
	wantUpdateData := uintptr(748)
	wantAIAction := uintptr(1360)
	wantStatus := uintptr(1440)
	wantPlayerUpdatePlayer := uintptr(276)
	wantPlayerIndex := uintptr(2064)
	wantSummonOrder := uintptr(3648)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantDirection1 = 128
		wantDirection2 = 130
		wantClass = 12
		wantSubclass = 16
		wantNetCode = 40
		wantTeamID = 56
		wantUpdateData = 872
		wantAIAction = 2100
		wantStatus = 2180
		wantPlayerUpdatePlayer = 336
		wantPlayerIndex = 2068
		wantSummonOrder = 4944
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.Direction1", unsafe.Offsetof(Object{}.Direction1), wantDirection1},
		{"Object.Direction2", unsafe.Offsetof(Object{}.Direction2), wantDirection2},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjSubClass", unsafe.Offsetof(Object{}.ObjSubClass), wantSubclass},
		{"Object.NetCode", unsafe.Offsetof(Object{}.NetCode), wantNetCode},
		{"Object.TeamVal.ID", unsafe.Offsetof(Object{}.TeamVal) + unsafe.Offsetof(ObjectTeam{}.ID), wantTeamID},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdateData},
		{"MonsterUpdateData.AIAction340", unsafe.Offsetof(MonsterUpdateData{}.AIAction340), wantAIAction},
		{"MonsterUpdateData.StatusFlags", unsafe.Offsetof(MonsterUpdateData{}.StatusFlags), wantStatus},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayerUpdatePlayer},
		{"Player.PlayerInd", unsafe.Offsetof(Player{}.PlayerInd), wantPlayerIndex},
		{"Player.SummonOrderAll", unsafe.Offsetof(Player{}.SummonOrderAll), wantSummonOrder},
		{"position X width", unsafe.Sizeof(types.Pointf{}.X), 4},
		{"position Y width", unsafe.Sizeof(types.Pointf{}.Y), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestSummonUnitNative5016C0PreservesCachedUpdatesAndReloadsPlayer(t *testing.T) {
	position := &types.Pointf{X: 12.5, Y: -7.25}
	staleUpdate := new(MonsterUpdateData)
	entryUpdate := &MonsterUpdateData{StatusFlags: object.MonsterStatus(0x12345601)}
	replacementUpdate := &MonsterUpdateData{AIAction340: 0xaaaaaaaa}
	created := &Object{
		ObjSubClass: object.SubClass(0x40000000),
		NetCode:     0x89abcdef,
		UpdateData:  unsafe.Pointer(staleUpdate),
	}
	orderPlayer := &Player{PlayerInd: 0x10, SummonOrderAll: 0xfedcba98}
	acquirePlayer := &Player{PlayerInd: 0x11}
	minimapPlayer := &Player{PlayerInd: 0x22}
	simplePlayer := &Player{PlayerInd: 0x33}
	ownerUpdate := &PlayerUpdateData{Player: orderPlayer}
	replacementOwnerUpdate := &PlayerUpdateData{Player: &Player{PlayerInd: 0x77}}
	owner := &Object{
		ObjClass:   object.ClassPlayer,
		TeamVal:    ObjectTeam{Field0: 0x12345678, ID: 0x7e},
		UpdateData: unsafe.Pointer(ownerUpdate),
	}

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"position":       unsafe.Pointer(position),
			"created":        unsafe.Pointer(created),
			"owner":          unsafe.Pointer(owner),
			"monster update": unsafe.Pointer(entryUpdate),
			"player update":  unsafe.Pointer(ownerUpdate),
			"player":         unsafe.Pointer(orderPlayer),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want native address above 4 GiB", name, pointer)
			}
		}
	}

	var events []string
	deps := summonUnitNativeDeps5016C0{
		newObject: func(typeID int32) *Object {
			events = append(events, "new")
			if typeID != int32(-0x76543211) {
				t.Fatalf("type ID = %#x", uint32(typeID))
			}
			return created
		},
		createObjectAt: func(gotCreated, gotOwner *Object, gotPosition types.Pointf) {
			events = append(events, "create")
			if gotCreated != created || gotOwner != owner || gotPosition != (types.Pointf{X: 12.5, Y: -7.25}) {
				t.Fatalf("create args = %p/%p/%+v", gotCreated, gotOwner, gotPosition)
			}
			created.UpdateData = unsafe.Pointer(entryUpdate)
			position.X = 99
			position.Y = 100
		},
		runtime: SummonUnitRuntime5016C0{
			OrderUnit: func(gotOwner, gotCreated *Object, order uint32) {
				events = append(events, "order")
				if gotOwner != owner || gotCreated != created || order != 0xfedcba98 {
					t.Fatalf("order args = %p/%p/%#x", gotOwner, gotCreated, order)
				}
				created.UpdateData = unsafe.Pointer(replacementUpdate)
				owner.UpdateData = unsafe.Pointer(replacementOwnerUpdate)
				ownerUpdate.Player = acquirePlayer
			},
			ReportAcquire: func(index uint8, got *Object) {
				events = append(events, "acquire")
				if index != acquirePlayer.PlayerInd || got != created {
					t.Fatalf("acquire args = %#x/%p", index, got)
				}
				ownerUpdate.Player = minimapPlayer
			},
			MarkMinimap: func(index uint8, got *Object, flags uint32) {
				events = append(events, "minimap")
				if index != minimapPlayer.PlayerInd || got != created || flags != summonUnitMinimapFlag5016C0 {
					t.Fatalf("minimap args = %#x/%p/%#x", index, got, flags)
				}
				ownerUpdate.Player = simplePlayer
			},
			SendSimpleObject: func(index uint8, got *Object) {
				events = append(events, "simple")
				if index != simplePlayer.PlayerInd || got != created {
					t.Fatalf("simple args = %#x/%p", index, got)
				}
			},
			CreateTeam: func(id TeamID, team *ObjectTeam, active int32, netCode uint32, flags int32) {
				events = append(events, "team")
				if id != owner.TeamVal.ID || team != &owner.TeamVal || active != 1 || netCode != 0x89abcdef || flags != 0 {
					t.Fatalf("team args = %d/%p/%d/%#x/%d", id, team, active, netCode, flags)
				}
				created.ObjSubClass = object.SubClass(0x20000000)
			},
		},
	}

	got := summonUnitNative5016C0(int32(-0x76543211), position, owner, 0xab, deps)
	if got != created {
		t.Fatalf("created = %p, want %p", got, created)
	}
	if want := []string{"new", "create", "order", "acquire", "minimap", "simple", "team"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
	if created.Direction1 != 0xab || created.Direction2 != 0xab {
		t.Fatalf("directions = %#x/%#x", created.Direction1, created.Direction2)
	}
	if got := uint32(entryUpdate.StatusFlags); got != 0x12345681 {
		t.Fatalf("entry status = %#x, want 0x12345681", got)
	}
	if entryUpdate.AIAction340 != summonUnitInvalidAction5016C0 {
		t.Fatalf("entry action = %#x, want %#x", entryUpdate.AIAction340, summonUnitInvalidAction5016C0)
	}
	if replacementUpdate.AIAction340 != 0xaaaaaaaa {
		t.Fatalf("replacement update action = %#x", replacementUpdate.AIAction340)
	}
	if got := uint32(created.ObjSubClass); got != 0x20000100 {
		t.Fatalf("subclass = %#x, want callback-live migrate state", got)
	}
	runtime.KeepAlive(position)
	runtime.KeepAlive(created)
	runtime.KeepAlive(owner)
}

func TestSummonUnitNative5016C0AllocationFailureDoesNotDereferencePosition(t *testing.T) {
	deps := summonUnitNativeDeps5016C0{
		newObject: func(typeID int32) *Object {
			if typeID != math.MinInt32 {
				t.Fatalf("type ID = %d, want MinInt32", typeID)
			}
			return nil
		},
		createObjectAt: func(*Object, *Object, types.Pointf) {
			t.Fatal("allocation failure created an object")
		},
	}
	if got := summonUnitNative5016C0(math.MinInt32, nil, &Object{}, 0xff, deps); got != nil {
		t.Fatalf("created = %p, want nil", got)
	}
}
