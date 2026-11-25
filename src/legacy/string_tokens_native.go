package legacy

import (
	"strings"
	"unsafe"
)

// legacyByteTokens scans a NUL-terminated buffer in place. Each call can use
// different delimiters, as the original narrow string tokenizer permits.
type legacyByteTokens struct {
	next *byte
}

func (t *legacyByteTokens) take(delimiters string) *byte {
	p := t.next
	if p == nil {
		return nil
	}
	for *p != 0 && strings.IndexByte(delimiters, *p) >= 0 {
		p = (*byte)(unsafe.Add(unsafe.Pointer(p), 1))
	}
	if *p == 0 {
		t.next = nil
		return nil
	}
	start := p
	for *p != 0 && strings.IndexByte(delimiters, *p) < 0 {
		p = (*byte)(unsafe.Add(unsafe.Pointer(p), 1))
	}
	t.next = nil
	if *p != 0 {
		*p = 0
		t.next = (*byte)(unsafe.Add(unsafe.Pointer(p), 1))
	}
	return start
}
