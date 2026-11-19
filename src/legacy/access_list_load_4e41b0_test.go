package legacy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

func TestAccessListLoadC4E41B0StockText(t *testing.T) {
	banned, doneBanned := alloc.NewHigh(accessListNode416720{})
	allowed, doneAllowed := alloc.NewHigh(accessListNode416720{})
	t.Cleanup(doneBanned)
	t.Cleanup(doneAllowed)
	accessListNativeClear416720(banned)
	accessListNativeClear416720(allowed)
	t.Cleanup(pushAccessListNativeHooks416720(&accessListNativeHooks416720{
		Banned:  banned,
		Allowed: allowed,
		Refresh: func() unsafe.Pointer { return nil },
		Allocate: func(size uintptr) unsafe.Pointer {
			p, free := alloc.MallocHigh(size)
			t.Cleanup(free)
			clear(unsafe.Slice((*byte)(p), size))
			return p
		},
	}))
	dir := t.TempDir()
	path := filepath.Join(dir, "ban.txt")
	input := strings.Repeat("#", 1023) + "[Banned Users]\r\n\tCaf\xe9\textra\r\n0\r\n\r\n[Allowed Users]\r\nSh\xe9"
	if err := os.WriteFile(path, []byte(input), 0600); err != nil {
		t.Fatal(err)
	}
	if got := accessListLoadC4E41B0(path); got != 1 {
		t.Fatalf("load returned %d", got)
	}
	bannedEntry := (*accessBanned416720)(sub416900Native())
	allowedEntry := (*accessAllowed416720)(sub4168E0Native())
	if bannedEntry == nil || bannedEntry.Name != ([26]uint16{'C', 'a', 'f', 0xe9}) ||
		bannedEntry.Serial[0] != 0 || bannedEntry.Expires != 0 ||
		allowedEntry == nil || allowedEntry.Name != ([26]uint16{'S', 'h', 0xe9}) {
		t.Fatalf("unexpected access entries %v, %v", bannedEntry, allowedEntry)
	}
	if got := accessListLoadC4E41B0(filepath.Join(dir, "missing")); got != 0 {
		t.Fatalf("failed open returned %d", got)
	}

	handles.InitHigh()
	t.Cleanup(handles.Release)
	if err := os.WriteFile(path, []byte("unused-name\n\t\nallowed-name"), 0600); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	file := NewFileHandle(binfile.NewTextFile(f))
	t.Cleanup(func() { nox_fs_close(file) })
	if got := accessListLoadSectionC4E41B0(file, true); got != 0 {
		t.Fatalf("blank serial returned %d", got)
	}
	if got := accessListLoadSectionC4E41B0(file, false); got != 1 {
		t.Fatalf("allowed section returned %d", got)
	}
	last := (*accessAllowed416720)(sub4168F0Native(unsafe.Pointer(allowedEntry)))
	if last == nil || GoWStringSlice(last.Name[:12]) != "allowed-name" ||
		sub416910Native(unsafe.Pointer(bannedEntry)) != nil {
		t.Fatal("section calls did not preserve the file position and existing entries")
	}
}
