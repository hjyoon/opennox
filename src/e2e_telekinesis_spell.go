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
	"github.com/opennox/opennox/v1/server"
)

func e2eTelekinesisMode(level int, mode string) bool {
	return mode == "player-script" && level >= 1 && level <= 5 || mode == "npc-animated" && level == 0
}

type e2eTelekinesisFixture struct {
	level, power                         int
	mode                                 string
	host, caster, hand                   *server.Object
	original                             types.Pointf
	aims                                 [2]types.Pointf
	cursorInputs                         [2]e2eTelekinesisCursorInput
	health, mana                         uint16
	wire, created, duration, previousAge uint32
	scriptID                             int32
	previousDuration                     int
	onSound, offSound                    sound.ID
	onAudio, offAudio, moves             int
	active, natural, removed             bool
}

func (f *e2eTelekinesisFixture) hands() []*server.Object {
	var out []*server.Object
	for unit, remaining := f.host.Field129, 4096; unit != nil; unit, remaining = unit.Field128, remaining-1 {
		if remaining == 0 {
			e2eError(fmt.Errorf("Telekinesis owned list did not terminate"))
			return nil
		}
		if typ := unit.ObjectTypeC(); typ != nil && typ.ID() == "TelekinesisHand" {
			out = append(out, unit)
		}
	}
	return out
}

func (f *e2eTelekinesisFixture) prepare() {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.ControllingPlayer() == nil || f.host.UpdateData == nil || f.host.HealthData == nil ||
		f.host.HealthData.Cur == 0 || f.host.Buffs != 0 || f.host.Flags().HasAny(object.FlagDead|object.FlagDestroyed|object.FlagNoUpdate) {
		e2eError(fmt.Errorf("Telekinesis requires a live unenchanted host"))
		return
	}
	f.original = f.host.PosVec
	origin, direction, err := e2eWarriorAbilityArena(f.original, f.host.Shape.Circle.R+16, func(from, to types.Pointf) bool { return e2eWarriorLaneClear(f.host, from, to) })
	if err != nil {
		e2eError(err)
		return
	}
	asObjectS(f.host).SetPos(origin)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	f.aims = [2]types.Pointf{origin.Add(direction.Mul(64)), origin.Add(direction.Mul(96))}
	f.caster, f.power = f.host, f.level
	if f.mode == "npc-animated" {
		f.caster = noxServer.NewObjectByTypeID("NPC")
		if f.caster == nil || f.caster.UpdateData == nil || f.caster.HealthData == nil {
			e2eError(fmt.Errorf("Telekinesis requires a stock NPC caster"))
			return
		}
		noxServer.CreateObjectAt(f.caster, nil, origin.Sub(direction.Mul(80)))
		f.caster.UpdateDataMonster().SetAggression(0)
		f.caster.ClearActionStack()
		f.caster.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
		f.power = int(noxServer.Server.SpellPower4FE7B0(spell.SPELL_TELEKINESIS, f.caster))
	}
	noxServer.ObjectsAddPending()
	f.duration = 20 * noxServer.TickRate()
	if f.duration < 100 || f.duration > 1800 || len(f.hands()) != 0 {
		e2eError(fmt.Errorf("Telekinesis stock duration/owned baseline is invalid"))
		return
	}
	for _, pointer := range []unsafe.Pointer{unsafe.Pointer(f.host), unsafe.Pointer(f.caster), f.host.UpdateData, unsafe.Pointer(f.host.ControllingPlayer())} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
			e2eError(fmt.Errorf("Telekinesis native fixture below 4 GiB: %p", pointer))
			return
		}
	}
	f.onSound = noxServer.Spells.DefByInd(spell.SPELL_TELEKINESIS).GetOnSound()
	f.offSound = noxServer.Spells.DefByInd(spell.SPELL_TELEKINESIS).GetOffSound()
	if f.onSound == 0 || f.offSound == 0 || f.onSound == f.offSound {
		e2eError(fmt.Errorf("Telekinesis stock on/off audio is missing"))
		return
	}
	e2eQueueInput(&seat.MouseMoveEvent{Pos: noxClient.Inp.GetMousePos()})
	noxServer.Audio.OnSound(f.observeSound)
	noxServer.TickHook(f.observeTick)
}

func (f *e2eTelekinesisFixture) beginCast() {
	f.health, f.mana = f.host.HealthData.Cur, f.host.UpdateDataPlayer().ManaCur
	f.active = true
	if f.mode == "npc-animated" {
		f.caster.SetDir(server.DirFromVec(f.host.PosVec.Sub(f.caster.PosVec)))
		f.caster.MonsterPushAction(ai.ACTION_CAST_SPELL_ON_OBJECT, uint32(spell.SPELL_TELEKINESIS), 0, f.host)
		return
	}
	// Use the normal post-reset callback queue, not a supplied buff/report
	// packet. The next normal client update consumes the real cast report.
	noxServer.TickCallback(func() {
		api := noxServer.noxScriptP()
		api.CastSpellLvl(nsp.Spell("SPELL_TELEKINESIS"), f.level, api.toObj(f.host), api.toObj(f.host))
		noxServer.ObjectsAddPending()
		f.observeCreation()
	})
}

func (f *e2eTelekinesisFixture) observeCreation() {
	if f.hand != nil {
		return
	}
	hands := f.hands()
	if len(hands) != 1 {
		e2eError(fmt.Errorf("Telekinesis created %d owned hands", len(hands)))
		return
	}
	hand := hands[0]
	if hand.ObjOwner != f.host || hand.Field32 != noxServer.Frame() || hand.PosVec != f.host.PosVec ||
		!f.host.HasEnchant(server.ENCHANT_TELEKINESIS) || f.host.EnchantDur(server.ENCHANT_TELEKINESIS) != int(f.duration) ||
		f.host.EnchantPower(server.ENCHANT_TELEKINESIS) != f.power || f.host.HealthData.Cur != f.health || f.host.UpdateDataPlayer().ManaCur != f.mana {
		e2eError(fmt.Errorf("Telekinesis creation owner/position/buff/timer/power/HP/mana mismatch"))
		return
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(hand)) <= math.MaxUint32 {
		e2eError(fmt.Errorf("Telekinesis hand below 4 GiB"))
		return
	}
	f.hand, f.wire, f.scriptID, f.created = hand, hand.NetCode, hand.ScriptIDVal, hand.Field32
	f.previousDuration = int(f.duration)
	e2eLog.Printf("TELEKINESIS CAST: mode=%s level=%d power=%d caster=%p hand=%p owner=%p timer=%d created=%d position=%v", f.mode, f.level, f.power, f.caster, hand, hand.ObjOwner, f.duration, f.created, hand.PosVec)
}

func (f *e2eTelekinesisFixture) observeSound(id sound.ID, kind int, target *server.Object, _ types.Pointf) {
	if !f.active {
		return
	}
	// A targeted NPC cast releases Magic at the animation gate. The target's
	// on-sound occurs later on impact, not during that release animation.
	if f.mode == "npc-animated" && id == noxServer.Spells.DefByInd(spell.SPELL_TELEKINESIS).GetCastSound() && target == f.caster {
		ud, head := f.caster.UpdateDataMonster(), f.caster.UpdateDataMonster().AIStackHead()
		if kind != 0 || f.natural || head == nil || head.Type() != ai.ACTION_CAST_SPELL_ON_OBJECT || head.ArgObj(2) != f.host ||
			head.ArgU32(0) != uint32(spell.SPELL_TELEKINESIS) || ud.MonsterDef == nil || ud.Field120_2 != 0 ||
			uint32(ud.Field120_1) != ud.MonsterDef.MissileAttackFrame216 {
			e2eError(fmt.Errorf("Telekinesis NPC bypassed natural animation/cast frame"))
			return
		}
		projectiles := 0
		for obj, remaining := f.caster.Field129, 4096; obj != nil; obj, remaining = obj.Field128, remaining-1 {
			if remaining == 0 {
				e2eError(fmt.Errorf("Telekinesis NPC projectile list did not terminate"))
				return
			}
			if typ := obj.ObjectTypeC(); typ == nil || typ.ID() != "Magic" || obj.Flags().Has(object.FlagDestroyed) {
				continue
			}
			data := obj.UpdateDataSpellProjectile()
			if data == nil || data.Spell12 != uint32(spell.SPELL_TELEKINESIS) {
				continue
			}
			if obj.ObjOwner != f.caster || obj.Field32 != noxServer.Frame() || data.Target != f.host ||
				data.Field0 != f.caster || data.Field8 != f.caster || data.Level16 != uint32(f.power) ||
				unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 || uintptr(obj.UpdateData) <= math.MaxUint32) {
				e2eError(fmt.Errorf("Telekinesis NPC projectile lost native source/target/power"))
				return
			}
			projectiles++
			e2eLog.Printf("TELEKINESIS NPC RELEASE: animation=%d frame=%d projectile=%p update=%p caster=%p target=%p power=%d", ud.Field120_1, noxServer.Frame(), obj, obj.UpdateData, f.caster, data.Target, data.Level16)
		}
		if projectiles != 1 {
			e2eError(fmt.Errorf("Telekinesis NPC released %d real Magic projectiles", projectiles))
			return
		}
		f.natural = true
		return
	}
	if id != f.onSound && id != f.offSound {
		return
	}
	if kind != 0 || target != f.host {
		e2eError(fmt.Errorf("Telekinesis audio lost the original target owner"))
		return
	}
	if id == f.onSound {
		f.onAudio++
		if f.mode == "npc-animated" && f.onAudio == 1 {
			if !f.natural {
				e2eError(fmt.Errorf("Telekinesis NPC effect preceded its real projectile release"))
				return
			}
			f.observeCreation()
			e2eLog.Printf("TELEKINESIS NPC IMPACT: frame=%d target=%p power=%d", noxServer.Frame(), target, f.host.EnchantPower(server.ENCHANT_TELEKINESIS))
		}
	} else {
		f.offAudio++
	}
}

func (f *e2eTelekinesisFixture) observeTick() {
	if !f.active || f.hand == nil || f.removed {
		return
	}
	// RunTickHooks observes the frame after IncFrame. Never dereference the
	// hand after identity-checked removal from the live object list.
	age := noxServer.Frame() - 1 - f.created
	if e2eFistInWorld(f.hand, f.wire, f.scriptID) {
		duration := f.host.EnchantDur(server.ENCHANT_TELEKINESIS)
		destroyed := f.hand.Flags().Has(object.FlagDestroyed)
		if age > f.duration+1 || age <= f.duration && destroyed ||
			age == f.duration+1 && (!destroyed || f.hand.DeletedAt != f.created+age || f.previousAge != f.duration) ||
			duration > f.previousDuration || f.previousDuration-duration > 1 {
			e2eError(fmt.Errorf("Telekinesis hand/timer changed before the original lifetime boundary: age=%d timer=%d->%d", age, f.previousDuration, duration))
			return
		}
		if age == f.duration+1 {
			e2eLog.Printf("TELEKINESIS DELETE QUEUED: mode=%s level=%d age=%d destroyed=%t deletedAt=%d", f.mode, f.level, age, destroyed, f.hand.DeletedAt)
		}
		f.previousAge, f.previousDuration = age, duration
	} else {
		// Original 004E5E20 retains a current-frame deletion and finalizes it
		// on the following tick. The strict hand expiry is a separate event.
		if age != f.duration+2 || f.previousAge != f.duration+1 {
			e2eError(fmt.Errorf("Telekinesis hand removed at age=%d after age=%d, want queued=%d finalized=%d", age, f.previousAge, f.duration+1, f.duration+2))
			return
		}
		f.removed = true
		e2eLog.Printf("TELEKINESIS EXPIRED: mode=%s level=%d age=%d timer=%d", f.mode, f.level, age, f.host.EnchantDur(server.ENCHANT_TELEKINESIS))
	}
}

func (f *e2eTelekinesisFixture) queueCursor(index int) {
	want := image.Pt(int(f.aims[index].X), int(f.aims[index].Y))
	mouse := noxClient.Viewport().ToScreenPos(want)
	if !mouse.In(noxClient.Viewport().Screen) {
		e2eError(fmt.Errorf("Telekinesis cursor aim is outside the viewport"))
		return
	}
	f.cursorInputs[index] = e2eTelekinesisCursorInput{canvas: mouse, previousWorld: noxClient.netPrevMouse}
	e2eQueueInput(&seat.MouseMoveEvent{Pos: mouse})
}

func (f *e2eTelekinesisFixture) cursorMoved(index int) bool {
	if f.hand != nil && !f.removed && e2eFistInWorld(f.hand, f.wire, f.scriptID) {
		age := noxServer.Frame() - f.created
		if age == 30 || age == 60 || age == 120 {
			drawable := noxClient.Objs.ByNetCode(uint16(f.wire))
			var drawPos image.Point
			if drawable != nil {
				drawPos = drawable.Pos()
			}
			host := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.host)))
			e2eLog.Printf("TELEKINESIS CURSOR STATE: age=%d sample=%d aim=%v cursor=%v mouse=%v position=%v drawable=%p drawPos=%v host=%p clientBuff=%t timer=%d power=%d HP=%d mana=%d on/off=%d/%d", age, index+1, f.aims[index], f.host.ControllingPlayer().CursorVec, noxClient.Inp.GetMousePos(), f.hand.PosVec, drawable, drawPos, host, host != nil && host.HasEnchant(server.ENCHANT_TELEKINESIS), f.host.EnchantDur(server.ENCHANT_TELEKINESIS), f.host.EnchantPower(server.ENCHANT_TELEKINESIS), f.host.HealthData.Cur, f.host.UpdateDataPlayer().ManaCur, f.onAudio, f.offAudio)
		}
	}
	if f.hand == nil || f.removed || !e2eFistInWorld(f.hand, f.wire, f.scriptID) || f.onAudio != 2 {
		return false
	}
	handDrawable := noxClient.Objs.ByNetCode(uint16(f.wire))
	hostDrawable := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.host)))
	// Stock TelekinesisHand has CLASS NULL and no draw record. Its position
	// follows the actual client MSG_MOUSE world point, which can differ from
	// the queued aim when the normal camera scrolls before consuming input.
	return e2eTelekinesisCursorMatches(f.cursorInputs[index], noxClient.Inp.GetMousePos(), noxClient.netPrevMouse, f.host.ControllingPlayer().CursorVec, f.hand.PosVec) &&
		f.hand.ObjClass == 0 && handDrawable == nil && hostDrawable != nil && hostDrawable.HasEnchant(server.ENCHANT_TELEKINESIS) &&
		f.host.HasEnchant(server.ENCHANT_TELEKINESIS) && f.host.EnchantPower(server.ENCHANT_TELEKINESIS) == f.power &&
		f.host.HealthData.Cur == f.health && f.host.UpdateDataPlayer().ManaCur == f.mana
}

func (f *e2eTelekinesisFixture) complete() bool {
	hostDrawable := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.host)))
	if !f.removed || f.moves != 2 || f.onAudio != 2 || f.offAudio != 2 || len(f.hands()) != 0 || noxClient.Objs.ByNetCode(uint16(f.wire)) != nil ||
		hostDrawable == nil || hostDrawable.HasEnchant(server.ENCHANT_TELEKINESIS) || f.host.HasEnchant(server.ENCHANT_TELEKINESIS) ||
		f.host.EnchantDur(server.ENCHANT_TELEKINESIS) != 0 || f.host.EnchantPower(server.ENCHANT_TELEKINESIS) != 0 ||
		f.host.HealthData.Cur != f.health || f.host.UpdateDataPlayer().ManaCur != f.mana ||
		f.mode == "npc-animated" && (!f.natural || f.caster.MonsterActionIsScheduled(ai.ACTION_CAST_SPELL_ON_OBJECT)) {
		return false
	}
	e2eLog.Printf("TELEKINESIS COMPLETE: mode=%s level=%d moves=2 on/off=2/2 HP=%d mana=%d unchanged world/owned/drawable/buff=removed", f.mode, f.level, f.health, f.mana)
	return true
}

func (f *e2eTelekinesisFixture) cleanup() {
	f.active = false
	if f.mode == "npc-animated" {
		noxServer.DelayedDelete(f.caster)
	}
	asObjectS(f.host).SetPos(f.original)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
}

// Setup moves units and supplies an ordinary idle NPC action. The stock cast,
// native hand/update, cursor packets, buff, audio and strict natural expiry are
// observed, never supplied. This does not simulate incantation/mana spending or
// autonomous NPC spell selection; the hand belongs to the player target.
func (sc *e2eScenario) CheckTelekinesisSpell(level int, mode, name string) {
	if !e2eTelekinesisMode(level, mode) {
		e2eError(fmt.Errorf("invalid Telekinesis mode/level %s/%d", mode, level))
		return
	}
	f := &e2eTelekinesisFixture{level: level, mode: mode}
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		host := noxServer.Players.HostUnit()
		return host != nil && host.Buffs == 0 && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, f.prepare)
	sc.Wait(12, name+" publish units")
	sc.add(0, name+" cast", f.beginCast)
	for index := range f.aims {
		sc.add(1, fmt.Sprintf("%s cursor input %d", name, index+1), func() { f.queueCursor(index) })
		sc.addWhen(1, fmt.Sprintf("%s real cursor movement %d", name, index+1), 180, func() bool { return f.cursorMoved(index) }, func() {
			f.moves++
			e2eLog.Printf("TELEKINESIS CURSOR: mode=%s level=%d sample=%d hand=%p cursor=%v position=%v client=matched", f.mode, f.level, index+1, f.hand, f.host.ControllingPlayer().CursorVec, f.hand.PosVec)
		})
		sc.CaptureMagicFrame(fmt.Sprintf("%s actual cursor frame %d", name, index+1))
	}
	sc.addWhen(1, name+" natural expiry", 2000, f.complete, f.cleanup)
	sc.Wait(3, name+" cleanup complete")
}
