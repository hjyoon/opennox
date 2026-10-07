package opennox

import (
	"fmt"
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

func e2eEarthquakeMode(level int, mode string) bool {
	return mode == "player-script" && level >= 1 && level <= 5 || mode == "npc-animated" && level == 0
}

func e2eEarthquakeDamageObserved(before, after uint16, attributed, caster *server.Object, typ uint32) bool {
	return caster != nil && attributed == caster && typ == uint32(object.DamageImpact) && after != 0 && after < before
}

type e2eEarthquakeFixture struct {
	level, castAudio                  int
	mode                              string
	host, caster, target              *server.Object
	original                          types.Pointf
	hostHP, hostMax, health, casterHP uint16
	frame                             uint32
	active, natural, hit, quake       bool
	verified                          bool
}

func (f *e2eEarthquakeFixture) prepare() {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.ControllingPlayer() == nil || f.host.HealthData == nil || f.host.HealthData.Cur == 0 ||
		f.host.Buffs != 0 || f.host.Poison540 != 0 || f.host.Flags().HasAny(object.FlagDead|object.FlagDestroyed|object.FlagNoUpdate) {
		e2eError(fmt.Errorf("Earthquake requires a live unenchanted Wizard host"))
		return
	}
	e2eQueueInput(&seat.MouseMoveEvent{Pos: noxClient.Inp.GetMousePos()})
	f.original, f.hostHP, f.hostMax = f.host.PosVec, f.host.HealthData.Cur, f.host.HealthData.Max
	origin, direction, err := e2eWarriorAbilityArena(f.original, f.host.Shape.Circle.R+24,
		func(from, to types.Pointf) bool { return e2eWarriorLaneClear(f.host, from, to) })
	if err != nil {
		e2eError(err)
		return
	}
	asObjectS(f.host).SetPos(origin)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	f.caster = f.host
	if f.mode == "npc-animated" {
		// An unowned stock NPC is hostile to the host. Reserve its ordinary
		// cast animation with WAIT; do not provide an enemy or frame result.
		f.caster, f.target = noxServer.NewObjectByTypeID("NPC"), f.host
	} else {
		f.target = noxServer.NewObjectByTypeID("Wolf")
	}
	unit := f.target
	if f.mode == "npc-animated" {
		unit = f.caster
	}
	if unit == nil || unit.HealthData == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) {
		e2eError(fmt.Errorf("Earthquake requires an initialized stock %s unit", f.mode))
		return
	}
	noxServer.CreateObjectAt(unit, nil, origin.Add(direction.Mul(64)))
	unit.UpdateDataMonster().SetAggression(0)
	unit.ClearActionStack()
	unit.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
	// Position, durability and WAIT are setup, never a supplied damage result.
	asObjectS(f.target).SetMaxHealth(2000)
	noxServer.ObjectsAddPending()
	if !noxServer.S().IsEnemyTo(f.target, f.caster.FindOwnerChainPlayer()) ||
		f.target.Flags().Has(object.FlagAirborne) || noxClient.Viewport().Jiggle12 != 0 {
		e2eError(fmt.Errorf("Earthquake requires a grounded enemy and settled viewport"))
		return
	}
	if radius := float32(noxServer.Balance.Float("EarthquakeRange")); radius <= 64 {
		e2eError(fmt.Errorf("stock EarthquakeRange does not cover the fixture"))
		return
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(f.caster), unsafe.Pointer(f.target),
			f.caster.UpdateData, f.target.UpdateData, unsafe.Pointer(f.target.HealthData)} {
			if uintptr(ptr) <= math.MaxUint32 {
				e2eError(fmt.Errorf("Earthquake fixture pointer below 4 GiB: %p", ptr))
				return
			}
		}
	}
	noxServer.Audio.OnSound(f.observeSound)
	noxServer.TickHook(f.observeTick)
}

func (f *e2eEarthquakeFixture) beginCast() {
	if !e2eObjectInWorld(f.target) || noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target))) == nil {
		e2eError(fmt.Errorf("Earthquake target was not network-published"))
		return
	}
	f.health, f.casterHP, f.frame, f.active = f.target.HealthData.Cur, f.caster.HealthData.Cur, noxServer.Frame(), true
	if f.mode == "npc-animated" {
		f.caster.SetDir(server.DirFromVec(f.target.PosVec.Sub(f.caster.PosVec)))
		f.caster.MonsterPushAction(ai.ACTION_CAST_SPELL_ON_OBJECT, uint32(spell.SPELL_EARTHQUAKE), 0, f.target)
		return
	}
	// E2E runs before the Kind1 reset. The normal callback queue runs after
	// that reset so the real quake packet survives until the client update.
	// Do not synthesize/replay a packet or change the production reset.
	noxServer.TickCallback(func() {
		api := noxServer.noxScriptP()
		api.CastSpellLvl(nsp.Spell("SPELL_EARTHQUAKE"), f.level, api.toObj(f.caster), api.toObj(f.target))
	})
}

func (f *e2eEarthquakeFixture) observeSound(id sound.ID, kind int, caster *server.Object, _ types.Pointf) {
	if !f.active || id != sound.SoundEarthquakeCast || caster != f.caster {
		return
	}
	f.castAudio++
	if kind != 0 || f.castAudio != 1 {
		e2eError(fmt.Errorf("Earthquake cast audio mismatch"))
		return
	}
	if f.mode == "npc-animated" {
		ud, head := f.caster.UpdateDataMonster(), f.caster.UpdateDataMonster().AIStackHead()
		if head == nil || head.Type() != ai.ACTION_CAST_SPELL_ON_OBJECT || head.ArgObj(2) != f.target ||
			head.ArgU32(0) != uint32(spell.SPELL_EARTHQUAKE) || ud.Field120_2 != 0 || uint32(ud.Field120_1) != ud.MonsterDef.MissileAttackFrame216 {
			e2eError(fmt.Errorf("Earthquake NPC bypassed natural animation/cast-frame"))
			return
		}
		f.natural = true
		e2eLog.Printf("EARTHQUAKE NPC ANIMATION: animation=%d elapsed=%d target=%p", ud.Field120_1, noxServer.Frame()-f.frame, head.ArgObj(2))
	}
	e2eLog.Printf("EARTHQUAKE CAST: mode=%s level=%d caster=%p target=%p owner=%p source=%v target-position=%v audio=1",
		f.mode, f.level, f.caster, f.target, f.caster.FindOwnerChainPlayer(), f.caster.PosVec, f.target.PosVec)
}

func (f *e2eEarthquakeFixture) observeTick() {
	if !f.active {
		return
	}
	if !e2eObjectInWorld(f.target) || f.target.HealthData == nil || f.caster.HealthData.Cur != f.casterHP {
		e2eError(fmt.Errorf("Earthquake target disappeared or caster received unexpected self damage"))
		return
	}
	if !f.hit && e2eEarthquakeDamageObserved(f.health, f.target.HealthData.Cur, f.target.Obj130, f.caster, f.target.Field131) && f.target.Frame134 >= f.frame {
		f.hit = true
		e2eLog.Printf("EARTHQUAKE HIT: mode=%s level=%d attribution=%p type=%d HP=%d->%d elapsed=%d",
			f.mode, f.level, f.target.Obj130, f.target.Field131, f.health, f.target.HealthData.Cur, noxServer.Frame()-f.frame)
	}
	if !f.quake && noxClient.Viewport().Jiggle12 != 0 {
		f.quake = true
		e2eLog.Printf("EARTHQUAKE CLIENT JIGGLE: mode=%s level=%d value=%d elapsed=%d", f.mode, f.level, noxClient.Viewport().Jiggle12, noxServer.Frame()-f.frame)
	}
}

func (f *e2eEarthquakeFixture) clientHit() bool {
	if !f.hit {
		if age := noxServer.Frame() - f.frame; age == 100 || age == 200 || age == 300 {
			e2eLog.Printf("EARTHQUAKE DAMAGE STATE: mode=%s level=%d age=%d HP=%d->%d attribution=%p type=%d audio=%d quake=%t natural-NPC=%t",
				f.mode, f.level, age, f.health, f.target.HealthData.Cur, f.target.Obj130, f.target.Field131, f.castAudio, f.quake, f.natural)
		}
		return false
	}
	if f.target == f.host {
		meter, ready := e2eClientHUDMeter(0)
		return ready && meter.Current == uint32(f.target.HealthData.Cur) && meter.Current < uint32(f.health)
	}
	drawable := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target)))
	if drawable == nil {
		return false
	}
	delta, ok := legacy.HealthChangeForDrawable(drawable.NetCode32)
	return ok && delta < 0
}

func (f *e2eEarthquakeFixture) complete() bool {
	if !f.clientHit() || !f.quake || f.castAudio != 1 || noxClient.Viewport().Jiggle12 != 0 ||
		f.mode == "npc-animated" && (!f.natural || f.caster.MonsterActionIsScheduled(ai.ACTION_CAST_SPELL_ON_OBJECT)) {
		return false
	}
	f.verified = true
	e2eLog.Printf("EARTHQUAKE COMPLETE: mode=%s level=%d HP=%d->%d type=11 audio=1 quake=received/settled natural-NPC=%t",
		f.mode, f.level, f.health, f.target.HealthData.Cur, f.natural)
	return true
}

func (f *e2eEarthquakeFixture) cleanup() {
	if !f.verified {
		e2eError(fmt.Errorf("Earthquake cleanup requires verified real damage and quake results"))
		return
	}
	f.active = false
	if f.mode == "npc-animated" {
		noxServer.DelayedDelete(f.caster)
		asObjectS(f.host).SetMaxHealth(int(f.hostMax))
		asObjectS(f.host).SetHealth(int(f.hostHP))
	} else {
		noxServer.DelayedDelete(f.target)
	}
	asObjectS(f.host).SetPos(f.original)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
}

// Real stock acceptance and NPC animation must produce IMPACT HP damage,
// client health replay and a real quake packet that naturally settles.
// No player incantation/mana or autonomous NPC spell choice is simulated.
func (sc *e2eScenario) CheckEarthquakeSpell(level int, mode, name string) {
	if !e2eEarthquakeMode(level, mode) {
		e2eError(fmt.Errorf("invalid Earthquake mode/level %s/%d", mode, level))
		return
	}
	f := &e2eEarthquakeFixture{level: level, mode: mode}
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		host := noxServer.Players.HostUnit()
		return host != nil && host.Buffs == 0 && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, f.prepare)
	sc.Wait(12, name+" publish units")
	sc.add(0, name+" cast", f.beginCast)
	sc.addWhen(1, name+" actual damage and client replay", 300, f.clientHit, func() {})
	sc.CaptureMagicFrame(name + " actual damage frame")
	sc.addWhen(1, name+" real quake packet and natural settling", 300, f.complete, func() {})
	sc.add(0, name+" cleanup", f.cleanup)
	sc.Wait(3, name+" cleanup complete")
}
