package server

// PickupUseRuntime4F34D0 supplies the root-owned object-list operations used
// by UsePickup's nested DefaultPickup call.
type PickupUseRuntime4F34D0 struct {
	DefaultPickup PickupDefaultRuntime4F31E0
}

type pickupUseNativeDeps4F34D0 struct {
	useByNetCode  func(*Object, *Object) int32
	defaultPickup func(*Object, *Object, int32, int32) int32
}

func pickupUseNative4F34D0(
	owner, item *Object,
	arg3, arg4 int32,
	deps pickupUseNativeDeps4F34D0,
) int32 {
	return pickupUse4F34D0(
		owner,
		item,
		arg3,
		arg4,
		pickupUseHooks4F34D0[*Object]{
			useByNetCode: deps.useByNetCode,
			loadFlagsLow: func(item *Object) uint8 {
				return uint8(item.ObjFlags)
			},
			defaultPickup: deps.defaultPickup,
		},
	)
}

// pickupUseByNetCode4F34D0 reproduces the observable part of GAME.EXE
// 0053F8E0 required by UsePickup without entering its ABI32 C body. A nil
// item or nil Use slot returns one before reading owner state. Otherwise the
// special-player gate suppresses Use, and every other owner calls the live
// native-width Use function with (owner, item). UsePickup discards the result.
func (s *Server) pickupUseByNetCode4F34D0(owner, item *Object) int32 {
	if item == nil {
		return 1
	}
	use := item.Use.Get()
	if use == nil {
		return 1
	}
	if s.Players.CheckXxx(owner) {
		return 1
	}
	if use(owner, item) {
		return 1
	}
	return 0
}

func pickupUseServerDeps4F34D0(
	s *Server,
	runtime PickupUseRuntime4F34D0,
) pickupUseNativeDeps4F34D0 {
	return pickupUseNativeDeps4F34D0{
		useByNetCode: s.pickupUseByNetCode4F34D0,
		defaultPickup: func(owner, item *Object, arg3, arg4 int32) int32 {
			return s.PickupDefault4F31E0(owner, item, arg3, arg4, runtime.DefaultPickup)
		},
	}
}

// PickupUse4F34D0 binds GAME.EXE's registered four-argument UsePickup
// callback to native-width Object, Use-function, and DefaultPickup paths.
func (s *Server) PickupUse4F34D0(
	owner, item *Object,
	arg3, arg4 int32,
	runtime PickupUseRuntime4F34D0,
) int32 {
	return pickupUseNative4F34D0(
		owner,
		item,
		arg3,
		arg4,
		pickupUseServerDeps4F34D0(s, runtime),
	)
}
