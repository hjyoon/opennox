package legacy

import (
	"fmt"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

func TestPlayerDamageUnarmedPlayerElectricNativeCallback4E17B0HP(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	srv.SetFrame(1400)
	oldGetServer := GetServer
	GetServer = func() Server { return &itemDurabilityLegacyServer4E1560{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })
	oldSetState := Nox_xxx_playerSetState_4FA020
	Nox_xxx_playerSetState_4FA020 = func(*server.Object, server.PlayerState) bool {
		t.Fatal("below-threshold electric hit requested hurt state")
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
	for _, toPlayer := range []bool{false, true} {
		for _, typ := range []object.DamageType{object.DamageElectric, object.DamageAirborneElectric} {
			t.Run(fmt.Sprintf("player-target-%t/%s", toPlayer, typ), func(t *testing.T) {
				sourcePlayer, targetPlayer := &server.Player{}, &server.Player{}
				sourceUpdate := &server.PlayerUpdateData{Player: sourcePlayer, State: server.PlayerState13}
				targetUpdate := &server.PlayerUpdateData{Player: targetPlayer, State: server.PlayerState13, Field40_0: 0x1122, Field40_1: 0xabcd}
				npcUpdate := &server.MonsterUpdateData{}
				source := &server.Object{TypeInd: 71, ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(sourceUpdate), HealthData: &server.HealthData{Cur: 200, Max: 200}, Material: 0x4000}
				target := &server.Object{TypeInd: 72, ObjClass: object.ClassMonster, ObjSubClass: 0x11012, UpdateData: unsafe.Pointer(npcUpdate), HealthData: &server.HealthData{Cur: 200, Max: 200}, Material: 0x4000}
				if toPlayer {
					target.ObjClass, target.ObjSubClass, target.UpdateData = object.ClassPlayer, 0, unsafe.Pointer(targetUpdate)
				}
				target.Damage = playerDamageMeleeCallbackNative4E17B0()
				var pin runtime.Pinner
				defer pin.Unpin()
				for _, pointer := range []unsafe.Pointer{unsafe.Pointer(source), unsafe.Pointer(target), unsafe.Pointer(sourcePlayer), unsafe.Pointer(targetPlayer), unsafe.Pointer(sourceUpdate), unsafe.Pointer(targetUpdate), unsafe.Pointer(npcUpdate), unsafe.Pointer(source.HealthData), unsafe.Pointer(target.HealthData)} {
					pin.Pin(pointer)
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
						t.Fatalf("pointer=%p, want native high address", pointer)
					}
				}
				if !srv.IsEnemyTo(target, source) {
					t.Fatal("native fixture must be hostile")
				}
				// The actual C dispatcher and registered production damage adapter
				// run through DefaultDamage and UnitSetHP. No HP/damage replacement.
				if !objectDamageDispatchCallNative(target, source, nil, 8, typ) || target.HealthData.Cur != 192 || target.Obj130 != source || target.Field131 != uint32(typ) || target.Frame134 != 1400 || target.Field38 != math.MaxUint32 {
					t.Fatalf("native electric: HP=200->%d attribution=%p/%d/%d synchronization=%#x", target.HealthData.Cur, target.Obj130, target.Field131, target.Frame134, target.Field38)
				}
				if toPlayer {
					if targetUpdate.Field40_0 != 2 || targetUpdate.Field40_1 != 0xabcd || targetUpdate.Field76 != 2 || targetUpdate.Field75 != uint32(typ) {
						t.Fatal("native player electric state")
					}
				} else if npcUpdate.Field523_2 != 2 || npcUpdate.Field547 != 2 || npcUpdate.Field546 != uint32(typ) {
					t.Fatal("native NPC electric state")
				}
				t.Logf("C->PlayerDamage->DefaultDamage->UnitSetHP: target=%p source=%p weapon=nil HP=200->192 type=%d", target, source, typ)
				runtime.KeepAlive(source)
				runtime.KeepAlive(target)
			})
		}
	}
}
