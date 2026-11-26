package legacy

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
)

func TestLegacyMemoryPointerPreservesBlobOrder(t *testing.T) {
	for _, base := range legacyMemoryBlobOrder {
		blob := memmap.BlobByAddr(base)
		address := unsafe.Pointer(&blob.Data[0])
		if got := legacyMemoryPointer(base, 0, 1); got != address {
			t.Fatalf("blob %#x starts at %p, want %p", base, got, address)
		}
		end := uintptr(len(blob.Data))
		if ptr, size := cfixture.MemmapNativeBlobStorage(base); ptr != address || size != end {
			t.Fatalf("blob %#x storage = %p/%d, want %p/%d", base, ptr, size, address, end)
		}
		if got := legacyMemoryPointer(base, end, 0); got != unsafe.Add(address, end) {
			t.Fatalf("blob %#x zero-length end = %p", base, got)
		}
	}
	for _, base := range []uintptr{0, 0x581451} {
		if ptr, size := cfixture.MemmapNativeBlobStorage(base); ptr != nil || size != 0 {
			t.Fatalf("blob storage at %#x = %p/%d, want nil/0", base, ptr, size)
		}
	}

	offset := uintptr(len(memmap.BlobByAddr(0x581450).Data))
	expect := memmap.PtrOff(0x587000, offset)
	if got := legacyMemoryPointer(0x581450, offset, 1); got != expect {
		t.Fatalf("fallthrough pointer = %p, want %p", got, expect)
	}

	for _, tc := range []struct {
		base, off, size uintptr
	}{
		{0x581451, 0, 1},
		{0x581450, ^uintptr(0), 2},
		{0x973A20, 0, ^uintptr(0)},
	} {
		if got := legacyMemoryPointer(tc.base, tc.off, tc.size); got != nil {
			t.Fatalf("out-of-bounds %#x+%d[%d] = %p", tc.base, tc.off, tc.size, got)
		}
	}
}

func TestLegacyMemoryNativeAccessors(t *testing.T) {
	const base, off = 0x581450, 16
	want := uintptr(memmap.PtrOff(base, off))
	for i, got := range cfixture.MemmapNativeAccessors(base, off) {
		if got != want {
			t.Errorf("accessor %d returned %#x, want %#x", i, got, want)
		}
	}
}

func TestLegacyMemoryNativeAccessorsAbort(t *testing.T) {
	const subprocess = "OPENNOX_TEST_MEMMAP_BAD_ADDRESS"
	if os.Getenv(subprocess) == "1" {
		cfixture.MemmapNativeAccessors(0, 0)
		t.Fatal("invalid memory access returned")
	}

	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(exe, "-test.run=^TestLegacyMemoryNativeAccessorsAbort$")
	cmd.Env = append(os.Environ(), subprocess+"=1")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("invalid memory access did not abort")
	}
	if !bytes.Contains(output, []byte("Invalid memory access! Requested = 0+0[4]\n")) &&
		!bytes.Contains(output, []byte("no blobs matching the address: 0x0\n")) {
		t.Fatalf("missing invalid memory access diagnostic (%v):\n%s", err, output)
	}
}
