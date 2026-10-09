package legacy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestStatsPacketRLE42A970ByteComparisons(t *testing.T) {
	fixture, err := os.ReadFile("testdata/stats_packet_rle_42a970_test.c")
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("GAME1_2.c")
	if err != nil {
		t.Fatal(err)
	}
	const signature = "int sub_42A970(uint8_t* a1, uint8_t* a2, int* a3) {"
	const marker = "// PRODUCTION_COMPRESSION"
	if strings.Count(string(source), signature) != 1 || strings.Count(string(fixture), marker) != 1 {
		t.Fatal("compression signature or fixture marker is not unique")
	}
	_, body, _ := strings.Cut(string(source), signature)
	body, _, ok := strings.Cut(body, "\n// 42A970:")
	if !ok || !strings.HasSuffix(strings.TrimSpace(body), "}") {
		t.Fatal("compression body boundary is missing")
	}
	program := strings.Replace(string(fixture), marker, signature+body, 1)
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	binary := filepath.Join(t.TempDir(), "stats-packet-rle")
	args := append(compiler[1:], "-std=c11", "-fsigned-char", "-O2", "-Wall", "-Wextra", "-Werror", "-x", "c", "-", "-o", binary)
	args = append(args, strings.Fields(os.Getenv("NOX_STATS_ENVELOPE_CFLAGS"))...)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = strings.NewReader(program)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile actual compression body: %v\n%s", err, output)
	}
	if output, err := exec.Command(binary).CombinedOutput(); err != nil {
		t.Fatalf("native BYTE comparison contract: %v\n%s", err, output)
	} else {
		t.Log(strings.TrimSpace(string(output)))
	}
	if receipt := os.Getenv("NOX_STATS_RLE_NATIVE_JSON"); receipt != "" {
		output, err := exec.Command(binary, "--json").Output()
		if err != nil {
			t.Fatalf("record native BYTE cases: %v", err)
		}
		if err := os.WriteFile(receipt, output, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
