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
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type idleInput5469B0 struct {
	Name, Root                     string
	UnitNil                        bool `json:"unit_nil"`
	Update                         int
	Cls                            uint32
	Aggression, Distance, Deadline [2]uint32
	Enemy, Sets                    [2]int
	Indices                        [2]int8
	Actions, IDs                   [2]uint32
	Frame, FPS, Seed               uint32
	Mutation                       struct {
		Kind  string
		Value int64
	}
}
type idleState5469B0 struct {
	Update int
	Cls    uint32
	Data   [2]struct {
		Set, Enemy                     int
		Distance, Deadline, Aggression uint32
		Index                          int8
		Actions                        [24]uint32
	}
	IDs             [4]uint32
	Frame, FPS, RNG uint32
}
type idleRow5469B0 struct {
	Name          string
	Input         idleInput5469B0
	Before, After idleState5469B0
	Events        []map[string]any
	Writes        []struct {
		Role  [2]string
		Value uint32
	}
	Fault  any
	Result int
	Void   bool `json:"root_void_has_no_semantic_EAX_result"`
	Whole  bool `json:"whole_original_heap_changes_within_declared_writes"`
}

func idleLoad5469B0(t *testing.T) []idleRow5469B0 {
	t.Helper()
	b, err := os.ReadFile("../server/testdata/monster_idle_sound_original_5469b0.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(b)) != "ae7d8a867fa0981d17f3c76c920415e45f96989fe07e49f460913eb59b35d31d" {
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
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "5315122cb88a7e9c2cf228139fc2dbb2bb9d4ba06b28a03ce12f5d5d0a6fa1d2" {
		t.Fatal("sealed decoded original")
	}
	var data struct {
		OriginalSHA string `json:"original_sha256"`
		Rows        []idleRow5469B0
	}
	if err = json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	if data.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(data.Rows) != 5670 {
		t.Fatal("original executable/dimensions")
	}
	return data.Rows
}

type idleWorld5469B0 struct {
	Unit            *server.Object
	Data            [2]*server.MonsterUpdateData
	Enemies         [2]*server.Object
	Sets            [4]unsafe.Pointer
	Base            unsafe.Pointer
	Raw, Before     []byte
	Frame, FPS, RNG uint32
	Events          []map[string]any
	Runtime         monsterIdleOriginalRuntime5469B0
}

func idleWorldNew5469B0(t *testing.T, row idleRow5469B0) *idleWorld5469B0 {
	t.Helper()
	const size = 32768
	base, free := alloc.MallocHigh(size)
	t.Cleanup(free)
	w := &idleWorld5469B0{Base: base, Raw: unsafe.Slice((*byte)(base), size), Frame: row.Input.Frame, FPS: row.Input.FPS, RNG: row.Input.Seed, Events: make([]map[string]any, 0)}
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
		ptr := unsafe.Add(base, start)
		if unsafe.Sizeof(ptr) == 8 && uintptr(ptr) <= math.MaxUint32 {
			t.Fatal("native allocation must exceed4GiB")
		}
		return ptr
	}
	w.Unit = (*server.Object)(take(unsafe.Sizeof(server.Object{})))
	for i := range w.Data {
		w.Data[i] = (*server.MonsterUpdateData)(take(unsafe.Sizeof(server.MonsterUpdateData{})))
	}
	for i := range w.Enemies {
		w.Enemies[i] = (*server.Object)(take(unsafe.Sizeof(server.Object{})))
	}
	w.Sets[0], w.Sets[1] = take(64), take(64)
	w.Sets[2], w.Sets[3] = unsafe.Pointer(&w.Data[0].Field128), unsafe.Pointer(&w.Data[1].Field128)
	w.Unit.ObjClass = object.Class(row.Input.Cls)
	w.Unit.UpdateData = unsafe.Pointer(w.data(row.Input.Update))
	for i, d := range w.Data {
		d.Aggression = math.Float32frombits(row.Input.Aggression[i])
		d.Field131 = row.Input.Distance[i]
		d.Field132 = row.Input.Deadline[i]
		d.CurrentEnemy = w.enemy(row.Input.Enemy[i])
		d.SoundSet122 = w.set(row.Input.Sets[i])
		d.AIStackInd = row.Input.Indices[i]
		for j := range d.AIStack {
			d.AIStack[j].Action = 0xa5000000 + uint32(j)
			for k := range d.AIStack[j].Args {
				high := uintptr(0x12345678)
				high <<= 32
				d.AIStack[j].Args[k] = high | 0xa5a5a5a5
			}
		}
		if d.AIStackInd >= 0 {
			d.AIStack[d.AIStackInd].Action = row.Input.Actions[i]
		}
	}
	for i := 0; i < 2; i++ {
		*(*uint32)(unsafe.Add(w.Sets[i], 16)) = row.Input.IDs[i]
	}
	w.Before = bytes.Clone(w.Raw)
	w.Runtime = monsterIdleOriginalRuntime5469B0{
		Frame: func() uint32 {
			w.Events = append(w.Events, map[string]any{"kind": "frame", "value": float64(w.Frame)})
			return w.Frame
		},
		TickRate: func() uint32 {
			w.Events = append(w.Events, map[string]any{"kind": "fps", "value": float64(w.FPS)})
			return w.FPS
		},
		Random: func(min, max int) int {
			i := len(w.Events)
			if i >= len(row.Events) || row.Events[i]["kind"] != "random" {
				t.Fatal("unexpected RNG call")
			}
			expected := row.Events[i]
			if min != int(int32(uint32(expected["min"].(float64)))) || max != int(int32(uint32(expected["max"].(float64)))) || w.RNG != uint32(expected["before"].(float64)) {
				t.Fatal("signed DWORD RNG bounds/index", min, max, w.RNG, expected)
			}
			r := prand.New(int(w.RNG))
			result := r.IntClamp(min, max)
			w.RNG = uint32(r.Index())
			w.Events = append(w.Events, map[string]any{"kind": "random", "min": float64(uint32(min)), "max": float64(uint32(max)), "before": expected["before"], "after": float64(w.RNG), "result": float64(uint32(result))})
			return int(int32(result))
		},
		Audio: func(id uint32, unit *server.Object, a, b uint32) {
			if unit != w.Unit || a != 0 || b != 0 {
				t.Fatal("native audio identities/flags")
			}
			w.Events = append(w.Events, map[string]any{"kind": "audio", "id": float64(id), "unit": "unit", "flags": []any{float64(a), float64(b)}})
			m := row.Input.Mutation
			switch m.Kind {
			case "none":
			case "update":
				w.Unit.UpdateData = unsafe.Pointer(w.data(int(m.Value)))
			case "class":
				w.Unit.ObjClass = object.Class(uint32(m.Value))
			case "timer0", "timer1":
				w.Data[int(m.Kind[len(m.Kind)-1]-'0')].Field132 = uint32(m.Value)
			case "set0", "set1":
				w.Data[int(m.Kind[len(m.Kind)-1]-'0')].SoundSet122 = w.set(int(m.Value))
			case "id0", "id1":
				*(*uint32)(unsafe.Add(w.Sets[int(m.Kind[len(m.Kind)-1]-'0')], 16)) = uint32(m.Value)
			case "frame":
				w.Frame = uint32(m.Value)
			case "fps":
				w.FPS = uint32(m.Value)
			default:
				t.Fatal("unknown frozen audio mutation", m)
			}
		},
	}
	if got := w.state(); !reflect.DeepEqual(got, row.Before) {
		t.Fatal("independent initial native state", got, row.Before)
	}
	return w
}
func (w *idleWorld5469B0) data(index int) *server.MonsterUpdateData {
	if index < 0 {
		return nil
	}
	return w.Data[index]
}
func (w *idleWorld5469B0) enemy(index int) *server.Object {
	if index < 0 {
		return nil
	}
	return w.Enemies[index]
}
func (w *idleWorld5469B0) set(index int) unsafe.Pointer {
	if index < 0 {
		return nil
	}
	return w.Sets[index]
}
func (w *idleWorld5469B0) setRole(ptr uintptr) int {
	if ptr == 0 {
		return -1
	}
	for i, p := range w.Sets {
		if uintptr(p) == ptr {
			return i
		}
	}
	panic("unknown native sound pointer")
}
func (w *idleWorld5469B0) state() idleState5469B0 {
	s := idleState5469B0{Update: -1, Cls: uint32(w.Unit.ObjClass), Frame: w.Frame, FPS: w.FPS, RNG: w.RNG}
	for i, d := range w.Data {
		if w.Unit.UpdateData == unsafe.Pointer(d) {
			s.Update = i
		}
		v := &s.Data[i]
		v.Set = w.setRole(uintptr(d.SoundSet122))
		v.Enemy = -1
		for j, e := range w.Enemies {
			if d.CurrentEnemy == e {
				v.Enemy = j
			}
		}
		v.Distance = d.Field131
		v.Deadline = d.Field132
		v.Aggression = math.Float32bits(d.Aggression)
		v.Index = d.AIStackInd
		for j := range d.AIStack {
			v.Actions[j] = d.AIStack[j].Action
		}
	}
	for i, ptr := range w.Sets {
		s.IDs[i] = *(*uint32)(unsafe.Add(ptr, 16))
	}
	return s
}

type idleNativeEntry5469B0 interface {
	MonsterIdleSoundNative5469B0(func() uint32, func() uint32, func(int, int) int, func(uint32, *server.Object, uint32, uint32))
}

func (w *idleWorld5469B0) call(t *testing.T, row idleRow5469B0, path string) int {
	t.Helper()
	unit := w.Unit
	if row.Input.UnitNil {
		unit = nil
	}
	if path == "C" {
		old := monsterIdleRuntimeOverride5469B0
		monsterIdleRuntimeOverride5469B0 = &w.Runtime
		defer func() { monsterIdleRuntimeOverride5469B0 = old }()
		result := idleOriginalC5469B0(row.Input.Root, unit)
		if row.Input.Root == "SoundSet" {
			return w.setRole(result)
		}
		return 0
	}
	if row.Input.Root == "SoundSet" {
		return w.setRole(idleOriginalSoundSetGo5469B0(unit))
	}
	entry, ok := any(unit).(idleNativeEntry5469B0)
	if !ok {
		t.Fatal("native Go idle entry not implemented")
	}
	entry.MonsterIdleSoundNative5469B0(w.Runtime.Frame, w.Runtime.TickRate, w.Runtime.Random, w.Runtime.Audio)
	return 0
}
func (w *idleWorld5469B0) check(t *testing.T, row idleRow5469B0) {
	t.Helper()
	if !row.Whole {
		t.Fatal("original whole-heap write coverage")
	}
	expected := bytes.Clone(w.Before)
	write := func(ptr unsafe.Pointer, value uintptr, size uintptr) {
		off := int(uintptr(ptr) - uintptr(w.Base))
		if off < 0 || off+int(size) > len(expected) {
			t.Fatal("projected original write outside native allocation")
		}
		if size == 8 {
			binary.LittleEndian.PutUint64(expected[off:], uint64(value))
		} else {
			binary.LittleEndian.PutUint32(expected[off:], uint32(value))
		}
	}
	decodePtr := func(value uint32, original uint32, stride uint32, count uint32) int {
		if value == 0 {
			return -1
		}
		i := (value - original) / stride
		if i >= count || value != original+i*stride {
			t.Fatal("original pointer role", value)
		}
		return int(i)
	}
	for _, change := range row.Writes {
		role, field, value := change.Role[0], change.Role[1], change.Value
		switch {
		case role == "unit":
			switch field {
			case "class":
				write(unsafe.Pointer(&w.Unit.ObjClass), uintptr(value), 4)
			case "update":
				i := decodePtr(value, 0x21003000, 0x2000, 2)
				write(unsafe.Pointer(&w.Unit.UpdateData), uintptr(unsafe.Pointer(w.data(i))), unsafe.Sizeof(uintptr(0)))
			default:
				t.Fatal("unit write", field)
			}
		case role == "data0" || role == "data1":
			d := w.Data[int(role[4]-'0')]
			switch field {
			case "deadline":
				write(unsafe.Pointer(&d.Field132), uintptr(value), 4)
			case "distance":
				write(unsafe.Pointer(&d.Field131), uintptr(value), 4)
			case "aggression":
				write(unsafe.Pointer(&d.Aggression), uintptr(value), 4)
			case "enemy":
				i := decodePtr(value, 0x21007000, 0x400, 2)
				write(unsafe.Pointer(&d.CurrentEnemy), uintptr(unsafe.Pointer(w.enemy(i))), unsafe.Sizeof(uintptr(0)))
			case "set":
				i := -1
				if value != 0 {
					original := [4]uint32{0x21008000, 0x21008100, 0x21003200, 0x21005200}
					for j, ptr := range original {
						if value == ptr {
							i = j
						}
					}
					if i < 0 {
						t.Fatal("original sound pointer role", value)
					}
				}
				write(unsafe.Pointer(&d.SoundSet122), uintptr(w.set(i)), unsafe.Sizeof(uintptr(0)))
			default:
				t.Fatal("data write", field)
			}
		case role == "set0" || role == "set1":
			if field != "id" {
				t.Fatal("set write", field)
			}
			write(unsafe.Add(w.Sets[int(role[3]-'0')], 16), uintptr(value), 4)
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
		t.Fatalf("original ordered effects got%#v want%#v", w.Events, row.Events)
	}
}
func TestMonsterIdleSoundOriginal5469B0(t *testing.T) {
	rows := idleLoad5469B0(t)
	for _, path := range []string{"C", "Go"} {
		t.Run(path, func(t *testing.T) {
			for _, row := range rows {
				t.Run(row.Name, func(t *testing.T) {
					w := idleWorldNew5469B0(t, row)
					if row.Fault != nil {
						panicked := false
						func() { defer func() { panicked = recover() != nil }(); w.call(t, row, path) }()
						if !panicked {
							t.Fatal("original nil fault must remain")
						}
					} else {
						result := w.call(t, row, path)
						if row.Input.Root == "Idle" {
							if !row.Void {
								t.Fatal("void root has no original EAX result")
							}
						} else if result != row.Result {
							t.Fatal("original full sound pointer role", result, row.Result)
						}
					}
					w.check(t, row)
				})
			}
		})
	}
}
func idleLayoutCheck5469B0(t *testing.T) {
	t.Helper()
	got := idleOriginalLayout5469B0()
	want := [11]uintptr{unsafe.Offsetof(server.Object{}.ObjClass), unsafe.Offsetof(server.Object{}.UpdateData), unsafe.Offsetof(server.MonsterUpdateData{}.SoundSet122), unsafe.Offsetof(server.MonsterUpdateData{}.Field128), unsafe.Offsetof(server.MonsterUpdateData{}.Field131), unsafe.Offsetof(server.MonsterUpdateData{}.Field132), unsafe.Offsetof(server.MonsterUpdateData{}.AIStackInd), unsafe.Offsetof(server.MonsterUpdateData{}.AIStack), unsafe.Offsetof(server.MonsterUpdateData{}.CurrentEnemy), unsafe.Offsetof(server.MonsterUpdateData{}.Aggression), unsafe.Sizeof(server.AIStackItem{})}
	if got != want {
		t.Fatal("independent C/Go native layouts", got, want)
	}
	if unsafe.Offsetof(server.MonsterUpdateData{}.Field132)-unsafe.Offsetof(server.MonsterUpdateData{}.Field128) != 16 {
		t.Fatal("native timer sound-set alias")
	}
}
func TestMonsterIdleSoundNativeLayout5469B0(t *testing.T) { idleLayoutCheck5469B0(t) }
func TestMonsterIdleSoundNativeCBefore5469B0(t *testing.T) {
	if os.Getenv("OPENNOX_IDLE_C_BEFORE") != "1" {
		return
	}
	if os.Getenv("OPENNOX_IDLE_C_BEFORE_CHILD") == "1" {
		idleLayoutCheck5469B0(t)
		for _, row := range idleLoad5469B0(t) {
			if row.Name == "clock/00000064/00000064/0000001e/2/0/0" {
				w := idleWorldNew5469B0(t, row)
				if unsafe.Sizeof(uintptr(0)) != 8 {
					t.Fatal("baseline requires64bit")
				}
				if idleOriginalC5469B0("SoundSet", w.Unit) != uintptr(w.Sets[2]) || idleOriginalSoundSetGo5469B0(w.Unit) != uintptr(w.Sets[2]) || *(*uint32)(unsafe.Add(w.Sets[2], 16)) != 100 || idleOriginalActionC5469B0(w.Unit) != 0 {
					t.Fatal("actual existing sound/action public C and Go controls")
				}
				if !bytes.Equal(w.Before, w.Raw) {
					t.Fatal("full initial guards/records changed")
				}
				fmt.Printf("IDLE_C_BEFORE_VALID_NATIVE_CONTROLS unit=%p MUD=%p sound=%p actual-existing-sound-getter-and-action-getter=true\n", w.Unit, w.Data[0], w.Sets[2])
				w.call(t, row, "C")
				t.Fatal("old public C unexpectedly returned")
			}
		}
		t.Fatal("missing frozen baseline input")
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestMonsterIdleSoundNativeCBefore5469B0$", "-test.v")
	cmd.Env = append(os.Environ(), "OPENNOX_IDLE_C_BEFORE_CHILD=1")
	out, err := cmd.CombinedOutput()
	t.Logf("%s", out)
	if err == nil || !bytes.Contains(out, []byte("IDLE_C_BEFORE_VALID_NATIVE_CONTROLS")) || (!strings.Contains(string(out), "SIGSEGV") && !strings.Contains(string(out), "SIGBUS")) || !bytes.Contains(out, []byte("_Cfunc_idle_original_public_5469b0")) {
		t.Fatal("actual public C fault after native controls", err)
	}
}
