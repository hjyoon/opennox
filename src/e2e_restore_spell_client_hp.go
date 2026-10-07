package opennox

import "github.com/opennox/libs/spell"

// e2eRestoreNPCClientHP follows the original owner-current-HP wire contract,
// not a synthesized healing-number event. 004D8620 transmits BYTE(Cur >> 1);
// client 0048EA70 opcode 65 applies WORD(2 * that unsigned BYTE). RestoreMana
// leaves a non-player's injury unchanged, including this client observation.
func e2eRestoreNPCClientHP(id spell.ID, injured, maximum uint16) uint16 {
	current := maximum
	if id == spell.SPELL_RESTORE_MANA {
		current = injured
	}
	return 2 * uint16(byte(current>>1))
}
