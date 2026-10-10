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

type elimScoreLegacyServer54D7A0 struct {
	Server
	srv    *server.Server
	deaths []*server.Object
}

func (s *elimScoreLegacyServer54D7A0) S() *server.Server { return s.srv }

func (s *elimScoreLegacyServer54D7A0) PlayerIncrementElimDeath4D8D40(unit *server.Object) {
	// Explicit count service boundary; the original count callback has other
	// game-round services outside this root/CGo fixture's certification.
	s.deaths = append(s.deaths, unit)
	unit.UpdateDataPlayer().Player.Field2140++
}

func TestPlayerHandleElimDeath54D7A0RawBodyOnlyABI32(t *testing.T) {
	source, err := os.ReadFile("GAME5.c")
	if err != nil {
		t.Fatal(err)
	}
	signature := "void nox_xxx_playerHandleElimDeath_54D7A0(int a1, int a2) {"
	if strings.Count(string(source), signature) != 1 {
		t.Fatal("nonunique production Elimination root")
	}
	_, tail, _ := strings.Cut(string(source), signature)
	body, _, ok := strings.Cut(tail, "\n//----- (0054D980)")
	if !ok {
		t.Fatal("missing production Elimination end")
	}
	// Compile the actual raw body and declare only its native export service.
	// This is ABI32 bit-pattern forwarding, not high-pointer reconstruction.
	program := `
#include <stdint.h>
#include <string.h>
typedef struct nox_object_t nox_object_t;
static uint32_t expected[2];
static unsigned calls;
void nox_server_player_handle_elim_death_native_54D7A0(nox_object_t* victim, nox_object_t* killer) {
    if ((uintptr_t)victim != expected[0] || (uintptr_t)killer != expected[1])
        __builtin_trap();
    calls++;
}
` + signature + body + `
int main(void) {
    _Static_assert(sizeof(int) == 4, "original DWORD arguments");
    const uint32_t values[4] = {0, 0x7fffffff, 0x80000000, 0xffffffff};
    for (unsigned i = 0; i != 4; ++i) {
        int args[2];
        for (unsigned j = 0; j != 2; ++j) {
            expected[j] = values[(i+j)%4];
            memcpy(&args[j], &expected[j], sizeof(args[j]));
        }
        nox_xxx_playerHandleElimDeath_54D7A0(args[0], args[1]);
    }
    return calls == 4 ? 0 : 1;
}
`
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	binary := filepath.Join(t.TempDir(), "elimination-abi32")
	args := append(compiler[1:], "-std=c11", "-O2", "-Wall", "-Wextra", "-Werror", "-x", "c", "-", "-o", binary)
	args = append(args, strings.Fields(os.Getenv("NOX_ELIM_SCORE_CFLAGS"))...)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = strings.NewReader(program)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile actual ABI32 Elimination body: %v\n%s", err, output)
	}
	if output, err := exec.Command(binary).CombinedOutput(); err != nil {
		t.Fatalf("actual ABI32 Elimination body: %v\n%s", err, output)
	}
}

func TestPlayerHandleElimDeath54D7A0CgoNativeTeamEntry(t *testing.T) {
	if os.Getenv("NOX_ELIM_SCORE_CGO_CHILD") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestPlayerHandleElimDeath54D7A0CgoNativeTeamEntry$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), "NOX_ELIM_SCORE_CGO_CHILD=1")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("actual CGo native Elimination score/team: %v\n%s", err, output)
		}
		return
	}
	srv := server.New(nil, nil, strman.New())
	var packets [][4]uint32
	srv.NetSendPacketXxx = func(recipient int, data []byte, related *server.Object, remove, sequence int) int {
		if related != nil || remove != 1 || sequence != 1 {
			t.Fatalf("score routing=%d/%p/%d/%d packet=%x", recipient, related, remove, sequence, data)
		}
		switch {
		case recipient == 255 && len(data) == 11 && data[0] == byte(netmsg.MSG_REPORT_LESSON):
			packets = append(packets, [4]uint32{0, uint32(binary.LittleEndian.Uint16(data[1:])), binary.LittleEndian.Uint32(data[3:]), binary.LittleEndian.Uint32(data[7:])})
		case recipient == 159 && len(data) == 10 && data[0] == byte(netmsg.MSG_TEAM_MSG) && data[1] == 8:
			packets = append(packets, [4]uint32{1, binary.LittleEndian.Uint32(data[2:]), binary.LittleEndian.Uint32(data[6:]), 0})
		default:
			t.Fatalf("unexpected actual packet recipient=%d bytes=%x", recipient, data)
		}
		return -17
	}
	bridge := &elimScoreLegacyServer54D7A0{srv: srv}
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
		*unit = server.Object{ObjClass: object.Class(0x80000004), NetCode: 0x80008300 + uint32(i), UpdateData: unsafe.Pointer(update)}
		units[i], players[i] = unit, player
		if unsafe.Sizeof(uintptr(0)) == 8 {
			for _, pointer := range []unsafe.Pointer{unsafe.Pointer(player), unsafe.Pointer(update), unsafe.Pointer(unit)} {
				if uintptr(pointer) <= math.MaxUint32 {
					t.Fatal("actual C-owned unit/update/Player must be above 4 GiB")
				}
			}
		}
	}
	units[2].ObjClass = object.ClassMonster
	players[0].Field2140, players[1].Lessons = math.MaxUint32, math.MaxInt32
	Nox_xxx_playerHandleElimDeathNative54D7A0(units[0], units[1])
	players[0].Lessons = math.MinInt32
	Nox_xxx_playerHandleElimDeathNative54D7A0(units[0], units[0])
	Nox_xxx_playerHandleElimDeathNative54D7A0(units[0], nil)
	Nox_xxx_playerHandleElimDeathNative54D7A0(units[0], units[2])

	// Actual native team creation, Has/ByID and TeamChangeLessons, including
	// real lesson/team packet construction. Only count/network are intercepted.
	noxflags.SetGame(noxflags.GameModeCoopTeam)
	first, second := srv.Teams.Create(1), srv.Teams.Create(2)
	if first == nil || second == nil || srv.Teams.ByID(1) != first || srv.Teams.ByID(2) != second {
		t.Fatal("actual native team lookup failed")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(first)) <= math.MaxUint32 || uintptr(unsafe.Pointer(second)) <= math.MaxUint32) {
		t.Fatal("actual C-owned teams must be above 4 GiB")
	}
	first.Lessons, second.Lessons = math.MaxInt32, -1
	units[0].TeamVal.ID, units[1].TeamVal.ID = 1, 2
	noxflags.SetGame(noxflags.GameHost)
	Nox_xxx_playerHandleElimDeathNative54D7A0(units[0], units[1])
	units[1].TeamVal.ID = 1
	Nox_xxx_playerHandleElimDeathNative54D7A0(units[0], units[1])
	Nox_xxx_playerHandleElimDeathNative54D7A0(units[0], units[0])
	if players[0].Lessons != math.MaxInt32-1 || players[0].Field2140 != 6 || players[1].Lessons != math.MinInt32 || first.Lessons != math.MinInt32+2 || second.Lessons != -1 || len(bridge.deaths) != 7 {
		t.Fatal("native score/death/team wrap or unchanged killer team mismatch")
	}
	for _, unit := range bridge.deaths {
		if unit != units[0] {
			t.Fatal("victim death count identity changed")
		}
	}
	want := [][4]uint32{
		{0, 0x8301, 0x80000000, 0}, {0, 0x8300, 0, 0},
		{0, 0x8300, math.MaxInt32, 1}, {0, 0x8300, math.MaxInt32, 2}, {0, 0x8300, math.MaxInt32, 3},
		{0, 0x8301, 0x80000001, 0}, {0, 0x8300, math.MaxInt32, 4}, {1, 1, 0x80000000, 0},
		{0, 0x8301, 0x80000000, 0}, {0, 0x8300, math.MaxInt32, 5}, {1, 1, 0x80000001, 0},
		{0, 0x8300, math.MaxInt32 - 1, 6}, {1, 1, 0x80000002, 0},
	}
	if len(packets) != len(want) {
		t.Fatalf("actual lesson/team packets=%d, want %d", len(packets), len(want))
	}
	for i, value := range want {
		if packets[i] != value {
			t.Fatalf("actual lesson/team packet %d=%x, want %x", i, packets[i], value)
		}
	}
}
