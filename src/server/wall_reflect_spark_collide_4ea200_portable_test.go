package server

import "testing"

func TestWallReflectSparkCollidePortableContract4EA200(t *testing.T) {
	data := int32(-19)
	events := make([]uint32, 0, 4)
	wallReflectSparkCollide4EA200(1, 2, 0, wallReflectSparkCollideHooks4EA200[int, int, *int32]{
		loadCollideData: func(source int) *int32 {
			if source != 1 {
				t.Fatalf("data source = %d", source)
			}
			return &data
		},
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
			return 0x100
		},
		delayedDelete: func(source int) {
			events = append(events, 4)
			if source != 1 {
				t.Fatalf("delete source = %d", source)
			}
		},
	})
	want := []uint32{1, 2, 3, 4}
	if len(events) != len(want) {
		t.Fatalf("events = %v", events)
	}
	for i := range want {
		if events[i] != want[i] {
			t.Fatalf("events = %v", events)
		}
	}
}
