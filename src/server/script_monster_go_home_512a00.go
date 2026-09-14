package server

import (
	"math"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// ScriptMonsterGoHome512A00 replaces the PE32 nox_server_gotoHome. The script
// queues a report, faces a point ten units ahead of the monster's saved home
// direction, then moves to its saved home position.
func (s *Server) ScriptMonsterGoHome512A00(unit *Object) {
	if unit == nil || !unit.Class().Has(object.ClassMonster) || unit.Flags().Has(object.FlagDead) {
		return
	}
	update := unit.UpdateDataMonster()
	unit.ClearActionStack()
	if report := unit.MonsterPushAction(ai.ACTION_REPORT); report != nil {
		report.Args[0] = uintptr(ai.ACTION_MOVE_TO_HOME)
	}
	if face := unit.MonsterPushAction(ai.ACTION_FACE_LOCATION); face != nil {
		cos, sin := SinCosDir(byte(update.Direction94))
		face.Args[0] = uintptr(math.Float32bits(float32(float64(cos)*10.0 + float64(unit.PosVec.X))))
		face.Args[1] = uintptr(math.Float32bits(float32(float64(sin)*10.0 + float64(unit.PosVec.Y))))
	}
	if home := unit.MonsterPushAction(ai.ACTION_MOVE_TO_HOME); home != nil {
		home.Args[0] = uintptr(math.Float32bits(update.Pos95.X))
		home.Args[1] = uintptr(math.Float32bits(update.Pos95.Y))
		home.Args[2] = 0
	}
}
