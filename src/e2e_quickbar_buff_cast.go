package opennox

import (
	"fmt"
	"image"
	"unsafe"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func e2eQuickbarBuffSpec(id spell.ID) (server.EnchantID, string, bool) {
	switch id {
	case spell.SPELL_HASTE:
		return server.ENCHANT_HASTED, "HasteEnchantDuration", true
	case spell.SPELL_PROTECTION_FROM_FIRE:
		return server.ENCHANT_PROTECT_FROM_FIRE, "ProtectFireEnchantDuration", true
	case spell.SPELL_PROTECTION_FROM_POISON:
		return server.ENCHANT_PROTECT_FROM_POISON, "ProtectPoisonEnchantDuration", true
	}
	return 0, "", false
}

type e2eQuickbarBuffCast struct {
	id                       spell.ID
	buff                     server.EnchantID
	duration, cost, slot     int
	host, npc, control       *server.Object
	original                 types.Pointf
	entries                  [25][2]uint32
	row                      int
	active, self             bool
	queueSeen, phonemeSeen   bool
	maxProgress              uint8
	debits                   int
	beforeMana, previousMana uint16
	start, applied           uint32
	hostHP, npcHP, controlHP uint16
	selfApplied              uint32
	selfDuration             int
}

func (f *e2eQuickbarBuffCast) prepare(key string) {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.HealthData == nil || f.host.UpdateData == nil || f.host.Buffs != 0 ||
		f.host.ControllingPlayer() == nil || f.host.ControllingPlayer().SpellLvl[f.id] != 1 || magicEntityHead != nil {
		e2eError(fmt.Errorf("quickbar buff fixture needs an unenchanted live player with the real level-one award"))
		return
	}
	var ok bool
	f.entries, f.row, ok = legacy.ClientQuickbarSnapshot()
	if !ok || f.row < 0 || f.row >= 5 {
		e2eError(fmt.Errorf("quickbar buff selected row unavailable: %d", f.row))
		return
	}
	f.slot = -1
	for slot := 0; slot < 5; slot++ {
		if f.entries[f.row*5+slot][0] == uint32(f.id) {
			f.slot = slot
		}
	}
	f.duration = int(noxServer.Balance.Float(key))
	f.cost = noxServer.Spells.ManaCost(f.id, 1)
	if !ok || f.slot < 0 || f.entries[f.row*5+f.slot][1]&1 != 1 || f.duration <= 120 || f.duration > 6000 || f.cost <= 0 {
		e2eError(fmt.Errorf("quickbar buff stock setup invalid: id=%s row=%d slot=%d duration=%d cost=%d", f.id, f.row, f.slot, f.duration, f.cost))
		return
	}
	f.original = f.host.PosVec
	origin, direction, err := e2eWarriorAbilityArena(f.original, f.host.Shape.Circle.R+16,
		func(from, to types.Pointf) bool { return e2eWarriorLaneClear(f.host, from, to) })
	if err != nil {
		e2eError(err)
		return
	}
	asObjectS(f.host).SetPos(origin)
	for index, pos := range []types.Pointf{
		origin.Add(direction.Mul(100)), origin.Add(types.Ptf(-direction.Y, direction.X).Mul(120)),
	} {
		unit := noxServer.NewObjectByTypeID("NPC")
		if unit == nil {
			e2eError(fmt.Errorf("quickbar buff fixture has no stock NPC"))
			return
		}
		noxServer.CreateObjectAt(unit, nil, pos)
		noxServer.ObjectsAddPending()
		if unit.HealthData == nil || unit.UpdateData == nil || unit.Buffs != 0 || !unit.SubClass().AsMonster().Has(object.MonsterNPC) {
			e2eError(fmt.Errorf("quickbar buff NPC is not initialized"))
			return
		}
		// Only the starting fixture uses ordinary waiting AI. Updates, damage,
		// networking and enchantment timers are never disabled.
		unit.UpdateDataMonster().SetAggression(0)
		unit.ClearActionStack()
		unit.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+20000)
		if index == 0 {
			f.npc = unit
		} else {
			f.control = unit
		}
	}
	f.hostHP, f.npcHP, f.controlHP = f.host.HealthData.Cur, f.npc.HealthData.Cur, f.control.HealthData.Cur
	noxServer.TickHook(f.observe)
	e2eLog.Printf("QUICKBAR BUFF PREPARED: spell=%s class=%s slot=%d native_player=%p native_npc=%p duration=%d stock_cost=%d real_award=true", f.id,
		f.host.ControllingPlayer().PlayerClass(), f.slot+1, f.host, f.npc, f.duration, f.cost)
}

func (f *e2eQuickbarBuffCast) aim() {
	pos := noxClient.Viewport().ToScreenPos(image.Pt(int(f.npc.PosVec.X), int(f.npc.PosVec.Y)))
	e2eQueueInput(&seat.MouseMoveEvent{Pos: pos, Relative: false})
}

func (f *e2eQuickbarBuffCast) begin(self bool) {
	entries, row, ok := legacy.ClientQuickbarSnapshot()
	want := f.entries
	if !self {
		want[f.row*5+f.slot][1] ^= 1
	}
	if !ok || row != f.row || entries != want || f.host.UpdateDataPlayer().CursorObj != f.npc ||
		f.host.UpdateDataPlayer().ManaCur < uint16(f.cost) || magicEntityHead != nil {
		e2eError(fmt.Errorf("quickbar buff cast not ready: spell=%s self=%t row=%d cursor=%p want=%p mana=%d cost=%d entries=%v want=%v", f.id, self,
			row, f.host.UpdateDataPlayer().CursorObj, f.npc, f.host.UpdateDataPlayer().ManaCur, f.cost, entries, want))
		return
	}
	f.self, f.active = self, true
	f.queueSeen, f.phonemeSeen, f.maxProgress, f.debits, f.applied = false, false, 0, 0, 0
	f.start = noxServer.Frame()
	f.beforeMana = f.host.UpdateDataPlayer().ManaCur
	f.previousMana = f.beforeMana
	keys := [...]keybind.Key{keybind.KeyA, keybind.KeyS, keybind.KeyD, keybind.KeyF, keybind.KeyG}
	e2eQueueInput(&seat.KeyboardEvent{Key: keys[f.slot], Pressed: true})
	e2eLog.Printf("QUICKBAR BUFF INPUT: spell=%s self=%t cursor=%p mana=%d key=%v", f.id, self, f.npc, f.beforeMana, keys[f.slot])
}

func (f *e2eQuickbarBuffCast) observe() {
	if !f.active {
		return
	}
	for node := magicEntityHead; node != nil; node = node.Next52 {
		if node.Obj4 != f.host {
			continue
		}
		mode := uint32(0)
		if f.self {
			mode = 1
		}
		if node.Spells8 != ([5]int32{int32(f.id)}) || node.Field48 != mode || node.SpellInd28 != 0 || node.Field29 != 0 ||
			unsafe.Sizeof(uintptr(0)) > 4 && (uintptr(unsafe.Pointer(node)) <= 0xffffffff || uintptr(unsafe.Pointer(node.Obj4)) <= 0xffffffff || uintptr(unsafe.Pointer(node.Field32)) <= 0xffffffff) {
			e2eError(fmt.Errorf("quickbar buff queue payload/native pointers mismatch: spell=%s mode=%d node=%+v", f.id, mode, *node))
			return
		}
		f.queueSeen = true
		f.maxProgress = max(f.maxProgress, node.Field36)
		if node.Field36 != 0 && node.Field32 != noxServer.Spells.PhonemeTree() && f.host.UpdateDataPlayer().SpellPhonemeLeaf == node.Field32 {
			f.phonemeSeen = true
		}
	}
	mana := f.host.UpdateDataPlayer().ManaCur
	if mana < f.previousMana {
		if int(f.previousMana-mana) != f.cost || f.debits != 0 {
			e2eError(fmt.Errorf("quickbar buff mana debit mismatch: spell=%s mana=%d->%d stock_cost=%d previous_debits=%d", f.id, f.previousMana, mana, f.cost, f.debits))
			return
		}
		f.debits++
	}
	f.previousMana = mana
	target := f.npc
	if f.self {
		target = f.host
	}
	if f.applied == 0 && target.HasEnchant(f.buff) {
		f.applied = noxServer.Frame()
		if !f.queueSeen || !f.phonemeSeen || f.debits != 1 || target.EnchantPower(f.buff) != 1 ||
			target.EnchantDur(f.buff) < f.duration-1 || target.EnchantDur(f.buff) > f.duration ||
			f.host.ControllingPlayer().Obj3640 != target {
			e2eError(fmt.Errorf("quickbar buff did not follow the actual incantation/mana/target path: spell=%s self=%t queue=%t phonemes=%t debit=%d target=%p timer=%d power=%d final_target=%p", f.id, f.self,
				f.queueSeen, f.phonemeSeen, f.debits, target, target.EnchantDur(f.buff), target.EnchantPower(f.buff), f.host.ControllingPlayer().Obj3640))
			return
		}
		if f.self {
			f.selfApplied, f.selfDuration = f.applied, target.EnchantDur(f.buff)
		}
		e2eLog.Printf("QUICKBAR BUFF CAST VERIFIED: spell=%s self=%t target=%p cursor=%p phonemes=%d stock_mana_debit=%d timer=%d power=%d elapsed=%d", f.id,
			f.self, target, f.npc, f.maxProgress, f.cost, target.EnchantDur(f.buff), target.EnchantPower(f.buff), f.applied-f.start)
	}
}

func (f *e2eQuickbarBuffCast) check() {
	target := f.npc
	if f.self {
		target = f.host
	}
	entries, row, ok := legacy.ClientQuickbarSnapshot()
	want := f.entries
	if !f.self {
		want[f.row*5+f.slot][1] ^= 1
	}
	clientTarget := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(target)))
	clientControl := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.control)))
	hud, hudOK := e2eClientHUDMeter(1)
	if !ok || row != f.row || entries != want || f.applied == 0 || f.debits != 1 || magicEntityHead != nil ||
		clientTarget == nil || !clientTarget.HasEnchant(f.buff) || clientControl == nil || clientControl.HasEnchant(f.buff) ||
		f.control.HasEnchant(f.buff) || f.host.HealthData.Cur != f.hostHP || f.npc.HealthData.Cur != f.npcHP || f.control.HealthData.Cur != f.controlHP ||
		!hudOK || hud.Current != uint32(f.host.UpdateDataPlayer().ManaCur) || hud.Maximum != uint32(f.host.UpdateDataPlayer().ManaMax) ||
		memmap.Int32(0x587000, 133484) != int32(f.slot) {
		e2eError(fmt.Errorf("quickbar buff live effect/client HUD/preservation mismatch: spell=%s self=%t applied=%d debit=%d hud=%+v mana=%d target=%p control=%p", f.id, f.self,
			f.applied, f.debits, hud, f.host.UpdateDataPlayer().ManaCur, clientTarget, clientControl))
		return
	}
	if f.self && f.npc.HasEnchant(f.buff) || !f.self &&
		f.host.EnchantDur(f.buff) != f.selfDuration-int(noxServer.Frame()-f.selfApplied) {
		e2eError(fmt.Errorf("quickbar buff also affected/refreshed the unintended unit: spell=%s self=%t host_timer=%d npc_timer=%d", f.id, f.self,
			f.host.EnchantDur(f.buff), f.npc.EnchantDur(f.buff)))
		return
	}
	f.active = false
	e2eLog.Printf("QUICKBAR BUFF CLIENT VERIFIED: spell=%s self=%t HUD_mana=%d/%d untouched_control=true preserved_other_entries=true", f.id, f.self, hud.Current, hud.Maximum)
}

// The book award is an explicit fixture; its insertion and default bit are
// live. Both casts below enter via actual seat keyboard input and MSG_TRY_SPELL.
// Cursor selection, gesture queue, phoneme tree, mana debit, target selection,
// server/client buffs and expiry are read-only observations. No direct cast,
// mana refill, injected spell packet, target pointer or timer is supplied.
func (sc *e2eScenario) CheckQuickbarBuffCast(id spell.ID, name string) {
	buff, key, ok := e2eQuickbarBuffSpec(id)
	if !ok {
		panic(fmt.Sprintf("unsupported quickbar buff spell %d", id))
	}
	f := &e2eQuickbarBuffCast{id: id, buff: buff}
	sc.CheckBookRewardDefault(id, false, name+" natural default")
	sc.addWhen(0, name+" prepare ordinary waiting NPC fixtures", 1200, func() bool {
		host := noxServer.Players.HostUnit()
		return host != nil && host.Buffs == 0 && host.UpdateData != nil &&
			host.UpdateDataPlayer().ManaCur >= uint16(2*noxServer.Spells.ManaCost(id, 1))
	}, func() { f.prepare(key) })
	sc.Wait(5, name+" publish fixture positions")
	for _, self := range []bool{true, false} {
		label := fmt.Sprintf("%s self=%t", name, self)
		if !self {
			sc.add(0, label+" press real target nugget", func() {
				win := legacy.ClientQuickbarNugget(f.slot)
				if win == nil || win.GetFlags().IsHidden() {
					e2eError(fmt.Errorf("quickbar buff target nugget unavailable"))
					return
				}
				e2eQueueInput(&seat.MouseMoveEvent{Pos: win.GlobalPos().Add(win.Size().Div(2))}, &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
			})
			sc.Input(1, label+" release real target nugget", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
		}
		sc.add(1, label+" aim by seat mouse input", f.aim)
		sc.addWhen(0, label+" observe natural cursor selection", 120, func() bool {
			return f.host.UpdateDataPlayer().CursorObj == f.npc
		}, func() { f.begin(self) })
		sc.add(1, label+" release real spell shortcut", func() {
			keys := [...]keybind.Key{keybind.KeyA, keybind.KeyS, keybind.KeyD, keybind.KeyF, keybind.KeyG}
			e2eQueueInput(&seat.KeyboardEvent{Key: keys[f.slot], Pressed: false})
		})
		sc.addWhen(0, label+" observe incantation mana and client effect", 120, func() bool {
			target := f.npc
			if self {
				target = f.host
			}
			client := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(target)))
			hud, hudOK := e2eClientHUDMeter(1)
			return f.applied != 0 && magicEntityHead == nil && client != nil && client.HasEnchant(f.buff) && hudOK && hud.Current == uint32(f.host.UpdateDataPlayer().ManaCur)
		}, f.check)
		sc.Screen(label + " actual quickbar effect")
	}
	sc.addWhen(0, name+" observe natural buff expiry", 6200, func() bool {
		for _, unit := range []*server.Object{f.host, f.npc, f.control} {
			client := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(unit)))
			if unit.HasEnchant(f.buff) || client == nil || client.HasEnchant(f.buff) {
				return false
			}
		}
		return true
	}, func() {
		if f.host.EnchantDur(f.buff) != 0 || f.npc.EnchantDur(f.buff) != 0 || f.debits != 1 {
			e2eError(fmt.Errorf("quickbar buff expiry did not clear live timers"))
			return
		}
		e2eLog.Printf("QUICKBAR BUFF EXPIRED: spell=%s both_targets=true natural_server_client_expiry=true", id)
		// Retire only the NPC fixtures, after their actual timers have expired.
		noxServer.DelayedDelete(f.npc)
		noxServer.DelayedDelete(f.control)
		asObjectS(f.host).SetPos(f.original)
	})
	sc.Wait(5, name+" retire only fixture NPCs")
}
