package server

import (
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
)

// WeaponDropRuntime53AB10 supplies dependencies whose state is still owned by
// the legacy boundary. Object fields, audio, game flags, TickRate, and decay
// scheduling are handled by native Server services.
type WeaponDropRuntime53AB10 struct {
	DefaultDrop   func(*Object, *Object, *types.Pointf) int32
	ServerSubFlag func(uint32) int32
}

// ArmorDropRuntime53EB70 is the armor counterpart to
// WeaponDropRuntime53AB10.
type ArmorDropRuntime53EB70 struct {
	DefaultDrop   func(*Object, *Object, *types.Pointf) int32
	ServerSubFlag func(uint32) int32
}

type equipmentDropNativeDeps53AB10 struct {
	defaultDrop   func(*Object, *Object, *types.Pointf) int32
	serverSubFlag func(uint32) int32
	audio         func(uint32, *Object, int32, uint32)
	gameFlag      func(uint32) int32
	loadGameFPS   func() uint32
	setDecay      func(*Object, uint32)
}

func equipmentDropNativeHooks53AB10(
	owner, item *Object,
	point *types.Pointf,
	dropSound func(*Object),
	deps equipmentDropNativeDeps53AB10,
) equipmentDropHooks53AB10[*Object, *types.Pointf] {
	return equipmentDropHooks53AB10[*Object, *types.Pointf]{
		loadPointArg:  func() *types.Pointf { return point },
		loadOwnerArg:  func() *Object { return owner },
		loadItemArg:   func() *Object { return item },
		defaultDrop:   deps.defaultDrop,
		dropSound:     dropSound,
		gameFlag:      deps.gameFlag,
		serverSubFlag: deps.serverSubFlag,
		loadGameFPS:   deps.loadGameFPS,
		setDecay:      deps.setDecay,
	}
}

func weaponDropNative53AB10(
	owner, item *Object,
	point *types.Pointf,
	deps equipmentDropNativeDeps53AB10,
) int32 {
	dropSound := func(item *Object) {
		weaponDropSound53AAB0(weaponDropSoundHooks53AAB0[*Object]{
			loadItemArg: func() *Object { return item },
			loadClass: func(item *Object) uint32 {
				return uint32(item.ObjClass)
			},
			loadMaterial: func(item *Object) uint16 {
				return item.Material
			},
			audio: deps.audio,
		})
	}
	return weaponDrop53AB10(equipmentDropNativeHooks53AB10(owner, item, point, dropSound, deps))
}

func armorDropNative53EB70(
	owner, item *Object,
	point *types.Pointf,
	deps equipmentDropNativeDeps53AB10,
) int32 {
	dropSound := func(item *Object) {
		armorDropSound53EAE0(armorDropSoundHooks53EAE0[*Object]{
			loadItemArg: func() *Object { return item },
			loadMaterial: func(item *Object) uint16 {
				return item.Material
			},
			loadSubClass: func(item *Object) uint32 {
				return uint32(item.ObjSubClass)
			},
			audio: deps.audio,
		})
	}
	return armorDrop53EB70(equipmentDropNativeHooks53AB10(owner, item, point, dropSound, deps))
}

func equipmentDropServerDeps53AB10(
	s *Server,
	defaultDrop func(*Object, *Object, *types.Pointf) int32,
	serverSubFlag func(uint32) int32,
) equipmentDropNativeDeps53AB10 {
	return equipmentDropNativeDeps53AB10{
		defaultDrop:   defaultDrop,
		serverSubFlag: serverSubFlag,
		audio: func(id uint32, item *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), item, int(kind), code)
		},
		gameFlag: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		loadGameFPS: s.TickRate,
		setDecay: func(item *Object, delay uint32) {
			s.DecaySetTime511660(item, delay)
		},
	}
}

// WeaponDrop53AB10 binds GAME.EXE 0053AB10 to native-width Object and Pointf
// pointers.
func (s *Server) WeaponDrop53AB10(
	owner, item *Object,
	point *types.Pointf,
	runtime WeaponDropRuntime53AB10,
) int32 {
	deps := equipmentDropServerDeps53AB10(s, runtime.DefaultDrop, runtime.ServerSubFlag)
	return weaponDropNative53AB10(owner, item, point, deps)
}

// ArmorDrop53EB70 binds GAME.EXE 0053EB70 to native-width Object and Pointf
// pointers.
func (s *Server) ArmorDrop53EB70(
	owner, item *Object,
	point *types.Pointf,
	runtime ArmorDropRuntime53EB70,
) int32 {
	deps := equipmentDropServerDeps53AB10(s, runtime.DefaultDrop, runtime.ServerSubFlag)
	return armorDropNative53EB70(owner, item, point, deps)
}
