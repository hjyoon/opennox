package server

import (
	"bytes"
	"testing"

	"github.com/opennox/libs/player"
	"github.com/opennox/libs/types"
)

func playerInfoWireFixture(t *testing.T) PlayerInfo {
	t.Helper()
	var info PlayerInfo
	info.SetName("Ada")
	info.SetField2235(0x12345678)
	info.SetField2239(0x90abcdef)
	info.SetField2243(0x10203040)
	info.SetField2247(0xfedcba98)
	info.SetPlayerClass(player.Conjurer)
	info.SetIsFemale(1)
	info.Colors = PlayerColors{
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
	info.SetNameSuff(" IV")
	return info
}

func TestPlayerInfoWireLayoutAndRoundTrip(t *testing.T) {
	info := playerInfoWireFixture(t)
	data := bytes.Repeat([]byte{0xcc}, PlayerInfoWireSize+4)
	if err := info.MarshalBinaryTo(data); err != nil {
		t.Fatal(err)
	}
	if got, want := data[:8], []byte{'A', 0, 'd', 0, 'a', 0, 0, 0}; !bytes.Equal(got, want) {
		t.Fatalf("name prefix = % x, want % x", got, want)
	}
	checks := []struct {
		off  int
		want []byte
	}{
		{50, []byte{0x78, 0x56, 0x34, 0x12}},
		{54, []byte{0xef, 0xcd, 0xab, 0x90}},
		{58, []byte{0x40, 0x30, 0x20, 0x10}},
		{62, []byte{0x98, 0xba, 0xdc, 0xfe}},
		{66, []byte{byte(player.Conjurer), 1}},
		{68, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21}},
		{89, []byte{' ', 0, 'I', 0, 'V', 0, 0, 0}},
	}
	for _, check := range checks {
		if got := data[check.off : check.off+len(check.want)]; !bytes.Equal(got, check.want) {
			t.Errorf("wire[%d:] = % x, want % x", check.off, got, check.want)
		}
	}
	if got := data[PlayerInfoWireSize:]; !bytes.Equal(got, bytes.Repeat([]byte{0xcc}, 4)) {
		t.Fatalf("trailing bytes changed: % x", got)
	}

	var decoded PlayerInfo
	if err := decoded.UnmarshalBinary(data); err != nil {
		t.Fatal(err)
	}
	roundTrip, err := decoded.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(roundTrip, data[:PlayerInfoWireSize]) {
		t.Fatalf("round trip = % x, want % x", roundTrip, data[:PlayerInfoWireSize])
	}
	data[0], data[50], data[68], data[89] = 0, 0, 0, 0
	again, err := decoded.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(roundTrip, again) {
		t.Fatal("decoded player info aliases the source wire buffer")
	}
}

func TestPlayerInfoWireRejectsShortBuffersWithoutMutation(t *testing.T) {
	info := playerInfoWireFixture(t)
	for size := 0; size < PlayerInfoWireSize; size++ {
		dst := bytes.Repeat([]byte{0xa5}, size)
		beforeDst := append([]byte(nil), dst...)
		if err := info.MarshalBinaryTo(dst); err == nil {
			t.Fatalf("%d-byte marshal buffer was accepted", size)
		}
		if !bytes.Equal(dst, beforeDst) {
			t.Fatalf("%d-byte marshal buffer changed", size)
		}

		decoded := info
		if err := decoded.UnmarshalBinary(make([]byte, size)); err == nil {
			t.Fatalf("%d-byte player info was accepted", size)
		}
		got, _ := decoded.MarshalBinary()
		want, _ := info.MarshalBinary()
		if !bytes.Equal(got, want) {
			t.Fatalf("%d-byte decode mutated player info", size)
		}
	}
	var nilInfo *PlayerInfo
	if _, err := nilInfo.MarshalBinary(); err == nil {
		t.Fatal("nil player info marshal was accepted")
	}
	if err := nilInfo.UnmarshalBinary(make([]byte, PlayerInfoWireSize)); err == nil {
		t.Fatal("nil player info unmarshal was accepted")
	}
}
