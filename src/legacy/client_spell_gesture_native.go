package legacy

/*
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "client__gui__window.h"

*/
import "C"

import (
	"image"
	"unsafe"

	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

var (
	clientGestureImages   [8]*noxrender.Image
	clientGestureProgress *server.PhonemeLeaf
)

func clientGestureFrame(index int) *uint32 {
	return g2ServerWord(1096596 + 4*uintptr(index))
}

//export nox_client_setPhonemeFrame_476E00
func nox_client_setPhonemeFrame_476E00(gesture C.int) C.int {
	if gesture >= 0 && int(gesture) < len(clientGestureImages) {
		*clientGestureFrame(int(gesture)) = gameFrame()
	}
	return gesture
}

//export sub_476E20
func sub_476E20() *nox_window {
	names := [...]string{
		"SpellGestureNW", "SpellGestureN", "SpellGestureNE", "SpellGestureW",
		"SpellGestureE", "SpellGestureSW", "SpellGestureS", "SpellGestureSE",
	}
	for i, name := range names {
		img := Nox_xxx_gLoadImg(name)
		clientGestureImages[i] = img
		if img == nil {
			return nil
		}
	}
	win := nox_window_new(nil, 64, (clientDisplayStorage.width-100)/2, (clientDisplayStorage.height-100)/2, 1, 1, nil)
	nox_window_set_all_funcs(win, nil, C.sub_476E90, nil)
	return win
}

//export sub_476E90
func sub_476E90(win *nox_window, data unsafe.Pointer) C.int {
	offsets := memmap.PtrT[[8][2]int32](0x587000, 151208)
	origin := image.Pt(int(clientDisplayStorage.width)/2-16, int(clientDisplayStorage.height)/2-41)
	for i, img := range clientGestureImages {
		start := clientGestureFrame(i)
		if *start == 0 {
			continue
		}
		offset := offsets[i]
		at := origin.Add(image.Pt(int(offset[0]), int(offset[1])))
		GetClient().R2().DrawImageAt(img, at)
		if gameFrame()-*start > 3 {
			*start = 0
		}
	}
	return 1
}

//export sub_49BB80
func sub_49BB80(id C.char) unsafe.Pointer {
	*g2ServerByte(1303504), *g2ServerByte(1303512) = byte(id), 0
	*g2ServerWord(1303516) = gameFrame()
	clientGestureProgress = GetServer().S().Spells.PhonemeTree()
	return clientGestureProgress.C()
}

//export sub_49BBB0
func sub_49BBB0() {
	*g2ServerByte(1303504) = 0
}

//export sub_49BBC0
func sub_49BBC0() {
	id := spell.ID(*g2ServerByte(1303504))
	if id == 0 {
		return
	}
	if clientGestureProgress == nil {
		sub_49BBB0()
		return
	}
	index := g2ServerByte(1303512)
	phoneme := GetServer().S().Spells.Phoneme(id, int(*index))
	if !phoneme.Valid() {
		sub_49BBB0()
		return
	}
	nextFrame := g2ServerWord(1303516)
	if gameFrame() >= *nextFrame {
		sound := nox_xxx_spellGetPhoneme_4FE1C0(C.uint32_t(ClientPlayerNetCode()), C.int8_t(phoneme))
		nox_xxx_clientPlaySoundSpecial_452D80(C.int(sound), 100)
		gesture := memmap.Uint32(0x587000, 163576+4*uintptr(phoneme))
		nox_client_setPhonemeFrame_476E00(C.int(gesture))
		*nextFrame = gameFrame() + 3
		clientGestureProgress = clientGestureProgress.Next(phoneme)
		*index = *index + 1
	}
	if clientGestureProgress == nil || clientGestureProgress.Ind == int32(id) {
		sub_49BBB0()
	}
}
