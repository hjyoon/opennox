package legacy

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestStatsEnd4264D0NativeBody(t *testing.T) {
	read := func(path string) string {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	cut := func(source, begin, end string) string {
		t.Helper()
		if strings.Count(source, begin) != 1 {
			t.Fatal("nonunique production boundary: " + begin)
		}
		_, body, _ := strings.Cut(source, begin)
		if end != "" {
			var ok bool
			body, _, ok = strings.Cut(body, end)
			if !ok {
				t.Fatal("missing production end: " + end)
			}
		}
		return body
	}
	services, _, ok := strings.Cut(read("testdata/stats_bulk_native_4263c0_test.c"), "// PRODUCTION_HEADER")
	if !ok {
		t.Fatal("R62 services boundary missing")
	}
	oldPlayer := "typedef struct {\n\tuint8_t prefix[2064];"
	before, tail, ok := strings.Cut(services, oldPlayer)
	if !ok {
		t.Fatal("R62 player fixture boundary missing")
	}
	_, tail, ok = strings.Cut(tail, "} nox_playerInfo;")
	if !ok {
		t.Fatal("R62 player fixture end missing")
	}
	services = before + tail
	services = strings.Replace(services, "events[32768]", "events[131072]", 1)
	for _, signature := range []string{"static void* getMemAt(", "static uint32_t* getMemU32Ptr(", "static uint8_t* getMemU8Ptr(", "static int16_t* getMemI16Ptr("} {
		services = strings.Replace(services, signature, strings.Replace(signature, " getMem", " bulk_getMem", 1), 1)
	}
	services = strings.Replace(services, "static uint16_t* sub_42A8B0(", "static uint16_t* bulk_envelope(", 1)
	program := "#include <stdint.h>\n" + read("testdata/stats_end_native_4264d0_test.c")
	program = strings.Replace(program, "// BULK_FIXTURE_SERVICES", services, 1)
	for _, part := range []struct{ token, path, begin, end string }{
		{"// PRODUCTION_BULK_HEADER", "stats_bulk_native.h", "", ""},
		{"// PRODUCTION_START_HEADER", "stats_start_native.h", "", ""},
		{"// PRODUCTION_END_HEADER", "stats_end_native.h", "", ""},
		{"// PRODUCTION_BULK_BODY", "stats_bulk_native.c", "// NATIVE_STATS_IMPLEMENTATION", ""},
		{"// PRODUCTION_END_BODY", "stats_end_native.c", "// NATIVE_STATS_END_IMPLEMENTATION", ""},
	} {
		text := read(part.path)
		if part.begin != "" {
			text = cut(text, part.begin, part.end)
		}
		if part.path == "stats_end_native.h" {
			text = strings.Replace(text, "#include \"stats_start_native.h\"", "", 1)
		}
		program = strings.Replace(program, part.token, text, 1)
	}
	getter := "nox_stats_quest_columns_native* nox_stats_quest_columns_native_get(void) {"
	program = strings.Replace(program, "// PRODUCTION_GETTER", getter+cut(read("stats_start_native.c"), getter, ""), 1)
	root := "int sub_4264D0() {"
	program = strings.Replace(program, "// PRODUCTION_ROOT", root+cut(read("GAME1_1.c"), root, "\n//----- (00426A30)"), 1)
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	binary := filepath.Join(t.TempDir(), "stats-end")
	args := append(compiler[1:], "-std=c11", "-O2", "-Wall", "-Wextra", "-Werror", "-Wno-unused-function", "-x", "c", "-", "-o", binary)
	args = append(args, strings.Fields(os.Getenv("NOX_STATS_END_CFLAGS"))...)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = strings.NewReader(program)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile actual end/Quest builder/packet: %v\n%s", err, output)
	}
	for _, name := range []string{"root", "builder", "packet"} {
		t.Run(name, func(t *testing.T) {
			if output, err := exec.Command(binary, "--"+name).CombinedOutput(); err != nil {
				t.Fatalf("native end %s: %v\n%s", name, err, output)
			} else {
				t.Log(strings.TrimSpace(string(output)))
			}
			output, err := exec.Command(binary, "--"+name, "--json").Output()
			if err != nil {
				t.Fatal(err)
			}
			// Frozen unchanged PE32 transcripts include every snapshot field,
			// wire byte and service prefix, not only the successful return.
			want := map[string]string{
				"root":    "d63a44a75762a2d1966d8f33a8eddaeb19018fab8741acccf457523c105d9f30",
				"builder": "fd9ae784606430c44b0dfa9a3869452fb92d6b8e1b69a2fee2962f6543268d3a",
				"packet":  "a45cbfbc7bdbc5e75781bce0060303f3d1615bddd6d83afeadd8698a9506240a",
			}[name]
			if got := fmt.Sprintf("%x", sha256.Sum256(output)); got != want {
				t.Fatalf("unchanged PE32 %s transcript changed: %s", name, got)
			}
			if receipt := os.Getenv("NOX_STATS_END_NATIVE_JSON"); receipt != "" {
				if err := os.WriteFile(receipt+"."+name+".jsonl", output, 0600); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
	for _, fault := range []string{"names_alloc", "node_alloc", "data_alloc", "packet_alloc", "transport"} {
		t.Run(fault, func(t *testing.T) {
			output, err := exec.Command(binary, "--fault="+fault).CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok {
				t.Fatalf("original fault became success: %v %q", err, output)
			}
			status, ok := exit.Sys().(syscall.WaitStatus)
			want := syscall.SIGSEGV
			if fault == "transport" {
				want = syscall.SIGABRT
			}
			if !ok || !status.Signaled() || status.Signal() != want {
				t.Fatalf("want %v: %v %q", want, exit, output)
			}
			prefix := "GQTQK" + strings.Repeat("A", 12) + "LUVIII"
			packet := prefix + strings.Repeat("NZD", 20) + "P"
			expected := map[string]string{
				"names_alloc": "GQTQKAA", "node_alloc": prefix + "N", "data_alloc": prefix + "NZD",
				"packet_alloc": packet, "transport": packet + strings.Repeat("FF", 20) + "EFB",
			}[fault]
			if string(output) != expected {
				t.Fatalf("original fault service prefix changed: %q, want %q", output, expected)
			}
			if receipt := os.Getenv("NOX_STATS_END_NATIVE_JSON"); receipt != "" {
				if err := os.WriteFile(receipt+"."+fault+".events", output, 0600); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

// The real CGo entry must not narrow the C-owned memmap header to an int.
// Isolate static C column ownership and use the production envelope, without
// enabling game statistics admission or starting a GUI/network sender.
func TestStatsEnd4264D0CgoEntry(t *testing.T) {
	if os.Getenv("NOX_STATS_END_CGO_CHILD") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestStatsEnd4264D0CgoEntry$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), "NOX_STATS_END_CGO_CHILD=1")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("real CGo statistics end: %v\n%s", err, output)
		}
		return
	}
	srv := server.New(nil, nil, strman.New())
	GetServer = func() Server { return &playerTrackingLegacyServer425F10{srv: srv} }
	noxflags.ResetEngine()
	for _, flags := range []noxflags.GameFlag{0, noxflags.GameOnline, noxflags.GameModeQuest, noxflags.GameOnline | noxflags.GameModeQuest} {
		noxflags.ResetGame()
		noxflags.SetGame(flags)
		header := memmap.PtrT[[640]byte](0x5D4594, 599476)
		quest := memmap.PtrT[[580]byte](0x5D4594, 739396)
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(header)) <= math.MaxUint32 || uintptr(unsafe.Pointer(quest)) <= math.MaxUint32) {
			t.Fatal("actual C-owned headers must be above 4 GiB")
		}
		*header, *quest = [640]byte{}, [580]byte{}
		Sub_4264D0()
		if binary.LittleEndian.Uint16(quest[:2]) != 0 {
			t.Fatal("empty Quest end count changed")
		}
		for _, data := range [][]byte{header[608:], quest[536:]} {
			for _, value := range data {
				if value != 0 {
					t.Fatal("native pointer escaped into an ABI32 header slot")
				}
			}
		}
	}
	// C-owned, nonempty player/unit/update chains exercise both typed list
	// traversal and the original Quest score callback, not a supplied count.
	slots := srv.Players.ListSlots()
	for _, index := range []int{0, 2, server.HostPlayerIndex} {
		player := slots[index]
		player.Active, player.PlayerInd, player.Field4792 = 1, byte(index), 1
		player.SetField2096("ArmQuest")
		update, releaseUpdate := alloc.New(server.PlayerUpdateData{})
		defer releaseUpdate()
		unit, releaseUnit := alloc.New(server.Object{})
		defer releaseUnit()
		update.Player = player
		*unit = server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
		player.PlayerUnit = unit
		if unsafe.Sizeof(uintptr(0)) == 8 {
			for _, pointer := range []unsafe.Pointer{player.C(), unsafe.Pointer(unit), unsafe.Pointer(update)} {
				if uintptr(pointer) <= math.MaxUint32 {
					t.Fatal("actual Quest callback pointer must be above 4 GiB")
				}
			}
		}
	}
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameModeQuest)
	for cycle, count := range []uint16{3, 0, 1, 3} {
		for i, index := range []int{0, 2, server.HostPlayerIndex} {
			slots[index].Field4792 = 0
			if i < int(count) {
				slots[index].Field4792 = 1
			}
		}
		if got := Nox_xxx_player_4E3CE0(); got != int(count) {
			t.Fatalf("cycle %d real Quest count = %d", cycle, got)
		}
		sequence := memmap.PtrT[uint32](0x5D4594, 741672)
		*sequence = math.MaxUint32
		Sub_4264D0()
		quest := memmap.PtrT[[580]byte](0x5D4594, 739396)
		if binary.LittleEndian.Uint16(quest[:2]) != count || *sequence != 0 {
			t.Fatalf("cycle %d Quest count/sequence changed", cycle)
		}
		for _, value := range quest[536:] {
			if value != 0 {
				t.Fatal("native column pointer escaped into packed Quest slots")
			}
		}
	}
}
