package legacy

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"runtime/debug"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type simplePolypRawContract54CAE0 interface {
	PolypDeathNative54CB10(*server.PolypDeathState54CB10, server.PolypDeathRawRuntime54CB10)
}

type simplePolypPlan54CAE0 struct {
	Position, Flags []uint32
	Update          []int
	Cache, FPS      *uint32
	Result          *int64
	Reenter         []json.RawMessage
}

type simplePolypInput54CAE0 struct {
	Name, Entry         string
	Unit, Cloud, Update int
	Cache, FPS, Flags   uint32
	NameResult          uint32 `json:"name_result"`
	LifetimeBits        uint32 `json:"lifetime_bits"`
	Script              map[string]simplePolypPlan54CAE0
}

type simplePolypOutput54CAE0 struct {
	Name   string           `json:"name"`
	Result *uint32          `json:"result"`
	Fault  map[string]any   `json:"fault"`
	Events []map[string]any `json:"events"`
	Cache  uint32           `json:"cache"`
	FPS    uint32           `json:"fps"`
	Heap   string           `json:"heap_sha256"`
}

func TestSimplePolypDeathOriginal54CAE0(t *testing.T) {
	TestSimplePolypDeathNativeLayout54CAE0(t)
	var inputs []simplePolypInput54CAE0
	trollDeathFixture54A270(t, "simple_polyp_death_independent_inputs_54cae0.json.gz", "54334f7371c06bc3f046293acc31e87004e286bb7293600b3d1819245d705d26", "b1875ee4b7a1e356ddfa8dd915919278c2e3597565b315a453c0c8713c19efc5", &inputs)
	if len(inputs) != 149 || inputs[0].NameResult != 73 || inputs[0].LifetimeBits != 0x3e800000 {
		t.Fatal("complete independent input count and decoded scalar transport")
	}
	if _, ok := any((*server.Object)(nil)).(simplePolypRawContract54CAE0); !ok {
		if os.Getenv("OPENNOX_SIMPLE_POLYP_RAW_REQUIRED") == "1" {
			t.Fatal("raw native Polyp contract not implemented")
		}
		t.Log("frozen original references compile before engine; no native leaf comparison credited")
		return
	}
	for _, profile := range []struct {
		name, encoded, decoded string
		nearest                bool
	}{
		{"tozero", "0cc6f398fbbf2b9a3671eb2071ecf3066dd9dfbddd2000322999a3f772af46eb", "8f839d3732dd1ca957b97845f9841ce290742f495b7a8c39737a901c4db451e8", false},
		{"nearest", "14427ed8cd5257fd981a0b3a191a036cd9eb50182e71422d986614e64855d8e0", "1ac2b826570efce639c37228620946ad6606cbf9fda4dea09fafc00e0b694f11", true},
	} {
		t.Run(profile.name, func(t *testing.T) {
			var outputs []simplePolypOutput54CAE0
			trollDeathFixture54A270(t, "simple_polyp_death_native_original_54cae0_pc53_"+profile.name+".json.gz", profile.encoded, profile.decoded, &outputs)
			if len(outputs) != len(inputs) {
				t.Fatal("complete original output count")
			}
			for _, mode := range []string{"Go", "C"} {
				t.Run(mode, func(t *testing.T) {
					for i, input := range inputs {
						t.Run(input.Name, func(t *testing.T) {
							if outputs[i].Name != input.Name {
								t.Fatal("frozen input/output identity")
							}
							got := simplePolypRunOriginal54CAE0(t, input, mode, profile.nearest)
							a, err := json.Marshal(got)
							if err != nil {
								t.Fatal(err)
							}
							b, err := json.Marshal(outputs[i])
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

func simplePolypRunOriginal54CAE0(t *testing.T, in simplePolypInput54CAE0, mode string, nearest bool) (out simplePolypOutput54CAE0) {
	t.Helper()
	base, page, free := simplePolypOriginalHeap54CAE0()
	if base == nil {
		t.Fatal("guarded native C fixture")
	}
	defer free()
	if page <= unsafe.Sizeof(server.Object{}) || unsafe.Sizeof(uintptr(0)) == 8 && uintptr(base) <= math.MaxUint32 {
		t.Fatal("native high-address complete object/guard domain")
	}
	buf := unsafe.Slice((*byte)(base), 0x10000)
	for i := range buf {
		buf[i] = byte(i*17 + (i>>7)*29)
	}
	canonical := bytes.Clone(buf)
	offsets := [7]int{0x1000, 0x2000, 0x3000, 0x4000, 0x5000, 0x6000, 0x7000}
	ptr := func(role int) unsafe.Pointer {
		if role < 0 {
			return simplePolypOriginalBad54CAE0(base, page, role)
		}
		return unsafe.Add(base, offsets[role])
	}
	role := func(p unsafe.Pointer) int {
		for n := -4; n < 7; n++ {
			if p == ptr(n) {
				return n
			}
		}
		t.Fatalf("undeclared native pointer %p", p)
		return -99
	}
	originalPointer := func(n int) uint32 {
		switch n {
		case -4:
			return 0x21000000 - 16
		case -3:
			return 0x21000000 - 60
		case -2:
			return 0x41000000
		case -1:
			return 0
		default:
			return 0x21000000 + uint32(offsets[n])
		}
	}
	o := server.Object{}
	flagOffset, xOffset, yOffset, updateOffset := int(unsafe.Offsetof(o.ObjFlags)), int(unsafe.Offsetof(o.PosVec)), int(unsafe.Offsetof(o.PosVec))+4, int(unsafe.Offsetof(o.UpdateData))
	put := func(b []byte, at int, value uint32) { binary.LittleEndian.PutUint32(b[at:], value) }
	putPointer := func(at int, value unsafe.Pointer) {
		if unsafe.Sizeof(uintptr(0)) == 4 {
			put(buf, at, uint32(uintptr(value)))
		} else {
			binary.LittleEndian.PutUint64(buf[at:], uint64(uintptr(value)))
		}
	}
	for n := range 4 {
		at := offsets[n]
		put(canonical, at+16, in.Flags+uint32(n))
		put(canonical, at+56, 0x40e00000+uint32(n))
		put(canonical, at+60, 0x41000000+uint32(n))
		put(canonical, at+748, originalPointer(in.Update))
	}
	for n := 4; n < 7; n++ {
		put(canonical, offsets[n], 0xdead0000+uint32(n))
	}
	copy(buf, canonical)
	for n := range 4 {
		at := offsets[n]
		put(buf, at+flagOffset, in.Flags+uint32(n))
		put(buf, at+xOffset, 0x40e00000+uint32(n))
		put(buf, at+yOffset, 0x41000000+uint32(n))
		putPointer(at+updateOffset, ptr(in.Update))
	}
	// In original role -4, flags alias heap+0 and source X/Y alias +40/+44.
	// Native field growth moves the latter two reads; preserve their declared
	// original values without treating unrelated native padding as PE memory.
	copy(buf[xOffset-flagOffset:xOffset-flagOffset+4], canonical[40:44])
	copy(buf[yOffset-flagOffset:yOffset-flagOffset+4], canonical[44:48])
	expected := bytes.Clone(buf)
	type guardedState struct {
		Before [16]byte
		State  server.PolypDeathState54CB10
		After  [16]byte
	}
	g, freeState := alloc.New(guardedState{})
	defer freeState()
	*g = guardedState{Before: [16]byte{0x12, 0x34, 0x56}, State: server.PolypDeathState54CB10{TypeIndex: in.Cache}, After: [16]byte{0x87, 0x65, 0x43}}
	initialState := *g
	fps := in.FPS
	events := []map[string]any{}
	occurrences := map[string]int{}
	nextIndex := 0
	type invocation struct {
		entry string
		unit  int
		phase string
		cloud *server.Object
	}
	var active, failed *invocation
	var invoke func(string, int) *uint32
	mutate := func(plan simplePolypPlan54CAE0) {
		for _, field := range []struct {
			values              []uint32
			nativeOff, original int
		}{{plan.Flags, flagOffset, 16}, {plan.Position, xOffset, 56}} {
			if len(field.values) == 0 {
				continue
			}
			n := int(field.values[0])
			for i, value := range field.values[1:] {
				at := offsets[n] + field.nativeOff + 4*i
				put(buf, at, value)
				put(expected, at, value)
				put(canonical, offsets[n]+field.original+4*i, value)
			}
		}
		if len(plan.Update) != 0 {
			n, value := plan.Update[0], plan.Update[1]
			at := offsets[n] + updateOffset
			putPointer(at, ptr(value))
			copy(expected[at:at+int(unsafe.Sizeof(uintptr(0)))], buf[at:at+int(unsafe.Sizeof(uintptr(0)))])
			put(canonical, offsets[n]+748, originalPointer(value))
		}
		if plan.Cache != nil {
			g.State.TypeIndex = *plan.Cache
		}
		if plan.FPS != nil {
			fps = *plan.FPS
		}
	}
	begin := func(kind string) (map[string]any, simplePolypPlan54CAE0) {
		sequence := occurrences[kind]
		occurrences[kind]++
		index := nextIndex
		nextIndex++
		return map[string]any{"kind": kind, "index": index, "cache": g.State.TypeIndex, "fps": fps}, in.Script[fmt.Sprintf("%s:%d", kind, sequence)]
	}
	finish := func(event map[string]any, plan simplePolypPlan54CAE0, result any) {
		events = append(events, event)
		mutate(plan)
		if len(plan.Reenter) != 0 {
			var entry string
			var unit int
			if len(plan.Reenter) != 2 || json.Unmarshal(plan.Reenter[0], &entry) != nil || json.Unmarshal(plan.Reenter[1], &unit) != nil {
				t.Fatal("declared reentry input")
			}
			events = append(events, map[string]any{"kind": "nested", "index": event["index"], "entry": entry, "unit": unit})
			r := invoke(entry, unit)
			events = append(events, map[string]any{"kind": "nested-return", "index": event["index"], "result": r})
		}
		events = append(events, map[string]any{"kind": event["kind"].(string) + "-return", "index": event["index"], "result": result})
	}
	simple := server.SimpleObjectDeathRuntime54CAE0{
		Audio: func(id uint32, source *server.Object) {
			e, plan := begin("audio")
			e["sound"], e["unit"] = id, role(unsafe.Pointer(source))
			finish(e, plan, nil)
		},
		DelayedDelete: func(source *server.Object) {
			e, plan := begin("delete")
			e["unit"] = role(unsafe.Pointer(source))
			finish(e, plan, nil)
		},
	}
	rt := server.PolypDeathRawRuntime54CB10{SimpleObjectDeathRuntime54CAE0: simple, RoundToNearest: nearest}
	rt.Resolve = func() uint32 {
		e, plan := begin("name")
		e["name"] = "ToxicCloud"
		value := in.NameResult
		if plan.Result != nil {
			value = uint32(*plan.Result)
		}
		finish(e, plan, value)
		return value
	}
	rt.New = func(id uint32) *server.Object {
		e, plan := begin("new")
		e["type_id"] = id
		value := in.Cloud
		if plan.Result != nil {
			value = int(*plan.Result)
		}
		cloud := (*server.Object)(ptr(value))
		finish(e, plan, value)
		active.cloud = cloud
		return cloud
	}
	rt.Create = func(cloud, owner *server.Object, position types.Pointf, extra uint32) {
		e, plan := begin("create")
		e["object"], e["owner"] = role(unsafe.Pointer(cloud)), role(unsafe.Pointer(owner))
		e["position"], e["extra"] = [2]uint32{math.Float32bits(position.X), math.Float32bits(position.Y)}, extra
		active.phase = "duration"
		finish(e, plan, nil)
	}
	rt.Balance = func() float64 {
		e, plan := begin("lifetime")
		e["name"] = "ToxicCloudLifetime"
		value := in.LifetimeBits
		if plan.Result != nil {
			value = uint32(*plan.Result)
		}
		finish(e, plan, value)
		return float64(math.Float32frombits(value))
	}
	rt.FPS = func() uint32 { return fps }
	oldSimple, oldRT, oldState := simplePolypOriginalSimpleRuntimeOverride54CAE0, simplePolypOriginalRuntimeOverride54CB10, simplePolypOriginalStateOverride54CB10
	simplePolypOriginalSimpleRuntimeOverride54CAE0, simplePolypOriginalRuntimeOverride54CB10, simplePolypOriginalStateOverride54CB10 = &simple, &rt, &g.State
	defer func() {
		simplePolypOriginalSimpleRuntimeOverride54CAE0, simplePolypOriginalRuntimeOverride54CB10, simplePolypOriginalStateOverride54CB10 = oldSimple, oldRT, oldState
	}()
	invoke = func(entry string, unit int) (result *uint32) {
		previous := active
		context := &invocation{entry: entry, unit: unit, phase: "source-y"}
		if entry == "54cae0" {
			context.phase = "flags"
		}
		active = context
		returned := false
		defer func() {
			if !returned && failed == nil {
				failed = context
			}
			active = previous
		}()
		index := map[string]int{"54cae0": 0, "54cb10": 1, "54cbb0": 2}[entry]
		source := (*server.Object)(ptr(unit))
		if mode == "C" {
			if simplePolypOriginalABI54CAE0(index) != 64 {
				t.Fatal("actual typed native public C entry required")
			}
			value := simplePolypOriginalC54CAE0(index, source)
			if entry == "54cae0" {
				result = &value
			}
		} else {
			switch entry {
			case "54cae0":
				value := server.ImpEggDieNative54CAE0(source, simple)
				result = &value
			case "54cb10":
				any(source).(simplePolypRawContract54CAE0).PolypDeathNative54CB10(&g.State, rt)
			case "54cbb0":
				server.PotionDieNative54CBB0(source, simple)
			default:
				t.Fatal("unknown original entry", entry)
			}
		}
		returned = true
		return
	}
	oldFault := debug.SetPanicOnFault(true)
	defer debug.SetPanicOnFault(oldFault)
	var panicValue any
	func() {
		defer func() { panicValue = recover() }()
		out.Result = invoke(in.Entry, in.Unit)
	}()
	if panicValue != nil {
		text := fmt.Sprint(panicValue)
		if failed == nil || !strings.Contains(text, "invalid memory") && !strings.Contains(text, "fault") {
			t.Fatal("not an actual native memory fault", panicValue)
		}
		// Resolve actual nonnil fault addresses to independent native C/Go
		// fields. Go canonicalizes nil faults; only those use the active
		// entry's observed world-call phase. No expected output is consulted.
		site := failed.phase
		if memory, ok := panicValue.(interface{ Addr() uintptr }); ok && memory.Addr() != 0 {
			address := memory.Addr()
			source := uintptr(ptr(failed.unit))
			cloud := uintptr(unsafe.Pointer(failed.cloud))
			switch address {
			case source + uintptr(flagOffset):
				site = "flags"
			case source + uintptr(yOffset):
				site = "source-y"
			case source + uintptr(xOffset):
				site = "source-x"
			case cloud + uintptr(updateOffset):
				site = "update"
			case uintptr(ptr(-2)):
				site = "duration"
			default:
				t.Fatalf("unclassified actual native memory fault at %x: %v", address, panicValue)
			}
		}
		out.Fault = map[string]any{"kind": "memory", "site": site}
	}
	// Only the original flag and duration destinations are projected back.
	// Callback pointer/position mutations have their own exact expected bytes.
	for n := range 7 {
		at := offsets[n]
		copy(expected[at:at+4], buf[at:at+4])
		copy(canonical[at:at+4], buf[at:at+4])
		if n < 4 {
			copy(expected[at+flagOffset:at+flagOffset+4], buf[at+flagOffset:at+flagOffset+4])
			copy(canonical[at+16:at+20], buf[at+flagOffset:at+flagOffset+4])
		}
	}
	copy(expected[:4], buf[:4])
	copy(canonical[:4], buf[:4])
	if !bytes.Equal(buf, expected) {
		t.Fatal("whole64KiB native frame changed outside original outputs and declared callbacks")
	}
	if g.Before != initialState.Before || g.After != initialState.After {
		t.Fatal("complete native type-cache guards")
	}
	out.Name, out.Events, out.Cache, out.FPS = in.Name, events, g.State.TypeIndex, fps
	out.Heap = fmt.Sprintf("%x", sha256.Sum256(canonical))
	return
}
