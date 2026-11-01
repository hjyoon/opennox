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
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type specialFlagsOutput527E50 struct {
	Events      []map[string]any `json:"events"`
	State       map[string]any   `json:"state"`
	Return      uint32           `json:"return_word"`
	Heap        bool             `json:"all_other_heap_bytes_preserved"`
	Stack       bool             `json:"stack_guards_preserved"`
	CW          bool             `json:"CW_and_TOP_preserved"`
	Nonvolatile bool             `json:"nonvolatile_preserved"`
	Fault       []map[string]any `json:"fault"`
}
type specialFlagsRow527E50 struct {
	Name   string                     `json:"name"`
	Input  map[string]any             `json:"input"`
	Output []specialFlagsOutput527E50 `json:"output"`
}

func specialFlagsNumber527E50(v any) uint32 { return uint32(int64(v.(float64))) }
func specialFlagsNormalize527E50(v any) any {
	b, e := json.Marshal(v)
	if e != nil {
		panic(e)
	}
	var out any
	if e = json.Unmarshal(b, &out); e != nil {
		panic(e)
	}
	return out
}
func specialFlagsLoad527E50(t *testing.T) ([]specialFlagsRow527E50, []specialFlagsRow527E50) {
	t.Helper()
	fixture := filepath.Join("..", "server", "testdata", "object_special_original_527e50.json.gz")
	if supplied := os.Getenv("NOX_SPECIAL_ORIGINAL_REFERENCE"); supplied != "" {
		fixture = supplied
	}
	b, e := os.ReadFile(fixture)
	if e != nil {
		t.Fatal(e)
	}
	if fmt.Sprintf("%x", sha256.Sum256(b)) != "5b06030fdcf4683598c04a69b61e65966f35a190df00d5b9c7b91c7f0c7a3a86" {
		t.Fatal("compressed original seal")
	}
	z, e := gzip.NewReader(bytes.NewReader(b))
	if e != nil {
		t.Fatal(e)
	}
	raw, e := io.ReadAll(z)
	if e != nil {
		t.Fatal(e)
	}
	if e = z.Close(); e != nil {
		t.Fatal(e)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "5ef4e65a74a11f708054c3d79a45d175d613ca69244695ecdb774dc2ed83c37e" {
		t.Fatal("decoded original seal")
	}
	var d struct {
		Contexts []specialFlagsRow527E50 `json:"contexts"`
		Faults   []specialFlagsRow527E50 `json:"expected_fault_contexts"`
	}
	if e = json.Unmarshal(raw, &d); e != nil {
		t.Fatal(e)
	}
	if len(d.Contexts) != 1252 || len(d.Faults) != 6 {
		t.Fatal("original dimensions")
	}
	for _, r := range append(append([]specialFlagsRow527E50{}, d.Contexts...), d.Faults...) {
		for _, o := range r.Output {
			if !o.Heap || !o.Stack || !o.CW {
				t.Fatal("original guards", r.Name)
			}
		}
	}
	return d.Contexts, d.Faults
}

type specialFlagsGuard527E50[T comparable] struct {
	Before [32]byte
	Value  T
	After  [32]byte
}
type specialFlagsMemory527E50 struct {
	Pointer  unsafe.Pointer
	Expected []byte
}
type specialFlagsFixture527E50 struct {
	t               *testing.T
	recipient, unit *server.Object
	updates         [2]*server.PlayerUpdateData
	players         [2]*server.Player
	input           map[string]any
	events          []map[string]any
	memory          []specialFlagsMemory527E50
}

func specialFlagsBytes527E50(ptr unsafe.Pointer, size uintptr) []byte {
	return append([]byte(nil), unsafe.Slice((*byte)(ptr), int(size))...)
}
func specialFlagsAllocate527E50[T comparable](f *specialFlagsFixture527E50) *T {
	f.t.Helper()
	g, free := alloc.New(specialFlagsGuard527E50[T]{})
	f.t.Cleanup(free)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(g)) <= math.MaxUint32 {
		f.t.Fatal("guarded native records must exceed4GiB")
	}
	for i := range g.Before {
		g.Before[i] = byte(0x91 + i)
		g.After[i] = byte(0x51 + i)
	}
	f.memory = append(f.memory, specialFlagsMemory527E50{Pointer: unsafe.Pointer(g), Expected: make([]byte, int(unsafe.Sizeof(*g)))})
	return &g.Value
}
func (f *specialFlagsFixture527E50) mark(ptr unsafe.Pointer, size uintptr) {
	f.t.Helper()
	for i := range f.memory {
		m := &f.memory[i]
		off := uintptr(ptr) - uintptr(m.Pointer)
		if uintptr(ptr) >= uintptr(m.Pointer) && off+size <= uintptr(len(m.Expected)) {
			copy(m.Expected[int(off):int(off+size)], unsafe.Slice((*byte)(ptr), int(size)))
			return
		}
	}
	f.t.Fatal("mutation outside allocated records")
}
func (f *specialFlagsFixture527E50) patch(field string, v any) {
	f.t.Helper()
	switch {
	case field == "flags":
		index := specialFlagsNumber527E50(f.input["index"])
		if index < 32 {
			f.unit.Field140[index] = specialFlagsNumber527E50(v)
			f.mark(unsafe.Pointer(&f.unit.Field140[index]), 4)
		}
	case strings.HasPrefix(field, "peer/"):
		i, e := strconv.Atoi(strings.TrimPrefix(field, "peer/"))
		if e != nil || i < 0 || i >= 32 {
			f.t.Fatal(field)
		}
		f.unit.Field140[i] = specialFlagsNumber527E50(v)
		f.mark(unsafe.Pointer(&f.unit.Field140[i]), 4)
	case field == "index" || field == "other_index":
		i := 0
		if field == "other_index" {
			i = 1
		}
		f.players[i].PlayerInd = byte(specialFlagsNumber527E50(v))
		f.mark(unsafe.Pointer(&f.players[i].PlayerInd), 1)
	case field == "recipient_link":
		var ptr unsafe.Pointer
		if v != nil {
			if v == "update" {
				ptr = unsafe.Pointer(f.updates[0])
			} else if v == "other_update" {
				ptr = unsafe.Pointer(f.updates[1])
			} else {
				f.t.Fatal(v)
			}
		}
		f.recipient.UpdateData = ptr
		f.mark(unsafe.Pointer(&f.recipient.UpdateData), unsafe.Sizeof(ptr))
	case field == "player_link":
		var ptr *server.Player
		if v != nil {
			if v == "player" {
				ptr = f.players[0]
			} else if v == "other_player" {
				ptr = f.players[1]
			} else {
				f.t.Fatal(v)
			}
		}
		f.updates[0].Player = ptr
		f.mark(unsafe.Pointer(&f.updates[0].Player), unsafe.Sizeof(ptr))
	default:
		f.t.Fatal("unknown declared mutation", field)
	}
}
func specialFlagsFixtureNew527E50(t *testing.T, input map[string]any) *specialFlagsFixture527E50 {
	f := &specialFlagsFixture527E50{t: t, input: make(map[string]any)}
	for k, v := range input {
		f.input[k] = v
	}
	f.recipient = specialFlagsAllocate527E50[server.Object](f)
	f.unit = specialFlagsAllocate527E50[server.Object](f)
	for i := 0; i < 2; i++ {
		f.updates[i] = specialFlagsAllocate527E50[server.PlayerUpdateData](f)
		f.players[i] = specialFlagsAllocate527E50[server.Player](f)
	}
	f.recipient.ObjClass = 0 // The original raw allocation has no Player class bit.
	for i := range f.unit.Field140 {
		f.unit.Field140[i] = 0x10000000 + uint32(i)
	}
	f.updates[1].Player = f.players[1]
	for _, field := range []string{"index", "other_index", "recipient_link", "player_link", "flags"} {
		f.patch(field, f.input[field])
	}
	for i := range f.memory {
		m := &f.memory[i]
		m.Expected = specialFlagsBytes527E50(m.Pointer, uintptr(len(m.Expected)))
	}
	return f
}
func (f *specialFlagsFixture527E50) state() map[string]any {
	var ud, pl, other any
	switch f.recipient.UpdateData {
	case nil:
	case unsafe.Pointer(f.updates[0]):
		ud = "update"
	case unsafe.Pointer(f.updates[1]):
		ud = "other_update"
	default:
		f.t.Fatal("unexpected update pointer")
	}
	label := func(p *server.Player) any {
		switch p {
		case nil:
			return nil
		case f.players[0]:
			return "player"
		case f.players[1]:
			return "other_player"
		default:
			f.t.Fatal("unexpected player pointer")
			return nil
		}
	}
	pl, other = label(f.updates[0].Player), label(f.updates[1].Player)
	return map[string]any{"flags": f.unit.Field140, "recipient_link": ud, "player_link": pl, "other_player_link": other, "index": f.players[0].PlayerInd, "other_index": f.players[1].PlayerInd}
}
func (f *specialFlagsFixture527E50) event(kind string, args []any) int32 {
	before := f.unit.Field140
	mutations := []any{}
	if v, ok := f.input["mutations"].(map[string]any)[kind]; ok {
		mutations = v.([]any)
	}
	for _, v := range mutations {
		pair := v.([]any)
		f.patch(pair[0].(string), pair[1])
	}
	ret := specialFlagsNumber527E50(f.input["report_return"])
	if kind == "enemy" {
		ret = specialFlagsNumber527E50(f.input["enemy_return"])
	}
	f.events = append(f.events, map[string]any{"service": kind, "arguments": args, "return_word": ret, "flags_before": before, "mutations": mutations})
	return int32(ret)
}
func (f *specialFlagsFixture527E50) runtime() server.ObjectSpecialRuntime527E50 {
	report := func(kind string) func(int32, *server.Object) int32 {
		return func(index int32, u *server.Object) int32 {
			if u != f.unit || index < 0 || index >= 32 {
				f.t.Fatal("report arguments", kind, index, u)
			}
			return f.event(kind, []any{index, "unit"})
		}
	}
	return server.ObjectSpecialRuntime527E50{Animation: report("animation"), Enemy: func(r, u *server.Object) int32 {
		if r != f.recipient || u != f.unit {
			f.t.Fatal("enemy arguments")
		}
		return f.event("enemy", []any{"recipient", "unit"})
	}, Health: report("health"), Hidden: report("hidden"), XStatus: report("xstatus"), Height: report("height"), Enchant: report("enchant"), TeamBase: report("team"), NPC: report("npc")}
}
func (f *specialFlagsFixture527E50) check(w specialFlagsOutput527E50) {
	f.t.Helper()
	if !reflect.DeepEqual(specialFlagsNormalize527E50(f.events), specialFlagsNormalize527E50(w.Events)) {
		f.t.Fatalf("service trace\n got %s\nwant %s", mustSpecialFlagsJSON527E50(f.events), mustSpecialFlagsJSON527E50(w.Events))
	}
	if !reflect.DeepEqual(specialFlagsNormalize527E50(f.state()), specialFlagsNormalize527E50(w.State)) {
		f.t.Fatalf("complete state got=%s want=%s", mustSpecialFlagsJSON527E50(f.state()), mustSpecialFlagsJSON527E50(w.State))
	}
	var flags [32]uint32
	values := w.State["flags"].([]any)
	for i := range flags {
		flags[i] = specialFlagsNumber527E50(values[i])
	}
	ptr := unsafe.Pointer(&f.unit.Field140)
	found := false
	for i := range f.memory {
		m := &f.memory[i]
		off := uintptr(ptr) - uintptr(m.Pointer)
		if uintptr(ptr) >= uintptr(m.Pointer) && off+unsafe.Sizeof(flags) <= uintptr(len(m.Expected)) {
			copy(m.Expected[int(off):int(off+unsafe.Sizeof(flags))], unsafe.Slice((*byte)(unsafe.Pointer(&flags)), int(unsafe.Sizeof(flags))))
			found = true
		}
	}
	if !found {
		f.t.Fatal("flags allocation")
	}
	for _, m := range f.memory {
		if got := specialFlagsBytes527E50(m.Pointer, uintptr(len(m.Expected))); !bytes.Equal(got, m.Expected) {
			f.t.Fatal("native complete record or neighboring guard changed", m.Pointer)
		}
	}
}
func mustSpecialFlagsJSON527E50(v any) string {
	b, e := json.Marshal(v)
	if e != nil {
		panic(e)
	}
	return string(b)
}
func specialFlagsRunOriginal527E50(t *testing.T, cEntry, fault bool) {
	rows, faults := specialFlagsLoad527E50(t)
	if fault {
		rows = faults
	}
	for _, row := range rows {
		t.Run(row.Name, func(t *testing.T) {
			f := specialFlagsFixtureNew527E50(t, row.Input)
			old := objectSpecialRuntimeNative527E50
			objectSpecialRuntimeNative527E50 = f.runtime
			t.Cleanup(func() { objectSpecialRuntimeNative527E50 = old })
			sequence := []map[string]any{{}}
			for _, v := range row.Input["sequence"].([]any) {
				sequence = append(sequence, v.(map[string]any))
			}
			if len(sequence) != len(row.Output) {
				t.Fatal("original call count")
			}
			for i, patch := range sequence {
				for k, v := range patch {
					f.input[k] = v
				}
				for k, v := range patch {
					f.patch(k, v)
				}
				f.events = []map[string]any{}
				r, u := f.recipient, f.unit
				if f.input["nil_recipient"].(bool) {
					r = nil
				}
				if f.input["nil_unit"].(bool) {
					u = nil
				}
				var ret int32
				var caught any
				func() {
					defer func() { caught = recover() }()
					if cEntry {
						ret = specialFlagsOriginalCEntry527E50(r, u)
					} else {
						ret = NetUpdateObjectSpecial527E50(r, u)
					}
				}()
				want := row.Output[i]
				if fault {
					if caught == nil || len(want.Fault) == 0 {
						t.Fatal("expected raw nil-chain fault", caught)
					}
				} else {
					if caught != nil || uint32(ret) != want.Return || !want.Nonvolatile {
						t.Fatal("root result", ret, caught, want.Return)
					}
				}
				f.check(want)
			}
		})
	}
}
func TestObjectSpecialOriginal527E50Go(t *testing.T) { specialFlagsRunOriginal527E50(t, false, false) }
func TestObjectSpecialOriginal527E50C(t *testing.T)  { specialFlagsRunOriginal527E50(t, true, false) }
func TestObjectSpecialOriginalFaults527E50Go(t *testing.T) {
	specialFlagsRunOriginal527E50(t, false, true)
}
func TestObjectSpecialOriginalFaults527E50C(t *testing.T) {
	specialFlagsRunOriginal527E50(t, true, true)
}
