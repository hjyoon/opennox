package opennox

import (
	"image"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
)

func TestDecodeClientPredictLinearState48EA70PacketWidth(t *testing.T) {
	data := []byte{
		0xb5,
		0x34, 0x92,
		0x78, 0x56,
		0x10, 0x27,
		0x20, 0x4e,
		0xef, 0xbe,
		0xf0,
		0x7f,
		0x80,
		0xaa, 0xbb,
	}
	state, ok := decodeClientPredictLinearState48EA70(data)
	if !ok {
		t.Fatal("valid predicted-linear packet was rejected")
	}
	want := clientPredictLinearState48EA70{
		Code:      0x9234,
		TypeID:    0x5678,
		Pos:       image.Pt(10000, 20000),
		Field127:  0xbeef,
		Damping:   -16,
		VelocityX: 127,
		VelocityY: -128,
	}
	if !reflect.DeepEqual(state, want) {
		t.Fatalf("decoded state = %+v, want %+v", state, want)
	}
	for n := 0; n < clientPredictLinearPacketSize48EA70; n++ {
		if _, ok := decodeClientPredictLinearState48EA70(data[:n]); ok {
			t.Fatalf("%d-byte predicted-linear packet was accepted", n)
		}
	}
}

func TestHandleClientPredictLinearNative48EA70Disconnected(t *testing.T) {
	var calls []string
	hooks := clientPredictLinearHooks48EA70{
		connected: func() bool {
			calls = append(calls, "connected")
			return false
		},
		create: func(int, uint16, int, int) *client.Drawable {
			calls = append(calls, "create")
			return nil
		},
	}
	if got := handleClientPredictLinearNative48EA70(make([]byte, 14), hooks); got != 14 {
		t.Fatalf("consumed bytes = %d, want 14", got)
	}
	if want := []string{"connected"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("callback order = %v, want %v", calls, want)
	}
}

func TestHandleClientPredictLinearNative48EA70HighAddress(t *testing.T) {
	dr := &client.Drawable{Field_127: 0xabcd0000}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", dr)
	}
	drawMarker := new(byte)
	updateMarker := new(byte)
	dr.DrawFuncPtr = unsafe.Pointer(drawMarker)
	dr.ClientUpdateFuncPtr = unsafe.Pointer(updateMarker)

	var calls []string
	secondaryMarker := new(byte)
	hooks := clientPredictLinearHooks48EA70{
		connected: func() bool {
			calls = append(calls, "connected")
			return true
		},
		create: func(typeID int, code uint16, x, y int) *client.Drawable {
			calls = append(calls, "create")
			if typeID != 0x5678 || code != 0x1234 || x != 1000 || y != 2000 {
				t.Fatalf("create args = (%#x, %#x, %d, %d)", typeID, code, x, y)
			}
			return dr
		},
		frame: func() uint32 {
			calls = append(calls, "frame")
			return 1234
		},
		secondaryUpdate: unsafe.Pointer(secondaryMarker),
		activate: func(got *client.Drawable) {
			calls = append(calls, "activate")
			if got != dr {
				t.Fatalf("activated drawable = %p, want %p", got, dr)
			}
		},
	}
	data := []byte{
		0xb5,
		0x34, 0x92,
		0x78, 0x56,
		0xe8, 0x03,
		0xd0, 0x07,
		0xef, 0xbe,
		0x04,
		0x18,
		0xe0,
	}
	if got := handleClientPredictLinearNative48EA70(data, hooks); got != 14 {
		t.Fatalf("consumed bytes = %d, want 14", got)
	}
	if dr.Field_127 != 0xabcdbeef || dr.AnimStart != 1234 || dr.Field_81 != 1000 || dr.Field_82 != 2000 {
		t.Fatalf("integer fields = field127:%#x start:%d origin:(%d,%d)", dr.Field_127, dr.AnimStart, dr.Field_81, dr.Field_82)
	}
	if got := math.Float32frombits(dr.Field_117); got != 1.5 {
		t.Fatalf("velocity X = %g, want 1.5", got)
	}
	if got := math.Float32frombits(dr.Field_118); got != -2 {
		t.Fatalf("velocity Y = %g, want -2", got)
	}
	if got := math.Float32frombits(dr.Field_119); got != 0.25 {
		t.Fatalf("damping = %g, want 0.25", got)
	}
	if dr.Field_115 != unsafe.Pointer(secondaryMarker) {
		t.Fatalf("secondary callback = %p, want %p", dr.Field_115, secondaryMarker)
	}
	if dr.DrawFuncPtr != unsafe.Pointer(drawMarker) || dr.ClientUpdateFuncPtr != unsafe.Pointer(updateMarker) {
		t.Fatalf("draw/update callbacks were corrupted: draw=%p update=%p", dr.DrawFuncPtr, dr.ClientUpdateFuncPtr)
	}
	if want := []string{"connected", "create", "frame", "activate"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("callback order = %v, want %v", calls, want)
	}
}

func TestHandleClientPredictLinearNative48EA70Guards(t *testing.T) {
	if got := handleClientPredictLinearNative48EA70(make([]byte, 13), clientPredictLinearHooks48EA70{}); got != -1 {
		t.Fatalf("short packet consumed bytes = %d, want -1", got)
	}
	created := 0
	hooks := clientPredictLinearHooks48EA70{
		connected: func() bool { return true },
		create: func(int, uint16, int, int) *client.Drawable {
			created++
			return nil
		},
	}
	if got := handleClientPredictLinearNative48EA70(make([]byte, 14), hooks); got != 14 || created != 1 {
		t.Fatalf("nil-create result = (%d, %d), want (14, 1)", got, created)
	}
}

func TestClientPredictLinearPosition4CA540(t *testing.T) {
	dr := &client.Drawable{
		AnimStart: 100,
		Field_81:  100,
		Field_82:  200,
		Field_117: math.Float32bits(2),
		Field_118: math.Float32bits(-1),
		Field_119: math.Float32bits(0.25),
	}
	if got := clientPredictLinearPosition4CA540(dr, 100); got != image.Pt(102, 199) {
		t.Fatalf("one-step position = %v, want (102,199)", got)
	}
	if got := clientPredictLinearPosition4CA540(dr, 101); got != image.Pt(103, 199) {
		t.Fatalf("two-step position = %v, want (103,199)", got)
	}
}

func TestUpdateClientPredictLinearNative4CA540HighAddress(t *testing.T) {
	dr := &client.Drawable{
		AnimStart: 100,
		Field_81:  100,
		Field_82:  200,
		Field_117: math.Float32bits(2),
		Field_118: math.Float32bits(-1),
		Field_119: math.Float32bits(0.25),
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", dr)
	}
	vp := &noxrender.Viewport{
		Screen: image.Rect(10, 20, 650, 500),
		World:  image.Rect(90, 190, 730, 670),
	}
	var calls []string
	hooks := clientPredictLinearUpdateHooks4CA540{
		frame: func() uint32 {
			calls = append(calls, "frame")
			return 100
		},
		move: func(got *client.Drawable, x, y int) {
			calls = append(calls, "move")
			if got != dr || x != 102 || y != 199 {
				t.Fatalf("move = (%p, %d, %d), want (%p, 102, 199)", got, x, y, dr)
			}
			got.PosVec = image.Pt(x, y)
		},
		visible: func(x, y int) int {
			calls = append(calls, "visible")
			if x != 22 || y != 29 {
				t.Fatalf("visibility point = (%d, %d), want (22, 29)", x, y)
			}
			return 1
		},
		remove: func(*client.Drawable) {
			calls = append(calls, "remove")
		},
	}
	if got := updateClientPredictLinearNative4CA540(vp, dr, hooks); got != 1 {
		t.Fatalf("update result = %d, want 1", got)
	}
	if want := []string{"frame", "move", "visible"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("callback order = %v, want %v", calls, want)
	}
}

func TestUpdateClientPredictLinearNative4CA540Removal(t *testing.T) {
	dr := &client.Drawable{AnimStart: 7, Field_81: 100, Field_82: 200}
	removed := 0
	hooks := clientPredictLinearUpdateHooks4CA540{
		frame: func() uint32 { return 7 },
		move: func(got *client.Drawable, x, y int) {
			got.PosVec = image.Pt(x, y)
		},
		visible: func(int, int) int { return 0 },
		remove: func(got *client.Drawable) {
			if got != dr {
				t.Fatalf("removed drawable = %p, want %p", got, dr)
			}
			removed++
		},
	}
	if got := updateClientPredictLinearNative4CA540(nil, dr, hooks); got != 0 || removed != 1 {
		t.Fatalf("invisible update = (%d, %d removals), want (0, 1)", got, removed)
	}

	dr.Field_81 = 0
	moveCalls := 0
	hooks.move = func(*client.Drawable, int, int) { moveCalls++ }
	hooks.visible = func(int, int) int {
		t.Fatal("out-of-bounds drawable reached visibility test")
		return 0
	}
	if got := updateClientPredictLinearNative4CA540(nil, dr, hooks); got != 0 || removed != 2 || moveCalls != 0 {
		t.Fatalf("out-of-bounds update = result:%d removals:%d moves:%d", got, removed, moveCalls)
	}
}
