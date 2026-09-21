package client

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/opennox/libs/noxnet"

	"github.com/opennox/opennox/v1/server"
)

func TestClassStatsFromNetWireOrder(t *testing.T) {
	want := server.ClassStats{Health: 1.25, Mana: -2.5, Speed: 3.75, Strength: -4.5}
	wire := make([]byte, 16)
	binary.LittleEndian.PutUint32(wire[0:], math.Float32bits(want.Health))
	binary.LittleEndian.PutUint32(wire[4:], math.Float32bits(want.Mana))
	binary.LittleEndian.PutUint32(wire[8:], math.Float32bits(want.Strength))
	binary.LittleEndian.PutUint32(wire[12:], math.Float32bits(want.Speed))

	var msg noxnet.MsgStatMult
	if n, err := msg.Decode(wire); err != nil || n != len(wire) {
		t.Fatalf("decode = %d, %v", n, err)
	}
	if got := classStatsFromNet(msg); got != want {
		t.Fatalf("decoded stats = %#v, want %#v", got, want)
	}
}
