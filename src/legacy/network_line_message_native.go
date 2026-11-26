package legacy

/*
#include "defs.h"
#include "noxstring_format_abi.h"
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet"
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/server"
)

const lineMessageWideFlag = noxnet.TextFlags(0x4)

// The wire length includes the terminator in a single byte. A 255-unit
// message therefore has a zero size, as in the original packet builder.
func buildLineMessagePacket(dst []byte, text []uint16, narrow bool) int {
	size := uint8(len(text))
	var payload [510]byte
	encoded := payload[:int(size)]
	flags := noxnet.TextUTF8
	if narrow {
		for i, unit := range text[:size] {
			encoded[i] = byte(unit)
		}
	} else {
		flags = lineMessageWideFlag
		encoded = payload[:2*int(size)]
		for i, unit := range text[:size] {
			binary.LittleEndian.PutUint16(encoded[2*i:2*i+2], unit)
		}
	}
	msg := noxnet.MsgText{Flags: flags, Size: size, Data: encoded}
	dst[0] = byte(netmsg.MSG_TEXT_MESSAGE)
	n, err := msg.Encode(dst[1:])
	if err != nil {
		return 0
	}
	return 1 + n
}

func sendLineMessageText(update unsafe.Pointer, text *wchar2_t) C.intptr_t {
	length := int(nox_wcslen(text)) + 1
	words := unsafe.Slice((*uint16)(unsafe.Pointer(text)), length)
	var packet [521]byte
	size := buildLineMessagePacket(packet[:], words, cliCanTalkNative4100F0(&words[0]) != 0)
	index := ntype.PlayerInd(nox_server_playerIndexFromUpdateData_4D9EB0(update))
	queued := GetServer().S().NetList.AddToMsgListCli(index, netlist.Kind1, packet[:size])
	return C.intptr_t(bool2int(queued))
}

//export nox_net_line_message_vprintf_4D9EB0
func nox_net_line_message_vprintf_4D9EB0(obj *nox_object_t, format *C.nox_format_word, args *C.nox_format_args) C.intptr_t {
	return sendFormattedLineMessage(asObjectS(obj), format, formatArguments{native: args})
}

func sendNumberedLineMessage(player *server.Object, format *wchar2_t, value uint32) {
	values := [...]formatValue{{number: int32(value)}}
	sendFormattedLineMessage(player, format, formatArguments{values: values[:]})
}

func sendFormattedLineMessage(player *server.Object, format *wchar2_t, args formatArguments) C.intptr_t {
	if !player.Class().Has(object.ClassPlayer) {
		return C.intptr_t(uintptr(unsafe.Pointer(player)))
	}

	// Cache update data before formatting reads any arguments. Reserve the
	// final zero-initialized word for the terminator of a truncated message.
	update := player.UpdateData
	var text [256]wchar2_t
	formatBoundedWide(text[:], format, args)
	return sendLineMessageText(update, &text[0])
}
