package server

import (
	"image"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
)

// CreateSpellProjectileRuntime4FDDA0 supplies the operations that remain
// owned by the outer game runtime. Object pointers remain native-width;
// spell, duration, power, and reserved arguments retain their original
// fixed-width contracts.
type CreateSpellProjectileRuntime4FDDA0 struct {
	SpellGetPower func(spell.ID, *Object) int32
	CreateAt      func(*Object, *Object, types.Pointf, int32)
	ApplyEnchant  func(*Object, EnchantID, int16, uint8)
}

type createSpellProjectileNativeDeps4FDDA0 struct {
	spellFlags       func(int32) uint32
	searchTarget     func(*types.Pointf, *Object, uint32, float32, int32, *Object) *Object
	mapTrace         func(types.Pointf, types.Pointf, *types.Pointf, *image.Point, int32) int32
	newObject        func(string) *Object
	directionX       func(int16) float32
	directionY       func(int16) float32
	indexedDirection func(int16, *types.Pointf)
	spellAudio       func(int32, int32) int32
	audio            func(int32, *Object, int32, uint32)
	runtime          CreateSpellProjectileRuntime4FDDA0
}

func createSpellProjectileNative4FDDA0(
	source, target *Object,
	spellID int32,
	deps createSpellProjectileNativeDeps4FDDA0,
) *Object {
	return createSpellProjectile4FDDA0(createSpellProjectileHooks4FDDA0[
		*Object,
		unsafe.Pointer,
		*Player,
	]{
		loadSourceArg: func() *Object { return source },
		loadTargetArg: func() *Object { return target },
		loadSpellArg:  func() int32 { return spellID },
		loadRadius: func(source *Object) float32 {
			return source.Shape.Circle.R
		},
		loadClassLow: func(source *Object) uint8 {
			return uint8(source.ObjClass)
		},
		loadUpdate: func(source *Object) unsafe.Pointer {
			return source.UpdateData
		},
		loadPlayer: func(update unsafe.Pointer) *Player {
			return (*PlayerUpdateData)(update).Player
		},
		loadCursorX: func(player *Player) int32 {
			return int32(player.CursorVec.X)
		},
		loadCursorY: func(player *Player) int32 {
			return int32(player.CursorVec.Y)
		},
		spellFlags: deps.spellFlags,
		searchTarget: func(
			aim *types.Pointf,
			source *Object,
			flags uint32,
			distance float32,
			mode int32,
			self *Object,
		) *Object {
			return deps.searchTarget(aim, source, flags, distance, mode, self)
		},
		loadDirection: func(object *Object) int16 {
			return int16(object.Direction1)
		},
		loadPosX:   func(object *Object) float32 { return object.PosVec.X },
		loadPosY:   func(object *Object) float32 { return object.PosVec.Y },
		directionX: deps.directionX,
		directionY: deps.directionY,
		loadVelX:   func(object *Object) float32 { return object.VelVec.X },
		loadVelY:   func(object *Object) float32 { return object.VelVec.Y },
		mapTrace: func(
			ray *createSpellProjectileRay4FDDA0,
			outPoint *types.Pointf,
			outGrid *image.Point,
			flags int32,
		) int32 {
			return deps.mapTrace(ray.Origin, ray.Destination, outPoint, outGrid, flags)
		},
		newObject: deps.newObject,
		loadProjectileUpdate: func(projectile *Object) unsafe.Pointer {
			return projectile.UpdateData
		},
		spellPower: func(id int32, source *Object) int32 {
			return deps.runtime.SpellGetPower(spell.ID(id), source)
		},
		storeLevel: func(update unsafe.Pointer, level uint32) {
			(*SpellProjectileUpdateData)(update).Level16 = level
		},
		createAt: func(projectile, owner *Object, position types.Pointf, reserved int32) {
			deps.runtime.CreateAt(projectile, owner, position, reserved)
		},
		storeDirection1: func(projectile *Object, direction uint16) {
			projectile.Direction1 = Dir16(direction)
		},
		storeDirection2: func(projectile *Object, direction uint16) {
			projectile.Direction2 = Dir16(direction)
		},
		storeField0: func(update unsafe.Pointer, value *Object) {
			(*SpellProjectileUpdateData)(update).Field0 = value
		},
		storeTarget: func(update unsafe.Pointer, value *Object) {
			(*SpellProjectileUpdateData)(update).Target = value
		},
		storeField8: func(update unsafe.Pointer, value *Object) {
			(*SpellProjectileUpdateData)(update).Field8 = value
		},
		storeSpell: func(update unsafe.Pointer, id uint32) {
			(*SpellProjectileUpdateData)(update).Spell12 = id
		},
		indexedDirection: deps.indexedDirection,
		loadSpeed:        func(projectile *Object) float32 { return projectile.SpeedCur },
		storeVelX:        func(projectile *Object, value float32) { projectile.VelVec.X = value },
		storeVelY:        func(projectile *Object, value float32) { projectile.VelVec.Y = value },
		hasEnchant: func(source *Object, enchant int32) int32 {
			if source.HasEnchant(EnchantID(enchant)) {
				return 1
			}
			return 0
		},
		enchantPower: func(source *Object, enchant int32) int32 {
			return int32(source.EnchantPower(EnchantID(enchant)))
		},
		enchantTimer: func(source *Object, enchant int32) int32 {
			return int32(source.EnchantDur(EnchantID(enchant)))
		},
		applyEnchant: func(projectile *Object, enchant int32, duration int16, power uint8) {
			deps.runtime.ApplyEnchant(projectile, EnchantID(enchant), duration, power)
		},
		spellAudio: deps.spellAudio,
		audio:      deps.audio,
	})
}

func createSpellProjectileDirection4FDDA0(direction int16) [2]float32 {
	// Deliberately index with the signed value. GAME.EXE sign-extends the
	// int16 direction before scaling it; wrapping through byte would turn a
	// corrupt negative direction into a valid row and conceal the fault.
	return sincosDirTable[int(direction)]
}

func createSpellProjectileClassifyDirection4FDDA0(value, threshold int32) int32 {
	if value > threshold {
		return 1
	}
	if value >= -threshold {
		return 0
	}
	return -1
}

func createSpellProjectileIndexedDirection4FDDA0(direction int16, scratch *types.Pointf) {
	// Force the same signed 0..255 domain before converting to uintptr for
	// the extracted integer table lookup.
	_ = createSpellProjectileDirection4FDDA0(direction)
	offset := uintptr(direction) * 8
	threshold := memmap.Int32(0x587000, 230092)
	bits := (*[2]int32)(unsafe.Pointer(scratch))
	bits[0] = createSpellProjectileClassifyDirection4FDDA0(
		memmap.Int32(0x587000, 192088+offset), threshold,
	)
	bits[1] = createSpellProjectileClassifyDirection4FDDA0(
		memmap.Int32(0x587000, 192092+offset), threshold,
	)
}

func createSpellProjectileServerDeps4FDDA0(
	s *Server,
	runtime CreateSpellProjectileRuntime4FDDA0,
) createSpellProjectileNativeDeps4FDDA0 {
	return createSpellProjectileNativeDeps4FDDA0{
		spellFlags: func(spellID int32) uint32 {
			return uint32(s.Spells.Flags(spell.ID(spellID)))
		},
		searchTarget: func(
			aim *types.Pointf,
			source *Object,
			flags uint32,
			distance float32,
			mode int32,
			self *Object,
		) *Object {
			return s.Nox_xxx_spellFlySearchTarget(
				aim, source, things.SpellFlags(flags), distance, int(mode), self,
			)
		},
		mapTrace: func(
			origin, destination types.Pointf,
			outPoint *types.Pointf,
			outGrid *image.Point,
			flags int32,
		) int32 {
			if s.MapTraceRayAt(origin, destination, outPoint, outGrid, MapTraceFlags(flags)) {
				return 1
			}
			return 0
		},
		newObject: s.NewObjectByTypeID,
		directionX: func(direction int16) float32 {
			return createSpellProjectileDirection4FDDA0(direction)[0]
		},
		directionY: func(direction int16) float32 {
			return createSpellProjectileDirection4FDDA0(direction)[1]
		},
		indexedDirection: createSpellProjectileIndexedDirection4FDDA0,
		spellAudio: func(spellID, field int32) int32 {
			return int32(s.Spells.DefByInd(spell.ID(spellID)).GetAudio(int(field)))
		},
		audio: func(audioID int32, object *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(audioID), object, int(kind), code)
		},
		runtime: runtime,
	}
}

// CreateSpellProjectile4FDDA0 binds GAME.EXE 004FDDA0 to native-width Object
// pointers. The generic model retains the original load/callback order and
// scalar narrowing; this adapter performs no pointer-to-integer conversion.
//
//go:noinline
func (s *Server) CreateSpellProjectile4FDDA0(
	source, target *Object,
	spellID spell.ID,
	runtime CreateSpellProjectileRuntime4FDDA0,
) *Object {
	return createSpellProjectileNative4FDDA0(
		source, target, int32(spellID),
		createSpellProjectileServerDeps4FDDA0(s, runtime),
	)
}
