package opennox

import (
	"image"
	"image/color"
	"unsafe"

	"github.com/opennox/libs/client/keybind"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/client/input"
	"github.com/opennox/opennox/v1/client/noxrender"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/internal/mainmenu"
	"github.com/opennox/opennox/v1/legacy"
)

var (
	winMainMenu           *gui.Window
	winMainMenuAnimTop    *gui.Anim
	winMainMenuAnimBottom *gui.Anim
	mainMenuEyes          = newMainMenuEyes()
	winCreditsNative      *gui.Window
	creditsStartTicks     uint64
)

type mainMenuEye struct {
	spec          mainmenu.EyeSpec
	image         *noxrender.Image
	hidden        bool
	phaseTicks    uint32
	blinkTicks    uint32
	blinkCooldown uint32
}

func newMainMenuEyes() []mainMenuEye {
	specs := mainmenu.EyeSpecs()
	eyes := make([]mainMenuEye, len(specs))
	for i, spec := range specs {
		eyes[i] = mainMenuEye{
			spec:          spec,
			hidden:        spec.Hidden,
			phaseTicks:    spec.InitialPhaseTicks,
			blinkTicks:    spec.InitialBlinkTicks,
			blinkCooldown: spec.InitialBlinkCooldown,
		}
	}
	return eyes
}

func loadMainMenuEyes() {
	for i := range mainMenuEyes {
		mainMenuEyes[i].image = nox_xxx_gLoadImg(mainMenuEyes[i].spec.Name)
	}
}

func drawMainMenuBackground(win *gui.Window, draw *gui.WindowData) int {
	// The original callback reads the sequence counter before drawing.
	_ = noxClient.GetInputSeq()
	pos := win.GlobalPos()
	if !win.GetFlags().Has(gui.StatusImage) {
		if draw.BgColorVal != 0x80000000 {
			sz := win.Size()
			noxClient.r.DrawRectFilledOpaque(pos.X, pos.Y, sz.X, sz.Y, draw.BackgroundColor())
		}
	} else {
		pos = pos.Add(draw.ImagePoint())
		img := draw.BackgroundImage()
		if draw.Field0&2 != 0 {
			img = draw.HighlightImage()
		}
		if img != nil {
			noxClient.r.DrawImageAt(img, pos)
		}
	}

	if len(mainMenuEyes) == 0 || mainMenuEyes[0].image == nil {
		return 1
	}
	rand := noxClient.srv.Rand.Other.Int
	for i := range mainMenuEyes {
		eye := &mainMenuEyes[i]
		if eye.blinkCooldown != 0 {
			eye.blinkCooldown--
		}
		if eye.blinkTicks != 0 {
			eye.blinkTicks--
			eye.blinkCooldown = uint32(rand(60, 120))
		}
		oldPhase := eye.phaseTicks
		eye.phaseTicks--
		if oldPhase == 1 {
			if eye.hidden {
				eye.hidden = false
				eye.phaseTicks = uint32(rand(int(eye.spec.VisibleMin), int(eye.spec.VisibleMax)))
				eye.blinkCooldown = uint32(rand(60, 90))
			} else {
				eye.hidden = true
				eye.phaseTicks = uint32(rand(int(eye.spec.HiddenMin), int(eye.spec.HiddenMax)))
			}
		} else if !eye.hidden && eye.blinkCooldown == 0 && eye.blinkTicks == 0 && rand(0, 100) > 75 {
			eye.blinkTicks = uint32(rand(4, 8))
		}
	}
	for i := range mainMenuEyes {
		eye := &mainMenuEyes[i]
		if !eye.hidden && eye.blinkTicks == 0 && eye.image != nil {
			noxClient.r.DrawImageAt(eye.image, image.Pt(eye.spec.X, eye.spec.Y))
		}
	}
	return 1
}

func sub_4A2490(win *gui.Window, ev gui.WindowEvent) gui.WindowEventResp {
	switch ev.(type) {
	case gui.WindowFocus:
		return gui.RawEventResp(1)
	}
	return nil
}

func sub_4A1A60() bool {
	win := newWindowFromFile(noxClient.GUI, "OptsBack.wnd", guiOptsBackProc)
	guiOptsBack = win
	if win == nil {
		return false
	}
	win.SetFunc93(sub4A18E0)
	return true
}

func guiOptsBackProc(a1 *gui.Window, ev gui.WindowEvent) gui.WindowEventResp {
	switch ev := ev.(type) {
	case *WindowEvent0x4005:
		clientPlaySoundSpecial(sound.SoundShellSelect, 100)
		return gui.RawEventResp(1)
	case *WindowEvent0x4007:
		if gui.AnimGlobalState() != gui.AnimOut && gui.AnimGlobalState() != gui.AnimIn || sub4D6F30() {
			v3 := ev.Win.ID() - 151
			if v3 != 0 {
				if v3 == 1 {
					if noxClient.GameGetStateCode() == client.StateMainMenu {
						v6 := strMan.GetStringInFile("GUIQuit.c:ReallyQuitMessage", "C:\\NoxPost\\src\\client\\shell\\OptsBack.c")
						v4 := strMan.GetStringInFile("GUIQuit.c:ReallyQuitTitle", "C:\\NoxPost\\src\\client\\shell\\OptsBack.c")
						NewDialogWindow(guiOptsBack, v4, v6, 56, sub_4A19D0, nil)
					} else {
						if sub4D6F30() {
							sub_4D6F90(2)
						}
						if noxClient.GameGetStateCode() == client.StateOptions {
							closeOptionsNative()
							clientPlaySoundSpecial(sound.SoundShellClick, 100)
							return gui.RawEventResp(1)
						}
						if noxClient.GameGetStateCode() == client.StateColorSelect {
							legacy.Sub_4A7A60(1)
						}
						noxClient.nox_game_checkStateSwitch_43C1E0()
					}
				}
			} else {
				if noxClient.GameGetStateCode() == client.StateColorSelect {
					legacy.Sub_4A7A60(0)
				}
				noxClient.nox_game_checkStateOptions_43C220()
			}
			clientPlaySoundSpecial(sound.SoundShellClick, 100)
			return nil
		}
		return gui.RawEventResp(1)
	}
	return nil
}

func (c *Client) nox_xxx_wndLoadMainBG_4A2210() int {
	nox_client_gui_flag_815132 = 1
	gui.MainBg = newWindowFromFile(c.GUI, "MainBG.wnd", sub_4A2490)
	if !sub_4A1A60() {
		return 0
	}
	v1 := gui.MainBg.ChildByID(98)
	v1.SetFunc93(sub4A18E0)
	v1.SetDraw(drawMainMenuBackground)
	loadMainMenuEyes()
	gui.FocusMainBg()
	return 1
}

func winMainMenuAnimOutStartFnc() int {
	winMainMenuAnimTop.SetState(gui.AnimOut)
	winMainMenuAnimBottom.SetState(gui.AnimOut)
	gui.SetAnimGlobalState(gui.AnimOut)
	clientPlaySoundSpecial(sound.SoundShellSlideOut, 100)
	return 1
}

func sub_44E320() {
	c := noxClient
	legacy.Get_dword_5d4594_831236().Capture(false)
	c.GUI.Focus(nil)
	legacy.Get_dword_5d4594_831236().Hide()
	sub_450580()
	sub_43DDA0()
	legacy.Set_nox_gameDisableMapDraw_5d4594_2650672(0)
	if legacy.Get_dword_5d4594_831220() == 255 {
		if nox_client_gui_flag_815132 == 1 {
			legacy.Sub_4505E0()
			sub_4A2500()
			sub_578E00()
		}
	} else if memmap.Uint8(0x5D4594, 832472)&0x5 != 0 {
		legacy.Nox_client_lockScreenBriefing_450160(254, 1, 2)
		return
	}
	v0 := int(nox_client_getIntroScreenDuration_44E3B0())
	c.r.FadeOutScreen(v0, true, func() {
		sub_450580()
		legacy.Set_dword_5d4594_831260(0)
		sub_413A00(0)
	})
}

func sub_4A2500() {
	setEnableFrameLimit(true)
	gui.MainBg.Show()
	winMainMenu.Show()
	gui.FocusMainBg()
}

func sub_4A2530() {
	setEnableFrameLimit(false)
	gui.MainBg.Hide()
	winMainMenu.Hide()
}

func showCreditsNative() bool {
	if winCreditsNative != nil && !winCreditsNative.GetFlags().Has(gui.StatusDestroyed) {
		winCreditsNative.ShowModal()
		winCreditsNative.Capture(true)
		winCreditsNative.Focus()
		return true
	}

	root := nox_new_window_from_file("Briefing.wnd", creditsWindowProc)
	if root == nil {
		return false
	}
	textWin := root.ChildByID(1010)
	if textWin == nil {
		root.Destroy()
		return false
	}

	credits := strMan.GetStringInFile("Nox:Credits", "C:\\NoxPost\\src\\client\\Gui\\GUIBrief.c")
	if credits == "" {
		credits = "NOX\n\nCredits"
	}
	textWin.Func94(&gui.StaticTextSetText{Str: credits, Val: -1})
	const textWidth = 520
	textSize := noxClient.r.GetStringSizeWrapped(textWin.DrawData().Font(), credits, textWidth)
	if textSize.Y < noxClient.r.FontHeight(textWin.DrawData().Font()) {
		textSize.Y = noxClient.r.FontHeight(textWin.DrawData().Font())
	}
	textWin.SizeVal = image.Pt(textWidth, textSize.Y)
	textWin.SetPos(image.Pt((640-textWidth)/2, 480))
	textWin.SetDraw(drawCreditsNative)

	root.Flags &^= gui.StatusImage | gui.StatusNoFocus
	root.DrawData().SetBackgroundImage(nil)
	root.DrawData().SetBackgroundColor(color.Black)
	root.SetFunc93(creditsWindowProc)
	root.SetFunc94(creditsWindowProc)
	if root.ShowModal() != 0 || !root.Capture(true) {
		root.Destroy()
		return false
	}

	winCreditsNative = root
	creditsStartTicks = platformTicks()
	root.Focus()
	legacy.MusicModule.Sub_43DD70(24, 50)
	sub_4A2530()
	return true
}

func drawCreditsNative(win *gui.Window, draw *gui.WindowData) int {
	elapsed := platformTicks() - creditsStartTicks
	y := 480 - int(elapsed/30)
	win.SetPos(image.Pt((640-win.Size().X)/2, y))
	return gui.StaticTextDrawNoImage(win, draw)
}

func creditsWindowProc(_ *gui.Window, ev gui.WindowEvent) gui.WindowEventResp {
	switch ev := ev.(type) {
	case gui.WindowFocus:
		return gui.RawEventResp(1)
	case gui.WindowKeyPress:
		if ev.Pressed {
			closeCreditsNative()
		}
		return gui.RawEventResp(1)
	case *gui.WindowMouseState:
		if ev.State == input.NOX_MOUSE_LEFT_DOWN || ev.State == input.NOX_MOUSE_RIGHT_DOWN {
			closeCreditsNative()
		}
		return gui.RawEventResp(1)
	}
	return nil
}

func closeCreditsNative() {
	if winCreditsNative == nil {
		return
	}
	winCreditsNative.Capture(false)
	winCreditsNative.Destroy()
	winCreditsNative = nil
	creditsStartTicks = 0
	sub_43DDA0()
	sub_4A2500()
	gui.SetAnimGlobalState(gui.AnimInDone)
}

func winMainMenuAnimOutDoneFnc() int {
	ani := *winMainMenuAnimTop // copy
	winMainMenuAnimTop.Free()
	winMainMenuAnimTop = nil
	winMainMenuAnimBottom.Free()
	winMainMenuAnimBottom = nil
	winMainMenu.Destroy()
	winMainMenu = nil
	ani.Func13()
	return 1
}

func sub_4A24C0(a1 int) int {
	sub4A24C0(a1 != 0)
	return 1
}

func sub4A24C0(a1 bool) {
	v1 := gui.MainBg.ChildByID(99)
	if a1 {
		v1.Hide()
	} else {
		v1.Show()
	}
	if !a1 {
		legacy.Sub_43E8C0(1)
	}
}

func nox_game_showMainMenu_4A1C00() int {
	if nox_game_showMainMenu4A1C00() {
		return 1
	}
	return 0
}
func nox_game_showMainMenu4A1C00() bool {
	sub4D6F40(false)
	sub_4D6F90(0)
	noxClient.GameAddStateCode(client.StateMainMenu)
	menuWin := newWindowFromFile(noxClient.GUI, "MainMenu.wnd", nox_xxx_windowMainMenuProc_4A1DC0)
	if menuWin == nil {
		return false
	}
	winMainMenu = menuWin
	menuWin.SetFunc93(sub4A18E0)
	topMenu := menuWin.ChildByID(110)
	topMenu.SetFunc94(nox_xxx_windowMainMenuProc_4A1DC0)
	winMainMenuAnimTop = nox_gui_makeAnimation(topMenu, 0, 0, 0, -270, 0, 20, 0, -40)
	if winMainMenuAnimTop == nil {
		return false
	}
	winMainMenuAnimTop.StateID = client.StateMainMenu
	_ = winMainMenuAnimOutStartFnc
	winMainMenuAnimTop.Func12Ptr = legacy.Get_winMainMenuAnimOutStartFnc()
	_ = winMainMenuAnimOutDoneFnc
	winMainMenuAnimTop.FncDoneOutPtr = legacy.Get_winMainMenuAnimOutDoneFnc()
	bottomMenu := menuWin.ChildByID(120)
	bottomMenu.SetFunc94(nox_xxx_windowMainMenuProc_4A1DC0)
	winMainMenuAnimBottom = nox_gui_makeAnimation(bottomMenu, 0, 270, 0, 510, 0, -20, 0, 40)
	if winMainMenuAnimBottom == nil {
		return false
	}
	guiSetBackButtonText("OptsBack.wnd:Quit")
	nox_xxx_unknown_libname_11_4D1650()
	sub_578CD0()
	legacy.Sub_43D9B0(25, 100)
	if noxflags.HasGame(noxflags.GameFlag26) {
		mpBtn := menuWin.ChildByID(112)
		menuWin.Func94(&WindowEvent0x4007{Win: mpBtn})
	}
	return true
}

func sub_43BE40(a1 int) {
	gui.SetAnimGlobalState(gui.AnimState(a1))
}

func sub_43BE30() int {
	return int(gui.AnimGlobalState())
}

func sub_4A19D0() {
	nox_xxx_setContinueMenuOrHost_43DDD0(0)
	nox_client_gui_flag_815132 = 0
}

func sub_4A18E0(a1 *gui.Window, a2, a3, a4 int) int {
	res := sub4A18E0(a1, gui.AsWindowEvent(a2, uintptr(a3), uintptr(a4)))
	return gui.EventRespInt(res)
}

func sub4A18E0(a1 *gui.Window, ev gui.WindowEvent) gui.WindowEventResp {
	if sub_450560() {
		return gui.RawEventResp(1)
	}
	switch ev := ev.(type) {
	case gui.WindowKeyPress:
		if ev.Key != keybind.KeyEsc {
			return nil
		}
		if !ev.Pressed {
			return gui.RawEventResp(1)
		}
		if gui.AnimGlobalState() != gui.AnimInDone {
			if sub4D6F30() {
				sub_4D6F90(2)
			}
			return gui.RawEventResp(1)
		}
		if !sub44A4A0() {
			if noxClient.GameGetStateCode() == client.StateServerList {
				sub_4373A0()
			} else if noxClient.GameGetStateCode() == client.StateMainMenu {
				v6 := strMan.GetStringInFile("GUIQuit.c:ReallyQuitMessage", "C:\\NoxPost\\src\\client\\shell\\OptsBack.c")
				v5 := strMan.GetStringInFile("GUIQuit.c:ReallyQuitTitle", "C:\\NoxPost\\src\\client\\shell\\OptsBack.c")
				NewDialogWindow(guiOptsBack, v5, v6, 56, sub_4A19D0, nil)
			} else {
				if noxClient.GameGetStateCode() == client.StateColorSelect {
					legacy.Sub_4A7A60(1)
				}
				noxClient.nox_game_checkStateSwitch_43C1E0()
			}
			if sub4D6F30() {
				sub_4D6F90(2)
			}
			return gui.RawEventResp(1)
		}
	}
	return nil
}

func nox_xxx_unknown_libname_11_4D1650() {
	noxServer.Balance.Free()
	legacy.Nox_xxx_monsterListFree_5174F0()
}

func nox_xxx_windowMainMenuProc_4A1DC0(a1 *gui.Window, ev gui.WindowEvent) gui.WindowEventResp {
	switch ev := ev.(type) {
	case *WindowEvent0x4005:
		clientPlaySoundSpecial(sound.SoundShellSelect, 100)
		return gui.RawEventResp(1)
	case *WindowEvent0x4007:
		if winMainMenuAnimTop.State() != gui.AnimInDone && !noxflags.HasGame(noxflags.GameFlag26) {
			clientPlaySoundSpecial(sound.SoundShellClick, 100)
			return gui.RawEventResp(1)
		}
		switch ev.Win.ID() {
		case 111: // Solo campaign button
			noxServer.Announce = false
			if nox_xxx_checkHasSoloMaps() {
				noxflags.SetGame(noxflags.GameModeCoop)
				noxflags.UnsetGame(noxflags.GameOnline)
				noxflags.UnsetGame(noxflags.GameNotQuest)
				noxServer.ai.nox_xxx_gameSetAudioFadeoutMb(0)
				noxflags.UnsetEngine(noxflags.EngineAdmin | noxflags.EngineGodMode)
				sub4D6F40(false)
				sub_4D6F90(0)
				noxServer.nox_xxx_setQuest_4D6F60(0)
				legacy.Sub_4D6F80(0)
				legacy.Nox_xxx_cliShowHideTubes_470AA0(0)
				legacy.Sub_461440(0)
				winMainMenuAnimOutStartFnc()
				legacy.Nox_xxx_cliSetMinimapZoom_472520(1110)
				if nox_xxx_parseGamedataBinPre_4D1630() == 0 {
					nox_xxx_setContinueMenuOrHost_43DDD0(0)
					nox_client_gui_flag_815132 = 0
					return nil
				}
				if legacy.Nox_client_countSaveFiles_4DC550() != 0 {
					legacy.Sub_4A7A70(1)
					_ = nox_game_showSelChar_4A4DB0
					winMainMenuAnimTop.Func13Ptr = unsafe.Pointer(legacy.Get_nox_game_showSelChar_4A4DB0())
				} else {
					legacy.Sub_4A7A70(0)
					winMainMenuAnimTop.Func13Ptr = unsafe.Pointer(legacy.Get_nox_game_showSelClass_4A4840())
				}
				clientPlaySoundSpecial(sound.SoundShellClick, 100)
			} else {
				v9 := strMan.GetStringInFile("caution", "mainmenu.c")
				v5 := strMan.GetStringInFile("solo", "mainmenu.c")
				NewDialogWindow(winMainMenu, v5, v9, gui.DialogOKButton|gui.DialogFlag6, nil, nil)
				sub_44A360(1)
				sub_44A4B0()
				clientPlaySoundSpecial(sound.SoundShellClick, 100)
			}
			return gui.RawEventResp(1)
		case 112: // Multiplayer button
			noxServer.Announce = true
			// prepare to start a server
			winMainMenuAnimOutStartFnc()
			noxflags.SetEngine(noxflags.EngineAdmin)
			noxflags.UnsetEngine(noxflags.EngineGodMode)
			noxflags.SetGame(noxflags.GameOnline)
			noxflags.SetGame(noxflags.GameNotQuest)
			noxflags.UnsetGame(noxflags.GameModeCoop)
			legacy.Sub_461440(0)
			sub4D6F40(false)
			sub_4D6F90(0)
			noxServer.nox_xxx_setQuest_4D6F60(0)
			legacy.Sub_4D6F80(0)
			if sub_473670() == 0 {
				nox_client_toggleMap_473610()
			}
			legacy.Nox_xxx_cliShowHideTubes_470AA0(0)
			legacy.Nox_xxx_cliSetMinimapZoom_472520(2300)
			if nox_xxx_parseGamedataBinPre_4D1630() == 0 {
				nox_xxx_setContinueMenuOrHost_43DDD0(0)
				nox_client_gui_flag_815132 = 0
				return nil
			}
			winMainMenuAnimTop.Func13Ptr = legacy.Get_nox_game_showGameSel_4379F0()
			legacy.Sub_43AF50()
			clientPlaySoundSpecial(sound.SoundShellClick, 100)
		case 121:
			path, ok := nox_game_setMovieFile_4CB230("intro.vqa")
			if !ok {
				clientPlaySoundSpecial(sound.SoundShellClick, 100)
				break
			}
			winMainMenuAnimOutStartFnc()
			pushMovieFile(path)
			sub4B0640(func() {
				noxClient.GameStateSwitch()
			})
			winMainMenuAnimTop.Func13Ptr = legacy.Get_nox_client_drawGeneralCallback_4A2200()
			clientPlaySoundSpecial(sound.SoundShellClick, 100)
			return gui.RawEventResp(1)
		case 122:
			showCreditsNative()
			ev.Win.DrawData().Field0 &= 0xFFFFFFFD
			clientPlaySoundSpecial(sound.SoundShellClick, 100)
		case 131: // Solo Quest
			noxServer.Announce = false
			winMainMenuAnimOutStartFnc()
			noxflags.SetEngine(noxflags.EngineAdmin)
			noxflags.UnsetEngine(noxflags.EngineGodMode)
			noxflags.SetGame(noxflags.GameOnline)
			noxflags.SetGame(noxflags.GameNotQuest)
			noxflags.UnsetGame(noxflags.GameModeCoop)
			sub4D6F40(true)
			sub_4D6F90(1)
			noxServer.nox_game_setQuestStage_4E3CD0(0)
			legacy.Sub_4D7440(0)
			legacy.Nox_xxx_cliSetMinimapZoom_472520(2300)
			if sub_473670() == 0 {
				nox_client_toggleMap_473610()
			}
			legacy.Sub_461440(0)
			if nox_xxx_parseGamedataBinPre_4D1630() == 0 {
				nox_xxx_setContinueMenuOrHost_43DDD0(0)
				nox_client_gui_flag_815132 = 0
				return nil
			}
			winMainMenuAnimTop.Func13Ptr = legacy.Get_nox_game_showGameSel_4379F0()
			legacy.Sub_43AF50()
			clientPlaySoundSpecial(sound.SoundShellClick, 100)
		default:
			clientPlaySoundSpecial(sound.SoundShellClick, 100)
		}
		return gui.RawEventResp(1)
	default:
		return nil
	}
}
