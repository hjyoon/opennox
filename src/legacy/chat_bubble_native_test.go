package legacy

import (
	"encoding/binary"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

func TestChatBubbleCandidate48E6A0NativeWord(t *testing.T) {
	g := newWidgetCallbackGUI(t)
	defer g.DestroyAll()

	moving := chatBubbleCandidateRecord{x: 30, y: 40, width: 11, height: 17}
	earlier := chatBubbleCandidateRecord{x: 100, y: 200, width: 23, height: 29}
	out, free := alloc.NewHigh([2]int32{})
	defer free()
	xWord, yWord := uintptr(unsafe.Pointer(&out[0])), uintptr(unsafe.Pointer(&out[1]))
	if unsafe.Sizeof(xWord) > 4 && xWord <= uintptr(^uint32(0)) {
		t.Fatalf("output address %#x does not exercise the native width", xWord)
	}
	want := map[byte]struct {
		coords [2]int32
		word   uintptr
	}{
		0x01: {[2]int32{89, 183}, 0},
		0x02: {[2]int32{123, 183}, yWord},
		0x04: {[2]int32{89, 229}, yWord},
		0x08: {[2]int32{123, 229}, yWord},
		0x10: {[2]int32{30, 183}, yWord},
		0x20: {[2]int32{30, 229}, yWord},
		0x40: {[2]int32{89, 40}, xWord},
		0x80: {[2]int32{123, 40}, yWord},
	}
	for i := range 256 {
		dir := byte(i)
		*out = [2]int32{-111, -222}
		if tc, ok := want[dir]; ok {
			if got := chatBubbleCandidateCallC(dir, &moving, &earlier, &out[0], &out[1]); got != tc.word || *out != tc.coords {
				t.Fatalf("direction %#02x returned %#x and %v, want %#x and %v", dir, got, *out, tc.word, tc.coords)
			}
		} else if got := chatBubbleCandidateCallC(dir, nil, nil, &out[0], &out[1]); got != 0 || *out != [2]int32{-111, -222} {
			t.Fatalf("invalid direction %#02x returned %#x and changed outputs to %v", dir, got, *out)
		}
	}

	aliasX := (*int32)(unsafe.Pointer(&earlier.y))
	got := chatBubbleCandidateCallC(1, &moving, &earlier, aliasX, &out[1])
	if got != 0 || earlier.y != 89 || out[1] != 72 {
		t.Fatalf("aliased outputs returned %#x with %d, %d, want 0 with 89, 72", got, earlier.y, out[1])
	}
}

func TestChatBubbleNativeListLifecycle(t *testing.T) {
	if got, min := ChatBubbleSize(), uintptr(692)+3*unsafe.Sizeof(uintptr(0)); got < min {
		t.Fatalf("chat bubble size = %d, want at least %d", got, min)
	}
	if old := Get_nox_alloc_chat_1197364(); old != nil {
		t.Fatalf("chat allocation class is already initialized: %p", old)
	}

	oldFrame, oldFPS := gameFrameHook, gameFPSHook
	gameFrameHook = func() uint32 { return 100 }
	gameFPSHook = func() uint32 { return 30 }
	handles.InitHigh()
	t.Cleanup(handles.Release)
	pool := alloc.NewClassHigh("chat-bubble-native-list", ChatBubbleSize(), 4)
	handle := pool.UPtr()
	Set_nox_alloc_chat_1197364(handle)
	t.Cleanup(func() {
		ResetChatBubbles()
		Set_nox_alloc_chat_1197364(nil)
		if alloc.AsClass(handle) != nil {
			pool.Free()
		}
		gameFrameHook, gameFPSHook = oldFrame, oldFPS
	})

	packet := func(netCode uint16, x, y uint16) []byte {
		out := make([]byte, 11)
		binary.LittleEndian.PutUint16(out[1:3], netCode)
		binary.LittleEndian.PutUint16(out[4:6], x)
		binary.LittleEndian.PutUint16(out[6:8], y)
		out[8] = 16
		binary.LittleEndian.PutUint16(out[9:11], 90)
		return out
	}

	Nox_xxx_createTextBubble_48D880(packet(0x1234, 100, 200), "Jerry")
	first := Nox_xxx_netCode2ChatBubble_48D850(0x1234)
	if first == 0 {
		t.Fatal("first chat bubble was not linked")
	}
	Nox_xxx_createTextBubble_48D880(packet(0x5678, 300, 400), "Bryan")
	second := Nox_xxx_netCode2ChatBubble_48D850(0x5678)
	if second == 0 || second == first {
		t.Fatalf("second chat bubble = %#x, first = %#x", second, first)
	}

	Nox_xxx_createTextBubble_48D880(packet(0x1234, 120, 220), "Jerry again")
	if got := Nox_xxx_netCode2ChatBubble_48D850(0x1234); got != first {
		t.Fatalf("updated chat bubble = %#x, want stable pointer %#x", got, first)
	}
	RemoveChatBubble(0x1234)
	if got := Nox_xxx_netCode2ChatBubble_48D850(0x1234); got != 0 {
		t.Fatalf("removed chat bubble still linked at %#x", got)
	}
	if got := Nox_xxx_netCode2ChatBubble_48D850(0x5678); got != second {
		t.Fatalf("remaining chat bubble = %#x, want %#x", got, second)
	}

	ResetChatBubbles()
	if got := Nox_xxx_netCode2ChatBubble_48D850(0x5678); got != 0 {
		t.Fatalf("reset chat bubble still linked at %#x", got)
	}
}
