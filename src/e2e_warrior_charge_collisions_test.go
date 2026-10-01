package opennox

import (
	"strings"
	"testing"
	"time"
)

func TestE2EWarriorChargeCollisionSchedule(t *testing.T) {
	for _, kind := range []string{"player", "wall"} {
		t.Run(kind, func(t *testing.T) {
			var sc e2eScenario
			sc.CheckWarriorChargeCollision(kind, "charge")
			for suffix, timeout := range map[string]time.Duration{
				" prepare": 1200, " server activation": 120,
				" actual collision effect": 180, " wait for effect end": 1200, " wait for ready": 1200,
			} {
				count := 0
				for _, step := range sc.steps {
					if strings.HasSuffix(step.name, suffix) {
						count++
						if step.ready == nil || step.fnc == nil || step.waitTimeout != timeout {
							t.Fatalf("gate %q lacks a bounded predicate and result check: %+v", step.name, step)
						}
					}
				}
				if count != 2 {
					t.Fatalf("gate %q occurs %d times, want two complete cycles", suffix, count)
				}
			}
			for _, suffix := range []string{" charge by actual keyboard input", " retry during cooldown", " verify retry rejected"} {
				count := 0
				for _, step := range sc.steps {
					if strings.HasSuffix(step.name, suffix) && step.fnc != nil {
						count++
					}
				}
				if count != 2 {
					t.Fatalf("input/retry %q occurs %d times, want two", suffix, count)
				}
			}
		})
	}
}

func TestE2EWarriorChargeCollisionRejectsUnknownKind(t *testing.T) {
	for _, kind := range []string{"", "monster", "Wall"} {
		t.Run(kind, func(t *testing.T) {
			var sc e2eScenario
			defer func() {
				if recover() == nil || len(sc.steps) != 0 {
					t.Fatal("invalid charge fixture must fail before scheduling any game mutation")
				}
			}()
			sc.CheckWarriorChargeCollision(kind, "invalid")
		})
	}
}
