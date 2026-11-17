package legacy

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"testing"
	"unicode/utf16"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

//go:embed testdata/wide_script_pointer_original_v1.json.gz
var wideScriptPointerNativeOriginal []byte

type wideScriptPointerNativeCase struct {
	Input struct {
		Name          string
		Count         uint32
		PointerRole   int `json:"pointer_role"`
		PointerOffset int `json:"pointer_offset"`
		Seed          int
	}
	ReturnWord      uint32 `json:"return_word"`
	CounterAfter    uint32 `json:"counter_after"`
	ArgumentPointer uint32 `json:"argument_pointer"`
	OrderedWrites   []struct {
		Address uint32
		Bytes   string
	} `json:"ordered_writes"`
}

func wideScriptPointerNativeCases(t *testing.T) []wideScriptPointerNativeCase {
	t.Helper()
	var cases []wideScriptPointerNativeCase
	npcVoiceOriginalDecode(t, wideScriptPointerNativeOriginal, "94797777584c082da5c2913a06703e298fc653a9f98974b495c3f6c4a85ca6d8", "f13b38d7a891c6cb6e2de8ef0442b55c0c81351d986cc41c05ce48d9c6c6e311", 1820, &cases)
	return cases
}

func TestWideScriptPointerNativeLayout(t *testing.T) {
	word := unsafe.Sizeof(uintptr(0))
	want := [10]uintptr{word, 4, 2, 4, word, 4, 2, 1024, 4, word}
	if got := wideScriptPointerNativeLayoutC(); got != want {
		t.Fatal("wide script scalar/pointer/callback ABI", got, want)
	}
	owners := wideScriptPointerNativeOwnersC()
	if owners[0] == owners[2] || owners[1] != uintptr(memmap.PtrOff(0x973F18, 0)) || owners[2] != uintptr(memmap.PtrOff(0x5D4594, 1599628)) {
		t.Fatal("extracted counter, packed blob and raw counter shadow identities", owners)
	}
	for _, value := range owners {
		if value == 0 || word == 8 && value <= math.MaxUint32 {
			t.Fatal("full native global owner required", owners)
		}
	}
}

type wideScriptPointerNativeState struct {
	blob, shadow []byte
	slots        [1024]*unsafe.Pointer
	word         int
	frees        []func()
}

func newWideScriptPointerNativeState(t *testing.T) *wideScriptPointerNativeState {
	t.Helper()
	s := &wideScriptPointerNativeState{
		blob:   memmap.Slice(0x973F18, 0),
		shadow: unsafe.Slice((*byte)(memmap.PtrOff(0x5D4594, 0x750000-0x5D4594)), 65536),
		word:   int(unsafe.Sizeof(uintptr(0))),
	}
	if len(s.blob) != 44881 {
		t.Fatal("packed wide script blob size", len(s.blob))
	}
	savedBlob, savedShadow := append([]byte(nil), s.blob...), append([]byte(nil), s.shadow...)
	savedCount := wideScriptPointerNativeCounterC(false, 0)
	var savedPointers [1024]unsafe.Pointer
	for i := range s.slots {
		s.slots[i] = memmap.PtrPtr(0x973F18, uintptr(26664+4*i))
		if s.word == 8 {
			savedPointers[i] = *s.slots[i]
		}
	}
	t.Cleanup(func() {
		copy(s.blob, savedBlob)
		copy(s.shadow, savedShadow)
		wideScriptPointerNativeCounterC(true, savedCount)
		if s.word == 8 {
			for i, value := range savedPointers {
				*s.slots[i] = value
			}
		}
		for _, free := range s.frees {
			free()
		}
	})
	return s
}

func (s *wideScriptPointerNativeState) owner(t *testing.T) []byte {
	t.Helper()
	raw, free := alloc.MakeHigh([]byte{}, 131072+512)
	s.frees = append(s.frees, free)
	if s.word == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
		t.Fatal("foreign pointee owners above4GiB required")
	}
	return raw
}

func (s *wideScriptPointerNativeState) put(i int, value unsafe.Pointer) {
	if s.word == 4 {
		*(*uintptr)(unsafe.Pointer(s.slots[i])) = uintptr(value)
	} else {
		*s.slots[i] = value
	}
}

func wideScriptPointerNativeProject(t *testing.T, raw []byte, placement int, address uint32) unsafe.Pointer {
	t.Helper()
	if address == 0 {
		return nil
	}
	if address < 0x21000000 || address >= 0x21020000 {
		t.Fatal("original pointer outside declared C pointee owners", address)
	}
	return unsafe.Pointer(&raw[placement+int(address-0x21000000)])
}

func (s *wideScriptPointerNativeState) initialize(t *testing.T, c wideScriptPointerNativeCase, raw []byte, placement int) [1024]unsafe.Pointer {
	t.Helper()
	for j := range raw {
		raw[j] = byte(j*17 + c.Input.Seed + 3)
	}
	for j := 0; j < 131072; j++ {
		raw[placement+j] = byte(j*37 + c.Input.Seed + 11)
	}
	for j := range s.blob {
		s.blob[j] = byte((j+0x3F18)*53 + c.Input.Seed)
	}
	for j := range s.shadow {
		s.shadow[j] = byte(j*31 + c.Input.Seed + 7)
	}
	var pointers [1024]unsafe.Pointer
	for j := range pointers {
		offset := uint32(2 * ((j * 19) % 32768))
		original := [4]uint32{0, 0x21000000, 0x21000000 + offset, 0x21010000 + offset}[j%4]
		pointers[j] = wideScriptPointerNativeProject(t, raw, placement, original)
		word := original
		if s.word == 4 {
			word = uint32(uintptr(pointers[j]))
		}
		binary.LittleEndian.PutUint32(s.blob[26664+4*j:], word)
		s.put(j, pointers[j])
	}
	wideScriptPointerNativeCounterC(true, c.Input.Count)
	return pointers
}

func wideScriptPointerNativeObservedStore(t *testing.T, c wideScriptPointerNativeCase, raw []byte, placement int) (int, unsafe.Pointer) {
	t.Helper()
	if len(c.OrderedWrites) == 0 {
		return -1, nil
	}
	if len(c.OrderedWrites) != 2 || c.OrderedWrites[1].Address != 0x75AE20 {
		t.Fatal("original table-then-counter write ledger", c.Input.Name)
	}
	w := c.OrderedWrites[0]
	if w.Address < 0x97A740 || w.Address >= 0x97B740 || (w.Address-0x97A740)%4 != 0 {
		t.Fatal("original logical pointer cell", c.Input.Name)
	}
	b, err := hex.DecodeString(w.Bytes)
	if err != nil || len(b) != 4 {
		t.Fatal("original pointer store bytes", c.Input.Name)
	}
	return int((w.Address - 0x97A740) / 4), wideScriptPointerNativeProject(t, raw, placement, binary.LittleEndian.Uint32(b))
}

func (s *wideScriptPointerNativeState) verify(t *testing.T, c wideScriptPointerNativeCase, pointers [1024]unsafe.Pointer, raw, wantRaw, wantBlob, wantShadow []byte) {
	t.Helper()
	if !bytes.Equal(raw, wantRaw) || !bytes.Equal(s.blob, wantBlob) || !bytes.Equal(s.shadow, wantShadow) || wideScriptPointerNativeCounterC(false, 0) != c.CounterAfter {
		t.Fatal("whole packed blob/pointee owners/counter shadow or extracted counter differs", c.Input.Name)
	}
	for i, want := range pointers {
		if got := *s.slots[i]; got != want {
			t.Fatalf("canonical cell%d: got%p want%p (%s)", i, got, want, c.Input.Name)
		}
	}
}

func TestWideScriptPointerNativeOriginal(t *testing.T) {
	cases := wideScriptPointerNativeCases(t)
	s := newWideScriptPointerNativeState(t)
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			raw := s.owner(t)
			for _, c := range cases {
				t.Run(fmt.Sprintf("route%d/place%d/%s", route, placement, c.Input.Name), func(t *testing.T) {
					pointers := s.initialize(t, c, raw, placement)
					wantRaw, wantBlob, wantShadow := append([]byte(nil), raw...), append([]byte(nil), s.blob...), append([]byte(nil), s.shadow...)
					cell, pointer := wideScriptPointerNativeObservedStore(t, c, raw, placement)
					if cell >= 0 {
						pointers[cell] = pointer
						if s.word == 4 {
							binary.LittleEndian.PutUint32(wantBlob[26664+4*cell:], uint32(uintptr(pointer)))
						}
					}
					argument := wideScriptPointerNativeProject(t, raw, placement, c.ArgumentPointer)
					if got := wideScriptPointerNativeCall(route, argument); got != c.ReturnWord {
						t.Fatalf("original signed32 return: got%#x want%#x", got, c.ReturnWord)
					}
					s.verify(t, c, pointers, raw, wantRaw, wantBlob, wantShadow)
					if cell < 0 {
						cell = 0
					}
					if wideScriptPointerNativeReadC(cell) != uintptr(pointers[cell]) || memmap.PtrPtr(0x973F18, uintptr(26664+4*cell)) != s.slots[cell] {
						t.Fatal("existing C canonical pointer reader or stable cell identity", c.Input.Name)
					}
				})
			}
		}
	}
}

func TestWideScriptPointerNativeWrapper(t *testing.T) {
	cases := wideScriptPointerNativeCases(t)
	byCount := make(map[uint32]wideScriptPointerNativeCase)
	for _, c := range cases {
		if _, ok := byCount[c.Input.Count]; !ok {
			byCount[c.Input.Count] = c
		}
	}
	s := newWideScriptPointerNativeState(t)
	raw := s.owner(t)
	pointers := s.initialize(t, byCount[0], raw, 64)
	wantRaw, wantBlob, wantShadow := append([]byte(nil), raw...), append([]byte(nil), s.blob...), append([]byte(nil), s.shadow...)
	for i, text := range []string{"", "CScrFunc.c", "환경음과 상점", "Nox 🧙", "a\x00b", "0123456789abcdefghijklmnopqrstuvwxyz"} {
		c := byCount[uint32(i)]
		if got := Sub_512E80(text); uint32(got) != c.ReturnWord {
			t.Fatal("existing Go wrapper signed return", i, got, c.ReturnWord)
		}
		pointer := *s.slots[i]
		if pointer == nil || s.word == 8 && uintptr(pointer) <= math.MaxUint32 {
			t.Fatal("existing C UTF16 producer full native allocation", i, pointer)
		}
		s.frees = append(s.frees, func() { alloc.FreePtr(pointer) })
		wantText := append(utf16.Encode([]rune(text)), 0)
		gotText := unsafe.Slice((*uint16)(pointer), len(wantText))
		for j, value := range wantText {
			if gotText[j] != value {
				t.Fatal("existing UTF16 producer bytes or terminator", i, j, gotText, wantText)
			}
		}
		pointers[i] = pointer
		if s.word == 4 {
			binary.LittleEndian.PutUint32(wantBlob[26664+4*i:], uint32(uintptr(pointer)))
		}
		s.verify(t, c, pointers, raw, wantRaw, wantBlob, wantShadow)
		if wideScriptPointerNativeReadC(i) != uintptr(pointer) {
			t.Fatal("existing C canonical reader of Go-produced UTF16", i)
		}
	}
}

func TestWideScriptPointerNativeCounterReset(t *testing.T) {
	cases := wideScriptPointerNativeCases(t)
	find := func(count, pointer uint32) wideScriptPointerNativeCase {
		for _, c := range cases {
			if c.Input.Count == count && c.ArgumentPointer == pointer {
				return c
			}
		}
		t.Fatalf("original sequence context absent: count%#x pointer%#x", count, pointer)
		return wideScriptPointerNativeCase{}
	}
	trace := []wideScriptPointerNativeCase{
		find(1023, 0x21000000), find(1024, 0),
		find(0, 0x21010000), find(1, 0), find(2, 0x21010000), find(3, 0x21010000),
	}
	s := newWideScriptPointerNativeState(t)
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			raw := s.owner(t)
			pointers := s.initialize(t, trace[0], raw, placement)
			wantRaw, wantBlob, wantShadow := append([]byte(nil), raw...), append([]byte(nil), s.blob...), append([]byte(nil), s.shadow...)
			for step, c := range trace {
				if step == 2 {
					// This is the existing setter used by noxScript.Reset. Reset
					// retains the table; it does not reset all memmap side slots.
					Set_dword_5d4594_1599628(0)
					s.verify(t, wideScriptPointerNativeCase{CounterAfter: 0}, pointers, raw, wantRaw, wantBlob, wantShadow)
				}
				if wideScriptPointerNativeCounterC(false, 0) != c.Input.Count {
					t.Fatal("original sequence counter before entry", route, placement, step)
				}
				cell, pointer := wideScriptPointerNativeObservedStore(t, c, raw, placement)
				if cell >= 0 {
					pointers[cell] = pointer
					if s.word == 4 {
						binary.LittleEndian.PutUint32(wantBlob[26664+4*cell:], uint32(uintptr(pointer)))
					}
				}
				argument := wideScriptPointerNativeProject(t, raw, placement, c.ArgumentPointer)
				if got := wideScriptPointerNativeCall(route, argument); got != c.ReturnWord {
					t.Fatal("original composed insert/saturation/reset/NULL/reinsertion result", route, placement, step, got, c.ReturnWord)
				}
				s.verify(t, c, pointers, raw, wantRaw, wantBlob, wantShadow)
			}
		}
	}
}
