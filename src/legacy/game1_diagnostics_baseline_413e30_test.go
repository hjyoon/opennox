package legacy

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/opennox/opennox/v1/common/memmap"
)

// This control also runs before replacing either public C body. It freezes the
// empty-head behavior without trying to encode a high host address in a DWORD.
func TestGAME1Diagnostics413E30EmptyPublicBaseline(t *testing.T) {
	const rowSize, rowCount = 84, 1024
	table := memmap.Slice(0x5D4594, 252284)[:rowSize*rowCount]
	before := bytes.Clone(table)
	packed := memmap.PtrUint32(0x5D4594, 338304)
	native := memmap.PtrPtr(0x5D4594, 338304)
	oldPacked, oldNative := *packed, *native
	t.Cleanup(func() { copy(table, before); *packed = oldPacked; *native = oldNative })
	for i := range table {
		table[i] = byte(i*31 + 7)
	}
	want := bytes.Clone(table)
	for i := 0; i < rowCount; i++ {
		binary.LittleEndian.PutUint32(want[i*rowSize+80:], 0)
	}
	*packed, *native = 0, nil
	memDumpFixturePublic413E30()
	if !bytes.Equal(table, want) {
		t.Fatal("actual public entry changed names or failed to clear all1024 DWORD counts")
	}
	if *packed != 0 || *native != nil {
		t.Fatal("empty head changed")
	}
}
