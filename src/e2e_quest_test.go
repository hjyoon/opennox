package opennox

import (
	"math"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func e2eQuestPlayerFixture() e2eQuestPlayerState {
	return e2eQuestPlayerState{
		mapName: "g_forest", stage: 5, quest: true, connected: true,
		serverCode: 2, clientCode: 2, drawCode: 2, phase: 3, status: 0x10,
		health: 450, maxHealth: 450, position: types.Pointf{X: 2000, Y: 700},
	}
}

func TestE2EQuestPlayerState(t *testing.T) {
	s := e2eQuestPlayerFixture()
	if err := s.validate(5, "G_Forest.map"); err != nil {
		t.Fatal(err)
	}
	if err := s.validate(5, ""); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		mutate func(*e2eQuestPlayerState)
	}{
		{"mode", func(s *e2eQuestPlayerState) { s.quest = false }},
		{"disconnected", func(s *e2eQuestPlayerState) { s.connected = false }},
		{"loading", func(s *e2eQuestPlayerState) { s.loading = true }},
		{"captured briefing", func(s *e2eQuestPlayerState) { s.briefing = true }},
		{"wrong stage", func(s *e2eQuestPlayerState) { s.stage++ }},
		{"empty map", func(s *e2eQuestPlayerState) { s.mapName = "" }},
		{"wrong map", func(s *e2eQuestPlayerState) { s.mapName = "g_cryptd" }},
		{"zero identity", func(s *e2eQuestPlayerState) { s.serverCode = 0 }},
		{"client identity", func(s *e2eQuestPlayerState) { s.clientCode++ }},
		{"drawable identity", func(s *e2eQuestPlayerState) { s.drawCode++ }},
		{"phase", func(s *e2eQuestPlayerState) { s.phase = 0 }},
		{"observer", func(s *e2eQuestPlayerState) { s.status |= 1 }},
		{"dead", func(s *e2eQuestPlayerState) { s.flags |= object.FlagDead }},
		{"destroyed", func(s *e2eQuestPlayerState) { s.flags |= object.FlagDestroyed }},
		{"zero health", func(s *e2eQuestPlayerState) { s.health = 0 }},
		{"zero maximum", func(s *e2eQuestPlayerState) { s.maxHealth = 0 }},
		{"excess health", func(s *e2eQuestPlayerState) { s.health++ }},
		{"NaN position", func(s *e2eQuestPlayerState) { s.position.X = float32(math.NaN()) }},
		{"infinite position", func(s *e2eQuestPlayerState) { s.position.Y = float32(math.Inf(1)) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := e2eQuestPlayerFixture()
			tc.mutate(&s)
			if err := s.validate(5, "g_forest"); err == nil {
				t.Fatal("unplayable Quest state accepted")
			}
		})
	}
	if err := s.validate(0, ""); err == nil {
		t.Fatal("invalid expected stage accepted")
	}
}

func TestE2EQuestMovementDistance(t *testing.T) {
	before := e2eQuestPlayerFixture()
	after := before
	after.position.X += 8
	if distance, err := e2eQuestMovementDistance(before, after); err != nil || distance != 8 {
		t.Fatalf("8-unit movement distance=%f err=%v", distance, err)
	}
	after.position.X = before.position.X + 7.99
	if _, err := e2eQuestMovementDistance(before, after); err == nil {
		t.Fatal("stationary/short movement accepted")
	}
	after = before
	after.position.X += 100
	after.serverCode, after.clientCode, after.drawCode = 3, 3, 3
	if _, err := e2eQuestMovementDistance(before, after); err == nil {
		t.Fatal("replacement player accepted as movement")
	}
	after = before
	after.position.Y += 100
	after.stage++
	if _, err := e2eQuestMovementDistance(before, after); err == nil {
		t.Fatal("map transition accepted as movement")
	}
	after = before
	after.position.Y += 100
	after.briefing = true
	if _, err := e2eQuestMovementDistance(before, after); err == nil {
		t.Fatal("movement with a captured briefing accepted")
	}
}

func TestE2EQuestGeneratorMaximum(t *testing.T) {
	for _, tc := range []struct {
		name     string
		stage    uint32
		selector uint8
		base     float64
		want     uint8
	}{
		{"high before", 19, 0, 8, 8},
		{"high at boundary", 20, 0, 8, 16},
		{"normal at boundary", 20, 1, 5, 10},
		{"low after", 21, 2, 2, 4},
		{"singular before", 19, 3, 1, 1},
		{"singular at boundary", 20, 3, 1, 1},
		{"singular after", ^uint32(0), 3, 1, 1},
		{"truncation before doubling", 20, 0, 8.75, 16},
		{"native byte wrap", 20, 0, 255, 254},
		{"zero maximum", 20, 0, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := e2eQuestGeneratorMaximum(tc.stage, 20, tc.selector, tc.base)
			if err != nil || got != tc.want {
				t.Fatalf("maximum=%d error=%v want=%d", got, err, tc.want)
			}
		})
	}
	for _, base := range []float64{-1, 256, math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, err := e2eQuestGeneratorMaximum(20, 20, 0, base); err == nil {
			t.Fatalf("invalid stock balance accepted: %g", base)
		}
	}
	if _, err := e2eQuestGeneratorMaximum(0, 20, 0, 6); err == nil {
		t.Fatal("uninitialized stage accepted")
	}
	if _, err := e2eQuestGeneratorMaximum(20, 0, 0, 6); err == nil {
		t.Fatal("uninitialized threshold accepted")
	}
	if _, err := e2eQuestGeneratorMaximum(20, 20, 4, 6); err == nil {
		t.Fatal("invalid stock selector accepted")
	}
}
