package legacy

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/music"
)

func g1sIsolate(t *testing.T) {
	t.Helper()
	var words [9]uint32
	for i := range words {
		words[i] = g1sWord(i, 0, false)
	}
	ticks, hooks, buffers := PlatformTicks, game1SplitTestHooks, game1TransferBuffers
	backdrop := g1sBackdrop()
	data := unsafe.Slice((*byte)(memmap.PtrOff(0x5D4594, 815220)), 1024)
	copyData := append([]byte(nil), data...)
	average := memmap.Uint64(0x587000, 91880)
	t.Cleanup(func() {
		for i, v := range words {
			g1sWord(i, v, true)
		}
		PlatformTicks, game1SplitTestHooks, game1TransferBuffers = ticks, hooks, buffers
		g1sSwapBackdrop(backdrop)
		copy(data, copyData)
		*memmap.PtrUint64(0x587000, 91880) = average
	})
}

func TestGame1SplitOriginalScalarCorpusCABI(t *testing.T) {
	g1sIsolate(t)
	raw, err := os.ReadFile("testdata/game1_split_scalar_original.json")
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != "8817f42d4bf16ddd609c598cc72f4fa6ccb42918a348570bb8806120c79da8d9" {
		t.Fatalf("sealed original fixture hash: %s", got)
	}
	var rows []struct {
		Kind               string
		Tick               uint32
		Previous           uint64
		Index, High, Count uint32
		Ring               []uint64
		Return             int64
		Delta              uint64
		NextIndex          uint64 `json:"next_index"`
		Average            uint64
		Fields             []uint32
		Case, Offset       int
		ReturnOffset       int `json:"return_offset"`
		Bytes              []byte
		Canary             byte
	}
	if err = json.Unmarshal(raw, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 588 {
		t.Fatalf("original cases: %d", len(rows))
	}
	formats, freeFormats := alloc.MakeHigh([]uint32{}, 5)
	defer freeFormats()
	lines := []string{"", "a", "abc\n", "abc\r", "abc\r\n", "\n\r\nX", "\r\n\rY", "a\rb\nc\r\nd", "한글\nend", "\r\n", "\n", "no final newline"}
	input, freeInput := alloc.MakeHigh([]byte{}, 256)
	defer freeInput()
	output, freeOutput := alloc.MakeHigh([]byte{}, 288)
	defer freeOutput()
	for index, row := range rows {
		t.Run(fmt.Sprintf("%s/%03d", row.Kind, index), func(t *testing.T) {
			switch row.Kind {
			case "timer":
				PlatformTicks = func() uint64 { return uint64(row.Tick) | uint64(0xdeadbeef)<<32 }
				*memmap.PtrUint64(0x5D4594, 815740) = row.Previous
				*memmap.PtrUint64(0x5D4594, 815756) = uint64(row.High)<<32 | uint64(row.Index)
				for i, v := range row.Ring {
					*memmap.PtrUint64(0x5D4594, 815220+uintptr(i)*8) = v
				}
				g1sWord(0, row.Count, true)
				if got := uint32(g1sCall(0, 0, nil)); got != uint32(row.Return) {
					t.Fatalf("counter return: %x != %x", got, row.Return)
				}
				if memmap.Uint64(0x5D4594, 815740) != uint64(row.Tick) || memmap.Uint64(0x5D4594, 815756) != row.NextIndex {
					t.Fatal("clock truncation or 64-bit index carry/modulo")
				}
				for i, v := range row.Ring {
					if uint32(i) == row.Index {
						v = row.Delta
					}
					if memmap.Uint64(0x5D4594, 815220+uintptr(i)*8) != v {
						t.Fatalf("ring slot %d", i)
					}
				}
				g1sCall(1, 0, nil)
				if got := memmap.Uint64(0x587000, 91880); got != row.Average {
					t.Fatalf("average: %x != %x", got, row.Average)
				}
			case "format":
				formats[1], formats[3], formats[4] = row.Fields[0], row.Fields[1], row.Fields[2]
				before := append([]uint32(nil), formats...)
				if got := int64(g1sCall(2, 0, unsafe.Pointer(&formats[0]))); got != row.Return {
					t.Fatalf("format %v: %d != %d", row.Fields, got, row.Return)
				}
				for i, v := range before {
					if formats[i] != v {
						t.Fatal("format record mutated")
					}
				}
			case "line":
				for i := range input {
					input[i] = 0xa7
				}
				for i := range output {
					output[i] = 0xa7
				}
				copy(input, lines[row.Case])
				input[len(lines[row.Case])] = 0
				before := append([]byte(nil), input...)
				p := g1sPointer(3, 0, unsafe.Pointer(&input[row.Offset]), unsafe.Pointer(&output[16]))
				if row.ReturnOffset == -1 {
					if p != nil {
						t.Fatal("expected EOF")
					}
				} else if p != unsafe.Pointer(&input[row.ReturnOffset]) {
					t.Fatalf("next native pointer: %p", p)
				}
				if !bytes.Equal(output[16:16+len(row.Bytes)], row.Bytes) || output[16+len(row.Bytes)] != 0 || output[17+len(row.Bytes)] != row.Canary {
					t.Fatal("line payload, NUL or canary")
				}
				if !bytes.Equal(input, before) || !bytes.Equal(output[:16], bytes.Repeat([]byte{0xa7}, 16)) {
					t.Fatal("input or prefix canary changed")
				}
			default:
				t.Fatal("unknown original record")
			}
		})
	}
}

func TestGame1SplitBackdropAndTransferNativeCABI(t *testing.T) {
	g1sIsolate(t)
	storage, free := alloc.MakeHigh([]byte{}, int(g1sLayout(0))+32)
	defer free()
	for i := range storage {
		storage[i] = 0xa7
	}
	p := unsafe.Pointer(&storage[16])
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
		t.Fatal("foreign pointer must be above 4 GiB")
	}
	before := append([]byte(nil), storage...)
	g1sWord(6, 0x13572468, true)
	var destroys int
	game1SplitTestHooks = &game1SplitHooks{
		NewBackdrop: func() *nox_window { return (*nox_window)(p) },
		DestroyWindow: func(w *nox_window) int32 {
			if unsafe.Pointer(w) != p {
				t.Fatalf("truncated window: %p", w)
			}
			destroys++
			return -7
		},
	}
	if g1sPointer(2, 1, nil, nil) != p || g1sBackdrop() != p {
		t.Fatal("native backdrop return/owner")
	}
	colorOffset := 16 + g1sLayout(1)
	*(*uint32)(unsafe.Pointer(&before[colorOffset])) = 0x13572468
	if !bytes.Equal(storage, before) {
		t.Fatal("native background field offset or canary changed")
	}
	if g1sPointer(2, 0, nil, nil) != nil || g1sBackdrop() != nil || destroys != 1 {
		t.Fatal("native destroy/clear")
	}
	g1sPointer(2, 0, nil, nil)
	if destroys != 1 {
		t.Fatal("double destroy")
	}
	game1SplitTestHooks.NewBackdrop = func() *nox_window { return nil }
	if g1sPointer(2, 1, nil, nil) != nil || g1sBackdrop() != nil {
		t.Fatal("failed creation must be safe")
	}
	lengths := [2]uint32{0x12345678, 0xfedcba98}
	var oldLengths [2]uint32
	for i, v := range lengths {
		oldLengths[i] = memmap.Uint32(0x5D4594, 826040+uintptr(i)*4)
		*memmap.PtrUint32(0x5D4594, 826040+uintptr(i)*4) = v
	}
	defer func() {
		for i, v := range oldLengths {
			*memmap.PtrUint32(0x5D4594, 826040+uintptr(i)*4) = v
		}
	}()
	g1sPointer(0, 0, p, nil)
	g1sPointer(0, 1, unsafe.Add(p, 16), nil)
	for _, i := range []int{-1, 2, 3, math.MaxInt32, math.MinInt32} {
		g1sPointer(0, i, nil, nil)
		if g1sPointer(1, i, nil, nil) != nil {
			t.Fatal("out-of-range transfer read")
		}
	}
	if g1sPointer(1, 0, nil, nil) != p || g1sPointer(1, 1, nil, nil) != unsafe.Add(p, 16) {
		t.Fatal("native transfer pointers")
	}
	for i, v := range lengths {
		if memmap.Uint32(0x5D4594, 826040+uintptr(i)*4) != v {
			t.Fatal("adjacent length overwritten")
		}
	}
}

func TestGame1SplitMusicStackCABI(t *testing.T) {
	g1sIsolate(t)
	module := MusicModule
	MusicModule = &music.Module{}
	t.Cleanup(func() { MusicModule = module })
	g1sWord(2, 0, true)
	g1sWord(3, 0, true)
	for level := uint32(0); level < 3; level++ {
		if g1sCall(9, 0, nil) != int32(level+1) || g1sWord(2, 0, false) != 0 || g1sWord(3, 0, false) != level+1 {
			t.Fatal("music group push")
		}
	}
	if g1sCall(9, 0, nil) != 3 {
		t.Fatal("music group overflow")
	}
	for level := uint32(3); level > 0; level-- {
		g1sCall(10, 0, nil)
		if g1sWord(3, 0, false) != level-1 {
			t.Fatal("music group pop")
		}
	}
	g1sCall(10, 0, nil)
	for i := 0; i < 6; i++ {
		if g1sCall(7, 0, nil) != 1 {
			t.Fatal("music frame push")
		}
	}
	if g1sCall(7, 0, nil) != 0 || g1sCall(6, 0, nil) != 6 {
		t.Fatal("music frame capacity")
	}
	for i := 0; i < 6; i++ {
		if got := g1sPointer(5, i, nil, nil); got != memmap.PtrOff(0x5D4594, 815772+16*uintptr(i)) {
			t.Fatal("music frame pointer")
		}
	}
	g1sCall(8, 0, nil)
	if g1sCall(6, 0, nil) != 0 {
		t.Fatal("music frame pop resets depth")
	}
	for _, v := range []uint32{0, 1, 0x7fffffff, 0x80000000, 0xffffffff} {
		if uint32(g1sCall(5, int(v), nil)) != v || uint32(g1sCall(6, 0, nil)) != v {
			t.Fatal("music DWORD ABI")
		}
		g1sWord(4, v, true)
		g1sWord(5, 42, true)
		if uint32(g1sCall(12, 0, nil)) != v {
			t.Fatal("audio availability return")
		}
		want := uint32(42)
		if v != 0 {
			want = 1
		}
		if g1sWord(5, 0, false) != want {
			t.Fatal("availability flag")
		}
		g1sCall(11, 0, nil)
		if g1sCall(13, 0, nil) != 0 {
			t.Fatal("availability clear")
		}
	}
}

func TestGame1SplitTextAndAddressNativeCABI(t *testing.T) {
	g1sIsolate(t)
	text, free := alloc.MakeHigh([]uint16{}, 64)
	defer free()
	game1SplitTestHooks = &game1SplitHooks{StringWidth: func(p *wchar2_t) int32 { return int32(len(GoWString(p)) * 7) }}
	for _, limit := range []int{0, 1, 4, 5, 12, 19, 26, 255} {
		for i := range text {
			text[i] = 0xa7a7
		}
		copy(text, []uint16{'a', 'b', 'c', 'd', 0})
		p := g1sPointer(4, limit, unsafe.Pointer(&text[0]), nil)
		if p != unsafe.Pointer(&text[0]) || text[5] != 0xa7a7 {
			t.Fatal("text pointer/canary")
		}
		got := GoWString((*wchar2_t)(p))
		want := ""
		switch limit {
		case 0, 1, 4, 5, 12:
			want = ""
		case 19:
			want = "a"
		case 26:
			want = "ab"
		case 255:
			want = "abcd"
		}
		if got != want {
			t.Fatalf("clip limit=%d got=%q want=%q", limit, got, want)
		}
	}
	buf, freeBuf := alloc.MakeHigh([]byte{}, 512)
	defer freeBuf()
	for _, addr := range []string{"127.0.0.1", "::1", strings.Repeat("x", 200)} {
		for _, port := range []uint16{0, 1, 65535} {
			for i := range buf {
				buf[i] = 0xa7
			}
			want := fmt.Sprintf("%s:%d", addr, port)
			if int(g1sFormat(addr, port, unsafe.Pointer(&buf[16]))) != len(want) || string(buf[16:16+len(want)]) != want || buf[16+len(want)] != 0 || buf[17+len(want)] != 0xa7 {
				t.Fatal("address format return/payload/canary")
			}
		}
	}
}

func TestGame1SplitOriginalTextCorpusCABI(t *testing.T) {
	g1sIsolate(t)
	raw, err := os.ReadFile("testdata/game1_split_text_original.json")
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != "646f4ed9ca81b95a64791ee43c929602eb20d49446ffcb90ef55640ae58df072" {
		t.Fatalf("sealed text fixture: %s", got)
	}
	var rows []struct {
		Kind                string
		Length, Limit, Case int
		Units               []uint16
	}
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 186 {
		t.Fatal("original text cases")
	}
	text, free := alloc.MakeHigh([]uint16{}, 160)
	defer free()
	game1SplitTestHooks = &game1SplitHooks{StringWidth: func(p *wchar2_t) int32 { return int32(len(GoWString(p)) * 7) }}
	suffix := unsafe.Slice((*byte)(memmap.PtrOff(0x587000, 103292)), 5)
	oldSuffix := append([]byte(nil), suffix...)
	defer copy(suffix, oldSuffix)
	copy(suffix, []byte(".wav\x00"))
	names := []string{"", "a", "abcdefghijklmnopqrstuvwxy", "\x80\xff"}
	name, freeName := alloc.MakeHigh([]byte{}, 128)
	defer freeName()
	for i, row := range rows {
		t.Run(fmt.Sprintf("%s/%03d", row.Kind, i), func(t *testing.T) {
			for j := range text {
				text[j] = 0xa7a7
			}
			p := unsafe.Pointer(&text[8])
			switch row.Kind {
			case "clip":
				for j := 0; j < row.Length; j++ {
					text[8+j] = uint16('a' + j)
				}
				text[8+row.Length] = 0
				if g1sPointer(4, row.Limit, p, nil) != p {
					t.Fatal("clip return identity")
				}
			case "suffix":
				for j := range name {
					name[j] = 0xa7
				}
				s := names[row.Case]
				copy(name, s)
				name[len(s)] = 0
				copy(text[8:], []uint16{'P', 'R', 'E', ':', 0})
				if g1sPointer(7, 0, unsafe.Pointer(&name[0]), p) != p {
					t.Fatal("suffix destination identity")
				}
			default:
				t.Fatal("unknown original case")
			}
			for j, v := range row.Units {
				if text[8+j] != v {
					t.Fatalf("unit[%d]=%x != %x", j, text[8+j], v)
				}
			}
			if text[8+len(row.Units)] != 0 || text[7] != 0xa7a7 {
				t.Fatal("text NUL/prefix canary")
			}
		})
	}
}
