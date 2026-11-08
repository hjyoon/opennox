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

const waInputSHA410390 = "ddf5d082e83be438c68a6a69079f5365e9c18df3cda28ee3a5ad79f703f988bd"
const waReferenceSHA410390 = "e032dde3b9dff724b966351693dcd683d3eaff8c73ccf74ec1ea52835dcc2701"
const waWorkerEnv410390 = "OPENNOX_WALL_ATTACH_410390_NATIVE_WORKER"
const waDirectoryEnv410390 = "OPENNOX_WALL_ATTACH_410390_NATIVE_DIRECTORY"

type waEvent410390 struct {
	Kind      string            `json:"kind"`
	Sequence  int               `json:"sequence"`
	PC        uint32            `json:"PC"`
	Address   uint32            `json:"address"`
	Width     int               `json:"width"`
	Value     json.RawMessage   `json:"value"`
	Completed bool              `json:"completed"`
	Registers map[string]uint32 `json:"registers"`
}
type waServiceReference410390 struct {
	Site          uint32         `json:"site"`
	Target        uint32         `json:"target"`
	Args          []uint32       `json:"arguments_u32"`
	Sequence      int            `json:"original_call_event_begin"`
	Occurrence    int            `json:"occurrence"`
	ResponseIndex int            `json:"response_index"`
	Packet        waPacket410390 `json:"packet"`
	LocalHex      string         `json:"local_int2_hex"`
	LocalAddress  uint32         `json:"local_int2_address"`
}
type waMemory410390 struct {
	Address uint32 `json:"address"`
	After   string `json:"after"`
	Exact   bool   `json:"replay_exact"`
}
type waReference410390 struct {
	Name            string                     `json:"name"`
	CW              uint16                     `json:"control_word"`
	Entry           uint32                     `json:"entry"`
	InputSHA        string                     `json:"input_sha256"`
	Status          string                     `json:"status"`
	Valid           bool                       `json:"certificate_valid"`
	Result          *uint32                    `json:"actual_EAX_u32"`
	Events          []waEvent410390            `json:"events"`
	Services        []waServiceReference410390 `json:"services"`
	Memory          []waMemory410390           `json:"memory"`
	Used            map[string][]int           `json:"used_mutations"`
	Unused          map[string][]int           `json:"unused_mutations"`
	UsedResponses   map[string][]int           `json:"used_response_indices"`
	UnusedResponses map[string][]int           `json:"unused_response_indices"`
	AfterFields     map[uint32][]byte          `json:"-"`
	FaultClosure    struct {
		PC             uint32 `json:"PC"`
		OperandAddress uint32 `json:"address"`
		OperandWidth   int    `json:"width"`
	} `json:"read_fault_mapping_closure"`
	RawSHA string `json:"-"`
}
type waCase410390 struct {
	Input      waInput410390
	Raw        json.RawMessage
	References map[uint16]*waReference410390
}
type waArgument410390 struct {
	Word      uint32
	Pointer   uint64
	IsPointer bool
}
type waTrace410390 struct {
	Kind                                                 string
	Site, PE                                             uint32
	PEWidth, Width                                       int
	Complete, ValueValid, AddressValid, Write, IsPointer bool
	Address, Base, Pointer                               uint64
	Offset                                               uint64
	Bits                                                 uint64
	LocalAddress                                         uint64
	LocalHex                                             string
	Word                                                 uint32
	Arguments                                            []waArgument410390
	ResponseIndex                                        int
	ReplyWord, ReplyECX, ReplyEDX                        uint32
	ReplyPointer                                         uint64
}
type waChildInput410390 struct {
	Raw                    json.RawMessage
	ControlWord            uint16
	Backend, ExecutableSHA string
}
type waChildResult410390 struct {
	PID                     int
	Backend                 string
	ControlWord             uint16
	Return                  uint64
	UsedReads, UsedServices []int
	UsedResponses           map[string][]int
	TraceEvents             int
}

func waCanonicalIndices410390(v []int) []int {
	out := append([]int{}, v...)
	sort.Ints(out)
	for i := 1; i < len(out); i++ {
		if out[i] == out[i-1] {
			panic("duplicate consumed stimulus")
		}
	}
	return out
}
func waResponseIndices410390(values map[string][]int) map[string][]int {
	out := map[string][]int{}
	for site, v := range values {
		unique := map[int]bool{}
		for _, n := range v {
			unique[n] = true
		}
		p := []int{}
		for n := range unique {
			p = append(p, n)
		}
		sort.Ints(p)
		out[site] = p
	}
	return out
}
func waReadGzipJSON410390(path, sha string, out any) error {
	f, z, e := rskOpenGzip4E23C0(path, sha)
	if e != nil {
		return e
	}
	defer f.Close()
	defer z.Close()
	d := json.NewDecoder(z)
	if e = d.Decode(out); e != nil {
		return e
	}
	var extra any
	if e = d.Decode(&extra); e != io.EOF {
		return fmt.Errorf("one gzip JSON document and complete EOF/CRC: %v", e)
	}
	return nil
}
func waLoadCases410390() ([]waCase410390, error) {
	base := filepath.Join("testdata", "wall_attach_410390")
	var inputs struct {
		Cases []json.RawMessage `json:"cases"`
		Count int               `json:"case_count"`
		CW    []string          `json:"FPU_controls"`
	}
	if e := waReadGzipJSON410390(filepath.Join(base, "inputs-v1.json.gz"), waInputSHA410390, &inputs); e != nil {
		return nil, e
	}
	if inputs.Count != 375 || len(inputs.Cases) != 375 || !reflect.DeepEqual(inputs.CW, []string{"0e3f", "023f"}) {
		return nil, fmt.Errorf("exact375 independent inputs")
	}
	cases := make([]waCase410390, 375)
	indices := map[string]int{}
	for i, p := range inputs.Cases {
		var b bytes.Buffer
		if e := json.Compact(&b, p); e != nil {
			return nil, e
		}
		cases[i].Raw = append(json.RawMessage(nil), b.Bytes()...)
		cases[i].References = map[uint16]*waReference410390{}
		if e := json.Unmarshal(p, &cases[i].Input); e != nil {
			return nil, e
		}
		name := cases[i].Input.Name
		if _, ok := indices[name]; ok {
			return nil, fmt.Errorf("unique independent name")
		}
		indices[name] = i
	}
	f, z, e := rskOpenGzip4E23C0(filepath.Join(base, "reference-v1.json.gz"), waReferenceSHA410390)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	defer z.Close()
	dec := json.NewDecoder(z)
	token, e := dec.Token()
	if e != nil || token != json.Delim('{') {
		return nil, fmt.Errorf("original reference object")
	}
	count := 0
	rowsSeen := false
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
		if rowsSeen {
			return nil, fmt.Errorf("one original rows stream")
		}
		rowsSeen = true
		token, e = dec.Token()
		if e != nil || token != json.Delim('[') {
			return nil, fmt.Errorf("original rows array")
		}
		for dec.More() {
			var raw json.RawMessage
			if e = dec.Decode(&raw); e != nil {
				return nil, e
			}
			var row waReference410390
			if e = json.Unmarshal(raw, &row); e != nil {
				return nil, e
			}
			i, ok := indices[row.Name]
			if !ok {
				return nil, fmt.Errorf("original unknown input")
			}
			item := &cases[i]
			if !row.Valid || row.Entry != item.Input.Entry || row.InputSHA != rskHash4E23C0(item.Raw) || (row.CW != 0x0e3f && row.CW != 0x023f) {
				return nil, fmt.Errorf("actual original source/input/CW certificate")
			}
			if _, ok := item.References[row.CW]; ok {
				return nil, fmt.Errorf("unique actual original CW")
			}
			row.RawSHA = rskHash4E23C0(raw)
			pages := map[uint32][]byte{}
			for _, m := range row.Memory {
				data := waHex410390(m.After)
				if !m.Exact || len(data) != 4096 {
					return nil, fmt.Errorf("source-qualified original full AFTER page")
				}
				pages[m.Address] = data
			}
			row.AfterFields = map[uint32][]byte{}
			add := func(pe uint32, width int) error {
				available := waAvailable410390(&item.Input, pe, width)
				if available == 0 {
					return nil
				}
				p, e := waOriginalBytes410390(pages, pe, available)
				if e != nil {
					return e
				}
				if old, ok := row.AfterFields[pe]; ok && !bytes.Equal(old, p) {
					return fmt.Errorf("aliased authentic AFTER projection")
				}
				row.AfterFields[pe] = p
				return nil
			}
			for role, t := range item.Input.Roles {
				if t == 0 {
					continue
				}
				switch {
				case role == "drawable":
					e = add(t+12, 4)
					if e == nil {
						e = add(t+16, 4)
					}
				case len(role) >= 4 && role[:4] == "wall":
					e = add(t+4, 1)
					if e == nil {
						e = add(t+8, 1)
					}
					if e == nil {
						e = add(t+32, 4)
					}
				case len(role) >= 7 && role[:7] == "polygon":
					e = add(t+130, 1)
				}
				if e != nil {
					return nil, e
				}
			}
			row.Memory = nil
			item.References[row.CW] = &row
			count++
		}
		if token, e = dec.Token(); e != nil || token != json.Delim(']') {
			return nil, fmt.Errorf("complete rows array")
		}
	}
	if token, e = dec.Token(); e != nil || token != json.Delim('}') {
		return nil, fmt.Errorf("complete original object")
	}
	var extra any
	if e = dec.Decode(&extra); e != io.EOF {
		return nil, fmt.Errorf("original gzip EOF/CRC %v", e)
	}
	if count != 750 || !rowsSeen {
		return nil, fmt.Errorf("exact750 authentic original rows")
	}
	for _, c := range cases {
		if len(c.References) != 2 {
			return nil, fmt.Errorf("two original CW rows per input")
		}
	}
	return cases, nil
}

// No original reference or predicted observation schedule enters this worker.
func TestWallAttachNativeWorker410390(t *testing.T) {
	if os.Getenv(waWorkerEnv410390) != "1" {
		return
	}
	dir := os.Getenv(waDirectoryEnv410390)
	if !filepath.IsAbs(dir) {
		t.Fatal("owned absolute child directory")
	}
	var input waChildInput410390
	p, e := os.ReadFile(filepath.Join(dir, "input.json.gz"))
	if e != nil {
		t.Fatal(e)
	}
	z, e := gzip.NewReader(bytes.NewReader(p))
	if e != nil {
		t.Fatal(e)
	}
	data, e := io.ReadAll(z)
	ce := z.Close()
	if e != nil || ce != nil {
		t.Fatal(e, ce)
	}
	if e = json.Unmarshal(data, &input); e != nil {
		t.Fatal(e)
	}
	if input.ControlWord != 0x0e3f && input.ControlWord != 0x023f {
		t.Fatal("one independent selected CW")
	}
	var in waInput410390
	if e = json.Unmarshal(input.Raw, &in); e != nil {
		t.Fatal(e)
	}
	exe, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	sha, e := rskFileSHA4E23C0(exe)
	if e != nil || sha != input.ExecutableSHA {
		t.Fatal("same preserved actual main", e)
	}
	g, e := waBuildGraph410390(&in, filepath.Join(dir, "actual-shared-banks.bin"))
	if e != nil {
		t.Fatal(e)
	}
	defer g.close()
	if e = rskWriteJSON4E23C0(filepath.Join(dir, "graph.json"), g.Meta); e != nil {
		t.Fatal(e)
	}
	before := g.snapshot()
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
	emit := func(v waTrace410390) {
		p, e := json.Marshal(v)
		must(e)
		events++
		traceBytes += len(p) + 1
		if events > 1024 || traceBytes > 256<<10 {
			must(fmt.Errorf("bounded native observation trace"))
		}
		must(enc.Encode(v))
	}
	state := func(value uint64) waChildResult410390 {
		return waChildResult410390{os.Getpid(), input.Backend, input.ControlWord, value, g.UsedReads, g.UsedServices, waResponseIndices410390(g.UsedResponses), events}
	}
	for site := range in.Services {
		g.UsedResponses[site] = []int{}
	}
	save := func() { must(rskWriteJSON4E23C0(filepath.Join(dir, "actual-input-state.json"), state(0))) }
	if e = rawDefaultDamageSharedBegin4E0B30(g.Header); e != nil {
		t.Fatal(e)
	}
	defer func() {
		if e := rawDefaultDamageSharedEnd4E0B30(); e != nil {
			t.Error(e)
		}
	}()
	observer := &wallAttachNativeObserver410390{}
	observer.Operand = func(op wallAttachNativeOperand410390) {
		pe, width, e := waOperand410390(g, op)
		must(e)
		if !op.Complete {
			rawDefaultDamageSharedProgress4E0B30(g.Header, uint64(events), op.Site, op.Address, int(op.Width), map[bool]int{false: rawDefaultDamageSharedRead4E0B30, true: rawDefaultDamageSharedWrite4E0B30}[op.Write], nil)
		}
		emit(waTrace410390{Kind: "operand", Site: op.Site, PE: pe, PEWidth: width, Width: int(op.Width), Complete: op.Complete, ValueValid: op.ValueValid, AddressValid: op.AddressValid, Write: op.Write, IsPointer: op.IsPointer, Address: uint64(uintptr(op.Address)), Base: uint64(uintptr(op.Base)), Offset: uint64(op.Offset), Word: uint32(op.ValueBits), Bits: op.ValueBits, Pointer: uint64(uintptr(op.ValuePointer))})
		if op.Complete {
			if !op.Write {
				must(g.afterRead(op.Site, pe, width))
			}
			save()
		}
	}
	observer.Service = func(site uint32, args []wallAttachNativeArg410390) (wallAttachNativeReply410390, bool) {
		key := fmt.Sprintf("%08x", site)
		s, ok := in.Services[key]
		if !ok || len(args) != s.Argc || len(s.Responses) == 0 {
			must(fmt.Errorf("declared actual service site/arity"))
		}
		index := g.Services[site]
		if index >= len(s.Responses) {
			if !s.Repeat {
				must(fmt.Errorf("undeclared repeated service"))
			}
			index = len(s.Responses) - 1
		}
		packet := s.Responses[index]
		reply := wallAttachNativeReply410390{Word: packet.Registers["eax"], ECX: packet.Registers["ecx"], EDX: packet.Registers["edx"]}
		if waPointerSite410390(site) {
			reply.Pointer, e = g.pointer(packet.Registers["eax"])
			must(e)
		}
		v := waTrace410390{Kind: "service", Site: site, Complete: true, ResponseIndex: index, ReplyWord: reply.Word, ReplyECX: reply.ECX, ReplyEDX: reply.EDX, ReplyPointer: uint64(uintptr(reply.Pointer))}
		for _, a := range args {
			v.Arguments = append(v.Arguments, waArgument410390{a.Word, uint64(uintptr(a.Pointer)), a.IsPointer})
		}
		if site == 0x4103e3 || site == 0x4103fa {
			if len(args) == 0 || !args[0].IsPointer || args[0].Pointer == nil || args[0].Pointer != g.Local {
				must(fmt.Errorf("actual query uses same private local8"))
			}
			v.LocalAddress = uint64(uintptr(args[0].Pointer))
			v.LocalHex = fmt.Sprintf("%x", ddiSnapshot4E0B30(args[0].Pointer, 8))
		}
		emit(v)
		g.UsedResponses[key] = append(g.UsedResponses[key], index)
		must(g.apply(packet.Writes))
		must(g.afterService(site))
		save()
		return reply, true
	}
	pop := pushWallAttachNativeObserver410390(observer)
	defer pop()
	backend := uint64(1)
	entry := reflect.ValueOf(sub410390Native).Pointer()
	if input.Backend == "C" {
		backend = 2
		entry = reflect.ValueOf(wallAttachFixtureCall410390).Pointer()
	} else if input.Backend != "Go" {
		t.Fatal("actual backend")
	}
	rawDefaultDamageSharedRootDispatched4E0B30(g.Header, entry, backend)
	save()
	must(rskWriteJSON4E23C0(filepath.Join(dir, "actual-child-PREEXEC.json"), map[string]any{"pid": os.Getpid(), "parent_pid": os.Getppid(), "backend": input.Backend, "control_word": input.ControlWord, "executable_sha256": sha, "input_sha256": rskHash4E23C0(input.Raw), "BEFORE_sha256": rskHash4E23C0(before), "prepared_before_actual_root": true, "original_observations_passed_to_child": false, "private_local_before_not_original_stack_equivalence": true}))
	if in.Entry != 0x410390 || len(in.Args) != 3 {
		t.Fatal("one root drawable and two raw grid DWORDs")
	}
	drawable, e := g.pointer(in.Args[0])
	must(e)
	var result unsafe.Pointer
	if input.Backend == "C" {
		result = wallAttachFixtureCall410390(drawable, int32(in.Args[1]), int32(in.Args[2]))
	} else {
		result = sub410390Native(drawable, int32(in.Args[1]), int32(in.Args[2]))
	}
	must(trace.Sync())
	must(rskWriteJSON4E23C0(filepath.Join(dir, "actual-return.json"), state(uint64(uintptr(result)))))
	runtime.KeepAlive(g)
}

func waTraceRead410390(path string) ([]waTrace410390, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	dec := json.NewDecoder(io.LimitReader(f, 256<<10+1))
	var out []waTrace410390
	var pending *waTrace410390
	for {
		var v waTrace410390
		e = dec.Decode(&v)
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, e
		}
		if v.Kind == "service" {
			if pending != nil || !v.Complete {
				return nil, fmt.Errorf("service during unfinished read")
			}
			out = append(out, v)
			if len(out) > 1024 {
				return nil, fmt.Errorf("trace bound")
			}
			continue
		}
		if v.Kind != "operand" {
			return nil, fmt.Errorf("actual trace kind")
		}
		if !v.Complete {
			if pending != nil {
				return nil, fmt.Errorf("duplicate read-before")
			}
			copy := v
			pending = &copy
			continue
		}
		if pending == nil {
			return nil, fmt.Errorf("completed read without actual before")
		}
		a, b := *pending, v
		a.Complete = b.Complete
		if !a.Write {
			a.Word = b.Word
			a.Bits = b.Bits
			a.Pointer = b.Pointer
			a.ValueValid = b.ValueValid
		}
		if !reflect.DeepEqual(a, b) {
			return nil, fmt.Errorf("read-before/after identity changed")
		}
		pending = nil
		out = append(out, v)
		if len(out) > 1024 {
			return nil, fmt.Errorf("trace bound")
		}
	}
	if pending != nil {
		if pending.Write {
			return nil, fmt.Errorf("pointer/write fault domain remains deferred")
		}
		pending.Kind = "fault"
		out = append(out, *pending)
	}
	return out, nil
}

type waStep410390 struct {
	Kind     string
	Sequence int
	Event    waEvent410390
	Service  waServiceReference410390
}

func waStaticWidth410390(site uint32) int {
	switch site {
	case 0x4103c4, 0x4103c7, 0x4103ca, 0x4103d1, 0x4103dc:
		return 4
	case 0x4103ce, 0x4103e0, 0x410406, 0x41040d, 0x410416:
		return 1
	}
	return 0
}
func waIdentity410390(meta *waMeta410390, token uint32) (uint64, error) {
	if token == 0 {
		return 0, nil
	}
	p, ok := meta.Identities[strconv.FormatUint(uint64(token), 10)]
	if !ok {
		return 0, fmt.Errorf("original canonical pointer %08x", token)
	}
	return p, nil
}
func waCompare410390(in *waInput410390, row *waReference410390, meta *waMeta410390, trace []waTrace410390, before, after []byte, result *waChildResult410390) error {
	var steps []waStep410390
	for _, e := range row.Events {
		if waStaticWidth410390(e.PC) != 0 && ((e.Kind == "read" || e.Kind == "write") && e.Completed || e.Kind == "fault") {
			kind := "operand"
			if e.Kind == "fault" {
				kind = "fault"
			}
			steps = append(steps, waStep410390{Kind: kind, Sequence: e.Sequence, Event: e})
		}
	}
	for _, s := range row.Services {
		steps = append(steps, waStep410390{Kind: "service", Sequence: s.Sequence, Service: s})
	}
	sort.Slice(steps, func(i, j int) bool { return steps[i].Sequence < steps[j].Sequence })
	if len(steps) != len(trace) {
		return fmt.Errorf("all ordered actual observations %d/%d", len(trace), len(steps))
	}
	var nativeLocal uint64
	for i, step := range steps {
		actual := trace[i]
		if actual.Kind != step.Kind {
			return fmt.Errorf("ordered observation kind %d", i)
		}
		if step.Kind == "service" {
			s := step.Service
			if actual.Site != s.Site || len(actual.Arguments) != len(s.Args) || actual.ResponseIndex != s.ResponseIndex || actual.ReplyWord != s.Packet.Registers["eax"] || actual.ReplyECX != s.Packet.Registers["ecx"] || actual.ReplyEDX != s.Packet.Registers["edx"] {
				return fmt.Errorf("actual service/packet/raw GPR %d", i)
			}
			for j, arg := range actual.Arguments {
				ptr := (s.Site == 0x4103e3 || s.Site == 0x4103fa) && j == 0
				if arg.IsPointer != ptr {
					return fmt.Errorf("actual argument kind")
				}
				if ptr {
					if s.Args[j] != 0x220007f8 || nativeLocal == 0 || arg.Pointer != nativeLocal || actual.LocalAddress != nativeLocal || actual.LocalHex != s.LocalHex || s.LocalAddress != 0x220007f8 {
						return fmt.Errorf("CURRENT same local8 actual snapshot/identity")
					}
				} else if arg.Word != s.Args[j] {
					return fmt.Errorf("raw DWORD service argument")
				}
			}
			p, e := waIdentity410390(meta, s.Packet.Registers["eax"])
			if e != nil {
				return e
			}
			if actual.ReplyPointer != p {
				return fmt.Errorf("full pointer reply identity")
			}
			continue
		}
		e := step.Event
		pw := waStaticWidth410390(e.PC)
		nw := pw
		if e.PC == 0x4103c4 || e.PC == 0x4103c7 || e.PC == 0x4103d1 {
			nw = 8
		}
		if actual.Site != e.PC || actual.PEWidth != pw || actual.Width != nw {
			return fmt.Errorf("static native width/site")
		}
		if step.Kind == "fault" {
			var logical uint32
			switch e.PC {
			case 0x4103c4:
				logical = e.Registers["eax"] + 12
			case 0x4103c7:
				logical = e.Registers["eax"] + 16
			case 0x4103ce:
				logical = e.Registers["esi"] + 4
			case 0x410406:
				logical = e.Registers["eax"] + 130
			default:
				return fmt.Errorf("nonread fault")
			}
			if actual.Write || actual.Complete || actual.ValueValid || actual.PE != logical || e.Address-logical >= uint32(pw) {
				return fmt.Errorf("semantic pending read and original missing span")
			}
			continue
		}
		if !actual.Complete || !actual.ValueValid || actual.PE != e.Address || actual.Write != (e.Kind == "write") {
			return fmt.Errorf("completed operand direction/address")
		}
		var word uint32
		if e.Kind == "read" {
			var value string
			if err := json.Unmarshal(e.Value, &value); err != nil {
				return err
			}
			word = waWord410390(waHex410390(value))
		} else {
			if err := json.Unmarshal(e.Value, &word); err != nil {
				return err
			}
		}
		if actual.IsPointer != (e.PC == 0x4103d1) {
			return fmt.Errorf("pointer store kind")
		}
		if actual.IsPointer {
			p, err := waIdentity410390(meta, word)
			if err != nil {
				return err
			}
			if actual.Pointer != p {
				return fmt.Errorf("full drawable pointer store")
			}
		} else {
			bits := uint64(word)
			if e.PC == 0x4103c4 || e.PC == 0x4103c7 {
				bits = uint64(int64(int32(word)))
			}
			if actual.Word != word || actual.Bits != bits {
				return fmt.Errorf("actual native scalar projection at%x", e.PC)
			}
		}
		if e.PC == 0x4103ca || e.PC == 0x4103dc {
			if actual.Base == 0 || actual.Address != actual.Base+actual.Offset {
				return fmt.Errorf("actual local DWORD store")
			}
			if nativeLocal == 0 {
				nativeLocal = actual.Base
			}
			if nativeLocal != actual.Base {
				return fmt.Errorf("one private local8")
			}
		}
	}
	expected, e := waExpectedAfter410390(meta, before, row.AfterFields)
	if e != nil {
		return e
	}
	if !bytes.Equal(expected, after) {
		return fmt.Errorf("whole actual native banks/canaries versus original field projection")
	}
	if !reflect.DeepEqual(waCanonicalIndices410390(result.UsedReads), waCanonicalIndices410390(row.Used["after_reads"])) || !reflect.DeepEqual(waCanonicalIndices410390(result.UsedServices), waCanonicalIndices410390(row.Used["after_services"])) || !reflect.DeepEqual(result.UsedResponses, row.UsedResponses) {
		return fmt.Errorf("actual independent input consumption")
	}
	if row.Status == "returned" {
		if row.Result == nil {
			return fmt.Errorf("missing actual original pointer return")
		}
		p, err := waIdentity410390(meta, *row.Result)
		if err != nil {
			return err
		}
		if result.Return != p {
			return fmt.Errorf("full polygon return pointer identity")
		}
	}
	return nil
}

func waRun410390(dir, exe, sha, backend string, cw uint16, item *waCase410390) (err error) {
	row, ok := item.References[cw]
	if !ok {
		return fmt.Errorf("one original matching CW")
	}
	if e := os.Mkdir(dir, 0700); e != nil {
		return e
	}
	raw, e := json.Marshal(waChildInput410390{item.Raw, cw, backend, sha})
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
		return fmt.Errorf("input-only child gzip %v %v %v", e, ze, fe)
	}
	stdout, e := os.OpenFile(filepath.Join(dir, "stdout.log"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	defer stdout.Close()
	stderr, e := os.OpenFile(filepath.Join(dir, "stderr.log"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	defer stderr.Close()
	outLog, errLog := &rskBoundedLog4E23C0{File: stdout}, &rskBoundedLog4E23C0{File: stderr}
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, "-test.run=^TestWallAttachNativeWorker410390$", "-test.count=1", "-test.timeout=10s")
	cmd.Env = append(os.Environ(), waWorkerEnv410390+"=1", waDirectoryEnv410390+"="+dir)
	cmd.Stdout = outLog
	cmd.Stderr = errLog
	var forced atomic.Bool
	defaultCancel := cmd.Cancel
	cmd.Cancel = func() error { forced.Store(true); return defaultCancel() }
	start := time.Now().UTC()
	startError := cmd.Start()
	pid := 0
	if cmd.Process != nil {
		pid = cmd.Process.Pid
	}
	var waitError error
	if startError == nil {
		waitError = cmd.Wait()
	}
	exit := -1
	if cmd.ProcessState != nil {
		exit = cmd.ProcessState.ExitCode()
	}
	terminal := map[string]any{"pid": pid, "backend": backend, "control_word": cw, "start_UTC": start, "Wait_terminal_UTC": time.Now().UTC(), "start_error": fmt.Sprint(startError), "Wait_error": fmt.Sprint(waitError), "exit_code": exit, "forced_request": forced.Load(), "context_error": fmt.Sprint(ctx.Err()), "stdout_exceeded": outLog.Exceeded, "stderr_exceeded": errLog.Exceeded, "executable_sha256": sha, "reference_CW_comparisons_planned": 1, "original_execution_credit": 0}
	if e = rskWriteJSON4E23C0(filepath.Join(dir, "actual-child-Wait.json"), terminal); e != nil {
		return e
	}
	defer func() {
		for _, name := range []string{"actual-shared-banks.bin", "actual-native-BEFORE.bin"} {
			path := filepath.Join(dir, name)
			if _, e := os.Lstat(path); e == nil {
				if e = rskGzipOwnedDuplicate4E23C0(path); e != nil && err == nil {
					err = e
				}
			}
		}
	}()
	if startError != nil || forced.Load() || ctx.Err() != nil || outLog.Exceeded || errLog.Exceeded {
		return fmt.Errorf("bounded native child did not terminate naturally: %v %v", startError, waitError)
	}
	var pre struct {
		PID       int    `json:"pid"`
		CW        uint16 `json:"control_word"`
		Backend   string `json:"backend"`
		SHA       string `json:"executable_sha256"`
		InputSHA  string `json:"input_sha256"`
		BeforeSHA string `json:"BEFORE_sha256"`
	}
	p, e := os.ReadFile(filepath.Join(dir, "actual-child-PREEXEC.json"))
	if e != nil {
		return e
	}
	if e = json.Unmarshal(p, &pre); e != nil {
		return e
	}
	if pre.PID != pid || pre.CW != cw || pre.Backend != backend || pre.SHA != sha || pre.InputSHA != rskHash4E23C0(item.Raw) {
		return fmt.Errorf("same-PID PREEXEC/input/main binding")
	}
	var meta waMeta410390
	p, e = os.ReadFile(filepath.Join(dir, "graph.json"))
	if e != nil {
		return e
	}
	if e = json.Unmarshal(p, &meta); e != nil {
		return e
	}
	before, e := os.ReadFile(filepath.Join(dir, "actual-native-BEFORE.bin"))
	if e != nil || rskHash4E23C0(before) != pre.BeforeSHA {
		return fmt.Errorf("whole native BEFORE binding")
	}
	raw, e = os.ReadFile(filepath.Join(dir, "actual-shared-banks.bin"))
	if e != nil {
		return e
	}
	if len(raw) != (len(meta.Banks)+1)*meta.Page {
		return fmt.Errorf("shared native bank geometry")
	}
	header, e := rskReadHeader4E23C0(raw)
	if e != nil {
		return e
	}
	after := raw[meta.Page:]
	if header[0] != rawDefaultDamageSharedMagic4E0B30 || header[1] != rawDefaultDamageSharedVersion4E0B30 || header[36] != 0 || header[38] != 1 || header[40] != (map[string]uint64{"Go": 1, "C": 2}[backend]) {
		return fmt.Errorf("actual generic header/root/validation state")
	}
	var state waChildResult410390
	p, e = os.ReadFile(filepath.Join(dir, "actual-input-state.json"))
	if e != nil {
		return e
	}
	if e = json.Unmarshal(p, &state); e != nil {
		return e
	}
	if state.PID != pid || state.Backend != backend || state.ControlWord != cw {
		return fmt.Errorf("actual durable input consumption PID/backend")
	}
	result := &state
	if exit == 0 {
		var r waChildResult410390
		p, e = os.ReadFile(filepath.Join(dir, "actual-return.json"))
		if e != nil {
			return e
		}
		if e = json.Unmarshal(p, &r); e != nil {
			return e
		}
		if r.PID != pid || r.Backend != backend || r.ControlWord != cw || header[2] != 2 || !reflect.DeepEqual(r.UsedReads, state.UsedReads) || !reflect.DeepEqual(r.UsedServices, state.UsedServices) || !reflect.DeepEqual(r.UsedResponses, state.UsedResponses) {
			return fmt.Errorf("actual natural return/header/state")
		}
		result = &r
	} else if exit == 172 {
		if header[2] != 5 || (header[3] != 10 && header[3] != 11) || header[20] == 0 || header[20] != header[21] || header[22] != 0 || header[24] != 0 || header[10] != 3 || header[6] == 0 || header[8] == 0 || header[8] > 8 || header[5] < header[7] || header[5]-header[7] >= header[8] {
			return fmt.Errorf("actual native read-fault capture prefix")
		}
	} else {
		return fmt.Errorf("unexpected actual exit %d", exit)
	}
	trace, e := waTraceRead410390(filepath.Join(dir, "actual-observations.jsonl"))
	if e != nil {
		return e
	}
	rawTrace, e := os.ReadFile(filepath.Join(dir, "actual-observations.jsonl"))
	if e != nil {
		return e
	}
	pending := 0
	if exit == 172 {
		pending = 1
	}
	if bytes.Count(rawTrace, []byte{'\n'}) != result.TraceEvents+pending {
		return fmt.Errorf("durable observation count plus only actual trailing pending-before")
	}
	if exit == 172 {
		if len(trace) == 0 {
			return fmt.Errorf("missing actual read-before")
		}
		last := trace[len(trace)-1]
		if last.Kind != "fault" || uint64(last.Site) != header[9] || last.Address != header[7] || uint64(last.Width) != header[8] {
			return fmt.Errorf("actual fault and unfinished read match")
		}
	}
	if (row.Status == "returned") != (exit == 0) {
		return fmt.Errorf("original/native terminal domain CW%04x", cw)
	}
	if e = waCompare410390(&item.Input, row, &meta, trace, before, after, result); e != nil {
		return fmt.Errorf("CW%04x: %w", cw, e)
	}
	return rskWriteJSON4E23C0(filepath.Join(dir, "actual-qualification.json"), map[string]any{
		"qualified_selected_observations": true, "actual_pid": pid, "backend": backend, "input": item.Input.Name, "input_sha256": rskHash4E23C0(item.Raw), "control_word": cw, "original_row_sha256": row.RawSHA, "actual_original_CW_comparisons": 1, "actual_native_calls": 1, "native_before_sha256": rskHash4E23C0(before), "native_after_sha256": rskHash4E23C0(after), "whole_foreign_banks_and_canaries": true, "full_native_polygon_return_identity_and_service_pointers": true, "service_reply_ECX_EDX_physical_equivalence": false, "whole_byte_generic_relocation_credit": false, "private_original_stack_FPU_equivalence": false, "physical_nil_EA_equivalence": false, "independent_native_fault_PC_qualification": false, "native_Go_heap_credit": false, "full_helper_body_or_caller_stack_FPU_credit": false, "signal_header": header})
}

func TestWallAttachNativeOriginal410390(t *testing.T) {
	if os.Getenv(waWorkerEnv410390) == "1" {
		return
	}
	cases, e := waLoadCases410390()
	if e != nil {
		t.Fatal(e)
	}
	initial := os.Getenv("OPENNOX_RAW_DEFAULT_DAMAGE_INITIAL_ARTIFACTS")
	if initial == "" {
		t.Fatal("prepared unique INITIAL_ARTIFACTS required")
	}
	root, e := filepath.Abs(filepath.Join(initial, "wall-attach-410390"))
	if e != nil {
		t.Fatal(e)
	}
	if e = os.MkdirAll(filepath.Dir(root), 0700); e != nil {
		t.Fatal(e)
	}
	if e = os.Mkdir(root, 0700); e != nil {
		t.Fatal("unique owned retained directory", e)
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
		for _, cw := range []uint16{0x0e3f, 0x023f} {
			for _, backend := range []string{"Go", "C"} {
				attempted++
				item := &cases[i]
				if t.Run(fmt.Sprintf("%04x/%s/%s", cw, backend, item.Input.Name), func(t *testing.T) {
					if e := waRun410390(filepath.Join(root, fmt.Sprintf("%03d-%04x-%s", i, cw, backend)), exe, sha, backend, cw, item); e != nil {
						t.Fatal(e)
					}
				}) {
					passed++
				}
			}
		}
	}
	if e = rskWriteJSON4E23C0(filepath.Join(root, "actual-corpus-summary.json"), map[string]any{"inputs": 375, "attempted_native_calls": attempted, "passed": passed, "failed": attempted - passed, "stream_complete": attempted == 1500, "qualified_original_CW_comparisons": passed, "actual_native_backends": []string{"Go", "C"}, "independently_selected_control_words": []uint16{0x0e3f, 0x023f}, "input375_sha256": waInputSHA410390, "original750_sha256": waReferenceSHA410390, "whole_graph_projection_and_native_canaries": true, "private_original_stack_or_FPU_equivalence": false, "native_fault_PC_and_physical_nil_EA_equivalence": false, "external_helper_body_credit": false}); e != nil {
		t.Fatal(e)
	}
	if attempted != 1500 || passed != 1500 {
		t.Fatalf("native corpus attempted=%d passed=%d", attempted, passed)
	}
	t.Logf("WALL_ATTACH_NATIVE_ORIGINAL_OK inputs=375 native=1500 matching-original-CW-comparisons=1500 full-polygon-return-and-service-pointer-identities=true private-original-stack-FPU-credit=0")
}
