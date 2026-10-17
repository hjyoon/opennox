package opennox

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestE2ECheckpointRebasesFollowingPhase(t *testing.T) {
	for _, tc := range []struct {
		name      string
		prefix    string
		phaseTime time.Duration
	}{
		{name: "populated", prefix: "  - action: wait\n    dt: 11\n", phaseTime: 11},
		{name: "empty"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "phases.yaml")
			data := "steps:\n" + tc.prefix + "  - action: checkpoint\n  - action: wait\n    dt: 7\n    name: following phase\n"
			if err := os.WriteFile(path, []byte(data), 0600); err != nil {
				t.Fatal(err)
			}
			previousJobs := e2eJobs
			previousPlatform, previousSteps, previousDone := e2e.p, e2e.steps, e2e.done
			defer func() {
				e2eJobs = previousJobs
				e2e.p, e2e.steps, e2e.done = previousPlatform, previousSteps, previousDone
			}()
			e2eJobs = make(chan *e2eScenario)
			clock := &platformE2E{}
			e2e.p, e2e.steps, e2e.done = clock, nil, nil
			var sc e2eScenario
			sc.wizard1LightningStart = 12345
			type phaseSnapshot struct {
				sameScenario bool
				steps        []e2eStep
				fixtureState uint32
			}
			phase := make(chan phaseSnapshot, 1)
			go func() {
				job := <-e2eJobs
				snapshot := phaseSnapshot{job == &sc, append([]e2eStep(nil), job.steps...), job.wizard1LightningStart}
				// Model the synchronous live-peer operation advancing virtual
				// time before the completed phase releases its loader.
				clock.tick(30 * time.Millisecond)
				close(job.done)
				phase <- snapshot
			}()
			sc.Load(path)
			first := <-phase
			if !first.sameScenario || first.fixtureState != 12345 || sc.wizard1LightningStart != 12345 {
				t.Fatal("checkpoint replaced the scenario or its fixture state")
			}
			if tc.phaseTime == 0 {
				if len(first.steps) != 0 {
					t.Fatalf("empty phase steps=%d", len(first.steps))
				}
			} else if len(first.steps) != 1 || first.steps[0].time != tc.phaseTime {
				t.Fatalf("first phase=%v", first.steps)
			}
			if len(sc.steps) != 1 || sc.steps[0].time != 7 {
				t.Fatalf("following phase lost its relative delay: %v", sc.steps)
			}
			e2eQueue(&sc)
			if len(e2e.steps) != 1 || e2e.steps[0].time != 30*time.Millisecond+7 {
				t.Fatalf("following phase was not rebased after the clock jump: %v", e2e.steps)
			}
		})
	}
}
