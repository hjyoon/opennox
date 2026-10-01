package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

func TestCastStun52C2C0ExactTraceAndCachedTarget(t *testing.T) {
	for _, tc := range []struct {
		name        string
		class       uint8
		playerClass uint8
		mass        float32
		buff        int32
		read        string
	}{
		{"warrior", 4, 0, 1000, 4, "player"},
		{"wizard", 4, 1, 0, 5, "player"},
		{"conjurer", 4, 2, 0, 5, "player"},
		{"raw-class", 4, 255, 0, 5, "player"},
		{"player-precedes-monster", 6, 1, 1000, 5, "player"},
		{"heavy-monster", 2, 0, 15.000001, 4, "mass"},
		{"mass-boundary", 2, 0, 15, 5, "mass"},
		{"light-monster", 2, 0, 14.999999, 5, "mass"},
		{"unordered-mass", 2, 0, float32(math.NaN()), 5, "mass"},
		{"positive-infinity", 2, 0, float32(math.Inf(1)), 4, "mass"},
		{"negative-infinity", 2, 0, float32(math.Inf(-1)), 5, "mass"},
		{"negative-zero", 2, 0, float32(math.Copysign(0, -1)), 5, "mass"},
		{"other-object", 1, 0, 1000, 5, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const source, gated, cached, changed, attributed = uint64(0x1111222233334444), uint64(0x5555666677778888), uint64(0x9999aaaabbbbcccc), uint64(0xeeee111122223333), uint64(0xdddd111122223333)
			target := gated
			var trace []string
			h := castStunHooks52C2C0[uint64]{
				target: func() uint64 { trace = append(trace, fmt.Sprintf("target:%x", target)); return target },
				balance: func(key string) float64 {
					trace = append(trace, "balance:"+key)
					target = cached
					return 32767.5
				},
				floatToInt: func(value float32) int32 {
					trace = append(trace, fmt.Sprintf("convert:%g", value))
					return aiPathFloatToInt419A70(value)
				},
				classLow: func(unit uint64) uint8 {
					trace = append(trace, fmt.Sprintf("class:%x", unit))
					target = changed
					return tc.class
				},
				playerClass: func(unit uint64) uint8 { trace = append(trace, fmt.Sprintf("player:%x", unit)); return tc.playerClass },
				mass:        func(unit uint64) float32 { trace = append(trace, fmt.Sprintf("mass:%x", unit)); return tc.mass },
				apply: func(unit uint64, buff int32, duration int16, power int8) {
					trace = append(trace, fmt.Sprintf("apply:%x/%d/%d/%d", unit, buff, duration, power))
					target = attributed
				},
				attribution: func(caster, unit uint64) { trace = append(trace, fmt.Sprintf("attribute:%x/%x", caster, unit)) },
			}
			if got := castStun52C2C0(source, 0x180, h); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			want := []string{"target:5555666677778888", "balance:StunEnchantDuration", "convert:32767.5", "target:9999aaaabbbbcccc", "class:9999aaaabbbbcccc"}
			if tc.read != "" {
				want = append(want, tc.read+":9999aaaabbbbcccc")
			}
			want = append(want, fmt.Sprintf("apply:9999aaaabbbbcccc/%d/-32768/-128", tc.buff), "target:dddd111122223333", "attribute:1111222233334444/dddd111122223333")
			if !reflect.DeepEqual(trace, want) {
				t.Fatalf("trace = %v, want %v", trace, want)
			}
		})
	}
}

func TestCastStun52C2C0NilTargetGate(t *testing.T) {
	calls := 0
	h := castStunHooks52C2C0[uint64]{target: func() uint64 { calls++; return 0 }}
	if got := castStun52C2C0(uint64(1), 1, h); got != 0 || calls != 1 {
		t.Fatalf("result/reads = %d/%d, want 0/1", got, calls)
	}
}

func TestCastStun52C2C0Binary32RoundingAndFixedWidths(t *testing.T) {
	for _, tc := range []struct {
		value     float64
		power     int32
		duration  int16
		wantPower int8
	}{
		{2.50000001, 0x180, 2, -128}, {3.5, 255, 4, -1}, {-2.5, -129, -2, 127},
		{32767.5, math.MaxInt32, math.MinInt16, -1}, {65535, math.MinInt32, -1, 0},
		{math.NaN(), 1, 0, 1}, {math.Inf(1), 1, 0, 1}, {math.Inf(-1), 1, 0, 1},
		{2147483648, 1, 0, 1}, {-2147483648, 1, 0, 1},
	} {
		t.Run(fmt.Sprintf("%g/%d", tc.value, tc.power), func(t *testing.T) {
			applyCalls, attributeCalls := 0, 0
			h := castStunHooks52C2C0[uint64]{
				target: func() uint64 { return 0x1111222233334444 }, balance: func(string) float64 { return tc.value },
				floatToInt: aiPathFloatToInt419A70, classLow: func(uint64) uint8 { return 0 },
				apply: func(_ uint64, buff int32, duration int16, power int8) {
					applyCalls++
					if buff != 5 || duration != tc.duration || power != tc.wantPower {
						t.Fatalf("buff/duration/power = %d/%d/%d", buff, duration, power)
					}
				},
				attribution: func(uint64, uint64) { attributeCalls++ },
			}
			if got := castStun52C2C0(uint64(0), tc.power, h); got != 1 || applyCalls != 1 || attributeCalls != 1 {
				t.Fatalf("result/calls = %d/%d/%d", got, applyCalls, attributeCalls)
			}
		})
	}
}

func TestCastStun52C2C0FaultPrefixes(t *testing.T) {
	for _, class := range []uint8{0, 2, 4} {
		steps := 9
		if class == 0 {
			steps = 8
		}
		for fault := 1; fault <= steps; fault++ {
			t.Run(fmt.Sprintf("class=%d/fault=%d", class, fault), func(t *testing.T) {
				calls := 0
				step := func() {
					calls++
					if calls == fault {
						panic("fixture fault")
					}
				}
				h := castStunHooks52C2C0[uint64]{
					target: func() uint64 { step(); return 1 }, balance: func(string) float64 { step(); return 5 },
					floatToInt: func(float32) int32 { step(); return 5 }, classLow: func(uint64) uint8 { step(); return class },
					playerClass: func(uint64) uint8 { step(); return 0 }, mass: func(uint64) float32 { step(); return 15 },
					apply: func(uint64, int32, int16, int8) { step() }, attribution: func(uint64, uint64) { step() },
				}
				func() {
					defer func() {
						if got := recover(); got != "fixture fault" {
							t.Fatalf("panic = %v", got)
						}
					}()
					castStun52C2C0(uint64(2), 1, h)
				}()
				if calls != fault {
					t.Fatalf("calls = %d, want prefix %d", calls, fault)
				}
			})
		}
	}
}
