package server

import "unsafe"

// SkullUpdateData is the exact pointer-independent 52-byte update record used
// by SkullUpdate. SkullInit resolves ProjectileName and stores ProjectileType;
// the other fields are named only to keep those adjacent bytes protected.
type SkullUpdateData struct {
	ScanDelay      uint32
	FireDelay      uint32
	TargetReady    uint8
	Field9         [3]byte
	ProjectileType uint32
	ProjectileName [32]byte
	Enabled        uint8
	Field49        [3]byte
}

var directionToAngleTable509E00 = [...]uint32{
	160, 192, 224,
	128, 0, 0,
	96, 64, 32,
}

func directionToAngleNative509E00(data *DirectionInitData) uint32 {
	return directionToAngle509E00(data, directionToAngleHooks509E00[*DirectionInitData]{
		loadY: func(data *DirectionInitData) int32 {
			return data.Y
		},
		loadX: func(data *DirectionInitData) int32 {
			return data.X
		},
		loadTable: func(index int32) uint32 {
			// GAME.EXE addresses this nine-dword table from its center. Valid
			// thing definitions produce indices -4 through 4. Reading adjacent
			// PE data for a malformed definition is intentionally not recreated.
			if index < -4 || index > 4 {
				panic("direction index outside sealed GAME.EXE table")
			}
			return directionToAngleTable509E00[index+4]
		},
	})
}

func skullProjectileName4F0450(update *SkullUpdateData) string {
	for index, value := range update.ProjectileName {
		if value == 0 {
			return string(update.ProjectileName[:index])
		}
	}
	return string(update.ProjectileName[:])
}

type skullInitNativeDeps4F0450 struct {
	lookupType func(string) int32
}

func skullInitNative4F0450(unit *Object, deps skullInitNativeDeps4F0450) int32 {
	return skullInit4F0450(unit, skullInitHooks4F0450[*Object, *DirectionInitData, *SkullUpdateData]{
		loadInitData: func(unit *Object) *DirectionInitData {
			return (*DirectionInitData)(unit.InitData)
		},
		loadUpdateData: func(unit *Object) *SkullUpdateData {
			return (*SkullUpdateData)(unit.UpdateData)
		},
		directionToAngle: directionToAngleNative509E00,
		storeDirection2: func(unit *Object, angle uint16) {
			unit.Direction2 = Dir16(angle)
		},
		storeDirection1: func(unit *Object, angle uint16) {
			unit.Direction1 = Dir16(angle)
		},
		resolveProjectileType: func(update *SkullUpdateData) int32 {
			return deps.lookupType(skullProjectileName4F0450(update))
		},
		storeProjectileType: func(update *SkullUpdateData, value uint32) {
			update.ProjectileType = value
		},
	})
}

// SkullInit4F0450 binds GAME.EXE 004F0450 to native-width Object pointers and
// the fixed-width DirectionInitData and SkullUpdateData records. There are
// deliberately no nil guards.
//
//go:noinline
func (s *Server) SkullInit4F0450(unit *Object) int32 {
	return skullInitNative4F0450(unit, skullInitNativeDeps4F0450{
		lookupType: func(name string) int32 {
			return int32(s.Types.IndByID(name))
		},
	})
}

var (
	_ = [1]struct{}{}[8-unsafe.Sizeof(DirectionInitData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(DirectionInitData{}.X)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(DirectionInitData{}.Y)]

	_ = [1]struct{}{}[52-unsafe.Sizeof(SkullUpdateData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(SkullUpdateData{}.ScanDelay)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(SkullUpdateData{}.FireDelay)]
	_ = [1]struct{}{}[8-unsafe.Offsetof(SkullUpdateData{}.TargetReady)]
	_ = [1]struct{}{}[12-unsafe.Offsetof(SkullUpdateData{}.ProjectileType)]
	_ = [1]struct{}{}[16-unsafe.Offsetof(SkullUpdateData{}.ProjectileName)]
	_ = [1]struct{}{}[48-unsafe.Offsetof(SkullUpdateData{}.Enabled)]
)
