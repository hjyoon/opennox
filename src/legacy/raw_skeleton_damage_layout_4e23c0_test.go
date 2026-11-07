//go:build darwin && arm64

package legacy

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestRawSkeletonNativeLayout4E23C0(t *testing.T) {
	offsets, cOffsets := rawSkeletonOffsets4E23C0(), rawSkeletonFixtureCObjectOffsets4E23C0()
	for i, cOffset := range cOffsets {
		if offsets[i] != cOffset {
			t.Fatalf("field%d Go=%d C=%d", i, offsets[i], cOffset)
		}
	}
	// Waypoints starts at304 on arm64: its eight-byte alignment adds four
	// bytes before sixteen widened pointers. Field119 widens once more.
	if unsafe.Offsetof(server.MonsterUpdateData{}.Waypoints) != 304 || offsets[5] != 552 || offsets[6] != 553 {
		t.Fatalf("actual compiled Monster layout differs: %v", offsets)
	}
	t.Logf("RAW_SKELETON_NATIVE_LAYOUT object=%v update_bytes=%d,%d pointer_width=%d", cOffsets, offsets[5], offsets[6], unsafe.Sizeof(unsafe.Pointer(nil)))
}

func TestRawSkeletonDefaultFallback4E23C0(t *testing.T) {
	if rawSkeletonObserverStack4E23C0 != nil {
		t.Fatal("actual default endpoints require no diagnostic observer")
	}
	for _, backend := range []string{"Go", "C"} {
		t.Run(backend, func(t *testing.T) {
			call := rawSkeletonDamageNative4E23C0
			if backend == "C" {
				call = rawSkeletonFixtureCall4E23C0
			}
			if result := call(nil, nil, nil, -2147483648, -1); result != 1 {
				t.Fatal("actual Buff and raw DefaultDamage nil-target fallback", result)
			}
		})
	}
}

func TestRawSkeletonPinnedGoHeap4E23C0(t *testing.T) {
	for _, backend := range []string{"Go", "C"} {
		t.Run(backend, func(t *testing.T) {
			target, source := new(server.Object), new(server.Object)
			update, alternate := new(server.MonsterUpdateData), new(server.MonsterUpdateData)
			target.UpdateData = unsafe.Pointer(update)
			target.Direction1 = 0xffff
			update.Field120_0, update.Field120_1 = 100, 80
			*(*uint32)(unsafe.Pointer(&source.PrevPos.X)) = 0x7fa12345
			*(*uint32)(unsafe.Pointer(&source.PrevPos.Y)) = 0x80000000
			var pin runtime.Pinner
			for _, value := range []any{target, source, update, alternate} {
				pin.Pin(value)
			}
			defer pin.Unpin()
			if uintptr(unsafe.Pointer(target)) <= 0xffffffff || uintptr(unsafe.Pointer(update)) <= 0xffffffff {
				t.Fatal("actual Go heap pointers above4GiB")
			}
			var calls []uint32
			pop := pushRawSkeletonObserver4E23C0(&rawSkeletonObserver4E23C0{Service: func(call rawSkeletonService4E23C0) (uint32, bool) {
				calls = append(calls, call.Site)
				a := call.Arguments
				switch call.Site {
				case 0x4e23ce:
					if len(a) != 2 || a[0].Pointer != unsafe.Pointer(target) || a[1].Word != 23 {
						t.Fatal("full native Buff arguments")
					}
					return 0, true
				case 0x4e243f:
					if len(a) != 3 || a[0].Pointer != unsafe.Pointer(&target.PosVec) || a[1].Word != 0xffffffff || *(*[2]uint32)(a[2].Pointer) != [2]uint32{0x7fa12345, 0x80000000} {
						t.Fatal("actual point pointers, signed direction and raw coordinate bits")
					}
					return 0x80000101, true
				case 0x4e2452:
					if len(a) != 1 || a[0].Pointer != unsafe.Pointer(target) {
						t.Fatal("full native Action target")
					}
					target.UpdateData = unsafe.Pointer(alternate)
					runtime.GC()
					return 21, true
				case 0x4e247b:
					if len(a) != 4 || a[0].Word != 878 || a[1].Pointer != unsafe.Pointer(target) || a[2].Word != 0 || a[3].Word != 0 {
						t.Fatal("raw Audio frame")
					}
					return 0x80000001, true // The void clobber must not replace return1.
				default:
					t.Fatal(fmt.Sprintf("unexpected raw service %#x", call.Site))
					return 0, false
				}
			}})
			defer pop()
			call := rawSkeletonDamageNative4E23C0
			if backend == "C" {
				call = rawSkeletonFixtureCall4E23C0
			}
			if result := call(unsafe.Pointer(target), unsafe.Pointer(source), nil, -1, -1); result != 1 {
				t.Fatal("cached native update pointer/raw void result", result)
			}
			if !reflect.DeepEqual(calls, []uint32{0x4e23ce, 0x4e243f, 0x4e2452, 0x4e247b}) {
				t.Fatal("complete raw service order", calls)
			}
			t.Logf("RAW_SKELETON_PINNED_GOHEAP backend=%s target=%p source=%p update=%p cached_after_GC=true", backend, target, source, update)
		})
	}
}
