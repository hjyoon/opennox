//go:build darwin && arm64

package legacy

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

const tlInputSHA411350 = "bcb75fa919bb8def777b843bce044236f33f14b3b3729e7628522080953c2cac"

// Actual recorded reference bytes; independent original qualification is a separate root install gate.
const tlReferenceSHA411350 = "6ed209ef758521fc0991fe4fb8de5ca641281a6509311430bcbdc0613c6dad96"
const tlWorkerEnv411350 = "OPENNOX_TILE_LIST_411350_NATIVE_WORKER"
const tlDirectoryEnv411350 = "OPENNOX_TILE_LIST_411350_NATIVE_DIRECTORY"

type tlCallReference411350 struct {
	Site       uint32   `json:"site"`
	Target     uint32   `json:"target"`
	Argc       int      `json:"argc32"`
	Args       []uint32 `json:"arguments_u32"`
	Begin      int      `json:"original_call_event_begin"`
	End        int      `json:"return_event_end"`
	Complete   bool     `json:"completed"`
	Result     uint32   `json:"actual_return_EAX_u32"`
	HostReturn bool     `json:"host_return_substitution"`
}
type tlMemory411350 struct {
	Address uint32 `json:"address"`
	Before  string `json:"before"`
	After   string `json:"after"`
	Exact   bool   `json:"replay_exact"`
}
type tlReference411350 struct {
	Name         string                  `json:"name"`
	CW           uint16                  `json:"control_word"`
	Entry        uint32                  `json:"entry"`
	InputSHA     string                  `json:"input_sha256"`
	Status       string                  `json:"status"`
	Valid        bool                    `json:"certificate_valid"`
	Result       *uint32                 `json:"actual_EAX_u32"`
	Events       []swEvent410550         `json:"events"`
	Memory       []tlMemory411350        `json:"memory"`
	Calls        []tlCallReference411350 `json:"internal_calls"`
	Used         map[string][]int        `json:"used_mutations"`
	Unused       map[string][]int        `json:"unused_mutations"`
	Immutable    bool                    `json:"immutable_code_pages_unchanged"`
	External     int                     `json:"external_helper_opcodes_executed"`
	HostReturns  int                     `json:"host_return_substitutions"`
	Modelled     bool                    `json:"branch_or_return_modelled_in_host"`
	FaultClosure struct {
		PC      uint32 `json:"PC"`
		Address uint32 `json:"address"`
		Width   int    `json:"width"`
	} `json:"read_fault_mapping_closure"`
	BeforePages map[uint32][]byte `json:"-"` // retained parent-only; never child input
	AfterPages  map[uint32][]byte `json:"-"`
	RawSHA      string            `json:"-"`
}
type tlCase411350 struct {
	Input      tlInput411350
	Raw        json.RawMessage
	References map[uint16]*tlReference411350
}
type tlTrace411350 struct {
	Kind                                          string
	Site, Target, PE                              uint32
	Width, PEWidth                                int
	Complete, ValueValid, AddressValid, IsPointer bool
	Address, Base, Offset, Pointer                uint64
	Word                                          uint32
	Words                                         [2]uint32
	Result                                        int32
}
type tlChildInput411350 struct {
	Raw                    json.RawMessage
	Backend, ExecutableSHA string
	NativeOnlyNextPrefix   int
}
type tlChildResult411350 struct {
	PID                    int
	Backend                string
	Return                 uint32
	UsedReads, UsedReturns []int
	TraceEvents            int
}

func tlLoadCases411350() ([]tlCase411350, error) {
	base := filepath.Join("testdata", "tile_list_411350")
	var inputs struct {
		Cases []json.RawMessage `json:"cases"`
		Count int               `json:"case_count"`
		CW    []string          `json:"FPU_controls"`
	}
	if e := swReadGzipJSON410550(filepath.Join(base, "inputs-v1.json.gz"), tlInputSHA411350, &inputs); e != nil {
		return nil, e
	}
	if inputs.Count != 1049 || len(inputs.Cases) != 1049 || !reflect.DeepEqual(inputs.CW, []string{"0e3f", "023f"}) {
		return nil, fmt.Errorf("frozen1049/twoCW inputs")
	}
	cases := make([]tlCase411350, len(inputs.Cases))
	indices := map[string]int{}
	for i, p := range inputs.Cases {
		var compact bytes.Buffer
		if e := json.Compact(&compact, p); e != nil {
			return nil, e
		}
		cases[i].Raw = append(json.RawMessage(nil), compact.Bytes()...)
		if e := json.Unmarshal(p, &cases[i].Input); e != nil {
			return nil, e
		}
		in := &cases[i].Input
		if _, ok := indices[in.Name]; ok {
			return nil, fmt.Errorf("unique input name")
		}
		indices[in.Name] = i
		cases[i].References = map[uint16]*tlReference411350{}
		argc := 2
		if in.Entry == 0x411350 {
			argc = 3
		} else if in.Entry != 0x4113a0 && in.Entry != 0x411490 {
			return nil, fmt.Errorf("selected root")
		}
		if len(in.Args) != argc {
			return nil, fmt.Errorf("root arity")
		}
	}
	if len(tlReferenceSHA411350) != 64 {
		return nil, fmt.Errorf("qualified original SHA admission not filled")
	}
	f, z, e := rskOpenGzip4E23C0(filepath.Join(base, "reference-v1.json.gz"), tlReferenceSHA411350)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	defer z.Close()
	dec := json.NewDecoder(z)
	token, e := dec.Token()
	if e != nil || token != json.Delim('{') {
		return nil, fmt.Errorf("reference object")
	}
	seen := false
	count := 0
	for dec.More() {
		key, e := dec.Token()
		if e != nil {
			return nil, e
		}
		if key != "rows" {
			var ignored json.RawMessage
			if e = dec.Decode(&ignored); e != nil {
				return nil, e
			}
			continue
		}
		if seen {
			return nil, fmt.Errorf("one rows stream")
		}
		seen = true
		token, e = dec.Token()
		if e != nil || token != json.Delim('[') {
			return nil, fmt.Errorf("rows array")
		}
		for dec.More() {
			var raw json.RawMessage
			if e = dec.Decode(&raw); e != nil {
				return nil, e
			}
			var row tlReference411350
			if e = json.Unmarshal(raw, &row); e != nil {
				return nil, e
			}
			i, ok := indices[row.Name]
			if !ok {
				return nil, fmt.Errorf("unknown original input")
			}
			item := &cases[i]
			if !row.Valid || !row.Immutable || row.Modelled || row.External != 0 || row.HostReturns != 0 || row.Entry != item.Input.Entry || row.InputSHA != rskHash4E23C0(item.Raw) || (row.CW != 0x0e3f && row.CW != 0x023f) {
				return nil, fmt.Errorf("qualified actual original identity/scope")
			}
			if _, ok := item.References[row.CW]; ok {
				return nil, fmt.Errorf("unique originalCW")
			}
			if row.Status != "returned" && row.Status != "original_read_fault" {
				return nil, fmt.Errorf("original terminal domain")
			}
			expected := map[uint32]bool{0x22000000: true}
			for _, r := range item.Input.Regions {
				expected[r.Page] = true
			}
			for _, p := range item.Input.EdgePages {
				if p.Mapped {
					expected[p.Page] = true
				}
			}
			if len(row.Memory) != len(expected) {
				return nil, fmt.Errorf("all actual mutable original pages")
			}
			row.BeforePages = map[uint32][]byte{}
			row.AfterPages = map[uint32][]byte{}
			for _, m := range row.Memory {
				a, b := swHex410550(m.Before), swHex410550(m.After)
				if !m.Exact || len(a) != 4096 || len(b) != 4096 || !expected[m.Address] || row.BeforePages[m.Address] != nil || row.AfterPages[m.Address] != nil {
					return nil, fmt.Errorf("authentic whole BEFORE/AFTER/replay")
				}
				row.BeforePages[m.Address] = a
				row.AfterPages[m.Address] = b
			}
			row.Memory = nil
			row.RawSHA = rskHash4E23C0(raw)
			item.References[row.CW] = &row
			count++
		}
		if token, e = dec.Token(); e != nil || token != json.Delim(']') {
			return nil, fmt.Errorf("rows end")
		}
	}
	if token, e = dec.Token(); e != nil || token != json.Delim('}') {
		return nil, fmt.Errorf("reference end")
	}
	var extra any
	if e = dec.Decode(&extra); e != io.EOF {
		return nil, fmt.Errorf("reference EOF/CRC %v", e)
	}
	if !seen || count != 2098 {
		return nil, fmt.Errorf("exact2098 original rows")
	}
	for _, c := range cases {
		if len(c.References) != 2 {
			return nil, fmt.Errorf("two authenticCW references")
		}
	}
	return cases, nil
}

func tlOperand411350(g *tlGraph411350, op tileListNativeOperand411350) (uint32, int, error) {
	if !op.AddressValid {
		return 0, 0, fmt.Errorf("actual operand address")
	}
	if op.Site == 0x4114a2 || op.Site == 0x4114a8 {
		if op.Width != 1 || op.IsPointer || op.Offset != 0 || op.Address != op.Base {
			return 0, 0, fmt.Errorf("diagnostic BYTE exact width/base")
		}
		if op.Address != g.edgeAddress(op.Site, 0, op.LogicalPEAddress-uint32(0x85b3fc), op.LogicalPEAddress) {
			return 0, 0, fmt.Errorf("actual diagnostic EDGE address identity")
		}
		return op.LogicalPEAddress, 1, nil
	}
	var token uint32
	found := op.Base == nil
	for value, p := range g.Tokens {
		if p == op.Base {
			if found {
				return 0, 0, fmt.Errorf("ambiguous base identity")
			}
			token = value
			found = true
		}
	}
	if !found {
		return 0, 0, fmt.Errorf("actual native pointer identity")
	}
	width := 4
	pointer := op.Site == 0x411380
	if op.IsPointer != pointer || op.Width != uint32(map[bool]int{false: 4, true: 8}[pointer]) {
		return 0, 0, fmt.Errorf("native DWORD/fullPTR read width")
	}
	expected := map[uint32]uintptr{0x411363: 12, 0x411366: 8, 0x41137e: 0, 0x411380: 16, 0x4113a7: 0, 0x4113a9: 4}
	off, ok := expected[op.Site]
	if !ok || op.Offset != off {
		return 0, 0, fmt.Errorf("actual named operand offset")
	}
	address := op.Base
	if address != nil {
		address = unsafe.Add(address, off)
	}
	if address != op.Address {
		return 0, 0, fmt.Errorf("actual native effective address")
	}
	return token + uint32(off), width, nil
}

func TestTileListNativeWorker411350(t *testing.T) {
	if os.Getenv(tlWorkerEnv411350) != "1" {
		return
	}
	dir := os.Getenv(tlDirectoryEnv411350)
	if !filepath.IsAbs(dir) {
		t.Fatal("owned absolute child directory")
	}
	var input tlChildInput411350
	p, e := os.ReadFile(filepath.Join(dir, "input.json.gz"))
	if e != nil {
		t.Fatal(e)
	}
	z, e := gzip.NewReader(bytes.NewReader(p))
	if e != nil {
		t.Fatal(e)
	}
	data, e := io.ReadAll(z)
	ze := z.Close()
	if e != nil || ze != nil {
		t.Fatal(e, ze)
	}
	if e = json.Unmarshal(data, &input); e != nil {
		t.Fatal(e)
	}
	var in tlInput411350
	if e = json.Unmarshal(input.Raw, &in); e != nil {
		t.Fatal(e)
	}
	exe, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	sha, e := rskFileSHA4E23C0(exe)
	if e != nil || sha != input.ExecutableSHA {
		t.Fatal("actual preserved main", e)
	}
	g, e := tlBuildGraph411350(&in, filepath.Join(dir, "actual-shared-banks.bin"))
	if e != nil {
		t.Fatal(e)
	}
	defer g.close()
	if input.NativeOnlyNextPrefix != 0 {
		if e = g.nativeOnlyNextPrefix(input.NativeOnlyNextPrefix); e != nil {
			t.Fatal(e)
		}
	}
	if e = rskWriteJSON4E23C0(filepath.Join(dir, "graph.json"), g.Meta); e != nil {
		t.Fatal(e)
	}
	before := g.snapshot()
	if e = tlCheckInputNativeBefore411350(&in, &g.Meta, before, input.NativeOnlyNextPrefix); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(dir, "actual-native-BEFORE.bin"), before, 0600); e != nil {
		t.Fatal(e)
	}
	trace, e := os.OpenFile(filepath.Join(dir, "actual-observations.jsonl"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer trace.Close()
	enc := json.NewEncoder(trace)
	events, traceBytes := 0, 0
	must := func(e error) {
		if e != nil {
			rawDefaultDamageSharedValidationFailed4E0B30(g.Header, uint64(events))
			panic(e)
		}
	}
	emit := func(v tlTrace411350) {
		p, e := json.Marshal(v)
		must(e)
		events++
		traceBytes += len(p) + 1
		if events > 4096 || traceBytes > 1<<20 {
			must(fmt.Errorf("bounded observation stream"))
		}
		must(enc.Encode(v))
	}
	state := func(value uint32) tlChildResult411350 {
		return tlChildResult411350{os.Getpid(), input.Backend, value, g.UsedReads, g.UsedReturns, events}
	}
	save := func() { must(rskWriteJSON4E23C0(filepath.Join(dir, "actual-input-state.json"), state(0))) }
	// Keep this goroutine on the handler-owner thread through every actual load.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if e = rawDefaultDamageSharedBegin4E0B30(g.Header); e != nil {
		t.Fatal(e)
	}
	defer func() {
		if e := rawDefaultDamageSharedEnd4E0B30(); e != nil {
			t.Error(e)
		}
	}()
	observer := &tileListNativeObserver411350{EdgeByteAddress: g.edgeAddress}
	observer.Operand = func(op tileListNativeOperand411350) {
		pe, pw, e := tlOperand411350(g, op)
		must(e)
		if !op.Complete {
			if op.ValueValid {
				must(fmt.Errorf("BEFORE has no value"))
			}
			rawDefaultDamageSharedProgress4E0B30(g.Header, uint64(events), op.Site, op.Address, int(op.Width), rawDefaultDamageSharedRead4E0B30, nil)
		} else if !op.ValueValid {
			must(fmt.Errorf("completed actual read value"))
		}
		emit(tlTrace411350{Kind: "operand", Site: op.Site, PE: pe, Width: int(op.Width), PEWidth: pw, Complete: op.Complete, ValueValid: op.ValueValid, AddressValid: op.AddressValid, IsPointer: op.IsPointer, Address: uint64(uintptr(op.Address)), Base: uint64(uintptr(op.Base)), Offset: uint64(op.Offset), Word: op.ValueWord, Pointer: uint64(uintptr(op.ValuePointer))})
		if op.Complete {
			must(g.afterRead(op.Site, pe, pw))
			save()
		}
	}
	observer.Call = func(call tileListNativeCall411350) {
		if (call.Site != 0x41136b || call.Target != 0x411490) && (call.Site != 0x411372 || call.Target != 0x4113a0) {
			must(fmt.Errorf("actual internal site/target"))
		}
		emit(tlTrace411350{Kind: "call", Site: call.Site, Target: call.Target, Complete: call.Complete, Words: call.Words, Pointer: uint64(uintptr(call.Pointer)), Result: call.Result})
		if call.Complete {
			must(g.afterReturn(call.Site))
		}
		save()
	}
	pop := pushTileListNativeObserver411350(observer)
	defer pop()
	var root uintptr
	switch in.Entry {
	case 0x411350:
		root = reflect.ValueOf(sub411350Native).Pointer()
	case 0x4113a0:
		root = reflect.ValueOf(sub4113A0Native).Pointer()
	case 0x411490:
		root = reflect.ValueOf(sub411490Native).Pointer()
	default:
		t.Fatal("selected root")
	}
	backend := uint64(1)
	if input.Backend == "C" {
		backend = 2
		switch in.Entry {
		case 0x411350:
			root = reflect.ValueOf(tileListFixtureList411350).Pointer()
		case 0x4113a0:
			root = reflect.ValueOf(tileListFixturePoint411350).Pointer()
		case 0x411490:
			root = reflect.ValueOf(tileListFixtureEdge411350).Pointer()
		}
	} else if input.Backend != "Go" {
		t.Fatal("backend")
	}
	rawDefaultDamageSharedRootDispatched4E0B30(g.Header, root, backend)
	save()
	must(rskWriteJSON4E23C0(filepath.Join(dir, "actual-child-PREEXEC.json"), map[string]any{"pid": os.Getpid(), "parent_pid": os.Getppid(), "backend": input.Backend, "executable_sha256": sha, "input_sha256": rskHash4E23C0(input.Raw), "BEFORE_sha256": rskHash4E23C0(before), "prepared_before_actual_root": true, "original_observations_passed_to_child": false, "native_CW_not_simulated": true, "OS_thread_locked_before_shared_begin": true, "independent_input_native_BEFORE_validated": true}))
	var result int32
	switch in.Entry {
	case 0x411350:
		if len(in.Args) != 3 {
			t.Fatal("350 arity")
		}
		head, e := g.pointer(in.Args[0])
		must(e)
		xy, e := g.pointer(in.Args[1])
		must(e)
		if input.Backend == "C" {
			result = tileListFixtureList411350(head, xy, int32(in.Args[2]))
		} else {
			result = sub411350Native(head, xy, int32(in.Args[2]))
		}
	case 0x4113a0:
		if len(in.Args) != 2 {
			t.Fatal("3A0 arity")
		}
		xy, e := g.pointer(in.Args[0])
		must(e)
		if input.Backend == "C" {
			result = tileListFixturePoint411350(xy, int32(in.Args[1]))
		} else {
			result = sub4113A0Native(xy, int32(in.Args[1]))
		}
	case 0x411490:
		if len(in.Args) != 2 {
			t.Fatal("490 arity")
		}
		if input.Backend == "C" {
			result = tileListFixtureEdge411350(int32(in.Args[0]), int32(in.Args[1]))
		} else {
			result = sub411490Native(int32(in.Args[0]), int32(in.Args[1]))
		}
	}
	must(trace.Sync())
	must(rskWriteJSON4E23C0(filepath.Join(dir, "actual-return.json"), state(uint32(result))))
	runtime.KeepAlive(g)
}

type tlStep411350 struct {
	Sequence int
	Kind     string
	PC, PE   uint32
	Width    int
	Bits     uint32
	Call     *tlCallReference411350
}

func tlExpectedSteps411350(row *tlReference411350) ([]tlStep411350, error) {
	semantic := map[uint32]int{0x411363: 4, 0x411366: 4, 0x41137e: 4, 0x411380: 4, 0x4113a7: 4, 0x4113a9: 4, 0x4114a2: 1, 0x4114a8: 1}
	var steps []tlStep411350
	for _, ev := range row.Events {
		if width, ok := semantic[ev.PC]; ok && ev.Kind == "read" && ev.Completed {
			var hex string
			if e := json.Unmarshal(ev.Value, &hex); e != nil {
				return nil, e
			}
			b := swHex410550(hex)
			if ev.Width != width || len(b) != width {
				return nil, fmt.Errorf("original semantic read width")
			}
			steps = append(steps, tlStep411350{Sequence: ev.Sequence, Kind: "operand", PC: ev.PC, PE: ev.Address, Width: width, Bits: swWord410550(b)})
		}
	}
	for i := range row.Calls {
		c := &row.Calls[i]
		if c.HostReturn || c.Argc != 2 || len(c.Args) != 2 {
			return nil, fmt.Errorf("actual original internal call frame")
		}
		steps = append(steps, tlStep411350{Sequence: c.Begin, Kind: "call-before", PC: c.Site, Call: c})
		if c.Complete {
			steps = append(steps, tlStep411350{Sequence: c.End, Kind: "call-after", PC: c.Site, Call: c})
		}
	}
	if row.Status == "original_read_fault" {
		f := row.FaultClosure
		if semantic[f.PC] != f.Width {
			return nil, fmt.Errorf("unsupported original fault site")
		}
		seq := -1
		for _, ev := range row.Events {
			if ev.Kind == "fault" && ev.PC == f.PC {
				if seq >= 0 {
					return nil, fmt.Errorf("one actual original fault")
				}
				seq = ev.Sequence
			}
		}
		if seq < 0 {
			return nil, fmt.Errorf("original raw fault event")
		}
		steps = append(steps, tlStep411350{Sequence: seq, Kind: "pending", PC: f.PC, PE: f.Address, Width: f.Width})
	}
	sort.SliceStable(steps, func(i, j int) bool { return steps[i].Sequence < steps[j].Sequence })
	return steps, nil
}
func tlIdentity411350(meta *tlMeta411350, token uint32) (uint64, error) {
	if token == 0 {
		return 0, nil
	}
	p, ok := meta.Identities[strconv.FormatUint(uint64(token), 10)]
	if !ok {
		return 0, fmt.Errorf("original token missing native identity")
	}
	return p, nil
}
func tlExpectedAfter411350(meta *tlMeta411350, before []byte, row *tlReference411350) ([]byte, error) {
	if len(before) != len(meta.Banks)*meta.Page {
		return nil, fmt.Errorf("native whole BEFORE")
	}
	out := append([]byte(nil), before...)
	for _, f := range meta.Fields {
		if f.Accessible == 0 {
			continue
		}
		data, e := swOriginalBytes410550(row.AfterPages, f.PE, f.Accessible)
		if e != nil {
			return nil, e
		}
		if f.Pointer && f.Accessible == 4 {
			p, e := tlIdentity411350(meta, swWord410550(data))
			if e != nil {
				return nil, e
			}
			data = swBytes410550(p, f.NativeWidth)
		}
		off := f.Bank*meta.Page + f.NativeOffset
		if off < 0 || off+len(data) > len(out) || f.NativeOffset+len(data) > meta.Page {
			return nil, fmt.Errorf("projected field extent")
		}
		copy(out[off:], data)
	}
	return out, nil
}
func tlCompare411350(in *tlInput411350, row *tlReference411350, meta *tlMeta411350, trace []tlTrace411350, before, after []byte, result *tlChildResult411350) error {
	expectedBefore, e := tlCheckOriginalNativeBefore411350(in, row, meta, before, meta.NativeOnlyNextPrefix)
	if e != nil {
		return e
	}
	steps, e := tlExpectedSteps411350(row)
	if e != nil {
		return e
	}
	at := 0
	for _, s := range steps {
		if at >= len(trace) {
			return fmt.Errorf("missing native semantic step")
		}
		actual := trace[at]
		at++
		if s.Kind == "operand" || s.Kind == "pending" {
			pointer := s.PC == 0x411380
			nw := s.Width
			if pointer {
				nw = 8
			}
			if actual.Kind != "operand" || actual.Site != s.PC || actual.PE != s.PE || actual.PEWidth != s.Width || actual.Width != nw || actual.IsPointer != pointer || actual.Complete || actual.ValueValid || !actual.AddressValid {
				return fmt.Errorf("actual BEFORE scalar/PTR field")
			}
			if s.Kind == "pending" {
				if at != len(trace) {
					return fmt.Errorf("only trailing actual pending read")
				}
				continue
			}
			if at >= len(trace) {
				return fmt.Errorf("missing actual completed load")
			}
			done := trace[at]
			at++
			if done.Kind != actual.Kind || done.Site != actual.Site || done.PE != actual.PE || done.Width != actual.Width || done.PEWidth != actual.PEWidth || done.Base != actual.Base || done.Offset != actual.Offset || done.Address != actual.Address || done.IsPointer != actual.IsPointer || !done.Complete || !done.ValueValid || !done.AddressValid {
				return fmt.Errorf("actual completed load metadata")
			}
			if pointer {
				p, e := tlIdentity411350(meta, s.Bits)
				if e != nil {
					return e
				}
				if done.Pointer != p {
					return fmt.Errorf("actual fullPTR value")
				}
			} else if done.Word != s.Bits {
				return fmt.Errorf("actual scalar value")
			}
		} else {
			c := s.Call
			if actual.Kind != "call" || actual.Site != c.Site || actual.Target != c.Target || actual.Complete != (s.Kind == "call-after") {
				return fmt.Errorf("actual helper call/return order")
			}
			if c.Site == 0x41136b && c.Target == 0x411490 {
				if actual.Words != [2]uint32{c.Args[0], c.Args[1]} || actual.Pointer != 0 {
					return fmt.Errorf("490 actual ABI words")
				}
			} else if c.Site == 0x411372 && c.Target == 0x4113a0 {
				p, e := tlIdentity411350(meta, c.Args[0])
				if e != nil {
					return e
				}
				if actual.Pointer != p || actual.Words != [2]uint32{0, c.Args[1]} {
					return fmt.Errorf("3A0 fullpointer arg0 andword arg1")
				}
			} else {
				return fmt.Errorf("original internal site/target")
			}
			if actual.Complete && uint32(actual.Result) != c.Result {
				return fmt.Errorf("actual helper full int return")
			}
		}
	}
	if at != len(trace) {
		return fmt.Errorf("extra native semantic events")
	}
	if row.Status == "returned" && (row.Result == nil || result.Return != *row.Result) {
		return fmt.Errorf("actual root full int return")
	}
	if !reflect.DeepEqual(swCanonicalIndices410550(result.UsedReads), swCanonicalIndices410550(row.Used["after_reads"])) || !reflect.DeepEqual(swCanonicalIndices410550(result.UsedReturns), swCanonicalIndices410550(row.Used["after_internal_returns"])) {
		return fmt.Errorf("actual input stimulus consumption")
	}
	for key, n := range map[string]int{"after_reads": len(in.AfterReads), "after_internal_returns": len(in.AfterReturns)} {
		used := row.Used[key]
		unused := row.Unused[key]
		all := append(append([]int{}, used...), unused...)
		sort.Ints(all)
		if len(all) != n {
			return fmt.Errorf("all stimulus indices retained")
		}
		for i, v := range all {
			if v != i {
				return fmt.Errorf("unique stimulus index partition")
			}
		}
	}
	expected, e := tlExpectedAfter411350(meta, expectedBefore, row)
	if e != nil {
		return e
	}
	if !bytes.Equal(expected, after) {
		return fmt.Errorf("whole native AFTER/bank canary vs actual original projection")
	}
	return nil
}

func tlRun411350(dir, exe, sha, backend string, item *tlCase411350) error {
	return tlRunMode411350(dir, exe, sha, backend, item, 0)
}
func tlRunMode411350(dir, exe, sha, backend string, item *tlCase411350, nativePrefix int) (err error) {
	if e := os.Mkdir(dir, 0700); e != nil {
		return e
	}
	raw, e := json.Marshal(tlChildInput411350{item.Raw, backend, sha, nativePrefix})
	if e != nil {
		return e
	}
	f, e := os.OpenFile(filepath.Join(dir, "input.json.gz"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	z := gzip.NewWriter(f)
	_, e = z.Write(raw)
	ze, fe := z.Close(), f.Close()
	if e != nil || ze != nil || fe != nil {
		return fmt.Errorf("input-only child transport")
	}
	out, e := os.OpenFile(filepath.Join(dir, "stdout.log"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	defer out.Close()
	er, e := os.OpenFile(filepath.Join(dir, "stderr.log"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	defer er.Close()
	outLog, errLog := &rskBoundedLog4E23C0{File: out}, &rskBoundedLog4E23C0{File: er}
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, "-test.run=^TestTileListNativeWorker411350$", "-test.count=1", "-test.timeout=10s")
	cmd.Env = append(os.Environ(), tlWorkerEnv411350+"=1", tlDirectoryEnv411350+"="+dir)
	cmd.Stdout = outLog
	cmd.Stderr = errLog
	var forced atomic.Bool
	defaultCancel := cmd.Cancel
	cmd.Cancel = func() error { forced.Store(true); return defaultCancel() }
	started := time.Now().UTC()
	startErr := cmd.Start()
	pid := 0
	if cmd.Process != nil {
		pid = cmd.Process.Pid
	}
	var waitErr error
	if startErr == nil {
		waitErr = cmd.Wait()
	}
	exit := -1
	if cmd.ProcessState != nil {
		exit = cmd.ProcessState.ExitCode()
	}
	if e = rskWriteJSON4E23C0(filepath.Join(dir, "actual-child-Wait.json"), map[string]any{"pid": pid, "backend": backend, "start_UTC": started, "Wait_terminal_UTC": time.Now().UTC(), "start_error": fmt.Sprint(startErr), "Wait_error": fmt.Sprint(waitErr), "exit_code": exit, "forced_request": forced.Load(), "context_error": fmt.Sprint(ctx.Err()), "stdout_exceeded": outLog.Exceeded, "stderr_exceeded": errLog.Exceeded, "executable_sha256": sha, "reference_CW_comparisons_planned": 2, "original_execution_credit": 0}); e != nil {
		return e
	}
	defer func() {
		for _, name := range []string{"actual-shared-banks.bin", "actual-native-BEFORE.bin"} {
			p := filepath.Join(dir, name)
			if _, e := os.Lstat(p); e == nil {
				if e = rskGzipOwnedDuplicate4E23C0(p); e != nil && err == nil {
					err = e
				}
			}
		}
	}()
	if startErr != nil || forced.Load() || ctx.Err() != nil || outLog.Exceeded || errLog.Exceeded {
		return fmt.Errorf("native child did not terminate naturally: %v %v", startErr, waitErr)
	}
	var pre struct {
		PID               int    `json:"pid"`
		ParentPID         int    `json:"parent_pid"`
		Backend           string `json:"backend"`
		SHA               string `json:"executable_sha256"`
		InputSHA          string `json:"input_sha256"`
		BeforeSHA         string `json:"BEFORE_sha256"`
		ThreadLocked      bool   `json:"OS_thread_locked_before_shared_begin"`
		IndependentBefore bool   `json:"independent_input_native_BEFORE_validated"`
	}
	p, e := os.ReadFile(filepath.Join(dir, "actual-child-PREEXEC.json"))
	if e != nil {
		return e
	}
	if e = json.Unmarshal(p, &pre); e != nil {
		return e
	}
	if pre.PID != pid || pre.ParentPID != os.Getpid() || pre.Backend != backend || pre.SHA != sha || pre.InputSHA != rskHash4E23C0(item.Raw) || !pre.ThreadLocked || !pre.IndependentBefore {
		return fmt.Errorf("actual samePID/parent/PREEXEC/input/thread binding")
	}
	var meta tlMeta411350
	p, e = os.ReadFile(filepath.Join(dir, "graph.json"))
	if e != nil {
		return e
	}
	if e = json.Unmarshal(p, &meta); e != nil {
		return e
	}
	before, e := os.ReadFile(filepath.Join(dir, "actual-native-BEFORE.bin"))
	if e != nil || rskHash4E23C0(before) != pre.BeforeSHA {
		return fmt.Errorf("actual BEFORE binding")
	}
	for _, cw := range []uint16{0x0e3f, 0x023f} {
		row := item.References[cw]
		if row == nil {
			return fmt.Errorf("both original BEFORE references")
		}
		if _, e = tlCheckOriginalNativeBefore411350(&item.Input, row, &meta, before, nativePrefix); e != nil {
			return fmt.Errorf("independent BEFORE CW%04x:%w", cw, e)
		}
	}
	raw, e = os.ReadFile(filepath.Join(dir, "actual-shared-banks.bin"))
	if e != nil {
		return e
	}
	if len(raw) != (len(meta.Banks)+1)*meta.Page {
		return fmt.Errorf("shared bank extent")
	}
	header, e := rskReadHeader4E23C0(raw)
	if e != nil {
		return e
	}
	after := raw[meta.Page:]
	if header[0] != rawDefaultDamageSharedMagic4E0B30 || header[1] != rawDefaultDamageSharedVersion4E0B30 || header[36] != 0 || header[38] != 1 || header[40] != (map[string]uint64{"Go": 1, "C": 2}[backend]) {
		return fmt.Errorf("actual shared header/root state")
	}
	var state tlChildResult411350
	p, e = os.ReadFile(filepath.Join(dir, "actual-input-state.json"))
	if e != nil {
		return e
	}
	if e = json.Unmarshal(p, &state); e != nil {
		return e
	}
	if state.PID != pid || state.Backend != backend {
		return fmt.Errorf("actual durable input state")
	}
	result := state
	if exit == 0 {
		p, e = os.ReadFile(filepath.Join(dir, "actual-return.json"))
		if e != nil {
			return e
		}
		if e = json.Unmarshal(p, &result); e != nil {
			return e
		}
		if result.PID != pid || result.Backend != backend || header[2] != 2 || !reflect.DeepEqual(result.UsedReads, state.UsedReads) || !reflect.DeepEqual(result.UsedReturns, state.UsedReturns) || result.TraceEvents != state.TraceEvents {
			return fmt.Errorf("actual returned state/header")
		}
	} else if exit == 172 {
		if header[2] != 5 || (header[3] != 10 && header[3] != 11) || header[20] == 0 || header[20] != header[21] || header[22] != 0 || header[23] != 0 || header[24] != 0 || header[10] != 3 || header[6] == 0 || header[8] == 0 || header[8] > 8 || header[5] < header[7] || header[5]-header[7] >= header[8] {
			return fmt.Errorf("actual native owner-thread read fault")
		}
	} else {
		return fmt.Errorf("unexpected actual exit%d", exit)
	}
	f, e = os.Open(filepath.Join(dir, "actual-observations.jsonl"))
	if e != nil {
		return e
	}
	dec := json.NewDecoder(f)
	var trace []tlTrace411350
	for {
		var v tlTrace411350
		e = dec.Decode(&v)
		if e == io.EOF {
			break
		}
		if e != nil {
			f.Close()
			return e
		}
		trace = append(trace, v)
		if len(trace) > 4096 {
			f.Close()
			return fmt.Errorf("trace bound")
		}
	}
	f.Close()
	pending := 0
	if exit == 172 {
		pending = 1
	}
	if len(trace) != result.TraceEvents+pending {
		return fmt.Errorf("durable trace plus one actual pending BEFORE")
	}
	if exit == 172 {
		if len(trace) == 0 {
			return fmt.Errorf("actual pending read missing")
		}
		v := trace[len(trace)-1]
		if v.Kind != "operand" || v.Complete || v.ValueValid || uint64(v.Site) != header[9] || v.Address != header[7] || uint64(v.Width) != header[8] || header[13] != uint64(len(trace)-1) || header[13] != uint64(state.TraceEvents) {
			return fmt.Errorf("fault header/pending read width/site/address/progress")
		}
	}
	if nativePrefix != 0 {
		if meta.NativeOnlyNextPrefix != nativePrefix || nativePrefix < 20 || nativePrefix > 24 || !bytes.Equal(before, after) {
			return fmt.Errorf("native-only actual geometry/full canaries")
		}
		if nativePrefix < 24 {
			if exit != 172 || len(trace) == 0 {
				return fmt.Errorf("actual8byteNext guard must be reached")
			}
			last := trace[len(trace)-1]
			if last.Site != 0x411380 || !last.IsPointer || last.Width != 8 || last.Complete || last.ValueValid {
				return fmt.Errorf("native-only unfinished fullPTR Next")
			}
			return rskWriteJSON4E23C0(filepath.Join(dir, "actual-qualification.json"), map[string]any{"qualified_native_only_width_control": true, "actual_native_calls": 1, "actual_original_CW_comparisons": 0, "actual_pid": pid, "backend": backend, "native_node_prefix": nativePrefix, "actual_pending_Next_width": 8, "whole_native_banks_unchanged": true, "original_prefix_equivalence": false, "independent_native_fault_PC_qualification": false, "signal_header": header})
		}
		if exit != 0 {
			return fmt.Errorf("complete native24 control must return naturally")
		}
	}
	rowSHAs := map[string]string{}
	for _, cw := range []uint16{0x0e3f, 0x023f} {
		row := item.References[cw]
		if row == nil {
			return fmt.Errorf("both actual originalCW rows")
		}
		if (row.Status == "returned") != (exit == 0) {
			return fmt.Errorf("actual terminal domain vsCW%04x", cw)
		}
		if e = tlCompare411350(&item.Input, row, &meta, trace, before, after, &result); e != nil {
			return fmt.Errorf("CW%04x:%w", cw, e)
		}
		rowSHAs[fmt.Sprintf("%04x", cw)] = row.RawSHA
	}
	return rskWriteJSON4E23C0(filepath.Join(dir, "actual-qualification.json"), map[string]any{"qualified_selected_observations": true, "actual_pid": pid, "backend": backend, "input": item.Input.Name, "input_sha256": rskHash4E23C0(item.Raw), "original_row_sha256s": rowSHAs, "actual_original_CW_comparisons": 2, "actual_native_calls": 1, "native_before_sha256": rskHash4E23C0(before), "native_after_sha256": rskHash4E23C0(after), "actual_native_returned": exit == 0, "whole_native_banks_and_canaries": true, "full_native_Next_identities": true, "actual_internal_body_order_and_returns": true, "physical_original_stack_FPU_or_nil_EA_equivalence": false, "independent_native_fault_PC_qualification": false, "native_Go_heap_credit": false, "signal_header": header})
}

func TestTileListNativeOriginal411350(t *testing.T) {
	if os.Getenv(tlWorkerEnv411350) == "1" {
		return
	}
	cases, e := tlLoadCases411350()
	if e != nil {
		t.Fatal(e)
	}
	initial := os.Getenv("OPENNOX_RAW_DEFAULT_DAMAGE_INITIAL_ARTIFACTS")
	if initial == "" {
		t.Fatal("prepared unique INITIAL_ARTIFACTS required")
	}
	root, e := filepath.Abs(filepath.Join(initial, "tile-list-three-411350"))
	if e != nil {
		t.Fatal(e)
	}
	if e = os.MkdirAll(filepath.Dir(root), 0700); e != nil {
		t.Fatal(e)
	}
	if e = os.Mkdir(root, 0700); e != nil {
		t.Fatal("unique artifact directory", e)
	}
	exe, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	sha, e := rskArchiveMain4E23C0(exe, root)
	if e != nil {
		t.Fatal(e)
	}
	attempted, passed := 0, 0
	for i := range cases {
		for _, backend := range []string{"Go", "C"} {
			attempted++
			item := &cases[i]
			if t.Run(backend+"/"+item.Input.Name, func(t *testing.T) {
				if e := tlRun411350(filepath.Join(root, fmt.Sprintf("%04d-%s", i, backend)), exe, sha, backend, item); e != nil {
					t.Fatal(e)
				}
			}) {
				passed++
			}
		}
	}
	widthCase := tlWidthControl411350(t, cases)
	widthAttempted, widthPassed := 0, 0
	for _, backend := range []string{"Go", "C"} {
		for _, prefix := range []int{20, 21, 22, 23, 24} {
			widthAttempted++
			if t.Run(fmt.Sprintf("%s/native-only-next-prefix-%d", backend, prefix), func(t *testing.T) {
				if e := tlRunMode411350(filepath.Join(root, fmt.Sprintf("native-only-%d-%s", prefix, backend)), exe, sha, backend, widthCase, prefix); e != nil {
					t.Fatal(e)
				}
			}) {
				widthPassed++
			}
		}
	}
	if e = rskWriteJSON4E23C0(filepath.Join(root, "actual-corpus-summary.json"), map[string]any{"inputs": 1049, "attempted_native_calls": attempted, "passed": passed, "failed": attempted - passed, "stream_complete": attempted == 2098, "qualified_original_CW_comparisons": passed * 2, "input1049_sha256": tlInputSHA411350, "original2098_sha256": tlReferenceSHA411350, "private_original_stack_FPU_or_nil_EA_equivalence": false, "native_fault_PC_equivalence": false, "native_Go_heap_credit": false}); e != nil {
		t.Fatal(e)
	}
	if attempted != 2098 || passed != 2098 {
		t.Fatalf("native calls attempted%d passed%d", attempted, passed)
	}
	t.Logf("TILE_LIST_NATIVE_ORIGINAL_OK inputs=1049 native=2098 original-CW-comparisons=4196 actual-internal-bodies=true full-native-Next=true")
	if e = rskWriteJSON4E23C0(filepath.Join(root, "actual-native-only-width-summary.json"), map[string]any{"attempted_native_calls": widthAttempted, "passed": widthPassed, "prefixes": []int{20, 21, 22, 23, 24}, "original_prefix_equivalence": false, "partial_prefix_original_comparisons": 0, "complete24_original_comparisons": 4}); e != nil {
		t.Fatal(e)
	}
	if widthAttempted != 10 || widthPassed != 10 {
		t.Fatalf("native-only width controls attempted%d passed%d", widthAttempted, widthPassed)
	}
}
