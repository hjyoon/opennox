package legacy

import (
	"bytes"
	"testing"

	"github.com/opennox/libs/noxnet/netmsg"
)

func TestClientChatNativeMessageWire(t *testing.T) {
	for _, test := range []struct {
		name         string
		text         []uint16
		narrow, team bool
		want         []byte
	}{
		{
			name: "single-byte team message",
			text: []uint16{' ', ' ', 'c', 'a', 'f', 0xe9}, narrow: true, team: true,
			want: []byte{0xa8, 0x34, 0x12, 3, 0x67, 0x45, 0xab, 0x89, 5, 0, 0, 'c', 'a', 'f', 0xe9, 0},
		},
		{
			name: "UTF-16 message",
			text: []uint16{' ', 0xd574, 0xb3c5},
			want: []byte{0xa8, 0x34, 0x12, 4, 0x67, 0x45, 0xab, 0x89, 3, 0, 0, 0x74, 0xd5, 0xc5, 0xb3, 0, 0},
		},
		{name: "only spaces", text: []uint16{' ', ' '}, narrow: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			message := clientChatMessage(test.text, test.narrow, test.team, 0x1234, 0x4567, 0x89ab)
			if test.want == nil {
				if message != nil {
					t.Fatalf("spaces produced a packet: %+v", message)
				}
				return
			}
			got, err := netmsg.Append(nil, message)
			if err != nil || !bytes.Equal(got, test.want) {
				t.Fatalf("packet %x, want %x, error %v", got, test.want, err)
			}
		})
	}
}
