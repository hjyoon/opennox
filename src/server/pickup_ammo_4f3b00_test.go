package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type pickupAmmoTestModifier4F3B00 struct{ name string }

type pickupAmmoTestInit4F3B00 struct {
	name      string
	modifiers [pickupAmmoModifierCount4F3B00]*pickupAmmoTestModifier4F3B00
}

type pickupAmmoTestUse4F3B00 struct {
	name   string
	values [3]uint8
}

type pickupAmmoTestPlayer4F3B00 struct {
	name string
	ind  uint8
}

type pickupAmmoTestUpdate4F3B00 struct {
	name   string
	player *pickupAmmoTestPlayer4F3B00
}

type pickupAmmoTestObject4F3B00 struct {
	name      string
	equip     uint32
	classLow  uint8
	typeInd   uint16
	class     uint32
	init      *pickupAmmoTestInit4F3B00
	use       *pickupAmmoTestUse4F3B00
	update    *pickupAmmoTestUpdate4F3B00
	firstItem *pickupAmmoTestObject4F3B00
	nextItem  *pickupAmmoTestObject4F3B00
}

type pickupAmmoTestWorld4F3B00 struct {
	arg3, arg4    int32
	defaultResult int32
	events        []string
	faultAt       int
	afterFlags    func(*pickupAmmoTestObject4F3B00)
	afterInit     func(*pickupAmmoTestObject4F3B00)
	afterUse      func(*pickupAmmoTestObject4F3B00)
	afterReport   func()
	afterDelete   func()
}

func pickupAmmoTestObjectName4F3B00(obj *pickupAmmoTestObject4F3B00) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func pickupAmmoTestInitName4F3B00(init *pickupAmmoTestInit4F3B00) string {
	if init == nil {
		return "nil"
	}
	return init.name
}

func pickupAmmoTestUseName4F3B00(use *pickupAmmoTestUse4F3B00) string {
	if use == nil {
		return "nil"
	}
	return use.name
}

func pickupAmmoTestModifierName4F3B00(modifier *pickupAmmoTestModifier4F3B00) string {
	if modifier == nil {
		return "nil"
	}
	return modifier.name
}

func pickupAmmoTestUpdateName4F3B00(update *pickupAmmoTestUpdate4F3B00) string {
	if update == nil {
		return "nil"
	}
	return update.name
}

func pickupAmmoTestPlayerName4F3B00(player *pickupAmmoTestPlayer4F3B00) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func (w *pickupAmmoTestWorld4F3B00) event(format string, args ...any) {
	value := fmt.Sprintf(format, args...)
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(value)
	}
}

func (w *pickupAmmoTestWorld4F3B00) hooks() pickupAmmoHooks4F3B00[
	*pickupAmmoTestObject4F3B00,
	*pickupAmmoTestInit4F3B00,
	*pickupAmmoTestUse4F3B00,
	*pickupAmmoTestModifier4F3B00,
	*pickupAmmoTestUpdate4F3B00,
	*pickupAmmoTestPlayer4F3B00,
] {
	return pickupAmmoHooks4F3B00[
		*pickupAmmoTestObject4F3B00,
		*pickupAmmoTestInit4F3B00,
		*pickupAmmoTestUse4F3B00,
		*pickupAmmoTestModifier4F3B00,
		*pickupAmmoTestUpdate4F3B00,
		*pickupAmmoTestPlayer4F3B00,
	]{
		weaponEquipFlags: func(obj *pickupAmmoTestObject4F3B00) uint32 {
			var value uint32
			if obj != nil {
				value = obj.equip
			}
			w.event("flags:%s=%08x", pickupAmmoTestObjectName4F3B00(obj), value)
			if w.afterFlags != nil {
				w.afterFlags(obj)
			}
			return value
		},
		loadOwnerClassLow: func(owner *pickupAmmoTestObject4F3B00) uint8 {
			value := owner.classLow
			w.event("owner-class:%s=%02x", pickupAmmoTestObjectName4F3B00(owner), value)
			return value
		},
		loadOwnerUpdate: func(owner *pickupAmmoTestObject4F3B00) *pickupAmmoTestUpdate4F3B00 {
			update := owner.update
			w.event("update:%s=%s", pickupAmmoTestObjectName4F3B00(owner), pickupAmmoTestUpdateName4F3B00(update))
			return update
		},
		loadInventoryHead: func(owner *pickupAmmoTestObject4F3B00) *pickupAmmoTestObject4F3B00 {
			item := owner.firstItem
			w.event("head:%s=%s", pickupAmmoTestObjectName4F3B00(owner), pickupAmmoTestObjectName4F3B00(item))
			return item
		},
		loadTypeInd: func(obj *pickupAmmoTestObject4F3B00) uint16 {
			value := obj.typeInd
			w.event("type:%s=%04x", pickupAmmoTestObjectName4F3B00(obj), value)
			return value
		},
		loadObjectClass: func(obj *pickupAmmoTestObject4F3B00) uint32 {
			value := obj.class
			w.event("class:%s=%08x", pickupAmmoTestObjectName4F3B00(obj), value)
			return value
		},
		loadInitData: func(obj *pickupAmmoTestObject4F3B00) *pickupAmmoTestInit4F3B00 {
			init := obj.init
			w.event("init:%s=%s", pickupAmmoTestObjectName4F3B00(obj), pickupAmmoTestInitName4F3B00(init))
			if w.afterInit != nil {
				w.afterInit(obj)
			}
			return init
		},
		loadUseData: func(obj *pickupAmmoTestObject4F3B00) *pickupAmmoTestUse4F3B00 {
			use := obj.use
			w.event("use:%s=%s", pickupAmmoTestObjectName4F3B00(obj), pickupAmmoTestUseName4F3B00(use))
			if w.afterUse != nil {
				w.afterUse(obj)
			}
			return use
		},
		loadModifier: func(init *pickupAmmoTestInit4F3B00, index int) *pickupAmmoTestModifier4F3B00 {
			modifier := init.modifiers[index]
			w.event("modifier:%s:%d=%s", pickupAmmoTestInitName4F3B00(init), index, pickupAmmoTestModifierName4F3B00(modifier))
			return modifier
		},
		loadUseByte: func(use *pickupAmmoTestUse4F3B00, index int) uint8 {
			value := use.values[index]
			w.event("use-byte:%s:%d=%02x", pickupAmmoTestUseName4F3B00(use), index, value)
			return value
		},
		storeUseByte: func(use *pickupAmmoTestUse4F3B00, index int, value uint8) {
			w.event("store-byte:%s:%d=%02x", pickupAmmoTestUseName4F3B00(use), index, value)
			use.values[index] = value
		},
		loadInventoryNext: func(obj *pickupAmmoTestObject4F3B00) *pickupAmmoTestObject4F3B00 {
			next := obj.nextItem
			w.event("next:%s=%s", pickupAmmoTestObjectName4F3B00(obj), pickupAmmoTestObjectName4F3B00(next))
			return next
		},
		loadUpdatePlayer: func(update *pickupAmmoTestUpdate4F3B00) *pickupAmmoTestPlayer4F3B00 {
			player := update.player
			w.event("player:%s=%s", pickupAmmoTestUpdateName4F3B00(update), pickupAmmoTestPlayerName4F3B00(player))
			return player
		},
		loadPlayerInd: func(player *pickupAmmoTestPlayer4F3B00) uint8 {
			value := player.ind
			w.event("player-ind:%s=%02x", pickupAmmoTestPlayerName4F3B00(player), value)
			return value
		},
		reportCharges: func(index uint8, item *pickupAmmoTestObject4F3B00, charge1, charge0 uint8) {
			w.event("report:%02x:%s:%02x:%02x", index, pickupAmmoTestObjectName4F3B00(item), charge1, charge0)
			if w.afterReport != nil {
				w.afterReport()
			}
		},
		delayedDelete: func(item *pickupAmmoTestObject4F3B00) {
			w.event("delete:%s", pickupAmmoTestObjectName4F3B00(item))
			if w.afterDelete != nil {
				w.afterDelete()
			}
		},
		pickupAudio: func(owner, item *pickupAmmoTestObject4F3B00) {
			w.event("audio:%s:%s", pickupAmmoTestObjectName4F3B00(owner), pickupAmmoTestObjectName4F3B00(item))
		},
		loadArg4: func() int32 {
			w.event("arg4=%08x", uint32(w.arg4))
			return w.arg4
		},
		loadArg3: func() int32 {
			w.event("arg3=%08x", uint32(w.arg3))
			return w.arg3
		},
		defaultPickup: func(owner, item *pickupAmmoTestObject4F3B00, arg3, arg4 int32) int32 {
			w.event("default:%s:%s:%08x:%08x", pickupAmmoTestObjectName4F3B00(owner), pickupAmmoTestObjectName4F3B00(item), uint32(arg3), uint32(arg4))
			return w.defaultResult
		},
	}
}

func verifyPickupAmmoFaultPrefixes4F3B00(
	t *testing.T,
	want []string,
	build func() (*pickupAmmoTestWorld4F3B00, *pickupAmmoTestObject4F3B00, *pickupAmmoTestObject4F3B00),
) {
	t.Helper()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w, owner, item := build()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %v, want %v", w.events, want[:faultAt])
				}
			}()
			pickupAmmo4F3B00(owner, item, w.hooks())
		})
	}
}

func pickupAmmoSuccessBuild4F3B00() (*pickupAmmoTestWorld4F3B00, *pickupAmmoTestObject4F3B00, *pickupAmmoTestObject4F3B00) {
	modifiers := [pickupAmmoModifierCount4F3B00]*pickupAmmoTestModifier4F3B00{
		{name: "A"}, {name: "B"}, nil, {name: "D"},
	}
	itemInit := &pickupAmmoTestInit4F3B00{name: "item-init", modifiers: modifiers}
	candidateInit := &pickupAmmoTestInit4F3B00{name: "candidate-init", modifiers: modifiers}
	itemUse := &pickupAmmoTestUse4F3B00{name: "item-use", values: [3]uint8{50, 250, 7}}
	candidateUse := &pickupAmmoTestUse4F3B00{name: "candidate-use", values: [3]uint8{200, 10, 0}}
	player := &pickupAmmoTestPlayer4F3B00{name: "cached-player", ind: 0xfe}
	update := &pickupAmmoTestUpdate4F3B00{name: "cached-update", player: player}
	replacementUpdate := &pickupAmmoTestUpdate4F3B00{
		name:   "replacement-update",
		player: &pickupAmmoTestPlayer4F3B00{name: "replacement-player", ind: 1},
	}
	item := &pickupAmmoTestObject4F3B00{
		name: "item", equip: 0x82, typeInd: 0x8001, init: itemInit, use: itemUse,
	}
	candidate := &pickupAmmoTestObject4F3B00{
		name: "candidate", equip: 0x80, typeInd: 0x8001,
		class: pickupAmmoWeaponClass4F3B00, init: candidateInit, use: candidateUse,
	}
	owner := &pickupAmmoTestObject4F3B00{
		name: "owner", classLow: 0x84, update: update, firstItem: candidate,
	}
	w := &pickupAmmoTestWorld4F3B00{arg3: math.MinInt32, arg4: math.MaxInt32}
	w.afterFlags = func(obj *pickupAmmoTestObject4F3B00) {
		switch obj {
		case item:
			item.equip = 0
		case candidate:
			owner.update = replacementUpdate
		}
	}
	w.afterInit = func(obj *pickupAmmoTestObject4F3B00) {
		if obj == item {
			item.init = &pickupAmmoTestInit4F3B00{name: "replacement-item-init"}
		}
	}
	w.afterUse = func(obj *pickupAmmoTestObject4F3B00) {
		switch obj {
		case item:
			item.use = &pickupAmmoTestUse4F3B00{name: "replacement-item-use"}
		case candidate:
			candidate.use = &pickupAmmoTestUse4F3B00{name: "replacement-candidate-use"}
		}
	}
	return w, owner, item
}

func pickupAmmoSuccessTrace4F3B00() []string {
	return []string{
		"flags:item=00000082",
		"owner-class:owner=84",
		"update:owner=cached-update",
		"head:owner=candidate",
		"init:item=item-init",
		"use:item=item-use",
		"type:candidate=8001",
		"type:item=8001",
		"class:candidate=01000000",
		"flags:candidate=00000080",
		"init:candidate=candidate-init",
		"use:candidate=candidate-use",
		"modifier:candidate-init:0=A",
		"modifier:item-init:0=A",
		"modifier:candidate-init:1=B",
		"modifier:item-init:1=B",
		"modifier:candidate-init:2=nil",
		"modifier:item-init:2=nil",
		"modifier:candidate-init:3=D",
		"modifier:item-init:3=D",
		"use-byte:candidate-use:2=00",
		"use-byte:candidate-use:0=c8",
		"use-byte:item-use:0=32",
		"use-byte:item-use:1=fa",
		"use-byte:candidate-use:1=0a",
		"store-byte:candidate-use:1=04",
		"use-byte:candidate-use:0=c8",
		"use-byte:item-use:0=32",
		"store-byte:candidate-use:0=fa",
		"player:cached-update=cached-player",
		"player-ind:cached-player=fe",
		"report:fe:candidate:04:fa",
		"delete:item",
		"audio:owner:item",
	}
}

func TestPickupAmmo4F3B00MergesCachedCompatibleAmmoAndReports(t *testing.T) {
	w, owner, item := pickupAmmoSuccessBuild4F3B00()
	if got := pickupAmmo4F3B00(owner, item, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := pickupAmmoSuccessTrace4F3B00()
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events =\n%v\nwant =\n%v", w.events, want)
	}
	candidateUse := owner.firstItem.use
	if candidateUse.name != "replacement-candidate-use" {
		t.Fatalf("live candidate UseData = %q, want replacement", candidateUse.name)
	}
	verifyPickupAmmoFaultPrefixes4F3B00(t, want, pickupAmmoSuccessBuild4F3B00)
}

func TestPickupAmmo4F3B00NonPlayerForwardsExactFourArgumentResult(t *testing.T) {
	build := func() (*pickupAmmoTestWorld4F3B00, *pickupAmmoTestObject4F3B00, *pickupAmmoTestObject4F3B00) {
		owner := &pickupAmmoTestObject4F3B00{name: "owner", classLow: 0x80}
		item := &pickupAmmoTestObject4F3B00{name: "item", equip: 0x82}
		w := &pickupAmmoTestWorld4F3B00{arg3: math.MinInt32, arg4: math.MaxInt32, defaultResult: math.MinInt32}
		return w, owner, item
	}
	w, owner, item := build()
	if got := pickupAmmo4F3B00(owner, item, w.hooks()); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	want := []string{
		"flags:item=00000082",
		"owner-class:owner=80",
		"arg4=7fffffff",
		"arg3=80000000",
		"default:owner:item:80000000:7fffffff",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	verifyPickupAmmoFaultPrefixes4F3B00(t, want, build)
}

func TestPickupAmmo4F3B00PlayerFallbackLoadOrder(t *testing.T) {
	t.Run("non-ammo caches update before fallback", func(t *testing.T) {
		owner := &pickupAmmoTestObject4F3B00{
			name: "owner", classLow: 4,
			update: &pickupAmmoTestUpdate4F3B00{name: "update"},
		}
		item := &pickupAmmoTestObject4F3B00{name: "item", equip: 0x40}
		w := &pickupAmmoTestWorld4F3B00{arg3: -3, arg4: -4, defaultResult: math.MaxInt32}
		if got := pickupAmmo4F3B00(owner, item, w.hooks()); got != math.MaxInt32 {
			t.Fatalf("result = %d, want %d", got, int32(math.MaxInt32))
		}
		want := []string{
			"flags:item=00000040", "owner-class:owner=04", "update:owner=update",
			"arg4=fffffffc", "arg3=fffffffd", "default:owner:item:fffffffd:fffffffc",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	})

	t.Run("empty inventory still caches item data", func(t *testing.T) {
		owner := &pickupAmmoTestObject4F3B00{
			name: "owner", classLow: 4,
			update: &pickupAmmoTestUpdate4F3B00{name: "update"},
		}
		item := &pickupAmmoTestObject4F3B00{
			name: "item", equip: 0x82,
			init: &pickupAmmoTestInit4F3B00{name: "item-init"},
			use:  &pickupAmmoTestUse4F3B00{name: "item-use"},
		}
		w := &pickupAmmoTestWorld4F3B00{arg3: 3, arg4: 4, defaultResult: -17}
		if got := pickupAmmo4F3B00(owner, item, w.hooks()); got != -17 {
			t.Fatalf("result = %d, want -17", got)
		}
		want := []string{
			"flags:item=00000082", "owner-class:owner=04", "update:owner=update",
			"head:owner=nil", "init:item=item-init", "use:item=item-use",
			"arg4=00000004", "arg3=00000003", "default:owner:item:00000003:00000004",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	})
}

func TestPickupAmmo4F3B00ComparesAllModifierSlotsBeforeFallback(t *testing.T) {
	shared := &pickupAmmoTestModifier4F3B00{name: "shared"}
	itemInit := &pickupAmmoTestInit4F3B00{
		name: "item-init",
		modifiers: [4]*pickupAmmoTestModifier4F3B00{
			{name: "item-only"}, shared, nil, shared,
		},
	}
	candidateInit := &pickupAmmoTestInit4F3B00{
		name: "candidate-init",
		modifiers: [4]*pickupAmmoTestModifier4F3B00{
			{name: "candidate-only"}, shared, nil, shared,
		},
	}
	candidate := &pickupAmmoTestObject4F3B00{
		name: "candidate", equip: 0x80, typeInd: 7,
		class: pickupAmmoWeaponClass4F3B00, init: candidateInit,
		use: &pickupAmmoTestUse4F3B00{name: "candidate-use"},
	}
	owner := &pickupAmmoTestObject4F3B00{
		name: "owner", classLow: 4,
		update: &pickupAmmoTestUpdate4F3B00{name: "update"}, firstItem: candidate,
	}
	item := &pickupAmmoTestObject4F3B00{
		name: "item", equip: 0x82, typeInd: 7, init: itemInit,
		use: &pickupAmmoTestUse4F3B00{name: "item-use"},
	}
	w := &pickupAmmoTestWorld4F3B00{arg3: 3, arg4: 4, defaultResult: -99}
	if got := pickupAmmo4F3B00(owner, item, w.hooks()); got != -99 {
		t.Fatalf("result = %d, want -99", got)
	}
	wantTail := []string{
		"modifier:candidate-init:0=candidate-only", "modifier:item-init:0=item-only",
		"modifier:candidate-init:1=shared", "modifier:item-init:1=shared",
		"modifier:candidate-init:2=nil", "modifier:item-init:2=nil",
		"modifier:candidate-init:3=shared", "modifier:item-init:3=shared",
		"next:candidate=nil", "arg4=00000004", "arg3=00000003",
		"default:owner:item:00000003:00000004",
	}
	if len(w.events) < len(wantTail) || !reflect.DeepEqual(w.events[len(w.events)-len(wantTail):], wantTail) {
		t.Fatalf("events tail = %v, want %v", w.events, wantTail)
	}
	for _, event := range w.events {
		if event == "use-byte:candidate-use:2=00" {
			t.Fatal("candidate use byte was read after modifier mismatch")
		}
	}
}

func TestPickupAmmo4F3B00PrimaryCapacityIsInclusiveAndUnsigned(t *testing.T) {
	for _, tc := range []struct {
		name          string
		candidate     uint8
		item          uint8
		wantResult    int32
		wantCandidate uint8
	}{
		{name: "exact 250 merges", candidate: 249, item: 1, wantResult: 1, wantCandidate: 250},
		{name: "251 falls back", candidate: 250, item: 1, wantResult: -7, wantCandidate: 250},
	} {
		t.Run(tc.name, func(t *testing.T) {
			init := &pickupAmmoTestInit4F3B00{name: "shared"}
			candidateUse := &pickupAmmoTestUse4F3B00{name: "candidate", values: [3]uint8{tc.candidate, 0, 0}}
			candidate := &pickupAmmoTestObject4F3B00{
				name: "candidate", equip: 0x80, typeInd: 9,
				class: pickupAmmoWeaponClass4F3B00, init: init, use: candidateUse,
			}
			player := &pickupAmmoTestPlayer4F3B00{name: "player", ind: 3}
			owner := &pickupAmmoTestObject4F3B00{
				name: "owner", classLow: 4, firstItem: candidate,
				update: &pickupAmmoTestUpdate4F3B00{name: "update", player: player},
			}
			item := &pickupAmmoTestObject4F3B00{
				name: "item", equip: 0x82, typeInd: 9, init: init,
				use: &pickupAmmoTestUse4F3B00{name: "item", values: [3]uint8{tc.item, 0, 0}},
			}
			w := &pickupAmmoTestWorld4F3B00{defaultResult: -7}
			if got := pickupAmmo4F3B00(owner, item, w.hooks()); got != tc.wantResult {
				t.Fatalf("result = %d, want %d", got, tc.wantResult)
			}
			if candidateUse.values[0] != tc.wantCandidate {
				t.Fatalf("candidate charge = %d, want %d", candidateUse.values[0], tc.wantCandidate)
			}
		})
	}
}
