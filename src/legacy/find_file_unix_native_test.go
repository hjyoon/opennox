//go:build !windows

package legacy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestFindFileNativeEnumeration(t *testing.T) {
	dir := t.TempDir()
	for name, data := range map[string]string{"b.plr": "beta", "a.plr": "alpha", ".hidden.plr": "hidden"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "c.plr"), 0700); err != nil {
		t.Fatal(err)
	}

	pattern := CString(strings.ReplaceAll(filepath.Join(dir, "*.plr"), "/", "\\"))
	defer StrFree(pattern)
	var first, second WIN32_FIND_DATAA
	h0 := FindFirstFileA((*byte)(unsafe.Pointer(pattern)), &first)
	if h0 == ^uintptr(0) {
		t.Fatal("no matching player files")
	}
	defer FindClose(h0)
	h1 := FindFirstFileA((*byte)(unsafe.Pointer(pattern)), &second)
	if h1 == ^uintptr(0) || h0 == h1 {
		t.Fatalf("expected independent search handles, got %#x and %#x", h0, h1)
	}
	defer FindClose(h1)

	for _, result := range []*WIN32_FIND_DATAA{&first, &second} {
		if name := alloc.GoStringS(result.cFileName[:]); name != "a.plr" ||
			result.nFileSizeLow != 5 || result.nFileSizeHigh != 0 || result.FileAttributes != 0x80 {
			t.Fatalf("unexpected first result %+v", *result)
		}
	}
	if FindNextFileA(h0, &first) == 0 || alloc.GoStringS(first.cFileName[:]) != "b.plr" || first.nFileSizeLow != 4 {
		t.Fatalf("unexpected second file %+v", first)
	}
	if FindNextFileA(h0, &first) == 0 || alloc.GoStringS(first.cFileName[:]) != "c.plr" || first.FileAttributes != 0x10 {
		t.Fatalf("unexpected directory result %+v", first)
	}
	if FindNextFileA(h0, &first) != 0 {
		t.Fatal("search included a hidden file")
	}
	if FindNextFileA(h1, &second) == 0 || alloc.GoStringS(second.cFileName[:]) != "b.plr" {
		t.Fatalf("advancing the first iterator changed the second: %+v", second)
	}

	missing := CString(filepath.Join(dir, "*.missing"))
	defer StrFree(missing)
	if handle := FindFirstFileA((*byte)(unsafe.Pointer(missing)), &first); handle != ^uintptr(0) {
		FindClose(handle)
		t.Fatalf("empty search returned handle %#x", handle)
	}
}
