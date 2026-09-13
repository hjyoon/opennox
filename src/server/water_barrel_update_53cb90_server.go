package server

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

// WaterBarrelUpdateRuntime53CB90 supplies the map, deletion, and audio services
// used by the original WaterBarrelUpdate callback.
type WaterBarrelUpdateRuntime53CB90 struct {
	DelayedDelete func(*Object)
}

type waterBarrelUpdateNativeDeps53CB90 struct {
	frame         func() uint32
	eachInRect    func(types.Rectf, func(*Object) bool)
	delayedDelete func(*Object)
	audio         func(sound.ID, *Object)
}

// GAME.EXE passes literal audio index 283, regardless of the current sound
// table's name for that index.
const waterBarrelAudioID53CB90 = sound.ID(283)

func waterBarrelUpdateNative53CB90(source *Object, deps waterBarrelUpdateNativeDeps53CB90) {
	// The original subtracts two unsigned 32-bit frames before branching.
	age := deps.frame() - source.Field32
	if age == 8 {
		pos := source.PosVec
		rect := types.Rectf{
			Min: types.Ptf(pos.X-40, pos.Y-40),
			Max: types.Ptf(pos.X+40, pos.Y+40),
		}
		deletedFire := false
		deps.eachInRect(rect, func(candidate *Object) bool {
			if candidate.ObjClass&object.ClassFire == 0 {
				return true
			}
			// The C callback receives a pointer to the barrel's live position,
			// not a copy of the rectangle center.
			dx := source.PosVec.X - candidate.PosVec.X
			dy := source.PosVec.Y - candidate.PosVec.Y
			distance := math.Sqrt(float64(dx)*float64(dx) + float64(dy)*float64(dy))
			if distance-float64(candidate.Shape.Circle.R) <= 40 {
				deps.delayedDelete(candidate)
				deletedFire = true
			}
			return true
		})
		if deletedFire {
			deps.audio(waterBarrelAudioID53CB90, source)
		}
	} else if age >= 30 {
		deps.delayedDelete(source)
	}
}

// WaterBarrelUpdate53CB90 restores GAME.EXE 0053CB90 and its fire-extinguish
// callback at 0053CC30 without passing native object pointers through int.
func (s *Server) WaterBarrelUpdate53CB90(source *Object, runtime WaterBarrelUpdateRuntime53CB90) {
	waterBarrelUpdateNative53CB90(source, waterBarrelUpdateNativeDeps53CB90{
		frame:         s.Frame,
		eachInRect:    s.Map.EachObjInRect,
		delayedDelete: runtime.DelayedDelete,
		audio: func(id sound.ID, obj *Object) {
			s.Audio.EventObj(id, obj, 0, 0)
		},
	})
}
