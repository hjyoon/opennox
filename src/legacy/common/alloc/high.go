package alloc

import (
	"unicode/utf16"
	"unsafe"
)

// MallocHigh allocates zeroed foreign memory for fixtures which must exercise
// pointer bits above 32. Windows' non-ASLR race executables have a low C heap,
// so their fixtures request a separate top-down mapping. Ordinary allocations
// and 32-bit targets retain the normal C allocator.
// The returned memory remains compatible with FreePtr and Realloc.
func MallocHigh(size uintptr) (unsafe.Pointer, func()) {
	if size == 0 {
		panic("zero alloc")
	}
	ptr, release := mallocHighMapping(size)
	if release == nil {
		return Malloc(size)
	}
	if ptr == nil {
		panic("cannot allocate high native mapping")
	}
	allocMu.Lock()
	allocs[ptr] = size
	allocReleases[ptr] = release
	allocMu.Unlock()
	return ptr, func() { FreePtr(ptr) }
}

// NewHigh is New with the high-address fixture allocation policy.
func NewHigh[T comparable](zero T) (*T, func()) {
	ptr, free := MallocHigh(unsafe.Sizeof(zero))
	return (*T)(ptr), free
}

// CallocHigh is Calloc with the high-address fixture allocation policy.
func CallocHigh(num int, size uintptr) (unsafe.Pointer, func()) {
	if num <= 0 || size == 0 || uintptr(num) > ^uintptr(0)/size {
		panic("invalid high allocation size")
	}
	return MallocHigh(uintptr(num) * size)
}

// MakeHigh is Make with the high-address fixture allocation policy.
func MakeHigh[T comparable, N number](src []T, n N) ([]T, func()) {
	if n < 0 || uint64(n) > uint64(^uint(0)>>1) {
		panic("invalid high allocation length")
	}
	var elem T
	ptr, free := CallocHigh(int(n), unsafe.Sizeof(elem))
	out := unsafe.Slice((*T)(ptr), int(n))
	copy(out, src)
	return out, free
}

func CloneSliceHigh[T comparable](src []T) ([]T, func()) {
	return MakeHigh(src, len(src))
}

func CStringHigh(s string) (*byte, func()) {
	arr, free := MakeHigh([]byte(nil), len(s)+1)
	copy(arr, s)
	return &arr[0], free
}

// CString16High is CString16 with the explicit high-address fixture policy.
func CString16High(s string) (*uint16, func()) {
	words := utf16.Encode([]rune(s))
	arr, free := MakeHigh(words, len(words)+1)
	return &arr[0], free
}
