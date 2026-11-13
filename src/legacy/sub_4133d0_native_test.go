//go:build darwin && arm64

package legacy

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func material7TestCall4133D0(backend string, item unsafe.Pointer) int32 {
	if backend == "Go" {
		return rawMaterial7Native4133D0(item)
	}
	return material7FixtureCall4133D0(item)
}

func TestMaterial7NativeLayout4133D0(t *testing.T) {
	goOffsets, cOffsets := material7Offsets4133D0(), material7FixtureCOffsets4133D0()
	if goOffsets[0] != cOffsets[0] || goOffsets[1] != cOffsets[1] || goOffsets[2] != unsafe.Offsetof(server.ModifierInitData{}.Modifiers)+unsafe.Sizeof((*server.ModifierEff)(nil)) {
		t.Fatalf("actual named C/Go InitData/Class/material-slot layout: Go=%v C=%v", goOffsets, cOffsets)
	}
	if unsafe.Sizeof(unsafe.Pointer(nil)) != 8 {
		t.Fatal("this local native-pointer control is Darwin ARM64")
	}
	t.Logf("MATERIAL7_ACTUAL_LAYOUT_OK InitData=%d ObjClass=%d slot1=%d pointer-width=8 typed-C-root=true", goOffsets[0], goOffsets[1], goOffsets[2])
}

func TestMaterial7NativeIdentityAndOrder4133D0(t *testing.T) {
	controls := []string{"warm-match", "same-name-other-pointer", "ineligible-lazy-slot", "cold-match", "cold-nil-retry", "eager-init-cache", "warm-cache-load-not-rechecked", "cold-cache-load-not-rechecked"}
	for _, backend := range []string{"Go", "C"} {
		for _, control := range controls {
			t.Run(backend+"/"+control, func(t *testing.T) {
				item, freeItem := alloc.NewHigh(server.Object{})
				defer freeItem()
				first, freeFirst := alloc.NewHigh(server.ModifierInitData{})
				defer freeFirst()
				second, freeSecond := alloc.NewHigh(server.ModifierInitData{})
				defer freeSecond()
				descriptor, freeDescriptor := alloc.NewHigh(server.ModifierEff{})
				defer freeDescriptor()
				other, freeOther := alloc.NewHigh(server.ModifierEff{})
				defer freeOther()
				cache, freeCache := alloc.NewHigh((*server.ModifierEff)(nil))
				defer freeCache()
				name, freeName := alloc.CStringHigh("Material7")
				defer freeName()
				// Name is the first native pointer field. The raw predicate does not
				// inspect it; equal text must not turn two descriptors into one identity.
				*(*unsafe.Pointer)(unsafe.Pointer(descriptor)) = unsafe.Pointer(name)
				*(*unsafe.Pointer)(unsafe.Pointer(other)) = unsafe.Pointer(name)
				if descriptor.Name() != "Material7" || other.Name() != "Material7" || descriptor == other {
					t.Fatal("two actual descriptors with the same name")
				}
				for _, p := range []unsafe.Pointer{unsafe.Pointer(item), unsafe.Pointer(first), unsafe.Pointer(second), unsafe.Pointer(descriptor), unsafe.Pointer(other), unsafe.Pointer(cache)} {
					if uintptr(p) <= 0xffffffff {
						t.Fatalf("actual C graph must exceed4GiB: %p", p)
					}
				}
				item.ObjClass = 0x10000000
				item.InitData = unsafe.Pointer(first)
				first.Modifiers[1], second.Modifiers[1] = descriptor, other
				*cache = descriptor
				wanted, calls := int32(1), 1
				cold := control == "cold-match" || control == "cold-nil-retry" || control == "eager-init-cache" || control == "cold-cache-load-not-rechecked"
				if cold {
					*cache = nil
				}
				switch control {
				case "same-name-other-pointer":
					first.Modifiers[1] = other
					wanted = 0
				case "ineligible-lazy-slot":
					item.ObjClass = 0
					item.InitData = nil
					wanted = 0
				case "cold-nil-retry":
					first.Modifiers[1] = nil
					calls = 2
				}
				var actual []string
				idCalls, descCalls := 0, 0
				observer := &material7Observer4133D0{Cache: cache}
				observer.Operand = func(event material7Operand4133D0) {
					if event.Base == nil || uintptr(event.Address) != uintptr(event.Base)+event.Offset {
						t.Fatal("actual native operand address metadata")
					}
					if !event.Complete {
						if !event.Write && (event.ValueValid || event.Pointer != nil || event.Word != 0) {
							t.Fatal("before-read invented a loaded value")
						}
						return
					}
					if !event.ValueValid {
						t.Fatal("completed actual operand lacks its value")
					}
					if event.Field == 2 {
						if event.Width != 4 || event.IsPointer {
							t.Fatal("actual DWORD class read width")
						}
					} else if event.Width != 8 || !event.IsPointer {
						t.Fatal("actual native-width pointer operand")
					}
					if event.Write && (event.Field != 0 || event.Address != unsafe.Pointer(cache) || *(*unsafe.Pointer)(unsafe.Pointer(cache)) != event.Pointer) {
						t.Fatal("completed actual cache pointer store")
					}
					kind := "read"
					if event.Write {
						kind = "write"
					}
					actual = append(actual, fmt.Sprintf("%s%d", kind, event.Field))
					if event.Field == 3 && event.Base != unsafe.Pointer(first) {
						t.Fatal("original eager InitData local was replaced by a later object-field value")
					}
					if event.Field == 0 && !event.Write {
						if control == "warm-cache-load-not-rechecked" {
							*cache = nil
						}
						if control == "cold-cache-load-not-rechecked" {
							*cache = other
						}
					}
				}
				observer.ResolveID = func(got string) int32 {
					idCalls++
					if !cold || got != "Material7" {
						t.Fatal("unexpected named lookup", got)
					}
					if control == "eager-init-cache" {
						item.InitData = unsafe.Pointer(second)
					}
					return -2147483647
				}
				observer.ResolveDesc = func(id int32) *server.ModifierEff {
					descCalls++
					if id != -2147483647 {
						t.Fatal("full signed lookup id changed", id)
					}
					if control == "cold-nil-retry" {
						return nil
					}
					return descriptor
				}
				pop := pushMaterial7Observer4133D0(observer)
				defer pop()
				for i := 0; i < calls; i++ {
					if got := material7TestCall4133D0(backend, unsafe.Pointer(item)); got != wanted {
						t.Fatalf("raw identity=%d expected source-control=%d", got, wanted)
					}
				}
				if idCalls != descCalls || idCalls != (map[bool]int{false: 0, true: calls})[cold] {
					t.Fatalf("external lookup invocations id=%d desc=%d", idCalls, descCalls)
				}
				one := []string{"read0", "read1"}
				if cold {
					one = append(one, "write0")
				}
				one = append(one, "read2")
				if control != "ineligible-lazy-slot" {
					one = append(one, "read3")
				}
				var expected []string
				for i := 0; i < calls; i++ {
					expected = append(expected, one...)
				}
				if !reflect.DeepEqual(actual, expected) {
					t.Fatalf("actual cached/eager/lazy operand order %v expected%v", actual, expected)
				}
				if control == "cold-nil-retry" && *cache != nil {
					t.Fatal("nil resolver result was incorrectly memoized")
				}
				t.Logf("MATERIAL7_NATIVE_IDENTITY_OK backend=%s control=%s actual-calls=%d pointer-width=8 original-helper-body-credit=0", backend, control, calls)
			})
		}
	}
}

func TestMaterial7NativeOpaqueWarm4133D0(t *testing.T) {
	for _, backend := range []string{"Go", "C"} {
		t.Run(backend, func(t *testing.T) {
			item, freeItem := alloc.NewHigh(server.Object{})
			defer freeItem()
			attributes, freeAttributes := alloc.NewHigh(server.ModifierInitData{})
			defer freeAttributes()
			cache, freeCache := alloc.NewHigh((*server.ModifierEff)(nil))
			defer freeCache()
			buffer, freeBuffer := alloc.MallocHigh(32)
			defer freeBuffer()
			raw := unsafe.Slice((*byte)(buffer), 32)
			for i := range raw {
				raw[i] = byte(0x31 + i*7)
			}
			before := bytes.Clone(raw)
			opaque := unsafe.Add(buffer, 1)
			nonmatch := unsafe.Add(buffer, 3)
			if uintptr(opaque) <= 0xffffffff || uintptr(opaque)%8 == 0 {
				t.Fatal("actual opaque C pointer above4GiB and unaligned")
			}
			*(*unsafe.Pointer)(unsafe.Pointer(cache)) = opaque
			item.ObjClass = 0x10000000
			item.InitData = unsafe.Pointer(attributes)
			slot := unsafe.Add(unsafe.Pointer(attributes), material7Offsets4133D0()[2])
			*(*unsafe.Pointer)(slot) = opaque
			pop := pushMaterial7Observer4133D0(&material7Observer4133D0{Cache: cache,
				ResolveID: func(string) int32 { t.Fatal("warm opaque identity unexpectedly called ID service"); return 0 },
				ResolveDesc: func(int32) *server.ModifierEff {
					t.Fatal("warm opaque identity unexpectedly called descriptor service")
					return nil
				}})
			defer pop()
			if got := material7TestCall4133D0(backend, unsafe.Pointer(item)); got != 1 {
				t.Fatalf("actual opaque identity match=%d", got)
			}
			*(*unsafe.Pointer)(slot) = nonmatch
			if got := material7TestCall4133D0(backend, unsafe.Pointer(item)); got != 0 {
				t.Fatalf("actual distinct opaque identity=%d", got)
			}
			if !bytes.Equal(raw, before) {
				t.Fatal("opaque descriptor bytes were dereferenced/written")
			}
			t.Logf("MATERIAL7_OPAQUE_WARM_OK backend=%s cached=%p match=true distinct=false no-typed-descriptor-cast=true", backend, opaque)
		})
	}
}

func TestMaterial7NativePinnedGo4133D0(t *testing.T) {
	for _, backend := range []string{"Go", "C"} {
		t.Run(backend, func(t *testing.T) {
			item := new(server.Object)
			attributes := new(server.ModifierInitData)
			descriptor := new(server.ModifierEff)
			cache := new(*server.ModifierEff)
			for _, p := range []unsafe.Pointer{unsafe.Pointer(item), unsafe.Pointer(attributes), unsafe.Pointer(descriptor), unsafe.Pointer(cache)} {
				if uintptr(p) <= 0xffffffff {
					t.Fatalf("actual high Go allocation: %p", p)
				}
			}
			item.ObjClass = 0x10000000
			item.InitData = unsafe.Pointer(attributes)
			attributes.Modifiers[1] = descriptor
			*cache = descriptor
			var pin runtime.Pinner
			pin.Pin(item)
			pin.Pin(attributes)
			pin.Pin(descriptor)
			pin.Pin(cache)
			defer pin.Unpin()
			GCs := 0
			pop := pushMaterial7Observer4133D0(&material7Observer4133D0{Cache: cache,
				Operand: func(e material7Operand4133D0) {
					if e.Complete {
						runtime.GC()
						GCs++
					}
				},
				ResolveID:   func(string) int32 { t.Fatal("warm pinned graph called lookup"); return 0 },
				ResolveDesc: func(int32) *server.ModifierEff { t.Fatal("warm pinned graph called descriptor lookup"); return nil }})
			defer pop()
			if got := material7TestCall4133D0(backend, unsafe.Pointer(item)); got != 1 || GCs != 4 {
				t.Fatalf("pinned graph rawresult/forcedGC=%d/%d", got, GCs)
			}
			runtime.KeepAlive(item)
			runtime.KeepAlive(attributes)
			runtime.KeepAlive(descriptor)
			runtime.KeepAlive(cache)
			t.Logf("MATERIAL7_PINNED_GO_OK backend=%s individually-pinned-allocations=4 forced-in-call-GC=4 cgo-check-bypass=false", backend)
		})
	}
}

type material7TestServer4133D0 struct {
	Server
	native *server.Server
}

func (s *material7TestServer4133D0) S() *server.Server { return s.native }

func TestMaterial7NativeDefaultCold4133D0(t *testing.T) {
	if material7ObserverStack4133D0 != nil {
		t.Fatal("default product-service control requires nil observer")
	}
	for _, backend := range []string{"Go", "C"} {
		t.Run(backend, func(t *testing.T) {
			srv := new(server.Server)
			facade := &material7TestServer4133D0{native: srv}
			previous := GetServer
			GetServer = func() Server { return facade }
			defer func() { GetServer = previous }()
			item, freeItem := alloc.NewHigh(server.Object{})
			defer freeItem()
			attributes, freeAttributes := alloc.NewHigh(server.ModifierInitData{})
			defer freeAttributes()
			item.ObjClass = 0x10000000
			item.InitData = unsafe.Pointer(attributes)
			if srv.Modif.Nox_xxx_modifGetIdByName413290("Material7") != 255 || srv.Modif.Nox_xxx_modifGetDescById413330(255) != nil {
				t.Fatal("actual empty-server modifier services ID255/descriptornil")
			}
			cell := srv.Modif.Material7CacheCell4133D0()
			for repeat := 0; repeat < 2; repeat++ {
				if *cell != nil || material7TestCall4133D0(backend, unsafe.Pointer(item)) != 1 || *cell != nil {
					t.Fatal("actual nil-observer services must retry nil descriptor and compare eligible nil slot equal")
				}
			}
			t.Logf("MATERIAL7_DEFAULT_COLD_OK backend=%s observer=nil actual-empty-server-id=255 actual-desc=nil calls=2 cold-retry=true original-helper-body-credit=0", backend)
		})
	}
}
