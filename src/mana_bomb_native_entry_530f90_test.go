package opennox

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/opennox/libs/datapath"
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
func TestManaBombNextPublicEntriesNativeRecordProbe530F90(t *testing.T) {
	const child = "NOX_MANA_BOMB_NEXT_ENTRY_PROBE_CHILD"
	mode := os.Getenv(child)
	if mode == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		for _, mode := range []string{"create-glyph-missing-type", "update-idle", "destroy-empty"} {
			t.Run(mode, func(t *testing.T) {
				cmd := exec.Command(binary, "-test.run=^TestManaBombNextPublicEntriesNativeRecordProbe530F90$", "-test.v")
				cmd.Env = append(os.Environ(), child+"="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil {
					t.Fatalf("actual ManaBomb %s C entry: %v", mode, err)
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
	dir, oldData := t.TempDir(), datapath.Data()
	datapath.SetData(dir)
	t.Cleanup(func() { datapath.SetData(oldData) })
	if err := os.WriteFile(filepath.Join(dir, "gamedata.yml"), []byte("ManaBombInitPower: [2, 2, 2]\nManaBombGlyphDuration: 10\nManaBombDeltaPower: [0, 0, 0]\nManaBombInRadius: 10\nManaBombOutRadius: 20\nManaBombShakeMag: 1\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := s.Balance.Read(); err != nil {
		t.Fatal(err)
	}
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	if s.Types.ByID("ManaBombCharge") != nil {
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
	f.Record = server.DurSpell{ID: 0xabcd, Spell: uint32(spell.SPELL_MANA_BOMB), Level: 3, Obj12: source, Caster16: source, Flag20: 1, Obj24: source, Field84: 1, Pos: types.Ptf(100, 200), Frame68: 400, Field72: 0x12340055, Field76: 0x55aa, Flags88: 0x77556601}
	if mode == "create-glyph-missing-type" {
		f.Record.Obj12 = nil
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(f.Record.C()) <= 0xffffffff || uintptr(source.CObj()) <= 0xffffffff) {
		t.Fatal("native C-owned ManaBomb links must exceed4GiB")
	}
	expected := *f
	raw := func(p *frame) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p))) }
	switch mode {
	case "create-glyph-missing-type":
		callback := legacy.Get_nox_xxx_manaBomb_530F90()
		want := s.spells.duration.callCreate4FEBA0(callback, &expected.Record)
		if want != 0 || expected.Record.Field72 != 2 || expected.Record.Field76 != 0 || expected.Record.Field84 != 0 || expected.Record.Frame68 != 160 {
			t.Fatal("ordinary native glyph/balance control")
		}
		t.Logf("ordinary create record=%p glyph/missing type return0 power=%#x", &f.Record, expected.Record.Field72)
		if got := ccall.CallIntPtr(callback, f.Record.C()); got != int(want) {
			t.Fatalf("C=%d ordinary=%d", got, want)
		}
	case "update-idle":
		callback := legacy.Get_nox_xxx_manaBombBoom_5310C0()
		want := s.spells.duration.callUpdate4FEEF0(callback, &expected.Record)
		if want != 0 {
			t.Fatal("ordinary native idle update control")
		}
		t.Logf("ordinary update record=%p monster-caster/glyph/idle return0", &f.Record)
		if got := ccall.CallIntPtr(callback, f.Record.C()); got != int(want) {
			t.Fatalf("C=%d ordinary=%d", got, want)
		}
	case "destroy-empty":
		callback := legacy.Get_sub_531290()
		s.spells.duration.callDestroy4FEDA0(callback, &expected.Record)
		t.Logf("ordinary destroy record=%p glyph/no native charge/FX completed", &f.Record)
		ccall.CallVoidPtr(callback, f.Record.C())
	default:
		t.Fatal("unknown probe mode")
	}
	if !bytes.Equal(raw(f), raw(&expected)) {
		t.Fatal("public C duration/canary differs from ordinary native dispatch")
	}
}
