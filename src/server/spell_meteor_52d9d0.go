package server

import (
	"github.com/opennox/libs/object"
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
	kind := h.loadCache()
	if kind == 0 {
		kind = h.lookupType("Meteor")
		h.storeCache(kind)
	}
	// The original keeps the full cached DWORD in EAX throughout this walk;
	// the object TypeInd is zero-extended, not the cache truncated to a WORD.
	for owned := owner.Field129; owned != nil; owned = owned.Field128 {
		if uint32(owned.TypeInd) == kind {
			h.priorityMsg(owner, "ExecSpel.c:TooManyMeteors", 0)
			return 0
		}
	}
	from, to := caster.PosVec, arg.Pos
	flags := MapTraceFlags(73)
	if h.hasGameFlag(2048) {
		flags = 9
	}
	if !h.traceRay(from, to, flags) {
		if caster.ObjClass.Has(object.ClassPlayer) {
			player := (*PlayerUpdateData)(caster.UpdateData).Player
			h.inform(player.PlayerInd, 0, 2)
		}
		return 0
	}
	// Unlike the owned-list comparison, allocation reloads the live cache.
	meteor := h.newObject(h.loadCache())
	if meteor != nil {
		data := (*MeteorUpdateData)(meteor.UpdateData)
		// 0052DAE8 spills the balance result to binary32 before FISTP.
		data.Damage = spellDurationRoundNearestEven(float32(h.balance("MeteorDamage", level-1)))
		h.createAt(meteor, owner, arg.Pos)
		meteor.Field5 |= 0x20
		h.raise(meteor, 255)
		meteor.Field27 = float32(-h.speed("MeteorSpeed"))
		castSound := h.castSound(id)
		h.audio(castSound, arg.Pos, 0, 0)
	}
	// Allocation failure still reports success in the original.
	return 1
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
