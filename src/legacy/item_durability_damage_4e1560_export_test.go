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
	"github.com/opennox/opennox/v1/server"
)

type itemDurabilityLegacyServer4E1560 struct {
	Server
	srv *server.Server
}

func (s *itemDurabilityLegacyServer4E1560) S() *server.Server {
	return s.srv
}

func TestPlayerDamageWeaponEntry4E1560KeepsNativePointers(t *testing.T) {
	var (
		recipient int
		packet    []byte
		related   *server.Object
		remove    int
		sequence  int
		sends     int
	)
	srv := &server.Server{
		NetSendPacketXxx: func(gotRecipient int, gotPacket []byte, gotRelated *server.Object, gotRemove, gotSequence int) int {
			sends++
			recipient = gotRecipient
			packet = append([]byte(nil), gotPacket...)
			related = gotRelated
			remove = gotRemove
			sequence = gotSequence
			return 1
		},
	}
	oldGetServer := GetServer
	GetServer = func() Server { return &itemDurabilityLegacyServer4E1560{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })

	oldCoop := noxflags.HasGame(noxflags.GameModeCoop)
	noxflags.SetGame(noxflags.GameModeCoop)
	t.Cleanup(func() {
		if !oldCoop {
			noxflags.UnsetGame(noxflags.GameModeCoop)
		}
	})

	health := &server.HealthData{Cur: 5, Max: 5}
	update := &server.WeaponArmorUpdateData{}
	initData := &server.ModifierInitData{}
	player := &server.Player{PlayerInd: 0xfe}
	ownerUpdate := &server.PlayerUpdateData{Player: player}
	item := &server.Object{
		ObjClass:   object.ClassWeapon,
		NetCode:    0x12345678,
		HealthData: health,
		UpdateData: unsafe.Pointer(update),
		InitData:   unsafe.Pointer(initData),
	}
	owner := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(ownerUpdate)}
	source := &server.Object{}
	effective := &server.Object{}

	damagePointer := objectDamageNativeProbePtr()
	damageCalls := 0
	server.RegisterObjectDamageGo(
		fmt.Sprintf("ItemDurabilityNativeWidth%d", objectDamageNativeTestSequence.Add(1)),
		damagePointer,
		func(gotItem, gotSource, gotEffective *server.Object, damage int32, typ object.DamageType) bool {
			damageCalls++
			if gotItem != item || gotSource != source || gotEffective != effective {
				t.Fatalf("damage objects = %p/%p/%p, want %p/%p/%p",
					gotItem, gotSource, gotEffective, item, source, effective)
			}
			if damage != 1 || typ != object.DamageBlade {
				t.Fatalf("damage values = %d/%d, want 1/%d", damage, typ, object.DamageBlade)
			}
			health.Cur -= uint16(damage)
			return true
		},
	)
	item.Damage = damagePointer

	var pin runtime.Pinner
	for name, pointer := range map[string]unsafe.Pointer{
		"item":         unsafe.Pointer(item),
		"health":       unsafe.Pointer(health),
		"item update":  unsafe.Pointer(update),
		"item init":    unsafe.Pointer(initData),
		"owner":        unsafe.Pointer(owner),
		"owner update": unsafe.Pointer(ownerUpdate),
		"player":       unsafe.Pointer(player),
		"source":       unsafe.Pointer(source),
		"effective":    unsafe.Pointer(effective),
	} {
		pin.Pin(pointer)
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
			pin.Unpin()
			t.Fatalf("%s pointer = %p, want address above the ABI32 range", name, pointer)
		}
	}
	defer pin.Unpin()

	playerDamageWeaponEntry4E1560(item, owner, source, effective, 1, object.DamageBlade)

	if damageCalls != 1 || health.Cur != 4 || update.Field0 != 0 {
		t.Fatalf("durability state = calls:%d health:%d carry:%#x, want 1/4/0",
			damageCalls, health.Cur, update.Field0)
	}
	wantPacket := server.BuildShopItemHealthPacket4D87A0(item)
	if sends != 1 || recipient != player.Index() || !bytes.Equal(packet, wantPacket[:]) ||
		related != nil || remove != 0 || sequence != 1 {
		t.Fatalf("health report = sends:%d recipient:%d packet:% x related:%p remove:%d sequence:%d",
			sends, recipient, packet, related, remove, sequence)
	}

	runtime.KeepAlive(item)
	runtime.KeepAlive(health)
	runtime.KeepAlive(update)
	runtime.KeepAlive(initData)
	runtime.KeepAlive(owner)
	runtime.KeepAlive(ownerUpdate)
	runtime.KeepAlive(player)
	runtime.KeepAlive(source)
	runtime.KeepAlive(effective)
}
