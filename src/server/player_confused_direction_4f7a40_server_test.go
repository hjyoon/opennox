package server

import (
	"testing"
	"unsafe"
)

func TestPlayerConfusedDirectionNativeLayout4F7A40(t *testing.T) {
	wantNetCode := uintptr(36)
	wantDirection2 := uintptr(126)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantNetCode = 40
		wantDirection2 = 130
	}
	for _, tc := range []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.NetCode", unsafe.Offsetof(Object{}.NetCode), wantNetCode},
		{"Object.Direction2", unsafe.Offsetof(Object{}.Direction2), wantDirection2},
		{"Object.BuffsPower element", unsafe.Sizeof(Object{}.BuffsPower[0]), 1},
	} {
		if tc.got != tc.want {
			t.Errorf("%s = %d, want %d", tc.name, tc.got, tc.want)
		}
	}
}

func TestPlayerConfusedDirectionNative4F7A40(t *testing.T) {
	s := new(Server)
	s.SetFrame(^uint32(0))
	unit := &Object{Direction2: Dir16(0xffff), NetCode: 1}
	unit.BuffsPower[ENCHANT_CONFUSED] = 2

	got := s.PlayerConfusedDirection4F7A40(unit)
	want := Dir16(playerConfusedDirectionReference4F7A40(0xffff, 2, ^uint32(0), 1))
	if got != want {
		t.Fatalf("direction = %d, want %d", got, want)
	}
}

func TestPlayerConfusedDirectionNative4F7A40NilUnitFaults(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil unit did not fault on the leading Direction2 load")
		}
	}()
	new(Server).PlayerConfusedDirection4F7A40(nil)
}
