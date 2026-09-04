package server

func dangerousUnitNative547120(
	candidate, unit *Object,
	cache *dangerousUnitTypeCache547120,
	lookupType func(string) uint32,
	clearLocationSafe func(),
) uint32 {
	return dangerousUnit547120(candidate, unit, dangerousUnitHooks547120[*Object]{
		loadToxicCloudCache: func() uint32 {
			return cache.toxicCloud
		},
		lookupType: lookupType,
		storeToxicCloudCache: func(value uint32) {
			cache.toxicCloud = value
		},
		storeSmallToxicCloudCache: func(value uint32) {
			cache.smallToxicCloud = value
		},
		loadClass: func(candidate *Object) uint32 {
			return uint32(candidate.ObjClass)
		},
		loadType: func(candidate *Object) uint16 {
			return candidate.TypeInd
		},
		loadSmallToxicCloudCache: func() uint32 {
			return cache.smallToxicCloud
		},
		loadSubClass: func(unit *Object) uint32 {
			return uint32(unit.ObjSubClass)
		},
		clearLocationSafe: clearLocationSafe,
	})
}

// UnitIsDangerous547120 binds GAME.EXE 00547120 to native-width Object
// pointers. It reports the callback's location-safe side effect; the original
// scalar return is intentionally internal because its sole caller discards it.
//
//go:noinline
func (s *Server) UnitIsDangerous547120(candidate, unit *Object) bool {
	dangerous := false
	dangerousUnitNative547120(
		candidate,
		unit,
		&s.dangerousUnitTypes547120,
		func(name string) uint32 {
			return uint32(s.Types.IndByID(name))
		},
		func() {
			dangerous = true
		},
	)
	return dangerous
}
