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

// These fixtures observe ordinary spatial collision, not a direct damage
// call. They never replace callbacks, fabricate damage packets or write HP
// after contact. Passive WAIT is placement setup, not an AI-attack test.
type e2eEnvironmentFixture struct {
	kind, victim               string
	host, unit, hazard         *server.Object
	original, contact, outside types.Pointf
	hostHP, hostMax, health    uint16
	baseline, start            uint32
	typ                        object.DamageType
	raw, damage                int32
	carry                      uint32
	data                       server.DamageCollideData
	switched, active, hit      bool
}

func (f *e2eEnvironmentFixture) prepare() {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.HealthData == nil || f.host.UpdateData == nil || f.host.Buffs != 0 || f.host.Poison540 != 0 {
		e2eError(fmt.Errorf("environment fixture requires a live unenchanted host"))
		return
	}
	f.original, f.hostHP, f.hostMax = f.host.PosVec, f.host.HealthData.Cur, f.host.HealthData.Max
	if f.victim == "player" && f.host.UpdateDataPlayer().Player.ArmorEquip&0x3000000 != 0 {
		e2eError(fmt.Errorf("environment HP-loss fixture requires an unshielded player; ordinary shield blocks are valid"))
		return
	}
	f.unit = f.host
	if f.victim != "player" {
		id := "Troll"
		if f.victim == "NPC" {
			id = "NPC"
		}
		f.unit = noxServer.NewObjectByTypeID(id)
		if f.unit == nil {
			e2eError(fmt.Errorf("missing stock environment victim %s", id))
			return
		}
	}
	f.typ, f.raw = object.DamageLava, 2
	if f.kind == "Lava" {
		first, found := e2eFindLavaTile()
		if !found || legacy.Nox_xxx_tileNFromPoint_411160(f.original) == 6 {
			e2eError(fmt.Errorf("environment fixture requires a lava tile and safe host baseline"))
			return
		}
		var err error
		f.contact, f.outside, err = e2eEnvironmentLavaArena(first, f.unit.Shape.Circle.R+4, legacy.Nox_xxx_tileNFromPoint_411160, func(a, b types.Pointf) bool { return e2eWarriorLaneClear(f.host, a, b) })
		if err != nil {
			e2eError(err)
			return
		}
	} else {
		f.hazard = noxServer.NewObjectByTypeID(f.kind)
		if f.hazard == nil || f.hazard.CollideData == nil {
			e2eError(fmt.Errorf("missing stock environment hazard %s", f.kind))
			return
		}
		f.data = *(*server.DamageCollideData)(f.hazard.CollideData)
		f.typ = object.DamageType(f.data.DamageType)
		f.raw = e2eEnvironmentCollisionDamage(f.data.Damage, noxServer.Frame())
		if f.kind == "Flame" {
			if f.typ != object.DamageFlame || f.data.Damage != 3 {
				e2eError(fmt.Errorf("stock Flame damage record drifted"))
				return
			}
		} else {
			raw, switched, ok := e2ePlayerSpikeKind(f.kind)
			if !ok || f.typ != object.DamageImpale || f.data.Damage != raw {
				e2eError(fmt.Errorf("stock spike damage record drifted: %s", f.kind))
				return
			}
			f.switched = switched
		}
		collide, size, registered := server.ObjectCollideHandler("DamageCollide")
		if !registered || size != 8 || f.hazard.Collide != collide || f.hazard.Shape != f.hazard.ObjectTypeC().Shape || f.hazard.Class() != f.hazard.ObjectTypeC().Class() {
			e2eError(fmt.Errorf("environment stock collider/shape/class changed: %s", f.kind))
			return
		}
		extent := f.hazard.Shape.Circle.R
		if f.hazard.Shape.Kind == server.ShapeKindBox {
			extent = float32(math.Hypot(float64(f.hazard.Shape.Box.W), float64(f.hazard.Shape.Box.H)) / 2)
		}
		origin, direction, err := e2ePlayerSpikeArena(f.original, extent+f.unit.Shape.Circle.R+6, func(a, b types.Pointf) bool { return e2eWarriorLaneClear(f.host, a, b) })
		if err != nil {
			e2eError(err)
			return
		}
		f.contact, f.outside = origin.Add(direction.Mul(6)), origin.Add(direction.Mul(80))
		if f.switched {
			asObjectS(f.hazard).Enable(false)
		}
		noxServer.CreateObjectAt(f.hazard, nil, origin)
	}
	if f.unit != f.host {
		// Keep the ordinary player/camera near the verified arena, outside
		// the hazard. A clear remote footprint alone need not be visible
		// through the original map's walls from the spawn position.
		asObjectS(f.host).SetPos(f.outside)
		f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
		noxServer.CreateObjectAt(f.unit, nil, f.outside)
		asObjectS(f.unit).SetAggression(0)
		asObjectS(f.unit).SetRetreatLevel(0)
		f.unit.ClearActionStack()
		f.unit.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
	}
	noxServer.ObjectsAddPending()
	if !e2eObjectInWorld(f.unit) || f.unit.HealthData == nil || f.unit.UpdateData == nil || f.unit.Buffs != 0 || f.unit.Flags().HasAny(object.FlagDead|object.FlagNoUpdate|object.FlagNoCollide) {
		e2eError(fmt.Errorf("environment stock victim is not collision-ready: %s flags=%x", f.victim, f.unit.Flags()))
		return
	}
	if f.victim == "NPC" && (f.unit.Damage != f.host.Damage || !f.unit.SubClass().AsMonster().Has(object.MonsterNPC)) ||
		f.victim == "monster" && (f.unit.Damage == f.host.Damage || f.unit.SubClass().AsMonster().Has(object.MonsterNPC)) {
		e2eError(fmt.Errorf("environment stock victim damage registration changed: %s", f.victim))
		return
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(f.unit)) <= math.MaxUint32 || uintptr(f.unit.UpdateData) <= math.MaxUint32) {
		e2eError(fmt.Errorf("environment victim/update must retain native pointer width"))
		return
	}
	asObjectS(f.unit).SetMaxHealth(2000)
	pos := f.outside
	if f.switched {
		pos = f.contact
	}
	asObjectS(f.unit).SetPos(pos)
	f.unit.VelVec, f.unit.ForceVec, f.unit.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	legacy.Nox_xxx_unitHasCollideOrUpdateFn_537610(f.unit)
	f.health, f.baseline = f.unit.HealthData.Cur, f.unit.Frame134
	noxServer.TickHook(f.observe)
}

func (f *e2eEnvironmentFixture) startContact() {
	if f.unit.HealthData.Cur != f.health || f.unit.Frame134 != f.baseline || f.switched && (f.hazard.IsEnabled() || !f.hazard.Flags().Has(object.FlagNoCollide)) {
		e2eError(fmt.Errorf("environment safe/disabled baseline changed: %s/%s HP=%d/%d frame=%d/%d", f.kind, f.victim, f.unit.HealthData.Cur, f.health, f.unit.Frame134, f.baseline))
		return
	}
	armor, carry := float32(0), float32(0)
	if f.victim == "player" {
		ud := f.unit.UpdateDataPlayer()
		armor, carry = math.Float32frombits(ud.Field57), math.Float32frombits(ud.Field21)
	} else {
		ud := f.unit.UpdateDataMonster()
		armor, carry = math.Float32frombits(ud.Field518), math.Float32frombits(ud.Field1)
	}
	f.damage, f.carry = f.raw, math.Float32bits(carry)
	if f.typ == object.DamageImpale {
		f.damage, f.carry = e2ePowderBarrelUnitDamage(f.raw, f.victim != "monster", false, armor, carry)
	}
	if f.switched {
		asObjectS(f.hazard).Enable(true)
	}
	f.health, f.start, f.active = f.unit.HealthData.Cur, noxServer.Frame(), true
	asObjectS(f.unit).SetPos(f.contact)
	f.unit.NewPos, f.unit.PrevPos = f.contact, f.contact
	f.unit.VelVec, f.unit.ForceVec, f.unit.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	legacy.Nox_xxx_unitHasCollideOrUpdateFn_537610(f.unit)
	e2eLog.Printf("ENVIRONMENT CONTACT: kind=%s target=%s unit=%p update=%p collider=%p type=%d stock-byte=%d incoming=%d armor=%g carry=%g expected=%d HP=%d safe/disabled-baseline=passed", f.kind, f.victim, f.unit, f.unit.UpdateData, f.hazard, f.typ, f.data.Damage, f.raw, armor, carry, f.damage, f.health)
	f.logContactGeometry()
}

func (f *e2eEnvironmentFixture) observe() {
	if !f.active || f.hit {
		return
	}
	if f.unit != f.host && f.host.HealthData.Cur != f.hostHP {
		e2eError(fmt.Errorf("environment spectator was damaged: %s/%s HP=%d/%d", f.kind, f.victim, f.host.HealthData.Cur, f.hostHP))
		return
	}
	if f.unit.HealthData.Cur >= f.health {
		elapsed := noxServer.Frame() - f.start
		if elapsed == 1 || elapsed == 10 || elapsed == 30 || elapsed == 60 {
			f.logContactGeometry()
		}
		return
	}
	u := f.unit
	marker, kind, carry := uint32(0), uint32(0), uint32(0)
	wantKind := uint32(f.typ)
	if f.victim == "player" {
		ud := u.UpdateDataPlayer()
		marker, kind, carry = ud.Field76, ud.Field75, ud.Field21
	} else {
		ud := u.UpdateDataMonster()
		marker, kind, carry = ud.Field547, ud.Field546, ud.Field1
	}
	if u.HealthData.Cur != f.health-uint16(f.damage) || u.Obj130 != f.hazard || u.Field131 != uint32(f.typ) || u.Frame134 < f.start || marker != 2 || kind != wantKind || carry != f.carry {
		e2eError(fmt.Errorf("environment actual hit mismatch: %s/%s HP=%d/%d source=%p/%p type=%d marker=%x/%x carry=%x/%x frame=%d/%d", f.kind, f.victim, u.HealthData.Cur, f.health-uint16(f.damage), u.Obj130, f.hazard, u.Field131, marker, kind, carry, f.carry, u.Frame134, f.start))
		return
	}
	if f.hazard != nil && *(*server.DamageCollideData)(f.hazard.CollideData) != f.data {
		e2eError(fmt.Errorf("stock environment damage record was altered"))
		return
	}
	f.hit = true
	e2eLog.Printf("ENVIRONMENT ACTUAL HIT: kind=%s target=%s HP=%d->%d loss=%d marker=%x/%x frame=%d native-source=%p stock-collision-data=unchanged", f.kind, f.victim, f.health, u.HealthData.Cur, f.damage, marker, kind, u.Frame134, u.Obj130)
	// Stop the fixture only after one complete natural collision has been
	// observed, before another tick can hide its exact damage/carry result.
	asObjectS(u).SetPos(f.outside)
	u.NewPos, u.PrevPos = f.outside, f.outside
	u.VelVec, u.ForceVec, u.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	if f.hazard != nil {
		noxServer.DelayedDelete(f.hazard)
	}
}

func (f *e2eEnvironmentFixture) logContactGeometry() {
	u := f.unit
	e2eLog.Printf("ENVIRONMENT GEOMETRY: kind=%s target=%s elapsed=%d position=%v next=%v previous=%v bounds=%v flags=%x HP=%d source=%p type=%d hit-frame=%d", f.kind, f.victim, noxServer.Frame()-f.start, u.PosVec, u.NewPos, u.PrevPos, u.CollideP1, uint32(u.Flags()), u.HealthData.Cur, u.Obj130, u.Field131, u.Frame134)
	if f.victim == "player" {
		ud := u.UpdateDataPlayer()
		e2eLog.Printf("ENVIRONMENT PLAYER GEOMETRY: marker=%x/%x carry=%x state=%d flags3680=%x equipment=%x/%x buffs=%x", ud.Field76, ud.Field75, ud.Field21, ud.State, ud.Player.Field3680, ud.Player.WeaponEquip, ud.Player.ArmorEquip, u.Buffs)
	}
	if f.hazard != nil {
		h := f.hazard
		e2eLog.Printf("ENVIRONMENT HAZARD GEOMETRY: kind=%s position=%v next=%v previous=%v bounds=%v shape=%+v flags=%x class=%x owner=%p terminal=%p collide=%p data=%+v", f.kind, h.PosVec, h.NewPos, h.PrevPos, h.CollideP1, h.Shape, uint32(h.Flags()), uint32(h.Class()), h.ObjOwner, h.FindOwnerChainPlayer(), h.Collide, *(*server.DamageCollideData)(h.CollideData))
	}
}

func (f *e2eEnvironmentFixture) clientResult() bool {
	if !f.hit {
		return false
	}
	if f.victim == "player" {
		meter, ready := e2eClientHUDMeter(0)
		// Script SetMaxHealth makes the fixture durable without publishing
		// a new HUD maximum. Require the original maximum and exact current
		// HP from the ordinary absolute-health report, as for Fireball.
		if !e2eFireballHUDHit(meter, ready, f.health, f.damage, f.hostMax) {
			return false
		}
	} else {
		wire := uint16(noxServer.GetUnitNetCode(f.unit))
		delta, got := legacy.HealthChangeForDrawable(uint32(wire))
		if noxClient.Objs.ByNetCode(wire) == nil || !got || int32(delta) != -f.damage {
			return false
		}
	}
	path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
	if err != nil {
		e2eError(err)
		return true
	}
	e2eLog.Printf("ENVIRONMENT PASS: kind=%s target=%s natural-collision/server-HP/client-HP/marker/carry=verified frame=%s", f.kind, f.victim, path)
	return true
}

func (f *e2eEnvironmentFixture) cleanup() {
	f.active = false
	if f.unit != f.host {
		noxServer.DelayedDelete(f.unit)
	}
	asObjectS(f.host).SetPos(f.original)
	f.host.NewPos, f.host.PrevPos = f.original, f.original
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	asObjectS(f.host).SetMaxHealth(int(f.hostMax))
	asObjectS(f.host).SetHealth(int(f.hostHP))
}

func (sc *e2eScenario) CheckEnvironmentDamage(kind, name string) {
	if _, _, ok := e2ePlayerSpikeKind(kind); !ok && kind != "Flame" && kind != "Lava" {
		e2eError(fmt.Errorf("invalid environment fixture %s", kind))
		return
	}
	for _, victim := range []string{"player", "monster", "NPC"} {
		f := &e2eEnvironmentFixture{kind: kind, victim: victim}
		label := name + " " + victim
		sc.addWhen(0, label+" prepare original hazard/victim", 1200, func() bool {
			u := noxServer.Players.HostUnit()
			return nox_client_isConnected() && noxClient.ClientPlayerUnit() != nil && u != nil && u.Buffs == 0 && u.Poison540 == 0
		}, f.prepare)
		sc.Wait(6, label+" safe baseline/publication")
		sc.add(0, label+" ordinary contact", f.startContact)
		sc.addWhen(1, label+" natural hit and client HP", 120, f.clientResult, f.cleanup)
		sc.Wait(6, label+" ordinary cleanup")
	}
}

// This is an independent expectation for GAME.EXE 004E9430, not a call to
// the collision implementation. The selected stock hazards use bytes 2, 3
// or 8, so their incoming damage does not depend on the frame parity.
func e2eEnvironmentCollisionDamage(stockByte uint8, frame uint32) int32 {
	if stockByte == 1 {
		return int32(frame & 1)
	}
	return int32(stockByte / 2)
}

// Lava must have a nearby safe spectator/cleanup point with an unobstructed
// view. This is placement only: floor tokens, damage and networking are never
// supplied by the fixture. Bound the search around the first stock lava tile.
func e2eEnvironmentLavaArena(first types.Pointf, radius float32, tile func(types.Pointf) int, clear func(types.Pointf, types.Pointf) bool) (types.Pointf, types.Pointf, error) {
	if radius <= 0 || radius > 32 || math.IsNaN(float64(radius)) || tile == nil || clear == nil {
		return types.Pointf{}, types.Pointf{}, fmt.Errorf("invalid lava footprint %g", radius)
	}
	for ring := 0; ring <= 230; ring += 23 {
		for y := -ring; y <= ring; y += 23 {
			for x := -ring; x <= ring; x += 23 {
				if ring != 0 && x != -ring && x != ring && y != -ring && y != ring {
					continue
				}
				contact := first.Add(types.Ptf(float32(x), float32(y)))
				if tile(contact) != 6 {
					continue
				}
				for dir := 0; dir < 256; dir += 32 {
					cx, cy := server.SinCosDir(byte(dir))
					outside := contact.Add(types.Ptf(cx*80, cy*80))
					if tile(outside) == 6 || !clear(contact, outside) {
						continue
					}
					valid := true
					for around := 0; around < 256; around += 16 {
						ax, ay := server.SinCosDir(byte(around))
						delta := types.Ptf(max(radius, 16)*ax, max(radius, 16)*ay)
						if !clear(contact, contact.Add(delta)) || !clear(outside, outside.Add(delta)) || tile(outside.Add(delta)) == 6 {
							valid = false
							break
						}
					}
					if valid {
						return contact, outside, nil
					}
				}
			}
		}
	}
	return types.Pointf{}, types.Pointf{}, fmt.Errorf("no visible lava/safe footprint near %v", first)
}
