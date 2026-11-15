package legacy

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/fireball_original_52c790_v1.json.gz
var fireballOriginalV1 []byte

//go:embed testdata/fireball_original_52c790_v2.json.gz
var fireballOriginal []byte

//go:embed testdata/fireball_original_direction_table_52c790_v1.bin
var fireballOriginalDirections []byte

type fireballObjectState struct {
	Position   [2]uint32 `json:"position"`
	Velocity   [2]uint32 `json:"velocity"`
	Radius     uint32    `json:"radius"`
	Speed      uint32    `json:"speed"`
	Directions [2]uint16 `json:"directions"`
}

type fireballOriginalEvent struct {
	Kind   string                `json:"kind"`
	Args   []any                 `json:"args"`
	Before []fireballObjectState `json:"before"`
}

type fireballOriginalCase struct {
	Input struct {
		Name       string                `json:"name"`
		Spell      int32                 `json:"spell"`
		Level      int32                 `json:"level"`
		Firing     int                   `json:"firing"`
		Allocation bool                  `json:"allocation"`
		Trace      int32                 `json:"trace"`
		Balance    uint64                `json:"balance_bits64"`
		Sound      uint32                `json:"sound"`
		Objects    []fireballObjectState `json:"objects"`
		Mutation   resourceSpellMutation `json:"mutation"`
	} `json:"input"`
	Result       int32                   `json:"result"`
	Events       []fireballOriginalEvent `json:"events"`
	Mutations    []resourceSpellMutation `json:"mutations"`
	Before       []fireballObjectState   `json:"before"`
	After        []fireballObjectState   `json:"after"`
	CallerBefore string                  `json:"original_caller_before_hex"`
	CallerAfter  string                  `json:"original_caller_after_hex"`
	HeapBefore   string                  `json:"original_heap_before_sha256"`
	HeapAfter    string                  `json:"original_heap_after_sha256"`
	XTemporary   []uint32                `json:"original_x_velocity_temporary"`
}

func fireballCases(t *testing.T) []fireballOriginalCase {
	t.Helper()
	var rows []fireballOriginalCase
	npcVoiceOriginalDecode(t, fireballOriginal, "2bd29c1894546f8a2c39f4ebb049aab32720e746563d4e307bfb1291a5876001", "65a01f46616e2ff348e5b4abbd17650cfab99649da01e8372baec1ebd6c7e57d", 243, &rows)
	var oldRaw, newRaw []json.RawMessage
	npcVoiceOriginalDecode(t, fireballOriginalV1, "3ce1fb1ff831c6c3f9f32907eeee6b30102b965a9e45abafccf577f58180b62e", "14fedb92da94f5a19479cfa4f5fde7f82d0b07dd3f75457c65dddc7d1c814146", 213, &oldRaw)
	npcVoiceOriginalDecode(t, fireballOriginal, "2bd29c1894546f8a2c39f4ebb049aab32720e746563d4e307bfb1291a5876001", "65a01f46616e2ff348e5b4abbd17650cfab99649da01e8372baec1ebd6c7e57d", 243, &newRaw)
	for i := range oldRaw {
		if !bytes.Equal(oldRaw[i], newRaw[i]) {
			t.Fatal("original pre-port213case prefix changed", i)
		}
	}
	return rows
}

func fireballOriginalHeap(s []fireballObjectState) []byte {
	out := make([]byte, 65536)
	for i := range out {
		out[i] = byte(i*31 + 19)
	}
	for i, v := range s {
		a := 0x1000 + i*0x2000
		for off, value := range map[int]uint32{56: v.Position[0], 60: v.Position[1], 80: v.Velocity[0], 84: v.Velocity[1], 176: v.Radius, 544: v.Speed} {
			binary.LittleEndian.PutUint32(out[a+off:], value)
		}
		binary.LittleEndian.PutUint16(out[a+124:], v.Directions[0])
		binary.LittleEndian.PutUint16(out[a+126:], v.Directions[1])
	}
	return out
}

func fireballVerifyOriginal(t *testing.T, c fireballOriginalCase) {
	t.Helper()
	if !reflect.DeepEqual(c.Input.Objects, c.Before) || fmt.Sprintf("%x", sha256.Sum256(fireballOriginalHeap(c.Before))) != c.HeapBefore || fmt.Sprintf("%x", sha256.Sum256(fireballOriginalHeap(c.After))) != c.HeapAfter {
		t.Fatal("original whole heap differs from exact input/declared fields and canaries")
	}
	a, b := npcVoiceUnhex(t, c.CallerBefore), npcVoiceUnhex(t, c.CallerAfter)
	if len(a) != 28 || len(b) != 28 || !bytes.Equal(a[:16], b[:16]) || !bytes.Equal(a[20:24], b[20:24]) {
		t.Fatal("original caller escaped two declared binary32 spill words")
	}
	if !c.Input.Allocation {
		if !bytes.Equal(a, b) || len(c.XTemporary) != 0 {
			t.Fatal("allocation failure changed caller or floating temporary")
		}
		return
	}
	if len(c.XTemporary) != 1 || binary.LittleEndian.Uint32(b[24:]) != c.XTemporary[0] {
		t.Fatal("original X temporary differs from caller word")
	}
	for _, event := range c.Events {
		if event.Kind == "trace" {
			direction := int16(event.Before[c.Input.Firing].Directions[0])
			if direction < 0 || direction > 255 || binary.LittleEndian.Uint32(b[16:]) != binary.LittleEndian.Uint32(fireballOriginalDirections[8*int(direction)+4:]) {
				t.Fatal("original caller directionY spill differs from immutable stock table")
			}
			return
		}
	}
	t.Fatal("missing original trace")
}

type fireballArena struct {
	raw     []byte
	objects [3]*server.Object
	offsets [3]int
	arg     *server.SpellAcceptArg
	layout  [10]uintptr
}

func fireballArenaNew(t *testing.T, placement int) *fireballArena {
	t.Helper()
	raw, free := alloc.MakeHigh([]byte{}, 0x40000)
	t.Cleanup(free)
	v := &fireballArena{raw: raw, layout: fireballLayoutC()}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
		t.Fatal("foreign native owners must be above4GiB")
	}
	if unsafe.Sizeof(server.Object{}) > 0x8000 {
		t.Fatal("foreign object stride too small")
	}
	for i := 0; i < 3; i++ {
		v.offsets[i] = 256 + i*0x10000 + placement
		v.objects[i] = (*server.Object)(unsafe.Pointer(&raw[v.offsets[i]]))
	}
	v.arg = (*server.SpellAcceptArg)(unsafe.Pointer(&raw[0x3e000+placement]))
	return v
}

func (v *fireballArena) object(i int) *server.Object {
	if i < 0 {
		return nil
	}
	return v.objects[i]
}
func (v *fireballArena) index(p *server.Object) int {
	if p == nil {
		return -1
	}
	for i, obj := range v.objects {
		if obj == p {
			return i
		}
	}
	panic("unmapped native Fireball object")
}
func (v *fireballArena) state() []fireballObjectState {
	out := make([]fireballObjectState, 3)
	for i, obj := range v.objects {
		out[i] = fireballObjectState{Position: [2]uint32{math.Float32bits(obj.PosVec.X), math.Float32bits(obj.PosVec.Y)}, Velocity: [2]uint32{math.Float32bits(obj.VelVec.X), math.Float32bits(obj.VelVec.Y)}, Radius: math.Float32bits(obj.Shape.Circle.R), Speed: math.Float32bits(obj.SpeedCur), Directions: [2]uint16{uint16(obj.Direction1), uint16(obj.Direction2)}}
	}
	return out
}

func (v *fireballArena) set(state []fireballObjectState) {
	for i, s := range state {
		obj := v.objects[i]
		obj.PosVec = types.Pointf{X: math.Float32frombits(s.Position[0]), Y: math.Float32frombits(s.Position[1])}
		obj.VelVec = types.Pointf{X: math.Float32frombits(s.Velocity[0]), Y: math.Float32frombits(s.Velocity[1])}
		obj.Shape.Circle.R = math.Float32frombits(s.Radius)
		obj.SpeedCur = math.Float32frombits(s.Speed)
		obj.Direction1 = server.Dir16(s.Directions[0])
		obj.Direction2 = server.Dir16(s.Directions[1])
	}
}

func (v *fireballArena) expected(before []byte, state []fireballObjectState) []byte {
	out := bytes.Clone(before)
	for i, s := range state {
		a := v.offsets[i]
		for slot, value := range []uint32{s.Position[0], s.Position[1], s.Velocity[0], s.Velocity[1], s.Radius, s.Speed} {
			binary.LittleEndian.PutUint32(out[a+int(v.layout[slot+1]):], value)
		}
		binary.LittleEndian.PutUint16(out[a+int(v.layout[7]):], s.Directions[0])
		binary.LittleEndian.PutUint16(out[a+int(v.layout[8]):], s.Directions[1])
	}
	return out
}

func fireballJSONValue(value any) any {
	switch x := value.(type) {
	case int:
		return float64(x)
	case int32:
		return float64(x)
	case uint32:
		return float64(x)
	case []uint32:
		out := make([]any, len(x))
		for i, n := range x {
			out[i] = float64(n)
		}
		return out
	default:
		return value
	}
}

func (v *fireballArena) run(t *testing.T, route int, c fireballOriginalCase) {
	t.Helper()
	for i := range v.raw {
		v.raw[i] = 0xa5
	}
	v.set(c.Before)
	argBytes := unsafe.Slice((*byte)(unsafe.Pointer(v.arg)), 16)
	for i := range argBytes {
		argBytes[i] = byte((0xa000+i)*31 + 19)
	}
	before := bytes.Clone(v.raw)
	if !reflect.DeepEqual(v.state(), c.Before) {
		t.Fatal("native normalized input state")
	}
	var got []fireballOriginalEvent
	var mutations []resourceSpellMutation
	counts := map[string]int{}
	var traceRay *types.Rectf
	emit := func(kind string, args ...any) {
		for i, a := range args {
			args[i] = fireballJSONValue(a)
		}
		got = append(got, fireballOriginalEvent{Kind: kind, Args: args, Before: v.state()})
		at := counts[kind]
		counts[kind]++
		m := c.Input.Mutation
		if m.Stage != kind || m.At != at {
			return
		}
		mutations = append(mutations, m)
		switch m.Action {
		case "firing-state":
			firing := v.object(c.Input.Firing)
			firing.PosVec = types.Pointf{X: 37.25, Y: -26.75}
			firing.VelVec = types.Pointf{X: 0.7, Y: -1.1}
			firing.Shape.Circle.R = 3.1
			firing.Direction1 = 255
		case "projectile-speed":
			v.objects[1].SpeedCur = 1.3
		case "rectangle":
			if traceRay == nil {
				t.Fatal("rectangle mutation escaped trace")
			}
			*traceRay = types.Rectf{Min: types.Pointf{X: -17.25, Y: 26.5}, Max: types.Pointf{X: 50.25, Y: -40.75}}
		default:
			t.Fatal("unknown sealed Fireball mutation", m)
		}
	}
	base := fireballDefaultRuntime52C790()
	rt := server.SpellFireballRuntime52C790{
		NewObject: func(level int32) *server.Object {
			if level != c.Input.Level {
				t.Fatal("creation level slot", level, c.Input.Level)
			}
			emit("new", fireballTypeName52C790(level))
			if !c.Input.Allocation {
				return nil
			}
			return v.objects[1]
		},
		Direction: base.Direction,
		Trace: func(ray *types.Rectf) bool {
			traceRay = ray
			emit("trace", []uint32{math.Float32bits(ray.Min.X), math.Float32bits(ray.Min.Y), math.Float32bits(ray.Max.X), math.Float32bits(ray.Max.Y)}, 0, 0, 5)
			traceRay = nil
			return c.Input.Trace != 0
		},
		Place: func(projectile, owner *server.Object, pos types.Pointf) {
			emit("create", v.index(projectile), v.index(owner), math.Float32bits(pos.X), math.Float32bits(pos.Y), 0)
		},
		Balance: func(index int32) float64 {
			emit("balance", "FireballSpeedCoeff", index)
			return math.Float64frombits(c.Input.Balance)
		},
		Sound: func(id int32) int32 { emit("sound", id, 0); return int32(c.Input.Sound) },
		Audio: func(id int32, owner *server.Object) { emit("audio", id, v.index(owner), 0, 0) },
	}
	previous := fireballRuntimeForCall52C790
	fireballRuntimeForCall52C790 = func() server.SpellFireballRuntime52C790 { return rt }
	defer func() { fireballRuntimeForCall52C790 = previous }()
	result := fireballActualC(route, c.Input.Spell, v.objects[2], v.objects[2], v.object(c.Input.Firing), v.arg, c.Input.Level)
	if result != c.Result || !reflect.DeepEqual(got, c.Events) {
		t.Fatalf("actual C/native result=%d want%d ordered events got%+v want%+v", result, c.Result, got, c.Events)
	}
	if len(mutations) != len(c.Mutations) {
		t.Fatal("provider mutation count", mutations, c.Mutations)
	}
	for i, m := range mutations {
		if m != c.Mutations[i] {
			t.Fatal("ordered provider mutations", mutations, c.Mutations)
		}
	}
	if !reflect.DeepEqual(v.state(), c.After) || !bytes.Equal(v.raw, v.expected(before, c.After)) {
		t.Fatal("whole native objects/unused arguments/padding/canaries differ")
	}
}

func TestFireballNativeLayoutCABI(t *testing.T) {
	layout := fireballLayoutC()
	want := [10]uintptr{layout[0], unsafe.Offsetof(server.Object{}.PosVec), unsafe.Offsetof(server.Object{}.PosVec) + 4, unsafe.Offsetof(server.Object{}.VelVec), unsafe.Offsetof(server.Object{}.VelVec) + 4, unsafe.Offsetof(server.Object{}.Shape) + unsafe.Offsetof(server.Shape{}.Circle), unsafe.Offsetof(server.Object{}.SpeedCur), unsafe.Offsetof(server.Object{}.Direction1), unsafe.Offsetof(server.Object{}.Direction2), unsafe.Sizeof(uintptr(0))}
	if layout != want || layout[0] > unsafe.Sizeof(server.Object{}) {
		t.Fatal("independent C/Go object field layouts differ", layout, want, unsafe.Sizeof(server.Object{}))
	}
}

func TestFireballNativeStockTables(t *testing.T) {
	InitBlobData()
	if len(fireballOriginalDirections) != 2048 || fmt.Sprintf("%x", sha256.Sum256(fireballOriginalDirections)) != "fbfdbd07ab3c6fbec2eb3768e69f1c32eb63c77b6e80e9957b9567c78d61bfcf" {
		t.Fatal("sealed original direction table bytes changed")
	}
	actual := unsafe.Slice((*byte)(memmap.PtrOff(0x587000, 194136)), 2048)
	if !bytes.Equal(actual, fireballOriginalDirections) {
		t.Fatal("production direction table differs from unchanged original image")
	}
	direction := fireballDefaultRuntime52C790().Direction
	for i := 0; i < 256; i++ {
		v := direction(int16(i))
		if math.Float32bits(v.X) != binary.LittleEndian.Uint32(fireballOriginalDirections[i*8:]) || math.Float32bits(v.Y) != binary.LittleEndian.Uint32(fireballOriginalDirections[i*8+4:]) {
			t.Fatal("actual direction lookup differs from original", i, v)
		}
	}
}

func TestFireballNativeOriginalCABI(t *testing.T) {
	InitBlobData()
	rows := fireballCases(t)
	for _, c := range rows {
		fireballVerifyOriginal(t, c)
	}
	for route, name := range []string{"direct-public-C", "actual-shared-callback", "production-Go-wrapper"} {
		for _, placement := range []int{0, 8, 24} {
			t.Run(fmt.Sprintf("%s/placement-%d", name, placement), func(t *testing.T) {
				v := fireballArenaNew(t, placement)
				for i, c := range rows {
					t.Run(fmt.Sprintf("%03d-%s", i, c.Input.Name), func(t *testing.T) { v.run(t, route, c) })
				}
				runtime.GC()
			})
		}
	}
	t.Log("unchanged original Fireball cases=243, pre-port213case prefix retained, root=312bytes, original startup=319executedbytes, canonical control0xe3f; actual stock packed type slots and immutable direction table, three foreign placements, direct/shared-C/production-Go routes, ordered service mutations, whole object/unused argument canaries")
}
