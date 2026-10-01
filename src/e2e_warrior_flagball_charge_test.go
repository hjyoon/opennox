package opennox

import (
	"strings"
	"testing"
	"time"

	"github.com/opennox/opennox/v1/server"
)

func TestE2EWarriorFlagBallChargeSchedule(t *testing.T) {
	var sc e2eScenario
	sc.CheckWarriorFlagBallCharge("charge carrier")
	for suffix, timeout := range map[string]time.Duration{
		" prepare": 1200, " stock ball pickup": 180, " wait for spawn protection": 1200,
		" server activation": 120, " actual damage and drop": 180,
		" wait for effect end": 1200, " wait for ready": 1200,
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
}

func TestE2EWarriorFlagBallOwnedList(t *testing.T) {
	first, ball, last := new(server.Object), new(server.Object), new(server.Object)
	owner := &server.Object{Field129: first}
	first.Field128, ball.Field128 = ball, last
	if !playerDamageOwnsBallE2E(owner, ball) || playerDamageOwnsBallE2E(owner, new(server.Object)) {
		t.Fatal("owned-ball observation must traverse native successor pointers")
	}
	first.Field128 = last
	if playerDamageOwnsBallE2E(owner, ball) {
		t.Fatal("detached ball is still considered owned")
	}
}
