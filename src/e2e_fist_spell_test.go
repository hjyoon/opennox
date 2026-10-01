package opennox

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/opennox/libs/types"
)

func TestE2EFistSpellSchedule(t *testing.T) {
	for level, want := range []string{"SmallFist", "MediumFist", "LargeFist", "LargeFist", "LargeFist"} {
		level++
		t.Run(fmt.Sprint(level), func(t *testing.T) {
			if got, ok := e2eFistType(level); !ok || got != want {
				t.Fatalf("type=%q/%t want=%q", got, ok, want)
			}
			var sc e2eScenario
			sc.CheckFistSpell(level, "Fist")
			if len(sc.steps) != 7 {
				t.Fatalf("steps=%d want=7", len(sc.steps))
			}
			for suffix, timeout := range map[string]time.Duration{
				" prepare": 1200, " falling drawable": 120, " natural collision": 180, " natural deletion": 300,
			} {
				found := 0
				for _, step := range sc.steps {
					if strings.HasSuffix(step.name, suffix) {
						found++
						if step.ready == nil || step.fnc == nil || step.waitTimeout != timeout {
							t.Fatalf("%q has no bounded live result check", suffix)
						}
					}
				}
				if found != 1 {
					t.Fatalf("%q found=%d want=1", suffix, found)
				}
			}
		})
	}
}

func TestE2EFistSpellRejectsInvalidLevelBeforeScheduling(t *testing.T) {
	for _, level := range []int{-1, 0, 6, 29, 0x7fffffff} {
		var sc e2eScenario
		func() {
			defer func() {
				if recover() == nil || len(sc.steps) != 0 {
					t.Errorf("level=%d scheduled a fixture", level)
				}
			}()
			sc.CheckFistSpell(level, "invalid")
		}()
	}
}

func TestE2EFistPositionRetainsScriptCoordinates(t *testing.T) {
	p := types.Ptf(-33.5, 12.25)
	target := e2eFistPosition{p}
	p.X = 900
	if got := target.Pos(); got != types.Ptf(-33.5, 12.25) {
		t.Fatalf("position=%v", got)
	}
}
