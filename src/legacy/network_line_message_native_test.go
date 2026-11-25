package legacy

import (
	"bytes"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet"
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/server"
)

func TestNetLineMessagePacket4D9EB0Encoding(t *testing.T) {
	wrap := make([]uint16, 256)
	for i := range len(wrap) - 1 {
		wrap[i] = 0x1111
	}
	for _, tc := range []struct {
		name    string
		words   []uint16
		narrow  bool
		flags   noxnet.TextFlags
		size    byte
		payload []byte
	}{
		{"empty", []uint16{0}, true, noxnet.TextUTF8, 1, []byte{0}},
		{"single-byte text", []uint16{'A', 0x00e9, 0}, true, noxnet.TextUTF8, 3, []byte{'A', 0xe9, 0}},
		{"wide text", []uint16{0x1111, 0xd83d, 0xde80, 0}, false, lineMessageWideFlag, 4, []byte{0x11, 0x11, 0x3d, 0xd8, 0x80, 0xde, 0, 0}},
		{"wire-length wrap", wrap, false, lineMessageWideFlag, 0, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var packet [521]byte
			length := buildLineMessagePacket(packet[:], tc.words, tc.narrow)
			want := make([]byte, 11, 11+len(tc.payload))
			want[0], want[3], want[8] = byte(netmsg.MSG_TEXT_MESSAGE), byte(tc.flags), tc.size
			want = append(want, tc.payload...)
			if !bytes.Equal(packet[:length], want) {
				t.Fatalf("message = %x, want %x", packet[:length], want)
			}
			var decoded noxnet.MsgText
			if _, err := decoded.Decode(packet[1:length]); err != nil || decoded.Flags != tc.flags || decoded.Size != tc.size ||
				!bytes.Equal(decoded.Data, tc.payload) {
				t.Fatalf("decoded = %+v, error=%v", decoded, err)
			}
		})
	}
}

func TestNetLineMessageGoTextPreservesPercentSigns(t *testing.T) {
	world := server.NewHigh(nil, nil, strman.New())
	t.Cleanup(world.Close)
	world.NetList.Init()
	t.Cleanup(world.NetList.Free)
	oldServer := GetServer
	GetServer = func() Server { return &chatLegacyServer528AC0{srv: world} }
	t.Cleanup(func() { GetServer = oldServer })

	const index = 3
	player := world.Players.ByIndRaw(index)
	player.PlayerInd = index
	update := &server.PlayerUpdateData{Player: player}
	unit := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	const message = "Progress: 100% (%s %% %d)"
	if !Nox_xxx_netSendLineMessage_4D9EB0(unit, message) {
		t.Fatal("could not queue line message")
	}
	packet := world.NetList.CopyPacketsA(index, netlist.Kind1)
	const header = 11
	if len(packet) != header+len(message)+1 || packet[0] != byte(netmsg.MSG_TEXT_MESSAGE) ||
		packet[8] != byte(len(message)+1) || string(packet[header:len(packet)-1]) != message || packet[len(packet)-1] != 0 {
		t.Fatalf("packet = %x, want literal %q", packet, message)
	}
}
