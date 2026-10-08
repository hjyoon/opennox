package legacy

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"testing"

	"github.com/opennox/libs/prand"
	"github.com/opennox/opennox/v1/server"
)

func commonRandomFloatFixture416090(t *testing.T, index int) (*server.Server, *int) {
	t.Helper()
	world, lookups := commonRandomFloatFixture416030(t, 93)
	world.Rand.Other = prand.New(index)
	return world, lookups
}

// Finite reference operations use math/big at precision 53/ToZero through the
// existing independent test model, never production rounding helpers.
func commonRandomFloatReferenceOp416090(a, b float64, operation byte) float64 {
	if !math.IsNaN(a) && !math.IsNaN(b) && !math.IsInf(a, 0) && !math.IsInf(b, 0) {
		return commonRandomFloatReferenceOp416030(a, b, operation)
	}
	switch operation {
	case '-':
		return a - b
	case '*':
		return a * b
	default:
		return a + b
	}
}

func commonRandomFloatReference416090(min, max float32, word int) float64 {
	delta := commonRandomFloatReferenceOp416090(float64(max), float64(min), '-')
	scaled := commonRandomFloatReferenceOp416090(float64(word), float64(math.Float32frombits(0x38000100)), '*')
	return commonRandomFloatReferenceOp416090(commonRandomFloatReferenceOp416090(delta, scaled, '*'), float64(min), '+')
}

func TestCommonRandomFloat416090CEntryOriginalGameplayPrecision(t *testing.T) {
	for _, bounds := range []struct {
		name     string
		min, max float32
	}{
		{"pi", math.Float32frombits(0xc0490fdb), math.Float32frombits(0x40490fdb)},
		{"light", 0, 100}, {"unit", 0, 1}, {"negative-unit", -1, 0},
		{"reverse", 3, -2}, {"motion", -2, 3}, {"speed", 12, 25},
		{"dense", math.Float32frombits(0xc2012345), math.Float32frombits(0x433abcde)},
		{"wide", math.SmallestNonzeroFloat32, math.MaxFloat32},
		{"negative-wide", -math.MaxFloat32, -math.SmallestNonzeroFloat32},
		{"opposite-extremes", -math.MaxFloat32, math.MaxFloat32},
		{"subnormal", math.Float32frombits(7), math.Float32frombits(0x00012345)},
		{"negative-zero-bound", math.Float32frombits(0x80000000), 12},
	} {
		t.Run(bounds.name, func(t *testing.T) {
			world, lookups := commonRandomFloatFixture416090(t, 0)
			table := prand.New(0)
			for index := 0; index < 4096; index++ {
				word := table.Int(0, 0x7fff)
				want := commonRandomFloatReference416090(bounds.min, bounds.max, word)
				got := commonRandomFloatCEntry416090(bounds.min, bounds.max)
				if math.Float64bits(got) != math.Float64bits(want) {
					t.Fatalf("C return at table[%d]=%d: %016x want %016x", index, word, math.Float64bits(got), math.Float64bits(want))
				}
				if world.Rand.Other.Index() != (index+1)%4096 || world.Rand.Logic.Index() != 93 || *lookups != index+1 {
					t.Fatalf("C entry changed stream/lookup order: other=%d logic=%d lookups=%d", world.Rand.Other.Index(), world.Rand.Logic.Index(), *lookups)
				}
			}
		})
	}
}

func TestCommonRandomFloat416090CEntryAlwaysSteps(t *testing.T) {
	for _, bounds := range []struct {
		name     string
		min, max uint32
	}{
		{"positive-zero", 0, 0}, {"negative-zero", 0x80000000, 0x80000000},
		{"minus-plus-zero", 0x80000000, 0}, {"plus-minus-zero", 0, 0x80000000},
		{"equal-finite", 0x3f800000, 0x3f800000},
		{"equal-positive-infinity", 0x7f800000, 0x7f800000},
		{"equal-negative-infinity", 0xff800000, 0xff800000},
		{"quiet-nan-min", 0x7fc12345, 0x40000000},
		{"signaling-nan-min", 0x7f812345, 0x40000000},
		{"quiet-nan-max", 0x3f800000, 0xffc12345},
		{"two-nans", 0x7fc12345, 0xffc12345},
		{"positive-infinity-max", 0x3f800000, 0x7f800000},
		{"negative-infinity-max", 0x3f800000, 0xff800000},
		{"opposite-infinities", 0xff800000, 0x7f800000},
	} {
		t.Run(bounds.name, func(t *testing.T) {
			world, lookups := commonRandomFloatFixture416090(t, 0)
			table := prand.New(0)
			min, max := math.Float32frombits(bounds.min), math.Float32frombits(bounds.max)
			for index := 0; index < 4096; index++ {
				word := table.Int(0, 0x7fff)
				want := commonRandomFloatReference416090(min, max, word)
				got := commonRandomFloatCEntry416090(min, max)
				if math.IsNaN(want) {
					if !math.IsNaN(got) {
						t.Fatalf("table[%d]=%d: exceptional C return=%v want NaN", index, word, got)
					}
				} else if math.Float64bits(got) != math.Float64bits(want) {
					t.Fatalf("table[%d]=%d: C return=%016x want=%016x", index, word, math.Float64bits(got), math.Float64bits(want))
				}
				if world.Rand.Other.Index() != (index+1)%4096 || world.Rand.Logic.Index() != 93 || *lookups != index+1 {
					t.Fatalf("unconditional C step at %d: other=%d logic=%d lookups=%d", index, world.Rand.Other.Index(), world.Rand.Logic.Index(), *lookups)
				}
			}
		})
	}
}

func TestCommonRandomFloat416090CEntryOriginalTableAndPiSeals(t *testing.T) {
	world, lookups := commonRandomFloatFixture416090(t, 0)
	table := prand.New(0)
	tableBytes, piBytes := make([]byte, 4096*4), make([]byte, 4096*4)
	for index := 0; index < 4096; index++ {
		binary.LittleEndian.PutUint32(tableBytes[4*index:], uint32(table.Int(0, 0x7fff)))
		value := commonRandomFloatCEntry416090(math.Float32frombits(0xc0490fdb), math.Float32frombits(0x40490fdb))
		binary.LittleEndian.PutUint32(piBytes[4*index:], math.Float32bits(float32(value)))
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(tableBytes)); got != "7dfccdce34d74e348fae4181630a4c41a20a560432658ba2526969bb7000605f" {
		t.Fatalf("table seal = %s", got)
	}
	// Preserve the existing binary32 nearest-spill Pi golden separately from
	// the retained double and renderer-specific Chop24 comparisons.
	if got := fmt.Sprintf("%x", sha256.Sum256(piBytes)); got != "3c278b2efdbe095d912b42ae8fb459963f76ec819a72d3962bb1425badffaf25" {
		t.Fatalf("original Pi golden = %s", got)
	}
	if world.Rand.Other.Index() != 0 || world.Rand.Logic.Index() != 93 || *lookups != 4096 {
		t.Fatal("whole-table C entry changed wrap/logic stream/lookups")
	}
}
