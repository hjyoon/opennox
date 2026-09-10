package server

const (
	audioEventZonePlayerClass501C00  = uint8(0x04)
	audioEventZoneMonsterClass501C00 = uint8(0x02)
)

// audioEventZoneHooks501C00 exposes every observable load and call made by
// GAME.EXE 00501C00. All handles remain native-width and comparable so the
// function can preserve the original null tests without narrowing pointers.
type audioEventZoneHooks501C00[P, O, U, L, G comparable] struct {
	loadClassLow         func(O) uint8
	loadUpdate           func(O) U
	loadPlayer           func(U) L
	loadPlayerZone       func(L) uint8
	loadMonsterPolygonID func(U) uint32
	polygonByID          func(uint32) G
	loadPolygonZone      func(G) uint8
	loadPositionX        func(P) float32
	floatToInt           func(float32) int32
	loadPositionY        func(P) float32
	polygonAtPoint       func([2]int32, uint32) G
}

// audioEventZone501C00 preserves GAME.EXE 00501C00's exact branch and access
// order. Player classification wins when both class bits are set. A nonzero
// object-derived zone returns before the position pointer is accessed; every
// other path converts X before loading and converting Y for the polygon
// fallback. The fallback's previous polygon ID is always zero.
func audioEventZone501C00[P, O, U, L, G comparable](
	position P,
	object O,
	hooks audioEventZoneHooks501C00[P, O, U, L, G],
) uint8 {
	var zone uint8
	var nilObject O
	var nilPolygon G
	if object != nilObject {
		class := hooks.loadClassLow(object)
		if class&audioEventZonePlayerClass501C00 != 0 {
			update := hooks.loadUpdate(object)
			player := hooks.loadPlayer(update)
			zone = hooks.loadPlayerZone(player)
			if zone != 0 {
				return zone
			}
		} else if class&audioEventZoneMonsterClass501C00 != 0 {
			update := hooks.loadUpdate(object)
			polygonID := hooks.loadMonsterPolygonID(update)
			polygon := hooks.polygonByID(polygonID)
			if polygon != nilPolygon {
				zone = hooks.loadPolygonZone(polygon)
				if zone != 0 {
					return zone
				}
			}
		}
	}

	x := hooks.loadPositionX(position)
	pointX := hooks.floatToInt(x)
	y := hooks.loadPositionY(position)
	pointY := hooks.floatToInt(y)
	polygon := hooks.polygonAtPoint([2]int32{pointX, pointY}, 0)
	if polygon != nilPolygon {
		return hooks.loadPolygonZone(polygon)
	}
	return zone
}
