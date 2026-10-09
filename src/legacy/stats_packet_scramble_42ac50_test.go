package legacy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestStatsPacketScramble42AC50NativeBody(t *testing.T) {
	fixture, err := os.ReadFile("testdata/stats_packet_scramble_42ac50_test.c")
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("GAME1_2.c")
	if err != nil {
		t.Fatal(err)
	}
	const signature = "uint8_t* sub_42AC50(uint8_t* a1, size_t* a2) {"
	const marker = "// PRODUCTION_SCRAMBLE"
	if strings.Count(string(source), signature) != 1 || strings.Count(string(fixture), marker) != 1 {
		t.Fatal("scrambler signature or fixture marker is not unique")
	}
	_, body, _ := strings.Cut(string(source), signature)
	body, _, ok := strings.Cut(body, "\n//----- (0042ADA0)")
	if !ok || !strings.HasSuffix(strings.TrimSpace(body), "}") {
		t.Fatal("scrambler body boundary is missing")
	}
	program := strings.Replace(string(fixture), marker, signature+body, 1)
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	binary := filepath.Join(t.TempDir(), "stats-packet-scramble")
	args := append(compiler[1:], "-std=c11", "-O2", "-Wall", "-Wextra", "-Werror",
		"-x", "c", "-", "-o", binary)
	args = append(args, strings.Fields(os.Getenv("NOX_STATS_PACKET_CFLAGS"))...)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = strings.NewReader(program)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile actual scrambler body: %v\n%s", err, output)
	}
	if output, err := exec.Command(binary).CombinedOutput(); err != nil {
		t.Fatalf("native scrambler contract: %v\n%s", err, output)
	} else {
		t.Log(strings.TrimSpace(string(output)))
	}
	if receipt := os.Getenv("NOX_STATS_PACKET_NATIVE_JSON"); receipt != "" {
		output, err := exec.Command(binary, "--json").Output()
		if err != nil {
			t.Fatalf("record native scrambler cases: %v", err)
		}
		if err := os.WriteFile(receipt, output, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
