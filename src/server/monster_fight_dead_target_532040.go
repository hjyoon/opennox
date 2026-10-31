package server

import "github.com/opennox/libs/object"

func monsterFightDeadTargetMatch532040(candidate *Object, candidateCode, targetCode uint32) bool {
	return candidateCode == targetCode && candidate.ObjFlags.Has(object.FlagDead)
}

// MonsterFightDeadTarget532040 reads the explicitly supplied cached record.
// GAME.EXE reads the candidate ID before the context ID, and reads flags only
// after a match. Neither class admission nor Object.UpdateData belongs here.
func MonsterFightDeadTarget532040(candidate *Object, context *MonsterUpdateData) bool {
	candidateCode := candidate.NetCode
	return monsterFightDeadTargetMatch532040(candidate, candidateCode, context.Field300)
}
