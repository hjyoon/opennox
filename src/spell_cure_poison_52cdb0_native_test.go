package opennox

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Both exported selector routes use actual C-owned pointers above 4 GiB.
// No class/status rejection is added by CurePoison itself; the downstream
// poison services still own their original class-dependent effects.
func TestCurePoisonGameAndLegacyNativeState52CDB0(t *testing.T) {
	previousServer, previousGetServer := noxServer, legacy.GetServer
	s := &Server{Server: new(server.Server)}
	noxServer, legacy.GetServer = s, func() legacy.Server { return s }
	t.Cleanup(func() { noxServer, legacy.GetServer = previousServer, previousGetServer })
	target, freeTarget := alloc.New(server.Object{})
	defer freeTarget()
	other, freeOther := alloc.New(server.Object{})
	defer freeOther()
	health, freeHealth := alloc.New(server.HealthData{})
	defer freeHealth()
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	defer freeArg()
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(other), unsafe.Pointer(health), unsafe.Pointer(arg)} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("actual game selector allocation=%p, want >4 GiB", ptr)
			}
		}
	}
	for _, route := range []string{"game", "legacy"} {
		for _, power := range []int32{math.MinInt32, -3, -1, 0, 1, 2, 3, 255, math.MaxInt32} {
			for _, flags := range []object.Flags{object.FlagActive, object.FlagDestroyed} {
				t.Run(fmt.Sprintf("%s/power-%d/flags-%x", route, power, uint32(flags)), func(t *testing.T) {
					*health = server.HealthData{Cur: 77, Max: 80, Field16: 91}
					*target = server.Object{HealthData: health, Poison540: 3, Field542: 1234, ObjFlags: flags}
					*arg = server.SpellAcceptArg{Obj: target, Pos: types.Ptf(-1000, 2000)}
					wantArg := *arg
					selector := castCurePoison
					if route == "legacy" {
						selector = legacy.Nox_xxx_castCurePoison_52CDB0
					}
					wantPoison, wantFrame := uint8(0), uint32(0)
					if power < 3 {
						wantPoison, wantFrame = uint8(3)-uint8(power), 91
					}
					if got := selector(spell.SPELL_CURE_POISON, target, other, other, arg, int(power)); got != 1 ||
						target.Poison540 != wantPoison || target.Field542 != 1234 || target.ObjFlags != flags ||
						health.Cur != 77 || health.Max != 80 || health.Field16 != wantFrame || *arg != wantArg {
						t.Fatalf("route=%s result=%d target=%+v health=%+v arg=%+v", route, got, target, health, arg)
					}
				})
			}
		}
	}
	*arg = server.SpellAcceptArg{Pos: types.Ptf(-1000, 2000)}
	if got := castCurePoison(spell.SPELL_CURE_POISON, target, other, other, arg, 3); got != 0 ||
		arg.Obj != nil || arg.Pos != types.Ptf(-1000, 2000) {
		t.Fatalf("nil game target result/arg=%d/%+v", got, arg)
	}
}

func TestCurePoisonGameRequiredAcceptancePointer52CDB0(t *testing.T) {
	previous := noxServer
	noxServer = &Server{Server: new(server.Server)}
	t.Cleanup(func() { noxServer = previous })
	defer func() {
		if recover() == nil {
			t.Fatal("required acceptance pointer no longer faults at its first load")
		}
	}()
	castCurePoison(spell.SPELL_CURE_POISON, nil, nil, nil, nil, 3)
}
