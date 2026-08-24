//go:build 386 || arm

package client

import "unsafe"

// DrawableUnion is the original five-dword drawable union on 32-bit targets.
type DrawableUnion struct {
	_ [0]uintptr
	_ [20]byte
}

var (
	_ = [1]struct{}{}[20-unsafe.Sizeof(DrawableUnion{})]
	_ = [1]struct{}{}[4-unsafe.Alignof(DrawableUnion{})]
	_ = [1]struct{}{}[unsafe.Sizeof(DrawableUnionMonster{})-unsafe.Sizeof(DrawableUnion{})]
	_ = [1]struct{}{}[unsafe.Sizeof(DrawableUnionItem{})-unsafe.Sizeof(DrawableUnion{})]
	_ = [1]struct{}{}[unsafe.Sizeof(DrawableUnionEffect{})-unsafe.Sizeof(DrawableUnion{})]
	_ = [1]struct{}{}[unsafe.Sizeof(DrawableUnionDoor{})-unsafe.Sizeof(DrawableUnion{})]

	_ = [1]struct{}{}[48-unsafe.Sizeof(AnimationVector{})]
	_ = [1]struct{}{}[148-unsafe.Sizeof(AnimationStateDrawData{})]
	_ = [1]struct{}{}[772-unsafe.Sizeof(MonsterDrawData{})]
	_ = [1]struct{}{}[40-unsafe.Sizeof(PlayerEquipAnimation{})]
	_ = [1]struct{}{}[264-unsafe.Sizeof(PlayerAnimation{})]
	_ = [1]struct{}{}[14524-unsafe.Sizeof(PlayerDrawData{})]
)
