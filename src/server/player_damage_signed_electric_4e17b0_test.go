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

func signedElectricFixture4E17B0(t *testing.T, playerSource, selfWeapon, observed bool, carry float32, events *[]string, damages *[]int32) (target, source, weapon *Object, cached, live *PlayerUpdateData, r PlayerDamageRuntime4E17B0) {
	t.Helper()
	target, source, weapon = possessionElectricFixture4E17B0(t, playerSource, selfWeapon)
	cached = target.UpdateDataPlayer()
	cached.Player.Field3680, cached.Player.CameraFollowObj = 0, nil
	cached.Field21 = math.Float32bits(carry)
	live = cached
	if observed {
		cached.Player.Field3680, cached.Player.CameraFollowObj = 2, source
		cached.Field21 = math.Float32bits(0.125)
		live = &PlayerUpdateData{Player: &Player{}, State: PlayerState21, Field21: math.Float32bits(carry), Field76: 31, Field75: 33, Field40_0: 0x1122, Field40_1: 0xabcd}
	}
	r = playerDamageRuntime4E17B0(t, nil, damages)
	r.ObserveClear = func(v *Object) {
		if !observed || v != target || cached.Field76 != 0 || cached.Field75 != 77 || len(*events) != 0 {
			t.Fatal("signed electric ObserveClear preceded the cached marker reset")
		}
		*events = append(*events, "observe")
		target.UpdateData = unsafe.Pointer(live)
	}
	exclusion := func(v *Object) bool {
		if v != source || cached.Field76 != 0 {
			t.Fatal("signed electric exclusion lost the original attack or marker")
		}
		*events = append(*events, "exclude")
		return false
	}
	r.BlockSourceExcluded, r.BlockSourceOnlyExcluded = exclusion, exclusion
	r.BlockDirection = func(v *Object, pos types.Pointf) bool {
		if v != target || pos != source.PrevPos || cached.Field76 != 0 {
			t.Fatal("signed electric direction lost the prefix snapshot")
		}
		*events = append(*events, "direction")
		return false
	}
	r.ElectricArmorScale = func(v *Object) float32 {
		if v != target || cached.Field76 != 0 || target.UpdateDataPlayer() != live {
			t.Fatal("signed electric scale preceded the entry prefix")
		}
		*events = append(*events, "scale")
		return 0.5
	}
	return
}

// The IA-32 switch accepts signed damage; only 004E2011's minimum is gated
// by raw damage > 0. DefaultDamage's electric-protection minimum is separate,
// and UnitDamageClear forwards the low WORD even for a negative subtraction.
func TestPlayerDamageSignedElectric4E17B0(t *testing.T) {
	electricPlayerCases4E17B0(t, func(t *testing.T, playerSource, selfWeapon bool, typ object.DamageType) {
		for _, observed := range []bool{false, true} {
			for _, tc := range []struct {
				name                        string
				raw, toDefault, toHP        int32
				carry, residual, questScale float32
				quest, god, shield          bool
				hp                          uint16
			}{
				{name: "zero positive tie", raw: 0, carry: 0.5, residual: 0.5, toHP: 1, hp: 59},
				{name: "zero negative tie", raw: 0, carry: -0.5, residual: -0.5, toHP: 1, hp: 59},
				{name: "negative cancellation", raw: -1, carry: 0.5, toHP: 1, hp: 59},
				{name: "negative tie", raw: -5, carry: 0.5, toDefault: -2, toHP: -2, hp: 62},
				{name: "negative fractional carry", raw: -3, carry: 0.25, residual: -0.25, toDefault: -1, toHP: -1, hp: 61},
				{name: "negative Quest", raw: -5, carry: 0.5, quest: true, questScale: 0.5, toDefault: -1, toHP: -1, hp: 61},
				{name: "negative Quest zero", raw: -5, carry: 0.5, quest: true, questScale: 0, toHP: 1, hp: 59},
				{name: "positive minimum control", raw: 1, carry: -0.5, toDefault: 1, toHP: 1, hp: 59},
				{name: "zero God", carry: 0.5, residual: 0.5, god: true, hp: 60},
				{name: "negative God", raw: -5, carry: 0.5, god: true, hp: 60},
				{name: "zero Shield", carry: 0.5, residual: 0.5, toHP: 1, shield: true, hp: 60},
				{name: "negative Shield", raw: -5, carry: 0.5, toDefault: -2, toHP: -2, shield: true, hp: 60},
			} {
				t.Run(fmt.Sprintf("observed-%t/%s", observed, tc.name), func(t *testing.T) {
					var events []string
					var damages []int32
					target, source, weapon, cached, live, r := signedElectricFixture4E17B0(t, playerSource, selfWeapon, observed, tc.carry, &events, &damages)
					r.GodMode = func() bool { return tc.god }
					r.QuestMode = func() bool { return tc.quest }
					r.QuestDamageScale = func() float32 { events = append(events, "Quest"); return tc.questScale }
					if tc.shield {
						target.Buffs |= 1 << defaultDamageShieldEnchant4E0B30
					}
					r.DefaultDamage = func(v, a, w *Object, d int32, gotType object.DamageType) bool {
						if tc.god || v != target || a != source || w != weapon || d != tc.toDefault || gotType != typ || cached.Field76 != 2 || cached.Field75 != uint32(typ) {
							t.Fatalf("signed electric default arguments/marker: damage=%d", d)
						}
						events = append(events, "default")
						return DefaultDamageWorld4E0B30(v, a, w, d, gotType, DefaultDamageWorldRuntime4E0B30{
							Frame: r.Frame, GameplayFlag1: func() bool { return true }, IsEnemy: r.IsEnemy,
							ElectricProtection: func(*Object) float64 { events = append(events, "protection"); return 0.25 },
							MonsterHasHitSound: func(*Object) bool { return false },
							PlayerSetState:     func(*Object, PlayerState) bool { t.Fatal("small signed electric hurt state"); return false },
							ShieldReduce: func(v *Object, d *int32, gotType object.DamageType, a *Object) {
								if !tc.shield || v != target || *d != tc.toHP || gotType != typ || a != source {
									t.Fatal("Shield lost the signed protected damage")
								}
								events = append(events, "Shield")
								*d = 0
							},
							DamageClear: r.DamageClear, Unsupported: r.Unsupported,
						})
					}
					h, result := PlayerDamageNative4E17B0(target, source, weapon, tc.raw, typ, r)
					var want []string
					if observed {
						want = append(want, "observe")
					}
					want = append(want, "exclude", "direction", "scale")
					if !tc.god {
						if tc.quest {
							want = append(want, "Quest")
						}
						want = append(want, "default", "protection")
						if tc.shield {
							want = append(want, "Shield")
						}
					}
					if !h || result == tc.shield || !slices.Equal(events, want) || target.HealthData.Cur != tc.hp || live.Field21 != math.Float32bits(tc.residual) || cached.Field76 != 2 || cached.Field75 != uint32(typ) {
						t.Fatalf("signed electric=%t/%t HP=%d carry=%g marker=%d/%d events=%v want=%v", h, result, target.HealthData.Cur, math.Float32frombits(live.Field21), cached.Field76, cached.Field75, events, want)
					}
					if observed && (cached.Field21 != math.Float32bits(0.125) || live.Field76 != 31 || live.Field75 != 33) {
						t.Fatal("signed electric overwrote the cached carry or replacement live marker")
					}
					if tc.god || tc.shield {
						if len(damages) != 0 {
							t.Fatal("God/Shield reached HP subtraction")
						}
					} else if !slices.Equal(damages, []int32{tc.toHP}) {
						t.Fatalf("signed HP damage=%v want=%d", damages, tc.toHP)
					}
					if tc.god {
						if target.Obj130 != nil || live.Field40_0 != 0x1122 || live.Field40_1 != 0xabcd {
							t.Fatal("God entered attribution/protection tail")
						}
					} else if target.Obj130 != source || target.Pos132 != source.PrevPos || target.Field131 != uint32(typ) || target.Frame134 != 700 || live.Field40_0 != 2 || live.Field40_1 != 0xabcd {
						t.Fatal("signed electric lost attribution or adjacent protection WORD")
					}
					if !tc.god && !playerSource && source.UpdateDataMonster().Field130 != 700 {
						t.Fatal("signed electric lost the NPC first-hit latch")
					}
				})
			}
		}
	})
}

func TestPlayerDamageSignedElectricPossessionReflect4E17B0(t *testing.T) {
	electricPlayerCases4E17B0(t, func(t *testing.T, playerSource, selfWeapon bool, typ object.DamageType) {
		for _, raw := range []int32{0, -5} {
			for _, front := range []bool{false, true} {
				t.Run(fmt.Sprintf("raw-%d/front-%t", raw, front), func(t *testing.T) {
					var events []string
					var damages []int32
					target, source, weapon, cached, live, r := signedElectricFixture4E17B0(t, playerSource, selfWeapon, true, 0.5, &events, &damages)
					target.Buffs |= 1 << playerDamageReflectEnchant4E17B0
					blocked := front && typ == object.DamageAirborneElectric
					r.BlockDirection = func(v *Object, pos types.Pointf) bool {
						if v != target || cached.Field76 != 0 || target.UpdateDataPlayer() != live {
							t.Fatal("signed Reflect ran before ObserveClear")
						}
						if pos == source.PosVec {
							events = append(events, "Reflect")
							return front
						}
						if blocked || pos != source.PrevPos {
							t.Fatal("signed Reflect/source-only direction")
						}
						events = append(events, "direction")
						return front
					}
					r.Audio = func(id int, v *Object) {
						if !blocked || id != 122 || v != target {
							t.Fatal("signed Reflect sound")
						}
						events = append(events, "audio")
					}
					r.DefaultDamage = func(v, a, w *Object, d int32, gotType object.DamageType) bool {
						wantDamage := int32(0)
						if raw < 0 {
							wantDamage = -2
						}
						if blocked || v != target || a != source || w != weapon || d != wantDamage || gotType != typ {
							t.Fatal("signed Reflect default")
						}
						events = append(events, "default")
						return true
					}
					h, result := PlayerDamageNative4E17B0(target, source, weapon, raw, typ, r)
					want := []string{"observe"}
					if typ == object.DamageAirborneElectric {
						want = append(want, "Reflect")
					}
					marker, markerType, residual := uint32(2), uint32(typ), float32(0.5)
					if blocked {
						want = append(want, "audio")
						marker, markerType = 0, 77
					} else {
						want = append(want, "exclude", "direction", "scale", "default")
						if raw < 0 {
							residual = 0
						}
					}
					if !h || result == blocked || !slices.Equal(events, want) || cached.Field76 != marker || cached.Field75 != markerType || live.Field21 != math.Float32bits(residual) || cached.Field21 != math.Float32bits(0.125) || live.Field76 != 31 || live.Field75 != 33 || target.HealthData.Cur != 60 {
						t.Fatalf("signed Reflect=%t/%t events=%v marker=%d/%d carry=%g", h, result, events, cached.Field76, cached.Field75, math.Float32frombits(live.Field21))
					}
				})
			}
		}
	})
}

func TestPlayerDamageSignedElectricArmor4E17B0(t *testing.T) {
	electricPlayerCases4E17B0(t, func(t *testing.T, playerSource, selfWeapon bool, typ object.DamageType) {
		for _, observed := range []bool{false, true} {
			for _, raw := range []int32{0, -5} {
				t.Run(fmt.Sprintf("observed-%t/raw-%d", observed, raw), func(t *testing.T) {
					var events []string
					var damages []int32
					target, source, weapon, cached, live, r := signedElectricFixture4E17B0(t, playerSource, selfWeapon, observed, 0.5, &events, &damages)
					_, _, armor, carry, modifier := playerDamageElectricSelfArmorFixture4E17B0(t, true, playerSource)
					target.InvFirstItem, live.Field57 = armor, math.Float32bits(0.4)
					r.ItemArmorValue = func(v *Object) float32 {
						if v != armor || cached.Field76 != 0 {
							t.Fatal("signed armor lookup order")
						}
						events = append(events, "armor lookup")
						return 0.4
					}
					r.CanApplyArmorDefend = func(m *ModifierEff) bool { return m == modifier }
					r.ApplyArmorDefend = func(m *ModifierEff, item, victim, w, a *Object, amount *float32) bool {
						if m != modifier || item != armor || victim != target || w != weapon || a != source || *amount != float32(raw) || cached.Field76 != 0 {
							t.Fatal("armor skipped signed raw damage or substituted the weapon")
						}
						events = append(events, "armor defend")
						*amount = 1.5
						return true
					}
					r.CanDamageArmor = func(v *Object) bool { return v == armor }
					r.DamageArmor = func(item, a, w *Object, d int32, gotType object.DamageType) bool {
						if item != armor || a != source || w != weapon || d != 2 || gotType != typ || cached.Field76 != 0 || *carry != -0.25 {
							t.Fatal("signed armor damage/carry order")
						}
						events = append(events, "armor damage")
						item.HealthData.Cur -= uint16(d)
						return true
					}
					r.ReportArmorHealth = func(owner, item *Object, before, after uint16) {
						if owner != target || item != armor || before != 100 || after != 98 {
							t.Fatal("signed armor report")
						}
						events = append(events, "report")
					}
					r.GodMode = func() bool { events = append(events, "God"); return true }
					r.DefaultDamage = func(*Object, *Object, *Object, int32, object.DamageType) bool {
						t.Fatal("signed armor God entered HP tail")
						return false
					}
					h, result := PlayerDamageNative4E17B0(target, source, weapon, raw, typ, r)
					var want []string
					if observed {
						want = append(want, "observe")
					}
					want = append(want, "exclude", "direction", "scale", "armor lookup", "armor defend", "armor damage", "report", "God")
					residual := float32(0.5)
					if raw < 0 {
						residual = 0
					}
					if !h || !result || !slices.Equal(events, want) || target.HealthData.Cur != 60 || armor.HealthData.Cur != 98 || *carry != -0.25 || live.Field21 != math.Float32bits(residual) || cached.Field76 != 2 || cached.Field75 != uint32(typ) {
						t.Fatalf("signed armor=%t/%t events=%v carry=%g marker=%d/%d", h, result, events, *carry, cached.Field76, cached.Field75)
					}
				})
			}
		}
	})
}
