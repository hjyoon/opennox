package legacy

/*
#include "game2_server_bridge.h"
*/
import "C"
import "unsafe"

//export sub_4532E0
func sub_4532E0() *C.nox_window {
	w, root := C.dword_5d4594_1045464, C.dword_5d4594_1045468
	if w == nil || root == nil {
		return nil
	}
	h := C.nox_xxx_guiFontHeightMB_43F320(w.draw_data.font) + 1
	C.sub_46AB20(w, C.int(w.width), 15*h+2)
	y := w.off_y + C.intptr_t(h) + 2
	var last *C.nox_window
	for id := 1520; id < 1534; id++ {
		last = g2ServerChild(root, id)
		if last != nil {
			last.off_y = y
		}
		y += C.intptr_t(h)
	}
	return last
}

//export sub_453350
func sub_453350(w *C.nox_window, d *C.nox_window_data) C.int {
	if w == nil || d == nil {
		return 0
	}
	return g2ServerDrawBackground(w, d, true)
}

//export sub_4533D0
func sub_4533D0(unused1 *C.nox_window, event C.int, control *C.nox_window, unused2 C.uintptr_t) C.int {
	root, w := C.dword_5d4594_1045468, C.dword_5d4594_1045464
	if root == nil || w == nil || control == nil {
		return 0
	}
	id := int(control.id)
	if event == 0x4000 {
		if control == g2ServerChild(root, 1513) || control == g2ServerChild(root, 1514) {
			g2ServerText(w, 0x4000, unsafe.Pointer(control), 0)
			sub_453750()
		}
		return 0
	}
	if event != 16391 {
		return 0
	}
	switch id {
	case 1513, 1514:
		g2ServerText(w, 0x4000, unsafe.Pointer(control), 0)
		sub_453750()
		return 0
	case 1515, 1516:
		off := uintptr(1045452)
		if C.dword_5d4594_1045460 != 0 {
			off = 1045456
		}
		*g2ServerWord(off) = 0
		if id == 1515 {
			*g2ServerWord(off) = 0xFFFFFFFF
		}
		sub_453750()
		sub_459D50(1)
	default:
		if id >= 1520 && id <= 1533 {
			d := g2ServerList(w)
			i := g2ServerFirstVisible(d) + id - 1520
			if d == nil || d.Items == nil || i < 0 || i >= int(d.Field_11_0) || g2ServerItems(d)[i].Text[0] == 0 {
				break
			}
			name := (*C.wchar2_t)(unsafe.Pointer(&g2ServerItems(d)[i].Text[0]))
			var flag C.int
			if C.dword_5d4594_1045460 != 0 {
				flag = C.sub_415DA0(name)
			} else {
				flag = C.sub_415960(name)
			}
			if flag == 0 {
				break
			}
			enabled := g2ServerBool(control.draw_data.field_0&4 == 0)
			if C.dword_5d4594_1045460 != 0 {
				sub_453640((*C.uint32_t)(g2ServerMem(1045456)), flag, enabled)
			} else {
				i := uintptr(0)
				for flag > 255 {
					flag >>= 8
					i++
				}
				sub_453620((*C.uint8_t)(g2ServerMem(1045452+i)), C.char(flag), enabled)
			}
			sub_459D50(1)
		}
	}
	C.nox_xxx_clientPlaySoundSpecial_452D80(766, 100)
	return 0
}

//export sub_4535E0
func sub_4535E0(p *C.int) *C.int { *g2ServerWord(1045452) = uint32(*p); return p }

//export sub_4535F0
func sub_4535F0(v C.int) C.int { *g2ServerWord(1045456) = uint32(v); return v }

//export sub_453600
func sub_453600() *C.char { return (*C.char)(g2ServerMem(1045452)) }

//export sub_453610
func sub_453610() C.int { return C.int(*g2ServerWord(1045456)) }

//export sub_453620
func sub_453620(p *C.uint8_t, mask C.char, set C.int) *C.uint8_t {
	if set != 0 {
		*p |= C.uint8_t(mask)
	} else {
		*p &^= C.uint8_t(mask)
	}
	return p
}

//export sub_453640
func sub_453640(p *C.uint32_t, mask, set C.int) *C.uint32_t {
	if set != 0 {
		*p |= C.uint32_t(mask)
	} else {
		*p &^= C.uint32_t(mask)
	}
	return p
}

//export sub_453660
func sub_453660(v C.int) C.int {
	index := uintptr(0)
	word := int32(v)
	if word > 0 {
		for {
			next := word >> 8
			if next > 0 {
				word = next
			}
			index++
			if next <= 0 {
				break
			}
		}
	}
	return g2ServerBool(byte(word)&*g2ServerByte(1045451 + index) != 0)
}

//export sub_453690
func sub_453690(v C.int) C.int { return g2ServerBool(uint32(v)&*g2ServerWord(1045456) != 0) }

//export sub_4536B0
func sub_4536B0(p *C.uint32_t) C.int {
	*p = 0xFFFFFFFF
	flag := uint32(1)
	bit := byte(1)
	off := uintptr(0)
	var result C.int
	for i := 0; i < 27; i++ {
		result = C.sub_415840(C.int(flag))
		if result != 0 {
			result = C.nox_xxx_getUnitDefDd10_4E3BA0(result)
			if result == 0 {
				result = C.int(^bit)
				*(*byte)(unsafe.Add(unsafe.Pointer(p), off)) &^= bit
			}
		}
		if bit == 128 {
			bit = 1
			off++
		} else {
			bit *= 2
		}
		flag *= 2
	}
	return result
}

//export sub_453710
func sub_453710() C.int {
	value := uint32(0xFFFFFFFF)
	flag := uint32(1)
	for i := 0; i < 26; i++ {
		v := C.sub_415CD0(C.int(flag))
		if v != 0 && C.nox_xxx_getUnitDefDd10_4E3BA0(v) == 0 {
			value &^= flag
		}
		flag *= 2
	}
	return C.int(value)
}

//export sub_453750
func sub_453750() C.int {
	w, root := C.dword_5d4594_1045464, C.dword_5d4594_1045468
	if w == nil || root == nil {
		return 0
	}
	index := g2ServerFirstVisible(g2ServerList(w))
	count := int32(*g2ServerWord(1045472 + 4*uintptr(C.dword_5d4594_1045460)))
	var result C.int
	for id := 1520; id <= 1533; id++ {
		box := g2ServerChild(root, id)
		for index >= 0 && index < 32 && (uint32(1)<<uint(index))&0x33 != 0 {
			index++
		}
		if box == nil {
			index++
			continue
		}
		if index < 0 || index >= int(count) || index >= 32 {
			result = g2ServerHidden(box, true)
		} else {
			g2ServerHidden(box, false)
			if C.sub_453080(C.char(index)) != 0 {
				box.draw_data.field_0 |= 4
			} else {
				box.draw_data.field_0 &^= 4
			}
			result = C.int(box.draw_data.field_0)
		}
		index++
	}
	return result
}

//export sub_453B00
func sub_453B00() *C.nox_window {
	w, other, root := C.dword_5d4594_1045480, C.dword_5d4594_1045508, C.dword_5d4594_1045484
	if w == nil || other == nil || root == nil {
		return nil
	}
	h := C.nox_xxx_guiFontHeightMB_43F320(w.draw_data.font) + 1
	C.sub_46AB20(w, C.int(w.width), 15*h+2)
	C.sub_46AB20(other, C.int(other.width), 15*h+2)
	y := w.off_y + C.intptr_t(h) + 2
	var last *C.nox_window
	for id := 1120; id < 1134; id++ {
		last = g2ServerChild(root, id)
		if last != nil {
			last.off_y = y
		}
		y += C.intptr_t(h)
	}
	return last
}

//export sub_453B80
func sub_453B80(w *C.nox_window, d *C.nox_window_data) C.int {
	if w == nil || d == nil {
		return 0
	}
	return g2ServerDrawBackground(w, d, true)
}

//export nox_game2_server_copySpells_453F70
func nox_game2_server_copySpells_453F70(p unsafe.Pointer) {
	copy(unsafe.Slice((*byte)(g2ServerMem(1045488)), 20), unsafe.Slice((*byte)(p), 20))
}

//export sub_453F90
func sub_453F90() *C.char { return (*C.char)(g2ServerMem(1045488)) }

//export sub_453FA0
func sub_453FA0(p unsafe.Pointer, index, set C.int) C.int {
	word := uint8(int32(index) / 32)
	mask := uint32(1) << uint(uint32(index)&31)
	slot := (*uint32)(unsafe.Add(p, uintptr(word)*4))
	if set != 0 {
		*slot |= mask
		return C.int(mask)
	}
	*slot &^= mask
	return C.int(^mask)
}

//export nox_game2_server_spellBit_454000
func nox_game2_server_spellBit_454000(p unsafe.Pointer, index C.int) C.int {
	word := uint32(int32(index)/32) & 255
	mask := uint32(1) << uint(uint32(index)&31)
	return g2ServerBool(*(*uint32)(unsafe.Add(p, uintptr(word)*4))&mask != 0)
}

//export sub_454040
func sub_454040(p *C.uint32_t) {
	if p == nil {
		return
	}
	words := unsafe.Slice((*uint32)(unsafe.Pointer(p)), 5)
	for i := range words {
		words[i] = 0xFFFFFFFF
	}
	for i := 1; i < int(C.NOX_SPELLS_MAX); i++ {
		if bool(C.nox_xxx_spellIsValid_424B50(C.int(i))) && uint32(C.nox_xxx_spellFlags_424A70(C.int(i)))&0x7000000 != 0 && !bool(C.nox_xxx_spellIsEnabled_424B70(C.int(i))) {
			words[i/32] &^= uint32(1) << uint(i%32)
		}
	}
}

//export nox_game2_server_applySpells_4540E0
func nox_game2_server_applySpells_4540E0(p unsafe.Pointer) C.int {
	var result C.int
	for i := 1; i < 137; i++ {
		if nox_game2_server_spellBit_454000(p, C.int(i)) != 0 {
			result = g2ServerBool(bool(C.nox_xxx_spellEnable_424B90(C.int(i))))
		} else {
			result = g2ServerBool(bool(C.nox_xxx_spellDisable_424BB0(C.int(i))))
		}
	}
	return result
}

//export sub_454120
func sub_454120() C.int {
	w, root := C.dword_5d4594_1045480, C.dword_5d4594_1045484
	if w == nil || root == nil {
		return 0
	}
	d := g2ServerList(w)
	index := g2ServerFirstVisible(d)
	var result C.int
	for id := 1120; id <= 1133; id++ {
		box := g2ServerChild(root, id)
		if box == nil {
			index++
			continue
		}
		var spell C.int
		if d != nil && d.Items != nil && index >= 0 && index < int(d.Field_11_0) && g2ServerItems(d)[index].Text[0] != 0 {
			spell = C.nox_xxx_spellByTitle_424960((*C.wchar2_t)(unsafe.Pointer(&g2ServerItems(d)[index].Text[0])))
			g2ServerHidden(box, false)
		} else {
			g2ServerHidden(box, true)
		}
		if spell != 0 && nox_game2_server_spellBit_454000(g2ServerMem(1045488), spell) != 0 {
			box.draw_data.field_0 |= 4
		} else {
			box.draw_data.field_0 &^= 4
		}
		result = C.int(box.draw_data.field_0)
		index++
	}
	return result
}
