package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type chakramReflectTestServer4EB3E0 struct {
	Server
	world *server.Server
}

func (s *chakramReflectTestServer4EB3E0) S() *server.Server {
	return s.world
}

func TestChakramRandomReflectPublicC4EB3E0(t *testing.T) {
	world := new(server.Server)
	world.Rand.Logic, world.Rand.Other = prand.New(0), prand.New(19)
	obj, done := alloc.NewHigh(server.Object{})
	defer done()
	*obj = server.Object{
		VelVec:  types.Ptf(3, 4),
		PosVec:  types.Ptf(100, 200),
		NewPos:  types.Ptf(300, 400),
		PrevPos: types.Ptf(-12.5, 33.25),
		Field37: 0x12345678,
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 {
		t.Fatalf("object address %p is below the native 64-bit test boundary", obj)
	}

	oldServer, oldMove := GetServer, Nox_xxx_moveUpdateSpecial_517970
	t.Cleanup(func() {
		GetServer, Nox_xxx_moveUpdateSpecial_517970 = oldServer, oldMove
	})
	GetServer = func() Server { return &chakramReflectTestServer4EB3E0{world: world} }
	calls := 0
	want := *obj
	// GAME.EXE 004EB3E0 with direction 38 and the first Logic draw (-32).
	want.VelVec = types.Ptf(math.Float32frombits(0xc09e44ab), math.Float32frombits(0xbf3bd0a3))
	want.PosVec, want.NewPos = obj.PrevPos, obj.PrevPos
	Nox_xxx_moveUpdateSpecial_517970 = func(receiver *server.Object) {
		calls++
		if receiver != obj || *receiver != want || world.Rand.Logic.Index() != 1 {
			t.Error("movement callback did not observe the original native object state after the RNG draw")
		}
	}

	chakramRandomReflectCallC4EB3E0(obj)
	if *obj != want || calls != 1 || world.Rand.Logic.Index() != 1 || world.Rand.Other.Index() != 19 {
		t.Fatalf("C reflection velocity %v, movement calls %d, RNG indexes %d/%d", obj.VelVec, calls,
			world.Rand.Logic.Index(), world.Rand.Other.Index())
	}
}
