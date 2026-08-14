package server

const (
	chestCollideQuestFlag4E9C40      = uint32(0x1000)
	chestCollidePlayerClass4E9C40    = uint8(0x04)
	chestCollideKeyClass4E9C40       = uint8(0x40)
	chestCollideDestroyedFlag4E9C40  = uint32(0x8000)
	chestCollideLockedSubclass4E9C40 = uint32(0x0f00)
	chestCollideUnlockAudio4E9C40    = uint32(234)
	chestCollideLockedAudio4E9C40    = uint32(1012)
	chestCollideFeedbackDelay4E9C40  = uint64(1500)
	chestCollideSilverKeyName4E9C40  = "SilverKey"
	chestCollideLockedMessage4E9C40  = "objcoll.c:ChestLockedSilver"
)

type chestCollideHooks4E9C40[O comparable, F comparable] struct {
	loadClassLow       func(O) uint8
	loadFlags          func(O) uint32
	gameFlagsCheck     func(uint32) int32
	loadSubclass       func(O) uint32
	firstItem          func(O) O
	loadTypeName       func(O) string
	nextItem           func(O) O
	delayedDelete      func(O)
	audio              func(uint32, O)
	ticks              func() uint64
	loadFeedbackTicks  func() uint64
	priorityMessage    func(O, string)
	storeFeedbackTicks func(uint64)
	loadDeath          func(O) F
	callDeath          func(F, O)
	chestOpen          func(O, O)
	dropAllItems       func(O)
}

// chestSilverKeyNameMatches4E9C40 models the ten-byte rep cmpsb against
// "SilverKey\0". Object type IDs normally omit the terminator; an embedded
// NUL lets tests model the original byte sequence and trailing storage.
func chestSilverKeyNameMatches4E9C40(name string) bool {
	const expected = chestCollideSilverKeyName4E9C40
	if len(name) < len(expected) || name[:len(expected)] != expected {
		return false
	}
	return len(name) == len(expected) || name[len(expected)] == 0
}

// chestCollide4E9C40 preserves GAME.EXE 004E9C40. The target and source
// gates precede all inventory reads. A matching SilverKey is deleted and its
// sound is emitted before the live Death callback is loaded. Failed unlocks
// use the shared unsigned 64-bit feedback timestamp and never enter the open
// path. The registered third collision argument is not read.
func chestCollide4E9C40[O comparable, F comparable](
	source, target O,
	collision any,
	hooks chestCollideHooks4E9C40[O, F],
) {
	_ = collision
	var zeroObject O
	if target == zeroObject {
		return
	}
	if hooks.loadClassLow(target)&chestCollidePlayerClass4E9C40 == 0 {
		return
	}
	if hooks.loadFlags(source)&chestCollideDestroyedFlag4E9C40 != 0 {
		return
	}

	if hooks.gameFlagsCheck(chestCollideQuestFlag4E9C40) != 0 &&
		hooks.loadSubclass(source)&chestCollideLockedSubclass4E9C40 != 0 {
		unlocked := false
		item := hooks.firstItem(target)
		for item != zeroObject {
			if hooks.loadClassLow(item)&chestCollideKeyClass4E9C40 != 0 {
				if chestSilverKeyNameMatches4E9C40(hooks.loadTypeName(item)) {
					hooks.delayedDelete(item)
					hooks.audio(chestCollideUnlockAudio4E9C40, source)
					unlocked = true
					break
				}
			}
			item = hooks.nextItem(item)
		}
		if !unlocked {
			now := hooks.ticks()
			last := hooks.loadFeedbackTicks()
			if now-last > chestCollideFeedbackDelay4E9C40 {
				hooks.audio(chestCollideLockedAudio4E9C40, source)
				hooks.priorityMessage(target, chestCollideLockedMessage4E9C40)
				hooks.storeFeedbackTicks(hooks.ticks())
			}
			return
		}
	}

	death := hooks.loadDeath(source)
	var zeroDeath F
	if death != zeroDeath {
		hooks.callDeath(death, source)
	}
	hooks.chestOpen(source, target)
	hooks.dropAllItems(source)
}
