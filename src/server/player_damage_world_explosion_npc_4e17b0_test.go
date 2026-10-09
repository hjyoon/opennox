package server

import (
	"fmt"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

func TestPlayerDamageWorldExplosionNPC4E17B0ArmorProtectionAndImmunity(t *testing.T) {
	for variant := 0; variant < 2; variant++ {
		for _, protected := range []bool{false, true} {
			for _, immune := range []bool{false, true} {
				t.Run(fmt.Sprintf("barrel-%d/protected-%t/immune-%t", variant+1, protected, immune), func(t *testing.T) {
					target, source := damageMeleeUnitFixture4E17B0(t, false), damageWorldExplosionSource4E17B0(t, variant)
					armor := damageMeleeArmorFixture4E17B0(target, .5, .25)
					ud := target.UpdateDataMonster()
					ud.Field547, ud.Field546 = 99, 77
					if immune {
						target.ObjSubClass |= 0x400
					}
					protection, damage := float64(0), 15
					if immune {
						damage = 7
					}
					if protected {
						protection = .5
						if immune {
							damage = 4
						} else {
							damage = 8
						}
					}
					r := damageFlameRuntime4E17B0(t, protection)
					damageMeleeArmorRuntime4E17B0(&r, armor, .5)
					r.GodMode = func() bool { t.Fatal("NPC barrel explosion queried player GodMode"); return true }
					if h, result := PlayerDamageNative4E17B0(target, source, nil, 30, object.DamageExplosion, r); !h || !result {
						t.Fatal("NPC barrel explosion rejected")
					}
					marker, typ, carry := damageMeleeMarker4E17B0(target)
					mask := object.MonStatusInjured | object.MonStatusOnFire
					if target.HealthData.Cur != uint16(200-damage) || armor.HealthData.Cur != 10 || marker != 2 || typ != 7 || carry != .25 || ud.StatusFlags&mask != mask ||
						target.Obj130 != source || target.Pos132 != source.PrevPos || target.Field131 != 7 || target.Frame134 != 1400 {
						t.Fatalf("HP=%d/%d armor=%d marker=%d/%d carry=%g status=%x", target.HealthData.Cur, 200-damage, armor.HealthData.Cur, marker, typ, carry, ud.StatusFlags)
					}
				})
			}
		}
	}
}

func TestPlayerDamageWorldExplosionNPC4E17B0Shield(t *testing.T) {
	for _, front := range []bool{false, true} {
		t.Run(fmt.Sprintf("front-%t", front), func(t *testing.T) {
			target, source := damageMeleeUnitFixture4E17B0(t, false), damageWorldExplosionSource4E17B0(t, 0)
			ud := target.UpdateDataMonster()
			ud.ArmorEquipFlags, ud.Field547, ud.Field546 = 0x3000000, 99, 77
			ud.AIStackInd, ud.AIStack[0].Action = 0, uint32(ai.ACTION_BLOCK_ATTACK)
			r := damageFlameRuntime4E17B0(t, 0)
			r.BlockDirection = func(v *Object, pos types.Pointf) bool {
				if v != target || pos != source.PrevPos {
					t.Fatal("NPC barrel shield used wrong position")
				}
				return front
			}
			sounds, wears := 0, 0
			r.Audio = func(id int, v *Object) {
				if id != 878 || v != target {
					t.Fatal("NPC barrel shield audio identity")
				}
				sounds++
			}
			r.BlockDamagePercent = func() float64 { return .25 }
			r.DamageBlockItem = func(shield, v, s, w *Object, amount float32, typ object.DamageType) bool {
				if shield != nil || v != target || s != source || w != nil || amount != 7.5 || typ != object.DamageExplosion {
					t.Fatal("NPC barrel shield wear identity/amount")
				}
				wears++
				return true
			}
			r.ProjectileReflect = func(*Object, *Object) { t.Fatal("NPC reflected a nonmissile barrel") }
			h, result := PlayerDamageNative4E17B0(target, source, nil, 30, object.DamageExplosion, r)
			if !h {
				t.Fatal("NPC world-source shield prefix rejected")
			}
			if front {
				if result || target.HealthData.Cur != 200 || sounds != 1 || wears != 1 || ud.Field547 != 0 || ud.Field546 != 77 {
					t.Fatal("NPC frontal barrel block failed")
				}
			} else if !result || target.HealthData.Cur != 170 || sounds != 0 || wears != 0 || ud.Field547 != 2 || ud.Field546 != 7 {
				t.Fatal("NPC rear barrel blast was blocked")
			}
		})
	}
}

func TestPlayerDamageWorldExplosionNPC4E17B0EarlyGates(t *testing.T) {
	for _, gate := range []string{"no-update", "dead", "invulnerable"} {
		t.Run(gate, func(t *testing.T) {
			target, source := damageMeleeUnitFixture4E17B0(t, false), damageWorldExplosionSource4E17B0(t, 0)
			ud := target.UpdateDataMonster()
			ud.Field547, ud.Field546 = 99, 77
			r := damageFlameRuntime4E17B0(t, 0)
			sounds := 0
			r.Audio = func(id int, v *Object) {
				if gate != "invulnerable" || id != 71 || v != target {
					t.Fatal("NPC early barrel audio")
				}
				sounds++
			}
			switch gate {
			case "no-update":
				target.ObjFlags |= object.FlagNoUpdate
			case "dead":
				target.ObjFlags |= object.FlagDead
			case "invulnerable":
				target.Buffs |= 1 << playerDamageInvulnerableEnchant4E17B0
			}
			r.DefaultDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
				t.Fatal("NPC early barrel reached HP")
				return false
			}
			h, result := PlayerDamageNative4E17B0(target, source, nil, 30, object.DamageExplosion, r)
			wantSounds := 0
			if gate == "invulnerable" {
				wantSounds = 1
			}
			if !h || result != (gate == "invulnerable") || target.HealthData.Cur != 200 || ud.Field547 != 99 || ud.Field546 != 77 || sounds != wantSounds {
				t.Fatal("NPC early barrel marker/HP changed")
			}
		})
	}
}
