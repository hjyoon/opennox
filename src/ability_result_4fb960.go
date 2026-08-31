package opennox

import "github.com/opennox/libs/strman"

const abilityResultSourcePath4FB960 = `C:\NoxPost\src\Server\Ability\Ability.c`

var abilityResultKeys4FB960 = [...]string{
	"AbilityOK",
	"AbilityInUse",
	"AbilityNotReady",
	"BadSkill",
	"Illegal",
	"AbilityRestrictedByFlag",
	"AbilityRestrictedWhileJumping",
	"player.c:TooHeavy",
}

type abilityResultHooks4FB960 struct {
	loadString    func(key, source string) string
	printCentered func(text string)
}

// abilityResult4FB960 preserves the defined input domain of GAME.EXE
// 004FB960. The original indexes an eight-entry PE32 pointer table with the
// complete uint32 status and has undefined behavior outside indices 0..7.
// Invalid wire values are rejected before callbacks so a malformed inform
// packet cannot turn that legacy table read into a native pointer fault.
func abilityResult4FB960(status uint32, hooks abilityResultHooks4FB960) bool {
	if status >= uint32(len(abilityResultKeys4FB960)) {
		return false
	}
	key := abilityResultKeys4FB960[status]
	text := hooks.loadString(key, abilityResultSourcePath4FB960)
	hooks.printCentered(text)
	return true
}

// AbilityResult4FB960 supplies live client services to the width-independent
// model of GAME.EXE 004FB960.
func (c *Client) AbilityResult4FB960(status uint32) bool {
	return abilityResult4FB960(status, abilityResultHooks4FB960{
		loadString: func(key, source string) string {
			return c.Strings().GetStringInFile(strman.ID(key), source)
		},
		printCentered: nox_xxx_printCentered_445490,
	})
}
