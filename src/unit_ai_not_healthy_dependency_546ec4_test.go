package opennox

import (
	"fmt"
	"math"
	"math/big"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Independent integer / big.Float reference: no production FMA or Nextafter.
func aiNotHealthyReferenceFraction546EC4(current, maximum uint16) float64 {
	fraction := new(big.Float).SetPrec(53).SetMode(big.ToZero)
	if maximum == 0 {
		fraction.SetUint64(1)
	} else {
		fraction.Quo(new(big.Float).SetUint64(uint64(current)), new(big.Float).SetUint64(uint64(maximum)))
	}
	value, accuracy := fraction.Float64()
	if accuracy != big.Exact {
		panic("53-bit health reference outside binary64 exponent range")
	}
	return value
}

func aiNotHealthyCheckFractionModel546EC4(t *testing.T, current, maximum uint16) {
	t.Helper()
	health := server.HealthData{Cur: current, Max: maximum, Field2: 0xabcd, Field16: 0x12345678}
	before := health
	want := aiNotHealthyReferenceFraction546EC4(current, maximum)
	got := aiDependencyHealthFraction546EC4(&health)
	if math.Float64bits(got) != math.Float64bits(want) {
		t.Fatalf("unsigned WORD fraction %d/%d bits=%016x want=%016x", current, maximum, math.Float64bits(got), math.Float64bits(want))
	}
	if health != before {
		t.Fatal("health fraction changed input words")
	}
}

func TestAINotHealthyDependency546EC4IndependentWordNumerators(t *testing.T) {
	for _, maximum := range []uint16{0, 1, 3, 10, 80, 150, 0x8000, 0xffff} {
		t.Run(fmt.Sprintf("max-%04x", maximum), func(t *testing.T) {
			for current := uint32(0); current < 1<<16; current++ {
				aiNotHealthyCheckFractionModel546EC4(t, uint16(current), maximum)
			}
		})
	}
}

func TestAINotHealthyDependency546EC4IndependentWordDenominators(t *testing.T) {
	for _, current := range []uint16{0, 1, 24, 45, 49, 0x8000, 0xfffe, 0xffff} {
		t.Run(fmt.Sprintf("cur-%04x", current), func(t *testing.T) {
			for maximum := uint32(0); maximum < 1<<16; maximum++ {
				aiNotHealthyCheckFractionModel546EC4(t, current, uint16(maximum))
			}
		})
	}
}

func TestAINotHealthyDependency546EC4RetainedFractionBits(t *testing.T) {
	for _, tc := range []struct {
		current, maximum uint16
		bits             uint64
	}{
		{0, 0, 0x3ff0000000000000},
		{0xffff, 0, 0x3ff0000000000000},
		{0, 80, 0},
		{40, 80, 0x3fe0000000000000},
		{24, 80, 0x3fd3333333333333},
		{45, 150, 0x3fd3333333333333},
		{49, 50, 0x3fef5c28f5c28f5c},
		{1, 80, 0x3f89999999999999},
		{1, 3, 0x3fd5555555555555},
		{2, 3, 0x3fe5555555555555},
		{1, 10, 0x3fb9999999999999},
		{76, 80, 0x3fee666666666666},
		{0x8000, 0xffff, 0x3fe0001000100010},
		{0xffff, 1, 0x40efffe000000000},
	} {
		t.Run(fmt.Sprintf("%d-%d", tc.current, tc.maximum), func(t *testing.T) {
			health := server.HealthData{Cur: tc.current, Max: tc.maximum}
			if got := math.Float64bits(aiDependencyHealthFraction546EC4(&health)); got != tc.bits {
				t.Fatalf("retained unsigned fraction bits=%016x want=%016x", got, tc.bits)
			}
		})
	}
}

func TestAINotHealthyDependency546EC4MissingInputFaultPrefix(t *testing.T) {
	s, unit, _, health := aiAliveDependencyNative546A70(t)
	unit.HealthData = health
	update := unit.UpdateDataMonster()
	update.ResumeLevel, update.Field97, update.Field101 = math.Float32frombits(0x3e99999a), 71, 72
	for _, tc := range []struct {
		name                                string
		noUnit, noHealth, noUpdate, maxZero bool
	}{
		{name: "nil-unit", noUnit: true},
		{name: "nil-health", noHealth: true},
		{name: "nil-update-after-quotient", noUpdate: true},
		{name: "nil-all", noUnit: true, noHealth: true, noUpdate: true},
		{name: "nil-health-and-update", noHealth: true, noUpdate: true},
		{name: "nil-unit-and-update", noUnit: true, noUpdate: true},
		{name: "nil-update-after-zero-max-fallback", noUpdate: true, maxZero: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			*health = server.HealthData{Cur: 24, Max: 80, Field2: 0xabcd, Field16: 0x12345678}
			if tc.maxZero {
				health.Max = 0
			}
			unit.HealthData = health
			if tc.noHealth {
				unit.HealthData = nil
			}
			selectedUnit, selectedUpdate := unit, update
			if tc.noUnit {
				selectedUnit = nil
			}
			if tc.noUpdate {
				selectedUpdate = nil
			}
			beforeHealth, beforeUpdate, beforePointer := *health, *update, unit.HealthData
			s.AI.StackChanged = false
			var fault any
			var result bool
			func() {
				defer func() { fault = recover() }()
				result = aiDependencyNotHealthy546EC4(selectedUnit, selectedUpdate)
			}()
			if fault == nil || result {
				t.Fatalf("missing-input prefix fault=%v result=%t, want fault before completed comparison", fault, result)
			}
			if *health != beforeHealth || *update != beforeUpdate || unit.HealthData != beforePointer || s.AI.StackChanged {
				t.Fatal("missing-input prefix changed native HP/update/identity/stack")
			}
		})
	}
}

func TestAINotHealthyDependency546EC4CachedUpdateAndLiveHealth(t *testing.T) {
	_, unit, _, health := aiAliveDependencyNative546A70(t)
	cached := unit.UpdateDataMonster()
	cached.ResumeLevel = math.Float32frombits(0x3e99999a)
	other, freeOther := alloc.New(server.MonsterUpdateData{ResumeLevel: 0, Field97: 91, Field101: 92})
	t.Cleanup(freeOther)
	replacement, freeReplacement := alloc.New(server.HealthData{Cur: 45, Max: 150, Field2: 0xabcd, Field16: 0x12345678})
	t.Cleanup(freeReplacement)
	for _, pointer := range []unsafe.Pointer{unsafe.Pointer(other), unsafe.Pointer(replacement)} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
			t.Fatalf("replacement native update/health below 4 GiB: %p", pointer)
		}
	}
	beforeUpdatePointer, beforeClass := unit.UpdateData, unit.ObjClass
	unit.UpdateData, unit.ObjClass = unsafe.Pointer(other), object.ClassFood
	defer func() { unit.UpdateData, unit.ObjClass = beforeUpdatePointer, beforeClass }()
	// Direct original case 64 does not inspect class or re-query UpdateData.
	// Replacement HP pointers remain live, while ResumeLevel is cached.
	for _, tc := range []struct {
		name         string
		selected     *server.HealthData
		cur, max     uint16
		valid, fault bool
	}{
		{"initial-health", health, 24, 80, true, false},
		{"replacement-health", replacement, 45, 150, true, false},
		{"replacement-health-full", replacement, 150, 150, false, false},
		{"zero-max-fallback", health, 0xffff, 0, false, false},
		{"missing-live-health", nil, 0, 0, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit.HealthData = tc.selected
			if tc.selected != nil {
				tc.selected.Cur, tc.selected.Max = tc.cur, tc.max
			}
			beforeHealth, beforeReplacement, beforeCached, beforeOther := *health, *replacement, *cached, *other
			var fault any
			var result bool
			func() {
				defer func() { fault = recover() }()
				result = aiDependencyNotHealthy546EC4(unit, cached)
			}()
			if (fault != nil) != tc.fault || result != tc.valid {
				t.Fatalf("cached update/live HP result=%t/%t fault=%v/%t", result, tc.valid, fault, tc.fault)
			}
			if *health != beforeHealth || *replacement != beforeReplacement || *cached != beforeCached || *other != beforeOther || unit.HealthData != tc.selected || unit.UpdateData != unsafe.Pointer(other) || unit.ObjClass != object.ClassFood {
				t.Fatal("health comparison changed cached/replacement native identities/data/class")
			}
		})
	}
}
