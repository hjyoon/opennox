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
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

// Actual C entries share the production runtime, allocator, movement gate,
// coordinate updates and spatial index. The simple bodies and glyph input are
// isolated synthetic controls, not earned campaign or sound-packet coverage.
func TestSpellSwapPublicCRealMovement530CA0(t *testing.T) {
	const childEnv = "NOX_SWAP_NATIVE_C_MOVEMENT_CHILD"
	if os.Getenv(childEnv) == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(binary, "-test.run=^TestSpellSwapPublicCRealMovement530CA0$", "-test.v")
		cmd.Env = append(os.Environ(), childEnv+"=1")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("public C Swap movement: %v\n%s", err, out)
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
	s.Audio.Init(s.Server)
	t.Cleanup(s.Audio.Free)
	oldServer, oldGet := noxServer, legacy.GetServer
	noxServer, legacy.GetServer = s, func() legacy.Server { return s }
	t.Cleanup(func() { noxServer, legacy.GetServer = oldServer, oldGet })
	oldFlags := noxflags.GetGame()
	noxflags.UnsetGame(oldFlags)
	t.Cleanup(func() { noxflags.UnsetGame(noxflags.GetGame()); noxflags.SetGame(oldFlags) })
	s.SetFrame(142)
	if !s.Objs.Init(8) {
		t.Fatal("native allocator initialization")
	}
	t.Cleanup(s.Objs.FreeObjects)
	caster, target := s.Objs.NewObject(&server.ObjectType{}), s.Objs.NewObject(&server.ObjectType{})
	caster.ObjClass, target.ObjClass = object.ClassSimple, object.ClassSimple
	origin, destination := types.Ptf(100, 200), types.Ptf(300, 400)
	s.CreateObjectAt(caster, nil, origin)
	s.CreateObjectAt(target, nil, destination)
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
	f.Record = server.DurSpell{ID: 0xabcd, Spell: uint32(spell.SPELL_SWAP), Level: 3, Obj12: target, Caster16: caster, Target48: target, Flag20: 1, Frame68: 0x87654321}
	for _, p := range []unsafe.Pointer{f.Record.C(), caster.CObj(), target.CObj()} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
			t.Fatalf("native C-owned pointer must be above4GiB: %p", p)
		}
	}
	want := *f
	want.Record.Frame68 = 143
	if got := ccall.CallIntPtr(legacy.Get_sub_530CA0(), f.Record.C()); got != 0 {
		t.Fatalf("public create return=%d", got)
	}
	raw := func(v *frame) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(v)), int(unsafe.Sizeof(*v))) }
	if !bytes.Equal(raw(f), raw(&want)) {
		t.Fatal("public create record/canary differs from scheduled native record")
	}
	// The original co-op movement gate admits simple map objects. Glyph mode
	// skips LOS/player bounds while retaining both real teleport calls.
	noxflags.SetGame(noxflags.GameModeCoop)
	if got := ccall.CallIntPtr(legacy.Get_sub_530D30(), f.Record.C()); got != 1 || !bytes.Equal(raw(f), raw(&want)) {
		t.Fatalf("public update return=%d or immutable native record/canary changed", got)
	}
	if caster.PosVec != destination || caster.NewPos != destination || caster.PrevPos != destination ||
		target.PosVec != origin || target.NewPos != origin || target.PrevPos != origin ||
		!caster.Flags().Has(object.FlagPartitioned) || !target.Flags().Has(object.FlagPartitioned) ||
		caster.Flags().Has(object.FlagDestroyed) || target.Flags().Has(object.FlagDestroyed) {
		t.Fatalf("real C Swap coordinates/index: caster=%v/%v/%v flags=%v target=%v/%v/%v flags=%v", caster.PosVec, caster.NewPos, caster.PrevPos, caster.Flags(), target.PosVec, target.NewPos, target.PrevPos, target.Flags())
	}
	t.Logf("public C Swap record=%p caster=%p %v->%v target=%p %v->%v spatial-index=true native-canaries=intact", &f.Record, caster, origin, caster.PosVec, target, destination, target.PosVec)
}
