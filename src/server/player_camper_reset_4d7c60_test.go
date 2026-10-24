package server

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/opennox/libs/balance"
	"github.com/opennox/libs/strman"
)

func TestPlayerCamperResetOriginal4D7C60(t *testing.T) {
	raw, err := os.ReadFile("testdata/player_camper_reset_4d7c60.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Unchanged bool   `json:"original_unchanged"`
		StartupCW uint32 `json:"actual_startup_control_word"`
		Rows      []struct {
			Name       string            `json:"name"`
			CW         uint32            `json:"fpcw"`
			Input      map[string]uint32 `json:"input_bits"`
			RadiusBits uint32            `json:"radius_bits"`
			Start      uint32            `json:"start"`
			Fade       uint32            `json:"fade"`
			LastCheck  uint32            `json:"last_check"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil || !fixture.Unchanged || len(fixture.Rows) != 90 || fixture.StartupCW&0xf3f != 0xe3f {
		t.Fatalf("original Reset fixture: %v", err)
	}
	srv := New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	selected := 0
	for _, row := range fixture.Rows {
		if row.CW != fixture.StartupCW {
			continue
		}
		selected++
		t.Run(fmt.Sprintf("%s", row.Name), func(t *testing.T) {
			value := func(key string) balance.Value {
				return balance.Float(float64(math.Float32frombits(row.Input[key])))
			}
			srv.Balance.file = &balance.File{Global: balance.Config{
				"camperradiussq": value("CamperRadiusSq"), "camperstarttime": value("CamperStartTime"), "camperfadetime": value("CamperFadeTime"),
			}}
			camper := &srv.Players.Camper
			camper.LastCheck, camper.RadiusSq, camper.StartTime, camper.FadeTime = 0xdeadbeef, 123.5, 0x89abcdef, 0xfedcba98
			camper.Reset() // Actual public Reset and server balance lookup, no replacement conversion.
			wantRadius := float64(math.Float32frombits(row.RadiusBits))
			radiusOK := math.IsNaN(wantRadius) && math.IsNaN(camper.RadiusSq) || math.Float64bits(camper.RadiusSq) == math.Float64bits(wantRadius)
			if camper.LastCheck != row.LastCheck || !radiusOK || camper.StartTime != row.Start || camper.FadeTime != row.Fade {
				t.Fatalf("native Reset last=%08x radius=%016x start=%08x fade=%08x; original last=%08x radius=%016x start=%08x fade=%08x", camper.LastCheck, math.Float64bits(camper.RadiusSq), camper.StartTime, camper.FadeTime, row.LastCheck, math.Float64bits(wantRadius), row.Start, row.Fade)
			}
		})
	}
	if selected != 30 {
		t.Fatalf("selected %d original startup rows", selected)
	}
}
