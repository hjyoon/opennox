package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestPlayerDamagePiercePrefixLiveTail4E17B0(t *testing.T) {
	playerDamagePierceSources4E17B0(t, func(t *testing.T, playerSource bool) {
		for _, clearMarker := range []bool{false, true} {
			t.Run(map[bool]string{false: "retain missile marker", true: "armor clears marker before fallback"}[clearMarker], func(t *testing.T) {
				target, source, arrow, r, damages := playerDamagePierceFixture4E17B0(t, playerSource)
				cached := target.UpdateDataPlayer()
				cached.Player.Field3680, cached.Player.CameraFollowObj = 2, source
				cached.Field57, cached.Field21 = math.Float32bits(0.9), math.Float32bits(0.125)
				cached.Field76, cached.Field75 = 1, uint32(arrow.TypeInd)
				live := &PlayerUpdateData{Player: &Player{}, Field57: math.Float32bits(0.5), Field21: math.Float32bits(0.5), Field76: 31, Field75: 33}
				target.UpdateData = unsafe.Pointer(live)
				r.playerPrefix = &playerDamagePrefix4E17B0{update: cached}
				r.ObserveClear = func(*Object) { t.Fatal("PIERCE tail repeated ObserveClear") }
				carry := damageArmorCarryFixture4E17B0(0.5)
				armor := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, Damage: unsafe.Pointer(new(byte)), HealthData: &HealthData{Cur: 30}, UpdateData: unsafe.Pointer(carry), InitData: unsafe.Pointer(&ModifierInitData{})}
				target.InvFirstItem = armor
				quest := false
				var events []string
				r.ItemArmorValue = func(*Object) float32 { return 0.25 }
				r.CanDamageArmor = func(item *Object) bool { return item == armor }
				r.DamageArmor = func(item, a, w *Object, d int32, typ object.DamageType) bool {
					if item != armor || a != source || w != arrow || d != 1 || typ != object.DamageImpale || cached.Field76 != 1 || cached.Field75 != 529 || live.Field21 != math.Float32bits(0.25) {
						t.Fatal("PIERCE armor/live carry order")
					}
					events = append(events, "armor")
					item.HealthData.Cur--
					quest = true
					if clearMarker {
						cached.Field76 = 0
					}
					return true
				}
				r.GodMode = func() bool { events = append(events, "god"); return false }
				r.QuestMode = func() bool { events = append(events, "quest"); return quest }
				r.QuestDamageScale = func() float32 { events = append(events, "scale"); return 0.5 }
				tail := r.DefaultDamage
				r.DefaultDamage = func(v, a, w *Object, d int32, typ object.DamageType) bool {
					events = append(events, "default")
					return tail(v, a, w, d, typ)
				}
				handled, result := playerDamageMissilePierce4E17B0(target, source, arrow, cached, 0.25, 5, object.DamageImpale, r)
				wantMarker, wantType := uint32(1), uint32(529)
				if clearMarker {
					wantMarker, wantType = 2, uint32(object.DamageImpale)
				}
				if !handled || !result || !reflect.DeepEqual(events, []string{"armor", "god", "quest", "scale", "default"}) || !reflect.DeepEqual(*damages, []int32{2}) || target.HealthData.Cur != 18 || armor.HealthData.Cur != 29 || *carry != 0 || cached.Field21 != math.Float32bits(0.125) || cached.Field57 != math.Float32bits(0.9) || cached.Field76 != wantMarker || cached.Field75 != wantType || live.Field76 != 31 || live.Field75 != 33 || live.Field21 != math.Float32bits(0.25) {
					t.Fatalf("tail=%t/%t events=%v damage=%v marker=%d/%d live carry=%g armor carry=%g", handled, result, events, *damages, cached.Field76, cached.Field75, math.Float32frombits(live.Field21), *carry)
				}
			})
		}
	})
}
