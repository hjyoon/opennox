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

func TestPlayerDamagePlayerPierceNativeCallback4E17B0(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	oldSetState := Nox_xxx_playerSetState_4FA020
	Nox_xxx_playerSetState_4FA020 = func(*server.Object, server.PlayerState) bool {
		t.Fatal("below-threshold PIERCE requested hurt state")
		return false
	}
	t.Cleanup(func() { Nox_xxx_playerSetState_4FA020 = oldSetState })
	for _, playerSource := range []bool{false, true} {
		for _, pure := range []bool{false, true} {
			t.Run(fmt.Sprintf("player-source-%t/pure-%t", playerSource, pure), func(t *testing.T) {
				var pin runtime.Pinner
				defer pin.Unpin()
				targetPlayer, sourcePlayer := &server.Player{PlayerInd: 7}, &server.Player{PlayerInd: 8}
				ud := &server.PlayerUpdateData{Player: targetPlayer, State: server.PlayerState13, Field57: math.Float32bits(0.25), Field40_0: 0x1234, Field40_1: 0xabcd}
				sourceUD := &server.MonsterUpdateData{}
				sourcePlayerUD := &server.PlayerUpdateData{Player: sourcePlayer, State: server.PlayerState13}
				target := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(ud), HealthData: &server.HealthData{Cur: 60, Max: 60}, Material: 0x4000, Damage: playerDamageMeleeCallbackNative4E17B0()}
				source := &server.Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(sourceUD), PrevPos: types.Ptf(44, 7)}
				if playerSource {
					source.ObjClass, source.UpdateData = object.ClassPlayer, unsafe.Pointer(sourcePlayerUD)
				}
				arrow := &server.Object{TypeInd: 529, ObjClass: object.Class(0x05200001), ObjSubClass: 0x10, PrevPos: types.Ptf(20, 0), ObjOwner: source}
				if pure {
					arrow.ObjClass = object.ClassMissile
				}
				for _, pointer := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(arrow), target.UpdateData, source.UpdateData, unsafe.Pointer(targetPlayer), unsafe.Pointer(sourcePlayer), unsafe.Pointer(sourcePlayerUD), unsafe.Pointer(target.HealthData)} {
					pin.Pin(pointer)
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
						t.Fatalf("pointer=%p, want above 4 GiB", pointer)
					}
				}
				beforeSource, beforePlayerUD, beforePlayer, beforeArrow := *source, *sourcePlayerUD, *sourcePlayer, *arrow
				for hit, carry := range []float32{0.25, 0.5, -0.25} {
					// Actual C dispatcher -> registered PlayerDamage -> DefaultDamage
					// -> UnitSetHP. No HP or damage callback is replaced.
					if !objectDamageDispatchCallNative(target, source, arrow, 3, object.DamageImpale) || target.HealthData.Cur != []uint16{58, 56, 53}[hit] ||
						ud.Field21 != math.Float32bits(carry) || ud.Field76 != 1 || ud.Field75 != 529 || ud.Field40_0 != 0x1234 || ud.Field40_1 != 0xabcd || ud.State != server.PlayerState13 ||
						target.Obj130 != arrow || target.Field131 != 3 || target.Frame134 != srv.Frame() || target.Field38 != math.MaxUint32 ||
						(playerSource && (*source != beforeSource || *sourcePlayerUD != beforePlayerUD || *sourcePlayer != beforePlayer)) || *arrow != beforeArrow {
						t.Fatalf("C entry hit=%d HP=%d marker=%d/%d carry=%g", hit, target.HealthData.Cur, ud.Field76, ud.Field75, math.Float32frombits(ud.Field21))
					}
				}
				t.Logf("C->player PlayerDamage: target=%p source=%p missile=%p HP=60->58->56->53", target, source, arrow)
				runtime.KeepAlive(target)
				runtime.KeepAlive(source)
				runtime.KeepAlive(arrow)
			})
		}
	}
}
