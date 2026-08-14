package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

// MonsterArrowCollideData is the fixed-width form of the registered
// eight-byte record. GAME.EXE selects CoopDamage when game flag 0x800 is set
// and OtherDamage otherwise. Neither field contains a pointer, so the layout
// remains eight bytes on native 32-bit and 64-bit targets.
type MonsterArrowCollideData struct {
	CoopDamage  int32
	OtherDamage int32
}

// MonsterArrowCollideRuntime4EB800 supplies effects whose implementations
// remain above package server. Object and collide-data reads stay native-width
// inside MonsterArrowCollide4EB800.
type MonsterArrowCollideRuntime4EB800 struct {
	TraceHitPoint func() *ntype.Point32
	DamageMap     func(int32, int32, int32, object.DamageType, *Object)
	DelayedDelete func(*Object)
}

type monsterArrowCollideNativeDeps4EB800 struct {
	gameFlag      func(uint32) bool
	findParent    func(*Object) *Object
	targetDamage  func(*Object, *Object, *Object, int32, object.DamageType) int32
	tracePoint    func() *ntype.Point32
	damageMap     func(int32, int32, int32, object.DamageType, *Object)
	delayedDelete func(*Object)
}

func monsterArrowCollideNative4EB800(
	source, target *Object,
	collision *types.Pointf,
	deps monsterArrowCollideNativeDeps4EB800,
) {
	monsterArrowCollide4EB800(source, target, collision, monsterArrowCollideHooks4EB800[
		*Object,
		*MonsterArrowCollideData,
		*ntype.Point32,
	]{
		loadCollideData: func(obj *Object) *MonsterArrowCollideData {
			return (*MonsterArrowCollideData)(obj.CollideData)
		},
		gameFlag: deps.gameFlag,
		loadCoopDamage: func(data *MonsterArrowCollideData) int32 {
			return data.CoopDamage
		},
		loadOtherDamage: func(data *MonsterArrowCollideData) int32 {
			return data.OtherDamage
		},
		loadTargetFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		findParent: deps.findParent,
		targetDamage: func(target, parent, source *Object, damage int32, damageType uint32) int32 {
			return deps.targetDamage(target, parent, source, damage, object.DamageType(damageType))
		},
		tracePoint: deps.tracePoint,
		loadTraceY: func(point *ntype.Point32) int32 {
			return point.Y
		},
		loadTraceX: func(point *ntype.Point32) int32 {
			return point.X
		},
		damageMap: func(x, y, damage int32, damageType uint32, source *Object) {
			deps.damageMap(x, y, damage, object.DamageType(damageType), source)
		},
		delayedDelete: deps.delayedDelete,
	})
}

// MonsterArrowCollide4EB800 binds the registered collision callback to
// native-width Object and fixed-width MonsterArrowCollideData layouts.
func (s *Server) MonsterArrowCollide4EB800(
	source, target *Object,
	collision *types.Pointf,
	runtime MonsterArrowCollideRuntime4EB800,
) {
	monsterArrowCollideNative4EB800(source, target, collision, monsterArrowCollideNativeDeps4EB800{
		gameFlag: func(flag uint32) bool {
			return noxflags.HasGame(noxflags.GameFlag(flag))
		},
		findParent: (*Object).FindOwnerChainPlayer,
		targetDamage: func(target, parent, source *Object, damage int32, damageType object.DamageType) int32 {
			return int32(ccall.CallIntUPtr5(
				target.Damage,
				uintptr(target.CObj()),
				uintptr(toObjectC(parent)),
				uintptr(source.CObj()),
				uintptr(uint32(damage)),
				uintptr(damageType),
			))
		},
		tracePoint:    runtime.TraceHitPoint,
		damageMap:     runtime.DamageMap,
		delayedDelete: runtime.DelayedDelete,
	})
}

var (
	_ = [1]struct{}{}[8-unsafe.Sizeof(MonsterArrowCollideData{})]
	_ = [1]struct{}{}[unsafe.Offsetof(MonsterArrowCollideData{}.CoopDamage)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(MonsterArrowCollideData{}.OtherDamage)]
)
