package opennox

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/opennox/v1/server"
)

func TestBerserkerConfusionWrapsFacing(t *testing.T) {
	s := &Server{Server: new(server.Server)}
	s.abilities.s = s
	s.NetSendPacketXxx = func(int, []byte, *server.Object, int, int) int { return 1 }
	previous := noxServer
	noxServer = s
	t.Cleanup(func() { noxServer = previous })

	for _, tc := range []struct {
		name      string
		seed      int
		direction server.Dir16
		want      server.Dir16
	}{
		{"overflow", 19, 249, 9},  // A turn of +16 crosses 255.
		{"underflow", 30, 3, 243}, // A turn of -16 crosses zero.
	} {
		t.Run(tc.name, func(t *testing.T) {
			s.Rand.Logic = prand.New(tc.seed)
			player := &server.Player{WeaponEquip: 1}
			update := &server.PlayerUpdateData{Player: player}
			unit := &server.Object{
				ObjClass:   object.ClassPlayer,
				Direction1: tc.direction,
				Direction2: tc.direction,
				UpdateData: unsafe.Pointer(update),
				Buffs:      1 << server.ENCHANT_CONFUSED,
			}
			player.PlayerUnit = unit
			nox_xxx_warriorBerserker_53FEB0(unit)
			if unit.Direction1 != tc.direction || unit.Direction2 != tc.want || update.State != server.PlayerState1 {
				t.Fatalf("Berserker facing = %d -> %d (state %d), want %d -> %d", unit.Direction1, unit.Direction2, update.State, tc.direction, tc.want)
			}
		})
	}
}
