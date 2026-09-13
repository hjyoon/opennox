package server

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

const (
	monsterFightHighUnit515D30   = uint64(0x7faedbb66e70)
	monsterFightHighTarget515D30 = uint64(0x7faedba15ba0)
	monsterFightHighUpdate515D30 = uint64(0x7fae10004c00)
)

type monsterFightWorld515D30 struct {
	events    []string
	class     uint8
	flags     uint32
	update    uint64
	frame     uint32
	next      uint32
	pos       [2]uint32
	pushNil   map[uint32]bool
	args      map[uint64][4]uint32
	preferred uint64
}

func (w *monsterFightWorld515D30) hooks() monsterSetFightTargetHooks515D30[uint64, uint64, uint64] {
	return monsterSetFightTargetHooks515D30[uint64, uint64, uint64]{
		loadUpdate: func(unit uint64) uint64 {
			w.events = append(w.events, fmt.Sprintf("update:%x", unit))
			return w.update
		},
		loadClassLow: func(unit uint64) uint8 {
			w.events = append(w.events, fmt.Sprintf("class:%x", unit))
			return w.class
		},
		loadFlags: func(unit uint64) uint32 {
			w.events = append(w.events, fmt.Sprintf("flags:%x", unit))
			return w.flags
		},
		clearActionStack: func(unit uint64) {
			w.events = append(w.events, fmt.Sprintf("clear:%x", unit))
		},
		storeTarget: func(update, target uint64) {
			w.events = append(w.events, fmt.Sprintf("target:%x:%x", update, target))
			w.preferred = target
		},
		setNextFrame: func() {
			w.events = append(w.events, "next-frame")
			w.next = w.frame + 1
		},
		pushAction: func(unit uint64, action uint32) uint64 {
			w.events = append(w.events, fmt.Sprintf("push:%x:%d", unit, action))
			if action == uint32(ai.ACTION_FIGHT) {
				// Original target coordinates and frame are loaded after this call.
				w.pos = [2]uint32{0x7fa12345, 0x80000000}
				w.frame = 78
			}
			if w.pushNil[action] {
				return 0
			}
			return 0x7fae20000000 + uint64(action)*0x100
		},
		storeArgBits: func(item uint64, index int, bits uint32) {
			w.events = append(w.events, fmt.Sprintf("arg:%x:%d:%08x", item, index, bits))
			args := w.args[item]
			args[index] = bits
			w.args[item] = args
		},
		loadTargetPosXBits: func(target uint64) uint32 {
			w.events = append(w.events, fmt.Sprintf("x:%x", target))
			return w.pos[0]
		},
		loadTargetPosYBits: func(target uint64) uint32 {
			w.events = append(w.events, fmt.Sprintf("y:%x", target))
			return w.pos[1]
		},
		loadFrame: func() uint32 {
			w.events = append(w.events, "frame")
			return w.frame
		},
	}
}

func newMonsterFightWorld515D30() *monsterFightWorld515D30 {
	return &monsterFightWorld515D30{
		class:   2,
		update:  monsterFightHighUpdate515D30,
		frame:   41,
		pos:     [2]uint32{0x3f800000, 0x40000000},
		pushNil: make(map[uint32]bool),
		args:    make(map[uint64][4]uint32),
	}
}

func TestMonsterSetFightTarget515D30ExactOrderAndHighPointers(t *testing.T) {
	w := newMonsterFightWorld515D30()
	monsterSetFightTarget515D30(monsterFightHighUnit515D30, monsterFightHighTarget515D30, w.hooks())
	want := []string{
		"update:7faedbb66e70", "class:7faedbb66e70", "flags:7faedbb66e70",
		"clear:7faedbb66e70", "target:7fae10004c00:7faedba15ba0", "next-frame",
		"push:7faedbb66e70:32", "arg:7fae20002000:0:0000000f",
		"push:7faedbb66e70:15", "x:7faedba15ba0", "arg:7fae20000f00:0:7fa12345",
		"y:7faedba15ba0", "arg:7fae20000f00:1:80000000", "frame",
		"arg:7fae20000f00:2:0000004e",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
	if w.preferred != monsterFightHighTarget515D30 || w.next != 42 ||
		w.args[0x7fae20000f00] != [4]uint32{0x7fa12345, 0x80000000, 78} {
		t.Fatalf("preferred=%x next=%d fight args=%#v", w.preferred, w.next, w.args[0x7fae20000f00])
	}
}

func TestMonsterSetFightTarget515D30PushFailuresAndEligibility(t *testing.T) {
	w := newMonsterFightWorld515D30()
	w.pushNil[uint32(ai.ACTION_REPORT)] = true
	w.pushNil[uint32(ai.ACTION_FIGHT)] = true
	monsterSetFightTarget515D30(monsterFightHighUnit515D30, monsterFightHighTarget515D30, w.hooks())
	if w.preferred != monsterFightHighTarget515D30 || w.next != 42 ||
		len(w.args) != 0 || w.frame != 78 {
		t.Fatalf("failed pushes lost prior effects: preferred=%x next=%d frame=%d args=%v", w.preferred, w.next, w.frame, w.args)
	}

	for _, tc := range []struct {
		name       string
		unit, targ uint64
		class      uint8
		flags      uint32
		update     uint64
	}{
		{"nil unit", 0, monsterFightHighTarget515D30, 2, 0, monsterFightHighUpdate515D30},
		{"nil target", monsterFightHighUnit515D30, 0, 2, 0, monsterFightHighUpdate515D30},
		{"self target", monsterFightHighUnit515D30, monsterFightHighUnit515D30, 2, 0, monsterFightHighUpdate515D30},
		{"not monster", monsterFightHighUnit515D30, monsterFightHighTarget515D30, 1, 0, monsterFightHighUpdate515D30},
		{"dead", monsterFightHighUnit515D30, monsterFightHighTarget515D30, 2, 0x8000, monsterFightHighUpdate515D30},
		{"no update", monsterFightHighUnit515D30, monsterFightHighTarget515D30, 2, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newMonsterFightWorld515D30()
			w.class, w.flags, w.update = tc.class, tc.flags, tc.update
			monsterSetFightTarget515D30(tc.unit, tc.targ, w.hooks())
			if w.preferred != 0 || w.next != 0 || len(w.args) != 0 {
				t.Fatalf("invalid request mutated state: %#v", w)
			}
		})
	}
}

func TestMonsterSetFightTargetNative515D30PreservesObjectAndPosition(t *testing.T) {
	// The standalone server package does not load the legacy C data blobs.
	if memmap.BlobByAddr(0x5D4594) == nil {
		memmap.RegisterBlobData(0x5D4594, "fight_target_test", make([]byte, 2487688))
	}
	s, unit, update := monsterScriptHitFixture515A30(t)
	s.SetFrame(0x1234567)
	target := &Object{PosVec: types.Ptf(math.Float32frombits(0x7fa12345), math.Float32frombits(0x80000000))}
	if unsafe.Sizeof(uintptr(0)) == 8 &&
		(uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 || uintptr(unsafe.Pointer(target)) <= math.MaxUint32) {
		t.Fatalf("object pointers are not both above 4 GiB: %p %p", unit, target)
	}
	next := memmap.PtrUint32(0x5D4594, 2487684)
	oldNext := *next
	t.Cleanup(func() { *next = oldNext })
	s.MonsterSetFightTarget515D30(unit, target)
	if update.PreferredEnemy != target || *next != s.Frame()+1 || update.AIStackInd != 1 ||
		update.AIStack[0].Type() != ai.ACTION_REPORT || update.AIStack[0].ArgU32(0) != uint32(ai.ACTION_FIGHT) ||
		update.AIStack[1].Type() != ai.ACTION_FIGHT || update.AIStack[1].ArgU32(0) != 0x7fa12345 ||
		update.AIStack[1].ArgU32(1) != 0x80000000 || update.AIStack[1].ArgU32(2) != s.Frame() ||
		!s.AI.StackChanged {
		t.Fatalf("fight state = preferred %p next %d stack %#v changed %t", update.PreferredEnemy, *next, update.AIStack[:2], s.AI.StackChanged)
	}
	runtime.KeepAlive(target)
	runtime.KeepAlive(unit)

	// Rejected requests must preserve both the scheduled fight and frame gate.
	unit.ObjFlags = object.FlagDead
	s.MonsterSetFightTarget515D30(unit, &Object{})
	if update.PreferredEnemy != target || *next != s.Frame()+1 || update.AIStackInd != 1 {
		t.Fatalf("dead unit changed fight state: preferred %p next %d index %d", update.PreferredEnemy, *next, update.AIStackInd)
	}
}
