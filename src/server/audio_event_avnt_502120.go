package server

const audioEventAVNTDistanceScale502120 = int32(15)

// audioEventAVNTHooks502120 describes every observable read, lookup, skip,
// descriptor selection, and store made by GAME.EXE 00502120. S is the PE32
// sound index and D may be a native-width descriptor handle.
type audioEventAVNTHooks502120[S comparable, D any] struct {
	loadInitialized func() bool
	readName        func() (string, error)
	resolveSound    func(string) S
	zeroSound       S

	readU8  func() uint8
	readI16 func() int16
	skip    func(int)

	descriptor       func(S) D
	storeMaxDistance func(D, int32)
	storeFlags       func(D, uint32)
	storeField8      func(D, uint32)
	incrementField12 func(D)
	storeField16     func(D, uint32)
	storeField20     func(D, uint32)
}

// readAudioEventAVNT502120 preserves the original AVNT command parser. The
// initialized flag is cached before the name is read, and a zero sound only
// disables the five gated stores. Opcode 7 deliberately still selects the
// descriptor for the resolved sound (including sound zero) and increments its
// dword count once per nonempty payload. Both word operands are signed.
func readAudioEventAVNT502120[S comparable, D any](
	hooks audioEventAVNTHooks502120[S, D],
) bool {
	initialized := hooks.loadInitialized()
	name, err := hooks.readName()
	if err != nil {
		return false
	}
	sound := hooks.resolveSound(name)
	update := initialized && sound != hooks.zeroSound

	for {
		switch opcode := hooks.readU8(); opcode {
		case 0:
			return true
		case 1, 5:
			hooks.skip(1)
		case 2:
			value := hooks.readU8()
			if update {
				descriptor := hooks.descriptor(sound)
				hooks.storeField16(descriptor, uint32(value))
			}
		case 3:
			value := hooks.readU8()
			if update {
				descriptor := hooks.descriptor(sound)
				hooks.storeField8(descriptor, uint32(value))
			}
		case 4:
			value := hooks.readU8()
			if update {
				descriptor := hooks.descriptor(sound)
				hooks.storeField20(descriptor, uint32(value))
			}
		case 6:
			hooks.skip(2)
		case 7:
			for {
				length := hooks.readU8()
				if length == 0 {
					break
				}
				hooks.skip(int(length))
				descriptor := hooks.descriptor(sound)
				hooks.incrementField12(descriptor)
			}
		case 8:
			hooks.skip(8)
		case 9:
			value := hooks.readI16()
			if update {
				descriptor := hooks.descriptor(sound)
				hooks.storeMaxDistance(descriptor, int32(value)*audioEventAVNTDistanceScale502120)
			}
		case 10:
			value := hooks.readI16()
			if update {
				descriptor := hooks.descriptor(sound)
				hooks.storeFlags(descriptor, uint32(int32(value)))
			}
		default:
			return false
		}
	}
}
