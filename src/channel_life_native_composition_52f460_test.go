package opennox

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/object"
	playerlib "github.com/opennox/libs/player"
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

//go:embed server/testdata/channel_life_precision_original.json
var channelLifeCOriginal52F460 []byte

// Arithmetic and effect amounts come from the unchanged complete original
// fixture. The public C entry composes production balance lookup, AddMana,
// ClearDamage, HP setter and inert-token protection on native-owned actors.
// Multi-tick actors remain away from the HP/mana termination gates excluded
// by the original fixture's nonmutating effects. No earned campaign credit.
func TestChannelLifePublicCNativeComposition52F460(t *testing.T) {
	const child = "NOX_CHANNEL_LIFE_NATIVE_COMPOSITION_CHILD"
	mode := os.Getenv(child)
	if mode == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		for _, mode := range []string{
			"sequence16_target4_caster4_level1", "sequence16_target4_caster2_level3",
			"sequence16_target6_caster4_level5", "sequence16_target2_caster4_level2",
			"nil-target", "dead-target", "destroyed-target", "glyph-one", "glyph-raw",
			"buffed-caster", "moved-monster", "full-mana", "zero-HP", "one-HP", "nil-caster", "missing-lookup",
			"exact_target_class_00000000", "exact_target_class_00010002", "exact_target_class_00010004", "exact_target_class_00010006",
		} {
			t.Run(mode, func(t *testing.T) {
				cmd := exec.Command(binary, "-test.run=^TestChannelLifePublicCNativeComposition52F460$", "-test.v")
				cmd.Env = append(os.Environ(), child+"="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil {
					t.Fatalf("actual ChannelLife C composition: %v", err)
				}
			})
		}
		return
	}
	var fixture struct {
		Contexts int `json:"contexts"`
		Cases    []struct {
			Name        string `json:"name"`
			CasterClass uint32 `json:"caster_class"`
			TargetClass uint32 `json:"target_class"`
			Level       uint32 `json:"level"`
			Coefficient uint32 `json:"coefficient_bits32"`
			Fraction    uint32 `json:"fraction_bits32"`
			TargetNil   bool   `json:"target_nil"`
			CasterNil   bool   `json:"caster_nil"`
			TargetFlags uint32 `json:"target_flags"`
			Buffs       uint32 `json:"caster_buffs"`
			Mana        uint16 `json:"mana"`
			MaxMana     uint16 `json:"maximum_mana"`
			HP          uint16 `json:"caster_HP"`
			Mode        uint32 `json:"mode"`
			Moved       bool   `json:"moved"`
			Missing     bool   `json:"lookup_missing"`
			Steps       []struct {
				Result   int    `json:"result"`
				Fraction uint32 `json:"fraction_bits32"`
				Services []struct {
					Service  string `json:"service"`
					Amount16 int16  `json:"amount16"`
					Amount   int32  `json:"amount"`
				} `json:"services"`
			} `json:"expected_steps"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(channelLifeCOriginal52F460, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Contexts != 322 || len(fixture.Cases) != 322 {
		t.Fatal("sealed complete original fixture")
	}
	index := -1
	for i, c := range fixture.Cases {
		if c.Name == mode {
			index = i
		}
	}
	if index < 0 {
		t.Fatal("sealed original context")
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
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	if err := s.Types.ReadObjectType(&things.Thing{Name: "NativeChannelLifeActor"}); err != nil {
		t.Fatal(err)
	}
	if !s.Objs.Init(8) {
		t.Fatal("native object pool")
	}
	t.Cleanup(s.Objs.FreeObjects)
	caster := s.Objs.NewObject(s.Types.ByID("NativeChannelLifeActor"))
	target := s.Objs.NewObject(s.Types.ByID("NativeChannelLifeActor"))
	ch, th := plasmaNativeAllocation531580[server.HealthData](t), plasmaNativeAllocation531580[server.HealthData](t)
	cu, tu := plasmaNativeAllocation531580[server.PlayerUpdateData](t), plasmaNativeAllocation531580[server.PlayerUpdateData](t)
	ci, ti := plasmaNativeAllocation531580[server.Player](t), plasmaNativeAllocation531580[server.Player](t)
	*ci, *ti = server.Player{PlayerInd: 6}, server.Player{PlayerInd: 7}
	ci.Info().SetPlayerClass(playerlib.Wizard)
	ti.Info().SetPlayerClass(playerlib.Wizard)
	*cu = server.PlayerUpdateData{Player: ci, ManaCur: c.Mana, ManaPrev: c.Mana, ManaMax: c.MaxMana}
	*tu = server.PlayerUpdateData{Player: ti, ManaCur: c.Mana, ManaPrev: c.Mana, ManaMax: c.MaxMana}
	*ch, *th = server.HealthData{Cur: c.HP, Max: 50000}, server.HealthData{Cur: 1000, Max: 50000}
	caster.ObjClass, caster.ObjFlags = object.Class(c.CasterClass), 0
	caster.PosVec, caster.NewPos, caster.VelVec = types.Ptf(300, 300), types.Ptf(300, 300), types.Pointf{}
	caster.HealthData, caster.UpdateData, caster.Buffs = ch, unsafe.Pointer(cu), c.Buffs
	target.ObjClass, target.ObjFlags = object.Class(c.TargetClass), object.Flags(c.TargetFlags)
	target.PosVec, target.NewPos, target.VelVec = types.Ptf(300, 300), types.Ptf(300, 300), types.Pointf{}
	target.HealthData, target.UpdateData = th, unsafe.Pointer(tu)
	// Both Player and Monster paths use valid native-sized backing storage.
	// Only Player effects access this PlayerUpdateData; Monster getters return
	// their original constants and the positive HP path uses no monster AI.
	dir, oldData := t.TempDir(), datapath.Data()
	datapath.SetData(dir)
	t.Cleanup(func() { datapath.SetData(oldData) })
	value := float64(math.Float32frombits(c.Coefficient))
	config := fmt.Sprintf("ChannelLifeCoeff: [%g, %g, %g, %g, %g]\n", value, value, value, value, value)
	if c.Missing {
		config = "UnusedChannelLifeControl: 0\n"
	}
	if err := os.WriteFile(filepath.Join(dir, "gamedata.yml"), []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	if err := s.Balance.Read(); err != nil {
		t.Fatal(err)
	}
	wantCoefficient := value
	if c.Missing {
		wantCoefficient = 0
	}
	if got := s.Balance.FloatInd("ChannelLifeCoeff", int(c.Level)-1); got != wantCoefficient {
		t.Fatalf("production balance coefficient=%g want=%g", got, wantCoefficient)
	}
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
	*r = server.DurSpell{ID: 0xabcd, Spell: uint32(spell.SPELL_CHANNEL_LIFE), Level: c.Level, Obj12: target, Caster16: caster, Flag20: c.Mode, Obj24: caster, Target48: target, Pos: types.Ptf(300, 300), Pos2: types.Ptf(310, 310), Frame60: 140, Frame64: 141, Frame68: 0x87654321, Field72: int32(c.Fraction), Field76: 0x55aa, Flags88: 0x77556601}
	if c.TargetNil {
		r.Target48 = nil
	}
	if c.CasterNil {
		r.Caster16 = nil
	}
	if c.Moved {
		r.Pos.X = 320
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(r.C()) <= 0xffffffff || uintptr(caster.CObj()) <= 0xffffffff || uintptr(target.CObj()) <= 0xffffffff || uintptr(unsafe.Pointer(tu)) <= 0xffffffff) {
		t.Fatal("native C-owned ChannelLife links must exceed4GiB")
	}
	raw := func(p *guarded) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p))) }
	expectedCasterHP, expectedTargetHP, expectedMana := ch.Cur, th.Cur, tu.ManaCur
	for tick, step := range c.Steps {
		expected := *f
		expected.Record.Field72 = int32(step.Fraction)
		oldMana := expectedMana
		added := false
		for _, event := range step.Services {
			switch event.Service {
			case "add_mana":
				if c.TargetClass&4 != 0 {
					added = true
					expectedMana = uint16(int32(expectedMana) + int32(event.Amount16))
					if expectedMana > c.MaxMana {
						expectedMana = c.MaxMana
					}
				}
			case "clear_damage":
				if c.Mode != 0 {
					expectedTargetHP = uint16(int32(expectedTargetHP) - event.Amount)
				} else {
					expectedCasterHP = uint16(int32(expectedCasterHP) - event.Amount)
				}
			}
		}
		got := ccall.CallIntPtr(legacy.Get_sub_52F460(), r.C())
		if got != step.Result || ch.Cur != expectedCasterHP || th.Cur != expectedTargetHP || tu.ManaCur != expectedMana {
			t.Fatalf("tick%d Creturn/casterHP/targetHP/mana=%d/%d/%d/%d original+effects=%d/%d/%d/%d", tick, got, ch.Cur, th.Cur, tu.ManaCur, step.Result, expectedCasterHP, expectedTargetHP, expectedMana)
		}
		if !bytes.Equal(raw(f), raw(&expected)) {
			t.Fatalf("tick%d changed duration/canary outside original fraction", tick)
		}
		if added && tu.ManaPrev != oldMana {
			t.Fatal("actual AddMana previous word")
		}
	}
	if ch.Max != 50000 || th.Max != 50000 || caster.HealthData != ch || target.HealthData != th || caster.UpdateData != unsafe.Pointer(cu) || target.UpdateData != unsafe.Pointer(tu) || cu.Player != ci || tu.Player != ti {
		t.Fatal("native actor/health/player links changed")
	}
	t.Logf("original case=%s public C record=%p caster=%p target=%p ticks=%d casterHP=%d targetHP=%d mana=%d fraction=%#x; production balance/mana/HP/protection and duration canaries verified", mode, r, caster, target, len(c.Steps), ch.Cur, th.Cur, tu.ManaCur, uint32(r.Field72))
}
