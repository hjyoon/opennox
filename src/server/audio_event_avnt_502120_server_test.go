package server

import (
	"encoding/binary"
	"io"
	"math"
	"testing"

	"github.com/opennox/opennox/v1/common/sound"
)

type audioEventAVNTFile502120 struct {
	data   []byte
	offset int
}

func (f *audioEventAVNTFile502120) take(size int) []byte {
	if size < 0 || f.offset+size > len(f.data) {
		panic("audioEventAVNTFile502120 out of bounds")
	}
	value := f.data[f.offset : f.offset+size]
	f.offset += size
	return value
}

func (f *audioEventAVNTFile502120) Skip(size int) { _ = f.take(size) }
func (f *audioEventAVNTFile502120) ReadU8() uint8 { return f.take(1)[0] }
func (f *audioEventAVNTFile502120) ReadI8() int8  { return int8(f.ReadU8()) }
func (f *audioEventAVNTFile502120) ReadU16() uint16 {
	return binary.LittleEndian.Uint16(f.take(2))
}
func (f *audioEventAVNTFile502120) ReadI16() int16 { return int16(f.ReadU16()) }
func (f *audioEventAVNTFile502120) ReadU32() uint32 {
	return binary.LittleEndian.Uint32(f.take(4))
}
func (f *audioEventAVNTFile502120) ReadI32() int32 { return int32(f.ReadU32()) }
func (f *audioEventAVNTFile502120) ReadString8() (string, error) {
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

func audioEventAVNTWire502120(name string, commands ...byte) []byte {
	if len(name) > math.MaxUint8 {
		panic("AVNT test name is too long")
	}
	wire := make([]byte, 1, 1+len(name)+len(commands))
	wire[0] = byte(len(name))
	wire = append(wire, name...)
	return append(wire, commands...)
}

func TestServerAudioReadAVNT502120SignedFieldsAndConsumption(t *testing.T) {
	const soundID = sound.SoundAnchorCast
	name := soundID.String()
	f := &audioEventAVNTFile502120{data: audioEventAVNTWire502120(name,
		2, 0xa2,
		3, 0xa3,
		4, 0xa4,
		7, 2, 'a', 'b', 1, 'c', 0,
		9, 0xff, 0xff,
		10, 0x00, 0x80,
		0,
	)}
	audio := &serverAudio{inited: true}

	if !audio.ReadAVNT(f) {
		t.Fatal("ReadAVNT = false")
	}
	if f.offset != len(f.data) {
		t.Fatalf("offset = %d, want %d", f.offset, len(f.data))
	}
	descriptor := &audio.bySound[soundID]
	if descriptor.MaxDist != -15 || descriptor.Flags != 0xffff8000 ||
		descriptor.Field8 != 0xa3 || descriptor.Field12 != 2 ||
		descriptor.Field16 != 0xa2 || descriptor.Field20 != 0xa4 {
		t.Fatalf("descriptor = %+v", descriptor)
	}
}

func TestServerAudioReadAVNT502120ZeroSoundCountsAndGates(t *testing.T) {
	f := &audioEventAVNTFile502120{data: audioEventAVNTWire502120("not-a-real-sound",
		2, 1,
		3, 2,
		4, 3,
		7, 1, 'x', 0,
		9, 0xff, 0xff,
		10, 0xfe, 0xff,
		0,
	)}
	audio := &serverAudio{inited: true}
	descriptor := &audio.bySound[sound.SoundNone]
	descriptor.MaxDist = 10
	descriptor.Flags = 11
	descriptor.Field8 = 12
	descriptor.Field12 = math.MaxUint32
	descriptor.Field16 = 14
	descriptor.Field20 = 15

	if !audio.ReadAVNT(f) {
		t.Fatal("ReadAVNT = false")
	}
	if descriptor.MaxDist != 10 || descriptor.Flags != 11 ||
		descriptor.Field8 != 12 || descriptor.Field12 != 0 ||
		descriptor.Field16 != 14 || descriptor.Field20 != 15 {
		t.Fatalf("sound-zero descriptor = %+v", descriptor)
	}
}

func TestServerAudioReadAVNT502120UninitializedStillCountsOpcode7(t *testing.T) {
	const soundID = sound.SoundAnchorCast
	f := &audioEventAVNTFile502120{data: audioEventAVNTWire502120(soundID.String(),
		2, 1,
		7, 1, 'x', 0,
		9, 1, 0,
		0,
	)}
	audio := &serverAudio{}
	descriptor := &audio.bySound[soundID]
	descriptor.MaxDist = 40
	descriptor.Field12 = 3
	descriptor.Field16 = 5

	if !audio.ReadAVNT(f) {
		t.Fatal("ReadAVNT = false")
	}
	if descriptor.MaxDist != 40 || descriptor.Field12 != 4 || descriptor.Field16 != 5 {
		t.Fatalf("uninitialized descriptor = %+v", descriptor)
	}
}
