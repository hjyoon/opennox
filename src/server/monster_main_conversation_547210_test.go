package server

import (
	"image"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

func monsterConversationFixture547210(t *testing.T) (*Server, *Object, MonsterMainRuntime547210, *int) {
	t.Helper()
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameModeCoop)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})
	s := unitFollowTestServer5158C0(t)
	s.SetTickRate(30)
	player := &Player{CursorVec: image.Pt(300, 300)}
	host := &Object{
		ObjClass:   object.ClassPlayer,
		UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: player}),
	}
	s.Players.SetHost(player, host)
	unit := passiveMonsterTestObject547210(t)
	unit.serverHandle = s.handle
	unit.Field5 = 0x10
	unit.PosVec = types.Ptf(300, 300)
	update := unit.UpdateDataMonster()
	update.AIStack[0].Action = uint32(ai.ACTION_GUARD)
	sounds := &[2]uint32{0, 1234}
	update.SoundSet122 = unsafe.Pointer(&sounds[0])
	count := new(int)
	runtime := MonsterMainRuntime547210{
		GUICursorActive:    func() bool { return false },
		FindObjectAtCursor: func(*Object) *Object { return unit },
		AudioEvent: func(id uint32, source *Object) {
			if id != sounds[1] || source != unit {
				t.Fatalf("hover audio = %d/%p, want %d/%p", id, source, sounds[1], unit)
			}
			*count++
		},
	}
	return s, unit, runtime, count
}

func TestMonsterMainConversation547210HeldCursorDoesNotReplaySound(t *testing.T) {
	s, unit, runtime, count := monsterConversationFixture547210(t)
	update := unit.UpdateDataMonster()
	if !s.monsterMainConversation547210(unit, update, runtime) {
		t.Fatal("first hover did not enter conversation")
	}
	before := *update
	for frame := uint32(1); frame <= 180; frame++ {
		s.SetFrame(frame)
		if s.monsterMainConversation547210(unit, update, runtime) {
			t.Errorf("held cursor re-entered conversation at frame %d", frame)
			break
		}
	}
	if *count != 1 || *update != before {
		t.Fatalf("held cursor changed conversation: audio calls %d, stack index %d (want one call and unchanged index %d)", *count, update.AIStackInd, before.AIStackInd)
	}
}

func TestMonsterMainConversation547210OriginalWaitActionGate(t *testing.T) {
	// GAME.EXE 005472D7 calls 0050A0D0 with literal 2, not 41. Check
	// both the current action and a scheduled wait below FACE_OBJECT.
	for _, tc := range []struct {
		name   string
		action uint32
		head   bool
		want   bool
	}{
		{"relative wait head", 2, true, false},
		{"scheduled relative wait", 2, false, false},
		{"time dependency head", 41, true, true},
		{"scheduled time dependency", 41, false, true},
		{"absolute wait", 1, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, unit, runtime, count := monsterConversationFixture547210(t)
			update := unit.UpdateDataMonster()
			update.AIStackInd = 1
			update.AIStack[1].Action = tc.action
			if !tc.head {
				update.AIStackInd = 2
				update.AIStack[2].Action = uint32(ai.ACTION_FACE_OBJECT)
			}
			before := *update
			if got := s.monsterMainConversation547210(unit, update, runtime); got != tc.want {
				t.Fatalf("action %d at head=%v: conversation = %v, want %v", tc.action, tc.head, got, tc.want)
			}
			if !tc.want && (*count != 0 || *update != before) {
				t.Fatal("relative wait guard changed state or played audio")
			}
			if tc.want && *count != 1 {
				t.Fatalf("accepted hover played %d sounds, want one", *count)
			}
		})
	}
}

func TestMonsterMainConversationImpossible547210OriginalWaitActionGate(t *testing.T) {
	// The eligibility predicate must agree with the literal action 2 check
	// in GAME.EXE 005472D7, even while FACE_OBJECT is the current action.
	for _, tc := range []struct {
		name   string
		action uint32
		head   bool
		want   bool
	}{
		{"relative wait head", 2, true, true},
		{"scheduled relative wait", 2, false, true},
		{"time dependency head", 41, true, false},
		{"scheduled time dependency", 41, false, false},
		{"absolute wait", 1, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, unit, _, _ := monsterConversationFixture547210(t)
			update := unit.UpdateDataMonster()
			update.AIStackInd = 1
			update.AIStack[1].Action = tc.action
			if !tc.head {
				update.AIStackInd = 2
				update.AIStack[2].Action = uint32(ai.ACTION_FACE_OBJECT)
			}
			before := *update
			if got := s.monsterMainConversationImpossible547210(unit, update); got != tc.want {
				t.Fatalf("action %d at head=%v: conversation impossible = %v, want %v", tc.action, tc.head, got, tc.want)
			}
			if *update != before {
				t.Fatal("conversation eligibility check changed the AI stack")
			}
		})
	}
}
