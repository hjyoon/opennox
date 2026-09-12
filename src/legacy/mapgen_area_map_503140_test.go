package legacy

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"
)

func TestMapgenPrepareAreaMap503140FileOperationOrder(t *testing.T) {
	for _, tc := range []struct {
		name      string
		removeErr error
		renameErr error
		want      bool
		calls     []string
	}{
		{"no old backup", os.ErrNotExist, nil, true, []string{"remove", "rename"}},
		{"old backup removed", nil, nil, true, []string{"remove", "rename"}},
		{"access denied", syscall.EACCES, nil, false, []string{"remove"}},
		{"other remove error", syscall.EIO, nil, true, []string{"remove", "rename"}},
		{"rename failure", nil, errors.New("rename failed"), false, []string{"remove", "rename"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls []string
			got := mapgenPrepareAreaMapWithFS503140("source", "directory", func(path string) error {
				if path != `directory\AreaMap.bak` {
					t.Fatalf("remove path = %q", path)
				}
				calls = append(calls, "remove")
				return tc.removeErr
			}, func(source, backup string) error {
				if source != "source" || backup != `directory\AreaMap.bak` {
					t.Fatalf("rename paths = %q, %q", source, backup)
				}
				calls = append(calls, "rename")
				return tc.renameErr
			})
			if got != tc.want || !reflect.DeepEqual(calls, tc.calls) {
				t.Fatalf("prepare = %t, calls %v; want %t, %v", got, calls, tc.want, tc.calls)
			}
		})
	}
}

func TestMapgenPrepareAreaMap503140ViaC(t *testing.T) {
	directory := t.TempDir()
	source := filepath.Join(directory, "source.tmp")
	backup := filepath.Join(directory, "AreaMap.bak")
	if err := os.WriteFile(source, []byte("new map"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := mapgenPrepareAreaMapViaC503140(source, directory); got != 1 {
		t.Fatalf("first prepare = %d", got)
	}
	if data, err := os.ReadFile(backup); err != nil || string(data) != "new map" {
		t.Fatalf("first backup = %q, %v", data, err)
	}
	if _, err := os.Stat(source); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("source after rename: %v", err)
	}
	if err := os.WriteFile(source, []byte("replacement"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := mapgenPrepareAreaMapViaC503140(source, directory); got != 1 {
		t.Fatalf("replacement prepare = %d", got)
	}
	if data, err := os.ReadFile(backup); err != nil || string(data) != "replacement" {
		t.Fatalf("replacement backup = %q, %v", data, err)
	}
	if got := mapgenPrepareAreaMapViaC503140("", directory); got != 0 {
		t.Fatalf("empty source prepare = %d", got)
	}
}
