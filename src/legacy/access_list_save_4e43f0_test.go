package legacy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestAccessListSaveC4E43F0StockBytes(t *testing.T) {
	banned, freeBanned := alloc.NewHigh(accessListNode416720{})
	defer freeBanned()
	allowed, freeAllowed := alloc.NewHigh(accessListNode416720{})
	defer freeAllowed()
	accessListNativeClear416720(banned)
	accessListNativeClear416720(allowed)
	defer pushAccessListNativeHooks416720(&accessListNativeHooks416720{
		Allowed: allowed, Banned: banned,
	})()

	for _, value := range []struct {
		name    [26]uint16
		expires uint64
		serial  string
	}{
		{name: [26]uint16{'C', 'a', 'f', 0xe9}},
		{name: [26]uint16{'X'}, expires: 1 << 32},
		{name: [26]uint16{'A', 0x20ac, 'B'}, serial: "1234-abcd"},
	} {
		entry, free := alloc.NewHigh(accessBanned416720{})
		defer free()
		*entry = accessBanned416720{Name: value.name, Expires: value.expires}
		copy(entry.Serial[:], value.serial)
		accessListNativeAppend416720(banned, &entry.List)
	}
	entry, freeEntry := alloc.NewHigh(accessAllowed416720{})
	defer freeEntry()
	*entry = accessAllowed416720{Name: [26]uint16{'B', 'o', 'b'}}
	accessListNativeAppend416720(allowed, &entry.List)

	dir := t.TempDir()
	path := filepath.Join(dir, "ban.txt")
	if got := accessListSaveC4E43F0(path); got != 1 {
		t.Fatalf("successful save returned %d", got)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	const want = "[Banned Users]\nCaf\xe9\n0\nA\n1234-abcd\n\n[Allowed Users]\nBob\n"
	if string(data) != want {
		t.Fatalf("saved bytes = %q, want %q", data, want)
	}
	if got := accessListSaveC4E43F0(filepath.Join(dir, "missing", "ban.txt")); got != 0 {
		t.Fatalf("failed open returned %d", got)
	}
}
