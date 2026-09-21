package opennox

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

// pixieCastHooks540440 isolates the world operations around the native-width
// restoration of GAME.EXE 00540440. Object pointers never cross the old
// six-int C callback ABI.
type pixieCastHooks540440 struct {
	pixieType   func() int
	countOwned  func(owner *server.Object, typeInd int32) int
	desired     func(levelIndex int) int
	randomInt   func(min, max int) int
	traceRay    func(from, to types.Pointf) bool
	newPixie    func() *server.Object
	createAt    func(pixie, owner *server.Object, pos types.Pointf)
	findTarget  func(pixie, owner *server.Object) *server.Object
	frame       func() uint32
	tickRate    func() uint32
	playCastAud func(spell.ID, *server.Object)
}

func castPixiesNative540440(
	spellID spell.ID,
	owner, caster *server.Object,
	level int,
	hooks pixieCastHooks540440,
) int {
	if owner == nil || caster == nil {
		return 0
	}
	pixieType := hooks.pixieType()
	current := hooks.countOwned(owner, int32(pixieType))
	desired := hooks.desired(level - 1)
	if current >= desired {
		return 1
	}

	radius := caster.Shape.Circle.R + 4
	for attempts := desired - current; attempts > 0; attempts-- {
		direction := hooks.randomInt(0, 255)
		cosine, sine := server.SinCosDir(byte(direction))
		spawn := caster.PosVec.Add(types.Ptf(radius*cosine, radius*sine))
		if !hooks.traceRay(caster.PosVec, spawn) {
			continue
		}
		pixie := hooks.newPixie()
		if pixie == nil {
			continue
		}
		hooks.createAt(pixie, owner, spawn)
		pixie.Direction1 = server.Dir16(direction)
		pixie.Direction2 = server.Dir16(direction)
		pixie.VelVec = types.Pointf{}

		update := pixie.UpdateDataPixie()
		update.Target = hooks.findTarget(pixie, owner)
		update.Owner = owner
		update.SpellID = int32(spellID)
		pixie.Pos39 = caster.PosVec
		update.Deadline = hooks.frame() + hooks.tickRate()*uint32(hooks.randomInt(30, 90))
		update.LastOwnerVisibleFrame = hooks.frame()
	}
	hooks.playCastAud(spellID, caster)
	return 1
}

func (s *Server) castPixies540440(
	spellID spell.ID,
	_ *server.Object,
	owner, caster *server.Object,
	_ *server.SpellAcceptArg,
	level int,
) int {
	return castPixiesNative540440(spellID, owner, caster, level, pixieCastHooks540440{
		pixieType: s.Types.PixieID,
		countOwned: func(owner *server.Object, typeInd int32) int {
			return int(owner.CountSubOfType(typeInd))
		},
		desired: func(levelIndex int) int {
			return int(s.Balance.FloatInd("PixieCount", levelIndex))
		},
		randomInt: s.Rand.Logic.IntClamp,
		traceRay: func(from, to types.Pointf) bool {
			return s.MapTraceRay(from, to, server.MapTraceFlag1|server.MapTraceFlag3)
		},
		newPixie: func() *server.Object {
			return s.NewObjectByTypeID("Pixie")
		},
		createAt: func(pixie, owner *server.Object, pos types.Pointf) {
			s.CreateObjectAt(pixie, owner, pos)
		},
		findTarget: func(pixie, owner *server.Object) *server.Object {
			return s.Nox_xxx_spellFlySearchTarget(nil, pixie, things.SpellOffensive, 600, 0, owner)
		},
		frame:    s.Frame,
		tickRate: s.TickRate,
		playCastAud: func(spellID spell.ID, caster *server.Object) {
			s.Audio.EventObj(s.Spells.DefByInd(spellID).GetCastSound(), caster, 0, 0)
		},
	})
}
