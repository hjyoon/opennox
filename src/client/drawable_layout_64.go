//go:build amd64 || arm64

package client

import "unsafe"

// DrawableUnion grows only for native 64-bit builds because the door view
// owns four image handles. Numeric union views keep their original 20-byte
// prefix, while the pointer-bearing view receives enough native-width space.
type DrawableUnion struct {
	_ [0]uintptr
	_ [40]byte
}

var (
	_ = [1]struct{}{}[40-unsafe.Sizeof(DrawableUnion{})]
	_ = [1]struct{}{}[8-unsafe.Alignof(DrawableUnion{})]
	_ = [unsafe.Sizeof(DrawableUnion{}) - unsafe.Sizeof(DrawableUnionMonster{})]byte{}
	_ = [unsafe.Sizeof(DrawableUnion{}) - unsafe.Sizeof(DrawableUnionItem{})]byte{}
	_ = [unsafe.Sizeof(DrawableUnion{}) - unsafe.Sizeof(DrawableUnionEffect{})]byte{}
	_ = [1]struct{}{}[unsafe.Sizeof(DrawableUnionDoor{})-unsafe.Sizeof(DrawableUnion{})]

	_ = [1]struct{}{}[88-unsafe.Sizeof(AnimationVector{})]
	_ = [1]struct{}{}[272-unsafe.Sizeof(AnimationStateDrawData{})]
	_ = [1]struct{}{}[1416-unsafe.Sizeof(MonsterDrawData{})]
	_ = [1]struct{}{}[80-unsafe.Sizeof(PlayerEquipAnimation{})]
	_ = [1]struct{}{}[520-unsafe.Sizeof(PlayerAnimation{})]
	_ = [1]struct{}{}[28608-unsafe.Sizeof(PlayerDrawData{})]
)
