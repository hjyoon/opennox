package opennox

import (
	"os"
	"strings"
	"testing"

	"github.com/opennox/libs/object"
)

func TestE2EQuestGolemPierceSample(t *testing.T) {
	arrow := e2eQuestGolemArrow{typeInd: 529, firstFrame: 102, lastFrame: 105, otherDamage: 3}
	check := func(a e2eQuestGolemArrow, known bool, typ uint16,
		frame, start, previous, hit, damage uint32, before, after uint16,
	) bool {
		return e2eQuestGolemPierceSample(a, known, typ, frame, start, previous, hit, damage, before, after)
	}
	if !check(arrow, true, 529, 107, 100, 104, 106, uint32(object.DamageImpale), 450, 447) {
		t.Fatal("recent stock arrow with a new PIERCE HP loss was rejected")
	}
	for _, tc := range []struct {
		name                                string
		known                               bool
		typ                                 uint16
		frame, start, previous, hit, damage uint32
		before, after                       uint16
	}{
		{"unseen or freed source", false, 529, 107, 100, 104, 106, 3, 450, 447},
		{"wrong stock type", true, 528, 107, 100, 104, 106, 3, 450, 447},
		{"missing type", true, 0, 107, 100, 104, 106, 3, 450, 447},
		{"preparation after missile", true, 529, 107, 103, 104, 106, 3, 450, 447},
		{"repeated hit", true, 529, 107, 100, 106, 106, 3, 450, 447},
		{"future hit", true, 529, 105, 100, 104, 106, 3, 450, 447},
		{"hit before last live sighting", true, 529, 106, 100, 100, 104, 3, 450, 447},
		{"stale cached identity", true, 529, 109, 100, 104, 108, 3, 450, 447},
		{"stale damage attribution", true, 529, 109, 100, 104, 106, 3, 450, 447},
		{"electric damage", true, 529, 107, 100, 104, 106, 17, 450, 447},
		{"unchanged health", true, 529, 107, 100, 104, 106, 3, 450, 450},
		{"healed player", true, 529, 107, 100, 104, 106, 3, 450, 451},
		{"dead player", true, 529, 107, 100, 104, 106, 3, 3, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if check(arrow, tc.known, tc.typ, tc.frame, tc.start, tc.previous, tc.hit, tc.damage, tc.before, tc.after) {
				t.Fatal("nonmatching hit was accepted")
			}
		})
	}
	arrow.otherDamage = 0
	if check(arrow, true, 529, 107, 100, 104, 106, 3, 450, 447) {
		t.Fatal("zero-damage stock data was accepted")
	}
	arrow.otherDamage, arrow.firstFrame = 3, 106
	if check(arrow, true, 529, 107, 100, 104, 106, 3, 450, 447) {
		t.Fatal("reversed live sighting interval was accepted")
	}
}

func TestE2EQuestGolemPierceReport(t *testing.T) {
	for _, tc := range []struct {
		name                      string
		before, reported, arrowHP uint16
		current                   uint16
		delta                     int16
		want                      bool
	}{
		{"arrow only", 194, 189, 189, 189, -5, true},
		{"arrow and another stock attack", 194, 163, 189, 163, -31, true},
		{"stale event before arrow", 194, 194, 189, 189, -122, false},
		{"report does not yet include arrow", 450, 194, 189, 194, -256, false},
		{"wrong delta", 194, 163, 189, 163, -5, false},
		{"new damage not reported", 194, 189, 189, 163, -5, false},
		{"healing", 194, 200, 189, 200, 6, false},
		{"zero HP", 194, 0, 189, 0, -194, false},
		{"missing arrow hit", 194, 163, 0, 163, -31, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := e2eQuestGolemPierceReport(tc.before, tc.reported, tc.arrowHP, tc.current, tc.delta); got != tc.want {
				t.Fatalf("report accepted=%t, want %t", got, tc.want)
			}
		})
	}
}

func TestE2EQuestGolemPierceSchedule(t *testing.T) {
	var sc e2eScenario
	sc.CheckQuestGolemPierce(3, "stock pierce")
	found := 0
	for _, step := range sc.steps {
		if step.name == "stock pierce wait for stock GolemArrow hits and client display" {
			found++
			if step.ready == nil || step.fnc == nil || step.waitTimeout != 1800 {
				t.Fatal("stock projectile observation is not bounded")
			}
		}
	}
	if found != 1 || len(sc.steps) != 4 || sc.steps[0].fnc == nil || sc.steps[3].fnc == nil ||
		sc.steps[1].ready == nil || sc.steps[1].waitTimeout != 600 {
		t.Fatal("missing preparation, live hit observation or screenshot")
	}
}

func TestE2EQuestGolemPierceScenario(t *testing.T) {
	raw, err := os.ReadFile("../scripts/e2e/host-quest-golem-pierce.yaml")
	if err != nil {
		t.Fatal(err)
	}
	scenario := string(raw)
	if strings.Count(scenario, "action: enter-quest-exit") != 4 ||
		!strings.Contains(scenario, "action: check-quest-golem-pierce\n    count: 3") ||
		!strings.Contains(scenario, "generate FlyingGolem") || !strings.Contains(scenario, "map: g_forest") {
		t.Fatal("stock golem requires four real Quest transitions and three actual hits")
	}
	for _, action := range []string{"spawn-monster", "set-monster-health", "set-player-health", "damage-monster", "call-noxscript-function", "switch-map"} {
		if strings.Contains(scenario, "action: "+action) {
			t.Fatalf("stock projectile scenario injects game behavior: %s", action)
		}
	}
}
