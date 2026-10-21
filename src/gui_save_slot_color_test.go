package opennox

import (
	"testing"

	"github.com/opennox/opennox/v1/server"
)

func TestSaveSlotColorsDistinguishAutosave(t *testing.T) {
	// Original 0046CD70 returns palette 13 for AUTOSAVE, 14 for manual
	// saves, and 10 when flag 8 is present, before inspecting the path.
	for _, path := range []string{
		`C:\Nox\Save\AUTOSAVE\Player.plr`,
		`/Nox/Save/AUTOSAVE/Player.plr`,
		`AUTOSAVE`,
		`C:\Nox\Save\SAVE0001\Player.plr`,
		`/Nox/Save/SAVE0013/Player.plr`,
	} {
		for _, flags := range []uint32{0, 8} {
			var save server.SaveGameInfo
			copy(save.PathBuf[:], path)
			save.Flags = flags
			want := 14
			if path == `AUTOSAVE` || path == `C:\Nox\Save\AUTOSAVE\Player.plr` || path == `/Nox/Save/AUTOSAVE/Player.plr` {
				want = 13
			}
			if flags&8 != 0 {
				want = 10
			}
			if got := sub46CD70(&save); got != want {
				t.Errorf("path=%q flags=%d: color index=%d, want %d", path, flags, got, want)
			}
		}
	}
}
