package opennox

import (
	"bytes"
	"encoding/binary"
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
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

// Public C calls compose with the real binary64 scalar/table reader, native
// facing/surface distance/LOS/index, registered DefaultDamage/UnitSetHP,
// player state/audio/mana and C important-packet sender. Units/definitions and
// gameplay flag1 are controlled. The observer delegates actual ray allocation;
// no gameplay service is replaced and no peer connection is opened.
func TestSpellEnergyBoltPublicCNativeComposition52E820(t *testing.T) {
	const childEnv = "NOX_ENERGY_BOLT_NATIVE_COMPOSITION_CHILD"
	mode := os.Getenv(childEnv)
	if mode == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		for _, mode := range []string{"player-sustain", "ray-switch", "cursor-acquire", "surface-boundary", "no-target", "recently-damaged", "monster-moved", "glyph-nil-caster"} {
			t.Run(mode, func(t *testing.T) {
				cmd := exec.Command(binary, "-test.run=^TestSpellEnergyBoltPublicCNativeComposition52E820$", "-test.v")
				cmd.Env = append(os.Environ(), childEnv+"="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil {
					t.Fatalf("actual Lightning C composition: %v", err)
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
	oldFlags, oldGameplay := noxflags.GetGame(), noxflags.GetGamePlay()
	noxflags.UnsetGame(oldFlags)
	noxflags.SetGamePlay(noxflags.GameplayFlag1)
	t.Cleanup(func() {
		noxflags.UnsetGame(noxflags.GetGame())
		noxflags.SetGame(oldFlags)
		noxflags.UnsetGamePlay(noxflags.GetGamePlay())
		noxflags.SetGamePlay(oldGameplay)
	})
	s.spells.duration.Init(s)
	t.Cleanup(s.spells.duration.Free)
	s.SetFrame(140)
	s.SetTickRate(30)
	s.Map.Init()
	t.Cleanup(s.Map.Free)
	if s.Walls.Init() == 0 {
		t.Fatal("native walls")
	}
	t.Cleanup(s.Walls.Free)
	s.Audio.Init(s.Server)
	t.Cleanup(s.Audio.Free)
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	dir, oldData := t.TempDir(), datapath.Data()
	datapath.SetData(dir)
	t.Cleanup(func() { datapath.SetData(oldData) })
	if err := os.WriteFile(filepath.Join(dir, "gamedata.yml"), []byte("LightningRange: 400\nEnergyBoltDamage: [1.5, 1.5, 1.5]\nEnergyBoltGlyphDamage: 3.5\nLightningSearchTime: 1.99999999\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := s.Balance.Read(); err != nil {
		t.Fatal(err)
	}
	if math.Float64bits(s.Balance.Float("LightningSearchTime")) != math.Float64bits(1.99999999) || s.Balance.FloatInd("EnergyBoltDamage", 2) != 1.5 {
		t.Fatal("production binary64 scalar/table reader")
	}
	if err := s.Types.ReadObjectType(&things.Thing{Name: "NativeLightningVictim", OnDamage: &things.ProcFunc{Name: "DefaultDamage"}}); err != nil {
		t.Fatal(err)
	}
	if !s.Objs.Init(16) {
		t.Fatal("native objects")
	}
	t.Cleanup(s.Objs.FreeObjects)
	newObj := func(typ *server.ObjectType, cls object.Class, code uint32) *server.Object {
		o := s.Objs.NewObject(typ)
		o.ObjClass = cls
		o.ObjFlags = 0
		o.NetCode = code
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(o.CObj()) <= math.MaxUint32 {
			t.Fatal("native C object below 4 GiB")
		}
		return o
	}
	source := newObj(&server.ObjectType{}, object.ClassPlayer, 0x1234)
	target := newObj(s.Types.ByID("NativeLightningVictim"), object.ClassMonster, 0x2345)
	previous := newObj(&server.ObjectType{}, object.ClassMonster, 0x3456)
	update := plasmaNativeAllocation531580[server.PlayerUpdateData](t)
	info := plasmaNativeAllocation531580[server.Player](t)
	monster := plasmaNativeAllocation531580[server.MonsterUpdateData](t)
	sourceMonster := plasmaNativeAllocation531580[server.MonsterUpdateData](t)
	health := plasmaNativeAllocation531580[server.HealthData](t)
	sourceHealth := plasmaNativeAllocation531580[server.HealthData](t)
	r := plasmaNativeAllocation531580[server.DurSpell](t)
	*info = server.Player{PlayerInd: 7}
	*update = server.PlayerUpdateData{Player: info, ManaCur: 3, ManaMax: 10, ManaPrev: 3, State: server.PlayerState13}
	*health = server.HealthData{Cur: 200, Max: 200, Field2: 200}
	*sourceHealth = server.HealthData{Cur: 200, Max: 200, Field2: 200}
	source.UpdateData = unsafe.Pointer(update)
	source.PosVec = types.Ptf(100, 200)
	source.PrevPos = types.Ptf(90, 190)
	source.Direction1 = 0
	target.UpdateData = unsafe.Pointer(monster)
	target.HealthData = health
	target.ObjSubClass = 0x10202
	target.PosVec = types.Ptf(200, 200)
	previous.PosVec = types.Ptf(220, 200)
	*r = server.DurSpell{ID: 0xabcd, Spell: uint32(spell.SPELL_LIGHTNING), Level: 3, Obj12: source, Caster16: source, Obj24: previous, Target48: target, Pos: source.PosVec, Frame60: 140, Frame68: 0x87654321, Field72: int32(math.Float32bits(.25)), Flags88: 0x77556601}
	if mode == "glyph-nil-caster" {
		r.Caster16 = nil
		r.Flag20 = 1
		r.Target48 = previous
	}
	if mode == "monster-moved" {
		source.ObjClass = object.ClassMonster
		source.UpdateData = unsafe.Pointer(sourceMonster)
	}
	rt := s.spells.duration.energyBoltRuntime52E820()
	if !rt.RoundTowardZero || rt.Balance64 == nil || rt.BalanceLevel64 == nil || rt.Distance == nil || rt.Selection == nil {
		t.Fatal("production startup/native services")
	}
	type sent struct {
		packet                              []byte
		recipient, remove, sequence, result int
	}
	var packets []sent
	productionSend := s.NetSendPacketXxx
	s.NetSendPacketXxx = func(recipient int, p []byte, related *server.Object, remove, sequence int) int {
		result := productionSend(recipient, p, related, remove, sequence)
		if len(p) == 7 && p[0] == 0x9e {
			packets = append(packets, sent{append([]byte(nil), p...), recipient, remove, sequence, result})
		}
		return result
	}
	t.Cleanup(func() { s.NetSendPacketXxx = productionSend })
	beforeCreate := *r
	if got := ccall.CallIntPtr(legacy.Get_nox_xxx_spellEnergyBoltStop_52E820(), r.C()); got != 0 || *r != beforeCreate {
		t.Fatal("actual native C create control", got)
	}
	wantTypes := []byte{}
	wantHP := uint16(200)
	wantMana := uint16(3)
	switch mode {
	case "ray-switch":
		s.spells.duration.durationRayTargets = map[*server.DurSpell]*server.Object{r: previous}
		wantTypes = append(wantTypes, 11)
	case "cursor-acquire":
		r.Target48 = nil
		update.CursorObj = target
		if !s.IsEnemyTo(source, target) {
			t.Fatal("cursor enemy prerequisite")
		}
	case "surface-boundary":
		source.Shape.Kind = server.ShapeKindCircle
		source.Shape.Circle.R = 10
		target.Shape.Kind = server.ShapeKindCircle
		target.Shape.Circle.R = 15
		target.PosVec = types.Ptf(510, 200)
		if distance := server.ObjectDistance4E6C00(source, target); distance >= 400 {
			t.Fatalf("surface distance %g must retain 410-unit center separation", distance)
		}
	case "no-target":
		r.Target48 = nil
		s.spells.duration.durationRayTargets = map[*server.DurSpell]*server.Object{r: previous}
	case "recently-damaged":
		r.Frame60 = 136
		source.Frame134 = 140
		source.HealthData = sourceHealth
	case "monster-moved":
		r.Pos.X -= 10
	case "glyph-nil-caster":
		target.NewPos = target.PosVec
		target.ObjFlags = object.FlagActive | object.FlagEnabled
		target.Shape.Kind = server.ShapeKindCircle
		target.Shape.Circle.R = 1
		s.Map.AddObjectToIndex(target)
		if !target.ObjFlags.Has(object.FlagPartitioned) {
			t.Fatal("real spatial index insertion")
		}
		t.Cleanup(func() { s.Map.RemoveObjectFromIndex(target) })
	}
	switch mode {
	case "player-sustain", "ray-switch", "cursor-acquire", "surface-boundary":
		wantTypes = append(wantTypes, 4)
		for hit, damage := range []uint16{1, 2, 1} {
			want := 0
			if hit == 2 {
				want = 1
			}
			got := ccall.CallIntPtr(legacy.Get_nox_xxx_spellEnergyBoltTick_52E850(), r.C())
			wantHP -= damage
			wantMana--
			carry := float32(.75)
			if hit == 1 {
				carry = .25
			}
			if got != want || health.Cur != wantHP || update.ManaCur != wantMana || r.Field72 != int32(math.Float32bits(carry)) || r.Frame68 != 141 || s.spells.duration.durationRayTargets[r] != target || update.State != 10 {
				t.Fatalf("C hit%d return%d/%d HP%d/%d mana%d/%d carry%#x expiry%d state%d", hit, got, want, health.Cur, wantHP, update.ManaCur, wantMana, uint32(r.Field72), r.Frame68, update.State)
			}
			if target.Obj130 != source || target.Field131 != 17 || target.Frame134 != 140 || monster.Field546 != 17 || monster.Field547 != 2 || !monster.StatusFlags.Has(object.MonStatusInjured) {
				t.Fatal("registered AirborneElectric damage attribution/update")
			}
		}
	case "no-target":
		if got := ccall.CallIntPtr(legacy.Get_nox_xxx_spellEnergyBoltTick_52E850(), r.C()); got != 0 || r.Target48 != nil || s.spells.duration.durationRayTargets[r] != nil {
			t.Fatal("no-target ray stop")
		}
		wantTypes = []byte{11}
	case "recently-damaged", "monster-moved":
		if got := ccall.CallIntPtr(legacy.Get_nox_xxx_spellEnergyBoltTick_52E850(), r.C()); got != 1 {
			t.Fatal("actual cancellation guard", got)
		}
	case "glyph-nil-caster":
		if got := ccall.CallIntPtr(legacy.Get_nox_xxx_spellEnergyBoltTick_52E850(), r.C()); got != 1 {
			t.Fatal("glyph completion", got)
		}
		wantHP = 197
		if health.Cur != wantHP || r.Target48 != previous || r.Frame68 != 0x87654321 || r.Field72 != int32(math.Float32bits(.25)) || s.spells.duration.energyBoltSelection.Target != target || s.spells.duration.energyBoltSelection.Mode != 1 || s.spells.duration.durationRayTargets[r] != nil || update.State != 13 {
			t.Fatal("glyph actual damage/shared selection/carry/expiry/native links")
		}
	default:
		t.Fatal("unknown mode")
	}
	if health.Cur != wantHP || update.ManaCur != wantMana {
		t.Fatal("unexpected damage/mana")
	}
	if len(packets) != len(wantTypes) {
		t.Fatalf("ray packets%d want%d", len(packets), len(wantTypes))
	}
	for i, p := range packets {
		first, second := uint16(source.NetCode), uint16(target.NetCode)
		if wantTypes[i] == 11 {
			first, second = uint16(previous.NetCode), uint16(source.NetCode)
		}
		want := []byte{0x9e, wantTypes[i], 3, 0, 0, 0, 0}
		binary.LittleEndian.PutUint16(want[3:5], first)
		binary.LittleEndian.PutUint16(want[5:7], second)
		if !bytes.Equal(p.packet, want) || p.recipient != 255 || p.remove != 1 || p.sequence != 1 || p.result != 1 {
			t.Fatalf("actual allocated ray packet %+v want%x", p, want)
		}
	}
	if r.Obj12 != source || r.Obj24 != previous || r.Flags88 != 0x77556601 || update.Player != info || target.UpdateData != unsafe.Pointer(monster) {
		t.Fatal("native links/canaries")
	}
	t.Logf("actual C EnergyBolt %s record=%p caster=%p target=%p HP200->%d mana3->%d rays%d startup=true", mode, r, source, target, health.Cur, update.ManaCur, len(packets))
}
