package opennox

import (
	"image"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/opennox/libs/noximage"
	"gopkg.in/yaml.v2"
)

func TestE2ESentryGlobeSchedule(t *testing.T) {
	for _, mode := range []string{"world-to-player", "world-to-npc", "world-to-monster", "player-to-npc", "npc-to-player", "npc-to-monster"} {
		t.Run(mode, func(t *testing.T) {
			var sc e2eScenario
			sc.CheckSentryGlobe(mode, "Sentry")
			if len(sc.steps) != 7 {
				t.Fatalf("steps=%d want=7", len(sc.steps))
			}
			for suffix, timeout := range map[string]time.Duration{
				" prepare stock hazard": 1200, " real damage and rendered ray": 12,
				" client damage": 120, " disabled ray cleanup": 120,
			} {
				found := 0
				for _, step := range sc.steps {
					if strings.HasSuffix(step.name, suffix) {
						found++
						if step.ready == nil || step.fnc == nil || step.waitTimeout != timeout {
							t.Fatalf("%q is not a bounded live check", suffix)
						}
					}
				}
				if found != 1 {
					t.Fatalf("%q found=%d", suffix, found)
				}
			}
		})
	}
}

func TestE2ESentryGlobeRejectsInvalidMode(t *testing.T) {
	for _, mode := range []string{"", "player-to-player", "injected"} {
		t.Run(mode, func(t *testing.T) {
			var sc e2eScenario
			defer func() {
				if recover() == nil || len(sc.steps) != 0 {
					t.Fatal("invalid mode was scheduled")
				}
			}()
			sc.CheckSentryGlobe(mode, "invalid")
		})
	}
}

func TestE2ESentryBeamPixels(t *testing.T) {
	for _, tc := range []struct {
		name     string
		from, to image.Point
	}{{"horizontal", image.Pt(80, 100), image.Pt(200, 100)}, {"diagonal", image.Pt(80, 100), image.Pt(180, 200)}, {"reverse", image.Pt(200, 180), image.Pt(80, 140)}} {
		t.Run(tc.name, func(t *testing.T) {
			pix := noximage.NewImage16(image.Rect(0, 0, 300, 300))
			for i := range pix.Pix {
				pix.Pix[i] = 0x4321
			}
			before := append([]uint16(nil), pix.Pix...)
			count, ok := e2eSentryBeamPixels(pix, tc.from, tc.to, 0x4321)
			if !ok || count < 100 || count > 369 || !reflect.DeepEqual(before, pix.Pix) {
				t.Fatalf("count=%d ok=%t or framebuffer mutated", count, ok)
			}
			count, ok = e2eSentryBeamPixels(pix, tc.from, tc.to, 0x1234)
			if !ok || count != 0 {
				t.Fatalf("unrelated color=%d/%t", count, ok)
			}
		})
	}
	for _, tc := range []struct {
		name     string
		pix      *noximage.Image16
		from, to image.Point
		color    uint16
	}{{"nil", nil, image.Point{}, image.Pt(100, 100), 1}, {"zero-color", noximage.NewImage16(image.Rect(0, 0, 300, 300)), image.Pt(80, 100), image.Pt(200, 100), 0}, {"short", noximage.NewImage16(image.Rect(0, 0, 300, 300)), image.Pt(80, 100), image.Pt(100, 100), 1}, {"outside", noximage.NewImage16(image.Rect(0, 0, 300, 300)), image.Pt(-100, -100), image.Pt(-200, -200), 1}} {
		t.Run(tc.name, func(t *testing.T) {
			count, ok := e2eSentryBeamPixels(tc.pix, tc.from, tc.to, tc.color)
			if ok || count != 0 {
				t.Fatalf("invalid pixels=%d/%t", count, ok)
			}
		})
	}
}

func TestE2ESentryGlobeScenario(t *testing.T) {
	const path = "../scripts/e2e/host-game-sentry-globe.yaml"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]bool)
	for _, step := range file.Steps {
		if step.Action != "check-sentry-globe" {
			continue
		}
		if _, _, ok := e2eSentryMode(step.Text); !ok || seen[step.Text] {
			t.Fatalf("invalid/repeated mode %q", step.Text)
		}
		seen[step.Text] = true
	}
	if len(seen) != 6 {
		t.Fatalf("modes=%d want=6", len(seen))
	}
	var sc e2eScenario
	sc.Load(path)
	checks := 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " real damage and rendered ray") {
			checks++
		}
	}
	if checks != 6 {
		t.Fatalf("live checks=%d want=6", checks)
	}
	for _, action := range []string{"damage-monster", "damage-player", "projectile-fx", "cast", "set-monster-health"} {
		if strings.Contains(string(raw), "action: "+action+"\n") {
			t.Fatalf("separate result injection %s", action)
		}
	}
}
