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
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	nsp "github.com/opennox/noxscript/ns/v4/spell"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

const (
	e2eAIRetreatSpell   = spell.SPELL_PROTECTION_FROM_FIRE
	e2eAIRetreatEnchant = server.ENCHANT_PROTECT_FROM_FIRE
)

func e2eAIRetreatMode(mode string) (kind string, descending, ok bool) {
	kind, slope, found := strings.Cut(mode, "/")
	if !found || (kind != "Troll" && kind != "NPC") {
		return "", false, false
	}
	switch slope {
	case "ascending":
		return kind, false, true
	case "descending":
		return kind, true, true
	default:
		return "", false, false
	}
}

// Read the real action and native target pointer; do not provide a cast result.
func e2eAIRetreatSelfCast(update *server.MonsterUpdateData, unit *server.Object) bool {
	if update == nil || unit == nil || !update.HasAction(ai.ACTION_RETREAT) {
		return false
	}
	head := update.AIStackHead()
	return head != nil && head.Type() == ai.ACTION_CAST_SPELL_ON_OBJECT &&
		head.ArgU32(0) == uint32(e2eAIRetreatSpell) && head.ArgObj(2) == unit
}

type e2eAIRetreatFixture struct {
	mode                     string
	combat                   e2eAIFirstAttackFixture
	update                   *server.MonsterUpdateData
	targetWire               uint32
	targetScript             int32
	armedFrame, oldDeadline  uint32
	queuedFrame, castFrame   uint32
	lastShot, lastLog        uint32
	shots, castSounds        int
	power                    int32
	duration                 int
	castSound                sound.ID
	missiles                 map[*server.Object]bool
	active, injured, queued  bool
	cast, applied, completed bool
}

func (f *e2eAIRetreatFixture) configure() {
	target, host := f.combat.enemy, f.combat.host
	if !f.combat.hit || !f.combat.acquired || !f.combat.fought || target.Buffs != 0 ||
		!e2eObjectInWorld(target) || noxflags.HasGame(noxflags.GameModeCoop|noxflags.GameModeQuest) ||
		!noxServer.Spells.HasFlags(spell.SPELL_HASTE, things.SpellMobsCanCast) {
		e2eError(fmt.Errorf("RETREAT probe requires natural first combat and stock mob-castable Haste"))
		return
	}
	f.update = target.UpdateDataMonster()
	f.targetWire, f.targetScript = target.NetCode, target.ScriptIDVal
	for _, ptr := range []unsafe.Pointer{unsafe.Pointer(target), target.UpdateData, unsafe.Pointer(host), host.UpdateData} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
			e2eError(fmt.Errorf("RETREAT object/update is not above 4 GiB: %p", ptr))
			return
		}
	}
	f.armedFrame, f.oldDeadline = noxServer.Frame(), f.update.Field371
	if f.oldDeadline > f.armedFrame || f.update.CurrentEnemy != host || target.HealthData.Cur != f.combat.enemyHP {
		e2eError(fmt.Errorf("RETREAT needs a naturally acquired enemy, ready slot and untouched health"))
		return
	}
	// Configure normal map/script ability and health-threshold properties on a
	// stock body. This does not claim stock Troll/NPC has Haste enabled. Do not
	// select an enemy, push an action, damage/heal a unit, reset a timer, change
	// spell definitions or write the buff/network/animation result.
	f.update.StatusFlags |= object.MonStatusCanCastSpells
	*(*uint32)(unsafe.Add(unsafe.Pointer(&f.update.Field372), 4*uintptr(spell.SPELL_HASTE))) |= 0x80000000
	f.update.Field370_0, f.update.Field370_2 = 300, 300
	asObjectS(target).SetRetreatLevel(0.98)
	asObjectS(target).SetRegroupLevel(1)
	f.power = noxServer.S().SpellPower4FE7B0(spell.SPELL_HASTE, target)
	f.duration = int(noxServer.Balance.Float("HasteEnchantDuration"))
	f.castSound = noxServer.Spells.DefByInd(spell.SPELL_HASTE).GetCastSound()
	if f.duration <= 0 || f.castSound == 0 {
		e2eError(fmt.Errorf("RETREAT stock Haste duration/sound unavailable"))
		return
	}
	f.missiles = make(map[*server.Object]bool)
	f.active = true
	noxServer.Audio.OnSound(f.observeSound)
	noxServer.TickHook(f.tick)
	e2eLog.Printf("AI RETREAT CONFIGURED: mode=%s native=%p/%p frame=%d original-deadline=%d first-attack=natural enemy=natural threshold=%g/%g spell-def=stock", f.mode, target, f.update, f.armedFrame, f.oldDeadline, f.update.RetreatLevel, f.update.ResumeLevel)
	f.fire()
}

func (f *e2eAIRetreatFixture) observeMissiles() {
	for _, missile := range noxServer.Objs.AllMissiles() {
		if missile.UpdateData == nil || missile.Flags().Has(object.FlagDestroyed) || missile.ObjOwner != f.combat.host ||
			!missile.Class().Has(object.ClassMissile) || !missile.SubClass().AsMissile().Has(object.MissileMagic) {
			continue
		}
		data := missile.UpdateDataMissile()
		if data.Owner == f.combat.host && data.Target == f.combat.enemy && data.SpellID == int32(spell.SPELL_MAGIC_MISSILE) {
			f.missiles[missile] = true
		}
	}
}

func (f *e2eAIRetreatFixture) fire() {
	if f.shots >= 12 || f.injured || f.queued || f.applied {
		return
	}
	// The player supplies an ordinary incoming projectile. The defender alone
	// selects RETREAT, its self-buff, native target and cast animation.
	host, target := f.combat.host, f.combat.enemy
	ud := host.UpdateDataPlayer()
	ud.CursorObj, ud.Player.Obj3640 = target, target
	ud.Field55, ud.Field56 = int(target.PosVec.X), int(target.PosVec.Y)
	ud.Player.CursorVec = image.Pt(int(target.PosVec.X), int(target.PosVec.Y))
	mouse := noxClient.Viewport().ToScreenPos(ud.Player.CursorVec)
	noxClient.ChangeMousePos(mouse, true)
	e2eQueueInput(&seat.MouseMoveEvent{Pos: mouse})
	host.SetDir(server.DirFromVec(target.PosVec.Sub(host.PosVec)))
	api := noxServer.noxScriptP()
	api.CastSpellLvl(nsp.Spell("SPELL_MAGIC_MISSILE"), 1, api.toObj(host), api.toObj(target))
	noxServer.ObjectsAddPending()
	f.shots++
	f.lastShot = noxServer.Frame()
	f.observeMissiles()
}

func (f *e2eAIRetreatFixture) observeQueue() {
	if f.queued || !e2eAIRetreatSelfCast(f.update, f.combat.enemy) {
		return
	}
	if !f.injured || f.update.Field371 == f.oldDeadline || f.update.Field371 < f.armedFrame+300 ||
		f.update.Field371 > noxServer.Frame()+300 {
		e2eError(fmt.Errorf("RETREAT self-cast lacks actual incoming damage/selector cooldown: %s", f.mode))
		return
	}
	f.queued, f.queuedFrame = true, f.update.Field371-300
	e2eLog.Printf("AI RETREAT SELF CAST: mode=%s queued-frame=%d deadline=%d target=%p HP=%d->%d native-self=%p action=RETREAT/CAST_OBJECT", f.mode, f.queuedFrame, f.update.Field371, f.combat.enemy, f.combat.enemyHP, f.combat.enemy.HealthData.Cur, f.update.AIStackHead().ArgObj(2))
}

func (f *e2eAIRetreatFixture) observeDamage() {
	target := f.combat.enemy
	if !f.injured && target.HealthData.Cur < f.combat.enemyHP {
		if !f.missiles[target.Obj130] || target.Frame134 < f.armedFrame {
			e2eError(fmt.Errorf("RETREAT incoming hit is not the observed player projectile: %s source=%p", f.mode, target.Obj130))
			return
		}
		f.injured = true
	}
}

func (f *e2eAIRetreatFixture) observeSound(id sound.ID, kind int, owner *server.Object, _ types.Pointf) {
	if !f.active || owner != f.combat.enemy || id != f.castSound {
		return
	}
	f.observeDamage()
	f.observeQueue()
	if !f.queued || kind != 0 || f.update.MonsterDef == nil || f.update.Field120_2 != 0 ||
		uint32(f.update.Field120_1) != f.update.MonsterDef.MissileAttackFrame216 {
		e2eError(fmt.Errorf("RETREAT Haste did not use the real self-cast animation: %s", f.mode))
		return
	}
	f.castSounds++
	if !f.cast {
		f.cast, f.castFrame = true, noxServer.Frame()
	}
}

func (f *e2eAIRetreatFixture) tick() {
	if !f.active {
		return
	}
	target := f.combat.enemy
	if !e2eObjectInWorld(target) || target.NetCode != f.targetWire || target.ScriptIDVal != f.targetScript ||
		target.UpdateData != unsafe.Pointer(f.update) || target.HealthData.Cur == 0 || f.combat.host.Buffs != 0 {
		e2eError(fmt.Errorf("RETREAT target identity/record/aliveness/self-buff ownership changed: %s", f.mode))
		return
	}
	f.observeMissiles()
	f.observeDamage()
	f.observeQueue()
	if target.HasEnchant(server.ENCHANT_HASTED) {
		if !f.queued || target.EnchantDur(server.ENCHANT_HASTED) <= 0 ||
			target.EnchantDur(server.ENCHANT_HASTED) > f.duration ||
			target.EnchantPower(server.ENCHANT_HASTED) != int(f.power) {
			e2eError(fmt.Errorf("RETREAT Haste has wrong natural source/duration/power: %s", f.mode))
			return
		}
		f.applied = true
	}
}

func (f *e2eAIRetreatFixture) observe() bool {
	f.tick()
	if !f.injured && noxServer.Frame()-f.lastShot >= 30 {
		f.fire()
	}
	if noxServer.Frame()-f.lastLog >= 30 {
		f.lastLog = noxServer.Frame()
		e2eLog.Printf("AI RETREAT TICK: mode=%s elapsed=%d shots=%d HP=%d/%d injured=%t queued=%t cast=%t applied=%t deadline=%d current=%p stack=%v", f.mode, noxServer.Frame()-f.armedFrame, f.shots, f.combat.enemy.HealthData.Cur, f.combat.enemyHP, f.injured, f.queued, f.cast, f.applied, f.update.Field371, f.update.CurrentEnemy, f.update.GetAIStack())
	}
	dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.combat.enemy)))
	if !f.injured || !f.queued || !f.cast || !f.applied || f.castSounds == 0 || dr == nil ||
		dr.Buffs&(uint32(1)<<uint(server.ENCHANT_HASTED)) == 0 {
		return false
	}
	if f.update.Field371 != f.queuedFrame+300 || f.castFrame < f.queuedFrame {
		e2eError(fmt.Errorf("RETREAT selector deadline was rewritten after the real cast"))
		return true
	}
	path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
	if err != nil {
		e2eError(err)
		return true
	}
	f.completed, f.active = true, false
	e2eLog.Printf("AI RETREAT PASS: mode=%s first-attack=natural incoming=player-projectile HP=%d->%d queued=%d cast=%d deadline=%d self-target=native buff=server/client power=%d duration=%d/%d sound=%d path=%s", f.mode, f.combat.enemyHP, f.combat.enemy.HealthData.Cur, f.queuedFrame, f.castFrame, f.update.Field371, f.combat.enemy.EnchantPower(server.ENCHANT_HASTED), f.combat.enemy.EnchantDur(server.ENCHANT_HASTED), f.duration, f.castSounds, path)
	return true
}

func (f *e2eAIRetreatFixture) cleanup() {
	if !f.completed {
		e2eError(fmt.Errorf("natural RETREAT self-buff did not complete: %s", f.mode))
		return
	}
	f.combat.cleanup()
}

func (sc *e2eScenario) CheckAIRetreat(mode, name string) {
	kind, descending, ok := e2eAIRetreatMode(mode)
	if !ok {
		e2eError(fmt.Errorf("invalid RETREAT self-buff mode %q", mode))
		return
	}
	f := &e2eAIRetreatFixture{mode: mode, combat: e2eAIFirstAttackFixture{mode: mode, kind: kind, descending: descending}}
	sc.addWhen(0, name+" prepare untouched hostile AI", 1200, func() bool {
		unit := noxServer.Players.HostUnit()
		return nox_client_isConnected() && unit != nil && noxClient.ClientPlayerUnit() != nil && unit.Buffs == 0 && unit.Poison540 == 0
	}, f.combat.prepare)
	sc.addWhen(1, name+" natural first attack then configure retreat ability", 600, f.combat.observe, f.configure)
	sc.addWhen(1, name+" natural incoming hit and RETREAT self-buff", 600, f.observe, f.cleanup)
	sc.Wait(12, name+" ordinary deletion cleanup")
}
