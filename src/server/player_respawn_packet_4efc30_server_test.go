package server

import (
	"reflect"
	"testing"
	"unsafe"
)

func TestPlayerRespawnPacket4EFC30NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantNetCode := uintptr(36)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantNetCode = 40
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.NetCode", unsafe.Offsetof(Object{}.NetCode), wantNetCode},
		{"Object.NetCode width", unsafe.Sizeof(Object{}.NetCode), 4},
		{"frame width", unsafe.Sizeof(uint32(0)), 4},
		{"packet byte width", unsafe.Sizeof(uint8(0)), 1},
		{"sender result width", unsafe.Sizeof(int32(0)), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestPlayerRespawnPacketNative4EFC30(t *testing.T) {
	unit := &Object{NetCode: 0xaabbccdd}
	events := make([]string, 0, 3)
	var gotRecipient int32
	var gotPacket []byte
	var gotRelated *Object
	var gotRemove int32
	got := playerRespawnPacketNative4EFC30(unit, 0x80, playerRespawnPacketNativeDeps4EFC30{
		loadFrame: func() uint32 {
			events = append(events, "frame")
			return 0x11223344
		},
		loadWeaponFlags: func() uint8 {
			events = append(events, "weapon")
			return 0xfe
		},
		sendSequence: func(recipient int32, packet []byte, related *Object, remove int32) int32 {
			events = append(events, "send")
			gotRecipient = recipient
			gotPacket = append([]byte(nil), packet...)
			gotRelated = related
			gotRemove = remove
			return -0x76543211
		},
	})
	if got != -0x76543211 {
		t.Fatalf("result = %#x", got)
	}
	wantPacket := []byte{0xe9, 0xdd, 0xcc, 0x44, 0x33, 0x22, 0x11, 0xfe, 0x80}
	if !reflect.DeepEqual(gotPacket, wantPacket) {
		t.Fatalf("packet = % x, want % x", gotPacket, wantPacket)
	}
	if gotRecipient != 255 || gotRelated != nil || gotRemove != 0 {
		t.Fatalf("send args = (%d, %p, %d)", gotRecipient, gotRelated, gotRemove)
	}
	if want := []string{"frame", "weapon", "send"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestPlayerRespawnPacketNative4EFC30NilFaultAfterFrame(t *testing.T) {
	events := make([]string, 0, 2)
	panicked := false
	func() {
		defer func() {
			panicked = recover() != nil
		}()
		playerRespawnPacketNative4EFC30(nil, 1, playerRespawnPacketNativeDeps4EFC30{
			loadFrame: func() uint32 {
				events = append(events, "frame")
				return 7
			},
			loadWeaponFlags: func() uint8 {
				events = append(events, "weapon")
				return 0
			},
			sendSequence: func(int32, []byte, *Object, int32) int32 {
				events = append(events, "send")
				return 0
			},
		})
	}()
	if !panicked {
		t.Fatal("nil unit did not preserve the original NetCode fault")
	}
	if want := []string{"frame"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}
