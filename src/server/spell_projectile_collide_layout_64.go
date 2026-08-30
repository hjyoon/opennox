//go:build amd64 || arm64

package server

import "unsafe"

// Native-pointer offsets corresponding to the fields used by 004E9500.
var (
	_ = [1]struct{}{}[88-unsafe.Offsetof(PlayerUpdateData{}.State)]
	_ = [1]struct{}{}[296-unsafe.Offsetof(PlayerUpdateData{}.Field59_0)]
	_ = [1]struct{}{}[336-unsafe.Offsetof(PlayerUpdateData{}.Player)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(Player{}.WeaponEquip)]
	_ = [1]struct{}{}[0-unsafe.Offsetof(ModifierInitData{}.Modifiers)]
	_ = [1]struct{}{}[128-unsafe.Offsetof(ModifierEff{}.DefendCollide88)]
)
