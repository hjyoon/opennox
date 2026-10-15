package maps_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/opennox/libs/binenc"
	"github.com/opennox/libs/maps"
	"github.com/opennox/libs/xfer"
)

const rawRecordType232 = xfer.Type("TestRawMapRecord232")

type rawRecord232 struct{ data []byte }

func (*rawRecord232) XferType() xfer.Type { return rawRecordType232 }
func (x *rawRecord232) DecodeXfer(_ xfer.ObjectRegistry, r *binenc.Reader) error {
	x.data = r.ReadAllBytes()
	return nil
}

type partialRecord232 struct{}

func (*partialRecord232) XferType() xfer.Type { return "TestPartialMapRecord232" }
func (*partialRecord232) DecodeXfer(_ xfer.ObjectRegistry, r *binenc.Reader) error {
	_, _ = r.ReadU8()
	return nil
}

func init() {
	xfer.Register(&rawRecord232{})
	xfer.Register(&partialRecord232{})
}

// Build the original wire layout independently of Reader.Align: a WORD
// section version, a WORD type, optional padding to the next eight-byte
// boundary, a DWORD length with four unused bytes, then the XFER payload.
func recordSection232(padding uint32, payloads ...[]byte) maps.Objects {
	data := binary.LittleEndian.AppendUint16(nil, 1)
	for _, payload := range payloads {
		data = binary.LittleEndian.AppendUint16(data, 1)
		for len(data)%8 != 0 {
			data = append(data, 0xa5)
		}
		data = binary.LittleEndian.AppendUint32(data, uint32(len(payload)))
		data = binary.LittleEndian.AppendUint32(data, padding)
		data = append(data, payload...)
	}
	data = binary.LittleEndian.AppendUint16(data, 0)
	return maps.Objects{Vers: 1, Data: data[2:]}
}

func readRecords232(t *testing.T, objects *maps.Objects, toc *maps.ObjectsTOC, typ xfer.Type) (out []maps.Xfer, err error) {
	t.Helper()
	defer func() {
		if value := recover(); value != nil {
			t.Fatalf("map record decoder panicked: %v", value)
		}
	}()
	return objects.ReadObjects(toc, &xfer.StaticRegistry{ByType: map[string]xfer.Type{"record": typ}})
}

func TestReadObjectsAlignedDWORDRecords232(t *testing.T) {
	toc := &maps.ObjectsTOC{Vers: 1, TOC: []maps.ObjectTOC{{Ind: 1, Type: "record"}}}
	for _, padding := range []uint32{0, 1, 0x7fffffff, 0xffffffff} {
		for size := 0; size < 32; size++ {
			t.Run(fmt.Sprintf("unused_%08x/payload_%d", padding, size), func(t *testing.T) {
				payloads := [][]byte{bytes.Repeat([]byte{0x31}, size), {1, 3, 5, 7, 9, 11}, {0x91, 0x92}}
				objects := recordSection232(padding, payloads...)
				input := append([]byte(nil), objects.Data...)
				got, err := readRecords232(t, &objects, toc, rawRecordType232)
				if err != nil || len(got) != len(payloads) {
					t.Fatalf("records=%d error=%v want=%d complete records", len(got), err, len(payloads))
				}
				for i, record := range got {
					if record.Type != "record" || !bytes.Equal(record.Xfer.(*rawRecord232).data, payloads[i]) {
						t.Fatalf("record %d type/payload changed", i)
					}
				}
				if !bytes.Equal(objects.Data, input) {
					t.Fatal("input bytes changed")
				}
			})
		}
	}
}

func TestReadObjectsLengthBounds232(t *testing.T) {
	toc := &maps.ObjectsTOC{TOC: []maps.ObjectTOC{{Ind: 1, Type: "record"}}}
	for cut := 0; cut < 8; cut++ {
		t.Run(fmt.Sprintf("truncated_header_%d", cut), func(t *testing.T) {
			objects := recordSection232(0, []byte{1, 2})
			objects.Data = objects.Data[:6+cut]
			got, err := readRecords232(t, &objects, toc, rawRecordType232)
			if len(got) != 0 || !errors.Is(err, io.ErrUnexpectedEOF) {
				t.Fatalf("records=%d error=%v want unexpected EOF", len(got), err)
			}
		})
	}
	for _, size := range []uint32{1, 0x80000000, 0xffffffff} {
		t.Run(fmt.Sprintf("size_%08x", size), func(t *testing.T) {
			objects := recordSection232(0, nil)
			binary.LittleEndian.PutUint32(objects.Data[6:10], size)
			objects.Data = objects.Data[:14]
			got, err := readRecords232(t, &objects, toc, rawRecordType232)
			if len(got) != 0 || !errors.Is(err, io.ErrUnexpectedEOF) {
				t.Fatalf("records=%d error=%v want unexpected EOF", len(got), err)
			}
		})
	}
}

func TestReadObjectsTOCAndPartialDecode232(t *testing.T) {
	for _, test := range []struct {
		name string
		toc  []maps.ObjectTOC
		want string
	}{
		{"zero", []maps.ObjectTOC{{Ind: 0, Type: "record"}}, "must not be zero"},
		{"duplicate", []maps.ObjectTOC{{Ind: 1, Type: "record"}, {Ind: 1, Type: "other"}}, "ambiguous"},
		{"missing", nil, "entry is missing"},
	} {
		t.Run(test.name, func(t *testing.T) {
			objects := recordSection232(0, []byte{1, 2})
			_, err := readRecords232(t, &objects, &maps.ObjectsTOC{TOC: test.toc}, rawRecordType232)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error=%v want %q", err, test.want)
			}
		})
	}
	t.Run("partial", func(t *testing.T) {
		objects := recordSection232(0, []byte{1, 2})
		got, err := readRecords232(t, &objects, &maps.ObjectsTOC{TOC: []maps.ObjectTOC{{Ind: 1, Type: "record"}}}, "TestPartialMapRecord232")
		if len(got) != 1 || err == nil || !strings.Contains(err.Error(), "partial decoding") {
			t.Fatalf("records=%d error=%v want partial decoding", len(got), err)
		}
	})
}
