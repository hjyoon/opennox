package legacy

import (
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

// The source blob remains PE32-packed: title at +0, char key at +4, mask at
// +8, stride 12. PtrPtr owns native pointer values outside those DWORD bytes.
type titleTable415960 struct {
	Base, Ready uintptr
	Source      string
	Line        int32
}

var titleWeapon415960 = titleTable415960{33392, 371248, "C:\\NoxPost\\src\\common\\Object\\WeapLook.c", 200}
var titleArmor415960 = titleTable415960{35496, 371256, "C:\\NoxPost\\src\\common\\Object\\ArmrLook.c", 380}

type titleAccess415960 struct {
	Table, Offset uintptr
	Address       unsafe.Pointer
	Pointer       unsafe.Pointer
	Word          uint32
	Width         uintptr
	Write         bool
	Complete      bool
}

type titleObserver415960 struct {
	Access  func(titleAccess415960)
	Compare func(query, title unsafe.Pointer) (int32, bool)
	Load    func(key unsafe.Pointer, source string, line int32) (unsafe.Pointer, bool)
}

type titleObserverFrame415960 struct {
	Observer *titleObserver415960
	Previous *titleObserverFrame415960
}

var titleObserverHead415960 *titleObserverFrame415960

func titlePushObserver415960(o *titleObserver415960) func() {
	f := &titleObserverFrame415960{o, titleObserverHead415960}
	titleObserverHead415960 = f
	return func() {
		if titleObserverHead415960 != f {
			panic("title table observer pop order")
		}
		titleObserverHead415960 = f.Previous
	}
}

func titleObserverCurrent415960() *titleObserver415960 {
	if titleObserverHead415960 == nil {
		return nil
	}
	return titleObserverHead415960.Observer
}

func titleObserve415960(e titleAccess415960) {
	if o := titleObserverCurrent415960(); o != nil && o.Access != nil {
		o.Access(e)
	}
}

func titleReadPointer415960(t titleTable415960, off uintptr) unsafe.Pointer {
	p := memmap.PtrPtr(0x587000, off)
	e := titleAccess415960{Table: t.Base, Offset: off, Address: unsafe.Pointer(p), Width: unsafe.Sizeof(*p)}
	titleObserve415960(e)
	v := *p
	e.Pointer, e.Complete = v, true
	titleObserve415960(e)
	runtime.KeepAlive(p)
	return v
}

func titleWritePointer415960(t titleTable415960, off uintptr, value unsafe.Pointer) {
	p := memmap.PtrPtr(0x587000, off)
	e := titleAccess415960{Table: t.Base, Offset: off, Address: unsafe.Pointer(p), Pointer: value, Width: unsafe.Sizeof(*p), Write: true}
	titleObserve415960(e)
	*p = value
	e.Complete = true
	titleObserve415960(e)
	runtime.KeepAlive(value)
	runtime.KeepAlive(p)
}

func titleReadWord415960(t titleTable415960, base, off uintptr) uint32 {
	p := memmap.PtrUint32(base, off)
	e := titleAccess415960{Table: t.Base, Offset: off, Address: unsafe.Pointer(p), Width: 4}
	titleObserve415960(e)
	v := *p
	e.Word, e.Complete = v, true
	titleObserve415960(e)
	return v
}

func titleWriteReady415960(t titleTable415960) {
	p := memmap.PtrUint32(0x5D4594, t.Ready)
	e := titleAccess415960{Table: t.Base, Offset: t.Ready, Address: unsafe.Pointer(p), Word: 1, Width: 4, Write: true}
	titleObserve415960(e)
	*p = 1
	e.Complete = true
	titleObserve415960(e)
}

func titleCompare415960(query, title unsafe.Pointer) int32 {
	if o := titleObserverCurrent415960(); o != nil && o.Compare != nil {
		if v, handled := o.Compare(query, title); handled {
			return v
		}
	}
	return titleDefaultCompare415960(query, title)
}

func titleLoadString415960(key unsafe.Pointer, t titleTable415960) unsafe.Pointer {
	if o := titleObserverCurrent415960(); o != nil && o.Load != nil {
		if v, handled := o.Load(key, t.Source, t.Line); handled {
			return v
		}
	}
	return titleDefaultLoadString415960(key, t.Source, t.Line)
}

func titleToMask415960(t titleTable415960, query unsafe.Pointer) uint32 {
	// Preserve the existing native API's nil-query policy. The original PE
	// delegates a nonempty-table/nil-query case to 5684F0 instead.
	if query == nil {
		return 0
	}
	if titleReadPointer415960(t, t.Base) == nil {
		return 0
	}
	for off := t.Base; ; off += 12 {
		current := titleReadPointer415960(t, off)
		if titleCompare415960(query, current) == 0 {
			v := titleReadWord415960(t, 0x587000, off+8)
			runtime.KeepAlive(query)
			return v
		}
		// Reload after the external comparison, as in the original body.
		if titleReadPointer415960(t, off+12) == nil {
			runtime.KeepAlive(query)
			return 0
		}
	}
}

func titleFromMask415960(t titleTable415960, mask uint32) unsafe.Pointer {
	if titleReadPointer415960(t, t.Base) == nil {
		return nil
	}
	for off := t.Base; ; off += 12 {
		if titleReadWord415960(t, 0x587000, off+8) == mask {
			return titleReadPointer415960(t, off)
		}
		if titleReadPointer415960(t, off+12) == nil {
			return nil
		}
	}
}

// The legacy loader's EAX is an opaque mixed return word: existing ready
// DWORD, zero, or the final char-key cell address. Only the thin C adapter
// reinterprets it as its historical char** ABI; Go never fabricates a pointer
// from the scalar branch and no caller dereferences that return word.
func titleLoadTable415960(t titleTable415960) uintptr {
	if ready := titleReadWord415960(t, 0x5D4594, t.Ready); ready != 0 {
		return uintptr(ready)
	}
	if titleReadPointer415960(t, t.Base+4) == nil {
		titleWriteReady415960(t)
		return 0
	}
	for off := t.Base + 4; ; off += 12 {
		key := titleReadPointer415960(t, off)
		text := titleLoadString415960(key, t)
		titleWritePointer415960(t, off-4, text)
		next := titleReadPointer415960(t, off+12)
		runtime.KeepAlive(key)
		if next == nil {
			titleWriteReady415960(t)
			owner := memmap.PtrPtr(0x587000, off+12)
			word := uintptr(unsafe.Pointer(owner))
			runtime.KeepAlive(owner)
			return word
		}
	}
}
