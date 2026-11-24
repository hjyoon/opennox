package legacy

import (
	"bytes"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestPriorityMessageNativeCEntries(t *testing.T) {
	s := server.NewHigh(nil, nil, strman.New())
	t.Cleanup(s.Close)
	s.NetList.Init()
	t.Cleanup(s.NetList.Free)
	oldGetServer := GetServer
	GetServer = func() Server { return &chatLegacyServer528AC0{srv: s} }
	t.Cleanup(func() { GetServer = oldGetServer })

	newPlayerUnit := func(index ntype.PlayerInd) *server.Object {
		p := s.Players.ByIndRaw(index)
		p.Active, p.PlayerInd = 1, byte(index)
		u, freeUpdate := alloc.NewHigh(server.PlayerUpdateData{})
		t.Cleanup(freeUpdate)
		u.Player = p
		unit, freeUnit := alloc.NewHigh(server.Object{})
		t.Cleanup(freeUnit)
		unit.ObjClass = object.ClassPlayer
		unit.UpdateData = unsafe.Pointer(u)
		p.PlayerUnit = unit
		return unit
	}
	first := newPlayerUnit(3)
	blocked := newPlayerUnit(7)
	newPlayerUnit(31)
	s.Players.ByIndRaw(5).Active = 1 // Include an active player with no unit.
	s.Players.SetXxx(blocked, 1)
	nonPlayer, freeNonPlayer := alloc.NewHigh(server.Object{})
	t.Cleanup(freeNonPlayer)
	nonPlayer.ObjClass = object.ClassMonster

	limit := bytes.Repeat([]byte{'a'}, 48)
	extended := []byte{0x81, 0xfe, 'a'}
	for _, tc := range []struct {
		name string
		unit *server.Object
		text []byte
		want []byte
	}{
		{"null-text", first, nil, nil},
		{"empty-text", first, []byte{}, nil},
		{"null-recipient", nil, extended, nil},
		{"non-player-recipient", nonPlayer, extended, nil},
		{"blocked-recipient", blocked, extended, nil},
		{"high-bytes", first, extended, extended},
		{"terminator-stops-copy", first, []byte("first\x00second"), []byte("first")},
		{"maximum-length", first, limit, limit},
		{"over-maximum-length", first, bytes.Repeat([]byte{'a'}, 49), nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var message *byte
			if tc.text != nil {
				data := append(bytes.Clone(tc.text), 0)
				message = &data[0]
			}
			callNetPriMsgToPlayerC(tc.unit, message, 0xa3)
			var want []byte
			if tc.want != nil {
				want = append([]byte{0xa9, 15, 0xa3}, tc.want...)
				want = append(want, 0)
			}
			if got := s.NetList.CopyPacketsA(3, netlist.Kind1); !bytes.Equal(got, want) {
				t.Fatalf("recipient packet = %x, want %x", got, want)
			}
			if got := s.NetList.CopyPacketsA(7, netlist.Kind1); len(got) != 0 {
				t.Fatalf("blocked recipient received %x", got)
			}
		})
	}

	text := []byte{'a', 0}
	if got := callNetPrintLineToAllC(&text[0]); got != 0 {
		t.Fatalf("broadcast return = %d, want 0", got)
	}
	want := []byte{0xa9, 15, 0, 'a', 0}
	for _, index := range []ntype.PlayerInd{3, 31} {
		if got := s.NetList.CopyPacketsA(index, netlist.Kind1); !bytes.Equal(got, want) {
			t.Fatalf("player %d broadcast = %x, want %x", index, got, want)
		}
	}
	for _, index := range []ntype.PlayerInd{5, 7} {
		if got := s.NetList.CopyPacketsA(index, netlist.Kind1); len(got) != 0 {
			t.Fatalf("unexpected broadcast to player %d: %x", index, got)
		}
	}
}
