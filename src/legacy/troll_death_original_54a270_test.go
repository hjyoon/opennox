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
	"runtime"
	"runtime/debug"
	"strconv"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type trollDeathRawContract54A270 interface {
	TrollDeathNative54A270(*server.TrollDeathState54A270, server.TrollDeathRuntime54A270) int32
}

type trollDeathInput54A270 struct {
	Name                 string
	Unit, New            int
	Cache, Resolved, FPS uint32
	Lifetime             uint64
	Positions            [4][2]uint32
	Updates              [4]int
	Script               map[string]struct {
		Position   []uint32
		Update     []int
		FPS, Cache *uint32
		Reenter    *int
	}
}

type trollDeathOutput54A270 struct {
	Name      string           `json:"name"`
	Result    *uint32          `json:"result"`
	Fault     map[string]any   `json:"fault"`
	Events    []map[string]any `json:"events"`
	Writes    []map[string]any `json:"writes"`
	Cache     uint32           `json:"cache"`
	FPS       uint32           `json:"fps"`
	Durations [2]uint32        `json:"durations"`
	Positions [4][2]uint32     `json:"positions"`
	Updates   [4]int           `json:"updates"`
	Heap      string           `json:"heap_sha256"`
}

func trollDeathFixture54A270(t *testing.T, name, encodedSHA, decodedSHA string, target any) {
	t.Helper()
	raw, err := os.ReadFile("../server/testdata/" + name)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != encodedSHA {
		t.Fatal("frozen compressed original", name, err)
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
	if fmt.Sprintf("%x", sha256.Sum256(decoded)) != decodedSHA {
		t.Fatal("frozen decoded original", name)
	}
	d := json.NewDecoder(bytes.NewReader(decoded))
	d.UseNumber()
	if err = d.Decode(target); err != nil {
		t.Fatal(err)
	}
}

func TestTrollDeathOriginal54A270(t *testing.T) {
	TestTrollDeathNativeLayout54A270(t)
	if _, ok := any((*server.Object)(nil)).(trollDeathRawContract54A270); !ok {
		if os.Getenv("OPENNOX_TROLL_RAW_REQUIRED") == "1" {
			t.Fatal("raw native contract is not implemented")
		}
		t.Log("raw native contract not yet implemented; references compile before engine")
		return
	}
	var inputs []trollDeathInput54A270
	trollDeathFixture54A270(t, "troll_death_independent_inputs_54a270.json.gz", "203cc04c74a1a0aa07dbb046c1fcf83ce685657e134fffdc03068b80883c42b2", "1537bdb164c257202c9c00b365cc82080434179818036b3492460462dd56c5a6", &inputs)
	if len(inputs) != 136 {
		t.Fatal("complete independent input count")
	}
	for _, profile := range []struct {
		name, encoded, decoded string
		nearest                bool
	}{
		{"pc53_tozero", "410eb49fb7016edd692668b5127a2f855c2c3bbf9de6c89193eae3ff985dd5b6", "3e93259e7cab1fd4ce49ca281aee27f005b7d2659de564881a7ffbfc20479f54", false},
		{"pc53_nearest", "cfb93ed618528730b40b5d40e370d94aae409c48addb066530a47201afeaa327", "388a6b0710dc68b1a0753e633bf30a047d0bc26847724ba0e3feacb8ad6a1b43", true},
	} {
		t.Run(profile.name, func(t *testing.T) {
			var outputs []trollDeathOutput54A270
			trollDeathFixture54A270(t, "troll_death_native_original_54a270_"+profile.name+".json.gz", profile.encoded, profile.decoded, &outputs)
			if len(outputs) != len(inputs) {
				t.Fatal("complete output count")
			}
			for _, mode := range []string{"Go", "C"} {
				t.Run(mode, func(t *testing.T) {
					for i, in := range inputs {
						t.Run(in.Name, func(t *testing.T) {
							want := outputs[i]
							if want.Name != in.Name {
								t.Fatal("input/output identity")
							}
							got := trollDeathRunOriginal54A270(t, in, mode, profile.nearest)
							a, err := json.Marshal(got)
							if err != nil {
								t.Fatal(err)
							}
							b, err := json.Marshal(want)
							if err != nil {
								t.Fatal(err)
							}
							if !bytes.Equal(a, b) {
								t.Fatalf("complete original output differs\ngot %s\nwant%s", a, b)
							}
						})
					}
				})
			}
		})
	}
}

func trollDeathRunOriginal54A270(t *testing.T, in trollDeathInput54A270, mode string, nearest bool) (out trollDeathOutput54A270) {
	t.Helper()
	base, free := alloc.MallocHigh(0x10000)
	defer free()
	buf := unsafe.Slice((*byte)(base), 0x10000)
	canonical := make([]byte, len(buf))
	expected := make([]byte, len(buf))
	for i := range buf {
		buf[i] = byte(i*17 + (i>>7)*29)
	}
	copy(canonical, buf)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(base) <= math.MaxUint32 {
		t.Fatal("native buffer must exceed4GiB")
	}
	offsets := [6]int{0x1000, 0x2000, 0x3000, 0x4000, 0x6000, 0x7000}
	var objects [4]*server.Object
	for i := range objects {
		objects[i] = (*server.Object)(unsafe.Add(base, offsets[i]))
	}
	ptr := func(role int) unsafe.Pointer {
		if role < 0 {
			return trollDeathOriginalBad54A270(role)
		}
		return unsafe.Add(base, offsets[role])
	}
	role := func(p unsafe.Pointer) int {
		for i := -2; i < 6; i++ {
			if ptr(i) == p {
				return i
			}
		}
		t.Fatalf("undeclared pointer %p", p)
		return -99
	}
	pePtr := func(n int) uint32 {
		if n == -1 {
			return 0
		}
		if n == -2 {
			return 0x41000000
		}
		return 0x21000000 + uint32(offsets[n])
	}
	storeUpdate := func(i int, p unsafe.Pointer) {
		at := offsets[i] + int(unsafe.Offsetof(server.Object{}.UpdateData))
		// These are C-owned records with arbitrary guard bytes. A typed
		// pointer assignment would feed the old poison word to Go's write
		// barrier; on 386 it can resemble an invalid Go heap pointer.
		// Store the same native pointer bits without tracing the old bytes.
		if unsafe.Sizeof(uintptr(0)) == 4 {
			binary.LittleEndian.PutUint32(buf[at:], uint32(uintptr(p)))
		} else {
			binary.LittleEndian.PutUint64(buf[at:], uint64(uintptr(p)))
		}
	}
	for i, o := range objects {
		o.PosVec = types.Ptf(math.Float32frombits(in.Positions[i][0]), math.Float32frombits(in.Positions[i][1]))
		storeUpdate(i, ptr(in.Updates[i]))
		binary.LittleEndian.PutUint32(canonical[offsets[i]+56:], in.Positions[i][0])
		binary.LittleEndian.PutUint32(canonical[offsets[i]+60:], in.Positions[i][1])
		binary.LittleEndian.PutUint32(canonical[offsets[i]+748:], pePtr(in.Updates[i]))
	}
	for i, v := range []uint32{0xffffff85, 0xfffffe38} {
		binary.LittleEndian.PutUint32(buf[offsets[i+4]:], v)
		binary.LittleEndian.PutUint32(canonical[offsets[i+4]:], v)
	}
	copy(expected, buf)
	stateGuard, freeState := alloc.NewHigh(struct {
		Before [16]byte
		State  server.TrollDeathState54A270
		After  [16]byte
	}{})
	defer freeState()
	for i := range stateGuard.Before {
		stateGuard.Before[i] = 0x5a
		stateGuard.After[i] = 0xa5
	}
	stateGuard.State.TypeIndex = in.Cache
	state := &stateGuard.State
	beforeGuard := *stateGuard
	fps := in.FPS
	out.Name = in.Name
	out.Events = []map[string]any{}
	out.Writes = []map[string]any{}
	type frame struct {
		unit, newRole, dataRole int
		allocated               bool
	}
	var stack []*frame
	var faultFrame *frame
	var pending []uint32
	serviceIndex := 0
	var rt server.TrollDeathRuntime54A270
	var call func(int) (uint32, any)
	mutate := func(planKey string) {
		plan := in.Script[planKey]
		if len(plan.Position) != 0 {
			i := int(plan.Position[0])
			objects[i].PosVec = types.Ptf(math.Float32frombits(plan.Position[1]), math.Float32frombits(plan.Position[2]))
			at := offsets[i] + int(unsafe.Offsetof(server.Object{}.PosVec))
			copy(expected[at:at+8], buf[at:at+8])
		}
		if len(plan.Update) != 0 {
			i := plan.Update[0]
			storeUpdate(i, ptr(plan.Update[1]))
			at := offsets[i] + int(unsafe.Offsetof(server.Object{}.UpdateData))
			copy(expected[at:at+int(unsafe.Sizeof(uintptr(0)))], buf[at:at+int(unsafe.Sizeof(uintptr(0)))])
		}
		if plan.FPS != nil {
			fps = *plan.FPS
		}
		if plan.Cache != nil {
			state.TypeIndex = *plan.Cache
		}
	}
	service := func(kind string, fields map[string]any, result uint64) {
		at := serviceIndex
		serviceIndex++
		key := strconv.Itoa(at)
		e := map[string]any{"kind": kind, "index": at, "cache": state.TypeIndex, "fps": fps}
		for k, v := range fields {
			e[k] = v
		}
		out.Events = append(out.Events, e)
		mutate(key)
		if nested := in.Script[key].Reenter; nested != nil {
			out.Events = append(out.Events, map[string]any{"kind": "nested", "index": at, "unit": *nested})
			value, panicValue := call(*nested)
			if panicValue != nil {
				panic(panicValue)
			}
			if value != 1 {
				t.Fatal("nested return")
			}
			out.Events = append(out.Events, map[string]any{"kind": "nested-return", "index": at, "result": uint32(1)})
		}
		out.Events = append(out.Events, map[string]any{"kind": kind + "-return", "index": at, "result": result})
	}
	rt = server.TrollDeathRuntime54A270{
		RoundToNearest: nearest,
		Resolve: func() uint32 {
			service("resolve", map[string]any{"type": "SmallToxicCloud"}, uint64(in.Resolved))
			pending = append(pending, in.Resolved)
			return in.Resolved
		},
		New: func(id uint32) *server.Object {
			if len(pending) != 0 {
				v := pending[len(pending)-1]
				pending = pending[:len(pending)-1]
				if state.TypeIndex != v {
					t.Fatal("original cached type store before constructor", state.TypeIndex, v)
				}
				out.Writes = append(out.Writes, map[string]any{"pc": "0x54a289", "role": nil, "cache": true, "value": v})
			}
			f := stack[len(stack)-1]
			service("new", map[string]any{"type_index": id}, uint64(pePtr(in.New)))
			f.newRole = in.New
			f.allocated = in.New != -1
			if in.New >= 0 {
				f.dataRole = role(objects[in.New].UpdateData)
			}
			return (*server.Object)(ptr(in.New))
		},
		Create: func(o, owner *server.Object, pos types.Pointf, extra uint32) uint32 {
			at := uint32(serviceIndex)
			service("create", map[string]any{"object": role(unsafe.Pointer(o)), "owner": role(unsafe.Pointer(owner)), "position": [2]uint32{math.Float32bits(pos.X), math.Float32bits(pos.Y)}, "trailing_zero": extra}, uint64(0xa5a50000+at))
			return uint32(0xa5a50000 + at)
		},
		Audio: func(id uint32, o *server.Object, kind, arg uint32) uint32 {
			at := uint32(serviceIndex)
			service("audio", map[string]any{"sound": id, "unit": role(unsafe.Pointer(o)), "kind_arg": kind, "arg": arg}, uint64(0xa5a50000+at))
			return uint32(0xa5a50000 + at)
		},
		Balance: func() float64 {
			service("balance", map[string]any{"key": "SmallToxicCloudLifetime"}, in.Lifetime)
			return math.Float64frombits(in.Lifetime)
		},
		FPS: func() uint32 { return fps },
	}
	oldRT, oldState := trollDeathOriginalRuntimeOverride54A270, trollDeathOriginalStateOverride54A270
	trollDeathOriginalRuntimeOverride54A270, trollDeathOriginalStateOverride54A270 = &rt, state
	defer func() {
		trollDeathOriginalRuntimeOverride54A270, trollDeathOriginalStateOverride54A270 = oldRT, oldState
	}()
	oldPanic := debug.SetPanicOnFault(true)
	defer debug.SetPanicOnFault(oldPanic)
	call = func(unit int) (value uint32, panicValue any) {
		f := &frame{unit: unit, newRole: -1, dataRole: -1}
		stack = append(stack, f)
		defer func() {
			panicValue = recover()
			if panicValue != nil {
				faultFrame = f
			}
			stack = stack[:len(stack)-1]
			if panicValue == nil && f.allocated {
				if f.dataRole < 4 {
					t.Fatal("returned despite invalid cached data", f.dataRole)
				}
				at := offsets[f.dataRole]
				v := binary.LittleEndian.Uint32(buf[at:])
				copy(expected[at:at+4], buf[at:at+4])
				out.Writes = append(out.Writes, map[string]any{"pc": "0x54a2f4", "role": f.dataRole, "cache": false, "value": v})
			}
		}()
		if mode == "Go" {
			value = uint32(any((*server.Object)(ptr(unit))).(trollDeathRawContract54A270).TrollDeathNative54A270(state, rt))
		} else {
			value = uint32(trollDeathOriginalC54A270((*server.Object)(ptr(unit))))
		}
		return
	}
	value, panicValue := call(in.Unit)
	if panicValue == nil {
		out.Result = &value
	} else {
		if _, ok := panicValue.(runtime.Error); !ok {
			t.Fatalf("fault is not actual native Go runtime access: %T %v", panicValue, panicValue)
		}
		pc := "0x54a2f4"
		address := uint32(0)
		access := 20
		if faultFrame == nil {
			t.Fatal("missing actual fault frame")
		}
		if faultFrame.newRole == -2 {
			pc = "0x54a2a1"
			address = pePtr(faultFrame.newRole) + 748
			access = 19
		} else if faultFrame.unit < 0 {
			pc = "0x54a2a7"
			address = pePtr(faultFrame.unit) + 60
			access = 19
		} else {
			address = pePtr(faultFrame.dataRole)
		}
		out.Fault = map[string]any{"kind": "original-memory-access", "pc": pc, "access": access, "address": address, "size": 4}
	}
	if !bytes.Equal(expected, buf) {
		for i := range buf {
			if expected[i] != buf[i] {
				t.Fatalf("complete64KiB native guard differs at%04x got%02x want%02x", i, buf[i], expected[i])
			}
		}
	}
	beforeGuard.State.TypeIndex = state.TypeIndex
	if !reflect.DeepEqual(beforeGuard, *stateGuard) {
		t.Fatal("whole guarded cache record")
	}
	out.Cache, out.FPS = state.TypeIndex, fps
	for i, o := range objects {
		out.Positions[i] = [2]uint32{math.Float32bits(o.PosVec.X), math.Float32bits(o.PosVec.Y)}
		out.Updates[i] = role(o.UpdateData)
		binary.LittleEndian.PutUint32(canonical[offsets[i]+56:], out.Positions[i][0])
		binary.LittleEndian.PutUint32(canonical[offsets[i]+60:], out.Positions[i][1])
		binary.LittleEndian.PutUint32(canonical[offsets[i]+748:], pePtr(out.Updates[i]))
	}
	for i := range out.Durations {
		out.Durations[i] = binary.LittleEndian.Uint32(buf[offsets[i+4]:])
		binary.LittleEndian.PutUint32(canonical[offsets[i+4]:], out.Durations[i])
	}
	out.Heap = fmt.Sprintf("%x", sha256.Sum256(canonical))
	return
}
