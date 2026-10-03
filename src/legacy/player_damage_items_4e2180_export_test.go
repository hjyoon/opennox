package legacy

import (
	"bytes"
	"fmt"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestPlayerDamageItemsCEntry4E2180NativePointersAndLiveLinks(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	if !srv.Objs.Init(40) {
		t.Fatal("native object allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	noxflags.SetGame(noxflags.GameModeCoop)
	for _, playerOwner := range []bool{false, true} {
		t.Run(fmt.Sprintf("player-owner-%t", playerOwner), func(t *testing.T) {
			newObject := func(ind uint16) *server.Object {
				obj := srv.Objs.NewObject(&server.ObjectType{})
				obj.TypeInd, obj.ObjFlags = ind, 0
				return obj
			}
			owner, source, effective := newObject(71), newObject(72), newObject(529)
			first, stale, later, noHealth := newObject(301), newObject(302), newObject(303), newObject(304)
			playerUD, freePlayerUD := alloc.New(server.PlayerUpdateData{})
			player, freePlayer := alloc.New(server.Player{})
			monsterUD, freeMonsterUD := alloc.New(server.MonsterUpdateData{})
			firstUD, freeFirstUD := alloc.New(server.WeaponArmorUpdateData{})
			laterUD, freeLaterUD := alloc.New(server.WeaponArmorUpdateData{})
			firstInit, freeFirstInit := alloc.New(server.ModifierInitData{})
			laterInit, freeLaterInit := alloc.New(server.ModifierInitData{})
			firstHP, freeFirstHP := alloc.New(server.HealthData{})
			liveHP, freeLiveHP := alloc.New(server.HealthData{})
			laterHP, freeLaterHP := alloc.New(server.HealthData{})
			firstDef, freeFirstDef := alloc.New(server.Modifier{})
			laterDef, freeLaterDef := alloc.New(server.Modifier{})
			armorEffect, freeArmorEffect := alloc.New(server.ModifierEff{})
			durabilityEffect, freeDurabilityEffect := alloc.New(server.ModifierEff{})
			for _, free := range []func(){freePlayerUD, freePlayer, freeMonsterUD, freeFirstUD, freeLaterUD, freeFirstInit, freeLaterInit, freeFirstHP, freeLiveHP, freeLaterHP, freeFirstDef, freeLaterDef, freeArmorEffect, freeDurabilityEffect} {
				t.Cleanup(free)
			}
			*player = server.Player{PlayerInd: 7}
			*playerUD = server.PlayerUpdateData{Player: player, Field57: math.Float32bits(0.5)}
			*monsterUD = server.MonsterUpdateData{Field518: math.Float32bits(0.5)}
			owner.ObjClass, owner.ObjSubClass, owner.UpdateData = object.ClassMonster, 0x10, unsafe.Pointer(monsterUD)
			if playerOwner {
				owner.ObjClass, owner.UpdateData = object.ClassPlayer, unsafe.Pointer(playerUD)
			}
			*firstUD, *laterUD = server.WeaponArmorUpdateData{Field0: math.Float32bits(0.25)}, server.WeaponArmorUpdateData{}
			*firstInit, *laterInit = server.ModifierInitData{}, server.ModifierInitData{}
			*firstHP, *liveHP, *laterHP = server.HealthData{Cur: 10, Max: 10}, server.HealthData{Cur: 6, Max: 10}, server.HealthData{Cur: 20, Max: 20}
			*firstDef = server.Modifier{TypeInd: uint32(first.TypeInd), DamageCoeffOrArmor64: 0.25, Next80: laterDef}
			*laterDef = server.Modifier{TypeInd: uint32(later.TypeInd), DamageCoeffOrArmor64: 0.125}
			fns := playerDamageLateDefendFunctionsNative4E1320()
			*armorEffect = server.ModifierEff{Defend76: server.ModifierEffFnc{Fnc: fns.armorMultiplier, Valf: 0.5}}
			*durabilityEffect = server.ModifierEff{Defend76: server.ModifierEffFnc{Fnc: fns.durabilityMultiplier, Valf: 1.5}}
			firstInit.Modifiers[0], firstInit.Modifiers[1] = armorEffect, durabilityEffect
			oldDefs := srv.Modif.Dword_5d4594_251608
			srv.Modif.Dword_5d4594_251608 = firstDef
			t.Cleanup(func() { srv.Modif.Dword_5d4594_251608 = oldDefs })
			for _, item := range []*server.Object{first, stale, later, noHealth} {
				item.ObjClass = object.ClassArmor
			}
			first.ObjFlags, stale.ObjFlags, noHealth.ObjFlags = object.FlagEquipped, object.FlagEquipped, object.FlagEquipped
			first.UpdateData, first.InitData, first.HealthData, first.InvHolder = unsafe.Pointer(firstUD), unsafe.Pointer(firstInit), firstHP, owner
			later.UpdateData, later.InitData, later.HealthData, later.InvHolder = unsafe.Pointer(laterUD), unsafe.Pointer(laterInit), laterHP, owner
			owner.InvFirstItem, first.InvNextItem, stale.InvNextItem, later.InvNextItem = first, stale, later, noHealth
			var wears []uint16
			damagePointer := objectDamageNativeProbePtr()
			server.RegisterObjectDamageGo(fmt.Sprintf("ArmorItemsNative%d", objectDamageNativeTestSequence.Add(1)), damagePointer,
				func(item, gotSource, gotEffective *server.Object, damage int32, typ object.DamageType) bool {
					if gotSource != source || gotEffective != effective || typ != object.DamageImpale {
						t.Fatal("native source/effective/type arguments")
					}
					wears = append(wears, item.TypeInd)
					if item == first {
						if damage != 1 || firstUD.Field0 != math.Float32bits(0.25) {
							t.Fatal("C armor lookup/slot-1 durability multiplier/carry order")
						}
						first.HealthData = liveHP
						first.InvNextItem = later
						later.ObjFlags = object.FlagEquipped
						playerUD.Field57, monsterUD.Field518 = math.Float32bits(2), math.Float32bits(2)
					} else if item == later {
						if damage != 2 || laterUD.Field0 != 0 {
							t.Fatal("C entry recaptured armor denominator or snapshotted later flags")
						}
						laterHP.Cur -= uint16(damage)
					} else {
						t.Fatal("native wear reached detached/health-less item")
					}
					return false
				})
			first.Damage, stale.Damage, later.Damage = damagePointer, damagePointer, damagePointer
			var packets [][]byte
			srv.NetSendPacketXxx = func(recipient int, packet []byte, related *server.Object, remove, sequence int) int {
				if !playerOwner || recipient != player.Index() || related != nil || remove != 0 || sequence != 1 {
					t.Fatal("native player-only health report contract")
				}
				packets = append(packets, append([]byte(nil), packet...))
				return 1
			}
			for _, pointer := range []unsafe.Pointer{unsafe.Pointer(owner), unsafe.Pointer(source), unsafe.Pointer(effective), unsafe.Pointer(first), unsafe.Pointer(stale), unsafe.Pointer(later), owner.UpdateData, first.UpdateData, later.UpdateData, first.InitData, later.InitData, unsafe.Pointer(firstHP), unsafe.Pointer(liveHP), unsafe.Pointer(laterHP), unsafe.Pointer(firstDef), unsafe.Pointer(laterDef), unsafe.Pointer(armorEffect), unsafe.Pointer(durabilityEffect), unsafe.Pointer(player)} {
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
					t.Fatalf("native pointer=%p, want above 4 GiB", pointer)
				}
			}
			// Actual C symbol -> exported Go 004E2180 -> unmodified C
			// 00415C00/slot-zero effect -> production EquipDamage/carry.
			Nox_xxx_playerDamageItems_4E2180(owner, source, effective, 8, object.DamageImpale)
			if len(wears) != 2 || wears[0] != first.TypeInd || wears[1] != later.TypeInd || first.HealthData != liveHP || firstHP.Cur != 10 || liveHP.Cur != 6 || laterHP.Cur != 18 || first.InvNextItem != later {
				t.Fatalf("native wears=%v oldHP=%d liveHP=%d laterHP=%d", wears, firstHP.Cur, liveHP.Cur, laterHP.Cur)
			}
			if playerOwner {
				if len(packets) != 2 {
					t.Fatalf("health reports=%d, want two after each wear", len(packets))
				}
				for i, item := range []*server.Object{first, later} {
					want := server.BuildShopItemHealthPacket4D87A0(item)
					if !bytes.Equal(packets[i], want[:]) {
						t.Fatalf("report %d=% x, want % x", i, packets[i], want)
					}
				}
			} else if len(packets) != 0 {
				t.Fatal("NPC armor wear emitted player health packets")
			}
			t.Logf("C armor-items entry: owner=%p source=%p effective=%p first=%p later=%p wears=%v", owner, source, effective, first, later, wears)
		})
	}
}

func TestPlayerDamageItemsCEntry4E2180ClassReturnBeforeServices(t *testing.T) {
	oldGetServer := GetServer
	GetServer = func() Server { t.Fatal("rejected/empty inventory consulted the server"); return nil }
	t.Cleanup(func() { GetServer = oldGetServer })
	for _, target := range []*server.Object{
		{ObjClass: object.ClassMonster},
		{ObjClass: object.ClassArmor, ObjSubClass: 0x10},
		{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(&server.PlayerUpdateData{})},
	} {
		var pin runtime.Pinner
		pin.Pin(target)
		if target.UpdateData != nil {
			pin.Pin(target.UpdateData)
		}
		Nox_xxx_playerDamageItems_4E2180(target, nil, nil, 0, object.DamageType(0x11223344))
		pin.Unpin()
	}
}
