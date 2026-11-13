package server

import (
	"bytes"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// Literal original e3f context direction/0, sealed before engine changes.
// Actual unit+176 is Shape.Circle.R; SpeedCur is the unrelated unit+544.
// The original outputs are not calculated by a production arithmetic helper.
func TestMonsterDodgeablesOriginalRadius547C50(t *testing.T) {
	s := NewHigh(nil, nil, strman.New())
	t.Cleanup(s.Close)
	if !s.Objs.InitHigh(1) {
		t.Fatal("native object allocator")
	}
	t.Cleanup(s.Objs.FreeObjects)
	unit := s.Objs.NewObject(&ObjectType{})
	update, free := alloc.NewHigh(MonsterUpdateData{})
	t.Cleanup(free)
	unit.ObjClass, unit.UpdateData = object.ClassMonster, unsafe.Pointer(update)
	unit.PosVec = types.Ptf(math.Float32frombits(1153884160), math.Float32frombits(1167323136))
	unit.Direction1 = 0
	unit.Shape.Circle.R = math.Float32frombits(1082130432)
	unit.SpeedCur = math.Float32frombits(0x3f123456)
	update.AIStackInd = 0
	update.AIStack[0].Action = uint32(ai.ACTION_GUARD)
	s.SetFrame(0xfffffff0)
	s.SetTickRate(30)
	before := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(unit)), int(unsafe.Sizeof(*unit))))
	want := types.Ptf(math.Float32frombits(1153884160), math.Float32frombits(1167299390))
	probes := 0
	runtime := MonsterMainRuntime547210{
		RandomFloat: func(min, max float32) float64 {
			if min != 2 || max != 3 {
				t.Fatal("original float bounds")
			}
			return math.Float64frombits(4613709456448946176)
		},
		RandomInt: func(min, max int) int {
			if min != 0 || max != 100 {
				t.Fatal("original integer bounds")
			}
			return 38
		},
		TraceRay: func(from, to types.Pointf, flags MapTraceFlags) bool {
			probes++
			if from != unit.PosVec || math.Float32bits(to.X) != 1153884160 || math.Float32bits(to.Y) != 1167299390 || flags != MapTraceFlag1 {
				t.Errorf("original radius176 destination mismatch: got=%08x/%08x want=%08x/%08x radius=%08x unrelatedSpeed=%08x", math.Float32bits(to.X), math.Float32bits(to.Y), uint32(1153884160), uint32(1167299390), math.Float32bits(unit.Shape.Circle.R), math.Float32bits(unit.SpeedCur))
			}
			return true
		},
		TraceObstacles: func(got *Object, from, to types.Pointf) bool {
			if got != unit || from != unit.PosVec || to != want {
				t.Error("original obstacle destination")
			}
			return true
		},
		TileAt: func(to types.Pointf) int {
			if to != want {
				t.Error("original tile destination")
			}
			return 0
		},
	}
	if !s.monsterMainCheckDodgeables547C50(unit, runtime) || probes != 1 {
		t.Fatal("one original successful attempt")
	}
	if update.AIStackInd != 2 || update.AIStack[1].Type() != ai.DEPENDENCY_TIME || update.AIStack[1].ArgU32(0) != 14 || update.AIStack[2].Type() != ai.ACTION_DODGE || update.AIStack[2].ArgPos(0) != want || update.AIStack[2].ArgU32(2) != 0 {
		t.Fatal("original wrapped timer and literal destination", update.GetAIStack())
	}
	if !bytes.Equal(before, unsafe.Slice((*byte)(unsafe.Pointer(unit)), int(unsafe.Sizeof(*unit)))) {
		t.Fatal("whole native object changed")
	}
}
