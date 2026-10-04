package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/balance"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

func init() {
	// Standalone server tests have no C data blobs. This zero-filled test
	// storage holds the original inversion scan DWORD, not oracle bytes.
	const size = 2489160
	if blob := memmap.BlobByAddr(0x5D4594); blob == nil {
		memmap.RegisterBlobData(0x5D4594, "inversion_test", make([]byte, size))
	} else if blob.Size < size {
		data := make([]byte, size)
		copy(data, blob.Data)
		memmap.RegisterBlobData(0x5D4594, "inversion_test", data)
	}
}

func monsterInversionNativeFixture5408D0(t *testing.T) (*Server, *Object, *MonsterUpdateData) {
	t.Helper()
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(oldFlags) })
	s, unit, update := monsterScriptHitFixture515A30(t)
	s.SetFrame(200)
	s.SetTickRate(30)
	s.Rand.Logic = prand.New(5408)
	s.Map.Init()
	s.Balance.file = &balance.File{Global: balance.Config{"inversionrange": balance.Float(100)}}
	s.Spells.byID = map[spell.ID]*SpellDef{
		spell.SPELL_INVERSION: {ID: spell.SPELL_INVERSION, Def: things.Spell{Flags: things.SpellMobsCanCast}},
	}
	unit.ObjFlags = object.FlagEnabled
	unit.PosVec = types.Ptf(100, 100)
	unit.HealthData = &HealthData{Cur: 100, Field2: 100, Max: 100}
	update.StatusFlags = object.MonStatusCanCastSpells
	update.Field362_0, update.Field362_2 = 7, 7
	update.Field363 = 200
	update.Aggression = 0
	monsterFightSetSpellFlag540B90(update, spell.SPELL_INVERSION, 0x08000000)
	missile := &Object{
		ObjClass: object.ClassMissile, ObjSubClass: object.SubClass(object.MissileMagic),
		ObjFlags: object.FlagActive, PosVec: types.Ptf(130, 140), NewPos: types.Ptf(130, 140),
		UpdateData: unsafe.Pointer(&MissileUpdateData{Target: unit}),
	}
	s.Map.AddObjectToIndex(missile)
	flag := memmap.PtrUint32(0x5D4594, 2489156)
	oldFlag := *flag
	*flag = 0xfeedbeef
	t.Cleanup(func() { *flag = oldFlag; runtime.KeepAlive(missile) })
	return s, unit, update
}

func TestMonsterMainInversion5408D0SchedulesNativeSelfCast(t *testing.T) {
	for _, tc := range []struct {
		name     string
		buff     uint32
		duration bool
	}{
		{name: "ordinary"},
		{name: "afraid", buff: 1 << ENCHANT_AFRAID},
		{name: "confused", buff: 1 << ENCHANT_CONFUSED},
		{name: "duration", duration: true},
		{name: "confused duration", buff: 1 << ENCHANT_CONFUSED, duration: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, unit, update := monsterInversionNativeFixture5408D0(t)
			unit.Buffs = tc.buff
			if tc.duration {
				s.Spells.byID[spell.SPELL_INVERSION].Def.Flags |= things.SpellDuration
			}
			base := update.AIStack[0]
			if !s.MonsterMainNativeRuntime547210(unit, MonsterMainRuntime547210{}) {
				t.Fatal("eligible inversion remained on the unported main-AI path")
			}
			i := 1
			if tc.buff&(1<<ENCHANT_CONFUSED) != 0 {
				if update.AIStack[i].Type() != ai.DEPENDENCY_IS_ENCHANTED || update.AIStack[i].ArgU32(0) != 3 ||
					update.AIStack[i+1].Type() != ai.ACTION_CONFUSED {
					t.Fatal("inversion did not retain the preceding confused transition")
				}
				i += 2
			}
			if update.AIStack[0] != base || update.AIStack[i].Type() != ai.DEPENDENCY_UNINTERRUPTABLE {
				t.Fatal("self-cast changed the previous action or omitted its dependency")
			}
			i++
			wantAction := ai.ACTION_CAST_SPELL_ON_OBJECT
			if tc.duration {
				wantAction = ai.ACTION_CAST_DURATION_SPELL
				if update.AIStack[i].Type() != ai.DEPENDENCY_TIME || update.AIStack[i].ArgU32(0) < 215 || update.AIStack[i].ArgU32(0) > 260 {
					t.Fatal("duration self-cast has no ordinary time dependency")
				}
				i++
			}
			head := update.AIStackHead()
			if int(update.AIStackInd) != i || head.Type() != wantAction ||
				head.ArgU32(0) != uint32(spell.SPELL_INVERSION) || head.ArgObj(2) != unit || head.Args[3] != 0 {
				t.Fatalf("inversion stack = index %d head %+v", update.AIStackInd, head)
			}
			if update.Field363 != 207 || memmap.Uint32(0x5D4594, 2489156) != 1 || !s.AI.StackChanged {
				t.Fatalf("cooldown/scan/stack changed = %d/%d/%t", update.Field363, memmap.Uint32(0x5D4594, 2489156), s.AI.StackChanged)
			}
			if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 || head.Args[2] <= math.MaxUint32) {
				t.Fatal("self-cast did not retain a native object pointer above 4 GiB")
			}
			runtime.KeepAlive(unit)
		})
	}
}

func TestMonsterMainInversion5408D0PreservesCallerAndSelectorGates(t *testing.T) {
	for _, mode := range []string{"disabled", "cannot cast", "future deadline", "anti magic", "cast head", "idle throttle", "uninterruptible", "dead"} {
		t.Run(mode, func(t *testing.T) {
			s, unit, update := monsterInversionNativeFixture5408D0(t)
			switch mode {
			case "disabled":
				unit.ObjFlags = 0
			case "cannot cast":
				update.StatusFlags = 0
			case "future deadline":
				update.Field363 = 201
			case "anti magic":
				unit.Buffs = 1 << ENCHANT_ANTI_MAGIC
			case "cast head":
				update.AIStack[0].Action = uint32(ai.ACTION_CAST_SPELL_ON_OBJECT)
			case "idle throttle":
				update.AIStack[0].Action = uint32(ai.ACTION_IDLE)
				unit.NetCode = 1
				update.Field137 = 0
			case "uninterruptible":
				update.AIStackInd = 1
				update.AIStack[0].Action = uint32(ai.DEPENDENCY_UNINTERRUPTABLE)
				update.AIStack[1].Action = uint32(ai.ACTION_WAIT)
			case "dead":
				unit.ObjFlags |= object.FlagDead
			}
			before := *update
			s.MonsterMainNativeRuntime547210(unit, MonsterMainRuntime547210{})
			if update.AIStackInd != before.AIStackInd || update.AIStack != before.AIStack || update.Field363 != before.Field363 || memmap.Uint32(0x5D4594, 2489156) != 0xfeedbeef || s.AI.StackChanged {
				t.Fatalf("main caller/selector gate did not remain inert: index=%d deadline=%d scan=%x", update.AIStackInd, update.Field363, memmap.Uint32(0x5D4594, 2489156))
			}
		})
	}
}

func TestMonsterMainInversion5408D0ConfusionWithoutThreatAndFullStack(t *testing.T) {
	for _, mode := range []string{"anti magic", "no threat", "already confused", "full stack"} {
		t.Run(mode, func(t *testing.T) {
			s, unit, update := monsterInversionNativeFixture5408D0(t)
			unit.Buffs = 1 << ENCHANT_CONFUSED
			switch mode {
			case "anti magic":
				unit.Buffs |= 1 << ENCHANT_ANTI_MAGIC
			case "no threat", "already confused":
				s.Map.Init()
			case "full stack":
				update.AIStackInd = int8(len(update.AIStack) - 1)
				for i := range update.AIStack {
					update.AIStack[i].Action = uint32(ai.ACTION_WAIT)
				}
			}
			if mode == "already confused" {
				update.AIStack[0].Action = uint32(ai.ACTION_CONFUSED)
			}
			before := *update
			s.MonsterMainNativeRuntime547210(unit, MonsterMainRuntime547210{})
			if mode == "full stack" {
				if update.AIStackInd != before.AIStackInd || update.AIStack != before.AIStack || update.Field363 != 207 || s.AI.StackChanged {
					t.Fatal("full stack changed actions or lost original successful-selector cooldown")
				}
				return
			}
			wantIndex := int8(2)
			if mode == "already confused" {
				wantIndex = 0
			}
			if update.AIStackInd != wantIndex || update.AIStackHead().Type() != ai.ACTION_CONFUSED || update.Field363 != 200 ||
				mode != "already confused" && (update.AIStack[0] != before.AIStack[0] || update.AIStack[1].Type() != ai.DEPENDENCY_IS_ENCHANTED || update.AIStack[1].ArgU32(0) != 3) {
				t.Fatal("confusion was skipped, duplicated or returned before original prefix")
			}
			wantFlag := uint32(0)
			if mode == "anti magic" {
				wantFlag = 0xfeedbeef
			}
			if memmap.Uint32(0x5D4594, 2489156) != wantFlag {
				t.Fatal("caller anti-magic gate incorrectly ran missile scan")
			}
		})
	}
}
