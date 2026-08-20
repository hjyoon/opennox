package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type boltDamageTestModifier4EF1E0 struct {
	name        string
	typeIndex   uint32
	required    uint16
	coefficient float32
	minimum     uint16
}

type boltDamageTestWorld4EF1E0 struct {
	cache        uint32
	lookupResult uint32
	flagResult   int32
	modifier     *boltDamageTestModifier4EF1E0
	strength     int32
	soloMinimum  float64

	events  []string
	stores  []uint32
	faultAt int
	after   map[string]func()
}

func boltDamageModifierName4EF1E0(modifier *boltDamageTestModifier4EF1E0) string {
	if modifier == nil {
		return "nil"
	}
	return modifier.name
}

func (w *boltDamageTestWorld4EF1E0) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *boltDamageTestWorld4EF1E0) finish(event string) {
	after := w.after[event]
	if after == nil {
		return
	}
	delete(w.after, event)
	after()
}

func (w *boltDamageTestWorld4EF1E0) hooks() boltDamageHooks4EF1E0[*boltDamageTestModifier4EF1E0] {
	return boltDamageHooks4EF1E0[*boltDamageTestModifier4EF1E0]{
		loadCachedArcherBoltType: func() uint32 {
			value := w.cache
			event := fmt.Sprintf("cache=%08x", value)
			w.record(event)
			w.finish(event)
			return value
		},
		lookupType: func(name string) uint32 {
			value := w.lookupResult
			event := fmt.Sprintf("lookup:%s=%08x", name, value)
			w.record(event)
			w.finish(event)
			return value
		},
		storeCachedArcherBoltType: func(value uint32) {
			event := fmt.Sprintf("store-cache=%08x", value)
			w.record(event)
			w.cache = value
			w.stores = append(w.stores, value)
			w.finish(event)
		},
		gameFlagsCheck: func(mask uint32) int32 {
			value := w.flagResult
			event := fmt.Sprintf("flag:%08x=%d", mask, value)
			w.record(event)
			w.finish(event)
			return value
		},
		loadModifierArg: func() *boltDamageTestModifier4EF1E0 {
			modifier := w.modifier
			event := "modifier=" + boltDamageModifierName4EF1E0(modifier)
			w.record(event)
			w.finish(event)
			return modifier
		},
		loadModifierType: func(modifier *boltDamageTestModifier4EF1E0) uint32 {
			name := boltDamageModifierName4EF1E0(modifier)
			if modifier == nil {
				event := "type:" + name
				w.record(event)
				panic("nil modifier TypeInd")
			}
			value := modifier.typeIndex
			event := fmt.Sprintf("type:%s=%08x", name, value)
			w.record(event)
			w.finish(event)
			return value
		},
		balanceFloat: func(key string) float64 {
			value := w.soloMinimum
			event := fmt.Sprintf("balance:%s=%016x", key, math.Float64bits(value))
			w.record(event)
			w.finish(event)
			return value
		},
		loadStrengthArg: func() int32 {
			value := w.strength
			event := fmt.Sprintf("strength=%d", value)
			w.record(event)
			w.finish(event)
			return value
		},
		loadRequiredStrength: func(modifier *boltDamageTestModifier4EF1E0) uint16 {
			name := boltDamageModifierName4EF1E0(modifier)
			if modifier == nil {
				event := "required:" + name
				w.record(event)
				panic("nil modifier required strength")
			}
			value := modifier.required
			event := fmt.Sprintf("required:%s=%04x", name, value)
			w.record(event)
			w.finish(event)
			return value
		},
		loadDamageMinimum: func(modifier *boltDamageTestModifier4EF1E0) uint16 {
			name := boltDamageModifierName4EF1E0(modifier)
			if modifier == nil {
				event := "minimum:" + name
				w.record(event)
				panic("nil modifier damage minimum")
			}
			value := modifier.minimum
			event := fmt.Sprintf("minimum:%s=%04x", name, value)
			w.record(event)
			w.finish(event)
			return value
		},
		loadCoefficient: func(modifier *boltDamageTestModifier4EF1E0) float32 {
			name := boltDamageModifierName4EF1E0(modifier)
			if modifier == nil {
				event := "coefficient:" + name
				w.record(event)
				panic("nil modifier coefficient")
			}
			value := modifier.coefficient
			event := fmt.Sprintf("coefficient:%s=%08x", name, math.Float32bits(value))
			w.record(event)
			w.finish(event)
			return value
		},
	}
}

func newBoltDamageTestWorld4EF1E0() *boltDamageTestWorld4EF1E0 {
	return &boltDamageTestWorld4EF1E0{
		cache:        73,
		lookupResult: 73,
		modifier: &boltDamageTestModifier4EF1E0{
			name:        "primary",
			typeIndex:   73,
			required:    12,
			coefficient: 1.25,
			minimum:     3,
		},
		strength:    37,
		soloMinimum: 5.5,
		after:       make(map[string]func()),
	}
}

func boltDamageNormalEvents4EF1E0(flag int32) []string {
	return []string{
		"cache=00000049",
		fmt.Sprintf("flag:00000800=%d", flag),
		"modifier=primary",
		"strength=37",
		"required:primary=000c",
		"minimum:primary=0003",
		"coefficient:primary=3fa00000",
	}
}

func boltDamageMustPanic4EF1E0(t *testing.T, run func()) {
	t.Helper()
	deferred := false
	func() {
		defer func() {
			deferred = recover() != nil
		}()
		run()
	}()
	if !deferred {
		t.Fatal("call did not panic")
	}
}

func TestBoltDamage4EF1E0ExactFlagAndBranchOrder(t *testing.T) {
	for _, flag := range []int32{0, 2, -1, math.MinInt32} {
		t.Run(fmt.Sprintf("normal-%d", flag), func(t *testing.T) {
			world := newBoltDamageTestWorld4EF1E0()
			world.flagResult = flag
			if got := boltDamage4EF1E0(world.hooks()); got != 34.25 {
				t.Fatalf("damage = %v, want 34.25", got)
			}
			if want := boltDamageNormalEvents4EF1E0(flag); !reflect.DeepEqual(world.events, want) {
				t.Fatalf("events = %q, want %q", world.events, want)
			}
		})
	}

	t.Run("cooperative ArcherBolt", func(t *testing.T) {
		world := newBoltDamageTestWorld4EF1E0()
		world.flagResult = 1
		if got := boltDamage4EF1E0(world.hooks()); got != 36.75 {
			t.Fatalf("damage = %v, want 36.75", got)
		}
		want := []string{
			"cache=00000049",
			"flag:00000800=1",
			"modifier=primary",
			"type:primary=00000049",
			"cache=00000049",
			"balance:BoltSoloDamageMin=4016000000000000",
			"strength=37",
			"required:primary=000c",
			"coefficient:primary=3fa00000",
		}
		if !reflect.DeepEqual(world.events, want) {
			t.Fatalf("events = %q, want %q", world.events, want)
		}
	})

	t.Run("cooperative other projectile falls through", func(t *testing.T) {
		world := newBoltDamageTestWorld4EF1E0()
		world.flagResult = 1
		world.modifier.typeIndex = 74
		if got := boltDamage4EF1E0(world.hooks()); got != 34.25 {
			t.Fatalf("damage = %v, want 34.25", got)
		}
		want := []string{
			"cache=00000049", "flag:00000800=1", "modifier=primary",
			"type:primary=0000004a", "cache=00000049", "strength=37",
			"required:primary=000c", "minimum:primary=0003",
			"coefficient:primary=3fa00000",
		}
		if !reflect.DeepEqual(world.events, want) {
			t.Fatalf("events = %q, want %q", world.events, want)
		}
	})
}

func TestBoltDamage4EF1E0ZeroCacheRetriesAndCanMatchZero(t *testing.T) {
	world := newBoltDamageTestWorld4EF1E0()
	world.cache = 0
	world.lookupResult = 0
	world.flagResult = 1
	world.modifier.typeIndex = 0
	for iteration := 0; iteration < 2; iteration++ {
		if got := boltDamage4EF1E0(world.hooks()); got != 36.75 {
			t.Fatalf("iteration %d damage = %v, want 36.75", iteration, got)
		}
	}
	if !reflect.DeepEqual(world.stores, []uint32{0, 0}) {
		t.Fatalf("cache stores = %v, want [0 0]", world.stores)
	}
	lookupEvent := "lookup:ArcherBolt=00000000"
	lookupCount := 0
	for _, event := range world.events {
		if event == lookupEvent {
			lookupCount++
		}
	}
	if lookupCount != 2 {
		t.Fatalf("lookup count = %d, want 2; events=%q", lookupCount, world.events)
	}
}

func TestBoltDamage4EF1E0ReloadsCacheAfterTypeRead(t *testing.T) {
	world := newBoltDamageTestWorld4EF1E0()
	world.flagResult = 1
	world.after["type:primary=00000049"] = func() {
		world.cache = 74
	}
	if got := boltDamage4EF1E0(world.hooks()); got != 34.25 {
		t.Fatalf("damage = %v, want normal-path 34.25", got)
	}
	wantPrefix := []string{
		"cache=00000049", "flag:00000800=1", "modifier=primary",
		"type:primary=00000049", "cache=0000004a", "strength=37",
	}
	if !reflect.DeepEqual(world.events[:len(wantPrefix)], wantPrefix) {
		t.Fatalf("events prefix = %q, want %q", world.events[:len(wantPrefix)], wantPrefix)
	}
}

func TestBoltDamage4EF1E0LoadsModifierAndFieldsAtOriginalTimes(t *testing.T) {
	t.Run("modifier after flag callback", func(t *testing.T) {
		world := newBoltDamageTestWorld4EF1E0()
		replacement := &boltDamageTestModifier4EF1E0{
			name: "replacement", required: 2, coefficient: 2, minimum: 7,
		}
		world.after["flag:00000800=0"] = func() {
			world.modifier = replacement
		}
		if got := boltDamage4EF1E0(world.hooks()); got != 77 {
			t.Fatalf("damage = %v, want 77", got)
		}
		if world.events[2] != "modifier=replacement" {
			t.Fatalf("modifier event = %q, want replacement", world.events[2])
		}
	})

	t.Run("cooperative fields after balance callback", func(t *testing.T) {
		world := newBoltDamageTestWorld4EF1E0()
		world.flagResult = 1
		balanceEvent := fmt.Sprintf(
			"balance:BoltSoloDamageMin=%016x", math.Float64bits(world.soloMinimum),
		)
		world.after[balanceEvent] = func() {
			world.strength = -5
			world.modifier.required = 40
			world.modifier.coefficient = 0.75
			world.modifier.minimum = 0xffff
		}
		if got := boltDamage4EF1E0(world.hooks()); got != -28.25 {
			t.Fatalf("damage = %v, want -28.25", got)
		}
		wantTail := []string{
			balanceEvent, "strength=-5", "required:primary=0028", "coefficient:primary=3f400000",
		}
		if !reflect.DeepEqual(world.events[len(world.events)-len(wantTail):], wantTail) {
			t.Fatalf("events tail = %q, want %q", world.events[len(world.events)-len(wantTail):], wantTail)
		}
		for _, event := range world.events {
			if event == "minimum:primary=ffff" {
				t.Fatalf("cooperative ArcherBolt read DamageMin: %q", world.events)
			}
		}
	})

	t.Run("normal path uses live ordered fields", func(t *testing.T) {
		world := newBoltDamageTestWorld4EF1E0()
		world.after["strength=37"] = func() {
			world.modifier.required = 20
		}
		world.after["required:primary=0014"] = func() {
			world.modifier.minimum = 9
		}
		world.after["minimum:primary=0009"] = func() {
			world.modifier.coefficient = 2
		}
		if got := boltDamage4EF1E0(world.hooks()); got != 43 {
			t.Fatalf("damage = %v, want 43", got)
		}
		wantTail := []string{
			"strength=37", "required:primary=0014", "minimum:primary=0009", "coefficient:primary=40000000",
		}
		if !reflect.DeepEqual(world.events[len(world.events)-len(wantTail):], wantTail) {
			t.Fatalf("events tail = %q, want %q", world.events[len(world.events)-len(wantTail):], wantTail)
		}
	})
}

func TestBoltDamage4EF1E0WrappingDeltaAndBinary32Widening(t *testing.T) {
	t.Run("signed subtraction wraps", func(t *testing.T) {
		world := newBoltDamageTestWorld4EF1E0()
		world.strength = math.MinInt32
		world.modifier.required = 1
		world.modifier.coefficient = 1
		world.modifier.minimum = 0
		if got := boltDamage4EF1E0(world.hooks()); got != float64(math.MaxInt32) {
			t.Fatalf("damage = %v, want %v", got, float64(math.MaxInt32))
		}
	})

	t.Run("coefficient widens from exact binary32", func(t *testing.T) {
		world := newBoltDamageTestWorld4EF1E0()
		coefficient := math.Float32frombits(0x3dcccccd)
		world.strength = -5
		world.modifier.required = 12
		world.modifier.coefficient = coefficient
		world.modifier.minimum = 0xffff
		want := boltDamageAdd64_4EF1E0(
			boltDamageMul64_4EF1E0(float64(int32(-17)), float64(coefficient)),
			float64(uint16(0xffff)),
		)
		got := boltDamage4EF1E0(world.hooks())
		if math.Float64bits(got) != math.Float64bits(want) {
			t.Fatalf("damage bits = %016x, want %016x", math.Float64bits(got), math.Float64bits(want))
		}
	})
}

func TestBoltDamage4EF1E0NilModifierFaultBoundaries(t *testing.T) {
	t.Run("normal reads strength before required field fault", func(t *testing.T) {
		world := newBoltDamageTestWorld4EF1E0()
		world.modifier = nil
		boltDamageMustPanic4EF1E0(t, func() {
			boltDamage4EF1E0(world.hooks())
		})
		want := []string{
			"cache=00000049", "flag:00000800=0", "modifier=nil", "strength=37", "required:nil",
		}
		if !reflect.DeepEqual(world.events, want) {
			t.Fatalf("events = %q, want %q", world.events, want)
		}
	})

	t.Run("exact cooperative mode faults on TypeInd first", func(t *testing.T) {
		world := newBoltDamageTestWorld4EF1E0()
		world.flagResult = 1
		world.modifier = nil
		boltDamageMustPanic4EF1E0(t, func() {
			boltDamage4EF1E0(world.hooks())
		})
		want := []string{
			"cache=00000049", "flag:00000800=1", "modifier=nil", "type:nil",
		}
		if !reflect.DeepEqual(world.events, want) {
			t.Fatalf("events = %q, want %q", world.events, want)
		}
	})
}

func TestBoltDamage4EF1E0EveryObservationFaultPrefix(t *testing.T) {
	paths := []struct {
		name  string
		build func() *boltDamageTestWorld4EF1E0
	}{
		{
			name: "cache miss cooperative match",
			build: func() *boltDamageTestWorld4EF1E0 {
				world := newBoltDamageTestWorld4EF1E0()
				world.cache = 0
				world.flagResult = 1
				return world
			},
		},
		{
			name:  "normal",
			build: newBoltDamageTestWorld4EF1E0,
		},
		{
			name: "cooperative mismatch",
			build: func() *boltDamageTestWorld4EF1E0 {
				world := newBoltDamageTestWorld4EF1E0()
				world.flagResult = 1
				world.modifier.typeIndex = 74
				return world
			},
		},
	}

	for _, path := range paths {
		t.Run(path.name, func(t *testing.T) {
			baseline := path.build()
			boltDamage4EF1E0(baseline.hooks())
			want := append([]string(nil), baseline.events...)
			for faultAt := 1; faultAt <= len(want); faultAt++ {
				t.Run(fmt.Sprintf("fault-%02d", faultAt), func(t *testing.T) {
					world := path.build()
					world.faultAt = faultAt
					boltDamageMustPanic4EF1E0(t, func() {
						boltDamage4EF1E0(world.hooks())
					})
					if prefix := want[:faultAt]; !reflect.DeepEqual(world.events, prefix) {
						t.Fatalf("events = %q, want prefix %q", world.events, prefix)
					}
				})
			}
		})
	}
}
