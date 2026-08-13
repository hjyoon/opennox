package legacy

import (
	"fmt"
	"reflect"
	"testing"
)

type unitPostCreateNotifyObject4E7F10 struct {
	name             string
	field35, field36 uint32
}

type unitPostCreateNotifyPlayer4E7F10 struct {
	name string
	ind  uint8
	unit *unitPostCreateNotifyObject4E7F10
	next *unitPostCreateNotifyPlayer4E7F10
}

func unitPostCreateNotifyTestHooks4E7F10(
	events *[]string,
	first **unitPostCreateNotifyPlayer4E7F10,
	hostile func(*unitPostCreateNotifyObject4E7F10, *unitPostCreateNotifyObject4E7F10) int32,
) unitPostCreateNotifyHooks4E7F10[*unitPostCreateNotifyObject4E7F10, *unitPostCreateNotifyPlayer4E7F10] {
	return unitPostCreateNotifyHooks4E7F10[*unitPostCreateNotifyObject4E7F10, *unitPostCreateNotifyPlayer4E7F10]{
		storeField35: func(obj *unitPostCreateNotifyObject4E7F10, value uint32) {
			*events = append(*events, fmt.Sprintf("store35:%#x", value))
			obj.field35 = value
		},
		storeField36: func(obj *unitPostCreateNotifyObject4E7F10, value uint32) {
			*events = append(*events, fmt.Sprintf("store36:%#x", value))
			obj.field36 = value
		},
		firstPlayer: func() *unitPostCreateNotifyPlayer4E7F10 {
			*events = append(*events, "first")
			return *first
		},
		loadPlayerInd: func(player *unitPostCreateNotifyPlayer4E7F10) uint8 {
			*events = append(*events, "index:"+player.name)
			return player.ind
		},
		loadPlayerUnit: func(player *unitPostCreateNotifyPlayer4E7F10) *unitPostCreateNotifyObject4E7F10 {
			*events = append(*events, "unit:"+player.name)
			return player.unit
		},
		isHostile: func(unit, obj *unitPostCreateNotifyObject4E7F10) int32 {
			*events = append(*events, "hostile:"+unit.name+":"+obj.name)
			return hostile(unit, obj)
		},
		loadField35: func(obj *unitPostCreateNotifyObject4E7F10) uint32 {
			*events = append(*events, "load35")
			return obj.field35
		},
		loadField36: func(obj *unitPostCreateNotifyObject4E7F10) uint32 {
			*events = append(*events, "load36")
			return obj.field36
		},
		nextPlayer: func(player *unitPostCreateNotifyPlayer4E7F10) *unitPostCreateNotifyPlayer4E7F10 {
			*events = append(*events, "next:"+player.name)
			return player.next
		},
	}
}

func TestUnitPostCreateNotify4E7F10ClearsBeforeEmptyList(t *testing.T) {
	obj := &unitPostCreateNotifyObject4E7F10{name: "created", field35: 0xffffffff, field36: 0xaaaaaaaa}
	var first *unitPostCreateNotifyPlayer4E7F10
	var events []string
	got := unitPostCreateNotify4E7F10(obj, unitPostCreateNotifyTestHooks4E7F10(
		&events, &first, func(*unitPostCreateNotifyObject4E7F10, *unitPostCreateNotifyObject4E7F10) int32 {
			t.Fatal("hostile callback called for an empty list")
			return 0
		},
	))
	if got != nil {
		t.Fatalf("return = %p, want nil", got)
	}
	if obj.field35 != 0 || obj.field36 != 0 {
		t.Fatalf("masks = (%#x, %#x), want zero", obj.field35, obj.field36)
	}
	want := []string{"store35:0x0", "store36:0x0", "first"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestUnitPostCreateNotify4E7F10ReadOrderExactOneAndMaskedShift(t *testing.T) {
	created := &unitPostCreateNotifyObject4E7F10{name: "created"}
	nonExact := &unitPostCreateNotifyObject4E7F10{name: "non-exact"}
	exact := &unitPostCreateNotifyObject4E7F10{name: "exact"}
	third := &unitPostCreateNotifyPlayer4E7F10{name: "third", ind: 34, unit: exact}
	second := &unitPostCreateNotifyPlayer4E7F10{name: "second", ind: 33, unit: nonExact, next: third}
	firstPlayer := &unitPostCreateNotifyPlayer4E7F10{name: "first", ind: 0xff, next: second}
	first := firstPlayer
	var events []string
	hooks := unitPostCreateNotifyTestHooks4E7F10(&events, &first, func(unit, obj *unitPostCreateNotifyObject4E7F10) int32 {
		switch unit {
		case nonExact:
			obj.field35 = 0x100
			obj.field36 = 0x200
			return 2
		case exact:
			obj.field35 = 0x100
			obj.field36 = 0x200
			return 1
		default:
			t.Fatalf("unexpected hostile unit %q", unit.name)
			return 0
		}
	})

	if got := unitPostCreateNotify4E7F10(created, hooks); got != nil {
		t.Fatalf("return = %p, want nil", got)
	}
	// Player index 34 is masked to x86 shift count 2.
	if created.field35 != 0x104 || created.field36 != 0x204 {
		t.Fatalf("masks = (%#x, %#x), want (0x104, 0x204)", created.field35, created.field36)
	}
	want := []string{
		"store35:0x0", "store36:0x0", "first",
		"index:first", "unit:first", "next:first",
		"index:second", "unit:second", "hostile:non-exact:created", "next:second",
		"index:third", "unit:third", "hostile:exact:created",
		"load35", "load36", "store35:0x104", "store36:0x204", "next:third",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestUnitPostCreateNotify4E7F10CachesBothMasksBeforeStoresAndUsesLiveNext(t *testing.T) {
	created := &unitPostCreateNotifyObject4E7F10{name: "created"}
	matched := &unitPostCreateNotifyObject4E7F10{name: "matched"}
	stale := &unitPostCreateNotifyPlayer4E7F10{name: "stale"}
	replacement := &unitPostCreateNotifyPlayer4E7F10{name: "replacement"}
	current := &unitPostCreateNotifyPlayer4E7F10{name: "current", ind: 3, unit: matched, next: stale}
	first := current
	var events []string
	hooks := unitPostCreateNotifyTestHooks4E7F10(&events, &first, func(_ *unitPostCreateNotifyObject4E7F10, obj *unitPostCreateNotifyObject4E7F10) int32 {
		obj.field35 = 0x10
		obj.field36 = 0x20
		current.next = replacement
		return 1
	})
	baseLoad35 := hooks.loadField35
	hooks.loadField35 = func(obj *unitPostCreateNotifyObject4E7F10) uint32 {
		value := baseLoad35(obj)
		obj.field36 = 0x40
		return value
	}
	baseStore35 := hooks.storeField35
	hooks.storeField35 = func(obj *unitPostCreateNotifyObject4E7F10, value uint32) {
		baseStore35(obj, value)
		obj.field36 = 0x80
	}

	unitPostCreateNotify4E7F10(created, hooks)
	// load36 observes load35's mutation (0x40), while store36 uses its cached
	// value and overwrites store35's later 0x80 mutation.
	if created.field35 != 0x18 || created.field36 != 0x48 {
		t.Fatalf("masks = (%#x, %#x), want (0x18, 0x48)", created.field35, created.field36)
	}
	if containsEvent4E7F10(events, "index:stale") || !containsEvent4E7F10(events, "index:replacement") {
		t.Fatalf("successor events = %v, want live replacement only", events)
	}
}

func TestUnitPostCreateNotify4E7F10NilObjectFaultsOnFirstStore(t *testing.T) {
	var first *unitPostCreateNotifyPlayer4E7F10
	var events []string
	defer func() {
		if recover() == nil {
			t.Fatal("nil object returned without a panic")
		}
		if want := []string{"store35:0x0"}; !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}()
	unitPostCreateNotify4E7F10(
		(*unitPostCreateNotifyObject4E7F10)(nil),
		unitPostCreateNotifyTestHooks4E7F10(&events, &first, func(*unitPostCreateNotifyObject4E7F10, *unitPostCreateNotifyObject4E7F10) int32 { return 0 }),
	)
}

func containsEvent4E7F10(events []string, want string) bool {
	for _, event := range events {
		if event == want {
			return true
		}
	}
	return false
}
