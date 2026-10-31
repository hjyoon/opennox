package opennox

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestObjectDispatchOrdinaryBindings519410(t *testing.T) {
	base := server.New(nil, nil, strman.New())
	t.Cleanup(base.Close)
	base.NetList.Init()
	t.Cleanup(base.NetList.Free)
	if base.Walls.Init() == 0 {
		t.Fatal("native empty wall grid")
	}
	t.Cleanup(base.Walls.Free)
	s := &Server{Server: base}
	recipient, freeRecipient := alloc.New(server.Object{})
	unit, freeUnit := alloc.New(server.Object{})
	update, freeUpdate := alloc.New(server.PlayerUpdateData{})
	playerUpdate, freePlayerUpdate := alloc.New(server.PlayerUpdateData{})
	monsterUpdate, freeMonsterUpdate := alloc.New(server.MonsterUpdateData{})
	elevatorUpdate, freeElevatorUpdate := alloc.New(server.ElevatorUpdateData{})
	track, freeTrack := alloc.New(server.MinimapItem{})
	for _, free := range []func(){freeRecipient, freeUnit, freeUpdate, freePlayerUpdate, freeMonsterUpdate, freeElevatorUpdate, freeTrack} {
		t.Cleanup(free)
	}
	player := base.Players.ByIndRaw(3)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, pointer := range []unsafe.Pointer{unsafe.Pointer(recipient), unsafe.Pointer(unit), unsafe.Pointer(update), unsafe.Pointer(player), unsafe.Pointer(playerUpdate), unsafe.Pointer(monsterUpdate), unsafe.Pointer(elevatorUpdate), unsafe.Pointer(track)} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatal("ordinary dispatcher inputs must exceed4GiB")
			}
		}
	}
	oldServer, oldGetter, oldQueue, oldFlags := noxServer, legacy.GetServer, legacy.Nox_netlist_addToMsgListSrv, noxflags.GetGame()
	noxServer = s
	legacy.GetServer = func() legacy.Server { return s }
	legacy.Nox_netlist_addToMsgListSrv = nox_netlist_addToMsgListSrv
	t.Cleanup(func() {
		noxServer, legacy.GetServer, legacy.Nox_netlist_addToMsgListSrv = oldServer, oldGetter, oldQueue
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})
	activeTest := t
	var transport [][]byte
	base.NetSendPacketXxx = func(index int, packet []byte, related *server.Object, remove, sequence int) int {
		if index != 3 || related != nil || remove != 1 || sequence < 0 || sequence > 1 {
			activeTest.Fatal("ordinary lower publication arguments", index, related, remove, sequence)
		}
		transport = append(transport, append([]byte(nil), packet...))
		return 1
	}
	for _, name := range []string{"sprite", "complex", "player", "phantom", "simple", "friend", "owner-denied", "owner-allowed", "game-allowed", "tracked", "ray-clear", "out", "shadow", "special", "no-sender", "destroyed", "prediction"} {
		t.Run(name, func(t *testing.T) {
			for mode := 0; mode < 3; mode++ {
				t.Run(fmt.Sprintf("mode%d", mode), func(t *testing.T) {
					activeTest = t
					noxflags.ResetGame()
					base.NetList.ResetAll()
					transport = nil
					*recipient = server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
					*update = server.PlayerUpdateData{Player: player}
					*player = server.Player{PlayerInd: 3, PlayerUnit: recipient, Pos3632Vec: types.Ptf(64, 64)}
					*playerUpdate = server.PlayerUpdateData{Player: base.Players.ByIndRaw(11)}
					*monsterUpdate = server.MonsterUpdateData{}
					*elevatorUpdate = server.ElevatorUpdateData{Field_4: 64}
					*track = server.MinimapItem{Field4: unit}
					track.Field8 = track
					*unit = server.Object{ObjClass: object.ClassImmobile | object.ClassElevator, TypeInd: 37, Extent: 0x1234, NetCode: 0x2345, Field38: 8, PosVec: types.Ptf(80, 80), UpdateData: unsafe.Pointer(elevatorUpdate)}
					wantSent, wantReturn, wantSeen, wantPending := true, byte(8), uint32(8), uint32(0)
					var wantA, wantB []byte
					wantTransport := 0
					wantA = []byte{byte(netmsg.MSG_DRAW_FRAME), 0x34, 0x92, 16}
					switch name {
					case "complex":
						unit.ObjClass = object.ClassClientPersist | object.ClassComplex | object.ClassMonster
						unit.UpdateData = unsafe.Pointer(monsterUpdate)
						wantA = nil
					case "player":
						unit.ObjClass = object.ClassClientPersist | object.ClassComplex | object.ClassPlayer
						unit.UpdateData = unsafe.Pointer(playerUpdate)
						wantA = nil
						wantB = []byte{0xc3, 0x34, 0x92, 37, 0, 80, 0, 80, 0, 0x40, 0xff, 4}
					case "phantom":
						unit.ObjClass = object.ClassClientPersist | object.ClassComplex
						unit.UpdateData = nil
						wantA = nil
						wantB = []byte{48, 0x34, 0x92, 37, 0, 80, 0, 80, 0, 0x40, 0xff}
					case "simple", "tracked", "ray-clear":
						unit.ObjClass = object.ClassSimple | object.ClassClientPersist
						unit.UpdateData = nil
						wantA = nil
						wantB = []byte{47, 0x34, 0x92, 37, 0, 80, 0, 80, 0}
						if name != "simple" {
							unit.ObjClass &^= object.ClassClientPersist
							wantB[1], wantB[2] = 0x45, 0x23
						}
						if name == "tracked" {
							player.Field4580 = track
							player.Pos3632Vec.X = -1000 // Tracking must precede the invalid ray origin.
						}
					case "friend":
						unit.Field35, unit.Field36 = 0x80000008, 8
						wantTransport = 1
					case "owner-denied", "owner-allowed", "game-allowed":
						unit.ObjFlags = object.FlagOwnerVisible
						if name == "owner-allowed" {
							unit.ObjOwner = recipient
						} else if name == "game-allowed" {
							noxflags.SetGame(noxflags.GameModeCoopTeam)
						} else {
							wantSent, wantReturn, wantSeen, wantPending, wantA = false, 0, 0, 8, nil
						}
					case "out", "shadow":
						unit.ObjClass, unit.UpdateData, unit.Field37 = object.ClassMonster, nil, 8
						if name == "shadow" {
							unit.ObjClass = object.ClassSimple
						}
						player.Pos3632Vec.X = -1000
						wantSent, wantReturn, wantSeen, wantPending, wantA = false, 0, 0, 8, nil
						wantTransport = 1
					case "special":
						unit.Field140[3] = 0x80000
						unit.Field5 = 0x12345678
						wantTransport = 1
					case "no-sender":
						unit.ObjClass, unit.UpdateData = object.ClassClientPersist, nil
						wantSent, wantReturn, wantSeen, wantPending, wantA = false, 0, 0, 0, nil
					case "destroyed", "prediction":
						if name == "destroyed" {
							unit.ObjFlags = object.FlagDestroyed
							wantReturn = 0x20
						} else {
							unit.ObjClass |= object.ClassClientPredict
							wantReturn = 0
						}
						wantSent, wantSeen, wantPending, wantA = false, 0, 8, nil
					}
					if mode == 2 {
						if got := s.netSendObjects2PlayerNative519410(recipient, unit); got != wantSent {
							t.Fatalf("ordinary sent%t want%t", got, wantSent)
						}
					} else {
						var result byte
						if mode == 0 {
							result = legacy.Nox_xxx_netSendObjects2Plr_519410(recipient, unit)
						} else {
							result = legacy.NetObjectDispatch519410(recipient, unit)
						}
						if result != wantReturn {
							t.Fatalf("actual public byte%#x want%#x", result, wantReturn)
						}
					}
					a, b := base.NetList.CopyPacketsA(3, netlist.Kind1), base.NetList.CopyPacketsB(3)
					if !bytes.Equal(a, wantA) || (name != "complex" && !bytes.Equal(b, wantB)) || len(transport) != wantTransport {
						t.Fatalf("native queues A=%x want%x B=%x want%x transport=%x wantcount%d", a, wantA, b, wantB, transport, wantTransport)
					}
					if name == "complex" && (len(b) != 11 || b[0] != 48 || binary.LittleEndian.Uint16(b[1:]) != 0x9234) {
						t.Fatalf("native complex sender identity/packet length=%x", b)
					}
					if name == "friend" && (!bytes.Equal(transport[0], []byte{byte(netmsg.MSG_OBJECT_FRIEND_ADD), 0x34, 0x92}) || unit.Field35 != 0x80000000) {
						t.Fatal("native friend publication or pending bit", transport, unit.Field35)
					}
					if name == "out" && !bytes.Equal(transport[0], []byte{byte(netmsg.MSG_OBJECT_OUT_OF_SIGHT), 0x45, 0x23}) {
						t.Fatal("native out-of-sight publication", transport)
					}
					if name == "shadow" && !bytes.Equal(transport[0], []byte{byte(netmsg.MSG_OBJECT_IN_SHADOWS), 0x45, 0x23}) {
						t.Fatal("native shadow publication", transport)
					}
					if name == "special" && (!bytes.Equal(transport[0], []byte{byte(netmsg.MSG_REPORT_X_STATUS), 0x34, 0x92, 0x78, 0x56, 0x34, 0x12}) || unit.Field140[3] != 0) {
						t.Fatal("native special publication or consumed flag", transport, unit.Field140[3])
					}
					if unit.Field37 != wantSeen || unit.Field38 != wantPending {
						t.Fatalf("live masks seen%#x/pending%#x want%#x/%#x", unit.Field37, unit.Field38, wantSeen, wantPending)
					}
				})
			}
		})
	}
}
