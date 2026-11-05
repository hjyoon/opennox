package server

// OwnerChainPlayerRead4EC580 exposes the actual typed memory operation to an
// optional diagnostic observer. Field zero is the owner pointer; field one
// is the low class byte. Ordinary production leaves the observer nil.
type OwnerChainPlayerRead4EC580 struct {
	Object   *Object
	Field    uint8
	Complete bool
	Owner    *Object
	Class    uint8
}

var ObserveOwnerChainPlayerRead4EC580 func(OwnerChainPlayerRead4EC580)

func ownerChainPlayerOwner4EC580(obj *Object) *Object {
	if ObserveOwnerChainPlayerRead4EC580 != nil {
		ObserveOwnerChainPlayerRead4EC580(OwnerChainPlayerRead4EC580{Object: obj})
	}
	owner := obj.ObjOwner
	if ObserveOwnerChainPlayerRead4EC580 != nil {
		ObserveOwnerChainPlayerRead4EC580(OwnerChainPlayerRead4EC580{Object: obj, Complete: true, Owner: owner})
	}
	return owner
}

func ownerChainPlayerClassLow4EC580(obj *Object) uint8 {
	if ObserveOwnerChainPlayerRead4EC580 != nil {
		ObserveOwnerChainPlayerRead4EC580(OwnerChainPlayerRead4EC580{Object: obj, Field: 1})
	}
	class := uint8(obj.ObjClass)
	if ObserveOwnerChainPlayerRead4EC580 != nil {
		ObserveOwnerChainPlayerRead4EC580(OwnerChainPlayerRead4EC580{Object: obj, Field: 1, Complete: true, Class: class})
	}
	return class
}
