package server

// DirectionInitData is the exact eight-byte record shared by SkullInit and
// DirectionInit. GAME.EXE treats both components as signed 32-bit integers.
type DirectionInitData struct {
	X int32
	Y int32
}

type directionToAngleHooks509E00[D any] struct {
	loadY     func(D) int32
	loadX     func(D) int32
	loadTable func(int32) uint32
}

// directionToAngle509E00 preserves the observable load order of GAME.EXE
// 00509E00. Y is loaded before X; x+3*y uses 32-bit arithmetic; the resulting
// signed index is then passed unchanged to the centered lookup table. The
// original has no nil or bounds guard.
func directionToAngle509E00[D any](data D, hooks directionToAngleHooks509E00[D]) uint32 {
	y := hooks.loadY(data)
	x := hooks.loadX(data)
	return hooks.loadTable(x + 3*y)
}

type skullInitHooks4F0450[O, I, U any] struct {
	loadInitData          func(O) I
	loadUpdateData        func(O) U
	directionToAngle      func(I) uint32
	storeDirection2       func(O, uint16)
	storeDirection1       func(O, uint16)
	resolveProjectileType func(U) int32
	storeProjectileType   func(U, uint32)
}

// skullInit4F0450 preserves GAME.EXE 004F0450's pointer caching, call order,
// low-word direction stores, full-width projectile type store, and return
// value. InitData and UpdateData are each loaded once before the first helper
// call. Direction2 is written before Direction1. The cached UpdateData is used
// both to resolve the name at offset 16 and to store the resulting type at
// offset 12, even if a callback changes the live object pointer fields. The
// original has no nil guards.
func skullInit4F0450[O, I, U any](unit O, hooks skullInitHooks4F0450[O, I, U]) int32 {
	initData := hooks.loadInitData(unit)
	updateData := hooks.loadUpdateData(unit)
	angle := hooks.directionToAngle(initData)
	hooks.storeDirection2(unit, uint16(angle))
	hooks.storeDirection1(unit, uint16(angle))
	projectileType := hooks.resolveProjectileType(updateData)
	hooks.storeProjectileType(updateData, uint32(projectileType))
	return projectileType
}
