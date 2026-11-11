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
	"strconv"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type reactiveInput545DA0 struct {
	Name, Root              string
	UnitNil                 bool `json:"unit_nil"`
	Update                  int
	Cls                     uint32
	Status, Markers, Stamps [2]uint32
	XY                      [2][2]uint32
	Position                [2]uint32
	Frame, FPS, Seed        uint32
	Mask                    uint32
	Returns                 [4]int
	Mutation                struct {
		At    int
		Kind  string
		Value int
	}
}
type reactiveState545DA0 struct {
	Update   int
	Position [2]uint32
	Cls      uint32
	Data     [2]struct {
		Marker, Stamp, Status uint32
		XY                    [2]uint32
	}
	Items           [4][5]uint32
	Frame, FPS, RNG uint32
}
type reactiveRow545DA0 struct {
	Name          string
	Input         reactiveInput545DA0
	Before, After reactiveState545DA0
	Events        []map[string]any
	Writes        []struct {
		Role  [2]string
		Value uint32
	}
	Fault  any
	Result uint32
	Whole  bool `json:"whole_original_heap_changes_within_declared_writes"`
}

func reactiveLoad545DA0(t *testing.T) []reactiveRow545DA0 {
	t.Helper()
	b, err := os.ReadFile("../server/testdata/monster_reactive_original_545da0.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(b)) != "21f8a0a310bf64bf68da1560facc6d3b5370cc9057ea8a3b5c79ec5373f01a9a" {
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
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "d503d68bda0d36bde27074c7fe7cee3c14d6156940017bfd84aaebf89737db77" {
		t.Fatal("sealed decoded original")
	}
	var data struct {
		OriginalSHA string `json:"original_sha256"`
		Rows        []reactiveRow545DA0
	}
	if err = json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	if data.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(data.Rows) != 2836 {
		t.Fatal("original executable/dimensions")
	}
	return data.Rows
}

type reactiveWorld545DA0 struct {
	Unit            *server.Object
	Data            [2]*server.MonsterUpdateData
	Items           [4]*server.AIStackItem
	Raw, Before     []byte
	Base            unsafe.Pointer
	Frame, FPS, RNG uint32
	Events          []map[string]any
	Pushes          int
	Runtime         server.MonsterReactiveRuntime545DA0
}

func reactiveWorldNew545DA0(t *testing.T, row reactiveRow545DA0) *reactiveWorld545DA0 {
	t.Helper()
	const size = 32768
	base, free := alloc.MallocHigh(size)
	t.Cleanup(free)
	w := &reactiveWorld545DA0{Base: base, Raw: unsafe.Slice((*byte)(base), size), Frame: row.Input.Frame, FPS: row.Input.FPS, RNG: row.Input.Seed, Events: make([]map[string]any, 0)}
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
			t.Fatal("guarded allocation")
		}
		p := unsafe.Add(base, start)
		if unsafe.Sizeof(p) > 4 && uintptr(p) <= math.MaxUint32 {
			t.Fatal("actual native address must exceed4GiB")
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
	w.Unit.UpdateData = unsafe.Pointer(w.data(row.Input.Update))
	w.Unit.Pos132.X = math.Float32frombits(row.Input.Position[0])
	w.Unit.Pos132.Y = math.Float32frombits(row.Input.Position[1])
	for i, d := range w.Data {
		d.Field97 = row.Input.Markers[i]
		d.Field99X = math.Float32frombits(row.Input.XY[i][0])
		d.Field99Y = math.Float32frombits(row.Input.XY[i][1])
		d.Field101 = row.Input.Stamps[i]
		d.StatusFlags = object.MonsterStatus(row.Input.Status[i])
	}
	for _, s := range w.Items {
		s.Action = 0xa5a5a5a5
		for i := range s.Args {
			s.Args[i] = 0xa5a5a5a5
		}
	}
	w.Before = bytes.Clone(w.Raw)
	appendEvent := func(e map[string]any) { w.Events = append(w.Events, e) }
	w.Runtime = server.MonsterReactiveRuntime545DA0{
		Frame:    func() uint32 { appendEvent(map[string]any{"kind": "frame", "value": float64(w.Frame)}); return w.Frame },
		TickRate: func() uint32 { appendEvent(map[string]any{"kind": "fps", "value": float64(w.FPS)}); return w.FPS },
		Random: func(min, max int) int {
			i := len(w.Events)
			if i >= len(row.Events) || row.Events[i]["kind"] != "random" {
				t.Fatal("unexpected RNG call")
			}
			expected := row.Events[i]
			if min != int(int32(uint32(expected["min"].(float64)))) || max != int(int32(uint32(expected["max"].(float64)))) || w.RNG != uint32(expected["before"].(float64)) {
				t.Fatal("exact signed32 RNG arguments/index", min, max, w.RNG, expected)
			}
			r := prand.New(int(w.RNG))
			result := r.IntClamp(min, max)
			w.RNG = uint32(r.Index())
			appendEvent(map[string]any{"kind": "random", "min": float64(uint32(min)), "max": float64(uint32(max)), "before": expected["before"], "after": float64(w.RNG), "result": float64(uint32(result))})
			return int(int32(result))
		},
		Push: func(unit *server.Object, action ai.ActionType) *server.AIStackItem {
			if unit != w.Unit || w.Pushes >= 4 {
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
			m := row.Input.Mutation
			if m.At == i {
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
						w.Unit.Pos132.X = math.Float32frombits(0x7f812345)
						w.Unit.Pos132.Y = math.Float32frombits(0xff800000)
					} else {
						w.Unit.Pos132.X = math.Float32frombits(0x80000000)
						w.Unit.Pos132.Y = math.Float32frombits(1)
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
				case "status0":
					if m.Value != 0 {
						w.Data[0].StatusFlags = 0
					} else {
						w.Data[0].StatusFlags = object.MonsterStatus(0xffffffff)
					}
				case "none":
				default:
					t.Fatal("unknown frozen mutation", m)
				}
			}
			return item
		},
	}
	if got := w.state(); !reflect.DeepEqual(got, row.Before) {
		t.Fatal("independent initial native state", got, row.Before)
	}
	return w
}
func (w *reactiveWorld545DA0) data(role int) *server.MonsterUpdateData {
	if role < 0 {
		return nil
	}
	return w.Data[role]
}
func (w *reactiveWorld545DA0) state() reactiveState545DA0 {
	s := reactiveState545DA0{Update: -1, Cls: uint32(w.Unit.ObjClass), Position: [2]uint32{math.Float32bits(w.Unit.Pos132.X), math.Float32bits(w.Unit.Pos132.Y)}, Frame: w.Frame, FPS: w.FPS, RNG: w.RNG}
	for i, d := range w.Data {
		if w.Unit.UpdateData == unsafe.Pointer(d) {
			s.Update = i
		}
		s.Data[i].Marker = d.Field97
		s.Data[i].Stamp = d.Field101
		s.Data[i].Status = uint32(d.StatusFlags)
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
func (w *reactiveWorld545DA0) call(t *testing.T, row reactiveRow545DA0, path string) uint32 {
	t.Helper()
	unit := w.Unit
	if row.Input.UnitNil {
		unit = nil
	}
	if path == "C" {
		old := monsterReactiveRuntimeOverride545DA0
		monsterReactiveRuntimeOverride545DA0 = &w.Runtime
		defer func() { monsterReactiveRuntimeOverride545DA0 = old }()
		return uint32(monsterReactiveOriginalC545DA0(row.Input.Root, unit))
	}
	type entries interface {
		MonsterRecentSound545DA0(server.MonsterReactiveRuntime545DA0) bool
		MonsterLookAtDamager5466B0(server.MonsterReactiveRuntime545DA0) bool
	}
	var result bool
	if row.Input.Root == "Sound" {
		result = any(unit).(entries).MonsterRecentSound545DA0(w.Runtime)
	} else {
		result = any(unit).(entries).MonsterLookAtDamager5466B0(w.Runtime)
	}
	if result {
		return 1
	}
	return 0
}
func (w *reactiveWorld545DA0) check(t *testing.T, row reactiveRow545DA0) {
	t.Helper()
	if !row.Whole {
		t.Fatal("original whole-heap write coverage")
	}
	expected := bytes.Clone(w.Before)
	write := func(p unsafe.Pointer, v uintptr, n uintptr) {
		off := int(uintptr(p) - uintptr(w.Base))
		if off < 0 || off+int(n) > len(expected) {
			t.Fatal("projected original write outside heap")
		}
		if n == 8 {
			binary.LittleEndian.PutUint64(expected[off:], uint64(v))
		} else {
			binary.LittleEndian.PutUint32(expected[off:], uint32(v))
		}
	}
	for _, change := range row.Writes {
		role, field, v := change.Role[0], change.Role[1], change.Value
		switch {
		case role == "unit":
			switch field {
			case "update":
				var p *server.MonsterUpdateData
				if v != 0 {
					i := (v - 0x21003000) / 0x2000
					if i > 1 || v != 0x21003000+i*0x2000 {
						t.Fatal("original pointer role", v)
					}
					p = w.Data[i]
				}
				write(unsafe.Pointer(&w.Unit.UpdateData), uintptr(unsafe.Pointer(p)), unsafe.Sizeof(uintptr(0)))
			case "x":
				write(unsafe.Pointer(&w.Unit.Pos132.X), uintptr(v), 4)
			case "y":
				write(unsafe.Pointer(&w.Unit.Pos132.Y), uintptr(v), 4)
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
			case "status":
				p = unsafe.Pointer(&d.StatusFlags)
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
func TestMonsterReactiveOriginal545DA0(t *testing.T) {
	rows := reactiveLoad545DA0(t)
	for _, path := range []string{"C", "Go"} {
		t.Run(path, func(t *testing.T) {
			for _, row := range rows {
				t.Run(row.Name, func(t *testing.T) {
					w := reactiveWorldNew545DA0(t, row)
					if row.Fault != nil {
						panicked := false
						func() { defer func() { panicked = recover() != nil }(); w.call(t, row, path) }()
						if !panicked {
							t.Fatal("original raw nil fault must remain")
						}
					} else if result := w.call(t, row, path); result != row.Result {
						t.Fatal("original DWORD result", result, row.Result)
					}
					w.check(t, row)
				})
			}
		})
	}
}
func TestMonsterReactiveNativeLayout545DA0(t *testing.T) {
	got := monsterReactiveOriginalLayout545DA0()
	want := [8]uintptr{unsafe.Offsetof(server.MonsterUpdateData{}.Field97), unsafe.Offsetof(server.MonsterUpdateData{}.Field99X), unsafe.Offsetof(server.MonsterUpdateData{}.Field101), unsafe.Offsetof(server.Object{}.Pos132), unsafe.Offsetof(server.Object{}.UpdateData), unsafe.Sizeof(server.AIStackItem{}), unsafe.Offsetof(server.AIStackItem{}.Args), unsafe.Offsetof(server.AIStackItem{}.Field5)}
	if got != want {
		t.Fatal("independent C/Go native layouts", got, want)
	}
}
