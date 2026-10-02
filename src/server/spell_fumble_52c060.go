package server

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
)

const (
	fumbleCastCacheBase52C060   = uintptr(0x5D4594)
	fumbleCastCacheOffset52C060 = uintptr(2487728)
)

type fumbleCastDeps52C060 struct {
	forceDrop      func(*Object, *Object) int32
	dropAll        func(*Object)
	loadTypeCache  func() uint32
	lookupType     func(string) uint32
	storeTypeCache func(uint32)
	applyForce     func(*Object, types.Pointf, float64)
	clearOwner     func(*Object)
	ballAudio      func(*Object)
	effectAudio    func(int32, *Object)
}

// fumbleCast52C060 restores GAME.EXE 0052C060..0052C183. Player/NPC
// equipment and ordinary-monster inventory take distinct paths. Save each
// inventory successor before force-drop, but reload arg.Obj at each original
// call site. The ball cache remains a whole DWORD compared to a zero-extended
// TypeInd; neither the unused spell level nor Destroyed flags gate the cast.
func fumbleCast52C060(id int32, caster *Object, arg *SpellAcceptArg, h fumbleCastDeps52C060) int32 {
	target := arg.Obj
	if target == nil {
		return 0
	}
	class := uint32(target.ObjClass)
	if class&4 != 0 || class&2 != 0 && uint32(target.ObjSubClass)&0x10 != 0 {
		for item := target.InvFirstItem; item != nil; {
			flags := uint32(item.ObjFlags)
			next := item.InvNextItem
			if flags&0x100 != 0 {
				itemClass := uint32(item.ObjClass)
				if itemClass&0x1001000 != 0 || itemClass&0x2000000 != 0 && uint32(item.ObjSubClass)&2 != 0 {
					h.forceDrop(arg.Obj, item)
				}
			}
			item = next
		}
		kind := h.loadTypeCache()
		if kind == 0 {
			kind = h.lookupType("GameBall")
			h.storeTypeCache(kind)
		}
		ownedTarget := arg.Obj
		for ball := ownedTarget.Field129; ball != nil; ball = ball.Field128 {
			if uint32(ball.TypeInd) != kind {
				continue
			}
			h.applyForce(ball, ownedTarget.PosVec, 100)
			h.clearOwner(ball)
			h.ballAudio(arg.Obj)
			break
		}
	} else if class&2 == 0 || uint32(target.ObjSubClass)&0x2000 == 0 {
		h.dropAll(target)
		h.applyForce(arg.Obj, caster.PosVec, 50)
	}
	h.effectAudio(id, arg.Obj)
	return 1
}

type FumbleCastRuntime52C060 struct {
	ForceDrop  func(*Object, *Object) int32
	DropAll    func(*Object)
	ApplyForce func(*Object, types.Pointf, float64)
}

// CastFumble52C060 retains the spell-selector signature and uses the existing
// native drop, physics and owner services. Unlike Wink, Fumble does not clear
// ball flags/Obj130 or send a ball-status message. Selector one emits the
// effect sound on the live target, not a cast sound on the audio owner.
func (s *Server) CastFumble52C060(id int32, _ *Object, _ *Object, caster *Object, arg *SpellAcceptArg, _ int32, runtime FumbleCastRuntime52C060) int32 {
	return fumbleCast52C060(id, caster, arg, fumbleCastDeps52C060{
		forceDrop: runtime.ForceDrop,
		dropAll:   runtime.DropAll,
		loadTypeCache: func() uint32 {
			return memmap.Uint32(fumbleCastCacheBase52C060, fumbleCastCacheOffset52C060)
		},
		lookupType: func(name string) uint32 { return uint32(s.Types.IndByID(name)) },
		storeTypeCache: func(value uint32) {
			*memmap.PtrUint32(fumbleCastCacheBase52C060, fumbleCastCacheOffset52C060) = value
		},
		applyForce: runtime.ApplyForce,
		clearOwner: s.ObjClearOwner,
		ballAudio:  func(target *Object) { s.Audio.EventObj(sound.ID(926), target, 0, 0) },
		effectAudio: func(id int32, target *Object) {
			s.Audio.EventObj(s.Spells.DefByInd(spell.ID(id)).GetOnSound(), target, 0, 0)
		},
	})
}
