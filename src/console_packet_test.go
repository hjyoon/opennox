package opennox

import (
	"encoding/binary"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/opennox/libs/noxnet/netmsg"
)

func TestConsoleServerCommandUTF16Packet(t *testing.T) {
	for _, tt := range []struct{ input, want string }{
		{"", ""}, {"watch Red Dragon", "watch Red Dragon"},
		{"say 안녕하세요", "say 안녕하세요"}, {"name 🐺 늑대", "name 🐺 늑대"},
		{strings.Repeat("x", 300), strings.Repeat("x", 254)},
		{strings.Repeat("한", 253) + "🐺", strings.Repeat("한", 253)},
		{strings.Repeat("🐺", 128), strings.Repeat("🐺", 127)},
		{"say abc\x00ignored", "say abc"},
	} {
		buf := consoleServerCommandPacket(5, 0xABCD, tt.input)
		if buf[0] != byte(netmsg.MSG_SERVER_CMD) || buf[1] != 5 || binary.LittleEndian.Uint16(buf[2:]) != 0xABCD {
			t.Fatal("stock console packet header changed")
		}
		units := utf16.Encode([]rune(tt.want))
		count := 0
		if len(units) != 0 {
			count = len(units) + 1
		}
		if int(buf[4]) != count || len(buf) != 5+2*(len(units)+1) || binary.LittleEndian.Uint16(buf[len(buf)-2:]) != 0 {
			t.Fatalf("invalid UTF-16 length/termination for %q: %d bytes count %d", tt.input, len(buf), buf[4])
		}
		for i, v := range units {
			if binary.LittleEndian.Uint16(buf[5+2*i:]) != v {
				t.Fatalf("UTF-16 unit %d corrupted", i)
			}
		}
	}
}
