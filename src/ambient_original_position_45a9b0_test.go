package opennox

import (
	_ "embed"
	"encoding/json"
	"image"
	"testing"
)

//go:embed server/testdata/ambient_position_original_45a9b0.json
var originalAmbientPosition45A9B0 []byte

func TestNativeAmbientOriginalPosition45A9B0(t *testing.T) {
	var fixture struct {
		ControlWord int `json:"control_word"`
		Cases       []struct {
			Input struct {
				Name                         string `json:"name"`
				Source, Listener             [2]int32
				Radius, Width                int32
				ScreenX                      int32 `json:"screen_x"`
				WorldX                       int32 `json:"world_x"`
				Flags, Flags70               uint32
				Definition, Viewport, Create bool
			} `json:"input"`
			Result struct {
				Events [][]json.RawMessage `json:"events"`
			} `json:"result"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(originalAmbientPosition45A9B0, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.ControlWord != 3647 || len(fixture.Cases) != 169 {
		t.Fatal("unexpected original input/profile set")
	}
	for _, c := range fixture.Cases {
		in := c.Input
		// Early opaque lookup and failed allocation produce no parameter setter
		// observations. The actual C playback fixture covers native admission.
		if !in.Definition || !in.Viewport || !in.Create {
			continue
		}
		t.Run(in.Name, func(t *testing.T) {
			var wantVolume, wantPan int32
			for _, event := range c.Result.Events {
				var kind string
				if err := json.Unmarshal(event[0], &kind); err != nil {
					t.Fatal(err)
				}
				switch kind {
				case "initial-volume", "volume":
					if err := json.Unmarshal(event[1], &wantVolume); err != nil {
						t.Fatal(err)
					}
				case "initial-pan", "pan":
					if err := json.Unmarshal(event[1], &wantPan); err != nil {
						t.Fatal(err)
					}
				}
			}
			volume, pan := ambientParameters45A9B0(image.Pt(int(in.Source[0]), int(in.Source[1])), image.Pt(int(in.Listener[0]), int(in.Listener[1])), in.Flags, in.Flags70, in.ScreenX, in.WorldX, in.Width, in.Radius)
			if volume != wantVolume || volume != 0 && pan != wantPan {
				t.Fatalf("native parameters=(%d,%d), original setter observations=(%d,%d)", volume, pan, wantVolume, wantPan)
			}
		})
	}
}
