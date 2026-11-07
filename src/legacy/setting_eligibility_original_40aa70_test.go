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
)

const seInputSHA40AA70 = "808b1c79535399ce4d08db7160dfd1f9453b86ac368314e3f777a15fbe42eb18"
const seReferenceSHA40AA70 = "03cad4011f68d89fe502ed93967281b994d5c623d525c85b0bb21dc4f5836ec2"
const seWorkerEnv40AA70 = "OPENNOX_SETTING_ELIGIBILITY_NATIVE_WORKER"
const seDirectoryEnv40AA70 = "OPENNOX_SETTING_ELIGIBILITY_NATIVE_DIRECTORY"

type seEvent40AA70 struct {
	Kind      string            `json:"kind"`
	Sequence  int               `json:"sequence"`
	PC        uint32            `json:"PC"`
	Address   uint32            `json:"address"`
	Width     int               `json:"width"`
	Value     json.RawMessage   `json:"value"`
	Completed bool              `json:"completed"`
	Registers map[string]uint32 `json:"registers"`
}
type seServiceReference40AA70 struct {
	Site          uint32         `json:"site"`
	Target        uint32         `json:"target"`
	Args          []uint32       `json:"arguments_u32"`
	Sequence      int            `json:"original_call_event_begin"`
	Occurrence    int            `json:"occurrence"`
	ResponseIndex int            `json:"response_index"`
	Packet        sePacket40AA70 `json:"packet"`
}
type seMemory40AA70 struct {
	Address uint32 `json:"address"`
	After   string `json:"after"`
	Exact   bool   `json:"replay_exact"`
}
type seReference40AA70 struct {
	Name            string                     `json:"name"`
	CW              uint16                     `json:"control_word"`
	Entry           uint32                     `json:"entry"`
	InputSHA        string                     `json:"input_sha256"`
	Status          string                     `json:"status"`
	Valid           bool                       `json:"certificate_valid"`
	Result          *uint32                    `json:"actual_EAX_u32"`
	Events          []seEvent40AA70            `json:"events"`
	Services        []seServiceReference40AA70 `json:"services"`
	Memory          []seMemory40AA70           `json:"memory"`
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
type seCase40AA70 struct {
	Input      seInput40AA70
	Raw        json.RawMessage
	References map[uint16]*seReference40AA70
}
type seArgument40AA70 struct {
	Word      uint32
	Pointer   uint64
	IsPointer bool
}
type seTrace40AA70 struct {
	Kind                               string
	Site, PE                           uint32
	PEWidth, Width                     int
	Complete, ValueValid, AddressValid bool
	Address, Base, Pointer             uint64
	Offset                             uint64
	Word                               uint32
	Arguments                          []seArgument40AA70
	ResponseIndex                      int
	ReplyWord, ReplyECX, ReplyEDX      uint32
	ReplyPointer                       uint64
}
type seChildInput40AA70 struct {
	Raw                    json.RawMessage
	ControlWord            uint16
	Backend, ExecutableSHA string
}
type seChildResult40AA70 struct {
	PID                     int
	Backend                 string
	ControlWord             uint16
	Return                  uint32
	UsedReads, UsedServices []int
	UsedResponses           map[string][]int
	TraceEvents             int
}

func seCanonicalIndices40AA70(v []int) []int {
	out := append([]int{}, v...)
	sort.Ints(out)
	for i := 1; i < len(out); i++ {
		if out[i] == out[i-1] {
			panic("duplicate consumed stimulus")
		}
	}
	return out
}
func seResponseIndices40AA70(values map[string][]int) map[string][]int {
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
func seReadGzipJSON40AA70(path, sha string, out any) error {
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
func seLoadCases40AA70() ([]seCase40AA70, error) {
	base := filepath.Join("testdata", "setting_eligibility_40aa70")
	var inputs struct {
		Cases []json.RawMessage `json:"cases"`
		Count int               `json:"case_count"`
		CW    []uint16          `json:"FPU_controls"`
	}
	if e := seReadGzipJSON40AA70(filepath.Join(base, "inputs-v1.json.gz"), seInputSHA40AA70, &inputs); e != nil {
		return nil, e
	}
	if inputs.Count != 662 || len(inputs.Cases) != 662 || !reflect.DeepEqual(inputs.CW, []uint16{0x0e3f, 0x023f}) {
		return nil, fmt.Errorf("exact662 independent inputs")
	}
	cases := make([]seCase40AA70, 662)
	indices := map[string]int{}
	for i, p := range inputs.Cases {
		var b bytes.Buffer
		if e := json.Compact(&b, p); e != nil {
			return nil, e
		}
		cases[i].Raw = append(json.RawMessage(nil), b.Bytes()...)
		cases[i].References = map[uint16]*seReference40AA70{}
		if e := json.Unmarshal(p, &cases[i].Input); e != nil {
			return nil, e
		}
		name := cases[i].Input.Name
		if _, ok := indices[name]; ok {
			return nil, fmt.Errorf("unique independent name")
		}
		indices[name] = i
	}
	f, z, e := rskOpenGzip4E23C0(filepath.Join(base, "reference-v1.json.gz"), seReferenceSHA40AA70)
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
			var row seReference40AA70
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
				data := seHex40AA70(m.After)
				if !m.Exact || len(data) != 4096 {
					return nil, fmt.Errorf("source-qualified original full AFTER page")
				}
				pages[m.Address] = data
			}
			row.AfterFields = map[uint32][]byte{}
			add := func(pe uint32, width int) error {
				available := seAvailable40AA70(&item.Input, pe, width)
				if available == 0 {
					return nil
				}
				p, e := seOriginalBytes40AA70(pages, pe, available)
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
				case len(role) >= 4 && role[:4] == "info":
					e = add(t+52, 1)
				case role == "argument":
					e = add(t+3680, 4)
					if e == nil {
						e = add(t+2068, 4)
					}
				case len(role) >= 8 && role[:8] == "iterator":
					e = add(t+2140, 4)
				}
				if e != nil {
					return nil, e
				}
			}
			// Global is always mapped in independently chosen original input.
			value, e := seOriginalBytes40AA70(pages, 0x5d5348, 4)
			if e != nil {
				return nil, e
			}
			row.AfterFields[0x5d5348] = value
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
	if count != 1324 || !rowsSeen {
		return nil, fmt.Errorf("exact1324 authentic original rows")
	}
	for _, c := range cases {
		if len(c.References) != 2 {
			return nil, fmt.Errorf("two original CW rows per input")
		}
	}
	return cases, nil
}

// No original reference or predicted observation schedule enters this worker.
func TestSettingEligibilityNativeWorker40AA70(t *testing.T) {
	if os.Getenv(seWorkerEnv40AA70) != "1" {
		return
	}
	dir := os.Getenv(seDirectoryEnv40AA70)
	if !filepath.IsAbs(dir) {
		t.Fatal("owned absolute child directory")
	}
	var input seChildInput40AA70
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
	var in seInput40AA70
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
	g, e := seBuildGraph40AA70(&in, filepath.Join(dir, "actual-shared-banks.bin"))
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
	emit := func(v seTrace40AA70) {
		p, e := json.Marshal(v)
		must(e)
		events++
		traceBytes += len(p) + 1
		if events > 1024 || traceBytes > 256<<10 {
			must(fmt.Errorf("bounded native observation trace"))
		}
		must(enc.Encode(v))
	}
	state := func(value uint32) seChildResult40AA70 {
		return seChildResult40AA70{os.Getpid(), input.Backend, input.ControlWord, value, g.UsedReads, g.UsedServices, seResponseIndices40AA70(g.UsedResponses), events}
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
	observer := &settingEligibilityNativeObserver40AA70{GlobalCell: g.globalCell()}
	observer.Operand = func(op settingEligibilityNativeOperand40AA70) {
		pe, width, e := seOperand40AA70(g, op)
		must(e)
		if !op.Complete {
			rawDefaultDamageSharedProgress4E0B30(g.Header, uint64(events), op.Site, op.Address, int(op.Width), rawDefaultDamageSharedRead4E0B30, nil)
		}
		emit(seTrace40AA70{Kind: "operand", Site: op.Site, PE: pe, PEWidth: width, Width: int(op.Width), Complete: op.Complete, ValueValid: op.ValueValid, AddressValid: op.AddressValid, Address: uint64(uintptr(op.Address)), Base: uint64(uintptr(op.Base)), Offset: uint64(op.Offset), Word: op.ValueWord})
		if op.Complete {
			must(g.afterRead(op.Site, pe, width))
			save()
		}
	}
	observer.Service = func(site uint32, args []settingEligibilityNativeArg40AA70) (settingEligibilityNativeReply40AA70, bool) {
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
		reply := settingEligibilityNativeReply40AA70{Word: packet.Registers["eax"], ECX: packet.Registers["ecx"], EDX: packet.Registers["edx"]}
		if sePointerSite40AA70(site) {
			reply.Pointer, e = g.pointer(packet.Registers["eax"])
			must(e)
		}
		v := seTrace40AA70{Kind: "service", Site: site, Complete: true, ResponseIndex: index, ReplyWord: reply.Word, ReplyECX: reply.ECX, ReplyEDX: reply.EDX, ReplyPointer: uint64(uintptr(reply.Pointer))}
		for _, a := range args {
			v.Arguments = append(v.Arguments, seArgument40AA70{a.Word, uint64(uintptr(a.Pointer)), a.IsPointer})
		}
		emit(v)
		g.UsedResponses[key] = append(g.UsedResponses[key], index)
		must(g.apply(packet.Writes))
		must(g.afterService(site))
		save()
		return reply, true
	}
	pop := pushSettingEligibilityNativeObserver40AA70(observer)
	defer pop()
	backend := uint64(1)
	entry := reflect.ValueOf(sub40AA70Native).Pointer()
	if input.Backend == "C" {
		backend = 2
		entry = reflect.ValueOf(settingEligibilityFixtureCall40AA70).Pointer()
	} else if input.Backend != "Go" {
		t.Fatal("actual backend")
	}
	rawDefaultDamageSharedRootDispatched4E0B30(g.Header, entry, backend)
	save()
	must(rskWriteJSON4E23C0(filepath.Join(dir, "actual-child-PREEXEC.json"), map[string]any{"pid": os.Getpid(), "parent_pid": os.Getppid(), "backend": input.Backend, "control_word": input.ControlWord, "executable_sha256": sha, "input_sha256": rskHash4E23C0(input.Raw), "BEFORE_sha256": rskHash4E23C0(before), "prepared_before_actual_root": true, "original_observations_passed_to_child": false, "diagnostic_global_cell_is_foreign": true}))
	if in.Entry != 0x40aa70 || len(in.Args) != 1 {
		t.Fatal("one AA70 root pointer argument")
	}
	player, e := g.pointer(in.Args[0])
	must(e)
	var result int32
	if input.Backend == "C" {
		result = settingEligibilityFixtureCall40AA70(player)
	} else {
		result = sub40AA70Native(player)
	}
	must(trace.Sync())
	must(rskWriteJSON4E23C0(filepath.Join(dir, "actual-return.json"), state(uint32(result))))
	runtime.KeepAlive(g)
}

func seTraceRead40AA70(path string) ([]seTrace40AA70, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	dec := json.NewDecoder(io.LimitReader(f, 256<<10+1))
	var out []seTrace40AA70
	var pending *seTrace40AA70
	for {
		var v seTrace40AA70
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
		a.Word = b.Word
		a.ValueValid = b.ValueValid
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
		pending.Kind = "fault"
		out = append(out, *pending)
	}
	return out, nil
}

type seStep40AA70 struct {
	Kind     string
	Sequence int
	Event    seEvent40AA70
	Service  seServiceReference40AA70
}

func seStaticWidth40AA70(site uint32) int {
	switch site {
	case 0x40aafd:
		return 1
	case 0x40aaa8, 0x40aab3, 0x40aadf, 0x40ab9d:
		return 4
	}
	return 0
}
func seIdentity40AA70(meta *seMeta40AA70, token uint32) (uint64, error) {
	if token == 0 {
		return 0, nil
	}
	p, ok := meta.Identities[strconv.FormatUint(uint64(token), 10)]
	if !ok {
		return 0, fmt.Errorf("original canonical pointer %08x", token)
	}
	return p, nil
}
func seCompare40AA70(in *seInput40AA70, row *seReference40AA70, meta *seMeta40AA70, trace []seTrace40AA70, before, after []byte, result *seChildResult40AA70) error {
	var steps []seStep40AA70
	for _, e := range row.Events {
		if seStaticWidth40AA70(e.PC) != 0 && ((e.Kind == "read" && e.Completed) || e.Kind == "fault") {
			kind := "operand"
			if e.Kind == "fault" {
				kind = "fault"
			}
			steps = append(steps, seStep40AA70{Kind: kind, Sequence: e.Sequence, Event: e})
		}
	}
	for _, s := range row.Services {
		steps = append(steps, seStep40AA70{Kind: "service", Sequence: s.Sequence, Service: s})
	}
	sort.Slice(steps, func(i, j int) bool { return steps[i].Sequence < steps[j].Sequence })
	if len(steps) != len(trace) {
		return fmt.Errorf("all ordered actual observations %d/%d", len(trace), len(steps))
	}
	for i, step := range steps {
		actual := trace[i]
		if actual.Kind != step.Kind {
			return fmt.Errorf("ordered actual kind %d", i)
		}
		if step.Kind == "service" {
			service := step.Service
			if actual.Site != service.Site || len(actual.Arguments) != len(service.Args) || actual.ResponseIndex != service.ResponseIndex || actual.ReplyWord != service.Packet.Registers["eax"] || actual.ReplyECX != service.Packet.Registers["ecx"] || actual.ReplyEDX != service.Packet.Registers["edx"] {
				return fmt.Errorf("original actual service/response/rawGPR %d", i)
			}
			for j, arg := range actual.Arguments {
				pointer := service.Site == 0x40aba8
				if arg.IsPointer != pointer {
					return fmt.Errorf("actual service argument type")
				}
				if pointer {
					p, e := seIdentity40AA70(meta, service.Args[j])
					if e != nil {
						return e
					}
					if arg.Pointer != p {
						return fmt.Errorf("cached full native service pointer")
					}
				} else if arg.Word != service.Args[j] {
					return fmt.Errorf("full DWORD service argument")
				}
			}
			if sePointerSite40AA70(service.Site) {
				p, e := seIdentity40AA70(meta, service.Packet.Registers["eax"])
				if e != nil {
					return e
				}
				if actual.ReplyPointer != p {
					return fmt.Errorf("full actual pointer reply")
				}
			}
			continue
		}
		event := step.Event
		if actual.Site != event.PC || actual.PEWidth != seStaticWidth40AA70(event.PC) || actual.Width != actual.PEWidth {
			return fmt.Errorf("static native scalar site/width")
		}
		if step.Kind == "fault" {
			var logical uint32
			switch event.PC {
			case 0x40aafd:
				logical = event.Registers["edi"] + 52
			case 0x40aaa8:
				logical = event.Registers["esi"] + 3680
			case 0x40aadf:
				logical = event.Registers["esi"] + 2068
			case 0x40ab9d:
				logical = event.Registers["eax"] + 2140
			case 0x40aab3:
				logical = 0x5d5348
			}
			if actual.Complete || actual.ValueValid || actual.PE != logical || event.Address-logical >= uint32(actual.PEWidth) {
				return fmt.Errorf("semantic pending read and raw original missing span")
			}
			continue
		}
		if !actual.Complete || !actual.ValueValid || actual.PE != event.Address {
			return fmt.Errorf("actual completed scalar operand")
		}
		var value string
		if e := json.Unmarshal(event.Value, &value); e != nil {
			return e
		}
		if actual.Word != seWord40AA70(seHex40AA70(value)) {
			return fmt.Errorf("actual BYTE/DWORD value versus original read")
		}
	}
	expected, e := seExpectedAfter40AA70(meta, before, row.AfterFields)
	if e != nil {
		return e
	}
	if !bytes.Equal(expected, after) {
		return fmt.Errorf("whole actual native banks/canaries versus original fields")
	}
	if !reflect.DeepEqual(seCanonicalIndices40AA70(result.UsedReads), seCanonicalIndices40AA70(row.Used["after_reads"])) || !reflect.DeepEqual(seCanonicalIndices40AA70(result.UsedServices), seCanonicalIndices40AA70(row.Used["after_services"])) || !reflect.DeepEqual(result.UsedResponses, row.UsedResponses) {
		return fmt.Errorf("actual independent input consumption")
	}
	if row.Status == "returned" && (row.Result == nil || result.Return != *row.Result) {
		return fmt.Errorf("full raw original EAX versus actual int32 return")
	}
	return nil
}

func seRun40AA70(dir, exe, sha, backend string, cw uint16, item *seCase40AA70) (err error) {
	row, ok := item.References[cw]
	if !ok {
		return fmt.Errorf("one original matching CW")
	}
	if e := os.Mkdir(dir, 0700); e != nil {
		return e
	}
	raw, e := json.Marshal(seChildInput40AA70{item.Raw, cw, backend, sha})
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
	cmd := exec.CommandContext(ctx, exe, "-test.run=^TestSettingEligibilityNativeWorker40AA70$", "-test.count=1", "-test.timeout=10s")
	cmd.Env = append(os.Environ(), seWorkerEnv40AA70+"=1", seDirectoryEnv40AA70+"="+dir)
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
	var meta seMeta40AA70
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
	var state seChildResult40AA70
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
		var r seChildResult40AA70
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
	trace, e := seTraceRead40AA70(filepath.Join(dir, "actual-observations.jsonl"))
	if e != nil {
		return e
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
	if e = seCompare40AA70(&item.Input, row, &meta, trace, before, after, result); e != nil {
		return fmt.Errorf("CW%04x: %w", cw, e)
	}
	return rskWriteJSON4E23C0(filepath.Join(dir, "actual-qualification.json"), map[string]any{
		"qualified_selected_observations": true, "actual_pid": pid, "backend": backend, "input": item.Input.Name, "input_sha256": rskHash4E23C0(item.Raw), "control_word": cw, "original_row_sha256": row.RawSHA, "actual_original_CW_comparisons": 1, "actual_native_calls": 1, "native_before_sha256": rskHash4E23C0(before), "native_after_sha256": rskHash4E23C0(after), "whole_foreign_banks_and_canaries": true, "full_raw_int32_return_and_native_service_pointers": true, "service_reply_ECX_EDX_physical_equivalence": false, "diagnostic_global_DWORD_is_not_production_memmap_owner": true, "physical_nil_EA_equivalence": false, "independent_native_fault_PC_qualification": false, "native_Go_heap_credit": false, "full_helper_body_or_caller_stack_FPU_credit": false, "signal_header": header})
}

func TestSettingEligibilityNativeOriginal40AA70(t *testing.T) {
	if os.Getenv(seWorkerEnv40AA70) == "1" {
		return
	}
	cases, e := seLoadCases40AA70()
	if e != nil {
		t.Fatal(e)
	}
	initial := os.Getenv("OPENNOX_RAW_DEFAULT_DAMAGE_INITIAL_ARTIFACTS")
	if initial == "" {
		t.Fatal("prepared unique INITIAL_ARTIFACTS required")
	}
	root, e := filepath.Abs(filepath.Join(initial, "setting-eligibility-40aa70"))
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
					if e := seRun40AA70(filepath.Join(root, fmt.Sprintf("%03d-%04x-%s", i, cw, backend)), exe, sha, backend, cw, item); e != nil {
						t.Fatal(e)
					}
				}) {
					passed++
				}
			}
		}
	}
	if e = rskWriteJSON4E23C0(filepath.Join(root, "actual-corpus-summary.json"), map[string]any{"inputs": 662, "attempted_native_calls": attempted, "passed": passed, "failed": attempted - passed, "stream_complete": attempted == 2648, "qualified_original_CW_comparisons": passed, "actual_native_backends": []string{"Go", "C"}, "independently_selected_control_words": []uint16{0x0e3f, 0x023f}, "input662_sha256": seInputSHA40AA70, "original1324_sha256": seReferenceSHA40AA70, "global_owner": "diagnostic-shared-C-DWORD; real-production-owner-separate", "native_fault_PC_and_physical_nil_EA_equivalence": false, "external_helper_body_credit": false}); e != nil {
		t.Fatal(e)
	}
	if attempted != 2648 || passed != 2648 {
		t.Fatalf("native corpus attempted=%d passed=%d", attempted, passed)
	}
	t.Logf("SETTING_ELIGIBILITY_NATIVE_ORIGINAL_OK inputs=662 native=2648 matching-original-CW-comparisons=2648 full-return-and-service-exact=true")
}
