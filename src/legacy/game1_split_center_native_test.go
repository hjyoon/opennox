package legacy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// Reuse the binary rows emitted from the sealed GAME.EXE fixture by the C
// reference test. Exercise the live public C -> Go entries against every
// complete event, ring byte, shadow byte, return value and boundary guard.
func g1sCompareCenterNative(t *testing.T, mode string, encoded []byte) {
	t.Helper()
	g1sIsolate(t)
	in := bytes.NewReader(encoded)
	var offsets [8]int32
	var count uint32
	read := func(p any) {
		if err := binary.Read(in, binary.LittleEndian, p); err != nil {
			t.Fatal(err)
		}
	}
	read(&offsets)
	read(&count)
	if count != 297 {
		t.Fatalf("original HUD row count: %d", count)
	}
	oldWidth := g1sSwapWidth(0)
	defer g1sSwapWidth(oldWidth)
	g1sWord(6, 0, true)
	g1sWord(7, 0x7fff, true)
	mainOwners, shadowOwners := make([][]byte, 3), make([][]byte, 3)
	for i := range mainOwners {
		var free func()
		mainOwners[i], free = alloc.MakeHigh([]byte{}, 0x80000)
		defer free()
		shadowOwners[i], free = alloc.MakeHigh([]byte{}, 0x80000)
		defer free()
		if unsafe.Sizeof(uintptr(0)) == 8 &&
			(uintptr(unsafe.Pointer(&mainOwners[i][0])) <= math.MaxUint32 || uintptr(unsafe.Pointer(&shadowOwners[i][0])) <= math.MaxUint32) {
			t.Fatal("live HUD foreign owners must be above 4 GiB")
		}
	}
	sources := [][]uint16{{0}, {0x41, 0}, {0xac00, 0xd83d, 0xde80, 0x42, 0}}
	clearOffsets := [...]int{16, 1280, 1920, 2048, 4096}
	shadowOffsets := [...]int{8, 16, 24, 32, 128}
	var nativeCases int
	for row := uint32(0); row < count; row++ {
		var fields [10]int64
		var eventCount uint32
		var expected [32][5]int64
		var returned uint32
		read(&fields)
		read(&eventCount)
		read(&expected)
		read(&returned)
		wantMode := int64(0)
		if mode == "draw" {
			wantMode = 1
		}
		if fields[0] != wantMode {
			continue
		}
		if eventCount > 32 {
			t.Fatal("original HUD event count")
		}
		for mapping := 0; mapping < 3; mapping++ {
			for placement := range clearOffsets {
				for pattern, fill := range []byte{0, 0x5a, 0xff} {
					main, shadow := mainOwners[mapping], shadowOwners[mapping]
					at := 0x40000 - clearOffsets[placement] - 636
					sh := 0x40000 - shadowOffsets[placement]
					if mapping == 2 {
						at = 64 + clearOffsets[placement]
						sh = 64 + shadowOffsets[placement]
					}
					for i := at - 32; i < at+1964; i++ {
						main[i] = fill
					}
					for i := sh - 32; i < sh+64; i++ {
						shadow[i] = fill
					}
					for i, v := range offsets {
						binary.LittleEndian.PutUint32(shadow[sh+4*i:], uint32(v))
					}
					if wantMode == 0 {
						for i, v := range sources[fields[9]] {
							binary.LittleEndian.PutUint16(main[at+1936+2*i:], v)
						}
					} else {
						for i := 0; i < 3; i++ {
							binary.LittleEndian.PutUint16(main[at+644*i:], uint16('A'+i))
							binary.LittleEndian.PutUint16(main[at+644*i+2:], 0)
							expires := uint32(fields[3]) - 1
							if fields[2]&(1<<i) != 0 {
								expires = uint32(fields[3])
							}
							binary.LittleEndian.PutUint32(main[at+644*i+636:], expires)
						}
					}
					want := append([]byte(nil), main[at-32:at+1964]...)
					shadowBefore := append([]byte(nil), shadow[sh-32:sh+64]...)
					var events [][5]int64
					emit := func(id, a, b, c int64) { events = append(events, [5]int64{id, a, b, c, 0}) }
					kind := func(p *wchar2_t) int64 {
						base, ptr := uintptr(unsafe.Pointer(&main[at])), uintptr(unsafe.Pointer(p))
						if ptr < base || ptr-base > 1288 || (ptr-base)%644 != 0 {
							t.Fatalf("wrong native HUD text pointer: %p", p)
						}
						return int64((ptr - base) / 644)
					}
					older := int32(0x3e0)
					game1SplitTestHooks = &game1SplitHooks{
						CenterMemory: func(base, off uintptr) unsafe.Pointer {
							switch {
							case base == 0x5d4594 && off == 2597996:
								return unsafe.Pointer(&older)
							case base == 0x5d4594 && off >= 823804 && off <= 825772:
								return unsafe.Pointer(&main[at+int(off-823804)])
							case base == 0x587000 && off >= 107848 && off <= 107880:
								return unsafe.Pointer(&shadow[sh+int(off-107848)])
							default:
								t.Fatalf("unexpected live HUD memory %#x:%d", base, off)
								return nil
							}
						},
						CopyText: func(dst, src *wchar2_t) *wchar2_t {
							if unsafe.Pointer(src) != unsafe.Pointer(&main[at+1936]) {
								t.Fatal("wrong native HUD copy source")
							}
							emit(6, kind(dst), 0, 0)
							for i, v := range sources[fields[9]] {
								*(*uint16)(unsafe.Add(unsafe.Pointer(dst), 2*i)) = v
							}
							return dst
						},
						Viewport:  func() (int32, int32) { emit(1, 0, 0, 0); return int32(fields[5]), int32(fields[6]) },
						TextColor: func(color uint32) { emit(2, int64(color), 0, 0) },
						StringWidth: func(text *wchar2_t) int32 {
							index := kind(text)
							emit(3, index, 0, 0)
							return int32(fields[7]) + 8*int32(index)
						},
						DrawText:   func(text *wchar2_t, x, y int32) int32 { emit(4, kind(text), int64(x), int64(y)); return -559038737 },
						FontHeight: func() int32 { emit(5, 0, 0, 0); return int32(fields[8]) },
						Frame:      func() uint32 { return uint32(fields[3]) },
					}
					g1sWord(8, uint32(fields[1]), true)
					g1sSwapWidth(int32(fields[4]))
					if wantMode == 0 {
						for i := 0; i < 3; i++ {
							pos := 32 + 644*i
							for j, v := range sources[fields[9]] {
								binary.LittleEndian.PutUint16(want[pos+2*j:], v)
							}
							clear(want[pos+636 : pos+641])
						}
						p := g1sPointer(6, 0, nil, nil)
						if p != unsafe.Pointer(&main[at+644*int(returned)]) || g1sWord(8, 0, false) != 0 {
							t.Fatal("native HUD clear return/index")
						}
					} else {
						if uint32(g1sCall(15, 0, nil)) != returned || g1sWord(8, 0, false) != uint32(fields[1]) {
							t.Fatal("native HUD draw return/index")
						}
					}
					if len(events) != int(eventCount) {
						t.Fatalf("%s row=%d mapping=%d placement=%d pattern=%d event count %d != %d", mode, row, mapping, placement, pattern, len(events), eventCount)
					}
					for i, event := range events {
						if event != expected[i] {
							t.Fatalf("%s row=%d event[%d]=%v != %v", mode, row, i, event, expected[i])
						}
					}
					if !bytes.Equal(main[at-32:at+1964], want) || !bytes.Equal(shadow[sh-32:sh+64], shadowBefore) {
						t.Fatal("live HUD packed ring/shadow bytes or guards changed")
					}
					nativeCases++
				}
			}
		}
	}
	if in.Len() != 0 {
		t.Fatal("trailing original HUD data")
	}
	t.Log(fmt.Sprintf("live public C -> Go HUD %s: %d native cases, complete original traces, packed bytes and guards", mode, nativeCases))
}
