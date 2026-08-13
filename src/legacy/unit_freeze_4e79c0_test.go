package legacy

import (
	"fmt"
	"reflect"
	"testing"
)

type unitFreezeFixture4E79C0 struct {
	name       string
	flags      uint32
	class      uint32
	first      *unitFreezeFixture4E79C0
	next       *unitFreezeFixture4E79C0
	loadByte   byte
	summoned   bool
	reportByte byte
	pushByte   byte
	popByte    byte
	onReport   func(*unitFreezeFixture4E79C0)
	onPush     func(*unitFreezeFixture4E79C0)
	onPop      func(*unitFreezeFixture4E79C0)
}

func unitFreezeTestHooks4E79C0(events *[]string, gate *uint32) unitFreezeHooks4E79C0[*unitFreezeFixture4E79C0] {
	return unitFreezeHooks4E79C0[*unitFreezeFixture4E79C0]{
		flags: func(obj *unitFreezeFixture4E79C0) uint32 {
			name := ""
			if obj != nil {
				name = obj.name
			}
			*events = append(*events, "flags:"+name)
			return obj.flags
		},
		setFlags: func(obj *unitFreezeFixture4E79C0, flags uint32) {
			*events = append(*events, fmt.Sprintf("set:%s:%08x", obj.name, flags))
			obj.flags = flags
		},
		class: func(obj *unitFreezeFixture4E79C0) uint32 {
			*events = append(*events, "class:"+obj.name)
			return obj.class
		},
		gate: func() uint32 {
			*events = append(*events, "gate")
			return *gate
		},
		setGate: func(value uint32) {
			*events = append(*events, fmt.Sprintf("set-gate:%08x", value))
			*gate = value
		},
		reportStatus: func(obj *unitFreezeFixture4E79C0) byte {
			*events = append(*events, "report:"+obj.name)
			if obj.onReport != nil {
				obj.onReport(obj)
			}
			return obj.reportByte
		},
		setPlayerIdle: func(obj *unitFreezeFixture4E79C0) {
			*events = append(*events, "idle:"+obj.name)
		},
		raiseZero: func(obj *unitFreezeFixture4E79C0) {
			*events = append(*events, "raise:"+obj.name)
		},
		resetPaths: func() {
			*events = append(*events, "paths")
		},
		firstOwned: func(obj *unitFreezeFixture4E79C0) *unitFreezeFixture4E79C0 {
			*events = append(*events, "first:"+obj.name)
			return obj.first
		},
		nextOwned: func(obj *unitFreezeFixture4E79C0) *unitFreezeFixture4E79C0 {
			*events = append(*events, "next:"+obj.name)
			return obj.next
		},
		monsterStatus: func(obj *unitFreezeFixture4E79C0) (byte, bool) {
			*events = append(*events, "status:"+obj.name)
			return obj.loadByte, obj.summoned
		},
		pushIdle: func(obj *unitFreezeFixture4E79C0) byte {
			*events = append(*events, "push:"+obj.name)
			if obj.onPush != nil {
				obj.onPush(obj)
			}
			return obj.pushByte
		},
		popAction: func(obj *unitFreezeFixture4E79C0) byte {
			*events = append(*events, "pop:"+obj.name)
			if obj.onPop != nil {
				obj.onPop(obj)
			}
			return obj.popByte
		},
	}
}

func TestUnitFreeze4E79C0AlreadyFrozenShortCircuits(t *testing.T) {
	obj := &unitFreezeFixture4E79C0{name: "obj", flags: 0xa5a50002, class: unitPlayerClass4E79C0}
	gate := uint32(9)
	var events []string
	got := unitFreeze4E79C0(obj, 7, unitFreezeTestHooks4E79C0(&events, &gate))
	if got != 2 || obj.flags != 0xa5a50002 || gate != 9 {
		t.Fatalf("result/state = (%#x, %#x, %#x), want (2, 0xa5a50002, 9)", got, obj.flags, gate)
	}
	if want := []string{"flags:obj"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestUnitFreeze4E79C0PreservesExistingGate(t *testing.T) {
	obj := &unitFreezeFixture4E79C0{name: "player", class: unitPlayerClass4E79C0}
	gate := uint32(0x12345678)
	var events []string
	got := unitFreeze4E79C0(obj, 0x80000001, unitFreezeTestHooks4E79C0(&events, &gate))
	if got != byte(unitPlayerClass4E79C0) || gate != 0x12345678 {
		t.Fatalf("result/gate = (%#x, %#x), want player class and existing gate", got, gate)
	}
	for _, event := range events {
		if event == "set-gate:80000001" {
			t.Fatal("nonzero gate was overwritten")
		}
	}
}

func TestUnitFreeze4E79C0PlayerRecursesAndReloadsSuccessor(t *testing.T) {
	root := &unitFreezeFixture4E79C0{name: "root", class: unitPlayerClass4E79C0}
	child1 := &unitFreezeFixture4E79C0{name: "child1", class: unitMonsterClass4E79C0, summoned: true, pushByte: 0x55}
	child2 := &unitFreezeFixture4E79C0{name: "child2", class: unitMonsterClass4E79C0, summoned: true}
	child3 := &unitFreezeFixture4E79C0{name: "child3"}
	root.first = child1
	child1.next = child2
	child1.onPush = func(*unitFreezeFixture4E79C0) { child1.next = child3 }
	gate := uint32(0)
	var events []string

	got := unitFreeze4E79C0(root, 0x80000001, unitFreezeTestHooks4E79C0(&events, &gate))
	if got != byte(unitPlayerClass4E79C0) || gate != 0x80000001 {
		t.Fatalf("result/gate = (%#x, %#x), want (4, 0x80000001)", got, gate)
	}
	if root.flags&unitFreezeFlag4E79C0 == 0 || child1.flags&unitFreezeFlag4E79C0 == 0 {
		t.Fatalf("frozen flags = (%#x, %#x), want both set", root.flags, child1.flags)
	}
	if child2.flags != 0 {
		t.Fatalf("stale successor was visited: flags %#x", child2.flags)
	}
	want := []string{
		"flags:root", "set:root:00000002", "class:root", "gate", "set-gate:80000001",
		"report:root", "idle:root", "raise:root", "paths", "first:root",
		"class:child1", "status:child1", "flags:child1", "set:child1:00000002",
		"class:child1", "class:child1", "flags:child1", "push:child1",
		"next:child1", "class:child3", "next:child3", "class:root",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events =\n%v\nwant\n%v", events, want)
	}
}

func TestUnitFreeze4E79C0DeadMonsterPreservesReturnByte(t *testing.T) {
	obj := &unitFreezeFixture4E79C0{name: "monster", flags: 0x123480f0, class: unitMonsterClass4E79C0, pushByte: 0xee}
	gate := uint32(0)
	var events []string
	got := unitFreeze4E79C0(obj, 0, unitFreezeTestHooks4E79C0(&events, &gate))
	if got != 0xf2 || obj.flags != 0x123480f2 {
		t.Fatalf("result/flags = (%#x, %#x), want (0xf2, 0x123480f2)", got, obj.flags)
	}
	for _, event := range events {
		if event == "push:monster" {
			t.Fatal("dead monster pushed an idle action")
		}
	}
}

func TestUnitUnfreeze4E7A60GateBlocksNonForcedPlayer(t *testing.T) {
	obj := &unitFreezeFixture4E79C0{name: "player", flags: 0x11220002, class: unitPlayerClass4E79C0}
	gate := uint32(0x123456a5)
	var events []string
	got := unitUnfreeze4E7A60(obj, 0, unitFreezeTestHooks4E79C0(&events, &gate))
	if got != 0xa5 || obj.flags != 0x11220002 || gate != 0x123456a5 {
		t.Fatalf("result/state = (%#x, %#x, %#x), want blocked original state", got, obj.flags, gate)
	}
	if want := []string{"flags:player", "class:player", "gate"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestUnitUnfreeze4E7A60AlreadyThawedShortCircuits(t *testing.T) {
	obj := &unitFreezeFixture4E79C0{name: "obj", flags: 0xa5a50001, class: unitPlayerClass4E79C0}
	gate := uint32(9)
	var events []string
	got := unitUnfreeze4E7A60(obj, 0xffffffff, unitFreezeTestHooks4E79C0(&events, &gate))
	if got != 1 || obj.flags != 0xa5a50001 || gate != 9 {
		t.Fatalf("result/state = (%#x, %#x, %#x), want unchanged", got, obj.flags, gate)
	}
	if want := []string{"flags:obj"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestUnitUnfreeze4E7A60ForcedPlayerReloadsAndRecurses(t *testing.T) {
	root := &unitFreezeFixture4E79C0{name: "root", flags: 0x00000102, class: unitPlayerClass4E79C0, reportByte: 0x91}
	child1 := &unitFreezeFixture4E79C0{name: "child1", flags: unitFreezeFlag4E79C0, class: unitMonsterClass4E79C0, summoned: true, popByte: 0x33}
	child2 := &unitFreezeFixture4E79C0{name: "child2", flags: unitFreezeFlag4E79C0, class: unitMonsterClass4E79C0, summoned: true}
	child3 := &unitFreezeFixture4E79C0{name: "child3", flags: unitFreezeFlag4E79C0, class: unitMonsterClass4E79C0, loadByte: 0x77}
	root.first = child1
	child1.next = child2
	child1.onPop = func(*unitFreezeFixture4E79C0) { child1.next = child3 }
	gate := uint32(0x11223344)
	var events []string
	hooks := unitFreezeTestHooks4E79C0(&events, &gate)
	originalGate := hooks.gate
	hooks.gate = func() uint32 {
		root.flags = 0x40000002
		return originalGate()
	}

	got := unitUnfreeze4E7A60(root, 0x80000000, hooks)
	if got != 0x77 || gate != 0 || root.flags != 0x40000000 || child1.flags != 0 {
		t.Fatalf("result/state = (%#x, %#x, %#x, %#x), want load artifact and cleared live flags", got, gate, root.flags, child1.flags)
	}
	if child2.flags != unitFreezeFlag4E79C0 || child3.flags != unitFreezeFlag4E79C0 {
		t.Fatalf("successor states = (%#x, %#x), want stale child2 skipped and unsummoned child3 unchanged", child2.flags, child3.flags)
	}
	want := []string{
		"flags:root", "class:root", "gate", "set-gate:00000000", "flags:root",
		"set:root:40000000", "report:root", "first:root", "class:child1", "status:child1",
		"flags:child1", "class:child1", "set:child1:00000000", "class:child1",
		"flags:child1", "pop:child1", "next:child1", "class:child3", "status:child3",
		"next:child3", "class:root",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events =\n%v\nwant\n%v", events, want)
	}
}

func TestUnitUnfreeze4E7A60DeadMonsterUsesClearedFlags(t *testing.T) {
	obj := &unitFreezeFixture4E79C0{name: "monster", flags: 0xabcd80f2, class: unitMonsterClass4E79C0, popByte: 0xee}
	gate := uint32(9)
	var events []string
	got := unitUnfreeze4E7A60(obj, 0, unitFreezeTestHooks4E79C0(&events, &gate))
	if got != 0xf0 || obj.flags != 0xabcd80f0 || gate != 9 {
		t.Fatalf("result/state = (%#x, %#x, %#x), want dead cleared monster", got, obj.flags, gate)
	}
	for _, event := range events {
		if event == "pop:monster" {
			t.Fatal("dead monster popped an action")
		}
	}
}

func TestUnitFreeze4E79C0NilFaultsOnFirstFlagsRead(t *testing.T) {
	gate := uint32(0)
	var events []string
	defer func() {
		if recover() == nil {
			t.Fatal("nil object did not fault")
		}
		if want := []string{"flags:"}; !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}()
	unitFreeze4E79C0[*unitFreezeFixture4E79C0](nil, 0, unitFreezeTestHooks4E79C0(&events, &gate))
}

func TestUnitUnfreeze4E7A60NilFaultsOnFirstFlagsRead(t *testing.T) {
	gate := uint32(0)
	var events []string
	defer func() {
		if recover() == nil {
			t.Fatal("nil object did not fault")
		}
		if want := []string{"flags:"}; !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}()
	unitUnfreeze4E7A60[*unitFreezeFixture4E79C0](nil, 0, unitFreezeTestHooks4E79C0(&events, &gate))
}
