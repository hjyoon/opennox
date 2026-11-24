package legacy

import (
	"bytes"
	"encoding/binary"
	"testing"
	"unicode/utf16"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/server"
)

func TestVoteWindowNativeSelectionPackets(t *testing.T) {
	g := newWidgetCallbackGUI(t)
	defer g.DestroyAll()
	s := GetServer().S()
	s.NetList.Init()
	defer s.NetList.Free()
	first, second := s.Players.ByIndRaw(0), s.Players.ByIndRaw(1)
	*first = server.Player{Active: 1, PlayerInd: 0}
	*second = server.Player{Active: 1, PlayerInd: 1}
	const firstName, secondName = "Åsa", "Béla"
	copy(first.NameFinal[:], utf16.Encode([]rune(firstName)))
	copy(second.NameFinal[:], utf16.Encode([]rune(secondName)))

	live := voteCurrentNames()
	before := memmap.PtrT[voteNames](0x5D4594, 1195512)
	oldLive, oldBefore := *live, *before
	defer func() { *live, *before = oldLive, oldBefore }()
	oldLoader := Nox_new_window_from_file
	defer func() { Nox_new_window_from_file = oldLoader }()
	var roster *gui.Window
	Nox_new_window_from_file = func(name string, fn gui.WindowFunc) *gui.Window {
		if name != "GuiKick.wnd" {
			t.Fatalf("loaded %q, want GuiKick.wnd", name)
		}
		root := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 240, 120, fn)
		draw := gui.WindowData{Window: root, Style: gui.StyleScrollListBox}
		roster = gui.NewScrollListBoxRaw(g, root, gui.StatusEnabled, 0, 0, 200, 100,
			&draw, &gui.ScrollListBoxData{Count: 4, Field_4: 1})
		roster.SetID(4320)
		g.NewWindowRaw(root, gui.StatusEnabled, 0, 0, 20, 20, nil).SetID(4321)
		return root
	}
	if got := sub_48D000_initGuiKick(); got != 1 {
		t.Fatalf("vote window init = %d, want 1", got)
	}
	defer sub_48D450()
	requireNativeWindowAddress(t, roster)
	win := (*nox_window)(roster.C())
	selectRow := func(row uintptr) {
		g2ServerEvent(win, 16399, 0, 0)
		g2ServerText(win, 16397, unsafe.Pointer(internWStr(firstName)), 4)
		g2ServerText(win, 16397, unsafe.Pointer(internWStr(secondName)), 4)
		g2ServerEvent(win, 16405, row, 0)
		if got := sub_48D120(); got != 1 {
			t.Fatalf("selected %d records for row %d, want 1", got, row)
		}
	}
	selectRow(0)
	selectRow(1)

	want := make([]byte, 3*52)
	writeVote := func(index int, command byte, name string) {
		p := want[index*52 : (index+1)*52]
		p[0], p[1] = byte(netmsg.MSG_VOTE), command
		for i, ch := range utf16.Encode([]rune(name)) {
			binary.LittleEndian.PutUint16(p[2+i*2:], ch)
		}
	}
	writeVote(0, 0, firstName)
	writeVote(1, 2, firstName)
	writeVote(2, 0, secondName)
	if got := s.NetList.CopyPacketsA(31, netlist.Kind0); !bytes.Equal(got, want) {
		t.Fatalf("vote packets = %x, want %x", got, want)
	}
}
