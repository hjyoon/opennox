package opennox

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// Stock placement and durable, passive targets are setup. Launch, aim,
// collision, damage, poison, timestamps and client reports are game outputs.
// The chapter-eight case never creates, moves, enables or casts for a Tower:
// it waits for the original map's five-second FireFONTrap01 timer.
type e2eTrapProjectileFixture struct {
	kind, victim, projectileID                                      string
	host, unit, source, projectile                                  *server.Object
	original, origin, contact, outside, spectator                   types.Pointf
	hostHP, hostMax, health                                         uint16
	start, baseline, marker, markerType, carry, wire                uint32
	direction                                                       server.Dir16
	typ                                                             object.DamageType
	raw, damage                                                     int32
	active, launched, hit, instantReported, dot, projectileReported bool
	poisonApplied, poisonFirst                                      uint32
	poisonHealth                                                    uint16
	idleInput                                                       e2eLockIdleInput
}

func e2eTrapProjectileKind(kind string) (projectile string, direction server.Dir16, ok bool) {
	switch kind {
	case "ArrowTrap1":
		return "MercArcherArrow", 32, true
	case "ArrowTrap2":
		return "MercArcherArrow", 96, true
	case "Skull1":
		return "StrongFireball", 160, true
	case "Skull2":
		return "StrongFireball", 224, true
	case "Skull3":
		return "StrongFireball", 96, true
	case "Skull4":
		return "StrongFireball", 32, true
	case "Polyp":
		return "ToxicCloud", 0, true
	case "chapter8-tower":
		return "DeathBall", 0, true
	}
	return "", 0, false
}

// Search only along the original SkullInit direction; do not turn a trap to
// fit a convenient lane. The observer can trigger its real enemy scan while
// staying beyond the impact radius and off the straight projectile lane.
func e2eTrapProjectileLane(origin, dir types.Pointf, distance float32, clear func(types.Pointf, types.Pointf) bool) bool {
	side := types.Ptf(-dir.Y, dir.X)
	target := origin.Add(dir.Mul(distance))
	observer := origin.Add(dir.Mul(288)).Add(side.Mul(100))
	outside := origin.Add(side.Mul(192))
	if !clear(origin, target) || !clear(observer, target) || !clear(outside, target) {
		return false
	}
	for _, pos := range []types.Pointf{origin, target, observer, outside} {
		for _, offset := range []types.Pointf{{32, 0}, {-32, 0}, {0, 32}, {0, -32}, {32, 32}, {-32, 32}, {32, -32}, {-32, -32}} {
			if !clear(pos, pos.Add(offset)) {
				return false
			}
		}
	}
	return true
}

func e2eTrapProjectileArena(original types.Pointf, direction server.Dir16, clear func(types.Pointf, types.Pointf) bool) (types.Pointf, error) {
	for ring := 0; ring <= 368; ring += 23 {
		for y := -ring; y <= ring; y += 23 {
			for x := -ring; x <= ring; x += 23 {
				if ring != 0 && x != -ring && x != ring && y != -ring && y != ring {
					continue
				}
				origin := original.Add(types.Ptf(float32(x), float32(y)))
				if e2eTrapProjectileLane(origin, direction.Vec(), 160, clear) {
					return origin, nil
				}
			}
		}
	}
	return types.Pointf{}, fmt.Errorf("no stock-direction trap lane near %v direction=%d", original, direction)
}

// Towers are built into the original corridor walls, unlike movable test
// traps. The stock ColorLight origin can launch from inside that wall, so
// locate the open part of its original flight lane before placing a victim.
// The real timer and projectile must still reach and damage that victim.
func e2eTrapTowerPositions(origin, dir types.Pointf, launchDistance float32, clear func(types.Pointf, types.Pointf) bool) (contact, spectator types.Pointf, ok bool) {
	contact = origin.Add(dir.Mul(60))
	footprint := func(pos types.Pointf) bool {
		for _, offset := range []types.Pointf{{18, 0}, {-18, 0}, {0, 18}, {0, -18}, {18, 18}, {-18, 18}, {18, -18}, {-18, -18}} {
			if !clear(pos, pos.Add(offset)) {
				return false
			}
		}
		return true
	}
	if launchDistance <= 0 || launchDistance >= 60 || !footprint(contact) {
		return contact, types.Pointf{}, false
	}
	lane := false
	for distance := launchDistance; distance < 60; distance += 2 {
		if clear(origin.Add(dir.Mul(distance)), contact) {
			lane = true
			break
		}
	}
	if !lane {
		return contact, types.Pointf{}, false
	}
	side := types.Ptf(-dir.Y, dir.X)
	for along := float32(-96); along <= 96; along += 23 {
		for across := float32(192); across <= 350; across += 23 {
			for _, sign := range []float32{-1, 1} {
				pos := contact.Add(dir.Mul(along)).Add(side.Mul(across * sign))
				if clear(pos, contact) && footprint(pos) {
					return contact, pos, true
				}
			}
		}
	}
	return contact, types.Pointf{}, false
}

func (f *e2eTrapProjectileFixture) place(unit *server.Object, pos types.Pointf) {
	asObjectS(unit).SetPos(pos)
	unit.NewPos, unit.PrevPos = pos, pos
	unit.VelVec, unit.ForceVec, unit.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	legacy.Nox_xxx_unitHasCollideOrUpdateFn_537610(unit)
}

func (f *e2eTrapProjectileFixture) prepare() {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.HealthData == nil || f.host.UpdateData == nil || f.host.Buffs != 0 || f.host.Poison540 != 0 ||
		f.host.UpdateDataPlayer().Player.ArmorEquip&0x3000000 != 0 || f.host.Flags().HasAny(object.FlagDead|object.FlagNoUpdate) {
		e2eError(fmt.Errorf("trap fixture requires a live unshielded, unenchanted host"))
		return
	}
	f.original, f.hostHP, f.hostMax = f.host.PosVec, f.host.HealthData.Cur, f.host.HealthData.Max
	f.projectileID, f.direction, _ = e2eTrapProjectileKind(f.kind)
	clear := func(a, b types.Pointf) bool { return e2eWarriorLaneClear(f.host, a, b) }
	distance := float32(160)
	if f.kind == "chapter8-tower" {
		if !noxflags.HasGame(noxflags.GameModeCoop) || e2eMapBaseName(legacy.Nox_xxx_mapGetMapName_409B40()) != "wiz08e" {
			e2eError(fmt.Errorf("Tower case requires original solo Wiz08e, not a fabricated gameplay mode"))
			return
		}
		distance = 60 // Direct collision before the ball's age>10 radial update.
		for i := 1; i <= 4; i++ {
			tower := noxServer.Objs.GetObjectByID(fmt.Sprintf("FON_Origin%02d", i))
			wp := noxServer.WPs.ByID(fmt.Sprintf("FON_Target%02d", i))
			if tower == nil || wp == nil {
				continue
			}
			e2eLog.Printf("CHAPTER8 ORIGINAL SOURCE: object=%s type=%s class=%x position=%v shape=%+v", tower.ID(), tower.ObjectTypeC().ID(), uint32(tower.Class()), tower.PosVec, tower.Shape)
			direction := server.DirFromVec(wp.PosVec.Sub(tower.PosVec))
			launchDistance := float32(24)
			switch tower.Shape.Kind {
			case server.ShapeKindCenter:
				launchDistance = 4
			case server.ShapeKindCircle:
				launchDistance = tower.Shape.Circle.R + 4
			case server.ShapeKindBox:
				launchDistance = max(tower.Shape.Box.W, tower.Shape.Box.H) + 4
			}
			// Search along the real flight ray without treating its original
			// source as an obstruction. Other solid props stay intact.
			towerClear := func(a, b types.Pointf) bool { return e2eWarriorLaneClear(tower, a, b) }
			contact, spectator, ready := e2eTrapTowerPositions(tower.PosVec, direction.Vec(), launchDistance, towerClear)
			if ready {
				f.source, f.origin, f.direction = tower, tower.PosVec, direction
				f.contact, f.spectator, f.outside = contact, spectator, f.original
				e2eLog.Printf("CHAPTER8 ORIGINAL TOWER: object=%s type=%s position=%v waypoint=%v direction=%d original-script=FireFONTrap01", tower.ID(), tower.ObjectTypeC().ID(), tower.PosVec, wp.PosVec, direction)
				break
			}
		}
		if f.source == nil {
			e2eError(fmt.Errorf("Wiz08e has no clear original Tower/waypoint lane"))
			return
		}
	} else {
		var err error
		f.origin, err = e2eTrapProjectileArena(f.original, f.direction, clear)
		if err != nil {
			e2eError(err)
			return
		}
		f.source = noxServer.NewObjectByTypeID(f.kind)
		if f.source == nil {
			e2eError(fmt.Errorf("missing stock trap %s", f.kind))
			return
		}
		noxServer.CreateObjectAt(f.source, nil, f.origin)
		if f.kind != "Polyp" {
			asObjectS(f.source).Enable(false)
		}
	}
	dir := f.direction.Vec()
	side := types.Ptf(-dir.Y, dir.X)
	if f.kind != "chapter8-tower" {
		f.contact, f.outside = f.origin.Add(dir.Mul(distance)), f.origin.Add(side.Mul(192))
		f.spectator = f.origin.Add(dir.Mul(288)).Add(side.Mul(100))
	}
	if f.kind == "Polyp" {
		f.contact = f.origin.Add(dir.Mul(6))
	}
	f.unit = f.host
	if f.victim != "player" {
		if f.kind == "chapter8-tower" {
			// Publish the victim in the observer's view before the real
			// timer launch. Returning it to the distant chapter spawn can
			// otherwise hide the damage report before it reaches the client.
			f.outside = f.spectator
		}
		id := "Troll"
		if f.kind == "Polyp" {
			id = "Urchin"
		} // Stock Troll is poison immune.
		if f.victim == "NPC" {
			id = "NPC"
		}
		f.unit = noxServer.NewObjectByTypeID(id)
		if f.unit == nil {
			e2eError(fmt.Errorf("missing stock trap victim %s", id))
			return
		}
		noxServer.CreateObjectAt(f.unit, nil, f.outside)
		asObjectS(f.unit).SetAggression(0)
		asObjectS(f.unit).SetRetreatLevel(0)
		f.unit.ClearActionStack()
		f.unit.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
		f.place(f.host, f.spectator)
	}
	noxServer.ObjectsAddPending()
	// The normal pending-object drain calls SkullInit, not CreateObjectAt.
	if f.kind != "Polyp" && f.kind != "chapter8-tower" &&
		(f.source.Direction1 != f.direction || f.source.UpdateDataSkull().ProjectileType != uint32(noxServer.Types.IndByID(f.projectileID))) {
		e2eError(fmt.Errorf("stock SkullInit direction/projectile drifted: %s dir=%d", f.kind, f.source.Direction1))
		return
	}
	if f.unit.HealthData == nil || f.unit.UpdateData == nil || f.unit.Buffs != 0 || f.unit.Poison540 != 0 ||
		f.unit.Flags().HasAny(object.FlagDead|object.FlagNoUpdate|object.FlagNoCollide) ||
		f.victim == "NPC" && (f.unit.Damage != f.host.Damage || !f.unit.SubClass().AsMonster().Has(object.MonsterNPC)) ||
		f.victim == "monster" && f.unit.Damage == f.host.Damage {
		e2eError(fmt.Errorf("stock trap victim is not collision-ready: %s", f.victim))
		return
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(f.unit)) <= math.MaxUint32 || uintptr(f.unit.UpdateData) <= math.MaxUint32 || uintptr(unsafe.Pointer(f.source)) <= math.MaxUint32) {
		e2eError(fmt.Errorf("trap fixture lost native pointer width"))
		return
	}
	asObjectS(f.unit).SetMaxHealth(2000)
	f.place(f.unit, f.outside)
	f.health, f.baseline = f.unit.HealthData.Cur, f.unit.Frame134
	noxServer.TickHook(f.observe)
}

// Independent PE32 armor expectation, including CRUSH's half absorption.
func e2eTrapProjectileUnitDamage(raw int32, typ object.DamageType, armored bool, armor, carry float32) (int32, uint32) {
	if !armored || typ == object.DamageFlame || typ == object.DamagePoison {
		return raw, math.Float32bits(carry)
	}
	if typ == object.DamageCrush {
		armor *= .5
	}
	return e2ePowderBarrelUnitDamage(raw, true, false, armor, carry)
}

// Model the unmodified natural regeneration and first poison tick. TickHook
// sees frame after IncFrame, so completed updates are [applied+1, frame).
// Player regeneration is five minutes to full; Monster/NPC is three minutes.
// This reads no live HP and never suppresses healing or supplies DOT damage.
func e2eTrapProjectilePoisonHP(applied, frame, fps uint32, maximum, initial uint16, player bool) (uint16, bool, bool) {
	first, ok := e2ePoisonFirstTick(applied, 1)
	elapsed := frame - applied
	if !ok || fps == 0 || maximum == 0 || initial <= 1 || initial > maximum || elapsed > 900 || elapsed > first-applied+128 {
		return 0, false, false
	}
	health, injury, dot := int32(initial), applied, false
	for offset := uint32(1); offset < elapsed; offset++ {
		tick := applied + offset
		if player {
			interval := max(uint32(1), 300*fps/uint32(maximum))
			if tick-injury > fps && health < int32(maximum) && tick%interval == 0 {
				health++
			}
		} else {
			health += e2eMeteorShowerRegenAmount(tick, injury, fps, int32(maximum), health)
		}
		if tick == first {
			health--
			injury, dot = tick, true
		}
	}
	return uint16(health), dot, true
}

func (f *e2eTrapProjectileFixture) begin() {
	if f.unit.HealthData.Cur != f.health || f.unit.Frame134 != f.baseline {
		e2eError(fmt.Errorf("trap safe baseline was damaged"))
		return
	}
	_, _, carry := e2eSpellUnitMatrixMarker(f.unit)
	var armor uint32
	if f.victim == "player" {
		armor = f.unit.UpdateDataPlayer().Field57
	} else {
		armor = f.unit.UpdateDataMonster().Field518
	}
	f.raw, f.typ = 96, object.DamageFlame
	switch f.kind {
	case "ArrowTrap1", "ArrowTrap2":
		f.raw, f.typ = int32(math.RoundToEven(noxServer.Balance.Float("ArrowTrapDamage"))), object.DamageImpale
	case "chapter8-tower":
		f.raw, f.typ = int32(math.RoundToEven(noxServer.Balance.Float("DeathBallCollideDamage"))), object.DamageCrush
	case "Polyp":
		f.raw, f.typ = 0, object.DamagePoison
	}
	f.damage, f.carry = e2eTrapProjectileUnitDamage(f.raw, f.typ, f.victim != "monster", math.Float32frombits(armor), math.Float32frombits(carry))
	f.start, f.active = noxServer.Frame(), true
	if f.kind != "chapter8-tower" {
		f.place(f.unit, f.contact)
		if f.kind != "Polyp" {
			asObjectS(f.source).Enable(true)
		}
	}
	e2eLog.Printf("TRAP ARMED: kind=%s target=%s source=%p victim=%p update=%p source-enemy-target=%t incoming=%d expected=%d HP=%d safe-baseline=passed", f.kind, f.victim, f.source, f.unit, f.unit.UpdateData, noxServer.S().IsEnemyTo(f.source, f.unit), f.raw, f.damage, f.health)
}

func (f *e2eTrapProjectileFixture) observeLaunch() {
	if f.projectile != nil {
		return
	}
	if f.kind == "Polyp" {
		noxServer.Map.EachObjInCircle(f.origin, 8, func(obj *server.Object) bool {
			if obj.ObjectTypeC().ID() == f.projectileID && obj.Field32 >= f.start {
				f.projectile = obj
			}
			return true
		})
	} else {
		for obj, limit := f.source.Field129, 4096; obj != nil && limit > 0; obj, limit = obj.Field128, limit-1 {
			if obj.ObjectTypeC().ID() == f.projectileID && obj.Field32 >= f.start && !obj.Flags().Has(object.FlagDestroyed) {
				f.projectile = obj
				break
			}
		}
	}
	p := f.projectile
	if p == nil {
		return
	}
	if p.Shape != p.ObjectTypeC().Shape || p.Class() != p.ObjectTypeC().Class() || unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(p)) <= math.MaxUint32 {
		e2eError(fmt.Errorf("trap projectile stock shape/class/native width changed"))
		return
	}
	if f.kind == "Polyp" {
		if p.ObjOwner != nil {
			e2eError(fmt.Errorf("Polyp did not use contact -> unowned ToxicCloud"))
			return
		}
	} else {
		if p.ObjOwner != f.source || p.Direction1 != f.direction || p.VelVec == (types.Pointf{}) {
			e2eError(fmt.Errorf("trap did not retain natural owner/aim/velocity"))
			return
		}
		if f.kind == "ArrowTrap1" || f.kind == "ArrowTrap2" {
			data := (*server.ProjectileCollideData)(p.CollideData)
			if data.Damage != f.raw || data.Field4 != f.raw {
				e2eError(fmt.Errorf("ArrowTrapDamage was not set by stock update"))
				return
			}
		} else if f.projectileID == "StrongFireball" && (*server.SparkExplosionCollideData)(p.CollideData).Power != 192 {
			e2eError(fmt.Errorf("stock Skull fireball power drifted"))
			return
		}
	}
	f.wire, f.launched = uint32(noxServer.GetUnitNetCode(p)), true
	if f.kind == "chapter8-tower" {
		f.place(f.unit, f.contact)
	}
	e2eLog.Printf("TRAP NATURAL LAUNCH: kind=%s target=%s object=%p type=%s owner=%p position=%v direction=%d velocity=%v frame=%d age=%d", f.kind, f.victim, p, p.ObjectTypeC().ID(), p.ObjOwner, p.PosVec, p.Direction1, p.VelVec, noxServer.Frame(), noxServer.Frame()-p.Field32)
}

func (f *e2eTrapProjectileFixture) observe() {
	if !f.active {
		return
	}
	f.idleInput.observe(noxServer.Frame(), noxClient.Inp.GetMousePos(), e2eQueueInput)
	if !f.hit {
		f.observeLaunch()
	}
	if f.unit != f.host && (f.host.HealthData.Cur != f.hostHP || f.host.Poison540 != 0) {
		e2eError(fmt.Errorf("trap damaged the off-lane spectator"))
		return
	}
	if f.launched && noxClient.Objs.ByNetCode(uint16(f.wire)) != nil {
		f.projectileReported = true
	}
	if !f.hit && f.unit.HealthData.Cur < f.health {
		marker, kind, carry := e2eSpellUnitMatrixMarker(f.unit)
		wantMarker, wantType := uint32(1), uint32(noxServer.Types.IndByID(f.projectileID))
		if f.kind == "Polyp" {
			wantMarker, wantType = 2, uint32(object.DamagePoison)
			f.damage = int32(f.health - f.unit.HealthData.Cur)
			if f.damage < 3 || f.damage > 10 || f.unit.Poison540 != 1 {
				e2eError(fmt.Errorf("Polyp cloud did not apply stock 3..10 damage + dose one: loss=%d dose=%d source=%p expected-cloud=%p frame=%d", f.damage, f.unit.Poison540, f.unit.Obj130, f.projectile, f.unit.Frame134))
				return
			}
			f.poisonApplied = f.unit.HealthData.Field16
			var ok bool
			f.poisonFirst, ok = e2ePoisonFirstTick(f.poisonApplied, 1)
			if !ok {
				e2eError(fmt.Errorf("Polyp poison has no natural application timestamp"))
				return
			}
		}
		if !f.launched || f.unit.HealthData.Cur != f.health-uint16(f.damage) || f.unit.Obj130 != f.projectile || f.unit.Field131 != uint32(f.typ) || f.unit.Frame134 < f.start || marker != wantMarker || kind != wantType || carry != f.carry {
			e2eError(fmt.Errorf("trap hit mismatch %s/%s launched=%t HP=%d/%d source=%p/%p type=%d marker=%d/%d expected=%d/%d carry=%x/%x frame=%d/%d", f.kind, f.victim, f.launched, f.unit.HealthData.Cur, f.health-uint16(f.damage), f.unit.Obj130, f.projectile, f.unit.Field131, marker, kind, wantMarker, wantType, carry, f.carry, f.unit.Frame134, f.start))
			return
		}
		f.marker, f.markerType, f.hit = marker, kind, true
		e2eLog.Printf("TRAP ACTUAL HIT: kind=%s target=%s HP=%d->%d loss=%d marker=%d/%d frame=%d source=%p", f.kind, f.victim, f.health, f.unit.HealthData.Cur, f.damage, marker, kind, f.unit.Frame134, f.unit.Obj130)
		f.place(f.unit, f.outside)
		noxServer.DelayedDelete(f.projectile)
		if f.kind != "Polyp" && f.kind != "chapter8-tower" {
			asObjectS(f.source).Enable(false)
		}
	}
	if f.kind == "Polyp" && f.hit && !f.dot {
		want, dot, ok := e2eTrapProjectilePoisonHP(f.poisonApplied, noxServer.Frame(), noxServer.TickRate(), f.unit.HealthData.Max, f.health-uint16(f.damage), f.victim == "player")
		if !ok || f.unit.HealthData.Cur != want || f.unit.Poison540 != 1 || f.unit.HealthData.Field16 != f.poisonApplied {
			e2eError(fmt.Errorf("Polyp natural regen/DOT HP mismatch: target=%s HP=%d/%d frame=%d applied=%d first=%d", f.victim, f.unit.HealthData.Cur, want, noxServer.Frame(), f.poisonApplied, f.poisonFirst))
			return
		}
		if dot {
			marker, kind, carry := e2eSpellUnitMatrixMarker(f.unit)
			if f.unit.Frame134 != f.poisonFirst || f.unit.Obj130 != nil || f.unit.Field131 != uint32(object.DamagePoison) || marker != 2 || kind != uint32(object.DamagePoison) || carry != f.carry {
				e2eError(fmt.Errorf("Polyp first natural DOT attribution mismatch"))
				return
			}
			f.poisonHealth, f.dot = want, true
			e2eLog.Printf("POLYP NATURAL DOT: target=%s HP=%d->%d applied=%d first-DOT=%d natural-regeneration=verified", f.victim, want+1, want, f.poisonApplied, f.poisonFirst)
		}
	}
	if !f.hit && (noxServer.Frame()-f.start)%120 == 0 {
		e2eLog.Printf("TRAP WAIT: kind=%s target=%s elapsed=%d launched=%t origin=%v unit=%v HP=%d", f.kind, f.victim, noxServer.Frame()-f.start, f.launched, f.origin, f.unit.PosVec, f.unit.HealthData.Cur)
	}
}

func (f *e2eTrapProjectileFixture) clientHP(damage int32) bool {
	if f.victim == "player" {
		meter, ready := e2eClientHUDMeter(0)
		return e2eFireballHUDHit(meter, ready, f.health, damage, f.hostMax)
	}
	code := uint16(noxServer.GetUnitNetCode(f.unit))
	delta, got := legacy.HealthChangeForDrawable(uint32(code))
	return noxClient.Objs.ByNetCode(code) != nil && got && int32(delta) == -damage
}

func (f *e2eTrapProjectileFixture) complete() bool {
	if !f.hit {
		return false
	}
	if !f.instantReported && f.clientHP(f.damage) {
		f.instantReported = true
		e2eLog.Printf("TRAP CLIENT INSTANT HP: kind=%s target=%s damage=%d", f.kind, f.victim, f.damage)
	}
	dotDamage := int32(1)
	if f.victim == "player" {
		dotDamage = int32(f.health - f.poisonHealth)
	}
	if !f.instantReported || f.kind == "Polyp" && (!f.dot || !f.clientHP(dotDamage)) {
		return false
	}
	if f.kind == "Polyp" {
		if f.victim == "player" {
			meter, ready := e2eClientHUDMeter(0)
			if !ready || !meter.Poisoned {
				return false
			}
		}
	}
	path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
	if err != nil {
		e2eError(err)
		return true
	}
	e2eLog.Printf("TRAP PROJECTILE PASS: kind=%s target=%s natural-launch/collision/server-HP/client-HP/marker/carry=verified poison-DOT=%t projectile-drawable-observed=%t frame=%s", f.kind, f.victim, f.dot, f.projectileReported, path)
	return true
}

func (f *e2eTrapProjectileFixture) cleanup() {
	f.active = false
	if f.kind == "Polyp" {
		noxServer.S().RemovePoison4EE9D0(f.unit)
	}
	if f.unit != f.host {
		noxServer.DelayedDelete(f.unit)
	}
	if f.kind != "chapter8-tower" && f.kind != "Polyp" {
		noxServer.DelayedDelete(f.source)
	}
	f.place(f.host, f.original)
	asObjectS(f.host).SetMaxHealth(int(f.hostMax))
	asObjectS(f.host).SetHealth(int(f.hostHP))
}

func (sc *e2eScenario) CheckTrapProjectileDamage(kind, name string) {
	if _, _, ok := e2eTrapProjectileKind(kind); !ok {
		e2eError(fmt.Errorf("invalid stock trap case %s", kind))
		return
	}
	for _, victim := range []string{"player", "monster", "NPC"} {
		f := &e2eTrapProjectileFixture{kind: kind, victim: victim}
		label := name + " " + victim
		sc.addWhen(0, label+" prepare original source/victim", 1200, func() bool {
			u := noxServer.Players.HostUnit()
			return nox_client_isConnected() && noxClient.ClientPlayerUnit() != nil && u != nil && u.Buffs == 0 && u.Poison540 == 0
		}, f.prepare)
		sc.Wait(6, label+" safe baseline/publication")
		sc.addWhen(0, label+" real client victim publication", 1200, func() bool {
			return f.unit != nil && noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.unit))) != nil
		}, nil)
		sc.add(0, label+" ordinary trigger", f.begin)
		sc.addWhen(1, label+" real projectile and HP reports", 900, f.complete, f.cleanup)
		sc.Wait(6, label+" ordinary cleanup")
	}
}
