//go:build amd64 || arm64

package server

import "unsafe"

// These are the current native Go offsets, not the original Win32 layout.
// Player and PlayerUpdateData still require a full legacy/runtime split.
var (
	_ = [1]struct{}{}[88-unsafe.Sizeof(PlayerJournal{})]
	_ = [1]struct{}{}[32-unsafe.Sizeof(MinimapItem{})]
	_ = [1]struct{}{}[48-unsafe.Sizeof(EquipmentData{})]

	_ = [1]struct{}{}[640-unsafe.Sizeof(PlayerUpdateData{})]
	_ = [1]struct{}{}[104-unsafe.Offsetof(PlayerUpdateData{}.EquippedWeapon)]
	_ = [1]struct{}{}[8-unsafe.Sizeof(PlayerUpdateData{}.EquippedWeapon)]
	_ = [1]struct{}{}[120-unsafe.Offsetof(PlayerUpdateData{}.Field29)]
	_ = [1]struct{}{}[288-unsafe.Offsetof(PlayerUpdateData{}.CurTraps)]
	_ = [1]struct{}{}[316-unsafe.Offsetof(PlayerUpdateData{}.Field68)]
	_ = [1]struct{}{}[320-unsafe.Offsetof(PlayerUpdateData{}.Player)]
	_ = [1]struct{}{}[336-unsafe.Offsetof(PlayerUpdateData{}.DialogWith)]
	_ = [1]struct{}{}[360-unsafe.Offsetof(PlayerUpdateData{}.CollisionWall)]
	_ = [1]struct{}{}[376-unsafe.Offsetof(PlayerUpdateData{}.SoulGate)]
	_ = [1]struct{}{}[384-unsafe.Offsetof(PlayerUpdateData{}.QuestExit)]
	_ = [1]struct{}{}[392-unsafe.Offsetof(PlayerUpdateData{}.QuestWarpGate)]
	_ = [1]struct{}{}[632-unsafe.Offsetof(PlayerUpdateData{}.Field138)]

	_ = [1]struct{}{}[6160-unsafe.Sizeof(Player{})]
	_ = [1]struct{}{}[2056-unsafe.Offsetof(Player{}.PlayerUnit)]
	_ = [1]struct{}{}[2068-unsafe.Offsetof(Player{}.PlayerInd)]
	_ = [1]struct{}{}[2072-unsafe.Offsetof(Player{}.Field2068)]
	_ = [1]struct{}{}[2096-unsafe.Offsetof(Player{}.Active)]
	_ = [1]struct{}{}[2100-unsafe.Offsetof(Player{}.Field2096Buf)]
	_ = [1]struct{}{}[2189-unsafe.Offsetof(Player{}.info)]
	_ = [1]struct{}{}[2288-unsafe.Offsetof(Player{}.CursorVec)]
	_ = [1]struct{}{}[4884-unsafe.Offsetof(Player{}.Field3600)]
	_ = [1]struct{}{}[4888-unsafe.Offsetof(Player{}.Field3604)]
	_ = [1]struct{}{}[4892-unsafe.Offsetof(Player{}.field3608)]
	_ = [1]struct{}{}[4912-unsafe.Offsetof(Player{}.CameraFollowObj)]
	_ = [1]struct{}{}[4920-unsafe.Offsetof(Player{}.Pos3632Vec)]
	_ = [1]struct{}{}[4968-unsafe.Offsetof(Player{}.Field3672)]
	_ = [1]struct{}{}[4976-unsafe.Offsetof(Player{}.Field3680)]
	_ = [1]struct{}{}[4984-unsafe.Offsetof(Player{}.Field3688)]
	_ = [1]struct{}{}[4988-unsafe.Offsetof(Player{}.Field3692)]
	_ = [1]struct{}{}[5956-unsafe.Offsetof(Player{}.field4652)]
	_ = [1]struct{}{}[5960-unsafe.Offsetof(Player{}.field4656)]
	_ = [1]struct{}{}[5964-unsafe.Offsetof(Player{}.field4660)]
	_ = [1]struct{}{}[5968-unsafe.Offsetof(Player{}.field4664)]
	_ = [1]struct{}{}[5972-unsafe.Offsetof(Player{}.field4668)]
	_ = [1]struct{}{}[5976-unsafe.Offsetof(Player{}.field4672)]
	_ = [1]struct{}{}[5980-unsafe.Offsetof(Player{}.field4676)]
	_ = [1]struct{}{}[5984-unsafe.Offsetof(Player{}.field4680)]
	_ = [1]struct{}{}[5988-unsafe.Offsetof(Player{}.field4684)]
	_ = [1]struct{}{}[5992-unsafe.Offsetof(Player{}.field4688)]
	_ = [1]struct{}{}[5996-unsafe.Offsetof(Player{}.field4692)]
	_ = [1]struct{}{}[6000-unsafe.Offsetof(Player{}.QuestStage)]
	_ = [1]struct{}{}[6096-unsafe.Offsetof(Player{}.Field4792)]
	_ = [1]struct{}{}[6104-unsafe.Offsetof(Player{}.QuestAnkhs)]
	_ = [1]struct{}{}[40-unsafe.Sizeof(Player{}.QuestAnkhs)]
	_ = [1]struct{}{}[66-unsafe.Offsetof(PlayerInfo{}.playerClass)]
)
