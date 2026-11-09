//go:build darwin && arm64

package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"strconv"
	"testing"
	"unsafe"
)

type spWrite4134D0 struct {
	Address uint32 `json:"address"`
	Hex     string `json:"hex"`
}
type spRegion4134D0 struct {
	Name        string `json:"name"`
	Page, Guard uint32
	Prefix      int  `json:"prefix_bytes"`
	Seed        byte `json:"byte_seed"`
}
type spPacket4134D0 struct {
	Registers map[string]uint32 `json:"register_effects_u32"`
	Writes    []spWrite4134D0   `json:"writes"`
}
type spService4134D0 struct {
	Packets []spPacket4134D0 `json:"packets"`
}
type spStimulus4134D0 struct {
	PC, Address, Width uint32
	Site               string
	Occurrence         int
	Writes             []spWrite4134D0
}
type spInput4134D0 struct {
	Name, Family  string
	Entry         uint32
	Args          []uint32 `json:"args_u32"`
	Regions       []spRegion4134D0
	Roles         map[string]uint32
	Words         map[string]uint32 `json:"initial_words"`
	Drive         byte              `json:"initial_drive_byte"`
	BufferSeed    byte              `json:"buffer_seed"`
	Writes        []spWrite4134D0   `json:"initial_writes"`
	Services      map[string]spService4134D0
	AfterReads    []spStimulus4134D0 `json:"after_reads"`
	AfterServices []spStimulus4134D0 `json:"after_services"`
}
type spPage4134D0 struct {
	Address       uint32 `json:"address"`
	Before, After string
}
type spOriginalEvent4134D0 struct {
	Kind              string
	PC, Site, Address uint32
	Width             int
	Value             json.RawMessage
	After             json.RawMessage
	Completed         bool
	Arguments         []uint32 `json:"arguments_u32"`
}
type spOriginalService4134D0 struct {
	Site      uint32
	Arguments []uint32 `json:"arguments_u32"`
}
type spReference4134D0 struct {
	Name, Status string
	CW           uint16  `json:"control_word"`
	Valid        bool    `json:"certificate_valid"`
	Result       *uint32 `json:"actual_EAX_u32"`
	Memory       []spPage4134D0
	Events       []spOriginalEvent4134D0
	Services     []spOriginalService4134D0
	Used         map[string][]int `json:"used_mutations"`
	Packets      map[string][]int `json:"used_service_packets"`
	Fault        *struct {
		PC, Address uint32
		Width       int
	} `json:"read_fault_mapping_closure"`
}
type spTrace4134D0 struct {
	Kind          string
	Site, Address uint32
	Width         int
	Value         string
	Args          []uint32
}
type spRawOperand4134D0 struct {
	Site                                uint32
	Field                               int
	Base, Address, Offset               uint64
	Width, Value                        uint32
	Write, Complete, ValueValid, Policy bool
}
type spRawCall4134D0 struct {
	Name, Backend, Status            string
	Entry, Result                    uint32
	Policy                           bool
	Pages                            []spPage4134D0
	Banks                            map[uint32]uint64
	Arguments                        []uint64
	Operands                         []spRawOperand4134D0
	Trace                            []spTrace4134D0
	Used                             map[string][]int
	Packets                          map[string][]int
	CallbackPointers, StringPointers map[uint32]uint64
	NativePageBytes                  int
}

var spCells4134D0 = [...]uint32{0x611cd0, 0x611cd4, 0x611cd8, 0x611cdc, 0x611cf0, 0x611cf4}

func spHex4134D0(s string) []byte {
	b, e := hex.DecodeString(s)
	if e != nil {
		panic(e)
	}
	return b
}
func spPattern4134D0(seed byte) []byte {
	b := make([]byte, 4096)
	for i := range b {
		b[i] = byte(int(seed) + 29*i + 7*(i>>4))
	}
	return b
}
func spInitial4134D0(in *spInput4134D0) map[uint32][]byte {
	m := map[uint32][]byte{0x611000: spPattern4134D0(0x71), 0x663000: spPattern4134D0(0x92)}
	for _, r := range in.Regions {
		m[r.Page] = spPattern4134D0(r.Seed)
	}
	for k, v := range in.Words {
		a, e := strconv.ParseUint(k, 16, 32)
		if e != nil {
			panic(e)
		}
		binary.LittleEndian.PutUint32(m[uint32(a)&^4095][uint32(a)&4095:], v)
	}
	m[0x663000][0xee8] = in.Drive
	copy(m[0x611000][0xc88:], spPattern4134D0(in.BufferSeed)[:12])
	for _, w := range in.Writes {
		copy(m[w.Address&^4095][w.Address&4095:], spHex4134D0(w.Hex))
	}
	return m
}

func spRun4134D0(t *testing.T, in *spInput4134D0, backend string) (out spRawCall4134D0) {
	t.Helper()
	page := os.Getpagesize()
	if page < 8192 {
		t.Fatal("native guard page")
	}
	out = spRawCall4134D0{Name: in.Name, Backend: backend, Entry: in.Entry, Policy: in.Entry == 0x4134d0 || in.Entry == 0x413890, Status: "returned", Banks: map[uint32]uint64{}, Used: map[string][]int{"after_reads": {}, "after_services": {}}, Packets: map[string][]int{}, CallbackPointers: map[uint32]uint64{}, StringPointers: map[uint32]uint64{}, NativePageBytes: page}
	initial := spInitial4134D0(in)
	banks := map[uint32]unsafe.Pointer{}
	var owners []unsafe.Pointer
	defer func() {
		for _, p := range owners {
			startupPauseGuardUnmap4134D0(p)
		}
	}()
	for _, a := range []uint32{0x611000, 0x663000, 0x21000000, 0x21004000} {
		p := startupPauseGuardMap4134D0()
		if p == nil {
			t.Fatal("guard mmap")
		}
		owners = append(owners, p)
		if uintptr(p) <= 0xffffffff {
			t.Fatal("native bank is not above4GiB")
		}
		whole := unsafe.Slice((*byte)(p), page)
		for i := range whole {
			whole[i] = 0xa6
		}
		copy(whole[page-4096:], initial[a])
		banks[a] = unsafe.Add(p, page-4096)
		out.Banks[a] = uint64(uintptr(banks[a]))
	}
	for _, a := range []uint32{0x21f02000, 0x81002000, 0xffffffff} {
		p := startupPauseGuardMap4134D0()
		if p == nil {
			t.Fatal("opaque guard mmap")
		}
		owners = append(owners, p)
		banks[a] = unsafe.Add(p, page)
		out.Banks[a] = uint64(uintptr(banks[a]))
	}
	address := func(a uint32) unsafe.Pointer {
		p, ok := banks[a&^4095]
		if !ok {
			panic(fmt.Sprintf("undeclared memory%08x", a))
		}
		return unsafe.Add(p, a&4095)
	}
	resolve := func(token uint32) unsafe.Pointer {
		if token == 0 {
			return nil
		}
		if p, ok := banks[token]; ok {
			return p
		}
		for _, r := range in.Regions {
			if token == r.Guard-uint32(r.Prefix) {
				return unsafe.Add(banks[r.Page], 4096-r.Prefix)
			}
		}
		panic("unknown pointer token")
	}
	var args [2]unsafe.Pointer
	if in.Entry == 0x413f60 {
		args = [2]unsafe.Pointer{resolve(in.Args[0]), resolve(in.Args[1])}
		out.Arguments = []uint64{uint64(uintptr(args[0])), uint64(uintptr(args[1]))}
	}
	for _, a := range []uint32{0x611000, 0x663000, 0x21000000, 0x21004000} {
		out.Pages = append(out.Pages, spPage4134D0{Address: a, Before: hex.EncodeToString(unsafe.Slice((*byte)(banks[a]), 4096))})
	}
	observer := &startupPauseObserver4134D0{Buffer: address(0x611c88)}
	for i, a := range spCells4134D0 {
		observer.Cells[i] = address(a)
	}
	reads := map[string]int{}
	counts := map[uint32]int{}
	var last startupPauseOperand4134D0
	apply := func(writes []spWrite4134D0) {
		for _, w := range writes {
			b := spHex4134D0(w.Hex)
			copy(unsafe.Slice((*byte)(address(w.Address)), len(b)), b)
			out.Trace = append(out.Trace, spTrace4134D0{Kind: "host-write", Address: w.Address, Width: len(b), Value: w.Hex})
		}
	}
	observer.Operand = func(e startupPauseOperand4134D0) {
		last = e
		out.Operands = append(out.Operands, spRawOperand4134D0{e.Site, e.Field, uint64(uintptr(e.Base)), uint64(uintptr(e.Address)), uint64(e.Offset), e.Width, e.Value, e.Write, e.Complete, e.ValueValid, e.Policy})
		if !e.Complete {
			return
		}
		var a uint32
		switch {
		case e.Field < 6:
			a = spCells4134D0[e.Field]
		case e.Field == 6:
			a = 0x611c88 + uint32(e.Offset)
		default:
			a = in.Args[e.Field-7] + uint32(e.Offset)
		}
		b := make([]byte, 4)
		binary.LittleEndian.PutUint32(b, e.Value)
		kind := "read"
		if e.Write {
			kind = "write"
		}
		out.Trace = append(out.Trace, spTrace4134D0{Kind: kind, Site: e.Site, Address: a, Width: int(e.Width), Value: hex.EncodeToString(b[:e.Width])})
		if !e.Write {
			key := fmt.Sprintf("%x/%x/%d", e.Site, a, e.Width)
			reads[key]++
			for i, s := range in.AfterReads {
				if s.PC == e.Site && s.Address == a && s.Width == e.Width && s.Occurrence == reads[key] {
					out.Used["after_reads"] = append(out.Used["after_reads"], i)
					apply(s.Writes)
				}
			}
		}
	}
	observer.Service = func(site uint32, args []startupPauseArg4134D0) (uint32, bool) {
		words := make([]uint32, 0, len(args))
		for _, arg := range args {
			switch arg.Kind {
			case "word":
				words = append(words, arg.Word)
			case "callback":
				kind := args[0].Word
				if arg.Pointer != startupPauseCallback4134D0(kind) || uintptr(arg.Pointer) <= 0xffffffff {
					t.Fatal("actual typed high callback")
				}
				out.CallbackPointers[kind] = uint64(uintptr(arg.Pointer))
				if kind == 1 {
					words = append(words, 0x413900)
				} else {
					words = append(words, 0x4138e0)
				}
			case "cstring":
				kind := args[0].Word
				if arg.Pointer != startupPauseString4134D0(kind) || uintptr(arg.Pointer) <= 0xffffffff || arg.Text != "Pause" {
					t.Fatal("actual Pause C string role")
				}
				out.StringPointers[kind] = uint64(uintptr(arg.Pointer))
				if kind == 1 {
					words = append(words, 0x58ee44)
				} else {
					words = append(words, 0x58ee4c)
				}
			default:
				t.Fatal("unknown native service arg")
			}
		}
		out.Trace = append(out.Trace, spTrace4134D0{Kind: "service", Site: site, Args: words})
		key := fmt.Sprintf("%08x", site)
		n := counts[site]
		packets := in.Services[key].Packets
		if n >= len(packets) {
			t.Fatal("finite native service packets exhausted")
		}
		p := packets[n]
		counts[site]++
		out.Packets[key] = append(out.Packets[key], n)
		apply(p.Writes)
		for i, s := range in.AfterServices {
			if s.Site == key && s.Occurrence == counts[site] {
				out.Used["after_services"] = append(out.Used["after_services"], i)
				apply(s.Writes)
			}
		}
		return p.Registers["eax"], true
	}
	pop := pushStartupPauseObserver4134D0(observer)
	defer pop()
	func() {
		old := debug.SetPanicOnFault(true)
		defer debug.SetPanicOnFault(old)
		defer func() {
			if r := recover(); r != nil {
				if last.Complete || last.Write || (last.Field != 7 && last.Field != 8) {
					panic(r)
				}
				out.Status = "original_read_fault"
				out.Trace = append(out.Trace, spTrace4134D0{Kind: "fault", Site: last.Site, Address: in.Args[last.Field-7] + 80, Width: 4})
			}
		}()
		if in.Entry == 0x413890 {
			var result unsafe.Pointer
			if backend == "Go" {
				result = sub413890Native()
			} else {
				result = startupPauseFixtureDrive4134D0()
			}
			if result != nil {
				t.Fatal("disabled current CD fallback")
			}
			return
		}
		arg := int32(0)
		if len(in.Args) == 1 {
			arg = int32(in.Args[0])
		}
		if backend == "C" {
			out.Result = uint32(startupPauseFixtureCall4134D0(in.Entry, arg, args[0], args[1]))
			return
		}
		switch in.Entry {
		case 0x4134d0:
			out.Result = uint32(sub4134D0Native())
		case 0x4137e0:
			sub4137E0Native()
		case 0x4138e0:
			sub4138E0Native(arg)
		case 0x413900:
			sub413900Native(arg)
		case 0x413920:
			out.Result = uint32(sub413920Native())
		case 0x4139b0:
			out.Result = uint32(sub4139B0Native())
		case 0x413f60:
			out.Result = uint32(sub413F60Native(args[0], args[1]))
		default:
			t.Fatal("entry")
		}
	}()
	for i := range out.Pages {
		a := out.Pages[i].Address
		out.Pages[i].After = hex.EncodeToString(unsafe.Slice((*byte)(banks[a]), 4096))
		whole := unsafe.Slice((*byte)(unsafe.Add(banks[a], -(page-4096))), page-4096)
		for _, b := range whole {
			if b != 0xa6 {
				t.Fatal("native bank extension canary")
			}
		}
	}
	return out
}

func spExpectedTrace4134D0(r *spReference4134D0) []spTrace4134D0 {
	var out []spTrace4134D0
	services := map[uint32][]uint32{}
	for _, s := range r.Services {
		services[s.Site] = s.Arguments
	}
	for _, e := range r.Events {
		switch e.Kind {
		case "read", "write":
			if e.Address&^4095 == 0x22000000 || !e.Completed {
				continue
			}
			var value string
			if e.Kind == "read" {
				if err := json.Unmarshal(e.Value, &value); err != nil {
					panic(err)
				}
			} else {
				if err := json.Unmarshal(e.After, &value); err != nil {
					panic(err)
				}
			}
			out = append(out, spTrace4134D0{Kind: e.Kind, Site: e.PC, Address: e.Address, Width: e.Width, Value: value})
		case "host-frame-read":
			out = append(out, spTrace4134D0{Kind: "service", Site: e.Site, Args: services[e.Site]})
		case "host-write":
			var value string
			if err := json.Unmarshal(e.After, &value); err != nil {
				panic(err)
			}
			out = append(out, spTrace4134D0{Kind: "host-write", Address: e.Address, Width: e.Width, Value: value})
		}
	}
	if r.Status == "original_read_fault" {
		if r.Fault == nil {
			panic("missing original fault closure")
		}
		out = append(out, spTrace4134D0{Kind: "fault", Site: r.Fault.PC, Address: r.Fault.Address, Width: r.Fault.Width})
	}
	return out
}

func TestStartupPauseNativeOriginal4134D0(t *testing.T) {
	var inputs struct{ Cases []spInput4134D0 }
	var reference struct{ Rows []spReference4134D0 }
	if e := swReadGzipJSON410550("testdata/startup_pause_4134d0/inputs.json.gz", "826b95ee3c8de7412bb8e0dbf3988f92cc0b63aa16d6f0886d4967fa8212f5b5", &inputs); e != nil {
		t.Fatal(e)
	}
	if e := swReadGzipJSON410550("testdata/startup_pause_4134d0/original.json.gz", "c976c73ec69f3f771e19068d4199bc0910677dd2fae1019ef79cbb04460cd5ea", &reference); e != nil {
		t.Fatal(e)
	}
	if len(inputs.Cases) != 535 || len(reference.Rows) != 1070 {
		t.Fatal("frozen corpus counts")
	}
	rows := map[string]map[uint16]*spReference4134D0{}
	for i := range reference.Rows {
		r := &reference.Rows[i]
		if !r.Valid {
			t.Fatal("original qualification")
		}
		if rows[r.Name] == nil {
			rows[r.Name] = map[uint16]*spReference4134D0{}
		}
		if rows[r.Name][r.CW] != nil {
			t.Fatal("duplicate original row")
		}
		rows[r.Name][r.CW] = r
	}
	var raw *os.File
	if dir := os.Getenv("OPENNOX_STARTUP_PAUSE_NATIVE_DIRECTORY"); dir != "" {
		if e := os.MkdirAll(dir, 0700); e != nil {
			t.Fatal(e)
		}
		f, e := os.OpenFile(filepath.Join(dir, "actual-calls.jsonl"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			t.Fatal(e)
		}
		raw = f
		defer raw.Close()
	}
	for _, backend := range []string{"Go", "C"} {
		t.Run(backend, func(t *testing.T) {
			for i := range inputs.Cases {
				in := &inputs.Cases[i]
				t.Run(in.Name, func(t *testing.T) {
					actual := spRun4134D0(t, in, backend)
					if raw != nil {
						if e := json.NewEncoder(raw).Encode(actual); e != nil {
							t.Fatal(e)
						}
						if e := raw.Sync(); e != nil {
							t.Fatal(e)
						}
					}
					initial := spInitial4134D0(in)
					for _, p := range actual.Pages {
						if p.Before != hex.EncodeToString(initial[p.Address]) {
							t.Fatal("actual native BEFORE independently chosen input")
						}
					}
					if actual.Policy {
						if actual.Status != "returned" || actual.Result != 0 {
							t.Fatal("current startup/CD policy result")
						}
						for _, p := range actual.Pages {
							want := append([]byte(nil), initial[p.Address]...)
							if p.Address == 0x611000 {
								if in.Entry == 0x4134d0 {
									clear(want[0xcd0:0xce0])
								} else {
									clear(want[0xc88:0xc8a])
								}
							}
							if !bytes.Equal(spHex4134D0(p.After), want) {
								t.Fatal("current policy whole-page canaries")
							}
						}
						return // explicit platform policy; no whole original equivalence credit.
					}
					for _, cw := range []uint16{0xe3f, 0x23f} {
						r := rows[in.Name][cw]
						if r == nil {
							t.Fatal("missing original CW")
						}
						if actual.Status != r.Status {
							t.Fatalf("status %s vs%s", actual.Status, r.Status)
						}
						if actual.Status == "returned" && (in.Entry == 0x413920 || in.Entry == 0x4139b0 || in.Entry == 0x413f60) {
							if r.Result == nil || actual.Result != *r.Result {
								t.Fatal("raw32 original return")
							}
						}
						for _, p := range actual.Pages {
							found := false
							for _, rp := range r.Memory {
								if p.Address == rp.Address {
									found = true
									if p.Before != rp.Before || p.After != rp.After {
										t.Fatalf("full page %08x original BEFORE/AFTER", p.Address)
									}
								}
							}
							if !found {
								t.Fatal("missing original data page")
							}
						}
						if want := spExpectedTrace4134D0(r); !reflect.DeepEqual(actual.Trace, want) {
							a, _ := json.Marshal(actual.Trace)
							b, _ := json.Marshal(want)
							t.Fatalf("ordered operand/service trace actual=%s original=%s", a, b)
						}
						for _, k := range []string{"after_reads", "after_services"} {
							if !reflect.DeepEqual(actual.Used[k], r.Used[k]) {
								t.Fatal("actual used input effects")
							}
						}
						for site, used := range r.Packets {
							if len(actual.Packets[site]) != len(used) {
								t.Fatal("actual finite packet consumption")
							}
						}
					}
				})
			}
		})
	}
}
