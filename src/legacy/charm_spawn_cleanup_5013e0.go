package legacy

import "github.com/opennox/opennox/v1/server"

func charmQuestSpawnCleanupNative5013E0(s *server.Server, target *server.Object) {
	if s != nil {
		s.MonsterSpawnCleanup50E140(target)
	}
}

func charmQuestSpawnCleanupRuntime5013E0(target *server.Object) {
	charmQuestSpawnCleanupNative5013E0(GetServer().S(), target)
}
