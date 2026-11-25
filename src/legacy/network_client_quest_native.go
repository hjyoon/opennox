package legacy

/*
#include "client__gui__guiggovr.h"
#include "cgo_const.h"
*/
import "C"

import (
	"bytes"
	"encoding/binary"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

var clientQuestPacketSizes = [...]uint8{
	0: 2, 1: 4, 2: 14,
	4: 5, 5: 4, 6: 4, 7: 4, 8: 4, 9: 4, 10: 4, 11: 16,
	12: 90, 13: 69, 14: 69, 15: 4, 16: 12, 17: 4, 18: 4, 19: 4,
	20: 2, 21: 8, 22: 5, 23: 5, 24: 3, 25: 7, 26: 6,
	28: 3, 29: 4, 30: 5, 31: 5, 32: 5, 33: 52,
}

func clientQuestSpriteType(offset uintptr, name string) uint32 {
	slot := memmap.PtrUint32(0x5D4594, offset)
	if *slot == 0 {
		*slot = uint32(nox_xxx_getTTByNameSpriteMB_44CFC0(internCStr(name)))
	}
	return *slot
}

func clientQuestGreenZap(data []byte) {
	typeID := clientQuestSpriteType(1200904, "GreenZap")
	if nox_client_isConnected_43C700() == 0 {
		return
	}
	x := binary.LittleEndian.Uint16(data[6:8])
	y := binary.LittleEndian.Uint16(data[8:10])
	drawable := nox_xxx_spriteLoadAdd_45A360_drawable(int32(typeID), int32(x), int32(y))
	if drawable == nil {
		return
	}
	payload := drawableEffectBytes(asDrawable(drawable))
	payload[0] = 0
	duration := binary.LittleEndian.Uint16(data[10:12])
	binary.LittleEndian.PutUint32(payload[1:5], uint32(duration))
	copy(payload[5:13], data[2:10])
	nox_xxx_spriteTransparentDecay_49B950(drawable, int32(duration))
}

func clientQuestFireBurst(data []byte) {
	if memmap.Uint32(0x5D4594, 1200908) == 0 {
		*memmap.PtrUint32(0x5D4594, 1200908) = uint32(nox_xxx_getTTByNameSpriteMB_44CFC0(internCStr("GreenSpark")))
		*memmap.PtrUint32(0x5D4594, 1200912) = uint32(nox_xxx_getTTByNameSpriteMB_44CFC0(internCStr("FireBoom")))
	}
	x := int16(binary.LittleEndian.Uint16(data[2:4]))
	y := int16(binary.LittleEndian.Uint16(data[4:6]))
	if drawable := nox_xxx_spriteLoadAdd_45A360_drawable(memmap.Int32(0x5D4594, 1200912), int32(x), int32(y)); drawable != nil {
		nox_xxx_sprite_45A110_drawable(drawable)
	}
	if nox_client_isConnected_43C700() != 0 {
		pos := C.int2{field_0: C.int(x), field_4: C.int(y)}
		sub_49A150(&pos, C.int(memmap.Int32(0x5D4594, 1200908)), C.uchar(data[6]))
	}
}

func clientQuestAwardNotice(data []byte) {
	id := int32(binary.LittleEndian.Uint16(data[3:5]))
	player := nox_common_playerInfoGetByID_417040(id)
	if nox_client_isConnected_43C700() == 0 || player == nil {
		return
	}
	self := asPlayerS(player) == Get_dword_8531A0_2576()
	value := int32(data[2])
	var title *wchar2_t
	var key string
	var line int32
	switch data[1] {
	case 30:
		title = nox_xxx_spellTitle_424930(value)
		key, line = "plyrspel.c:AwardSpell", 8112
		if !self {
			key, line = "plyrspel.c:AwardSpellToOther", 8115
		}
	case 31:
		title = nox_xxx_guiCreatureGetName_427240(C.int(value))
		key, line = "PlyrGide.c:AwardGuide", 8145
		if !self {
			key, line = "PlyrGide.c:AwardGuideToOther", 8148
		}
	case 32:
		title = nox_xxx_abilityGetName_0_425260(value)
		key, line = "ComAblty.c:AwardAbility", 8178
		if !self {
			key, line = "ComAblty.c:AwardAbilityToOther", 8181
		}
	}
	format := clientDecodeString(key, line)
	var text [256]wchar2_t
	if self {
		formatWideStrings(text[:], format, title, nil)
	} else {
		formatWideStrings(text[:], format, &player.name_final[0], title)
	}
	nox_xxx_printCentered_445490(&text[0])
}

func clientQuestKeyNotice(data []byte) {
	keyOffsets := [...]uintptr{160948, 160988, 161028, 161068, 161112}
	if int(data[51]) >= len(keyOffsets) {
		return
	}
	item := (*C.char)(memmap.PtrOff(0x587000, keyOffsets[data[51]]))
	title := nox_strman_loadString_40F1D0(item, nil, internCStr(clientDecodeStringsFile), 8217)
	key := data[2:51]
	if n := bytes.IndexByte(key, 0); n >= 0 {
		key = key[:n]
	}
	format := clientDecodeString(string(key), 8216)
	var text [256]wchar2_t
	formatWideStrings(text[:], format, title, nil)
	nox_xxx_printCentered_445490(&text[0])
}

func clientQuestPacket(data []byte) int {
	if len(data) < 2 || int(data[1]) >= len(clientQuestPacketSizes) {
		return -1
	}
	subtype := data[1]
	size := int(clientQuestPacketSizes[subtype])
	if size == 0 || len(data) < size {
		return -1
	}
	switch subtype {
	case 30, 31, 32:
		clientQuestAwardNotice(data)
		return size
	case 16:
		clientQuestGreenZap(data)
		return size
	case 25:
		clientQuestFireBurst(data)
		return size
	}
	connected := nox_client_isConnected_43C700() != 0
	if subtype >= 12 && subtype <= 14 {
		if connected {
			packet := (*C.nox_cgo_const_uchar)(unsafe.Pointer(&data[0]))
			switch subtype {
			case 12:
				nox_xxx_clientQuestWinScreen_450770(packet)
			case 13:
				nox_client_showQuestBriefing2_450980(packet, C.int(data[4]&1))
			case 14:
				nox_client_showQuestBriefing_450A30(packet, 1)
			}
		}
		Set_nox_client_gui_flag_1556112(0)
		return size
	}
	if !connected {
		return size
	}
	word := func(index int) uint16 { return binary.LittleEndian.Uint16(data[index : index+2]) }
	switch subtype {
	case 0:
		if g2ServerGame(1) {
			sub_460380()
			nox_xxx_cliPrepareGameplay1_460E60()
			path := (*C.char)(memmap.PtrOff(0x85B3FC, 10984))
			sub_41CC00(path)
			nox_xxx_plrLoad_41A480(path)
		}
		nox_xxx_cliShowHideTubes_470AA0(1)
		sub_48D4A0()
		nox_xxx_cliPrepareGameplay2_4721D0()
		sub_4705B0()
	case 1:
		nox_xxx_clientPlaySoundSpecial_452D80(1008, 100)
		if player := nox_common_playerInfoGetByID_417040(int32(word(2))); player != nil && !g2ServerGame(1) {
			player.field_4792 = 1
		}
	case 2:
		sub_49B4B0((*C.ushort)(unsafe.Pointer(&data[0])))
	case 4:
		code := word(3)
		if player := nox_common_playerInfoGetByID_417040(int32(code)); player != nil {
			*(*byte)(unsafe.Pointer(&player.tail_padding[0])) = data[2]
		}
		if uint32(code) == uint32(ClientPlayerNetCode()) {
			sub_463420(C.int(data[2]))
		}
	case 15:
		if drawable := nox_xxx_netSpriteByCodeStatic_45A720(int32(word(2))); drawable != nil {
			drawableEffectBytes(asDrawable(drawable))[0] = 0
		}
	case 17:
		sub_45D320(C.int(word(2)))
	case 18:
		game2_book_disable_ability_word(C.int(word(2)))
	case 19:
		sub_45D400(C.int(word(2)))
	case 20:
		if sub_470580() == 0 {
			sub_4705B0()
		}
	case 21:
		if player := nox_common_playerInfoGetByID_417040(int32(word(6))); player != nil {
			player.tail_padding[1] = C.uint32_t(binary.LittleEndian.Uint32(data[2:6]))
		}
	case 22, 23:
		if player := nox_common_playerInfoGetByID_417040(int32(word(3))); player != nil {
			*(*byte)(unsafe.Add(unsafe.Pointer(&player.tail_padding[0]), int(subtype)-14)) = data[2]
		}
	case 24:
		serverKeyVisibleRequest(uintptr(data[2]))
	case 26:
		nox_xxx_makePointFxCli_499610(C.int(memmap.Int32(0x5D4594, 1200788)), 25, 500, 25, C.int(int16(word(2))), C.int(int16(word(4))))
	case 28:
		sub_41D1A0(C.int(data[2]))
	case 29:
		sub_465DE0(C.int(word(2)))
	case 33:
		clientQuestKeyNotice(data)
	}
	return size
}
