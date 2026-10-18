//go:build opennox_sdl_input_test

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
	"strings"
	"testing"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/noximage"
	"github.com/opennox/opennox/v1/client/input"
	"github.com/opennox/opennox/v1/client/render"
	"github.com/opennox/opennox/v1/client/seat/headless"
	native "github.com/veandco/go-sdl2/sdl"
)

const inputNativeChild = "OPENNOX_TEST_SDL_INPUT_HEADLESS_R312"

var inputNativeCases = []string{
	"mapped-key", "unmapped-prefix", "unmapped-release", "out-of-range",
	"unmapped-middle",
}

// Exercise the real SDL dummy driver/event queue, production PollEvent adapter,
// renderer viewport, and game input handler. There is no physical window,
// OpenGL context or compositor; this test does not certify their output.
func TestSDLInputNativeHeadless(t *testing.T) {
	if name := os.Getenv(inputNativeChild); name != "" {
		testSDLInputNativeHeadless(t, name)
		return
	}
	for _, profile := range []string{"menu", "game-letterbox"} {
		for _, name := range inputNativeCases {
			t.Run(profile+"/"+name, func(t *testing.T) {
				name := profile + "/" + name
				cmd := exec.Command(os.Args[0], "-test.run=^TestSDLInputNativeHeadless$", "-test.v")
				cmd.Env = append(os.Environ(), inputNativeChild+"="+name, "SDL_VIDEODRIVER=dummy")
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("native SDL input failed: %v\n%s", err, out)
				}
				if !bytes.Contains(out, []byte("SDL INPUT HEADLESS PASS: "+name)) {
					t.Fatalf("missing native queue/input receipt: %s", out)
				}
				t.Logf("SDL INPUT HEADLESS PASS: %s", name)
			})
		}
	}
}

// sdl2-compat can normalize an invalid scancode while translating its queue.
// Cover the adapter boundary directly as well so its own bounds check is
// established independently of the installed SDL implementation.
func TestSDLKeyboardScancodeBoundary(t *testing.T) {
	for _, code := range []native.Scancode{native.Scancode(^uint32(0)), 1 << 30, native.SCANCODE_UNKNOWN, native.SCANCODE_LANG1, native.SCANCODE_LANG2} {
		t.Run(fmt.Sprint(int(code)), func(t *testing.T) {
			w := &Window{}
			var keys []keybind.Key
			w.OnInput(func(ev seat.InputEvent) {
				if ev, ok := ev.(*seat.KeyboardEvent); ok {
					keys = append(keys, ev.Key)
				}
			})
			defer func() {
				if v := recover(); v != nil {
					t.Fatalf("unsupported SDL scancode panicked: %v", v)
				}
			}()
			w.processKeyboardEvent(&native.KeyboardEvent{State: native.PRESSED, Keysym: native.Keysym{Scancode: code}})
			w.processKeyboardEvent(&native.KeyboardEvent{State: native.PRESSED, Keysym: native.Keysym{Scancode: native.SCANCODE_A}})
			if !slices.Equal(keys, []keybind.Key{keybind.KeyA}) {
				t.Fatalf("unsupported scancode became a game key: %v", keys)
			}
		})
	}
}

type inputHeadlessSeat struct {
	*Window
	output *headless.Seat
}

func (s *inputHeadlessSeat) NewSurface(sz image.Point, filtering bool) seat.Surface {
	return s.output.NewSurface(sz, filtering)
}
func (s *inputHeadlessSeat) Clear()   { s.output.Clear() }
func (s *inputHeadlessSeat) Present() { s.output.Present() }

func testSDLInputNativeHeadless(t *testing.T, fullName string) {
	profile, name, found := strings.Cut(fullName, "/")
	if !found || !slices.Contains(inputNativeCases, name) || (profile != "menu" && profile != "game-letterbox") {
		t.Fatal("unknown SDL native input case")
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
	size, canvas := image.Pt(640, 480), image.Pt(640, 480)
	if profile == "game-letterbox" {
		size, canvas = image.Pt(800, 450), image.Pt(1024, 768)
	}
	win, err := native.CreateWindow("input-native-headless", 0, 0, int32(size.X), int32(size.Y), native.WINDOW_HIDDEN)
	if err != nil {
		t.Fatal(err)
	}
	defer win.Destroy()
	if win.GetFlags()&native.WINDOW_OPENGL != 0 {
		t.Fatal("OpenGL window must never be created")
	}
	id, err := win.GetID()
	if err != nil {
		t.Fatal(err)
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	w := &Window{win: win, log: log}
	w.InputTick() // drain native creation events before observing
	screen := headless.New(size)
	defer screen.Close()
	r, err := render.New(&inputHeadlessSeat{Window: w, output: screen})
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	pix := noximage.NewImage16(image.Rectangle{Max: canvas})
	pix.Pix[0], pix.Pix[len(pix.Pix)-1] = 0x1f, 0x7c00
	r.CopyBuffer(pix)
	h := input.New(log, w, false, 0)
	h.SetDrawWinSize(canvas)
	r.OnViewResize(h.SetWinSize)
	h.Tick()
	var keys []keybind.Key
	h.OnKeyPress(func(key keybind.Key) { keys = append(keys, key) })
	var emittedKeys []keybind.Key
	w.OnInput(func(ev seat.InputEvent) {
		if ev, ok := ev.(*seat.KeyboardEvent); ok {
			emittedKeys = append(emittedKeys, ev.Key)
		}
	})
	push := func(got int) {
		t.Helper()
		if got != 1 {
			t.Fatalf("real SDL queue rejected input: %d", got)
		}
	}
	key := func(scancode native.Scancode, down bool) {
		t.Helper()
		push(inputNativePushKey(id, int(scancode), down))
	}
	switch name {
	case "mapped-key":
		key(native.SCANCODE_A, true)
		h.Tick()
		if !h.IsPressed(keybind.KeyA) || !slices.Equal(keys, []keybind.Key{keybind.KeyA}) {
			t.Fatal("native keydown did not reach the game handler")
		}
		key(native.SCANCODE_A, false)
		h.Tick()
		if h.IsPressed(keybind.KeyA) || !h.IsReleased(keybind.KeyA) {
			t.Fatal("native keyup did not release the game key")
		}
	case "unmapped-prefix":
		key(native.SCANCODE_LANG1, true) // ordinary Korean Hangul/English key, no DirectInput mapping
		key(native.SCANCODE_A, true)
		h.Tick()
		if !h.IsPressed(keybind.KeyA) || !slices.Equal(keys, []keybind.Key{keybind.KeyA}) || !slices.Equal(emittedKeys, []keybind.Key{keybind.KeyA}) {
			t.Fatalf("unmapped SDL key swallowed the following mapped key: pressed=%v keys=%v emitted=%v", h.IsPressed(keybind.KeyA), keys, emittedKeys)
		}
	case "unmapped-release":
		key(native.SCANCODE_A, true)
		h.Tick()
		key(native.SCANCODE_LANG1, false)
		key(native.SCANCODE_A, false)
		key(native.SCANCODE_B, true)
		h.Tick()
		if h.IsPressed(keybind.KeyA) || !h.IsPressed(keybind.KeyB) || slices.Contains(emittedKeys, 0) {
			t.Fatalf("unmapped key prevented release or the next press: A=%v B=%v emitted=%v", h.IsPressed(keybind.KeyA), h.IsPressed(keybind.KeyB), emittedKeys)
		}
	case "out-of-range":
		push(inputNativePushKey(id, -1, true))
		push(inputNativePushKey(id, int(native.NUM_SCANCODES), true))
		key(native.SCANCODE_A, true)
		h.Tick()
		if !h.IsPressed(keybind.KeyA) || !slices.Equal(emittedKeys, []keybind.Key{keybind.KeyA}) {
			t.Fatal("out-of-range scancodes interfered with valid input")
		}
	case "unmapped-middle":
		key(native.SCANCODE_A, true)
		key(native.SCANCODE_LANG2, true) // ordinary Korean Hanja key, also unmapped
		key(native.SCANCODE_B, true)
		h.Tick()
		if !h.IsPressed(keybind.KeyA) || !h.IsPressed(keybind.KeyB) || !slices.Equal(emittedKeys, []keybind.Key{keybind.KeyA, keybind.KeyB}) {
			t.Fatalf("unmapped key truncated the mapped batch: A=%v B=%v emitted=%v", h.IsPressed(keybind.KeyA), h.IsPressed(keybind.KeyB), emittedKeys)
		}
	}
	shown, _ := screen.Snapshot()
	if shown == nil || !slices.Equal(shown.Pix, pix.Pix) {
		t.Fatal("native input changed or lost the retained game frame")
	}
	fmt.Println("SDL INPUT HEADLESS PASS: " + fullName)
}
