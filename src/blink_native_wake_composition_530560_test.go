package opennox

import (
	"math"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

// The public C helper uses the production type reader, allocator, cache and
// CreateObjectAt. Only its isolated object definitions are synthetic; this
// fixture makes no earned-campaign or transitive collision-handler claim.
func TestSpellBlinkPublicCNativeWakeComposition530560(t *testing.T) {
	const childEnv = "NOX_BLINK_NATIVE_WAKE_COMPOSITION_CHILD"
	if os.Getenv(childEnv) == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(binary, "-test.run=^TestSpellBlinkPublicCNativeWakeComposition530560$", "-test.v")
		cmd.Env = append(os.Environ(), childEnv+"=1")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("public C wake composition: %v\n%s", err, out)
		} else {
			t.Logf("%s", out)
		}
		return
	}
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	s.spells.duration.Init(s)
	oldServer, oldGet := noxServer, legacy.GetServer
	noxServer, legacy.GetServer = s, func() legacy.Server { return s }
	t.Cleanup(func() { noxServer, legacy.GetServer = oldServer, oldGet })
	oldFlags := noxflags.GetGame()
	noxflags.UnsetGame(oldFlags)
	t.Cleanup(func() { noxflags.UnsetGame(noxflags.GetGame()); noxflags.SetGame(oldFlags) })
	s.SetFrame(0xfffffff0)
	s.SetTickRate(30)
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	if err := s.Types.ReadObjectType(&things.Thing{Name: "TeleportWake", Class: []things.Class{"SIMPLE"}, OnCollide: &things.ProcFunc{Name: "TeleportWakeCollide"}}); err != nil {
		t.Fatal(err)
	}
	if !s.Objs.Init(8) {
		t.Fatal("native allocator initialization")
	}
	t.Cleanup(s.Objs.FreeObjects)
	owner := s.Objs.NewObject(&server.ObjectType{})
	owner.ObjClass, owner.PosVec = object.ClassSimple, types.Ptf(2, 3)
	cache := memmap.PtrUint32(0x5D4594, 2487916)
	previousType := *cache
	*cache = 0
	t.Cleanup(func() { *cache = previousType })
	type pointFrame struct {
		Before [32]byte
		Point  types.Pointf
		After  [32]byte
	}
	origin, freeOrigin := alloc.New(pointFrame{})
	destination, freeDestination := alloc.New(pointFrame{})
	t.Cleanup(freeOrigin)
	t.Cleanup(freeDestination)
	*origin, *destination = pointFrame{Point: types.Ptf(123.5, -456.25)}, pointFrame{Point: types.Ptf(-17.5, 23.25)}
	for i := range origin.Before {
		origin.Before[i], origin.After[i] = byte(0x31+i), byte(0xe7-i)
		destination.Before[i], destination.After[i] = byte(0x71+i), byte(0xa7-i)
	}
	beforeOrigin, beforeDestination := *origin, *destination
	kind := uint32(s.Types.IndByID("TeleportWake"))
	if kind == 0 {
		t.Fatal("registered TeleportWake type")
	}
	for _, alias := range []bool{false, true} {
		from := &origin.Point
		if alias {
			from = &destination.Point
		}
		previousPending := s.Objs.Pending
		ccall.CallVoidPtr3(legacy.Get_nox_xxx_spellTeleportCreateWake_530560(), owner.CObj(), unsafe.Pointer(from), unsafe.Pointer(&destination.Point))
		wake := s.Objs.Pending
		if wake == nil || wake == previousPending || wake.CollideData == nil {
			t.Fatal("public C helper did not allocate a pending native wake")
		}
		for _, p := range []unsafe.Pointer{owner.CObj(), wake.CObj(), wake.CollideData, unsafe.Pointer(from), unsafe.Pointer(&destination.Point)} {
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
				t.Fatalf("native C-owned wake link must be above4GiB: %p", p)
			}
		}
		if *cache != kind || uint32(wake.TypeInd) != kind || wake.ObjOwner != owner || wake.PosVec != *from ||
			(*server.TeleportWakeCollideData)(wake.CollideData).Destination != destination.Point || wake.Field34 != 14 ||
			wake.Flags()&(object.FlagActive|object.FlagPending) != object.FlagActive|object.FlagPending || wake.ObjNext != previousPending ||
			*origin != beforeOrigin || *destination != beforeDestination || owner.PosVec != types.Ptf(2, 3) {
			t.Fatalf("public C helper/cache/CreateObjectAt: alias=%t cache=%d wake=%+v", alias, *cache, wake)
		}
		t.Logf("public C helper alias=%t cache=%d owner=%p wake=%p collision=%p origin=%v destination=%v expiry=%d", alias, *cache, owner, wake, wake.CollideData, wake.PosVec, destination.Point, wake.Field34)
	}
}
