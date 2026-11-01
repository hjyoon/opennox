package opennox

import (
	"bytes"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// These checks reach the production factory through public C, public Go and
// both ordinary adapters. The lower important transport is a declared capture
// boundary; height uses the real client netlist. Child-body equivalence remains
// covered by each child's existing original-reference suite.
func TestObjectSpecialOrdinaryBindings527E50(t *testing.T) {
	s, unit, _, hp := basicSenderProductionFixture5187E0(t)
	r, fr := alloc.New(server.Object{})
	d, fd := alloc.New(server.PlayerUpdateData{})
	m, fm := alloc.New(server.MonsterUpdateData{})
	for _, free := range []func(){fr, fd, fm} {
		t.Cleanup(free)
	}
	p := s.Players.ByIndRaw(3)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(r), unsafe.Pointer(d), unsafe.Pointer(m), unsafe.Pointer(p)} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatal("production records must exceed4GiB")
			}
		}
	}
	active := t
	var transport [][]byte
	var sequence []int
	s.NetSendPacketXxx = func(index int, packet []byte, related *server.Object, remove, seq int) int {
		if index != 3 || related != nil || remove != 1 {
			active.Fatal("lower transport arguments", index, related, remove)
		}
		transport = append(transport, append([]byte(nil), packet...))
		sequence = append(sequence, seq)
		return math.MinInt32
	}
	npc := []byte{byte(netmsg.MSG_REPORT_NPC), 0x45, 0x23}
	for i := 0; i < 18; i++ {
		npc = append(npc, byte(i+1))
	}
	cases := []struct {
		name     string
		bit      uint32
		packet   []byte
		sequence int
		height   bool
	}{
		{"animation", 0x10000, []byte{byte(netmsg.MSG_REPORT_ANIMATION_FRAME), 0x45, 0x23, 0x88, 0x77, 0x66, 0x55}, 0, false},
		{"health", 0x20000, []byte{65, 0x45, 0x23, 99}, 1, false},
		{"hidden", 0x40000, []byte{byte(netmsg.MSG_ENABLE_OBJECT), 0x45, 0x23}, 0, false},
		{"xstatus", 0x80000, []byte{byte(netmsg.MSG_REPORT_X_STATUS), 0x45, 0x23, 4, 0x33, 0x22, 0x11}, 0, false},
		{"height", 0x400000, []byte{byte(netmsg.MSG_REPORT_Z_PLUS), 0x45, 0x23, 12}, 0, true},
		{"enchant", 0x800000, []byte{byte(netmsg.MSG_REPORT_ENCHANTMENT), 0x45, 0x23, 0x78, 0x56, 0x34, 0x12}, 1, false},
		{"team", 0x2000000, []byte{byte(netmsg.MSG_REPORT_MODIFIER), 0x45, 0x23, 0xff, 0xff, 0xff, 0xff}, 0, false},
		{"npc", 0x4000000, npc, 1, false},
		{"unhandled", 0, nil, 0, false},
		{"no-special-flags", 0, nil, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for path := 0; path < 4; path++ {
				t.Run(fmt.Sprintf("path%d", path), func(t *testing.T) {
					active = t
					s.NetList.ResetAll()
					transport = nil
					sequence = nil
					*r = server.Object{ObjClass: object.ClassPlayer, TypeInd: 73, UpdateData: unsafe.Pointer(d)}
					*d = server.PlayerUpdateData{Player: p}
					p.PlayerInd = 3
					*m = server.MonsterUpdateData{}
					for i := range m.Color {
						m.Color[i] = server.Color3{R: byte(3*i + 1), G: byte(3*i + 2), B: byte(3*i + 3)}
					}
					*hp = server.HealthData{Cur: 199}
					*unit = server.Object{ObjClass: object.ClassMonster | object.ClassVisibleEnable | object.ClassWeapon, ObjSubClass: 0x20, ObjFlags: object.FlagEnabled, TypeInd: 73, NetCode: 0x2345, UpdateData: unsafe.Pointer(m), HealthData: hp, Field33: 0x55667788, Field5: 0x11223304, ZVal: 12.75, Buffs: 0x12345678}
					for i := range unit.Field140 {
						unit.Field140[i] = 0xaabb0000 + uint32(i)
					}
					unit.Field140[3] = tc.bit | 0x001000ab
					wantReturn := int32(1)
					if tc.name == "no-special-flags" {
						unit.Field140[3] = 0xf00000ab
						wantReturn = 0
					}
					wantFlags := unit.Field140
					wantFlags[3] &^= tc.bit
					var got int32
					switch path {
					case 0:
						legacy.Nox_xxx_netUpdateObjectSpecial_527E50(r, unit)
					case 1:
						got = legacy.NetUpdateObjectSpecial527E50(r, unit)
					case 2:
						if s.NetUpdateObjectSpecialNative527E50(r, unit) {
							got = 1
						}
					case 3:
						got = int32(s.ObjectDispatchRuntimeNative519410().Special(r, unit))
					}
					if path != 0 && got != wantReturn {
						t.Fatal("root return", got, wantReturn)
					}
					if unit.Field140 != wantFlags || r.UpdateData != unsafe.Pointer(d) || d.Player != p || p.PlayerInd != 3 || hp.Cur != 199 {
						t.Fatal("production state/peer neighbors")
					}
					a, b := s.NetList.CopyPacketsA(3, netlist.Kind1), s.NetList.CopyPacketsB(3)
					if tc.height {
						if len(transport) != 0 || !bytes.Equal(a, tc.packet) || len(b) != 0 {
							t.Fatal("real height netlist", transport, a, b)
						}
					} else {
						if len(a) != 0 || len(b) != 0 {
							t.Fatal("unexpected object netlist", a, b)
						}
						if tc.packet == nil {
							if len(transport) != 0 {
								t.Fatal("unhandled/no-flags publication", transport)
							}
						} else if len(transport) != 1 || !bytes.Equal(transport[0], tc.packet) || len(sequence) != 1 || sequence[0] != tc.sequence {
							t.Fatal("actual production child packet/mode", transport, sequence, tc.packet, tc.sequence)
						}
					}
				})
			}
		})
	}
}

func TestObjectSpecialOrdinaryPolicy527E50(t *testing.T) {
	s := &Server{Server: &server.Server{}}
	if !s.NetUpdateObjectSpecialNative527E50(nil, nil) || !s.NetUpdateObjectSpecialNative527E50(&server.Object{}, nil) || !s.NetUpdateObjectSpecialNative527E50(nil, &server.Object{}) {
		t.Fatal("ordinary early nil policy")
	}
	d := &server.PlayerUpdateData{}
	r := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(d)}
	u := &server.Object{}
	if !s.NetUpdateObjectSpecialNative527E50(r, u) {
		t.Fatal("ordinary missing Player policy")
	}
	d.Player = &server.Player{PlayerInd: 255}
	if !s.NetUpdateObjectSpecialNative527E50(r, u) {
		t.Fatal("ordinary byte index gate")
	}
	for _, name := range []string{"class", "update"} {
		t.Run(name, func(t *testing.T) {
			r := &server.Object{}
			if name == "update" {
				r.ObjClass = object.ClassPlayer
			}
			defer func() {
				if recover() == nil {
					t.Fatal("existing class/nil-update fault was changed")
				}
			}()
			s.NetUpdateObjectSpecialNative527E50(r, u)
		})
	}
}
