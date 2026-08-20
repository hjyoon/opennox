package server

import (
	"fmt"
	"reflect"
	"testing"
)

type respawnWeaponFlagsWorld4EF580 struct {
	armor  map[uint32]int
	weapon map[uint32]int
	allow  map[int]bool
	events []string
	fault  int
	after  map[string]func()
}

func (w *respawnWeaponFlagsWorld4EF580) record(event string) {
	w.events = append(w.events, event)
	if after := w.after[event]; after != nil {
		delete(w.after, event)
		after()
	}
	if w.fault != 0 && len(w.events) == w.fault {
		panic(event)
	}
}

func (w *respawnWeaponFlagsWorld4EF580) hooks() respawnWeaponFlagsHooks4EF580 {
	return respawnWeaponFlagsHooks4EF580{
		lookupArmor: func(mask uint32) int {
			result := w.armor[mask]
			w.record(fmt.Sprintf("armor:%08x=%d", mask, result))
			return result
		},
		lookupWeapon: func(mask uint32) int {
			result := w.weapon[mask]
			w.record(fmt.Sprintf("weapon:%08x=%d", mask, result))
			return result
		},
		allowed: func(ind int) bool {
			result := w.allow[ind]
			w.record(fmt.Sprintf("allowed:%d=%t", ind, result))
			return result
		},
	}
}

func newRespawnWeaponFlagsWorld4EF580(allowed uint8) *respawnWeaponFlagsWorld4EF580 {
	w := &respawnWeaponFlagsWorld4EF580{
		armor: map[uint32]int{
			0x400:     11,
			0x4:       22,
			0x1:       33,
			0x4000:    55,
			0x1000000: 88,
		},
		weapon: map[uint32]int{
			0x8000: 44,
			0x100:  66,
			0x200:  77,
		},
		allow: make(map[int]bool),
		after: make(map[string]func()),
	}
	for bit, ind := range []int{11, 22, 33, 44, 55, 66, 77, 88} {
		w.allow[ind] = allowed&(1<<bit) != 0
	}
	return w
}

func TestRespawnWeaponFlags4EF580ExactOrderAndBitMapping(t *testing.T) {
	w := newRespawnWeaponFlagsWorld4EF580(0xa5)
	got := respawnWeaponFlags4EF580(w.hooks())
	if got != 0xa5 {
		t.Fatalf("flags = %#02x, want 0xa5", got)
	}
	want := []string{
		"armor:00000400=11", "allowed:11=true",
		"armor:00000004=22", "allowed:22=false",
		"armor:00000001=33", "allowed:33=true",
		"weapon:00008000=44", "allowed:44=false",
		"armor:00004000=55", "allowed:55=false",
		"weapon:00000100=66", "allowed:66=true",
		"weapon:00000200=77", "allowed:77=false",
		"armor:01000000=88", "allowed:88=true",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestRespawnWeaponFlags4EF580ExhaustiveResultsAndNoShortCircuit(t *testing.T) {
	for want := 0; want <= 0xff; want++ {
		w := newRespawnWeaponFlagsWorld4EF580(uint8(want))
		got := respawnWeaponFlags4EF580(w.hooks())
		if got != uint8(want) {
			t.Fatalf("allowed %#02x: flags = %#02x", want, got)
		}
		if len(w.events) != 16 {
			t.Fatalf("allowed %#02x: events = %d, want 16", want, len(w.events))
		}
	}
}

func TestRespawnWeaponFlags4EF580CallbacksObserveLiveState(t *testing.T) {
	w := newRespawnWeaponFlagsWorld4EF580(0)
	w.after["allowed:11=false"] = func() {
		w.armor[0x4] = 99
		w.allow[99] = true
	}
	w.after["allowed:99=true"] = func() {
		w.weapon[0x8000] = 100
		w.allow[100] = true
	}
	got := respawnWeaponFlags4EF580(w.hooks())
	if got != 0x0a {
		t.Fatalf("flags = %#02x, want 0x0a", got)
	}
	if w.events[2] != "armor:00000004=99" || w.events[6] != "weapon:00008000=100" {
		t.Fatalf("live lookup events = %v", w.events)
	}
}

func TestRespawnWeaponFlags4EF580EveryCallbackFaultPrefix(t *testing.T) {
	base := newRespawnWeaponFlagsWorld4EF580(0xff)
	respawnWeaponFlags4EF580(base.hooks())
	want := append([]string(nil), base.events...)

	for fault := 1; fault <= len(want); fault++ {
		t.Run(fmt.Sprintf("event-%02d", fault), func(t *testing.T) {
			w := newRespawnWeaponFlagsWorld4EF580(0xff)
			w.fault = fault
			defer func() {
				if got := recover(); got != want[fault-1] {
					t.Fatalf("panic = %v, want %q", got, want[fault-1])
				}
				if prefix := want[:fault]; !reflect.DeepEqual(w.events, prefix) {
					t.Fatalf("events = %v, want %v", w.events, prefix)
				}
			}()
			respawnWeaponFlags4EF580(w.hooks())
		})
	}
}
