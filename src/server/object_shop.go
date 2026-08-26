package server

import "unsafe"

type GoldInitData struct {
	Amount uint32 // 0, 0
}

const ShopkeeperItemDefinitionCount = 60

// ShopkeeperItemDefinition is the exact fixed-width 28-byte record stored in
// ShopkeeperInitData. ModifierSlots are PE32 pointer slots in GAME.EXE. They
// remain fixed-width provenance here; the native loader must resolve them
// before supporting modified merchant items on a 64-bit target.
type ShopkeeperItemDefinition struct {
	TypeInd       uint32    // 0, 0
	Count         uint8     // 1, 4
	_             [3]byte   // 1, 5
	Param         uint32    // 2, 8
	ModifierSlots [4]uint32 // 3, 12
}

// ShopkeeperInitData preserves the exact 1,724-byte GAME.EXE layout. The
// record array is fixed-width serialized/map data rather than live native
// pointers, so its offsets are identical on 32- and 64-bit targets.
type ShopkeeperInitData struct {
	Count          uint8                                                   // 0, 0
	_              [3]byte                                                 // 0, 1
	Items          [ShopkeeperItemDefinitionCount]ShopkeeperItemDefinition // 1, 4
	ShopText       [32]byte                                                // 421, 1684
	BuyMultiplier  float32                                                 // 429, 1716
	SellMultiplier float32                                                 // 430, 1720
}

var (
	_ = [1]struct{}{}[28-unsafe.Sizeof(ShopkeeperItemDefinition{})]
	_ = [1]struct{}{}[unsafe.Sizeof(ShopkeeperItemDefinition{})-28]
	_ = [1]struct{}{}[4-unsafe.Offsetof(ShopkeeperItemDefinition{}.Count)]
	_ = [1]struct{}{}[8-unsafe.Offsetof(ShopkeeperItemDefinition{}.Param)]
	_ = [1]struct{}{}[12-unsafe.Offsetof(ShopkeeperItemDefinition{}.ModifierSlots)]

	_ = [1]struct{}{}[1724-unsafe.Sizeof(ShopkeeperInitData{})]
	_ = [1]struct{}{}[unsafe.Sizeof(ShopkeeperInitData{})-1724]
	_ = [1]struct{}{}[4-unsafe.Offsetof(ShopkeeperInitData{}.Items)]
	_ = [1]struct{}{}[1684-unsafe.Offsetof(ShopkeeperInitData{}.ShopText)]
	_ = [1]struct{}{}[1716-unsafe.Offsetof(ShopkeeperInitData{}.BuyMultiplier)]
	_ = [1]struct{}{}[1720-unsafe.Offsetof(ShopkeeperInitData{}.SellMultiplier)]
)
