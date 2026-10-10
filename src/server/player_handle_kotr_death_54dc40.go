package server

import (
	"math"

	"github.com/opennox/libs/object"
)

// The shared score runtime and these callbacks describe the services called
// by GAME.EXE 0054DC40. This dependency does not admit competitive PlayerDie.
type PlayerHandleKotrDeathRuntime54DC40 struct {
	PlayerUpdateScoreRuntime54D980
	IsCrown      func(*Object) bool
	BalanceFloat func(string) float64
	FloatToInt   func(float32) uint32
	GameplayFlag func(uint32) bool
	DropCrowns   func(*Object, *Object)
}

// The original root spills the balance double to binary32 before calling
// 00419A70's FISTP, using the game's round-to-nearest-even x87 control mode.
// Invalid conversions retain the original integer-indefinite DWORD.
func PlayerKotrPoints54DC40(value float32) uint32 {
	if math.IsNaN(float64(value)) || value >= 2147483648 || value < -2147483648 {
		return 0x80000000
	}
	return uint32(int32(math.RoundToEven(float64(value))))
}

// Preserve the NaN payload/sign observed in the original root's FSTPS spill.
// The PE32 replay retains signalling payload bits; a payload that vanishes
// when narrowed becomes floating-point indefinite. ARM FCVT quiets SNaNs,
// so the original spill's bits must be formed explicitly for NaN inputs.
func PlayerKotrSpill54DC40(value float64) float32 {
	bits := math.Float64bits(value)
	if bits&0x7ff0000000000000 == 0x7ff0000000000000 && bits&0x000fffffffffffff != 0 {
		payload := uint32((bits & 0x000fffffffffffff) >> 29)
		if payload == 0 {
			return math.Float32frombits(0xffc00000)
		}
		return math.Float32frombits(uint32(bits>>32)&0x80000000 | 0x7f800000 | payload)
	}
	return float32(value)
}

func PlayerHandleKotrDeath54DC40(victim, killer *Object, rt PlayerHandleKotrDeathRuntime54DC40) {
	victimUpdate := (*PlayerUpdateData)(victim.UpdateData)
	var victimTeam, killerTeam *Team
	if rt.HasTeam(&victim.TeamVal) {
		victimTeam = rt.TeamByID(uint8(victim.TeamVal.ID))
	}
	if killer == nil {
		return
	}
	killerUpdate := (*PlayerUpdateData)(killer.UpdateData)
	if rt.HasTeam(&killer.TeamVal) {
		killerTeam = rt.TeamByID(uint8(killer.TeamVal.ID))
	}
	if !killer.ObjClass.Has(object.ClassPlayer) {
		return
	}
	if killer == victim || victimTeam != nil && victimTeam == killerTeam {
		if rt.IsCrown(killer) {
			rt.SubtractScore(killer, 1)
			rt.ReportLesson(killer)
			if killerTeam != nil {
				rt.TeamChangeLessons(killerTeam, int32(uint32(killerTeam.Lessons)-1))
			}
			if rt.ObserverMode() != 0 && killerUpdate != nil {
				player := killerUpdate.Player
				rt.ObserverUpdate(player, player)
			}
		}
	} else if killerTeam == nil || killerTeam == victimTeam {
		if rt.IsCrown(killer) || rt.IsCrown(victim) {
			key := "KotRPawnKillsKingPoints"
			if rt.IsCrown(killer) {
				key = "KotRKingKillsPawnPoints"
			}
			points := rt.FloatToInt(PlayerKotrSpill54DC40(rt.BalanceFloat(key)))
			rt.AddScore(killer, points)
			rt.ReportLesson(killer)
			if rt.ObserverMode() != 0 && killerUpdate != nil && victimUpdate != nil {
				second := victimUpdate.Player
				first := killerUpdate.Player
				rt.ObserverUpdate(first, second)
			}
			if !rt.GameplayFlag(4) && rt.IsCrown(victim) {
				rt.DropCrowns(victim, killer)
			}
		}
	} else {
		key := ""
		if rt.IsCrown(killer) {
			key = "KotRKingKillsPawnPoints"
			if rt.IsCrown(victim) {
				key = "KotRKingKillsKingPoints"
			}
		} else if rt.IsCrown(victim) {
			key = "KotRPawnKillsKingPoints"
		}
		if key != "" {
			points := rt.FloatToInt(PlayerKotrSpill54DC40(rt.BalanceFloat(key)))
			rt.AddScore(killer, points)
			rt.TeamChangeLessons(killerTeam, int32(uint32(killerTeam.Lessons)+points))
			rt.ReportLesson(killer)
			if rt.ObserverMode() != 0 && killerUpdate != nil && victimUpdate != nil {
				second := victimUpdate.Player
				first := killerUpdate.Player
				rt.ObserverUpdate(first, second)
			}
		}
	}
	rt.IncrementElimDeath(victim)
	rt.ReportLesson(victim)
}
