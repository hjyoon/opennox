package legacy

import (
	"bytes"
	"fmt"
	"testing"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestGame1PublicEntriesNativeCABI(t *testing.T) {
	world := playerVitalReportWorld(t)
	for route := range 2 {
		t.Run(fmt.Sprintf("route%d", route), func(t *testing.T) {
			name, freeName := alloc.CStringHigh("NativeFacadeFixture")
			other, freeOther := alloc.CStringHigh("NativeRegularTile")
			t.Cleanup(freeName)
			t.Cleanup(freeOther)
			first, last := memmap.PtrPtr(0x587000, 26488), memmap.PtrPtr(0x587000, 26492)
			oldFirst, oldLast := *first, *last
			defer func() { *first, *last = oldFirst, oldLast }()
			*first, *last = unsafe.Pointer(name), nil
			if callGame1FacadeNative(route, name) != 1 || callGame1FacadeNative(route, other) != 0 {
				t.Fatal("C facade lookup changed the tile name")
			}

			marker := game4ConsumerValue(t, byte(0x7a))
			icon := itemEnchantmentIconCall413420
			defer func() { itemEnchantmentIconCall413420 = icon }()
			itemEnchantmentIconCall413420 = func(flag byte) unsafe.Pointer {
				if flag != 0x84 {
					t.Fatalf("enchantment flag %#x, want 0x84", flag)
				}
				return unsafe.Pointer(marker)
			}
			if got := callGame1IconNative(route, 0x84); got != unsafe.Pointer(marker) {
				t.Fatalf("image pointer %p, want %p", got, marker)
			}

			const code = 0x9234
			player := world.Players.ByIndRaw(19)
			*player = server.Player{Active: 1, NetCodeVal: code, Field3680: 0x200}
			defer func() { *player = server.Player{} }()
			if cgoABIPointerSize == 8 && uintptr(player.C()) <= 0xffffffff {
				t.Fatal("C calls need a native player above 4 GiB")
			}
			noxflags.ResetGame()
			noxflags.SetGame(noxflags.GameHost)
			world.NetSendPacketXxx = func(index int, data []byte, related *server.Object, remove, sequence int) int {
				if index != 255 || !bytes.Equal(data, []byte{106, 0x34, 0x92, 0, 4, 0, 0}) ||
					related != nil || remove != 0 || sequence != 1 || player.Field3680 != 0x600 {
					t.Fatal("C player status changed the order or bytes of the report")
				}
				return -17
			}
			if got := callGame1StatusNative(route, player, 0x400); got != -17 {
				t.Fatalf("C status result = %d, want -17", got)
			}
			noxflags.ResetGame()

			ids := [4]byte{0xfe, 0xfd, 0xfc, 0xfb}
			before := ids
			const weapon, armor = 0x800, 0x200
			if callGame1EquipNative(route, 80, code, weapon, &ids) != player ||
				player.WeaponEquip != weapon || player.Weapon[0].Field0 != weapon || ids != before {
				t.Fatal("weapon equip changed the pointer, mask, record or modifier bytes")
			}
			if callGame1DequipNative(route, 84, code, weapon) != player ||
				player.WeaponEquip != 0 || player.Weapon[0].Field0 != 0 {
				t.Fatal("weapon dequip changed the pointer or failed to clear equipment")
			}
			if callGame1EquipNative(route, 82, code, armor, nil) != player ||
				player.ArmorEquip != armor || player.Armor[0].Field0 != armor {
				t.Fatal("armor equip changed the pointer or nil modifier handling")
			}
			player.Armor[0].Field20 = 0x8900
			callGame1RespawnNative(route, player, 0xff)
			if player.ArmorEquip != 0 || player.Armor[0].Field0 != 0 || player.Armor[0].Field20 != 0x8900 {
				t.Fatal("respawn failed to clear the armor record and preserve its trailing word")
			}
		})
	}
}
