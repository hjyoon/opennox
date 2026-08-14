package opennox

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func castGlyph(sp spell.ID, a2, caster, targ *server.Object, sa *server.SpellAcceptArg, lvl int) int {
	s := noxServer
	if !caster.Class().Has(object.ClassPlayer) {
		return 1
	}
	ud := caster.UpdateDataPlayer()
	pl := ud.Player
	if int(ud.CurTraps) >= int(s.Balance.Float("MaxTrapCount")) {
		s.NetInformTextMsg(pl.PlayerIndex(), 0, 5)
		return 0
	}
	if ud.TrapSpellsCnt == 0 {
		s.NetInformTextMsg(pl.PlayerIndex(), 0, 7)
		return 0
	}
	trap := s.NewObjectByTypeID("Glyph")
	if trap == nil {
		return 0
	}
	if pl.PlayerClass() != player.Conjurer {
		s.CreateObjectAt(trap, caster, targ.Pos())
		snd := s.Spells.DefByInd(sp).GetCastSound()
		s.Audio.EventObj(snd, targ, 0, 0)
	} else {
		if countBombers(caster) >= int(s.Balance.Float("MaxBomberCount")) {
			s.NetInformTextMsg(pl.PlayerIndex(), 0, 5)
			s.DelayedDelete(trap)
			return 0
		}
		if !nox_xxx_checkSummonedCreaturesLimit_500D70(caster, 5) {
			s.DelayedDelete(trap)
			return 0
		}
		pos := s.RandomReachablePointAround(50.0, targ.Pos())
		var bomb *server.Object
		if caster != nil {
			bomb = nox_xxx_unitDoSummonAt_5016C0(s.Types.BomberID(), pos, caster, caster.Direction1)
		} else {
			bomb = nox_xxx_unitDoSummonAt_5016C0(s.Types.BomberID(), pos, nil, 0)
		}
		if bomb != nil {
			legacy.Nox_xxx_inventoryPutImpl_4F3070(bomb, trap, 1)
		}
		s.Audio.EventObj(sound.SoundBomberSummon, targ, 0, 0)
	}
	idata := trap.InitDataGlyph()
	*idata = server.GlyphInitData{
		SpellsCnt: 0,
		SpellArg: server.SpellAcceptArg{
			Pos: sa.Pos,
		},
	}
	for i := 0; i < int(ud.TrapSpellsCnt); i++ {
		tsp := spell.ID(ud.TrapSpells[i])
		if spl := s.Spells.DefByInd(tsp); spl != nil && spl.Enabled {
			idata.Spells[idata.SpellsCnt] = uint32(tsp)
			idata.SpellsCnt++
		}
	}
	ud.TrapSpellsCnt = 0
	trap.Direction1 = targ.Direction1
	trap.Direction2 = targ.Direction1
	if pl.PlayerClass() == player.Wizard {
		ud.CurTraps++
	}
	return 1
}

func setBomberSpells(u *server.Object, spells ...spell.ID) {
	s := noxServer
	if u == nil {
		return
	}
	if !u.Class().Has(object.ClassMonster) {
		return
	}
	if !u.SubClass().AsMonster().Has(object.MonsterBomber) {
		return
	}
	for it := u.FirstItem(); it != nil; it = it.NextItem() {
		if int(it.TypeInd) != s.Types.GlyphID() {
			s.DelayedDelete(it)
			break
		}
	}
	trap := s.NewObjectByTypeID("Glyph")
	if trap == nil {
		return
	}
	idata := trap.InitDataGlyph()
	*idata = server.GlyphInitData{
		SpellsCnt: 0,
		SpellArg: server.SpellAcceptArg{
			Pos: u.Pos(),
		},
	}
	for _, sp := range spells {
		if !sp.Valid() {
			continue
		}
		idata.Spells[idata.SpellsCnt] = uint32(sp)
		idata.SpellsCnt++
	}
	legacy.Nox_xxx_inventoryPutImpl_4F3070(u, trap, 1)
}

func countBombers(u *server.Object) int {
	cnt := 0
	for it := u.FirstOwned516(); it != nil; it = it.NextOwned512() {
		if it.Class().Has(object.ClassMonster) && it.SubClass().AsMonster().Has(object.MonsterBomber) {
			cnt++
		}
	}
	return cnt
}

func nox_xxx___mkgmtime_538280(obj *server.Object) {
	triggerTrap(obj, nil)
}

func nox_xxx_dieGlyph_54DF30(obj *server.Object) {
	triggerTrap(obj, nil)
}

func nox_xxx_collideGlyph_4E9A00(a1, a2 *server.Object, collision unsafe.Pointer) {
	noxServer.GlyphCollide4E9A00(a1, a2, collision, server.GlyphCollideRuntime4E9A00{
		Trigger: triggerTrap,
	})
}

func castDetonateGlyphs(sp spell.ID, a2, a3, caster *server.Object, sa *server.SpellAcceptArg, lvl int) int {
	s := noxServer
	pos := caster.Pos()
	const dist = 300
	rect := types.RectFromPointsf(pos.Sub(types.Ptf(dist, dist)), pos.Add(types.Ptf(dist, dist)))
	snd := s.Spells.DefByInd(sp).GetCastSound()
	s.Audio.EventObj(snd, a3, 0, 0)
	for {
		var found *server.Object
		s.Map.EachObjInRect(rect, func(it *server.Object) bool {
			if int(it.TypeInd) != s.Types.GlyphID() || it.Flags().Has(object.FlagDestroyed) {
				return true
			}
			owner := caster.FindOwnerChainPlayer()
			if it.HasOwner(owner) || it.ObjOwner == nil && !owner.Class().Has(object.ClassPlayer) {
				if s.MapTraceRayAt(caster.Pos(), it.Pos(), nil, nil, 5) {
					found = it
					return false
				}
			}
			return true
		})
		if found == nil {
			break
		}
		triggerTrap(found, nil)
	}
	return 1
}

func triggerTrap(trap, a2 *server.Object) {
	s := noxServer
	idata := trap.InitDataGlyph()
	if trap.Flags().Has(object.FlagDestroyed) {
		return
	}
	owner := trap.FindOwnerChainPlayer()
	s.DelayedDelete(trap)
	if owner != nil && owner.Class().Has(object.ClassPlayer) {
		ud := owner.UpdateDataPlayer()
		if ud.Player.PlayerClass() == player.Wizard {
			if ud.CurTraps != 0 {
				ud.CurTraps--
			}
		}
	}
	sa := &idata.SpellArg
	if a2 != nil {
		idata.SpellArg.Obj = a2
	} else {
		v10 := float32(s.Balance.Float("GlyphRange"))
		idata.SpellArg.Obj = sub_4E6EA0(trap, v10, &trapSearchArg{
			Field0:             15,
			Field4:             1,
			Field8:             0,
			ClassAllow12:       object.MaskUnits,
			ClassDisallow16:    0,
			SubClassAllow20:    math.MaxUint32,
			SubClassDisallow24: 0,
			FlagsAllow28:       math.MaxUint32,
			FlagsDisallow32:    object.FlagDead,
		})
	}

	for i := 0; i < int(idata.SpellsCnt); i++ {
		sp := spell.ID(idata.Spells[i])
		if (!s.Spells.HasFlags(sp, things.SpellFlagUnk1) || a2 != nil) && legacy.Sub_4FD0E0(owner, sp) == 0 {
			if owner.Class().Has(object.ClassPlayer) {
				lvl := legacy.Nox_xxx_spellGetPower_4FE7B0(sp, owner)
				s.Nox_xxx_spellAccept4FD400(sp, owner, owner, trap, sa, lvl)
			} else {
				s.Nox_xxx_spellAccept4FD400(sp, owner, owner, trap, sa, 2)
			}
		}
	}
	pos := trap.Pos()
	s.Nox_xxx_netSendPointFx_522FF0(netmsg.MSG_FX_BLUE_SPARKS, pos)
	s.Audio.EventPos(sound.SoundGlyphDetonate, pos, 0, 0)
	const dist = 100
	rect := types.RectFromPointsf(pos.Sub(types.Ptf(dist, dist)), pos.Add(types.Ptf(dist, dist)))
	s.Map.EachObjInRect(rect, func(it *server.Object) bool {
		if it != trap && (int32(uint8(*(*float32)(unsafe.Add(unsafe.Pointer(it), unsafe.Sizeof(float32(0))*4))))&0x20) == 0 {
			if int(it.TypeInd) == s.Types.GlyphID() {
				if s.MapTraceRayAt(trap.Pos(), it.Pos(), nil, nil, 5) {
					_ = nox_xxx___mkgmtime_538280
					it.Update = legacy.Get_nox_xxx___mkgmtime_538280()
					s.Objs.AddToUpdatable(it)
				}
			}
		}
		return true
	})
}

type trapSearchArg = targetSearchArg4E6EA0[*server.Object]

func sub_4E6EA0(a1 *server.Object, r float32, ta *trapSearchArg) *server.Object {
	s := noxServer
	return targetSearch4E6EA0(a1, r, ta, targetSearch4E6EA0Hooks[*server.Object]{
		eachInCircle: s.Map.EachObjInCircle,
		class:        func(it *server.Object) object.Class { return it.ObjClass },
		subClass:     func(it *server.Object) object.SubClass { return it.ObjSubClass },
		flags:        func(it *server.Object) object.Flags { return it.ObjFlags },
		position:     func(it *server.Object) types.Pointf { return it.PosVec },
		directionInd: func(it *server.Object) int16 { return int16(it.Direction1) },
		sameTeam:     legacy.Nox_xxx_unitsHaveSameTeam_4EC520,
		playerStatus: func(it *server.Object) uint32 {
			return it.ControllingPlayer().Field3680
		},
		isEnemy: s.IsEnemyTo,
		direction: func(a types.Pointf, dir int16, b types.Pointf) uint32 {
			return uint32(legacy.Nox_server_testTwoPointsAndDirection_4E6E50(a, dir, b))
		},
		canInteract: s.CanInteract,
	})
}

func sub_4E71F0(obj *server.Object) {
	s := noxServer
	spellProjectileExpireObject4E71F0(obj, sub_4E6EA0, s.Nox_xxx_spellAccept4FD400, s.DelayedDelete)
}

func nox_bomberDead_54A150(u *server.Object) int {
	s := noxServer
	ud := u.UpdateDataMonster()
	s.Nox_xxx_netSendPointFx_522FF0(netmsg.MSG_FX_EXPLOSION, u.Pos())
	s.Audio.EventObj(sound.SoundBomberDie, u, 0, 0)

	if it := u.FirstItem(); it != nil {
		// TODO: this assumes inventory of exactly 1 item, which is a trap
		//       instead, should scan for all traps, check the type of it, etc
		idata := it.InitDataGlyph()
		owner := it.ObjOwner
		legacy.Sub_4ED0C0(u, it)
		s.ObjSetOwner(owner, it)
		idata.SpellArg.Obj = nil
		idata.SpellArg.Pos = u.Pos()
		it.PosVec = u.PosVec
		it.NewPos = u.NewPos
		it.Direction1 = u.Direction1
		it.Direction2 = u.Direction1
		triggerTrap(it, ud.BombCollideTarget)
	} else {
		s.Nox_xxx_mapDamageUnitsAround(u.Pos(), 50.0, 30.0, 10, object.DamageExplosion, u, nil, doDamageWalls)
		legacy.Nox_xxx_mapPushUnitsAround_52E040(u.Pos(), 50.0, 30.0, 30.0, u, 0, 0)
	}
	return 1
}
