package server

import (
	"math"
	"testing"
)

func TestQuestDifficulty4E3D50OriginalPrecision(t *testing.T) {
	// GAME.EXE 004E3D50 with the gameplay x87 control word 0x0e3f.
	for _, test := range []struct {
		stage   uint32
		players int32
		delta   uint32
		want    uint32
	}{
		{0, 0, 0x3f800001, 0x80000000},
		{0, 7, 0xff800000, 0xffc00000},
		{3, 0, 0x3e199999, 0x40233333},
		{3, 2, 0xbf800000, 0},
		{3, 2, 0xff812345, 0xffc12345},
		{3, 7, 0x3e199999, 0x40b66666},
		{0x01000001, 2, 0x3e199999, 0x4b933333},
		{0x80000000, 1, 0x3f800000, 0x4f000000},
		{0xffffffff, 7, 0xbe199999, 0x4dccccd3},
		{0xffffffff, 7, 0x7f7fffff, 0x7f7fffff},
	} {
		if got := math.Float32bits(QuestDifficulty4E3D50(test.stage, test.players, math.Float32frombits(test.delta))); got != test.want {
			t.Errorf("stage=%08x players=%d delta=%08x result=%08x, want %08x", test.stage, test.players, test.delta, got, test.want)
		}
	}
}
