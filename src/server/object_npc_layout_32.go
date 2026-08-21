//go:build 386 || arm

package server

import "unsafe"

var (
	_ = [1]struct{}{}[1316-unsafe.Sizeof(NPC{})]
	_ = [1]struct{}{}[32-unsafe.Offsetof(NPC{}.Weapon)]
	_ = [1]struct{}{}[680-unsafe.Offsetof(NPC{}.Armor)]
	_ = [1]struct{}{}[1304-unsafe.Offsetof(NPC{}.WeaponEquip)]
	_ = [1]struct{}{}[1308-unsafe.Offsetof(NPC{}.ArmorEquip)]
	_ = [1]struct{}{}[1312-unsafe.Offsetof(NPC{}.Field1312)]
)
