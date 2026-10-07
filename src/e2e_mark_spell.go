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
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func e2eMarkLevel(level int) bool { return level >= 1 && level <= 5 }

type e2eMarkFixture struct {
	level               int
	host, npc           *server.Object
	markers             [4]*server.Object
	wires               [4]uint32
	clientCodes         [4]uint16
	scriptIDs           [4]int32
	points              [4]types.Pointf
	stamps              [4]uint32
	charges             uint32
	original, direction types.Pointf
	origin              types.Pointf
	health, mana, npcHP uint16
	audio               int
	castSound           sound.ID
	active              bool
}

func (f *e2eMarkFixture) prepare() {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.HealthData == nil || f.host.ControllingPlayer() == nil || f.host.HealthData.Cur == 0 ||
		f.host.Buffs != 0 || f.host.Flags().HasAny(object.FlagDead|object.FlagDestroyed|object.FlagNoUpdate) ||
		f.host.UpdateDataPlayer().Field29 != [4]*server.Object{} {
		e2eError(fmt.Errorf("Mark requires a live unenchanted host with four naturally empty marker slots"))
		return
	}
	death, _, ok := server.ObjectDeathHandler("MarkerDie")
	for index := 0; index < 4; index++ {
		name := fmt.Sprintf("TeleportGlyph%d", index+1)
		typ := noxServer.Types.ByID(name)
		if !ok || typ == nil || typ.Death != death {
			e2eError(fmt.Errorf("Mark requires stock %s with the registered MarkerDie callback", name))
			return
		}
	}
	f.original = f.host.PosVec
	var err error
	f.origin, f.direction, err = e2eWarriorAbilityArena(f.original, f.host.Shape.Circle.R+16, func(from, to types.Pointf) bool { return e2eWarriorLaneClear(f.host, from, to) })
	if err != nil {
		e2eError(err)
		return
	}
	f.npc = noxServer.NewObjectByTypeID("NPC")
	if f.npc == nil || f.npc.UpdateData == nil || f.npc.HealthData == nil {
		e2eError(fmt.Errorf("Mark requires a stock NPC for the original non-player no-op"))
		return
	}
	noxServer.CreateObjectAt(f.npc, nil, e2eMarkNPCPosition(f.origin, f.direction, f.host.Shape.Circle.R, f.npc.Shape.Circle.R))
	f.npc.UpdateDataMonster().SetAggression(0)
	f.npc.ClearActionStack()
	f.npc.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
	noxServer.ObjectsAddPending()
	if !e2eObjectInWorld(f.npc) {
		e2eError(fmt.Errorf("Mark NPC fixture is not in the live world"))
		return
	}
	laneEnd := f.origin.Add(f.direction.Mul(48))
	laneClear := e2eWarriorLaneClear(f.host, f.origin, laneEnd)
	oldPoint := f.origin.Sub(f.direction.Mul(24))
	bound := f.host.Shape.Circle.R + f.npc.Shape.Circle.R + 4
	old24Intersects := !e2eWarriorLaneMissesCircle(f.origin, laneEnd, oldPoint, bound)
	if !laneClear {
		e2eError(fmt.Errorf("Mark stock NPC placement still obstructs the verified player input lane: host-radius=%g NPC-radius=%g NPC-position=%v", f.host.Shape.Circle.R, f.npc.Shape.Circle.R, f.npc.PosVec))
		return
	}
	e2eLog.Printf("MARK INPUT LAYOUT: level=%d host-radius=%g npc-radius=%g old24-intersects=%t actual-lane-clear=%t", f.level, f.host.Shape.Circle.R, f.npc.Shape.Circle.R, old24Intersects, laneClear)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(f.host), f.host.UpdateData, unsafe.Pointer(f.npc), f.npc.UpdateData} {
			if uintptr(ptr) <= math.MaxUint32 {
				e2eError(fmt.Errorf("Mark live fixture pointer is not native/high: %p", ptr))
				return
			}
		}
	}
	f.castSound = noxServer.Spells.DefByInd(spell.SPELL_MARK).GetCastSound()
	if f.castSound == 0 {
		e2eError(fmt.Errorf("Mark stock cast sound is missing"))
		return
	}
	noxServer.Audio.OnSound(func(id sound.ID, kind int, unit *server.Object, _ types.Pointf) {
		if !f.active || id != f.castSound {
			return
		}
		if kind != 0 || unit != f.host || f.audio >= 5 {
			e2eError(fmt.Errorf("Mark audio lost caster identity/kind/count"))
			return
		}
		f.audio++
		e2eLog.Printf("MARK AUDIO: level=%d event=%d caster=%p kind=%d", f.level, f.audio, unit, kind)
	})
	f.health, f.mana, f.npcHP = f.host.HealthData.Cur, f.host.UpdateDataPlayer().ManaCur, f.npc.HealthData.Cur
	f.active = true
	e2eLog.Printf("MARK STOCK: level=%d host=%p player-data=%p npc=%p frame=%d", f.level, f.host, f.host.UpdateData, f.npc, noxServer.Frame())
}

func (f *e2eMarkFixture) unchangedUnits() {
	if f.host.HealthData.Cur != f.health || f.host.UpdateDataPlayer().ManaCur != f.mana || f.npc.HealthData.Cur != f.npcHP {
		e2eError(fmt.Errorf("Mark changed live HP/mana: host=%d want=%d mana=%d want=%d NPC=%d want=%d", f.host.HealthData.Cur, f.health, f.host.UpdateDataPlayer().ManaCur, f.mana, f.npc.HealthData.Cur, f.npcHP))
	}
}

func (f *e2eMarkFixture) cast(ordinal int) {
	index := ordinal
	if ordinal == 4 {
		index = 0
		for i := 1; i < 4; i++ {
			if f.stamps[i] <= f.stamps[i-1] {
				e2eError(fmt.Errorf("Mark creation timestamps did not naturally increase"))
				return
			}
		}
	}
	point := f.origin.Add(f.direction.Mul(float32(ordinal * 12)))
	if !e2eWarriorLaneClear(f.host, f.origin, point) {
		e2eError(fmt.Errorf("Mark input position is outside the clear stock lane"))
		return
	}
	// Position and WAIT are input setup only. No marker, charge, timestamp,
	// owner, health, mana, packet, sound, or pixel output is supplied.
	asObjectS(f.host).SetPos(point)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	data := f.host.UpdateDataPlayer()
	before, charges, frame := data.Field29, data.Field39, noxServer.Frame()
	api := noxServer.noxScriptP()
	api.CastSpellLvl(nsp.Spell("SPELL_MARK"), f.level, api.toObj(f.host), api.toObj(f.host))
	marker := data.Field29[index]
	if marker == nil || marker.ObjectTypeC() == nil || marker.ObjectTypeC().ID() != fmt.Sprintf("TeleportGlyph%d", index+1) ||
		marker.ObjOwner != f.host || marker.PosVec != point || marker.Field34 != frame ||
		unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(marker)) <= math.MaxUint32 {
		e2eError(fmt.Errorf("Mark real allocation/placement/timestamp/owner/native-width mismatch: level=%d ordinal=%d marker=%p frame=%d point=%v", f.level, ordinal, marker, frame, point))
		return
	}
	if ordinal < 4 {
		if before[index] != nil {
			e2eError(fmt.Errorf("Mark first-empty allocation replaced an occupied slot"))
			return
		}
	} else if marker != f.markers[0] || marker != before[0] {
		e2eError(fmt.Errorf("Mark full slots allocated instead of reusing the oldest marker"))
		return
	}
	for i := 0; i < 4; i++ {
		if i != index && (data.Field29[i] != before[i] || before[i] != nil && (before[i].PosVec != f.points[i] || before[i].Field34 != f.stamps[i])) {
			e2eError(fmt.Errorf("Mark changed an unrelated marker slot %d", i))
			return
		}
	}
	shift := uint(index * 8)
	wantCharges := charges&^(uint32(0xff)<<shift) | uint32(3)<<shift
	if data.Field39 != wantCharges || *legacy.Get_dword_5d4594_2487712_ptr() != uint32(noxServer.Types.IndByID("Glyph")) {
		e2eError(fmt.Errorf("Mark live packed recharge/shared Glyph cache mismatch"))
		return
	}
	f.markers[index], f.points[index], f.stamps[index] = marker, point, frame
	f.wires[index], f.scriptIDs[index], f.charges = marker.NetCode, marker.ScriptIDVal, wantCharges
	f.clientCodes[index] = uint16(noxServer.GetUnitNetCode(marker))
	f.unchangedUnits()
	e2eLog.Printf("MARK CAST: level=%d ordinal=%d slot=%d marker=%p type=%s owner=%p position=%v frame=%d charges=%08x reuse=%t", f.level, ordinal, index, marker, marker.ObjectTypeC().ID(), marker.ObjOwner, marker.PosVec, frame, data.Field39, ordinal == 4)
}

func (f *e2eMarkFixture) published(ordinal int) bool {
	f.unchangedUnits()
	data := f.host.UpdateDataPlayer()
	if data.Field29 != f.markers || data.Field39 != f.charges {
		e2eError(fmt.Errorf("Mark slots/charges changed between normal casts"))
		return false
	}
	for index, marker := range f.markers {
		if marker == nil {
			continue
		}
		if !e2eFistInWorld(marker, f.wires[index], f.scriptIDs[index]) {
			return false
		}
		if marker.ObjOwner != f.host || marker.PosVec != f.points[index] || marker.Field34 != f.stamps[index] {
			e2eError(fmt.Errorf("Mark live owner/position/timestamp changed in slot %d", index))
			return false
		}
		if noxClient.Objs.ByNetCode(f.clientCodes[index]) == nil {
			return false
		}
	}
	return f.audio == ordinal+1
}

func (f *e2eMarkFixture) npcNoOp() {
	api := noxServer.noxScriptP()
	api.CastSpellLvl(nsp.Spell("SPELL_MARK"), f.level, api.toObj(f.npc), api.toObj(f.npc))
	if !f.npcUnchanged() {
		e2eError(fmt.Errorf("Mark non-player no-op was not observed intact"))
		return
	}
	e2eLog.Printf("MARK NPC NO-OP: level=%d npc=%p slots/charges/audio=unchanged", f.level, f.npc)
}

func (f *e2eMarkFixture) npcUnchanged() bool {
	if !f.published(4) {
		e2eError(fmt.Errorf("Mark non-player no-op changed the player markers/audio"))
		return false
	}
	for _, unit := range noxServer.Objs.AllObjects() {
		if unit.ObjOwner != f.npc || unit.ObjectTypeC() == nil {
			continue
		}
		for index := 0; index < 4; index++ {
			if unit.ObjectTypeC().ID() == fmt.Sprintf("TeleportGlyph%d", index+1) {
				e2eError(fmt.Errorf("Mark original non-player no-op allocated an NPC marker"))
				return false
			}
		}
	}
	return true
}

func (f *e2eMarkFixture) cleanup() {
	f.unchangedUnits()
	f.active = false
	// Invoke the stock registered death dispatch. Do not clear the player
	// slots or fabricate an expiry/combat hit; MarkerDie owns their release.
	for _, marker := range f.markers {
		server.CallObjectDeath(marker.Death, marker)
	}
	data := f.host.UpdateDataPlayer()
	if data.Field29 != [4]*server.Object{} || data.Field39 != f.charges {
		e2eError(fmt.Errorf("Mark registered MarkerDie failed to clear exactly the live slots"))
		return
	}
	e2eLog.Printf("MARK COMPLETE: level=%d audio=%d oldest-reused=true registered-death-cleared=4 HP=%d mana=%d NPC-HP=%d charges=%08x", f.level, f.audio, f.health, f.mana, f.npcHP, data.Field39)
	noxServer.DelayedDelete(f.npc)
	asObjectS(f.host).SetPos(f.original)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
}

// Normal script casts create four stock markers and reuse the oldest one at
// five levels. The NPC cast is the original no-op. Marker slots/charges,
// timestamps, owners, HP, mana, packets, sound and pixels are only observed.
// Cleanup tests registered MarkerDie, not natural expiry or physical combat.
func (sc *e2eScenario) CheckMarkSpell(level int, name string) {
	if !e2eMarkLevel(level) {
		e2eError(fmt.Errorf("invalid Mark level %d", level))
		return
	}
	f := &e2eMarkFixture{level: level}
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		host := noxServer.Players.HostUnit()
		return host != nil && host.Buffs == 0 && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, f.prepare)
	sc.Wait(12, name+" publish stock NPC")
	for ordinal := 0; ordinal < 5; ordinal++ {
		ordinal := ordinal
		sc.add(12, fmt.Sprintf("%s cast %d", name, ordinal), func() { f.cast(ordinal) })
		sc.addWhen(1, fmt.Sprintf("%s marker %d published", name, ordinal), 120, func() bool { return f.published(ordinal) }, func() {
			e2eLog.Printf("MARK PUBLISHED: level=%d ordinal=%d audio=%d world/drawables=live", level, ordinal, f.audio)
		})
		if ordinal >= 3 {
			sc.CaptureMagicFrame(fmt.Sprintf("%s actual markers ordinal %d", name, ordinal))
		}
	}
	sc.add(1, name+" NPC no-op", f.npcNoOp)
	sc.addWhen(12, name+" NPC no-op observation", 120, f.npcUnchanged, f.unchangedUnits)
	sc.add(1, name+" cleanup", f.cleanup)
	sc.addWhen(1, name+" deleted units", 120, func() bool {
		if e2eObjectInWorld(f.npc) {
			return false
		}
		for index, marker := range f.markers {
			if e2eFistInWorld(marker, f.wires[index], f.scriptIDs[index]) || noxClient.Objs.ByNetCode(f.clientCodes[index]) != nil {
				return false
			}
		}
		return true
	}, func() { e2eLog.Printf("MARK CLEANUP: level=%d stock NPC and markers world/drawables=removed", level) })
}
