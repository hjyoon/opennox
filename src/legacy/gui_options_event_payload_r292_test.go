package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
)

func TestInputConfigRootCallbacksDecodeEventPayload(t *testing.T) {
	for _, shell := range []bool{false, true} {
		t.Run(map[bool]string{false: "in_game", true: "shell"}[shell], func(t *testing.T) {
			g := gui.New(nil)
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
