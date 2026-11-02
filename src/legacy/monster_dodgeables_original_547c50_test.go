package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type dodgeMutation547C50 struct {
	Ray       *[4]uint32
	Position  *[2]uint32
	Radius176 *uint32 `json:"speed"` // Frozen wire name; original +176 is Shape.Circle.R, never SpeedCur (+544).
	Direction *uint16
	Frame     *uint32
	FPS       *uint32
}
type dodgeAttempt547C50 struct {
	Trace, Obstacle, Tile uint32
	Mutations             map[string]dodgeMutation547C50
}
type dodgeInput547C50 struct {
	Name      string
	Direction uint16
	Position  [2]uint32
	Radius176 uint32 `json:"speed"`
	Seed      uint32
	Frame     uint32
	FPS       uint32
	Index     int8
	Actions   [24]uint32
	Prefix    uint32
	PopSteps  []struct {
		Index  int8
		Prefix uint32
		Update int
	} `json:"pop_steps"`
	NilUnit       bool                           `json:"nil_unit"`
	NilUpdate     bool                           `json:"nil_update"`
	PushNull      [2]bool                        `json:"push_null"`
	PushMutations map[string]dodgeMutation547C50 `json:"push_mutations"`
	Attempts      [5]dodgeAttempt547C50
}
type dodgeState547C50 struct {
	Position  [2]uint32    `json:"position"`
	Radius176 uint32       `json:"speed"`
	Direction uint16       `json:"direction"`
	Update    int          `json:"update"`
	Index     int8         `json:"index"`
	Prefix    uint32       `json:"prefix"`
	Actions   [24]uint32   `json:"actions"`
	Frame     uint32       `json:"frame"`
	FPS       uint32       `json:"fps"`
	RNG       uint32       `json:"rng"`
	Pushes    [2][6]uint32 `json:"pushes"`
}
type dodgeWrite547C50 struct {
	Target [2]json.RawMessage `json:"target"`
	Size   int                `json:"size"`
	Value  uint64             `json:"value"`
}
type dodgeRow547C50 struct {
	Input  dodgeInput547C50
	Output struct {
		Events []json.RawMessage
		Writes []dodgeWrite547C50
		State  dodgeState547C50
		Ray    []uint32
		Fault  []json.RawMessage
		Result *uint32
	}
}

func dodgeOriginalRows547C50(t *testing.T) []dodgeRow547C50 {
	t.Helper()
	raw, err := os.ReadFile("../server/testdata/monster_dodgeables_original_547c50.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "c0cec2ea8230ea72deb7cba8a0ed850bf895b57bf2af04d92edf9c8cebae9166" {
		t.Fatal("sealed original gzip changed")
	}
	z, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err = z.Close(); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(decoded)) != "c6513a855625a0deb4e0d978230bcc123689e3489b8eebbd1fce06ffe9b79579" {
		t.Fatal("sealed original decoded bytes changed")
	}
	var data struct {
		Rows      []dodgeRow547C50
		Auxiliary []json.RawMessage
	}
	if err = json.Unmarshal(decoded, &data); err != nil {
		t.Fatal(err)
	}
	if len(data.Rows) != 2361 || len(data.Auxiliary) != 4 {
		t.Fatal("original context counts")
	}
	return data.Rows
}

type dodgeWorld547C50 struct {
	Unit                     *server.Object
	Data                     *server.MonsterUpdateData
	Pushes                   [2]*server.AIStackItem
	Heap                     []byte
	Frame, FPS, RNG          uint32
	Ray                      *[4]float32
	Attempt, Pops, PushCount int
	Events                   []any
	Writes                   []dodgeWrite547C50
}

func newDodgeWorld547C50(t *testing.T, row dodgeRow547C50) *dodgeWorld547C50 {
	t.Helper()
	p, free := alloc.Malloc(32768)
	t.Cleanup(free)
	heap := unsafe.Slice((*byte)(p), 32768)
	for i := range heap {
		heap[i] = byte(i*37 + 0xa5)
	}
	w := &dodgeWorld547C50{Unit: (*server.Object)(unsafe.Add(p, 1024)), Data: (*server.MonsterUpdateData)(unsafe.Add(p, 4096)), Pushes: [2]*server.AIStackItem{(*server.AIStackItem)(unsafe.Add(p, 16384)), (*server.AIStackItem)(unsafe.Add(p, 16896))}, Heap: heap, Frame: row.Input.Frame, FPS: row.Input.FPS, RNG: row.Input.Seed, Attempt: -1, Events: make([]any, 0), Writes: make([]dodgeWrite547C50, 0)}
	w.Unit.ObjClass = object.ClassMonster
	w.Unit.PosVec = types.Ptf(math.Float32frombits(row.Input.Position[0]), math.Float32frombits(row.Input.Position[1]))
	w.Unit.Direction1 = server.Dir16(row.Input.Direction)
	w.Unit.Shape.Circle.R = math.Float32frombits(row.Input.Radius176)
	// A different, guarded value makes an accidental SpeedCur read observable.
	w.Unit.SpeedCur = math.Float32frombits(0x3f123456)
	w.Unit.UpdateData = unsafe.Pointer(w.Data)
	if row.Input.NilUpdate {
		w.Unit.UpdateData = nil
	}
	w.Data.AIStackInd, w.Data.Field132 = row.Input.Index, row.Input.Prefix
	word := uintptr(0xa5a5a5a5)
	if unsafe.Sizeof(word) == 8 {
		high := uint64(0x12345678)
		word |= uintptr(high << 32)
	}
	for i := range w.Data.AIStack {
		w.Data.AIStack[i].Action = row.Input.Actions[i]
	}
	for _, item := range w.Pushes {
		item.Action, item.Field5 = 0xa5a5a5a5, 0xa5a5a5a5
		for i := range item.Args {
			item.Args[i] = word
		}
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(w.Unit), unsafe.Pointer(w.Data), unsafe.Pointer(w.Pushes[0]), unsafe.Pointer(w.Pushes[1])} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatal("native C-owned address must exceed4GiB")
			}
		}
	}
	return w
}
func (w *dodgeWorld547C50) snapshot() dodgeState547C50 {
	x := dodgeState547C50{Position: [2]uint32{math.Float32bits(w.Unit.PosVec.X), math.Float32bits(w.Unit.PosVec.Y)}, Radius176: math.Float32bits(w.Unit.Shape.Circle.R), Direction: uint16(w.Unit.Direction1), Update: -1, Index: w.Data.AIStackInd, Prefix: w.Data.Field132, Frame: w.Frame, FPS: w.FPS, RNG: w.RNG}
	if w.Unit.UpdateData != nil {
		if w.Unit.UpdateData != unsafe.Pointer(w.Data) {
			panic("invalid native MUD role")
		}
		x.Update = 0
	}
	for i := range w.Data.AIStack {
		x.Actions[i] = w.Data.AIStack[i].Action
	}
	for i, item := range w.Pushes {
		x.Pushes[i][0] = item.Action
		for j := range item.Args {
			x.Pushes[i][j+1] = uint32(item.Args[j])
		}
		x.Pushes[i][5] = item.Field5
	}
	return x
}
func (w *dodgeWorld547C50) rayBits() []uint32 {
	if w.Ray == nil {
		return nil
	}
	out := make([]uint32, 4)
	for i, v := range w.Ray {
		out[i] = math.Float32bits(v)
	}
	return out
}
func (w *dodgeWorld547C50) record(namespace string, field any, size int, value uint32) {
	a, _ := json.Marshal(namespace)
	b, _ := json.Marshal(field)
	w.Writes = append(w.Writes, dodgeWrite547C50{Target: [2]json.RawMessage{a, b}, Size: size, Value: uint64(value)})
}
func (w *dodgeWorld547C50) mutate(m dodgeMutation547C50) {
	if m.Ray != nil {
		if w.Ray == nil {
			panic("ray mutation without pointer")
		}
		for i, v := range m.Ray {
			w.Ray[i] = math.Float32frombits(v)
		}
	}
	if m.Position != nil {
		w.Unit.PosVec = types.Ptf(math.Float32frombits(m.Position[0]), math.Float32frombits(m.Position[1]))
		for i, v := range m.Position {
			w.record("unit", 56+4*i, 4, v)
		}
	}
	if m.Radius176 != nil {
		w.Unit.Shape.Circle.R = math.Float32frombits(*m.Radius176)
		w.record("unit", 176, 4, *m.Radius176)
	}
	if m.Direction != nil {
		w.Unit.Direction1 = server.Dir16(*m.Direction)
		w.record("unit", 124, 2, uint32(*m.Direction))
	}
	if m.Frame != nil {
		w.Frame = *m.Frame
		w.record("world", "frame", 4, w.Frame)
	}
	if m.FPS != nil {
		w.FPS = *m.FPS
		w.record("world", "fps", 4, w.FPS)
	}
}
func dodgeCanonical547C50(t *testing.T, value any) any {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var out any
	if err = d.Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out
}
func (w *dodgeWorld547C50) randomResult(t *testing.T, row dodgeRow547C50, name string) uint64 {
	t.Helper()
	if len(w.Events) >= len(row.Output.Events) {
		t.Fatal("extra RNG callback")
	}
	var expected struct {
		Service string
		Result  uint64
		Before  uint32
	}
	if err := json.Unmarshal(row.Output.Events[len(w.Events)], &expected); err != nil {
		t.Fatal(err)
	}
	if expected.Service != name || expected.Before != w.RNG {
		t.Fatal("original RNG identity/index order", expected, w.RNG)
	}
	return expected.Result
}
func (w *dodgeWorld547C50) advanceRNG() {
	w.RNG++
	w.record("world", "rng", 4, w.RNG)
	if w.RNG >= 4096 {
		w.RNG = 0
		w.record("world", "rng", 4, 0)
	}
}
func (w *dodgeWorld547C50) runtime(t *testing.T, row dodgeRow547C50) monsterDodgeOriginalRuntime547C50 {
	return monsterDodgeOriginalRuntime547C50{
		RandomFloat: func(min, max float32) float64 {
			result := w.randomResult(t, row, "FloatRandom416030")
			before := w.RNG
			w.advanceRNG()
			w.Events = append(w.Events, map[string]any{"service": "FloatRandom416030", "min": math.Float32bits(min), "max": math.Float32bits(max), "before": before, "after": w.RNG, "result": result})
			return math.Float64frombits(result)
		},
		RandomInt: func(min, max int) int {
			result := w.randomResult(t, row, "IntRandom415FA0")
			before := w.RNG
			w.advanceRNG()
			w.Events = append(w.Events, map[string]any{"service": "IntRandom415FA0", "min": uint32(min), "max": uint32(max), "before": before, "after": w.RNG, "result": result})
			return int(int32(uint32(result)))
		},
		Trace: func(ray *[4]float32, flags uint32) int {
			w.Attempt++
			if w.Attempt >= 5 {
				t.Fatal("more than five attempts")
			}
			if w.Ray == nil {
				w.Ray = ray
			}
			if ray != w.Ray {
				t.Fatal("local ray identity changed")
			}
			step := row.Input.Attempts[w.Attempt]
			w.Events = append(w.Events, map[string]any{"service": "Trace535250", "attempt": w.Attempt, "ray": w.rayBits(), "flags": []uint32{0, 0, flags}, "before": w.snapshot(), "result": step.Trace})
			w.mutate(step.Mutations["trace"])
			return int(int32(step.Trace))
		},
		Obstacles: func(unit *server.Object, ray *[4]float32) int {
			if unit != w.Unit || ray != w.Ray {
				t.Fatal("obstacle pointer identity")
			}
			step := row.Input.Attempts[w.Attempt]
			w.Events = append(w.Events, map[string]any{"service": "Obstacles50B580", "attempt": w.Attempt, "ray": w.rayBits(), "before": w.snapshot(), "result": step.Obstacle})
			w.mutate(step.Mutations["obstacle"])
			return int(int32(step.Obstacle))
		},
		Tile: func(point *types.Pointf) int {
			if w.Ray == nil || unsafe.Pointer(point) != unsafe.Pointer(&w.Ray[2]) {
				t.Fatal("tile destination does not alias same local ray")
			}
			step := row.Input.Attempts[w.Attempt]
			w.Events = append(w.Events, map[string]any{"service": "Tile411160", "attempt": w.Attempt, "point": w.rayBits()[2:], "before": w.snapshot(), "result": step.Tile})
			w.mutate(step.Mutations["tile"])
			return int(int32(step.Tile))
		},
		Pop: func(unit *server.Object) int {
			if unit != w.Unit || w.Pops >= len(row.Input.PopSteps) {
				t.Fatal("unexpected actual attack-pop child service")
			}
			step := row.Input.PopSteps[w.Pops]
			w.Pops++
			w.Events = append(w.Events, map[string]any{"service": "Pop50A160", "before": w.snapshot(), "result": uint32(0xa5a55a5a)})
			w.Data.AIStackInd = step.Index
			w.record("mud", 544, 1, uint32(uint8(step.Index)))
			w.Data.Field132 = step.Prefix
			w.record("mud", 528, 4, step.Prefix)
			value := uint32(0)
			w.Unit.UpdateData = nil
			if step.Update == 0 {
				w.Unit.UpdateData = unsafe.Pointer(w.Data)
				value = 0x21003000
			}
			w.record("unit", 748, 4, value)
			return int(int32(-1515890086))
		},
		Push: func(unit *server.Object, action ai.ActionType) *server.AIStackItem {
			if unit != w.Unit || w.PushCount >= 2 {
				t.Fatal("push pointer/order")
			}
			index := w.PushCount
			w.PushCount++
			want, key := ai.DEPENDENCY_TIME, "push41"
			if index == 1 {
				want, key = ai.ACTION_DODGE, "push9"
			}
			if action != want {
				t.Fatal("push action order", action, want)
			}
			result := index
			if row.Input.PushNull[index] {
				result = -1
			}
			w.Events = append(w.Events, map[string]any{"service": "Push50A260", "action": uint32(action), "before": w.snapshot(), "result": result})
			w.mutate(row.Input.PushMutations[key])
			if result < 0 {
				return nil
			}
			return w.Pushes[index]
		},
		Frame: func() uint32 {
			w.Events = append(w.Events, map[string]any{"service": "Frame", "result": w.Frame})
			return w.Frame
		},
		FPS: func() uint32 {
			w.Events = append(w.Events, map[string]any{"service": "FPS", "result": w.FPS})
			return w.FPS
		},
	}
}

// Apply the ORIGINAL write log independently of the input/service adapter.
// Root stores are checked through final native bytes and the next service's
// before-state; only RNG/service writes are directly logged by this observer.
func (w *dodgeWorld547C50) expected(t *testing.T, row dodgeRow547C50) ([]byte, []dodgeWrite547C50) {
	t.Helper()
	out := bytes.Clone(w.Heap)
	service := make([]dodgeWrite547C50, 0)
	for _, wr := range row.Output.Writes {
		var namespace string
		if err := json.Unmarshal(wr.Target[0], &namespace); err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(namespace, "push") {
			service = append(service, wr)
		}
		if namespace == "world" {
			continue
		}
		var field int
		if err := json.Unmarshal(wr.Target[1], &field); err != nil {
			t.Fatal(err)
		}
		var ptr unsafe.Pointer
		size := wr.Size
		value := wr.Value
		switch namespace {
		case "unit":
			switch field {
			case 56:
				ptr = unsafe.Pointer(&w.Unit.PosVec.X)
			case 60:
				ptr = unsafe.Pointer(&w.Unit.PosVec.Y)
			case 124:
				ptr = unsafe.Pointer(&w.Unit.Direction1)
			case 176:
				ptr = unsafe.Pointer(&w.Unit.Shape.Circle.R)
			case 748:
				ptr = unsafe.Pointer(&w.Unit.UpdateData)
				size = int(unsafe.Sizeof(uintptr(0)))
				if value == 0x21003000 {
					value = uint64(uintptr(unsafe.Pointer(w.Data)))
				} else if value != 0 {
					t.Fatal("unclassified original MUD pointer")
				}
			default:
				t.Fatal("unclassified original unit write", field)
			}
		case "mud":
			if field == 544 {
				ptr = unsafe.Pointer(&w.Data.AIStackInd)
			} else if field == 528 {
				ptr = unsafe.Pointer(&w.Data.Field132)
			} else {
				t.Fatal("unclassified MUD store", field)
			}
		case "push0", "push1":
			index, _ := strconv.Atoi(namespace[4:])
			if field < 4 || field > 12 || field%4 != 0 {
				t.Fatal("unclassified stack-argument store", field)
			}
			ptr = unsafe.Pointer(&w.Pushes[index].Args[field/4-1])
			size = int(unsafe.Sizeof(uintptr(0)))
		default:
			t.Fatal("unclassified original store", namespace)
		}
		base := uintptr(unsafe.Pointer(&w.Heap[0]))
		off := uintptr(ptr) - base
		if ptr == nil || off+uintptr(size) > uintptr(len(out)) {
			t.Fatal("projected original store outside native allocation")
		}
		var raw [8]byte
		binary.LittleEndian.PutUint64(raw[:], value)
		copy(out[int(off):int(off)+size], raw[:size])
	}
	return out, service
}
func TestMonsterDodgeablesNativeLayout547C50(t *testing.T) {
	o := server.Object{}
	end := unsafe.Offsetof(o.ScriptPickup) + unsafe.Sizeof(o.ScriptPickup)
	a, _ := reflect.TypeOf(o).FieldByName("serverHandle")
	b, _ := reflect.TypeOf(o).FieldByName("objectHandle")
	if a.Offset != end || b.Offset != end+unsafe.Sizeof(uintptr(0)) || unsafe.Sizeof(o) != end+2*unsafe.Sizeof(uintptr(0)) {
		t.Fatal("shared C prefix and Go-only EXT handles")
	}
	want := [13]uintptr{end, unsafe.Offsetof(o.PosVec) + unsafe.Offsetof(types.Pointf{}.X), unsafe.Offsetof(o.PosVec) + unsafe.Offsetof(types.Pointf{}.Y), unsafe.Offsetof(o.Direction1), unsafe.Offsetof(o.Shape) + unsafe.Offsetof(server.Shape{}.Circle), unsafe.Offsetof(o.SpeedCur), unsafe.Offsetof(o.UpdateData), unsafe.Offsetof(server.MonsterUpdateData{}.Field132), unsafe.Offsetof(server.MonsterUpdateData{}.AIStackInd), unsafe.Offsetof(server.MonsterUpdateData{}.AIStack), unsafe.Sizeof(server.AIStackItem{}), unsafe.Offsetof(server.AIStackItem{}.Args), unsafe.Offsetof(server.AIStackItem{}.Field5)}
	if got := dodgeOriginalLayout547C50(); got != want {
		t.Fatal("independent C/Go offsets", got, want)
	}
	if want[4] == want[5] {
		t.Fatal("original radius176 must differ from unrelated speed544")
	}
}

type dodgeEntry547C50 interface {
	MonsterCheckDodgeablesNative547C50(func(float32, float32) float64, func(int, int) int, func(*[4]float32, uint32) int, func(*server.Object, *[4]float32) int, func(*types.Pointf) int, func(*server.Object) int, func(*server.Object, ai.ActionType) *server.AIStackItem, func() uint32, func() uint32) int
}

func TestMonsterDodgeablesOriginal547C50(t *testing.T) {
	for _, path := range []string{"C", "Go"} {
		t.Run(path, func(t *testing.T) {
			for _, row := range dodgeOriginalRows547C50(t) {
				t.Run(row.Input.Name, func(t *testing.T) {
					w := newDodgeWorld547C50(t, row)
					want, writes := w.expected(t, row)
					runtime := w.runtime(t, row)
					unit := w.Unit
					if row.Input.NilUnit {
						unit = nil
					}
					var fault any
					result := 0
					func() {
						defer func() { fault = recover() }()
						if path == "C" {
							old := monsterDodgeRuntimeOverride547C50
							monsterDodgeRuntimeOverride547C50 = &runtime
							defer func() { monsterDodgeRuntimeOverride547C50 = old }()
							result = dodgeOriginalC547C50(unit)
						} else {
							entry, ok := any(unit).(dodgeEntry547C50)
							if !ok {
								t.Fatal("native raw entry not implemented")
							}
							result = entry.MonsterCheckDodgeablesNative547C50(runtime.RandomFloat, runtime.RandomInt, runtime.Trace, runtime.Obstacles, runtime.Tile, runtime.Pop, runtime.Push, runtime.Frame, runtime.FPS)
						}
					}()
					if (fault != nil) != (len(row.Output.Fault) > 0) {
						t.Fatal("raw original fault prefix", fault, row.Output.Fault)
					}
					if fault == nil && (row.Output.Result == nil || uint32(result) != *row.Output.Result) {
						t.Fatal("original return", result, row.Output.Result)
					}
					if !reflect.DeepEqual(dodgeCanonical547C50(t, w.Events), dodgeCanonical547C50(t, row.Output.Events)) {
						t.Fatalf("ordered original services/boundary stores differ\ngot %s\nwant %s", mustDodgeJSON547C50(w.Events), mustDodgeJSON547C50(row.Output.Events))
					}
					if !reflect.DeepEqual(w.Writes, writes) || w.snapshot() != row.Output.State || !reflect.DeepEqual(w.rayBits(), row.Output.Ray) {
						t.Fatal("original service writes/final scalar state/ray", w.Writes, writes, w.snapshot(), row.Output.State, w.rayBits(), row.Output.Ray)
					}
					if !bytes.Equal(w.Heap, want) {
						for i := range want {
							if want[i] != w.Heap[i] {
								t.Fatalf("whole native allocation differs at%d got%02x want%02x", i, w.Heap[i], want[i])
							}
						}
					}
				})
			}
		})
	}
}
func mustDodgeJSON547C50(value any) []byte { raw, _ := json.Marshal(value); return raw }
func TestMonsterDodgeablesNativeCBefore547C50(t *testing.T) {
	if os.Getenv("OPENNOX_DODGEABLES_C_BEFORE") != "1" {
		return
	}
	if os.Getenv("OPENNOX_DODGEABLES_C_CHILD") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestMonsterDodgeablesNativeCBefore547C50$", "-test.v")
		cmd.Env = append(os.Environ(), "OPENNOX_DODGEABLES_C_CHILD=1")
		out, err := cmd.CombinedOutput()
		t.Logf("%s", out)
		if err == nil || !bytes.Contains(out, []byte("DODGE_C_BEFORE_VALID_NATIVE_CONTROLS")) || !bytes.Contains(out, []byte("SIGSEGV")) || !bytes.Contains(out, []byte("_Cfunc_dodge_original_public_547c50")) {
			t.Fatal("actual old C must fault after valid native controls", err)
		}
		return
	}
	TestMonsterDodgeablesNativeLayout547C50(t)
	w := newDodgeWorld547C50(t, dodgeOriginalRows547C50(t)[0])
	w.Data.AIStackInd = 0
	w.Data.AIStack[0].Action = 0
	w.Data.SoundSet122 = unsafe.Pointer(w.Pushes[0])
	before := bytes.Clone(w.Heap)
	set, action := dodgeOriginalControls547C50(w.Unit)
	if set != uintptr(w.Data.SoundSet122) || action != 0 || !bytes.Equal(before, w.Heap) {
		t.Fatal("actual typed getter and complete initial record controls")
	}
	if dodgeOriginalABI547C50() != 32 {
		t.Fatal("unchanged legacy int ABI required")
	}
	t.Logf("DODGE_C_BEFORE_VALID_NATIVE_CONTROLS unit=%p update=%p radius=%08x speed=%08x publicABI=legacy32-explicit-narrowing", w.Unit, w.Data, math.Float32bits(w.Unit.Shape.Circle.R), math.Float32bits(w.Unit.SpeedCur))
	dodgeOriginalC547C50(w.Unit)
	t.Fatal("old C unexpectedly returned")
}
