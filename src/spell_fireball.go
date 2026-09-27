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
	direction := types.Ptf(cosine, sine)
	spawn := caster.PosVec.Add(caster.VelVec).Add(direction.Mul(2 * caster.Shape.Circle.R))
	if !hooks.traceRay(caster.PosVec, spawn) {
		spawn = caster.PosVec
	}
	hooks.createAt(projectile, caster, spawn)

	speed := float32(hooks.speedCoeff(level-1) * float64(projectile.SpeedCur))
	projectile.SpeedCur = speed
	projectile.VelVec = caster.VelVec.Add(direction.Mul(speed))
	projectile.Direction1 = caster.Direction1
	projectile.Direction2 = caster.Direction1
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
