package server

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

const monsterConfusedAttackDistance545140 = float32(10)

type monsterActionConfusedHooks545140 struct {
	Nearest    bool
	random     func(int, int) int
	randomWalk func(*Object) bool
	canMelee   func(*Object) bool
	canShoot   func(*Object) bool
	push       func(*Object, ai.ActionType) *AIStackItem
}

// monsterActionConfused545140 restores GAME.EXE 00545140. The function is
// reachable only through the ACTION_CONFUSED dispatch row, so requiring that
// action at the native stack head rejects accidental cross-action calls.
// Engine calls remain hooks to preserve and test the original RNG, capability,
// and push order without passing Object through the PE32 int callback ABI.
func monsterActionConfused545140(unit *Object, hooks monsterActionConfusedHooks545140) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) ||
		hooks.random == nil || hooks.randomWalk == nil || hooks.canMelee == nil ||
		hooks.canShoot == nil || hooks.push == nil {
		return false
	}
	if unit.UpdateDataMonster().AIStackHead().Type() != ai.ACTION_CONFUSED {
		return false
	}

	if hooks.random(0, 100) >= 15 {
		hooks.randomWalk(unit)
		return true
	}
	if hooks.canMelee(unit) {
		if !hooks.canShoot(unit) || hooks.random(0, 100) < 50 {
			hooks.push(unit, ai.ACTION_MELEE_ATTACK)
			return true
		}
	} else if !hooks.canShoot(unit) {
		return true
	}

	item := hooks.push(unit, ai.ACTION_MISSILE_ATTACK)
	if item == nil {
		return true
	}

	// Preserve the signed WORD lookup, reload for Y, and original store order.
	// Each multiply/add uses PC53, then each coordinate spills to binary32.
	cosine := monsterConfusedDirectionFloat545140(int32(int16(unit.Direction1)) * 2)
	xWide := monsterRandomWalkMul53_545020(float64(cosine), float64(monsterConfusedAttackDistance545140), hooks.Nearest)
	xWide = monsterRandomWalkAdd53_545020(xWide, objectForceMemory32_52DF80(unit.PosVec.X), hooks.Nearest)
	x := monsterRandomWalkSpill545020(xWide, hooks.Nearest)
	item.Args[0] = uintptr(math.Float32bits(x))
	sine := monsterConfusedDirectionFloat545140(int32(int16(unit.Direction1))*2 + 1)
	yWide := monsterRandomWalkMul53_545020(float64(sine), float64(monsterConfusedAttackDistance545140), hooks.Nearest)
	yWide = monsterRandomWalkAdd53_545020(yWide, objectForceMemory32_52DF80(unit.PosVec.Y), hooks.Nearest)
	item.Args[2] = 0
	item.Args[1] = uintptr(math.Float32bits(monsterRandomWalkSpill545020(yWide, hooks.Nearest)))
	return true
}

func monsterConfusedDirectionFloat545140(index int32) float32 {
	if index >= 0 && index < 512 {
		return sincosDirTable[index/2][index&1]
	}
	// Direction -1 addresses the two original DWORDs immediately preceding
	// the direction table, rather than direction 255. Keep their exact bits
	// available with the native table, including without legacy blob setup.
	if index == -2 {
		return math.Float32frombits(0x00000010)
	}
	if index == -1 {
		return math.Float32frombits(0)
	}
	return *(*float32)(memmap.Ptr(uintptr(int64(0x5b6658) + int64(index)*4)))
}

// MonsterActionConfused545140 binds confused movement to native-width monster
// state, the live logic RNG, and the already-restored random-walk movement.
func (s *Server) MonsterActionConfused545140(unit *Object, tileAt func(types.Pointf) int) bool {
	return monsterActionConfused545140(unit, monsterActionConfusedHooks545140{
		random: s.Rand.Logic.IntClamp,
		randomWalk: func(unit *Object) bool {
			return monsterActionRandomWalkMove545020(unit, s.monsterActionRandomWalkHooks545020(tileAt))
		},
		canMelee: func(unit *Object) bool {
			return monsterFightCanMelee534220(unit, unit.UpdateDataMonster())
		},
		canShoot: func(unit *Object) bool {
			return monsterFightCanShoot534280(unit, unit.UpdateDataMonster())
		},
		push: func(unit *Object, action ai.ActionType) *AIStackItem {
			return unit.MonsterPushAction(action)
		},
	})
}
