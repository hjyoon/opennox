package legacy

import (
	"runtime"
	"unsafe"
)

type accessListNode416720 struct{ Next, Prev, Owner *accessListNode416720 }
type accessAllowed416720 struct {
	List accessListNode416720
	Name [26]uint16
}
type accessBanned416720 struct {
	List    accessListNode416720
	Name    [26]uint16
	Expires uint64
	Serial  [24]byte
}

// Diagnostic heads are actual native objects, never four-byte PE cells. All
// hooks are serial/LIFO test seams; nil selects existing production owners.
type accessListNativeHooks416720 struct {
	Allowed, Banned *accessListNode416720
	Allocate        func(uintptr) unsafe.Pointer
	Free            func(unsafe.Pointer)
	Clock           func() uint32
	Refresh         func() unsafe.Pointer
	Load            func(unsafe.Pointer) int32
	Save            func(string)
}
type accessListNativeFrame416720 struct {
	Hooks    *accessListNativeHooks416720
	Previous *accessListNativeFrame416720
}

var accessListNativeTop416720 *accessListNativeFrame416720

func accessListNativeCurrent416720() *accessListNativeHooks416720 {
	if accessListNativeTop416720 == nil {
		return nil
	}
	return accessListNativeTop416720.Hooks
}
func pushAccessListNativeHooks416720(h *accessListNativeHooks416720) func() {
	f := &accessListNativeFrame416720{h, accessListNativeTop416720}
	accessListNativeTop416720 = f
	return func() {
		if accessListNativeTop416720 != f {
			panic("access-list observer pop order")
		}
		accessListNativeTop416720 = f.Previous
	}
}
func accessListNativeHead416720(banned bool) *accessListNode416720 {
	if h := accessListNativeCurrent416720(); h != nil {
		if banned && h.Banned != nil {
			return h.Banned
		}
		if !banned && h.Allowed != nil {
			return h.Allowed
		}
	}
	return (*accessListNode416720)(accessListNativeDefaultHead416720(banned))
}
func accessListNativeClear416720(n *accessListNode416720) { n.Next = n; n.Prev = n; n.Owner = n }
func accessListNativeInit416720(n *accessListNode416720)  { n.Next = n; n.Prev = n; n.Owner = nil }
func accessListNativeNext416720(n *accessListNode416720) *accessListNode416720 {
	if n == nil {
		return nil
	}
	next := n.Next
	if next != nil && next == next.Owner {
		return nil
	}
	return next
}
func accessListNativeAppend416720(head, n *accessListNode416720) {
	if head == nil || n == nil {
		panic("access-list append nil")
	}
	previous := head.Prev
	n.Next = head
	n.Prev = previous
	head.Prev = n
	if previous != nil {
		previous.Next = n
	}
}
func accessListNativeRemove416720(n *accessListNode416720) {
	if n == nil || n.Next == nil || n.Prev == nil {
		return
	}
	n.Prev.Next = n.Next
	n.Next.Prev = n.Prev
	n.Next = n
	n.Prev = n
}
func accessListNativeAlloc416720(size uintptr) unsafe.Pointer {
	if h := accessListNativeCurrent416720(); h != nil && h.Allocate != nil {
		return h.Allocate(size)
	}
	return accessListNativeDefaultAlloc416720(size)
}
func accessListNativeFree416720(p unsafe.Pointer) {
	if h := accessListNativeCurrent416720(); h != nil && h.Free != nil {
		h.Free(p)
		runtime.KeepAlive(p)
		return
	}
	accessListNativeDefaultFree416720(p)
}
func accessListNativeClock416720() uint32 {
	if h := accessListNativeCurrent416720(); h != nil && h.Clock != nil {
		return h.Clock()
	}
	return uint32(PlatformTicks())
}
func accessListNativeRefresh416720() unsafe.Pointer {
	if h := accessListNativeCurrent416720(); h != nil && h.Refresh != nil {
		return h.Refresh()
	}
	return accessListNativeDefaultRefresh416720()
}
func accessListNativeCopyName416720(dst unsafe.Pointer, src unsafe.Pointer) {
	// Preserve the original unbounded UTF16 copy. Public callers provide a
	// terminated name fitting26 code units; this port adds no truncation guard.
	for i := uintptr(0); ; i += 2 {
		v := *(*uint16)(unsafe.Add(src, i))
		*(*uint16)(unsafe.Add(dst, i)) = v
		if v == 0 {
			break
		}
	}
	runtime.KeepAlive(src)
}
func sub416720Native() {
	index := int32(0)
	for n := sub416900Native(); n != nil; {
		next := sub416910Native(n) // Cache before expiry read/clock/removal.
		expires := (*accessBanned416720)(n).Expires
		if expires != 0 && uint64(accessListNativeClock416720()) > expires {
			sub416820Native(index)
		}
		index++
		n = next
	}
}
func sub416770Native(minutes int32, name, serial unsafe.Pointer) unsafe.Pointer {
	head := accessListNativeHead416720(true)
	p := accessListNativeAlloc416720(unsafe.Sizeof(accessBanned416720{}))
	if p == nil {
		return nil
	}
	n := (*accessBanned416720)(p)
	accessListNativeInit416720(&n.List)
	accessListNativeCopyName416720(unsafe.Pointer(&n.Name[0]), name)
	if serial != nil {
		ended := false
		for i := 0; i < 23; i++ {
			var b byte
			if !ended {
				b = *(*byte)(unsafe.Add(serial, i))
				ended = b == 0
			}
			n.Serial[i] = b
		}
		n.Serial[23] = 0
	} else {
		n.Serial[0] = 0
	}
	accessListNativeAppend416720(head, &n.List)
	if minutes != 0 {
		n.Expires = 60000*uint64(minutes) + uint64(accessListNativeClock416720())
	} else {
		n.Expires = 0
	}
	result := accessListNativeRefresh416720()
	runtime.KeepAlive(name)
	runtime.KeepAlive(serial)
	return result
}
func sub416820Native(index int32) {
	n := sub416900Native()
	if n == nil {
		return
	}
	for index != 0 {
		index--
		n = sub416910Native(n)
		if n == nil {
			return
		}
	}
	accessListNativeRemove416720((*accessListNode416720)(n))
	accessListNativeFree416720(n)
}
func sub416860Native(index int32) unsafe.Pointer {
	result := sub4168E0Native()
	n := result
	if n != nil {
		for index != 0 {
			index--
			result = sub4168F0Native(n)
			n = result
			if n == nil {
				return nil
			}
		}
		accessListNativeRemove416720((*accessListNode416720)(n))
		accessListNativeFree416720(n)
	}
	// The old C returns the removed (freed) pointer. Opaque identity only;
	// callers/tests must never dereference it after Free.
	return result
}
func sub4168A0Native(name unsafe.Pointer) unsafe.Pointer {
	head := accessListNativeHead416720(false)
	p := accessListNativeAlloc416720(unsafe.Sizeof(accessAllowed416720{}))
	if p == nil {
		return nil
	}
	n := (*accessAllowed416720)(p)
	accessListNativeInit416720(&n.List)
	accessListNativeCopyName416720(unsafe.Pointer(&n.Name[0]), name)
	accessListNativeAppend416720(head, &n.List)
	result := accessListNativeRefresh416720()
	runtime.KeepAlive(name)
	return result
}
func sub4168E0Native() unsafe.Pointer {
	return unsafe.Pointer(accessListNativeNext416720(accessListNativeHead416720(false)))
}
func sub4168F0Native(p unsafe.Pointer) unsafe.Pointer {
	// &a1->list is at offset0; current native C safe-next accepts nil.
	n := accessListNativeNext416720((*accessListNode416720)(p))
	runtime.KeepAlive(p)
	return unsafe.Pointer(n)
}
func sub416900Native() unsafe.Pointer {
	return unsafe.Pointer(accessListNativeNext416720(accessListNativeHead416720(true)))
}
func sub416910Native(p unsafe.Pointer) unsafe.Pointer {
	n := accessListNativeNext416720((*accessListNode416720)(p))
	runtime.KeepAlive(p)
	return unsafe.Pointer(n)
}
func sub416920Native() int32 {
	allowed, banned := accessListNativeHead416720(false), accessListNativeHead416720(true)
	accessListNativeClear416720(allowed)
	accessListNativeClear416720(banned)
	path := accessListNativeConfig416720()
	if h := accessListNativeCurrent416720(); h != nil && h.Load != nil {
		return h.Load(path)
	}
	return accessListNativeDefaultLoad416720(path)
}
func sub416950Native() unsafe.Pointer {
	if h := accessListNativeCurrent416720(); h != nil && h.Save != nil {
		h.Save("ban.txt")
	} else {
		accessListNativeDefaultSave416720()
	}
	for n := sub4168E0Native(); n != nil; {
		next := sub4168F0Native(n)
		accessListNativeRemove416720((*accessListNode416720)(n))
		accessListNativeFree416720(n)
		n = next
	}
	first := sub416900Native()
	for n := first; n != nil; {
		next := sub416910Native(n)
		accessListNativeRemove416720((*accessListNode416720)(n))
		accessListNativeFree416720(n)
		n = next
	}
	return first // Old C's first banned allocation, opaque after Free.
}
