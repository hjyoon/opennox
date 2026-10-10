package legacy

import (
	"encoding/binary"
	"math"
	"os"
	"os/exec"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
	"github.com/opennox/libs/strman"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type competitiveDeathLegacyServer54D2B0 struct {
	Server
	srv     *server.Server
	deaths  []*server.Object
	deleted []*server.Object
}

func (s *competitiveDeathLegacyServer54D2B0) S() *server.Server { return s.srv }

// The match-countdown/death increment and delayed deletion are declared outer
// services of this entry test. The production death binding, three score roots,
// real score/report packet builders, rivals/limit readers, and owned-object
// cleanup remain unchanged. This does not execute a whole competitive match.
func (s *competitiveDeathLegacyServer54D2B0) PlayerIncrementElimDeath4D8D40(unit *server.Object) {
	s.deaths = append(s.deaths, unit)
	unit.UpdateDataPlayer().Player.Field2140++
}

func (s *competitiveDeathLegacyServer54D2B0) DelayedDelete(unit *server.Object) {
	s.deleted = append(s.deleted, unit)
}

func TestPlayerDieCompetitive54D2B0ProductionCEntry(t *testing.T) {
	if os.Getenv("NOX_COMPETITIVE_DEATH_CGO_CHILD") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestPlayerDieCompetitive54D2B0ProductionCEntry$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), "NOX_COMPETITIVE_DEATH_CGO_CHILD=1")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("production competitive death C entry: %v\n%s", err, output)
		}
		t.Logf("isolated actual C entry:\n%s", output)
		return
	}
	srv := server.New(nil, nil, strman.New())
	defer srv.Close()
	srv.SetTickRate(30)
	srv.SetFrame(1000)
	bridge := &competitiveDeathLegacyServer54D2B0{srv: srv}
	GetServer = func() Server { return bridge }
	noxflags.ResetGame()
	noxflags.ResetEngine()
	noxflags.UnsetGamePlay(noxflags.GameplayFlag(math.MaxUint32))
	Set_dword_5d4594_2650652(0)
	playerDieAnkhType54D2B0 = 77              // Already-initialized native type-cache branch.
	*memmap.PtrUint32(0x5D4594, 1567716) = 83 // Already-initialized Crown lookup.
	var units [3]*server.Object
	var updates [3]*server.PlayerUpdateData
	var players [3]*server.Player
	var health [3]*server.HealthData
	for i := range units {
		unit, freeUnit := alloc.New(server.Object{})
		defer freeUnit()
		update, freeUpdate := alloc.New(server.PlayerUpdateData{})
		defer freeUpdate()
		hp, freeHealth := alloc.New(server.HealthData{})
		defer freeHealth()
		*hp = server.HealthData{Cur: 0, Max: 100}
		p := srv.Players.ByIndRaw(ntype.PlayerInd(i))
		units[i], updates[i], players[i], health[i] = unit, update, p, hp
		questPenaltyLegacyHighPointers54CBD0(t, unsafe.Pointer(unit), unsafe.Pointer(update), unsafe.Pointer(p), unsafe.Pointer(hp))
	}
	monster, freeMonster := alloc.New(server.Object{})
	defer freeMonster()
	*monster = server.Object{ObjClass: object.ClassMonster, ObjFlags: object.FlagEnabled, NetCode: 0x9300}
	ownedAttack, freeAttack := alloc.New(server.Object{})
	defer freeAttack()
	*ownedAttack = server.Object{ObjClass: object.ClassMissile, ObjOwner: units[1]}
	spawned, freeSpawned := alloc.New(server.Object{})
	defer freeSpawned()
	*spawned = server.Object{ObjClass: object.ClassMissile, TypeInd: 77, ObjOwner: units[0]}
	questPenaltyLegacyHighPointers54CBD0(t, unsafe.Pointer(monster), unsafe.Pointer(ownedAttack), unsafe.Pointer(spawned))
	states, abilities := 0, 0
	// legacy is below the package registering these two callbacks. Only those
	// missing outer services are supplied, as in the existing Quest C-entry test.
	Nox_xxx_playerSetState_4FA020 = func(unit *server.Object, state server.PlayerState) bool {
		if unit != units[0] || state != server.PlayerState3 {
			t.Fatal("production state callback identity")
		}
		states++
		unit.UpdateDataPlayer().State = state
		return false
	}
	Nox_xxx_playerCancelAbils_4FC180 = func(unit *server.Object) {
		if unit != units[0] {
			t.Fatal("production ability callback identity")
		}
		abilities++
	}
	var packets [][4]uint32
	srv.NetSendPacketXxx = func(recipient int, data []byte, related *server.Object, remove, sequence int) int {
		if recipient != 255 || related != nil || sequence != 1 {
			t.Fatalf("production packet route=%d/%p/%d/%d bytes=%x", recipient, related, remove, sequence, data)
		}
		switch {
		case len(data) == 11 && data[0] == byte(netmsg.MSG_REPORT_LESSON) && remove == 1:
			packets = append(packets, [4]uint32{0, uint32(binary.LittleEndian.Uint16(data[1:])), binary.LittleEndian.Uint32(data[3:]), binary.LittleEndian.Uint32(data[7:])})
		case len(data) == 3 && data[0] == byte(netmsg.MSG_PLAYER_DIED) && remove == 0:
			packets = append(packets, [4]uint32{1, uint32(binary.LittleEndian.Uint16(data[1:])), 0, 0})
		default:
			t.Fatalf("unexpected production packet remove=%d bytes=%x", remove, data)
		}
		return -17 // The original death/report roots ignore these transport returns.
	}
	cases := []struct {
		name         string
		mode         noxflags.GameFlag
		source       int // -1 environment, 0/1 player, 3 monster, 4 owned attack.
		assist       bool
		noRivals     bool
		limit        uint16
		deathsBefore uint32
		wantLessons  [3]int32
		wantDeaths   uint32
		increments   int
		removals     int
		wantPackets  [][4]uint32
	}{
		{"arena killer", noxflags.GameModeArena, 1, false, false, 0, 0, [3]int32{0, 1, 0}, 1, 1, 0, [][4]uint32{{0, 0x9201, 1, 0}, {0, 0x9200, 0, 1}}},
		{"arena assist", noxflags.GameModeArena, 1, true, false, 0, 0, [3]int32{0, 1, 1}, 1, 1, 0, [][4]uint32{{0, 0x9201, 1, 0}, {0, 0x9200, 0, 1}, {0, 0x9202, 1, 0}}},
		{"arena owned killer", noxflags.GameModeArena, 4, false, false, 0, 0, [3]int32{0, 1, 0}, 1, 1, 0, [][4]uint32{{0, 0x9201, 1, 0}, {0, 0x9200, 0, 1}}},
		{"arena self", noxflags.GameModeArena, 0, false, false, 0, 0, [3]int32{-1, 0, 0}, 0, 0, 0, [][4]uint32{{0, 0x9200, math.MaxUint32, 0}}},
		{"arena environment", noxflags.GameModeArena, -1, false, false, 0, 0, [3]int32{-1, 0, 0}, 0, 0, 0, [][4]uint32{{0, 0x9200, math.MaxUint32, 0}}},
		{"arena monster", noxflags.GameModeArena, 3, false, false, 0, 0, [3]int32{}, 1, 1, 0, [][4]uint32{{0, 0x9200, 0, 1}}},
		{"kotr pawn", noxflags.GameModeKOTR, 1, false, false, 0, 0, [3]int32{}, 1, 1, 0, [][4]uint32{{0, 0x9200, 0, 1}}},
		{"kotr environment", noxflags.GameModeKOTR, -1, false, false, 0, 0, [3]int32{}, 0, 0, 0, nil},
		{"elim below limit", noxflags.GameModeElimination, 1, false, false, 2, 0, [3]int32{0, 1, 0}, 1, 1, 0, [][4]uint32{{0, 0x9201, 1, 0}, {0, 0x9200, 0, 1}}},
		{"elim at limit", noxflags.GameModeElimination, 1, false, false, 1, 0, [3]int32{0, 1, 0}, 1, 1, 1, [][4]uint32{{0, 0x9201, 1, 0}, {0, 0x9200, 0, 1}}},
		{"elim self", noxflags.GameModeElimination, 0, false, false, 1, 0, [3]int32{-1, 0, 0}, 1, 1, 1, [][4]uint32{{0, 0x9200, math.MaxUint32, 1}}},
		{"elim monster", noxflags.GameModeElimination, 3, false, false, 1, 0, [3]int32{}, 1, 1, 1, [][4]uint32{{0, 0x9200, 0, 1}}},
		{"elim environment", noxflags.GameModeElimination, -1, false, false, 1, 0, [3]int32{}, 1, 1, 1, [][4]uint32{{0, 0x9200, 0, 1}}},
		{"elim zero limit", noxflags.GameModeElimination, 1, false, false, 0, 0, [3]int32{0, 1, 0}, 1, 1, 0, [][4]uint32{{0, 0x9201, 1, 0}, {0, 0x9200, 0, 1}}},
		{"elim unsigned word", noxflags.GameModeElimination, 1, false, false, 0xffff, 0xfffe, [3]int32{0, 1, 0}, 0xffff, 1, 1, [][4]uint32{{0, 0x9201, 1, 0}, {0, 0x9200, 0, 0xffff}}},
		{"elim signed death", noxflags.GameModeElimination, 1, false, false, 1, math.MaxInt32, [3]int32{0, 1, 0}, 0x80000000, 1, 0, [][4]uint32{{0, 0x9201, 1, 0}, {0, 0x9200, 0, 0x80000000}}},
		{"elim no rivals cleanup", noxflags.GameModeElimination, 1, false, true, 1, 1, [3]int32{}, 1, 0, 1, nil},
	}
	totalPackets, totalIncrements, totalRemovals := 0, 0, 0
	for _, item := range cases {
		noxflags.ResetGame()
		noxflags.SetGame(noxflags.GameHost | item.mode) // Deliberately not GameOnline.
		for i := range players {
			*units[i] = server.Object{ObjClass: object.ClassPlayer, ObjFlags: object.FlagEnabled, NetCode: 0x80009200 + uint32(i), UpdateData: unsafe.Pointer(updates[i]), HealthData: health[i]}
			*updates[i] = server.PlayerUpdateData{Player: players[i], ManaCur: 17, TrapSpellsCnt: 0xaabbcc05, TrapSpells: [5]uint32{1, 2, 3, 4, 5}}
			*players[i] = server.Player{Active: 1, PlayerUnit: units[i], PlayerInd: byte(i), NetCodeVal: units[i].NetCode}
			players[i].Info().SetPlayerClass(player.Warrior)
			if item.noRivals && i != 0 {
				players[i].Active = 0
			}
		}
		units[0].ObjFlags |= object.FlagDead
		units[0].Field129 = spawned
		players[0].Field2140 = item.deathsBefore
		if item.assist {
			players[0].Field3600, players[0].Field3604 = 0x80000000, 2
			players[0].SetLastAggressorFrame(srv.Frame())
		}
		switch item.source {
		case 0, 1:
			units[0].Obj130 = units[item.source]
		case 3:
			units[0].Obj130 = monster
		case 4:
			units[0].Obj130 = ownedAttack
		}
		for slot := 0; slot < 6; slot++ {
			*memmap.PtrUint16(0x5D4594, 3488+2*uintptr(slot)) = item.limit
		}
		if uint16(Nox_xxx_servGamedataGet_40A020(0x400)) != item.limit {
			t.Fatal("real C limit WORD reader")
		}
		if (Nox_xxx_gamePlayIsAnyPlayers_40A8A0() != 0) == item.noRivals {
			t.Fatal("real C rivals reader")
		}
		packets, bridge.deaths, bridge.deleted = nil, nil, nil
		states, abilities = 0, 0
		playerDieExportCall54D2B0(units[0])
		wantPackets := append(append([][4]uint32(nil), item.wantPackets...), [4]uint32{1, 0x9200, 0, 0})
		if !reflect.DeepEqual(packets, wantPackets) || states != 1 || abilities != 1 || len(bridge.deaths) != item.increments || len(bridge.deleted) != item.removals {
			t.Fatalf("%s packets=%x want=%x state/abilities=%d/%d increments=%d/%d removals=%d/%d", item.name, packets, wantPackets, states, abilities, len(bridge.deaths), item.increments, len(bridge.deleted), item.removals)
		}
		for i, p := range players {
			wantDeaths := uint32(0)
			if i == 0 {
				wantDeaths = item.wantDeaths
			}
			if p.Lessons != item.wantLessons[i] || p.Field2140 != wantDeaths {
				t.Fatalf("%s player %d lessons/deaths=%d/%x want=%d/%x", item.name, i, p.Lessons, p.Field2140, item.wantLessons[i], wantDeaths)
			}
		}
		for _, unit := range bridge.deaths {
			if unit != units[0] {
				t.Fatal("death increment identity")
			}
		}
		for _, unit := range bridge.deleted {
			if unit != spawned {
				t.Fatal("spawned cleanup identity")
			}
		}
		update := updates[0]
		if units[0].ObjFlags != object.FlagEnabled|object.FlagDead|object.FlagShort || update.State != server.PlayerState3 || update.ManaCur != 0 || update.TrapSpells != [5]uint32{} || update.TrapSpellsCnt != 0xaabbcc00 || players[0].Field3600 != 0 || update.Trade70 != nil {
			t.Fatal("common death tail did not finish")
		}
		totalPackets += len(packets)
		totalIncrements += len(bridge.deaths)
		totalRemovals += len(bridge.deleted)
	}
	t.Logf("production C death entries=%d real lesson/death packets=%d declared increments=%d declared deletions=%d; all unit/update/player/health/owned chains C-owned above 4 GiB", len(cases), totalPackets, totalIncrements, totalRemovals)
}
