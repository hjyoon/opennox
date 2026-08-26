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
		{name: "inventory", setup: func(unit *Object, _ *MonsterUpdateData) { unit.InvFirstItem = unit }},
		{name: "enemy", setup: func(unit *Object, update *MonsterUpdateData) { update.CurrentEnemy = unit }},
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
