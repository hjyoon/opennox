package server

import "github.com/opennox/libs/types"

func signCollideNative4EAB40(
	source, target *Object,
	collision *types.Pointf,
) {
	signCollide4EAB40(
		source,
		target,
		collision,
		signCollideHooks4EAB40[*Object, UseFunc]{
			classLow: func(obj *Object) uint8 {
				return uint8(obj.ObjClass)
			},
			loadUse: func(obj *Object) UseFunc {
				return obj.Use.Get()
			},
			callUse: func(use UseFunc, target, source *Object) int32 {
				if use(target, source) {
					return 1
				}
				return 0
			},
		},
	)
}

// SignCollide4EAB40 binds the original registered callback to native Object
// class and Use-function fields.
func (s *Server) SignCollide4EAB40(
	source, target *Object,
	collision *types.Pointf,
) {
	signCollideNative4EAB40(source, target, collision)
}
