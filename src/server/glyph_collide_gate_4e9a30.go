package server

const (
	glyphCollideCoopFlag4E9A30     = uint32(0x800)
	glyphCollideCoopTeamFlag4E9A30 = uint32(0x200)
	glyphCollideNoUpdateFlag4E9A30 = uint32(0x2)
	glyphCollidePlayerClass4E9A30  = uint8(0x4)
	glyphCollideTreadLightly4E9A30 = int32(4)
)

type glyphCollideGateHooks4E9A30[O comparable] struct {
	gameFlag        func(uint32) int32
	firstPlayerUnit func() O
	loadFlags       func(O) uint32
	unitsOnSameTeam func(O, O) int32
	findParent      func(O) O
	classLow        func(O) uint8
	abilityActive   func(O, int32) int32
}

// glyphCollideGate4E9A30 preserves GAME.EXE 004E9A30. Coop suppression is
// entered only when gameFlags_check(0x800) returns exactly one. Outside that
// early return, ability 4 is queried last even for a nil target. For non-nil
// targets, same-team wins before the 0x200 parent-player test; parent lookup
// order is source then target, and both lookups finish before either class is
// read. Every accepted/rejected result is canonical one/zero.
func glyphCollideGate4E9A30[O comparable](
	source, target O,
	hooks glyphCollideGateHooks4E9A30[O],
) int32 {
	if hooks.gameFlag(glyphCollideCoopFlag4E9A30) == 1 {
		first := hooks.firstPlayerUnit()
		if hooks.loadFlags(first)&glyphCollideNoUpdateFlag4E9A30 == glyphCollideNoUpdateFlag4E9A30 {
			return 0
		}
	}

	result := int32(1)
	var zero O
	if target != zero {
		if hooks.unitsOnSameTeam(source, target) != 0 {
			result = 0
		} else if hooks.gameFlag(glyphCollideCoopTeamFlag4E9A30) != 0 {
			sourceParent := hooks.findParent(source)
			targetParent := hooks.findParent(target)
			if hooks.classLow(sourceParent)&glyphCollidePlayerClass4E9A30 != 0 &&
				hooks.classLow(targetParent)&glyphCollidePlayerClass4E9A30 != 0 {
				result = 0
			}
		}
	}
	if hooks.abilityActive(target, glyphCollideTreadLightly4E9A30) != 0 {
		result = 0
	}
	return result
}
