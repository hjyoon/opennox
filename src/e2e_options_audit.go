package opennox

import (
	"fmt"
	"image"
	"math"
	"unsafe"

	"github.com/spf13/viper"

	"github.com/opennox/libs/cfg"
	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/client/seat"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/timer"
)

// The audit only queues user input and reads state. It does not call option
// handlers, repair settings, or enable configuration writes suppressed in E2E.
type optionsAudit struct {
	mode           int
	passed, failed int
}

func (a *optionsAudit) check(ok bool, name, detail string) {
	if ok {
		a.passed++
		e2eLog.Printf("OPTIONS AUDIT PASS: mode=%d %s %s", a.mode, name, detail)
	} else {
		a.failed++
		e2eLog.Printf("OPTIONS AUDIT FAIL: mode=%d %s %s", a.mode, name, detail)
	}
}

func (a *optionsAudit) root() *gui.Window {
	return noxClient.GUI.ChildByID(300)
}

func optionsAuditEnabled(win *gui.Window) bool {
	for cur := win; cur != nil; cur = cur.Parent() {
		if !cur.GetFlags().Has(gui.StatusEnabled) || cur.GetFlags().Has(gui.StatusHidden|gui.StatusDestroyed) {
			return false
		}
	}
	return win != nil
}

// The stock InputCfg.wnd leaves 932 disabled. The in-game constructor enables
// it, while the shell constructor applies pending bindings from its animated
// exit callback (004CBB70), reached through the separate Back button 152.
func optionsAuditInputExitControl(mode int) (uint, string) {
	switch mode {
	case 0:
		return 152, "Back"
	case 1:
		return 932, "Apply"
	default:
		panic(fmt.Sprintf("invalid options input audit mode %d", mode))
	}
}

// The original general-section writer (004332E0) does not persist tooltip
// visibility. NoSoftLights is also absent from its keys/formats, and ID 2052
// is a Go-side option. Do not invent legacy persistence requirements for
// these session settings. A future persistence extension needs its own test.
func optionsAuditLegacyVideoKey(id uint) string {
	switch id {
	case 2012:
		return "SoftShadowEdge"
	case 2014:
		return "TranslucentConsole"
	case 2015:
		return "RenderGlow"
	case 2016:
		return "FadeObjects"
	case 2020:
		return "DrawFrontWalls"
	case 2021:
		return "TranslucentFrontWalls"
	case 2022:
		return "HighResFrontWalls"
	case 2031:
		return "HighResFloors"
	case 2032:
		return "LockHighResFloors"
	case 2033:
		return "TexturedFloors"
	case 2040:
		return "RenderGUI"
	default:
		return ""
	}
}

func optionsAuditMove(pos image.Point) {
	e2eQueueRawInput(&seat.MouseMoveEvent{Pos: noxClient.Inp.DrawPosToWindow(pos), Relative: false})
}

func (a *optionsAudit) click(sc *e2eScenario, id uint) {
	a.clickWindow(sc, fmt.Sprintf("option %d", id), func() *gui.Window { return a.root().ChildByID(id) })
}

func (a *optionsAudit) clickWindow(sc *e2eScenario, name string, target func() *gui.Window) {
	var available bool
	sc.add(0, "audit click "+name, func() {
		win := target()
		available = optionsAuditEnabled(win)
		if !optionsAuditEnabled(win) {
			e2eLog.Printf("OPTIONS AUDIT UNAVAILABLE: mode=%d %s", a.mode, name)
			return
		}
		optionsAuditMove(win.GlobalPos().Add(image.Pt(8, win.Size().Y/2)))
	})
	sc.add(1, "", func() {
		if available {
			e2eQueueRawInput(&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
		}
	})
	sc.add(1, "", func() {
		if available {
			e2eQueueRawInput(&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
		}
	})
	sc.Wait(8, "")
}

func optionsAuditVolume(ind VolumeControl) (current, target int, enabled bool) {
	var ptr unsafe.Pointer
	switch ind {
	case VolumeFX:
		ptr, enabled = legacy.Get_dword_587000_127004(), legacy.Sub_453070() != 0
	case VolumeDialog:
		ptr, enabled = legacy.Get_dword_587000_122852(), legacy.Sub_44D990() != 0
	case VolumeMusic:
		ptr, enabled = legacy.Get_dword_587000_93164(), legacy.Sub_43DC30() != 0
	}
	if ptr == nil {
		return -1, -1, enabled
	}
	t := (*timer.Timer)(ptr)
	return int(t.Current >> 16), int(t.Target >> 16), enabled
}

func optionsAuditVideoValue(id uint) bool {
	switch id {
	case 2051:
		return noxClient.GetFiltering()
	case 2050:
		return noxClient.GetStretch()
	case 2012:
		return noxflags.HasEngine(noxflags.EngineSoftShadowEdge)
	case 2014:
		return guiCon.translucent
	case 2015:
		return noxClient.r.Part.RenderGlow
	case 2016:
		return legacy.Get_nox_client_fadeObjects_80836() != 0
	case 2017:
		return nox_client_showTooltips_80840
	case 2020:
		return nox_client_drawFrontWalls_80812
	case 2021:
		return legacy.Get_nox_client_translucentFrontWalls_805844() != 0
	case 2022:
		return legacy.Get_nox_client_highResFrontWalls_80820() != 0
	case 2031:
		return legacy.Get_nox_client_highResFloors_154952() != 0
	case 2032:
		return nox_client_lockHighResFloors_1193152
	case 2033:
		return nox_client_texturedFloors2_154960
	case 2052:
		return noxflags.HasEngine(noxflags.EngineNoSoftLights)
	case 2040:
		return nox_client_renderGUI_80828
	}
	panic(fmt.Sprintf("unknown option %d", id))
}

func (a *optionsAudit) dragSlider(sc *e2eScenario, id uint, fraction float64, target func() *gui.Window) {
	var available bool
	sc.add(0, fmt.Sprintf("audit drag slider %d to %.2f", id, fraction), func() {
		win := target()
		if win == nil {
			return
		}
		available = optionsAuditEnabled(win) && win.Field100() != nil
		if !available {
			a.check(false, fmt.Sprintf("slider %d available", id), fmt.Sprintf("window=%p", win))
			return
		}
		thumb := win.Field100()
		optionsAuditMove(thumb.GlobalPos().Add(thumb.Size().Div(2)))
	})
	sc.add(1, "", func() {
		if available {
			e2eQueueRawInput(&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
		}
	})
	sc.add(3, "", func() {
		if !available {
			return
		}
		win := target()
		if win == nil || win.Field100() == nil {
			return
		}
		thumb := win.Field100()
		offset := thumb.Size().Div(2)
		if win.DrawData().Style.IsVertSlider() {
			offset.Y += int(math.Round((1 - fraction) * float64(win.Size().Y-thumb.Size().Y)))
		} else {
			offset.X += int(math.Round(fraction * float64(win.Size().X-thumb.Size().X)))
		}
		optionsAuditMove(win.GlobalPos().Add(offset))
	})
	sc.add(2, "", func() {
		if available {
			e2eQueueRawInput(&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
		}
	})
	sc.Wait(8, "")
}

func (a *optionsAudit) slider(sc *e2eScenario, id uint, fraction float64) {
	var startupFX int
	if id == 351 {
		sc.add(0, "audit FX startup value before slider", func() {
			startupFX = configGetVolume(VolumeFX)
		})
	}
	a.dragSlider(sc, id, fraction, func() *gui.Window { return a.root().ChildByID(id) })
	sc.add(0, "", func() {
		win := a.root().ChildByID(id)
		if win == nil || win.WidgetData == nil {
			return
		}
		data := (*gui.SliderData)(win.WidgetData)
		want := float64(data.Min) + fraction*float64(data.Max-data.Min)
		a.check(math.Abs(float64(data.Field3)-want) <= float64(data.Max-data.Min)/60+1,
			fmt.Sprintf("slider %d input", id), fmt.Sprintf("range=%d..%d value=%d want=%.2f", data.Min, data.Max, data.Field3, want))
		switch id {
		case 316:
			want := min(gammaMax, gammaMin+gammaMax*float32(data.Field3)/100)
			a.check(math.Abs(float64(getGamma()-want)) < 0.0001, "gamma applied", fmt.Sprintf("got=%g want=%g", getGamma(), want))
			a.check(math.Abs(viper.GetFloat64(configVideoGamma)-float64(getGamma())) < 0.0001,
				"gamma persistence value", fmt.Sprintf("yaml=%g live=%g", viper.GetFloat64(configVideoGamma), getGamma()))
		case 318:
			want := float32(math.Pow(10, float64(data.Field3)/50-1))
			a.check(math.Abs(float64(noxClient.GetSensitivity()-want)) < 0.0001,
				"sensitivity applied", fmt.Sprintf("got=%g want=%g", noxClient.GetSensitivity(), want))
		case 351, 352, 353:
			ind := VolumeControl(id - 351)
			current, target, enabled := optionsAuditVolume(ind)
			// Mock AIL timers do not tick: Target verifies the GUI command,
			// while Current and physical playback are deliberately not asserted.
			a.check(data.Min == 0 && data.Max == VolumeMax && target == int(data.Field3),
				fmt.Sprintf("volume %d target applied", ind), fmt.Sprintf("current=%d target=%d startupGain=%d slider=%d max=%d enabled=%t", current, target, configGetVolume(ind), data.Field3, data.Max, enabled))
			a.check(enabled == (data.Field3 != 0), fmt.Sprintf("volume %d zero/enable link", ind), fmt.Sprintf("value=%d enabled=%t", data.Field3, enabled))
			if ind == VolumeFX {
				// This scalar is the saved startup setting, not live source gain.
				// The Target/enable checks above observe the GUI command; actual
				// native source mixing is covered by the OpenAL integration tests.
				a.check(configGetVolume(VolumeFX) == startupFX, "FX startup value preserved", fmt.Sprintf("before=%d after=%d liveTarget=%d", startupFX, configGetVolume(VolumeFX), target))
			}
		}
	})
}

func (sc *e2eScenario) AuditClientOptions(mode int, name string) {
	if mode != 0 && mode != 1 {
		panic(fmt.Sprintf("invalid options audit mode %d: use 0 (menu) or 1 (in-game)", mode))
	}
	a := &optionsAudit{mode: mode}
	if mode == 1 {
		sc.Key(keybind.KeyEsc, "audit open pause menu")
		sc.Wait(60, "")
		a.clickWindow(sc, "pause menu Options", func() *gui.Window { return noxClient.GUI.ChildByID(9005) })
		sc.Wait(80, "")
	}
	sc.add(0, name, func() {
		root := a.root()
		a.check(root != nil && optionsAuditEnabled(root), "root available", fmt.Sprintf("window=%p state=%d", root, noxClient.GameGetStateCode()))
		e2eLog.Printf("OPTIONS AUDIT NOTE: audio validation checks requested volume and mute flags, not audible playback; mock AIL timers and buffers are inactive")
		for id := uint(351); id <= 353; id++ {
			win := root.ChildByID(id)
			if win == nil || win.WidgetData == nil {
				a.check(false, fmt.Sprintf("volume %d initialized", id), "missing widget")
				continue
			}
			data := (*gui.SliderData)(win.WidgetData)
			vol, _, enabled := optionsAuditVolume(VolumeControl(id - 351))
			check := root.ChildByID(id + 10)
			checked := check != nil && check.DrawData().Field0&4 != 0
			a.check(data.Min == 0 && data.Max == VolumeMax && int(data.Field3) == vol && checked == enabled,
				fmt.Sprintf("volume %d initialized", id), fmt.Sprintf("range=%d..%d value=%d timer=%d checked=%t enabled=%t", data.Min, data.Max, data.Field3, vol, checked, enabled))
		}
	})
	ids := []uint{2051, 2050, 2012, 2014, 2015, 2016, 2017, 2020, 2021, 2022, 2031, 2032, 2033, 2052, 2040}
	for _, id := range ids {
		var before bool
		sc.add(0, "", func() {
			before = optionsAuditVideoValue(id)
			win := a.root().ChildByID(id)
			a.check(optionsAuditEnabled(win) && (win.DrawData().Field0&4 != 0) == before,
				fmt.Sprintf("video %d initial", id), fmt.Sprintf("state=%t window=%p", before, win))
		})
		a.click(sc, id)
		sc.add(0, "", func() {
			live := optionsAuditVideoValue(id)
			win := a.root().ChildByID(id)
			checked := win != nil && win.DrawData().Field0&4 != 0
			a.check(live != before && checked == live, fmt.Sprintf("video %d toggled", id), fmt.Sprintf("before=%t live=%t checked=%t", before, live, checked))
			if id == 2033 {
				a.check(nox_client_texturedFloors_154956 == live, "floor renderer linked", fmt.Sprintf("draw=%t option=%t", nox_client_texturedFloors_154956, live))
			}
			if id == 2040 {
				a.check(nox_xxx_xxxRenderGUI_587000_80832 == live, "GUI renderer linked", fmt.Sprintf("draw=%t option=%t", nox_xxx_xxxRenderGUI_587000_80832, live))
			}
			if id == 2051 || id == 2050 {
				key := configVideoFiltering
				if id == 2050 {
					key = configVideoStretch
				}
				a.check(viper.GetBool(key) == live, fmt.Sprintf("video %d persistence value", id), fmt.Sprintf("key=%s stored=%t live=%t", key, viper.GetBool(key), live))
				if id == 2050 {
					a.check(configDirty, "stretch schedules configuration save", "live GUI toggle; E2E still suppresses disk writes")
				}
			} else if key := optionsAuditLegacyVideoKey(id); key != "" {
				var section cfg.Section
				writeConfigLegacyMain(&section)
				got, present := section.Get(key)
				a.check(present && got == fmt.Sprint(bool2int(live)), fmt.Sprintf("video %d persistence value", id), fmt.Sprintf("key=%s present=%t serialized=%q live=%t", key, present, got, live))
			} else {
				e2eLog.Printf("OPTIONS AUDIT NOTE: mode=%d video %d is session-only; no original legacy persistence key; live=%t (not a disk round-trip claim)", a.mode, id, live)
			}
		})
		a.click(sc, id)
		sc.add(0, "", func() {
			live := optionsAuditVideoValue(id)
			a.check(live == before, fmt.Sprintf("video %d restored by click", id), fmt.Sprintf("live=%t original=%t", live, before))
		})
	}
	for i, res := range getResolutionOptions() {
		if res == (image.Point{}) {
			continue
		}
		id := uint(guiIDMenuExt + i)
		a.click(sc, id)
		sc.add(0, "", func() {
			win := a.root().ChildByID(id)
			checked := win != nil && win.DrawData().Field0&4 != 0
			stored := image.Pt(viper.GetInt(configVideoWidth), viper.GetInt(configVideoHeight))
			a.check(checked && guiOptionsRes == res && stored == res, fmt.Sprintf("resolution %d selected", id),
				fmt.Sprintf("checked=%t pending=%v yaml=%v want=%v", checked, guiOptionsRes, stored, res))
		})
	}
	if mode == 0 {
		var savedWindowMode string
		sc.add(0, "audit window-mode save checkpoint", func() {
			var section cfg.Section
			writeConfigLegacyMain(&section)
			savedWindowMode = sectionValueOptionsAudit(section, "Fullscreen")
		})
		for _, id := range []uint{331, 332, 331} {
			a.click(sc, id)
			sc.add(0, "", func() {
				want := int(id - 331)
				a.check(nox_video_getFullScreen() == want, fmt.Sprintf("window mode %d", id), fmt.Sprintf("got=%d want=%d", nox_video_getFullScreen(), want))
				lastResolutionID := uint(guiIDMenuExt + len(getResolutionOptions()) - 1)
				resolution := a.root().ChildByID(lastResolutionID)
				a.check(resolution != nil && resolution.DrawData().Field0&4 != 0, "window mode preserves resolution selection", fmt.Sprintf("mode=%d resolution=%d", want, lastResolutionID))
				var section cfg.Section
				writeConfigLegacyMain(&section)
				// The native setter changes the renderer immediately. The legacy
				// save checkpoint is synchronized by videoUpdateGameMode on
				// options close, a path deliberately suppressed during E2E.
				a.check(sectionValueOptionsAudit(section, "Fullscreen") == savedWindowMode, "open window mode preserves save checkpoint",
					fmt.Sprintf("live=%d serialized=%q checkpoint=%q", nox_video_getFullScreen(), sectionValueOptionsAudit(section, "Fullscreen"), savedWindowMode))
				e2eLog.Printf("OPTIONS AUDIT NOTE: mode=%d fullscreen live=%d saved=%q; options-close synchronization and disk restart are not exercised here", a.mode, nox_video_getFullScreen(), savedWindowMode)
			})
		}
	} else {
		for _, id := range []uint{331, 332} {
			sc.add(0, "", func() {
				win := a.root().ChildByID(id)
				e2eLog.Printf("OPTIONS AUDIT NOTE: mode=%d window mode %d enabled=%t flags=%v", a.mode, id, optionsAuditEnabled(win), win.GetFlags())
			})
		}
	}
	sc.Screen(fmt.Sprintf("options audit resolution selection mode %d", mode))
	for _, id := range []uint{316, 318, 351, 352, 353} {
		for _, value := range []float64{0, .5, 1, .3} {
			a.slider(sc, id, value)
		}
	}
	for _, id := range []uint{361, 362, 363} {
		ind := VolumeControl(id - 361)
		var before bool
		sc.add(0, "", func() { _, _, before = optionsAuditVolume(ind) })
		a.click(sc, id)
		sc.add(0, "", func() {
			_, _, enabled := optionsAuditVolume(ind)
			win := a.root().ChildByID(id)
			checked := win != nil && win.DrawData().Field0&4 != 0
			a.check(enabled != before && enabled == checked, fmt.Sprintf("mute %d toggled", id), fmt.Sprintf("before=%t enabled=%t checked=%t", before, enabled, checked))
		})
		a.click(sc, id)
	}
	sc.Screen(fmt.Sprintf("options audit controls mode %d", mode))
	a.click(sc, 341)
	sc.Wait(100, "")
	sc.add(0, "", func() {
		win := noxClient.GUI.ChildByID(900)
		a.check(optionsAuditEnabled(win), "input configuration opened", fmt.Sprintf("window=%p state=%d", win, noxClient.GameGetStateCode()))
	})
	sc.Screen(fmt.Sprintf("options audit input mode %d", mode))
	a.input(sc)
	if mode == 0 {
		sc.Key(keybind.KeyEsc, "audit leave main options")
		sc.Wait(80, "")
		sc.add(0, "", func() {
			a.check(a.root() == nil && noxClient.GameGetStateCode() == client.StateMainMenu, "escape closes options", fmt.Sprintf("window=%p state=%d", a.root(), noxClient.GameGetStateCode()))
		})
		a.clickWindow(sc, "Back fallback", func() *gui.Window {
			if a.root() == nil {
				return nil
			}
			return noxClient.GUI.ChildByID(152)
		})
		sc.Wait(80, "")
		sc.add(0, "", func() {
			a.check(a.root() == nil && noxClient.GameGetStateCode() == client.StateMainMenu, "Back closes options", fmt.Sprintf("window=%p state=%d", a.root(), noxClient.GameGetStateCode()))
			var section cfg.Section
			writeConfigLegacyMain(&section)
			// E2E intentionally suppresses actual resolution changes in the
			// legacy close path. Do not turn that guard into a false failure.
			e2eLog.Printf("OPTIONS AUDIT NOTE: mode=%d resolution on return actual=%v requested=%v legacy=%q; E2E resolution application and disk writes are suppressed", a.mode, noxClient.videoGetGameMode(), guiOptionsRes, sectionValueOptionsAudit(section, "VideoMode"))
		})
		sc.Screen("options audit Back returned to menu")
	} else {
		a.click(sc, 371)
		sc.Wait(40, "")
		sc.add(0, "", func() {
			a.check(!optionsAuditEnabled(a.root()), "in-game Close hides options", fmt.Sprintf("window=%p", a.root()))
		})
	}
	sc.add(0, "client options audit summary", func() {
		e2eLog.Printf("OPTIONS AUDIT SUMMARY: mode=%d passed=%d failed=%d", mode, a.passed, a.failed)
		if a.failed != 0 {
			e2eError(fmt.Errorf("client options audit: %d failed checks (%d passed)", a.failed, a.passed))
		}
	})
}

func optionsAuditListText(win *gui.Window, index int) string {
	if win == nil || win.WidgetData == nil {
		return ""
	}
	data := (*gui.ScrollListBoxData)(win.WidgetData)
	if data.Items == nil || index < 0 || index >= int(data.Field_11_0) {
		return ""
	}
	return alloc.GoString16(&unsafe.Slice(data.Items, int(data.Field_11_0))[index].Text[0])
}

func optionsAuditListFirst(win *gui.Window) int {
	if win == nil || win.WidgetData == nil {
		return 0
	}
	data := (*gui.ScrollListBoxData)(win.WidgetData)
	if data.Items == nil {
		return 0
	}
	for i, item := range unsafe.Slice(data.Items, int(data.Field_11_0)) {
		if item.Field_0 > uint32(data.Field_13_1) {
			return i
		}
	}
	return 0
}

func (a *optionsAudit) inputScroll(root *gui.Window, label string) {
	first := optionsAuditListFirst(root.ChildByID(910))
	offset := (*gui.ScrollListBoxData)(root.ChildByID(911).WidgetData).Field_13_1
	for _, id := range []uint{911, 912, 913} {
		win := root.ChildByID(id)
		row := optionsAuditListFirst(win)
		got := (*gui.ScrollListBoxData)(win.WidgetData).Field_13_1
		// The container may be partway through a row after dragging; the
		// original input dialog synchronizes columns by first visible row.
		a.check(row == first && got == offset, fmt.Sprintf("input %s synchronized %d", label, id),
			fmt.Sprintf("parentRow=%d columnRow=%d columnOffset=%d referenceOffset=%d", first, row, got, offset))
	}
}

func (a *optionsAudit) input(sc *e2eScenario) {
	var available bool
	var originalPrimary string
	root := func() *gui.Window { return noxClient.GUI.ChildByID(900) }
	sc.add(0, "audit input configuration controls", func() {
		available = optionsAuditEnabled(root())
		if !available {
			e2eLog.Printf("OPTIONS AUDIT NOTE: mode=%d input configuration controls not reachable", a.mode)
			return
		}
		for _, id := range []uint{910, 911, 912, 913} {
			win := root().ChildByID(id)
			if win == nil || win.WidgetData == nil {
				a.check(false, fmt.Sprintf("input list %d populated", id), "missing widget")
				continue
			}
			data := (*gui.ScrollListBoxData)(win.WidgetData)
			a.check(data.Field_11_0 > 0, fmt.Sprintf("input list %d populated", id), fmt.Sprintf("count=%d first=%q", data.Field_11_0, optionsAuditListText(win, 0)))
		}
		originalPrimary = optionsAuditListText(root().ChildByID(912), 0)
	})
	inputWin := func(id uint) func() *gui.Window {
		return func() *gui.Window {
			if !available {
				return nil
			}
			return root().ChildByID(id)
		}
	}
	// Return from Middle to Left so Left is tested as an actual setting change,
	// not only as a click on the already-selected startup default.
	for _, id := range []uint{971, 972, 973, 971} {
		sc.add(0, "", func() {
			if available {
				a.check(optionsAuditEnabled(root().ChildByID(id)), fmt.Sprintf("input pickup %d available", id), "mouse pickup selector")
			}
		})
		a.clickWindow(sc, fmt.Sprintf("input pickup %d", id), inputWin(id))
		sc.add(0, "", func() {
			if !available || !optionsAuditEnabled(root().ChildByID(id)) {
				return
			}
			want := int(id - 971)
			var section cfg.Section
			writeConfigHotkeys(&section)
			win := root().ChildByID(id)
			checked := win != nil && win.DrawData().Field0&4 != 0
			exclusive := true
			for _, other := range []uint{971, 972, 973} {
				peer := root().ChildByID(other)
				exclusive = exclusive && peer != nil && (peer.DrawData().Field0&4 != 0) == (other == id)
			}
			serialized := sectionValueOptionsAudit(section, "MousePickup")
			a.check(int(legacy.Nox_client_mousePriKey_430AF0()) == want && checked && exclusive && serialized == noxMouseSelectOpt[want],
				fmt.Sprintf("input pickup %d applied", id), fmt.Sprintf("live=%d checked=%t exclusive=%t serialized=%q", legacy.Nox_client_mousePriKey_430AF0(), checked, exclusive, serialized))
		})
	}
	for _, id := range []uint{933, 931} {
		a.rebind(sc, &available, 912, keybind.KeyF10)
		a.clickWindow(sc, fmt.Sprintf("input defaults/reset %d", id), inputWin(id))
		sc.add(0, "", func() {
			if !available {
				return
			}
			win := root().ChildByID(911)
			data := (*gui.ScrollListBoxData)(win.WidgetData)
			got := optionsAuditListText(root().ChildByID(912), 0)
			a.check(data.Field_11_0 > 0 && got == originalPrimary,
				fmt.Sprintf("input defaults/reset %d discards unapplied edit", id), fmt.Sprintf("count=%d primary=%q original=%q", data.Field_11_0, got, originalPrimary))
			e2eLog.Printf("OPTIONS AUDIT NOTE: E2E uses deterministic hotkeys for both nox.cfg and default.cfg; original defaults and disk reset require a separate configuration round trip")
		})
	}
	for _, id := range []uint{922, 922, 922, 921, 921, 921} {
		a.clickWindow(sc, fmt.Sprintf("input scroll %d", id), inputWin(id))
		sc.add(0, "", func() {
			if !available {
				return
			}
			first := (*gui.ScrollListBoxData)(root().ChildByID(910).WidgetData).Field_13_1
			a.inputScroll(root(), fmt.Sprintf("scroll %d", id))
			if id == 922 {
				a.check(first > 0, "input down scroll moved", fmt.Sprintf("offset=%d", first))
			}
		})
	}
	for _, fraction := range []float64{.5, 0, 1} {
		a.dragSlider(sc, 920, fraction, inputWin(920))
		sc.add(0, "", func() {
			if !available {
				return
			}
			first := (*gui.ScrollListBoxData)(root().ChildByID(910).WidgetData).Field_13_1
			a.inputScroll(root(), "scrollbar")
			a.check((fraction == 1 && first == 0) || (fraction != 1 && first > 0), "input scrollbar moved", fmt.Sprintf("offset=%d fraction=%.2f", first, fraction))
		})
	}
	var event keybind.Event
	sc.add(0, "audit first event identity", func() {
		if !available {
			return
		}
		for _, ev := range keyBinding.Events() {
			if ev.Event != keybind.EventToggleQuitMenu {
				event = ev.Event
				break
			}
		}
	})
	a.rebind(sc, &available, 912, keybind.KeyF10)
	a.rebind(sc, &available, 913, keybind.KeyF11)
	a.rebind(sc, &available, 913, keybind.KeyEsc)
	exitID, exitName := optionsAuditInputExitControl(a.mode)
	sc.add(0, "audit input exit control", func() {
		if !available {
			return
		}
		apply := root().ChildByID(932)
		if a.mode == 0 {
			a.check(apply != nil && !apply.GetFlags().Has(gui.StatusEnabled), "menu input Apply disabled as in stock", fmt.Sprintf("window=%p", apply))
		} else {
			a.check(optionsAuditEnabled(apply), "in-game input Apply available", fmt.Sprintf("window=%p", apply))
		}
	})
	a.clickWindow(sc, "input "+exitName, func() *gui.Window {
		if !available {
			return nil
		}
		if a.mode == 0 {
			return noxClient.GUI.ChildByID(exitID)
		}
		return root().ChildByID(exitID)
	})
	// The shell exits InputCfg and then animates a new Options window in.
	sc.Wait(80, "")
	sc.add(0, "", func() {
		if !available {
			return
		}
		a.check(noxClient.ctrl.hasDefBinding(event, keybind.KeyF10), "input "+exitName+" updates binding", fmt.Sprintf("event=%v title=%q", event, noxClient.ctrl.Sub_42E8E0_go(event, 1)))
		a.check(noxClient.ctrl.hasDefBinding(event, keybind.KeyF11), "input "+exitName+" updates secondary binding", fmt.Sprintf("event=%v", event))
		a.check(!optionsAuditEnabled(root()) && optionsAuditEnabled(a.root()), "input "+exitName+" returns to options", fmt.Sprintf("input=%p options=%p", root(), a.root()))
		var section cfg.Section
		writeConfigHotkeys(&section)
		for _, key := range []keybind.Key{keybind.KeyF10, keybind.KeyF11} {
			got, present := section.Get(key.String())
			a.check(present && got != "", "input binding serialized", fmt.Sprintf("key=%s present=%t event=%q", key, present, got))
		}
	})
	if a.mode == 0 {
		sc.Screen("options audit input Back applied")
		a.clickWindow(sc, "reopen input after Back", func() *gui.Window {
			if !optionsAuditEnabled(a.root()) {
				return nil
			}
			return a.root().ChildByID(341)
		})
		sc.Wait(100, "")
		sc.add(0, "audit input bindings after reopening", func() {
			available = optionsAuditEnabled(root())
			a.check(available, "input configuration reopened", fmt.Sprintf("window=%p state=%d", root(), noxClient.GameGetStateCode()))
			if !available {
				return
			}
			for _, binding := range []struct {
				column uint
				key    keybind.Key
			}{{912, keybind.KeyF10}, {913, keybind.KeyF11}} {
				got, want := optionsAuditListText(root().ChildByID(binding.column), 0), binding.key.Title(noxClient.Strings())
				a.check(got == want, "input Back binding survives reopen", fmt.Sprintf("column=%d title=%q want=%q", binding.column, got, want))
			}
		})
		a.rebind(sc, &available, 912, keybind.KeyF9)
		a.rebind(sc, &available, 913, keybind.KeyF12)
		sc.add(0, "audit input focus before Escape", func() {
			e2eLog.Printf("OPTIONS AUDIT NOTE: mode=%d before input Escape focus=%p capture=%p prompt=%p animation_global=%d", a.mode,
				noxClient.GUI.Focused(), noxClient.GUI.Captured(), noxClient.GUI.ChildByID(980), gui.AnimGlobalState())
		})
		sc.Key(keybind.KeyEsc, "audit leave input with Escape")
		sc.Wait(80, "")
		sc.add(0, "audit input Escape applied", func() {
			if !available {
				return
			}
			for _, key := range []keybind.Key{keybind.KeyF9, keybind.KeyF12} {
				a.check(noxClient.ctrl.hasDefBinding(event, key), "input Escape updates binding", fmt.Sprintf("key=%s event=%v", key, event))
			}
			for _, key := range []keybind.Key{keybind.KeyF10, keybind.KeyF11} {
				a.check(!noxClient.ctrl.hasDefBinding(event, key), "input Escape replaces previous binding", fmt.Sprintf("key=%s event=%v", key, event))
			}
			a.check(root() == nil && optionsAuditEnabled(a.root()) && noxClient.GameGetStateCode() == client.StateOptions,
				"input Escape returns to options", fmt.Sprintf("input=%p options=%p state=%d", root(), a.root(), noxClient.GameGetStateCode()))
		})
		sc.Screen("options audit input Escape exit attempt")
		// Preserve the Escape failures above, but use real Back input if needed
		// so one failed route does not prevent the later options-close checks.
		a.clickWindow(sc, "input Back fallback after Escape", func() *gui.Window {
			if !optionsAuditEnabled(root()) {
				return nil
			}
			return noxClient.GUI.ChildByID(152)
		})
		sc.Wait(80, "")
		sc.add(0, "audit input return after exit attempts", func() {
			if !available {
				return
			}
			a.check(root() == nil && optionsAuditEnabled(a.root()) && noxClient.GameGetStateCode() == client.StateOptions,
				"input exit attempts return to options", fmt.Sprintf("input=%p options=%p state=%d", root(), a.root(), noxClient.GameGetStateCode()))
			for _, key := range []keybind.Key{keybind.KeyF9, keybind.KeyF12} {
				a.check(noxClient.ctrl.hasDefBinding(event, key), "input exit attempts apply binding", fmt.Sprintf("key=%s event=%v", key, event))
			}
		})
	}
}

func (a *optionsAudit) rebind(sc *e2eScenario, available *bool, column uint, key keybind.Key) {
	var ready, prompted bool
	var before string
	root := func() *gui.Window { return noxClient.GUI.ChildByID(900) }
	sc.add(0, fmt.Sprintf("audit rebind column %d to %s", column, key), func() {
		if !*available {
			return
		}
		win := root().ChildByID(column)
		ready = optionsAuditEnabled(win) && win.WidgetData != nil
		if !ready {
			a.check(false, "input binding column available", fmt.Sprintf("column=%d", column))
			return
		}
		before = optionsAuditListText(win, 0)
		data := (*gui.ScrollListBoxData)(win.WidgetData)
		optionsAuditMove(win.GlobalPos().Add(image.Pt(8, int(data.Line_height)/2)))
	})
	for _, pressed := range []bool{true, false} {
		sc.add(1, "", func() {
			if ready {
				e2eQueueRawInput(&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: pressed})
			}
		})
	}
	sc.Wait(15, "")
	sc.add(0, "", func() {
		if ready {
			prompted = optionsAuditEnabled(noxClient.GUI.ChildByID(980))
			a.check(prompted, "input rebind prompt opens", fmt.Sprintf("column=%d", column))
		}
	})
	for _, pressed := range []bool{true, false} {
		sc.add(1, "", func() {
			if prompted {
				e2eQueueRawInput(&seat.KeyboardEvent{Key: key, Pressed: pressed})
			}
		})
	}
	sc.Wait(15, "")
	sc.add(0, "", func() {
		if !prompted {
			return
		}
		got := optionsAuditListText(root().ChildByID(column), 0)
		want, label := key.Title(noxClient.Strings()), "input key captured"
		if key == keybind.KeyEsc {
			want, label = before, "input rebind cancellation preserves binding"
		}
		a.check(!optionsAuditEnabled(noxClient.GUI.ChildByID(980)) && got == want, label, fmt.Sprintf("column=%d title=%q want=%q", column, got, want))
	})
}

func sectionValueOptionsAudit(section cfg.Section, key string) string {
	value, _ := section.Get(key)
	return value
}
