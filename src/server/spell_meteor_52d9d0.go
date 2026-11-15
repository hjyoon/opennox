package server

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
)

type meteorCastDeps52D9D0 struct {
	loadCache   func() uint32
	storeCache  func(uint32)
	lookupType  func(string) uint32
	priorityMsg func(*Object, string, byte)
	hasGameFlag func(uint32) bool
	traceRay    func(types.Pointf, types.Pointf, MapTraceFlags) bool
	newObject   func(uint32) *Object
	balance     func(string, int32) float64
	createAt    func(*Object, *Object, types.Pointf)
	raise       func(*Object, float32)
	speed       func(string) float64
	castSound   func(int32) sound.ID
	audio       func(sound.ID, types.Pointf, int, uint32)
	inform      func(uint8, byte, int32)
}

// meteorCast52D9D0 restores GAME.EXE 0052D9D0. Object, owned-list, argument,
// update-data and Player pointers are native-width. Type cache, level, damage
// and flags retain their original DWORD widths. The second object argument
// and SpellAcceptArg.Obj are deliberately unused by the original.
func meteorCast52D9D0(id int32, owner, caster *Object, arg *SpellAcceptArg, level int32, h meteorCastDeps52D9D0) int32 {
	return MeteorNative52D9D0(id, owner, caster, arg, level, MeteorNativeRuntime52D9D0{
		LoadCache:   h.loadCache,
		StoreCache:  h.storeCache,
		LookupType:  h.lookupType,
		PriorityMsg: h.priorityMsg,
		HasGameFlag: h.hasGameFlag,
		TraceRay:    h.traceRay,
		NewObject:   h.newObject,
		Balance:     h.balance,
		CreateAt:    h.createAt,
		Raise:       h.raise,
		Speed:       h.speed,
		CastSound:   h.castSound,
		Audio:       h.audio,
		Inform:      h.inform,
	})
}

// MeteorCastRuntime52D9D0 binds the original C-owned cache DWORD and ordinary
// world placement. The cache is not a pointer stored in the retired PE32 blob.
type MeteorCastRuntime52D9D0 struct {
	TypeCache *uint32
	CreateAt  func(*Object, *Object, types.Pointf)
}

func (s *Server) CastMeteor52D9D0(id int32, _ *Object, owner, caster *Object, arg *SpellAcceptArg, level int32, runtime MeteorCastRuntime52D9D0) int32 {
	return meteorCast52D9D0(id, owner, caster, arg, level, meteorCastDeps52D9D0{
		loadCache:  func() uint32 { return *runtime.TypeCache },
		storeCache: func(kind uint32) { *runtime.TypeCache = kind },
		lookupType: func(name string) uint32 { return uint32(s.Types.IndByID(name)) },
		priorityMsg: func(unit *Object, message string, arg byte) {
			s.NetPriMsgToPlayer(unit, strman.ID(message), arg)
		},
		hasGameFlag: func(flag uint32) bool { return noxflags.HasGame(noxflags.GameFlag(flag)) },
		traceRay:    s.MapTraceRay,
		newObject:   func(kind uint32) *Object { return s.NewObjectByTypeInd(int(kind)) },
		balance:     func(key string, index int32) float64 { return s.Balance.FloatInd(key, int(index)) },
		createAt:    runtime.CreateAt,
		raise:       (*Object).Raise,
		speed:       s.Balance.Float,
		castSound:   func(id int32) sound.ID { return s.Spells.DefByInd(spell.ID(id)).GetCastSound() },
		audio:       s.Audio.EventPos,
		inform: func(index uint8, code byte, value int32) {
			s.NetInformTextMsg(ntype.PlayerInd(index), code, int(value))
		},
	})
}
