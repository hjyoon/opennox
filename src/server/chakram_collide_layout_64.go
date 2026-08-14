//go:build amd64 || arm64

package server

import "unsafe"

var (
	_ = [1]struct{}{}[40-unsafe.Sizeof(ChakramUpdateData{})]
	_ = [1]struct{}{}[16-unsafe.Offsetof(ChakramUpdateData{}.LastHit)]
	_ = [1]struct{}{}[24-unsafe.Offsetof(ChakramUpdateData{}.OwnerPos)]
	_ = [1]struct{}{}[32-unsafe.Offsetof(ChakramUpdateData{}.ReturnState)]
	_ = [1]struct{}{}[48-unsafe.Sizeof(ChakramAttackData{})]
	_ = [1]struct{}{}[16-unsafe.Offsetof(ChakramAttackData{}.Owner)]
	_ = [1]struct{}{}[24-unsafe.Offsetof(ChakramAttackData{}.PosX)]
	_ = [1]struct{}{}[28-unsafe.Offsetof(ChakramAttackData{}.PosY)]
	_ = [1]struct{}{}[32-unsafe.Offsetof(ChakramAttackData{}.Field24)]
	_ = [1]struct{}{}[40-unsafe.Offsetof(ChakramAttackData{}.Source)]
)
