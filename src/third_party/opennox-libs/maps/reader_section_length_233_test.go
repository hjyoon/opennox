package maps_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/opennox/libs/maps"
	crypt "github.com/opennox/noxcrypt"
)

// Encode the wire framing directly, independently of maps.Writer. Each
// encrypted length block contains a DWORD length followed by unused padding.
func sectionFile233(t *testing.T, magic, unused uint32, sections []maps.RawSection) []byte {
	t.Helper()
	var plain []byte
	u32 := func(v uint32) { plain = binary.LittleEndian.AppendUint32(plain, v) }
	align := func() {
		for len(plain)%8 != 0 {
			plain = append(plain, 0)
		}
	}
	u32(magic)
	if magic == maps.Magic {
		align()
		u32(0x12345678)
		u32(0)
	}
	u32(7)
	u32(11)
	for _, section := range sections {
		plain = append(plain, byte(len(section.Name)+1))
		plain = append(plain, section.Name...)
		plain = append(plain, 0)
		align()
		u32(uint32(len(section.Data)))
		u32(unused)
		plain = append(plain, section.Data...)
	}
	align()
	cipher, err := crypt.NewCipher(crypt.MapKey)
	if err != nil {
		t.Fatal(err)
	}
	encoded := make([]byte, len(plain))
	for i := 0; i < len(plain); i += 8 {
		cipher.Encrypt(encoded[i:i+8], plain[i:i+8])
	}
	return encoded
}

func TestMapSectionLengthDWORD233(t *testing.T) {
	for _, magic := range []uint32{maps.MagicOld, maps.Magic} {
		for _, unused := range []uint32{0, 1, 0x7fffffff, 0x80000000, 0xffffffff} {
			for n := 0; n < 32; n++ {
				t.Run(fmt.Sprintf("%08x/%08x/%02d", magic, unused, n), func(t *testing.T) {
					sections := []maps.RawSection{
						{Name: "First", Data: make([]byte, n)},
						{Name: "Next233", Data: make([]byte, 31-n)},
						{Name: "LastSection233", Data: []byte{0x5a, 0xa5}},
					}
					for i := range sections {
						for j := range sections[i].Data {
							sections[i].Data[j] = byte(i*53 + j*37 + 11)
						}
					}
					input := sectionFile233(t, magic, unused, sections)
					before := append([]byte(nil), input...)
					r, err := maps.NewReader(bytes.NewReader(input))
					if err != nil {
						t.Fatal(err)
					}
					got, err := r.ReadSectionsRaw()
					if err != nil {
						t.Fatal(err)
					}
					if len(got) != len(sections) {
						t.Fatalf("sections: got %d, want %d", len(got), len(sections))
					}
					for i, want := range sections {
						if got[i].Name != want.Name || !bytes.Equal(got[i].Data, want.Data) {
							t.Fatalf("section %d: got %q/%x, want %q/%x", i, got[i].Name, got[i].Data, want.Name, want.Data)
						}
					}
					if !bytes.Equal(input, before) {
						t.Fatal("reader modified original encrypted input")
					}
				})
			}
		}
	}
}
