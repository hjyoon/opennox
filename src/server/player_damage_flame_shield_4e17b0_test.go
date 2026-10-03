package server

import (
	"fmt"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestPlayerDamageNative4E17B0MissileFlameShieldMarker(t *testing.T) {
	for _, splash := range []bool{false, true} {
		t.Run(fmt.Sprintf("splash-%t", splash), func(t *testing.T) {
			target, source, weapon, missile := damageFlameFixture4E17B0(t, false, true, splash)
			ud := target.UpdateDataPlayer()
			ud.Field76, ud.Field75 = 99, 77
			ud.State, ud.Player.ArmorEquip = PlayerState16, 0x1000000
			shield := &Object{ObjClass: object.ClassArmor, ObjSubClass: 2, ObjFlags: object.FlagEquipped, HealthData: &HealthData{Cur: 10}}
			target.InvFirstItem = shield
			r := damageFlameRuntime4E17B0(t, 0)
			r.BlockDirection = func(got *Object, p types.Pointf) bool { return got == target && p == missile.PrevPos }
			r.BlockSourceExcluded = func(*Object) bool { return false }
			r.ProjectileReflect = func(got, owner *Object) {
				if got != missile || owner != target {
					t.Fatal("wrong blocked projectile")
				}
			}
			r.ClearOwner = func(got *Object) { got.ObjOwner = nil }
			r.SetOwner = func(owner, got *Object) { got.ObjOwner = owner }
			wantMarker, wantKind := uint32(1), uint32(missile.TypeInd)
			if splash {
				wantMarker, wantKind = 0, 77
			}
			r.Audio = func(id int, got *Object) {
				if id != 878 || got != target || ud.Field76 != wantMarker || ud.Field75 != wantKind {
					t.Fatal("shield attribution/audio order")
				}
			}
			r.BlockDamagePercent = func() float64 { return 0.5 }
			r.CanDamageBlockItem = func(got *Object) bool { return got == shield }
			r.PlayerSetState = func(*Object, PlayerState) bool { t.Fatal("intact shield changed state"); return false }
			r.DamageBlockItem = func(got, owner, s, w *Object, amount float32, typ object.DamageType) bool {
				if got != shield || owner != target || s != source || w != missile || amount != 4.5 || typ != object.DamageFlame {
					t.Fatal("wrong shield wear arguments")
				}
				shield.HealthData.Cur -= 4
				return true
			}
			if h, result := PlayerDamageNative4E17B0(target, source, weapon, 9, object.DamageFlame, r); !h || result || target.HealthData.Cur != 200 ||
				shield.HealthData.Cur != 6 || target.Obj130 != nil || ud.Field76 != wantMarker || ud.Field75 != wantKind {
				t.Fatal("shield block reached FLAME HP/armor pass")
			}
		})
	}
}
