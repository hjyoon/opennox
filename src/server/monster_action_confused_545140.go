package server

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

const monsterConfusedAttackDistance545140 = float32(10)

type monsterActionConfusedHooks545140 struct {
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

	// The original x87 code rounds once when each coordinate is stored. It
	// reloads Direction1 for Y, clears Arg2 before storing Y, and leaves Arg3
	// untouched. Using float64 intermediates reproduces the extended multiply
	// and add for binary32 inputs before the final binary32 rounding.
	cosine, _ := SinCosDir(byte(unit.Direction1))
	x := float32(float64(cosine)*float64(monsterConfusedAttackDistance545140) + float64(unit.PosVec.X))
	item.Args[0] = uintptr(math.Float32bits(x))
	_, sine := SinCosDir(byte(unit.Direction1))
	y := float32(float64(sine)*float64(monsterConfusedAttackDistance545140) + float64(unit.PosVec.Y))
	item.Args[2] = 0
	item.Args[1] = uintptr(math.Float32bits(y))
	return true
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
