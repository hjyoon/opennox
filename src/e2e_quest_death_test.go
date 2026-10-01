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
	"github.com/opennox/opennox/v1/client/gui"
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
		for name, timeout := range map[string]time.Duration{label + " wait for natural lethal damage": 15000, label + " wait for real-input respawn": 1200, label + " wait for ordinary life HUD packet": 1200} {
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
	for i, step := range sc.steps {
		if strings.HasSuffix(step.name, " wait for ordinary life HUD packet") {
			label := strings.TrimSuffix(step.name, " wait for ordinary life HUD packet")
			if i == 0 || i+1 >= len(sc.steps) || sc.steps[i-1].name != label+" wait for natural lethal damage" || sc.steps[i+1].name != label+" dead" {
				t.Fatal("read-only life packet gate must follow natural death and precede its capture")
			}
		}
	}
	label := "Quest death cycle 3"
	for i, step := range sc.steps {
		if step.name == label+" host result timer" {
			if i < 2 || sc.steps[i-1].name != label+" locate ordinary respawn input" || sc.steps[i-2].name != label+" finish dead-state input delay" {
				t.Fatal("host result capture must follow the ordinary dead-frame delay and read-only timer observation")
			}
			return
		}
	}
	t.Fatal("missing host result timer capture before Continue input")
}

func TestE2EQuestHostResultTimerReadOnly(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(**gui.Window, **server.Player, *gui.Window, *gui.StaticTextData)
		wantOK bool
	}{
		{"empty", func(_ **gui.Window, _ **server.Player, _ *gui.Window, _ *gui.StaticTextData) {}, true},
		{"nil text", func(_ **gui.Window, _ **server.Player, _ *gui.Window, data *gui.StaticTextData) { data.Text = nil }, true},
		{"countdown", func(_ **gui.Window, _ **server.Player, _ *gui.Window, data *gui.StaticTextData) {
			text := []uint16{'T', 'i', 'm', 'e', ' ', '-', ' ', '3', '0', 0}
			data.Text = &text[0]
		}, false},
		{"nil root", func(root **gui.Window, _ **server.Player, _ *gui.Window, _ *gui.StaticTextData) { *root = nil }, false},
		{"wrong root", func(root **gui.Window, _ **server.Player, _ *gui.Window, _ *gui.StaticTextData) { (*root).SetID(10600) }, false},
		{"hidden root", func(root **gui.Window, _ **server.Player, _ *gui.Window, _ *gui.StaticTextData) {
			(*root).Flags |= gui.StatusHidden
		}, false},
		{"nil player", func(_ **gui.Window, player **server.Player, _ *gui.Window, _ *gui.StaticTextData) { *player = nil }, false},
		{"non-host", func(_ **gui.Window, player **server.Player, _ *gui.Window, _ *gui.StaticTextData) {
			(*player).PlayerInd = 0
		}, false},
		{"missing child", func(root **gui.Window, _ **server.Player, _ *gui.Window, _ *gui.StaticTextData) {
			(*root).Field100Ptr = nil
		}, false},
		{"hidden child", func(_ **gui.Window, _ **server.Player, child *gui.Window, _ *gui.StaticTextData) {
			child.Flags |= gui.StatusHidden
		}, false},
		{"wrong widget type", func(_ **gui.Window, _ **server.Player, child *gui.Window, _ *gui.StaticTextData) {
			child.DrawData().Style = gui.StylePushButton
		}, false},
		{"missing widget data", func(_ **gui.Window, _ **server.Player, child *gui.Window, _ *gui.StaticTextData) {
			child.WidgetData = nil
		}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			empty := uint16(0)
			data := &gui.StaticTextData{Text: &empty, Center: 1, Glow: 2}
			child := &gui.Window{Flags: gui.StatusEnabled, WidgetData: unsafe.Pointer(data)}
			child.SetID(10712)
			child.DrawData().Style = gui.StyleStaticText
			root := &gui.Window{Flags: gui.StatusEnabled, Field100Ptr: child}
			root.SetID(10700)
			player := &server.Player{PlayerInd: 31, GoldVal: 0xaabbccdd}
			observedRoot, observedPlayer := root, player
			tc.mutate(&observedRoot, &observedPlayer, child, data)
			beforeRoot, beforeChild, beforePlayer, beforeData := *root, *child, *player, *data
			timer, err := e2eQuestHostResultTimer(observedRoot, observedPlayer)
			if (err == nil) != tc.wantOK || tc.wantOK && timer != child || !tc.wantOK && timer != nil {
				t.Fatalf("timer=%p error=%v want-success=%t", timer, err, tc.wantOK)
			}
			if *root != beforeRoot || *child != beforeChild || *player != beforePlayer || *data != beforeData || empty != 0 {
				t.Fatal("timer observation mutated native window/Player/text data")
			}
		})
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

func TestE2EQuestLifeHUDRejectsPartialSuccessAndIsReadOnly(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(**server.Object, **server.Player, *server.Object, *server.PlayerUpdateData, *uint32, *uint32)
		wantOK bool
	}{
		{"synchronized", func(**server.Object, **server.Player, *server.Object, *server.PlayerUpdateData, *uint32, *uint32) {}, true},
		{"nil unit", func(unit **server.Object, _ **server.Player, _ *server.Object, _ *server.PlayerUpdateData, _ *uint32, _ *uint32) {
			*unit = nil
		}, false},
		{"wrong class", func(_ **server.Object, _ **server.Player, unit *server.Object, _ *server.PlayerUpdateData, _ *uint32, _ *uint32) {
			unit.ObjClass = object.ClassMonster
		}, false},
		{"nil update", func(_ **server.Object, _ **server.Player, unit *server.Object, _ *server.PlayerUpdateData, _ *uint32, _ *uint32) {
			unit.UpdateData = nil
		}, false},
		{"nil server player", func(_ **server.Object, _ **server.Player, _ *server.Object, update *server.PlayerUpdateData, _ *uint32, _ *uint32) {
			update.Player = nil
		}, false},
		{"nil client player", func(_ **server.Object, player **server.Player, _ *server.Object, _ *server.PlayerUpdateData, _ *uint32, _ *uint32) {
			*player = nil
		}, false},
		{"different player", func(_ **server.Object, player **server.Player, _ *server.Object, _ *server.PlayerUpdateData, _ *uint32, _ *uint32) {
			*player = &server.Player{}
		}, false},
		{"wrong code", func(_ **server.Object, _ **server.Player, _ *server.Object, _ *server.PlayerUpdateData, code *uint32, _ *uint32) {
			*code = 0x1235
		}, false},
		{"high client code", func(_ **server.Object, _ **server.Player, _ *server.Object, _ *server.PlayerUpdateData, code *uint32, _ *uint32) {
			*code = 0x10001234
		}, false},
		{"stale zero HUD", func(_ **server.Object, _ **server.Player, _ *server.Object, _ *server.PlayerUpdateData, _ *uint32, hud *uint32) {
			*hud = 0
		}, false},
		{"high HUD bits", func(_ **server.Object, _ **server.Player, _ *server.Object, _ *server.PlayerUpdateData, _ *uint32, hud *uint32) {
			*hud = 0x102
		}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			player := &server.Player{PlayerInd: 31, GoldVal: 0x89abcdef}
			update := &server.PlayerUpdateData{Player: player, ExtraLives: 2}
			unit := &server.Object{ObjClass: object.ClassPlayer, NetCode: 0x89ab1234, UpdateData: unsafe.Pointer(update)}
			observedUnit, clientPlayer := unit, player
			code, hud := uint32(0x1234), uint32(2)
			tc.mutate(&observedUnit, &clientPlayer, unit, update, &code, &hud)
			beforeUnit, beforeUpdate, beforePlayer, beforeCode, beforeHUD := *unit, *update, *player, code, hud
			if err := e2eQuestLifeHUD(observedUnit, clientPlayer, code, hud, 2); (err == nil) != tc.wantOK {
				t.Fatalf("error=%v want-success=%t", err, tc.wantOK)
			}
			if *unit != beforeUnit || *update != beforeUpdate || *player != beforePlayer || code != beforeCode || hud != beforeHUD {
				t.Fatal("HUD observation mutated native state")
			}
		})
	}
	player := &server.Player{PlayerInd: 31}
	update := &server.PlayerUpdateData{Player: player}
	unit := &server.Object{ObjClass: object.ClassPlayer, NetCode: 0x1234, UpdateData: unsafe.Pointer(update)}
	for value := uint32(0); value < 256; value++ {
		update.ExtraLives = value
		if err := e2eQuestLifeHUD(unit, player, 0x1234, value, value); err != nil {
			t.Fatal(err)
		}
	}
	// Server reset prepares two lives during the zero-life result window.
	// Its original marker suppresses the new packet until real respawn.
	update.ExtraLives = 2
	if err := e2eQuestLifeHUD(unit, player, 0x1234, 0, 0); err != nil {
		t.Fatal(err)
	}
}
