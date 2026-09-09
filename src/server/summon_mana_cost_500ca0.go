package server

const (
	summonManaCostPlayerClassLow500CA0 = uint8(0x04)
	summonManaCostTableAddress500CA0   = uint32(0x005bc244)
)

// summonManaCostHooks500CA0 exposes the two observable memory reads made by
// GAME.EXE 00500CA0. A generic object handle keeps its identity at native
// width, while table addresses retain the oracle's 32-bit address arithmetic.
type summonManaCostHooks500CA0[O comparable] struct {
	loadClassLow func(O) uint8
	loadCost     func(uint32) int32
}

// summonManaCostAddress500CA0 reproduces the PE32 indexed address expression
// 005BC244 + spellID*4, including uint32 wraparound for unchecked IDs.
func summonManaCostAddress500CA0(spellID int32) uint32 {
	return summonManaCostTableAddress500CA0 + uint32(spellID)*4
}

// summonManaCost500CA0 returns the summon mana cost for player units. The
// original ABI accepts any signed spell ID; its two known callers constrain
// that ID to the inclusive range 75..114 before reaching this function.
func summonManaCost500CA0[O comparable](spellID int32, unit O, hooks summonManaCostHooks500CA0[O]) int32 {
	var nilObject O
	if unit == nilObject {
		return 0
	}
	if hooks.loadClassLow(unit)&summonManaCostPlayerClassLow500CA0 == 0 {
		return 0
	}
	return hooks.loadCost(summonManaCostAddress500CA0(spellID))
}
