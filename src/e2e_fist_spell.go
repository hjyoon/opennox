package opennox

import (
	"fmt"
	"math"
	"strings"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	nsp "github.com/opennox/noxscript/ns/v4/spell"

	"github.com/opennox/opennox/v1/server"
)

func e2eFistType(level int) (string, bool) {
	switch level {
	case 1:
		return "SmallFist", true
	case 2:
		return "MediumFist", true
	case 3, 4, 5:
		return "LargeFist", true
	default:
		return "", false
	}
}

type e2eFistPosition struct{ types.Pointf }

func (p e2eFistPosition) Pos() types.Pointf { return p.Pointf }

type e2eFistFixture struct {
	level            int
	typeID           string
	unit, prop, fist *server.Object
	original         types.Pointf
	health           uint16
	damage           int32
	wire             uint32
	scriptID         int32
	frame            uint32
	seenHit          bool
}

func e2eFistInWorld(want *server.Object, wire uint32, scriptID int32) bool {
	// Never dereference the retained pointer after normal delayed deletion.
	// Matching the live identity also rejects reuse of its allocator address.
	for _, obj := range noxServer.Objs.AllObjects() {
		if obj == want && obj.NetCode == wire && obj.ScriptIDVal == scriptID {
			return true
		}
	}
	return false
}

func (f *e2eFistFixture) ownedFists() []*server.Object {
	var out []*server.Object
	for obj := f.unit.Field129; obj != nil; obj = obj.Field128 {
		typ := obj.ObjectTypeC()
		if typ != nil && (typ.ID() == "SmallFist" || typ.ID() == "MediumFist" || typ.ID() == "LargeFist") {
			out = append(out, obj)
		}
	}
	return out
}

func (f *e2eFistFixture) cast() {
	api := noxServer.noxScriptP()
	api.CastSpellLvl(nsp.Spell("SPELL_FIST"), f.level, api.toObj(f.unit), e2eFistPosition{f.prop.PosVec})
	noxServer.ObjectsAddPending()
}

func (f *e2eFistFixture) prepare() {
	f.unit = noxServer.Players.HostUnit()
	if f.unit == nil || f.unit.ControllingPlayer() == nil || f.unit.HealthData == nil || f.unit.HealthData.Cur == 0 ||
		f.unit.Flags().HasAny(object.FlagDead|object.FlagDestroyed|object.FlagNoUpdate) || len(f.ownedFists()) != 0 {
		e2eError(fmt.Errorf("Fist level %d requires a live host without another owned Fist", f.level))
		return
	}
	f.original = f.unit.PosVec
	origin, direction, err := e2eWarriorAbilityArena(f.original, f.unit.Shape.Circle.R+16, func(from, to types.Pointf) bool {
		return e2eWarriorLaneClear(f.unit, from, to)
	})
	if err != nil {
		e2eError(err)
		return
	}
	// A stock non-unit prop isolates cast/physics/collision from the separately
	// incomplete unit-damage and death admissions. Only fixture placement and
	// durability are changed; no cast, damage, Z, flags or results are injected.
	for _, typ := range noxServer.Types.List() {
		if strings.HasPrefix(typ.ID(), "Barrel") && typ.ID() != "BarrelBreaking" &&
			!typ.Class().HasAny(object.MaskUnits) && typ.Health() != nil && typ.Health().Cur != 0 {
			f.prop = noxServer.NewObjectByTypeID(typ.ID())
			break
		}
	}
	if f.prop == nil {
		e2eError(fmt.Errorf("Fist fixture has no stock damageable barrel"))
		return
	}
	asObjectS(f.unit).SetPos(origin)
	f.unit.VelVec, f.unit.ForceVec, f.unit.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	noxServer.CreateObjectAt(f.prop, nil, origin.Add(direction.Mul(112)))
	noxServer.ObjectsAddPending()
	if f.prop.HealthData == nil || f.prop.Damage == nil || !e2eObjectInWorld(f.prop) {
		e2eError(fmt.Errorf("Fist fixture stock prop was not initialized"))
		return
	}
	asObjectS(f.prop).SetMaxHealth(2000)
	f.health = f.prop.HealthData.Cur
	balance := noxServer.Balance.FloatInd("FistOfVengeanceDamage", f.level-1)
	f.damage = int32(math.RoundToEven(float64(float32(balance))))
	if f.damage <= 0 || f.damage >= int32(f.health) {
		e2eError(fmt.Errorf("Fist level %d stock damage %g/%d is outside durable fixture health %d", f.level, balance, f.damage, f.health))
		return
	}
	f.cast()
	owned := f.ownedFists()
	if len(owned) != 1 {
		e2eError(fmt.Errorf("Fist level %d produced %d owned projectiles, want one", f.level, len(owned)))
		return
	}
	f.fist = owned[0]
	f.wire, f.scriptID, f.frame = f.fist.NetCode, f.fist.ScriptIDVal, noxServer.Frame()
	data := (*server.FistUpdateData)(f.fist.UpdateData)
	if !e2eFistInWorld(f.fist, f.wire, f.scriptID) || f.fist.ObjectTypeC().ID() != f.typeID ||
		f.fist.ObjOwner != f.unit || f.fist.PosVec != f.prop.PosVec || data == nil || data.Damage != f.damage ||
		f.fist.ZVal != 255 || f.fist.Field27 != float32(-noxServer.Balance.Float("FistSpeed")) ||
		f.fist.Field29 != 0x41100000 || f.fist.Field5&0x20 == 0 || uint32(f.fist.ObjFlags)&0x800000 == 0 {
		e2eError(fmt.Errorf("Fist level %d native creation mismatch: type=%s owner=%p pos=%v damage=%v Z=%g velocity=%g flags=%#x field5=%#x field29=%#x",
			f.level, f.fist.ObjectTypeC().ID(), f.fist.ObjOwner, f.fist.PosVec, data, f.fist.ZVal, f.fist.Field27, f.fist.ObjFlags, f.fist.Field5, f.fist.Field29))
		return
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, pointer := range []unsafe.Pointer{unsafe.Pointer(f.unit), unsafe.Pointer(f.fist), f.fist.UpdateData} {
			if uintptr(pointer) <= math.MaxUint32 {
				e2eError(fmt.Errorf("Fist actual native allocation is below 4 GiB: %p", pointer))
				return
			}
		}
	}
	// Enter the same script API again before physics advances. The regular
	// native cast must reject this duplicate through the actual owned list.
	f.cast()
	if owned := f.ownedFists(); len(owned) != 1 || owned[0] != f.fist {
		e2eError(fmt.Errorf("Fist level %d duplicate cast changed ownership", f.level))
		return
	}
	e2eLog.Printf("FIST CAST: level=%d type=%s owner=%p fist=%p update=%p wire=%d stock-damage=%d prop=%s health=%d duplicate=refused frame=%d",
		f.level, f.typeID, f.unit, f.fist, f.fist.UpdateData, f.wire, f.damage, f.prop.ObjectTypeC().ID(), f.health, f.frame)
}

func (f *e2eFistFixture) observeHit() bool {
	if f.seenHit {
		return true
	}
	if !e2eObjectInWorld(f.prop) || f.prop.HealthData == nil || f.prop.HealthData.Cur == 0 {
		e2eError(fmt.Errorf("Fist level %d durable prop disappeared before collision verification", f.level))
		return true
	}
	if f.prop.HealthData.Cur == f.health {
		return false
	}
	if f.prop.HealthData.Cur != f.health-uint16(f.damage) || f.prop.Obj130 != f.fist || f.prop.Field131 != uint32(object.DamageCrush) {
		e2eError(fmt.Errorf("Fist level %d collision mismatch: health=%d->%d want=%d weapon=%p/%p damage-type=%d",
			f.level, f.health, f.prop.HealthData.Cur, f.health-uint16(f.damage), f.prop.Obj130, f.fist, f.prop.Field131))
		return true
	}
	f.seenHit = true
	e2eLog.Printf("FIST HIT: level=%d health=%d->%d damage=%d attribution=%p type=%d elapsed=%d",
		f.level, f.health, f.prop.HealthData.Cur, f.damage, f.prop.Obj130, f.prop.Field131, noxServer.Frame()-f.frame)
	return true
}

// CheckFistSpell uses real menu startup and the object-to-position NoxScript
// API, normal selector, creation, physics, collision, packets and client draw.
// It does not execute a stock map's trigger bytecode, player incantation/mana,
// or unit-target damage/death. Invalid scenario levels fail before scheduling.
func (sc *e2eScenario) CheckFistSpell(level int, name string) {
	typeID, ok := e2eFistType(level)
	if !ok {
		e2eError(fmt.Errorf("Fist scenario level %d is outside 1..5", level))
		return
	}
	f := &e2eFistFixture{level: level, typeID: typeID}
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		return noxServer.Players.HostUnit() != nil && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, f.prepare)
	sc.addWhen(1, name+" falling drawable", 120, func() bool {
		f.observeHit()
		return e2eFistInWorld(f.fist, f.wire, f.scriptID) && f.fist.ZVal > 0 && f.fist.ZVal < 255 &&
			noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.fist))) != nil
	}, func() {
		e2eLog.Printf("FIST FALL: level=%d Z=%g velocity=%g drawable=live elapsed=%d", f.level, f.fist.ZVal, f.fist.Field27, noxServer.Frame()-f.frame)
	})
	sc.Screen(fmt.Sprintf("Fist level %d falling", level))
	sc.addWhen(1, name+" natural collision", 180, f.observeHit, func() {})
	sc.Screen(fmt.Sprintf("Fist level %d impact", level))
	sc.addWhen(1, name+" natural deletion", 300, func() bool {
		f.observeHit()
		return !e2eFistInWorld(f.fist, f.wire, f.scriptID) && len(f.ownedFists()) == 0 &&
			noxClient.Objs.ByNetCode(uint16(f.wire)) == nil
	}, func() {
		if !f.seenHit || f.prop.HealthData.Cur != f.health-uint16(f.damage) {
			e2eError(fmt.Errorf("Fist level %d deletion lacked exactly one verified collision", level))
			return
		}
		e2eLog.Printf("FIST DELETED: level=%d world/owned/drawable=absent elapsed=%d", level, noxServer.Frame()-f.frame)
		noxServer.DelayedDelete(f.prop)
		asObjectS(f.unit).SetPos(f.original)
	})
	sc.Wait(3, name+" fixture cleanup")
}
