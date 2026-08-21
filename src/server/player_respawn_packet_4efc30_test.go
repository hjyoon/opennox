package server

import (
	"fmt"
	"reflect"
	"testing"
)

type playerRespawnPacketWorld4EFC30 struct {
	events  []string
	faultAt int

	unit        int
	frame       uint32
	netCode     uint32
	weaponFlags uint8
	keepItems   uint8
	sendResult  int32

	recipient int32
	packet    [9]byte
	related   int
	remove    int32
}

func (w *playerRespawnPacketWorld4EFC30) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic("injected fault")
	}
}

func (w *playerRespawnPacketWorld4EFC30) hooks() playerRespawnPacketHooks4EFC30[int] {
	return playerRespawnPacketHooks4EFC30[int]{
		loadUnitArg: func() int {
			value := w.unit
			w.event(fmt.Sprintf("unit:%d", value))
			return value
		},
		loadFrame: func() uint32 {
			value := w.frame
			w.event(fmt.Sprintf("frame:%#x", value))
			return value
		},
		loadNetCode: func(unit int) uint32 {
			value := w.netCode
			w.event(fmt.Sprintf("net:%d=%#x", unit, value))
			return value
		},
		loadWeaponFlags: func() uint8 {
			value := w.weaponFlags
			w.event(fmt.Sprintf("weapon:%#x", value))
			return value
		},
		loadKeepItemsArg: func() uint8 {
			value := w.keepItems
			w.event(fmt.Sprintf("keep:%#x", value))
			return value
		},
		sendSequence: func(recipient int32, packet [9]byte, related int, remove int32) int32 {
			w.event(fmt.Sprintf("send:%d:%x:%d:%d", recipient, packet, related, remove))
			w.recipient = recipient
			w.packet = packet
			w.related = related
			w.remove = remove
			return w.sendResult
		},
	}
}

func newPlayerRespawnPacketWorld4EFC30() *playerRespawnPacketWorld4EFC30 {
	return &playerRespawnPacketWorld4EFC30{
		unit:        7,
		frame:       0x11223344,
		netCode:     0xaabbccdd,
		weaponFlags: 0xfe,
		keepItems:   0x80,
		sendResult:  -0x76543211,
	}
}

func TestPlayerRespawnPacket4EFC30ExactPacketAndReturn(t *testing.T) {
	w := newPlayerRespawnPacketWorld4EFC30()
	got := playerRespawnPacket4EFC30(w.hooks())
	if got != w.sendResult {
		t.Fatalf("result = %#x, want %#x", got, w.sendResult)
	}
	wantPacket := [9]byte{0xe9, 0xdd, 0xcc, 0x44, 0x33, 0x22, 0x11, 0xfe, 0x80}
	if w.packet != wantPacket {
		t.Fatalf("packet = % x, want % x", w.packet, wantPacket)
	}
	if w.recipient != 255 || w.related != 0 || w.remove != 0 {
		t.Fatalf("send args = (%d, %d, %d), want (255, 0, 0)", w.recipient, w.related, w.remove)
	}
	wantEvents := []string{
		"unit:7",
		"frame:0x11223344",
		"net:7=0xaabbccdd",
		"weapon:0xfe",
		"keep:0x80",
		"send:255:e9ddcc44332211fe80:0:0",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events = %v, want %v", w.events, wantEvents)
	}
}

func TestPlayerRespawnPacket4EFC30CachedAndLiveReads(t *testing.T) {
	w := newPlayerRespawnPacketWorld4EFC30()
	hooks := w.hooks()
	originalFrame := hooks.loadFrame
	hooks.loadFrame = func() uint32 {
		value := originalFrame()
		w.unit = 99
		w.frame = 0x55667788
		return value
	}
	originalNetCode := hooks.loadNetCode
	hooks.loadNetCode = func(unit int) uint32 {
		value := originalNetCode(unit)
		w.netCode = 0x01020304
		w.weaponFlags = 0x5a
		w.keepItems = 0x33
		return value
	}
	originalWeaponFlags := hooks.loadWeaponFlags
	hooks.loadWeaponFlags = func() uint8 {
		value := originalWeaponFlags()
		w.weaponFlags = 0xa5
		w.keepItems = 0x77
		return value
	}

	playerRespawnPacket4EFC30(hooks)
	wantPacket := [9]byte{0xe9, 0xdd, 0xcc, 0x44, 0x33, 0x22, 0x11, 0x5a, 0x77}
	if w.packet != wantPacket {
		t.Fatalf("packet = % x, want cached/live % x", w.packet, wantPacket)
	}
	if got := w.events[2]; got != "net:7=0xaabbccdd" {
		t.Fatalf("cached unit net-code event = %q", got)
	}
}

func TestPlayerRespawnPacket4EFC30FaultPrefixes(t *testing.T) {
	base := newPlayerRespawnPacketWorld4EFC30()
	playerRespawnPacket4EFC30(base.hooks())
	want := append([]string(nil), base.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := newPlayerRespawnPacketWorld4EFC30()
			w.faultAt = faultAt
			panicked := false
			func() {
				defer func() {
					panicked = recover() != nil
				}()
				playerRespawnPacket4EFC30(w.hooks())
			}()
			if !panicked {
				t.Fatal("expected injected fault")
			}
			if !reflect.DeepEqual(w.events, want[:faultAt]) {
				t.Fatalf("events = %v, want prefix %v", w.events, want[:faultAt])
			}
		})
	}
}
