package legacy

import (
	"fmt"
	"reflect"
	"testing"
)

type unitPetFixture4E7B00 struct {
	name     string
	subclass uint32
	update   *unitPetUpdateFixture4E7B00
	owner    *unitPetFixture4E7B00
}

type unitPetUpdateFixture4E7B00 struct {
	name   string
	player *unitPetPlayerFixture4E7B00
}

type unitPetPlayerFixture4E7B00 struct {
	name string
	ind  byte
}

func unitPetTestHooks4E7B00(events *[]string) unitPetHooks4E7B00[
	*unitPetFixture4E7B00,
	*unitPetUpdateFixture4E7B00,
	*unitPetPlayerFixture4E7B00,
] {
	return unitPetHooks4E7B00[*unitPetFixture4E7B00, *unitPetUpdateFixture4E7B00, *unitPetPlayerFixture4E7B00]{
		subclass: func(obj *unitPetFixture4E7B00) uint32 {
			*events = append(*events, "subclass:"+obj.name)
			return obj.subclass
		},
		setSubclass: func(obj *unitPetFixture4E7B00, subclass uint32) {
			*events = append(*events, fmt.Sprintf("set-subclass:%s:%08x", obj.name, subclass))
			obj.subclass = subclass
		},
		updateData: func(obj *unitPetFixture4E7B00) *unitPetUpdateFixture4E7B00 {
			if obj == nil {
				*events = append(*events, "update:nil")
				panic("nil owner")
			}
			*events = append(*events, "update:"+obj.name)
			return obj.update
		},
		player: func(update *unitPetUpdateFixture4E7B00) *unitPetPlayerFixture4E7B00 {
			if update == nil {
				*events = append(*events, "player:nil")
				panic("nil update data")
			}
			if update.player == nil {
				*events = append(*events, "player:"+update.name+":nil")
				return nil
			}
			*events = append(*events, "player:"+update.name+":"+update.player.name)
			return update.player
		},
		playerInd: func(player *unitPetPlayerFixture4E7B00) byte {
			if player == nil {
				*events = append(*events, "index:nil")
				panic("nil player")
			}
			*events = append(*events, "index:"+player.name)
			return player.ind
		},
		monitor: func(ind byte, obj *unitPetFixture4E7B00) {
			*events = append(*events, fmt.Sprintf("monitor:%02x:%s", ind, obj.name))
		},
		mark: func(ind byte, obj *unitPetFixture4E7B00, flags uint32) {
			*events = append(*events, fmt.Sprintf("mark:%02x:%s:%d", ind, obj.name, flags))
		},
		setOwner: func(owner, obj *unitPetFixture4E7B00) {
			*events = append(*events, "set-owner:"+owner.name+":"+obj.name)
			obj.owner = owner
		},
		unmonitor: func(ind byte, obj *unitPetFixture4E7B00) {
			*events = append(*events, fmt.Sprintf("unmonitor:%02x:%s", ind, obj.name))
		},
		unmark: func(ind byte, obj *unitPetFixture4E7B00, flags uint32) {
			*events = append(*events, fmt.Sprintf("unmark:%02x:%s:%d", ind, obj.name, flags))
		},
		clearOwner: func(obj *unitPetFixture4E7B00) {
			*events = append(*events, "clear-owner:"+obj.name)
			obj.owner = nil
		},
	}
}

func TestUnitBecomePet4E7B00GuardsBeforeAnyLoad(t *testing.T) {
	owner := &unitPetFixture4E7B00{name: "owner"}
	pet := &unitPetFixture4E7B00{name: "pet"}
	for _, tc := range []struct {
		name       string
		owner, pet *unitPetFixture4E7B00
	}{
		{name: "nil owner", pet: pet},
		{name: "nil pet", owner: owner},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			unitBecomePet4E7B00(tc.owner, tc.pet, unitPetTestHooks4E7B00(&events))
			if len(events) != 0 {
				t.Fatalf("events = %v, want none", events)
			}
		})
	}
}

func TestUnitBecomePet4E7B00CachesUpdateAndReloadsPlayer(t *testing.T) {
	first := &unitPetPlayerFixture4E7B00{name: "first", ind: 0x81}
	second := &unitPetPlayerFixture4E7B00{name: "second", ind: 0xfe}
	replacement := &unitPetPlayerFixture4E7B00{name: "replacement", ind: 0x44}
	cachedUpdate := &unitPetUpdateFixture4E7B00{name: "cached", player: first}
	owner := &unitPetFixture4E7B00{name: "owner", update: cachedUpdate}
	pet := &unitPetFixture4E7B00{name: "pet", subclass: 0xa5a50001}
	var events []string
	hooks := unitPetTestHooks4E7B00(&events)
	originalMonitor := hooks.monitor
	hooks.monitor = func(ind byte, obj *unitPetFixture4E7B00) {
		originalMonitor(ind, obj)
		owner.update = &unitPetUpdateFixture4E7B00{name: "new", player: replacement}
		cachedUpdate.player = second
	}

	unitBecomePet4E7B00(owner, pet, hooks)
	want := []string{
		"subclass:pet", "update:owner", "set-subclass:pet:a5a50081",
		"player:cached:first", "index:first", "monitor:81:pet",
		"player:cached:second", "index:second", "mark:fe:pet:1",
		"set-owner:owner:pet",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if pet.subclass != 0xa5a50081 {
		t.Fatalf("pet subclass = %#08x, want 0xa5a50081", pet.subclass)
	}
	if pet.owner != owner {
		t.Fatalf("pet owner = %p, want %p", pet.owner, owner)
	}
}

func TestUnitBecomePet4E7B00NilUpdateFaultsAfterSubclassWrite(t *testing.T) {
	owner := &unitPetFixture4E7B00{name: "owner"}
	pet := &unitPetFixture4E7B00{name: "pet", subclass: 0xff00aa01}
	var events []string
	defer func() {
		if recover() == nil {
			t.Fatal("nil update data returned without a panic")
		}
		want := []string{
			"subclass:pet", "update:owner", "set-subclass:pet:ff00aa81", "player:nil",
		}
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
		if pet.subclass != 0xff00aa81 {
			t.Fatalf("pet subclass = %#08x, want 0xff00aa81", pet.subclass)
		}
	}()
	unitBecomePet4E7B00(owner, pet, unitPetTestHooks4E7B00(&events))
}

func TestUnitBecomeEnemy4E7B60NilOwnerFaultsBeforePetCheck(t *testing.T) {
	var events []string
	defer func() {
		if recover() == nil {
			t.Fatal("nil owner returned without a panic")
		}
		if want := []string{"update:nil"}; !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}()
	unitBecomeEnemy4E7B60[*unitPetFixture4E7B00](nil, nil, unitPetTestHooks4E7B00(&events))
}

func TestUnitBecomeEnemy4E7B60NilPetReadsOnlyOwnerUpdate(t *testing.T) {
	owner := &unitPetFixture4E7B00{name: "owner"}
	var events []string
	unitBecomeEnemy4E7B60(owner, nil, unitPetTestHooks4E7B00(&events))
	if want := []string{"update:owner"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestUnitBecomeEnemy4E7B60CachesUpdateAndReloadsPlayer(t *testing.T) {
	first := &unitPetPlayerFixture4E7B00{name: "first", ind: 0x80}
	second := &unitPetPlayerFixture4E7B00{name: "second", ind: 0xff}
	replacement := &unitPetPlayerFixture4E7B00{name: "replacement", ind: 0x33}
	cachedUpdate := &unitPetUpdateFixture4E7B00{name: "cached", player: first}
	owner := &unitPetFixture4E7B00{name: "owner", update: cachedUpdate}
	pet := &unitPetFixture4E7B00{name: "pet", subclass: 0x5a5a80ff, owner: owner}
	var events []string
	hooks := unitPetTestHooks4E7B00(&events)
	originalUnmonitor := hooks.unmonitor
	hooks.unmonitor = func(ind byte, obj *unitPetFixture4E7B00) {
		originalUnmonitor(ind, obj)
		owner.update = &unitPetUpdateFixture4E7B00{name: "new", player: replacement}
		cachedUpdate.player = second
	}

	unitBecomeEnemy4E7B60(owner, pet, hooks)
	want := []string{
		"update:owner", "subclass:pet", "set-subclass:pet:5a5a807f",
		"player:cached:first", "index:first", "unmonitor:80:pet",
		"player:cached:second", "index:second", "unmark:ff:pet:1",
		"clear-owner:pet",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if pet.subclass != 0x5a5a807f {
		t.Fatalf("pet subclass = %#08x, want 0x5a5a807f", pet.subclass)
	}
	if pet.owner != nil {
		t.Fatalf("pet owner = %p, want nil", pet.owner)
	}
}

func TestUnitBecomeEnemy4E7B60NilUpdateFaultsAfterSubclassWrite(t *testing.T) {
	owner := &unitPetFixture4E7B00{name: "owner"}
	pet := &unitPetFixture4E7B00{name: "pet", subclass: 0x123480ff}
	var events []string
	defer func() {
		if recover() == nil {
			t.Fatal("nil update data returned without a panic")
		}
		want := []string{
			"update:owner", "subclass:pet", "set-subclass:pet:1234807f", "player:nil",
		}
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
		if pet.subclass != 0x1234807f {
			t.Fatalf("pet subclass = %#08x, want 0x1234807f", pet.subclass)
		}
	}()
	unitBecomeEnemy4E7B60(owner, pet, unitPetTestHooks4E7B00(&events))
}
