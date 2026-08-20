package server

import noxflags "github.com/opennox/opennox/v1/common/flags"

type boltDamageNativeDeps4EF1E0 struct {
	loadCachedArcherBoltType  func() uint32
	lookupType                func(string) uint32
	storeCachedArcherBoltType func(uint32)
	gameFlagsCheck            func(uint32) int32
	balanceFloat              func(string) float64
}

// BoltDamageModifierValues4EF1E0 is the pointer-free form used when an
// original caller synthesized only the four relevant modifier fields on its
// stack instead of passing a modifier-table entry.
type BoltDamageModifierValues4EF1E0 struct {
	TypeIndex        uint32
	RequiredStrength uint16
	Coefficient      float32
	Minimum          uint16
}

func boltDamageNative4EF1E0(
	strength int32,
	modifier *Modifier,
	deps boltDamageNativeDeps4EF1E0,
) float64 {
	return boltDamage4EF1E0(boltDamageHooks4EF1E0[*Modifier]{
		loadCachedArcherBoltType:  deps.loadCachedArcherBoltType,
		lookupType:                deps.lookupType,
		storeCachedArcherBoltType: deps.storeCachedArcherBoltType,
		gameFlagsCheck:            deps.gameFlagsCheck,
		loadModifierArg: func() *Modifier {
			return modifier
		},
		loadModifierType: func(modifier *Modifier) uint32 {
			return modifier.TypeInd
		},
		balanceFloat: deps.balanceFloat,
		loadStrengthArg: func() int32 {
			return strength
		},
		loadRequiredStrength: func(modifier *Modifier) uint16 {
			return modifier.ReqStrength60
		},
		loadDamageMinimum: func(modifier *Modifier) uint16 {
			return modifier.DamageMin72
		},
		loadCoefficient: func(modifier *Modifier) float32 {
			return modifier.DamageCoeffOrArmor64
		},
	})
}

func (s *Server) boltDamageNativeDeps4EF1E0() boltDamageNativeDeps4EF1E0 {
	return boltDamageNativeDeps4EF1E0{
		loadCachedArcherBoltType: func() uint32 {
			return s.Modif.boltDamageArcherType4EF1E0
		},
		lookupType: func(name string) uint32 {
			return uint32(s.Types.IndByID(name))
		},
		storeCachedArcherBoltType: func(value uint32) {
			s.Modif.boltDamageArcherType4EF1E0 = value
		},
		gameFlagsCheck: func(mask uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(mask)) {
				return 1
			}
			return 0
		},
		balanceFloat: s.Balance.Float,
	}
}

// CalcBoltDamage4EF1E0 binds the ordered GAME.EXE contract to the native-width
// Modifier. The cache is owned by this Server and remains distinct from the
// other ArcherBolt caches used by Arrow collision and registration code.
func (s *Server) CalcBoltDamage4EF1E0(strength int32, modifier *Modifier) float64 {
	return boltDamageNative4EF1E0(strength, modifier, s.boltDamageNativeDeps4EF1E0())
}

// CalcBoltDamageValues4EF1E0 serves the one original attack caller that built
// an 88-byte Win32 modifier-shaped temporary. It creates a native Modifier so
// no fixed 32-bit pointer layout crosses the restored boundary.
func (s *Server) CalcBoltDamageValues4EF1E0(
	strength int32,
	values BoltDamageModifierValues4EF1E0,
) float64 {
	modifier := Modifier{
		TypeInd:              values.TypeIndex,
		ReqStrength60:        values.RequiredStrength,
		DamageCoeffOrArmor64: values.Coefficient,
		DamageMin72:          values.Minimum,
	}
	return s.CalcBoltDamage4EF1E0(strength, &modifier)
}
