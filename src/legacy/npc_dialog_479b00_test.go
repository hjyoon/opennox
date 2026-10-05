package legacy

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/dialog"
	"github.com/opennox/opennox/v1/server"
)

const npcDialogWidthChild479B00 = "OPENNOX_TEST_NPC_DIALOG_REPEAT_WIDTH"

func npcDialogTestState479B00(t *testing.T) (*gui.Window, *gui.Window, *uint32, *netlist.List) {
	t.Helper()
	g := gui.New(nil)
	t.Cleanup(g.DestroyAll)
	root := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 400, 140, nil)
	button := g.NewWindowRaw(root, gui.StatusEnabled, 0, 0, 80, 20, nil)
	root.SetFunc94C(NPCDialogProc479B00())
	if unsafe.Sizeof(uintptr(0)) > 4 && (uintptr(root.C()) <= math.MaxUint32 || uintptr(button.C()) <= math.MaxUint32) {
		t.Fatalf("windows %p/%p did not exceed 4 GiB", root, button)
	}
	pause, filename, choice := npcDialogPauseSlot479B00(), memmap.PtrPtr(0x5D4594, 1115312), memmap.PtrUint8(0x5D4594, 1123516)
	packed := memmap.PtrUint32(0x5D4594, 1115312)
	oldPause, oldFilename, oldChoice, oldWindow, oldPacked := *pause, *filename, *choice, *npcDialogWindowSlot479B00(), *packed
	oldSound, oldDialogs, oldGetServer := ClientPlaySoundSpecial, Dialogs, GetServer
	t.Cleanup(func() {
		*pause, *filename, *choice, *npcDialogWindowSlot479B00() = oldPause, oldFilename, oldChoice, oldWindow
		*packed = oldPacked
		ClientPlaySoundSpecial, Dialogs, GetServer = oldSound, oldDialogs, oldGetServer
	})
	*pause, *choice, *npcDialogWindowSlot479B00() = 0, 0xA5, root
	*filename = nil
	if unsafe.Sizeof(uintptr(0)) > 4 {
		// The obsolete packed DWORD must not participate in native Repeat.
		*packed = 0x7FFF0000
	}
	list := netlist.New()
	list.Init()
	t.Cleanup(list.Free)
	GetServer = func() Server { return &netClientSendTestServer{srv: &server.Server{NetList: list}} }
	ClientPlaySoundSpecial = func(id sound.ID, volume int) {
		if id != sound.ID(766) || volume != 100 {
			t.Fatalf("click sound=%d/%d", id, volume)
		}
	}
	// Only PlayFile is reached here. No device/stream or state result is
	// supplied; the real Dialog service must retain the filename request.
	enabled := new(uint32)
	*enabled = 1
	Dialogs = dialog.NewDialog("dialog", enabled, new(uint32), new(uint32), new(uint32), new(uint32), nil,
		new(uint32), nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	return root, button, enabled, list
}

func TestNPCDialogNative479B00RepeatAbove4GiB(t *testing.T) {
	if os.Getenv(npcDialogWidthChild479B00) != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestNPCDialogNative479B00RepeatAbove4GiB$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), npcDialogWidthChild479B00+"=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("isolated native Repeat regression: %v\n%s", err, out)
		}
		t.Logf("%s", out)
		return
	}
	root, button, _, list := npcDialogTestState479B00(t)
	button.SetID(3907)
	filename, free := alloc.CString("War01A.scr:Opening-voice-file")
	defer free()
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(filename)) <= math.MaxUint32 {
		t.Fatalf("filename %p did not exceed 4 GiB", filename)
	}
	*memmap.PtrPtr(0x5D4594, 1115312) = unsafe.Pointer(filename)
	if got := gui.EventRespInt(root.Func94(&gui.RawEvent{Event: 0x4007, Arg1: uintptr(button.C()), Arg2: ^uintptr(0)})); got != 0 {
		t.Fatalf("Repeat return=%d", got)
	}
	if got := Dialogs.FileToRead(); got != "War01A.scr:Opening-voice-file" {
		t.Fatalf("Repeat requested %q, want full native filename", got)
	}
	if got := *memmap.PtrUint8(0x5D4594, 1123516); got != 0xA5 || root.GetFlags().IsHidden() {
		t.Fatalf("Repeat answered/closed dialog: choice=%d hidden=%t", got, root.GetFlags().IsHidden())
	}
	if got := list.CopyPacketsA(31, netlist.Kind0); len(got) != 0 {
		t.Fatalf("Repeat sent a choice packet: %x", got)
	}
	t.Logf("native callback=%p root=%p button=%p filename=%p request=%q", NPCDialogProc479B00(), root, button, filename, Dialogs.FileToRead())
}

func TestNPCDialogNative479B00ButtonsPauseAndHidden(t *testing.T) {
	for _, id := range []uint{3906, 3907, 3908, 3909, 0, math.MaxUint32} {
		for _, pause := range []uint32{0, 1, math.MaxUint32} {
			for _, hidden := range []bool{false, true} {
				t.Run(fmt.Sprintf("id_%d_pause_%x_hidden_%t", id, pause, hidden), func(t *testing.T) {
					root, button, _, list := npcDialogTestState479B00(t)
					button.SetID(id)
					root.SetHidden(hidden)
					*npcDialogPauseSlot479B00() = pause
					file, free := alloc.CString("native-repeat-file")
					defer free()
					*memmap.PtrPtr(0x5D4594, 1115312) = unsafe.Pointer(file)
					var calls int
					ClientPlaySoundSpecial = func(id sound.ID, volume int) {
						if id != 766 || volume != 100 || *memmap.PtrUint8(0x5D4594, 1123516) != 0xA5 {
							t.Fatal("sound did not precede button side effects")
						}
						calls++
					}
					if got := gui.EventRespInt(root.Func94(&gui.RawEvent{Event: 0x4007, Arg1: uintptr(button.C()), Arg2: ^uintptr(0)})); got != 0 {
						t.Fatalf("callback return=%d", got)
					}
					wantChoice, wantFile, wantSound := byte(0xA5), "", 0
					var wantPacket []byte
					if pause == 0 {
						wantSound = 1
						switch id {
						case 3906:
							if !hidden {
								wantChoice, wantPacket = 0, []byte{0xD0, 0x02, 0}
							}
						case 3907:
							wantFile = "native-repeat-file"
						case 3908:
							wantChoice, wantPacket = 1, []byte{0xD0, 0x02, 1}
						case 3909:
							wantChoice, wantPacket = 2, []byte{0xD0, 0x02, 2}
						}
					}
					gotPacket := list.CopyPacketsA(31, netlist.Kind0)
					if choice := *memmap.PtrUint8(0x5D4594, 1123516); choice != wantChoice || calls != wantSound || Dialogs.FileToRead() != wantFile || !bytes.Equal(gotPacket, wantPacket) || root.GetFlags().IsHidden() != hidden {
						t.Fatalf("choice=%d sound=%d file=%q packet=%x hidden=%t; want %d/%d/%q/%x/%t", choice, calls, Dialogs.FileToRead(), gotPacket, root.GetFlags().IsHidden(), wantChoice, wantSound, wantFile, wantPacket, hidden)
					}
				})
			}
		}
	}
}

func TestNPCDialogNative479B00ReadsLiveFileAfterSound(t *testing.T) {
	for _, id := range []uint{3906, 3907, 3908, 3909} {
		t.Run(fmt.Sprintf("cached_id_%d", id), func(t *testing.T) {
			root, button, _, list := npcDialogTestState479B00(t)
			button.SetID(id)
			before, freeBefore := alloc.CString("before-click")
			after, freeAfter := alloc.CString("live-file-after-click")
			defer freeBefore()
			defer freeAfter()
			*memmap.PtrPtr(0x5D4594, 1115312) = unsafe.Pointer(before)
			ClientPlaySoundSpecial = func(sound.ID, int) {
				button.SetID(3909)
				*npcDialogPauseSlot479B00() = 1
				*memmap.PtrPtr(0x5D4594, 1115312) = unsafe.Pointer(after)
			}
			if got := gui.EventRespInt(root.Func94(&gui.RawEvent{Event: 0x4007, Arg1: uintptr(button.C())})); got != 0 {
				t.Fatalf("return=%d", got)
			}
			wantFile, wantChoice := "", byte(0xA5)
			var wantPacket []byte
			switch id {
			case 3906:
				wantChoice, wantPacket = 0, []byte{0xD0, 0x02, 0}
			case 3907:
				wantFile = "live-file-after-click"
			case 3908:
				wantChoice, wantPacket = 1, []byte{0xD0, 0x02, 1}
			case 3909:
				wantChoice, wantPacket = 2, []byte{0xD0, 0x02, 2}
			}
			if got := list.CopyPacketsA(31, netlist.Kind0); !bytes.Equal(got, wantPacket) || Dialogs.FileToRead() != wantFile || *memmap.PtrUint8(0x5D4594, 1123516) != wantChoice {
				t.Fatalf("cached ID=%d packet=%x file=%q choice=%d", id, got, Dialogs.FileToRead(), *memmap.PtrUint8(0x5D4594, 1123516))
			}
		})
	}
}

func TestNPCDialogNative479B00NonClickAndNull(t *testing.T) {
	for _, event := range []int{0, 1, 2, 21, 23, 0x4000, 0x4006, 0x4008} {
		t.Run(fmt.Sprintf("event_%x", event), func(t *testing.T) {
			root, _, _, list := npcDialogTestState479B00(t)
			ClientPlaySoundSpecial = func(sound.ID, int) { t.Fatal("non-click played sound") }
			if got := gui.EventRespInt(root.Func94(&gui.RawEvent{Event: event, Arg1: ^uintptr(0)})); got != 0 || len(list.CopyPacketsA(31, netlist.Kind0)) != 0 {
				t.Fatalf("non-click return=%d", got)
			}
		})
	}
	t.Run("null_button", func(t *testing.T) {
		root, _, _, list := npcDialogTestState479B00(t)
		var sounds int
		ClientPlaySoundSpecial = func(sound.ID, int) { sounds++ }
		if got := gui.EventRespInt(root.Func94(&gui.RawEvent{Event: 0x4007})); got != 0 || sounds != 1 || len(list.CopyPacketsA(31, netlist.Kind0)) != 0 || *memmap.PtrUint8(0x5D4594, 1123516) != 0xA5 {
			t.Fatalf("null-button return=%d sound=%d", got, sounds)
		}
	})
}

func TestNPCDialogNative479B00EmptyAndDisabledRepeat(t *testing.T) {
	for _, mode := range []string{"nil", "empty", "disabled"} {
		t.Run(mode, func(t *testing.T) {
			root, button, enabled, list := npcDialogTestState479B00(t)
			button.SetID(3907)
			Dialogs.PlayFile("previous-request", 33)
			if mode != "nil" {
				name := ""
				if mode == "disabled" {
					name, *enabled = "new-file", 0
				}
				file, free := alloc.CString(name)
				defer free()
				*memmap.PtrPtr(0x5D4594, 1115312) = unsafe.Pointer(file)
			}
			if got := gui.EventRespInt(root.Func94(&gui.RawEvent{Event: 0x4007, Arg1: uintptr(button.C())})); got != 0 || Dialogs.FileToRead() != "previous-request" || len(list.CopyPacketsA(31, netlist.Kind0)) != 0 {
				t.Fatalf("return=%d file=%q", got, Dialogs.FileToRead())
			}
		})
	}
}

func TestNPCDialogNative479B00FullQueueDoesNotUndoChoice(t *testing.T) {
	for _, id := range []uint{3906, 3908, 3909} {
		t.Run(fmt.Sprintf("id_%d", id), func(t *testing.T) {
			root, button, _, list := npcDialogTestState479B00(t)
			button.SetID(id)
			full := bytes.Repeat([]byte{0x55}, 2048)
			if !list.AddToMsgListCli(31, netlist.Kind0, full) {
				t.Fatal("could not fill queue")
			}
			want := byte(0)
			if id != 3906 {
				want = byte(id - 3907)
			}
			if got := gui.EventRespInt(root.Func94(&gui.RawEvent{Event: 0x4007, Arg1: uintptr(button.C())})); got != 0 || *memmap.PtrUint8(0x5D4594, 1123516) != want || !bytes.Equal(list.CopyPacketsA(31, netlist.Kind0), full) {
				t.Fatalf("return=%d choice=%d", got, *memmap.PtrUint8(0x5D4594, 1123516))
			}
		})
	}
}
