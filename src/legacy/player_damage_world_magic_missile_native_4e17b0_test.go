package legacy

import (
	"fmt"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func TestPlayerDamageWorldMagicMissileNativeCallback4E17B0(t *testing.T) {
	_ = npcReflectLegacyServer4E17B0(t)
	oldState := Nox_xxx_playerSetState_4FA020
	Nox_xxx_playerSetState_4FA020 = func(*server.Object, server.PlayerState) bool {
		t.Fatal("below-threshold explosion requested a hurt state")
		return false
	}
	t.Cleanup(func() { Nox_xxx_playerSetState_4FA020 = oldState })
	for _, owner := range []string{"terminal", "immobile", "imaginary"} {
		for _, player := range []bool{false, true} {
			for _, immune := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/player-%t/immune-%t", owner, player, immune), func(t *testing.T) {
					var pin runtime.Pinner
					defer pin.Unpin()
					npc, hero := npcReflectLegacyObjects4E17B0(t, &pin)
					npc.Buffs = 0
					target := npc
					if player {
						target = hero
						target.Damage = playerDamageMeleeCallbackNative4E17B0()
					}
					if immune {
						target.ObjSubClass |= 0x400
					}
					missile := &server.Object{TypeInd: 697, ObjClass: object.ClassMissile | object.ClassSimple | object.ClassLight,
						PrevPos: types.Ptf(23, 9), PosVec: types.Ptf(25, 11)}
					pin.Pin(missile)
					source := missile
					if owner != "terminal" {
						source = &server.Object{TypeInd: 1399}
						if owner == "immobile" {
							source.ObjClass = object.ClassImmobile
						}
						pin.Pin(source)
						missile.ObjOwner = source
					}
					if missile.FindOwnerChainPlayer() != source || source.UpdateData != nil ||
						(unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(missile)) <= math.MaxUint32) {
						t.Fatal("expected a native-width missile with a terminal non-unit owner")
					}
					var marker, kind, carry *uint32
					if player {
						ud := target.UpdateDataPlayer()
						ud.Field57 = math.Float32bits(.5)
						marker, kind, carry = &ud.Field76, &ud.Field75, &ud.Field21
					} else {
						ud := target.UpdateDataMonster()
						ud.Field518 = math.Float32bits(.5)
						marker, kind, carry = &ud.Field547, &ud.Field546, &ud.Field1
					}
					*marker, *kind, *carry = 99, 77, math.Float32bits(.25)
					wantHP := []uint16{195, 191, 186}
					if immune && !player {
						wantHP = []uint16{198, 196, 194}
					}
					for hit, hp := range wantHP {
						if !objectDamageDispatchCallNative(target, source, missile, 9, object.DamageExplosion) || target.HealthData.Cur != hp {
							t.Fatalf("C->PlayerDamage->DefaultDamage->UnitSetHP hit=%d HP=%d want=%d", hit, target.HealthData.Cur, hp)
						}
						wantMarker, wantKind := uint32(1), uint32(missile.TypeInd)
						if owner == "terminal" {
							wantMarker, wantKind = 2, 7
						}
						wantCarry := []float32{-.25, .25, -.25}[hit]
						if *marker != wantMarker || *kind != wantKind || *carry != math.Float32bits(wantCarry) ||
							target.Field38 != math.MaxUint32 || target.Obj130 != missile || target.Pos132 != missile.PrevPos || target.Field131 != 7 || target.Frame134 != 1400 {
							t.Fatalf("marker=%d/%d carry=%x attribution=%p sync=%x", *marker, *kind, *carry, target.Obj130, target.Field38)
						}
					}
					runtime.KeepAlive(npc)
					runtime.KeepAlive(hero)
					runtime.KeepAlive(source)
					runtime.KeepAlive(missile)
				})
			}
		}
	}
}
