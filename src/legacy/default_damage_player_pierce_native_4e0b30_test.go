package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func TestDefaultDamagePlayerPierceNativeCallback4E0B30(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	if err := srv.Types.ReadObjectType(&things.Thing{Name: "NativePlayerPierceDefault", OnDamage: &things.ProcFunc{Name: "DefaultDamage"}}); err != nil {
		t.Fatal(err)
	}
	for _, pure := range []bool{false, true} {
		t.Run(map[bool]string{false: "ranged missile weapon", true: "pure missile"}[pure], func(t *testing.T) {
			var pin runtime.Pinner
			defer pin.Unpin()
			ud := &server.MonsterUpdateData{}
			player := &server.Player{}
			playerUD := &server.PlayerUpdateData{Player: player, State: server.PlayerState13}
			target := &server.Object{ObjClass: object.ClassMonster, ObjSubClass: 0x10, UpdateData: unsafe.Pointer(ud),
				HealthData: &server.HealthData{Cur: 60, Max: 60}, Material: 0x4000, Damage: srv.Types.ByID("NativePlayerPierceDefault").Damage}
			source := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(playerUD), PrevPos: types.Ptf(44, 7)}
			arrow := &server.Object{TypeInd: 529, ObjClass: object.Class(0x05200001), ObjSubClass: 0x10, PrevPos: types.Ptf(20, 0), ObjOwner: source, InitData: unsafe.Pointer(&server.ModifierInitData{})}
			if pure {
				arrow.ObjClass = object.ClassMissile
			}
			for _, pointer := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(arrow), target.UpdateData, source.UpdateData, unsafe.Pointer(player), unsafe.Pointer(target.HealthData), arrow.InitData} {
				pin.Pin(pointer)
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
					t.Fatalf("pointer=%p, want above 4 GiB", pointer)
				}
			}
			wantPos := source.PrevPos
			if pure {
				wantPos = arrow.PrevPos
			}
			for hit, wantHP := range []uint16{57, 54, 51} {
				// The real C damage dispatcher enters the registered production
				// DefaultDamage handler and unmodified UnitSetHP.
				if !objectDamageDispatchCallNative(target, source, arrow, 3, object.DamageImpale) || target.HealthData.Cur != wantHP ||
					ud.Field547 != 1 || ud.Field546 != 529 || target.Obj130 != arrow || target.Field131 != 3 || target.Frame134 != srv.Frame() || target.Pos132 != wantPos {
					t.Fatalf("C entry hit=%d HP=%d marker=%d/%d", hit, target.HealthData.Cur, ud.Field547, ud.Field546)
				}
			}
			t.Logf("C->DefaultDamage player PIERCE: target=%p source=%p missile=%p HP=60->57->54->51", target, source, arrow)
			runtime.KeepAlive(target)
			runtime.KeepAlive(source)
			runtime.KeepAlive(arrow)
		})
	}
}
