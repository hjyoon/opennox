package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func TestDefaultDropDispatchKeepsNativePointerWidth(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("native-width routing regression applies to 64-bit builds")
	}

	owner := &server.Object{}
	item := &server.Object{}
	point := &types.Pointf{X: 12.5, Y: -3.25}
	for name, ptr := range map[string]unsafe.Pointer{
		"owner": unsafe.Pointer(owner),
		"item":  unsafe.Pointer(item),
		"point": unsafe.Pointer(point),
	} {
		if uintptr(ptr) <= math.MaxUint32 {
			t.Fatalf("%s pointer = %p, want address above the ABI32 range", name, ptr)
		}
	}

	old := defaultDropCall4ED290
	t.Cleanup(func() {
		defaultDropCall4ED290 = old
	})
	const wantResult int32 = -0x7654321
	var calls int
	defaultDropCall4ED290 = func(gotOwner, gotItem *server.Object, gotPoint *types.Pointf) int32 {
		calls++
		if gotOwner != owner || gotItem != item || gotPoint != point {
			t.Fatalf("DefaultDrop args = %p/%p/%p, want %p/%p/%p", gotOwner, gotItem, gotPoint, owner, item, point)
		}
		return wantResult
	}

	handler, ok := server.ObjectDropHandler("DefaultDrop")
	if !ok || handler.Ptr == nil {
		t.Fatalf("ObjectDropHandler(DefaultDrop) = %p/%t, want non-nil/true", handler.Ptr, ok)
	}
	if got := handler.Get()(owner, item, point); got != wantResult {
		t.Fatalf("DefaultDrop result = %d, want %d", got, wantResult)
	}
	if calls != 1 {
		t.Fatalf("DefaultDrop calls = %d, want 1", calls)
	}
	runtime.KeepAlive(owner)
	runtime.KeepAlive(item)
	runtime.KeepAlive(point)
}

func TestEquipmentDropDispatchKeepsNativePointerWidth(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("native-width routing regression applies to 64-bit builds")
	}

	owner := &server.Object{}
	item := &server.Object{}
	point := &types.Pointf{X: -18.5, Y: 27.25}
	for name, ptr := range map[string]unsafe.Pointer{
		"owner": unsafe.Pointer(owner),
		"item":  unsafe.Pointer(item),
		"point": unsafe.Pointer(point),
	} {
		if uintptr(ptr) <= math.MaxUint32 {
			t.Fatalf("%s pointer = %p, want address above the ABI32 range", name, ptr)
		}
	}

	tests := []struct {
		name string
		set  func(server.DropFunc) func()
	}{
		{
			name: "WeaponDrop",
			set: func(f server.DropFunc) func() {
				old := weaponDropCall53AB10
				weaponDropCall53AB10 = f
				return func() { weaponDropCall53AB10 = old }
			},
		},
		{
			name: "ArmorDrop",
			set: func(f server.DropFunc) func() {
				old := armorDropCall53EB70
				armorDropCall53EB70 = f
				return func() { armorDropCall53EB70 = old }
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			const wantResult int32 = -0x7654321
			var calls int
			restore := tc.set(func(gotOwner, gotItem *server.Object, gotPoint *types.Pointf) int32 {
				calls++
				if gotOwner != owner || gotItem != item || gotPoint != point {
					t.Fatalf("%s args = %p/%p/%p, want %p/%p/%p", tc.name, gotOwner, gotItem, gotPoint, owner, item, point)
				}
				return wantResult
			})
			t.Cleanup(restore)

			handler, ok := server.ObjectDropHandler(tc.name)
			if !ok || handler.Ptr == nil {
				t.Fatalf("ObjectDropHandler(%s) = %p/%t, want non-nil/true", tc.name, handler.Ptr, ok)
			}
			if got := handler.Get()(owner, item, point); got != wantResult {
				t.Fatalf("%s result = %d, want %d", tc.name, got, wantResult)
			}
			if calls != 1 {
				t.Fatalf("%s calls = %d, want 1", tc.name, calls)
			}
		})
	}
	runtime.KeepAlive(owner)
	runtime.KeepAlive(item)
	runtime.KeepAlive(point)
}
