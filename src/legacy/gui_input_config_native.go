package legacy

/*
#include "GAME2_1.h"
#include "GAME3_1.h"
#include "client__gui__guiinput.h"
#include "client__shell__inputcfg__inputcfg.h"
#include "noxstring.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

//export sub_4C3500
func sub_4C3500() C.int {
	root := nox_new_window_from_file(internCStr("yesno.wnd"), nil)
	if root != nil {
		nox_window_setPos_46A9B0(root, int32(clientDisplayStorage.width-320)/2, int32(clientDisplayStorage.height-240)/2)
	}
	nox_window_set_hidden(root, 1)
	return C.int(bool2int(root != nil))
}

//export sub_4C35B0
func sub_4C35B0(cancel C.int) C.int {
	if wndIsShown_nox_xxx_wndIsShown_46ACC0(gameInputBindings.root) != 0 {
		return 0
	}
	sub_413A00(0)
	if cancel != 0 {
		g2ServerHidden(gameInputBindings.root, true)
	} else {
		sub_4C3620()
		nox_common_writecfgfile(internCStr("nox.cfg"))
		g2ServerHidden(gameInputBindings.root, true)
		C.sub_472280()
		sub_4ADA40()
	}
	return 1
}

//export sub_4C3A60
func sub_4C3A60(win *nox_window, event C.int, a, b C.uintptr_t) C.int {
	if event == 19 || event == 20 {
		return 0
	}
	return nox_xxx_wndListboxProcWithoutData10_4A28E0(win, event, a, b)
}

type inputBindingPanel struct {
	root         *nox_window
	actions      *nox_window
	labels       *nox_window
	prompt       *nox_window
	selected     *nox_window
	columns      [2]*nox_window
	promptOffset uintptr
	promptUnits  int
	keyEmpty     uintptr
	source       string
	escEventLine int32
	pressKeyLine int32
	restoreFocus bool
}

var gameInputBindings = inputBindingPanel{
	promptOffset: 1321256,
	promptUnits:  (1321520 - 1321256) / 2,
	keyEmpty:     185444,
	source:       `C:\NoxPost\src\client\Gui\GuiInput.c`,
	escEventLine: 191,
	pressKeyLine: 436,
}

var shellInputBindings = inputBindingPanel{
	promptOffset: 1522636,
	promptUnits:  (1522892 - 1522636) / 2,
	keyEmpty:     187824,
	source:       `C:\NoxPost\src\Client\shell\InputCfg\inputcfg.c`,
	escEventLine: 192,
	pressKeyLine: 424,
	restoreFocus: true,
}

func (p *inputBindingPanel) loadString(id string, line int32) *wchar2_t {
	return nox_strman_loadString_40F1D0(internCStr(id), nil, internCStr(p.source), line)
}

func inputBindingEventName(title *wchar2_t) string {
	event := Nox_xxx_bindevent_bindNameByTitle_42EA40(GoWString(title))
	if event == nil {
		return "(null)"
	}
	return event.Name
}

func (p *inputBindingPanel) apply() C.int {
	list := g2ServerList(p.labels)
	sub_42CD90()
	if list != nil {
		for i := 0; i < int(int16(list.Field_11_0)); i++ {
			event := inputBindingEventName(listBoxRowText(p.labels, uintptr(i)))
			primary := Nox_xxx_keybind_nameByTitle_42E960(GoWString(listBoxRowText(p.columns[0], uintptr(i))))
			secondary := Nox_xxx_keybind_nameByTitle_42E960(GoWString(listBoxRowText(p.columns[1], uintptr(i))))
			if secondary != 0 {
				Nox_client_parseConfigHotkeysLine_42CF50(secondary.String() + " = " + event)
			}
			if primary != 0 {
				Nox_client_parseConfigHotkeysLine_42CF50(primary.String() + " = " + event)
			}
		}
	}
	event := inputBindingEventName(p.loadString("bindevent:ToggleQuitMenu", p.escEventLine))
	escape := Nox_xxx_keybind_nameByTitle_42E960(GoWString(p.loadString("keybind:Esc", p.escEventLine+1)))
	if escape == 0 {
		return 0
	}
	return C.int(Nox_client_parseConfigHotkeysLine_42CF50(escape.String() + " = " + event))
}

//export sub_4C3620
func sub_4C3620() C.int { return gameInputBindings.apply() }

//export sub_4CBD30
func sub_4CBD30() C.int { return shellInputBindings.apply() }

func (p *inputBindingPanel) listColumns() [3]*nox_window {
	return [3]*nox_window{p.labels, p.columns[0], p.columns[1]}
}

func (p *inputBindingPanel) attachLists(actionsProc, keyProc unsafe.Pointer) bool {
	root := p.root
	actions := g2ServerChild(root, 910)
	p.actions = actions
	p.labels = g2ServerChild(root, 911)
	p.columns[0] = g2ServerChild(root, 912)
	p.columns[1] = g2ServerChild(root, 913)
	if actions == nil || p.labels == nil || p.columns[0] == nil || p.columns[1] == nil {
		return false
	}
	list := g2ServerList(actions)
	if list == nil || list.Field_7 == nil || list.Field_8 == nil || list.Field_9 == nil {
		return false
	}
	nox_xxx_wndSetID_46B080((*nox_window)(list.Field_7), 921)
	nox_xxx_wndSetID_46B080((*nox_window)(list.Field_8), 922)
	nox_xxx_wndSetID_46B080((*nox_window)(list.Field_9), 920)
	nox_xxx_wndSetProc_46B2C0(actions, actionsProc)
	for _, column := range p.listColumns() {
		sub_46B120(column, actions)
	}
	nox_xxx_wndSetWindowProc_46B300(p.columns[0], keyProc)
	nox_xxx_wndSetWindowProc_46B300(p.columns[1], keyProc)
	up, down := g2ServerChild(root, 921), g2ServerChild(root, 922)
	for _, column := range p.listColumns() {
		g2ServerText(column, 16408, unsafe.Pointer(up), 0)
		g2ServerText(column, 16409, unsafe.Pointer(down), 0)
	}
	return true
}

func (p *inputBindingPanel) setMouseButtons() {
	root := p.root
	for id, end := 971, 971+int(InputMouseButtonCount47DBC0()); id < end; id++ {
		g2ServerEnabled(g2ServerChild(root, id), true)
	}
	if button := g2ServerChild(root, 971+int(nox_client_mousePriKey_430AF0())); button != nil {
		g2ServerEvent(button, 16392, 1, 0)
	}
}

func (p *inputBindingPanel) bindPromptLabel() {
	if label := g2ServerChild(p.prompt, 981); label != nil {
		g2ServerText(label, 16385, g2ServerMem(p.promptOffset), 0)
	}
}

//export sub_4C3760
func sub_4C3760() C.int {
	p := &gameInputBindings
	root := nox_new_window_from_file(internCStr("InputCfg.wnd"), C.sub_4C3A90)
	p.root = root
	if root == nil || !p.attachLists(C.sub_4C3CD0, C.sub_4C3A60) {
		return 0
	}

	p.setMouseButtons()
	nox_window_setPos_46A9B0(root, (int32(clientDisplayStorage.width)-int32(root.width))/2, 0)
	prompt := g2ServerChild(root, 980)
	p.prompt = prompt
	if prompt == nil {
		return 0
	}
	sub_46B120(prompt, nil)
	nox_xxx_wndSetProc_46B2C0(prompt, C.sub_4C3A90)
	nox_xxx_wndSetWindowProc_46B300(prompt, C.sub_4C3EB0)
	g2ServerHidden(prompt, true)
	nox_window_setPos_46A9B0(prompt, (int32(clientDisplayStorage.width)-int32(prompt.width))/2, int32(prompt.off_y))
	p.bindPromptLabel()
	g2ServerEnabled(g2ServerChild(root, 932), true)
	g2ServerHidden(root, true)
	return 1
}

//export sub_4C3CD0
func sub_4C3CD0(win *nox_window, event C.int, a, b C.uintptr_t) C.int {
	return gameInputBindings.handleList(win, event, a, b)
}

//export sub_4CBF60
func sub_4CBF60(win *nox_window, event C.int, a, b C.uintptr_t) C.int {
	return shellInputBindings.handleList(win, event, a, b)
}

func (p *inputBindingPanel) handleList(win *nox_window, event C.int, a, b C.uintptr_t) C.int {
	switch event {
	case 23:
		return 1
	case 16391, 16384:
		root := p.root
		up := C.uintptr_t(uintptr(unsafe.Pointer(g2ServerChild(root, 921))))
		down := C.uintptr_t(uintptr(unsafe.Pointer(g2ServerChild(root, 922))))
		if a == up || a == down {
			for _, column := range p.listColumns() {
				g2ServerEvent(column, int(event), uintptr(a), 0)
			}
		}
	case 16393:
		// The actions list updates its scroll state before all columns are synchronized.
		nox_xxx_wndListboxProcPre_4A30D0(win, C.uint(event), a, b)
		first := uintptr(nox_xxx_wndListBoxFirstVisible(win))
		for _, column := range p.listColumns() {
			g2ServerEvent(column, 16412, first, 0)
		}
	case 16400:
		sender := (*nox_window)(unsafe.Pointer(uintptr(a)))
		index := int32(g2ServerEvent(sender, 16404, 0, 0))
		if index >= 0 {
			p.selected = sender
			p.showCapture(uintptr(index))
		}
	}
	return C.int(nox_xxx_wndListboxProcPre_4A30D0(win, C.uint(event), a, b))
}

func inputPromptPart(s *wchar2_t) *uint16 {
	if s == nil {
		s = internWStr("(null)")
	}
	return (*uint16)(unsafe.Pointer(s))
}

func (p *inputBindingPanel) showCapture(index uintptr) {
	item := listBoxRowText(p.labels, index)
	prompt := p.loadString("InputCfg.wnd:PressKey", p.pressKeyLine)
	text := unsafe.Slice((*uint16)(g2ServerMem(p.promptOffset)), p.promptUnits)
	body := text[:len(text)-1]
	n := alloc.StrCopy16P(body, inputPromptPart(prompt))
	n += copy(body[n:], []uint16{'\n', '\''})
	n += alloc.StrCopy16P(body[n:], inputPromptPart(item))
	n += copy(body[n:], []uint16{'\''})
	text[n] = 0
	nox_xxx_wndShowModalMB_46A8C0(p.prompt)
	nox_xxx_windowFocus_46B500(p.prompt)
	sub_46C690(p.prompt)
}

func (p *inputBindingPanel) hideCapture() {
	nox_xxx_windowFocus_46B500(nil)
	nox_xxx_wnd_46C6E0(p.prompt)
	g2ServerHidden(p.prompt, true)
	if p.restoreFocus {
		nox_gui_input_cfg_restore_focus()
	}
}

//export sub_4C3EB0
func sub_4C3EB0(win *nox_window, event C.int, a, b C.uintptr_t) C.int {
	return gameInputBindings.handleCapture(event, a, b)
}

func (p *inputBindingPanel) handleCapture(event C.int, a, b C.uintptr_t) C.int {
	var button C.uint
	switch event {
	case 6, 7:
		button = 0x10000
	case 10, 11:
		button = 0x10002
	case 14, 15:
		button = 0x10001
	case 19:
		button = 0x10003
	case 20:
		button = 0x10004
	case 21:
		if a == 1 {
			if b == 2 {
				p.hideCapture()
				if selection := p.selected; selection != nil {
					g2ServerEvent(selection, 16403, ^uintptr(0), 0)
				}
			}
			return 1
		}
		if b == 1 && p.setKey(C.uint(a)) != 0 {
			button = 0x10000
		} else {
			return 0
		}
	default:
		return 0
	}
	p.setButton(button)
	p.hideCapture()
	return 1
}

func clearInputBinding(list *nox_window, name *wchar2_t, empty unsafe.Pointer) {
	data := g2ServerList(list)
	if data == nil {
		return
	}
	count := int(data.Field_11_0)
	for i := 0; i < count; i++ {
		current := listBoxRowText(list, uintptr(i))
		if current != nil && C.nox_wcscmp(current, name) == 0 {
			g2ServerText(list, 16407, empty, uintptr(i))
		}
	}
}

func (p *inputBindingPanel) set(name *wchar2_t, offset uintptr) {
	for i, column := range p.columns {
		clearInputBinding(column, name, memmap.PtrOff(0x587000, offset+4*uintptr(i)))
	}
	index := int32(g2ServerEvent(p.selected, 16404, 0, 0))
	g2ServerText(p.selected, 16407, unsafe.Pointer(name), uintptr(index))
	g2ServerEvent(p.selected, 16403, ^uintptr(0), 0)
	p.selected = nil
}

func (p *inputBindingPanel) setKey(key C.uint) C.int {
	name := nox_xxx_keybind_titleByKeyZero_42EA00(key)
	if name == nil {
		return 0
	}
	if p.selected != nil {
		p.set(name, p.keyEmpty)
	}
	return 1
}

func (p *inputBindingPanel) setButton(key C.uint) C.int {
	name := nox_xxx_keybind_titleByKey_42EA00(key)
	if p.selected == nil {
		return 1
	}
	if name == nil {
		return 0
	}
	p.set(name, p.keyEmpty+8)
	return 1
}

//export sub_4C3FC0
func sub_4C3FC0(key C.uint) C.int { return gameInputBindings.setKey(key) }

//export sub_4C4100
func sub_4C4100(key C.uint) C.int { return gameInputBindings.setButton(key) }

func (p *inputBindingPanel) destroy() {
	nox_xxx_windowDestroyMB_46C4E0(p.root)
	nox_xxx_windowDestroyMB_46C4E0(p.prompt)
	p.root, p.prompt, p.selected = nil, nil, nil
	p.actions, p.labels = nil, nil
	clear(p.columns[:])
}

//export sub_4C4220
func sub_4C4220() C.int {
	gameInputBindings.destroy()
	return 0
}

//export sub_4C4260
func sub_4C4260() {
	C.sub_4C3B70()
	nox_xxx_wndShowModalMB_46A8C0(gameInputBindings.root)
	sub_413A00(1)
}

//export sub_4C4280
func sub_4C4280() C.int {
	return C.int(bool2int(wndIsShown_nox_xxx_wndIsShown_46ACC0(gameInputBindings.root) == 0))
}
