package legacy

/*
#include <stdlib.h>
#include "GAME5_2.h"

typedef const wchar2_t rule_const_wchar2_t;

*/
import "C"

import (
	"bytes"
	"math/bits"
	"unsafe"

	"github.com/opennox/libs/ifs"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

const mapRuleLineWidth = 256

//export sub_57A3F0
func sub_57A3F0(filename, data *C.char, notes *C.int, mode C.int) C.int {
	settings := (*server.Settings2)(unsafe.Pointer(data))
	list := (*intrusiveListNode425760)(unsafe.Pointer(notes))
	return C.int(bool2int(readMapRuleFile(GoString(filename), settings, list, noxflags.GameFlag(uint32(mode)))))
}

func readMapRuleFile(path string, settings *server.Settings2, notes *intrusiveListNode425760, mode noxflags.GameFlag) bool {
	legacyWords.dword_5d4594_2523764 = C.uint32_t(noxflags.GameModeMask)
	r, err := ifs.Open(path)
	if err != nil {
		return false
	}
	file := binfile.NewTextFile(r)
	defer file.Close()

	var wide [mapRuleLineWidth - 1]uint16
	for {
		data, err := file.ReadStringLimit(mapRuleLineWidth - 1)
		data, _, _ = bytes.Cut(data, []byte{0})
		data = bytes.TrimSuffix(data, []byte{'\n'})
		for i, b := range data {
			wide[i] = uint16(b)
		}
		if len(data) != 0 {
			readMapRuleLine(wide[:len(data)], settings, notes, mode)
		}
		if err != nil {
			return true
		}
	}
}

//export sub_57A4D0
func sub_57A4D0(line *C.wchar2_t, data *C.char, notes *C.int, mode C.int) {
	settings := (*server.Settings2)(unsafe.Pointer(data))
	list := (*intrusiveListNode425760)(unsafe.Pointer(notes))
	readMapRuleLine(mapRuleWide(unsafe.Pointer(line)), settings, list, noxflags.GameFlag(uint32(mode)))
}

func readMapRuleLine(line []uint16, settings *server.Settings2, notes *intrusiveListNode425760, mode noxflags.GameFlag) {
	sub_416580()
	tokens := splitMapRuleLine(line)
	if len(tokens) == 0 || applyMapRuleTokens(tokens, settings, mode) || notes == nil {
		return
	}
	size := unsafe.Sizeof(C.nox_rule_list_item_t{})
	ptr, _ := alloc.Calloc(1, size)
	if ptr == nil {
		return
	}
	note := (*C.nox_rule_list_item_t)(ptr)
	copy(unsafe.Slice((*uint16)(unsafe.Pointer(&note.text[0])), mapRuleLineWidth-1), line)
	intrusiveListBefore4258E0(notes, (*intrusiveListNode425760)(ptr))
}

func mapRuleWide(ptr unsafe.Pointer) []uint16 {
	n := alloc.StrLenN((*uint16)(ptr), mapRuleLineWidth-1)
	return unsafe.Slice((*uint16)(ptr), n)
}

func splitMapRuleLine(line []uint16) [][]uint16 {
	var tokens [][]uint16
	pos := 0
	quoted := len(line) > 0 && line[0] == '"'
	if quoted {
		pos++
	}
	for pos < len(line) {
		for pos < len(line) && mapRuleSeparator(line[pos], quoted) {
			pos++
		}
		start := pos
		for pos < len(line) && !mapRuleSeparator(line[pos], quoted) {
			pos++
		}
		if start == pos {
			break
		}
		tokens = append(tokens, line[start:pos])
		if pos == len(line) {
			break
		}
		pos++

		// Stock checks for a quote before skipping separators. A quoted token
		// advances the quote check by one extra code unit.
		look := pos
		if quoted {
			look++
		}
		quoted = look < len(line) && line[look] == '"'
		if quoted {
			pos = look + 1
		}
	}
	return tokens
}

func mapRuleSeparator(ch uint16, quoted bool) bool {
	if quoted {
		return ch == '"' || ch == '\n' || ch == '\r'
	}
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func mapRuleKeyword(token []uint16, keyword string) bool {
	if len(token) != len(keyword) {
		return false
	}
	for i, ch := range token {
		if ch >= 'A' && ch <= 'Z' {
			ch += 'a' - 'A'
		}
		if ch != uint16(keyword[i]) {
			return false
		}
	}
	return true
}

func mapRuleNarrow(token []uint16) string {
	// nox_sprintf("%S") stores the low byte of each UTF-16 code unit.
	narrow := make([]byte, 0, len(token))
	for _, ch := range token {
		if byte(ch) == 0 {
			break
		}
		narrow = append(narrow, byte(ch))
	}
	return string(narrow)
}

//export sub_57A620
func sub_57A620(count C.uchar, args **C.rule_const_wchar2_t, data *C.char, mode C.int) C.int {
	tokens := make([][]uint16, int(count))
	for i, p := range unsafe.Slice(args, len(tokens)) {
		tokens[i] = mapRuleWide(unsafe.Pointer(p))
	}
	settings := (*server.Settings2)(unsafe.Pointer(data))
	return C.int(bool2int(applyMapRuleTokens(tokens, settings, noxflags.GameFlag(uint32(mode)))))
}

func applyMapRuleTokens(tokens [][]uint16, settings *server.Settings2, mode noxflags.GameFlag) bool {
	if len(tokens) == 0 {
		return false
	}
	if section := mapRuleModeByName(mapRuleNarrow(tokens[0])); section != 0 {
		legacyWords.dword_5d4594_2523764 = C.uint32_t(section)
		return mode == section
	}
	active := noxflags.GameFlag(legacyWords.dword_5d4594_2523764)
	if active&mode == 0 || len(tokens) != 4 || !mapRuleKeyword(tokens[0], "set") {
		return false
	}
	off := mapRuleKeyword(tokens[3], "off")
	switch {
	case mapRuleKeyword(tokens[1], "spell"):
		if !applyMapRuleSpell(settings, tokens[2], off) {
			return false
		}
	case mapRuleKeyword(tokens[1], "weapon"):
		if !noxflags.HasGame(noxflags.GameHost) || !applyMapRuleWeapon(settings, tokens[2], off) {
			return false
		}
	case mapRuleKeyword(tokens[1], "armor"):
		if !noxflags.HasGame(noxflags.GameHost) || !applyMapRuleArmor(settings, tokens[2], off) {
			return false
		}
	default:
		return false
	}
	return mode == noxflags.GameFlag(legacyWords.dword_5d4594_2523764)
}

func applyMapRuleSpell(settings *server.Settings2, token []uint16, off bool) bool {
	name := CString(mapRuleNarrow(token))
	defer StrFree(name)
	spell := nox_xxx_spellNameToN_4243F0(name)
	if spell == 0 {
		title, free := CWString(GoWStringSlice(token))
		defer free()
		spell = nox_xxx_spellByTitle_424960(title)
	}
	if spell == 0 || !nox_xxx_spellIsValid_424B50(spell) {
		return false
	}
	if nox_xxx_spellFlags_424A70(spell)&0x7000000 != 0 && off {
		sub_453FA0(unsafe.Pointer(&settings.Field24), C.int(spell), 0)
	}
	return true
}

func applyMapRuleWeapon(settings *server.Settings2, token []uint16, off bool) bool {
	name := CString(mapRuleNarrow(token))
	defer StrFree(name)
	idx := sub_415A30(name)
	if idx < 0 {
		return false
	}
	mask := nox_xxx_ammoCheck_415880(idx)
	if mask == 0 {
		return false
	}

	// The original signed shift loop starts one byte before the weapon bits.
	// Positive masks advance it to the byte containing the highest set bit.
	offset := -1
	if mask > 0 {
		offset = (bits.Len32(uint32(mask)) - 1) / 8
		mask >>= 8 * offset
	}
	slot := (*byte)(unsafe.Add(unsafe.Pointer(&settings.Field44[0]), offset))
	if off {
		*slot &^= byte(mask)
	} else {
		*slot |= byte(mask)
	}
	return true
}

func applyMapRuleArmor(settings *server.Settings2, token []uint16, off bool) bool {
	name := CString(mapRuleNarrow(token))
	defer StrFree(name)
	idx := sub_415EC0(name)
	if idx < 0 {
		return false
	}
	mask := uint32(sub_415D10(idx))
	if mask == 0 {
		return false
	}
	if off {
		settings.Field48 &^= mask
	} else {
		settings.Field48 |= mask
	}
	return true
}

//export sub_57ADF0
func sub_57ADF0(notes *C.int) *C.int {
	first := clearMapRuleNotes((*intrusiveListNode425760)(unsafe.Pointer(notes)))
	return (*C.int)(unsafe.Pointer(first))
}

func clearMapRuleNotes(notes *intrusiveListNode425760) *intrusiveListNode425760 {
	first := intrusiveListFirstSafe425890(notes)
	for item := first; item != nil; {
		next := intrusiveListNextSafe4258A0(item)
		intrusiveListRemove425920(item)
		if ptr := unsafe.Pointer(item); !alloc.TryFreePtr(ptr) {
			C.free(ptr)
		}
		item = next
	}
	return first
}
