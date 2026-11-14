package legacy

/*
#include <stdlib.h>
#include "audio_source_native.h"
extern void *dword_587000_155144;
extern void *dword_5d4594_805984;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

var audioGlobalOwner *audioGlobalNative

//export sub_431270
func sub_431270() {
	if p := C.dword_5d4594_805984; p != nil {
		sub_487680(p)
		C.dword_5d4594_805984 = nil
	}
}

//export sub_431290
func sub_431290() {
	if p := C.dword_5d4594_805984; p != nil {
		sub_487970(p, -1)
	}
}

func audioNativeGlobal() *audioGlobalNative {
	if audioGlobalOwner == nil {
		audioGlobalOwner = (*audioGlobalNative)(audioNativeAlloc(unsafe.Sizeof(audioGlobalNative{})))
		audioGlobalOwner.Devices.init()
		audioGlobalOwner.Contexts.init()
		audioGlobalOwner.Group.Init()
	}
	return audioGlobalOwner
}

// The blob's old 128-byte DWORD/list record is no longer a native owner.
// Reconnect its external global after each blob initialization.
func initAudioNativeGlobal() { C.dword_587000_155144 = unsafe.Pointer(audioNativeGlobal()) }

func AudioInitNative() {
	g := audioNativeGlobal()
	g.Devices.init()
	g.Contexts.init()
	g.Busy = 0
	g.Group.Init()
	initAudioNativeGlobal()
}

func AudioServeNative() {
	g := audioNativeGlobal()
	if g.Busy != 0 {
		return
	}
	for n := g.Contexts.first(); n != nil; {
		next := n.next()
		ctx := (*AudioContextNative)(unsafe.Pointer(n))
		if ctx.Kind&2 == 0 {
			audioNativeCall(ctx.Update, unsafe.Pointer(ctx))
		}
		n = next
	}
}

func audioBusyEnd(busy *int32) int32 {
	*busy--
	result := *busy
	if result < 0 {
		*busy = 0
	}
	return result
}

//export sub_486E30
func sub_486E30(p, voice unsafe.Pointer) C.int {
	ctx, v := (*AudioContextNative)(p), (*AudioVoiceNative)(voice)
	v.Context = ctx
	ctx.Count++
	ctx.Busy++
	ctx.Voices.append(&v.Link)
	return C.int(audioBusyEnd(&ctx.Busy))
}

//export sub_486E90
func sub_486E90(p unsafe.Pointer) C.int {
	v := (*AudioVoiceNative)(p)
	ctx := v.Context
	v.Link.remove()
	ctx.Count--
	ctx.Busy++
	v.Link.remove()
	return C.int(audioBusyEnd(&ctx.Busy))
}

//export sub_486FA0
func sub_486FA0(p unsafe.Pointer) unsafe.Pointer {
	result := sub_486FE0(p)
	if result != nil {
		class := (*AudioClassNative)(p)
		class.State |= 1
		sub_487050(result)
		if class.Flags&2 != 0 {
			*memmap.PtrUint32(0x5D4594, 1193332) = 1
		}
	}
	return result
}

//export sub_486FE0
func sub_486FE0(p unsafe.Pointer) unsafe.Pointer {
	d := (*AudioDeviceNative)(audioNativeAlloc(unsafe.Sizeof(AudioDeviceNative{})))
	if d == nil {
		return nil
	}
	d.Link.init()
	d.Class = (*AudioClassNative)(p)
	if audioNativeCall(d.Class.Create, unsafe.Pointer(d)) == 0 {
		return unsafe.Pointer(d)
	}
	sub_487030(unsafe.Pointer(d))
	return nil
}

//export sub_487030
func sub_487030(p unsafe.Pointer) {
	d := (*AudioDeviceNative)(p)
	audioNativeVoid(d.Class.Destroy, p)
	d.Class.State &^= 1
	audioNativeFree(p)
}

//export sub_487050
func sub_487050(p unsafe.Pointer) { audioNativeGlobal().Devices.append(&(*AudioDeviceNative)(p).Link) }

//export sub_487070
func sub_487070(p unsafe.Pointer) {
	sub_487090(p)
	sub_487030(p)
	*memmap.PtrUint32(0x5D4594, 1193332) = 0
}

//export sub_487090
func sub_487090(p unsafe.Pointer) { (*AudioDeviceNative)(p).Link.remove() }

//export sub_4870A0
func sub_4870A0() {
	for n := audioNativeGlobal().Devices.first(); n != nil; {
		next := n.next()
		sub_487070(unsafe.Pointer(n))
		n = next
	}
}

//export sub_4870E0
func sub_4870E0(out unsafe.Pointer) unsafe.Pointer {
	n := audioNativeGlobal().Devices.first()
	*(*unsafe.Pointer)(out) = unsafe.Pointer(n)
	return unsafe.Pointer(n)
}

//export sub_487100
func sub_487100(out unsafe.Pointer) unsafe.Pointer {
	p := (*unsafe.Pointer)(out)
	if *p != nil {
		*p = unsafe.Pointer((*AudioDeviceNative)(*p).Link.next())
	}
	return *p
}

//export sub_487150_go
func sub_487150_go(index C.int, format unsafe.Pointer) unsafe.Pointer {
	requested := int32(index)
	if requested == -1 {
		requested = 0
	}
	if requested < 0 {
		return nil
	}
	var device unsafe.Pointer
	var slot int32
	sub_487360(C.int(requested), unsafe.Pointer(&device), (*C.int)(unsafe.Pointer(&slot)))
	if device == nil {
		return nil
	}
	d := (*AudioDeviceNative)(device)
	if slot < 0 || slot >= int32(len(d.Contexts)) {
		return nil
	}
	v := d.Contexts[slot]
	if v == nil {
		v = (*AudioContextNative)(sub_4871C0_go(device, C.int(slot), format))
		if v == nil {
			return nil
		}
		v.Priority = requested
		sub_487310(unsafe.Pointer(v))
	}
	v.Users++
	return unsafe.Pointer(v)
}

//export sub_4871C0_go
func sub_4871C0_go(device unsafe.Pointer, index C.int, format unsafe.Pointer) unsafe.Pointer {
	d := (*AudioDeviceNative)(device)
	// Native arrays have sixteen slots, matching the original device owner.
	if index < 0 || index >= C.int(len(d.Contexts)) {
		return nil
	}
	ctx := (*AudioContextNative)(audioNativeAlloc(unsafe.Sizeof(AudioContextNative{})))
	if ctx == nil {
		return nil
	}
	ctx.Link.init()
	ctx.Index, ctx.Device = int32(index), d
	d.Count++
	d.Contexts[index] = ctx
	ctx.Methods = d.Class.Methods
	ctx.Voices.init()
	ctx.Group.Init()
	ctx.Interval = 33
	ctx.Update = C.nox_audio_context_update_callback()
	if format != nil {
		sub_487590_go(unsafe.Pointer(ctx), format)
	}
	if audioNativeCall(d.Class.ContextCreate, unsafe.Pointer(ctx)) == 0 {
		return unsafe.Pointer(ctx)
	}
	sub_4872C0(unsafe.Pointer(ctx))
	return nil
}

//export sub_4872C0
func sub_4872C0(p unsafe.Pointer) {
	ctx := (*AudioContextNative)(p)
	sub_487910(p, -1)
	d := ctx.Device
	audioNativeVoid(d.Class.ContextDestroy, p)
	d.Contexts[ctx.Index] = nil
	d.Count--
	if int32(d.Count) < 0 {
		d.Count = 0
	}
	audioNativeFree(p)
}

//export sub_487310
func sub_487310(p unsafe.Pointer) C.int {
	g := audioNativeGlobal()
	g.Busy++
	g.Contexts.append(&(*AudioContextNative)(p).Link)
	return C.int(audioBusyEnd(&g.Busy))
}

//export sub_487360
func sub_487360(index C.int, device unsafe.Pointer, slot *C.int) unsafe.Pointer {
	remaining := int32(index)
	n := audioNativeGlobal().Devices.first()
	for n != nil {
		d := (*AudioDeviceNative)(unsafe.Pointer(n))
		if remaining < int32(d.Limit) {
			break
		}
		remaining -= int32(d.Limit)
		n = n.next()
	}
	*(*unsafe.Pointer)(device) = unsafe.Pointer(n)
	if n != nil {
		*slot = C.int(remaining)
		return unsafe.Pointer(slot)
	}
	*slot = -1
	return nil
}

//export sub_4873C0
func sub_4873C0(p unsafe.Pointer) C.int {
	ctx := (*AudioContextNative)(p)
	if ctx.Busy != 0 {
		return -2146304000
	}
	now := PlatformTicks()
	delta := now - ctx.Last
	if delta < ctx.Interval {
		return 0
	}
	ctx.Delta = delta
	if ctx.Peak > 10*ctx.Interval {
		ctx.Peak = 0
	}
	if delta > ctx.Peak {
		ctx.Peak = delta
	}
	ctx.Group.Update()
	if ctx.Master != nil {
		ctx.Master.Update()
	}
	updated := ctx.Master != nil && ctx.Master.IsUpdated() || ctx.Group.IsUpdated()
	ctx.Last = now
	for n := ctx.Voices.first(); n != nil; {
		next := n.next()
		v := (*AudioVoiceNative)(unsafe.Pointer(n))
		if v.Flags&1 != 0 && v.Buffer != nil {
			v.Group.Update()
			if updated || v.Group.IsUpdated() || v.Master != nil && v.Master.IsUpdated() || v.OwnerGroup != nil && v.OwnerGroup.IsUpdated() {
				sub_4BD840(unsafe.Pointer(v))
				audioNativeVoid(v.Methods.Refresh, unsafe.Pointer(v))
			}
		}
		n = next
	}
	if ctx.Master != nil {
		ctx.Master.ClearUpdated()
	}
	ctx.Group.ClearUpdated()
	return 0
}

//export sub_487590_go
func sub_487590_go(p, format unsafe.Pointer) unsafe.Pointer {
	(*AudioContextNative)(p).Format60 = *(*[7]uint32)(format)
	return p
}

//export sub_4875B0
func sub_4875B0(out unsafe.Pointer) unsafe.Pointer {
	n := audioNativeGlobal().Contexts.first()
	*(*unsafe.Pointer)(out) = unsafe.Pointer(n)
	return unsafe.Pointer(n)
}

//export sub_4875D0
func sub_4875D0(out unsafe.Pointer) unsafe.Pointer {
	p := (*unsafe.Pointer)(out)
	if *p != nil {
		*p = unsafe.Pointer((*AudioContextNative)(*p).Link.next())
	}
	return *p
}

//export sub_4875F0
func sub_4875F0() C.int {
	g := audioNativeGlobal()
	g.Busy++
	for n := g.Contexts.first(); n != nil; {
		next := n.next()
		sub_487680(unsafe.Pointer(n))
		n = next
	}
	return C.int(audioBusyEnd(&g.Busy))
}

//export sub_487680
func sub_487680(p unsafe.Pointer) {
	sub_4876A0(p)
	sub_4872C0(p)
}

//export sub_4876A0
func sub_4876A0(p unsafe.Pointer) unsafe.Pointer {
	g := audioNativeGlobal()
	g.Busy++
	(*AudioContextNative)(p).Link.remove()
	audioBusyEnd(&g.Busy)
	// The historical mixed pointer/status return is ignored by every caller.
	return nil
}

//export sub_487750
func sub_487750(p unsafe.Pointer) unsafe.Pointer {
	ctx := (*AudioContextNative)(p)
	if ctx.Count >= ctx.Limit {
		return nil
	}
	v := sub_4BD720(p)
	if v != nil {
		sub_486E30(p, v)
	}
	return v
}

//export sub_487790
func sub_487790(p unsafe.Pointer, count C.int) C.int {
	var n int32
	if sub_487750(p) != nil {
		n++
		for left := int32(count) - 1; left != 0 && sub_487750(p) != nil; left-- {
			n++
		}
	}
	return C.int(n)
}

//export sub_4877D0
func sub_4877D0(p, out unsafe.Pointer) unsafe.Pointer {
	n := (*AudioContextNative)(p).Voices.first()
	*(*unsafe.Pointer)(out) = unsafe.Pointer(n)
	return unsafe.Pointer(n)
}

//export sub_4877F0
func sub_4877F0(out unsafe.Pointer) unsafe.Pointer {
	p := (*unsafe.Pointer)(out)
	if *p != nil {
		*p = unsafe.Pointer((*AudioVoiceNative)(*p).Link.next())
	}
	return *p
}

//export sub_487810
func sub_487810(p unsafe.Pointer, kind C.int) unsafe.Pointer {
	if kind == -1 {
		kind = 1
	}
	level := uint32(0xffffffff)
	activePriority, idlePriority := int32(127), int32(127)
	var active, idle *AudioVoiceNative
	for n := (*AudioContextNative)(p).Voices.first(); n != nil; n = n.next() {
		v := (*AudioVoiceNative)(unsafe.Pointer(n))
		if v.Kind != int32(kind) {
			continue
		}
		if v.Flags&0x15 == 0 {
			return unsafe.Pointer(v)
		}
		if v.Flags&1 != 0 {
			gain := v.Mixed.Timers[0].Current
			if v.Priority < activePriority || v.Priority == activePriority && gain < level && level-gain >= 0x666 {
				active, activePriority, level = v, v.Priority, gain
			}
		} else if v.Priority < idlePriority {
			idle, idlePriority = v, v.Priority
		}
	}
	if idle != nil && idlePriority <= activePriority {
		return unsafe.Pointer(idle)
	}
	return unsafe.Pointer(active)
}

//export sub_487910
func sub_487910(p unsafe.Pointer, kind C.int) C.int {
	for n := (*AudioContextNative)(p).Voices.first(); n != nil; {
		next := n.next()
		v := (*AudioVoiceNative)(unsafe.Pointer(n))
		if kind == -1 || v.Kind == int32(kind) {
			sub_4BDA60(unsafe.Pointer(v))
		}
		n = next
	}
	return 0
}

//export sub_487970
func sub_487970(p unsafe.Pointer, kind C.int) unsafe.Pointer {
	for n := (*AudioContextNative)(p).Voices.first(); n != nil; {
		next := n.next()
		v := (*AudioVoiceNative)(unsafe.Pointer(n))
		if kind == -1 || v.Kind == int32(kind) {
			sub_4BDA80(unsafe.Pointer(v))
		}
		n = next
	}
	return nil // discarded historical callback/status residue
}

//export sub_487C30
func sub_487C30(p unsafe.Pointer) {
	b := (*AudioBufferNative)(p)
	b.Data, b.Size, b.Format, b.Field24 = nil, 0, nil, 0
	b.Chunks.init()
}

//export sub_487C50
func sub_487C50(p, chunk unsafe.Pointer) C.int {
	b, c := (*AudioBufferNative)(p), (*AudioChunkNative)(chunk)
	b.Chunks.append(&c.Link)
	b.Size += c.Size
	c.Buffer = b
	return C.int(b.Size)
}

//export sub_487C80
func sub_487C80(p unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer((*AudioBufferNative)(p).Chunks.first())
}

//export sub_487D00
func sub_487D00(p *C.uint32_t) C.int {
	v := (*[6]uint32)(unsafe.Pointer(p))
	result := v[2] * v[3] * v[4]
	if v[1] == 1 {
		result >>= 2
	}
	v[5] = result
	return C.int(result)
}

//export sub_487D30
func sub_487D30(p, data unsafe.Pointer, size C.int) unsafe.Pointer {
	c := (*AudioChunkNative)(p)
	c.Size, c.Data = uint32(size), (*byte)(data)
	c.Link.init()
	c.Buffer = nil
	return p
}

//export sub_487D60
func sub_487D60(p unsafe.Pointer) unsafe.Pointer {
	(*AudioChunkNative)(p).Buffer = nil
	return p
}
