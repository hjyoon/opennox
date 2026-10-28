package server

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	playerlib "github.com/opennox/libs/player"
)

// The inputs were copied from an ordinary Wizard3 Load. GAME.EXE 004EEDC0,
// both stat getters and the real integer conversions supplied the expectations.
// Its actual startup PC53/ToZero mode produces 32/98 at level 3; a nearest-mode
// counterfactual produces 33/99 and must not be used as the gameplay oracle.
func TestPlayerReadValuesWizard3Original(t *testing.T) {
	var fixture struct {
		OriginalSHA string `json:"original_sha256"`
		StartupCW   uint16 `json:"startup_fpcw"`
		Inputs      struct {
			Class   uint8     `json:"class"`
			Flags   uint32    `json:"game_flags"`
			Base    [4]uint32 `json:"base_bits32"`
			Maximum [4]uint32 `json:"maximum_bits32"`
			Warrior [4]uint32 `json:"warrior_bits32"`
		} `json:"inputs"`
		Cases []struct {
			Level      uint8             `json:"level"`
			Result     map[string]uint32 `json:"result"`
			Protection []uint32          `json:"protection_values"`
		} `json:"cases"`
	}
	blob, err := os.ReadFile("testdata/player_read_values_wizard3_original.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.StartupCW != 0xe3f || fixture.Inputs.Class != uint8(playerlib.Wizard) || len(fixture.Cases) != 10 {
		t.Fatal("wrong original provenance or captured Wizard inputs")
	}
	stats := func(words [4]uint32) ClassStats {
		return ClassStats{Health: math.Float32frombits(words[0]), Mana: math.Float32frombits(words[1]), Speed: math.Float32frombits(words[2]), Strength: math.Float32frombits(words[3])}
	}
	for _, row := range fixture.Cases {
		t.Run(fmt.Sprintf("level%d", row.Level), func(t *testing.T) {
			players := &serverPlayers{}
			players.Stats.Base, players.Stats.Wizard, players.Stats.Warrior = stats(fixture.Inputs.Base), stats(fixture.Inputs.Maximum), stats(fixture.Inputs.Warrior)
			pl := &Player{Level: row.Level, field3652: 0xa5a5a5a5, ProtPlayerField2239: 0x11111111, ProtPlayerField2235: 0x22222222, ProtUnitManaMax: 0x33333333, ProtUnitHPMax: 0x44444444, ProtPlayerOrigName: 0x55555555}
			pl.Info().SetPlayerClass(playerlib.Wizard)
			pl.Info().SetName("")
			ud := &PlayerUpdateData{Player: pl, ManaCur: 99, ManaMax: 99}
			hp := &HealthData{Cur: 33, Max: 33}
			unit := &Object{UpdateData: unsafe.Pointer(ud), HealthData: hp}
			var protected []uint32
			var tokens []uint32
			deps := defaultPlayerReadValuesNativeDeps4EEDC0()
			deps.gameFlagsCheck = func(flag uint32) int32 { return int32(fixture.Inputs.Flags & flag) }
			deps.protectInt = func(token, value uint32) { tokens = append(tokens, token); protected = append(protected, value) }
			deps.protectUint16 = func(token uint32, value uint16) {
				tokens = append(tokens, token)
				protected = append(protected, uint32(value))
			}
			deps.protectName = func(info *PlayerInfo, size, token uint32) int32 {
				if info != pl.Info() || size != 0 || token != 0x55555555 {
					t.Fatal("wrong native name protection arguments")
				}
				return -0x1234567
			}
			result := playerReadValuesNative4EEDC0(unit, 0, players, deps)
			got := map[string]uint32{
				"health": uint32(hp.Cur), "maximum_health": uint32(hp.Max), "mana": uint32(ud.ManaCur), "maximum_mana": uint32(ud.ManaMax),
				"strength": pl.Info().Field2239(), "speed": pl.Info().Field2235(), "speed_base_bits32": math.Float32bits(unit.SpeedBase), "mass_bits32": math.Float32bits(unit.Mass),
				"capacity": uint32(uint16(pl.field3652)), "carry": uint32(unit.CarryCapacity), "overweight": pl.Field3656,
				"initialized": uint32(*playerReadValuesInitializedPtr4EEDC0(pl)), "return_bits32": uint32(result),
			}
			if !reflect.DeepEqual(got, row.Result) {
				t.Fatalf("native state differs from unchanged original: got %v want %v", got, row.Result)
			}
			if !reflect.DeepEqual(protected, row.Protection) || !reflect.DeepEqual(tokens, []uint32{0x11111111, 0x22222222, 0x33333333, 0x44444444}) || pl.field3652>>16 != 0xa5a5 || ud.Player != pl || unit.HealthData != hp {
				t.Fatal("protection order, nonoutput capacity bits or native links changed")
			}
		})
	}
}
