package server

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

// LockCastState52CE90 replaces the original shared selected-door and group-
// owner DWORDs with native pointers. The state remains shared across nested
// casts, rather than silently changing the original globals to local caches.
type LockCastState52CE90 struct {
	Selected   *Object
	Nearest    float32
	GroupOwner *Object
}

type lockCastHooks52CE90 struct {
	state     *LockCastState52CE90
	rect      func(types.Rectf, func(*Object))
	trace     func(types.Pointf, types.Pointf, MapTraceFlags) bool
	fps       func() uint32
	frame     func() uint32
	message   func(*Object, string, uint8)
	castSound func(int32) sound.ID
	audio     func(sound.ID, *Object, int, uint32)
}

func lockCastSpill52CE90(value float64) float32 {
	return float32(monsterMoveToRunSpill544440(value))
}

// 0052CF90 spills the precision-53/chop Y-square plus X-square before
// both comparisons. Unlike TriggerGlyph, the comparison is NOT against
// the unspilled register. Unordered sets C0|C3 and passes both gates.
func lockCastDistance52CF90(aim, door types.Pointf) float32 {
	dx := monsterMoveToRunAddChop53_544434(float64(aim.X), -float64(door.X))
	dy := monsterMoveToRunAddChop53_544434(float64(aim.Y), -float64(door.Y))
	ySquare := monsterMoveToRunSquareChop53_544434(dy)
	xSquare := monsterMoveToRunSquareChop53_544434(dx)
	return lockCastSpill52CE90(monsterMoveToRunAddChop53_544434(ySquare, xSquare))
}

func lockCastCandidate52CF90(door, aim *Object, h lockCastHooks52CE90) {
	if uint8(door.ObjClass)&0x80 == 0 {
		return
	}
	distance := lockCastDistance52CF90(aim.PosVec, door.PosVec)
	if distance > 22500 || distance >= h.state.Nearest {
		return
	}
	fromX := aim.PosVec.X
	update := door.UpdateDataDoor()
	fromY := aim.PosVec.Y
	x := float64(DoorDirectionX(update.CurrentDirection)) * 0.5
	toX := lockCastSpill52CE90(monsterMoveToRunAddChop53_544434(x, float64(door.PosVec.X)))
	y := float64(DoorDirectionY(update.CurrentDirection)) * 0.5
	toY := lockCastSpill52CE90(monsterMoveToRunAddChop53_544434(y, float64(door.PosVec.Y)))
	if h.trace(types.Ptf(fromX, fromY), types.Ptf(toX, toY), 0) {
		// Reload the shared destination after trace, but retain this
		// candidate and the pre-trace spilled distance.
		h.state.Selected, h.state.Nearest = door, distance
	}
}

// 0052CE60 applies the shared owner to every Door in the group rectangle,
// regardless of its previous owner, direction, lock code, or object flags.
func lockCastGroupDoor52CE60(door *Object, h lockCastHooks52CE90) {
	if uint8(door.ObjClass)&0x80 == 0 {
		return
	}
	door.ObjOwner = h.state.GroupOwner
	fps := h.fps()
	frame := h.frame()
	door.Field34 = frame + 60*fps
}

// 0052D060 caches the selected DoorUpdate pointer, uses wrapping signed
// DWORD tile*23 for ALL four bounds (including max Y), and clears the shared
// group owner only after the ordinary iterator return, not on a fault.
func lockCastGroup52D060(door, caster *Object, h lockCastHooks52CE90) {
	update := door.UpdateDataDoor()
	minX := lockCastSpill52CE90(float64(int32(uint32(update.TileX)*23)) - 34)
	minY := lockCastSpill52CE90(float64(int32(uint32(update.TileY)*23)) - 34)
	maxX := lockCastSpill52CE90(float64(int32(uint32(update.TileX)*23)) + 34)
	tileY := update.TileY
	h.state.GroupOwner = caster
	maxY := lockCastSpill52CE90(float64(int32(uint32(tileY)*23)) + 34)
	h.rect(types.Rectf{Min: types.Ptf(minX, minY), Max: types.Ptf(maxX, maxY)}, func(obj *Object) {
		lockCastGroupDoor52CE60(obj, h)
	})
	h.state.GroupOwner = nil
}

// lockCast52CE90 restores GAME.EXE 0052CE90..0052CF85. The third object
// centers the spatial query and owns the lock; the fourth object is used by
// the candidate callback for its aim/visibility. There are no new nil, class,
// dead, level, existing-expiry, key, or protection gates.
func lockCast52CE90(id int32, caster, aim *Object, h lockCastHooks52CE90) int32 {
	rect := types.Rectf{
		Min: types.Ptf(lockCastSpill52CE90(monsterMoveToRunAddChop53_544434(float64(caster.PosVec.X), -150)), lockCastSpill52CE90(monsterMoveToRunAddChop53_544434(float64(caster.PosVec.Y), -150))),
		Max: types.Ptf(lockCastSpill52CE90(monsterMoveToRunAddChop53_544434(float64(caster.PosVec.X), 150)), lockCastSpill52CE90(monsterMoveToRunAddChop53_544434(float64(caster.PosVec.Y), 150))),
	}
	h.state.Selected, h.state.Nearest = nil, 100000000
	h.rect(rect, func(obj *Object) { lockCastCandidate52CF90(obj, aim, h) })
	selected := h.state.Selected
	if selected == nil {
		return 0
	}
	if owner := selected.ObjOwner; owner != nil && owner != caster {
		h.message(caster, "ExecSpel.c:DoorAlreadyLocked", 0)
		return 0
	}
	selected.ObjOwner = caster
	fps := h.fps()
	frame := h.frame()
	h.state.Selected.Field34 = frame + 60*fps
	lockCastGroup52D060(h.state.Selected, caster, h)
	selected = h.state.Selected
	idSound := h.castSound(id)
	h.audio(idSound, selected, 0, 0)
	return 1
}

type LockCastRuntime52CE90 struct {
	State           *LockCastState52CE90
	PriorityMessage func(*Object, string, uint8)
}

func (s *Server) CastLock52CE90(id int32, _ *Object, caster, aim *Object, _ *SpellAcceptArg, _ int32, runtime LockCastRuntime52CE90) int32 {
	message := runtime.PriorityMessage
	if message == nil {
		message = func(unit *Object, id string, value uint8) { s.NetPriMsgToPlayer(unit, strman.ID(id), value) }
	}
	return lockCast52CE90(id, caster, aim, lockCastHooks52CE90{
		state: runtime.State,
		rect: func(rect types.Rectf, callback func(*Object)) {
			s.Map.EachObjInRect(rect, func(obj *Object) bool { callback(obj); return true })
		},
		trace:     s.MapTraceRay,
		fps:       s.TickRate,
		frame:     s.Frame,
		message:   message,
		castSound: func(id int32) sound.ID { return s.Spells.DefByInd(spell.ID(id)).GetCastSound() },
		audio:     func(id sound.ID, unit *Object, kind int, code uint32) { s.Audio.EventObj(id, unit, kind, code) },
	})
}
