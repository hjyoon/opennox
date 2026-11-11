package legacy

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"unicode/utf16"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type accessControlOp416720 struct {
	Kind      string  `json:"kind"`
	Value     int32   `json:"value"`
	Name      string  `json:"name"`
	Serial    *string `json:"serial"`
	Clock     uint32  `json:"clock"`
	PointerID int     `json:"pointer_id"`
}
type accessControlCase416720 struct {
	Name string                  `json:"name"`
	Ops  []accessControlOp416720 `json:"ops"`
}
type accessControlNode416720 struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Expires uint64 `json:"expires"`
	Serial  string `json:"serial"`
}
type accessControlStep416720 struct {
	ReturnRole int                       `json:"return_role"`
	Word       uint32                    `json:"word"`
	Allowed    []accessControlNode416720 `json:"allowed"`
	Banned     []accessControlNode416720 `json:"banned"`
	Events     []string                  `json:"events"`
}
type accessControlRow416720 struct {
	Name  string                    `json:"name"`
	Steps []accessControlStep416720 `json:"steps"`
}
type accessControlDocument416720 struct {
	Schema       string                   `json:"schema"`
	PointerBytes int                      `json:"pointer_bytes"`
	Rows         []accessControlRow416720 `json:"rows"`
}

func accessLoadControl416720(t *testing.T) ([]accessControlCase416720, accessControlDocument416720) {
	t.Helper()
	var in struct {
		Cases []accessControlCase416720 `json:"cases"`
	}
	p := filepath.Join("testdata", "access_list_416720", "inputs-v1.json")
	b, e := os.ReadFile(p)
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(b, &in); e != nil {
		t.Fatal(e)
	}
	p = os.Getenv("OPENNOX_ACCESS_LIST_BEFORE_C_REFERENCE")
	if p == "" {
		p = filepath.Join("testdata", "access_list_416720", "before-C-reference-v1.json")
	}
	b, e = os.ReadFile(p)
	if e != nil {
		t.Fatalf("root must record preserved pre-port C controls first: %v", e)
	}
	// Reference records the qualified64-bit old-C host; pointer roles and
	// UTF16/name/expiry data are logical, not current native slot bytes.
	// TestAccessListNativeCLayout416720 independently checks current Go/C layout.
	var doc accessControlDocument416720
	if e = json.Unmarshal(b, &doc); e != nil {
		t.Fatal(e)
	}
	if doc.Schema != "opennox-access-list-preserved-C-control-v1" || doc.PointerBytes != 8 || len(doc.Rows) != len(in.Cases) {
		t.Fatal("wrong preserved-C control domain/layout")
	}
	return in.Cases, doc
}
func accessNameHex416720(name [26]uint16) string {
	var b [52]byte
	for i, v := range name {
		b[2*i] = byte(v)
		b[2*i+1] = byte(v >> 8)
	}
	return hex.EncodeToString(b[:])
}
func TestAccessListNativeCLayout416720(t *testing.T) {
	var a accessAllowed416720
	var b accessBanned416720
	got := [6]uintptr{unsafe.Sizeof(a.List), unsafe.Sizeof(a), unsafe.Sizeof(b), unsafe.Offsetof(a.Name), unsafe.Offsetof(b.Expires), unsafe.Offsetof(b.Serial)}
	if want := accessListNativeCLayout416720(); got != want {
		t.Fatalf("named Go/C layout got=%v C=%v", got, want)
	}
}
func TestAccessListNativePreservedC416720(t *testing.T) {
	cases, doc := accessLoadControl416720(t)
	for _, backend := range []string{"Go", "C"} {
		t.Run(backend, func(t *testing.T) {
			for i, item := range cases {
				t.Run(item.Name, func(t *testing.T) {
					if len(doc.Rows[i].Steps) != len(item.Ops) {
						t.Fatal("reference step count")
					}
					if doc.Rows[i].Name != item.Name {
						t.Fatal("reference case order mismatch")
					}
					var ah, bh accessListNode416720
					var headPins runtime.Pinner
					headPins.Pin(&ah)
					headPins.Pin(&bh)
					defer headPins.Unpin()
					accessListNativeClear416720(&ah)
					accessListNativeClear416720(&bh)
					ids := map[unsafe.Pointer]int{}
					var owners []unsafe.Pointer
					freed := map[unsafe.Pointer]bool{}
					clock := uint32(0)
					failAlloc := false
					events := []string{}
					refresh := accessListNativeFixtureAlloc416720(4)
					if refresh == nil {
						t.Fatal("refresh anchor calloc")
					}
					defer accessListNativeFixtureFree416720(refresh)
					role := func(p unsafe.Pointer) int {
						if p == nil {
							return 0
						}
						if p == refresh {
							return -1
						}
						v, ok := ids[p]
						if !ok {
							t.Fatalf("unknown full pointer %p", p)
						}
						return v
					}
					h := &accessListNativeHooks416720{Allowed: &ah, Banned: &bh}
					h.Allocate = func(size uintptr) unsafe.Pointer {
						if failAlloc {
							failAlloc = false
							events = append(events, "alloc-fail")
							return nil
						}
						p := accessListNativeFixtureAlloc416720(size)
						if p == nil {
							t.Fatal("actual C allocation failed")
						}
						if uintptr(p) <= 0xffffffff && unsafe.Sizeof(p) == 8 {
							t.Fatal("expected high native C allocation")
						}
						owners = append(owners, p)
						freed[p] = false
						ids[p] = len(owners)
						events = append(events, fmt.Sprintf("alloc:%d", ids[p]))
						return p
					}
					h.Free = func(p unsafe.Pointer) {
						n := (*accessListNode416720)(p)
						if n.Next != n || n.Prev != n {
							t.Fatal("unlink must precede free")
						}
						if freed[p] {
							t.Fatal("double free")
						}
						events = append(events, fmt.Sprintf("free:%d", role(p)))
						freed[p] = true
						accessListNativeFixtureFree416720(p)
					}
					h.Clock = func() uint32 { events = append(events, fmt.Sprintf("clock:%d", clock)); return clock }
					h.Refresh = func() unsafe.Pointer { events = append(events, "refresh"); return refresh }
					h.Load = func(p unsafe.Pointer) int32 {
						if p != accessListNativeConfig416720() {
							t.Fatal("actual config address")
						}
						events = append(events, "load")
						return -2147483647
					}
					h.Save = func(s string) { events = append(events, "save:"+s) }
					pop := pushAccessListNativeHooks416720(h)
					defer pop()
					defer func() {
						for _, p := range owners {
							if !freed[p] {
								freed[p] = true
								accessListNativeFixtureFree416720(p)
							}
						}
					}()
					snapshot := func(p unsafe.Pointer, word uint32) accessControlStep416720 {
						out := accessControlStep416720{ReturnRole: role(p), Word: word, Allowed: []accessControlNode416720{}, Banned: []accessControlNode416720{}, Events: append([]string{}, events...)}
						for n := accessListNativeNext416720(&ah); n != nil; n = accessListNativeNext416720(n) {
							a := (*accessAllowed416720)(unsafe.Pointer(n))
							out.Allowed = append(out.Allowed, accessControlNode416720{ID: role(unsafe.Pointer(n)), Name: accessNameHex416720(a.Name)})
						}
						for n := accessListNativeNext416720(&bh); n != nil; n = accessListNativeNext416720(n) {
							b := (*accessBanned416720)(unsafe.Pointer(n))
							out.Banned = append(out.Banned, accessControlNode416720{ID: role(unsafe.Pointer(n)), Name: accessNameHex416720(b.Name), Expires: b.Expires, Serial: hex.EncodeToString(b.Serial[:])})
						}
						return out
					}
					for j, op := range item.Ops {
						clock = op.Clock
						events = []string{}
						entry := map[string]uint32{"allow": 0x4168a0, "ban": 0x416770, "remove-allow": 0x416860, "remove-ban": 0x416820, "expire": 0x416720, "reset": 0x416920, "clear": 0x416950, "first-allow": 0x4168e0, "first-ban": 0x416900, "next-allow": 0x4168f0, "next-ban": 0x416910}[op.Kind]
						if op.Kind == "fail-next-allocation" {
							failAlloc = true
							entry = 0
						}
						units := utf16.Encode([]rune(op.Name))
						units = append(units, 0)
						name, freeName := alloc.MakeHigh([]uint16{}, len(units))
						copy(name, units)
						np := unsafe.Pointer(&name[0])
						var sp unsafe.Pointer
						freeSerial := func() {}
						if op.Serial != nil {
							b := append([]byte(*op.Serial), 0)
							arr, free := alloc.MakeHigh([]byte{}, len(b))
							copy(arr, b)
							sp = unsafe.Pointer(&arr[0])
							freeSerial = free
						}
						if op.Kind == "next-allow" || op.Kind == "next-ban" {
							np = nil
							if op.PointerID > 0 {
								if op.PointerID > len(owners) {
									t.Fatal("literal input pointer role")
								}
								np = owners[op.PointerID-1]
							}
						}
						var p unsafe.Pointer
						var word uint32
						if entry != 0 {
							if backend == "C" {
								p, word = accessListFixtureCall416720(entry, op.Value, np, sp)
							} else {
								switch entry {
								case 0x416720:
									sub416720Native()
								case 0x416770:
									p = sub416770Native(op.Value, np, sp)
								case 0x416820:
									sub416820Native(op.Value)
								case 0x416860:
									p = sub416860Native(op.Value)
								case 0x4168a0:
									p = sub4168A0Native(np)
								case 0x4168e0:
									p = sub4168E0Native()
								case 0x4168f0:
									p = sub4168F0Native(np)
								case 0x416900:
									p = sub416900Native()
								case 0x416910:
									p = sub416910Native(np)
								case 0x416920:
									word = uint32(sub416920Native())
								case 0x416950:
									p = sub416950Native()
								}
							}
						}
						freeSerial()
						freeName()
						got := snapshot(p, word)
						if !reflect.DeepEqual(got, doc.Rows[i].Steps[j]) {
							t.Fatalf("step%d %s got=%+v oldC=%+v", j, op.Kind, got, doc.Rows[i].Steps[j])
						}
					}
				})
			}
		})
	}
}
func TestAccessListNativePinnedCachedNext416720(t *testing.T) {
	// Actual typed Go owners are independently pinned before public C entry.
	// Free is observed, not passed to C.free for these Go allocations.
	for _, backend := range []string{"Go", "C"} {
		t.Run(backend, func(t *testing.T) {
			ah, bh := new(accessListNode416720), new(accessListNode416720)
			a, b, c := new(accessAllowed416720), new(accessAllowed416720), new(accessBanned416720)
			var pin runtime.Pinner
			for _, p := range []any{ah, bh, a, b, c} {
				pin.Pin(p)
			}
			defer pin.Unpin()
			accessListNativeClear416720(ah)
			accessListNativeClear416720(bh)
			accessListNativeInit416720(&a.List)
			accessListNativeInit416720(&b.List)
			accessListNativeInit416720(&c.List)
			accessListNativeAppend416720(ah, &a.List)
			accessListNativeAppend416720(ah, &b.List)
			accessListNativeAppend416720(bh, &c.List)
			freed := []unsafe.Pointer{}
			h := &accessListNativeHooks416720{Allowed: ah, Banned: bh, Save: func(s string) {
				if s != "ban.txt" {
					t.Fatal(s)
				}
				runtime.GC()
			}, Free: func(p unsafe.Pointer) { runtime.GC(); freed = append(freed, p) }}
			pop := pushAccessListNativeHooks416720(h)
			defer pop()
			var got unsafe.Pointer
			if backend == "C" {
				got, _ = accessListFixtureCall416720(0x416950, 0, nil, nil)
			} else {
				got = sub416950Native()
			}
			if got != unsafe.Pointer(c) || !reflect.DeepEqual(freed, []unsafe.Pointer{unsafe.Pointer(a), unsafe.Pointer(b), unsafe.Pointer(c)}) || ah.Next != ah || bh.Next != bh {
				t.Fatal("cached next/free order/full opaque return")
			}
			runtime.KeepAlive([]any{ah, bh, a, b, c})
		})
	}
}
func TestAccessListNativeProductionOwner416720(t *testing.T) {
	ah := (*accessListNode416720)(accessListNativeDefaultHead416720(false))
	bh := (*accessListNode416720)(accessListNativeDefaultHead416720(true))
	oldA, oldB := *ah, *bh
	defer func() { *ah = oldA; *bh = oldB }()
	accessListNativeClear416720(ah)
	accessListNativeClear416720(bh)
	p := accessListNativeFixtureAlloc416720(unsafe.Sizeof(accessAllowed416720{}))
	if p == nil {
		t.Fatal("calloc")
	}
	defer accessListNativeFixtureFree416720(p)
	n := (*accessAllowed416720)(p)
	accessListNativeInit416720(&n.List)
	accessListNativeAppend416720(ah, &n.List)
	if sub4168E0Native() != p || Nox_server_accessAllowedFirst_4168E0() != p || sub4168F0Native(p) != nil || Nox_server_accessAllowedNext_4168F0(p) != nil || sub416900Native() != nil {
		t.Fatal("actual C owner and existing consumer identity")
	}
}
