package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
)

type cleansingFlameCastDeps52D5C0 struct {
	loadCache       func(int32) uint32
	storeCache      func(int32, uint32)
	lookupType      func(string) uint32
	priorityMsg     func(*Object, string, byte)
	hasGameFlag     func(uint32) bool
	random          func(int32, int32) int32
	newObject       func(uint32) *Object
	directionVector func(int16) types.Pointf
	traceRay        func(types.Pointf, types.Pointf, MapTraceFlags) bool
	createAt        func(*Object, *Object, types.Pointf)
	delayedDelete   func(*Object)
	fps             func() uint32
	frame           func() uint32
	updateCallback  unsafe.Pointer
	addUpdatable    func(*Object)
	predictLinear   func(*Object)
	castSound       func(int32) sound.ID
	audio           func(sound.ID, *Object, int, uint32)
}

func cleansingFlameCacheOffset52D5C0(slot int32) uintptr {
	return uintptr(uint32(2487760) + 4*uint32(slot))
}

// cleansingFlameCast52D5C0 restores GAME.EXE 0052D5C0..0052D89B. All
// owned/caster/flame pointers are native-width; cache indices, RNG bounds
// and timer sums retain DWORD arithmetic. The third object is only the
// priority-message recipient, the fourth owns each flame, and the argument
// is unused. Do not add caster/argument guards or clamp the original level.
func cleansingFlameCast52D5C0(id int32, second, recipient, caster *Object, level int32, h cleansingFlameCastDeps52D5C0) int32 {
	if h.loadCache(0) == 0 {
		for slot, name := range [10]string{
			"SmallFlameCleanse", "SmallFlameCleanse", "MediumFlameCleanse", "FlameCleanse", "LargeFlameCleanse",
			"SmallBlueFlameCleanse", "SmallBlueFlameCleanse", "MediumBlueFlameCleanse", "BlueFlameCleanse", "LargeBlueFlameCleanse",
		} {
			h.storeCache(int32(slot), h.lookupType(name))
		}
	}
	if second != nil {
		for owned := second.Field129; owned != nil; owned = owned.Field128 {
			kind := uint32(owned.TypeInd)
			for slot := int32(0); slot < 10; slot++ {
				if kind == h.loadCache(slot) {
					h.priorityMsg(recipient, "plyrspel.c:TooManySpells", 0)
					return 0
				}
			}
		}
	}
	base := int32(5)
	if id == 10 {
		base = 0
	}
	if !h.hasGameFlag(2048) {
		level = 4
	}
	for remaining := 48; remaining != 0; remaining-- {
		power := level - h.random(0, 1)
		if power < 1 {
			continue
		}
		flame := h.newObject(h.loadCache(base + power - 1))
		if flame == nil {
			continue
		}
		direction := int16(h.random(0, 255))
		radius := float64(flame.Shape.Circle.R)
		flame.Direction1 = Dir16(direction)
		fromX := caster.PosVec.X
		distance := monsterMoveToRunAddChop53_544434(radius, float64(caster.Shape.Circle.R))
		distance = monsterMoveToRunAddChop53_544434(distance, 4)
		vector := h.directionVector(direction)
		// Keep the original separate FMUL, FADD and binary32 ToZero spill;
		// ARM64 must not contract this sequence into a fused multiply-add.
		x := monsterMoveForceMulChop53_50D4FE(distance, float64(vector.X))
		fromY := caster.PosVec.Y
		x = monsterMoveToRunAddChop53_544434(x, float64(caster.PosVec.X))
		y := monsterMoveForceMulChop53_50D4FE(distance, float64(vector.Y))
		y = monsterMoveToRunAddChop53_544434(y, float64(caster.PosVec.Y))
		position := types.Ptf(float32(monsterMoveToRunSpill544440(x)), float32(monsterMoveToRunSpill544440(y)))
		if !h.traceRay(types.Ptf(fromX, fromY), position, MapTraceFlags(65)) {
			h.delayedDelete(flame)
			continue
		}
		h.createAt(flame, caster, position)
		flame.Direction2 = flame.Direction1
		flame.VelVec.X = float32(monsterMoveToRunSpill544440(float64(vector.X) * 4))
		flame.VelVec.Y = float32(monsterMoveToRunSpill544440(float64(vector.Y) * 4))
		fps := h.fps()
		duration := h.random(int32(3*fps), int32(6*fps))
		flame.Field34 = h.frame() + uint32(duration)
		flame.Pos39 = caster.PosVec
		flame.Update = h.updateCallback
		h.addUpdatable(flame)
		class := flame.ObjClass | object.ClassClientPredict
		flame.Float28 = 0
		flame.ObjClass = class
		h.predictLinear(flame)
	}
	castSound := h.castSound(id)
	h.audio(castSound, caster, 0, 0)
	return 1
}

type CleansingFlameCastRuntime52D5C0 struct {
	CreateAt       func(*Object, *Object, types.Pointf)
	DelayedDelete  func(*Object)
	UpdateCallback unsafe.Pointer
}

func (s *Server) CastCleansingFlame52D5C0(id int32, second, recipient, caster *Object, _ *SpellAcceptArg, level int32, runtime CleansingFlameCastRuntime52D5C0) int32 {
	return cleansingFlameCast52D5C0(id, second, recipient, caster, level, cleansingFlameCastDeps52D5C0{
		loadCache: func(slot int32) uint32 { return memmap.Uint32(0x5D4594, cleansingFlameCacheOffset52D5C0(slot)) },
		storeCache: func(slot int32, kind uint32) {
			*memmap.PtrUint32(0x5D4594, cleansingFlameCacheOffset52D5C0(slot)) = kind
		},
		lookupType:  func(name string) uint32 { return uint32(s.Types.IndByID(name)) },
		priorityMsg: func(unit *Object, id string, arg byte) { s.NetPriMsgToPlayer(unit, strman.ID(id), arg) },
		hasGameFlag: func(flag uint32) bool { return noxflags.HasGame(noxflags.GameFlag(flag)) },
		random:      func(min, max int32) int32 { return int32(s.Rand.Logic.IntClamp(int(min), int(max))) },
		newObject:   func(kind uint32) *Object { return s.NewObjectByTypeInd(int(int32(kind))) },
		directionVector: func(direction int16) types.Pointf {
			offset := uint32(194136) + 8*uint32(int32(direction))
			return types.Ptf(memmap.Float32(0x587000, uintptr(offset)), memmap.Float32(0x587000, uintptr(offset+4)))
		},
		traceRay:       s.MapTraceRay,
		createAt:       runtime.CreateAt,
		delayedDelete:  runtime.DelayedDelete,
		fps:            s.TickRate,
		frame:          s.Frame,
		updateCallback: runtime.UpdateCallback,
		addUpdatable:   s.Objs.AddToUpdatable,
		predictLinear:  func(obj *Object) { s.NetClientPredictLinear523530(obj) },
		castSound:      func(id int32) sound.ID { return s.Spells.DefByInd(spell.ID(id)).GetCastSound() },
		audio:          func(id sound.ID, unit *Object, kind int, code uint32) { s.Audio.EventObj(id, unit, kind, code) },
	})
}
