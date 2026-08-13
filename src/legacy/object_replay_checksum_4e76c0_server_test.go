package legacy

import (
	"math"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func float32FromBits4E7700(v uint32) float32 {
	return math.Float32frombits(v)
}

func TestObjectReplayChecksumNative4E7700NamedFieldMapping(t *testing.T) {
	health := &server.HealthData{Cur: 0xa003, Field2: 0x9002, Max: 0x8001, Field16: 0xdeadbeef}
	obj := &server.Object{
		TypeInd:     0x8001,
		ObjFlags:    object.Flags(0x31323334),
		Field5:      0x21222324,
		NetCode:     0x11121314,
		Extent:      0x01020304,
		ScriptIDVal: -2,
		TeamVal:     server.ObjectTeam{ID: server.TeamID(0x85)},
		PosVec:      types.Pointf{X: float32FromBits4E7700(0x80000000), Y: float32FromBits4E7700(0x7fc12345)},
		NewPos:      types.Pointf{X: float32FromBits4E7700(0x41424344), Y: float32FromBits4E7700(0x51525354)},
		PrevPos:     types.Pointf{X: float32FromBits4E7700(0x61626364), Y: float32FromBits4E7700(0x71727374)},
		VelVec:      types.Pointf{X: float32FromBits4E7700(0x81828384), Y: float32FromBits4E7700(0x91929394)},
		ForceVec:    types.Pointf{X: float32FromBits4E7700(0xa1a2a3a4), Y: float32FromBits4E7700(0xb1b2b3b4)},
		Pos24:       types.Pointf{X: float32FromBits4E7700(0xc1c2c3c4), Y: float32FromBits4E7700(0xd1d2d3d4)},
		ZVal:        float32FromBits4E7700(0xe1e2e3e4),
		Field27:     float32FromBits4E7700(0x7f812345),
		Mass:        float32FromBits4E7700(0xff800000),
		Direction1:  server.Dir16(0x8000),
		Direction2:  server.Dir16(0xfffe),
		Field32:     0x32323232,
		Field33:     0x33333333,
		Field34:     0x34343434,
		Field37:     0x37373737,
		Field38:     0x38383838,
		Buffs:       0x85858585,
		HealthData:  health,
	}
	obj.Field62[0] = 0x62626262
	obj.Field62[1] = 0xffffffff

	want := objectReplayChecksum4E7700(objectReplayChecksumInput4E7700{
		teamID:        0x85,
		typeInd:       0x8001,
		scriptID:      -2,
		posXBits:      0x80000000,
		extent:        0x01020304,
		netCode:       0x11121314,
		field5:        0x21222324,
		objFlags:      0x31323334,
		posYBits:      0x7fc12345,
		newPosXBits:   0x41424344,
		newPosYBits:   0x51525354,
		prevPosXBits:  0x61626364,
		prevPosYBits:  0x71727374,
		velXBits:      0x81828384,
		velYBits:      0x91929394,
		forceXBits:    0xa1a2a3a4,
		forceYBits:    0xb1b2b3b4,
		pos24XBits:    0xc1c2c3c4,
		pos24YBits:    0xd1d2d3d4,
		zBits:         0xe1e2e3e4,
		field27Bits:   0x7f812345,
		direction1:    -32768,
		direction2:    -2,
		field38:       0x38383838,
		field37:       0x37373737,
		field34:       0x34343434,
		field33:       0x33333333,
		field32:       0x32323232,
		massBits:      0xff800000,
		buffs:         0x85858585,
		field62:       0x62626262,
		healthPresent: true,
		healthMax:     0x8001,
		healthField2:  0x9002,
		healthCur:     0xa003,
	})
	if got := Sub_4E7700(obj); got != want {
		t.Fatalf("checksum = %#08x, want %#08x", got, want)
	}

	before := Sub_4E7700(obj)
	obj.ObjClass = object.Class(0xffffffff)
	obj.ObjSubClass = object.SubClass(0xffffffff)
	obj.TeamVal.Field0 = 0xffffffff
	obj.Material = 0xffff
	obj.Experience = float32FromBits4E7700(0xffffffff)
	obj.Worth = 0xffffffff
	obj.Field35 = 0xffffffff
	obj.Field36 = 0xffffffff
	obj.Field62[1] ^= 0xffffffff
	obj.BuffsDur[0] = 0xffff
	health.Field16 ^= 0xffffffff
	if got := Sub_4E7700(obj); got != before {
		t.Fatalf("checksum changed for unobserved fields: got %#08x, want %#08x", got, before)
	}
}

func TestObjectReplayChecksumNative4E7700NilHealth(t *testing.T) {
	obj := &server.Object{TeamVal: server.ObjectTeam{ID: 7}}
	if got := Sub_4E7700(obj); got != 7 {
		t.Fatalf("checksum = %#08x, want 0x00000007", got)
	}
}

func TestObjectReplayChecksumNative4E7700NilObjectFault(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil object did not fault")
		}
	}()
	Sub_4E7700(nil)
}

func TestObjectReplayChecksumPassNative4E76C0(t *testing.T) {
	a := &server.Object{TeamVal: server.ObjectTeam{ID: 1}}
	b := &server.Object{TeamVal: server.ObjectTeam{ID: 2}}
	a.ObjNext = b
	firstCalls := 0
	got := objectReplayChecksumPassNative4E76C0(func() *server.Object {
		firstCalls++
		return a
	})
	if got != nil {
		t.Fatalf("terminal object = %p, want nil", got)
	}
	if firstCalls != 1 {
		t.Fatalf("first calls = %d, want 1", firstCalls)
	}
	if a.ObjNext != b || b.ObjNext != nil {
		t.Fatalf("no-op/checksum pass mutated the object list")
	}
}

func TestSub4E76F0AcceptsNil(t *testing.T) {
	Sub_4E76F0(nil)
}
