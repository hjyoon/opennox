package opennox

import (
	"fmt"
	"image"
	"math"
	"unsafe"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	nsp "github.com/opennox/noxscript/ns/v4/spell"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func e2eMagicMissileUnitMode(level int, direction string) (fromNPC, ok bool) {
	fromNPC, ok = e2eFistUnitDirection(direction)
	return fromNPC, ok && (level >= 1 && level <= 5 || fromNPC && level == 0)
}

// Independent expected-value calculation for case 7's binary32 armor/carry
// spill, round-to-even and positive minimum, then DefaultDamage's fire minimum.
func e2eMagicMissileDamage(raw int32, armor, carry float32) (int32, float32) {
	accumulated := float32((1-float64(armor))*float64(raw)) + carry
	effective := int32(math.RoundToEven(float64(accumulated)))
	next := accumulated - float32(effective)
	if raw > 0 && effective == 0 {
		effective = 1
	}
	if effective == 0 {
		effective = 1
	}
	return effective, next
}

func e2eMagicMissileSplash(raw int32, radius float32, from, to types.Pointf) (int32, bool) {
	dx, dy := to.X-from.X, to.Y-from.Y
	distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	if radius <= 5 || distance > radius {
		return 0, false
	}
	amount := float32(raw)
	if distance >= 5 {
		amount *= 1 - (distance-5)/(radius-5)
	}
	return int32(amount), true
}

// Deleted pointers are retained only as identity keys, never dereferenced.
type e2eMagicMissileIdentity struct {
	wire   uint32
	script int32
	hit    bool
}

type e2eMagicMissileUnitFixture struct {
	level, actualLevel, count, castAudio, hits              int
	direction, projectileType                               string
	fromNPC, active, natural                                bool
	host, npc, caster, target                               *server.Object
	original                                                types.Pointf
	hostHP, hostMax, health, lastHP, reportHP, reportBefore uint16
	frame                                                   uint32
	armor, carry, radius                                    float32
	direct, splash, totalDamage                             int32
	projectiles                                             map[*server.Object]e2eMagicMissileIdentity
}

func (f *e2eMagicMissileUnitFixture) prepare() {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.HealthData == nil || f.host.HealthData.Cur == 0 || f.host.ControllingPlayer() == nil {
		e2eError(fmt.Errorf("Magic Missile requires a live host"))
		return
	}
	f.original, f.hostHP, f.hostMax = f.host.PosVec, f.host.HealthData.Cur, f.host.HealthData.Max
	origin, direction, err := e2eWarriorAbilityArena(f.original, f.host.Shape.Circle.R+24, func(from, to types.Pointf) bool {
		return e2eWarriorLaneClear(f.host, from, to)
	})
	if err != nil {
		e2eError(err)
		return
	}
	f.npc = noxServer.NewObjectByTypeID("NPC")
	if f.npc == nil {
		e2eError(fmt.Errorf("Magic Missile requires the stock NPC"))
		return
	}
	asObjectS(f.host).SetPos(origin)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	noxServer.CreateObjectAt(f.npc, nil, origin.Add(direction.Mul(96)))
	noxServer.ObjectsAddPending()
	if !e2eObjectInWorld(f.npc) || f.npc.UpdateData == nil || f.npc.HealthData == nil ||
		f.npc.Damage != f.host.Damage || !f.npc.SubClass().AsMonster().Has(object.MonsterNPC) {
		e2eError(fmt.Errorf("Magic Missile stock NPC lacks the real PlayerDamage callback"))
		return
	}
	// Placement, durable HP and waiting AI are setup; hit results are untouched.
	asObjectS(f.npc).SetMaxHealth(2000)
	f.npc.UpdateDataMonster().SetAggression(0)
	f.npc.ClearActionStack()
	f.npc.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
	f.caster, f.target = f.host, f.npc
	if f.fromNPC {
		f.caster, f.target = f.npc, f.host
		asObjectS(f.host).SetMaxHealth(2000)
	}
	if f.target.Buffs != 0 || f.caster.Buffs != 0 {
		e2eError(fmt.Errorf("Magic Missile baseline units are enchanted"))
		return
	}
	for item := f.target.FirstItem(); item != nil; item = item.NextItem() {
		if !item.Flags().Has(object.FlagEquipped) || item.InitData == nil ||
			!item.Class().HasAny(object.ClassArmor|object.ClassWeapon|object.ClassWand) {
			continue
		}
		for _, m := range item.InitDataModifier().Modifiers {
			if m != nil && (m.Defend76.Fnc != nil || m.Engage112 != nil) {
				e2eError(fmt.Errorf("Magic Missile baseline has an equipped protection modifier"))
				return
			}
		}
	}
	f.actualLevel = f.level
	if f.level == 0 {
		f.actualLevel = int(noxServer.Server.SpellPower4FE7B0(spell.SPELL_MAGIC_MISSILE, f.caster))
	}
	spl := noxServer.Spells.DefByInd(spell.SPELL_MAGIC_MISSILE)
	if spl == nil || f.actualLevel < 1 || f.actualLevel > 5 {
		e2eError(fmt.Errorf("Magic Missile definition/selected level unavailable: %d", f.actualLevel))
		return
	}
	opts := spl.Def.Missiles.Level(f.actualLevel)
	f.projectileType, f.count = opts.Projectile, opts.Count
	if f.count <= 0 {
		f.count = int(noxServer.Balance.FloatInd("MagicMissileCount", f.actualLevel-1))
	}
	f.direct = int32(math.RoundToEven(float64(float32(noxServer.Balance.Float("MagicMissileDamage")))))
	f.splash = int32(math.RoundToEven(float64(float32(noxServer.Balance.Float("MagicMissileSplashDamage")))))
	f.radius = float32(noxServer.Balance.Float("MagicMissileRange"))
	if f.count <= 0 || f.count > 20 || f.direct <= 0 || f.splash <= 0 || f.radius <= 5 {
		e2eError(fmt.Errorf("Magic Missile invalid stock count/damage/range: %d/%d/%d/%g", f.count, f.direct, f.splash, f.radius))
		return
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, p := range []unsafe.Pointer{unsafe.Pointer(f.host), unsafe.Pointer(f.npc), f.host.UpdateData, f.npc.UpdateData} {
			if uintptr(p) <= math.MaxUint32 {
				e2eError(fmt.Errorf("Magic Missile unit pointer is below 4 GiB: %p", p))
				return
			}
		}
	}
	f.projectiles = make(map[*server.Object]e2eMagicMissileIdentity)
	noxServer.Audio.OnSound(func(id sound.ID, kind int, owner *server.Object, _ types.Pointf) {
		if !f.active {
			return
		}
		if id == spl.GetCastSound() && owner == f.caster {
			f.castAudio++
			if kind != 0 || f.castAudio != 1 {
				e2eError(fmt.Errorf("Magic Missile cast sound kind/count=%d/%d", kind, f.castAudio))
				return
			}
			if len(f.projectiles) == 0 {
				f.observeProjectiles()
			}
			if f.level == 0 {
				ud, head := f.npc.UpdateDataMonster(), f.npc.UpdateDataMonster().AIStackHead()
				if head == nil || head.Type() != ai.ACTION_CAST_SPELL_ON_OBJECT || head.ArgObj(2) != f.target ||
					head.ArgU32(0) != uint32(spell.SPELL_MAGIC_MISSILE) || ud.Field120_2 != 0 ||
					uint32(ud.Field120_1) != ud.MonsterDef.MissileAttackFrame216 {
					e2eError(fmt.Errorf("Magic Missile NPC bypassed the natural animation/cast frame"))
					return
				}
				f.natural = true
			}
		}
		if id == sound.SoundMagicMissileDetonate {
			if _, known := f.projectiles[owner]; known {
				f.observeDetonation(owner, kind)
			}
		}
	})
	noxServer.TickHook(f.observeReports)
}

func (f *e2eMagicMissileUnitFixture) beginCast() {
	if !e2eObjectInWorld(f.target) || !noxServer.MapTraceRay(f.caster.PosVec, f.target.PosVec, server.MapTraceFlag1) ||
		noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target))) == nil {
		e2eError(fmt.Errorf("Magic Missile target placement/publication did not settle"))
		return
	}
	f.health, f.lastHP, f.frame = f.target.HealthData.Cur, f.target.HealthData.Cur, noxServer.Frame()
	f.armor, f.carry = f.readArmorCarry()
	sample := unitHealthSampleNative4D8760(f.target, int(f.host.ControllingPlayer().PlayerIndex()))
	if sample == nil || *sample != f.health {
		e2eError(fmt.Errorf("Magic Missile real recipient HP cache did not settle"))
		return
	}
	f.reportHP = *sample
	f.caster.SetDir(server.DirFromVec(f.target.PosVec.Sub(f.caster.PosVec)))
	opts := noxServer.Spells.DefByInd(spell.SPELL_MAGIC_MISSILE).Def.Missiles.Level(f.actualLevel)
	for i := 0; i < f.count; i++ {
		doff := int16(opts.Spread * uint16((i+1)/2))
		if i%2 == 1 {
			doff = -doff
		}
		dir := server.RoundDir(int(int16(f.caster.Direction1) + doff))
		spawn := f.caster.PosVec.Add(dir.Vec().Mul(f.caster.Shape.Circle.R + opts.Offset))
		if !noxServer.MapTraceRay(f.caster.PosVec, spawn, server.MapTraceFlag1|server.MapTraceFlag3) {
			e2eError(fmt.Errorf("Magic Missile spread ray %d is blocked", i))
			return
		}
	}
	if !f.fromNPC {
		// Ordinary cursor/aim setup, not a MissileUpdateData.Target override.
		ud := f.host.UpdateDataPlayer()
		ud.CursorObj, ud.Player.Obj3640 = f.target, f.target
		ud.Field55, ud.Field56 = int(f.target.PosVec.X), int(f.target.PosVec.Y)
		ud.Player.CursorVec = image.Pt(int(f.target.PosVec.X), int(f.target.PosVec.Y))
		mouse := noxClient.Viewport().ToScreenPos(ud.Player.CursorVec)
		noxClient.ChangeMousePos(mouse, true)
		e2eQueueInput(&seat.MouseMoveEvent{Pos: mouse, Relative: false})
	}
	f.active = true
	if f.level == 0 {
		f.npc.MonsterPushAction(ai.ACTION_CAST_SPELL_ON_OBJECT, uint32(spell.SPELL_MAGIC_MISSILE), 0, f.target)
		return
	}
	api := noxServer.noxScriptP()
	api.CastSpellLvl(nsp.Spell("SPELL_MAGIC_MISSILE"), f.level, api.toObj(f.caster), api.toObj(f.target))
	noxServer.ObjectsAddPending()
	// Calls outside the server audio tick enqueue their real sound until the
	// next tick. Observe actual objects now, and require the sound in complete.
	if len(f.projectiles) == 0 {
		f.observeProjectiles()
	}
}

func (f *e2eMagicMissileUnitFixture) readArmorCarry() (float32, float32) {
	if f.fromNPC {
		ud := f.target.UpdateDataPlayer()
		return math.Float32frombits(ud.Field57), math.Float32frombits(ud.Field21)
	}
	ud := f.target.UpdateDataMonster()
	return math.Float32frombits(ud.Field518), math.Float32frombits(ud.Field1)
}

func (f *e2eMagicMissileUnitFixture) observeProjectiles() {
	for obj := f.caster.Field129; obj != nil; obj = obj.Field128 {
		if obj.ObjectTypeC().ID() != f.projectileType || obj.Flags().Has(object.FlagDestroyed) {
			continue
		}
		if obj.ObjOwner != f.caster || obj.UpdateData == nil || obj.Collide == nil || !obj.Class().Has(object.ClassMissile) ||
			unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 || uintptr(obj.UpdateData) <= math.MaxUint32) {
			e2eError(fmt.Errorf("Magic Missile lost native missile/owner/update/collision"))
			return
		}
		mud := obj.UpdateDataMissile()
		if mud.Owner != f.caster || mud.Target != f.target || mud.SpellID != int32(spell.SPELL_MAGIC_MISSILE) {
			e2eError(fmt.Errorf("Magic Missile real target search mismatch: owner=%p target=%p want=%p spell=%d", mud.Owner, mud.Target, f.target, mud.SpellID))
			return
		}
		if _, duplicate := f.projectiles[obj]; duplicate {
			e2eError(fmt.Errorf("Magic Missile duplicate projectile"))
			return
		}
		f.projectiles[obj] = e2eMagicMissileIdentity{wire: obj.NetCode, script: obj.ScriptIDVal}
	}
	if len(f.projectiles) != f.count {
		e2eError(fmt.Errorf("Magic Missile projectile count=%d want=%d", len(f.projectiles), f.count))
		return
	}
	e2eLog.Printf("MAGIC MISSILE CAST: direction=%s requested=%d actual=%d missiles=%d direct=%d splash=%d range=%g armor=%g carry=%g",
		f.direction, f.level, f.actualLevel, f.count, f.direct, f.splash, f.radius, f.armor, f.carry)
}

func (f *e2eMagicMissileUnitFixture) observeDetonation(missile *server.Object, kind int) {
	id := f.projectiles[missile]
	rawSplash, inRange := e2eMagicMissileSplash(f.splash, f.radius, missile.PosVec, f.target.PosVec)
	armor, carry := f.readArmorCarry()
	direct, nextCarry := e2eMagicMissileDamage(f.direct, f.armor, f.carry)
	// A direct collision can touch the NPC radius outside the 15-unit splash
	// circle. MapDamageUnits also requires its separate line-of-sight check.
	splashApplied := inRange && noxServer.MapTraceRay(missile.PosVec, f.target.PosVec, server.MapTraceFlag1)
	splash := int32(0)
	if splashApplied {
		splash, nextCarry = e2eMagicMissileDamage(rawSplash, f.armor, nextCarry)
	}
	damage := direct + splash
	wantMarker, wantType := uint32(1), uint32(missile.TypeInd)
	if splashApplied {
		wantMarker, wantType = 2, uint32(object.DamageExplosion)
	}
	var marker, markerType uint32
	if f.fromNPC {
		ud := f.target.UpdateDataPlayer()
		marker, markerType = ud.Field76, ud.Field75
	} else {
		ud := f.target.UpdateDataMonster()
		marker, markerType = ud.Field547, ud.Field546
		if !ud.StatusFlags.Has(object.MonStatusInjured) || !ud.StatusFlags.Has(object.MonStatusOnFire) {
			e2eError(fmt.Errorf("Magic Missile NPC lacks injured/on-fire status"))
			return
		}
	}
	if kind != 0 || id.hit || noxServer.Frame()-missile.Field32 > 3*noxServer.TickRate() ||
		missile.UpdateDataMissile().Target != f.target || missile.UpdateDataMissile().Owner != f.caster ||
		f.lastHP <= uint16(damage) || f.target.HealthData.Cur != f.lastHP-uint16(damage) ||
		armor != f.armor || math.Float32bits(carry) != math.Float32bits(nextCarry) ||
		f.target.Obj130 != missile || f.target.Field131 != uint32(object.DamageExplosion) ||
		f.target.Frame134 != noxServer.Frame() || marker != wantMarker || markerType != wantType {
		e2eError(fmt.Errorf("Magic Missile direct+splash mismatch: direction=%s level=%d HP=%d->%d want=%d direct=%d splash=%d marker=%d/%d armor=%g/%g carry=%#x/%#x",
			f.direction, f.level, f.lastHP, f.target.HealthData.Cur, f.lastHP-uint16(damage), direct, splash,
			marker, markerType, armor, f.armor, math.Float32bits(carry), math.Float32bits(nextCarry)))
		return
	}
	id.hit = true
	f.projectiles[missile] = id
	f.hits++
	f.totalDamage += damage
	f.lastHP, f.carry = f.target.HealthData.Cur, nextCarry
	e2eLog.Printf("MAGIC MISSILE HIT: direction=%s requested=%d actual=%d hit=%d/%d direct=%d splash=%d HP=%d carry=%g",
		f.direction, f.level, f.actualLevel, f.hits, f.count, direct, splash, f.lastHP, f.carry)
}

func (f *e2eMagicMissileUnitFixture) observeReports() {
	if !f.active {
		return
	}
	sample := unitHealthSampleNative4D8760(f.target, int(f.host.ControllingPlayer().PlayerIndex()))
	if sample == nil {
		e2eError(fmt.Errorf("Magic Missile lost recipient HP cache"))
		return
	}
	if *sample != f.reportHP {
		if *sample >= f.reportHP || *sample < f.lastHP {
			e2eError(fmt.Errorf("Magic Missile unexpected cached HP transition: %d->%d actual=%d", f.reportHP, *sample, f.lastHP))
			return
		}
		f.reportBefore, f.reportHP = f.reportHP, *sample
	}
}

func (f *e2eMagicMissileUnitFixture) complete() bool {
	if f.hits != f.count || f.castAudio != 1 || f.health-f.lastHP != uint16(f.totalDamage) ||
		f.reportHP != f.lastHP || f.target.HealthData.Cur != f.lastHP ||
		f.level == 0 && (!f.natural || f.npc.MonsterActionIsScheduled(ai.ACTION_CAST_SPELL_ON_OBJECT)) {
		return false
	}
	for obj, id := range f.projectiles {
		if !id.hit || e2eFistInWorld(obj, id.wire, id.script) || noxClient.Objs.ByNetCode(uint16(id.wire)) != nil {
			return false
		}
	}
	dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target)))
	if dr == nil {
		return false
	}
	delta, ready := legacy.HealthChangeForDrawable(dr.NetCode32)
	if !ready || delta != int16(f.reportHP-f.reportBefore) || delta >= 0 {
		return false
	}
	if f.fromNPC {
		hud, ready := e2eClientHUDMeter(0)
		if !ready || hud.Current != uint32(f.lastHP) || hud.Maximum != uint32(f.hostMax) {
			return false
		}
	}
	e2eLog.Printf("MAGIC MISSILE COMPLETE: direction=%s requested=%d actual=%d hits=%d/%d total-damage=%d server/client-HP=verified missiles=removed natural-NPC=%t",
		f.direction, f.level, f.actualLevel, f.hits, f.count, f.totalDamage, f.natural)
	return true
}

func (f *e2eMagicMissileUnitFixture) cleanup() {
	f.active = false
	noxServer.DelayedDelete(f.npc)
	asObjectS(f.host).SetPos(f.original)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	if f.fromNPC {
		asObjectS(f.host).SetMaxHealth(int(f.hostMax))
		asObjectS(f.host).SetHealth(int(f.hostHP))
	}
}

// Script requests exercise real creation, target search, homing, collision,
// direct AND splash damage, armor/carry, sound, client reports and deletion.
// No hit result is supplied. Player mana/incantation input is not simulated.
func (sc *e2eScenario) CheckMagicMissileUnitDamage(level int, direction, name string) {
	fromNPC, ok := e2eMagicMissileUnitMode(level, direction)
	if !ok {
		e2eError(fmt.Errorf("invalid Magic Missile fixture: %s/%d", direction, level))
		return
	}
	f := &e2eMagicMissileUnitFixture{level: level, direction: direction, fromNPC: fromNPC}
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		host := noxServer.Players.HostUnit()
		return host != nil && host.Buffs == 0 && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, f.prepare)
	sc.Wait(12, name+" publish units")
	sc.add(0, name+" cast", f.beginCast)
	sc.addWhen(1, name+" actual hits and client replay", 300, f.complete, f.cleanup)
	sc.Wait(3, name+" cleanup")
}
