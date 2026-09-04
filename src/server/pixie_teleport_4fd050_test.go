package server

import (
	"fmt"
	"reflect"
	"testing"
)

func TestPixieTeleport4FD050ExactTraceAndRawCoordinateBits(t *testing.T) {
	const (
		owner = uint64(0x1_0000_1234)
		pixie = uint64(0x2_0000_5678)
	)
	xValues := []uint32{0x80000000, 0x7fa12345, 0xff800000}
	yValues := []uint32{0x00000000, 0x7fcabcde, 0x7f800000}
	var (
		events []string
		got    [6]uint32
	)
	recordStore := func(name string, index int) func(uint64, uint32) {
		return func(object uint64, value uint32) {
			events = append(events, name)
			if object != pixie {
				t.Fatalf("%s object = %#x, want %#x", name, object, pixie)
			}
			got[index] = value
		}
	}
	pixieTeleport4FD050(pixieTeleportHooks4FD050[uint64]{
		loadOwnerArg: func() uint64 {
			events = append(events, "owner-arg")
			return owner
		},
		loadPixieArg: func() uint64 {
			events = append(events, "pixie-arg")
			return pixie
		},
		loadOwnerXBits: func(object uint64) uint32 {
			events = append(events, "owner-x")
			if object != owner {
				t.Fatalf("X owner = %#x, want %#x", object, owner)
			}
			value := xValues[0]
			xValues = xValues[1:]
			return value
		},
		loadOwnerYBits: func(object uint64) uint32 {
			events = append(events, "owner-y")
			if object != owner {
				t.Fatalf("Y owner = %#x, want %#x", object, owner)
			}
			value := yValues[0]
			yValues = yValues[1:]
			return value
		},
		storeNewPosXBits:  recordStore("new-x", 0),
		storeNewPosYBits:  recordStore("new-y", 1),
		storePosXBits:     recordStore("pos-x", 2),
		storePosYBits:     recordStore("pos-y", 3),
		storePrevPosXBits: recordStore("prev-x", 4),
		storePrevPosYBits: recordStore("prev-y", 5),
		moveUpdate: func(object uint64) {
			events = append(events, "move")
			if object != pixie {
				t.Fatalf("move object = %#x, want %#x", object, pixie)
			}
			want := [6]uint32{
				0x80000000, 0x00000000,
				0x7fa12345, 0x7fcabcde,
				0xff800000, 0x7f800000,
			}
			if got != want {
				t.Fatalf("coordinates at move = %#v, want %#v", got, want)
			}
		},
	})

	wantEvents := []string{
		"owner-arg", "pixie-arg",
		"owner-x", "new-x", "owner-y", "new-y",
		"owner-x", "pos-x", "owner-y", "pos-y",
		"owner-x", "prev-x", "owner-y", "prev-y",
		"move",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if len(xValues) != 0 || len(yValues) != 0 {
		t.Fatalf("unconsumed coordinate loads = X:%d Y:%d", len(xValues), len(yValues))
	}
}

func TestPixieTeleport4FD050AllFaultPrefixes(t *testing.T) {
	want := []string{
		"owner-arg", "pixie-arg",
		"owner-x", "new-x", "owner-y", "new-y",
		"owner-x", "pos-x", "owner-y", "pos-y",
		"owner-x", "prev-x", "owner-y", "prev-y",
		"move",
	}
	for failAt := 1; failAt <= len(want); failAt++ {
		t.Run(fmt.Sprintf("step_%02d", failAt), func(t *testing.T) {
			var events []string
			emit := func(event string) {
				events = append(events, event)
				if len(events) == failAt {
					panic(event)
				}
			}
			hooks := pixieTeleportHooks4FD050[int]{
				loadOwnerArg: func() int {
					emit("owner-arg")
					return 1
				},
				loadPixieArg: func() int {
					emit("pixie-arg")
					return 2
				},
				loadOwnerXBits: func(int) uint32 {
					emit("owner-x")
					return 3
				},
				loadOwnerYBits: func(int) uint32 {
					emit("owner-y")
					return 4
				},
				storeNewPosXBits: func(int, uint32) { emit("new-x") },
				storeNewPosYBits: func(int, uint32) { emit("new-y") },
				storePosXBits:    func(int, uint32) { emit("pos-x") },
				storePosYBits:    func(int, uint32) { emit("pos-y") },
				storePrevPosXBits: func(int, uint32) {
					emit("prev-x")
				},
				storePrevPosYBits: func(int, uint32) {
					emit("prev-y")
				},
				moveUpdate: func(int) { emit("move") },
			}
			func() {
				defer func() {
					if recover() == nil {
						t.Fatalf("step %d did not panic", failAt)
					}
				}()
				pixieTeleport4FD050(hooks)
			}()
			if expected := want[:failAt]; !reflect.DeepEqual(events, expected) {
				t.Fatalf("events = %v, want prefix %v", events, expected)
			}
		})
	}
}
