//go:build darwin || linux

package legacy

import (
	"os"
	"strings"
	"testing"
	"unsafe"

	"golang.org/x/sys/unix"
)

func TestMapgenBoundedCStringStopsAtGuardPage(t *testing.T) {
	pageSize := os.Getpagesize()
	region, err := unix.Mmap(-1, 0, 2*pageSize, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_PRIVATE|unix.MAP_ANON)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := unix.Munmap(region); err != nil {
			t.Error(err)
		}
	})
	if err := unix.Mprotect(region[pageSize:], unix.PROT_NONE); err != nil {
		t.Fatal(err)
	}

	short := region[pageSize-4 : pageSize]
	copy(short, "map\x00")
	if value, ok := mapgenBoundedCString(unsafe.Pointer(&short[0])); !ok || value != "map" {
		t.Fatalf("short path: %q, %t", value, ok)
	}

	full := region[pageSize-mapgenPathBufferSize503140 : pageSize]
	want := strings.Repeat("x", len(full))
	copy(full, want)
	if value, ok := mapgenBoundedCString(unsafe.Pointer(&full[0])); ok || value != "" {
		t.Fatalf("unterminated path accepted: length %d", len(value))
	}
	full[len(full)-1] = 0
	if value, ok := mapgenBoundedCString(unsafe.Pointer(&full[0])); !ok || value != want[:len(want)-1] {
		t.Fatalf("bounded path: length %d, %t", len(value), ok)
	}
}
