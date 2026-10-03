package opennox

import (
	"fmt"
	"image"
	"math"
	"unsafe"

	"github.com/opennox/libs/client/seat"
	noxcolor "github.com/opennox/libs/color"
	"github.com/opennox/libs/noximage"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	ns4 "github.com/opennox/noxscript/ns/v4"
	nsp "github.com/opennox/noxscript/ns/v4/spell"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func e2eShockMode(level int, mode string) bool {
	switch mode {
	case "player", "npc":
		return level >= 1 && level <= 5
	case "npc-natural", "glyph-to-npc", "glyph-to-player":
		return level == 0
	default:
		return false
	}
}

// The original stores a binary32 spill rounded to nearest/even in a WORD.
// Keep this expectation independent of the production conversion helper.
func e2eShockDuration(balance float64) (uint32, bool) {
	rounded := math.RoundToEven(float64(float32(balance)))
	if math.IsNaN(rounded) || math.IsInf(rounded, 0) || rounded < 25 || rounded > 2900 {
		return 0, false
	}
	return uint32(uint16(int16(rounded))), true
}

func e2eShockWhitePixel(pix *noximage.Image16, point image.Point, view image.Rectangle) bool {
	return point.X-10 >= view.Min.X && point.Y-10 >= view.Min.Y &&
		point.X+10 < view.Max.X && point.Y+10 < view.Max.Y && point.In(pix.Rect) &&
		pix.Pix[pix.PixOffset(point.X, point.Y)] == uint16(noxcolor.RGB5551Color(255, 255, 255))
}

type e2eShockFixture struct {
	mode                      string
	level                     int
	host, npc, target, caster *server.Object
	glyph                     *server.Object
	original                  types.Pointf
	glyphPos                  types.Pointf
	active, applied, hit      bool
	natural                   bool
	frame, appliedFrame       uint32
	duration, power           uint32
	glyphWire                 uint32
	glyphType                 uint16
	glyphScript               int32
	health                    uint16
	damage                    uint16
	onAudio, offAudio         int
	detonateAudio             int
	firstBirth                uint32
	keepaliveFrame            uint32
	keepalivePressed          bool
	walkFrom                  types.Pointf
}

func (f *e2eShockFixture) isGlyph() bool {
	return f.mode == "glyph-to-npc" || f.mode == "glyph-to-player"
}

func (f *e2eShockFixture) prepare() {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.HealthData == nil || f.host.HealthData.Cur <= 1 || f.host.Buffs != 0 ||
		f.host.Flags().HasAny(object.FlagDead|object.FlagDestroyed|object.FlagNoUpdate) {
		e2eError(fmt.Errorf("Shock requires a live unenchanted host"))
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
		e2eError(fmt.Errorf("Shock has no stock NPC type"))
		return
	}
	asObjectS(f.host).SetPos(origin)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	noxServer.CreateObjectAt(f.npc, nil, origin.Add(direction.Mul(112)))
	noxServer.ObjectsAddPending()
	if !e2eObjectInWorld(f.npc) || f.npc.UpdateData == nil || f.npc.HealthData == nil ||
		f.npc.HealthData.Cur <= 1 || f.npc.UpdateDataMonster().MonsterDef == nil ||
		!f.npc.SubClass().AsMonster().Has(object.MonsterNPC) {
		e2eError(fmt.Errorf("Shock stock NPC is not initialized"))
		return
	}
	f.npc.UpdateDataMonster().SetAggression(0)
	f.npc.ClearActionStack()
	f.npc.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
	f.caster, f.target = f.host, f.host
	switch f.mode {
	case "npc", "npc-natural":
		f.caster, f.target = f.npc, f.npc
	case "glyph-to-npc":
		f.target = f.npc
	case "glyph-to-player":
		f.caster = f.npc
	}
	if f.target.Damage == nil {
		e2eError(fmt.Errorf("Shock target has no damage callback"))
		return
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(f.host), unsafe.Pointer(f.npc), f.host.UpdateData, f.npc.UpdateData, unsafe.Pointer(f.target.HealthData)} {
			if uintptr(ptr) <= math.MaxUint32 {
				e2eError(fmt.Errorf("Shock native fixture pointer is below 4 GiB: %p", ptr))
				return
			}
		}
	}
	if !f.isGlyph() {
		var ok bool
		f.duration, ok = e2eShockDuration(noxServer.Balance.Float("ShockEnchantDuration"))
		if !ok {
			e2eError(fmt.Errorf("Shock stock duration is outside the bounded scenario"))
			return
		}
		f.power = uint32(f.level)
		if f.mode == "npc-natural" {
			f.power = uint32(noxServer.Server.SpellPower4FE7B0(spell.SPELL_SHOCK, f.caster))
		}
	} else {
		if f.host.ControllingPlayer().PlayerClass() != player.Wizard || noxServer.SpellPrecheck4FD0E0(f.caster, spell.SPELL_SHOCK) != 0 {
			e2eError(fmt.Errorf("Shock Glyph needs an allowed stock Wizard/NPC owner"))
			return
		}
		f.power = 2
		if f.mode == "glyph-to-npc" {
			f.power = uint32(noxServer.Server.SpellPower4FE7B0(spell.SPELL_SHOCK, f.caster))
		}
		value := noxServer.Balance.FloatInd("ShockTrapDamage", int(f.power)-1)
		rounded := math.RoundToEven(float64(float32(value)))
		if rounded <= 0 || rounded >= float64(f.target.HealthData.Cur) {
			e2eError(fmt.Errorf("Shock stock Glyph damage is not a survivable fixture: damage=%g HP=%d", rounded, f.target.HealthData.Cur))
			return
		}
		f.damage = uint16(rounded)
	}
	noxServer.Audio.OnSound(func(id sound.ID, kind int, owner *server.Object, pos types.Pointf) {
		if !f.active {
			return
		}
		if owner == f.target && (id == sound.SoundShockOn || id == sound.SoundShockOff) {
			if kind != 0 || f.isGlyph() {
				e2eError(fmt.Errorf("Shock unexpected enchant audio mode=%s kind=%d", f.mode, kind))
				return
			}
			if id == sound.SoundShockOn {
				f.onAudio++
				f.observeApplication()
			} else {
				f.offAudio++
			}
		}
		if id == sound.SoundGlyphDetonate && f.isGlyph() && pos.Sub(f.glyphPos).Len() < 1 {
			f.detonateAudio++
		}
	})
	noxServer.TickHook(f.observe)
	e2eLog.Printf("SHOCK PREPARED: mode=%s level=%d power=%d caster=%p target=%p update=%p damage=%p", f.mode, f.level, f.power, f.caster, f.target, f.target.UpdateData, f.target.Damage)
}

func (f *e2eShockFixture) begin() {
	if f.target.Buffs != 0 || f.target.HasEnchant(server.ENCHANT_INVULNERABLE) ||
		noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target))) == nil {
		e2eError(fmt.Errorf("Shock placement/protection/client publication did not settle: mode=%s buffs=%#x flags=%#x status=%#x drawable=%p", f.mode, f.target.Buffs, f.target.ObjFlags, f.host.ControllingPlayer().Field3680, noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target)))))
		return
	}
	f.frame, f.health, f.active = noxServer.Frame(), f.target.HealthData.Cur, true
	f.keepaliveFrame = f.frame
	if f.mode == "npc-natural" {
		f.npc.MonsterPushAction(ai.ACTION_CAST_SPELL_ON_OBJECT, uint32(spell.SPELL_SHOCK), 0, f.target)
		return
	}
	api := noxServer.noxScriptP()
	if !f.isGlyph() {
		api.CastSpellLvl(nsp.Spell("SPELL_SHOCK"), f.level, api.toObj(f.caster), api.toObj(f.target))
		f.observeApplication()
		return
	}
	// Normal NewTrap creation and the normal owner service. The actual world
	// collision must trigger it: neither TriggerTrap nor Damage is called here.
	position := f.host.PosVec.Add(f.npc.PosVec).Mul(0.5)
	trap := api.NewTrap(position, []ns4.TrapSpell{{Spell: nsp.Spell("SPELL_SHOCK")}})
	if trap == nil {
		e2eError(fmt.Errorf("Shock NewTrap failed"))
		return
	}
	f.glyph = server.ToObject(trap.(server.Obj))
	noxServer.ObjSetOwner(f.caster, f.glyph)
	noxServer.ObjectsAddPending()
	f.glyphWire, f.glyphScript, f.glyphPos = f.glyph.NetCode, f.glyph.ScriptIDVal, f.glyph.PosVec
	f.glyphType = f.glyph.TypeInd
	idata := f.glyph.InitDataGlyph()
	if idata == nil || idata.SpellsCnt != 1 || idata.Spells[0] != uint32(spell.SPELL_SHOCK) || idata.SpellArg.Obj != nil ||
		unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(f.glyph)) <= math.MaxUint32 || uintptr(unsafe.Pointer(idata)) <= math.MaxUint32) {
		e2eError(fmt.Errorf("Shock real Glyph lost its native initialization"))
	}
	e2eLog.Printf("SHOCK GLYPH CREATED: object=%p init=%p owner=%p power=%d damage=%d", f.glyph, idata, f.glyph.ObjOwner, f.power, f.damage)
	if noxServer.GlyphCollideAllowed4E9A30(f.glyph, f.target) != 1 {
		e2eError(fmt.Errorf("Shock Glyph is not eligible for its actual target"))
		return
	}
	f.walkFrom = f.target.PosVec
	if f.mode == "glyph-to-npc" {
		api.toObj(f.target).WalkTo(position)
	} else {
		pos := noxClient.Viewport().ToScreenPos(image.Pt(int(position.X), int(position.Y)))
		e2eQueueInput(&seat.MouseMoveEvent{Pos: pos}, &seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: true})
	}
}

func (f *e2eShockFixture) observeApplication() {
	if f.applied || !f.target.HasEnchant(server.ENCHANT_SHOCK) {
		return
	}
	if f.target.EnchantPower(server.ENCHANT_SHOCK) != int(f.power) ||
		f.target.EnchantDur(server.ENCHANT_SHOCK) < int(f.duration)-1 || f.target.EnchantDur(server.ENCHANT_SHOCK) > int(f.duration) ||
		f.target.HealthData.Cur != f.health || noxServer.Frame()-f.frame > 150 {
		e2eError(fmt.Errorf("Shock application mismatch mode=%s timer=%d/%d power=%d/%d HP=%d/%d", f.mode, f.target.EnchantDur(server.ENCHANT_SHOCK), f.duration, f.target.EnchantPower(server.ENCHANT_SHOCK), f.power, f.target.HealthData.Cur, f.health))
		return
	}
	if f.mode == "npc-natural" {
		ud := f.npc.UpdateDataMonster()
		head := ud.AIStackHead()
		if head == nil || head.Type() != ai.ACTION_CAST_SPELL_ON_OBJECT || head.ArgObj(2) != f.target ||
			head.ArgU32(0) != uint32(spell.SPELL_SHOCK) || ud.Field120_2 != 0 || uint32(ud.Field120_1) != ud.MonsterDef.MissileAttackFrame216 {
			e2eError(fmt.Errorf("Shock did not use the real NPC cast-frame gate"))
			return
		}
		f.natural = true
		e2eLog.Printf("SHOCK NPC CAST FRAME: animation=%d elapsed=%d", ud.Field120_1, noxServer.Frame()-f.frame)
	}
	f.applied, f.appliedFrame = true, noxServer.Frame()
	e2eLog.Printf("SHOCK APPLIED: mode=%s level=%d power=%d timer=%d HP=%d unchanged", f.mode, f.level, f.power, f.target.EnchantDur(server.ENCHANT_SHOCK), f.health)
}

func (f *e2eShockFixture) observe() {
	if !f.active {
		return
	}
	if !f.isGlyph() {
		f.observeApplication()
		return
	}
	if f.hit || f.target.HealthData.Cur == f.health {
		return
	}
	var marker, typ uint32
	if f.mode == "glyph-to-player" {
		ud := f.target.UpdateDataPlayer()
		marker, typ = ud.Field76, ud.Field75
	} else {
		ud := f.target.UpdateDataMonster()
		marker, typ = ud.Field547, ud.Field546
		if !ud.StatusFlags.Has(object.MonStatusInjured) {
			e2eError(fmt.Errorf("Shock Glyph NPC lost the injured latch"))
			return
		}
	}
	if f.target.PosVec.Sub(f.walkFrom).Len() < 8 || f.target.HealthData.Cur != f.health-f.damage || f.target.Buffs != 0 ||
		f.target.Obj130 != f.caster || f.target.Field131 != uint32(object.DamageElectric) ||
		marker != 2 || typ != uint32(object.DamageElectric) || *legacy.Get_dword_5d4594_2487712_ptr() != uint32(f.glyphType) {
		e2eError(fmt.Errorf("Shock Glyph damage mismatch mode=%s HP=%d->%d expected=%d marker=%d/%d source=%p expected=%p cache=%d", f.mode, f.health, f.target.HealthData.Cur, f.damage, marker, typ, f.target.Obj130, f.caster, *legacy.Get_dword_5d4594_2487712_ptr()))
		return
	}
	f.hit = true
	if f.mode == "glyph-to-player" {
		e2eQueueInput(&seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: false})
	}
	e2eLog.Printf("SHOCK GLYPH HIT: mode=%s HP=%d->%d electric=%d frame=%d native-source=%p", f.mode, f.health, f.target.HealthData.Cur, f.damage, f.target.Frame134, f.target.Obj130)
}

func (f *e2eShockFixture) particles() (alive, pixels int, newest uint32) {
	typeID := noxClient.Things.IndByID("WhiteSpark")
	vp, pix := noxClient.Viewport(), noxClient.r.PixBuffer()
	frame := noxServer.Frame()
	for _, dr := range noxClient.Objs.AllList1() {
		if int(dr.TypeIDVal) != typeID || math.Abs(float64(dr.PosVec.X)-float64(f.target.PosVec.X)) > 40 || math.Abs(float64(dr.PosVec.Y)-float64(f.target.PosVec.Y)) > 40 {
			continue
		}
		effect := dr.UnionEffect()
		if effect.Field_111 >= frame || int32(effect.Field_112-frame) <= 0 {
			continue
		}
		alive++
		if effect.Field_111 > newest {
			newest = effect.Field_111
		}
		point := vp.ToScreenPos(dr.PosVec).Add(image.Pt(0, -int(int16(dr.ZVal2))-int(int16(dr.ZVal))))
		if e2eShockWhitePixel(pix, point, vp.Screen) {
			pixels++
		}
	}
	return
}

func (f *e2eShockFixture) visible(label string, elapsed uint32) bool {
	if !f.applied || noxServer.Frame()-f.appliedFrame < elapsed {
		return false
	}
	dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target)))
	if dr == nil || !dr.HasEnchant(server.ENCHANT_SHOCK) {
		return false
	}
	if dr.Buffs != f.target.Buffs || f.target.HealthData.Cur != f.health ||
		f.target == f.host && memmap.Uint32(0x5D4594, 1062540) != f.target.Buffs {
		e2eError(fmt.Errorf("Shock client enchant replay mismatch"))
		return true
	}
	alive, pixels, birth := f.particles()
	if alive == 0 || pixels == 0 || label == "advanced" && birth <= f.firstBirth {
		return false
	}
	if label == "first" {
		f.firstBirth = birth
	}
	e2eLog.Printf("SHOCK VISIBLE: mode=%s level=%d sample=%s timer=%d particles=%d actual-white-centers=%d newest=%d", f.mode, f.level, label, f.target.EnchantDur(server.ENCHANT_SHOCK), alive, pixels, birth)
	return true
}

func (f *e2eShockFixture) complete() bool {
	// Eleven natural 600-frame enchant intervals exceed the stock inactivity
	// limit. Use ordinary one-frame movement input, not observer/clock writes.
	if f.keepalivePressed {
		e2eQueueInput(&seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: false})
		f.keepalivePressed = false
	} else if noxServer.Frame()-f.keepaliveFrame >= 300 {
		pos := noxClient.Viewport().ToScreenPos(image.Pt(int(f.host.PosVec.X)+16, int(f.host.PosVec.Y)))
		e2eQueueInput(&seat.MouseMoveEvent{Pos: pos}, &seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: true})
		f.keepaliveFrame, f.keepalivePressed = noxServer.Frame(), true
	}
	if f.isGlyph() {
		if f.detonateAudio > 0 && !f.hit && noxServer.Frame()-f.frame > 4 {
			e2eError(fmt.Errorf("Shock Glyph detonated but inflicted no electric damage: mode=%s HP=%d/%d", f.mode, f.target.HealthData.Cur, f.health))
			return true
		}
		if !f.hit || f.detonateAudio != 1 || e2eFistInWorld(f.glyph, f.glyphWire, f.glyphScript) {
			return false
		}
		if f.mode == "glyph-to-player" {
			meter, ready := e2eClientHUDMeter(0)
			if !ready || meter.Current != uint32(f.health-f.damage) {
				return false
			}
		} else {
			dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target)))
			if dr == nil {
				return false
			}
			if delta, ok := legacy.HealthChangeForDrawable(dr.NetCode32); !ok || delta != -int16(f.damage) {
				return false
			}
		}
		e2eLog.Printf("SHOCK GLYPH COMPLETE: mode=%s power=%d HP=%d->%d native-damage=verified client-replay=verified", f.mode, f.power, f.health, f.target.HealthData.Cur)
		return true
	}
	dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target)))
	if !f.applied || dr == nil || noxServer.Frame()-f.appliedFrame < f.duration+25 || f.target.HasEnchant(server.ENCHANT_SHOCK) || dr.HasEnchant(server.ENCHANT_SHOCK) {
		return false
	}
	alive, _, _ := f.particles()
	if alive != 0 || f.onAudio != 1 || f.offAudio != 1 || f.target.EnchantDur(server.ENCHANT_SHOCK) != 0 ||
		f.target.HealthData.Cur != f.health || f.target == f.host && memmap.Uint32(0x5D4594, 1062540) != 0 ||
		f.mode == "npc-natural" && (!f.natural || f.npc.MonsterActionIsScheduled(ai.ACTION_CAST_SPELL_ON_OBJECT)) {
		e2eError(fmt.Errorf("Shock expiry mismatch mode=%s particles=%d on/off=%d/%d timer=%d", f.mode, alive, f.onAudio, f.offAudio, f.target.EnchantDur(server.ENCHANT_SHOCK)))
		return true
	}
	e2eLog.Printf("SHOCK EXPIRED: mode=%s level=%d power=%d elapsed=%d timer=0 client-buff=cleared particles=0 audio=1/1 natural-NPC=%t", f.mode, f.level, f.power, noxServer.Frame()-f.appliedFrame, f.natural)
	return true
}

// Placement, waiting AI, normal self-targeted script casts and NewTrap are
// fixture inputs. Buffs, HP, timers, collision, damage callbacks, packets,
// effect drawables and pixels are all produced by the actual game pipeline.
func (sc *e2eScenario) CheckShockSpell(level int, mode, name string) {
	if !e2eShockMode(level, mode) {
		e2eError(fmt.Errorf("invalid Shock fixture: level=%d mode=%q", level, mode))
		return
	}
	f := &e2eShockFixture{mode: mode, level: level}
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		host := noxServer.Players.HostUnit()
		return host != nil && host.Buffs == 0 && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, f.prepare)
	sc.Wait(12, name+" publish placement")
	sc.add(0, name+" cast", f.begin)
	if !f.isGlyph() {
		for _, sample := range []struct {
			label   string
			elapsed uint32
		}{{"first", 12}, {"advanced", 24}} {
			sc.addWhen(0, name+" "+sample.label+" visible", 180, func() bool { return f.visible(sample.label, sample.elapsed) }, func() {})
			// visible asserts real, advancing particle pixels and synchronized
			// buffs. Whole-frame snapshots are private diagnostics only: the
			// surrounding map continues animating during the bounded wait.
			sc.CaptureMagicFrame(name + " " + sample.label)
		}
	}
	sc.addWhen(1, name+" complete", 3000, f.complete, func() {
		f.active = false
		noxServer.DelayedDelete(f.npc)
		asObjectS(f.host).SetPos(f.original)
		f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	})
}
