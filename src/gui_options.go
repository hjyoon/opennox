package opennox

import (
	"fmt"
	"image"
	"math"
	"strings"
	"unsafe"

	"github.com/spf13/viper"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/noxfont"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
)

const (
	guiIDMenuExt = 380

	configVideoWidth  = "video.size.width"
	configVideoHeight = "video.size.height"
	configVideoGamma  = "video.gamma"
)

func init() {
	viper.SetDefault(configVideoWidth, 1024)
	viper.SetDefault(configVideoHeight, 768)
	viper.SetDefault(configVideoGamma, gammaDef)
	registerOnConfigRead(func() {
		setGamma(float32(viper.GetFloat64(configVideoGamma)))
	})
}

var (
	guiOptionsRes = image.Point{
		X: noxDefaultWidth,
		Y: noxDefaultHeight,
	}
	guiOptsBack      *gui.Window
	winOptionsNative *gui.Window
)

func noxGameShowOptionsNative() int {
	if winOptionsNative != nil && !winOptionsNative.GetFlags().Has(gui.StatusDestroyed) {
		winOptionsNative.ShowModal()
		gui.SetAnimGlobalState(gui.AnimInDone)
		return 1
	}

	root := nox_new_window_from_file("Options.wnd", optionsWindowProcNative)
	if root == nil {
		return 0
	}
	noxClient.NewGUIAdvOptsOn(root)
	root.SetFunc93(optionsWindowProcNative)
	if root.ShowModal() != 0 {
		root.Destroy()
		return 0
	}

	winOptionsNative = root
	noxClient.GameAddStateCode(client.StateOptions)
	guiSetBackButtonText("OptsBack.wnd:Back")
	sub_4A1A40(0)
	nox_video_setMenuOptions(root)
	fullScreenID := uint(331)
	if nox_video_getFullScreen() != 0 {
		fullScreenID = 332
	}
	if w := root.ChildByID(fullScreenID); w != nil {
		w.Func94(gui.AsWindowEvent(0x4008, 1, 0))
	}
	initNativeVolumeSliders(root)
	gui.SetAnimGlobalState(gui.AnimInDone)
	return 1
}

func initNativeVolumeSliders(root *gui.Window) {
	bg := nox_xxx_gLoadImg("OptionsVolumeSlider")
	lit := nox_xxx_gLoadImg("OptionsVolumeSliderLit")
	for id := uint(351); id <= 353; id++ {
		slider := root.ChildByID(id)
		if slider == nil {
			continue
		}
		thumb := slider.Field100()
		if thumb == nil {
			continue
		}
		thumb.SizeVal = image.Pt(24, 20)
		thumb.SetEnd(thumb.Offs().Add(thumb.Size()))
		thumb.DrawData().SetBackgroundImage(bg)
		thumb.DrawData().SetHighlightImage(lit)
		thumb.DrawData().SetSelectedImage(lit)
	}
}

func optionsWindowProcNative(_ *gui.Window, ev gui.WindowEvent) gui.WindowEventResp {
	switch ev := ev.(type) {
	case *WindowEvent0x4005:
		clientPlaySoundSpecial(sound.SoundShellSelect, 100)
		return gui.RawEventResp(1)
	case *WindowEvent0x4007:
		if ev.Win == nil {
			return nil
		}
		id := ev.Win.ID()
		switch {
		case id >= guiIDMenuExt && id < uint(guiIDMenuExt+len(getResolutionOptions())):
			nox_gui_menu_proc_ext(int(id))
		case id == 331:
			nox_video_setFullScreen(0)
		case id == 332:
			nox_video_setFullScreen(1)
		}
		clientPlaySoundSpecial(sound.SoundShellClick, 100)
		return gui.RawEventResp(1)
	case *WindowEvent0x4009:
		if ev.Win == nil {
			return nil
		}
		switch ev.Win.ID() {
		case 316:
			setGammaSlider(ev.Val)
		case 318:
			noxClient.SetSensitivity(float32(math.Pow(10, float64(ev.Val)/50.0-1.0)))
		}
		return gui.RawEventResp(1)
	case gui.WindowKeyPress:
		if ev.Key == keybind.KeyEsc && ev.Pressed {
			closeOptionsNative()
			return gui.RawEventResp(1)
		}
	}
	return nil
}

func closeOptionsNative() {
	if winOptionsNative != nil {
		winOptionsNative.Destroy()
		winOptionsNative = nil
	}
	if noxClient.GameGetStateCode() == client.StateOptions {
		noxClient.GamePopState()
	}
	sub_4A1A40(1)
	guiSetBackButtonText("OptsBack.wnd:Quit")
	gui.SetAnimGlobalState(gui.AnimInDone)
	nox_game_showMainMenu_4A1C00()
}

func getResolutionOptions() []image.Point {
	if noxHighRes {
		return []image.Point{
			{640, 480},
			{800, 600},
			{1024, 768},
			{1280, 720},
			{1600, 900},
			{1920, 1080},
			{2144, 1206},
			{2560, 1440},
			{3200, 1800},
			{3840, 2160},
		}
	}
	return []image.Point{
		{640, 480},
		{720, 540},
		{800, 600},
		{928, 696},
		{1024, 768},

		{},
		{},
		{864, 486},
		{928, 522},
		{1024, 576},
	}
}

func nox_video_setMenuOptions(root *gui.Window) {
	mode := noxClient.videoGetGameMode()
	for i, res := range getResolutionOptions() {
		if res == (image.Point{}) {
			continue
		}
		id := uint(guiIDMenuExt + i)
		if res == mode {
			root.ChildByID(id).Func94(gui.AsWindowEvent(0x4008, 1, 0))
			return
		}
	}
}

func nox_gui_menu_proc_ext(id int) int {
	opts := getResolutionOptions()
	if id >= guiIDMenuExt && id < guiIDMenuExt+len(opts) {
		guiOptionsRes = opts[id-guiIDMenuExt]
		viper.Set(configVideoWidth, guiOptionsRes.X)
		viper.Set(configVideoHeight, guiOptionsRes.Y)
		writeConfigLater()
	}
	clientPlaySoundSpecial(sound.SoundShellClick, 100)
	return 1
}

func guiParseHook(name string, win *gui.Window) {
	name = strings.ToLower(name)
	switch name {
	case "options.wnd":
		guiEnhanceOptions(win)
	case "inputcfg.wnd":
		guiEnhanceInputCfg(win)
	}
}

func guiEnhanceInputCfg(root *gui.Window) {
	// "Back" is confusing, we rename it to "Apply"
	if w := root.ChildByID(932); w != nil {
		w.DrawData().SetText("Apply")
	}
}

func guiEnhanceOptions(root *gui.Window) {
	c := noxClient
	small := c.r.Fonts.FontPtrByName(noxfont.SmallName)
	// change resolution options to a new ones
	// if you decide to change these, check carefully in other places, especially in C
	resOpts := getResolutionOptions()
	// exactly 3 options in the menu (321-323)
	for i := 0; i < 3; i++ {
		if w := root.ChildByID(uint(321 + i)); w != nil {
			w.Hide()
		}
	}
	// add new resolution options
	const (
		resRows   = 5
		resHeight = 12
		resWidth  = 70
	)
	for i, res := range resOpts {
		if res == (image.Point{}) {
			continue
		}
		xi := i / resRows
		yi := i % resRows
		text := fmt.Sprintf("%4dx%d", res.X, res.Y)
		b := NewRadioButton(c.GUI, root, uint(guiIDMenuExt+i), 112+resWidth*xi, 135+resHeight*yi, resWidth, resHeight, 1, text)
		draw := b.DrawData()
		draw.SetFont(small)
		draw.SetImagePoint(image.Point{Y: -2})
	}

	// replace 8bit/16bit switch with window/fullscreen
	if w := root.ChildByID(331); w != nil { // 8bit
		w.DrawData().SetText("\tWindowed")
		w.Show() // keep enabled even in-game
	}
	if w := root.ChildByID(332); w != nil { // 16bit
		w.DrawData().SetText("\tFullscreen")
		w.Show() // keep enabled even in-game
	}

	// hide window size selection
	for i := uint(0); i < 5; i++ { // 310-314
		root.ChildByID(310 + i).Hide()
	}
	// add gamma and sensitivity sliders instead
	c.GUI.NewStaticText(root, 315, 112, 220, 140, 16, true, false, "Gamma")
	NewHorizontalSlider(root, 316, 120, 236, 120, 16, 1, 100).
		Func94(gui.AsWindowEvent(0x400A, uintptr(getGammaSlider()), 0))
	c.GUI.NewStaticText(root, 317, 112, 258, 140, 16, true, false, "Sensitivity")
	NewHorizontalSlider(root, 318, 120, 274, 120, 16, 1, 100).
		Func94(gui.AsWindowEvent(0x400A, uintptr((math.Log10(float64(noxClient.GetSensitivity()))+1.0)*50), 0))
}

func sub_4A1A40(a1 int) {
	v1 := guiOptsBack.ChildByID(151)
	legacy.Nox_xxx_wnd_46ABB0(v1, a1)
}

func guiSetBackButtonText(name strman.ID) {
	win := guiOptsBack
	backBtn := win.ChildByID(152)
	str := strMan.GetStringInFile(name, "C:\\NoxPost\\src\\client\\shell\\OptsBack.c")
	backBtn.Func94(&gui.StaticTextSetText{Str: str, Val: -1})
}

func sub_4AAA10() int {
	v0p := legacy.Get_nox_wnd_xxx_1309740()
	v0 := *v0p // copy before deletion
	v0p.Free()
	legacy.Get_dword_5d4594_1309720().Destroy()
	sub_4A1A40(1)
	v0.Func13()
	return 1
}

func sub_4C3A90(a1, a2 int, a3 unsafe.Pointer, a4 int) int {
	if a2 == 23 {
		return 1
	}
	if a2 != 16391 {
		return 0
	}
	c := noxClient
	win := legacy.AsWindowP(a3)
	clientPlaySoundSpecial(sound.SoundShellClick, 100)
	switch win.ID() {
	case 931:
		c.ctrl.Reset()
		nox_common_readcfgfile("nox.cfg", true)
		sub_4C3B70()
	case 932:
		legacy.Sub_4C35B0(0)
	case 933:
		sub_4C3CB0()
	case 971, 972, 973:
		legacy.Sub_430AA0(int(win.ID() - 971))
	}
	return 0
}

func sub_4CBE70(a1, a2 int, a3 unsafe.Pointer, a4 int) int {
	if a2 == 23 {
		return 1
	}
	if a2 != 16391 {
		return 0
	}
	c := noxClient
	win := legacy.AsWindowP(a3)
	clientPlaySoundSpecial(sound.SoundShellClick, 100)
	switch win.ID() {
	case 931:
		c.ctrl.Reset()
		nox_common_readcfgfile("nox.cfg", true)
		sub_4CBBF0()
	case 932:
		legacy.Sub_4CBD30()
	case 933:
		sub_4CBF40()
	case 971, 972, 973:
		legacy.Sub_430AA0(int(win.ID() - 971))
	}
	return 0
}
