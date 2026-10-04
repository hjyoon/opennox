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
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func e2eAIFearMode(mode string) (kind string, descending, ok bool) {
	kind, slope, found := strings.Cut(mode, "/")
	if !found {
		return "", false, false
	}
	switch kind {
	case "Spider", "Troll", "Urchin", "NPC":
	default:
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

func e2eAIFearDuration(balance float64) (uint16, bool) {
	if math.IsNaN(balance) || math.IsInf(balance, 0) || balance < 25 || balance >= 2901 {
		return 0, false
	}
	return uint16(int(balance)), true
}

// The stock non-Coop IMMUNE_FEAR gate must not be bypassed by a probe.
func e2eAIFearImmune(unit *server.Object, coop bool) bool {
	return unit != nil && !coop && unit.Class().Has(object.ClassMonster) &&
		unit.MonsterClass().Has(object.MonsterImmuneFear)
}

// Read the live native stack; a partially rejected prefix is not a FLEE.
func e2eAIFearScheduled(unit *server.Object) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) || !unit.HasEnchant(server.ENCHANT_AFRAID) {
		return false
	}
	data := unit.UpdateDataMonster()
	if data.AIStackInd < 1 || int(data.AIStackInd) >= len(data.AIStack) {
		return false
	}
	for i := 0; i < int(data.AIStackInd); i++ {
		if data.AIStack[i].Type() == ai.DEPENDENCY_IS_ENCHANTED &&
			data.AIStack[i].ArgU32(0) == uint32(server.ENCHANT_AFRAID) &&
			data.AIStack[i+1].Type() == ai.ACTION_FLEE {
			return true
		}
	}
	return false
}

func e2eAIFearAway(from, to, away types.Pointf) bool {
	delta := to.Sub(from)
	for _, value := range []float32{delta.X, delta.Y, away.X, away.Y} {
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			return false
		}
	}
	initial := math.Hypot(float64(away.X), float64(away.Y))
	separation := math.Hypot(float64(away.X+delta.X), float64(away.Y+delta.Y))
	return initial > 0 && delta.Len() >= 16 && separation >= initial+16
}

// A real player may need to re-enter sight after the full stock flee timer.
// Return only a short mouse aim point, not a replacement world position.
func e2eAIFearApproach(from, to types.Pointf) (types.Pointf, bool) {
	delta := to.Sub(from)
	if !e2eAIFearAway(from, to, delta) || delta.Len() <= 48 {
		return types.Pointf{}, false
	}
	return from.Add(delta.Normalize().Mul(128)), true
}

type e2eAIFearFixture struct {
	mode                         string
	combat                       e2eAIFirstAttackFixture
	startPos, away, clientStart  types.Pointf
	clientAway                   types.Pointf
	duration                     uint16
	power                        int
	castFrame, appliedFrame      uint32
	keepFrame, lastLog           uint32
	keepPressed                  bool
	playerRoute                  []types.Pointf
	playerRouteIndex             int
	expiredFrame, targetWire     uint32
	targetScript                 int32
	active, applied, moved       bool
	expired, resumed, hit        bool
	clientDamageCleared          bool
	expiredHP                    uint16
	targetHP                     uint16
	castSound, onSound, offSound sound.ID
	fleeSound                    sound.ID
	castCount, onCount, offCount int
	fleeCount                    int
	targeted                     bool
	immune, immuneVerified       bool
	magic                        *server.Object
	magicWire                    uint32
	magicScript                  int32
}

func (f *e2eAIFearFixture) prepare() {
	var ok bool
	f.duration, ok = e2eAIFearDuration(noxServer.Balance.Float("FearEnchantDuration"))
	if !ok {
		e2eError(fmt.Errorf("fear duration outside bounded stock observer"))
		return
	}
	f.combat.prepare()
	host, target := f.combat.host, f.combat.enemy
	for _, p := range []unsafe.Pointer{unsafe.Pointer(host), host.UpdateData, unsafe.Pointer(target), target.UpdateData} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
			e2eError(fmt.Errorf("fear object/update below 4 GiB: %p", p))
			return
		}
	}
	f.targetWire, f.targetScript = target.NetCode, target.ScriptIDVal
	f.immune = e2eAIFearImmune(target, noxflags.HasGame(noxflags.GameModeCoop))
	if f.immune != (f.combat.kind == "NPC") {
		e2eError(fmt.Errorf("stock fear immunity changed: mode=%s subclass=%x immune=%t", f.mode, uint32(target.ObjSubClass), f.immune))
		return
	}
	e2eLog.Printf("AI FEAR STOCK GATE: mode=%s subclass=%x non-Coop-immune=%t", f.mode, uint32(target.ObjSubClass), f.immune)
	f.targeted = noxServer.Spells.HasFlags(spell.SPELL_FEAR, things.SpellTargeted)
	f.power = 1
	if f.targeted {
		f.power = int(noxServer.Server.SpellPower4FE7B0(spell.SPELL_FEAR, host))
	}
	def := noxServer.Spells.DefByInd(spell.SPELL_FEAR)
	f.castSound, f.onSound, f.offSound = def.GetAudio(0), def.GetAudio(1), def.GetAudio(2)
	if ptr := target.UpdateDataMonster().SoundSet122; ptr != nil {
		f.fleeSound = sound.ID(*(*uint32)(unsafe.Add(ptr, 48)))
	}
	noxServer.Audio.OnSound(f.observeSound)
	noxServer.TickHook(f.tick)
}

func (f *e2eAIFearFixture) cast() {
	host, target := f.combat.host, f.combat.enemy
	dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(target)))
	if dr == nil || target.Buffs != 0 || host.Buffs != 0 || dr.Buffs != 0 ||
		!f.combat.hit || !f.combat.acquired || !f.combat.fought || target.UpdateDataMonster().HasAction(ai.ACTION_FLEE) {
		e2eError(fmt.Errorf("fear needs ordinary first combat and an unenchanted published target"))
		return
	}
	// Clear the real Spider's initial poison before the timed probe. Do not
	// heal, select an enemy, schedule FLEE or change the enchant duration.
	noxServer.RemovePoison4EE9D0(host)
	pos := dr.Pos()
	f.startPos, f.away, f.clientStart = target.PosVec, target.PosVec.Sub(host.PosVec), types.Ptf(float32(pos.X), float32(pos.Y))
	playerPos := noxClient.ClientPlayerUnit().Pos()
	f.clientAway = f.clientStart.Sub(types.Ptf(float32(playerPos.X), float32(playerPos.Y)))
	f.targetHP = target.HealthData.Cur
	f.castFrame, f.active = noxServer.Frame(), true
	f.keepFrame = f.castFrame
	api := noxServer.noxScriptP()
	api.CastSpellLvl(nsp.Spell("SPELL_FEAR"), 1, api.toObj(host), api.toObj(target))
	e2eLog.Printf("AI FEAR CAST: mode=%s target=%p update=%p stock-duration=%d expected-power=%d targeted=%t", f.mode, target, target.UpdateData, f.duration, f.power, f.targeted)
}

func (f *e2eAIFearFixture) observeMagic() {
	for obj := f.combat.host.Field129; obj != nil; obj = obj.Field128 {
		if obj.ObjectTypeC().ID() != "Magic" || obj.Flags().Has(object.FlagDestroyed) {
			continue
		}
		data := obj.UpdateDataSpellProjectile()
		if data == nil || data.Spell12 != uint32(spell.SPELL_FEAR) {
			continue
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 || uintptr(obj.UpdateData) <= math.MaxUint32) {
			e2eError(fmt.Errorf("fear Magic object/update below 4 GiB: %p/%p", obj, obj.UpdateData))
			return
		}
		if f.magic != nil || data.Target != f.combat.enemy || data.Field0 != f.combat.host ||
			data.Field8 != f.combat.host || int(data.Level16) != f.power {
			e2eError(fmt.Errorf("fear Magic lost its ordinary target/owner/power"))
			return
		}
		f.magic, f.magicWire, f.magicScript = obj, obj.NetCode, obj.ScriptIDVal
	}
	if f.magic == nil {
		e2eError(fmt.Errorf("targeted fear produced no ordinary Magic"))
	}
}

func (f *e2eAIFearFixture) observeApplied() {
	if f.immune {
		e2eError(fmt.Errorf("stock immune NPC accepted fear: %s", f.mode))
		return
	}
	if f.applied {
		return
	}
	target := f.combat.enemy
	duration := target.EnchantDur(server.ENCHANT_AFRAID)
	if target.Buffs != uint32(1)<<uint(server.ENCHANT_AFRAID) || f.combat.host.Buffs != 0 ||
		target.EnchantPower(server.ENCHANT_AFRAID) != f.power ||
		(duration != int(f.duration) && duration != int(f.duration)-1) ||
		target.HealthData.Cur != f.targetHP {
		e2eError(fmt.Errorf("fear applied to wrong unit/timer/power: mode=%s buffs=%x timer=%d/%d power=%d/%d", f.mode, target.Buffs, duration, f.duration, target.EnchantPower(server.ENCHANT_AFRAID), f.power))
		return
	}
	f.applied, f.appliedFrame = true, noxServer.Frame()
	e2eLog.Printf("AI FEAR APPLIED: mode=%s target-only=true duration=%d power=%d HP=unchanged", f.mode, duration, f.power)
}

func (f *e2eAIFearFixture) observeSound(id sound.ID, kind int, owner *server.Object, _ types.Pointf) {
	if !f.active || id == 0 {
		return
	}
	if id == f.castSound && owner == f.combat.host {
		f.castCount++
		if kind != 0 || f.castCount != 1 {
			e2eError(fmt.Errorf("fear cast sound kind/count=%d/%d", kind, f.castCount))
		}
		if f.targeted {
			f.observeMagic()
		}
	}
	if owner != f.combat.enemy {
		return
	}
	if id == f.onSound {
		f.onCount++
		f.observeApplied()
		if kind != 0 || f.onCount != 1 {
			e2eError(fmt.Errorf("fear on sound kind/count=%d/%d", kind, f.onCount))
		}
	}
	if id == f.offSound {
		f.offCount++
		if kind != 0 || f.offCount != 1 {
			e2eError(fmt.Errorf("fear off sound kind/count=%d/%d", kind, f.offCount))
		}
	}
	if id == f.fleeSound {
		f.fleeCount++
		if kind != 0 {
			e2eError(fmt.Errorf("fear FLEE sound kind=%d", kind))
		}
	}
}

func (f *e2eAIFearFixture) tick() {
	if !f.active {
		return
	}
	target, host := f.combat.enemy, f.combat.host
	if _, ready := legacy.HealthChangeForDrawable(uint32(noxServer.GetUnitNetCode(host))); !ready {
		f.clientDamageCleared = true
	}
	if target.HasEnchant(server.ENCHANT_AFRAID) {
		f.observeApplied()
	}
	if !f.applied && noxServer.Frame()-f.lastLog >= 60 {
		f.lastLog = noxServer.Frame()
		live := f.magic != nil && e2eFistInWorld(f.magic, f.magicWire, f.magicScript)
		var magicPos types.Pointf
		if live {
			magicPos = f.magic.PosVec
		}
		e2eLog.Printf("AI FEAR PENDING: mode=%s since-cast=%d buffs=%x pos=%v->%v direction=%d magic=%p/%t/%v sound=%d/%d/%d/%d", f.mode, noxServer.Frame()-f.castFrame, target.Buffs, target.PosVec, host.PosVec, host.Direction1, f.magic, live, magicPos, f.castCount, f.onCount, f.offCount, f.fleeCount)
	}
	if f.applied && !f.moved && noxServer.Frame()-f.lastLog >= 60 {
		f.lastLog = noxServer.Frame()
		dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(target)))
		e2eLog.Printf("AI FEAR MOVE TICK: mode=%s since-cast=%d scheduled=%t pos=%v->%v actual-away=%t client=%p target-HP=%d", f.mode, noxServer.Frame()-f.castFrame, e2eAIFearScheduled(target), f.startPos, target.PosVec, e2eAIFearAway(f.startPos, target.PosVec, f.away), dr, target.HealthData.Cur)
	}
	if f.applied && !f.expired && !target.HasEnchant(server.ENCHANT_AFRAID) {
		elapsed := noxServer.Frame() - f.appliedFrame
		if elapsed+2 < uint32(f.duration) || elapsed > uint32(f.duration)+2 ||
			target.EnchantDur(server.ENCHANT_AFRAID) != 0 {
			e2eError(fmt.Errorf("fear cleared before/after its stock deadline: %s elapsed=%d duration=%d", f.mode, elapsed, f.duration))
			return
		}
		f.expired, f.expiredFrame, f.expiredHP = true, noxServer.Frame(), host.HealthData.Cur
		f.combat.projectiles = make(map[*server.Object]bool)
		e2eLog.Printf("AI FEAR EXPIRED: mode=%s elapsed=%d stock-duration=%d HP-checkpoint=%d", f.mode, elapsed, f.duration, f.expiredHP)
	}
	if !f.expired {
		return
	}
	data := target.UpdateDataMonster()
	if data.CurrentEnemy == host && !data.HasAction(ai.ACTION_FLEE) && data.HasAction(ai.ACTION_FIGHT) {
		f.resumed = true
	}
	// Only new, live enemy-owned missiles after expiry qualify for a return
	// hit. Stored pointers are identity keys, not reads of freed missiles.
	for _, missile := range noxServer.Objs.AllMissiles() {
		if missile.ObjOwner == target && missile.Field32 > f.expiredFrame &&
			!missile.Flags().Has(object.FlagDestroyed) {
			f.combat.projectiles[missile] = true
		}
	}
}

func (f *e2eAIFearFixture) moving() bool {
	target := f.combat.enemy
	dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(target)))
	if !f.applied || f.expired || !e2eAIFearScheduled(target) || dr == nil ||
		dr.Buffs != uint32(1)<<uint(server.ENCHANT_AFRAID) || target.HealthData.Cur != f.targetHP {
		return false
	}
	pos := dr.Pos()
	client := types.Ptf(float32(pos.X), float32(pos.Y))
	return e2eAIFearAway(f.startPos, target.PosVec, f.away) &&
		e2eAIFearAway(f.clientStart, client, f.clientAway)
}

func (f *e2eAIFearFixture) immuneResponse() bool {
	host, target := f.combat.host, f.combat.enemy
	dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(target)))
	data := target.UpdateDataMonster()
	return f.immune && !f.applied && noxServer.Frame()-f.castFrame >= 60 &&
		target.Buffs == 0 && target.EnchantDur(server.ENCHANT_AFRAID) == 0 && target.HealthData.Cur == f.targetHP &&
		dr != nil && dr.Buffs == 0 && data.CurrentEnemy == host && data.HasAction(ai.ACTION_FIGHT) && !data.HasAction(ai.ACTION_FLEE) &&
		f.castCount == 1 && f.onCount == 0 && f.offCount == 0 && f.fleeCount == 0 &&
		f.magic != nil && !e2eFistInWorld(f.magic, f.magicWire, f.magicScript) && noxClient.Objs.ByNetCode(uint16(f.magicWire)) == nil &&
		e2eAIFearFreshHit(host, host.HealthData.Max, f.castFrame) && f.combat.acceptsDamageSource(host.Obj130)
}

func (f *e2eAIFearFixture) response() bool {
	if f.immune {
		return f.immuneResponse()
	}
	return f.moving()
}

func (f *e2eAIFearFixture) captureResponse() {
	if !f.immune {
		f.captureMoving()
		return
	}
	if !f.immuneResponse() {
		e2eError(fmt.Errorf("stock fear immunity/combat/client state lost: %s", f.mode))
		return
	}
	path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
	if err != nil {
		e2eError(err)
		return
	}
	f.immuneVerified = true
	e2eLog.Printf("AI FEAR IMMUNE: mode=%s stock=IMMUNE_FEAR non-Coop=true duration=0 server/client=unenchanted Magic=normally-removed FIGHT=natural fresh-server-hit=%d HP=%d/%d cast/on/off/flee=%d/%d/%d/%d path=%s", f.mode, f.combat.host.Frame134, f.combat.host.HealthData.Cur, f.combat.host.HealthData.Max, f.castCount, f.onCount, f.offCount, f.fleeCount, path)
}

func (f *e2eAIFearFixture) captureMoving() {
	if !f.moving() {
		e2eError(fmt.Errorf("fear movement/client state lost"))
		return
	}
	path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
	if err != nil {
		e2eError(err)
		return
	}
	f.moved = true
	e2eLog.Printf("AI FEAR MOVED: mode=%s stack=natural-FEAR/FLEE server=%v->%v client=AFRAID away=true path=%s", f.mode, f.startPos, f.combat.enemy.PosVec, path)
}

func e2eAIFearFreshHit(host *server.Object, checkpoint uint16, after uint32) bool {
	return host != nil && host.HealthData != nil && host.HealthData.Cur < checkpoint && host.Frame134 > after && host.Obj130 != nil
}

func (f *e2eAIFearFixture) complete() bool {
	if f.immune {
		f.hit = f.immuneVerified
		return f.hit
	}
	host, target := f.combat.host, f.combat.enemy
	// Keep the bounded wait below ordinary inactivity/spectator limits using
	// real input. After expiry, walk into close visible proximity, even if the
	// NPC reacquires first. There is no attack input or AI/damage injection.
	if aim, walk := f.approach(); walk {
		pos := noxClient.Viewport().ToScreenPos(image.Pt(int(aim.X), int(aim.Y)))
		e2eQueueInput(&seat.MouseMoveEvent{Pos: pos}, &seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: true})
		f.keepFrame, f.keepPressed = noxServer.Frame(), true
	} else if f.keepPressed {
		e2eQueueInput(&seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: false})
		f.keepPressed = false
	} else if noxServer.Frame()-f.keepFrame >= 300 {
		pos := noxClient.Viewport().ToScreenPos(image.Pt(int(host.PosVec.X)+16, int(host.PosVec.Y)))
		e2eQueueInput(&seat.MouseMoveEvent{Pos: pos}, &seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: true})
		f.keepFrame, f.keepPressed = noxServer.Frame(), true
	}
	if noxServer.Frame()-f.lastLog >= 60 {
		f.lastLog = noxServer.Frame()
		data := target.UpdateDataMonster()
		delta, ready := legacy.HealthChangeForDrawable(uint32(noxServer.GetUnitNetCode(host)))
		e2eLog.Printf("AI FEAR TICK: mode=%s since-cast=%d expired=%t resumed=%t action=%v pos=%v->%v current=%p sight=%t HP=%d checkpoint=%d fresh=%t source=%p accepted=%t client=%t/%d cleared=%t sound=%d/%d/%d/%d", f.mode, noxServer.Frame()-f.castFrame, f.expired, f.resumed, data.AIStackHead().Type(), target.PosVec, host.PosVec, data.CurrentEnemy, noxServer.S().CanSee(target, host, 0), host.HealthData.Cur, f.expiredHP, e2eAIFearFreshHit(host, f.expiredHP, f.expiredFrame), host.Obj130, f.combat.acceptsDamageSource(host.Obj130), ready, delta, f.clientDamageCleared, f.castCount, f.onCount, f.offCount, f.fleeCount)
	}
	if !f.moved || !f.expired || !f.resumed || target.HasEnchant(server.ENCHANT_AFRAID) ||
		target.UpdateDataMonster().HasAction(ai.ACTION_FLEE) || !e2eAIFearFreshHit(host, f.expiredHP, f.expiredFrame) {
		return false
	}
	if !f.combat.acceptsDamageSource(host.Obj130) {
		return false
	}
	dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(target)))
	delta, ready := legacy.HealthChangeForDrawable(uint32(noxServer.GetUnitNetCode(host)))
	if dr == nil || dr.HasEnchant(server.ENCHANT_AFRAID) || !f.clientDamageCleared || !ready || delta >= 0 ||
		f.targeted && (f.magic == nil || e2eFistInWorld(f.magic, f.magicWire, f.magicScript) || noxClient.Objs.ByNetCode(uint16(f.magicWire)) != nil) {
		return false
	}
	if f.castSound != 0 && f.castCount != 1 || f.onSound != 0 && f.onCount != 1 ||
		f.offSound != 0 && f.offCount != 1 || f.fleeSound != 0 && f.fleeCount == 0 {
		return false
	}
	path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
	if err != nil {
		e2eError(err)
		return true
	}
	f.hit = true
	e2eLog.Printf("AI FEAR PASS: mode=%s expiry=%d resumed=natural-FIGHT fresh-HP=%d->%d hit-frame=%d old-client-damage=expired client-delta=%d cast/on/off/flee=%d/%d/%d/%d path=%s", f.mode, f.expiredFrame, f.expiredHP, host.HealthData.Cur, host.Frame134, delta, f.castCount, f.onCount, f.offCount, f.fleeCount, path)
	return true
}

func (f *e2eAIFearFixture) cleanup() {
	if !f.hit {
		e2eError(fmt.Errorf("fear did not return to ordinary combat"))
		return
	}
	f.active = false
	if f.keepPressed {
		e2eQueueInput(&seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: false})
		f.keepPressed = false
	}
	f.combat.cleanup()
}

func (f *e2eAIFearFixture) removed() bool {
	if e2eFistInWorld(f.combat.enemy, f.targetWire, f.targetScript) ||
		noxClient.Objs.ByNetCode(uint16(f.targetWire)) != nil {
		return false
	}
	e2eLog.Printf("AI FEAR REMOVED: mode=%s server/client=absent", f.mode)
	return true
}

// Stock placement/hostile NPC equipment, an ordinary FEAR request, and real
// non-attack movement input after expiry are supplied. No enemy, FLEE/FIGHT
// action, path, timer, damage, packet or rendered result is injected; the
// target must first attack autonomously and later reacquire the player. Stock
// non-Coop immune NPCs must instead reject FEAR and continue real combat.
func (sc *e2eScenario) CheckAIFear(mode, name string) {
	kind, descending, ok := e2eAIFearMode(mode)
	if !ok {
		e2eError(fmt.Errorf("invalid stock AI fear mode %q", mode))
		return
	}
	f := &e2eAIFearFixture{mode: mode, combat: e2eAIFirstAttackFixture{mode: mode + "/clear", kind: kind, descending: descending}}
	sc.addWhen(0, name+" prepare stock hostile", 1200, func() bool {
		host := noxServer.Players.HostUnit()
		return nox_client_isConnected() && host != nil && noxClient.ClientPlayerUnit() != nil && host.Buffs == 0 && host.Poison540 == 0
	}, f.prepare)
	sc.addWhen(1, name+" autonomous first attack then ordinary fear", 600, f.combat.observe, f.cast)
	sc.addWhen(1, name+" stock fear response and client state", 600, f.response, f.captureResponse)
	sc.addWhen(1, name+" stock result and continued combat", 3600, f.complete, f.cleanup)
	sc.addWhen(1, name+" ordinary deletion", 300, f.removed, func() {})
	sc.Wait(12, name+" fear cleanup")
}
