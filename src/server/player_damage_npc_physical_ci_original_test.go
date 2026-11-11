package server

import (
	"encoding/json"
	"math"
	"os"
	"testing"
)

type npcPhysicalCIOriginal4E17B0 struct {
	Name    string `json:"name"`
	Mode    string `json:"mode"`
	Raw     int32  `json:"raw"`
	Armor   uint32 `json:"armor_bits32"`
	Initial uint32 `json:"initial_carry_bits32"`
	Damage  int32  `json:"damage_bits32"`
	Wear    int32  `json:"wear_bits32"`
	Carry   uint32 `json:"carry_bits32"`
}

// These exact historical inputs execute unchanged GAME.EXE PlayerDamage,
// carry and FISTP under its actual startup control word, rather than host
// RoundToEven. The adjacent script reproduces both startup and nearest controls.
func npcPhysicalOriginal4E17B0(t *testing.T, name string, raw int32, armor, carry float32) npcPhysicalCIOriginal4E17B0 {
	t.Helper()
	b, err := os.ReadFile("testdata/npc_physical_carry_ci_original_4e17b0.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Startup uint16                        `json:"original_startup_FPCW"`
		Rows    []npcPhysicalCIOriginal4E17B0 `json:"rows"`
	}
	if err = json.Unmarshal(b, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Startup != 0xe3f || len(fixture.Rows) != 32 {
		t.Fatal("incomplete original NPC physical CI controls")
	}
	for _, row := range fixture.Rows {
		if row.Name == name && row.Mode == "actual-startup" {
			if row.Raw != raw || row.Armor != math.Float32bits(armor) || row.Initial != math.Float32bits(carry) {
				t.Fatal("original NPC physical input differs from test")
			}
			return row
		}
	}
	t.Fatal("missing original NPC physical CI input", name)
	return npcPhysicalCIOriginal4E17B0{}
}
