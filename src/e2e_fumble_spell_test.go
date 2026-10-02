package opennox

import (
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"gopkg.in/yaml.v2"

	"github.com/opennox/opennox/v1/server"
)

func TestE2EFumbleSpellModes(t *testing.T) {
	for _, direction := range []string{"player-to-npc", "npc-to-player", "player-to-monster", ""} {
		for level := -1; level <= 6; level++ {
			fromNPC, ordinary, ok := e2eFumbleMode(level, direction)
			want := direction == "player-to-npc" && level >= 1 && level <= 5 ||
				direction == "npc-to-player" && level == 0 || direction == "player-to-monster" && level == 3
			if ok != want || ok && (fromNPC != (direction == "npc-to-player") || ordinary != (direction == "player-to-monster")) {
				t.Fatalf("Fumble mode=%s/%d got=%t/%t/%t want ok=%t", direction, level, fromNPC, ordinary, ok, want)
			}
		}
	}
}

func TestE2EFumbleSpellDropPredicate(t *testing.T) {
	for _, tc := range []struct {
		class object.Class
		sub   uint32
		want  bool
	}{{object.ClassWeapon, 0, true}, {object.ClassWand, 0, true},
		{object.ClassArmor, 2, true}, {object.ClassArmor, 3, true},
		{object.ClassArmor, 1, false}, {object.ClassArmor, 0x2000, false},
		{object.ClassFood, 2, false}, {object.ClassFlag, 0, false}, {0, 2, false}} {
		for _, equipped := range []bool{false, true} {
			item := &server.Object{ObjClass: tc.class, ObjSubClass: object.SubClass(tc.sub)}
			if equipped {
				item.ObjFlags = object.FlagEquipped
			}
			if got := e2eFumbleShouldDrop(item, false); got != (tc.want && equipped) {
				t.Fatalf("class/sub/equipped=%s/%#x/%t drop=%t", tc.class, tc.sub, equipped, got)
			}
			if !e2eFumbleShouldDrop(item, true) {
				t.Fatal("ordinary monster must use drop-all, not the humanoid equipment predicate")
			}
		}
	}
}

func TestE2EFumbleSpellForcePrediction(t *testing.T) {
	for _, tc := range []struct {
		origin   types.Pointf
		position types.Pointf
		mass     float32
		want     types.Pointf
	}{{types.Ptf(0, 0), types.Ptf(0, 0), 1, types.Ptf(0, 0)},
		{types.Ptf(0, 0), types.Ptf(3, 0), 50, types.Ptf(30.0/3.1, 0)},
		{types.Ptf(0, 0), types.Ptf(0, -3), 50, types.Ptf(0, -30.0/3.1)},
		{types.Ptf(109, 100), types.Ptf(112, 100), 50, types.Ptf(30.0/3.1, 0)}} {
		got := e2eFumbleExpectedForce(tc.origin, tc.position, tc.mass)
		if math.Abs(float64(got.X-tc.want.X)) > 1e-6 || math.Abs(float64(got.Y-tc.want.Y)) > 1e-6 {
			t.Fatalf("position/mass=%v/%g force=%v want=%v", tc.position, tc.mass, got, tc.want)
		}
	}
}

func TestE2EFumbleSpellSchedule(t *testing.T) {
	for _, direction := range []string{"player-to-npc", "npc-to-player", "player-to-monster"} {
		level := 3
		if direction == "npc-to-player" {
			level = 0
		}
		var sc e2eScenario
		sc.CheckFumbleSpell(level, direction, "Fumble")
		if len(sc.steps) != 5 {
			t.Fatalf("%s steps=%d want=5", direction, len(sc.steps))
		}
		for suffix, timeout := range map[string]time.Duration{" prepare": 1200, " actual drop and client replay": 300} {
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

func TestE2EFumbleSpellInvalidFixture(t *testing.T) {
	for _, tc := range []struct {
		level int
		dir   string
	}{{0, "player-to-npc"}, {6, "player-to-npc"}, {1, "npc-to-player"}, {1, "player-to-monster"}, {1, ""}} {
		var sc e2eScenario
		func() {
			defer func() {
				if recover() == nil || len(sc.steps) != 0 {
					t.Errorf("invalid Fumble fixture scheduled: %+v", tc)
				}
			}()
			sc.CheckFumbleSpell(tc.level, tc.dir, "invalid")
		}()
	}
}

func TestE2EFumbleSpellScenario(t *testing.T) {
	path := "../scripts/e2e/host-game-fumble-spell.yaml"
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
		if step.Action != "check-fumble-spell" {
			continue
		}
		if _, _, ok := e2eFumbleMode(step.Count, step.Text); !ok || seen[step.Text][step.Count] {
			t.Fatalf("invalid or repeated Fumble scenario=%s/%d", step.Text, step.Count)
		}
		if seen[step.Text] == nil {
			seen[step.Text] = map[int]bool{}
		}
		seen[step.Text][step.Count] = true
	}
	if len(seen["player-to-npc"]) != 5 || len(seen["npc-to-player"]) != 1 || len(seen["player-to-monster"]) != 1 {
		t.Fatalf("Fumble casts=%v want five script requests, one NPC cast and one drop-all branch", seen)
	}
	var sc e2eScenario
	sc.Load(path)
	loaded := 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " actual drop and client replay") {
			loaded++
		}
	}
	if loaded != 7 {
		t.Fatalf("loaded Fumble checks=%d want=7", loaded)
	}
	for _, action := range []string{"cast", "damage-monster", "damage-player", "set-monster-health", "set-player-health"} {
		if strings.Contains(string(raw), "action: "+action+"\n") {
			t.Fatalf("Fumble scenario has a separate result injection: %s", action)
		}
	}
}
