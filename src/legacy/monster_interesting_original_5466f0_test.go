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
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type interestingInput5466F0 struct {
	Name                   string
	UnitNil                bool `json:"unit_nil"`
	Update                 int
	Cls, Subclass          uint32
	Marker, Stamp          [2]uint32
	XY                     [2][2]uint32
	Position               [2]uint32
	Frame, FPS, Seed, Mask uint32
	Tile, Path, Trace      int
	Returns                [5]int
	Mutations              []struct {
		At, Kind string
		Value    int
	}
}
type interestingState5466F0 struct {
	Update        int
	Cls, Subclass uint32
	Position      [2]uint32
	Data          [2]struct {
		Marker, Stamp uint32
		XY            [2]uint32
	}
	Items           [4][5]uint32
	Frame, FPS, RNG uint32
}
type interestingRow5466F0 struct {
	Name          string
	Input         interestingInput5466F0
	Before, After interestingState5466F0
	Events        []map[string]any
	Writes        []struct {
		Role  [2]string
		Value uint32
	}
	Fault  any
	Result uint32
	Whole  bool `json:"whole_original_heap_changes_within_declared_writes"`
}

func interestingLoad5466F0(t *testing.T) []interestingRow5466F0 {
	t.Helper()
	b, err := os.ReadFile("../server/testdata/monster_interesting_original_5466f0.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(b)) != "7dc57fa677f9de65736d89bf2f13dd26b11cc86b051e3650e77a18208065199e" {
		t.Fatal("sealed original gzip")
	}
	z, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err = z.Close(); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "b8cda36a229995b93f0c44f8d8c381226eeb7b30a2942b6bf87b4a0a0d82b01a" {
		t.Fatal("sealed decoded original")
	}
	var data struct {
		OriginalSHA string `json:"original_sha256"`
		Rows        []interestingRow5466F0
	}
	if err = json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	if data.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(data.Rows) != 5511 {
		t.Fatal("original executable/dimensions")
	}
	return data.Rows
}

type interestingWorld5466F0 struct {
	Unit            *server.Object
	Data            [2]*server.MonsterUpdateData
	Items           [4]*server.AIStackItem
	Base            unsafe.Pointer
	Raw, Before     []byte
	Frame, FPS, RNG uint32
	Pushes          int
	Events          []map[string]any
	Runtime         monsterInterestingSoundOriginalRuntime5466F0
}

func interestingWorldNew5466F0(t *testing.T, row interestingRow5466F0) *interestingWorld5466F0 {
	t.Helper()
	const size = 32768
	base, free := alloc.MallocHigh(size)
	t.Cleanup(free)
	w := &interestingWorld5466F0{Base: base, Raw: unsafe.Slice((*byte)(base), size), Frame: row.Input.Frame, FPS: row.Input.FPS, RNG: row.Input.Seed, Events: make([]map[string]any, 0)}
	for i := range w.Raw {
		w.Raw[i] = 0xa5
	}
	next := 0
	take := func(n uintptr) unsafe.Pointer {
		next = (next + 63) &^ 63
		next += 32
		start := next
		next += int(n) + 32
		if next > size {
			t.Fatal("guarded native allocation")
		}
		p := unsafe.Add(base, start)
		if unsafe.Sizeof(p) == 8 && uintptr(p) <= math.MaxUint32 {
			t.Fatal("native address must exceed4GiB")
		}
		return p
	}
	w.Unit = (*server.Object)(take(unsafe.Sizeof(server.Object{})))
	for i := range w.Data {
		w.Data[i] = (*server.MonsterUpdateData)(take(unsafe.Sizeof(server.MonsterUpdateData{})))
	}
	for i := 0; i < 2; i++ {
		w.Items[i] = (*server.AIStackItem)(take(unsafe.Sizeof(server.AIStackItem{})))
	}
	w.Items[2] = &w.Data[0].AIStack[0]
	w.Items[3] = &w.Data[1].AIStack[0]
	w.Unit.ObjClass = object.Class(row.Input.Cls)
	w.Unit.ObjSubClass = object.SubClass(row.Input.Subclass)
	w.Unit.UpdateData = unsafe.Pointer(w.data(row.Input.Update))
	w.Unit.PosVec = interestingPoint5466F0(row.Input.Position)
	for i, d := range w.Data {
		d.Field97 = row.Input.Marker[i]
		d.Field101 = row.Input.Stamp[i]
		d.Field99X = math.Float32frombits(row.Input.XY[i][0])
		d.Field99Y = math.Float32frombits(row.Input.XY[i][1])
	}
	for _, item := range w.Items {
		item.Action = 0xa5a5a5a5
		for i := range item.Args {
			high := uintptr(0x12345678)
			high <<= 32
			item.Args[i] = high | 0xa5a5a5a5
		}
	}
	w.Before = bytes.Clone(w.Raw)
	appendEvent := func(e map[string]any) { w.Events = append(w.Events, e) }
	w.Runtime.Reactive = server.MonsterReactiveRuntime545DA0{
		Frame:    func() uint32 { appendEvent(map[string]any{"kind": "frame", "value": float64(w.Frame)}); return w.Frame },
		TickRate: func() uint32 { appendEvent(map[string]any{"kind": "fps", "value": float64(w.FPS)}); return w.FPS },
		Random: func(min, max int) int {
			i := len(w.Events)
			if i >= len(row.Events) || row.Events[i]["kind"] != "random" {
				t.Fatal("unexpected RNG call")
			}
			expected := row.Events[i]
			if min != int(int32(uint32(expected["min"].(float64)))) || max != int(int32(uint32(expected["max"].(float64)))) || w.RNG != uint32(expected["before"].(float64)) {
				t.Fatal("signed DWORD RNG arguments/index", min, max, w.RNG, expected)
			}
			r := prand.New(int(w.RNG))
			result := r.IntClamp(min, max)
			w.RNG = uint32(r.Index())
			appendEvent(map[string]any{"kind": "random", "min": float64(uint32(min)), "max": float64(uint32(max)), "before": expected["before"], "after": float64(w.RNG), "result": float64(uint32(result))})
			return int(int32(result))
		},
		Push: func(unit *server.Object, action ai.ActionType) *server.AIStackItem {
			if unit != w.Unit || w.Pushes >= 5 {
				t.Fatal("Push native object/count")
			}
			i := w.Pushes
			w.Pushes++
			role := -1
			if row.Input.Mask&(1<<i) != 0 {
				role = row.Input.Returns[i]
			}
			appendEvent(map[string]any{"kind": "push", "action": float64(action), "result": float64(role)})
			var item *server.AIStackItem
			if role >= 0 {
				item = w.Items[role]
				item.Action = uint32(action)
			}
			w.mutate(t, row, "push"+strconv.Itoa(i))
			return item
		},
	}
	pointRole := func(point *types.Pointf) int {
		for i, d := range w.Data {
			if unsafe.Pointer(point) == unsafe.Pointer(&d.Field99X) {
				return i
			}
		}
		t.Fatal("world service lost live native coordinate pointer")
		return -1
	}
	w.Runtime.Tile = func(point *types.Pointf) int {
		i := pointRole(point)
		appendEvent(map[string]any{"kind": "tile", "data": float64(i), "xy": interestingWordsJSON5466F0(*point), "result": float64(row.Input.Tile)})
		w.mutate(t, row, "tile")
		return row.Input.Tile
	}
	w.Runtime.Path = func(unit *server.Object, point *types.Pointf) bool {
		if unit != w.Unit {
			t.Fatal("path native object")
		}
		i := pointRole(point)
		appendEvent(map[string]any{"kind": "path", "data": float64(i), "xy": interestingWordsJSON5466F0(*point), "result": float64(row.Input.Path)})
		w.mutate(t, row, "path")
		return row.Input.Path != 0
	}
	w.Runtime.Trace = func(from, to types.Pointf) bool {
		xy := append(interestingWordsJSON5466F0(from), interestingWordsJSON5466F0(to)...)
		appendEvent(map[string]any{"kind": "trace", "xy": xy, "result": float64(row.Input.Trace)})
		w.mutate(t, row, "trace")
		return row.Input.Trace != 0
	}
	if got := w.state(); !reflect.DeepEqual(got, row.Before) {
		t.Fatal("independent initial native state", got, row.Before)
	}
	return w
}
func interestingPoint5466F0(words [2]uint32) types.Pointf {
	return types.Pointf{X: math.Float32frombits(words[0]), Y: math.Float32frombits(words[1])}
}
func interestingWordsJSON5466F0(point types.Pointf) []any {
	return []any{float64(math.Float32bits(point.X)), float64(math.Float32bits(point.Y))}
}
func (w *interestingWorld5466F0) data(role int) *server.MonsterUpdateData {
	if role < 0 {
		return nil
	}
	return w.Data[role]
}
func (w *interestingWorld5466F0) mutate(t *testing.T, row interestingRow5466F0, stage string) {
	t.Helper()
	for _, m := range row.Input.Mutations {
		if m.At != stage {
			continue
		}
		switch m.Kind {
		case "update":
			w.Unit.UpdateData = unsafe.Pointer(w.data(m.Value))
		case "sound0", "sound1":
			d := w.Data[int(m.Kind[len(m.Kind)-1]-'0')]
			if m.Value != 0 {
				d.Field97 = 0
				d.Field99X = math.Float32frombits(0x7f812345)
				d.Field99Y = math.Float32frombits(0xff800000)
				d.Field101 = 0xfffffffe
			} else {
				d.Field97 = 0x87654321
				d.Field99X = math.Float32frombits(0x80000000)
				d.Field99Y = math.Float32frombits(0x3f800001)
				d.Field101 = 17
			}
		case "position":
			if m.Value != 0 {
				w.Unit.PosVec = interestingPoint5466F0([2]uint32{0x7f812345, 0xff800000})
			} else {
				w.Unit.PosVec = interestingPoint5466F0([2]uint32{0x80000000, 1})
			}
		case "frame":
			if m.Value != 0 {
				w.Frame = 17
			} else {
				w.Frame = 0xfffffffe
			}
		case "fps":
			if m.Value != 0 {
				w.FPS = 0x80000001
			} else {
				w.FPS = 0xffffffff
			}
		case "subclass":
			if m.Value != 0 {
				w.Unit.ObjSubClass = 0x400
			} else {
				w.Unit.ObjSubClass = 0
			}
		default:
			t.Fatal("unknown frozen mutation", m)
		}
	}
}
func (w *interestingWorld5466F0) state() interestingState5466F0 {
	s := interestingState5466F0{Update: -1, Cls: uint32(w.Unit.ObjClass), Subclass: uint32(w.Unit.ObjSubClass), Position: [2]uint32{math.Float32bits(w.Unit.PosVec.X), math.Float32bits(w.Unit.PosVec.Y)}, Frame: w.Frame, FPS: w.FPS, RNG: w.RNG}
	for i, d := range w.Data {
		if w.Unit.UpdateData == unsafe.Pointer(d) {
			s.Update = i
		}
		s.Data[i].Marker = d.Field97
		s.Data[i].Stamp = d.Field101
		s.Data[i].XY = [2]uint32{math.Float32bits(d.Field99X), math.Float32bits(d.Field99Y)}
	}
	for i, item := range w.Items {
		s.Items[i][0] = item.Action
		for j := range item.Args {
			s.Items[i][j+1] = uint32(item.Args[j])
		}
	}
	return s
}
func (w *interestingWorld5466F0) call(row interestingRow5466F0) uint32 {
	old := monsterInterestingSoundRuntimeOverride5466F0
	monsterInterestingSoundRuntimeOverride5466F0 = &w.Runtime
	defer func() { monsterInterestingSoundRuntimeOverride5466F0 = old }()
	unit := w.Unit
	if row.Input.UnitNil {
		unit = nil
	}
	return uint32(monsterInterestingOriginalC5466F0(unit))
}
func (w *interestingWorld5466F0) check(t *testing.T, row interestingRow5466F0) {
	t.Helper()
	if !row.Whole {
		t.Fatal("original whole-heap coverage")
	}
	expected := bytes.Clone(w.Before)
	write := func(p unsafe.Pointer, value uintptr, size uintptr) {
		off := int(uintptr(p) - uintptr(w.Base))
		if off < 0 || off+int(size) > len(expected) {
			t.Fatal("projected original write outside heap")
		}
		if size == 8 {
			binary.LittleEndian.PutUint64(expected[off:], uint64(value))
		} else {
			binary.LittleEndian.PutUint32(expected[off:], uint32(value))
		}
	}
	for _, change := range row.Writes {
		role, field, v := change.Role[0], change.Role[1], change.Value
		switch {
		case role == "unit":
			switch field {
			case "update":
				var ptr *server.MonsterUpdateData
				if v != 0 {
					i := (v - 0x21003000) / 0x2000
					if i > 1 || v != 0x21003000+i*0x2000 {
						t.Fatal("original pointer role", v)
					}
					ptr = w.Data[i]
				}
				write(unsafe.Pointer(&w.Unit.UpdateData), uintptr(unsafe.Pointer(ptr)), unsafe.Sizeof(uintptr(0)))
			case "x":
				write(unsafe.Pointer(&w.Unit.PosVec.X), uintptr(v), 4)
			case "y":
				write(unsafe.Pointer(&w.Unit.PosVec.Y), uintptr(v), 4)
			case "subclass":
				write(unsafe.Pointer(&w.Unit.ObjSubClass), uintptr(v), 4)
			case "class":
				write(unsafe.Pointer(&w.Unit.ObjClass), uintptr(v), 4)
			default:
				t.Fatal("unit write", field)
			}
		case role == "data0" || role == "data1":
			d := w.Data[int(role[4]-'0')]
			var p unsafe.Pointer
			switch field {
			case "marker":
				p = unsafe.Pointer(&d.Field97)
			case "stamp":
				p = unsafe.Pointer(&d.Field101)
			case "x":
				p = unsafe.Pointer(&d.Field99X)
			case "y":
				p = unsafe.Pointer(&d.Field99Y)
			default:
				t.Fatal("data write", field)
			}
			write(p, uintptr(v), 4)
		case len(role) == 5 && role[:4] == "item":
			s := w.Items[int(role[4]-'0')]
			if field == "action" {
				write(unsafe.Pointer(&s.Action), uintptr(v), 4)
			} else {
				i, err := strconv.Atoi(field[3:])
				if err != nil || i < 0 || i >= 4 {
					t.Fatal("item write", field)
				}
				write(unsafe.Pointer(&s.Args[i]), uintptr(v), unsafe.Sizeof(uintptr(0)))
			}
		case role == "world":
			if field != "frame" && field != "fps" && field != "rng" {
				t.Fatal("world write", field)
			}
		default:
			t.Fatal("unknown original write", role, field)
		}
	}
	if !bytes.Equal(expected, w.Raw) {
		for i := range expected {
			if expected[i] != w.Raw[i] {
				t.Fatalf("whole native records/padding/guards byte%d got%02x want%02x", i, w.Raw[i], expected[i])
			}
		}
	}
	if got := w.state(); !reflect.DeepEqual(got, row.After) {
		t.Fatal("original final native state", got, row.After)
	}
	if !reflect.DeepEqual(w.Events, row.Events) {
		t.Fatalf("original ordered services got%#v want%#v", w.Events, row.Events)
	}
}
func TestMonsterInterestingOriginal5466F0(t *testing.T) {
	for _, row := range interestingLoad5466F0(t) {
		t.Run(row.Name, func(t *testing.T) {
			w := interestingWorldNew5466F0(t, row)
			if row.Fault != nil {
				panicked := false
				func() { defer func() { panicked = recover() != nil }(); w.call(row) }()
				if !panicked {
					t.Fatal("original raw nil fault must remain")
				}
			} else if result := w.call(row); result != row.Result {
				t.Fatal("original DWORD result", result, row.Result)
			}
			w.check(t, row)
		})
	}
}
func interestingLayoutCheck5466F0(t *testing.T) {
	t.Helper()
	got := monsterInterestingOriginalLayout5466F0()
	want := [10]uintptr{unsafe.Offsetof(server.Object{}.ObjClass), unsafe.Offsetof(server.Object{}.ObjSubClass), unsafe.Offsetof(server.Object{}.PosVec), unsafe.Offsetof(server.Object{}.UpdateData), unsafe.Offsetof(server.MonsterUpdateData{}.Field97), unsafe.Offsetof(server.MonsterUpdateData{}.Field99X), unsafe.Offsetof(server.MonsterUpdateData{}.Field101), unsafe.Sizeof(server.AIStackItem{}), unsafe.Offsetof(server.AIStackItem{}.Args), unsafe.Offsetof(server.AIStackItem{}.Field5)}
	if got != want {
		t.Fatal("independent C/Go native layouts", got, want)
	}
}
func TestMonsterInterestingNativeLayout5466F0(t *testing.T) { interestingLayoutCheck5466F0(t) }
func TestMonsterInterestingNativeCBefore5466F0(t *testing.T) {
	if os.Getenv("OPENNOX_INTERESTING_C_BEFORE") != "1" {
		return
	}
	if os.Getenv("OPENNOX_INTERESTING_C_BEFORE_CHILD") == "1" {
		interestingLayoutCheck5466F0(t)
		for _, row := range interestingLoad5466F0(t) {
			if row.Name == "push/1/31/0/01230" {
				w := interestingWorldNew5466F0(t, row)
				if unsafe.Sizeof(uintptr(0)) != 8 {
					t.Fatal("baseline requires64bit")
				}
				if !bytes.Equal(w.Raw, w.Before) {
					t.Fatal("initial full records/canaries")
				}
				fmt.Printf("INTERESTING_C_BEFORE_VALID_NATIVE_CONTROLS unit=%p MUD=%p pos=%p\n", w.Unit, w.Data[0], &w.Data[0].Field99X)
				w.call(row)
				t.Fatal("old C entry unexpectedly returned")
			}
		}
		t.Fatal("missing frozen baseline context")
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestMonsterInterestingNativeCBefore5466F0$", "-test.v")
	cmd.Env = append(os.Environ(), "OPENNOX_INTERESTING_C_BEFORE_CHILD=1")
	out, err := cmd.CombinedOutput()
	t.Log(string(out))
	if err == nil || !bytes.Contains(out, []byte("INTERESTING_C_BEFORE_VALID_NATIVE_CONTROLS")) || (!strings.Contains(string(out), "SIGSEGV") && !strings.Contains(string(out), "SIGBUS")) || !bytes.Contains(out, []byte("_Cfunc_interesting_public_5466f0")) {
		t.Fatalf("actual public C fault after valid native controls: err=%v", err)
	}
}
