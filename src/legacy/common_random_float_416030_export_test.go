package legacy

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/opennox/libs/prand"
	"github.com/opennox/opennox/v1/server"
)

type commonRandomFloatLegacyServer416030 struct {
	Server
	world *server.Server
}

func (s *commonRandomFloatLegacyServer416030) S() *server.Server { return s.world }

func commonRandomFloatFixture416030(t *testing.T, index int) (*server.Server, *int) {
	t.Helper()
	s := new(server.Server)
	s.Rand.Logic, s.Rand.Other = prand.New(index), prand.New(93)
	previous := GetServer
	lookups := 0
	GetServer = func() Server {
		lookups++
		return &commonRandomFloatLegacyServer416030{world: s}
	}
	t.Cleanup(func() { GetServer = previous })
	return s, &lookups
}

// Independent x87 instruction-order reference: binary32 operands and scale,
// 53-bit ToZero intermediate registers, and no binary32 return-value spill.
func commonRandomFloatReferenceOp416030(a, b float64, operation byte) float64 {
	x := new(big.Float).SetPrec(53).SetMode(big.ToZero).SetFloat64(a)
	y := new(big.Float).SetPrec(53).SetMode(big.ToZero).SetFloat64(b)
	switch operation {
	case '-':
		x.Sub(x, y)
	case '*':
		x.Mul(x, y)
	default:
		x.Add(x, y)
	}
	result, _ := x.Float64()
	return result
}

func commonRandomFloatReference416030(min, max float32, word int) float64 {
	delta := commonRandomFloatReferenceOp416030(float64(max), float64(min), '-')
	scaled := commonRandomFloatReferenceOp416030(float64(word), float64(math.Float32frombits(0x38000100)), '*')
	return commonRandomFloatReferenceOp416030(commonRandomFloatReferenceOp416030(delta, scaled, '*'), float64(min), '+')
}

func TestCommonRandomFloat416030CEntryOriginalGameplayPrecision(t *testing.T) {
	for _, bounds := range []struct {
		name     string
		min, max float32
	}{
		{"pi", math.Float32frombits(0xc0490fdb), math.Float32frombits(0x40490fdb)},
		{"unit", 0, 1}, {"negative-unit", -1, 0}, {"reverse", 3, -2},
		{"motion", -2, 3}, {"speed", 12, 25},
		{"dense", math.Float32frombits(0xc2012345), math.Float32frombits(0x433abcde)},
		{"wide", math.SmallestNonzeroFloat32, math.MaxFloat32},
		{"negative-wide", -math.MaxFloat32, -math.SmallestNonzeroFloat32},
		{"opposite-extremes", -math.MaxFloat32, math.MaxFloat32},
		{"subnormal", math.Float32frombits(7), math.Float32frombits(0x00012345)},
		{"negative-zero-bound", math.Float32frombits(0x80000000), 12},
	} {
		t.Run(bounds.name, func(t *testing.T) {
			world, lookups := commonRandomFloatFixture416030(t, 0)
			table := prand.New(0)
			for index := 0; index < 4096; index++ {
				word := table.Int(0, 0x7fff)
				want := commonRandomFloatReference416030(bounds.min, bounds.max, word)
				got := commonRandomFloatCEntry416030(bounds.min, bounds.max)
				if math.Float64bits(got) != math.Float64bits(want) {
					t.Fatalf("C return at table[%d]=%d: %016x want %016x", index, word, math.Float64bits(got), math.Float64bits(want))
				}
				if world.Rand.Logic.Index() != (index+1)%4096 || world.Rand.Other.Index() != 93 || *lookups != index+1 {
					t.Fatalf("C entry changed stream/lookup order: logic=%d other=%d lookups=%d", world.Rand.Logic.Index(), world.Rand.Other.Index(), *lookups)
				}
			}
		})
	}
}

func TestCommonRandomFloat416030CEntryNoStep(t *testing.T) {
	for _, bounds := range [][2]uint32{
		{0, 0}, {0x80000000, 0x80000000}, {0x80000000, 0}, {0, 0x80000000},
		{0x3f800000, 0x3f800000}, {0x7f800000, 0x7f800000}, {0xff800000, 0xff800000},
		{0x7fc12345, 0x40000000}, {0x7f812345, 0x40000000},
		{0x3f800000, 0xffc12345}, {0x7fc12345, 0xffc12345},
	} {
		for _, index := range []int{17, 4095} {
			t.Run(fmt.Sprintf("%08x-%08x/at-%d", bounds[0], bounds[1], index), func(t *testing.T) {
				world, lookups := commonRandomFloatFixture416030(t, index)
				min, max := math.Float32frombits(bounds[0]), math.Float32frombits(bounds[1])
				got := commonRandomFloatCEntry416030(min, max)
				if math.IsNaN(float64(max)) {
					if !math.IsNaN(got) {
						t.Fatalf("unordered result = %v", got)
					}
				} else if math.Float64bits(got) != math.Float64bits(float64(max)) {
					t.Fatalf("no-step result=%016x want max=%016x", math.Float64bits(got), math.Float64bits(float64(max)))
				}
				if world.Rand.Logic.Index() != index || world.Rand.Other.Index() != 93 || *lookups != 1 {
					t.Fatalf("no-step C entry changed state: logic=%d other=%d lookups=%d", world.Rand.Logic.Index(), world.Rand.Other.Index(), *lookups)
				}
			})
		}
	}
}

func TestCommonRandomFloat416030CEntryOriginalTableAndPiSeals(t *testing.T) {
	world, lookups := commonRandomFloatFixture416030(t, 0)
	table := prand.New(0)
	tableBytes, piBytes := make([]byte, 4096*4), make([]byte, 4096*4)
	for index := 0; index < 4096; index++ {
		binary.LittleEndian.PutUint32(tableBytes[4*index:], uint32(table.Int(0, 0x7fff)))
		value := commonRandomFloatCEntry416030(math.Float32frombits(0xc0490fdb), math.Float32frombits(0x40490fdb))
		binary.LittleEndian.PutUint32(piBytes[4*index:], math.Float32bits(float32(value)))
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(tableBytes)); got != "7dfccdce34d74e348fae4181630a4c41a20a560432658ba2526969bb7000605f" {
		t.Fatalf("table seal = %s", got)
	}
	// Preserve the existing binary32 nearest-spill golden, independently of
	// the retained double-precision return comparisons above.
	if got := fmt.Sprintf("%x", sha256.Sum256(piBytes)); got != "3c278b2efdbe095d912b42ae8fb459963f76ec819a72d3962bb1425badffaf25" {
		t.Fatalf("original Pi golden = %s", got)
	}
	if world.Rand.Logic.Index() != 0 || world.Rand.Other.Index() != 93 || *lookups != 4096 {
		t.Fatal("whole-table C entry changed wrap/other stream/lookups")
	}
}
