package opennox

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// A synthetic native-width fixture of DefaultDamage itself, not a claim about
// NPC PlayerDamage armor admission or a stock-map player attack. The registered
// callback and its HP/owner-report services are production implementations.
// Only the outgoing network transport is captured; every C-bound record and
// link is C-owned. The original assets and object definitions are not changed.
func TestDefaultDamagePlayerPierceNative4E0B30HPAndOwnerReport(t *testing.T) {
	base := server.New(nil, nil, strman.New())
	t.Cleanup(base.Close)
	s := &Server{Server: base}
	oldServer, oldGetServer := noxServer, legacy.GetServer
	noxServer, legacy.GetServer = s, func() legacy.Server { return s }
	t.Cleanup(func() { noxServer, legacy.GetServer = oldServer, oldGetServer })
	oldGame, oldEngine, oldGameplay := noxflags.GetGame(), noxflags.GetEngine(), noxflags.GetGamePlay()
	noxflags.UnsetGame(oldGame)
	noxflags.UnsetEngine(oldEngine)
	noxflags.UnsetGamePlay(oldGameplay)
	noxflags.SetGamePlay(noxflags.GameplayFlag1)
	t.Cleanup(func() {
		noxflags.UnsetGame(noxflags.GetGame())
		noxflags.UnsetEngine(noxflags.GetEngine())
		noxflags.UnsetGamePlay(noxflags.GetGamePlay())
		noxflags.SetGame(oldGame)
		noxflags.SetEngine(oldEngine)
		noxflags.SetGamePlay(oldGameplay)
	})
	s.SetFrame(1400)
	s.Audio.Init(s.Server)
	t.Cleanup(s.Audio.Free)
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	if err := s.Types.ReadObjectType(&things.Thing{Name: "NativePlayerPierceDefault", OnDamage: &things.ProcFunc{Name: "DefaultDamage"}}); err != nil {
		t.Fatal(err)
	}
	if !s.Objs.Init(24) {
		t.Fatal("native allocator initialization")
	}
	t.Cleanup(s.Objs.FreeObjects)
	for _, npc := range []bool{false, true} {
		for _, pure := range []bool{false, true} {
			t.Run(fmt.Sprintf("NPC-%t/pure-%t", npc, pure), func(t *testing.T) {
				def := s.Types.ByID("NativePlayerPierceDefault")
				target, source, arrow := s.Objs.NewObject(def), s.Objs.NewObject(&server.ObjectType{}), s.Objs.NewObject(&server.ObjectType{})
				ud, freeUD := alloc.New(server.MonsterUpdateData{})
				playerUD, freePlayerUD := alloc.New(server.PlayerUpdateData{})
				player, freePlayer := alloc.New(server.Player{})
				hp, freeHP := alloc.New(server.HealthData{})
				for _, free := range []func(){freeUD, freePlayerUD, freePlayer, freeHP} {
					t.Cleanup(free)
				}
				*ud, *playerUD, *player = server.MonsterUpdateData{}, server.PlayerUpdateData{}, server.Player{}
				*hp = server.HealthData{Cur: 60, Max: 60}
				target.TypeInd, target.ObjClass, target.ObjFlags, target.Material = 71, object.ClassMonster, 0, 0x4000
				if npc {
					target.ObjSubClass = 0x10
				}
				target.UpdateData, target.HealthData = unsafe.Pointer(ud), hp
				source.TypeInd, source.ObjClass, source.ObjFlags, source.UpdateData = 72, object.ClassPlayer, 0, unsafe.Pointer(playerUD)
				source.PrevPos = types.Ptf(44, 7)
				playerUD.Player, playerUD.State, player.PlayerInd = player, server.PlayerState13, 7
				// Also prove an owned unit's real HP wire report. GameplayFlag1
				// admits this ranged self-owned hit, as in the original prefix.
				target.ObjOwner, source.Field129, target.NetCode = source, target, 0x1234
				arrow.TypeInd, arrow.ObjClass, arrow.ObjSubClass, arrow.ObjFlags = 529, object.Class(0x05200001), 0x10, 0
				if pure {
					arrow.ObjClass = object.ClassMissile
				}
				arrow.ObjOwner, arrow.PrevPos = source, types.Ptf(20, 0)
				for _, pointer := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(arrow), target.UpdateData, source.UpdateData, unsafe.Pointer(player), unsafe.Pointer(hp)} {
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
						t.Fatalf("pointer=%p, want above 4 GiB", pointer)
					}
				}
				beforeSource, beforeArrow, beforePlayerUD := *source, *arrow, *playerUD
				previousSend, reports := s.NetSendPacketXxx, 0
				s.NetSendPacketXxx = func(recipient int, packet []byte, related *server.Object, remove, sequence int) int {
					want := [4]byte{65, 0x34, 0x12, byte(hp.Cur >> 1)}
					if recipient != 7 || len(packet) != 4 || [4]byte(packet) != want || related != nil || remove != 1 || sequence != 1 {
						t.Fatalf("owner HP report: recipient=%d packet=%x", recipient, packet)
					}
					reports++
					return 1
				}
				t.Cleanup(func() { s.NetSendPacketXxx = previousSend })
				wantPos := source.PrevPos
				if pure {
					wantPos = arrow.PrevPos
				}
				for hit, wantHP := range []uint16{57, 54, 51} {
					if !target.CallDamage(source, arrow, 3, object.DamageImpale) || hp.Cur != wantHP || ud.Field547 != 1 || ud.Field546 != 529 ||
						target.Obj130 != arrow || target.Field131 != 3 || target.Frame134 != 1400 || target.Pos132 != wantPos ||
						*source != beforeSource || *arrow != beforeArrow || *playerUD != beforePlayerUD || reports != hit+1 || !ud.StatusFlags.Has(object.MonStatusInjured) || target.Field38 != math.MaxUint32 {
						t.Fatalf("hit=%d HP=%d marker=%d/%d owner reports=%d", hit, hp.Cur, ud.Field547, ud.Field546, reports)
					}
				}
				t.Logf("registered C-owned DefaultDamage: target=%p player=%p missile=%p HP=60->57->54->51 owner reports=%d", target, source, arrow, reports)
			})
		}
	}
}
