package legacy

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/player"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/server"
)

const (
	quickBarRows45DA50        = 5
	quickBarSlots45DA50       = 5
	quickBarRowSize45DA50     = 40
	quickBarSlotSize45DA50    = 8
	quickBarSelected45DA50    = 200
	clientSpellPacket45DA50   = 22
	clientAbilityPacket45DA50 = 2
)

func clientPlayerAnimationBlocked45DA50(dr *client.Drawable) bool {
	if dr == nil {
		return true
	}
	switch dr.AnimInd {
	case 1, 2, 51:
		return true
	default:
		return false
	}
}

func quickBarSpellData45DA50(base unsafe.Pointer, slot int) (spell uint32, flags byte, ok bool) {
	if base == nil || slot < 0 || slot >= quickBarSlots45DA50 {
		return 0, 0, false
	}
	row := int(*(*byte)(unsafe.Add(base, quickBarSelected45DA50)))
	if row < 0 || row >= quickBarRows45DA50 {
		return 0, 0, false
	}
	off := row*quickBarRowSize45DA50 + slot*quickBarSlotSize45DA50
	data := unsafe.Slice((*byte)(unsafe.Add(base, off)), quickBarSlotSize45DA50)
	return binary.LittleEndian.Uint32(data), data[4], true
}

func clientSpellSlotPacket45DA50(class player.Class, spell uint32, flags byte) (packet [clientSpellPacket45DA50]byte, size int, ok bool) {
	if spell == 0 {
		return packet, 0, false
	}
	if class == player.Warrior {
		packet[0] = byte(netmsg.MSG_TRY_ABILITY)
		packet[1] = byte(spell)
		return packet, clientAbilityPacket45DA50, true
	}
	packet[0] = byte(netmsg.MSG_TRY_SPELL)
	binary.LittleEndian.PutUint32(packet[1:5], spell)
	packet[21] = flags & 1
	return packet, clientSpellPacket45DA50, true
}

func clientInvokeSpellSlot45DA50(slot int) {
	pl := Get_dword_8531A0_2576()
	dr := AsDrawableP(*memmap.PtrPtr(0x852978, 8))
	if pl == nil || clientPlayerAnimationBlocked45DA50(dr) {
		return
	}
	spell, flags, ok := quickBarSpellData45DA50(quickBarBase45DA50(), slot)
	if !ok {
		return
	}
	packet, size, send := clientSpellSlotPacket45DA50(pl.PlayerClass(), spell, flags)
	if memmap.Uint32(0x5D4594, 1096672) == 0 && send {
		GetServer().S().NetList.AddToMsgListCli(server.HostPlayerIndex, netlist.Kind0, packet[:size])
	}
	*memmap.PtrUint32(0x587000, 133484) = uint32(slot)
	*memmap.PtrUint32(0x5D4594, 1049540) = gameFrameHook()
}
