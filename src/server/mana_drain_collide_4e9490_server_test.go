package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestManaDrainCollideNative4E9490(t *testing.T) {
	data := &ManaDrainCollideData{Amount: 10, Reserved: [7]uint8{1, 2, 3, 4, 5, 6, 7}}
	source := &Object{CollideData: unsafe.Pointer(data), Field542: 100}
	player := &Player{ProtUnitManaCur: 0xfedcba98}
	update := &PlayerUpdateData{ManaCur: 100, ManaPrev: 7, Player: player}
	target := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	frames := []uint32{131, 0x12345}
	var collisionMarker byte
	var gotToken uint32
	var gotDelta int16
	var gotAudio *Object
	manaDrainCollideNative4E9490(source, target, unsafe.Pointer(&collisionMarker), manaDrainNativeDeps4E9490{
		loadFrame: func() uint32 {
			frame := frames[0]
			frames = frames[1:]
			return frame
		},
		loadFPS: func() uint32 { return 60 },
		godMode: func() bool { return false },
		protectMana: func(token uint32, delta int16) {
			gotToken, gotDelta = token, delta
		},
		audio: func(obj *Object) { gotAudio = obj },
	})
	if update.ManaPrev != 100 || update.ManaCur != 90 {
		t.Fatalf("mana prev/current = %d/%d, want 100/90", update.ManaPrev, update.ManaCur)
	}
	if gotToken != player.ProtUnitManaCur || gotDelta != -10 {
		t.Fatalf("protect = (%#x,%d)", gotToken, gotDelta)
	}
	if gotAudio != source || source.Field542 != 0x2345 {
		t.Fatalf("audio/source frame = %p/%#x", gotAudio, source.Field542)
	}
	if data.Reserved != [7]uint8{1, 2, 3, 4, 5, 6, 7} {
		t.Fatalf("reserved bytes changed: %v", data.Reserved)
	}
}

func TestManaDrainCollideNative4E9490ZeroManaDoesNotReadSource(t *testing.T) {
	target := &Object{
		ObjClass:   object.ClassPlayer,
		UpdateData: unsafe.Pointer(&PlayerUpdateData{}),
	}
	manaDrainCollideNative4E9490(nil, target, nil, manaDrainNativeDeps4E9490{
		loadFrame:   func() uint32 { t.Fatal("frame read"); return 0 },
		loadFPS:     func() uint32 { t.Fatal("FPS read"); return 0 },
		godMode:     func() bool { t.Fatal("god-mode read"); return false },
		protectMana: func(uint32, int16) { t.Fatal("protection update") },
		audio:       func(*Object) { t.Fatal("audio") },
	})
}

func TestManaDrainCollideNative4E9490GodModeStillRunsThrottle(t *testing.T) {
	data := &ManaDrainCollideData{Amount: 20}
	source := &Object{CollideData: unsafe.Pointer(data)}
	player := &Player{ProtUnitManaCur: 9}
	update := &PlayerUpdateData{ManaCur: 15, ManaPrev: 3, Player: player}
	target := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	frames := []uint32{1, 2}
	audio := 0
	manaDrainCollideNative4E9490(source, target, nil, manaDrainNativeDeps4E9490{
		loadFrame: func() uint32 {
			frame := frames[0]
			frames = frames[1:]
			return frame
		},
		loadFPS: func() uint32 { return 0 },
		godMode: func() bool { return true },
		protectMana: func(uint32, int16) {
			t.Fatal("protection update in god mode")
		},
		audio: func(*Object) { audio++ },
	})
	if update.ManaCur != 15 || update.ManaPrev != 3 || audio != 1 || source.Field542 != 2 {
		t.Fatalf("god-mode result: mana=%d/%d audio=%d frame=%d", update.ManaCur, update.ManaPrev, audio, source.Field542)
	}
}

func TestManaDrainCollide4E9490Layouts(t *testing.T) {
	wantClass, wantTimer, wantCollideData, wantUpdate := uintptr(8), uintptr(542), uintptr(700), uintptr(748)
	wantPlayerPtr, wantProtection := uintptr(276), uintptr(4596)
	wantObjectSize, wantUpdateSize, wantPlayerSize := uintptr(780), uintptr(556), uintptr(4828)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantClass, wantTimer, wantCollideData, wantUpdate = 12, 602, 776, 872
		wantPlayerPtr, wantProtection = 320, 5900
		wantObjectSize, wantUpdateSize, wantPlayerSize = 928, 640, 6160
	}
	tests := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"ManaDrainCollideData size", unsafe.Sizeof(ManaDrainCollideData{}), 8},
		{"ManaDrainCollideData.Amount", unsafe.Offsetof(ManaDrainCollideData{}.Amount), 0},
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.Field542", unsafe.Offsetof(Object{}.Field542), wantTimer},
		{"Object.CollideData", unsafe.Offsetof(Object{}.CollideData), wantCollideData},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantUpdateSize},
		{"PlayerUpdateData.ManaCur", unsafe.Offsetof(PlayerUpdateData{}.ManaCur), 4},
		{"PlayerUpdateData.ManaPrev", unsafe.Offsetof(PlayerUpdateData{}.ManaPrev), 6},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayerPtr},
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.ProtUnitManaCur", unsafe.Offsetof(Player{}.ProtUnitManaCur), wantProtection},
	}
	for _, tc := range tests {
		if tc.got != tc.want {
			t.Errorf("%s = %d, want %d", tc.name, tc.got, tc.want)
		}
	}
}
