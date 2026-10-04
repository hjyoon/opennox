package server

import (
	"fmt"
	"image"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// Stock SentryGlobe is LIGHT|SIMPLE|IMMOBILE|VISIBLE_ENABLE. An unowned
// globe is its own terminal owner, not a nil source. These contract tests
// record the HP service's input; the registered native test checks real HP.
func defaultDamageZapRayFixture4E0B30(t *testing.T, owner string, subclass uint32) (target, source, ray *Object) {
	t.Helper()
	target = damageMeleeUnitFixture4E17B0(t, false)
	target.ObjSubClass = object.SubClass(subclass)
	target.HealthData = &HealthData{Cur: 2000, Field2: 2000, Max: 2000}
	*target.UpdateDataMonster() = MonsterUpdateData{Field547: 99, Field546: 77, Field523_2: 88, Field1: math.Float32bits(0.125)}
	ray = &Object{TypeInd: 801, ObjClass: object.ClassLight | object.ClassSimple | object.ClassImmobile | object.ClassVisibleEnable,
		PrevPos: types.Ptf(44, 7), PosVec: types.Ptf(10, 20)}
	switch owner {
	case "world":
		source = ray
	case "Player", "NPC":
		source = damageMeleeUnitFixture4E17B0(t, owner == "Player")
		ray.ObjOwner = source
	default:
		t.Fatal("invalid ray owner fixture")
	}
	return
}

func TestDefaultDamageWorld4E0B30ZapRaySignedTail(t *testing.T) {
	for _, owner := range []string{"world", "Player", "NPC"} {
		for _, subclass := range []uint32{0x202, 0x11012, 0xe00} {
			for _, raw := range []int32{-7, 0, 1, 19, 20, 500} {
				t.Run(fmt.Sprintf("%s/subclass-%x/raw-%d", owner, subclass, raw), func(t *testing.T) {
					target, source, ray := defaultDamageZapRayFixture4E0B30(t, owner, subclass)
					target.Buffs = 1 << defaultDamageShockEnchant4E0B30
					var events []string
					calls := 0
					r := damageMeleeWorldRuntime4E0B30(t)
					r.Frame = func() uint32 { events = append(events, "frame"); return 1400 }
					r.GameplayFlag1 = func() bool { events = append(events, "gameplay"); return true }
					r.IsEnemy = func(got, attacker *Object) bool {
						if got != target || attacker != source {
							t.Fatal("ray enemy identity")
						}
						events = append(events, "enemy")
						return false // Non-melee ray must not use the melee friendly gate.
					}
					r.BuffOff = func(got *Object, id EnchantID) {
						if got != target || id != defaultDamageInvisibleEnchant4E0B30 {
							t.Fatal("ray BuffOff identity")
						}
						want := uint32(1)
						if owner == "world" {
							want = 0
						}
						if target.UpdateDataMonster().Field547 != want || target.Frame134 != 0 {
							t.Fatal("ray attribution/BuffOff order")
						}
						events = append(events, "buff")
					}
					r.DefaultDamageSound = func(got, attacker *Object) {
						if got != target || attacker != ray || target.Frame134 != 1400 {
							t.Fatal("ray sound identity/order")
						}
						events = append(events, "sound")
					}
					r.MonsterHasHitSound = func(got *Object) bool {
						if got != source || owner != "NPC" {
							t.Fatal("ray monster hit-sound identity")
						}
						events = append(events, "hit-sound")
						return false
					}
					r.AdjustFieldGuide = func(gotSource, gotTarget *Object, d int32) int32 {
						if gotSource != source || gotTarget != target || d != raw {
							t.Fatal("ray field-guide identity/raw damage")
						}
						events = append(events, "guide")
						return d
					}
					r.FireProtection = func(*Object) float64 { t.Fatal("ZAP_RAY is not fire damage"); return 1 }
					r.ElectricProtection = func(*Object) float64 { t.Fatal("ZAP_RAY is not electric damage 9/17"); return 1 }
					r.CallDamage = func(*Object, *Object, *Object, int32, object.DamageType) bool {
						t.Fatal("non-melee ray triggered Shock")
						return false
					}
					r.PlayerSetState = func(*Object, PlayerState) bool { t.Fatal("monster ray requested player state"); return false }
					r.DamageClear = func(got *Object, d int32) {
						if got != target || d != raw {
							t.Fatalf("ray HP input=%p/%d want=%p/%d", got, d, target, raw)
						}
						calls++
						events = append(events, "HP")
					}
					if !DefaultDamageWorld4E0B30(target, source, ray, raw, object.DamageZapRay, r) || calls != 1 {
						t.Fatal("ray tail did not reach HP exactly once")
					}
					want := []string{"gameplay", "enemy", "buff", "frame"}
					if owner == "NPC" {
						want = append(want, "hit-sound")
					}
					want = append(want, "sound", "guide")
					if owner == "NPC" {
						want = append(want, "enemy")
					}
					want = append(want, "HP")
					if !slices.Equal(events, want) {
						t.Fatalf("ray events=%v want=%v", events, want)
					}
					ud := target.UpdateDataMonster()
					wantMarker, wantType := uint32(1), uint32(ray.TypeInd)
					if owner == "world" {
						wantMarker, wantType = 2, uint32(object.DamageZapRay)
					}
					if target.Obj130 != ray || target.Pos132 != ray.PrevPos || target.Field131 != 16 || target.Frame134 != 1400 || ud.Field547 != wantMarker || ud.Field546 != wantType || ud.Field1 != math.Float32bits(0.125) || ud.Field523_2 != 88 || !ud.StatusFlags.Has(object.MonStatusInjured) || ud.StatusFlags.Has(object.MonStatusOnFire) || !target.HasEnchant(defaultDamageShockEnchant4E0B30) || target.HealthData.Cur != 2000 {
						t.Fatalf("ray metadata marker=%d/%d carry=%g state=%d status=%x HP=%d", ud.Field547, ud.Field546, math.Float32frombits(ud.Field1), ud.Field523_2, ud.StatusFlags, target.HealthData.Cur)
					}
				})
			}
		}
	}
}

func TestDefaultDamageWorld4E0B30ZapRayOwnerFriendlyGate(t *testing.T) {
	for _, owner := range []string{"world", "Player", "NPC"} {
		for _, gameplay := range []bool{false, true} {
			for _, enemy := range []bool{false, true} {
				for _, quest := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/gameplay-%t/enemy-%t/quest-%t", owner, gameplay, enemy, quest), func(t *testing.T) {
						target, source, ray := defaultDamageZapRayFixture4E0B30(t, owner, 0x202)
						r := damageMeleeWorldRuntime4E0B30(t)
						r.GameplayFlag1 = func() bool { return gameplay }
						r.QuestMode = func() bool { return quest }
						r.IsEnemy = func(*Object, *Object) bool { return enemy }
						calls := 0
						r.DamageClear = func(*Object, int32) { calls++ }
						DefaultDamageWorld4E0B30(target, source, ray, 500, object.DamageZapRay, r)
						want := 1
						if owner != "world" && !gameplay && !enemy {
							want = 0
						}
						if calls != want || (want == 0 && (target.Obj130 != nil || target.Frame134 != 0 || target.UpdateDataMonster().Field547 != 0)) {
							t.Fatalf("ray campaign gate calls=%d want=%d", calls, want)
						}
					})
				}
			}
		}
	}
	for _, quest := range []bool{false, true} {
		t.Run(fmt.Sprintf("NPC-self/quest-%t", quest), func(t *testing.T) {
			target, _, ray := defaultDamageZapRayFixture4E0B30(t, "world", 0x11012)
			ray.ObjOwner = target
			r := damageMeleeWorldRuntime4E0B30(t)
			r.GameplayFlag1 = func() bool { return false }
			r.IsEnemy = func(*Object, *Object) bool { return false }
			r.QuestMode = func() bool { return quest }
			calls := 0
			r.DamageClear = func(*Object, int32) { calls++ }
			DefaultDamageWorld4E0B30(target, target, ray, 19, object.DamageZapRay, r)
			want := 1
			if quest {
				want = 0
			}
			if calls != want {
				t.Fatalf("self-owned ray calls=%d want=%d", calls, want)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30ZapRayShieldAndVampirism(t *testing.T) {
	for _, owner := range []string{"world", "Player", "NPC"} {
		for _, remaining := range []int32{-3, 0, 7} {
			t.Run(fmt.Sprintf("%s/shield-%d", owner, remaining), func(t *testing.T) {
				target, source, ray := defaultDamageZapRayFixture4E0B30(t, owner, 0x202)
				target.Buffs = 1 << defaultDamageShieldEnchant4E0B30
				source.Buffs |= 1 << damageVampirismEnchant4E0B30
				source.BuffsPower[damageVampirismEnchant4E0B30] = 2
				var events []string
				r := damageMeleeWorldRuntime4E0B30(t)
				r.Audio = func(id int, got *Object) {
					if id != 163 || got != ray {
						t.Fatal("ray Vampirism sound")
					}
					events = append(events, "audio")
				}
				r.BalanceFloatInd = func(key string, index int) float64 {
					if key != "VampirismCoeff" || index != 1 {
						t.Fatal("ray Vampirism coefficient")
					}
					events = append(events, "balance")
					return 0.5
				}
				r.AdjustHP = func(got *Object, d int32) {
					if got != source || d != 10 {
						t.Fatal("ray Vampirism healing")
					}
					events = append(events, "heal")
				}
				r.VampirismFX = func(id int, from, to image.Point, amount uint16) {
					if id != 162 || from != damageVampirismPoint4E0B30(source.PosVec) || to != damageVampirismPoint4E0B30(target.PosVec) || amount != 10 {
						t.Fatal("ray Vampirism FX")
					}
					events = append(events, "fx")
				}
				r.ShieldReduce = func(got *Object, d *int32, typ object.DamageType, attacker *Object) {
					if got != target || *d != 19 || typ != object.DamageZapRay || attacker != ray {
						t.Fatal("ray Shield identity/raw damage")
					}
					events = append(events, "shield")
					*d = remaining
				}
				calls := 0
				r.DamageClear = func(got *Object, d int32) {
					if got != target || d != remaining {
						t.Fatal("ray post-Shield damage")
					}
					calls++
					events = append(events, "HP")
				}
				result := DefaultDamageWorld4E0B30(target, source, ray, 19, object.DamageZapRay, r)
				want := []string{"audio", "balance", "heal", "fx", "shield"}
				wantCalls := 0
				if remaining != 0 {
					want = append(want, "HP")
					wantCalls = 1
				}
				if result != (remaining != 0) || calls != wantCalls || !slices.Equal(events, want) {
					t.Fatalf("ray Shield result=%t calls=%d events=%v want=%v", result, calls, events, want)
				}
			})
		}
	}
}

func TestDefaultDamageWorld4E0B30ZapRayMissingServices(t *testing.T) {
	for _, owner := range []string{"world", "Player", "NPC"} {
		for _, missing := range []string{"enemy", "buff", "HP", "hit-sound"} {
			t.Run(owner+"/"+missing, func(t *testing.T) {
				target, source, ray := defaultDamageZapRayFixture4E0B30(t, owner, 0x202)
				r := damageMeleeWorldRuntime4E0B30(t)
				why, calls := "", 0
				r.Unsupported = func(reason string, got, attacker, weapon *Object, d int32, typ object.DamageType) {
					if got != target || attacker != source || weapon != ray || d != 500 || typ != object.DamageZapRay {
						t.Fatal("ray unsupported identity")
					}
					why = reason
				}
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
				}
				DefaultDamageWorld4E0B30(target, source, ray, 500, object.DamageZapRay, r)
				unused := missing == "hit-sound" && owner != "NPC"
				if unused {
					if why != "" || calls != 1 {
						t.Fatalf("unused hit sound was required: %q calls=%d", why, calls)
					}
				} else {
					wantReason := "missing monster ZAP_RAY tail service"
					if missing == "enemy" {
						wantReason = "missing monster ZAP_RAY enemy service"
					}
					if why != wantReason || calls != 0 || target.Obj130 != nil || target.Pos132 != (types.Pointf{}) || target.Frame134 != 0 || target.UpdateDataMonster().Field547 != 0 || target.UpdateDataMonster().Field546 != 77 {
						t.Fatalf("missing ray %s committed tail: why=%q want=%q", missing, why, wantReason)
					}
				}
			})
		}
	}
}

func TestDefaultDamageWorld4E0B30ZapRayKeepsShapeBoundary(t *testing.T) {
	for _, invalid := range []string{"nil source", "nil ray", "distinct world source", "nil unit update", "mixed source weapon", "no simple", "no immobile", "weapon ray", "wand ray", "missile ray", "unit ray", "wrong type", "player without update"} {
		t.Run(invalid, func(t *testing.T) {
			target, source, ray := defaultDamageZapRayFixture4E0B30(t, "Player", 0x202)
			typ := object.DamageZapRay
			switch invalid {
			case "nil source":
				source = nil
			case "nil ray":
				ray = nil
			case "distinct world source":
				source = &Object{ObjClass: object.ClassImmobile}
			case "nil unit update":
				source.UpdateData = nil
			case "mixed source weapon":
				source.ObjClass |= object.ClassWeapon
			case "no simple":
				ray.ObjClass &^= object.ClassSimple
			case "no immobile":
				ray.ObjClass &^= object.ClassImmobile
			case "weapon ray":
				ray.ObjClass |= object.ClassWeapon
			case "wand ray":
				ray.ObjClass |= object.ClassWand
			case "missile ray":
				ray.ObjClass |= object.ClassMissile
			case "unit ray":
				ray.ObjClass |= object.ClassMonster
			case "wrong type":
				typ = object.DamageManaBomb
			case "player without update":
				// Player ray tails are restored separately; missing native
				// update data must still fail before any hit metadata/HP.
				target = damageMeleeUnitFixture4E17B0(t, true)
				target.UpdateData = nil
			}
			r := damageMeleeWorldRuntime4E0B30(t)
			why := ""
			r.Unsupported = func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) { why = reason }
			r.DamageClear = func(*Object, int32) { t.Fatal("unported ray shape reached HP") }
			DefaultDamageWorld4E0B30(target, source, ray, 500, typ, r)
			if why == "" || target.Obj130 != nil || target.Frame134 != 0 {
				t.Fatalf("ray shape admitted: %q", why)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30ZapRayUsesLiveMetadata(t *testing.T) {
	for _, owner := range []string{"world", "Player", "NPC"} {
		t.Run(owner, func(t *testing.T) {
			target, source, ray := defaultDamageZapRayFixture4E0B30(t, owner, 0x11012)
			cached := target.UpdateDataMonster()
			live := &MonsterUpdateData{Field547: 31, Field546: 33, Field1: math.Float32bits(0.25)}
			frame := uint32(1400)
			r := damageMeleeWorldRuntime4E0B30(t)
			r.Frame = func() uint32 { return frame }
			r.BuffOff = func(*Object, EnchantID) { target.UpdateData = unsafe.Pointer(live); frame = 1470 }
			r.DefaultDamageSound = func(got, attacker *Object) {
				if got != target || attacker != ray || target.Frame134 != 1470 {
					t.Fatal("ray sound metadata")
				}
				frame = 1500
				target.Buffs |= 1 << defaultDamageShieldEnchant4E0B30
			}
			r.AdjustFieldGuide = func(gotSource, gotTarget *Object, d int32) int32 {
				if gotSource != source || gotTarget != target || d != 19 {
					t.Fatal("live ray guide")
				}
				return 27
			}
			r.ShieldReduce = func(got *Object, d *int32, typ object.DamageType, attacker *Object) {
				if got != target || *d != 27 || typ != object.DamageZapRay || attacker != ray || target.Frame134 != 1470 || live.Field547 != 31 {
					t.Fatal("live ray Shield")
				}
				*d = 7
			}
			calls := 0
			r.DamageClear = func(got *Object, d int32) {
				if got != target || d != 7 {
					t.Fatal("live ray HP")
				}
				calls++
			}
			DefaultDamageWorld4E0B30(target, source, ray, 19, object.DamageZapRay, r)
			wantMarker := uint32(1)
			if owner == "world" {
				wantMarker = 0
			}
			if calls != 1 || cached.Field547 != wantMarker || cached.StatusFlags.Has(object.MonStatusInjured) || live.Field547 != 31 || live.Field546 != 33 || !live.StatusFlags.Has(object.MonStatusInjured) || live.Field1 != math.Float32bits(0.25) || target.Frame134 != 1470 {
				t.Fatalf("live ray replayed/lost metadata: HP calls=%d cached=%d live=%d/%d frame=%d", calls, cached.Field547, live.Field547, live.Field546, target.Frame134)
			}
			if owner == "NPC" && source.UpdateDataMonster().Field130 != 1500 {
				t.Fatal("ray source first-hit frame is not live")
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30ZapRayEarlyGates(t *testing.T) {
	for _, gate := range []string{"Invulnerable", "Dead", "Zombie", "NoUpdate", "campaign-friendly"} {
		t.Run(gate, func(t *testing.T) {
			target, source, ray := defaultDamageZapRayFixture4E0B30(t, "Player", 0x202)
			r := damageMeleeWorldRuntime4E0B30(t)
			r.BuffOff, r.DamageClear, r.MonsterHasHitSound = nil, nil, nil
			r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) {
				t.Fatal("early gate required unused ray service: " + why)
			}
			r.Frame = func() uint32 { return 1400 }
			audio := 0
			r.Audio = func(id int, got *Object) {
				if id != defaultDamageInvulnerableSound4E0B30 || got != target || gate != "Invulnerable" {
					t.Fatal("ray early audio")
				}
				audio++
			}
			switch gate {
			case "Invulnerable":
				target.Buffs = 1 << defaultDamageInvulnerableEnchant4E0B30
				r.IsEnemy = nil
			case "Dead", "Zombie":
				target.ObjFlags |= object.FlagDead
				r.IsEnemy = nil
				r.IsZombie = func(*Object) bool { return gate == "Zombie" }
			case "NoUpdate":
				target.ObjFlags |= object.FlagNoUpdate
			case "campaign-friendly":
				r.GameplayFlag1 = func() bool { return false }
				r.IsEnemy = func(*Object, *Object) bool { return false }
			}
			if !DefaultDamageWorld4E0B30(target, source, ray, 500, object.DamageZapRay, r) {
				t.Fatal("ray early gate result")
			}
			wantMarker := uint32(0)
			if gate == "Invulnerable" {
				wantMarker = 99
			}
			if target.UpdateDataMonster().Field547 != wantMarker || target.UpdateDataMonster().Field546 != 77 || target.HealthData.Cur != 2000 {
				t.Fatal("ray early gate changed HP/latch")
			}
			if gate == "Zombie" {
				if target.Obj130 != ray || target.Field131 != 16 || target.Frame134 != 1400 {
					t.Fatal("ray zombie attribution")
				}
			} else if target.Obj130 != nil || target.Field131 != 0 || target.Frame134 != 0 {
				t.Fatal("ray early gate committed attribution")
			}
			if audio != 0 && gate != "Invulnerable" || gate == "Invulnerable" && audio != 1 {
				t.Fatal("ray invulnerable cadence")
			}
		})
	}
}

// This joins the production ray trace/circle/owner scan to DefaultDamage's
// restored tail. Geometry/HP services are recorded, not stock-map GUI input.
func TestDefaultDamageWorld4E0B30ZapRaySentryScan(t *testing.T) {
	for _, owner := range []string{"world", "Player", "NPC"} {
		for _, test := range []struct {
			name       string
			position   types.Pointf
			flags      object.Flags
			quest, hit bool
		}{
			{"crossing", types.Ptf(5, 2), 0, false, true},
			{"tangent", types.Ptf(5, 3), 0, false, false},
			{"outside", types.Ptf(5, 6), 0, false, false},
			{"before-start", types.Ptf(-5, 0), 0, false, false},
			{"past-end", types.Ptf(15, 0), 0, false, false},
			{"below", types.Ptf(5, 2), object.FlagBelow, false, false},
			{"no-collide", types.Ptf(5, 2), object.FlagNoCollide, false, false},
			{"short-outside-Quest", types.Ptf(5, 2), object.FlagShort, false, false},
			{"short-in-Quest", types.Ptf(5, 2), object.FlagShort, true, true},
			{"short-dead-in-Quest", types.Ptf(5, 2), object.FlagShort | object.FlagDead, true, false},
		} {
			t.Run(owner+"/"+test.name, func(t *testing.T) {
				target, source, ray := defaultDamageZapRayFixture4E0B30(t, owner, 0x202)
				target.PosVec, target.ObjFlags, target.Shape.Circle.R = test.position, test.flags, 3
				data := &SentryUpdateData{}
				ray.PosVec, ray.ObjFlags, ray.UpdateData = types.Ptf(0, 0), object.FlagEnabled, unsafe.Pointer(data)
				var events []string
				r := damageMeleeWorldRuntime4E0B30(t)
				r.DamageClear = func(got *Object, d int32) {
					if got != target || d != 500 {
						t.Fatal("scan ray HP input")
					}
					events = append(events, "HP")
				}
				var state sentryGlobeState510E60
				state.update(ray, sentryGlobeUpdateDeps510E60{
					traceRay: func(start, end types.Pointf) (types.Pointf, bool) {
						if start != ray.PosVec || end != types.Ptf(600, 0) {
							t.Fatal("ray trace endpoints")
						}
						return types.Ptf(10, 0), false
					},
					eachInRect: func(rect types.Rectf, fn func(*Object) bool) {
						if rect != types.RectFromPointsf(types.Ptf(0, 0), types.Ptf(10, 0)) || !fn(target) {
							t.Fatal("ray scan rectangle/continuation")
						}
					},
					questMode: test.quest,
					damage: func(got, parent, weapon *Object, d int, typ object.DamageType) {
						if got != target || parent != source || weapon != ray || d != 500 || typ != object.DamageZapRay {
							t.Fatal("ray scan owner/damage")
						}
						DefaultDamageWorld4E0B30(got, parent, weapon, int32(d), typ, r)
					},
					audio: func(got *Object) {
						if got != target {
							t.Fatal("ray hit sound identity")
						}
						events = append(events, "ray-hit")
					},
				})
				var want []string
				if test.hit {
					want = []string{"HP", "ray-hit"}
				}
				if !slices.Equal(events, want) || state.head != ray || ray.Pos39 != types.Ptf(10, 0) {
					t.Fatalf("ray scan events=%v want=%v end=%v", events, want, ray.Pos39)
				}
			})
		}
	}
}
