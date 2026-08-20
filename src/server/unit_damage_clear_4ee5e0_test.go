package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type unitDamageClearTestHealth4EE5E0 struct {
	name    string
	current uint16
	maximum uint16
}

type unitDamageClearTestPlayer4EE5E0 struct {
	name  string
	class uint8
}

type unitDamageClearTestUpdate4EE5E0 struct {
	name    string
	player  *unitDamageClearTestPlayer4EE5E0
	harpoon *unitDamageClearTestObject4EE5E0
}

type unitDamageClearTestObject4EE5E0 struct {
	name   string
	class  uint8
	flags  uint32
	health *unitDamageClearTestHealth4EE5E0
	update *unitDamageClearTestUpdate4EE5E0
	death  int
}

type unitDamageClearTestWorld4EE5E0 struct {
	unit        *unitDamageClearTestObject4EE5E0
	damage      int32
	engine      int32
	zombie      int32
	events      []string
	classLoads  int
	afterClass  func(*unitDamageClearTestWorld4EE5E0, int)
	afterBreak  func(*unitDamageClearTestWorld4EE5E0)
	afterSetHP  func(*unitDamageClearTestWorld4EE5E0)
	afterBuff   func(*unitDamageClearTestWorld4EE5E0)
	afterZombie func(*unitDamageClearTestWorld4EE5E0)
	afterReward func(*unitDamageClearTestWorld4EE5E0)
	afterDie    func(*unitDamageClearTestWorld4EE5E0)
	afterDeath  func(*unitDamageClearTestWorld4EE5E0)
	afterDelete func(*unitDamageClearTestWorld4EE5E0)
}

func unitDamageClearObjectName4EE5E0(obj *unitDamageClearTestObject4EE5E0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func unitDamageClearHealthName4EE5E0(health *unitDamageClearTestHealth4EE5E0) string {
	if health == nil {
		return "nil"
	}
	return health.name
}

func unitDamageClearUpdateName4EE5E0(update *unitDamageClearTestUpdate4EE5E0) string {
	if update == nil {
		return "nil"
	}
	return update.name
}

func unitDamageClearPlayerName4EE5E0(player *unitDamageClearTestPlayer4EE5E0) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func (w *unitDamageClearTestWorld4EE5E0) event(format string, args ...any) {
	w.events = append(w.events, fmt.Sprintf(format, args...))
}

func (w *unitDamageClearTestWorld4EE5E0) hooks() unitDamageClearHooks4EE5E0[
	*unitDamageClearTestObject4EE5E0,
	*unitDamageClearTestHealth4EE5E0,
	*unitDamageClearTestUpdate4EE5E0,
	*unitDamageClearTestPlayer4EE5E0,
	int,
] {
	return unitDamageClearHooks4EE5E0[
		*unitDamageClearTestObject4EE5E0,
		*unitDamageClearTestHealth4EE5E0,
		*unitDamageClearTestUpdate4EE5E0,
		*unitDamageClearTestPlayer4EE5E0,
		int,
	]{
		loadUnitArg: func() *unitDamageClearTestObject4EE5E0 {
			w.event("arg:%s", unitDamageClearObjectName4EE5E0(w.unit))
			return w.unit
		},
		loadHealth: func(obj *unitDamageClearTestObject4EE5E0) *unitDamageClearTestHealth4EE5E0 {
			w.event("health:%s=%s", unitDamageClearObjectName4EE5E0(obj), unitDamageClearHealthName4EE5E0(obj.health))
			return obj.health
		},
		loadMaximum: func(health *unitDamageClearTestHealth4EE5E0) uint16 {
			w.event("max:%s=%d", unitDamageClearHealthName4EE5E0(health), health.maximum)
			return health.maximum
		},
		engineFlag: func(flag uint32) int32 {
			w.event("engine:%#x=%d", flag, w.engine)
			return w.engine
		},
		loadClassLow: func(obj *unitDamageClearTestObject4EE5E0) uint8 {
			value := obj.class
			w.classLoads++
			w.event("class:%s=%#x", unitDamageClearObjectName4EE5E0(obj), value)
			if w.afterClass != nil {
				w.afterClass(w, w.classLoads)
			}
			return value
		},
		loadUpdateData: func(obj *unitDamageClearTestObject4EE5E0) *unitDamageClearTestUpdate4EE5E0 {
			w.event("update:%s=%s", unitDamageClearObjectName4EE5E0(obj), unitDamageClearUpdateName4EE5E0(obj.update))
			return obj.update
		},
		loadPlayer: func(update *unitDamageClearTestUpdate4EE5E0) *unitDamageClearTestPlayer4EE5E0 {
			w.event("player:%s=%s", unitDamageClearUpdateName4EE5E0(update), unitDamageClearPlayerName4EE5E0(update.player))
			return update.player
		},
		loadPlayerClass: func(player *unitDamageClearTestPlayer4EE5E0) uint8 {
			w.event("playerClass:%s", unitDamageClearPlayerName4EE5E0(player))
			return player.class
		},
		loadHarpoonTarget: func(update *unitDamageClearTestUpdate4EE5E0) *unitDamageClearTestObject4EE5E0 {
			w.event("harpoon:%s=%s", unitDamageClearUpdateName4EE5E0(update), unitDamageClearObjectName4EE5E0(update.harpoon))
			return update.harpoon
		},
		breakHarpoon: func(obj *unitDamageClearTestObject4EE5E0) {
			w.event("break:%s", unitDamageClearObjectName4EE5E0(obj))
			if w.afterBreak != nil {
				w.afterBreak(w)
			}
		},
		loadDamageArg: func() int32 {
			w.event("damage=%d", w.damage)
			return w.damage
		},
		loadCurrent: func(health *unitDamageClearTestHealth4EE5E0) uint16 {
			w.event("cur:%s", unitDamageClearHealthName4EE5E0(health))
			return health.current
		},
		setHP: func(obj *unitDamageClearTestObject4EE5E0, value uint16) {
			w.event("set:%s=%d", unitDamageClearObjectName4EE5E0(obj), value)
			obj.health.current = value
			if w.afterSetHP != nil {
				w.afterSetHP(w)
			}
		},
		loadFlags: func(obj *unitDamageClearTestObject4EE5E0) uint32 {
			w.event("flags:%s=%#x", unitDamageClearObjectName4EE5E0(obj), obj.flags)
			return obj.flags
		},
		storeFlags: func(obj *unitDamageClearTestObject4EE5E0, value uint32) {
			w.event("flags:%s<-%#x", unitDamageClearObjectName4EE5E0(obj), value)
			obj.flags = value
		},
		buffOff: func(obj *unitDamageClearTestObject4EE5E0, buff int32) {
			w.event("buff:%s=%d", unitDamageClearObjectName4EE5E0(obj), buff)
			if w.afterBuff != nil {
				w.afterBuff(w)
			}
		},
		isZombie: func(obj *unitDamageClearTestObject4EE5E0) int32 {
			value := w.zombie
			w.event("zombie:%s=%d", unitDamageClearObjectName4EE5E0(obj), value)
			if w.afterZombie != nil {
				w.afterZombie(w)
			}
			return value
		},
		soloReward: func(obj *unitDamageClearTestObject4EE5E0) {
			w.event("reward:%s", unitDamageClearObjectName4EE5E0(obj))
			if w.afterReward != nil {
				w.afterReward(w)
			}
		},
		monsterDie: func(obj *unitDamageClearTestObject4EE5E0) {
			w.event("monsterDie:%s", unitDamageClearObjectName4EE5E0(obj))
			if w.afterDie != nil {
				w.afterDie(w)
			}
		},
		loadDeath: func(obj *unitDamageClearTestObject4EE5E0) int {
			w.event("death:%s=%d", unitDamageClearObjectName4EE5E0(obj), obj.death)
			return obj.death
		},
		callDeath: func(death int, obj *unitDamageClearTestObject4EE5E0) {
			w.event("callDeath:%d:%s", death, unitDamageClearObjectName4EE5E0(obj))
			if w.afterDeath != nil {
				w.afterDeath(w)
			}
		},
		delayedDelete: func(obj *unitDamageClearTestObject4EE5E0) {
			w.event("delete:%s", unitDamageClearObjectName4EE5E0(obj))
			if w.afterDelete != nil {
				w.afterDelete(w)
			}
		},
		informOwnerHP: func(obj *unitDamageClearTestObject4EE5E0) {
			w.event("inform:%s", unitDamageClearObjectName4EE5E0(obj))
		},
	}
}

func newUnitDamageClearWorld4EE5E0() *unitDamageClearTestWorld4EE5E0 {
	health := &unitDamageClearTestHealth4EE5E0{name: "entry", current: 20, maximum: 100}
	return &unitDamageClearTestWorld4EE5E0{
		unit:   &unitDamageClearTestObject4EE5E0{name: "unit", health: health},
		damage: 3,
	}
}

func assertUnitDamageClearEvents4EE5E0(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("events:\n got: %q\nwant: %q", got, want)
	}
}

func TestUnitDamageClear4EE5E0EntryGatesAndGodMode(t *testing.T) {
	tests := []struct {
		name string
		edit func(*unitDamageClearTestWorld4EE5E0)
		want []string
	}{
		{
			name: "nil unit",
			edit: func(w *unitDamageClearTestWorld4EE5E0) { w.unit = nil },
			want: []string{"arg:nil"},
		},
		{
			name: "nil health",
			edit: func(w *unitDamageClearTestWorld4EE5E0) { w.unit.health = nil },
			want: []string{"arg:unit", "health:unit=nil"},
		},
		{
			name: "zero maximum",
			edit: func(w *unitDamageClearTestWorld4EE5E0) { w.unit.health.maximum = 0 },
			want: []string{"arg:unit", "health:unit=entry", "max:entry=0"},
		},
		{
			name: "god mode player",
			edit: func(w *unitDamageClearTestWorld4EE5E0) {
				w.engine = 1
				w.unit.class = unitDamageClearPlayerBit4EE5E0
			},
			want: []string{"arg:unit", "health:unit=entry", "max:entry=100", "engine:0x20=1", "class:unit=0x4"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newUnitDamageClearWorld4EE5E0()
			tc.edit(w)
			unitDamageClear4EE5E0(w.hooks())
			assertUnitDamageClearEvents4EE5E0(t, w.events, tc.want)
		})
	}
}

func TestUnitDamageClear4EE5E0GodModeNonPlayerReloadsClass(t *testing.T) {
	w := newUnitDamageClearWorld4EE5E0()
	w.engine = 1
	w.afterClass = func(w *unitDamageClearTestWorld4EE5E0, load int) {
		if load == 1 {
			w.unit.class = unitDamageClearPlayerBit4EE5E0
		}
	}
	unitDamageClear4EE5E0(w.hooks())
	assertUnitDamageClearEvents4EE5E0(t, w.events, []string{
		"arg:unit", "health:unit=entry", "max:entry=100", "engine:0x20=1",
		"class:unit=0x0", "class:unit=0x4", "update:unit=nil",
		"health:unit=entry", "damage=3", "cur:entry", "set:unit=17", "class:unit=0x4",
	})
}

func TestUnitDamageClear4EE5E0HarpoonUsesCachedUpdateAndLiveHealth(t *testing.T) {
	w := newUnitDamageClearWorld4EE5E0()
	live := &unitDamageClearTestHealth4EE5E0{name: "live", current: 33, maximum: 90}
	target := &unitDamageClearTestObject4EE5E0{name: "target"}
	w.unit.class = unitDamageClearPlayerBit4EE5E0
	w.unit.update = &unitDamageClearTestUpdate4EE5E0{
		name: "update", player: &unitDamageClearTestPlayer4EE5E0{name: "player"}, harpoon: target,
	}
	w.damage = 5
	w.afterBreak = func(w *unitDamageClearTestWorld4EE5E0) {
		w.unit.health = live
		w.unit.update = &unitDamageClearTestUpdate4EE5E0{name: "replacement"}
		w.unit.class = unitDamageClearMonsterBit4EE5E0
	}
	unitDamageClear4EE5E0(w.hooks())
	assertUnitDamageClearEvents4EE5E0(t, w.events, []string{
		"arg:unit", "health:unit=entry", "max:entry=100", "engine:0x20=0", "class:unit=0x4",
		"update:unit=update", "player:update=player", "playerClass:player", "harpoon:update=target", "break:unit",
		"health:unit=live", "damage=5", "cur:live", "set:unit=28", "class:unit=0x2", "inform:unit",
	})
	if live.current != 28 {
		t.Fatalf("live HP = %d, want 28", live.current)
	}
}

func TestUnitDamageClear4EE5E0PlayerHarpoonShortCircuitAndFaults(t *testing.T) {
	t.Run("class skips target", func(t *testing.T) {
		w := newUnitDamageClearWorld4EE5E0()
		w.unit.class = unitDamageClearPlayerBit4EE5E0
		w.unit.update = &unitDamageClearTestUpdate4EE5E0{
			name: "update", player: &unitDamageClearTestPlayer4EE5E0{name: "player", class: 2},
		}
		unitDamageClear4EE5E0(w.hooks())
		for _, event := range w.events {
			if event == "harpoon:update=nil" || event == "break:unit" {
				t.Fatalf("player class gate touched harpoon: %q", w.events)
			}
		}
	})

	t.Run("nil player faults", func(t *testing.T) {
		w := newUnitDamageClearWorld4EE5E0()
		w.unit.class = unitDamageClearPlayerBit4EE5E0
		w.unit.update = &unitDamageClearTestUpdate4EE5E0{name: "update"}
		defer func() {
			if recover() == nil {
				t.Fatal("nil Player did not preserve the original fault")
			}
			assertUnitDamageClearEvents4EE5E0(t, w.events, []string{
				"arg:unit", "health:unit=entry", "max:entry=100", "engine:0x20=0", "class:unit=0x4",
				"update:unit=update", "player:update=nil", "playerClass:nil",
			})
		}()
		unitDamageClear4EE5E0(w.hooks())
	})

	t.Run("reloaded nil health faults after damage load", func(t *testing.T) {
		w := newUnitDamageClearWorld4EE5E0()
		w.unit.class = unitDamageClearPlayerBit4EE5E0
		w.unit.update = &unitDamageClearTestUpdate4EE5E0{
			name: "update", player: &unitDamageClearTestPlayer4EE5E0{name: "player"},
			harpoon: &unitDamageClearTestObject4EE5E0{name: "target"},
		}
		w.afterBreak = func(w *unitDamageClearTestWorld4EE5E0) { w.unit.health = nil }
		defer func() {
			if recover() == nil {
				t.Fatal("nil live HealthData did not preserve the original fault")
			}
			wantTail := []string{"break:unit", "health:unit=nil", "damage=3", "cur:nil"}
			if got := w.events[len(w.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
				t.Fatalf("fault tail = %q, want %q", got, wantTail)
			}
		}()
		unitDamageClear4EE5E0(w.hooks())
	})
}

func TestUnitDamageClear4EE5E0SignedDamageAndLowWord(t *testing.T) {
	tests := []struct {
		name    string
		current uint16
		damage  int32
		want    uint16
	}{
		{name: "ordinary", current: 100, damage: 99, want: 1},
		{name: "negative heals", current: 10, damage: -1, want: 11},
		{name: "zero heals", current: 0, damage: -1, want: 1},
		{name: "low word wraps", current: math.MaxUint16, damage: -1, want: 0},
		{name: "minimum int wraps", current: math.MaxUint16, damage: math.MinInt32, want: math.MaxUint16},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newUnitDamageClearWorld4EE5E0()
			w.unit.health.current = tc.current
			w.damage = tc.damage
			unitDamageClear4EE5E0(w.hooks())
			if got := w.unit.health.current; got != tc.want {
				t.Fatalf("HP = %d, want %d (current=%d damage=%d)", got, tc.want, tc.current, tc.damage)
			}
			for _, event := range w.events {
				if len(event) >= 6 && event[:6] == "flags" {
					t.Fatalf("nonlethal damage entered death path: %q", w.events)
				}
			}
		})
	}
}

func TestUnitDamageClear4EE5E0LethalCallbacksUseLiveState(t *testing.T) {
	w := newUnitDamageClearWorld4EE5E0()
	w.unit.health.current = 5
	w.damage = 5
	w.afterSetHP = func(w *unitDamageClearTestWorld4EE5E0) {
		w.unit.flags = 0x40
	}
	w.afterReward = func(w *unitDamageClearTestWorld4EE5E0) {
		w.unit.class = unitDamageClearMonsterBit4EE5E0
	}
	w.afterDie = func(w *unitDamageClearTestWorld4EE5E0) {
		w.unit.class = 0
	}
	unitDamageClear4EE5E0(w.hooks())
	assertUnitDamageClearEvents4EE5E0(t, w.events, []string{
		"arg:unit", "health:unit=entry", "max:entry=100", "engine:0x20=0", "class:unit=0x0",
		"health:unit=entry", "damage=5", "cur:entry", "set:unit=0", "flags:unit=0x40", "flags:unit<-0x8040",
		"buff:unit=16", "zombie:unit=0", "reward:unit", "class:unit=0x2", "monsterDie:unit", "class:unit=0x0",
	})
	if w.unit.flags != 0x8040 {
		t.Fatalf("flags = %#x, want 0x8040", w.unit.flags)
	}
}

func TestUnitDamageClear4EE5E0NilDeathUsesDelayedDeleteThenLiveClass(t *testing.T) {
	w := newUnitDamageClearWorld4EE5E0()
	w.unit.health.current = 2
	w.damage = 2
	w.afterDelete = func(w *unitDamageClearTestWorld4EE5E0) {
		w.unit.class = unitDamageClearMonsterBit4EE5E0
	}
	unitDamageClear4EE5E0(w.hooks())
	assertUnitDamageClearEvents4EE5E0(t, w.events, []string{
		"arg:unit", "health:unit=entry", "max:entry=100", "engine:0x20=0", "class:unit=0x0",
		"health:unit=entry", "damage=2", "cur:entry", "set:unit=0", "flags:unit=0x0", "flags:unit<-0x8000",
		"buff:unit=16", "zombie:unit=0", "reward:unit", "class:unit=0x0", "death:unit=0", "delete:unit",
		"class:unit=0x2", "inform:unit",
	})
}

func TestUnitDamageClear4EE5E0ZombieDeathCallbackCanEnableFinalReport(t *testing.T) {
	w := newUnitDamageClearWorld4EE5E0()
	w.unit.health.current = 5
	w.unit.death = 7
	w.damage = 5
	w.zombie = 1
	w.afterDeath = func(w *unitDamageClearTestWorld4EE5E0) {
		w.unit.class = unitDamageClearMonsterBit4EE5E0
	}
	unitDamageClear4EE5E0(w.hooks())
	assertUnitDamageClearEvents4EE5E0(t, w.events, []string{
		"arg:unit", "health:unit=entry", "max:entry=100", "engine:0x20=0", "class:unit=0x0",
		"health:unit=entry", "damage=5", "cur:entry", "set:unit=0", "flags:unit=0x0", "flags:unit<-0x8000",
		"buff:unit=16", "zombie:unit=1", "class:unit=0x0", "death:unit=7", "callDeath:7:unit",
		"class:unit=0x2", "inform:unit",
	})
}

func TestUnitDamageClear4EE5E0ExistingDeadSkipsDeathSideEffects(t *testing.T) {
	w := newUnitDamageClearWorld4EE5E0()
	w.unit.health.current = 1
	w.unit.class = unitDamageClearMonsterBit4EE5E0
	w.unit.flags = unitDamageClearDeadFlag4EE5E0 | 0x20
	w.damage = math.MaxInt32
	unitDamageClear4EE5E0(w.hooks())
	assertUnitDamageClearEvents4EE5E0(t, w.events, []string{
		"arg:unit", "health:unit=entry", "max:entry=100", "engine:0x20=0", "class:unit=0x2",
		"health:unit=entry", fmt.Sprintf("damage=%d", int32(math.MaxInt32)), "cur:entry", "set:unit=0",
		"flags:unit=0x8020", "class:unit=0x2", "inform:unit",
	})
}
