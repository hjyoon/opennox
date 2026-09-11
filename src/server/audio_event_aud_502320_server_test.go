package server

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"
	"testing"

	"github.com/opennox/opennox/v1/common/sound"
)

type audioEventAUDFile502320 struct {
	data   []byte
	offset int
}

func (f *audioEventAUDFile502320) take(size int) []byte {
	if size < 0 || f.offset+size > len(f.data) {
		panic("audioEventAUDFile502320 out of bounds")
	}
	value := f.data[f.offset : f.offset+size]
	f.offset += size
	return value
}

func (f *audioEventAUDFile502320) Skip(size int) { _ = f.take(size) }
func (f *audioEventAUDFile502320) ReadU8() uint8 { return f.take(1)[0] }
func (f *audioEventAUDFile502320) ReadI8() int8  { return int8(f.ReadU8()) }
func (f *audioEventAUDFile502320) ReadU16() uint16 {
	return binary.LittleEndian.Uint16(f.take(2))
}
func (f *audioEventAUDFile502320) ReadI16() int16 { return int16(f.ReadU16()) }
func (f *audioEventAUDFile502320) ReadU32() uint32 {
	return binary.LittleEndian.Uint32(f.take(4))
}
func (f *audioEventAUDFile502320) ReadI32() int32 { return int32(f.ReadU32()) }
func (f *audioEventAUDFile502320) ReadString8() (string, error) {
	if f.offset >= len(f.data) {
		return "", io.EOF
	}
	size := int(f.ReadU8())
	if f.offset+size > len(f.data) {
		f.offset = len(f.data)
		return "", io.ErrUnexpectedEOF
	}
	return string(f.take(size)), nil
}

func audioEventAUDRecordWire502370(
	name string,
	flags int16,
	field8 uint8,
	maxDistance int16,
	field20 uint8,
	samples ...string,
) []byte {
	if len(name) > math.MaxInt8 {
		panic("AUD test sound name is too long")
	}
	wire := make([]byte, 1, 1+len(name)+9)
	wire[0] = byte(len(name))
	wire = append(wire, name...)
	wire = binary.LittleEndian.AppendUint16(wire, uint16(flags))
	wire = append(wire, field8)
	wire = binary.LittleEndian.AppendUint16(wire, uint16(maxDistance))
	wire = append(wire, field20, 0xa1, 0xa2, 0xa3)
	for _, sample := range samples {
		if len(sample) == 0 || len(sample) > math.MaxInt8 {
			panic("AUD test sample length must be in 1..127")
		}
		wire = append(wire, byte(len(sample)))
		wire = append(wire, sample...)
	}
	return append(wire, 0)
}

func audioEventAUDListWire502320(count int32, records ...[]byte) []byte {
	wire := binary.LittleEndian.AppendUint32(nil, uint32(count))
	for _, record := range records {
		wire = append(wire, record...)
	}
	return wire
}

func TestServerAudioReadAUD502320SignedNonpositiveCounts(t *testing.T) {
	for _, count := range []int32{math.MinInt32, -1, 0} {
		t.Run(fmt.Sprintf("count-%d", count), func(t *testing.T) {
			f := &audioEventAUDFile502320{
				data: append(audioEventAUDListWire502320(count), 0xff),
			}
			audio := &serverAudio{inited: true}
			if !audio.ReadAUD(f) {
				t.Fatalf("ReadAUD(%d) = false", count)
			}
			if f.offset != 4 {
				t.Fatalf("offset = %d, want signed count only", f.offset)
			}
		})
	}
}

func TestServerAudioReadAUD502320PositiveCompletionAndSignedFields(t *testing.T) {
	const soundID = sound.SoundAnchorCast
	record := audioEventAUDRecordWire502370(soundID.String(), -1, 0xfe, 2, 0xfd, "ab", "c")
	f := &audioEventAUDFile502320{data: audioEventAUDListWire502320(1, record)}
	audio := &serverAudio{inited: true}
	descriptor := &audio.bySound[soundID]
	descriptor.MaxDist = 600
	descriptor.Field12 = math.MaxUint32
	descriptor.Field16 = 99

	if !audio.ReadAUD(f) {
		t.Fatal("ReadAUD = false, want positive-list completion success")
	}
	if f.offset != len(f.data) {
		t.Fatalf("offset = %d, want %d", f.offset, len(f.data))
	}
	if descriptor.MaxDist != 30 || descriptor.Flags != math.MaxUint32 ||
		descriptor.Field8 != 0xfe || descriptor.Field12 != 1 ||
		descriptor.Field16 != 2 || descriptor.Field20 != 0xfd {
		t.Fatalf("descriptor = %+v", descriptor)
	}
}

func TestServerAudioReadAUD502320StopsOnTruncatedRecord(t *testing.T) {
	const soundID = sound.SoundAnchorCast
	first := audioEventAUDRecordWire502370(soundID.String(), 7, 8, 0, 9)
	truncated := []byte{5, 'b', 'a'}
	f := &audioEventAUDFile502320{data: audioEventAUDListWire502320(2, first, truncated)}
	audio := &serverAudio{inited: true}

	if audio.ReadAUD(f) {
		t.Fatal("ReadAUD = true, want second-record name failure")
	}
	if f.offset != len(f.data) || audio.bySound[soundID].Flags != 7 {
		t.Fatalf("offset/first flags = %d/%d, want %d/7", f.offset, audio.bySound[soundID].Flags, len(f.data))
	}
}

func TestServerAudioReadAUD502320InactiveRecordConsumption(t *testing.T) {
	name := "not-a-real-sound"
	record := append([]byte{byte(len(name))}, name...)
	record = append(record,
		0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa7, 0xa8,
		2, 'x', 'y', 1, 'z', 0,
	)
	f := &audioEventAUDFile502320{data: audioEventAUDListWire502320(1, record)}
	audio := &serverAudio{inited: true}
	descriptor := &audio.bySound[sound.SoundNone]
	*descriptor = audioEvent2{MaxDist: 11, Flags: 12, Field8: 13, Field12: 14, Field16: 15, Field20: 16}

	if !audio.ReadAUD(f) {
		t.Fatal("ReadAUD = false")
	}
	if f.offset != len(f.data) {
		t.Fatalf("offset = %d, want %d", f.offset, len(f.data))
	}
	want := audioEvent2{MaxDist: 11, Flags: 12, Field8: 13, Field12: 14, Field16: 15, Field20: 16}
	if !reflect.DeepEqual(*descriptor, want) {
		t.Fatalf("sound-zero descriptor = %+v, want %+v", *descriptor, want)
	}
}

type audioEventAUDSignedFile502370 struct {
	name  string
	i16   []int16
	u8    []uint8
	i8    []int8
	skips []int
}

func (f *audioEventAUDSignedFile502370) Skip(size int) {
	f.skips = append(f.skips, size)
}
func (f *audioEventAUDSignedFile502370) ReadU8() uint8 {
	value := f.u8[0]
	f.u8 = f.u8[1:]
	return value
}
func (f *audioEventAUDSignedFile502370) ReadI8() int8 {
	value := f.i8[0]
	f.i8 = f.i8[1:]
	return value
}
func (f *audioEventAUDSignedFile502370) ReadU16() uint16 {
	panic("AUD record binding must use ReadI16")
}
func (f *audioEventAUDSignedFile502370) ReadI16() int16 {
	value := f.i16[0]
	f.i16 = f.i16[1:]
	return value
}
func (f *audioEventAUDSignedFile502370) ReadU32() uint32 {
	panic("unexpected ReadU32")
}
func (f *audioEventAUDSignedFile502370) ReadI32() int32 {
	panic("unexpected ReadI32")
}
func (f *audioEventAUDSignedFile502370) ReadString8() (string, error) {
	return f.name, nil
}

func TestServerAudioReadAUDRecord502370UsesSignedReadersAndSkips(t *testing.T) {
	const soundID = sound.SoundAnchorCast
	f := &audioEventAUDSignedFile502370{
		name: soundID.String(),
		i16:  []int16{-1, 0},
		u8:   []uint8{3, 4},
		i8:   []int8{-1, 0},
	}
	audio := &serverAudio{inited: true}
	descriptor := &audio.bySound[soundID]
	descriptor.MaxDist = 123
	descriptor.Field12 = 5

	if !audio.readAUDRec(f) {
		t.Fatal("readAUDRec = false")
	}
	if want := []int{3, -1}; !reflect.DeepEqual(f.skips, want) {
		t.Fatalf("skips = %v, want %v", f.skips, want)
	}
	if descriptor.MaxDist != 123 || descriptor.Flags != math.MaxUint32 ||
		descriptor.Field8 != 3 || descriptor.Field12 != 6 ||
		descriptor.Field16 != 2 || descriptor.Field20 != 4 {
		t.Fatalf("descriptor = %+v", descriptor)
	}
}
