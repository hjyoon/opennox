package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

type monsterActionRandomWalkHooks545020 struct {
	Nearest   bool
	random    func(int, int) int
	tileAt    func(types.Pointf) int
	moveAudio func(*Object)
}

func monsterRandomWalkDirection545090(unit *Object, hooks monsterActionRandomWalkHooks545020) byte {
	offset := hooks.random(-20, 20)
	direction := byte(int(int16(unit.Direction1)) + offset)
	if uint32(unit.ObjSubClass)&0x400 == 0 && hooks.tileAt != nil {
		cosine, sine := SinCosDir(direction)
		x := monsterRandomWalkMul53_545020(float64(cosine), 30, hooks.Nearest)
		x = monsterRandomWalkAdd53_545020(x, objectForceMemory32_52DF80(unit.PosVec.X), hooks.Nearest)
		y := monsterRandomWalkMul53_545020(float64(sine), 30, hooks.Nearest)
		y = monsterRandomWalkAdd53_545020(y, objectForceMemory32_52DF80(unit.PosVec.Y), hooks.Nearest)
		probe := types.Ptf(monsterRandomWalkSpill545020(x, hooks.Nearest), monsterRandomWalkSpill545020(y, hooks.Nearest))
		if hooks.tileAt(probe) == 6 {
			direction += 64
		}
	}
	return direction
}

// monsterActionRandomWalk545020 restores GAME.EXE 00545020 and its direction
// helper at 00545090 without reading Object or MonsterUpdateData through PE32
// offsets. UpdateData is cached before the direction helper; the helper may
// change the object's live class or update pointer through its services.
func monsterActionRandomWalkMove545020(unit *Object, hooks monsterActionRandomWalkHooks545020) bool {
	if unit == nil || unit.UpdateData == nil || hooks.random == nil {
		return false
	}
	update := (*MonsterUpdateData)(unit.UpdateData)
	direction := monsterRandomWalkDirection545090(unit, hooks)
	unit.Direction1 = Dir16(direction)
	unit.Direction2 = Dir16(direction)
	var speed float64
	if update.StatusFlags.Has(object.MonStatusRunning) {
		if update.MonsterDef == nil {
			return false
		}
		speed = monsterRandomWalkMul53_545020(float64(update.MonsterDef.RunMultiplier96), objectForceMemory32_52DF80(unit.SpeedCur), hooks.Nearest)
	} else {
		speed = float64(unit.SpeedCur)
	}
	cosine, sine := SinCosDir(direction)
	unit.ForceVec.X = monsterRandomWalkSpill545020(monsterRandomWalkMul53_545020(speed, objectForceMemory32_52DF80(cosine), hooks.Nearest), hooks.Nearest)
	unit.ForceVec.Y = monsterRandomWalkSpill545020(monsterRandomWalkMul53_545020(speed, objectForceMemory32_52DF80(sine), hooks.Nearest), hooks.Nearest)
	if hooks.moveAudio != nil {
		hooks.moveAudio(unit)
	}
	return true
}

func monsterActionRandomWalk545020(unit *Object, hooks monsterActionRandomWalkHooks545020) bool {
	if unit == nil || unit.UpdateData == nil || !unit.ObjClass.Has(object.ClassMonster) ||
		unit.UpdateDataMonster().AIStackHead().Type() != ai.ACTION_RANDOM_WALK {
		return false
	}
	return monsterActionRandomWalkMove545020(unit, hooks)
}

func (s *Server) monsterActionRandomWalkHooks545020(tileAt func(types.Pointf) int) monsterActionRandomWalkHooks545020 {
	return monsterActionRandomWalkHooks545020{
		random:    s.Rand.Logic.IntClamp,
		tileAt:    tileAt,
		moveAudio: s.monsterMoveAudio534030,
	}
}

// MonsterActionRandomWalk545020 binds the native-width action to the live
// logic RNG, tile query, and already-restored movement audio path.
func (s *Server) MonsterActionRandomWalk545020(unit *Object, tileAt func(types.Pointf) int) bool {
	return monsterActionRandomWalk545020(unit, s.monsterActionRandomWalkHooks545020(tileAt))
}

// MonsterRandomWalkMoveNative545020 is the original void update body. Confused
// also calls this body while ACTION_CONFUSED remains at the stack head.
func (s *Server) MonsterRandomWalkMoveNative545020(unit *Object, tileAt func(types.Pointf) int) {
	monsterActionRandomWalkMove545020(unit, s.monsterActionRandomWalkHooks545020(tileAt))
}

// MonsterRandomWalkDirectionNative545090 retains the helper's public int.
func (s *Server) MonsterRandomWalkDirectionNative545090(unit *Object, tileAt func(types.Pointf) int) int {
	return int(monsterRandomWalkDirection545090(unit, s.monsterActionRandomWalkHooks545020(tileAt)))
}
