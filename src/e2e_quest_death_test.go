package opennox

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

func TestE2EQuestDeathReadOnlyNativeDWORD(t *testing.T) {
	player := &server.Player{GoldVal: 0xaabbccdd}
	for _, name := range []string{"field4660", "field4664", "field4668", "field4672", "field4688"} {
		field, ok := reflect.TypeOf(*player).FieldByName(name)
		if !ok || field.Type.Kind() != reflect.Uint32 {
			t.Fatalf("missing native DWORD %q", name)
		}
		// Seed native declared fields only in this unit test, not the GUI fixture.
		*(*uint32)(unsafe.Add(unsafe.Pointer(player), field.Offset)) = 0x89abcdef
		before := *player
		value, err := e2eQuestDeathPlayerDWORD(player, name)
		if err != nil || value != 0x89abcdef || *player != before {
			t.Fatalf("native read %q = %#x/%v changed=%t", name, value, err, *player != before)
		}
	}
	for _, name := range []string{"missing", "PlayerInd", "PlayerUnit"} {
		if _, err := e2eQuestDeathPlayerDWORD(player, name); err == nil {
			t.Fatalf("invalid DWORD %q accepted", name)
		}
	}
	if _, err := e2eQuestDeathPlayerDWORD(nil, "field4660"); err == nil {
		t.Fatal("nil Player accepted")
	}
}

func TestE2EQuestDeathTransitionsRejectPartialSuccess(t *testing.T) {
	for _, lives := range []uint32{2, 1, 0} {
		t.Run(fmt.Sprint(lives), func(t *testing.T) {
			before := e2eQuestDeathState{lives: lives, deaths: 5, gold: 101, health: 450, maximum: 450, stage: 5}
			after := e2eQuestDeathState{lives: 2, gold: 51, maximum: 450, stage: 5, blocker: 1234, state: server.PlayerState4, flags: object.FlagDead, scoreVisible: true}
			if lives != 0 {
				after.lives, after.deaths, after.gold, after.blocker, after.scoreVisible = lives-1, 6, 101, 0, false
			}
			if err := e2eQuestDeathTransition(before, after); err != nil {
				t.Fatal(err)
			}
			for name, mutate := range map[string]func(*e2eQuestDeathState){
				"still alive":          func(s *e2eQuestDeathState) { s.health = 1 },
				"not dead":             func(s *e2eQuestDeathState) { s.flags = 0 },
				"destroyed":            func(s *e2eQuestDeathState) { s.flags |= object.FlagDestroyed },
				"animation unfinished": func(s *e2eQuestDeathState) { s.state = server.PlayerState3 },
				"wrong life count":     func(s *e2eQuestDeathState) { s.lives++ },
				"wrong death count":    func(s *e2eQuestDeathState) { s.deaths++ },
				"wrong timestamp": func(s *e2eQuestDeathState) {
					if lives == 0 {
						s.blocker = 0
					} else {
						s.blocker = 1
					}
				},
				"wrong score visibility": func(s *e2eQuestDeathState) { s.scoreVisible = !s.scoreVisible },
				"wrong gold":             func(s *e2eQuestDeathState) { s.gold = 102 },
			} {
				bad := after
				mutate(&bad)
				if err := e2eQuestDeathTransition(before, bad); err == nil {
					t.Fatalf("false success %q: %+v", name, bad)
				}
			}
			if lives == 0 {
				for _, mutate := range []func(*e2eQuestDeathState){func(s *e2eQuestDeathState) { s.generators = 1 }, func(s *e2eQuestDeathState) { s.monsters = 1 }, func(s *e2eQuestDeathState) { s.secrets = 1 }, func(s *e2eQuestDeathState) { s.stage = 4 }} {
					bad := after
					mutate(&bad)
					if err := e2eQuestDeathTransition(before, bad); err == nil {
						t.Fatal("partial statistics reset accepted")
					}
				}
			}
		})
	}
}

func TestE2EQuestDeathEvidenceRejectsMissingObservations(t *testing.T) {
	for bits := 0; bits < 8; bits++ {
		f := &e2eQuestDeathFixture{incoming: bits&1 != 0, clientIncoming: bits&2 != 0, lethalSeen: bits&4 != 0}
		if got, want := f.hasNaturalDamageEvidence(), bits == 7; got != want {
			t.Fatalf("evidence %#x = %t, want %t", bits, got, want)
		}
		// A stale or different final projectile cannot establish exclusive
		// Necromancer attribution, and is deliberately reported separately.
		f.lethalMinion = true
		if got, want := f.hasNaturalDamageEvidence(), bits == 7; got != want {
			t.Fatalf("lethal attribution hides missing evidence %#x", bits)
		}
	}
}

func TestE2EQuestDeathSchedule(t *testing.T) {
	var sc e2eScenario
	sc.CheckQuestPlayerDeaths("Quest death")
	for cycle := 1; cycle <= 3; cycle++ {
		label := fmt.Sprintf("Quest death cycle %d", cycle)
		for name, timeout := range map[string]time.Duration{label + " wait for natural lethal damage": 15000, label + " wait for real-input respawn": 1200} {
			found := 0
			for _, step := range sc.steps {
				if step.name == name {
					found++
					if step.ready == nil || step.waitTimeout != timeout {
						t.Fatalf("unbounded gate %q", name)
					}
				}
			}
			if found != 1 {
				t.Fatalf("gate %q occurs %d times", name, found)
			}
		}
	}
}

func TestE2EQuestDeathScenarioDoesNotInjectOutcome(t *testing.T) {
	raw, err := os.ReadFile("../scripts/e2e/host-quest-player-death.yaml")
	if err != nil {
		t.Fatal(err)
	}
	scenario := string(raw)
	if strings.Count(scenario, "action: enter-quest-exit") != 4 || strings.Count(scenario, "action: check-quest-player-deaths") != 1 || !strings.Contains(scenario, "creature: Necromancer") || !strings.Contains(scenario, "map: g_forest") {
		t.Fatal("Quest death requires stock menus, four real exits and a generated Necromancer")
	}
	for _, action := range []string{"spawn-monster", "set-monster-health", "damage-monster", "call-noxscript-function", "switch-map", "damage-player", "set-player-health", "set-player-lives"} {
		if strings.Contains(scenario, "action: "+action) {
			t.Fatalf("natural death scenario injects outcome: %s", action)
		}
	}
}
