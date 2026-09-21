package opennox

import "github.com/opennox/libs/strman"

const spellResultSourcePath4FB0B0 = `C:\NoxPost\src\Server\Magic\plyrspel.c`

var spellResultKeys4FB0B0 = [...]string{
	"SpellOK",
	"SpellInUse",
	"execspel.c:UnseenTarget",
	"TooManySpells",
	"summon.c:CreatureControlFailed",
	"glyph.c:TooManyGlyphs",
	"Spell.c:DuplicateGlyphSpell",
	"glyph.c:CantCastGlyph",
	"BadTarget",
	"BadSkill",
	"Illegal",
	"NotEnoughManaCast",
	"spell.c:NotEnoughManaGlyph",
	"SpellRestrictedByFlag",
	"SpellNotStartedWarCry",
	"spell.c:SpellCancelledByWarCry",
	"SpellRestrictedByBall",
	"SpellRestrictedByCrown",
}

type spellResultHooks4FB0B0 struct {
	loadString    func(key, source string) string
	printCentered func(text string)
}

// spellResult4FB0B0 preserves the defined input domain of GAME.EXE
// 004FB0B0. The original indexes an eighteen-entry packed PE32 pointer table.
// Reading that table as native char** combines adjacent entries on 64-bit
// hosts, so the active implementation keeps the exact keys independently of
// pointer width. Invalid wire values are rejected before callbacks.
func spellResult4FB0B0(status uint32, hooks spellResultHooks4FB0B0) bool {
	if status >= uint32(len(spellResultKeys4FB0B0)) {
		return false
	}
	key := spellResultKeys4FB0B0[status]
	text := hooks.loadString(key, spellResultSourcePath4FB0B0)
	hooks.printCentered(text)
	return true
}

// SpellResult4FB0B0 supplies live client services to the width-independent
// model of GAME.EXE 004FB0B0.
func (c *Client) SpellResult4FB0B0(status uint32) bool {
	return spellResult4FB0B0(status, spellResultHooks4FB0B0{
		loadString: func(key, source string) string {
			return c.Strings().GetStringInFile(strman.ID(key), source)
		},
		printCentered: nox_xxx_printCentered_445490,
	})
}
