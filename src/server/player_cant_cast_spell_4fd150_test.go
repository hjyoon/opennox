package server

import (
	"fmt"
	"reflect"
	"testing"
)

type playerCantCastTestObject4FD150 struct {
	typeInd uint16
	flags   uint32
	owned   *playerCantCastTestObject4FD150
	next    *playerCantCastTestObject4FD150
	inv     *playerCantCastTestObject4FD150
	invNext *playerCantCastTestObject4FD150
}

type playerCantCastTestRuntime4FD150 struct {
	events        []string
	gameFlags     uint32
	crownCache    uint32
	gameBallCache uint32
	lookup        map[string]uint32
	spellFlags    int32
	team          int32
	enchant       int32
	allowed       int32
	types         map[playerCantCastSummonType4FD150]uint32
	counts        map[uint32]int32
	power         int32
	balance       map[string]float64
	onSpellFlags  func()
}

func (r *playerCantCastTestRuntime4FD150) add(event string) {
	r.events = append(r.events, event)
}

func (r *playerCantCastTestRuntime4FD150) hooks() playerCantCastSpellHooks4FD150[
	*playerCantCastTestObject4FD150,
	*playerCantCastTestObject4FD150,
] {
	return playerCantCastSpellHooks4FD150[*playerCantCastTestObject4FD150, *playerCantCastTestObject4FD150]{
		findParent: func(obj *playerCantCastTestObject4FD150) *playerCantCastTestObject4FD150 {
			r.add("find-parent")
			return obj
		},
		hasGameFlag: func(flag uint32) int32 {
			r.add(fmt.Sprintf("game:%#x", flag))
			if r.gameFlags&flag != 0 {
				return -1
			}
			return 0
		},
		loadCrownTypeCache: func() uint32 {
			r.add("load-crown")
			return r.crownCache
		},
		storeCrownTypeCache: func(value uint32) {
			r.add(fmt.Sprintf("store-crown:%d", value))
			r.crownCache = value
		},
		loadGameBallTypeCache: func() uint32 {
			r.add("load-ball")
			return r.gameBallCache
		},
		storeGameBallTypeCache: func(value uint32) {
			r.add(fmt.Sprintf("store-ball:%d", value))
			r.gameBallCache = value
		},
		lookupObjectType: func(name string) uint32 {
			r.add("lookup:" + name)
			return r.lookup[name]
		},
		spellHasFlags: func(spell int32, flags uint32) int32 {
			r.add(fmt.Sprintf("spell-flags:%d:%#x", spell, flags))
			if r.onSpellFlags != nil {
				r.onSpellFlags()
			}
			return r.spellFlags
		},
		loadFirstOwned: func(obj *playerCantCastTestObject4FD150) *playerCantCastTestObject4FD150 {
			r.add("first-owned")
			return obj.owned
		},
		loadOwnedType: func(obj *playerCantCastTestObject4FD150) uint16 {
			r.add(fmt.Sprintf("owned-type:%d", obj.typeInd))
			return obj.typeInd
		},
		loadNextOwned: func(obj *playerCantCastTestObject4FD150) *playerCantCastTestObject4FD150 {
			r.add("next-owned")
			return obj.next
		},
		hasTeam: func(*playerCantCastTestObject4FD150) int32 {
			r.add("team")
			return r.team
		},
		loadFirstInventory: func(obj *playerCantCastTestObject4FD150) *playerCantCastTestObject4FD150 {
			r.add("first-inventory")
			return obj.inv
		},
		loadInventoryFlags: func(obj *playerCantCastTestObject4FD150) uint32 {
			r.add(fmt.Sprintf("inventory-flags:%#x", obj.flags))
			return obj.flags
		},
		loadNextInventory: func(obj *playerCantCastTestObject4FD150) *playerCantCastTestObject4FD150 {
			r.add("next-inventory")
			return obj.invNext
		},
		hasEnchant: func(_ *playerCantCastTestObject4FD150, enchant uint8) int32 {
			r.add(fmt.Sprintf("enchant:%d", enchant))
			return r.enchant
		},
		spellAllowed: func(spell int32) int32 {
			r.add(fmt.Sprintf("allowed:%d", spell))
			return r.allowed
		},
		loadSummonType: func(kind playerCantCastSummonType4FD150) uint32 {
			r.add(fmt.Sprintf("summon-type:%d", kind))
			return r.types[kind]
		},
		countOwnedType: func(_ *playerCantCastTestObject4FD150, typ uint32) int32 {
			r.add(fmt.Sprintf("count:%d", typ))
			return r.counts[typ]
		},
		spellPower: func(spell int32, _ *playerCantCastTestObject4FD150) int32 {
			r.add(fmt.Sprintf("power:%d", spell))
			return r.power
		},
		balanceFloat: func(key string, index int32) float64 {
			r.add(fmt.Sprintf("balance:%s:%d", key, index))
			return r.balance[key]
		},
	}
}

func newPlayerCantCastTestRuntime4FD150() *playerCantCastTestRuntime4FD150 {
	return &playerCantCastTestRuntime4FD150{
		lookup:  map[string]uint32{"Crown": 7, "GameBall": 8},
		allowed: 1,
		types: map[playerCantCastSummonType4FD150]uint32{
			playerCantCastPixie4FD150:        11,
			playerCantCastMagicMissile4FD150: 12,
			playerCantCastSmallFist4FD150:    13,
			playerCantCastMediumFist4FD150:   14,
			playerCantCastLargeFist4FD150:    15,
			playerCantCastDeathBall4FD150:    16,
			playerCantCastMeteor4FD150:       17,
		},
		counts:  make(map[uint32]int32),
		balance: make(map[string]float64),
	}
}

func TestPlayerCantCastSpell4FD150KOTRReloadsLazyCrownCache(t *testing.T) {
	runtime := newPlayerCantCastTestRuntime4FD150()
	runtime.gameFlags = playerCantCastModeKOTR4FD150 | playerCantCastModeFlagBall4FD150 | playerCantCastModeCTF4FD150
	runtime.spellFlags = 1
	runtime.team = -1
	unit := &playerCantCastTestObject4FD150{owned: &playerCantCastTestObject4FD150{typeInd: 9}}
	runtime.onSpellFlags = func() { runtime.crownCache = 9 }

	if got := playerCantCastSpell4FD150(unit, 22, 0, runtime.hooks()); got != 17 {
		t.Fatalf("result = %d, want 17", got)
	}
	want := []string{
		"find-parent", "game:0x10", "load-crown", "lookup:Crown", "store-crown:7",
		"spell-flags:22:0x80000", "first-owned", "load-crown", "owned-type:9", "team",
	}
	if !reflect.DeepEqual(runtime.events, want) {
		t.Fatalf("events = %#v, want %#v", runtime.events, want)
	}
}

func TestPlayerCantCastSpell4FD150FlagBallAndCTF(t *testing.T) {
	t.Run("flag-ball", func(t *testing.T) {
		runtime := newPlayerCantCastTestRuntime4FD150()
		runtime.gameFlags = playerCantCastModeFlagBall4FD150 | playerCantCastModeCTF4FD150
		runtime.spellFlags = 1
		runtime.gameBallCache = 8
		unit := &playerCantCastTestObject4FD150{owned: &playerCantCastTestObject4FD150{typeInd: 8}}
		if got := playerCantCastSpell4FD150(unit, 23, 0, runtime.hooks()); got != 16 {
			t.Fatalf("result = %d, want 16", got)
		}
		want := []string{
			"find-parent", "game:0x10", "game:0x40", "load-ball", "spell-flags:23:0x80000",
			"first-owned", "load-ball", "owned-type:8",
		}
		if !reflect.DeepEqual(runtime.events, want) {
			t.Fatalf("events = %#v, want %#v", runtime.events, want)
		}
	})

	t.Run("ctf", func(t *testing.T) {
		runtime := newPlayerCantCastTestRuntime4FD150()
		runtime.gameFlags = playerCantCastModeCTF4FD150
		runtime.spellFlags = 1
		second := &playerCantCastTestObject4FD150{flags: playerCantCastInventoryFlag4FD150}
		unit := &playerCantCastTestObject4FD150{inv: &playerCantCastTestObject4FD150{invNext: second}}
		if got := playerCantCastSpell4FD150(unit, 24, 0, runtime.hooks()); got != 13 {
			t.Fatalf("result = %d, want 13", got)
		}
		want := []string{
			"find-parent", "game:0x10", "game:0x40", "game:0x20", "spell-flags:24:0x80000",
			"first-inventory", "inventory-flags:0x0", "next-inventory", "inventory-flags:0x10000000",
		}
		if !reflect.DeepEqual(runtime.events, want) {
			t.Fatalf("events = %#v, want %#v", runtime.events, want)
		}
	})
}

func TestPlayerCantCastSpell4FD150BypassAndCommonGates(t *testing.T) {
	unit := &playerCantCastTestObject4FD150{}

	runtime := newPlayerCantCastTestRuntime4FD150()
	runtime.gameFlags = ^uint32(0)
	runtime.enchant = 1
	if got := playerCantCastSpell4FD150(unit, 29, 1, runtime.hooks()); got != 14 {
		t.Fatalf("anti-magic result = %d, want 14", got)
	}
	if want := []string{"find-parent", "enchant:29"}; !reflect.DeepEqual(runtime.events, want) {
		t.Fatalf("anti-magic events = %#v, want %#v", runtime.events, want)
	}

	runtime = newPlayerCantCastTestRuntime4FD150()
	runtime.allowed = 0
	if got := playerCantCastSpell4FD150(unit, 111, 1, runtime.hooks()); got != 10 {
		t.Fatalf("disabled result = %d, want 10", got)
	}
	if want := []string{"find-parent", "enchant:29", "allowed:111"}; !reflect.DeepEqual(runtime.events, want) {
		t.Fatalf("disabled events = %#v, want %#v", runtime.events, want)
	}
}

func TestPlayerCantCastSpell4FD150SummonCases(t *testing.T) {
	tests := []struct {
		name   string
		spell  int32
		counts map[uint32]int32
		want   int32
	}{
		{name: "fist empty", spell: 29, want: 0},
		{name: "large fist", spell: 29, counts: map[uint32]int32{15: 1}, want: 3},
		{name: "medium fist", spell: 29, counts: map[uint32]int32{14: 1}, want: 3},
		{name: "small fist", spell: 29, counts: map[uint32]int32{13: 1}, want: 3},
		{name: "death ball empty", spell: 31, want: 0},
		{name: "death ball present", spell: 31, counts: map[uint32]int32{16: 1}, want: 3},
		{name: "meteor empty", spell: 52, want: 0},
		{name: "meteor present", spell: 52, counts: map[uint32]int32{17: 1}, want: 3},
		{name: "ordinary spell", spell: 90, want: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runtime := newPlayerCantCastTestRuntime4FD150()
			for typ, count := range tc.counts {
				runtime.counts[typ] = count
			}
			if got := playerCantCastSpell4FD150(&playerCantCastTestObject4FD150{}, tc.spell, 1, runtime.hooks()); got != tc.want {
				t.Fatalf("result = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestPlayerCantCastSpell4FD150MissileAndPixieOrder(t *testing.T) {
	for _, tc := range []struct {
		name       string
		spell      int32
		kind       playerCantCastSummonType4FD150
		typeID     uint32
		balanceKey string
	}{
		{name: "missile", spell: 50, kind: playerCantCastMagicMissile4FD150, typeID: 12, balanceKey: "MagicMissileCount"},
		{name: "pixie", spell: 58, kind: playerCantCastPixie4FD150, typeID: 11, balanceKey: "PixieCount"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runtime := newPlayerCantCastTestRuntime4FD150()
			runtime.counts[tc.typeID] = 2
			runtime.power = 3
			runtime.balance[tc.balanceKey] = 2.9
			if got := playerCantCastSpell4FD150(&playerCantCastTestObject4FD150{}, tc.spell, 1, runtime.hooks()); got != 3 {
				t.Fatalf("result = %d, want 3", got)
			}
			wantTail := []string{
				fmt.Sprintf("summon-type:%d", tc.kind), fmt.Sprintf("count:%d", tc.typeID),
				fmt.Sprintf("power:%d", tc.spell), fmt.Sprintf("balance:%s:2", tc.balanceKey),
			}
			if got := runtime.events[len(runtime.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
				t.Fatalf("event tail = %#v, want %#v", got, wantTail)
			}

			runtime = newPlayerCantCastTestRuntime4FD150()
			runtime.counts[tc.typeID] = 1
			runtime.power = 1
			runtime.balance[tc.balanceKey] = 2
			if got := playerCantCastSpell4FD150(&playerCantCastTestObject4FD150{}, tc.spell, 1, runtime.hooks()); got != 0 {
				t.Fatalf("below-limit result = %d, want 0", got)
			}
		})
	}
}
