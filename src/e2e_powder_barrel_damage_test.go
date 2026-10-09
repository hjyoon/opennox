package opennox

import (
	"math"
	"os"
	"strings"
	"testing"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
	"gopkg.in/yaml.v2"
)

func TestE2EPowderBarrelArenaChecksRadialFootprints(t *testing.T) {
	trace := func(from, to types.Pointf) bool {
		return math.Abs(float64(from.X)) <= 180 && math.Abs(float64(to.X)) <= 180 &&
			math.Abs(float64(from.Y)) <= 80 && math.Abs(float64(to.Y)) <= 80
	}
	if _, _, err := e2eWarriorAbilityArena(types.Pointf{}, 96, trace); err == nil {
		t.Fatal("test room unexpectedly fits a 96-wide warrior charge lane")
	}
	center, direction, err := e2ePowderBarrelArena(types.Pointf{}, 12, trace)
	if err != nil || center != (types.Pointf{}) || direction != types.Ptf(1, 0) {
		t.Fatalf("clear radial fixture rejected: center=%v direction=%v err=%v", center, direction, err)
	}
	if _, _, err := e2ePowderBarrelArena(types.Pointf{}, 12, func(types.Pointf, types.Pointf) bool { return false }); err == nil {
		t.Fatal("blocked radial fixture accepted")
	}
}

func TestE2EPowderBarrelTracksRecycledObjectIdentity(t *testing.T) {
	unit := &server.Object{ScriptIDVal: 12}
	f := e2ePowderBarrelFixture{existing: map[*server.Object]int32{unit: 12}}
	if f.newObject(unit) {
		t.Fatal("existing stock object counted as a death spawn")
	}
	unit.ScriptIDVal = 13
	if !f.newObject(unit) || !f.newObject(&server.Object{ScriptIDVal: 14}) {
		t.Fatal("new/recycled death spawn was excluded by its pool address")
	}
}

func TestE2EPowderBarrelDamageContracts(t *testing.T) {
	for _, tc := range []struct {
		distance float32
		damage   int32
	}{
		{0, 30}, {29, 30}, {30, 30}, {64, 15}, {99, 0}, {100, 0}, {128, 0},
	} {
		if got := e2ePowderBarrelRadialDamage(tc.distance); got != tc.damage {
			t.Fatalf("radial damage at %g=%d want=%d", tc.distance, got, tc.damage)
		}
	}
	for _, tc := range []struct {
		raw             int32
		armored, immune bool
		armor, carry    float32
		damage          int32
		remainder       float32
	}{
		{15, false, false, 0, 0, 15, 0}, {15, false, true, 0, 0, 7, 0},
		{15, true, false, .5, .25, 8, -.25}, {15, true, true, .5, .25, 4, -.25},
		{9, true, false, .5, .25, 5, -.25}, {9, true, false, .5, -.25, 4, .25},
		{1, true, false, 1, 0, 1, 0},
	} {
		damage, carry := e2ePowderBarrelUnitDamage(tc.raw, tc.armored, tc.immune, tc.armor, tc.carry)
		if damage != tc.damage || carry != math.Float32bits(tc.remainder) {
			t.Fatalf("unit contract %+v got=%d/%#x", tc, damage, carry)
		}
	}
}

func TestE2EPowderBarrelDamageSchedule(t *testing.T) {
	for _, kind := range []string{"BlackPowderBarrel", "BlackPowderBarrel2"} {
		breaking, ok := e2ePowderBarrelBreakingType(kind)
		if !ok || breaking != kind+"Breaking" {
			t.Fatal("stock barrel breaking type mismatch")
		}
		var sc e2eScenario
		sc.CheckPowderBarrelDamage(kind, "barrel")
		if len(sc.steps) != 5 || sc.steps[0].waitTimeout != 1200 || sc.steps[3].waitTimeout != 300 ||
			sc.steps[0].ready == nil || sc.steps[3].ready == nil || sc.steps[2].fnc == nil ||
			sc.steps[1].time-sc.steps[0].time != 12 || sc.steps[4].time-sc.steps[3].time != 12 {
			t.Fatal("unbounded or missing real barrel prepare/ignite/observe/cleanup")
		}
	}
}

func TestE2EPowderBarrelDamageRejectsNonExplosiveBarrels(t *testing.T) {
	for _, kind := range []string{"", "Barrel", "WaterBarrel", "BlackPowderBarrelBreaking", "BlackPowderBarrel/injected"} {
		if _, ok := e2ePowderBarrelBreakingType(kind); ok {
			t.Fatalf("non-explosive fixture accepted: %q", kind)
		}
		func() {
			var sc e2eScenario
			defer func() {
				if recover() == nil || len(sc.steps) != 0 {
					t.Fatal("invalid barrel was scheduled")
				}
			}()
			sc.CheckPowderBarrelDamage(kind, "invalid")
		}()
	}
}

func TestE2EPowderBarrelDamageScenario(t *testing.T) {
	const path = "../scripts/e2e/host-game-powder-barrel-damage.yaml"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	barrels := make(map[string]bool)
	for _, step := range file.Steps {
		if step.Action != "check-powder-barrel-damage" {
			continue
		}
		if _, ok := e2ePowderBarrelBreakingType(step.Item); !ok || barrels[step.Item] {
			t.Fatalf("invalid/repeated barrel fixture: %s", step.Item)
		}
		barrels[step.Item] = true
	}
	if len(barrels) != 2 {
		t.Fatalf("stock barrel variants=%d want=2", len(barrels))
	}
	var sc e2eScenario
	sc.Load(path)
	hits := 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " natural blast and client HP") {
			hits++
		}
	}
	if hits != 2 {
		t.Fatalf("real explosion checks=%d want=2", hits)
	}
	for _, action := range []string{"damage-player", "damage-monster", "set-monster-health", "cast", "prepare-urchin-ranged-attack"} {
		if strings.Contains(string(raw), "action: "+action+"\n") {
			t.Fatalf("injected victim result: %s", action)
		}
	}
}
