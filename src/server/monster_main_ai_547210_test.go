package server

import (
	"image"
	"math"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

func passiveMonsterTestObject547210(t *testing.T) *Object {
	t.Helper()
	unit := monsterActionTestObject50A910(t)
	unit.ObjSubClass = object.SubClass(object.MonsterFemaleNPC | object.MonsterNoTarget | object.MonsterHasSoul)
	unit.ObjFlags = object.FlagActive | object.FlagEnabled
	unit.HealthData = &HealthData{Field2: 75}
	unit.SpeedBase = 1.95
	update := unit.UpdateDataMonster()
	update.AIStackInd = 0
	update.AIStack[0].Action = uint32(ai.ACTION_IDLE)
	update.Aggression = 0
	update.RetreatLevel = 0.5
	update.StatusFlags = object.MonStatusCanSeeFriends | object.MonStatusRunning | object.MonStatusAlwaysRun
	return unit
}

func TestMonsterMainNative547210JenniferPassiveState(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameHost | noxflags.GameModeCoop)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	unit := passiveMonsterTestObject547210(t)
	beforeUnit := *unit
	beforeUpdate := *unit.UpdateDataMonster()
	if !new(Server).MonsterMainNative547210(unit) {
		t.Fatal("Jennifer passive state was not handled")
	}
	if *unit != beforeUnit || *unit.UpdateDataMonster() != beforeUpdate {
		t.Fatal("Jennifer passive state was changed")
	}
}

func TestMonsterMainNative547210ExactEarlyReturns(t *testing.T) {
	t.Run("staggered idle throttle", func(t *testing.T) {
		unit := passiveMonsterTestObject547210(t)
		unit.Buffs = ^uint32(0)
		server := new(Server)
		server.SetFrame(1)
		if !server.MonsterMainNative547210(unit) {
			t.Fatal("staggered IDLE tick was not handled")
		}
	})
	t.Run("uninterruptible dependency", func(t *testing.T) {
		unit := passiveMonsterTestObject547210(t)
		update := unit.UpdateDataMonster()
		update.AIStackInd = 1
		update.AIStack[0].Action = uint32(ai.DEPENDENCY_UNINTERRUPTABLE)
		update.AIStack[1].Action = uint32(ai.ACTION_FIGHT)
		if !new(Server).MonsterMainNative547210(unit) {
			t.Fatal("uninterruptible dependency was not handled")
		}
	})
	t.Run("dead", func(t *testing.T) {
		unit := passiveMonsterTestObject547210(t)
		unit.ObjFlags |= object.FlagDead
		unit.UpdateDataMonster().AIStack[0].Action = uint32(ai.ACTION_FIGHT)
		if !new(Server).MonsterMainNative547210(unit) {
			t.Fatal("dead monster was not handled")
		}
	})
}

func TestMonsterMainNative547210RejectsUnportedPassiveBranches(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*Object, *MonsterUpdateData)
	}{
		{name: "buff", setup: func(unit *Object, _ *MonsterUpdateData) { unit.Buffs = 1 << ENCHANT_CONFUSED }},
		{name: "moderate aggression", setup: func(_ *Object, update *MonsterUpdateData) { update.Aggression = 0.08 }},
		{name: "casting", setup: func(_ *Object, update *MonsterUpdateData) { update.StatusFlags |= object.MonStatusCanCastSpells }},
		{name: "blocking", setup: func(_ *Object, update *MonsterUpdateData) { update.StatusFlags |= object.MonStatusCanBlock }},
		{name: "mimic bot", setup: func(_ *Object, update *MonsterUpdateData) { update.StatusFlags |= object.MonStatusBot }},
		{name: "NPC weapon block", setup: func(unit *Object, update *MonsterUpdateData) {
			unit.ObjSubClass |= object.SubClass(object.MonsterNPC)
			update.WeaponEquipFlags = 0x400
		}},
		{name: "NPC shield block", setup: func(unit *Object, update *MonsterUpdateData) {
			unit.ObjSubClass |= object.SubClass(object.MonsterNPC)
			update.ArmorEquipFlags = 0x1000000
		}},
		{name: "active fight", setup: func(_ *Object, update *MonsterUpdateData) {
			update.AIStack[0].Action = uint32(ai.ACTION_FIGHT)
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			unit := passiveMonsterTestObject547210(t)
			tc.setup(unit, unit.UpdateDataMonster())
			if new(Server).MonsterMainNative547210(unit) {
				t.Fatal("unported main-AI branch was handled")
			}
		})
	}
}

func TestMonsterMainNative547210LoadedLowAggressionNPC(t *testing.T) {
	s := new(Server)
	s.SetFrame(13)
	unit := passiveMonsterTestObject547210(t)
	unit.NetCode = 1348
	unit.ObjSubClass = object.SubClass(0x11012)
	unit.HealthData = &HealthData{}
	unit.InvFirstItem = &Object{}
	update := unit.UpdateDataMonster()
	update.AIStack[0].Args[0] = 1
	update.AIStack[0].Field5 = 1
	update.Field137 = 1
	update.Field127 = 1
	update.WeaponEquipFlags = 0x800
	update.ArmorEquipFlags = 0x22a8a
	update.CurrentEnemy = &Object{}
	update.MonsterDef = &MonsterDef{}
	beforeUnit := *unit
	beforeUpdate := *update

	if !s.MonsterMainNative547210(unit) {
		t.Fatal("loaded low-aggression NPC was not handled")
	}
	if *unit != beforeUnit || *update != beforeUpdate {
		t.Fatal("loaded low-aggression NPC state changed")
	}
}

func TestMonsterMainNative547210RejectsCoopConversationCandidate(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameModeCoop)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})
	unit := passiveMonsterTestObject547210(t)
	unit.Field5 = 0x10
	if new(Server).MonsterMainNative547210(unit) {
		t.Fatal("co-op conversation candidate was handled as passive")
	}
}

func TestMonsterMainNative547210PassiveAfterConversationCheck(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameModeCoop)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	s := new(Server)
	s.SetTickRate(30)
	s.SetFrame(678)
	unit := passiveMonsterTestObject547210(t)
	unit.NetCode = 3800
	unit.Field5 = 0x10
	unit.ObjSubClass = object.SubClass(object.MonsterNPC)
	unit.HealthData = &HealthData{Cur: 5000, Field2: 5000, Max: 5000}
	unit.SpeedBase = 1.9052104
	update := unit.UpdateDataMonster()
	update.AIStack[0].Action = uint32(ai.ACTION_GUARD)
	update.Aggression = 0
	update.RetreatLevel = 0.5
	update.StatusFlags = object.MonStatusCanSeeFriends
	update.CurrentEnemy = &Object{}
	update.MonsterDef = &MonsterDef{}
	before := *update

	if !s.MonsterMainNativeRuntime547210(unit, MonsterMainRuntime547210{
		GUICursorActive:    func() bool { return false },
		FindObjectAtCursor: func(*Object) *Object { return nil },
	}) {
		t.Fatal("passive NPC was not handled after its conversation predicate was checked")
	}
	if *update != before {
		t.Fatal("passive NPC state changed after the conversation predicate")
	}

	update.AIStackInd = 4
	update.AIStack[4] = AIStackItem{Action: uint32(ai.ACTION_FACE_LOCATION)}
	before = *update
	if !s.MonsterMainNativeRuntime547210(unit, MonsterMainRuntime547210{
		GUICursorActive:    func() bool { return false },
		FindObjectAtCursor: func(*Object) *Object { return nil },
	}) {
		t.Fatal("passive NPC face action was not handled after its conversation predicate was checked")
	}
	if *update != before {
		t.Fatal("passive NPC face action changed main-AI state")
	}
}

func TestMonsterMainNative547210QuiescentDodgeMonster(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameModeCoop)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	s := new(Server)
	s.SetFrame(1)
	unit := passiveMonsterTestObject547210(t)
	unit.NetCode = 15
	unit.HealthData = &HealthData{Cur: 8, Max: 8, Field2: 8}
	unit.UpdateDataMonster().Aggression = 0.5
	unit.UpdateDataMonster().RetreatLevel = 0.25
	unit.UpdateDataMonster().StatusFlags = object.MonStatusCanDodge
	unit.UpdateDataMonster().MonsterDef = &MonsterDef{StatusFlags92: object.MonStatusCanDodge}
	before := *unit.UpdateDataMonster()
	if !s.MonsterMainNative547210(unit) {
		t.Fatal("quiescent dodge monster was not handled")
	}
	if *unit.UpdateDataMonster() != before {
		t.Fatal("quiescent dodge monster state changed")
	}

	s.Map.Init()
	missile := &Object{ObjClass: object.ClassMissile, ObjFlags: object.FlagActive, PosVec: unit.PosVec, NewPos: unit.PosVec}
	s.Map.AddObjectToIndex(missile)
	if s.monsterMainQuiescentNoop547210(unit, unit.UpdateDataMonster()) {
		t.Fatal("nearby missile was ignored")
	}
}

func TestMonsterMainNative547210QuiescentRetreatCooldown(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameModeCoop)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	s := new(Server)
	s.SetTickRate(30)
	s.SetFrame(3)
	unit := passiveMonsterTestObject547210(t)
	unit.ObjSubClass = object.SubClass(object.MonsterFemaleNPC | object.MonsterNPC | object.MonsterNoTarget)
	unit.HealthData = &HealthData{Max: 75}
	update := unit.UpdateDataMonster()
	update.AIStack[0].Action = uint32(ai.ACTION_GUARD)
	update.StatusFlags = 0
	update.MonsterDef = &MonsterDef{}
	update.Field127 = 0
	before := *update

	if !s.monsterMainPassiveNoop547210(unit, update) {
		t.Fatal("newly loaded NPC retreat cooldown was not handled")
	}
	if *update != before {
		t.Fatal("retreat cooldown path changed monster state")
	}

	s.SetFrame(96)
	if s.monsterMainPassiveNoop547210(unit, update) {
		t.Fatal("expired retreat cooldown hid the original retreat branch")
	}
}

func TestMonsterMainNative547210RetreatTransition(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameModeCoop)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	s := new(Server)
	s.handle = atomic.AddUintptr(&serverLast, 1)
	servers.Store(s.handle, s)
	t.Cleanup(func() { servers.Delete(s.handle) })
	s.SetTickRate(30)
	s.SetFrame(90)
	unit := passiveMonsterTestObject547210(t)
	unit.serverHandle = s.handle
	unit.NetCode = 6
	unit.ObjSubClass = object.SubClass(object.MonsterMedium | object.MonsterUndead | object.MonsterImmuneFear | object.MonsterHasSoul)
	unit.HealthData = &HealthData{Cur: 4, Field2: 4, Max: 400}
	unit.SpeedBase = 2.99
	update := unit.UpdateDataMonster()
	update.AIStack[0].Action = uint32(ai.ACTION_GUARD)
	update.Aggression = 0
	update.RetreatLevel = 0.07
	update.StatusFlags = 0
	update.Field127 = 0
	sounds := [14]uint32{}
	sounds[13] = 773
	update.SoundSet122 = unsafe.Pointer(&sounds[0])
	update.ScriptRetreat = ScriptCallback{Flags: 0xa5, Func: 17}

	var soundID uint32
	var scriptBlock *ScriptCallback
	var scriptCaller, scriptTrigger *Object
	var scriptEvent ScriptEventType
	if !s.MonsterMainNativeRuntime547210(unit, MonsterMainRuntime547210{
		AudioEvent: func(id uint32, got *Object) {
			soundID = id
			if got != unit {
				t.Fatalf("sound unit = %p, want %p", got, unit)
			}
		},
		ScriptCallback: func(block *ScriptCallback, caller, trigger *Object, event ScriptEventType) {
			scriptBlock, scriptCaller, scriptTrigger, scriptEvent = block, caller, trigger, event
		},
	}) {
		t.Fatal("eligible retreat transition was not handled")
	}
	if update.AIStackInd != 2 || update.AIStack[1].Type() != ai.DEPENDENCY_NOT_CORNERED ||
		update.AIStack[2].Type() != ai.ACTION_RETREAT {
		t.Fatalf("retreat stack = %#v", update.GetAIStack())
	}
	if soundID != 773 {
		t.Fatalf("retreat sound = %d, want 773", soundID)
	}
	if scriptBlock != &update.ScriptRetreat || scriptCaller != nil || scriptTrigger != unit ||
		scriptEvent != NoxEventMonsterMoveXXX {
		t.Fatalf("retreat callback = %p/%p/%p/%v", scriptBlock, scriptCaller, scriptTrigger, scriptEvent)
	}
}

func TestMonsterMainNative547210MaidenScriptedMoveRetreat(t *testing.T) {
	s := new(Server)
	s.handle = atomic.AddUintptr(&serverLast, 1)
	servers.Store(s.handle, s)
	t.Cleanup(func() { servers.Delete(s.handle) })
	s.SetTickRate(30)
	s.SetFrame(842)

	unit := passiveMonsterTestObject547210(t)
	unit.serverHandle = s.handle
	unit.NetCode = 605
	unit.ObjSubClass = object.SubClass(0x11022)
	unit.Field5 = 0
	unit.HealthData = &HealthData{Cur: 11, Field2: 11, Max: 75}
	unit.SpeedBase = 1.95
	update := unit.UpdateDataMonster()
	update.AIStackInd = 6
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_GUARD)}
	update.AIStack[1] = AIStackItem{Action: uint32(ai.DEPENDENCY_NO_VISIBLE_ENEMY)}
	update.AIStack[2] = AIStackItem{Action: uint32(ai.ACTION_WAIT_RELATIVE), Args: [4]uintptr{900}}
	update.AIStack[3] = AIStackItem{Action: uint32(ai.DEPENDENCY_LOCATION_FARTHER_THAN)}
	update.AIStack[4] = AIStackItem{Action: uint32(ai.ACTION_FACE_LOCATION)}
	update.AIStack[5] = AIStackItem{Action: uint32(ai.DEPENDENCY_LOCATION_CLOSER_THAN)}
	update.AIStack[6] = AIStackItem{Action: uint32(ai.ACTION_MOVE_TO)}
	update.Aggression = 0
	update.RetreatLevel = 0.5
	update.StatusFlags = object.MonStatusCanSeeFriends | object.MonStatusRunning | object.MonStatusAlwaysRun
	update.Field127 = 0
	update.MonsterDef = &MonsterDef{}
	before := update.AIStack

	if !s.MonsterMainNative547210(unit) {
		t.Fatal("War01A Maiden scripted MOVE_TO retreat was not handled")
	}
	if update.AIStackInd != 8 {
		t.Fatalf("AIStackInd = %d, want 8", update.AIStackInd)
	}
	for i := 0; i <= 6; i++ {
		if update.AIStack[i] != before[i] {
			t.Fatalf("scripted stack[%d] changed: %#v, want %#v", i, update.AIStack[i], before[i])
		}
	}
	if update.AIStack[7].Type() != ai.DEPENDENCY_NOT_CORNERED ||
		update.AIStack[8].Type() != ai.ACTION_RETREAT {
		t.Fatalf("retreat suffix = %#v", update.GetAIStack()[7:])
	}
}

func TestMonsterMainPopAttackActions5471B0(t *testing.T) {
	s := new(Server)
	s.handle = atomic.AddUintptr(&serverLast, 1)
	servers.Store(s.handle, s)
	t.Cleanup(func() { servers.Delete(s.handle) })

	unit := passiveMonsterTestObject547210(t)
	unit.serverHandle = s.handle
	update := unit.UpdateDataMonster()
	update.AIStackInd = 4
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_GUARD)}
	update.AIStack[1] = AIStackItem{Action: uint32(ai.DEPENDENCY_CAN_SEE)}
	update.AIStack[2] = AIStackItem{Action: uint32(ai.ACTION_MELEE_ATTACK)}
	update.AIStack[3] = AIStackItem{Action: uint32(ai.DEPENDENCY_TIME)}
	update.AIStack[4] = AIStackItem{Action: uint32(ai.ACTION_FACE_OBJECT)}

	s.monsterMainPopAttackActions5471B0(unit)
	if update.AIStackInd != 0 || update.AIStack[0].Type() != ai.ACTION_GUARD {
		t.Fatalf("remaining stack = %#v, want GUARD", update.GetAIStack())
	}
}

func TestMonsterMainNative547210ConjurerModerateRetreat(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameModeCoop)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	s := new(Server)
	s.handle = atomic.AddUintptr(&serverLast, 1)
	servers.Store(s.handle, s)
	t.Cleanup(func() { servers.Delete(s.handle) })
	s.SetTickRate(30)
	s.SetFrame(105)
	unit := passiveMonsterTestObject547210(t)
	unit.serverHandle = s.handle
	unit.NetCode = 1096
	unit.ObjSubClass = object.SubClass(0x10002)
	unit.Field5 = 0x10
	unit.PosVec = types.Ptf(1953, 4361)
	unit.HealthData = &HealthData{Cur: 75, Field2: 74, Max: 5000}
	unit.SpeedBase = 1.8085278
	update := unit.UpdateDataMonster()
	update.AIStackInd = 0
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_GUARD), Field5: 1}
	update.Aggression = 0.5
	update.RetreatLevel = 0.5
	update.StatusFlags = 0
	update.Field127 = 0
	update.Field137 = 1
	update.MonsterDef = &MonsterDef{}
	sounds := [14]uint32{}
	sounds[13] = 0x11223344
	update.SoundSet122 = unsafe.Pointer(&sounds[0])
	update.ScriptRetreat = ScriptCallback{Flags: 0xa5, Func: 17}

	var soundID uint32
	var scriptBlock *ScriptCallback
	if !s.MonsterMainNativeRuntime547210(unit, MonsterMainRuntime547210{
		GUICursorActive:    func() bool { return true },
		FindObjectAtCursor: func(*Object) *Object { return unit },
		AudioEvent: func(id uint32, got *Object) {
			soundID = id
			if got != unit {
				t.Fatalf("sound unit = %p, want %p", got, unit)
			}
		},
		ScriptCallback: func(block *ScriptCallback, caller, trigger *Object, event ScriptEventType) {
			if caller != nil || trigger != unit || event != NoxEventMonsterMoveXXX {
				t.Fatalf("retreat callback args = %p/%p/%v", caller, trigger, event)
			}
			scriptBlock = block
		},
	}) {
		t.Fatal("Con01A moderate-aggression retreat was not handled")
	}
	if update.AIStackInd != 2 || update.AIStack[1].Type() != ai.DEPENDENCY_NOT_CORNERED ||
		update.AIStack[2].Type() != ai.ACTION_RETREAT {
		t.Fatalf("retreat stack = %#v", update.GetAIStack())
	}
	if soundID != sounds[13] || scriptBlock != &update.ScriptRetreat {
		t.Fatalf("retreat effects = sound %#x callback %p", soundID, scriptBlock)
	}
}

func TestMonsterMainNative547210ModerateMeleeCombat(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	s := new(Server)
	s.SetTickRate(30)
	s.SetFrame(670)
	unit := passiveMonsterTestObject547210(t)
	unit.NetCode = 3816
	unit.ObjSubClass = object.SubClass(object.MonsterMedium | object.MonsterImmunePoison)
	unit.PosVec = types.Ptf(4484.5, 2104.5)
	unit.HealthData = &HealthData{Cur: 12, Field2: 12, Max: 12}
	unit.SpeedBase = 1.8233368
	update := unit.UpdateDataMonster()
	update.AIStack[0].Action = uint32(ai.ACTION_IDLE)
	update.Aggression = 0.5
	update.SightRange = 150
	update.FleeRange = 0
	update.RetreatLevel = 0
	update.StatusFlags = 0
	update.MonsterDef = &MonsterDef{MeleeAttackRange112: 15}
	update.Field127 = 662
	update.Field137 = 662
	player := &Object{PosVec: types.Ptf(4404.5, 2104.5)}
	update.CurrentEnemy = player
	update.PreferredEnemy = player
	before := *update

	if !s.MonsterMainNative547210(unit) {
		t.Fatal("full-health moderate melee IDLE was not handled")
	}
	if *update != before {
		t.Fatal("moderate melee IDLE main AI changed state")
	}

	update.AIStackInd = 1
	update.AIStack[1] = AIStackItem{Action: uint32(ai.ACTION_FIGHT), Field5: 1}
	before = *update
	if !s.MonsterMainNative547210(unit) {
		t.Fatal("full-health moderate melee FIGHT was not handled")
	}
	if *update != before {
		t.Fatal("moderate melee FIGHT main AI changed state")
	}

	unit.HealthData.Cur = 7
	update.StatusFlags = object.MonStatusAlert | object.MonStatusRunning | object.MonStatusInjured
	before = *update
	if !s.MonsterMainNative547210(unit) {
		t.Fatal("injured moderate melee FIGHT was not handled")
	}
	if *update != before {
		t.Fatal("injured moderate melee FIGHT main AI changed state")
	}
	unit.HealthData.Cur = unit.HealthData.Max

	update.StatusFlags = object.MonStatusAlert | object.MonStatusRunning
	update.AIStackInd = 7
	update.AIStack[7] = AIStackItem{Action: uint32(ai.ACTION_MOVE_TO)}
	update.Field124 = s.Frame()
	update.Field125 = math.Float32bits(unit.PosVec.X)
	update.Field126 = math.Float32bits(unit.PosVec.Y)
	before = *update
	if !s.MonsterMainNative547210(unit) {
		t.Fatal("active moderate melee pursuit was not handled")
	}
	if *update != before {
		t.Fatal("active moderate melee pursuit main AI changed state")
	}

	update.StatusFlags |= object.MonStatusCanCastSpells
	if s.monsterMainModerateCombatStable547210(unit, update, MonsterMainRuntime547210{}) {
		t.Fatal("caster combat state was treated as a proved melee no-op")
	}
	update.StatusFlags = object.MonStatusAlert | object.MonStatusRunning

	update.FleeRange = 100
	if s.monsterMainModerateCombatStable547210(unit, update, MonsterMainRuntime547210{}) {
		t.Fatal("nonzero flee range was treated as a proved combat no-op")
	}
}

func TestMonsterMainNative547210WizardModerateScriptedFace(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameModeCoop)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	s := new(Server)
	s.SetTickRate(30)
	s.SetFrame(14)
	unit := passiveMonsterTestObject547210(t)
	unit.NetCode = 1379
	unit.ObjSubClass = object.SubClass(0x10002)
	unit.Field5 = 0x10
	unit.PosVec = types.Ptf(2508, 3289)
	unit.HealthData = &HealthData{Max: 200}
	unit.SpeedBase = 1.8085278
	update := unit.UpdateDataMonster()
	update.AIStackInd = 2
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_GUARD), Field5: 1}
	update.AIStack[1] = AIStackItem{Action: uint32(ai.DEPENDENCY_NO_VISIBLE_ENEMY)}
	update.AIStack[2] = AIStackItem{Action: uint32(ai.ACTION_FACE_ANGLE), Args: [4]uintptr{85}}
	update.Aggression = 0.5
	update.RetreatLevel = 0.07
	update.StatusFlags = object.MonStatusCanDodge | object.MonStatusCanCastSpells |
		object.MonStatusHoldYourGround | object.MonStatusCanSeeFriends
	update.Field124 = 13
	update.Field127 = 0
	update.Field137 = 13
	update.Field363 = 15
	update.MonsterDef = &MonsterDef{StatusFlags92: object.MonStatusCanDodge | object.MonStatusCanCastSpells}
	beforeUnit := *unit
	beforeUpdate := *update

	if !s.MonsterMainNativeRuntime547210(unit, MonsterMainRuntime547210{
		GUICursorActive:    func() bool { return false },
		FindObjectAtCursor: func(*Object) *Object { return nil },
	}) {
		t.Fatal("Wiz01A moderate-aggression scripted face was not handled")
	}
	if *unit != beforeUnit || *update != beforeUpdate {
		t.Fatal("stable scripted-face state changed")
	}
}

func TestMonsterMainNative547210GuardMoveNearbyEnemyStimulus(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	s := new(Server)
	s.SetTickRate(30)
	s.SetFrame(672)
	unit := passiveMonsterTestObject547210(t)
	unit.ObjFlags |= object.FlagEnabled
	unit.NetCode = 1426
	unit.PosVec = types.Ptf(4503.3633, 2105.1082)
	unit.NewPos = unit.PosVec
	unit.PrevPos = types.Ptf(4504.5, 2104.5)
	unit.SpeedBase = 2.4523568
	unit.HealthData = &HealthData{Cur: 8, Field2: 8, Max: 8}
	update := unit.UpdateDataMonster()
	update.AIStackInd = 2
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_GUARD), Field5: 1}
	update.AIStack[1] = AIStackItem{Action: uint32(ai.DEPENDENCY_NOT_UNDER_ATTACK), Field5: 1}
	update.AIStack[2] = AIStackItem{Action: uint32(ai.ACTION_MOVE_TO), Field5: 1}
	update.Field124 = 671
	update.Field125 = math.Float32bits(4504.5)
	update.Field126 = math.Float32bits(2104.5)
	update.Field137 = 670
	update.Aggression = 0.5
	update.MonsterDef = &MonsterDef{}

	found := &Object{ObjClass: object.ClassPlayer}
	var calls int
	if !s.MonsterMainNativeRuntime547210(unit, MonsterMainRuntime547210{
		EnemyAggro: func(got *Object, radius float32) *Object {
			calls++
			if got != unit || radius != 100 {
				t.Fatalf("enemy query = (%p,%g), want (%p,100)", got, radius, unit)
			}
			return found
		},
	}) {
		t.Fatal("War01A placed Bat guard-move transition was not handled")
	}
	if calls != 1 {
		t.Fatalf("enemy query calls = %d, want 1", calls)
	}
	if !update.StatusFlags.Has(object.MonStatusInjured) || unit.Obj130 != nil ||
		unit.Field131 != 11 || unit.Frame134 != 672 {
		t.Fatalf("nearby-enemy stimulus = status:%#x source:%p type:%d frame:%d",
			uint32(update.StatusFlags), unit.Obj130, unit.Field131, unit.Frame134)
	}
	if update.Field124 != 671 || math.Float32frombits(update.Field125) != 4504.5 ||
		math.Float32frombits(update.Field126) != 2104.5 {
		t.Fatalf("movement tracking changed = %d/{%g %g}", update.Field124,
			math.Float32frombits(update.Field125), math.Float32frombits(update.Field126))
	}

	s.SetFrame(688)
	update.StatusFlags = 0
	update.CurrentEnemy = found
	update.Field124 = 687
	update.Field125 = math.Float32bits(unit.PosVec.X)
	update.Field126 = math.Float32bits(unit.PosVec.Y)
	unit.Obj130 = nil
	unit.Field131 = 0
	unit.Frame134 = 0
	if !s.MonsterMainNativeRuntime547210(unit, MonsterMainRuntime547210{
		EnemyAggro: func(got *Object, radius float32) *Object {
			calls++
			if got != unit || radius != 100 {
				t.Fatalf("combat enemy query = (%p,%g), want (%p,100)", got, radius, unit)
			}
			return found
		},
	}) {
		t.Fatal("War01A placed Bat guard-move combat transition was not handled")
	}
	if calls != 2 {
		t.Fatalf("enemy query calls = %d, want 2", calls)
	}
	if !update.StatusFlags.Has(object.MonStatusInjured) || unit.Obj130 != found ||
		unit.Field131 != 11 || unit.Frame134 != 688 {
		t.Fatalf("combat nearby-enemy stimulus = status:%#x source:%p type:%d frame:%d",
			uint32(update.StatusFlags), unit.Obj130, unit.Field131, unit.Frame134)
	}
}

func TestMonsterMainNative547210RetreatAllowsMovementStatus(t *testing.T) {
	s := new(Server)
	s.handle = atomic.AddUintptr(&serverLast, 1)
	servers.Store(s.handle, s)
	t.Cleanup(func() { servers.Delete(s.handle) })
	s.SetTickRate(30)
	s.SetFrame(100)
	unit := passiveMonsterTestObject547210(t)
	unit.serverHandle = s.handle
	unit.NetCode = 12
	unit.HealthData = &HealthData{Cur: 1, Field2: 1, Max: 75}
	update := unit.UpdateDataMonster()
	update.MonsterDef = &MonsterDef{}
	update.Field127 = 0
	if !s.MonsterMainNative547210(unit) {
		t.Fatal("eligible retreat with movement status was not handled")
	}
	if !update.HasAction(ai.ACTION_RETREAT) {
		t.Fatalf("retreat stack = %#v", update.GetAIStack())
	}
	update.AIStackInd = 0
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_IDLE)}
	update.StatusFlags |= object.MonStatusCanCastSpells
	if s.monsterMainRetreat547210(unit, update, MonsterMainRuntime547210{}) {
		t.Fatal("unported caster branch entered native retreat")
	}
}

func TestMonsterMainNative547210PassiveCaster(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameModeCoop)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	s := new(Server)
	s.SetTickRate(30)
	s.SetFrame(15)
	unit := passiveMonsterTestObject547210(t)
	unit.NetCode = 1
	unit.ObjSubClass = object.SubClass(object.MonsterMedium | object.MonsterNPC | object.MonsterImmuneFear | object.MonsterHasSoul)
	unit.HealthData = &HealthData{Cur: 1000, Field2: 1000, Max: 1000}
	unit.InvFirstItem = &Object{}
	unit.Buffs = 1<<ENCHANT_INVISIBLE | 1<<ENCHANT_SHIELD
	update := unit.UpdateDataMonster()
	update.StatusFlags = object.MonStatusCanCastSpells | object.MonStatusHoldYourGround | object.MonStatusNeverRun
	update.MonsterDef = &MonsterDef{}
	update.RetreatLevel = 0.25
	update.ArmorEquipFlags = 0x84848
	update.CurrentEnemy = &Object{PosVec: types.Ptf(130, 130)}
	update.Field127 = 15
	// The active inversion predicate short-circuits on its exact cooldown,
	// allowing this test to remain independent of an on-disk balance table.
	update.Field363 = 16
	before := *update

	if !s.MonsterMainNative547210(unit) {
		t.Fatal("passive caster state was not handled")
	}
	if *update != before {
		t.Fatal("passive caster main AI changed state")
	}
}

func TestMonsterMainNative547210FullHealthWaitCadence(t *testing.T) {
	s := new(Server)
	s.SetTickRate(30)
	s.SetFrame(16)
	unit := passiveMonsterTestObject547210(t)
	unit.ObjSubClass = object.SubClass(object.MonsterSmall | object.MonsterWarcryStun | object.MonsterLookAround)
	unit.HealthData = &HealthData{Cur: 1, Field2: 1, Max: 1}
	unit.SpeedBase = 1.7
	update := unit.UpdateDataMonster()
	update.AIStack[0].Action = uint32(ai.ACTION_WAIT)
	update.Aggression = 0.16
	update.RetreatLevel = 0.5
	update.StatusFlags = 0
	update.MonsterDef = &MonsterDef{}
	update.Field127 = 16
	before := *update

	if !s.MonsterMainNative547210(unit) {
		t.Fatal("full-health WAIT was not handled on the edible-search cadence")
	}
	if *update != before {
		t.Fatal("full-health WAIT changed monster state")
	}

	unit.HealthData = &HealthData{Max: 1}
	if s.MonsterMainNative547210(unit) {
		t.Fatal("hungry WAIT hid the original edible-search branch")
	}
}

func TestMonsterMainNative547210AmbientIdle(t *testing.T) {
	s := new(Server)
	s.SetTickRate(30)
	s.SetFrame(43)
	unit := passiveMonsterTestObject547210(t)
	unit.NetCode = 5
	unit.ObjSubClass = object.SubClass(object.MonsterSmall | object.MonsterWarcryStun | object.MonsterLookAround)
	unit.PosVec = types.Ptf(100, 100)
	unit.HealthData = &HealthData{Cur: 1, Field2: 1, Max: 1}
	unit.SpeedBase = 1.35
	update := unit.UpdateDataMonster()
	update.Aggression = 0.16
	update.RetreatLevel = 0.5
	update.FleeRange = 50
	update.StatusFlags = 0
	update.MonsterDef = &MonsterDef{}
	update.Field127 = 43
	update.CurrentEnemy = &Object{PosVec: types.Ptf(310, 100)}
	before := *update

	if !s.MonsterMainNative547210(unit) {
		t.Fatal("ambient IDLE with a far retained enemy was not handled")
	}
	if *update != before {
		t.Fatal("ambient IDLE changed monster state")
	}

	s.SetFrame(100)
	unit.NetCode = 12
	update.Field127 = 0
	update.CurrentEnemy = &Object{PosVec: types.Ptf(120, 100)}
	if s.monsterMainAmbientIdleNoop547210(unit, update) {
		t.Fatal("nearby enemy flee branch was handled after cooldown")
	}
}

func TestMonsterMainNative547210LowAggressionRandomWalk(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	s := new(Server)
	s.SetTickRate(30)
	s.SetFrame(662)
	unit := passiveMonsterTestObject547210(t)
	unit.NetCode = 3816
	unit.ObjSubClass = object.SubClass(object.MonsterSmall | object.MonsterWarcryStun | object.MonsterLookAround)
	unit.PosVec = types.Ptf(4484.5, 2104.5)
	unit.HealthData = &HealthData{Cur: 1, Field2: 1, Max: 1}
	unit.SpeedBase = 1.2155579
	update := unit.UpdateDataMonster()
	update.AIStack[0].Action = uint32(ai.ACTION_RANDOM_WALK)
	update.Aggression = 0.16
	update.SightRange = 150
	update.FleeRange = 40
	update.RetreatLevel = 0.5
	update.StatusFlags = 0
	update.MonsterDef = &MonsterDef{}
	update.Field127 = 662
	update.Field137 = 662
	update.CurrentEnemy = &Object{PosVec: types.Ptf(4404.5, 2104.5)}
	before := *update

	if !s.MonsterMainNative547210(unit) {
		t.Fatal("full-health low-aggression RANDOM_WALK was not handled")
	}
	if *update != before {
		t.Fatal("low-aggression RANDOM_WALK main AI changed state")
	}

	update.CurrentEnemy = &Object{PosVec: types.Ptf(4464.5, 2104.5)}
	if s.monsterMainLowAggressionRandomWalkNoop547210(unit, update) {
		t.Fatal("nearby enemy flee branch was treated as a no-op")
	}
	update.CurrentEnemy = nil
	unit.HealthData.Cur = 0
	if s.monsterMainLowAggressionRandomWalkNoop547210(unit, update) {
		t.Fatal("hungry random-walk state was treated as a no-op")
	}
}

func TestMonsterMainInversionThreat547210(t *testing.T) {
	s := new(Server)
	s.Map.Init()
	unit := passiveMonsterTestObject547210(t)
	unit.PosVec = types.Ptf(100, 100)

	addMissile := func(pos types.Pointf, subclass object.MissileClass, target *Object) {
		missile := &Object{
			ObjClass:    object.ClassMissile,
			ObjSubClass: object.SubClass(subclass),
			ObjFlags:    object.FlagActive,
			PosVec:      pos,
			NewPos:      pos,
			UpdateData:  unsafe.Pointer(&MissileUpdateData{Target: target}),
		}
		s.Map.AddObjectToIndex(missile)
	}

	addMissile(types.Ptf(100, 100), 0, unit)
	addMissile(types.Ptf(100, 100), object.MissileMagic, &Object{})
	addMissile(types.Ptf(300, 300), object.MissileMagic, unit)
	if s.monsterMainInversionThreat547210(unit, 50) {
		t.Fatal("non-magic, wrong-target, or out-of-range missile triggered inversion")
	}
	addMissile(types.Ptf(110, 110), object.MissileMagic, unit)
	if !s.monsterMainInversionThreat547210(unit, 50) {
		t.Fatal("targeted magic missile did not trigger inversion")
	}
}

func TestMonsterMainNative547210ScriptedFaceObject(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameModeCoop)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	s := new(Server)
	s.SetTickRate(30)
	s.SetFrame(14)
	unit := passiveMonsterTestObject547210(t)
	unit.ObjSubClass = object.SubClass(object.MonsterNPC | object.MonsterFemaleNPC | object.MonsterNoTarget)
	unit.HealthData = &HealthData{}
	unit.InvFirstItem = &Object{}
	update := unit.UpdateDataMonster()
	update.AIStackInd = 1
	update.AIStack[0].Action = uint32(ai.ACTION_GUARD)
	update.AIStack[1].Action = uint32(ai.ACTION_FACE_OBJECT)
	target := monsterActionTestObject50A910(t)
	update.AIStack[1].SetArgs(target)
	update.StatusFlags = object.MonStatusCanSeeFriends | object.MonStatusCanDodge
	update.MonsterDef = &MonsterDef{}
	update.WeaponEquipFlags = 0x200
	update.ArmorEquipFlags = 0x422b0a
	before := *update

	if !s.MonsterMainNative547210(unit) {
		t.Fatal("passive scripted FACE_OBJECT state was not handled")
	}
	if *update != before {
		t.Fatal("scripted FACE_OBJECT main AI changed state")
	}
	if update.AIStack[1].ArgObj(0) != target {
		t.Fatal("native face target pointer changed")
	}

	for _, tc := range []struct {
		name  string
		setup func(*MonsterUpdateData)
	}{
		{"weapon block", func(update *MonsterUpdateData) { update.WeaponEquipFlags |= 0x400 }},
		{"shield block", func(update *MonsterUpdateData) { update.ArmorEquipFlags |= 0x1000000 }},
		{"definition dodge", func(update *MonsterUpdateData) { update.MonsterDef.StatusFlags92 |= object.MonStatusCanDodge }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			copyUpdate := before
			tc.setup(&copyUpdate)
			unit.UpdateData = unsafe.Pointer(&copyUpdate)
			if s.MonsterMainNative547210(unit) {
				t.Fatal("active FACE_OBJECT branch was treated as a no-op")
			}
		})
	}
}

func TestMonsterMainNative547210LoadedScriptedFaceLocation(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameModeCoop)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	s := new(Server)
	s.SetTickRate(30)
	s.SetFrame(1480)
	player := &Player{CursorVec: image.Pt(4580, 1930)}
	host := &Object{
		ObjClass:   object.ClassPlayer,
		ObjFlags:   object.FlagActive | object.FlagEnabled,
		PosVec:     types.Ptf(4443.74, 2072.0867),
		UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: player}),
	}
	s.Players.SetHost(player, host)
	unit := passiveMonsterTestObject547210(t)
	unit.NetCode = 1098
	unit.ObjSubClass = object.SubClass(0x10002)
	unit.ObjFlags = object.Flags(0x1080204)
	unit.Field5 = 0x10
	unit.PosVec = types.Ptf(4441.215, 2049.9827)
	unit.HealthData = &HealthData{Cur: 5000, Field2: 5000, Max: 5000}
	unit.SpeedBase = 1.8544486
	update := unit.UpdateDataMonster()
	update.AIStackInd = 1
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_GUARD), Args: [4]uintptr{uintptr(math.Float32bits(4441)), uintptr(math.Float32bits(2052)), 64}, Field5: 1}
	update.AIStack[1] = AIStackItem{Action: uint32(ai.ACTION_FACE_LOCATION), Args: [4]uintptr{uintptr(math.Float32bits(4441)), uintptr(math.Float32bits(2052))}}
	update.Aggression = 0
	update.RetreatLevel = 0.5
	update.StatusFlags = object.MonStatusCanSeeFriends
	update.MonsterDef = &MonsterDef{}
	update.Field124 = 1479
	update.Field127 = 1
	beforeUnit := *unit
	beforeUpdate := *update

	if !s.MonsterMainNative547210(unit) {
		t.Fatal("loaded scripted FACE_LOCATION state was not handled")
	}
	if *unit != beforeUnit || *update != beforeUpdate {
		t.Fatal("loaded scripted FACE_LOCATION state changed")
	}

	player.CursorVec = image.Pt(4441, 2050)
	if s.monsterMainScriptedFaceNoop547210(unit, update) {
		t.Fatal("near-cursor conversation candidate was treated as a no-op")
	}
}

func TestMonsterMainNative547210RoamTracking(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	newFish := func() (*Server, *Object, *MonsterUpdateData) {
		s := new(Server)
		s.SetTickRate(30)
		s.SetFrame(1)
		unit := passiveMonsterTestObject547210(t)
		unit.ObjSubClass = object.SubClass(object.MonsterSmall | object.MonsterWarcryStun | object.MonsterLookAround)
		unit.PosVec.X, unit.PosVec.Y = 100, 200
		unit.HealthData = &HealthData{Cur: 1, Field2: 1, Max: 1}
		unit.SpeedBase = 1.7
		update := unit.UpdateDataMonster()
		update.AIStack[0].Action = uint32(ai.ACTION_ROAM)
		update.Aggression = 0.16
		update.RetreatLevel = 0.5
		update.StatusFlags = 0
		update.MonsterDef = &MonsterDef{}
		return s, unit, update
	}

	t.Run("records movement", func(t *testing.T) {
		s, unit, update := newFish()
		if !s.MonsterMainNative547210(unit) {
			t.Fatal("ambient roam movement was not handled")
		}
		if update.Field124 != 1 || math.Float32frombits(update.Field125) != 100 || math.Float32frombits(update.Field126) != 200 {
			t.Fatalf("tracking = %d/%g/%g, want 1/100/200", update.Field124, math.Float32frombits(update.Field125), math.Float32frombits(update.Field126))
		}
	})
	t.Run("stationary before frustration deadline", func(t *testing.T) {
		s, unit, update := newFish()
		update.Field125 = math.Float32bits(unit.PosVec.X)
		update.Field126 = math.Float32bits(unit.PosVec.Y)
		if !s.MonsterMainNative547210(unit) {
			t.Fatal("pre-deadline ambient roam was not handled")
		}
	})
	t.Run("far enemy does not trigger flee", func(t *testing.T) {
		s, unit, update := newFish()
		update.FleeRange = 50
		update.CurrentEnemy = &Object{PosVec: types.Ptf(300, 200)}
		if !s.MonsterMainNative547210(unit) {
			t.Fatal("far enemy ambient roam was not handled")
		}
	})
	t.Run("rejects nearby enemy flee branch", func(t *testing.T) {
		s, unit, update := newFish()
		update.FleeRange = 50
		update.CurrentEnemy = &Object{PosVec: types.Ptf(120, 200)}
		if s.MonsterMainNative547210(unit) {
			t.Fatal("nearby enemy flee branch was handled")
		}
	})
	t.Run("rejects frustration branch", func(t *testing.T) {
		s, unit, update := newFish()
		update.Field125 = math.Float32bits(unit.PosVec.X)
		update.Field126 = math.Float32bits(unit.PosVec.Y)
		s.SetFrame(16)
		if s.MonsterMainNative547210(unit) {
			t.Fatal("randomized frustration branch was handled")
		}
	})
}

func TestMonsterMainNative547210PassiveRetreatRoamTracking(t *testing.T) {
	s := new(Server)
	s.SetTickRate(30)
	s.SetFrame(91)
	unit := passiveMonsterTestObject547210(t)
	unit.ObjSubClass = object.SubClass(object.MonsterMedium | object.MonsterUndead | object.MonsterHasSoul)
	unit.ObjFlags |= object.FlagEnabled
	unit.PosVec = types.Ptf(400, 500)
	unit.HealthData = &HealthData{Cur: 5, Field2: 4, Max: 400}
	unit.SpeedBase = 2.99
	update := unit.UpdateDataMonster()
	update.Aggression = 0
	update.RetreatLevel = 0.07
	update.MonsterDef = &MonsterDef{}
	update.AIStackInd = 5
	update.AIStack[0].Action = uint32(ai.DEPENDENCY_NOT_CORNERED)
	update.AIStack[1].Action = uint32(ai.ACTION_RETREAT)
	update.AIStack[2].Action = uint32(ai.DEPENDENCY_NOT_HEALTHY)
	update.AIStack[3].Action = uint32(ai.DEPENDENCY_NO_VISIBLE_ENEMY)
	update.AIStack[4].Action = uint32(ai.DEPENDENCY_NO_VISIBLE_FOOD)
	update.AIStack[5].Action = uint32(ai.ACTION_ROAM)

	if !s.MonsterMainNative547210(unit) {
		t.Fatal("passive retreat ROAM was not handled")
	}
	if update.Field124 != 91 || math.Float32frombits(update.Field125) != 400 || math.Float32frombits(update.Field126) != 500 {
		t.Fatalf("tracking = %d/%g/%g", update.Field124, math.Float32frombits(update.Field125), math.Float32frombits(update.Field126))
	}
	update.AIStack[5].Action = uint32(ai.ACTION_MOVE_TO)
	unit.PosVec = types.Ptf(440, 500)
	s.SetFrame(92)
	if !s.MonsterMainNative547210(unit) || update.Field124 != 92 || math.Float32frombits(update.Field125) != 440 {
		t.Fatal("MOVE_TO above passive RETREAT was not tracked")
	}
	s.SetFrame(106)
	if !s.MonsterMainNative547210(unit) {
		t.Fatal("passive retreat ROAM was not handled at the deadline")
	}
	s.SetFrame(108)
	if s.MonsterMainNative547210(unit) {
		t.Fatal("frustration branch was hidden after the deadline")
	}
}

func TestMonsterMainNative547210MaidenFrustrationWait(t *testing.T) {
	s := new(Server)
	s.handle = atomic.AddUintptr(&serverLast, 1)
	servers.Store(s.handle, s)
	t.Cleanup(func() { servers.Delete(s.handle) })
	s.SetTickRate(30)
	s.SetFrame(842)

	unit := passiveMonsterTestObject547210(t)
	unit.serverHandle = s.handle
	unit.NetCode = 605
	unit.ObjSubClass = object.SubClass(0x11022)
	unit.Field5 = 0
	unit.PosVec = types.Ptf(1121.1504, 3458.7422)
	unit.SpeedCur = 1.95
	unit.SpeedBase = 1.95
	unit.HealthData = &HealthData{Cur: 11, Field2: 11, Max: 75}
	update := unit.UpdateDataMonster()
	update.AIStackInd = 6
	update.AIStack[0] = AIStackItem{Action: uint32(ai.DEPENDENCY_NOT_CORNERED)}
	update.AIStack[1] = AIStackItem{Action: uint32(ai.ACTION_RETREAT), Field5: 1}
	update.AIStack[2] = AIStackItem{Action: uint32(ai.DEPENDENCY_NOT_HEALTHY)}
	update.AIStack[3] = AIStackItem{Action: uint32(ai.DEPENDENCY_NO_VISIBLE_ENEMY)}
	update.AIStack[4] = AIStackItem{Action: uint32(ai.DEPENDENCY_OBJECT_AT_VISIBLE_LOCATION)}
	update.AIStack[5] = AIStackItem{Action: uint32(ai.ACTION_PICKUP_OBJECT)}
	update.AIStack[6] = AIStackItem{Action: uint32(ai.ACTION_MOVE_TO), Field5: 1}
	update.Aggression = 0
	update.RetreatLevel = 0.5
	update.StatusFlags = object.MonStatusCanSeeFriends | object.MonStatusRunning | object.MonStatusAlwaysRun
	update.MonsterDef = &MonsterDef{}
	update.Field124 = 800
	update.Field125 = math.Float32bits(unit.PosVec.X)
	update.Field126 = math.Float32bits(unit.PosVec.Y)

	var randomCalls [][2]int
	runtime := MonsterMainRuntime547210{
		RandomInt: func(min, max int) int {
			randomCalls = append(randomCalls, [2]int{min, max})
			if len(randomCalls) == 1 {
				return 33
			}
			return 47
		},
		RandomFloat: func(float32, float32) float64 {
			t.Fatal("frustration wait path unexpectedly attempted a dodge")
			return 0
		},
		TileAt: func(types.Pointf) int { return 0 },
	}
	if !s.MonsterMainNativeRuntime547210(unit, runtime) {
		t.Fatal("War01A Maiden frustration state was not handled")
	}
	if !update.StatusFlags.Has(object.MonStatusFrustrated) {
		t.Fatal("frustrated status was not set")
	}
	if update.Field127 != 842 || update.Field124 != 842 ||
		math.Float32frombits(update.Field125) != unit.PosVec.X ||
		math.Float32frombits(update.Field126) != unit.PosVec.Y {
		t.Fatalf("frustration tracking = field127 %d field124 %d pos (%g,%g)",
			update.Field127, update.Field124, math.Float32frombits(update.Field125), math.Float32frombits(update.Field126))
	}
	if update.AIStackInd != 7 || update.AIStack[7].Type() != ai.ACTION_WAIT ||
		update.AIStack[7].ArgU32(0) != 889 {
		t.Fatalf("frustration stack = %#v", update.GetAIStack())
	}
	wantCalls := [][2]int{{0, 100}, {15, 60}}
	if len(randomCalls) != len(wantCalls) {
		t.Fatalf("random calls = %#v, want %#v", randomCalls, wantCalls)
	}
	for i := range wantCalls {
		if randomCalls[i] != wantCalls[i] {
			t.Fatalf("random call[%d] = %v, want %v", i, randomCalls[i], wantCalls[i])
		}
	}
}

func TestMonsterMainCheckDodgeables547C50(t *testing.T) {
	s := new(Server)
	s.handle = atomic.AddUintptr(&serverLast, 1)
	servers.Store(s.handle, s)
	t.Cleanup(func() { servers.Delete(s.handle) })
	s.SetTickRate(30)
	s.SetFrame(100)

	unit := passiveMonsterTestObject547210(t)
	unit.serverHandle = s.handle
	unit.PosVec = types.Ptf(100, 200)
	unit.Direction1 = 0
	unit.SpeedCur = 10
	update := unit.UpdateDataMonster()
	update.AIStackInd = 2
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_GUARD)}
	update.AIStack[1] = AIStackItem{Action: uint32(ai.DEPENDENCY_CAN_SEE)}
	update.AIStack[2] = AIStackItem{Action: uint32(ai.ACTION_FACE_OBJECT)}

	wantDestination := types.Ptf(100, 185)
	var gotRayFrom, gotRayTo types.Pointf
	var gotFlags MapTraceFlags
	var gotObstacleUnit *Object
	var gotObstacleFrom, gotObstacleTo types.Pointf
	runtime := MonsterMainRuntime547210{
		RandomFloat: func(min, max float32) float64 {
			if min != 2 || max != 3 {
				t.Fatalf("random float bounds = %g..%g", min, max)
			}
			return 2
		},
		RandomInt: func(min, max int) int {
			if min != 0 || max != 100 {
				t.Fatalf("random int bounds = %d..%d", min, max)
			}
			return 40
		},
		TraceRay: func(from, to types.Pointf, flags MapTraceFlags) bool {
			gotRayFrom, gotRayTo, gotFlags = from, to, flags
			return true
		},
		TraceObstacles: func(got *Object, from, to types.Pointf) bool {
			gotObstacleUnit, gotObstacleFrom, gotObstacleTo = got, from, to
			return true
		},
		TileAt: func(pos types.Pointf) int {
			if pos != wantDestination {
				t.Fatalf("tile point = %v, want %v", pos, wantDestination)
			}
			return 0
		},
	}
	if !s.monsterMainCheckDodgeables547C50(unit, runtime) {
		t.Fatal("clear lateral dodge was not scheduled")
	}
	if gotRayFrom != unit.PosVec || gotRayTo != wantDestination || gotFlags != MapTraceFlag1 {
		t.Fatalf("ray = %v -> %v flags %#x", gotRayFrom, gotRayTo, gotFlags)
	}
	if gotObstacleUnit != unit || gotObstacleFrom != unit.PosVec || gotObstacleTo != wantDestination {
		t.Fatalf("obstacle trace = %p %v -> %v", gotObstacleUnit, gotObstacleFrom, gotObstacleTo)
	}
	if update.AIStackInd != 2 || update.AIStack[0].Type() != ai.ACTION_GUARD ||
		update.AIStack[1].Type() != ai.DEPENDENCY_TIME || update.AIStack[1].ArgU32(0) != 130 ||
		update.AIStack[2].Type() != ai.ACTION_DODGE || update.AIStack[2].ArgPos(0) != wantDestination ||
		update.AIStack[2].ArgU32(2) != 0 {
		t.Fatalf("dodge stack = %#v", update.GetAIStack())
	}
}

func TestMonsterMainNative547210PassiveRetreatWait(t *testing.T) {
	s := new(Server)
	s.SetTickRate(30)
	s.SetFrame(92)
	unit := passiveMonsterTestObject547210(t)
	unit.HealthData = &HealthData{Cur: 5, Field2: 5, Max: 400}
	update := unit.UpdateDataMonster()
	update.Aggression = 0
	update.RetreatLevel = 0.07
	update.MonsterDef = &MonsterDef{}
	update.AIStackInd = 2
	update.AIStack[0].Action = uint32(ai.DEPENDENCY_NOT_CORNERED)
	update.AIStack[1].Action = uint32(ai.ACTION_RETREAT)
	update.AIStack[2].Action = uint32(ai.ACTION_WAIT)
	before := *update
	if !s.MonsterMainNative547210(unit) {
		t.Fatal("WAIT above passive RETREAT was not handled")
	}
	if *update != before {
		t.Fatal("passive RETREAT WAIT changed monster state")
	}
	update.AIStack[2].Action = uint32(ai.ACTION_FLEE)
	if s.MonsterMainNative547210(unit) {
		t.Fatal("moving FLEE head bypassed tracking")
	}
}

func TestMonsterMainNative547210ConversationNPCSoundWait(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameHost | noxflags.GameModeCoop)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	s := new(Server)
	s.SetTickRate(30)
	s.SetFrame(705)
	unit := passiveMonsterTestObject547210(t)
	unit.ObjSubClass = object.SubClass(0x10002)
	unit.Field5 = 0x10
	unit.HealthData = &HealthData{Cur: 5000, Field2: 5000, Max: 5000}
	unit.SpeedBase = 1.9052104
	update := unit.UpdateDataMonster()
	update.AIStackInd = 3
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_GUARD), Field5: 1}
	update.AIStack[1] = AIStackItem{Action: uint32(ai.DEPENDENCY_NO_INTERESTING_SOUND)}
	update.AIStack[2] = AIStackItem{Action: uint32(ai.DEPENDENCY_NO_VISIBLE_ENEMY)}
	update.AIStack[3] = AIStackItem{Action: uint32(ai.ACTION_WAIT), Args: [4]uintptr{757}}
	update.Field137 = 704
	update.StatusFlags = object.MonStatusCanSeeFriends
	update.Aggression = 0
	update.FleeRange = 0
	update.RetreatLevel = 0.5
	update.MonsterDef = &MonsterDef{}
	beforeUnit := *unit
	beforeUpdate := *update
	runtime := MonsterMainRuntime547210{
		GUICursorActive:    func() bool { return false },
		FindObjectAtCursor: func(*Object) *Object { return nil },
	}

	if !s.MonsterMainNativeRuntime547210(unit, runtime) {
		t.Fatal("conversation NPC sound WAIT stack was not handled")
	}
	if *unit != beforeUnit || *update != beforeUpdate {
		t.Fatal("conversation NPC sound WAIT stack changed state")
	}

	runtime.FindObjectAtCursor = nil
	if s.MonsterMainNativeRuntime547210(unit, runtime) {
		t.Fatal("conversation NPC WAIT without a cursor oracle was handled")
	}
}

func TestMonsterMainNative547210ConversationTransition(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameModeCoop)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	s := new(Server)
	s.handle = atomic.AddUintptr(&serverLast, 1)
	servers.Store(s.handle, s)
	t.Cleanup(func() { servers.Delete(s.handle) })
	s.SetTickRate(30)
	s.SetFrame(105)
	player := &Player{CursorVec: image.Pt(1953, 4361)}
	hostUpdate := &PlayerUpdateData{Player: player}
	host := &Object{
		ObjClass:   object.ClassPlayer,
		ObjFlags:   object.FlagActive | object.FlagEnabled,
		UpdateData: unsafe.Pointer(hostUpdate),
	}
	s.Players.SetHost(player, host)
	unit := passiveMonsterTestObject547210(t)
	unit.serverHandle = s.handle
	unit.NetCode = 1096
	unit.Field5 = 0x10
	unit.PosVec = types.Ptf(1953, 4361)
	unit.PrevPos = types.Ptf(-1, -2)
	unit.VelVec = types.Ptf(1, 2)
	unit.ForceVec = types.Ptf(3, 4)
	unit.Pos24 = types.Ptf(5, 6)
	update := unit.UpdateDataMonster()
	update.AIStackInd = 0
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_GUARD)}
	update.Field137 = 1
	sounds := [2]uint32{0x10203040, 0x50607080}
	update.SoundSet122 = unsafe.Pointer(&sounds[0])
	var foundWith *Object
	var audioID uint32
	var audioUnit *Object
	runtime := MonsterMainRuntime547210{
		GUICursorActive: func() bool { return false },
		FindObjectAtCursor: func(got *Object) *Object {
			foundWith = got
			return unit
		},
		AudioEvent: func(id uint32, got *Object) {
			audioID = id
			audioUnit = got
		},
	}

	if !s.MonsterMainNativeRuntime547210(unit, runtime) {
		t.Fatal("Con01A under-cursor conversation was not handled")
	}
	if foundWith != host {
		t.Fatalf("cursor scan source = %p, want host %p", foundWith, host)
	}
	if unit.PrevPos != unit.PosVec || unit.VelVec != (types.Pointf{}) ||
		unit.ForceVec != (types.Pointf{}) || unit.Pos24 != (types.Pointf{}) {
		t.Fatalf("conversation motion state = prev %v vel %v force %v pos24 %v", unit.PrevPos, unit.VelVec, unit.ForceVec, unit.Pos24)
	}
	wantActions := [...]ai.ActionType{
		ai.ACTION_GUARD,
		ai.DEPENDENCY_NOT_MOVED,
		ai.ACTION_WAIT_RELATIVE,
		ai.DEPENDENCY_UNDER_CURSOR,
		ai.ACTION_WAIT_RELATIVE,
		ai.ACTION_FACE_OBJECT,
	}
	if update.AIStackInd != int8(len(wantActions)-1) {
		t.Fatalf("AIStackInd = %d, want %d", update.AIStackInd, len(wantActions)-1)
	}
	for i, want := range wantActions {
		if got := update.AIStack[i].Type(); got != want {
			t.Fatalf("stack[%d] = %v, want %v", i, got, want)
		}
	}
	if got := update.AIStack[2].ArgU32(0); got != s.TickRate() {
		t.Fatalf("first relative wait = %d, want FPS %d", got, s.TickRate())
	}
	if got := update.AIStack[4].ArgU32(0); got != 999999 {
		t.Fatalf("second relative wait = %d, want 999999", got)
	}
	if got := update.AIStack[5].Args[0]; got != uintptr(unsafe.Pointer(host)) {
		t.Fatalf("face target = %#x, want %#x", got, uintptr(unsafe.Pointer(host)))
	}
	if audioID != sounds[1] || audioUnit != unit {
		t.Fatalf("idle audio = %#x/%p, want %#x/%p", audioID, audioUnit, sounds[1], unit)
	}
}

func TestMonsterMainConversation547210RejectsFailedGates(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameModeCoop)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	tests := []struct {
		name  string
		setup func(*Object, *MonsterUpdateData, *Object, *Player, *MonsterMainRuntime547210)
	}{
		{name: "GUI cursor", setup: func(_ *Object, _ *MonsterUpdateData, _ *Object, _ *Player, runtime *MonsterMainRuntime547210) {
			runtime.GUICursorActive = func() bool { return true }
		}},
		{name: "wrong cursor object", setup: func(_ *Object, _ *MonsterUpdateData, _ *Object, _ *Player, runtime *MonsterMainRuntime547210) {
			runtime.FindObjectAtCursor = func(*Object) *Object { return nil }
		}},
		{name: "far cursor", setup: func(_ *Object, _ *MonsterUpdateData, _ *Object, player *Player, _ *MonsterMainRuntime547210) {
			player.CursorVec = image.Pt(120, 100)
		}},
		{name: "timed dependency", setup: func(_ *Object, update *MonsterUpdateData, _ *Object, _ *Player, _ *MonsterMainRuntime547210) {
			update.AIStackInd = 1
			update.AIStack[1].Action = uint32(ai.DEPENDENCY_TIME)
		}},
		{name: "host no update", setup: func(_ *Object, _ *MonsterUpdateData, host *Object, _ *Player, _ *MonsterMainRuntime547210) {
			host.ObjFlags |= object.FlagNoUpdate
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := new(Server)
			player := &Player{CursorVec: image.Pt(100, 100)}
			host := &Object{
				ObjClass:   object.ClassPlayer,
				UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: player}),
			}
			s.Players.SetHost(player, host)
			unit := passiveMonsterTestObject547210(t)
			unit.Field5 = 0x10
			unit.PosVec = types.Ptf(100, 100)
			update := unit.UpdateDataMonster()
			runtime := MonsterMainRuntime547210{
				GUICursorActive:    func() bool { return false },
				FindObjectAtCursor: func(*Object) *Object { return unit },
			}
			tc.setup(unit, update, host, player, &runtime)
			beforeUnit := *unit
			beforeUpdate := *update
			if s.monsterMainConversation547210(unit, update, runtime) {
				t.Fatal("failed conversation gate was handled")
			}
			if *unit != beforeUnit || *update != beforeUpdate {
				t.Fatal("failed conversation gate changed state")
			}
		})
	}
}

func TestMonsterMainConversationImpossible547210(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameModeCoop)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	s := new(Server)
	player := &Player{CursorVec: image.Pt(50, 50)}
	host := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: player})}
	s.Players.SetHost(player, host)
	unit := passiveMonsterTestObject547210(t)
	unit.Field5 = 0x10
	unit.PosVec.X, unit.PosVec.Y = 100, 100
	update := unit.UpdateDataMonster()
	if !s.monsterMainConversationImpossible547210(unit, update) {
		t.Fatal("far cursor should make conversation impossible")
	}
	player.CursorVec = image.Pt(105, 105)
	if s.monsterMainConversationImpossible547210(unit, update) {
		t.Fatal("near cursor may select the NPC")
	}
	update.AIStackInd = 1
	update.AIStack[1].Action = uint32(ai.DEPENDENCY_TIME)
	if !s.monsterMainConversationImpossible547210(unit, update) {
		t.Fatal("wait dependency should suppress conversation")
	}
}

func passiveShopkeeperTestObject547210(t *testing.T) *Object {
	t.Helper()
	unit := monsterActionTestObject50A910(t)
	unit.ObjSubClass = object.SubClass(object.MonsterShopkeeper)
	unit.ObjFlags = object.FlagActive | object.FlagEnabled
	unit.HealthData = &HealthData{}
	update := unit.UpdateDataMonster()
	update.MonsterDef = &MonsterDef{}
	update.Aggression = 0.5
	update.AIStackInd = 0
	update.AIStack[0].Action = uint32(ai.ACTION_IDLE)
	return unit
}

func TestMonsterMainPassiveShopkeeper547210(t *testing.T) {
	unit := passiveShopkeeperTestObject547210(t)
	before := *unit.UpdateDataMonster()
	if !new(Server).MonsterMainPassiveShopkeeper547210(unit) {
		t.Fatal("passive shopkeeper path was not handled")
	}
	after := *unit.UpdateDataMonster()
	if after != before {
		t.Fatal("passive shopkeeper path changed monster state")
	}
}

func TestMonsterMainPassiveShopkeeperGuard547210(t *testing.T) {
	unit := passiveShopkeeperTestObject547210(t)
	unit.UpdateDataMonster().AIStack[0].Action = uint32(ai.ACTION_GUARD)
	before := *unit.UpdateDataMonster()
	if !new(Server).MonsterMainPassiveShopkeeper547210(unit) {
		t.Fatal("passive shopkeeper guard path was not handled")
	}
	after := *unit.UpdateDataMonster()
	if after != before {
		t.Fatal("passive shopkeeper guard path changed monster state")
	}
}

func TestMonsterMainPassiveShopkeeper547210RejectsActiveBranches(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*Object, *MonsterUpdateData)
	}{
		{name: "not shopkeeper", setup: func(unit *Object, _ *MonsterUpdateData) { unit.ObjSubClass = 0 }},
		{name: "destroyed", setup: func(unit *Object, _ *MonsterUpdateData) { unit.ObjFlags |= object.FlagDestroyed }},
		{name: "buffed", setup: func(unit *Object, _ *MonsterUpdateData) { unit.Buffs = 1 }},
		{name: "inventory", setup: func(unit *Object, _ *MonsterUpdateData) { unit.InvFirstItem = unit }},
		{name: "active action", setup: func(_ *Object, update *MonsterUpdateData) { update.AIStack[0].Action = uint32(ai.ACTION_FIGHT) }},
		{name: "stacked", setup: func(_ *Object, update *MonsterUpdateData) { update.AIStackInd = 1 }},
		{name: "enemy", setup: func(unit *Object, update *MonsterUpdateData) { update.CurrentEnemy = unit }},
		{name: "status", setup: func(_ *Object, update *MonsterUpdateData) { update.StatusFlags = object.MonStatusCanBlock }},
		{name: "weapon", setup: func(_ *Object, update *MonsterUpdateData) { update.WeaponEquipFlags = 0x400 }},
		{name: "armor", setup: func(_ *Object, update *MonsterUpdateData) { update.ArmorEquipFlags = 0x1000000 }},
		{name: "dodge definition", setup: func(_ *Object, update *MonsterUpdateData) { update.MonsterDef.StatusFlags92 = 8 }},
		{name: "missing definition", setup: func(_ *Object, update *MonsterUpdateData) { update.MonsterDef = nil }},
		{name: "hungry", setup: func(unit *Object, _ *MonsterUpdateData) { unit.HealthData.Field2 = 1 }},
		{name: "injured", setup: func(unit *Object, _ *MonsterUpdateData) { unit.HealthData = &HealthData{Cur: 4, Max: 5} }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			unit := passiveShopkeeperTestObject547210(t)
			tc.setup(unit, unit.UpdateDataMonster())
			if new(Server).MonsterMainPassiveShopkeeper547210(unit) {
				t.Fatal("active main-AI branch was treated as passive")
			}
		})
	}
}
