package server

import (
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

// MakeScorchRuntime537AF0 supplies object placement, which remains owned by
// the outer game server. Type lookup, random selection and decay scheduling
// stay in the native-width server representation.
type MakeScorchRuntime537AF0 struct {
	CreateObjectAt func(*Object, *Object, types.Pointf)
}

// MakeScorch537AF0 creates one of the original three floor scorch objects and
// gives it the original mode-dependent lifetime.
func (s *Server) MakeScorch537AF0(pos types.Pointf, kind int, runtime MakeScorchRuntime537AF0) {
	if runtime.CreateObjectAt == nil {
		return
	}
	makeScorch537AF0(pos, kind, makeScorchHooks537AF0[*Object]{
		newObject: s.NewObjectByTypeID,
		createObjectAt: func(obj *Object, pos types.Pointf) {
			runtime.CreateObjectAt(obj, nil, pos)
		},
		gameFlag: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		randomInt:   s.Rand.Logic.IntClamp,
		loadGameFPS: s.TickRate,
		setDecay: func(obj *Object, delay uint32) {
			s.DecaySetTime511660(obj, delay)
		},
	})
}
