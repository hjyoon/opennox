package server

import (
	"math"
	"reflect"
	"testing"
)

type inventoryDetachTestPlayer4ED0C0 struct {
	name       string
	field4     uint32
	index      uint8
	protect    uint32
	overweight uint32
}

type inventoryDetachTestUpdate4ED0C0 struct {
	name   string
	player *inventoryDetachTestPlayer4ED0C0
}

type inventoryDetachTestObject4ED0C0 struct {
	name     string
	class    uint32
	flags    uint32
	subclass uint8
	update   *inventoryDetachTestUpdate4ED0C0
	weight   uint8
	capacity uint16
	holder   *inventoryDetachTestObject4ED0C0
	next     *inventoryDetachTestObject4ED0C0
	previous *inventoryDetachTestObject4ED0C0
	first    *inventoryDetachTestObject4ED0C0
}

type inventoryDetachTestWorld4ED0C0 struct {
	owner *inventoryDetachTestObject4ED0C0
	item  *inventoryDetachTestObject4ED0C0

	events  []string
	faultAt int

	itemClassLoads []uint32
	gameFlags      map[uint32]uint32

	initialUpdate *inventoryDetachTestUpdate4ED0C0
	dropPlayer    *inventoryDetachTestPlayer4ED0C0
	protectPlayer *inventoryDetachTestPlayer4ED0C0
	postFirst     *inventoryDetachTestObject4ED0C0
	postUpdate    *inventoryDetachTestUpdate4ED0C0

	replacementNext *inventoryDetachTestObject4ED0C0
	replacementPrev *inventoryDetachTestObject4ED0C0
}

func (w *inventoryDetachTestWorld4ED0C0) event(name string) {
	w.events = append(w.events, name)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(name)
	}
}

func inventoryDetachObjectName4ED0C0(obj *inventoryDetachTestObject4ED0C0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func inventoryDetachUpdateName4ED0C0(update *inventoryDetachTestUpdate4ED0C0) string {
	if update == nil {
		return "nil"
	}
	return update.name
}

func inventoryDetachPlayerName4ED0C0(player *inventoryDetachTestPlayer4ED0C0) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func (w *inventoryDetachTestWorld4ED0C0) hooks() inventoryDetachHooks4ED0C0[
	*inventoryDetachTestObject4ED0C0,
	*inventoryDetachTestUpdate4ED0C0,
	*inventoryDetachTestPlayer4ED0C0,
] {
	return inventoryDetachHooks4ED0C0[
		*inventoryDetachTestObject4ED0C0,
		*inventoryDetachTestUpdate4ED0C0,
		*inventoryDetachTestPlayer4ED0C0,
	]{
		loadOwnerArg: func() *inventoryDetachTestObject4ED0C0 {
			w.event("owner-arg")
			return w.owner
		},
		loadItemArg: func() *inventoryDetachTestObject4ED0C0 {
			w.event("item-arg")
			return w.item
		},
		loadObjectClass: func(obj *inventoryDetachTestObject4ED0C0) uint32 {
			w.event("class:" + inventoryDetachObjectName4ED0C0(obj))
			if obj == w.item && len(w.itemClassLoads) != 0 {
				value := w.itemClassLoads[0]
				w.itemClassLoads = w.itemClassLoads[1:]
				return value
			}
			return obj.class
		},
		loadObjectFlags: func(obj *inventoryDetachTestObject4ED0C0) uint32 {
			w.event("flags:" + obj.name)
			return obj.flags
		},
		loadObjectSubclass: func(obj *inventoryDetachTestObject4ED0C0) uint8 {
			w.event("subclass:" + obj.name)
			return obj.subclass
		},
		loadObjectUpdate: func(obj *inventoryDetachTestObject4ED0C0) *inventoryDetachTestUpdate4ED0C0 {
			w.event("update:" + obj.name)
			return obj.update
		},
		gameFlag: func(flag uint32) uint32 {
			w.event("game-flag:" + inventoryDetachUint4ED0C0(flag))
			return w.gameFlags[flag]
		},
		loadUpdatePlayer: func(update *inventoryDetachTestUpdate4ED0C0) *inventoryDetachTestPlayer4ED0C0 {
			w.event("player:" + inventoryDetachUpdateName4ED0C0(update))
			return update.player
		},
		loadPlayerField4: func(player *inventoryDetachTestPlayer4ED0C0) uint32 {
			w.event("player-field4:" + inventoryDetachPlayerName4ED0C0(player))
			return player.field4
		},
		storePlayerField4: func(player *inventoryDetachTestPlayer4ED0C0, value uint32) {
			w.event("store-player-field4:" + player.name + ":" + inventoryDetachUint4ED0C0(value))
			player.field4 = value
		},
		netReportDequip: func(index uint8, item *inventoryDetachTestObject4ED0C0) {
			w.event("net-dequip:" + inventoryDetachUint4ED0C0(uint32(index)) + ":" + item.name)
		},
		dequipArmor: func(owner, item *inventoryDetachTestObject4ED0C0, mode, report int32) {
			w.event("dequip-armor:" + owner.name + ":" + item.name + ":" + inventoryDetachInt4ED0C0(mode) + ":" + inventoryDetachInt4ED0C0(report))
			if w.initialUpdate != nil && w.dropPlayer != nil {
				w.initialUpdate.player = w.dropPlayer
			}
		},
		dequipWeapon: func(owner, item *inventoryDetachTestObject4ED0C0, mode, report int32) {
			w.event("dequip-weapon:" + owner.name + ":" + item.name + ":" + inventoryDetachInt4ED0C0(mode) + ":" + inventoryDetachInt4ED0C0(report))
		},
		loadPlayerIndex: func(player *inventoryDetachTestPlayer4ED0C0) uint8 {
			w.event("player-index:" + inventoryDetachPlayerName4ED0C0(player))
			return player.index
		},
		netReportDrop: func(index uint8, item *inventoryDetachTestObject4ED0C0) {
			w.event("net-drop:" + inventoryDetachUint4ED0C0(uint32(index)) + ":" + item.name)
			if w.initialUpdate != nil && w.protectPlayer != nil {
				w.initialUpdate.player = w.protectPlayer
			}
		},
		loadPlayerProtect: func(player *inventoryDetachTestPlayer4ED0C0) uint32 {
			w.event("player-protect:" + inventoryDetachPlayerName4ED0C0(player))
			return player.protect
		},
		protectItem: func(value uint32, item *inventoryDetachTestObject4ED0C0) {
			w.event("protect:" + inventoryDetachUint4ED0C0(value) + ":" + item.name)
		},
		npcSetItemEquip: func(owner, item *inventoryDetachTestObject4ED0C0, value int32) {
			w.event("npc-equip:" + owner.name + ":" + item.name + ":" + inventoryDetachInt4ED0C0(value))
		},
		loadInventoryPrev: func(item *inventoryDetachTestObject4ED0C0) *inventoryDetachTestObject4ED0C0 {
			w.event("previous:" + item.name)
			return item.previous
		},
		loadInventoryNext: func(item *inventoryDetachTestObject4ED0C0) *inventoryDetachTestObject4ED0C0 {
			w.event("next:" + item.name)
			return item.next
		},
		storeInventoryNext: func(item, next *inventoryDetachTestObject4ED0C0) {
			w.event("store-next:" + item.name + ":" + inventoryDetachObjectName4ED0C0(next))
			item.next = next
			if item != w.item && w.replacementNext != nil {
				w.item.next = w.replacementNext
				w.item.previous = w.replacementPrev
			}
		},
		storeInventoryPrev: func(item, previous *inventoryDetachTestObject4ED0C0) {
			w.event("store-previous:" + item.name + ":" + inventoryDetachObjectName4ED0C0(previous))
			item.previous = previous
		},
		loadInventoryFirst: func(owner *inventoryDetachTestObject4ED0C0) *inventoryDetachTestObject4ED0C0 {
			w.event("first:" + owner.name)
			return owner.first
		},
		storeInventoryFirst: func(owner, first *inventoryDetachTestObject4ED0C0) {
			w.event("store-first:" + owner.name + ":" + inventoryDetachObjectName4ED0C0(first))
			owner.first = first
		},
		storeInventoryHolder: func(item, holder *inventoryDetachTestObject4ED0C0) {
			w.event("store-holder:" + item.name + ":" + inventoryDetachObjectName4ED0C0(holder))
			item.holder = holder
		},
		clearOwner: func(item *inventoryDetachTestObject4ED0C0) {
			w.event("clear-owner:" + item.name)
			if w.postFirst != nil {
				w.owner.first = w.postFirst
			}
			if w.postUpdate != nil {
				w.owner.update = w.postUpdate
			}
		},
		loadItemWeight: func(item *inventoryDetachTestObject4ED0C0) uint8 {
			w.event("weight:" + item.name)
			return item.weight
		},
		loadCarryCapacity: func(owner *inventoryDetachTestObject4ED0C0) uint16 {
			w.event("capacity:" + owner.name)
			return owner.capacity
		},
		storePlayerOverweight: func(player *inventoryDetachTestPlayer4ED0C0, value uint32) {
			w.event("store-overweight:" + inventoryDetachPlayerName4ED0C0(player) + ":" + inventoryDetachUint4ED0C0(value))
			player.overweight = value
		},
	}
}

func inventoryDetachUint4ED0C0(value uint32) string {
	const digits = "0123456789"
	if value == 0 {
		return "0"
	}
	var buf [10]byte
	i := len(buf)
	for value != 0 {
		i--
		buf[i] = digits[value%10]
		value /= 10
	}
	return string(buf[i:])
}

func inventoryDetachInt4ED0C0(value int32) string {
	if value >= 0 {
		return inventoryDetachUint4ED0C0(uint32(value))
	}
	return "-" + inventoryDetachUint4ED0C0(uint32(-int64(value)))
}

func newInventoryDetachFullTrace4ED0C0() (*inventoryDetachTestWorld4ED0C0, []string) {
	equipPlayer := &inventoryDetachTestPlayer4ED0C0{name: "equip-player", field4: 7}
	dropPlayer := &inventoryDetachTestPlayer4ED0C0{name: "drop-player", index: 9}
	protectPlayer := &inventoryDetachTestPlayer4ED0C0{name: "protect-player", protect: 77}
	weightPlayer := &inventoryDetachTestPlayer4ED0C0{name: "weight-player"}
	initialUpdate := &inventoryDetachTestUpdate4ED0C0{name: "initial-update", player: equipPlayer}
	postUpdate := &inventoryDetachTestUpdate4ED0C0{name: "post-update", player: weightPlayer}
	owner := &inventoryDetachTestObject4ED0C0{
		name:     "owner",
		class:    inventoryDetachPlayerClass4ED0C0,
		flags:    inventoryDetachDeadFlag4ED0C0,
		update:   initialUpdate,
		capacity: 400,
	}
	previous := &inventoryDetachTestObject4ED0C0{name: "previous"}
	staleNext := &inventoryDetachTestObject4ED0C0{name: "stale-next"}
	replacementNext := &inventoryDetachTestObject4ED0C0{name: "replacement-next"}
	replacementPrev := &inventoryDetachTestObject4ED0C0{name: "replacement-previous"}
	item := &inventoryDetachTestObject4ED0C0{
		name:     "item",
		holder:   owner,
		next:     staleNext,
		previous: previous,
	}
	previous.next = item
	weight2 := &inventoryDetachTestObject4ED0C0{name: "weight-2", weight: 250}
	weight1 := &inventoryDetachTestObject4ED0C0{name: "weight-1", weight: 250, next: weight2}
	w := &inventoryDetachTestWorld4ED0C0{
		owner:           owner,
		item:            item,
		itemClassLoads:  []uint32{0, inventoryDetachFlagClass4ED0C0},
		gameFlags:       map[uint32]uint32{inventoryDetachOnlineFlag4ED0C0: 5},
		initialUpdate:   initialUpdate,
		dropPlayer:      dropPlayer,
		protectPlayer:   protectPlayer,
		postFirst:       weight1,
		postUpdate:      postUpdate,
		replacementNext: replacementNext,
		replacementPrev: replacementPrev,
	}
	want := []string{
		"owner-arg",
		"item-arg",
		"class:owner",
		"update:owner",
		"game-flag:4096",
		"flags:owner",
		"class:item",
		"class:item",
		"game-flag:32",
		"player:initial-update",
		"player-field4:equip-player",
		"store-player-field4:equip-player:6",
		"net-dequip:255:item",
		"dequip-armor:owner:item:0:1",
		"dequip-weapon:owner:item:0:1",
		"player:initial-update",
		"player-index:drop-player",
		"net-drop:9:item",
		"player:initial-update",
		"player-protect:protect-player",
		"protect:77:item",
		"previous:item",
		"next:item",
		"store-next:previous:stale-next",
		"next:item",
		"previous:item",
		"store-previous:replacement-next:replacement-previous",
		"store-holder:item:nil",
		"clear-owner:item",
		"class:owner",
		"first:owner",
		"update:owner",
		"weight:weight-1",
		"next:weight-1",
		"weight:weight-2",
		"next:weight-2",
		"capacity:owner",
		"player:post-update",
		"store-overweight:weight-player:1",
	}
	return w, want
}

func TestInventoryDetach4ED0C0ExactPlayerTrace(t *testing.T) {
	w, want := newInventoryDetachFullTrace4ED0C0()
	detachInventory4ED0C0(w.hooks())
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events =\n%v\nwant =\n%v", w.events, want)
	}
	if w.initialUpdate.player != w.protectPlayer {
		t.Fatalf("initial update player = %p, want protect player %p", w.initialUpdate.player, w.protectPlayer)
	}
	if w.item.holder != nil {
		t.Fatalf("item holder = %p, want nil", w.item.holder)
	}
	if w.replacementNext.previous != w.replacementPrev {
		t.Fatalf("replacement next previous = %p, want %p", w.replacementNext.previous, w.replacementPrev)
	}
	if w.postUpdate.player.overweight != 1 {
		t.Fatalf("post player overweight = %d, want 1", w.postUpdate.player.overweight)
	}
}

func TestInventoryDetach4ED0C0FaultOrder(t *testing.T) {
	_, wantEvents := newInventoryDetachFullTrace4ED0C0()
	for faultAt := range wantEvents {
		faultAt++
		t.Run(wantEvents[faultAt-1], func(t *testing.T) {
			w, _ := newInventoryDetachFullTrace4ED0C0()
			w.faultAt = faultAt
			defer func() {
				gotPanic := recover()
				if gotPanic != wantEvents[faultAt-1] {
					t.Fatalf("panic = %v, want %q", gotPanic, wantEvents[faultAt-1])
				}
				if want := wantEvents[:faultAt]; !reflect.DeepEqual(w.events, want) {
					t.Fatalf("events = %v, want %v", w.events, want)
				}
			}()
			detachInventory4ED0C0(w.hooks())
		})
	}
}

func TestInventoryDetach4ED0C0NilArgumentOrder(t *testing.T) {
	for _, tc := range []struct {
		name   string
		owner  *inventoryDetachTestObject4ED0C0
		item   *inventoryDetachTestObject4ED0C0
		events []string
	}{
		{name: "owner", events: []string{"owner-arg"}},
		{name: "item", owner: &inventoryDetachTestObject4ED0C0{name: "owner"}, events: []string{"owner-arg", "item-arg"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := &inventoryDetachTestWorld4ED0C0{owner: tc.owner, item: tc.item, gameFlags: map[uint32]uint32{}}
			detachInventory4ED0C0(w.hooks())
			if !reflect.DeepEqual(w.events, tc.events) {
				t.Fatalf("events = %v, want %v", w.events, tc.events)
			}
		})
	}
}

func TestInventoryDetach4ED0C0SuppressedPlayerReport(t *testing.T) {
	player := &inventoryDetachTestPlayer4ED0C0{name: "player", field4: 3}
	update := &inventoryDetachTestUpdate4ED0C0{name: "update", player: player}
	owner := &inventoryDetachTestObject4ED0C0{
		name:   "owner",
		class:  inventoryDetachPlayerClass4ED0C0,
		flags:  inventoryDetachDeadFlag4ED0C0,
		update: update,
	}
	item := &inventoryDetachTestObject4ED0C0{name: "item"}
	w := &inventoryDetachTestWorld4ED0C0{
		owner:          owner,
		item:           item,
		itemClassLoads: []uint32{inventoryDetachReportMask4ED0C0, inventoryDetachFlagClass4ED0C0},
		gameFlags:      map[uint32]uint32{inventoryDetachOnlineFlag4ED0C0: 1},
	}
	detachInventory4ED0C0(w.hooks())
	if player.field4 != 2 {
		t.Fatalf("player field4 = %#x, want 2", player.field4)
	}
	for _, event := range w.events {
		if event == "net-dequip:255:item" {
			t.Fatal("suppressed path reported dequip")
		}
	}
	wantPair := []string{"dequip-armor:owner:item:0:0", "dequip-weapon:owner:item:0:0"}
	var gotPair []string
	for _, event := range w.events {
		if event == wantPair[0] || event == wantPair[1] {
			gotPair = append(gotPair, event)
		}
	}
	if !reflect.DeepEqual(gotPair, wantPair) {
		t.Fatalf("dequip pair = %v, want %v", gotPair, wantPair)
	}
}

func TestInventoryDetach4ED0C0MonsterUsesCachedEntryClass(t *testing.T) {
	owner := &inventoryDetachTestObject4ED0C0{
		name:     "owner",
		class:    inventoryDetachMonsterClass4ED0C0,
		subclass: inventoryDetachNPCEquipSubclass4ED0C0,
	}
	item := &inventoryDetachTestObject4ED0C0{name: "item"}
	w := &inventoryDetachTestWorld4ED0C0{
		owner:          owner,
		item:           item,
		itemClassLoads: []uint32{inventoryDetachFlagClass4ED0C0},
		gameFlags:      map[uint32]uint32{inventoryDetachOnlineFlag4ED0C0: 3},
	}
	hooks := w.hooks()
	originalNPCEquip := hooks.npcSetItemEquip
	hooks.npcSetItemEquip = func(owner, item *inventoryDetachTestObject4ED0C0, value int32) {
		originalNPCEquip(owner, item, value)
		owner.class = 0
	}
	detachInventory4ED0C0(hooks)
	wantPrefix := []string{
		"owner-arg", "item-arg", "class:owner", "subclass:owner", "class:item", "game-flag:32",
		"npc-equip:owner:item:0", "dequip-armor:owner:item:1:1", "dequip-weapon:owner:item:1:1",
	}
	if !reflect.DeepEqual(w.events[:len(wantPrefix)], wantPrefix) {
		t.Fatalf("event prefix = %v, want %v", w.events[:len(wantPrefix)], wantPrefix)
	}
	if got := w.events[len(w.events)-1]; got != "class:owner" {
		t.Fatalf("last event = %q, want live class reload", got)
	}
}

func TestInventoryDetach4ED0C0HeadUnlinkAndSignedWeight(t *testing.T) {
	player := &inventoryDetachTestPlayer4ED0C0{name: "player", overweight: 9}
	update := &inventoryDetachTestUpdate4ED0C0{name: "update", player: player}
	next := &inventoryDetachTestObject4ED0C0{name: "next"}
	item := &inventoryDetachTestObject4ED0C0{name: "item", next: next}
	owner := &inventoryDetachTestObject4ED0C0{
		name:     "owner",
		class:    inventoryDetachPlayerClass4ED0C0,
		update:   update,
		first:    item,
		capacity: math.MaxUint16,
	}
	w := &inventoryDetachTestWorld4ED0C0{owner: owner, item: item, gameFlags: map[uint32]uint32{inventoryDetachQuestFlag4ED0C0: 1}}
	detachInventory4ED0C0(w.hooks())
	if owner.first != next || next.previous != nil {
		t.Fatalf("head unlink = first %p previous %p, want %p/nil", owner.first, next.previous, next)
	}
	if player.overweight != 0 {
		t.Fatalf("overweight = %d, want 0", player.overweight)
	}
	if item.next != next || item.previous != nil {
		t.Fatalf("detached links changed = next %p previous %p", item.next, item.previous)
	}
}

func TestInventoryDetach4ED0C0Int32WeightRules(t *testing.T) {
	if got := addInventoryWeight4ED0C0(math.MaxInt32, 1); got != math.MinInt32 {
		t.Fatalf("wrapped weight = %d, want %d", got, int32(math.MinInt32))
	}
	if got := inventoryOverweight4ED0C0(math.MinInt32, 0); got != 0 {
		t.Fatalf("negative wrapped weight overweight = %d, want 0", got)
	}
	if got := inventoryOverweight4ED0C0(math.MaxInt32, math.MaxUint16); got != 1 {
		t.Fatalf("large signed weight overweight = %d, want 1", got)
	}
}
