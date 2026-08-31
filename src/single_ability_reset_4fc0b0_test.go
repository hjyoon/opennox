package opennox

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/opennox/opennox/v1/server"
)

type singleAbilityResetTestRecord4FC0B0 struct {
	name    string
	unit    uint64
	ability server.Ability
	active  uint32
	next    *singleAbilityResetTestRecord4FC0B0
	prev    *singleAbilityResetTestRecord4FC0B0
	freed   bool
}

type singleAbilityResetCooldownKey4FC0B0 struct {
	index   uint8
	ability server.Ability
}

type singleAbilityResetTestWorld4FC0B0 struct {
	unitArg       uint64
	unitClass     map[uint64]uint8
	updates       map[uint64]string
	players       map[string]string
	playerClasses map[string]uint8
	playerIndices map[string]uint8
	abilityArg    server.Ability
	cooldowns     map[singleAbilityResetCooldownKey4FC0B0]int32
	head          *singleAbilityResetTestRecord4FC0B0
	allocator     string
	events        []string
	faultAt       int
	after         map[string]func()
}

func singleAbilityResetRecordName4FC0B0(record *singleAbilityResetTestRecord4FC0B0) string {
	if record == nil {
		return "nil"
	}
	return record.name
}

func singleAbilityResetOpaqueName4FC0B0(value string) string {
	if value == "" {
		return "nil"
	}
	return value
}

func (w *singleAbilityResetTestWorld4FC0B0) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *singleAbilityResetTestWorld4FC0B0) hooks() singleAbilityResetHooks4FC0B0[
	uint64,
	string,
	string,
	*singleAbilityResetTestRecord4FC0B0,
	string,
] {
	return singleAbilityResetHooks4FC0B0[
		uint64,
		string,
		string,
		*singleAbilityResetTestRecord4FC0B0,
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
			w.record(fmt.Sprintf("update:%016x=%s", unit, singleAbilityResetOpaqueName4FC0B0(update)))
			return update
		},
		loadPlayer: func(update string) string {
			if update == "" {
				w.record("player:nil-update")
				panic("nil UpdateData")
			}
			player := w.players[update]
			w.record("player:" + update + "=" + singleAbilityResetOpaqueName4FC0B0(player))
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
		loadAbilityArg: func() server.Ability {
			ability := w.abilityArg
			w.record(fmt.Sprintf("ability-arg=%d", ability))
			return ability
		},
		loadPlayerIndex: func(player string) uint8 {
			index := w.playerIndices[player]
			w.record(fmt.Sprintf("player-index:%s=%d", player, index))
			return index
		},
		storeCooldown: func(index uint8, ability server.Ability, value int32) {
			w.record(fmt.Sprintf("cooldown-store:%d:%d=%d", index, ability, value))
			w.cooldowns[singleAbilityResetCooldownKey4FC0B0{index: index, ability: ability}] = value
		},
		resetAbility: func(unit uint64, ability server.Ability) {
			w.record(fmt.Sprintf("reset:%016x:%d", unit, ability))
		},
		loadExecHead: func() *singleAbilityResetTestRecord4FC0B0 {
			head := w.head
			w.record("head=" + singleAbilityResetRecordName4FC0B0(head))
			return head
		},
		storeExecHead: func(record *singleAbilityResetTestRecord4FC0B0) {
			w.record("head-store=" + singleAbilityResetRecordName4FC0B0(record))
			w.head = record
		},
		loadExecUnit: func(record *singleAbilityResetTestRecord4FC0B0) uint64 {
			unit := record.unit
			w.record(fmt.Sprintf("exec-unit:%s=%016x", record.name, unit))
			return unit
		},
		loadExecAbility: func(record *singleAbilityResetTestRecord4FC0B0) server.Ability {
			ability := record.ability
			w.record(fmt.Sprintf("exec-ability:%s=%d", record.name, ability))
			return ability
		},
		loadExecNext: func(record *singleAbilityResetTestRecord4FC0B0) *singleAbilityResetTestRecord4FC0B0 {
			next := record.next
			w.record("exec-next:" + record.name + "=" + singleAbilityResetRecordName4FC0B0(next))
			return next
		},
		loadExecPrev: func(record *singleAbilityResetTestRecord4FC0B0) *singleAbilityResetTestRecord4FC0B0 {
			prev := record.prev
			w.record("exec-prev:" + record.name + "=" + singleAbilityResetRecordName4FC0B0(prev))
			return prev
		},
		storeExecNext: func(record, next *singleAbilityResetTestRecord4FC0B0) {
			w.record("exec-next-store:" + record.name + "=" + singleAbilityResetRecordName4FC0B0(next))
			record.next = next
		},
		storeExecPrev: func(record, prev *singleAbilityResetTestRecord4FC0B0) {
			w.record("exec-prev-store:" + record.name + "=" + singleAbilityResetRecordName4FC0B0(prev))
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
		freeExec: func(allocator string, record *singleAbilityResetTestRecord4FC0B0) {
			w.record("free:" + allocator + ":" + record.name)
			record.freed = true
		},
	}
}

func singleAbilityResetWarriorWorld4FC0B0(unit uint64, ability server.Ability) singleAbilityResetTestWorld4FC0B0 {
	return singleAbilityResetTestWorld4FC0B0{
		unitArg:       unit,
		unitClass:     map[uint64]uint8{unit: singleAbilityResetPlayerClass4FC0B0},
		updates:       map[uint64]string{unit: "update"},
		players:       map[string]string{"update": "player"},
		playerClasses: map[string]uint8{"player": singleAbilityResetWarrior4FC0B0},
		playerIndices: map[string]uint8{"player": 255},
		abilityArg:    ability,
		cooldowns:     make(map[singleAbilityResetCooldownKey4FC0B0]int32),
		allocator:     "pool",
		after:         make(map[string]func()),
	}
}

func TestSingleAbilityReset4FC0B0SignedAbilityNativeIdentityAndAllMatches(t *testing.T) {
	const (
		unit     = uint64(0x1234567889abcdef)
		lowAlias = uint64(0x0000000089abcdef)
		other    = uint64(0xfedcba9876543210)
	)
	ability := server.Ability(-7)
	tail := &singleAbilityResetTestRecord4FC0B0{name: "tail", unit: other, ability: ability, active: 0x44}
	match2 := &singleAbilityResetTestRecord4FC0B0{name: "match-2", unit: unit, ability: ability, active: 0x33, next: tail}
	match1 := &singleAbilityResetTestRecord4FC0B0{name: "match-1", unit: unit, ability: ability, active: 0x22, next: match2}
	wrongAbility := &singleAbilityResetTestRecord4FC0B0{name: "wrong-ability", unit: unit, ability: server.AbilityWarcry, active: 0x11, next: match1}
	low := &singleAbilityResetTestRecord4FC0B0{name: "low-alias", unit: lowAlias, ability: ability, next: wrongAbility}
	wrongAbility.prev = low
	match1.prev = wrongAbility
	match2.prev = match1
	tail.prev = match2
	stale := &singleAbilityResetTestRecord4FC0B0{name: "stale"}

	w := singleAbilityResetWarriorWorld4FC0B0(unit, ability)
	w.head = stale
	w.cooldowns[singleAbilityResetCooldownKey4FC0B0{index: 255, ability: ability}] = 123
	w.after[fmt.Sprintf("reset:%016x:%d", unit, ability)] = func() {
		w.head = low
	}

	singleAbilityReset4FC0B0(w.hooks())

	if got := w.cooldowns[singleAbilityResetCooldownKey4FC0B0{index: 255, ability: ability}]; got != 0 {
		t.Fatalf("cooldown = %d, want 0", got)
	}
	if !match1.freed || !match2.freed || low.freed || wrongAbility.freed || tail.freed || stale.freed {
		t.Fatalf("freed low/wrong/match1/match2/tail/stale = %v/%v/%v/%v/%v/%v", low.freed, wrongAbility.freed, match1.freed, match2.freed, tail.freed, stale.freed)
	}
	if w.head != low || low.next != wrongAbility || wrongAbility.prev != low || wrongAbility.next != tail || tail.prev != wrongAbility {
		t.Fatal("ordinary unlink did not preserve the surrounding doubly linked list")
	}
	if match1.active != 0x22 || match2.active != 0x33 {
		t.Fatal("reset inspected or changed an Active field")
	}

	want := []string{
		"unit-arg=1234567889abcdef", "unit-class:1234567889abcdef=04",
		"update:1234567889abcdef=update", "player:update=player", "player-class:player=0",
		"ability-arg=-7", "player-index:player=255", "cooldown-store:255:-7=0",
		"reset:1234567889abcdef:-7", "head=low-alias",
		"exec-unit:low-alias=0000000089abcdef", "exec-next:low-alias=wrong-ability",
		"exec-unit:wrong-ability=1234567889abcdef", "exec-next:wrong-ability=match-1", "exec-ability:wrong-ability=2",
		"exec-unit:match-1=1234567889abcdef", "exec-next:match-1=match-2", "exec-ability:match-1=-7", "active:1234567889abcdef:-7=0",
		"exec-next:match-1=match-2", "exec-prev:match-1=wrong-ability", "exec-prev-store:match-2=wrong-ability",
		"exec-prev:match-1=wrong-ability", "exec-next:match-1=match-2", "exec-next-store:wrong-ability=match-2", "allocator=pool", "free:pool:match-1",
		"exec-unit:match-2=1234567889abcdef", "exec-next:match-2=tail", "exec-ability:match-2=-7", "active:1234567889abcdef:-7=0",
		"exec-next:match-2=tail", "exec-prev:match-2=wrong-ability", "exec-prev-store:tail=wrong-ability",
		"exec-prev:match-2=wrong-ability", "exec-next:match-2=tail", "exec-next-store:wrong-ability=tail", "allocator=pool", "free:pool:match-2",
		"exec-unit:tail=fedcba9876543210", "exec-next:tail=nil",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("event order:\n got: %q\nwant: %q", w.events, want)
	}
}

func TestSingleAbilityReset4FC0B0CachedReportArgsLiveUnlinkAndCachedTraversal(t *testing.T) {
	const unit = uint64(7)
	other := uint64(9)
	tail := &singleAbilityResetTestRecord4FC0B0{name: "tail", unit: other}
	middle := &singleAbilityResetTestRecord4FC0B0{name: "middle", unit: unit, ability: server.AbilityBerserk, next: tail}
	head := &singleAbilityResetTestRecord4FC0B0{name: "head", unit: unit, ability: server.AbilityBerserk, next: middle}
	middle.prev = head
	tail.prev = middle

	liveHeadNext := &singleAbilityResetTestRecord4FC0B0{name: "live-head-next", prev: head}
	liveMiddlePrev := &singleAbilityResetTestRecord4FC0B0{name: "live-middle-prev", next: middle}
	liveMiddleNext := &singleAbilityResetTestRecord4FC0B0{name: "live-middle-next", prev: middle}

	w := singleAbilityResetWarriorWorld4FC0B0(unit, server.AbilityBerserk)
	w.head = head
	w.after["exec-ability:head=1"] = func() {
		head.unit = 99
		head.ability = server.AbilityWarcry
	}
	w.after["active:0000000000000007:1=0"] = func() {
		if !head.freed {
			head.next = liveHeadNext
			return
		}
		middle.next = liveMiddleNext
		middle.prev = liveMiddlePrev
	}

	singleAbilityReset4FC0B0(w.hooks())

	if !head.freed || !middle.freed || tail.freed || liveHeadNext.freed || liveMiddleNext.freed || liveMiddlePrev.freed {
		t.Fatal("cached traversal or callback-time record selection freed the wrong records")
	}
	if w.head != liveHeadNext || liveHeadNext.prev != nil {
		t.Fatalf("live head unlink = (%s,%s), want (live-head-next,nil)", singleAbilityResetRecordName4FC0B0(w.head), singleAbilityResetRecordName4FC0B0(liveHeadNext.prev))
	}
	if liveMiddleNext.prev != liveMiddlePrev || liveMiddlePrev.next != liveMiddleNext {
		t.Fatal("non-head unlink used cached rather than callback-mutated live links")
	}

	want := []string{
		"unit-arg=0000000000000007", "unit-class:0000000000000007=04",
		"update:0000000000000007=update", "player:update=player", "player-class:player=0",
		"ability-arg=1", "player-index:player=255", "cooldown-store:255:1=0", "reset:0000000000000007:1", "head=head",
		"exec-unit:head=0000000000000007", "exec-next:head=middle", "exec-ability:head=1", "active:0000000000000007:1=0",
		"exec-next:head=live-head-next", "exec-prev:head=nil", "exec-prev-store:live-head-next=nil",
		"exec-prev:head=nil", "exec-next:head=live-head-next", "head-store=live-head-next", "allocator=pool", "free:pool:head",
		"exec-unit:middle=0000000000000007", "exec-next:middle=tail", "exec-ability:middle=1", "active:0000000000000007:1=0",
		"exec-next:middle=live-middle-next", "exec-prev:middle=live-middle-prev", "exec-prev-store:live-middle-next=live-middle-prev",
		"exec-prev:middle=live-middle-prev", "exec-next:middle=live-middle-next", "exec-next-store:live-middle-prev=live-middle-next", "allocator=pool", "free:pool:middle",
		"exec-unit:tail=0000000000000009", "exec-next:tail=nil",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("event order:\n got: %q\nwant: %q", w.events, want)
	}
}

func TestSingleAbilityReset4FC0B0GatesAndRequiredPointerFaults(t *testing.T) {
	t.Run("nil unit", func(t *testing.T) {
		w := singleAbilityResetWarriorWorld4FC0B0(0, server.AbilityBerserk)
		singleAbilityReset4FC0B0(w.hooks())
		if want := []string{"unit-arg=0000000000000000"}; !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})

	t.Run("non-player", func(t *testing.T) {
		w := singleAbilityResetWarriorWorld4FC0B0(1, server.AbilityBerserk)
		w.unitClass[1] = 0x82
		singleAbilityReset4FC0B0(w.hooks())
		want := []string{"unit-arg=0000000000000001", "unit-class:0000000000000001=82"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})

	t.Run("non-Warrior", func(t *testing.T) {
		w := singleAbilityResetWarriorWorld4FC0B0(2, server.Ability(-99))
		w.playerClasses["player"] = 2
		singleAbilityReset4FC0B0(w.hooks())
		want := []string{
			"unit-arg=0000000000000002", "unit-class:0000000000000002=04",
			"update:0000000000000002=update", "player:update=player", "player-class:player=2",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})

	t.Run("nil UpdateData faults", func(t *testing.T) {
		w := singleAbilityResetWarriorWorld4FC0B0(3, server.AbilityBerserk)
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
		singleAbilityReset4FC0B0(w.hooks())
	})

	t.Run("nil Player faults", func(t *testing.T) {
		w := singleAbilityResetWarriorWorld4FC0B0(4, server.AbilityBerserk)
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
		singleAbilityReset4FC0B0(w.hooks())
	})
}

func TestSingleAbilityReset4FC0B0FaultPrefixes(t *testing.T) {
	const unit = uint64(7)
	all := []string{
		"unit-arg=0000000000000007", "unit-class:0000000000000007=04",
		"update:0000000000000007=update", "player:update=player", "player-class:player=0",
		"ability-arg=1", "player-index:player=255", "cooldown-store:255:1=0", "reset:0000000000000007:1",
		"head=record", "exec-unit:record=0000000000000007", "exec-next:record=nil", "exec-ability:record=1",
		"active:0000000000000007:1=0", "exec-next:record=nil", "exec-prev:record=nil", "exec-next:record=nil",
		"head-store=nil", "allocator=pool", "free:pool:record",
	}
	for faultAt := 1; faultAt <= len(all); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			record := &singleAbilityResetTestRecord4FC0B0{name: "record", unit: unit, ability: server.AbilityBerserk}
			w := singleAbilityResetWarriorWorld4FC0B0(unit, server.AbilityBerserk)
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
			singleAbilityReset4FC0B0(w.hooks())
		})
	}
}
