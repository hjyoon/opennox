package legacy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
	"unsafe"

	"github.com/opennox/libs/console"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

type playerIDNativeServer417040 struct {
	Server
	base *server.Server
}

func (s *playerIDNativeServer417040) S() *server.Server { return s.base }

func playerIDNativeFixture417040(t *testing.T) *server.Server {
	t.Helper()
	s := server.New(nil, console.NewMultiPrinter(), strman.New())
	t.Cleanup(s.Close)
	old := GetServer
	GetServer = func() Server { return &playerIDNativeServer417040{base: s} }
	t.Cleanup(func() { GetServer = old })
	return s
}

func playerIDBytes417040(p *server.Player) []byte {
	return bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p))))
}

// GAME.EXE 417040 compares the complete 32-bit code with CMP/JE, skips
// inactive records, and writes the selected array index into PlayerInd before
// returning. The CGo int32 argument must not become a different LP64 int.
// Registry records are real C-owned player slots, not lookup result hooks.
func TestPlayerInfoGetByID417040RawBitsAndOriginalIndex(t *testing.T) {
	s := playerIDNativeFixture417040(t)
	slots := s.Players.ListSlots()
	for _, raw := range []uint32{0, 1, 0xffff, 0x10000, 0x7fffffff, 0x80000000, 0xfffffffe, 0xffffffff} {
		for _, index := range []int{0, 17, 31} {
			for _, active := range []byte{0, 1, 2} {
				t.Run(fmt.Sprintf("%08x/index%d/active%d", raw, index, active), func(t *testing.T) {
					for _, p := range slots {
						*p = server.Player{}
					}
					first := slots[index]
					*first = server.Player{NetCodeVal: raw, PlayerInd: 99, Active: active, Field3680: 0x76543210}
					duplicateIndex := 31
					if index == 31 {
						duplicateIndex = 0
					}
					duplicate := slots[duplicateIndex]
					*duplicate = server.Player{NetCodeVal: raw, PlayerInd: 98, Active: 1, Field3680: 0x12345678}
					expectedIndex := -1
					for i, p := range slots {
						if p.Active != 0 && p.NetCodeVal == raw {
							expectedIndex = i
							break
						}
					}
					if expectedIndex < 0 {
						t.Fatal("fixture lacks active matching slot")
					}
					before := make([][]byte, len(slots))
					for i, p := range slots {
						if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(p)) <= 0xffffffff {
							t.Fatalf("player array is not above 4 GiB: %p", p)
						}
						before[i] = playerIDBytes417040(p)
					}
					got := (*server.Player)(unsafe.Pointer(nox_common_playerInfoGetByID_417040(int32(raw))))
					if got != slots[expectedIndex] {
						t.Errorf("raw code=%08x CGo int32=%d pointer=%p want=%p", raw, int32(raw), got, slots[expectedIndex])
					}
					for i, p := range slots {
						want := bytes.Clone(before[i])
						if i == expectedIndex {
							want[int(unsafe.Offsetof(server.Player{}.PlayerInd))] = byte(i)
						}
						if !bytes.Equal(playerIDBytes417040(p), want) {
							t.Errorf("slot %d changed outside original index store, or omitted it: index=%d want=%d", i, p.PlayerInd, expectedIndex)
						}
					}
				})
			}
		}
	}
	t.Run("inactive and missing codes leave every record unchanged", func(t *testing.T) {
		for _, p := range slots {
			*p = server.Player{}
		}
		*slots[17] = server.Player{NetCodeVal: 0xffffffff, PlayerInd: 99, Active: 0, Field3680: 0x76543210}
		before := playerIDBytes417040(slots[17])
		for _, raw := range []uint32{0, 0x1337, 0x80000000, 0xffffffff} {
			if got := nox_common_playerInfoGetByID_417040(int32(raw)); got != nil {
				t.Errorf("inactive/missing raw code=%08x returned %p", raw, got)
			}
		}
		if !bytes.Equal(before, playerIDBytes417040(slots[17])) {
			t.Fatal("failed lookup changed inactive record")
		}
	})
}

// The actual C 419570 caller queries 417040 before the 418E40 player reset.
// Read the genuine native Important queue, without replacing sender or HP,
// membership, relation, or player services. This is not a GUI/network peer test.
func TestPlayerInfoGetByID417040ActualCTeamRemovalRefresh(t *testing.T) {
	for _, raw := range []uint32{0x1337, 0x7fffffff, 0x80000000, 0xf123abcd, 0xffffffff} {
		t.Run(fmt.Sprintf("%08x", raw), func(t *testing.T) {
			s := playerIDNativeFixture417040(t)
			s.SetFrame(37)
			s.SetTickRate(30)
			oldFlags := noxflags.GetGame()
			noxflags.ResetGame()
			noxflags.SetGame(noxflags.GameHost | noxflags.GameOnline)
			t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(oldFlags) })
			handles.Init()
			t.Cleanup(handles.Release)
			preserveImportantPacketState(t)
			setImportantNativeListC(nil, nil)
			setImportantAllocClassC(nil)
			setImportantRecipientMaskC(1)
			*memmap.PtrT[[32]uint16](0x5D4594, 1565524) = [32]uint16{}
			player := s.Players.NewRaw(int(raw))
			if player == nil {
				t.Fatal("real player registry allocation failed")
			}
			update, freeUpdate := alloc.New(server.PlayerUpdateData{})
			t.Cleanup(freeUpdate)
			unit, freeUnit := alloc.New(server.Object{})
			t.Cleanup(freeUnit)
			*update = server.PlayerUpdateData{Player: player, State: 9}
			*unit = server.Object{TypeInd: 713, ObjClass: object.ClassPlayer, ObjFlags: object.FlagEnabled, NetCode: raw, UpdateData: unsafe.Pointer(update), Extent: 0x89abcdef}
			player.PlayerUnit = unit
			t.Cleanup(func() { player.PlayerUnit = nil })
			team := s.Teams.Create(7)
			if s.Teams.AttachObject(unit.TeamPtr(), team.ID()) != team {
				t.Fatal("real team attachment failed")
			}
			playerBefore, updateBefore := playerIDBytes417040(player), bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(update)), int(unsafe.Sizeof(*update))))
			Nox_xxx_netChangeTeamMb_419570(unit.TeamPtr(), raw)
			if unit.TeamVal.ID != 0 || s.Teams.MemberCount(team) != 0 || unit.Field35 != 1 || unit.Field36 != 1 {
				t.Fatalf("real C removal/relations failed: team=%d members=%d masks=%x/%x", unit.TeamVal.ID, s.Teams.MemberCount(team), unit.Field35, unit.Field36)
			}
			if !bytes.Equal(playerBefore, playerIDBytes417040(player)) || !bytes.Equal(updateBefore, unsafe.Slice((*byte)(unsafe.Pointer(update)), int(unsafe.Sizeof(*update)))) {
				t.Fatal("C removal corrupted player/update")
			}
			var packets [][]byte
			for p, _ := importantNativeListC(); p != nil; p = importantPacketNextC(p) {
				if len(packets) > 3 {
					t.Fatal("real C queue repeats packets")
				}
				packet := (*importantPacketLegacy)(p)
				if packet.Recipient != -97 && packet.Recipient != 0 || packet.SequenceEnabled != 1 || packet.RemoveIfDisconnected != 1 {
					t.Fatal("real packet metadata differs")
				}
				packets = append(packets, bytes.Clone(packet.Payload[:packet.PayloadSize]))
			}
			remove := []byte{0xc4, 2, 0, 0, 0, 0}
			binary.LittleEndian.PutUint32(remove[2:], raw)
			if len(packets) != 2 || !bytes.Equal(packets[0], []byte{54}) || !bytes.Equal(packets[1], remove) {
				t.Fatalf("actual C player refresh/removal packets=%x, want=36/%x", packets, remove)
			}
			Nox_xxx_netChangeTeamMb_419570(unit.TeamPtr(), raw)
			count := 0
			for p, _ := importantNativeListC(); p != nil; p = importantPacketNextC(p) {
				count++
			}
			if count != 2 {
				t.Fatal("repeated removal emitted extra packet")
			}
			if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(player)) <= 0xffffffff || uintptr(unsafe.Pointer(unit)) <= 0xffffffff || uintptr(unsafe.Pointer(update)) <= 0xffffffff || uintptr(team.C()) <= 0xffffffff) {
				t.Fatal("native fixtures must be above 4 GiB")
			}
		})
	}
}
