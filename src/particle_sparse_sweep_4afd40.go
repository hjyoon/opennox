package opennox

import (
	"image/color"
	"reflect"
	"slices"
	"unsafe"

	noxcolor "github.com/opennox/libs/color"

	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

// This is only a cost choice. Both traversals preserve the same live forward
// cursor; unrelated keys are not particle slots in the original bitmap.
func particleSparseRange4AFD40(first, last unsafe.Pointer, active int) bool {
	span := (uintptr(last) - uintptr(first)) / unsafe.Alignof(uintptr(0))
	return span > uintptr(active)*64
}

func (pfx *partFXes) particleKeysAfter4AFD40(after unsafe.Pointer, keys []unsafe.Pointer) []unsafe.Pointer {
	keys = keys[:0]
	for handle := range pfx.byHandle {
		if uintptr(handle) > uintptr(after) {
			keys = append(keys, handle)
		}
	}
	slices.SortFunc(keys, func(a, b unsafe.Pointer) int {
		if uintptr(a) < uintptr(b) {
			return -1
		}
		if uintptr(a) > uintptr(b) {
			return 1
		}
		return 0
	})
	return keys
}

// Default update changes only this record (free also removes only its own
// key). Unknown Go/C drawing and arbitrary color.Color methods are barriers.
// Restrict the known point-draw case to the audited scalar color conversions.
func particleCallbacksMayMutate4AFD40(p *particleFx, drawPC, updatePC uintptr) bool {
	if p.updateFunc != nil && reflect.ValueOf(p.updateFunc).Pointer() != updatePC {
		return true
	}
	if p.flags&8 != 0 || p.drawFunc == nil {
		return false
	}
	if reflect.ValueOf(p.drawFunc).Pointer() != drawPC || p.image != nil || p.rendPart != nil || p.drawable12 != nil && p.drawableVp != nil {
		return true
	}
	if p.pointSize == 0 {
		return false
	}
	if p.pointSize < 1 || p.pointSize > 3 {
		return true
	}
	switch p.color.(type) {
	case noxcolor.RGBA5551, noxcolor.RGB555,
		color.RGBA, color.NRGBA, color.Gray, color.Gray16, color.Alpha, color.Alpha16:
		return false
	default:
		return true
	}
}

func (pfx *partFXes) drawSparseParticles4AFD40(after unsafe.Pointer) {
	drawPC := reflect.ValueOf(partfxDraw).Pointer()
	updatePC := reflect.ValueOf(partfxUpdateDef).Pointer()
	for handle, p := range pfx.byHandle {
		if uintptr(handle) > uintptr(after) && particleCallbacksMayMutate4AFD40(p, drawPC, updatePC) {
			// Preserve the live native cursor for arbitrary callbacks. Do
			// not repeatedly sort the map for custom effect updates.
			pfx.drawNativeParticlesAfter4AFD40(after)
			return
		}
	}
	// All remaining callbacks are the audited local-only default paths.
	// They cannot create/reactivate/change future slots; only self-free is
	// possible. Unknown color/image/C callbacks never use this snapshot.
	keys := pfx.particleKeysAfter4AFD40(after, make([]unsafe.Pointer, 0, len(pfx.byHandle)))
	for i := 0; i < len(keys); {
		handle := keys[i]
		i++
		p := pfx.byHandle[handle]
		if p == nil {
			continue
		}
		if p.flags&8 == 0 && p.drawFunc != nil {
			p.drawFunc(p)
		}
		if p.updateFunc != nil && !p.updateFunc(p) {
			p.free()
		}
	}
}

func (pfx *partFXes) drawNativeParticlesAfter4AFD40(after unsafe.Pointer) {
	var first, last unsafe.Pointer
	for handle := range pfx.byHandle {
		if uintptr(handle) <= uintptr(after) {
			continue
		}
		if first == nil || uintptr(handle) < uintptr(first) {
			first = handle
		}
		if last == nil || uintptr(handle) > uintptr(last) {
			last = handle
		}
	}
	if first == nil {
		return
	}
	frontier := handles.FrontierPtr()
	for handle := first; ; handle = unsafe.Add(handle, unsafe.Alignof(uintptr(0))) {
		if p := pfx.byHandle[handle]; p != nil {
			if p.flags&8 == 0 && p.drawFunc != nil {
				p.drawFunc(p)
			}
			if p.updateFunc != nil && !p.updateFunc(p) {
				p.free()
			}
			if current := handles.FrontierPtr(); current != frontier {
				frontier, last = current, current
			}
		}
		if uintptr(handle) >= uintptr(last) {
			break
		}
	}
}
