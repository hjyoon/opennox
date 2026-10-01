package opennox

import (
	"strings"
	"testing"
	"time"

	"github.com/opennox/opennox/v1/server"
)

func TestE2EPlayerStatusAnimationSchedule(t *testing.T) {
	for kind, want := range map[string]server.EnchantID{
		"stun": server.ENCHANT_HELD, "confused": server.ENCHANT_CONFUSED,
		"nullify": server.ENCHANT_ANTI_MAGIC, "charming": server.ENCHANT_CHARMING,
		"shield": server.ENCHANT_SHIELD, "slow": server.ENCHANT_SLOWED,
	} {
		t.Run(kind, func(t *testing.T) {
			if got, ok := e2ePlayerStatusBuff(kind); !ok || got != want {
				t.Fatalf("buff = %d/%t, want %d", got, ok, want)
			}
			var sc e2eScenario
			sc.CheckPlayerStatusAnimation(kind, "animation")
			if len(sc.steps) != 7 {
				t.Fatalf("steps = %d, want prepare/two pixel checks/expiry/three screens", len(sc.steps))
			}
			for suffix, timeout := range map[string]time.Duration{
				" prepare": 1200, " first visible": 120, " advanced visible": 120, " natural expiry": 180,
			} {
				found := 0
				for _, step := range sc.steps {
					if strings.HasSuffix(step.name, suffix) {
						found++
						if step.ready == nil || step.fnc == nil || step.waitTimeout != timeout {
							t.Fatalf("%q lacks a bounded live predicate/result check", suffix)
						}
					}
				}
				if found != 1 {
					t.Fatalf("%q found %d times, want one", suffix, found)
				}
			}
		})
	}
}

func TestE2EPlayerStatusAnimationRejectsUnknown(t *testing.T) {
	for _, kind := range []string{"", "STUN", "poison", "freeze"} {
		var sc e2eScenario
		func() {
			defer func() {
				if recover() == nil || len(sc.steps) != 0 {
					t.Errorf("unknown status %q scheduled a mutation", kind)
				}
			}()
			sc.CheckPlayerStatusAnimation(kind, "invalid")
		}()
	}
}
