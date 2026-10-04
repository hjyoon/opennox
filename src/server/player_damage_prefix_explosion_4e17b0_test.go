package server

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestPlayerDamageExplosionPrefixLiveTail4E17B0(t *testing.T) {
	for _, playerSource := range []bool{false, true} {
		for _, splash := range []bool{false, true} {
			for _, markerAction := range []string{"retain", "clear", "replace"} {
				t.Run(fmt.Sprintf("player-source-%t/splash-%t/%s", playerSource, splash, markerAction), func(t *testing.T) {
					target, source, weapon, missile := damageExplosionFixture4E17B0(t, playerSource, true, splash)
					cached := target.UpdateDataPlayer()
					cached.Player.Field3680, cached.Player.CameraFollowObj = 2, source
					cached.Field57, cached.Field21 = math.Float32bits(0.9), math.Float32bits(0.125)
					cached.Field76, cached.Field75 = 1, 777
					if splash {
						cached.Field76, cached.Field75 = 0, 77
					}
					marker, markerType := cached.Field76, cached.Field75
					live := &PlayerUpdateData{Player: &Player{Field3680: 2, CameraFollowObj: missile}, Field57: math.Float32bits(0.5), Field21: math.Float32bits(0.5), Field76: 31, Field75: 33}
					target.UpdateData = unsafe.Pointer(live)
					carry := damageArmorCarryFixture4E17B0(0.5)
					armor := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, Damage: unsafe.Pointer(new(byte)), HealthData: &HealthData{Cur: 30}, UpdateData: unsafe.Pointer(carry), InitData: unsafe.Pointer(&ModifierInitData{})}
					target.InvFirstItem = armor
					r := damageFlameRuntime4E17B0(t, 0)
					r.playerPrefix = &playerDamagePrefix4E17B0{update: cached}
					r.ObserveClear = func(*Object) { t.Fatal("EXPLOSION tail repeated ObserveClear") }
					r.ItemArmorValue = func(*Object) float32 { return 0.25 }
					r.CanDamageArmor = func(item *Object) bool { return item == armor }
					quest := false
					var events []string
					r.DamageArmor = func(item, a, w *Object, damage int32, typ object.DamageType) bool {
						if item != armor || a != source || w != weapon || damage != 1 || typ != object.DamageExplosion || cached.Field76 != marker || cached.Field75 != markerType || live.Field21 != math.Float32bits(0.25) {
							t.Fatalf("EXPLOSION wear/carry/marker order: damage=%d marker=%d/%d want=%d/%d carry=%g", damage, cached.Field76, cached.Field75, marker, markerType, math.Float32frombits(live.Field21))
						}
						events = append(events, "armor")
						item.HealthData.Cur -= uint16(damage)
						quest = true
						if markerAction == "clear" {
							cached.Field76 = 0
						} else if markerAction == "replace" {
							cached.Field76, cached.Field75 = 81, 82
						}
						return true
					}
					r.GodMode = func() bool { events = append(events, "god"); return false }
					r.QuestMode = func() bool { events = append(events, "quest"); return quest }
					r.QuestDamageScale = func() float32 { events = append(events, "scale"); return 0.5 }
					tail := r.DefaultDamage
					r.DefaultDamage = func(v, a, w *Object, damage int32, typ object.DamageType) bool {
						if damage != 2 {
							t.Fatalf("EXPLOSION Quest damage=%d", damage)
						}
						events = append(events, "default")
						return tail(v, a, w, damage, typ)
					}
					h, result := playerDamageMissileExplosionTail4E17B0(target, source, weapon, &cached.Field76, &cached.Field75, 0.25, 5, object.DamageExplosion, r)
					wantMarker, wantType := marker, markerType
					if markerAction == "replace" {
						wantMarker, wantType = 81, 82
					} else if markerAction == "clear" || splash {
						wantMarker, wantType = 2, uint32(object.DamageExplosion)
					}
					if !h || !result || !slices.Equal(events, []string{"armor", "god", "quest", "scale", "default"}) || target.HealthData.Cur != 198 || armor.HealthData.Cur != 29 || *carry != 0 || cached.Field76 != wantMarker || cached.Field75 != wantType || cached.Field21 != math.Float32bits(0.125) || live.Field21 != math.Float32bits(0.25) || live.Field76 != 31 || live.Field75 != 33 {
						t.Fatalf("EXPLOSION tail=%t/%t HP=%d armor=%d events=%v marker=%d/%d", h, result, target.HealthData.Cur, armor.HealthData.Cur, events, cached.Field76, cached.Field75)
					}
				})
			}
		}
	}
}
