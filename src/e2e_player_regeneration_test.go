package opennox

import (
	"strings"
	"testing"
)

func TestE2EPlayerRegenerationSchedule(t *testing.T) {
	var sc e2eScenario
	sc.CheckPlayerRegeneration("regeneration")
	if len(sc.steps) != 7 {
		t.Fatalf("steps=%d, want 7", len(sc.steps))
	}
	for _, tc := range []struct {
		index   int
		timeout uint32
	}{{0, 1200}, {3, 15000}} {
		step := sc.steps[tc.index]
		if step.ready == nil || step.fnc == nil || uint32(step.waitTimeout) != tc.timeout {
			t.Fatalf("step %d lacks a bounded live assertion", tc.index)
		}
	}
	if !strings.HasSuffix(sc.steps[3].name, " natural equipped-item healing") || !strings.HasSuffix(sc.steps[6].name, " restore fixture HP") {
		t.Fatal("natural healing or post-observation cleanup is missing")
	}
}

func TestE2EPlayerPassiveHealingBound(t *testing.T) {
	for _, tc := range []struct{ elapsed, fps, interval, want uint32 }{
		{0, 30, 20, 0}, {29, 30, 20, 0}, {30, 30, 20, 0},
		{31, 30, 20, 2}, {50, 30, 20, 2}, {51, 30, 20, 3},
		{90, 30, 20, 4}, {150, 30, 20, 7}, {4350, 30, 60, 73},
		{0xffffffff, 30, 0xffffffff, 2},
	} {
		if got := e2ePlayerPassiveHealingBound(tc.elapsed, tc.fps, tc.interval); got != tc.want {
			t.Fatalf("bound(%d,%d,%d)=%d, want %d", tc.elapsed, tc.fps, tc.interval, got, tc.want)
		}
	}
	// Every phase of ordinary regeneration fits below the bound. An observer
	// accepting passive healing alone would hide the old wrong-owner bug.
	for phase := uint32(0); phase < 20; phase++ {
		var passive uint32
		for elapsed := uint32(1); elapsed <= 150; elapsed++ {
			if elapsed > 30 && (phase+elapsed)%20 == 0 {
				passive++
			}
			if passive > e2ePlayerPassiveHealingBound(elapsed, 30, 20) {
				t.Fatalf("passive healing escaped bound at phase=%d elapsed=%d", phase, elapsed)
			}
		}
	}
}
