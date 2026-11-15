package server

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
)

type meteorShowerCastDeps52D8A0 struct {
	loadCache   func() uint32
	storeCache  func(uint32)
	lookupType  func(string) uint32
	hasGameFlag func(uint32) bool
	traceRay    func(types.Pointf, types.Pointf, MapTraceFlags) bool
	newObject   func(uint32) *Object
	balance     func(string, int32) float64
	createAt    func(*Object, *Object, types.Pointf)
	castSound   func(int32) sound.ID
	audio       func(sound.ID, *Object, int, uint32)
	inform      func(uint8, byte, int32)
}

// meteorShowerCast52D8A0 restores GAME.EXE 0052D8A0 with native-width
// object, argument, update-data and player pointers. The second object and
// SpellAcceptArg.Obj are unused; unlike Meteor, no owned-object gate exists.
func meteorShowerCast52D8A0(id int32, owner, caster *Object, arg *SpellAcceptArg, level int32, h meteorShowerCastDeps52D8A0) int32 {
	return MeteorShowerNative52D8A0(id, owner, caster, arg, level, MeteorShowerNativeRuntime52D8A0{
		LoadCache:   h.loadCache,
		StoreCache:  h.storeCache,
		LookupType:  h.lookupType,
		HasGameFlag: h.hasGameFlag,
		TraceRay:    h.traceRay,
		NewObject:   h.newObject,
		Balance:     h.balance,
		CreateAt:    h.createAt,
		CastSound:   h.castSound,
		Audio:       h.audio,
		Inform:      h.inform,
	})
}

// MeteorShowerCastRuntime52D8A0 keeps the original packed cache DWORD;
// this scalar is not a pointer and must not be widened with the objects.
type MeteorShowerCastRuntime52D8A0 struct {
	TypeCache *uint32
	CreateAt  func(*Object, *Object, types.Pointf)
}

func (s *Server) CastMeteorShower52D8A0(id int32, _ *Object, owner, caster *Object, arg *SpellAcceptArg, level int32, runtime MeteorShowerCastRuntime52D8A0) int32 {
	return meteorShowerCast52D8A0(id, owner, caster, arg, level, meteorShowerCastDeps52D8A0{
		loadCache:   func() uint32 { return *runtime.TypeCache },
		storeCache:  func(kind uint32) { *runtime.TypeCache = kind },
		lookupType:  func(name string) uint32 { return uint32(s.Types.IndByID(name)) },
		hasGameFlag: func(flag uint32) bool { return noxflags.HasGame(noxflags.GameFlag(flag)) },
		traceRay:    s.MapTraceRay,
		newObject:   func(kind uint32) *Object { return s.NewObjectByTypeInd(int(kind)) },
		balance:     func(key string, index int32) float64 { return s.Balance.FloatInd(key, int(index)) },
		createAt:    runtime.CreateAt,
		castSound:   func(id int32) sound.ID { return s.Spells.DefByInd(spell.ID(id)).GetCastSound() },
		audio: func(id sound.ID, owner *Object, kind int, code uint32) {
			s.Audio.EventObj(id, owner, kind, code)
		},
		inform: func(index uint8, code byte, value int32) {
			s.NetInformTextMsg(ntype.PlayerInd(index), code, int(value))
		},
	})
}
