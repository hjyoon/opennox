package alloc

import (
	"bytes"
	"testing"
	"unsafe"
)

func TestNativeByteCopyAndFill(t *testing.T) {
	ptr, release := MallocHigh(8)
	t.Cleanup(release)
	buf := unsafe.Slice((*byte)(ptr), 8)
	copy(buf, bytes.Repeat([]byte{0xa5}, len(buf)))
	dst := unsafe.Add(ptr, 2)
	src := []byte{0x80, 0, 0xff}

	if got := Memcpy(dst, unsafe.Pointer(&src[0]), uintptr(len(src))); got != dst ||
		!bytes.Equal(buf, []byte{0xa5, 0xa5, 0x80, 0, 0xff, 0xa5, 0xa5, 0xa5}) {
		t.Fatalf("native copy returned %p and wrote %x", got, buf)
	}
	if got := Memset(dst, 0xc3, 3); got != dst ||
		!bytes.Equal(buf, []byte{0xa5, 0xa5, 0xc3, 0xc3, 0xc3, 0xa5, 0xa5, 0xa5}) {
		t.Fatalf("native fill returned %p and wrote %x", got, buf)
	}
	if got := Memset(dst, 0, 2); got != dst ||
		!bytes.Equal(buf, []byte{0xa5, 0xa5, 0, 0, 0xc3, 0xa5, 0xa5, 0xa5}) {
		t.Fatalf("native clear returned %p and wrote %x", got, buf)
	}
}

func TestNativeCStringCopyAndAppend(t *testing.T) {
	ptr, release := MallocHigh(8)
	t.Cleanup(release)
	buf := unsafe.Slice((*byte)(ptr), 8)
	copy(buf, bytes.Repeat([]byte{0xa5}, len(buf)))
	dst := unsafe.Add(ptr, 1)
	word := []byte{0xff, 'x', 0, 'y'}
	suffix := []byte{0x80, 0, 'z'}

	if got := Strcpy(dst, unsafe.Pointer(&word[0])); got != dst ||
		!bytes.Equal(buf, []byte{0xa5, 0xff, 'x', 0, 0xa5, 0xa5, 0xa5, 0xa5}) {
		t.Fatalf("native string copy returned %p and wrote %x", got, buf)
	}
	if got := Strcat(dst, unsafe.Pointer(&suffix[0])); got != dst ||
		!bytes.Equal(buf, []byte{0xa5, 0xff, 'x', 0x80, 0, 0xa5, 0xa5, 0xa5}) {
		t.Fatalf("native string append returned %p and wrote %x", got, buf)
	}
}
