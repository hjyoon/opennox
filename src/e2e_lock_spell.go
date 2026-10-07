package opennox

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	nsp "github.com/opennox/noxscript/ns/v4/spell"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

type e2eLockFixture struct {
	level                 int
	host, npc             *server.Object
	doors                 [3]*server.Object
	original              types.Pointf
	outsideOwner          *server.Object
	outsideExpiry, expiry uint32
	health, mana, npcHP   uint16
	castSound             sound.ID
	audio                 int
	active                bool
}

func e2eLockLevel(level int) bool { return level >= 1 && level <= 5 }

func (f *e2eLockFixture) prepare() {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.HealthData == nil || f.host.ControllingPlayer() == nil || f.host.HealthData.Cur == 0 ||
		f.host.Buffs != 0 || f.host.Flags().HasAny(object.FlagDead|object.FlagDestroyed|object.FlagNoUpdate) {
		e2eError(fmt.Errorf("Lock requires a live unenchanted host"))
		return
	}
	f.original = f.host.PosVec
	origin, direction, err := e2eWarriorAbilityArena(f.original, f.host.Shape.Circle.R+16, func(from, to types.Pointf) bool { return e2eWarriorLaneClear(f.host, from, to) })
	if err != nil {
		e2eError(err)
		return
	}
	doorCollide, _, collideOK := server.ObjectCollideHandler("DoorCollide")
	doorUpdate, size, updateOK := server.ObjectUpdateHandler("DoorUpdate")
	var doorType *server.ObjectType
	for _, typ := range noxServer.Types.List() {
		if typ.Class().Has(object.ClassDoor) && typ.Collide == doorCollide && typ.Update == doorUpdate && typ.UpdateDataSize == size {
			doorType = typ
			break
		}
	}
	if !collideOK || !updateOK || size != unsafe.Sizeof(server.DoorUpdateData{}) || doorType == nil {
		e2eError(fmt.Errorf("Lock lacks a stock Door with native DoorUpdate/DoorCollide handlers"))
		return
	}
	asObjectS(f.host).SetPos(origin)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	f.npc = noxServer.NewObjectByTypeID("NPC")
	if f.npc == nil || f.npc.UpdateData == nil || f.npc.HealthData == nil {
		e2eError(fmt.Errorf("Lock requires a stock NPC caster"))
		return
	}
	noxServer.CreateObjectAt(f.npc, nil, origin.Sub(direction.Mul(24)))
	f.npc.UpdateDataMonster().SetAggression(0)
	f.npc.ClearActionStack()
	f.npc.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
	center := origin.Add(direction.Mul(64))
	center = types.Ptf(float32(math.Round(float64(center.X)/23)*23), float32(math.Round(float64(center.Y)/23)*23))
	for i, distance := range []float32{0, 16, 70} {
		pos := center.Add(direction.Mul(distance))
		door := noxServer.NewObjectByTypeID(doorType.ID())
		if door == nil || door.UpdateData == nil {
			e2eError(fmt.Errorf("Lock stock Door was not initialized"))
			return
		}
		// Only map-record inputs are placed here. Direction 16 has positive
		// X/Y half-offsets; the grid-aligned first Door is centered in its
		// original tile*23 +/-34 group rectangle.
		update := door.UpdateDataDoor()
		update.CurrentDirection, update.TargetDirection, update.SyncedDirection, update.FractionalDir = 16, 16, 16, 128
		update.TileX = e2eDoorTileCoordinate4F4CB0(server.DoorDirectionX(16), pos.X)
		update.TileY = e2eDoorTileCoordinate4F4CB0(server.DoorDirectionY(16), pos.Y)
		noxServer.CreateObjectAt(door, nil, pos)
		f.doors[i] = door
	}
	noxServer.ObjectsAddPending()
	for _, unit := range append([]*server.Object{f.host, f.npc}, f.doors[:]...) {
		if !e2eObjectInWorld(unit) || unsafe.Sizeof(uintptr(0)) == 8 &&
			(uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 || uintptr(unit.UpdateData) <= math.MaxUint32) {
			e2eError(fmt.Errorf("Lock fixture is not live/native: object=%p update=%p", unit, unit.UpdateData))
			return
		}
	}
	for _, door := range f.doors {
		if door.ObjOwner != nil || door.UpdateDataDoor().LockCode != 0 {
			e2eError(fmt.Errorf("Lock requires initially unowned, key-free stock Doors"))
			return
		}
	}
	f.outsideOwner, f.outsideExpiry = f.doors[2].ObjOwner, f.doors[2].Field34
	f.castSound = noxServer.Spells.DefByInd(spell.SPELL_LOCK).GetCastSound()
	if f.castSound == 0 || noxServer.TickRate() == 0 || noxServer.TickRate() > 60 {
		e2eError(fmt.Errorf("Lock stock audio/tick rate is invalid"))
		return
	}
	noxServer.Audio.OnSound(func(id sound.ID, kind int, unit *server.Object, _ types.Pointf) {
		if !f.active || id != f.castSound {
			return
		}
		if kind != 0 || unit != f.doors[0] || f.audio >= 3 {
			e2eError(fmt.Errorf("Lock cast sound lost nearest-door identity/kind/count"))
			return
		}
		f.audio++
		e2eLog.Printf("LOCK AUDIO: level=%d event=%d door=%p kind=%d", f.level, f.audio, unit, kind)
	})
	e2eLog.Printf("LOCK STOCK: level=%d type=%s host=%p npc=%p doors=%p/%p/%p updates=%p/%p/%p", f.level, doorType.ID(), f.host, f.npc, f.doors[0], f.doors[1], f.doors[2], f.doors[0].UpdateData, f.doors[1].UpdateData, f.doors[2].UpdateData)
}

func (f *e2eLockFixture) unchangedOutsideAndUnits() {
	if f.doors[2].ObjOwner != f.outsideOwner || f.doors[2].Field34 != f.outsideExpiry ||
		f.host.HealthData.Cur != f.health || f.host.UpdateDataPlayer().ManaCur != f.mana || f.npc.HealthData.Cur != f.npcHP {
		e2eError(fmt.Errorf("Lock changed outside Door ownership/expiry or live HP/mana: outside=%p owner=%p want=%p expiry=%d want=%d HP=%d want=%d mana=%d want=%d NPC-HP=%d want=%d", f.doors[2], f.doors[2].ObjOwner, f.outsideOwner, f.doors[2].Field34, f.outsideExpiry, f.host.HealthData.Cur, f.health, f.host.UpdateDataPlayer().ManaCur, f.mana, f.npc.HealthData.Cur, f.npcHP))
	}
}

func (f *e2eLockFixture) castOwned(caster *server.Object, stage string) {
	frame, fps := noxServer.Frame(), noxServer.TickRate()
	api := noxServer.noxScriptP()
	api.CastSpellLvl(nsp.Spell("SPELL_LOCK"), f.level, api.toObj(caster), api.toObj(caster))
	want := frame + 60*fps
	for _, door := range f.doors[:2] {
		if door.ObjOwner != caster || door.Field34 != want {
			e2eError(fmt.Errorf("Lock %s nearest/group owner/timer mismatch: door=%p owner=%p expiry=%d want=%p/%d", stage, door, door.ObjOwner, door.Field34, caster, want))
			return
		}
	}
	f.expiry = want
	f.unchangedOutsideAndUnits()
	e2eLog.Printf("LOCK CAST: level=%d stage=%s caster=%p selected=%p group=%p owner=%p frame=%d fps=%d expiry=%d outside-unchanged=true HP=%d mana=%d", f.level, stage, caster, f.doors[0], f.doors[1], f.doors[0].ObjOwner, frame, fps, want, f.health, f.mana)
}

func (f *e2eLockFixture) begin() {
	f.health, f.mana, f.npcHP = f.host.HealthData.Cur, f.host.UpdateDataPlayer().ManaCur, f.npc.HealthData.Cur
	f.active = true
	noxServer.TickCallback(func() { f.castOwned(f.host, "player") })
}

func (f *e2eLockFixture) denied(caster *server.Object, stage string) {
	owner := f.doors[0].ObjOwner
	api := noxServer.noxScriptP()
	api.CastSpellLvl(nsp.Spell("SPELL_LOCK"), f.level, api.toObj(caster), api.toObj(caster))
	for _, door := range f.doors[:2] {
		if door.ObjOwner != owner || door.Field34 != f.expiry {
			e2eError(fmt.Errorf("Lock %s changed a foreign-owned Door", stage))
			return
		}
	}
	f.unchangedOutsideAndUnits()
	e2eLog.Printf("LOCK DENIED: level=%d stage=%s caster=%p owner=%p frame=%d expiry=%d", f.level, stage, caster, owner, noxServer.Frame(), f.expiry)
}

func (f *e2eLockFixture) reclaim() {
	// The original cast does not clear an expired foreign owner. Only an
	// ordinary DoorCollide callback clears it. This dispatch checks that
	// registered handler; it is not a claim of physically walking into it.
	f.denied(f.npc, "expired-foreign")
	for _, door := range f.doors[:2] {
		server.CallObjectCollide(door.Collide, door, f.npc, nil)
		if door.ObjOwner != nil || door.Field34 != f.expiry {
			e2eError(fmt.Errorf("Lock normal DoorCollide failed to release naturally expired owner"))
			return
		}
	}
	e2eLog.Printf("LOCK EXPIRED: level=%d frame=%d expiry=%d owner-cleared=registered-collision", f.level, noxServer.Frame(), f.expiry)
	f.castOwned(f.npc, "npc-after-expiry")
	f.denied(f.host, "player-foreign")
}

func (f *e2eLockFixture) cleanup() {
	f.unchangedOutsideAndUnits()
	e2eLog.Printf("LOCK COMPLETE: level=%d audio=%d same-owner-refresh=true live/expired-foreign-denied=true npc-reclaimed=true HP=%d mana=%d", f.level, f.audio, f.health, f.mana)
	f.active = false
	for _, door := range f.doors {
		noxServer.DelayedDelete(door)
	}
	noxServer.DelayedDelete(f.npc)
	asObjectS(f.host).SetPos(f.original)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
}

// Five normal script API levels, live stock Doors and a stock idle NPC. No
// owner, expiry, HP, mana, packet, audio or pixel outcome is supplied. This
// tests script casts and the registered collision callback, not incantation,
// autonomous NPC decisions, physical Door contact or audible audio hardware.
func (sc *e2eScenario) CheckLockSpell(level int, name string) {
	if !e2eLockLevel(level) {
		e2eError(fmt.Errorf("invalid Lock level %d", level))
		return
	}
	f := &e2eLockFixture{level: level}
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		host := noxServer.Players.HostUnit()
		return host != nil && host.Buffs == 0 && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, f.prepare)
	sc.Wait(12, name+" publish stock units")
	sc.add(0, name+" player cast", f.begin)
	sc.addWhen(1, name+" player audio", 120, func() bool { return f.audio == 1 }, func() { f.unchangedOutsideAndUnits() })
	sc.add(20, name+" same-owner refresh", func() { f.castOwned(f.host, "refresh") })
	sc.addWhen(1, name+" refresh audio", 120, func() bool { return f.audio == 2 }, func() { f.denied(f.npc, "live-foreign") })
	sc.CaptureMagicFrame(name + " player lock frame")
	sc.addWhen(1, name+" natural expiry", 4000, func() bool {
		f.unchangedOutsideAndUnits()
		for _, door := range f.doors[:2] {
			if door.ObjOwner != f.host || door.Field34 != f.expiry {
				e2eError(fmt.Errorf("Lock owner/timer changed before registered expiry collision"))
				return false
			}
		}
		return noxServer.Frame() >= f.expiry
	}, f.reclaim)
	sc.addWhen(1, name+" NPC audio", 120, func() bool { return f.audio == 3 }, func() { f.unchangedOutsideAndUnits() })
	sc.CaptureMagicFrame(name + " NPC lock frame")
	sc.add(1, name+" cleanup", f.cleanup)
	sc.addWhen(1, name+" deleted units", 120, func() bool {
		for _, unit := range append([]*server.Object{f.npc}, f.doors[:]...) {
			if e2eObjectInWorld(unit) {
				return false
			}
		}
		return true
	}, func() { e2eLog.Printf("LOCK CLEANUP: level=%d stock fixture removed", level) })
}
