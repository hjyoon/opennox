package server

func playerConfusedDirectionNative4F7A40(s *Server, unit *Object) uint32 {
	return playerConfusedDirection4F7A40(unit, playerConfusedDirectionHooks4F7A40[*Object]{
		loadDirection2: func(obj *Object) uint16 {
			return uint16(obj.Direction2)
		},
		loadBuffPower: func(obj *Object, buff uint32) uint8 {
			return uint8(obj.EnchantPower(EnchantID(buff)))
		},
		loadFrame: s.Frame,
		loadNetCode: func(obj *Object) uint32 {
			return obj.NetCode
		},
	})
}

// PlayerConfusedDirection4F7A40 binds the original PE32 calculation to the
// native-width Object layout while retaining its fixed-width field semantics.
func (s *Server) PlayerConfusedDirection4F7A40(unit *Object) Dir16 {
	return Dir16(playerConfusedDirectionNative4F7A40(s, unit))
}
