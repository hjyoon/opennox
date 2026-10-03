package opennox

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	nsp "github.com/opennox/noxscript/ns/v4/spell"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func e2eFireballUnitMode(level int, direction string) (fromNPC, ok bool) {
	fromNPC, ok = e2eFistUnitDirection(direction)
	return fromNPC, ok && (level >= 1 && level <= 5 || fromNPC && level == 0)
}

func e2eFireballUnitType(level int) string {
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

type e2eFireballUnitFixture struct {
	level, actualLevel, castAudio, explodeAudio int
	direction                                   string
	fromNPC, active, natural, hit               bool
	host, npc, caster, target, projectile       *server.Object
	original                                    types.Pointf
	hostHP, hostMax, health, projectileType     uint16
	wire, frame                                 uint32
	script                                      int32
	damage                                      int32
	carry                                       uint32
}

func (f *e2eFireballUnitFixture) prepare() {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.HealthData == nil || f.host.HealthData.Cur == 0 || f.host.ControllingPlayer() == nil ||
		f.host.Flags().HasAny(object.FlagDead|object.FlagDestroyed|object.FlagNoUpdate) {
		e2eError(fmt.Errorf("Fireball requires a live host"))
		return
	}
	f.original, f.hostHP, f.hostMax = f.host.PosVec, f.host.HealthData.Cur, f.host.HealthData.Max
	origin, direction, err := e2eWarriorAbilityArena(f.original, f.host.Shape.Circle.R+16, func(from, to types.Pointf) bool {
		return e2eWarriorLaneClear(f.host, from, to)
	})
	if err != nil {
		e2eError(err)
		return
	}
	f.npc = noxServer.NewObjectByTypeID("NPC")
	if f.npc == nil {
		e2eError(fmt.Errorf("Fireball requires the stock NPC"))
		return
	}
	asObjectS(f.host).SetPos(origin)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	noxServer.CreateObjectAt(f.npc, nil, origin.Add(direction.Mul(112)))
	noxServer.ObjectsAddPending()
	if !e2eObjectInWorld(f.npc) || f.npc.UpdateData == nil || f.npc.HealthData == nil ||
		f.npc.Damage != f.host.Damage || !f.npc.SubClass().AsMonster().Has(object.MonsterNPC) {
		e2eError(fmt.Errorf("Fireball stock NPC lacks the real PlayerDamage callback"))
		return
	}
	// Durable HP, placement and ordinary waiting AI are setup, not hit results.
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
		e2eError(fmt.Errorf("Fireball baseline units are enchanted"))
		return
	}
	// Keep stock armor. This baseline excludes modifier/buff resistance;
	// resistance, Shield and live callbacks have separate native unit tests.
	for item := f.target.FirstItem(); item != nil; item = item.NextItem() {
		if !item.Flags().Has(object.FlagEquipped) || item.InitData == nil ||
			!item.Class().HasAny(object.ClassArmor|object.ClassWeapon|object.ClassWand) {
			continue
		}
		for _, m := range item.InitDataModifier().Modifiers {
			if m != nil && (m.Defend76.Fnc != nil || m.Engage112 != nil) {
				e2eError(fmt.Errorf("Fireball baseline has an equipped protection modifier"))
				return
			}
		}
	}
	f.actualLevel = f.level
	if f.level == 0 {
		f.actualLevel = int(noxServer.Server.SpellPower4FE7B0(spell.SPELL_FIREBALL, f.caster))
	}
	if e2eFireballUnitType(f.actualLevel) == "" {
		e2eError(fmt.Errorf("Fireball mode-selected level is outside stock range: %d", f.actualLevel))
		return
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, p := range []unsafe.Pointer{unsafe.Pointer(f.host), unsafe.Pointer(f.npc), f.host.UpdateData, f.npc.UpdateData} {
			if uintptr(p) <= math.MaxUint32 {
				e2eError(fmt.Errorf("Fireball unit pointer is below 4 GiB: %p", p))
				return
			}
		}
	}
	noxServer.Audio.OnSound(func(id sound.ID, kind int, owner *server.Object, _ types.Pointf) {
		if !f.active {
			return
		}
		if id == sound.SoundFireballCast && owner == f.caster {
			f.castAudio++
			if kind != 0 || f.castAudio != 1 {
				e2eError(fmt.Errorf("Fireball cast sound kind/count=%d/%d", kind, f.castAudio))
				return
			}
			if f.projectile == nil {
				f.observeProjectile()
			}
			if f.level == 0 {
				ud, head := f.npc.UpdateDataMonster(), f.npc.UpdateDataMonster().AIStackHead()
				if head == nil || head.Type() != ai.ACTION_CAST_SPELL_ON_OBJECT || head.ArgObj(2) != f.target ||
					head.ArgU32(0) != uint32(spell.SPELL_FIREBALL) || ud.Field120_2 != 0 ||
					uint32(ud.Field120_1) != ud.MonsterDef.MissileAttackFrame216 {
					e2eError(fmt.Errorf("Fireball NPC bypassed the natural animation/cast frame"))
					return
				}
				f.natural = true
			}
		}
		if id == sound.SoundFireballExplode && owner == f.projectile {
			f.explodeAudio++
			if kind != 0 || f.explodeAudio != 1 {
				e2eError(fmt.Errorf("Fireball explosion sound kind/count=%d/%d", kind, f.explodeAudio))
			}
		}
	})
	noxServer.TickHook(f.observeHit)
}

func (f *e2eFireballUnitFixture) beginCast() {
	if !e2eObjectInWorld(f.target) || !noxServer.MapTraceRay(f.caster.PosVec, f.target.PosVec, 0) ||
		noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target))) == nil {
		e2eError(fmt.Errorf("Fireball target placement/client publication did not settle"))
		return
	}
	f.health, f.frame, f.active = f.target.HealthData.Cur, noxServer.Frame(), true
	if f.fromNPC {
		f.carry = f.target.UpdateDataPlayer().Field21
	} else {
		f.carry = f.target.UpdateDataMonster().Field1
	}
	// Explicit script aim is setup; the projectile/collision remains real.
	f.caster.SetDir(server.DirFromVec(f.target.PosVec.Sub(f.caster.PosVec)))
	if f.level == 0 {
		f.npc.MonsterPushAction(ai.ACTION_CAST_SPELL_ON_OBJECT, uint32(spell.SPELL_FIREBALL), 0, f.target)
		return
	}
	api := noxServer.noxScriptP()
	api.CastSpellLvl(nsp.Spell("SPELL_FIREBALL"), f.level, api.toObj(f.caster), api.toObj(f.target))
	noxServer.ObjectsAddPending()
	f.observeProjectile()
}

func (f *e2eFireballUnitFixture) observeProjectile() {
	want := e2eFireballUnitType(f.actualLevel)
	for obj := f.caster.Field129; obj != nil; obj = obj.Field128 {
		if obj.ObjectTypeC().ID() != want || obj.Flags().Has(object.FlagDestroyed) {
			continue
		}
		if f.projectile != nil || obj.ObjOwner != f.caster || !obj.Class().Has(object.ClassMissile) || obj.CollideData == nil ||
			unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 || uintptr(obj.CollideData) <= math.MaxUint32) {
			e2eError(fmt.Errorf("Fireball projectile lost native type/owner/collision data"))
			return
		}
		f.projectile, f.wire, f.script, f.projectileType = obj, obj.NetCode, obj.ScriptIDVal, obj.TypeInd
		f.damage = int32((*server.SparkExplosionCollideData)(obj.CollideData).Power >> 1)
		if f.damage <= 0 || f.damage >= int32(f.health) {
			e2eError(fmt.Errorf("Fireball collision damage is outside durable baseline: %d", f.damage))
			return
		}
		e2eLog.Printf("FIREBALL PROJECTILE: direction=%s requested=%d actual=%d type=%s object=%p collide=%p damage=%d HP=%d",
			f.direction, f.level, f.actualLevel, want, obj, obj.CollideData, f.damage, f.health)
	}
	if f.projectile == nil {
		e2eError(fmt.Errorf("Fireball produced no real %s projectile", want))
	}
}

func (f *e2eFireballUnitFixture) observeHit() {
	if !f.active || f.hit || f.projectile == nil || f.target.HealthData.Cur == f.health {
		return
	}
	var marker, markerType, carry uint32
	if f.fromNPC {
		ud := f.target.UpdateDataPlayer()
		marker, markerType, carry = ud.Field76, ud.Field75, ud.Field21
	} else {
		ud := f.target.UpdateDataMonster()
		marker, markerType, carry = ud.Field547, ud.Field546, ud.Field1
		if !ud.StatusFlags.Has(object.MonStatusInjured) || !ud.StatusFlags.Has(object.MonStatusOnFire) {
			e2eError(fmt.Errorf("Fireball NPC lacks the injured/on-fire flags"))
			return
		}
	}
	if f.target.HealthData.Cur != f.health-uint16(f.damage) || f.target.Obj130 != f.projectile ||
		f.target.Field131 != uint32(object.DamageFlame) || marker != 1 || markerType != uint32(f.projectileType) || carry != f.carry {
		e2eError(fmt.Errorf("Fireball hit mismatch: direction=%s level=%d HP=%d->%d want=%d attribution=%p/%p marker=%d/%d carry=%#x/%#x",
			f.direction, f.level, f.health, f.target.HealthData.Cur, f.health-uint16(f.damage), f.target.Obj130, f.projectile, marker, markerType, carry, f.carry))
		return
	}
	f.hit = true
	e2eLog.Printf("FIREBALL HIT: direction=%s requested=%d actual=%d HP=%d->%d marker=%d/%d elapsed=%d natural-NPC=%t",
		f.direction, f.level, f.actualLevel, f.health, f.target.HealthData.Cur, marker, markerType, noxServer.Frame()-f.frame, f.natural)
}

func (f *e2eFireballUnitFixture) complete() bool {
	if !f.hit || f.castAudio != 1 || f.explodeAudio != 1 || f.level == 0 &&
		(!f.natural || f.npc.MonsterActionIsScheduled(ai.ACTION_CAST_SPELL_ON_OBJECT)) {
		return false
	}
	if e2eFistInWorld(f.projectile, f.wire, f.script) || noxClient.Objs.ByNetCode(uint16(f.wire)) != nil {
		return false
	}
	dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target)))
	if dr == nil {
		return false
	}
	if f.fromNPC {
		meter, ready := e2eClientHUDMeter(0)
		if !ready || meter.Current != uint32(f.health)-uint32(f.damage) || meter.Maximum != uint32(f.target.HealthData.Max) {
			return false
		}
	} else if delta, ok := legacy.HealthChangeForDrawable(dr.NetCode32); !ok || int32(delta) != -f.damage {
		return false
	}
	e2eLog.Printf("FIREBALL COMPLETE: direction=%s requested=%d actual=%d server/client-HP=verified projectile=removed cast/explode=1/1 natural-NPC=%t",
		f.direction, f.level, f.actualLevel, f.natural)
	return true
}

func (f *e2eFireballUnitFixture) cleanup() {
	f.active = false
	noxServer.DelayedDelete(f.npc)
	asObjectS(f.host).SetPos(f.original)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	if f.fromNPC {
		asObjectS(f.host).SetMaxHealth(int(f.hostMax))
		asObjectS(f.host).SetHealth(int(f.hostHP))
	}
}

// Regular script casts in both directions and one naturally animated NPC cast.
// No damage callback, collision, hit metadata, packet or rendered result is
// supplied by the observer. This does not simulate mana/incantation input.
func (sc *e2eScenario) CheckFireballUnitDamage(level int, direction, name string) {
	fromNPC, ok := e2eFireballUnitMode(level, direction)
	if !ok {
		e2eError(fmt.Errorf("invalid Fireball fixture: %s/%d", direction, level))
		return
	}
	f := &e2eFireballUnitFixture{level: level, direction: direction, fromNPC: fromNPC}
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		host := noxServer.Players.HostUnit()
		return host != nil && host.Buffs == 0 && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, f.prepare)
	sc.Wait(12, name+" publish units")
	sc.add(0, name+" cast", f.beginCast)
	sc.addWhen(1, name+" actual hit and client replay", 300, f.complete, f.cleanup)
	sc.Wait(3, name+" cleanup")
}
