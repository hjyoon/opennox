package legacy

import (
	"math"
	"strconv"
	"testing"

	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/server"
)

type mapgenAdjacentReservationLegacyServer524FB0 struct {
	Server
	srv *server.Server
}

func (s *mapgenAdjacentReservationLegacyServer524FB0) S() *server.Server {
	return s.srv
}

func TestMapgenAdjacentReservation524FB0PreservesNativePointers(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	if srv.Walls.Init() == 0 {
		t.Fatal("cannot initialize server wall grid")
	}
	t.Cleanup(srv.Walls.Free)
	oldGetServer := GetServer
	GetServer = func() Server {
		return &mapgenAdjacentReservationLegacyServer524FB0{srv: srv}
	}
	t.Cleanup(func() { GetServer = oldGetServer })

	got := mapgenAdjacentReservationFixture524FB0()
	if !got.usedLegacyWrapper {
		t.Fatal("direction zero did not exercise the legacy token wrapper")
	}

	const tile = float32(32.526913)
	roomWidth := 4 * tile
	roomHeight := 3 * tile
	neighborWidth := 2 * tile
	want := [4][4]float32{
		{110, 200, 110 + neighborWidth, 200 + tile},
		{110, 200 + roomHeight - tile, 110 + neighborWidth, 200 + roomHeight},
		{100 + roomWidth - tile, 210, 100 + roomWidth, 210 + roomHeight},
		{100, 210, 100 + tile, 210 + roomHeight},
	}

	for direction, item := range got.cases {
		if !item.added || item.rectangleAddress == 0 || item.rectangleToken == 0 {
			t.Fatalf("direction %d did not add an occupied rectangle: %+v", direction, item)
		}
		values := [4]float32{item.minX, item.minY, item.maxX, item.maxY}
		for index := range values {
			if diff := math.Abs(float64(values[index] - want[direction][index])); diff > 0.001 {
				t.Fatalf("direction %d bound %d = %f, want %f", direction, index, values[index], want[direction][index])
			}
		}

		if item.roomToken == 0 || item.neighborToken == 0 {
			t.Fatalf("direction %d stored a zero room token: %+v", direction, item)
		}
		if strconv.IntSize == 64 {
			for name, address := range map[string]uintptr{
				"room":      item.roomAddress,
				"neighbor":  item.neighborAddress,
				"rectangle": item.rectangleAddress,
			} {
				if address <= 1<<32 {
					t.Fatalf("direction %d %s fixture address = %#x, want address above PE32 range", direction, name, address)
				}
			}
			if uintptr(item.roomToken) == item.roomAddress || uintptr(item.neighborToken) == item.neighborAddress {
				t.Fatalf("direction %d room token unexpectedly equals a native address: %+v", direction, item)
			}
			if uintptr(item.rectangleToken) == item.rectangleAddress {
				t.Fatalf("direction %d rectangle token %#x unexpectedly equals native address %#x", direction, item.rectangleToken, item.rectangleAddress)
			}
		}
	}
}
