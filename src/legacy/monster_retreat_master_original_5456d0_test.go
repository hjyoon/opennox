package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type retreatMasterMutation5456D0 struct {
	Owner, Update          *int
	NilHealth              *bool `json:"nil_health"`
	Health                 *[2]uint16
	Positions              *[4][2]uint32
	Ranges, Resume, Status *[2]uint32
	Buffs                  *uint32
}
type retreatMasterInput5456D0 struct {
	Name                   string
	Owner, Update          int
	Positions              [4][2]uint32
	Ranges, Resume, Status [2]uint32
	Buffs                  uint32
	Cur, Maximum           uint16
	UnitClass              uint32 `json:"unit_class"`
	NilUnit                bool   `json:"nil_unit"`
	NilHealth              bool   `json:"nil_health"`
	Steps                  [2]struct {
		ReturnItem int `json:"return_item"`
		Mutation   retreatMasterMutation5456D0
	}
	PopMutation retreatMasterMutation5456D0 `json:"pop_mutation"`
}
type retreatMasterState5456D0 struct {
	Owner     int          `json:"owner"`
	Update    int          `json:"update"`
	NilHealth bool         `json:"nil_health"`
	Positions [4][2]uint32 `json:"positions"`
	Ranges    [2]uint32    `json:"ranges"`
	Resume    [2]uint32    `json:"resume"`
	Status    [2]uint32    `json:"status"`
	Buffs     uint32       `json:"buffs"`
	Health    [2]uint16    `json:"health"`
	Items     [2][6]uint32 `json:"items"`
}
type retreatMasterWrite5456D0 struct {
	Item, Field, Size int
	Value             uint32
	PC                string
}
type retreatMasterRow5456D0 struct {
	Input  retreatMasterInput5456D0
	Output struct {
		Events []json.RawMessage
		Writes []retreatMasterWrite5456D0
		State  retreatMasterState5456D0
		Fault  []json.RawMessage
	}
}

func retreatMasterRows5456D0(t *testing.T) []retreatMasterRow5456D0 {
	t.Helper()
	raw, err := os.ReadFile("../server/testdata/monster_retreat_master_original_5456d0.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "f7e863caa08ddcc57f5bafc7f345087fcb5f85b09bc5a2839bbc5e2e93559086" {
		t.Fatal("frozen original gzip changed")
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
	if fmt.Sprintf("%x", sha256.Sum256(decoded)) != "61bf4dd194efdf884b7dbc62861cd8b8d58590f35c7b48d20d13c2ab72cb806f" {
		t.Fatal("frozen original decoded bytes changed")
	}
	var data struct{ Rows []retreatMasterRow5456D0 }
	if err = json.Unmarshal(decoded, &data); err != nil {
		t.Fatal(err)
	}
	if len(data.Rows) != 942 {
		t.Fatal("original context count", len(data.Rows))
	}
	return data.Rows
}

type retreatMasterWorld5456D0 struct {
	Objects [4]*server.Object
	Updates [2]*server.MonsterUpdateData
	Health  *server.HealthData
	Items   [2]*server.AIStackItem
	Heap    []byte
	Events  []any
	Pushes  int
}

func newRetreatMasterWorld5456D0(t *testing.T, input retreatMasterInput5456D0) *retreatMasterWorld5456D0 {
	t.Helper()
	ptr, free := alloc.MallocHigh(65536)
	t.Cleanup(free)
	w := &retreatMasterWorld5456D0{Heap: unsafe.Slice((*byte)(ptr), 65536), Events: make([]any, 0)}
	for i := range w.Heap {
		w.Heap[i] = byte(i*37 + 0xa5)
	}
	for i, offset := range []uintptr{4096, 6144, 8192, 10240} {
		w.Objects[i] = (*server.Object)(unsafe.Add(ptr, offset))
	}
	w.Updates = [2]*server.MonsterUpdateData{(*server.MonsterUpdateData)(unsafe.Add(ptr, 16384)), (*server.MonsterUpdateData)(unsafe.Add(ptr, 32768))}
	w.Health = (*server.HealthData)(unsafe.Add(ptr, 49152))
	w.Items = [2]*server.AIStackItem{(*server.AIStackItem)(unsafe.Add(ptr, 53248)), (*server.AIStackItem)(unsafe.Add(ptr, 55296))}
	unit := w.Objects[0]
	unit.ObjClass, unit.ObjOwner, unit.UpdateData = object.Class(input.UnitClass), w.obj(input.Owner), w.update(input.Update)
	unit.HealthData, unit.Buffs = w.Health, input.Buffs
	if input.NilHealth {
		unit.HealthData = nil
	}
	w.Health.Cur, w.Health.Max = input.Cur, input.Maximum
	for i, point := range input.Positions {
		w.Objects[i].PosVec = types.Ptf(math.Float32frombits(point[0]), math.Float32frombits(point[1]))
	}
	for i, update := range w.Updates {
		update.Field329, update.ResumeLevel = math.Float32frombits(input.Ranges[i]), math.Float32frombits(input.Resume[i])
		update.StatusFlags = object.MonsterStatus(input.Status[i])
	}
	for i, item := range w.Items {
		item.Action, item.Field5 = uint32(0x5a000001+i*256), uint32(0x5a000006+i*256)
		for field := range item.Args {
			item.Args[field] = uintptr(0x5a000002 + i*256 + field)
			if unsafe.Sizeof(uintptr(0)) == 8 {
				high := uint64(0xface1234)
				item.Args[field] |= uintptr(high << 32)
			}
		}
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, value := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(w.Objects[1]), unsafe.Pointer(w.Updates[0]), unsafe.Pointer(w.Updates[1]), unsafe.Pointer(w.Health), unsafe.Pointer(w.Items[0]), unsafe.Pointer(w.Items[1])} {
			if uintptr(value) <= math.MaxUint32 {
				t.Fatal("native C-owned object/data/health/action pointer must exceed4GiB")
			}
		}
	}
	return w
}
func (w *retreatMasterWorld5456D0) obj(index int) *server.Object {
	if index < 0 {
		return nil
	}
	return w.Objects[index]
}
func (w *retreatMasterWorld5456D0) update(index int) unsafe.Pointer {
	if index < 0 {
		return nil
	}
	return unsafe.Pointer(w.Updates[index])
}
func (w *retreatMasterWorld5456D0) item(index int) *server.AIStackItem {
	if index < 0 {
		return nil
	}
	return w.Items[index]
}
func (w *retreatMasterWorld5456D0) ref(unit *server.Object) int {
	if unit == nil {
		return -1
	}
	for i, value := range w.Objects {
		if value == unit {
			return i
		}
	}
	panic("native owner outside frozen roles")
}
func (w *retreatMasterWorld5456D0) canonical(value uintptr) uint32 {
	for i, obj := range w.Objects {
		if value == uintptr(unsafe.Pointer(obj)) {
			return 0xf0000000 + uint32(i)
		}
	}
	return uint32(value)
}
func (w *retreatMasterWorld5456D0) snapshot() retreatMasterState5456D0 {
	unit := w.Objects[0]
	out := retreatMasterState5456D0{Owner: w.ref(unit.ObjOwner), Update: -1, NilHealth: unit.HealthData == nil, Buffs: unit.Buffs, Health: [2]uint16{w.Health.Cur, w.Health.Max}}
	if unit.UpdateData != nil {
		if unit.UpdateData == unsafe.Pointer(w.Updates[0]) {
			out.Update = 0
		} else if unit.UpdateData == unsafe.Pointer(w.Updates[1]) {
			out.Update = 1
		} else {
			panic("native MUD outside frozen roles")
		}
	}
	for i, obj := range w.Objects {
		out.Positions[i] = [2]uint32{math.Float32bits(obj.PosVec.X), math.Float32bits(obj.PosVec.Y)}
	}
	for i, update := range w.Updates {
		out.Ranges[i], out.Resume[i], out.Status[i] = math.Float32bits(update.Field329), math.Float32bits(update.ResumeLevel), uint32(update.StatusFlags)
	}
	for i, item := range w.Items {
		out.Items[i][0], out.Items[i][5] = item.Action, item.Field5
		for arg, value := range item.Args {
			out.Items[i][arg+1] = uint32(value)
			if arg == 2 {
				out.Items[i][arg+1] = w.canonical(value)
			}
		}
	}
	return out
}
func (w *retreatMasterWorld5456D0) mutate(m retreatMasterMutation5456D0) {
	unit := w.Objects[0]
	if m.Owner != nil {
		unit.ObjOwner = w.obj(*m.Owner)
	}
	if m.Update != nil {
		unit.UpdateData = w.update(*m.Update)
	}
	if m.NilHealth != nil {
		unit.HealthData = w.Health
		if *m.NilHealth {
			unit.HealthData = nil
		}
	}
	if m.Health != nil {
		w.Health.Cur, w.Health.Max = m.Health[0], m.Health[1]
	}
	if m.Positions != nil {
		for i, point := range m.Positions {
			w.Objects[i].PosVec = types.Ptf(math.Float32frombits(point[0]), math.Float32frombits(point[1]))
		}
	}
	for i, update := range w.Updates {
		if m.Ranges != nil {
			update.Field329 = math.Float32frombits(m.Ranges[i])
		}
		if m.Resume != nil {
			update.ResumeLevel = math.Float32frombits(m.Resume[i])
		}
		if m.Status != nil {
			update.StatusFlags = object.MonsterStatus(m.Status[i])
		}
	}
	if m.Buffs != nil {
		unit.Buffs = *m.Buffs
	}
}
func (w *retreatMasterWorld5456D0) hooks(t *testing.T, input retreatMasterInput5456D0) retreatMasterOriginalRuntime5456D0 {
	return retreatMasterOriginalRuntime5456D0{
		Push: func(unit *server.Object, action ai.ActionType) *server.AIStackItem {
			if unit != w.Objects[0] || w.Pushes >= 2 || action != []ai.ActionType{ai.DEPENDENCY_OBJECT_FARTHER_THAN, ai.ACTION_MOVE_TO}[w.Pushes] {
				t.Fatal("original push pointer/order", unit, action, w.Pushes)
			}
			index := w.Pushes
			w.Events = append(w.Events, map[string]any{"kind": "push", "index": index, "action": uint32(action), "before": w.snapshot()})
			step := input.Steps[index]
			w.mutate(step.Mutation)
			item := w.item(step.ReturnItem)
			if item != nil {
				item.Action = uint32(action)
			}
			w.Events = append(w.Events, map[string]any{"kind": "push-return", "index": index, "returned_item": step.ReturnItem, "after": w.snapshot()})
			w.Pushes++
			return item
		},
		Pop: func(unit *server.Object) int {
			if unit != w.Objects[0] {
				t.Fatal("original pop pointer")
			}
			w.Events = append(w.Events, map[string]any{"kind": "pop", "before": w.snapshot()})
			w.mutate(input.PopMutation)
			w.Events = append(w.Events, map[string]any{"kind": "pop-return", "after": w.snapshot()})
			return -2023406815
		},
	}
}

type retreatMasterNative5456D0 interface {
	MonsterRetreatToMasterNative5456D0(func(ai.ActionType) *server.AIStackItem, func() int)
}

func TestMonsterRetreatMasterOriginal5456D0(t *testing.T) {
	for _, path := range []string{"C", "Go"} {
		t.Run(path, func(t *testing.T) {
			for _, row := range retreatMasterRows5456D0(t) {
				t.Run(row.Input.Name, func(t *testing.T) {
					if path == "C" && retreatMasterOriginalABI5456D0() != 64 {
						t.Fatal("native public entry pending; actual legacy fault qualified separately")
					}
					w := newRetreatMasterWorld5456D0(t, row.Input)
					hooks := w.hooks(t, row.Input)
					old := retreatMasterRuntimeOverride5456D0
					retreatMasterRuntimeOverride5456D0 = &hooks
					defer func() { retreatMasterRuntimeOverride5456D0 = old }()
					unit := w.Objects[0]
					if row.Input.NilUnit {
						unit = nil
					}
					before := bytes.Clone(w.Heap)
					var fault any
					func() {
						defer func() { fault = recover() }()
						if path == "C" {
							retreatMasterOriginalC5456D0(unit)
							return
						}
						native, ok := any(unit).(retreatMasterNative5456D0)
						if !ok {
							t.Fatal("raw direct-Go entry pending")
						}
						native.MonsterRetreatToMasterNative5456D0(func(action ai.ActionType) *server.AIStackItem { return hooks.Push(unit, action) }, func() int { return hooks.Pop(unit) })
					}()
					if (fault != nil) != (len(row.Output.Fault) != 0) {
						t.Fatal("original fault boundary", fault, row.Output.Fault)
					}
					actual, err := json.Marshal(w.Events)
					if err != nil {
						t.Fatal(err)
					}
					wanted, err := json.Marshal(row.Output.Events)
					if err != nil {
						t.Fatal(err)
					}
					var a, b any
					if json.Unmarshal(actual, &a) != nil || json.Unmarshal(wanted, &b) != nil || !reflect.DeepEqual(a, b) {
						t.Fatalf("original ordered world-service states\ngot%s\nwant%s", actual, wanted)
					}
					if got := w.snapshot(); got != row.Output.State {
						t.Fatalf("original final state got=%+v want=%+v", got, row.Output.State)
					}
					after := bytes.Clone(w.Heap)
					copy(w.Heap, before)
					state := row.Output.State
					w.mutate(retreatMasterMutation5456D0{Owner: &state.Owner, Update: &state.Update, NilHealth: &state.NilHealth, Health: &state.Health, Positions: &state.Positions, Ranges: &state.Ranges, Resume: &state.Resume, Status: &state.Status, Buffs: &state.Buffs})
					for i, item := range w.Items {
						item.Action = state.Items[i][0]
					}
					for _, write := range row.Output.Writes {
						if write.Size != 4 || write.Field < 1 || write.Field > 3 {
							t.Fatal("original payload store descriptor", write)
						}
						value := uintptr(write.Value)
						if write.Field == 3 && write.Value >= 0xf0000000 && write.Value < 0xf0000004 {
							value = uintptr(unsafe.Pointer(w.Objects[write.Value-0xf0000000]))
						}
						w.Items[write.Item].Args[write.Field-1] = value
					}
					if !bytes.Equal(after, w.Heap) {
						t.Fatal("whole native objects/health/cached and live MUD/items/unwritten high uintptr words/canaries differ")
					}
				})
			}
		})
	}
}
