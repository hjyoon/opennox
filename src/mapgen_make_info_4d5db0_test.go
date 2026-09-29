package opennox

import (
	"bytes"
	"encoding/binary"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
)

func mapgenInfoString4D5DB0(data []byte, off, size int) string {
	data = data[off : off+size]
	if end := bytes.IndexByte(data, 0); end >= 0 {
		data = data[:end]
	}
	return string(data)
}

func TestMapgenMakeInfo4D5DB0PreservesNativeBuffer(t *testing.T) {
	data, address := legacy.MapgenMakeInfoViaC4D5DB0()
	if len(data) != 0x5B8 {
		t.Fatalf("map info size = %#x, want %#x", len(data), 0x5B8)
	}
	if strconv.IntSize == 64 && address <= math.MaxUint32 {
		t.Fatalf("map info source address = %#x, want address above PE32 range", address)
	}

	for _, tc := range []struct {
		off  int
		size int
		want string
	}{
		{0, 64, "Generated Map"},
		{64, 64, "Generated Map"},
		{656, 64, "http://www.westwood.com"},
		{720, 128, "http://www.westwood.com"},
		{848, 128, "http://www.westwood.com"},
		{976, 256, "Generated Map"},
		{1232, 128, "Westwood Studios"},
	} {
		if got := mapgenInfoString4D5DB0(data, tc.off, tc.size); got != tc.want {
			t.Fatalf("map info string at %#x = %q, want %q", tc.off, got, tc.want)
		}
	}
	if got, want := binary.LittleEndian.Uint16(data[576:]), memmap.Uint16(0x587000, 198380); got != want {
		t.Fatalf("map info field 576 = %#x, want %#x", got, want)
	}
	if got, want := binary.LittleEndian.Uint32(data[592:]), memmap.Uint32(0x587000, 198384); got != want {
		t.Fatalf("map info field 592 = %#x, want %#x", got, want)
	}
	if got := binary.LittleEndian.Uint32(data[1392:]); got != 3 {
		t.Fatalf("map info field 1392 = %d, want 3", got)
	}
	date := mapgenInfoString4D5DB0(data, 1360, 32)
	if _, err := time.Parse("Mon, Jan 2 2006", date); err != nil {
		t.Fatalf("map info date = %q: %v", date, err)
	}
}
