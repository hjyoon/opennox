package opennox

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/opennox/opennox/v1/server"
)

type activeAbilityDisableTestRecord4FC300 struct {
	name     string
	unit     uint64
	ability  server.Ability
	active   uint32
	deadline uint32
	next     *activeAbilityDisableTestRecord4FC300
	prev     *activeAbilityDisableTestRecord4FC300
	freed    bool
}

type activeAbilityDisableTestWorld4FC300 struct {
	unitArg    uint64
	abilityArg server.Ability
	updates    map[uint64]string
	bolts      map[string]uint64
	head       *activeAbilityDisableTestRecord4FC300
	allocator  string
	events     []string
	faultAt    int
	after      map[string]func()
}

func activeAbilityDisableRecordName4FC300(record *activeAbilityDisableTestRecord4FC300) string {
	if record == nil {
		return "nil"
	}
	return record.name
}

func activeAbilityDisableOpaqueName4FC300(value string) string {
	if value == "" {
		return "nil"
	}
	return value
}

func (w *activeAbilityDisableTestWorld4FC300) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *activeAbilityDisableTestWorld4FC300) hooks() activeAbilityDisableHooks4FC300[
	uint64,
	string,
	*activeAbilityDisableTestRecord4FC300,
	string,
] {
	return activeAbilityDisableHooks4FC300[
		uint64,
		string,
		*activeAbilityDisableTestRecord4FC300,
		string,
	]{
		loadUnitArg: func() uint64 {
			unit := w.unitArg
			w.record(fmt.Sprintf("unit-arg=%016x", unit))
			return unit
		},
		loadAbilityArg: func() server.Ability {
			ability := w.abilityArg
			w.record(fmt.Sprintf("ability-arg=%d", ability))
			return ability
		},
		loadUpdateData: func(unit uint64) string {
			update := w.updates[unit]
			w.record(fmt.Sprintf("update:%016x=%s", unit, activeAbilityDisableOpaqueName4FC300(update)))
			return update
		},
		loadHarpoonBolt: func(update string) uint64 {
			if update == "" {
				w.record("bolt:nil-update")
				panic("nil UpdateData")
			}
			bolt := w.bolts[update]
			w.record(fmt.Sprintf("bolt:%s=%016x", update, bolt))
			return bolt
		},
		breakHarpoon: func(unit, bolt uint64) {
			w.record(fmt.Sprintf("harpoon-break:%016x:%016x", unit, bolt))
		},
		disableEnchant: func(unit uint64, enchant int32) {
			w.record(fmt.Sprintf("disable-enchant:%016x:%d", unit, enchant))
		},
		reportActive: func(unit uint64, ability server.Ability, active int32) {
			w.record(fmt.Sprintf("active:%016x:%d=%d", unit, ability, active))
		},
		loadExecHead: func() *activeAbilityDisableTestRecord4FC300 {
			head := w.head
			w.record("head=" + activeAbilityDisableRecordName4FC300(head))
			return head
		},
		storeExecHead: func(record *activeAbilityDisableTestRecord4FC300) {
			w.record("head-store=" + activeAbilityDisableRecordName4FC300(record))
			w.head = record
		},
		loadExecUnit: func(record *activeAbilityDisableTestRecord4FC300) uint64 {
			unit := record.unit
			w.record(fmt.Sprintf("exec-unit:%s=%016x", record.name, unit))
			return unit
		},
		loadExecAbility: func(record *activeAbilityDisableTestRecord4FC300) server.Ability {
			ability := record.ability
			w.record(fmt.Sprintf("exec-ability:%s=%d", record.name, ability))
			return ability
		},
		loadExecNext: func(record *activeAbilityDisableTestRecord4FC300) *activeAbilityDisableTestRecord4FC300 {
			next := record.next
			w.record("exec-next:" + record.name + "=" + activeAbilityDisableRecordName4FC300(next))
			return next
		},
		loadExecPrev: func(record *activeAbilityDisableTestRecord4FC300) *activeAbilityDisableTestRecord4FC300 {
			prev := record.prev
			w.record("exec-prev:" + record.name + "=" + activeAbilityDisableRecordName4FC300(prev))
			return prev
		},
		storeExecNext: func(record, next *activeAbilityDisableTestRecord4FC300) {
			w.record("exec-next-store:" + record.name + "=" + activeAbilityDisableRecordName4FC300(next))
			record.next = next
		},
		storeExecPrev: func(record, prev *activeAbilityDisableTestRecord4FC300) {
			w.record("exec-prev-store:" + record.name + "=" + activeAbilityDisableRecordName4FC300(prev))
			record.prev = prev
		},
		loadExecAllocator: func() string {
			allocator := w.allocator
			w.record("allocator=" + allocator)
			return allocator
		},
		freeExec: func(allocator string, record *activeAbilityDisableTestRecord4FC300) {
			w.record("free:" + allocator + ":" + record.name)
			record.freed = true
		},
	}
}

func activeAbilityDisableWorld4FC300(unit uint64, ability server.Ability) activeAbilityDisableTestWorld4FC300 {
	return activeAbilityDisableTestWorld4FC300{
		unitArg:    unit,
		abilityArg: ability,
		updates:    map[uint64]string{unit: "update"},
		bolts:      map[string]uint64{"update": 0xfedcba9876543210},
		allocator:  "pool",
		after:      make(map[string]func()),
	}
}

func TestActiveAbilityDisable4FC300SignedGatesAndSpecialPaths(t *testing.T) {
	const unit = uint64(0x1234567889abcdef)
	tests := []struct {
		name    string
		unit    uint64
		ability server.Ability
		want    []string
	}{
		{name: "nil unit", unit: 0, ability: server.AbilityHarpoon, want: []string{
			"unit-arg=0000000000000000",
		}},
		{name: "minimum signed", unit: unit, ability: server.Ability(-1 << 31), want: []string{
			"unit-arg=1234567889abcdef", "ability-arg=-2147483648",
		}},
		{name: "negative", unit: unit, ability: -1, want: []string{
			"unit-arg=1234567889abcdef", "ability-arg=-1",
		}},
		{name: "zero", unit: unit, ability: server.AbilityInvalid, want: []string{
			"unit-arg=1234567889abcdef", "ability-arg=0",
		}},
		{name: "maximum", unit: unit, ability: server.AbilityMax, want: []string{
			"unit-arg=1234567889abcdef", "ability-arg=6",
		}},
		{name: "maximum signed", unit: unit, ability: server.Ability(1<<31 - 1), want: []string{
			"unit-arg=1234567889abcdef", "ability-arg=2147483647",
		}},
		{name: "berserk", unit: unit, ability: server.AbilityBerserk, want: []string{
			"unit-arg=1234567889abcdef", "ability-arg=1",
			"active:1234567889abcdef:1=0", "head=nil",
		}},
		{name: "warcry", unit: unit, ability: server.AbilityWarcry, want: []string{
			"unit-arg=1234567889abcdef", "ability-arg=2",
			"active:1234567889abcdef:2=0", "head=nil",
		}},
		{name: "harpoon", unit: unit, ability: server.AbilityHarpoon, want: []string{
			"unit-arg=1234567889abcdef", "ability-arg=3",
			"update:1234567889abcdef=update", "bolt:update=fedcba9876543210",
			"harpoon-break:1234567889abcdef:fedcba9876543210",
			"active:1234567889abcdef:3=0", "head=nil",
		}},
		{name: "tread lightly", unit: unit, ability: server.AbilityTreadLightly, want: []string{
			"unit-arg=1234567889abcdef", "ability-arg=4",
			"disable-enchant:1234567889abcdef:31",
			"active:1234567889abcdef:4=0", "head=nil",
		}},
		{name: "infravis", unit: unit, ability: server.AbilityInfravis, want: []string{
			"unit-arg=1234567889abcdef", "ability-arg=5",
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := activeAbilityDisableWorld4FC300(tc.unit, tc.ability)
			activeAbilityDisable4FC300(w.hooks())
			if !reflect.DeepEqual(w.events, tc.want) {
				t.Fatalf("events = %q, want %q", w.events, tc.want)
			}
		})
	}
}

func TestActiveAbilityDisable4FC300HarpoonRequiredUpdateDataFault(t *testing.T) {
	const unit = uint64(7)
	w := activeAbilityDisableWorld4FC300(unit, server.AbilityHarpoon)
	w.updates[unit] = ""
	defer func() {
		if recover() == nil {
			t.Fatal("nil UpdateData did not fault while loading HarpoonBolt")
		}
		want := []string{
			"unit-arg=0000000000000007", "ability-arg=3",
			"update:0000000000000007=nil", "bolt:nil-update",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	}()
	activeAbilityDisable4FC300(w.hooks())
}

func TestActiveAbilityDisable4FC300NativeIdentityReportBeforeHeadAndAllMatches(t *testing.T) {
	const (
		unit     = uint64(0x1234567889abcdef)
		lowAlias = uint64(0x0000000089abcdef)
		other    = uint64(0xfedcba9876543210)
	)
	tail := &activeAbilityDisableTestRecord4FC300{name: "tail", unit: other, ability: server.AbilityBerserk, active: 0x55, deadline: 0x65}
	match2 := &activeAbilityDisableTestRecord4FC300{name: "match-2", unit: unit, ability: server.AbilityBerserk, active: 0x44, deadline: 0x64, next: tail}
	match1 := &activeAbilityDisableTestRecord4FC300{name: "match-1", unit: unit, ability: server.AbilityBerserk, active: 0x33, deadline: 0x63, next: match2}
	wrongAbility := &activeAbilityDisableTestRecord4FC300{name: "wrong-ability", unit: unit, ability: server.AbilityWarcry, active: 0x22, deadline: 0x62, next: match1}
	low := &activeAbilityDisableTestRecord4FC300{name: "low-alias", unit: lowAlias, ability: server.AbilityBerserk, active: 0x11, deadline: 0x61, next: wrongAbility}
	wrongAbility.prev = low
	match1.prev = wrongAbility
	match2.prev = match1
	tail.prev = match2
	stale := &activeAbilityDisableTestRecord4FC300{name: "stale"}

	w := activeAbilityDisableWorld4FC300(unit, server.AbilityBerserk)
	w.head = stale
	w.after["unit-arg=1234567889abcdef"] = func() { w.unitArg = lowAlias }
	w.after["ability-arg=1"] = func() { w.abilityArg = server.AbilityWarcry }
	w.after["active:1234567889abcdef:1=0"] = func() { w.head = low }

	activeAbilityDisable4FC300(w.hooks())

	if !match1.freed || !match2.freed || low.freed || wrongAbility.freed || tail.freed || stale.freed {
		t.Fatalf("freed low/wrong/match1/match2/tail/stale = %v/%v/%v/%v/%v/%v", low.freed, wrongAbility.freed, match1.freed, match2.freed, tail.freed, stale.freed)
	}
	if w.head != low || low.next != wrongAbility || wrongAbility.prev != low || wrongAbility.next != tail || tail.prev != wrongAbility {
		t.Fatal("ordinary unlink did not preserve the surrounding doubly linked list")
	}
	if match1.active != 0x33 || match1.deadline != 0x63 || match2.active != 0x44 || match2.deadline != 0x64 {
		t.Fatal("disable inspected or changed an Active/deadline field")
	}

	want := []string{
		"unit-arg=1234567889abcdef", "ability-arg=1", "active:1234567889abcdef:1=0", "head=low-alias",
		"exec-unit:low-alias=0000000089abcdef", "exec-next:low-alias=wrong-ability",
		"exec-unit:wrong-ability=1234567889abcdef", "exec-next:wrong-ability=match-1", "exec-ability:wrong-ability=2",
		"exec-unit:match-1=1234567889abcdef", "exec-next:match-1=match-2", "exec-ability:match-1=1",
		"exec-next:match-1=match-2", "exec-prev:match-1=wrong-ability", "exec-prev-store:match-2=wrong-ability",
		"exec-prev:match-1=wrong-ability", "exec-next:match-1=match-2", "exec-next-store:wrong-ability=match-2", "allocator=pool", "free:pool:match-1",
		"exec-unit:match-2=1234567889abcdef", "exec-next:match-2=tail", "exec-ability:match-2=1",
		"exec-next:match-2=tail", "exec-prev:match-2=wrong-ability", "exec-prev-store:tail=wrong-ability",
		"exec-prev:match-2=wrong-ability", "exec-next:match-2=tail", "exec-next-store:wrong-ability=tail", "allocator=pool", "free:pool:match-2",
		"exec-unit:tail=fedcba9876543210", "exec-next:tail=nil",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("event order:\n got: %q\nwant: %q", w.events, want)
	}
}

func TestActiveAbilityDisable4FC300CachedSelectionLiveUnlinkAndCachedTraversal(t *testing.T) {
	const unit = uint64(7)
	tail := &activeAbilityDisableTestRecord4FC300{name: "tail", unit: 9}
	middle := &activeAbilityDisableTestRecord4FC300{name: "middle", unit: unit, ability: server.AbilityBerserk, next: tail}
	head := &activeAbilityDisableTestRecord4FC300{name: "head", unit: unit, ability: server.AbilityBerserk, next: middle}
	middle.prev = head
	tail.prev = middle

	liveHeadNext := &activeAbilityDisableTestRecord4FC300{name: "live-head-next", prev: head}
	liveMiddlePrev := &activeAbilityDisableTestRecord4FC300{name: "live-middle-prev", next: middle}
	liveMiddleNext := &activeAbilityDisableTestRecord4FC300{name: "live-middle-next", prev: middle}

	w := activeAbilityDisableWorld4FC300(unit, server.AbilityBerserk)
	w.head = head
	w.after["exec-next:head=middle"] = func() { head.unit = 99 }
	w.after["exec-ability:head=1"] = func() {
		head.ability = server.AbilityWarcry
		head.next = liveHeadNext
	}
	w.after["exec-next:middle=tail"] = func() { middle.unit = 99 }
	w.after["exec-ability:middle=1"] = func() {
		middle.ability = server.AbilityWarcry
		middle.next = liveMiddleNext
		middle.prev = liveMiddlePrev
	}

	activeAbilityDisable4FC300(w.hooks())

	if !head.freed || !middle.freed || tail.freed || liveHeadNext.freed || liveMiddleNext.freed || liveMiddlePrev.freed {
		t.Fatal("cached traversal or cached selection freed the wrong records")
	}
	if w.head != liveHeadNext || liveHeadNext.prev != nil {
		t.Fatalf("live head unlink = (%s,%s), want (live-head-next,nil)", activeAbilityDisableRecordName4FC300(w.head), activeAbilityDisableRecordName4FC300(liveHeadNext.prev))
	}
	if liveMiddleNext.prev != liveMiddlePrev || liveMiddlePrev.next != liveMiddleNext {
		t.Fatal("non-head unlink used cached rather than live links")
	}

	want := []string{
		"unit-arg=0000000000000007", "ability-arg=1", "active:0000000000000007:1=0", "head=head",
		"exec-unit:head=0000000000000007", "exec-next:head=middle", "exec-ability:head=1",
		"exec-next:head=live-head-next", "exec-prev:head=nil", "exec-prev-store:live-head-next=nil",
		"exec-prev:head=nil", "exec-next:head=live-head-next", "head-store=live-head-next", "allocator=pool", "free:pool:head",
		"exec-unit:middle=0000000000000007", "exec-next:middle=tail", "exec-ability:middle=1",
		"exec-next:middle=live-middle-next", "exec-prev:middle=live-middle-prev", "exec-prev-store:live-middle-next=live-middle-prev",
		"exec-prev:middle=live-middle-prev", "exec-next:middle=live-middle-next", "exec-next-store:live-middle-prev=live-middle-next", "allocator=pool", "free:pool:middle",
		"exec-unit:tail=0000000000000009", "exec-next:tail=nil",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("event order:\n got: %q\nwant: %q", w.events, want)
	}
}

func TestActiveAbilityDisable4FC300FaultPrefixes(t *testing.T) {
	const unit = uint64(7)
	tail := &activeAbilityDisableTestRecord4FC300{name: "tail", unit: 9}
	nonHead := &activeAbilityDisableTestRecord4FC300{name: "non-head", unit: unit, ability: server.AbilityHarpoon, next: tail}
	decoy := &activeAbilityDisableTestRecord4FC300{name: "decoy", unit: 9, next: nonHead}
	head := &activeAbilityDisableTestRecord4FC300{name: "head", unit: unit, ability: server.AbilityHarpoon, next: decoy}
	decoy.prev = head
	nonHead.prev = decoy
	tail.prev = nonHead

	all := []string{
		"unit-arg=0000000000000007", "ability-arg=3", "update:0000000000000007=update",
		"bolt:update=fedcba9876543210", "harpoon-break:0000000000000007:fedcba9876543210",
		"active:0000000000000007:3=0", "head=head",
		"exec-unit:head=0000000000000007", "exec-next:head=decoy", "exec-ability:head=3",
		"exec-next:head=decoy", "exec-prev:head=nil", "exec-prev-store:decoy=nil",
		"exec-prev:head=nil", "exec-next:head=decoy", "head-store=decoy", "allocator=pool", "free:pool:head",
		"exec-unit:decoy=0000000000000009", "exec-next:decoy=non-head",
		"exec-unit:non-head=0000000000000007", "exec-next:non-head=tail", "exec-ability:non-head=3",
		"exec-next:non-head=tail", "exec-prev:non-head=decoy", "exec-prev-store:tail=decoy",
		"exec-prev:non-head=decoy", "exec-next:non-head=tail", "exec-next-store:decoy=tail", "allocator=pool", "free:pool:non-head",
		"exec-unit:tail=0000000000000009", "exec-next:tail=nil",
	}
	for faultAt := 1; faultAt <= len(all); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			localTail := *tail
			localNonHead := *nonHead
			localDecoy := *decoy
			localHead := *head
			localHead.next = &localDecoy
			localDecoy.prev = &localHead
			localDecoy.next = &localNonHead
			localNonHead.prev = &localDecoy
			localNonHead.next = &localTail
			localTail.prev = &localNonHead

			w := activeAbilityDisableWorld4FC300(unit, server.AbilityHarpoon)
			w.head = &localHead
			w.faultAt = faultAt
			defer func() {
				if recover() == nil {
					t.Fatal("fault was not propagated")
				}
				if want := all[:faultAt]; !reflect.DeepEqual(w.events, want) {
					t.Fatalf("events = %q, want prefix %q", w.events, want)
				}
			}()
			activeAbilityDisable4FC300(w.hooks())
		})
	}
}
