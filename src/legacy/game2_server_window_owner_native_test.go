package legacy

import (
	"bytes"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
	"github.com/opennox/opennox/v1/server"
)

func TestGame2ServerNativeRulesHead(t *testing.T) {
	head := cfixture.Game2RulesHead()
	if head != unsafe.Pointer(serverOptionsRules) {
		t.Fatalf("rules head = %p, want %p", head, serverOptionsRules)
	}
	runtime.GC()
	if got := cfixture.Game2RulesHead(); got != head {
		t.Fatalf("rules head after GC = %p, want %p", got, head)
	}
	for i, link := range unsafe.Slice((*unsafe.Pointer)(head), 3) {
		if link != head {
			t.Fatalf("empty rules head field %d = %p, want %p", i, link, head)
		}
	}
}

func TestGame2ServerNativeOptionsWindowGetter(t *testing.T) {
	// Fixture field8 is the actual C-owned nox_window* global. The getter
	// must return that pointer, rather than reading its low DWORD as an int.
	old := game2ServerFixtureOwner(8)
	var allocations [2]unsafe.Pointer
	defer func() {
		game2ServerFixtureSetOwner(8, old)
		for _, p := range allocations {
			if p != nil {
				game2ServerFixtureFree(p)
			}
		}
	}()

	windowBytes := game2ServerFixtureLayout(0)
	var windows [2]unsafe.Pointer
	var before [2][]byte
	for i := range allocations {
		p := game2ServerFixtureAlloc(windowBytes + 32)
		if p == nil {
			t.Fatal("C window allocation")
		}
		allocations[i] = p
		windows[i] = unsafe.Add(p, 16)
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(windows[i]) <= 0xFFFFFFFF {
			t.Fatal("C window must exercise address bits above the low DWORD")
		}
		storage := unsafe.Slice((*byte)(p), int(windowBytes)+32)
		for j := range storage {
			storage[j] = byte(0x39 + i*0x21 + j*7)
		}
		before[i] = append([]byte(nil), storage...)
	}
	if windows[0] == windows[1] {
		t.Fatal("independent C window allocations alias")
	}

	for _, want := range []unsafe.Pointer{nil, windows[0], windows[1], windows[0], nil} {
		game2ServerFixtureSetOwner(8, want)
		for gc := 0; gc < 3; gc++ {
			runtime.GC()
			var got unsafe.Pointer = Get_dword_5d4594_1046492()
			if got != want || game2ServerFixtureOwner(8) != want {
				t.Fatalf("public getter lost live C window identity: got %p want %p", got, want)
			}
			for i, p := range allocations {
				if !bytes.Equal(unsafe.Slice((*byte)(p), int(windowBytes)+32), before[i]) {
					t.Fatal("getter changed a complete C window or its surrounding canaries")
				}
			}
		}
	}
}

func TestGame2ServerApply459AA0SignedSelection(t *testing.T) {
	g := newWidgetCallbackGUI(t)
	defer g.DestroyAll()
	oldRoot, oldName, oldChildren := serverSharedWindows.options, serverOptionsNameInput, serverOptionsChildren
	defer func() {
		serverSharedWindows.options, serverOptionsNameInput, serverOptionsChildren = oldRoot, oldName, oldChildren
	}()

	root := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 10, 10, nil)
	serverSharedWindows.options = (*nox_window)(root.C())
	check := g.NewWindowRaw(root, gui.StatusEnabled, 0, 0, 10, 10, nil)
	check.SetID(10122)
	(*nox_window)(check.C()).draw_data.field_0 = 4

	wideWord := func(text []uint16) gui.RawEventResp {
		buf, free := alloc.MakeHigh(text, len(text))
		t.Cleanup(free)
		return gui.RawEventResp(uintptr(unsafe.Pointer(&buf[0])))
	}
	input := func(text []uint16) *nox_window {
		value := wideWord(text)
		win := g.NewWindowRaw(root, gui.StatusEnabled, 0, 0, 10, 10, func(_ *gui.Window, event gui.WindowEvent) gui.WindowEventResp {
			if event.EventCode() == 16413 {
				return value
			}
			return nil
		})
		return (*nox_window)(win.C())
	}
	serverOptionsNameInput = input([]uint16{'H', 'o', 's', 't', 0})
	serverOptionsChildren.lessonLimit = input([]uint16{'1', '2', 0})
	serverOptionsChildren.timeLimit = input([]uint16{0})
	mapText := wideWord([]uint16{'c', 'a', 's', 't', 'l', 'e', '\t', 'x', 0})
	noToken := wideWord([]uint16{'\t', '\t', 0})

	var selection, textCalls, requested uintptr
	var line gui.RawEventResp
	maps := g.NewWindowRaw(root, gui.StatusEnabled, 0, 0, 10, 10, func(_ *gui.Window, event gui.WindowEvent) gui.WindowEventResp {
		switch event.EventCode() {
		case 16404:
			return gui.RawEventResp(selection)
		case 16406:
			textCalls++
			requested, _ = event.EventArgsC()
			return line
		}
		return nil
	})
	serverOptionsChildren.mapList = (*nox_window)(maps.C())

	for _, route := range []struct {
		name  string
		apply func(unsafe.Pointer) int32
	}{{"Go", game2ServerApply459AA0}, {"C", cfixture.Game2ServerApply459AA0}} {
		t.Run(route.name, func(t *testing.T) {
			for _, tc := range []struct {
				name    string
				word    uint32
				text    gui.RawEventResp
				want    int32
				mapName string
			}{
				{"missing", 0xffffffff, mapText, -1, ""},
				{"minimum", 0x80000000, mapText, -2147483648, ""},
				{"selected", 4, mapText, 0, "castle"},
				{"empty token", 0, noToken, 0, ""},
			} {
				t.Run(tc.name, func(t *testing.T) {
					selection, line = uintptr(tc.word), tc.text
					textCalls, requested = 0, 0
					record := server.Settings2{Field56: 22}
					copy(record.Field0[:], "stale")
					got := route.apply(unsafe.Pointer(&record))
					if got != tc.want {
						t.Fatalf("selection %#x returned %d, want %d", tc.word, got, tc.want)
					}
					if tc.want < 0 && textCalls != 0 || tc.want >= 0 && (textCalls != 1 || requested != selection) {
						t.Fatalf("selection %#x requested row %#x %d times", tc.word, requested, textCalls)
					}
					name, _, _ := bytes.Cut(record.Field0[:9], []byte{0})
					if string(name) != tc.mapName || string(record.Field0[9:14]) != "Host\x00" ||
						record.Field54 != 12 || record.Field56 != 22 || record.Field57 != 1 {
						t.Fatalf("applied settings = %+v", record)
					}
				})
			}
		})
	}
}
