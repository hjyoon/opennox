package legacy

import "github.com/opennox/opennox/v1/server"

// Nox_xxx_cancelAllSpells_4FEE90 is retained as a source-compatible Go name.
// Production callers use the restored native server method directly.
func Nox_xxx_cancelAllSpells_4FEE90(a1 *server.Object) {
	GetServer().S().Spells.Dur.SpellDurationCancelSelected4FEE90(a1)
}
