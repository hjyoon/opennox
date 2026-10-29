package opennox

import (
	"bytes"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/datapath"
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

// Both actual C entries use real balance/type caches, typed allocation,
// native collision records, create-at/ownership, active-list/index activation
// and delayed deletion. The object definitions and blank map are synthetic;
// the stock hosted fixture separately exercises actual projectile updates.
func TestSpellTurnUndeadPublicCNativeComposition531310(t *testing.T) {
	const childEnv = "NOX_TURN_UNDEAD_NATIVE_COMPOSITION_CHILD"
	mode := os.Getenv(childEnv)
	if mode == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		modes := []string{"caster:cold", "caster:warm", "anchor:cold", "anchor:warm", "glyph:cold", "glyph:warm", "allocation-fail"}
		for _, mode := range modes {
			t.Run(mode, func(t *testing.T) {
				cmd := exec.Command(binary, "-test.run=^TestSpellTurnUndeadPublicCNativeComposition531310$", "-test.v")
				cmd.Env = append(os.Environ(), childEnv+"="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil {
					t.Fatalf("actual C TurnUndead composition: %v", err)
				}
			})
		}
		return
	}
	parts := strings.Split(mode, ":")
	family := parts[0]
	warm := len(parts) == 2 && parts[1] == "warm"
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	oldServer, oldGet := noxServer, legacy.GetServer
	noxServer, legacy.GetServer = s, func() legacy.Server { return s }
	t.Cleanup(func() { noxServer, legacy.GetServer = oldServer, oldGet })
	oldFlags := noxflags.GetGame()
	noxflags.UnsetGame(oldFlags)
	t.Cleanup(func() { noxflags.UnsetGame(noxflags.GetGame()); noxflags.SetGame(oldFlags) })
	s.spells.duration.Init(s)
	s.SetFrame(142)
	s.SetTickRate(30)
	s.Map.Init()
	t.Cleanup(s.Map.Free)
	s.Audio.Init(s.Server)
	t.Cleanup(s.Audio.Free)
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	dir, oldData := t.TempDir(), datapath.Data()
	datapath.SetData(dir)
	t.Cleanup(func() { datapath.SetData(oldData) })
	if err := os.WriteFile(filepath.Join(dir, "gamedata.yml"), []byte("TurnUndeadKillPoints: [3.9]\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := s.Balance.Read(); err != nil {
		t.Fatal(err)
	}
	if s.Balance.FloatInd("TurnUndeadKillPoints", 0) != 3.9 {
		t.Fatal("actual binary64 array reader")
	}
	if family != "allocation-fail" {
		for _, name := range []string{"UndeadKiller", "OtherTurnUndeadControl"} {
			if err := s.Types.ReadObjectType(&things.Thing{Name: name, Class: []things.Class{"SIMPLE"}, OnCollide: &things.ProcFunc{Name: "UndeadKillerCollide"}}); err != nil {
				t.Fatal(err)
			}
		}
	}
	createCache, destroyCache := memmap.PtrUint32(0x5D4594, 2487924), memmap.PtrUint32(0x5D4594, 2487928)
	oldCreate, oldDestroy := *createCache, *destroyCache
	t.Cleanup(func() { *createCache, *destroyCache = oldCreate, oldDestroy })
	*createCache, *destroyCache = 0, 0
	typeID := uint32(s.Types.IndByID("UndeadKiller"))
	if warm {
		*createCache, *destroyCache = typeID, typeID
	}
	type recordFrame struct {
		Before [32]byte
		Record server.DurSpell
		After  [32]byte
	}
	type sourceFrame struct {
		Before [32]byte
		Object server.Object
		After  [32]byte
	}
	f, freeRecord := alloc.New(recordFrame{})
	t.Cleanup(freeRecord)
	*f = recordFrame{}
	other, freeOther := alloc.New(server.DurSpell{})
	t.Cleanup(freeOther)
	*other = server.DurSpell{}
	caster, freeCaster := alloc.New(sourceFrame{})
	t.Cleanup(freeCaster)
	*caster = sourceFrame{}
	anchor, freeAnchor := alloc.New(sourceFrame{})
	t.Cleanup(freeAnchor)
	*anchor = sourceFrame{}
	for i := range f.Before {
		f.Before[i], f.After[i] = byte(0x51+i), byte(0xe7-i)
		caster.Before[i], caster.After[i] = byte(0x31+i), byte(0xc7-i)
		anchor.Before[i], anchor.After[i] = byte(0x71+i), byte(0xa7-i)
	}
	caster.Object = server.Object{ObjClass: object.ClassSimple, PosVec: types.Ptf(100, 200)}
	anchor.Object = server.Object{ObjClass: object.ClassSimple, PosVec: types.Ptf(300, 400)}
	f.Record = server.DurSpell{ID: 0xabcd, Spell: uint32(spell.SPELL_TURN_UNDEAD), Level: 1, Obj12: &caster.Object, Caster16: &caster.Object, Obj24: &anchor.Object, Field72: 0x55aa, Pos: types.Ptf(11.5, -22.25), Frame68: 0x87654321, Flags88: 0x77556601}
	point, owner := caster.Object.PosVec, &caster.Object
	if family == "anchor" || family == "glyph" {
		f.Record.Flag20 = 1
		point = anchor.Object.PosVec
	}
	if family == "glyph" {
		f.Record.Caster16 = nil
		owner = nil
	}
	if !s.Objs.Init(64) {
		t.Fatal("native allocator")
	}
	t.Cleanup(s.Objs.FreeObjects)
	raw := func(p unsafe.Pointer, n uintptr) []byte { return unsafe.Slice((*byte)(p), n) }
	expected := *f
	expected.Record.Field72 = 3
	casterBefore := append([]byte(nil), raw(unsafe.Pointer(caster), unsafe.Sizeof(*caster))...)
	anchorBefore := append([]byte(nil), raw(unsafe.Pointer(anchor), unsafe.Sizeof(*anchor))...)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, p := range []unsafe.Pointer{f.Record.C(), caster.Object.CObj(), anchor.Object.CObj(), other.C()} {
			if uintptr(p) <= 0xffffffff {
				t.Fatal("native C-owned source/duration must be above4GiB")
			}
		}
	}
	if got := ccall.CallIntPtr(legacy.Get_nox_xxx_spellTurnUndeadCreate_531310(), f.Record.C()); got != 0 {
		t.Fatalf("actual C create=%d", got)
	}
	if *createCache != typeID || !bytes.Equal(raw(unsafe.Pointer(f), unsafe.Sizeof(*f)), raw(unsafe.Pointer(&expected), unsafe.Sizeof(expected))) {
		t.Fatal("actual cache/budget and full record canary")
	}
	var killers [43]*server.Object
	count := 0
	for o := s.Objs.Pending; o != nil; o = o.ObjNext {
		if count >= 43 || o.TypeInd != uint16(typeID) || o.CollideData == nil || (*server.UndeadKillerCollideData)(o.CollideData).Spell != &f.Record || o.ObjOwner != owner || o.PosVec != point || o.Direction1 != o.Direction2 || uint16(o.Direction1)%6 != 0 || uint16(o.Direction1) > 252 || o.Float28 != 0 || math.Abs(math.Hypot(float64(o.VelVec.X), float64(o.VelVec.Y))-4) > 1e-5 {
			t.Fatal("actual native projectile/collision/owner state")
		}
		i := int(o.Direction1) / 6
		if killers[i] != nil {
			t.Fatal("duplicate direction")
		}
		killers[i] = o
		count++
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(o.CObj()) <= 0xffffffff || uintptr(o.CollideData) <= 0xffffffff) {
			t.Fatal("native projectile and collision links must be above4GiB")
		}
	}
	if family == "allocation-fail" {
		if count != 0 || typeID != 0 {
			t.Fatal("unknown type must allocate no projectiles")
		}
	} else if count != 43 {
		t.Fatalf("created%d want43", count)
	}
	var wrongType, wrongSpell *server.Object
	if family != "allocation-fail" {
		wrongType = s.NewObjectByTypeID("OtherTurnUndeadControl")
		wrongSpell = s.NewObjectByTypeID("UndeadKiller")
		if wrongType == nil || wrongSpell == nil {
			t.Fatal("actual unrelated controls")
		}
		(*server.UndeadKillerCollideData)(wrongType.CollideData).Spell = &f.Record
		(*server.UndeadKillerCollideData)(wrongSpell.CollideData).Spell = other
		s.CreateObjectAt(wrongType, nil, types.Ptf(500, 600))
		s.CreateObjectAt(wrongSpell, nil, types.Ptf(700, 800))
		s.ObjectsAddPending()
		if s.Objs.Pending != nil {
			t.Fatal("actual pending activation")
		}
	}
	ccall.CallVoidPtr(legacy.Get_nox_xxx_spellTurnUndeadDelete_531420(), f.Record.C())
	if *destroyCache != typeID {
		t.Fatal("actual destroy cache")
	}
	deleted := 0
	for o := s.Objs.DeletedList; o != nil; o = o.DeletedNext {
		deleted++
	}
	if deleted != count {
		t.Fatalf("delayed deleted=%d want=%d", deleted, count)
	}
	for _, o := range killers {
		if o != nil && !o.Flags().Has(object.FlagDestroyed) {
			t.Fatal("matching active projectile not retired")
		}
	}
	if wrongType != nil && (wrongType.Flags().Has(object.FlagDestroyed) || wrongSpell.Flags().Has(object.FlagDestroyed)) {
		t.Fatal("type and spell identity filters must preserve unrelated objects")
	}
	if !bytes.Equal(raw(unsafe.Pointer(f), unsafe.Sizeof(*f)), raw(unsafe.Pointer(&expected), unsafe.Sizeof(expected))) {
		t.Fatal("destroy changed duration record/canary")
	}
	allowedOffset := unsafe.Offsetof(sourceFrame{}.Object) + unsafe.Offsetof(server.Object{}.Field129)
	allowedSize := unsafe.Sizeof(server.Object{}.Field129)
	for _, pair := range []struct {
		source *sourceFrame
		before []byte
	}{{caster, casterBefore}, {anchor, anchorBefore}} {
		for i, value := range raw(unsafe.Pointer(pair.source), unsafe.Sizeof(*pair.source)) {
			allowed := pair.source == caster && owner != nil && uintptr(i) >= allowedOffset && uintptr(i) < allowedOffset+allowedSize
			if value != pair.before[i] && !allowed {
				t.Fatalf("nonoutput source frame changed at%d", i)
			}
		}
	}
	t.Logf("TURN UNDEAD PUBLIC C COMPOSITION PASS:mode=%s record=%p projectiles=%d deleted=%d cache=%d/%d", mode, &f.Record, count, deleted, *createCache, *destroyCache)
}
