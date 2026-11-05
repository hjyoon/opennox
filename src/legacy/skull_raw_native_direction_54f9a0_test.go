package legacy

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func TestSkullRawNativeDirectionData54F9A0(t *testing.T) {
	if skullRawOriginalStateOverride54F9A0 != nil || skullRawOriginalRuntimeOverride54F9A0 != nil {
		t.Fatal("actual default native direction reader required")
	}
	data, err := os.ReadFile("../server/testdata/skull_raw_direction_data_54f9a0.json")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != "3e1fe81326cb6c663fc63e12dcc35ac13e78b5c7c00a088ee78fc946e329d93d" {
		t.Fatal("independent original direction seal")
	}
	var reference struct {
		Rows []struct {
			Direction int16
			Words     [2]uint32
		}
	}
	if json.Unmarshal(data, &reference) != nil || len(reference.Rows) != 260 {
		t.Fatal("complete independent direction transport")
	}
	stock, err := os.ReadFile("../common/memmap/nox/blobdata/blob_587000.dat")
	if err != nil {
		t.Fatal(err)
	}
	want := make([]byte, 2048)
	for _, row := range reference.Rows {
		if row.Direction >= 0 && row.Direction < 256 {
			for component, value := range row.Words {
				binary.LittleEndian.PutUint32(want[int(row.Direction)*8+component*4:], value)
			}
		}
	}
	const offset = 0x5b6658 - 0x587000
	if !bytes.Equal(want, stock[offset:offset+len(want)]) {
		t.Fatal("actual stock table differs from independently sealed original")
	}
	// Standalone legacy TestMain does not load application blob data. Prepare
	// only this byte-exact stock region and restore every previous byte.
	actual := unsafe.Slice((*byte)(memmap.Ptr(0x5b6658)), len(want))
	before := bytes.Clone(actual)
	copy(actual, stock[offset:offset+len(want)])
	t.Cleanup(func() { copy(actual, before) })
	runtime := skullRawRuntime54F9A0()
	seen := 0
	for _, row := range reference.Rows {
		// Malformed indices use explicitly declared initialized-PE projections
		// in the original consumer, without a general native alias certificate.
		if row.Direction < 0 || row.Direction > 255 {
			continue
		}
		for component, expected := range row.Words {
			if actual := math.Float32bits(runtime.DirectionComponent(row.Direction, uint32(component))); actual != expected {
				t.Fatalf("native direction%d component%d=%08x want%08x", row.Direction, component, actual, expected)
			}
		}
		seen++
	}
	if seen != 256 || runtime.RoundToNearest || unsafe.Sizeof(server.SkullRawState54F9A0{}) != 24 {
		t.Fatal("complete default direction domain, startup rounding and six fixed-width numeric dwords")
	}
}
