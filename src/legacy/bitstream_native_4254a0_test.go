package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

//go:embed testdata/bitstream_original_4254a0_425550.json.gz
var bitstreamOriginalFixture4254A0 []byte

type bitstreamOriginalCase4254A0 struct {
	Kind              string          `json:"kind"`
	Index             uint8           `json:"index"`
	Value             byte            `json:"value"`
	Before            string          `json:"before"`
	After             string          `json:"after"`
	CursorBeforeDelta int             `json:"cursor_before_delta"`
	CursorDelta       int             `json:"cursor_delta"`
	CursorIndex       uint8           `json:"cursor_index"`
	Bits              int32           `json:"bits"`
	SourceOffset      int             `json:"source_offset"`
	TargetOffset      int             `json:"target_offset"`
	Result            json.RawMessage `json:"result"`
}

func TestBitstreamNativeCursorLayoutCABI(t *testing.T) {
	size, bitOffset := bitstreamNativeCursorLayoutC()
	if size != unsafe.Sizeof(bitCursorNative4254A0{}) ||
		bitOffset != unsafe.Offsetof(bitCursorNative4254A0{}.bit) ||
		bitOffset != unsafe.Sizeof(uintptr(0)) {
		t.Fatalf("C cursor layout=%d/%d Go=%d/%d", size, bitOffset,
			unsafe.Sizeof(bitCursorNative4254A0{}), unsafe.Offsetof(bitCursorNative4254A0{}.bit))
	}
}

func TestBitstreamNativeOriginalCABI(t *testing.T) {
	if got := fmt.Sprintf("%x", sha256.Sum256(bitstreamOriginalFixture4254A0)); got != "a3c3f71917af1e966cbf191a2767520821e4c0a36bf66c56378e4c5cc1201d72" {
		t.Fatal("sealed original compressed fixture changed", got)
	}
	z, err := gzip.NewReader(bytes.NewReader(bitstreamOriginalFixture4254A0))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != "d267d8bfc89696d1d41b1b49a3feb5dd1170827278fce5a9d8e650dd27031b42" {
		t.Fatal("sealed original fixture changed", got)
	}
	var cases []bitstreamOriginalCase4254A0
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 9092 {
		t.Fatal("sealed case count", len(cases))
	}
	counts := make(map[string]int)
	for _, c := range cases {
		counts[c.Kind]++
	}
	for kind, want := range map[string]int{"read_init": 24, "read_next": 2048, "write_init": 192, "write_next": 6144, "encode": 684} {
		if counts[kind] != want {
			t.Fatal("sealed operation count", kind, counts[kind], want)
		}
	}
	t.Logf("unchanged original cases=%d, instruction bytes=291, declared services=0", len(cases))
	for _, placement := range []int{0, 8, 24} {
		t.Run(fmt.Sprintf("placement-%d", placement), func(t *testing.T) {
			cursorBytes, freeCursor := alloc.MakeHigh([]byte{}, 128)
			defer freeCursor()
			dataBytes, freeData := alloc.MakeHigh([]byte{}, 640)
			defer freeData()
			for _, b := range [][]byte{cursorBytes, dataBytes} {
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&b[0])) <= math.MaxUint32 {
					t.Fatal("foreign native owners must be above the PE32 range")
				}
			}
			start := 32 + placement
			cursorPointer := unsafe.Pointer(&cursorBytes[start])
			cursor := (*bitCursorNative4254A0)(cursorPointer)
			for i, c := range cases {
				t.Run(fmt.Sprintf("%04d-%s", i, c.Kind), func(t *testing.T) {
					for i := range cursorBytes {
						cursorBytes[i] = 0xa5
					}
					for i := range dataBytes {
						dataBytes[i] = 0xa5
					}
					before, err := hex.DecodeString(c.Before)
					if err != nil {
						t.Fatal(err)
					}
					after, err := hex.DecodeString(c.After)
					if err != nil || len(before) != len(after) {
						t.Fatal("invalid sealed owner size", err)
					}
					if (c.Kind == "encode" && len(before) != 512) ||
						(c.Kind != "encode" && len(before) != 32) {
						t.Fatal("invalid sealed owner", c.Kind, len(before))
					}
					copy(dataBytes[start:], before)
					expectedData := append([]byte(nil), dataBytes...)
					copy(expectedData[start:], after)
					if c.Kind == "encode" {
						var want int32
						if err := json.Unmarshal(c.Result, &want); err != nil {
							t.Fatal(err)
						}
						source := unsafe.Pointer(&dataBytes[start+c.SourceOffset])
						target := unsafe.Pointer(&dataBytes[start+c.TargetOffset])
						if got := bitstreamNativeEncodeC(source, target, c.Bits); got != want {
							t.Fatalf("C encoder bits=%d result=%d want=%d", c.Bits, got, want)
						}
						if !bytes.Equal(cursorBytes, bytes.Repeat([]byte{0xa5}, len(cursorBytes))) {
							t.Fatal("encoder changed the independent cursor owner")
						}
					} else {
						dataPointer := unsafe.Pointer(&dataBytes[start+8])
						cursor.data = &dataBytes[start+8+c.CursorBeforeDelta]
						cursor.bit = c.Index
						got := bitstreamNativeCursorCallC(c.Kind, cursorPointer, dataPointer, c.Value)
						if c.Kind == "write_init" {
							if string(c.Result) != `"data"` || got != uintptr(dataPointer) {
								t.Fatalf("C writer lost the full data pointer: got=%#x want=%p", got, dataPointer)
							}
						} else {
							var want uintptr
							if err := json.Unmarshal(c.Result, &want); err != nil {
								t.Fatal(err)
							}
							if got != want {
								t.Fatalf("C %s index=%d value=%02x result=%#x want=%#x", c.Kind, c.Index, c.Value, got, want)
							}
						}
						wantData := uintptr(dataPointer) + uintptr(c.CursorDelta)
						if uintptr(unsafe.Pointer(cursor.data)) != wantData || cursor.bit != c.CursorIndex {
							t.Fatalf("C cursor data=%p bit=%d want=%#x/%d", cursor.data, cursor.bit, wantData, c.CursorIndex)
						}
						for i, b := range cursorBytes {
							if (start <= i && i < start+int(unsafe.Sizeof(uintptr(0)))) ||
								i == start+int(unsafe.Offsetof(cursor.bit)) {
								continue
							}
							if b != 0xa5 {
								t.Fatalf("C cursor changed padding or canary at %d: %02x", i, b)
							}
						}
					}
					if !bytes.Equal(dataBytes, expectedData) {
						t.Fatalf("C %s changed data or canaries outside the original write set", c.Kind)
					}
				})
			}
		})
	}
}
