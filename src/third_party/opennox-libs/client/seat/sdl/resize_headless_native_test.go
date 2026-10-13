//go:build opennox_sdl_resize_test

package sdl

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"testing"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/noximage"
	"github.com/opennox/opennox/v1/client/input"
	"github.com/opennox/opennox/v1/client/render"
	"github.com/opennox/opennox/v1/client/seat/headless"
	native "github.com/veandco/go-sdl2/sdl"
)

const resizeNativeChild = "OPENNOX_TEST_SDL_RESIZE_R179"

var resizeNativeCases = []string{"api", "duplicates", "resized-payload", "restored", "display-change", "other-window", "mouse-order", "retired-renderer", "focus-quit", "empty"}

// Run the real SDL dummy video driver/event queue, Window.InputTick and game
// renderer/input handler without an OS window or an OpenGL context. Only the
// surface/presentation backend is the production headless seat. This does not
// certify a physical Retina display or OpenGL output.
func TestSDLResizeNativeHeadless(t *testing.T) {
	if name := os.Getenv(resizeNativeChild); name != "" {
		testSDLResizeNativeHeadless(t, name)
		return
	}
	for _, name := range resizeNativeCases {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestSDLResizeNativeHeadless$", "-test.v")
			cmd.Env = append(os.Environ(), resizeNativeChild+"="+name, "SDL_VIDEODRIVER=dummy")
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("native SDL resize failed: %v\n%s", err, out)
			}
			if !bytes.Contains(out, []byte("SDL RESIZE HEADLESS PASS: "+name)) {
				t.Fatalf("missing native event/render/input receipt: %s", out)
			}
			t.Logf("SDL RESIZE HEADLESS PASS: %s", name)
		})
	}
}

type resizeHeadlessSeat struct {
	*Window
	output *headless.Seat
}

func (s *resizeHeadlessSeat) NewSurface(sz image.Point, filtering bool) seat.Surface {
	return s.output.NewSurface(sz, filtering)
}
func (s *resizeHeadlessSeat) Clear()   { s.output.Clear() }
func (s *resizeHeadlessSeat) Present() { s.output.Present() }

func testSDLResizeNativeHeadless(t *testing.T, name string) {
	if !slices.Contains(resizeNativeCases, name) {
		t.Fatal("unknown SDL resize fixture")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if os.Getenv("SDL_VIDEODRIVER") != "dummy" {
		t.Fatal("physical video driver must never be used")
	}
	if err := native.Init(native.INIT_VIDEO | native.INIT_TIMER); err != nil {
		t.Fatal(err)
	}
	defer native.Quit()
	if driver, err := native.GetCurrentVideoDriver(); err != nil || driver != "dummy" {
		t.Fatal("SDL did not select the required headless video driver")
	}
	win, err := native.CreateWindow("resize-native-headless", 0, 0, 640, 480, native.WINDOW_HIDDEN|native.WINDOW_RESIZABLE)
	if err != nil {
		t.Fatal(err)
	}
	defer win.Destroy()
	id, err := win.GetID()
	if err != nil {
		t.Fatal(err)
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	w := &Window{win: win, log: log}
	if w.ScreenSize() != image.Pt(640, 480) || win.GetFlags()&native.WINDOW_OPENGL != 0 {
		t.Fatal("SDL dummy drawable size differs or an OpenGL window was created")
	}
	// Drain native creation events before installing the observers.
	w.InputTick()
	screen := headless.New(image.Pt(640, 480))
	defer screen.Close()
	st := &resizeHeadlessSeat{Window: w, output: screen}
	r, err := render.New(st)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	pix := noximage.NewImage16(image.Rect(0, 0, 64, 48))
	for i := range pix.Pix {
		pix.Pix[i] = uint16(1 + i%0x7fff)
	}
	r.CopyBuffer(pix)
	h := input.New(log, w, false, 0)
	h.SetDrawWinSize(image.Pt(640, 480))
	r.OnViewResize(h.SetWinSize)
	quit := 0
	h.OnQuit(func() { quit++ })
	var inputEvents []seat.WindowEvent
	w.OnInput(func(ev seat.InputEvent) {
		if v, ok := ev.(seat.WindowEvent); ok {
			inputEvents = append(inputEvents, v)
		}
	})
	var sizes []image.Point
	second := 0
	w.OnScreenResize(func(sz image.Point) { sizes = append(sizes, sz) })
	w.OnScreenResize(func(sz image.Point) {
		second++
		if sz != w.ScreenSize() {
			t.Fatal("resize notification used event payload instead of current drawable size")
		}
	})
	beforePresent := screen.PresentCount()
	pushWindow := func(subtype native.WindowEventID, windowID uint32) {
		t.Helper()
		if got := resizeNativePushWindow(windowID, uint8(subtype)); got != 1 {
			t.Fatalf("native SDL queue rejected window event: %d", got)
		}
	}
	changed := name != "other-window" && name != "focus-quit" && name != "empty"
	if changed || name == "other-window" {
		win.SetSize(800, 450)
		if w.ScreenSize() != image.Pt(800, 450) {
			t.Fatal("real SDL API did not resize the dummy window")
		}
	}
	switch name {
	case "duplicates":
		pushWindow(native.WINDOWEVENT_RESIZED, id)
		pushWindow(native.WINDOWEVENT_SIZE_CHANGED, id)
	case "resized-payload", "restored", "display-change":
		// Isolate each event subtype while keeping the actual resized window.
		native.FlushEvents(native.WINDOWEVENT, native.WINDOWEVENT)
		subtype := map[string]native.WindowEventID{"resized-payload": native.WINDOWEVENT_RESIZED, "restored": native.WINDOWEVENT_RESTORED, "display-change": native.WINDOWEVENT_DISPLAY_CHANGED}[name]
		pushWindow(subtype, id)
	case "other-window":
		other, err := native.CreateWindow("other-headless", 0, 0, 320, 240, native.WINDOW_HIDDEN)
		if err != nil {
			t.Fatal(err)
		}
		defer other.Destroy()
		otherID, err := other.GetID()
		if err != nil {
			t.Fatal(err)
		}
		native.FlushEvents(native.WINDOWEVENT, native.WINDOWEVENT)
		pushWindow(native.WINDOWEVENT_SIZE_CHANGED, otherID)
	case "mouse-order":
		if got := resizeNativePushMouse(id); got != 1 {
			t.Fatalf("native SDL queue rejected mouse event: %d", got)
		}
	case "retired-renderer":
		old, err := render.New(st)
		if err != nil {
			t.Fatal(err)
		}
		old.CopyBuffer(pix)
		old.Close()
		beforePresent = screen.PresentCount()
	case "focus-quit":
		pushWindow(native.WINDOWEVENT_FOCUS_LOST, id)
		pushWindow(native.WINDOWEVENT_FOCUS_GAINED, id)
		if got := resizeNativePushQuit(); got != 1 {
			t.Fatalf("native SDL queue rejected quit event: %d", got)
		}
	}
	h.Tick() // real SDL PollEvent -> seat callbacks -> renderer and mouse handler
	wantCalls := 0
	wantView := image.Rect(0, 0, 640, 480)
	if changed {
		wantCalls = 1
	}
	// The renderer's existing float32 viewport arithmetic is part of the
	// contract. Compare with an independently resized headless renderer,
	// including its mouse coordinate mapping, instead of idealized geometry.
	controlSize := image.Pt(640, 480)
	if changed {
		controlSize = image.Pt(800, 450)
	}
	control := headless.New(controlSize)
	defer control.Close()
	controlRender, err := render.New(control)
	if err != nil {
		t.Fatal(err)
	}
	defer controlRender.Close()
	controlRender.CopyBuffer(pix)
	_, wantView = control.Snapshot()
	controlInput := input.New(log, control, false, 0)
	controlInput.SetDrawWinSize(image.Pt(640, 480))
	controlRender.OnViewResize(controlInput.SetWinSize)
	control.QueueInput(&seat.MouseMoveEvent{Pos: image.Pt(400, 225)})
	controlInput.Tick()
	if len(sizes) != wantCalls || second != wantCalls {
		t.Fatalf("native resize notification count=%d/%d want=%d", len(sizes), second, wantCalls)
	}
	if changed && sizes[0] != image.Pt(800, 450) {
		t.Fatalf("native drawable size=%v", sizes[0])
	}
	shown, view := screen.Snapshot()
	if shown == nil || view != wantView || !slices.Equal(shown.Pix, pix.Pix) || screen.PresentCount() != beforePresent+uint64(wantCalls) {
		t.Fatalf("SDL event did not preserve frame/redraw viewport: view=%v want=%v presentations=%d/%d", view, wantView, screen.PresentCount(), beforePresent+uint64(wantCalls))
	}
	if name == "mouse-order" && h.GetMousePos() != controlInput.GetMousePos() {
		t.Fatalf("first mouse event used old viewport: canvas=%v want=%v", h.GetMousePos(), controlInput.GetMousePos())
	}
	if name == "focus-quit" && (!slices.Equal(inputEvents, []seat.WindowEvent{seat.WindowUnfocused, seat.WindowFocused, seat.WindowClosed}) || quit != 1) {
		t.Fatal("resize handling changed existing focus/quit events")
	}
	if name == "duplicates" {
		pushWindow(native.WINDOWEVENT_RESIZED, id)
		h.Tick()
		if len(sizes) != 1 || second != 1 || screen.PresentCount() != beforePresent+1 {
			t.Fatal("same-size resize duplicated notification or redraw")
		}
	}
	fmt.Println("SDL RESIZE HEADLESS PASS: " + name)
}
