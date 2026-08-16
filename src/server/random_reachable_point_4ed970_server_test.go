package server

import (
	"math"
	"testing"

	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/types"
)

func TestRandomReachablePointServerDeps4ED970UseExactLogicRNG(t *testing.T) {
	s := new(Server)
	s.Rand.Logic = prand.New(0)
	deps := randomReachablePointServerDeps4ED970(s)
	got := deps.randomFloat(
		-randomReachablePointPi4ED970,
		randomReachablePointPi4ED970,
		randomReachablePointSource4ED970,
		randomReachablePointSourceLine4ED970,
	)
	if bits := math.Float64bits(got); bits != 0x40040900c579a06c {
		t.Fatalf("random result bits = %#016x, want 0x40040900c579a06c", bits)
	}
	if index := s.Rand.Logic.Index(); index != 1 {
		t.Fatalf("logic index = %d, want 1", index)
	}
}

func TestRandomReachablePointAroundInto4ED970ReturnsExactOutputPointer(t *testing.T) {
	s := new(Server)
	s.Rand.Logic = prand.New(0)
	center := &types.Pointf{X: -1000, Y: -1000}
	output := &types.Pointf{X: 1, Y: 2}

	got := s.RandomReachablePointAroundInto4ED970(64, center, output)
	if got != output || *got != *center {
		t.Fatalf("result = %p %+v, want output %p and center %+v", got, *got, output, *center)
	}
	if index := s.Rand.Logic.Index(); index != 1 {
		t.Fatalf("logic index = %d, want 1", index)
	}
}

func TestRandomReachablePointAround4ED970ValueWrapper(t *testing.T) {
	s := new(Server)
	s.Rand.Logic = prand.New(0)
	center := types.Pointf{X: -1000, Y: -1000}
	if got := s.RandomReachablePointAround(64, center); got != center {
		t.Fatalf("result = %+v, want %+v", got, center)
	}
	if index := s.Rand.Logic.Index(); index != 1 {
		t.Fatalf("logic index = %d, want 1", index)
	}
}
