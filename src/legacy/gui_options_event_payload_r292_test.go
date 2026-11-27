package legacy

import (
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type inputApplyTestClient struct {
	widgetCallbackClient
	CtrlEventHandler
	resets int
}

func (c *inputApplyTestClient) GetCtrlEvent() CtrlEventHandler { return c }
func (c *inputApplyTestClient) Reset()                         { c.resets++ }

type inputApplyTestServer struct {
	Server
	world *server.Server
}

func (s *inputApplyTestServer) S() *server.Server { return s.world }

func TestInputConfigApplyUsesLiveSignedRowCountAndReturnsParserStatus(t *testing.T) {
	g := gui.NewHigh(nil)
	defer g.DestroyAll()
	list, free := alloc.NewHigh(gui.ScrollListBoxData{})
	defer free()
	world := server.NewHigh(nil, nil, strman.New())
	defer world.Close()
	cl := &inputApplyTestClient{widgetCallbackClient: widgetCallbackClient{cli: &client.Client{GUI: g}}}
	oldClient, oldServer := GetClient, GetServer
	oldGame, oldShell := gameInputBindings, shellInputBindings
	oldKey, oldEvent := Nox_xxx_keybind_nameByTitle_42E960, Nox_xxx_bindevent_bindNameByTitle_42EA40
	oldParse := Nox_client_parseConfigHotkeysLine_42CF50
	defer func() {
		GetClient, GetServer = oldClient, oldServer
		gameInputBindings, shellInputBindings = oldGame, oldShell
		Nox_xxx_keybind_nameByTitle_42E960, Nox_xxx_bindevent_bindNameByTitle_42EA40 = oldKey, oldEvent
		Nox_client_parseConfigHotkeysLine_42CF50 = oldParse
	}()
	GetClient = func() Client { return cl }
	GetServer = func() Server { return &inputApplyTestServer{world: world} }
	Nox_xxx_keybind_nameByTitle_42E960 = func(title string) keybind.Key {
		switch title {
		case "primary":
			return keybind.KeyA
		case "secondary":
			return keybind.KeyB
		default:
			return keybind.KeyEsc
		}
	}
	Nox_xxx_bindevent_bindNameByTitle_42EA40 = func(string) *keybind.BindEvent { return nil }

	var rows []string
	rowWindow := func(text string) *gui.Window {
		return g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 80, 30, func(_ *gui.Window, event gui.WindowEvent) gui.WindowEventResp {
			if event.EventCode() != 16406 {
				return nil
			}
			index, _ := event.EventArgsC()
			if index != 0 || len(rows) >= 3 {
				t.Errorf("read row %d of %s after %v", index, text, rows)
				list.Field_11_0 = 0
				return nil
			}
			rows = append(rows, text)
			return gui.RawEventResp(uintptr(unsafe.Pointer(internWStr(text))))
		})
	}
	labels := rowWindow("label")
	labels.WidgetData = unsafe.Pointer(list)
	defer func() { labels.WidgetData = nil }()
	main, secondary := rowWindow("primary"), rowWindow("secondary")
	for _, panel := range []*inputBindingPanel{&gameInputBindings, &shellInputBindings} {
		panel.labels = (*nox_window)(labels.C())
		panel.columns = [2]*nox_window{(*nox_window)(main.C()), (*nox_window)(secondary.C())}
	}
	primaryLine := keybind.KeyA.String() + " = (null)"
	secondaryLine := keybind.KeyB.String() + " = (null)"
	escapeLine := keybind.KeyEsc.String() + " = (null)"
	var parsed []string
	Nox_client_parseConfigHotkeysLine_42CF50 = func(line string) int {
		parsed = append(parsed, line)
		if line == primaryLine {
			list.Field_11_0 = 0
		}
		return -73
	}
	for _, shell := range []bool{false, true} {
		for _, count := range []uint16{0xffff, 2} {
			rows, parsed = nil, nil
			cl.resets = 0
			list.Field_11_0 = count
			var wantRows []string
			wantParsed := []string{escapeLine}
			if count == 2 {
				wantRows = []string{"label", "primary", "secondary"}
				wantParsed = []string{secondaryLine, primaryLine, escapeLine}
			}
			if got := inputConfigApplyBindingsC(shell); got != -73 || cl.resets != 1 ||
				!slices.Equal(rows, wantRows) || !slices.Equal(parsed, wantParsed) {
				t.Fatalf("shell=%t rows=%#x returned %d with %d resets, rows %v and parsed %v", shell, count, got, cl.resets, rows, parsed)
			}
		}
	}
}

func TestInputConfigRootCallbacksDecodeEventPayload(t *testing.T) {
	for _, shell := range []bool{false, true} {
		t.Run(map[bool]string{false: "in_game", true: "shell"}[shell], func(t *testing.T) {
			g := gui.NewHigh(nil)
			defer g.DestroyAll()
			root := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 640, 480, nil)
			sender := g.NewWindowRaw(root, gui.StatusEnabled, 10, 20, 120, 20, nil)
			requireNativeWindowAddress(t, root)
			requireNativeWindowAddress(t, sender)

			oldGame, oldShell := Sub_4C3A90, Sub_4CBE70
			t.Cleanup(func() { Sub_4C3A90, Sub_4CBE70 = oldGame, oldShell })
			cases := []struct {
				name    string
				event   int
				a3, a4  uintptr
				pointer unsafe.Pointer
			}{
				// This exact notification crashed during Host Load while the
				// InputCfg.wnd parser attached child ID 980 to its root.
				{name: "new_child_980", event: 22, a3: 980},
				{name: "focus", event: 23, a3: 1},
				{name: "key", event: 21, a3: 65, a4: 2},
				{name: "mouse_position", event: 17, a3: 0x005c0210},
				{name: "ignored_native_scalars", event: 0x7fff, a3: ^uintptr(0), a4: ^uintptr(0)},
				{name: "button_sender", event: 16391, a3: uintptr(sender.C()), a4: uintptr(root.C()), pointer: sender.C()},
				{name: "nil_button_sender", event: 16391},
			}
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					calls := 0
					hook := func(a1, event int, ptr unsafe.Pointer, a4 int) int {
						calls++
						if uintptr(a1) != uintptr(root.C()) || event != tc.event || ptr != tc.pointer || uintptr(a4) != tc.a4 {
							t.Fatalf("callback = (%#x, %d, %p, %#x), want (%p, %d, %p, %#x)",
								a1, event, ptr, a4, root.C(), tc.event, tc.pointer, tc.a4)
						}
						return -17
					}
					if shell {
						Sub_4CBE70 = hook
					} else {
						Sub_4C3A90 = hook
					}
					if got := inputConfigRootCallbackC(root, shell, tc.event, tc.a3, tc.a4); got != -17 {
						t.Fatalf("C-to-Go root callback response = %d, want -17", got)
					}
					if calls != 1 {
						t.Fatalf("root callback calls = %d, want 1", calls)
					}
				})
			}
		})
	}
}
