package legacy

import (
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

func TestPlayerAttackBow538960NativePlayerRoute(t *testing.T) {
	for _, flag := range []object.WeaponClass{object.WeaponBow, object.WeaponCrossbow} {
		t.Run(flag.String(), func(t *testing.T) {
			srv := installPlayerAttackProjectileServer538960(t)
			owner := &server.Object{ObjClass: object.ClassPlayer}
			weapon := &server.Object{TypeInd: 0x3214, ObjClass: object.ClassWeapon, ObjSubClass: object.SubClass(flag)}
			player := &server.Player{WeaponEquip: uint32(flag)}
			player.Info().SetField2239(37)
			update := &server.PlayerUpdateData{Player: player, EquippedWeapon: weapon, Field59_1: 0x65, Field59_2: 0x9876}
			owner.UpdateData = unsafe.Pointer(update)
			modifier := &server.Modifier{TypeInd: uint32(weapon.TypeInd)}
			srv.Modif.Dword_5d4594_251600 = modifier
			var pin runtime.Pinner
			defer pin.Unpin()
			pinPlayerAttackProjectilePointers538960(t, &pin, unsafe.Pointer(owner), unsafe.Pointer(weapon), unsafe.Pointer(player), unsafe.Pointer(update), unsafe.Pointer(modifier))
			old := playerAttackBowRuntimeFactory538960
			playerAttackBowRuntimeFactory538960 = func() playerAttackBowRuntime538960 {
				return playerAttackBowRuntime538960{playerAttackProjectileRuntime538960: playerAttackProjectileRuntime538960{
					Frame: srv.Frame, Readiness: func(*server.Object) int32 { return 0 },
					AnimFrames: func(action int) (int, int) {
						if flag == object.WeaponBow && action != 33 || flag == object.WeaponCrossbow && action != 34 {
							t.Fatalf("bow animation %d", action)
						}
						return 4, 0
					},
				}}
			}
			defer func() { playerAttackBowRuntimeFactory538960 = old }()
			if got := playerAttackNativeEntry538960(owner); got != 1 {
				t.Fatalf("native player bow start = %d, want 1", got)
			}
			if update.Field0 != 4 || update.Field59_0 != 0 || update.Field59_1 != 0x65 || update.Field59_2 != 0x9876 {
				t.Fatalf("bow deadline/frame/neighbors = %+v", *update)
			}
		})
	}
}
