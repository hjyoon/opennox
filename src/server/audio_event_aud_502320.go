package server

const audioEventAUDDistanceScale502370 = int32(15)

// audioEventAUDListHooks502320 describes the signed count read and record
// calls made by GAME.EXE 00502320.
type audioEventAUDListHooks502320 struct {
	readCount  func() int32
	readRecord func() bool
}

// readAudioEventAUDList502320 accepts every nonpositive signed record count.
// A positive list succeeds only after every record reader call succeeds.
func readAudioEventAUDList502320(hooks audioEventAUDListHooks502320) bool {
	count := hooks.readCount()
	if count <= 0 {
		return true
	}
	for index := int32(0); index < count; index++ {
		if !hooks.readRecord() {
			return false
		}
	}
	return true
}

// audioEventAUDRecordHooks502370 describes every observable lookup, gate,
// read, skip, descriptor selection, and store made by GAME.EXE 00502370. S is
// the PE32 sound index and D may be a native-width descriptor handle.
type audioEventAUDRecordHooks502370[S comparable, D any] struct {
	readName        func() (string, error)
	resolveSound    func(string) S
	zeroSound       S
	loadInitialized func() bool

	readI16 func() int16
	readU8  func() uint8
	readI8  func() int8
	skip    func(int)

	descriptor       func(S) D
	storeMaxDistance func(D, int32)
	storeFlags       func(D, uint32)
	storeField8      func(D, uint32)
	incrementField12 func(D)
	storeField16     func(D, uint32)
	storeField20     func(D, uint32)
}

// readAudioEventAUDRecord502370 preserves the original AUD record reader.
// Initialization is sampled after sound resolution and only for a nonzero
// sound. The two words and every sample length retain their signed PE32
// interpretation; a negative sample length therefore produces a negative
// skip just as the original pointer arithmetic did.
func readAudioEventAUDRecord502370[S comparable, D any](
	hooks audioEventAUDRecordHooks502370[S, D],
) bool {
	name, err := hooks.readName()
	if err != nil {
		return false
	}
	sound := hooks.resolveSound(name)
	if sound == hooks.zeroSound || !hooks.loadInitialized() {
		hooks.skip(9)
		for {
			length := hooks.readI8()
			if length == 0 {
				return true
			}
			hooks.skip(int(length))
		}
	}

	flags := hooks.readI16()
	field8 := hooks.readU8()
	maxDistance := hooks.readI16()
	field20 := hooks.readU8()
	descriptor := hooks.descriptor(sound)
	if maxDistance > 0 {
		hooks.storeMaxDistance(descriptor, int32(maxDistance)*audioEventAUDDistanceScale502370)
	}
	hooks.storeFlags(descriptor, uint32(int32(flags)))
	hooks.storeField8(descriptor, uint32(field8))
	hooks.storeField20(descriptor, uint32(field20))
	hooks.skip(3)

	for {
		length := hooks.readI8()
		if length == 0 {
			break
		}
		hooks.skip(int(length))
		hooks.incrementField12(descriptor)
	}
	hooks.storeField16(descriptor, 2)
	return true
}
