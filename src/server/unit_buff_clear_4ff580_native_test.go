package server

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestUnitBuffClear4FF580NativeLayout(t *testing.T) {
	wantBuffs, wantDuration, wantPower := uintptr(340), uintptr(344), uintptr(408)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantBuffs, wantDuration, wantPower = 344, 348, 412
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Buffs offset", unsafe.Offsetof(Object{}.Buffs), wantBuffs},
		{"Buffs width", unsafe.Sizeof(Object{}.Buffs), 4},
		{"BuffsDur offset", unsafe.Offsetof(Object{}.BuffsDur), wantDuration},
		{"BuffsDur width", unsafe.Sizeof(Object{}.BuffsDur), 64},
		{"BuffsDur element width", unsafe.Sizeof(Object{}.BuffsDur[0]), 2},
		{"BuffsPower offset", unsafe.Offsetof(Object{}.BuffsPower), wantPower},
		{"BuffsPower width", unsafe.Sizeof(Object{}.BuffsPower), 32},
		{"BuffsPower element width", unsafe.Sizeof(Object{}.BuffsPower[0]), 1},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("Object %s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestUnitBuffClearNative4FF580PreservesPointerAndAccessorOrder(t *testing.T) {
	unit, freeUnit := alloc.New(Object{})
	t.Cleanup(freeUnit)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	var events []string
	deps := unitBuffClearNativeDeps4FF580{
		loadUnitArg: func(got *Object) *Object {
			events = append(events, "unit")
			if got != unit {
				t.Fatalf("unit = %p, want full native identity %p", got, unit)
			}
			return got
		},
		setBuffFlags: func(got *Object, flags uint32) {
			events = append(events, "set")
			if got != unit || flags != 0 {
				t.Fatalf("SetBuffFlags = (%p, %#x), want (%p, 0)", got, flags, unit)
			}
		},
		storeDuration: func(got *Object, buff int32, value uint16) {
			events = append(events, fmt.Sprintf("duration-%d", buff))
			if got != unit || value != 0 {
				t.Fatalf("StoreDuration(%d) = (%p, %#x), want (%p, 0)", buff, got, value, unit)
			}
		},
		storePower: func(got *Object, buff int32, value uint8) {
			events = append(events, fmt.Sprintf("power-%d", buff))
			if got != unit || value != 0 {
				t.Fatalf("StorePower(%d) = (%p, %#x), want (%p, 0)", buff, got, value, unit)
			}
		},
	}
	unitBuffClearNative4FF580(unit, deps)

	want := []string{"unit", "set"}
	for buff := int32(0); buff < unitBuffClearSlots4FF580; buff++ {
		want = append(want, fmt.Sprintf("duration-%d", buff), fmt.Sprintf("power-%d", buff))
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want exact native accessor order %q", events, want)
	}
	runtime.KeepAlive(unit)
}

func TestUnitBuffClear4FF580ServerBinding(t *testing.T) {
	unit, freeUnit := alloc.New(Object{})
	t.Cleanup(freeUnit)
	unit.ObjClass = object.ClassClientPersist
	unit.Buffs = math.MaxUint32
	for i := range unit.BuffsDur {
		unit.BuffsDur[i] = uint16(0x8000 + i)
		unit.BuffsPower[i] = uint8(0x80 + i)
		unit.Field140[i] = 0xa5000080 | uint32(i<<12)
	}

	unit.UnitBuffClear4FF580(UnitBuffClearRuntime4FF580{})
	if unit.Buffs != 0 || unit.BuffsDur != [32]uint16{} || unit.BuffsPower != [32]uint8{} {
		t.Fatalf("native buff state was not cleared: flags=%#x durations=%#v powers=%#v", unit.Buffs, unit.BuffsDur, unit.BuffsPower)
	}
	if unit.Field38 != math.MaxUint32 {
		t.Fatalf("sync marker = %#08x, want %#08x", unit.Field38, uint32(math.MaxUint32))
	}
	for i, value := range unit.Field140 {
		want := (uint32(0xa5000080)|uint32(i<<12))&0xfffff000 | 0x800000
		if value != want {
			t.Fatalf("Field140[%d] = %#08x, want %#08x", i, value, want)
		}
	}
	runtime.KeepAlive(unit)
}

func TestUnitBuffClear4FF580PlayerProtectionBoundary(t *testing.T) {
	player := &Player{ProtUnitBuffs: 0x13579bdf}
	update := &PlayerUpdateData{Player: player}
	unit := &Object{
		ObjClass:   object.ClassPlayer,
		UpdateData: unsafe.Pointer(update),
		Buffs:      math.MaxUint32,
	}
	for i := range unit.BuffsDur {
		unit.BuffsDur[i] = uint16(0x8000 + i)
		unit.BuffsPower[i] = uint8(0x80 + i)
		unit.Field140[i] = 0xa5000080 | uint32(i)
	}
	beforeDuration, beforePower, beforeSync := unit.BuffsDur, unit.BuffsPower, unit.Field140

	calls := 0
	unit.UnitBuffClear4FF580(UnitBuffClearRuntime4FF580{
		ResetPlayerProtection: func(got *Player, flags uint32) {
			calls++
			if got != player || flags != 0 {
				t.Fatalf("protection callback = (%p, %#x), want (%p, 0)", got, flags, player)
			}
			if unit.Field38 != math.MaxUint32 || unit.Buffs != 0 {
				t.Fatal("protection callback ran before the flag/sync writes")
			}
			if unit.BuffsDur != beforeDuration || unit.BuffsPower != beforePower || unit.Field140 != beforeSync {
				t.Fatal("duration, power, or per-player sync stores preceded the protection callback")
			}
		},
	})
	if calls != 1 {
		t.Fatalf("protection callback count = %d, want 1", calls)
	}
	if unit.BuffsDur != [32]uint16{} || unit.BuffsPower != [32]uint8{} {
		t.Fatal("player buff arrays were not cleared after the protection callback")
	}
	for i, value := range unit.Field140 {
		want := beforeSync[i]&0xfffff000 | 0x800000
		if value != want {
			t.Fatalf("Field140[%d] = %#08x, want %#08x", i, value, want)
		}
	}
}

func TestUnitBuffClear4FF580NativeNilObjectFaults(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil native object returned instead of faulting")
		}
	}()
	(*Object)(nil).UnitBuffClear4FF580(UnitBuffClearRuntime4FF580{})
}
