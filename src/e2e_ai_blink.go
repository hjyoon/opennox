package opennox

import (
	"fmt"
	"math"
	"strings"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func e2eAIBlinkMode(mode string) (kind string, descending, ok bool) {
	kind, slope, found := strings.Cut(mode, "/")
	if !found || (kind != "Troll" && kind != "Urchin" && kind != "NPC") {
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

func e2eAIBlinkDuration(record *server.DurSpell, target *server.Object, frame uint32) bool {
	return record != nil && target != nil && record.Spell == uint32(spell.SPELL_BLINK) &&
		record.Obj12 == target && record.Caster16 == target && record.Flag20 == 0 && record.Obj24 == nil && record.Target48 == target &&
		record.Frame68-1 == frame && record.Create == legacy.Get_nox_xxx_spellBlink2_530310() &&
		record.Update == legacy.Get_nox_xxx_spellBlink1_530380()
}

// Observe only live native owner links. The wake can still be pending when
// Blink emits its first sound, before its normal teleport call.
func e2eAIBlinkWake(target *server.Object) *server.Object {
	if target == nil {
		return nil
	}
	obj := target.Field129
	for count := 0; obj != nil && count < 4096; count++ {
		if obj.ObjOwner == target && !obj.Flags().Has(object.FlagDestroyed) &&
			obj.CollideData != nil && obj.ObjectTypeC().ID() == "TeleportWake" {
			return obj
		}
		obj = obj.Field128
	}
	return nil
}

type e2eAIBlinkFixture struct {
	mode                             string
	combat                           e2eAIFirstAttackFixture
	update                           *server.MonsterUpdateData
	targetWire                       uint32
	targetScript                     int32
	armedFrame, oldDeadline          uint32
	castFrame, deadline, wakeExpiry  uint32
	lastLog                          uint32
	castSound                        sound.ID
	soundCount                       int
	durationSeen, castSeen, wakeSeen bool
	active, teleported, completed    bool
	origin, destination              types.Pointf
}

func (f *e2eAIBlinkFixture) configure() {
	target, host := f.combat.enemy, f.combat.host
	if !f.combat.hit || !f.combat.acquired || !f.combat.fought || target.Buffs != 0 ||
		!e2eObjectInWorld(target) || noxflags.HasGame(noxflags.GameModeCoop|noxflags.GameModeQuest) {
		e2eError(fmt.Errorf("MainAI Blink requires real first combat in an ordinary game"))
		return
	}
	f.update = target.UpdateDataMonster()
	f.targetWire, f.targetScript = target.NetCode, target.ScriptIDVal
	for _, ptr := range []unsafe.Pointer{unsafe.Pointer(target), target.UpdateData, unsafe.Pointer(host), host.UpdateData} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
			e2eError(fmt.Errorf("MainAI Blink object/update is not above 4 GiB: %p", ptr))
			return
		}
	}
	f.armedFrame, f.oldDeadline = noxServer.Frame(), f.update.Field371
	if f.oldDeadline > f.armedFrame || f.update.CurrentEnemy != host {
		e2eError(fmt.Errorf("MainAI Blink needs a naturally acquired enemy and an initially ready slot"))
		return
	}
	// This is a configurable Blink-capable stock body, not a claim that stock
	// Troll/Urchin/NPC definitions have Blink enabled. Change ability properties
	// only; never select an enemy, inject an action/path/cast, move either unit,
	// reset recent-move timers, or adjust duration/teleport output.
	f.update.StatusFlags |= object.MonStatusCanCastSpells
	f.update.Field376 = 1
	f.update.Field370_0, f.update.Field370_2 = 300, 300
	f.update.FleeRange = float32(server.ObjectDistance4E6C00(target, host)*2 + 64)
	f.castSound = noxServer.Spells.DefByInd(spell.SPELL_BLINK).GetCastSound()
	f.active = true
	noxServer.Audio.OnSound(f.observeSound)
	noxServer.TickHook(f.tick)
	e2eLog.Printf("AI MAIN BLINK CONFIGURED: mode=%s native=%p/%p frame=%d range=%g original-deadline=%d first-attack=natural selected-enemy=natural cast-input=none recent-move-timer=untouched", f.mode, target, f.update, f.armedFrame, f.update.FleeRange, f.oldDeadline)
}

func (f *e2eAIBlinkFixture) observeSound(id sound.ID, _ int, owner *server.Object, _ types.Pointf) {
	if !f.active || owner != f.combat.enemy || id != f.castSound {
		return
	}
	f.soundCount++
	for record := noxServer.Spells.Dur.List; record != nil; record = record.Next {
		if !e2eAIBlinkDuration(record, owner, noxServer.Frame()) {
			continue
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(record)) <= math.MaxUint32 {
			e2eError(fmt.Errorf("MainAI Blink duration is not a native high pointer"))
			return
		}
		if !f.durationSeen {
			f.durationSeen, f.castFrame, f.origin = true, noxServer.Frame(), owner.PosVec
			e2eLog.Printf("AI MAIN BLINK DURATION: mode=%s record=%p power=%d caster=target trigger-frame=%d origin=%v", f.mode, record, record.Level, f.castFrame, f.origin)
		}
	}
}

func (f *e2eAIBlinkFixture) tick() {
	if !f.active {
		return
	}
	target := f.combat.enemy
	if !e2eObjectInWorld(target) || target.NetCode != f.targetWire || target.ScriptIDVal != f.targetScript ||
		target.UpdateData != unsafe.Pointer(f.update) || target.HealthData.Cur != f.combat.enemyHP ||
		target.Obj130 != nil || f.update.StatusFlags.Has(object.MonStatusInjured) {
		e2eError(fmt.Errorf("MainAI Blink target identity/record/untouched health changed: %s", f.mode))
		return
	}
	if f.update.Field371 != f.oldDeadline && !f.castSeen {
		f.castSeen, f.deadline = true, f.update.Field371
		if f.deadline < f.armedFrame+300 || f.deadline > noxServer.Frame()+300 ||
			f.update.HasAction(ai.ACTION_CAST_SPELL_ON_OBJECT) || f.update.HasAction(ai.ACTION_CAST_SPELL_ON_LOCATION) ||
			f.update.HasAction(ai.ACTION_CAST_DURATION_SPELL) {
			e2eError(fmt.Errorf("MainAI Blink did not use the immediate service/cooldown: %s", f.mode))
			return
		}
	}
	if wake := e2eAIBlinkWake(target); wake != nil && !f.wakeSeen {
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(wake)) <= math.MaxUint32 || uintptr(wake.CollideData) <= math.MaxUint32) {
			e2eError(fmt.Errorf("MainAI Blink wake/collision data is not native-width"))
			return
		}
		f.wakeSeen, f.destination, f.wakeExpiry = true, (*server.TeleportWakeCollideData)(wake.CollideData).Destination, wake.Field34
		if !f.durationSeen || wake.PosVec != f.origin || f.destination.Sub(f.origin).Len() < 64 ||
			f.wakeExpiry != f.castFrame+noxServer.TickRate() {
			e2eError(fmt.Errorf("MainAI Blink wake has wrong origin/destination/lifetime: %s origin=%v/%v destination=%v expires=%d cast=%d", f.mode, f.origin, wake.PosVec, f.destination, f.wakeExpiry, f.castFrame))
			return
		}
	}
	if f.wakeSeen && target.PosVec.Sub(f.destination).Len() <= 16 && target.PosVec.Sub(f.origin).Len() >= 64 {
		f.teleported = true
	}
}

func (f *e2eAIBlinkFixture) observe() bool {
	if noxServer.Frame()-f.lastLog >= 30 {
		f.lastLog = noxServer.Frame()
		e2eLog.Printf("AI MAIN BLINK TICK: mode=%s elapsed=%d deadline=%d seen=%t/%t/%t teleport=%t sound=%d current=%p action=%v recent=%t pos=%v", f.mode, noxServer.Frame()-f.armedFrame, f.update.Field371, f.castSeen, f.durationSeen, f.wakeSeen, f.teleported, f.soundCount, f.update.CurrentEnemy, f.update.AIStackHead().Type(), noxServer.S().MonsterMoveAttemptRecent534810(f.combat.enemy), f.combat.enemy.PosVec)
	}
	if !f.castSeen || !f.durationSeen || !f.wakeSeen || !f.teleported || f.soundCount < 2 {
		return false
	}
	for record := noxServer.Spells.Dur.List; record != nil; record = record.Next {
		if record.Spell == uint32(spell.SPELL_BLINK) && record.Caster16 == f.combat.enemy {
			return false
		}
	}
	if f.deadline != f.castFrame+300 || f.update.Field371 != f.deadline {
		e2eError(fmt.Errorf("MainAI Blink cooldown did not remain tied to its real cast frame"))
		return true
	}
	path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
	if err != nil {
		e2eError(err)
		return true
	}
	f.completed, f.active = true, false
	e2eLog.Printf("AI MAIN BLINK PASS: mode=%s first-attack=natural native-object/update/duration/wake=true immediate=true cast-frame=%d deadline=%d origin=%v destination=%v actual=%v sounds=%d duration=removed path=%s", f.mode, f.castFrame, f.deadline, f.origin, f.destination, f.combat.enemy.PosVec, f.soundCount, path)
	return true
}

func (f *e2eAIBlinkFixture) cleanup() {
	if !f.completed {
		e2eError(fmt.Errorf("autonomous MainAI Blink did not complete: %s", f.mode))
		return
	}
	f.combat.cleanup()
}

func (sc *e2eScenario) CheckAIBlink(mode, name string) {
	kind, descending, ok := e2eAIBlinkMode(mode)
	if !ok {
		e2eError(fmt.Errorf("invalid MainAI Blink mode %q", mode))
		return
	}
	f := &e2eAIBlinkFixture{mode: mode, combat: e2eAIFirstAttackFixture{mode: mode, kind: kind, descending: descending}}
	sc.addWhen(0, name+" prepare untouched hostile AI", 1200, func() bool {
		unit := noxServer.Players.HostUnit()
		return nox_client_isConnected() && unit != nil && noxClient.ClientPlayerUnit() != nil && unit.Buffs == 0 && unit.Poison540 == 0
	}, f.combat.prepare)
	sc.addWhen(1, name+" natural first attack then configure Blink ability", 600, f.combat.observe, f.configure)
	sc.addWhen(1, name+" autonomous MainAI Blink and natural teleport", 600, f.observe, f.cleanup)
	sc.Wait(12, name+" ordinary deletion cleanup")
}
