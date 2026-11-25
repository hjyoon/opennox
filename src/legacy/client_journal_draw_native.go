package legacy

/*
#include "client__gui__guijourn.h"
*/
import "C"

import (
	"image"

	noxcolor "github.com/opennox/libs/color"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

const (
	clientJournalWidth     = 260
	clientJournalHeight    = 150
	clientJournalTextInset = 10
	clientJournalTextWidth = clientJournalWidth - 2*clientJournalTextInset
)

func clientJournalText(sm *strman.StringManager, entry *server.PlayerJournal) string {
	const file = `C:\NoxPost\src\client\Gui\GUIJourn.c`
	var label strman.ID
	switch entry.Field3 {
	case 2:
		label = "Journal:QuestLabel"
	case 4:
		label = "Journal:CompletedLabel"
	case 8:
		label = "Journal:HintLabel"
	}
	var prefix string
	if label != "" {
		prefix = sm.GetStringInFile(label, file)
	}
	key := strman.ID("Journal:" + alloc.GoStringS(entry.EntryBuf[:]))
	return prefix + " " + sm.GetStringInFile(key, file)
}

func clientJournalColor(kind uint16) noxcolor.RGBA5551 {
	switch kind {
	case 1:
		return noxcolor.RGBA5551(Get_nox_color_white_2523948())
	case 2:
		return noxcolor.RGBA5551(memmap.Uint32(0x85B3FC, 940))
	case 4:
		return noxcolor.RGBA5551(memmap.Uint32(0x85B3FC, 956))
	case 8:
		return noxcolor.RGBA5551(Get_nox_color_yellow_2589772())
	default:
		return noxcolor.RGBA5551(Get_dword_8531A0_2572())
	}
}

//export nox_xxx_cliBuildJournalString_469BC0
func nox_xxx_cliBuildJournalString_469BC0() {
	player := Get_dword_8531A0_2576()
	if player == nil {
		return
	}
	render := GetClient().R2()
	strings := GetServer().S().Strings()
	fontHeight := render.FontHeight(nil)
	height := -fontHeight
	for entry := player.Journal; entry != nil; entry = entry.Next {
		text := clientJournalText(strings, entry)
		height += fontHeight + render.GetStringSizeWrapped(nil, text, clientJournalTextWidth).Y
	}
	*memmap.PtrUint32(0x5D4594, 1064848) = uint32(max(0, height))
}

//export nox_xxx_guiDrawJournal_469D40
func nox_xxx_guiDrawJournal_469D40(left, top, scroll C.int) {
	player := Get_dword_8531A0_2576()
	if player == nil {
		return
	}
	x, y := int(left), int(top)
	render := GetClient().R2()
	black := noxcolor.RGBA5551(Get_nox_color_black_2650656())
	render.Data().SetColor2(black)
	render.DrawRectFilledOpaque(x, y, clientJournalWidth, clientJournalHeight, black)
	entry := player.Journal
	if entry == nil {
		return
	}
	for entry.Next != nil {
		entry = entry.Next
	}
	sm := GetServer().S().Strings()
	fontHeight := render.FontHeight(nil)
	drawY := y - int(scroll)
	textX := x + clientJournalTextInset
	for ; entry != nil; entry = entry.Prev {
		text := clientJournalText(sm, entry)
		bottom := drawY + render.GetStringSizeWrapped(nil, text, clientJournalTextWidth).Y
		if bottom > y {
			render.Data().SetTextColor(clientJournalColor(entry.Field3))
			render.DrawStringWrapped(nil, text, image.Rect(textX, drawY, textX+clientJournalTextWidth, drawY))
		}
		drawY = bottom + fontHeight
		if drawY > y+clientJournalHeight {
			break
		}
	}
}
