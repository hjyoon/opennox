package server

import (
	"image"

	"github.com/opennox/libs/types"
)

const (
	createSpellProjectilePlayerClass4FDDA0 = uint8(0x04)
	createSpellProjectileType4FDDA0        = "Magic"
	createSpellProjectileRadiusAdd4FDDA0   = float32(4)
	createSpellProjectileSearchDist4FDDA0  = float32(600)
	createSpellProjectileTraceFlag4FDDA0   = int32(5)
	createSpellProjectileEnchant4FDDA0     = int32(21)
)

// Each helper is a separate instruction boundary. The original executable
// runs this routine with x87 53-bit precision and spills at the indicated
// binary32 stores.
//
//go:noinline
func createSpellProjectileAdd64_4FDDA0(a, b float64) float64 { return a + b }

//go:noinline
func createSpellProjectileMul64_4FDDA0(a, b float64) float64 { return a * b }

//go:noinline
func createSpellProjectileSpill32_4FDDA0(value float64) float32 { return float32(value) }

//go:noinline
func createSpellProjectileI32ToF32_4FDDA0(value int32) float32 { return float32(value) }

type createSpellProjectileRay4FDDA0 struct {
	Origin      types.Pointf
	Destination types.Pointf
}

type createSpellProjectileHooks4FDDA0[Object comparable, Update, Player any] struct {
	loadSourceArg func() Object
	loadTargetArg func() Object
	loadSpellArg  func() int32

	loadRadius   func(Object) float32
	loadClassLow func(Object) uint8
	loadUpdate   func(Object) Update
	loadPlayer   func(Update) Player
	loadCursorX  func(Player) int32
	loadCursorY  func(Player) int32
	spellFlags   func(int32) uint32
	searchTarget func(*types.Pointf, Object, uint32, float32, int32, Object) Object

	loadDirection func(Object) int16
	loadPosX      func(Object) float32
	loadPosY      func(Object) float32
	directionX    func(int16) float32
	directionY    func(int16) float32
	loadVelX      func(Object) float32
	loadVelY      func(Object) float32
	mapTrace      func(*createSpellProjectileRay4FDDA0, *types.Pointf, *image.Point, int32) int32

	newObject            func(string) Object
	loadProjectileUpdate func(Object) Update
	spellPower           func(int32, Object) int32
	storeLevel           func(Update, uint32)
	createAt             func(Object, Object, types.Pointf, int32)
	storeDirection1      func(Object, uint16)
	storeDirection2      func(Object, uint16)
	storeField0          func(Update, Object)
	storeTarget          func(Update, Object)
	storeField8          func(Update, Object)
	storeSpell           func(Update, uint32)
	indexedDirection     func(int16, *types.Pointf)
	loadSpeed            func(Object) float32
	storeVelX            func(Object, float32)
	storeVelY            func(Object, float32)

	hasEnchant   func(Object, int32) int32
	enchantPower func(Object, int32) int32
	enchantTimer func(Object, int32) int32
	applyEnchant func(Object, int32, int16, uint8)
	spellAudio   func(int32, int32) int32
	audio        func(int32, Object, int32, uint32)
}

// createSpellProjectile4FDDA0 preserves GAME.EXE 004FDDA0's load order,
// callback order, x87 spill points, nil exits, and fixed-width stores. Object,
// update-data, and player values stay opaque and native-width.
//
// Source, target, and spell arguments are captured in that order. A nil
// target is searched lazily; player cursor X/Y convert from signed dwords and
// reload the Player pointer between components. The source origin is cached,
// while position and velocity contributions are deliberately loaded live.
// Direction table indices are signed int16 values and must not be wrapped.
//
// The CreateAt callback includes the original fifth reserved zero argument;
// GAME.EXE's 004DAA50 ignores it. The indexed-direction callback reuses the
// same eight-byte scratch object as the optional player aim, exactly as the
// original stack frame does, although its output is not consumed here.
func createSpellProjectile4FDDA0[Object comparable, Update, Player any](
	hooks createSpellProjectileHooks4FDDA0[Object, Update, Player],
) Object {
	var zero Object

	source := hooks.loadSourceArg()
	target := hooks.loadTargetArg()
	spellID := hooks.loadSpellArg()

	radius := createSpellProjectileSpill32_4FDDA0(createSpellProjectileAdd64_4FDDA0(
		float64(hooks.loadRadius(source)),
		float64(createSpellProjectileRadiusAdd4FDDA0),
	))

	var scratch types.Pointf
	if target == zero {
		if hooks.loadClassLow(source)&createSpellProjectilePlayerClass4FDDA0 != 0 {
			update := hooks.loadUpdate(source)
			playerX := hooks.loadPlayer(update)
			scratch.X = createSpellProjectileI32ToF32_4FDDA0(hooks.loadCursorX(playerX))
			playerY := hooks.loadPlayer(update)
			scratch.Y = createSpellProjectileI32ToF32_4FDDA0(hooks.loadCursorY(playerY))
			flags := hooks.spellFlags(spellID)
			target = hooks.searchTarget(
				&scratch, source, flags, createSpellProjectileSearchDist4FDDA0, 0, source,
			)
		} else {
			flags := hooks.spellFlags(spellID)
			target = hooks.searchTarget(
				nil, source, flags, createSpellProjectileSearchDist4FDDA0, 0, source,
			)
		}
	}

	direction := hooks.loadDirection(source)
	originX := hooks.loadPosX(source)
	originY := hooks.loadPosY(source)

	xBeforeVelocity := createSpellProjectileAdd64_4FDDA0(
		createSpellProjectileMul64_4FDDA0(float64(radius), float64(hooks.directionX(direction))),
		float64(hooks.loadPosX(source)),
	)
	yBeforeVelocity := createSpellProjectileSpill32_4FDDA0(createSpellProjectileAdd64_4FDDA0(
		createSpellProjectileMul64_4FDDA0(float64(radius), float64(hooks.directionY(direction))),
		float64(hooks.loadPosY(source)),
	))
	destinationX := createSpellProjectileSpill32_4FDDA0(createSpellProjectileAdd64_4FDDA0(
		xBeforeVelocity, float64(hooks.loadVelX(source)),
	))
	destinationY := createSpellProjectileSpill32_4FDDA0(createSpellProjectileAdd64_4FDDA0(
		float64(yBeforeVelocity), float64(hooks.loadVelY(source)),
	))
	ray := createSpellProjectileRay4FDDA0{
		Origin:      types.Pointf{X: originX, Y: originY},
		Destination: types.Pointf{X: destinationX, Y: destinationY},
	}
	if hooks.mapTrace(&ray, nil, nil, createSpellProjectileTraceFlag4FDDA0) == 0 {
		return zero
	}

	projectile := hooks.newObject(createSpellProjectileType4FDDA0)
	if projectile == zero {
		return zero
	}

	projectileUpdate := hooks.loadProjectileUpdate(projectile)
	level := hooks.spellPower(spellID, source)
	hooks.storeLevel(projectileUpdate, uint32(level))
	hooks.createAt(projectile, source, ray.Destination, 0)

	directionBits := uint16(hooks.loadDirection(source))
	hooks.storeDirection1(projectile, directionBits)
	hooks.storeDirection2(projectile, directionBits)
	hooks.storeField0(projectileUpdate, source)
	hooks.storeTarget(projectileUpdate, target)
	hooks.storeField8(projectileUpdate, source)
	hooks.storeSpell(projectileUpdate, uint32(spellID))
	hooks.indexedDirection(hooks.loadDirection(source), &scratch)

	projectileDirectionX := hooks.loadDirection(projectile)
	projectileDirectionY := hooks.loadDirection(projectile)
	velocityX := createSpellProjectileSpill32_4FDDA0(createSpellProjectileMul64_4FDDA0(
		float64(hooks.directionX(projectileDirectionX)),
		float64(hooks.loadSpeed(projectile)),
	))
	hooks.storeVelX(projectile, velocityX)
	velocityY := createSpellProjectileSpill32_4FDDA0(createSpellProjectileMul64_4FDDA0(
		float64(hooks.directionY(projectileDirectionY)),
		float64(hooks.loadSpeed(projectile)),
	))
	hooks.storeVelY(projectile, velocityY)

	velocityX = createSpellProjectileSpill32_4FDDA0(createSpellProjectileAdd64_4FDDA0(
		float64(hooks.loadVelX(projectile)),
		float64(hooks.loadVelX(source)),
	))
	hooks.storeVelX(projectile, velocityX)
	velocityY = createSpellProjectileSpill32_4FDDA0(createSpellProjectileAdd64_4FDDA0(
		float64(hooks.loadVelY(projectile)),
		float64(hooks.loadVelY(source)),
	))
	hooks.storeVelY(projectile, velocityY)

	if hooks.hasEnchant(source, createSpellProjectileEnchant4FDDA0) != 0 {
		power := hooks.enchantPower(source, createSpellProjectileEnchant4FDDA0)
		duration := hooks.enchantTimer(source, createSpellProjectileEnchant4FDDA0)
		hooks.applyEnchant(
			projectile,
			createSpellProjectileEnchant4FDDA0,
			int16(duration),
			uint8(power),
		)
	}
	audioID := hooks.spellAudio(spellID, 0)
	hooks.audio(audioID, source, 0, 0)
	return projectile
}
