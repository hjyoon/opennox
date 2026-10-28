package server

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
)

func TestMapIndexCoordinateOriginal5175E0(t *testing.T) {
	blob, err := os.ReadFile("testdata/map_index_coordinate_5175e0_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		OriginalSHA string `json:"original_sha256"`
		Contexts    int    `json:"contexts"`
		Rows        []struct {
			Mode      string `json:"mode"`
			Input     uint32 `json:"input"`
			Result    int32  `json:"result"`
			Preserved bool   `json:"nonoutput_bytes_preserved"`
		} `json:"rows"`
	}
	if err = json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if f.Contexts != 451 || len(f.Rows) != 902 || f.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" {
		t.Fatal("original coordinate provenance")
	}
	seen := 0
	for _, row := range f.Rows {
		if row.Mode != "actual-startup" {
			continue
		}
		seen++
		t.Run(fmt.Sprintf("input-%08x", row.Input), func(t *testing.T) {
			if !row.Preserved {
				t.Fatal("original output canary")
			}
			if got := RoundCoord(math.Float32frombits(row.Input)); got != int(row.Result) {
				t.Fatalf("native coordinate=%d original=%d input=%08x", got, row.Result, row.Input)
			}
		})
	}
	if seen != 451 {
		t.Fatal("missing original coordinate conditions")
	}
}
