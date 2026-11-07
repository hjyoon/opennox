package legacy

import (
	"bytes"
	"encoding/binary"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func TestSub409A70OriginalTableVariants(t *testing.T) {
	original := networkTableOriginal256(t)
	live := memmap.Slice(0x587000, 4704)[:24]
	before := bytes.Clone(live)
	t.Cleanup(func() { copy(live, before) })
	if uintptr(unsafe.Pointer(&live[0])) <= 0xffffffff && unsafe.Sizeof(uintptr(0)) > 4 {
		t.Fatal("native table must exercise an address above 4GiB")
	}
	variants := 0
	for _, row := range original.Rows {
		if row.Name != "settings" {
			continue
		}
		variants++
		t.Run(row.Variant, func(t *testing.T) {
			if len(row.Table) != 6 || len(row.Cases) != 256 {
				t.Fatal("complete independently recorded table domain")
			}
			for i, word := range row.Table {
				binary.LittleEndian.PutUint32(live[4*i:], word)
			}
			installed := bytes.Clone(live)
			for _, pair := range row.Cases {
				// The mask ignores every bit outside 0x17f0. Exercise both
				// signs through Go and the actual C short entry using the
				// existing original result, rather than computing an oracle.
				for _, input := range []uint16{uint16(pair[0]), uint16(pair[0]) | 0xe80f} {
					if got := uint32(sub409A70Native(int16(input))); got != pair[1] {
						t.Fatalf("Go input=%#x got=%d original=%d", input, got, pair[1])
					}
					if got := uint32(Sub_409A70(int(input))); got != pair[1] {
						t.Fatalf("C entry input=%#x got=%d original=%d", input, got, pair[1])
					}
				}
			}
			if !bytes.Equal(live, installed) {
				t.Fatal("lookup modified the live table")
			}
		})
	}
	if variants != 3 {
		t.Fatal("all three original settings variants")
	}
}
