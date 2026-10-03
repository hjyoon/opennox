package server

import (
	"fmt"
	"math"
	"slices"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func damageGreatSwordUnblocked4E17B0(t *testing.T, player bool, typ object.DamageType) {
	t.Helper()
	for _, mode := range []string{"rear", "excluded", "unready"} {
		t.Run(mode, func(t *testing.T) {
			target, source, weapon, _, sword, _, r, events := damageGreatSwordFixture4E17B0(t, player, false)
			switch mode {
			case "rear":
				r.BlockDirection = func(*Object, types.Pointf) bool { return false }
			case "excluded":
				r.BlockSourceExcluded = func(*Object) bool { return true }
			case "unready":
				if player {
					target.UpdateDataPlayer().State = PlayerState0
				} else {
					target.UpdateDataMonster().AIStack[0].Action = uint32(ai.ACTION_BLOCK_ATTACK)
				}
			}
			r.DefaultDamage = func(v, s, w *Object, amount int32, got object.DamageType) bool {
				if v != target || s != source || w != weapon || amount != 9 || got != typ {
					t.Fatalf("unblocked tail args=%p/%p/%p/%d/%d", v, s, w, amount, got)
				}
				*events = append(*events, "default")
				v.HealthData.Cur -= uint16(amount)
				return true
			}
			h, result := PlayerDamageNative4E17B0(target, source, weapon, 9, typ, r)
			marker, markerType, carry := damageMeleeMarker4E17B0(target)
			wantCarry := float32(0.4)
			if typ == object.DamageExplosion || typ == object.DamageImpale {
				wantCarry = float32(9.4) - 9
			}
			if !h || !result || target.HealthData.Cur != 191 || sword.HealthData.Cur != 25 ||
				marker != 1 || markerType != uint32(weapon.TypeInd) || math.Float32bits(carry) != math.Float32bits(wantCarry) ||
				!slices.Equal(*events, []string{"default"}) {
				t.Fatalf("unblocked player=%t type=%d result=%t/%t HP=%d carry=%g marker=%d/%d events=%v",
					player, typ, h, result, target.HealthData.Cur, carry, marker, markerType, *events)
			}
		})
	}
}

func TestPlayerDamageNative4E17B0NPCGreatSwordUnblockedFlame(t *testing.T) {
	damageGreatSwordUnblocked4E17B0(t, false, object.DamageFlame)
}

func TestPlayerDamageNative4E17B0PlayerGreatSwordUnblockedFlame(t *testing.T) {
	damageGreatSwordUnblocked4E17B0(t, true, object.DamageFlame)
}

func TestPlayerDamageGreatSwordMissile4E17B0NPCActionAdmission(t *testing.T) {
	for action := 0; action < 72; action++ {
		t.Run(fmt.Sprint(action), func(t *testing.T) {
			target, source, weapon, _, _, cached, r, events := damageGreatSwordFixture4E17B0(t, false, false)
			target.UpdateDataMonster().AIStack[0].Action = uint32(action)
			a, h, result := playerDamageGreatSwordMissileBlock4E17B0(target, source, weapon, cached, 9, object.DamageFlame, r)
			want := slices.Contains([]int{0, 1, 4, 23, 25, 26, 27}, action)
			if a != want || h != want || result || !want && (*cached.marker != 88 || len(*events) != 0) {
				t.Fatalf("action=%d block=%t/%t/%t want=%t events=%v", action, a, h, result, want, *events)
			}
		})
	}
}
