package opennox

import (
	"math"
	"testing"
)

func TestE2ETrollCloudLifetime54A270NearestAndBounds(t *testing.T) {
	for _, row := range []struct {
		name     string
		lifetime float32
		fps      uint32
		want     int32
		ok       bool
	}{
		{"nearest-not-truncate", 1.25, 30, 38, true},
		{"nearest-even-down", 2.5, 1, 2, true},
		{"nearest-even-up", 1.5, 1, 2, true},
		{"ordinary", 5, 30, 150, true},
		{"maximum", 60, 30, 1800, true},
		{"too-long", 61, 30, 0, false},
		{"zero", 0, 30, 0, false},
		{"signed-negative-fps", 1.25, 0xffffffff, 0, false},
		{"NaN", float32(math.NaN()), 30, 0, false},
		{"infinite", float32(math.Inf(1)), 30, 0, false},
	} {
		t.Run(row.name, func(t *testing.T) {
			got, ok := e2eTrollCloudLifetime54A270(row.lifetime, row.fps)
			if got != row.want || ok != row.ok {
				t.Fatalf("lifetime=%d/%t, want %d/%t", got, ok, row.want, row.ok)
			}
		})
	}
	if e2ePlay == "" && consoleCommandAt(noxConsole, "e2e-troll-cloud") != nil {
		t.Fatal("E2E-only Troll command leaked into an ordinary game")
	}
}
