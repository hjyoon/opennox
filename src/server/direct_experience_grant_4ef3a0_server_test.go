package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"
)

func TestDirectExperienceGrantNative4EF3A0BindsFieldsAndServices(t *testing.T) {
	player := &Player{ProtUnitExperience: 0x89abcdef}
	update := &PlayerUpdateData{Player: player}
	unit := &Object{Experience: 10, UpdateData: unsafe.Pointer(update)}
	var events []string
	directExperienceGrantNative4EF3A0(unit, 2.5, directExperienceGrantNativeDeps4EF3A0{
		protectExperience: func(token uint32, award float32) {
			events = append(events, "protect")
			if token != 0x89abcdef || math.Float32bits(award) != 0x40200000 || math.Float32bits(unit.Experience) != 0x41480000 {
				t.Fatalf("protection args/state = %08x/%08x/%08x", token, math.Float32bits(award), math.Float32bits(unit.Experience))
			}
		},
		reportExperience: func(got *Object) {
			events = append(events, "report")
			if got != unit {
				t.Fatalf("report unit = %p, want %p", got, unit)
			}
		},
		loadString: func(key, path string, line int) string {
			events = append(events, "string")
			if key != "health.c:gainpoints" || path != `C:\NoxPost\src\Server\GameMech\explevel.c` || line != 381 {
				t.Fatalf("string args = %q/%q/%d", key, path, line)
			}
			return "native-gain-message"
		},
		sendLineMessage: func(got *Object, message string, points uint32) {
			events = append(events, "line")
			if got != unit || message != "native-gain-message" || points != 2 {
				t.Fatalf("line args = %p/%q/%08x", got, message, points)
			}
		},
		syncLevel: func(got *Object) {
			events = append(events, "sync")
			if got != unit {
				t.Fatalf("sync unit = %p, want %p", got, unit)
			}
			unit.Experience = 777
		},
	})
	want := []string{"protect", "report", "string", "line", "sync"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
	if unit.Experience != 777 {
		t.Fatalf("sync mutation was lost: %v", unit.Experience)
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update)
	runtime.KeepAlive(player)
}

func TestDirectExperienceGrantNative4EF3A0HasNoClassOrAwardGate(t *testing.T) {
	for _, award := range []float32{0, -1, math.Float32frombits(0x7fc12345)} {
		player := &Player{ProtUnitExperience: 7}
		update := &PlayerUpdateData{Player: player}
		unit := &Object{ObjClass: 0, Experience: 1, UpdateData: unsafe.Pointer(update)}
		called := 0
		directExperienceGrantNative4EF3A0(unit, award, directExperienceGrantNativeDeps4EF3A0{
			protectExperience: func(uint32, float32) { called++ },
			reportExperience:  func(*Object) { called++ },
			loadString: func(string, string, int) string {
				called++
				return "message"
			},
			sendLineMessage: func(*Object, string, uint32) { called++ },
			syncLevel:       func(*Object) { called++ },
		})
		if called != 5 {
			t.Fatalf("award %08x service calls = %d, want 5", math.Float32bits(award), called)
		}
		runtime.KeepAlive(unit)
		runtime.KeepAlive(update)
		runtime.KeepAlive(player)
	}
}

func TestDirectExperienceGrant4EF3A0NativeLayouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantExperience := uintptr(28)
	wantUpdate := uintptr(748)
	wantUpdateSize := uintptr(556)
	wantPlayer := uintptr(276)
	wantPlayerSize := uintptr(4828)
	wantProtection := uintptr(4604)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantExperience = 32
		wantUpdate = 872
		wantUpdateSize = 656
		wantPlayer = 336
		wantPlayerSize = 6160
		wantProtection = 5908
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.Experience", unsafe.Offsetof(Object{}.Experience), wantExperience},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantUpdateSize},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayer},
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.ProtUnitExperience", unsafe.Offsetof(Player{}.ProtUnitExperience), wantProtection},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}
