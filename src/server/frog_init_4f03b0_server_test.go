package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/prand"
)

func TestFrogInit4F03B0NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantUpdateData := uintptr(748)
	wantDirection2 := uintptr(126)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantUpdateData = 872
		wantDirection2 = 130
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdateData},
		{"Object.Direction2", unsafe.Offsetof(Object{}.Direction2), wantDirection2},
		{"FrogInitUpdateData size", unsafe.Sizeof(FrogInitUpdateData{}), 3},
		{"FrogInitUpdateData.Delay", unsafe.Offsetof(FrogInitUpdateData{}.Delay), 0},
		{"FrogInitUpdateData.Field1", unsafe.Offsetof(FrogInitUpdateData{}.Field1), 1},
		{"FrogInitUpdateData.Field2", unsafe.Offsetof(FrogInitUpdateData{}.Field2), 2},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Fatalf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestFrogInit4F03B0NativeUsesCachedPointerAndExactNarrowing(t *testing.T) {
	type guardedData struct {
		data  FrogInitUpdateData
		guard uint8
	}
	entry := &guardedData{
		data:  FrogInitUpdateData{Delay: 0x11, Field1: 0x22, Field2: 0x33},
		guard: 0xa5,
	}
	live := &guardedData{
		data:  FrogInitUpdateData{Delay: 0x44, Field1: 0x55, Field2: 0x66},
		guard: 0x5a,
	}
	unit := &Object{
		Direction1: 0x1111,
		Direction2: 0x2222,
		UpdateData: unsafe.Pointer(&entry.data),
	}
	randomCall := 0
	got := frogInitNative4F03B0(unit, frogInitNativeDeps4F03B0{
		randomInt: func(minimum, maximum int32, path string, line int32) int32 {
			randomCall++
			switch randomCall {
			case 1:
				if minimum != 55 || maximum != 60 || path != frogInitDelayRandomPath4F03B0 || line != 943 {
					t.Fatalf("delay RNG args = %d/%d/%q/%d", minimum, maximum, path, line)
				}
				unit.UpdateData = unsafe.Pointer(&live.data)
				return 0x1234563b
			case 2:
				if minimum != 0 || maximum != 255 || path != frogInitDirectionRandomPath4F03B0 || line != 947 {
					t.Fatalf("direction RNG args = %d/%d/%q/%d", minimum, maximum, path, line)
				}
				if entry.data != (FrogInitUpdateData{Delay: 0x3b, Field1: 1, Field2: 0}) {
					t.Fatalf("second RNG observed update = %+v", entry.data)
				}
				return 0x1234cdef
			default:
				t.Fatalf("unexpected RNG call %d", randomCall)
				return 0
			}
		},
	})
	if got != 0x1234cdef {
		t.Fatalf("return = %#x, want full RNG result", got)
	}
	if entry.data != (FrogInitUpdateData{Delay: 0x3b, Field1: 1, Field2: 0}) || entry.guard != 0xa5 {
		t.Fatalf("entry record = %+v guard=%#x", entry.data, entry.guard)
	}
	if live.data != (FrogInitUpdateData{Delay: 0x44, Field1: 0x55, Field2: 0x66}) || live.guard != 0x5a {
		t.Fatalf("live record changed = %+v guard=%#x", live.data, live.guard)
	}
	if unit.Direction1 != 0x1111 || unit.Direction2 != 0xcdef {
		t.Fatalf("directions = %#x/%#x", unit.Direction1, unit.Direction2)
	}
	if unit.UpdateData != unsafe.Pointer(&live.data) {
		t.Fatalf("callback UpdateData mutation was lost: %p", unit.UpdateData)
	}
}

func TestFrogInit4F03B0NativeUsesExactLogicRNG(t *testing.T) {
	s := new(Server)
	s.Rand.Logic = prand.New(0)
	wantRandom := prand.New(0)
	wantDelay := wantRandom.IntClamp(55, 60)
	wantDirection := wantRandom.IntClamp(0, 255)
	update := &FrogInitUpdateData{Delay: 0xaa, Field1: 0xbb, Field2: 0xcc}
	unit := &Object{Direction1: 0x1357, Direction2: 0x2468, UpdateData: unsafe.Pointer(update)}

	got := s.FrogInit4F03B0(unit)
	if got != int32(wantDirection) {
		t.Fatalf("return = %d, want %d", got, wantDirection)
	}
	if update.Delay != uint8(wantDelay) || update.Field1 != 1 || update.Field2 != 0 {
		t.Fatalf("update = %+v, want delay %d and bytes 1/0", *update, wantDelay)
	}
	if unit.Direction1 != 0x1357 || unit.Direction2 != Dir16(uint16(wantDirection)) {
		t.Fatalf("directions = %#x/%#x, want Direction2 %#x", unit.Direction1, unit.Direction2, wantDirection)
	}
	if index := s.Rand.Logic.Index(); index != 2 {
		t.Fatalf("logic RNG index = %d, want 2", index)
	}
}

func TestFrogInit4F03B0NativeNilUnitFaultsBeforeRandom(t *testing.T) {
	randomCalls := 0
	defer func() {
		if recover() == nil {
			t.Fatal("nil Object did not preserve the original UpdateData-load fault")
		}
		if randomCalls != 0 {
			t.Fatalf("RNG calls = %d, want 0", randomCalls)
		}
	}()
	frogInitNative4F03B0(nil, frogInitNativeDeps4F03B0{
		randomInt: func(int32, int32, string, int32) int32 {
			randomCalls++
			return 0
		},
	})
}

func TestFrogInit4F03B0NativeNilUpdateFaultsAfterFirstRandom(t *testing.T) {
	randomCalls := 0
	unit := &Object{Direction1: 0x1111, Direction2: 0x2222}
	defer func() {
		if recover() == nil {
			t.Fatal("nil UpdateData did not preserve the original first-byte-store fault")
		}
		if randomCalls != 1 {
			t.Fatalf("RNG calls = %d, want first call only", randomCalls)
		}
		if unit.Direction1 != 0x1111 || unit.Direction2 != 0x2222 {
			t.Fatalf("directions changed to %#x/%#x", unit.Direction1, unit.Direction2)
		}
	}()
	frogInitNative4F03B0(unit, frogInitNativeDeps4F03B0{
		randomInt: func(int32, int32, string, int32) int32 {
			randomCalls++
			return 58
		},
	})
}
