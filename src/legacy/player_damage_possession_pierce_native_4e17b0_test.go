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

func TestPlayerDamagePossessionPierceNativeCallback4E17B0(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	oldObserve, oldSetState := Nox_xxx_playerObserveClear_4DDEF0, Nox_xxx_playerSetState_4FA020
	t.Cleanup(func() { Nox_xxx_playerObserveClear_4DDEF0, Nox_xxx_playerSetState_4FA020 = oldObserve, oldSetState })
	Nox_xxx_playerSetState_4FA020 = func(*server.Object, server.PlayerState) bool {
		t.Fatal("below-threshold PIERCE hurt state")
		return false
	}
	for _, playerSource := range []bool{false, true} {
		for _, pure := range []bool{false, true} {
			t.Run(fmt.Sprintf("player-source-%t/pure-%t", playerSource, pure), func(t *testing.T) {
				var pin runtime.Pinner
				defer pin.Unpin()
				targetPlayer, newPlayer, sourcePlayer := &server.Player{PlayerInd: 7, Field3680: 2}, &server.Player{PlayerInd: 7, Field3680: 2}, &server.Player{PlayerInd: 8}
				cached := &server.PlayerUpdateData{Player: targetPlayer, State: server.PlayerState13, Field57: math.Float32bits(0.25), Field21: math.Float32bits(0.125), Field76: 88, Field75: 77}
				live := &server.PlayerUpdateData{Player: newPlayer, State: server.PlayerState13, Field57: math.Float32bits(0.5), Field21: math.Float32bits(0.5), Field76: 31, Field75: 33}
				sourceUD, sourcePlayerUD := &server.MonsterUpdateData{}, &server.PlayerUpdateData{Player: sourcePlayer, State: server.PlayerState13}
				target := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(cached), HealthData: &server.HealthData{Cur: 60, Max: 60}, Material: 0x4000, Damage: playerDamageMeleeCallbackNative4E17B0()}
				source := &server.Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(sourceUD), PrevPos: types.Ptf(44, 7)}
				if playerSource {
					source.ObjClass, source.UpdateData = object.ClassPlayer, unsafe.Pointer(sourcePlayerUD)
				}
				arrow := &server.Object{TypeInd: 529, ObjClass: object.Class(0x05200001), ObjSubClass: 0x10, PrevPos: types.Ptf(20, 0), ObjOwner: source, InitData: unsafe.Pointer(&server.ModifierInitData{})}
				if pure {
					arrow.ObjClass = object.ClassMissile
				}
				targetPlayer.CameraFollowObj, newPlayer.CameraFollowObj = source, arrow
				for _, pointer := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(arrow), unsafe.Pointer(cached), unsafe.Pointer(live), source.UpdateData, unsafe.Pointer(targetPlayer), unsafe.Pointer(newPlayer), unsafe.Pointer(sourcePlayer), unsafe.Pointer(sourcePlayerUD), unsafe.Pointer(target.HealthData), arrow.InitData} {
					pin.Pin(pointer)
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
						t.Fatalf("pointer=%p, want above 4 GiB", pointer)
					}
				}
				observed := 0
				Nox_xxx_playerObserveClear_4DDEF0 = func(v *server.Object) {
					if v != target || observed != 0 || target.UpdateDataPlayer() != cached || cached.Field76 != 0 || cached.Field75 != 77 {
						t.Fatal("C entry possession-prefix order/repetition")
					}
					observed++
					cached.Field57 = math.Float32bits(0.9)
					target.UpdateData = unsafe.Pointer(live)
				}
				beforeSource, beforePlayerUD, beforePlayer, beforeArrow := *source, *sourcePlayerUD, *sourcePlayer, *arrow
				// Real C dispatcher -> registered PlayerDamage -> DefaultDamage ->
				// UnitSetHP. Only the external ObserveClear service is intercepted.
				if !objectDamageDispatchCallNative(target, source, arrow, 5, object.DamageImpale) || observed != 1 || target.HealthData.Cur != 56 || cached.Field21 != math.Float32bits(0.125) || cached.Field76 != 1 || cached.Field75 != 529 || live.Field21 != math.Float32bits(0.25) || live.Field76 != 31 || live.Field75 != 33 || target.Obj130 != arrow || target.Field131 != 3 || target.Frame134 != srv.Frame() || *source != beforeSource || *sourcePlayerUD != beforePlayerUD || *sourcePlayer != beforePlayer || *arrow != beforeArrow {
					t.Fatalf("C possession: observe=%d HP=%d cached/live marker=%d/%d carry=%g", observed, target.HealthData.Cur, cached.Field76, live.Field76, math.Float32frombits(live.Field21))
				}
				t.Logf("C->possessed player PIERCE: target=%p cached=%p live=%p source=%p missile=%p HP=60->56", target, cached, live, source, arrow)
				runtime.KeepAlive(target)
				runtime.KeepAlive(source)
				runtime.KeepAlive(arrow)
			})
		}
	}
}
