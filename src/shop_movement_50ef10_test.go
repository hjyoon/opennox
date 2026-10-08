package opennox

import (
	"fmt"
	"testing"

	"github.com/opennox/opennox/v1/common/flags"
)

func TestShopMovementFrozen50EF10OriginalGameModes(t *testing.T) {
	for _, mode := range []noxflags.GameFlag{0, noxflags.GameModeArena, noxflags.GameModeQuest, noxflags.GameModeCoop} {
		for _, online := range []noxflags.GameFlag{0, noxflags.GameOnline} {
			flags := noxflags.GameHost | noxflags.GameClient | mode | online
			t.Run(fmt.Sprintf("%#x", uint32(flags)), func(t *testing.T) {
				if got, want := shopMovementFrozen50EF10(flags), uint32(flags)&0x800 != 0; got != want {
					t.Fatalf("movement freeze = %t, want original 0x800 gate %t", got, want)
				}
			})
		}
	}
}
