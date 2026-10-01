package opennox

import (
	"strings"
	"testing"

	"github.com/opennox/libs/object"
)

func TestE2EPlayerThrownWeaponSchedule(t *testing.T) {
	for _, tc := range []struct {
		item, projectile string
		flag             object.WeaponClass
	}{
		{"FanChakram", "FanChakramInMotion", object.WeaponShuriken},
		{"RoundChakram", "RoundChakramInMotion", object.WeaponChakram},
	} {
		t.Run(tc.item, func(t *testing.T) {
			if projectile, flag, ok := e2ePlayerThrownWeaponType(tc.item); !ok || projectile != tc.projectile || flag != tc.flag {
				t.Fatalf("type=%s/%#x/%t", projectile, flag, ok)
			}
			var sc e2eScenario
			sc.CheckPlayerThrownWeapon(tc.item, tc.item)
			if len(sc.steps) != 5 {
				t.Fatalf("steps=%d, want 5", len(sc.steps))
			}
			for _, index := range []int{0, 4} {
				step := sc.steps[index]
				if step.ready == nil || step.fnc == nil || step.waitTimeout == 0 {
					t.Fatalf("step %d has no bounded live check", index)
				}
			}
			if !strings.HasSuffix(sc.steps[4].name, " actual projectile launch") || sc.steps[4].waitTimeout != 90 {
				t.Fatal("no strict launch assertion")
			}
		})
	}
}

func TestE2EPlayerThrownWeaponRejectsUnknownItem(t *testing.T) {
	var sc e2eScenario
	defer func() {
		if recover() == nil || len(sc.steps) != 0 {
			t.Fatal("unknown weapon scheduled")
		}
	}()
	sc.CheckPlayerThrownWeapon("not-a-weapon", "invalid")
}

func TestE2EPlayerThrownRoundDoesNotReadAmmo(t *testing.T) {
	f := &e2ePlayerThrownWeaponFixture{flag: object.WeaponChakram}
	if f.currentCharge() != 0 {
		t.Fatal("RoundChakram invented ammo")
	}
}
