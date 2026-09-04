package opennox

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

type spellCancellationTestRecord4FCEB0 struct {
	name   string
	next   *spellCancellationTestRecord4FCEB0
	target uint64
}

type spellCancellationTestWorld4FCEB0 struct {
	head    *spellCancellationTestRecord4FCEB0
	classes map[uint64]uint32
	events  []string
	faultAt int
	after   map[string]func()
}

func spellCancellationRecordName4FCEB0(record *spellCancellationTestRecord4FCEB0) string {
	if record == nil {
		return "nil"
	}
	return record.name
}

func (w *spellCancellationTestWorld4FCEB0) record(event string) {
	if w.faultAt != 0 && len(w.events)+1 == w.faultAt {
		panic(event)
	}
	w.events = append(w.events, event)
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *spellCancellationTestWorld4FCEB0) hooks() spellCancellationTraversalHooks4FCEB0[*spellCancellationTestRecord4FCEB0, uint64] {
	return spellCancellationTraversalHooks4FCEB0[*spellCancellationTestRecord4FCEB0, uint64]{
		firstSpell: func() *spellCancellationTestRecord4FCEB0 {
			head := w.head
			w.record("first=" + spellCancellationRecordName4FCEB0(head))
			return head
		},
		loadNext: func(current *spellCancellationTestRecord4FCEB0) *spellCancellationTestRecord4FCEB0 {
			next := current.next
			w.record("next:" + current.name + "=" + spellCancellationRecordName4FCEB0(next))
			return next
		},
		loadTarget: func(current *spellCancellationTestRecord4FCEB0) uint64 {
			target := current.target
			w.record(fmt.Sprintf("target:%s=%016x", current.name, target))
			return target
		},
		loadTargetClass: func(target uint64) uint32 {
			class := w.classes[target]
			w.record(fmt.Sprintf("class:%016x=%08x", target, class))
			return class
		},
		cancelSpell: func(current *spellCancellationTestRecord4FCEB0) {
			w.record("cancel:" + current.name)
		},
	}
}

func TestSpellCancellationTraversal4FCEB0EmptyReturnsCanonicalZero(t *testing.T) {
	w := spellCancellationTestWorld4FCEB0{
		classes: make(map[uint64]uint32),
		after:   make(map[string]func()),
	}
	if got := spellCancellationTraversal4FCEB0(1, w.hooks()); got != 0 {
		t.Fatalf("result = %d, want canonical 0", got)
	}
	if want := []string{"first=nil"}; !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestSpellCancellationTraversal4FCEB0NonOneSkipsTargetsAndUsesCachedNext(t *testing.T) {
	const (
		target1 = uint64(0x100001234)
		target2 = uint64(0x200002345)
	)
	decoy := &spellCancellationTestRecord4FCEB0{name: "decoy", target: target1}
	second := &spellCancellationTestRecord4FCEB0{name: "second", target: target2}
	first := &spellCancellationTestRecord4FCEB0{name: "first", next: second, target: target1}
	w := spellCancellationTestWorld4FCEB0{
		head:    first,
		classes: map[uint64]uint32{target1: spellCancellationPlayerClass4FCEB0, target2: spellCancellationPlayerClass4FCEB0},
		after:   make(map[string]func()),
	}
	w.after["cancel:first"] = func() { first.next = decoy }

	for _, mode := range []int32{0, 2, -1, -2147483648} {
		w.events = nil
		first.next = second
		if got := spellCancellationTraversal4FCEB0(mode, w.hooks()); got != 0 {
			t.Fatalf("mode %d result = %d, want canonical 0", mode, got)
		}
		want := []string{
			"first=first", "next:first=second", "cancel:first",
			"next:second=nil", "cancel:second",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("mode %d events = %q, want %q", mode, w.events, want)
		}
	}
}

func TestSpellCancellationTraversal4FCEB0ModeOneTargetGatesAndLoadOrder(t *testing.T) {
	const (
		playerTarget    = uint64(0x1234567889abcdef)
		nonPlayerTarget = uint64(0xfedcba9876543210)
		lateTarget      = uint64(0x700007777)
	)
	last := &spellCancellationTestRecord4FCEB0{name: "last", target: lateTarget}
	nonPlayer := &spellCancellationTestRecord4FCEB0{name: "non-player", next: last, target: nonPlayerTarget}
	nilTarget := &spellCancellationTestRecord4FCEB0{name: "nil-target", next: nonPlayer}
	player := &spellCancellationTestRecord4FCEB0{name: "player", next: nilTarget, target: 1}
	w := spellCancellationTestWorld4FCEB0{
		head: player,
		classes: map[uint64]uint32{
			playerTarget:    0xa5a50004,
			nonPlayerTarget: 0xfffffff8,
			lateTarget:      0x04,
		},
		after: make(map[string]func()),
	}
	w.after["next:player=nil-target"] = func() { player.target = playerTarget }
	decoyTarget := uint64(0x900009999)
	w.after[fmt.Sprintf("target:player=%016x", playerTarget)] = func() { player.target = decoyTarget }

	if got := spellCancellationTraversal4FCEB0(1, w.hooks()); got != 0 {
		t.Fatalf("result = %d, want canonical 0", got)
	}
	want := []string{
		"first=player",
		"next:player=nil-target",
		fmt.Sprintf("target:player=%016x", playerTarget),
		fmt.Sprintf("class:%016x=%08x", playerTarget, uint32(0xa5a50004)),
		"next:nil-target=non-player",
		"target:nil-target=0000000000000000",
		"cancel:nil-target",
		"next:non-player=last",
		fmt.Sprintf("target:non-player=%016x", nonPlayerTarget),
		fmt.Sprintf("class:%016x=%08x", nonPlayerTarget, uint32(0xfffffff8)),
		"cancel:non-player",
		"next:last=nil",
		fmt.Sprintf("target:last=%016x", lateTarget),
		fmt.Sprintf("class:%016x=00000004", lateTarget),
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestSpellCancellationTraversal4FCEB0CancelMutationCannotRedirectTraversal(t *testing.T) {
	const target = uint64(0x100000111)
	decoy := &spellCancellationTestRecord4FCEB0{name: "decoy", target: target}
	second := &spellCancellationTestRecord4FCEB0{name: "second", target: target}
	first := &spellCancellationTestRecord4FCEB0{name: "first", next: second, target: target}
	w := spellCancellationTestWorld4FCEB0{
		head:    first,
		classes: map[uint64]uint32{target: 0},
		after:   make(map[string]func()),
	}
	w.after["cancel:first"] = func() {
		first.next = decoy
		second.target = 0
	}

	spellCancellationTraversal4FCEB0(1, w.hooks())
	want := []string{
		"first=first", "next:first=second",
		"target:first=0000000100000111", "class:0000000100000111=00000000", "cancel:first",
		"next:second=nil", "target:second=0000000000000000", "cancel:second",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestSpellCancellationTraversalNative4FCEB0BindsNativePointersAndFields(t *testing.T) {
	playerTarget := &server.Object{ObjClass: object.ClassPlayer | object.ClassMonster}
	nonPlayerTarget := &server.Object{ObjClass: object.ClassMonster}
	third := &server.DurSpell{}
	second := &server.DurSpell{Target48: playerTarget, Next: third}
	first := &server.DurSpell{Target48: nonPlayerTarget, Next: second}
	decoy := &server.DurSpell{Target48: nonPlayerTarget}

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"player target":     unsafe.Pointer(playerTarget),
			"non-player target": unsafe.Pointer(nonPlayerTarget),
			"first spell":       unsafe.Pointer(first),
			"second spell":      unsafe.Pointer(second),
			"third spell":       unsafe.Pointer(third),
			"decoy spell":       unsafe.Pointer(decoy),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want a native address above 4 GiB", name, pointer)
			}
		}
	}

	var cancelled []*server.DurSpell
	got := spellCancellationTraversalNative4FCEB0(
		1,
		spellCancellationTraversalNativeDeps4FCEB0{
			firstSpell: func() *server.DurSpell { return first },
			cancelSpell: func(current *server.DurSpell) {
				cancelled = append(cancelled, current)
				if current == first {
					first.Next = decoy
				}
			},
		},
	)
	if got != 0 {
		t.Fatalf("result = %d, want canonical 0", got)
	}
	if want := []*server.DurSpell{first, third}; !reflect.DeepEqual(cancelled, want) {
		t.Fatalf("cancelled = %v, want %v", cancelled, want)
	}
	runtime.KeepAlive(playerTarget)
	runtime.KeepAlive(nonPlayerTarget)
	runtime.KeepAlive(first)
	runtime.KeepAlive(second)
	runtime.KeepAlive(third)
	runtime.KeepAlive(decoy)
}

func TestSpellCancellationTraversal4FCEB0ProductionBinding(t *testing.T) {
	first := &server.DurSpell{}
	second := &server.DurSpell{}
	first.Next = second
	base := new(server.Server)
	base.Spells.Dur.List = first
	s := &Server{Server: base}

	if got := s.SpellCancellationTraversal4FCEB0(2); got != 0 {
		t.Fatalf("result = %d, want canonical 0", got)
	}
	if first.Flags88&1 == 0 || second.Flags88&1 == 0 {
		t.Fatalf("cancel flags = %#x/%#x, want low bits set", first.Flags88, second.Flags88)
	}
}

func TestSpellCancellationTraversal4FCEB0FaultPrefixes(t *testing.T) {
	const (
		playerTarget = uint64(0x100001111)
		otherTarget  = uint64(0x200002222)
	)
	second := &spellCancellationTestRecord4FCEB0{name: "second", target: otherTarget}
	first := &spellCancellationTestRecord4FCEB0{name: "first", next: second, target: playerTarget}
	allEvents := []string{
		"first=first",
		"next:first=second", "target:first=0000000100001111", "class:0000000100001111=00000004",
		"next:second=nil", "target:second=0000000200002222", "class:0000000200002222=00000000", "cancel:second",
	}

	for faultAt := 1; faultAt <= len(allEvents); faultAt++ {
		t.Run(fmt.Sprintf("fault-%02d-%s", faultAt, allEvents[faultAt-1]), func(t *testing.T) {
			w := spellCancellationTestWorld4FCEB0{
				head:    first,
				classes: map[uint64]uint32{playerTarget: spellCancellationPlayerClass4FCEB0, otherTarget: 0},
				events:  make([]string, 0, len(allEvents)),
				faultAt: faultAt,
				after:   make(map[string]func()),
			}
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				spellCancellationTraversal4FCEB0(1, w.hooks())
			}()
			if recovered != allEvents[faultAt-1] {
				t.Fatalf("recovered = %#v, want %q", recovered, allEvents[faultAt-1])
			}
			if want := allEvents[:faultAt-1]; !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %q, want prefix %q", w.events, want)
			}
		})
	}
}
