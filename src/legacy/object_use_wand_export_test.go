package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

func TestWandUseExportsAndRegistrationPreserveNativePointers(t *testing.T) {
	owner := &server.Object{ObjClass: object.ClassPlayer}
	wand := new(server.Object)
	var pin runtime.Pinner
	pin.Pin(owner)
	pin.Pin(wand)
	defer pin.Unpin()
	assertWandUseHighPointers(t, owner, wand)

	tests := []struct {
		name    string
		current func() server.UseFunc
		set     func(server.UseFunc)
		export  func(*server.Object, *server.Object) int32
		pointer func() unsafe.Pointer
		want    bool
	}{
		{
			name:    "WandUse",
			current: func() server.UseFunc { return Nox_xxx_useWand_53F290 },
			set:     func(f server.UseFunc) { Nox_xxx_useWand_53F290 = f },
			export:  wandUseExportCall53F290,
			pointer: Get_nox_xxx_useLesserFireballStaff_53F290,
			want:    true,
		},
		{
			name:    "WandCastUse",
			current: func() server.UseFunc { return Nox_xxx_useWandCast_53F4F0 },
			set:     func(f server.UseFunc) { Nox_xxx_useWandCast_53F4F0 = f },
			export:  wandCastUseExportCall53F4F0,
			pointer: Get_nox_xxx_useWandCastSpell_53F4F0,
			want:    false,
		},
		{
			name:    "FireWandUse",
			current: func() server.UseFunc { return Nox_xxx_useFireWand_53F670 },
			set:     func(f server.UseFunc) { Nox_xxx_useFireWand_53F670 = f },
			export:  fireWandUseExportCall53F670,
			pointer: Get_nox_xxx_useFireWand_53F670,
			want:    true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			old := tc.current()
			t.Cleanup(func() { tc.set(old) })
			calls := 0
			tc.set(func(gotOwner, gotWand *server.Object) bool {
				calls++
				if gotOwner != owner || gotWand != wand {
					t.Fatalf("callback pointers = %p/%p, want %p/%p", gotOwner, gotWand, owner, wand)
				}
				return tc.want
			})

			wantInt := int32(0)
			if tc.want {
				wantInt = 1
			}
			if got := tc.export(owner, wand); got != wantInt {
				t.Fatalf("export result = %d, want %d", got, wantInt)
			}

			use := server.UseFuncPtr{Ptr: tc.pointer()}.Get()
			if use == nil {
				t.Fatal("registered callback is nil")
			}
			if got := use(owner, wand); got != tc.want {
				t.Fatalf("registered callback result = %v, want %v", got, tc.want)
			}

			wand.Use = server.UseFuncPtr{Ptr: tc.pointer()}
			new(server.Server).SignCollide4EAB40(wand, owner, nil)
			if calls != 3 {
				t.Fatalf("callback calls = %d, want 3", calls)
			}
		})
	}
	runtime.KeepAlive(owner)
	runtime.KeepAlive(wand)
}

func assertWandUseHighPointers(t *testing.T, owner, wand *server.Object) {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) != 8 {
		return
	}
	if uintptr(unsafe.Pointer(owner)) <= math.MaxUint32 || uintptr(unsafe.Pointer(wand)) <= math.MaxUint32 {
		t.Fatalf("test pointers do not exercise native high addresses: owner=%p wand=%p", owner, wand)
	}
}
