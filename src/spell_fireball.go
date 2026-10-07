package opennox

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

// fireballCastHooks52C790 isolates the world operations around the
// native-width restoration of GAME.EXE 0052C790. In particular, object
// pointers no longer cross the original six-int C callback ABI.
type fireballCastHooks52C790 struct {
	newObject     func(string) *server.Object
	traceRay      func(types.Pointf, types.Pointf) bool
	createAt      func(*server.Object, *server.Object, types.Pointf)
	speedCoeff    func(int) float64
	playCastAudio func(spell.ID, *server.Object)
}

func fireballProjectileType52C790(level int) string {
	switch level {
	case 1:
		return "Fireball"
	case 2:
		return "StrongFireball"
	case 3, 4, 5:
		return "TitanFireball"
	default:
		return ""
	}
}

// castFireballNative52C790 follows GAME.EXE 0052C790 using native Go object
// pointers. The legacy routine accepted the caster as a 32-bit int and
// truncated every 64-bit address before reading its position and radius.
func castFireballNative52C790(
	spellID spell.ID,
	caster *server.Object,
	level int,
	hooks fireballCastHooks52C790,
) int {
	if caster == nil {
		return 0
	}
	typeID := fireballProjectileType52C790(level)
	if typeID == "" {
		return 1
	}
	projectile := hooks.newObject(typeID)
	if projectile == nil {
		return 1
	}

	cosine, sine := server.SinCosDir(byte(caster.Direction1))
	radius := float64(caster.Shape.Circle.R)
	radius2 := fireballAddChop53_52C790(radius, radius)
	from := caster.PosVec
	// X is spilled at 0052C7FD before velocity is added. Y remains in the
	// x87 register through both additions and is spilled only at 0052C816.
	x := fireballMulChop53_52C790(radius2, float64(cosine))
	x = fireballAddChop53_52C790(x, float64(from.X))
	x = float64(fireballSpill32_52C790(x))
	x = fireballAddChop53_52C790(x, float64(caster.VelVec.X))
	y := fireballMulChop53_52C790(radius2, float64(sine))
	y = fireballAddChop53_52C790(y, float64(from.Y))
	y = fireballAddChop53_52C790(y, float64(caster.VelVec.Y))
	spawn := types.Ptf(fireballSpill32_52C790(x), fireballSpill32_52C790(y))
	if !hooks.traceRay(from, spawn) {
		spawn = from // The original fallback uses the pre-trace snapshot.
	}
	hooks.createAt(projectile, caster, spawn)

	coefficient := hooks.speedCoeff(level - 1)
	speed := fireballMulChop53_52C790(coefficient, float64(projectile.SpeedCur))
	projectile.SpeedCur = fireballSpill32_52C790(speed)
	// The non-popping speed/Y stores retain their 53-bit register values;
	// X alone is reloaded from its binary32 temporary at 0052C884.
	xProduct := fireballSpill32_52C790(fireballMulChop53_52C790(speed, float64(cosine)))
	projectile.VelVec.X = xProduct
	yProduct := fireballMulChop53_52C790(speed, float64(sine))
	projectile.VelVec.Y = fireballSpill32_52C790(yProduct)
	projectile.VelVec.X = fireballSpill32_52C790(fireballAddChop53_52C790(float64(xProduct), float64(caster.VelVec.X)))
	projectile.VelVec.Y = fireballSpill32_52C790(fireballAddChop53_52C790(yProduct, float64(caster.VelVec.Y)))
	direction := caster.Direction1
	projectile.Direction1 = direction
	projectile.Direction2 = direction
	hooks.playCastAudio(spellID, caster)
	return 1
}

func (s *Server) castFireball52C790(
	spellID spell.ID,
	_ *server.Object,
	_ *server.Object,
	caster *server.Object,
	_ *server.SpellAcceptArg,
	level int,
) int {
	return castFireballNative52C790(spellID, caster, level, fireballCastHooks52C790{
		newObject: s.NewObjectByTypeID,
		traceRay: func(from, to types.Pointf) bool {
			return s.MapTraceRay(from, to, server.MapTraceFlag1|server.MapTraceFlag3)
		},
		createAt: func(projectile, owner *server.Object, position types.Pointf) {
			s.CreateObjectAt(projectile, owner, position)
		},
		speedCoeff: func(levelIndex int) float64 {
			return s.Balance.FloatInd("FireballSpeedCoeff", levelIndex)
		},
		playCastAudio: func(id spell.ID, caster *server.Object) {
			s.Audio.EventObj(s.Spells.DefByInd(id).GetCastSound(), caster, 0, 0)
		},
	})
}
