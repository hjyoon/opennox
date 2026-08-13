package legacy

import "github.com/opennox/libs/object"

type teleportToMB4E7190Hooks[T any] struct {
	anchored func(T) bool
	flags    func(T) object.Flags
	quest    func() bool
	class    func(T) object.Class
	subclass func(T) object.SubClass
	coop     func() bool
	move     func(T)
}

// teleportToMB4E7190 preserves the observable load and call order of the
// movement gate at 004E7190. In particular, class is reloaded after the Coop
// check instead of being cached from the Quest shopkeeper check.
func teleportToMB4E7190[T any](obj T, h teleportToMB4E7190Hooks[T]) {
	if h.anchored(obj) {
		return
	}
	if h.flags(obj).Has(object.FlagNoUpdate) {
		return
	}
	if h.quest() && h.class(obj).Has(object.ClassMonster) &&
		h.subclass(obj).AsMonster().Has(object.MonsterShopkeeper) {
		return
	}
	if !h.coop() && !h.class(obj).HasAny(object.MaskUnits) {
		return
	}
	h.move(obj)
}
