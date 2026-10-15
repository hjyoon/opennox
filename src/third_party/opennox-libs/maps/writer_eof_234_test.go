package maps_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"io"
	"testing"

	"github.com/opennox/libs/maps"
	crypt "github.com/opennox/noxcrypt"
)

type memory234 struct {
	data               []byte
	writes, patches    int
	writeErr, patchErr error
}

func (m *memory234) Write(p []byte) (int, error) {
	m.writes++
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	m.data = append(m.data, p...)
	return len(p), nil
}
func (m *memory234) WriteAt(p []byte, off int64) (int, error) {
	m.patches++
	if m.patchErr != nil {
		return 0, m.patchErr
	}
	if off < 0 || off > int64(len(m.data)) {
		return 0, io.ErrShortWrite
	}
	n := copy(m.data[int(off):], p)
	if n != len(p) {
		return n, io.ErrShortWrite
	}
	return n, nil
}

// Check decrypted wire bytes directly. maps.Reader accepts physical EOF, so
// using it as the oracle would not catch an omitted zero-length section name.
func requireWireEOF234(t *testing.T, input []byte, magic uint32, sections []maps.RawSection) {
	t.Helper()
	r, err := crypt.NewReader(bytes.NewReader(input), crypt.MapKey)
	if err != nil {
		t.Fatal(err)
	}
	p, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(p) < 16 || binary.LittleEndian.Uint32(p) != magic {
		t.Fatal("map header changed")
	}
	off := 4
	if magic == maps.Magic {
		off = 16
		if len(p) < 24 {
			t.Fatal("short checksum header")
		}
		// Independent bitwise form of the existing per-block checksum. The
		// checksum placeholder contributes an all-zero block, not its value.
		crc := uint32(0xffffffff)
		for at := 0; at < len(p); at += 8 {
			for i := 0; i < 8; i++ {
				v := p[at+i]
				if at == 8 {
					v = 0
				}
				crc ^= uint32(v)
				for bit := 0; bit < 8; bit++ {
					if crc&1 != 0 {
						crc = crc>>1 ^ 0xedb88320
					} else {
						crc >>= 1
					}
				}
			}
			crc = ^crc
		}
		if got := binary.LittleEndian.Uint32(p[8:]); got != crc {
			t.Fatalf("CRC excludes final bytes: got %08x want %08x", got, crc)
		}
	}
	if binary.LittleEndian.Uint32(p[off:]) != 7 || binary.LittleEndian.Uint32(p[off+4:]) != 11 {
		t.Fatal("wall offsets changed")
	}
	off += 8
	for _, want := range sections {
		if off >= len(p) {
			t.Fatal("missing section name")
		}
		n := int(p[off])
		off++
		if n != len(want.Name)+1 || off+n > len(p) || string(p[off:off+n-1]) != want.Name || p[off+n-1] != 0 {
			t.Fatal("section name changed")
		}
		off = (off + n + 7) &^ 7
		if off+8 > len(p) || binary.LittleEndian.Uint32(p[off:]) != uint32(len(want.Data)) || binary.LittleEndian.Uint32(p[off+4:]) != 0 {
			t.Fatal("section length changed")
		}
		off += 8
		if off+len(want.Data) > len(p) || !bytes.Equal(p[off:off+len(want.Data)], want.Data) {
			t.Fatal("section payload changed")
		}
		off += len(want.Data)
	}
	if off >= len(p) {
		t.Fatal("missing EOF zero-length section name at crypt block boundary")
	}
	if p[off] != 0 {
		t.Fatal("EOF section name is not zero")
	}
	if len(p) != (off+1+7)&^7 {
		t.Fatalf("extra/missing EOF block: payload end=%d file=%d", off, len(p))
	}
	for _, v := range p[off:] {
		if v != 0 {
			t.Fatal("nonzero bytes after EOF marker")
		}
	}
}

func TestMapWriterZeroNameEOF234(t *testing.T) {
	for _, magic := range []uint32{maps.MagicOld, maps.Magic} {
		for flushes := 0; flushes < 3; flushes++ {
			for n := 0; n < 32; n++ {
				t.Run(fmt.Sprintf("%08x/flush%d/len%02d", magic, flushes, n), func(t *testing.T) {
					sections := []maps.RawSection{{Name: "First234", Data: []byte{11, 22, 33}}, {Name: "Last", Data: make([]byte, n)}}
					for i := range sections[1].Data {
						sections[1].Data[i] = byte(i*37 + 11)
					}
					var out memory234
					w, err := maps.NewWriter(&out, maps.Header{Magic: magic, Offs: image.Pt(7, 11)})
					if err != nil {
						t.Fatal(err)
					}
					if err = w.WriteRawSections(sections); err != nil {
						t.Fatal(err)
					}
					for i := 0; i < flushes; i++ {
						if err = w.Flush(); err != nil {
							t.Fatal(err)
						}
					}
					if err = w.Close(); err != nil {
						t.Fatal(err)
					}
					requireWireEOF234(t, out.data, magic, sections)
					before := append([]byte(nil), out.data...)
					if err = w.Close(); err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(out.data, before) {
						t.Fatal("repeated Close appended a second EOF marker")
					}
				})
			}
		}
	}
}

func TestMapWriterEmptyEOF234(t *testing.T) {
	for _, magic := range []uint32{maps.MagicOld, maps.Magic} {
		for flushes := 0; flushes < 3; flushes++ {
			t.Run(fmt.Sprintf("%08x/flush%d", magic, flushes), func(t *testing.T) {
				var out memory234
				w, err := maps.NewWriter(&out, maps.Header{Magic: magic, Offs: image.Pt(7, 11)})
				if err != nil {
					t.Fatal(err)
				}
				for i := 0; i < flushes; i++ {
					if err = w.Flush(); err != nil {
						t.Fatal(err)
					}
				}
				if err = w.Close(); err != nil {
					t.Fatal(err)
				}
				requireWireEOF234(t, out.data, magic, nil)
				before := append([]byte(nil), out.data...)
				if err = w.Close(); err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(out.data, before) {
					t.Fatal("empty map repeated Close changed bytes")
				}
			})
		}
	}
}
