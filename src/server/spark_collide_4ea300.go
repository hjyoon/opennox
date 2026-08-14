package server

const (
	sparkCollideNoEffectKind4EA300   = uint32(4)
	sparkCollideWebbingKind4EA300    = uint32(5)
	sparkCollideWebbingAudio4EA300   = uint32(351)
	sparkCollidePlayerClassBit4EA300 = uint8(0x04)
	sparkCollideWebbingTimer4EA300   = uint16(1000)
	sparkCollideWebbingMessage4EA300 = "objcoll.c:WebbingSlow"
)

type sparkCollideHooks4EA300[O comparable, C, D any] struct {
	loadUpdateData  func(O) D
	loadKind        func(D) uint32
	wallReflect     func(O, O, C)
	audio           func(uint32, O)
	delayedDelete   func(O)
	loadSlowCount   func(O) uint8
	loadClassLow    func(O) uint8
	storeSlowCount  func(O, uint8)
	storeSlowTimer  func(O, uint16)
	priorityMessage func(O, string)
}

// sparkCollide4EA300 preserves GAME.EXE 004EA300. The source update pointer
// and kind are read before either callback argument is inspected. Kind four
// returns, kind five applies the webbing counters to a non-nil target, and all
// other values forward the original three arguments to WallReflectCollide.
// The webbing path loads its counter and class only after audio and deletion;
// both values are cached before the counter and timer stores.
func sparkCollide4EA300[O comparable, C, D any](
	source, target O,
	collision C,
	hooks sparkCollideHooks4EA300[O, C, D],
) {
	update := hooks.loadUpdateData(source)
	kind := hooks.loadKind(update)
	if kind == sparkCollideNoEffectKind4EA300 {
		return
	}
	if kind != sparkCollideWebbingKind4EA300 {
		hooks.wallReflect(source, target, collision)
		return
	}

	var zero O
	if target == zero {
		return
	}
	hooks.audio(sparkCollideWebbingAudio4EA300, source)
	hooks.delayedDelete(source)
	count := hooks.loadSlowCount(target)
	class := hooks.loadClassLow(target)
	hooks.storeSlowCount(target, count+1)
	hooks.storeSlowTimer(target, sparkCollideWebbingTimer4EA300)
	if class&sparkCollidePlayerClassBit4EA300 != 0 {
		hooks.priorityMessage(target, sparkCollideWebbingMessage4EA300)
	}
}
