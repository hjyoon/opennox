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

const tspInputSHA40A740 = "ab8ee32fc0c6afc64769a021277e7544c42aeb270de05c236518493d300b641b"
const tspReferenceSHA40A740 = "bf18aaf6293c85a0cd99f4384304734e67bea9694ef9c269e9326929ac1d7568"
const tspWorkerEnv40A740 = "OPENNOX_TWO_SETTINGS_POINTER_NATIVE_WORKER"
const tspDirectoryEnv40A740 = "OPENNOX_TWO_SETTINGS_POINTER_NATIVE_DIRECTORY"

type tspEvent40A740 struct {
	Kind      string            `json:"kind"`
	Sequence  int               `json:"sequence"`
	PC        uint32            `json:"PC"`
	Address   uint32            `json:"address"`
	Width     int               `json:"width"`
	Value     json.RawMessage   `json:"value"`
	Completed bool              `json:"completed"`
	Registers map[string]uint32 `json:"registers"`
}
type tspServiceReference40A740 struct {
	Site          uint32          `json:"site"`
	Target        uint32          `json:"target"`
	Args          []uint32        `json:"arguments_u32"`
	Sequence      int             `json:"original_call_event_begin"`
	Occurrence    int             `json:"occurrence"`
	ResponseIndex int             `json:"response_index"`
	Packet        tspPacket40A740 `json:"packet"`
}
type tspMemory40A740 struct {
	Address uint32 `json:"address"`
	After   string `json:"after"`
	Exact   bool   `json:"replay_exact"`
}
type tspReference40A740 struct {
	Name            string                      `json:"name"`
	CW              uint16                      `json:"control_word"`
	Entry           uint32                      `json:"entry"`
	InputSHA        string                      `json:"input_sha256"`
	Status          string                      `json:"status"`
	Valid           bool                        `json:"certificate_valid"`
	Result          *uint32                     `json:"actual_EAX_u32"`
	Events          []tspEvent40A740            `json:"events"`
	Services        []tspServiceReference40A740 `json:"services"`
	Memory          []tspMemory40A740           `json:"memory"`
	Used            map[string][]int            `json:"used_mutations"`
	Unused          map[string][]int            `json:"unused_mutations"`
	UsedResponses   map[string][]int            `json:"used_response_indices"`
	UnusedResponses map[string][]int            `json:"unused_response_indices"`
	AfterFields     map[uint32][]byte           `json:"-"`
	RawSHA          string                      `json:"-"`
}
type tspCase40A740 struct {
	Input      tspInput40A740
	Raw        json.RawMessage
	References []*tspReference40A740
}
type tspArgument40A740 struct {
	Word      uint32
	Pointer   uint64
	IsPointer bool
}
type tspTrace40A740 struct {
	Kind                              string
	Site, PE                          uint32
	PEWidth, Width                    int
	Complete, IsPointer, AddressValid bool
	Address, Base, Pointer            uint64
	Offset                            uint64
	Word                              uint32
	Arguments                         []tspArgument40A740
	ResponseIndex                     int
	ReplyWord, ReplyECX, ReplyEDX     uint32
	ReplyPointer                      uint64
}
type tspChildInput40A740 struct {
	Raw                    json.RawMessage
	Backend, ExecutableSHA string
}
type tspChildResult40A740 struct {
	PID                     int
	Backend                 string
	Return                  uint32
	UsedReads, UsedServices []int
	UsedResponses           map[string][]int
	TraceEvents             int
}

func tspCanonicalIndices40A740(v []int) []int {
	out := append([]int{}, v...)
	sort.Ints(out)
	for i := 1; i < len(out); i++ {
		if out[i] == out[i-1] {
			panic("duplicate consumed stimulus")
		}
	}
	return out
}
func tspResponseIndices40A740(values map[string][]int) map[string][]int {
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
func tspReadGzipJSON40A740(path, sha string, out any) error {
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
func tspLoadCases40A740() ([]tspCase40A740, error) {
	base := filepath.Join("testdata", "two_settings_pointers_40a740")
	var inputs struct {
		Cases []json.RawMessage `json:"cases"`
		Count int               `json:"case_count"`
	}
	if e := tspReadGzipJSON40A740(filepath.Join(base, "inputs-v1.json.gz"), tspInputSHA40A740, &inputs); e != nil {
		return nil, e
	}
	if inputs.Count != 394 || len(inputs.Cases) != 394 {
		return nil, fmt.Errorf("exact394 independent inputs")
	}
	cases := make([]tspCase40A740, 394)
	indices := map[string]int{}
	for i, p := range inputs.Cases {
		var b bytes.Buffer
		if e := json.Compact(&b, p); e != nil {
			return nil, e
		}
		cases[i].Raw = append(json.RawMessage(nil), b.Bytes()...)
		if e := json.Unmarshal(p, &cases[i].Input); e != nil {
			return nil, e
		}
		name := cases[i].Input.Name
		if _, ok := indices[name]; ok {
			return nil, fmt.Errorf("unique independent name")
		}
		indices[name] = i
	}
	f, z, e := rskOpenGzip4E23C0(filepath.Join(base, "reference-v2.json.gz"), tspReferenceSHA40A740)
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
			var row tspReference40A740
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
			for _, old := range item.References {
				if old.CW == row.CW {
					return nil, fmt.Errorf("unique actual original CW")
				}
			}
			row.RawSHA = rskHash4E23C0(raw)
			pages := map[uint32][]byte{}
			for _, m := range row.Memory {
				data := tspHex40A740(m.After)
				if !m.Exact || len(data) != 4096 {
					return nil, fmt.Errorf("source-qualified original full AFTER page")
				}
				pages[m.Address] = data
			}
			row.AfterFields = map[uint32][]byte{}
			add := func(pe uint32, width int) error {
				available := tspAvailable40A740(&item.Input, pe, width)
				if pe == 0x85b7a0 && item.Input.Engine.Mapped {
					available = 4
				}
				if available == 0 {
					return nil
				}
				p, e := tspOriginalBytes40A740(pages, pe, available)
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
					e = add(t+53, 1)
				case len(role) >= 4 && role[:4] == "unit":
					e = add(t+748, 4)
				case len(role) >= 6 && role[:6] == "update":
					e = add(t+276, 4)
				case len(role) >= 6 && role[:6] == "player":
					e = add(t+3680, 4)
					if e == nil {
						e = add(t+2064, 1)
					}
				}
				if e != nil {
					return nil, e
				}
			}
			if e = add(0x85b7a0, 4); e != nil {
				return nil, e
			}
			row.Memory = nil
			item.References = append(item.References, &row)
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
	if count != 788 || !rowsSeen {
		return nil, fmt.Errorf("exact788 authentic original rows")
	}
	for _, c := range cases {
		if len(c.References) != 2 {
			return nil, fmt.Errorf("two original CW rows per input")
		}
	}
	return cases, nil
}

// No original reference or predicted observation schedule enters this worker.
func TestTwoSettingsPointerNativeWorker40A740(t *testing.T) {
	if os.Getenv(tspWorkerEnv40A740) != "1" {
		return
	}
	dir := os.Getenv(tspDirectoryEnv40A740)
	if !filepath.IsAbs(dir) {
		t.Fatal("owned absolute child directory")
	}
	var input tspChildInput40A740
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
	var in tspInput40A740
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
	g, e := tspBuildGraph40A740(&in, filepath.Join(dir, "actual-shared-banks.bin"))
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
	emit := func(v tspTrace40A740) {
		p, e := json.Marshal(v)
		must(e)
		events++
		traceBytes += len(p) + 1
		if events > 1024 || traceBytes > 256<<10 {
			must(fmt.Errorf("bounded native observation trace"))
		}
		must(enc.Encode(v))
	}
	state := func(value uint32) tspChildResult40A740 {
		return tspChildResult40A740{os.Getpid(), input.Backend, value, g.UsedReads, g.UsedServices, tspResponseIndices40A740(g.UsedResponses), events}
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
	observer := &twoSettingsPointerNativeObserver40A740{EngineCell: g.engineCell()}
	observer.Operand = func(op twoSettingsPointerNativeOperand40A740) {
		pe, width, e := tspOperand40A740(g, op)
		must(e)
		if !op.Complete {
			rawDefaultDamageSharedProgress4E0B30(g.Header, uint64(events), op.Site, op.Address, int(op.Width), rawDefaultDamageSharedRead4E0B30, nil)
		}
		emit(tspTrace40A740{Kind: "operand", Site: op.Site, PE: pe, PEWidth: width, Width: int(op.Width), Complete: op.Complete, IsPointer: op.IsPointer, AddressValid: op.AddressValid, Address: uint64(uintptr(op.Address)), Base: uint64(uintptr(op.Base)), Pointer: uint64(uintptr(op.ValuePointer)), Offset: uint64(op.Offset), Word: op.ValueWord})
		if op.Complete {
			must(g.afterRead(op.Site, pe, width))
			save()
		}
	}
	observer.Service = func(site uint32, args []twoSettingsPointerNativeArg40A740) (twoSettingsPointerNativeReply40A740, bool) {
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
		reply := twoSettingsPointerNativeReply40A740{Word: packet.Registers["eax"], ECX: packet.Registers["ecx"], EDX: packet.Registers["edx"]}
		if tspPointerSite40A740(site) {
			reply.Pointer, e = g.pointer(packet.Registers["eax"])
			must(e)
		}
		v := tspTrace40A740{Kind: "service", Site: site, Complete: true, ResponseIndex: index, ReplyWord: reply.Word, ReplyECX: reply.ECX, ReplyEDX: reply.EDX, ReplyPointer: uint64(uintptr(reply.Pointer))}
		for _, a := range args {
			v.Arguments = append(v.Arguments, tspArgument40A740{a.Word, uint64(uintptr(a.Pointer)), a.IsPointer})
		}
		emit(v)
		g.UsedResponses[key] = append(g.UsedResponses[key], index)
		must(g.apply(packet.Writes))
		must(g.afterService(site))
		save()
		return reply, true
	}
	pop := pushTwoSettingsPointerNativeObserver40A740(observer)
	defer pop()
	backend := uint64(1)
	entry := reflect.ValueOf(sub40A740Native).Pointer()
	if in.Entry == 0x40a770 {
		entry = reflect.ValueOf(sub40A770Native).Pointer()
	}
	if input.Backend == "C" {
		backend = 2
		entry = reflect.ValueOf(twoSettingsPointerFixtureCall40A740).Pointer()
	} else if input.Backend != "Go" {
		t.Fatal("actual backend")
	}
	rawDefaultDamageSharedRootDispatched4E0B30(g.Header, entry, backend)
	save()
	must(rskWriteJSON4E23C0(filepath.Join(dir, "actual-child-PREEXEC.json"), map[string]any{"pid": os.Getpid(), "parent_pid": os.Getppid(), "backend": input.Backend, "executable_sha256": sha, "input_sha256": rskHash4E23C0(input.Raw), "BEFORE_sha256": rskHash4E23C0(before), "prepared_before_actual_root": true, "original_observations_passed_to_child": false, "diagnostic_engine_cell_is_foreign": true}))
	var result int32
	if input.Backend == "C" {
		result = twoSettingsPointerFixtureCall40A740(in.Entry)
	} else {
		switch in.Entry {
		case 0x40a740:
			result = sub40A740Native()
		case 0x40a770:
			result = sub40A770Native()
		default:
			t.Fatal("input entry")
		}
	}
	must(trace.Sync())
	must(rskWriteJSON4E23C0(filepath.Join(dir, "actual-return.json"), state(uint32(result))))
	runtime.KeepAlive(g)
}

func tspTraceRead40A740(path string) ([]tspTrace40A740, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	dec := json.NewDecoder(io.LimitReader(f, 256<<10+1))
	var out []tspTrace40A740
	var pending *tspTrace40A740
	for {
		var v tspTrace40A740
		e = dec.Decode(&v)
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, e
		}
		if v.Kind == "service" {
			if pending != nil {
				return nil, fmt.Errorf("service during unfinished read")
			}
			out = append(out, v)
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
		a.Pointer = b.Pointer
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

type tspStep40A740 struct {
	Kind     string
	Sequence int
	Event    tspEvent40A740
	Service  tspServiceReference40A740
}

func tspStaticWidth40A740(site uint32) int {
	switch site {
	case 0x40a759, 0x40a7af, 0x40a7b7, 0x40a7fe, 0x40a806:
		return 1
	case 0x40a7a3, 0x40a7a9, 0x40a7c0, 0x40a80f:
		return 4
	}
	return 0
}
func tspIdentity40A740(meta *tspMeta40A740, token uint32) (uint64, error) {
	if token == 0 {
		return 0, nil
	}
	p, ok := meta.Identities[strconv.FormatUint(uint64(token), 10)]
	if !ok {
		return 0, fmt.Errorf("original canonical pointer %08x", token)
	}
	return p, nil
}
func tspCompare40A740(in *tspInput40A740, row *tspReference40A740, meta *tspMeta40A740, trace []tspTrace40A740, before, after []byte, result *tspChildResult40A740) error {
	var steps []tspStep40A740
	for _, e := range row.Events {
		if tspStaticWidth40A740(e.PC) != 0 && ((e.Kind == "read" && e.Completed) || e.Kind == "fault") {
			kind := "operand"
			if e.Kind == "fault" {
				kind = "fault"
			}
			steps = append(steps, tspStep40A740{Kind: kind, Sequence: e.Sequence, Event: e})
		}
	}
	for _, s := range row.Services {
		steps = append(steps, tspStep40A740{Kind: "service", Sequence: s.Sequence, Service: s})
	}
	sort.Slice(steps, func(i, j int) bool { return steps[i].Sequence < steps[j].Sequence })
	if len(steps) != len(trace) {
		return fmt.Errorf("ordered actual observation count %d/%d", len(trace), len(steps))
	}
	for i, s := range steps {
		a := trace[i]
		if a.Kind != s.Kind {
			return fmt.Errorf("ordered observation kind %d", i)
		}
		if s.Kind == "service" {
			s := s.Service
			if a.Site != s.Site || len(a.Arguments) != len(s.Args) || a.ResponseIndex != s.ResponseIndex || a.ReplyWord != s.Packet.Registers["eax"] || a.ReplyECX != s.Packet.Registers["ecx"] || a.ReplyEDX != s.Packet.Registers["edx"] {
				return fmt.Errorf("original actual service/response-index/declared raw packet %d", i)
			}
			for j, arg := range a.Arguments {
				pointer := s.Site == 0x40a7c9 || s.Site == 0x40a7d9 || s.Site == 0x40a819
				if arg.IsPointer != pointer {
					return fmt.Errorf("service argument kind")
				}
				if pointer {
					p, e := tspIdentity40A740(meta, s.Args[j])
					if e != nil {
						return e
					}
					if p != arg.Pointer {
						return fmt.Errorf("all native service pointer bits")
					}
				} else if arg.Word != s.Args[j] {
					return fmt.Errorf("raw DWORD service argument")
				}
			}
			if tspPointerSite40A740(s.Site) {
				p, e := tspIdentity40A740(meta, s.Packet.Registers["eax"])
				if e != nil {
					return e
				}
				if p != a.ReplyPointer {
					return fmt.Errorf("full native input pointer response")
				}
			}
			continue
		}
		e := s.Event
		if a.Site != e.PC || a.PEWidth != tspStaticWidth40A740(e.PC) {
			return fmt.Errorf("static scalar/pointer width and site")
		}
		if s.Kind == "fault" {
			var logical uint32
			switch e.PC {
			case 0x40a759:
				logical = e.Registers["esi"] + 53
			case 0x40a7a3:
				logical = e.Registers["eax"] + 748
			case 0x40a7a9:
				logical = e.Registers["ecx"] + 276
			case 0x40a7af:
				logical = e.Registers["ecx"] + 3680
			case 0x40a7fe:
				logical = e.Registers["eax"] + 3680
			case 0x40a7b7:
				logical = e.Registers["ecx"] + 2064
			case 0x40a806:
				logical = e.Registers["eax"] + 2064
			case 0x40a7c0, 0x40a80f:
				logical = 0x85b7a0
			}
			if a.Complete || a.PE != logical || uint64(e.Address) < uint64(logical) || uint64(e.Address) >= uint64(logical)+uint64(a.PEWidth) {
				return fmt.Errorf("semantic fault operand and authentic raw missing-byte span")
			}
			continue
		}
		if !a.Complete || a.PE != e.Address {
			return fmt.Errorf("actual completed operand address")
		}
		var value string
		if e := json.Unmarshal(e.Value, &value); e != nil {
			return e
		}
		word := tspWord40A740(tspHex40A740(value))
		if a.IsPointer {
			p, e := tspIdentity40A740(meta, word)
			if e != nil {
				return e
			}
			if p != a.Pointer || a.Width != int(unsafe.Sizeof(unsafe.Pointer(nil))) {
				return fmt.Errorf("completed full native pointer")
			}
		} else if a.Word != word {
			return fmt.Errorf("actual byte/DWORD load value")
		}
	}
	expected, e := tspExpectedAfter40A740(meta, before, row.AfterFields)
	if e != nil {
		return e
	}
	if !bytes.Equal(expected, after) {
		return fmt.Errorf("whole native AFTER/canaries versus authentic original fields")
	}
	if !reflect.DeepEqual(tspCanonicalIndices40A740(result.UsedReads), tspCanonicalIndices40A740(row.Used["after_reads"])) || !reflect.DeepEqual(tspCanonicalIndices40A740(result.UsedServices), tspCanonicalIndices40A740(row.Used["after_services"])) || !reflect.DeepEqual(result.UsedResponses, row.UsedResponses) {
		return fmt.Errorf("actual consumed versus unused independent inputs")
	}
	if row.Status == "returned" && (row.Result == nil || result.Return != *row.Result) {
		return fmt.Errorf("full raw int32 return")
	}
	return nil
}

func tspRun40A740(dir, exe, sha, backend string, item *tspCase40A740) (err error) {
	if e := os.Mkdir(dir, 0700); e != nil {
		return e
	}
	raw, e := json.Marshal(tspChildInput40A740{item.Raw, backend, sha})
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
	cmd := exec.CommandContext(ctx, exe, "-test.run=^TestTwoSettingsPointerNativeWorker40A740$", "-test.count=1", "-test.timeout=10s")
	cmd.Env = append(os.Environ(), tspWorkerEnv40A740+"=1", tspDirectoryEnv40A740+"="+dir)
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
	terminal := map[string]any{"pid": pid, "backend": backend, "start_UTC": start, "Wait_terminal_UTC": time.Now().UTC(), "start_error": fmt.Sprint(startError), "Wait_error": fmt.Sprint(waitError), "exit_code": exit, "forced_request": forced.Load(), "context_error": fmt.Sprint(ctx.Err()), "stdout_exceeded": outLog.Exceeded, "stderr_exceeded": errLog.Exceeded, "executable_sha256": sha, "reference_CW_comparisons_planned": 2, "original_execution_credit": 0}
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
	if pre.PID != pid || pre.SHA != sha || pre.InputSHA != rskHash4E23C0(item.Raw) {
		return fmt.Errorf("same-PID PREEXEC/input/main binding")
	}
	var meta tspMeta40A740
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
	var state tspChildResult40A740
	p, e = os.ReadFile(filepath.Join(dir, "actual-input-state.json"))
	if e != nil {
		return e
	}
	if e = json.Unmarshal(p, &state); e != nil {
		return e
	}
	if state.PID != pid || state.Backend != backend {
		return fmt.Errorf("actual durable input consumption PID/backend")
	}
	result := &state
	if exit == 0 {
		var r tspChildResult40A740
		p, e = os.ReadFile(filepath.Join(dir, "actual-return.json"))
		if e != nil {
			return e
		}
		if e = json.Unmarshal(p, &r); e != nil {
			return e
		}
		if r.PID != pid || r.Backend != backend || header[2] != 2 || !reflect.DeepEqual(r.UsedReads, state.UsedReads) || !reflect.DeepEqual(r.UsedServices, state.UsedServices) || !reflect.DeepEqual(r.UsedResponses, state.UsedResponses) {
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
	trace, e := tspTraceRead40A740(filepath.Join(dir, "actual-observations.jsonl"))
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
	var refSHAs []string
	for _, row := range item.References {
		if (row.Status == "returned") != (exit == 0) {
			return fmt.Errorf("authentic original/native terminal domain differs CW%04x", row.CW)
		}
		if e = tspCompare40A740(&item.Input, row, &meta, trace, before, after, result); e != nil {
			return fmt.Errorf("CW%04x: %w", row.CW, e)
		}
		refSHAs = append(refSHAs, row.RawSHA)
	}
	return rskWriteJSON4E23C0(filepath.Join(dir, "actual-qualification.json"), map[string]any{"qualified_selected_observations": true, "actual_pid": pid, "backend": backend, "input": item.Input.Name, "input_sha256": rskHash4E23C0(item.Raw), "original_CW_references": []uint16{0x0e3f, 0x023f}, "original_row_sha256": refSHAs, "actual_original_CW_comparisons": 2, "actual_native_calls": 1, "native_before_sha256": rskHash4E23C0(before), "native_after_sha256": rskHash4E23C0(after), "whole_foreign_banks_and_canaries": true, "full_raw_int32_return_and_native_service_pointers": true, "service_reply_ECX_EDX_physical_equivalence": false, "diagnostic_engine_DWORD_is_not_production_Go_flags_owner": true, "physical_nil_EA_equivalence": false, "independent_native_fault_PC_qualification": false, "native_Go_heap_credit": false, "full_helper_body_or_caller_stack_FPU_credit": false, "signal_header": header})
}

func TestTwoSettingsPointerNativeOriginal40A740(t *testing.T) {
	if os.Getenv(tspWorkerEnv40A740) == "1" {
		return
	}
	cases, e := tspLoadCases40A740()
	if e != nil {
		t.Fatal(e)
	}
	initial := os.Getenv("OPENNOX_RAW_DEFAULT_DAMAGE_INITIAL_ARTIFACTS")
	if initial == "" {
		t.Fatal("prepared unique INITIAL_ARTIFACTS required")
	}
	root, e := filepath.Abs(filepath.Join(initial, "two-settings-pointer-40a740"))
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
		for _, backend := range []string{"Go", "C"} {
			attempted++
			item := &cases[i]
			if t.Run(backend+"/"+item.Input.Name, func(t *testing.T) {
				if e := tspRun40A740(filepath.Join(root, fmt.Sprintf("%03d-%s", i, backend)), exe, sha, backend, item); e != nil {
					t.Fatal(e)
				}
			}) {
				passed++
			}
		}
	}
	if e = rskWriteJSON4E23C0(filepath.Join(root, "actual-corpus-summary.json"), map[string]any{"attempted_native_calls": attempted, "passed": passed, "failed": attempted - passed, "stream_complete": attempted == 788, "qualified_original_CW_comparisons": 2 * passed, "actual_native_backends": []string{"Go", "C"}, "input394_sha256": tspInputSHA40A740, "original788_sha256": tspReferenceSHA40A740, "engine_flags_owner": "diagnostic-C-owned-or-protected-DWORD; production-Go-owner-separate", "native_fault_PC_and_physical_nil_EA_equivalence": false, "external_helper_body_credit": false}); e != nil {
		t.Fatal(e)
	}
	if attempted != 788 || passed != 788 {
		t.Fatalf("native corpus attempted=%d passed=%d", attempted, passed)
	}
	t.Logf("TWO_SETTINGS_POINTER_NATIVE_ORIGINAL_OK native=788 original_CW_comparisons=1576")
}
