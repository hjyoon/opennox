package opennox

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/opennox/opennox/v1/server"
)

type allAbilityCancelTestRecord4FC180 struct {
	name    string
	unit    uint64
	ability server.Ability
	active  uint32
	next    *allAbilityCancelTestRecord4FC180
	prev    *allAbilityCancelTestRecord4FC180
	freed   bool
}

type allAbilityCancelCooldownKey4FC180 struct {
	index   uint8
	ability server.Ability
}

type allAbilityCancelTestWorld4FC180 struct {
	unitArg       uint64
	unitClass     map[uint64]uint8
	updates       map[uint64]string
	players       map[string]string
	playerClasses map[string]uint8
	playerIndices map[string]uint8
	cooldowns     map[allAbilityCancelCooldownKey4FC180]int32
	head          *allAbilityCancelTestRecord4FC180
	allocator     string
	events        []string
	faultAt       int
	after         map[string]func()
}

func allAbilityCancelRecordName4FC180(record *allAbilityCancelTestRecord4FC180) string {
	if record == nil {
		return "nil"
	}
	return record.name
}

func allAbilityCancelOpaqueName4FC180(value string) string {
	if value == "" {
		return "nil"
	}
	return value
}

func (w *allAbilityCancelTestWorld4FC180) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *allAbilityCancelTestWorld4FC180) hooks() allAbilityCancelHooks4FC180[
	uint64,
	string,
	string,
	*allAbilityCancelTestRecord4FC180,
	string,
] {
	return allAbilityCancelHooks4FC180[
		uint64,
		string,
		string,
		*allAbilityCancelTestRecord4FC180,
		string,
	]{
		loadUnitArg: func() uint64 {
			unit := w.unitArg
			w.record(fmt.Sprintf("unit-arg=%016x", unit))
			return unit
		},
		loadUnitClassLow: func(unit uint64) uint8 {
			class := w.unitClass[unit]
			w.record(fmt.Sprintf("unit-class:%016x=%02x", unit, class))
			return class
		},
		loadUpdateData: func(unit uint64) string {
			update := w.updates[unit]
			w.record(fmt.Sprintf("update:%016x=%s", unit, allAbilityCancelOpaqueName4FC180(update)))
			return update
		},
		loadPlayer: func(update string) string {
			if update == "" {
				w.record("player:nil-update")
				panic("nil UpdateData")
			}
			player := w.players[update]
			w.record("player:" + update + "=" + allAbilityCancelOpaqueName4FC180(player))
			return player
		},
		loadPlayerClass: func(player string) uint8 {
			if player == "" {
				w.record("player-class:nil")
				panic("nil Player")
			}
			class := w.playerClasses[player]
			w.record(fmt.Sprintf("player-class:%s=%d", player, class))
			return class
		},
		loadPlayerIndex: func(player string) uint8 {
			if player == "" {
				w.record("player-index:nil")
				panic("nil Player")
			}
			index := w.playerIndices[player]
			w.record(fmt.Sprintf("player-index:%s=%d", player, index))
			return index
		},
		storeCooldown: func(index uint8, ability server.Ability, value int32) {
			w.record(fmt.Sprintf("cooldown-store:%d:%d=%d", index, ability, value))
			w.cooldowns[allAbilityCancelCooldownKey4FC180{index: index, ability: ability}] = value
		},
		resetAbilities: func(unit uint64, ability server.Ability) {
			w.record(fmt.Sprintf("reset:%016x:%d", unit, ability))
		},
		loadExecHead: func() *allAbilityCancelTestRecord4FC180 {
			head := w.head
			w.record("head=" + allAbilityCancelRecordName4FC180(head))
			return head
		},
		storeExecHead: func(record *allAbilityCancelTestRecord4FC180) {
			w.record("head-store=" + allAbilityCancelRecordName4FC180(record))
			w.head = record
		},
		loadExecUnit: func(record *allAbilityCancelTestRecord4FC180) uint64 {
			unit := record.unit
			w.record(fmt.Sprintf("exec-unit:%s=%016x", record.name, unit))
			return unit
		},
		loadExecAbility: func(record *allAbilityCancelTestRecord4FC180) server.Ability {
			ability := record.ability
			w.record(fmt.Sprintf("exec-ability:%s=%d", record.name, ability))
			return ability
		},
		loadExecNext: func(record *allAbilityCancelTestRecord4FC180) *allAbilityCancelTestRecord4FC180 {
			next := record.next
			w.record("exec-next:" + record.name + "=" + allAbilityCancelRecordName4FC180(next))
			return next
		},
		loadExecPrev: func(record *allAbilityCancelTestRecord4FC180) *allAbilityCancelTestRecord4FC180 {
			prev := record.prev
			w.record("exec-prev:" + record.name + "=" + allAbilityCancelRecordName4FC180(prev))
			return prev
		},
		storeExecNext: func(record, next *allAbilityCancelTestRecord4FC180) {
			w.record("exec-next-store:" + record.name + "=" + allAbilityCancelRecordName4FC180(next))
			record.next = next
		},
		storeExecPrev: func(record, prev *allAbilityCancelTestRecord4FC180) {
			w.record("exec-prev-store:" + record.name + "=" + allAbilityCancelRecordName4FC180(prev))
			record.prev = prev
		},
		reportActive: func(unit uint64, ability server.Ability, active int32) {
			w.record(fmt.Sprintf("active:%016x:%d=%d", unit, ability, active))
		},
		loadExecAllocator: func() string {
			allocator := w.allocator
			w.record("allocator=" + allocator)
			return allocator
		},
		freeExec: func(allocator string, record *allAbilityCancelTestRecord4FC180) {
			w.record("free:" + allocator + ":" + record.name)
			record.freed = true
		},
	}
}

func allAbilityCancelWarriorWorld4FC180(unit uint64) allAbilityCancelTestWorld4FC180 {
	return allAbilityCancelTestWorld4FC180{
		unitArg:       unit,
		unitClass:     map[uint64]uint8{unit: allAbilityCancelPlayerClass4FC180},
		updates:       map[uint64]string{unit: "update"},
		players:       map[string]string{"update": "player"},
		playerClasses: map[string]uint8{"player": allAbilityCancelWarrior4FC180},
		playerIndices: map[string]uint8{"player": 31},
		cooldowns:     make(map[allAbilityCancelCooldownKey4FC180]int32),
		allocator:     "pool",
		after:         make(map[string]func()),
	}
}

func TestAllAbilityCancel4FC180CooldownRangeReloadsAndAllUnitRecords(t *testing.T) {
	const (
		unit     = uint64(0x1234567889abcdef)
		lowAlias = uint64(0x0000000089abcdef)
		other    = uint64(0xfedcba9876543210)
	)
	tail := &allAbilityCancelTestRecord4FC180{name: "tail", unit: other, ability: server.AbilityBerserk, active: 0x44}
	match2 := &allAbilityCancelTestRecord4FC180{name: "match-2", unit: unit, ability: server.Ability(-99), active: 0x33, next: tail}
	match1 := &allAbilityCancelTestRecord4FC180{name: "match-1", unit: unit, ability: server.AbilityWarcry, active: 0x22, next: match2}
	low := &allAbilityCancelTestRecord4FC180{name: "low-alias", unit: lowAlias, ability: server.AbilityWarcry, next: match1}
	match1.prev = low
	match2.prev = match1
	tail.prev = match2
	stale := &allAbilityCancelTestRecord4FC180{name: "stale"}

	w := allAbilityCancelWarriorWorld4FC180(unit)
	w.head = stale
	w.players["update"] = "class-player"
	w.playerClasses["class-player"] = allAbilityCancelWarrior4FC180
	for i := 1; i < int(server.AbilityMax); i++ {
		name := fmt.Sprintf("slot-%d", i)
		w.playerIndices[name] = uint8(i)
	}
	w.after["player-class:class-player=0"] = func() { w.players["update"] = "slot-1" }
	for i := 1; i < int(server.AbilityMax)-1; i++ {
		ability := server.Ability(i)
		next := fmt.Sprintf("slot-%d", i+1)
		key := fmt.Sprintf("cooldown-store:%d:%d=0", i, ability)
		w.after[key] = func() { w.players["update"] = next }
	}
	w.cooldowns[allAbilityCancelCooldownKey4FC180{index: 31, ability: server.AbilityInvalid}] = 77
	w.after[fmt.Sprintf("reset:%016x:%d", unit, server.AbilityMax)] = func() { w.head = low }

	allAbilityCancel4FC180(w.hooks())

	if got := w.cooldowns[allAbilityCancelCooldownKey4FC180{index: 31, ability: server.AbilityInvalid}]; got != 77 {
		t.Fatalf("invalid cooldown slot = %d, want untouched 77", got)
	}
	for i := 1; i < int(server.AbilityMax); i++ {
		key := allAbilityCancelCooldownKey4FC180{index: uint8(i), ability: server.Ability(i)}
		if got := w.cooldowns[key]; got != 0 {
			t.Fatalf("cooldown [%d][%d] = %d, want 0", i, i, got)
		}
	}
	if !match1.freed || !match2.freed || low.freed || tail.freed || stale.freed {
		t.Fatalf("freed low/match1/match2/tail/stale = %v/%v/%v/%v/%v", low.freed, match1.freed, match2.freed, tail.freed, stale.freed)
	}
	if w.head != low || low.next != tail || tail.prev != low {
		t.Fatal("all-unit removal did not preserve the surrounding doubly linked list")
	}
	if match1.active != 0x22 || match2.active != 0x33 {
		t.Fatal("cancellation inspected or changed an Active field")
	}

	want := []string{
		"unit-arg=1234567889abcdef", "unit-class:1234567889abcdef=04",
		"update:1234567889abcdef=update", "player:update=class-player", "player-class:class-player=0",
		"player:update=slot-1", "player-index:slot-1=1", "cooldown-store:1:1=0",
		"player:update=slot-2", "player-index:slot-2=2", "cooldown-store:2:2=0",
		"player:update=slot-3", "player-index:slot-3=3", "cooldown-store:3:3=0",
		"player:update=slot-4", "player-index:slot-4=4", "cooldown-store:4:4=0",
		"player:update=slot-5", "player-index:slot-5=5", "cooldown-store:5:5=0",
		"reset:1234567889abcdef:6", "head=low-alias",
		"exec-unit:low-alias=0000000089abcdef", "exec-next:low-alias=match-1",
		"exec-unit:match-1=1234567889abcdef", "exec-next:match-1=match-2", "exec-ability:match-1=2", "active:1234567889abcdef:2=0",
		"exec-next:match-1=match-2", "exec-prev:match-1=low-alias", "exec-prev-store:match-2=low-alias",
		"exec-prev:match-1=low-alias", "exec-next:match-1=match-2", "exec-next-store:low-alias=match-2", "allocator=pool", "free:pool:match-1",
		"exec-unit:match-2=1234567889abcdef", "exec-next:match-2=tail", "exec-ability:match-2=-99", "active:1234567889abcdef:-99=0",
		"exec-next:match-2=tail", "exec-prev:match-2=low-alias", "exec-prev-store:tail=low-alias",
		"exec-prev:match-2=low-alias", "exec-next:match-2=tail", "exec-next-store:low-alias=tail", "allocator=pool", "free:pool:match-2",
		"exec-unit:tail=fedcba9876543210", "exec-next:tail=nil",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("event order:\n got: %q\nwant: %q", w.events, want)
	}
}

func TestAllAbilityCancel4FC180CachedReportArgsLiveUnlinkAndCachedTraversal(t *testing.T) {
	const unit = uint64(7)
	tail := &allAbilityCancelTestRecord4FC180{name: "tail", unit: 9}
	middle := &allAbilityCancelTestRecord4FC180{name: "middle", unit: unit, ability: server.AbilityWarcry, next: tail}
	head := &allAbilityCancelTestRecord4FC180{name: "head", unit: unit, ability: server.AbilityBerserk, next: middle}
	middle.prev = head
	tail.prev = middle

	liveHeadNext := &allAbilityCancelTestRecord4FC180{name: "live-head-next", prev: head}
	liveMiddlePrev := &allAbilityCancelTestRecord4FC180{name: "live-middle-prev", next: middle}
	liveMiddleNext := &allAbilityCancelTestRecord4FC180{name: "live-middle-next", prev: middle}

	w := allAbilityCancelWarriorWorld4FC180(unit)
	w.head = head
	w.after["exec-ability:head=1"] = func() {
		head.unit = 99
		head.ability = server.Ability(-7)
	}
	w.after["active:0000000000000007:1=0"] = func() { head.next = liveHeadNext }
	w.after["active:0000000000000007:2=0"] = func() {
		middle.next = liveMiddleNext
		middle.prev = liveMiddlePrev
	}

	allAbilityCancel4FC180(w.hooks())

	if !head.freed || !middle.freed || tail.freed || liveHeadNext.freed || liveMiddleNext.freed || liveMiddlePrev.freed {
		t.Fatal("cached traversal or callback-time record selection freed the wrong records")
	}
	if w.head != liveHeadNext || liveHeadNext.prev != nil {
		t.Fatalf("live head unlink = (%s,%s), want (live-head-next,nil)", allAbilityCancelRecordName4FC180(w.head), allAbilityCancelRecordName4FC180(liveHeadNext.prev))
	}
	if liveMiddleNext.prev != liveMiddlePrev || liveMiddlePrev.next != liveMiddleNext {
		t.Fatal("non-head unlink used cached rather than callback-mutated live links")
	}
}

func TestAllAbilityCancel4FC180GatesAndRequiredPointerFaults(t *testing.T) {
	t.Run("nil unit", func(t *testing.T) {
		w := allAbilityCancelWarriorWorld4FC180(0)
		allAbilityCancel4FC180(w.hooks())
		if want := []string{"unit-arg=0000000000000000"}; !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})

	t.Run("non-player", func(t *testing.T) {
		w := allAbilityCancelWarriorWorld4FC180(1)
		w.unitClass[1] = 0x82
		allAbilityCancel4FC180(w.hooks())
		want := []string{"unit-arg=0000000000000001", "unit-class:0000000000000001=82"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})

	t.Run("non-Warrior", func(t *testing.T) {
		w := allAbilityCancelWarriorWorld4FC180(2)
		w.playerClasses["player"] = 2
		allAbilityCancel4FC180(w.hooks())
		want := []string{
			"unit-arg=0000000000000002", "unit-class:0000000000000002=04",
			"update:0000000000000002=update", "player:update=player", "player-class:player=2",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})

	t.Run("nil UpdateData faults", func(t *testing.T) {
		w := allAbilityCancelWarriorWorld4FC180(3)
		w.updates[3] = ""
		defer func() {
			if recover() == nil {
				t.Fatal("nil UpdateData did not fault while loading Player")
			}
			want := []string{
				"unit-arg=0000000000000003", "unit-class:0000000000000003=04",
				"update:0000000000000003=nil", "player:nil-update",
			}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %q, want %q", w.events, want)
			}
		}()
		allAbilityCancel4FC180(w.hooks())
	})

	t.Run("nil Player faults at class", func(t *testing.T) {
		w := allAbilityCancelWarriorWorld4FC180(4)
		w.players["update"] = ""
		defer func() {
			if recover() == nil {
				t.Fatal("nil Player did not fault while loading class")
			}
			want := []string{
				"unit-arg=0000000000000004", "unit-class:0000000000000004=04",
				"update:0000000000000004=update", "player:update=nil", "player-class:nil",
			}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %q, want %q", w.events, want)
			}
		}()
		allAbilityCancel4FC180(w.hooks())
	})

	t.Run("nil Player faults at cooldown index", func(t *testing.T) {
		w := allAbilityCancelWarriorWorld4FC180(5)
		w.after["player-class:player=0"] = func() { w.players["update"] = "" }
		defer func() {
			if recover() == nil {
				t.Fatal("reloaded nil Player did not fault at PlayerInd")
			}
			want := []string{
				"unit-arg=0000000000000005", "unit-class:0000000000000005=04",
				"update:0000000000000005=update", "player:update=player", "player-class:player=0",
				"player:update=nil", "player-index:nil",
			}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %q, want %q", w.events, want)
			}
		}()
		allAbilityCancel4FC180(w.hooks())
	})
}

func TestAllAbilityCancel4FC180FaultPrefixes(t *testing.T) {
	const unit = uint64(7)
	all := []string{
		"unit-arg=0000000000000007", "unit-class:0000000000000007=04",
		"update:0000000000000007=update", "player:update=player", "player-class:player=0",
		"player:update=player", "player-index:player=31", "cooldown-store:31:1=0",
		"player:update=player", "player-index:player=31", "cooldown-store:31:2=0",
		"player:update=player", "player-index:player=31", "cooldown-store:31:3=0",
		"player:update=player", "player-index:player=31", "cooldown-store:31:4=0",
		"player:update=player", "player-index:player=31", "cooldown-store:31:5=0",
		"reset:0000000000000007:6", "head=record",
		"exec-unit:record=0000000000000007", "exec-next:record=nil", "exec-ability:record=-7",
		"active:0000000000000007:-7=0", "exec-next:record=nil", "exec-prev:record=nil", "exec-next:record=nil",
		"head-store=nil", "allocator=pool", "free:pool:record",
	}
	for faultAt := 1; faultAt <= len(all); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			record := &allAbilityCancelTestRecord4FC180{name: "record", unit: unit, ability: server.Ability(-7)}
			w := allAbilityCancelWarriorWorld4FC180(unit)
			w.head = record
			w.faultAt = faultAt
			defer func() {
				if recover() == nil {
					t.Fatal("fault was not propagated")
				}
				if want := all[:faultAt]; !reflect.DeepEqual(w.events, want) {
					t.Fatalf("events = %q, want prefix %q", w.events, want)
				}
			}()
			allAbilityCancel4FC180(w.hooks())
		})
	}
}
