package legacy

/*
#include "game2_audio_native.h"
int sub_453660(int);
int sub_453690(int);
*/
import "C"
import (
	"github.com/opennox/opennox/v1/internal/binfile"
	"unsafe"
)

func game2AudioRandomC(lo, hi int32, line int) int32 {
	return nox_common_randomIntMinMax_415FF0(lo, hi, internCStr("C:\\NoxPost\\src\\client\\Audio\\AudEvent.c"), int32(line))
}
func game2AudioMaskC(mask uint32, weapon bool) int32 {
	if weapon {
		return int32(C.sub_453690(C.int(mask)))
	}
	return int32(C.sub_453660(C.int(mask)))
}

// Every public entry serializes mutations. Audio-thread callbacks only enqueue
// completion tokens and never re-enter this lock while AIL End/Init is running.
//
//export sub_451850
func sub_451850(context, bank unsafe.Pointer) C.int {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return C.int(game2AudioInit451850(context, bank))
}

//export sub_451920
func sub_451920(p unsafe.Pointer) C.int {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return C.int(game2AudioReset451920(p))
}

//export sub_451970
func sub_451970() { game2AudioMu.Lock(); defer game2AudioMu.Unlock(); game2AudioShutdown451970() }

//export sub_4519C0
func sub_4519C0() { game2AudioMu.Lock(); defer game2AudioMu.Unlock(); game2AudioFrame4519C0() }

//export sub_451BE0
func sub_451BE0(p unsafe.Pointer) C.int {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return C.int(game2AudioRank451BE0((*game2AudioEvent)(p)))
}

//export sub_451CA0
func sub_451CA0(p unsafe.Pointer) unsafe.Pointer {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return game2AudioStart451CA0((*game2AudioEvent)(p))
}

//export sub_451F30
func sub_451F30(p unsafe.Pointer, index C.int) C.int {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return C.int(game2AudioLoadSample451F30((*game2AudioEvent)(p), int32(index)))
}

//export sub_451F90
func sub_451F90(p unsafe.Pointer) C.int {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return C.int(game2AudioReleaseSamples451F90((*game2AudioEvent)(p)))
}

//export sub_451FE0
func sub_451FE0(p unsafe.Pointer) unsafe.Pointer {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return game2AudioRelease451FE0((*game2AudioEvent)(p))
}

//export sub_452010
func sub_452010() C.int {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return C.int(game2AudioBucketReset452010())
}

//export sub_452050
func sub_452050(p unsafe.Pointer) {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	game2AudioBucket452050((*game2AudioEvent)(p))
}

//export sub_452120
func sub_452120(p unsafe.Pointer) C.int {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return C.int(game2AudioSteal452120((*game2AudioEvent)(p)))
}

//export sub_452190
func sub_452190(p unsafe.Pointer) {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	game2AudioBucketRemove452190(game2AudioDef(p))
}

//export sub_4521A0
func sub_4521A0(limit C.int) unsafe.Pointer {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	d := game2AudioFindBucket4521A0(int32(limit))
	if d == nil {
		return nil
	}
	return d.Pointer
}

//export sub_4521F0
func sub_4521F0() C.uintptr_t {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return C.uintptr_t(game2AudioClear4521F0())
}

//export sub_452230
func sub_452230() unsafe.Pointer {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return game2AudioPurge452230()
}

//export nox_xxx_draw_452270
func nox_xxx_draw_452270(id C.int) unsafe.Pointer {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return game2AudioLookup452270(int32(id))
}

//export sub_4522A0
func sub_4522A0(p unsafe.Pointer) C.int {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return C.int(game2AudioDuration4522A0(p))
}

//export nox_xxx_draw_452300
func nox_xxx_draw_452300(p unsafe.Pointer) unsafe.Pointer {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return unsafe.Pointer(game2AudioAllocate452300(p))
}

//export sub_4523D0
func sub_4523D0(p unsafe.Pointer) C.int {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return C.int(game2AudioStop4523D0((*game2AudioEvent)(p)))
}

//export sub_452410
func sub_452410(p unsafe.Pointer) unsafe.Pointer {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return game2AudioDetach452410((*game2AudioEvent)(p))
}

//export sub_452490
func sub_452490(p unsafe.Pointer) C.int {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return C.int(game2AudioPlay452490((*game2AudioEvent)(p)))
}

//export sub_452510
func sub_452510(p unsafe.Pointer) {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	game2AudioAdvance452510((*game2AudioEvent)(p))
}

//export sub_452690
func sub_452690(p unsafe.Pointer, delta C.int64_t, next C.int) C.int64_t {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return C.int64_t(game2AudioDelay452690((*game2AudioEvent)(p), int64(delta), int32(next)))
}

//export sub_4526D0
func sub_4526D0(p unsafe.Pointer) C.int {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return C.int(game2AudioEnd4526D0((*game2AudioVoice)(p)))
}

//export sub_4526F0
func sub_4526F0(p unsafe.Pointer) C.int {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return C.int(game2AudioDone4526F0((*game2AudioVoice)(p)))
}

//export sub_452810
func sub_452810(priority C.int, flags C.char) unsafe.Pointer {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return unsafe.Pointer(game2AudioAcquire452810(int32(priority), int8(flags)))
}

//export nox_thing_read_AVNT_452890
func nox_thing_read_AVNT_452890(f *C.nox_memfile, scratch unsafe.Pointer) C.int {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return C.int(game2AudioParseAVNT452890((*binfile.MemFile)(unsafe.Pointer(f)), scratch))
}

//export nox_thing_skip_AVNT_452B00
func nox_thing_skip_AVNT_452B00(f *C.nox_memfile) C.int {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return C.int(game2AudioSkip452B00((*binfile.MemFile)(unsafe.Pointer(f))))
}

//export nox_thing_skip_AVNT_inner_452B30
func nox_thing_skip_AVNT_inner_452B30(f *C.nox_memfile) C.int {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return C.int(game2AudioSkip452B30((*binfile.MemFile)(unsafe.Pointer(f))))
}

//export sub_452BD0
func sub_452BD0(f *C.nox_memfile, scratch *C.char) C.int {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return C.int(game2AudioParseAUD452BD0((*binfile.MemFile)(unsafe.Pointer(f)), unsafe.Pointer(scratch)))
}

//export nox_xxx_clientPlaySoundSpecial_452D80
func nox_xxx_clientPlaySoundSpecial_452D80(id, volume C.int) {
	nox_xxx_clientPlaySoundSpecial_native_452D80(id, volume)
}

//export sub_452DC0
func sub_452DC0(id, volume, pan C.int) {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	game2AudioPlayAt452DC0(int32(id), int32(volume), int32(pan), 0)
}

//export sub_452E10
func sub_452E10(id, volume, pan C.int) {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	game2AudioPlayAt452DC0(int32(id), int32(volume), int32(pan), 2)
}

//export sub_452E90
func sub_452E90(handle, p unsafe.Pointer) unsafe.Pointer {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return game2AudioSetHandle452E90((*game2AudioHandle)(handle), (*game2AudioEvent)(p))
}

//export sub_452EB0
func sub_452EB0(handle unsafe.Pointer) unsafe.Pointer {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return unsafe.Pointer(game2AudioGetHandle452EB0((*game2AudioHandle)(handle)))
}

//export sub_452EE0
func sub_452EE0(p unsafe.Pointer, v C.int) C.int {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return C.int(game2AudioVolume452EE0((*game2AudioEvent)(p), int32(v)))
}

//export sub_452F10
func sub_452F10(p unsafe.Pointer, v C.int) C.uint {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return C.uint(game2AudioVolume452F10((*game2AudioEvent)(p), int32(v)))
}

//export sub_452F50
func sub_452F50(p unsafe.Pointer, v C.int) C.int {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return C.int(game2AudioVolume452F50((*game2AudioEvent)(p), int32(v)))
}

//export sub_452F80
func sub_452F80(p unsafe.Pointer, v C.int) unsafe.Pointer {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return game2AudioPan452F80((*game2AudioEvent)(p), int32(v))
}

//export sub_452FA0
func sub_452FA0(v C.int) C.int { return C.int(game2AudioPan452FA0(int32(v))) }

//export sub_452FE0
func sub_452FE0(p unsafe.Pointer, v C.int) C.int {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return C.int(game2AudioPan452FE0((*game2AudioEvent)(p), int32(v)))
}

//export sub_453050
func sub_453050() { game2AudioMu.Lock(); defer game2AudioMu.Unlock(); game2AudioSetFXFlag(0) }

//export nox_xxx____setargv_9_453060
func nox_xxx____setargv_9_453060() {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	game2AudioSetFXFlag(1)
}

//export sub_453070
func sub_453070() C.int {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return C.int(game2AudioFXFlag())
}

//export sub_453080
func sub_453080(bit C.char) C.int { return C.int(game2AudioMask453080(int8(bit))) }

//export sub_451CF0
func sub_451CF0(p unsafe.Pointer) unsafe.Pointer {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return game2AudioNext451CF0((*game2AudioEvent)(p))
}

//export sub_451DC0
func sub_451DC0(p unsafe.Pointer) C.int {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return C.int(game2AudioLoad451DC0((*game2AudioEvent)(p)))
}

//export sub_451E80
func sub_451E80(p unsafe.Pointer) C.int {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return C.int(game2AudioShuffle451E80((*game2AudioEvent)(p)))
}

//export sub_452580
func sub_452580(p unsafe.Pointer) C.int {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return C.int(game2AudioSetup452580((*game2AudioEvent)(p)))
}

//export sub_452770
func sub_452770(p unsafe.Pointer) C.int {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	return C.int(game2AudioNeedBuffer452770((*game2AudioVoice)(p)))
}
