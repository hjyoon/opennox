package opennox

import (
	"fmt"
	"image"
	"math"
	"strings"
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

func e2eCleansingFlameMode(level int, mode string) bool {
	return mode == "red-player" && level >= 1 && level <= 5 ||
		(mode == "red-npc" || mode == "blue-npc") && level == 0
}

func e2eCleansingFlameType(name string) bool {
	switch name {
	case "SmallFlameCleanse", "MediumFlameCleanse", "FlameCleanse", "LargeFlameCleanse",
		"SmallBlueFlameCleanse", "MediumBlueFlameCleanse", "BlueFlameCleanse", "LargeBlueFlameCleanse":
		return true
	}
	return false
}

type e2eCleansingFlameRecord struct {
	obj                                    *server.Object
	wire, deadline, last                   uint32
	script                                 int32
	kind                                   uint16
	direction                              server.Dir16
	position, velocity                     types.Pointf
	clientPosition                         image.Point
	predicted, moved, clientMoved, removed bool
}

type e2eCleansingFlameFixture struct {
	level, castAudio, expired, early int
	mode                             string
	id                               spell.ID
	host, caster, target             *server.Object
	original                         types.Pointf
	hostHP, hostMax, hostMana        uint16
	health, mana, previousMana       uint16
	frame                            uint32
	flames                           []e2eCleansingFlameRecord
	active, natural, hit, verified   bool
}

func (f *e2eCleansingFlameFixture) ownedFlames() []*server.Object {
	var out []*server.Object
	for obj, remaining := f.caster.Field129, 4096; obj != nil; obj, remaining = obj.Field128, remaining-1 {
		if remaining == 0 {
			e2eError(fmt.Errorf("Cleansing Flame owned list did not terminate"))
			return nil
		}
		if typ := obj.ObjectTypeC(); typ != nil && e2eCleansingFlameType(typ.ID()) {
			out = append(out, obj)
		}
	}
	return out
}

func (f *e2eCleansingFlameFixture) prepare() {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.ControllingPlayer() == nil || f.host.HealthData == nil ||
		f.host.HealthData.Cur == 0 || f.host.Buffs != 0 ||
		f.host.Flags().HasAny(object.FlagDead|object.FlagDestroyed|object.FlagNoUpdate) {
		e2eError(fmt.Errorf("Cleansing Flame requires a live unenchanted Wizard host"))
		return
	}
	e2eQueueInput(&seat.MouseMoveEvent{Pos: noxClient.Inp.GetMousePos()})
	f.original, f.hostHP, f.hostMax = f.host.PosVec, f.host.HealthData.Cur, f.host.HealthData.Max
	f.hostMana = f.host.UpdateDataPlayer().ManaCur
	origin, direction, err := e2eWarriorAbilityArena(f.original, f.host.Shape.Circle.R+40,
		func(from, to types.Pointf) bool { return e2eWarriorLaneClear(f.host, from, to) })
	if err != nil {
		e2eError(err)
		return
	}
	asObjectS(f.host).SetPos(origin)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	f.id = spell.SPELL_CLEANSING_FLAME
	f.caster = f.host
	if f.mode != "red-player" {
		f.caster, f.target = noxServer.NewObjectByTypeID("NPC"), f.host
		if f.caster == nil || f.caster.UpdateData == nil {
			e2eError(fmt.Errorf("Cleansing Flame requires the stock NPC caster"))
			return
		}
		// The unowned stock NPC is hostile to the host. Placement and WAIT
		// reserve the normal animation; no cast-frame or enemy result is set.
		noxServer.CreateObjectAt(f.caster, nil, origin.Add(direction.Mul(64)))
		f.caster.UpdateDataMonster().SetAggression(0)
		f.caster.ClearActionStack()
		f.caster.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
	} else {
		f.target = noxServer.NewObjectByTypeID("Wolf")
		if f.target == nil || f.target.UpdateData == nil || f.target.SubClass().Has(0x400) {
			e2eError(fmt.Errorf("Cleansing Flame requires the stock fire-susceptible Wolf"))
			return
		}
		noxServer.CreateObjectAt(f.target, nil, origin.Add(direction.Mul(64)))
		f.target.UpdateDataMonster().SetAggression(0)
		f.target.ClearActionStack()
		f.target.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
	}
	if f.mode == "blue-npc" {
		f.id = spell.SPELL_CLEANSING_MANA_FLAME
		if f.hostMana == 0 {
			e2eError(fmt.Errorf("blue flame requires actual stock Wizard mana"))
			return
		}
	} else {
		// Durability is setup only, not an injected damage outcome.
		asObjectS(f.target).SetMaxHealth(2000)
	}
	noxServer.ObjectsAddPending()
	if !noxServer.S().IsEnemyTo(f.target, f.caster.FindOwnerChainPlayer()) || len(f.ownedFlames()) != 0 {
		e2eError(fmt.Errorf("Cleansing Flame requires enemy units without a prior owned flame"))
		return
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(f.caster), unsafe.Pointer(f.target),
			f.caster.UpdateData, f.target.UpdateData, unsafe.Pointer(f.target.HealthData)} {
			if uintptr(ptr) <= math.MaxUint32 {
				e2eError(fmt.Errorf("Cleansing Flame unit/update/health below 4 GiB: %p", ptr))
				return
			}
		}
	}
	noxServer.Audio.OnSound(f.observeSound)
	noxServer.TickHook(f.observeTick)
}

func (f *e2eCleansingFlameFixture) beginCast() {
	if !e2eObjectInWorld(f.target) || noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target))) == nil {
		e2eError(fmt.Errorf("Cleansing Flame target was not network-published"))
		return
	}
	f.health, f.frame, f.active = f.target.HealthData.Cur, noxServer.Frame(), true
	if f.target == f.host {
		f.mana = f.host.UpdateDataPlayer().ManaCur
		f.previousMana = f.mana
	}
	if f.mode != "red-player" {
		f.caster.SetDir(server.DirFromVec(f.target.PosVec.Sub(f.caster.PosVec)))
		f.caster.MonsterPushAction(ai.ACTION_CAST_SPELL_ON_OBJECT, uint32(f.id), 0, f.target)
		return
	}
	api := noxServer.noxScriptP()
	api.CastSpellLvl(nsp.Spell("SPELL_CLEANSING_FLAME"), f.level, api.toObj(f.caster), api.toObj(f.target))
	noxServer.ObjectsAddPending()
}

func (f *e2eCleansingFlameFixture) observeSound(id sound.ID, kind int, source *server.Object, _ types.Pointf) {
	if !f.active {
		return
	}
	if id == sound.ID(228) && f.mode == "blue-npc" {
		for i := range f.flames {
			p := &f.flames[i]
			if source == p.obj && e2eFistInWorld(p.obj, p.wire, p.script) &&
				f.host.UpdateDataPlayer().ManaCur < f.previousMana {
				if kind != 0 || f.host.HealthData.Cur != f.health {
					e2eError(fmt.Errorf("blue flame changed HP or used non-object drain audio"))
					return
				}
				f.hit = true
				e2eLog.Printf("CLEANSING FLAME MANA HIT: flame=%p mana=%d->%d HP=%d elapsed=%d",
					source, f.previousMana, f.host.UpdateDataPlayer().ManaCur, f.health, noxServer.Frame()-f.frame)
				return
			}
		}
		return
	}
	if id != sound.SoundCleansingFlameCast || source != f.caster {
		return
	}
	f.castAudio++
	if kind != 0 || f.castAudio != 1 {
		e2eError(fmt.Errorf("Cleansing Flame cast audio mismatch"))
		return
	}
	if f.mode != "red-player" {
		ud := f.caster.UpdateDataMonster()
		head := ud.AIStackHead()
		if head == nil || head.Type() != ai.ACTION_CAST_SPELL_ON_OBJECT || head.ArgObj(2) != f.target ||
			head.ArgU32(0) != uint32(f.id) || ud.Field120_2 != 0 || uint32(ud.Field120_1) != ud.MonsterDef.MissileAttackFrame216 {
			e2eError(fmt.Errorf("Cleansing Flame NPC bypassed natural animation/cast-frame"))
			return
		}
		f.natural = true
	}
	owned := f.ownedFlames()
	if len(owned) == 0 || len(owned) > 48 || len(f.flames) != 0 {
		e2eError(fmt.Errorf("Cleansing Flame produced %d owned objects", len(owned)))
		return
	}
	now, fps := noxServer.Frame(), noxServer.TickRate()
	for _, obj := range owned {
		blue := strings.Contains(obj.ObjectTypeC().ID(), "Blue")
		if blue != (f.mode == "blue-npc") || obj.ObjOwner != f.caster ||
			obj.Update != legacy.Get_nox_xxx_updateFlameCleanse_53D510() || obj.IsUpdatable == 0 ||
			!obj.Class().Has(object.ClassClientPredict) || obj.Float28 != 0 ||
			obj.Direction1 != obj.Direction2 || obj.Field34-now < 3*fps || obj.Field34-now > 6*fps {
			e2eError(fmt.Errorf("Cleansing Flame lost stock type/owner/update/prediction/lifetime contract"))
			return
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 {
			e2eError(fmt.Errorf("Cleansing Flame object below 4 GiB"))
			return
		}
		f.flames = append(f.flames, e2eCleansingFlameRecord{
			obj: obj, wire: obj.NetCode, script: obj.ScriptIDVal, kind: obj.TypeInd,
			deadline: obj.Field34, last: now, direction: obj.Direction1,
			position: obj.PosVec, velocity: obj.VelVec,
		})
	}
	e2eLog.Printf("CLEANSING FLAME CAST: mode=%s requested=%d caster=%p target=%p flames=%d HP=%d mana=%d natural-NPC=%t",
		f.mode, f.level, f.caster, f.target, len(f.flames), f.health, f.mana, f.natural)
}

func (f *e2eCleansingFlameFixture) observeTick() {
	if !f.active || len(f.flames) == 0 {
		return
	}
	now := noxServer.Frame()
	for i := range f.flames {
		p := &f.flames[i]
		if p.removed {
			continue
		}
		if e2eFistInWorld(p.obj, p.wire, p.script) {
			p.last = now
			p.moved = p.moved || p.obj.PosVec != p.position
			if p.obj.Field34 != p.deadline || p.obj.ObjOwner != f.caster {
				e2eError(fmt.Errorf("Cleansing Flame update altered deadline/owner"))
				return
			}
			dr := noxClient.Objs.ByNetCode(uint16(p.wire))
			if dr != nil && dr.Field_115 == legacy.Get_nox_xxx_sprite_4CA540() && dr.InClientUpdateList != 0 {
				if dr.TypeIDVal != uint32(p.kind) || uint16(dr.Field_127) != uint16(p.direction) ||
					dr.Field_81 != uint32(uint16(p.position.X)) || dr.Field_82 != uint32(uint16(p.position.Y)) ||
					dr.Field_119 != 0 ||
					dr.Field_117 != math.Float32bits(float32(int8(int32(p.velocity.X*16)))*0.0625) ||
					dr.Field_118 != math.Float32bits(float32(int8(int32(p.velocity.Y*16)))*0.0625) {
					e2eError(fmt.Errorf("Cleansing Flame real prediction packet/drawable mismatch"))
					return
				}
				if !p.predicted {
					p.clientPosition, p.predicted = dr.PosVec, true
				} else if !p.clientMoved && dr.PosVec != p.clientPosition && p.moved {
					// Both server movement and a real client prediction update
					// must be observed; a synthesized packet cannot satisfy this.
					e2eLog.Printf("CLEANSING FLAME PREDICTED: mode=%s flame=%p wire=%d client=%v->%v",
						f.mode, p.obj, p.wire, p.clientPosition, dr.PosVec)
					p.clientPosition = dr.PosVec
					p.clientMoved = true
				}
			}
		} else {
			p.removed = true
			if p.last+1 >= p.deadline {
				f.expired++
			} else {
				f.early++
			}
		}
	}
	if f.mode == "blue-npc" {
		f.previousMana = f.host.UpdateDataPlayer().ManaCur
		return
	}
	if !f.hit && f.target.HealthData.Cur < f.health {
		for i := range f.flames {
			p := &f.flames[i]
			if f.target.Obj130 == p.obj && e2eFistInWorld(p.obj, p.wire, p.script) &&
				f.target.Field131 == uint32(object.DamageFlame) && f.target.HealthData.Cur != 0 {
				f.hit = true
				e2eLog.Printf("CLEANSING FLAME HP HIT: mode=%s flame=%p HP=%d->%d elapsed=%d",
					f.mode, p.obj, f.health, f.target.HealthData.Cur, now-f.frame)
				return
			}
		}
	}
}

func (f *e2eCleansingFlameFixture) predicted() bool {
	for _, p := range f.flames {
		if p.predicted && p.moved && p.clientMoved {
			return true
		}
	}
	return false
}

func (f *e2eCleansingFlameFixture) clientHit() bool {
	if !f.hit {
		return false
	}
	if f.mode == "red-player" {
		dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target)))
		if dr == nil {
			return false
		}
		delta, ok := legacy.HealthChangeForDrawable(dr.NetCode32)
		return ok && delta < 0
	}
	slot := 0
	if f.mode == "blue-npc" {
		slot = 1
	}
	meter, ready := e2eClientHUDMeter(slot)
	baseline := f.health
	if slot == 1 {
		baseline = f.mana
	}
	return ready && meter.Current < uint32(baseline)
}

func (f *e2eCleansingFlameFixture) complete() bool {
	if !f.hit || !f.predicted() || f.castAudio != 1 || f.expired+f.early != len(f.flames) ||
		f.expired == 0 || len(f.ownedFlames()) != 0 ||
		f.mode != "red-player" && (!f.natural || f.caster.MonsterActionIsScheduled(ai.ACTION_CAST_SPELL_ON_OBJECT)) {
		return false
	}
	for _, p := range f.flames {
		if !p.removed || e2eFistInWorld(p.obj, p.wire, p.script) || noxClient.Objs.ByNetCode(uint16(p.wire)) != nil {
			return false
		}
	}
	f.verified = true
	e2eLog.Printf("CLEANSING FLAME COMPLETE: mode=%s requested=%d spawned=%d expired=%d early=%d HP=%d->%d owner/world/drawable=removed natural-NPC=%t",
		f.mode, f.level, len(f.flames), f.expired, f.early, f.health, f.target.HealthData.Cur, f.natural)
	return true
}

func (f *e2eCleansingFlameFixture) cleanup() {
	if !f.verified {
		e2eError(fmt.Errorf("Cleansing Flame cleanup requires verified natural results"))
		return
	}
	f.active = false
	if f.caster != f.host {
		noxServer.DelayedDelete(f.caster)
	} else {
		noxServer.DelayedDelete(f.target)
	}
	asObjectS(f.host).SetPos(f.original)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	if f.mode == "red-npc" {
		asObjectS(f.host).SetMaxHealth(int(f.hostMax))
		asObjectS(f.host).SetHealth(int(f.hostHP))
	}
	if f.mode == "blue-npc" {
		asObjectS(f.host).SetMana(int(f.hostMana))
	}
}

// Script acceptance and normal NPC CAST_ON_OBJECT animation use stock
// objects, RNG, physics, collision, prediction and expiry. This does not
// simulate player incantation/mana spending or autonomous AI spell choice.
func (sc *e2eScenario) CheckCleansingFlameSpell(level int, mode, name string) {
	if !e2eCleansingFlameMode(level, mode) {
		e2eError(fmt.Errorf("invalid Cleansing Flame mode/level %s/%d", mode, level))
		return
	}
	f := &e2eCleansingFlameFixture{level: level, mode: mode}
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		host := noxServer.Players.HostUnit()
		return host != nil && host.Buffs == 0 && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, f.prepare)
	sc.Wait(12, name+" publish units")
	sc.add(0, name+" cast", f.beginCast)
	sc.addWhen(1, name+" real moving prediction", 180, f.predicted, func() {})
	sc.CaptureMagicFrame(name + " actual flame frame")
	sc.addWhen(1, name+" actual collision and client replay", 300, f.clientHit, func() {})
	sc.CaptureMagicFrame(name + " actual collision frame")
	sc.addWhen(1, name+" natural expiry", 600, f.complete, func() {})
	sc.add(0, name+" cleanup", f.cleanup)
	sc.Wait(3, name+" cleanup complete")
}
