package legacy

import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
)

func clientProgressPacket(op netmsg.Op, data []byte) (int, bool) {
	var size int
	switch op {
	case netmsg.MSG_RESET_ABILITIES, netmsg.MSG_REPORT_CREATURE_CMD:
		size = 2
	case netmsg.MSG_REPORT_EXPERIENCE:
		size = 5
	case netmsg.MSG_REPORT_SPELL_AWARD:
		size = 4
	case netmsg.MSG_REPORT_SPELL_STAT:
		size = 6
	case netmsg.MSG_REPORT_ABILITY_AWARD, netmsg.MSG_REPORT_ABILITY_STATE,
		netmsg.MSG_REPORT_ACTIVE_ABILITIES, netmsg.MSG_REPORT_GUIDE_AWARD, netmsg.MSG_CHAPTER_END,
		netmsg.MSG_FX_GENERATING_MAP, netmsg.MSG_FX_ASSEMBLING_MAP, netmsg.MSG_FX_POPULATING_MAP:
		size = 3
	default:
		return 0, false
	}
	if len(data) < size {
		return 0, true
	}
	if op != netmsg.MSG_REPORT_SPELL_STAT && nox_client_isConnected_43C700() == 0 {
		return size, true
	}
	switch op {
	case netmsg.MSG_REPORT_EXPERIENCE:
		sub_467440(C.int(binary.LittleEndian.Uint32(data[1:5])))
	case netmsg.MSG_REPORT_SPELL_AWARD:
		flags := data[3]
		nox_xxx_netSpellRewardCli_45CFE0(C.int(data[1]), C.int(data[2]), C.int(flags&0x7f), C.int(flags>>7))
	case netmsg.MSG_REPORT_ABILITY_AWARD:
		flags := data[2]
		clientAbilityReward(C.int(data[1]), C.int(flags&0x7f), flags&0x80 != 0)
	case netmsg.MSG_REPORT_ABILITY_STATE:
		sub_461090(C.int(data[1]), C.int(data[2]))
	case netmsg.MSG_REPORT_ACTIVE_ABILITIES:
		sub_461120(C.int(data[1]), C.int(data[2]))
	case netmsg.MSG_REPORT_GUIDE_AWARD:
		nox_xxx_netGuideRewardCli_45D140(C.int(data[1]), C.int(data[2]&0x7f))
	case netmsg.MSG_CHAPTER_END:
		nox_client_lockScreenBriefing_450160(C.int(data[1]), C.int(data[2]), 0)
	case netmsg.MSG_RESET_ABILITIES:
		sub_4610D0(C.uchar(data[1]))
	case netmsg.MSG_REPORT_CREATURE_CMD:
		sub_4C1CA0(C.int(data[1]))
	case netmsg.MSG_REPORT_SPELL_STAT:
		sub_460EB0(C.int(binary.LittleEndian.Uint32(data[1:5])), C.char(data[5]))
	default:
		nox_xxx_mapGenClientText_4A9D00((*C.uchar)(unsafe.Pointer(&data[0])))
	}
	return size, true
}
