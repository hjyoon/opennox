package server

import (
	"encoding/binary"
	"math"

	"github.com/opennox/libs/types"
)

// PlayerBotUpdate4FAB20 retains the entry-time player record while respawn
// and monster AI may replace the unit's live player or monster records.
func (unit *Object) PlayerBotUpdate4FAB20(rt PlayerBotRuntime4FAB20) uint32 {
	player := (*PlayerUpdateData)(unit.UpdateData)
	if player.Field73 == nil {
		unit.PlayerBotCreate4FA700(rt.Create)
		if player.Field73 == nil {
			unit.Update = rt.RegularPlayerUpdate
			return 0
		}
	}
	if result := unit.PlayerBotRespawn4FAC70(rt); result != 0 {
		return result
	}
	monster := player.Field73
	monster.StatusFlags |= 0x100
	unit.MorphPlayerBotFromPlayer4FAAC0()
	rt.UpdateMonster(unit)
	unit.MorphPlayerBotToPlayer4FAAF0()
	state := unit.PlayerBotActionToState4FABC0()
	info := player.Player
	player.State = PlayerState(state)
	player.Field59_0 = monster.Field120_1
	info.Pos3632Vec.X = unit.PosVec.X
	y := unit.PosVec.Y
	info.Pos3632Vec.Y = y
	return math.Float32bits(y)
}

// PlayerBotRespawn4FAC70 loads settings even for a living player or nil bot.
// Only a dead player reads the cached bot's unsigned frame timer. A failed
// constructor does not suppress the original default-item and respawn tail.
func (unit *Object) PlayerBotRespawn4FAC70(rt PlayerBotRuntime4FAB20) uint32 {
	monster := (*PlayerUpdateData)(unit.UpdateData).Field73
	settings := rt.LoadSettings()
	if unit.HealthData.Cur != 0 {
		return 0
	}
	if rt.Frame()-monster.Field137 < 2*rt.Create.FPS() {
		return 1
	}
	unit.PlayerBotCreate4FA700(rt.Create)
	rt.DefaultItems(unit, 1, 0)
	if rt.NetworkMode() == 0 || binary.LittleEndian.Uint32(settings.PlayerSkeletons58[:]) != 0 {
		rt.RespawnCorpse(&unit.PosVec, int16(unit.Direction1))
	}
	var position types.Pointf
	rt.FindStart(&position, unit)
	rt.Move(unit, &position)
	rt.Audio(148, unit, 0, 0)
	if rt.GameFlag(0x2000) {
		rt.Buff(unit, 23, int16(5*rt.Create.FPS()), 5)
	}
	return 0
}
