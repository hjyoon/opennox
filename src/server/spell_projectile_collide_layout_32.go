//go:build 386 || arm

package server

import "unsafe"

// GAME.EXE offsets consumed by SpellProjectileCollide and its inversion scan.
var (
	_ = [1]struct{}{}[88-unsafe.Offsetof(PlayerUpdateData{}.State)]
	_ = [1]struct{}{}[236-unsafe.Offsetof(PlayerUpdateData{}.Field59_0)]
	_ = [1]struct{}{}[276-unsafe.Offsetof(PlayerUpdateData{}.Player)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(Player{}.WeaponEquip)]
	_ = [1]struct{}{}[0-unsafe.Offsetof(ModifierInitData{}.Modifiers)]
	_ = [1]struct{}{}[88-unsafe.Offsetof(ModifierEff{}.DefendCollide88)]
)
