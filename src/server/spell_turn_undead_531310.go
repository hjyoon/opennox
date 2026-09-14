package server

import (
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/types"
)

// SpellTurnUndeadRuntime531310 supplies the game-owned effects used by the
// Turn Undead duration callbacks. Object and spell identities remain native
// pointers; only the kill-point budget has the original signed dword width.
type SpellTurnUndeadRuntime531310 struct {
	KillPoints    func(uint32) float32
	NewObject     func(string) *Object
	CreateAt      func(*Object, *Object, types.Pointf)
	SendPointFX   func(netmsg.Op, types.Pointf)
	FirstObject   func() *Object
	TypeInd       func(string) int
	DelayedDelete func(*Object)
}

// SpellTurnUndeadCreate531310 follows GAME.EXE 00531310 without decoding a
// native-width DurSpell or Object through the original PE32 dword layout.
//
//go:noinline
func SpellTurnUndeadCreate531310(record *DurSpell, runtime SpellTurnUndeadRuntime531310) int32 {
	record.Field72 = spellDurationRoundNearestEven(runtime.KillPoints(record.Level - 1))
	source := record.Caster16
	if record.Flag20 != 0 {
		source = record.Obj24
	}
	if source == nil {
		// The PE32 callback dereferences a missing source. Cancel the duration
		// before creating orphaned projectiles instead.
		return 1
	}
	position := source.PosVec
	for direction := 0; direction < 256; direction += 6 {
		killer := runtime.NewObject("UndeadKiller")
		if killer == nil {
			continue
		}
		(*UndeadKillerCollideData)(killer.CollideData).Spell = record
		runtime.CreateAt(killer, record.Caster16, position)
		killer.Direction2 = Dir16(direction)
		killer.Direction1 = Dir16(direction)
		cos, sin := SinCosDir(byte(direction))
		killer.VelVec.X = cos * 4
		killer.Float28 = 0
		killer.VelVec.Y = sin * 4
	}
	runtime.SendPointFX(netmsg.MSG_FX_TURN_UNDEAD, position)
	return 0
}

// SpellTurnUndeadUpdate531410 is the original no-op duration callback.
func SpellTurnUndeadUpdate531410(*DurSpell) int32 { return 0 }

// SpellTurnUndeadDestroy531420 schedules only this duration's active
// UndeadKiller objects for deletion. Pending objects retain their normal
// 70-frame lifetime, matching the original active-object traversal.
//
//go:noinline
func SpellTurnUndeadDestroy531420(record *DurSpell, runtime SpellTurnUndeadRuntime531310) {
	typeInd := uint16(runtime.TypeInd("UndeadKiller"))
	for obj := runtime.FirstObject(); obj != nil; {
		next := obj.ObjNext
		if obj.TypeInd == typeInd && obj.CollideData != nil &&
			(*UndeadKillerCollideData)(obj.CollideData).Spell == record {
			runtime.DelayedDelete(obj)
		}
		obj = next
	}
}
