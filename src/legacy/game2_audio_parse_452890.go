package legacy

import (
	"encoding/binary"
	"fmt"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"strings"
	"unsafe"
)

// The native client loader already decodes these sections for its platform FX
// implementation. These taps keep the public event-service scalar owner in
// sync without changing the caller's stream position or retaining its buffer.
// Root wires them into nativeAudioFX.readAUD/readAVNT after successful parsing.
func Game2AudioResetDefinitions() {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	if !game2Audio.Initialized {
		return
	}
	game2AudioClear4521F0()
	game2AudioBucketReset452010()
	for i := range game2Audio.Definitions {
		d := &game2Audio.Definitions[i]
		game2AudioReset451920(d.Pointer)
		d.Ranked = nil
		d.BucketLinked = false
	}
}
func Game2AudioShutdownNative() {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	game2AudioShutdown451970()
}
func Game2AudioReadAVNTData(data []byte) (int, error) {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	if !game2Audio.Initialized {
		return 0, nil
	}
	p, _ := alloc.Malloc(uintptr(len(data) + 1))
	if p == nil {
		return 0, fmt.Errorf("AVNT mirror allocation")
	}
	copy(unsafe.Slice((*byte)(p), len(data)), data)
	f := binfile.NewMemFile(p, len(data))
	defer f.Free()
	if game2AudioParseAVNT452890(f, nil) == 0 {
		return len(data) - len(f.Data()), fmt.Errorf("AVNT event definition")
	}
	return len(data) - len(f.Data()), nil
}
func Game2AudioReadAUDData(data []byte) (int, error) {
	game2AudioMu.Lock()
	defer game2AudioMu.Unlock()
	if !game2Audio.Initialized {
		return 0, nil
	}
	p, _ := alloc.Malloc(uintptr(len(data) + 1))
	if p == nil {
		return 0, fmt.Errorf("AUD mirror allocation")
	}
	copy(unsafe.Slice((*byte)(p), len(data)), data)
	f := binfile.NewMemFile(p, len(data))
	defer f.Free()
	r := game2AudioReader{f: f}
	count := int32(r.u32())
	if r.bad {
		return len(data) - len(f.Data()), fmt.Errorf("AUD count")
	}
	for i := int32(0); i < count; i++ {
		if game2AudioParseAUD452BD0(f, nil) == 0 {
			return len(data) - len(f.Data()), fmt.Errorf("AUD event definition %d", i)
		}
	}
	return len(data) - len(f.Data()), nil
}

type game2AudioReader struct {
	f          *binfile.MemFile
	bad        bool
	scratch    unsafe.Pointer
	lastLength int
}

func (r *game2AudioReader) take(n int) []byte {
	if r.bad || r.f == nil {
		return nil
	}
	d := r.f.Data()
	if len(d) < n {
		r.f.Skip(len(d))
		r.bad = true
		return nil
	}
	p := d[:n]
	r.f.Skip(n)
	return p
}
func (r *game2AudioReader) u8() uint8 {
	p := r.take(1)
	if p == nil {
		return 0
	}
	return p[0]
}
func (r *game2AudioReader) i8() int8 { return int8(r.u8()) }
func (r *game2AudioReader) i16() int16 {
	p := r.take(2)
	if p == nil {
		return 0
	}
	return int16(binary.LittleEndian.Uint16(p))
}
func (r *game2AudioReader) u32() uint32 {
	p := r.take(4)
	if p == nil {
		return 0
	}
	return binary.LittleEndian.Uint32(p)
}
func (r *game2AudioReader) name() string {
	return r.nameBytes(int(r.u8()))
}
func (r *game2AudioReader) sampleName() string {
	n := int(r.u8())
	if n == 0 {
		// The list terminator is consumed before the C parser touches scratch.
		// An initial empty sound name still writes NUL through name().
		r.lastLength = 0
		return ""
	}
	return r.nameBytes(n)
}
func (r *game2AudioReader) nameBytes(n int) string {
	r.lastLength = n
	p := r.take(n)
	if r.bad {
		return ""
	}
	if r.scratch != nil {
		dst := unsafe.Slice((*byte)(r.scratch), n+1)
		copy(dst, p)
		dst[n] = 0
	}
	// C string consumers stop at the first NUL even when the byte-counted
	// record contains more bytes. Cursor consumption remains the full length.
	for i, c := range p {
		if c == 0 {
			p = p[:i]
			break
		}
	}
	return string(p)
}
func game2AudioParseAVNT452890(f *binfile.MemFile, scratch unsafe.Pointer) int32 {
	r := game2AudioReader{f: f, scratch: scratch}
	name := r.name()
	if r.bad {
		return 0
	}
	id := int32(sound.ByName(name))
	p := game2AudioLookup452270(id)
	if id == 0 || p == nil {
		game2AudioSkipInner(&r)
		if r.bad {
			return 0
		}
		return 1
	}
	d := game2AudioDef(p)
	for !r.bad {
		op := r.u8()
		if r.bad {
			return 0
		}
		switch op {
		case 0:
			d.set(0, 1)
			return 1
		case 1:
			v := r.u8()
			if !r.bad {
				d.set(12, uint32(v))
			}
		case 2:
			v := r.u8()
			if !r.bad {
				d.set(1, uint32(v))
			}
		case 3:
			v := r.u8()
			if !r.bad {
				d.volume().Init(163 * int32(v))
			}
		case 4:
			v := r.u8()
			if !r.bad {
				d.set(14, uint32(v))
			}
		case 5:
			v := r.u8()
			if !r.bad {
				d.set(15, uint32(v))
			}
		case 6:
			a, b := r.i8(), r.i8()
			if !r.bad {
				d.set(19, uint32(int32(a)))
				d.set(20, uint32(int32(b)))
			}
		case 7:
			n := 0
			for !r.bad {
				name := r.sampleName()
				if r.bad {
					return 0
				}
				if r.lastLength == 0 {
					break
				}
				if n < 32 {
					id := game2AudioSampleLookup(name)
					d.setSample(n, id)
					if id != -1 {
						n++
					}
				}
			}
			d.set(48, uint32(n))
		case 8:
			a, b := r.u32(), r.u32()
			if !r.bad {
				d.set(17, a)
				d.set(18, b)
			}
		case 9:
			v := r.i16()
			if !r.bad && v > 0 {
				d.set(16, 15*uint32(v))
			}
		case 10:
			v := r.i16()
			if !r.bad {
				d.set(2, uint32(int32(v)))
			}
		default:
			return 0
		}
	}
	return 0
}
func game2AudioSkipInner(r *game2AudioReader) int32 {
	for !r.bad {
		op := r.u8()
		if r.bad {
			return 0
		}
		switch op {
		case 0:
			return 1
		case 1, 2, 3, 4, 5:
			r.take(1)
		case 6, 9, 10:
			r.take(2)
		case 7:
			for !r.bad {
				n := int(r.u8())
				if n == 0 {
					break
				}
				r.take(n)
			}
		case 8:
			r.take(8)
		default:
			return 0
		}
	}
	return 0
}
func game2AudioSkip452B30(f *binfile.MemFile) int32 {
	return game2AudioSkipInner(&game2AudioReader{f: f})
}
func game2AudioSkip452B00(f *binfile.MemFile) int32 {
	r := game2AudioReader{f: f}
	n := int(r.u8())
	r.take(n)
	return game2AudioSkipInner(&r)
}
func game2AudioParseAUD452BD0(f *binfile.MemFile, scratch unsafe.Pointer) int32 {
	r := game2AudioReader{f: f, scratch: scratch}
	name := r.name()
	if r.bad {
		return 0
	}
	id := int32(sound.ByName(name))
	p := game2AudioLookup452270(id)
	if id == 0 || p == nil {
		r.take(9)
		for !r.bad {
			n := int(r.u8())
			if n == 0 {
				return game2AudioBool(!r.bad)
			}
			r.take(n)
		}
		return 0
	}
	d := game2AudioDef(p)
	d.set(1, 2)
	priority := r.i16()
	if r.bad {
		return 0
	}
	d.set(2, uint32(int32(priority)))
	gain := r.u8()
	if r.bad {
		return 0
	}
	d.volume().Init(163 * int32(gain))
	duration := r.i16()
	if r.bad {
		return 0
	}
	if duration > 0 {
		d.set(16, 15*uint32(duration))
	}
	for _, field := range []int{14, 19, 20} {
		v := r.i8()
		if r.bad {
			return 0
		}
		d.set(field, uint32(int32(v)))
	}
	mode := r.i8()
	if r.bad {
		return 0
	}
	d.set(12, uint32(int32(mode)))
	if mode >= 3 {
		return 0
	}
	count := 0
	for !r.bad {
		name := r.sampleName()
		if r.bad {
			return 0
		}
		if r.lastLength == 0 {
			break
		}
		if i := strings.LastIndexByte(name, '.'); i >= 0 {
			name = name[:i]
			if scratch != nil {
				dst := unsafe.Slice((*byte)(scratch), len(name)+1)
				dst[len(name)] = 0
			}
		}
		if count < 32 {
			id := game2AudioSampleLookup(name)
			d.setSample(count, id)
			if id != -1 {
				count++
			}
		}
		// More than32 valid names formerly overwrote adjacent definitions.
		// Consume them without storing or issuing further bank lookups.
	}
	if r.bad {
		return 0
	}
	d.set(0, 1)
	d.set(48, uint32(count))
	return 1
}
