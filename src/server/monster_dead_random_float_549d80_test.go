package server

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/opennox/libs/prand"
)

func TestMonsterDeadRandomFloat549D80OriginalCallerSpills(t *testing.T) {
	digest := sha256.New()
	for _, bounds := range []struct {
		name     string
		min, max float32
	}{
		{"golem-height", 10, 70}, {"vertical-velocity", -2, 0},
		{"golem-force", 5, 20}, {"bone-height", 10, 35}, {"bone-force", 5, 25},
	} {
		t.Run(bounds.name, func(t *testing.T) {
			s := new(Server)
			s.SetFrame(1400)
			s.Rand.Logic, s.Rand.Other = prand.New(0), prand.New(17)
			table := prand.New(0)
			for index := range 4096 {
				word := table.Int(0, 0x7fff)
				retained := logicRandomFloatReferenceValue416030(bounds.min, bounds.max, word)
				// All five live ranges produce normal binary32 results.
				// Independent 24-bit ToZero models their original FSTPS.
				want, _ := new(big.Float).SetPrec(24).SetMode(big.ToZero).SetFloat64(retained).Float32()
				got := s.MonsterDeadRandomFloat549D80(bounds.min, bounds.max)
				if math.Float32bits(got) != math.Float32bits(want) {
					t.Fatalf("table[%d]=%d: stored=%08x want original=%08x", index, word, math.Float32bits(got), math.Float32bits(want))
				}
				after := (index + 1) % 4096
				if s.Rand.Logic.Index() != after || s.Rand.Other.Index() != 17 || s.Frame() != 1400 {
					t.Fatal("caller spill changed stream/index/frame")
				}
				var row [20]byte
				for i, value := range [...]uint32{math.Float32bits(bounds.min), math.Float32bits(bounds.max), uint32(index), math.Float32bits(got), uint32(after)} {
					binary.LittleEndian.PutUint32(row[4*i:], value)
				}
				digest.Write(row[:])
			}
		})
	}
	// Unchanged original GAME.EXE 00416030 + 00549F20 FSTPS, all 4096
	// table indices in each live range, gameplay FPCW 0x0e7f, no code hooks.
	got := fmt.Sprintf("%x", digest.Sum(nil))
	if got != "e5368033a096fdb9a47afa2d3171d9f9ae45c91499fddb8994e7ed4b95ca2425" {
		t.Fatalf("original instruction execution checksum mismatch: %s", got)
	}
	t.Logf("original FSTPS comparison: rows=20480 result_sha256=%s", got)
}

func TestMonsterDeadRandomFloat549D80NoStep(t *testing.T) {
	for _, bounds := range [][2]uint32{
		{0, 0}, {0x80000000, 0x80000000}, {0x80000000, 0}, {0, 0x80000000},
		{0x3f800000, 0x3f800000}, {0x7f800000, 0x7f800000}, {0xff800000, 0xff800000},
		{0x7fc12345, 0x40000000}, {0x7f812345, 0x40000000},
		{0x3f800000, 0xffc12345}, {0x7fc12345, 0xffc12345},
	} {
		for _, index := range []int{17, 4095} {
			t.Run(fmt.Sprintf("%08x-%08x/at-%d", bounds[0], bounds[1], index), func(t *testing.T) {
				s := new(Server)
				s.SetFrame(1400)
				s.Rand.Logic, s.Rand.Other = prand.New(index), prand.New(17)
				min, max := math.Float32frombits(bounds[0]), math.Float32frombits(bounds[1])
				got := s.MonsterDeadRandomFloat549D80(min, max)
				if math.IsNaN(float64(max)) {
					if !math.IsNaN(float64(got)) {
						t.Fatalf("unordered max produced %v", got)
					}
				} else if math.Float32bits(got) != bounds[1] {
					t.Fatalf("stored max=%08x want %08x", math.Float32bits(got), bounds[1])
				}
				if s.Rand.Logic.Index() != index || s.Rand.Other.Index() != 17 || s.Frame() != 1400 {
					t.Fatal("no-step path changed stream/index/frame")
				}
			})
		}
	}
}
