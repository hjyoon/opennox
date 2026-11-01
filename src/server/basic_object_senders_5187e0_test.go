package server

import (
	"bytes"
	"github.com/opennox/libs/object"
	"testing"
	"unsafe"
)

func basicSenderOrderRuntime5187E0() BasicObjectSendRuntime5187E0 {
	return BasicObjectSendRuntime5187E0{
		Frame:        func() uint32 { return 100 },
		HealthDelta:  func(int32, uint16, int16) {},
		UnitCode:     func(*Object) uint32 { return 0x1234 },
		FloatToInt:   func(v float32) int32 { return int32(v) },
		Direction:    func(int32) IndexedDirectionVector509E20 { return IndexedDirectionVector509E20{X: 1} },
		ServerPacket: func(int32, []byte) int32 { return -2023406815 },
	}
}
func TestBasicSenderHealthSnapshotAndPostFrameUpdate5188A0(t *testing.T) {
	before, after := &MonsterGenUpdateData{}, &MonsterGenUpdateData{}
	before.HealthSamples[3], after.HealthSamples[3] = 211, 222
	oldHealth, newHealth := &HealthData{Cur: 199}, &HealthData{Cur: 444}
	unit := &Object{ObjClass: object.ClassMonsterGenerator, UpdateData: unsafe.Pointer(before), HealthData: oldHealth, NetCode: 0x2345, Frame134: 97}
	rt := basicSenderOrderRuntime5187E0()
	frames, reports := 0, 0
	rt.Frame = func() uint32 {
		frames++
		unit.HealthData = newHealth
		unit.UpdateData = unsafe.Pointer(after)
		return 100
	}
	rt.HealthDelta = func(index int32, code uint16, delta int16) {
		reports++
		if index != 3 || code != 0x2345 || delta != -23 {
			t.Fatalf("health snapshot/update timing %d/%x/%d", index, code, delta)
		}
	}
	if got := SimpleObjectSendNative5188A0(3, unit, rt); uint32(got) != 0x87654321 || frames != 1 || reports != 1 || before.HealthSamples[3] != 211 || after.HealthSamples[3] != 444 {
		t.Fatalf("return=%x frames=%d reports=%d samples=%d/%d", uint32(got), frames, reports, before.HealthSamples[3], after.HealthSamples[3])
	}
}
func TestBasicSenderDamageTimestampAfterFrame5188A0(t *testing.T) {
	update := &MonsterGenUpdateData{}
	update.HealthSamples[3] = 211
	unit := &Object{ObjClass: object.ClassMonsterGenerator, UpdateData: unsafe.Pointer(update), HealthData: &HealthData{Cur: 199}, Frame134: 97}
	rt := basicSenderOrderRuntime5187E0()
	frames := 0
	rt.Frame = func() uint32 { frames++; unit.Frame134 = 100; unit.UpdateData = nil; return 100 }
	rt.HealthDelta = func(int32, uint16, int16) { t.Fatal("damage timestamp read before Frame") }
	if got := SimpleObjectSendNative5188A0(3, unit, rt); uint32(got) != 0x87654321 || frames != 1 || update.HealthSamples[3] != 211 {
		t.Fatal("young damage or nil cached update was eagerly dereferenced")
	}
}
func TestBasicSenderTypeCoordinatesAndDirectionOrder5187E0(t *testing.T) {
	unit := &Object{TypeInd: 73}
	rt := basicSenderOrderRuntime5187E0()
	converts := 0
	rt.UnitCode = func(got *Object) uint32 {
		if got != unit {
			t.Fatal("code object")
		}
		unit.TypeInd = 91
		unit.PosVec.X = 256.75
		unit.PosVec.Y = -123.75
		unit.ObjClass = object.ClassMissile
		unit.ObjSubClass = 0x30
		unit.Direction1 = 0x8123
		return 0x1234
	}
	rt.FloatToInt = func(v float32) int32 {
		converts++
		if converts == 1 {
			if v != 256.75 {
				t.Fatal("X before code callback")
			}
			unit.PosVec.Y = 7.75
		} else if converts == 2 {
			if v != 7.75 {
				t.Fatal("Y cached before X converter")
			}
		} else {
			t.Fatal("extra conversion")
		}
		return int32(v)
	}
	rt.Direction = func(index int32) IndexedDirectionVector509E20 {
		if index != -32477 {
			t.Fatal("direction WORD sign extension/live reload", index)
		}
		unit.Direction1 = 0
		unit.ObjClass = 0
		return IndexedDirectionVector509E20{X: -1, Y: -1}
	}
	rt.ServerPacket = func(index int32, packet []byte) int32 {
		want := []byte{48, 0x34, 0x12, 73, 0, 0, 1, 7, 0, 0, 0x24}
		if index != 3 || !bytes.Equal(packet, want) {
			t.Fatalf("packet=%x original-order=%x index=%d", packet, want, index)
		}
		return -2023406815
	}
	if got := PhantomObjectSendNative5187E0(3, unit, rt); uint32(got) != 0x87654321 || converts != 2 {
		t.Fatal("full transport return or converter count")
	}
}
func TestBasicSenderOrdinaryGuardedHealthPolicy5188A0(t *testing.T) {
	for _, name := range []string{"missing cache", "negative sample slot", "past last sample slot"} {
		t.Run(name, func(t *testing.T) {
			update := &MonsterGenUpdateData{}
			unit := &Object{ObjClass: object.ClassMonsterGenerator, UpdateData: unsafe.Pointer(update), HealthData: &HealthData{Cur: 199}, Frame134: 97}
			index := int32(3)
			switch name {
			case "missing cache":
				unit.UpdateData = nil
			case "negative sample slot":
				index = -1
			case "past last sample slot":
				index = 32
			}
			rt := basicSenderOrderRuntime5187E0()
			rt.GuardedInputs = true
			rt.HealthDelta = func(int32, uint16, int16) { t.Fatal("ordinary invalid sample policy") }
			rt.ServerPacket = func(got int32, packet []byte) int32 {
				if got != index || len(packet) != 9 || packet[0] != 47 {
					t.Fatal("guarded policy publication")
				}
				return 1
			}
			if SimpleObjectSendNative5188A0(index, unit, rt) != 1 {
				t.Fatal("ordinary nil/sample policy")
			}
		})
	}
}
