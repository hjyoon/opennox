package opennox

import (
	"fmt"
	"image"
	"math"
	"strings"
	"unsafe"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/noximage"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	nsp "github.com/opennox/noxscript/ns/v4/spell"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func e2eMutualStatusMode(mode string) (kind string, id spell.ID, fromNPC, ok bool) {
	kind, direction, found := strings.Cut(mode, "/")
	fromNPC, ok = e2eFistUnitDirection(direction)
	if !found || !ok {
		return kind, 0, fromNPC, false
	}
	switch kind {
	case "confused":
		id = spell.SPELL_CONFUSE
	case "stun":
		id = spell.SPELL_STUN
	case "slow":
		id = spell.SPELL_SLOW
	case "freeze":
		id = spell.SPELL_FREEZE
	case "blind":
		id = spell.SPELL_BLIND
	default:
		return kind, 0, fromNPC, false
	}
	return kind, id, fromNPC, true
}

// Expected mechanics are independent of the live caster's result: Stun on a
// Warrior or a monster heavier than 15 slows; other units are held. Confuse
// and Stun spill/round their balance durations; Slow truncates instead.
func e2eMutualStatusExpected(kind string, isPlayer, warrior bool, mass float32, balance float64, tickRate int) (server.EnchantID, uint16) {
	switch kind {
	case "confused":
		return server.ENCHANT_CONFUSED, uint16(int16(int32(math.RoundToEven(float64(float32(balance))))))
	case "stun":
		buff := server.ENCHANT_HELD
		if isPlayer && warrior || !isPlayer && mass > 15 {
			buff = server.ENCHANT_SLOWED
		}
		return buff, uint16(int16(int32(math.RoundToEven(float64(float32(balance))))))
	case "slow":
		return server.ENCHANT_SLOWED, uint16(int(balance))
	case "freeze":
		return server.ENCHANT_FREEZE, uint16(4 * tickRate)
	case "blind":
		return server.ENCHANT_BLINDED, uint16(4 * tickRate)
	default:
		return 0, 0
	}
}

type e2eMutualStatusFixture struct {
	kind                                                     string
	id                                                       spell.ID
	fromNPC, targeted, active, applied, natural, keepPressed bool
	host, npc, caster, target, magic                         *server.Object
	original                                                 types.Pointf
	health, casterHealth, duration                           uint16
	buff                                                     server.EnchantID
	power                                                    int
	frame, appliedFrame, keepFrame, magicWire                uint32
	magicScript                                              int32
	castSound, onSound, offSound                             sound.ID
	castCount, onCount, offCount                             int
	visual                                                   *e2ePlayerStatusAnimation
}

func (f *e2eMutualStatusFixture) prepare() {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.HealthData == nil || f.host.HealthData.Cur == 0 || f.host.Buffs != 0 {
		e2eError(fmt.Errorf("status spell needs a live unenchanted host"))
		return
	}
	f.original = f.host.PosVec
	origin, dir, err := e2eWarriorAbilityArena(f.original, f.host.Shape.Circle.R+16, func(from, to types.Pointf) bool { return e2eWarriorLaneClear(f.host, from, to) })
	if err != nil {
		e2eError(err)
		return
	}
	f.npc = noxServer.NewObjectByTypeID("NPC")
	if f.npc == nil {
		e2eError(fmt.Errorf("status spell has no stock NPC"))
		return
	}
	asObjectS(f.host).SetPos(origin)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	// Stay in the open half of the verified lane so a wall's visibility mask
	// cannot hide the NPC's front-effect sprite from the pixel observer.
	noxServer.CreateObjectAt(f.npc, nil, origin.Add(dir.Mul(64)))
	noxServer.ObjectsAddPending()
	if !e2eObjectInWorld(f.npc) || f.npc.UpdateData == nil || f.npc.HealthData == nil || f.npc.UpdateDataMonster().MonsterDef == nil || f.npc.Buffs != 0 {
		e2eError(fmt.Errorf("status NPC initialization failed"))
		return
	}
	f.npc.UpdateDataMonster().SetAggression(0)
	f.npc.ClearActionStack()
	f.npc.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
	f.caster, f.target = f.host, f.npc
	if f.fromNPC {
		f.caster, f.target = f.npc, f.host
	}
	key := ""
	switch f.kind {
	case "confused":
		key = "ConfuseEnchantDuration"
	case "stun":
		key = "StunEnchantDuration"
	case "slow":
		key = "SlowEnchantDuration"
	}
	balance := float64(0)
	if key != "" {
		balance = noxServer.Balance.Float(key)
	}
	isPlayer := f.target.Class().Has(object.ClassPlayer)
	warrior := isPlayer && f.target.ControllingPlayer().PlayerClass() == player.Warrior
	f.buff, f.duration = e2eMutualStatusExpected(f.kind, isPlayer, warrior, f.target.Mass, balance, int(noxServer.TickRate()))
	if f.duration <= 24 || f.duration > 2900 {
		e2eError(fmt.Errorf("status duration outside bounded stock test: %s/%d", f.kind, f.duration))
		return
	}
	f.targeted = noxServer.Spells.HasFlags(f.id, things.SpellTargeted)
	f.power = 3
	if f.fromNPC || f.targeted {
		f.power = int(noxServer.Server.SpellPower4FE7B0(f.id, f.caster))
	}
	f.castSound = noxServer.Spells.DefByInd(f.id).GetAudio(0)
	f.onSound = noxServer.Spells.DefByInd(f.buff.Spell()).GetAudio(1)
	f.offSound = noxServer.Spells.DefByInd(f.buff.Spell()).GetAudio(2)
	for _, p := range []unsafe.Pointer{unsafe.Pointer(f.host), unsafe.Pointer(f.npc), f.host.UpdateData, f.npc.UpdateData} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
			e2eError(fmt.Errorf("status native pointer below 4 GiB: %p", p))
			return
		}
	}
	noxServer.Audio.OnSound(func(id sound.ID, kind int, owner *server.Object, _ types.Pointf) {
		if !f.active {
			return
		}
		if id == f.castSound && id != 0 && owner == f.caster {
			f.castCount++
			if kind != 0 || f.castCount != 1 {
				e2eError(fmt.Errorf("status cast sound kind/count=%d/%d", kind, f.castCount))
				return
			}
			if f.targeted {
				f.observeProjectile()
			}
			if f.fromNPC {
				f.observeNatural()
			}
		}
		if id == f.onSound && id != 0 && owner == f.target {
			f.onCount++
			f.observeEffect()
			if kind != 0 || f.onCount != 1 {
				e2eError(fmt.Errorf("status on sound kind/count=%d/%d", kind, f.onCount))
			}
		}
		if id == f.offSound && id != 0 && owner == f.target {
			f.offCount++
			if kind != 0 || f.offCount != 1 {
				e2eError(fmt.Errorf("status off sound kind/count=%d/%d", kind, f.offCount))
			}
		}
	})
	noxServer.TickHook(func() {
		if f.active && !f.applied && f.target.HasEnchant(f.buff) {
			f.observeEffect()
		}
	})
}

func (f *e2eMutualStatusFixture) beginCast() {
	dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target)))
	if dr == nil || dr.Buffs != 0 || f.target.Buffs != 0 || f.caster.Buffs != 0 || !noxServer.MapTraceRay(f.caster.PosVec, f.target.PosVec, 0) {
		e2eError(fmt.Errorf("status target was not published/clear"))
		return
	}
	f.health, f.casterHealth = f.target.HealthData.Cur, f.caster.HealthData.Cur
	f.frame, f.keepFrame, f.active = noxServer.Frame(), noxServer.Frame(), true
	if f.buff == server.ENCHANT_CONFUSED || f.buff == server.ENCHANT_HELD {
		ref := legacy.AsImageRefP(*memmap.PtrPtr(0x5D4594, 1096456))
		if ref == nil || ref.Kind() != 2 || len(ref.Field24ptr().Images()) < 2 {
			e2eError(fmt.Errorf("status birdies native animation cache missing"))
			return
		}
		f.visual = &e2ePlayerStatusAnimation{kind: f.kind, buff: f.buff, unit: f.target, ref: ref, first: -1}
	}
	if f.fromNPC {
		f.npc.MonsterPushAction(ai.ACTION_CAST_SPELL_ON_OBJECT, uint32(f.id), 0, f.target)
		return
	}
	api := noxServer.noxScriptP()
	api.CastSpellLvl(nsp.Spell(f.id.String()), 3, api.toObj(f.caster), api.toObj(f.target))
}

func (f *e2eMutualStatusFixture) observeProjectile() {
	for obj := f.caster.Field129; obj != nil; obj = obj.Field128 {
		if obj.ObjectTypeC().ID() != "Magic" || obj.Flags().Has(object.FlagDestroyed) {
			continue
		}
		ud := obj.UpdateDataSpellProjectile()
		if ud == nil || ud.Spell12 != uint32(f.id) {
			continue
		}
		if f.magic != nil || ud.Target != f.target || ud.Field0 != f.caster || ud.Field8 != f.caster || int(ud.Level16) != f.power ||
			unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 || uintptr(obj.UpdateData) <= math.MaxUint32) {
			e2eError(fmt.Errorf("status Magic lost target/native owner/power: %s", f.kind))
			return
		}
		f.magic, f.magicWire, f.magicScript = obj, obj.NetCode, obj.ScriptIDVal
		e2eLog.Printf("MUTUAL STATUS PROJECTILE: kind=%s from-NPC=%t object=%p target=%p actual-power=%d", f.kind, f.fromNPC, obj, ud.Target, f.power)
	}
	if f.magic == nil {
		e2eError(fmt.Errorf("status targeted cast produced no Magic: %s", f.kind))
	}
}

func (f *e2eMutualStatusFixture) observeNatural() {
	ud, head := f.npc.UpdateDataMonster(), f.npc.UpdateDataMonster().AIStackHead()
	if head == nil || head.Type() != ai.ACTION_CAST_SPELL_ON_OBJECT || head.ArgObj(2) != f.target || head.ArgU32(0) != uint32(f.id) ||
		ud.Field120_2 != 0 || uint32(ud.Field120_1) != ud.MonsterDef.MissileAttackFrame216 {
		e2eError(fmt.Errorf("status NPC bypassed natural animation/cast frame"))
		return
	}
	f.natural = true
}

func (f *e2eMutualStatusFixture) observeEffect() {
	if f.applied {
		return
	}
	dur := f.target.EnchantDur(f.buff)
	if f.target.Buffs != uint32(1)<<uint(f.buff) || f.caster.Buffs != 0 || f.target.EnchantPower(f.buff) != f.power ||
		(dur != int(f.duration) && dur != int(f.duration)-1) || f.target.HealthData.Cur != f.health || f.caster.HealthData.Cur != f.casterHealth {
		e2eError(fmt.Errorf("status applied to wrong unit/timer/power: %s buffs=%#x/%#x timer=%d/%d power=%d/%d", f.kind, f.target.Buffs, f.caster.Buffs, dur, f.duration, f.target.EnchantPower(f.buff), f.power))
		return
	}
	f.applied, f.appliedFrame = true, noxServer.Frame()
	e2eLog.Printf("MUTUAL STATUS APPLIED: kind=%s from-NPC=%t caster=%p target=%p enchant=%d duration=%d actual-power=%d", f.kind, f.fromNPC, f.caster, f.target, f.buff, f.duration, f.power)
}

func (f *e2eMutualStatusFixture) sample(label string) {
	dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target)))
	cr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.caster)))
	mask := uint32(1) << uint(f.buff)
	if dr == nil || cr == nil || f.target.Buffs != mask || dr.Buffs != mask || f.caster.Buffs != 0 || cr.Buffs != 0 ||
		f.fromNPC && memmap.Uint32(0x5D4594, 1062540) != mask {
		e2eError(fmt.Errorf("mutual status server/client/HUD target mismatch"))
		return
	}
	if f.visual != nil {
		index, matched, total := e2eMutualStatusBirdiesPixels(f.visual.ref, dr)
		if total < 10 || matched*100 < total*80 || label == "advanced" && index == f.visual.first {
			e2eError(fmt.Errorf("mutual status animation missing/still: kind=%s sample=%s frame=%d matched=%d/%d", f.kind, label, index, matched, total))
			return
		}
		if label == "first" {
			f.visual.first = index
		}
		e2eLog.Printf("MUTUAL STATUS PIXELS: kind=%s from-NPC=%t sample=%s image-frame=%d matched=%d/%d", f.kind, f.fromNPC, label, index, matched, total)
	}
	if f.buff == server.ENCHANT_SLOWED {
		count, typ := 0, noxClient.Things.IndByID("YellowBubbleParticle")
		for _, p := range noxClient.Objs.AllList1() {
			delta := p.Pos().Sub(dr.Pos())
			if int(p.TypeIDVal) == typ && delta.X*delta.X+delta.Y*delta.Y <= 64*64 && noxClient.Viewport().ToScreenPos(p.Pos()).In(noxClient.Viewport().Screen) {
				count++
			}
		}
		if count == 0 {
			e2eError(fmt.Errorf("slowed target has no nearby real particles"))
			return
		}
		e2eLog.Printf("MUTUAL STATUS PARTICLES: kind=%s from-NPC=%t sample=%s nearby=%d", f.kind, f.fromNPC, label, count)
	}
	e2eLog.Printf("MUTUAL STATUS CLIENT: kind=%s from-NPC=%t sample=%s buffs=%#x HP=unchanged", f.kind, f.fromNPC, label, mask)
}

func (f *e2eMutualStatusFixture) complete() bool {
	// Ordinary input keeps the long Slow timer test within the inactivity limit.
	if f.keepPressed {
		e2eQueueInput(&seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: false})
		f.keepPressed = false
	} else if noxServer.Frame()-f.keepFrame >= 300 {
		pos := noxClient.Viewport().ToScreenPos(image.Pt(int(f.host.PosVec.X)+16, int(f.host.PosVec.Y)))
		e2eQueueInput(&seat.MouseMoveEvent{Pos: pos}, &seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: true})
		f.keepFrame, f.keepPressed = noxServer.Frame(), true
	}
	if !f.applied || noxServer.Frame() < f.appliedFrame+uint32(f.duration)+8 {
		return false
	}
	dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target)))
	cr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.caster)))
	if dr == nil || cr == nil || f.target.Buffs != 0 || dr.Buffs != 0 || f.caster.Buffs != 0 || cr.Buffs != 0 || f.target.EnchantDur(f.buff) != 0 ||
		f.targeted && (f.magic == nil || e2eFistInWorld(f.magic, f.magicWire, f.magicScript) || noxClient.Objs.ByNetCode(uint16(f.magicWire)) != nil) {
		return false
	}
	if f.castSound != 0 && f.castCount != 1 || f.onSound != 0 && f.onCount != 1 || f.offSound != 0 && f.offCount != 1 ||
		f.fromNPC && f.castSound != 0 && !f.natural || f.fromNPC && f.npc.MonsterActionIsScheduled(ai.ACTION_CAST_SPELL_ON_OBJECT) {
		return false
	}
	if f.target.HealthData.Cur != f.health || f.caster.HealthData.Cur != f.casterHealth || f.fromNPC && memmap.Uint32(0x5D4594, 1062540) != 0 {
		e2eError(fmt.Errorf("expired mutual status changed HP or remains in HUD"))
		return true
	}
	if f.visual != nil {
		_, matched, total := e2eMutualStatusBirdiesPixels(f.visual.ref, dr)
		if total >= 10 && matched*100 >= total*80 {
			e2eError(fmt.Errorf("status sprite remains after expiry"))
			return true
		}
	}
	e2eLog.Printf("MUTUAL STATUS COMPLETE: kind=%s from-NPC=%t server/client=expired target-only=true HP=unchanged cast/on/off=%d/%d/%d natural-NPC=%t", f.kind, f.fromNPC, f.castCount, f.onCount, f.offCount, f.natural)
	return true
}

// A moving NPC/camera invalidates the pre-cast background used by a translucent
// reference sprite. Render each stock frame over black AND white in a separate
// renderer: equal output pixels are opaque and independent of that background.
// The live buffer is only read. Retain the same minimum and 80% match threshold.
func e2eMutualStatusBirdiesPixels(ref *legacy.ImageRef, dr *client.Drawable) (best, matched, total int) {
	live := noxClient.r.PixBuffer()
	render := *noxClient.r.NoxRender
	data := *render.Data()
	render.SetData(&data)
	render.HookImageDrawXxx = nil
	data.SetAlphaEnabled(false)
	data.SetMultiply14(0)
	data.SetColorize17(0)
	data.SetClip(true)
	data.SetClipRect(live.Rect)
	pos := noxClient.Viewport().ToScreenPos(dr.Pos())
	pos = pos.Add(image.Pt(-64, 5-int(int16(dr.ZVal2))-dr.Z()-int(dr.GetZSizeMax())-64))
	if dr.Class().Has(object.ClassPlayer) && dr.AnimInd == 6 {
		off := 8 * uintptr(dr.AnimDir)
		pos = pos.Add(image.Pt(int(memmap.Int32(0x587000, 149432+off)), int(memmap.Int32(0x587000, 149436+off))))
	}
	rect := image.Rect(pos.X, pos.Y, pos.X+128, pos.Y+128).Intersect(live.Rect)
	for i, handle := range ref.Field24ptr().Images() {
		black, white := noximage.NewImage16(live.Rect), noximage.NewImage16(live.Rect)
		for j := range white.Pix {
			white.Pix[j] = 0xffff
		}
		img := render.Bag.AsImage(handle)
		render.SetPixBuffer(black)
		render.DrawImage16(img, pos)
		render.SetPixBuffer(white)
		render.DrawImage16(img, pos)
		m, n := e2eMutualStatusOpaquePixels(black, white, live, rect)
		if i == 0 || n > 0 && m*total > matched*n {
			best, matched, total = i, m, n
		}
	}
	return best, matched, total
}

func e2eMutualStatusOpaquePixels(black, white, live *noximage.Image16, rect image.Rectangle) (matched, total int) {
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			b, w := black.Pix[black.PixOffset(x, y)], white.Pix[white.PixOffset(x, y)]
			if b == w {
				total++
				if live.Pix[live.PixOffset(x, y)] == b {
					matched++
				}
			}
		}
	}
	return matched, total
}

func (f *e2eMutualStatusFixture) cleanup() {
	f.active = false
	noxServer.DelayedDelete(f.npc)
	asObjectS(f.host).SetPos(f.original)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
}

// Only placement and waiting AI are setup. The regular spell selector, owned
// Magic projectile (when targeted), animation-frame AI, buff packets, pixels
// and natural expiry remain real. This does not simulate player mana/input.
func (sc *e2eScenario) CheckMutualStatusSpell(mode, name string) {
	kind, id, fromNPC, ok := e2eMutualStatusMode(mode)
	if !ok {
		e2eError(fmt.Errorf("invalid mutual status mode %q", mode))
		return
	}
	f := &e2eMutualStatusFixture{kind: kind, id: id, fromNPC: fromNPC}
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		host := noxServer.Players.HostUnit()
		return host != nil && host.Buffs == 0 && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, f.prepare)
	sc.Wait(12, name+" publish units")
	sc.add(0, name+" cast", f.beginCast)
	for _, sample := range []struct {
		label string
		dt    uint32
	}{{"first", 12}, {"advanced", 24}} {
		sc.addWhen(0, name+" "+sample.label+" visible", 300, func() bool {
			dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target)))
			return f.applied && noxServer.Frame() >= f.appliedFrame+sample.dt && dr != nil && dr.HasEnchant(f.buff)
		}, func() { f.sample(sample.label) })
	}
	sc.addWhen(1, name+" natural expiry", 3000, f.complete, f.cleanup)
	sc.Wait(3, name+" cleanup")
}
