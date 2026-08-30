package server

const (
	mapFindPlayerStartType4F7AB0       = "PlayerStart"
	mapFindPlayerStartSource4F7AB0     = `C:\NoxPost\src\Server\Object\update\Player.c`
	mapFindPlayerStartSourceLine4F7AB0 = int32(0x116)
	mapFindPlayerStartEnabled4F7CE0    = uint32(0x01000000)
	mapFindPlayerStartEmpty4F7AB0      = float32(2000.0)
	mapFindPlayerStartNearest4F7AB0    = float32(10000000.0)
)

type mapFindPlayerStartHooks4F7AB0[O comparable] struct {
	loadCachedType  func() uint32
	lookupType      func(string) uint32
	storeCachedType func(uint32)

	hasTeam    func(O) bool
	loadTeamID func(O) uint8
	touchTeam  func(uint8)

	firstObject     func() O
	nextObject      func(O) O
	loadTypeIndex   func(O) uint16
	loadObjectFlags func(O) uint32
	teamContains    func(O, uint8) bool

	firstPlayer func() O
	nextPlayer  func(O) O
	isEnemyTo   func(O, O) bool
	loadPosX    func(O) float32
	loadPosY    func(O) float32

	randomInt func(minimum, maximum int32, source string, line int32) int32

	storeOutputX func(float32)
	storeOutputY func(float32)
}

// mapFindPlayerStartEligible4F7CE0 preserves GAME.EXE 004F7CE0. The enabled
// flag is read before the requested team. An unscoped request and an unteamed
// start short-circuit before the active-membership comparison.
func mapFindPlayerStartEligible4F7CE0[O comparable](
	object O,
	teamID uint8,
	hooks mapFindPlayerStartHooks4F7AB0[O],
) bool {
	if hooks.loadObjectFlags(object)&mapFindPlayerStartEnabled4F7CE0 == 0 {
		return false
	}
	if teamID == 0 {
		return true
	}
	if !hooks.hasTeam(object) {
		return true
	}
	return hooks.teamContains(object, teamID)
}

// Keep each x87 arithmetic instruction at an explicit binary64 boundary.
// GAME.EXE runs this code with 53-bit precision control; the noinline helpers
// also prevent a target compiler from contracting the multiply/add into FMA.
//
//go:noinline
func mapFindPlayerStartSub64_4F7AB0(left, right float64) float64 {
	return left - right
}

//go:noinline
func mapFindPlayerStartMul64_4F7AB0(left, right float64) float64 {
	return left * right
}

//go:noinline
func mapFindPlayerStartAdd64_4F7AB0(left, right float64) float64 {
	return left + right
}

// mapFindPlayerStartDistance4F7AB0 models the original x87 expression
// dy*dy + dx*dx. The caller performs the sole FSTP/F32 rounding when it
// updates nearest.
func mapFindPlayerStartDistance4F7AB0(startX, otherX, startY, otherY float32) float64 {
	dx := mapFindPlayerStartSub64_4F7AB0(float64(startX), float64(otherX))
	dy := mapFindPlayerStartSub64_4F7AB0(float64(startY), float64(otherY))
	dySquared := mapFindPlayerStartMul64_4F7AB0(dy, dy)
	dxSquared := mapFindPlayerStartMul64_4F7AB0(dx, dx)
	return mapFindPlayerStartAdd64_4F7AB0(dySquared, dxSquared)
}

// mapFindPlayerStart4F7AB0 preserves GAME.EXE 004F7AB0. In particular, the
// type cache is populated before the nil-player return and reloaded for every
// object comparison. The first traversal remembers the last matching start,
// even when it is ineligible. Enemy-distance selection uses the x87 status-bit
// behavior: unordered values replace nearest, but only an ordered strict
// nearest > best value can become the selected start. The random fallback and
// final position read deliberately retain the original fault behavior.
func mapFindPlayerStart4F7AB0[O comparable](
	player O,
	hooks mapFindPlayerStartHooks4F7AB0[O],
) {
	if hooks.loadCachedType() == 0 {
		hooks.storeCachedType(hooks.lookupType(mapFindPlayerStartType4F7AB0))
	}

	var zero O
	if player == zero {
		return
	}

	var teamID uint8
	if hooks.hasTeam(player) {
		teamID = hooks.loadTeamID(player)
		hooks.touchTeam(teamID)
	}

	var eligibleCount int32
	var fallback O
	object := hooks.firstObject()
	if object == zero {
		hooks.storeOutputX(mapFindPlayerStartEmpty4F7AB0)
		hooks.storeOutputY(mapFindPlayerStartEmpty4F7AB0)
		return
	}
	for object != zero {
		cachedType := hooks.loadCachedType()
		typeIndex := uint32(hooks.loadTypeIndex(object))
		if typeIndex == cachedType {
			fallback = object
			if mapFindPlayerStartEligible4F7CE0(object, teamID, hooks) {
				eligibleCount++
			}
		}
		object = hooks.nextObject(object)
	}

	if eligibleCount == 0 {
		if fallback == zero {
			hooks.storeOutputX(mapFindPlayerStartEmpty4F7AB0)
			hooks.storeOutputY(mapFindPlayerStartEmpty4F7AB0)
			return
		}
		x := hooks.loadPosX(fallback)
		hooks.storeOutputX(x)
		y := hooks.loadPosY(fallback)
		hooks.storeOutputY(y)
		return
	}

	bestDistance := float32(0)
	var selected O
	noEnemies := true
	for object = hooks.firstObject(); object != zero; object = hooks.nextObject(object) {
		cachedType := hooks.loadCachedType()
		typeIndex := uint32(hooks.loadTypeIndex(object))
		if typeIndex != cachedType || !mapFindPlayerStartEligible4F7CE0(object, teamID, hooks) {
			continue
		}

		nearest := mapFindPlayerStartNearest4F7AB0
		for other := hooks.firstPlayer(); other != zero; other = hooks.nextPlayer(other) {
			if other == player || !hooks.isEnemyTo(player, other) {
				continue
			}

			startX := hooks.loadPosX(object)
			otherX := hooks.loadPosX(other)
			startY := hooks.loadPosY(object)
			otherY := hooks.loadPosY(other)
			distance := mapFindPlayerStartDistance4F7AB0(startX, otherX, startY, otherY)

			// FCOM nearest tests C0 only. C0 is set for ordered less-than and
			// unordered, so NaN replaces nearest; a later finite value also
			// replaces a NaN nearest.
			if !(distance >= float64(nearest)) {
				nearest = float32(distance)
			}
			noEnemies = false
		}

		// FCOMP best tests C0|C3, accepting ordered strict greater only.
		if nearest > bestDistance {
			selected = object
			bestDistance = nearest
		}
	}

	if noEnemies || selected == zero {
		index := hooks.randomInt(
			0,
			eligibleCount-1,
			mapFindPlayerStartSource4F7AB0,
			mapFindPlayerStartSourceLine4F7AB0,
		)
		selected = zero
		for object = hooks.firstObject(); object != zero; object = hooks.nextObject(object) {
			cachedType := hooks.loadCachedType()
			typeIndex := uint32(hooks.loadTypeIndex(object))
			if typeIndex != cachedType || !mapFindPlayerStartEligible4F7CE0(object, teamID, hooks) {
				continue
			}
			if index == 0 {
				selected = object
				break
			}
			index--
		}
	}

	x := hooks.loadPosX(selected)
	hooks.storeOutputX(x)
	y := hooks.loadPosY(selected)
	hooks.storeOutputY(y)
}
