package legacy

import (
	"bytes"
	"fmt"
	"runtime"
	"testing"
	"unicode/utf16"
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestGame3ConfigCEntriesMatchRetainedReference(t *testing.T) {
	bank := memmap.Slice(0x5D4594, 1308172)[:628]
	before := append([]byte(nil), bank...)
	defer copy(bank, before)
	for which := 0; which <= 10; which++ {
		inputs := []string{"", "a", "127.0.0.1", " -17tail", "0", "2147483647", "-2147483648"}
		if which == 5 {
			inputs = append(inputs, string(bytes.Repeat([]byte{'x'}, 128)), string(bytes.Repeat([]byte{'y'}, 129)))
		}
		if which == 10 {
			inputs = []string{"12,34", " 12, 34", "4,5,6", "1\t2", "1 2", "0,1", "1,0", "", "-9,7"}
		}
		for _, text := range inputs {
			t.Run(fmt.Sprintf("%d/%q", which, text), func(t *testing.T) {
				for i := range bank {
					bank[i] = byte(0x31 + i%127)
				}
				seed := append([]byte(nil), bank...)
				want, wantInput, _ := game3FixtureConfig(true, which, &text)
				wantBank := append([]byte(nil), bank...)
				copy(bank, seed)
				got, gotInput, address := game3FixtureConfig(false, which, &text)
				if got != want || !bytes.Equal(bank, wantBank) || !bytes.Equal(gotInput, wantInput) {
					t.Fatalf("C→Go config entry %d differs: got%d want%d", which, got, want)
				}
				if unsafe.Sizeof(uintptr(0)) > 4 && address <= 0xffffffff {
					t.Fatal("foreign input must exercise full native address")
				}
			})
		}
		copy(bank, before)
		want, _, _ := game3FixtureConfig(true, which, nil)
		got, _, _ := game3FixtureConfig(false, which, nil)
		if want != 0 || got != 0 || !bytes.Equal(bank, before) {
			t.Fatal("nil config input changed state")
		}
	}
	if game3FixtureConfigAddress() != memmap.PtrOff(0x5D4594, 1308732) {
		t.Fatal("borrowed config pair address narrowed or copied")
	}
}

func TestGame3ColorNameMatchesRetainedC(t *testing.T) {
	for _, text := range []string{"", " ", "\t\r\n", "name", "  Name  ", "*first?", " ?first/second ", "a|b", `\\foo`, "  :  ", "ab\tcd  ", "abcdefghijklmnopqrstuvwxy"} {
		t.Run(text, func(t *testing.T) {
			words := append(utf16.Encode([]rune(text)), 0)
			want, wb := game3FixtureName(true, words)
			got, gb := game3FixtureName(false, words)
			if got != want || !bytes.Equal(wb, gb) {
				t.Fatalf("name/scratch-tail/canary differs got%d want%d", got, want)
			}
		})
	}
}

func TestGame3ConfigLookupNativeSidecars(t *testing.T) {
	const off = uintptr(171856)
	var old [3]unsafe.Pointer
	var words [3]uint32
	for i := range old {
		old[i] = *memmap.PtrPtr(0x587000, off+8*uintptr(i))
		words[i] = memmap.Uint32(0x587000, off+4+8*uintptr(i))
	}
	previous := memmap.Uint32(0x5D4594, 1308184)
	a, fa := alloc.CString("First")
	b, fb := alloc.CString("Second")
	defer func() {
		for i := range old {
			*memmap.PtrPtr(0x587000, off+8*uintptr(i)) = old[i]
			*memmap.PtrUint32(0x587000, off+4+8*uintptr(i)) = words[i]
		}
		*memmap.PtrUint32(0x5D4594, 1308184) = previous
		fb()
		fa()
	}()
	*memmap.PtrPtr(0x587000, off) = unsafe.Pointer(a)
	*memmap.PtrPtr(0x587000, off+8) = unsafe.Pointer(b)
	*memmap.PtrPtr(0x587000, off+16) = nil
	*memmap.PtrUint32(0x587000, off+4) = 0x92345678
	*memmap.PtrUint32(0x587000, off+12) = 0xfedcba98
	if unsafe.Sizeof(uintptr(0)) > 4 && (uintptr(unsafe.Pointer(a)) <= 0xffffffff || uintptr(unsafe.Pointer(b)) <= 0xffffffff) {
		t.Fatal("native table strings must be high addresses")
	}
	// Actual public call is separate from the historical packed-pointer C body:
	// the native sidecars must never overwrite adjacent DWORD enum fields.
	for _, tc := range []struct {
		s    string
		want uint32
	}{{"fIrSt", 0x92345678}, {"SECOND", 0xfedcba98}, {"absent", 0xfedcba98}} {
		if game3FixtureLookup(tc.s) != 1 || memmap.Uint32(0x5D4594, 1308184) != tc.want {
			t.Fatal("full-width lookup/case/unknown-state contract")
		}
	}
	if memmap.Uint32(0x587000, off+4) != 0x92345678 || memmap.Uint32(0x587000, off+12) != 0xfedcba98 {
		t.Fatal("native pointers clobbered packed DWORD neighbors")
	}
}

func TestGame3PaletteNativePointerAndMixedReturn(t *testing.T) {
	g := gui.New(nil)
	t.Cleanup(g.DestroyAll)
	root := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 100, 100, nil)
	old := game3FixtureSwapColorRoot(root)
	t.Cleanup(func() { game3FixtureSwapColorRoot(old) })
	w := g.NewWindowRaw(root, gui.StatusEnabled, 0, 0, 8, 8, nil)
	w.SetID(721)
	left := g.NewWindowRaw(root, gui.StatusEnabled, 0, 0, 8, 8, nil)
	left.SetID(711)
	right := g.NewWindowRaw(root, gui.StatusEnabled, 0, 0, 8, 8, nil)
	right.SetID(731)
	palette := memmap.Slice(0x5D4594, 1307796)[:288]
	saved := append([]byte(nil), palette...)
	defer copy(palette, saved)
	for i := range palette {
		palette[i] = byte(i)
	}
	for _, group := range []int{0, 1, 2} {
		for _, index := range []int{0, 7, 31, 32} {
			rgb := [3]byte{0xff, 0xff, 0xff}
			if index < 32 {
				copy(rgb[:], palette[group*96+3*index:])
			}
			flags, draw := left.Flags, right.DrawData().Field0
			want := game3FixtureMatch(true, w, group, rgb)
			wantFlags, wantDraw := left.Flags, right.DrawData().Field0
			left.Flags = flags
			right.DrawData().Field0 = draw
			got := game3FixtureMatch(false, w, group, rgb)
			if got != want || left.Flags != wantFlags || right.DrawData().Field0 != wantDraw {
				t.Fatalf("palette pointer/status return differs group%d index%d", group, index)
			}
			if group == 1 && index < 32 && got != uintptr(right.C()) {
				t.Fatal("matched palette window address was narrowed")
			}
		}
	}
	runtime.GC()
}

func TestGame3RadioNativeCEntryAndGC(t *testing.T) {
	g := gui.New(nil)
	t.Cleanup(g.DestroyAll)
	var sender uintptr
	parent := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 100, 100, func(_ *gui.Window, e gui.WindowEvent) gui.WindowEventResp {
		if e.EventCode() == 0x4007 {
			sender, _ = e.EventArgsC()
			runtime.GC()
		}
		return gui.RawEventResp(1)
	})
	makeRadio := func() *gui.Window {
		w := g.NewWindowRaw(parent, gui.StatusEnabled, 0, 0, 10, 10, gui.RadioButtonProcPre)
		d := gui.WindowData{Window: parent, Style: gui.StyleRadioButton}
		d.SetGroup(17)
		w.CopyDrawData(&d)
		if game3FixtureRadioInit(w) != 0 {
			t.Fatal("C native radio init")
		}
		return w
	}
	a, b := makeRadio(), makeRadio()
	proc, draw := game3FixtureRadioCallbackWords(a)
	if proc == 0 || draw == 0 || (unsafe.Sizeof(uintptr(0)) > 4 && (proc <= 0xffffffff || draw <= 0xffffffff)) {
		t.Fatal("actual full-width C callback registrations missing")
	}
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(a.C()) <= 0xffffffff {
		t.Fatal("C entry must use full-width native windows")
	}
	if game3FixtureRadio(a, 6, 0, 0, false) != 1 || sender != uintptr(a.C()) || a.DrawData().Field0&4 == 0 {
		t.Fatal("C→Go radio click/parent pointer")
	}
	game3FixtureRadio(b, 16392, 1, 0, true)
	if b.DrawData().Field0&4 == 0 || a.DrawData().Field0&4 != 0 {
		t.Fatal("programmatic native sibling traversal")
	}
	old := game3FixtureSwapOptionsRoot(nil)
	defer game3FixtureSwapOptionsRoot(old)
	if got := game3FixtureUnknownOptions(a); got != [3]int{0, 0, 1} {
		t.Fatalf("bounded nil-options/unknown-event/key states %v", got)
	}
	if game3FixtureLayout(0) != unsafe.Sizeof(gui.Window{}) || game3FixtureLayout(5) != unsafe.Sizeof(gui.WindowData{}) {
		t.Fatal("actual C/Go native window layout")
	}
}

func TestGame3RadioCTextPreservesRawUTF16Bound(t *testing.T) {
	g := gui.New(nil)
	t.Cleanup(g.DestroyAll)
	w := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 10, 10, nil)
	words := make([]uint16, 63)
	for i := range words {
		words[i] = uint16(0xD800 + i)
	}
	got := game3FixtureRadioText(w, words)
	for i := range words {
		if got[i] != words[i] {
			t.Fatal("C bounded text converted raw surrogate words")
		}
	}
	if got[63] != 0 {
		t.Fatal("bounded text lacks forced last NUL")
	}
	got = game3FixtureRadioText(w, []uint16{'x', 0, 0xD8EF})
	if got[0] != 'x' {
		t.Fatal("C text prefix")
	}
	for _, v := range got[1:] {
		if v != 0 {
			t.Fatal("C text strncpy padding")
		}
	}
}
