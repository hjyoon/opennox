package legacy

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestQuestGameOverTimer49B6E0CEntryDelegation(t *testing.T) {
	old := questGameOverTimerCall49B6E0
	t.Cleanup(func() { questGameOverTimerCall49B6E0 = old })
	for _, value := range []int32{0, 1, -1, math.MinInt32, math.MaxInt32} {
		calls := 0
		questGameOverTimerCall49B6E0 = func() int32 { calls++; return value }
		if got := questGameOverTimerCEntry49B6E0(); got != value || calls != 1 {
			t.Fatalf("return=%d want=%d calls=%d", got, value, calls)
		}
	}
}

type questGameOverTimerClient49B6E0 struct {
	Client
	cli *client.Client
}

func (c *questGameOverTimerClient49B6E0) Cli() *client.Client { return c.cli }

type questGameOverTimerServer49B6E0 struct {
	Server
	srv *server.Server
}

func (s *questGameOverTimerServer49B6E0) S() *server.Server { return s.srv }

func TestQuestGameOverTimer49B6E0NativeHost(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	cli := client.NewClient(nil, nil, srv)
	t.Cleanup(cli.Close)
	t.Cleanup(cli.GUI.DestroyAll)
	oldClient, oldServer := GetClient, GetServer
	GetClient = func() Client { return &questGameOverTimerClient49B6E0{cli: cli} }
	GetServer = func() Server { return &questGameOverTimerServer49B6E0{srv: srv} }
	t.Cleanup(func() { GetClient, GetServer = oldClient, oldServer })
	root := cli.GUI.NewWindowRaw(nil, 0, 0, 0, 100, 100, nil)
	var gotText string
	var calls int
	child := cli.GUI.NewWindowRaw(root, 0, 0, 0, 10, 10, func(_ *gui.Window, ev gui.WindowEvent) gui.WindowEventResp {
		if ev, ok := ev.(*gui.StaticTextSetText); ok {
			calls++
			gotText = ev.Str
		}
		return gui.RawEventResp(math.MaxUint32)
	})
	child.SetID(10712)
	player, freePlayer := alloc.New(server.Player{})
	t.Cleanup(freePlayer)
	player.PlayerInd = 31
	if unsafe.Sizeof(uintptr(0)) > 4 {
		for _, p := range []unsafe.Pointer{unsafe.Pointer(root), unsafe.Pointer(child), unsafe.Pointer(player)} {
			if uintptr(p) <= math.MaxUint32 {
				t.Fatalf("native C-owned pointer=%p, want above 4 GiB", p)
			}
		}
		if off := unsafe.Offsetof(player.PlayerInd); off == 2064 {
			t.Fatal("64-bit native PlayerInd unexpectedly shares PE32 offset")
		}
	}
	t.Logf("Player=%p PlayerInd offset=%d PE32 byte=%d", player, unsafe.Offsetof(player.PlayerInd), *(*byte)(unsafe.Add(unsafe.Pointer(player), 2064)))
	rootSlot := questGameOverTimerRootSlot49B6E0()
	startSlot := memmap.PtrUint32(0x5D4594, 1303456)
	buffer := unsafe.Slice((*byte)(memmap.Ptr(0x5D4594+1301852)), 64)
	empty := unsafe.Slice((*byte)(memmap.Ptr(0x5D4594+1303464)), 4)
	oldBuffer, oldEmpty := append([]byte(nil), buffer...), append([]byte(nil), empty...)
	oldRoot, oldStart, oldPlayer := *rootSlot, *startSlot, Get_dword_8531A0_2576()
	oldFPS, oldFrame := gameFPSHook, gameFrameHook
	t.Cleanup(func() {
		*rootSlot, *startSlot = oldRoot, oldStart
		Set_dword_8531A0_2576(oldPlayer)
		copy(buffer, oldBuffer)
		copy(empty, oldEmpty)
		gameFPSHook, gameFrameHook = oldFPS, oldFrame
	})
	*rootSlot, *startSlot = root, 100
	Set_dword_8531A0_2576(player)
	clear(empty)
	gameFPSHook = func() uint32 { return 30 }
	gameFrameHook = func() uint32 { return 100 }
	if got := questGameOverTimerCEntry49B6E0(); got != 0 || calls != 1 || gotText != "" {
		t.Fatalf("host timer return=%d calls=%d text=%q, want 0 and one empty text update", got, calls, gotText)
	}
	// A non-host with a misleading PE32 byte must still receive the timer.
	player.PlayerInd = 0
	if unsafe.Offsetof(player.PlayerInd) != 2064 {
		*(*byte)(unsafe.Add(unsafe.Pointer(player), 2064)) = 31
	}
	fpsCalls := 0
	gameFPSHook = func() uint32 {
		fpsCalls++
		if fpsCalls == 1 {
			return 30
		}
		return 7
	}
	gameFrameHook = func() uint32 { *startSlot = 201; return 100 }
	before := *player
	wantText := fmt.Sprintf("%s - 33", srv.Strings().GetStringInFile("Rules.c:Time", `C:\NoxPost\src\client\Gui\GUIGGOvr.c`))
	if got := questGameOverTimerCEntry49B6E0(); got != 0 || calls != 2 || gotText != wantText || fpsCalls != 1 || *player != before {
		t.Fatalf("non-host return=%d calls=%d text=%q FPS calls=%d", got, calls, gotText, fpsCalls)
	}
	// Host suppression reads the inline string and bypasses DIV at FPS 0.
	player.PlayerInd = 31
	empty[0], empty[1], empty[2], empty[3] = 0x1e, 0x22, 0, 0 // UTF-16 infinity.
	gameFPSHook = func() uint32 { return 0 }
	if got := questGameOverTimerCEntry49B6E0(); got != 0 || calls != 3 || gotText != "∞" {
		t.Fatalf("inline host text return=%d calls=%d text=%q", got, calls, gotText)
	}
	// Invisible and absent roots must not query the clock or touch the buffer.
	gameFPSHook = func() uint32 { t.Fatal("clock read for hidden/nil root"); return 0 }
	root.SetHidden(true)
	if got := questGameOverTimerCEntry49B6E0(); got != 1 || calls != 3 {
		t.Fatalf("hidden return=%d calls=%d", got, calls)
	}
	*rootSlot = nil
	if got := questGameOverTimerCEntry49B6E0(); got != 0 || calls != 3 {
		t.Fatalf("nil return=%d calls=%d", got, calls)
	}
}

func TestQuestGameOverTimer49B6E0NativeFormatAndFreshRoot(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	cli := client.NewClient(nil, nil, srv)
	t.Cleanup(cli.Close)
	t.Cleanup(cli.GUI.DestroyAll)
	oldClient := GetClient
	GetClient = func() Client { return &questGameOverTimerClient49B6E0{cli: cli} }
	t.Cleanup(func() { GetClient = oldClient })
	rootSlot := questGameOverTimerRootSlot49B6E0()
	startSlot := memmap.PtrUint32(0x5D4594, 1303456)
	buffer := unsafe.Slice((*byte)(memmap.Ptr(0x5D4594+1301852)), 64)
	oldRoot, oldStart, oldCall := *rootSlot, *startSlot, questGameOverTimerCall49B6E0
	oldPlayer := Get_dword_8531A0_2576()
	oldFPS, oldFrame := gameFPSHook, gameFrameHook
	oldBuffer := append([]byte(nil), buffer...)
	t.Cleanup(func() {
		*rootSlot, *startSlot, questGameOverTimerCall49B6E0 = oldRoot, oldStart, oldCall
		Set_dword_8531A0_2576(oldPlayer)
		gameFPSHook, gameFrameHook = oldFPS, oldFrame
		copy(buffer, oldBuffer)
	})
	root := cli.GUI.NewWindowRaw(nil, 0, 0, 0, 100, 100, nil)
	replacement := cli.GUI.NewWindowRaw(nil, 0, 0, 0, 100, 100, nil)
	var gotText string
	var gotBuffer uintptr
	child := cli.GUI.NewWindowRaw(replacement, 0, 0, 0, 10, 10, func(_ *gui.Window, ev gui.WindowEvent) gui.WindowEventResp {
		if ev, ok := ev.(*gui.StaticTextSetText); ok {
			gotText = ev.Str
			gotBuffer, _ = ev.EventArgsC()
		}
		return gui.RawEventResp(math.MaxUint32)
	})
	child.SetID(10712)
	title := unsafe.Pointer(alloc.InternCString16("시간"))
	for _, p := range []unsafe.Pointer{root.C(), replacement.C(), child.C(), title, unsafe.Pointer(questGameOverTimerBuffer49B6E0())} {
		if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(p) <= math.MaxUint32 {
			t.Fatalf("native pointer=%p, want above 4 GiB", p)
		}
	}
	Set_dword_8531A0_2576(nil)
	gameFPSHook = func() uint32 { return 1 }
	gameFrameHook = func() uint32 { return 0 }
	*startSlot = math.MaxInt32 - 30
	h := questGameOverTimerNativeHooks49B6E0()
	h.loadText = func(key, source string, line int32) unsafe.Pointer {
		if key != "Rules.c:Time" || source != `C:\NoxPost\src\client\Gui\GUIGGOvr.c` || line != 265 {
			t.Fatalf("lookup=%q/%q/%d", key, source, line)
		}
		*rootSlot = replacement
		return title
	}
	questGameOverTimerCall49B6E0 = func() int32 { return questGameOverTimer49B6E0(h) }
	*rootSlot = root
	if got := questGameOverTimerCEntry49B6E0(); got != 0 || gotText != "시간 - 2147483647" || gotBuffer != uintptr(unsafe.Pointer(questGameOverTimerBuffer49B6E0())) {
		t.Fatalf("native format return=%d text=%q buffer=%#x", got, gotText, gotBuffer)
	}
}
