package legacy

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"math"
	"os"
	"reflect"
	"runtime"
	"testing"
	"unsafe"
)

func g2qTestIsolate(t *testing.T) {
	t.Helper()
	array := unsafe.Slice((*byte)(g2qTestArray()), 8*int(g2qTestLayout(0)))
	before := append([]byte(nil), array...)
	clear(array)
	var ptrs [11]unsafe.Pointer
	for i := range ptrs {
		ptrs[i] = g2qTestPointer(i, nil, false)
		g2qTestPointer(i, nil, true)
	}
	var words [4]uint32
	for i := range words {
		words[i] = g2qTestWord(i, 0, false)
		g2qTestWord(i, 0, true)
	}
	packed := memmap.Slice(0x5D4594, 1047764)[:2000]
	saved := append([]byte(nil), packed...)
	clear(packed)
	hooks := game2QuickbarTestHooks
	game2QuickbarTestHooks = &game2QuickbarHooks{Sound: func(int32, int32) {}, Destroy: func(unsafe.Pointer) int32 { return 0 }, Hidden: func(unsafe.Pointer, int32) int32 { return 0 }, Style: func(unsafe.Pointer, int32) int32 { return 0 }, Tooltip: func(unsafe.Pointer, uint32) {}}
	t.Cleanup(func() {
		game2QuickbarTestHooks = hooks
		copy(array, before)
		for i, v := range ptrs {
			g2qTestPointer(i, v, true)
		}
		for i, v := range words {
			g2qTestWord(i, v, true)
		}
		copy(packed, saved)
	})
}
func g2qTestObjects(t *testing.T, count int) unsafe.Pointer {
	t.Helper()
	b, free := alloc.MakeHigh([]byte{}, count*int(g2qTestLayout(6)))
	t.Cleanup(free)
	return unsafe.Pointer(&b[0])
}

func TestGame2QuickbarNativeRegistryABIAndGC(t *testing.T) {
	g2qTestIsolate(t)
	type owner struct {
		Data, Root                 unsafe.Pointer
		Buttons, Nuggets, Children [5]unsafe.Pointer
	}
	want := []uintptr{unsafe.Sizeof(owner{}), unsafe.Offsetof(owner{}.Data), unsafe.Offsetof(owner{}.Root), unsafe.Offsetof(owner{}.Buttons), unsafe.Offsetof(owner{}.Nuggets), unsafe.Offsetof(owner{}.Children)}
	for i, v := range want {
		if g2qTestLayout(i) != v {
			t.Fatalf("C/Go native owner field%d", i)
		}
	}
	data := make([]byte, 8*256)
	var pin runtime.Pinner
	pin.Pin(&data[0])
	defer func() {
		clear(unsafe.Slice((*byte)(g2qTestArray()), 8*int(g2qTestLayout(0))))
		pin.Unpin()
	}()
	windows := g2qTestObjects(t, 8*16)
	stride := g2qTestLayout(6)
	for i := 0; i < 8; i++ {
		p := unsafe.Pointer(&data[i*256])
		w := unsafe.Add(windows, uintptr(i*16)*stride)
		g2qTestOwner(i, p, w, unsafe.Add(w, stride), unsafe.Add(w, 6*stride), unsafe.Add(w, 11*stride))
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(p) <= math.MaxUint32 || uintptr(w) <= math.MaxUint32) {
			t.Fatal("full native owners must exceed4GiB")
		}
		runtime.GC()
		if g2qTestLookup(0, p, 0) != w {
			t.Fatal("root lookup C entry")
		}
		for j := 0; j < 5; j++ {
			for kind, start := range []int{1, 6, 11} {
				actual := g2qTestLookup(kind+1, p, j)
				expected := unsafe.Add(w, uintptr(start+j)*stride)
				if actual != expected || g2qTestLookup(4, expected, 0) != p {
					t.Fatal("full pointer child registry")
				}
			}
		}
	}
	extra := new([256]byte)
	pin.Pin(extra)
	g2qTestSet(false, unsafe.Pointer(extra), 0, windows)
	if g2qTestLookup(2, unsafe.Pointer(extra), 0) != nil {
		t.Fatal("ninth registry key must not replace existing owner")
	}
	for _, index := range []int{-1, 5, math.MaxInt32} {
		if g2qTestLookup(1, unsafe.Pointer(&data[0]), index) != nil {
			t.Fatal("slot guard")
		}
	}
	g2qTestCall(0, 0, unsafe.Pointer(&data[0]))
	g2qTestSet(true, unsafe.Pointer(extra), 4, windows)
	if g2qTestLookup(3, unsafe.Pointer(extra), 4) != windows {
		t.Fatal("cleared registry slot reuse")
	}
	// Release every retained Go pointer before unpinning, including callback keys.
	clear(unsafe.Slice((*byte)(g2qTestArray()), 8*int(g2qTestLayout(0))))
	runtime.KeepAlive(data)
	runtime.KeepAlive(extra)
}

func TestGame2QuickbarNativePackedRowsAndExpandedRoundtrip(t *testing.T) {
	g2qTestIsolate(t)
	storage, free := alloc.MakeHigh([]byte{}, 288)
	defer free()
	for i := range storage {
		storage[i] = byte(i*29 + 7)
	}
	base := unsafe.Pointer(&storage[16])
	g2qTestPointer(0, base, true)
	g2qTestPointer(1, nil, true)
	expanded := memmap.Slice(0x5D4594, 1048196)[:1280]
	for i := range expanded {
		expanded[i] = 0xa7
	}
	original := append([]byte(nil), storage...)
	*(*byte)(unsafe.Add(base, 200)) = 2
	original[16+200] = 2
	oldPause := pauseFXStateNative57AF30.active
	pauseFXStateNative57AF30.active = false
	defer func() { pauseFXStateNative57AF30.active = oldPause }()
	var sounds []int32
	hidden := 0
	game2QuickbarTestHooks.Sound = func(id, volume int32) {
		if volume != 100 {
			t.Fatal("volume")
		}
		sounds = append(sounds, id)
	}
	game2QuickbarTestHooks.Hidden = func(unsafe.Pointer, int32) int32 { hidden++; return 0 }
	g2qTestCall(5, 0, nil)
	if memmap.Uint32(0x5D4594, 1049476) != 1 || memmap.Uint32(0x5D4594, 1047912) != 2 || *(*byte)(unsafe.Add(base, 200)) != 4 {
		t.Fatal("expanded opening row state")
	}
	for row := 0; row < 5; row++ {
		for slot := 0; slot < 5; slot++ {
			src := 16 + 40*row + 8*slot
			dst := 256*row + 8*slot
			if !bytes.Equal(expanded[dst:dst+5], original[src:src+5]) {
				t.Fatal("copied DWORD+flag")
			}
			if !bytes.Equal(expanded[dst+5:dst+8], []byte{0xa7, 0xa7, 0xa7}) {
				t.Fatal("padding copied unexpectedly")
			}
		}
	}
	// Edits are restricted to the serialized DWORD and flag; spare bytes remain.
	for row := 0; row < 5; row++ {
		for slot := 0; slot < 5; slot++ {
			o := 256*row + 8*slot
			expanded[o] = byte(40 + row*5 + slot)
			expanded[o+4] = byte(0x80 | slot)
		}
	}
	if g2qTestCall(4, 0, nil) != 1 {
		t.Fatal("close result")
	}
	if memmap.Uint32(0x5D4594, 1049476) != 0 || *(*byte)(unsafe.Add(base, 200)) != 2 {
		t.Fatal("restored selected row")
	}
	for row := 0; row < 5; row++ {
		for slot := 0; slot < 5; slot++ {
			o := 16 + 40*row + 8*slot
			if !bytes.Equal(storage[o:o+5], expanded[256*row+8*slot:256*row+8*slot+5]) || !bytes.Equal(storage[o+5:o+8], original[o+5:o+8]) {
				t.Fatal("roundtrip fields/padding")
			}
		}
	}
	if unsafe.Sizeof(uintptr(0)) == 4 {
		// The PE32 row-pointer mirror is a real output at offset 204. On a
		// native 64-bit host these same bytes remain guarded scalar padding.
		binary.LittleEndian.PutUint32(original[16+204:], uint32(uintptr(unsafe.Add(base, 2*40))))
	}
	if !bytes.Equal(storage[:16], original[:16]) || !bytes.Equal(storage[16+201:], original[16+201:]) {
		t.Fatal("packed row outer canary/64bit pointer mirror overwritten")
	}
	if fmt.Sprint(sounds) != "[799 800]" || hidden != 62 {
		t.Fatalf("service sequence %v hidden%d", sounds, hidden)
	}
}

func TestGame2QuickbarNativeSelectorsAndSwap(t *testing.T) {
	g2qTestIsolate(t)
	data, free := alloc.MakeHigh([]byte{}, 256)
	defer free()
	for i := range data {
		data[i] = byte(i*13 + 3)
	}
	base := unsafe.Pointer(&data[0])
	g2qTestPointer(0, base, true)
	for _, v := range []int{-1, 0, 1, 2, 3, 255} {
		before := *memmap.PtrUint8(0x5D4594, 1048140)
		want := int32(v)
		if v >= 0 && v < 3 {
			want = int32(v * 5)
		}
		if g2qTestCall(6, v, nil) != want {
			t.Fatal("trap selector result")
		}
		if v < 0 || v >= 3 {
			if *memmap.PtrUint8(0x5D4594, 1048140) != before {
				t.Fatal("invalid trap mutation")
			}
		}
	}
	*memmap.PtrUint8(0x5D4594, 1048140) = 2
	g2qTestCall(7, 0, nil)
	if g2qTestCall(9, 0, nil) != 0 {
		t.Fatal("trap next wrap")
	}
	g2qTestCall(8, 0, nil)
	if g2qTestCall(9, 0, nil) != 2 {
		t.Fatal("trap prev wrap")
	}
	*memmap.PtrUint32(0x5D4594, 1049476) = 0
	g2qTestWord(0, 0, true)
	if g2qTestCall(2, 3, nil) != 0 || g2qTestCall(3, 0, nil) != 3 {
		t.Fatal("main row selector")
	}
	before := append([]byte(nil), data...)
	g2qTestSwap(base, 0, 4)
	want := append([]byte(nil), before...)
	copy(want[120:128], before[152:160])
	copy(want[152:160], before[120:128])
	if !bytes.Equal(data, want) {
		t.Fatal("swap full8 scalar bytes only")
	}
	for _, invalid := range []int{-1, 5} {
		g2qTestSwap(base, invalid, 0)
		if !bytes.Equal(data, want) {
			t.Fatal("invalid swap changed bytes")
		}
	}
	if g2qTestCall(10, 0, nil) != 1 || g2qTestCall(11, 0, nil) != 0 {
		t.Fatal("draw stubs")
	}
}

func TestGame2QuickbarNativeDestroyOrder(t *testing.T) {
	g2qTestIsolate(t)
	data, free := alloc.MakeHigh([]byte{}, 256)
	defer free()
	objects := g2qTestObjects(t, 16)
	s := g2qTestLayout(6)
	g2qTestOwner(0, unsafe.Pointer(&data[0]), objects, unsafe.Add(objects, s), unsafe.Add(objects, 6*s), unsafe.Add(objects, 11*s))
	var calls []unsafe.Pointer
	game2QuickbarTestHooks.Destroy = func(p unsafe.Pointer) int32 { calls = append(calls, p); return 0 }
	g2qTestCall(1, 0, nil)
	if len(calls) != 128 {
		t.Fatalf("all8sets*16 actual destroy calls: %d", len(calls))
	}
	if calls[0] != objects {
		t.Fatal("root first")
	}
	for i := 0; i < 5; i++ {
		for j, start := range []int{1, 6, 11} {
			if calls[1+3*i+j] != unsafe.Add(objects, uintptr(start+i)*s) {
				t.Fatal("button/nugget/child destroy order")
			}
		}
	}
	for _, b := range unsafe.Slice((*byte)(g2qTestArray()), 8*int(g2qTestLayout(0))) {
		if b != 0 {
			t.Fatal("registry reset")
		}
	}
}

// This consumes the frozen actual PE32 handler rows unchanged. Hooked callees
// are the original explicitly declared services; this qualifies this handler's
// branches, memory writes and call ordering only, not those callee bodies.
func TestGame2Quickbar45EF50ActualGoFrozenOriginal245(t *testing.T) {
	raw, err := os.ReadFile("testdata/quickbar_drag_out_45ef50.json")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "2ab9d21d12fa7c8ddf4ad3f87742b25b498b10f9e86554535c0d26c2baca1c73" {
		t.Fatal("immutable original245 changed")
	}
	var original struct {
		Original string `json:"original_sha256"`
		Rows     []struct {
			Case struct {
				Name                                            string `json:"name"`
				Bar, Row, Slot, Flag, Event, X, Y, Target, Mode int
				Dropped                                         uint32 `json:"dropped"`
				PlayerClass                                     int    `json:"player_class"`
			} `json:"case"`
			Changes        [][3]int `json:"changed_bytes"`
			Calls          []string `json:"calls"`
			Press, Release int
			Captured       bool `json:"captured"`
			LastButtonNil  bool `json:"last_button_nil"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(raw, &original); err != nil {
		t.Fatal(err)
	}
	if original.Original != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(original.Rows) != 245 {
		t.Fatal("original source/count")
	}
	for _, row := range original.Rows {
		t.Run(row.Case.Name, func(t *testing.T) {
			g2qTestIsolate(t)
			c := row.Case
			main, free := alloc.MakeHigh([]byte{}, 256)
			defer free()
			trap := memmap.Slice(0x5D4594, 1047940)[:256]
			bars := [2][]byte{main, trap}
			bases := [2]unsafe.Pointer{unsafe.Pointer(&main[0]), unsafe.Pointer(&trap[0])}
			local := *memmap.PtrPtr(0x852978, 8)
			*memmap.PtrPtr(0x852978, 8) = nil
			defer func() { *memmap.PtrPtr(0x852978, 8) = local }()
			objects := g2qTestObjects(t, 32)
			stride := g2qTestLayout(6)
			var buttons [2][5]unsafe.Pointer
			for b := 0; b < 2; b++ {
				for i := range bars[b] {
					bars[b][i] = 0xa5
				}
				for r := 0; r < 5; r++ {
					for s := 0; s < 5; s++ {
						o := r*40 + s*8
						binary.LittleEndian.PutUint32(bars[b][o:], uint32(27+r*5+s))
						bars[b][o+4] = byte((r + s) & 1)
					}
				}
				bars[b][200] = 0
				w := unsafe.Add(objects, uintptr(b*16)*stride)
				g2qTestOwner(b, bases[b], w, unsafe.Add(w, stride), unsafe.Add(w, 6*stride), unsafe.Add(w, 11*stride))
				for s := 0; s < 5; s++ {
					buttons[b][s] = unsafe.Add(w, uintptr(s+1)*stride)
				}
			}
			bars[c.Bar][200] = byte(c.Row)
			bars[c.Bar][40*c.Row+8*c.Slot+4] = byte(c.Flag)
			var expected [2][]byte
			for b := range bars {
				expected[b] = append([]byte(nil), bars[b][:201]...)
			}
			for _, change := range row.Changes {
				expected[change[0]][change[1]] = byte(change[2])
			}
			player, fp := alloc.MakeHigh([]byte{}, int(g2qTestLayout(8)))
			defer fp()
			player[g2qTestLayout(9)] = byte(c.PlayerClass)
			g2qTestPointer(1, unsafe.Pointer(&player[0]), true)
			win := buttons[c.Bar][c.Slot]
			var capture unsafe.Pointer
			var calls []string
			record := func(s string) { calls = append(calls, s) }
			selected := func() []byte { return bars[c.Bar][int(bars[c.Bar][200])*40:] }
			h := game2QuickbarTestHooks
			h.Blocked = func() int32 { return 0 }
			h.Enabled = func(spell uint32) bool {
				if spell == 0 {
					t.Fatal("enabled zero")
				}
				return true
			}
			h.Capture = func() unsafe.Pointer { return capture }
			h.SetCapture = func(p unsafe.Pointer) {
				if p != win {
					t.Fatal("set capture arg")
				}
				capture = p
				record("capture")
			}
			h.ClearCapture = func(p unsafe.Pointer) {
				if p != win {
					t.Fatal("clear capture arg")
				}
				capture = nil
				record("release")
			}
			h.DragSave = func(spell uint32, kind int32) {
				if spell != binary.LittleEndian.Uint32(selected()[8*c.Slot:]) || kind != 2 {
					t.Fatal("drag args")
				}
				record("drag")
			}
			h.DragClear = func() { record("clear-dnd") }
			h.Sound = func(id, volume int32) {
				if volume != 100 {
					t.Fatal("sound volume")
				}
				switch id {
				case 793:
					record("pickup")
				case 794:
					record("swap-sound")
				default:
					t.Fatalf("unexpected sound%d", id)
				}
			}
			h.PointIn = func(p unsafe.Pointer, x, y int32) bool {
				if x != int32(c.X) || y != int32(c.Y) {
					t.Fatal("hit coords")
				}
				for i, w := range buttons[c.Bar] {
					if p == w {
						if i == 0 {
							record("hit")
						}
						return i == c.Target
					}
				}
				t.Fatal("unknown button")
				return false
			}
			rectangle := memmap.Slice(0x587000, 133656)[:16]
			savedRect := append([]byte(nil), rectangle...)
			defer copy(rectangle, savedRect)
			for i, v := range []uint32{352, 677, 672, 768} {
				binary.LittleEndian.PutUint32(rectangle[4*i:], v)
			}
			h.PointRect = nil // actual unchanged native rectangle entry, not a copied branch model
			h.Drop = func(spell uint32, flag byte, x, y int32, data unsafe.Pointer) uint32 {
				if data != bases[c.Bar] || spell != binary.LittleEndian.Uint32(selected()[8*c.Slot:]) || flag != selected()[8*c.Slot+4] || x != int32(c.X) || y != int32(c.Y) {
					t.Fatal("drop args")
				}
				record("drop")
				return c.Dropped
			}
			h.EventSwap = func(data unsafe.Pointer, a, b int32) {
				if data != bases[c.Bar] || a != int32(c.Slot) || b != int32(c.Target) {
					t.Fatal("swap args")
				}
				r := selected()
				tmp := append([]byte(nil), r[8*a:8*a+8]...)
				copy(r[8*a:8*a+8], r[8*b:8*b+8])
				copy(r[8*b:8*b+8], tmp)
				record("swap")
			}
			h.EventUpdate = func(data unsafe.Pointer) {
				if data != bases[c.Bar] {
					t.Fatal("icons data")
				}
				record("icons")
			}
			h.SpellCursor = func(spell uint32, flag byte) int32 {
				if spell != binary.LittleEndian.Uint32(selected()[8*c.Slot:]) || flag != selected()[8*c.Slot+4] {
					t.Fatal("cursor args")
				}
				record("cursor")
				return 0
			}
			h.AbilityCursor = func(spell uint32) int32 {
				if spell != binary.LittleEndian.Uint32(selected()[8*c.Slot:]) {
					t.Fatal("ability args")
				}
				record("ability")
				return 0
			}
			h.StoreButton = func(slot int32) {
				if slot != int32(c.Slot) {
					t.Fatal("store arg")
				}
				record("store")
			}
			h.CancelAbility = func() { t.Fatal("unselected original cancel") }
			h.TargetClick = func(uint32) { t.Fatal("unselected original target") }
			press := int32(-1)
			if c.Mode != 3 {
				press = g2qTestEvent(win, 5, 743<<16|438)
				if g2qTestPointer(9, nil, false) != bases[c.Bar] || g2qTestPointer(10, nil, false) != unsafe.Add(bases[c.Bar], 40*c.Row) {
					t.Fatal("native drag owner/row")
				}
			}
			if c.Mode == 1 {
				g2qTestPointer(10, nil, true)
			}
			if c.Mode == 2 {
				bars[c.Bar][200] = byte((c.Row + 1) % 5)
			}
			release := g2qTestEvent(win, c.Event, uint32(c.X)|uint32(c.Y)<<16)
			for b := range bars {
				if !bytes.Equal(bars[b][:201], expected[b]) {
					t.Fatalf("bar%d Go/publicC bytes%x != original%x", b, bars[b][:201], expected[b])
				}
				if !bytes.Equal(bars[b][201:], bytes.Repeat([]byte{0xa5}, 55)) {
					t.Fatal("packed pointer/padding canary")
				}
			}
			if len(calls) != len(row.Calls) || (len(calls) > 0 && !reflect.DeepEqual(calls, row.Calls)) {
				t.Fatalf("actual Go service order %v != original %v", calls, row.Calls)
			}
			if press != int32(row.Press) || release != int32(row.Release) || (capture != nil) != row.Captured || (g2qTestPointer(8, nil, false) == nil) != row.LastButtonNil {
				t.Fatal("actual Go/publicC return/capture/cleanup differs frozen original")
			}
			for _, v := range unsafe.Slice((*byte)(objects), int(32*stride)) {
				if v != 0 {
					t.Fatal("window bytes mutated")
				}
			}
		})
	}
}

func TestGame2PlayerClassNativePackedField(t *testing.T) {
	size, off := int(g2qTestLayout(8)), int(g2qTestLayout(9))
	data, free := alloc.MakeHigh([]byte{}, size+32)
	defer free()
	p := unsafe.Pointer(&data[16])
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
		t.Fatal("native C player above 4GiB required")
	}
	for i := range data {
		data[i] = byte(0xa3 + i*7)
	}
	for _, class := range []byte{0, 1, 2, 127, 255} {
		data[16+off] = class
		before := bytes.Clone(data)
		runtime.GC()
		if got := g2qTestPlayerClass(p); got != class {
			t.Fatalf("packed C offsetof class %d, native Go accessor %d", class, got)
		}
		if !bytes.Equal(data, before) {
			t.Fatal("class accessor changed player bytes or canaries")
		}
	}
	if g2qTestPlayerClass(nil) != 0 {
		t.Fatal("nil native player")
	}
}
