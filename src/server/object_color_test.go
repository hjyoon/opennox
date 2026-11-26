package server

import (
	"image/color"
	"testing"
	"unsafe"

	noxcolor "github.com/opennox/libs/color"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestPlayerObjectSetColor(t *testing.T) {
	readMaterials := func(p *Player) [6]uint32 {
		c := p.Colors
		return [6]uint32{c.Hair, c.Goatee, c.UnkColor, c.Beard, c.Mustache, c.Skin}
	}
	const packetSize = 129
	for _, tc := range []struct {
		name   string
		index  int
		offset int
	}{
		{"hair", 0, 68},
		{"goatee", 1, 77},
		{"third material", 2, -1},
		{"beard", 3, 80},
		{"mustache", 4, 74},
		{"skin", 5, 71},
		{"negative index", -1, -1},
		{"past last material", 6, -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			player := &Player{NetCodeVal: 123, Lessons: 9}
			player.Info().Colors = PlayerColors{
				Hair: types.RGB{R: 40}, Skin: types.RGB{R: 64},
				Mustache: types.RGB{G: 72}, Goatee: types.RGB{G: 80}, Beard: types.RGB{B: 88},
				Pants: 2, Shirt1: 3, Shirt2: 4, Shoes1: 5, Shoes2: 6,
			}
			player.InitColors()
			update := &PlayerUpdateData{Player: player}
			obj := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}

			wantColors := readMaterials(player)
			var wantPacket [packetSize]byte
			player.EncodeNewPlayerPacket(wantPacket[:])
			input := color.NRGBA{R: 217, G: 139, B: 73, A: 255}
			obj.SetColor(tc.index, input)
			if tc.index >= 0 && tc.index < len(wantColors) {
				wantColors[tc.index] = noxcolor.RGB5551Color(input.R, input.G, input.B).Color32()
			}
			if got := readMaterials(player); got != wantColors {
				t.Fatalf("render materials = %08x, want %08x", got, wantColors)
			}
			if tc.offset >= 0 {
				copy(wantPacket[3+tc.offset:], []byte{input.R, input.G, input.B})
			}
			var packet [packetSize]byte
			player.EncodeNewPlayerPacket(packet[:])
			if packet != wantPacket {
				t.Fatalf("player join packet = % x, want % x", packet, wantPacket)
			}
			if tc.offset >= 0 {
				var reloaded Player
				if err := reloaded.Info().UnmarshalBinary(packet[3 : 3+PlayerInfoWireSize]); err != nil {
					t.Fatal(err)
				}
				reloaded.InitColors()
				if got := readMaterials(&reloaded); got != wantColors {
					t.Fatalf("reloaded materials = %08x, want %08x", got, wantColors)
				}
			}
		})
	}
}

func TestPlayerObjectSetColorBeforeJoin(t *testing.T) {
	for _, update := range []unsafe.Pointer{nil, unsafe.Pointer(&PlayerUpdateData{})} {
		obj := &Object{ObjClass: object.ClassPlayer, UpdateData: update}
		obj.SetColor(0, color.White)
	}
}

func TestAsColor3Handles16BitColor(t *testing.T) {
	for _, tc := range []struct {
		input color.Color
		want  Color3
	}{
		{color.Gray16{Y: 0xc040}, Color3{R: 0xc0, G: 0xc0, B: 0xc0}},
		{color.NRGBA64{R: 0xabcd, G: 0x4500, B: 0x2200, A: 0xffff}, Color3{R: 0xab, G: 0x45, B: 0x22}},
		{color.NRGBA64{R: 0xab00, G: 0x4500, B: 0x2200, A: 0x7fff}, Color3{}},
	} {
		if got := AsColor3(tc.input); got != tc.want {
			t.Errorf("AsColor3(%#v) = %#v, want %#v", tc.input, got, tc.want)
		}
	}
}
