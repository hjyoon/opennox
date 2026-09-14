package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func TestScriptMonsterGoHome512A00BuildsNativeActionStack(t *testing.T) {
	s := unitIdleTestServer515820(t)
	unit := monsterActionTestObject50A910(t)
	unit.serverHandle = s.handle
	unit.PosVec = types.Ptf(100.5, -20.25)
	update := unit.UpdateDataMonster()
	update.AIStackInd = 0
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_WAIT)}
	update.Direction94 = 32
	update.Pos95 = types.Pointf{
		X: math.Float32frombits(0x7fa12345),
		Y: math.Float32frombits(0x80000000),
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(update)) <= math.MaxUint32 {
		t.Fatalf("update pointer = %p, want native address above 4 GiB", update)
	}

	s.ScriptMonsterGoHome512A00(unit)

	if update.AIStackInd != 2 {
		t.Fatalf("AIStackInd = %d, want 2", update.AIStackInd)
	}
	report, face, home := &update.AIStack[0], &update.AIStack[1], &update.AIStack[2]
	if report.Type() != ai.ACTION_REPORT || report.Args != [4]uintptr{uintptr(ai.ACTION_MOVE_TO_HOME), 0, 0, 0} {
		t.Fatalf("report = %#v, want ACTION_REPORT with MOVE_TO_HOME argument", report)
	}
	cos, sin := SinCosDir(byte(update.Direction94))
	wantFaceX := math.Float32bits(float32(float64(cos)*10.0 + float64(unit.PosVec.X)))
	wantFaceY := math.Float32bits(float32(float64(sin)*10.0 + float64(unit.PosVec.Y)))
	if face.Type() != ai.ACTION_FACE_LOCATION || face.ArgU32(0) != wantFaceX || face.ArgU32(1) != wantFaceY ||
		face.Args[2] != 0 || face.Args[3] != 0 {
		t.Fatalf("face = %#v, want location bits (%#x, %#x)", face, wantFaceX, wantFaceY)
	}
	if home.Type() != ai.ACTION_MOVE_TO_HOME || home.ArgU32(0) != 0x7fa12345 ||
		home.ArgU32(1) != 0x80000000 || home.Args[2] != 0 || home.Args[3] != 0 {
		t.Fatalf("home = %#v, want exact saved home position and nil target", home)
	}
	if !s.AI.StackChanged {
		t.Fatal("native action-stack operations did not mark the stack changed")
	}
	runtime.KeepAlive(unit)
}

func TestScriptMonsterGoHome512A00PreservesOriginalGates(t *testing.T) {
	s := unitIdleTestServer515820(t)
	s.ScriptMonsterGoHome512A00(nil)
	s.ScriptMonsterGoHome512A00(&Object{ObjClass: object.ClassPlayer})
	s.ScriptMonsterGoHome512A00(&Object{ObjClass: object.ClassMonster, ObjFlags: object.FlagDead})
}

func TestScriptMonsterGoHome512A00DoesNotHideMissingUpdateData(t *testing.T) {
	s := unitIdleTestServer515820(t)
	defer func() {
		if recover() == nil {
			t.Fatal("eligible monster without UpdateData did not preserve the original fault contract")
		}
	}()
	s.ScriptMonsterGoHome512A00(&Object{ObjClass: object.ClassMonster, serverHandle: s.handle})
}
