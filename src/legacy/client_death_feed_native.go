package legacy

/*
#include "GAME2_3.h"

*/
import "C"

import (
	"encoding/binary"
	"image"
	"strings"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/memmap"
)

const (
	deathFeedCapacity  = 100
	deathFeedRowHeight = 36
)

type deathFeedEntry struct {
	attacker, assistant, victim uint32
	weapon, kind, frame         uint32
}

var (
	deathFeedFallbackIcon *nox_video_bag_image_t
	deathFeedNameEscapes  = strings.NewReplacer("%%", "%", "%!", "!")
)

//export sub_495210
func sub_495210(data *C.uint8_t) C.int {
	packet := unsafe.Slice((*byte)(unsafe.Pointer(data)), 11)
	index := worldWordStorage.death_message_word_1203836
	next := (index + 1) % deathFeedCapacity
	if next == worldWordStorage.death_message_word_1203840 {
		worldWordStorage.death_message_word_1203840 = (worldWordStorage.death_message_word_1203840 + 1) % deathFeedCapacity
	}
	read := binary.LittleEndian.Uint16
	if packet[10] == 1 {
		switch uint32(read(packet[8:10])) {
		case memmap.Uint32(0x5D4594, 1203844):
			binary.LittleEndian.PutUint16(packet[8:10], memmap.Uint16(0x5D4594, 1203856))
		case memmap.Uint32(0x5D4594, 1203848):
			binary.LittleEndian.PutUint16(packet[8:10], memmap.Uint16(0x5D4594, 1203852))
		}
	}
	ring := memmap.PtrT[[deathFeedCapacity]deathFeedEntry](0x5D4594, 1201428)
	ring[index] = deathFeedEntry{
		attacker:  uint32(read(packet[2:4])),
		assistant: uint32(read(packet[4:6])),
		victim:    uint32(read(packet[6:8])),
		weapon:    uint32(read(packet[8:10])),
		kind:      uint32(packet[10]),
		frame:     gameFrame(),
	}
	worldWordStorage.death_message_word_1203836 = next
	words := [4]C.uint16_t{0, C.uint16_t(ring[index].attacker), C.uint16_t(ring[index].assistant), C.uint16_t(ring[index].victim)}
	return sub_4952E0(&words[0])
}

//export sub_495430
func sub_495430() C.int {
	legacyWords.dword_5d4594_1203832 = 0
	nox_client_drawEnableAlpha_434560(0)
	sub_4345F0(0)
	nox_xxx_draw_434600(0)
	ring := memmap.PtrT[[deathFeedCapacity]deathFeedEntry](0x5D4594, 1201428)
	maxRow := C.uint32_t(clientDisplayStorage.height / 4 / deathFeedRowHeight)
	oldest := worldWordStorage.death_message_word_1203840
	for index := oldest; index != worldWordStorage.death_message_word_1203836; index = (index + 1) % deathFeedCapacity {
		if legacyWords.dword_5d4594_1203832 > maxRow {
			break
		}
		if gameFrame()-ring[oldest].frame > 90 {
			oldest = (oldest + 1) % deathFeedCapacity
			worldWordStorage.death_message_word_1203840 = oldest
			continue
		}
		drawDeathFeedEntry(&ring[index])
		oldest = worldWordStorage.death_message_word_1203840
		legacyWords.dword_5d4594_1203832++
	}
	return C.int(worldWordStorage.death_message_word_1203836)
}

//export sub_495500
func sub_495500(entry *C.int) C.uintptr_t {
	return C.uintptr_t(drawDeathFeedEntry((*deathFeedEntry)(unsafe.Pointer(entry))))
}

func drawDeathFeedEntry(entry *deathFeedEntry) uintptr {
	r := GetClient().R2()
	face := r.GetFonts().FontByName("large")
	name := func(netCode uint32) (string, bool) {
		if netCode == 0 {
			return "", false
		}
		p := asPlayerS(nox_common_playerInfoGetByID_417040(int32(netCode)))
		if p == nil {
			return "", false
		}
		return p.Name(), true
	}
	attacker, hasAttacker := name(entry.attacker)
	attacker = deathFeedNameEscapes.Replace(attacker)
	assistant, hasAssistant := name(entry.assistant)
	if hasAttacker && hasAssistant {
		assistant = "+" + assistant
	} else {
		assistant = deathFeedNameEscapes.Replace(assistant)
	}
	victim, _ := name(entry.victim)
	victim = deathFeedNameEscapes.Replace(victim)

	icon := deathFeedIcon(entry)
	var iconOffset, iconSize image.Point
	if icon != nil {
		if offset, size, ok := icon.Meta(); ok {
			iconOffset, iconSize = offset, size
		}
	}
	width := r.GetStringSizeWrapped(face, attacker, 0).X
	width += r.GetStringSizeWrapped(face, assistant, 0).X
	victimSize := r.GetStringSizeWrapped(face, victim, 0)
	width += iconSize.X + victimSize.X
	left := (int(clientDisplayStorage.width) - width - 10) / 2
	top := int(legacyWords.dword_5d4594_1203832) * deathFeedRowHeight
	r.DrawRectFilledAlpha(left-5, top, width+20, deathFeedRowHeight)
	textY := top + (deathFeedRowHeight-victimSize.Y)/2
	drawName := func(code *uint32, text string, x int) int {
		if *code == 0 {
			return x
		}
		nox_xxx_drawSetTextColor_434390(memmap.Int32(0x5D4594, 2597996))
		if *code == uint32(ClientPlayerNetCode()) {
			nox_xxx_drawSetTextColor_434390(int32(worldWordStorage.local_player_color))
		}
		return r.DrawString(face, text, image.Pt(x, textY))
	}
	left = drawName(&entry.attacker, attacker, left)
	left = drawName(&entry.assistant, assistant, left)
	left += 5
	if icon != nil {
		r.DrawImageAt(icon, image.Pt(left-iconOffset.X, top+(deathFeedRowHeight-iconSize.Y)/2-iconOffset.Y))
	}
	if entry.victim == 0 {
		return uintptr(unsafe.Pointer(entry))
	}
	return uintptr(int32(drawName(&entry.victim, victim, left+iconSize.X+5)))
}

func deathFeedIcon(entry *deathFeedEntry) *noxrender.Image {
	var icon *nox_video_bag_image_t
	switch entry.kind {
	case 1:
		thing := GetClient().Cli().Things.TypeByInd(int(entry.weapon))
		if thing != nil {
			if thing.ObjClass.HasAny(object.ClassWeapon | object.ClassWand) {
				sub_4B9650(C.int(entry.weapon))
			}
			icon = (*nox_video_bag_image_t)(unsafe.Pointer(uintptr(thing.PrettyImage)))
		}
	case 2:
		switch entry.weapon {
		case 1, 12:
			icon = (*nox_video_bag_image_t)(nox_xxx_spellIcon_424A90(5))
		case 2:
			icon = nox_xxx_spellGetAbilityIcon_425310(1, 0)
		case 4:
			icon = (*nox_video_bag_image_t)(nox_xxx_spellIcon_424A90(130))
		case 5:
			icon = (*nox_video_bag_image_t)(nox_xxx_spellIcon_424A90(60))
		case 9, 17:
			icon = (*nox_video_bag_image_t)(nox_xxx_spellIcon_424A90(43))
		case 15:
			icon = (*nox_video_bag_image_t)(nox_xxx_spellIcon_424A90(56))
		case 16:
			icon = (*nox_video_bag_image_t)(nox_xxx_spellIcon_424A90(16))
		}
	}
	if icon == nil {
		icon = deathFeedFallbackIcon
	}
	return asImage(icon)
}

//export sub_4958F0
func sub_4958F0() C.int {
	worldWordStorage.death_message_word_1203840 = 0
	worldWordStorage.death_message_word_1203836 = 0
	types := [...]string{"ArcherBolt", "ArcherArrow", "Bow", "CrossBow"}
	for i, typeName := range types {
		id := memmap.PtrT[uint32](0x5D4594, 1203844+uintptr(i)*4)
		if *id == 0 {
			*id = uint32(GetClient().Cli().Things.IndByID(typeName))
		}
	}
	deathFeedFallbackIcon = (*nox_video_bag_image_t)(nox_xxx_spellIcon_424A90(15))
	legacyIcon := uint32(uintptr(unsafe.Pointer(deathFeedFallbackIcon)))
	*memmap.PtrT[uint32](0x5D4594, 1203828) = legacyIcon
	return C.int(legacyIcon)
}
