package legacy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestStatsPacketEnvelope42A8B0NativeBody(t *testing.T) {
	fixture, err := os.ReadFile("testdata/stats_packet_envelope_42a8b0_test.c")
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("GAME1_2.c")
	if err != nil {
		t.Fatal(err)
	}
	program := string(fixture)
	for _, part := range []struct{ signature, end, marker string }{
		{"uint16_t* sub_42A8B0(uint8_t* a1, int* a2) {", "\n//----- (0042A970)", "// PRODUCTION_ENVELOPE"},
		{"int sub_42A970(uint8_t* a1, uint8_t* a2, int* a3) {", "\n// 42A970:", "// PRODUCTION_COMPRESSION"},
		{"uint8_t* sub_42AC50(uint8_t* a1, size_t* a2) {", "\n//----- (0042ADA0)", "// PRODUCTION_SCRAMBLE"},
	} {
		if strings.Count(string(source), part.signature) != 1 || strings.Count(program, part.marker) != 1 {
			t.Fatal("production signature or fixture marker is not unique: " + part.signature)
		}
		_, body, _ := strings.Cut(string(source), part.signature)
		body, _, ok := strings.Cut(body, part.end)
		if !ok || !strings.HasSuffix(strings.TrimSpace(body), "}") {
			t.Fatal("production body boundary is missing: " + part.signature)
		}
		program = strings.Replace(program, part.marker, part.signature+body, 1)
	}
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	binary := filepath.Join(t.TempDir(), "stats-packet-envelope")
	args := append(compiler[1:], "-std=c11", "-O2", "-Wall", "-Wextra", "-Werror", "-x", "c", "-", "-o", binary)
	args = append(args, strings.Fields(os.Getenv("NOX_STATS_ENVELOPE_CFLAGS"))...)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = strings.NewReader(program)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile actual packet/compression/scrambler bodies: %v\n%s", err, output)
	}
	t.Run("native_packet_and_compression_chain", func(t *testing.T) {
		if output, err := exec.Command(binary).CombinedOutput(); err != nil {
			t.Fatalf("native envelope contract: %v\n%s", err, output)
		} else {
			t.Log(strings.TrimSpace(string(output)))
		}
		if receipt := os.Getenv("NOX_STATS_ENVELOPE_NATIVE_JSON"); receipt != "" {
			output, err := exec.Command(binary, "--json").Output()
			if err != nil {
				t.Fatalf("record native envelope cases: %v", err)
			}
			if err := os.WriteFile(receipt, output, 0600); err != nil {
				t.Fatal(err)
			}
		}
	})
	for _, fault := range []struct{ name, events string }{
		{"nil_length", ""}, {"nil_input", "AC"}, {"compression_alloc", "AC"},
		{"node_alloc", "ACSFN"}, {"data_alloc", "ACSFNZD"}, {"packet_alloc", "ACSFNZDFP"},
	} {
		t.Run(fault.name, func(t *testing.T) {
			output, err := exec.Command(binary, "--fault="+fault.name).CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok {
				t.Fatalf("original fault must not become an early return: err=%v\n%s", err, output)
			}
			status, ok := exit.Sys().(syscall.WaitStatus)
			if !ok || !status.Signaled() || status.Signal() != syscall.SIGSEGV {
				t.Fatalf("expected native SIGSEGV, got %v\n%s", exit, output)
			}
			if string(output) != fault.events {
				t.Fatalf("fault moved across service boundary: events=%q want=%q", output, fault.events)
			}
		})
	}
}
