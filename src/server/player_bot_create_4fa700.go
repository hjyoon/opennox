package server

import "math"

// PlayerBotCreate4FA700 implements the raw public constructor. Cache the
// player record before allocation and world lookups, which may replace the
// unit's live record. Nil selected records preserve the original faults.
func (unit *Object) PlayerBotCreate4FA700(rt PlayerBotCreateRuntime4FA700) uint32 {
	player := (*PlayerUpdateData)(unit.UpdateData)
	if player.Field73 == nil {
		player.Field73 = rt.Allocate()
	}
	monster := player.Field73
	if monster == nil {
		return 0
	}
	*monster = MonsterUpdateData{}
	monster.Field545 = player
	typ := rt.TypeID("NPC")
	monster.MonsterDef = rt.Definition(typ)
	monster.RetreatLevel = 0.25
	monster.Field335 = 1
	monster.ResumeLevel = math.Float32frombits(1061997773)
	monster.Field337 = 1
	monster.StatusFlags = 186376
	monster.AIStackInd = 0
	monster.AIStack[0].Action = 5
	monster.AIAction340 = 38
	monster.Aggression2 = 0.5
	monster.Aggression = math.Float32frombits(1062501089)
	monster.SightRange = 150
	monster.Field329 = 30
	monster.Field333 = 255
	monster.Field330 = 1
	monster.Field331 = 30
	monster.Field332 = 0.5
	monster.Field338 = 1
	monster.ScriptLookingForEnemy.Func = -1
	monster.ScriptEnemySighted.Func = -1
	monster.ScriptChangeFocus.Func = -1
	monster.ScriptIsHit.Func = -1
	monster.ScriptRetreat.Func = -1
	monster.ScriptDeath.Func = -1
	monster.ScriptCollision.Func = -1
	monster.ScriptHearEnemy.Func = -1
	monster.ScriptEndOfWaypoint.Func = -1
	monster.ScriptLostEnemy.Func = -1
	monster.Field510 = 3
	monster.DialogStartFunc = -1
	monster.DialogEndFunc = -1
	monster.Field0 = 0xdeadface
	monster.Direction94 = uint32(int32(int16(unit.Direction1)))
	monster.Pos95 = unit.PosVec
	// Read the raw class only after initialization and lookup. The guarded
	// PlayerClass helpers would conceal the original nil-player fault here.
	class := uint32(player.Player.Info().playerClass)
	if class == 0 {
		monster.FleeRange = 0
		return 0
	}
	if class != 1 && class != 2 {
		return class - 2
	}
	monster.FleeRange = 50
	monster.Field410 = 0x08000000
	monster.StatusFlags |= 0x20
	monster.Field362_0 = 0
	rate := rt.FPS()
	if class == 2 {
		monster.Field430 = 0x10000000
	}
	monster.Field362_2 = uint16(rate >> 1)
	monster.Field364_0 = 0
	if class == 1 {
		monster.Field423 = 0x10000000
		monster.Field408 = 0x10000000
		monster.Field411 = 0x10000000
		monster.Field364_2 = uint16(rt.FPS())
		monster.Field384 = 0x20000000
		monster.Field405 = 0x20000000
	} else {
		monster.Field364_2 = 6 * uint16(rt.FPS())
		monster.Field432 = 0x20000000
		monster.Field446 = 0x20000000
	}
	monster.Field366_0 = 3 * uint16(rt.FPS())
	monster.Field366_2 = 30 * uint16(rt.FPS())
	if class == 1 {
		switch chance := rt.Random(0, 100); {
		case chance < 33:
			monster.Field399 = 0x40000000
		case chance < 66:
			monster.Field388 = 0x40000000
		default:
			monster.Field415 = 0x40000000
			monster.Field422 = 0x40000000
		}
	} else {
		monster.Field401 = 0x40000000
		monster.Field424 = 0x40000000
		monster.Field456 = 0x40000000
		monster.Field455 = 0x40000000
		monster.Field464 = 0x40000000
	}
	monster.Field368_0 = 0
	rate = rt.FPS()
	monster.Field376 = 0x80000000
	monster.Field370_0 = 0
	monster.Field368_2 = 2 * uint16(rate)
	result := rt.FPS()
	if class == 1 {
		result *= 6
	}
	monster.Field370_2 = 6 * uint16(rt.FPS())
	return result
}
