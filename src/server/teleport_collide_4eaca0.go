package server

const (
	teleportCollideDoorClass4EACA0 = uint32(0x00000080)
	teleportCollidePreFX4EACA0     = uint32(138)
	teleportCollidePostFX4EACA0    = uint32(137)
	teleportCollideSound4EACA0     = uint32(147)
)

type teleportCollideHooks4EACA0[O, D, P, T any] struct {
	loadCollideData   func(O) D
	loadClass         func(O) uint32
	cachePosition     func(O) P
	pointFX           func(uint32, P)
	audio             func(uint32, O)
	loadDestinationX  func(D) int32
	cacheDestination  func(O) T
	storeDestinationX func(O, float32)
	loadDestinationY  func(D) int32
	storeDestinationY func(O, float32)
	teleport          func(O, T)
}

// teleportCollide4EACA0 preserves GAME.EXE 004EACA0. The source collide-data
// pointer is cached before either target guard, while its two signed values
// stay live until their individual FILD instructions. The target position
// reference is cached once and reused by both point effects.
//
// On the accepted path, the pre-effect and sound precede both destination
// loads. Destination X is stored before Y is loaded, the teleport receives an
// alias of those two stored fields, and the post-effect and sound run even if
// the nested teleport gate elects not to move the target. Collision is never
// observed.
func teleportCollide4EACA0[O comparable, D, P, T, C any](
	source, target O,
	_ C,
	hooks teleportCollideHooks4EACA0[O, D, P, T],
) {
	data := hooks.loadCollideData(source)

	var zero O
	if target == zero {
		return
	}
	if hooks.loadClass(target)&teleportCollideDoorClass4EACA0 != 0 {
		return
	}

	position := hooks.cachePosition(target)
	hooks.pointFX(teleportCollidePreFX4EACA0, position)
	hooks.audio(teleportCollideSound4EACA0, target)

	x := hooks.loadDestinationX(data)
	destination := hooks.cacheDestination(target)
	hooks.storeDestinationX(target, float32(x))
	y := hooks.loadDestinationY(data)
	hooks.storeDestinationY(target, float32(y))
	hooks.teleport(target, destination)

	hooks.pointFX(teleportCollidePostFX4EACA0, position)
	hooks.audio(teleportCollideSound4EACA0, target)
}
