package legacy

import (
	"bytes"
	"encoding/binary"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func TestStatsBulk4263C0NativeBody(t *testing.T) {
	fixture, err := os.ReadFile("testdata/stats_bulk_native_4263c0_test.c")
	if err != nil {
		t.Fatal(err)
	}
	header, err := os.ReadFile("stats_bulk_native.h")
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("stats_bulk_native.c")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(source), "// NATIVE_STATS_IMPLEMENTATION") != 1 {
		t.Fatal("native implementation boundary is not unique")
	}
	_, implementation, _ := strings.Cut(string(source), "// NATIVE_STATS_IMPLEMENTATION")
	root, err := os.ReadFile("GAME1_1.c")
	if err != nil {
		t.Fatal(err)
	}
	signature := "char* nox_xxx_net_4263C0() {"
	if strings.Count(string(root), signature) != 1 {
		t.Fatal("existing production root signature is not unique")
	}
	_, body, _ := strings.Cut(string(root), signature)
	body, _, ok := strings.Cut(body, "\n//----- (004264D0)")
	if !ok || !strings.HasSuffix(strings.TrimSpace(body), "}") {
		t.Fatal("existing production root boundary is missing")
	}
	program := strings.Replace(string(fixture), "// PRODUCTION_HEADER", string(header), 1)
	program = strings.Replace(program, "// PRODUCTION_NATIVE_BODY", implementation, 1)
	program = strings.Replace(program, "// PRODUCTION_ROOT", signature+body, 1)
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	binary := filepath.Join(t.TempDir(), "stats-bulk")
	args := append(compiler[1:], "-std=c11", "-O2", "-Wall", "-Wextra", "-Werror", "-x", "c", "-", "-o", binary)
	args = append(args, strings.Fields(os.Getenv("NOX_STATS_BULK_CFLAGS"))...)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = strings.NewReader(program)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile actual native report helpers and original C entry: %v\n%s", err, output)
	}
	for _, name := range []string{"packet", "flush", "lifecycle"} {
		t.Run(name, func(t *testing.T) {
			if output, err := exec.Command(binary, "--"+name).CombinedOutput(); err != nil {
				t.Fatalf("native report %s contract: %v\n%s", name, err, output)
			} else {
				t.Log(strings.TrimSpace(string(output)))
			}
			if receipt := os.Getenv("NOX_STATS_BULK_NATIVE_JSON"); receipt != "" {
				output, err := exec.Command(binary, "--"+name, "--json").Output()
				if err != nil {
					t.Fatalf("record native report cases: %v", err)
				}
				if err := os.WriteFile(receipt+"."+name+".jsonl", output, 0600); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
	for _, fault := range []struct {
		name   string
		signal syscall.Signal
		events string
	}{
		{"names_alloc", syscall.SIGSEGV, "AA"},
		{"pair_alloc", syscall.SIGSEGV, "A"},
		{"node_alloc", syscall.SIGSEGV, "N"},
		{"data_alloc", syscall.SIGSEGV, "NZD"},
		{"packet_alloc", syscall.SIGSEGV, strings.Repeat("NZD", 31) + "P"},
		{"nil_length", syscall.SIGSEGV, strings.Repeat("NZD", 31)},
		{"transport", syscall.SIGABRT, strings.Repeat("NZD", 31) + "P" + strings.Repeat("FF", 31) + "EFB"},
	} {
		t.Run(fault.name, func(t *testing.T) {
			output, err := exec.Command(binary, "--fault="+fault.name).CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok {
				t.Fatalf("fault boundary became success: err=%v output=%q", err, output)
			}
			status, ok := exit.Sys().(syscall.WaitStatus)
			if !ok || !status.Signaled() || status.Signal() != fault.signal {
				t.Fatalf("want native %v, got %v output=%q", fault.signal, exit, output)
			}
			if string(output) != fault.events {
				t.Fatalf("original fault service prefix changed: got %q, want %q", output, fault.events)
			}
			if receipt := os.Getenv("NOX_STATS_BULK_NATIVE_JSON"); receipt != "" {
				if err := os.WriteFile(receipt+"."+fault.name+".events", output, 0600); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

// Exercise the real no-argument Go -> C -> Go player-list callback chain and
// the production compressor/scrambler/envelope, not the standalone E service.
// A child process isolates the persistent C columns and all memmap mutations.
func TestStatsBulk4263C0CgoEntry(t *testing.T) {
	if os.Getenv("NOX_STATS_BULK_CGO_CHILD") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestStatsBulk4263C0CgoEntry$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), "NOX_STATS_BULK_CGO_CHILD=1")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("real CGo bulk flush: %v\n%s", err, output)
		}
		return
	}
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameOnline)
	srv := &server.Server{}
	GetServer = func() Server { return &playerTrackingLegacyServer425F10{srv: srv} }
	header := memmap.PtrT[[640]byte](0x5D4594, 599476)
	records := memmap.PtrT[[8192]byte](0x5D4594, 600124)
	pairs := memmap.PtrT[[131072]byte](0x5D4594, 608320)
	*header = [640]byte{}
	for cycle, count := range []uint32{2, 0, 256, 1, 0} {
		for i := range records {
			records[i] = 0xa5
		}
		for i := range pairs {
			pairs[i] = byte(i*37 + cycle)
		}
		for i := uint32(0); i < count; i++ {
			record := records[i*32 : (i+1)*32]
			copy(record[:10], "ArmReport\x00")
			binary.LittleEndian.PutUint32(record[12:16], 0x80123456+i)
			binary.LittleEndian.PutUint32(record[16:20], 0xabcdef00+i)
			record[20], record[21], record[28] = byte(i+0x80), 0xfe, 0x81
			binary.LittleEndian.PutUint32(record[24:28], 0)
		}
		Set_dword_5d4594_608316(count)
		playerKillStatsSetPairCount425CA0(uint32(cycle))
		Nox_xxx_net_4263C0()
		if got := binary.LittleEndian.Uint16(header[6:8]); got != uint16(count) {
			t.Fatalf("cycle %d header count = %d, want %d", cycle, got, count)
		}
		if Get_dword_5d4594_608316() != 0 || playerKillStatsPairCount425CA0() != 0 {
			t.Fatalf("cycle %d did not reset scalar counts", cycle)
		}
		if !bytes.Equal(records[:], make([]byte, len(records))) || !bytes.Equal(pairs[:], make([]byte, len(pairs))) {
			t.Fatalf("cycle %d did not clear exact original record/pair buffers", cycle)
		}
		if !bytes.Equal(header[608:], make([]byte, 32)) {
			t.Fatalf("cycle %d wrote a pointer into the eight ABI32 header slots", cycle)
		}
	}
}
