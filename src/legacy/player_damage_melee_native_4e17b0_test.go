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

func TestPlayerDamageMeleeNativeCallback4E17B0BidirectionalHP(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	srv.SetFrame(1400)
	oldGetServer := GetServer
	GetServer = func() Server { return &itemDurabilityLegacyServer4E1560{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })
	// Package opennox installs this service in the game. These 10-damage hits
	// are below its 20-damage hurt threshold, but still require the binding.
	oldSetState := Nox_xxx_playerSetState_4FA020
	Nox_xxx_playerSetState_4FA020 = func(*server.Object, server.PlayerState) bool {
		t.Fatal("below-threshold melee unexpectedly requested a hurt state")
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
		for _, attack := range []struct {
			name     string
			typ      object.DamageType
			subclass object.SubClass
		}{
			{"Sword", object.DamageBlade, object.SubClass(object.WeaponSword)},
			{"MorningStar", object.DamageCrush, object.SubClass(object.WeaponMace)},
			{"Unarmed", object.DamageClaw, 0},
		} {
			t.Run(fmt.Sprintf("player-target-%t/%s", toPlayer, attack.name), func(t *testing.T) {
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
				var weapon *server.Object
				if attack.subclass != 0 {
					weapon = &server.Object{TypeInd: 444, ObjClass: object.ClassWeapon, ObjSubClass: attack.subclass, Material: 0x4000}
				}
				var pin runtime.Pinner
				defer pin.Unpin()
				for _, pointer := range []unsafe.Pointer{unsafe.Pointer(playerUnit), unsafe.Pointer(npc), unsafe.Pointer(player), unsafe.Pointer(playerUpdate), unsafe.Pointer(npcUpdate), unsafe.Pointer(playerUnit.HealthData), unsafe.Pointer(npc.HealthData), unsafe.Pointer(weapon)} {
					if pointer == nil {
						continue
					}
					pin.Pin(pointer)
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
						t.Fatalf("pointer=%p, want native high address", pointer)
					}
				}
				if !srv.IsEnemyTo(target, source) {
					t.Fatal("native fixture must be hostile")
				}
				// The real C dispatcher resolves the registered production adapter,
				// DefaultDamage and UnitSetHP. No damage or HP callback is replaced.
				if !objectDamageDispatchCallNative(target, source, weapon, 10, attack.typ) || target.HealthData.Cur != 190 {
					t.Fatalf("native callback HP=200->%d, want 190", target.HealthData.Cur)
				}
				if target.Field38 != math.MaxUint32 {
					t.Fatal("native HP setter did not mark synchronization")
				}
				t.Logf("C->registered PlayerDamage->DefaultDamage->UnitSetHP: target=%p source=%p weapon=%p HP=200->190", target, source, weapon)
				runtime.KeepAlive(playerUnit)
				runtime.KeepAlive(npc)
				runtime.KeepAlive(weapon)
			})
		}
	}
}
