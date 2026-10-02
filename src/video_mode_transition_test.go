//go:build !server

package opennox

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/env"
	"github.com/opennox/libs/noxfont"
	"github.com/opennox/libs/strman"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/client/input"
	"github.com/opennox/opennox/v1/client/render"
	"github.com/opennox/opennox/v1/client/seat/headless"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

const videoModeTransitionChild = "NOX_TEST_VIDEO_MODE_TRANSITION_CHILD"

// This is a non-E2E video/API integration, not a stock GUI/game-loop playback.
// It uses the actual C apply entry, gameResetVideoMode, render initialization,
// fonts, C-owned pixbuffer/row table, upload and queued headless mouse events.
// Only the seat and font assets are fixtures. No production video callback is
// replaced; no stock/personal data, config or Save directory is accessed.
func TestVideoModeTransitionsHeadless(t *testing.T) {
	runVideoModeTransitionChild(t, "frames")
}

func TestVideoModePersistentGUIFontsHeadless(t *testing.T) {
	runVideoModeTransitionChild(t, "fonts")
}

func runVideoModeTransitionChild(t *testing.T, phase string) {
	t.Helper()
	dir := t.TempDir()
	for _, name := range []string{"default", "large", "number", "small"} {
		if err := os.WriteFile(filepath.Join(dir, name+".ttf"), goregular.TTF, 0600); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestVideoModeTransitionsHeadlessChild$", "-test.count=1", "-test.v")
	cmd.Dir = dir
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if key == videoModeTransitionChild || key == "NOX_DATA" || strings.HasPrefix(key, "NOX_E2E") {
			continue
		}
		cmd.Env = append(cmd.Env, kv)
	}
	cmd.Env = append(cmd.Env, videoModeTransitionChild+"="+phase)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("video transition subprocess failed: %v\n%s", err, out)
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "HEADLESS VIDEO:") {
			t.Log(strings.TrimSpace(line))
		}
	}
}

func TestVideoModeTransitionsHeadlessChild(t *testing.T) {
	phase := os.Getenv(videoModeTransitionChild)
	if phase == "" {
		t.Skip("only run in an isolated headless video subprocess")
	}
	if env.IsE2E() {
		t.Fatal("E2E suppression must not mask actual video mode changes")
	}
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	datapath.SetData(dir)
	legacy.InitBlobData()
	handles.Init()
	defer handles.Release()
	s := server.New(nil, nil, strman.New())
	defer s.Close()
	noxServer = &Server{Server: s}
	c, err := NewClient(nil, nil, noxServer)
	if err != nil {
		t.Fatal(err)
	}
	noxClient = c
	defer c.Client.Close()
	defer c.GUI.DestroyAll()
	menu := image.Pt(640, 480)
	sc := headless.New(menu)
	defer sc.Close()
	c.Seat = sc
	c.Win, err = render.New(sc)
	if err != nil {
		t.Fatal(err)
	}
	c.Inp = input.New(c.Log, sc, false, c.Strings().Lang())
	c.GUI.SetInput(c.Inp)
	// The same viewport/canvas hooks as initSeat, without SDL/Retina units.
	// Observers below never repair input coordinates or resize the buffers.
	c.Win.OnViewResize(c.Inp.SetWinSize)
	OnPixBufferResize(c.Inp.SetDrawWinSize)
	resizes := 0
	OnPixBufferResize(func(image.Point) { resizes++ })
	c.videoSetGameMode(menu)
	if err := c.gameResetVideoMode(true, true); err != nil {
		t.Fatal(err)
	}
	defer c.nox_video_freeFloorBuffer_430EC0()
	if phase == "fonts" {
		verifyHeadlessPersistentGUIFonts(t, c, sc)
		return
	}
	if phase != "frames" {
		t.Fatalf("unknown video test phase %q", phase)
	}
	sentinel := []byte("isolated video test: do not write configuration\n")
	for _, name := range []string{"nox.cfg", "opennox.yml"} {
		if err := os.WriteFile(name, sentinel, 0600); err != nil {
			t.Fatal(err)
		}
	}
	steps, highRows := 0, 0
	verify := func(label string, want image.Point) {
		t.Helper()
		steps++
		if verifyHeadlessVideoFrame(t, c, sc, label, want) {
			highRows++
		}
	}
	verify("initial menu", menu)
	for _, windowMode := range []int{-3, -1, -2} {
		c.Win.SetWindowMode(windowMode)
		for _, size := range getResolutionOptions() {
			if size == (image.Point{}) {
				continue
			}
			label := fmt.Sprintf("mode %d size %v", windowMode, size)
			before, beforeResize := noxPixBuffer.img, resizes
			guiOptionsRes = size
			ccall.CallVoidVoid(legacy.OptionsApplyVideoModeCEntry)
			if c.videoGetGameMode() != size || sc.ScreenSize() != size || g_fullscreen_cfg != windowMode {
				t.Fatalf("%s: C apply did not update game/seat/checkpoint", label)
			}
			// Shell options stage the game resolution; the menu canvas remains
			// 640x480 until the existing game-entry reset is explicitly called.
			if noxPixBuffer.img != before || resizes != beforeResize || before.Size() != menu {
				t.Fatalf("%s: staged option unexpectedly recreated the menu canvas", label)
			}
			if err := gameUpdateVideoMode(false); err != nil {
				t.Fatal(err)
			}
			wantResizes := beforeResize
			if size != menu {
				wantResizes++
			}
			if resizes != wantResizes {
				t.Fatalf("%s: resize callbacks = %d, want %d", label, resizes, wantResizes)
			}
			verify(label+" game entry", size)
			before, beforeResize = noxPixBuffer.img, resizes
			if err := gameUpdateVideoMode(false); err != nil {
				t.Fatal(err)
			}
			if noxPixBuffer.img != before || resizes != beforeResize {
				t.Fatalf("%s: unchanged mode recreated the pixel buffer", label)
			}
			verify(label+" unchanged", size)
			// A window resize presents the previous canvas, not a new game
			// resolution. Its letterboxed input must continue to address it.
			sc.ResizeScreen(size.Add(image.Pt(200, 100)))
			if noxPixBuffer.img != before || resizes != beforeResize {
				t.Fatalf("%s: seat resize recreated the game canvas", label)
			}
			verify(label+" window resize", size)
			seatSize := sc.ScreenSize()
			if err := gameUpdateVideoMode(true); err != nil {
				t.Fatal(err)
			}
			if c.videoGetGameMode() != size || sc.ScreenSize() != seatSize {
				t.Fatalf("%s: menu return discarded the selected game/seat mode", label)
			}
			verify(label+" menu return", menu)
		}
	}
	beforeResize := resizes
	if err := c.gameResetVideoMode(true, true); err != nil {
		t.Fatal(err)
	}
	if resizes != beforeResize+1 {
		t.Fatal("forced same-size reset did not recreate the canvas")
	}
	verify("forced menu reset", menu)
	for _, name := range []string{"nox.cfg", "opennox.yml"} {
		got, err := os.ReadFile(name)
		if err != nil || !bytes.Equal(got, sentinel) {
			t.Fatalf("video transitions changed %s: %v", name, err)
		}
	}
	c.nox_video_freeFloorBuffer_430EC0()
	if noxPixBuffer.img != nil || noxPixBuffer.rows != nil || *legacy.VideoModeRowsC != nil || legacy.Get_dword_5d4594_823776() != 0 {
		t.Fatal("video cleanup left live pixel/row/initialized state")
	}
	t.Logf("HEADLESS VIDEO: %d frame checks, %d resize callbacks, %d native row tables above 4 GiB; 3 signed window modes; configs unchanged", steps, resizes, highRows)
}

func verifyHeadlessPersistentGUIFonts(t *testing.T, c *Client, sc *headless.Seat) {
	t.Helper()
	type label struct {
		name    string
		win     *gui.Window
		ptr     unsafe.Pointer
		metrics font.Metrics
		pixels  []uint16
	}
	var labels []label
	for i, name := range []string{noxfont.DefaultName, noxfont.LargeName, noxfont.NumbersName, noxfont.SmallName} {
		draw, free := gui.NewWindowData()
		draw.Style = gui.StyleStaticText
		draw.SetBackgroundColor(color.Transparent)
		draw.SetEnabledColor(color.Transparent)
		draw.SetDisabledColor(color.Transparent)
		draw.SetTextColor(color.White)
		draw.SetFont(c.r.Fonts.FontPtrByName(name))
		win := c.GUI.NewStaticTextRaw(nil, gui.StatusEnabled|gui.StatusNoFocus, 12, 12+70*i, 320, 48, draw, &gui.StaticTextData{
			Text: alloc.InternCString16("Persistent GUI font\nVideo mode transition"),
		})
		free()
		if win == nil || win.DrawData().Font() == nil {
			t.Fatalf("could not initialize persistent %s font label", name)
		}
		labels = append(labels, label{name: name, win: win, ptr: win.DrawData().FontC(), metrics: win.DrawData().Font().Metrics()})
	}
	// The windows, their native FontPtr fields and all expected pixels are
	// retained across real resets. Only ordinary Draw uploads the text.
	capture := func(win *gui.Window) []uint16 {
		c.r.ClearScreen(color.Black)
		win.Draw()
		c.copyPixBuffer()
		frame, _ := sc.Snapshot()
		var pixels []uint16
		pos, size := win.GlobalPos(), win.Size()
		for y := pos.Y; y < pos.Y+size.Y; y++ {
			row := frame.Pix[y*frame.Stride+pos.X : y*frame.Stride+pos.X+size.X]
			pixels = append(pixels, row...)
		}
		return pixels
	}
	for i := range labels {
		labels[i].pixels = capture(labels[i].win)
		if !slices.ContainsFunc(labels[i].pixels, func(p uint16) bool { return p != 0 }) {
			t.Fatalf("initial %s label rendered no pixels", labels[i].name)
		}
	}
	checks := 0
	for _, step := range []struct {
		size  image.Point
		menu  bool
		force bool
	}{
		{size: image.Pt(800, 600)},
		{size: image.Pt(1024, 768)},
		{menu: true},
		{menu: true, force: true},
	} {
		if !step.menu {
			guiOptionsRes = step.size
			ccall.CallVoidVoid(legacy.OptionsApplyVideoModeCEntry)
		}
		if err := c.gameResetVideoMode(step.menu, step.force); err != nil {
			t.Fatal(err)
		}
		for _, label := range labels {
			checks++
			draw := label.win.DrawData()
			if draw.FontC() != label.ptr {
				t.Errorf("%s: reset rewrote a live window's font field", label.name)
			}
			fnt := draw.Font()
			if fnt == nil || fnt.Metrics() != label.metrics {
				t.Errorf("%s: video reset lost the persistent GUI font binding", label.name)
			}
			if got := capture(label.win); !slices.Equal(got, label.pixels) {
				t.Errorf("%s: video reset changed the actual persistent label pixels", label.name)
			}
		}
	}
	t.Logf("HEADLESS VIDEO: %d retained GUI font/pixel checks across 4 actual resets; window font fields unchanged", checks)
}

func verifyHeadlessVideoFrame(t *testing.T, c *Client, sc *headless.Seat, label string, want image.Point) bool {
	t.Helper()
	pix := noxPixBuffer.img
	if pix == nil || pix.Size() != want || pix.Stride != want.X || len(pix.Pix) != want.X*want.Y || c.r.PixBuffer() != pix {
		t.Fatalf("%s: wrong live pixel buffer for %v", label, want)
	}
	if len(noxPixBuffer.rows) != want.Y || *legacy.VideoModeRowsC != &noxPixBuffer.rows[0] || *legacy.VideoModeRenderDataC != c.r.Data() {
		t.Fatalf("%s: C row/render-data globals are stale", label)
	}
	crows := unsafe.Slice(*legacy.VideoModeRowsC, want.Y)
	for y := 0; y < want.Y; y++ {
		if crows[y] != &pix.Pix[y*want.X] || noxPixBuffer.rows[y] != crows[y] {
			t.Fatalf("%s: native row %d has the wrong width/pointer", label, y)
		}
	}
	if pitch := ccall.CallIntVoid(legacy.VideoModePitchCEntry); pitch != 2*want.X {
		t.Fatalf("%s: actual C backbuffer pitch = %d, want %d", label, pitch, 2*want.X)
	}
	d := c.r.Data()
	if d.Rect3() != pix.Rect || d.ClipRect() != pix.Rect || d.ClipRect2() != image.Rect(0, 0, want.X-1, want.Y-1) {
		t.Fatalf("%s: stale full-screen clipping: %v / %v / %v", label, d.Rect3(), d.ClipRect(), d.ClipRect2())
	}
	c.r.ClearScreen(color.Black)
	c.r.DrawRectFilledOpaque(0, 0, want.X/2, want.Y/2, color.RGBA{R: 255, A: 255})
	c.r.DrawRectFilledOpaque(want.X/2, want.Y/2, want.X-want.X/2, want.Y-want.Y/2, color.RGBA{B: 255, A: 255})
	if pix.Pix[0] == 0 || pix.Pix[len(pix.Pix)-1] == 0 || pix.Pix[0] == pix.Pix[len(pix.Pix)-1] {
		t.Fatalf("%s: resized canvas did not render distinct first/last pixels", label)
	}
	d.SetTextColor(color.White)
	font := c.r.Fonts.DefaultFont()
	if font == nil || c.r.DrawString(font, "Video transition", image.Pt(8, want.Y/2)) <= 0 {
		t.Fatalf("%s: reloaded font could not draw", label)
	}
	white := 0
	for y := want.Y / 2; y < want.Y/2+32; y++ {
		for x := 8; x < 160; x++ {
			if pix.Pix[y*want.X+x] != 0 {
				white++
			}
		}
	}
	if white == 0 {
		t.Fatalf("%s: reloaded font produced no pixels", label)
	}
	beforePresents := sc.PresentCount()
	c.copyPixBuffer()
	frame, view := sc.Snapshot()
	if sc.PresentCount() != beforePresents+1 || frame == nil || frame.Rect != pix.Rect || frame.Stride != pix.Stride || !slices.Equal(frame.Pix, pix.Pix) {
		t.Fatalf("%s: headless upload/presentation did not preserve the frame", label)
	}
	if view.Empty() || !view.In(image.Rectangle{Max: sc.ScreenSize()}) {
		t.Fatalf("%s: invalid presentation viewport %v for %v", label, view, sc.ScreenSize())
	}
	center := view.Min.Add(image.Pt(view.Dx()/2, view.Dy()/2))
	for _, tc := range []struct{ raw, want image.Point }{
		{view.Min, image.Point{}},
		{center, image.Pt(want.X/2, want.Y/2)},
		{view.Max.Sub(image.Pt(1, 1)), want.Sub(image.Pt(1, 1))},
	} {
		sc.QueueInput(&seat.MouseMoveEvent{Pos: tc.raw})
		c.Inp.Tick()
		got := c.Inp.GetMousePos()
		// Downscaling and float32 viewport rounding can skip one canvas
		// pixel; a stale 640x480 scale/bounds is never within this tolerance.
		if abs(got.X-tc.want.X) > 1 || abs(got.Y-tc.want.Y) > 1 {
			t.Fatalf("%s: queued mouse %v mapped to %v, want near %v", label, tc.raw, got, tc.want)
		}
	}
	return uint64(uintptr(unsafe.Pointer(*legacy.VideoModeRowsC))) > 0xffffffff
}
