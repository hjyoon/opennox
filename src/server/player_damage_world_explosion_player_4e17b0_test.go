package server

import (
	"fmt"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestPlayerDamageWorldExplosionPlayer4E17B0Armor(t *testing.T) {
	for variant := 0; variant < 2; variant++ {
		for _, protected := range []bool{false, true} {
			t.Run(fmt.Sprintf("barrel-%d/protected-%t", variant+1, protected), func(t *testing.T) {
				target, source := damageMeleeUnitFixture4E17B0(t, true), damageWorldExplosionSource4E17B0(t, variant)
				armor := damageMeleeArmorFixture4E17B0(target, .5, .25)
				ud := target.UpdateDataPlayer()
				ud.Field76, ud.Field75 = 99, 77
				protection, wantHP := float64(0), uint16(185)
				if protected {
					protection, wantHP = .5, 192
				}
				r := damageFlameRuntime4E17B0(t, protection)
				damageMeleeArmorRuntime4E17B0(&r, armor, .5)
				if h, result := PlayerDamageNative4E17B0(target, source, nil, 30, object.DamageExplosion, r); !h || !result {
					t.Fatal("player barrel explosion rejected")
				}
				marker, typ, carry := damageMeleeMarker4E17B0(target)
				if target.HealthData.Cur != wantHP || armor.HealthData.Cur != 10 || marker != 2 || typ != 7 || carry != .25 || target.Obj130 != source || target.Pos132 != source.PrevPos || target.Field131 != 7 || target.Frame134 != 1400 {
					t.Fatalf("HP=%d/%d armor=%d marker=%d/%d carry=%g source=%p/%p", target.HealthData.Cur, wantHP, armor.HealthData.Cur, marker, typ, carry, target.Obj130, source)
				}
			})
		}
	}
}

func TestPlayerDamageWorldExplosionPlayer4E17B0EarlyGates(t *testing.T) {
	for _, gate := range []string{"no-update", "dead", "invulnerable", "Coop-owner", "status"} {
		t.Run(gate, func(t *testing.T) {
			target, source := damageMeleeUnitFixture4E17B0(t, true), damageWorldExplosionSource4E17B0(t, 0)
			ud := target.UpdateDataPlayer()
			ud.Field76, ud.Field75 = 99, 77
			r := damageFlameRuntime4E17B0(t, 0)
			sounds := 0
			r.Audio = func(id int, v *Object) {
				if gate != "invulnerable" || id != 71 || v != target {
					t.Fatal("early invulnerability sound")
				}
				sounds++
			}
			switch gate {
			case "no-update":
				target.ObjFlags |= object.FlagNoUpdate
				target.UpdateData = nil
			case "dead":
				target.ObjFlags |= object.FlagDead
				target.UpdateData = nil
			case "invulnerable":
				target.Buffs |= 1 << playerDamageInvulnerableEnchant4E17B0
				target.UpdateData = nil
			case "Coop-owner":
				source.ObjOwner = target
				target.UpdateData = nil
				r.CoopMode = func() bool { return true }
			case "status":
				ud.Player.Field3680 |= 1
			}
			r.DefaultDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
				t.Fatal("early world explosion reached HP")
				return false
			}
			r.BlockSourceOnlyExcluded = func(*Object) bool { t.Fatal("early world explosion reached source exclusion"); return false }
			h, result := PlayerDamageNative4E17B0(target, source, nil, 30, object.DamageExplosion, r)
			wantSounds := 0
			if gate == "invulnerable" {
				wantSounds = 1
			}
			if !h || result != (gate == "invulnerable") || target.HealthData.Cur != 200 || ud.Field76 != 99 || ud.Field75 != 77 || sounds != wantSounds {
				t.Fatalf("handled/result=%t/%t marker=%d/%d sounds=%d/%d", h, result, ud.Field76, ud.Field75, sounds, wantSounds)
			}
		})
	}
}

func TestPlayerDamageWorldExplosionPlayer4E17B0Shield(t *testing.T) {
	for _, front := range []bool{false, true} {
		t.Run(fmt.Sprintf("front-%t", front), func(t *testing.T) {
			target, source := damageMeleeUnitFixture4E17B0(t, true), damageWorldExplosionSource4E17B0(t, 0)
			ud := target.UpdateDataPlayer()
			ud.State, ud.Player.ArmorEquip, ud.Field76, ud.Field75 = PlayerState16, 0x3000000, 99, 77
			target.Buffs = 1 << playerDamageReflectEnchant4E17B0
			r := damageFlameRuntime4E17B0(t, 0)
			r.BlockDirection = func(v *Object, pos types.Pointf) bool {
				if v != target || pos != source.PrevPos {
					t.Fatal("barrel shield used missile position")
				}
				return front
			}
			sounds, wears := 0, 0
			r.Audio = func(id int, v *Object) {
				if id != 878 || v != target {
					t.Fatal("barrel shield audio identity")
				}
				sounds++
			}
			r.BlockDamagePercent = func() float64 { return .25 }
			r.DamageBlockItem = func(shield, v, s, w *Object, amount float32, typ object.DamageType) bool {
				if shield != nil || v != target || s != source || w != nil || amount != 7.5 || typ != object.DamageExplosion {
					t.Fatal("barrel shield wear identity/amount")
				}
				wears++
				return true
			}
			r.ProjectileReflect = func(*Object, *Object) { t.Fatal("world barrel was reflected as a missile") }
			h, result := PlayerDamageNative4E17B0(target, source, nil, 30, object.DamageExplosion, r)
			if !h {
				t.Fatal("world-source shield prefix rejected")
			}
			if front {
				if result || target.HealthData.Cur != 200 || sounds != 1 || wears != 1 || ud.Field76 != 0 || ud.Field75 != 77 {
					t.Fatal("frontal barrel shield did not preserve original block")
				}
			} else if !result || target.HealthData.Cur != 170 || sounds != 0 || wears != 0 || ud.Field76 != 2 || ud.Field75 != 7 {
				t.Fatal("rear barrel blast was incorrectly blocked")
			}
		})
	}
}
