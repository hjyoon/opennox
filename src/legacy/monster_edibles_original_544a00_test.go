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
	"reflect"
	"runtime/debug"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type edibleRawContract544A00 interface {
	MonsterSearchEdibleNative544A00(float32, *server.MonsterEdibleSearchState544A00, server.MonsterEdibleNativeRuntime544A00) *server.Object
	MonsterSearchEdibleCandidateNative544A40(*server.Object, *server.MonsterEdibleSearchState544A00, server.MonsterEdibleNativeRuntime544A00)
	MonsterRetreatCheckEdiblesNative5455E0(*server.MonsterEdibleSearchState544A00, server.MonsterEdibleNativeRuntime544A00) *server.AIStackItem
}
type edibleObject544A00 struct {
	Class  uint32    `json:"cls"`
	Sub    uint32    `json:"sub"`
	Poison uint8     `json:"poison"`
	Pos    [2]uint32 `json:"pos"`
}
type edibleState544A00 struct {
	Objects []edibleObject544A00 `json:"objects"`
	Flags   uint32               `json:"flags"`
	Best    uint32               `json:"best"`
	Chosen  int                  `json:"chosen"`
	Items   [4][6]uint32         `json:"items"`
}
type edibleMutation544A00 struct {
	Objects []struct {
		ID     int
		Class  *uint32 `json:"cls"`
		Sub    *uint32
		Poison *uint8
		Pos    *[2]uint32
	}
	Flags, Best *uint32
	Chosen      *int
	Items       []struct {
		ID    int
		Words []uint32
	}
}
type edibleNested544A00 struct {
	Unit   int
	Radius uint32
}
type edibleStep544A00 struct {
	Item                  int
	Accept                *uint32
	Mutation, AfterNested edibleMutation544A00
	Reenter               *edibleNested544A00
}

// Underscores are explicit: Go's case folding does not remove them.
func (s *edibleStep544A00) UnmarshalJSON(raw []byte) error {
	type alias edibleStep544A00
	var v struct {
		*alias
		AfterNested edibleMutation544A00 `json:"after_nested"`
	}
	v.alias = (*alias)(s)
	if err := json.Unmarshal(raw, &v); err != nil {
		return err
	}
	s.AfterNested = v.AfterNested
	return nil
}

type edibleEnum544A00 struct {
	Candidates    []int
	Before, After edibleMutation544A00
	AfterEach     []edibleMutation544A00 `json:"after_each"`
}
type edibleInput544A00 struct {
	Name, Entry          string
	Unit, Candidate      int
	Radius, Flags, Best  uint32
	Chosen               int
	Objects              []edibleObject544A00
	Enums                []edibleEnum544A00
	Interactions, Pushes []edibleStep544A00
}
type edibleWrite544A00 struct {
	AfterEvent        int `json:"after_event"`
	PC, Target        string
	Item, Field, Size int
	Value             uint32
}
type edibleRow544A00 struct {
	Input  edibleInput544A00
	Output struct {
		Result *int
		Events []json.RawMessage
		Writes []edibleWrite544A00
		State  edibleState544A00
		Fault  []json.RawMessage
		Guard  bool `json:"whole_heap_scratch_guarded_caller_valid_nonvolatile_and_FPU_preserved"`
	}
}

func TestMonsterSearchEdibleOriginal544A00(t *testing.T) {
	TestMonsterSearchEdibleNativeLayout544A00(t)
	raw, err := os.ReadFile("../server/testdata/monster_edibles_original_544a00.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "12bb2143ec39e6f694d12e5a6fc591e0d618609f5de31b374270580eb9f7c188" {
		t.Fatal("frozen gzip changed")
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
	if fmt.Sprintf("%x", sha256.Sum256(decoded)) != "6524d1ee641e146a990c61be4e539cd6b415fa2b3234ac64be9f30378f4e29d3" {
		t.Fatal("frozen decoded original changed")
	}
	var data struct{ Rows []edibleRow544A00 }
	if err = json.Unmarshal(decoded, &data); err != nil {
		t.Fatal(err)
	}
	if len(data.Rows) != 3406 {
		t.Fatal("context count", len(data.Rows))
	}
	for _, mode := range []string{"C", "Go"} {
		t.Run(mode, func(t *testing.T) {
			for _, row := range data.Rows {
				t.Run(row.Input.Name, func(t *testing.T) { edibleCompareOriginal544A00(t, mode, row) })
			}
		})
	}
}

func edibleCompareOriginal544A00(t *testing.T, mode string, row edibleRow544A00) {
	t.Helper()
	if !row.Output.Guard {
		t.Fatal("original guards not certified")
	}
	buf, free := alloc.Make([]byte{}, 0x10000)
	defer free()
	for i := range buf {
		buf[i] = byte(i*37 + 0xa5)
	}
	layout := edibleOriginalLayout544A00()
	ptrSize := int(unsafe.Sizeof(uintptr(0)))
	objects := [5]*server.Object{}
	items := [4]*server.AIStackItem{}
	for i := range objects {
		objects[i] = (*server.Object)(unsafe.Pointer(&buf[(i+1)*0x1000]))
	}
	for i := range items {
		items[i] = (*server.AIStackItem)(unsafe.Pointer(&buf[(i+8)*0x1000]))
	}
	objectAt := func(i int) *server.Object {
		if i < 0 {
			return nil
		}
		return objects[i]
	}
	objectRole := func(p *server.Object) int {
		if p == nil {
			return -1
		}
		for i, v := range objects {
			if p == v {
				return i
			}
		}
		t.Fatalf("unknown full object pointer %p", p)
		return -2
	}
	itemRole := func(p *server.AIStackItem) int {
		if p == nil {
			return -1
		}
		for i, v := range items {
			if p == v {
				return i
			}
		}
		t.Fatalf("unknown full item pointer %p", p)
		return -2
	}
	decode := func(v uint32) uintptr {
		if v >= 0xf0000000 && v < 0xf0000005 {
			return uintptr(unsafe.Pointer(objects[v-0xf0000000]))
		}
		return uintptr(v)
	}
	canonical := func(v uintptr) uint32 {
		for i, p := range objects {
			if v == uintptr(unsafe.Pointer(p)) {
				return 0xf0000000 + uint32(i)
			}
		}
		return uint32(v)
	}
	itemOffset := func(i, field int) int {
		base := (i + 8) * 0x1000
		if field == 0 {
			return base
		}
		if field == 5 {
			return base + int(layout[10])
		}
		return base + int(layout[9]) + (field-1)*ptrSize
	}
	put := func(dst []byte, offset, n int, value uint64) {
		switch n {
		case 1:
			dst[offset] = byte(value)
		case 4:
			binary.LittleEndian.PutUint32(dst[offset:], uint32(value))
		case 8:
			binary.LittleEndian.PutUint64(dst[offset:], value)
		default:
			t.Fatal("write width", n)
		}
	}
	itemWord := func(i, field int) uint32 {
		if field == 0 {
			return items[i].Action
		}
		if field == 5 {
			return items[i].Field5
		}
		return canonical(items[i].Args[field-1])
	}
	for i, v := range row.Input.Objects {
		base := (i + 1) * 0x1000
		put(buf, base+int(layout[0]), 4, uint64(v.Class))
		put(buf, base+int(layout[1]), 4, uint64(v.Sub))
		put(buf, base+int(layout[3]), 1, uint64(v.Poison))
		put(buf, base+int(layout[2]), 4, uint64(v.Pos[0]))
		put(buf, base+int(layout[2])+4, 4, uint64(v.Pos[1]))
	}
	for i := range items {
		for f := 0; f < 6; f++ {
			v := uint64(0x5a000000 + i*256 + f + 1)
			n := 4
			if f > 0 && f < 5 {
				n = ptrSize
				if ptrSize == 8 {
					v |= 0xface123400000000
				}
			}
			put(buf, itemOffset(i, f), n, v)
		}
	}
	oldFlags := noxflags.GetGame()
	defer func() { noxflags.ResetGame(); noxflags.SetGame(oldFlags) }()
	setFlags := func(v uint32) { noxflags.ResetGame(); noxflags.SetGame(noxflags.GameFlag(v)) }
	setFlags(row.Input.Flags)
	oldScratch, oldRuntime := edibleNativeScratch544A00, edibleNativeRuntimeOverride544A00
	defer func() { edibleNativeScratch544A00, edibleNativeRuntimeOverride544A00 = oldScratch, oldRuntime }()
	edibleNativeScratch544A00 = server.MonsterEdibleSearchState544A00{BestDistance: math.Float32frombits(row.Input.Best), BestObject: objectAt(row.Input.Chosen)}
	want := bytes.Clone(buf)
	wantFlags, wantBest, wantChosen := row.Input.Flags, row.Input.Best, objectAt(row.Input.Chosen)
	checkBytes := func() {
		t.Helper()
		if !bytes.Equal(buf, want) {
			for i := range buf {
				if buf[i] != want[i] {
					t.Fatalf("whole64KiB byte offset=%#x got=%02x want=%02x", i, buf[i], want[i])
				}
			}
		}
		if uint32(noxflags.GetGame()) != wantFlags || math.Float32bits(edibleNativeScratch544A00.BestDistance) != wantBest || edibleNativeScratch544A00.BestObject != wantChosen {
			t.Fatalf("whole live scratch flags=%x best=%x chosen=%d want=%x/%x/%d", uint32(noxflags.GetGame()), math.Float32bits(edibleNativeScratch544A00.BestDistance), objectRole(edibleNativeScratch544A00.BestObject), wantFlags, wantBest, objectRole(wantChosen))
		}
	}
	state := func() edibleState544A00 {
		v := edibleState544A00{Flags: uint32(noxflags.GetGame()), Best: math.Float32bits(edibleNativeScratch544A00.BestDistance), Chosen: objectRole(edibleNativeScratch544A00.BestObject)}
		for _, p := range objects {
			v.Objects = append(v.Objects, edibleObject544A00{Class: uint32(p.ObjClass), Sub: uint32(p.ObjSubClass), Poison: p.Poison540, Pos: [2]uint32{math.Float32bits(p.PosVec.X), math.Float32bits(p.PosVec.Y)}})
		}
		for i := range items {
			for f := 0; f < 6; f++ {
				v.Items[i][f] = itemWord(i, f)
			}
		}
		return v
	}
	eventCursor, writeCursor, enumIndex, interactIndex, pushIndex := 0, 0, 0, 0, 0
	advanceWrites := func() {
		for writeCursor < len(row.Output.Writes) && row.Output.Writes[writeCursor].AfterEvent <= eventCursor {
			w := row.Output.Writes[writeCursor]
			writeCursor++
			switch w.Target {
			case "chosen":
				wantChosen = nil
				if w.Value >= 0xf0000000 && w.Value < 0xf0000005 {
					wantChosen = objects[w.Value-0xf0000000]
				} else if w.Value != 0 {
					t.Fatal("unknown original chosen write", w)
				}
			case "best":
				wantBest = w.Value
			case "item":
				n := 4
				v := uint64(w.Value)
				if w.Field > 0 && w.Field < 5 {
					n = ptrSize
				}
				if w.Size == 1 {
					n = 1
				}
				if w.PC == "0x54563d" || w.PC == "0x54564f" || w.PC == "0x54566a" {
					v = uint64(decode(w.Value))
				}
				put(want, itemOffset(w.Item, w.Field), n, v)
			default:
				t.Fatal("unknown original target", w)
			}
		}
	}
	normalized := func(raw []byte) any {
		var v any
		if err := json.Unmarshal(raw, &v); err != nil {
			t.Fatal(err)
		}
		return v
	}
	emit := func(v map[string]any) {
		t.Helper()
		advanceWrites()
		checkBytes()
		if eventCursor >= len(row.Output.Events) {
			t.Fatal("extra native world event", v)
		}
		raw, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(normalized(raw), normalized(row.Output.Events[eventCursor])) {
			t.Fatalf("event%d got=%s want=%s", eventCursor, raw, row.Output.Events[eventCursor])
		}
		eventCursor++
	}
	mutate := func(m edibleMutation544A00) {
		for _, c := range m.Objects {
			base := (c.ID + 1) * 0x1000
			for _, dst := range [][]byte{buf, want} {
				if c.Class != nil {
					put(dst, base+int(layout[0]), 4, uint64(*c.Class))
				}
				if c.Sub != nil {
					put(dst, base+int(layout[1]), 4, uint64(*c.Sub))
				}
				if c.Poison != nil {
					put(dst, base+int(layout[3]), 1, uint64(*c.Poison))
				}
				if c.Pos != nil {
					put(dst, base+int(layout[2]), 4, uint64(c.Pos[0]))
					put(dst, base+int(layout[2])+4, 4, uint64(c.Pos[1]))
				}
			}
		}
		if m.Flags != nil {
			setFlags(*m.Flags)
			wantFlags = *m.Flags
		}
		if m.Best != nil {
			edibleNativeScratch544A00.BestDistance = math.Float32frombits(*m.Best)
			wantBest = *m.Best
		}
		if m.Chosen != nil {
			edibleNativeScratch544A00.BestObject = objectAt(*m.Chosen)
			wantChosen = objectAt(*m.Chosen)
		}
		for _, c := range m.Items {
			for f, v := range c.Words {
				n := 4
				value := uint64(v)
				if f > 0 && f < 5 {
					n = ptrSize
					value = uint64(decode(v))
				}
				for _, dst := range [][]byte{buf, want} {
					put(dst, itemOffset(c.ID, f), n, value)
				}
			}
		}
	}
	var hooks server.MonsterEdibleNativeRuntime544A00
	var search func(*server.Object, float32) *server.Object
	search = func(unit *server.Object, radius float32) *server.Object {
		if mode == "C" {
			return (*server.Object)(edibleOriginalC544A00(1, unit, nil, radius))
		}
		native, ok := any(unit).(edibleRawContract544A00)
		if !ok {
			t.Fatal("raw native contract not implemented")
		}
		return native.MonsterSearchEdibleNative544A00(radius, &edibleNativeScratch544A00, hooks)
	}
	nested := func(kind string, index int, step edibleStep544A00) {
		if step.Reenter == nil {
			return
		}
		v := step.Reenter
		emit(map[string]any{"kind": "nested-search", "service": kind, "index": index, "unit": v.Unit, "radius": v.Radius, "before": state()})
		result := search(objectAt(v.Unit), math.Float32frombits(v.Radius))
		emit(map[string]any{"kind": "nested-return", "service": kind, "index": index, "result": objectRole(result), "after": state()})
	}
	flags := func(mask uint32) bool {
		emit(map[string]any{"kind": "flags", "mask": mask, "before": state()})
		return edibleOriginalNativeFlags544A00(mask)
	}
	hooks.Quest = func() bool { return flags(4096) }
	hooks.Online = func() bool { return flags(8192) }
	hooks.EachCircle = func(point *types.Pointf, radius float32, callback func(*server.Object, *server.Object), unit *server.Object) {
		// Force the same declared spatial point-read admission before emitting enum.
		position := [2]uint32{math.Float32bits(point.X), math.Float32bits(point.Y)}
		index := enumIndex
		enumIndex++
		plan := edibleEnum544A00{}
		if index < len(row.Input.Enums) {
			plan = row.Input.Enums[index]
		}
		emit(map[string]any{"kind": "enum", "index": index, "unit": objectRole(unit), "radius": math.Float32bits(radius), "position": position, "before": state()})
		mutate(plan.Before)
		for i, c := range plan.Candidates {
			callback(objectAt(c), unit)
			emit(map[string]any{"kind": "callback-return", "enum": index, "index": i, "candidate": c, "after": state()})
			if i < len(plan.AfterEach) {
				mutate(plan.AfterEach[i])
			}
		}
		mutate(plan.After)
		emit(map[string]any{"kind": "enum-return", "index": index, "after": state()})
	}
	hooks.CanInteract = func(unit, candidate *server.Object, options int) bool {
		index := interactIndex
		interactIndex++
		step := edibleStep544A00{}
		if index < len(row.Input.Interactions) {
			step = row.Input.Interactions[index]
		}
		emit(map[string]any{"kind": "interact", "index": index, "unit": objectRole(unit), "candidate": objectRole(candidate), "options": options, "before": state()})
		mutate(step.Mutation)
		nested("interact", index, step)
		mutate(step.AfterNested)
		value := uint32(1)
		if step.Accept != nil {
			value = *step.Accept
		}
		emit(map[string]any{"kind": "interact-return", "index": index, "value": value, "after": state()})
		return value != 0
	}
	hooks.Push = func(unit *server.Object, action ai.ActionType) *server.AIStackItem {
		index := pushIndex
		pushIndex++
		step := edibleStep544A00{Item: index % 4}
		if index < len(row.Input.Pushes) {
			step = row.Input.Pushes[index]
		}
		emit(map[string]any{"kind": "push", "index": index, "unit": objectRole(unit), "action": uint32(action), "before": state()})
		mutate(step.Mutation)
		nested("push", index, step)
		mutate(step.AfterNested)
		var result *server.AIStackItem
		if step.Item >= 0 {
			result = items[step.Item]
			result.Action = uint32(action)
			put(want, itemOffset(step.Item, 0), 4, uint64(action))
		}
		emit(map[string]any{"kind": "push-return", "index": index, "item": step.Item, "after": state()})
		return result
	}
	edibleNativeRuntimeOverride544A00 = &hooks
	var result unsafe.Pointer
	var fault any
	oldFault := debug.SetPanicOnFault(true)
	defer debug.SetPanicOnFault(oldFault)
	func() {
		defer func() { fault = recover() }()
		unit, candidate := objectAt(row.Input.Unit), objectAt(row.Input.Candidate)
		entry := 0
		if row.Input.Entry == "search" {
			entry = 1
		} else if row.Input.Entry == "callback" {
			entry = 2
		}
		if mode == "C" {
			result = edibleOriginalC544A00(entry, unit, candidate, math.Float32frombits(row.Input.Radius))
			return
		}
		native, ok := any(unit).(edibleRawContract544A00)
		if !ok {
			t.Fatal("raw native contract not implemented")
		}
		switch entry {
		case 0:
			result = unsafe.Pointer(native.MonsterRetreatCheckEdiblesNative5455E0(&edibleNativeScratch544A00, hooks))
		case 1:
			result = unsafe.Pointer(native.MonsterSearchEdibleNative544A00(math.Float32frombits(row.Input.Radius), &edibleNativeScratch544A00, hooks))
		case 2:
			c, ok := any(candidate).(edibleRawContract544A00)
			if !ok {
				t.Fatal("raw candidate contract not implemented")
			}
			c.MonsterSearchEdibleCandidateNative544A40(unit, &edibleNativeScratch544A00, hooks)
		}
	}()
	if (fault != nil) != (len(row.Output.Fault) != 0) {
		t.Fatal("original fault prefix", fault, row.Output.Fault)
	}
	advanceWrites()
	checkBytes()
	if eventCursor != len(row.Output.Events) || writeCursor != len(row.Output.Writes) {
		t.Fatal("incomplete original event/write history", eventCursor, len(row.Output.Events), writeCursor, len(row.Output.Writes))
	}
	if !reflect.DeepEqual(state(), row.Output.State) {
		t.Fatalf("complete final state got=%+v want=%+v", state(), row.Output.State)
	}
	if row.Output.Result != nil {
		got := -1
		if row.Input.Entry == "root" {
			got = itemRole((*server.AIStackItem)(result))
		} else {
			got = objectRole((*server.Object)(result))
		}
		if got != *row.Output.Result {
			t.Fatal("full native return pointer", got, *row.Output.Result)
		}
	}
}
