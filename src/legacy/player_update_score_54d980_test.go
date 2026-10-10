package legacy

import (
	"encoding/binary"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type arenaScoreLegacyServer54D980 struct {
	Server
	srv    *server.Server
	deaths []*server.Object
}

func TestPlayerUpdateScore54D980RawBodyOnlyABI32(t *testing.T) {
	source, err := os.ReadFile("GAME5.c")
	if err != nil {
		t.Fatal(err)
	}
	signature := "void nox_xxx_playerUpdateScore_54D980(int a1, int a2, int a3, int a4) {"
	if strings.Count(string(source), signature) != 1 {
		t.Fatal("nonunique production Arena root")
	}
	_, tail, _ := strings.Cut(string(source), signature)
	body, _, ok := strings.Cut(tail, "\n//----- (0054DC40)")
	if !ok {
		t.Fatal("missing production Arena end")
	}
	// This compiles the actual existing body, with a declared export service.
	// It certifies ABI32 bit-pattern forwarding only, never recovery of already
	// truncated 64-bit object pointers or competitive death admission.
	program := `
#include <stdint.h>
#include <string.h>
typedef struct nox_object_t nox_object_t;
static uint32_t expected[4];
static unsigned calls;
void nox_server_player_update_score_native_54D980(nox_object_t* victim,
    nox_object_t* killer, nox_object_t* assist, uint32_t tracking) {
    if ((uintptr_t)victim != expected[0] || (uintptr_t)killer != expected[1] ||
        (uintptr_t)assist != expected[2] || tracking != expected[3])
        __builtin_trap();
    calls++;
}
` + signature + body + `
int main(void) {
    _Static_assert(sizeof(int) == 4, "original DWORD arguments");
    const uint32_t values[4] = {0, 0x7fffffff, 0x80000000, 0xffffffff};
    for (unsigned i = 0; i != 4; ++i) {
        int args[4];
        for (unsigned j = 0; j != 4; ++j) {
            expected[j] = values[(i+j)%4];
            memcpy(&args[j], &expected[j], sizeof(args[j]));
        }
        nox_xxx_playerUpdateScore_54D980(args[0], args[1], args[2], args[3]);
    }
    return calls == 4 ? 0 : 1;
}
`
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	binary := filepath.Join(t.TempDir(), "arena-abi32")
	args := append(compiler[1:], "-std=c11", "-O2", "-Wall", "-Wextra", "-Werror", "-x", "c", "-", "-o", binary)
	args = append(args, strings.Fields(os.Getenv("NOX_ARENA_SCORE_CFLAGS"))...)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = strings.NewReader(program)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile actual ABI32 Arena body: %v\n%s", err, output)
	}
	if output, err := exec.Command(binary).CombinedOutput(); err != nil {
		t.Fatalf("actual ABI32 Arena body: %v\n%s", err, output)
	}
}

func (s *arenaScoreLegacyServer54D980) S() *server.Server { return s.srv }

func (s *arenaScoreLegacyServer54D980) PlayerIncrementElimDeath4D8D40(unit *server.Object) {
	s.deaths = append(s.deaths, unit)
	unit.UpdateDataPlayer().Player.Field2140++
}

func TestPlayerUpdateScore54D980CgoNativeEntry(t *testing.T) {
	if os.Getenv("NOX_ARENA_SCORE_CGO_CHILD") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestPlayerUpdateScore54D980CgoNativeEntry$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), "NOX_ARENA_SCORE_CGO_CHILD=1")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("actual CGo native Arena score: %v\n%s", err, output)
		}
		return
	}
	srv := server.New(nil, nil, strman.New())
	var packets [][]byte
	srv.NetSendPacketXxx = func(recipient int, data []byte, related *server.Object, remove, sequence int) int {
		if recipient != 255 || related != nil || remove != 1 || sequence != 1 || len(data) != 11 || data[0] != byte(netmsg.MSG_REPORT_LESSON) {
			t.Fatalf("lesson routing=%d/%p/%d/%d packet=%x", recipient, related, remove, sequence, data)
		}
		packets = append(packets, append([]byte(nil), data...))
		return -17
	}
	bridge := &arenaScoreLegacyServer54D980{srv: srv}
	GetServer = func() Server { return bridge }
	noxflags.ResetGame()
	noxflags.ResetEngine()
	Set_dword_5d4594_2650652(0)
	var units [3]*server.Object
	var players [3]*server.Player
	for i := range units {
		player, releasePlayer := alloc.New(server.Player{})
		defer releasePlayer()
		update, releaseUpdate := alloc.New(server.PlayerUpdateData{})
		defer releaseUpdate()
		unit, releaseUnit := alloc.New(server.Object{})
		defer releaseUnit()
		player.PlayerUnit, update.Player = unit, player
		*unit = server.Object{ObjClass: object.ClassPlayer, NetCode: 0x80008100 + uint32(i), UpdateData: unsafe.Pointer(update)}
		units[i], players[i] = unit, player
		if unsafe.Sizeof(uintptr(0)) == 8 {
			for _, pointer := range []unsafe.Pointer{unsafe.Pointer(player), unsafe.Pointer(update), unsafe.Pointer(unit)} {
				if uintptr(pointer) <= math.MaxUint32 {
					t.Fatal("actual C-owned unit/update/player must be above 4 GiB")
				}
			}
		}
	}
	// Uses the actual C->Go export and real native score/report bindings. The
	// elimination-death and network transport services are explicitly intercepted.
	Nox_xxx_playerUpdateScoreNative54D980(units[0], units[1], units[2], 0x80000000)
	if players[0].Lessons != 0 || players[0].Field2140 != 1 || players[1].Lessons != 1 || players[2].Lessons != 1 || len(bridge.deaths) != 1 || bridge.deaths[0] != units[0] {
		t.Fatal("native killer/assist score or victim death identity changed")
	}
	players[0].Lessons = math.MinInt32
	Nox_xxx_playerUpdateScoreNative54D980(units[0], units[0], nil, 0)
	if players[0].Lessons != math.MaxInt32 || players[0].Field2140 != 1 || len(bridge.deaths) != 1 {
		t.Fatal("native suicide subtraction/wrap unexpectedly incremented deaths")
	}
	Nox_xxx_playerUpdateScoreNative54D980(units[0], nil, nil, 0)
	if players[0].Lessons != math.MaxInt32-1 || players[0].Field2140 != 1 {
		t.Fatal("native environmental death penalty changed")
	}
	if len(packets) != 5 {
		t.Fatalf("actual native lesson reports=%d, want 5", len(packets))
	}
	for i, want := range [][3]uint32{{0x8101, 1, 0}, {0x8100, 0, 1}, {0x8102, 1, 0}, {0x8100, math.MaxInt32, 1}, {0x8100, math.MaxInt32 - 1, 1}} {
		data := packets[i]
		got := [3]uint32{uint32(binary.LittleEndian.Uint16(data[1:])), binary.LittleEndian.Uint32(data[3:]), binary.LittleEndian.Uint32(data[7:])}
		if got != want {
			t.Fatalf("actual lesson packet %d=%x, want %x", i, got, want)
		}
	}
}
