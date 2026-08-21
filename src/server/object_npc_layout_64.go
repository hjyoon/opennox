//go:build amd64 || arm64

package server

import "unsafe"

// These are native Go offsets. The original Win32 NPC layout remains covered
// by object_npc_layout_32.go and by the legacy C ABI assertions.
var (
	_ = [1]struct{}{}[2592-unsafe.Sizeof(NPC{})]
	_ = [1]struct{}{}[32-unsafe.Offsetof(NPC{}.Weapon)]
	_ = [1]struct{}{}[1328-unsafe.Offsetof(NPC{}.Armor)]
	_ = [1]struct{}{}[2576-unsafe.Offsetof(NPC{}.WeaponEquip)]
	_ = [1]struct{}{}[2580-unsafe.Offsetof(NPC{}.ArmorEquip)]
	_ = [1]struct{}{}[2584-unsafe.Offsetof(NPC{}.Field1312)]
)
