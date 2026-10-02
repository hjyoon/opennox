package opennox

import (
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v2"
)

func TestE2EPoisonSpellModes(t *testing.T) {
	for _, direction := range []string{"player-to-npc", "npc-to-player", "player-to-monster", ""} {
		for level := -1; level <= 6; level++ {
			fromNPC, ok := e2ePoisonMode(level, direction)
			want := direction == "player-to-npc" && level >= 1 && level <= 5 || direction == "npc-to-player" && level == 0
			if ok != want || ok && fromNPC != (direction == "npc-to-player") {
				t.Fatalf("Poison mode=%s/%d got=%t/%t want ok=%t", direction, level, fromNPC, ok, want)
			}
		}
	}
}

func TestE2EPoisonSpellFirstTick(t *testing.T) {
	for _, tc := range []struct {
		applied uint32
		power   uint8
		want    uint32
	}{{1, 1, 128}, {67, 1, 128}, {68, 1, 256},
		{1, 2, 64}, {1, 3, 64}, {4, 3, 96}, {64, 3, 128},
		{128, 8, 189}, {128, 9, 189}, {128, 255, 189},
		{math.MaxUint32 - 60, 3, 0}, {math.MaxUint32 - 47, 3, 32}} {
		if got, ok := e2ePoisonFirstTick(tc.applied, tc.power); !ok || got != tc.want {
			t.Fatalf("applied/power=%d/%d first=%d/%t want=%d", tc.applied, tc.power, got, ok, tc.want)
		}
	}
	for _, tc := range []struct {
		applied uint32
		power   uint8
	}{{0, 3}, {100, 0}} {
		if _, ok := e2ePoisonFirstTick(tc.applied, tc.power); ok {
			t.Fatalf("invalid Poison timing accepted: %+v", tc)
		}
	}
}

func TestE2EPoisonSpellSchedule(t *testing.T) {
	for _, direction := range []string{"player-to-npc", "npc-to-player"} {
		level := 3
		if direction == "npc-to-player" {
			level = 0
		}
		var sc e2eScenario
		sc.CheckPoisonSpell(level, direction, "Poison")
		if len(sc.steps) != 5 {
			t.Fatalf("%s steps=%d want=5", direction, len(sc.steps))
		}
		for suffix, timeout := range map[string]time.Duration{" prepare": 1200, " actual DOT and client replay": 300, " cure replay": 120} {
			found := 0
			for _, step := range sc.steps {
				if strings.HasSuffix(step.name, suffix) {
					found++
					if step.ready == nil || step.fnc == nil || step.waitTimeout != timeout {
						t.Fatalf("%q has no bounded check", suffix)
					}
				}
			}
			if found != 1 {
				t.Fatalf("%q found=%d", suffix, found)
			}
		}
	}
}

func TestE2EPoisonSpellInvalidFixture(t *testing.T) {
	for _, tc := range []struct {
		level int
		dir   string
	}{{0, "player-to-npc"}, {6, "player-to-npc"}, {1, "npc-to-player"}, {1, "player-to-monster"}, {1, ""}} {
		var sc e2eScenario
		func() {
			defer func() {
				if recover() == nil || len(sc.steps) != 0 {
					t.Errorf("invalid Poison fixture scheduled: %+v", tc)
				}
			}()
			sc.CheckPoisonSpell(tc.level, tc.dir, "invalid")
		}()
	}
}

func TestE2EPoisonSpellScenario(t *testing.T) {
	path := "../scripts/e2e/host-game-poison-spell.yaml"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	seen := map[string]map[int]bool{}
	for _, step := range file.Steps {
		if step.Action != "check-poison-spell" {
			continue
		}
		if _, ok := e2ePoisonMode(step.Count, step.Text); !ok || seen[step.Text][step.Count] {
			t.Fatalf("invalid or repeated Poison scenario=%s/%d", step.Text, step.Count)
		}
		if seen[step.Text] == nil {
			seen[step.Text] = map[int]bool{}
		}
		seen[step.Text][step.Count] = true
	}
	if len(seen["player-to-npc"]) != 5 || len(seen["npc-to-player"]) != 1 {
		t.Fatalf("Poison casts=%v want five script requests and one natural NPC cast", seen)
	}
	var sc e2eScenario
	sc.Load(path)
	loaded := 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " actual DOT and client replay") {
			loaded++
		}
	}
	if loaded != 6 {
		t.Fatalf("loaded Poison checks=%d want=6", loaded)
	}
	for _, action := range []string{"cast", "damage-monster", "damage-player", "set-monster-health", "set-player-health", "arm-player-poison"} {
		if strings.Contains(string(raw), "action: "+action+"\n") {
			t.Fatalf("Poison scenario has a separate result injection: %s", action)
		}
	}
}
