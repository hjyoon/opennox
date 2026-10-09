package legacy

import (
	"fmt"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func TestPlayerAttackNPC538960CompleteFrameCycle(t *testing.T) {
	for _, weaponKind := range []struct {
		name      string
		flag      object.WeaponClass
		class     object.Class
		animation int
	}{
		{"LongSword", object.WeaponSword, object.ClassWeapon, 27},
		{"WoodenStaff", object.WeaponStaff, object.ClassWand, 29},
	} {
		for _, duration := range []int{0, 2} {
			t.Run(fmt.Sprintf("%s/duration-%d", weaponKind.name, duration), func(t *testing.T) {
				srv := server.New(nil, nil, strman.New())
				t.Cleanup(srv.Close)
				srv.Map.Init()
				t.Cleanup(srv.Map.Free)
				bridge := &playerAttackLegacyServer538960{srv: srv}
				oldServer, oldFrames := GetServer, playerAnimFrames4F9F90
				GetServer = func() Server { return bridge }
				playerAnimFrames4F9F90 = func(animation int) (int, int) {
					if animation != weaponKind.animation {
						t.Fatalf("animation=%d want=%d", animation, weaponKind.animation)
					}
					return 6, duration
				}
				t.Cleanup(func() { GetServer, playerAnimFrames4F9F90 = oldServer, oldFrames })
				unit := &server.Object{ObjClass: object.ClassMonster, ObjSubClass: object.SubClass(object.MonsterNPC), PosVec: types.Ptf(321, 654)}
				unit.Shape.Kind, unit.Shape.Circle.R = server.ShapeKindCircle, 5
				update := &server.MonsterUpdateData{Field331: 37, WeaponEquipFlags: uint32(weaponKind.flag), Field481: 0xdeadbeef, Field517: 0x99aabbcc}
				weapon := &server.Object{TypeInd: 0x1234, ObjClass: weaponKind.class, ObjSubClass: object.SubClass(weaponKind.flag), ObjFlags: object.FlagEquipped, InvHolder: unit}
				modifier := &server.Modifier{TypeInd: uint32(weapon.TypeInd), ReqStrength60: 20, DamageCoeffOrArmor64: 1.5, Range68: 40, DamageMin72: 10}
				unit.UpdateData, unit.InvFirstItem = unsafe.Pointer(update), weapon
				srv.Modif.Dword_5d4594_251600 = modifier
				var pin runtime.Pinner
				for _, p := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(update), unsafe.Pointer(weapon), unsafe.Pointer(modifier)} {
					pin.Pin(p)
				}
				defer pin.Unpin()
				step := uint32(duration + 1)
				for cycle := uint32(0); cycle < 3; cycle++ {
					unit.Field34, update.Field120_1 = 100+cycle*100, 0
					beforeHits := bridge.wallDamageCalls
					for tick := uint32(0); tick <= 6*step; tick++ {
						srv.SetFrame(unit.Field34 + tick)
						wantFrame, wantResult := uint8(min(tick/step, 5)), 1
						if tick >= 6*step {
							wantResult = 0
						}
						// Repeated updates within one frame must not repeat the
						// midpoint hit; the byte also feeds NPC network packets.
						for repeat := 0; repeat < 2; repeat++ {
							got := playerAttackNativeEntry538960(unit)
							if got != wantResult || update.Field120_1 != wantFrame || update.Field481 != 0xdeadbeef || update.Field517 != 0x99aabbcc {
								t.Fatalf("cycle=%d tick=%d repeat=%d result=%d frame=%d unrelated=%#x animation=%#x want=%d/%d/unchanged", cycle, tick, repeat, got, update.Field120_1, update.Field481, update.Field517, wantResult, wantFrame)
							}
						}
					}
					if bridge.wallDamageCalls != beforeHits+1 || bridge.wallDamageAttacker != weapon {
						t.Fatalf("cycle %d hits=%d want exactly one", cycle, bridge.wallDamageCalls-beforeHits)
					}
				}
			})
		}
	}
}
