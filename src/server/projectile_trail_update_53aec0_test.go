package server

import (
	"testing"

	"github.com/opennox/libs/types"
)

func TestProjectileTrailUpdate53AEC0CreatesEightNativeOwnedSparks(t *testing.T) {
	projectile := &Object{
		PosVec:     types.Ptf(18, 22),
		PrevPos:    types.Ptf(10.5, 10.5),
		ForceVec:   types.Ptf(1, -2),
		Direction1: 0,
		SpeedCur:   8,
	}
	var sparks []types.Pointf
	var randomRanges []types.Pointf
	runtime := ProjectileTrailRuntime53AEC0{
		RandomFloat: func(min, max float32) float32 {
			randomRanges = append(randomRanges, types.Ptf(min, max))
			return 0
		},
		CreateSpark: func(pos types.Pointf, kind, lifetime int, velocity types.Pointf, z float32, owner *Object) {
			if kind != 1 || lifetime != 6 || velocity != (types.Pointf{}) || z != 0 || owner != projectile {
				t.Fatalf("spark = %v/%d/%d/%v/%g/%p", pos, kind, lifetime, velocity, z, owner)
			}
			sparks = append(sparks, pos)
		},
	}
	ProjectileTrailUpdate53AEC0(projectile, runtime)
	if got, want := projectile.ForceVec, (types.Ptf(3, -2)); got != want {
		t.Fatalf("force = %v, want %v", got, want)
	}
	if len(sparks) != 8 || len(randomRanges) != 32 {
		t.Fatalf("sparks/random calls = %d/%d", len(sparks), len(randomRanges))
	}
	for i, want := range []types.Pointf{types.Ptf(10, 10), types.Ptf(12, 13), types.Ptf(14, 16), types.Ptf(16, 19)} {
		if sparks[i*2] != want || sparks[i*2+1] != want {
			t.Fatalf("spark %d positions = %v/%v, want %v", i, sparks[i*2], sparks[i*2+1], want)
		}
	}
	for i, got := range randomRanges {
		want := types.Ptf(-2, 2)
		if i%4 >= 2 {
			want = types.Ptf(-4, 4)
		}
		if got != want {
			t.Fatalf("random range %d = %v, want %v", i, got, want)
		}
	}
}
