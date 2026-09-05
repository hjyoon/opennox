package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

const (
	collisionEnchantTestSource4FDF90        = uint64(0x1_0000_1111)
	collisionEnchantTestInitialTarget4FDF90 = uint64(0x2_0000_2222)
	collisionEnchantTestTarget4FDF90        = uint64(0x3_0000_3333)
)

type collisionEnchantTestState4FDF90 struct {
	t *testing.T

	sourceArg uint64
	targetArg uint64
	hasShock  int32

	targetFlags    uint32
	targetClassLow uint8
	targetClass    uint32
	sourceClassLow uint8
	isEnemyResult  int32
	sameTeamResult int32
	shockPower     uint8
	balanceValue   float64
	convertedValue int32

	flagLoads        int
	targetClassLoads int
	invisibleOffs    int
	damageCalls      int
	events           []string
	faultAt          int
}

func newCollisionEnchantTestState4FDF90(t *testing.T) *collisionEnchantTestState4FDF90 {
	return &collisionEnchantTestState4FDF90{
		t:              t,
		sourceArg:      collisionEnchantTestSource4FDF90,
		targetArg:      collisionEnchantTestInitialTarget4FDF90,
		hasShock:       math.MinInt32,
		targetClassLow: 0x86,
		isEnemyResult:  math.MinInt32,
		shockPower:     7,
		convertedValue: math.MinInt32,
	}
}

func (s *collisionEnchantTestState4FDF90) event(name string) {
	s.events = append(s.events, name)
	if s.faultAt != 0 && len(s.events) == s.faultAt {
		panic(name)
	}
}

func (s *collisionEnchantTestState4FDF90) hooks() collisionEnchantHooks4FDF90[uint64] {
	return collisionEnchantHooks4FDF90[uint64]{
		loadSourceArg: func() uint64 {
			s.event("source-arg")
			return s.sourceArg
		},
		hasEnchant: func(source uint64, enchant uint32) int32 {
			s.event("has-shock")
			if source != collisionEnchantTestSource4FDF90 || enchant != collisionEnchantShock4FDF90 {
				s.t.Fatalf("has-enchant args = %#x/%d", source, enchant)
			}
			s.targetArg = collisionEnchantTestTarget4FDF90
			return s.hasShock
		},
		loadTargetArg: func() uint64 {
			s.event("target-arg")
			return s.targetArg
		},
		loadTargetFlags: func(target uint64) uint32 {
			s.flagLoads++
			s.event(fmt.Sprintf("target-flags:%d", s.flagLoads))
			if target != collisionEnchantTestTarget4FDF90 {
				s.t.Fatalf("flags target = %#x", target)
			}
			return s.targetFlags
		},
		loadTargetClassLow: func(target uint64) uint8 {
			s.event("target-class-low")
			if target != collisionEnchantTestTarget4FDF90 {
				s.t.Fatalf("class-low target = %#x", target)
			}
			return s.targetClassLow
		},
		isEnemy: func(source, target uint64) int32 {
			s.event("enemy")
			if source != collisionEnchantTestSource4FDF90 || target != collisionEnchantTestTarget4FDF90 {
				s.t.Fatalf("enemy args = %#x/%#x", source, target)
			}
			s.shockPower = 0
			return s.isEnemyResult
		},
		loadShockPower: func(source uint64) uint8 {
			s.event("shock-power")
			if source != collisionEnchantTestSource4FDF90 {
				s.t.Fatalf("power source = %#x", source)
			}
			return s.shockPower
		},
		audio: func(id int32, source uint64, kind int32, code uint32) {
			s.event("audio")
			if id != collisionEnchantShockAudio4FDF90 || source != collisionEnchantTestSource4FDF90 || kind != 0 || code != 0 {
				s.t.Fatalf("audio args = %d/%#x/%d/%d", id, source, kind, code)
			}
		},
		disableEnchant: func(source uint64, enchant uint32) {
			s.event(fmt.Sprintf("disable:%d", enchant))
			if source != collisionEnchantTestSource4FDF90 {
				s.t.Fatalf("disable source = %#x", source)
			}
			switch enchant {
			case collisionEnchantShock4FDF90:
				s.balanceValue = 16_777_217
			case collisionEnchantInvisible4FDF90:
				s.invisibleOffs++
				if s.invisibleOffs == 1 {
					s.sourceClassLow = 0x84
					s.targetClass = collisionEnchantWallClass4FDF90
				}
			default:
				s.t.Fatalf("unexpected enchant %d", enchant)
			}
		},
		balanceFloatTable: func(key string, power int32) float64 {
			s.event("balance")
			if key != collisionEnchantShockDamageBalance4FDF90 || power != -1 {
				s.t.Fatalf("balance args = %q/%d", key, power)
			}
			return s.balanceValue
		},
		floatToInt: func(value float32) int32 {
			s.event("float-to-int")
			if got := math.Float32bits(value); got != math.Float32bits(16_777_216) {
				s.t.Fatalf("spilled damage bits = %08x", got)
			}
			return s.convertedValue
		},
		callTargetDamage: func(target, source, weapon uint64, damage int32, damageType uint32) int32 {
			s.event("damage")
			if target != collisionEnchantTestTarget4FDF90 || source != collisionEnchantTestSource4FDF90 ||
				weapon != source || damage != math.MinInt32 || damageType != collisionEnchantShockDamageType4FDF90 {
				s.t.Fatalf("damage args = %#x/%#x/%#x/%d/%d", target, source, weapon, damage, damageType)
			}
			s.damageCalls++
			s.targetClass = collisionEnchantUnitOrWallClass4FDF90
			return math.MinInt32
		},
		loadTargetClass: func(target uint64) uint32 {
			s.targetClassLoads++
			s.event(fmt.Sprintf("target-class:%d", s.targetClassLoads))
			if target != collisionEnchantTestTarget4FDF90 {
				s.t.Fatalf("class target = %#x", target)
			}
			return s.targetClass
		},
		unitsOnSameTeam: func(target, source uint64) int32 {
			s.event("same-team")
			if target != collisionEnchantTestTarget4FDF90 || source != collisionEnchantTestSource4FDF90 {
				s.t.Fatalf("same-team args = %#x/%#x", target, source)
			}
			return s.sameTeamResult
		},
		loadSourceClassLow: func(source uint64) uint8 {
			s.event("source-class-low")
			if source != collisionEnchantTestSource4FDF90 {
				s.t.Fatalf("source class object = %#x", source)
			}
			return s.sourceClassLow
		},
	}
}

var collisionEnchantExactEvents4FDF90 = []string{
	"source-arg", "has-shock", "target-arg",
	"target-flags:1", "target-class-low", "enemy", "shock-power",
	"audio", "disable:22", "balance", "float-to-int", "damage",
	"target-class:1", "target-flags:2", "same-team", "disable:0",
	"source-class-low", "target-class:2", "target-flags:3", "disable:0",
}

func assertCollisionEnchantEvents4FDF90(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("events:\n got  %#v\n want %#v", got, want)
	}
}

func TestCollisionEnchant4FDF90ExactOrderWidthsAndLiveReloads(t *testing.T) {
	state := newCollisionEnchantTestState4FDF90(t)
	collisionEnchant4FDF90(state.hooks())
	assertCollisionEnchantEvents4FDF90(t, state.events, collisionEnchantExactEvents4FDF90)
	if state.damageCalls != 1 || state.invisibleOffs != 2 {
		t.Fatalf("damage/invisibility calls = %d/%d, want 1/2", state.damageCalls, state.invisibleOffs)
	}
}

func TestCollisionEnchant4FDF90ExactFaultPrefixes(t *testing.T) {
	for faultAt := 1; faultAt <= len(collisionEnchantExactEvents4FDF90); faultAt++ {
		t.Run(fmt.Sprintf("%02d_%s", faultAt, collisionEnchantExactEvents4FDF90[faultAt-1]), func(t *testing.T) {
			state := newCollisionEnchantTestState4FDF90(t)
			state.faultAt = faultAt
			panicked := false
			func() {
				defer func() { panicked = recover() != nil }()
				collisionEnchant4FDF90(state.hooks())
			}()
			if !panicked {
				t.Fatal("expected injected fault")
			}
			assertCollisionEnchantEvents4FDF90(t, state.events, collisionEnchantExactEvents4FDF90[:faultAt])
		})
	}
}

func TestCollisionEnchant4FDF90CapturesTargetAfterShockPredicateEvenWhenFalse(t *testing.T) {
	state := newCollisionEnchantTestState4FDF90(t)
	state.hasShock = 0
	hooks := state.hooks()
	hooks.loadTargetFlags = func(uint64) uint32 { t.Fatal("zero Shock read target flags"); return 0 }
	hooks.loadTargetClassLow = func(uint64) uint8 { t.Fatal("zero Shock read low class"); return 0 }
	hooks.isEnemy = func(uint64, uint64) int32 { t.Fatal("zero Shock tested enemy"); return 0 }
	collisionEnchant4FDF90(hooks)
	assertCollisionEnchantEvents4FDF90(t, state.events, []string{
		"source-arg", "has-shock", "target-arg", "target-class:1", "source-class-low",
	})
}

func TestCollisionEnchant4FDF90ShockGatesAreOrdered(t *testing.T) {
	t.Run("flags precede class", func(t *testing.T) {
		state := newCollisionEnchantTestState4FDF90(t)
		state.targetFlags = collisionEnchantShockRejectFlags4FDF90
		hooks := state.hooks()
		hooks.loadTargetClassLow = func(uint64) uint8 { t.Fatal("blocked flags read class"); return 0 }
		collisionEnchant4FDF90(hooks)
		assertCollisionEnchantEvents4FDF90(t, state.events, []string{
			"source-arg", "has-shock", "target-arg", "target-flags:1", "target-class:1", "source-class-low",
		})
	})

	t.Run("class precedes enemy", func(t *testing.T) {
		state := newCollisionEnchantTestState4FDF90(t)
		state.targetClassLow = 0xf9
		hooks := state.hooks()
		hooks.isEnemy = func(uint64, uint64) int32 { t.Fatal("blocked class tested enemy"); return 0 }
		collisionEnchant4FDF90(hooks)
		assertCollisionEnchantEvents4FDF90(t, state.events, []string{
			"source-arg", "has-shock", "target-arg", "target-flags:1", "target-class-low",
			"target-class:1", "source-class-low",
		})
	})

	t.Run("enemy precedes power", func(t *testing.T) {
		state := newCollisionEnchantTestState4FDF90(t)
		state.isEnemyResult = 0
		hooks := state.hooks()
		hooks.loadShockPower = func(uint64) uint8 { t.Fatal("friendly target read Shock power"); return 0 }
		collisionEnchant4FDF90(hooks)
		assertCollisionEnchantEvents4FDF90(t, state.events, []string{
			"source-arg", "has-shock", "target-arg", "target-flags:1", "target-class-low", "enemy",
			"target-class:1", "source-class-low",
		})
	})
}

func TestCollisionEnchant4FDF90ShockPowerZeroExtendsThenDecrements(t *testing.T) {
	for _, tc := range []struct {
		power uint8
		want  int32
	}{{0, -1}, {1, 0}, {255, 254}} {
		t.Run(fmt.Sprintf("power_%d", tc.power), func(t *testing.T) {
			state := newCollisionEnchantTestState4FDF90(t)
			state.shockPower = tc.power
			hooks := state.hooks()
			hooks.isEnemy = func(uint64, uint64) int32 {
				state.event("enemy")
				return -1
			}
			hooks.balanceFloatTable = func(key string, power int32) float64 {
				state.event("balance")
				if key != collisionEnchantShockDamageBalance4FDF90 || power != tc.want {
					t.Fatalf("balance args = %q/%d, want power %d", key, power, tc.want)
				}
				return 16_777_217
			}
			collisionEnchant4FDF90(hooks)
		})
	}
}

func TestCollisionEnchant4FDF90LaterPhasesRemainIndependentAndOrdered(t *testing.T) {
	t.Run("inactive target stops before same-team", func(t *testing.T) {
		state := newCollisionEnchantTestState4FDF90(t)
		state.hasShock = 0
		state.targetClass = collisionEnchantUnitOrWallClass4FDF90
		state.targetFlags = collisionEnchantInactiveFlags4FDF90
		hooks := state.hooks()
		hooks.unitsOnSameTeam = func(uint64, uint64) int32 { t.Fatal("inactive target tested team"); return 0 }
		collisionEnchant4FDF90(hooks)
		assertCollisionEnchantEvents4FDF90(t, state.events, []string{
			"source-arg", "has-shock", "target-arg", "target-class:1", "target-flags:1", "source-class-low",
		})
	})

	t.Run("any nonzero same-team result suppresses first removal but later fields are live", func(t *testing.T) {
		for _, sameTeam := range []int32{-1, 1, math.MinInt32, math.MaxInt32} {
			state := newCollisionEnchantTestState4FDF90(t)
			state.hasShock = 0
			state.targetClass = collisionEnchantUnitOrWallClass4FDF90
			state.sameTeamResult = sameTeam
			hooks := state.hooks()
			hooks.unitsOnSameTeam = func(target, source uint64) int32 {
				state.event("same-team")
				state.sourceClassLow = collisionEnchantPlayerClassLow4FDF90
				state.targetClass = collisionEnchantWallClass4FDF90
				return sameTeam
			}
			collisionEnchant4FDF90(hooks)
			if state.invisibleOffs != 1 {
				t.Fatalf("same-team %d removals = %d, want one later wall removal", sameTeam, state.invisibleOffs)
			}
		}
	})
}
