package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSealAndVerify(t *testing.T) {
	root := makeTree(t)
	manifestPath := filepath.Join(t.TempDir(), "oracle.json")
	var out bytes.Buffer
	if err := run([]string{"seal", "-root", root, "-out", manifestPath, "-id", "test-copy"}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "sealed test-copy: 2 files") {
		t.Fatalf("unexpected seal output: %q", out.String())
	}
	out.Reset()
	if err := run([]string{"verify", "-root", root, "-manifest", manifestPath}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "verified test-copy: 2 files") {
		t.Fatalf("unexpected verify output: %q", out.String())
	}
	if err := run([]string{"seal", "-root", root, "-out", manifestPath, "-id", "replacement"}, &out); err == nil || !strings.Contains(err.Error(), "refusing to replace") {
		t.Fatalf("seal should refuse replacement, got %v", err)
	}
}

func TestVerifyDetectsEveryChangeClass(t *testing.T) {
	tests := []struct {
		name string
		edit func(t *testing.T, root string)
		want string
	}{
		{
			name: "changed",
			edit: func(t *testing.T, root string) {
				writeFile(t, filepath.Join(root, "GAME.EXE"), "mutated")
			},
			want: "changed=1",
		},
		{
			name: "missing",
			edit: func(t *testing.T, root string) {
				if err := os.Remove(filepath.Join(root, "maps", "demo.map")); err != nil {
					t.Fatal(err)
				}
			},
			want: "missing=1",
		},
		{
			name: "extra",
			edit: func(t *testing.T, root string) {
				writeFile(t, filepath.Join(root, "unexpected.dll"), "extra")
			},
			want: "extra=1",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := makeTree(t)
			manifestPath := filepath.Join(t.TempDir(), "oracle.json")
			if err := run([]string{"seal", "-root", root, "-out", manifestPath, "-id", "test-copy"}, &bytes.Buffer{}); err != nil {
				t.Fatal(err)
			}
			tc.edit(t, root)
			err := run([]string{"verify", "-root", root, "-manifest", manifestPath}, &bytes.Buffer{})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("verify error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestReadManifestRejectsTampering(t *testing.T) {
	root := makeTree(t)
	manifestPath := filepath.Join(t.TempDir(), "oracle.json")
	if err := run([]string{"seal", "-root", root, "-out", manifestPath, "-id", "test-copy"}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	b = bytes.Replace(b, []byte(`"tree_sha256": "`), []byte(`"tree_sha256": "00`), 1)
	if err := os.WriteFile(manifestPath, b, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readManifest(manifestPath); err == nil || !strings.Contains(err.Error(), "tree_sha256 mismatch") {
		t.Fatalf("expected tree digest rejection, got %v", err)
	}
}

func makeTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "GAME.EXE"), "executable")
	writeFile(t, filepath.Join(root, "maps", "demo.map"), "map data")
	return root
}

func writeFile(t *testing.T, name, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
}
