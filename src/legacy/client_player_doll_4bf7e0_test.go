package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func TestClientPlayerDollState4BF7E0UsesNativePlayerLayout(t *testing.T) {
	oldPlayer := Get_dword_8531A0_2576()
	t.Cleanup(func() { Set_dword_8531A0_2576(oldPlayer) })

	Set_dword_8531A0_2576(nil)
	if state, ok := clientPlayerDollStateNative4BF7E0(); ok || state != (clientPlayerDollState4BF7E0{}) {
		t.Fatalf("nil player state = %+v, %t; want zero, false", state, ok)
	}

	player := new(server.Player)
	var pin runtime.Pinner
	pin.Pin(player)
	t.Cleanup(pin.Unpin)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(player)) <= math.MaxUint32 {
		t.Fatalf("player pointer = %p, want native address above 4 GiB", player)
	}
	player.ArmorEquip = 0x03000007
	player.WeaponEquip = 0x04000009
	player.Colors.Skin = 0x11111111
	player.Colors.Hair = 0x22222222
	player.Colors.Mustache = 0x33333333
	player.Colors.Goatee = 0x44444444
	player.Colors.Beard = 0x55555555
	player.Colors.UnkColor = 0x66666666
	player.Info().SetIsFemale(1)
	Set_dword_8531A0_2576(player)

	want := clientPlayerDollState4BF7E0{
		armorMask:     player.ArmorEquip,
		weaponMask:    player.WeaponEquip,
		colorSkin:     player.Colors.Skin,
		colorHair:     player.Colors.Hair,
		colorMustache: player.Colors.Mustache,
		colorGoatee:   player.Colors.Goatee,
		colorBeard:    player.Colors.Beard,
		colorUnknown:  player.Colors.UnkColor,
		variant:       1,
	}
	if got, ok := clientPlayerDollStateNative4BF7E0(); !ok || got != want {
		t.Fatalf("player doll state = %+v, %t; want %+v, true", got, ok, want)
	}
	runtime.KeepAlive(player)
}

func TestClientPlayerDollImage4BF9F0UsesNativePointerSlots(t *testing.T) {
	const (
		tableOffset = uintptr(256 + 108)
		layer       = 26
	)
	slot := memmap.PtrPtr(0x973A20, tableOffset+4*layer)
	old := *slot
	t.Cleanup(func() { *slot = old })

	marker := new(byte)
	var pin runtime.Pinner
	pin.Pin(marker)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(marker)) <= math.MaxUint32 {
		t.Fatalf("image pointer = %p, want native address above 4 GiB", marker)
	}
	*slot = unsafe.Pointer(marker)
	if got := clientPlayerDollImageNative4BF9F0(tableOffset, layer); got != unsafe.Pointer(marker) {
		t.Fatalf("player doll image pointer = %p, want %p", got, marker)
	}
	runtime.KeepAlive(marker)
}
