package opennox

import (
	"bytes"
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
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

// External read-only overlay on the clean preceding runtime. Each ordinary
// native control completes before the same native record reaches actual C.
func TestForceOfNatureNextPublicEntriesNativeRecordProbe52EF30(t *testing.T) {
	const child = "NOX_FORCE_OF_NATURE_NEXT_ENTRY_PROBE_CHILD"
	mode := os.Getenv(child)
	if mode == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		for _, mode := range []string{"create-missing-type", "update-idle", "destroy-empty"} {
			t.Run(mode, func(t *testing.T) {
				cmd := exec.Command(binary, "-test.run=^TestForceOfNatureNextPublicEntriesNativeRecordProbe52EF30$", "-test.v")
				cmd.Env = append(os.Environ(), child+"="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil {
					t.Fatalf("actual ForceOfNature %s C entry: %v", mode, err)
				}
			})
		}
		return
	}
	handles.Init()
	t.Cleanup(handles.Release)
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	oldServer, oldGet := noxServer, legacy.GetServer
	noxServer, legacy.GetServer = s, func() legacy.Server { return s }
	t.Cleanup(func() { noxServer, legacy.GetServer = oldServer, oldGet })
	oldFlags := noxflags.GetGame()
	noxflags.UnsetGame(oldFlags)
	t.Cleanup(func() { noxflags.UnsetGame(noxflags.GetGame()); noxflags.SetGame(oldFlags) })
	s.spells.duration.Init(s)
	t.Cleanup(s.spells.duration.Free)
	s.SetFrame(150)
	s.SetTickRate(30)
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	if s.Types.ByID("ForceOfNatureCharge") != nil || s.Types.ByID("DeathBall") != nil {
		t.Fatal("declared missing native object types")
	}
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
	*source = server.Object{ObjClass: object.ClassMonster, PosVec: types.Ptf(11.5, -22.25), Direction1: 13}
	for i := range f.Before {
		f.Before[i], f.After[i] = byte(0x51+i), byte(0xe7-i)
	}
	f.Record = server.DurSpell{ID: 0xabcd, Spell: uint32(spell.SPELL_FORCE_OF_NATURE), Level: 3, Obj12: source, Obj24: source, Pos: types.Ptf(100, 200), Frame68: 400, Field72: 0x12340055, Field76: 0x55aa, Flags88: 0x77556601}
	if mode == "create-missing-type" {
		f.Record.Caster16 = source
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(f.Record.C()) <= 0xffffffff || uintptr(source.CObj()) <= 0xffffffff) {
		t.Fatal("native C-owned ForceOfNature links must exceed4GiB")
	}
	expected := *f
	raw := func(p *frame) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p))) }
	switch mode {
	case "create-missing-type":
		callback := legacy.Get_sub_52EF30()
		want := s.spells.duration.callCreate4FEBA0(callback, &expected.Record)
		if want != 0 || expected.Record.Field72 != 0x1234000d {
			t.Fatal("ordinary native create/direction control")
		}
		t.Logf("ordinary create record=%p missing type return0 direction=%#x", &f.Record, expected.Record.Field72)
		if got := ccall.CallIntPtr(callback, f.Record.C()); got != int(want) {
			t.Fatalf("C=%d ordinary=%d", got, want)
		}
	case "update-idle":
		callback := legacy.Get_sub_52EFD0()
		want := s.spells.duration.callUpdate4FEEF0(callback, &expected.Record)
		if want != 0 {
			t.Fatal("ordinary native idle update control")
		}
		t.Logf("ordinary update record=%p nilcaster/idle return0", &f.Record)
		if got := ccall.CallIntPtr(callback, f.Record.C()); got != int(want) {
			t.Fatalf("C=%d ordinary=%d", got, want)
		}
	case "destroy-empty":
		callback := legacy.Get_sub_52F1D0()
		s.spells.duration.callDestroy4FEDA0(callback, &expected.Record)
		t.Logf("ordinary destroy record=%p nilcaster/no native charge completed", &f.Record)
		ccall.CallVoidPtr(callback, f.Record.C())
	default:
		t.Fatal("unknown probe mode")
	}
	if !bytes.Equal(raw(f), raw(&expected)) {
		t.Fatal("public C duration/canary differs from ordinary native dispatch")
	}
}
