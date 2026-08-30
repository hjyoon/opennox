package opennox

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func nox_xxx_harpoonBreakForPlr_537520(u *server.Object) {
	noxServer.abilities.harpoon.breakForOwner(u, true)
}

func nox_xxx_collideHarpoon_4EB6A0(a1c *server.Object, a2c *server.Object, collision *types.Pointf) {
	noxServer.abilities.harpoon.Collide(a1c, a2c, collision)
}

func nox_xxx_updateHarpoon_54F380(a1c *server.Object) {
	noxServer.abilities.harpoon.Update(a1c)
}

type harpoonData struct {
	*harpoonPtr
	getAim func() types.Pointf
}

var _ = [1]struct{}{}[16+2*unsafe.Sizeof(uintptr(0))-unsafe.Sizeof(harpoonPtr{})]

type harpoonPtr struct {
	target  *server.Object // 33, 132
	bolt    *server.Object // 34, 136
	frame35 uint32         // 35, 140
	targPos types.Pointf   // 36, 144
	frame38 uint32         // 38, 152
}

type abilityHarpoon struct {
	s         *Server
	damage    int32
	maxDist   float32
	minDist   float32
	maxFlight float32
	lifetime  float32
}

func (a *abilityHarpoon) Init(s *Server) {
	a.s = s
}

func (a *abilityHarpoon) Free() {
}

func (a *abilityHarpoon) getHarpoonData(u *server.Object) *harpoonData {
	if u == nil {
		return nil
	}
	switch {
	case u.Class().Has(object.ClassPlayer):
		ud := u.UpdateDataPlayer()
		pl := ud.Player
		p := (*harpoonPtr)(unsafe.Pointer(&ud.HarpoonTarg))
		return &harpoonData{harpoonPtr: p, getAim: func() types.Pointf {
			return pl.CursorPos()
		}}
	default:
		panic(u.Class().String())
	}
}

func (a *abilityHarpoon) Do(u *server.Object) {
	nox_xxx_playerSetState_4FA020(u, server.PlayerState32)
	if u == nil {
		return
	}
	d := a.getHarpoonData(u)
	if d == nil {
		return
	}
	a.createBolt(u)
	d.frame35 = 0
}

func (a *abilityHarpoon) createBolt(u *server.Object) {
	if u == nil {
		return
	}
	d := a.getHarpoonData(u)
	if d == nil {
		return
	}
	bolt := a.s.NewObjectByTypeID("HarpoonBolt")
	if bolt == nil {
		return
	}
	r := u.Shape.Circle.R + 1.0
	(*server.HarpoonCollideData)(bolt.CollideData).Owner = u
	dv := u.Direction1.Vec()
	hpos := u.Pos().Add(dv.Mul(r))
	a.s.CreateObjectAt(bolt, u, hpos)
	bolt.VelVec = dv.Mul(bolt.SpeedCur)
	dir := u.Direction1
	bolt.Direction1 = dir
	bolt.Direction2 = dir
	d.bolt = bolt
	d.frame35 = 0
}

func (a *abilityHarpoon) netHarpoonAttach(u1, u2 *server.Object) {
	a.s.NetHarpoonAttach(u1, u2)
}

func (a *abilityHarpoon) netHarpoonBreak(u1 *server.Object, u2 *server.Object) {
	a.s.NetHarpoonBreak(u1, u2)
}

func (a *abilityHarpoon) updatePlayer4F8100(u *server.Object, ud *server.PlayerUpdateData) {
	playerUpdateHarpoon4F8100(playerUpdateHarpoonHooks4F8100[*server.Object]{
		loadTarget: func() *server.Object {
			return ud.HarpoonTarg
		},
		loadForce: func() float64 {
			return a.s.Balance.Float("HarpoonForce")
		},
		destroyed: func(target *server.Object) bool {
			// Direct access intentionally preserves the original fault when the
			// force callback clears a target that was non-nil on entry.
			return uint8(target.ObjFlags)&uint8(object.FlagDestroyed) != 0
		},
		breakOwner: func() {
			a.breakForOwner(u, true)
		},
		attribution: func(target *server.Object) {
			sub_4E7540(u, target)
		},
		applyForce: func(target *server.Object, force float64) {
			asObjectS(target).applyForce(u.PosVec, force)
		},
	})
}

func (a *abilityHarpoon) breakForOwner(u *server.Object, emitSound bool) {
	if u == nil {
		return
	}
	d := a.getHarpoonData(u)
	if d == nil {
		return
	}
	if d.bolt != nil {
		d.target = nil
		a.s.abilities.DisableAbility(u, server.AbilityHarpoon)
		asObjectS(d.bolt).Delete()
		d.bolt = nil
	}
	if emitSound {
		a.s.Audio.EventObj(sound.SoundHarpoonBroken, u, 0, 0)
	}
}

func (a *abilityHarpoon) Collide(bolt *server.Object, targ *server.Object, collision *types.Pointf) {
	a.s.HarpoonCollide4EB6A0(bolt, targ, collision, server.HarpoonCollideRuntime4EB6A0{
		LoadDamage:  func() int32 { return a.damage },
		StoreDamage: func(value int32) { a.damage = value },
		DamageMap: func(x, y, damage int32, typ object.DamageType, source *server.Object) {
			a.s.Nox_xxx_damageToMap_534BC0(int(x), int(y), int(damage), typ, source)
		},
		DisableAbility: a.s.abilities.DisableAbility,
		DelayedDelete:  a.s.DelayedDelete,
		MarkRelation: func(owner, target *server.Object) {
			sub_4E7540(owner, target)
		},
	})
}

func (a *abilityHarpoon) disable(u *server.Object) {
	ud := u.UpdateDataPlayer()
	a.netHarpoonBreak(u, ud.HarpoonBolt)
}

func (a *abilityHarpoon) Update(bolt *server.Object) {
	if bolt == nil || bolt.Owner() == nil {
		return
	}

	if a.maxDist == 0 {
		a.maxDist = float32(a.s.Balance.Float("MaxHarpoonDistance"))
		a.minDist = float32(a.s.Balance.Float("MinHarpoonDistance"))
		a.maxFlight = float32(a.s.Balance.Float("MaxHarpoonFlightDistance"))
		a.lifetime = float32(a.s.Balance.Float("MaxHarpoonExistence"))
	}
	owner := bolt.Owner()
	if owner.Flags().HasAny(object.FlagDestroyed | object.FlagDead) {
		a.breakForOwner(owner, true)
		return
	}
	bud := bolt.UpdateData
	obj4 := asObjectS(*(**server.Object)(bud))
	if obj4 != nil && obj4.Flags().HasAny(object.FlagDestroyed|object.FlagDead) {
		a.breakForOwner(owner, true)
		return
	}
	d := a.getHarpoonData(owner)
	if d == nil {
		return
	}
	if d.target == nil {
		if obj4 == nil {
			aim := d.getAim()
			obj6 := a.s.Nox_xxx_spellFlySearchTarget(&aim, bolt, 32, a.maxDist, 0, owner)
			*(**server.Object)(bud) = obj6
			if obj6 != nil {
				if legacy.Nox_server_testTwoPointsAndDirection_4E6E50(bolt.Pos(), int16(bolt.Direction1), obj6.Pos())&0x1 == 0 {
					*(**server.Object)(bud) = nil
				}
			}
		} else {
			vel := obj4.Pos().Sub(bolt.Pos())
			bolt.VelVec = vel.Normalize().Mul(bolt.SpeedCur)
		}
	}
	dist := nox_xxx_calcDistance_4E6C00(bolt, owner)
	if targ := asObjectS(d.target); targ != nil {
		if dist > a.maxDist {
			a.breakForOwner(owner, true)
			return
		}
		if dist < a.minDist {
			a.breakForOwner(owner, false)
			return
		}

		if df := a.s.Frame() - d.frame35; float32(df) > a.lifetime {
			a.breakForOwner(owner, true)
			return
		}
		tpos := targ.Pos()
		if a.s.Frame()-d.frame38 > 30 {
			d.frame38 = a.s.Frame()
			dx := d.targPos.X - tpos.X
			dy := d.targPos.Y - tpos.Y
			if dx*dx+dy*dy < 1.0 {
				a.breakForOwner(owner, true)
				return
			}
			d.targPos = tpos
		}
		if !a.s.MapTraceRayAt(owner.Pos(), tpos, nil, nil, 9) {
			a.breakForOwner(owner, true)
			return
		}
		if targ.Flags().HasAny(object.FlagDestroyed | object.FlagDead) {
			a.breakForOwner(owner, true)
			return
		}
		bolt.NewPos = tpos
		bolt.PosVec = tpos
		bolt.PrevPos = tpos
		bolt.VelVec = types.Pointf{}
		bolt.ForceVec = types.Pointf{}
		bolt.Direction1 = targ.Direction1
		a.s.nox_xxx_moveUpdateSpecial_517970(bolt)
	} else if dist > a.maxFlight {
		a.breakForOwner(owner, true)
		return
	}
	if d.frame35 == 0 {
		a.netHarpoonAttach(owner, bolt)
		d.frame35 = a.s.Frame()
	}
}
