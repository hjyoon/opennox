package server

import "math"

const (
	deathBallDoorClassMask4E9E90      = uint8(0x80)
	deathBallDamageType4E9E90         = uint32(2)
	deathBallWallReflectAudio4E9E90   = uint32(37)
	deathBallDoorReflectAudio4E9E90   = uint32(283)
	deathBallCollideDamageKey4E9E90   = "DeathBallCollideDamage"
	deathBallReflectEpsilonBits57B770 = uint32(0x3dcccccd)
	deathBallDoorDirectionCount4E9E90 = int32(32)
)

type deathBallCollideHooks4E9E90[O, C, U, P comparable] struct {
	loadClassByte     func(O) uint8
	loadDoorUpdate    func(O) U
	loadPrevX         func(O) float32
	loadPrevY         func(O) float32
	storeNewX         func(O, float32)
	storeNewY         func(O, float32)
	loadPosX          func(O) float32
	loadPosY          func(O) float32
	loadDoorDirection func(U) int32
	loadDirectionY    func(int32) int32
	loadDirectionX    func(int32) int32
	doorReflect       func(O, float32, float32)
	audio             func(uint32, O)
	balanceFloat      func(string) float64
	floatToInt        func(float32) int32
	findParent        func(O) O
	targetDamage      func(O, O, O, int32, uint32) int32
	wallReflect       func(C, O)
	traceHitPoint     func() P
	loadTraceY        func(P) int32
	loadTraceX        func(P) int32
	damageMap         func(int32, int32, int32, uint32, O)
}

// Keep the Win32 x87 53-bit precision boundaries explicit and prevent a
// target compiler from contracting multiply-plus-add sequences to FMA.
//
//go:noinline
func deathBallAdd64(a, b float64) float64 { return a + b }

//go:noinline
func deathBallSub64(a, b float64) float64 { return a - b }

//go:noinline
func deathBallMul64(a, b float64) float64 { return a * b }

//go:noinline
func deathBallDiv64(a, b float64) float64 { return a / b }

//go:noinline
func deathBallSqrt64(value float64) float64 { return math.Sqrt(value) }

// deathBallCollide4E9E90 preserves GAME.EXE 004E9E90. The target class is
// the first object field observed. Door collisions copy both previous-position
// words before any live position reload, construct the original table normal,
// and return after reflection and audio. Non-Door targets use a freshly rounded
// balance value, then a live owner-chain and Damage callback. The wall path
// reflects and emits audio before acquiring the trace point; after balance
// conversion it reads trace Y before X.
func deathBallCollide4E9E90[O, C, U, P comparable](
	source, target O,
	collision C,
	hooks deathBallCollideHooks4E9E90[O, C, U, P],
) {
	var zeroObject O
	if target != zeroObject {
		if hooks.loadClassByte(target)&deathBallDoorClassMask4E9E90 != 0 {
			update := hooks.loadDoorUpdate(target)
			previousX := hooks.loadPrevX(source)
			previousY := hooks.loadPrevY(source)
			hooks.storeNewX(source, previousX)
			hooks.storeNewY(source, previousY)

			targetX := hooks.loadPosX(target)
			deltaX := deathBallSub64(float64(targetX), float64(hooks.loadPrevX(source)))
			targetY := hooks.loadPosY(target)
			directionY := hooks.loadDoorDirection(update)
			deltaY := deathBallSub64(float64(targetY), float64(hooks.loadPrevY(source)))

			normalX := float32(-float64(hooks.loadDirectionY(directionY)))
			directionX := hooks.loadDoorDirection(update)
			normalY := float32(hooks.loadDirectionX(directionX))
			dot := deathBallAdd64(
				deathBallMul64(float64(normalY), deltaY),
				deathBallMul64(float64(normalX), deltaX),
			)
			// GAME.EXE negates only for an ordered, strict positive result.
			// Equal, negative, and unordered values keep the original normal.
			if dot > 0 {
				normalX = -normalX
				normalY = -normalY
			}
			hooks.doorReflect(source, normalX, normalY)
			hooks.audio(deathBallDoorReflectAudio4E9E90, source)
			return
		}

		damage := hooks.floatToInt(float32(hooks.balanceFloat(deathBallCollideDamageKey4E9E90)))
		parent := hooks.findParent(source)
		_ = hooks.targetDamage(target, parent, source, damage, deathBallDamageType4E9E90)
		return
	}

	var zeroCollision C
	if collision == zeroCollision {
		return
	}
	hooks.wallReflect(collision, source)
	hooks.audio(deathBallWallReflectAudio4E9E90, source)
	point := hooks.traceHitPoint()
	var zeroPoint P
	if point == zeroPoint {
		return
	}
	damage := hooks.floatToInt(float32(hooks.balanceFloat(deathBallCollideDamageKey4E9E90)))
	y := hooks.loadTraceY(point)
	x := hooks.loadTraceX(point)
	hooks.damageMap(x, y, damage, deathBallDamageType4E9E90, source)
}

// deathBallTraceHitResult537760 preserves GAME.EXE 00537760: the fixed trace
// point is not acquired at all when the ready word is zero.
func deathBallTraceHitResult537760[P comparable](ready func() uint32, point func() P) P {
	if ready() == 0 {
		var zero P
		return zero
	}
	return point()
}

type deathBallDoorReflectHooks57B770[V comparable] struct {
	loadVelocityX  func(V) float32
	loadVelocityY  func(V) float32
	storeVelocityX func(V, float32)
	storeVelocityY func(V, float32)
}

// deathBallDoorReflectCore57B770 preserves GAME.EXE 0057B770. Inputs are
// binary32, intermediate x87 operations use the executable's 53-bit precision
// mode, and the four explicit local FSTP sites are represented by float32
// conversions. Velocity is reloaded in X,Y,Y,X order and stored X before Y.
func deathBallDoorReflectCore57B770[V comparable](
	velocity V,
	normalX, normalY float32,
	hooks deathBallDoorReflectHooks57B770[V],
) {
	nx := float64(normalX)
	ny := float64(normalY)
	lengthSquared := deathBallAdd64(deathBallMul64(nx, nx), deathBallMul64(ny, ny))
	denominator := deathBallAdd64(
		deathBallSqrt64(lengthSquared),
		float64(math.Float32frombits(deathBallReflectEpsilonBits57B770)),
	)

	velocityXForParallel := hooks.loadVelocityX(velocity)
	velocityYForParallel := hooks.loadVelocityY(velocity)
	parallel := deathBallDiv64(
		deathBallAdd64(
			deathBallMul64(float64(velocityXForParallel), nx),
			deathBallMul64(float64(velocityYForParallel), ny),
		),
		denominator,
	)
	negativeY := -ny
	velocityYForOrthogonal := hooks.loadVelocityY(velocity)
	velocityXForOrthogonal := hooks.loadVelocityX(velocity)
	orthogonal := float32(deathBallDiv64(
		deathBallAdd64(
			deathBallMul64(nx, float64(velocityYForOrthogonal)),
			deathBallMul64(negativeY, float64(velocityXForOrthogonal)),
		),
		denominator,
	))
	parallelX := float32(deathBallDiv64(deathBallMul64(parallel, nx), denominator))
	parallelY := float32(deathBallDiv64(deathBallMul64(parallel, ny), denominator))
	orthogonalX := float32(deathBallDiv64(
		deathBallMul64(float64(orthogonal), negativeY),
		denominator,
	))
	orthogonalY := deathBallDiv64(
		deathBallMul64(float64(orthogonal), nx),
		denominator,
	)
	resultX := float32(deathBallSub64(float64(orthogonalX), float64(parallelX)))
	resultY := float32(deathBallSub64(orthogonalY, float64(parallelY)))
	hooks.storeVelocityX(velocity, resultX)
	hooks.storeVelocityY(velocity, resultY)
}

func deathBallDoorReflect57B770(velocityX, velocityY, normalX, normalY float32) (float32, float32) {
	type velocity57B770 struct {
		x float32
		y float32
	}
	velocity := &velocity57B770{x: velocityX, y: velocityY}
	deathBallDoorReflectCore57B770(velocity, normalX, normalY, deathBallDoorReflectHooks57B770[*velocity57B770]{
		loadVelocityX: func(value *velocity57B770) float32 { return value.x },
		loadVelocityY: func(value *velocity57B770) float32 { return value.y },
		storeVelocityX: func(value *velocity57B770, result float32) {
			value.x = result
		},
		storeVelocityY: func(value *velocity57B770, result float32) {
			value.y = result
		},
	})
	return velocity.x, velocity.y
}

type deathBallDirection4E9E90 struct {
	x int32
	y int32
}

// Exact signed int32 pairs from GAME.EXE 005B6E58..005B6F57.
var deathBallDoorDirections4E9E90 = [...]deathBallDirection4E9E90{
	{x: -23, y: -23}, {x: -18, y: -27}, {x: -12, y: -30}, {x: -6, y: -31},
	{x: 0, y: -32}, {x: 6, y: -31}, {x: 12, y: -30}, {x: 18, y: -27},
	{x: 23, y: -23}, {x: 27, y: -18}, {x: 30, y: -12}, {x: 31, y: -6},
	{x: 32, y: 0}, {x: 31, y: 6}, {x: 30, y: 12}, {x: 27, y: 18},
	{x: 23, y: 23}, {x: 18, y: 27}, {x: 12, y: 30}, {x: 6, y: 31},
	{x: 0, y: 32}, {x: -6, y: 31}, {x: -12, y: 30}, {x: -18, y: 27},
	{x: -23, y: 23}, {x: -27, y: 18}, {x: -30, y: 12}, {x: -31, y: 6},
	{x: -32, y: 0}, {x: -31, y: -6}, {x: -30, y: -12}, {x: -27, y: -18},
}

func deathBallDoorDirection4E9E90(direction int32) deathBallDirection4E9E90 {
	if direction < 0 || direction >= deathBallDoorDirectionCount4E9E90 {
		panic("DeathBall Door direction is outside the original 0..31 table")
	}
	return deathBallDoorDirections4E9E90[direction]
}

// DoorDirectionX exposes the native signed X component shared by Door
// transfer and DeathBall collision handling.
func DoorDirectionX(direction int32) int32 {
	return deathBallDoorDirection4E9E90(direction).x
}

// DoorDirectionY exposes the native signed Y component shared by Door
// transfer and DeathBall collision handling.
func DoorDirectionY(direction int32) int32 {
	return deathBallDoorDirection4E9E90(direction).y
}
