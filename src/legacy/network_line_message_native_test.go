package legacy

import (
	"bytes"
	"testing"

	"github.com/opennox/libs/noxnet"
	"github.com/opennox/libs/noxnet/netmsg"
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
