package server

const (
	monsterGeneratorMaxActiveHighKey4F0590     = "GeneratorMaxActiveCreaturesHigh"
	monsterGeneratorMaxActiveNormalKey4F0590   = "GeneratorMaxActiveCreaturesNormal"
	monsterGeneratorMaxActiveLowKey4F0590      = "GeneratorMaxActiveCreaturesLow"
	monsterGeneratorMaxActiveSingularKey4F0590 = "GeneratorMaxActiveCreaturesSingular"
)

var monsterGeneratorMaxActiveKeys4F0590 = [...]string{
	monsterGeneratorMaxActiveHighKey4F0590,
	monsterGeneratorMaxActiveNormalKey4F0590,
	monsterGeneratorMaxActiveLowKey4F0590,
	monsterGeneratorMaxActiveSingularKey4F0590,
}

type directionIndexToAngleHooks509E90 struct {
	loadTable func(uint32) uint32
}

// directionIndexToAngle509E90 preserves the complete dword loaded by GAME.EXE
// helper 00509E90. The original helper performs no bounds check.
func directionIndexToAngle509E90(index uint32, hooks directionIndexToAngleHooks509E90) uint32 {
	return hooks.loadTable(index)
}

type monsterGeneratorInitHooks4F0590[O, U any] struct {
	loadUpdateData      func(O) U
	currentQuestGroup   func() uint32
	loadQuestSpawnRate  func(U, uint32) uint8
	loadBalanceFloat    func(string) float64
	truncQwordLow       func(float64) int32
	storeMaxActive      func(U, uint8)
	loadObjSubClass     func(O) uint32
	directionIndexAngle func(uint32) uint32
	storeDirection1     func(O, uint16)
	loadDirection1      func(O) uint16
	storeDirection2     func(O, uint16)
}

// monsterGeneratorInit4F0590 preserves GAME.EXE 004F0590's observable order.
// UpdateData is cached before the quest-group callback. A selector from zero
// through three chooses the exact balance key, whose original binary32 result
// is supplied widened to float64, truncates it through helper 00566DCC, and
// stores only the low byte. Other selectors skip that entire path.
//
// The full ObjSubClass dword is loaded afterward. Only its low byte is tested,
// with bits 1, 2, 4, and 8 taking priority and mapping to direction-table
// indices 0, 2, 8, and 6. A match stores the helper's low word to Direction1;
// all paths then reload Direction1 and copy it to Direction2. The returned
// dword is the full direction-table value on a match and full ObjSubClass when
// no bit matches. The original has no nil or bounds guards.
func monsterGeneratorInit4F0590[O, U any](unit O, hooks monsterGeneratorInitHooks4F0590[O, U]) int32 {
	update := hooks.loadUpdateData(unit)
	group := hooks.currentQuestGroup()
	selector := hooks.loadQuestSpawnRate(update, group)
	if selector <= 3 {
		value := hooks.loadBalanceFloat(monsterGeneratorMaxActiveKeys4F0590[selector])
		maximum := hooks.truncQwordLow(value)
		hooks.storeMaxActive(update, uint8(maximum))
	}

	subClass := hooks.loadObjSubClass(unit)
	result := subClass
	index := uint32(0)
	matched := true
	switch {
	case subClass&1 != 0:
		index = 0
	case subClass&2 != 0:
		index = 2
	case subClass&4 != 0:
		index = 8
	case subClass&8 != 0:
		index = 6
	default:
		matched = false
	}
	if matched {
		result = hooks.directionIndexAngle(index)
		hooks.storeDirection1(unit, uint16(result))
	}
	hooks.storeDirection2(unit, hooks.loadDirection1(unit))
	return int32(result)
}
