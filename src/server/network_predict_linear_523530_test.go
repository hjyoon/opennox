package server

import (
	"encoding/binary"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
)

func TestPredictLinear523530PacketAndLiveTraversal(t *testing.T) {
	obj := &Object{}
	p1, p2 := &Player{PlayerInd: 255}, &Player{PlayerInd: 17}
	u1 := &Object{UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: p1})}
	u2 := &Object{UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: p2})}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 {
		t.Fatalf("object=%p, want an actual pointer above 4 GiB", obj)
	}
	want := []byte{0xb5, 0x45, 0x23, 0xef, 0xcd, 1, 0, 0xff, 0xff, 0x80, 0xfe, 0xe1, 0x7f, 0x7f}
	var events []string
	got := predictLinear523530(obj, predictLinearDeps523530{
		netCode: func(got *Object) uint32 {
			if got != obj {
				t.Fatal("net code lost native pointer")
			}
			events = append(events, "code")
			*obj = Object{TypeInd: 0xcdef, PosVec: types.Ptf(65537.75, -1.75), Direction1: 0xfe80, Float28: -1.96875, VelVec: types.Ptf(7.96875, -8.0625)}
			return 0x12345
		},
		first: func() *Object { events = append(events, "first"); return u1 },
		send: func(index uint8, packet []byte) {
			if !reflect.DeepEqual(packet, want) {
				t.Fatalf("packet=% x, want % x", packet, want)
			}
			if len(events) == 2 {
				if index != 255 {
					t.Fatalf("first raw player byte=%d", index)
				}
				events = append(events, "send1")
				*obj = Object{} // The packet is a snapshot, not rebuilt per recipient.
				u1.Field128 = u2
				p2.PlayerInd = 31
			} else {
				if index != 31 {
					t.Fatalf("live second player byte=%d", index)
				}
				events = append(events, "send2")
			}
		},
		next: func(unit *Object) *Object {
			if unit == u1 {
				events = append(events, "next1")
			} else {
				events = append(events, "next2")
			}
			return unit.Field128
		},
	})
	if got != 0 || !reflect.DeepEqual(events, []string{"code", "first", "send1", "next1", "send2", "next2"}) {
		t.Fatalf("result/events=%d/%v", got, events)
	}
}

func TestPredictLinear523530QwordConversions(t *testing.T) {
	for _, tc := range []struct {
		name   string
		value  float32
		word   uint16
		scaled byte
	}{
		{"positive fraction", 1.75, 1, 28}, {"negative fraction", -1.75, 0xffff, 228},
		{"word wrap", 65537.75, 1, 28}, {"signed dword limit", 2147483648, 0, 0},
		{"qword last binary32", math.Float32frombits(0x5effffff), 0, 0},
		{"qword positive overflow", math.Float32frombits(0x5f000000), 0, 0},
		{"qword negative limit", math.Float32frombits(0xdf000000), 0, 0},
		{"qword negative overflow", math.Float32frombits(0xdf000001), 0, 0},
		{"NaN", math.Float32frombits(0x7fc01234), 0, 0},
		{"positive infinity", float32(math.Inf(1)), 0, 0}, {"negative infinity", float32(math.Inf(-1)), 0, 0},
		{"negative zero", math.Float32frombits(0x80000000), 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			obj := &Object{PosVec: types.Ptf(tc.value, tc.value), Float28: tc.value, VelVec: types.Ptf(tc.value, tc.value)}
			unit := &Object{UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: &Player{PlayerInd: 7}})}
			calls := 0
			predictLinear523530(obj, predictLinearDeps523530{
				netCode: func(*Object) uint32 { return 0 }, first: func() *Object { return unit }, next: func(*Object) *Object { return nil },
				send: func(index uint8, packet []byte) {
					calls++
					if index != 7 || binary.LittleEndian.Uint16(packet[5:7]) != tc.word || binary.LittleEndian.Uint16(packet[7:9]) != tc.word ||
						packet[11] != tc.scaled || packet[12] != tc.scaled || packet[13] != tc.scaled {
						t.Fatalf("conversion packet=% x", packet)
					}
				},
			})
			if calls != 1 {
				t.Fatalf("send count=%d", calls)
			}
		})
	}
}

func TestPredictLinear523530NativeMessageQueues(t *testing.T) {
	s := &Server{NetList: netlist.New()}
	s.NetList.Init()
	t.Cleanup(s.NetList.Free)
	s.Players.list = make([]Player, 4)
	for i := range s.Players.list {
		p := &s.Players.list[i]
		p.PlayerInd, p.Active = byte(i), 1
		if i != 1 {
			p.PlayerUnit = &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: p})}
		}
	}
	s.Players.list[2].Active = 0
	obj := &Object{TypeInd: 0x2345, NetCode: 0x4567, PosVec: types.Ptf(12.75, 34.25), Direction1: 64, VelVec: types.Ptf(4, -4)}
	if got := s.NetClientPredictLinear523530(obj); got != 0 {
		t.Fatalf("result=%d", got)
	}
	want := []byte{0xb5, 0x67, 0x45, 0x45, 0x23, 12, 0, 34, 0, 64, 0, 0, 64, 192}
	for i := ntype.PlayerInd(0); i < 4; i++ {
		packet := s.NetList.ByInd(i, netlist.Kind1).Get()
		if i == 0 || i == 3 {
			if !reflect.DeepEqual(packet, want) {
				t.Fatalf("player %d packet=% x, want % x", i, packet, want)
			}
		} else if len(packet) != 0 {
			t.Fatalf("absent/inactive player %d received packet", i)
		}
		if s.NetList.ByInd(i, netlist.Kind2).Count() != 0 {
			t.Fatal("prediction used wrong message queue")
		}
	}
}

func TestPredictLinear523530FaultPrefix(t *testing.T) {
	called := false
	defer func() {
		if recover() == nil || !called {
			t.Fatal("null object must fault after net-code callback")
		}
	}()
	predictLinear523530(nil, predictLinearDeps523530{netCode: func(*Object) uint32 { called = true; return 3 }})
}
