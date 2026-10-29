package opennox

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	playerlib "github.com/opennox/libs/player"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

//go:embed server/testdata/greater_heal_precision_original.json
var greaterHealCUpdateOriginal52F2E0 []byte

// The C export composes production visibility, native HP setter, mana debit
// and protection. Arithmetic/fraction/amount expectations come from the
// pre-existing complete original-update fixture; it is never rewritten here.
// Synthetic actors, empty geometry, explicit coefficients/multipliers and
// inert protection tokens bound this test; no earned campaign credit.
func TestGreaterHealPublicCUpdateNativeComposition52F2E0(t *testing.T) {
	const child = "NOX_GREATER_HEAL_NATIVE_UPDATE_CHILD"
	mode := os.Getenv(child)
	if mode == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		for _, mode := range []string{
			"sequence16_class0_level1_mult3f800000", "sequence16_class1_level3_mult40400000", "sequence16_class2_level5_mult3f800000",
			"stock_class0_level1_mult3f800000_fraction3f000001", "stock_class1_level1_mult40400000_fraction3f7fffff",
			"unknown-player-class_level2_fraction3f400000", "monster-no-multiplier_level3_fraction3f7fffff",
			"nil-target", "dead-target", "destroyed-target", "buffed-caster", "hidden-by-blindness", "no-mana", "moved-monster", "recent-damage", "full-health", "nonunit-no-mana",
		} {
			t.Run(mode, func(t *testing.T) {
				cmd := exec.Command(binary, "-test.run=^TestGreaterHealPublicCUpdateNativeComposition52F2E0$", "-test.v")
				cmd.Env = append(os.Environ(), child+"="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil {
					t.Fatalf("actual GreaterHeal C update: %v", err)
				}
			})
		}
		return
	}
	var fixture struct {
		Cases []struct {
			Name        string `json:"name"`
			CasterClass uint32 `json:"caster_class"`
			PlayerClass uint8  `json:"player_class"`
			Level       uint32 `json:"level"`
			Coefficient uint32 `json:"coefficient_bits32"`
			Fraction    uint32 `json:"fraction_bits32"`
			Multiplier  uint32 `json:"multiplier_bits32"`
			TargetNil   bool   `json:"target_nil"`
			TargetFlags uint32 `json:"target_flags"`
			Buffs       uint32 `json:"caster_buffs"`
			Mana        uint16 `json:"mana"`
			Moved       bool   `json:"moved"`
			Damaged     bool   `json:"damaged"`
			FullHealth  bool   `json:"full_health"`
			CanInteract bool   `json:"can_interact"`
			Steps       []struct {
				Result   int    `json:"result"`
				Fraction uint32 `json:"fraction_bits32"`
				Services []struct {
					Service string `json:"service"`
					Amount  int32  `json:"amount"`
				} `json:"services"`
			} `json:"expected_steps"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(greaterHealCUpdateOriginal52F2E0, &fixture); err != nil {
		t.Fatal(err)
	}
	index := -1
	for i, c := range fixture.Cases {
		if c.Name == mode {
			index = i
		}
	}
	if index < 0 {
		t.Fatal("sealed original update context")
	}
	c := fixture.Cases[index]
	handles.Init()
	t.Cleanup(handles.Release)
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	oldServer, oldGet := noxServer, legacy.GetServer
	noxServer, legacy.GetServer = s, func() legacy.Server { return s }
	t.Cleanup(func() { noxServer, legacy.GetServer = oldServer, oldGet })
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
	s.SetFrame(150)
	s.SetTickRate(30)
	s.Map.Init()
	t.Cleanup(s.Map.Free)
	if s.Walls.Init() == 0 {
		t.Fatal("native walls")
	}
	t.Cleanup(s.Walls.Free)
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	if err := s.Types.ReadObjectType(&things.Thing{Name: "NativeGreaterHealUpdateActor"}); err != nil {
		t.Fatal(err)
	}
	if !s.Objs.Init(8) {
		t.Fatal("native object pool")
	}
	t.Cleanup(s.Objs.FreeObjects)
	caster := s.Objs.NewObject(s.Types.ByID("NativeGreaterHealUpdateActor"))
	target := s.Objs.NewObject(s.Types.ByID("NativeGreaterHealUpdateActor"))
	ch, th := plasmaNativeAllocation531580[server.HealthData](t), plasmaNativeAllocation531580[server.HealthData](t)
	update := plasmaNativeAllocation531580[server.PlayerUpdateData](t)
	info := plasmaNativeAllocation531580[server.Player](t)
	targetUpdate := plasmaNativeAllocation531580[server.MonsterUpdateData](t)
	*ch, *th = server.HealthData{Cur: 1000, Max: 50000}, server.HealthData{Cur: 1000, Max: 50000}
	if c.FullHealth {
		th.Cur = th.Max
	}
	*info = server.Player{PlayerInd: 7}
	info.Info().SetPlayerClass(playerlib.Class(c.PlayerClass))
	*update = server.PlayerUpdateData{Player: info, ManaCur: c.Mana, ManaPrev: c.Mana, ManaMax: 1000}
	caster.ObjClass, caster.ObjFlags = object.Class(c.CasterClass), 0
	caster.PosVec, caster.NewPos, caster.VelVec = types.Ptf(300, 300), types.Ptf(300, 300), types.Pointf{}
	caster.HealthData, caster.UpdateData, caster.Buffs = ch, unsafe.Pointer(update), c.Buffs
	target.ObjClass, target.ObjFlags = object.ClassMonster, object.Flags(c.TargetFlags)
	target.PosVec, target.NewPos, target.VelVec = types.Ptf(310, 310), types.Ptf(310, 310), types.Pointf{}
	target.HealthData, target.UpdateData = th, unsafe.Pointer(targetUpdate)
	if c.Damaged {
		caster.Frame134 = 150
	}
	oldMult := s.Players.Mult
	t.Cleanup(func() { s.Players.Mult = oldMult })
	m := math.Float32frombits(c.Multiplier)
	s.Players.Mult.Warrior.Health, s.Players.Mult.Wizard.Health, s.Players.Mult.Conjurer.Health = m, m, m
	coefficient := memmap.PtrUint32(0x587000, 260360+uintptr(4*c.Level))
	oldCoefficient := *coefficient
	*coefficient = c.Coefficient
	t.Cleanup(func() { *coefficient = oldCoefficient })
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
	r := &f.Record
	*r = server.DurSpell{ID: 0xabcd, Spell: uint32(spell.SPELL_GREATER_HEAL), Level: c.Level, Obj12: target, Caster16: caster, Obj24: caster, Target48: target, Pos: types.Ptf(300, 300), Pos2: types.Ptf(310, 310), Frame60: 140, Frame64: 141, Frame68: 0x87654321, Field72: int32(c.Fraction), Field76: 0x55aa, Flags88: 0x77556601}
	if c.TargetNil {
		r.Target48 = nil
	}
	if c.Moved {
		r.Pos.X = 320
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(r.C()) <= 0xffffffff || uintptr(caster.CObj()) <= 0xffffffff || uintptr(target.CObj()) <= 0xffffffff) {
		t.Fatal("native C-owned update links must exceed4GiB")
	}
	if !c.TargetNil && s.CanInteract(caster, target, 0) != c.CanInteract {
		t.Fatal("original declared empty-geometry visibility")
	}
	raw := func(p *guarded) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p))) }
	expectedHP, expectedMana := th.Cur, update.ManaCur
	for tick, step := range c.Steps {
		expected := *f
		expected.Record.Field72 = int32(step.Fraction)
		oldMana := expectedMana
		for _, event := range step.Services {
			if event.Service == "adjust_HP" {
				expectedHP = uint16(int32(expectedHP) + event.Amount)
			}
			if event.Service == "mana_sub" && c.CasterClass&4 != 0 {
				expectedMana--
			}
		}
		got := ccall.CallIntPtr(legacy.Get_sub_52F2E0(), r.C())
		if got != step.Result || th.Cur != expectedHP || update.ManaCur != expectedMana {
			t.Fatalf("tick%d Creturn/HP/mana=%d/%d/%d original+native effects=%d/%d/%d", tick, got, th.Cur, update.ManaCur, step.Result, expectedHP, expectedMana)
		}
		if !bytes.Equal(raw(f), raw(&expected)) {
			t.Fatalf("tick%d changed native duration/canary outside original fraction", tick)
		}
		if step.Result == 0 && c.CasterClass&4 != 0 && update.ManaPrev != oldMana {
			t.Fatal("actual mana previous word")
		}
	}
	if th.Max != 50000 || caster.HealthData != ch || target.HealthData != th || caster.UpdateData != unsafe.Pointer(update) || update.Player != info {
		t.Fatal("native actor/health/player links changed")
	}
	t.Logf("original case=%s public C record=%p caster=%p target=%p ticks=%d HP=%d mana=%d fraction=%#x; production visibility/HP/mana/protection and full duration canaries verified", mode, r, caster, target, len(c.Steps), th.Cur, update.ManaCur, uint32(r.Field72))
}
