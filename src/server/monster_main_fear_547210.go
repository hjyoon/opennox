package server

import (
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// monsterMainFear547210 restores GAME.EXE 0054744D..005474CE after the
// caller's confusion and inversion probes. Fear has no enable, aggression,
// enemy, cast-head, or other-enchant gate. The original movement predicate
// rejects unordered speeds as well as values below binary32 0.01.
//
// The action check reloads the live update-data pointer. Position DWORDs are
// read only after the FLEE push, which may cancel the preceding action. The
// sound-set pointer was cached before the prefix callbacks, but entry +48
// is read after both push attempts. Rejected pushes still sound and return.
func monsterMainFear547210(unit *Object, soundSet unsafe.Pointer, runtime MonsterMainRuntime547210) bool {
	if !unit.HasEnchant(ENCHANT_AFRAID) || !(unit.SpeedBase >= float32(0.0099999998)) ||
		unit.UpdateDataMonster().HasAction(ai.ACTION_FLEE) {
		return false
	}
	unit.MonsterPushAction(ai.DEPENDENCY_IS_ENCHANTED, uint32(ENCHANT_AFRAID))
	if action := unit.MonsterPushAction(ai.ACTION_FLEE); action != nil {
		action.Args[0] = uintptr(math.Float32bits(unit.PosVec.X))
		action.Args[1] = uintptr(math.Float32bits(unit.PosVec.Y))
		action.Args[2] = 0
	}
	if soundSet != nil && runtime.AudioEvent != nil {
		runtime.AudioEvent(*(*uint32)(unsafe.Add(soundSet, 48)), unit)
	}
	return true
}
