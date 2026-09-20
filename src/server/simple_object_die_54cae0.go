package server

import "github.com/opennox/libs/object"

// SimpleObjectDeathRuntime54CAE0 supplies the side effects shared by the
// ImpEggDie and PotionDie callbacks.
type SimpleObjectDeathRuntime54CAE0 struct {
	Audio         func(uint32, *Object)
	DelayedDelete func(*Object)
}

// ImpEggDieNative54CAE0 restores GAME.EXE 0054CAE0 without passing the live
// object through the original callback's PE32-sized int parameter.
func ImpEggDieNative54CAE0(obj *Object, runtime SimpleObjectDeathRuntime54CAE0) uint32 {
	runtime.Audio(764, obj)
	obj.ObjFlags |= object.FlagNoCollide
	return uint32(obj.ObjFlags)
}

// PotionDieNative54CBB0 restores GAME.EXE 0054CBB0 with a native-width object
// pointer. Its audio event precedes the delayed deletion.
func PotionDieNative54CBB0(obj *Object, runtime SimpleObjectDeathRuntime54CAE0) {
	runtime.Audio(753, obj)
	runtime.DelayedDelete(obj)
}
