package server

import (
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

func TestUnitAdjustHPClusterNativeLayouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantClass := uintptr(8)
	wantOwner := uintptr(508)
	wantHealth := uintptr(556)
	wantUpdate := uintptr(748)
	wantPlayerUpdateSize := uintptr(556)
	wantPlayerField := uintptr(276)
	wantPlayerSize := uintptr(4828)
	wantPlayerInd := uintptr(2064)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantClass = 12
		wantOwner = 552
		wantHealth = 616
		wantUpdate = 872
		wantPlayerUpdateSize = 656
		wantPlayerField = 336
		wantPlayerSize = 6160
		wantPlayerInd = 2068
	}
	for _, tc := range []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjOwner", unsafe.Offsetof(Object{}.ObjOwner), wantOwner},
		{"Object.HealthData", unsafe.Offsetof(Object{}.HealthData), wantHealth},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"HealthData size", unsafe.Sizeof(HealthData{}), 20},
		{"HealthData.Cur", unsafe.Offsetof(HealthData{}.Cur), 0},
		{"HealthData.Max", unsafe.Offsetof(HealthData{}.Max), 4},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantPlayerUpdateSize},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayerField},
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.PlayerInd", unsafe.Offsetof(Player{}.PlayerInd), wantPlayerInd},
	} {
		if tc.got != tc.want {
			t.Errorf("%s on %s/%s = %d, want %d", tc.name, runtime.GOOS, runtime.GOARCH, tc.got, tc.want)
		}
	}
}

func TestUnitAdjustHP4EE460NativeUsesFieldsAndLiveClass(t *testing.T) {
	health := &HealthData{Cur: 10, Max: 20}
	unit := &Object{HealthData: health}
	var events []string
	unitAdjustHPNative4EE460(unit, 3, unitAdjustHPNativeDeps4EE460{
		gameFlag: func(flag uint32) int32 {
			events = append(events, "flag")
			if flag != uint32(noxflags.GameSuddenDeath) {
				t.Fatalf("flag = %08x, want %08x", flag, uint32(noxflags.GameSuddenDeath))
			}
			return 0
		},
		setHP: func(got *Object, value uint16) {
			events = append(events, "set")
			if got != unit || value != 13 {
				t.Fatalf("set = (%p,%d), want (%p,13)", got, value, unit)
			}
			unit.ObjClass = object.ClassMonster
		},
		informOwnerHP: func(got *Object) {
			events = append(events, "inform")
			if got != unit {
				t.Fatalf("inform object = %p, want %p", got, unit)
			}
		},
	})
	if want := []string{"flag", "set", "inform"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestUnitAdjustHP4EE460ServerBindingUsesSuddenDeathGate(t *testing.T) {
	oldFlags := noxflags.GetGame()
	defer noxflags.SetGame(oldFlags)
	noxflags.SetGame(oldFlags | noxflags.GameSuddenDeath)

	s := &Server{}
	s.UnitAdjustHP4EE460(nil, 1, UnitAdjustHPRuntime4EE460{
		SetHP: func(*Object, uint16) {
			t.Fatal("setter called under SuddenDeath")
		},
	})
}

func TestUnitAdjustHP4EE460NativePreservesNullUnitFault(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected nil unit HealthData panic")
		}
	}()
	unitAdjustHPNative4EE460(nil, 1, unitAdjustHPNativeDeps4EE460{
		gameFlag:      func(uint32) int32 { return 0 },
		setHP:         func(*Object, uint16) {},
		informOwnerHP: func(*Object) {},
	})
}

func TestMobInformOwnerHP4EE4C0NativePreservesIdentityAndIndex(t *testing.T) {
	player := &Player{PlayerInd: 0xfe}
	update := &PlayerUpdateData{Player: player}
	owner := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	obj := &Object{ObjOwner: owner}
	called := false
	mobInformOwnerHPNative4EE4C0(obj, func(index uint8, got *Object) {
		called = true
		if index != player.PlayerInd || got != obj {
			t.Fatalf("report = (%d,%p), want (%d,%p)", index, got, player.PlayerInd, obj)
		}
	})
	if !called {
		t.Fatal("report was not called")
	}
}

func TestMobInformOwnerHP4EE4C0NativePreservesIntermediateNullFaults(t *testing.T) {
	for _, tc := range []struct {
		name   string
		update *PlayerUpdateData
	}{
		{name: "nil-update"},
		{name: "nil-player", update: &PlayerUpdateData{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			owner := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(tc.update)}
			obj := &Object{ObjOwner: owner}
			defer func() {
				if recover() == nil {
					t.Fatal("expected intermediate nil panic")
				}
			}()
			mobInformOwnerHPNative4EE4C0(obj, func(uint8, *Object) {
				t.Fatal("report called after intermediate nil")
			})
		})
	}
}
