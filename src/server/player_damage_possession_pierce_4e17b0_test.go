package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestPlayerDamagePossessionPierce4E17B0(t *testing.T) {
	playerDamagePierceSources4E17B0(t, func(t *testing.T, playerSource bool) {
		for _, defense := range []string{"none", "Reflect Shield", "ordinary shield", "GreatSword", "excluded shield", "rear Reflect Shield"} {
			t.Run(defense, func(t *testing.T) {
				target, source, arrow, r, damages := playerDamagePierceFixture4E17B0(t, playerSource)
				cached := target.UpdateDataPlayer()
				cached.Player.Field3680, cached.Player.CameraFollowObj = 2, source
				cached.Field57, cached.Field21 = math.Float32bits(0.25), math.Float32bits(0.125)
				cached.Field76, cached.Field75 = 88, 77
				cached.State = PlayerState13
				if defense == "ordinary shield" || defense == "excluded shield" {
					cached.State, cached.Player.ArmorEquip = PlayerState16, 0x1000000
				}
				if defense == "GreatSword" {
					cached.Player.WeaponEquip = 0x400
				}
				livePlayer := &Player{Field3680: 2, CameraFollowObj: arrow}
				live := &PlayerUpdateData{Player: livePlayer, State: PlayerState13, Field57: math.Float32bits(0.5), Field21: math.Float32bits(0.5), Field76: 31, Field75: 33}
				previousPosition := arrow.PrevPos
				reflectShield := defense == "Reflect Shield" || defense == "rear Reflect Shield"
				var events []string
				r.ObserveClear = func(v *Object) {
					if v != target || target.UpdateDataPlayer() != cached || cached.Field76 != 0 || cached.Field75 != 77 || len(events) != 0 {
						t.Fatal("ObserveClear must follow only the original marker reset")
					}
					events = append(events, "observe")
					cached.Field57 = math.Float32bits(0.9)
					cached.Player.ArmorEquip, cached.Player.WeaponEquip = 0, 0
					target.UpdateData = unsafe.Pointer(live)
					// Original 004E18F3 queries Reflect Shield after ObserveClear.
					if reflectShield {
						target.Buffs |= 1 << playerDamageReflectEnchant4E17B0
					}
				}
				r.BlockSourceExcluded = func(w *Object) bool {
					before := []string{"observe"}
					if defense == "rear Reflect Shield" {
						before = append(before, "direction")
					}
					if w != arrow || !reflect.DeepEqual(events, before) || cached.Field76 != 0 || cached.Field75 != 77 {
						t.Fatal("exclusion must follow ObserveClear before missile attribution exactly once")
					}
					events = append(events, "exclude")
					arrow.TypeInd = 530
					// Original facing consumes the position snapshot taken before
					// this callback, not the current missile's new position.
					arrow.PrevPos = types.Ptf(-72, 87)
					return defense == "excluded shield"
				}
				r.BlockDirection = func(v *Object, pos types.Pointf) bool {
					if v != target || target.UpdateDataPlayer() != live {
						t.Fatal("direction preceded live update replacement")
					}
					if reflectShield && len(events) == 1 {
						if !reflect.DeepEqual(events, []string{"observe"}) || pos != arrow.PosVec || cached.Field76 != 0 {
							t.Fatal("Reflect facing/prefix order")
						}
					} else {
						before := []string{"observe", "exclude"}
						if defense == "rear Reflect Shield" {
							before = []string{"observe", "direction", "exclude"}
						}
						if !reflect.DeepEqual(events, before) || pos != previousPosition || cached.Field76 != 1 || cached.Field75 != 530 {
							t.Fatal("ordinary facing/prefix order, stale snapshot or duplicate call")
						}
					}
					events = append(events, "direction")
					return defense == "Reflect Shield" || defense == "ordinary shield" || defense == "GreatSword"
				}
				r.ProjectileReflect = func(w, v *Object) {
					if w != arrow || v != target || live.Field76 != 31 || live.Field75 != 33 {
						t.Fatal("reflection arguments/live marker")
					}
					events = append(events, "reflect")
				}
				r.ClearOwner = func(w *Object) {
					if w != arrow {
						t.Fatal("clear owner")
					}
					events = append(events, "clear")
				}
				r.SetOwner = func(v, w *Object) {
					if v != target || w != arrow {
						t.Fatal("set owner")
					}
					events = append(events, "owner")
				}
				r.Audio = func(id int, v *Object) {
					want := map[string]int{"Reflect Shield": 122, "ordinary shield": 878, "GreatSword": 890}[defense]
					if v != target || id != want {
						t.Fatal("defense audio")
					}
					events = append(events, "audio")
				}
				r.BlockDamagePercent = func() float64 { events = append(events, "balance"); return 0.5 }
				r.PlayerSetState = func(v *Object, state PlayerState) bool {
					if v != target || state != PlayerState19 {
						t.Fatal("GreatSword state")
					}
					events = append(events, "state")
					live.State = state
					return true
				}
				r.Melee.RandomInt = func(min, max int) int {
					if min != 18 || max != 20 {
						t.Fatal("GreatSword RNG")
					}
					events = append(events, "rng")
					return 19
				}
				blockItem := &Object{HealthData: &HealthData{Cur: 30}, ObjFlags: object.FlagEquipped}
				if defense == "ordinary shield" {
					blockItem.ObjClass, blockItem.ObjSubClass = object.ClassArmor, 2
					target.InvFirstItem = blockItem
				}
				if defense == "GreatSword" {
					blockItem.ObjClass, blockItem.ObjSubClass = object.ClassWeapon, 0x400
					target.InvFirstItem = blockItem
				}
				r.CanDamageBlockItem = func(item *Object) bool { return item == blockItem }
				r.Melee.CanDamageBlockWeapon = r.CanDamageBlockItem
				wear := func(item, v, a, w *Object, amount float32, typ object.DamageType) bool {
					if item != blockItem || v != target || a != source || w != arrow || amount != 2.5 || typ != object.DamageImpale || cached.Field76 != 1 || cached.Field75 != 530 {
						t.Fatal("block wear/cached marker")
					}
					events = append(events, "wear")
					return true
				}
				r.DamageBlockItem, r.Melee.DamageBlockWeapon = wear, wear
				tail := r.DefaultDamage
				r.DefaultDamage = func(v, a, w *Object, d int32, typ object.DamageType) bool {
					events = append(events, "default")
					return tail(v, a, w, d, typ)
				}
				handled, result := PlayerDamageNative4E17B0(target, source, arrow, 5, object.DamageImpale, r)
				want := []string{"observe", "exclude", "direction", "default"}
				if defense == "Reflect Shield" {
					want = []string{"observe", "direction", "reflect", "clear", "owner", "audio"}
				}
				if defense == "ordinary shield" {
					want = []string{"observe", "exclude", "direction", "audio", "balance", "wear"}
				}
				if defense == "GreatSword" {
					want = []string{"observe", "exclude", "direction", "reflect", "clear", "owner", "audio", "rng", "state", "balance", "wear"}
				}
				if defense == "excluded shield" {
					want = []string{"observe", "exclude", "default"}
				}
				if defense == "rear Reflect Shield" {
					want = []string{"observe", "direction", "exclude", "direction", "default"}
				}
				wantDamage := defense == "none" || defense == "excluded shield" || defense == "rear Reflect Shield"
				if !handled || result != wantDamage || !reflect.DeepEqual(events, want) || live.Field76 != 31 || live.Field75 != 33 || cached.Field21 != math.Float32bits(0.125) {
					t.Fatalf("possession=%t/%t events=%v want=%v", handled, result, events, want)
				}
				if wantDamage {
					if target.HealthData.Cur != 16 || !reflect.DeepEqual(*damages, []int32{4}) || live.Field21 != math.Float32bits(0.25) || cached.Field76 != 1 || cached.Field75 != 530 {
						t.Fatal("cached absorption/live carry/actual HP damage")
					}
				} else if target.HealthData.Cur != 20 || len(*damages) != 0 || live.Field21 != math.Float32bits(0.5) {
					t.Fatal("defense reached HP/carry")
				}
				if defense == "Reflect Shield" && (cached.Field76 != 0 || cached.Field75 != 77) {
					t.Fatal("Reflect changed cached attribution")
				}
			})
		}
	})
}

func TestPlayerDamagePossessionPierceAdmission4E17B0(t *testing.T) {
	for _, service := range []string{"observe", "direction", "exclusion", "default", "Quest scale"} {
		t.Run(service, func(t *testing.T) {
			target, source, arrow, r, damages := playerDamagePierceFixture4E17B0(t, false)
			ud := target.UpdateDataPlayer()
			ud.Player.Field3680, ud.Player.CameraFollowObj = 2, source
			ud.Field76, ud.Field75 = 88, 77
			r.ObserveClear = func(*Object) { t.Fatal("missing-service admission reached ObserveClear") }
			if service == "observe" {
				r.ObserveClear = nil
			}
			if service == "direction" {
				r.BlockDirection = nil
			}
			if service == "exclusion" {
				r.BlockSourceExcluded = nil
			}
			if service == "default" {
				r.DefaultDamage = nil
			}
			if service == "Quest scale" {
				r.QuestMode = func() bool { return true }
				r.QuestDamageScale = nil
			}
			var reason string
			r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
			before, beforeUD, beforePlayer := *target, *ud, *ud.Player
			if h, result := PlayerDamageNative4E17B0(target, source, arrow, 5, object.DamageImpale, r); h || result || reason == "" || *target != before || *ud != beforeUD || *ud.Player != beforePlayer || len(*damages) != 0 {
				t.Fatalf("admission=%t/%t reason=%q", h, result, reason)
			}
		})
	}
}

func TestPlayerDamagePossessionPierceEarlyGates4E17B0(t *testing.T) {
	playerDamagePierceSources4E17B0(t, func(t *testing.T, playerSource bool) {
		for _, gate := range []string{"no update", "dead", "invulnerable", "observer", "Coop self"} {
			t.Run(gate, func(t *testing.T) {
				target, source, arrow, r, damages := playerDamagePierceFixture4E17B0(t, playerSource)
				ud := target.UpdateDataPlayer()
				ud.Player.Field3680, ud.Player.CameraFollowObj = 2, arrow
				ud.Field76, ud.Field75 = 88, 77
				switch gate {
				case "no update":
					target.ObjFlags |= object.FlagNoUpdate
				case "dead":
					target.ObjFlags |= object.FlagDead
				case "invulnerable":
					target.Buffs |= 1 << playerDamageInvulnerableEnchant4E17B0
				case "observer":
					ud.Player.Field3680 |= 1
				case "Coop self":
					source, arrow.ObjOwner = target, target
					r.CoopMode = func() bool { return true }
				}
				r.ObserveClear = func(*Object) { t.Fatal("early gate cleared possession") }
				r.BlockSourceExcluded = func(*Object) bool { t.Fatal("early gate reached exclusion"); return false }
				r.BlockDirection = func(*Object, types.Pointf) bool { t.Fatal("early gate reached direction"); return false }
				audio := 0
				r.Audio = func(id int, v *Object) {
					if gate != "invulnerable" || v != target || id != 71 {
						t.Fatal("unexpected early audio")
					}
					audio++
				}
				before, beforeUD, beforePlayer, beforeArrow := *target, *ud, *ud.Player, *arrow
				h, result := PlayerDamageNative4E17B0(target, source, arrow, 5, object.DamageImpale, r)
				wantResult := gate == "invulnerable"
				if !h || result != wantResult || audio != map[bool]int{false: 0, true: 1}[wantResult] || *target != before || *ud != beforeUD || *ud.Player != beforePlayer || *arrow != beforeArrow || target.HealthData.Cur != 20 || len(*damages) != 0 {
					t.Fatalf("early gate=%t/%t audio=%d damage=%v", h, result, audio, *damages)
				}
			})
		}
	})
}

func TestPlayerDamagePossessionPierceSignedDamage4E17B0(t *testing.T) {
	playerDamagePierceSources4E17B0(t, func(t *testing.T, playerSource bool) {
		for _, tc := range []struct {
			name                 string
			damage, wantDamage   int32
			armor, carry, remain float32
		}{
			{"positive", 5, 4, 0.25, 0.5, 0.25},
			{"zero flushes live carry", 0, 1, 0.25, 0.75, -0.25},
			{"negative", -3, -2, 0.25, 0.25, 0},
			{"negative is not minimum one", -1, 0, 1, 0, 0},
		} {
			t.Run(tc.name, func(t *testing.T) {
				target, source, arrow, r, damages := playerDamagePierceFixture4E17B0(t, playerSource)
				cached := target.UpdateDataPlayer()
				cached.Player.Field3680, cached.Player.CameraFollowObj = 2, source
				cached.Field57, cached.Field21 = math.Float32bits(tc.armor), math.Float32bits(0.125)
				cached.Field76, cached.Field75 = 88, 77
				live := &PlayerUpdateData{Player: &Player{Field3680: 2, CameraFollowObj: arrow}, State: PlayerState13, Field57: math.Float32bits(0.5), Field21: math.Float32bits(tc.carry), Field76: 31, Field75: 33}
				var events []string
				r.ObserveClear = func(v *Object) {
					if v != target || cached.Field76 != 0 || cached.Field75 != 77 || len(events) != 0 {
						t.Fatal("signed possession prefix order")
					}
					events = append(events, "observe")
					cached.Field57 = math.Float32bits(0.9)
					target.UpdateData = unsafe.Pointer(live)
				}
				r.BlockSourceExcluded = func(w *Object) bool {
					if w != arrow || cached.Field76 != 0 || cached.Field75 != 77 {
						t.Fatal("signed possession marker order")
					}
					events = append(events, "exclude")
					return false
				}
				r.BlockDirection = func(*Object, types.Pointf) bool { events = append(events, "direction"); return false }
				h, result := PlayerDamageNative4E17B0(target, source, arrow, tc.damage, object.DamageImpale, r)
				if !h || !result || !reflect.DeepEqual(events, []string{"observe", "exclude", "direction"}) || !reflect.DeepEqual(*damages, []int32{tc.wantDamage}) || target.HealthData.Cur != uint16(20-tc.wantDamage) || cached.Field21 != math.Float32bits(0.125) || live.Field21 != math.Float32bits(tc.remain) || cached.Field76 != 1 || cached.Field75 != 529 || live.Field76 != 31 || live.Field75 != 33 {
					t.Fatalf("signed possession=%t/%t events=%v damage=%v HP=%d carry=%g", h, result, events, *damages, target.HealthData.Cur, math.Float32frombits(live.Field21))
				}
			})
		}
	})
}
