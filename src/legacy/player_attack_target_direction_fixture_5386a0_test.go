package legacy

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math"
	"os"
	"testing"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

// The legacy package alone registers zero C blobs, without the application's
// original-data loader. Supply only the unchanged original direction table
// for this package's native C tests. Do not change any frozen test, expected
// value, input coordinate, health, animation, clock, seed or gameplay RNG.
// Restore the entire pre-test table after m.Run, including failed runs.
func TestMain(m *testing.M) {
	const originalSHA256 = "fbfdbd07ab3c6fbec2eb3768e69f1c32eb63c77b6e80e9957b9567c78d61bfcf"
	original := make([]byte, 2048)
	previous := make([]uint32, 512)
	for direction := 0; direction < 256; direction++ {
		x, y := server.SinCosDir(byte(direction))
		binary.LittleEndian.PutUint32(original[8*direction:], math.Float32bits(x))
		binary.LittleEndian.PutUint32(original[8*direction+4:], math.Float32bits(y))
	}
	sum := sha256.Sum256(original)
	if hex.EncodeToString(sum[:]) != originalSHA256 {
		panic("test initialization is not the unchanged original direction table")
	}
	for i := range previous {
		previous[i] = *memmap.PtrUint32(0x587000, 194136+uintptr(i)*4)
		*memmap.PtrUint32(0x587000, 194136+uintptr(i)*4) = binary.LittleEndian.Uint32(original[i*4:])
	}
	result := m.Run()
	for i, bits := range previous {
		*memmap.PtrUint32(0x587000, 194136+uintptr(i)*4) = bits
	}
	os.Exit(result)
}
