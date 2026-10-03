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

func TestPlayerDamageNPCPierceNativeCallback4E17B0(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	for _, playerSource := range []bool{false, true} {
		for _, pure := range []bool{false, true} {
			t.Run(fmt.Sprintf("player-source-%t/pure-%t", playerSource, pure), func(t *testing.T) {
				var pin runtime.Pinner
				defer pin.Unpin()
				ud, sourceUD := &server.MonsterUpdateData{Field518: math.Float32bits(0.25)}, &server.MonsterUpdateData{}
				target := &server.Object{ObjClass: object.ClassMonster, ObjSubClass: 0x11012, UpdateData: unsafe.Pointer(ud),
					HealthData: &server.HealthData{Cur: 60, Max: 60}, Material: 0x4000, Damage: playerDamageMeleeCallbackNative4E17B0()}
				source := &server.Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(sourceUD), PrevPos: types.Ptf(44, 7)}
				player := &server.Player{PlayerInd: 7}
				playerUD := &server.PlayerUpdateData{Player: player, State: server.PlayerState13, Field57: math.Float32bits(0.125)}
				if playerSource {
					source.ObjClass, source.UpdateData = object.ClassPlayer, unsafe.Pointer(playerUD)
				}
				arrow := &server.Object{TypeInd: 529, ObjClass: object.Class(0x05200001), ObjSubClass: 0x10, PrevPos: types.Ptf(20, 0), InitData: unsafe.Pointer(&server.ModifierInitData{})}
				if pure {
					arrow.ObjClass = object.ClassMissile
				}
				for _, pointer := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(arrow), target.UpdateData, source.UpdateData, unsafe.Pointer(player), unsafe.Pointer(target.HealthData), arrow.InitData} {
					pin.Pin(pointer)
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
						t.Fatalf("pointer=%p, want above 4 GiB", pointer)
					}
				}
				beforeSource, beforePlayerUD, beforePlayer := *source, *playerUD, *player
				for hit, wantHP := range []uint16{58, 56, 53} {
					// The C damage dispatcher invokes the registered production
					// PlayerDamage shim and its unmodified DefaultDamage/UnitSetHP.
					if !objectDamageDispatchCallNative(target, source, arrow, 3, object.DamageImpale) || target.HealthData.Cur != wantHP ||
						ud.Field547 != 1 || ud.Field546 != 529 || target.Obj130 != arrow || target.Field131 != 3 || target.Frame134 != srv.Frame() ||
						(playerSource && (*source != beforeSource || *playerUD != beforePlayerUD || *player != beforePlayer)) {
						t.Fatalf("C entry hit=%d HP=%d marker=%d/%d", hit, target.HealthData.Cur, ud.Field547, ud.Field546)
					}
				}
				t.Logf("C->NPC PlayerDamage: target=%p source=%p missile=%p HP=60->58->56->53 carry=%g", target, source, arrow, math.Float32frombits(ud.Field1))
				runtime.KeepAlive(target)
				runtime.KeepAlive(source)
				runtime.KeepAlive(arrow)
			})
		}
	}
}
