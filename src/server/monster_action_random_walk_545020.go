package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

type monsterActionRandomWalkHooks545020 struct {
	random    func(int, int) int
	tileAt    func(types.Pointf) int
	moveAudio func(*Object)
}

func monsterRandomWalkDirection545090(unit *Object, hooks monsterActionRandomWalkHooks545020) byte {
	direction := byte(int(unit.Direction1) + hooks.random(-20, 20))
	if uint32(unit.ObjSubClass)&0x400 == 0 && hooks.tileAt != nil {
		cosine, sine := SinCosDir(direction)
		probe := types.Ptf(unit.PosVec.X+30*cosine, unit.PosVec.Y+30*sine)
		if hooks.tileAt(probe) == 6 {
			direction += 64
		}
	}
	return direction
}

// monsterActionRandomWalk545020 restores GAME.EXE 00545020 and its direction
// helper at 00545090 without reading Object or MonsterUpdateData through PE32
// offsets. Float64 intermediates preserve the original x87 multiply before
// the force components are stored back as float32.
func monsterActionRandomWalk545020(unit *Object, hooks monsterActionRandomWalkHooks545020) bool {
	if unit == nil || unit.UpdateData == nil || hooks.random == nil ||
		!unit.ObjClass.Has(object.ClassMonster) {
		return false
	}
	update := unit.UpdateDataMonster()
	if update.AIStackHead().Type() != ai.ACTION_RANDOM_WALK {
		return false
	}
	direction := monsterRandomWalkDirection545090(unit, hooks)
	unit.Direction1 = Dir16(direction)
	unit.Direction2 = Dir16(direction)
	speed := float64(unit.SpeedCur)
	if update.StatusFlags.Has(object.MonStatusRunning) {
		if update.MonsterDef == nil {
			return false
		}
		speed *= float64(update.MonsterDef.RunMultiplier96)
	}
	cosine, sine := SinCosDir(direction)
	unit.ForceVec.X = float32(speed * float64(cosine))
	unit.ForceVec.Y = float32(speed * float64(sine))
	if hooks.moveAudio != nil {
		hooks.moveAudio(unit)
	}
	return true
}

// MonsterActionRandomWalk545020 binds the native-width action to the live
// logic RNG, tile query, and already-restored movement audio path.
func (s *Server) MonsterActionRandomWalk545020(unit *Object, tileAt func(types.Pointf) int) bool {
	return monsterActionRandomWalk545020(unit, monsterActionRandomWalkHooks545020{
		random:    s.Rand.Logic.IntClamp,
		tileAt:    tileAt,
		moveAudio: s.monsterMoveAudio534030,
	})
}
