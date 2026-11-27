//go:build windows

package legacy

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestFindFileNativeWindowsANSI(t *testing.T) {
	dir := t.TempDir()
	for name, data := range map[string]string{"a.plr": "alpha", "b.plr": "second"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	pattern, err := syscall.BytePtrFromString(filepath.Join(dir, "*.plr"))
	if err != nil {
		t.Fatal(err)
	}
	var entry WIN32_FIND_DATAA
	handle := FindFirstFileA(pattern, &entry)
	if handle == invalidFindFileHandle || handle == 0 {
		t.Fatalf("FindFirstFileA returned %#x", handle)
	}
	defer func() {
		if handle != invalidFindFileHandle {
			FindClose(handle)
		}
	}()

	files := make(map[string]uint32)
	for {
		name := alloc.GoStringS(entry.cFileName[:])
		if _, seen := files[name]; seen {
			t.Fatalf("find returned %q more than once", name)
		}
		files[name] = entry.nFileSizeLow
		if entry.FileAttributes&0x10 != 0 || entry.nFileSizeHigh != 0 {
			t.Fatalf("unexpected find record %+v", entry)
		}
		if FindNextFileA(handle, &entry) == 0 {
			break
		}
	}
	if len(files) != 2 || files["a.plr"] != 5 || files["b.plr"] != 6 {
		t.Fatalf("find results = %v", files)
	}
	if closeFileSearch(handle) == 0 {
		t.Fatal("FindClose rejected a live handle")
	}
	handle = invalidFindFileHandle
	missing, err := syscall.BytePtrFromString(filepath.Join(dir, "*.missing"))
	if err != nil {
		t.Fatal(err)
	}
	if got := FindFirstFileA(missing, &entry); got != invalidFindFileHandle {
		FindClose(got)
		t.Fatalf("missing search returned %#x", got)
	}
}
