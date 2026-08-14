package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestUnusedManaTransfer4EAD50NativeLayouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantClass := uintptr(8)
	wantTeam := uintptr(48)
	wantUpdate := uintptr(748)
	wantTeamSize := uintptr(80)
	wantTeamID := uintptr(57)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantClass = 12
		wantTeam = 52
		wantUpdate = 872
		wantTeamSize = 88
		wantTeamID = 57
	}

	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.TeamVal", unsafe.Offsetof(Object{}.TeamVal), wantTeam},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"ObjectTeam size", unsafe.Sizeof(ObjectTeam{}), 8},
		{"ObjectTeam.ID", unsafe.Offsetof(ObjectTeam{}.ID), 4},
		{"PlayerUpdateData.ManaCur", unsafe.Offsetof(PlayerUpdateData{}.ManaCur), 4},
		{"PlayerUpdateData.ManaMax", unsafe.Offsetof(PlayerUpdateData{}.ManaMax), 8},
		{"ObeliskUpdateData size", unsafe.Sizeof(ObeliskUpdateData{}), 4},
		{"ObeliskUpdateData.Mana", unsafe.Offsetof(ObeliskUpdateData{}.Mana), 0},
		{"Team size", unsafe.Sizeof(Team{}), wantTeamSize},
		{"Team.IDVal", unsafe.Offsetof(Team{}.IDVal), wantTeamID},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestUnusedManaTransferNative4EAD50UntetheredLiveDecrement(t *testing.T) {
	sourceUpdate := &ObeliskUpdateData{Mana: 3}
	playerUpdate := &PlayerUpdateData{ManaCur: 7, ManaMax: 9}
	source := &Object{UpdateData: unsafe.Pointer(sourceUpdate)}
	target := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(playerUpdate)}
	var events []string

	unusedManaTransferNative4EAD50(source, target, unusedManaTransferNativeDeps4EAD50{
		findTeamByID: func(TeamID) *Team {
			t.Fatal("team lookup for unteamed source")
			return nil
		},
		teamContains: func(*ObjectTeam, TeamID) int32 {
			t.Fatal("team containment for unteamed source")
			return 0
		},
		addPlayerMana: func(got *Object, amount int16) uint16 {
			events = append(events, "add")
			if got != target || amount != 1 {
				t.Fatalf("add args = %p/%d", got, amount)
			}
			sourceUpdate.Mana = 11
			return 0xffff
		},
	})

	if !reflect.DeepEqual(events, []string{"add"}) {
		t.Fatalf("events = %#v", events)
	}
	if sourceUpdate.Mana != 10 {
		t.Fatalf("source mana = %d, want live 11 - 1", sourceUpdate.Mana)
	}
	if playerUpdate.ManaCur != 7 || playerUpdate.ManaMax != 9 {
		t.Fatalf("player mana fields changed = %d/%d", playerUpdate.ManaCur, playerUpdate.ManaMax)
	}
}

func TestUnusedManaTransferNative4EAD50TeamPath(t *testing.T) {
	sourceUpdate := &ObeliskUpdateData{Mana: 2}
	playerUpdate := &PlayerUpdateData{ManaCur: 1, ManaMax: 5}
	source := &Object{TeamVal: ObjectTeam{ID: 4}, UpdateData: unsafe.Pointer(sourceUpdate)}
	target := &Object{
		ObjClass:   object.ClassPlayer,
		TeamVal:    ObjectTeam{ID: 7},
		UpdateData: unsafe.Pointer(playerUpdate),
	}
	team := &Team{IDVal: 9}
	var events []string

	unusedManaTransferNative4EAD50(source, target, unusedManaTransferNativeDeps4EAD50{
		findTeamByID: func(id TeamID) *Team {
			events = append(events, "find")
			if id != 7 {
				t.Fatalf("lookup ID = %d, want 7", id)
			}
			return team
		},
		teamContains: func(got *ObjectTeam, id TeamID) int32 {
			events = append(events, "contains")
			if got != &source.TeamVal || id != 9 {
				t.Fatalf("contains args = %p/%d", got, id)
			}
			return -1
		},
		addPlayerMana: func(got *Object, amount int16) uint16 {
			events = append(events, "add")
			if got != target || amount != 1 {
				t.Fatalf("add args = %p/%d", got, amount)
			}
			return 123
		},
	})

	if want := []string{"find", "contains", "add"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
	if sourceUpdate.Mana != 1 {
		t.Fatalf("source mana = %d, want 1", sourceUpdate.Mana)
	}
}

func TestUnusedManaTransfer4EAD50ServerBinding(t *testing.T) {
	sourceUpdate := &ObeliskUpdateData{Mana: 1}
	playerUpdate := &PlayerUpdateData{ManaCur: 2, ManaMax: 3}
	source := &Object{TeamVal: ObjectTeam{ID: 4}, UpdateData: unsafe.Pointer(sourceUpdate)}
	target := &Object{
		ObjClass:   object.ClassPlayer,
		TeamVal:    ObjectTeam{ID: 7},
		UpdateData: unsafe.Pointer(playerUpdate),
	}
	s := &Server{}
	s.Teams.Arr = []Team{{}, {IDVal: 7, ind: 1, active: 1}}
	containsCalls := 0
	addCalls := 0

	s.UnusedManaTransfer4EAD50(source, target, UnusedManaTransferRuntime4EAD50{
		TeamContains: func(got *ObjectTeam, id TeamID) int32 {
			containsCalls++
			if got != &source.TeamVal || id != 7 {
				t.Fatalf("contains args = %p/%d", got, id)
			}
			return 1
		},
		AddPlayerMana: func(got *Object, amount int16) uint16 {
			addCalls++
			if got != target || amount != 1 {
				t.Fatalf("add args = %p/%d", got, amount)
			}
			return 0
		},
	})

	if containsCalls != 1 || addCalls != 1 || sourceUpdate.Mana != 0 {
		t.Fatalf("binding result = contains %d, add %d, source mana %d", containsCalls, addCalls, sourceUpdate.Mana)
	}
}

func TestUnusedManaTransferNative4EAD50FullTargetSkipsNilSourceRecord(t *testing.T) {
	source := &Object{}
	playerUpdate := &PlayerUpdateData{ManaCur: 5, ManaMax: 5}
	target := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(playerUpdate)}
	unusedManaTransferNative4EAD50(source, target, unusedManaTransferNativeDeps4EAD50{
		findTeamByID:  func(TeamID) *Team { t.Fatal("team lookup"); return nil },
		teamContains:  func(*ObjectTeam, TeamID) int32 { t.Fatal("team containment"); return 0 },
		addPlayerMana: func(*Object, int16) uint16 { t.Fatal("mana add"); return 0 },
	})
}

func TestUnusedManaTransferNative4EAD50NilSourceFaultsAfterTargetCurrent(t *testing.T) {
	playerUpdate := &PlayerUpdateData{ManaCur: 1, ManaMax: 5}
	target := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(playerUpdate)}
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		unusedManaTransferNative4EAD50(nil, target, unusedManaTransferNativeDeps4EAD50{})
	}()
	if recovered == nil {
		t.Fatal("nil source did not fault")
	}
}
