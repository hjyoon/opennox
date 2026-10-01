package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
)

const fistCastCacheBase52D3C0 = uintptr(0x5D4594)

type fistCastDeps52D3C0 struct {
	loadCache   func(int32) uint32
	storeCache  func(int32, uint32)
	lookupType  func(string) uint32
	priorityMsg func(*Object, string, byte)
	hasGameFlag func(uint32) bool
	traceRay    func(types.Pointf, types.Pointf, MapTraceFlags) bool
	newObject   func(uint32) *Object
	balance     func(string, int32) float64
	createAt    func(*Object, *Object, types.Pointf)
	raise       func(*Object, float32)
	speed       func(string) float64
	castAudio   func(int32, *Object)
	inform      func(uint8, byte, int32)
}

// fistCast52D3C0 restores GAME.EXE 0052D3C0. Owner-list, caster, argument,
// update-data and Player pointers are native-width; only the original type
// cache, damage, flags and packet values remain DWORDs. The second object
// argument and SpellAcceptArg.Obj are deliberately unused by the original.
func fistCast52D3C0(id int32, owner, caster *Object, arg *SpellAcceptArg, level int32, h fistCastDeps52D3C0) int32 {
	if h.loadCache(0) == 0 {
		for i, name := range [5]string{"SmallFist", "MediumFist", "LargeFist", "LargeFist", "LargeFist"} {
			h.storeCache(int32(i+1), h.lookupType(name))
		}
		h.storeCache(0, 1)
	}
	for owned := owner.Field129; owned != nil; owned = owned.Field128 {
		kind := uint32(owned.TypeInd)
		for slot := int32(1); slot <= 5; slot++ {
			if kind == h.loadCache(slot) {
				h.priorityMsg(owner, "ExecSpel.c:TooManyFists", 0)
				return 0
			}
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
	fist := h.newObject(h.loadCache(level))
	if fist != nil {
		data := (*FistUpdateData)(fist.UpdateData)
		// 0052D534 spills the balance value to binary32 before FISTP.
		data.Damage = spellDurationRoundNearestEven(float32(h.balance("FistOfVengeanceDamage", level-1)))
		h.createAt(fist, caster, arg.Pos)
		fist.Field5 |= 0x20
		h.raise(fist, 255)
		velocity := h.speed("FistSpeed")
		flags := fist.ObjFlags | object.Flags(0x800000)
		fist.Field27 = float32(-velocity)
		fist.ObjFlags = flags
		fist.Field29 = 0x41100000
		h.castAudio(id, caster)
	}
	// The original reports success even when object allocation returns nil.
	return 1
}

func fistCastCacheOffset52D3C0(slot int32) uintptr {
	// Preserve the original DWORD index arithmetic, including slot zero's
	// initialized marker; do not clamp a level to another projectile type.
	return uintptr(uint32(2487736) + 4*uint32(slot))
}

type FistCastRuntime52D3C0 struct {
	CreateAt func(*Object, *Object, types.Pointf)
}

func (s *Server) CastFist52D3C0(id int32, _ *Object, owner, caster *Object, arg *SpellAcceptArg, level int32, runtime FistCastRuntime52D3C0) int32 {
	return fistCast52D3C0(id, owner, caster, arg, level, fistCastDeps52D3C0{
		loadCache: func(slot int32) uint32 {
			return memmap.Uint32(fistCastCacheBase52D3C0, fistCastCacheOffset52D3C0(slot))
		},
		storeCache: func(slot int32, value uint32) {
			*memmap.PtrUint32(fistCastCacheBase52D3C0, fistCastCacheOffset52D3C0(slot)) = value
		},
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
		castAudio: func(id int32, caster *Object) {
			s.Audio.EventObj(s.Spells.DefByInd(spell.ID(id)).GetCastSound(), caster, 0, 0)
		},
		inform: func(index uint8, code byte, value int32) {
			s.NetInformTextMsg(ntype.PlayerInd(index), code, int(value))
		},
	})
}
