package server

import (
	"math"
	"testing"

	"github.com/opennox/libs/object"
)

// Select unchanged-original observations for the exact existing test inputs.
func selfStrikeOriginal4E17B0(t *testing.T, sourcePlayer, player bool, typ object.DamageType, raw int32, armor, carry float32, quest bool, scale float32) physicalOriginalExpected4E17B0 {
	t.Helper()
	rows := selfStrikeStartupOriginalRows4E17B0(t, "testdata/self_strike_compatibility_original.json", 764)
	key := [8]uint32{0, 0, uint32(typ), uint32(raw), math.Float32bits(armor), math.Float32bits(carry), 0, math.Float32bits(scale)}
	if sourcePlayer {
		key[0] = 1
	}
	if player {
		key[1] = 1
	}
	if quest {
		key[6] = 1
	}
	for _, row := range rows {
		if [8]uint32(row[:8]) == key {
			return physicalOriginalExpected4E17B0{Damage: int32(row[10]), Carry: row[11], Armor: int32(row[9])}
		}
	}
	t.Fatalf("missing original self-strike input sourcePlayer=%t player=%t type=%d raw=%d armor=%g carry=%g quest=%t scale=%g", sourcePlayer, player, typ, raw, armor, carry, quest, scale)
	return physicalOriginalExpected4E17B0{}
}
