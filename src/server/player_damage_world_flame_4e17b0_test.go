package server

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func worldFlameArmorFixture4E17B0(t *testing.T, target *Object) *Object {
	t.Helper()
	item, freeItem := alloc.New(Object{})
	health, freeHealth := alloc.New(HealthData{})
	update, freeUpdate := alloc.New(WeaponArmorUpdateData{})
	init, freeInit := alloc.New(ModifierInitData{})
	callback, freeCallback := alloc.New(byte(0))
	for _, free := range []func(){freeItem, freeHealth, freeUpdate, freeInit, freeCallback} {
		t.Cleanup(free)
	}
	*health = HealthData{Cur: 50, Max: 50}
	*update = WeaponArmorUpdateData{Field0: math.Float32bits(.25)}
	*init = ModifierInitData{}
	*item = Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, HealthData: health, UpdateData: unsafe.Pointer(update), InitData: unsafe.Pointer(init), Damage: unsafe.Pointer(callback)}
	target.InvFirstItem = item
	if target.Class().Has(object.ClassPlayer) {
		target.UpdateDataPlayer().Field57 = math.Float32bits(1)
	} else {
		target.UpdateDataMonster().Field518 = math.Float32bits(1)
	}
	return item
}

func TestPlayerDamageWorldFlame4E17B0SignedArmorQuest(t *testing.T) {
	for _, victim := range []string{"NPC", "player"} {
		for _, owner := range []string{"nil", "self", "imaginary", "player", "NPC", "proxy-NPC"} {
			for _, raw := range []int32{-7, 0, 3, 10, 19} {
				for _, scale := range []float32{-1, 0, .5, 1.25} {
					t.Run(fmt.Sprintf("%s/%s/raw-%d/scale-%g", victim, owner, raw, scale), func(t *testing.T) {
						v, a, w := worldFlameFixture4E17B0(t, owner, victim)
						armor := worldFlameArmorFixture4E17B0(t, v)
						v.Buffs |= 1 << 27 // Pure world FLAME does not enter Reflect.
						var marker, kind *uint32
						if victim == "player" {
							ud := v.UpdateDataPlayer()
							marker, kind = &ud.Field76, &ud.Field75
						} else {
							ud := v.UpdateDataMonster()
							marker, kind = &ud.Field547, &ud.Field546
						}
						wantMarker, wantKind := uint32(1), uint32(w.TypeInd)
						if a == nil || a == w {
							wantMarker, wantKind = 2, 1
						}
						var events []string
						r := PlayerDamageRuntime4E17B0{
							Frame: func() uint32 { t.Fatal("ordinary fire read invulnerability frame"); return 0 },
							BlockSourceExcluded: func(got *Object) bool {
								if got != w || a == nil || *marker != 0 {
									t.Fatal("exclusion prefix")
								}
								events = append(events, "exclude")
								return false
							},
							BlockDirection: func(got *Object, pos types.Pointf) bool {
								if got != v || pos != w.PrevPos {
									t.Fatal("facing identity")
								}
								events = append(events, "face")
								return true
							},
							ItemArmorValue: func(got *Object) float32 {
								if got != armor {
									t.Fatal("armor lookup")
								}
								events = append(events, "armor")
								return 1
							},
							CanDamageArmor: func(got *Object) bool { return got == armor },
							DamageArmor: func(got, attacker, effective *Object, amount int32, typ object.DamageType) bool {
								if got != armor || attacker != a || effective != w || amount != raw || typ != object.DamageFlame {
									t.Fatalf("FLAME armor amount=%d want=%d", amount, raw)
								}
								events = append(events, "wear")
								armor.HealthData.Cur -= uint16(amount)
								return true
							},
							GodMode: func() bool {
								if *marker != wantMarker || *kind != wantKind {
									t.Fatal("GodMode preceded marker")
								}
								events = append(events, "god")
								return victim == "NPC"
							},
							QuestMode:        func() bool { events = append(events, "quest"); return scale >= 0 },
							QuestDamageScale: func() float32 { events = append(events, "scale"); return scale },
							DefaultDamage: func(got, attacker, effective *Object, amount int32, typ object.DamageType) bool {
								want := raw
								if scale >= 0 {
									want = int32(math.RoundToEven(float64(float32(float64(raw) * float64(scale)))))
									if raw > 0 && want < 1 {
										want = 1
									}
								}
								if got != v || attacker != a || effective != w || typ != object.DamageFlame || amount != want {
									t.Fatalf("default amount=%d want=%d", amount, want)
								}
								events = append(events, "default")
								return true
							},
							Audio:       func(int, *Object) { t.Fatal("world fire reflected/blocked") },
							Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) },
						}
						h, result := PlayerDamageNative4E17B0(v, a, w, raw, object.DamageFlame, r)
						var want []string
						if a != nil {
							want = append(want, "exclude", "face")
						}
						want = append(want, "armor")
						if raw > 0 {
							want = append(want, "wear")
						}
						want = append(want, "god", "quest")
						if scale >= 0 {
							want = append(want, "scale")
						}
						want = append(want, "default")
						if !h || !result || !slices.Equal(events, want) || v.HealthData.Cur != 20 || armor.HealthData.Cur != uint16(50-max(raw, 0)) || armor.UpdateDataWeaponArmor().Field0 != math.Float32bits(.25) || *marker != wantMarker || *kind != wantKind {
							t.Fatalf("result=%t/%t events=%v want=%v", h, result, events, want)
						}
						if victim == "player" && v.UpdateDataPlayer().Field21 != math.Float32bits(.25) || victim == "NPC" && v.UpdateDataMonster().Field1 != math.Float32bits(.25) {
							t.Fatal("FLAME consumed HP fractional carry")
						}
					})
				}
			}
		}
	}
}

func TestPlayerDamageWorldFlame4E17B0ShieldAndSourceLess(t *testing.T) {
	for _, victim := range []string{"NPC", "player"} {
		for _, owner := range []string{"nil", "self", "NPC"} {
			for _, front := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/front-%t", victim, owner, front), func(t *testing.T) {
					v, a, w := worldFlameFixture4E17B0(t, owner, victim)
					if victim == "player" {
						ud := v.UpdateDataPlayer()
						ud.Player.ArmorEquip = 0x3000000
						ud.Player.WeaponEquip = 0x400
						ud.State = PlayerState16
					} else {
						ud := v.UpdateDataMonster()
						ud.ArmorEquipFlags = 0x3000000
						ud.WeaponEquipFlags = 0x400
						ud.AIStack[0].Action = 21
					}
					blocked := a != nil && front
					var events []string
					r := PlayerDamageRuntime4E17B0{
						BlockSourceExcluded: func(*Object) bool { events = append(events, "exclude"); return false },
						BlockDirection:      func(*Object, types.Pointf) bool { events = append(events, "face"); return front },
						Audio: func(id int, got *Object) {
							if !blocked || id != 878 || got != v {
								t.Fatal("shield audio")
							}
							events = append(events, "audio")
						},
						BlockDamagePercent: func() float64 { events = append(events, "balance"); return .5 },
						DamageBlockItem: func(item, got, attacker, effective *Object, amount float32, typ object.DamageType) bool {
							if item != nil || got != v || attacker != a || effective != w || amount != 4.5 || typ != object.DamageFlame {
								t.Fatal("shield wear")
							}
							events = append(events, "shield")
							return true
						},
						DefaultDamage: func(*Object, *Object, *Object, int32, object.DamageType) bool {
							events = append(events, "default")
							return true
						},
						Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) },
					}
					h, result := PlayerDamageNative4E17B0(v, a, w, 9, object.DamageFlame, r)
					var want []string
					if a != nil {
						want = append(want, "exclude", "face")
					}
					if blocked {
						want = append(want, "audio", "balance", "shield")
					} else {
						want = append(want, "default")
					}
					if !h || result == blocked || !slices.Equal(events, want) {
						t.Fatalf("result=%t/%t events=%v want=%v", h, result, events, want)
					}
				})
			}
		}
	}
}

func TestPlayerDamageWorldFlame4E17B0EarlyGates(t *testing.T) {
	for _, victim := range []string{"NPC", "player"} {
		for _, gate := range []string{"no-update", "dead", "invulnerable", "coop", "observer-disabled", "god"} {
			if victim == "NPC" && (gate == "coop" || gate == "observer-disabled") {
				continue
			}
			t.Run(victim+"/"+gate, func(t *testing.T) {
				v, a, w := worldFlameFixture4E17B0(t, "NPC", victim)
				switch gate {
				case "no-update":
					v.ObjFlags |= object.FlagNoUpdate
					v.Buffs |= 1 << 23
				case "dead":
					v.ObjFlags |= object.FlagDead
				case "invulnerable":
					v.Buffs |= 1 << 23
				case "coop":
					a = v
				case "observer-disabled":
					v.UpdateDataPlayer().Player.Field3680 |= 1
				}
				calls, audio := 0, 0
				r := PlayerDamageRuntime4E17B0{
					Frame: func() uint32 { return 1400 }, Audio: func(id int, got *Object) {
						if id != 71 || got != v {
							t.Fatal("invulnerability voice")
						}
						audio++
					},
					CoopMode: func() bool { return gate == "coop" }, BlockSourceExcluded: func(*Object) bool { return true },
					GodMode:       func() bool { return gate == "god" },
					DefaultDamage: func(*Object, *Object, *Object, int32, object.DamageType) bool { calls++; return true },
					Unsupported:   func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) },
				}
				h, result := PlayerDamageNative4E17B0(v, a, w, 9, object.DamageFlame, r)
				wantCalls := 0
				if victim == "NPC" && gate == "god" {
					wantCalls = 1
				}
				wantAudio := 0
				if gate == "invulnerable" {
					wantAudio = 1
				}
				wantResult := gate == "invulnerable" || gate == "god"
				if !h || result != wantResult || calls != wantCalls || audio != wantAudio {
					t.Fatalf("result=%t/%t default=%d audio=%d", h, result, calls, audio)
				}
			})
		}
	}
}

func TestPlayerDamageWorldFlame4E17B0CachedPrefixLiveArmor(t *testing.T) {
	for _, victim := range []string{"NPC", "player"} {
		t.Run(victim, func(t *testing.T) {
			v, a, w := worldFlameFixture4E17B0(t, "NPC", victim)
			armor := worldFlameArmorFixture4E17B0(t, v)
			pos := w.PrevPos
			var marker, kind, liveMarker, liveArmor *uint32
			var cachedArmor *uint32
			var live unsafe.Pointer
			if victim == "player" {
				cached := v.UpdateDataPlayer()
				ud, free := alloc.New(PlayerUpdateData{})
				t.Cleanup(free)
				*ud = PlayerUpdateData{Player: cached.Player, State: PlayerState16, Field76: 79, Field57: math.Float32bits(2)}
				marker, kind, cachedArmor = &cached.Field76, &cached.Field75, &cached.Player.ArmorEquip
				liveMarker, liveArmor, live = &ud.Field76, &ud.Field57, unsafe.Pointer(ud)
			} else {
				cached := v.UpdateDataMonster()
				ud, free := alloc.New(MonsterUpdateData{})
				t.Cleanup(free)
				*ud = MonsterUpdateData{Field547: 79, Field518: math.Float32bits(2)}
				ud.AIStack[0].Action = 21
				marker, kind, cachedArmor = &cached.Field547, &cached.Field546, &cached.ArmorEquipFlags
				liveMarker, liveArmor, live = &ud.Field547, &ud.Field518, unsafe.Pointer(ud)
			}
			var events []string
			r := PlayerDamageRuntime4E17B0{
				BlockSourceExcluded: func(*Object) bool {
					if *marker != 0 {
						t.Fatal("entry marker was not cleared")
					}
					events = append(events, "exclude")
					v.UpdateData, w.TypeInd = live, 222
					w.PrevPos = types.Pointf{X: 777, Y: 888}
					*cachedArmor = 0x3000000 // Entry equipment must not be reloaded.
					return false
				},
				BlockDirection: func(got *Object, attackPos types.Pointf) bool {
					if got != v || attackPos != pos || *marker != 1 || *kind != 222 {
						t.Fatal("cached PrevPos / live type ordering")
					}
					events = append(events, "face")
					return true
				},
				ItemArmorValue: func(got *Object) float32 {
					if got != armor {
						t.Fatal("live armor identity")
					}
					events = append(events, "armor")
					*liveArmor = math.Float32bits(4) // Denominator was already captured.
					w.TypeInd = 333
					return 1
				},
				CanDamageArmor: func(got *Object) bool { return got == armor },
				DamageArmor: func(got, attacker, effective *Object, amount int32, typ object.DamageType) bool {
					if got != armor || attacker != a || effective != w || amount != 5 || typ != object.DamageFlame {
						t.Fatalf("live armor amount=%d want=5", amount)
					}
					events = append(events, "wear")
					return true
				},
				GodMode: func() bool {
					if *marker != 1 || *kind != 222 || *liveMarker != 79 {
						t.Fatal("armor replaced the entry marker/type")
					}
					events = append(events, "god")
					return false
				},
				DefaultDamage: func(got, attacker, effective *Object, amount int32, typ object.DamageType) bool {
					if got != v || attacker != a || effective != w || amount != 10 || typ != object.DamageFlame || v.UpdateData != live {
						t.Fatal("default lost live record or full HP damage")
					}
					events = append(events, "default")
					return true
				},
				Audio:       func(int, *Object) { t.Fatal("live equipment replaced cached equipment") },
				Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) },
			}
			h, result := PlayerDamageNative4E17B0(v, a, w, 10, object.DamageFlame, r)
			if !h || !result || !slices.Equal(events, []string{"exclude", "face", "armor", "wear", "god", "default"}) {
				t.Fatalf("result=%t/%t events=%v", h, result, events)
			}
		})
	}
}

func TestPlayerDamageWorldFlame4E17B0ShieldLateRecords(t *testing.T) {
	for _, victim := range []string{"NPC", "player"} {
		for _, afterReflection := range []string{"missile", "not-missile", "owner-subclass"} {
			t.Run(victim+"/"+afterReflection, func(t *testing.T) {
				v, a, w := worldFlameFixture4E17B0(t, "NPC", victim)
				shield := worldFlameArmorFixture4E17B0(t, v)
				shield.ObjSubClass = 2
				v.InvFirstItem = nil // Shield is installed by the late balance callback.
				if victim == "player" {
					ud := v.UpdateDataPlayer()
					ud.Player.ArmorEquip, ud.State = 0x3000000, PlayerState16
				} else {
					ud := v.UpdateDataMonster()
					ud.ArmorEquipFlags, ud.AIStack[0].Action = 0x3000000, 21
				}
				var events []string
				r := PlayerDamageRuntime4E17B0{
					BlockSourceExcluded: func(*Object) bool { return false },
					BlockDirection:      func(*Object, types.Pointf) bool { return true },
					Audio: func(id int, got *Object) {
						if id != 878 || got != v {
							t.Fatal("shield audio")
						}
						events = append(events, "audio")
						w.ObjClass |= object.ClassMissile
						w.ObjSubClass = 0
					},
					ProjectileReflect: func(got, target *Object) {
						if got != w || target != v {
							t.Fatal("reflection identity")
						}
						events = append(events, "reflect")
						switch afterReflection {
						case "not-missile":
							w.ObjClass &^= object.ClassMissile
						case "owner-subclass":
							w.ObjSubClass = 2
						}
					},
					ClearOwner: func(got *Object) {
						if got != w {
							t.Fatal("clear identity")
						}
						events = append(events, "clear")
					},
					SetOwner: func(target, got *Object) {
						if got != w || target != v {
							t.Fatal("owner identity")
						}
						events = append(events, "owner")
					},
					BlockDamagePercent: func() float64 {
						events = append(events, "balance")
						v.InvFirstItem = shield
						return .5
					},
					CanDamageBlockItem: func(got *Object) bool { return got == shield },
					DamageBlockItem: func(item, got, attacker, effective *Object, amount float32, typ object.DamageType) bool {
						if item != shield || got != v || attacker != a || effective != w || amount != -1.5 || typ != object.DamageFlame {
							t.Fatal("late live shield durability")
						}
						events = append(events, "wear")
						shield.ObjFlags |= object.FlagDestroyed
						return true
					},
					PlayerSetState: func(got *Object, state PlayerState) bool {
						if victim != "player" || got != v || state != PlayerState13 {
							t.Fatal("broken-shield player state")
						}
						events = append(events, "broken")
						return true
					},
					Melee: PlayerDamageMeleeRuntime4E17B0{MonsterPopBlockAction: func(got *Object) {
						if victim != "NPC" || got != v {
							t.Fatal("broken-shield NPC action")
						}
						events = append(events, "broken")
					}},
					DefaultDamage: func(*Object, *Object, *Object, int32, object.DamageType) bool {
						t.Fatal("shield entered HP damage")
						return false
					},
					Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) },
				}
				h, result := PlayerDamageNative4E17B0(v, a, w, -3, object.DamageFlame, r)
				want := []string{"audio", "reflect"}
				if afterReflection == "missile" {
					want = append(want, "clear", "owner")
				}
				want = append(want, "balance", "wear", "broken")
				if !h || result || !slices.Equal(events, want) {
					t.Fatalf("result=%t/%t events=%v want=%v", h, result, events, want)
				}
			})
		}
	}
}
