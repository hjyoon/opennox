package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/wall"
)

func TestSecretWallNativeLayout(t *testing.T) {
	ptrSize := unsafe.Sizeof(uintptr(0))
	wantSize := uintptr(32)
	wantX := uintptr(4)
	wantY := uintptr(8)
	wantWall := uintptr(12)
	wantOpenWait := uintptr(16)
	wantFlags := uintptr(20)
	wantState := uintptr(21)
	wantOpenDelay := uintptr(22)
	wantLastOpen := uintptr(24)
	wantPlayerBits := uintptr(28)
	if ptrSize == 8 {
		wantSize = 40
		wantX = 8
		wantY = 12
		wantWall = 16
		wantOpenWait = 24
		wantFlags = 28
		wantState = 29
		wantOpenDelay = 30
		wantLastOpen = 32
		wantPlayerBits = 36
	}
	tests := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"size", unsafe.Sizeof(SecretWall{}), wantSize},
		{"next", unsafe.Offsetof(SecretWall{}.Next), 0},
		{"x", unsafe.Offsetof(SecretWall{}.X), wantX},
		{"y", unsafe.Offsetof(SecretWall{}.Y), wantY},
		{"wall", unsafe.Offsetof(SecretWall{}.Wall), wantWall},
		{"open wait", unsafe.Offsetof(SecretWall{}.OpenWait), wantOpenWait},
		{"flags", unsafe.Offsetof(SecretWall{}.Flags), wantFlags},
		{"state", unsafe.Offsetof(SecretWall{}.State), wantState},
		{"open delay", unsafe.Offsetof(SecretWall{}.OpenDelay), wantOpenDelay},
		{"last open", unsafe.Offsetof(SecretWall{}.LastOpen), wantLastOpen},
		{"player bits", unsafe.Offsetof(SecretWall{}.PlayerBits), wantPlayerBits},
	}
	for _, tc := range tests {
		if tc.got != tc.want {
			t.Errorf("%s offset/size = %d, want %d", tc.name, tc.got, tc.want)
		}
	}
}

func TestWallAttachSecretNativePointer(t *testing.T) {
	secret := &SecretWall{X: 7, Y: 11, State: 3}
	wl := &Wall{}
	wl.AttachSecret(secret, 0x1234)
	if !wl.Flags4.Has(wall.FlagSecret) {
		t.Fatal("secret-wall flag was not set")
	}
	if wl.Field10 != 0x1234 {
		t.Fatalf("wall ID = %#x, want 0x1234", wl.Field10)
	}
	if got := wl.Secret(); got != secret {
		t.Fatalf("secret pointer = %p, want %p", got, secret)
	}
}
