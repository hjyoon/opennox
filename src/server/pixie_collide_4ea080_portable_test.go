package server

import "testing"

func TestPixieCollidePortableContract4EA080(t *testing.T) {
	data := int32(-19)
	events := make([]uint32, 0, 4)
	pixieCollide4EA080(1, 2, 0, pixieCollideHooks4EA080[int, int, *int32]{
		loadCollideData: func(source int) *int32 {
			if source != 1 {
				t.Fatalf("data source = %d", source)
			}
			return &data
		},
		isEnemy:   func(source, target int) int32 { return 1 },
		loadClass: func(int) uint32 { return 2 },
		loadFlags: func(int) uint32 { return 0 },
		loadOwner: func(int) int { return 0 },
		loadDamage: func(got *int32) int32 {
			events = append(events, 1)
			return *got
		},
		findParent: func(source int) int {
			events = append(events, 2)
			return 3
		},
		targetDamage: func(target, parent, source int, damage int32, damageType uint32) int32 {
			events = append(events, 3)
			if target != 2 || parent != 3 || source != 1 || damage != -19 || damageType != 11 {
				t.Fatalf("damage = %d/%d/%d/%d/%d", target, parent, source, damage, damageType)
			}
			return -1
		},
		audio: func(id uint32, obj int) {
			events = append(events, 4)
			if id != 96 || obj != 1 {
				t.Fatalf("audio = %d/%d", id, obj)
			}
		},
		delayedDelete: func(source int) {
			events = append(events, 5)
			if source != 1 {
				t.Fatalf("delete = %d", source)
			}
		},
	})
	want := []uint32{1, 2, 3, 4, 5}
	if len(events) != len(want) {
		t.Fatalf("events = %v", events)
	}
	for i := range want {
		if events[i] != want[i] {
			t.Fatalf("events = %v", events)
		}
	}
}
