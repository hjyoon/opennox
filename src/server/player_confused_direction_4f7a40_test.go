package server

import (
	"reflect"
	"testing"
)

type playerConfusedDirectionTestObject4F7A40 struct {
	direction uint16
	power     uint8
	netCode   uint32
}

func playerConfusedDirectionReference4F7A40(
	direction uint16,
	power uint8,
	frame, netCode uint32,
) uint32 {
	phase := (frame + netCode) % 40
	if phase > 20 {
		phase = 40 - phase
	}
	value := int32(int16(direction)) + (int32(power)+3)*(int32(phase)-10)
	return uint32((value%256 + 256) % 256)
}

func TestPlayerConfusedDirection4F7A40LoadOrderAndLiveFields(t *testing.T) {
	obj := &playerConfusedDirectionTestObject4F7A40{
		direction: 0xffff,
		power:     2,
		netCode:   7,
	}
	frame := uint32(12)
	var events []string
	got := playerConfusedDirection4F7A40(obj,
		playerConfusedDirectionHooks4F7A40[*playerConfusedDirectionTestObject4F7A40]{
			loadDirection2: func(obj *playerConfusedDirectionTestObject4F7A40) uint16 {
				events = append(events, "direction")
				return obj.direction
			},
			loadBuffPower: func(obj *playerConfusedDirectionTestObject4F7A40, buff uint32) uint8 {
				events = append(events, "power")
				if buff != 3 {
					t.Fatalf("buff = %d, want 3", buff)
				}
				obj.direction = 1000
				frame = ^uint32(0)
				obj.netCode = 8
				return obj.power
			},
			loadFrame: func() uint32 {
				events = append(events, "frame")
				obj.netCode = 1
				return frame
			},
			loadNetCode: func(obj *playerConfusedDirectionTestObject4F7A40) uint32 {
				events = append(events, "netcode")
				return obj.netCode
			},
		},
	)
	if want := uint32(205); got != want {
		t.Fatalf("direction = %d, want %d", got, want)
	}
	if want := []string{"direction", "power", "frame", "netcode"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestPlayerConfusedDirection4F7A40FaultPrefix(t *testing.T) {
	var events []string
	defer func() {
		if recover() == nil {
			t.Fatal("Direction2 fault was not propagated")
		}
		if want := []string{"direction"}; !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %#v, want %#v", events, want)
		}
	}()
	playerConfusedDirection4F7A40(0, playerConfusedDirectionHooks4F7A40[int]{
		loadDirection2: func(int) uint16 {
			events = append(events, "direction")
			panic("fault")
		},
		loadBuffPower: func(int, uint32) uint8 {
			events = append(events, "power")
			return 0
		},
		loadFrame: func() uint32 {
			events = append(events, "frame")
			return 0
		},
		loadNetCode: func(int) uint32 {
			events = append(events, "netcode")
			return 0
		},
	})
}

func TestPlayerConfusedDirection4F7A40BoundariesAndNormalization(t *testing.T) {
	tests := []struct {
		name      string
		direction uint16
		power     uint8
		frame     uint32
		netCode   uint32
	}{
		{"phase zero negative", 0, 0, 0, 0},
		{"phase ten center", 0x8000, 255, 10, 0},
		{"phase twenty", 0x7fff, 255, 20, 0},
		{"phase twenty one fold", 255, 4, 21, 0},
		{"phase thirty nine fold", 0xffff, 2, 39, 0},
		{"unsigned sum wraps", 0xffff, 2, ^uint32(0), 1},
		{"large net code", 0x8000, 255, 0xffffff00, 0x80000123},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			obj := playerConfusedDirectionTestObject4F7A40{
				direction: tc.direction,
				power:     tc.power,
				netCode:   tc.netCode,
			}
			got := playerConfusedDirection4F7A40(&obj,
				playerConfusedDirectionHooks4F7A40[*playerConfusedDirectionTestObject4F7A40]{
					loadDirection2: func(obj *playerConfusedDirectionTestObject4F7A40) uint16 { return obj.direction },
					loadBuffPower:  func(obj *playerConfusedDirectionTestObject4F7A40, _ uint32) uint8 { return obj.power },
					loadFrame:      func() uint32 { return tc.frame },
					loadNetCode:    func(obj *playerConfusedDirectionTestObject4F7A40) uint32 { return obj.netCode },
				},
			)
			want := playerConfusedDirectionReference4F7A40(tc.direction, tc.power, tc.frame, tc.netCode)
			if got != want {
				t.Fatalf("direction = %d, want %d", got, want)
			}
			if got > 255 {
				t.Fatalf("direction = %d, want canonical byte", got)
			}
		})
	}
}

func TestPlayerConfusedDirection4F7A40Differential(t *testing.T) {
	obj := playerConfusedDirectionTestObject4F7A40{}
	frame := uint32(0)
	hooks := playerConfusedDirectionHooks4F7A40[*playerConfusedDirectionTestObject4F7A40]{
		loadDirection2: func(obj *playerConfusedDirectionTestObject4F7A40) uint16 { return obj.direction },
		loadBuffPower:  func(obj *playerConfusedDirectionTestObject4F7A40, _ uint32) uint8 { return obj.power },
		loadFrame:      func() uint32 { return frame },
		loadNetCode:    func(obj *playerConfusedDirectionTestObject4F7A40) uint32 { return obj.netCode },
	}
	for direction := uint32(0); direction <= 0xffff; direction++ {
		obj.direction = uint16(direction)
		for _, power := range []uint8{0, 255} {
			obj.power = power
			for _, phase := range []uint32{0, 10, 20, 39} {
				frame = phase
				got := playerConfusedDirection4F7A40(&obj, hooks)
				want := playerConfusedDirectionReference4F7A40(obj.direction, power, frame, 0)
				if got != want {
					t.Fatalf("direction=%#04x power=%d phase=%d: got %d, want %d",
						obj.direction, power, phase, got, want)
				}
			}
		}
	}
}
