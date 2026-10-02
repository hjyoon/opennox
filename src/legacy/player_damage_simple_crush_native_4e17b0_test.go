package legacy

import (
	"fmt"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

func TestPlayerDamageSimpleCrushNativeCallback4E17B0BidirectionalHP(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	srv.SetFrame(1400)
	oldGetServer := GetServer
	GetServer = func() Server { return &itemDurabilityLegacyServer4E1560{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })
	oldSetState := Nox_xxx_playerSetState_4FA020
	Nox_xxx_playerSetState_4FA020 = func(*server.Object, server.PlayerState) bool {
		t.Fatal("below-threshold Fist unexpectedly requested a hurt state")
		return false
	}
	t.Cleanup(func() { Nox_xxx_playerSetState_4FA020 = oldSetState })
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
	for _, path := range []string{"C-dispatcher", "FistCollide"} {
		for _, toPlayer := range []bool{false, true} {
			for i, name := range []string{"SmallFist", "MediumFist", "LargeFist"} {
				t.Run(fmt.Sprintf("%s/player-target-%t/%s", path, toPlayer, name), func(t *testing.T) {
					player := &server.Player{}
					playerUpdate := &server.PlayerUpdateData{Player: player, State: server.PlayerState13}
					npcUpdate := &server.MonsterUpdateData{}
					playerUnit := &server.Object{TypeInd: 71, ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(playerUpdate), HealthData: &server.HealthData{Cur: 200, Max: 200, Field2: 200}, Material: 0x4000}
					npc := &server.Object{TypeInd: 72, ObjClass: object.ClassMonster, ObjSubClass: 0x11012, UpdateData: unsafe.Pointer(npcUpdate), HealthData: &server.HealthData{Cur: 200, Max: 200, Field2: 200}, Material: 0x4000}
					target, source := npc, playerUnit
					if toPlayer {
						target, source = playerUnit, npc
					}
					target.Damage = playerDamageMeleeCallbackNative4E17B0()
					fistUpdate := &server.FistUpdateData{Damage: 10}
					fist := &server.Object{TypeInd: uint16(777 + i), ObjClass: object.ClassSimple, ObjOwner: source,
						UpdateData: unsafe.Pointer(fistUpdate), PrevPos: types.Pointf{X: 33.5, Y: 44.25}}
					var pin runtime.Pinner
					defer pin.Unpin()
					for _, pointer := range []unsafe.Pointer{unsafe.Pointer(playerUnit), unsafe.Pointer(npc), unsafe.Pointer(player), unsafe.Pointer(playerUpdate), unsafe.Pointer(npcUpdate), unsafe.Pointer(playerUnit.HealthData), unsafe.Pointer(npc.HealthData), unsafe.Pointer(fist), unsafe.Pointer(fistUpdate)} {
						pin.Pin(pointer)
						if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
							t.Fatalf("pointer=%p, want native high address", pointer)
						}
					}
					if !srv.IsEnemyTo(target, source) || fist.FindOwnerChainPlayer() != source {
						t.Fatal("native fixture must retain hostile unit ownership")
					}
					// Both paths use the registered production PlayerDamage adapter and
					// real DefaultDamage/UnitSetHP. No damage or HP service is replaced.
					if path == "C-dispatcher" {
						if !objectDamageDispatchCallNative(target, source, fist, fistUpdate.Damage, object.DamageCrush) {
							t.Fatalf("native Fist callback rejected HP=200->%d", target.HealthData.Cur)
						}
					} else {
						srv.FistCollide4EADF0(fist, target, nil)
					}
					if target.HealthData.Cur != 190 || target.Field38 != math.MaxUint32 || target.Obj130 != fist ||
						target.Field131 != uint32(object.DamageCrush) || target.Pos132 != fist.PrevPos {
						t.Fatalf("native %s HP=200->%d sync=%x attribution=%p position=%v", path, target.HealthData.Cur, target.Field38, target.Obj130, target.Pos132)
					}
					if toPlayer {
						if playerUpdate.Field76 != 1 || playerUpdate.Field75 != uint32(fist.TypeInd) {
							t.Fatal("native player Fist marker was lost")
						}
					} else if npcUpdate.Field547 != 1 || npcUpdate.Field546 != uint32(fist.TypeInd) || !npcUpdate.StatusFlags.Has(object.MonStatusInjured) {
						t.Fatal("native NPC Fist marker was lost")
					}
					t.Logf("%s->registered PlayerDamage->DefaultDamage->UnitSetHP: target=%p source=%p fist=%p HP=200->190", path, target, source, fist)
					runtime.KeepAlive(playerUnit)
					runtime.KeepAlive(npc)
					runtime.KeepAlive(fist)
				})
			}
		}
	}
}
