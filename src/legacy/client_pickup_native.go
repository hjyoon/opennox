package legacy

/*
#include "client__gui__gamewin__gamewin.h"
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_clientPickup_46C140
func nox_xxx_clientPickup_46C140(drawable *C.nox_drawable) {
	Nox_xxx_clientPickup_46C140((*client.Drawable)(unsafe.Pointer(drawable)))
}

func Nox_xxx_clientPickup_46C140(drawable *client.Drawable) {
	gold := memmap.PtrT[[3]uint32](0x5D4594, 1064928)
	if gold[0] == 0 {
		things := &GetClient().Cli().Things
		gold[0] = uint32(things.IndByID("Gold"))
		gold[1] = uint32(things.IndByID("QuestGoldPile"))
		gold[2] = uint32(things.IndByID("QuestGoldChest"))
	}
	if drawable == nil {
		return
	}
	typeID := drawable.TypeIDVal
	if typeID == gold[0] || typeID == gold[1] || typeID == gold[2] || sub_467B00(C.int(typeID), 1) != 0 {
		packet := [3]byte{byte(netmsg.MSG_TRY_GET)}
		binary.LittleEndian.PutUint16(packet[1:], uint16(objectFieldDrawableCode578B00(drawable)))
		GetServer().S().NetList.AddToMsgListCli(server.HostPlayerIndex, netlist.Kind0, packet[:])
		return
	}
	nox_xxx_clientPlaySoundSpecial_452D80(925, 100)
	message := nox_strman_loadString_40F1D0(internCStr("pickup.c:CarryingTooMuch"), nil,
		internCStr(`C:\NoxPost\src\Client\Gui\GameWin\gamewin.c`), 83)
	nox_xxx_printCentered_445490(message)
}
