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

func e2eAIInversionMode(mode string) (kind string, descending, ok bool) {
	kind, slope, found := strings.Cut(mode, "/")
	if !found {
		return "", false, false
	}
	switch kind {
	case "Wizard", "WizardGreen", "UrchinShaman":
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

// Called only with a live missile obtained from the server's missile list.
// A stored pointer is an identity key, never permission to read a deleted unit.
func e2eAIInversionReflection(missile, attacker, defender *server.Object, wire uint32, script int32, frame uint32) bool {
	if missile == nil || missile.UpdateData == nil || missile.NetCode != wire || missile.ScriptIDVal != script ||
		missile.Flags().HasAny(object.FlagDestroyed|object.FlagDead) ||
		!missile.Class().Has(object.ClassMissile) || !missile.SubClass().AsMissile().Has(object.MissileMagic) ||
		missile.ObjOwner != defender || missile.Field32 != frame {
		return false
	}
	data := missile.UpdateDataMissile()
	return data.Owner == defender && data.Target == attacker && data.SpellID == int32(spell.SPELL_MAGIC_MISSILE)
}

func e2eAIInversionOwned(owner, missile *server.Object) bool {
	for it, remaining := owner.Field129, 4096; it != nil && remaining > 0; it, remaining = it.Field128, remaining-1 {
		if it == missile {
			return true
		}
	}
	return false
}

type e2eAIInversionMissile struct {
	wire                uint32
	script              int32
	reflected, returned bool
	frame               uint32
	pos                 types.Pointf
}

type e2eAIInversionFixture struct {
	mode, kind                string
	descending                bool
	host, defender            *server.Object
	original                  types.Pointf
	hostHP, hostMax           uint16
	start, lastShot, lastLog  uint32
	shots, casts, reflections int
	active, natural, hit      bool
	hitHP                     uint16
	hitFrame                  uint32
	hitWire                   uint32
	castFrames                map[uint32]bool
	missiles                  map[*server.Object]*e2eAIInversionMissile
}

func (f *e2eAIInversionFixture) prepare() {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.HealthData == nil || f.host.HealthData.Cur == 0 || f.host.Buffs != 0 || f.host.Poison540 != 0 || f.host.ControllingPlayer() == nil {
		e2eError(fmt.Errorf("AI inversion needs a live unenchanted host"))
		return
	}
	f.original, f.hostHP, f.hostMax = f.host.PosVec, f.host.HealthData.Cur, f.host.HealthData.Max
	f.defender = noxServer.NewObjectByTypeID(f.kind)
	if f.defender == nil || f.defender.HealthData == nil || f.defender.UpdateData == nil {
		e2eError(fmt.Errorf("AI inversion stock monster %s unavailable", f.kind))
		return
	}
	from, to, err := e2eAIFirstAttackArena(f.host, max(f.host.Shape.Circle.R, f.defender.Shape.Circle.R)+12, f.descending)
	if err != nil {
		e2eError(err)
		return
	}
	asObjectS(f.host).SetPos(to)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	noxServer.CreateObjectAt(f.defender, nil, from)
	noxServer.ObjectsAddPending()
	f.defender.SetDir(server.DirFromVec(to.Sub(from)))
	// Durable health and ordinary placement are setup. Inversion capability,
	// its cooldown/RNG, spell flags, enemy selection and every action are stock.
	asObjectS(f.host).SetMaxHealth(2000)
	asObjectS(f.defender).SetMaxHealth(2000)
	for _, p := range []unsafe.Pointer{unsafe.Pointer(f.host), unsafe.Pointer(f.defender), f.host.UpdateData, f.defender.UpdateData} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
			e2eError(fmt.Errorf("AI inversion object/update pointer below 4 GiB: %p", p))
			return
		}
	}
	f.castFrames = make(map[uint32]bool)
	f.missiles = make(map[*server.Object]*e2eAIInversionMissile)
	noxServer.Audio.OnSound(f.observeSound)
	noxServer.TickHook(f.tick)
}

func (f *e2eAIInversionFixture) begin() {
	ud := f.defender.UpdateDataMonster()
	if !e2eObjectInWorld(f.defender) || ud.MonsterDef == nil ||
		!ud.StatusFlags.Has(object.MonStatusCanCastSpells) || ud.Field410&0x08000000 == 0 ||
		f.defender.Buffs != 0 || !noxServer.S().IsEnemyTo(f.defender, f.host) ||
		!noxServer.MapTraceRay(f.host.PosVec, f.defender.PosVec, server.MapTraceFlag1) ||
		noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.defender))) == nil {
		e2eError(fmt.Errorf("stock AI inversion unavailable: mode=%s status=%x spell38=%x buffs=%x stack=%v", f.mode, ud.StatusFlags, ud.Field410, f.defender.Buffs, ud.GetAIStack()))
		return
	}
	f.start, f.active = noxServer.Frame(), true
	e2eLog.Printf("AI INVERSION PREPARED: mode=%s frame=%d unit=%p update=%p stock-mask=%x cooldown=%d bounds=%d..%d cast-frame=%d range=%g stack=%v", f.mode, f.start, f.defender, f.defender.UpdateData, ud.Field410, ud.Field363, ud.Field362_0, ud.Field362_2, ud.MonsterDef.MissileAttackFrame216, noxServer.Balance.Float("InversionRange"), ud.GetAIStack())
	f.fire()
}

func (f *e2eAIInversionFixture) fire() {
	if f.shots >= 30 || f.hit {
		return
	}
	// An ordinary script attack supplies the threat, not the monster's choice
	// of spell, self-cast action, animation frame or the projectile's result.
	ud := f.host.UpdateDataPlayer()
	ud.CursorObj, ud.Player.Obj3640 = f.defender, f.defender
	ud.Field55, ud.Field56 = int(f.defender.PosVec.X), int(f.defender.PosVec.Y)
	ud.Player.CursorVec = image.Pt(int(f.defender.PosVec.X), int(f.defender.PosVec.Y))
	mouse := noxClient.Viewport().ToScreenPos(ud.Player.CursorVec)
	noxClient.ChangeMousePos(mouse, true)
	e2eQueueInput(&seat.MouseMoveEvent{Pos: mouse})
	f.host.SetDir(server.DirFromVec(f.defender.PosVec.Sub(f.host.PosVec)))
	api := noxServer.noxScriptP()
	api.CastSpellLvl(nsp.Spell("SPELL_MAGIC_MISSILE"), 1, api.toObj(f.host), api.toObj(f.defender))
	noxServer.ObjectsAddPending()
	f.shots++
	f.lastShot = noxServer.Frame()
	f.observeMissiles()
}

func (f *e2eAIInversionFixture) observeMissiles() {
	for _, missile := range noxServer.Objs.AllMissiles() {
		if missile.UpdateData == nil || missile.Flags().Has(object.FlagDestroyed) ||
			!missile.SubClass().AsMissile().Has(object.MissileMagic) {
			continue
		}
		data := missile.UpdateDataMissile()
		if data.SpellID != int32(spell.SPELL_MAGIC_MISSILE) {
			continue
		}
		record := f.missiles[missile]
		if record == nil || record.wire != missile.NetCode || record.script != missile.ScriptIDVal {
			if missile.ObjOwner != f.host || data.Owner != f.host || data.Target != f.defender {
				continue
			}
			if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(missile)) <= math.MaxUint32 || uintptr(missile.UpdateData) <= math.MaxUint32) {
				e2eError(fmt.Errorf("AI inversion projectile pointer below 4 GiB"))
				return
			}
			record = &e2eAIInversionMissile{wire: missile.NetCode, script: missile.ScriptIDVal}
			f.missiles[missile] = record
		}
		if !record.reflected && f.castFrames[noxServer.Frame()] && e2eAIInversionReflection(missile, f.host, f.defender, record.wire, record.script, noxServer.Frame()) {
			if !e2eAIInversionOwned(f.defender, missile) || e2eAIInversionOwned(f.host, missile) {
				e2eError(fmt.Errorf("inverted missile was not transferred between ordinary owner lists"))
				return
			}
			record.reflected, record.frame, record.pos = true, noxServer.Frame(), missile.PosVec
			f.reflections++
			e2eLog.Printf("AI INVERSION REFLECTED: mode=%s frame=%d missile=%p update=%p native-owner=%p target=%p lifetime-reset=%d", f.mode, record.frame, missile, missile.UpdateData, data.Owner, data.Target, missile.Field32)
		}
		if record.reflected && noxServer.Frame() > record.frame && !record.returned {
			movement, toward := missile.PosVec.Sub(record.pos), f.host.PosVec.Sub(record.pos)
			if data.Owner == f.defender && data.Target == f.host && movement.X*toward.X+movement.Y*toward.Y > 0 && noxClient.Objs.ByNetCode(uint16(record.wire)) != nil {
				record.returned = true
				e2eLog.Printf("AI INVERSION RETURNED: mode=%s wire=%d client-drawable=present movement=%v", f.mode, record.wire, movement)
			}
		}
	}
}

func (f *e2eAIInversionFixture) observeSound(id sound.ID, kind int, owner *server.Object, _ types.Pointf) {
	if !f.active {
		return
	}
	if id == sound.SoundInversionCast && owner == f.defender {
		ud, head := f.defender.UpdateDataMonster(), f.defender.UpdateDataMonster().AIStackHead()
		if kind != 0 || head == nil || head.Type() != ai.ACTION_CAST_SPELL_ON_OBJECT ||
			head.ArgU32(0) != uint32(spell.SPELL_INVERSION) || head.ArgObj(2) != f.defender ||
			ud.Field120_2 != 0 || uint32(ud.Field120_1) != ud.MonsterDef.MissileAttackFrame216 {
			e2eError(fmt.Errorf("AI inversion did not use its natural self-cast action and animation frame"))
			return
		}
		if !f.castFrames[noxServer.Frame()] {
			f.casts++
			f.castFrames[noxServer.Frame()] = true
		}
		f.natural = true
		f.observeMissiles()
	}
	if id == sound.SoundMagicMissileDetonate {
		record := f.missiles[owner]
		if record == nil || !record.reflected || !record.returned || record.wire != owner.NetCode || record.script != owner.ScriptIDVal {
			return
		}
		data := owner.UpdateDataMissile()
		if kind != 0 || data.Owner != f.defender || data.Target != f.host ||
			f.host.Obj130 != owner || f.host.Frame134 != noxServer.Frame() || f.host.HealthData.Cur >= 2000 {
			e2eError(fmt.Errorf("reflected missile did not deal ordinary return-hit damage"))
			return
		}
		f.hit, f.hitHP, f.hitFrame, f.hitWire = true, f.host.HealthData.Cur, noxServer.Frame(), record.wire
		e2eLog.Printf("AI INVERSION RETURN HIT: mode=%s wire=%d HP=%d damage-source=reflected-missile frame=%d", f.mode, f.hitWire, f.hitHP, f.hitFrame)
	}
}

func (f *e2eAIInversionFixture) tick() {
	if !f.active {
		return
	}
	f.observeMissiles()
	if !f.hit && noxServer.Frame()-f.lastShot >= 18 {
		f.fire()
	}
	if noxServer.Frame()-f.lastLog >= 30 {
		f.lastLog = noxServer.Frame()
		ud := f.defender.UpdateDataMonster()
		e2eLog.Printf("AI INVERSION TICK: mode=%s elapsed=%d shots=%d casts=%d reflected=%d HP=%d/%d deadline=%d action=%v animation=%d/%d", f.mode, noxServer.Frame()-f.start, f.shots, f.casts, f.reflections, f.host.HealthData.Cur, f.defender.HealthData.Cur, ud.Field363, ud.AIStackHead().Type(), ud.Field120_1, ud.Field120_2)
	}
}

func (f *e2eAIInversionFixture) complete() bool {
	if !f.natural || !f.hit || f.casts == 0 || f.reflections == 0 {
		return false
	}
	dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.host)))
	delta, ready := legacy.HealthChangeForDrawable(uint32(noxServer.GetUnitNetCode(f.host)))
	if dr == nil || !ready || delta >= 0 {
		return false
	}
	path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
	if err != nil {
		e2eError(err)
		return true
	}
	e2eLog.Printf("AI INVERSION PASS: mode=%s shots=%d natural-casts=%d reflections=%d return-hit-wire=%d actual-HP=%d client-delta=%d path=%s", f.mode, f.shots, f.casts, f.reflections, f.hitWire, f.hitHP, delta, path)
	return true
}

func (f *e2eAIInversionFixture) removed() bool {
	for missile, record := range f.missiles {
		if e2eFistInWorld(missile, record.wire, record.script) || noxClient.Objs.ByNetCode(uint16(record.wire)) != nil {
			return false
		}
	}
	e2eLog.Printf("AI INVERSION REMOVED: mode=%s tracked-missiles=%d server/client=absent", f.mode, len(f.missiles))
	return true
}

func (f *e2eAIInversionFixture) cleanup() {
	f.active = false
	noxServer.DelayedDelete(f.defender)
	noxServer.RemovePoison4EE9D0(f.host)
	asObjectS(f.host).SetPos(f.original)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	asObjectS(f.host).SetMaxHealth(int(f.hostMax))
	asObjectS(f.host).SetHealth(int(f.hostHP))
}

// No inversion request, self-cast action, cooldown, RNG, projectile target,
// ownership, velocity, hit result, packet or pixels are supplied by this probe.
func (sc *e2eScenario) CheckAIInversion(mode, name string) {
	kind, descending, ok := e2eAIInversionMode(mode)
	if !ok {
		e2eError(fmt.Errorf("invalid stock AI inversion mode %q", mode))
		return
	}
	f := &e2eAIInversionFixture{mode: mode, kind: kind, descending: descending}
	sc.addWhen(0, name+" prepare stock inversion caster", 1200, func() bool {
		unit := noxServer.Players.HostUnit()
		return nox_client_isConnected() && unit != nil && noxClient.ClientPlayerUnit() != nil && unit.Buffs == 0 && unit.Poison540 == 0
	}, f.prepare)
	sc.add(5, name+" ordinary incoming magic attack", f.begin)
	sc.addWhen(1, name+" autonomous inversion and return hit", 1200, f.complete, func() {})
	sc.addWhen(1, name+" natural missile removal", 300, f.removed, f.cleanup)
	sc.Wait(12, name+" ordinary inversion cleanup")
}
