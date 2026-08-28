package server

// PickupAmmoRuntime4F3B00 contains the object-bearing operations still owned
// by the legacy/root boundary. Every object stays a native pointer; only
// callback arguments, charges, and PlayerInd retain fixed widths.
type PickupAmmoRuntime4F3B00 struct {
	WeaponPickup  func(*Object, *Object, int32, int32) int32
	ReportCharges func(uint8, *Object, uint8, uint8)
	DelayedDelete func(*Object)
	PickupAudio   func(*Object, *Object)
}

type pickupAmmoNativeDeps4F3B00 struct {
	weaponEquipFlags func(*Object) uint32
	weaponPickup     func(*Object, *Object, int32, int32) int32
	reportCharges    func(uint8, *Object, uint8, uint8)
	delayedDelete    func(*Object)
	pickupAudio      func(*Object, *Object)
}

func pickupAmmoUseByte4F3B00(data *AmmoUseData, index int) uint8 {
	switch index {
	case 0:
		return data.Charge0
	case 1:
		return data.Charge1
	case 2:
		return data.Field2
	default:
		panic("invalid AmmoUseData byte index")
	}
}

func pickupAmmoStoreUseByte4F3B00(data *AmmoUseData, index int, value uint8) {
	switch index {
	case 0:
		data.Charge0 = value
	case 1:
		data.Charge1 = value
	case 2:
		data.Field2 = value
	default:
		panic("invalid AmmoUseData byte index")
	}
}

func pickupAmmoNative4F3B00(
	owner, item *Object,
	arg3, arg4 int32,
	deps pickupAmmoNativeDeps4F3B00,
) int32 {
	return pickupAmmo4F3B00(owner, item, pickupAmmoHooks4F3B00[
		*Object,
		*ModifierInitData,
		*AmmoUseData,
		*ModifierEff,
		*PlayerUpdateData,
		*Player,
	]{
		weaponEquipFlags: deps.weaponEquipFlags,
		loadOwnerClassLow: func(owner *Object) uint8 {
			return uint8(owner.ObjClass)
		},
		loadOwnerUpdate: func(owner *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(owner.UpdateData)
		},
		loadInventoryHead: func(owner *Object) *Object {
			return owner.InvFirstItem
		},
		loadTypeInd: func(object *Object) uint16 {
			return object.TypeInd
		},
		loadObjectClass: func(object *Object) uint32 {
			return uint32(object.ObjClass)
		},
		loadInitData: func(object *Object) *ModifierInitData {
			return (*ModifierInitData)(object.InitData)
		},
		loadUseData: func(object *Object) *AmmoUseData {
			return (*AmmoUseData)(object.UseData.Ptr)
		},
		loadModifier: func(data *ModifierInitData, index int) *ModifierEff {
			return data.Modifiers[index]
		},
		loadUseByte:  pickupAmmoUseByte4F3B00,
		storeUseByte: pickupAmmoStoreUseByte4F3B00,
		loadInventoryNext: func(object *Object) *Object {
			return object.InvNextItem
		},
		loadUpdatePlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadPlayerInd: func(player *Player) uint8 {
			return player.PlayerInd
		},
		reportCharges: deps.reportCharges,
		delayedDelete: deps.delayedDelete,
		pickupAudio:   deps.pickupAudio,
		loadArg4: func() int32 {
			return arg4
		},
		loadArg3: func() int32 {
			return arg3
		},
		defaultPickup: deps.weaponPickup,
	})
}

func pickupAmmoServerDeps4F3B00(
	s *Server,
	runtime PickupAmmoRuntime4F3B00,
) pickupAmmoNativeDeps4F3B00 {
	return pickupAmmoNativeDeps4F3B00{
		weaponEquipFlags: s.Weapons.Nox_xxx_weaponInventoryEquipFlags_415820,
		weaponPickup:     runtime.WeaponPickup,
		reportCharges:    runtime.ReportCharges,
		delayedDelete:    runtime.DelayedDelete,
		pickupAudio:      runtime.PickupAudio,
	}
}

// PickupAmmo4F3B00 binds GAME.EXE's registered four-argument AmmoPickup to
// native-width Object, ModifierInitData, ModifierEff, PlayerUpdateData, and
// Player pointers plus the exact three-byte AmmoUseData payload.
func (s *Server) PickupAmmo4F3B00(
	owner, item *Object,
	arg3, arg4 int32,
	runtime PickupAmmoRuntime4F3B00,
) int32 {
	return pickupAmmoNative4F3B00(
		owner,
		item,
		arg3,
		arg4,
		pickupAmmoServerDeps4F3B00(s, runtime),
	)
}
