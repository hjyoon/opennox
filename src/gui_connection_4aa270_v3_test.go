package opennox

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"image"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

// Constants and ordering below come from identity-checked GAME.EXE
// 4AA270..4AA64C, 43AF50..43AF76 and stock window/ArnaMain.wnd.
// These tests execute real native-width windows/animations/C callbacks;
// only external GUI, audio and network services are observed by fixtures.
type connectionFixture4AA270 struct {
	g        *gui.GUI
	root     *gui.Window
	services connectionServices4AA270
	request  uint32
	solo     bool
	events   []string
	sounds   []sound.ID
	wol      int
	back     int
	next     int
}

func newConnectionFixture4AA270(t *testing.T) *connectionFixture4AA270 {
	t.Helper()
	oldOwner, oldBG := winConnection4AA270, gui.MainBg
	oldAnim, oldSpeed := gui.AnimGlobalState(), gui.AnimSpeed
	oldFlags := noxflags.GetGame()
	oldMode, oldWOL := connectionMode43AF50, legacy.Get_dword_5d4594_2650652()
	oldNext := legacy.NoxGameShowGameSelNative
	if oldOwner != nil || gui.FindAnimForStateID(connectionState4AA270) != nil {
		t.Fatal("connection fixture must not replace a live owner")
	}
	noxflags.UnsetGame(oldFlags)
	gui.AnimSpeed = 1
	f := &connectionFixture4AA270{g: gui.New(nil)}
	f.root = f.g.NewWindowRaw(nil, gui.StatusEnabled|gui.StatusNoFocus, 0, 0, 640, 480, nil)
	f.root.SetID(400)
	top := f.g.NewWindowRaw(f.root, gui.StatusEnabled|gui.StatusNoFocus, 0, 0, 640, 240, nil)
	top.SetID(410)
	f.g.NewWindowRaw(top, gui.StatusEnabled|gui.StatusNoFocus, 237, 182, 166, 35, nil).SetID(411)
	bottom := f.g.NewWindowRaw(f.root, gui.StatusEnabled, 0, 240, 640, 240, nil)
	bottom.SetID(420)
	f.g.NewWindowRaw(bottom, gui.StatusEnabled|gui.StatusNoFocus, 237, 16, 166, 35, nil).SetID(421)
	gui.MainBg = f.g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 640, 480, func(_ *gui.Window, ev gui.WindowEvent) gui.WindowEventResp {
		if _, ok := ev.(gui.WindowFocus); ok {
			return gui.RawEventResp(1)
		}
		return nil
	})
	f.services = connectionServices4AA270{
		closeDialog:   func() { f.events = append(f.events, "close-dialog") },
		pushState:     func(id gui.StateID) { f.events = append(f.events, fmt.Sprintf("state:%d", id)) },
		enableOptions: func(v int) { f.events = append(f.events, fmt.Sprintf("options:%d", v)) },
		setMode: func(v int) {
			f.events = append(f.events, fmt.Sprintf("mode:%d", v))
			connectionSetMode43AF50(v)
		},
		newWindow:    func() *gui.Window { f.events = append(f.events, "window"); return f.root },
		setBackLabel: func() { f.events = append(f.events, "back-label") },
		back:         func() { f.back++ },
		sharedProc: func(_ *gui.Window, _ gui.WindowEvent) gui.WindowEventResp {
			return gui.RawEventResp(123)
		},
		flags: noxflags.GetGame,
		solo:  func() bool { return f.solo },
		request: func() uint32 {
			return f.request
		},
		playSound: func(id sound.ID, volume int) {
			f.sounds = append(f.sounds, id)
			if volume != 100 {
				t.Errorf("volume=%d, want 100", volume)
			}
		},
		wolUnavailable: func(root *gui.Window) {
			f.wol++
			if root != f.root {
				t.Errorf("WOL dialog owner=%p, want %p", root, f.root)
			}
		},
	}
	legacy.NoxGameShowGameSelNative = func() int {
		f.next++
		if winConnection4AA270 != nil || gui.FindAnimForStateID(connectionState4AA270) != nil || !f.root.Flags.Has(gui.StatusDestroyed) {
			t.Error("next C callback ran before both animations and root were released")
		}
		return 1
	}
	t.Cleanup(func() {
		if winConnection4AA270 != nil {
			winConnection4AA270.free()
			winConnection4AA270 = nil
		}
		f.g.DestroyAll()
		legacy.NoxGameShowGameSelNative = oldNext
		connectionSetMode43AF50(int(oldMode))
		legacy.Set_dword_5d4594_2650652(oldWOL)
		noxflags.UnsetGame(noxflags.GetGame())
		noxflags.SetGame(oldFlags)
		gui.SetAnimGlobalState(oldAnim)
		gui.AnimSpeed = oldSpeed
		gui.MainBg, winConnection4AA270 = oldBG, oldOwner
	})
	return f
}

func TestConnectionSelector4AA270MainMenuRoute(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "gui_main_menu.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var body *ast.BlockStmt
	for _, d := range f.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok && fn.Name.Name == "nox_xxx_windowMainMenuProc_4A1DC0" {
			body = fn.Body
		}
	}
	if body == nil {
		t.Fatal("main-menu callback absent")
	}
	for _, tc := range []struct{ name, id, want string }{
		{"Solo Quest retains original connection stage", "131", "GetNoxGameShowConnection4AA270"},
		{"regular multiplayer entry not changed in Solo Quest port", "112", "Get_nox_game_showGameSel_4379F0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := ""
			ast.Inspect(body, func(n ast.Node) bool {
				clause, ok := n.(*ast.CaseClause)
				if !ok || len(clause.List) != 1 {
					return true
				}
				id, ok := clause.List[0].(*ast.BasicLit)
				if !ok || id.Value != tc.id {
					return true
				}
				ast.Inspect(clause, func(n ast.Node) bool {
					a, ok := n.(*ast.AssignStmt)
					if !ok || len(a.Lhs) != 1 || len(a.Rhs) != 1 {
						return true
					}
					lhs, ok := a.Lhs[0].(*ast.SelectorExpr)
					if !ok || lhs.Sel.Name != "Func13Ptr" {
						return true
					}
					call, ok := a.Rhs[0].(*ast.CallExpr)
					if !ok {
						return true
					}
					if target, ok := call.Fun.(*ast.SelectorExpr); ok {
						got = target.Sel.Name
					}
					return true
				})
				return false
			})
			if got != tc.want {
				t.Errorf("next callback=%s, want %s", got, tc.want)
			}
		})
	}
}

func TestConnectionSelector4AA270Construction(t *testing.T) {
	f := newConnectionFixture4AA270(t)
	if showConnection4AA270(f.services) != 1 {
		t.Fatal("entry failed")
	}
	if want := []string{"close-dialog", "state:400", "options:1", "mode:2", "window", "back-label"}; !reflect.DeepEqual(f.events, want) {
		t.Fatalf("service order=%v, want %v", f.events, want)
	}
	st := winConnection4AA270
	if st == nil || st.root != f.root || st.top.StateID != gui.StateID(400) || st.bottom.StateID != 0 {
		t.Fatal("original native owner/state IDs not retained")
	}
	if st.top.Window().Offs() != image.Pt(0, -240) || st.bottom.Window().Offs() != image.Pt(0, 480) || st.top.State() != gui.AnimIn || st.bottom.State() != gui.AnimIn {
		t.Fatal("original off-screen positions and slide-in states not retained")
	}
	if st.top.Func12Ptr != legacy.GetConnectionOutStart4AA450() || st.top.FncDoneOutPtr != legacy.GetConnectionOutDone4AA490() || st.top.Func13Ptr != nil {
		t.Fatal("animation C callback wiring differs")
	}
	if gui.EventRespInt(f.root.Func93(gui.WindowFocus(true))) != 123 {
		t.Error("shared key handler not installed")
	}
	if !f.root.GetFlags().Has(gui.StatusActive) {
		t.Error("connection window was not shown modal")
	}
	for _, p := range []unsafe.Pointer{unsafe.Pointer(f.root), unsafe.Pointer(st.top), unsafe.Pointer(st.bottom)} {
		if unsafe.Sizeof(p) > 4 && uintptr(p) <= uintptr(^uint32(0)) {
			t.Fatalf("fixture pointer %p does not exercise >4GiB native ownership", p)
		}
	}
	for i := 1; i <= 12; i++ {
		gui.AnimTick()
		if got, want := st.top.Window().Offs(), image.Pt(0, -240+20*i); got != want {
			t.Fatalf("top tick %d=%v, want %v", i, got, want)
		}
		if got, want := st.bottom.Window().Offs(), image.Pt(0, 480-20*i); got != want {
			t.Fatalf("bottom tick %d=%v, want %v", i, got, want)
		}
	}
	if st.top.State() != gui.AnimInDone || st.bottom.State() != gui.AnimInDone || len(f.sounds) != 0 || f.next != 0 {
		t.Fatal("slide-in changed sound/callback/state behavior")
	}
}

func TestConnectionSelector4AA270Modes(t *testing.T) {
	for _, mode := range []int{0, 1, 2} {
		t.Run(fmt.Sprint(mode), func(t *testing.T) {
			newConnectionFixture4AA270(t)
			connectionSetMode43AF50(mode)
			if connectionMode43AF50 != uint32(mode) || legacy.Get_dword_5d4594_2650652() != bool2int(mode == 1) {
				t.Fatal("primary mode and WOL predicate were conflated")
			}
		})
	}
}

func TestConnectionSelector4AA270AutomaticBranches(t *testing.T) {
	for _, tc := range []struct {
		name    string
		request uint32
		flags   noxflags.GameFlag
		solo    bool
		want    string
	}{
		{"ordinary waits for selection", 0, 0, false, "wait"},
		{"Solo request1 selects LAN during slide-in", 1, noxflags.GameFlag25, true, "LAN"},
		{"request2 selects shared Back", 2, noxflags.GameFlag25, true, "Back"},
		{"flag26 WOL wins over request1", 1, noxflags.GameFlag25 | noxflags.GameFlag26, true, "WOL"},
		{"flag26 WOL wins over request2", 2, noxflags.GameFlag25 | noxflags.GameFlag26, true, "WOL"},
		{"unknown request is not auto LAN", 3, noxflags.GameFlag25, true, "wait"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newConnectionFixture4AA270(t)
			f.request, f.solo = tc.request, tc.solo
			noxflags.SetGame(tc.flags)
			if showConnection4AA270(f.services) != 1 {
				t.Fatal("entry failed")
			}
			st := winConnection4AA270
			switch tc.want {
			case "LAN":
				if st.top.State() != gui.AnimOut || st.bottom.State() != gui.AnimOut || st.top.Func13Ptr != legacy.Get_nox_game_showGameSel_4379F0() {
					t.Fatal("Solo Quest did not select LAN before slide-in completion")
				}
				if connectionMode43AF50 != 0 || legacy.Get_dword_5d4594_2650652() != 0 || noxflags.HasGame(noxflags.GameFlag25) {
					t.Fatal("LAN mode/flag cleanup differs from original")
				}
				if want := []sound.ID{sound.SoundShellSlideOut, sound.SoundShellClick}; !reflect.DeepEqual(f.sounds, want) {
					t.Fatalf("sounds=%v, want %v", f.sounds, want)
				}
				if f.next != 0 {
					t.Fatal("network entry ran before animation completion")
				}
				gui.AnimTick()
				if f.next != 1 || winConnection4AA270 != nil {
					t.Fatal("C animation completion did not invoke network entry once")
				}
				gui.AnimTick()
				if f.next != 1 {
					t.Fatal("network entry repeated after released animation")
				}
			case "Back":
				if f.back != 1 || f.wol != 0 || st.top.Func13Ptr != nil {
					t.Fatal("request2 did not dispatch only shared Back")
				}
			case "WOL":
				if f.wol != 1 || f.back != 0 || st.top.Func13Ptr != nil || st.top.State() != gui.AnimIn || !noxflags.HasGame(noxflags.GameFlag25) || len(f.sounds) != 0 {
					t.Fatal("unavailable WOL fell through to LAN/Back or altered original error branch")
				}
			case "wait":
				if f.back != 0 || f.wol != 0 || st.top.Func13Ptr != nil || st.top.State() != gui.AnimIn || len(f.sounds) != 0 {
					t.Fatal("ordinary selection was not left pending")
				}
			}
		})
	}
}

func TestConnectionSelector4AA270ClickGuard(t *testing.T) {
	for _, tc := range []struct {
		name    string
		state   gui.AnimState
		solo    bool
		allowed bool
	}{
		{"ordinary slide-in blocked", gui.AnimIn, false, false},
		{"ordinary slide-out blocked", gui.AnimOut, false, false},
		{"ordinary completed selection allowed", gui.AnimInDone, false, true},
		{"Solo slide-in allowed", gui.AnimIn, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newConnectionFixture4AA270(t)
			f.solo = tc.solo
			noxflags.SetGame(noxflags.GameFlag25)
			if showConnection4AA270(f.services) != 1 {
				t.Fatal("entry failed")
			}
			st := winConnection4AA270
			st.top.SetState(tc.state)
			if gui.EventRespInt(f.root.Func94(&WindowEvent0x4007{Win: f.root.ChildByID(421)})) != 1 {
				t.Fatal("click result differs")
			}
			if tc.allowed {
				if st.top.Func13Ptr == nil || noxflags.HasGame(noxflags.GameFlag25) {
					t.Fatal("eligible LAN click was not handled")
				}
			} else if st.top.State() != tc.state || st.top.Func13Ptr != nil || !noxflags.HasGame(noxflags.GameFlag25) {
				t.Fatal("blocked click changed mode/animation")
			}
			if f.sounds[len(f.sounds)-1] != sound.SoundShellClick {
				t.Fatal("click feedback missing")
			}
		})
	}
}

func TestConnectionSelector4AA270NativeStartAndBackCallbacks(t *testing.T) {
	f := newConnectionFixture4AA270(t)
	if showConnection4AA270(f.services) != 1 {
		t.Fatal("entry failed")
	}
	st := winConnection4AA270
	for range 12 {
		gui.AnimTick()
	}
	st.top.Func13Ptr = legacy.Get_nox_game_showGameSel_4379F0()
	if st.top.Func12() != 1 {
		t.Fatal("native out-start C callback failed")
	}
	for i := 1; i <= 5; i++ {
		gui.AnimTick()
		if st.top.Window().Offs() != image.Pt(0, -40*i) || st.bottom.Window().Offs() != image.Pt(0, 240+40*i) || f.next != 0 {
			t.Fatalf("out tick %d differs from original", i)
		}
	}
	gui.AnimTick()
	if f.next != 1 || winConnection4AA270 != nil || gui.FindAnimForStateID(connectionState4AA270) != nil {
		t.Fatal("native done-out retained owner/animations")
	}
	if ccall.CallIntVoid(legacy.GetConnectionOutStart4AA450()) != 0 || ccall.CallIntVoid(legacy.GetConnectionOutDone4AA490()) != 0 {
		t.Fatal("stale completion callback was not rejected")
	}
}

func TestConnectionSelector4AA270WindowFaults(t *testing.T) {
	for _, missing := range []uint{400, 410, 411, 420, 421} {
		t.Run(fmt.Sprint(missing), func(t *testing.T) {
			f := newConnectionFixture4AA270(t)
			if missing == 400 {
				f.services.newWindow = func() *gui.Window { return nil }
			} else {
				f.root.ChildByID(missing).SetID(9999)
			}
			if showConnection4AA270(f.services) != 0 || winConnection4AA270 != nil || gui.FindAnimForStateID(connectionState4AA270) != nil || f.next != 0 {
				t.Fatal("bad stock window leaked into animation/network callbacks")
			}
			if missing != 400 && !f.root.Flags.Has(gui.StatusDestroyed) {
				t.Error("partial root was not released")
			}
		})
	}
}

func TestConnectionSelector4AA270HoverAndUnknownEvents(t *testing.T) {
	f := newConnectionFixture4AA270(t)
	if showConnection4AA270(f.services) != 1 {
		t.Fatal("entry failed")
	}
	if gui.EventRespInt(f.root.Func94(&WindowEvent0x4005{Win: f.root.ChildByID(421)})) != 1 || !reflect.DeepEqual(f.sounds, []sound.ID{sound.SoundShellSelect}) {
		t.Fatal("hover feedback differs")
	}
	if f.root.Func94(gui.WindowFocus(true)) != nil || f.next != 0 || f.wol != 0 {
		t.Fatal("unknown event was handled as a click")
	}
}
