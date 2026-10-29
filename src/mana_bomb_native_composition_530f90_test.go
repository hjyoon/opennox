package opennox

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"path/filepath"
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
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

//go:embed server/testdata/spell_mana_bomb_native_original_530f90.json
var manaBombCOriginal530F90 []byte

// Actual public C calls compose the type cache/factory, CreateObjectAt,
// native sidecar, buff/mana/protection and radial map/ray/C-damage paths.
// Each stage loads its own frozen original input before C and checks that
// original output; this does not claim an original whole-world replay.
// The one inner-radius target captures damage rather than changing HP.
// FX/quake run with no connected clients, so no packet-delivery claim is made.
func TestManaBombPublicCNativeComposition530F90(t *testing.T) {
	const child = "NOX_MANA_BOMB_NATIVE_COMPOSITION_CHILD"
	mode := os.Getenv(child)
	if mode == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		for _, mode := range []string{"glyph-idle", "glyph-explode", "player-charge", "player-glyph-zero-mana", "player-cancel", "glyph-raw-cancel"} {
			t.Run(mode, func(t *testing.T) {
				cmd := exec.Command(binary, "-test.run=^TestManaBombPublicCNativeComposition530F90$", "-test.v")
				cmd.Env = append(os.Environ(), child+"="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil {
					t.Fatalf("actual Mana Bomb C composition: %v", err)
				}
			})
		}
		return
	}
	var fixture struct {
		Cases []struct {
			Input  map[string]any `json:"input"`
			Result map[string]any `json:"result"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(manaBombCOriginal530F90, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 1969 {
		t.Fatal("sealed original corpus")
	}
	u := func(m map[string]any, k string) uint32 { return uint32(m[k].(float64)) }
	pt := func(v any) types.Pointf {
		p := v.([]any)
		return types.Ptf(math.Float32frombits(uint32(p[0].(float64))), math.Float32frombits(uint32(p[1].(float64))))
	}
	bits := func(p types.Pointf) [2]uint32 { return [2]uint32{math.Float32bits(p.X), math.Float32bits(p.Y)} }
	recordMode, cls := uint32(1), uint32(2)
	if mode == "player-charge" || mode == "player-cancel" {
		recordMode, cls = 0, 4
	}
	if mode == "player-glyph-zero-mana" {
		cls = 4
	}
	if mode == "glyph-raw-cancel" {
		recordMode = math.MaxUint32
	}
	destroyCharge := true
	selectCase := func(function string, frame, mana uint32) int {
		for i, row := range fixture.Cases {
			c := row.Input
			if c["function"] != function || c["mutation"].(map[string]any)["stage"] != "" || u(c, "mode") != recordMode || u(c, "class") != cls || u(c, "level") != 3 || c["caster_nil"].(bool) || c["charge"].(bool) != (function != "destroy" || destroyCharge) || !c["allocate"].(bool) || len(c["missing"].([]any)) != 0 {
				continue
			}
			if function == "create" && u(c, "cache") == 0 {
				return i
			}
			if function == "update" && u(c, "frame") == frame && u(c, "expiry") == 120 && u(c, "mana") == mana {
				return i
			}
			if function == "destroy" && u(c, "exploded") == 0 {
				return i
			}
		}
		t.Fatalf("no frozen original %s context", function)
		return -1
	}
	createIndex := selectCase("create", 100, 20)
	handles.Init()
	t.Cleanup(handles.Release)
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	oldServer, oldGet, oldWalls := noxServer, legacy.GetServer, doDamageWalls
	noxServer, legacy.GetServer = s, func() legacy.Server { return s }
	t.Cleanup(func() { noxServer, legacy.GetServer, doDamageWalls = oldServer, oldGet, oldWalls })
	oldFlags, oldEngine := noxflags.GetGame(), noxflags.GetEngine()
	noxflags.UnsetGame(oldFlags)
	noxflags.UnsetEngine(noxflags.EngineGodMode)
	t.Cleanup(func() {
		noxflags.UnsetGame(noxflags.GetGame())
		noxflags.SetGame(oldFlags)
		noxflags.UnsetEngine(noxflags.GetEngine())
		noxflags.SetEngine(oldEngine)
	})
	s.spells.duration.Init(s)
	t.Cleanup(s.spells.duration.Free)
	s.SetFrame(100)
	s.SetTickRate(60)
	s.Map.Init()
	t.Cleanup(s.Map.Free)
	if s.Walls.Init() == 0 {
		t.Fatal("native walls")
	}
	t.Cleanup(s.Walls.Free)
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	for _, name := range []string{"NativeManaBombActor", "ManaBombCharge"} {
		if err := s.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
			t.Fatal(err)
		}
	}
	if !s.Objs.Init(16) {
		t.Fatal("native object pool")
	}
	t.Cleanup(s.Objs.FreeObjects)
	cache := memmap.PtrUint32(0x5D4594, 2487920)
	oldCache := *cache
	*cache = 0
	t.Cleanup(func() { *cache = oldCache })
	dir, oldData := t.TempDir(), datapath.Data()
	datapath.SetData(dir)
	t.Cleanup(func() { datapath.SetData(oldData) })
	config := "ManaBombInitPower: [10.75, 10.75, 10.75, 10.75, 10.75]\nManaBombGlyphDuration: 12.75\nManaBombInRadius: 20.25\nManaBombOutRadius: 60.5\nManaBombShakeMag: 4.75\nManaBombDeltaPower: [3.75, 3.75, 3.75, 3.75, 3.75]\n"
	if err := os.WriteFile(filepath.Join(dir, "gamedata.yml"), []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	if err := s.Balance.Read(); err != nil {
		t.Fatal(err)
	}
	s.Audio.Init(s.Server)
	t.Cleanup(s.Audio.Free)
	name := sound.ID(81).String()
	wire := binary.LittleEndian.AppendUint32(nil, 1)
	wire = append(wire, byte(len(name)))
	wire = append(wire, name...)
	wire = binary.LittleEndian.AppendUint16(wire, 8)
	wire = append(wire, 0)
	wire = binary.LittleEndian.AppendUint16(wire, 40)
	wire = append(wire, 0, 0, 0, 0, 1, 's', 0)
	file := luaCompatAudioFile{bytes.NewReader(wire)}
	if !s.Audio.ReadAUD(file) || file.Len() != 0 || s.Audio.Field12(81) != 1 {
		t.Fatal("real AUD descriptor")
	}
	s.Audio.Reset()
	caster := s.Objs.NewObject(s.Types.ByID("NativeManaBombActor"))
	info, update := plasmaNativeAllocation531580[server.Player](t), plasmaNativeAllocation531580[server.PlayerUpdateData](t)
	*info = server.Player{PlayerInd: 7}
	*update = server.PlayerUpdateData{Player: info, ManaCur: 20, ManaPrev: 20, ManaMax: 100}
	caster.ObjClass, caster.ObjFlags = object.Class(cls), 0
	caster.UpdateData = unsafe.Pointer(update)
	caster.PosVec, caster.NewPos = types.Ptf(123.125, -71.25), types.Ptf(123.125, -71.25)
	caster.Mass, caster.VelVec, caster.ForceVec, caster.Pos24 = 17.5, types.Ptf(1, 2), types.Ptf(3, 4), types.Ptf(5, 6)
	type guarded struct {
		Before [32]byte
		Record server.DurSpell
		After  [32]byte
	}
	f := plasmaNativeAllocation531580[guarded](t)
	for i := range f.Before {
		f.Before[i], f.After[i] = byte(0x51+i), byte(0xe7-i)
	}
	r := &f.Record
	*r = server.DurSpell{ID: 0xabcd, Spell: uint32(spell.SPELL_MANA_BOMB), Caster16: caster, Obj12: caster, Obj24: caster, Flags88: 0x77556601, Field76: 0x3fdccccc}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(r.C()) <= math.MaxUint32 || uintptr(caster.CObj()) <= math.MaxUint32) {
		t.Fatal("native C-owned links above4GiB")
	}
	raw := func(p *guarded) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p))) }
	prepare := func(i int) {
		c := fixture.Cases[i].Input
		r.Level, r.Flag20, r.Pos, r.Frame68 = u(c, "level"), u(c, "mode"), pt(c["pos"]), u(c, "expiry")
		r.Field72, r.Field80, r.Field84 = int32(u(c, "power")), u(c, "saved_mass"), u(c, "exploded")
		s.SetFrame(u(c, "frame"))
		update.ManaCur = uint16(u(c, "mana"))
	}
	expect := func(i int) guarded {
		e := *f
		st := fixture.Cases[i].Result["state"].(map[string]any)
		e.Record.Field72, e.Record.Frame68, e.Record.Field80, e.Record.Field84 = int32(u(st, "power")), u(st, "expiry"), u(st, "saved_mass"), u(st, "exploded")
		if fixture.Cases[i].Input["function"] == "create" {
			e.Record.Field76 = 0
		}
		return e
	}
	prepare(createIndex)
	expected := expect(createIndex)
	if got := ccall.CallIntPtr(legacy.Get_nox_xxx_manaBomb_530F90(), r.C()); got != 0 {
		t.Fatalf("C create=%d", got)
	}
	charge := s.spells.duration.manaBombCharges[r]
	if charge == nil || int(charge.TypeInd) != s.Types.IndByID("ManaBombCharge") || *cache != uint32(charge.TypeInd) || charge.ObjOwner != nil || charge.Flags().Has(object.FlagDestroyed) {
		t.Fatal("real native cache/factory/ownership/sidecar")
	}
	if bits(charge.PosVec) != bits(pt(fixture.Cases[createIndex].Result["state"].(map[string]any)["charge_pos"])) || !bytes.Equal(raw(f), raw(&expected)) {
		t.Fatal("original create fields/position or canary")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(charge.CObj()) <= math.MaxUint32 {
		t.Fatal("native charge above4GiB")
	}
	if recordMode == 0 {
		for _, buff := range []server.EnchantID{5, 14, 29} {
			if !caster.HasEnchant(buff) || caster.EnchantDur(buff) != 600 || caster.EnchantPower(buff) != 5 {
				t.Fatal("real buff application")
			}
		}
		if math.Float32bits(caster.Mass) != 1203982323 || caster.VelVec != (types.Pointf{}) || caster.ForceVec != (types.Pointf{}) || caster.Pos24 != (types.Pointf{}) {
			t.Fatal("original player mass and vectors")
		}
	}
	var target *server.Object
	if mode == "glyph-explode" || mode == "player-glyph-zero-mana" {
		radialDamageCaptureSetup4E25B0()
		target = s.Objs.NewObject(s.Types.ByID("NativeManaBombActor"))
		target.PosVec, target.NewPos = types.Ptf(101, 200), types.Ptf(101, 200)
		target.ObjClass, target.ObjFlags = object.ClassSimple, object.FlagActive
		target.Shape.Kind, target.Shape.Circle.R, target.Shape.Circle.R2 = server.ShapeKindCircle, 1, 1
		target.Damage = legacy.RadialDamageCapturePointer4E25B0(false)
		s.Map.AddObjectToIndex(target)
		t.Cleanup(func() { s.Map.RemoveObjectFromIndex(target) })
		if !s.MapTraceRay(r.Pos, target.PosVec, server.MapTraceFlag1) {
			t.Fatal("real map ray precondition")
		}
	}
	audioCalls := 0
	s.Audio.OnSound(func(id sound.ID, kind int, obj *server.Object, pos types.Pointf) {
		if id != 81 {
			return
		}
		if kind != 0 || obj != caster || bits(pos) != bits(caster.PosVec) {
			t.Fatal("real audio arguments")
		}
		audioCalls++
	})
	calls := 1
	if mode != "player-cancel" && mode != "glyph-raw-cancel" {
		frame, mana := uint32(111), uint32(20)
		if mode == "glyph-explode" {
			frame = 119
		}
		if mode == "player-charge" {
			frame = 100
		}
		if mode == "player-glyph-zero-mana" {
			frame, mana = 100, 0
		}
		i := selectCase("update", frame, mana)
		prepare(i)
		r.Field76 = 0x3fdccccc
		expected = expect(i)
		if got := ccall.CallIntPtr(legacy.Get_nox_xxx_manaBombBoom_5310C0(), r.C()); got != int(u(fixture.Cases[i].Result, "return_value")) {
			t.Fatalf("C update=%d", got)
		}
		calls++
		if !bytes.Equal(raw(f), raw(&expected)) {
			t.Fatal("original update fields/canary")
		}
		st := fixture.Cases[i].Result["state"].(map[string]any)
		wantCharge := int(st["charge"].(float64)) >= 0
		destroyCharge = wantCharge
		if (s.spells.duration.manaBombCharges[r] != nil) != wantCharge || charge.Flags().Has(object.FlagDestroyed) == wantCharge {
			t.Fatal("original charge retirement through real deletion list")
		}
		if !wantCharge && s.Objs.DeletedList != charge {
			t.Fatal("real charge deletion list")
		}
		for _, ev := range fixture.Cases[i].Result["events"].([]any) {
			e := ev.(map[string]any)
			a := e["args"].([]any)
			switch e["kind"] {
			case "mana":
				if update.ManaPrev != uint16(mana) || update.ManaCur != uint16(mana-uint32(a[1].(float64))) {
					t.Fatal("real mana subtraction/protection")
				}
			case "damage":
				count, damage, typ, gotTarget, source, weapon := legacy.RadialDamageCaptureResult4E25B0()
				if count != 1 || uint32(damage) != uint32(a[3].(float64)) || typ != 15 || gotTarget != target || source != caster || weapon != nil {
					t.Fatalf("real map/C damage count=%d damage=%d type=%d target=%p source=%p", count, damage, typ, gotTarget, source)
				}
			}
		}
		if audioCalls != int(u(st, "exploded")) {
			t.Fatal("real explosion AUD event")
		}
	}
	i := selectCase("destroy", 100, 20)
	prepare(i)
	r.Field76 = 0x3fdccccc
	expected = expect(i)
	ccall.CallVoidPtr(legacy.Get_sub_531290(), r.C())
	calls++
	if s.spells.duration.manaBombCharges[r] != nil || !charge.Flags().Has(object.FlagDestroyed) || !bytes.Equal(raw(f), raw(&expected)) {
		t.Fatal("real C destroy sidecar/deletion/canary")
	}
	if recordMode == 0 {
		for _, buff := range []server.EnchantID{5, 14, 29} {
			if caster.HasEnchant(buff) {
				t.Fatal("real buff removal")
			}
		}
		if math.Float32bits(caster.Mass) != u(fixture.Cases[i].Result["state"].(map[string]any)["actors"].([]any)[0].(map[string]any), "mass") {
			t.Fatal("original mass restore")
		}
	}
	t.Logf("C mode=%s calls=%d record=%p charge=%p target=%p audio=%d cache=%d; real factory/buff/mana/map-damage/delete and native canaries verified", mode, calls, r, charge, target, audioCalls, *cache)
}
