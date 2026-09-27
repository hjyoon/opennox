package legacy

import (
	"encoding/binary"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

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
	handles.Init()
	t.Cleanup(handles.Release)
	pool := alloc.NewClass("chat-bubble-native-list", ChatBubbleSize(), 4)
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
