package server

import (
	"fmt"
	"math"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

func worldMagicMissileFixture4E17B0(owner string) (source, missile *Object) {
	missile = &Object{TypeInd: 697, ObjClass: object.ClassMissile | object.ClassSimple | object.ClassLight,
		PrevPos: types.Ptf(23, 9), PosVec: types.Ptf(25, 11)}
	switch owner {
	case "terminal":
		return missile, missile
	case "immobile":
		source = &Object{TypeInd: 600, ObjClass: object.ClassImmobile}
	case "imaginary":
		source = &Object{TypeInd: 1399}
	default:
		panic(owner)
	}
	missile.ObjOwner = source
	return source, missile
}

func TestPlayerDamageWorldMagicMissile4E17B0Defenses(t *testing.T) {
	for _, owner := range []string{"terminal", "immobile", "imaginary"} {
		for _, player := range []bool{false, true} {
			for _, gate := range []string{"reflect", "shield", "invulnerable", "dead", "disabled", "coop-owner", "god"} {
				t.Run(fmt.Sprintf("%s/player-%t/%s", owner, player, gate), func(t *testing.T) {
					v := damageMeleeUnitFixture4E17B0(t, player)
					a, w := worldMagicMissileFixture4E17B0(owner)
					r := damageMeleeRuntimeFixture4E17B0(t)
					reflections, owners, audio, hpCalls := 0, 0, 0, 0
					wear := float32(0)
					r.BlockDirection = func(*Object, types.Pointf) bool { return true }
					r.ProjectileReflect = func(m, target *Object) {
						if m != w || target != v {
							t.Fatal("wrong reflected objects")
						}
						reflections++
					}
					r.ClearOwner = func(m *Object) { owners++; m.ObjOwner = nil }
					r.SetOwner = func(target, m *Object) { owners++; m.ObjOwner = target }
					r.Audio = func(id int, _ *Object) { audio = id }
					r.BlockDamagePercent = func() float64 { return .25 }
					r.DamageBlockItem = func(item, target, source, weapon *Object, d float32, typ object.DamageType) bool {
						if item != nil || target != v || source != a || weapon != w || typ != object.DamageExplosion {
							t.Fatal("wrong shield wear arguments")
						}
						wear = d
						return true
					}
					r.DefaultDamage = func(target, source, weapon *Object, d int32, typ object.DamageType) bool {
						if target != v || source != a || weapon != w || d != 9 || typ != object.DamageExplosion {
							t.Fatal("wrong default damage arguments")
						}
						hpCalls++
						return true
					}
					switch gate {
					case "reflect":
						v.Buffs = 1 << playerDamageReflectEnchant4E17B0
					case "shield":
						if player {
							ud := v.UpdateDataPlayer()
							ud.State, ud.Player.ArmorEquip = PlayerState16, 0x1000000
						} else {
							ud := v.UpdateDataMonster()
							ud.AIStack[0].Action, ud.ArmorEquipFlags = uint32(ai.ACTION_BLOCK_ATTACK), 0x1000000
						}
					case "invulnerable":
						v.Buffs = 1 << playerDamageInvulnerableEnchant4E17B0
					case "dead":
						v.ObjFlags |= object.FlagDead
					case "disabled":
						v.ObjFlags |= object.FlagNoUpdate
					case "coop-owner":
						a.ObjOwner = v
						r.CoopMode = func() bool { return true }
					case "god":
						r.GodMode = func() bool { return true }
					}
					h, ok := PlayerDamageNative4E17B0(v, a, w, 9, object.DamageExplosion, r)
					wantHP := !player && (gate == "coop-owner" || gate == "god")
					wantResult := wantHP || gate == "invulnerable" || gate == "god" && player
					if !h || ok != wantResult || (hpCalls != 0) != wantHP || v.HealthData.Cur != 200 {
						t.Fatalf("gate handled/result=%t/%t HP calls=%d", h, ok, hpCalls)
					}
					if gate == "reflect" || gate == "shield" {
						wantAudio := 122
						if gate == "shield" {
							wantAudio = 878
							if wear != 2.25 {
								t.Fatalf("shield wear=%g want=2.25", wear)
							}
						}
						if reflections != 1 || owners != 2 || w.ObjOwner != v || audio != wantAudio {
							t.Fatalf("reflect=%d ownership=%d audio=%d", reflections, owners, audio)
						}
					}
				})
			}
		}
	}
}

func TestWorldMagicMissileShape4E17B0Boundaries(t *testing.T) {
	for _, owner := range []string{"terminal", "immobile", "imaginary"} {
		for _, extra := range []object.Class{0, object.ClassPlayer, object.ClassMonster, object.ClassWeapon, object.ClassWand} {
			t.Run(fmt.Sprintf("%s/%x", owner, extra), func(t *testing.T) {
				a, w := worldMagicMissileFixture4E17B0(owner)
				w.ObjClass |= extra
				if got := playerDamageWorldProjectileShape4E17B0(a, w, object.DamageExplosion); got != (extra == 0) {
					t.Fatalf("missile class=%x admitted=%t", w.Class(), got)
				}
			})
		}
	}
	if playerDamageWorldProjectileShape4E17B0(nil, nil, object.DamageExplosion) {
		t.Fatal("nil explosion admitted")
	}
}

// BoomCollide passes FindOwnerChainPlayer(missile), missile, type 7. A
// terminal missile or a non-unit script caster has no unit update record.
func TestPlayerDamageWorldMagicMissile4E17B0ArmorAndHP(t *testing.T) {
	for _, owner := range []string{"terminal", "immobile", "imaginary"} {
		for _, player := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/player-%t", owner, player), func(t *testing.T) {
				v := damageMeleeUnitFixture4E17B0(t, player)
				a, w := worldMagicMissileFixture4E17B0(owner)
				if w.FindOwnerChainPlayer() != a {
					t.Fatal("fixture does not represent BoomCollide's real arguments")
				}
				armor := damageMeleeArmorFixture4E17B0(v, .5, .25)
				r := damageMeleeRuntimeFixture4E17B0(t)
				damageMeleeArmorRuntime4E17B0(&r, armor, .5)
				world := damageMeleeWorldRuntime4E0B30(t)
				world.FireProtection = func(*Object) float64 { return 0 }
				r.DefaultDamage = func(v, a, w *Object, d int32, typ object.DamageType) bool {
					return DefaultDamageWorld4E0B30(v, a, w, d, typ, world)
				}
				before := *w
				for hit, hp := range []uint16{195, 191, 186} {
					if h, ok := PlayerDamageNative4E17B0(v, a, w, 9, object.DamageExplosion, r); !h || !ok || v.HealthData.Cur != hp {
						t.Fatalf("hit=%d handled/result=%t/%t HP=%d want=%d", hit, h, ok, v.HealthData.Cur, hp)
					}
					marker, kind, carry := damageMeleeMarker4E17B0(v)
					wantMarker, wantKind := uint32(1), uint32(w.TypeInd)
					if owner == "terminal" {
						wantMarker, wantKind = 2, 7
					}
					if marker != wantMarker || kind != wantKind || carry != []float32{-.25, .25, -.25}[hit] ||
						v.Obj130 != w || v.Pos132 != w.PrevPos || v.Field131 != 7 || *w != before {
						t.Fatalf("attribution=%p marker=%d/%d carry=%g", v.Obj130, marker, kind, carry)
					}
				}
				if armor.HealthData.Cur != 12 {
					t.Fatalf("armor HP=%d want=12", armor.HealthData.Cur)
				}
			})
		}
	}
}

func TestDefaultDamageWorldMagicMissile4E0B30SignedFireImmunity(t *testing.T) {
	for _, owner := range []string{"terminal", "immobile", "imaginary"} {
		for _, player := range []bool{false, true} {
			for _, raw := range []int32{math.MinInt32, -11, -9, -1, 0, 1, 9, 11, math.MaxInt32} {
				t.Run(fmt.Sprintf("%s/player-%t/%d", owner, player, raw), func(t *testing.T) {
					v := damageMeleeUnitFixture4E17B0(t, player)
					v.ObjSubClass |= 0x400
					a, w := worldMagicMissileFixture4E17B0(owner)
					r := damageMeleeWorldRuntime4E0B30(t)
					r.FireProtection = func(*Object) float64 { return 0 }
					got := int32(123456)
					r.DamageClear = func(_ *Object, d int32) { got = d }
					want := raw
					if !player {
						want /= 2 // Original signed division precedes protection/minimum.
					}
					want = defaultDamageOriginalProtected4E0B30(t, want, 0)
					if !DefaultDamageWorld4E0B30(v, a, w, raw, object.DamageExplosion, r) || got != want {
						t.Fatalf("HP input=%d want=%d", got, want)
					}
				})
			}
		}
	}
}
