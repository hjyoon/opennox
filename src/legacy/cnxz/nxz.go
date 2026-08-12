package cnxz

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/icza/bitio"
	"github.com/opennox/libs/ifs"
)

const nxzSymbolCount = 274

type initialCode struct {
	bits   uint8
	offset uint16
}

var initialCodes = [...]initialCode{
	{bits: 2, offset: 0x00},
	{bits: 3, offset: 0x04},
	{bits: 3, offset: 0x0c},
	{bits: 4, offset: 0x14},
	{bits: 4, offset: 0x24},
	{bits: 4, offset: 0x34},
	{bits: 4, offset: 0x44},
	{bits: 4, offset: 0x54},
	{bits: 4, offset: 0x64},
	{bits: 4, offset: 0x74},
	{bits: 4, offset: 0x84},
	{bits: 4, offset: 0x94},
	{bits: 4, offset: 0xa4},
	{bits: 5, offset: 0xb4},
	{bits: 5, offset: 0xd4},
	{bits: 5, offset: 0xf4},
}

type symbolCode struct {
	prefix uint8
	bits   uint8
	value  uint16
}

var initialSymbolCodes = makeInitialSymbolCodes()

func makeInitialSymbolOrder() [nxzSymbolCount]int16 {
	var order [nxzSymbolCount]int16
	var seen [nxzSymbolCount]bool
	seen[0], seen[0x20], seen[0x30], seen[0xff] = true, true, true, true
	pos := 0
	for ; pos < 16; pos++ {
		order[pos] = int16(pos + 0x100)
		seen[order[pos]] = true
	}
	for _, sym := range []int{0, 0x20, 0x30, 0xff} {
		order[pos] = int16(sym)
		pos++
	}
	for sym := 1; sym <= 0x111; sym++ {
		if seen[sym] {
			continue
		}
		order[pos] = int16(sym)
		pos++
	}
	if pos != len(order) {
		panic(fmt.Sprintf("NXZ initial symbol table has %d entries, want %d", pos, len(order)))
	}
	return order
}

func makeInitialSymbolCodes() [nxzSymbolCount]symbolCode {
	order := makeInitialSymbolOrder()
	var out [nxzSymbolCount]symbolCode
	var found [nxzSymbolCount]bool
	for ind, raw := range order {
		sym := int(raw)
		for prefix, code := range initialCodes {
			start := int(code.offset)
			end := start + 1<<code.bits
			if ind < start || ind >= end {
				continue
			}
			out[sym] = symbolCode{
				prefix: uint8(prefix),
				bits:   code.bits,
				value:  uint16(ind - start),
			}
			found[sym] = true
			break
		}
	}
	for sym, ok := range found {
		if !ok {
			panic(fmt.Sprintf("NXZ initial codebook has no symbol for %#03x", sym))
		}
	}
	return out
}

func validatePaths(src, dst string) error {
	if src == "" {
		return errors.New("empty source path")
	}
	if dst == "" {
		return errors.New("empty destination path")
	}
	return nil
}

func CompressFile(src, dst string) (retErr error) {
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
	if fi.Size() < 0 || fi.Size() > math.MaxUint32 {
		return fmt.Errorf("NXZ input size %d is outside the uint32 format limit", fi.Size())
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

	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], uint32(fi.Size()))
	if _, err = w.Write(hdr[:]); err != nil {
		return err
	}
	return writeLiteralStream(w, r)
}

func writeLiteralStream(dst io.Writer, src io.Reader) error {
	bw := bitio.NewWriter(dst)
	var buf [32 * 1024]byte
	for {
		n, rerr := src.Read(buf[:])
		for _, b := range buf[:n] {
			code := initialSymbolCodes[b]
			if err := bw.WriteBits(uint64(code.prefix), 4); err != nil {
				return err
			}
			if err := bw.WriteBits(uint64(code.value), code.bits); err != nil {
				return err
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return rerr
		}
		if n == 0 {
			return io.ErrNoProgress
		}
	}
	end := initialSymbolCodes[273]
	if err := bw.WriteBits(uint64(end.prefix), 4); err != nil {
		return err
	}
	if err := bw.WriteBits(uint64(end.value), end.bits); err != nil {
		return err
	}
	return bw.Close()
}
