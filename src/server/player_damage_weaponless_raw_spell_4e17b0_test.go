package server

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

func TestPlayerDamageNative4E17B0WeaponlessRawSpellPairs(t *testing.T) {
	for _, to := range []string{"player", "NPC"} {
		for _, from := range []string{"player", "monster", "NPC"} {
			if from == to {
				continue
			}
			for _, typ := range []object.DamageType{object.DamageManaBomb, object.DamageZapRay} {
				t.Run(fmt.Sprintf("%s-to-%s/type-%d", from, to, typ), func(t *testing.T) {
					target, source := damageRawSpellUnit4E17B0(t, to), damageRawSpellUnit4E17B0(t, from)
					armor := damageMeleeArmorFixture4E17B0(target, .5, .4)
					r := damageMeleeRuntimeFixture4E17B0(t)
					r.Unsupported = func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
						t.Errorf("raw spell prefix rejected: %s", reason)
					}
					r.ItemArmorValue = func(*Object) float32 { t.Fatal("raw spell read armor value"); return 1 }
					r.DamageArmor = func(*Object, *Object, *Object, int32, object.DamageType) bool {
						t.Fatal("raw spell wore armor")
						return true
					}
					r.ElectricArmorScale = func(*Object) float32 { t.Fatal("raw spell used electric armor"); return 1 }
					if h, result := PlayerDamageNative4E17B0(target, source, nil, 25, typ, r); !h || !result || target.HealthData.Cur != 175 || source.HealthData.Cur != 200 {
						t.Fatalf("handled/result/HP=%t/%t/%d", h, result, target.HealthData.Cur)
					}
					marker, kind, carry := damageMeleeMarker4E17B0(target)
					if marker != 2 || kind != uint32(typ) || carry != .4 || armor.HealthData.Cur != 25 || target.Obj130 != source || target.Pos132 != source.PrevPos {
						t.Fatalf("raw marker/carry/armor/source=%d/%d/%g/%d/%p", marker, kind, carry, armor.HealthData.Cur, target.Obj130)
					}
				})
			}
		}
	}
}

func TestPlayerDamageNative4E17B0WeaponlessRawSpellEarlyGates(t *testing.T) {
	for _, to := range []string{"player", "NPC"} {
		for _, typ := range []object.DamageType{object.DamageManaBomb, object.DamageZapRay} {
			for _, gate := range []string{"Dead", "NoUpdate", "Invulnerable-1400", "Invulnerable-1401", "player-disabled"} {
				if gate == "player-disabled" && to != "player" {
					continue
				}
				t.Run(fmt.Sprintf("%s/%s/type-%d", to, gate, typ), func(t *testing.T) {
					target, source := damageRawSpellUnit4E17B0(t, to), damageRawSpellUnit4E17B0(t, "monster")
					before := target.UpdateData
					r := PlayerDamageRuntime4E17B0{Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) }}
					wantResult, audio := false, 0
					switch gate {
					case "Dead":
						target.ObjFlags |= object.FlagDead
						target.UpdateData = nil
					case "NoUpdate":
						target.ObjFlags |= object.FlagNoUpdate
						target.UpdateData = nil
					case "player-disabled":
						target.UpdateDataPlayer().Player.Field3680 = 1
					default:
						target.Buffs = 1 << ENCHANT_INVULNERABLE
						target.UpdateData = nil // Buff gate is before update admission.
						wantResult = true
						r.Frame = func() uint32 {
							if gate == "Invulnerable-1400" {
								return 1400
							}
							return 1401
						}
						r.Audio = func(id int, got *Object) {
							if id != 71 || got != target {
								t.Fatal("wrong invulnerability audio")
							}
							audio++
						}
					}
					if h, result := PlayerDamageNative4E17B0(target, source, nil, 25, typ, r); !h || result != wantResult || target.HealthData.Cur != 200 || target.Obj130 != nil {
						t.Fatalf("early handled/result/HP=%t/%t/%d", h, result, target.HealthData.Cur)
					}
					wantAudio := 0
					if gate == "Invulnerable-1400" {
						wantAudio = 1
					}
					if audio != wantAudio {
						t.Fatalf("audio=%d want %d", audio, wantAudio)
					}
					target.UpdateData = before
				})
			}
		}
	}
}

func TestPlayerDamageNative4E17B0WeaponlessRawSpellSignedQuestAndGod(t *testing.T) {
	for _, to := range []string{"player", "NPC"} {
		for _, typ := range []object.DamageType{object.DamageManaBomb, object.DamageZapRay} {
			for _, god := range []bool{false, true} {
				for _, raw := range []int32{-3, 0, 1, 5, 9} {
					t.Run(fmt.Sprintf("%s/type-%d/god-%t/raw-%d", to, typ, god, raw), func(t *testing.T) {
						target := damageRawSpellUnit4E17B0(t, to)
						r := damageMeleeRuntimeFixture4E17B0(t)
						var events []string
						r.GodMode = func() bool {
							m, k, _ := damageMeleeMarker4E17B0(target)
							if m != 2 || k != uint32(typ) {
								t.Fatal("GodMode before raw marker")
							}
							events = append(events, "god")
							return god
						}
						r.QuestMode = func() bool { events = append(events, "quest"); return true }
						r.QuestDamageScale = func() float32 { events = append(events, "scale"); return .5 }
						r.BlockDirection = func(*Object, types.Pointf) bool { t.Fatal("nil source read facing"); return true }
						r.BlockSourceOnlyExcluded = func(*Object) bool { t.Fatal("nil source read exclusions"); return false }
						got, calls := int32(999), 0
						r.DefaultDamage = func(v, a, w *Object, d int32, k object.DamageType) bool {
							if v != target || a != nil || w != nil || k != typ {
								t.Fatal("raw tail lost context")
							}
							got, calls = d, calls+1
							events = append(events, "default")
							return true
						}
						if h, result := PlayerDamageNative4E17B0(target, nil, nil, raw, typ, r); !h || !result {
							t.Fatal("raw prefix rejected")
						}
						if god && to == "player" {
							if calls != 0 || !slices.Equal(events, []string{"god"}) {
								t.Fatalf("GodMode calls/events=%d/%v", calls, events)
							}
						} else {
							want := int32(math.RoundToEven(float64(float32(float64(raw) * .5))))
							if raw > 0 && want < 1 {
								want = 1
							}
							if calls != 1 || got != want || !slices.Equal(events, []string{"god", "quest", "scale", "default"}) {
								t.Fatalf("raw scaled=%d want %d events=%v", got, want, events)
							}
						}
					})
				}
			}
		}
	}
}

func TestPlayerDamageNative4E17B0WeaponlessRawSpellCoopSelf(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageManaBomb, object.DamageZapRay} {
		t.Run(fmt.Sprint(typ), func(t *testing.T) {
			target := damageRawSpellUnit4E17B0(t, "player")
			r := damageMeleeRuntimeFixture4E17B0(t)
			r.CoopMode = func() bool { return true }
			h, result := PlayerDamageNative4E17B0(target, target, nil, 9, typ, r)
			want := uint16(200)
			if typ == object.DamageManaBomb {
				want = 191
			}
			if !h || result != (typ == object.DamageManaBomb) || target.HealthData.Cur != want {
				t.Fatalf("Coop self result/HP=%t/%d", result, target.HealthData.Cur)
			}
		})
	}
}

func rawSpellBlockFixture4E17B0(t *testing.T, kind string) (target, source, shield *Object) {
	t.Helper()
	target, source = damageRawSpellUnit4E17B0(t, kind), damageRawSpellUnit4E17B0(t, "monster")
	source.PosVec, source.PrevPos = types.Ptf(20, 0), types.Ptf(-20, 7)
	if kind == "player" {
		target.UpdateDataPlayer().State, target.UpdateDataPlayer().Player.ArmorEquip = PlayerState16, 0x1000000
	} else {
		ud := target.UpdateDataMonster()
		ud.ArmorEquipFlags, ud.AIStack[0].Action = 0x1000000, uint32(ai.ACTION_BLOCK_ATTACK)
	}
	shield = &Object{ObjClass: object.ClassArmor, ObjSubClass: 2, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(&ModifierInitData{}), HealthData: &HealthData{Cur: 25, Max: 25}}
	target.InvFirstItem = shield
	return
}

func TestPlayerDamageNative4E17B0WeaponlessRawSpellReflectAndShield(t *testing.T) {
	for _, to := range []string{"player", "NPC"} {
		for _, reflect := range []bool{false, true} {
			for _, front := range []bool{false, true} {
				for _, typ := range []object.DamageType{object.DamageManaBomb, object.DamageZapRay} {
					t.Run(fmt.Sprintf("%s/reflect-%t/front-%t/type-%d", to, reflect, front, typ), func(t *testing.T) {
						target, source, shield := rawSpellBlockFixture4E17B0(t, to)
						if reflect {
							target.Buffs |= 1 << ENCHANT_REFLECTIVE_SHIELD
						}
						r := damageMeleeRuntimeFixture4E17B0(t)
						var events []string
						r.BlockSourceExcluded = func(*Object) bool { t.Fatal("raw nil weapon used six exclusions"); return true }
						r.BlockSourceOnlyExcluded = func(got *Object) bool {
							if got != source {
								t.Fatal("exclusion source")
							}
							events = append(events, "exclude")
							return false
						}
						r.BlockDirection = func(got *Object, pos types.Pointf) bool {
							if typ == object.DamageManaBomb {
								t.Fatal("Mana Bomb entered Reflect/direction")
							}
							if reflect && len(events) == 0 {
								if pos != source.PosVec {
									t.Fatal("Reflect used PrevPos")
								}
								events = append(events, "reflect-front")
							} else {
								if pos != source.PrevPos {
									t.Fatal("physical block used PosVec")
								}
								events = append(events, "shield-front")
							}
							return front
						}
						r.PointFX = func(id int, pos types.Pointf) {
							if id != 132 || pos != target.PosVec {
								t.Fatal("wrong ray FX")
							}
							events = append(events, "fx")
						}
						r.Audio = func(id int, got *Object) {
							if got != target {
								t.Fatal("wrong sound target")
							}
							events = append(events, fmt.Sprint(id))
						}
						r.BlockDamagePercent = func() float64 { events = append(events, "balance"); return .5 }
						r.CanDamageBlockItem = func(item *Object) bool { return item == shield }
						r.DamageBlockItem = func(item, owner, attacker, weapon *Object, amount float32, got object.DamageType) bool {
							if item != shield || owner != target || attacker != source || weapon != nil || amount != 12.5 || got != typ {
								t.Fatal("wrong shield wear context")
							}
							events = append(events, "wear")
							return true
						}
						r.BerserkShieldBlock = func(*Object) bool { t.Fatal("raw type entered berserker block"); return true }
						h, result := PlayerDamageNative4E17B0(target, source, nil, 25, typ, r)
						blocked := typ == object.DamageZapRay && front
						want := uint16(175)
						if blocked {
							want = 200
						}
						if !h || result == blocked || target.HealthData.Cur != want {
							t.Fatalf("raw result/HP=%t/%d events=%v", result, target.HealthData.Cur, events)
						}
						if blocked {
							wantEvents := []string{"exclude", "shield-front", "878", "balance", "wear"}
							if reflect {
								wantEvents = []string{"reflect-front", "fx", "122"}
							}
							if !slices.Equal(events, wantEvents) {
								t.Fatalf("block order=%v want %v", events, wantEvents)
							}
							m, _, _ := damageMeleeMarker4E17B0(target)
							if m != 0 || target.Obj130 != nil {
								t.Fatal("blocked raw damage entered tail")
							}
						}
					})
				}
			}
		}
	}
}

func TestPlayerDamageNative4E17B0WeaponlessRawSpellCachedAndLive(t *testing.T) {
	for _, to := range []string{"player", "NPC"} {
		for _, broken := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/broken-%t", to, broken), func(t *testing.T) {
				target, source, shield := rawSpellBlockFixture4E17B0(t, to)
				cached := target.UpdateData
				r := damageMeleeRuntimeFixture4E17B0(t)
				var events []string
				if to == "player" {
					old := target.UpdateDataPlayer()
					old.Player.Field3680, old.Player.CameraFollowObj = 2, source
					r.ObserveClear = func(*Object) {
						if old.Field76 != 0 {
							t.Fatal("observer preceded marker clear")
						}
						events = append(events, "observe")
						target.UpdateData = unsafe.Pointer(&PlayerUpdateData{Player: &Player{}, State: PlayerState13, Field76: 99})
					}
				}
				pos := source.PrevPos
				r.BlockSourceOnlyExcluded = func(*Object) bool {
					events = append(events, "exclude")
					source.PrevPos = types.Ptf(77, 99)
					if to == "NPC" {
						target.UpdateData = unsafe.Pointer(&MonsterUpdateData{Field547: 99, AIStack: [24]AIStackItem{{Action: uint32(ai.ACTION_BLOCK_ATTACK)}}})
					}
					return false
				}
				r.BlockDirection = func(_ *Object, got types.Pointf) bool {
					if got != pos {
						t.Fatal("source position snapshot moved after exclusion")
					}
					events = append(events, "front")
					return true
				}
				liveShield := &Object{ObjClass: object.ClassArmor, ObjSubClass: 2, ObjFlags: object.FlagEquipped}
				r.Audio = func(id int, _ *Object) {
					if id != 878 {
						t.Fatal("wrong shield sound")
					}
					events = append(events, "audio")
					target.InvFirstItem = liveShield
				}
				r.BlockDamagePercent = func() float64 { events = append(events, "balance"); return .25 }
				r.CanDamageBlockItem = func(item *Object) bool { return item == liveShield }
				r.DamageBlockItem = func(item, _, _, w *Object, amount float32, _ object.DamageType) bool {
					if item != liveShield || item == shield || w != nil || amount != 2 {
						t.Fatal("stale shield/nonnil weapon")
					}
					events = append(events, "wear")
					if broken {
						item.ObjFlags |= object.FlagDestroyed
					}
					return true
				}
				r.PlayerSetState = func(v *Object, state PlayerState) bool {
					if v != target || state != PlayerState13 || v.UpdateData == cached {
						t.Fatal("broken shield used cached state")
					}
					events = append(events, "state")
					return true
				}
				r.Melee.MonsterPopBlockAction = func(v *Object) {
					if v != target || v.UpdateData == cached {
						t.Fatal("broken shield used cached action")
					}
					events = append(events, "pop")
				}
				if h, result := PlayerDamageNative4E17B0(target, source, nil, 8, object.DamageZapRay, r); !h || result || target.HealthData.Cur != 200 {
					t.Fatal("cached equipment/state block lost")
				}
				want := []string{"exclude", "front", "audio", "balance", "wear"}
				if to == "player" {
					want = append([]string{"observe"}, want...)
				}
				if broken {
					if to == "player" {
						want = append(want, "state")
					} else {
						want = append(want, "pop")
					}
				}
				if !slices.Equal(events, want) {
					t.Fatalf("cached/live events=%v want %v", events, want)
				}
			})
		}
	}
}

func TestPlayerDamageNative4E17B0WeaponlessRawSpellLiveReflection(t *testing.T) {
	for _, to := range []string{"player", "NPC"} {
		for _, sub := range []object.SubClass{0, 0x40, 2} {
			t.Run(fmt.Sprintf("%s/sub-%x", to, sub), func(t *testing.T) {
				target, source, _ := rawSpellBlockFixture4E17B0(t, to)
				target.Buffs = 1 << ENCHANT_REFLECTIVE_SHIELD
				r := damageMeleeRuntimeFixture4E17B0(t)
				var events []string
				r.CoopMode = func() bool { source.ObjClass = object.ClassMissile; return false }
				r.BlockDirection = func(*Object, types.Pointf) bool { events = append(events, "front"); return true }
				r.ProjectileReflect = func(a, v *Object) {
					if a != source || v != target {
						t.Fatal("wrong reflection")
					}
					events = append(events, "reflect")
					source.ObjSubClass = sub
				}
				r.ClearOwner = func(*Object) { events = append(events, "clear") }
				r.SetOwner = func(v, a *Object) {
					if v != target || a != source {
						t.Fatal("wrong owner transfer")
					}
					events = append(events, "owner")
				}
				r.ChangeOwner = func(*Object, *Object) { events = append(events, "change") }
				r.PointFX = func(id int, _ types.Pointf) {
					if id != 132 {
						t.Fatal("ray FX")
					}
					events = append(events, "fx")
				}
				r.Audio = func(id int, _ *Object) {
					if id != 122 {
						t.Fatal("ray audio")
					}
					events = append(events, "audio")
				}
				if h, result := PlayerDamageNative4E17B0(target, source, nil, 8, object.DamageZapRay, r); !h || result || target.HealthData.Cur != 200 {
					t.Fatal("live missile Reflect lost")
				}
				want := []string{"front", "reflect"}
				if sub&0x40 == 0 {
					want = append(want, "clear", "owner")
				}
				if sub&2 != 0 {
					want = append(want, "change")
				}
				want = append(want, "fx", "audio")
				if !slices.Equal(events, want) {
					t.Fatalf("live reflection=%v want %v", events, want)
				}
			})
		}
	}
}
