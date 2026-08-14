package server

import "testing"

func TestWebbingCollidePortableContract4EA380(t *testing.T) {
	events := make([]uint32, 0, 9)
	class := uint8(2)
	webbingCollide4EA380(1, 2, 3, webbingCollideHooks4EA380[int]{
		audio: func(id uint32, source int) {
			events = append(events, 1)
			if id != 351 || source != 1 {
				t.Fatalf("audio = %d/%d", id, source)
			}
		},
		delayedDelete: func(source int) {
			events = append(events, 2)
		},
		findParent: func(source int) int {
			events = append(events, 3)
			return 4
		},
		targetDamage: func(target, parent, source int, damage int32, damageType uint32) int32 {
			events = append(events, 4)
			if target != 2 || parent != 4 || source != 1 || damage != 0 || damageType != 2 {
				t.Fatalf("damage = %d/%d/%d/%d/%d", target, parent, source, damage, damageType)
			}
			return 1
		},
		loadClassLow: func(target int) uint8 {
			events = append(events, 5)
			return class
		},
		loadFPS: func() uint32 {
			events = append(events, 6)
			return 0x40004001
		},
		applyEnchant: func(target int, enchant, duration, power uint32) {
			events = append(events, 7)
			if target != 2 || enchant != 4 || duration != 0x10004 || power != 3 {
				t.Fatalf("enchant = %d/%d/%#x/%d", target, enchant, duration, power)
			}
			class = 4
		},
		priorityMessage: func(target int, message string) {
			events = append(events, 8)
			if target != 2 || message != "objcoll.c:WebbingSlow" {
				t.Fatalf("message = %d/%q", target, message)
			}
		},
	})
	want := []uint32{1, 2, 3, 4, 5, 6, 7, 5, 8}
	if len(events) != len(want) {
		t.Fatalf("events = %v", events)
	}
	for i := range want {
		if events[i] != want[i] {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}
}
