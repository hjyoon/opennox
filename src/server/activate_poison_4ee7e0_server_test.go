package server

import (
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
)

func defaultActivatePoisonNativeDeps4EE7E0() activatePoisonNativeDeps4EE7E0 {
	return activatePoisonNativeDeps4EE7E0{
		poisonProtection: func(*Object) float64 { return 0 },
		randomInt:        func(int32, int32, string, int32) int32 { return 100 },
		priorityMessage:  func(*Object, string, uint8) {},
		setPoison: func(unit *Object, value int32) {
			unit.Poison540 = uint8(value)
		},
		audio: func(uint32, *Object, int32, uint32) {},
		frame: func() uint32 { return 0 },
	}
}

func TestActivatePoison4EE7E0NativeLayouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantClass := uintptr(8)
	wantSubClass := uintptr(12)
	wantFlags := uintptr(16)
	wantPoison := uintptr(540)
	wantHealth := uintptr(556)
	wantUpdate := uintptr(748)
	wantUpdateSize := uintptr(556)
	wantPlayer := uintptr(276)
	wantPlayerSize := uintptr(4828)
	wantPlayerFlags := uintptr(3680)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantClass = 12
		wantSubClass = 16
		wantFlags = 20
		wantPoison = 600
		wantHealth = 616
		wantUpdate = 872
		wantUpdateSize = 640
		wantPlayer = 320
		wantPlayerSize = 6160
		wantPlayerFlags = 4976
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjSubClass", unsafe.Offsetof(Object{}.ObjSubClass), wantSubClass},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.Poison540", unsafe.Offsetof(Object{}.Poison540), wantPoison},
		{"Object.HealthData", unsafe.Offsetof(Object{}.HealthData), wantHealth},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantUpdateSize},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayer},
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.Field3680", unsafe.Offsetof(Player{}.Field3680), wantPlayerFlags},
		{"HealthData size", unsafe.Sizeof(HealthData{}), 20},
		{"HealthData.Field16", unsafe.Offsetof(HealthData{}.Field16), 16},
		{"poison argument width", unsafe.Sizeof(int32(0)), 4},
		{"poison storage width", unsafe.Sizeof(Object{}.Poison540), 1},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestActivatePoisonNative4EE7E0BindsEffectsAndLiveHealth(t *testing.T) {
	health := new(HealthData)
	player := new(Player)
	update := &PlayerUpdateData{Player: player}
	unit := &Object{
		ObjClass:   object.ClassPlayer,
		HealthData: health,
		UpdateData: unsafe.Pointer(update),
	}
	var events []string
	deps := defaultActivatePoisonNativeDeps4EE7E0()
	deps.poisonProtection = func(got *Object) float64 {
		events = append(events, "protection")
		if got != unit {
			t.Fatalf("protection object = %p, want %p", got, unit)
		}
		return 0.25
	}
	deps.randomInt = func(minimum, maximum int32, path string, line int32) int32 {
		events = append(events, "random")
		if minimum != 0 || maximum != 100 || path != activatePoisonRandomPath4EE7E0 || line != 361 {
			t.Fatalf("random args = (%d,%d,%q,%d)", minimum, maximum, path, line)
		}
		return 25
	}
	deps.setPoison = func(got *Object, value int32) {
		events = append(events, "set")
		if got != unit || value != 2 {
			t.Fatalf("set poison = (%p,%d), want (%p,2)", got, value, unit)
		}
		got.Poison540 = uint8(value)
	}
	deps.audio = func(id uint32, got *Object, kind int32, code uint32) {
		events = append(events, "audio")
		if id != 100 || got != unit || kind != 0 || code != 0 {
			t.Fatalf("audio args = (%d,%p,%d,%d)", id, got, kind, code)
		}
	}
	deps.frame = func() uint32 {
		events = append(events, "frame")
		return 77
	}

	if got := activatePoisonNative4EE7E0(unit, 2, 3, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(events, []string{"protection", "random", "set", "audio", "frame"}) {
		t.Fatalf("events = %q", events)
	}
	if unit.Poison540 != 2 || health.Field16 != 77 {
		t.Fatalf("poison/frame = %d/%d, want 2/77", unit.Poison540, health.Field16)
	}
}

func TestActivatePoisonNative4EE7E0PlayerAndMonsterGates(t *testing.T) {
	failDeps := defaultActivatePoisonNativeDeps4EE7E0()
	failDeps.poisonProtection = func(*Object) float64 {
		t.Fatal("protection reached after class gate")
		return 0
	}

	player := &Player{Field3680: 1}
	playerUnit := &Object{
		ObjClass:   object.ClassPlayer,
		UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: player}),
	}
	if got := activatePoisonNative4EE7E0(playerUnit, 1, 10, failDeps); got != 0 {
		t.Fatalf("observer result = %d, want 0", got)
	}

	monster := &Object{ObjClass: object.ClassMonster, ObjSubClass: object.SubClass(0x200)}
	if got := activatePoisonNative4EE7E0(monster, 1, 10, failDeps); got != 0 {
		t.Fatalf("immune monster result = %d, want 0", got)
	}

	buffed := &Object{Buffs: uint32(1) << activatePoisonBlockingEnchant4EE7E0}
	if got := activatePoisonNative4EE7E0(buffed, 1, 10, failDeps); got != 0 {
		t.Fatalf("buffed result = %d, want 0", got)
	}
}

func TestActivatePoisonNative4EE7E0NilFaultsBeforeNominalGate(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil unit did not fault on the entry poison load")
		}
	}()
	activatePoisonNative4EE7E0(nil, 1, 10, defaultActivatePoisonNativeDeps4EE7E0())
}

func TestActivatePoison4EE7E0ServerBindsLogicRNG(t *testing.T) {
	s := new(Server)
	s.Rand.Logic = prand.New(0)
	marker := unsafe.Pointer(new(byte))
	modifier := &ModifierEff{Engage112: marker, EngageFloat120: 0.5}
	init := &ModifierInitData{Modifiers: [4]*ModifierEff{modifier}}
	unit := &Object{
		ObjClass:  object.Class(poisonProtectionClassMask4E0040),
		Poison540: 9,
		InitData:  unsafe.Pointer(init),
	}
	got := s.ActivatePoison4EE7E0(unit, 0, 9, ActivatePoisonRuntime4EE7E0{
		PoisonProtectEngage: marker,
		PoisonState: PoisonStateRuntime4EE8F0{
			NeedPlayerStatus:  func(*Player, uint32) { t.Fatal("unchanged poison reached need status") },
			UnsetPlayerStatus: func(*Player, uint32) { t.Fatal("unchanged poison reached unset status") },
			ReportPoison:      func(*Object, *Object, int32) { t.Fatal("unchanged poison reached report") },
		},
	})
	if got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if index := s.Rand.Logic.Index(); index != 1 {
		t.Fatalf("logic RNG index = %d, want 1", index)
	}
}

func TestActivatePoison4EE7E0ServerUsesRestoredSetter(t *testing.T) {
	s := new(Server)
	s.Rand.Logic = prand.New(0)
	s.SetFrame(91)
	health := new(HealthData)
	unit := &Object{HealthData: health}
	runtime := ActivatePoisonRuntime4EE7E0{
		PoisonState: PoisonStateRuntime4EE8F0{
			NeedPlayerStatus:  func(*Player, uint32) { t.Fatal("need status reached for unclassified object") },
			UnsetPlayerStatus: func(*Player, uint32) { t.Fatal("unset status reached for unclassified object") },
			ReportPoison:      func(*Object, *Object, int32) { t.Fatal("report reached for unclassified object") },
		},
	}
	if got := s.ActivatePoison4EE7E0(unit, 2, 3, runtime); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if unit.Poison540 != 2 || unit.Field542 != 1000 || health.Field16 != 91 {
		t.Fatalf("poison/timer/frame = %d/%d/%d, want 2/1000/91", unit.Poison540, unit.Field542, health.Field16)
	}
}
