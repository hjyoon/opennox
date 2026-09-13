package server

import (
	"encoding/binary"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet"
	"github.com/opennox/libs/noxnet/netmsg"
)

func TestChatPacket528AC0NativePointer(t *testing.T) {
	obj := &Object{NetCode: 0x1234}
	obj.PosVec.X = 123.75
	obj.PosVec.Y = -2.9
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(obj)) <= 0xffffffff {
		t.Skip("allocator did not provide a high-address object")
	}
	packet := new(Server).chatPacket528AC0(obj, "Hello", 99)
	want := []byte{byte(netmsg.MSG_TEXT_MESSAGE), 0x34, 0x12, 2, 123, 0, 0xfe, 0xff, 6, 99, 0, 'H', 'e', 'l', 'l', 'o', 0}
	if string(packet) != string(want) {
		t.Fatalf("chat packet = %x, want %x", packet, want)
	}
	var decoded noxnet.MsgText
	if _, err := decoded.Decode(packet[1:]); err != nil {
		t.Fatal(err)
	}
	if decoded.Text() != "Hello" {
		t.Fatalf("decoded text = %q", decoded.Text())
	}
}

func TestChatPacket528AC0WideAndNarrow(t *testing.T) {
	obj := &Object{NetCode: 0x104}
	tests := []struct {
		name    string
		message string
		flags   byte
		size    byte
		data    []byte
	}{
		{"latin1", "é", 2, 2, []byte{0xe9, 0}},
		{"wide", "漢🙂", 4, 4, []byte{0x22, 0x6f, 0x3d, 0xd8, 0x42, 0xde, 0, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packet := new(Server).chatPacket528AC0(obj, tt.message, 0x7bcd)
			if packet[3] != tt.flags || packet[8] != tt.size || binary.LittleEndian.Uint16(packet[9:11]) != 0x7bcd || string(packet[11:]) != string(tt.data) {
				t.Fatalf("chat packet = %x, flags %d size %d data %x", packet, tt.flags, tt.size, tt.data)
			}
		})
	}
}

func TestChatPacket528AC0Bounds(t *testing.T) {
	obj := &Object{}
	narrow := new(Server).chatPacket528AC0(obj, strings.Repeat("a", 300), 0)
	if len(narrow) != 266 || narrow[8] != 255 || narrow[len(narrow)-1] != 0 {
		t.Fatalf("bounded narrow packet: length %d, size %d", len(narrow), narrow[8])
	}
	wide := new(Server).chatPacket528AC0(obj, strings.Repeat("🙂", 200), 0)
	if len(wide) > 520 || wide[8] != 253 || binary.LittleEndian.Uint16(wide[len(wide)-4:len(wide)-2]) != 0xde42 || binary.LittleEndian.Uint16(wide[len(wide)-2:]) != 0 {
		t.Fatalf("bounded wide packet: length %d, size %d, tail %x", len(wide), wide[8], wide[len(wide)-4:])
	}
}
