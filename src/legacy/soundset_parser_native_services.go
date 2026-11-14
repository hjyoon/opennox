package legacy

/*
#include <stdio.h>
*/
import "C"
import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// Serial service seam for the independently captured original token-level
// contract. NameStored observes Go ownership; it is not a native malloc call.
type soundSetParserObserver424170 struct {
	Open       func(unsafe.Pointer, int32) unsafe.Pointer
	Key        func(unsafe.Pointer, int32) int32
	Token      func(unsafe.Pointer, unsafe.Pointer) int32
	Sound      func(unsafe.Pointer) uint32
	Allocate   func(uintptr) unsafe.Pointer
	NameStored func(string)
	Close      func(unsafe.Pointer) int32
}

var soundSetParserObserved424170 *soundSetParserObserver424170

func soundSetParserObserve424170(o *soundSetParserObserver424170) func() {
	previous := soundSetParserObserved424170
	soundSetParserObserved424170 = o
	return func() {
		if soundSetParserObserved424170 != o {
			panic("sound-set parser observer restore order")
		}
		soundSetParserObserved424170 = previous
	}
}

func soundSetParserOpen424170(path unsafe.Pointer) unsafe.Pointer {
	if o := soundSetParserObserved424170; o != nil && o.Open != nil {
		return o.Open(path, 0)
	}
	return unsafe.Pointer(nox_binfile_open_408CC0((*C.char)(path), 0))
}

func soundSetParserKey424170(file unsafe.Pointer) int32 {
	if o := soundSetParserObserved424170; o != nil && o.Key != nil {
		return o.Key(file, 5)
	}
	return nox_binfile_cryptSet_408D40((*C.FILE)(file), 5)
}

func soundSetParserToken424170(file, out unsafe.Pointer) int32 {
	if o := soundSetParserObserved424170; o != nil && o.Token != nil {
		return o.Token(file, out)
	}
	return headSettingsParse409470(file, out)
}

func soundSetParserSound424170(name unsafe.Pointer) uint32 {
	if o := soundSetParserObserved424170; o != nil && o.Sound != nil {
		return o.Sound(name)
	}
	return uint32(sound.ByName(alloc.GoString((*byte)(name))))
}

func soundSetParserAllocate424170() unsafe.Pointer {
	if o := soundSetParserObserved424170; o != nil && o.Allocate != nil {
		return o.Allocate(soundSetRecordSize)
	}
	data, _ := alloc.Malloc(soundSetRecordSize)
	return data
}

func soundSetParserNameStored424170(name string) {
	if o := soundSetParserObserved424170; o != nil && o.NameStored != nil {
		o.NameStored(name)
	}
}

func soundSetParserClose424170(file unsafe.Pointer) {
	if o := soundSetParserObserved424170; o != nil && o.Close != nil {
		o.Close(file)
		return
	}
	nox_binfile_close_408D90((*C.FILE)(file))
}
