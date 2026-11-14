package legacy

/*
#include <stdlib.h>
#include "audio_source_native.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/legacy/timer"
)

// These private layouts replace the packed PE32 owners as a unit. In
// particular Cursor and Available no longer overlap at old offsets 296/300.
// All stored pointers refer to C allocations, interned data or C callbacks.
type audioNativeLink struct{ Next, Prev, Head *audioNativeLink }

// Synchronous fixture ownership; ordinary products retain calloc/free.
var audioNativeAllocationService func(uintptr) unsafe.Pointer
var audioNativeReleaseService func(unsafe.Pointer)

func audioNativeAlloc(n uintptr) unsafe.Pointer {
	if audioNativeAllocationService != nil {
		return audioNativeAllocationService(n)
	}
	return C.calloc(1, C.size_t(n))
}
func audioNativeFree(p unsafe.Pointer) {
	if p == nil {
		return
	}
	if audioNativeReleaseService != nil {
		audioNativeReleaseService(p)
	} else {
		C.free(p)
	}
}

func (n *audioNativeLink) init() { n.Next, n.Prev, n.Head = n, n, n }
func (n *audioNativeLink) first() *audioNativeLink {
	if n.Next == n || n.Next == nil {
		return nil
	}
	return n.Next
}
func (n *audioNativeLink) next() *audioNativeLink {
	if n.Next == n.Head || n.Next == nil {
		return nil
	}
	return n.Next
}
func (n *audioNativeLink) remove() {
	if n.Next != nil && n.Prev != nil {
		n.Next.Prev, n.Prev.Next = n.Prev, n.Next
	}
	n.init()
}
func (n *audioNativeLink) append(v *audioNativeLink) {
	v.Next, v.Prev, v.Head = n, n.Prev, n
	n.Prev.Next, n.Prev = v, v
}
func (n *audioNativeLink) prepend(v *audioNativeLink) {
	v.Next, v.Prev, v.Head = n.Next, n, n
	n.Next.Prev, n.Next = v, v
}

type AudioChunkNative struct {
	Link   audioNativeLink
	Data   *byte
	Size   uint32
	Buffer *AudioBufferNative
}
type AudioBufferNative struct {
	Data    *byte
	Size    uint32
	Chunks  audioNativeLink
	Format  *[7]uint32
	Field24 uint32
}
type AudioMethodsNative struct {
	Reserved, Create, Destroy, Play, Stop   unsafe.Pointer
	Pause, Resume, Status, Refresh, Restart unsafe.Pointer
}
type AudioClassNative struct {
	Name                                                     unsafe.Pointer
	Field4, Flags, State                                     uint32
	Reserved, Create, Destroy, ContextCreate, ContextDestroy unsafe.Pointer
	Methods                                                  *AudioMethodsNative
}
type AudioDeviceNative struct {
	Link         audioNativeLink
	Class        *AudioClassNative
	Count, Limit uint32
	Contexts     [16]*AudioContextNative
}
type AudioContextNative struct {
	Link                        audioNativeLink
	Kind                        int32
	Users                       uint32
	Device                      *AudioDeviceNative
	Index                       int32
	Field28                     uint32
	Format32, Format60          [7]uint32
	Group                       timer.TimerGroup
	Master                      *timer.TimerGroup
	Priority, Count, Limit      int32
	Voices                      audioNativeLink
	Busy                        int32
	Update                      unsafe.Pointer
	Field220                    uint32
	Interval, Delta, Peak, Last uint64
	Methods                     *AudioMethodsNative
	Field260                    uint32
}
type AudioVoiceNative struct {
	Link                           audioNativeLink
	Kind                           int32
	Group                          timer.TimerGroup
	OwnerGroup, Master             *timer.TimerGroup
	Priority                       int32
	Flags                          uint32
	Repeat                         int32
	Context                        *AudioContextNative
	OnRefill, OnRestart, OnEnd     unsafe.Pointer
	OnStop, UserData               unsafe.Pointer
	Reserved                       [4]uint32
	Methods                        *AudioMethodsNative
	Mixed                          timer.TimerGroup
	Sample                         *AudioSample
	Refill, Restart, Finish        unsafe.Pointer
	Buffer                         *AudioBufferNative
	Chunk                          *AudioChunkNative
	Cursor                         *byte
	Available, Remaining, Field308 uint32
}
type audioGlobalNative struct {
	Devices, Contexts audioNativeLink
	Busy              int32
	Group             timer.TimerGroup
	GroupInitialized  uint32
}

func audioNativeCall(f unsafe.Pointer, p unsafe.Pointer) int32 {
	if f == nil {
		return 0
	}
	return int32(ccall.CallIntPtr(f, p))
}
func audioNativeVoid(f unsafe.Pointer, p unsafe.Pointer) {
	if f != nil {
		ccall.CallVoidPtr(f, p)
	}
}

// The synchronous seams substitute only the Miles services at the original
// boundary; production uses the real AIL handle and complete native pointers.
type audioSourceServices struct {
	UserData   func(ail.Sample) *AudioSample
	Ready      func(ail.Sample) int
	Load       func(ail.Sample, uint32, unsafe.Pointer, uint32)
	Dependency func(string, *AudioVoiceNative)
}

var audioSourceTestServices *audioSourceServices

func audioSourceDependency(name string, v *AudioVoiceNative) {
	if svc := audioSourceTestServices; svc != nil && svc.Dependency != nil {
		svc.Dependency(name, v)
	}
}

func audioSourceEnd(s *AudioSample) {
	if s.Flag7 == 0 {
		v := (*AudioVoiceNative)(s.Field1)
		audioNativeCall(v.Finish, unsafe.Pointer(v))
		// The callback can write Flag7; the original overwrites it afterwards.
		s.Flag7 = 1
	}
}

func AudioSampleFinishNative(s *AudioSample) { audioSourceEnd(s) }

//export sub_43EDB0
func sub_43EDB0(handle C.HSAMPLE) {
	h := ail.Sample(unsafe.Pointer(handle))
	var s *AudioSample
	if svc := audioSourceTestServices; svc != nil && svc.UserData != nil {
		s = svc.UserData(h)
	} else {
		s = h.UserData().(*AudioSample)
	}
	audioSourceEnd(s)
}

//export sub_43EE00
func sub_43EE00(p unsafe.Pointer) C.int {
	s := (*AudioSample)(p)
	ready := func() int {
		if svc := audioSourceTestServices; svc != nil && svc.Ready != nil {
			return svc.Ready(s.Smp)
		}
		return s.Smp.BufferReady()
	}
	for index := ready(); index != -1; index = ready() {
		v := (*AudioVoiceNative)(s.Field1)
		dest := s.Data1
		if index != 0 {
			dest = s.Data2
		}
		var start, deferred, deferredDest *byte
		var count, deferredSize uint32
		if s.Field3 == 0 {
			for int32(count) < 0x4000 {
				n := v.Available
				if n == 0 {
					audioNativeCall(v.Refill, unsafe.Pointer(v))
					n = v.Available
					if n == 0 {
						audioNativeCall(v.Restart, unsafe.Pointer(v))
						n = v.Available
						if n == 0 {
							s.Field3 = 1
							break
						}
					}
				}
				if count == 0 {
					start = v.Cursor
				}
				if count != 0 || int32(n) < 0x4000 || deferred != nil {
					if int32(n+count) > 0x4000 {
						n = 0x4000 - count
					}
					if n != 0 {
						if deferred != nil {
							if deferredSize != 0 {
								copy(unsafe.Slice(deferredDest, int(deferredSize)), unsafe.Slice(deferred, int(deferredSize)))
								start, deferred = deferredDest, nil
							}
							copy(unsafe.Slice(dest, int(n)), unsafe.Slice(v.Cursor, int(n)))
						} else if deferredSize == 0 {
							deferred, deferredDest, deferredSize = v.Cursor, dest, n
						} else {
							copy(unsafe.Slice(dest, int(n)), unsafe.Slice(v.Cursor, int(n)))
						}
						dest = (*byte)(unsafe.Add(unsafe.Pointer(dest), uintptr(n)))
					}
				}
				// A full first source uses all available bytes, even above 0x4000.
				// Only concatenated/deferred copies are bounded to the scratch size.
				v.Available -= n
				v.Remaining -= n
				v.Cursor = (*byte)(unsafe.Add(unsafe.Pointer(v.Cursor), uintptr(n)))
				count += n
				if s.Field3 != 0 {
					break
				}
			}
		}
		if svc := audioSourceTestServices; svc != nil && svc.Load != nil {
			svc.Load(s.Smp, uint32(index), unsafe.Pointer(start), count)
		} else {
			s.Smp.LoadBuffer(uint32(index), unsafe.Slice(start, int(count)))
		}
	}
	return -1
}

//export sub_4BD720
func sub_4BD720(p unsafe.Pointer) unsafe.Pointer {
	ctx := (*AudioContextNative)(p)
	v := (*AudioVoiceNative)(audioNativeAlloc(unsafe.Sizeof(AudioVoiceNative{})))
	if v == nil {
		return nil
	}
	v.Link.init()
	v.Mixed.Init()
	sub_4BD7C0(unsafe.Pointer(v))
	v.Context, v.Methods = ctx, ctx.Methods
	if audioNativeCall(v.Methods.Create, unsafe.Pointer(v)) == 0 {
		return unsafe.Pointer(v)
	}
	sub_4BD7A0(unsafe.Pointer(v))
	return nil
}

//export sub_4BD7A0
func sub_4BD7A0(p unsafe.Pointer) {
	v := (*AudioVoiceNative)(p)
	audioNativeVoid(v.Methods.Destroy, p)
	audioNativeFree(p)
}

//export sub_4BD7C0
func sub_4BD7C0(p unsafe.Pointer) unsafe.Pointer {
	v := (*AudioVoiceNative)(p)
	C.nox_audio_source_callbacks(p)
	v.OnRefill, v.OnRestart, v.OnEnd, v.UserData = nil, nil, nil, nil
	v.Kind = 1
	v.Flags, v.Repeat, v.Priority = 0, 0, 0
	v.Master = audioNativeMasterGroup()
	v.OwnerGroup = nil
	v.Group.Init()
	v.Buffer = nil
	return unsafe.Pointer(&v.Group)
}

//export sub_4BD840
func sub_4BD840(p unsafe.Pointer) {
	v := (*AudioVoiceNative)(p)
	v.Mixed.Init()
	v.Mixed.Mix(&v.Group)
	v.Group.ClearUpdated()
	if v.OwnerGroup != nil {
		v.Mixed.Mix(v.OwnerGroup)
		v.OwnerGroup.ClearUpdated()
	}
	if v.Master != nil {
		v.Mixed.Mix(v.Master)
	}
	v.Mixed.Mix(&v.Context.Group)
	if v.Context.Master != nil {
		v.Mixed.Mix(v.Context.Master)
	}
}

//export sub_4BD8C0
func sub_4BD8C0(p unsafe.Pointer) C.int {
	v := (*AudioVoiceNative)(p)
	if v.OnRefill != nil {
		if result := audioNativeCall(v.OnRefill, p); result != 0 {
			v.Available, v.Remaining, v.Cursor = 0, 0, nil
			return C.int(result)
		}
	} else {
		if v.Chunk != nil {
			v.Chunk = (*AudioChunkNative)(unsafe.Pointer(v.Chunk.Link.next()))
			if v.Chunk != nil {
				v.Cursor, v.Available, v.Remaining = v.Chunk.Data, v.Chunk.Size, v.Chunk.Size
				return 0
			}
		}
		v.Available = 0
	}
	return 0
}

//export sub_4BD940
func sub_4BD940(p unsafe.Pointer) C.int {
	v := (*AudioVoiceNative)(p)
	var buffer *AudioBufferNative
	if v.Repeat != 0 {
		if v.Repeat != -1 {
			v.Repeat--
		}
		buffer = v.Buffer
	}
	sub_4BDB90(p, unsafe.Pointer(buffer))
	audioNativeVoid(v.OnRestart, p)
	if v.Buffer != nil {
		audioNativeVoid(v.Methods.Restart, p)
	}
	return 0
}

//export sub_4BD9B0
func sub_4BD9B0(p unsafe.Pointer) C.int {
	v := (*AudioVoiceNative)(p)
	v.Buffer = nil
	v.Flags &^= 5
	v.Repeat = 0
	audioSourceDependency("timer_init", v)
	v.Group.Init()
	return C.int(audioNativeCall(v.OnEnd, p))
}

//export sub_4BDA60
func sub_4BDA60(p unsafe.Pointer) {
	sub_4BDA80(p)
	sub_486E90(p)
	sub_4BD7A0(p)
}

//export sub_4BDA80
func sub_4BDA80(p unsafe.Pointer) C.int {
	v := (*AudioVoiceNative)(p)
	if v.Flags&5 != 0 {
		audioNativeVoid(v.Methods.Stop, p)
	}
	result := audioNativeCall(v.OnStop, p)
	v.Buffer = nil
	return C.int(result)
}

//export sub_4BDB20
func sub_4BDB20(p unsafe.Pointer) unsafe.Pointer {
	(*AudioVoiceNative)(p).Flags |= 0x10
	return p
}

//export sub_4BDB30
func sub_4BDB30(p unsafe.Pointer) unsafe.Pointer {
	(*AudioVoiceNative)(p).Flags &^= 0x10
	return p
}

//export sub_4BDB40
func sub_4BDB40(p unsafe.Pointer) C.int {
	v := (*AudioVoiceNative)(p)
	if v.Flags&5 != 0 {
		return -2146500608
	}
	if v.Buffer == nil {
		return -2147024896
	}
	audioSourceDependency("timer_update", v)
	v.Group.Update()
	audioSourceDependency("mix", v)
	sub_4BD840(p)
	result := audioNativeCall(v.Methods.Play, p)
	if result == 0 {
		v.Flags |= 1
	}
	return C.int(result)
}

//export sub_4BDB90
func sub_4BDB90(p, data unsafe.Pointer) {
	v, b := (*AudioVoiceNative)(p), (*AudioBufferNative)(data)
	v.Buffer = b
	if b != nil {
		v.Chunk = (*AudioChunkNative)(unsafe.Pointer(b.Chunks.first()))
		if v.Chunk != nil {
			v.Cursor, v.Available, v.Remaining = v.Chunk.Data, v.Chunk.Size, v.Chunk.Size
			b.Data = nil
		} else {
			v.Cursor, v.Available, v.Remaining = b.Data, b.Size, b.Size
		}
	}
}

//export sub_4BDC00
func sub_4BDC00(p unsafe.Pointer) unsafe.Pointer {
	// This helper receives the adjacent priority/flags/repeat scalar triplet.
	v := unsafe.Slice((*uint32)(p), 3)
	v[1], v[2] = 0, 0
	return p
}
