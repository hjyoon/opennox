//go:build darwin && arm64

package legacy

import (
	"bytes"
	"encoding/binary"
	"runtime"
	"testing"
	"unicode/utf16"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

type titleTestRow415960 struct {
	Title, Key unsafe.Pointer
	Mask       uint32
}

func titleTestWide415960(t *testing.T, text string) unsafe.Pointer {
	t.Helper()
	b := append(utf16.Encode([]rune(text)), 0)
	p := new(runtime.Pinner)
	p.Pin(&b[0])
	t.Cleanup(func() { p.Unpin(); runtime.KeepAlive(b) })
	return unsafe.Pointer(&b[0])
}

func titleTestKey415960(t *testing.T, text string) unsafe.Pointer {
	t.Helper()
	b := make([]byte, len(text)+8)
	copy(b, text)
	p := new(runtime.Pinner)
	p.Pin(&b[0])
	t.Cleanup(func() { p.Unpin(); runtime.KeepAlive(b) })
	return unsafe.Pointer(&b[0])
}

func titleTestTable415960(armor bool) titleTable415960 {
	if armor {
		return titleArmor415960
	}
	return titleWeapon415960
}

// Install only this family's fixed blob range; retain and restore all native
// side slots, all packed bytes, and the scalar flag's surrounding canaries.
func titleTestInstall415960(t *testing.T, table titleTable415960, rows []titleTestRow415960) []byte {
	t.Helper()
	count := 25
	if table.Base == titleArmor415960.Base {
		count = 26
	}
	if len(rows) > count {
		t.Fatal("fixture exceeds source table")
	}
	packed := unsafe.Slice((*byte)(memmap.PtrOff(0x587000, table.Base)), 12*(count+1))
	prior := append([]byte(nil), packed...)
	pointers := make([]unsafe.Pointer, 2*(count+1))
	for i := 0; i <= count; i++ {
		pointers[2*i] = *memmap.PtrPtr(0x587000, table.Base+uintptr(12*i))
		pointers[2*i+1] = *memmap.PtrPtr(0x587000, table.Base+uintptr(12*i+4))
	}
	flagBytes := unsafe.Slice((*byte)(memmap.PtrOff(0x5D4594, table.Ready-4)), 12)
	priorFlag := append([]byte(nil), flagBytes...)
	t.Cleanup(func() {
		copy(packed, prior)
		copy(flagBytes, priorFlag)
		for i := 0; i <= count; i++ {
			*memmap.PtrPtr(0x587000, table.Base+uintptr(12*i)) = pointers[2*i]
			*memmap.PtrPtr(0x587000, table.Base+uintptr(12*i+4)) = pointers[2*i+1]
		}
	})
	for i := range packed {
		packed[i] = byte(0xa1 + i%19)
	}
	for i := 0; i <= count; i++ {
		*memmap.PtrPtr(0x587000, table.Base+uintptr(12*i)) = nil
		*memmap.PtrPtr(0x587000, table.Base+uintptr(12*i+4)) = nil
	}
	for i, row := range rows {
		off := table.Base + uintptr(12*i)
		*memmap.PtrPtr(0x587000, off) = row.Title
		*memmap.PtrPtr(0x587000, off+4) = row.Key
		*memmap.PtrUint32(0x587000, off+8) = row.Mask
	}
	for i := range flagBytes {
		flagBytes[i] = byte(0x61 + i)
	}
	*memmap.PtrUint32(0x5D4594, table.Ready) = 0
	return packed
}

func titleTestTo415960(backend string, armor bool, query unsafe.Pointer) uint32 {
	if backend == "C" {
		return titleFixtureToMask415960(armor, query)
	}
	return titleToMask415960(titleTestTable415960(armor), query)
}

func titleTestFrom415960(backend string, armor bool, mask uint32) unsafe.Pointer {
	if backend == "C" {
		return titleFixtureFromMask415960(armor, mask)
	}
	return titleFromMask415960(titleTestTable415960(armor), mask)
}

func titleTestLoad415960(backend string, armor bool) uintptr {
	if backend == "C" {
		return titleFixtureLoad415960(armor)
	}
	return titleLoadTable415960(titleTestTable415960(armor))
}

func TestTitleTablesNative415960(t *testing.T) {
	for _, armor := range []bool{false, true} {
		name := "weapon"
		if armor {
			name = "armor"
		}
		for _, backend := range []string{"Go", "C"} {
			t.Run(name+"/"+backend+"/localized_field_and_scalar", func(t *testing.T) {
				a := titleTestWide415960(t, "Fire Staff")
				b := titleTestWide415960(t, "Leather")
				q := titleTestWide415960(t, "fIrE sTaFf")
				key := titleTestKey415960(t, "Modifier.db:UnrelatedKey")
				for _, p := range []unsafe.Pointer{a, b, q, key} {
					if uintptr(p) <= 0xffffffff {
						t.Fatal("control requires actual >4GiB owner")
					}
				}
				table := titleTestTable415960(armor)
				packed := titleTestInstall415960(t, table, []titleTestRow415960{{a, key, 0x80000001}, {b, key, 0xffffffff}})
				before := append([]byte(nil), packed...)
				if titleTestTo415960(backend, armor, q) != 0x80000001 || titleTestFrom415960(backend, armor, 0xffffffff) != b {
					t.Fatal("localized +0 or full DWORD mask was lost")
				}
				if titleTestFrom415960(backend, armor, 7) != nil || titleTestTo415960(backend, armor, titleTestWide415960(t, "missing")) != 0 {
					t.Fatal("missing lookup")
				}
				if !bytes.Equal(packed, before) {
					t.Fatal("lookup wrote packed 12-byte source rows")
				}
			})
			t.Run(name+"/"+backend+"/first_match_aliases", func(t *testing.T) {
				table := titleTestTable415960(armor)
				a := titleTestWide415960(t, "Same")
				b := titleTestWide415960(t, "Other")
				q := titleTestWide415960(t, "sAME")
				titleTestInstall415960(t, table, []titleTestRow415960{{a, nil, 0x1234}, {a, nil, 0x5678}, {b, nil, 0x1234}})
				if titleTestTo415960(backend, armor, q) != 0x1234 || titleTestFrom415960(backend, armor, 0x1234) != a {
					t.Fatal("source-order first matching title/mask or pointer alias changed")
				}
			})
			t.Run(name+"/"+backend+"/title_nil_sentinel_and_query_policy", func(t *testing.T) {
				table := titleTestTable415960(armor)
				later := titleTestWide415960(t, "Later")
				key := titleTestKey415960(t, "nonnull")
				titleTestInstall415960(t, table, []titleTestRow415960{{nil, key, 1}, {later, key, 2}})
				if titleTestTo415960(backend, armor, later) != 0 || titleTestFrom415960(backend, armor, 2) != nil {
					t.Fatal("key +4 replaced localized +0 sentinel")
				}
				accesses := 0
				pop := titlePushObserver415960(&titleObserver415960{Access: func(titleAccess415960) { accesses++ }})
				defer pop()
				if titleTestTo415960(backend, armor, nil) != 0 || accesses != 0 {
					t.Fatal("explicit native nil-query compatibility policy")
				}
			})
			t.Run(name+"/"+backend+"/fresh_mask_and_result_reads", func(t *testing.T) {
				table := titleTestTable415960(armor)
				a := titleTestWide415960(t, "A")
				b := titleTestWide415960(t, "B")
				q := titleTestWide415960(t, "query")
				titleTestInstall415960(t, table, []titleTestRow415960{{a, nil, 1}})
				compared := false
				func() {
					pop := titlePushObserver415960(&titleObserver415960{Compare: func(query, title unsafe.Pointer) (int32, bool) {
						if query != q || title != a {
							t.Fatal("native comparison arguments")
						}
						compared = true
						*memmap.PtrUint32(0x587000, table.Base+8) = 0xf1234567
						return 0, true
					}})
					defer pop()
					if titleTestTo415960(backend, armor, q) != 0xf1234567 {
						t.Fatal("mask cached across comparison")
					}
				}()
				if !compared {
					t.Fatal("comparison boundary not used")
				}
				pop := titlePushObserver415960(&titleObserver415960{Access: func(e titleAccess415960) {
					if !e.Write && e.Complete && e.Offset == table.Base+8 {
						*memmap.PtrPtr(0x587000, table.Base) = b
					}
				}})
				defer pop()
				if titleTestFrom415960(backend, armor, 0xf1234567) != b {
					t.Fatal("title cached before matching mask")
				}
			})
			t.Run(name+"/"+backend+"/next_title_reload_after_compare", func(t *testing.T) {
				table := titleTestTable415960(armor)
				a := titleTestWide415960(t, "A")
				b := titleTestWide415960(t, "B")
				titleTestInstall415960(t, table, []titleTestRow415960{{a, nil, 1}, {b, nil, 2}})
				calls := 0
				pop := titlePushObserver415960(&titleObserver415960{Compare: func(query, title unsafe.Pointer) (int32, bool) {
					calls++
					if title != a {
						t.Fatal("read stale next title after external service")
					}
					*memmap.PtrPtr(0x587000, table.Base+12) = nil
					return 1, true
				}})
				defer pop()
				if titleTestTo415960(backend, armor, b) != 0 || calls != 1 {
					t.Fatal("next localized sentinel reload")
				}
			})
			t.Run(name+"/"+backend+"/loader_sidecar_width_and_idempotence", func(t *testing.T) {
				table := titleTestTable415960(armor)
				a := titleTestWide415960(t, "First translated")
				b := titleTestWide415960(t, "Second translated")
				ka := titleTestKey415960(t, "key:first")
				kb := titleTestKey415960(t, "key:second")
				packed := titleTestInstall415960(t, table, []titleTestRow415960{{nil, ka, 1}, {nil, kb, 2}})
				before := append([]byte(nil), packed...)
				guard := unsafe.Slice((*byte)(memmap.PtrOff(0x5D4594, table.Ready-4)), 12)
				wantGuard := append([]byte(nil), guard...)
				binary.LittleEndian.PutUint32(wantGuard[4:8], 1)
				calls := 0
				pop := titlePushObserver415960(&titleObserver415960{Load: func(key unsafe.Pointer, source string, line int32) (unsafe.Pointer, bool) {
					if source != table.Source || line != table.Line {
						t.Fatal("localization source identity")
					}
					calls++
					if calls == 1 && key == ka {
						return a, true
					}
					if calls == 2 && key == kb {
						return b, true
					}
					t.Fatal("key order/full native identity")
					return nil, true
				}})
				defer pop()
				wantReturn := uintptr(unsafe.Pointer(memmap.PtrPtr(0x587000, table.Base+28)))
				if titleTestLoad415960(backend, armor) != wantReturn || calls != 2 {
					t.Fatal("native terminal keycell return")
				}
				if *memmap.PtrPtr(0x587000, table.Base) != a || *memmap.PtrPtr(0x587000, table.Base+12) != b || titleTestFrom415960(backend, armor, 2) != b {
					t.Fatal("full localized pointer ownership")
				}
				if !bytes.Equal(packed, before) || !bytes.Equal(guard, wantGuard) {
					t.Fatal("native pointer escaped sidecar or ready DWORD overrun")
				}
				if titleTestLoad415960(backend, armor) != 1 || calls != 2 {
					t.Fatal("ready scalar return/idempotence")
				}
			})
			t.Run(name+"/"+backend+"/loader_next_key_reload", func(t *testing.T) {
				table := titleTestTable415960(armor)
				a := titleTestWide415960(t, "Loaded")
				ka := titleTestKey415960(t, "first")
				kb := titleTestKey415960(t, "removed during service")
				titleTestInstall415960(t, table, []titleTestRow415960{{nil, ka, 1}, {nil, kb, 2}})
				calls := 0
				pop := titlePushObserver415960(&titleObserver415960{Load: func(key unsafe.Pointer, source string, line int32) (unsafe.Pointer, bool) {
					calls++
					if calls != 1 || key != ka {
						t.Fatal("stale next key entered localization service")
					}
					*memmap.PtrPtr(0x587000, table.Base+16) = nil
					return a, true
				}})
				defer pop()
				want := uintptr(unsafe.Pointer(memmap.PtrPtr(0x587000, table.Base+16)))
				if titleTestLoad415960(backend, armor) != want || calls != 1 || *memmap.PtrPtr(0x587000, table.Base+12) != nil {
					t.Fatal("next key reload/early terminal")
				}
			})
			t.Run(name+"/"+backend+"/empty_loader_and_ready_scalar", func(t *testing.T) {
				table := titleTestTable415960(armor)
				titleTestInstall415960(t, table, nil)
				calls := 0
				pop := titlePushObserver415960(&titleObserver415960{Load: func(unsafe.Pointer, string, int32) (unsafe.Pointer, bool) { calls++; return nil, true }})
				defer pop()
				if titleTestLoad415960(backend, armor) != 0 || calls != 0 || *memmap.PtrUint32(0x5D4594, table.Ready) != 1 {
					t.Fatal("empty key sentinel")
				}
				*memmap.PtrUint32(0x5D4594, table.Ready) = 0xfedcba98
				if titleTestLoad415960(backend, armor) != 0xfedcba98 || calls != 0 {
					t.Fatal("mixed scalar return narrowed or made a Go pointer")
				}
			})
		}
	}
}

func TestTitleTablesNativeGoOwners415960(t *testing.T) {
	a := titleTestWide415960(t, "Pinned title")
	q := titleTestWide415960(t, "PINNED TITLE")
	k := titleTestKey415960(t, "different key")
	titleTestInstall415960(t, titleWeapon415960, []titleTestRow415960{{a, k, 0x12345678}})
	outerCalls, innerCalls := 0, 0
	outer := &titleObserver415960{Compare: func(x, y unsafe.Pointer) (int32, bool) {
		outerCalls++
		func() {
			pop := titlePushObserver415960(&titleObserver415960{Compare: func(x, y unsafe.Pointer) (int32, bool) { innerCalls++; return 0, true }})
			defer pop()
			if titleFixtureToMask415960(false, q) != 0x12345678 {
				t.Fatal("nested C->Go lookup")
			}
		}()
		return titleDefaultCompare415960(x, y), true
	}}
	pop := titlePushObserver415960(outer)
	defer pop()
	for i := 0; i < 12; i++ {
		runtime.GC()
		if titleFixtureToMask415960(false, q) != 0x12345678 || titleFixtureFromMask415960(false, 0x12345678) != a {
			t.Fatal("pinned native owner lost through GC/C ABI")
		}
	}
	if outerCalls != 12 || innerCalls != 12 || titleObserverCurrent415960() != outer {
		t.Fatal("scoped observer reentry restoration")
	}
	t.Logf("actual GC cycles=%d nested comparison calls=%d; UTF16 owners explicitly runtime.Pinner-pinned", 12, innerCalls)
}

func TestTitleTablesObserverLIFO415960(t *testing.T) {
	before := titleObserverHead415960
	defer func() { titleObserverHead415960 = before }()
	outer := &titleObserver415960{}
	inner := &titleObserver415960{}
	popOuter := titlePushObserver415960(outer)
	popInner := titlePushObserver415960(inner)
	panicked := false
	func() {
		defer func() { panicked = recover() != nil }()
		popOuter()
	}()
	if !panicked || titleObserverCurrent415960() != inner {
		t.Fatal("out-of-order pop must reject without discarding the current frame")
	}
	popInner()
	if titleObserverCurrent415960() != outer {
		t.Fatal("inner pop lost outer frame")
	}
	popOuter()
	if titleObserverHead415960 != before {
		t.Fatal("observer stack did not restore")
	}
}
