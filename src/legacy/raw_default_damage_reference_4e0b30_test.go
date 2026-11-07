//go:build darwin && arm64

package legacy

import (
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"container/list"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const rddBundleSHA4E0B30 = "213a4752c6e991ecbb6037a1333d04dcaf933aeaefe29587c3f1fd85925e606b"
const rddCaller4E0B30 = uint64(0x2200e000)
const rddStop4E0B30 = uint64(0x23000100)

type rddMember4E0B30 struct {
	Bytes int64  `json:"bytes"`
	SHA   string `json:"sha256"`
}
type rddSnapshot4E0B30 struct {
	Bytes     int    `json:"bytes"`
	File      string `json:"file"`
	SHA       string `json:"sha256"`
	GzipBytes int    `json:"gzip_bytes"`
	GzipSHA   string `json:"gzip_sha256"`
}
type rddAreaRef4E0B30 struct {
	Name, Address string
	Permissions   int
	Size          int
	Snapshot      rddSnapshot4E0B30
}
type rddRun4E0B30 struct {
	Rows     string `json:"rows_member"`
	Inputs   string `json:"inputs_member"`
	Services string `json:"services_member"`
	Calls    int    `json:"actual_root_calls"`
	Cases    int    `json:"input_case_count"`
}
type rddCache4E0B30 struct {
	Key   string
	Bytes []byte
}
type rddBundle4E0B30 struct {
	Zip     *zip.ReadCloser
	Members map[string]*zip.File
	Index   struct {
		Schema  string                     `json:"schema"`
		Members map[string]rddMember4E0B30 `json:"members"`
		Runs    map[string]rddRun4E0B30    `json:"runs"`
	}
	Cached       map[string]*list.Element
	LRU          *list.List
	CacheBytes   int
	Verified     map[string]bool
	Instructions map[string]rddInstruction4E0B30
	Calls        map[string]rddCallSpec4E0B30
}

func rddOpenBundle4E0B30() (b *rddBundle4E0B30, err error) {
	path := "testdata/default_damage_parent_4e0b30/default-damage-parent-4e0b30-reference-v1.zip"
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	h := sha256.New()
	_, e = io.Copy(h, f)
	f.Close()
	if e != nil || fmt.Sprintf("%x", h.Sum(nil)) != rddBundleSHA4E0B30 {
		return nil, fmt.Errorf("immutable original bundle identity: %v", e)
	}
	z, e := zip.OpenReader(path)
	if e != nil {
		return nil, e
	}
	b = &rddBundle4E0B30{Zip: z, Members: map[string]*zip.File{}, Cached: map[string]*list.Element{}, LRU: list.New(), Verified: map[string]bool{}}
	defer func() {
		if err != nil {
			z.Close()
		}
	}()
	for _, file := range z.File {
		if b.Members[file.Name] != nil || strings.HasPrefix(file.Name, "/") || strings.Contains(file.Name, "../") {
			return b, fmt.Errorf("duplicate/unsafe ZIP member")
		}
		b.Members[file.Name] = file
	}
	index := b.Members["bundle-index-v1.json"]
	if index == nil {
		return b, fmt.Errorf("portable index absent")
	}
	r, e := index.Open()
	if e != nil {
		return b, e
	}
	e = json.NewDecoder(r).Decode(&b.Index)
	r.Close()
	if e != nil {
		return b, e
	}
	if b.Index.Schema != "opennox-default-damage-parent-portable-reference-bundle-v1" || len(b.Members) != len(b.Index.Members)+1 {
		return b, fmt.Errorf("portable member set")
	}
	for name := range b.Index.Members {
		if b.Members[name] == nil {
			return b, fmt.Errorf("indexed member absent")
		}
	}
	if e = b.loadContracts(); e != nil {
		return b, e
	}
	return b, nil
}
func (b *rddBundle4E0B30) close() { b.Zip.Close(); b.Cached = nil; b.LRU.Init() }
func (b *rddBundle4E0B30) member(name string) ([]byte, error) {
	m, ok := b.Index.Members[name]
	if !ok {
		return nil, fmt.Errorf("unindexed reference member %s", name)
	}
	f := b.Members[name]
	r, e := f.Open()
	if e != nil {
		return nil, e
	}
	data, e := io.ReadAll(r)
	r.Close()
	if e != nil {
		return nil, e
	}
	if int64(len(data)) != m.Bytes || ddiSHA4E0B30(data) != m.SHA {
		return nil, fmt.Errorf("reference member fullSHA %s", name)
	}
	return data, nil
}
func (b *rddBundle4E0B30) snapshot(ref rddSnapshot4E0B30) ([]byte, error) {
	if ref.File != "memory/"+ref.SHA+".bin.gz" {
		return nil, fmt.Errorf("snapshot content-address contract")
	}
	indexed, ok := b.Index.Members[ref.File]
	if !ok || indexed.Bytes != int64(ref.GzipBytes) || indexed.SHA != ref.GzipSHA {
		return nil, fmt.Errorf("every snapshot compressed identity")
	}
	if el := b.Cached[ref.SHA]; el != nil {
		cached := el.Value.(rddCache4E0B30).Bytes
		if len(cached) != ref.Bytes {
			return nil, fmt.Errorf("cached snapshot extent")
		}
		b.LRU.MoveToBack(el)
		return cached, nil
	}
	packed, e := b.member(ref.File)
	if e != nil {
		return nil, e
	}
	if len(packed) != ref.GzipBytes || ddiSHA4E0B30(packed) != ref.GzipSHA {
		return nil, fmt.Errorf("snapshot compressed identity")
	}
	z, e := gzip.NewReader(bytes.NewReader(packed))
	if e != nil {
		return nil, e
	}
	raw, e := io.ReadAll(z)
	ce := z.Close()
	if e != nil || ce != nil {
		return nil, fmt.Errorf("snapshot gzip %v/%v", e, ce)
	}
	if len(raw) != ref.Bytes || ddiSHA4E0B30(raw) != ref.SHA {
		return nil, fmt.Errorf("complete snapshot bytes/SHA")
	}
	for b.CacheBytes+len(raw) > 16<<20 && b.LRU.Len() > 0 {
		el := b.LRU.Front()
		old := el.Value.(rddCache4E0B30)
		delete(b.Cached, old.Key)
		b.CacheBytes -= len(old.Bytes)
		b.LRU.Remove(el)
	}
	if len(raw) <= 16<<20 {
		b.Cached[ref.SHA] = b.LRU.PushBack(rddCache4E0B30{ref.SHA, raw})
		b.CacheBytes += len(raw)
	}
	b.Verified[ref.SHA] = true
	return raw, nil
}
func (b *rddBundle4E0B30) rows(run rddRun4E0B30, fn func(json.RawMessage) error) error {
	// The compressed member hash is checked before streaming any certificates.
	packed, e := b.member(run.Rows)
	if e != nil {
		return e
	}
	z, e := gzip.NewReader(bytes.NewReader(packed))
	if e != nil {
		return e
	}
	defer z.Close()
	reader := bufio.NewReaderSize(z, 1<<20)
	count := 0
	for {
		line, e := reader.ReadBytes('\n')
		if e == io.EOF && len(line) == 0 {
			break
		}
		if e != nil {
			return e
		}
		if len(line) == 0 || line[len(line)-1] != '\n' {
			return fmt.Errorf("partial original row")
		}
		if e = fn(json.RawMessage(line)); e != nil {
			return e
		}
		count++
	}
	if count != run.Calls {
		return fmt.Errorf("original row count %d want%d", count, run.Calls)
	}
	return nil
}
func rddCanonical4E0B30(raw json.RawMessage) ([]byte, error) {
	var value any
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	if e := d.Decode(&value); e != nil {
		return nil, e
	}
	var out bytes.Buffer
	e := json.NewEncoder(&out)
	e.SetEscapeHTML(false)
	if err := e.Encode(value); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(out.Bytes(), []byte{'\n'}), nil
}
func rddSameJSON4E0B30(a, b json.RawMessage) bool {
	x, e := rddCanonical4E0B30(a)
	if e != nil {
		return false
	}
	y, e := rddCanonical4E0B30(b)
	return e == nil && bytes.Equal(x, y)
}
func rddNumber4E0B30(text string) (uint64, error) {
	return strconv.ParseUint(strings.TrimPrefix(text, "0x"), 16, 64)
}
func rddHex4E0B30(data string) ([]byte, error) { return hex.DecodeString(data) }

type rddEffect4E0B30 struct {
	Address *uint32 `json:"address"`
	Arg     *int    `json:"arg_index"`
	Role    string  `json:"role"`
	Offset  int     `json:"offset"`
	Hex     string  `json:"hex"`
}
type rddResponse4E0B30 struct {
	EAX     *uint32                       `json:"eax"`
	ECX     *uint32                       `json:"ecx"`
	EDX     *uint32                       `json:"edx"`
	Effects []rddEffect4E0B30             `json:"effects"`
	X87     *struct{ Format, Hex string } `json:"x87"`
}
type rddSchedule4E0B30 struct {
	Responses []json.RawMessage `json:"responses"`
	Repeat    bool              `json:"repeat_last"`
}
type rddStimulus4E0B30 struct {
	PC         string            `json:"pc"`
	Site       string            `json:"site"`
	Address    uint32            `json:"address"`
	Width      int               `json:"width"`
	Occurrence int               `json:"occurrence"`
	Effects    []rddEffect4E0B30 `json:"effects"`
}
type rddInput4E0B30 struct {
	ddiInput4E0B30
	ImageInitializers []struct {
		Address uint32 `json:"address"`
		Hex     string `json:"hex"`
	} `json:"image_initializers"`
	Services      map[string]rddSchedule4E0B30 `json:"services"`
	AfterReads    []rddStimulus4E0B30          `json:"after_reads"`
	AfterServices []rddStimulus4E0B30          `json:"after_services"`
}
type rddFPU4E0B30 struct {
	Control, Status, Tag uint32
	Raw                  [8][2]uint64 `json:"raw_physical80"`
}
type rddFrame4E0B30 struct {
	Kind, Name, Entry string
	CallSite          string `json:"call_site"`
	StackEntry        string `json:"stack_entry"`
	ReturnPC          string `json:"return_pc"`
	Args              []uint32
	Before            map[string]uint32 `json:"registers_before"`
	After             map[string]uint32 `json:"registers_after"`
	FPUBefore         rddFPU4E0B30      `json:"FPU_before"`
	FPUAfter          rddFPU4E0B30      `json:"FPU_after"`
	Completed         bool
	Preserved         bool            `json:"callee_saved_preserved"`
	ReturnedTo        string          `json:"returned_to"`
	StackExit         string          `json:"stack_exit"`
	Occurrence        int             `json:"service_occurrence"`
	X87               json.RawMessage `json:"x87_return_input"`
}
type rddEvent4E0B30 struct {
	Kind              string
	Sequence          int
	PC, Address       string
	Width, Size       int
	Name              string
	RangeName         string            `json:"range_name"`
	Depth             int               `json:"frame_depth"`
	RegistersAtWrite  map[string]uint32 `json:"registers_at_write"`
	Encoding, Source  string
	Completed, Fault  bool
	Before, After     string
	ValueHex          *string `json:"value_hex"`
	Value             uint64  `json:"hook_value_low64"`
	Site, Target      string
	Args              []uint32
	ArgumentWords     int    `json:"argument_words"`
	StackBefore       string `json:"stack_before"`
	Registers         map[string]uint32
	FPU               rddFPU4E0B30
	Frame             json.RawMessage
	Reason            string
	Occurrence        int
	ResponseIndex     int `json:"response_index"`
	Repeated          bool
	InputResponse     json.RawMessage `json:"input_response"`
	HelperExecuted    bool            `json:"original_helper_executed"`
	StimulusIndex     int             `json:"stimulus_index"`
	ReadSequence      int             `json:"read_sequence"`
	ServiceSite       string          `json:"service_site"`
	OriginalWrite     int             `json:"original_write_sequence"`
	WriteOffset       int             `json:"write_offset"`
	HookSlice         string          `json:"original_hook_value_slice"`
	OriginalCompleted bool            `json:"original_instruction_completed"`
	PhysicalOnly      bool            `json:"emulator_physical_observation_only"`
	ChangedSequences  []int           `json:"changed_fragment_sequences"`
	Observation       json.RawMessage `json:"faulting_write_observation"`
	MappedBefore      []struct {
		Address string
		Width   int
		Offset  int    `json:"write_offset"`
		Before  string `json:"before_hex"`
	} `json:"mapped_spans_before"`
	Access     int
	FaultValue uint64 `json:"value_low64"`
}
type rddRow4E0B30 struct {
	Name, Mode, Status string
	Control            uint32 `json:"control_word"`
	Input              json.RawMessage
	InputSHA           string             `json:"input_sha256"`
	Before, After      []rddAreaRef4E0B30 `json:"-"`
	MemoryBefore       []rddAreaRef4E0B30 `json:"complete_memory_before"`
	MemoryAfter        []rddAreaRef4E0B30 `json:"complete_memory_after"`
	Events             []json.RawMessage  `json:"complete_events"`
	RegistersBefore    map[string]uint32  `json:"registers_before"`
	RegistersAfter     map[string]uint32  `json:"registers_after"`
	FPUAfter           rddFPU4E0B30       `json:"FPU_after"`
	CallerBefore       string             `json:"caller_before_hex"`
	CallerAfter        string             `json:"caller_after_hex"`
	Prefix             int                `json:"input_initialization_events_before_snapshot"`
	Instructions       int
	Entered            map[string]int    `json:"instruction_coverage"`
	Finished           map[string]int    `json:"completed_original_instruction_coverage"`
	Entries            map[string]int    `json:"actual_original_entry_counts"`
	Services           map[string]int    `json:"service_call_counts"`
	Returned           []json.RawMessage `json:"completed_call_frames"`
	Open               []json.RawMessage `json:"open_call_frames"`
	PendingReturn      json.RawMessage   `json:"pending_return_at_terminal"`
	PendingCall        json.RawMessage   `json:"pending_call_at_terminal"`
	Faults             []json.RawMessage
	UsedReads          []int `json:"used_after_read_stimuli"`
	UnusedReads        []int `json:"unused_after_read_stimuli"`
	UsedServices       []int `json:"used_after_service_stimuli"`
	UnusedServices     []int `json:"unused_after_service_stimuli"`
	PartialCount       int   `json:"observed_partial_write_fragments"`
	FaultWriteCount    int   `json:"faulting_write_observations"`
	Reconstruction     []struct {
		Name          string
		Bytes         int
		Actual        string `json:"actual_sha256"`
		Reconstructed string `json:"reconstructed_sha256"`
		Complete      bool   `json:"reconstructed_from_completed_writes_and_observed_fault_fragments"`
	} `json:"all_memory_write_reconstruction"`
	NoNative       bool            `json:"no_native_algorithm"`
	NoHostFloat    bool            `json:"no_host_float_computation"`
	CodeUnchanged  bool            `json:"original_executable_bytes_unchanged"`
	ExternalCredit bool            `json:"original_world_or_HP_helper_execution_credit"`
	HarnessError   json.RawMessage `json:"harness_error"`
	EmulationError json.RawMessage `json:"emulation_error"`
	Cleanup        struct {
		Registered int `json:"registered_hooks"`
		Removed    int `json:"removed_hooks"`
		Errors     []json.RawMessage
		Released   bool `json:"native_engine_reference_released"`
	} `json:"recorder_cleanup"`
}
type rddArea4E0B30 struct {
	Ref     rddAreaRef4E0B30
	Address uint64
	Bytes   []byte
	Dirty   bool
}
type rddMemory4E0B30 struct{ Areas []rddArea4E0B30 }

func (m *rddMemory4E0B30) area(address uint64, width int) *rddArea4E0B30 {
	for i := range m.Areas {
		a := &m.Areas[i]
		if a.Address <= address && address+uint64(width) <= a.Address+uint64(len(a.Bytes)) {
			return a
		}
	}
	return nil
}
func (m *rddMemory4E0B30) read(address uint64, width int) ([]byte, error) {
	a := m.area(address, width)
	if a == nil {
		return nil, fmt.Errorf("unmapped reference read %#x/%d", address, width)
	}
	return a.Bytes[address-a.Address : address-a.Address+uint64(width)], nil
}
func (m *rddMemory4E0B30) write(address uint64, data []byte) error {
	a := m.area(address, len(data))
	if a == nil || a.Ref.Permissions&2 == 0 {
		return fmt.Errorf("unmapped/unwritable certificate write")
	}
	if !a.Dirty {
		a.Bytes = append([]byte(nil), a.Bytes...)
		a.Dirty = true
	}
	copy(a.Bytes[address-a.Address:], data)
	return nil
}
func (m *rddMemory4E0B30) word(address uint64) (uint32, error) {
	v, e := m.read(address, 4)
	if e != nil {
		return 0, e
	}
	return binary.LittleEndian.Uint32(v), nil
}

type rddHost4E0B30 struct {
	Reason       string
	Address      uint64
	Bytes        []byte
	Stimulus     int
	ReadSequence int
	Service      string
}
type rddCertificate4E0B30 struct {
	Calls, Reads, Writes, Faults, Fragments, FaultingWrites, Services, UsedReads, UsedServices int
	Returned                                                                                   bool
	Return                                                                                     uint32
}

// Copy-on-first-write reconstruction keeps every byte of every area while
// avoiding rehashing/copying unchanged5.7MiB image data for every row.
func rddValidateRow4E0B30(bundle *rddBundle4E0B30, row *rddRow4E0B30) (summary rddCertificate4E0B30, err error) {
	bad := func(format string, args ...any) error {
		return fmt.Errorf("%s/%s: %s", row.Name, row.Mode, fmt.Sprintf(format, args...))
	}
	if row.Control != 0x0e3f && row.Control != 0x023f {
		return summary, bad("CW")
	}
	if (row.Control == 0x0e3f && row.Mode != "startup-53-chop") || (row.Control == 0x023f && row.Mode != "nearest-53") {
		return summary, bad("CW label")
	}
	if !row.NoNative || !row.NoHostFloat || !row.CodeUnchanged || row.ExternalCredit || string(row.HarnessError) != "null" {
		return summary, bad("reference scope")
	}
	inputCanonical, e := rddCanonical4E0B30(row.Input)
	if e != nil || ddiSHA4E0B30(inputCanonical) != row.InputSHA {
		return summary, bad("input canonicalSHA")
	}
	var in rddInput4E0B30
	if e = json.Unmarshal(row.Input, &in); e != nil {
		return summary, e
	}
	if in.Name != row.Name {
		return summary, bad("input name")
	}
	mem := rddMemory4E0B30{}
	for _, ref := range row.MemoryBefore {
		address, e := rddNumber4E0B30(ref.Address)
		if e != nil {
			return summary, e
		}
		data, e := bundle.snapshot(ref.Snapshot)
		if e != nil {
			return summary, e
		}
		if len(data) != ref.Size {
			return summary, bad("initial full area size")
		}
		for _, old := range mem.Areas {
			if address < old.Address+uint64(len(old.Bytes)) && old.Address < address+uint64(len(data)) {
				return summary, bad("area overlap")
			}
		}
		mem.Areas = append(mem.Areas, rddArea4E0B30{Ref: ref, Address: address, Bytes: data})
	}
	caller := make([]byte, 24)
	binary.LittleEndian.PutUint32(caller, uint32(rddStop4E0B30))
	for i, v := range in.Args {
		binary.LittleEndian.PutUint32(caller[4+4*i:], v)
	}
	actual, e := mem.read(rddCaller4E0B30, 24)
	if e != nil || !bytes.Equal(caller, actual) || hex.EncodeToString(caller) != row.CallerBefore {
		return summary, bad("complete caller before")
	}
	if row.RegistersBefore["eip"] != 0x4e0b30 || row.RegistersBefore["esp"] != uint32(rddCaller4E0B30) || row.RegistersBefore["eax"] != 0 || row.RegistersBefore["ecx"] != 0 || row.RegistersBefore["edx"] != 0 || row.RegistersBefore["eflags"] != 0x202 {
		return summary, bad("entry registers")
	}
	for reg, want := range map[string]uint32{"ebx": 0x12345678, "esi": 0x23456789, "edi": 0x3456789a, "ebp": 0x456789ab} {
		if row.RegistersBefore[reg] != want {
			return summary, bad("entry saved register")
		}
	}
	frames := []rddFrame4E0B30{{Kind: "actual-original", Name: "parent", Entry: "0x004e0b30", StackEntry: "0x2200e000", ReturnPC: "0x23000100", Args: append([]uint32(nil), in.Args[:]...), Before: row.RegistersBefore, FPUBefore: rddFPU4E0B30{Control: row.Control, Tag: 0xffff}}}
	var pendingReturn *rddFrame4E0B30
	var issued *rddEvent4E0B30
	var issuedReturn uint32
	var current *rddEvent4E0B30
	entered, completed, entries, services := map[string]int{}, map[string]int{}, map[string]int{}, map[string]int{}
	readOccurrence := map[string]int{}
	readUsed, serviceUsed := map[int]bool{}, map[int]bool{}
	writes := map[int]rddEvent4E0B30{}
	fragments := map[int]rddEvent4E0B30{}
	var observedReturns, observedFaults []json.RawMessage
	var expectedHosts []rddHost4E0B30
	var activeResponses = map[string]json.RawMessage{}
	observedWrites := map[int]bool{}
	externalEntries, trampoline := 0, 0
	resolve := func(effect rddEffect4E0B30, args []uint32) (uint64, []byte, error) {
		address := uint64(0)
		selectors := 0
		if effect.Address != nil {
			address = uint64(*effect.Address)
			selectors++
		}
		if effect.Arg != nil {
			if *effect.Arg < 0 || *effect.Arg >= len(args) {
				return 0, nil, bad("effect arg")
			}
			address = uint64(args[*effect.Arg])
			selectors++
		}
		if effect.Role != "" {
			v, ok := in.Roles[effect.Role]
			if !ok {
				return 0, nil, bad("effect role")
			}
			address = uint64(v)
			selectors++
		}
		if selectors != 1 {
			return 0, nil, bad("effect selector")
		}
		n := int64(address) + int64(effect.Offset)
		if n < 0 || uint64(n) > 0xffffffff {
			return 0, nil, bad("effect address")
		}
		v, e := hex.DecodeString(effect.Hex)
		return uint64(n), v, e
	}
	queue := func(reason string, effects []rddEffect4E0B30, args []uint32, stim, read int, site string) error {
		for _, effect := range effects {
			address, data, e := resolve(effect, args)
			if e != nil {
				return e
			}
			expectedHosts = append(expectedHosts, rddHost4E0B30{reason, address, data, stim, read, site})
		}
		return nil
	}
	prefix, e := rddInitialPrefix4E0B30(in, caller, row.Control)
	if e != nil {
		return summary, e
	}
	if row.Prefix != len(prefix) {
		return summary, bad("declared initial prefix count")
	}
	for i, raw := range row.Events {
		var event rddEvent4E0B30
		if e = json.Unmarshal(raw, &event); e != nil {
			return summary, e
		}
		if event.Sequence != i {
			return summary, bad("event sequence")
		}
		if i < row.Prefix {
			if event.Kind != "host-write" {
				return summary, bad("initial prefix kind")
			}
			want := prefix[i]
			if event.Reason != want.Reason || event.Address != fmt.Sprintf("0x%08x", want.Address) || event.After != hex.EncodeToString(want.Bytes) || event.Width != len(want.Bytes) {
				return summary, bad("independent initial prefix declaration")
			}
			address, e := rddNumber4E0B30(event.Address)
			if e != nil {
				return summary, e
			}
			data, e := hex.DecodeString(event.After)
			if e != nil {
				return summary, e
			}
			v, e := mem.read(address, len(data))
			if e != nil || !bytes.Equal(v, data) {
				return summary, bad("input initialization snapshot")
			}
			continue
		}
		if len(expectedHosts) > 0 && (event.Kind == "original-instruction" || event.Kind == "external-x87-trampoline-instruction" || event.Kind == "call-entered" || event.Kind == "return-completed" || event.Kind == "fault") {
			return summary, bad("deferred/missing declared host effect")
		}
		switch event.Kind {
		case "host-write":
			if len(expectedHosts) == 0 {
				return summary, bad("undeclared host write %s", event.Reason)
			}
			want := expectedHosts[0]
			expectedHosts = expectedHosts[1:]
			address, e := rddNumber4E0B30(event.Address)
			if e != nil {
				return summary, e
			}
			before, e := hex.DecodeString(event.Before)
			if e != nil {
				return summary, e
			}
			after, e := hex.DecodeString(event.After)
			if e != nil {
				return summary, e
			}
			v, e := mem.read(address, event.Width)
			if e != nil || address != want.Address || event.Reason != want.Reason || !bytes.Equal(after, want.Bytes) || !bytes.Equal(v, before) || event.Width != len(after) || want.Stimulus >= 0 && event.StimulusIndex != want.Stimulus || want.ReadSequence >= 0 && event.ReadSequence != want.ReadSequence || want.Service != "" && event.ServiceSite != want.Service {
				return summary, bad("exact input/service effect timing/value")
			}
			if e = mem.write(address, after); e != nil {
				return summary, e
			}
		case "original-instruction", "external-x87-trampoline-instruction":
			current = &event
			address, e := rddNumber4E0B30(event.PC)
			if e != nil {
				return summary, e
			}
			encoded, e := hex.DecodeString(event.Encoding)
			if e != nil {
				return summary, e
			}
			v, e := mem.read(address, event.Size)
			if e != nil || len(encoded) != event.Size || !bytes.Equal(v, encoded) {
				return summary, bad("instruction bytes")
			}
			if event.Kind == "original-instruction" {
				definition, ok := bundle.Instructions[event.PC]
				if !ok || event.Size != definition.Size || event.Encoding != definition.Encoding || event.RangeName != definition.Range || event.Depth != len(frames) {
					return summary, bad("sealed instruction boundary/encoding/frame depth")
				}
				entered[event.PC]++
				if event.Completed {
					completed[event.PC]++
				}
				if address == 0x4e0b30 || address == 0x4e1400 || address == 0x4e1470 || address == 0x419a70 {
					entries[event.PC]++
				}
			} else {
				if (address != 0x23000500 && address != 0x23000506) || len(frames) == 0 || frames[len(frames)-1].Kind != "external-input-boundary" {
					return summary, bad("explicit x87 trampoline-only scope")
				}
				trampoline++
			}
			if len(encoded) == 1 && encoded[0] == 0xc3 {
				if pendingReturn != nil || len(frames) == 0 {
					return summary, bad("RET frame")
				}
				f := frames[len(frames)-1]
				frames = frames[:len(frames)-1]
				pendingReturn = &f
			}
		case "read":
			address, e := rddNumber4E0B30(event.Address)
			if e != nil {
				return summary, e
			}
			if current == nil || event.PC != current.PC || event.Width <= 0 || event.Source != rddInstructionSource4E0B30(*current) || event.Completed && !current.Completed || event.Fault && event.Completed {
				return summary, bad("ordered read instruction")
			}
			if event.ValueHex != nil {
				expected, e := hex.DecodeString(*event.ValueHex)
				if e != nil {
					return summary, e
				}
				v, e := mem.read(address, event.Width)
				if e != nil || len(expected) != event.Width || !bytes.Equal(v, expected) {
					return summary, bad("actual original read bytes")
				}
			} else if event.Completed {
				return summary, bad("completed read without bytes")
			}
			key := fmt.Sprintf("%s/%s/%d", event.PC, event.Address, event.Width)
			readOccurrence[key]++
			if readOccurrence[key] != event.Occurrence {
				return summary, bad("read occurrence")
			}
			summary.Reads++
			if event.Completed && event.Source == "actual-original" {
				for n, s := range in.AfterReads {
					pc, e := rddNumber4E0B30(s.PC)
					if e != nil {
						return summary, e
					}
					epc, _ := rddNumber4E0B30(event.PC)
					if pc == epc && uint64(s.Address) == address && s.Width == event.Width && s.Occurrence == event.Occurrence {
						if readUsed[n] {
							return summary, bad("duplicate read stimulus")
						}
						readUsed[n] = true
						if e = queue("after-completed-original-read", s.Effects, nil, n, event.Sequence, ""); e != nil {
							return summary, e
						}
					}
				}
			}
		case "write":
			if current == nil || event.PC != current.PC || event.Width <= 0 || event.Source != "actual-original" || event.Completed && !current.Completed || event.Fault && event.Completed || event.RegistersAtWrite["eip"] != uint32(mustRDDNumber4E0B30(event.PC)) {
				return summary, bad("ordered write instruction")
			}
			address, e := rddNumber4E0B30(event.Address)
			if e != nil {
				return summary, e
			}
			writes[i] = event
			summary.Writes++
			if event.Completed {
				before, e := hex.DecodeString(event.Before)
				if e != nil {
					return summary, e
				}
				after, e := hex.DecodeString(event.After)
				if e != nil {
					return summary, e
				}
				v, e := mem.read(address, event.Width)
				if e != nil || len(before) != event.Width || len(after) != event.Width || !bytes.Equal(v, before) {
					return summary, bad("completed write before/width")
				}
				if event.Width <= 8 && !bytes.Equal(after, ddiBytes4E0B30(event.Value, event.Width)) {
					return summary, bad("write operand bits")
				}
				if e = mem.write(address, after); e != nil {
					return summary, e
				}
			} else {
				for _, span := range event.MappedBefore {
					at, e := rddNumber4E0B30(span.Address)
					if e != nil {
						return summary, e
					}
					v, e := mem.read(at, span.Width)
					data, de := hex.DecodeString(span.Before)
					if e != nil || de != nil || !bytes.Equal(v, data) {
						return summary, bad("incomplete write mapped prefix")
					}
				}
			}
		case "call-issued":
			if issued != nil || current == nil || event.Site != current.PC {
				return summary, bad("ordered CALL issuance")
			}
			spec, ok := bundle.Calls[event.Site]
			if !ok || event.ArgumentWords != spec.Argc || !spec.Indirect && event.Target != spec.Target {
				return summary, bad("actual sealed CALL contract")
			}
			issuedReturn = uint32(mustRDDNumber4E0B30(event.Site)) + uint32(current.Size)
			sp, e := rddNumber4E0B30(event.StackBefore)
			if e != nil {
				return summary, e
			}
			if len(event.Args) != event.ArgumentWords || event.Registers["esp"] != uint32(sp) {
				return summary, bad("actual CALL argument words")
			}
			for n, want := range event.Args {
				v, e := mem.word(sp + uint64(4*n))
				if e != nil || v != want {
					return summary, bad("outgoing stack argument")
				}
			}
			copy := event
			issued = &copy
		case "call-entered":
			var f rddFrame4E0B30
			if e = json.Unmarshal(event.Frame, &f); e != nil {
				return summary, e
			}
			if issued == nil || f.CallSite != issued.Site || f.Entry != issued.Target || !reflect.DeepEqual(f.Args, issued.Args) {
				return summary, bad("actual CALL entered identity")
			}
			sp, e := rddNumber4E0B30(f.StackEntry)
			if e != nil {
				return summary, e
			}
			before, _ := rddNumber4E0B30(issued.StackBefore)
			ret, _ := rddNumber4E0B30(f.ReturnPC)
			value, e := mem.word(sp)
			if e != nil || sp+4 != before || uint64(value) != ret || f.Before["esp"] != uint32(sp) || f.FPUBefore.Control != row.Control || ret != uint64(issuedReturn) || f.Before["eip"] != uint32(mustRDDNumber4E0B30(f.Entry)) {
				return summary, bad("CALL entry frame/return word")
			}
			spec := bundle.Calls[f.CallSite]
			if f.Name != spec.Name || f.Kind != spec.Kind || len(f.Args) != spec.Argc || f.Completed {
				return summary, bad("entered actual/external callee contract")
			}
			if f.Kind == "external-input-boundary" {
				externalEntries++
			}
			frames = append(frames, f)
			issued = nil
			summary.Calls++
		case "external-service-input":
			if len(frames) == 0 {
				return summary, bad("service without frame")
			}
			f := &frames[len(frames)-1]
			if f.Kind != "external-input-boundary" || f.CallSite != event.Site || !reflect.DeepEqual(f.Args, event.Args) || event.HelperExecuted || event.Name != f.Name {
				return summary, bad("external boundary scope/args")
			}
			services[event.Site]++
			if event.Occurrence != services[event.Site] {
				return summary, bad("service occurrence")
			}
			site, _ := rddNumber4E0B30(event.Site)
			schedule, ok := in.Services[fmt.Sprintf("%08x", site)]
			if !ok || len(schedule.Responses) == 0 {
				return summary, bad("undeclared service")
			}
			selected := min(event.Occurrence-1, len(schedule.Responses)-1)
			repeated := event.Occurrence > len(schedule.Responses)
			if repeated && !schedule.Repeat || event.Repeated != repeated || event.ResponseIndex != selected || !rddSameJSON4E0B30(event.InputResponse, schedule.Responses[selected]) {
				return summary, bad("input response selection")
			}
			activeResponses[event.Site] = event.InputResponse
			f.Occurrence = event.Occurrence
			var response rddResponse4E0B30
			if e = json.Unmarshal(event.InputResponse, &response); e != nil {
				return summary, e
			}
			if e = queue("external-service-input-effect", response.Effects, f.Args, -1, -1, event.Site); e != nil {
				return summary, e
			}
			spec := bundle.Calls[event.Site]
			if (response.X87 != nil) != (spec.Return == "x87") {
				return summary, bad("explicit x87 service response contract")
			}
			if response.X87 != nil {
				var originalResponse map[string]json.RawMessage
				if e = json.Unmarshal(event.InputResponse, &originalResponse); e != nil {
					return summary, e
				}
				f.X87 = originalResponse["x87"]
				payload, e := hex.DecodeString(response.X87.Hex)
				if e != nil {
					return summary, e
				}
				opcode := []byte{}
				switch response.X87.Format {
				case "f32":
					opcode = []byte{0xd9, 0x05}
				case "f64":
					opcode = []byte{0xdd, 0x05}
				case "ext80":
					opcode = []byte{0xdb, 0x2d}
				default:
					return summary, bad("FPU payload format")
				}
				if len(payload) != map[string]int{"f32": 4, "f64": 8, "ext80": 10}[response.X87.Format] {
					return summary, bad("exact raw FPU input width")
				}
				opcode = append(opcode, 0x00, 0x04, 0x00, 0x23, 0xc3)
				expectedHosts = append(expectedHosts, rddHost4E0B30{"framework-exact-x87-service-payload", 0x23000400, payload, -1, -1, event.Site}, rddHost4E0B30{"framework-exact-x87-fld-ret", 0x23000500, opcode, -1, -1, event.Site})
			}
			summary.Services++
		case "return-completed":
			var f rddFrame4E0B30
			if e = json.Unmarshal(event.Frame, &f); e != nil {
				return summary, e
			}
			var before rddFrame4E0B30
			if pendingReturn != nil {
				before = *pendingReturn
				pendingReturn = nil
			} else if len(frames) > 0 && frames[len(frames)-1].Kind == "external-input-boundary" {
				before = frames[len(frames)-1]
				frames = frames[:len(frames)-1]
			} else {
				return summary, bad("return without actual frame")
			}
			if f.Kind != before.Kind || f.Name != before.Name || f.Entry != before.Entry || f.CallSite != before.CallSite || f.ReturnPC != before.ReturnPC || f.StackEntry != before.StackEntry || !reflect.DeepEqual(f.Args, before.Args) || !reflect.DeepEqual(f.Before, before.Before) || f.FPUBefore != before.FPUBefore || f.Occurrence != before.Occurrence || !rddOptionalSameJSON4E0B30(f.X87, before.X87) || !f.Completed || !f.Preserved {
				return summary, bad("full returned frame")
			}
			sp, _ := rddNumber4E0B30(f.StackEntry)
			exit, _ := rddNumber4E0B30(f.StackExit)
			ret, _ := rddNumber4E0B30(f.ReturnPC)
			if uint64(f.After["esp"]) != sp+4 || exit != sp+4 || uint64(f.After["eip"]) != ret || f.ReturnedTo != f.ReturnPC || f.FPUAfter.Control != row.Control {
				return summary, bad("cdecl return cleanup")
			}
			for _, reg := range []string{"ebx", "esi", "edi", "ebp"} {
				if f.After[reg] != f.Before[reg] {
					return summary, bad("callee-saved register")
				}
			}
			if f.Kind == "external-input-boundary" {
				var response rddResponse4E0B30
				if e = json.Unmarshal(activeResponses[f.CallSite], &response); e != nil {
					return summary, e
				}
				for name, value := range map[string]*uint32{"eax": response.EAX, "ecx": response.ECX, "edx": response.EDX} {
					want := f.Before[name]
					if value != nil {
						want = *value
					}
					if f.After[name] != want {
						return summary, bad("external declared register response")
					}
				}
				if f.After["eflags"] != f.Before["eflags"] {
					return summary, bad("external EFLAGS")
				}
				for n, s := range in.AfterServices {
					site, _ := rddNumber4E0B30(s.Site)
					actualSite, _ := rddNumber4E0B30(f.CallSite)
					if site == actualSite && s.Occurrence == f.Occurrence {
						if serviceUsed[n] {
							return summary, bad("duplicate service stimulus")
						}
						serviceUsed[n] = true
						if e = queue("after-completed-external-service", s.Effects, f.Args, n, -1, f.CallSite); e != nil {
							return summary, e
						}
					}
				}
			}
			observedReturns = append(observedReturns, event.Frame)
		case "fault":
			if e = rddValidateFault4E0B30(event, current, issued); e != nil {
				return summary, bad("%v", e)
			}
			observedFaults = append(observedFaults, raw)
			summary.Faults++
		case "observed-partial-write-fragment":
			write, ok := writes[event.OriginalWrite]
			if !ok || write.Completed || !write.Fault || event.OriginalCompleted || !event.PhysicalOnly || event.PC != write.PC {
				return summary, bad("fault fragment provenance")
			}
			start, _ := rddNumber4E0B30(write.Address)
			address, _ := rddNumber4E0B30(event.Address)
			if event.WriteOffset < 0 || event.WriteOffset+event.Width > write.Width || address != start+uint64(event.WriteOffset) || write.Width > 8 {
				return summary, bad("fragment extent")
			}
			before, e := hex.DecodeString(event.Before)
			if e != nil {
				return summary, e
			}
			after, e := hex.DecodeString(event.After)
			if e != nil {
				return summary, e
			}
			desired := ddiBytes4E0B30(write.Value, write.Width)[event.WriteOffset : event.WriteOffset+event.Width]
			v, e := mem.read(address, event.Width)
			if e != nil || !bytes.Equal(v, before) || !bytes.Equal(after, desired) || event.HookSlice != hex.EncodeToString(desired) {
				return summary, bad("unchanged original observed fragment bytes")
			}
			if e = mem.write(address, after); e != nil {
				return summary, e
			}
			fragments[i] = event
			summary.Fragments++
		case "faulting-write-observed":
			write, ok := writes[event.OriginalWrite]
			if !ok || write.Completed || !write.Fault || event.OriginalCompleted || !event.PhysicalOnly || event.PC != write.PC || event.Address != write.Address || event.Width != write.Width {
				return summary, bad("faulting write observation")
			}
			for _, n := range event.ChangedSequences {
				if f, ok := fragments[n]; !ok || f.OriginalWrite != event.OriginalWrite {
					return summary, bad("changed fragment identity")
				}
			}
			if observedWrites[event.OriginalWrite] {
				return summary, bad("duplicate faulting write observation")
			}
			observedWrites[event.OriginalWrite] = true
			if e = rddValidateFaultWrite4E0B30(write, event, fragments, &mem); e != nil {
				return summary, bad("%v", e)
			}
			summary.FaultingWrites++
		default:
			return summary, bad("unknown complete event kind %s", event.Kind)
		}
	}
	if len(expectedHosts) != 0 {
		return summary, bad("unconsumed expected host effect")
	}
	if !reflect.DeepEqual(entered, row.Entered) || !reflect.DeepEqual(completed, row.Finished) || !reflect.DeepEqual(entries, row.Entries) || !reflect.DeepEqual(services, row.Services) || entries["0x004e0b30"] != 1 {
		return summary, bad("exact original coverage/service counters")
	}
	if row.Instructions != sumRDDCounts4E0B30(entered)+trampoline+externalEntries {
		return summary, bad("instruction/external entry accounting")
	}
	if len(observedReturns) != len(row.Returned) || len(observedFaults) != len(row.Faults) {
		return summary, bad("full frame/fault counts")
	}
	for i, v := range observedReturns {
		if !rddSameJSON4E0B30(v, row.Returned[i]) {
			return summary, bad("complete call frame certificate")
		}
	}
	for i, v := range observedFaults {
		if !rddSameJSON4E0B30(v, row.Faults[i]) {
			return summary, bad("fault prefix certificate")
		}
	}
	if summary.Fragments != row.PartialCount || summary.FaultingWrites != row.FaultWriteCount {
		return summary, bad("fragment/fault-write counts")
	}
	for sequence, write := range writes {
		if !write.Completed && write.Fault && !observedWrites[sequence] {
			return summary, bad("faulting write certificate absent")
		}
	}
	if e = rddTerminalFrames4E0B30(row, frames, pendingReturn, issued, issuedReturn); e != nil {
		return summary, bad("%v", e)
	}
	if len(row.MemoryAfter) != len(mem.Areas) || len(row.Reconstruction) != len(mem.Areas) {
		return summary, bad("whole mapped area extent")
	}
	for i, a := range mem.Areas {
		ref := row.MemoryAfter[i]
		if ref.Name != a.Ref.Name || ref.Address != a.Ref.Address || ref.Size != a.Ref.Size || ref.Permissions != a.Ref.Permissions {
			return summary, bad("whole area identity")
		}
		after, e := bundle.snapshot(ref.Snapshot)
		if e != nil {
			return summary, e
		}
		if !bytes.Equal(a.Bytes, after) {
			return summary, bad("complete reconstructed memory bytes %s", ref.Name)
		}
		hash := ref.Snapshot.SHA
		claim := row.Reconstruction[i]
		if a.Dirty {
			hash = ddiSHA4E0B30(a.Bytes)
		}
		if hash != ref.Snapshot.SHA || claim.Name != ref.Name || claim.Bytes != len(a.Bytes) || claim.Actual != hash || claim.Reconstructed != hash || !claim.Complete {
			return summary, bad("whole memory SHA reconstruction")
		}
	}
	callerAfter, e := mem.read(rddCaller4E0B30, 24)
	if e != nil || hex.EncodeToString(callerAfter) != row.CallerAfter {
		return summary, bad("complete caller after incl scratch writes")
	}
	for _, rangeInfo := range []struct {
		Address uint64
		Bytes   int
		SHA     string
	}{{0x4e0b30, 1783, "4cf4af9b19b58d3912660df35a7b47010c363e5f595aa4ac1e74c3adae541fe1"}, {0x4e1400, 112, "0ac2ef45c5c13d020b99663c9fcac15321cdf3b825cabb3b44c746aae5a4fb73"}, {0x4e1470, 34, "fa632520ae8b970b679dd6c5b8bd947f81f4b41d65295f6d3a26edbcb975b92e"}, {0x419a70, 17, "10e536e2a987ea0fe14c91f007b2d5361cd87d4f0ca4cf5385fc82e53913d186"}} {
		data, e := mem.read(rangeInfo.Address, rangeInfo.Bytes)
		if e != nil || ddiSHA4E0B30(data) != rangeInfo.SHA {
			return summary, bad("unchanged sealed original code")
		}
	}
	if row.FPUAfter.Control != row.Control || row.Cleanup.Registered != 4 || row.Cleanup.Removed != 4 || len(row.Cleanup.Errors) != 0 || !row.Cleanup.Released {
		return summary, bad("FPU/recorder lifecycle certificate")
	}
	if row.Status == "returned" {
		if len(observedFaults) != 0 || len(frames) != 0 || pendingReturn != nil || issued != nil || row.RegistersAfter["eip"] != uint32(rddStop4E0B30) || row.RegistersAfter["esp"] != uint32(rddCaller4E0B30+4) || string(row.EmulationError) != "null" {
			return summary, bad("root returned terminal")
		}
		summary.Returned = true
		summary.Return = row.RegistersAfter["eax"]
	} else if row.Status == "original-fault" {
		if len(observedFaults) == 0 {
			return summary, bad("original fault absent")
		}
		var last rddEvent4E0B30
		_ = json.Unmarshal(observedFaults[len(observedFaults)-1], &last)
		if !reflect.DeepEqual(row.RegistersAfter, last.Registers) || row.FPUAfter != last.FPU {
			return summary, bad("terminal fault registers/FPU")
		}
	} else {
		return summary, bad("status")
	}
	for _, item := range []struct {
		Used, Unused []int
		Actual       map[int]bool
		Total        int
	}{{row.UsedReads, row.UnusedReads, readUsed, len(in.AfterReads)}, {row.UsedServices, row.UnusedServices, serviceUsed, len(in.AfterServices)}} {
		if len(item.Used) != len(item.Actual) || len(item.Used)+len(item.Unused) != item.Total {
			return summary, bad("stimulus used/unconsumed counts")
		}
		seen := map[int]bool{}
		previous := -1
		for _, n := range item.Used {
			if !item.Actual[n] || n < 0 || n >= item.Total || seen[n] || n <= previous {
				return summary, bad("actual used stimulus")
			}
			seen[n] = true
			previous = n
		}
		previous = -1
		for _, n := range item.Unused {
			if item.Actual[n] || n < 0 || n >= item.Total || seen[n] || n <= previous {
				return summary, bad("exact unused stimulus complement")
			}
			seen[n] = true
			previous = n
		}
	}
	summary.UsedReads = len(readUsed)
	summary.UsedServices = len(serviceUsed)
	return summary, nil
}
func sumRDDCounts4E0B30(values map[string]int) int {
	n := 0
	for _, v := range values {
		n += v
	}
	return n
}

func rddDecodeGzip4E0B30(packed []byte, out any) error {
	z, e := gzip.NewReader(bytes.NewReader(packed))
	if e != nil {
		return e
	}
	defer z.Close()
	d := json.NewDecoder(z)
	if e = d.Decode(out); e != nil {
		return e
	}
	var extra any
	if e = d.Decode(&extra); e != io.EOF {
		return fmt.Errorf("gzip JSON exact EOF: %v", e)
	}
	return nil
}
