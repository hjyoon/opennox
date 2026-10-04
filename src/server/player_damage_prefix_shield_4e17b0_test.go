package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestPlayerDamagePiercePrefixShield4E17B0(t *testing.T) {
	playerDamagePierceSources4E17B0(t, func(t *testing.T, playerSource bool) {
		for _, cachedShield := range []bool{false, true} {
			t.Run(map[bool]string{false: "live shield cannot add cached equipment", true: "cached shield survives player replacement"}[cachedShield], func(t *testing.T) {
				target, source, arrow, r, damages := playerDamagePierceFixture4E17B0(t, playerSource)
				cached := target.UpdateDataPlayer()
				cached.State, cached.Field76, cached.Field75 = PlayerState16, 1, uint32(arrow.TypeInd)
				cached.Player.Field3680, cached.Player.CameraFollowObj = 2, source
				live := &PlayerUpdateData{Player: &Player{Field3680: 2, CameraFollowObj: arrow}, Field76: 31, Field75: 33, State: PlayerState13}
				prefix := &playerDamagePrefix4E17B0{update: cached}
				if cachedShield {
					prefix.armorFlags = 0x1000000
				} else {
					live.State, live.Player.ArmorEquip = PlayerState16, 0x1000000
				}
				// The cached equipment mask remains independent of both players.
				cached.Player.ArmorEquip = 0
				target.UpdateData = unsafe.Pointer(live)
				r.playerPrefix = prefix
				r.ObserveClear = func(*Object) { t.Fatal("shield repeated ObserveClear") }
				shield := &Object{ObjClass: object.ClassArmor, ObjSubClass: 2, ObjFlags: object.FlagEquipped, HealthData: &HealthData{Cur: 30}}
				target.InvFirstItem = shield
				beforeCached, beforeLive := *cached, *live
				var events []string
				r.BlockSourceExcluded = func(w *Object) bool {
					if w != arrow || cached.Field76 != 1 || cached.Field75 != 529 {
						t.Fatal("exclusion preceded cached missile attribution")
					}
					events = append(events, "exclude")
					arrow.TypeInd = 530
					return false
				}
				r.BlockDirection = func(v *Object, pos types.Pointf) bool {
					if v != target || pos != arrow.PrevPos {
						t.Fatal("shield facing arguments")
					}
					events = append(events, "direction")
					return true
				}
				r.Audio = func(id int, v *Object) {
					if id != 878 || v != target || *cached != beforeCached || *live != beforeLive {
						t.Fatal("shield rewrote consumed prefix/live marker")
					}
					events = append(events, "audio")
				}
				r.BlockDamagePercent = func() float64 { events = append(events, "balance"); return 0.5 }
				r.CanDamageBlockItem = func(item *Object) bool { return item == shield }
				r.PlayerSetState = func(*Object, PlayerState) bool { t.Fatal("intact shield changed state"); return false }
				r.DamageBlockItem = func(item, v, a, w *Object, amount float32, typ object.DamageType) bool {
					if item != shield || v != target || a != source || w != arrow || amount != 4 || typ != object.DamageImpale || *cached != beforeCached || *live != beforeLive {
						t.Fatal("shield durability/cached marker")
					}
					events = append(events, "wear")
					item.HealthData.Cur -= 4
					return true
				}
				applicable, handled, result := playerDamageShieldBlock4E17B0(target, source, arrow, 8, object.DamageImpale, r)
				if applicable != cachedShield || handled != cachedShield || result {
					t.Fatalf("shield = %t/%t/%t cached=%t", applicable, handled, result, cachedShield)
				}
				want := []string(nil)
				wantHP := uint16(30)
				if cachedShield {
					want, wantHP = []string{"exclude", "direction", "audio", "balance", "wear"}, 26
				}
				if !reflect.DeepEqual(events, want) || *cached != beforeCached || *live != beforeLive || shield.HealthData.Cur != wantHP || target.HealthData.Cur != 20 || len(*damages) != 0 {
					t.Fatalf("events=%v want=%v shieldHP=%d", events, want, shield.HealthData.Cur)
				}
			})
		}
	})
}
