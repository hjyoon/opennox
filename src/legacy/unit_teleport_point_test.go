package legacy

import (
	"testing"

	"github.com/opennox/libs/object"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

func TestTeleportToMBPoint4E7190GatesBeforePointRead(t *testing.T) {
	old := noxflags.GetGame()
	noxflags.UnsetGame(old)
	t.Cleanup(func() { noxflags.UnsetGame(noxflags.GetGame()); noxflags.SetGame(old) })
	for _, tc := range []struct {
		name string
		unit server.Object
		coop bool
	}{
		{"anchored", server.Object{Buffs: 1 << server.ENCHANT_ANCHORED}, false},
		{"no update", server.Object{ObjFlags: object.FlagNoUpdate}, false},
		{"nonunit outside coop", server.Object{ObjClass: object.ClassSimple}, false},
		{"immobile admitted by coop", server.Object{ObjClass: object.ClassImmobile}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			noxflags.UnsetGame(noxflags.GameModeCoop)
			if tc.coop {
				noxflags.SetGame(noxflags.GameModeCoop)
			}
			TeleportToMBPoint4E7190(&tc.unit, nil)
		})
	}
}
