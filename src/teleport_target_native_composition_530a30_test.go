package opennox

import (
	"bytes"
	"math"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
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

// Exercise both actual C entry points through the production runtime, floor
// query, ray, allocator, wake cache, movement and spatial index. This isolated
// blank map and synthetic wake definition do not establish campaign coverage.
func TestSpellTeleportTargetPublicCNativeComposition530A30(t *testing.T) {
	const childEnv = "NOX_TELEPORT_TARGET_NATIVE_COMPOSITION_CHILD"
	if os.Getenv(childEnv) == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(binary, "-test.run=^TestSpellTeleportTargetPublicCNativeComposition530A30$", "-test.v")
		cmd.Env = append(os.Environ(), childEnv+"=1")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("public C TeleportToTarget composition: %v\n%s", err, out)
		} else {
			t.Logf("%s", out)
		}
		return
	}
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	s.spells.duration.Init(s)
	s.Map.Init()
	t.Cleanup(s.Map.Free)
	if s.Walls.Init() == 0 || legacy.Nox_xxx_tileAlloc_410F60_init() == 0 {
		t.Fatal("isolated blank map initialization")
	}
	t.Cleanup(s.Walls.Free)
	// The zero-filled floor is tile 0; these five registered tile indices are
	// the nonzero teleport-blocking kinds. The child owns the floor allocation.
	for i := 0; i < 5; i++ {
		index := memmap.PtrUint32(0x587000, 26516+4*uintptr(i))
		previous := *index
		*index = 100 + uint32(i)
		t.Cleanup(func() { *index = previous })
	}
	s.Audio.Init(s.Server)
	t.Cleanup(s.Audio.Free)
	oldServer, oldGet := noxServer, legacy.GetServer
	noxServer, legacy.GetServer = s, func() legacy.Server { return s }
	t.Cleanup(func() { noxServer, legacy.GetServer = oldServer, oldGet })
	oldFlags := noxflags.GetGame()
	noxflags.UnsetGame(oldFlags)
	t.Cleanup(func() { noxflags.UnsetGame(noxflags.GetGame()); noxflags.SetGame(oldFlags) })
	s.SetFrame(142)
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
	caster, target := s.Objs.NewObject(&server.ObjectType{}), s.Objs.NewObject(&server.ObjectType{})
	caster.ObjClass, target.ObjClass = object.ClassSimple, object.ClassSimple
	origin, destination := types.Ptf(100, 200), types.Ptf(300, 400)
	s.CreateObjectAt(caster, nil, types.Ptf(110, 210))
	s.CreateObjectAt(target, nil, origin)
	cache := memmap.PtrUint32(0x5D4594, 2487916)
	previousType := *cache
	*cache = 0
	t.Cleanup(func() { *cache = previousType })
	type frame struct {
		Before [32]byte
		Record server.DurSpell
		After  [32]byte
	}
	f, free := alloc.New(frame{})
	t.Cleanup(free)
	*f = frame{}
	for i := range f.Before {
		f.Before[i], f.After[i] = byte(0x51+i), byte(0xe7-i)
	}
	f.Record = server.DurSpell{ID: 0xabcd, Spell: uint32(spell.SPELL_TELEPORT_TO_TARGET), Level: 3, Obj12: target, Caster16: caster, Target48: target,
		Pos: types.Ptf(11.5, -22.25), Pos2: destination, Frame68: 0x87654321, Field36: 0x13579bdf, Flags88: 0x77556601}
	for _, p := range []unsafe.Pointer{f.Record.C(), caster.CObj(), target.CObj()} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
			t.Fatalf("native C-owned pointer must be above4GiB: %p", p)
		}
	}
	want := *f
	want.Record.Frame68 = 143
	if got := ccall.CallIntPtr(legacy.Get_sub_530A30_spell_execdur(), f.Record.C()); got != 0 {
		t.Fatalf("public create return=%d", got)
	}
	raw := func(v *frame) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(v)), int(unsafe.Sizeof(*v))) }
	if !bytes.Equal(raw(f), raw(&want)) {
		t.Fatal("public create record/canary differs from scheduled native record")
	}
	noxflags.SetGame(noxflags.GameModeCoop)
	previousPending := s.Objs.Pending
	if got := ccall.CallIntPtr(legacy.Get_nox_xxx_castTTT_530B70(), f.Record.C()); got != 1 || !bytes.Equal(raw(f), raw(&want)) {
		t.Fatalf("public update return=%d or immutable native record/canary changed", got)
	}
	wake := s.Objs.Pending
	kind := uint32(s.Types.IndByID("TeleportWake"))
	if kind == 0 || *cache != kind || wake == nil || wake == previousPending || wake.CollideData == nil ||
		uint32(wake.TypeInd) != kind || wake.ObjOwner != target || wake.PosVec != origin ||
		(*server.TeleportWakeCollideData)(wake.CollideData).Destination != destination || wake.Field34 != 172 {
		t.Fatalf("public update did not create its native wake: cache=%d wake=%p", *cache, wake)
	}
	for _, p := range []unsafe.Pointer{wake.CObj(), wake.CollideData} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
			t.Fatalf("native wake pointer must be above4GiB: %p", p)
		}
	}
	if target.PosVec != destination || target.NewPos != destination || target.PrevPos != destination ||
		!target.Flags().Has(object.FlagPartitioned) || target.Flags().Has(object.FlagDestroyed) || caster.PosVec != types.Ptf(110, 210) {
		t.Fatalf("real C teleport coordinates/index: target=%v/%v/%v flags=%v caster=%v", target.PosVec, target.NewPos, target.PrevPos, target.Flags(), caster.PosVec)
	}
	t.Logf("public C create/update record=%p target=%p %v->%v wake=%p collision=%p cache=%d native-canaries=intact",
		&f.Record, target, origin, target.PosVec, wake, wake.CollideData, *cache)
}
