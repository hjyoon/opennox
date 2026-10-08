package server

import (
	"fmt"
	"image"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

func monsterMainFearFixture547210(t *testing.T) (*Server, *Object, *MonsterUpdateData) {
	t.Helper()
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(oldFlags) })
	s := unitFollowTestServer5158C0(t)
	s.SetFrame(200)
	s.SetTickRate(30)
	unit := passiveMonsterTestObject547210(t)
	unit.serverHandle = s.handle
	unit.Buffs = 1 << ENCHANT_AFRAID
	unit.PosVec = types.Ptf(8640, -3120)
	unit.PrevPos, unit.VelVec = types.Ptf(1, 2), types.Ptf(3, 4)
	unit.ForceVec, unit.Pos24 = types.Ptf(5, 6), types.Ptf(7, 8)
	unit.Direction1, unit.Direction2 = 13, 29
	update := unit.UpdateDataMonster()
	update.StatusFlags = 0
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_WAIT), Args: [4]uintptr{1, 2, 3, 4}}
	update.CurrentEnemy = &Object{PosVec: types.Ptf(-100, -200)}
	update.Field2, update.Field67, update.Field74, update.Field91 = 2, 67, 74, 91
	update.Field120_0, update.Field120_1, update.Field120_2, update.Field120_3 = 7, 1, 2, 3
	update.Field124, update.Field137 = 124, 137
	return s, unit, update
}

func TestMonsterMainFear547210Transition(t *testing.T) {
	for _, mode := range []string{
		"wait", "idle", "fight", "object cast", "location cast", "duration cast",
		"confused", "existing confused", "anti magic", "disabled", "destroyed",
		"can cast and block", "stunned and frozen", "no enemy", "NaN aggression",
	} {
		t.Run(mode, func(t *testing.T) {
			s, unit, update := monsterMainFearFixture547210(t)
			switch mode {
			case "idle":
				update.AIStack[0].Action = uint32(ai.ACTION_IDLE)
				update.Field137 = s.Frame()
			case "fight":
				update.AIStack[0].Action = uint32(ai.ACTION_FIGHT)
			case "object cast":
				update.AIStack[0].Action = uint32(ai.ACTION_CAST_SPELL_ON_OBJECT)
			case "location cast":
				update.AIStack[0].Action = uint32(ai.ACTION_CAST_SPELL_ON_LOCATION)
			case "duration cast":
				update.AIStack[0].Action = uint32(ai.ACTION_CAST_DURATION_SPELL)
			case "confused":
				unit.Buffs |= 1 << ENCHANT_CONFUSED
			case "existing confused":
				unit.Buffs |= 1 << ENCHANT_CONFUSED
				update.AIStack[0].Action = uint32(ai.ACTION_CONFUSED)
			case "anti magic":
				unit.Buffs |= 1 << ENCHANT_ANTI_MAGIC
			case "disabled":
				unit.ObjFlags &^= object.FlagEnabled
			case "destroyed":
				unit.ObjFlags |= object.FlagDestroyed
			case "can cast and block":
				update.StatusFlags = object.MonStatusCanCastSpells | object.MonStatusCanBlock
				update.Field363 = s.Frame() + 1
			case "stunned and frozen":
				unit.Buffs |= 1<<ENCHANT_HELD | 1<<ENCHANT_FREEZE
			case "no enemy":
				update.CurrentEnemy = nil
			case "NaN aggression":
				update.Aggression = float32(math.NaN())
			}
			beforeUnit, base, enemy := *unit, update.AIStack[0], update.CurrentEnemy
			sounds := [13]uint32{12: 0x87654321}
			update.SoundSet122 = unsafe.Pointer(&sounds[0])
			audioCalls := 0
			runtime := MonsterMainRuntime547210{
				AudioEvent: func(id uint32, got *Object) {
					audioCalls++
					if got != unit || id != sounds[12] || update.AIStackHead().Type() != ai.ACTION_FLEE {
						t.Fatalf("fear audio preceded scheduling or lost its source: %x/%p", id, got)
					}
				},
				RandomInt:    func(int, int) int { t.Fatal("fear consumed random numbers"); return 0 },
				SearchEdible: func(*Object, float32) *Object { t.Fatal("fear reached food search"); return nil },
			}
			if !s.MonsterMainNativeRuntime547210(unit, runtime) {
				t.Fatal("fear transition remained unported")
			}
			index := 1
			if mode == "idle" {
				index = 0
			} else if update.AIStack[0] != base {
				t.Fatal("fear changed the preceding action")
			}
			if mode == "confused" {
				if update.AIStack[index].Type() != ai.DEPENDENCY_IS_ENCHANTED || update.AIStack[index].ArgU32(0) != uint32(ENCHANT_CONFUSED) ||
					update.AIStack[index+1].Type() != ai.ACTION_CONFUSED {
					t.Fatal("fear lost the preceding confusion transition")
				}
				index += 2
			}
			dependency, flee := &update.AIStack[index], &update.AIStack[index+1]
			dependencyArgs := [4]uintptr{uintptr(ENCHANT_AFRAID)}
			if mode == "idle" {
				// 0054748B writes only Arg0 after reusing the idle slot.
				dependencyArgs = base.Args
				dependencyArgs[0] = uintptr(ENCHANT_AFRAID)
			}
			if update.AIStackInd != int8(index+1) || dependency.Type() != ai.DEPENDENCY_IS_ENCHANTED ||
				dependency.Args != dependencyArgs || dependency.Field5 != 0 ||
				flee.Type() != ai.ACTION_FLEE || flee.Args != [4]uintptr{uintptr(math.Float32bits(unit.PosVec.X)), uintptr(math.Float32bits(unit.PosVec.Y)), 0, 0} || flee.Field5 != 0 {
				t.Fatalf("fear stack = %+v", update.GetAIStack())
			}
			if *unit != beforeUnit || update.CurrentEnemy != enemy || audioCalls != 1 || !s.AI.StackChanged {
				t.Fatal("fear changed motion/enemy state, lost audio, or bypassed ordinary stack services")
			}
			if update.Field2 != 0 || update.Field67 != 0 || update.Field74 != 0 || update.Field91 != 0 || update.Field120_0 != 7 ||
				update.Field120_1 != 0 || update.Field120_2 != 0 || update.Field120_3 != 0 || update.Field124 != 200 || update.Field137 != 200 {
				t.Fatal("fear did not use the real action-reset services")
			}
		})
	}
}

func TestMonsterMainFear547210MovementBoundary(t *testing.T) {
	for _, bits := range []uint32{0, 0x80000000, 0x3c23d709, 0x3c23d70a, 0x3c23d70b, 0x7f800000, 0xff800000, 0x7fc12345, 0x7f812345} {
		t.Run(fmt.Sprintf("speed:%08x", bits), func(t *testing.T) {
			s, unit, update := monsterMainFearFixture547210(t)
			unit.SpeedBase = math.Float32frombits(bits)
			before := *update
			handled := s.MonsterMainNative547210(unit)
			eligible := bits == 0x3c23d70a || bits == 0x3c23d70b || bits == 0x7f800000
			if eligible {
				if !handled || update.AIStackInd != 2 || update.AIStackHead().Type() != ai.ACTION_FLEE || !s.AI.StackChanged {
					t.Fatal("ordered moving speed did not schedule fear")
				}
			} else if *update != before || s.AI.StackChanged {
				t.Fatal("immobile or unordered speed changed the stack")
			}
		})
	}
}

func TestMonsterMainFear547210RawCoordinatesAndOptionalAudio(t *testing.T) {
	for _, bits := range [][2]uint32{{0x80000000, 0}, {0x7fc12345, 0xff800000}, {0x7f812345, 0x7f800000}} {
		for _, mode := range []string{"nil sound", "zero sound", "no audio hook"} {
			t.Run(fmt.Sprintf("%08x:%08x:%s", bits[0], bits[1], mode), func(t *testing.T) {
				s, unit, update := monsterMainFearFixture547210(t)
				unit.PosVec = types.Ptf(math.Float32frombits(bits[0]), math.Float32frombits(bits[1]))
				sounds := [13]uint32{}
				if mode != "nil sound" {
					update.SoundSet122 = unsafe.Pointer(&sounds[0])
				}
				calls := 0
				runtime := MonsterMainRuntime547210{AudioEvent: func(id uint32, got *Object) {
					calls++
					if id != 0 || got != unit {
						t.Fatal("zero sound call lost its original arguments")
					}
				}}
				if mode == "no audio hook" {
					runtime.AudioEvent = nil
				}
				if !s.MonsterMainNativeRuntime547210(unit, runtime) || update.AIStackHead().Args != [4]uintptr{uintptr(bits[0]), uintptr(bits[1]), 0, 0} {
					t.Fatalf("fear did arithmetic on raw coordinates: %+v", update.AIStackHead())
				}
				wantCalls := 0
				if mode == "zero sound" {
					wantCalls = 1
				}
				if calls != wantCalls {
					t.Fatalf("audio calls=%d want %d", calls, wantCalls)
				}
			})
		}
	}
}

func TestMonsterMainFear547210DoesNotDuplicateExistingFlee(t *testing.T) {
	for _, index := range []int{0, 1, 5} {
		t.Run(fmt.Sprintf("flee index:%d", index), func(t *testing.T) {
			s, unit, update := monsterMainFearFixture547210(t)
			update.AIStackInd = 5
			update.AIStack[5].Action = uint32(ai.ACTION_WAIT)
			update.AIStack[index].Action = uint32(ai.ACTION_FLEE)
			before := *update
			if index == 5 {
				// An existing moving FLEE skips fear scheduling, but the common
				// tail still records this tick's displacement in the original.
				before.Field124, before.Field125, before.Field126 = s.Frame(), math.Float32bits(unit.PosVec.X), math.Float32bits(unit.PosVec.Y)
			}
			calls := 0
			s.MonsterMainNativeRuntime547210(unit, MonsterMainRuntime547210{AudioEvent: func(uint32, *Object) { calls++ }})
			if *update != before || s.AI.StackChanged || calls != 0 {
				t.Fatal("existing FLEE anywhere in the stack was duplicated")
			}
		})
	}
}

func TestMonsterMainFear547210RejectedAndPartialPushStillReturnsAndSounds(t *testing.T) {
	for _, mode := range []string{"full", "one slot", "dead head"} {
		t.Run(mode, func(t *testing.T) {
			s, unit, update := monsterMainFearFixture547210(t)
			switch mode {
			case "full":
				update.AIStackInd = int8(len(update.AIStack) - 1)
				update.AIStackHead().Action = uint32(ai.ACTION_WAIT)
			case "one slot":
				update.AIStackInd = int8(len(update.AIStack) - 2)
				update.AIStackHead().Action = uint32(ai.ACTION_WAIT)
			case "dead head":
				update.AIStack[0].Action = uint32(ai.ACTION_DEAD)
			}
			before := *update
			sounds := [13]uint32{12: 1200}
			update.SoundSet122 = unsafe.Pointer(&sounds[0])
			before.SoundSet122 = update.SoundSet122
			calls := 0
			if !s.MonsterMainNativeRuntime547210(unit, MonsterMainRuntime547210{AudioEvent: func(id uint32, got *Object) {
				calls++
				if id != 1200 || got != unit {
					t.Fatal("rejected push changed fear audio arguments")
				}
			}}) || calls != 1 {
				t.Fatal("rejected action push did not retain the fear early return and sound")
			}
			if mode == "one slot" {
				if update.AIStackInd != int8(len(update.AIStack)-1) || update.AIStackHead().Type() != ai.DEPENDENCY_IS_ENCHANTED || update.AIStackHead().ArgU32(0) != 11 || !s.AI.StackChanged {
					t.Fatal("partial fear push did not retain its dependency")
				}
			} else if *update != before || s.AI.StackChanged {
				t.Fatal("rejected fear push changed state")
			}
		})
	}
}

func TestMonsterMainFear547210PrefixGates(t *testing.T) {
	for _, mode := range []string{"idle throttle", "guard throttle", "dead", "uninterruptible", "no fear", "invalid stack", "negative stack", "nonmonster", "no update", "nil"} {
		t.Run(mode, func(t *testing.T) {
			s, unit, update := monsterMainFearFixture547210(t)
			switch mode {
			case "idle throttle":
				update.AIStack[0].Action = uint32(ai.ACTION_IDLE)
				update.Field137 = 199
			case "guard throttle":
				update.AIStack[0].Action = uint32(ai.ACTION_GUARD)
				update.Field137 = 199
			case "dead":
				unit.ObjFlags |= object.FlagDead
			case "uninterruptible":
				update.AIStackInd = 1
				update.AIStack[0].Action = uint32(ai.DEPENDENCY_UNINTERRUPTABLE)
				update.AIStack[1].Action = uint32(ai.ACTION_WAIT)
			case "no fear":
				unit.Buffs = 1 << ENCHANT_HELD
			case "invalid stack":
				update.AIStackInd = int8(len(update.AIStack))
			case "negative stack":
				update.AIStackInd = -1
			case "nonmonster":
				unit.ObjClass = object.ClassPlayer
			case "no update":
				unit.UpdateData = nil
			}
			before := *update
			sounds := [13]uint32{12: 1200}
			update.SoundSet122 = unsafe.Pointer(&sounds[0])
			before.SoundSet122 = update.SoundSet122
			arg := unit
			if mode == "nil" {
				arg = nil
			}
			calls := 0
			s.MonsterMainNativeRuntime547210(arg, MonsterMainRuntime547210{AudioEvent: func(uint32, *Object) { calls++ }})
			if *update != before || s.AI.StackChanged || calls != 0 {
				t.Fatal("fear bypassed an original prefix gate")
			}
		})
	}
}

type monsterMainFearCancel547210 struct{ cancel func(*Object) }

func (monsterMainFearCancel547210) Type() ai.ActionType   { return ai.ACTION_WAIT }
func (monsterMainFearCancel547210) Start(*Object)         {}
func (monsterMainFearCancel547210) Update(*Object)        {}
func (monsterMainFearCancel547210) End(*Object)           {}
func (a monsterMainFearCancel547210) Cancel(unit *Object) { a.cancel(unit) }

func TestMonsterMainFear547210CachesSoundPointerButReadsLiveEntryAndPostPushPosition(t *testing.T) {
	s, unit, update := monsterMainFearFixture547210(t)
	noxflags.SetGame(noxflags.GameModeCoop)
	update.AIStack[0].Field5 = 1
	originalSound, replacementSound := [13]uint32{12: 100}, [13]uint32{12: 200}
	update.SoundSet122 = unsafe.Pointer(&originalSound[0])
	previous, exists := aiActions[ai.ACTION_WAIT]
	t.Cleanup(func() {
		if exists {
			aiActions[ai.ACTION_WAIT] = previous
		} else {
			delete(aiActions, ai.ACTION_WAIT)
		}
	})
	cancelCalls := 0
	aiActions[ai.ACTION_WAIT] = monsterMainFearCancel547210{cancel: func(got *Object) {
		cancelCalls++
		if got != unit {
			t.Fatal("cancel lost the native unit")
		}
		unit.PosVec = types.Ptf(-17.25, math.Float32frombits(0x80000000))
		originalSound[12] = 300
	}}
	cursorCalls, audioCalls := 0, 0
	if !s.MonsterMainNativeRuntime547210(unit, MonsterMainRuntime547210{
		GUICursorActive: func() bool { cursorCalls++; update.SoundSet122 = unsafe.Pointer(&replacementSound[0]); return true },
		AudioEvent: func(id uint32, got *Object) {
			audioCalls++
			if id != 300 || got != unit || cancelCalls != 1 || update.AIStackHead().Args != [4]uintptr{uintptr(math.Float32bits(-17.25)), 0x80000000, 0, 0} {
				t.Fatalf("cached pointer/live entry/post-push position lost: id=%d stack=%+v", id, update.AIStackHead())
			}
		},
	}) || cursorCalls != 1 || cancelCalls != 1 || audioCalls != 1 {
		t.Fatal("fear changed original callback ordering")
	}
}

func TestMonsterMainFear547210ReadsLiveActionStackAfterCursorCallback(t *testing.T) {
	for _, replacementFlee := range []bool{false, true} {
		t.Run(fmt.Sprintf("replacement flee:%t", replacementFlee), func(t *testing.T) {
			s, unit, update := monsterMainFearFixture547210(t)
			noxflags.SetGame(noxflags.GameModeCoop)
			update.AIStack[0].Action = uint32(ai.ACTION_FLEE)
			replacement := &MonsterUpdateData{AIStackInd: 0}
			replacement.AIStack[0].Action = uint32(ai.ACTION_WAIT)
			if replacementFlee {
				replacement.AIStack[0].Action = uint32(ai.ACTION_FLEE)
				update.AIStack[0].Action = uint32(ai.ACTION_WAIT)
			}
			before, beforeReplacement := *update, *replacement
			s.MonsterMainNativeRuntime547210(unit, MonsterMainRuntime547210{GUICursorActive: func() bool { unit.UpdateData = unsafe.Pointer(replacement); return true }})
			if *update != before {
				t.Fatal("fear pushed into the stale update data")
			}
			if replacementFlee {
				if *replacement != beforeReplacement || s.AI.StackChanged {
					t.Fatal("fear ignored a live FLEE action")
				}
			} else if replacement.AIStackInd != 2 || replacement.AIStackHead().Type() != ai.ACTION_FLEE || replacement.AIStack[1].ArgU32(0) != 11 || !s.AI.StackChanged {
				t.Fatal("fear consulted the stale FLEE stack")
			}
		})
	}
}

func TestMonsterMainFear547210InversionPreemptsFear(t *testing.T) {
	s, unit, update := monsterInversionNativeFixture5408D0(t)
	unit.Buffs = 1 << ENCHANT_AFRAID
	unit.SpeedBase = 1.95
	sounds := [13]uint32{12: 1200}
	update.SoundSet122 = unsafe.Pointer(&sounds[0])
	calls := 0
	if !s.MonsterMainNativeRuntime547210(unit, MonsterMainRuntime547210{AudioEvent: func(uint32, *Object) { calls++ }}) ||
		update.AIStackHead().Type() != ai.ACTION_CAST_SPELL_ON_OBJECT || update.HasAction(ai.ACTION_FLEE) || calls != 0 {
		t.Fatal("fear preceded an eligible inversion cast")
	}
}

func TestMonsterMainFear547210ConversationPreemptsFear(t *testing.T) {
	s, unit, update := monsterMainFearFixture547210(t)
	noxflags.SetGame(noxflags.GameModeCoop)
	unit.Field5 = 0x10
	unit.PosVec = types.Ptf(300, 300)
	player := &Player{CursorVec: image.Pt(300, 300)}
	host := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: player})}
	s.Players.SetHost(player, host)
	if !s.MonsterMainNativeRuntime547210(unit, MonsterMainRuntime547210{
		GUICursorActive: func() bool { return false }, FindObjectAtCursor: func(*Object) *Object { return unit },
	}) || update.AIStackHead().Type() != ai.ACTION_FACE_OBJECT || update.HasAction(ai.ACTION_FLEE) || update.HasAction(ai.ACTION_CONFUSED) {
		t.Fatal("fear preceded the original conversation transition")
	}
}
