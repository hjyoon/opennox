package legacy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/opennox/libs/datapath"
)

func TestMapgenGenerationRetries(t *testing.T) {
	for _, tc := range []struct {
		name      string
		terminal  int
		attempt   int
		want      bool
		wantCalls int
	}{
		{"immediate success", 1, 1, true, 1},
		{"hard failure stops retries", 0, 5, false, 5},
		{"last attempt succeeds", 1, 100, true, 100},
		{"attempt limit", 1, 101, false, 100},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			ok := mapgenGenerateWithRetries(func() int {
				calls++
				if calls == tc.attempt {
					return tc.terminal
				}
				return 2
			})
			if ok != tc.want || calls != tc.wantCalls {
				t.Fatalf("generation result = %t after %d attempts", ok, calls)
			}
		})
	}
}

func TestMapgenScratchObjectFiles(t *testing.T) {
	old := datapath.Data()
	dir := t.TempDir()
	datapath.SetData(dir)
	t.Cleanup(func() { datapath.SetData(old) })
	if !mapgenMoveScratchObjects(true) || mapgenMoveScratchObjects(false) {
		t.Fatal("optional and required scratch object file behavior differs")
	}
	previous := filepath.Join(dir, "nc.obj")
	blend := filepath.Join(dir, "blend.obj")
	if err := os.WriteFile(previous, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(blend, []byte("new"), 0600); err != nil {
		t.Fatal(err)
	}
	if !mapgenMoveScratchObjects(false) {
		t.Fatal("required object stream rename failed")
	}
	if data, err := os.ReadFile(previous); err != nil || string(data) != "new" {
		t.Fatalf("renamed stream = %q (%v)", data, err)
	}
	if _, err := os.Stat(blend); !os.IsNotExist(err) {
		t.Fatalf("old stream name remains, stat returned %v", err)
	}
	if !mapgenMoveScratchObjects(true) {
		t.Fatal("optional stream removal failed")
	}
	if _, err := os.Stat(previous); !os.IsNotExist(err) {
		t.Fatalf("temporary object file remains, stat returned %v", err)
	}
}
