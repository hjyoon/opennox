package noxscriptwire

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"
)

type testMetadata map[byte][]uint32

func (m testMetadata) ArgCount(function byte) byte { return byte(len(m[function])) }
func (m testMetadata) ArgKind(function byte, argument byte) uint32 {
	return m[function][argument]
}

func appendWord(dst []byte, value uint32) []byte {
	var word [4]byte
	binary.LittleEndian.PutUint32(word[:], value)
	return append(dst, word[:]...)
}

func TestScanAllArgumentKinds(t *testing.T) {
	kinds := []uint32{0, 1, 2, 3, 4, 5, 6, 7, 8}
	wire := appendWord(nil, 3)
	wire = append(wire, 'n', 'o', 'x')
	wire = appendWord(wire, 0xa1b2c3d4)
	wire = appendWord(wire, uint32(len(kinds)))
	meta := make(testMetadata)
	for i, kind := range kinds {
		wire = append(wire, byte(i), 0x91) // function ID and ignored byte
		meta[byte(i)] = []uint32{kind}
		switch kind {
		case 0, 3, 4, 5, 6:
			wire = append(wire, 1, 2, 3, 4)
		case 1:
			wire = append(wire, 1, 2, 3, 4, 5, 6, 7, 8)
		case 2, 7:
			wire = append(wire, 3, 0x80, 0x81, 0x82)
		}
	}
	reader := bytes.NewReader(append(wire, 0xee))
	flags := uint32(0x11223344)
	if got := Scan(reader, meta, &flags); got != int32(len(kinds)) {
		t.Fatalf("count = %d, want %d", got, len(kinds))
	}
	if flags != 0xa1b2c3d4 {
		t.Fatalf("flags = %#x", flags)
	}
	if off, _ := reader.Seek(0, io.SeekCurrent); off != int64(len(wire)) {
		t.Fatalf("cursor = %d, want %d", off, len(wire))
	}
}

func TestScanMultipleArgumentsAndUnsignedStringLength(t *testing.T) {
	wire := appendWord(nil, 0)
	wire = appendWord(wire, 0x12345678)
	wire = appendWord(wire, 1)
	wire = append(wire, 0xff, 0x35)
	wire = append(wire, 0x80)
	wire = append(wire, bytes.Repeat([]byte{0xaa}, 128)...)
	wire = append(wire, 1, 2, 3, 4)
	reader := bytes.NewReader(append(wire, 0xee))
	flags := uint32(0)
	if got := Scan(reader, testMetadata{0xff: {7, 6}}, &flags); got != 1 {
		t.Fatalf("count = %d", got)
	}
	if off, _ := reader.Seek(0, io.SeekCurrent); off != int64(len(wire)) {
		t.Fatalf("cursor = %d, want %d", off, len(wire))
	}
}

func TestScanOversizedNameIsDiscardedWithoutOverwritingFlags(t *testing.T) {
	wire := appendWord(nil, 1025)
	wire = append(wire, bytes.Repeat([]byte{0x5a}, 1025)...)
	wire = appendWord(wire, 0xdeadbeef)
	wire = appendWord(wire, 0)
	reader := bytes.NewReader(wire)
	flags := uint32(0x11223344)
	if got := Scan(reader, testMetadata{}, &flags); got != 0 || flags != 0xdeadbeef {
		t.Fatalf("count = %d, flags = %#x", got, flags)
	}
	if off, _ := reader.Seek(0, io.SeekCurrent); off != int64(len(wire)) {
		t.Fatalf("cursor = %d, want %d", off, len(wire))
	}
}

func TestScanNegativeLengthsAndPartialFlags(t *testing.T) {
	t.Run("negative name", func(t *testing.T) {
		reader := bytes.NewReader(appendWord(nil, ^uint32(0)))
		flags := uint32(0x11223344)
		if got := Scan(reader, testMetadata{}, &flags); got != 0 || flags != 0x11223344 {
			t.Fatalf("count = %d, flags = %#x", got, flags)
		}
	})
	t.Run("negative record count", func(t *testing.T) {
		wire := appendWord(nil, 0)
		wire = appendWord(wire, 0x01020304)
		wire = appendWord(wire, ^uint32(0))
		reader := bytes.NewReader(wire)
		flags := uint32(0)
		if got := Scan(reader, testMetadata{}, &flags); got != -1 || flags != 0x01020304 {
			t.Fatalf("count = %d, flags = %#x", got, flags)
		}
	})
	t.Run("partial flags", func(t *testing.T) {
		reader := bytes.NewReader([]byte{0, 0, 0, 0, 0xaa, 0xbb})
		flags := uint32(0x11223344)
		if got := Scan(reader, testMetadata{}, &flags); got != 0 || flags != 0x1122bbaa {
			t.Fatalf("count = %d, flags = %#x", got, flags)
		}
	})
	t.Run("truncated huge record count", func(t *testing.T) {
		wire := appendWord(nil, 0)
		wire = appendWord(wire, 0xaabbccdd)
		wire = appendWord(wire, 0x7fffffff)
		reader := bytes.NewReader(wire)
		flags := uint32(0)
		if got := Scan(reader, testMetadata{}, &flags); got != 0x7fffffff || flags != 0xaabbccdd {
			t.Fatalf("count = %d, flags = %#x", got, flags)
		}
	})
}
