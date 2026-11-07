//go:build darwin && arm64

package legacy

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

const rskInputSHA4E23C0 = "a345b89795dd111b843f5bde10deb8bf49ec6546e627222d54b02a90d5e13a16"
const rskReferenceSHA4E23C0 = "77a1721674ad583738af9e30bf139d4ca7f00711f4178387d7ae026f96ecb2b0"
const rskChildEnv4E23C0 = "OPENNOX_RAW_SKELETON_NATIVE_CHILD"
const rskDirectoryEnv4E23C0 = "OPENNOX_RAW_SKELETON_NATIVE_CHILD_DIRECTORY"
const rskLogsBudget4E23C0 = 10 << 20

type rskMemory4E23C0 struct {
	Name      string `json:"name"`
	Address   uint32 `json:"address"`
	Size      int    `json:"size"`
	Before    string `json:"before_hex"`
	After     string `json:"after_hex"`
	BeforeSHA string `json:"before_sha256"`
	AfterSHA  string `json:"after_sha256"`
}
type rskEvent4E23C0 struct {
	Kind      string            `json:"kind"`
	Sequence  int               `json:"sequence"`
	PC        uint32            `json:"PC"`
	Address   uint32            `json:"address"`
	Width     int               `json:"width"`
	Value     string            `json:"value_hex"`
	Completed bool              `json:"completed"`
	Registers map[string]uint32 `json:"registers"`
}
type rskServiceReference4E23C0 struct {
	Site          uint32   `json:"site"`
	Args          []uint32 `json:"arguments_u32"`
	Sequence      int      `json:"actual_CALL_write_sequence"`
	LocalSequence *int     `json:"direction_local8_host_read_sequence"`
}
type rskReference4E23C0 struct {
	Name         string                      `json:"name"`
	Mode         string                      `json:"mode"`
	CW           uint16                      `json:"control_word"`
	InputSHA     string                      `json:"input_case_sha256"`
	Status       string                      `json:"status"`
	Result       *uint32                     `json:"actual_result_u32"`
	Valid        bool                        `json:"certificate_valid"`
	Exact        bool                        `json:"whole_memory_reconstruction_exact"`
	Events       []rskEvent4E23C0            `json:"events"`
	Services     []rskServiceReference4E23C0 `json:"external_services"`
	Memory       []rskMemory4E23C0           `json:"complete_memory"`
	UsedReads    []int                       `json:"used_after_read_stimuli"`
	UsedServices []int                       `json:"used_after_service_stimuli"`
	Consumed     []uint32                    `json:"consumed_service_sites"`
}
type rskCase4E23C0 struct {
	Input      rskInput4E23C0
	Raw        json.RawMessage
	References []*rskReference4E23C0
}
type rskArgumentTrace4E23C0 struct {
	Pointer   uint64 `json:"pointer"`
	Word      uint32 `json:"word"`
	IsPointer bool   `json:"is_pointer"`
}
type rskTrace4E23C0 struct {
	Kind             string                   `json:"kind"`
	Site             uint32                   `json:"site"`
	Complete         bool                     `json:"complete"`
	PE               uint32                   `json:"pe"`
	PEWidth          int                      `json:"pe_width"`
	Address          uint64                   `json:"address"`
	EffectiveAddress uint64                   `json:"effective_address"`
	Base             uint64                   `json:"base"`
	Offset           uint64                   `json:"offset"`
	Width            int                      `json:"width"`
	Word             uint32                   `json:"word"`
	Pointer          uint64                   `json:"pointer"`
	IsPointer        bool                     `json:"is_pointer"`
	Arguments        []rskArgumentTrace4E23C0 `json:"arguments,omitempty"`
	DirectionBytes   string                   `json:"direction_local8_hex,omitempty"`
	Reply            uint32                   `json:"reply,omitempty"`
}
type rskChildInput4E23C0 struct {
	Input         rskInput4E23C0 `json:"input"`
	Raw           []byte         `json:"raw_input_base64"`
	Backend       string         `json:"backend"`
	ExecutableSHA string         `json:"executable_sha256"`
}
type rskChildResult4E23C0 struct {
	PID          int      `json:"pid"`
	Backend      string   `json:"backend"`
	Return       uint32   `json:"return_u32"`
	UsedReads    []int    `json:"used_after_reads"`
	UsedServices []int    `json:"used_after_services"`
	Consumed     []uint32 `json:"consumed_service_sites"`
	TraceEvents  int      `json:"trace_events"`
	ResultScope  string   `json:"scope"`
}

func rskFileSHA4E23C0(path string) (string, error) {
	f, e := os.Open(path)
	if e != nil {
		return "", e
	}
	defer f.Close()
	h := sha256.New()
	if _, e = io.Copy(h, f); e != nil {
		return "", e
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
func rskOpenGzip4E23C0(path, sha string) (*os.File, *gzip.Reader, error) {
	actual, e := rskFileSHA4E23C0(path)
	if e != nil || actual != sha {
		return nil, nil, fmt.Errorf("immutable fixture SHA %s: %s %v", path, actual, e)
	}
	f, e := os.Open(path)
	if e != nil {
		return nil, nil, e
	}
	z, e := gzip.NewReader(f)
	if e != nil {
		f.Close()
		return nil, nil, e
	}
	return f, z, nil
}
func rskLoadCases4E23C0() ([]rskCase4E23C0, error) {
	base := filepath.Join("testdata", "retained_skeleton_4e23c0")
	f, z, e := rskOpenGzip4E23C0(filepath.Join(base, "inputs-v1.json.gz"), rskInputSHA4E23C0)
	if e != nil {
		return nil, e
	}
	var doc struct {
		Cases []json.RawMessage `json:"cases"`
	}
	e = json.NewDecoder(z).Decode(&doc)
	ze := z.Close()
	fe := f.Close()
	if e != nil || ze != nil || fe != nil {
		return nil, fmt.Errorf("input gzip decode: %v %v %v", e, ze, fe)
	}
	if len(doc.Cases) != 251 {
		return nil, fmt.Errorf("independent 251 inputs")
	}
	cases := make([]rskCase4E23C0, 251)
	indices := make(map[string]int)
	for i, raw := range doc.Cases {
		if e = json.Unmarshal(raw, &cases[i].Input); e != nil {
			return nil, e
		}
		var compact bytes.Buffer
		if e = json.Compact(&compact, raw); e != nil {
			return nil, e
		}
		cases[i].Raw = append(json.RawMessage(nil), compact.Bytes()...)
		name := cases[i].Input.Name
		if _, ok := indices[name]; ok {
			return nil, fmt.Errorf("duplicate independent input")
		}
		indices[name] = i
	}
	f, z, e = rskOpenGzip4E23C0(filepath.Join(base, "reference-v2.json.gz"), rskReferenceSHA4E23C0)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	defer z.Close()
	d := json.NewDecoder(z)
	token, e := d.Token()
	if e != nil || token != json.Delim('{') {
		return nil, fmt.Errorf("reference object")
	}
	count := 0
	found := false
	for d.More() {
		key, e := d.Token()
		if e != nil {
			return nil, e
		}
		if key != "rows" {
			var discarded json.RawMessage
			if e = d.Decode(&discarded); e != nil {
				return nil, e
			}
			continue
		}
		if found {
			return nil, fmt.Errorf("duplicate rows")
		}
		found = true
		token, e = d.Token()
		if e != nil || token != json.Delim('[') {
			return nil, fmt.Errorf("reference rows array")
		}
		for d.More() {
			var row rskReference4E23C0
			if e = d.Decode(&row); e != nil {
				return nil, e
			}
			i, ok := indices[row.Name]
			if !ok || !row.Valid || !row.Exact || row.InputSHA != rskHash4E23C0(cases[i].Raw) {
				return nil, fmt.Errorf("original row input/certificate binding")
			}
			if row.CW != 0x0e3f && row.CW != 0x023f {
				return nil, fmt.Errorf("original exact CW")
			}
			if row.Status != "returned" && row.Status != "original-fault" {
				return nil, fmt.Errorf("original terminal status")
			}
			for _, old := range cases[i].References {
				if old.CW == row.CW {
					return nil, fmt.Errorf("duplicate original case/CW")
				}
			}
			cases[i].References = append(cases[i].References, &row)
			count++
		}
		if _, e = d.Token(); e != nil {
			return nil, e
		}
	}
	if _, e = d.Token(); e != nil {
		return nil, e
	}
	var tail any
	if e = d.Decode(&tail); e != io.EOF {
		return nil, fmt.Errorf("reference trailing data/CRC: %v", e)
	}
	if !found || count != 502 {
		return nil, fmt.Errorf("actual original 502 rows")
	}
	for i := range cases {
		if len(cases[i].References) != 2 {
			return nil, fmt.Errorf("two original CW observations per input")
		}
	}
	return cases, nil
}

func rskPEOperand4E23C0(g *rskGraph4E23C0, event rawSkeletonOperand4E23C0) (uint32, int, error) {
	offsets := map[uint32]struct {
		pe    uint32
		width int
		index int
	}{0x4e2413: {72, 4, 1}, 0x4e2416: {76, 4, 2}, 0x4e2423: {72, 4, 1}, 0x4e2426: {76, 4, 2}, 0x4e2431: {124, 2, 3}, 0x4e244b: {748, 4, 4}, 0x4e245f: {480, 1, 5}, 0x4e2465: {481, 1, 6}}
	desc, ok := offsets[event.Site]
	if !ok {
		return 0, 0, fmt.Errorf("unsupported actual operand site")
	}
	expectedWidth := desc.width
	if event.IsPointer {
		expectedWidth = int(unsafe.Sizeof(unsafe.Pointer(nil)))
	}
	if event.Offset != g.Meta.Offsets[desc.index] || event.Width != expectedWidth || event.IsPointer != (event.Site == 0x4e244b) {
		return 0, 0, fmt.Errorf("actual operand named offset/type/width")
	}
	token := uint32(0)
	found := event.Base == nil
	for pe, p := range g.Tokens {
		if p == event.Base {
			if found {
				return 0, 0, fmt.Errorf("native identity collision")
			}
			token = pe
			found = true
		}
	}
	if !found {
		return 0, 0, fmt.Errorf("undeclared actual operand base")
	}
	numeric := uintptr(event.Base) + event.Offset
	if event.EffectiveAddress != numeric {
		return 0, 0, fmt.Errorf("numeric named effective address")
	}
	if event.Base == nil {
		if event.Address != nil {
			return 0, 0, fmt.Errorf("nil must not enter low unsafe.Pointer")
		}
	} else if uintptr(event.Address) != numeric {
		return 0, 0, fmt.Errorf("actual foreign operand EA")
	}
	return token + desc.pe, desc.width, nil
}

// No original observation is opened by a child. Responses/effects are only
// independent inputs. The parent consumes authentic original rows after Wait.
func TestRawSkeletonNativeWorker4E23C0(t *testing.T) {
	if os.Getenv(rskChildEnv4E23C0) != "1" {
		return
	}
	directory := os.Getenv(rskDirectoryEnv4E23C0)
	if !filepath.IsAbs(directory) {
		t.Fatal("absolute owned child directory")
	}
	var child rskChildInput4E23C0
	p, e := os.ReadFile(filepath.Join(directory, "input.json"))
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(p, &child); e != nil {
		t.Fatal(e)
	}
	var rawInput rskInput4E23C0
	if e = json.Unmarshal(child.Raw, &rawInput); e != nil || !reflect.DeepEqual(rawInput, child.Input) {
		t.Fatal("exact input bytes/typed projection", e)
	}
	executable, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	sha, e := rskFileSHA4E23C0(executable)
	if e != nil || sha != child.ExecutableSHA {
		t.Fatal("same actual PREEXEC main bytes", e)
	}
	g, e := rskBuildGraph4E23C0(&child.Input, filepath.Join(directory, "actual-shared-banks.bin"))
	if e != nil {
		t.Fatal(e)
	}
	defer g.close()
	if e = rskWriteJSON4E23C0(filepath.Join(directory, "graph.json"), g.Meta); e != nil {
		t.Fatal(e)
	}
	before := g.snapshot()
	if e = os.WriteFile(filepath.Join(directory, "actual-native-BEFORE.bin"), before, 0600); e != nil {
		t.Fatal(e)
	}
	trace, e := os.OpenFile(filepath.Join(directory, "actual-observations.jsonl"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer trace.Close()
	enc := json.NewEncoder(trace)
	events := 0
	traceBytes := 0
	emit := func(value rskTrace4E23C0) {
		data, e := json.Marshal(value)
		if e != nil {
			panic(e)
		}
		events++
		traceBytes += len(data) + 1
		if events > 64 || traceBytes > 16384 {
			panic("bounded actual Skeleton trace")
		}
		if e = enc.Encode(value); e != nil {
			panic(e)
		}
	}
	must := func(e error) {
		if e != nil {
			rawDefaultDamageSharedValidationFailed4E0B30(g.Header, uint64(events))
			panic(e)
		}
	}
	saveState := func() {
		must(rskWriteJSON4E23C0(filepath.Join(directory, "actual-input-state.json"), rskChildResult4E23C0{PID: os.Getpid(), Backend: child.Backend, UsedReads: g.UsedReads, UsedServices: g.UsedServices, Consumed: g.UsedPackets, TraceEvents: events, ResultScope: "actual input-effect consumption; no original observation schedule supplied"}))
	}
	if e = rawDefaultDamageSharedBegin4E0B30(g.Header); e != nil {
		t.Fatal(e)
	}
	defer func() {
		if e := rawDefaultDamageSharedEnd4E0B30(); e != nil {
			t.Error(e)
		}
	}()
	observer := &rawSkeletonObserver4E23C0{}
	observer.Operand = func(event rawSkeletonOperand4E23C0) {
		pe, width, e := rskPEOperand4E23C0(g, event)
		must(e)
		if !event.Complete {
			rawDefaultDamageSharedProgress4E0B30(g.Header, uint64(events), event.Site, event.Address, event.Width, rawDefaultDamageSharedRead4E0B30, nil)
		}
		emit(rskTrace4E23C0{Kind: "operand", Site: event.Site, Complete: event.Complete, PE: pe, PEWidth: width, Address: uint64(uintptr(event.Address)), EffectiveAddress: uint64(event.EffectiveAddress), Base: uint64(uintptr(event.Base)), Offset: uint64(event.Offset), Width: event.Width, Word: event.Word, Pointer: uint64(uintptr(event.Pointer)), IsPointer: event.IsPointer})
		if event.Complete {
			must(g.afterRead(event.Site, pe, width))
			saveState()
		}
	}
	observer.Frame = func() uint8 {
		address, f, e := g.address(0x84ea04, 1)
		must(e)
		rawDefaultDamageSharedProgress4E0B30(g.Header, uint64(events), 0x4e23da, address, 1, rawDefaultDamageSharedRead4E0B30, nil)
		base := uint64(uintptr(g.Pointers[f.Region]))
		entry := rskTrace4E23C0{Kind: "operand", Site: 0x4e23da, PE: 0x84ea04, PEWidth: 1, Address: uint64(uintptr(address)), EffectiveAddress: uint64(uintptr(address)), Base: base, Offset: uint64(f.NativeOffset), Width: 1}
		emit(entry)
		value := *(*uint8)(address)
		entry.Word = uint32(value)
		entry.Complete = true
		emit(entry)
		must(g.afterRead(0x4e23da, 0x84ea04, 1))
		saveState()
		return value
	}
	observer.Service = func(call rawSkeletonService4E23C0) (uint32, bool) {
		packet, ok := g.Input.Services[fmt.Sprintf("%08x", call.Site)]
		if !ok || len(call.Arguments) != packet.Argc || g.Services[call.Site] != 0 {
			must(fmt.Errorf("independent one-shot service ABI/input"))
		}
		record := rskTrace4E23C0{Kind: "service", Site: call.Site, Complete: true, Reply: packet.Registers["eax"]}
		for _, argument := range call.Arguments {
			record.Arguments = append(record.Arguments, rskArgumentTrace4E23C0{Pointer: uint64(uintptr(argument.Pointer)), Word: argument.Word, IsPointer: argument.IsPointer})
		}
		if call.Site == 0x4e243f {
			p := call.Arguments[2].Pointer
			if p == nil {
				must(fmt.Errorf("actual local vector address"))
			}
			sample := append([]byte(nil), unsafe.Slice((*byte)(p), 8)...)
			record.DirectionBytes = hex.EncodeToString(sample)
			runtime.KeepAlive(p)
		}
		emit(record)
		g.UsedPackets = append(g.UsedPackets, call.Site)
		must(g.apply(packet.Writes))
		must(g.afterService(call.Site))
		saveState()
		return packet.Registers["eax"], true
	}
	pop := pushRawSkeletonObserver4E23C0(observer)
	defer pop()
	target, e := g.pointer(g.Input.Args[0])
	must(e)
	source, e := g.pointer(g.Input.Args[1])
	must(e)
	weapon, e := g.pointer(g.Input.Args[2])
	must(e)
	backend := uint64(1)
	entry := reflect.ValueOf(rawSkeletonDamageNative4E23C0).Pointer()
	if child.Backend == "C" {
		backend = 2
		entry = reflect.ValueOf(rawSkeletonFixtureCall4E23C0).Pointer()
	} else if child.Backend != "Go" {
		t.Fatal("actual backend")
	}
	rawDefaultDamageSharedRootDispatched4E0B30(g.Header, entry, backend)
	saveState()
	must(rskWriteJSON4E23C0(filepath.Join(directory, "actual-child-PREEXEC.json"), map[string]any{"pid": os.Getpid(), "parent_pid": os.Getppid(), "backend": child.Backend, "executable_sha256": sha, "input_sha256": rskHash4E23C0(child.Raw), "BEFORE_sha256": rskHash4E23C0(before), "graph": g.Meta, "prepared_before_actual_root": true, "original_observations_passed_to_child": false}))
	var result int32
	if child.Backend == "Go" {
		result = rawSkeletonDamageNative4E23C0(target, source, weapon, int32(g.Input.Args[3]), int32(g.Input.Args[4]))
	} else {
		result = rawSkeletonFixtureCall4E23C0(target, source, weapon, int32(g.Input.Args[3]), int32(g.Input.Args[4]))
	}
	must(trace.Sync())
	must(rskWriteJSON4E23C0(filepath.Join(directory, "actual-return.json"), rskChildResult4E23C0{PID: os.Getpid(), Backend: child.Backend, Return: uint32(result), UsedReads: g.UsedReads, UsedServices: g.UsedServices, Consumed: g.UsedPackets, TraceEvents: events, ResultScope: "real caller with independently chosen external boundary inputs; helper/product/campaign service execution credit zero"}))
	runtime.KeepAlive(g)
}

// Bounded output is retained in full; exceeding the budget rejects, it never
// silently truncates a log and then qualifies a successful child.
type rskBoundedLog4E23C0 struct {
	File     *os.File
	N        int
	Exceeded bool
}

func (w *rskBoundedLog4E23C0) Write(p []byte) (int, error) {
	if w.N+len(p) > 8192 {
		w.Exceeded = true
		return 0, fmt.Errorf("bounded child log exceeded")
	}
	n, e := w.File.Write(p)
	w.N += n
	return n, e
}
func rskArchiveMain4E23C0(executable, directory string) (string, error) {
	sha, e := rskFileSHA4E23C0(executable)
	if e != nil {
		return "", e
	}
	archive := filepath.Join(directory, "actual-main-before-first-child.zip")
	f, e := os.OpenFile(archive, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return "", e
	}
	z := zip.NewWriter(f)
	member, e := z.Create("actual-native-test-main")
	if e != nil {
		return "", e
	}
	source, e := os.Open(executable)
	if e != nil {
		return "", e
	}
	_, e = io.Copy(member, source)
	source.Close()
	if e != nil {
		return "", e
	}
	if e = z.Close(); e != nil {
		return "", e
	}
	if e = f.Close(); e != nil {
		return "", e
	}
	zr, e := zip.OpenReader(archive)
	if e != nil {
		return "", e
	}
	defer zr.Close()
	if len(zr.File) != 1 {
		return "", fmt.Errorf("one actual main archive member")
	}
	r, e := zr.File[0].Open()
	if e != nil {
		return "", e
	}
	h := sha256.New()
	_, e = io.Copy(h, r)
	r.Close()
	if e != nil || hex.EncodeToString(h.Sum(nil)) != sha {
		return "", fmt.Errorf("byte-exact actual main PREEXEC archive")
	}
	info, _ := debug.ReadBuildInfo()
	if e = rskWriteJSON4E23C0(filepath.Join(directory, "actual-main-PREEXEC.json"), map[string]any{"path": executable, "sha256": sha, "archive": archive, "archive_member": "actual-native-test-main", "archive_sha256": func() string { s, _ := rskFileSHA4E23C0(archive); return s }(), "build_info": info, "runtime_version": runtime.Version(), "archived_before_first_child": true}); e != nil {
		return "", e
	}
	return sha, nil
}
func rskGzipOwnedDuplicate4E23C0(path string) error {
	info, e := os.Lstat(path)
	if e != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("owned duplicate must be a regular non-symlink file")
	}
	original, e := os.ReadFile(path)
	if e != nil {
		return e
	}
	f, e := os.OpenFile(path+".gz", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	z := gzip.NewWriter(f)
	if _, e = z.Write(original); e != nil {
		f.Close()
		return e
	}
	if e = z.Close(); e != nil {
		f.Close()
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	rf, e := os.Open(path + ".gz")
	if e != nil {
		return e
	}
	rz, e := gzip.NewReader(rf)
	if e != nil {
		rf.Close()
		return e
	}
	restored, e := io.ReadAll(rz)
	rz.Close()
	rf.Close()
	if e != nil || !bytes.Equal(original, restored) {
		return fmt.Errorf("byte-exact owned native bank compression")
	}
	current, e := os.ReadFile(path)
	if e != nil || !bytes.Equal(current, original) {
		return fmt.Errorf("owned raw changed before cleanup")
	}
	if e = os.Remove(path); e != nil {
		return e
	}
	return rskWriteJSON4E23C0(path+".raw-cleanup.json", map[string]any{"path": path, "bytes": len(original), "sha256": rskHash4E23C0(original), "gzip_verified_byte_exact": true, "child_Wait_terminal": true, "removed_only_owned_duplicate": true})
}

type rskStep4E23C0 struct {
	Kind     string
	Sequence int
	Event    rskEvent4E23C0
	Service  rskServiceReference4E23C0
}

func rskSteps4E23C0(row *rskReference4E23C0) []rskStep4E23C0 {
	sites := map[uint32]bool{0x4e23da: true, 0x4e2413: true, 0x4e2416: true, 0x4e2423: true, 0x4e2426: true, 0x4e2431: true, 0x4e244b: true, 0x4e245f: true, 0x4e2465: true}
	var out []rskStep4E23C0
	for _, e := range row.Events {
		if sites[e.PC] && ((e.Kind == "read" && e.Completed) || e.Kind == "fault") {
			kind := "operand"
			if e.Kind == "fault" {
				kind = "fault"
			}
			out = append(out, rskStep4E23C0{Kind: kind, Sequence: e.Sequence, Event: e})
		}
	}
	for _, s := range row.Services {
		out = append(out, rskStep4E23C0{Kind: "service", Sequence: s.Sequence, Service: s})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Sequence < out[j].Sequence })
	return out
}

// Fixed widths come from the sealed MOV/TEST operands, not invalid-hook chunks.
func rskSemanticWidth4E23C0(site uint32) int {
	switch site {
	case 0x4e23da, 0x4e245f, 0x4e2465:
		return 1
	case 0x4e2431:
		return 2
	case 0x4e2413, 0x4e2416, 0x4e2423, 0x4e2426, 0x4e244b:
		return 4
	}
	return 0
}
func rskNativeIdentity4E23C0(meta *rskGraphMetadata4E23C0, token uint32) (uint64, error) {
	if token == 0 {
		return 0, nil
	}
	v, ok := meta.Identities[strconv.FormatUint(uint64(token), 10)]
	if !ok {
		return 0, fmt.Errorf("canonical reference pointer identity %08x", token)
	}
	return v, nil
}
func rskReadTrace4E23C0(path string) ([]rskTrace4E23C0, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	info, e := f.Stat()
	if e != nil || info.Size() > 16384 {
		return nil, fmt.Errorf("bounded actual trace extent")
	}
	d := json.NewDecoder(f)
	var out []rskTrace4E23C0
	var pending *rskTrace4E23C0
	for {
		var v rskTrace4E23C0
		e = d.Decode(&v)
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, e
		}
		if v.Kind == "operand" && !v.Complete {
			if pending != nil {
				return nil, fmt.Errorf("duplicate unmatched before operand")
			}
			copy := v
			pending = &copy
			continue
		}
		if v.Kind == "operand" {
			if pending == nil || v.Site != pending.Site || v.PE != pending.PE || v.Address != pending.Address || v.Width != pending.Width || v.IsPointer != pending.IsPointer {
				return nil, fmt.Errorf("actual before/after read pairing")
			}
			pending = nil
		} else if v.Kind != "service" || pending != nil {
			return nil, fmt.Errorf("actual trace event ordering")
		}
		out = append(out, v)
	}
	if pending != nil {
		pending.Kind = "fault"
		out = append(out, *pending)
	}
	return out, nil
}
func rskCompareReference4E23C0(in *rskInput4E23C0, row *rskReference4E23C0, meta *rskGraphMetadata4E23C0, trace []rskTrace4E23C0, before, after []byte, result *rskChildResult4E23C0) error {
	steps := rskSteps4E23C0(row)
	if len(steps) != len(trace) {
		return fmt.Errorf("actual/reference logical observations %d != %d", len(trace), len(steps))
	}
	for i, s := range steps {
		actual := trace[i]
		if actual.Kind != s.Kind {
			return fmt.Errorf("ordered observation kind %d", i)
		}
		if s.Kind == "service" {
			s := s.Service
			if actual.Site != s.Site || len(actual.Arguments) != len(s.Args) {
				return fmt.Errorf("actual service site/arity %d", i)
			}
			packet := in.Services[fmt.Sprintf("%08x", s.Site)]
			if actual.Reply != packet.Registers["eax"] {
				return fmt.Errorf("actual independent boundary reply")
			}
			masks := map[uint32][]bool{0x4e23ce: {true, false}, 0x4e23ea: {false, true, false, false}, 0x4e243f: {true, false, true}, 0x4e2452: {true}, 0x4e247b: {false, true, false, false}, 0x4e249d: {true, true, true, false, false}}
			mask, ok := masks[s.Site]
			if !ok || len(mask) != len(s.Args) {
				return fmt.Errorf("independent service pointer ABI")
			}
			for j, arg := range actual.Arguments {
				if arg.IsPointer != mask[j] {
					return fmt.Errorf("actual service argument kind")
				}
				if !arg.IsPointer {
					if arg.Word != s.Args[j] {
						return fmt.Errorf("raw DWORD service argument %d/%d", i, j)
					}
					continue
				}
				if s.Site == 0x4e243f && j == 2 {
					if arg.Pointer <= 0xffffffff || len(rskHex4E23C0(actual.DirectionBytes)) != 8 {
						return fmt.Errorf("actual local8 full address/extent")
					}
					var expected string
					for _, e := range row.Events {
						if s.LocalSequence != nil && e.Sequence == *s.LocalSequence {
							expected = e.Value
						}
					}
					if expected == "" || actual.DirectionBytes != expected {
						return fmt.Errorf("actual copied direction X/Y bytes")
					}
					continue
				}
				var pointer uint64
				var e error
				if s.Site == 0x4e243f && j == 0 {
					if s.Args[j] != in.Args[0]+56 {
						return fmt.Errorf("actual original Direction target-position argument")
					}
					target, e := rskNativeIdentity4E23C0(meta, in.Args[0])
					if e != nil {
						return e
					}
					pointer = target + uint64(meta.Offsets[0])
				} else {
					pointer, e = rskNativeIdentity4E23C0(meta, s.Args[j])
					if e != nil {
						return e
					}
				}
				if arg.Pointer != pointer {
					return fmt.Errorf("full native service pointer %d/%d", i, j)
				}
			}
			continue
		}
		e := s.Event
		if actual.Site != e.PC || actual.PEWidth != rskSemanticWidth4E23C0(e.PC) {
			return fmt.Errorf("ordered canonical operand %d site/width", i)
		}
		if s.Kind == "fault" {
			if actual.Complete {
				return fmt.Errorf("faulting read cannot be marked complete")
			}
			base, offset := uint32(0), uint32(0)
			switch e.PC {
			case 0x4e2413:
				base, offset = e.Registers["ebp"], 72
			case 0x4e2416:
				base, offset = e.Registers["ebp"], 76
			case 0x4e2423:
				base, offset = e.Registers["ebx"], 72
			case 0x4e2426:
				base, offset = e.Registers["ebx"], 76
			case 0x4e2431:
				base, offset = e.Registers["esi"], 124
			case 0x4e244b:
				base, offset = e.Registers["esi"], 748
			case 0x4e245f:
				base, offset = e.Registers["edi"], 480
			case 0x4e2465:
				base, offset = e.Registers["edi"], 481
			default:
				return fmt.Errorf("unsupported original fault operand")
			}
			logical := base + offset
			if actual.PE != logical || uint64(e.Address) < uint64(logical) || uint64(e.Address) >= uint64(logical)+uint64(rskSemanticWidth4E23C0(e.PC)) || e.Width <= 0 || e.Width > 8 {
				return fmt.Errorf("actual logical fault operand and raw original invalid-byte extent")
			}
			continue
		}
		if actual.PE != e.Address {
			return fmt.Errorf("actual completed PE operand identity")
		}
		if !actual.Complete {
			return fmt.Errorf("original completed read missing")
		}
		if actual.IsPointer {
			pointer, e := rskNativeIdentity4E23C0(meta, rskWord4E23C0(rskHex4E23C0(e.Value)))
			if e != nil {
				return e
			}
			if actual.Pointer != pointer || actual.Width != int(unsafe.Sizeof(unsafe.Pointer(nil))) {
				return fmt.Errorf("all native pointer bits")
			}
		} else if actual.Word != rskWord4E23C0(rskHex4E23C0(e.Value)) {
			return fmt.Errorf("raw scalar load %d", i)
		}
	}
	pages := make(map[uint32][]byte)
	for _, m := range row.Memory {
		if m.After == "" {
			continue
		}
		data := rskHex4E23C0(m.After)
		if len(data) != m.Size || rskHash4E23C0(data) != m.AfterSHA {
			return fmt.Errorf("original whole AFTER bytes/SHA")
		}
		pages[m.Address] = data
	}
	expected, e := rskExpectedAfter4E23C0(meta, before, pages)
	if e != nil {
		return e
	}
	if !bytes.Equal(expected, after) {
		return fmt.Errorf("whole foreign native AFTER/canaries differ from original field projection")
	}
	if result == nil {
		return fmt.Errorf("actual durable input-consumption state")
	}
	if !reflect.DeepEqual(rskInts4E23C0(result.UsedReads), rskInts4E23C0(row.UsedReads)) || !reflect.DeepEqual(rskInts4E23C0(result.UsedServices), rskInts4E23C0(row.UsedServices)) || !reflect.DeepEqual(result.Consumed, row.Consumed) {
		return fmt.Errorf("consumed vs unused independent schedules")
	}
	if row.Status == "returned" {
		if row.Result == nil || result.Return != *row.Result {
			return fmt.Errorf("raw signed32 caller return bits")
		}
	}
	return nil
}
func rskInts4E23C0(v []int) []int {
	if v == nil {
		return []int{}
	}
	return v
}

func rskRunCase4E23C0(directory, executable, sha, backend string, item *rskCase4E23C0) (err error) {
	if e := os.Mkdir(directory, 0700); e != nil {
		return e
	}
	if e := rskWriteJSON4E23C0(filepath.Join(directory, "input.json"), rskChildInput4E23C0{Input: item.Input, Raw: item.Raw, Backend: backend, ExecutableSHA: sha}); e != nil {
		return e
	}
	stdout, e := os.OpenFile(filepath.Join(directory, "stdout.log"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	defer stdout.Close()
	stderr, e := os.OpenFile(filepath.Join(directory, "stderr.log"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	defer stderr.Close()
	outLog, errLog := &rskBoundedLog4E23C0{File: stdout}, &rskBoundedLog4E23C0{File: stderr}
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, executable, "-test.run=^TestRawSkeletonNativeWorker4E23C0$", "-test.count=1", "-test.timeout=10s")
	command.Env = append(os.Environ(), rskChildEnv4E23C0+"=1", rskDirectoryEnv4E23C0+"="+directory)
	command.Stdout = outLog
	command.Stderr = errLog
	var forced atomic.Bool
	defaultCancel := command.Cancel
	command.Cancel = func() error { forced.Store(true); return defaultCancel() }
	start := time.Now().UTC()
	startErr := command.Start()
	pid := 0
	if command.Process != nil {
		pid = command.Process.Pid
	}
	var waitErr error
	if startErr == nil {
		waitErr = command.Wait()
	}
	end := time.Now().UTC()
	exit := -1
	if command.ProcessState != nil {
		exit = command.ProcessState.ExitCode()
	}
	terminal := map[string]any{"pid": pid, "backend": backend, "start_UTC": start, "Wait_terminal_UTC": end, "start_error": fmt.Sprint(startErr), "wait_error": fmt.Sprint(waitErr), "exit_code": exit, "forced_request": forced.Load(), "context_error": fmt.Sprint(ctx.Err()), "stdout_exceeded": outLog.Exceeded, "stderr_exceeded": errLog.Exceeded, "executable_sha256": sha, "reference_CW_comparisons": 2, "original_execution_credit": 0}
	if e = rskWriteJSON4E23C0(filepath.Join(directory, "actual-child-Wait.json"), terminal); e != nil {
		return e
	}
	// Archive exactly observed raw banks even when subsequent qualification
	// fails. Only byte-verified owned duplicates are removed, after Wait.
	defer func() {
		for _, name := range []string{"actual-shared-banks.bin", "actual-native-BEFORE.bin"} {
			path := filepath.Join(directory, name)
			if _, e := os.Stat(path); e == nil {
				if ce := rskGzipOwnedDuplicate4E23C0(path); ce != nil && err == nil {
					err = ce
				}
			}
		}
	}()
	if startErr != nil || forced.Load() || ctx.Err() != nil || outLog.Exceeded || errLog.Exceeded {
		return fmt.Errorf("bounded native child did not finish naturally: %v %v", startErr, waitErr)
	}
	var pre struct {
		PID       int    `json:"pid"`
		SHA       string `json:"executable_sha256"`
		InputSHA  string `json:"input_sha256"`
		BeforeSHA string `json:"BEFORE_sha256"`
	}
	p, e := os.ReadFile(filepath.Join(directory, "actual-child-PREEXEC.json"))
	if e != nil {
		return e
	}
	if e = json.Unmarshal(p, &pre); e != nil {
		return e
	}
	if pre.PID != pid || pre.SHA != sha || pre.InputSHA != rskHash4E23C0(item.Raw) {
		return fmt.Errorf("actual same-PID/root PREEXEC input/main binding")
	}
	var meta rskGraphMetadata4E23C0
	p, e = os.ReadFile(filepath.Join(directory, "graph.json"))
	if e != nil {
		return e
	}
	if e = json.Unmarshal(p, &meta); e != nil {
		return e
	}
	before, e := os.ReadFile(filepath.Join(directory, "actual-native-BEFORE.bin"))
	if e != nil || rskHash4E23C0(before) != pre.BeforeSHA {
		return fmt.Errorf("whole BEFORE binding")
	}
	raw, e := os.ReadFile(filepath.Join(directory, "actual-shared-banks.bin"))
	if e != nil {
		return e
	}
	if len(raw) != (len(meta.Banks)+1)*meta.Page {
		return fmt.Errorf("shared native bank extent")
	}
	header, e := rskReadHeader4E23C0(raw)
	if e != nil {
		return e
	}
	after := raw[meta.Page:]
	if header[0] != rawDefaultDamageSharedMagic4E0B30 || header[1] != rawDefaultDamageSharedVersion4E0B30 || header[36] != 0 || header[38] != 1 || header[40] != (map[string]uint64{"Go": 1, "C": 2}[backend]) {
		return fmt.Errorf("actual generic capture/root/validation header")
	}
	var state rskChildResult4E23C0
	p, e = os.ReadFile(filepath.Join(directory, "actual-input-state.json"))
	if e != nil {
		return e
	}
	if e = json.Unmarshal(p, &state); e != nil {
		return e
	}
	if state.PID != pid || state.Backend != backend {
		return fmt.Errorf("durable actual same-PID input state")
	}
	result := &state
	if exit == 0 {
		var v rskChildResult4E23C0
		p, e = os.ReadFile(filepath.Join(directory, "actual-return.json"))
		if e != nil {
			return e
		}
		if e = json.Unmarshal(p, &v); e != nil {
			return e
		}
		if v.PID != pid || v.Backend != backend || header[2] != 2 || !reflect.DeepEqual(v.UsedReads, state.UsedReads) || !reflect.DeepEqual(v.UsedServices, state.UsedServices) || !reflect.DeepEqual(v.Consumed, state.Consumed) {
			return fmt.Errorf("healthy actual return/header/state")
		}
		result = &v
	} else if exit == 172 {
		if header[2] != 5 || header[3] != 10 && header[3] != 11 || header[20] == 0 || header[20] != header[21] || header[22] != 0 || header[24] != 0 || header[10] != 3 || header[6] == 0 || header[8] == 0 || header[8] > 8 || header[5] < header[7] || header[5]-header[7] >= header[8] {
			return fmt.Errorf("actual native operand-fault capture scope")
		}
	} else {
		return fmt.Errorf("unexpected natural native exit %d", exit)
	}
	trace, e := rskReadTrace4E23C0(filepath.Join(directory, "actual-observations.jsonl"))
	if e != nil {
		return e
	}
	if exit == 172 {
		if len(trace) == 0 || trace[len(trace)-1].Kind != "fault" || uint64(trace[len(trace)-1].Site) != header[9] || trace[len(trace)-1].Address != header[7] || uint64(trace[len(trace)-1].Width) != header[8] {
			return fmt.Errorf("actual fault and unmatched operand prefix")
		}
	}
	for _, row := range item.References {
		if (row.Status == "returned") != (exit == 0) {
			return fmt.Errorf("observed original/native terminal domain differs %s", row.Mode)
		}
		if e = rskCompareReference4E23C0(&item.Input, row, &meta, trace, before, after, result); e != nil {
			return fmt.Errorf("%s: %w", row.Mode, e)
		}
	}
	return rskWriteJSON4E23C0(filepath.Join(directory, "actual-qualification.json"), map[string]any{"qualified_selected_observations": true, "actual_pid": pid, "backend": backend, "input": item.Input.Name, "original_CW_references": []uint16{0x0e3f, 0x023f}, "actual_native_calls": 1, "native_algorithm_floating_arithmetic": false, "native_before_sha256": rskHash4E23C0(before), "native_after_sha256": rskHash4E23C0(after), "signal_header": header, "whole_foreign_banks_and_canaries": true, "raw_return_and_external_arguments": true, "physical_nil_EA_equivalence": false, "independent_Skeleton_native_fault_PC_qualification": false, "full_helper_or_product_or_campaign_credit": false, "native_Go_heap_credit": false, "caller_stack_ABI_equivalence": false})
}

func TestRawSkeletonNativeOriginal4E23C0(t *testing.T) {
	if os.Getenv(rskChildEnv4E23C0) == "1" {
		return
	}
	cases, e := rskLoadCases4E23C0()
	if e != nil {
		t.Fatal(e)
	}
	initial := os.Getenv("OPENNOX_RAW_DEFAULT_DAMAGE_INITIAL_ARTIFACTS")
	if initial == "" {
		t.Fatal("prepared unique INITIAL_ARTIFACTS root required")
	}
	root, e := filepath.Abs(filepath.Join(initial, "raw-skeleton-4e23c0"))
	if e != nil {
		t.Fatal(e)
	}
	if e = os.MkdirAll(filepath.Dir(root), 0700); e != nil {
		t.Fatal(e)
	}
	if e = os.Mkdir(root, 0700); e != nil {
		t.Fatal("unique retained Skeleton directory", e)
	}
	executable, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	sha, e := rskArchiveMain4E23C0(executable, root)
	if e != nil {
		t.Fatal(e)
	}
	attempted, passed := 0, 0
	for i := range cases {
		for _, backend := range []string{"Go", "C"} {
			attempted++
			item := &cases[i]
			ok := t.Run(backend+"/"+item.Input.Name, func(t *testing.T) {
				directory := filepath.Join(root, fmt.Sprintf("%03d-%s", i, backend))
				if e := rskRunCase4E23C0(directory, executable, sha, backend, item); e != nil {
					t.Fatal(e)
				}
			})
			if ok {
				passed++
			}
		}
	}
	var smallBytes int64
	e = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && !strings.HasSuffix(path, "actual-main-before-first-child.zip") {
			smallBytes += info.Size()
		}
		return nil
	})
	if e != nil {
		t.Error(e)
	}
	summary := map[string]any{"scheduled_backend_cases": attempted, "qualified_native_calls": passed, "actual_native_call_budget": 502, "original_reference_comparisons_budget": 1004, "original_inputs": 251, "reference_SHA256": rskReferenceSHA4E23C0, "input_SHA256": rskInputSHA4E23C0, "one_canonical_main_SHA256": sha, "retained_small_artifact_bytes": smallBytes, "small_artifact_budget": rskLogsBudget4E23C0, "full_conformance_credit": false, "native_algorithm_floating_arithmetic": false, "fault_PC_and_physical_nil_EA_equivalence_credit": false}
	if e = rskWriteJSON4E23C0(filepath.Join(root, "actual-native-summary.json"), summary); e != nil {
		t.Error(e)
	}
	if smallBytes > rskLogsBudget4E23C0 {
		t.Logf("retention advisory: actual immutable evidence exceeds target 10MiB: %d; no evidence discarded", smallBytes)
	}
	if attempted != 502 || passed != 502 {
		t.Fatalf("actual Skeleton calls: attempted=%d passed=%d", attempted, passed)
	}
	t.Logf("RAW_SKELETON_NATIVE_ORIGINAL_OK actual-calls=%d original-CW-comparisons=%d inputs=251 full-helper-credit=false fault-PC-credit=false", passed, passed*2)
}
