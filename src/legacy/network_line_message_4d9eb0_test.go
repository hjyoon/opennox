package legacy

import (
	"bytes"
	"strings"
	"testing"
	"unicode/utf16"
	"unsafe"

	"github.com/opennox/libs/noxnet"
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
	"github.com/opennox/opennox/v1/server"
)

func TestNetLineMessage4D9EB0VariadicCEntry(t *testing.T) {
	unit, freeUnit := alloc.NewHigh(server.Object{})
	t.Cleanup(freeUnit)
	if got := cfixture.LineMessage4D9EB0(unsafe.Pointer(unit), nil, 37); got != uintptr(unsafe.Pointer(unit)) {
		t.Fatalf("non-player result = %#x, want original address %p", got, unit)
	}

	world := lineMessageTestWorld(t)
	const index = 7
	player := world.Players.ByIndRaw(index)
	player.PlayerInd = index
	update, freeUpdate := alloc.NewHigh(server.PlayerUpdateData{})
	t.Cleanup(freeUpdate)
	update.Player = player
	unit.ObjClass, unit.UpdateData = object.ClassPlayer, unsafe.Pointer(update)

	for _, test := range []struct {
		format string
		want   []byte
	}{
		{"%S %04d%s", []byte("xp 0037!\x00")},
		{strings.Repeat("x", 256), nil}, // 255 UTF-16 units plus NUL wrap the wire size.
	} {
		format := append(utf16.Encode([]rune(test.format)), 0)
		if got := cfixture.LineMessage4D9EB0(unsafe.Pointer(unit), &format[0], 37); got != 1 {
			t.Fatalf("send result = %#x", got)
		}
		packet := world.NetList.CopyPacketsA(index, netlist.Kind1)
		if len(packet) == 0 || packet[0] != byte(netmsg.MSG_TEXT_MESSAGE) {
			t.Fatalf("unexpected packet %x", packet)
		}
		var message noxnet.MsgText
		if used, err := message.Decode(packet[1:]); err != nil || used != len(packet)-1 ||
			message.Flags != noxnet.TextUTF8 || message.Size != uint8(len(test.want)) || !bytes.Equal(message.Data, test.want) {
			t.Fatalf("packet %x decoded as %+v (used=%d err=%v), want %q", packet, message, used, err, test.want)
		}
	}
}

type netLineMessagePlayer4D9EB0 struct {
	index byte
}

type netLineMessageUpdateData4D9EB0 struct {
	player *netLineMessagePlayer4D9EB0
}

func TestNetLineMessagePlayerIndex4D9EB0(t *testing.T) {
	player := &netLineMessagePlayer4D9EB0{index: 0xfe}
	updateData := &netLineMessageUpdateData4D9EB0{player: player}
	var events []string
	got := netLineMessagePlayerIndex4D9EB0(
		updateData,
		func(updateData *netLineMessageUpdateData4D9EB0) *netLineMessagePlayer4D9EB0 {
			events = append(events, "player")
			return updateData.player
		},
		func(player *netLineMessagePlayer4D9EB0) byte {
			events = append(events, "index")
			return player.index
		},
	)
	if want := byte(0xfe); got != want {
		t.Fatalf("player index = %#x, want %#x", got, want)
	}
	if got, want := len(events), 2; got != want || events[0] != "player" || events[1] != "index" {
		t.Fatalf("events = %v, want [player index]", events)
	}

	player.index = 0x7b
	if got, want := netLineMessagePlayerIndex4D9EB0(
		updateData,
		func(updateData *netLineMessageUpdateData4D9EB0) *netLineMessagePlayer4D9EB0 {
			return updateData.player
		},
		func(player *netLineMessagePlayer4D9EB0) byte { return player.index },
	), byte(0x7b); got != want {
		t.Fatalf("reloaded player index = %#x, want %#x", got, want)
	}
}

func TestNetLineMessagePlayerIndex4D9EB0FaultsOnMissingPointers(t *testing.T) {
	tests := []struct {
		name       string
		updateData *netLineMessageUpdateData4D9EB0
	}{
		{name: "update data", updateData: nil},
		{name: "player", updateData: &netLineMessageUpdateData4D9EB0{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("missing legacy pointer did not fault")
				}
			}()
			_ = netLineMessagePlayerIndex4D9EB0(
				test.updateData,
				func(updateData *netLineMessageUpdateData4D9EB0) *netLineMessagePlayer4D9EB0 {
					return updateData.player
				},
				func(player *netLineMessagePlayer4D9EB0) byte { return player.index },
			)
		})
	}
}
