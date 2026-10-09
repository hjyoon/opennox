package legacy

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"unsafe"

	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func TestStatsStart426060NativeBody(t *testing.T) {
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
	for _, signature := range []string{"static void* getMemAt(", "static uint32_t* getMemU32Ptr(", "static uint8_t* getMemU8Ptr(", "static int16_t* getMemI16Ptr("} {
		services = strings.Replace(services, signature, strings.Replace(signature, " getMem", " bulk_getMem", 1), 1)
	}
	program := read("testdata/stats_start_native_426060_test.c")
	// Headers must precede the forward declarations in the reused services.
	program = strings.Replace(program, "// BULK_FIXTURE_SERVICES", services, 1)
	program = "#include <stdint.h>\n" + program
	program = strings.Replace(program, "// PRODUCTION_BULK_HEADER", read("stats_bulk_native.h"), 1)
	program = strings.Replace(program, "// PRODUCTION_START_HEADER", read("stats_start_native.h"), 1)
	program = strings.Replace(program, "// PRODUCTION_BULK_BODY", cut(read("stats_bulk_native.c"), "// NATIVE_STATS_IMPLEMENTATION", ""), 1)
	program = strings.Replace(program, "// PRODUCTION_START_BODY", cut(read("stats_start_native.c"), "// NATIVE_STATS_START_IMPLEMENTATION", ""), 1)
	ipSignature := "char* sub_4282D0(char* a1, int a2) {"
	program = strings.Replace(program, "// PRODUCTION_IP_ROOT", ipSignature+cut(read("GAME1_2.c"), ipSignature, "\n//----- (004282F0)"), 1)
	rootSignature := "void sub_426060() {"
	program = strings.Replace(program, "// PRODUCTION_START_ROOT", rootSignature+cut(read("server__system__server.c"), rootSignature, "\n//----- (004D0CF0)"), 1)
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	binary := filepath.Join(t.TempDir(), "stats-start")
	args := append(compiler[1:], "-std=c11", "-O2", "-Wall", "-Wextra", "-Werror", "-Wno-int-to-void-pointer-cast", "-x", "c", "-", "-o", binary)
	args = append(args, strings.Fields(os.Getenv("NOX_STATS_START_CFLAGS"))...)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = strings.NewReader(program)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile actual statistics startup/packet chain: %v\n%s", err, output)
	}
	for _, name := range []string{"header", "root", "cleanup"} {
		t.Run(name, func(t *testing.T) {
			if output, err := exec.Command(binary, "--"+name).CombinedOutput(); err != nil {
				t.Fatalf("native start %s: %v\n%s", name, err, output)
			} else {
				t.Log(strings.TrimSpace(string(output)))
			}
			if receipt := os.Getenv("NOX_STATS_START_NATIVE_JSON"); receipt != "" {
				output, err := exec.Command(binary, "--"+name, "--json").Output()
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(receipt+"."+name+".jsonl", output, 0600); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
	packetPrefix := "TOUGQLHLGQCJVXYYY" + strings.Repeat("NZD", 31)
	faultPrefixes := map[string]string{
		"settings_nil":    "TOUGQLHLGQCJ",
		"game_nil":        "TOUGQLHLGQCJV",
		"node_alloc":      "TOUGQLHLGQCJVXYYYN",
		"data_alloc":      "TOUGQLHLGQCJVXYYYNZD",
		"packet_alloc":    packetPrefix + "P",
		"quest_names_nil": "TOUGQQCJKV",
		"transport":       packetPrefix + "P" + strings.Repeat("FF", 31) + "EFB",
	}
	for _, fault := range []string{"settings_nil", "game_nil", "node_alloc", "data_alloc", "packet_alloc", "quest_names_nil", "transport"} {
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
			if string(output) != faultPrefixes[fault] {
				t.Fatalf("original fault service prefix changed: %q, want %q", output, faultPrefixes[fault])
			}
			if receipt := os.Getenv("NOX_STATS_START_NATIVE_JSON"); receipt != "" {
				if err := os.WriteFile(receipt+"."+fault+".events", output, 0600); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

// The child uses real C-owned player/team arrays, list callbacks and the
// production compressor/envelope. It does not start a GUI or network sender.
func TestStatsStart426060CgoEntry(t *testing.T) {
	if os.Getenv("NOX_STATS_START_CGO_CHILD") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestStatsStart426060CgoEntry$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), "NOX_STATS_START_CGO_CHILD=1")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("real CGo statistics start: %v\n%s", err, output)
		}
		return
	}
	srv := server.New(nil, nil, strman.New())
	srv.OwnIPStr = "203.0.113.250"
	GetServer = func() Server { return &playerTrackingLegacyServer425F10{srv: srv} }
	noxflags.ResetGame()
	noxflags.ResetEngine()
	slots := srv.Players.ListSlots()
	for _, index := range []int{0, 2, 31} {
		player := slots[index]
		player.Active = 1
		player.PlayerInd = byte(index)
		player.Field4648 = 75
		player.SetField2096("ArmHost")
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(player.C()) <= math.MaxUint32 {
			t.Fatalf("player address %p is not above 4 GiB", player)
		}
	}
	// First/Next must visit a nonempty native array. Only field60, not active
	// alone, contributes to the original BYTE team count.
	for _, index := range []int{1, 3, 5} {
		team := &srv.Teams.Arr[index]
		data := unsafe.Slice((*byte)(team.C()), int(unsafe.Sizeof(*team)))
		data[58] = byte(index)
		binary.LittleEndian.PutUint32(data[64:68], 1)
		if index != 3 {
			binary.LittleEndian.PutUint32(data[60:64], 0x80000001)
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(team.C()) <= math.MaxUint32 {
			t.Fatalf("team address %p is not above 4 GiB", team)
		}
	}
	header := memmap.PtrT[[640]byte](0x5D4594, 599476)
	quest := memmap.PtrT[[580]byte](0x5D4594, 739396)
	settings := memmap.PtrT[[128]byte](0x5D4594, 371516)
	game := memmap.PtrT[[128]byte](0x5D4594, 371380)
	address := memmap.PtrT[[16]byte](0x5D4594, 741316)
	for _, gameFlags := range []noxflags.GameFlag{0, noxflags.GameModeQuest,
		noxflags.GameOnline, noxflags.GameOnline | noxflags.GameModeQuest} {
		for _, port := range []int{32767, 32768, 65535} {
			noxflags.ResetGame()
			noxflags.SetGame(gameFlags)
			srv.SetServerPort(port)
			*header = [640]byte{}
			*quest = [580]byte{}
			for i := range settings {
				settings[i] = byte(i*13 + 7)
			}
			for i := range game {
				game[i] = byte(i*17 + 3)
			}
			copy(game[:9], "map-1234\x00")
			copy(game[9:24], "server-123456789")
			binary.LittleEndian.PutUint16(game[52:54], 0x4320)
			Set_dword_5d4594_608316(75)
			Sub_426060()
			if *address != [16]byte{'2', '0', '3', '.', '0', '.', '1', '1', '3', '.', '2', '5', '0'} {
				t.Fatalf("IP truncation/padding changed: %x", address)
			}
			if got := *memmap.PtrUint32(0x5D4594, 741304); got != uint32(int32(int16(port))) {
				t.Fatalf("port=%d DWORD=%x, want signed WORD promotion", port, got)
			}
			online := gameFlags.Has(noxflags.GameOnline) && !gameFlags.Has(noxflags.GameModeQuest)
			if online {
				if Get_dword_5d4594_608316() != 3 || slots[31].Field4648 != 0 || slots[0].Field4648 != 1 || slots[2].Field4648 != 2 {
					t.Fatal("real player callbacks did not register host first, then list order")
				}
				if header[25] != 2 || binary.LittleEndian.Uint32(header[20:24]) != 0 || binary.LittleEndian.Uint16(header[6:8]) != 0 ||
					!bytes.Equal(header[608:], make([]byte, 32)) {
					t.Fatalf("team count, post-report count or ABI32 slots changed: %x", header[:32])
				}
			} else if Get_dword_5d4594_608316() != 0 {
				t.Fatal("non-online start registered players")
			}
			if gameFlags.Has(noxflags.GameModeQuest) {
				if binary.LittleEndian.Uint16(quest[:2]) != 3 || quest[16] != 1 ||
					binary.LittleEndian.Uint32(quest[12:16]) != 5 || string(quest[24:32]) != "map-1234" ||
					string(quest[280:295]) != "server-12345678" || !bytes.Equal(quest[536:], make([]byte, 44)) {
					t.Fatalf("Quest initialization changed: %x", quest[:40])
				}
			}
		}
	}
}
