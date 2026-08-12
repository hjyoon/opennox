package cnxz

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"math/bits"

	"github.com/opennox/libs/ifs"
)

const nxzWindowSize = 64 * 1024

var nxzLengthTable = [...]struct {
	bits   uint8
	offset uint16
}{
	{bits: 1, offset: 0x008},
	{bits: 2, offset: 0x00a},
	{bits: 3, offset: 0x00e},
	{bits: 4, offset: 0x016},
	{bits: 5, offset: 0x026},
	{bits: 6, offset: 0x046},
	{bits: 7, offset: 0x086},
	{bits: 8, offset: 0x106},
}

var nxzDistanceTable = [...]struct {
	bits   uint8
	offset uint8
}{
	{},
	{offset: 0x01},
	{bits: 1, offset: 0x02},
	{bits: 2, offset: 0x04},
	{bits: 3, offset: 0x08},
	{bits: 4, offset: 0x10},
	{bits: 5, offset: 0x20},
	{bits: 6, offset: 0x40},
}

// DecompressFile decodes the original NXZ bitstream without relying on the
// pointer-sized C translation. The decoder is based on the repository's
// historical custom-map fix (ecd645ab0), with fixed-width format tables and a
// Go-sized ring-buffer index so that the same code works on 32- and 64-bit
// targets.
func DecompressFile(src, dst string) (retErr error) {
	if err := validatePaths(src, dst); err != nil {
		return err
	}
	r, err := ifs.Open(src)
	if err != nil {
		return err
	}
	defer r.Close()

	fi, err := r.Stat()
	if err != nil {
		return err
	}
	if fi.Size() < 4 {
		return fmt.Errorf("NXZ input is too short: %d bytes", fi.Size())
	}
	maxInt := uint64(^uint(0) >> 1)
	compressedSize := uint64(fi.Size() - 4)
	if compressedSize > maxInt {
		return fmt.Errorf("NXZ compressed size %d exceeds this architecture's int limit", compressedSize)
	}

	var hdr [4]byte
	if _, err = io.ReadFull(r, hdr[:]); err != nil {
		return err
	}
	decodedSize := uint64(binary.LittleEndian.Uint32(hdr[:]))
	if decodedSize > maxInt {
		return fmt.Errorf("NXZ decoded size %d exceeds this architecture's int limit", decodedSize)
	}

	compressed := make([]byte, int(compressedSize))
	if _, err = io.ReadFull(r, compressed); err != nil {
		return err
	}
	decoded := make([]byte, int(decodedSize))
	dec := newNXZDecoder(compressed)
	written := 0
	for written < len(decoded) {
		sourceBefore := len(dec.src)
		n, err := dec.decodeChunk(decoded[written:])
		if err != nil {
			return err
		}
		written += n
		if n == 0 && len(dec.src) == sourceBefore {
			return errors.New("NXZ decoder made no progress")
		}
	}
	if written != len(decoded) {
		return fmt.Errorf("NXZ decoded size mismatch: got %d, want %d", written, len(decoded))
	}

	w, err := ifs.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if err := w.Close(); retErr == nil {
			retErr = err
		}
	}()
	_, err = w.Write(decoded)
	return err
}

type nxzDecoder struct {
	src     []byte
	bits    nxzBitReader
	window  [nxzWindowSize]byte
	windowI int
	counts  [nxzSymbolCount]int16
	symbols [nxzSymbolCount]int16
	codes   [16]initialCode
}

type nxzDecoderRecord struct {
	index int16
	count int16
}

func newNXZDecoder(src []byte) *nxzDecoder {
	d := &nxzDecoder{
		src:     src,
		symbols: makeInitialSymbolOrder(),
		codes:   initialCodes,
	}
	d.bits.readByte = d.readByte
	return d
}

func (d *nxzDecoder) readByte() (byte, bool) {
	if len(d.src) == 0 {
		return 0, false
	}
	b := d.src[0]
	d.src = d.src[1:]
	return b, true
}

func (d *nxzDecoder) rebuild(records []nxzDecoderRecord) {
	for i := range d.counts {
		records[i] = nxzDecoderRecord{index: int16(i), count: d.counts[i]}
		d.counts[i] /= 2
	}
	// Preserve the original shell-sort ordering: descending frequency, then
	// descending symbol number. The tie-break is part of the file format.
	for gap := 40; gap > 0; gap /= 3 {
		for i, cur := range records {
			j := i + 1
			for ; j > gap; j -= gap {
				prev := records[j-gap-1]
				delta := int(prev.count) - int(cur.count)
				if delta == 0 {
					delta = int(prev.index) - int(cur.index)
				}
				if delta >= 0 {
					break
				}
				records[j-1] = prev
			}
			records[j-1] = cur
		}
	}
	for i := range d.symbols {
		d.symbols[i] = records[i].index
	}
}

func (d *nxzDecoder) decodeChunk(dst []byte) (int, error) {
	var records [nxzSymbolCount]nxzDecoderRecord
	written := 0
	d.bits.reset()
	for {
		// Several original/custom NXZ files omit the end opcode. When the last
		// source byte has been consumed, the size header is the terminator.
		if len(d.src) == 0 {
			return written, nil
		}
		prefix := int(d.bits.readBits(4))
		if prefix < 0 || prefix >= len(d.codes) {
			return 0, errors.New("NXZ invalid prefix")
		}
		code := d.codes[prefix]
		index := int(code.offset)
		if code.bits != 0 {
			index += int(d.bits.readBits(int(code.bits)))
		}
		if d.bits.failed || index < 0 || index >= len(d.symbols) {
			return 0, fmt.Errorf("NXZ invalid symbol index %d", index)
		}

		symbol := int(d.symbols[index])
		d.counts[symbol]++
		switch {
		case symbol < 256:
			if written >= len(dst) {
				return 0, io.ErrShortBuffer
			}
			dst[written] = byte(symbol)
			d.window[d.windowI%nxzWindowSize] = byte(symbol)
			d.windowI++
			written++
			continue
		case symbol == 272:
			d.rebuild(records[:])
			bitOffset, symbolOffset := 0, 0
			for i := range d.codes {
				zeroes := 0
				for d.bits.readBits(1) == 0 {
					zeroes++
					if d.bits.failed || zeroes > 31 {
						return 0, errors.New("NXZ invalid rebuilt code table")
					}
				}
				bitOffset += zeroes
				if bitOffset > 31 {
					return 0, errors.New("NXZ rebuilt code width is too large")
				}
				d.codes[i] = initialCode{bits: uint8(bitOffset), offset: uint16(symbolOffset)}
				symbolOffset += 1 << bitOffset
			}
			continue
		case symbol == 273:
			return written, nil
		}

		length := 4
		if symbol < 264 {
			length += symbol - 256
		} else {
			entry := nxzLengthTable[symbol-264]
			length += int(entry.offset) + int(d.bits.readBits(int(entry.bits)))
			if d.bits.failed {
				return 0, io.ErrUnexpectedEOF
			}
		}

		distanceCode := int(d.bits.readBits(3))
		if d.bits.failed || distanceCode >= len(nxzDistanceTable) {
			return 0, errors.New("NXZ invalid distance code")
		}
		entry := nxzDistanceTable[distanceCode]
		distance := (int(entry.offset) << 9) + int(d.bits.readBits(int(entry.bits)+9))
		if d.bits.failed {
			return 0, io.ErrUnexpectedEOF
		}
		if distance < 0 || distance > nxzWindowSize || distance > d.windowI {
			return 0, fmt.Errorf("NXZ invalid back-reference distance %d at output %d", distance, d.windowI)
		}
		if length > len(dst)-written {
			return 0, io.ErrShortBuffer
		}

		readI := (d.windowI - distance) % nxzWindowSize
		for i := 0; i < length; i++ {
			b := d.window[(readI+i)%nxzWindowSize]
			dst[written+i] = b
			d.window[d.windowI%nxzWindowSize] = b
			d.windowI++
		}
		written += length
	}
}

type nxzBitReader struct {
	readByte func() (byte, bool)
	buf      uint
	count    int
	failed   bool
}

func (r *nxzBitReader) reset() {
	r.buf = 0
	r.count = 0
	r.failed = false
}

func (r *nxzBitReader) readBits(n int) uint {
	if n == 0 {
		return 0
	}
	for r.count < n {
		b, ok := r.readByte()
		if !ok {
			r.failed = true
			return math.MaxUint
		}
		r.buf |= uint(b) << (bits.UintSize - 8 - r.count)
		r.count += 8
	}
	value := r.buf >> (bits.UintSize - n)
	r.buf <<= n
	r.count -= n
	return value
}
