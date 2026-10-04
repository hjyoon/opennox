package server

import (
	"fmt"
	"image"
	"math"
	"slices"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// This is the DefaultDamage dependency, not PlayerDamage's Reflect/physical
// block prefix. The original default tail must not clear Player hit markers,
// apply armor absorption/wear, or turn raw zero into the electric minimum.
func defaultDamagePlayerZapRayFixture4E0B30(t *testing.T, owner string) (target, source, ray *Object) {
	t.Helper()
	_, source, ray = defaultDamageZapRayFixture4E0B30(t, owner, 0x202)
	target = damageMeleeUnitFixture4E17B0(t, true)
	target.HealthData = &HealthData{Cur: 2000, Field2: 2000, Max: 2000}
	ud := target.UpdateDataPlayer()
	ud.Field75, ud.Field76 = 77, 88
	ud.Field40_0, ud.Field40_1 = 0x1122, 0xabcd
	ud.Field57, ud.Field21 = math.Float32bits(0.75), math.Float32bits(0.125)
	ud.Player.ArmorEquip, ud.Player.WeaponEquip = 0x405, 0x400
	return
}

func TestDefaultDamageWorld4E0B30PlayerZapRaySignedTail(t *testing.T) {
	for _, owner := range []string{"world", "Player", "NPC"} {
		for _, state := range []PlayerState{PlayerState13, PlayerState1, PlayerState15} {
			for _, raw := range []int32{-7, 0, 1, 19, 20, 500} {
				for _, playerSound := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/state-%d/raw-%d/player-sound-%t", owner, state, raw, playerSound), func(t *testing.T) {
						target, source, ray := defaultDamagePlayerZapRayFixture4E0B30(t, owner)
						ud := target.UpdateDataPlayer()
						ud.State = state
						target.Buffs = 1 << defaultDamageShockEnchant4E0B30
						before, beforePlayer := *ud, *ud.Player
						var events []string
						calls := 0
						r := damageMeleeWorldRuntime4E0B30(t)
						r.GameplayFlag1 = func() bool { events = append(events, "gameplay"); return true }
						r.IsEnemy = func(victim, attacker *Object) bool {
							if victim != target || attacker != source {
								t.Fatal("ray enemy identity")
							}
							events = append(events, "enemy")
							return false // A non-melee ray has no second friendly-hit gate.
						}
						r.BuffOff = func(victim *Object, id EnchantID) {
							if victim != target || id != 0 || victim.Pos132 != ray.PrevPos || victim.Obj130 != nil || *ud != before {
								t.Fatal("default ray BuffOff/marker order")
							}
							events = append(events, "buff")
						}
						r.Frame = func() uint32 { events = append(events, "frame"); return 1400 }
						r.MonsterHasHitSound = func(attacker *Object) bool {
							if owner != "NPC" || attacker != source {
								t.Fatal("NPC ray hit sound identity")
							}
							events = append(events, "hit-sound")
							return false
						}
						play := func(victim, attacker *Object) {
							if victim != target || attacker != ray || victim.Obj130 != ray || victim.Field131 != 16 || victim.Frame134 != 1400 {
								t.Fatal("default ray sound/metadata order")
							}
							events = append(events, "sound")
						}
						r.DefaultDamageSound = play
						if playerSound {
							r.PlayerDamageSoundC = unsafe.Pointer(new(byte))
							target.DamageSound = r.PlayerDamageSoundC
							r.PlayerDamageSound = play
							r.DefaultDamageSound = func(*Object, *Object) { t.Fatal("player sound used fallback") }
						}
						r.GameBallOnDamage = func(attacker, victim *Object, d int32) {
							if attacker != source || victim != target || d != raw {
								t.Fatal("ray ball tail signed input")
							}
							events = append(events, "ball")
						}
						r.PlayerSetState = nil // Unused services are not required below 20 or in states 1/15.
						wantState := state
						if raw >= 20 && state != PlayerState1 && state != PlayerState15 {
							wantState = PlayerState30
							r.PlayerSetState = func(victim *Object, st PlayerState) bool {
								if victim != target || st != PlayerState30 || ud.State != state {
									t.Fatal("ray hurt identity/order")
								}
								events = append(events, "hurt")
								ud.State = st
								return true
							}
						}
						r.FireProtection = func(*Object) float64 { t.Fatal("ray used fire resistance"); return 1 }
						r.ElectricProtection = func(*Object) float64 { t.Fatal("type 16 used electric resistance"); return 1 }
						r.CallDamage = func(*Object, *Object, *Object, int32, object.DamageType) bool {
							t.Fatal("ray retaliated Shock")
							return false
						}
						r.AdjustFieldGuide = func(*Object, *Object, int32) int32 { t.Fatal("Player ray used monster guide"); return 0 }
						r.DamageClear = func(victim *Object, d int32) {
							if victim != target || d != raw || ud.State != wantState {
								t.Fatal("ray raw HP/state input")
							}
							calls++
							events = append(events, "HP")
						}
						if !DefaultDamageWorld4E0B30(target, source, ray, raw, object.DamageZapRay, r) || calls != 1 {
							t.Fatal("Player ray did not reach HP exactly once")
						}
						want := []string{"gameplay", "enemy", "buff", "frame"}
						if owner == "NPC" {
							want = append(want, "hit-sound")
						}
						want = append(want, "sound", "ball")
						if wantState == PlayerState30 {
							want = append(want, "hurt")
						}
						if owner == "NPC" {
							want = append(want, "enemy")
						}
						want = append(want, "HP")
						before.State = wantState
						if !slices.Equal(events, want) || *ud != before || *ud.Player != beforePlayer || target.HealthData.Cur != 2000 || !target.HasEnchant(defaultDamageShockEnchant4E0B30) {
							t.Fatalf("Player ray events=%v want=%v markers=%d/%d state=%d", events, want, ud.Field76, ud.Field75, ud.State)
						}
					})
				}
			}
		}
	}
}

func TestDefaultDamageWorld4E0B30PlayerZapRayOwnerGate(t *testing.T) {
	for _, owner := range []string{"world", "Player", "NPC"} {
		for _, gameplay := range []bool{false, true} {
			for _, enemy := range []bool{false, true} {
				for _, quest := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/gameplay-%t/enemy-%t/quest-%t", owner, gameplay, enemy, quest), func(t *testing.T) {
						target, source, ray := defaultDamagePlayerZapRayFixture4E0B30(t, owner)
						r := damageMeleeWorldRuntime4E0B30(t)
						r.GameplayFlag1 = func() bool { return gameplay }
						r.IsEnemy = func(*Object, *Object) bool { return enemy }
						r.QuestMode = func() bool { return quest }
						calls := 0
						r.DamageClear = func(*Object, int32) { calls++ }
						DefaultDamageWorld4E0B30(target, source, ray, 19, object.DamageZapRay, r)
						admitted := gameplay || owner == "world" || enemy
						if (calls == 1) != admitted || calls > 1 || (target.Obj130 == ray) != admitted {
							t.Fatalf("owner gate calls=%d admitted=%t", calls, admitted)
						}
					})
				}
			}
		}
	}
	for _, quest := range []bool{false, true} {
		t.Run(fmt.Sprintf("self/quest-%t", quest), func(t *testing.T) {
			target, _, ray := defaultDamagePlayerZapRayFixture4E0B30(t, "world")
			ray.ObjOwner = target
			r := damageMeleeWorldRuntime4E0B30(t)
			r.GameplayFlag1 = func() bool { return false }
			r.IsEnemy = func(*Object, *Object) bool { return false }
			r.QuestMode = func() bool { return quest }
			calls := 0
			r.DamageClear = func(*Object, int32) { calls++ }
			DefaultDamageWorld4E0B30(target, target, ray, 19, object.DamageZapRay, r)
			if (calls == 1) == quest {
				t.Fatal("self ray must obey campaign/Quest owner rule")
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30PlayerZapRayDefendVampirismShield(t *testing.T) {
	for _, owner := range []string{"world", "Player", "NPC"} {
		for _, suppress := range []bool{false, true} {
			for _, remaining := range []int32{-3, 0, 7} {
				t.Run(fmt.Sprintf("%s/suppress-%t/remaining-%d", owner, suppress, remaining), func(t *testing.T) {
					target, source, ray := defaultDamagePlayerZapRayFixture4E0B30(t, owner)
					modifier := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
					armor := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, HealthData: &HealthData{Cur: 99}, InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, modifier}})}
					target.InvFirstItem = armor
					source.Buffs |= 1 << damageVampirismEnchant4E0B30
					source.BuffsPower[damageVampirismEnchant4E0B30] = 1
					var events []string
					frame, calls := uint32(1400), 0
					r := damageMeleeWorldRuntime4E0B30(t)
					queries := 0
					r.IsEnemy = func(*Object, *Object) bool {
						queries++
						if queries > 1 {
							events = append(events, "first-hit")
							frame = 1490
						}
						return true
					}
					frames := 0
					r.Frame = func() uint32 { frames++; return frame }
					before := *target.UpdateDataPlayer()
					r.BuffOff = func(victim *Object, id EnchantID) {
						if victim != target || id != 0 || *victim.UpdateDataPlayer() != before {
							t.Fatal("ray default marker clear")
						}
						events = append(events, "buff")
					}
					readies := 0
					r.CanApplyLateDefend = func(m *ModifierEff) bool { readies++; return m == modifier }
					r.ApplyLateDefend = func(m *ModifierEff, item, victim, weapon, attacker *Object, d int32, typ object.DamageType) int32 {
						if m != modifier || item != armor || victim != target || weapon != ray || attacker != source || d != 19 || typ != object.DamageZapRay || target.Frame134 != 0 {
							t.Fatal("ray late defend identity/order")
						}
						events = append(events, "defend")
						target.Buffs |= 1 << defaultDamageShieldEnchant4E0B30
						return 31
					}
					r.MonsterHasHitSound = func(*Object) bool { events = append(events, "hit-sound"); return suppress }
					r.DefaultDamageSound = func(victim, attacker *Object) {
						if victim != target || attacker != ray || target.Frame134 != 1400 {
							t.Fatal("ray sound")
						}
						events = append(events, "sound")
					}
					r.Audio = func(id int, got *Object) {
						if id != 163 || got != ray {
							t.Fatal("Vampirism ray audio")
						}
						events = append(events, "vamp-audio")
					}
					r.BalanceFloatInd = func(key string, index int) float64 {
						if key != damageVampirismBalance4E0B30 || index != 0 {
							t.Fatal("Vampirism coefficient")
						}
						return 0.5
					}
					r.AdjustHP = func(got *Object, heal int32) {
						if got != source || heal != 16 {
							t.Fatal("Vampirism pre-Shield rounding")
						}
						events = append(events, "heal")
					}
					r.VampirismFX = func(id int, _, _ image.Point, amount uint16) {
						if id != 162 || amount != 16 {
							t.Fatal("Vampirism FX")
						}
						events = append(events, "vamp-FX")
					}
					r.GameBallType = 777
					r.GameBallOnDamage = func(attacker, victim *Object, d int32) {
						if attacker != source || victim != target || d != 31 {
							t.Fatal("GameBall pre-Shield input")
						}
						events = append(events, "ball")
					}
					r.PlayerSetState = func(victim *Object, st PlayerState) bool {
						if victim != target || st != PlayerState30 {
							t.Fatal("hurt input")
						}
						events = append(events, "hurt")
						victim.UpdateDataPlayer().State = st
						return true
					}
					r.ShieldReduce = func(victim *Object, d *int32, typ object.DamageType, attacker *Object) {
						if victim != target || *d != 31 || typ != object.DamageZapRay || attacker != ray || victim.UpdateDataPlayer().State != PlayerState30 || (owner == "NPC" && source.UpdateDataMonster().Field130 != 1490) {
							t.Fatal("Shield after hurt/first-hit order")
						}
						events = append(events, "shield")
						target.Buffs &^= 1 << defaultDamageShieldEnchant4E0B30
						*d = remaining
					}
					r.DamageClear = func(victim *Object, d int32) {
						if victim != target || d != remaining {
							t.Fatal("Shield raw tail input")
						}
						calls++
						events = append(events, "HP")
					}
					result := DefaultDamageWorld4E0B30(target, source, ray, 19, object.DamageZapRay, r)
					want := []string{"buff", "defend"}
					if owner == "NPC" {
						want = append(want, "hit-sound")
					}
					if owner != "NPC" || !suppress {
						want = append(want, "sound")
					}
					want = append(want, "vamp-audio", "heal", "vamp-FX", "ball", "hurt")
					if owner == "NPC" {
						want = append(want, "first-hit")
					}
					want = append(want, "shield")
					wantCalls, wantFrames := 0, 1
					if owner == "NPC" {
						wantFrames++
					}
					if remaining != 0 {
						want = append(want, "HP")
						wantCalls = 1
					}
					before.State = PlayerState30
					if result != (remaining != 0) || calls != wantCalls || frames != wantFrames || !slices.Equal(events, want) || *target.UpdateDataPlayer() != before || armor.HealthData.Cur != 99 || readies != 2 {
						t.Fatalf("ray defense result=%t calls=%d frames=%d events=%v want=%v readiness=%d", result, calls, frames, events, want, readies)
					}
				})
			}
		}
	}
}

func TestDefaultDamageWorld4E0B30PlayerZapRayLiveHurt(t *testing.T) {
	for _, hook := range []string{"buff", "defend", "sound", "Vampirism", "ball"} {
		for _, state := range []PlayerState{PlayerState13, PlayerState1, PlayerState15} {
			t.Run(fmt.Sprintf("%s/state-%d", hook, state), func(t *testing.T) {
				target, source, ray := defaultDamagePlayerZapRayFixture4E0B30(t, "Player")
				old := target.UpdateDataPlayer()
				live := &PlayerUpdateData{Player: old.Player, State: state, Field75: 91, Field76: 92, Field40_0: 0x5678}
				r := damageMeleeWorldRuntime4E0B30(t)
				replace := func() { target.UpdateData = unsafe.Pointer(live) }
				switch hook {
				case "buff":
					r.BuffOff = func(*Object, EnchantID) { replace() }
				case "defend":
					m := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
					target.InvFirstItem = &Object{ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, m}})}
					r.CanApplyLateDefend = func(*ModifierEff) bool { return true }
					r.ApplyLateDefend = func(_ *ModifierEff, _, _, _, _ *Object, d int32, _ object.DamageType) int32 { replace(); return d }
				case "sound":
					r.DefaultDamageSound = func(*Object, *Object) { replace() }
				case "Vampirism":
					source.Buffs = 1 << damageVampirismEnchant4E0B30
					source.BuffsPower[damageVampirismEnchant4E0B30] = 1
					r.Audio = func(int, *Object) {}
					r.BalanceFloatInd = func(string, int) float64 { return 0.5 }
					r.AdjustHP = func(*Object, int32) {}
					r.VampirismFX = func(int, image.Point, image.Point, uint16) { replace() }
				case "ball":
					r.GameBallOnDamage = func(*Object, *Object, int32) { replace() }
				}
				hurt, calls := 0, 0
				r.PlayerSetState = nil
				if state == PlayerState13 {
					r.PlayerSetState = func(victim *Object, st PlayerState) bool {
						if victim != target || st != PlayerState30 || victim.UpdateDataPlayer() != live {
							t.Fatal("hurt retained stale Player update")
						}
						hurt++
						live.State = st
						return true
					}
				}
				r.DamageClear = func(victim *Object, d int32) {
					if victim != target || d != 27 {
						t.Fatal("live hurt HP input")
					}
					calls++
				}
				DefaultDamageWorld4E0B30(target, source, ray, 27, object.DamageZapRay, r)
				wantHurt := 0
				if state == PlayerState13 {
					wantHurt = 1
				}
				if calls != 1 || hurt != wantHurt || old.State != PlayerState13 || old.Field75 != 77 || old.Field76 != 88 || live.Field75 != 91 || live.Field76 != 92 || live.Field40_0 != 0x5678 {
					t.Fatalf("live ray hurt=%d calls=%d old/live=%d/%d", hurt, calls, old.State, live.State)
				}
			})
		}
	}
	for _, mutation := range []string{"NPC", "non-unit", "nil update", "missing hurt"} {
		t.Run("ball/live-"+mutation, func(t *testing.T) {
			target, source, ray := defaultDamagePlayerZapRayFixture4E0B30(t, "world")
			r := damageMeleeWorldRuntime4E0B30(t)
			r.PlayerSetState = nil
			r.GameBallOnDamage = func(*Object, *Object, int32) {
				switch mutation {
				case "NPC":
					target.ObjClass = object.ClassMonster
					target.UpdateData = unsafe.Pointer(&MonsterUpdateData{})
				case "non-unit":
					target.ObjClass = object.ClassSimple
				case "nil update":
					target.UpdateData = nil
				}
			}
			why, calls := "", 0
			r.Unsupported = func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) { why = reason }
			r.DamageClear = func(*Object, int32) { calls++ }
			DefaultDamageWorld4E0B30(target, source, ray, 27, object.DamageZapRay, r)
			fault := mutation == "nil update" || mutation == "missing hurt"
			if (why != "") != fault || (calls == 0) != fault || calls > 1 || target.Obj130 != ray || target.Pos132 != ray.PrevPos || target.Frame134 != 1400 {
				t.Fatalf("live class/hurt fault=%s reason=%q calls=%d", mutation, why, calls)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30PlayerZapRayRequiredServices(t *testing.T) {
	for _, owner := range []string{"world", "Player", "NPC"} {
		for _, missing := range []string{"enemy", "buff", "HP", "hit-sound", "player-sound"} {
			t.Run(owner+"/"+missing, func(t *testing.T) {
				target, source, ray := defaultDamagePlayerZapRayFixture4E0B30(t, owner)
				before := *target.UpdateDataPlayer()
				r := damageMeleeWorldRuntime4E0B30(t)
				why, calls := "", 0
				r.Unsupported = func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) { why = reason }
				r.DamageClear = func(*Object, int32) { calls++ }
				switch missing {
				case "enemy":
					r.IsEnemy = nil
				case "buff":
					r.BuffOff = nil
				case "HP":
					r.DamageClear = nil
				case "hit-sound":
					r.MonsterHasHitSound = nil
				case "player-sound":
					r.PlayerDamageSoundC = unsafe.Pointer(new(byte))
					target.DamageSound = r.PlayerDamageSoundC
				}
				DefaultDamageWorld4E0B30(target, source, ray, 19, object.DamageZapRay, r)
				unused := owner != "NPC" && missing == "hit-sound"
				if unused {
					if why != "" || calls != 1 {
						t.Fatalf("unused service required: %q calls=%d", why, calls)
					}
					return
				}
				if !strings.HasPrefix(why, "missing player") || calls != 0 || target.Obj130 != nil || target.Frame134 != 0 || *target.UpdateDataPlayer() != before {
					t.Fatalf("missing %s committed default Player ray: %q calls=%d", missing, why, calls)
				}
			})
		}
	}
	for _, liveService := range []string{"Shield", "GameBall type", "GameBall drop"} {
		t.Run("live/"+liveService, func(t *testing.T) {
			target, source, ray := defaultDamagePlayerZapRayFixture4E0B30(t, "world")
			r := damageMeleeWorldRuntime4E0B30(t)
			why, calls := "", 0
			r.Unsupported = func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) { why = reason }
			r.DamageClear = func(*Object, int32) { calls++ }
			r.DefaultDamageSound = func(*Object, *Object) {
				if liveService == "Shield" {
					target.Buffs |= 1 << defaultDamageShieldEnchant4E0B30
				} else {
					target.Field129 = &Object{TypeInd: 777, ObjOwner: target}
				}
			}
			if liveService == "GameBall drop" {
				r.GameBallType = 777
			}
			DefaultDamageWorld4E0B30(target, source, ray, 31, object.DamageZapRay, r)
			if !strings.Contains(why, "unsupported live "+liveService) || calls != 0 || target.Obj130 != ray || target.Frame134 != 1400 || target.UpdateDataPlayer().Field76 != 88 {
				t.Fatalf("live service=%s reason=%q calls=%d", liveService, why, calls)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30PlayerZapRayEarlyGatesAndBoundary(t *testing.T) {
	for _, gate := range []string{"Invulnerable", "Dead", "Zombie", "NoUpdate"} {
		t.Run(gate, func(t *testing.T) {
			target, source, ray := defaultDamagePlayerZapRayFixture4E0B30(t, "world")
			before := *target.UpdateDataPlayer()
			r := DefaultDamageWorldRuntime4E0B30{Frame: func() uint32 { return 1400 }, IsEnemy: func(*Object, *Object) bool { return true }, Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(reason) }}
			switch gate {
			case "Invulnerable":
				target.Buffs = 1 << defaultDamageInvulnerableEnchant4E0B30
			case "Dead", "Zombie":
				target.ObjFlags |= object.FlagDead
				r.IsZombie = func(*Object) bool { return gate == "Zombie" }
			case "NoUpdate":
				target.ObjFlags |= object.FlagNoUpdate
			}
			if !DefaultDamageWorld4E0B30(target, source, ray, 500, object.DamageZapRay, r) || *target.UpdateDataPlayer() != before || target.HealthData.Cur != 2000 {
				t.Fatal("Player ray early gate changed update/HP")
			}
			if gate == "Zombie" {
				if target.Obj130 != ray || target.Frame134 != 1400 || target.Field131 != 16 {
					t.Fatal("zombie ray metadata")
				}
			} else if target.Obj130 != nil || target.Frame134 != 0 {
				t.Fatal("early ray gate committed metadata")
			}
		})
	}
	for _, invalid := range []string{"nil source", "nil ray", "distinct world", "nil source update", "source weapon", "no simple", "no immobile", "ray weapon", "ray wand", "ray missile", "ray unit", "wrong type", "nil target update", "nil health"} {
		t.Run("boundary/"+invalid, func(t *testing.T) {
			target, source, ray := defaultDamagePlayerZapRayFixture4E0B30(t, "Player")
			typ := object.DamageZapRay
			switch invalid {
			case "nil source":
				source = nil
			case "nil ray":
				ray = nil
			case "distinct world":
				source = &Object{ObjClass: object.ClassSimple}
			case "nil source update":
				source.UpdateData = nil
			case "source weapon":
				source.ObjClass |= object.ClassWeapon
			case "no simple":
				ray.ObjClass &^= object.ClassSimple
			case "no immobile":
				ray.ObjClass &^= object.ClassImmobile
			case "ray weapon":
				ray.ObjClass |= object.ClassWeapon
			case "ray wand":
				ray.ObjClass |= object.ClassWand
			case "ray missile":
				ray.ObjClass |= object.ClassMissile
			case "ray unit":
				ray.ObjClass |= object.ClassPlayer
			case "wrong type":
				typ = object.DamageManaBomb
			case "nil target update":
				target.UpdateData = nil
			case "nil health":
				target.HealthData = nil
			}
			why := ""
			r := damageMeleeWorldRuntime4E0B30(t)
			r.Unsupported = func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) { why = reason }
			r.DamageClear = func(*Object, int32) { t.Fatal("unsupported ray shape reached HP") }
			DefaultDamageWorld4E0B30(target, source, ray, 500, typ, r)
			if why == "" || target.Obj130 != nil || target.Pos132 != (types.Pointf{}) || target.Frame134 != 0 {
				t.Fatalf("ray boundary admitted: %s reason=%q", invalid, why)
			}
		})
	}
}
