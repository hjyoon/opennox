package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/balance"
	"github.com/opennox/libs/types"
)

func TestCastConfuseNative52C1E0BalanceSpillAndLiveObjectPointers(t *testing.T) {
	s := &Server{}
	s.Balance.file = &balance.File{Global: balance.Config{"confuseenchantduration": balance.Float(2.50000001)}}
	source, target, reloaded := new(Object), new(Object), new(Object)
	arg := &SpellAcceptArg{Obj: target, Pos: types.Pointf{X: -4.5, Y: 10.25}}
	if unsafe.Sizeof(uintptr(0)) > 4 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(source), unsafe.Pointer(target), unsafe.Pointer(reloaded), unsafe.Pointer(arg)} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("native pointer = %p, want above 4 GiB", ptr)
			}
		}
	}
	applyCalls, attributeCalls := 0, 0
	r := CastConfuseRuntime52C1E0{
		BuffApply: func(unit *Object, buff int32, duration int16, power int8) {
			applyCalls++
			if unit != target || buff != 3 || duration != 2 || power != -128 {
				t.Fatalf("buff = %p/%d/%d/%d", unit, buff, duration, power)
			}
			arg.Obj = reloaded
		},
		Attribution: func(caster, unit *Object) {
			attributeCalls++
			if caster != source || unit != reloaded {
				t.Fatalf("attribution = %p/%p", caster, unit)
			}
		},
	}
	if got := s.CastConfuse52C1E0(source, arg, 0x180, r); got != 1 || applyCalls != 1 || attributeCalls != 1 {
		t.Fatalf("result/calls = %d/%d/%d", got, applyCalls, attributeCalls)
	}
	if arg.Obj != reloaded || arg.Pos != (types.Pointf{X: -4.5, Y: 10.25}) {
		t.Fatal("unexpected argument change")
	}
	arg.Obj = nil
	if got := s.CastConfuse52C1E0(source, arg, 1, CastConfuseRuntime52C1E0{}); got != 0 {
		t.Fatalf("nil target = %d", got)
	}
}
