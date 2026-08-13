package server

type objectPixieTargetClearNativeDeps4E81D0 struct {
	loadPixieTypeID  func() uint32
	lookupObjectType func(string) uint32
	storePixieTypeID func(uint32)
}

func objectPixieTargetClearNative4E81D0(
	obj *Object,
	deps objectPixieTargetClearNativeDeps4E81D0,
) objectPixieTargetClearResult4E81D0[*PixieUpdateData] {
	return objectPixieTargetClear4E81D0(obj, objectPixieTargetClearHooks4E81D0[*Object, *PixieUpdateData]{
		loadPixieTypeID:  deps.loadPixieTypeID,
		lookupObjectType: deps.lookupObjectType,
		storePixieTypeID: deps.storePixieTypeID,
		loadTypeInd: func(obj *Object) uint16 {
			return obj.TypeInd
		},
		loadUpdateData: func(obj *Object) *PixieUpdateData {
			return obj.UpdateDataPixie()
		},
		clearTarget: func(updateData *PixieUpdateData) {
			updateData.Target = nil
		},
	})
}

// ClearPixieTarget4E81D0 clears the current target of a Pixie object. Both
// original callers discard EAX, so the mixed scalar/pointer return artifact is
// retained only by the internal semantic contract.
func (s *Server) ClearPixieTarget4E81D0(obj *Object) {
	_ = objectPixieTargetClearNative4E81D0(obj, objectPixieTargetClearNativeDeps4E81D0{
		loadPixieTypeID: func() uint32 {
			return uint32(s.Types.fast.pixie)
		},
		lookupObjectType: func(id string) uint32 {
			return uint32(s.Types.IndByID(id))
		},
		storePixieTypeID: func(typeID uint32) {
			s.Types.fast.pixie = int(typeID)
		},
	})
}
