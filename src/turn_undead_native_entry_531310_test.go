package opennox

import (
	"bytes"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
	"os"
	"os/exec"
	"testing"
	"unsafe"
)

// Regression for both originally faulting public C entries, compared with
// ordinary native dispatch using the same C-owned records and active object.
func TestTurnUndeadNextPublicEntryNativeRecordProbe531310(t *testing.T) {
	const childEnv = "NOX_TURN_UNDEAD_NEXT_ENTRY_PROBE_CHILD"
	mode := os.Getenv(childEnv)
	if mode == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		for _, mode := range []string{"create", "destroy-active-nonprojectile"} {
			t.Run(mode, func(t *testing.T) {
				cmd := exec.Command(binary, "-test.run=^TestTurnUndeadNextPublicEntryNativeRecordProbe531310$", "-test.v")
				cmd.Env = append(os.Environ(), childEnv+"="+mode)
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("actual TurnUndead %s C entry: %v\n%s", mode, err, out)
				}
			})
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
	s.Map.Init()
	t.Cleanup(s.Map.Free)
	s.Audio.Init(s.Server)
	t.Cleanup(s.Audio.Free)
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	type frame struct {
		Before [32]byte
		Record server.DurSpell
		After  [32]byte
	}
	f, free := alloc.New(frame{})
	t.Cleanup(free)
	*f = frame{}
	source, freeSource := alloc.New(server.Object{})
	t.Cleanup(freeSource)
	*source = server.Object{ObjClass: object.ClassMonster, PosVec: types.Ptf(11.5, -22.25)}
	for i := range f.Before {
		f.Before[i], f.After[i] = byte(0x51+i), byte(0xe7-i)
	}
	f.Record = server.DurSpell{ID: 0xabcd, Spell: 0x66, Level: 3, Obj12: source, Caster16: source, Obj24: source, Target48: source, Field72: 0x55aa, Pos: types.Ptf(100, 200), Frame68: 0x87654321, Flags88: 0x77556601}
	if !s.Objs.Init(8) {
		t.Fatal("native allocator initialization")
	}
	t.Cleanup(s.Objs.FreeObjects)
	var active *server.Object
	if mode == "destroy-active-nonprojectile" {
		if err := s.Types.ReadObjectType(&things.Thing{Name: "NextTurnUndeadNonProjectile", Class: []things.Class{"SIMPLE"}}); err != nil {
			t.Fatal(err)
		}
		active = s.NewObjectByTypeID("NextTurnUndeadNonProjectile")
		if active == nil || active.TypeInd == 0 {
			t.Fatal("valid distinct type")
		}
		s.CreateObjectAt(active, nil, types.Ptf(100, 200))
		s.ObjectsAddPending()
		if s.Objs.First() != active || active.ObjNext != nil {
			t.Fatal("one actual active nonprojectile")
		}
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(f.Record.C()) <= 0xffffffff || uintptr(source.CObj()) <= 0xffffffff || active != nil && uintptr(active.CObj()) <= 0xffffffff) {
		t.Fatal("native C-owned TurnUndead links must be above4GiB")
	}
	expected := *f
	raw := func(v *frame) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(v)), int(unsafe.Sizeof(*v))) }
	switch mode {
	case "create":
		callback := legacy.Get_nox_xxx_spellTurnUndeadCreate_531310()
		want := s.spells.duration.callCreate4FEBA0(callback, &expected.Record)
		if want != 0 {
			t.Fatalf("ordinary native create=%d", want)
		}
		t.Logf("native TurnUndead create record=%p source=%p expected-return=%d expected-budget=%d", &f.Record, source, want, expected.Record.Field72)
		if got := ccall.CallIntPtr(callback, f.Record.C()); got != int(want) {
			t.Fatalf("public C return=%d ordinary=%d", got, want)
		}
	case "destroy-active-nonprojectile":
		callback := legacy.Get_nox_xxx_spellTurnUndeadDelete_531420()
		s.spells.duration.callDestroy4FEDA0(callback, &expected.Record)
		if s.Objs.First() != active || active.Flags().Has(object.FlagDestroyed) {
			t.Fatal("ordinary native destroy must preserve the unrelated active object")
		}
		t.Logf("native TurnUndead destroy record=%p active=%p ordinary native dispatch completed", &f.Record, active)
		ccall.CallVoidPtr(callback, f.Record.C())
	default:
		t.Fatal("unknown mode")
	}
	if !bytes.Equal(raw(f), raw(&expected)) {
		t.Fatal("public C record/canary differs from ordinary native dispatch")
	}
}
