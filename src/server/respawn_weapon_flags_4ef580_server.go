package server

// RespawnWeaponFlags4EF580 binds GAME.EXE 004EF580 to the native armor,
// weapon, and object-type tables. Masks remain fixed-width dwords and the
// result remains the exact packet byte on every host pointer width.
func (s *Server) RespawnWeaponFlags4EF580() uint8 {
	return respawnWeaponFlags4EF580(respawnWeaponFlagsHooks4EF580{
		lookupArmor:  s.Armor.Sub_415CD0,
		lookupWeapon: s.Weapons.Sub_415840,
		allowed: func(ind int) bool {
			return s.Types.ByInd(ind).Allowed()
		},
	})
}
