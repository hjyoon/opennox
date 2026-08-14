package opennox

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func nox_xxx_updateDeathBall_53D080(obj *server.Object) {
	sobj := asObjectS(obj)
	s := sobj.getServer()
	df := s.Frame() - obj.Field32
	if s.Frame()%(s.TickRate()/3) != 0 {
		found := false
		r := float32(s.Balance.Float("DeathBallCancelRange"))
		s.Map.EachMissileInCircle(obj.PosVec, r, func(it *server.Object) bool {
			if it == obj {
				return true
			}
			if int(it.TypeInd) != s.Types.DeathBallID() {
				return true
			}
			if !s.MapTraceVision(obj, it) {
				return true
			}
			found = true
			if !it.Flags().Has(object.FlagDestroyed) {
				s.Nox_xxx_netSendFxGreenBolt_523790(obj.PosVec.Point(), it.PosVec.Point(), 10)
				legacy.Nox_xxx_sMakeScorch_537AF0(it.PosVec, 1)
				asObjectS(it).Delete()
			}
			return true
		})
		if found {
			legacy.Nox_xxx_sMakeScorch_537AF0(obj.PosVec, 1)
			sobj.Delete()
		}
	}
	if df > 10 {
		r1 := float32(s.Balance.Float("DeathBallOutRadius"))
		r2 := float32(s.Balance.Float("DeathBallInRadius"))
		dmg := int(s.Balance.Float("DeathBallNearbyDamage"))
		s.Nox_xxx_mapDamageUnitsAround(obj.PosVec, r1, r2, dmg, object.DamageCrush, obj, nil, doDamageWalls)
	}
	if df > s.SecToFrames(4) {
		sobj.Delete()
	}
}

func nox_xxx_collideDeathBall_4E9E90(ball *server.Object, targ *server.Object, pos *types.Pointf) {
	s := noxServer
	s.S().DeathBallCollide4E9E90(ball, targ, pos, server.DeathBallCollideRuntime4E9E90{
		TraceReady: legacy.Get_dword_5d4594_2488620,
		TracePoint: func() *ntype.Point32 {
			return memmap.PtrT[ntype.Point32](0x5D4594, 2488612)
		},
		DamageMap: func(x, y, damage int32, damageType object.DamageType, source *server.Object) {
			s.Nox_xxx_damageToMap_534BC0(int(x), int(y), int(damage), damageType, source)
		},
	})
}

func nox_xxx_deathBallCreateFragments_52BD30(obj *server.Object) {
	s := asObjectS(obj).getServer()
	for i := 0; i < 3; i++ {
		fr := s.NewObjectByTypeID("DeathBallFragment")
		if fr != nil {
			s.CreateObjectAt(fr, obj.ObjOwner, obj.PosVec)
			fr.Direction1 = server.Dir16(s.Rand.Logic.IntClamp(0, math.MaxUint8))
			cos, sin := server.SinCosDir(byte(fr.Direction1))
			fr.VelVec.X = cos * fr.SpeedCur
			fr.VelVec.Y = sin * fr.SpeedCur
		}
	}
}
