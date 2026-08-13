package server

const (
	pickupCollideMonsterClassByte4E8DF0 = uint8(0x02)
	pickupCollidePlayerClassByte4E8DF0  = uint8(0x04)
	pickupCollideMovementGate4E8DF0     = uint8(0x01)
	pickupCollideInventoryFlag4E8DF0    = int32(1)
)

type pickupCollideResult4E8DF0[O any] struct {
	unit               O
	inventoryResult    uint32
	inventoryAttempted bool
}

type pickupCollideHooks4E8DF0[O, U any] struct {
	loadClassByte         func(O) uint8
	frame                 func() uint32
	loadPickupFrame       func(O) uint32
	fps                   func() uint32
	loadUpdateData        func(O) U
	loadMovementFlagsByte func(U) uint8
	placeInventory        func(O, O, int32, int32) uint32
}

// pickupCollide4E8DF0 preserves GAME.EXE 004E8DF0. The unit class byte is
// cached before the frame reads, elapsed time uses wrapping uint32 arithmetic,
// and FPS is shifted logically. Guard paths return the original unit while an
// inventory attempt forwards the callee's 32-bit result. The registered third
// collision argument is intentionally ignored.
func pickupCollide4E8DF0[O comparable, U, C any](
	item, unit O,
	collision C,
	hooks pickupCollideHooks4E8DF0[O, U],
) pickupCollideResult4E8DF0[O] {
	_ = collision
	result := pickupCollideResult4E8DF0[O]{unit: unit}

	var zero O
	if unit == zero {
		return result
	}
	class := hooks.loadClassByte(unit)
	if class&pickupCollideMonsterClassByte4E8DF0 != 0 {
		return result
	}

	frame := hooks.frame()
	pickupFrame := hooks.loadPickupFrame(item)
	fps := hooks.fps()
	if frame-pickupFrame < fps>>1 {
		return result
	}

	if class&pickupCollidePlayerClassByte4E8DF0 != 0 {
		update := hooks.loadUpdateData(unit)
		if hooks.loadMovementFlagsByte(update)&pickupCollideMovementGate4E8DF0 == 0 {
			return result
		}
	}

	result.inventoryAttempted = true
	result.inventoryResult = hooks.placeInventory(
		unit,
		item,
		pickupCollideInventoryFlag4E8DF0,
		pickupCollideInventoryFlag4E8DF0,
	)
	return result
}
