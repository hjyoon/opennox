package opennox

import (
	"fmt"
	"image"
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	nsp "github.com/opennox/noxscript/ns/v4/spell"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func e2eFumbleMode(level int, direction string) (fromNPC, ordinary, ok bool) {
	switch direction {
	case "player-to-npc":
		return false, false, level >= 1 && level <= 5
	case "npc-to-player":
		return true, false, level == 0
	case "player-to-monster":
		return false, true, level == 3
	default:
		return false, false, false
	}
}

// Independently express the original inventory predicate using named classes:
// humanoids lose equipped weapons/wands/shields, not body armor or spare gear.
// The ordinary-monster fixture instead exercises the real drop-all routine.
func e2eFumbleShouldDrop(item *server.Object, ordinary bool) bool {
	return ordinary || item.Flags().Has(object.FlagEquipped) &&
		(item.Class().HasAny(object.ClassWeapon|object.ClassWand) ||
			item.Class().Has(object.ClassArmor) && item.SubClass().AsArmor().Has(object.ArmorShield))
}

func e2eFumbleExpectedForce(origin, position types.Pointf, mass float32) types.Pointf {
	delta := position.Sub(origin)
	length := float32(math.Hypot(float64(delta.X), float64(delta.Y)) + 0.1)
	strength := float32(500) / mass
	return types.Ptf(float32(float64(delta.X)*float64(strength)/float64(length)),
		float32(float64(delta.Y)*float64(strength)/float64(length)))
}

type e2eFumbleFixture struct {
	level                               int
	direction                           string
	fromNPC, ordinary, targeted         bool
	host, npc, caster, target           *server.Object
	magic                               *server.Object
	magicWire                           uint32
	magicScript                         int32
	original, targetOrigin, forceOrigin types.Pointf
	clientOrigin                        image.Point
	dropped, retained, granted          []*server.Object
	retainedFlags                       map[*server.Object]object.Flags
	health                              uint16
	frame                               uint32
	active, effectSeen, naturalCastSeen bool
	castAudio, effectAudio              int
}

func (f *e2eFumbleFixture) grant(typeID string, equip bool) {
	item := legacy.Nox_xxx_playerRespawnItem_4EF750(f.target, typeID, nil, 1, 0)
	if item == nil {
		e2eError(fmt.Errorf("Fumble stock item %s was not created", typeID))
		return
	}
	if item.InvHolder != f.target || !f.target.HasItem(item) ||
		!equip && item.Flags().Has(object.FlagEquipped) ||
		equip && !item.Flags().Has(object.FlagEquipped) && !asObjectS(f.target).Equip(item) {
		e2eError(fmt.Errorf("Fumble stock item %s was not placed/equipped through normal services: holder=%p flags=%#x weight=%d target-capacity=%d", typeID, item.InvHolder, uint32(item.Flags()), item.Weight, f.target.CarryCapacity))
		return
	}
	f.granted = append(f.granted, item)
}

func (f *e2eFumbleFixture) prepare() {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.HealthData == nil || f.host.HealthData.Cur == 0 ||
		f.host.Flags().HasAny(object.FlagDead|object.FlagDestroyed|object.FlagNoUpdate) {
		e2eError(fmt.Errorf("Fumble requires a live host"))
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
	typeID := "NPC"
	if f.ordinary {
		typeID = "Troll"
	}
	f.npc = noxServer.NewObjectByTypeID(typeID)
	if f.npc == nil {
		e2eError(fmt.Errorf("Fumble has no stock %s type", typeID))
		return
	}
	hostPos, npcPos := origin, origin.Add(direction.Mul(112))
	f.caster, f.target = f.host, f.npc
	if f.fromNPC {
		hostPos, npcPos = npcPos, hostPos
		f.caster, f.target = f.npc, f.host
	}
	asObjectS(f.host).SetPos(hostPos)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	noxServer.CreateObjectAt(f.npc, nil, npcPos)
	noxServer.ObjectsAddPending()
	if !e2eObjectInWorld(f.npc) || f.npc.UpdateData == nil || f.npc.HealthData == nil ||
		f.npc.UpdateDataMonster().MonsterDef == nil || f.target.CarryCapacity == 0 ||
		f.npc.SubClass().AsMonster().Has(object.MonsterNPC) == f.ordinary ||
		f.ordinary && (!f.target.IsMovable() || f.target.Mass <= 0) {
		e2eError(fmt.Errorf("Fumble stock %s/target was not initialized", typeID))
		return
	}
	f.npc.UpdateDataMonster().SetAggression(0)
	f.npc.ClearActionStack()
	f.npc.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
	if f.ordinary {
		f.grant("StreetShirt", false)
		f.grant("StreetSneakers", false)
	} else {
		f.grant("Longsword", true)
		f.grant("WoodenShield", true)
		f.grant("StreetShirt", true)
		spare := "StaffWooden"
		if f.fromNPC {
			spare = "Sword"
		}
		f.grant(spare, false)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		pointers := []unsafe.Pointer{unsafe.Pointer(f.host), unsafe.Pointer(f.npc), f.npc.UpdateData}
		for _, item := range f.granted {
			pointers = append(pointers, unsafe.Pointer(item))
		}
		for _, pointer := range pointers {
			if uintptr(pointer) <= math.MaxUint32 {
				e2eError(fmt.Errorf("Fumble native fixture allocation is below 4 GiB: %p", pointer))
				return
			}
		}
	}
	f.targeted = noxServer.Spells.HasFlags(spell.SPELL_FUMBLE, things.SpellTargeted)
	noxServer.Audio.OnSound(func(id sound.ID, kind int, owner *server.Object, _ types.Pointf) {
		if !f.active {
			return
		}
		if id == sound.SoundFumbleCast && owner == f.caster {
			f.castAudio++
			if kind != 0 || !f.targeted || f.castAudio != 1 {
				e2eError(fmt.Errorf("Fumble cast sound kind/count=%d/%d", kind, f.castAudio))
				return
			}
			f.observeProjectile()
			if f.fromNPC {
				f.observeNaturalCast()
			}
		}
		if id == sound.SoundFumbleEffect && owner == f.target {
			f.effectAudio++
			if kind != 0 || f.effectAudio != 1 {
				e2eError(fmt.Errorf("Fumble effect sound kind/count=%d/%d", kind, f.effectAudio))
				return
			}
			if f.fromNPC && !f.targeted {
				f.observeNaturalCast()
			}
			f.verifyEffect()
		}
	})
	e2eLog.Printf("FUMBLE PREPARED: direction=%s requested-level=%d targeted=%t caster=%p target=%p update=%p", f.direction, f.level, f.targeted, f.caster, f.target, f.npc.UpdateData)
}

func (f *e2eFumbleFixture) beginCast() {
	if !e2eObjectInWorld(f.target) || !e2eObjectInWorld(f.caster) || f.target.ForceVec != (types.Pointf{}) ||
		f.target.VelVec != (types.Pointf{}) || !noxServer.MapTraceRay(f.caster.PosVec, f.target.PosVec, 0) {
		e2eError(fmt.Errorf("Fumble placement did not settle without external force"))
		return
	}
	drawable := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target)))
	if drawable == nil {
		e2eError(fmt.Errorf("Fumble target drawable has not been published"))
		return
	}
	f.retainedFlags = make(map[*server.Object]object.Flags)
	for item := f.target.InvFirstItem; item != nil; item = item.InvNextItem {
		if e2eFumbleShouldDrop(item, f.ordinary) {
			f.dropped = append(f.dropped, item)
		} else {
			f.retained = append(f.retained, item)
			f.retainedFlags[item] = item.Flags() & object.FlagEquipped
		}
	}
	if len(f.dropped) < 2 || !f.ordinary && len(f.retained) < 2 {
		e2eError(fmt.Errorf("Fumble inventory fixture lacks dropped/control items: %d/%d", len(f.dropped), len(f.retained)))
		return
	}
	f.targetOrigin, f.clientOrigin, f.health = f.target.PosVec, drawable.PosVec, f.target.HealthData.Cur
	f.frame, f.active = noxServer.Frame(), true
	if f.fromNPC {
		f.npc.MonsterPushAction(ai.ACTION_CAST_SPELL_ON_OBJECT, uint32(spell.SPELL_FUMBLE), 0, f.target)
		return
	}
	api := noxServer.noxScriptP()
	api.CastSpellLvl(nsp.Spell("SPELL_FUMBLE"), f.level, api.toObj(f.caster), api.toObj(f.target))
}

func (f *e2eFumbleFixture) observeProjectile() {
	for obj := f.caster.Field129; obj != nil; obj = obj.Field128 {
		if obj.ObjectTypeC().ID() != "Magic" || obj.Flags().Has(object.FlagDestroyed) {
			continue
		}
		ud := obj.UpdateDataSpellProjectile()
		if ud == nil || ud.Spell12 != uint32(spell.SPELL_FUMBLE) {
			continue
		}
		if f.magic != nil || ud.Target != f.target || ud.Field0 != f.caster || ud.Field8 != f.caster ||
			ud.Level16 != uint32(noxServer.Server.SpellPower4FE7B0(spell.SPELL_FUMBLE, f.caster)) ||
			unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 || uintptr(obj.UpdateData) <= math.MaxUint32) {
			e2eError(fmt.Errorf("Fumble projectile lost its native source/target or mode-selected power"))
			return
		}
		f.magic, f.magicWire, f.magicScript = obj, obj.NetCode, obj.ScriptIDVal
		e2eLog.Printf("FUMBLE PROJECTILE: object=%p update=%p target=%p requested-level=%d actual-power=%d", obj, obj.UpdateData, ud.Target, f.level, ud.Level16)
	}
	if f.magic == nil {
		e2eError(fmt.Errorf("targeted Fumble cast has no real Magic projectile"))
	}
}

func (f *e2eFumbleFixture) observeNaturalCast() {
	ud := f.npc.UpdateDataMonster()
	head := ud.AIStackHead()
	if f.naturalCastSeen || head.Type() != ai.ACTION_CAST_SPELL_ON_OBJECT || head.ArgObj(2) != f.target ||
		head.ArgU32(0) != uint32(spell.SPELL_FUMBLE) || ud.Field120_2 != 0 ||
		uint32(ud.Field120_1) != ud.MonsterDef.MissileAttackFrame216 {
		e2eError(fmt.Errorf("Fumble did not use the real NPC action/cast-frame gate"))
		return
	}
	f.naturalCastSeen = true
	e2eLog.Printf("FUMBLE NPC CAST FRAME: animation=%d duration=%d target=%p elapsed=%d", ud.Field120_1, ud.Field120_2, head.ArgObj(2), noxServer.Frame()-f.frame)
}

func (f *e2eFumbleFixture) verifyEffect() {
	if f.effectSeen || f.target.HealthData.Cur != f.health {
		e2eError(fmt.Errorf("Fumble repeated its effect or caused HP damage"))
		return
	}
	for _, item := range f.dropped {
		if f.target.HasItem(item) || item.InvHolder != nil || item.Flags().HasAny(object.FlagEquipped|object.FlagDestroyed) {
			e2eError(fmt.Errorf("Fumble did not naturally drop %s: holder=%p flags=%#x", item.ObjectTypeC().ID(), item.InvHolder, uint32(item.Flags())))
			return
		}
	}
	for _, item := range f.retained {
		if !f.target.HasItem(item) || item.InvHolder != f.target || item.Flags().Has(object.FlagDestroyed) ||
			item.Flags()&object.FlagEquipped != f.retainedFlags[item] {
			e2eError(fmt.Errorf("Fumble unexpectedly removed body armor/spare %s", item.ObjectTypeC().ID()))
			return
		}
	}
	want := types.Pointf{}
	if f.ordinary {
		f.forceOrigin = f.caster.PosVec
		if f.targeted {
			if !e2eFistInWorld(f.magic, f.magicWire, f.magicScript) {
				e2eError(fmt.Errorf("Fumble impact projectile disappeared before its effect callback"))
				return
			}
			// The original projectile collision passes the Magic object as the
			// fourth SpellAccept argument. Fumble uses that object's position,
			// not the initial spellcaster's position, as its force origin.
			f.forceOrigin = f.magic.PosVec
		}
		want = e2eFumbleExpectedForce(f.forceOrigin, f.target.PosVec, f.target.Mass)
	}
	if f.target.ForceVec != want {
		e2eError(fmt.Errorf("Fumble force=%v, want=%v (ordinary=%t)", f.target.ForceVec, want, f.ordinary))
		return
	}
	f.effectSeen = true
	e2eLog.Printf("FUMBLE EFFECT: direction=%s dropped=%d retained=%d force=%v force-origin=%v HP=%d unchanged elapsed=%d", f.direction, len(f.dropped), len(f.retained), want, f.forceOrigin, f.health, noxServer.Frame()-f.frame)
}

func (f *e2eFumbleFixture) waitFor(reason string) bool {
	elapsed := noxServer.Frame() - f.frame
	if elapsed == 30 || elapsed != 0 && elapsed%100 == 0 {
		e2eLog.Printf("FUMBLE WAIT: direction=%s requested-level=%d elapsed=%d %s", f.direction, f.level, elapsed, reason)
	}
	return false
}

func (f *e2eFumbleFixture) complete() bool {
	if !f.effectSeen || f.effectAudio != 1 || f.targeted && (f.castAudio != 1 || f.magic == nil) ||
		f.fromNPC && (!f.naturalCastSeen || f.npc.MonsterActionIsScheduled(ai.ACTION_CAST_SPELL_ON_OBJECT)) {
		return f.waitFor(fmt.Sprintf("effect=%t effect-audio=%d cast-audio=%d projectile=%p NPC-natural=%t", f.effectSeen, f.effectAudio, f.castAudio, f.magic, f.naturalCastSeen))
	}
	if !e2eObjectInWorld(f.target) || f.target.HealthData.Cur != f.health {
		e2eError(fmt.Errorf("Fumble target disappeared or took damage"))
		return true
	}
	if f.targeted && (e2eFistInWorld(f.magic, f.magicWire, f.magicScript) || noxClient.Objs.ByNetCode(uint16(f.magicWire)) != nil) {
		return f.waitFor(fmt.Sprintf("projectile wire=%#x world=%t client=%t", f.magicWire, e2eFistInWorld(f.magic, f.magicWire, f.magicScript), noxClient.Objs.ByNetCode(uint16(f.magicWire)) != nil))
	}
	visibleDropped, occludedDropped := 0, 0
	for _, item := range f.dropped {
		if !e2eObjectInWorld(item) || item.InvHolder != nil {
			return f.waitFor(fmt.Sprintf("drop %s wire=%#x world=%t holder=%p client=%t pos=%v host=%v target=%v ray=%t flags=%#x", item.ObjectTypeC().ID(), noxServer.GetUnitNetCode(item), e2eObjectInWorld(item), item.InvHolder, noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(item))) != nil, item.PosVec, f.host.PosVec, f.target.PosVec, noxServer.MapTraceRay(f.host.PosVec, item.PosVec, 0), uint32(item.Flags())))
		}
		// The stock force-drop helper picks random reachable points around the
		// target, not around the host. A corner can therefore hide a valid drop.
		// Demand normal client publication only for drops inside the host's
		// real vision; never inject position, visibility or a drawable result.
		if noxServer.MapTraceVision(f.host, item) {
			visibleDropped++
			if noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(item))) == nil {
				return f.waitFor(fmt.Sprintf("visible drop %s wire=%#x has no client drawable", item.ObjectTypeC().ID(), noxServer.GetUnitNetCode(item)))
			}
		} else {
			occludedDropped++
		}
	}
	weapon, armor, _ := e2eEquippedNPCMasks(f.target)
	if !f.ordinary {
		var clientWeapon, clientArmor uint32
		if f.fromNPC {
			drawable := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target)))
			if drawable == nil {
				return false
			}
			player := noxClient.Server.Players.ByID(int(drawable.NetCode32))
			if player == nil {
				return false
			}
			clientWeapon, clientArmor = player.WeaponEquip, player.ArmorEquip
		} else {
			npc := noxClient.Server.NPCs.ByID(noxServer.GetUnitNetCode(f.target))
			if npc == nil {
				return false
			}
			clientWeapon, clientArmor = npc.WeaponEquip, npc.ArmorEquip
		}
		if clientWeapon&^1 != weapon&^1 || clientArmor != armor {
			return f.waitFor(fmt.Sprintf("target wire=%#x equipment client=%#x/%#x server=%#x/%#x", noxServer.GetUnitNetCode(f.target), clientWeapon, clientArmor, weapon, armor))
		}
	}
	if f.ordinary {
		drawable := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target)))
		if drawable == nil {
			return false
		}
		direction := f.targetOrigin.Sub(f.forceOrigin).Normalize()
		delta := f.target.PosVec.Sub(f.targetOrigin)
		clientDelta := drawable.PosVec.Sub(f.clientOrigin)
		if delta.X*direction.X+delta.Y*direction.Y <= 1 || float32(clientDelta.X)*direction.X+float32(clientDelta.Y)*direction.Y <= 1 {
			return false
		}
	}
	e2eLog.Printf("FUMBLE COMPLETE: direction=%s requested-level=%d dropped=%d retained=%d visible-drawables=%d occluded-drops=%d client-weapon=%#x client-armor=%#x projectile=removed NPC-natural=%t elapsed=%d", f.direction, f.level, len(f.dropped), len(f.retained), visibleDropped, occludedDropped, weapon, armor, f.naturalCastSeen, noxServer.Frame()-f.frame)
	return true
}

// Real script object casts and a naturally animated queued NPC cast. Placement,
// waiting AI, normal item grants/equip and explicit target selection are setup.
// No drop, equipment flag, force, cast-frame, HP, projectile or client result is
// injected. Targeted spells retain the stock mode-selected projectile power;
// requested script levels are not claimed to override it. Fumble ignores level.
func (sc *e2eScenario) CheckFumbleSpell(level int, direction, name string) {
	fromNPC, ordinary, ok := e2eFumbleMode(level, direction)
	if !ok {
		e2eError(fmt.Errorf("invalid Fumble fixture: level=%d direction=%q", level, direction))
		return
	}
	f := &e2eFumbleFixture{level: level, direction: direction, fromNPC: fromNPC, ordinary: ordinary}
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		return noxServer.Players.HostUnit() != nil && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, f.prepare)
	sc.Wait(12, name+" publish placement and equipment")
	sc.add(0, name+" cast", f.beginCast)
	sc.addWhen(1, name+" actual drop and client replay", 300, f.complete, func() {
		f.active = false
		for _, item := range f.granted {
			if item.InvHolder == nil {
				noxServer.DelayedDelete(item)
			}
		}
		noxServer.DelayedDelete(f.npc)
		asObjectS(f.host).SetPos(f.original)
		f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	})
	sc.Wait(3, name+" cleanup")
}
