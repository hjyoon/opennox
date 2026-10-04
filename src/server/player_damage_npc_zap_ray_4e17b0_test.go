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

// GAME.EXE case 16 is raw signed damage: it skips armor absorption, carry,
// ordinary armor wear and electric protection. These are callback contracts,
// not stock Sentry placement or real HP evidence (the native test owns that).
func npcZapRayFixture4E17B0(t *testing.T, owner string) (target, source, ray *Object, cached *MonsterUpdateData, r PlayerDamageRuntime4E17B0) {
	t.Helper()
	target, source, ray = defaultDamageZapRayFixture4E0B30(t, owner, 0x11012)
	cached = target.UpdateDataMonster()
	cached.Field518, cached.WeaponEquipFlags = math.Float32bits(0.75), 0x400
	target.InvFirstItem = &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped,
		HealthData: &HealthData{Cur: 25, Max: 25}, InitData: unsafe.Pointer(new(ModifierInitData))}
	r = PlayerDamageRuntime4E17B0{
		Frame:               func() uint32 { return 1400 },
		BlockSourceExcluded: func(*Object) bool { return false },
		BlockDirection:      func(*Object, types.Pointf) bool { return false },
		GodMode:             func() bool { return false },
		QuestMode:           func() bool { return false },
		ItemArmorValue:      func(*Object) float32 { t.Fatal("NPC ZAP_RAY entered armor wear"); return 0 },
		CanDamageArmor:      func(*Object) bool { t.Fatal("NPC ZAP_RAY preflighted armor wear"); return false },
		DamageArmor: func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
			t.Fatal("NPC ZAP_RAY damaged armor")
			return false
		},
		ElectricArmorScale: func(*Object) float32 { t.Fatal("NPC ZAP_RAY is not electric case 9/17"); return 0 },
		FireProtection:     func(*Object) float64 { t.Fatal("NPC ZAP_RAY is not fire"); return 0 },
		ObserveClear:       func(*Object) { t.Fatal("NPC used the player observer layout") },
		PlayerSetState:     func(*Object, PlayerState) bool { t.Fatal("NPC used the player state layout"); return false },
		Unsupported:        func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) },
	}
	return
}

func TestPlayerDamageNPCZapRaySignedQuest4E17B0(t *testing.T) {
	for _, owner := range []string{"world", "Player", "NPC"} {
		for _, raw := range []int32{-7, 0, 1, 19, 500} {
			for _, scale := range []float32{-1, 0, 0.5, 1.25} {
				for _, god := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/raw-%d/scale-%g/god-%t", owner, raw, scale, god), func(t *testing.T) {
						target, source, ray, cached, r := npcZapRayFixture4E17B0(t, owner)
						before, armorBefore := *cached, *target.InvFirstItem
						var events []string
						marker, kind := uint32(1), uint32(ray.TypeInd)
						if owner == "world" {
							marker, kind = 2, 16
						}
						r.Frame = func() uint32 {
							if cached.Field547 != 99 {
								t.Fatal("NPC entry marker cleared before frame read")
							}
							events = append(events, "frame")
							return 1400
						}
						r.BlockSourceExcluded = func(got *Object) bool {
							if got != ray || cached.Field547 != 0 || cached.Field546 != 77 {
								t.Fatal("NPC prefix must precede exclusions")
							}
							events = append(events, "excluded")
							return false
						}
						r.BlockDirection = func(got *Object, pos types.Pointf) bool {
							want := uint32(1)
							if owner == "world" {
								want = 0
							}
							if got != target || pos != ray.PrevPos || cached.Field547 != want {
								t.Fatal("NPC facing/attribution order")
							}
							events = append(events, "direction")
							return false
						}
						r.GodMode = func() bool {
							if cached.Field547 != marker || cached.Field546 != kind {
								t.Fatal("NPC fallback marker must precede GodMode")
							}
							events = append(events, "god")
							return god // GodMode protects players, not NPCs.
						}
						r.QuestMode = func() bool { events = append(events, "quest"); return scale >= 0 }
						r.QuestDamageScale = func() float32 { events = append(events, "scale"); return scale }
						wantDamage := raw
						if scale >= 0 {
							wantDamage = int32(math.RoundToEven(float64(float32(float64(scale) * float64(raw)))))
							if raw > 0 && wantDamage < 1 {
								wantDamage = 1
							}
						}
						r.DefaultDamage = func(got, attacker, attack *Object, d int32, typ object.DamageType) bool {
							if got != target || attacker != source || attack != ray || d != wantDamage || typ != object.DamageZapRay {
								t.Fatalf("NPC ray default input=%d want=%d", d, wantDamage)
							}
							events = append(events, "default")
							return true
						}
						h, result := PlayerDamageNative4E17B0(target, source, ray, raw, object.DamageZapRay, r)
						want := []string{"frame", "excluded", "direction", "god", "quest"}
						if scale >= 0 {
							want = append(want, "scale")
						}
						want = append(want, "default")
						before.Field547, before.Field546 = marker, kind
						if !h || !result || !slices.Equal(events, want) || *cached != before || *target.InvFirstItem != armorBefore || target.HealthData.Cur != 2000 {
							t.Fatalf("NPC ray=%t/%t events=%v want=%v marker=%d/%d carry=%g", h, result, events, want, cached.Field547, cached.Field546, math.Float32frombits(cached.Field1))
						}
					})
				}
			}
		}
	}
}

func TestPlayerDamageNPCZapRayCachedLivePrefix4E17B0(t *testing.T) {
	for _, owner := range []string{"world", "Player", "NPC"} {
		for _, stage := range []string{"reflect", "excluded", "direction", "scale"} {
			t.Run(owner+"/"+stage, func(t *testing.T) {
				target, source, ray, cached, r := npcZapRayFixture4E17B0(t, owner)
				target.Buffs |= 1 << playerDamageReflectEnchant4E17B0
				live := &MonsterUpdateData{Field547: 31, Field546: 33, Field1: math.Float32bits(-0.5), Field523_2: 66}
				before, liveBefore := *cached, *live
				mutate := func(at string) {
					if stage == at {
						target.UpdateData = unsafe.Pointer(live)
					}
				}
				var events []string
				calls := 0
				r.BlockDirection = func(got *Object, pos types.Pointf) bool {
					if got != target {
						t.Fatal("NPC ray direction target")
					}
					calls++
					if calls == 1 {
						if pos != ray.PosVec || cached.Field547 != 0 || cached.Field546 != 77 {
							t.Fatal("Reflect must observe the cleared entry marker")
						}
						events = append(events, "reflect")
						mutate("reflect")
						ray.PrevPos = types.Ptf(30, 7)
					} else {
						want := uint32(1)
						if owner == "world" {
							want = 0
						}
						if pos != types.Ptf(30, 7) || cached.Field547 != want {
							t.Fatal("NPC ray must use post-Reflect, pre-exclusion PrevPos")
						}
						events = append(events, "direction")
						mutate("direction")
					}
					return false
				}
				r.BlockSourceExcluded = func(got *Object) bool {
					if got != ray || cached.Field547 != 0 {
						t.Fatal("NPC exclusion must precede distinct attribution")
					}
					events = append(events, "excluded")
					mutate("excluded")
					ray.PrevPos, ray.TypeInd = types.Ptf(-100, 99), 888
					return false
				}
				r.GodMode = func() bool { events = append(events, "god"); return false }
				r.QuestMode = func() bool { events = append(events, "quest"); return true }
				r.QuestDamageScale = func() float32 { events = append(events, "scale"); mutate("scale"); return 0.5 }
				r.DefaultDamage = func(got, attacker, attack *Object, d int32, typ object.DamageType) bool {
					if got != target || attacker != source || attack != ray || d != 10 || typ != object.DamageZapRay || target.UpdateDataMonster() != live || *live != liveBefore {
						t.Fatal("NPC ray default must retain the untouched replacement update")
					}
					events = append(events, "default")
					return false
				}
				h, result := PlayerDamageNative4E17B0(target, source, ray, 19, object.DamageZapRay, r)
				before.Field547, before.Field546 = 1, 888
				if owner == "world" {
					before.Field547, before.Field546 = 2, 16
				}
				want := []string{"reflect", "excluded", "direction", "god", "quest", "scale", "default"}
				if !h || result || !slices.Equal(events, want) || calls != 2 || *cached != before || *live != liveBefore {
					t.Fatalf("NPC cached/live=%t/%t events=%v", h, result, events)
				}
			})
		}
	}
}

func TestPlayerDamageNPCZapRayReflectEarlyReturn4E17B0(t *testing.T) {
	for _, owner := range []string{"world", "Player", "NPC"} {
		for _, raw := range []int32{-7, 0, 500} {
			t.Run(fmt.Sprintf("%s/raw-%d", owner, raw), func(t *testing.T) {
				target, source, ray, cached, r := npcZapRayFixture4E17B0(t, owner)
				target.Buffs |= 1 << playerDamageReflectEnchant4E17B0
				live := &MonsterUpdateData{Field547: 31, Field546: 33}
				before, liveBefore := *cached, *live
				var events []string
				r.BlockDirection = func(got *Object, pos types.Pointf) bool {
					if got != target || pos != ray.PosVec || cached.Field547 != 0 || cached.Field546 != 77 {
						t.Fatal("NPC Reflect entry/current position")
					}
					events = append(events, "direction")
					target.UpdateData, target.PosVec = unsafe.Pointer(live), types.Ptf(55, 66)
					ray.ObjClass |= object.ClassMissile // The original nonmissile branch was already selected.
					return true
				}
				r.PointFX = func(id int, pos types.Pointf) {
					if id != 132 || pos != types.Ptf(55, 66) {
						t.Fatal("NPC Reflect live point FX")
					}
					events = append(events, "fx")
				}
				r.Audio = func(id int, got *Object) {
					if id != 122 || got != target {
						t.Fatal("NPC Reflect audio")
					}
					events = append(events, "audio")
				}
				r.BlockSourceExcluded, r.DefaultDamage, r.QuestDamageScale = nil, nil, nil
				r.GodMode = func() bool { t.Fatal("Reflect queried the unused tail"); return false }
				r.QuestMode = func() bool { t.Fatal("Reflect queried Quest"); return true }
				h, result := PlayerDamageNative4E17B0(target, source, ray, raw, object.DamageZapRay, r)
				before.Field547 = 0
				if !h || result || !slices.Equal(events, []string{"direction", "fx", "audio"}) || *cached != before || *live != liveBefore || target.HealthData.Cur != 2000 {
					t.Fatalf("NPC ray Reflect=%t/%t events=%v", h, result, events)
				}
			})
		}
	}
}

func TestPlayerDamageNPCZapRayShieldCachedFlagsLiveItem4E17B0(t *testing.T) {
	for _, owner := range []string{"world", "Player", "NPC"} {
		for _, facing := range []string{"front", "rear", "excluded"} {
			for _, raw := range []int32{-7, 0, 19} {
				for _, broken := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/%s/raw-%d/broken-%t", owner, facing, raw, broken), func(t *testing.T) {
						target, source, ray, cached, r := npcZapRayFixture4E17B0(t, owner)
						cached.ArmorEquipFlags, cached.AIStackInd = 0x2000000, 1
						cached.AIStack[1].Action = uint32(ai.ACTION_BLOCK_ATTACK)
						first, afterAudio, afterBalance := &Object{ObjFlags: object.FlagEquipped, ObjSubClass: 2}, &Object{ObjFlags: object.FlagEquipped, ObjSubClass: 2}, &Object{ObjFlags: object.FlagEquipped, ObjSubClass: 2}
						target.InvFirstItem = first
						before := *cached
						var events []string
						r.BlockSourceExcluded = func(got *Object) bool {
							if got != ray || cached.Field547 != 0 {
								t.Fatal("shield prefix/exclusion identity")
							}
							events = append(events, "excluded")
							cached.ArmorEquipFlags, ray.TypeInd, ray.PrevPos = 0, 888, types.Ptf(-100, 99)
							return facing == "excluded"
						}
						r.BlockDirection = func(got *Object, pos types.Pointf) bool {
							if got != target || pos != types.Ptf(44, 7) {
								t.Fatal("NPC shield lost the saved PrevPos")
							}
							events = append(events, "direction")
							return facing == "front"
						}
						r.Audio = func(id int, got *Object) {
							if id != 878 || got != target {
								t.Fatal("NPC shield sound")
							}
							events = append(events, "audio")
							target.InvFirstItem = afterAudio
						}
						r.BlockDamagePercent = func() float64 { events = append(events, "percent"); target.InvFirstItem = afterBalance; return 0.25 }
						r.CanDamageBlockItem = func(item *Object) bool {
							if item != afterBalance {
								t.Fatal("NPC shield item must be selected after audio/balance")
							}
							events = append(events, "can-wear")
							return true
						}
						r.DamageBlockItem = func(item, owner, attacker, attack *Object, amount float32, typ object.DamageType) bool {
							if item != afterBalance || owner != target || attacker != source || attack != ray || amount != float32(float64(raw)*0.25) || typ != object.DamageZapRay {
								t.Fatal("NPC ray shield wear arguments")
							}
							events = append(events, "wear")
							if broken {
								item.ObjFlags |= object.FlagDestroyed
							}
							return true
						}
						if broken {
							r.Melee.MonsterPopBlockAction = func(got *Object) {
								if got != target || !afterBalance.Flags().Has(object.FlagDestroyed) {
									t.Fatal("NPC shield pop preceded break")
								}
								events = append(events, "pop")
								cached.AIStackInd--
							}
						}
						r.DefaultDamage = func(got, attacker, attack *Object, d int32, typ object.DamageType) bool {
							if got != target || attacker != source || attack != ray || d != raw || typ != object.DamageZapRay {
								t.Fatal("NPC unblocked shield ray input")
							}
							events = append(events, "default")
							return true
						}
						blocked := facing == "front"
						if blocked {
							r.DefaultDamage, r.QuestDamageScale = nil, nil
							r.GodMode = func() bool { t.Fatal("shield queried GodMode"); return false }
							r.QuestMode = func() bool { t.Fatal("shield queried Quest"); return true }
						}
						h, result := PlayerDamageNative4E17B0(target, source, ray, raw, object.DamageZapRay, r)
						want := []string{"excluded"}
						if facing != "excluded" {
							want = append(want, "direction")
						}
						if blocked {
							want = append(want, "audio", "percent", "can-wear", "wear")
							if broken {
								want = append(want, "pop")
								before.AIStackInd--
							}
						} else {
							want = append(want, "default")
						}
						before.ArmorEquipFlags, before.Field547, before.Field546 = 0, 1, 888
						if owner == "world" {
							before.Field547, before.Field546 = 0, 77
							if !blocked {
								before.Field547, before.Field546 = 2, 16
							}
						}
						if !h || result == blocked || !slices.Equal(events, want) || *cached != before || first.Flags().Has(object.FlagDestroyed) || afterAudio.Flags().Has(object.FlagDestroyed) || target.HealthData.Cur != 2000 {
							t.Fatalf("NPC ray shield=%t/%t events=%v want=%v", h, result, events, want)
						}
					})
				}
			}
		}
	}
}

func TestPlayerDamageNPCZapRayShieldLiveMissile4E17B0(t *testing.T) {
	for _, transition := range []string{"missile", "excluded-subclass", "clear-class", "retain-owner"} {
		t.Run(transition, func(t *testing.T) {
			target, source, ray, cached, r := npcZapRayFixture4E17B0(t, "Player")
			cached.ArmorEquipFlags = 0x1000000
			cached.AIStack[0].Action = uint32(ai.ACTION_BLOCK_ATTACK)
			target.InvFirstItem = nil // No shield record: wear is a successful no-op.
			var events []string
			r.BlockDirection = func(*Object, types.Pointf) bool { return true }
			r.Audio = func(id int, got *Object) {
				if id != 878 || got != target {
					t.Fatal("NPC live shield sound")
				}
				events = append(events, "audio")
				ray.ObjClass |= object.ClassMissile
				if transition == "excluded-subclass" {
					ray.ObjSubClass = 0x40
				}
			}
			r.ProjectileReflect = func(attack, got *Object) {
				if attack != ray || got != target {
					t.Fatal("NPC live reflection identity")
				}
				events = append(events, "reflect")
				if transition == "clear-class" {
					ray.ObjClass &^= object.ClassMissile
				}
				if transition == "retain-owner" {
					ray.ObjSubClass |= 2
				}
			}
			r.ClearOwner = func(attack *Object) {
				if attack != ray {
					t.Fatal("NPC live clear owner")
				}
				events = append(events, "clear")
			}
			r.SetOwner = func(owner, attack *Object) {
				if owner != target || attack != ray {
					t.Fatal("NPC live set owner")
				}
				events = append(events, "set")
			}
			r.BlockDamagePercent = func() float64 { events = append(events, "percent"); return 0.25 }
			r.DamageBlockItem = func(item, owner, attacker, attack *Object, amount float32, typ object.DamageType) bool {
				if item != nil || owner != target || attacker != source || attack != ray || amount != 4.75 || typ != object.DamageZapRay {
					t.Fatal("NPC nil-shield wear input")
				}
				events = append(events, "wear")
				return true
			}
			h, result := PlayerDamageNative4E17B0(target, source, ray, 19, object.DamageZapRay, r)
			want := []string{"audio"}
			if transition != "excluded-subclass" {
				want = append(want, "reflect")
			}
			if transition == "missile" {
				want = append(want, "clear", "set")
			}
			want = append(want, "percent", "wear")
			if !h || result || !slices.Equal(events, want) || cached.Field547 != 1 || cached.Field546 != 801 || target.HealthData.Cur != 2000 {
				t.Fatalf("NPC live missile shield=%t/%t events=%v want=%v", h, result, events, want)
			}
		})
	}
}

func TestPlayerDamageNPCZapRayMissingLiveService4E17B0(t *testing.T) {
	for _, service := range []string{"exclusions", "direction", "reflect-direction", "reflect-fx", "reflect-audio", "shield-audio", "shield-percent", "shield-can-wear", "shield-wear", "shield-pop", "quest-scale", "default", "live-exclusions", "live-direction", "live-god", "live-scale", "live-player"} {
		t.Run(service, func(t *testing.T) {
			target, source, ray, cached, r := npcZapRayFixture4E17B0(t, "Player")
			r.DefaultDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
				t.Fatal("missing live service reached DefaultDamage")
				return true
			}
			r.Audio = func(int, *Object) {}
			r.PointFX = func(int, types.Pointf) {}
			if len(service) >= 8 && service[:8] == "reflect-" {
				target.Buffs |= 1 << playerDamageReflectEnchant4E17B0
				r.BlockDirection = func(*Object, types.Pointf) bool { return true }
			}
			if len(service) >= 7 && service[:7] == "shield-" {
				cached.ArmorEquipFlags = 0x1000000
				cached.AIStack[0].Action = uint32(ai.ACTION_BLOCK_ATTACK)
				target.InvFirstItem = &Object{ObjFlags: object.FlagEquipped, ObjSubClass: 2}
				r.BlockDirection = func(*Object, types.Pointf) bool { return true }
				r.BlockDamagePercent = func() float64 { return 0.25 }
				r.CanDamageBlockItem = func(*Object) bool { return true }
				r.DamageBlockItem = func(item, _, _, _ *Object, _ float32, _ object.DamageType) bool {
					if service == "shield-pop" {
						item.ObjFlags |= object.FlagDestroyed
					}
					return true
				}
			}
			switch service {
			case "exclusions":
				r.BlockSourceExcluded = nil
			case "direction", "reflect-direction":
				r.BlockDirection = nil
			case "reflect-fx":
				r.PointFX = nil
			case "reflect-audio", "shield-audio":
				r.Audio = nil
			case "shield-percent":
				r.BlockDamagePercent = nil
			case "shield-can-wear":
				r.CanDamageBlockItem = nil
			case "shield-wear":
				r.DamageBlockItem = nil
			case "quest-scale":
				r.QuestMode = func() bool { return true }
				r.QuestDamageScale = nil
			case "default":
				r.DefaultDamage = nil
			case "live-exclusions":
				r.BlockSourceExcluded = func(*Object) bool { target.UpdateData = nil; return false }
			case "live-direction":
				r.BlockDirection = func(*Object, types.Pointf) bool { target.UpdateData = nil; return true }
			case "live-god":
				r.GodMode = func() bool { target.UpdateData = nil; return false }
			case "live-scale":
				r.QuestMode = func() bool { return true }
				r.QuestDamageScale = func() float32 { target.UpdateData = nil; return 0.5 }
			case "live-player":
				r.BlockSourceExcluded = func(*Object) bool { target.ObjClass = object.ClassPlayer; return false }
			}
			var reason string
			r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
			h, result := PlayerDamageNative4E17B0(target, source, ray, 19, object.DamageZapRay, r)
			if h || result || reason == "" || cached.Field547 == 99 || target.HealthData.Cur != 2000 || cached.Field1 != math.Float32bits(0.125) || cached.Field523_2 != 88 {
				t.Fatalf("NPC ray missing %s=%t/%t reason=%q marker=%d", service, h, result, reason, cached.Field547)
			}
		})
	}
}

func TestPlayerDamageNPCZapRayAdmission4E17B0(t *testing.T) {
	for _, gate := range []string{"no-update", "dead", "invulnerable", "non-NPC", "missing-update", "missing-source", "source-update", "source-wand", "ray-missile", "ray-wand", "ray-unit", "not-simple", "not-immobile"} {
		t.Run(gate, func(t *testing.T) {
			target, source, ray, cached, r := npcZapRayFixture4E17B0(t, "Player")
			r.DefaultDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
				t.Fatal("invalid NPC ray reached DefaultDamage")
				return true
			}
			before := *cached
			var reason string
			r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
			sounds := 0
			r.Audio = func(id int, got *Object) {
				if id != 71 || got != target {
					t.Fatal("NPC invulnerable sound")
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
			case "non-NPC":
				target.ObjSubClass &^= 0x10
			case "missing-update":
				target.UpdateData = nil
			case "missing-source":
				source = nil
			case "source-update":
				source.UpdateData = nil
			case "source-wand":
				source.ObjClass |= object.ClassWand
			case "ray-missile":
				ray.ObjClass |= object.ClassMissile
			case "ray-wand":
				ray.ObjClass |= object.ClassWand
			case "ray-unit":
				ray.ObjClass |= object.ClassMonster
			case "not-simple":
				ray.ObjClass &^= object.ClassSimple
			case "not-immobile":
				ray.ObjClass &^= object.ClassImmobile
			}
			h, result := PlayerDamageNative4E17B0(target, source, ray, 19, object.DamageZapRay, r)
			wantHandled, wantResult := gate == "no-update" || gate == "dead" || gate == "invulnerable", gate == "invulnerable"
			wantSounds := 0
			if wantResult {
				wantSounds = 1
			}
			if h != wantHandled || result != wantResult || sounds != wantSounds || (!wantHandled && reason == "") || *cached != before || target.HealthData.Cur != 2000 {
				t.Fatalf("NPC ray gate %s=%t/%t reason=%q sounds=%d", gate, h, result, reason, sounds)
			}
		})
	}
}
