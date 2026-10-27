package server

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sync"
	"testing"

	"github.com/opennox/opennox/v1/internal/oracletest"
)

var npcCasterImpactStartup4E17B0 = sync.OnceValues(func() ([]struct {
	Input struct {
		Armor uint32 `json:"absorption_bits32"`
		Raw   int32  `json:"raw"`
		Quest bool   `json:"quest"`
	} `json:"input"`
	Carry  uint32 `json:"carry_bits32"`
	Events []struct {
		Service string `json:"service"`
		Damage  uint32 `json:"damage_bits32"`
	} `json:"events"`
}, error) {
	data, err := os.ReadFile("testdata/npc_caster_impact_startup_original.json")
	if err != nil {
		return nil, err
	}
	var fixture struct {
		OriginalSHA string `json:"original_sha256"`
		ControlWord uint16 `json:"original_startup_control_word"`
		Rows        []struct {
			Input struct {
				Armor uint32 `json:"absorption_bits32"`
				Raw   int32  `json:"raw"`
				Quest bool   `json:"quest"`
			} `json:"input"`
			Carry  uint32 `json:"carry_bits32"`
			Events []struct {
				Service string `json:"service"`
				Damage  uint32 `json:"damage_bits32"`
			} `json:"events"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		return nil, err
	}
	if fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.ControlWord != 0xe3f || len(fixture.Rows) != 36 {
		return nil, fmt.Errorf("incomplete original NPC caster startup observations")
	}
	return fixture.Rows, nil
})

func originalNPCCasterImpactStartup4E17B0(t *testing.T, absorption float32, raw int32, quest bool) (uint32, int32) {
	t.Helper()
	rows, err := npcCasterImpactStartup4E17B0()
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if row.Input.Armor == math.Float32bits(absorption) && row.Input.Raw == raw && row.Input.Quest == quest {
			for _, event := range row.Events {
				if event.Service == "default" {
					return row.Carry, int32(event.Damage)
				}
			}
		}
	}
	t.Fatalf("missing independent NPC caster input %g %d %t", absorption, raw, quest)
	return 0, 0
}

func readOriginalNonItemStartupFixture4E17B0(path string) ([]byte, error) {
	return oracletest.ReadNonItemStartup(path, "testdata/nonitem_damage_historical_startup_original.json")
}

// The original nearest-control seals remain named at each existing call site.
// Re-execution reproduces all three before supplying startup-mode observations.
func originalNonItemStartupPrefixSeal4E17B0(t *testing.T, shape int, nearestControl string) (string, int) {
	t.Helper()
	data, err := os.ReadFile("testdata/nonitem_prefix_startup_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA string `json:"original_sha256"`
		Startup     uint16 `json:"startup_control_word"`
		Rows        []struct {
			Shape int    `json:"shape"`
			Mode  string `json:"mode"`
			Rows  int    `json:"rows"`
			Bytes int    `json:"bytes"`
			SHA   string `json:"sha256"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.Startup != 0xe3f || len(fixture.Rows) != 6 {
		t.Fatal("incomplete unchanged original prefix observations")
	}
	var control, startup string
	var size int
	for _, row := range fixture.Rows {
		if row.Shape == shape && row.Rows == 1164 {
			if row.Mode == "nearest-control" {
				control = row.SHA
			} else if row.Mode == "actual-startup" {
				startup, size = row.SHA, row.Bytes
			}
		}
	}
	if control != nearestControl || startup == "" || size == 0 {
		t.Fatal("original historical control or startup seal did not match")
	}
	return startup, size
}
