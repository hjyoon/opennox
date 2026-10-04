package server

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func electricPlayerCases4E17B0(t *testing.T, fn func(*testing.T, bool, bool, object.DamageType)) {
	t.Helper()
	for _, playerSource := range []bool{false, true} {
		for _, selfWeapon := range []bool{false, true} {
			for _, typ := range []object.DamageType{object.DamageElectric, object.DamageAirborneElectric} {
				t.Run(fmt.Sprintf("player-source-%t/self-weapon-%t/%s", playerSource, selfWeapon, typ), func(t *testing.T) {
					fn(t, playerSource, selfWeapon, typ)
				})
			}
		}
	}
}

// 004E1DF2 calls ElectricArmorScale before 004E20F0 reads the live carry.
// The marker base remains the update captured before that call instead.
func TestPlayerDamageElectricPlayerLiveCarry4E17B0(t *testing.T) {
	electricPlayerCases4E17B0(t, func(t *testing.T, playerSource, selfWeapon bool, typ object.DamageType) {
		for _, prefix := range []bool{false, true} {
			t.Run(fmt.Sprintf("consumed-prefix-%t", prefix), func(t *testing.T) {
				target, source := defaultDamageElectricSelfFixture4E0B30(t, true, playerSource)
				weapon := source
				if !selfWeapon {
					weapon = nil
				}
				cached := target.UpdateDataPlayer()
				cached.Field21, cached.Field76, cached.Field75 = math.Float32bits(0.125), 99, 77
				live := &PlayerUpdateData{Player: &Player{}, Field21: math.Float32bits(0.5), Field76: 31, Field75: 33}
				var events []string
				r := PlayerDamageRuntime4E17B0{Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) }}
				if prefix {
					cached.Field76, cached.Field75 = 9, 321
					r.playerPrefix = &playerDamagePrefix4E17B0{update: cached}
					target.UpdateData = unsafe.Pointer(live)
					// This is live observer data, not an unconsumed entry gate.
					live.Player.Field3680, live.Player.CameraFollowObj = 3, source
				}
				r.ObserveClear = func(*Object) { t.Fatal("consumed prefix ran ObserveClear twice") }
				r.ElectricArmorScale = func(v *Object) float32 {
					wantMarker := uint32(0)
					if prefix {
						wantMarker = 9
					}
					if v != target || cached.Field76 != wantMarker {
						t.Fatal("scale ran before the marker reset, or reset a consumed marker")
					}
					events = append(events, "electric scale")
					target.UpdateData = unsafe.Pointer(live)
					return 0.5
				}
				r.GodMode = func() bool { events = append(events, "God"); return false }
				r.QuestMode = func() bool { events = append(events, "Quest"); return false }
				r.QuestDamageScale = func() float32 { t.Fatal("non-Quest scale"); return 1 }
				r.DefaultDamage = func(v, a, w *Object, d int32, gotType object.DamageType) bool {
					if v != target || a != source || w != weapon || d != 3 || gotType != typ {
						t.Fatal("electric default lost its live carry or original weapon")
					}
					events = append(events, "default")
					return true
				}
				h, result := playerDamageElectricPlayer4E17B0(target, source, weapon, 5, typ, r)
				marker, markerType := uint32(2), uint32(typ)
				if prefix {
					marker, markerType = 9, 321
				}
				if !h || !result || !slices.Equal(events, []string{"electric scale", "God", "Quest", "default"}) || cached.Field21 != math.Float32bits(0.125) || cached.Field76 != marker || cached.Field75 != markerType || live.Field21 != 0 || live.Field76 != 31 || live.Field75 != 33 {
					t.Fatalf("electric carry=%t/%t events=%v cached=%d/%d/%g live=%d/%d/%g", h, result, events, cached.Field76, cached.Field75, math.Float32frombits(cached.Field21), live.Field76, live.Field75, math.Float32frombits(live.Field21))
				}
			})
		}
	})
}

func TestPlayerDamageElectricPlayerLiveTail4E17B0(t *testing.T) {
	electricPlayerCases4E17B0(t, func(t *testing.T, playerSource, selfWeapon bool, typ object.DamageType) {
		for _, tc := range []struct {
			name              string
			clear, quest, god bool
			dropPlayer        string
		}{
			{"keep callback marker", false, false, false, ""},
			{"fallback cleared marker", true, false, false, ""},
			{"Quest enabled by wear", false, true, false, ""},
			{"God after wear", false, true, true, ""},
			{"wear changes live class", true, true, true, "wear"},
			{"God flag precedes live class", true, true, true, "God"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				target, source, armor, _, modifier := playerDamageElectricSelfArmorFixture4E17B0(t, true, playerSource)
				weapon := source
				if !selfWeapon {
					weapon = nil
				}
				cached := target.UpdateDataPlayer()
				cached.Field76, cached.Field75, cached.Field21 = 9, 321, math.Float32bits(0.125)
				live := &PlayerUpdateData{Player: &Player{}, Field57: math.Float32bits(0.4), Field21: math.Float32bits(0.5), Field76: 31, Field75: 33}
				target.UpdateData = unsafe.Pointer(live)
				var events []string
				quest := false
				r := PlayerDamageRuntime4E17B0{playerPrefix: &playerDamagePrefix4E17B0{update: cached}, Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) }}
				r.ElectricArmorScale = func(*Object) float32 { events = append(events, "electric scale"); return 0.5 }
				r.ItemArmorValue = func(item *Object) float32 {
					if item != armor || live.Field21 != 0 || cached.Field76 != 9 {
						t.Fatal("armor lookup lost live carry or cached marker")
					}
					events = append(events, "lookup")
					return 0.4
				}
				r.CanApplyArmorDefend = func(m *ModifierEff) bool { return m == modifier }
				r.ApplyArmorDefend = func(m *ModifierEff, item, v, w, a *Object, d *float32) bool {
					if m != modifier || item != armor || v != target || w != weapon || a != source || *d != 5 || cached.Field76 != 9 {
						t.Fatal("electric wear did not receive raw damage/original weapon")
					}
					events = append(events, "defend")
					quest = tc.quest
					if tc.clear {
						cached.Field76, cached.Field75 = 0, 99
					}
					return true
				}
				r.CanDamageArmor = func(item *Object) bool { return item == armor }
				r.DamageArmor = func(item, a, w *Object, d int32, gotType object.DamageType) bool {
					if item != armor || a != source || w != weapon || d != 5 || gotType != typ {
						t.Fatal("electric armor damage arguments")
					}
					events = append(events, "armor damage")
					if tc.dropPlayer == "wear" {
						target.ObjClass = object.ClassSimple
					}
					return true
				}
				r.GodMode = func() bool {
					events = append(events, "God")
					if tc.dropPlayer == "God" {
						target.ObjClass = object.ClassSimple
					}
					return tc.god
				}
				r.QuestMode = func() bool { events = append(events, "Quest"); return quest }
				r.QuestDamageScale = func() float32 { events = append(events, "Quest scale"); return 0.5 }
				god := tc.god && tc.dropPlayer == ""
				wantDamage := int32(3)
				if tc.quest {
					wantDamage = 2
				}
				r.DefaultDamage = func(v, a, w *Object, d int32, gotType object.DamageType) bool {
					if v != target || a != source || w != weapon || d != wantDamage || gotType != typ {
						t.Fatal("electric live damage tail")
					}
					events = append(events, "default")
					return true
				}
				h, result := playerDamageElectricPlayer4E17B0(target, source, weapon, 5, typ, r)
				want := []string{"electric scale", "lookup", "defend", "armor damage", "God"}
				if !god {
					want = append(want, "Quest")
					if tc.quest {
						want = append(want, "Quest scale")
					}
					want = append(want, "default")
				}
				marker, markerType := uint32(9), uint32(321)
				if tc.clear {
					marker, markerType = 2, uint32(typ)
					if tc.dropPlayer == "wear" {
						marker, markerType = 0, 99
					}
				}
				if !h || !result || !slices.Equal(events, want) || cached.Field76 != marker || cached.Field75 != markerType || cached.Field21 != math.Float32bits(0.125) || live.Field21 != 0 || live.Field76 != 31 || live.Field75 != 33 {
					t.Fatalf("electric tail=%t/%t marker=%d/%d events=%v want=%v", h, result, cached.Field76, cached.Field75, events, want)
				}
			})
		}
	})
}

func TestPlayerDamageElectricPlayerInvalidLiveCarry4E17B0(t *testing.T) {
	for _, change := range []string{"class", "nil update"} {
		t.Run(change, func(t *testing.T) {
			target, source := defaultDamageElectricSelfFixture4E0B30(t, true, false)
			cached := target.UpdateDataPlayer()
			cached.Field21, cached.Field76, cached.Field75 = math.Float32bits(0.125), 99, 77
			reason := ""
			r := PlayerDamageRuntime4E17B0{
				ElectricArmorScale: func(*Object) float32 {
					if change == "class" {
						target.ObjClass = object.ClassMonster
					} else {
						target.UpdateData = nil
					}
					return 1
				},
				DefaultDamage: func(*Object, *Object, *Object, int32, object.DamageType) bool {
					t.Fatal("invalid carry entered default damage")
					return false
				},
				Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why },
			}
			if h, result := playerDamageElectricPlayer4E17B0(target, source, nil, 5, object.DamageElectric, r); h || result || reason == "" || cached.Field21 != math.Float32bits(0.125) || cached.Field76 != 0 || cached.Field75 != 77 || target.HealthData.Cur != 60 {
				t.Fatalf("live electric guard=%t/%t reason=%q carry=%g", h, result, reason, math.Float32frombits(cached.Field21))
			}
		})
	}
}
