package legacy

import "github.com/opennox/opennox/v1/server"

func unitDamageClearRuntime4EE5E0() server.UnitDamageClearRuntime4EE5E0 {
	srv := GetServer()
	return server.UnitDamageClearRuntime4EE5E0{
		BreakHarpoon:  Nox_xxx_harpoonBreakForPlr_537520,
		SetHP:         Nox_xxx_unitSetHP_4E4560,
		BuffOff:       Nox_xxx_spellBuffOff_4FF5B0,
		SoloReward:    soloMonsterKillRewardCall4EE500,
		MonsterDie:    unitDamageClearMonsterDie4EE5E0,
		DelayedDelete: srv.DelayedDelete,
	}
}

func unitDamageClearCall4EE5E0(unit *server.Object, damage int32) {
	GetServer().S().UnitDamageClear4EE5E0(unit, damage, unitDamageClearRuntime4EE5E0())
}
