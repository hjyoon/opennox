package server

import (
	"fmt"
	"image"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

// 004E11BF queries buff 26 after every hit callback, not at admission.
// These fixtures change only Shield; the admission shape stays supported.
func TestDefaultDamageWorld4E0B30LiveShieldAfterCallbacks(t *testing.T) {
	for _, stage := range []string{
		"fire", "electric", "protection-audio", "buff-off", "late-defend", "pre-damage",
		"monster-hit-sound", "damage-sound", "vampirism-fx", "game-ball", "hurt-state", "field-guide", "source-hit-enemy",
	} {
		for _, initial := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/initial-%t", stage, initial), func(t *testing.T) {
				target := damageMeleeUnitFixture4E17B0(t, stage == "game-ball" || stage == "hurt-state")
				source := damageMeleeUnitFixture4E17B0(t, stage != "monster-hit-sound" && stage != "source-hit-enemy")
				weapon := &Object{TypeInd: 777, ObjClass: object.ClassWand, InitData: unsafe.Pointer(&ModifierInitData{})}
				if initial {
					target.Buffs = 1 << defaultDamageShieldEnchant4E0B30
				}
				mutations := 0
				mutate := func() {
					mutations++
					target.Buffs ^= 1 << defaultDamageShieldEnchant4E0B30
				}
				r := damageMeleeWorldRuntime4E0B30(t)
				typ, incoming, beforeShield := object.DamageBlade, int32(8), int32(8)
				switch stage {
				case "fire", "protection-audio":
					target.ObjClass = 0 // World-object fire shape; no NPC admission expansion.
					typ, beforeShield = object.DamageFlame, 6
					r.FireProtection = func(*Object) float64 {
						if stage == "fire" {
							mutate()
						}
						return 0.25
					}
					r.Audio = func(id int, got *Object) {
						if id != 104 || got != target {
							t.Fatal("incorrect fire protection audio")
						}
						if stage == "protection-audio" {
							mutate()
						}
					}
				case "electric":
					weapon, typ, beforeShield = nil, object.DamageElectric, 6
					r.ElectricProtection = func(*Object) float64 { mutate(); return 0.25 }
				case "buff-off":
					r.BuffOff = func(got *Object, id EnchantID) {
						if got != target || id != defaultDamageInvisibleEnchant4E0B30 {
							t.Fatal("incorrect BuffOff prefix")
						}
						mutate()
					}
				case "late-defend":
					effect := defaultDamageLateEffectFixture4E0B30()
					item := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped,
						InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, effect}})}
					target.InvFirstItem = item
					r.CanApplyLateDefend = func(m *ModifierEff) bool { return m == effect }
					r.ApplyLateDefend = func(*ModifierEff, *Object, *Object, *Object, *Object, int32, object.DamageType) int32 {
						mutate()
						return incoming
					}
				case "pre-damage":
					effect := defaultDamagePreEffectFixture4E0B30()
					(*ModifierInitData)(weapon.InitData).Modifiers[0] = effect
					r.CanApplyPreDamage = func(m *ModifierEff) bool { return m == effect }
					r.ApplyPreDamage = func(*ModifierEff, *Object, *Object, *Object, *int32) { mutate() }
				case "monster-hit-sound":
					r.MonsterHasHitSound = func(*Object) bool { mutate(); return false }
				case "damage-sound":
					r.DefaultDamageSound = func(*Object, *Object) { mutate() }
				case "vampirism-fx":
					source.Buffs, source.BuffsPower[damageVampirismEnchant4E0B30] = 1<<damageVampirismEnchant4E0B30, 1
					r.Audio = func(int, *Object) {}
					r.BalanceFloatInd = func(string, int) float64 { return 0.5 }
					r.AdjustHP = func(*Object, int32) {}
					r.VampirismFX = func(int, image.Point, image.Point, uint16) { mutate() }
				case "game-ball":
					r.GameBallOnDamage = func(*Object, *Object, int32) { mutate() }
				case "hurt-state":
					incoming, beforeShield = 32, 32
					r.PlayerSetState = func(*Object, PlayerState) bool { mutate(); return true }
				case "field-guide":
					r.AdjustFieldGuide = func(*Object, *Object, int32) int32 { mutate(); return incoming }
				case "source-hit-enemy":
					r.IsEnemy = func(*Object, *Object) bool {
						if target.Frame134 == 1400 {
							mutate()
						}
						return true
					}
				}
				shieldCalls, clearCalls := 0, 0
				attribution := weapon
				if attribution == nil {
					attribution = source
				}
				r.ShieldReduce = func(got *Object, damage *int32, gotType object.DamageType, gotSource *Object) {
					shieldCalls++
					if got != target || gotSource != attribution || gotType != typ || *damage != beforeShield ||
						!target.HasEnchant(defaultDamageShieldEnchant4E0B30) || target.Obj130 != attribution || target.Frame134 != 1400 ||
						(source.Class().Has(object.ClassMonster) && source.UpdateDataMonster().Field130 != 1400) {
						t.Fatal("Shield used stale state or ran before attribution/source timestamp")
					}
					*damage /= 2
				}
				clear := r.DamageClear
				r.DamageClear = func(got *Object, damage int32) { clearCalls++; clear(got, damage) }
				if !DefaultDamageWorld4E0B30(target, source, weapon, incoming, typ, r) {
					t.Fatal("nonzero live Shield tail returned false")
				}
				wantDamage, wantShield := beforeShield, 0
				if !initial {
					wantDamage, wantShield = beforeShield/2, 1
				}
				if mutations != 1 || clearCalls != 1 || shieldCalls != wantShield || target.HealthData.Cur != uint16(200-wantDamage) {
					t.Fatalf("live Shield: mutations=%d clears=%d reductions=%d HP=%d, want reductions=%d HP=%d",
						mutations, clearCalls, shieldCalls, target.HealthData.Cur, wantShield, 200-wantDamage)
				}
			})
		}
	}
}

func TestDefaultDamageWorld4E0B30LiveShieldBypassAndRawZero(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamagePoison, object.DamageManaBomb} {
		for _, live := range []bool{false, true} {
			for _, damage := range []int32{0, 8, -1, math.MinInt32, math.MaxInt32} {
				t.Run(fmt.Sprintf("%s/live-%t/raw-%d", typ, live, damage), func(t *testing.T) {
					target := damageMeleeUnitFixture4E17B0(t, false)
					var source *Object
					if typ == object.DamageManaBomb {
						target.ObjClass = 0 // Original non-unit self-ManaBomb gate.
						source = target
					}
					// Flip at the last sound callback, so the entry snapshot is wrong.
					if !live {
						target.Buffs = 1 << defaultDamageShieldEnchant4E0B30
					}
					r := damageMeleeWorldRuntime4E0B30(t)
					r.DefaultDamageSound = func(*Object, *Object) { target.Buffs ^= 1 << defaultDamageShieldEnchant4E0B30 }
					// Neither POISON nor self-ManaBomb requires ShieldReduce.
					r.ShieldReduce = nil
					clears := 0
					r.DamageClear = func(got *Object, raw int32) {
						clears++
						if got != target || raw != damage {
							t.Fatal("bypass changed raw signed damage")
						}
					}
					want := !(live && typ == object.DamageManaBomb && damage == 0)
					got := DefaultDamageWorld4E0B30(target, source, nil, damage, typ, r)
					if got != want || clears != boolIntLiveShield4E0B30(want) || target.HasEnchant(defaultDamageShieldEnchant4E0B30) != live {
						t.Fatalf("raw Shield bypass: result=%t clears=%d live=%t, want result=%t", got, clears, target.HasEnchant(defaultDamageShieldEnchant4E0B30), want)
					}
				})
			}
		}
	}
}

func boolIntLiveShield4E0B30(value bool) int {
	if value {
		return 1
	}
	return 0
}

func TestDefaultDamageWorld4E0B30LiveShieldMissingService(t *testing.T) {
	for _, initial := range []bool{false, true} {
		t.Run(fmt.Sprintf("initial-%t", initial), func(t *testing.T) {
			target, source := damageMeleeUnitFixture4E17B0(t, false), damageMeleeUnitFixture4E17B0(t, true)
			if initial {
				target.Buffs = 1 << defaultDamageShieldEnchant4E0B30
			}
			r := damageMeleeWorldRuntime4E0B30(t)
			var events []string
			r.BuffOff = func(*Object, EnchantID) {
				events = append(events, "buff-off")
				target.Buffs |= 1 << defaultDamageShieldEnchant4E0B30
			}
			r.DefaultDamageSound = func(*Object, *Object) { events = append(events, "sound") }
			r.ShieldReduce = nil
			r.Unsupported = func(reason string, got, gotSource, weapon *Object, damage int32, typ object.DamageType) {
				events = append(events, reason)
				if got != target || gotSource != source || weapon != nil || damage != 8 || typ != object.DamageClaw {
					t.Fatal("missing Shield service lost original arguments")
				}
			}
			r.DamageClear = func(*Object, int32) { t.Fatal("missing Shield service silently applied unreduced damage") }
			if !DefaultDamageWorld4E0B30(target, source, nil, 8, object.DamageClaw, r) || target.HealthData.Cur != 200 {
				t.Fatal("unsupported Shield result or HP")
			}
			want := []string{"missing Shield reduction service"}
			if !initial {
				want = []string{"buff-off", "sound", "unsupported live Shield reduction service"}
				if target.Obj130 != source || target.Frame134 != 1400 || target.Field131 != uint32(object.DamageClaw) || target.UpdateDataMonster().Field547 != 1 {
					t.Fatal("unsupported live Shield replayed or lost the executed prefix")
				}
			} else if target.Obj130 != nil || target.Frame134 != 0 || target.UpdateDataMonster().Field547 != 0 {
				t.Fatal("entry-time missing service ran the hit prefix")
			}
			if !slices.Equal(events, want) {
				t.Fatalf("missing Shield events=%v, want %v", events, want)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30LiveShieldResultAndFaultPrefix(t *testing.T) {
	for _, fault := range []bool{false, true} {
		t.Run(fmt.Sprintf("fault-%t", fault), func(t *testing.T) {
			target, source := damageMeleeUnitFixture4E17B0(t, false), damageMeleeUnitFixture4E17B0(t, true)
			r := damageMeleeWorldRuntime4E0B30(t)
			r.BuffOff = func(*Object, EnchantID) { target.Buffs |= 1 << defaultDamageShieldEnchant4E0B30 }
			calls := 0
			r.ShieldReduce = func(got *Object, damage *int32, typ object.DamageType, gotSource *Object) {
				calls++
				if got != target || gotSource != source || typ != object.DamageClaw || *damage != 8 {
					t.Fatal("live Shield arguments")
				}
				// No post-helper re-query: raw zero still returns false after
				// ShieldReduce itself removes the buff (e.g. depletion).
				target.Buffs &^= 1 << defaultDamageShieldEnchant4E0B30
				*damage = 0
				if fault {
					panic("Shield fault")
				}
			}
			r.DamageClear = func(*Object, int32) { t.Fatal("zero/fault Shield reached HP clear") }
			var recovered any
			result := true
			func() {
				defer func() { recovered = recover() }()
				result = DefaultDamageWorld4E0B30(target, source, nil, 8, object.DamageClaw, r)
			}()
			if calls != 1 || target.HealthData.Cur != 200 || target.Obj130 != source || target.Frame134 != 1400 || target.Field131 != uint32(object.DamageClaw) ||
				target.UpdateDataMonster().Field547 != 1 || target.HasEnchant(defaultDamageShieldEnchant4E0B30) {
				t.Fatal("Shield result/fault did not preserve the executed hit prefix")
			}
			if fault {
				if recovered != "Shield fault" {
					t.Fatalf("Shield fault=%v", recovered)
				}
			} else if recovered != nil || result {
				t.Fatalf("Shield zero result=%t panic=%v", result, recovered)
			}
		})
	}
}
