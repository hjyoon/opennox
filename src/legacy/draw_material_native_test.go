package legacy

import (
	"encoding/binary"
	"testing"

	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/memmap"
)

func TestMaterialRGBReadsMappedExpansionFlag(t *testing.T) {
	flag := memmap.BlobByAddr(0x973F18).Data[2376:2380]
	old := [4]byte(flag)
	t.Cleanup(func() { copy(flag, old[:]) })

	data := new(noxrender.RenderData)
	binary.LittleEndian.PutUint32(flag, 0)
	if got := setMaterialRGB4340A0(data, 4, 22, 47, 130); got != 0 {
		t.Fatalf("packed material result = %#x, want 0", got)
	}
	binary.LittleEndian.PutUint32(flag, 1)
	if got := setMaterialRGB4340A0(data, 4, 22, 47, 130); got != 0x00820082 {
		t.Fatalf("expanded material result = %#x, want 0x00820082", got)
	}
}
