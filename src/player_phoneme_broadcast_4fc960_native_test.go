package opennox

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestPlayerPhonemeBroadcastNative4FC960PreservesPointersAndScalars(t *testing.T) {
	source := &server.Object{NetCode: 0xfedcba98}
	listener := &server.Object{NetCode: 0x89abcdef}
	var pin runtime.Pinner
	pin.Pin(source)
	pin.Pin(listener)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 &&
		(uintptr(unsafe.Pointer(source)) <= math.MaxUint32 ||
			uintptr(unsafe.Pointer(listener)) <= math.MaxUint32) {
		t.Fatalf("test pointers do not exercise native high addresses: source=%p listener=%p", source, listener)
	}

	var (
		gotCode     uint32
		gotPhoneme  int8
		gotSound    int32
		gotSource   *server.Object
		gotKind     int32
		gotListener uint32
	)
	got := playerPhonemeBroadcastNative4FC960(
		source,
		-128,
		playerPhonemeBroadcastNativeDeps4FC960{
			firstUnit: func() *server.Object { return listener },
			nextUnit:  func(*server.Object) *server.Object { return nil },
			spellGetPhoneme: func(code uint32, phoneme int8) int32 {
				gotCode, gotPhoneme = code, phoneme
				return math.MinInt32 + 0x4321
			},
			audioEvent: func(sound int32, object *server.Object, kind int32, code uint32) {
				gotSound, gotSource, gotKind, gotListener = sound, object, kind, code
			},
		},
	)
	if got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if gotCode != source.NetCode || gotPhoneme != -128 {
		t.Fatalf("phoneme call = %#x/%d, want %#x/-128", gotCode, gotPhoneme, source.NetCode)
	}
	if gotSound != math.MinInt32+0x4321 || gotSource != source || gotKind != 2 || gotListener != listener.NetCode {
		t.Fatalf("audio call = %d/%p/%d/%#x, want %d/%p/2/%#x",
			gotSound, gotSource, gotKind, gotListener,
			math.MinInt32+0x4321, source, listener.NetCode)
	}
	runtime.KeepAlive(source)
	runtime.KeepAlive(listener)
}
