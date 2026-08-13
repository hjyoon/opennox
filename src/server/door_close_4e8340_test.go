package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type doorCloseTestObject4E8340 struct {
	class  uint32
	update *doorCloseTestUpdate4E8340
}

type doorCloseTestUpdate4E8340 struct {
	lockCode uint8
	tileX    int32
	tileY    int32
	syncByte uint8
}

type doorCloseTestPoint4E8340 struct {
	x int32
	y int32
}

func doorCloseTestHooks4E8340(
	events *[]string,
	quest int32,
	syncResult int32,
) doorCloseHooks4E8340[*doorCloseTestObject4E8340, *doorCloseTestUpdate4E8340, *doorCloseTestPoint4E8340] {
	add := func(event string) { *events = append(*events, event) }
	return doorCloseHooks4E8340[*doorCloseTestObject4E8340, *doorCloseTestUpdate4E8340, *doorCloseTestPoint4E8340]{
		class: func(obj *doorCloseTestObject4E8340) uint32 {
			add("class")
			return obj.class
		},
		updateData: func(obj *doorCloseTestObject4E8340) *doorCloseTestUpdate4E8340 {
			add("update")
			return obj.update
		},
		tileX: func(update *doorCloseTestUpdate4E8340) int32 {
			add("tile-x")
			return update.tileX
		},
		targetX: func(target *doorCloseTestPoint4E8340) int32 {
			add("target-x")
			return target.x
		},
		tileY: func(update *doorCloseTestUpdate4E8340) int32 {
			add("tile-y")
			return update.tileY
		},
		targetY: func(target *doorCloseTestPoint4E8340) int32 {
			add("target-y")
			return target.y
		},
		storeLockCode: func(update *doorCloseTestUpdate4E8340, value uint8) {
			add(fmt.Sprintf("lock=%d", value))
			update.lockCode = value
		},
		quest: func() int32 {
			add("quest")
			return quest
		},
		questSync: func(obj *doorCloseTestObject4E8340) int32 {
			add("sync")
			obj.update.syncByte = 1
			return syncResult
		},
	}
}

func TestDoorClose4E8340ClassLowByteShortCircuit(t *testing.T) {
	for _, class := range []uint32{0, 0x00008000, 0xffffff7f} {
		events := []string{}
		update := &doorCloseTestUpdate4E8340{lockCode: 4, tileX: 10, tileY: 20}
		doorClose4E8340(
			&doorCloseTestObject4E8340{class: class, update: update},
			&doorCloseTestPoint4E8340{x: 10, y: 20},
			doorCloseTestHooks4E8340(&events, 1, -1),
		)
		if want := []string{"class"}; !reflect.DeepEqual(events, want) {
			t.Fatalf("class %#x events = %v, want %v", class, events, want)
		}
		if update.lockCode != 4 || update.syncByte != 0 {
			t.Fatalf("class %#x changed update: %#v", class, update)
		}
	}
}

func TestDoorClose4E8340CoordinateReadAndShortCircuitOrder(t *testing.T) {
	tests := []struct {
		name   string
		target doorCloseTestPoint4E8340
		want   []string
	}{
		{
			name:   "x mismatch skips both y reads",
			target: doorCloseTestPoint4E8340{x: 11, y: 20},
			want:   []string{"class", "update", "tile-x", "target-x"},
		},
		{
			name:   "y mismatch stops before lock and quest",
			target: doorCloseTestPoint4E8340{x: 10, y: 21},
			want:   []string{"class", "update", "tile-x", "target-x", "tile-y", "target-y"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			events := []string{}
			update := &doorCloseTestUpdate4E8340{lockCode: 4, tileX: 10, tileY: 20}
			doorClose4E8340(
				&doorCloseTestObject4E8340{class: 0xa5000080, update: update},
				&tc.target,
				doorCloseTestHooks4E8340(&events, 1, math.MinInt32),
			)
			if !reflect.DeepEqual(events, tc.want) {
				t.Fatalf("events = %v, want %v", events, tc.want)
			}
			if update.lockCode != 4 || update.syncByte != 0 {
				t.Fatalf("mismatch changed update: %#v", update)
			}
		})
	}
}

func TestDoorClose4E8340ClearsBeforeQuestAndTreatsAnyNonzeroAsQuest(t *testing.T) {
	for _, quest := range []int32{0, 1, 2, -1, math.MinInt32} {
		events := []string{}
		update := &doorCloseTestUpdate4E8340{lockCode: 0xff, tileX: -17, tileY: math.MaxInt32}
		doorClose4E8340(
			&doorCloseTestObject4E8340{class: 0xffffff80, update: update},
			&doorCloseTestPoint4E8340{x: -17, y: math.MaxInt32},
			doorCloseTestHooks4E8340(&events, quest, math.MinInt32),
		)
		want := []string{"class", "update", "tile-x", "target-x", "tile-y", "target-y", "lock=0", "quest"}
		if quest != 0 {
			want = append(want, "sync")
		}
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("quest %d events = %v, want %v", quest, events, want)
		}
		if update.lockCode != 0 {
			t.Fatalf("quest %d lock = %d, want 0", quest, update.lockCode)
		}
		wantSync := uint8(0)
		if quest != 0 {
			wantSync = 1
		}
		if update.syncByte != wantSync {
			t.Fatalf("quest %d sync = %d, want %d", quest, update.syncByte, wantSync)
		}
	}
}

func TestDoorClose4E8340NilUpdateFaultsBeforeTargetRead(t *testing.T) {
	events := []string{}
	defer func() {
		if recover() == nil {
			t.Fatal("nil update returned without a panic")
		}
		want := []string{"class", "update", "tile-x"}
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}()
	doorClose4E8340(
		&doorCloseTestObject4E8340{class: 0x80},
		(*doorCloseTestPoint4E8340)(nil),
		doorCloseTestHooks4E8340(&events, 1, 0),
	)
}

func TestDoorQuestSync4E8390StoresBeforeSendAndReturnsExactResult(t *testing.T) {
	for _, result := range []int32{math.MinInt32, -1, 0, 1, math.MaxInt32} {
		events := []string{}
		update := &doorCloseTestUpdate4E8340{syncByte: 0xa5}
		obj := &doorCloseTestObject4E8340{update: update}
		got := doorQuestSync4E8390(obj, doorQuestSyncHooks4E8390[*doorCloseTestObject4E8340, *doorCloseTestUpdate4E8340]{
			updateData: func(got *doorCloseTestObject4E8340) *doorCloseTestUpdate4E8340 {
				events = append(events, "update")
				if got != obj {
					t.Fatal("update received a different object")
				}
				return got.update
			},
			storeSyncByte: func(got *doorCloseTestUpdate4E8340, value uint8) {
				events = append(events, fmt.Sprintf("sync=%d", value))
				got.syncByte = value
			},
			sendExtent: func(recipient int32, got *doorCloseTestObject4E8340) int32 {
				events = append(events, "send")
				if recipient != 255 || got != obj || got.update.syncByte != 1 {
					t.Fatalf("send args/state = (%d, %p, %d)", recipient, got, got.update.syncByte)
				}
				return result
			},
		})
		if got != result {
			t.Fatalf("result = %d, want %d", got, result)
		}
		if want := []string{"update", "sync=1", "send"}; !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}
}

func TestDoorExtentPacket4D6A20PacketAndOptions(t *testing.T) {
	events := []string{}
	obj := &doorCloseTestObject4E8340{}
	got := doorExtentPacket4D6A20(math.MinInt32, obj, doorExtentPacketHooks4D6A20[*doorCloseTestObject4E8340]{
		extent: func(got *doorCloseTestObject4E8340) uint16 {
			events = append(events, "extent")
			if got != obj {
				t.Fatal("extent received a different object")
			}
			return 0xabcd
		},
		send: func(recipient int32, packet [4]byte, relatedObject uintptr, removeIfDisconnected int32) int32 {
			events = append(events, "send")
			if recipient != math.MinInt32 || packet != [4]byte{0xf0, 0x0f, 0xcd, 0xab} ||
				relatedObject != 0 || removeIfDisconnected != 1 {
				t.Fatalf("send = (%d, % x, %#x, %d)", recipient, packet, relatedObject, removeIfDisconnected)
			}
			return math.MaxInt32
		},
	})
	if got != math.MaxInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MaxInt32))
	}
	if want := []string{"extent", "send"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}
