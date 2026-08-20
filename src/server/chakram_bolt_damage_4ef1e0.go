package server

import noxflags "github.com/opennox/opennox/v1/common/flags"

const (
	boltDamageCooperativeFlag4EF1E0 = uint32(0x800)
	boltDamageArcherBoltName4EF1E0  = "ArcherBolt"
	boltDamageSoloMinimumKey4EF1E0  = "BoltSoloDamageMin"

	// Kept while the already-restored Chakram and Arrow callers migrate to the
	// authoritative native adapter in the next porting slice.
	chakramArcherBoltTypeName4EF1E0 = boltDamageArcherBoltName4EF1E0
)

type boltDamageHooks4EF1E0[M any] struct {
	loadCachedArcherBoltType  func() uint32
	lookupType                func(string) uint32
	storeCachedArcherBoltType func(uint32)
	gameFlagsCheck            func(uint32) int32
	loadModifierArg           func() M
	loadModifierType          func(M) uint32
	balanceFloat              func(string) float64
	loadStrengthArg           func() int32
	loadRequiredStrength      func(M) uint16
	loadDamageMinimum         func(M) uint16
	loadCoefficient           func(M) float32
}

// Keep the Win32 x87 53-bit arithmetic boundaries explicit and prevent a
// host compiler from contracting the multiply and add into one operation.
//
//go:noinline
func boltDamageMul64_4EF1E0(a, b float64) float64 { return a * b }

//go:noinline
func boltDamageAdd64_4EF1E0(a, b float64) float64 { return a + b }

// boltDamage4EF1E0 preserves GAME.EXE 004EF1E0. The zero-valued ArcherBolt
// cache is looked up and stored before any argument or game-mode observation;
// a zero lookup is deliberately retried by the next call. The modifier
// argument is loaded only after the mode callback, and only an exact callback
// result of one enters the cooperative type test. That test reloads the live
// cache after reading TypeInd. A matching ArcherBolt loads the balance minimum
// before strength and modifier fields and never reads DamageMin. Every other
// path reads strength, unsigned required strength, unsigned minimum, then the
// binary32 coefficient. Signed subtraction wraps at 32 bits, while the x87
// multiply and final add are represented by explicit binary64 boundaries.
func boltDamage4EF1E0[M any](hooks boltDamageHooks4EF1E0[M]) float64 {
	if hooks.loadCachedArcherBoltType() == 0 {
		projectileType := hooks.lookupType(boltDamageArcherBoltName4EF1E0)
		hooks.storeCachedArcherBoltType(projectileType)
	}

	mode := hooks.gameFlagsCheck(boltDamageCooperativeFlag4EF1E0)
	modifier := hooks.loadModifierArg()
	if mode == 1 {
		modifierType := hooks.loadModifierType(modifier)
		archerBoltType := hooks.loadCachedArcherBoltType()
		if modifierType == archerBoltType {
			minimum := hooks.balanceFloat(boltDamageSoloMinimumKey4EF1E0)
			strength := hooks.loadStrengthArg()
			required := hooks.loadRequiredStrength(modifier)
			delta := strength - int32(required)
			coefficient := hooks.loadCoefficient(modifier)
			scaled := boltDamageMul64_4EF1E0(float64(delta), float64(coefficient))
			return boltDamageAdd64_4EF1E0(minimum, scaled)
		}
	}

	strength := hooks.loadStrengthArg()
	required := hooks.loadRequiredStrength(modifier)
	delta := strength - int32(required)
	minimum := hooks.loadDamageMinimum(modifier)
	coefficient := hooks.loadCoefficient(modifier)
	scaled := boltDamageMul64_4EF1E0(float64(delta), float64(coefficient))
	return boltDamageAdd64_4EF1E0(scaled, float64(minimum))
}

// chakramCalcBoltDamage4EF1E0 temporarily preserves the value-oriented API
// used by the restored collision paths. It delegates arithmetic and branch
// semantics to the authoritative ordered contract above.
func chakramCalcBoltDamage4EF1E0(
	strength int32,
	modifier *Modifier,
	cooperative bool,
	archerBoltType uint32,
	boltSoloDamageMin float64,
) float64 {
	cache := archerBoltType
	mode := int32(0)
	if cooperative {
		mode = 1
	}
	return boltDamage4EF1E0(boltDamageHooks4EF1E0[*Modifier]{
		loadCachedArcherBoltType: func() uint32 {
			return cache
		},
		lookupType: func(string) uint32 {
			return archerBoltType
		},
		storeCachedArcherBoltType: func(value uint32) {
			cache = value
		},
		gameFlagsCheck: func(uint32) int32 {
			return mode
		},
		loadModifierArg: func() *Modifier {
			return modifier
		},
		loadModifierType: func(modifier *Modifier) uint32 {
			return modifier.TypeInd
		},
		balanceFloat: func(string) float64 {
			return boltSoloDamageMin
		},
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

func (s *Server) chakramCalcBoltDamageNative4EF1E0(strength int32, modifier *Modifier) float32 {
	return float32(chakramCalcBoltDamage4EF1E0(
		strength,
		modifier,
		noxflags.HasGame(noxflags.GameModeCoop),
		uint32(s.Types.IndByID(chakramArcherBoltTypeName4EF1E0)),
		s.Balance.Float(boltDamageSoloMinimumKey4EF1E0),
	))
}
