package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

var (
	createFuncs    = make(map[string]unsafe.Pointer)
	initFuncs      = make(map[string]objectDefFunc)
	initParseFuncs = map[string]ObjectParseFunc{
		"SkullInit":     objectDirectionInitParse,
		"DirectionInit": objectDirectionInitParse,
		"BreakInit":     objectBreakInitParse536910,
	}
	updateFuncs       = make(map[string]objectDefFunc)
	updateParseFuncs  = make(map[string]ObjectParseFunc)
	collideFuncs      = make(map[string]objectDefFunc)
	collideParseFuncs = make(map[string]ObjectParseFunc)
	useFuncs          = make(map[string]objectDefFunc)
	useParseFuncs     = make(map[string]ObjectParseFunc)
	damageFuncs       = make(map[string]unsafe.Pointer)
	damageSoundFuncs  = make(map[string]unsafe.Pointer)
	deathFuncs        = make(map[string]objectDefFunc)
	deathParseFuncs   = make(map[string]ObjectParseFunc)
	dropFuncs         = make(map[string]unsafe.Pointer)
	dropParseFuncs    = map[string]ObjectParseFunc{
		"AudEventDrop": func(t *ObjectType, args []string) error {
			t.s.dropSoundTable.parse(t, args)
			return nil
		},
	}
	pickupFuncs      = make(map[string]unsafe.Pointer)
	pickupParseFuncs = map[string]ObjectParseFunc{
		"AudEventPickup": func(t *ObjectType, args []string) error {
			t.s.pickupSoundTable.parse(t, args)
			return nil
		},
	}
	xferFuncs = make(map[string]unsafe.Pointer)
)

var (
	DefaultDamage      unsafe.Pointer
	DefaultDamageSound unsafe.Pointer
	DefaultXfer        unsafe.Pointer
)

func init() {
	RegisterObjectCreate("NoCreate", nil)
	RegisterObjectCreate("PlayerCreate", nil)

	RegisterObjectInit("NoInit", nil, 0)
	RegisterObjectInit("RewardMarkerInit", nil, unsafe.Sizeof(RewardMarkerInitData{}))
	RegisterObjectInit("AnkhInit", nil, unsafe.Sizeof(AnkhInitData{}))

	RegisterObjectUpdate("NoUpdate", nil, 0)
	RegisterObjectUpdate("HomingProjectileUpdate", nil, 4)
	RegisterObjectUpdate("SpikeBlockUpdate", nil, 2200)
	RegisterObjectUpdate("TowerUpdate", nil, 8)
	RegisterObjectUpdate("WeaponArmorUpdate", nil, unsafe.Sizeof(WeaponArmorUpdateData{}))
	RegisterObjectUpdate("DamageRoundoffUpdate", nil, unsafe.Sizeof(WeaponArmorUpdateData{})) // used in demo instead of WeaponArmorUpdate

	RegisterObjectCollide("NoCollide", nil, 0)

	RegisterObjectUseC("AmmoUse", nil, unsafe.Sizeof(AmmoUseData{}))
	RegisterObjectUseC("BowUse", nil, 1)
}

type ObjectParseFunc func(objt *ObjectType, args []string) error

type objectDefFunc struct {
	Func     unsafe.Pointer
	DataSize uintptr
}

func RegisterObjectCreate(name string, fnc unsafe.Pointer) {
	if _, ok := createFuncs[name]; ok {
		panic("already registered")
	}
	createFuncs[name] = fnc
}

func RegisterObjectInit(name string, fnc unsafe.Pointer, sz uintptr) {
	if _, ok := initFuncs[name]; ok {
		panic("already registered")
	}
	initFuncs[name] = objectDefFunc{Func: fnc, DataSize: sz}
}

func RegisterObjectUpdate(name string, fnc unsafe.Pointer, sz uintptr) {
	if _, ok := updateFuncs[name]; ok {
		panic("already registered")
	}
	updateFuncs[name] = objectDefFunc{Func: fnc, DataSize: sz}
}

func RegisterObjectUpdateParse(name string, fnc ObjectParseFunc) {
	if _, ok := updateParseFuncs[name]; ok {
		panic("already registered")
	}
	updateParseFuncs[name] = fnc
}

func RegisterObjectCollide(name string, fnc unsafe.Pointer, sz uintptr) {
	if _, ok := collideFuncs[name]; ok {
		panic("already registered")
	}
	collideFuncs[name] = objectDefFunc{Func: fnc, DataSize: sz}
}

func RegisterObjectCollideParse(name string, fnc ObjectParseFunc) {
	if _, ok := collideParseFuncs[name]; ok {
		panic("already registered")
	}
	collideParseFuncs[name] = fnc
}

type UseFuncPtr struct {
	Ptr unsafe.Pointer
}

func (p UseFuncPtr) Get() UseFunc {
	if p.Ptr == nil {
		return nil
	}
	return objUse.Get(p.Ptr)
}

type UseFunc func(obj, obj2 *Object) bool

var objUse = ccall.NewFuncs(func(cfnc unsafe.Pointer) UseFunc {
	return func(obj, obj2 *Object) bool {
		return ccall.CallIntPtr2(cfnc, obj.CObj(), obj2.CObj()) != 0
	}
})

func RegisterObjectUseC(name string, cfnc unsafe.Pointer, sz uintptr) {
	if _, ok := useFuncs[name]; ok {
		panic("already registered")
	}
	useFuncs[name] = objectDefFunc{Func: cfnc, DataSize: sz}
}

func RegisterObjectUse(name string, cfnc unsafe.Pointer, fnc UseFunc, sz uintptr) {
	if _, ok := useFuncs[name]; ok {
		panic("already registered")
	}
	useFuncs[name] = objectDefFunc{Func: cfnc, DataSize: sz}
	objUse.Register(cfnc, fnc)
}

func RegisterObjectUseParse(name string, fnc ObjectParseFunc) {
	if _, ok := useParseFuncs[name]; ok {
		panic("already registered")
	}
	useParseFuncs[name] = fnc
}

func RegisterObjectDamage(name string, fnc unsafe.Pointer) {
	if _, ok := damageFuncs[name]; ok {
		panic("already registered")
	}
	damageFuncs[name] = fnc
}

// DamageFunc is the native-width counterpart of a thing.bin damage callback.
// Damage callbacks are especially likely to be reached from legacy C code, so
// restored handlers must not round their object arguments through PE32 ints.
type DamageFunc func(target, source, weapon *Object, damage int32, typ object.DamageType) bool

var objDamage = ccall.NewFuncs(func(cfnc unsafe.Pointer) DamageFunc {
	return func(target, source, weapon *Object, damage int32, typ object.DamageType) bool {
		// Unrestored callbacks still use five C ints internally. They are valid
		// on PE32 targets only; invoking one with a 64-bit object address would
		// silently truncate it before the first field access.
		if unsafe.Sizeof(uintptr(0)) != 4 {
			return false
		}
		return ccall.CallIntUPtr5(
			cfnc,
			uintptr(target.CObj()),
			uintptr(source.CObj()),
			uintptr(weapon.CObj()),
			uintptr(uint32(damage)),
			uintptr(uint32(typ)),
		) != 0
	}
})

// RegisterObjectDamageGo preserves the public C callback identity stored in
// object definitions while associating it with a pointer-width Go handler.
func RegisterObjectDamageGo(name string, cfnc unsafe.Pointer, fnc DamageFunc) {
	RegisterObjectDamage(name, cfnc)
	objDamage.Register(cfnc, fnc)
}

// CallObjectDamage dispatches restored callbacks without an indirect PE32 C
// call. Unknown callbacks retain their original path on 32-bit targets and
// fail closed on wider targets until their implementation is restored.
func CallObjectDamage(
	fnc unsafe.Pointer,
	target, source, weapon *Object,
	damage int32,
	typ object.DamageType,
) bool {
	if fnc == nil {
		return false
	}
	return objDamage.Get(fnc)(target, source, weapon, damage, typ)
}

func RegisterObjectDamageSound(name string, fnc unsafe.Pointer) {
	if _, ok := damageSoundFuncs[name]; ok {
		panic("already registered")
	}
	damageSoundFuncs[name] = fnc
}

func RegisterObjectDeath(name string, fnc unsafe.Pointer, sz uintptr) {
	if _, ok := deathFuncs[name]; ok {
		panic("already registered")
	}
	deathFuncs[name] = objectDefFunc{Func: fnc, DataSize: sz}
}

// DeathFunc is the native-width counterpart of a registered legacy death
// callback. The object remains a Go pointer all the way to handlers restored
// in Go, avoiding an unnecessary indirect C call through the PE32 boundary.
type DeathFunc func(obj *Object)

var objDeath = ccall.NewFuncs(func(cfnc unsafe.Pointer) DeathFunc {
	return func(obj *Object) {
		ccall.CallVoidPtr(cfnc, obj.CObj())
	}
})

// RegisterObjectDeathGo registers the public C callback identity used by
// thing.bin together with its native Go implementation.
func RegisterObjectDeathGo(name string, cfnc unsafe.Pointer, fnc DeathFunc, sz uintptr) {
	RegisterObjectDeath(name, cfnc, sz)
	objDeath.Register(cfnc, fnc)
}

// CallObjectDeath dispatches restored handlers without converting the object
// through C. Unrestored handlers retain the exact legacy indirect-call path.
func CallObjectDeath(fnc unsafe.Pointer, obj *Object) {
	objDeath.Get(fnc)(obj)
}

func RegisterObjectDeathParse(name string, fnc ObjectParseFunc) {
	if _, ok := deathParseFuncs[name]; ok {
		panic("already registered")
	}
	deathParseFuncs[name] = fnc
}

// ObjectDeathHandler returns the exact registered death callback and parser
// data size for name. It is useful for checking that a type loaded from
// thing.bin retained the intended native callback contract.
func ObjectDeathHandler(name string) (unsafe.Pointer, uintptr, bool) {
	def, ok := deathFuncs[name]
	return def.Func, def.DataSize, ok
}

type DropFuncPtr struct {
	Ptr unsafe.Pointer
}

func (p DropFuncPtr) Get() DropFunc {
	if p.Ptr == nil {
		return nil
	}
	return objDrop.Get(p.Ptr)
}

type DropFunc func(obj, obj2 *Object, pos *types.Pointf) int32

var objDrop = ccall.NewFuncs(func(cfnc unsafe.Pointer) DropFunc {
	return func(obj, obj2 *Object, pos *types.Pointf) int32 {
		return int32(ccall.CallIntPtr3(cfnc, obj.CObj(), obj2.CObj(), unsafe.Pointer(pos)))
	}
})

func RegisterObjectDropC(name string, cfnc unsafe.Pointer) {
	if _, ok := dropFuncs[name]; ok {
		panic("already registered")
	}
	dropFuncs[name] = cfnc
}

func RegisterObjectDrop(name string, cfnc unsafe.Pointer, fnc DropFunc) {
	if _, ok := dropFuncs[name]; ok {
		panic("already registered")
	}
	dropFuncs[name] = cfnc
	objDrop.Register(cfnc, fnc)
}

type PickupFuncPtr struct {
	Ptr unsafe.Pointer
}

func (p PickupFuncPtr) Get() PickupFunc {
	if p.Ptr == nil {
		return nil
	}
	return objPickup.Get(p.Ptr)
}

// CallInt32 invokes the registered C callback with its original four-argument
// ABI and preserves the complete 32-bit return value. Get deliberately keeps
// the older bool-facing Go API for callers that only need success/failure.
func (p PickupFuncPtr) CallInt32(who, it *Object, a3, a4 int32) int32 {
	return int32(ccall.CallIntUPtr4(
		p.Ptr,
		uintptr(who.CObj()),
		uintptr(it.CObj()),
		uintptr(a3),
		uintptr(a4),
	))
}

type PickupFunc func(who, it *Object, a3, a4 int) bool

var objPickup = ccall.NewFuncs(func(cfnc unsafe.Pointer) PickupFunc {
	return func(who, it *Object, a3, a4 int) bool {
		return ccall.CallIntUPtr4(cfnc, uintptr(who.CObj()), uintptr(it.CObj()), uintptr(a3), uintptr(a4)) != 0
	}
})

func RegisterObjectPickupC(name string, cfnc unsafe.Pointer) {
	if _, ok := pickupFuncs[name]; ok {
		panic("already registered")
	}
	pickupFuncs[name] = cfnc
}

func RegisterObjectPickup(name string, cfnc unsafe.Pointer, fnc PickupFunc) {
	if _, ok := pickupFuncs[name]; ok {
		panic("already registered")
	}
	pickupFuncs[name] = cfnc
	objPickup.Register(cfnc, fnc)
}

// ObjectPickupHandler returns the exact registered pickup callback for name.
// It is useful for constructing objects whose handler is selected at runtime
// instead of inherited from thing.bin.
func ObjectPickupHandler(name string) (PickupFuncPtr, bool) {
	cfnc, ok := pickupFuncs[name]
	return PickupFuncPtr{Ptr: cfnc}, ok
}

func RegisterObjectXfer(name string, fnc unsafe.Pointer) {
	if _, ok := xferFuncs[name]; ok {
		panic("already registered")
	}
	xferFuncs[name] = fnc
}
