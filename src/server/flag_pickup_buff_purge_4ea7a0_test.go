package server

import (
	"fmt"
	"reflect"
	"testing"
)

func defaultFlagPickupBuffPurgeHooks4EA7A0() flagPickupBuffPurgeHooks4EA7A0[int] {
	return flagPickupBuffPurgeHooks4EA7A0[int]{
		hasBuff:       func(int, uint32) int32 { return 0 },
		enchantSpell:  func(uint32) int32 { return 0 },
		spellHasFlags: func(int32, uint32) int32 { return 0 },
		buffOff:       func(int, uint32) int32 { return 0 },
	}
}

func TestFlagPickupBuffPurge4EA7A0ScansAllSlots(t *testing.T) {
	var slots []uint32
	hooks := defaultFlagPickupBuffPurgeHooks4EA7A0()
	hooks.hasBuff = func(obj int, enchant uint32) int32 {
		if obj != 17 {
			t.Fatalf("object = %d", obj)
		}
		slots = append(slots, enchant)
		return 0
	}
	got := flagPickupBuffPurge4EA7A0(17, hooks)
	if got != 0 {
		t.Fatalf("result = %#x", got)
	}
	if len(slots) != 32 {
		t.Fatalf("slots = %v", slots)
	}
	for i, slot := range slots {
		if slot != uint32(i) {
			t.Fatalf("slots[%d] = %d", i, slot)
		}
	}
}

func TestFlagPickupBuffPurge4EA7A0PreservesDoubleLookupAndOrder(t *testing.T) {
	events := make([]string, 0, 40)
	lookups := 0
	hooks := defaultFlagPickupBuffPurgeHooks4EA7A0()
	hooks.hasBuff = func(_ int, enchant uint32) int32 {
		events = append(events, fmt.Sprintf("has:%d", enchant))
		if enchant == 7 {
			return -1
		}
		return 0
	}
	hooks.enchantSpell = func(enchant uint32) int32 {
		lookups++
		events = append(events, fmt.Sprintf("spell:%d:%d", enchant, lookups))
		if lookups == 1 {
			return 0x1111
		}
		return 0x2222
	}
	hooks.spellHasFlags = func(ind int32, flags uint32) int32 {
		events = append(events, fmt.Sprintf("flags:%x:%x", ind, flags))
		if ind != 0x2222 || flags != 0x80000 {
			t.Fatalf("HasFlags args = %#x/%#x", ind, flags)
		}
		return 1
	}
	hooks.buffOff = func(obj int, enchant uint32) int32 {
		events = append(events, fmt.Sprintf("off:%d:%d", obj, enchant))
		if obj != 9 || enchant != 7 {
			t.Fatalf("BuffOff args = %d/%d", obj, enchant)
		}
		return 0x3333
	}
	got := flagPickupBuffPurge4EA7A0(9, hooks)
	if got != 0 {
		t.Fatalf("slot 31 result = %#x, want 0", got)
	}
	wantAtSeven := []string{
		"has:7", "spell:7:1", "spell:7:2", "flags:2222:80000", "off:9:7",
	}
	if !reflect.DeepEqual(events[7:12], wantAtSeven) {
		t.Fatalf("slot 7 events = %v, want %v", events[7:12], wantAtSeven)
	}
	if len(events) != 36 {
		t.Fatalf("event count = %d, events = %v", len(events), events)
	}
}

func TestFlagPickupBuffPurge4EA7A0ReturnsLastOperationEAX(t *testing.T) {
	hooks := defaultFlagPickupBuffPurgeHooks4EA7A0()
	hooks.hasBuff = func(_ int, enchant uint32) int32 {
		if enchant == 31 {
			return 1
		}
		return 0
	}
	hooks.enchantSpell = func(uint32) int32 { return 5 }
	hooks.spellHasFlags = func(int32, uint32) int32 { return -1 }
	hooks.buffOff = func(int, uint32) int32 { return -0x7654321 }
	if got := flagPickupBuffPurge4EA7A0(0, hooks); got != -0x7654321 {
		t.Fatalf("result = %#x", got)
	}
}

func TestFlagPickupBuffPurge4EA7A0FirstLookupZeroSkipsSecond(t *testing.T) {
	hooks := defaultFlagPickupBuffPurgeHooks4EA7A0()
	hooks.hasBuff = func(_ int, enchant uint32) int32 {
		if enchant == 31 {
			return 1
		}
		return 0
	}
	lookups := 0
	hooks.enchantSpell = func(uint32) int32 {
		lookups++
		return 0
	}
	hooks.spellHasFlags = func(int32, uint32) int32 {
		t.Fatal("zero first lookup reached HasFlags")
		return 0
	}
	if got := flagPickupBuffPurge4EA7A0(0, hooks); got != 0 || lookups != 1 {
		t.Fatalf("result/lookups = %d/%d", got, lookups)
	}
}
