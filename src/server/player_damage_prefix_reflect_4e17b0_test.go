package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestPlayerDamagePiercePrefixReflect4E17B0(t *testing.T) {
	playerDamagePierceSources4E17B0(t, func(t *testing.T, playerSource bool) {
		for _, front := range []bool{false, true} {
			t.Run(map[bool]string{false: "rear", true: "front"}[front], func(t *testing.T) {
				target, source, arrow, r, damages := playerDamagePierceFixture4E17B0(t, playerSource)
				cached := target.UpdateDataPlayer()
				cached.Field76, cached.Field75 = 37, 41
				cached.Player.Field3680 = 2
				cached.Player.CameraFollowObj = source
				live := &PlayerUpdateData{Player: &Player{Field3680: 2, CameraFollowObj: arrow}, Field76: 51, Field75: 53}
				target.UpdateData = unsafe.Pointer(live)
				target.Buffs |= 1 << playerDamageReflectEnchant4E17B0
				beforeCached, beforeLive := *cached, *live
				r.playerPrefix = &playerDamagePrefix4E17B0{update: cached}
				r.ObserveClear = func(*Object) { t.Fatal("defense repeated the already-consumed possession prefix") }
				var events []string
				r.BlockDirection = func(v *Object, pos types.Pointf) bool {
					if v != target || pos != arrow.PosVec {
						t.Fatal("Reflect Shield facing arguments")
					}
					events = append(events, "direction")
					return front
				}
				r.ProjectileReflect = func(w, v *Object) {
					if w != arrow || v != target {
						t.Fatal("reflect arguments")
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
					if id != 122 || v != target {
						t.Fatal("reflect audio")
					}
					events = append(events, "audio")
				}
				applicable, handled, result := playerDamageReflectShield4E17B0(target, source, arrow, 8, object.DamageImpale, r)
				if applicable != front || handled != front || result {
					t.Fatalf("reflect = %t/%t/%t front=%t", applicable, handled, result, front)
				}
				want := []string{"direction"}
				if front {
					want = append(want, "reflect", "clear", "owner", "audio")
				}
				if !reflect.DeepEqual(events, want) || *cached != beforeCached || *live != beforeLive || target.HealthData.Cur != 20 || len(*damages) != 0 {
					t.Fatalf("events=%v want=%v cached/live marker=%d/%d", events, want, cached.Field76, live.Field76)
				}
			})
		}
	})
}
