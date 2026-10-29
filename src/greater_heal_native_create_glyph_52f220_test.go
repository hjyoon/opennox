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
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

// Actual glyph search, native spatial index and HP setter; no search/heal
// result is injected. The original create fixture's glyph branch heals20.
func TestGreaterHealPublicCCreateGlyphNative52F220(t *testing.T) {
	const child = "NOX_GREATER_HEAL_CREATE_GLYPH_CHILD"
	if os.Getenv(child) == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(binary, "-test.run=^TestGreaterHealPublicCCreateGlyphNative52F220$", "-test.v")
		cmd.Env = append(os.Environ(), child+"=true")
		out, err := cmd.CombinedOutput()
		t.Logf("%s", out)
		if err != nil {
			t.Fatalf("actual GreaterHeal C glyph create: %v", err)
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
	s.Map.Init()
	t.Cleanup(s.Map.Free)
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	if err := s.Types.ReadObjectType(&things.Thing{Name: "NativeGreaterHealTarget"}); err != nil {
		t.Fatal(err)
	}
	if !s.Objs.Init(8) {
		t.Fatal("native object pool")
	}
	t.Cleanup(s.Objs.FreeObjects)
	target := s.Objs.NewObject(s.Types.ByID("NativeGreaterHealTarget"))
	health := plasmaNativeAllocation531580[server.HealthData](t)
	update := plasmaNativeAllocation531580[server.MonsterUpdateData](t)
	*health = server.HealthData{Cur: 6, Max: 100}
	target.ObjClass, target.ObjFlags = object.ClassMonster, object.FlagActive|object.FlagEnabled
	target.HealthData, target.UpdateData = health, unsafe.Pointer(update)
	target.PosVec, target.NewPos = types.Ptf(300, 300), types.Ptf(300, 300)
	target.Shape.Kind = server.ShapeKindCircle
	target.Shape.Circle.R, target.Shape.Circle.R2 = 1, 1
	s.Map.AddObjectToIndex(target)
	t.Cleanup(func() { s.Map.RemoveObjectFromIndex(target) })
	if !target.ObjFlags.Has(object.FlagPartitioned) {
		t.Fatal("real spatial insertion")
	}
	type guarded struct {
		Before [32]byte
		Record server.DurSpell
		After  [32]byte
	}
	f := plasmaNativeAllocation531580[guarded](t)
	for i := range f.Before {
		f.Before[i] = byte(0x51 + i)
		f.After[i] = byte(0xe7 - i)
	}
	f.Record = server.DurSpell{ID: 0xabcd, Spell: uint32(spell.SPELL_GREATER_HEAL), Level: 3, Flag20: 1, Pos: types.Ptf(100, 200), Pos2: target.PosVec, Frame68: 0x87654321, Field76: 0x55aa, Flags88: 0x77556601}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(f.Record.C()) <= 0xffffffff || uintptr(target.CObj()) <= 0xffffffff) {
		t.Fatal("native C-owned glyph links must exceed4GiB")
	}
	expected := *f
	beforeHealth, beforeTarget := *health, *target
	callback := legacy.Get_sub_52F220()
	if got := s.spells.duration.callCreate4FEBA0(callback, &f.Record); got != 1 || health.Cur != 26 {
		t.Fatalf("ordinary native glyph control return=%d HP=%d, want1/26", got, health.Cur)
	}
	t.Logf("ordinary glyph record=%p target=%p nilcaster/mode1 healed6->26 through production search and setter", &f.Record, target)
	*health, *target = beforeHealth, beforeTarget
	if got := ccall.CallIntPtr(callback, f.Record.C()); got != 1 || health.Cur != 26 {
		t.Fatalf("public C glyph create return=%d HP=%d; original/native want1/26", got, health.Cur)
	}
	raw := func(p *guarded) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p))) }
	if !bytes.Equal(raw(f), raw(&expected)) {
		t.Fatal("glyph create changed native duration/canary")
	}
	if health.Max != 100 || target.HealthData != health || target.UpdateData != unsafe.Pointer(update) {
		t.Fatal("native target links changed")
	}
	t.Logf("public C glyph record=%p target=%p healed6->26; production spatial search/HP and duration canaries verified", &f.Record, target)
}
