package legacy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

var headSettingsRegions409470 = []struct {
	base, off uintptr
	size      int
}{
	{0x5d4594, 1320, 24}, {0x5d4594, 3450, 144},
	{0x5d4594, 2598184, 88}, {0x85b3fc, 32, 88},
	{0x587000, 4704, 36},
}

type headSettingsState409470 struct {
	Regions         [][]byte
	Subflags, Dirty uint32
}

func headSettingsSnapshot409470() headSettingsState409470 {
	v := headSettingsState409470{Subflags: *(*uint32)(settingFlagsNativeCell409E40(0)), Dirty: *(*uint32)(settingFlagsNativeCell409E40(1))}
	for _, r := range headSettingsRegions409470 {
		v.Regions = append(v.Regions, bytes.Clone(unsafe.Slice((*byte)(memmap.PtrOff(r.base, r.off)), r.size)))
	}
	return v
}
func headSettingsRestore409470(v headSettingsState409470) {
	for i, r := range headSettingsRegions409470 {
		copy(unsafe.Slice((*byte)(memmap.PtrOff(r.base, r.off)), r.size), v.Regions[i])
	}
	*(*uint32)(settingFlagsNativeCell409E40(0)) = v.Subflags
	*(*uint32)(settingFlagsNativeCell409E40(1)) = v.Dirty
}
func headSettingsSetup409470(t *testing.T) headSettingsState409470 {
	t.Helper()
	old := headSettingsSnapshot409470()
	t.Cleanup(func() { headSettingsRestore409470(old) })
	for i, r := range headSettingsRegions409470 {
		b := unsafe.Slice((*byte)(memmap.PtrOff(r.base, r.off)), r.size)
		for j := range b {
			b[j] = byte(0xa0 + i)
		}
	}
	for _, p := range []unsafe.Pointer{headSettingsPath409470(), headSettingsMap409470(), headSettingsName409470(), headSettingsPass409470()} {
		*(*byte)(p) = 0
	}
	for i := 0; i < 6; i++ {
		*(*uint32)(memmap.PtrOff(0x587000, uintptr(4704+4*i))) = uint32(i * 0x10)
		*(*uint16)(memmap.PtrOff(0x5d4594, uintptr(3488+2*i))) = uint16(0x8000 + i)
	}
	*(*uint32)(settingFlagsNativeCell409E40(0)) = 0xfedcba98
	*(*uint32)(settingFlagsNativeCell409E40(1)) = 0x11223344
	return headSettingsSnapshot409470()
}
func headSettingsHigh409470(t *testing.T, p unsafe.Pointer, label string) {
	t.Helper()
	if p == nil || unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= 0xffffffff {
		t.Fatalf("%s native allocation/address %#x", label, uintptr(p))
	}
}
func headSettingsString409470(t *testing.T, s string, wide bool) unsafe.Pointer {
	t.Helper()
	n := len(s) + 1
	if wide {
		n *= 2
	}
	p := headSettingsFixtureAlloc409470(uintptr(n))
	if p == nil {
		t.Fatal("calloc")
	}
	t.Cleanup(func() { headSettingsFixtureFree409470(p) })
	if wide {
		for i, b := range []byte(s) {
			*(*uint16)(unsafe.Add(p, 2*i)) = uint16(b)
		}
	} else {
		copy(unsafe.Slice((*byte)(p), n), s)
	}
	headSettingsHigh409470(t, p, "actual C source")
	return p
}

// Current-C differential reference, not original-PE equivalence. Every C leaf
// is actually called, including returned native addresses and the dirty owner.
func TestHeadSettings21ActualCABIAndCurrentCReference(t *testing.T) {
	initial := headSettingsSetup409470(t)
	wordEntries := []uint32{0x409e60, 0x409f80, 0x409fa0, 0x40a020, 0x40a340, 0x40a350, 0x40a5e0, 0x40a5f0, 0x40a670, 0x40a680, 0x40a690, 0x40a6c0, 0x40a6d0}
	for _, entry := range wordEntries {
		for _, arg := range []int32{0, 1, -1, -2147483648, 2147483647, 0x50} {
			for _, flag := range []bool{false, true} {
				t.Run(fmt.Sprintf("%06x/%08x/%t", entry, uint32(arg), flag), func(t *testing.T) {
					var traces [2][][2]uint32
					var results [2]uint32
					var states [2]headSettingsState409470
					for backend, ref := range []bool{true, false} {
						headSettingsRestore409470(initial)
						o := &headSettingsObserver409470{Flags: func(mask uint32) bool { traces[backend] = append(traces[backend], [2]uint32{1, mask}); return flag }, Notify: func(v int32) int32 {
							traces[backend] = append(traces[backend], [2]uint32{2, uint32(v)})
							return -7654321
						}}
						func() {
							defer headSettingsObserve409470(o)()
							results[backend] = headSettingsFixtureWord409470(ref, entry, arg)
						}()
						states[backend] = headSettingsSnapshot409470()
					}
					if results[0] != results[1] || !reflect.DeepEqual(states[0], states[1]) || !reflect.DeepEqual(traces[0], traces[1]) {
						t.Fatalf("current C comparison entry%x: result%v traces%v", entry, results, traces)
					}
				})
			}
		}
	}
	for _, x := range []struct {
		entry uint32
		p     unsafe.Pointer
	}{{0x409b30, headSettingsPath409470()}, {0x409b40, headSettingsMap409470()}, {0x40a4c0, headSettingsName409470()}, {0x40a630, headSettingsPass409470()}} {
		got := headSettingsFixturePointer409470(false, x.entry, nil)
		if got != x.p || got != headSettingsFixturePointer409470(true, x.entry, nil) {
			t.Fatalf("getter%x native alias", x.entry)
		}
		headSettingsHigh409470(t, got, "public C getter")
	}
	if settingFlagsNativeCellWidth409E40(0) != 4 || settingFlagsNativeCellWidth409E40(1) != 4 {
		t.Fatal("actual owner DWORD widths")
	}
	for _, p := range []unsafe.Pointer{settingFlagsNativeCell409E40(0), settingFlagsNativeCell409E40(1)} {
		headSettingsHigh409470(t, p, "actual C global")
	}
}

func TestHeadSettingsStringsCurrentCAndNativePolicy(t *testing.T) {
	initial := headSettingsSetup409470(t)
	for _, tc := range []struct {
		entry uint32
		value string
		wide  bool
	}{
		{0x409d70, "War01a.map", false}, {0x409d70, "x", false}, {0x409d70, "", false}, {0x409d70, "dir\\War01a.map", false},
		{0x409d70, strings.Repeat("d", 90) + "\\map.map", false},
		{0x409d70, "dir\\" + strings.Repeat("m", 79) + ".map", false},
		{0x40a440, "", false}, {0x40a440, "Host", false}, {0x40a440, strings.Repeat("n", 14), false}, {0x40a440, strings.Repeat("n", 15), false}, {0x40a440, strings.Repeat("n", 40), false},
		{0x40a610, "", true}, {0x40a610, "ABCxyz", true}, {0x40a610, strings.Repeat("p", 21), true},
	} {
		t.Run(fmt.Sprintf("%x/%d", tc.entry, len(tc.value)), func(t *testing.T) {
			p := headSettingsString409470(t, tc.value, tc.wide)
			var states [2]headSettingsState409470
			var results [2]unsafe.Pointer
			for i, ref := range []bool{true, false} {
				headSettingsRestore409470(initial)
				results[i] = headSettingsFixturePointer409470(ref, tc.entry, p)
				states[i] = headSettingsSnapshot409470()
			}
			if results[0] != results[1] || !reflect.DeepEqual(states[0], states[1]) {
				t.Fatalf("current-C string return/whole-region mismatch%x", tc.entry)
			}
			if tc.entry != 0x40a610 {
				// Stored path/server name compare uses the truncated live string.
				// Valid short unchanged names must return nil and preserve dirty.
				headSettingsDirty409470(0x12345678)
				if tc.entry == 0x40a440 || len(tc.value) <= 79 {
					if got := headSettingsFixturePointer409470(false, tc.entry, p); got != nil || headSettingsDirtyGet409470() != 0x12345678 {
						t.Fatal("unchanged string rewrote state")
					}
				}
			}
		})
	}
	// Nil server-name policy is OLD behavior: clear byte0 only, dirty=1.
	headSettingsRestore409470(initial)
	name := unsafe.Slice((*byte)(headSettingsName409470()), 16)
	for i := range name {
		name[i] = 0x6f
	}
	before := bytes.Clone(name)
	if headSettingsFixturePointer409470(false, 0x40a440, nil) != nil {
		t.Fatal("nil name return")
	}
	before[0] = 0
	if !bytes.Equal(name, before) || headSettingsDirtyGet409470() != 1 {
		t.Fatal("nil name byte-only side effect")
	}
	// Native-only safety domain: never run the old undefined huge strncpy.
	for _, s := range []string{"dir\\x", "dir\\", "dir\\" + strings.Repeat("b", 120) + ".map"} {
		headSettingsRestore409470(initial)
		p := headSettingsString409470(t, s, false)
		guard := bytes.Clone(unsafe.Slice((*byte)(memmap.PtrOff(0x85b3fc, 116)), 4))
		if headSettingsFixturePointer409470(false, 0x409d70, p) != headSettingsMap409470() {
			t.Fatal("native policy result")
		}
		n := 0
		if len(s) > 8 {
			n = 79
		}
		if *(*byte)(unsafe.Add(headSettingsMap409470(), n)) != 0 || !bytes.Equal(guard, unsafe.Slice((*byte)(memmap.PtrOff(0x85b3fc, 116)), 4)) {
			t.Fatal("native basename policy boundary")
		}
	}
	headSettingsRestore409470(initial)
	if headSettingsFixturePointer409470(false, 0x409d70, nil) != nil || !reflect.DeepEqual(initial, headSettingsSnapshot409470()) {
		t.Fatal("native nil-map policy")
	}
}

func TestHeadSettingsRateServiceOrderingAndObserverReentry(t *testing.T) {
	headSettingsSetup409470(t)
	p := (*uint32)(memmap.PtrOff(0x587000, 4728))
	for _, ref := range []bool{true, false} {
		*p = 0x13572468
		var trace []string
		outer := &headSettingsObserver409470{}
		outer.Flags = func(mask uint32) bool {
			if mask != 0x20000 || *p != 0x13572468 {
				t.Fatal("flags read/write order")
			}
			trace = append(trace, "flags")
			func() {
				inner := &headSettingsObserver409470{Flags: func(m uint32) bool { trace = append(trace, "inner"); return false }}
				defer headSettingsObserve409470(inner)()
				if headSettingsFixtureWord409470(false, 0x40a5f0, 0) != 0xffffff9d {
					t.Fatal("nested flag result")
				}
			}()
			return true
		}
		outer.Notify = func(v int32) int32 {
			if v != 159 || *p != 0x13572468 {
				t.Fatal("notify order")
			}
			trace = append(trace, "notify")
			*p = 0xfefefefe
			runtime.GC()
			return -7
		}
		func() {
			defer headSettingsObserve409470(outer)()
			headSettingsSetRespawn409470(-99)
			if got := headSettingsFixtureWord409470(ref, 0x40a6d0, -2147483648); got != 0xfffffff9 {
				t.Fatalf("notify return%x", got)
			}
		}()
		if *p != 0x80000000 || !reflect.DeepEqual(trace, []string{"flags", "inner", "notify"}) || headSettingsObserved409470 != nil {
			t.Fatal("final DWORD/reentry/LIFO")
		}
		func() {
			defer headSettingsObserve409470(&headSettingsObserver409470{Flags: func(uint32) bool { t.Fatal("equal rate called service"); return false }})()
			if headSettingsFixtureWord409470(ref, 0x40a6d0, -2147483648) != 0x80000000 {
				t.Fatal("same-rate return")
			}
		}()
	}
}

type headSettingsParsed409470 struct {
	Result   int32
	Bytes    []byte
	Position int64
	Error    bool
}

func TestHeadSettingsParseActualBinfilePlainAndEncrypted(t *testing.T) {
	// Isolate the real opaque-handle registry exactly as existing binfile tests.
	files.Lock()
	oldFiles := files.byHandle
	files.byHandle = make(map[unsafe.Pointer]*binfile.File)
	files.Unlock()
	handles.Init()
	t.Cleanup(func() {
		files.Lock()
		owned := files.byHandle
		files.byHandle = oldFiles
		files.Unlock()
		for _, f := range owned {
			_ = f.Close()
		}
		handles.Release()
	})
	for _, key := range []int{-1, 5} {
		for i, text := range []string{"", "   \t\r\n", "alpha beta\n", "  first\tsecond ", "partialEOF", "slash/ tail ", "//abc\nnext ", "///\nnext ", "x // q\nend ", "A\vB\fC\rD\tE\n", "nul\x00byte ", strings.Repeat("x", 255) + " "} {
			t.Run(fmt.Sprintf("key%d/case%d", key, i), func(t *testing.T) {
				path := filepath.Join(t.TempDir(), "parse.bin")
				writer, err := binfile.BinfileOpen(path, binfile.WriteOnly)
				if err != nil {
					t.Fatal(err)
				}
				if err = writer.SetKey(key); err != nil {
					t.Fatal(err)
				}
				if n, e := writer.Write([]byte(text)); e != nil || n != len(text) {
					t.Fatalf("write %d/%v", n, e)
				}
				if err = writer.Close(); err != nil {
					t.Fatal(err)
				}
				var rows [2][]headSettingsParsed409470
				for backend, ref := range []bool{true, false} {
					reader, err := binfile.BinfileOpen(path, binfile.ReadOnly)
					if err != nil {
						t.Fatal(err)
					}
					if err = reader.SetKey(key); err != nil {
						t.Fatal(err)
					}
					h := NewFileHandle(reader.File)
					headSettingsHigh409470(t, unsafe.Pointer(h), "actual opaque binfile handle")
					func() {
						defer Nox_fs_close(h)
						out := headSettingsFixtureAlloc409470(288)
						if out == nil {
							t.Fatal("calloc token")
						}
						defer headSettingsFixtureFree409470(out)
						headSettingsHigh409470(t, out, "actual C token output")
						for call := 0; call < 32; call++ {
							b := unsafe.Slice((*byte)(out), 288)
							for i := range b {
								b[i] = 0xcc
							}
							v := headSettingsFixtureParse409470(ref, unsafe.Pointer(h), unsafe.Add(out, 16))
							pos, e := reader.File.File.Seek(0, io.SeekCurrent)
							if e != nil {
								t.Fatal(e)
							}
							rows[backend] = append(rows[backend], headSettingsParsed409470{v, bytes.Clone(b), pos, reader.File.Err != nil})
							if !bytes.Equal(b[:16], bytes.Repeat([]byte{0xcc}, 16)) || !bytes.Equal(b[272:], bytes.Repeat([]byte{0xcc}, 16)) {
								t.Fatal("token canary")
							}
							if v == 0 {
								break
							}
							if call == 31 {
								t.Fatal("parser did not reach bounded EOF")
							}
						}
					}()
				}
				if !reflect.DeepEqual(rows[0], rows[1]) {
					t.Fatalf("actual binfile/current-C partial-write or handle-position mismatch: %+v/%+v", rows[0], rows[1])
				}
				if key == -1 && text == "alpha beta\n" {
					if rows[1][0].Result != 1 || !bytes.Equal(rows[1][0].Bytes[16:22], []byte("alpha\x00")) {
						t.Fatal("literal first token")
					}
				}
				if key == -1 && text == "partialEOF" {
					r := rows[1][0]
					if r.Result != 0 || !r.Error || !bytes.Equal(r.Bytes[16:26], []byte("partialEOF")) || r.Bytes[26] != 0xcc {
						t.Fatal("literal EOF partial bytes must remain unterminated")
					}
				}
			})
		}
	}
}

func TestHeadSettingsDWORDNeighborsAndCaseInsensitiveNoChange(t *testing.T) {
	headSettingsSetup409470(t)
	limit := unsafe.Slice((*byte)(memmap.PtrOff(0x5d4594, 3460)), 12)
	before := bytes.Clone(limit)
	if headSettingsFixtureWord409470(false, 0x409f80, -1) != 0xffffffff {
		t.Fatal("DWORD return")
	}
	binary.LittleEndian.PutUint32(before[4:8], 0xffffffff)
	if !bytes.Equal(limit, before) {
		t.Fatal("player limit write exceeded four bytes")
	}
	for _, tc := range []struct {
		entry         uint32
		first, second string
	}{{0x409d70, "maps\\War01a.map", "MAPS\\war01A.MAP"}, {0x40a440, "ServerOne", "sERVERoNE"}, {0x40a440, "123456789012345A", "123456789012345B"}} {
		p := headSettingsString409470(t, tc.first, false)
		headSettingsFixturePointer409470(false, tc.entry, p)
		headSettingsDirty409470(0x76543210)
		snapshot := headSettingsSnapshot409470()
		q := headSettingsString409470(t, tc.second, false)
		if headSettingsFixturePointer409470(false, tc.entry, q) != nil || !reflect.DeepEqual(snapshot, headSettingsSnapshot409470()) {
			t.Fatal("case-insensitive/truncated-name no-change contract")
		}
	}
}
