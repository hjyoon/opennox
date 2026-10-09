package opennox

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func e2ePowderBarrelBreakingType(kind string) (string, bool) {
	switch kind {
	case "BlackPowderBarrel", "BlackPowderBarrel2":
		return kind + "Breaking", true
	default:
		return "", false
	}
}

// Independent observation contracts from 0053C9A0 -> 005350C0 and
// 004E1F84/004E20F0 -> 004E0B30. Do not call the production damage helpers
// to manufacture expected HP, and never supply these values to the game.
func e2ePowderBarrelRadialDamage(distance float32) int32 {
	if distance > 100 {
		return 0
	}
	damage := float32(30)
	if distance >= 30 {
		damage *= 1 - (distance-30)/70
	}
	return int32(damage)
}

func e2ePowderBarrelUnitDamage(raw int32, armored, immune bool, armor, carry float32) (int32, uint32) {
	damage, remainder := raw, math.Float32bits(carry)
	if armored {
		accumulated := float32((1-float64(armor))*float64(raw)) + carry
		damage = int32(math.RoundToEven(float64(accumulated)))
		remainder = math.Float32bits(accumulated - float32(damage))
		if raw > 0 && damage == 0 {
			damage = 1
		}
	}
	if immune {
		damage /= 2
	}
	return damage, remainder
}

type e2ePowderBarrelVictim struct {
	name    string
	unit    *server.Object
	health  uint16
	damage  int32
	carry   uint32
	armored bool
	hit     bool
}

type e2ePowderBarrelFixture struct {
	kind, breakingType            string
	host, barrel, source, control *server.Object
	victims                       [3]e2ePowderBarrelVictim
	original, center              types.Pointf
	hostHP, hostMax, controlHP    uint16
	controlFrame, start, fuse     uint32
	audio                         int
	active                        bool
	existing                      map[*server.Object]int32
	flames                        []*server.Object
}

func (f *e2ePowderBarrelFixture) prepare() {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.HealthData == nil || f.host.HealthData.Cur == 0 || f.host.UpdateData == nil ||
		f.host.Buffs != 0 || f.host.Poison540 != 0 || noxflags.HasGame(noxflags.GameModeQuest) {
		e2eError(fmt.Errorf("powder barrel requires an unenchanted, live regular-game host"))
		return
	}
	f.original, f.hostHP, f.hostMax = f.host.PosVec, f.host.HealthData.Cur, f.host.HealthData.Max
	origin, direction, err := e2eWarriorAbilityArena(f.original, 96, func(from, to types.Pointf) bool {
		return e2eWarriorLaneClear(f.host, from, to)
	})
	if err != nil {
		e2eError(err)
		return
	}
	f.center = origin
	npc, monster := noxServer.NewObjectByTypeID("NPC"), noxServer.NewObjectByTypeID("Troll")
	f.control, f.barrel = noxServer.NewObjectByTypeID("Troll"), noxServer.NewObjectByTypeID(f.kind)
	if npc == nil || monster == nil || f.control == nil || f.barrel == nil {
		e2eError(fmt.Errorf("powder barrel requires stock NPC, Troll and %s definitions", f.kind))
		return
	}
	side := types.Ptf(-direction.Y, direction.X)
	asObjectS(f.host).SetPos(origin.Add(direction.Mul(64)))
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	noxServer.CreateObjectAt(monster, nil, origin.Sub(direction.Mul(64)))
	noxServer.CreateObjectAt(npc, nil, origin.Add(side.Mul(64)))
	noxServer.CreateObjectAt(f.control, nil, origin.Add(direction.Mul(128)))
	noxServer.CreateObjectAt(f.barrel, nil, origin)
	noxServer.ObjectsAddPending()
	f.victims = [3]e2ePowderBarrelVictim{
		{name: "player", unit: f.host, armored: true},
		{name: "monster", unit: monster},
		{name: "NPC", unit: npc, armored: true},
	}
	for _, unit := range []*server.Object{f.host, npc, monster, f.control} {
		if !e2eObjectInWorld(unit) || unit.UpdateData == nil || unit.HealthData == nil || unit.Buffs != 0 {
			e2eError(fmt.Errorf("powder barrel unit lacks a real stock live record: %p", unit))
			return
		}
		asObjectS(unit).SetMaxHealth(2000)
		if unit != f.host {
			// Ordinary passive properties/WAIT are placement setup; no
			// enemy selection, hit, callback, packet or result is injected.
			asObjectS(unit).SetAggression(0)
			asObjectS(unit).SetRetreatLevel(0)
			unit.ClearActionStack()
			unit.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
		}
		for item := unit.FirstItem(); item != nil; item = item.NextItem() {
			if !item.Flags().Has(object.FlagEquipped) || item.InitData == nil ||
				!item.Class().HasAny(object.ClassArmor|object.ClassWeapon|object.ClassWand) {
				continue
			}
			for _, modifier := range item.InitDataModifier().Modifiers {
				if modifier != nil && (modifier.Defend76.Fnc != nil || modifier.Engage112 != nil) {
					e2eError(fmt.Errorf("powder barrel baseline has equipped protection"))
					return
				}
			}
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 || uintptr(unit.UpdateData) <= math.MaxUint32) {
			e2eError(fmt.Errorf("powder barrel unit/update is not above 4 GiB"))
			return
		}
	}
	if npc.Damage != f.host.Damage || monster.Damage == f.host.Damage ||
		!npc.SubClass().AsMonster().Has(object.MonsterNPC) || monster.SubClass().AsMonster().Has(object.MonsterNPC) ||
		f.barrel.Damage == nil || f.barrel.HealthData == nil || f.barrel.HealthData.Cur != 7 {
		e2eError(fmt.Errorf("powder barrel lost stock PlayerDamage/DefaultDamage/FlammableDamage records"))
		return
	}
	f.existing = make(map[*server.Object]int32)
	for unit := noxServer.Objs.First(); unit != nil; unit = unit.Next() {
		f.existing[unit] = unit.ScriptIDVal
	}
	noxServer.Audio.OnSound(func(id sound.ID, kind int, owner *server.Object, _ types.Pointf) {
		if f.active && id == sound.SoundPowderBarrelExplode && owner == f.barrel {
			f.audio++
			if kind != 0 || f.audio != 1 {
				e2eError(fmt.Errorf("powder barrel explosion sound kind/count=%d/%d", kind, f.audio))
			}
		}
	})
	noxServer.TickHook(f.observeHit)
}

func (f *e2ePowderBarrelFixture) ignite() {
	for i := range f.victims {
		v := &f.victims[i]
		if noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(v.unit))) == nil ||
			!noxServer.MapTraceRay(f.center, v.unit.PosVec, server.MapTraceFlag1) {
			e2eError(fmt.Errorf("powder barrel %s placement/client publication is not ready", v.name))
			return
		}
		v.health = v.unit.HealthData.Cur
		delta := v.unit.PosVec.Sub(f.center)
		distance := float32(math.Sqrt(float64(delta.X*delta.X + delta.Y*delta.Y)))
		raw := e2ePowderBarrelRadialDamage(distance)
		armor, carry := float32(0), float32(0)
		if v.name == "player" {
			ud := v.unit.UpdateDataPlayer()
			armor, carry = math.Float32frombits(ud.Field57), math.Float32frombits(ud.Field21)
		} else if v.armored {
			ud := v.unit.UpdateDataMonster()
			armor, carry = math.Float32frombits(ud.Field518), math.Float32frombits(ud.Field1)
		}
		immune := v.unit.Class().Has(object.ClassMonster) && uint32(v.unit.SubClass())&0x400 != 0
		v.damage, v.carry = e2ePowderBarrelUnitDamage(raw, v.armored, immune, armor, carry)
		if distance < 60 || distance > 68 || v.damage <= 0 || v.damage >= int32(v.health) {
			e2eError(fmt.Errorf("powder barrel %s baseline drifted: distance=%g raw=%d effective=%d HP=%d", v.name, distance, raw, v.damage, v.health))
			return
		}
		e2eLog.Printf("POWDER BARREL PREPARED: kind=%s target=%s object=%p update=%p distance=%g raw=%d armor=%g carry=%g expected=%d HP=%d",
			f.kind, v.name, v.unit, v.unit.UpdateData, distance, raw, armor, carry, v.damage, v.health)
	}
	f.controlHP, f.controlFrame = f.control.HealthData.Cur, f.control.Frame134
	f.start, f.active = noxServer.Frame(), true
	// Ignite only the barrel with the ordinary script damage entry. Its real
	// FlammableDamage/death spawn and natural 1..5-tick fuse must perform
	// the radial damage, knockback and flames; never call them from here.
	if !asObjectS(f.barrel).DoDamage(f.host, 1, object.DamageFlame) {
		e2eError(fmt.Errorf("powder barrel ignition was rejected"))
		return
	}
	noxServer.ObjectsAddPending()
	for unit := noxServer.Objs.First(); unit != nil; unit = unit.Next() {
		if !f.newObject(unit) || unit.ObjectTypeC().ID() != f.breakingType || unit.PosVec != f.center {
			continue
		}
		if f.source != nil || unit.UpdateData != nil || unit.ObjOwner != nil ||
			unit.Class()&(object.ClassSimple|object.ClassLight) != object.ClassSimple|object.ClassLight ||
			!unit.Flags().Has(object.FlagNoCollide) || unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
			e2eError(fmt.Errorf("powder barrel death spawn lost its unowned native SIMPLE|LIGHT/no-update shape"))
			return
		}
		f.source = unit
	}
	if f.source == nil || f.audio != 1 {
		e2eError(fmt.Errorf("powder barrel produced no real breaking object/audio: source=%p count=%d", f.source, f.audio))
	}
}

func (f *e2ePowderBarrelFixture) observeHit() {
	if !f.active || f.source == nil {
		return
	}
	if f.control.HealthData.Cur != f.controlHP || f.control.Frame134 != f.controlFrame {
		e2eError(fmt.Errorf("powder barrel damaged the outside-radius control"))
		return
	}
	if f.fuse == 0 && f.source.Field34 != f.source.Field32 {
		f.fuse = f.source.Field34
		if f.fuse <= f.start || f.fuse-f.source.Field32 > 6 {
			e2eError(fmt.Errorf("powder barrel fuse was not naturally scheduled: creation=%d fuse=%d start=%d", f.source.Field32, f.fuse, f.start))
			return
		}
	}
	for i := range f.victims {
		v := &f.victims[i]
		if v.hit || v.unit.HealthData.Cur == v.health {
			continue
		}
		if v.unit.HealthData.Cur != v.health-uint16(v.damage) || v.unit.Obj130 != f.source ||
			v.unit.Field131 != uint32(object.DamageExplosion) || v.unit.Pos132 != f.source.PrevPos || v.unit.Frame134 != f.fuse {
			e2eError(fmt.Errorf("powder barrel %s hit mismatch: HP=%d->%d want=%d source=%p/%p type=%d frame=%d fuse=%d",
				v.name, v.health, v.unit.HealthData.Cur, v.health-uint16(v.damage), v.unit.Obj130, f.source, v.unit.Field131, v.unit.Frame134, f.fuse))
			return
		}
		if v.armored {
			var marker, typ, carry uint32
			if v.name == "player" {
				ud := v.unit.UpdateDataPlayer()
				marker, typ, carry = ud.Field76, ud.Field75, ud.Field21
			} else {
				ud := v.unit.UpdateDataMonster()
				marker, typ, carry = ud.Field547, ud.Field546, ud.Field1
			}
			if marker != 2 || typ != 7 || carry != v.carry {
				e2eError(fmt.Errorf("powder barrel %s armor marker/carry mismatch: %d/%d %#x/%#x", v.name, marker, typ, carry, v.carry))
				return
			}
		}
		if v.unit.Class().Has(object.ClassMonster) {
			mask := object.MonStatusInjured | object.MonStatusOnFire
			if v.unit.UpdateDataMonster().StatusFlags&mask != mask {
				e2eError(fmt.Errorf("powder barrel %s lacks injured/on-fire status", v.name))
				return
			}
		}
		v.hit = true
		e2eLog.Printf("POWDER BARREL HIT: kind=%s target=%s HP=%d->%d damage=%d type=%d source=%p natural-fuse=%d elapsed=%d",
			f.kind, v.name, v.health, v.unit.HealthData.Cur, v.damage, v.unit.Field131, f.source, f.fuse, noxServer.Frame()-f.start)
	}
}

func (f *e2ePowderBarrelFixture) complete() bool {
	if f.audio != 1 || f.fuse == 0 {
		return false
	}
	for i := range f.victims {
		v := &f.victims[i]
		if !v.hit {
			return false
		}
		if v.name == "player" {
			meter, ready := e2eClientHUDMeter(0)
			if !e2eFireballHUDHit(meter, ready, v.health, v.damage, f.hostMax) {
				return false
			}
		} else {
			wire := uint16(noxServer.GetUnitNetCode(v.unit))
			delta, received := legacy.HealthChangeForDrawable(uint32(wire))
			if noxClient.Objs.ByNetCode(wire) == nil || !received || int32(delta) != -v.damage {
				return false
			}
		}
	}
	for unit := noxServer.Objs.First(); unit != nil; unit = unit.Next() {
		if !f.newObject(unit) || unit.Flags().Has(object.FlagDestroyed) {
			continue
		}
		id := unit.ObjectTypeC().ID()
		if (id == "SmallFlame" || id == "MediumFlame") && unit.PosVec.Sub(f.center).Len() <= 26 {
			f.flames = append(f.flames, unit)
		}
	}
	if len(f.flames) != 4 {
		e2eError(fmt.Errorf("powder barrel spawned %d real flames, want 4", len(f.flames)))
		return false
	}
	path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
	if err != nil {
		e2eError(err)
		return false
	}
	e2eLog.Printf("POWDER BARREL PASS: kind=%s player/monster/NPC server/client-HP=verified outside-radius-HP=%d unchanged natural-fuse=%d explosion-audio=1 flames=4 path=%s",
		f.kind, f.control.HealthData.Cur, f.fuse-f.source.Field32, path)
	return true
}

// Object pools can reuse the dead barrel's address for a new flame. Observe
// script identity as well as its pointer instead of excluding a recycled slot.
func (f *e2ePowderBarrelFixture) newObject(unit *server.Object) bool {
	before, present := f.existing[unit]
	return !present || before != unit.ScriptIDVal
}

func (f *e2ePowderBarrelFixture) cleanup() {
	f.active = false
	for _, unit := range []*server.Object{f.victims[1].unit, f.victims[2].unit, f.control, f.source} {
		if e2eObjectInWorld(unit) && !unit.Flags().Has(object.FlagDestroyed) {
			noxServer.DelayedDelete(unit)
		}
	}
	for _, flame := range f.flames {
		noxServer.DelayedDelete(flame)
	}
	asObjectS(f.host).SetPos(f.original)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	asObjectS(f.host).SetMaxHealth(int(f.hostMax))
	asObjectS(f.host).SetHealth(int(f.hostHP))
}

func (sc *e2eScenario) CheckPowderBarrelDamage(kind, name string) {
	breaking, ok := e2ePowderBarrelBreakingType(kind)
	if !ok {
		e2eError(fmt.Errorf("invalid powder barrel fixture: %s", kind))
		return
	}
	f := &e2ePowderBarrelFixture{kind: kind, breakingType: breaking}
	sc.addWhen(0, name+" prepare real barrel and three unit classes", 1200, func() bool {
		host := noxServer.Players.HostUnit()
		return nox_client_isConnected() && noxClient.ClientPlayerUnit() != nil && host != nil && host.Buffs == 0 && host.Poison540 == 0
	}, f.prepare)
	sc.Wait(12, name+" publish units")
	sc.add(0, name+" ignite barrel through ordinary script damage", f.ignite)
	sc.addWhen(1, name+" natural blast and client HP", 300, f.complete, f.cleanup)
	sc.Wait(12, name+" ordinary deletion cleanup")
}
