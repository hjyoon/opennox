package opennox

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Execute the actual launch subshell with an argv-only executable. This is
// a shell argument/quoting check; the official server E2E separately checks
// the real engine's maps, counters, authenticated commands and process exit.
func TestHeadlessServerRunnerDisablesDefaultCLIMapLoad(t *testing.T) {
	source, err := os.ReadFile("../scripts/run-headless-server-e2e.sh")
	if err != nil {
		t.Fatal(err)
	}
	begin := "(\n\tcd \"$runtime_data_dir\"\n"
	end := "\n) >\"$server_log\" 2>&1 &"
	text := string(source)
	if strings.Count(text, begin) != 1 || strings.Count(text, end) != 1 {
		t.Fatal("real server launch block is missing or ambiguous")
	}
	start := strings.Index(text, begin)
	finish := strings.Index(text, end)
	if finish <= start {
		t.Fatal("invalid server launch block")
	}
	body := text[start:finish] + "\n)\n"
	dir := t.TempDir()
	data := filepath.Join(dir, "data with spaces")
	if err := os.Mkdir(data, 0700); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(dir, "argv-only server")
	if err := os.WriteFile(binary, []byte("#!/bin/sh\nprintf '%s\\000' \"$@\"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("bash", "-c", body)
	cmd.Env = append(os.Environ(),
		"runtime_data_dir="+data, "server_binary="+binary,
		"server_port=52940", "metrics_port=52941", "script_dir="+dir,
		"test_token=unit-only-token")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("production launch block: %v", err)
	}
	parts := bytes.Split(output, []byte{0})
	if len(parts) == 0 || len(parts[len(parts)-1]) != 0 {
		t.Fatal("argv capture did not preserve argument boundaries")
	}
	var got []string
	for _, part := range parts[:len(parts)-1] {
		got = append(got, string(part))
	}
	want := []string{"-config", filepath.Join(data, "opennox.yml"), "-data", data,
		"-port", "52940", "-autoexec", "", "-pprof", "127.0.0.1:52941", "-noaudio"}
	if len(got) == len(want) {
		// Bash appends /opennox.yml to a native Windows directory. Compare
		// only path spellings canonically; all flags and argv boundaries stay exact.
		got[1], got[3] = filepath.Clean(got[1]), filepath.Clean(got[3])
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("isolated map-transition launch argv=%q, want %q; CLI startup load must be explicitly disabled", got, want)
	}
}
