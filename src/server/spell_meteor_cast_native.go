package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
)

// MeteorShowerNativeRuntime52D8A0 supplies the original cast observation boundaries.
// Balance carries original binary32 game-data values as float64.
type MeteorShowerNativeRuntime52D8A0 struct {
	LoadCache   func() uint32
	StoreCache  func(uint32)
	LookupType  func(string) uint32
	HasGameFlag func(uint32) bool
	TraceRay    func(types.Pointf, types.Pointf, MapTraceFlags) bool
	NewObject   func(uint32) *Object
	Balance     func(string, int32) float64
	CreateAt    func(*Object, *Object, types.Pointf)
	CastSound   func(int32) sound.ID
	Audio       func(sound.ID, *Object, int, uint32)
	Inform      func(uint8, byte, int32)
}

// MeteorShowerNative52D8A0 uses native owners, arguments and update/player links.
// Original startup FISTP truncates signed damage; invalid values yield MinInt32.
func MeteorShowerNative52D8A0(id int32, owner, caster *Object, arg *SpellAcceptArg, level int32, rt MeteorShowerNativeRuntime52D8A0) int32 {
	if rt.LoadCache() == 0 {
		rt.StoreCache(rt.LookupType("MeteorShower"))
	}
	from, to := caster.PosVec, arg.Pos
	flags := MapTraceFlags(73)
	if rt.HasGameFlag(2048) {
		flags = 9
	}
	if !rt.TraceRay(from, to, flags) {
		if caster.ObjClass.Has(object.ClassPlayer) {
			player := (*PlayerUpdateData)(caster.UpdateData).Player
			rt.Inform(player.PlayerInd, 0, 2)
		}
		return 0
	}
	shower := rt.NewObject(rt.LoadCache())
	if shower != nil {
		data := (*MeteorUpdateData)(shower.UpdateData)
		// 0052D974 spills binary32 before the original x87 FISTP conversion.
		data.Damage = netFXTruncateSignedDWORD419A70(float32(rt.Balance("MeteorDamage", level-1)))
		rt.CreateAt(shower, owner, arg.Pos)
		castSound := rt.CastSound(id)
		// The original attaches cast audio to the owner, not the caster or
		// destination, and does not set falling-Meteor height or velocity.
		rt.Audio(castSound, owner, 0, 0)
	}
	// Allocation failure still reports success at 0052D9B8.
	return 1
}

// MeteorNativeRuntime52D9D0 supplies the original cast observation boundaries.
// Balance and Speed carry original binary32 game-data values as float64.
type MeteorNativeRuntime52D9D0 struct {
	LoadCache   func() uint32
	StoreCache  func(uint32)
	LookupType  func(string) uint32
	PriorityMsg func(*Object, string, byte)
	HasGameFlag func(uint32) bool
	TraceRay    func(types.Pointf, types.Pointf, MapTraceFlags) bool
	NewObject   func(uint32) *Object
	Balance     func(string, int32) float64
	CreateAt    func(*Object, *Object, types.Pointf)
	Raise       func(*Object, float32)
	Speed       func(string) float64
	CastSound   func(int32) sound.ID
	Audio       func(sound.ID, types.Pointf, int, uint32)
	Inform      func(uint8, byte, int32)
}

// MeteorNative52D9D0 uses native owners, arguments and update/player links.
// Original startup FISTP truncates signed damage; invalid values yield MinInt32.
func MeteorNative52D9D0(id int32, owner, caster *Object, arg *SpellAcceptArg, level int32, rt MeteorNativeRuntime52D9D0) int32 {
	kind := rt.LoadCache()
	if kind == 0 {
		kind = rt.LookupType("Meteor")
		rt.StoreCache(kind)
	}
	// The original keeps the full cached DWORD in EAX throughout this walk;
	// the object TypeInd is zero-extended, not the cache truncated to a WORD.
	for owned := owner.Field129; owned != nil; owned = owned.Field128 {
		if uint32(owned.TypeInd) == kind {
			rt.PriorityMsg(owner, "ExecSpel.c:TooManyMeteors", 0)
			return 0
		}
	}
	from, to := caster.PosVec, arg.Pos
	flags := MapTraceFlags(73)
	if rt.HasGameFlag(2048) {
		flags = 9
	}
	if !rt.TraceRay(from, to, flags) {
		if caster.ObjClass.Has(object.ClassPlayer) {
			player := (*PlayerUpdateData)(caster.UpdateData).Player
			rt.Inform(player.PlayerInd, 0, 2)
		}
		return 0
	}
	// Unlike the owned-list comparison, allocation reloads the live cache.
	meteor := rt.NewObject(rt.LoadCache())
	if meteor != nil {
		data := (*MeteorUpdateData)(meteor.UpdateData)
		// 0052DAE8 spills the balance result to binary32 before FISTP.
		data.Damage = netFXTruncateSignedDWORD419A70(float32(rt.Balance("MeteorDamage", level-1)))
		rt.CreateAt(meteor, owner, arg.Pos)
		meteor.Field5 |= 0x20
		rt.Raise(meteor, 255)
		meteor.Field27 = float32(-rt.Speed("MeteorSpeed"))
		castSound := rt.CastSound(id)
		rt.Audio(castSound, arg.Pos, 0, 0)
	}
	// Allocation failure still reports success in the original.
	return 1
}
