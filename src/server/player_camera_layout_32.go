//go:build 386 || arm

package server

import "unsafe"

var (
	_ = [1]struct{}{}[556-unsafe.Sizeof(PlayerUpdateData{})]
	_ = [1]struct{}{}[116-unsafe.Offsetof(PlayerUpdateData{}.Field29)]
	_ = [1]struct{}{}[244-unsafe.Offsetof(PlayerUpdateData{}.CurTraps)]
	_ = [1]struct{}{}[276-unsafe.Offsetof(PlayerUpdateData{}.Player)]
	_ = [1]struct{}{}[312-unsafe.Offsetof(PlayerUpdateData{}.Field78)]
	_ = [1]struct{}{}[316-unsafe.Offsetof(PlayerUpdateData{}.Field79)]
	_ = [1]struct{}{}[552-unsafe.Offsetof(PlayerUpdateData{}.Field138)]

	_ = [1]struct{}{}[4828-unsafe.Sizeof(Player{})]
	_ = [1]struct{}{}[2056-unsafe.Offsetof(Player{}.PlayerUnit)]
	_ = [1]struct{}{}[2064-unsafe.Offsetof(Player{}.PlayerInd)]
	_ = [1]struct{}{}[2068-unsafe.Offsetof(Player{}.Field2068)]
	_ = [1]struct{}{}[2092-unsafe.Offsetof(Player{}.Active)]
	_ = [1]struct{}{}[2096-unsafe.Offsetof(Player{}.Field2096Buf)]
	_ = [1]struct{}{}[2185-unsafe.Offsetof(Player{}.info)]
	_ = [1]struct{}{}[2284-unsafe.Offsetof(Player{}.CursorVec)]
	_ = [1]struct{}{}[3628-unsafe.Offsetof(Player{}.CameraFollowObj)]
	_ = [1]struct{}{}[3632-unsafe.Offsetof(Player{}.Pos3632Vec)]
	_ = [1]struct{}{}[3672-unsafe.Offsetof(Player{}.Field3672)]
	_ = [1]struct{}{}[3680-unsafe.Offsetof(Player{}.Field3680)]
	_ = [1]struct{}{}[3688-unsafe.Offsetof(Player{}.Field3688)]
	_ = [1]struct{}{}[3692-unsafe.Offsetof(Player{}.Field3692)]
	_ = [1]struct{}{}[4792-unsafe.Offsetof(Player{}.Field4792)]
	_ = [1]struct{}{}[66-unsafe.Offsetof(PlayerInfo{}.playerClass)]
)
