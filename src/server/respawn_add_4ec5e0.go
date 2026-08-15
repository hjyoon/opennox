package server

// RespawnAddHooks4EC5E0 separates the pointer-width-independent GAME.EXE
// contract from the legacy allocator, object, record, and list storage.
type RespawnAddHooks4EC5E0[O any, R comparable, A any, U any] struct {
	LoadAllow          func() uint32
	LoadAllocator      func() A
	AllocZero          func(A) R
	LoadTypeInd        func(O) uint16
	StoreObject        func(R, O)
	StoreTypeInd       func(R, uint32)
	LoadPositionXBits  func(O) uint32
	StorePositionXBits func(R, uint32)
	LoadPositionYBits  func(O) uint32
	StorePositionYBits func(R, uint32)
	LoadDirection      func(O) uint16
	StoreDirection     func(R, uint16)
	LoadClass          func(O) uint32
	CopyModifierAttrs  func(R, O)
	WeaponEquipFlags   func(O) uint32
	LoadUseData        func(O) U
	LoadUseByte        func(U, uint32) uint8
	StoreCharge1       func(R, uint8)
	StoreCharge0       func(R, uint8)
	LoadHead           func() R
	StoreNext          func(R, R)
	StorePrev          func(R, R)
	StoreHead          func(R)
}

// RespawnAdd4EC5E0 preserves GAME.EXE 004EC5E0. In particular, the list head
// is not cached at entry: it is loaded once for the new record's next link and
// then loaded again for the old head's prev link and the function result.
func RespawnAdd4EC5E0[O any, R comparable, A any, U any](obj O, hooks RespawnAddHooks4EC5E0[O, R, A, U]) R {
	var zero R
	if hooks.LoadAllow() == 0 {
		return zero
	}
	allocator := hooks.LoadAllocator()
	rec := hooks.AllocZero(allocator)
	if rec == zero {
		return zero
	}

	typeInd := hooks.LoadTypeInd(obj)
	hooks.StoreObject(rec, obj)
	hooks.StoreTypeInd(rec, uint32(typeInd))
	x := hooks.LoadPositionXBits(obj)
	hooks.StorePositionXBits(rec, x)
	y := hooks.LoadPositionYBits(obj)
	hooks.StorePositionYBits(rec, y)
	direction := hooks.LoadDirection(obj)
	hooks.StoreDirection(rec, direction)

	if hooks.LoadClass(obj)&0x13001000 != 0 {
		hooks.CopyModifierAttrs(rec, obj)
	}
	if hooks.LoadClass(obj)&0x01000000 != 0 {
		flags := hooks.WeaponEquipFlags(obj)
		if uint8(flags)&0x82 != 0 {
			useData := hooks.LoadUseData(obj)
			charge1 := hooks.LoadUseByte(useData, 1)
			hooks.StoreCharge1(rec, charge1)
			charge0 := hooks.LoadUseByte(useData, 0)
			hooks.StoreCharge0(rec, charge0)
		}
	}

	hooks.StorePrev(rec, zero)
	firstHead := hooks.LoadHead()
	hooks.StoreNext(rec, firstHead)
	secondHead := hooks.LoadHead()
	if secondHead != zero {
		hooks.StorePrev(secondHead, rec)
	}
	hooks.StoreHead(rec)
	return secondHead
}
