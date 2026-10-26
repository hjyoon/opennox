package legacy

import (
	"image"
	"image/color"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noximage"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// The 32-bit C draw callbacks use the client's renderer directly. Supply the
// same software renderer to the owning GUI and client on both architectures.
// Unused root-renderer extensions still fail if unexpectedly called.
type widgetCallbackUnusedRender interface {
	SetRectFullScreen()
	SetColorMultAndIntensityRGB(cr, cg, cb byte) byte
	DrawCircleSegment(cx, cy, rad, ang int, cl color.Color)
	DrawImageAt(img *noxrender.Image, pos image.Point)
	Sub_49F7C0_def_go()
}

type widgetCallbackRender struct {
	*noxrender.NoxRender
	widgetCallbackUnusedRender
}

type widgetCallbackClient struct {
	Client
	cli    *client.Client
	render *widgetCallbackRender
}

func (c *widgetCallbackClient) Cli() *client.Client { return c.cli }
func (c *widgetCallbackClient) R2() Render2         { return c.render }

func newWidgetCallbackGUI(t *testing.T) *gui.GUI {
	t.Helper()
	data, free := noxrender.NewRenderData()
	r := noxrender.NewRender(nil, nil)
	r.SetData(data)
	r.SetPixBuffer(noximage.NewImage16(image.Rect(0, 0, 640, 480)))
	g := gui.New(r)
	c := &widgetCallbackClient{cli: &client.Client{GUI: g}, render: &widgetCallbackRender{NoxRender: r}}
	old := GetClient
	GetClient = func() Client { return c }
	t.Cleanup(func() { GetClient = old; free() })
	return g
}

func requireNativeWindowAddress(t *testing.T, win *gui.Window) {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(win.C()) <= math.MaxUint32 {
		t.Fatalf("test window address = %p, want a pointer above the PE32 range", win.C())
	}
}

func TestEditEventCallbackPreservesNativeWindowPointer(t *testing.T) {
	g := newWidgetCallbackGUI(t)
	defer g.DestroyAll()

	parent := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 640, 480, nil)
	draw := gui.WindowData{Window: parent, Style: gui.StyleEntryField}
	win := gui.NewEntryFieldRaw(g, parent, gui.StatusEnabled, 0, 0, 120, 20, &draw, &gui.EntryFieldData{Field_1040: 32})
	if win == nil {
		t.Fatal("NewEntryFieldRaw returned nil")
	}
	requireNativeWindowAddress(t, win)

	// This is the exact event and packed cursor position from the reported
	// crash. The PE32 implementation faulted while reading window offset 44.
	if got := editEventCallbackC(win, 17, 0x005c0210, 0); got != 1 {
		t.Fatalf("edit callback response = %d, want 1", got)
	}
	if got, want := editPreCallbackC(win, 0x401d, 0, 0), uintptr(win.WidgetData); got != want {
		t.Fatalf("edit data pointer = %#x, want %#x", got, want)
	}
	if got := editDrawCallbackC(win); got != 1 {
		t.Fatalf("edit draw callback response = %d, want 1", got)
	}

	cDraw := gui.WindowData{Window: parent, Style: gui.StyleEntryField}
	cWin := editConstructorC(parent, &cDraw, &gui.EntryFieldData{Field_1040: 32})
	if cWin == nil {
		t.Fatal("C entry-field constructor returned nil")
	}
	requireNativeWindowAddress(t, cWin)
	if got, want := editPreCallbackC(cWin, 0x401d, 0, 0), uintptr(cWin.WidgetData); got != want {
		t.Fatalf("C entry-field data pointer = %#x, want %#x", got, want)
	}
}

func TestListBoxCallbacksPreserveNativePointers(t *testing.T) {
	for _, multi := range []bool{false, true} {
		t.Run(map[bool]string{false: "single", true: "multi"}[multi], func(t *testing.T) {
			g := newWidgetCallbackGUI(t)
			defer g.DestroyAll()

			parent := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 640, 480, nil)
			draw := gui.WindowData{Window: parent, Style: gui.StyleScrollListBox | gui.StyleMouseTrack}
			opts := gui.ScrollListBoxData{Count: 4, Line_height: 10}
			if multi {
				opts.Field_4 = 1
			}
			win := gui.NewScrollListBoxRaw(g, parent, gui.StatusEnabled, 10, 20, 120, 60, &draw, &opts)
			if win == nil {
				t.Fatal("NewScrollListBoxRaw returned nil")
			}
			requireNativeWindowAddress(t, win)

			text := alloc.InternCString16("alpha")
			if got := listBoxPreCallbackC(win, 0x400d, uintptr(unsafe.Pointer(text)), ^uintptr(0)); got != 1 {
				t.Fatalf("add-line callback response = %d, want 1", got)
			}
			gotText := listBoxPreCallbackC(win, 0x4016, 0, 0)
			if gotText == 0 || alloc.GoString16((*uint16)(unsafe.Pointer(gotText))) != "alpha" {
				t.Fatalf("item text pointer = %#x, want alpha", gotText)
			}
			if unsafe.Sizeof(uintptr(0)) > 4 && gotText <= math.MaxUint32 {
				t.Fatalf("item text pointer = %#x, want a native-width address", gotText)
			}
			if got := listBoxEventCallbackC(win, multi, 17, 0x005c0210, 0); got != 1 {
				t.Fatalf("list-box callback response = %d, want 1", got)
			}

			// The legacy initializer used to reinstall PE32 draw and event
			// callbacks. It must retain the native Go callbacks instead.
			listBoxInitCallbackC(win)
			win.Draw()
		})
	}
}

func TestInputConfigCallbacksPreserveNativeWindowPointer(t *testing.T) {
	for _, shell := range []bool{false, true} {
		t.Run(map[bool]string{false: "in_game", true: "shell"}[shell], func(t *testing.T) {
			g := newWidgetCallbackGUI(t)
			defer g.DestroyAll()

			parent := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 640, 480, nil)
			draw := gui.WindowData{Window: parent, Style: gui.StyleScrollListBox | gui.StyleMouseTrack}
			win := gui.NewScrollListBoxRaw(
				g, parent, gui.StatusEnabled, 10, 20, 120, 60, &draw,
				&gui.ScrollListBoxData{Count: 4, Line_height: 10},
			)
			if win == nil {
				t.Fatal("NewScrollListBoxRaw returned nil")
			}
			requireNativeWindowAddress(t, win)

			// Event 17 makes the list callback inspect the native Window, while
			// the control callback immediately reads its native widget data.
			if got := inputConfigListCallbackC(win, shell, 17, 0x005c0210, 0); got != 1 {
				t.Fatalf("list callback response = %d, want 1", got)
			}
			if got := inputConfigControlCallbackC(win, shell, 17, 0x005c0210, 0); got != 0 {
				t.Fatalf("control callback response = %d, want 0", got)
			}
		})
	}
}
