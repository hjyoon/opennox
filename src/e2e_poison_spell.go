package opennox

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	nsp "github.com/opennox/noxscript/ns/v4/spell"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func e2ePoisonMode(level int, direction string) (fromNPC, ok bool) {
	switch direction {
	case "player-to-npc":
		return false, level >= 1 && level <= 5
	case "npc-to-player":
		return true, level == 0
	default:
		return false, false
	}
}

// For a nonzero application timestamp, the live damage loop waits more than
// sixty frames, then tests frame modulo 128 >> (poison-1). Unsigned wrap is
// retained at both additions. This fixture cannot start on frame zero.
func e2ePoisonFirstTick(applied uint32, power uint8) (uint32, bool) {
	if applied == 0 || power == 0 {
		return 0, false
	}
	period := uint32(1)
	if power <= 8 {
		period = uint32(128 >> (power - 1))
	}
	next := applied + 61
	if remainder := next % period; remainder != 0 {
		next += period - remainder
	}
	return next, true
}

type e2ePoisonFixture struct {
	level                                             int
	direction                                         string
	fromNPC, active, effectSeen, naturalSeen, hitSeen bool
	host, npc, caster, target, magic                  *server.Object
	original                                          types.Pointf
	magicWire                                         uint32
	magicScript                                       int32
	health                                            uint16
	frame, applied, firstTick                         uint32
	power                                             uint8
	castAudio, effectAudio                            int
}

func (f *e2ePoisonFixture) prepare() {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.HealthData == nil || f.host.HealthData.Cur <= 1 ||
		f.host.Poison540 != 0 || f.host.Flags().HasAny(object.FlagDead|object.FlagDestroyed|object.FlagNoUpdate) {
		e2eError(fmt.Errorf("Poison requires a live, unpoisoned host"))
		return
	}
	f.original = f.host.PosVec
	origin, direction, err := e2eWarriorAbilityArena(f.original, f.host.Shape.Circle.R+16, func(from, to types.Pointf) bool {
		return e2eWarriorLaneClear(f.host, from, to)
	})
	if err != nil {
		e2eError(err)
		return
	}
	f.npc = noxServer.NewObjectByTypeID("NPC")
	if f.npc == nil {
		e2eError(fmt.Errorf("Poison has no stock NPC type"))
		return
	}
	hostPos, npcPos := origin, origin.Add(direction.Mul(112))
	f.caster, f.target = f.host, f.npc
	if f.fromNPC {
		hostPos, npcPos = npcPos, hostPos
		f.caster, f.target = f.npc, f.host
	}
	asObjectS(f.host).SetPos(hostPos)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	noxServer.CreateObjectAt(f.npc, nil, npcPos)
	noxServer.ObjectsAddPending()
	if !e2eObjectInWorld(f.npc) || f.npc.UpdateData == nil || f.npc.HealthData == nil ||
		f.npc.HealthData.Cur <= 1 || f.npc.UpdateDataMonster().MonsterDef == nil ||
		!f.npc.SubClass().AsMonster().Has(object.MonsterNPC) || f.target.Damage == nil {
		e2eError(fmt.Errorf("Poison stock NPC/damage callback was not initialized"))
		return
	}
	f.npc.UpdateDataMonster().SetAggression(0)
	f.npc.ClearActionStack()
	f.npc.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
	if !noxServer.Spells.HasFlags(spell.SPELL_POISON, things.SpellTargeted) {
		e2eError(fmt.Errorf("stock Poison is not targeted"))
		return
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(f.host), unsafe.Pointer(f.npc), f.host.UpdateData, f.npc.UpdateData, unsafe.Pointer(f.target.HealthData)} {
			if uintptr(ptr) <= math.MaxUint32 {
				e2eError(fmt.Errorf("Poison native fixture pointer is below 4 GiB: %p", ptr))
				return
			}
		}
	}
	noxServer.Audio.OnSound(func(id sound.ID, kind int, owner *server.Object, _ types.Pointf) {
		if !f.active {
			return
		}
		if id == sound.SoundPoisonCast && owner == f.caster {
			f.castAudio++
			if kind != 0 || f.castAudio != 1 {
				e2eError(fmt.Errorf("Poison cast sound kind/count=%d/%d", kind, f.castAudio))
				return
			}
			f.observeProjectile()
			if f.fromNPC {
				f.observeNaturalCast()
			}
		}
		if id == sound.SoundPoisonEffect && owner == f.target {
			f.effectAudio++
			if kind != 0 || f.effectAudio != 1 {
				e2eError(fmt.Errorf("Poison effect sound kind/count=%d/%d", kind, f.effectAudio))
				return
			}
			f.observeEffect()
		}
	})
	// Observe the server result before another AI update consumes its injured
	// latch. Client packet replay and rendering can complete on later frames.
	noxServer.TickHook(f.observeDOT)
	e2eLog.Printf("POISON PREPARED: direction=%s requested-level=%d caster=%p target=%p update=%p damage=%p", f.direction, f.level, f.caster, f.target, f.target.UpdateData, f.target.Damage)
}

func (f *e2ePoisonFixture) beginCast() {
	if !e2eObjectInWorld(f.target) || !e2eObjectInWorld(f.caster) || f.target.Poison540 != 0 ||
		f.target.HasEnchant(server.ENCHANT_INVULNERABLE) || !noxServer.MapTraceRay(f.caster.PosVec, f.target.PosVec, 0) ||
		noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target))) == nil {
		e2eError(fmt.Errorf("Poison target placement/protection/client publication did not settle"))
		return
	}
	f.health, f.frame, f.active = f.target.HealthData.Cur, noxServer.Frame(), true
	if f.fromNPC {
		f.npc.MonsterPushAction(ai.ACTION_CAST_SPELL_ON_OBJECT, uint32(spell.SPELL_POISON), 0, f.target)
		return
	}
	api := noxServer.noxScriptP()
	api.CastSpellLvl(nsp.Spell("SPELL_POISON"), f.level, api.toObj(f.caster), api.toObj(f.target))
}

func (f *e2ePoisonFixture) observeProjectile() {
	for obj := f.caster.Field129; obj != nil; obj = obj.Field128 {
		if obj.ObjectTypeC().ID() != "Magic" || obj.Flags().Has(object.FlagDestroyed) {
			continue
		}
		ud := obj.UpdateDataSpellProjectile()
		if ud == nil || ud.Spell12 != uint32(spell.SPELL_POISON) {
			continue
		}
		if f.magic != nil || ud.Target != f.target || ud.Field0 != f.caster || ud.Field8 != f.caster ||
			ud.Level16 != uint32(noxServer.Server.SpellPower4FE7B0(spell.SPELL_POISON, f.caster)) ||
			ud.Level16 == 0 || ud.Level16 > 255 || unsafe.Sizeof(uintptr(0)) == 8 &&
			(uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 || uintptr(obj.UpdateData) <= math.MaxUint32) {
			e2eError(fmt.Errorf("Poison projectile lost native pointers or mode-selected power"))
			return
		}
		f.magic, f.magicWire, f.magicScript, f.power = obj, obj.NetCode, obj.ScriptIDVal, uint8(ud.Level16)
		e2eLog.Printf("POISON PROJECTILE: object=%p update=%p target=%p requested-level=%d actual-power=%d", obj, obj.UpdateData, ud.Target, f.level, f.power)
	}
	if f.magic == nil {
		e2eError(fmt.Errorf("Poison cast has no real Magic projectile"))
	}
}

func (f *e2ePoisonFixture) observeNaturalCast() {
	ud := f.npc.UpdateDataMonster()
	head := ud.AIStackHead()
	if f.naturalSeen || head == nil || head.Type() != ai.ACTION_CAST_SPELL_ON_OBJECT || head.ArgObj(2) != f.target ||
		head.ArgU32(0) != uint32(spell.SPELL_POISON) || ud.Field120_2 != 0 ||
		uint32(ud.Field120_1) != ud.MonsterDef.MissileAttackFrame216 {
		e2eError(fmt.Errorf("Poison did not use the real NPC animation/cast-frame gate"))
		return
	}
	f.naturalSeen = true
	e2eLog.Printf("POISON NPC CAST FRAME: animation=%d duration=%d target=%p elapsed=%d", ud.Field120_1, ud.Field120_2, head.ArgObj(2), noxServer.Frame()-f.frame)
}

func (f *e2ePoisonFixture) observeEffect() {
	f.applied = f.target.HealthData.Field16
	var ok bool
	f.firstTick, ok = e2ePoisonFirstTick(f.applied, f.power)
	// Events queued outside the audio loop may be replayed on the next frame;
	// the real timer can already have received that frame's single decrement.
	if f.effectSeen || !ok || f.target.Poison540 != f.power || f.target.HealthData.Cur != f.health ||
		f.applied-f.frame > 150 || noxServer.Frame()-f.applied > 1 ||
		f.target.Field542 < 999 || f.target.Field542 > 1000 {
		e2eError(fmt.Errorf("Poison application mismatch: power=%d/%d HP=%d/%d applied=%d frame=%d timer=%d", f.target.Poison540, f.power, f.target.HealthData.Cur, f.health, f.applied, noxServer.Frame(), f.target.Field542))
		return
	}
	f.effectSeen = true
	e2eLog.Printf("POISON APPLIED: direction=%s dose=%d HP=%d unchanged applied=%d first-DOT=%d timer=%d", f.direction, f.power, f.health, f.applied, f.firstTick, f.target.Field542)
}

func (f *e2ePoisonFixture) observeDOT() {
	if !f.active || !f.effectSeen || f.hitSeen {
		return
	}
	if !e2eObjectInWorld(f.target) || f.target.HealthData.Cur == 0 || f.target.Poison540 != f.power {
		e2eError(fmt.Errorf("Poison target disappeared or lost its live state"))
		return
	}
	if f.target.HealthData.Cur == f.health {
		return
	}
	var marker, markerType, wantType uint32
	if f.fromNPC {
		ud := f.target.UpdateDataPlayer()
		marker, markerType, wantType = ud.Field76, ud.Field75, math.Float32bits(float32(object.DamagePoison))
	} else {
		ud := f.target.UpdateDataMonster()
		marker, markerType, wantType = ud.Field547, ud.Field546, uint32(object.DamagePoison)
		if !ud.StatusFlags.Has(object.MonStatusInjured) {
			e2eError(fmt.Errorf("Poison NPC injured latch missing: HP=%d->%d hit=%d expected=%d observed=%d status=%#x", f.health, f.target.HealthData.Cur, f.target.Frame134, f.firstTick, noxServer.Frame(), ud.StatusFlags))
			return
		}
	}
	if f.target.HealthData.Cur != f.health-1 || f.target.Frame134 != f.firstTick || f.target.HealthData.Field16 != f.applied ||
		f.target.Obj130 != nil || f.target.Field131 != uint32(object.DamagePoison) || f.target.Pos132 != (types.Pointf{}) ||
		marker != 2 || markerType != wantType {
		e2eError(fmt.Errorf("Poison DOT mismatch: direction=%s HP=%d->%d hit=%d/%d marker=%d/%#x/%#x attribution=%p", f.direction, f.health, f.target.HealthData.Cur, f.target.Frame134, f.firstTick, marker, markerType, wantType, f.target.Obj130))
		return
	}
	f.hitSeen = true
	e2eLog.Printf("POISON DOT: direction=%s HP=%d->%d applied=%d hit=%d observed=%d marker=%d/%#x", f.direction, f.health, f.target.HealthData.Cur, f.applied, f.target.Frame134, noxServer.Frame(), marker, markerType)
}

func (f *e2ePoisonFixture) complete() bool {
	if !f.hitSeen || f.castAudio != 1 || f.effectAudio != 1 || f.magic == nil ||
		f.fromNPC && (!f.naturalSeen || f.npc.MonsterActionIsScheduled(ai.ACTION_CAST_SPELL_ON_OBJECT)) {
		return false
	}
	if !e2eObjectInWorld(f.target) || f.target.HealthData.Cur != f.health-1 || f.target.Poison540 != f.power {
		e2eError(fmt.Errorf("Poison target changed after its verified first DOT"))
		return true
	}
	if e2eFistInWorld(f.magic, f.magicWire, f.magicScript) || noxClient.Objs.ByNetCode(uint16(f.magicWire)) != nil {
		return false
	}
	drawable := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target)))
	if drawable == nil {
		return false
	}
	if f.fromNPC {
		meter, ready := e2eClientHUDMeter(0)
		if !ready || !meter.Poisoned || !meter.PoisonTubeReady || meter.Current != uint32(f.health-1) ||
			meter.Maximum != uint32(f.target.HealthData.Max) || f.target.UpdateDataPlayer().Player.Field3680&0x400 == 0 {
			return false
		}
		pixels, err := e2eAssertHUDMeterPixels(noxClient.r.CopyPixBuffer(), meter, 'g', "natural Poison health")
		if err != nil {
			e2eError(err)
			return true
		}
		e2eLog.Printf("POISON PLAYER HUD: HP=%d/%d poisoned=%t green-pixels=%d status=%#x", meter.Current, meter.Maximum, meter.Poisoned, pixels.Filled, f.target.UpdateDataPlayer().Player.Field3680)
	} else if delta, ok := legacy.HealthChangeForDrawable(drawable.NetCode32); !ok || delta != -1 {
		return false
	}
	e2eLog.Printf("POISON COMPLETE: direction=%s requested-level=%d actual-power=%d HP=%d->%d applied=%d DOT=%d client-replay=verified projectile=removed NPC-natural=%t", f.direction, f.level, f.power, f.health, f.target.HealthData.Cur, f.applied, f.target.Frame134, f.naturalSeen)
	return true
}

// Normal script object casts and one naturally animated queued NPC cast.
// Only placement, waiting AI and explicit target selection are fixture setup.
// No poison, HP, timing, damage, projectile, status, packet or pixel result is
// injected. Cure is cleanup after all spell/DOT/client assertions have passed.
func (sc *e2eScenario) CheckPoisonSpell(level int, direction, name string) {
	fromNPC, ok := e2ePoisonMode(level, direction)
	if !ok {
		e2eError(fmt.Errorf("invalid Poison fixture: level=%d direction=%q", level, direction))
		return
	}
	f := &e2ePoisonFixture{level: level, direction: direction, fromNPC: fromNPC}
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		host := noxServer.Players.HostUnit()
		return host != nil && !host.HasEnchant(server.ENCHANT_INVULNERABLE) &&
			noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, f.prepare)
	sc.Wait(12, name+" publish placement")
	sc.add(0, name+" cast", f.beginCast)
	sc.addWhen(1, name+" actual DOT and client replay", 300, f.complete, func() {
		f.active = false
		noxServer.Server.RemovePoison4EE9D0(f.target)
		if f.target.Poison540 != 0 || f.target.HealthData.Field16 != 0 {
			e2eError(fmt.Errorf("Poison cleanup did not use the normal cure service"))
			return
		}
		noxServer.DelayedDelete(f.npc)
		asObjectS(f.host).SetPos(f.original)
		f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	})
	sc.addWhen(3, name+" cure replay", 120, func() bool {
		if !f.fromNPC {
			return true
		}
		meter, ready := e2eClientHUDMeter(0)
		return ready && !meter.Poisoned && f.host.UpdateDataPlayer().Player.Field3680&0x400 == 0
	}, func() { e2eLog.Printf("POISON CLEANUP: direction=%s player-status=cleared", f.direction) })
}
