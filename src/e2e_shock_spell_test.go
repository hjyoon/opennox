package opennox

import (
	"image"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	noxcolor "github.com/opennox/libs/color"
	"github.com/opennox/libs/noximage"
	"gopkg.in/yaml.v2"
)

func TestE2EShockSpellModes(t *testing.T) {
	for _, mode := range []string{"player", "npc", "npc-natural", "glyph-to-npc", "glyph-to-player", "", "monster"} {
		for level := -1; level <= 6; level++ {
			want := (mode == "player" || mode == "npc") && level >= 1 && level <= 5 ||
				(mode == "npc-natural" || mode == "glyph-to-npc" || mode == "glyph-to-player") && level == 0
			if got := e2eShockMode(level, mode); got != want {
				t.Fatalf("mode=%s/%d accepted=%t want=%t", mode, level, got, want)
			}
		}
	}
}

func TestE2EShockDuration(t *testing.T) {
	for _, tc := range []struct {
		value float64
		want  uint32
		ok    bool
	}{{24, 0, false}, {25, 25, true}, {25.5, 26, true}, {26.5, 26, true},
		{26.500000001, 26, true}, {2900, 2900, true}, {2901, 0, false},
		{65536 + 300, 0, false}, {math.NaN(), 0, false}, {math.Inf(1), 0, false}} {
		if got, ok := e2eShockDuration(tc.value); got != tc.want || ok != tc.ok {
			t.Fatalf("balance=%g duration=%d/%t want=%d/%t", tc.value, got, ok, tc.want, tc.ok)
		}
	}
}

func TestE2EShockWhitePixel(t *testing.T) {
	pix := noximage.NewImage16(image.Rect(0, 0, 100, 100))
	point := image.Pt(50, 50)
	before := append([]uint16(nil), pix.Pix...)
	if e2eShockWhitePixel(pix, point, pix.Rect) {
		t.Fatal("missing particle pixel accepted")
	}
	pix.Pix[pix.PixOffset(point.X, point.Y)] = uint16(noxcolor.RGB5551Color(255, 255, 255))
	if !e2eShockWhitePixel(pix, point, pix.Rect) {
		t.Fatal("live white center rejected")
	}
	for _, view := range []image.Rectangle{image.Rect(55, 0, 100, 100), image.Rect(0, 0, 60, 100), image.Rect(0, 0, 100, 60)} {
		if e2eShockWhitePixel(pix, point, view) {
			t.Fatalf("offscreen/clipped center accepted: %v", view)
		}
	}
	if e2eShockWhitePixel(pix, image.Pt(-1, 50), image.Rect(-100, -100, 200, 200)) {
		t.Fatal("out-of-buffer pixel accepted")
	}
	for i, original := range before {
		if i != pix.PixOffset(point.X, point.Y) && pix.Pix[i] != original {
			t.Fatal("read-only checker wrote pixels")
		}
	}
}

func TestE2EShockSpellSchedule(t *testing.T) {
	for _, tc := range []struct {
		mode         string
		level, steps int
	}{{"player", 3, 8}, {"npc", 3, 8}, {"npc-natural", 0, 8}, {"glyph-to-npc", 0, 4}, {"glyph-to-player", 0, 4}} {
		var sc e2eScenario
		sc.CheckShockSpell(tc.level, tc.mode, "Shock")
		if len(sc.steps) != tc.steps {
			t.Fatalf("%s steps=%d want=%d", tc.mode, len(sc.steps), tc.steps)
		}
		for _, step := range sc.steps {
			if strings.HasSuffix(step.name, " prepare") || strings.HasSuffix(step.name, " complete") || strings.HasSuffix(step.name, " visible") {
				if step.ready == nil || step.fnc == nil || step.waitTimeout <= 0 || step.waitTimeout > time.Duration(3000) {
					t.Fatalf("Shock %s has no bounded observer", step.name)
				}
			}
		}
	}
}

func TestE2EShockSpellInvalidFixture(t *testing.T) {
	for _, tc := range []struct {
		level int
		mode  string
	}{{0, "player"}, {6, "npc"}, {1, "npc-natural"}, {3, "glyph-to-npc"}, {0, "unknown"}} {
		var sc e2eScenario
		func() {
			defer func() {
				if recover() == nil || len(sc.steps) != 0 {
					t.Errorf("invalid Shock fixture scheduled: %+v", tc)
				}
			}()
			sc.CheckShockSpell(tc.level, tc.mode, "invalid")
		}()
	}
}

func TestE2EShockSpellScenario(t *testing.T) {
	path := "../scripts/e2e/host-game-shock-spell.yaml"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	seen := map[string]map[int]bool{}
	wizard := false
	for _, step := range file.Steps {
		if step.Action == "click" && step.Name == "select wizard" && step.X == 768 && step.Y == 208 {
			wizard = true
		}
		if step.Action != "check-shock-spell" {
			continue
		}
		if !e2eShockMode(step.Count, step.Text) || seen[step.Text][step.Count] {
			t.Fatalf("invalid/repeated Shock scenario=%s/%d", step.Text, step.Count)
		}
		if seen[step.Text] == nil {
			seen[step.Text] = map[int]bool{}
		}
		seen[step.Text][step.Count] = true
	}
	if !wizard || len(seen["player"]) != 5 || len(seen["npc"]) != 5 || len(seen["npc-natural"]) != 1 || len(seen["glyph-to-npc"]) != 1 || len(seen["glyph-to-player"]) != 1 {
		t.Fatalf("Shock coverage=%v wizard=%t", seen, wizard)
	}
	var sc e2eScenario
	sc.Load(path)
	loaded := 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " complete") {
			loaded++
		}
	}
	if loaded != 13 {
		t.Fatalf("loaded Shock checks=%d want=13", loaded)
	}
	for _, action := range []string{"cast", "damage-monster", "damage-player", "set-monster-health", "set-player-health", "set-player-buff"} {
		if strings.Contains(string(raw), "action: "+action+"\n") {
			t.Fatalf("Shock scenario injects a result: %s", action)
		}
	}
}
