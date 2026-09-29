package opennox

import (
	"bytes"
	"encoding/binary"
	"image"
	"reflect"
	"testing"

	"github.com/opennox/libs/player"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func playerInfoPacketFixture(t *testing.T) server.PlayerInfo {
	t.Helper()
	var info server.PlayerInfo
	info.SetName("Wire Player")
	info.SetField2235(0x11223344)
	info.SetField2239(0x55667788)
	info.SetField2243(0x99aabbcc)
	info.SetField2247(0xddeeff00)
	info.SetPlayerClass(player.Wizard)
	info.SetIsFemale(1)
	info.Colors = server.PlayerColors{
		Hair:     types.RGB{R: 1, G: 2, B: 3},
		Skin:     types.RGB{R: 4, G: 5, B: 6},
		Mustache: types.RGB{R: 7, G: 8, B: 9},
		Goatee:   types.RGB{R: 10, G: 11, B: 12},
		Beard:    types.RGB{R: 13, G: 14, B: 15},
		Pants:    16,
		Shirt1:   17,
		Shirt2:   18,
		Shoes1:   19,
		Shoes2:   20,
	}
	info.Field2273 = 21
	info.SetNameSuff(" 2")
	return info
}

func TestPlayerOptsUsesPackedPlayerInfoWire(t *testing.T) {
	want := PlayerOpts{
		Info:      playerInfoPacketFixture(t),
		Screen:    image.Pt(-12345, 67890),
		Serial:    "SERIAL-123",
		Field2096: "profile",
		Field2068: -987654,
		Field2072: "slot",
		Byte152:   0xa5,
	}
	data, err := want.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 153 {
		t.Fatalf("player opts size = %d, want 153", len(data))
	}
	infoWire, err := want.Info.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data[:server.PlayerInfoWireSize], infoWire) {
		t.Fatalf("player info prefix = % x, want % x", data[:server.PlayerInfoWireSize], infoWire)
	}
	var got PlayerOpts
	if err := got.UnmarshalBinary(data); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip = %#v, want %#v", got, want)
	}
	data[0] ^= 0xff
	gotWire, err := got.Info.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotWire, infoWire) {
		t.Fatal("decoded player options alias the source packet")
	}
	if err := got.UnmarshalBinary(data[:152]); err == nil {
		t.Fatal("short player options packet was accepted")
	}
}

func TestNewPlayerPacketBoundsPackedFields(t *testing.T) {
	field2140 := int32(-2345)
	pl := &server.Player{
		NetCodeVal:  0x9234,
		Lessons:     -1234,
		Field2140:   uint32(field2140),
		ArmorEquip:  0x11223344,
		WeaponEquip: 0x55667788,
		Field2152:   0x1aa,
		Field2156:   0x2bb,
		Field3676:   3,
		Field3680:   0xffffffff,
	}
	*pl.Info() = playerInfoPacketFixture(t)
	pl.SetField2096("ABCDEFGHIJK")
	var packet [132]byte
	for i := range packet {
		packet[i] = 0xcc
	}
	nox_xxx_netNewPlayerMakePacket_4DDA90(packet[:], pl)
	if packet[0] != 45 {
		t.Fatalf("opcode = %#x, want 45", packet[0])
	}
	if got := binary.LittleEndian.Uint16(packet[1:3]); got != 0x9234 {
		t.Fatalf("net code = %#x, want 0x9234", got)
	}
	infoWire, err := pl.Info().MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(packet[3:100], infoWire) {
		t.Fatalf("player info = % x, want % x", packet[3:100], infoWire)
	}
	if got := int16(binary.LittleEndian.Uint16(packet[100:102])); got != -1234 {
		t.Fatalf("lessons = %d, want -1234", got)
	}
	if got := int16(binary.LittleEndian.Uint16(packet[102:104])); got != -2345 {
		t.Fatalf("field 2140 = %d, want -2345", got)
	}
	if got := binary.LittleEndian.Uint32(packet[104:108]); got != 0x11223344 {
		t.Fatalf("armor = %#x, want 0x11223344", got)
	}
	if got := binary.LittleEndian.Uint32(packet[108:112]); got != 0x55667788 {
		t.Fatalf("weapon = %#x, want 0x55667788", got)
	}
	if got := binary.LittleEndian.Uint32(packet[112:116]); got != 0x423 {
		t.Fatalf("flags = %#x, want 0x423", got)
	}
	if packet[116] != 0xaa || packet[117] != 0xbb || packet[118] != 1 {
		t.Fatalf("status bytes = % x, want aa bb 01", packet[116:119])
	}
	if got, want := packet[119:129], []byte{'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 0}; !bytes.Equal(got, want) {
		t.Fatalf("profile field = % x, want % x", got, want)
	}
	if got := packet[129:]; !bytes.Equal(got, []byte{0xcc, 0xcc, 0xcc}) {
		t.Fatalf("bytes after packet changed: % x", got)
	}
}
