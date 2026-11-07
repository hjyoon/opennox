//go:build darwin && arm64

package legacy

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/hex"
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

const sdInputSHA40A970 = "6896556370c4c649a794bd60328ffe253245116aa0319c703c0b9e1d66d27778"
const sdReferenceSHA40A970 = "debd6b5c669a61d955b31744c67a44ca17876df7aa051bf0275036e9de56cc81"
const sdWorkerEnv40A970 = "OPENNOX_SUDDEN_DEATH_NATIVE_WORKER"
const sdDirectoryEnv40A970 = "OPENNOX_SUDDEN_DEATH_NATIVE_DIRECTORY"

type sdEvent40A970 struct {
	Kind      string            `json:"kind"`
	Sequence  int               `json:"sequence"`
	PC        uint32            `json:"PC"`
	Address   uint32            `json:"address"`
	Width     int               `json:"width"`
	Value     json.RawMessage   `json:"value"`
	Completed bool              `json:"completed"`
	Registers map[string]uint32 `json:"registers"`
}
type sdServiceReference40A970 struct {
	Site          uint32         `json:"site"`
	Target        uint32         `json:"target"`
	Args          []uint32       `json:"arguments_u32"`
	Sequence      int            `json:"original_call_event_begin"`
	Occurrence    int            `json:"occurrence"`
	ResponseIndex int            `json:"response_index"`
	Packet        sdPacket40A970 `json:"packet"`
}
type sdConverterReference40A970 struct {
	Site     uint32 `json:"site"`
	Sequence int    `json:"original_call_event_begin"`
	Result   uint32 `json:"actual_EAX_u32"`
	Complete bool   `json:"completed"`
}
type sdMemory40A970 struct {
	Address uint32 `json:"address"`
	After   string `json:"after"`
	Exact   bool   `json:"replay_exact"`
}
type sdReference40A970 struct {
	Name          string                       `json:"name"`
	CW            uint16                       `json:"control_word"`
	Entry         uint32                       `json:"entry"`
	InputSHA      string                       `json:"input_sha256"`
	Status        string                       `json:"status"`
	Valid         bool                         `json:"certificate_valid"`
	PublicReturn  string                       `json:"public_return_kind"`
	Events        []sdEvent40A970              `json:"events"`
	Services      []sdServiceReference40A970   `json:"services"`
	Converters    []sdConverterReference40A970 `json:"actual_converter_frames"`
	Memory        []sdMemory40A970             `json:"memory"`
	Used          map[string][]int             `json:"used_mutations"`
	Unused        map[string][]int             `json:"unused_mutations"`
	UsedResponses map[string][]int             `json:"used_response_indices"`
	AfterFields   map[uint32][]byte            `json:"-"`
	RawSHA        string                       `json:"-"`
}
type sdCase40A970 struct {
	Input      sdInput40A970
	Raw        json.RawMessage
	References map[uint16]*sdReference40A970
}
type sdArgument40A970 struct {
	Kind    string
	Word    uint32
	Pointer uint64
	String  string
}
type sdTrace40A970 struct {
	Kind                                      string
	Site, PE                                  uint32
	Width                                     int
	Complete, Write, ValueValid, AddressValid bool
	Address, Base, Offset                     uint64
	Value                                     uint32
	Arguments                                 []sdArgument40A970
	ResponseIndex                             int
	ReplyWord, DeclaredECX, DeclaredEDX       uint32
	ReplyPointer                              uint64
	FloatFormat, FloatHex                     string
	NumericKind                               string
	Bits, Result                              uint32
	BitsValid, ResultValid                    bool
}
type sdChildInput40A970 struct {
	Raw                    json.RawMessage
	ControlWord            uint16
	Backend, ExecutableSHA string
}
type sdChildResult40A970 struct {
	PID                                 int
	Backend                             string
	ControlWord                         uint16
	UsedReads, UsedWrites, UsedServices []int
	UsedResponses                       map[string][]int
	TraceEvents                         int
}

func sdSorted40A970(v []int) []int {
	out := append([]int{}, v...)
	sort.Ints(out)
	for i := 1; i < len(out); i++ {
		if out[i] == out[i-1] {
			panic("duplicate consumed stimulus")
		}
	}
	return out
}
func sdResponses40A970(v map[string][]int) map[string][]int {
	out := map[string][]int{}
	for site, values := range v {
		unique := map[int]bool{}
		for _, n := range values {
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
func sdOriginalBytes40A970(pages map[uint32][]byte, address uint32, n int) ([]byte, error) {
	p := make([]byte, n)
	for i := range p {
		a := address + uint32(i)
		page, ok := pages[a&^uint32(4095)]
		if !ok || len(page) != 4096 {
			return nil, fmt.Errorf("authentic original AFTER page")
		}
		p[i] = page[a&4095]
	}
	return p, nil
}

func sdLoadCases40A970() ([]sdCase40A970, error) {
	base := filepath.Join("testdata", "sudden_death_40a970")
	var input struct {
		Cases []json.RawMessage `json:"cases"`
		CW    []uint16          `json:"FPU_controls"`
	}
	f, z, e := rskOpenGzip4E23C0(filepath.Join(base, "inputs-v1.json.gz"), sdInputSHA40A970)
	if e != nil {
		return nil, e
	}
	dec := json.NewDecoder(z)
	e = dec.Decode(&input)
	var extra any
	if e == nil {
		e = dec.Decode(&extra)
		if e == io.EOF {
			e = nil
		} else if e == nil {
			e = fmt.Errorf("one independent input document")
		}
	}
	ze, fe := z.Close(), f.Close()
	if e != nil || ze != nil || fe != nil {
		return nil, fmt.Errorf("input gzip/CRC %v %v %v", e, ze, fe)
	}
	if len(input.Cases) != 403 || !reflect.DeepEqual(input.CW, []uint16{0x0e3f, 0x023f}) {
		return nil, fmt.Errorf("403 input cases and two independent CW")
	}
	cases := make([]sdCase40A970, 403)
	indices := map[string]int{}
	for i, p := range input.Cases {
		var compact bytes.Buffer
		if e = json.Compact(&compact, p); e != nil {
			return nil, e
		}
		cases[i].Raw = append(json.RawMessage(nil), compact.Bytes()...)
		cases[i].References = map[uint16]*sdReference40A970{}
		if e = json.Unmarshal(p, &cases[i].Input); e != nil {
			return nil, e
		}
		name := cases[i].Input.Name
		if _, ok := indices[name]; ok {
			return nil, fmt.Errorf("unique input name")
		}
		indices[name] = i
	}
	f, z, e = rskOpenGzip4E23C0(filepath.Join(base, "reference-v1.json.gz"), sdReferenceSHA40A970)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	defer z.Close()
	dec = json.NewDecoder(z)
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
			return nil, fmt.Errorf("one original row stream")
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
			var row sdReference40A970
			if e = json.Unmarshal(raw, &row); e != nil {
				return nil, e
			}
			i, ok := indices[row.Name]
			if !ok {
				return nil, fmt.Errorf("original unknown input")
			}
			item := &cases[i]
			if !row.Valid || row.Entry != 0x40a970 || row.PublicReturn != "void" || row.InputSHA != rskHash4E23C0(item.Raw) || (row.CW != 0x0e3f && row.CW != 0x023f) {
				return nil, fmt.Errorf("actual original input/CW/source certificate")
			}
			if _, ok = item.References[row.CW]; ok {
				return nil, fmt.Errorf("unique original CW per input")
			}
			row.RawSHA = rskHash4E23C0(raw)
			pages := map[uint32][]byte{}
			for _, m := range row.Memory {
				p := sdHex40A970(m.After)
				if !m.Exact || len(p) != 4096 {
					return nil, fmt.Errorf("source-qualified original full AFTER")
				}
				pages[m.Address] = p
			}
			row.AfterFields = map[uint32][]byte{}
			for _, pe := range sdCellPE40A970 {
				p, e := sdOriginalBytes40A970(pages, pe, 4)
				if e != nil {
					return nil, e
				}
				row.AfterFields[pe] = p
			}
			for _, region := range item.Input.Regions {
				pe := region.Base + 3680
				n := sdAvailable40A970(&item.Input, pe, 4)
				if n == 0 {
					continue
				}
				p, e := sdOriginalBytes40A970(pages, pe, n)
				if e != nil {
					return nil, e
				}
				row.AfterFields[pe] = p
			}
			if len(row.Converters) != 2 || !row.Converters[0].Complete || !row.Converters[1].Complete {
				return nil, fmt.Errorf("two actual original converter calls")
			}
			row.Memory = nil
			item.References[row.CW] = &row
			count++
		}
		if token, e = dec.Token(); e != nil || token != json.Delim(']') {
			return nil, fmt.Errorf("complete original rows")
		}
	}
	if token, e = dec.Token(); e != nil || token != json.Delim('}') {
		return nil, fmt.Errorf("complete original object")
	}
	extra = nil
	if e = dec.Decode(&extra); e != io.EOF {
		return nil, fmt.Errorf("original gzip EOF/CRC %v", e)
	}
	if count != 806 || !rowsSeen {
		return nil, fmt.Errorf("exact806 authentic original rows")
	}
	for _, c := range cases {
		if len(c.References) != 2 {
			return nil, fmt.Errorf("both selected original CW")
		}
	}
	return cases, nil
}

// This worker receives one independent input and one independently selected CW.
// Original results, observed helper calls and AFTER observations never enter it.
func TestSuddenDeathNativeWorker40A970(t *testing.T) {
	if os.Getenv(sdWorkerEnv40A970) != "1" {
		return
	}
	dir := os.Getenv(sdDirectoryEnv40A970)
	if !filepath.IsAbs(dir) {
		t.Fatal("absolute unique child directory")
	}
	p, e := os.ReadFile(filepath.Join(dir, "input.json.gz"))
	if e != nil {
		t.Fatal(e)
	}
	z, e := gzip.NewReader(bytes.NewReader(p))
	if e != nil {
		t.Fatal(e)
	}
	raw, e := io.ReadAll(z)
	ce := z.Close()
	if e != nil || ce != nil {
		t.Fatal(e, ce)
	}
	var child sdChildInput40A970
	if e = json.Unmarshal(raw, &child); e != nil {
		t.Fatal(e)
	}
	if child.ControlWord != 0x0e3f && child.ControlWord != 0x023f {
		t.Fatal("independent two-CW selection")
	}
	var in sdInput40A970
	if e = json.Unmarshal(child.Raw, &in); e != nil {
		t.Fatal(e)
	}
	exe, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	sha, e := rskFileSHA4E23C0(exe)
	if e != nil || sha != child.ExecutableSHA {
		t.Fatal("same preserved actual test main", e)
	}
	g, e := sdBuildGraph40A970(&in, filepath.Join(dir, "actual-shared-banks.bin"))
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
	emit := func(v sdTrace40A970) {
		p, e := json.Marshal(v)
		must(e)
		events++
		traceBytes += len(p) + 1
		if events > 1024 || traceBytes > 256<<10 {
			must(fmt.Errorf("bounded actual observation trace"))
		}
		must(enc.Encode(v))
	}
	state := func() sdChildResult40A970 {
		return sdChildResult40A970{os.Getpid(), child.Backend, child.ControlWord, g.UsedReads, g.UsedWrites, g.UsedServices, sdResponses40A970(g.UsedResponses), events}
	}
	for site := range in.Services {
		g.UsedResponses[site] = []int{}
	}
	save := func() { must(rskWriteJSON4E23C0(filepath.Join(dir, "actual-input-state.json"), state())) }
	if e = rawDefaultDamageSharedBegin4E0B30(g.Header); e != nil {
		t.Fatal(e)
	}
	defer func() {
		if e := rawDefaultDamageSharedEnd4E0B30(); e != nil {
			t.Error(e)
		}
	}()
	observer := &suddenDeathNativeObserver40A970{ControlWord: child.ControlWord, Cells: g.cells()}
	observer.Operand = func(op suddenDeathNativeOperand40A970) {
		pe, width, e := sdOperand40A970(g, op)
		must(e)
		if !op.Complete {
			operation := rawDefaultDamageSharedRead4E0B30
			if op.Write {
				operation = rawDefaultDamageSharedWrite4E0B30
			}
			rawDefaultDamageSharedProgress4E0B30(g.Header, uint64(events), op.Site, op.Address, int(op.Width), operation, nil)
		}
		emit(sdTrace40A970{Kind: "operand", Site: op.Site, PE: pe, Width: width, Complete: op.Complete, Write: op.Write, ValueValid: op.ValueValid, AddressValid: op.AddressValid, Address: uint64(uintptr(op.Address)), Base: uint64(uintptr(op.Base)), Offset: uint64(op.Offset), Value: op.Value})
		if op.Complete {
			if op.Write {
				must(g.afterWrite(op.Site, pe, width))
			} else {
				must(g.afterRead(op.Site, pe, width))
			}
			save()
		}
	}
	observer.Numeric = func(n suddenDeathNativeNumeric40A970) {
		emit(sdTrace40A970{Kind: "numeric", Site: n.Site, NumericKind: n.Kind, Bits: n.Bits, Result: n.Result, Complete: n.Complete, BitsValid: n.BitsValid, ResultValid: n.ResultValid})
	}
	observer.Service = func(site uint32, args []suddenDeathNativeArg40A970) (suddenDeathNativeReply40A970, bool) {
		key := fmt.Sprintf("%08x", site)
		schedule, ok := in.Services[key]
		if !ok || len(args) != schedule.Argc || len(schedule.Responses) == 0 {
			must(fmt.Errorf("declared actual service site/arity"))
		}
		index := g.Services[site]
		if index >= len(schedule.Responses) {
			if !schedule.Repeat {
				must(fmt.Errorf("undeclared repeated service"))
			}
			index = len(schedule.Responses) - 1
		}
		packet := schedule.Responses[index]
		reply := suddenDeathNativeReply40A970{Word: packet.Registers["eax"]}
		v := sdTrace40A970{Kind: "service", Site: site, Complete: true, ResponseIndex: index, ReplyWord: reply.Word, DeclaredECX: packet.Registers["ecx"], DeclaredEDX: packet.Registers["edx"]}
		if site == 0x40a9ba || site == 0x40a9df {
			reply.Pointer, e = g.pointer(packet.Registers["eax"])
			must(e)
			v.ReplyPointer = uint64(uintptr(reply.Pointer))
		}
		if site == 0x40a98a || site == 0x40a9a1 {
			reply.FloatFormat, reply.FloatPayload, e = sdRawFPInput40A970(packet.X87)
			must(e)
			v.FloatFormat = reply.FloatFormat
			n := 8
			if reply.FloatFormat == "ext80" {
				n = 10
			}
			v.FloatHex = hex.EncodeToString(reply.FloatPayload[:n])
		}
		for _, a := range args {
			v.Arguments = append(v.Arguments, sdArgument40A970{a.Kind, a.Word, uint64(uintptr(a.Pointer)), a.String})
		}
		emit(v)
		g.UsedResponses[key] = append(g.UsedResponses[key], index)
		must(g.apply(packet.Writes))
		must(g.afterService(site))
		save()
		return reply, true
	}
	pop := pushSuddenDeathNativeObserver40A970(observer)
	defer pop()
	backend := uint64(1)
	entry := reflect.ValueOf(sub40A970Native).Pointer()
	if child.Backend == "C" {
		backend = 2
		entry = reflect.ValueOf(suddenDeathFixtureCall40A970).Pointer()
	} else if child.Backend != "Go" {
		t.Fatal("actual backend")
	}
	rawDefaultDamageSharedRootDispatched4E0B30(g.Header, entry, backend)
	save()
	must(rskWriteJSON4E23C0(filepath.Join(dir, "actual-child-PREEXEC.json"), map[string]any{"pid": os.Getpid(), "parent_pid": os.Getppid(), "backend": child.Backend, "control_word": child.ControlWord, "executable_sha256": sha, "input_sha256": rskHash4E23C0(child.Raw), "BEFORE_sha256": rskHash4E23C0(before), "prepared_before_actual_root": true, "original_observations_passed_to_child": false, "diagnostic_cells_are_foreign_not_production_alias": true}))
	if child.Backend == "C" {
		suddenDeathFixtureCall40A970()
	} else {
		sub40A970Native()
	}
	must(trace.Sync())
	must(rskWriteJSON4E23C0(filepath.Join(dir, "actual-return.json"), state()))
	runtime.KeepAlive(g)
}

func sdReadTrace40A970(path string) ([]sdTrace40A970, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	dec := json.NewDecoder(io.LimitReader(f, 256<<10+1))
	var out []sdTrace40A970
	var pending *sdTrace40A970
	for {
		var event sdTrace40A970
		e = dec.Decode(&event)
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, e
		}
		switch event.Kind {
		case "service":
			if pending != nil || !event.Complete {
				return nil, fmt.Errorf("service during unfinished operation")
			}
			out = append(out, event)
		case "operand", "numeric":
			if !event.Complete {
				if pending != nil {
					return nil, fmt.Errorf("duplicate before operation")
				}
				copy := event
				pending = &copy
				continue
			}
			if pending == nil || pending.Kind != event.Kind {
				return nil, fmt.Errorf("completed operation without actual before")
			}
			a, b := *pending, event
			a.Complete = true
			if event.Kind == "operand" {
				if a.Write {
					if !a.ValueValid || a.Value != b.Value {
						return nil, fmt.Errorf("write value changed across actual store")
					}
				} else {
					a.Value, a.ValueValid = b.Value, b.ValueValid
				}
			} else {
				if a.NumericKind == "store32" {
					if a.BitsValid || a.ResultValid || !b.BitsValid || b.ResultValid {
						return nil, fmt.Errorf("numeric store32 actual validity")
					}
					a.Bits, a.BitsValid = b.Bits, b.BitsValid
				} else if a.NumericKind == "convert32" {
					if !a.BitsValid || a.ResultValid || !b.BitsValid || !b.ResultValid || a.Bits != b.Bits {
						return nil, fmt.Errorf("numeric convert32 actual validity")
					}
					a.Result, a.ResultValid = b.Result, b.ResultValid
				} else {
					return nil, fmt.Errorf("declared numeric boundary")
				}
			}
			if !reflect.DeepEqual(a, b) {
				return nil, fmt.Errorf("before/after operation identity changed")
			}
			pending = nil
			out = append(out, event)
		default:
			return nil, fmt.Errorf("actual observation kind")
		}
		if len(out) > 1024 {
			return nil, fmt.Errorf("actual trace bound")
		}
	}
	if pending != nil {
		if pending.Kind != "operand" || pending.Write {
			return nil, fmt.Errorf("only unfinished public/status READ may be a raw fault")
		}
		pending.Kind = "fault"
		out = append(out, *pending)
	}
	return out, nil
}

type sdStep40A970 struct {
	Kind         string
	Sequence     int
	Event        sdEvent40A970
	Service      sdServiceReference40A970
	NumericKind  string
	NumericValue uint32
	Site         uint32
}

func sdPublicSite40A970(site uint32) bool {
	switch site {
	case 0x40a970, 0x40a97b, 0x40a980, 0x40a99c, 0x40a9b5, 0x40a9c5:
		return true
	}
	return false
}
func sdNativeIdentity40A970(meta *sdMeta40A970, token uint32) (uint64, error) {
	if token == 0 {
		return 0, nil
	}
	p, ok := meta.Identities[strconv.FormatUint(uint64(token), 10)]
	if !ok {
		return 0, fmt.Errorf("unknown authentic PE player token %08x", token)
	}
	return p, nil
}

// The oracle contributes recorded values, call frames and AFTER fields only.
// This comparison does not evaluate a branch, floating conversion or helper.
func sdCompare40A970(row *sdReference40A970, meta *sdMeta40A970, trace []sdTrace40A970, before, after []byte, result *sdChildResult40A970) (privateStoreBitDifferences int, err error) {
	var steps []sdStep40A970
	for _, event := range row.Events {
		if sdPublicSite40A970(event.PC) && ((event.Kind == "read" || event.Kind == "write") && event.Completed || event.Kind == "fault") {
			kind := "operand"
			if event.Kind == "fault" {
				kind = "fault"
			}
			steps = append(steps, sdStep40A970{Kind: kind, Sequence: event.Sequence, Event: event})
		}
		if (event.PC == 0x40a98f || event.PC == 0x40a9a6) && event.Kind == "write" && event.Completed {
			var word uint32
			if e := json.Unmarshal(event.Value, &word); e != nil {
				return 0, e
			}
			steps = append(steps, sdStep40A970{Kind: "numeric", Sequence: event.Sequence, NumericKind: "store32", NumericValue: word, Site: event.PC})
		}
	}
	for _, service := range row.Services {
		steps = append(steps, sdStep40A970{Kind: "service", Sequence: service.Sequence, Service: service})
	}
	for _, converter := range row.Converters {
		steps = append(steps, sdStep40A970{Kind: "numeric", Sequence: converter.Sequence, NumericKind: "convert32", NumericValue: converter.Result, Site: converter.Site})
	}
	sort.Slice(steps, func(i, j int) bool { return steps[i].Sequence < steps[j].Sequence })
	if len(steps) != len(trace) {
		return 0, fmt.Errorf("all ordered public/status/service/numeric observations %d/%d", len(trace), len(steps))
	}
	var latestNativeStore uint32
	haveStore := false
	for i, step := range steps {
		a := trace[i]
		if a.Kind != step.Kind {
			return 0, fmt.Errorf("ordered operation kind %d got%s want%s", i, a.Kind, step.Kind)
		}
		switch step.Kind {
		case "service":
			s := step.Service
			if a.Site != s.Site || a.ResponseIndex != s.ResponseIndex || len(a.Arguments) != len(s.Args) || a.ReplyWord != s.Packet.Registers["eax"] || a.DeclaredECX != s.Packet.Registers["ecx"] || a.DeclaredEDX != s.Packet.Registers["edx"] {
				return 0, fmt.Errorf("actual service and independent raw packet %d", i)
			}
			for j, arg := range a.Arguments {
				word := s.Args[j]
				switch s.Site {
				case 0x40a98a, 0x40a9a1:
					want, ok := map[uint32]string{0x5883f8: "SuddenDeathPlayerThreshold", 0x588414: "SuddenDeathLifeTime"}[word]
					if !ok || arg.Kind != "string" || arg.String != want {
						return 0, fmt.Errorf("original immutable string identity versus actual native string argument")
					}
				case 0x40a9d6, 0x40a9df:
					if j == 0 {
						p, e := sdNativeIdentity40A970(meta, word)
						if e != nil {
							return 0, e
						}
						if arg.Kind != "player" || arg.Pointer != p {
							return 0, fmt.Errorf("actual cached full native player argument")
						}
					} else if arg.Kind != "word" || arg.Word != word {
						return 0, fmt.Errorf("actual full DWORD status mask")
					}
				default:
					if arg.Kind != "word" || arg.Word != word {
						return 0, fmt.Errorf("actual full DWORD service argument")
					}
				}
			}
			if s.Site == 0x40a9ba || s.Site == 0x40a9df {
				p, e := sdNativeIdentity40A970(meta, s.Packet.Registers["eax"])
				if e != nil {
					return 0, e
				}
				if a.ReplyPointer != p {
					return 0, fmt.Errorf("actual native pointer reply transport")
				}
			}
			if s.Site == 0x40a98a || s.Site == 0x40a9a1 {
				if s.Packet.X87 == nil || a.FloatFormat != s.Packet.X87.Format || a.FloatHex != s.Packet.X87.Hex {
					return 0, fmt.Errorf("actual raw8/10B input transport, not expected FP result")
				}
			}
		case "numeric":
			if a.Site != step.Site || a.NumericKind != step.NumericKind || !a.Complete || !a.BitsValid {
				return 0, fmt.Errorf("actual selected numeric boundary")
			}
			if a.NumericKind == "store32" {
				latestNativeStore, haveStore = a.Bits, true
				if a.Bits != step.NumericValue {
					privateStoreBitDifferences++
				}
			} else {
				if !haveStore || a.Bits != latestNativeStore || !a.ResultValid || a.Result != step.NumericValue {
					return 0, fmt.Errorf("actual numerical converter result versus original FISTP EAX")
				}
				haveStore = false
			}
		default:
			e := step.Event
			if a.Site != e.PC || a.Width != 4 {
				return 0, fmt.Errorf("static actual DWORD operand site/width")
			}
			if step.Kind == "fault" {
				logical := e.Registers["esi"] + 3680
				if a.Complete || a.Write || a.PE != logical || e.PC != 0x40a9c5 || e.Address < logical || uint64(e.Address) >= uint64(logical)+4 {
					return 0, fmt.Errorf("authentic semantic status fault and raw missing span, physicalEA separate")
				}
				continue
			}
			if !a.Complete || !a.ValueValid || a.PE != e.Address || a.Write != (e.Kind == "write") {
				return 0, fmt.Errorf("actual completed public/status operand")
			}
			var word uint32
			if e.Kind == "read" {
				var raw string
				if err = json.Unmarshal(e.Value, &raw); err != nil {
					return 0, err
				}
				p, e := hex.DecodeString(raw)
				if e != nil {
					return 0, e
				}
				if len(p) != 4 {
					return 0, fmt.Errorf("recorded original DWORD read")
				}
				word = sdWord40A970(p)
			} else if err = json.Unmarshal(e.Value, &word); err != nil {
				return 0, err
			}
			if a.Value != word {
				return 0, fmt.Errorf("actual DWORD value versus recorded original")
			}
		}
	}
	if haveStore {
		return 0, fmt.Errorf("unmatched actual numeric store/converter chain")
	}
	expected, e := sdExpectedAfter40A970(meta, before, row.AfterFields)
	if e != nil {
		return 0, e
	}
	if !bytes.Equal(expected, after) {
		return 0, fmt.Errorf("whole actual native banks/canaries versus authentic original field AFTER")
	}
	if !reflect.DeepEqual(sdSorted40A970(result.UsedReads), sdSorted40A970(row.Used["after_reads"])) || !reflect.DeepEqual(sdSorted40A970(result.UsedWrites), sdSorted40A970(row.Used["after_writes"])) || !reflect.DeepEqual(sdSorted40A970(result.UsedServices), sdSorted40A970(row.Used["after_services"])) || !reflect.DeepEqual(result.UsedResponses, row.UsedResponses) {
		return 0, fmt.Errorf("all actual independent input consumption versus original matching CW")
	}
	return privateStoreBitDifferences, nil
}

func sdRun40A970(dir, exe, sha, backend string, cw uint16, item *sdCase40A970) (err error) {
	row, ok := item.References[cw]
	if !ok {
		return fmt.Errorf("one authentic reference for selected CW")
	}
	if e := os.Mkdir(dir, 0700); e != nil {
		return e
	}
	raw, e := json.Marshal(sdChildInput40A970{item.Raw, cw, backend, sha})
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
		return fmt.Errorf("independent-only child gzip %v %v %v", e, ze, fe)
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
	cmd := exec.CommandContext(ctx, exe, "-test.run=^TestSuddenDeathNativeWorker40A970$", "-test.count=1", "-test.timeout=10s")
	cmd.Env = append(os.Environ(), sdWorkerEnv40A970+"=1", sdDirectoryEnv40A970+"="+dir)
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
	terminal := map[string]any{"pid": pid, "backend": backend, "control_word": cw, "start_UTC": start, "Wait_terminal_UTC": time.Now().UTC(), "start_error": fmt.Sprint(startError), "Wait_error": fmt.Sprint(waitError), "exit_code": exit, "forced_request": forced.Load(), "context_error": fmt.Sprint(ctx.Err()), "stdout_exceeded": outLog.Exceeded, "stderr_exceeded": errLog.Exceeded, "executable_sha256": sha, "matching_original_CW_comparisons_planned": 1, "original_execution_credit": 0}
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
		return fmt.Errorf("bounded child did not terminate naturally %v %v", startError, waitError)
	}
	var pre struct {
		PID       int    `json:"pid"`
		Backend   string `json:"backend"`
		CW        uint16 `json:"control_word"`
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
	if pre.PID != pid || pre.Backend != backend || pre.CW != cw || pre.SHA != sha || pre.InputSHA != rskHash4E23C0(item.Raw) {
		return fmt.Errorf("same PID/input/CW/main PREEXEC")
	}
	var meta sdMeta40A970
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
		return fmt.Errorf("whole actual shared bank extent")
	}
	header, e := rskReadHeader4E23C0(raw)
	if e != nil {
		return e
	}
	after := raw[meta.Page:]
	if header[0] != rawDefaultDamageSharedMagic4E0B30 || header[1] != rawDefaultDamageSharedVersion4E0B30 || header[36] != 0 || header[38] != 1 || header[39] == 0 || header[40] != (map[string]uint64{"Go": 1, "C": 2}[backend]) {
		return fmt.Errorf("actual root/backend/fixed41header/validation")
	}
	var state sdChildResult40A970
	p, e = os.ReadFile(filepath.Join(dir, "actual-input-state.json"))
	if e != nil {
		return e
	}
	if e = json.Unmarshal(p, &state); e != nil {
		return e
	}
	if state.PID != pid || state.Backend != backend || state.ControlWord != cw {
		return fmt.Errorf("durable native input consumption identity")
	}
	result := &state
	if exit == 0 {
		var r sdChildResult40A970
		p, e = os.ReadFile(filepath.Join(dir, "actual-return.json"))
		if e != nil {
			return e
		}
		if e = json.Unmarshal(p, &r); e != nil {
			return e
		}
		if r.PID != pid || r.Backend != backend || r.ControlWord != cw || header[2] != 2 || !reflect.DeepEqual(r.UsedReads, state.UsedReads) || !reflect.DeepEqual(r.UsedWrites, state.UsedWrites) || !reflect.DeepEqual(r.UsedServices, state.UsedServices) || !reflect.DeepEqual(r.UsedResponses, state.UsedResponses) {
			return fmt.Errorf("natural actual void return/header/state")
		}
		result = &r
	} else if exit == 172 {
		if header[2] != 5 || (header[3] != 10 && header[3] != 11) || header[20] == 0 || header[20] != header[21] || header[22] != 0 || header[24] != 0 || header[10] != 3 || header[6] == 0 || header[8] != 4 || header[5] < header[7] || header[5]-header[7] >= header[8] {
			return fmt.Errorf("actual unqualified-PC status DWORD readfault capture")
		}
	} else {
		return fmt.Errorf("unexpected actual native exit%d", exit)
	}
	trace, e := sdReadTrace40A970(filepath.Join(dir, "actual-observations.jsonl"))
	if e != nil {
		return e
	}
	if exit == 172 {
		if len(trace) == 0 {
			return fmt.Errorf("missing pending actual read")
		}
		last := trace[len(trace)-1]
		if last.Kind != "fault" || last.Site != 0x40a9c5 || uint64(last.Site) != header[9] || last.Address != header[7] || uint64(last.Width) != header[8] {
			return fmt.Errorf("actual fault binds pending status operand")
		}
	}
	if (row.Status == "returned") != (exit == 0) {
		return fmt.Errorf("actual original/native terminal domain differs CW%04x", cw)
	}
	differences, e := sdCompare40A970(row, &meta, trace, before, after, result)
	if e != nil {
		return fmt.Errorf("CW%04x: %w", cw, e)
	}
	return rskWriteJSON4E23C0(filepath.Join(dir, "actual-qualification.json"), map[string]any{"qualified_selected_observations": true, "actual_pid": pid, "backend": backend, "input": item.Input.Name, "input_sha256": rskHash4E23C0(item.Raw), "control_word": cw, "original_row_sha256": row.RawSHA, "actual_original_CW_comparisons": 1, "actual_native_calls": 1, "native_before_sha256": rskHash4E23C0(before), "native_after_sha256": rskHash4E23C0(after), "whole_foreign_banks_and_canaries": true, "public_four_DWORDs_and_player_status_and_service_order": true, "public_return_kind": "void", "private_numeric_store_bit_differences_observed": differences, "private_FSTP_bit_equality_required": false, "native_numeric_bits_retained": true, "original_helper_body_or_physical_ECX_EDX_equivalence": false, "diagnostic_cells_are_not_production_owner_aliases": true, "physical_nil_EA_equivalence": false, "independent_native_fault_PC_qualification": false, "native_Go_heap_credit": false, "native_private_stack_or_physical_FPU_equivalence": false, "signal_header": header})
}

func TestSuddenDeathNativeOriginal40A970(t *testing.T) {
	if os.Getenv(sdWorkerEnv40A970) == "1" {
		return
	}
	cases, e := sdLoadCases40A970()
	if e != nil {
		t.Fatal(e)
	}
	initial := os.Getenv("OPENNOX_RAW_DEFAULT_DAMAGE_INITIAL_ARTIFACTS")
	if initial == "" {
		t.Fatal("prepared unique INITIAL_ARTIFACTS required")
	}
	root, e := filepath.Abs(filepath.Join(initial, "sudden-death-40a970"))
	if e != nil {
		t.Fatal(e)
	}
	if e = os.MkdirAll(filepath.Dir(root), 0700); e != nil {
		t.Fatal(e)
	}
	if e = os.Mkdir(root, 0700); e != nil {
		t.Fatal("unique owned native corpus directory", e)
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
					if e := sdRun40A970(filepath.Join(root, fmt.Sprintf("%03d-%04x-%s", i, cw, backend)), exe, sha, backend, cw, item); e != nil {
						t.Fatal(e)
					}
				}) {
					passed++
				}
			}
		}
	}
	if e = rskWriteJSON4E23C0(filepath.Join(root, "actual-corpus-summary.json"), map[string]any{"inputs": 403, "attempted_native_calls": attempted, "passed": passed, "failed": attempted - passed, "stream_complete": attempted == 1612, "qualified_original_CW_comparisons": passed, "actual_native_backends": []string{"Go", "C"}, "independently_selected_control_words": []uint16{0x0e3f, 0x023f}, "input403_sha256": sdInputSHA40A970, "original806_sha256": sdReferenceSHA40A970, "private_FSTP_bit_equality_required": false, "native_numeric_bits_retained": true, "native_private_stack_FPU_PC_and_physical_nil_EA_equivalence": false, "external_helper_body_credit": false}); e != nil {
		t.Fatal(e)
	}
	if attempted != 1612 || passed != 1612 {
		t.Fatalf("native corpus attempted=%d passed=%d", attempted, passed)
	}
	t.Logf("SUDDEN_DEATH_NATIVE_ORIGINAL_OK inputs=403 native=1612 matching-original-CW-comparisons=1612 public-state-and-service-exact=true private-FSTP-NaN-payload-physical-credit=false")
}
