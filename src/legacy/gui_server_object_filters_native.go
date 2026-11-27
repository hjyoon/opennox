package legacy

/*
#include "game2_server_bridge.h"
#include "client__gui__servopts__objlst.h"

*/
import "C"

import "unsafe"

const serverObjectFilterStrings = `C:\NoxPost\src\client\Gui\ServOpts\objlst.c`

//export nox_xxx_guiObjlistLoad_4530C0
func nox_xxx_guiObjlistLoad_4530C0(parent *C.nox_window, classMask C.int) *C.nox_window {
	root := nox_new_window_from_file(internCStr("objlst.wnd"), unsafe.Pointer(C.sub_4533D0))
	serverSharedWindows.object_filter = root
	if root == nil {
		return nil
	}
	asWindow(root).AddCleanup(func() {
		if serverSharedWindows.object_filter == root {
			serverSharedWindows.object_filter, serverSharedWindows.object_list = nil, nil
		}
	})
	nox_xxx_wndSetDrawFn_46B340(root, unsafe.Pointer(C.sub_453350))
	sub_46B120(root, parent)
	nox_xxx_wnd_46B280(root, parent)
	list := g2ServerChild(root, 1510)
	serverSharedWindows.object_list = list
	if list == nil {
		nox_xxx_windowDestroyMB_46C4E0(root)
		serverSharedWindows.object_filter, serverSharedWindows.object_list = nil, nil
		return nil
	}
	sub_4532E0()
	g2ServerEvent(list, 16399, 0, 0)
	var title [66]wchar2_t
	var mask C.int
	var rows int
	var itemName func(C.int) *wchar2_t
	switch classMask {
	case 0x1000000:
		legacyWords.dword_5d4594_1045460 = 0
		copyWideString(&title[0], g2ServerLoad("Weapons", serverObjectFilterStrings, 321))
		mask, rows, itemName = 4, 25, sub_4159F0
	case 0x2000000:
		legacyWords.dword_5d4594_1045460 = 1
		copyWideString(&title[0], g2ServerLoad("servopts.wnd:Armor", serverObjectFilterStrings, 308))
		mask, rows, itemName = 1, 26, sub_415E80
	}
	var count uint32
	for range rows {
		if name := itemName(mask); name != nil {
			g2ServerText(list, 16397, unsafe.Pointer(name), ^uintptr(0))
			count++
		}
		mask *= 2
	}
	g2ServerText(list, 16385, unsafe.Pointer(&title[0]), 0)
	g2ServerText(list, 16408, unsafe.Pointer(g2ServerChild(root, 1513)), 0)
	g2ServerText(list, 16409, unsafe.Pointer(g2ServerChild(root, 1514)), 0)
	*g2ServerWord(1045472 + 4*uintptr(legacyWords.dword_5d4594_1045460)) = count
	sub_453750()
	if !g2ServerGame(1) || g2ServerGame(49152) {
		sub_46AD20(root, 1515, 1533, 0)
	}
	return serverSharedWindows.object_filter
}
