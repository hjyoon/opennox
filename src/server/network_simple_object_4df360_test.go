package server

import (
	"bytes"
	"math"
	"reflect"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestNetSendSimpleObject4DF360PacketAndCallbackOrder(t *testing.T) {
	s := new(Server)
	obj := &Object{TypeInd: 0xabcd, NetCode: 0x1234, PosVec: types.Pointf{X: 12.5, Y: 33}}
	var rounded []float32
	round := func(value float32) int32 {
		rounded = append(rounded, value)
		if len(rounded) == 1 {
			// Type/code/X are already read; Y is read after this callback.
			obj.TypeInd = 2
			obj.NetCode = 3
			obj.PosVec = types.Pointf{X: 100, Y: -7.5}
			return 0x12345
		}
		return -9
	}
	s.NetSendPacketXxx = func(recipient int, packet []byte, related *Object, remove, sequence int) int {
		if recipient != -0x123456 || related != nil || remove != 1 || sequence != 1 {
			t.Fatalf("routing=%d/%p/%d/%d", recipient, related, remove, sequence)
		}
		want := []byte{0x2f, 0x34, 0x12, 0xcd, 0xab, 0x45, 0x23, 0xf7, 0xff}
		if !bytes.Equal(packet, want) {
			t.Fatalf("packet=%x, want %x", packet, want)
		}
		return -17
	}
	if got := s.NetSendSimpleObject4DF360(-0x123456, obj, round); got != -17 {
		t.Fatalf("return=%d, want -17", got)
	}
	if !reflect.DeepEqual(rounded, []float32{12.5, -7.5}) {
		t.Fatalf("coordinate loads=%v", rounded)
	}
}

func TestNetSendSimpleObject4DF360StaticWireCode(t *testing.T) {
	for _, class := range []object.Class{object.ClassClientPersist, object.ClassImmobile} {
		s := new(Server)
		obj := &Object{ObjClass: class, TypeInd: 9, NetCode: 4, Extent: 0x1234, PosVec: types.Pointf{X: 2.5, Y: -3.5}}
		s.NetSendPacketXxx = func(recipient int, packet []byte, related *Object, remove, sequence int) int {
			want := []byte{0x2f, 0x34, 0x92, 9, 0, 2, 0, 0xfc, 0xff}
			if recipient != 31 || !bytes.Equal(packet, want) || related != nil || remove != 1 || sequence != 1 {
				t.Fatalf("class=%#x routing=%d/%p/%d/%d packet=%x", class, recipient, related, remove, sequence, packet)
			}
			return 1
		}
		if got := s.NetSendSimpleObject4DF360(31, obj, func(value float32) int32 { return int32(math.RoundToEven(float64(value))) }); got != 1 {
			t.Fatalf("return=%d", got)
		}
	}
}
