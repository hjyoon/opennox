package legacy

import "github.com/opennox/opennox/v1/server"

func monsterInitRuntime4F0040() server.MonsterInitRuntime4F0040 {
	return server.MonsterInitRuntime4F0040{
		SetHealth: Nox_xxx_unitSetHP_4E4560,
	}
}

func monsterInitCall4F0040(unit *server.Object) {
	unit.Server().MonsterInit4F0040(unit, monsterInitRuntime4F0040())
}
