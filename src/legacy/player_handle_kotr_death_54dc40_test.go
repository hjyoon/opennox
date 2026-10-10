package legacy

import (
	"encoding/binary"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestPlayerHandleKotrDeath54DC40RawBodyOnlyABI32(t *testing.T) {
	source, err := os.ReadFile("GAME5.c")
	if err != nil {
		t.Fatal(err)
	}
	signature := "void nox_xxx_playerHandleKotrDeath_54DC40(int a1, int a2) {"
	if strings.Count(string(source), signature) != 1 {
		t.Fatal("nonunique production KotR root")
	}
	_, tail, _ := strings.Cut(string(source), signature)
	body, _, ok := strings.Cut(tail, "\n//----- (0054DF00)")
	if !ok {
		t.Fatal("missing production KotR end")
	}
	// Compile the actual raw body and declare only its native export service.
	// This is ABI32 bit-pattern forwarding, not high-pointer reconstruction.
	program := `
#include <stdint.h>
#include <string.h>
typedef struct nox_object_t nox_object_t;
static uint32_t expected[2];
static unsigned calls;
void nox_server_player_handle_kotr_death_native_54DC40(nox_object_t* victim, nox_object_t* killer) {
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
        nox_xxx_playerHandleKotrDeath_54DC40(args[0], args[1]);
    }
    return calls == 4 ? 0 : 1;
}
`
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	binary := filepath.Join(t.TempDir(), "kotr-abi32")
	args := append(compiler[1:], "-std=c11", "-O2", "-Wall", "-Wextra", "-Werror", "-x", "c", "-", "-o", binary)
	args = append(args, strings.Fields(os.Getenv("NOX_KOTR_SCORE_CFLAGS"))...)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = strings.NewReader(program)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile actual ABI32 KotR body: %v\n%s", err, output)
	}
	if output, err := exec.Command(binary).CombinedOutput(); err != nil {
		t.Fatalf("actual ABI32 KotR body: %v\n%s", err, output)
	}
}

type kotrScoreLegacyServer54DC40 struct {
	Server
	srv    *server.Server
	deaths []*server.Object
}

func (s *kotrScoreLegacyServer54DC40) S() *server.Server { return s.srv }
func (s *kotrScoreLegacyServer54DC40) PlayerIncrementElimDeath4D8D40(unit *server.Object) {
	s.deaths = append(s.deaths, unit)
	unit.UpdateDataPlayer().Player.Field2140++
}

func TestPlayerHandleKotrDeath54DC40CgoNativeCrownTeamEntry(t *testing.T) {
	if os.Getenv("NOX_KOTR_SCORE_CGO_CHILD") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestPlayerHandleKotrDeath54DC40CgoNativeCrownTeamEntry$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), "NOX_KOTR_SCORE_CGO_CHILD=1")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("actual CGo native KotR crown/team: %v\n%s", err, output)
		}
		return
	}
	srv := server.New(nil, nil, strman.New())
	bridge := &kotrScoreLegacyServer54DC40{srv: srv}
	GetServer = func() Server { return bridge }
	noxflags.ResetGame()
	noxflags.ResetEngine()
	noxflags.UnsetGamePlay(noxflags.GameplayFlag(math.MaxUint32))
	Set_dword_5d4594_2650652(0)
	// Generated balance service input, not an edited stock/retained YAML file.
	dataRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(dataRoot, "gamedata.yml"), []byte("KotRKingKillsKingPoints: 2.5\nKotRKingKillsPawnPoints: 1.5\nKotRPawnKillsKingPoints: -2.5\n"), 0600); err != nil {
		t.Fatal(err)
	}
	datapath.SetData(dataRoot)
	if err := srv.Balance.Read(); err != nil {
		t.Fatal(err)
	}
	if srv.Balance.Float("KotRKingKillsPawnPoints") != 1.5 {
		t.Fatal("real balance reader unavailable")
	}
	var packets [][4]uint32
	srv.NetSendPacketXxx = func(recipient int, data []byte, related *server.Object, remove, sequence int) int {
		if related != nil || remove != 1 || sequence != 1 {
			t.Fatal("actual score packet routing")
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
	var units [3]*server.Object
	var players [3]*server.Player
	for i := range units {
		unit, fu := alloc.New(server.Object{})
		defer fu()
		update, fd := alloc.New(server.PlayerUpdateData{})
		defer fd()
		player, fp := alloc.New(server.Player{})
		defer fp()
		*unit = server.Object{ObjClass: object.Class(0x80000004), NetCode: 0x80008400 + uint32(i), UpdateData: unsafe.Pointer(update)}
		update.Player, player.PlayerUnit = player, unit
		units[i], players[i] = unit, player
		for _, p := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(update), unsafe.Pointer(player)} {
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
				t.Fatal("C-owned chain must be above 4 GiB")
			}
		}
	}
	units[2].ObjClass = object.ClassMonster
	noxflags.SetGame(noxflags.GameModeCoopTeam)
	first, second := srv.Teams.Create(1), srv.Teams.Create(2)
	if first == nil || second == nil || srv.Teams.ByID(1) != first || srv.Teams.ByID(2) != second {
		t.Fatal("real native team lookup")
	}
	for _, team := range []*server.Team{first, second} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(team)) <= math.MaxUint32 {
			t.Fatal("C-owned team must be above 4 GiB")
		}
	}
	noxflags.SetGame(noxflags.GameHost)
	var crowns [2]*server.Object
	var crownUpdates [2]*server.CrownUpdateData
	slot, fs := alloc.New(uint64(0))
	defer fs()
	drops := 0
	server.RegisterObjectDrop("KotR54DC40DeclaredDropBoundary", unsafe.Pointer(slot), func(owner, item *server.Object, point *types.Pointf) int32 {
		if owner != units[0] || item != crowns[0] || point != &units[0].PosVec {
			t.Fatal("native crown drop identity")
		}
		drops++
		return -17
	})
	for i := range crowns {
		crown, fc := alloc.New(server.Object{})
		defer fc()
		data, fd := alloc.New(server.CrownUpdateData{})
		defer fd()
		*crown = server.Object{TypeInd: 83, UpdateData: unsafe.Pointer(data), Drop: server.DropFuncPtr{Ptr: unsafe.Pointer(slot)}}
		crowns[i], crownUpdates[i] = crown, data
		for _, p := range []unsafe.Pointer{unsafe.Pointer(crown), unsafe.Pointer(data)} {
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
				t.Fatal("C-owned crown/update must be above 4 GiB")
			}
		}
	}
	*memmap.PtrUint32(0x5D4594, 1567716) = 83
	*memmap.PtrUint32(0x5D4594, dropOwnedCrownsTypeCacheOffset4ED050) = 83
	cases := []struct {
		name                   string
		killer                 int
		victimTeam, killerTeam server.TeamID
		crownMask              uint32
		want                   [][4]uint32
		deaths, drops          int
	}{
		{"environment", -1, 0, 0, 0, nil, 0, 0},
		{"monster", 2, 0, 0, 3, nil, 0, 0},
		{"pawn", 1, 0, 0, 0, [][4]uint32{{0, 0x8400, 0x80000000, 0}}, 1, 0},
		{"self pawn", 0, 0, 0, 0, [][4]uint32{{0, 0x8400, 0x80000000, 0}}, 1, 0},
		{"self king", 0, 0, 0, 1, [][4]uint32{{0, 0x8400, math.MaxInt32, math.MaxUint32}, {0, 0x8400, math.MaxInt32, 0}}, 1, 0},
		{"free for all kings", 1, 0, 0, 3, [][4]uint32{{0, 0x8401, 0x80000001, math.MaxUint32}, {0, 0x8400, 0x80000000, 0}}, 1, 1},
		{"different team kings", 1, 1, 2, 3, [][4]uint32{{1, 2, 0x80000001, 0}, {0, 0x8401, 0x80000001, math.MaxUint32}, {0, 0x8400, 0x80000000, 0}}, 1, 0},
		{"pawn kills team king", 1, 1, 2, 1, [][4]uint32{{1, 2, math.MaxInt32 - 2, 0}, {0, 0x8401, math.MaxInt32 - 2, math.MaxUint32}, {0, 0x8400, 0x80000000, 0}}, 1, 0},
		{"same team king", 1, 1, 1, 2, [][4]uint32{{0, 0x8401, math.MaxInt32 - 1, math.MaxUint32}, {1, 1, math.MaxInt32, 0}, {0, 0x8400, 0x80000000, 0}}, 1, 0},
		{"same team pawn", 1, 1, 1, 0, [][4]uint32{{0, 0x8400, 0x80000000, 0}}, 1, 0},
	}
	totalPackets, totalDeaths, totalDrops := 0, 0, 0
	for _, item := range cases {
		packets = nil
		bridge.deaths = nil
		drops = 0
		for i := range players {
			players[i].Lessons = math.MaxInt32
			players[i].Field2140 = math.MaxUint32
			units[i].Field129 = nil
		}
		players[0].Lessons = math.MinInt32
		first.Lessons, second.Lessons = math.MinInt32, math.MaxInt32
		units[0].TeamVal.ID, units[1].TeamVal.ID = item.victimTeam, item.killerTeam
		for i := range crowns {
			crownUpdates[i].PickupTarget = nil
			if item.crownMask&(1<<i) != 0 {
				units[i].Field129 = crowns[i]
			}
		}
		var killer *server.Object
		if item.killer >= 0 {
			killer = units[item.killer]
		}
		Nox_xxx_playerHandleKotrDeathNative54DC40(units[0], killer)
		if !reflect.DeepEqual(packets, item.want) || len(bridge.deaths) != item.deaths || drops != item.drops {
			t.Fatalf("%s native packets=%x want=%x deaths=%d/%d drops=%d/%d", item.name, packets, item.want, len(bridge.deaths), item.deaths, drops, item.drops)
		}
		if drops != 0 && crownUpdates[0].PickupTarget != killer {
			t.Fatal("cached native Crown update lost pickup target")
		}
		for _, unit := range bridge.deaths {
			if unit != units[0] {
				t.Fatal("wrong death identity")
			}
		}
		totalPackets += len(packets)
		totalDeaths += len(bridge.deaths)
		totalDrops += drops
	}
	t.Logf("actual typed CGo calls=%d packets=%d deaths=%d declared crown drops=%d; observer mode0", len(cases), totalPackets, totalDeaths, totalDrops)
}
