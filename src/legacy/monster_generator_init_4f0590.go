package legacy

import "github.com/opennox/opennox/v1/server"

func monsterGeneratorInitCall4F0590(unit *server.Object, currentQuestGroup func() uint32) int32 {
	return GetServer().S().MonsterGeneratorInit4F0590(unit, server.MonsterGeneratorInitRuntime4F0590{
		CurrentQuestGroup: currentQuestGroup,
	})
}
