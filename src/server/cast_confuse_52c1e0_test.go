package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

func TestCastConfuse52C1E0ExactTraceAndLiveTargets(t *testing.T) {
	const caster, gated, applied, attributed = uint64(0x1111222233334444), uint64(0x5555666677778888), uint64(0x9999aaaabbbbcccc), uint64(0xdddd111122223333)
	var trace []string
	target := gated
	h := castConfuseHooks52C1E0[uint64]{
		target: func() uint64 { trace = append(trace, fmt.Sprintf("target:%x", target)); return target },
		balance: func(key string) float64 {
			trace = append(trace, "balance:"+key)
			target = applied
			return 32767.5
		},
		floatToInt: func(value float32) int32 {
			trace = append(trace, fmt.Sprintf("convert:%g", value))
			return aiPathFloatToInt419A70(value)
		},
		apply: func(unit uint64, buff int32, duration int16, power int8) {
			trace = append(trace, fmt.Sprintf("apply:%x/%d/%d/%d", unit, buff, duration, power))
			target = attributed
		},
		attribution: func(source, unit uint64) { trace = append(trace, fmt.Sprintf("attribute:%x/%x", source, unit)) },
	}
	if got := castConfuse52C1E0(caster, 0x180, h); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{
		"target:5555666677778888", "balance:ConfuseEnchantDuration", "convert:32767.5",
		"target:9999aaaabbbbcccc", "apply:9999aaaabbbbcccc/3/-32768/-128",
		"target:dddd111122223333", "attribute:1111222233334444/dddd111122223333",
	}
	if !reflect.DeepEqual(trace, want) {
		t.Fatalf("trace = %v, want %v", trace, want)
	}
}

func TestCastConfuse52C1E0NilGateAndLiveNilReloads(t *testing.T) {
	for _, initiallyNil := range []bool{true, false} {
		t.Run(fmt.Sprintf("initial-nil=%t", initiallyNil), func(t *testing.T) {
			target := uint64(0x1111222233334444)
			if initiallyNil {
				target = 0
			}
			balanceCalls, applyCalls, attributeCalls := 0, 0, 0
			h := castConfuseHooks52C1E0[uint64]{
				target:     func() uint64 { return target },
				balance:    func(key string) float64 { balanceCalls++; target = 0; return 2.5 },
				floatToInt: aiPathFloatToInt419A70,
				apply: func(unit uint64, buff int32, duration int16, power int8) {
					applyCalls++
					if unit != 0 || buff != 3 || duration != 2 || power != 1 {
						t.Fatalf("apply = %x/%d/%d/%d", unit, buff, duration, power)
					}
				},
				attribution: func(source, unit uint64) {
					attributeCalls++
					if source != 0 || unit != 0 {
						t.Fatal("unexpected attribution")
					}
				},
			}
			want := int32(1)
			if initiallyNil {
				want = 0
			}
			if got := castConfuse52C1E0(uint64(0), 1, h); got != want || balanceCalls != int(want) || applyCalls != int(want) || attributeCalls != int(want) {
				t.Fatalf("result/calls = %d/%d/%d/%d, want %d each", got, balanceCalls, applyCalls, attributeCalls, want)
			}
		})
	}
}

func TestCastConfuse52C1E0Binary32RoundingAndFixedWidthBoundaries(t *testing.T) {
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
			h := castConfuseHooks52C1E0[uint64]{
				target:  func() uint64 { return 0x1111222233334444 },
				balance: func(string) float64 { return tc.value }, floatToInt: aiPathFloatToInt419A70,
				apply: func(_ uint64, _ int32, duration int16, power int8) {
					if duration != tc.duration || power != tc.wantPower {
						t.Fatalf("duration/power = %d/%d, want %d/%d", duration, power, tc.duration, tc.wantPower)
					}
				},
				attribution: func(uint64, uint64) {},
			}
			if got := castConfuse52C1E0(uint64(0), tc.power, h); got != 1 {
				t.Fatalf("result = %d", got)
			}
		})
	}
}

func TestCastConfuse52C1E0FaultPrefixes(t *testing.T) {
	for fault := 1; fault <= 7; fault++ {
		t.Run(fmt.Sprint(fault), func(t *testing.T) {
			calls := 0
			step := func() {
				calls++
				if calls == fault {
					panic("fixture fault")
				}
			}
			h := castConfuseHooks52C1E0[uint64]{
				target: func() uint64 { step(); return 1 }, balance: func(string) float64 { step(); return 5 },
				floatToInt: func(float32) int32 { step(); return 5 }, apply: func(uint64, int32, int16, int8) { step() },
				attribution: func(uint64, uint64) { step() },
			}
			func() {
				defer func() {
					if got := recover(); got != "fixture fault" {
						t.Fatalf("panic = %v", got)
					}
				}()
				castConfuse52C1E0(uint64(2), 1, h)
			}()
			if calls != fault {
				t.Fatalf("calls = %d, want prefix %d", calls, fault)
			}
		})
	}
}
