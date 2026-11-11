package opennox

import (
	"bytes"
	"fmt"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
	"math"
	"testing"
	"unsafe"
)

func basicSenderProductionFixture5187E0(t *testing.T) (*Server, *server.Object, *server.MonsterGenUpdateData, *server.HealthData) {
	t.Helper()
	base := server.New(nil, nil, strman.New())
	t.Cleanup(base.Close)
	base.NetList.Init()
	t.Cleanup(base.NetList.Free)
	base.SetFrame(100)
	s := &Server{Server: base}
	unit, fu := alloc.NewHigh(server.Object{})
	update, fd := alloc.NewHigh(server.MonsterGenUpdateData{})
	hp, fh := alloc.NewHigh(server.HealthData{})
	for _, free := range []func(){fu, fd, fh} {
		t.Cleanup(free)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, p := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(update), unsafe.Pointer(hp)} {
			if uintptr(p) <= math.MaxUint32 {
				t.Fatal("production records must exceed4GiB")
			}
		}
	}
	oldServer, oldGetter, oldQueue, oldFlags := noxServer, legacy.GetServer, legacy.Nox_netlist_addToMsgListSrv, noxflags.GetGame()
	noxServer = s
	legacy.GetServer = func() legacy.Server { return s }
	legacy.Nox_netlist_addToMsgListSrv = nox_netlist_addToMsgListSrv
	noxflags.ResetGame()
	t.Cleanup(func() {
		noxServer, legacy.GetServer, legacy.Nox_netlist_addToMsgListSrv = oldServer, oldGetter, oldQueue
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})
	return s, unit, update, hp
}
func TestBasicSenderOrdinaryBindings5187E0(t *testing.T) {
	s, unit, update, hp := basicSenderProductionFixture5187E0(t)
	cases := []struct {
		name   string
		simple bool
		class  object.Class
		sub    object.SubClass
		health bool
		frame  byte
		code   uint16
	}{
		{"Phantom idle", false, object.ClassMissile, 0, false, 0xff, 0x2345},
		{"Phantom missile frame", false, object.ClassMissile, 0x30, false, 0, 0x2345},
		{"Phantom persistent code", false, object.ClassComplex | object.ClassClientPersist, 0, false, 0xff, 0x9234},
		{"simple", true, object.ClassSimple, 0, false, 0, 0x2345},
		{"generator nil health", true, object.ClassSimple | object.ClassMonsterGenerator, 0, false, 0, 0x2345},
		{"generator young damage", true, object.ClassSimple | object.ClassMonsterGenerator, 0, true, 0, 0x2345},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for path := 0; path < 3; path++ {
				t.Run(fmt.Sprintf("path%d", path), func(t *testing.T) {
					s.NetList.ResetAll()
					*unit = server.Object{ObjClass: tc.class, ObjSubClass: tc.sub, TypeInd: 73, NetCode: 0x2345, Extent: 0x1234, PosVec: types.Ptf(256.75, -123.75), Frame134: 100}
					*update = server.MonsterGenUpdateData{}
					update.HealthSamples[3] = 211
					*hp = server.HealthData{Cur: 199}
					if tc.health {
						unit.UpdateData = unsafe.Pointer(update)
						unit.HealthData = hp
					}
					var got int
					if tc.simple {
						switch path {
						case 0:
							got = legacy.Nox_xxx_netSendSimpleObj_5188A0(3, unit)
						case 1:
							got = int(legacy.NetSendSimpleObject5188A0(3, unit))
						case 2:
							got = int(s.ObjectDispatchRuntimeNative519410().Simple(3, unit))
						}
					} else {
						switch path {
						case 0:
							got = legacy.Nox_xxx_netSendPhantomPlrMb_5187E0(3, unit)
						case 1:
							got = int(legacy.NetSendPhantomObject5187E0(3, unit))
						case 2:
							got = int(s.ObjectDispatchRuntimeNative519410().Phantom(3, unit))
						}
					}
					want := []byte{48, byte(tc.code), byte(tc.code >> 8), 73, 0, 0, 1, 0x85, 0xff, 0x40, tc.frame}
					if tc.simple {
						want = want[:9]
						want[0] = 47
					}
					a, b := s.NetList.CopyPacketsA(3, netlist.Kind1), s.NetList.CopyPacketsB(3)
					if got != 1 || len(a) != 0 || !bytes.Equal(b, want) {
						t.Fatalf("return=%d A=%x B=%x original=%x", got, a, b, want)
					}
					if update.HealthSamples[3] != 211 || hp.Cur != 199 || s.Frame() != 100 {
						t.Fatal("young/nil-health production path changed health cache/frame")
					}
				})
			}
		})
	}
}
func TestBasicSenderOrdinaryMissingGeneratorCache5188A0(t *testing.T) {
	s, unit, _, hp := basicSenderProductionFixture5187E0(t)
	*hp = server.HealthData{Cur: 199}
	*unit = server.Object{ObjClass: object.ClassSimple | object.ClassMonsterGenerator, TypeInd: 73, NetCode: 0x2345, HealthData: hp, Frame134: 97}
	if s.ObjectDispatchRuntimeNative519410().Simple(3, unit) != 1 {
		t.Fatal("existing ordinary missing-cache policy")
	}
	packet := s.NetList.CopyPacketsB(3)
	if len(packet) != 9 || packet[0] != 47 || unit.HealthData != hp || unit.UpdateData != nil {
		t.Fatal("ordinary missing-cache publication/state", packet)
	}
}
