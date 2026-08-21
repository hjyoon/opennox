package server

type monsterGeneratorInitNativeDeps4F0590 struct {
	currentQuestGroup func() uint32
	balanceFloat      func(string) float64
}

// MonsterGeneratorInitRuntime4F0590 supplies the Quest generator-group index
// whose storage remains in the legacy runtime.
type MonsterGeneratorInitRuntime4F0590 struct {
	CurrentQuestGroup func() uint32
}

func directionIndexToAngleNative509E90(index uint32) uint32 {
	return directionIndexToAngle509E90(index, directionIndexToAngleHooks509E90{
		loadTable: func(index uint32) uint32 {
			// GAME.EXE performs an unchecked dword load. All three decoded callers
			// use the sealed 0..8 table domain. Reading adjacent PE data for a
			// malformed index is intentionally not recreated.
			if index >= uint32(len(directionToAngleTable509E00)) {
				panic("direction index outside sealed GAME.EXE table")
			}
			return directionToAngleTable509E00[index]
		},
	})
}

func monsterGeneratorInitNative4F0590(unit *Object, deps monsterGeneratorInitNativeDeps4F0590) int32 {
	return monsterGeneratorInit4F0590(unit, monsterGeneratorInitHooks4F0590[*Object, *MonsterGenUpdateData]{
		loadUpdateData: func(unit *Object) *MonsterGenUpdateData {
			// Do not use UpdateDataMonsterGen: 004F0590 has no class or nil gate.
			return (*MonsterGenUpdateData)(unit.UpdateData)
		},
		currentQuestGroup: deps.currentQuestGroup,
		loadQuestSpawnRate: func(update *MonsterGenUpdateData, group uint32) uint8 {
			// The legacy global is 0..2. Preserve a deterministic fault instead
			// of reading adjacent native memory if it is corrupted.
			return update.QuestSpawnRate[group]
		},
		loadBalanceFloat: func(key string) float64 {
			// Original helper 00419D40 executes FLD dword. Balance stores float64
			// today, so reproduce that binary32 boundary before x87 truncation.
			return float64(float32(deps.balanceFloat(key)))
		},
		truncQwordLow: x87TruncSignedQwordLow566DCC,
		storeMaxActive: func(update *MonsterGenUpdateData, value uint8) {
			update.MaxActive = value
		},
		loadObjSubClass: func(unit *Object) uint32 {
			return uint32(unit.ObjSubClass)
		},
		directionIndexAngle: directionIndexToAngleNative509E90,
		storeDirection1: func(unit *Object, value uint16) {
			unit.Direction1 = Dir16(value)
		},
		loadDirection1: func(unit *Object) uint16 {
			return uint16(unit.Direction1)
		},
		storeDirection2: func(unit *Object, value uint16) {
			unit.Direction2 = Dir16(value)
		},
	})
}

// MonsterGeneratorInit4F0590 binds GAME.EXE 004F0590 to native-width Object
// pointers and the portable MonsterGenUpdateData record. There are
// deliberately no nil or class guards.
//
//go:noinline
func (s *Server) MonsterGeneratorInit4F0590(unit *Object, runtime MonsterGeneratorInitRuntime4F0590) int32 {
	return monsterGeneratorInitNative4F0590(unit, monsterGeneratorInitNativeDeps4F0590{
		currentQuestGroup: runtime.CurrentQuestGroup,
		balanceFloat:      s.Balance.Float,
	})
}
