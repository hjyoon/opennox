package legacy

/*
#include "memfile.h"
*/
import "C"

import (
	"encoding/binary"
	"runtime"
	"unsafe"
)

// Wire words remain little-endian DWORD4. Only the memfile owner/cursor and
// tile/image sidecars are native pointers. Short reads consume available bytes
// to end, return zero, and retain completed writes; there is no rollback.
type thingParserCursor struct {
	f  *nox_memfile
	ok bool
}

func newThingParserCursor(f *nox_memfile) *thingParserCursor {
	s := &thingParserCursor{f: f}
	if f == nil || f.data == nil || f.size < 0 {
		return s
	}
	b, c, e := uintptr(unsafe.Pointer(f.data)), uintptr(unsafe.Pointer(f.cur)), uintptr(unsafe.Pointer(f.end))
	s.ok = e >= b && e-b == uintptr(uint32(f.size)) && c >= b && c <= e
	return s
}

func (s *thingParserCursor) take(n int) []byte {
	if !s.ok {
		return nil
	}
	if n < 0 || uintptr(n) > uintptr(unsafe.Pointer(s.f.end))-uintptr(unsafe.Pointer(s.f.cur)) {
		s.ok = false
		s.f.cur = s.f.end
		return nil
	}
	if n == 0 {
		return nil
	}
	p := unsafe.Pointer(s.f.cur)
	s.f.cur = (*C.char)(unsafe.Add(p, n))
	return unsafe.Slice((*byte)(p), n)
}
func (s *thingParserCursor) skip(n int) { s.take(n) }
func (s *thingParserCursor) u8() byte {
	b := s.take(1)
	if !s.ok {
		return 0
	}
	return b[0]
}
func (s *thingParserCursor) i16() int16 {
	b := s.take(2)
	if !s.ok {
		return 0
	}
	return int16(binary.LittleEndian.Uint16(b))
}
func (s *thingParserCursor) u32() uint32 {
	b := s.take(4)
	if !s.ok {
		return 0
	}
	return binary.LittleEndian.Uint32(b)
}
func (s *thingParserCursor) alignedByte() byte {
	if !s.ok {
		return 0
	}
	off := uintptr(unsafe.Pointer(s.f.cur)) - uintptr(unsafe.Pointer(s.f.data))
	s.skip(int((-off) & 7))
	b := s.take(8)
	if !s.ok {
		return 0
	}
	return b[0]
}
func (s *thingParserCursor) stringSkip() {
	n := s.u8()
	if s.ok {
		s.skip(int(n))
	}
}
func (s *thingParserCursor) imageRef() {
	v := s.u32()
	if !s.ok {
		return
	}
	if v == 0xffffffff {
		s.skip(1)
		s.stringSkip()
	}
}
func (s *thingParserCursor) result(v bool) int32 {
	if s.ok && v {
		return 1
	}
	return 0
}

// The unchanged public buffer ABI requires at least 256 writable bytes, as its
// BYTE lengths can request 255+NUL. Existing Go callers supply 256 KiB. No Go
// pointer is reconstructed from a wire DWORD. Empty nonterminated copies are lazy.
func thingParserScratch(p unsafe.Pointer, data []byte, nul bool) bool {
	n := len(data)
	if nul {
		n++
	}
	if n == 0 {
		return true
	}
	if p == nil {
		return false
	}
	out := unsafe.Slice((*byte)(p), n)
	copy(out, data)
	if nul {
		out[len(data)] = 0
	}
	runtime.KeepAlive(p)
	return true
}

func thingReadFloorNative411540(f *nox_memfile, scratch unsafe.Pointer) int32 {
	if *thingTileCountCell411540() >= 176 {
		return 0
	}
	p := thingTileOwner411540(*thingTileCountCell411540())
	s := newThingParserCursor(f)
	s.skip(4)
	n := s.u8()
	if !s.ok {
		return 0
	}
	// The old stack name is 32 bytes. Long BYTE names are a declared rejected
	// malformed domain, with the cursor immediately after the length byte.
	if n >= 32 {
		return 0
	}
	name := s.take(int(n))
	if !s.ok {
		return 0
	}
	ended := false
	for i := 0; i < 31; i++ {
		if i >= len(name) || name[i] == 0 {
			ended = true
		}
		if ended {
			p.NameBuf[i] = 0
		} else {
			p.NameBuf[i] = name[i]
		}
	}
	rgb := s.take(3)
	if !s.ok {
		return 0
	}
	if rgb[0] == 255 && rgb[1] == 0 && rgb[2] == 255 {
		p.Color48 = 0x80000000
	} else {
		// The real RGB endpoint narrows each original int argument to a BYTE;
		// do not read the old uninitialized cbuf high bytes merely to discard them.
		p.Color48 = thingParserColor411540(rgb[0], rgb[1], rgb[2])
	}
	p.Field58 = 0
	hasNUL := false
	for _, b := range p.NameBuf {
		if b == 0 {
			hasNUL = true
			break
		}
	}
	if !hasNUL {
		return 0
	} // Do not scan beyond an unterminated native name.
	if thingParserFacade411540(p) == 1 {
		p.Field58 |= 1
	}
	v := s.u32()
	if !s.ok {
		return 0
	}
	p.Field36 = int32(v)
	v = s.u32()
	if !s.ok {
		return 0
	}
	p.Field40 = int32(v)
	b := s.u8()
	if !s.ok {
		return 0
	}
	p.Field57 = b
	b = s.u8()
	if !s.ok {
		return 0
	}
	p.Field53 = b
	b = s.u8()
	if !s.ok {
		return 0
	}
	p.Field52 = b
	b = s.u8()
	if !s.ok {
		return 0
	}
	p.Field54 = b
	p.Field46 = 0
	b = s.u8()
	if !s.ok {
		return 0
	}
	p.Field55, p.Field56 = b, b
	p.Field44 = uint16(uint32(p.Field52) * uint32(p.Field53))
	frames := int(p.Field52) * int(p.Field53) * int(p.Field54)
	p.Data32 = nil
	for i := 0; i < frames; i++ {
		ref := s.u32()
		if !s.ok {
			return 0
		}
		if !thingParserScratch(scratch, []byte{thingFloorScratchByte411540()}, false) {
			return 0
		}
		if ref == 0xffffffff {
			s.skip(1)
			size := s.u8()
			data := s.take(int(size))
			if !s.ok {
				return 0
			}
			if !thingParserScratch(scratch, data, true) {
				return 0
			}
		}
	}
	(*thingTileCountCell411540())++
	runtime.KeepAlive(f)
	runtime.KeepAlive(scratch)
	return 1
}

func thingReadEdgeNative411850(f *nox_memfile, scratch unsafe.Pointer) int32 {
	i := *thingEdgeCountCell411850()
	if i >= 64 {
		return 0
	}
	s := newThingParserCursor(f)
	s.skip(4)
	n := s.u8()
	if !s.ok {
		return 0
	}
	if n >= 32 {
		return 0
	}
	name := s.take(int(n))
	if !s.ok {
		return 0
	}
	// strcpy stops at an embedded NUL, while the file cursor consumes all n bytes.
	end := len(name)
	for j, b := range name {
		if b == 0 {
			end = j
			break
		}
	}
	copy(thingEdgeBytes411850(i, 28644, end+1), append(append([]byte(nil), name[:end]...), 0))
	v := s.u32()
	if !s.ok {
		return 0
	}
	thingEdgeStore32411850(i, 28680, v)
	v = s.u32()
	if !s.ok {
		return 0
	}
	thingEdgeStore32411850(i, 28684, v)
	a := s.u8()
	if !s.ok {
		return 0
	}
	thingEdgeBytes411850(i, 28701, 1)[0] = a
	b := s.u8()
	if !s.ok {
		return 0
	}
	thingEdgeBytes411850(i, 28698, 1)[0] = b
	thingEdgeStore16411850(i, 28690, 0)
	c := s.u8()
	if !s.ok {
		return 0
	}
	thingEdgeBytes411850(i, 28699, 1)[0], thingEdgeBytes411850(i, 28700, 1)[0] = c, c
	flag := s.u8()
	if !s.ok || flag == 1 {
		return 0
	}
	d := s.u8()
	if !s.ok {
		return 0
	}
	thingEdgeBytes411850(i, 28697, 1)[0] = d
	e := s.u8()
	if !s.ok {
		return 0
	}
	thingEdgeBytes411850(i, 28696, 1)[0] = e
	thingEdgeStore16411850(i, 28688, uint16(2*(uint32(d)+uint32(e))))
	thingEdgeStore32411850(i, 28676, 0)
	*thingEdgeImageCell411850(i) = nil
	for j := 0; j < 2*int(b)*(int(d)+int(e)); j++ {
		ref := s.u32()
		if !s.ok {
			return 0
		}
		if !thingParserScratch(scratch, []byte{thingEdgeScratchByte411850()}, false) {
			return 0
		}
		if ref == 0xffffffff {
			s.skip(1)
			n := s.u8()
			data := s.take(int(n))
			if !s.ok {
				return 0
			}
			if !thingParserScratch(scratch, data, true) {
				return 0
			}
		}
	}
	if s.u32() != 0x454e4420 || !s.ok {
		return 0
	}
	(*thingEdgeCountCell411850())++
	runtime.KeepAlive(f)
	runtime.KeepAlive(scratch)
	return 1
}

func thingSkipAudioNative414D40(f *nox_memfile) int32 {
	s := newThingParserCursor(f)
	count := int32(s.u32())
	if !s.ok {
		return 0
	}
	if count <= 0 {
		return 1
	}
	for i := int32(0); i < count; i++ {
		n := s.u8()
		s.skip(int(n) + 9)
		for {
			n = s.u8()
			if !s.ok {
				return 0
			}
			if n == 0 {
				break
			}
			s.skip(int(n))
		}
	}
	return s.result(true)
}
func thingReadFloorNative414DB0(f *nox_memfile) int32 {
	s := newThingParserCursor(f)
	s.skip(4)
	n := s.u8()
	s.skip(int(n) + 12)
	a, b, c := s.u8(), s.u8(), s.u8()
	s.skip(1)
	for i := 0; i < int(a)*int(b)*int(c); i++ {
		s.imageRef()
		if !s.ok {
			return 0
		}
	}
	return s.result(s.u32() == 0x454e4420)
}
func thingReadEdgeNative414E70(f *nox_memfile, scratch unsafe.Pointer) int32 {
	s := newThingParserCursor(f)
	s.skip(4)
	n := s.u8()
	data := s.take(int(n))
	if !s.ok {
		return 0
	}
	if !thingParserScratch(scratch, data, false) {
		return 0
	}
	s.skip(9)
	count := s.u8()
	s.skip(1)
	flag := s.u8()
	if !s.ok || flag == 1 {
		return 0
	}
	a, b := s.u8(), s.u8()
	for i := 0; i < 2*int(count)*(int(a)+int(b)); i++ {
		s.imageRef()
		if !s.ok {
			return 0
		}
	}
	return s.result(s.u32() == 0x454e4420)
}
func thingReadWallNative414F60(f *nox_memfile, scratch unsafe.Pointer) int32 {
	s := newThingParserCursor(f)
	readName := func() bool {
		n := s.u8()
		data := s.take(int(n))
		return s.ok && thingParserScratch(scratch, data, true)
	}
	s.skip(4)
	if !readName() {
		return 0
	}
	s.skip(14)
	groups := s.alignedByte()
	for i := 0; i < int(groups); i++ {
		if i >= 8 {
			return 0
		}
		if !readName() {
			return 0
		}
	}
	for i := 0; i < 3; i++ {
		if !readName() {
			return 0
		}
	}
	s.skip(1)
	for i := 0; i < 15; i++ {
		records := s.alignedByte()
		if !s.ok {
			return 0
		}
		for j := 0; j < int(records); j++ {
			for k := 0; k < 4; k++ {
				s.skip(8)
				s.imageRef()
				if !s.ok {
					return 0
				}
			}
		}
	}
	return s.result(s.u32() == 0x454e4420)
}
func thingSkipSpellsNative415100(f *nox_memfile) int32 {
	s := newThingParserCursor(f)
	count := int32(s.u32())
	if !s.ok {
		return 0
	}
	if count <= 0 {
		return 1
	}
	for i := int32(0); i < count; i++ {
		n := s.u8()
		s.skip(int(n) + 3)
		s.stringSkip()
		for j := 0; j < 2; j++ {
			s.imageRef()
		}
		s.skip(4)
		s.stringSkip()
		next := s.i16()
		s.skip(int(next))
		for j := 0; j < 3; j++ {
			s.stringSkip()
		}
		if !s.ok {
			return 0
		}
	}
	return 1
}
func thingReadImageNative415240(f *nox_memfile) int32 {
	s := newThingParserCursor(f)
	count := int32(s.u32())
	if !s.ok {
		return 0
	}
	if count <= 0 {
		return 1
	}
	for i := int32(0); i < count; i++ {
		s.stringSkip()
		kind := s.u8()
		refs := byte(1)
		if kind == 2 {
			refs = s.u8()
			s.skip(1)
			s.stringSkip()
		}
		for j := 0; j < int(refs); j++ {
			s.imageRef()
			if !s.ok {
				return 0
			}
		}
		if !s.ok {
			return 0
		}
	}
	return 1
}
func thingReadAbilityNative415320(f *nox_memfile) int32 {
	s := newThingParserCursor(f)
	count := int32(s.u32())
	if !s.ok {
		return 0
	}
	if count <= 0 {
		return 1
	}
	for i := int32(0); i < count; i++ {
		n := s.u8()
		s.skip(int(n) + 1)
		for j := 0; j < 3; j++ {
			s.imageRef()
		}
		s.stringSkip()
		next := s.i16()
		s.skip(int(next))
		for j := 0; j < 3; j++ {
			s.stringSkip()
		}
		if !s.ok {
			return 0
		}
	}
	return 1
}
func thingReadAudioNative415660(f *nox_memfile, scratch unsafe.Pointer) int32 {
	s := newThingParserCursor(f)
	count := int32(s.u32())
	if !s.ok {
		return 0
	}
	if count <= 0 {
		return 1
	}
	for i := int32(0); i < count; i++ {
		// The existing external sound helper stays real. Preflight a complete
		// serialized record on a private CURSOR copy, never a copied payload.
		// On malformed/truncated records do not enter its unchecked C reads.
		probeFile := *f
		probe := newThingParserCursor(&probeFile)
		probe.stringSkip()
		probe.skip(9)
		for {
			n := probe.u8()
			if !probe.ok {
				f.cur = f.end
				return 0
			}
			if n == 0 {
				break
			}
			probe.skip(int(n))
		}
		if scratch == nil {
			return 0
		}
		if thingParserAudio415660(f, scratch) == 0 {
			return 0
		}
		s = newThingParserCursor(f)
		if !s.ok {
			return 0
		}
	}
	runtime.KeepAlive(f)
	runtime.KeepAlive(scratch)
	return 1
}
