package opennox

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Public synthetic wall input supplies one ordinary blocking definition, not
// a substituted CanInteract/ray result. Production wall parsing is used.
func aiVisibleDependencyWall546A70(t *testing.T, s *Server) {
	t.Helper()
	var data bytes.Buffer
	word := func(value uint32) {
		if err := binary.Write(&data, binary.LittleEndian, value); err != nil {
			t.Fatal(err)
		}
	}
	alignedZero := func() {
		for data.Len()%8 != 0 {
			data.WriteByte(0)
		}
		data.Write(make([]byte, 8))
	}
	word(0)
	name := "visible-dependency-wall"
	data.WriteByte(byte(len(name)))
	data.WriteString(name)
	word(0) // ordinary opaque wall flags
	word(0)
	word(0)
	data.WriteByte(0)
	data.WriteByte(100)
	alignedZero()                  // no bricks
	data.Write([]byte{0, 0, 0, 0}) // sounds and Field749
	for direction := server.WallDirUp; direction <= server.WallDirLeftHalfArrowUp; direction++ {
		alignedZero() // no sprites required by geometric ray tracing
	}
	word(0x454e4420)
	raw, _ := alloc.Make([]byte{}, data.Len())
	copy(raw, data.Bytes())
	f := binfile.NewMemFile(unsafe.Pointer(&raw[0]), len(raw))
	defer f.Free()
	if err := s.Walls.ReadWall(f); err != nil {
		t.Fatal(err)
	}
	if s.Walls.DefByInd(0) == nil || s.Walls.DefByInd(0).Flags32 != 0 {
		t.Fatal("ordinary wall definition was not loaded")
	}
}

func aiVisibleDependencyNative546A70(t *testing.T) (*Server, *server.Object, *server.Object) {
	t.Helper()
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	if !s.Objs.Init(2) || s.Walls.Init() == 0 {
		t.Fatal("cannot initialize native object/wall allocation")
	}
	t.Cleanup(s.Objs.FreeObjects)
	t.Cleanup(s.Walls.Free)
	s.Map.Init()
	t.Cleanup(s.Map.Free)
	aiVisibleDependencyWall546A70(t, s)
	oldEngine := noxflags.GetEngine() & noxflags.EngineShowAI
	noxflags.UnsetEngine(noxflags.EngineShowAI)
	t.Cleanup(func() { noxflags.SetEngine(oldEngine) })
	s.SetFrame(702)
	unit, target := s.Objs.NewObject(&server.ObjectType{}), s.Objs.NewObject(&server.ObjectType{})
	update, free := alloc.New(server.MonsterUpdateData{})
	t.Cleanup(free)
	unit.ObjClass, unit.UpdateData = object.ClassMonster, unsafe.Pointer(update)
	target.ObjClass, target.ObjFlags = object.ClassFood, object.FlagActive
	for _, pointer := range []unsafe.Pointer{unsafe.Pointer(unit), unit.UpdateData, unsafe.Pointer(target)} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
			t.Fatalf("native AI object/update below 4 GiB: %p", pointer)
		}
	}
	return s, unit, target
}

func TestAIVisibleLocationDependency546A70NativeStack(t *testing.T) {
	s, unit, target := aiVisibleDependencyNative546A70(t)
	for _, descending := range []bool{false, true} {
		for _, mode := range []string{"nil-target", "blind-unit", "ordinary-target"} {
			for _, blocked := range []bool{false, true} {
				for _, withOR := range []bool{false, true} {
					t.Run(fmt.Sprintf("descending-%t/%s/blocked-%t/or-%t", descending, mode, blocked, withOR), func(t *testing.T) {
						s.Walls.Reset()
						unit.PosVec, target.PosVec = types.Ptf(300, 300), types.Ptf(428, 396)
						if descending {
							unit.PosVec.Y, target.PosVec.Y = target.PosVec.Y, unit.PosVec.Y
						}
						unit.Buffs = 0
						if mode == "blind-unit" {
							unit.Buffs = 1 << server.ENCHANT_BLINDED
						}
						if blocked {
							wall := s.Walls.CreateAtGrid(image.Pt(15, 15))
							if wall == nil {
								t.Fatal("cannot create indexed wall")
							}
							wall.Dir0 = byte(server.WallDirCross)
						}
						var selected *server.Object
						if mode != "nil-target" {
							selected = target
						}
						visible := mode == "ordinary-target" && !blocked
						if selected != nil && s.CanInteract(unit, selected, 0) != visible {
							t.Fatal("native CanInteract fixture differs from the literal visibility case")
						}
						if got := s.MapTraceRay(unit.PosVec, target.PosVec, server.MapTraceFlag1); got != !blocked {
							t.Fatalf("native wall ray clear=%t want=%t", got, !blocked)
						}
						update := unit.UpdateDataMonster()
						*update = server.MonsterUpdateData{Field124: 11, Field137: 12}
						base := server.AIStackItem{Action: uint32(ai.ACTION_WAIT), Args: [4]uintptr{7, 8, 9, 10}}
						update.AIStack[0] = base
						conditionIndex := 1
						if withOR {
							update.AIStack[1] = server.AIStackItem{Action: uint32(ai.DEPENDENCY_TIME)} // false companion
							conditionIndex = 2
						}
						condition := &update.AIStack[conditionIndex]
						*condition = server.AIStackItem{Action: uint32(ai.DEPENDENCY_OBJECT_AT_VISIBLE_LOCATION), Field5: 51}
						condition.SetArgs(target.PosVec, unsafe.Pointer(selected), uint32(0xfedcba98))
						headIndex := conditionIndex + 1
						if withOR {
							update.AIStack[headIndex] = server.AIStackItem{Action: uint32(ai.DEPENDENCY_OR)}
							headIndex++
						}
						head := server.AIStackItem{Action: uint32(ai.ACTION_WAIT), Args: [4]uintptr{900, 77, 88, 99}}
						update.AIStack[headIndex], update.AIStackInd = head, int8(headIndex)
						before := *condition
						s.AI.StackChanged = false
						(&aiData{s: s}).nox_xxx_mobActionDependency(unit)
						// 00546E65..00546E75 fails only for a clear ray and no
						// interaction. A blocked ray retains the action even
						// with no target. Expectations are not derived from Go.
						wantRetained := blocked || mode == "ordinary-target"
						if wantRetained {
							if int(update.AIStackInd) != headIndex || update.AIStack[headIndex] != head || s.AI.StackChanged || update.Field124 != 11 || update.Field137 != 12 {
								t.Fatalf("original retained condition popped/reset the action: index=%d head=%+v changed=%t reset=%d/%d", update.AIStackInd, update.AIStackHead(), s.AI.StackChanged, update.Field124, update.Field137)
							}
						} else if update.AIStackInd != 0 || !s.AI.StackChanged || update.Field124 != 702 || update.Field137 != 702 {
							t.Fatalf("failed condition did not perform real native stack pop/reset: index=%d changed=%t reset=%d/%d", update.AIStackInd, s.AI.StackChanged, update.Field124, update.Field137)
						}
						if *condition != before || update.AIStack[0] != base {
							t.Fatal("condition payload or preceding action was damaged")
						}
					})
				}
			}
		}
	}
}

func TestAIVisibleLocationDependency546A70NativeLivePosition(t *testing.T) {
	s, unit, target := aiVisibleDependencyNative546A70(t)
	for _, descending := range []bool{false, true} {
		t.Run(fmt.Sprintf("descending-%t", descending), func(t *testing.T) {
			unit.PosVec, target.PosVec = types.Ptf(300, 300), types.Ptf(428, 396)
			if descending {
				unit.PosVec.Y, target.PosVec.Y = target.PosVec.Y, unit.PosVec.Y
			}
			update := unit.UpdateDataMonster()
			*update = server.MonsterUpdateData{AIStackInd: 2}
			update.AIStack[0].Action = uint32(ai.ACTION_WAIT)
			condition := &update.AIStack[1]
			condition.Action = uint32(ai.DEPENDENCY_OBJECT_AT_VISIBLE_LOCATION)
			condition.SetArgs(types.Ptf(412, 400), unsafe.Pointer(target), uint32(0x12345678))
			update.AIStack[2].Action = uint32(ai.ACTION_MOVE_TO)
			before := *condition
			if !s.CanInteract(unit, target, 0) {
				t.Fatal("ordinary live target not visible")
			}
			(&aiData{s: s}).nox_xxx_mobActionDependency(unit)
			if update.AIStackInd != 2 || condition.ArgPos(0) != target.PosVec || condition.Args[2] != before.Args[2] || condition.Args[3] != before.Args[3] || condition.Field5 != before.Field5 {
				t.Fatalf("live target position/native identity not retained: index=%d condition=%+v target=%p pos=%v", update.AIStackInd, condition, target, target.PosVec)
			}
		})
	}
}
