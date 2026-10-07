package opennox

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	nsp "github.com/opennox/noxscript/ns/v4/spell"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

// Input-only preparation is shared with the ordinary Mark fixture. This
// observer uses separate sound accounting and never supplies marker results.
type e2eMarkSlotsFixture struct {
	base          e2eMarkFixture
	audio         int
	expectedSound sound.ID
	active        bool
}

func (f *e2eMarkSlotsFixture) prepare() {
	f.base.prepare()
	f.base.active = false // ordinary Mark's observer is not this scenario
	for index := 0; index < 4; index++ {
		if noxServer.Spells.DefByInd(spell.ID(46+index)).GetCastSound() == 0 {
			e2eError(fmt.Errorf("Mark slot %d lacks its stock cast sound", index+1))
			return
		}
	}
	noxServer.Audio.OnSound(func(id sound.ID, kind int, unit *server.Object, _ types.Pointf) {
		if !f.active || id != f.expectedSound {
			return
		}
		if kind != 0 || unit != f.base.host || f.audio >= 8 {
			e2eError(fmt.Errorf("Mark slots audio lost caster identity/kind/count"))
			return
		}
		f.audio++
		e2eLog.Printf("MARK SLOTS AUDIO: level=%d event=%d caster=%p kind=%d", f.base.level, f.audio, unit, kind)
	})
	f.active = true
}

func (f *e2eMarkSlotsFixture) cast(ordinal, index int) {
	b := &f.base
	point := b.origin.Add(b.direction.Mul(float32(ordinal * 6)))
	if !e2eWarriorLaneClear(b.host, b.origin, point) {
		e2eError(fmt.Errorf("Mark slots input position is outside the verified stock lane"))
		return
	}
	// Position is input setup; slots, charges, timestamps, ownership, HP,
	// mana, sounds, packets and rendered objects remain actual game outputs.
	asObjectS(b.host).SetPos(point)
	b.host.VelVec, b.host.ForceVec, b.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	data := b.host.UpdateDataPlayer()
	before, charges, frame := data.Field29, data.Field39, noxServer.Frame()
	f.expectedSound = noxServer.Spells.DefByInd(spell.ID(46 + index)).GetCastSound()
	api := noxServer.noxScriptP()
	api.CastSpellLvl(nsp.Spell(fmt.Sprintf("SPELL_MARK_%d", index+1)), b.level, api.toObj(b.host), api.toObj(b.host))
	marker := data.Field29[index]
	if marker == nil || marker.ObjectTypeC() == nil || marker.ObjectTypeC().ID() != fmt.Sprintf("TeleportGlyph%d", index+1) ||
		marker.ObjOwner != b.host || marker.PosVec != point || marker.Field34 != frame ||
		unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(marker)) <= math.MaxUint32 {
		e2eError(fmt.Errorf("Mark named-slot allocation/placement/timestamp/owner/native identity mismatch: level=%d ordinal=%d slot=%d marker=%p", b.level, ordinal, index, marker))
		return
	}
	if ordinal < 4 {
		if before[index] != nil {
			e2eError(fmt.Errorf("Mark named-slot allocation replaced an occupied marker"))
			return
		}
	} else if marker != before[index] || marker != b.markers[index] {
		e2eError(fmt.Errorf("Mark named-slot relocation allocated instead of reusing its marker"))
		return
	}
	for i := 0; i < 4; i++ {
		if i != index && (data.Field29[i] != before[i] || before[i] != nil && (before[i].PosVec != b.points[i] || before[i].Field34 != b.stamps[i])) {
			e2eError(fmt.Errorf("Mark slots cast changed unrelated slot %d", i))
			return
		}
	}
	shift := uint(index * 8)
	wantCharges := charges&^(uint32(0xff)<<shift) | uint32(3)<<shift
	if data.Field39 != wantCharges {
		e2eError(fmt.Errorf("Mark slots cast changed unrelated packed charge bytes"))
		return
	}
	b.markers[index], b.points[index], b.stamps[index] = marker, point, frame
	b.wires[index], b.scriptIDs[index], b.charges = marker.NetCode, marker.ScriptIDVal, wantCharges
	b.clientCodes[index] = uint16(noxServer.GetUnitNetCode(marker))
	b.unchangedUnits()
	e2eLog.Printf("MARK SLOTS CAST: level=%d ordinal=%d slot=%d marker=%p type=%s owner=%p position=%v frame=%d charges=%08x reuse=%t", b.level, ordinal, index, marker, marker.ObjectTypeC().ID(), marker.ObjOwner, marker.PosVec, frame, data.Field39, ordinal >= 4)
}

func (f *e2eMarkSlotsFixture) published(casts int) bool {
	b := &f.base
	b.unchangedUnits()
	data := b.host.UpdateDataPlayer()
	if data.Field29 != b.markers || data.Field39 != b.charges {
		e2eError(fmt.Errorf("Mark named slots/charges changed between normal script casts"))
		return false
	}
	for index, marker := range b.markers {
		if marker == nil {
			continue
		}
		if !e2eFistInWorld(marker, b.wires[index], b.scriptIDs[index]) || noxClient.Objs.ByNetCode(b.clientCodes[index]) == nil {
			return false
		}
		if marker.ObjOwner != b.host || marker.PosVec != b.points[index] || marker.Field34 != b.stamps[index] {
			e2eError(fmt.Errorf("Mark slot %d owner/position/timestamp changed in the live world", index))
			return false
		}
	}
	return f.audio == casts
}

func (f *e2eMarkSlotsFixture) npcNoOp() {
	api := noxServer.noxScriptP()
	for index := 0; index < 4; index++ {
		api.CastSpellLvl(nsp.Spell(fmt.Sprintf("SPELL_MARK_%d", index+1)), f.base.level, api.toObj(f.base.npc), api.toObj(f.base.npc))
	}
	if !f.npcUnchanged() {
		e2eError(fmt.Errorf("Mark named-slot non-player no-op changed observed outputs"))
		return
	}
	e2eLog.Printf("MARK SLOTS NPC NO-OP: level=%d npc=%p casts=4 slots/charges/audio=unchanged", f.base.level, f.base.npc)
}

func (f *e2eMarkSlotsFixture) npcUnchanged() bool {
	if !f.published(8) {
		return false
	}
	for _, unit := range noxServer.Objs.AllObjects() {
		if unit.ObjOwner != f.base.npc || unit.ObjectTypeC() == nil {
			continue
		}
		for index := 0; index < 4; index++ {
			if unit.ObjectTypeC().ID() == fmt.Sprintf("TeleportGlyph%d", index+1) {
				e2eError(fmt.Errorf("Mark named-slot no-op allocated an NPC marker"))
				return false
			}
		}
	}
	return true
}

func (f *e2eMarkSlotsFixture) cleanup() {
	b := &f.base
	b.unchangedUnits()
	f.active = false
	// Registered MarkerDie releases slots; cleanup is not a claimed natural
	// expiry or combat hit, and does not clear the slots in the fixture.
	for _, marker := range b.markers {
		server.CallObjectDeath(marker.Death, marker)
	}
	if b.host.UpdateDataPlayer().Field29 != [4]*server.Object{} || b.host.UpdateDataPlayer().Field39 != b.charges {
		e2eError(fmt.Errorf("Mark slots registered death failed to release the actual marker slots"))
		return
	}
	noxServer.DelayedDelete(b.npc)
	asObjectS(b.host).SetPos(b.original)
	b.host.VelVec, b.host.ForceVec, b.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	e2eLog.Printf("MARK SLOTS COMPLETE: level=%d allocations=4 named-reuse=4 audio=%d registered-death-cleared=4 HP=%d mana=%d NPC-HP=%d", b.level, f.audio, b.health, b.mana, b.npcHP)
}

func (sc *e2eScenario) CheckMarkSlotsSpell(level int, name string) {
	if !e2eMarkLevel(level) {
		e2eError(fmt.Errorf("invalid Mark slots level %d", level))
		return
	}
	f := &e2eMarkSlotsFixture{base: e2eMarkFixture{level: level}}
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		host := noxServer.Players.HostUnit()
		return host != nil && host.Buffs == 0 && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, f.prepare)
	sc.Wait(12, name+" publish stock NPC")
	// Allocation skips empty slots; relocation deliberately does not select
	// the oldest marker. All four named slots execute both real branches.
	order := [8]int{2, 0, 3, 1, 3, 1, 0, 2}
	for ordinal, index := range order {
		ordinal, index := ordinal, index
		sc.add(12, fmt.Sprintf("%s cast %d", name, ordinal), func() { f.cast(ordinal, index) })
		sc.addWhen(1, fmt.Sprintf("%s marker %d published", name, ordinal), 120, func() bool { return f.published(ordinal + 1) }, func() {
			e2eLog.Printf("MARK SLOTS PUBLISHED: level=%d ordinal=%d slot=%d audio=%d world/drawables=live", level, ordinal, index, f.audio)
		})
		if ordinal == 3 || ordinal == 7 {
			sc.CaptureMagicFrame(fmt.Sprintf("%s actual named markers ordinal %d", name, ordinal))
		}
	}
	sc.add(1, name+" NPC no-op", f.npcNoOp)
	sc.addWhen(12, name+" NPC no-op observation", 120, f.npcUnchanged, f.base.unchangedUnits)
	sc.add(1, name+" cleanup", f.cleanup)
	sc.addWhen(1, name+" deleted units", 120, func() bool {
		if e2eObjectInWorld(f.base.npc) {
			return false
		}
		for index, marker := range f.base.markers {
			if e2eFistInWorld(marker, f.base.wires[index], f.base.scriptIDs[index]) || noxClient.Objs.ByNetCode(f.base.clientCodes[index]) != nil {
				return false
			}
		}
		return true
	}, func() {
		e2eLog.Printf("MARK SLOTS CLEANUP: level=%d stock NPC and markers world/drawables=removed", level)
	})
}
