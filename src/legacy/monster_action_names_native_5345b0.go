package legacy

import "github.com/opennox/opennox/v1/common/unit/ai"

func monsterSelectedActionNameNative5345B0(action uint32) string {
	switch ai.ActionType(action) {
	case ai.ACTION_IDLE, ai.ACTION_GUARD, ai.ACTION_ROAM, ai.ACTION_ESCORT:
		return ai.ActionType(action).String()
	default:
		return ai.ACTION_INVALID.String()
	}
}

func monsterActionNameNative534650(action uint32) (string, bool) {
	if action >= 72 {
		return "", false
	}
	// GAME.EXE's slot 39 points at an empty BSS string. String() retains
	// its diagnostic spelling; serialized and public C names use this table.
	if action == 39 {
		return "", true
	}
	return ai.ActionType(action).String(), true
}

func monsterActionByName528DB0(name string, limit int, fallback ai.ActionType) ai.ActionType {
	for i := 0; i < limit; i++ {
		if value, ok := monsterActionNameNative534650(uint32(i)); ok && value == name {
			return ai.ActionType(i)
		}
	}
	return fallback
}
