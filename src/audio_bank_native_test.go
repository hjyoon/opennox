package opennox

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

func TestAudioBankNativePublicCFileAndIndex(t *testing.T) {
	handles.InitHigh()
	t.Cleanup(handles.Release)
	dir := t.TempDir()
	bagPath := filepath.Join(dir, "bank.bag")
	if err := os.WriteFile(bagPath, []byte("abcdefuvwxyz"), 0600); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(bagPath)
	if err != nil {
		t.Fatal(err)
	}
	banks, freeBank := alloc.MakeHigh([]audioStructXxx{}, 1)
	defer freeBank()
	rows, freeRows := alloc.MakeHigh([]audioStructYyy{}, 2)
	defer freeRows()
	b := &banks[0]
	b.arr0, b.size4 = &rows[0], 2
	b.bagfile268 = legacy.NewFileHandle(binfile.NewFile(f))
	defer func() {
		if b.field272 != nil {
			legacy.Nox_fs_close(b.field272)
		}
		legacy.Nox_fs_close(b.bagfile268)
	}()
	for i, name := range []string{"Alpha", "Zulu"} {
		copy(rows[i].field0[:], name)
		record := (*[9]uint32)(unsafe.Pointer(&rows[i]))
		record[4], record[5], record[6], record[7], record[8] = uint32(i*6), 6, uint32(22050-i*11025), uint32(i*5), uint32(i*2)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(b)) <= math.MaxUint32 {
		t.Fatal("native bank must be foreign above 4 GiB")
	}
	if legacy.AudioNativeBankPointer(0, unsafe.Pointer(b), 1) != unsafe.Pointer(&rows[1]) || legacy.AudioNativeBankPointer(0, unsafe.Pointer(b), -1) != nil || legacy.AudioNativeBankPointer(0, unsafe.Pointer(b), 2) != nil {
		t.Fatal("native index pointer/limit")
	}
	name, freeName := alloc.MakeHigh([]byte{}, 32)
	defer freeName()
	for _, tc := range []struct {
		name string
		want int32
	}{{"alpha", 0}, {"ALPHA", 0}, {"Zulu", 1}, {"missing", -1}} {
		clear(name)
		copy(name, tc.name)
		if got := legacy.AudioNativeBankScalar(1, unsafe.Pointer(b), unsafe.Pointer(&name[0]), 0); got != tc.want {
			t.Fatalf("lookup %q: %d", tc.name, got)
		}
	}
	format, freeFormat := alloc.MakeHigh([]uint32{}, 7)
	defer freeFormat()
	for i := range format {
		format[i] = 0xa7a7a7a7
	}
	if got := legacy.AudioNativeBankScalar(2, unsafe.Pointer(b), unsafe.Pointer(&format[0]), 1); got != 2 || *(*[7]uint32)(unsafe.Pointer(&format[0])) != [7]uint32{4, 0, 11025, 2, 2, 0xa7a7a7a7, 2} {
		t.Fatal("native index format/unchanged word5")
	}
	data, freeData := alloc.MakeHigh([]byte{}, 64)
	defer freeData()
	read := func(n int32) int32 {
		return legacy.AudioNativeBankScalar(4, unsafe.Pointer(b), unsafe.Pointer(&data[16]), n)
	}
	for _, id := range []int32{0, 1} {
		if legacy.AudioNativeBankScalar(3, unsafe.Pointer(b), nil, id) != 1 || b.field280 != b.bagfile268 || b.field284 != 6 {
			t.Fatal("selected native bag handle/length")
		}
		for i := range data {
			data[i] = 0xa7
		}
		if read(-1) != 0 || read(0) != 0 || read(100) != 6 || b.field284 != 0 || read(1) != 0 {
			t.Fatal("bounded signed read/EOF")
		}
		want := []byte("abcdef")
		if id == 1 {
			want = []byte("uvwxyz")
		}
		if !bytes.Equal(data[16:22], want) || !bytes.Equal(data[:16], bytes.Repeat([]byte{0xa7}, 16)) || !bytes.Equal(data[22:], bytes.Repeat([]byte{0xa7}, 42)) {
			t.Fatal("read payload/canary")
		}
		if legacy.AudioNativeBankPointer(1, unsafe.Pointer(b), 0) != nil || b.field280 != nil || b.field272 != nil {
			t.Fatal("read close native handles")
		}
	}
	// A missing override retains the selected bag stream.
	b.field276 = 1
	copy(b.path2_8[:], filepath.ToSlash(dir)+"/")
	if legacy.AudioNativeBankScalar(3, unsafe.Pointer(b), nil, 0) != 1 || b.field280 != b.bagfile268 || read(6) != 6 || string(data[16:22]) != "abcdef" {
		t.Fatal("missing override fallback")
	}
	writeWAV := func(channels, align uint16, rate uint32, payload []byte) {
		var body bytes.Buffer
		body.WriteString("WAVEJUNK")
		binary.Write(&body, binary.LittleEndian, uint32(4))
		body.WriteString("skip")
		body.WriteString("fmt ")
		binary.Write(&body, binary.LittleEndian, uint32(18))
		for _, v := range []uint16{1, channels} {
			binary.Write(&body, binary.LittleEndian, v)
		}
		binary.Write(&body, binary.LittleEndian, rate)
		binary.Write(&body, binary.LittleEndian, rate*uint32(align))
		binary.Write(&body, binary.LittleEndian, align)
		binary.Write(&body, binary.LittleEndian, uint16(8*align/channels))
		binary.Write(&body, binary.LittleEndian, uint16(0))
		body.WriteString("data")
		binary.Write(&body, binary.LittleEndian, uint32(len(payload)))
		body.Write(payload)
		var wav bytes.Buffer
		wav.WriteString("RIFF")
		binary.Write(&wav, binary.LittleEndian, uint32(body.Len()))
		wav.Write(body.Bytes())
		if err := os.WriteFile(filepath.Join(dir, "Alpha.wav"), wav.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		channels, align uint16
		rate, flags     uint32
		payload         []byte
	}{{1, 1, 8000, 2, []byte{1, 2, 3, 4}}, {2, 4, 44100, 7, []byte{0, 1, 2, 3, 4, 5, 6, 7}}} {
		writeWAV(tc.channels, tc.align, tc.rate, tc.payload)
		if legacy.AudioNativeBankScalar(3, unsafe.Pointer(b), nil, 0) != 1 || b.field272 == nil || b.field280 != b.field272 || b.field284 != uint32(len(tc.payload)) {
			t.Fatal("selected native WAV override")
		}
		record := (*[9]uint32)(unsafe.Pointer(&rows[0]))
		if record[6] != tc.rate || record[7] != tc.flags || read(100) != int32(len(tc.payload)) || !bytes.Equal(data[16:16+len(tc.payload)], tc.payload) {
			t.Fatal("WAV metadata/payload")
		}
		legacy.AudioNativeBankPointer(1, unsafe.Pointer(b), 0)
		if b.field272 != nil || b.field280 != nil {
			t.Fatal("override close")
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "Alpha.wav"), []byte("invalid RIFF"), 0600); err != nil {
		t.Fatal(err)
	}
	if legacy.AudioNativeBankScalar(3, unsafe.Pointer(b), nil, 0) != 1 || b.field272 != nil || b.field280 != b.bagfile268 || read(6) != 6 || string(data[16:22]) != "abcdef" {
		t.Fatal("malformed RIFF fallback")
	}
	b.field284 = 0xffffffff
	if read(5) != 0 || b.field284 != 0xffffffff {
		t.Fatal("signed remaining-length semantics")
	}
	legacy.AudioNativeBankPointer(1, unsafe.Pointer(b), 0)
}
