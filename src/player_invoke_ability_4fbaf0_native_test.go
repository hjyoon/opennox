package opennox

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"strconv"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

func TestPlayerInvokeAbilityNative4FBAF0Layout(t *testing.T) {
	wantFlagsOffset := uintptr(16)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantFlagsOffset = 20
	}
	if got := unsafe.Offsetof(server.Object{}.ObjFlags); got != wantFlagsOffset {
		t.Fatalf("Object.ObjFlags offset = %d, want %d", got, wantFlagsOffset)
	}
	if got := unsafe.Sizeof(server.Ability(0)); got != 4 {
		t.Fatalf("Ability size = %d, want 4", got)
	}
}

func TestPlayerInvokeAbilityNative4FBAF0PreservesPointerDispatchAndDuration(t *testing.T) {
	unit := &server.Object{ObjFlags: object.FlagSelected | object.FlagAirborne}
	wantDuration := int32(math.MinInt32 + 0x123456)
	sourceDuration := int(wantDuration)
	if strconv.IntSize == 64 {
		sourceDuration = int((int64(3) << 32) | int64(uint32(wantDuration)))
	}

	var events []string
	checkUnit := func(name string, got *server.Object) {
		if got != unit {
			t.Fatalf("%s unit = %p, want %p", name, got, unit)
		}
		events = append(events, name)
	}
	deps := playerInvokeAbilityNativeDeps4FBAF0{
		berserk: func(got *server.Object) {
			checkUnit("berserk", got)
		},
		warcry: func(got *server.Object) {
			checkUnit("warcry", got)
		},
		harpoon: func(got *server.Object) {
			checkUnit("harpoon", got)
		},
		loadDuration: func(ability server.Ability) int {
			events = append(events, fmt.Sprintf("duration:%d", ability))
			return sourceDuration
		},
		treadLightly: func(got *server.Object, duration int) {
			checkUnit("tread-lightly", got)
			if duration != int(wantDuration) {
				t.Fatalf("Tread Lightly duration = %d, want signed PE32 %d", duration, wantDuration)
			}
		},
		infravis: func(got *server.Object, duration int) {
			checkUnit("infravis", got)
			if duration != int(wantDuration) {
				t.Fatalf("Infravision duration = %d, want signed PE32 %d", duration, wantDuration)
			}
		},
	}

	cases := []struct {
		ability server.Ability
		want    []string
	}{
		{server.Ability(math.MinInt32), nil},
		{server.AbilityInvalid, nil},
		{server.AbilityBerserk, []string{"berserk"}},
		{server.AbilityWarcry, []string{"warcry"}},
		{server.AbilityHarpoon, []string{"harpoon"}},
		{server.AbilityTreadLightly, []string{"duration:4", "tread-lightly"}},
		{server.AbilityInfravis, []string{"duration:5", "infravis"}},
		{server.AbilityMax, nil},
		{server.Ability(math.MaxInt32), nil},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("ability-%08x", uint32(tc.ability)), func(t *testing.T) {
			events = nil
			playerInvokeAbilityNative4FBAF0(unit, tc.ability, deps)
			if !reflect.DeepEqual(events, tc.want) {
				t.Fatalf("events = %v, want %v", events, tc.want)
			}
		})
	}

	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %#x, want native address above 4 GiB", uintptr(unsafe.Pointer(unit)))
	}
	runtime.KeepAlive(unit)
}

func TestPlayerInvokeAbilityNative4FBAF0DeadDestroyedAndNilSemantics(t *testing.T) {
	called := false
	deps := playerInvokeAbilityNativeDeps4FBAF0{
		berserk: func(*server.Object) { called = true },
		warcry:  func(*server.Object) { called = true },
		harpoon: func(*server.Object) { called = true },
		loadDuration: func(server.Ability) int {
			called = true
			return 0
		},
		treadLightly: func(*server.Object, int) { called = true },
		infravis:     func(*server.Object, int) { called = true },
	}
	for _, flags := range []object.Flags{
		object.FlagDestroyed,
		object.FlagDead,
		object.FlagDestroyed | object.FlagDead,
	} {
		unit := &server.Object{ObjFlags: flags}
		playerInvokeAbilityNative4FBAF0(unit, server.AbilityTreadLightly, deps)
		if called {
			t.Fatalf("flags %#x invoked a callback", uint32(flags))
		}
	}

	defer func() {
		if recover() == nil {
			t.Fatal("nil unit did not fault at direct ObjFlags read")
		}
		if called {
			t.Fatal("nil unit invoked a callback after the flag fault")
		}
	}()
	playerInvokeAbilityNative4FBAF0(nil, server.AbilityBerserk, deps)
}
