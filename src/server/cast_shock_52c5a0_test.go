package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

const (
	shockCasterTest52C5A0  = uint64(0x7fd4c759df40)
	shockContextTest52C5A0 = uint64(0x7fd4c75743d0)
	shockTargetTest52C5A0  = uint64(0x7fd4c759db90)
)

type shockTestWorld52C5A0 struct {
	target, context uint64
	cache, lookup   uint32
	contextType     uint16
	value           float64
	faultAt         int
	events          []string
}

func (w *shockTestWorld52C5A0) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic("Shock fixture fault")
	}
}

func (w *shockTestWorld52C5A0) hooks() castShockHooks52C5A0[uint64] {
	return castShockHooks52C5A0[uint64]{
		target:         func() uint64 { w.observe(fmt.Sprintf("target:%x", w.target)); return w.target },
		loadGlyphType:  func() uint32 { w.observe(fmt.Sprintf("cache:%x", w.cache)); return w.cache },
		storeGlyphType: func(id uint32) { w.observe(fmt.Sprintf("store:%x", id)); w.cache = id },
		lookupType:     func(name string) uint32 { w.observe("lookup:" + name); return w.lookup },
		typeIndex:      func(unit uint64) uint16 { w.observe(fmt.Sprintf("type:%x", unit)); return w.contextType },
		balance:        func(key string) float64 { w.observe("balance:" + key); return w.value },
		balanceIndex: func(key string, index int32) float64 {
			w.observe(fmt.Sprintf("table:%s:%d", key, index))
			return w.value
		},
		floatToInt: func(value float32) int32 {
			w.observe(fmt.Sprintf("round:%g", value))
			return aiPathFloatToInt419A70(value)
		},
		damage: func(unit, source, weapon uint64, damage, typ int32) bool {
			w.observe(fmt.Sprintf("damage:%x:%x:%x:%d:%d", unit, source, weapon, damage, typ))
			return false // The original cast ignores damage rejection.
		},
		apply: func(unit uint64, buff int32, duration int16, power int8) {
			w.observe(fmt.Sprintf("apply:%x:%d:%d:%d", unit, buff, duration, power))
		},
	}
}

func TestCastShock52C5A0ExactBranchesAndTrace(t *testing.T) {
	for _, trap := range []bool{false, true} {
		t.Run(fmt.Sprintf("trap=%t", trap), func(t *testing.T) {
			w := &shockTestWorld52C5A0{target: shockTargetTest52C5A0, lookup: 0x8123, contextType: 0x8123, value: 32767.5}
			if trap {
				w.context = shockContextTest52C5A0
			}
			if got := castShock52C5A0(shockCasterTest52C5A0, w.context, 0x180, w.hooks()); got != 1 || w.cache != 0x8123 {
				t.Fatalf("result/cache=%d/%x", got, w.cache)
			}
			want := []string{"target:7fd4c759db90", "cache:0", "lookup:Glyph", "store:8123"}
			if trap {
				want = append(want, "type:7fd4c75743d0", "table:ShockTrapDamage:383", "round:32767.5", "target:7fd4c759db90", "damage:7fd4c759db90:7fd4c759df40:7fd4c759df40:32768:9")
			} else {
				want = append(want, "balance:ShockEnchantDuration", "round:32767.5", "target:7fd4c759db90", "apply:7fd4c759db90:22:-32768:-128")
			}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events=%q want=%q", w.events, want)
			}
		})
	}
}

func TestCastShock52C5A0OnlyEntryTargetGatesServices(t *testing.T) {
	// All other hooks are absent. A nonnull, deliberately opaque context must
	// not be read if the entry target is null; nor may Glyph be initialized.
	if got := castShock52C5A0(shockCasterTest52C5A0, shockContextTest52C5A0, math.MinInt32, castShockHooks52C5A0[uint64]{target: func() uint64 { return 0 }}); got != 0 {
		t.Fatalf("nil target=%d", got)
	}
}

func TestCastShock52C5A0CacheUsesWholeDwordAndLocalLookupResult(t *testing.T) {
	for _, tc := range []struct {
		cache, lookup uint32
		typ           uint16
		trap          bool
	}{
		{0x8123, 7, 0x8123, true}, {0x18123, 7, 0x8123, false},
		{math.MaxUint32, 7, math.MaxUint16, false}, {0, 0, 0, true},
		{0, 0x8123, 0x8123, true}, {0, 0x18123, 0x8123, false},
	} {
		t.Run(fmt.Sprintf("%x/%x/%x", tc.cache, tc.lookup, tc.typ), func(t *testing.T) {
			w := &shockTestWorld52C5A0{target: shockTargetTest52C5A0, cache: tc.cache, lookup: tc.lookup, contextType: tc.typ}
			h := w.hooks()
			lookups, stores, traps, buffs := 0, 0, 0, 0
			h.lookupType = func(name string) uint32 {
				lookups++
				if name != "Glyph" {
					t.Fatal(name)
				}
				return tc.lookup
			}
			h.storeGlyphType = func(id uint32) {
				stores++
				if id != tc.lookup {
					t.Fatalf("stored cache=%x", id)
				}
				w.cache = 0xfedcba98 // Comparison must still use the lookup result.
			}
			h.damage = func(uint64, uint64, uint64, int32, int32) bool { traps++; return false }
			h.apply = func(uint64, int32, int16, int8) { buffs++ }
			if got := castShock52C5A0(uint64(0), shockContextTest52C5A0, 1, h); got != 1 {
				t.Fatal(got)
			}
			wantLookups := 0
			if tc.cache == 0 {
				wantLookups = 1
			}
			if lookups != wantLookups || stores != wantLookups || (traps == 1) != tc.trap || (buffs == 1) == tc.trap {
				t.Fatalf("lookup/store/trap/buff=%d/%d/%d/%d", lookups, stores, traps, buffs)
			}
		})
	}
	w := &shockTestWorld52C5A0{target: shockTargetTest52C5A0}
	for i := 0; i < 2; i++ {
		castShock52C5A0(uint64(0), uint64(0), 1, w.hooks())
	}
	lookups := 0
	for _, event := range w.events {
		if event == "lookup:Glyph" {
			lookups++
		}
	}
	if lookups != 2 {
		t.Fatalf("zero cache must retry on the next call: %d", lookups)
	}
}

func TestCastShock52C5A0ReloadAfterCallbacksAndNoPostEffectRead(t *testing.T) {
	const afterLookup, afterBalance, afterRound = uint64(0x1234567890), uint64(0x2345678901), uint64(0x3456789012)
	for _, trap := range []bool{false, true} {
		w := &shockTestWorld52C5A0{target: shockTargetTest52C5A0, lookup: 17, contextType: 17}
		h := w.hooks()
		reads := 0
		h.target = func() uint64 { reads++; return w.target }
		h.lookupType = func(string) uint32 { w.target = afterLookup; return 17 }
		h.balance = func(string) float64 { w.target = afterBalance; return 2.50000001 }
		h.balanceIndex = func(string, int32) float64 { w.target = afterBalance; return 2.50000001 }
		h.floatToInt = func(value float32) int32 {
			if value != 2.5 {
				t.Fatalf("missing binary32 spill: %g", value)
			}
			w.target = afterRound
			return 2
		}
		calls := 0
		h.damage = func(unit, source, weapon uint64, damage, typ int32) bool {
			calls++
			if unit != afterRound || source != shockCasterTest52C5A0 || weapon != source || damage != 2 || typ != 9 {
				t.Fatal("stale trap arguments")
			}
			w.target = 0
			return false
		}
		h.apply = func(unit uint64, buff int32, duration int16, power int8) {
			calls++
			if unit != afterRound || buff != 22 || duration != 2 || power != 127 {
				t.Fatal("stale buff arguments")
			}
			w.target = 0
		}
		context := uint64(0)
		if trap {
			context = shockContextTest52C5A0
		}
		if got := castShock52C5A0(shockCasterTest52C5A0, context, -129, h); got != 1 || reads != 2 || calls != 1 || w.target != 0 {
			t.Fatalf("result/reads/effects/live=%d/%d/%d/%x", got, reads, calls, w.target)
		}
	}
}

func TestCastShock52C5A0RoundingAndSignedBoundaries(t *testing.T) {
	for _, tc := range []struct {
		value   float64
		rounded int32
	}{
		{2.50000001, 2}, {3.5, 4}, {-2.5, -2}, {32767.5, 32768}, {65535, 65535},
		{65535.5, 65536}, {2147483647, math.MinInt32}, {-2147483648, math.MinInt32},
		{math.NaN(), math.MinInt32}, {math.Inf(1), math.MinInt32}, {math.Inf(-1), math.MinInt32},
	} {
		for _, trap := range []bool{false, true} {
			t.Run(fmt.Sprintf("%g/trap=%t", tc.value, trap), func(t *testing.T) {
				w := &shockTestWorld52C5A0{target: shockTargetTest52C5A0, cache: 17, contextType: 17, value: tc.value}
				h := w.hooks()
				h.damage = func(_ uint64, _, _ uint64, value, typ int32) bool {
					if value != tc.rounded || typ != 9 {
						t.Fatalf("damage=%d/%d", value, typ)
					}
					return false
				}
				h.apply = func(_ uint64, _ int32, value int16, power int8) {
					if value != int16(tc.rounded) || power != -1 {
						t.Fatalf("buff=%d/%d", value, power)
					}
				}
				context := uint64(0)
				if trap {
					context = shockContextTest52C5A0
				}
				if got := castShock52C5A0(shockCasterTest52C5A0, context, math.MaxInt32, h); got != 1 {
					t.Fatal(got)
				}
			})
		}
	}
	for _, tc := range []struct {
		power, index int32
		bytePower    int8
	}{
		{math.MinInt32, math.MaxInt32, 0}, {math.MaxInt32, math.MaxInt32 - 1, -1},
		{-129, -130, 127}, {-1, -2, -1}, {0, -1, 0}, {1, 0, 1}, {5, 4, 5}, {0x180, 383, -128},
	} {
		w := &shockTestWorld52C5A0{target: shockTargetTest52C5A0, cache: 17, contextType: 17}
		h := w.hooks()
		h.balanceIndex = func(key string, index int32) float64 {
			if key != "ShockTrapDamage" || index != tc.index {
				t.Fatalf("table index for %d=%s/%d", tc.power, key, index)
			}
			return 0
		}
		h.apply = func(_ uint64, _ int32, _ int16, power int8) {
			if power != tc.bytePower {
				t.Fatalf("power for %d=%d", tc.power, power)
			}
		}
		castShock52C5A0(uint64(0), shockContextTest52C5A0, tc.power, h)
		castShock52C5A0(uint64(0), uint64(0), tc.power, h)
	}
}

func TestCastShock52C5A0EveryFaultPrefix(t *testing.T) {
	for _, trap := range []bool{false, true} {
		baseline := &shockTestWorld52C5A0{target: shockTargetTest52C5A0, lookup: 17, contextType: 17, value: 4}
		context := uint64(0)
		if trap {
			context = shockContextTest52C5A0
		}
		castShock52C5A0(shockCasterTest52C5A0, context, 1, baseline.hooks())
		for fault := 1; fault <= len(baseline.events); fault++ {
			t.Run(fmt.Sprintf("trap=%t/fault=%d", trap, fault), func(t *testing.T) {
				w := &shockTestWorld52C5A0{target: shockTargetTest52C5A0, lookup: 17, contextType: 17, value: 4, faultAt: fault}
				var recovered any
				func() {
					defer func() { recovered = recover() }()
					castShock52C5A0(shockCasterTest52C5A0, context, 1, w.hooks())
				}()
				if recovered != "Shock fixture fault" || !reflect.DeepEqual(w.events, baseline.events[:fault]) {
					t.Fatalf("fault=%v events=%q want=%q", recovered, w.events, baseline.events[:fault])
				}
			})
		}
	}
}
