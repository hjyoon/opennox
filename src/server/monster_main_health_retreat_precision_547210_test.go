package server

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// GAME.EXE 005477E8..00547807 zero-extends two health WORDs, FILD/FIDIVs
// at precision 53, and FSTPs to binary32 under gameplay's ToZero mode.
// This independent reference uses neither production arithmetic helpers nor
// nearest-even/Nextafter correction, and does not touch the thread FPU mode.
func monsterHealthRetreatReference547807(cur, maximum uint16) float32 {
	if maximum == 0 {
		panic("the original health gate excludes zero maximum")
	}
	input := func(value uint16) *big.Float {
		return new(big.Float).SetPrec(53).SetMode(big.ToZero).SetUint64(uint64(value))
	}
	ratio := new(big.Float).SetPrec(53).SetMode(big.ToZero).Quo(input(cur), input(maximum))
	spill := new(big.Float).SetPrec(24).SetMode(big.ToZero).Set(ratio)
	result, accuracy := spill.Float32()
	if accuracy != big.Exact {
		panic("health-ratio reference spill must be exactly binary32")
	}
	return result
}

func monsterHealthRetreatNativeFixture547807(t *testing.T) (*Server, *Object, *MonsterUpdateData) {
	t.Helper()
	s := New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	if !s.Objs.Init(1) {
		t.Fatal("cannot initialize native object allocator")
	}
	t.Cleanup(s.Objs.FreeObjects)
	oldFlags, oldEngine := noxflags.GetGame(), noxflags.GetEngine()
	noxflags.ResetGame()
	noxflags.UnsetEngine(noxflags.EngineShowAI)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
		if oldEngine.Has(noxflags.EngineShowAI) {
			noxflags.SetEngine(noxflags.EngineShowAI)
		}
	})
	s.SetTickRate(30)
	s.SetFrame(101) // neither the staggered IDLE nor periodic food tick
	s.Rand.Logic, s.Rand.Other = prand.New(1496), prand.New(3905)
	unit := s.Objs.NewObject(&ObjectType{})
	unit.ObjClass, unit.ObjFlags, unit.SpeedBase = object.ClassMonster, object.FlagActive|object.FlagEnabled, 2
	unit.PosVec, unit.NewPos = types.Ptf(300, 300), types.Ptf(300, 300)
	update, freeUpdate := alloc.New(MonsterUpdateData{})
	t.Cleanup(freeUpdate)
	health, freeHealth := alloc.New(HealthData{})
	t.Cleanup(freeHealth)
	sounds, freeSounds := alloc.New([14]uint32{})
	t.Cleanup(freeSounds)
	*sounds = [14]uint32{13: 773}
	unit.UpdateData, unit.HealthData = unsafe.Pointer(update), health
	*update = MonsterUpdateData{AIStackInd: 0, Field363: 102, SoundSet122: unsafe.Pointer(sounds)}
	update.AIStack[0].Action = uint32(ai.ACTION_FIGHT)
	for name, p := range map[string]unsafe.Pointer{
		"unit": unsafe.Pointer(unit), "update": unsafe.Pointer(update),
		"health": unsafe.Pointer(health), "sounds": unsafe.Pointer(sounds),
	} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
			t.Fatalf("native %s below 4 GiB: %p", name, p)
		}
	}
	return s, unit, update
}

func monsterHealthRetreatAssertNative547807(t *testing.T, s *Server, unit *Object, update *MonsterUpdateData, cur, maximum uint16, threshold float32, fullEntry bool) {
	t.Helper()
	*unit.HealthData = HealthData{Cur: cur, Field2: 0x1234, Max: maximum, field6: 0x5678}
	*update = MonsterUpdateData{AIStackInd: 0, Field363: 102, RetreatLevel: threshold, SoundSet122: update.SoundSet122}
	update.AIStack[0].Action = uint32(ai.ACTION_FIGHT)
	s.AI.StackChanged = false
	want := monsterHealthRetreatReference547807(cur, maximum) <= threshold || math.IsNaN(float64(threshold))
	unitBefore := monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))
	updateBefore := monsterMoveSelectionBytes50D3B0(unsafe.Pointer(update), unsafe.Sizeof(*update))
	healthBefore := *unit.HealthData
	base := update.AIStack[0]
	var callbacks []string
	services := MonsterMainRuntime547210{
		AudioEvent: func(id uint32, got *Object) {
			if id != 773 || got != unit || update.AIStackHead().Type() != ai.ACTION_RETREAT {
				t.Fatalf("original retreat sound did not follow native action scheduling: id=%d unit=%p want=%p head=%v", id, got, unit, update.AIStackHead().Type())
			}
			callbacks = append(callbacks, "audio")
		},
		ScriptCallback: func(block *ScriptCallback, caller, trigger *Object, event ScriptEventType) {
			if block != &update.ScriptRetreat || caller != nil || trigger != unit || event != NoxEventMonsterMoveXXX {
				t.Fatal("native retreat script lost cached record, trigger or event")
			}
			callbacks = append(callbacks, "script")
		},
	}
	handled := false
	if fullEntry {
		handled = s.MonsterMainNativeRuntime547210(unit, services)
		if !handled {
			t.Fatal("complete native MainAI entry was not handled")
		}
	} else {
		handled = s.monsterMainHealthRetreat547210(unit, update, update.SoundSet122, services)
		if handled != want {
			t.Errorf("health retreat=%t want=%t: health=%d/%d threshold=%08x original-spill=%08x nearest-spill=%08x",
				handled, want, cur, maximum, math.Float32bits(threshold), math.Float32bits(monsterHealthRetreatReference547807(cur, maximum)), math.Float32bits(float32(float64(cur)/float64(maximum))))
		}
	}
	if !bytes.Equal(unitBefore, monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) ||
		*unit.HealthData != healthBefore || s.Rand.Logic.Index() != 1496 || s.Rand.Other.Index() != 3905 {
		t.Fatal("health-ratio comparison changed the object, health or either RNG")
	}
	if want {
		if update.AIStackInd != 2 || update.AIStack[0] != base || update.AIStack[1].Type() != ai.DEPENDENCY_NOT_CORNERED ||
			update.AIStack[2].Type() != ai.ACTION_RETREAT || update.AIStack[1].Args != [4]uintptr{} || update.AIStack[2].Args != [4]uintptr{} ||
			!s.AI.StackChanged || fmt.Sprint(callbacks) != "[audio script]" {
			t.Fatalf("original health-ratio retreat absent: health=%d/%d threshold=%08x stack=%+v callbacks=%v changed=%t",
				cur, maximum, math.Float32bits(threshold), update.GetAIStack(), callbacks, s.AI.StackChanged)
		}
	} else if !bytes.Equal(updateBefore, monsterMoveSelectionBytes50D3B0(unsafe.Pointer(update), unsafe.Sizeof(*update))) || s.AI.StackChanged || len(callbacks) != 0 {
		t.Fatal("ordered-above-threshold branch changed update, stack or callbacks")
	}
}

func TestMonsterMainHealthRetreat547210NativeRatioBoundaries(t *testing.T) {
	for _, health := range [][2]uint16{{0, 65535}, {1, 2}, {1, 3}, {2, 3}, {1, 10}, {1, 25}, {32767, 65535}, {65534, 65535}, {65535, 65535}, {65535, 1}} {
		for _, boundary := range []string{"below", "equal", "above"} {
			for _, npc := range []bool{false, true} {
				for _, fullEntry := range []bool{false, true} {
					t.Run(fmt.Sprintf("%d-of-%d/%s/npc-%t/main-%t", health[0], health[1], boundary, npc, fullEntry), func(t *testing.T) {
						s, unit, update := monsterHealthRetreatNativeFixture547807(t)
						if npc {
							unit.ObjSubClass = object.SubClass(object.MonsterNPC)
						}
						threshold := monsterHealthRetreatReference547807(health[0], health[1])
						switch boundary {
						case "below":
							threshold = math.Nextafter32(threshold, float32(math.Inf(-1)))
						case "above":
							threshold = math.Nextafter32(threshold, float32(math.Inf(1)))
						}
						monsterHealthRetreatAssertNative547807(t, s, unit, update, health[0], health[1], threshold, fullEntry)
					})
				}
			}
		}
	}
}

func TestMonsterMainHealthRetreat547210IndependentRatioCalibration(t *testing.T) {
	for _, tc := range []struct {
		cur, max uint16
		bits     uint32
	}{
		{0, 65535, 0}, {1, 2, 0x3f000000}, {1, 3, 0x3eaaaaaa}, {2, 3, 0x3f2aaaaa},
		{1, 10, 0x3dcccccc}, {1, 25, 0x3d23d70a}, {65535, 65535, 0x3f800000}, {65535, 1, 0x477fff00},
	} {
		if got := math.Float32bits(monsterHealthRetreatReference547807(tc.cur, tc.max)); got != tc.bits {
			t.Fatalf("independent reference %d/%d=%08x want=%08x", tc.cur, tc.max, got, tc.bits)
		}
	}
	if monsterHealthRetreatReference547807(1, 3) == float32(float64(1)/float64(3)) {
		t.Fatal("original ToZero and host nearest-even calibration must differ")
	}
}

func TestMonsterMainHealthRetreat547210IndependentNativeRatioMatrix(t *testing.T) {
	s, unit, update := monsterHealthRetreatNativeFixture547807(t)
	data := rand.New(rand.NewSource(0x547807)) // fixture data only, not game RNG
	for fixture := 0; fixture < 4096; fixture++ {
		cur, maximum := uint16(data.Intn(65536)), uint16(data.Intn(65535)+1)
		threshold := monsterHealthRetreatReference547807(cur, maximum)
		switch fixture % 3 {
		case 0:
			threshold = math.Nextafter32(threshold, float32(math.Inf(-1)))
		case 2:
			threshold = math.Nextafter32(threshold, float32(math.Inf(1)))
		}
		monsterHealthRetreatAssertNative547807(t, s, unit, update, cur, maximum, threshold, fixture&1 != 0)
	}
}
