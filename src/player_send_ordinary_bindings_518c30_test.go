package opennox

import (
	"bytes"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// These cases exercise the production factory and both real netlists. The
// complete original-reference suite separately checks the actual C important
// health queue, independent flags, every load/callback order and raw returns.
func TestPlayerSendOrdinaryBindings518C30(t *testing.T) {
	base := server.New(nil, nil, strman.New())
	t.Cleanup(base.Close)
	base.NetList.Init()
	t.Cleanup(base.NetList.Free)
	s := &Server{Server: base}
	recipient, freeR := alloc.New(server.Object{})
	unit, freeU := alloc.New(server.Object{})
	ru, freeRU := alloc.New(server.PlayerUpdateData{})
	uu, freeUU := alloc.New(server.PlayerUpdateData{})
	for _, free := range []func(){freeR, freeU, freeRU, freeUU} {
		t.Cleanup(free)
	}
	rp, up := base.Players.ByIndRaw(3), base.Players.ByIndRaw(11)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, p := range []unsafe.Pointer{unsafe.Pointer(recipient), unsafe.Pointer(unit), unsafe.Pointer(ru), unsafe.Pointer(uu), unsafe.Pointer(rp), unsafe.Pointer(up)} {
			if uintptr(p) <= math.MaxUint32 {
				t.Fatal("production records must exceed4GiB")
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
	cases := []struct {
		name             string
		state            server.PlayerState
		field131         uint32
		self             bool
		animation, frame byte
	}{
		{"idle", 0, 0, false, 4, 0xff},
		{"walk", 2, 0, false, 21, 0xa5},
		{"emote", 25, 0, false, 20, 0xa5},
		{"original override", 3, 16, false, 51, 0xff},
		{"self gold report", 0, 0, true, 4, 0xff},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for stream := 0; stream < 2; stream++ {
				t.Run(fmt.Sprintf("stream%d", stream), func(t *testing.T) {
					for path := 0; path < 3; path++ {
						t.Run([]string{"C", "publicGo", "ordinaryGo"}[path], func(t *testing.T) {
							noxflags.ResetGame()
							base.NetList.ResetAll()
							base.SetFrame(0x12340064)
							*recipient = server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(ru)}
							*unit = server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(uu), TypeInd: 73, NetCode: 0x2345, PosVec: types.Ptf(256.75, -123.75), Field131: tc.field131, Direction1: 0}
							*ru = server.PlayerUpdateData{Player: rp}
							*uu = server.PlayerUpdateData{Player: up, State: tc.state, Field59_0: 0xa5}
							*rp = server.Player{PlayerInd: 3, PlayerUnit: recipient}
							*up = server.Player{PlayerInd: 11, PlayerUnit: unit}
							for i := range rp.PeerUpdateFrames {
								rp.PeerUpdateFrames[i] = 0xa0000000 + uint32(i)
							}
							wantFrames := rp.PeerUpdateFrames
							target, update, targetPlayer := unit, uu, up
							if tc.self {
								*recipient = *unit
								recipient.UpdateData = unsafe.Pointer(ru)
								target, update, targetPlayer = recipient, ru, rp
								rp.GoldVal, rp.Field2168 = 37, 12
							}
							wantFrames[targetPlayer.PlayerInd] = base.Frame()
							var reports [][]byte
							base.NetSendPacketXxx = func(index int, packet []byte, related *server.Object, remove, sequence int) int {
								if index != 3 || related != nil || remove != 1 || sequence != 0 {
									t.Fatalf("self publication %d/%p/%d/%d", index, related, remove, sequence)
								}
								reports = append(reports, append([]byte(nil), packet...))
								return 1
							}
							var got int
							switch path {
							case 0:
								got = legacy.Nox_xxx_netPlayerObjSend_518C30(recipient, target, stream, 0)
							case 1:
								got = int(legacy.NetPlayerObjectSend518C30(recipient, target, int32(stream), 0))
							case 2:
								if s.netPlayerObjectSendNative518C30(recipient, target, stream != 0) {
									got = 1
								}
							}
							want := []byte{0xc3, 0x45, 0x23, 73, 0, 0, 1, 0x85, 0xff, 0x40, tc.frame, tc.animation}
							a, b := base.NetList.CopyPacketsA(3, netlist.Kind1), base.NetList.CopyPacketsB(3)
							var wantA, wantB []byte
							if stream != 0 {
								wantB = want
							} else {
								wantA = want
							}
							if got != 1 || !bytes.Equal(a, wantA) || !bytes.Equal(b, wantB) {
								t.Fatalf("return=%d A=%x want%x B=%x want%x", got, a, wantA, b, wantB)
							}
							if rp.PeerUpdateFrames != wantFrames || update.Field41 != 0 || update.HealthSamples != [32]uint16{} {
								t.Fatal("production peer stamp or unrelated state")
							}
							if tc.self {
								if len(reports) != 1 || !bytes.Equal(reports[0], []byte{74, 37, 0, 0, 0}) || rp.Field2168 != 37 {
									t.Fatalf("actual self report=%x gold cache=%d", reports, rp.Field2168)
								}
							} else if len(reports) != 0 {
								t.Fatalf("unexpected self reports=%x", reports)
							}
						})
					}
				})
			}
		})
	}
}
