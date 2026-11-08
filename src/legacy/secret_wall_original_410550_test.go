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

const swInputSHA410550 = "8bfe33a17556423af70a2945102bed8ff5cf46206babc99e35d005e5d83ef921"
const swReferenceSHA410550 = "b417e2adba6fbded1e572d01d588357767c0f9262ea1b1fa5f3d74d5d36d839f"
const swWorkerEnv410550 = "OPENNOX_SECRET_WALL_410550_NATIVE_WORKER"
const swDirectoryEnv410550 = "OPENNOX_SECRET_WALL_410550_NATIVE_DIRECTORY"

type swEvent410550 struct {
	Kind      string            `json:"kind"`
	Sequence  int               `json:"sequence"`
	PC        uint32            `json:"PC"`
	Address   uint32            `json:"address"`
	Width     int               `json:"width"`
	Value     json.RawMessage   `json:"value"`
	Completed bool              `json:"completed"`
	Registers map[string]uint32 `json:"registers"`
}
type swServiceReference410550 struct {
	Site          uint32         `json:"site"`
	Target        uint32         `json:"target"`
	Args          []uint32       `json:"arguments_u32"`
	Sequence      int            `json:"original_call_event_begin"`
	Occurrence    int            `json:"occurrence"`
	ResponseIndex int            `json:"response_index"`
	Packet        swPacket410550 `json:"packet"`
	LocalHex      string         `json:"local_int2_hex"`
	LocalAddress  uint32         `json:"local_int2_address"`
}
type swMemory410550 struct {
	Address uint32 `json:"address"`
	After   string `json:"after"`
	Exact   bool   `json:"replay_exact"`
}
type swReference410550 struct {
	Name            string                     `json:"name"`
	CW              uint16                     `json:"control_word"`
	Entry           uint32                     `json:"entry"`
	InputSHA        string                     `json:"input_sha256"`
	Status          string                     `json:"status"`
	Valid           bool                       `json:"certificate_valid"`
	Result          *uint32                    `json:"actual_EAX_u32"`
	Events          []swEvent410550            `json:"events"`
	Services        []swServiceReference410550 `json:"services"`
	Memory          []swMemory410550           `json:"memory"`
	Used            map[string][]int           `json:"used_mutations"`
	Unused          map[string][]int           `json:"unused_mutations"`
	UsedResponses   map[string][]int           `json:"used_response_indices"`
	UnusedResponses map[string][]int           `json:"unused_response_indices"`
	AfterPages      map[uint32][]byte          `json:"-"`
	FaultClosure    struct {
		PC             uint32 `json:"PC"`
		OperandAddress uint32 `json:"address"`
		OperandWidth   int    `json:"width"`
	} `json:"read_fault_mapping_closure"`
	RawSHA string `json:"-"`
}
type swCase410550 struct {
	Input      swInput410550
	Raw        json.RawMessage
	References map[uint16]*swReference410550
}
type swArgument410550 struct {
	Word      uint32
	Pointer   uint64
	IsPointer bool
}
type swTrace410550 struct {
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
	Arguments                                            []swArgument410550
	ResponseIndex                                        int
	ReplyWord, ReplyECX, ReplyEDX                        uint32
	ReplyPointer                                         uint64
}
type swChildInput410550 struct {
	Raw                    json.RawMessage
	ControlWord            uint16
	Backend, ExecutableSHA string
}
type swChildResult410550 struct {
	PID                     int
	Backend                 string
	ControlWord             uint16
	Return                  uint64
	UsedReads, UsedServices []int
	UsedResponses           map[string][]int
	TraceEvents             int
}

func swCanonicalIndices410550(v []int) []int {
	out := append([]int{}, v...)
	sort.Ints(out)
	for i := 1; i < len(out); i++ {
		if out[i] == out[i-1] {
			panic("duplicate consumed stimulus")
		}
	}
	return out
}
func swResponseIndices410550(values map[string][]int) map[string][]int {
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
func swReadGzipJSON410550(path, sha string, out any) error {
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
func swLoadCases410550() ([]swCase410550, error) {
	base := filepath.Join("testdata", "secret_wall_410550")
	var inputs struct {
		Cases []json.RawMessage `json:"cases"`
		Count int               `json:"case_count"`
		CW    []string          `json:"FPU_controls"`
	}
	if e := swReadGzipJSON410550(filepath.Join(base, "inputs-v1.json.gz"), swInputSHA410550, &inputs); e != nil {
		return nil, e
	}
	if inputs.Count != 273 || len(inputs.Cases) != 273 || !reflect.DeepEqual(inputs.CW, []string{"0e3f", "023f"}) {
		return nil, fmt.Errorf("exact273 independent inputs")
	}
	cases := make([]swCase410550, 273)
	indices := map[string]int{}
	for i, p := range inputs.Cases {
		var b bytes.Buffer
		if e := json.Compact(&b, p); e != nil {
			return nil, e
		}
		cases[i].Raw = append(json.RawMessage(nil), b.Bytes()...)
		cases[i].References = map[uint16]*swReference410550{}
		if e := json.Unmarshal(p, &cases[i].Input); e != nil {
			return nil, e
		}
		name := cases[i].Input.Name
		if _, ok := indices[name]; ok {
			return nil, fmt.Errorf("unique independent name")
		}
		indices[name] = i
		in := &cases[i].Input
		if (in.Entry != 0x410550 && in.Entry != 0x410730 && in.Entry != 0x4107a0) || len(in.Regions) != 5 || len(in.Services) != 8 {
			return nil, fmt.Errorf("three roots/five regions/eight site schedules")
		}
		argc := 1
		if in.Entry == 0x410730 {
			argc = 0
		}
		if len(in.Args) != argc {
			return nil, fmt.Errorf("root argument arity")
		}
	}
	f, z, e := rskOpenGzip4E23C0(filepath.Join(base, "reference-v1.json.gz"), swReferenceSHA410550)
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
			var row swReference410550
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
			if row.Status != "returned" && row.Status != "original_read_fault" {
				return nil, fmt.Errorf("qualified original terminal domain")
			}
			row.RawSHA = rskHash4E23C0(raw)
			row.AfterPages = map[uint32][]byte{}
			expected := map[uint32]bool{0x611000: true, 0x22000000: true}
			for _, r := range item.Input.Regions {
				expected[r.Page] = true
			}
			if len(expected) != 7 || len(row.Memory) != 7 {
				return nil, fmt.Errorf("seven authentic original pages")
			}
			for _, m := range row.Memory {
				data := swHex410550(m.After)
				if !m.Exact || len(data) != 4096 || !expected[m.Address] {
					return nil, fmt.Errorf("source-qualified original full AFTER page")
				}
				if _, ok := row.AfterPages[m.Address]; ok {
					return nil, fmt.Errorf("duplicate original page")
				}
				row.AfterPages[m.Address] = data
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
	if count != 546 || !rowsSeen {
		return nil, fmt.Errorf("exact546 authentic original rows")
	}
	for _, c := range cases {
		if len(c.References) != 2 {
			return nil, fmt.Errorf("two original CW rows per input")
		}
	}
	return cases, nil
}

func TestSecretWallNativeWorker410550(t *testing.T) {
	if os.Getenv(swWorkerEnv410550) != "1" {
		return
	}
	dir := os.Getenv(swDirectoryEnv410550)
	if !filepath.IsAbs(dir) {
		t.Fatal("owned absolute child directory")
	}
	var input swChildInput410550
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
	var in swInput410550
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
	g, e := swBuildGraph410550(&in, filepath.Join(dir, "actual-shared-banks.bin"))
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
	emit := func(v swTrace410550) {
		p, e := json.Marshal(v)
		must(e)
		events++
		traceBytes += len(p) + 1
		if events > 1024 || traceBytes > 256<<10 {
			must(fmt.Errorf("bounded native observation trace"))
		}
		must(enc.Encode(v))
	}
	state := func(value uint64) swChildResult410550 {
		return swChildResult410550{os.Getpid(), input.Backend, input.ControlWord, value, g.UsedReads, g.UsedServices, swResponseIndices410550(g.UsedResponses), events}
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
	observer := &secretWallNativeObserver410550{HeadCell: g.Head}
	observer.Operand = func(op secretWallNativeOperand410550) {
		pe, width, e := swOperand410550(g, op)
		must(e)
		if !op.Complete {
			rawDefaultDamageSharedProgress4E0B30(g.Header, uint64(events), op.Site, op.Address, int(op.Width), map[bool]int{false: rawDefaultDamageSharedRead4E0B30, true: rawDefaultDamageSharedWrite4E0B30}[op.Write], nil)
		}
		emit(swTrace410550{Kind: "operand", Site: op.Site, PE: pe, PEWidth: width, Width: int(op.Width), Complete: op.Complete, ValueValid: op.ValueValid, AddressValid: op.AddressValid, Write: op.Write, IsPointer: op.IsPointer, Address: uint64(uintptr(op.Address)), Base: uint64(uintptr(op.Base)), Offset: uint64(op.Offset), Word: op.ValueWord, Bits: uint64(op.ValueWord), Pointer: uint64(uintptr(op.ValuePointer))})
		if op.Complete {
			if !op.Write {
				must(g.afterRead(op.Site, pe, width))
			}
			save()
		}
	}
	observer.Service = func(site uint32, args []secretWallNativeArg410550) (secretWallNativeReply410550, bool) {
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
		reply := secretWallNativeReply410550{Word: packet.Registers["eax"], ECX: packet.Registers["ecx"], EDX: packet.Registers["edx"]}
		if swPointerSite410550(site) {
			reply.Pointer, e = g.pointer(packet.Registers["eax"])
			must(e)
		}
		v := swTrace410550{Kind: "service", Site: site, Complete: true, ResponseIndex: index, ReplyWord: reply.Word, ReplyECX: reply.ECX, ReplyEDX: reply.EDX, ReplyPointer: uint64(uintptr(reply.Pointer))}
		for _, a := range args {
			v.Arguments = append(v.Arguments, swArgument410550{a.Word, uint64(uintptr(a.Pointer)), a.IsPointer})
		}

		emit(v)
		g.UsedResponses[key] = append(g.UsedResponses[key], index)
		must(g.apply(packet.Writes))
		must(g.afterService(site))
		save()
		return reply, true
	}
	pop := pushSecretWallNativeObserver410550(observer)
	defer pop()
	backend := uint64(1)
	var entry uintptr
	switch in.Entry {
	case 0x410550:
		entry = reflect.ValueOf(secretWallLookupNative410550).Pointer()
	case 0x410730:
		entry = reflect.ValueOf(secretWallClearNative410730).Pointer()
	case 0x4107a0:
		entry = reflect.ValueOf(secretWallRemoveNative4107A0).Pointer()
	default:
		t.Fatal("one selected SecretWall root")
	}
	if input.Backend == "C" {
		backend = 2
		entry = reflect.ValueOf(secretWallFixtureCall410550).Pointer()
	} else if input.Backend != "Go" {
		t.Fatal("actual backend")
	}
	rawDefaultDamageSharedRootDispatched4E0B30(g.Header, entry, backend)
	save()
	must(rskWriteJSON4E23C0(filepath.Join(dir, "actual-child-PREEXEC.json"), map[string]any{"pid": os.Getpid(), "parent_pid": os.Getppid(), "backend": input.Backend, "control_word": input.ControlWord, "executable_sha256": sha, "input_sha256": rskHash4E23C0(input.Raw), "BEFORE_sha256": rskHash4E23C0(before), "prepared_before_actual_root": true, "original_observations_passed_to_child": false, "private_argument_cells_not_original_stack_equivalence": true}))
	argc := 1
	if in.Entry == 0x410730 {
		argc = 0
	}
	if len(in.Args) != argc {
		t.Fatal("actual selected root arguments")
	}
	var argPointer unsafe.Pointer
	var argWord uint32
	if argc == 1 {
		argWord = in.Args[0]
	}
	if in.Entry == 0x4107a0 {
		argPointer, e = g.pointer(in.Args[0])
		must(e)
	}
	var result unsafe.Pointer
	if input.Backend == "C" {
		result = secretWallFixtureCall410550(in.Entry, argWord, argPointer)
	} else {
		switch in.Entry {
		case 0x410550:
			result = secretWallLookupNative410550(argWord)
		case 0x410730:
			result = secretWallClearNative410730()
		case 0x4107a0:
			result = secretWallRemoveNative4107A0(argPointer)
		}
	}
	must(trace.Sync())
	must(rskWriteJSON4E23C0(filepath.Join(dir, "actual-return.json"), state(uint64(uintptr(result)))))
	runtime.KeepAlive(g)
}

func swTraceRead410550(path string) ([]swTrace410550, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	dec := json.NewDecoder(io.LimitReader(f, 256<<10+1))
	var out []swTrace410550
	var pending *swTrace410550
	for {
		var v swTrace410550
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

type swStep410550 struct {
	Kind     string
	Sequence int
	Event    swEvent410550
	Service  swServiceReference410550
}

func swStaticWidth410550(site uint32) int {
	switch site {
	case 0x41055a, 0x410562:
		return 2
	case 0x41055f, 0x410579, 0x410730, 0x41073a, 0x41074b, 0x410753, 0x4107a0, 0x4107ad, 0x4107c7, 0x4107da, 0x4107f3:
		return 4
	}
	return 0
}

func swIdentity410550(meta *swMeta410550, token uint32) (uint64, error) {
	if token == 0 {
		return 0, nil
	}
	p, ok := meta.Identities[strconv.FormatUint(uint64(token), 10)]
	if !ok {
		return 0, fmt.Errorf("original canonical pointer %08x", token)
	}
	return p, nil
}
func swCompare410550(in *swInput410550, row *swReference410550, meta *swMeta410550, trace []swTrace410550, before, after []byte, result *swChildResult410550) error {
	var steps []swStep410550
	for _, e := range row.Events {
		if swStaticWidth410550(e.PC) != 0 && (((e.Kind == "read" || e.Kind == "write") && e.Completed) || e.Kind == "fault") {
			kind := "operand"
			if e.Kind == "fault" {
				kind = "fault"
			}
			steps = append(steps, swStep410550{Kind: kind, Sequence: e.Sequence, Event: e})
		}
	}
	for _, service := range row.Services {
		steps = append(steps, swStep410550{Kind: "service", Sequence: service.Sequence, Service: service})
	}
	sort.Slice(steps, func(i, j int) bool { return steps[i].Sequence < steps[j].Sequence })
	if len(steps) != len(trace) {
		return fmt.Errorf("all ordered actual observations %d/%d", len(trace), len(steps))
	}
	originalFree, nativeFree := 0, 0
	for i, step := range steps {
		actual := trace[i]
		if actual.Kind != step.Kind {
			return fmt.Errorf("ordered observation kind %d", i)
		}
		if step.Kind == "service" {
			service := step.Service
			if actual.Site != service.Site || len(actual.Arguments) != len(service.Args) || actual.ResponseIndex != service.ResponseIndex || actual.ReplyWord != service.Packet.Registers["eax"] || actual.ReplyECX != service.Packet.Registers["ecx"] || actual.ReplyEDX != service.Packet.Registers["edx"] {
				return fmt.Errorf("actual service/packet/raw GPR %d", i)
			}
			for j, arg := range actual.Arguments {
				if !arg.IsPointer || arg.Word != 0 {
					return fmt.Errorf("all original list service arguments are full pointers")
				}
				p, e := swIdentity410550(meta, service.Args[j])
				if e != nil {
					return e
				}
				if arg.Pointer != p {
					return fmt.Errorf("full service pointer identity")
				}
			}
			if swPointerSite410550(service.Site) {
				p, e := swIdentity410550(meta, service.Packet.Registers["eax"])
				if e != nil {
					return e
				}
				if actual.ReplyPointer != p {
					return fmt.Errorf("full First/Next reply identity")
				}
			} else {
				originalFree++
				nativeFree++
				if actual.ReplyPointer != 0 {
					return fmt.Errorf("Free DWORD must not manufacture pointer")
				}
			}
			continue
		}
		e := step.Event
		pw := swStaticWidth410550(e.PC)
		nw := pw
		pointer := pw == 4
		if pointer {
			nw = 8
		}
		if actual.Site != e.PC || actual.PEWidth != pw || actual.Width != nw || actual.IsPointer != pointer {
			return fmt.Errorf("static native width/site/kind")
		}
		if step.Kind == "fault" {
			if actual.Write || actual.Complete || actual.ValueValid || row.FaultClosure.PC != e.PC || row.FaultClosure.OperandWidth != pw || actual.PE != row.FaultClosure.OperandAddress {
				return fmt.Errorf("original pending read semantic span versus actual attempt")
			}
			continue
		}
		if !actual.Complete || !actual.ValueValid || !actual.AddressValid || actual.PE != e.Address || actual.Write != (e.Kind == "write") {
			return fmt.Errorf("completed operand direction/address")
		}
		var word uint32
		if e.Kind == "read" {
			var value string
			if err := json.Unmarshal(e.Value, &value); err != nil {
				return err
			}
			word = swWord410550(swHex410550(value))
		} else {
			if err := json.Unmarshal(e.Value, &word); err != nil {
				return err
			}
		}
		if pointer {
			p, err := swIdentity410550(meta, word)
			if err != nil {
				return err
			}
			if actual.Pointer != p || actual.Word != 0 || actual.Bits != 0 {
				return fmt.Errorf("full native pointer operand %x", e.PC)
			}
		} else if actual.Word != word || actual.Bits != uint64(word) {
			return fmt.Errorf("actual unsigned WORD bits at%x", e.PC)
		}
		if e.PC == 0x41055a || e.PC == 0x4107ad {
			if actual.Base == 0 || actual.Address != actual.Base || actual.Offset != 0 {
				return fmt.Errorf("actual lazy argument cell; not original stack equivalence")
			}
		}
	}
	fields := map[uint32][]byte{}
	for _, field := range meta.Fields {
		if field.Accessible == 0 {
			continue
		}
		data, e := swOriginalBytes410550(row.AfterPages, field.PE, field.Accessible)
		if e != nil {
			return e
		}
		if old, ok := fields[field.PE]; ok && !bytes.Equal(old, data) {
			return fmt.Errorf("aliased authentic AFTER field")
		}
		fields[field.PE] = data
	}
	expected, e := swExpectedAfter410550(meta, before, fields)
	if e != nil {
		return e
	}
	if !bytes.Equal(expected, after) {
		return fmt.Errorf("whole actual six native banks/canaries versus authentic original field projection")
	}
	if !reflect.DeepEqual(swCanonicalIndices410550(result.UsedReads), swCanonicalIndices410550(row.Used["after_reads"])) || !reflect.DeepEqual(swCanonicalIndices410550(result.UsedServices), swCanonicalIndices410550(row.Used["after_services"])) || !reflect.DeepEqual(result.UsedResponses, row.UsedResponses) {
		return fmt.Errorf("actual independent input consumption")
	}
	// Native unused accounting is derived only from actual consumed input
	// indices, not from a predicted caller branch or expected native schedule.
	for site, schedule := range in.Services {
		seen := map[int]bool{}
		for _, i := range result.UsedResponses[site] {
			if i < 0 || i >= len(schedule.Responses) {
				return fmt.Errorf("actual response index domain")
			}
			seen[i] = true
		}
		unused := []int{}
		for i := range schedule.Responses {
			if !seen[i] {
				unused = append(unused, i)
			}
		}
		if !reflect.DeepEqual(unused, row.UnusedResponses[site]) {
			return fmt.Errorf("actual unused service response accounting")
		}
	}
	for kind, size := range map[string]int{"after_reads": len(in.AfterReads), "after_services": len(in.AfterServices)} {
		consumed := result.UsedReads
		if kind == "after_services" {
			consumed = result.UsedServices
		}
		seen := map[int]bool{}
		for _, i := range consumed {
			if i < 0 || i >= size {
				return fmt.Errorf("actual stimulus index domain")
			}
			seen[i] = true
		}
		unused := []int{}
		for i := 0; i < size; i++ {
			if !seen[i] {
				unused = append(unused, i)
			}
		}
		if !reflect.DeepEqual(unused, row.Unused[kind]) {
			return fmt.Errorf("actual unused stimulus accounting")
		}
	}
	if row.Status == "returned" {
		if row.Result == nil {
			return fmt.Errorf("missing actual original raw return")
		}
		if in.Entry == 0x4107a0 && originalFree > 0 {
			if nativeFree != originalFree || len(in.Args) != 1 {
				return fmt.Errorf("actual post-Free policy domain")
			}
			p, e := swIdentity410550(meta, in.Args[0])
			if e != nil {
				return e
			}
			if result.Return != p {
				return fmt.Errorf("current typed C cached matched native return contract")
			}
			// Original post-Free EAX remains raw diagnostic; no equality or
			// uintptr-to-pointer construction follows from that DWORD.
		} else {
			p, e := swIdentity410550(meta, *row.Result)
			if e != nil {
				return e
			}
			if result.Return != p {
				return fmt.Errorf("full qualified returned pointer identity")
			}
			if in.Entry == 0x410730 || in.Entry == 0x4107a0 {
				if *row.Result != 0 || result.Return != 0 {
					return fmt.Errorf("natural clear/no-Free remove nil return")
				}
			}
		}
	}
	return nil
}

func swRun410550(dir, exe, sha, backend string, cw uint16, item *swCase410550) (err error) {
	row, ok := item.References[cw]
	if !ok {
		return fmt.Errorf("one original matching CW")
	}
	if e := os.Mkdir(dir, 0700); e != nil {
		return e
	}
	raw, e := json.Marshal(swChildInput410550{item.Raw, cw, backend, sha})
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
	cmd := exec.CommandContext(ctx, exe, "-test.run=^TestSecretWallNativeWorker410550$", "-test.count=1", "-test.timeout=10s")
	cmd.Env = append(os.Environ(), swWorkerEnv410550+"=1", swDirectoryEnv410550+"="+dir)
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
	var meta swMeta410550
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
	var state swChildResult410550
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
		var r swChildResult410550
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
	trace, e := swTraceRead410550(filepath.Join(dir, "actual-observations.jsonl"))
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
	if e = swCompare410550(&item.Input, row, &meta, trace, before, after, result); e != nil {
		return fmt.Errorf("CW%04x: %w", cw, e)
	}
	postFreeReturn := false
	if exit == 0 && item.Input.Entry == 0x4107a0 {
		for _, service := range row.Services {
			if service.Site == 0x4107e0 || service.Site == 0x4107f6 {
				postFreeReturn = true
			}
		}
	}
	return rskWriteJSON4E23C0(filepath.Join(dir, "actual-qualification.json"), map[string]any{
		"qualified_selected_observations": true, "actual_pid": pid, "backend": backend, "input": item.Input.Name, "input_sha256": rskHash4E23C0(item.Raw), "control_word": cw, "original_row_sha256": row.RawSHA, "actual_original_CW_comparisons": 1, "actual_native_calls": 1, "native_before_sha256": rskHash4E23C0(before), "native_after_sha256": rskHash4E23C0(after), "whole_foreign_banks_and_canaries": true, "full_native_secret_wall_and_service_pointer_identities": true, "post_Free_original_EAX_native_pointer_equivalence": false, "post_Free_cached_native_return_policy_checked": postFreeReturn, "original_return_comparison_qualified": exit == 0 && !postFreeReturn, "actual_native_returned": exit == 0, "unused_responses_and_stimuli_accounted": true, "service_reply_ECX_EDX_physical_equivalence": false, "whole_byte_generic_relocation_credit": false, "private_original_stack_FPU_equivalence": false, "physical_nil_EA_equivalence": false, "independent_native_fault_PC_qualification": false, "native_Go_heap_credit": false, "full_helper_body_or_caller_stack_FPU_credit": false, "signal_header": header})
}

func TestSecretWallNativeOriginal410550(t *testing.T) {
	if os.Getenv(swWorkerEnv410550) == "1" {
		return
	}
	cases, e := swLoadCases410550()
	if e != nil {
		t.Fatal(e)
	}
	initial := os.Getenv("OPENNOX_RAW_DEFAULT_DAMAGE_INITIAL_ARTIFACTS")
	if initial == "" {
		t.Fatal("prepared unique INITIAL_ARTIFACTS required")
	}
	root, e := filepath.Abs(filepath.Join(initial, "secret-wall-three-410550"))
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
					if e := swRun410550(filepath.Join(root, fmt.Sprintf("%03d-%04x-%s", i, cw, backend)), exe, sha, backend, cw, item); e != nil {
						t.Fatal(e)
					}
				}) {
					passed++
				}
			}
		}
	}
	if e = rskWriteJSON4E23C0(filepath.Join(root, "actual-corpus-summary.json"), map[string]any{"inputs": 273, "attempted_native_calls": attempted, "passed": passed, "failed": attempted - passed, "stream_complete": attempted == 1092, "qualified_original_CW_comparisons": passed, "actual_native_backends": []string{"Go", "C"}, "independently_selected_control_words": []uint16{0x0e3f, 0x023f}, "input273_sha256": swInputSHA410550, "original546_sha256": swReferenceSHA410550, "whole_graph_projection_and_native_canaries": true, "private_original_stack_or_FPU_equivalence": false, "native_fault_PC_and_physical_nil_EA_equivalence": false, "external_helper_body_credit": false}); e != nil {
		t.Fatal(e)
	}
	if attempted != 1092 || passed != 1092 {
		t.Fatalf("native corpus attempted=%d passed=%d", attempted, passed)
	}
	t.Logf("SECRET_WALL_NATIVE_ORIGINAL_OK inputs=273 native=1092 matching-original-CW-comparisons=1092 full-native-pointer-and-service-identities=true post-Free-original-EAX-return-credit=0 private-original-stack-FPU-credit=0")
}
