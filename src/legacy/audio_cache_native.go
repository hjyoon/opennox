package legacy

/*
#include <stdlib.h>
#include "audio_source_native.h"
*/
import "C"

import (
	"github.com/opennox/opennox/v1/internal/binfile"
	"unsafe"
)

// AudioBankFileNative resolves the existing native file handle without
// reconstructing it from the packed bank's old DWORD offsets.
func AudioBankFileNative(f *FILE) *binfile.File { return fileByHandle(f) }

// The file/index owner lives in the client package. These native services
// replace the old casts of audioStructXxx's pointer fields into DWORD slots.
var (
	AudioBankRecordNative    func(unsafe.Pointer, int32) unsafe.Pointer
	AudioBankLookupNative    func(unsafe.Pointer, string) uint32
	AudioBankSelectNative    func(unsafe.Pointer, int32) int32
	AudioBankReadNative      func(unsafe.Pointer, *byte, int32) int32
	AudioBankCloseReadNative func(unsafe.Pointer) unsafe.Pointer
	AudioBankRemainingNative func(unsafe.Pointer) uint32
)

type audioCacheNative struct {
	Bank, Chunks, Samples unsafe.Pointer
	LRU                   audioNativeLink
	ChunkSize             int32
}
type audioCachedSampleNative struct {
	Link   audioNativeLink
	Refs   uint32
	ID     int32
	Loaded uint32
	Buffer AudioBufferNative
	Cache  *audioCacheNative
	Format [7]uint32
}

//export sub_4BD280
func sub_4BD280(count, size C.int) unsafe.Pointer {
	// Invalid sizes used to underflow the PE32 allocation/initialization loop.
	if count <= 0 || size < 0 {
		return nil
	}
	word := unsafe.Sizeof(uintptr(0))
	stride := (uintptr(size) + word + word - 1) &^ (word - 1)
	if uintptr(count) > (^uintptr(0)-word)/stride {
		return nil
	}
	p := audioNativeAlloc(word + uintptr(count)*stride)
	if p == nil {
		return nil
	}
	first := unsafe.Add(p, word)
	*(*unsafe.Pointer)(p) = first
	for i := uintptr(0); i+1 < uintptr(count); i++ {
		*(*unsafe.Pointer)(unsafe.Add(first, i*stride)) = unsafe.Add(first, (i+1)*stride)
	}
	return p
}

//export sub_4BD2D0
func sub_4BD2D0(p unsafe.Pointer) { audioNativeFree(p) }

//export sub_4BD2E0
func sub_4BD2E0(p unsafe.Pointer) unsafe.Pointer {
	item := *(*unsafe.Pointer)(p)
	if item != nil {
		*(*unsafe.Pointer)(p) = *(*unsafe.Pointer)(item)
		return unsafe.Add(item, unsafe.Sizeof(uintptr(0)))
	}
	return nil
}

//export sub_4BD300
func sub_4BD300(p, item unsafe.Pointer) unsafe.Pointer {
	header := unsafe.Add(item, -int(unsafe.Sizeof(uintptr(0))))
	*(*unsafe.Pointer)(header) = *(*unsafe.Pointer)(p)
	*(*unsafe.Pointer)(p) = header
	return header
}

//export sub_4BD340
func sub_4BD340(bank unsafe.Pointer, bytes, count, size C.int) unsafe.Pointer {
	if size <= 0 || bytes <= 0 || count <= 0 || int64(size)+int64(unsafe.Sizeof(AudioChunkNative{})) > 0x7fffffff {
		return nil
	}
	v := (*audioCacheNative)(audioNativeAlloc(unsafe.Sizeof(audioCacheNative{})))
	if v == nil {
		return nil
	}
	v.Bank, v.ChunkSize = bank, int32(size)
	v.Chunks = sub_4BD280(bytes/(size+24), C.int(unsafe.Sizeof(AudioChunkNative{}))+size)
	v.Samples = sub_4BD280(count, C.int(unsafe.Sizeof(audioCachedSampleNative{})))
	v.LRU.init()
	if v.Chunks != nil && v.Samples != nil {
		return unsafe.Pointer(v)
	}
	sub_4BD3C0(unsafe.Pointer(v))
	return nil
}

//export sub_4BD3C0
func sub_4BD3C0(p unsafe.Pointer) {
	v := (*audioCacheNative)(p)
	for n := v.LRU.first(); n != nil; n = v.LRU.first() {
		sub_4BD690(unsafe.Pointer(n))
	}
	sub_4BD2D0(v.Chunks)
	sub_4BD2D0(v.Samples)
	audioNativeFree(p)
}

//export sub_4BD420
func sub_4BD420(p unsafe.Pointer, id C.int) unsafe.Pointer {
	for n := (*audioCacheNative)(p).LRU.first(); n != nil; n = n.next() {
		v := (*audioCachedSampleNative)(unsafe.Pointer(n))
		if v.ID == int32(id) && v.Loaded != 0 {
			return unsafe.Pointer(v)
		}
	}
	return nil
}

//export sub_4BD470
func sub_4BD470(p unsafe.Pointer, id C.int) unsafe.Pointer {
	cache := (*audioCacheNative)(p)
	if hit := sub_4BD420(p, id); hit != nil {
		v := (*audioCachedSampleNative)(hit)
		v.Link.remove()
		cache.LRU.prepend(&v.Link)
		return hit
	}
	if sub_486B60(cache.Bank, id) == 0 {
		return nil
	}
	item := sub_4BD2E0(cache.Samples)
	if item == nil {
		sub_4BD600(p)
		item = sub_4BD2E0(cache.Samples)
		if item == nil {
			sub_486E00(cache.Bank)
			return nil
		}
	}
	v := (*audioCachedSampleNative)(item)
	v.ID, v.Cache = int32(id), cache
	v.Link.init()
	v.Refs = 0
	sub_487C30(unsafe.Pointer(&v.Buffer))
	v.Buffer.Format = &v.Format
	remaining := AudioBankRemainingNative(cache.Bank)
	for remaining != 0 {
		n := uint32(cache.ChunkSize)
		if int32(n) > int32(remaining) {
			n = remaining
		}
		chunk := sub_4BD2E0(cache.Chunks)
		for chunk == nil && sub_4BD600(p) != 0 {
			chunk = sub_4BD2E0(cache.Chunks)
		}
		if chunk == nil {
			sub_4BD690(item)
			return nil
		}
		data := unsafe.Add(chunk, unsafe.Sizeof(AudioChunkNative{}))
		sub_487D30(chunk, data, C.int(n))
		sub_487C50(unsafe.Pointer(&v.Buffer), chunk)
		read := int32(sub_486DB0(cache.Bank, (*C.char)(data), C.int(n)))
		if uint32(read) != n {
			sub_4BD690(item)
			return nil
		}
		remaining -= uint32(read)
	}
	sub_486AA0(cache.Bank, id, (*C.uint32_t)(unsafe.Pointer(&v.Format[0])))
	cache.LRU.prepend(&v.Link)
	v.Loaded = 1
	sub_486E00(cache.Bank)
	return item
}

//export sub_4BD600
func sub_4BD600(p unsafe.Pointer) C.int {
	head := &(*audioCacheNative)(p).LRU
	for n := head.Prev; n != nil && n != head; n = n.Prev {
		if (*audioCachedSampleNative)(unsafe.Pointer(n)).Refs == 0 {
			sub_4BD690(unsafe.Pointer(n))
			return 1
		}
	}
	return 0
}

//export sub_4BD650
func sub_4BD650(p unsafe.Pointer) unsafe.Pointer {
	(*audioCachedSampleNative)(p).Refs++
	return p
}

//export sub_4BD660
func sub_4BD660(p unsafe.Pointer) C.int {
	v := (*audioCachedSampleNative)(p)
	v.Refs--
	result := int32(v.Refs)
	if result < 0 {
		v.Refs = 0
	}
	return C.int(result)
}

//export sub_4BD680
func sub_4BD680(p unsafe.Pointer) C.int { return C.int((*audioCachedSampleNative)(p).Refs) }

//export sub_4BD690
func sub_4BD690(p unsafe.Pointer) unsafe.Pointer {
	v := (*audioCachedSampleNative)(p)
	v.Link.remove()
	for n := v.Buffer.Chunks.first(); n != nil; n = v.Buffer.Chunks.first() {
		c := (*AudioChunkNative)(unsafe.Pointer(n))
		c.Link.remove()
		sub_487D60(unsafe.Pointer(c))
		sub_4BD300(v.Cache.Chunks, unsafe.Pointer(c))
	}
	return sub_4BD300(v.Cache.Samples, p)
}

//export sub_4BD710
func sub_4BD710(p unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(&(*audioCachedSampleNative)(p).Buffer)
}

//export sub_4866D0
func sub_4866D0(p unsafe.Pointer, index C.int) unsafe.Pointer {
	return AudioBankRecordNative(p, int32(index))
}

//export sub_486A10
func sub_486A10(p, name unsafe.Pointer) C.uint {
	return C.uint(AudioBankLookupNative(p, GoString((*C.char)(name))))
}

//export sub_486AA0
func sub_486AA0(p unsafe.Pointer, id C.int, out *C.uint32_t) C.int {
	record := (*[9]uint32)(sub_4866D0(p, id))
	v := (*[7]uint32)(unsafe.Pointer(out))
	v[0], v[2], v[3], v[6] = 4, record[6], (record[7]&1)+1, record[8]
	if record[7]&8 != 0 {
		v[1], v[4] = 2, 2
	} else {
		v[1], v[4] = 0, ((record[7]>>2)&1)+1
	}
	return C.int(v[4])
}

//export sub_486B60
func sub_486B60(p unsafe.Pointer, id C.int) C.int { return C.int(AudioBankSelectNative(p, int32(id))) }

//export sub_486DB0
func sub_486DB0(p unsafe.Pointer, data *C.char, size C.int) C.int {
	return C.int(AudioBankReadNative(p, (*byte)(unsafe.Pointer(data)), int32(size)))
}

//export sub_486E00
func sub_486E00(p unsafe.Pointer) unsafe.Pointer { return AudioBankCloseReadNative(p) }
