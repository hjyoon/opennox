package opennox

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/opennox/opennox/v1/common/sound"
)

const restoreHighTarget52BF20 = uint64(0x7fb8ed1723d0)

type restoreTestArg52BF20 struct {
	target uint64
}

func TestRestoreHealth52BF20PreservesNativeWidthAndOrder(t *testing.T) {
	arg := &restoreTestArg52BF20{target: restoreHighTarget52BF20}
	var events []string
	got := restoreHealth52BF20(arg, restoreHealthHooks52BF20[uint64, *restoreTestArg52BF20]{
		loadTarget: func(arg *restoreTestArg52BF20) uint64 {
			events = append(events, fmt.Sprintf("target:%#x", arg.target))
			return arg.target
		},
		setMaxHP: func(target uint64) {
			events = append(events, fmt.Sprintf("max-hp:%#x", target))
		},
		audio: func(id sound.ID, target uint64) {
			events = append(events, fmt.Sprintf("audio:%d:%#x", id, target))
		},
	})
	if got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{
		"target:0x7fb8ed1723d0",
		"max-hp:0x7fb8ed1723d0",
		"audio:754:0x7fb8ed1723d0",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestRestoreHealth52BF20NilTargetFailsWithoutEffects(t *testing.T) {
	var effects int
	got := restoreHealth52BF20(&restoreTestArg52BF20{}, restoreHealthHooks52BF20[uint64, *restoreTestArg52BF20]{
		loadTarget: func(arg *restoreTestArg52BF20) uint64 { return arg.target },
		setMaxHP:   func(uint64) { effects++ },
		audio:      func(sound.ID, uint64) { effects++ },
	})
	if got != 0 || effects != 0 {
		t.Fatalf("result/effects = %d/%d, want 0/0", got, effects)
	}
}

func TestRestoreMana52BF50PlayerPreservesNativeWidthAndOrder(t *testing.T) {
	arg := &restoreTestArg52BF20{target: restoreHighTarget52BF20}
	var events []string
	got := restoreMana52BF50(arg, restoreManaHooks52BF50[uint64, *restoreTestArg52BF20]{
		loadTarget: func(arg *restoreTestArg52BF20) uint64 {
			events = append(events, fmt.Sprintf("target:%#x", arg.target))
			return arg.target
		},
		loadClassLow: func(target uint64) uint8 {
			events = append(events, fmt.Sprintf("class:%#x", target))
			return 0x04
		},
		loadMaxMana: func(target uint64) uint16 {
			events = append(events, fmt.Sprintf("max-mana:%#x", target))
			return 0xfedc
		},
		addMana: func(target uint64, amount uint16) {
			events = append(events, fmt.Sprintf("add:%#x:%#x", target, amount))
		},
		audio: func(id sound.ID, target uint64) {
			events = append(events, fmt.Sprintf("audio:%d:%#x", id, target))
		},
	})
	if got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{
		"target:0x7fb8ed1723d0",
		"class:0x7fb8ed1723d0",
		"max-mana:0x7fb8ed1723d0",
		"add:0x7fb8ed1723d0:0xfedc",
		"audio:755:0x7fb8ed1723d0",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestRestoreMana52BF50NonPlayerStillSucceedsWithoutEffects(t *testing.T) {
	arg := &restoreTestArg52BF20{target: restoreHighTarget52BF20}
	var events []string
	got := restoreMana52BF50(arg, restoreManaHooks52BF50[uint64, *restoreTestArg52BF20]{
		loadTarget: func(arg *restoreTestArg52BF20) uint64 {
			events = append(events, "target")
			return arg.target
		},
		loadClassLow: func(uint64) uint8 {
			events = append(events, "class")
			return 0
		},
		loadMaxMana: func(uint64) uint16 {
			events = append(events, "max-mana")
			return 1
		},
		addMana: func(uint64, uint16) { events = append(events, "add") },
		audio:   func(sound.ID, uint64) { events = append(events, "audio") },
	})
	if got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{"target", "class"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestRestoreMana52BF50NilTargetFailsWithoutEffects(t *testing.T) {
	var effects int
	got := restoreMana52BF50(&restoreTestArg52BF20{}, restoreManaHooks52BF50[uint64, *restoreTestArg52BF20]{
		loadTarget:   func(arg *restoreTestArg52BF20) uint64 { return arg.target },
		loadClassLow: func(uint64) uint8 { effects++; return 0x04 },
		loadMaxMana:  func(uint64) uint16 { effects++; return 1 },
		addMana:      func(uint64, uint16) { effects++ },
		audio:        func(sound.ID, uint64) { effects++ },
	})
	if got != 0 || effects != 0 {
		t.Fatalf("result/effects = %d/%d, want 0/0", got, effects)
	}
}
