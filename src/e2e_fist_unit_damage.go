package opennox

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func e2eFistUnitDirection(direction string) (fromNPC, ok bool) {
	switch direction {
	case "player-to-npc":
		return false, true
	case "npc-to-player":
		return true, true
	default:
		return false, false
	}
}

// CRUSH spills the scaled damage to binary32 before adding the unit's live
// fractional carry and rounding to even. Read this expectation independently
// of the production damage service; never replace the service or set its result.
func e2eFistUnitExpectedDamage(damage int32, armor, carry float32) (int32, float32) {
	scaled := float32((1 - 0.5*float64(armor)) * float64(damage))
	accumulated := scaled + carry
	effective := int32(math.RoundToEven(float64(accumulated)))
	nextCarry := accumulated - float32(effective)
	if damage > 0 && effective == 0 {
		effective = 1
	}
	return effective, nextCarry
}

type e2eFistUnitFixture struct {
	e2eFistFixture
	fromNPC         bool
	direction       string
	host, npc       *server.Object
	hostHP, hostMax uint16
	expectedDamage  int32
	expectedCarry   float32
	hitHP           uint16
	clientHit       bool
}

func (f *e2eFistUnitFixture) prepareUnits() {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.ControllingPlayer() == nil || f.host.HealthData == nil || f.host.HealthData.Cur == 0 ||
		f.host.Flags().HasAny(object.FlagDead|object.FlagDestroyed|object.FlagNoUpdate) {
		e2eError(fmt.Errorf("Fist unit fixture requires a live host"))
		return
	}
	f.original, f.hostHP, f.hostMax = f.host.PosVec, f.host.HealthData.Cur, f.host.HealthData.Max
	origin, direction, err := e2eWarriorAbilityArena(f.original, f.host.Shape.Circle.R+16, func(from, to types.Pointf) bool {
		return e2eWarriorLaneClear(f.host, from, to)
	})
	if err != nil {
		e2eError(err)
		return
	}
	f.npc = noxServer.NewObjectByTypeID("NPC")
	if f.npc == nil {
		e2eError(fmt.Errorf("Fist unit fixture has no stock NPC type"))
		return
	}
	asObjectS(f.host).SetPos(origin)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	noxServer.CreateObjectAt(f.npc, nil, origin.Add(direction.Mul(112)))
	noxServer.ObjectsAddPending()
	if !e2eObjectInWorld(f.npc) || f.npc.HealthData == nil || f.npc.UpdateData == nil || f.npc.Damage == nil ||
		f.npc.Damage != f.npc.ObjectTypeC().Damage || f.npc.Damage != f.host.Damage ||
		!f.npc.Class().Has(object.ClassMonster) || uint32(f.npc.SubClass())&0x10 == 0 {
		e2eError(fmt.Errorf("Fist stock NPC was not initialized with its PlayerDamage callback"))
		return
	}
	// Durable, positioned units and ordinary waiting AI are fixture setup.
	// No spell, damage, hit marker, physics state, packet or result is injected.
	// This is not an autonomous NPC casting/target-acquisition test.
	asObjectS(f.npc).SetMaxHealth(2000)
	f.npc.UpdateDataMonster().SetAggression(0)
	f.npc.ClearActionStack()
	f.npc.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
	f.unit, f.prop = f.host, f.npc
	if f.fromNPC {
		f.unit, f.prop = f.npc, f.host
		asObjectS(f.host).SetMaxHealth(2000)
	}
	if len(f.ownedFists()) != 0 || f.prop.HasEnchant(23) || f.prop.HasEnchant(26) {
		e2eError(fmt.Errorf("Fist unit fixture is not ready for an unshielded cast"))
		return
	}
	// The real starting Wizard armor is retained. Other equipped Defend effects
	// need separate expectations and must not silently bypass this HP assertion.
	for item := f.prop.FirstItem(); item != nil; item = item.NextItem() {
		if !item.Flags().Has(object.FlagEquipped) || !item.Class().HasAny(object.ClassArmor|object.ClassWeapon|object.ClassWand) || item.InitData == nil {
			continue
		}
		for _, mod := range item.InitDataModifier().Modifiers {
			if mod != nil && mod.Defend76.Fnc != nil {
				e2eError(fmt.Errorf("Fist unit fixture has an unexpected equipped Defend effect"))
				return
			}
		}
	}
	f.health = f.prop.HealthData.Cur
	balance := noxServer.Balance.FloatInd("FistOfVengeanceDamage", f.level-1)
	f.damage = int32(math.RoundToEven(float64(float32(balance))))
	var armor, carry float32
	if f.fromNPC {
		ud := f.prop.UpdateDataPlayer()
		armor, carry = math.Float32frombits(ud.Field57), math.Float32frombits(ud.Field21)
	} else {
		ud := f.prop.UpdateDataMonster()
		armor, carry = math.Float32frombits(ud.Field518), math.Float32frombits(ud.Field1)
	}
	f.expectedDamage, f.expectedCarry = e2eFistUnitExpectedDamage(f.damage, armor, carry)
	if f.damage <= 0 || f.expectedDamage <= 0 || f.expectedDamage >= int32(f.health) {
		e2eError(fmt.Errorf("Fist unit damage %d/%d exceeds durable health %d", f.damage, f.expectedDamage, f.health))
		return
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, pointer := range []unsafe.Pointer{unsafe.Pointer(f.host), unsafe.Pointer(f.npc), f.host.UpdateData, f.npc.UpdateData} {
			if uintptr(pointer) <= math.MaxUint32 {
				e2eError(fmt.Errorf("Fist unit fixture pointer is below 4 GiB: %p", pointer))
				return
			}
		}
	}
	e2eLog.Printf("FIST UNIT PREPARED: direction=%s level=%d caster=%p target=%p health=%d stock-damage=%d armor=%g carry=%g expected=%d",
		f.direction, f.level, f.unit, f.prop, f.health, f.damage, armor, carry, f.expectedDamage)
}

func (f *e2eFistUnitFixture) castUnitSpell() {
	f.cast()
	owned := f.ownedFists()
	if len(owned) != 1 {
		e2eError(fmt.Errorf("Fist unit cast produced %d owned projectiles, want one", len(owned)))
		return
	}
	f.fist = owned[0]
	f.wire, f.scriptID, f.frame = f.fist.NetCode, f.fist.ScriptIDVal, noxServer.Frame()
	data := (*server.FistUpdateData)(f.fist.UpdateData)
	if !e2eFistInWorld(f.fist, f.wire, f.scriptID) || f.fist.ObjectTypeC().ID() != f.typeID ||
		f.fist.Class() != object.ClassSimple || f.fist.ObjOwner != f.unit || f.fist.PosVec != f.prop.PosVec ||
		data == nil || data.Damage != f.damage || f.fist.ZVal != 255 ||
		f.fist.Field27 != float32(-noxServer.Balance.Float("FistSpeed")) || f.fist.Field29 != 0x41100000 {
		e2eError(fmt.Errorf("Fist unit cast native type/ownership/damage/physics mismatch"))
		return
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(f.fist)) <= math.MaxUint32 || uintptr(f.fist.UpdateData) <= math.MaxUint32) {
		e2eError(fmt.Errorf("Fist unit projectile allocation is below 4 GiB"))
		return
	}
	f.cast()
	if owned = f.ownedFists(); len(owned) != 1 || owned[0] != f.fist {
		e2eError(fmt.Errorf("Fist unit duplicate cast changed ownership"))
		return
	}
	e2eLog.Printf("FIST UNIT CAST: direction=%s level=%d type=%s owner=%p fist=%p wire=%d damage=%d duplicate=refused",
		f.direction, f.level, f.typeID, f.unit, f.fist, f.wire, f.damage)
}

func (f *e2eFistUnitFixture) observeUnitHit() bool {
	if !e2eObjectInWorld(f.prop) || f.prop.HealthData == nil || f.prop.HealthData.Cur == 0 || f.prop.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
		e2eError(fmt.Errorf("Fist durable unit disappeared before hit verification"))
		return true
	}
	if !f.seenHit {
		if f.prop.HealthData.Cur == f.health {
			return false
		}
		if !e2eFistInWorld(f.fist, f.wire, f.scriptID) {
			e2eError(fmt.Errorf("Fist projectile disappeared before unit attribution verification"))
			return true
		}
		var marker, markerType uint32
		var carry float32
		if f.fromNPC {
			ud := f.prop.UpdateDataPlayer()
			marker, markerType, carry = ud.Field76, ud.Field75, math.Float32frombits(ud.Field21)
		} else {
			ud := f.prop.UpdateDataMonster()
			marker, markerType, carry = ud.Field547, ud.Field546, math.Float32frombits(ud.Field1)
			if !ud.StatusFlags.Has(object.MonStatusInjured) {
				e2eError(fmt.Errorf("Fist NPC hit did not set the injured flag"))
				return true
			}
		}
		f.hitHP = f.health - uint16(f.expectedDamage)
		if f.prop.HealthData.Cur != f.hitHP || f.prop.Obj130 != f.fist || f.prop.Field131 != uint32(object.DamageCrush) ||
			marker != 1 || markerType != uint32(f.fist.TypeInd) || math.Abs(float64(carry-f.expectedCarry)) > 1e-6 {
			e2eError(fmt.Errorf("Fist unit hit mismatch: direction=%s level=%d HP=%d->%d want=%d weapon=%p/%p marker=%d/%d carry=%g/%g",
				f.direction, f.level, f.health, f.prop.HealthData.Cur, f.hitHP, f.prop.Obj130, f.fist, marker, markerType, carry, f.expectedCarry))
			return true
		}
		f.seenHit = true
		e2eLog.Printf("FIST UNIT HIT: direction=%s level=%d health=%d->%d damage=%d attribution=%p type=%d marker=%d/%d carry=%g elapsed=%d",
			f.direction, f.level, f.health, f.hitHP, f.expectedDamage, f.prop.Obj130, f.prop.Field131, marker, markerType, carry, noxServer.Frame()-f.frame)
	}
	if f.prop.HealthData.Cur < f.hitHP {
		e2eError(fmt.Errorf("Fist unit took additional damage after its verified collision: %d->%d", f.hitHP, f.prop.HealthData.Cur))
		return true
	}
	if drawable := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.prop))); drawable != nil {
		if delta, ok := legacy.HealthChangeForDrawable(drawable.NetCode32); ok && delta < 0 {
			if int32(delta) != -f.expectedDamage {
				e2eError(fmt.Errorf("Fist client damage mismatch: direction=%s level=%d delta=%d want=%d", f.direction, f.level, delta, -f.expectedDamage))
				return true
			}
			if !f.clientHit {
				e2eLog.Printf("FIST UNIT CLIENT DAMAGE: direction=%s level=%d drawable=%p delta=%d", f.direction, f.level, drawable, delta)
			}
			f.clientHit = true
		}
	}
	return f.seenHit && f.clientHit
}

// CheckFistUnitDamage uses real host startup, stock NPC initialization and the
// regular object-to-position script cast, collision, PlayerDamage, packets and
// client drawing at all five levels. Placement, durable HP and waiting NPC AI
// are fixtures. It does not simulate autonomous casting, incantation/mana,
// unit death, arbitrary spell damage or a particular campaign trigger.
func (sc *e2eScenario) CheckFistUnitDamage(level int, direction, name string) {
	typeID, levelOK := e2eFistType(level)
	fromNPC, directionOK := e2eFistUnitDirection(direction)
	if !levelOK || !directionOK {
		e2eError(fmt.Errorf("invalid Fist unit scenario: level=%d direction=%q", level, direction))
		return
	}
	f := &e2eFistUnitFixture{e2eFistFixture: e2eFistFixture{level: level, typeID: typeID}, fromNPC: fromNPC, direction: direction}
	sc.addWhen(0, name+" prepare units", 1200, func() bool {
		return noxServer.Players.HostUnit() != nil && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, f.prepareUnits)
	sc.Wait(12, name+" publish unit placement and durability")
	sc.add(0, name+" regular script cast", f.castUnitSpell)
	sc.addWhen(1, name+" falling drawable", 120, func() bool {
		f.observeUnitHit()
		return e2eFistInWorld(f.fist, f.wire, f.scriptID) && f.fist.ZVal > 0 && f.fist.ZVal < 255 &&
			noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.fist))) != nil
	}, func() {
		e2eLog.Printf("FIST UNIT FALL: direction=%s level=%d Z=%g drawable=live elapsed=%d", f.direction, f.level, f.fist.ZVal, noxServer.Frame()-f.frame)
	})
	sc.addWhen(1, name+" natural unit collision and client damage", 180, f.observeUnitHit, func() {})
	sc.addWhen(1, name+" natural deletion", 300, func() bool {
		f.observeUnitHit()
		return !e2eFistInWorld(f.fist, f.wire, f.scriptID) && len(f.ownedFists()) == 0 &&
			noxClient.Objs.ByNetCode(uint16(f.wire)) == nil
	}, func() {
		if !f.seenHit || !f.clientHit {
			e2eError(fmt.Errorf("Fist unit deletion lacked a verified server/client hit"))
			return
		}
		e2eLog.Printf("FIST UNIT DELETED: direction=%s level=%d world/owned/drawable=absent health=%d elapsed=%d", f.direction, f.level, f.prop.HealthData.Cur, noxServer.Frame()-f.frame)
		noxServer.DelayedDelete(f.npc)
		asObjectS(f.host).SetPos(f.original)
		if f.fromNPC {
			// Restore only the durable fixture HP after the real hit was verified.
			asObjectS(f.host).SetMaxHealth(int(f.hostMax))
			asObjectS(f.host).SetHealth(int(f.hostHP))
		}
	})
	sc.Wait(3, name+" fixture cleanup")
}
