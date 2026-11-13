package legacy

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Only S is used by the production runtime's RNG and clock bindings.
// Unexpected interface services remain unbound rather than simulated.
type dodgeProductionServer547C50 struct {
	Server
	core *server.Server
}

func (s dodgeProductionServer547C50) S() *server.Server { return s.core }

// Real native objects, default exported RNG/clock/Pop/Push bindings, with
// declared geometry observers. End/Cancel handlers remain inactive; this
// does not certify complete map geometry or all action callbacks.
func TestMonsterDodgeablesProduction547C50(t *testing.T) {
	rows := dodgeOriginalRows547C50(t)
	for _, path := range []string{"C", "Go"} {
		t.Run(path, func(t *testing.T) {
			count := 0
			for _, row := range rows {
				if !strings.HasPrefix(row.Input.Name, "speed/") || row.Input.Seed != 0 {
					continue
				}
				count++
				t.Run(row.Input.Name, func(t *testing.T) {
					s := server.NewHigh(nil, nil, strman.New())
					t.Cleanup(s.Close)
					if !s.Objs.InitHigh(1) {
						t.Fatal("native object allocator")
					}
					t.Cleanup(s.Objs.FreeObjects)
					unit := s.Objs.NewObject(&server.ObjectType{})
					update, free := alloc.NewHigh(server.MonsterUpdateData{})
					t.Cleanup(free)
					unit.ObjClass, unit.UpdateData = object.ClassMonster, unsafe.Pointer(update)
					unit.Direction1 = server.Dir16(row.Input.Direction)
					unit.PosVec = types.Ptf(math.Float32frombits(row.Input.Position[0]), math.Float32frombits(row.Input.Position[1]))
					unit.Shape.Circle.R = math.Float32frombits(row.Input.Radius176)
					unit.SpeedCur = math.Float32frombits(0x3f123456)
					update.AIStackInd = 0
					update.AIStack[0].Action = uint32(ai.ACTION_IDLE)
					s.SetFrame(row.Input.Frame)
					s.SetTickRate(row.Input.FPS)
					s.Rand.Logic = prand.New(0)
					other := s.Rand.Other.Index()
					before := *unit
					want := *update
					want.AIStackInd = 1 // Actual Push replaces the initial IDLE.
					want.AIStack[0] = server.AIStackItem{Action: uint32(ai.DEPENDENCY_TIME), Args: [4]uintptr{uintptr(row.Output.State.Pushes[0][1])}}
					want.AIStack[1] = server.AIStackItem{Action: uint32(ai.ACTION_DODGE), Args: [4]uintptr{uintptr(row.Output.Ray[2]), uintptr(row.Output.Ray[3]), 0}}
					want.Field124, want.Field137 = row.Input.Frame, row.Input.Frame
					oldServer := GetServer
					GetServer = func() Server { return dodgeProductionServer547C50{core: s} }
					defer func() { GetServer = oldServer }()
					if monsterDodgeRuntimeOverride547C50 != nil {
						t.Fatal("production runtime unexpectedly overridden")
					}
					runtime := monsterDodgeRuntime547C50()
					order := make([]string, 0)
					runtime.Trace = func(ray *[4]float32, flags uint32) int {
						order = append(order, "ray")
						for i, value := range ray {
							if math.Float32bits(value) != row.Output.Ray[i] {
								t.Fatalf("actual RNG/radius original ray[%d]=%08x want%08x", i, math.Float32bits(value), row.Output.Ray[i])
							}
						}
						if flags != 1 {
							t.Fatal("original trace flags")
						}
						return 1
					}
					runtime.Obstacles = func(obj *server.Object, ray *[4]float32) int {
						order = append(order, "obstacles")
						if obj != unit {
							t.Fatal("native obstacle unit")
						}
						return 1
					}
					runtime.Tile = func(point *types.Pointf) int {
						order = append(order, "tile")
						if math.Float32bits(point.X) != row.Output.Ray[2] || math.Float32bits(point.Y) != row.Output.Ray[3] {
							t.Fatal("original tile coordinates")
						}
						return 0
					}
					result := 0
					if path == "C" {
						if dodgeOriginalABI547C50() != 64 {
							t.Fatal("typed pointer ABI")
						}
						monsterDodgeRuntimeOverride547C50 = &runtime
						defer func() { monsterDodgeRuntimeOverride547C50 = nil }()
						result = dodgeOriginalC547C50(unit)
					} else {
						result = unit.MonsterCheckDodgeablesNative547C50(runtime.RandomFloat, runtime.RandomInt, runtime.Trace, runtime.Obstacles, runtime.Tile, runtime.Pop, runtime.Push, runtime.Frame, runtime.FPS)
					}
					if result != 1 || fmt.Sprint(order) != "[ray obstacles tile]" || s.Rand.Logic.Index() != 2 || s.Rand.Other.Index() != other || !s.AI.StackChanged {
						t.Fatal("production return/order/RNG/stack marker", result, order, s.Rand.Logic.Index())
					}
					for _, pair := range []struct {
						got, want unsafe.Pointer
						size      uintptr
					}{
						{unsafe.Pointer(unit), unsafe.Pointer(&before), unsafe.Sizeof(before)},
						{unsafe.Pointer(update), unsafe.Pointer(&want), unsafe.Sizeof(want)},
					} {
						if !bytes.Equal(unsafe.Slice((*byte)(pair.got), int(pair.size)), unsafe.Slice((*byte)(pair.want), int(pair.size))) {
							t.Fatal("whole native production records or padding differ")
						}
					}
				})
			}
			if count != 136 {
				t.Fatal("original production geometry context count", count)
			}
		})
	}
}
