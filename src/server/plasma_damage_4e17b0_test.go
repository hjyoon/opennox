package server

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// The sealed PE switch at 004E20A8 sends PLASMA (14), Mana Bomb (15)
// and Death Ray (16) to the same raw HP/marker tail at 004E1E83.
func TestPlasmaDamageNative4E17B0UnitPairs(t *testing.T) {
	for _, from := range []string{"player", "monster", "NPC"} {
		for _, to := range []string{"player", "monster", "NPC"} {
			t.Run(from+"-to-"+to, func(t *testing.T) {
				target, source := damageRawSpellUnit4E17B0(t, to), damageRawSpellUnit4E17B0(t, from)
				world := damageMeleeWorldRuntime4E0B30(t)
				world.FireProtection = func(*Object) float64 { t.Fatal("Plasma used fire protection"); return 1 }
				world.ElectricProtection = func(*Object) float64 { t.Fatal("Plasma used electric protection"); return 1 }
				var armor *Object
				prefix := damageMeleeRuntimeFixture4E17B0(t)
				if to != "monster" {
					armor = damageMeleeArmorFixture4E17B0(target, .9, .4)
					prefix.ItemArmorValue = func(*Object) float32 { t.Fatal("Plasma absorbed by armor"); return 1 }
					prefix.DamageArmor = func(*Object, *Object, *Object, int32, object.DamageType) bool {
						t.Fatal("Plasma wore armor")
						return false
					}
					prefix.ElectricArmorScale = func(*Object) float32 { t.Fatal("Plasma used electric armor"); return 1 }
				}
				prefix.DefaultDamage = func(v, a, w *Object, d int32, k object.DamageType) bool {
					return DefaultDamageWorld4E0B30(v, a, w, d, k, world)
				}
				for hit, want := range []uint16{191, 182, 173} {
					if to == "monster" {
						if !DefaultDamageWorld4E0B30(target, source, nil, 9, object.DamagePlasma, world) {
							t.Fatal("ordinary monster Plasma rejected")
						}
					} else if handled, result := PlayerDamageNative4E17B0(target, source, nil, 9, object.DamagePlasma, prefix); !handled || !result {
						t.Fatal("player/NPC Plasma prefix rejected")
					}
					if target.HealthData.Cur != want || source.HealthData.Cur != 200 || target.Obj130 != source || target.Pos132 != source.PrevPos || target.Field131 != 14 || target.Frame134 != 1400 {
						t.Fatalf("hit=%d HP=%d source=%p type=%d frame=%d", hit, target.HealthData.Cur, target.Obj130, target.Field131, target.Frame134)
					}
					if to != "player" {
						ud := target.UpdateDataMonster()
						if ud.Field547 != 2 || ud.Field546 != 14 || !ud.StatusFlags.Has(object.MonStatusInjured) || !ud.StatusFlags.Has(object.MonStatusOnFire) {
							t.Fatalf("Plasma monster marker=%d/%d flags=%x", ud.Field547, ud.Field546, ud.StatusFlags)
						}
					}
					if armor != nil {
						marker, kind, carry := damageMeleeMarker4E17B0(target)
						if marker != 2 || kind != 14 || carry != .4 || armor.HealthData.Cur != 25 {
							t.Fatalf("raw prefix marker/carry/armor=%d/%d/%g/%d", marker, kind, carry, armor.HealthData.Cur)
						}
					}
				}
			})
		}
	}
}

func TestPlasmaDamageShapeNative4E17B0(t *testing.T) {
	for _, class := range []object.Class{0, object.ClassPlayer, object.ClassMonster, object.ClassPlayer | object.ClassMonster, object.ClassMissile, object.ClassWeapon, object.ClassWand, object.ClassPlayer | object.ClassMissile} {
		source := &Object{ObjClass: class}
		want := class == 0 || (class.HasAny(object.MaskUnits) && !class.HasAny(object.ClassMissile|object.ClassWeapon|object.ClassWand))
		if got := playerDamageWeaponlessRawSpellShape4E17B0(source, nil, object.DamagePlasma); got != want {
			t.Errorf("class=%x admitted=%t want=%t", class, got, want)
		}
		if playerDamageWeaponlessRawSpellShape4E17B0(source, source, object.DamagePlasma) {
			t.Errorf("class=%x admitted nonnil weapon", class)
		}
	}
	if !playerDamageWeaponlessRawSpellShape4E17B0(nil, nil, object.DamagePlasma) || playerDamageWeaponlessRawSpellShape4E17B0(nil, nil, object.DamageAirborneElectric) {
		t.Fatal("nil-source Plasma/electric shape changed")
	}
}

func TestPlasmaDamageNative4E17B0ReflectAndPhysicalShield(t *testing.T) {
	for _, to := range []string{"player", "NPC"} {
		for _, reflect := range []bool{false, true} {
			for _, front := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/reflect-%t/front-%t", to, reflect, front), func(t *testing.T) {
					target, source, shield := rawSpellBlockFixture4E17B0(t, to)
					if reflect {
						target.Buffs |= 1 << ENCHANT_REFLECTIVE_SHIELD
					}
					r := damageMeleeRuntimeFixture4E17B0(t)
					var events []string
					r.PointFX = func(int, types.Pointf) { t.Fatal("Plasma reflected as Death Ray") }
					r.ProjectileReflect = func(*Object, *Object) { t.Fatal("unit Plasma reflected as missile") }
					r.BlockSourceOnlyExcluded = func(got *Object) bool {
						if got != source {
							t.Fatal("wrong source-only exclusion")
						}
						events = append(events, "exclude")
						return false
					}
					r.BlockDirection = func(got *Object, pos types.Pointf) bool {
						if got != target || pos != source.PrevPos {
							t.Fatal("Plasma used Reflect facing instead of shield facing")
						}
						events = append(events, "front")
						return front
					}
					r.Audio = func(id int, got *Object) {
						if id != 878 || got != target {
							t.Fatal("wrong shield audio")
						}
						events = append(events, "audio")
					}
					r.BlockDamagePercent = func() float64 { events = append(events, "balance"); return .5 }
					r.CanDamageBlockItem = func(got *Object) bool { return got == shield }
					r.DamageBlockItem = func(item, owner, attacker, weapon *Object, amount float32, typ object.DamageType) bool {
						if item != shield || owner != target || attacker != source || weapon != nil || amount != 12.5 || typ != object.DamagePlasma {
							t.Fatal("wrong Plasma shield wear")
						}
						events = append(events, "wear")
						return true
					}
					h, result := PlayerDamageNative4E17B0(target, source, nil, 25, object.DamagePlasma, r)
					wantHP, wantEvents := uint16(175), []string{"exclude", "front"}
					if front {
						wantHP, wantEvents = 200, []string{"exclude", "front", "audio", "balance", "wear"}
					}
					if !h || result == front || target.HealthData.Cur != wantHP || !slices.Equal(events, wantEvents) {
						t.Fatalf("handled/result/HP=%t/%t/%d events=%v", h, result, target.HealthData.Cur, events)
					}
				})
			}
		}
	}
}

func TestPlasmaDamageWorld4E0B30FriendlyAndElementalImmunity(t *testing.T) {
	for _, to := range []string{"monster", "NPC"} {
		for _, gameplay := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/gameplay-%t", to, gameplay), func(t *testing.T) {
				target, source := damageRawSpellUnit4E17B0(t, to), damageRawSpellUnit4E17B0(t, "player")
				target.ObjSubClass |= 0xe00
				r := damageMeleeWorldRuntime4E0B30(t)
				r.GameplayFlag1 = func() bool { return gameplay }
				r.IsEnemy = func(*Object, *Object) bool { return false }
				DefaultDamageWorld4E0B30(target, source, nil, 9, object.DamagePlasma, r)
				want := uint16(200)
				if gameplay {
					want = 191
				}
				if target.HealthData.Cur != want || !target.UpdateDataMonster().StatusFlags.Has(object.MonStatusOnFire) {
					t.Fatalf("friendly/immune Plasma HP=%d want=%d flags=%x", target.HealthData.Cur, want, target.UpdateDataMonster().StatusFlags)
				}
			})
		}
	}
}

func TestPlasmaDamageNative4E17B0SignedQuestAndGod(t *testing.T) {
	for _, to := range []string{"player", "NPC"} {
		for _, god := range []bool{false, true} {
			for _, raw := range []int32{-3, 0, 1, 5, 9} {
				t.Run(fmt.Sprintf("%s/god-%t/raw-%d", to, god, raw), func(t *testing.T) {
					target := damageRawSpellUnit4E17B0(t, to)
					r := damageMeleeRuntimeFixture4E17B0(t)
					r.GodMode, r.QuestMode = func() bool { return god }, func() bool { return true }
					r.QuestDamageScale = func() float32 { return .5 }
					calls, damage := 0, int32(999)
					r.DefaultDamage = func(got, a, w *Object, d int32, k object.DamageType) bool {
						if got != target || a != nil || w != nil || k != object.DamagePlasma {
							t.Fatal("signed Plasma context")
						}
						calls, damage = calls+1, d
						return true
					}
					if h, result := PlayerDamageNative4E17B0(target, nil, nil, raw, object.DamagePlasma, r); !h || !result {
						t.Fatal("signed Plasma rejected")
					}
					if god && to == "player" {
						if calls != 0 {
							t.Fatal("GodMode still damaged player")
						}
					} else {
						want := int32(math.RoundToEven(float64(raw) * .5))
						if raw > 0 && want < 1 {
							want = 1
						}
						if calls != 1 || damage != want {
							t.Fatalf("calls/raw=%d/%d want=%d", calls, damage, want)
						}
					}
				})
			}
		}
	}
}

func TestPlasmaDamageWorld4E0B30LateDefenseAndShield(t *testing.T) {
	for _, to := range []string{"player", "NPC"} {
		t.Run(to, func(t *testing.T) {
			target, source := damageRawSpellUnit4E17B0(t, to), damageRawSpellUnit4E17B0(t, "monster")
			late := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
			item := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, late}})}
			target.InvFirstItem, target.Buffs = item, 1<<defaultDamageShieldEnchant4E0B30
			var events []string
			r := damageMeleeWorldRuntime4E0B30(t)
			r.BuffOff = func(*Object, EnchantID) { events = append(events, "invisible") }
			r.CanApplyLateDefend = func(m *ModifierEff) bool { return m == late }
			r.ApplyLateDefend = func(m *ModifierEff, i, v, w, a *Object, d int32, k object.DamageType) int32 {
				if m != late || i != item || v != target || w != nil || a != source || d != 19 || k != object.DamagePlasma {
					t.Fatal("Plasma late defense context")
				}
				events = append(events, "defend")
				return 11
			}
			r.DefaultDamageSound = func(*Object, *Object) { events = append(events, "sound") }
			r.ShieldReduce = func(got *Object, d *int32, k object.DamageType, a *Object) {
				if got != target || *d != 11 || k != object.DamagePlasma || a != source || target.Obj130 != source {
					t.Fatal("Plasma Shield context")
				}
				events = append(events, "shield")
				*d = 3
			}
			clear := r.DamageClear
			r.DamageClear = func(got *Object, raw int32) { events = append(events, "hp"); clear(got, raw) }
			if !DefaultDamageWorld4E0B30(target, source, nil, 19, object.DamagePlasma, r) || target.HealthData.Cur != 197 || !slices.Equal(events, []string{"invisible", "defend", "sound", "shield", "hp"}) {
				t.Fatalf("HP=%d events=%v", target.HealthData.Cur, events)
			}
		})
	}
}

func TestPlasmaDamageNative4E17B0EarlyGates(t *testing.T) {
	for _, to := range []string{"player", "NPC"} {
		for _, gate := range []string{"Dead", "NoUpdate", "Invulnerable", "CoopSelf"} {
			if gate == "CoopSelf" && to != "player" {
				continue // Original self-owner gate is player-only.
			}
			t.Run(to+"/"+gate, func(t *testing.T) {
				target, source := damageRawSpellUnit4E17B0(t, to), damageRawSpellUnit4E17B0(t, "monster")
				r := damageMeleeRuntimeFixture4E17B0(t)
				r.DefaultDamage = func(*Object, *Object, *Object, int32, object.DamageType) bool {
					t.Fatal("early gate entered HP tail")
					return false
				}
				want, audio := false, 0
				switch gate {
				case "Dead":
					target.ObjFlags |= object.FlagDead
				case "NoUpdate":
					target.ObjFlags |= object.FlagNoUpdate
				case "Invulnerable":
					target.Buffs = 1 << ENCHANT_INVULNERABLE
					want = true
					r.Audio = func(id int, got *Object) {
						if id != 71 || got != target {
							t.Fatal("invulnerability audio")
						}
						audio++
					}
				case "CoopSelf":
					source = target
					r.CoopMode = func() bool { return true }
				}
				h, result := PlayerDamageNative4E17B0(target, source, nil, 9, object.DamagePlasma, r)
				wantAudio := 0
				if gate == "Invulnerable" {
					wantAudio = 1
				}
				if !h || result != want || target.HealthData.Cur != 200 || target.Obj130 != nil || audio != wantAudio {
					t.Fatalf("gate handled/result/HP/audio=%t/%t/%d/%d", h, result, target.HealthData.Cur, audio)
				}
			})
		}
	}
}
