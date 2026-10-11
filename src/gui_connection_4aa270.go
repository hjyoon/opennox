package opennox

import (
	"image"

	"github.com/opennox/opennox/v1/client/gui"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

const connectionState4AA270 = gui.StateID(400)

var winConnection4AA270 *connectionScreen4AA270

// Original 59C56C was extracted from the blob; keep its primary mode in
// Go-owned storage, separately from the legacy WOL predicate at 85B7B0.
var connectionMode43AF50 uint32

// Go owns callbacks and UI services; only native C function pointers are
// stored in the C-allocated animation records.
type connectionServices4AA270 struct {
	closeDialog    func()
	pushState      func(gui.StateID)
	enableOptions  func(int)
	setMode        func(int)
	newWindow      func() *gui.Window
	setBackLabel   func()
	back           func()
	sharedProc     gui.WindowFunc
	flags          func() noxflags.GameFlag
	solo           func() bool
	request        func() uint32
	playSound      func(sound.ID, int)
	wolUnavailable func(*gui.Window)
}

type connectionScreen4AA270 struct {
	root        *gui.Window
	top, bottom *gui.Anim
	services    connectionServices4AA270
}

func init() {
	legacy.NoxGameShowConnection4AA270 = noxGameShowConnection4AA270
	legacy.ConnectionOutStart4AA450 = connectionOutStart4AA450
	legacy.ConnectionOutDone4AA490 = connectionOutDone4AA490
	gui.RegisterState(connectionState4AA270, "Connection", func() bool {
		return noxGameShowConnection4AA270() != 0
	})
}

func connectionSetMode43AF50(mode int) {
	// GAME.EXE 43AF50 writes the primary mode at 59C56C, then stores the
	// separate WOL predicate (mode == 1) at 85B7B0. Mode 2 is not WOL.
	connectionMode43AF50 = uint32(mode)
	legacy.Set_dword_5d4594_2650652(bool2int(mode == 1))
}

func noxGameShowConnection4AA270() int {
	return showConnection4AA270(connectionServices4AA270{
		closeDialog:   sub_44A400,
		pushState:     func(id gui.StateID) { noxClient.GameAddStateCode(id) },
		enableOptions: sub_4A1A40,
		setMode:       connectionSetMode43AF50,
		newWindow:     func() *gui.Window { return nox_new_window_from_file("ArnaMain.wnd", nil) },
		setBackLabel:  func() { guiSetBackButtonText("OptsBack.wnd:Back") },
		back: func() {
			guiOptsBack.Func94(&WindowEvent0x4007{Win: guiOptsBack.ChildByID(152)})
		},
		sharedProc: sub4A18E0,
		flags:      noxflags.GetGame,
		solo:       sub4D6F30,
		request:    func() uint32 { return memmap.Uint32(0x5D4594, 1556104) },
		playSound:  clientPlaySoundSpecial,
		wolUnavailable: func(root *gui.Window) {
			// The legacy WOL API/login service is not implemented by this
			// cross-platform client. Preserve 41D440 == 0's error branch;
			// do not silently route a WOL request to LAN/Solo Quest.
			const file = "C:\\NoxPost\\src\\client\\shell\\ArnaMain.c"
			NewDialogWindow(root, strMan.GetStringInFile("Error", file),
				strMan.GetStringInFile("WOLAPIINIT_Failed", file), gui.DialogOKButton|gui.DialogFlag6, nil, nil)
		},
	})
}

func showConnection4AA270(services connectionServices4AA270) int {
	services.closeDialog()
	services.pushState(connectionState4AA270)
	services.enableOptions(1)
	services.setMode(2)
	root := services.newWindow()
	if root == nil {
		return 0
	}
	top, bottom := root.ChildByID(410), root.ChildByID(420)
	if top == nil || bottom == nil || root.ChildByID(411) == nil || root.ChildByID(421) == nil {
		root.Destroy()
		return 0
	}
	st := &connectionScreen4AA270{root: root, services: services}
	root.SetFunc94(st.proc4AA4D0)
	root.SetFunc93(services.sharedProc)
	top.SetFunc94(st.proc4AA4D0)
	st.top = gui.NewAnim(top, image.Pt(0, 0), image.Pt(0, -240), image.Pt(0, 20), image.Pt(0, -40))
	st.top.StateID = connectionState4AA270
	st.top.Func12Ptr = legacy.GetConnectionOutStart4AA450()
	st.top.FncDoneOutPtr = legacy.GetConnectionOutDone4AA490()
	bottom.SetFunc94(st.proc4AA4D0)
	st.bottom = gui.NewAnim(bottom, image.Pt(0, 240), image.Pt(0, 480), image.Pt(0, -20), image.Pt(0, 40))
	winConnection4AA270 = st
	services.setBackLabel()
	if root.ShowModal() < 0 {
		st.free()
		winConnection4AA270 = nil
		return 0
	}
	// GAME.EXE dispatches these clicks before the slide-in has completed.
	// Flag 26 has priority over both Quest request words.
	if services.flags().Has(noxflags.GameFlag26) {
		root.Func94(&WindowEvent0x4007{Win: root.ChildByID(411)})
	} else if services.request() == 1 {
		root.Func94(&WindowEvent0x4007{Win: root.ChildByID(421)})
	} else if services.request() == 2 {
		services.back()
	}
	return 1
}

func connectionOutStart4AA450() int {
	st := winConnection4AA270
	if st == nil || st.top == nil || st.bottom == nil {
		return 0
	}
	st.top.SetState(gui.AnimOut)
	st.bottom.SetState(gui.AnimOut)
	gui.SetAnimGlobalState(gui.AnimOut)
	st.services.playSound(sound.SoundShellSlideOut, 100)
	return 1
}

func connectionOutDone4AA490() int {
	st := winConnection4AA270
	if st == nil || st.top == nil {
		return 0
	}
	next := st.top.Func13Ptr
	winConnection4AA270 = nil
	st.free()
	if next != nil {
		ccall.CallIntVoid(next)
	}
	return 1
}

func (st *connectionScreen4AA270) free() {
	st.top.Free()
	st.bottom.Free()
	st.top, st.bottom = nil, nil
	st.root.Destroy()
	st.root = nil
}

func (st *connectionScreen4AA270) proc4AA4D0(_ *gui.Window, ev gui.WindowEvent) gui.WindowEventResp {
	switch ev := ev.(type) {
	case *WindowEvent0x4005:
		st.services.playSound(sound.SoundShellSelect, 100)
		return gui.RawEventResp(1)
	case *WindowEvent0x4007:
		if st.top == nil || ev.Win == nil {
			return nil
		}
		if st.top.State() == gui.AnimInDone || st.services.flags().Has(noxflags.GameFlag26) || st.services.solo() {
			switch ev.Win.ID() {
			case 421:
				connectionOutStart4AA450()
				st.top.Func13Ptr = legacy.Get_nox_game_showGameSel_4379F0()
				st.services.setMode(0)
				noxflags.UnsetGame(noxflags.GameFlag25)
			case 411:
				st.services.wolUnavailable(st.root)
				return nil
			}
		}
		st.services.playSound(sound.SoundShellClick, 100)
		return gui.RawEventResp(1)
	default:
		return nil
	}
}
