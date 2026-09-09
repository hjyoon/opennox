package server

import (
	"bytes"
	"math"
	"strings"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

const monsterEscortNameSize546600 = 19 * 4

type monsterEscortNameHooks546600 struct {
	firstPlayer func() *Object
	nextPlayer  func(*Object) *Object
	random      func(int, int) int
	owner       func() *Object
	objectByID  func(string) *Object
}

func monsterEscortNameBuffer546600(update *MonsterUpdateData) []byte {
	if update == nil {
		return nil
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(&update.Field341)), monsterEscortNameSize546600)
}

func monsterEscortScriptNameMatches546600(id, name string) bool {
	if id == "" {
		return false
	}
	if strings.Contains(name, ":") {
		return id == name
	}
	if separator := strings.IndexByte(id, ':'); separator >= 0 {
		id = id[separator+1:]
	}
	return id == name
}

func monsterEscortFindByScriptName546600(root *Object, name string) *Object {
	if monsterEscortScriptNameMatches546600(root.ID(), name) {
		return root
	}
	for item := root.InvFirstItem; item != nil; item = item.InvNextItem {
		if monsterEscortScriptNameMatches546600(item.ID(), name) {
			return item
		}
	}
	return nil
}

func (s *Server) monsterEscortObjectByScriptName546600(name string) *Object {
	for obj := s.Objs.List; obj != nil; obj = obj.ObjNext {
		if found := monsterEscortFindByScriptName546600(obj, name); found != nil {
			return found
		}
	}
	for obj := s.Objs.Pending; obj != nil; obj = obj.ObjNext {
		if found := monsterEscortFindByScriptName546600(obj, name); found != nil {
			return found
		}
	}
	return nil
}

// monsterGetObjEscortName546600 restores GAME.EXE 00546600 while keeping the
// 76-byte PE32 name field in MonsterUpdateData and all resolved object
// pointers in their native-width Go representation. As in the original, only
// the first name byte is cleared after every lookup attempt.
func monsterGetObjEscortName546600(unit *Object, hooks monsterEscortNameHooks546600) *Object {
	if unit == nil || unit.UpdateData == nil || !unit.ObjClass.Has(object.ClassMonster) {
		return nil
	}
	nameBuf := monsterEscortNameBuffer546600(unit.UpdateDataMonster())
	if len(nameBuf) == 0 {
		return nil
	}
	defer func() { nameBuf[0] = 0 }()

	end := bytes.IndexByte(nameBuf, 0)
	if end < 0 {
		end = len(nameBuf)
	}
	name := string(nameBuf[:end])
	switch name {
	case "**PLAYER**":
		count := 0
		for player := hooks.firstPlayer(); player != nil; player = hooks.nextPlayer(player) {
			count++
		}
		selected := hooks.random(0, count-1)
		result := hooks.firstPlayer()
		for result != nil {
			index := selected
			selected--
			if index == 0 {
				break
			}
			result = hooks.nextPlayer(result)
		}
		return result
	case "**OWNER**":
		return hooks.owner()
	default:
		return hooks.objectByID(name)
	}
}

type monsterActionEscortHooks546430 struct {
	resolveTarget    func() *Object
	canAttackAtWill  func() bool
	mediumAggression func() bool
	noticeThreat     func() int
	lookAtDamager    func() bool
	interestingSound func() int
	hasAntiMagic     func() bool
	healSomeone      func() int
	frame            func() uint32
	push             func(ai.ActionType) *AIStackItem
	pop              func() int
}

func monsterEscortSetPosition546430(item *AIStackItem, target *Object) {
	if item == nil || target == nil {
		return
	}
	item.Args[0] = uintptr(math.Float32bits(target.PosVec.X))
	item.Args[1] = uintptr(math.Float32bits(target.PosVec.Y))
}

func monsterEscortSetTarget546430(item *AIStackItem, target *Object) {
	if item == nil {
		return
	}
	item.Args[2] = uintptr(unsafe.Pointer(target))
}

// monsterActionEscort546430 restores GAME.EXE 00546430 without passing an
// Object, MonsterUpdateData, or AI-stack pointer through the PE32 C ABI. The
// original reloads mutable enemy/escort state after several action pushes;
// those reload points are intentionally retained below.
func monsterActionEscort546430(unit *Object, hooks monsterActionEscortHooks546430) bool {
	if unit == nil || unit.UpdateData == nil || !unit.ObjClass.Has(object.ClassMonster) ||
		hooks.resolveTarget == nil || hooks.push == nil || hooks.pop == nil {
		return false
	}
	update := unit.UpdateDataMonster()
	if update.AIStackInd < 0 || int(update.AIStackInd) >= len(update.AIStack) {
		return false
	}
	escort := update.AIStackHead()
	if escort == nil || escort.Type() != ai.ACTION_ESCORT {
		return false
	}

	target := escort.ArgObj(2)
	if target == nil {
		target = hooks.resolveTarget()
		monsterEscortSetTarget546430(escort, target)
		if target == nil {
			hooks.pop()
			return true
		}
		monsterEscortSetPosition546430(escort, target)
	}

	if hooks.canAttackAtWill() {
		if update.CurrentEnemy != nil {
			fight := hooks.push(ai.ACTION_FIGHT)
			if fight != nil {
				// The push may cancel the previous action and run callbacks, so
				// GAME.EXE reloads CurrentEnemy before reading its live position.
				enemy := update.CurrentEnemy
				monsterEscortSetPosition546430(fight, enemy)
				fight.Args[2] = uintptr(hooks.frame())
			}
			return true
		}
		if hooks.lookAtDamager() {
			return true
		}
	} else if hooks.mediumAggression() {
		if hooks.noticeThreat() != 0 {
			return true
		}
		if hooks.lookAtDamager() {
			return true
		}
	}

	target = escort.ArgObj(2)
	if target == nil {
		return true
	}
	dx := float64(unit.PosVec.X) - float64(target.PosVec.X)
	dy := float64(unit.PosVec.Y) - float64(target.PosVec.Y)
	radius := float64(update.Field329) + 30.0
	if radius*radius >= dx*dx+dy*dy {
		if !hooks.canAttackAtWill() || hooks.interestingSound() == 0 {
			if !hooks.hasAntiMagic() {
				hooks.healSomeone()
			}
		}
		return true
	}

	if hooks.mediumAggression() || hooks.canAttackAtWill() {
		hooks.push(ai.DEPENDENCY_NOT_UNDER_ATTACK)
	}
	if hooks.canAttackAtWill() {
		hooks.push(ai.DEPENDENCY_NO_VISIBLE_ENEMY)
	}
	if farther := hooks.push(ai.DEPENDENCY_OBJECT_FARTHER_THAN); farther != nil {
		farther.Args[0] = uintptr(math.Float32bits(update.Field329))
		monsterEscortSetTarget546430(farther, escort.ArgObj(2))
	}
	if move := hooks.push(ai.ACTION_MOVE_TO); move != nil {
		// Both the dependency and move pushes may run callbacks. Reload the
		// escort target at the same points as the original routine.
		target = escort.ArgObj(2)
		monsterEscortSetPosition546430(move, target)
		monsterEscortSetTarget546430(move, escort.ArgObj(2))
	}
	return true
}

// MonsterGetObjEscortName546600 binds the original escort-name resolver to
// the live player list, deterministic logic RNG, owner, and script ID lookup.
func (s *Server) MonsterGetObjEscortName546600(unit *Object) *Object {
	return monsterGetObjEscortName546600(unit, monsterEscortNameHooks546600{
		firstPlayer: s.Players.FirstUnit,
		nextPlayer:  s.Players.NextUnit,
		random:      s.Rand.Logic.IntClamp,
		owner:       func() *Object { return unit.ObjOwner },
		objectByID:  s.monsterEscortObjectByScriptName546600,
	})
}

// MonsterActionEscort546430 binds the restored action to the live server.
// interestingSound remains supplied by the top-level package because its
// compatibility path and tile callbacks live there.
func (s *Server) MonsterActionEscort546430(unit *Object, interestingSound func(*Object) int) bool {
	return monsterActionEscort546430(unit, monsterActionEscortHooks546430{
		resolveTarget:    func() *Object { return s.MonsterGetObjEscortName546600(unit) },
		canAttackAtWill:  unit.Nox_xxx_monsterCanAttackAtWill_534390,
		mediumAggression: unit.Sub_5343C0,
		noticeThreat:     unit.Sub_545E60,
		lookAtDamager:    unit.MonsterLookAtDamager,
		interestingSound: func() int { return interestingSound(unit) },
		hasAntiMagic:     func() bool { return unit.HasEnchant(ENCHANT_ANTI_MAGIC) },
		healSomeone:      func() int { return s.MonsterHealSomeone5411A0(unit) },
		frame:            s.Frame,
		push:             func(action ai.ActionType) *AIStackItem { return unit.MonsterPushAction(action) },
		pop:              unit.MonsterPopAction,
	})
}
