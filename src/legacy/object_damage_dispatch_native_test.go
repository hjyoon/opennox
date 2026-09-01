package legacy

import (
	"fmt"
	"math"
	"runtime"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

var objectDamageNativeTestSequence atomic.Uint64

func TestObjectDamageDispatchKeepsNativePointers(t *testing.T) {
	target := &server.Object{}
	source := &server.Object{}
	weapon := &server.Object{}
	var pin runtime.Pinner
	pin.Pin(target)
	pin.Pin(source)
	pin.Pin(weapon)
	defer pin.Unpin()

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"target": unsafe.Pointer(target),
			"source": unsafe.Pointer(source),
			"weapon": unsafe.Pointer(weapon),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want native high address", name, pointer)
			}
		}
	}

	fnc := objectDamageNativeProbePtr()
	called := false
	server.RegisterObjectDamageGo(
		fmt.Sprintf("NativeWidthDamageTest%d", objectDamageNativeTestSequence.Add(1)),
		fnc,
		func(gotTarget, gotSource, gotWeapon *server.Object, damage int32, typ object.DamageType) bool {
			called = true
			if gotTarget != target || gotSource != source || gotWeapon != weapon {
				t.Fatalf("damage objects = %p/%p/%p, want %p/%p/%p",
					gotTarget, gotSource, gotWeapon, target, source, weapon)
			}
			if damage != -31 || typ != object.DamageElectric {
				t.Fatalf("damage values = %d/%d, want -31/%d", damage, typ, object.DamageElectric)
			}
			return true
		},
	)
	target.Damage = fnc

	if !objectDamageDispatchCallNative(target, source, weapon, -31, object.DamageElectric) {
		t.Fatal("native damage dispatch returned false")
	}
	if !called {
		t.Fatal("registered native-width damage callback was not called")
	}
	runtime.KeepAlive(target)
	runtime.KeepAlive(source)
	runtime.KeepAlive(weapon)
}
