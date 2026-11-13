package legacy

import (
	"fmt"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"image"
	"math"
	"runtime"
	"testing"
	"unsafe"
)

func TestGame3NearbyNativeCReferences(t *testing.T) {
	for _, ref := range []bool{false, true} {
		if p, ok := game3WidgetServerCount(ref); !ok || (unsafe.Sizeof(uintptr(0)) == 8 && p <= math.MaxUint32) {
			t.Fatal("native linked server-node count/identity")
		}
	}
	if p, ok := game3WidgetRegistryCheck(); !ok || (unsafe.Sizeof(uintptr(0)) == 8 && p <= math.MaxUint32) {
		t.Fatal("public nearby registry narrowed a server pointer")
	}

	for _, xy := range [][2]int32{{0, 0}, {215, 26}, {216, 27}, {400, 251}, {601, 452}, {math.MinInt32, math.MaxInt32}, {math.MaxInt32, math.MinInt32}, {-1, -1}} {
		t.Run(fmt.Sprintf("position/%d,%d", xy[0], xy[1]), func(t *testing.T) {
			got, want := game3WidgetReferencePosition(false, xy[0], xy[1]), game3WidgetReferencePosition(true, xy[0], xy[1])
			if got != want {
				t.Fatalf("native=%v C=%v", got, want)
			}
		})
	}
	t.Run("position/native-owners", func(t *testing.T) {
		want := game3WidgetReferencePosition(true, 216, 27)
		for _, ref := range []bool{false, true} {
			got, address, intact := game3WidgetPositionCOwner(ref, 216, 27)
			if !intact || got != want || (unsafe.Sizeof(uintptr(0)) == 8 && address <= math.MaxUint32) {
				t.Fatalf("C-owned position: ref=%t value=%v want=%v address=%x intact=%t", ref, got, want, address, intact)
			}
		}
		if got, same := game3WidgetPositionPinnedOwner(216, 27); !same || got != want {
			t.Fatalf("pinned Go position: value=%v want=%v returned-same-owner=%t", got, want, same)
		}
	})
	for _, r := range []float64{0, 1, 5, 15.5, 10000000000, math.Inf(1), math.NaN()} {
		for _, xy := range [][2]uint32{{0, 0}, {3, 4}, {0xffffffff, 0x80000000}, {0x7fffffff, 0xffffffff}} {
			t.Run(fmt.Sprintf("distance/%08x,%08x/%g", xy[0], xy[1], r), func(t *testing.T) {
				got, p := game3WidgetReferenceDistance(false, xy[0], xy[1], -32768, 32767, r)
				want, _ := game3WidgetReferenceDistance(true, xy[0], xy[1], -32768, 32767, r)
				if got != want {
					t.Fatalf("native=%v C=%v", got, want)
				}
				if unsafe.Sizeof(uintptr(0)) == 8 && p <= math.MaxUint32 {
					t.Fatal("C server allocation did not exercise full native pointer")
				}
			})
		}
	}
}

func TestGame3ListBoxNativeCReferences(t *testing.T) {
	for _, scroll := range []int{0, 7, 8, 20, 31, 65535} {
		for _, delta := range []int{-9, -1, 0, 1, 9} {
			for _, update := range []bool{false, true} {
				t.Run(fmt.Sprintf("scroll/%d/delta/%d/update/%t", scroll, delta, update), func(t *testing.T) {
					got, want := game3WidgetReferenceList(false, 1, scroll, delta, update), game3WidgetReferenceList(true, 1, scroll, delta, update)
					if got != want {
						t.Fatalf("native=%v C=%v", got, want)
					}
					if got[13] != 1 || got[15] != 1 {
						t.Fatal("list bank canary or full-pointer identity changed")
					}
					if unsafe.Sizeof(uintptr(0)) == 8 && got[14] != 1 {
						t.Fatal("list was not allocated above32bits")
					}
				})
			}
		}
	}
	for _, op := range []int{0, 2, 3} {
		for _, scroll := range []int{0, 7, 8, 20, 31, 65535} {
			t.Run(fmt.Sprintf("op/%d/scroll/%d", op, scroll), func(t *testing.T) {
				got, want := game3WidgetReferenceList(false, op, scroll, 0, true), game3WidgetReferenceList(true, op, scroll, 0, true)
				if got != want {
					t.Fatalf("native=%v C=%v", got, want)
				}
			})
		}
	}
	for i := -2; i <= 18; i++ {
		if got, want := game3ListColor(int32(i)), game3WidgetReferenceColor(i); got != want {
			t.Fatalf("palette %d native=%08x C=%08x", i, got, want)
		}
	}
}

func TestGame3SliderNativeCReferences(t *testing.T) {
	cases := []struct {
		name  string
		pre   bool
		event int
		a, b  uintptr
	}{
		{"mouse-down", false, 5, 0, 0}, {"drag", false, 6, uintptr(45 | 50<<16), 0}, {"drag-past-end", false, 7, uintptr(500 | 600<<16), 0},
		{"release-no-owner-style", false, 8, 0, 0}, {"enter-no-tab-style", false, 17, 0, 0}, {"leave-no-tab-style", false, 18, 0, 0},
		{"key-left", false, 21, 203, 2}, {"key-right", false, 21, 205, 2}, {"key-up", false, 21, 200, 2}, {"key-down", false, 21, 208, 2}, {"key-unknown", false, 21, 222, 2},
		{"set-value", true, 16394, 7, 0}, {"set-outside", true, 16394, 999, 0}, {"set-range", true, 16395, 3, 97},
		{"set-equal-range", true, 16395, 8, 8}, {"horizontal-drag", true, 0x4000, 0, uintptr(45 | 50<<16)},
		{"drag-before", true, 0x4000, 0, 0}, {"drag-after", true, 0x4000, 0, uintptr(500 | 600<<16)},
		{"focus", true, 23, 1, 0}, {"blur", true, 23, 0, 0}, {"drag-complete", true, 0x4007, 0, 0},
	}
	for _, vertical := range []bool{false, true} {
		for _, scale := range []float32{0, 0.125, 0.9, 2.25} {
			for _, tc := range cases {
				t.Run(fmt.Sprintf("vertical/%t/scale/%g/%s", vertical, scale, tc.name), func(t *testing.T) {
					got, want := game3WidgetReferenceSlider(false, vertical, tc.pre, tc.event, tc.a, tc.b, 3, 97, math.Float32bits(scale), 15), game3WidgetReferenceSlider(true, vertical, tc.pre, tc.event, tc.a, tc.b, 3, 97, math.Float32bits(scale), 15)
					if got != want {
						t.Fatalf("native=%v C=%v", got, want)
					}
					if got[14] != 1 || got[15] != 1 || got[17] != 1 {
						t.Fatal("widget ownership/canaries changed")
					}
					if unsafe.Sizeof(uintptr(0)) == 8 && got[16] != 1 {
						t.Fatal("native full-pointer case not exercised")
					}
				})
			}
		}
	}
}

func TestGame3WidgetsNativeLayoutAndPinnedSelection(t *testing.T) {
	layout := game3WidgetLayout()
	if layout[0] != unsafe.Sizeof(gui.Window{}) || layout[1] != unsafe.Sizeof(gui.WindowData{}) || layout[2] != unsafe.Sizeof(gui.ScrollListBoxData{}) || layout[3] != unsafe.Sizeof(gui.ScrollListBoxItem{}) || layout[4] != unsafe.Offsetof(gui.Window{}.WidgetData) || layout[5] != unsafe.Offsetof(gui.Window{}.Field100Ptr) || layout[6] != unsafe.Offsetof(gui.ScrollListBoxData{}.Field_12) || layout[7] != unsafe.Sizeof(game3SliderData{}) {
		t.Fatalf("C/Go layout mismatch %v", layout)
	}
	if p, ok := game3WidgetSelectionCheck(); !ok || (unsafe.Sizeof(uintptr(0)) == 8 && p <= math.MaxUint32) {
		t.Fatal("native selection copy/full-pointer/canary contract failed")
	}
	if runtime.GOOS != "windows" {
		if p, ok := game3WidgetGuardedSelectionCheck(); !ok || (unsafe.Sizeof(uintptr(0)) == 8 && p <= math.MaxUint32) {
			t.Fatal("exact Count C selection at guard-page boundary/canary failed")
		}
	}
	for _, vertical := range []bool{false, true} {
		if !game3WidgetDestroyCheck(vertical) {
			t.Fatal("destroy did not free and clear native slider data")
		}
	}
	// Go-owned pointer-bearing list metadata is individually pinned across a
	// synchronous public C call. Never retain a temporary Go pointer in C.
	d := new(gui.ScrollListBoxData)
	a := new([4]uint32)
	for i := range a {
		a[i] = uint32(10 + i)
	}
	d.Count = 4
	d.Field_12 = &a[0]
	var pin runtime.Pinner
	pin.Pin(d)
	pin.Pin(a)
	defer pin.Unpin()
	runtime.GC()
	game3WidgetPinnedSelection(unsafe.Pointer(d))
	runtime.GC()
	if *d.Field_12 != 11 || a[1] != 12 || a[2] != 13 || a[3] != math.MaxUint32 {
		t.Fatal("pinned Go full-pointer selection copy failed")
	}
	runtime.KeepAlive(a)
	runtime.KeepAlive(d)
}

func TestGame3WidgetsNativeConstructorAndExistingDelegates(t *testing.T) {
	g := gui.NewHigh(nil)
	defer g.DestroyAll()
	parent := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 320, 200, nil)
	for _, vertical := range []bool{false, true} {
		draw, freeDraw := alloc.NewHigh(gui.WindowData{})
		*draw = gui.WindowData{Window: parent, Style: gui.StyleHorizSlider}
		if vertical {
			draw.Style = gui.StyleVertSlider
		}
		opts, freeOpts := alloc.NewHigh(game3SliderData{})
		*opts = game3SliderData{Min: 7, Max: 7, Value: 100}
		win := game3WidgetCreateSlider(parent, draw, opts)
		if win == nil {
			freeOpts()
			freeDraw()
			t.Fatal("public C slider constructor failed")
		}
		if opts.Max != 8 || opts.Value != 100 || opts.Scale != float32(90) {
			t.Fatalf("constructor did not preserve caller options mutation: %+v", opts)
		}
		data := (*game3SliderData)(win.WidgetData)
		if *data != *opts || win.Field100() == nil || win.DrawData().Window != parent {
			t.Fatal("typed constructor ownership/copy mismatch")
		}
		win.Destroy()
		g.FreeDestroyed()
		freeOpts()
		freeDraw()
	}
	draw := gui.WindowData{Window: parent, Style: gui.StyleScrollListBox}
	list := gui.NewScrollListBoxRaw(g, parent, gui.StatusEnabled, 0, 0, 100, 50, &draw, &gui.ScrollListBoxData{Count: 2, Line_height: 10})
	if list == nil {
		t.Fatal("native listbox constructor")
	}
	game3WidgetInitList(list)
	// Existing native selection registry and pre-handler remain the same owner.
	text, free := alloc.CString16High("entry")
	defer free()
	if !gui.EventRespBool(list.Func94(gui.AsWindowEvent(0x400d, uintptr(unsafe.Pointer(text)), ^uintptr(0)))) {
		t.Fatal("existing Go native list handler no longer installed")
	}
	d := (*gui.ScrollListBoxData)(list.WidgetData)
	if d.Field_11_0 != 1 || d.Items == nil {
		t.Fatal("native list data was not updated")
	}
	list.SetPos(image.Pt(5, 9))
	if game3WidgetListDrawWithoutRenderer(list) != 1 {
		t.Fatal("public draw bridge")
	}
}

func TestGame3WidgetsNativeSliderCleanup(t *testing.T) {
	for _, vertical := range []bool{false, true} {
		for _, explicit := range []bool{false, true} {
			t.Run(fmt.Sprintf("vertical/%t/explicit/%t", vertical, explicit), func(t *testing.T) {
				g := gui.NewHigh(nil)
				defer g.DestroyAll()
				parent := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 320, 200, nil)
				draw, freeDraw := alloc.NewHigh(gui.WindowData{})
				defer freeDraw()
				*draw = gui.WindowData{Window: parent, Style: gui.StyleHorizSlider}
				if vertical {
					draw.Style = gui.StyleVertSlider
				}
				opts, freeOpts := alloc.NewHigh(game3SliderData{})
				defer freeOpts()
				*opts = game3SliderData{Min: 2, Max: 9, Value: 3}
				win := game3WidgetCreateSlider(parent, draw, opts)
				if win == nil || win.WidgetData == nil {
					t.Fatal("native slider must own its scalar data")
				}
				if explicit {
					win.Func94(gui.WindowDestroy{})
					if win.WidgetData != nil {
						t.Fatal("explicit event 2 did not clear the scalar data owner")
					}
				}
				calls := 0
				win.AddCleanup(func() {
					calls++
					if win.WidgetData != nil {
						t.Error("slider data survived its owner cleanup before pool free")
					}
				})
				win.Destroy()
				if calls != 0 {
					t.Fatal("slider cleanup ran before deferred GUI disposal")
				}
				g.FreeDestroyed()
				if calls != 1 {
					t.Fatalf("cleanup observer calls=%d, want 1", calls)
				}
				g.FreeDestroyed()
				if calls != 1 {
					t.Fatal("repeated GUI disposal reran owner cleanup")
				}
			})
		}
	}
}
