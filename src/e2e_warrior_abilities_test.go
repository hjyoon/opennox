package opennox

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func TestE2EWarriorAbilityHUDDecode(t *testing.T) {
	for ability := server.AbilityBerserk; ability <= server.AbilityInfravis; ability++ {
		t.Run(ability.String(), func(t *testing.T) {
			data := make([]byte, 24)
			values := []uint32{uint32(ability), 0xdeadbeef, 1, 1 << ability, 5, 0xfedcba98}
			for i, v := range values {
				binary.LittleEndian.PutUint32(data[i*4:], v)
			}
			got, err := e2eDecodeAbilityHUD(data)
			want := e2eAbilityHUD{uint32(ability), 1, 1 << ability, 5, 0xfedcba98}
			if err != nil || got != want {
				t.Fatalf("HUD = %+v, %v; want %+v", got, err, want)
			}
		})
	}
	for size := 0; size <= 25; size++ {
		if size == 24 {
			continue
		}
		if _, err := e2eDecodeAbilityHUD(make([]byte, size)); err == nil {
			t.Fatalf("accepted malformed %d-byte HUD record", size)
		}
	}
}

func TestE2EWarriorAbilityInputKeys(t *testing.T) {
	for i, want := range []keybind.Key{keybind.KeyA, keybind.KeyS, keybind.KeyD, keybind.KeyF, keybind.KeyG} {
		ability := server.Ability(i + 1)
		got, err := e2eWarriorAbilityKey(ability)
		if err != nil || got != want {
			t.Fatalf("ability %d key = %v, %v; want %v", ability, got, err, want)
		}
	}
	for _, ability := range []server.Ability{-1, 0, server.AbilityMax, 100} {
		if _, err := e2eWarriorAbilityKey(ability); err == nil {
			t.Fatalf("accepted invalid ability %d", ability)
		}
	}
}

func TestE2EWarriorAbilityArena(t *testing.T) {
	original := types.Ptf(200, 300)
	t.Run("open original position", func(t *testing.T) {
		pos, direction, err := e2eWarriorAbilityArena(original, 14, func(types.Pointf, types.Pointf) bool { return true })
		if err != nil || pos != original || direction != types.Ptf(1, 0) {
			t.Fatalf("arena = %v, %v, %v", pos, direction, err)
		}
	})
	t.Run("nearby lane when spawn is confined", func(t *testing.T) {
		trace := func(from, to types.Pointf) bool {
			return from != original || math.Hypot(float64(to.X-from.X), float64(to.Y-from.Y)) < 70
		}
		pos, direction, err := e2eWarriorAbilityArena(original, 14, trace)
		if err != nil || pos == original || direction != types.Ptf(1, 0) || !trace(original, pos) {
			t.Fatalf("arena = %v, %v, %v", pos, direction, err)
		}
	})
	t.Run("blocked map fails closed", func(t *testing.T) {
		if _, _, err := e2eWarriorAbilityArena(original, 14, func(types.Pointf, types.Pointf) bool { return false }); err == nil {
			t.Fatal("selected an arena through blocked rays")
		}
	})
	t.Run("center ray alone is insufficient", func(t *testing.T) {
		trace := func(from, to types.Pointf) bool { return from.Y == original.Y && to.Y == original.Y }
		if _, _, err := e2eWarriorAbilityArena(original, 14, trace); err == nil {
			t.Fatal("selected a lane too narrow for a player")
		}
	})
}

func TestE2EWarriorAbilityLaneCircleClearance(t *testing.T) {
	for _, tc := range []struct {
		name           string
		from, to, prop types.Pointf
		radius         float32
		clear          bool
	}{
		{"direct obstacle", types.Ptf(0, 0), types.Ptf(10, 0), types.Ptf(5, 0), 1, false},
		{"nearby obstacle", types.Ptf(0, 0), types.Ptf(10, 0), types.Ptf(5, 2), 1, true},
		{"tangent blocks", types.Ptf(0, 0), types.Ptf(10, 0), types.Ptf(5, 1), 1, false},
		{"behind origin", types.Ptf(0, 0), types.Ptf(10, 0), types.Ptf(-2, 0), 1, true},
		{"beyond target", types.Ptf(0, 0), types.Ptf(10, 0), types.Ptf(12, 0), 1, true},
		{"end overlap", types.Ptf(0, 0), types.Ptf(10, 0), types.Ptf(11, 0), 1, false},
		{"diagonal", types.Ptf(0, 0), types.Ptf(10, 10), types.Ptf(5, 5), 1, false},
		{"point inside", types.Ptf(5, 5), types.Ptf(5, 5), types.Ptf(5, 5), 1, false},
		{"point outside", types.Ptf(5, 5), types.Ptf(5, 5), types.Ptf(7, 5), 1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := e2eWarriorLaneMissesCircle(tc.from, tc.to, tc.prop, tc.radius); got != tc.clear {
				t.Fatalf("clear = %t, want %t", got, tc.clear)
			}
		})
	}
}
