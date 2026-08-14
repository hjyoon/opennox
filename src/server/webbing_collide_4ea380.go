package server

const (
	webbingCollideAudio4EA380       = uint32(351)
	webbingCollideUnitClass4EA380   = uint8(0x06)
	webbingCollidePlayerClass4EA380 = uint8(0x04)
	webbingCollideDamage4EA380      = int32(0)
	webbingCollideDamageType4EA380  = uint32(2)
	webbingCollideEnchant4EA380     = uint32(4)
	webbingCollideDuration4EA380    = uint32(4)
	webbingCollidePower4EA380       = uint32(3)
	webbingCollideMessage4EA380     = "objcoll.c:WebbingSlow"
)

type webbingCollideHooks4EA380[O comparable] struct {
	audio           func(uint32, O)
	delayedDelete   func(O)
	findParent      func(O) O
	targetDamage    func(O, O, O, int32, uint32) int32
	loadClassLow    func(O) uint8
	loadFPS         func() uint32
	applyEnchant    func(O, uint32, uint32, uint32)
	priorityMessage func(O, string)
}

// webbingCollide4EA380 preserves GAME.EXE 004EA380. A nil target returns
// before the source or collision argument is inspected. A live target receives
// source audio and deletion before its live Damage callback is invoked with a
// parent resolved after both effects. Successful damage reads the target class
// twice: once to decide whether to load FPS and apply Slow, and again after the
// enchant callback to decide whether to send the Player message.
func webbingCollide4EA380[O comparable, C any](
	source, target O,
	_ C,
	hooks webbingCollideHooks4EA380[O],
) {
	var zero O
	if target == zero {
		return
	}
	hooks.audio(webbingCollideAudio4EA380, source)
	hooks.delayedDelete(source)
	parent := hooks.findParent(source)
	if hooks.targetDamage(
		target,
		parent,
		source,
		webbingCollideDamage4EA380,
		webbingCollideDamageType4EA380,
	) == 0 {
		return
	}
	if hooks.loadClassLow(target)&webbingCollideUnitClass4EA380 != 0 {
		fps := hooks.loadFPS()
		hooks.applyEnchant(
			target,
			webbingCollideEnchant4EA380,
			fps*webbingCollideDuration4EA380,
			webbingCollidePower4EA380,
		)
	}
	if hooks.loadClassLow(target)&webbingCollidePlayerClass4EA380 != 0 {
		hooks.priorityMessage(target, webbingCollideMessage4EA380)
	}
}
