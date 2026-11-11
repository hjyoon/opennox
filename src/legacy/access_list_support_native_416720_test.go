package legacy

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"runtime"
	"testing"
	"unicode/utf16"
	"unsafe"
)

func TestAccessListSupportCFieldAddresses416720(t *testing.T) {
	var allowed accessAllowed416720
	var banned accessBanned416720
	for _, item := range []struct {
		name       string
		field      int
		banned     bool
		goOffset   uintptr
		width      int
		last       int
		firstValue uint32
		lastValue  uint32
	}{
		{"allowed-name", 0, false, unsafe.Offsetof(allowed.Name), 2, 25, 0x9ABC, 0xFEDC},
		{"banned-name", 1, true, unsafe.Offsetof(banned.Name), 2, 25, 0x8FED, 0xCDEF},
		{"banned-serial", 2, true, unsafe.Offsetof(banned.Serial), 1, 23, 0xEA, 0xD7},
	} {
		t.Run(item.name, func(t *testing.T) {
			if accessSupportFixtureField416720(item.field, nil) != nil {
				t.Fatal("public C getter must preserve nil")
			}
			offset := accessSupportFixtureOffset416720(item.field)
			if offset != item.goOffset {
				t.Fatalf("named C/Go field offset: C=%d Go=%d", offset, item.goOffset)
			}
			const guard = 16
			size := accessSupportFixtureSize416720(item.banned)
			base := accessListNativeFixtureAlloc416720(size + 2*guard)
			if base == nil {
				t.Fatal("C entry allocation")
			}
			defer accessListNativeFixtureFree416720(base)
			storage := unsafe.Slice((*byte)(base), int(size)+2*guard)
			copy(storage, bytes.Repeat([]byte{0xA7}, len(storage)))
			want := append([]byte(nil), storage...)
			entry := unsafe.Add(base, guard)
			field := accessSupportFixtureField416720(item.field, entry)
			if field != unsafe.Add(entry, offset) {
				t.Fatalf("public getter must alias the complete native field: entry=%p field=%p", entry, field)
			}
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(entry) <= 0xFFFFFFFF {
				t.Fatal("control needs a C entry above4GiB")
			}
			accessSupportFixtureStore416720(item.field, entry, 0, item.firstValue)
			last := unsafe.Add(field, uintptr(item.width*item.last))
			if item.width == 2 {
				if *(*uint16)(field) != uint16(item.firstValue) {
					t.Fatal("C named-field write did not reach the returned UTF16 address")
				}
				*(*uint16)(last) = uint16(item.lastValue)
				binary.NativeEndian.PutUint16(want[guard+int(offset):], uint16(item.firstValue))
				binary.NativeEndian.PutUint16(want[guard+int(offset)+item.width*item.last:], uint16(item.lastValue))
			} else {
				if *(*byte)(field) != byte(item.firstValue) {
					t.Fatal("C named-field write did not reach the returned serial address")
				}
				*(*byte)(last) = byte(item.lastValue)
				want[guard+int(offset)] = byte(item.firstValue)
				want[guard+int(offset)+item.last] = byte(item.lastValue)
			}
			for i := 0; i < 3; i++ {
				runtime.GC()
				if accessSupportFixtureField416720(item.field, entry) != field {
					t.Fatal("public field address changed after GC")
				}
				if accessSupportFixtureRead416720(item.field, entry, item.last) != item.lastValue {
					t.Fatal("Go alias write did not reach the final named C field element")
				}
				if !bytes.Equal(storage, want) {
					t.Fatal("field alias writes changed surrounding fields, padding or allocation canaries")
				}
			}
		})
	}
}

func TestAccessListSupportDefaultHeadsAndGC416720(t *testing.T) {
	if accessListNativeCurrent416720() != nil {
		t.Fatal("test requires actual default owners")
	}
	ah := (*accessListNode416720)(accessSupportFixtureHead416720(0))
	bh := (*accessListNode416720)(accessSupportFixtureHead416720(1))
	if ah == nil || bh == nil || ah == bh {
		t.Fatal("independent non-nil native C head owners")
	}
	if unsafe.Pointer(ah) != accessSupportFixtureOwner416720(false) || unsafe.Pointer(bh) != accessSupportFixtureOwner416720(true) {
		t.Fatal("public head must preserve each actual C data symbol address")
	}
	for _, head := range []*accessListNode416720{ah, bh} {
		p := unsafe.Pointer(head)
		links := accessSupportFixtureLinks416720(p)
		if links[0] == nil || links[1] == nil || links[2] != p {
			t.Fatal("head accessor must initialize the native sentinel before returning")
		}
	}
	oldA, oldB := *ah, *bh
	defer func() { *ah = oldA; *bh = oldB }()
	accessListNativeClear416720(ah)
	accessListNativeClear416720(bh)
	for i := 0; i < 3; i++ {
		runtime.GC()
		if accessSupportFixtureHead416720(0) != unsafe.Pointer(ah) || accessListNativeDefaultHead416720(false) != unsafe.Pointer(ah) {
			t.Fatal("allowed head identity changed")
		}
		for _, selector := range []int32{1, -1, 0x7FFFFFFF, -0x80000000} {
			if accessSupportFixtureHead416720(selector) != unsafe.Pointer(bh) {
				t.Fatal("nonzero banned selector changed native owner identity")
			}
		}
		for _, head := range []*accessListNode416720{ah, bh} {
			p := unsafe.Pointer(head)
			if got := accessSupportFixtureLinks416720(p); got != [3]unsafe.Pointer{p, p, p} {
				t.Fatalf("C-observed native sentinel links: %v", got)
			}
		}
		if sub4168E0Native() != nil || sub416900Native() != nil {
			t.Fatal("existing Go list functions must recognize the actual empty sentinels")
		}
	}
}

func TestAccessListSupportExistingListInterop416720(t *testing.T) {
	for _, backend := range []string{"Go", "C"} {
		t.Run(backend, func(t *testing.T) {
			if accessListNativeCurrent416720() != nil {
				t.Fatal("test requires actual default owners")
			}
			ah := (*accessListNode416720)(accessSupportFixtureHead416720(0))
			bh := (*accessListNode416720)(accessSupportFixtureHead416720(1))
			if ah == nil || bh == nil || ah == bh {
				t.Fatal("default native owners")
			}
			oldA, oldB := *ah, *bh
			accessListNativeClear416720(ah)
			accessListNativeClear416720(bh)
			var allocations, freed []unsafe.Pointer
			dead := map[unsafe.Pointer]bool{}
			refreshes, saves := 0, 0
			h := &accessListNativeHooks416720{
				// Allowed/Banned stay nil: every list operation uses the real C owners.
				Allocate: func(size uintptr) unsafe.Pointer {
					p := accessListNativeFixtureAlloc416720(size)
					if p == nil {
						t.Fatal("C-owned entry allocation")
					}
					allocations = append(allocations, p)
					return p
				},
				Free: func(p unsafe.Pointer) {
					known := false
					for _, q := range allocations {
						known = known || p == q
					}
					if !known || dead[p] {
						t.Fatal("only this control's live C allocations may be freed")
					}
					links := accessSupportFixtureLinks416720(p)
					if links[0] != p || links[1] != p {
						t.Fatal("entry must be unlinked before Free")
					}
					freed = append(freed, p)
					dead[p] = true
					accessListNativeFixtureFree416720(p)
				},
				Refresh: func() unsafe.Pointer { refreshes++; return nil },
				Save: func(path string) {
					if path != "ban.txt" {
						t.Fatal("save service argument")
					}
					saves++
				},
			}
			pop := pushAccessListNativeHooks416720(h)
			defer func() {
				// Restore actual roots before releasing any surviving test allocations.
				*ah = oldA
				*bh = oldB
				pop()
				for _, p := range allocations {
					if !dead[p] {
						accessListNativeFixtureFree416720(p)
					}
				}
			}()
			name := func(s string) unsafe.Pointer {
				units := append(utf16.Encode([]rune(s)), 0)
				p := accessListNativeFixtureAlloc416720(uintptr(len(units) * 2))
				if p == nil {
					t.Fatal("C name allocation")
				}
				copy(unsafe.Slice((*uint16)(p), len(units)), units)
				t.Cleanup(func() { accessListNativeFixtureFree416720(p) })
				return p
			}
			firstName, secondName, bannedName := name("Alpha"), name("Omega Ω"), name("Banned 中")
			serialBytes := append([]byte("serial-0001"), 0)
			serial := accessListNativeFixtureAlloc416720(uintptr(len(serialBytes)))
			if serial == nil {
				t.Fatal("C serial allocation")
			}
			defer accessListNativeFixtureFree416720(serial)
			copy(unsafe.Slice((*byte)(serial), len(serialBytes)), serialBytes)
			for _, p := range []unsafe.Pointer{firstName, secondName} {
				var result unsafe.Pointer
				if backend == "Go" {
					result = sub4168A0Native(p)
				} else {
					result, _ = accessListFixtureCall416720(0x4168A0, 0, p, nil)
				}
				if result != nil {
					t.Fatal("add must propagate the isolated Refresh return")
				}
			}
			if backend == "Go" {
				sub416770Native(0, bannedName, serial)
			} else {
				accessListFixtureCall416720(0x416770, 0, bannedName, serial)
			}
			if len(allocations) != 3 || refreshes != 3 {
				t.Fatal("two allowed entries and one banned entry must be allocated and refreshed")
			}
			a, b, ban := allocations[0], allocations[1], allocations[2]
			for i := 0; i < 3; i++ {
				runtime.GC()
				if accessSupportFixtureHead416720(0) != unsafe.Pointer(ah) || accessSupportFixtureHead416720(1) != unsafe.Pointer(bh) {
					t.Fatal("head lookup/GC must not reset or replace a populated owner")
				}
				if sub4168E0Native() != a || Nox_server_accessAllowedFirst_4168E0() != a || sub4168F0Native(a) != b || Nox_server_accessAllowedNext_4168F0(a) != b || Nox_server_accessAllowedNext_4168F0(b) != nil {
					t.Fatal("Go and public C traversal must share the full allowed links")
				}
				if sub416900Native() != ban || Nox_server_accessBannedFirst_416900() != ban || Nox_server_accessBannedNext_416910(ban) != nil {
					t.Fatal("Go and public C traversal must share the banned link")
				}
				if Nox_server_accessAllowedName_4168E0(a) != "Alpha" || Nox_server_accessAllowedName_4168E0(b) != "Omega Ω" || Nox_server_accessBannedName_416900(ban) != "Banned 中" || Nox_server_accessBannedSerial_416900(ban) != "serial-0001" {
					t.Fatal("existing consumers must use the same native name/serial fields")
				}
			}
			accessSupportFixtureStore416720(0, a, 0, 'Z')
			accessSupportFixtureStore416720(2, ban, 0, 'X')
			if Nox_server_accessAllowedName_4168E0(a) != "Zlpha" || Nox_server_accessBannedSerial_416900(ban) != "Xerial-0001" {
				t.Fatal("named C mutations must be visible through existing Go consumers")
			}
			// Remove through the other entry path, keeping the actual sentinel owner.
			var removed unsafe.Pointer
			if backend == "Go" {
				removed, _ = accessListFixtureCall416720(0x416860, 0, nil, nil)
			} else {
				removed = sub416860Native(0)
			}
			if removed != a || Nox_server_accessAllowedFirst_4168E0() != b || Nox_server_accessAllowedNext_4168F0(b) != nil {
				t.Fatal("cross-entry removal must preserve the remaining native links")
			}
			// removed is an opaque identity only; it is never dereferenced after Free.
			if backend == "Go" {
				accessListFixtureCall416720(0x416950, 0, nil, nil)
			} else {
				sub416950Native()
			}
			if saves != 1 || !reflect.DeepEqual(freed, []unsafe.Pointer{a, b, ban}) {
				t.Fatal("existing clear/save/free order across actual C owners")
			}
			for _, head := range []*accessListNode416720{ah, bh} {
				p := unsafe.Pointer(head)
				if accessSupportFixtureLinks416720(p) != [3]unsafe.Pointer{p, p, p} {
					t.Fatal("clear must restore C-observed sentinel self-links")
				}
			}
		})
	}
}
