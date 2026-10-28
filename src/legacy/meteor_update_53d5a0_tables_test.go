package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func TestMeteorNativeDirectionTable53D5A0(t *testing.T) {
	blob, err := os.ReadFile("../server/testdata/meteor_update_53d5a0_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture meteorOriginalFixture53D5A0
	if err = json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	stock, err := os.ReadFile("../common/memmap/nox/blobdata/blob_587000.dat")
	if err != nil {
		t.Fatal(err)
	}
	const off = 0x5b6658 - 0x587000
	want := make([]byte, 2048)
	for n, v := range fixture.Table.Words {
		binary.LittleEndian.PutUint32(want[n*4:], v)
	}
	if !bytes.Equal(stock[off:off+len(want)], want) {
		t.Fatal("production direction table differs from original")
	}
	// Standalone legacy tests omit application blob loading. Prepare only this
	// unchanged production region, preserving every previous byte and TestMain.
	actual := unsafe.Slice((*byte)(memmap.Ptr(0x5b6658)), len(want))
	before := append([]byte(nil), actual...)
	t.Cleanup(func() { copy(actual, before) })
	copy(actual, stock[off:off+len(want)])
	read := meteorShowerRawRuntime53D5A0().DirectionFloat
	for n, v := range fixture.Table.Words {
		if math.Float32bits(read(int32(n))) != v {
			t.Fatalf("native direction word %d", n)
		}
	}
}
