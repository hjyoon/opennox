package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"testing"

	"github.com/opennox/opennox/v1/common/memmap"
)

const fourTimersInputSHA40A1F0 = "5a1e175d954288b72f677d4047a77d261eac9a10bcb52d21fcceaf21aae3a78e"
const fourTimersReferenceSHA40A1F0 = "f8065153a575ba5df6c4b43db3738830a5e948676d45fc3a4d708744f6082374"

type timerWrite40A1F0 struct {
	Address uint32 `json:"address"`
	Hex     string `json:"hex"`
}
type timerStimulus40A1F0 struct {
	PC         uint32             `json:"pc"`
	Address    uint32             `json:"address"`
	Width      uint32             `json:"width"`
	Site       string             `json:"site"`
	Occurrence int                `json:"occurrence"`
	Writes     []timerWrite40A1F0 `json:"writes"`
}
type timerPacket40A1F0 struct {
	Target    uint32             `json:"target"`
	Argc      int                `json:"argc32"`
	Registers map[string]uint32  `json:"register_effects_u32"`
	Writes    []timerWrite40A1F0 `json:"writes"`
}
type timerInput40A1F0 struct {
	Name          string                       `json:"name"`
	Entry         uint32                       `json:"entry"`
	Args          []uint32                     `json:"args_u32"`
	Words         map[string]uint32            `json:"initial_words"`
	Settings      []uint32                     `json:"settings6_u32"`
	Limits        []byte                       `json:"limits6_u8"`
	Services      map[string]timerPacket40A1F0 `json:"services"`
	AfterReads    []timerStimulus40A1F0        `json:"after_reads"`
	AfterWrites   []timerStimulus40A1F0        `json:"after_writes"`
	AfterServices []timerStimulus40A1F0        `json:"after_services"`
}
type timerOriginalEvent40A1F0 struct {
	Kind      string          `json:"kind"`
	PC        uint32          `json:"PC"`
	Address   uint32          `json:"address"`
	Width     uint32          `json:"width"`
	Value     json.RawMessage `json:"value"`
	Completed bool            `json:"completed"`
	Sequence  int             `json:"sequence"`
}
type timerOriginalService40A1F0 struct {
	Site   uint32            `json:"site"`
	Target uint32            `json:"target"`
	Argc   int               `json:"argc"`
	Args   []uint32          `json:"arguments_u32"`
	Frame  string            `json:"frame_hex"`
	Packet timerPacket40A1F0 `json:"packet"`
}
type timerOriginalMemory40A1F0 struct {
	Address uint32 `json:"address"`
	Before  string `json:"before"`
	After   string `json:"after"`
	Exact   bool   `json:"replay_exact"`
}
type timerOriginalRow40A1F0 struct {
	Name           string                       `json:"name"`
	Entry          uint32                       `json:"entry"`
	CW             uint16                       `json:"control_word"`
	InputSHA       string                       `json:"input_sha256"`
	Status         string                       `json:"status"`
	Valid          bool                         `json:"certificate_valid"`
	EAX            *uint32                      `json:"actual_EAX_u32"`
	EDX            *uint32                      `json:"actual_EDX_u32"`
	Pair           *uint64                      `json:"raw_EDX_EAX_pair_u64"`
	Events         []timerOriginalEvent40A1F0   `json:"events"`
	Services       []timerOriginalService40A1F0 `json:"services"`
	Memory         []timerOriginalMemory40A1F0  `json:"memory"`
	Used           map[string][]int             `json:"used_mutations"`
	Unused         map[string][]int             `json:"unused_mutations"`
	UnusedServices []uint32                     `json:"unused_service_sites"`
	Helpers        int                          `json:"helper_original_opcodes_executed"`
	Cleanup        []string                     `json:"hook_cleanup_errors"`
	Immutable      bool                         `json:"immutable_code_pages_unchanged"`
	Faults         []json.RawMessage            `json:"faults"`
}
type timerStep40A1F0 struct {
	Kind    string    `json:"kind"`
	Site    uint32    `json:"site"`
	Field   int       `json:"field,omitempty"`
	Address uint32    `json:"PE_address,omitempty"`
	Width   uint32    `json:"width,omitempty"`
	Word    uint32    `json:"word,omitempty"`
	Args    []uint32  `json:"arguments,omitempty"`
	Reply   [3]uint32 `json:"declared_EAX_ECX_EDX,omitempty"`
}
type timerReference40A1F0 struct {
	CW                               uint16
	Result                           uint64
	Steps                            []timerStep40A1F0
	BeforeA, AfterA, BeforeB, AfterB []byte
	Used, Unused                     map[string][]int
	UnusedServices                   []uint32
}
type timerChosen40A1F0 struct {
	Input timerInput40A1F0
	SHA   string
	Refs  []*timerReference40A1F0
}

var timerCalls40A1F0 = map[uint32][2]uint32{0x40a1fd: {0x40a5c0, 1}, 0x40a20f: {0x4d8f50, 2}, 0x40a233: {0x416bb0, 0}, 0x40a252: {0x40a5b0, 0}, 0x40a258: {0x409a70, 1}, 0x40a262: {0x416bb0, 0}, 0x40a311: {0x416bb0, 0}}

func timerHash40A1F0(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
func timerOpen40A1F0(name, digest string) (*os.File, *gzip.Reader, error) {
	f, e := os.Open(filepath.Join("testdata", "four_timers_40a1f0", name))
	if e != nil {
		return nil, nil, e
	}
	h := sha256.New()
	if _, e = io.Copy(h, f); e != nil {
		f.Close()
		return nil, nil, e
	}
	if hex.EncodeToString(h.Sum(nil)) != digest {
		f.Close()
		return nil, nil, fmt.Errorf("frozen timer gzip SHA %s", name)
	}
	if _, e = f.Seek(0, 0); e != nil {
		f.Close()
		return nil, nil, e
	}
	z, e := gzip.NewReader(f)
	if e != nil {
		f.Close()
		return nil, nil, e
	}
	return f, z, nil
}
func timerEOF40A1F0(d *json.Decoder, z *gzip.Reader) error {
	var extra any
	if e := d.Decode(&extra); e != io.EOF {
		return fmt.Errorf("exact JSON EOF: %v", e)
	}
	_, e := io.Copy(io.Discard, z)
	return e
}
func timerField40A1F0(address, width uint32) (int, error) {
	switch {
	case address == 0x588234 && width == 4:
		return 0, nil
	case address == 0x5d5320 && width == 4:
		return 1, nil
	case address == 0x5d5324 && width == 4:
		return 2, nil
	case address >= 0x5d5340 && address < 0x5d5346 && width == 1:
		return 3, nil
	}
	return -1, fmt.Errorf("unselected timer operand %x/%d", address, width)
}
func timerPacketWords40A1F0(p timerPacket40A1F0) ([3]uint32, error) {
	if len(p.Registers) != 3 {
		return [3]uint32{}, fmt.Errorf("raw three-register packet")
	}
	out := [3]uint32{}
	for i, k := range []string{"eax", "ecx", "edx"} {
		v, ok := p.Registers[k]
		if !ok {
			return out, fmt.Errorf("missing register packet %s", k)
		}
		out[i] = v
	}
	return out, nil
}
func timerSameWords40A1F0(a, b []uint32) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func timerReduce40A1F0(row *timerOriginalRow40A1F0, in timerInput40A1F0) (*timerReference40A1F0, error) {
	if !row.Valid || row.Status != "returned" || row.EAX == nil || row.EDX == nil || row.Pair == nil || row.Helpers != 0 || !row.Immutable || len(row.Cleanup) != 0 || len(row.Faults) != 0 {
		return nil, fmt.Errorf("authentic returned original timer certificate")
	}
	if *row.Pair != (uint64(*row.EDX)<<32 | uint64(*row.EAX)) {
		return nil, fmt.Errorf("original full register pair")
	}
	r := &timerReference40A1F0{CW: row.CW, Result: uint64(*row.EAX), Used: row.Used, Unused: row.Unused, UnusedServices: row.UnusedServices}
	if in.Entry == 0x40a250 || in.Entry == 0x40a310 {
		r.Result = *row.Pair
	}
	for _, m := range row.Memory {
		if !m.Exact {
			return nil, fmt.Errorf("original whole page replay")
		}
		before, e := hex.DecodeString(m.Before)
		if e != nil {
			return nil, e
		}
		after, e := hex.DecodeString(m.After)
		if e != nil || len(before) != 4096 || len(after) != 4096 {
			return nil, fmt.Errorf("whole original mapped page bytes")
		}
		switch m.Address {
		case 0x588000:
			o := int(0x58822c - m.Address)
			r.BeforeA = append([]byte(nil), before[o:o+84]...)
			r.AfterA = append([]byte(nil), after[o:o+84]...)
		case 0x5d5000:
			o := int(0x5d5318 - m.Address)
			r.BeforeB = append([]byte(nil), before[o:o+56]...)
			r.AfterB = append([]byte(nil), after[o:o+56]...)
		}
	}
	if len(r.BeforeA) != 84 || len(r.BeforeB) != 56 || len(in.Words) != 3 || len(in.Settings) != 6 || len(in.Limits) != 6 {
		return nil, fmt.Errorf("native scalar buffer/input domain")
	}
	if binary.LittleEndian.Uint32(r.BeforeA[8:]) != in.Words["00588234"] || binary.LittleEndian.Uint32(r.BeforeB[8:]) != in.Words["005d5320"] || binary.LittleEndian.Uint32(r.BeforeB[12:]) != in.Words["005d5324"] || !bytes.Equal(r.BeforeB[40:46], in.Limits) {
		return nil, fmt.Errorf("independent input words/limit bytes equal original BEFORE")
	}
	for i, v := range in.Settings {
		if binary.LittleEndian.Uint32(r.BeforeA[52+i*4:]) != v {
			return nil, fmt.Errorf("input settings6 original BEFORE")
		}
	}
	type ordered struct {
		sequence int
		step     timerStep40A1F0
	}
	var steps []ordered
	for _, e := range row.Events {
		if e.Kind != "read" && e.Kind != "write" {
			continue
		}
		if e.Address >= 0x22000000 {
			continue
		} // caller stack is original-only evidence
		field, err := timerField40A1F0(e.Address, e.Width)
		if err != nil || !e.Completed {
			return nil, fmt.Errorf("original semantic scalar event: %v", err)
		}
		s := timerStep40A1F0{Kind: e.Kind, Site: e.PC, Field: field, Address: e.Address, Width: e.Width}
		if e.Kind == "write" {
			if err = json.Unmarshal(e.Value, &s.Word); err != nil {
				return nil, err
			}
		} else {
			var text string
			if err = json.Unmarshal(e.Value, &text); err != nil {
				return nil, err
			}
			bits, err := hex.DecodeString(text)
			if err != nil || len(bits) != int(e.Width) {
				return nil, fmt.Errorf("original exact read value")
			}
			if len(bits) == 1 {
				s.Word = uint32(bits[0])
			} else {
				s.Word = binary.LittleEndian.Uint32(bits)
			}
		}
		steps = append(steps, ordered{e.Sequence, s})
	}
	for _, svc := range row.Services {
		contract, ok := timerCalls40A1F0[svc.Site]
		packet, declared := in.Services[fmt.Sprintf("%08x", svc.Site)]
		if !ok || !declared || svc.Target != contract[0] || svc.Argc != int(contract[1]) || len(svc.Args) != svc.Argc || packet.Target != svc.Target || packet.Argc != svc.Argc || !reflect.DeepEqual(packet, svc.Packet) {
			return nil, fmt.Errorf("original actual service/input packet contract")
		}
		words, err := timerPacketWords40A1F0(packet)
		if err != nil {
			return nil, err
		}
		seq := -1
		for _, ev := range row.Events {
			if ev.Kind != "host-frame-read" || ev.PC != svc.Target {
				continue
			}
			var text string
			if err = json.Unmarshal(ev.Value, &text); err != nil {
				return nil, err
			}
			frame, err := hex.DecodeString(text)
			if err != nil || len(frame) < 4 {
				return nil, fmt.Errorf("actual CALL frame")
			}
			if binary.LittleEndian.Uint32(frame) != svc.Site+5 {
				continue
			}
			if seq >= 0 || text != svc.Frame || len(frame) != 4*(1+svc.Argc) {
				return nil, fmt.Errorf("unique actual original service frame")
			}
			for i, v := range svc.Args {
				if binary.LittleEndian.Uint32(frame[4+i*4:]) != v {
					return nil, fmt.Errorf("original frame argument")
				}
			}
			seq = ev.Sequence
		}
		if seq < 0 {
			return nil, fmt.Errorf("actual service frame absent")
		}
		steps = append(steps, ordered{seq, timerStep40A1F0{Kind: "service", Site: svc.Site, Args: append([]uint32(nil), svc.Args...), Reply: words}})
	}
	sort.Slice(steps, func(i, j int) bool { return steps[i].sequence < steps[j].sequence })
	for _, s := range steps {
		r.Steps = append(r.Steps, s.step)
	}
	return r, nil
}
func timerLoad40A1F0() ([]timerChosen40A1F0, error) {
	f, z, e := timerOpen40A1F0("inputs-v1.json.gz", fourTimersInputSHA40A1F0)
	if e != nil {
		return nil, e
	}
	var input struct {
		Cases []json.RawMessage `json:"cases"`
	}
	dec := json.NewDecoder(z)
	e = dec.Decode(&input)
	if e == nil {
		e = timerEOF40A1F0(dec, z)
	}
	z.Close()
	f.Close()
	if e != nil {
		return nil, e
	}
	if len(input.Cases) != 1123 {
		return nil, fmt.Errorf("1123 independent timer inputs")
	}
	out := make([]timerChosen40A1F0, 1123)
	indices := map[string]int{}
	for i, raw := range input.Cases {
		if e = json.Unmarshal(raw, &out[i].Input); e != nil {
			return nil, e
		}
		in := out[i].Input
		if _, ok := indices[in.Name]; ok {
			return nil, fmt.Errorf("unique timer input")
		}
		indices[in.Name] = i
		var compact bytes.Buffer
		if e = json.Compact(&compact, raw); e != nil {
			return nil, e
		}
		out[i].SHA = timerHash40A1F0(compact.Bytes())
		argc := 0
		if in.Entry == 0x40a1f0 || in.Entry == 0x40a310 {
			argc = 1
		}
		if in.Entry != 0x40a1f0 && in.Entry != 0x40a230 && in.Entry != 0x40a250 && in.Entry != 0x40a310 || len(in.Args) != argc || len(in.Services) != 7 {
			return nil, fmt.Errorf("timer entry/raw argument/service count")
		}
	}
	f, z, e = timerOpen40A1F0("reference-v1.json.gz", fourTimersReferenceSHA40A1F0)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	defer z.Close()
	dec = json.NewDecoder(z)
	tok, e := dec.Token()
	if e != nil || tok != json.Delim('{') {
		return nil, fmt.Errorf("original object")
	}
	found, count := false, 0
	for dec.More() {
		key, e := dec.Token()
		if e != nil {
			return nil, e
		}
		if key != "rows" {
			var raw json.RawMessage
			if e = dec.Decode(&raw); e != nil {
				return nil, e
			}
			if (key == "fatal" || key == "fatal_recording_error") && !bytes.Equal(raw, []byte("null")) {
				return nil, fmt.Errorf("original recording fatal")
			}
			continue
		}
		if found {
			return nil, fmt.Errorf("duplicate rows")
		}
		found = true
		tok, e = dec.Token()
		if e != nil || tok != json.Delim('[') {
			return nil, fmt.Errorf("original rows array")
		}
		for dec.More() {
			var row timerOriginalRow40A1F0
			if e = dec.Decode(&row); e != nil {
				return nil, e
			}
			i, ok := indices[row.Name]
			if !ok || row.Entry != out[i].Input.Entry || row.InputSHA != out[i].SHA || row.CW != 0x0e3f && row.CW != 0x023f {
				return nil, fmt.Errorf("original identity/input/CW")
			}
			for _, prior := range out[i].Refs {
				if prior.CW == row.CW {
					return nil, fmt.Errorf("duplicate original case/CW")
				}
			}
			ref, e := timerReduce40A1F0(&row, out[i].Input)
			if e != nil {
				return nil, e
			}
			out[i].Refs = append(out[i].Refs, ref)
			count++
		}
		tok, e = dec.Token()
		if e != nil || tok != json.Delim(']') {
			return nil, fmt.Errorf("row array close")
		}
	}
	tok, e = dec.Token()
	if e != nil || tok != json.Delim('}') {
		return nil, fmt.Errorf("original object close")
	}
	if e = timerEOF40A1F0(dec, z); e != nil {
		return nil, e
	}
	if !found || count != 2246 {
		return nil, fmt.Errorf("2246 actual original rows")
	}
	for i := range out {
		if len(out[i].Refs) != 2 {
			return nil, fmt.Errorf("two original control words per input")
		}
	}
	return out, nil
}

type timerRun40A1F0 struct {
	t        *testing.T
	chosen   *timerChosen40A1F0
	next     int
	pending  *timerNativeOperand40A1F0
	observed []timerStep40A1F0
	occ      map[string]int
	used     map[string]map[int]bool
	services map[uint32]bool
}

func timerWindows40A1F0() ([]byte, []byte) {
	return memmap.Slice(0x587000, 4652)[:84], memmap.Slice(0x5d4594, 3460)[:56]
}
func timerSave40A1F0() func() {
	a, b := timerWindows40A1F0()
	aa, bb := bytes.Clone(a), bytes.Clone(b)
	return func() { copy(a, aa); copy(b, bb) }
}
func timerWriteInput40A1F0(t *testing.T, writes []timerWrite40A1F0) {
	t.Helper()
	a, b := timerWindows40A1F0()
	for _, w := range writes {
		bits, e := hex.DecodeString(w.Hex)
		if e != nil || len(bits) != 1 && len(bits) != 4 && len(bits) != 8 {
			t.Fatal("declared raw memory effect")
		}
		end := uint64(w.Address) + uint64(len(bits))
		var target []byte
		switch {
		case w.Address >= 0x588234 && end <= 0x588238:
			target = a[int(w.Address-0x58822c):]
		case w.Address >= 0x5d5320 && end <= 0x5d5328:
			target = b[int(w.Address-0x5d5318):]
		case w.Address >= 0x588260 && end <= 0x588278:
			target = a[int(w.Address-0x58822c):]
		case w.Address >= 0x5d5340 && end <= 0x5d5346:
			target = b[int(w.Address-0x5d5318):]
		default:
			t.Fatalf("effect outside selected42-byte domain %x/%d", w.Address, len(bits))
		}
		copy(target, bits)
	}
}
func (r *timerRun40A1F0) expect() timerStep40A1F0 {
	r.t.Helper()
	if r.next >= len(r.chosen.Refs[0].Steps) {
		r.t.Fatal("unexpected extra native observation")
	}
	return r.chosen.Refs[0].Steps[r.next]
}
func (r *timerRun40A1F0) mutate(kind string, site, address, width uint32) {
	key := fmt.Sprintf("%s/%x/%x/%d", kind, site, address, width)
	r.occ[key]++
	var list []timerStimulus40A1F0
	switch kind {
	case "after_reads":
		list = r.chosen.Input.AfterReads
	case "after_writes":
		list = r.chosen.Input.AfterWrites
	case "after_services":
		list = r.chosen.Input.AfterServices
	default:
		r.t.Fatal("known mutation kind")
	}
	for i, s := range list {
		match := s.PC == site && s.Address == address && s.Width == width
		if kind == "after_services" {
			v, e := strconv.ParseUint(s.Site, 16, 32)
			if e != nil {
				r.t.Fatal(e)
			}
			match = uint32(v) == site
		}
		if match && s.Occurrence == r.occ[key] {
			if r.used[kind][i] {
				r.t.Fatal("duplicate mutation use")
			}
			timerWriteInput40A1F0(r.t, s.Writes)
			r.used[kind][i] = true
		}
	}
}
func (r *timerRun40A1F0) operand(e timerNativeOperand40A1F0) {
	r.t.Helper()
	actualAddress := uint32(0)
	width := uint32(4)
	if e.Field == 3 {
		base := uintptr(fourTimersCellC40A1F0(0x5d4594, 3500))
		p := uintptr(e.Address)
		if p < base || p-base >= 6 {
			r.t.Fatal("actual live limit pointer domain")
		}
		actualAddress = 0x5d5340 + uint32(p-base)
		width = 1
	} else {
		offsets := []uint32{4660, 3468, 3472}
		if e.Field < 0 || e.Field >= 3 {
			r.t.Fatal("native field")
		}
		base := uint32(0x5d4594)
		if e.Field == 0 {
			base = 0x587000
		}
		if e.Address != fourTimersCellC40A1F0(base, offsets[e.Field]) {
			r.t.Fatal("actual native C/Go cell identity")
		}
		actualAddress = base + offsets[e.Field]
	}
	step := timerStep40A1F0{Kind: "read", Site: e.Site, Field: e.Field, Address: actualAddress, Width: e.Width, Word: e.Value}
	if e.Write {
		step.Kind = "write"
	}
	want := r.expect()
	if e.Width != width || want.Kind != step.Kind || want.Site != step.Site || want.Field != step.Field || want.Address != step.Address || want.Width != step.Width {
		r.t.Fatalf("actual operand identity/order got=%+v want=%+v", step, want)
	}
	if !e.Complete {
		if r.pending != nil || e.ValueValid != e.Write || !e.Write && e.Value != 0 || e.Write && e.Value != want.Word {
			r.t.Fatalf("before actual load/store contract: actual=%+v expected=%+v", e, want)
		}
		copy := e
		r.pending = &copy
		return
	}
	if r.pending == nil || !e.ValueValid || r.pending.Site != e.Site || r.pending.Address != e.Address || r.pending.Width != e.Width || r.pending.Write != e.Write || e.Value != want.Word {
		r.t.Fatalf("actual completed load/store value/order: actual=%+v expected=%+v", e, want)
	}
	if e.Write && *(*uint32)(e.Address) != e.Value {
		r.t.Fatal("actual DWORD store missing from live cell")
	}
	r.pending = nil
	r.observed = append(r.observed, step)
	r.next++
	kind := "after_reads"
	if e.Write {
		kind = "after_writes"
	}
	r.mutate(kind, e.Site, actualAddress, e.Width)
}
func (r *timerRun40A1F0) service(site uint32, args []uint32) (timerNativeReply40A1F0, bool) {
	r.t.Helper()
	want := r.expect()
	packet, ok := r.chosen.Input.Services[fmt.Sprintf("%08x", site)]
	contract, known := timerCalls40A1F0[site]
	if r.pending != nil || want.Kind != "service" || want.Site != site || !timerSameWords40A1F0(args, want.Args) || !ok || !known || packet.Target != contract[0] || packet.Argc != int(contract[1]) || len(args) != packet.Argc || r.services[site] {
		r.t.Fatalf("actual external site/full DWORD arguments/order: site=%x arguments=%v expected=%+v", site, args, want)
	}
	words, e := timerPacketWords40A1F0(packet)
	if e != nil {
		r.t.Fatal(e)
	}
	r.observed = append(r.observed, timerStep40A1F0{Kind: "service", Site: site, Args: append([]uint32(nil), args...), Reply: words})
	r.next++
	r.services[site] = true
	timerWriteInput40A1F0(r.t, packet.Writes)
	r.mutate("after_services", site, 0, 0)
	return timerNativeReply40A1F0{EAX: words[0], ECX: words[1], EDX: words[2]}, true
}
func timerCall40A1F0(backend string, in timerInput40A1F0) uint64 {
	arg := uint32(0)
	if len(in.Args) > 0 {
		arg = in.Args[0]
	}
	if backend == "C" {
		return fourTimersCallC40A1F0(in.Entry, arg)
	}
	switch in.Entry {
	case 0x40a1f0:
		return uint64(uint32(sub40A1F0Native(int32(arg))))
	case 0x40a230:
		return uint64(uint32(sub40A230Native()))
	case 0x40a250:
		return uint64(sub40A250Native())
	case 0x40a310:
		return uint64(sub40A310Native(int32(arg)))
	default:
		panic("selected timer entry")
	}
}
func timerIndexLists40A1F0(r *timerRun40A1F0) (map[string][]int, map[string][]int, []uint32) {
	u, n := map[string][]int{}, map[string][]int{}
	for kind, count := range map[string]int{"after_reads": len(r.chosen.Input.AfterReads), "after_writes": len(r.chosen.Input.AfterWrites), "after_services": len(r.chosen.Input.AfterServices)} {
		u[kind] = []int{}
		n[kind] = []int{}
		for i := 0; i < count; i++ {
			if r.used[kind][i] {
				u[kind] = append(u[kind], i)
			} else {
				n[kind] = append(n[kind], i)
			}
		}
	}
	var unused []uint32
	for site := range timerCalls40A1F0 {
		if !r.services[site] {
			unused = append(unused, site)
		}
	}
	sort.Slice(unused, func(i, j int) bool { return unused[i] < unused[j] })
	return u, n, unused
}
func timerCheckResult40A1F0(t *testing.T, c *timerChosen40A1F0, result uint64, a, b []byte) {
	t.Helper()
	for _, ref := range c.Refs {
		if result != ref.Result || !bytes.Equal(a, ref.AfterA) || !bytes.Equal(b, ref.AfterB) {
			t.Fatalf("actual native raw return/full140B AFTER vs original CW%04x input=%s actual=%016x original=%016x", ref.CW, c.Input.Name, result, ref.Result)
		}
	}
}

func TestFourTimersNativeABI40A1F0(t *testing.T) {
	for field := 0; field < 4; field++ {
		count := 1
		if field == 3 {
			count = 6
		}
		for i := 0; i < count; i++ {
			offset := []uint32{4660, 3468, 3472, 3500}[field] + uint32(i)
			base := uint32(0x5d4594)
			if field == 0 {
				base = 0x587000
			}
			p := timerNativeCell40A1F0(field, uint32(i))
			if uintptr(p) <= 0xffffffff || p != fourTimersCellC40A1F0(base, offset) {
				t.Fatal("actual full native scalar cell aliases")
			}
		}
	}
	t.Log("FOUR_TIMERS_NATIVE_ABI_OK Ctyped4 int32=2 int64=2 native-cell-pointers>4GiB declared42bytes packed-canary98bytes physical-stack-FP-credit=false")
}
func TestFourTimersOriginal40A1F0(t *testing.T) {
	chosen, e := timerLoad40A1F0()
	if e != nil {
		t.Fatal(e)
	}
	restore := timerSave40A1F0()
	defer restore()
	root := os.Getenv("OPENNOX_LOCAL_TEST_ARTIFACTS")
	if root == "" {
		root = os.Getenv("OPENNOX_RAW_DEFAULT_DAMAGE_INITIAL_ARTIFACTS")
	}
	if !filepath.IsAbs(root) {
		t.Fatal("absolute retained artifact root required")
	}
	if e = os.MkdirAll(root, 0700); e != nil {
		t.Fatal(e)
	}
	directory, e := os.MkdirTemp(root, "four-timers-native-")
	if e != nil {
		t.Fatal(e)
	}
	trace, e := os.OpenFile(filepath.Join(directory, "actual-four-timers-native.jsonl"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer trace.Close()
	encoder := json.NewEncoder(trace)
	calls, comparisons, mutations, services := 0, 0, 0, 0
	for index := range chosen {
		c := &chosen[index]
		for _, backend := range []string{"Go", "C"} {
			t.Run(c.Input.Name+"/"+backend, func(t *testing.T) {
				a, b := timerWindows40A1F0()
				copy(a, c.Refs[0].BeforeA)
				copy(b, c.Refs[0].BeforeB)
				for _, ref := range c.Refs {
					if !bytes.Equal(a, ref.BeforeA) || !bytes.Equal(b, ref.BeforeB) {
						t.Fatal("two original CW initial-state identity")
					}
				}
				r := &timerRun40A1F0{t: t, chosen: c, occ: map[string]int{}, used: map[string]map[int]bool{"after_reads": {}, "after_writes": {}, "after_services": {}}, services: map[uint32]bool{}}
				pop := pushTimerNativeObserver40A1F0(&timerNativeObserver40A1F0{Operand: r.operand, Service: r.service})
				var result uint64
				func() { defer pop(); result = timerCall40A1F0(backend, c.Input) }()
				calls++
				used, unused, unusedServices := timerIndexLists40A1F0(r)
				actualA, actualB := bytes.Clone(a), bytes.Clone(b)
				if e := encoder.Encode(map[string]any{"name": c.Input.Name, "backend": backend, "input_sha256": c.SHA, "entry": c.Input.Entry, "actual_return_bits_u64": result, "actual_steps": r.observed, "BEFORE_A": c.Refs[0].BeforeA, "BEFORE_B": c.Refs[0].BeforeB, "AFTER_A": actualA, "AFTER_B": actualB, "used_mutations": used, "unused_mutations": unused, "unused_services": unusedServices, "original_control_words": []uint16{c.Refs[0].CW, c.Refs[1].CW}}); e != nil {
					t.Fatal(e)
				}
				if r.pending != nil || r.next != len(c.Refs[0].Steps) {
					t.Fatal("all before/complete/service observations consumed")
				}
				timerCheckResult40A1F0(t, c, result, actualA, actualB)
				for _, ref := range c.Refs {
					if !reflect.DeepEqual(r.observed, ref.Steps) || !reflect.DeepEqual(used, ref.Used) || !reflect.DeepEqual(unused, ref.Unused) || !timerSameWords40A1F0(unusedServices, ref.UnusedServices) {
						t.Fatalf("raw actual event/service/used-unused sequence vs original CW%04x", ref.CW)
					}
					comparisons++
				}
				for _, list := range used {
					mutations += len(list)
				}
				services += len(r.services)
				t.Logf("FOUR_TIMERS_ORIGINAL_OK input=%s backend=%s original-CW-comparisons=2 raw-return=%016x native-full-buffer=140 declared-field=42 packed-canary=98", c.Input.Name, backend, result)
			})
		}
	}
	if e = trace.Sync(); e != nil {
		t.Fatal(e)
	}
	if e = trace.Close(); e != nil {
		t.Fatal(e)
	}
	summary := map[string]any{"schema": "opennox-four-timers-native-actual-summary-v1", "inputs": len(chosen), "native_Go_C_calls": calls, "original_CW_comparisons": comparisons, "actual_input_mutations": mutations, "actual_service_packets": services, "input_fixture_sha256": fourTimersInputSHA40A1F0, "reference_fixture_sha256": fourTimersReferenceSHA40A1F0, "actual_full_buffer_bytes": 140, "declared_field_bytes": 42, "packed_blob_canary_bytes": 98, "original_helper_opcode_credit": 0, "native_physical_ECX_EDX_stack_FP_equivalence": false, "new_original_execution": 0}
	data, e := json.MarshalIndent(summary, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(directory, "actual-four-timers-summary.json"), append(data, '\n'), 0600); e != nil {
		t.Fatal(e)
	}
	if calls != 2246 || comparisons != 4492 {
		t.Fatalf("all actual native calls/reference comparisons calls=%d comparisons=%d", calls, comparisons)
	}
}

func timerClockChosen40A1F0(t *testing.T, chosen []timerChosen40A1F0, entry uint32) *timerChosen40A1F0 {
	t.Helper()
	site := uint32(0x40a233)
	if entry == 0x40a310 {
		site = 0x40a311
	}
	for i := range chosen {
		c := &chosen[i]
		if c.Input.Entry != entry || len(c.Input.AfterReads)+len(c.Input.AfterWrites)+len(c.Input.AfterServices) != 0 {
			continue
		}
		p := c.Input.Services[fmt.Sprintf("%08x", site)]
		if len(p.Writes) == 0 && p.Registers["eax"] == 1 && p.Registers["edx"] == 1 {
			return c
		}
	}
	t.Fatal("frozen full64 clock input control")
	return nil
}
func TestFourTimersDefaultClock40A1F0(t *testing.T) {
	chosen, e := timerLoad40A1F0()
	if e != nil {
		t.Fatal(e)
	}
	restore := timerSave40A1F0()
	defer restore()
	old := PlatformTicks
	defer func() { PlatformTicks = old }()
	for _, entry := range []uint32{0x40a230, 0x40a310} {
		c := timerClockChosen40A1F0(t, chosen, entry)
		for _, backend := range []string{"Go", "C"} {
			t.Run(fmt.Sprintf("%06X/%s", entry, backend), func(t *testing.T) {
				a, b := timerWindows40A1F0()
				copy(a, c.Refs[0].BeforeA)
				copy(b, c.Refs[0].BeforeB)
				calls := 0
				PlatformTicks = func() uint64 { calls++; return 0x100000001 }
				pop := pushTimerNativeObserver40A1F0(nil)
				var result uint64
				func() { defer pop(); result = timerCall40A1F0(backend, c.Input) }()
				if calls != 1 {
					t.Fatal("actual nil-observer PlatformTicks dispatch")
				}
				timerCheckResult40A1F0(t, c, result, a, b)
				t.Logf("FOUR_TIMERS_DEFAULT_CLOCK_OK entry=%06x backend=%s observer=nil actual-PlatformTicks-dispatch=1 full64-input=0000000100000001 original-helper-body-credit=0 Notify-executions=0", entry, backend)
			})
		}
	}
}
func TestFourTimersObserverReentry40A1F0(t *testing.T) {
	chosen, e := timerLoad40A1F0()
	if e != nil {
		t.Fatal(e)
	}
	restore := timerSave40A1F0()
	defer restore()
	c := timerClockChosen40A1F0(t, chosen, 0x40a230)
	p := c.Input.Services["0040a233"]
	words, e := timerPacketWords40A1F0(p)
	if e != nil {
		t.Fatal(e)
	}
	reply := timerNativeReply40A1F0{EAX: words[0], ECX: words[1], EDX: words[2]}
	for _, backend := range []string{"Go", "C"} {
		t.Run(backend, func(t *testing.T) {
			a, b := timerWindows40A1F0()
			copy(a, c.Refs[0].BeforeA)
			copy(b, c.Refs[0].BeforeB)
			outer, inner := 0, 0
			pop := pushTimerNativeObserver40A1F0(&timerNativeObserver40A1F0{Service: func(site uint32, args []uint32) (timerNativeReply40A1F0, bool) {
				if site != 0x40a233 || len(args) != 0 {
					t.Fatal("outer actual Clock site")
				}
				outer++
				nestedPop := pushTimerNativeObserver40A1F0(&timerNativeObserver40A1F0{Service: func(site uint32, args []uint32) (timerNativeReply40A1F0, bool) {
					if site != 0x40a233 || len(args) != 0 {
						t.Fatal("inner actual Clock site")
					}
					inner++
					return reply, true
				}})
				var result uint64
				func() { defer nestedPop(); result = timerCall40A1F0(backend, c.Input) }()
				timerCheckResult40A1F0(t, c, result, a, b)
				return reply, true
			}})
			var result uint64
			func() { defer pop(); result = timerCall40A1F0(backend, c.Input) }()
			timerCheckResult40A1F0(t, c, result, a, b)
			if outer != 1 || inner != 1 {
				t.Fatal("nested observer scope restoration")
			}
			t.Logf("FOUR_TIMERS_REENTRY_OK backend=%s outer=1 inner=1 rawreply-preserved=true original-helper-body-credit=0", backend)
		})
	}
	savedStack := timerNativeObserverStack40A1F0
	defer func() { timerNativeObserverStack40A1F0 = savedStack }() // restore serial test scope even when the guard itself regresses
	outer := pushTimerNativeObserver40A1F0(nil)
	inner := pushTimerNativeObserver40A1F0(nil)
	innerTop := timerNativeObserverStack40A1F0
	panicked := false
	func() { defer func() { panicked = recover() != nil }(); outer() }()
	if !panicked || timerNativeObserverStack40A1F0 != innerTop {
		t.Fatal("non-LIFO pop must fail")
	}
	inner()
	outer()
}
