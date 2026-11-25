package legacy

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
)

//go:embed testdata/inventory_lookup_original_v1.json.gz
var inventoryLookupNativeOriginal []byte

type inventoryLookupNativeCase struct {
	Input struct {
		Name   string
		Query  uint32
		Seed   int
		Counts [84]byte
		Edits  []struct {
			Cell, Word int
			Value      uint32
		}
		Ghost struct {
			Count, Word int
			Value       uint32
		}
		InitialPointer int    `json:"initial_pointer"`
		InitialIndex   uint32 `json:"initial_index"`
	}
	ReturnPointer      uint32 `json:"return_pointer"`
	Found              *struct{ Cell, Index int }
	ResultPointerAfter uint32 `json:"result_pointer_after"`
	ResultIndexAfter   uint32 `json:"result_index_after"`
	OrderedWrites      []struct {
		Address uint32
		Bytes   string
	} `json:"ordered_writes"`
	InitialGlobalSHA string `json:"initial_global_sha256"`
	FinalGlobalSHA   string `json:"final_global_sha256"`
	ForeignOwnerSHA  string `json:"foreign_owner_sha256"`
}

func inventoryLookupNativeCases(t *testing.T) []inventoryLookupNativeCase {
	t.Helper()
	var cases []inventoryLookupNativeCase
	npcVoiceOriginalDecode(t, inventoryLookupNativeOriginal, "ecbce9f9f44eb4f0770cd1849f5107d96b60431fbfe43cda0d58eeea6fec9abf", "d74f0773fd59ae8165184578d5b90412b9449f5b4cb3f7abfb1ca926e80bd703", 3647, &cases)
	return cases
}

func TestInventoryLookupNativeLayout(t *testing.T) {
	w := unsafe.Sizeof(uintptr(0))
	want := [17]uintptr{w, 4, 4, 1, w + 144, w, w + 136, w + 140, 1, 21, 4, 84, (w + 144) * 84, w, w, 4, 144}
	if got := cfixture.InventoryLookupNativeLayoutC(); got != want {
		t.Fatal("inventory scalar/pointer/callback ABI", got, want)
	}
	o := cfixture.InventoryLookupNativeOwnersC()
	if o[0] == o[1]+0x4b38 || o[1] != uintptr(memmap.PtrOff(0x5D4594, 0x6d0000-0x5D4594)) || o[2] != o[1]+0x4a50 || o[3] != o[2]+4 {
		t.Fatal("extracted grid and packed result owner identities", o)
	}
	for _, v := range o {
		if v == 0 || w == 8 && v <= math.MaxUint32 {
			t.Fatal("full native C owner required", o)
		}
	}
}

type inventoryLookupNativeState struct {
	grid, global []byte
	slot         *unsafe.Pointer
	word, cell   int
	frees        []func()
}

func newInventoryLookupNativeState(t *testing.T) *inventoryLookupNativeState {
	t.Helper()
	s := &inventoryLookupNativeState{grid: cfixture.InventoryLookupNativeGridC(), global: unsafe.Slice((*byte)(memmap.PtrOff(0x5D4594, 0x6d0000-0x5D4594)), 65536), slot: memmap.PtrPtr(0x5D4594, 1049788), word: int(unsafe.Sizeof(uintptr(0)))}
	s.cell = s.word + 144
	grid, global := append([]byte(nil), s.grid...), append([]byte(nil), s.global...)
	var pointer unsafe.Pointer
	if s.word == 8 {
		pointer = *s.slot
	}
	t.Cleanup(func() {
		copy(s.grid, grid)
		copy(s.global, global)
		if s.word == 8 {
			*s.slot = pointer
		}
		for _, free := range s.frees {
			free()
		}
	})
	return s
}

func (s *inventoryLookupNativeState) owner(t *testing.T) []byte {
	t.Helper()
	raw, free := alloc.MakeHigh([]byte{}, 65536+512)
	s.frees = append(s.frees, free)
	if s.word == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
		t.Fatal("foreign C owner above4GiB required")
	}
	return raw
}

func (s *inventoryLookupNativeState) put(pointer unsafe.Pointer) {
	if s.word == 4 {
		*(*uintptr)(unsafe.Pointer(s.slot)) = uintptr(pointer)
	} else {
		*s.slot = pointer
	}
}

func inventoryLookupNativeHash(t *testing.T, raw []byte, want string) {
	t.Helper()
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != want {
		t.Fatalf("original owner SHA: got%s want%s", got, want)
	}
}

// Encode declared input bytes; expected search results come only from the stock CPU fixture.
func inventoryLookupNativeInput(c inventoryLookupNativeCase) ([]byte, []byte) {
	g, f := make([]byte, 65536), make([]byte, 65536)
	for j := range g {
		g[j] = byte(j*53 + c.Input.Seed)
		f[j] = byte(j*37 + c.Input.Seed + 11)
	}
	for cell := 0; cell < 85; cell++ {
		at := 0x4b38 + 148*cell
		var ptr uint32
		if cell%3 != 0 {
			ptr = 0x21000100 + 64*uint32(cell)
		}
		binary.LittleEndian.PutUint32(g[at:], ptr)
		for word := 0; word < 32; word++ {
			binary.LittleEndian.PutUint32(g[at+4+4*word:], 0x40000000|uint32(cell<<16)|uint32(word<<8)|uint32(c.Input.Seed))
		}
		for i, off := range []int{132, 136, 144} {
			binary.LittleEndian.PutUint32(g[at+off:], uint32(0x50000000+i*0x10000000)|uint32(cell<<16)|uint32(c.Input.Seed))
		}
		copy(g[at+140:], []byte{0, byte(c.Input.Seed), byte(cell), 0xa0})
	}
	for cell, count := range c.Input.Counts {
		g[0x4b38+148*cell+140] = count
	}
	for _, e := range c.Input.Edits {
		binary.LittleEndian.PutUint32(g[0x4b38+148*e.Cell+4+4*e.Word:], e.Value)
	}
	at := 0x4b38 + 148*84
	g[at+140] = byte(c.Input.Ghost.Count)
	if c.Input.Ghost.Count != 0 {
		binary.LittleEndian.PutUint32(g[at+4+4*c.Input.Ghost.Word:], c.Input.Ghost.Value)
	}
	var prior uint32
	switch c.Input.InitialPointer {
	case -1:
	case -2:
		prior = 0x21000080
	default:
		prior = 0x6d4b38 + 148*uint32(c.Input.InitialPointer)
	}
	binary.LittleEndian.PutUint32(g[0x4a50:], prior)
	binary.LittleEndian.PutUint32(g[0x4a54:], c.Input.InitialIndex)
	return g, f
}

func (s *inventoryLookupNativeState) project(t *testing.T, raw []byte, placement int, ptr uint32) unsafe.Pointer {
	t.Helper()
	switch {
	case ptr == 0:
		return nil
	case ptr >= 0x6d4b38 && ptr < 0x6d4b38+84*148 && (ptr-0x6d4b38)%148 == 0:
		return unsafe.Pointer(&s.grid[int((ptr-0x6d4b38)/148)*s.cell])
	case ptr >= 0x21000000 && ptr < 0x21010000:
		return unsafe.Pointer(&raw[placement+int(ptr-0x21000000)])
	default:
		t.Fatalf("original pointer outside declared C owners: %#x", ptr)
		return nil
	}
}

func (s *inventoryLookupNativeState) initialize(t *testing.T, c inventoryLookupNativeCase, raw []byte, placement int) unsafe.Pointer {
	t.Helper()
	g, f := inventoryLookupNativeInput(c)
	inventoryLookupNativeHash(t, g, c.InitialGlobalSHA)
	inventoryLookupNativeHash(t, f, c.ForeignOwnerSHA)
	for j := range raw {
		raw[j] = byte(j*17 + c.Input.Seed + 3)
	}
	copy(raw[placement:], f)
	copy(s.global, g)
	cfixture.InventoryLookupNativeSeedC(unsafe.Pointer(&s.global[0x4b38]), unsafe.Pointer(&raw[placement]))
	wantGrid := make([]byte, len(s.grid))
	for cell := 0; cell < 84; cell++ {
		at := 0x4b38 + 148*cell
		p := s.project(t, raw, placement, binary.LittleEndian.Uint32(g[at:]))
		if s.word == 4 {
			binary.LittleEndian.PutUint32(wantGrid[cell*s.cell:], uint32(uintptr(p)))
		} else {
			binary.LittleEndian.PutUint64(wantGrid[cell*s.cell:], uint64(uintptr(p)))
		}
		copy(wantGrid[cell*s.cell+s.word:], g[at+4:at+148])
	}
	if !bytes.Equal(s.grid, wantGrid) {
		t.Fatal("complete native cell projection including scalar tail", c.Input.Name)
	}
	prior := s.project(t, raw, placement, binary.LittleEndian.Uint32(g[0x4a50:]))
	s.put(prior)
	// Validate the immutable CPU write ledger and its whole original owner hash.
	for _, w := range c.OrderedWrites {
		b, err := hex.DecodeString(w.Bytes)
		if err != nil || len(b) != 4 || w.Address < 0x6d0000 || w.Address >= 0x6e0000 {
			t.Fatal("original write ledger", c.Input.Name)
		}
		copy(g[int(w.Address-0x6d0000):], b)
	}
	inventoryLookupNativeHash(t, g, c.FinalGlobalSHA)
	if binary.LittleEndian.Uint32(g[0x4a50:]) != c.ResultPointerAfter || binary.LittleEndian.Uint32(g[0x4a54:]) != c.ResultIndexAfter {
		t.Fatal("original final pair/ledger", c.Input.Name)
	}
	if c.Found == nil {
		if len(c.OrderedWrites) != 0 || c.ReturnPointer != 0 {
			t.Fatal("original miss ledger", c.Input.Name)
		}
	} else if len(c.OrderedWrites) != 2 || c.OrderedWrites[0].Address != 0x6d4a54 || c.OrderedWrites[1].Address != 0x6d4a50 || c.ReturnPointer != 0x6d4a50 || c.ResultPointerAfter != 0x6d4b38+148*uint32(c.Found.Cell) || c.ResultIndexAfter != uint32(c.Found.Index) {
		t.Fatal("original index-before-pointer hit ledger", c.Input.Name)
	}
	return prior
}

func (s *inventoryLookupNativeState) verify(t *testing.T, c inventoryLookupNativeCase, pointer unsafe.Pointer, index uint32, raw, wantRaw, wantGrid, wantGlobal []byte) {
	t.Helper()
	if !bytes.Equal(raw, wantRaw) || !bytes.Equal(s.grid, wantGrid) || !bytes.Equal(s.global, wantGlobal) {
		t.Fatal("whole native grid/global/foreign owners or guards differ", c.Input.Name)
	}
	if got, gotIndex := cfixture.InventoryLookupNativeResultC(); got != pointer || gotIndex != index {
		t.Fatalf("canonical C result reader: got%p/%#x want%p/%#x (%s)", got, gotIndex, pointer, index, c.Input.Name)
	}
	if *s.slot != pointer || memmap.PtrPtr(0x5D4594, 1049788) != s.slot {
		t.Fatal("stable canonical result slot", c.Input.Name)
	}
}

func TestInventoryLookupNativeOriginal(t *testing.T) {
	cases := inventoryLookupNativeCases(t)
	s := newInventoryLookupNativeState(t)
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			raw := s.owner(t)
			for _, c := range cases {
				t.Run(fmt.Sprintf("route%d/place%d/%s", route, placement, c.Input.Name), func(t *testing.T) {
					prior := s.initialize(t, c, raw, placement)
					wantRaw, wantGrid, wantGlobal := append([]byte(nil), raw...), append([]byte(nil), s.grid...), append([]byte(nil), s.global...)
					pointer, index := prior, c.Input.InitialIndex
					var legacyReturn, cell unsafe.Pointer
					if c.Found != nil {
						cell = s.project(t, raw, placement, c.ResultPointerAfter)
						pointer, index = cell, c.ResultIndexAfter
						legacyReturn = unsafe.Pointer(&s.global[0x4a50])
						binary.LittleEndian.PutUint32(wantGlobal[0x4a54:], index)
						if s.word == 4 {
							binary.LittleEndian.PutUint32(wantGlobal[0x4a50:], uint32(uintptr(pointer)))
						}
					}
					if got := inventoryLookupNativeCall(route, 0, c.Input.Query, nil); got != legacyReturn {
						t.Fatalf("legacy result owner: got%p want%p", got, legacyReturn)
					}
					s.verify(t, c, pointer, index, raw, wantRaw, wantGrid, wantGlobal)
					out := 0x6000 + placement
					if c.Found != nil {
						binary.LittleEndian.PutUint32(wantRaw[out:], c.ResultIndexAfter)
					}
					if got := inventoryLookupNativeCall(route, 1, c.Input.Query, unsafe.Pointer(&raw[out])); got != cell {
						t.Fatalf("typed cell pointer: got%p want%p", got, cell)
					}
					s.verify(t, c, pointer, index, raw, wantRaw, wantGrid, wantGlobal)
					if got := inventoryLookupNativeCall(route, 1, c.Input.Query, nil); got != cell {
						t.Fatalf("typed cell pointer with NULL output: got%p want%p", got, cell)
					}
					s.verify(t, c, pointer, index, raw, wantRaw, wantGrid, wantGlobal)
				})
			}
		}
	}
}

func TestInventoryLookupNativeBoundedCParent(t *testing.T) {
	s := newInventoryLookupNativeState(t)
	for _, placement := range []int{0, 64, 512} {
		raw := s.owner(t)
		for _, c := range inventoryLookupNativeCases(t) {
			prior := s.initialize(t, c, raw, placement)
			wantRaw, wantGrid, wantGlobal := append([]byte(nil), raw...), append([]byte(nil), s.grid...), append([]byte(nil), s.global...)
			want := [2]uint32{0, 1}
			if c.Found != nil {
				input, _ := inventoryLookupNativeInput(c)
				want = [2]uint32{uint32(input[0x4b38+148*c.Found.Cell+140]), 0}
			}
			if got := cfixture.InventoryLookupNativeParentC(c.Input.Query); got != want {
				t.Fatal("unchanged bounded C parent count/NULL-drag service", c.Input.Name, got, want)
			}
			s.verify(t, c, prior, c.Input.InitialIndex, raw, wantRaw, wantGrid, wantGlobal)
		}
	}
}

func TestInventoryLookupNativeSequence(t *testing.T) {
	s := newInventoryLookupNativeState(t)
	raw := s.owner(t)
	var c inventoryLookupNativeCase
	for _, candidate := range inventoryLookupNativeCases(t) {
		if candidate.Input.Name == "hit-cell83-word31" {
			c = candidate
			break
		}
	}
	if c.Found == nil || c.Found.Cell != 83 || c.Found.Index != 31 {
		t.Fatal("fixed late-cell golden context missing")
	}
	s.initialize(t, c, raw, 64)
	if inventoryLookupNativeCall(1, 0, c.Input.Query, nil) != unsafe.Pointer(&s.global[0x4a50]) {
		t.Fatal("late row hit")
	}
	ptr, idx := cfixture.InventoryLookupNativeResultC()
	if ptr != unsafe.Pointer(&s.grid[83*s.cell]) || idx != 31 {
		t.Fatal("late-cell publication")
	}
	// Empty the input counts, retaining the prior published pair and every other byte.
	for cell := 0; cell < 84; cell++ {
		s.grid[cell*s.cell+s.word+136] = 0
	}
	wantRaw, wantGrid, wantGlobal := append([]byte(nil), raw...), append([]byte(nil), s.grid...), append([]byte(nil), s.global...)
	for route := 0; route < 3; route++ {
		if inventoryLookupNativeCall(route, 0, c.Input.Query, nil) != nil || inventoryLookupNativeCall(route, 1, c.Input.Query, unsafe.Pointer(&raw[0x6040])) != nil || inventoryLookupNativeCall(route, 1, c.Input.Query, nil) != nil {
			t.Fatal("empty successor search must miss")
		}
		s.verify(t, c, ptr, idx, raw, wantRaw, wantGrid, wantGlobal)
	}
}

func inventoryLookupNativeCall(route, entry int, code uint32, output unsafe.Pointer) unsafe.Pointer {
	if route != 2 {
		return cfixture.InventoryLookupNativeCall(route, entry, code, output)
	}
	if entry == 0 {
		return inventoryLookupLegacyResult(code)
	}
	return unsafe.Pointer(inventoryLookupFindCell(code, (*uint32)(output)))
}
