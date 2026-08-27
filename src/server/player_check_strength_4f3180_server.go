package server

type playerCheckStrengthNativeDeps4F3180 struct {
	getUnitStrength func(*Object) int32
	findArmorDef    func(uint16) *Modifier
	findWeaponDef   func(uint16) *Modifier
}

func playerCheckStrengthNative4F3180(
	player, item *Object,
	deps playerCheckStrengthNativeDeps4F3180,
) int32 {
	return playerCheckStrength4F3180(player, item, playerCheckStrengthHooks4F3180[
		*Object,
		*Modifier,
	]{
		loadPlayerClassLow: func(object *Object) uint8 {
			return uint8(object.ObjClass)
		},
		getUnitStrength: deps.getUnitStrength,
		loadItemClass: func(object *Object) uint32 {
			return uint32(object.ObjClass)
		},
		loadItemType: func(object *Object) uint16 {
			return uint16(object.TypeInd)
		},
		findArmorDef:  deps.findArmorDef,
		findWeaponDef: deps.findWeaponDef,
		loadRequired: func(modifier *Modifier) uint16 {
			return modifier.ReqStrength60
		},
	})
}

// PlayerCheckStrength4F3180 checks the live equipment definition using native
// Object and Modifier pointers while retaining the original signed 32-bit
// strength result and canonical 0/1 return.
//
//go:noinline
func (s *Server) PlayerCheckStrength4F3180(player, item *Object) int32 {
	return playerCheckStrengthNative4F3180(player, item, playerCheckStrengthNativeDeps4F3180{
		getUnitStrength: func(object *Object) int32 {
			return int32(object.Strength())
		},
		findArmorDef: func(typeInd uint16) *Modifier {
			return s.Modif.Nox_xxx_equipClothFindDefByTT413270(int(typeInd))
		},
		findWeaponDef: func(typeInd uint16) *Modifier {
			return s.Modif.Nox_xxx_getProjectileClassById413250(int(typeInd))
		},
	})
}
