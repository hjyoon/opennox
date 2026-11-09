package legacy

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

type gameSettingsInput4161E0 struct {
	Name       string
	Entry      uint32
	Args       [2]int32
	ReturnKind string "json:\"return_kind\""
	StateHex   string "json:\"state_hex\""
	Dirty      uint32
	Words      [9][2]uint32
	Spells     [5]uint32
	ItemsHex   string "json:\"items_hex\""
	Strings    [2][2]string
}
type gameSettingsDefinition4161E0 struct {
	Name  string
	Bytes int
	SHA   string "json:\"sha256\""
}
type gameSettingsInputs4161E0 struct {
	Schema      string
	Cases       []gameSettingsInput4161E0
	Definitions []gameSettingsDefinition4161E0 "json:\"old_C_definitions\""
}
type gameSettingsReference4161E0 struct {
	Index         int
	Entry         uint32
	Word          uint32
	Dirty         uint32
	PointerOffset int64  "json:\"pointer_offset\""
	ActualPointer string "json:\"actual_pointer\""
	Allocation    string
	CharSigned    bool   "json:\"char_signed\""
	Outside       bool   "json:\"outside_unchanged\""
	AfterHex      string "json:\"after_hex\""
	Trace         [][2]int64
}

func gameSettingsLoad4161E0(t *testing.T) (gameSettingsInputs4161E0, []gameSettingsReference4161E0) {
	t.Helper()
	raw, e := os.ReadFile("testdata/game_settings_4161e0_inputs_v1.json")
	if e != nil {
		t.Fatal(e)
	}
	var inputs gameSettingsInputs4161E0
	if e = json.Unmarshal(raw, &inputs); e != nil {
		t.Fatal(e)
	}
	if inputs.Schema != "opennox-settings12-independent-current-C-controls-v1" || len(inputs.Cases) != 22 || len(inputs.Definitions) != 12 {
		t.Fatal("settings independent-input scope")
	}
	f, e := os.Open("testdata/game_settings_4161e0_old_control.jsonl")
	if e != nil {
		t.Fatalf("root must run the preserved old-C control BEFORE installation and retain its exact reference: %v", e)
	}
	defer f.Close()
	scan := bufio.NewScanner(f)
	scan.Buffer(make([]byte, 4096), 2<<20)
	var header struct {
		Schema      string
		InputsSHA   string                         "json:\"inputs_sha256\""
		Count       int                            "json:\"case_count\""
		Definitions []gameSettingsDefinition4161E0 "json:\"old_C_definitions\""
		OriginalCPU bool                           "json:\"original_CPU_equivalence\""
		Wait        struct {
			Exit     int "json:\"exit_code\""
			Forced   bool
			Terminal bool "json:\"Wait_terminal\""
		} "json:\"actual_control_Wait\""
	}
	if !scan.Scan() {
		t.Fatal("missing old-C header")
	}
	if e = json.Unmarshal(scan.Bytes(), &header); e != nil {
		t.Fatal(e)
	}
	sum := sha256.Sum256(raw)
	if header.Schema != "opennox-settings12-current-preserved-C-controlled-reference-v1" || header.InputsSHA != hex.EncodeToString(sum[:]) || header.Count != len(inputs.Cases) || header.OriginalCPU || header.Wait.Exit != 0 || header.Wait.Forced || !header.Wait.Terminal || !reflect.DeepEqual(header.Definitions, inputs.Definitions) {
		t.Fatal("actual preserved C provenance/input binding")
	}
	refs := make([]gameSettingsReference4161E0, 0, len(inputs.Cases))
	for scan.Scan() {
		var row gameSettingsReference4161E0
		if e = json.Unmarshal(scan.Bytes(), &row); e != nil {
			t.Fatal(e)
		}
		refs = append(refs, row)
	}
	if e = scan.Err(); e != nil {
		t.Fatal(e)
	}
	if len(refs) != len(inputs.Cases) {
		t.Fatal("exact old-C row count")
	}
	for i, r := range refs {
		if r.Index != i || r.Entry != inputs.Cases[i].Entry || !r.Outside || r.CharSigned != gameSettingsPlainCharSigned4161E0 {
			t.Fatalf("old-C identity/char ABI %d", i)
		}
		base, e := strconv.ParseUint(r.Allocation[2:], 16, 64)
		if e != nil {
			t.Fatal(e)
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && base <= 0xffffffff {
			t.Fatal("old-C high native allocation missing")
		}
		if inputs.Cases[i].ReturnKind == "pointer" {
			p, e := strconv.ParseUint(r.ActualPointer[2:], 16, 64)
			if e != nil || p != base+uint64(r.PointerOffset) {
				t.Fatal("old-C pointer result/address relation")
			}
		}
	}
	return inputs, refs
}
func gameSettingsCall4161E0(backend string, in gameSettingsInput4161E0) (uint32, unsafe.Pointer) {
	if backend == "C" {
		if in.ReturnKind == "pointer" {
			return 0, gameSettingsFixturePointer4161E0(in.Entry, in.Args[0])
		}
		return uint32(gameSettingsFixtureWord4161E0(in.Entry, in.Args[0], in.Args[1])), nil
	}
	switch in.Entry {
	case 0x4161e0:
		return uint32(sub4161E0Native()), nil
	case 0x4164f0:
		sub4164F0Native()
		return 0, nil
	case 0x416580:
		return uint32(sub416580Native()), nil
	case 0x4165b0:
		return 0, sub4165B0Native()
	case 0x4165d0:
		return 0, sub4165D0Native(in.Args[0])
	case 0x4165f0:
		return uint32(sub4165F0Native(in.Args[0], in.Args[1])), nil
	case 0x416630:
		return 0, sub416630Native()
	case 0x416640:
		return 0, sub416640Native()
	case 0x416650:
		return uint32(sub416650Native()), nil
	case 0x416690:
		sub416690Native()
		return 0, nil
	case 0x4169c0:
		return uint32(sub4169C0Native()), nil
	case 0x4169f0:
		return 0, sub4169F0Native()
	}
	panic("unknown settings entry")
}
func gameSettingsRunInput4161E0(t *testing.T, backend string, in gameSettingsInput4161E0, want gameSettingsReference4161E0) {
	t.Helper()
	p := gameSettingsFixtureAlloc4161E0(640)
	dirty := gameSettingsFixtureAlloc4161E0(48)
	if p == nil || dirty == nil {
		t.Fatal("calloc")
	}
	defer gameSettingsFixtureFree4161E0(p)
	defer gameSettingsFixtureFree4161E0(dirty)
	block := unsafe.Slice((*byte)(p), 640)
	for i := range block {
		block[i] = 0xbd
	}
	window := block[64:576]
	initial, e := hex.DecodeString(in.StateHex)
	if e != nil || len(initial) != 512 {
		t.Fatal("initial512")
	}
	copy(window, initial)
	db := unsafe.Slice((*byte)(dirty), 48)
	for i := range db {
		db[i] = 0xce
	}
	binary.LittleEndian.PutUint32(db[16:20], in.Dirty)
	var strings [2][2]unsafe.Pointer
	for i := range strings {
		for j := range strings[i] {
			strings[i][j] = gameSettingsFixtureAlloc4161E0(32)
			if strings[i][j] == nil {
				t.Fatal("string calloc")
			}
			defer gameSettingsFixtureFree4161E0(strings[i][j])
			copy(unsafe.Slice((*byte)(strings[i][j]), 32), in.Strings[i][j])
		}
	}
	trace := make([][2]int64, 0, 24)
	var occurrences [14]int
	o := &gameSettingsNativeObserver4161E0{Region: unsafe.Pointer(&window[80]), Dirty: unsafe.Pointer(&db[16])}
	o.Word = func(id int, arg int32) (uint32, bool) {
		if id < 1 || id > 9 {
			panic("word service")
		}
		trace = append(trace, [2]int64{int64(id), int64(arg)})
		n := occurrences[id]
		occurrences[id]++
		if n > 1 {
			n = 1
		}
		return in.Words[id-1][n], true
	}
	o.Fill = func(id int, out []byte) bool {
		trace = append(trace, [2]int64{int64(id), 0})
		switch id {
		case 10:
			if len(out) != 20 {
				panic("spell extent")
			}
			for i, v := range in.Spells {
				binary.LittleEndian.PutUint32(out[4*i:], v)
			}
		case 11:
			b, e := hex.DecodeString(in.ItemsHex)
			if e != nil || len(b) != 4 || len(out) != 4 {
				panic("item extent")
			}
			copy(out, b)
		default:
			panic("fill service")
		}
		return true
	}
	o.String = func(id int) (unsafe.Pointer, bool) {
		if id != 12 && id != 13 {
			panic("string service")
		}
		trace = append(trace, [2]int64{int64(id), 0})
		n := occurrences[id]
		occurrences[id]++
		if n > 1 {
			n = 1
		}
		return strings[id-12][n], true
	}
	pop := pushGameSettingsNativeObserver4161E0(o)
	defer pop()
	word, pointer := gameSettingsCall4161E0(backend, in)
	after, e := hex.DecodeString(want.AfterHex)
	if e != nil || len(after) != 512 {
		t.Fatal("old-C AFTER512")
	}
	if !bytes.Equal(window, after) || binary.LittleEndian.Uint32(db[16:20]) != want.Dirty || !reflect.DeepEqual(trace, want.Trace) {
		t.Fatalf("old-C full-window/dirty/service order mismatch\ntrace %v\nwant %v", trace, want.Trace)
	}
	if in.ReturnKind == "word" && word != want.Word {
		t.Fatalf("full return bits %08x want%08x", word, want.Word)
	}
	if in.ReturnKind == "pointer" && pointer != unsafe.Add(o.Region, want.PointerOffset-371380) {
		t.Fatal("full native pointer return")
	}
	for _, v := range append(append([]byte{}, block[:64]...), block[576:]...) {
		if v != 0xbd {
			t.Fatal("region outer canary")
		}
	}
	for i, v := range db {
		if (i < 16 || i >= 20) && v != 0xce {
			t.Fatal("dirty-cell canary")
		}
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= 0xffffffff {
		t.Fatal("native high allocation witness")
	}
}
func TestGameSettingsNativeOldC4161E0(t *testing.T) {
	inputs, refs := gameSettingsLoad4161E0(t)
	for _, backend := range []string{"Go", "C"} {
		t.Run(backend, func(t *testing.T) {
			for i, in := range inputs.Cases {
				t.Run(in.Name, func(t *testing.T) { gameSettingsRunInput4161E0(t, backend, in, refs[i]) })
			}
		})
	}
}
func TestGameSettingsNativeDefaultOwners4161E0(t *testing.T) {
	if gameSettingsCurrent4161E0() != nil {
		t.Fatal("leaked observer")
	}
	window := unsafe.Slice((*byte)(memmap.PtrOff(0x5d4594, 371300)), 512)
	before := append([]byte{}, window...)
	cell := (*uint32)(gameSettingsActualDirtyCell4161E0())
	old := *cell
	t.Cleanup(func() { copy(window, before); *cell = old })
	if unsafe.Pointer(cell) == memmap.PtrOff(0x5d4594, 371692) {
		t.Fatal("C dirty owner replaced by memmap shadow")
	}
	for _, backend := range []string{"Go", "C"} {
		t.Run(backend, func(t *testing.T) {
			*cell = 0x80000001
			gameSettingsCall4161E0(backend, gameSettingsInput4161E0{Entry: 0x4164f0})
			if *cell != 0 {
				t.Fatal("actual C dirty owner clear")
			}
			_, p := gameSettingsCall4161E0(backend, gameSettingsInput4161E0{Entry: 0x4165d0, Args: [2]int32{1, 0}, ReturnKind: "pointer"})
			if p != memmap.PtrOff(0x5d4594, 371438) || *memmap.PtrUint32(0x5d4594, 371688) != 1 {
				t.Fatal("actual index58 and pointer")
			}
			gameSettingsPutDWORD4161E0(371700, 0x80000100)
			w, _ := gameSettingsCall4161E0(backend, gameSettingsInput4161E0{Entry: 0x416650})
			if w != 0x80000100 {
				t.Fatal("full DWORD getter")
			}
			gameSettingsPutDWORD4161E0(371704, 0xffffffff)
			w, _ = gameSettingsCall4161E0(backend, gameSettingsInput4161E0{Entry: 0x4169c0})
			if w != 0xffffffff {
				t.Fatal("acquired full DWORD")
			}
			gameSettingsPutByte4161E0(371616, 0xff)
			*cell = 0x89abcdef
			openBefore := append([]byte{}, window...)
			openWant := append([]byte{}, openBefore...)
			openWant[316] = 0xef
			_, openPointer := gameSettingsCall4161E0(backend, gameSettingsInput4161E0{Entry: 0x4169f0, ReturnKind: "pointer"})
			if openPointer != memmap.PtrOff(0x5d4594, 371516) || window[316] != 0xef || !bytes.Equal(window, openWant) || *cell != 0x89abcdef {
				t.Fatal("open clears only BYTE100, returns full same settings pointer and preserves actual C dirty owner")
			}
			snapshot := append([]byte{}, window...)
			*cell = 0x12345678
			gameSettingsCall4161E0(backend, gameSettingsInput4161E0{Entry: 0x416690})
			if !bytes.Equal(snapshot, window) || *cell != 0x12345678 {
				t.Fatal("current disabled690 no-op")
			}
		})
	}
}
func TestGameSettingsNativePinnedGoHeap4161E0(t *testing.T) {
	for _, backend := range []string{"Go", "C"} {
		t.Run(backend, func(t *testing.T) {
			block := new([640]byte)
			dirty := new([12]uint32)
			name := new([32]byte)
			for i := range block {
				block[i] = 0xbd
			}
			for i := 64; i < 576; i++ {
				block[i] = 0
			}
			var pins runtime.Pinner
			pins.Pin(block)
			pins.Pin(dirty)
			pins.Pin(name)
			defer pins.Unpin()
			gcs := 0
			gc := func() { runtime.GC(); runtime.GC(); gcs += 2 }
			o := &gameSettingsNativeObserver4161E0{Region: unsafe.Pointer(&block[144]), Dirty: unsafe.Pointer(&dirty[4])}
			o.Word = func(id int, arg int32) (uint32, bool) { gc(); return 0, true }
			o.Fill = func(id int, out []byte) bool { gc(); clear(out); return true }
			o.String = func(id int) (unsafe.Pointer, bool) { gc(); return unsafe.Pointer(&name[0]), true }
			pop := pushGameSettingsNativeObserver4161E0(o)
			defer pop()
			w, _ := gameSettingsCall4161E0(backend, gameSettingsInput4161E0{Entry: 0x4161e0})
			if w != 0 || dirty[4] != 0 || gcs == 0 {
				t.Fatal("pinned in-call GC state")
			}
			_, p := gameSettingsCall4161E0(backend, gameSettingsInput4161E0{Entry: 0x4165d0, Args: [2]int32{1, 0}, ReturnKind: "pointer"})
			if p != unsafe.Pointer(&block[202]) {
				t.Fatal("pinned full native interior return")
			}
			if binary.LittleEndian.Uint32(block[452:456]) != 1 {
				t.Fatal("pinned active index DWORD")
			}
			for i, v := range block {
				if (i < 64 || i >= 576) && v != 0xbd {
					t.Fatal("pinned canary")
				}
			}
			runtime.KeepAlive(block)
			runtime.KeepAlive(dirty)
			runtime.KeepAlive(name)
		})
	}
}
func TestGameSettingsNativeObserverLIFO4161E0(t *testing.T) {
	p := gameSettingsFixtureAlloc4161E0(512)
	d := gameSettingsFixtureAlloc4161E0(4)
	if p == nil || d == nil {
		t.Fatal("calloc")
	}
	defer gameSettingsFixtureFree4161E0(p)
	defer gameSettingsFixtureFree4161E0(d)
	outer := &gameSettingsNativeObserver4161E0{Region: unsafe.Add(p, 80), Dirty: d}
	pop := pushGameSettingsNativeObserver4161E0(outer)
	defer pop()
	func() {
		inner := &gameSettingsNativeObserver4161E0{Region: outer.Region, Dirty: d}
		undo := pushGameSettingsNativeObserver4161E0(inner)
		defer undo()
		if gameSettingsCurrent4161E0() != inner {
			t.Fatal("inner scope")
		}
		defer func() {
			if recover() == nil {
				t.Error("out-of-order pop must reject")
			}
		}()
		pop()
	}()
	if gameSettingsCurrent4161E0() != outer {
		t.Fatal("outer observer restored")
	}
	t.Log(fmt.Sprintf("settings12 C-char-signed=%t; controls are current-C, not original CPU", gameSettingsPlainCharSigned4161E0))
}
