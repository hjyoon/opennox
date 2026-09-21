package opennox

import (
	"encoding/binary"
	"image"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

func TestDecodeTurnUndeadFXPosition48EA70PacketWidth(t *testing.T) {
	data := []byte{0xa0, 0x00, 0x80, 0xff, 0x7f, 0xaa, 0xbb}
	pos, ok := decodeTurnUndeadFXPosition48EA70(data)
	if !ok || pos != image.Pt(-32768, 32767) {
		t.Fatalf("decoded position = %v, ok=%t", pos, ok)
	}
	for n := 0; n < turnUndeadFXPacketSize48EA70; n++ {
		if _, ok := decodeTurnUndeadFXPosition48EA70(data[:n]); ok {
			t.Fatalf("%d-byte turn-undead packet was accepted", n)
		}
	}
}

func TestHandleTurnUndeadFXNative48EA70HighAddress(t *testing.T) {
	const (
		typeID = 91
		frame  = 0x12345678
	)
	pos := image.Pt(-32768, 32767)
	data := make([]byte, turnUndeadFXPacketSize48EA70)
	data[0] = 0xa0
	binary.LittleEndian.PutUint16(data[1:3], uint16(int16(pos.X)))
	binary.LittleEndian.PutUint16(data[3:5], uint16(int16(pos.Y)))

	drawMarker := new(byte)
	updateMarker := new(byte)
	secondaryMarker := new(byte)
	directions := make(map[*client.Drawable]int, turnUndeadFXDrawableCount48EA70)
	activated := make(map[*client.Drawable]bool, turnUndeadFXDrawableCount48EA70)
	var drawables []*client.Drawable
	spawnCount := 0
	vectorCount := 0
	frameCount := 0
	sightCount := 0
	hooks := turnUndeadFXHooks48EA70{
		connected: func() bool { return true },
		typeID:    func() int { return typeID },
		spawn: func(gotType int, gotPos image.Point) *client.Drawable {
			if gotType != typeID || gotPos != pos {
				t.Fatalf("spawn %d = type %d at %v, want type %d at %v", spawnCount, gotType, gotPos, typeID, pos)
			}
			direction := spawnCount * 6
			dr := &client.Drawable{
				Field_127:           0xabcd0000,
				Field_119:           0xffffffff,
				DrawFuncPtr:         unsafe.Pointer(drawMarker),
				ClientUpdateFuncPtr: unsafe.Pointer(updateMarker),
			}
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
				t.Skipf("allocator returned a low address: %p", dr)
			}
			directions[dr] = direction
			drawables = append(drawables, dr)
			spawnCount++
			return dr
		},
		directionVector: func(direction byte) (float32, float32) {
			want := byte(vectorCount * 6)
			if direction != want {
				t.Fatalf("direction vector %d = %d, want %d", vectorCount, direction, want)
			}
			vectorCount++
			return float32(direction) / 8, -float32(direction) / 16
		},
		frame: func() uint32 {
			frameCount++
			return frame
		},
		secondaryUpdate: unsafe.Pointer(secondaryMarker),
		activate: func(dr *client.Drawable) {
			if _, ok := directions[dr]; !ok {
				t.Fatalf("activated unknown drawable %p", dr)
			}
			activated[dr] = true
		},
		sightDestroy: func(dr *client.Drawable) {
			if !activated[dr] {
				t.Fatalf("sight-destroy registration preceded activation for %p", dr)
			}
			sightCount++
		},
	}

	if got := handleTurnUndeadFXNative48EA70(data, hooks); got != turnUndeadFXPacketSize48EA70 {
		t.Fatalf("consumed bytes = %d, want %d", got, turnUndeadFXPacketSize48EA70)
	}
	if spawnCount != turnUndeadFXDrawableCount48EA70 || vectorCount != spawnCount || frameCount != spawnCount || sightCount != spawnCount {
		t.Fatalf("callback counts = spawn:%d vector:%d frame:%d sight:%d, want %d each",
			spawnCount, vectorCount, frameCount, sightCount, turnUndeadFXDrawableCount48EA70)
	}
	for _, dr := range drawables {
		direction := directions[dr]
		if dr.Field_127 != 0xabcd0000|uint32(direction) {
			t.Errorf("direction %d field 127 = %#x", direction, dr.Field_127)
		}
		if got, want := math.Float32frombits(dr.Field_117), float32(direction)/2; got != want {
			t.Errorf("direction %d velocity X = %g, want %g", direction, got, want)
		}
		if got, want := math.Float32frombits(dr.Field_118), -float32(direction)/4; got != want {
			t.Errorf("direction %d velocity Y = %g, want %g", direction, got, want)
		}
		if dr.Field_119 != 0 || dr.AnimStart != frame || dr.Field_81 != uint32(pos.X) || dr.Field_82 != uint32(pos.Y) {
			t.Errorf("direction %d state = damping:%#x frame:%#x origin:(%#x,%#x)",
				direction, dr.Field_119, dr.AnimStart, dr.Field_81, dr.Field_82)
		}
		if dr.Field_115 != unsafe.Pointer(secondaryMarker) {
			t.Errorf("direction %d secondary update = %p, want %p", direction, dr.Field_115, secondaryMarker)
		}
		if dr.DrawFuncPtr != unsafe.Pointer(drawMarker) || dr.ClientUpdateFuncPtr != unsafe.Pointer(updateMarker) {
			t.Errorf("direction %d callbacks corrupted: draw=%p update=%p", direction, dr.DrawFuncPtr, dr.ClientUpdateFuncPtr)
		}
	}
}

func TestHandleTurnUndeadFXNative48EA70Guards(t *testing.T) {
	for n := 0; n < turnUndeadFXPacketSize48EA70; n++ {
		if got := handleTurnUndeadFXNative48EA70(make([]byte, n), turnUndeadFXHooks48EA70{}); got != -1 {
			t.Fatalf("%d-byte packet consumed = %d, want -1", n, got)
		}
	}

	connectedCalls := 0
	hooks := turnUndeadFXHooks48EA70{
		connected: func() bool {
			connectedCalls++
			return false
		},
		typeID: func() int {
			t.Fatal("type lookup ran while disconnected")
			return 0
		},
	}
	if got := handleTurnUndeadFXNative48EA70(make([]byte, turnUndeadFXPacketSize48EA70), hooks); got != turnUndeadFXPacketSize48EA70 || connectedCalls != 1 {
		t.Fatalf("disconnected result = consumed:%d connected:%d", got, connectedCalls)
	}

	spawnCount := 0
	hooks = turnUndeadFXHooks48EA70{
		connected: func() bool { return true },
		typeID:    func() int { return 1 },
		spawn: func(int, image.Point) *client.Drawable {
			spawnCount++
			return nil
		},
		directionVector: func(byte) (float32, float32) {
			t.Fatal("direction lookup ran for a nil drawable")
			return 0, 0
		},
		frame: func() uint32 {
			t.Fatal("frame lookup ran for a nil drawable")
			return 0
		},
		activate: func(*client.Drawable) { t.Fatal("nil drawable was activated") },
		sightDestroy: func(*client.Drawable) {
			t.Fatal("nil drawable entered sight-destroy list")
		},
	}
	if got := handleTurnUndeadFXNative48EA70(make([]byte, turnUndeadFXPacketSize48EA70), hooks); got != turnUndeadFXPacketSize48EA70 || spawnCount != turnUndeadFXDrawableCount48EA70 {
		t.Fatalf("nil-spawn result = consumed:%d spawns:%d", got, spawnCount)
	}
}
