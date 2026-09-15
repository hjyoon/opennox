package legacy

import (
	"encoding/binary"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/player"

	"github.com/opennox/opennox/v1/client"
)

func TestClientPlayerAnimationBlocked45DA50(t *testing.T) {
	for _, tc := range []struct {
		name string
		dr   *client.Drawable
		want bool
	}{
		{name: "missing drawable", want: true},
		{name: "death animation", dr: &client.Drawable{AnimInd: 1}, want: true},
		{name: "dying animation", dr: &client.Drawable{AnimInd: 2}, want: true},
		{name: "special blocked animation", dr: &client.Drawable{AnimInd: 51}, want: true},
		{name: "idle", dr: &client.Drawable{AnimInd: 0}},
		{name: "casting", dr: &client.Drawable{AnimInd: 3}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := clientPlayerAnimationBlocked45DA50(tc.dr); got != tc.want {
				t.Fatalf("blocked = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestQuickBarSpellData45DA50(t *testing.T) {
	quickbar := make([]byte, 240)
	quickbar[quickBarSelected45DA50] = 2
	off := 2*quickBarRowSize45DA50 + 3*quickBarSlotSize45DA50
	binary.LittleEndian.PutUint32(quickbar[off:off+4], 0x12345678)
	quickbar[off+4] = 3

	spell, flags, ok := quickBarSpellData45DA50(unsafe.Pointer(&quickbar[0]), 3)
	if spell != 0x12345678 || flags != 3 || !ok {
		t.Fatalf("spell data = %#x, %d, %t; want 0x12345678, 3, true", spell, flags, ok)
	}

	for _, tc := range []struct {
		name string
		base unsafe.Pointer
		slot int
	}{
		{name: "nil base", slot: 2},
		{name: "negative slot", base: unsafe.Pointer(&quickbar[0]), slot: -1},
		{name: "slot past end", base: unsafe.Pointer(&quickbar[0]), slot: quickBarSlots45DA50},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, ok := quickBarSpellData45DA50(tc.base, tc.slot); ok {
				t.Fatal("spell data accepted invalid input")
			}
		})
	}

	quickbar[quickBarSelected45DA50] = quickBarRows45DA50
	if _, _, ok := quickBarSpellData45DA50(unsafe.Pointer(&quickbar[0]), 2); ok {
		t.Fatal("spell data accepted an invalid selected row")
	}
}

func TestClientSpellSlotPacket45DA50(t *testing.T) {
	tests := []struct {
		name  string
		class player.Class
		spell uint32
		flags byte
		want  []byte
		ok    bool
	}{
		{name: "empty warrior slot", class: player.Warrior},
		{name: "empty wizard slot", class: player.Wizard},
		{name: "warrior ability", class: player.Warrior, spell: 0x12345678, want: []byte{byte(netmsg.MSG_TRY_ABILITY), 0x78}, ok: true},
		{name: "wizard spell", class: player.Wizard, spell: 0x12345678, flags: 3, want: []byte{byte(netmsg.MSG_TRY_SPELL), 0x78, 0x56, 0x34, 0x12, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}, ok: true},
		{name: "conjurer spell", class: player.Conjurer, spell: 7, flags: 2, want: []byte{byte(netmsg.MSG_TRY_SPELL), 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, ok: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			packet, size, ok := clientSpellSlotPacket45DA50(tc.class, tc.spell, tc.flags)
			if ok != tc.ok {
				t.Fatalf("ok = %t, want %t", ok, tc.ok)
			}
			if got := packet[:size]; string(got) != string(tc.want) {
				t.Fatalf("packet = % X, want % X", got, tc.want)
			}
		})
	}
}
