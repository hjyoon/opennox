package legacy

import (
	"encoding/json"
	"math"
	"os"
	"testing"
	"unsafe"
)

func TestManaBombPayloadOriginal499520(t *testing.T) {
	raw, err := os.ReadFile("../testdata/mana_bomb_original_499520.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA       string `json:"original_sha256"`
		Unchanged bool   `json:"original_unchanged"`
		Rows      []struct {
			Input struct {
				Name    string `json:"name"`
				Kind    string `json:"kind"`
				Path    [4]int `json:"path"`
				Angle   uint16 `json:"angle"`
				Reverse byte   `json:"reverse"`
				Fill    byte   `json:"fill"`
				Mode    byte   `json:"mode"`
			} `json:"input"`
			Activations []struct {
				State struct {
					Effect   [5]uint32 `json:"effect"`
					Field127 uint32    `json:"field127"`
					Update   uint32    `json:"update"`
				} `json:"state"`
			} `json:"activations"`
		} `json:"rows"`
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.Unchanged || len(fixture.Rows) != 754 {
		t.Fatal("unchanged original ManaBomb fixture")
	}
	count := 0
	for _, r := range fixture.Rows {
		if (r.Input.Kind != "spawn" && r.Input.Kind != "payload") || len(r.Activations) == 0 {
			continue
		}
		count++
		t.Run(r.Input.Name, func(t *testing.T) {
			var path [4]uint16
			for i, v := range r.Input.Path {
				path[i] = uint16(v)
			}
			e, marker, address, canary, callback := manaBombPayloadProbe499520(path, r.Input.Angle, r.Input.Reverse, 5, r.Input.Mode, r.Input.Fill)
			want := r.Activations[0].State
			if unsafe.Sizeof(uintptr(0)) == 8 && address <= math.MaxUint32 {
				t.Fatal("native C drawable address must exceed4GiB")
			}
			if e != want.Effect || marker != want.Field127 || !canary || !callback || want.Update != 0x4ca720 {
				t.Fatalf("payload=%x/%x marker=%x/%x canary=%v callback=%v", e, want.Effect, marker, want.Field127, canary, callback)
			}
		})
	}
	if count != 186 {
		t.Fatal("original C payload vector count", count)
	}
}
