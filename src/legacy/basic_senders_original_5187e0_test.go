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
	"path/filepath"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

type basicSenderOutput5187E0 struct {
	Events      []map[string]any `json:"events"`
	Reads       [][]any          `json:"global_frame_reads"`
	State       map[string]any   `json:"state"`
	Return      uint32           `json:"return_word"`
	Heap        bool             `json:"all_other_heap_bytes_preserved"`
	Stack       bool             `json:"stack_guards_preserved"`
	CW          bool             `json:"CW_and_TOP_preserved"`
	Nonvolatile bool             `json:"nonvolatile_preserved"`
	Fault       []map[string]any `json:"fault"`
}
type basicSenderRow5187E0 struct {
	Name   string                    `json:"name"`
	Input  map[string]any            `json:"input"`
	Output []basicSenderOutput5187E0 `json:"output"`
}

func basicSenderNumber5187E0(v any) uint32      { return uint32(int64(v.(float64))) }
func basicSenderMap5187E0(v any) map[string]any { return v.(map[string]any) }
func basicSenderLoad5187E0(t *testing.T, nearest bool) ([]basicSenderRow5187E0, []basicSenderRow5187E0) {
	t.Helper()
	name, compressed, decoded := "basic_senders_original_startup_5187e0.json.gz", "c0d1b43e0d6db61ceeae3c21a4f16f7dfe5d48e8cfae346534a883a087facd4c", "49886052a67a6c3c8abd68159ddfd3d2437b072be32bb3ecd769f7c08888068f"
	if nearest {
		name, compressed, decoded = "basic_senders_original_nearest_5187e0.json.gz", "58ba3d577703db411e4089ef0f8376093c8b67c0eeba7d16abd6191598658a67", "4c5ffe91906c5dfea4eccd056d0d6c863fcda3c1ee2c9695cf9f48a15443c759"
	}
	raw, err := os.ReadFile(filepath.Join("..", "server", "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != compressed {
		t.Fatal("original compressed seal")
	}
	z, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	blob, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err = z.Close(); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(blob)) != decoded {
		t.Fatal("original decoded seal")
	}
	var data struct {
		Contexts []basicSenderRow5187E0 `json:"contexts"`
		Faults   []basicSenderRow5187E0 `json:"expected_fault_contexts"`
	}
	if err = json.Unmarshal(blob, &data); err != nil {
		t.Fatal(err)
	}
	if len(data.Contexts) != 3443 || len(data.Faults) != 4 {
		t.Fatal("original dimensions")
	}
	return data.Contexts, data.Faults
}

type basicSenderGuard5187E0[T comparable] struct {
	Before [32]byte
	Value  T
	After  [32]byte
}
type basicSenderMemory5187E0 struct {
	Pointer  unsafe.Pointer
	Size     uintptr
	Expected []byte
}
type basicSenderFixture5187E0 struct {
	t       *testing.T
	s       *server.Server
	unit    *server.Object
	updates [2]*server.MonsterGenUpdateData
	health  [2]*server.HealthData
	input   map[string]any
	nearest bool
	want    *basicSenderOutput5187E0
	events  []map[string]any
	frames  []uint32
	memory  []basicSenderMemory5187E0
}

func basicSenderBytes5187E0(p unsafe.Pointer, n uintptr) []byte {
	return append([]byte(nil), unsafe.Slice((*byte)(p), int(n))...)
}
func basicSenderAllocate5187E0[T comparable](f *basicSenderFixture5187E0) *T {
	f.t.Helper()
	g, free := alloc.NewHigh(basicSenderGuard5187E0[T]{})
	f.t.Cleanup(free)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(g)) <= math.MaxUint32 {
		f.t.Fatal("native guarded records must exceed4GiB")
	}
	for i := range g.Before {
		g.Before[i] = 0x5a
		g.After[i] = 0xa5
	}
	f.memory = append(f.memory, basicSenderMemory5187E0{Pointer: unsafe.Pointer(g), Size: unsafe.Sizeof(*g)})
	return &g.Value
}
func (f *basicSenderFixture5187E0) pointer(v any) unsafe.Pointer {
	if v == nil {
		return nil
	}
	switch v.(string) {
	case "unit":
		return unsafe.Pointer(f.unit)
	case "update":
		return unsafe.Pointer(f.updates[0])
	case "other_update":
		return unsafe.Pointer(f.updates[1])
	case "health":
		return unsafe.Pointer(f.health[0])
	case "other_health":
		return unsafe.Pointer(f.health[1])
	default:
		f.t.Fatal("unknown original pointer role", v)
		return nil
	}
}
func newBasicSenderFixture5187E0(t *testing.T, s *server.Server, input map[string]any, nearest bool) *basicSenderFixture5187E0 {
	f := &basicSenderFixture5187E0{t: t, s: s, input: input, nearest: nearest}
	f.unit = basicSenderAllocate5187E0[server.Object](f)
	for i := range f.updates {
		f.updates[i] = basicSenderAllocate5187E0[server.MonsterGenUpdateData](f)
		f.health[i] = basicSenderAllocate5187E0[server.HealthData](f)
	}
	for _, field := range []string{"unit_class", "subclass", "type_index", "net_code", "direction", "damage_frame", "health_link", "update_link", "current", "other_current", "frame"} {
		f.change(field, input[field])
	}
	pos := input["position_bits"].([]any)
	f.unit.PosVec.X = math.Float32frombits(basicSenderNumber5187E0(pos[0]))
	f.unit.PosVec.Y = math.Float32frombits(basicSenderNumber5187E0(pos[1]))
	index := basicSenderNumber5187E0(input["player"])
	for k, u := range f.updates {
		for i := range u.HealthSamples {
			u.HealthSamples[i] = uint16(0x2200 + k*0x1100 + i)
		}
		if index < 32 {
			u.HealthSamples[index] = uint16(basicSenderNumber5187E0(input["sample"]))
		}
	}
	for i := range f.memory {
		m := &f.memory[i]
		m.Expected = basicSenderBytes5187E0(m.Pointer, m.Size)
	}
	return f
}
func (f *basicSenderFixture5187E0) change(name string, value any) {
	switch name {
	case "frame":
		f.s.SetFrame(basicSenderNumber5187E0(value))
	case "current":
		f.health[0].Cur = uint16(basicSenderNumber5187E0(value))
	case "other_current":
		f.health[1].Cur = uint16(basicSenderNumber5187E0(value))
	case "health_link":
		f.unit.HealthData = (*server.HealthData)(f.pointer(value))
	case "update_link":
		f.unit.UpdateData = f.pointer(value)
	case "unit_class":
		f.unit.ObjClass = object.Class(basicSenderNumber5187E0(value))
	case "subclass":
		f.unit.ObjSubClass = object.SubClass(basicSenderNumber5187E0(value))
	case "type_index":
		f.unit.TypeInd = uint16(basicSenderNumber5187E0(value))
	case "net_code":
		f.unit.NetCode = basicSenderNumber5187E0(value)
	case "position_x":
		f.unit.PosVec.X = math.Float32frombits(basicSenderNumber5187E0(value))
	case "position_y":
		f.unit.PosVec.Y = math.Float32frombits(basicSenderNumber5187E0(value))
	case "direction":
		f.unit.Direction1 = server.Dir16(basicSenderNumber5187E0(value))
	case "damage_frame":
		f.unit.Frame134 = basicSenderNumber5187E0(value)
	default:
		f.t.Fatal("unknown mutation", name)
	}
}
func (f *basicSenderFixture5187E0) mutations(kind string) []any {
	v := basicSenderMap5187E0(f.input["mutations"])[kind]
	if v == nil {
		return []any{}
	}
	return v.([]any)
}
func (f *basicSenderFixture5187E0) mutate(kind string) {
	for _, raw := range f.mutations(kind) {
		pair := raw.([]any)
		f.change(pair[0].(string), pair[1])
	}
}
func (f *basicSenderFixture5187E0) patch(pointer unsafe.Pointer, data []byte) {
	f.t.Helper()
	target := uintptr(pointer)
	for i := range f.memory {
		m := &f.memory[i]
		base := uintptr(m.Pointer)
		if target >= base && target+uintptr(len(data)) <= base+m.Size {
			copy(m.Expected[target-base:], data)
			return
		}
	}
	f.t.Fatal("expected field outside guarded native records")
}
func basicSenderExpected5187E0[T any](f *basicSenderFixture5187E0, p *T, v T) {
	f.patch(unsafe.Pointer(p), basicSenderBytes5187E0(unsafe.Pointer(&v), unsafe.Sizeof(v)))
}
func (f *basicSenderFixture5187E0) checkState() {
	f.t.Helper()
	s := f.want.State
	unit := basicSenderMap5187E0(s["unit"])
	n := func(key string) uint32 { return basicSenderNumber5187E0(unit[key]) }
	basicSenderExpected5187E0(f, &f.unit.TypeInd, uint16(n("type_index")))
	basicSenderExpected5187E0(f, &f.unit.ObjClass, object.Class(n("unit_class")))
	basicSenderExpected5187E0(f, &f.unit.ObjSubClass, object.SubClass(n("subclass")))
	basicSenderExpected5187E0(f, &f.unit.NetCode, n("net_code"))
	pos := unit["position_bits"].([]any)
	basicSenderExpected5187E0(f, &f.unit.PosVec.X, math.Float32frombits(basicSenderNumber5187E0(pos[0])))
	basicSenderExpected5187E0(f, &f.unit.PosVec.Y, math.Float32frombits(basicSenderNumber5187E0(pos[1])))
	basicSenderExpected5187E0(f, &f.unit.Direction1, server.Dir16(n("direction")))
	basicSenderExpected5187E0(f, &f.unit.Frame134, n("damage_frame"))
	basicSenderExpected5187E0(f, &f.unit.HealthData, (*server.HealthData)(f.pointer(unit["health_link"])))
	basicSenderExpected5187E0(f, &f.unit.UpdateData, f.pointer(unit["update_link"]))
	hp := s["health"].([]any)
	for i, h := range f.health {
		basicSenderExpected5187E0(f, &h.Cur, uint16(basicSenderNumber5187E0(hp[i])))
	}
	updates := basicSenderMap5187E0(s["updates"])
	for i, name := range []string{"update", "other_update"} {
		samples := updates[name].([]any)
		if len(samples) != 32 {
			f.t.Fatal("sample dimensions")
		}
		for j, v := range samples {
			basicSenderExpected5187E0(f, &f.updates[i].HealthSamples[j], uint16(basicSenderNumber5187E0(v)))
		}
	}
	if f.s.Frame() != basicSenderNumber5187E0(s["frame"]) {
		f.t.Fatal("global frame differs")
	}
	for i, m := range f.memory {
		if !bytes.Equal(m.Expected, basicSenderBytes5187E0(m.Pointer, m.Size)) {
			f.t.Fatalf("whole native record/guard%d differs beyond original outputs", i)
		}
	}
}
func basicSenderCanonical5187E0(t *testing.T, v any) any {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var out any
	if err = json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	return out
}
func basicSenderWord5187E0(v uint32) *uint32 { return &v }
func (f *basicSenderFixture5187E0) event(kind string, args []any, ret *uint32, out []any) {
	f.t.Helper()
	i := len(f.events)
	if i >= len(f.want.Events) {
		f.t.Fatal("unexpected child", kind, args)
	}
	got := map[string]any{"service": kind, "arguments": args}
	if kind != "convert" && kind != "health_delta" {
		got["mutations"] = f.mutations(kind)
	}
	if ret != nil {
		got["return_word"] = *ret
	}
	if out != nil {
		got["out"] = out
	}
	want := make(map[string]any)
	for k, v := range f.want.Events[i] {
		if k != "ignored_high_word_diagnostic" {
			want[k] = v
		}
	}
	if !reflect.DeepEqual(basicSenderCanonical5187E0(f.t, got), want) {
		f.t.Fatalf("child%d=%v original=%v", i, got, want)
	}
	f.events = append(f.events, got)
	f.mutate(kind)
}
func (f *basicSenderFixture5187E0) runtime() server.BasicObjectSendRuntime5187E0 {
	return server.BasicObjectSendRuntime5187E0{
		Frame: func() uint32 { v := f.s.Frame(); f.frames = append(f.frames, v); return v },
		HealthDelta: func(index int32, code uint16, delta int16) {
			f.event("health_delta", []any{uint32(index), int16(code), delta}, nil, nil)
			first, _ := importantNativeListC()
			if first != nil {
				f.t.Fatal("actual C health queue not empty")
			}
			complexObjectReportHealthC518960(index, code, delta)
			recipient, packet := complexObjectTakeImportantPacketC518960()
			if packet != nil {
				// The original network service's unused return word is a declared input.
				// Actual C builder and real queue packet/recipient are checked here.
				word := basicSenderNumber5187E0(f.input["health_packet_return"])
				f.event("health_packet", []any{uint32(recipient), fmt.Sprintf("%x", packet), 5, 0, 1}, &word, nil)
			}
			first, _ = importantNativeListC()
			if first != nil {
				f.t.Fatal("extra actual C health publication")
			}
		},
		UnitCode: func(unit *server.Object) uint32 {
			if unit != f.unit {
				f.t.Fatal("unit code pointer")
			}
			v := basicSenderNumber5187E0(f.input["code_return"])
			f.event("unit_code", []any{"unit"}, &v, nil)
			return v
		},
		FloatToInt: func(value float32) int32 {
			var v int32
			if f.nearest {
				i := len(f.events)
				if i >= len(f.want.Events) || f.want.Events[i]["service"] != "convert" {
					f.t.Fatal("alternate converter order")
				}
				v = int32(basicSenderNumber5187E0(f.want.Events[i]["return_word"]))
			} else {
				v = complexObjectFloatToIntC518960(value)
			}
			f.event("convert", []any{math.Float32bits(value)}, basicSenderWord5187E0(uint32(v)), nil)
			return v
		},
		Direction: func(index int32) server.IndexedDirectionVector509E20 {
			v := f.input["direction_out"].([]any)
			x, y := basicSenderNumber5187E0(v[0]), basicSenderNumber5187E0(v[1])
			f.event("direction", []any{index}, basicSenderWord5187E0(0x76543210), []any{x, y})
			return server.IndexedDirectionVector509E20{X: int32(x), Y: int32(y)}
		},
		ServerPacket: func(index int32, packet []byte) int32 {
			v := basicSenderNumber5187E0(f.input["publication_return"])
			f.event("server_packet", []any{uint32(index), fmt.Sprintf("%x", packet), len(packet)}, &v, nil)
			return int32(v)
		},
	}
}

type basicSenderLegacyServer5187E0 struct {
	Server
	native *server.Server
}

func (s *basicSenderLegacyServer5187E0) S() *server.Server { return s.native }
func compareBasicSenderOriginal5187E0(t *testing.T, invoke func(string, int32, *server.Object) int32, nearest, faults bool) {
	t.Helper()
	rows, bad := basicSenderLoad5187E0(t, nearest)
	if faults {
		rows = bad
	}
	handles.InitHigh()
	t.Cleanup(handles.Release)
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	preserveImportantPacketState(t)
	setImportantNativeListC(nil, nil)
	setImportantListHeadsC(0, 0)
	setImportantAllocClassC(nil)
	setImportantRecipientMaskC(0)
	old := GetServer
	GetServer = func() Server { return &basicSenderLegacyServer5187E0{native: s} }
	t.Cleanup(func() { GetServer = old })
	for _, row := range rows {
		t.Run(row.Name, func(t *testing.T) {
			f := newBasicSenderFixture5187E0(t, s, row.Input, nearest)
			old := basicObjectSendRuntimeNative5187E0
			basicObjectSendRuntimeNative5187E0 = f.runtime
			t.Cleanup(func() { basicObjectSendRuntimeNative5187E0 = old })
			sequence := f.input["sequence"].([]any)
			if len(row.Output) != len(sequence)+1 {
				t.Fatal("repeat dimensions")
			}
			for i, want := range row.Output {
				if i != 0 {
					for field, value := range basicSenderMap5187E0(sequence[i-1]) {
						f.input[field] = value
						f.change(field, value)
					}
				}
				f.want = &want
				f.events = nil
				f.frames = nil
				unit := f.unit
				if f.input["nil_unit"] == true {
					unit = nil
				}
				var got int32
				var recovered any
				func() {
					defer func() { recovered = recover() }()
					got = invoke(f.input["kind"].(string), int32(basicSenderNumber5187E0(f.input["player"])), unit)
				}()
				if faults {
					if recovered == nil || len(want.Fault) == 0 {
						t.Fatal("original nil fault swallowed")
					}
				} else {
					if recovered != nil {
						t.Fatalf("unexpected native panic: %v", recovered)
					}
					if uint32(got) != want.Return {
						t.Fatalf("return=%x original=%x", uint32(got), want.Return)
					}
					if !want.Nonvolatile {
						t.Fatal("original nonvolatile guard missing")
					}
				}
				if !want.Heap || !want.Stack || !want.CW {
					t.Fatal("original guard proof missing")
				}
				if len(f.events) != len(want.Events) || len(f.frames) != len(want.Reads) {
					t.Fatalf("child/frame counts=%d/%d original=%d/%d", len(f.events), len(f.frames), len(want.Events), len(want.Reads))
				}
				for j, v := range f.frames {
					if v != basicSenderNumber5187E0(want.Reads[j][1]) {
						t.Fatal("frame read value")
					}
				}
				f.checkState()
			}
		})
	}
}
func basicSenderOriginalGoInvoke5187E0(kind string, index int32, unit *server.Object) int32 {
	if kind == "phantom" {
		return NetSendPhantomObject5187E0(index, unit)
	}
	return NetSendSimpleObject5188A0(index, unit)
}
func TestBasicSenderOriginalCGoStartup5187E0(t *testing.T) {
	compareBasicSenderOriginal5187E0(t, basicSenderOriginalCEntry5187E0, false, false)
}
func TestBasicSenderOriginalGoStartup5187E0(t *testing.T) {
	compareBasicSenderOriginal5187E0(t, basicSenderOriginalGoInvoke5187E0, false, false)
}
func TestBasicSenderOriginalCGoAlternateConverterBoundary5187E0(t *testing.T) {
	compareBasicSenderOriginal5187E0(t, basicSenderOriginalCEntry5187E0, true, false)
}
func TestBasicSenderOriginalGoAlternateConverterBoundary5187E0(t *testing.T) {
	compareBasicSenderOriginal5187E0(t, basicSenderOriginalGoInvoke5187E0, true, false)
}
func TestBasicSenderOriginalCGoFaults5187E0(t *testing.T) {
	compareBasicSenderOriginal5187E0(t, basicSenderOriginalCEntry5187E0, false, true)
}
func TestBasicSenderOriginalGoFaults5187E0(t *testing.T) {
	compareBasicSenderOriginal5187E0(t, basicSenderOriginalGoInvoke5187E0, false, true)
}
