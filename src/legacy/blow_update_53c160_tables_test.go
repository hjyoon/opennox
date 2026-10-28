package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"strconv"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func TestBlowNativeDirectionTables53C160(t *testing.T) {
	blob, err := os.ReadFile("../server/testdata/blow_update_53c160_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture blowOriginalFixture53C160
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	// The standalone legacy package omits the application's blob loader.
	// Use its unchanged production data for these regions only, preserving
	// every pre-test byte and the existing TestMain initialization.
	stock, err := os.ReadFile("../common/memmap/nox/blobdata/blob_587000.dat")
	if err != nil {
		t.Fatal(err)
	}
	for _, table := range fixture.Tables {
		addr, err := strconv.ParseUint(table.Address, 0, 64)
		if err != nil {
			t.Fatal(err)
		}
		if addr != 0x5b5e58 && addr != 0x5b6658 && addr != 0x5bf2cc {
			continue
		}
		expected, err := hex.DecodeString(table.Hex)
		if err != nil {
			t.Fatal(err)
		}
		off := int(addr - 0x587000)
		if off < 0 || off+len(expected) > len(stock) || !bytes.Equal(stock[off:off+len(expected)], expected) {
			t.Fatalf("production blob %#x differs from unchanged original", addr)
		}
		actual := unsafe.Slice((*byte)(memmap.Ptr(uintptr(addr))), len(expected))
		previous := append([]byte(nil), actual...)
		t.Cleanup(func() { copy(actual, previous) })
		copy(actual, stock[off:off+len(expected)])
		if !bytes.Equal(actual, expected) {
			for word := 0; word < len(expected)/4; word++ {
				got, want := binary.LittleEndian.Uint32(actual[word*4:]), binary.LittleEndian.Uint32(expected[word*4:])
				if got != want {
					t.Fatalf("native direction table %#x word %d: got %#x want %#x", addr, word, got, want)
				}
			}
		}
		if addr == 0x5b6658 {
			read := blowUpdateRuntime53C160().DirectionFloat
			for index := int32(0); index < 512; index++ {
				if math.Float32bits(read(index)) != binary.LittleEndian.Uint32(expected[index*4:]) {
					t.Fatalf("native float word %d", index)
				}
			}
		}
	}
	for direction := int16(0); direction < 256; direction++ {
		var expected server.IndexedDirectionVector509E20
		server.IndexedDirection509E20(int32(direction), &expected)
		actual := blowIndexedDirectionNative53C160(direction)
		if int32(actual.X) != expected.X || int32(actual.Y) != expected.Y {
			t.Fatalf("native indexed direction %d: %v want %v", direction, actual, expected)
		}
	}
}
