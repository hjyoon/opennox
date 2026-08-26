package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// MonsterActionRunStart534750 and MonsterActionRunEnd534780 restore the two
// status-bit helpers used by FLEE and return-to-home actions.
func (s *Server) MonsterActionRunStart534750(unit *Object) {
	if unit == nil || unit.UpdateData == nil {
		return
	}
	update := unit.UpdateDataMonster()
	if !update.StatusFlags.Has(object.MonStatusNeverRun) {
		update.StatusFlags |= object.MonStatusRunning
	}
}

func (s *Server) MonsterActionRunEnd534780(unit *Object) {
	if unit == nil || unit.UpdateData == nil {
		return
	}
	update := unit.UpdateDataMonster()
	if !update.StatusFlags.Has(object.MonStatusAlwaysRun) {
		update.StatusFlags &^= object.MonStatusRunning
	}
}

type monsterActionFleeHooks544760 struct {
	frame        func() uint32
	tickRate     func() uint32
	generatePath func([]types.Pointf, *Object, *types.Pointf) int
	actuallyMove func(*Object) bool
	moveAudio    func(*Object)
	pop          func() int
}

// monsterActionFlee544760 restores the native-width movement portion of
// GAME.EXE 00544760. The spell side branch is inactive for aggression below
// 0.08, which is the War01A retreat path that first reaches this routine.
func monsterActionFlee544760(unit *Object, hooks monsterActionFleeHooks544760) {
	if unit == nil || unit.UpdateData == nil || hooks.pop == nil {
		return
	}
	update := unit.UpdateDataMonster()
	head := update.AIStackHead()
	if head == nil || head.Type() != ai.ACTION_FLEE {
		return
	}
	if unit.SpeedBase < 0.0099999998 {
		hooks.pop()
		return
	}
	if enemy := update.CurrentEnemy; enemy != nil {
		head.SetArgs(enemy.PosVec, uint32(0))
		delta := enemy.PosVec.Sub(unit.PosVec)
		if update.FleeRange*update.FleeRange > delta.X*delta.X+delta.Y*delta.Y &&
			hooks.frame()-update.Field70 > hooks.tickRate()/2 {
			update.Field2 = 0
		}
	}
	frame := hooks.frame()
	if update.Field2 != 0 && frame-update.Field70 > 2*hooks.tickRate() {
		update.Field2 = 0
		frame = hooks.frame()
	}
	move := update.Field2 != 0 || frame-update.Field70 <= 10
	if !move && hooks.generatePath != nil {
		count := hooks.generatePath(update.Path[:], unit, &types.Pointf{X: head.ArgPos(0).X, Y: head.ArgPos(0).Y})
		update.Field2 = uint32(count)
		update.Field70 = hooks.frame()
		update.Field67 = 0
		move = count > 1
		if !move {
			head.SetArgs(unit.PosVec, uint32(0))
		}
	}
	if move {
		if hooks.actuallyMove != nil && hooks.actuallyMove(unit) {
			update.Field2 = 0
		}
		if hooks.moveAudio != nil {
			hooks.moveAudio(unit)
		}
	}
}

// MonsterActionFlee544760 binds the FLEE action to the live retreat
// pathfinder and native movement implementation.
func (s *Server) MonsterActionFlee544760(unit *Object, generatePath func([]types.Pointf, *Object, *types.Pointf) int) {
	monsterActionFlee544760(unit, monsterActionFleeHooks544760{
		frame:        s.Frame,
		tickRate:     s.TickRate,
		generatePath: generatePath,
		actuallyMove: func(unit *Object) bool { return monsterCreatureActuallyMove50D3B0(unit, s.MapTraceRay) },
		moveAudio:    s.monsterMoveAudio534030,
		pop:          unit.MonsterPopAction,
	})
}
