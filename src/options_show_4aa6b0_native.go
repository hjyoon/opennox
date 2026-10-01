package opennox

import (
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/timer"
)

func optionsShowNativeHooks4AA6B0() optionsShowHooks4AA6B0[*gui.Window, *gui.Anim, noxrender.ImageHandle] {
	return optionsShowHooks4AA6B0[*gui.Window, *gui.Anim, noxrender.ImageHandle]{
		addState:  func(id int32) { noxClient.GameAddStateCode(gui.StateID(id)) },
		newWindow: legacy.OptionsNewWindow4AA6B0,
		storeRoot: legacy.OptionsStoreRoot4AA6B0,
		advanced:  legacy.OptionsNewAdvanced4AA6B0,
		root:      legacy.Get_dword_5d4594_1309720,
		setProc:   func(win *gui.Window) { win.SetFunc93C(legacy.Get_sub_4A18E0()) },
		tabWidth:  func(width int32) { noxClient.r.SetTabWidth(int(width)) },
		newAnimation: func(win *gui.Window, args [8]int32) *gui.Anim {
			return nox_gui_makeAnimation(win, int(args[0]), int(args[1]), int(args[2]), int(args[3]), int(args[4]), int(args[5]), int(args[6]), int(args[7]))
		},
		storeAnimation: legacy.OptionsStoreAnimation4AA6B0,
		animation:      legacy.Get_nox_wnd_xxx_1309740,
		stateID:        func(anim *gui.Anim, id int32) { anim.StateID = gui.StateID(id) },
		startOut:       func(anim *gui.Anim) { anim.Func12Ptr = legacy.OptionsStartOut4AA6B0() },
		doneOut:        func(anim *gui.Anim) { anim.FncDoneOutPtr = legacy.OptionsDoneOut4AA6B0() },
		child:          func(win *gui.Window, id int32) *gui.Window { return win.ChildByID(uint(id)) },
		thumb:          func(slider *gui.Window) *gui.Window { return slider.Field100Ptr },
		width:          func(win *gui.Window, width int32) { win.SizeVal.X = int(width) },
		height:         func(win *gui.Window, height int32) { win.SizeVal.Y = int(height) },
		loadImage:      func(name string) noxrender.ImageHandle { return nox_xxx_gLoadImg(name).C() },
		images: func(slider *gui.Window, enabled, selected, highlight noxrender.ImageHandle) {
			gui.ButtonSetImage(slider, nil, nil, enabled, selected, highlight)
		},
		event: func(win *gui.Window, event int32, first, second uint32) {
			win.Func94(gui.AsWindowEvent(int(event), uintptr(first), uintptr(second)))
		},
		current: func(channel int32) uint32 {
			switch channel {
			case 0:
				return (*timer.Timer)(legacy.Get_dword_587000_127004()).Current
			case 1:
				return (*timer.Timer)(legacy.Get_dword_587000_122852()).Current
			default:
				return (*timer.Timer)(legacy.Get_dword_587000_93164()).Current
			}
		},
		storeCheckbox: legacy.OptionsStoreCheckbox4AA6B0,
		enabled: func(channel int32) int32 {
			switch channel {
			case 0:
				return int32(legacy.Sub_453070())
			case 1:
				return int32(legacy.Sub_44D990())
			default:
				return int32(legacy.Sub_43DC30())
			}
		},
		checkbox:    legacy.OptionsCheckbox4AA6B0,
		flags:       func(win *gui.Window) uint32 { return win.DrawData().Field0 },
		storeFlags:  func(win *gui.Window, flags uint32) { win.DrawData().Field0 = flags },
		returnNull:  legacy.OptionsReturnNull4AA6B0,
		backText:    func(name string) { guiSetBackButtonText(strman.ID(name)) },
		backEnabled: func(enabled int32) { sub_4A1A40(int(enabled)) },
		video:       legacy.OptionsInitVideo4AA6B0,
	}
}
