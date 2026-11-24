package legacy

import (
	"os"
	"os/exec"
	"testing"

	"github.com/opennox/opennox/v1/common/memmap"
)

func TestWOLWorldRegionsLoadFromNativePointers(t *testing.T) {
	const child = "OPENNOX_TEST_WOL_WORLD_REGIONS"
	if os.Getenv(child) != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestWOLWorldRegionsLoadFromNativePointers$", "-test.count=1")
		cmd.Env = append(os.Environ(), child+"=1")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("isolated world region load: %v\n%s", err, out)
		}
		return
	}
	InitBlobData()
	const data, world = 0x587000, 0x5D4594
	const recordBase, polygonBase = 165744, 552228
	const recordSize, polygonSize = 16, 140
	sub_49FDB0(0)
	defer sub_49FF20()
	for i := range 17 {
		record := uintptr(recordBase + i*recordSize)
		polygon := uintptr(polygonBase + (i+1)*polygonSize)
		name := GoStringP(*memmap.PtrPtr(data, record+12))
		got := GoStringP(memmap.PtrOff(world, polygon+4))
		count := memmap.Uint16(world, polygon+128)
		if name == "" || got != name || count != uint16(memmap.Int8(data, record)) {
			t.Fatalf("region %d = %q with %d vertices, want %q with %d", i, got, count, name, memmap.Int8(data, record))
		}
	}
	sub_49FDB0(0)
	if duplicate := memmap.Uint32(world, polygonBase+18*polygonSize+84); duplicate != 0 {
		t.Fatal("already loaded world regions were appended again")
	}
}
