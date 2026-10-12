//go:build !server

package opennox

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opennox/libs/maps"
)

// Execute the actual runner's filesystem-preparation block, without replacing
// its map-copy/link operations. Only its cleanup is replaced so evidence stays
// inside the test-owned directory. This is not a game/Windows runtime oracle.
func TestHeadlessRuntimeMapsCanBeDownloaded(t *testing.T) {
	raw, err := os.ReadFile("../scripts/run-headless-gui-e2e.sh")
	if err != nil {
		t.Fatal(err)
	}
	script := string(raw)
	start := strings.Index(script, "temp_root=")
	end := strings.Index(script, "\ncd \"$runtime_data_dir\"")
	if start < 0 || end <= start || strings.Count(script, "rm -rf -- \"$runtime_data_dir\"") != 1 {
		t.Fatal("runner preparation/cleanup boundaries changed")
	}
	preparation := script[start:end]
	preparation = strings.Replace(preparation, "rm -rf -- \"$runtime_data_dir\"", "echo \"preserved isolated runtime: $runtime_data_dir\"", 1)
	for _, name := range []string{"Estate", "G_Castle", "Map With Space"} {
		for _, nested := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/nested-%t", name, nested), func(t *testing.T) {
				task := t.TempDir()
				source := filepath.Join(task, "original data")
				mapDir := filepath.Join(source, "maps", name)
				if err := os.MkdirAll(mapDir, 0755); err != nil {
					t.Fatal(err)
				}
				mapBody := []byte("opaque original map bytes: " + name)
				rules := []byte("original rules: " + name)
				expected := map[string][]byte{
					strings.ToLower(name + ".map"): mapBody,
					strings.ToLower(name + ".rul"): rules,
				}
				sourceFiles := map[string][]byte{
					filepath.Join(name, name+".map"): mapBody,
					filepath.Join(name, name+".rul"): rules,
					filepath.Join(name, name+".nxz"): []byte("original compressed map"),
					filepath.Join(name, "UsEr.RuL"):  []byte("private user rules must not be imported"),
				}
				if nested {
					sourceFiles[filepath.Join(name, "code", "helper.lua")] = []byte("return true")
					expected["code/helper.lua"] = []byte("return true")
				}
				for path, body := range sourceFiles {
					path = filepath.Join(source, "maps", path)
					if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(path, body, 0644); err != nil {
						t.Fatal(err)
					}
				}
				if err := os.MkdirAll(filepath.Join(source, "Save"), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(source, "Save", "personal.plr"), []byte("private save"), 0644); err != nil {
					t.Fatal(err)
				}
				for _, file := range []string{"nox.cfg", "opennox.yml", "nc.obj", "thing.bin"} {
					if err := os.WriteFile(filepath.Join(source, file), []byte("original "+file), 0644); err != nil {
						t.Fatal(err)
					}
				}
				cmd := exec.Command("bash", "-s", "--", source)
				cmd.Env = append(os.Environ(), "TMPDIR="+task)
				cmd.Stdin = strings.NewReader("#!/usr/bin/env bash\nset -euo pipefail\ndata_dir=\"$1\"\n" + preparation + "\n")
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("actual runner preparation: %v\n%s", err, out)
				}
				const prefix = "preserved isolated runtime: "
				runtimeData := strings.TrimSpace(strings.TrimPrefix(string(out), prefix))
				if !strings.HasPrefix(string(out), prefix) || !strings.HasPrefix(runtimeData, task+string(filepath.Separator)) {
					t.Fatalf("unexpected runtime path: %q", out)
				}
				// The actual map server intentionally excludes symbolic links.
				srv := maps.NewServer(slog.New(slog.NewTextHandler(io.Discard, nil)), filepath.Join(runtimeData, "maps"))
				req := httptest.NewRequest("GET", "/api/v0/maps/"+strings.ToLower(strings.ReplaceAll(name, " ", "%20"))+"/download", nil)
				req.Header.Set("Accept", "application/zip")
				rec := httptest.NewRecorder()
				srv.ServeHTTP(rec, req)
				if rec.Code != 200 {
					t.Fatalf("actual map server response=%d", rec.Code)
				}
				zr, err := zip.NewReader(bytes.NewReader(rec.Body.Bytes()), int64(rec.Body.Len()))
				if err != nil {
					t.Fatal(err)
				}
				got := make(map[string][]byte)
				for _, f := range zr.File {
					r, err := f.Open()
					if err != nil {
						t.Fatal(err)
					}
					b, err := io.ReadAll(r)
					_ = r.Close()
					if err != nil {
						t.Fatal(err)
					}
					got[f.Name] = b
				}
				if len(got) != len(expected) {
					t.Fatalf("isolated runtime ZIP has %d files, want %d; linked map files are excluded by production map server", len(got), len(expected))
				}
				for name, want := range expected {
					if !bytes.Equal(got[name], want) {
						t.Errorf("actual ZIP file %q bytes differ", name)
					}
				}
				for path, body := range sourceFiles {
					got, err := os.ReadFile(filepath.Join(source, "maps", path))
					if err != nil || !bytes.Equal(got, body) {
						t.Fatal("runner modified source map assets")
					}
				}
				saves, err := os.ReadDir(filepath.Join(runtimeData, "Save"))
				if err != nil || len(saves) != 0 {
					t.Fatal("runner imported personal Save")
				}
				for _, private := range []string{"nox.cfg", "opennox.yml", "nc.obj", filepath.Join("maps", name, "UsEr.RuL")} {
					if _, err := os.Lstat(filepath.Join(runtimeData, private)); !os.IsNotExist(err) {
						t.Fatalf("runner imported private configuration: %s", private)
					}
				}
				// Modifying a runtime map must never modify the input asset.
				runtimeMap := filepath.Join(runtimeData, "maps", name, name+".map")
				fi, err := os.Lstat(runtimeMap)
				if err != nil || !fi.Mode().IsRegular() {
					t.Fatal("isolated map is not a regular private copy")
				}
				if err := os.WriteFile(runtimeMap, []byte("runtime save"), 0644); err != nil {
					t.Fatal(err)
				}
				gotMap, err := os.ReadFile(filepath.Join(mapDir, name+".map"))
				if err != nil || !bytes.Equal(gotMap, mapBody) {
					t.Fatal("runtime writes reached source map")
				}
			})
		}
	}
}
