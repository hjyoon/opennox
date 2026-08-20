package server

import (
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

func defaultUnitDamageClearNativeDeps4EE5E0() unitDamageClearNativeDeps4EE5E0 {
	return unitDamageClearNativeDeps4EE5E0{
		engineFlag:    func(uint32) int32 { return 0 },
		breakHarpoon:  func(*Object) {},
		setHP:         func(*Object, uint16) {},
		buffOff:       func(*Object, EnchantID) {},
		isZombie:      func(*Object) bool { return false },
		soloReward:    func(*Object) {},
		monsterDie:    func(*Object) {},
		callDeath:     func(unsafe.Pointer, *Object) {},
		delayedDelete: func(*Object) {},
		informOwnerHP: func(*Object) {},
	}
}

func TestUnitDamageClear4EE5E0NativeLayouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantClass := uintptr(8)
	wantFlags := uintptr(16)
	wantHealth := uintptr(556)
	wantDeath := uintptr(724)
	wantUpdate := uintptr(748)
	wantUpdateSize := uintptr(556)
	wantHarpoon := uintptr(132)
	wantPlayer := uintptr(276)
	wantPlayerSize := uintptr(4828)
	wantPlayerClass := uintptr(2251)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantClass = 12
		wantFlags = 20
		wantHealth = 616
		wantDeath = 824
		wantUpdate = 872
		wantUpdateSize = 640
		wantHarpoon = 152
		wantPlayer = 320
		wantPlayerSize = 6160
		wantPlayerClass = 2255
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.HealthData", unsafe.Offsetof(Object{}.HealthData), wantHealth},
		{"Object.Death", unsafe.Offsetof(Object{}.Death), wantDeath},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"HealthData size", unsafe.Sizeof(HealthData{}), 20},
		{"HealthData.Cur", unsafe.Offsetof(HealthData{}.Cur), 0},
		{"HealthData.Max", unsafe.Offsetof(HealthData{}.Max), 4},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantUpdateSize},
		{"PlayerUpdateData.HarpoonTarg", unsafe.Offsetof(PlayerUpdateData{}.HarpoonTarg), wantHarpoon},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayer},
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{
			"Player.Info.playerClass",
			unsafe.Offsetof(Player{}.info) + unsafe.Offsetof(PlayerInfo{}.playerClass),
			wantPlayerClass,
		},
		{"damage width", unsafe.Sizeof(int32(0)), 4},
		{"HP width", unsafe.Sizeof(uint16(0)), 2},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestUnitDamageClearNative4EE5E0BindsHarpoonAndLiveHealth(t *testing.T) {
	entryHealth := &HealthData{Cur: 40, Max: 100}
	liveHealth := &HealthData{Cur: 30, Max: 70}
	harpoon := &Object{}
	pl := &Player{}
	pl.Info().SetPlayerClass(player.Warrior)
	update := &PlayerUpdateData{Player: pl, HarpoonTarg: harpoon}
	unit := &Object{
		ObjClass:   object.ClassPlayer,
		HealthData: entryHealth,
		UpdateData: unsafe.Pointer(update),
	}
	var events []string
	deps := defaultUnitDamageClearNativeDeps4EE5E0()
	deps.engineFlag = func(flag uint32) int32 {
		events = append(events, "engine")
		if flag != uint32(noxflags.EngineGodMode) {
			t.Fatalf("engine flag = %#x, want %#x", flag, noxflags.EngineGodMode)
		}
		return 0
	}
	deps.breakHarpoon = func(got *Object) {
		events = append(events, "break")
		if got != unit || update.Player != pl || update.HarpoonTarg != harpoon {
			t.Fatalf("harpoon binding lost native identity")
		}
		unit.HealthData = liveHealth
		unit.ObjClass = object.ClassMonster
	}
	deps.setHP = func(got *Object, value uint16) {
		events = append(events, "set")
		if got != unit || value != 23 {
			t.Fatalf("set HP = (%p,%d), want (%p,23)", got, value, unit)
		}
		got.HealthData.Cur = value
	}
	deps.informOwnerHP = func(got *Object) {
		events = append(events, "inform")
		if got != unit {
			t.Fatalf("inform object = %p, want %p", got, unit)
		}
	}
	unitDamageClearNative4EE5E0(unit, 7, deps)
	if !reflect.DeepEqual(events, []string{"engine", "break", "set", "inform"}) {
		t.Fatalf("events = %q, want [engine break set inform]", events)
	}
	if liveHealth.Cur != 23 || entryHealth.Cur != 40 {
		t.Fatalf("live/entry HP = %d/%d, want 23/40", liveHealth.Cur, entryHealth.Cur)
	}
}

func TestUnitDamageClearNative4EE5E0LethalServicesAndLiveClass(t *testing.T) {
	health := &HealthData{Cur: 5, Max: 10}
	unit := &Object{HealthData: health, ObjFlags: object.Flags(0x40)}
	var events []string
	deps := defaultUnitDamageClearNativeDeps4EE5E0()
	deps.setHP = func(got *Object, value uint16) {
		events = append(events, "set")
		if got != unit || value != 0 {
			t.Fatalf("set HP = (%p,%d), want (%p,0)", got, value, unit)
		}
		health.Cur = value
	}
	deps.buffOff = func(got *Object, enchant EnchantID) {
		events = append(events, "buff")
		if got != unit || enchant != ENCHANT_DEATH {
			t.Fatalf("buff off = (%p,%d), want (%p,%d)", got, enchant, unit, ENCHANT_DEATH)
		}
	}
	deps.isZombie = func(got *Object) bool {
		events = append(events, "zombie")
		return got != unit
	}
	deps.soloReward = func(got *Object) {
		events = append(events, "reward")
		if got != unit {
			t.Fatalf("reward object = %p, want %p", got, unit)
		}
		unit.ObjClass = object.ClassMonster
	}
	deps.monsterDie = func(got *Object) {
		events = append(events, "die")
		if got != unit {
			t.Fatalf("die object = %p, want %p", got, unit)
		}
		unit.ObjClass = 0
	}
	deps.callDeath = func(unsafe.Pointer, *Object) { t.Fatal("Monster path called Death") }
	deps.delayedDelete = func(*Object) { t.Fatal("Monster path delayed deletion") }
	deps.informOwnerHP = func(*Object) { t.Fatal("post-die non-Monster reported HP") }
	unitDamageClearNative4EE5E0(unit, 5, deps)
	if !reflect.DeepEqual(events, []string{"set", "buff", "zombie", "reward", "die"}) {
		t.Fatalf("events = %q, want [set buff zombie reward die]", events)
	}
	if uint32(unit.ObjFlags) != 0x8040 {
		t.Fatalf("flags = %#x, want 0x8040", unit.ObjFlags)
	}
}

func TestUnitDamageClearNative4EE5E0DeathPointerAndFinalClass(t *testing.T) {
	token := new(uint32)
	health := &HealthData{Cur: 1, Max: 10}
	unit := &Object{HealthData: health, Death: unsafe.Pointer(token)}
	deps := defaultUnitDamageClearNativeDeps4EE5E0()
	deps.setHP = func(*Object, uint16) {}
	deps.isZombie = func(*Object) bool { return true }
	var events []string
	deps.callDeath = func(death unsafe.Pointer, got *Object) {
		events = append(events, "death")
		if death != unsafe.Pointer(token) || got != unit {
			t.Fatalf("death = (%p,%p), want (%p,%p)", death, got, token, unit)
		}
		unit.ObjClass = object.ClassMonster
	}
	deps.delayedDelete = func(*Object) { t.Fatal("non-nil Death used fallback") }
	deps.informOwnerHP = func(got *Object) {
		events = append(events, "inform")
		if got != unit {
			t.Fatalf("inform object = %p, want %p", got, unit)
		}
	}
	unitDamageClearNative4EE5E0(unit, 1, deps)
	if !reflect.DeepEqual(events, []string{"death", "inform"}) {
		t.Fatalf("events = %q, want [death inform]", events)
	}
}

func TestUnitDamageClearNative4EE5E0PreservesNilPlayerFault(t *testing.T) {
	health := &HealthData{Cur: 5, Max: 10}
	update := &PlayerUpdateData{}
	unit := &Object{
		ObjClass:   object.ClassPlayer,
		HealthData: health,
		UpdateData: unsafe.Pointer(update),
	}
	defer func() {
		if recover() == nil {
			t.Fatal("nil Player did not preserve GAME.EXE fault")
		}
	}()
	unitDamageClearNative4EE5E0(unit, 1, defaultUnitDamageClearNativeDeps4EE5E0())
}

func TestUnitDamageClear4EE5E0ServerBindingUsesGodMode(t *testing.T) {
	oldEngine := noxflags.GetEngine()
	defer func() {
		noxflags.ResetEngine()
		noxflags.SetEngine(oldEngine)
	}()
	noxflags.ResetEngine()
	noxflags.SetEngine(noxflags.EngineGodMode)

	unit := &Object{ObjClass: object.ClassPlayer, HealthData: &HealthData{Cur: 5, Max: 10}}
	s := &Server{}
	s.UnitDamageClear4EE5E0(unit, 1, UnitDamageClearRuntime4EE5E0{
		BreakHarpoon:  func(*Object) { t.Fatal("GodMode broke harpoon") },
		SetHP:         func(*Object, uint16) { t.Fatal("GodMode changed HP") },
		BuffOff:       func(*Object, EnchantID) { t.Fatal("GodMode removed buff") },
		SoloReward:    func(*Object) { t.Fatal("GodMode rewarded kill") },
		MonsterDie:    func(*Object) { t.Fatal("GodMode killed Monster") },
		DelayedDelete: func(*Object) { t.Fatal("GodMode deleted unit") },
	})
	if unit.HealthData.Cur != 5 || uint32(unit.ObjFlags) != 0 {
		t.Fatalf("GodMode mutated unit: HP=%d flags=%#x", unit.HealthData.Cur, unit.ObjFlags)
	}
}
