package server

import "github.com/opennox/libs/object"

// PlayerHandleElimDeath54D7A0 follows the executed GAME.EXE 0054D7A0 root.
// The shared runtime declares its team/score/report/count/statistics services.
// This dependency alone does not admit the still-unported PlayerDie branches.
func PlayerHandleElimDeath54D7A0(victim, killer *Object, rt PlayerUpdateScoreRuntime54D980) {
	victimUpdate := (*PlayerUpdateData)(victim.UpdateData)
	var victimTeam, killerTeam *Team
	var killerUpdate *PlayerUpdateData
	if rt.HasTeam(&victim.TeamVal) {
		victimTeam = rt.TeamByID(uint8(victim.TeamVal.ID))
	}
	if killer != nil {
		killerUpdate = (*PlayerUpdateData)(killer.UpdateData)
		if rt.HasTeam(&killer.TeamVal) {
			killerTeam = rt.TeamByID(uint8(killer.TeamVal.ID))
		}
	}

	if killer == victim {
		rt.SubtractScore(victim, 1)
		rt.IncrementElimDeath(victim)
		rt.ReportLesson(victim)
		if victimTeam != nil {
			rt.TeamChangeLessons(victimTeam, int32(uint32(victimTeam.Lessons)+1))
		}
		if rt.ObserverMode() != 0 && killerUpdate != nil {
			player := killerUpdate.Player
			rt.ObserverUpdate(player, player)
		}
		return
	}

	if killer != nil {
		if killer.ObjClass.Has(object.ClassPlayer) {
			if killerTeam != nil && killerTeam == victimTeam {
				rt.SubtractScore(killer, 1)
				rt.ReportLesson(killer)
				if rt.ObserverMode() != 0 && killerUpdate != nil {
					player := killerUpdate.Player
					rt.ObserverUpdate(player, player)
				}
			} else {
				rt.AddScore(killer, 1)
				rt.ReportLesson(killer)
				if rt.ObserverMode() != 0 && killerUpdate != nil && victimUpdate != nil {
					// Both 0054D8B7 and 0054D8F1 read victim's Player first.
					second := victimUpdate.Player
					first := killerUpdate.Player
					rt.ObserverUpdate(first, second)
				}
			}
		}
	} else if rt.ObserverMode() != 0 && victimUpdate != nil {
		player := victimUpdate.Player
		rt.ObserverUpdate(player, player)
	}

	rt.IncrementElimDeath(victim)
	rt.ReportLesson(victim)
	if victimTeam != nil {
		rt.TeamChangeLessons(victimTeam, int32(uint32(victimTeam.Lessons)+1))
	}
}
