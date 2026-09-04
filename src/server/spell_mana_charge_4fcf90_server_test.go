package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

func TestSpellManaChargeNative4FCF90PreservesPointersAndCachedUpdate(t *testing.T) {
	update1 := &PlayerUpdateData{ManaCur: 20}
	update2 := &PlayerUpdateData{ManaCur: 0}
	unit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update1)}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]uintptr{
			"unit":    uintptr(unsafe.Pointer(unit)),
			"update1": uintptr(unsafe.Pointer(update1)),
			"update2": uintptr(unsafe.Pointer(update2)),
		} {
			if pointer <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, pointer)
			}
		}
	}

	subtractCalls := 0
	got := spellManaChargeNative4FCF90(unit, 74, math.MinInt32, spellManaChargeNativeDeps4FCF90{
		loadGodMode: func() bool {
			unit.UpdateData = unsafe.Pointer(update2)
			return false
		},
		summonCost: func(int32, *Object) int32 {
			t.Fatal("ordinary spell used summon cost")
			return 0
		},
		spellManaCost: func(spellID, costType int32) int32 {
			if spellID != 74 || costType != math.MinInt32 {
				t.Fatalf("ordinary cost args = %d/%d, want 74/INT32_MIN", spellID, costType)
			}
			return 7
		},
		subtractMana: func(gotUnit *Object, cost int32) {
			subtractCalls++
			if gotUnit != unit || cost != 7 {
				t.Fatalf("subtract args = %p/%d, want %p/7", gotUnit, cost, unit)
			}
		},
		loadTickRate: func() uint32 {
			t.Fatal("sufficient mana loaded tick rate")
			return 0
		},
	})
	if got != 7 || subtractCalls != 1 {
		t.Fatalf("result/subtract calls = %d/%d, want 7/1", got, subtractCalls)
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update1)
	runtime.KeepAlive(update2)
}

func TestSpellManaChargeNative4FCF90InsufficientSummonStoresCachedUpdate(t *testing.T) {
	update1 := &PlayerUpdateData{ManaCur: 9, Field20_0: 1, Field20_1: 2}
	update2 := &PlayerUpdateData{ManaCur: 100, Field20_0: 3, Field20_1: 4}
	unit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update1)}
	ordinaryCalls := 0

	got := spellManaChargeNative4FCF90(unit, 75, 2, spellManaChargeNativeDeps4FCF90{
		loadGodMode: func() bool { return false },
		summonCost: func(spellID int32, gotUnit *Object) int32 {
			if spellID != 75 || gotUnit != unit {
				t.Fatalf("summon cost args = %d/%p, want 75/%p", spellID, gotUnit, unit)
			}
			unit.UpdateData = unsafe.Pointer(update2)
			return 10
		},
		spellManaCost: func(spellID, costType int32) int32 {
			ordinaryCalls++
			if spellID != 75 || costType != 1 {
				t.Fatalf("recharge cost args = %d/%d, want 75/1", spellID, costType)
			}
			return 0x12345
		},
		subtractMana: func(*Object, int32) {
			t.Fatal("insufficient mana subtracted mana")
		},
		loadTickRate: func() uint32 {
			if update1.Field20_0 != 0x2345 || update1.Field20_1 != 2 {
				t.Fatalf("state before tick = %#x/%#x, want 0x2345/2", update1.Field20_0, update1.Field20_1)
			}
			return 0xffff8001
		},
	})
	if got != -1 || ordinaryCalls != 1 {
		t.Fatalf("result/ordinary calls = %d/%d, want -1/1", got, ordinaryCalls)
	}
	if update1.Field20_0 != 0x2345 || update1.Field20_1 != 0x8001 {
		t.Fatalf("cached update recharge = %#x/%#x, want 0x2345/0x8001", update1.Field20_0, update1.Field20_1)
	}
	if update2.Field20_0 != 3 || update2.Field20_1 != 4 {
		t.Fatalf("replacement update changed: %#x/%#x", update2.Field20_0, update2.Field20_1)
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update1)
	runtime.KeepAlive(update2)
}

func TestSpellManaChargeNative4FCF90HasOriginalFaultAndGateSemantics(t *testing.T) {
	forbidden := spellManaChargeNativeDeps4FCF90{
		loadGodMode: func() bool {
			t.Fatal("GodMode read across entry gate")
			return false
		},
		summonCost: func(int32, *Object) int32 {
			t.Fatal("summon cost across entry gate")
			return 0
		},
		spellManaCost: func(int32, int32) int32 {
			t.Fatal("ordinary cost across entry gate")
			return 0
		},
		subtractMana: func(*Object, int32) {
			t.Fatal("mana subtraction across entry gate")
		},
		loadTickRate: func() uint32 {
			t.Fatal("tick rate read across entry gate")
			return 0
		},
	}

	t.Run("nil unit faults at class load", func(t *testing.T) {
		var recovered any
		func() {
			defer func() { recovered = recover() }()
			_ = spellManaChargeNative4FCF90(nil, 74, 2, forbidden)
		}()
		if recovered == nil {
			t.Fatal("nil unit did not fault")
		}
	})

	t.Run("non-Player may cache nil update without dereferencing it", func(t *testing.T) {
		unit := &Object{ObjClass: object.ClassMonster}
		if got := spellManaChargeNative4FCF90(unit, 74, 2, forbidden); got != -1 {
			t.Fatalf("result = %d, want canonical -1", got)
		}
	})

	t.Run("zero spell may cache nil update before returning", func(t *testing.T) {
		unit := &Object{ObjClass: object.ClassPlayer}
		if got := spellManaChargeNative4FCF90(unit, 0, 2, forbidden); got != -1 {
			t.Fatalf("result = %d, want canonical -1", got)
		}
	})

	t.Run("nil Player update faults after initial cost callback", func(t *testing.T) {
		unit := &Object{ObjClass: object.ClassPlayer}
		costCalls := 0
		deps := spellManaChargeNativeDeps4FCF90{
			loadGodMode: func() bool { return false },
			summonCost: func(int32, *Object) int32 {
				t.Fatal("ordinary spell used summon cost")
				return 0
			},
			spellManaCost: func(spellID, costType int32) int32 {
				costCalls++
				if spellID != 74 || costType != 2 {
					t.Fatalf("cost args = %d/%d, want 74/2", spellID, costType)
				}
				return 1
			},
			subtractMana: func(*Object, int32) { t.Fatal("subtract after nil update") },
			loadTickRate: func() uint32 {
				t.Fatal("tick after nil update")
				return 0
			},
		}
		var recovered any
		func() {
			defer func() { recovered = recover() }()
			_ = spellManaChargeNative4FCF90(unit, 74, 2, deps)
		}()
		if recovered == nil || costCalls != 1 {
			t.Fatalf("recovered/cost calls = %#v/%d, want fault/1", recovered, costCalls)
		}
	})
}

func TestSpellManaCharge4FCF90ServerMethodUsesEngineGodMode(t *testing.T) {
	oldEngine := noxflags.GetEngine()
	noxflags.ResetEngine()
	t.Cleanup(func() {
		noxflags.ResetEngine()
		noxflags.SetEngine(oldEngine)
	})

	unit := &Object{ObjClass: object.ClassPlayer}
	noxflags.SetEngine(noxflags.EngineGodMode)
	if got := new(Server).SpellManaCharge4FCF90(unit, 74, 2, func(*Object, int32) {
		t.Fatal("GodMode subtracted mana")
	}); got != 0 {
		t.Fatalf("GodMode result = %d, want canonical 0", got)
	}
}

func TestSpellManaCharge4FCF90ServerSummonCostRechecksLivePlayerClass(t *testing.T) {
	deps := spellManaChargeServerDeps4FCF90(new(Server), func(*Object, int32) {
		t.Fatal("summon cost subtracted mana")
	})

	if got := deps.summonCost(75, nil); got != 0 {
		t.Fatalf("nil unit summon cost = %d, want 0", got)
	}
	unit := &Object{ObjClass: object.ClassPlayer}
	unit.ObjClass = object.ClassMonster
	if got := deps.summonCost(75, unit); got != 0 {
		t.Fatalf("live non-Player summon cost = %d, want 0", got)
	}
}

func TestSpellManaCharge4FCF90NativeLayouts(t *testing.T) {
	wantClass := uintptr(8)
	wantUpdate := uintptr(748)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantClass = 12
		wantUpdate = 872
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"PlayerUpdateData.ManaCur", unsafe.Offsetof(PlayerUpdateData{}.ManaCur), 4},
		{"PlayerUpdateData.Field20_0", unsafe.Offsetof(PlayerUpdateData{}.Field20_0), 80},
		{"PlayerUpdateData.Field20_1", unsafe.Offsetof(PlayerUpdateData{}.Field20_1), 82},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}
