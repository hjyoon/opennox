package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func TestPlayerDamageNPCBerserkerNativeCallback4E17B0(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	var pin runtime.Pinner
	defer pin.Unpin()
	ud := &server.MonsterUpdateData{Field518: math.Float32bits(0.25), Field547: 99, Field546: 77}
	target := &server.Object{ObjClass: object.ClassMonster, ObjSubClass: 0x11012, UpdateData: unsafe.Pointer(ud),
		HealthData: &server.HealthData{Cur: 60, Max: 60}, Material: 0x4000, Damage: playerDamageMeleeCallbackNative4E17B0()}
	player := &server.Player{PlayerInd: 7}
	playerUD := &server.PlayerUpdateData{Player: player, State: server.PlayerState13, Field57: math.Float32bits(0.125)}
	source := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(playerUD), PrevPos: types.Ptf(20, 7)}
	for _, pointer := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(source), target.UpdateData, source.UpdateData, unsafe.Pointer(player), unsafe.Pointer(target.HealthData)} {
		pin.Pin(pointer)
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
			t.Fatalf("pointer=%p, want above 4 GiB", pointer)
		}
	}
	beforeSource, beforeUD, beforePlayer := *source, *playerUD, *player
	for hit, wantHP := range []uint16{57, 55, 52} {
		// Real C damage dispatcher, production PlayerDamage adapter, shared
		// DefaultDamage and UnitSetHP. Only fixture memory is pinned; no HP,
		// damage, armor or collision callback is replaced.
		if !objectDamageDispatchCallNative(target, source, source, 3, object.DamageCrush) || target.HealthData.Cur != wantHP ||
			ud.Field547 != 2 || ud.Field546 != uint32(object.DamageCrush) || target.Obj130 != source ||
			target.Field131 != uint32(object.DamageCrush) || target.Frame134 != srv.Frame() ||
			*source != beforeSource || *playerUD != beforeUD || *player != beforePlayer {
			t.Fatalf("C entry charge=%d HP=%d marker=%d/%d", hit, target.HealthData.Cur, ud.Field547, ud.Field546)
		}
	}
	t.Logf("C->NPC charge: target=%p source/self-weapon=%p HP=60->57->55->52 carry=%g", target, source, math.Float32frombits(ud.Field1))
	runtime.KeepAlive(target)
	runtime.KeepAlive(source)
}
