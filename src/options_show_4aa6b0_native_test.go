package opennox

import (
	"image"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/timer"
)

func TestOptionsShow4AA6B0CEntryNativeWidgetsAndAnimation(t *testing.T) {
	oldAnimState := gui.AnimGlobalState()
	t.Cleanup(func() { gui.SetAnimGlobalState(oldAnimState) })
	g := gui.New(nil)
	t.Cleanup(g.DestroyAll)
	root := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 640, 480, nil)
	root.SetID(300)
	oldRoot, oldAnim := legacy.Get_dword_5d4594_1309720(), legacy.Get_nox_wnd_xxx_1309740()
	oldChecks := [3]*gui.Window{legacy.OptionsCheckbox4AA6B0(0), legacy.OptionsCheckbox4AA6B0(1), legacy.OptionsCheckbox4AA6B0(2)}
	oldHook := legacy.NoxGameShowOptionsNative
	t.Cleanup(func() {
		legacy.OptionsStoreRoot4AA6B0(oldRoot)
		legacy.OptionsStoreAnimation4AA6B0(oldAnim)
		for index, win := range oldChecks {
			legacy.OptionsStoreCheckbox4AA6B0(int32(index), win)
		}
		legacy.NoxGameShowOptionsNative = oldHook
	})
	var sliders, checks [3]*gui.Window
	var currents [3]*timer.Timer
	values := [3]uint32{0x1234abcd, 0x40000123, 0x0001ffff}
	for index := range sliders {
		draw := gui.WindowData{Window: root, Style: gui.StyleHorizSlider}
		sliders[index] = gui.NewSliderRaw(g, root, gui.StatusEnabled, 10, 30*index, 160, 20, &draw, &gui.SliderData{Min: 0, Max: 100})
		sliders[index].SetID(uint(351 + index))
		checks[index] = g.NewWindowRaw(root, gui.StatusEnabled, 0, 30*index, 20, 20, nil)
		checks[index].SetID(uint(361 + index))
		checks[index].DrawData().Field0 = 0xa5a50104
		current, free := alloc.New(timer.Timer{})
		t.Cleanup(free)
		current.Current, currents[index] = values[index], current
	}
	h := optionsShowNativeHooks4AA6B0()
	h.addState = func(id int32) {
		if id != 300 {
			t.Fatal("state ID")
		}
	}
	h.newWindow = func(name string) *gui.Window {
		if name != "Options.wnd" {
			t.Fatal("window name")
		}
		return root
	}
	h.advanced = func(win *gui.Window) int32 {
		if win != root {
			t.Fatal("advanced root")
		}
		return -1
	}
	h.tabWidth = func(width int32) {
		if width != 15 {
			t.Fatal("tab width")
		}
	}
	h.newAnimation = func(win *gui.Window, args [8]int32) *gui.Anim {
		anim := gui.NewAnim(win, image.Pt(int(args[0]), int(args[1])), image.Pt(int(args[2]), int(args[3])), image.Pt(int(args[4]), int(args[5])), image.Pt(int(args[6]), int(args[7])))
		t.Cleanup(anim.Free)
		return anim
	}
	h.loadImage = func(string) noxrender.ImageHandle { return nil }
	h.current = func(index int32) uint32 { return currents[index].Current }
	h.enabled = func(index int32) int32 { return []int32{1, -1, 1}[index] }
	h.backText = func(string) {}
	h.backEnabled = func(enabled int32) {
		if enabled != 0 {
			t.Fatal("back enabled")
		}
	}
	h.video = func() {}
	legacy.NoxGameShowOptionsNative = func() int { return int(optionsShow4AA6B0(h)) }
	if got := legacy.OptionsShowCEntry4AA6B0(); got != 1 {
		t.Fatalf("C entry returned %d", got)
	}
	anim := legacy.Get_nox_wnd_xxx_1309740()
	if anim == nil || anim.Window() != root || anim.StateID != 300 || anim.State() != gui.AnimIn || root.Offs() != image.Pt(0, -480) {
		t.Fatalf("native animation=%p root=%p", anim, root)
	}
	if anim.Func12Ptr != legacy.OptionsStartOut4AA6B0() || anim.FncDoneOutPtr != legacy.OptionsDoneOut4AA6B0() {
		t.Fatal("animation callback identities were lost")
	}
	if legacy.Get_dword_5d4594_1309720() != root {
		t.Fatal("native root was not published")
	}
	for index, slider := range sliders {
		data := (*gui.SliderData)(slider.WidgetData)
		if data.Min != 0 || data.Max != 16384 || data.Field3 != values[index]>>16 || slider.Field100().Size() != image.Pt(24, 20) {
			t.Fatalf("slider %d=%+v thumb=%v", index, data, slider.Field100().Size())
		}
		if legacy.OptionsCheckbox4AA6B0(int32(index)) != checks[index] {
			t.Fatal("native checkbox identity")
		}
		want := uint32(0xa5a50104)
		if index == 1 {
			want &^= 4
		}
		if checks[index].DrawData().Field0 != want {
			t.Fatalf("checkbox %d flags=%08x", index, checks[index].DrawData().Field0)
		}
		if unsafe.Sizeof(uintptr(0)) == 8 {
			for _, ptr := range []unsafe.Pointer{root.C(), slider.C(), slider.WidgetData, slider.Field100().C(), checks[index].C(), unsafe.Pointer(anim), unsafe.Pointer(currents[index])} {
				if uintptr(ptr) <= math.MaxUint32 {
					t.Fatalf("C-owned native pointer %p did not exceed 4 GiB", ptr)
				}
			}
		}
	}
	t.Logf("C entry root=%p anim=%p sliders=%p/%p/%p", root, anim, sliders[0], sliders[1], sliders[2])
}

func TestOptionsShow4AA6B0CEntrySignedDwordReturn(t *testing.T) {
	oldHook := legacy.NoxGameShowOptionsNative
	t.Cleanup(func() { legacy.NoxGameShowOptionsNative = oldHook })
	for _, result := range []int32{math.MinInt32, -1, 0, 1, math.MaxInt32} {
		legacy.NoxGameShowOptionsNative = func() int { return int(result) }
		if got := legacy.OptionsShowCEntry4AA6B0(); got != result {
			t.Fatalf("C return=%d want=%d", got, result)
		}
	}
}
