package server

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/common/sound"
)

type earthquakeCastDeps52DE40 struct {
	storeLevel func(int32)
	loadLevel  func() int32
	balance    func(string) float32
	indexed    func(string, int32) float32
	circle     func(*Object, float32, func(*Object))
	owner      func(*Object) *Object
	hp         func(*Object) uint16
	distance   func(*Object, *Object) float64
	damage     func(*Object, *Object, *Object, int32, object.DamageType)
	castSound  func(int32) sound.ID
	audio      func(sound.ID, *Object, int, uint32)
	quake      func(*Object, int32)
}

// earthquakeJiggle52DE40 models the 00419A70 signed-DWORD FISTP under
// gameplay's round-toward-zero mode. An invalid conversion stores integer
// indefinite. Damage instead uses 00566DCC's signed-QWORD/low-DWORD contract.
func earthquakeJiggle52DE40(value float32) int32 {
	if math.IsNaN(float64(value)) || value >= 2147483648 || value < -2147483648 {
		return math.MinInt32
	}
	return int32(value)
}

// earthquakeDamage52DEC0 preserves GAME.EXE 0052DEC0..0052DF3F. Resolve
// and cache the owner before even the self/airborne/HP gates. The iterator
// ignores the callback's EAX; only the live damage callback's side effects
// matter. No enemy, visibility, invulnerability or positive-damage gate is
// added here: the original damage handler makes those decisions.
func earthquakeDamage52DEC0(target, caster *Object, h earthquakeCastDeps52DE40) {
	owner := h.owner(caster)
	if target == caster || target.ObjFlags.Has(object.FlagAirborne) || h.hp(target) == 0 {
		return
	}
	// 0052DEF7 spills surface distance to binary32. The division remains
	// precision-53/ToZero, followed by the live cached-level read, subtraction
	// and second binary32 spill. Reuse the qualified local x87 operations;
	// neither change thread state nor contract original instructions.
	distance := monsterMoveToRunSpill544440(h.distance(target, caster))
	rangeValue := h.balance("EarthquakeRange")
	portion := monsterMoveForceDivChop53_50D581(distance, float64(rangeValue))
	index := h.loadLevel() - 1
	falloff := monsterMoveToRunSpill544440(monsterMoveToRunAddChop53_544434(1, -portion))
	base := h.indexed("EarthquakeDamage", index)
	amount := x87TruncSignedQwordLow566DCC(monsterMoveForceMulChop53_50D4FE(float64(base), falloff))
	h.damage(target, owner, caster, amount, object.DamageImpact)
}

// earthquakeCast52DE40 restores GAME.EXE 0052DE40..0052DEB4 without
// narrowing object addresses through six C ints. Damage uses the live global
// DWORD level (including reentrant casts); Jiggle retains the entry level.
// Position is read by circle/quake after the corresponding balance lookup,
// from the same cached caster pointer. The other three pointer args are unused.
func earthquakeCast52DE40(id int32, caster *Object, level int32, h earthquakeCastDeps52DE40) int32 {
	h.storeLevel(level)
	rangeValue := h.balance("EarthquakeRange")
	h.circle(caster, rangeValue, func(target *Object) { earthquakeDamage52DEC0(target, caster, h) })
	castSound := h.castSound(id)
	h.audio(castSound, caster, 0, 0)
	jiggle := h.indexed("EarthquakeJiggle", level-1)
	h.quake(caster, earthquakeJiggle52DE40(jiggle))
	return 1
}

type EarthquakeCastRuntime52DE40 struct {
	LevelCache *uint32
}

func (s *Server) earthquakeCastDeps52DE40(runtime EarthquakeCastRuntime52DE40) earthquakeCastDeps52DE40 {
	return earthquakeCastDeps52DE40{
		storeLevel: func(level int32) { *runtime.LevelCache = uint32(level) },
		loadLevel:  func() int32 { return int32(*runtime.LevelCache) },
		balance:    func(key string) float32 { return float32(s.Balance.Float(key)) },
		indexed:    func(key string, index int32) float32 { return float32(s.Balance.FloatInd(key, int(index))) },
		circle: func(caster *Object, radius float32, callback func(*Object)) {
			s.Map.EachObjInCircle(caster.PosVec, radius, func(target *Object) bool {
				callback(target)
				return true // original iterator ignores the damage callback's return
			})
		},
		owner:    (*Object).FindOwnerChainPlayer,
		hp:       UnitGetHP4EE780,
		distance: ObjectDistance4E6C00,
		damage: func(target, owner, caster *Object, amount int32, typ object.DamageType) {
			objDamage.Get(target.Damage)(target, owner, caster, amount, typ)
		},
		castSound: func(id int32) sound.ID { return s.Spells.DefByInd(spell.ID(id)).GetCastSound() },
		audio:     func(id sound.ID, caster *Object, kind int, code uint32) { s.Audio.EventObj(id, caster, kind, code) },
		quake:     func(caster *Object, jiggle int32) { s.Nox_xxx_earthquakeSend_4D9110(caster.PosVec, int(jiggle)) },
	}
}

func (s *Server) CastEarthquake52DE40(id int32, _, _ *Object, caster *Object, _ *SpellAcceptArg, level int32, runtime EarthquakeCastRuntime52DE40) int32 {
	return earthquakeCast52DE40(id, caster, level, s.earthquakeCastDeps52DE40(runtime))
}
