package server

type questLoseGemsNativeDeps54D080 struct {
	lookupType    func(string) uint32
	price         func(*Object) int32
	delayedDelete func(*Object)
	addGold       func(*Object, int32)
}

func questLoseGemsNative54D080(unit *Object, types [3]*uint32, deps questLoseGemsNativeDeps54D080) {
	questLoseGems54D080(unit, questLoseGemsHooks54D080[*Object]{
		loadType:      func(slot int) uint32 { return *types[slot] },
		storeType:     func(slot int, value uint32) { *types[slot] = value },
		lookupType:    deps.lookupType,
		first:         func(unit *Object) *Object { return unit.InvFirstItem },
		next:          func(item *Object) *Object { return item.InvNextItem },
		typeIndex:     func(item *Object) uint16 { return item.TypeInd },
		price:         deps.price,
		delayedDelete: deps.delayedDelete,
		addGold:       deps.addGold,
	})
}

// QuestLoseGems54D080 uses the original three DWORD cache locations and the
// native mode-one, nil-merchant pricing path. Gold/deletion remain the game's
// existing services; no inventory snapshots or optional-binding guards enter
// the original helper.
func (s *Server) QuestLoseGems54D080(unit *Object, types [3]*uint32, delayedDelete func(*Object), addGold func(*Object, int32)) {
	questLoseGemsNative54D080(unit, types, questLoseGemsNativeDeps54D080{
		lookupType:    func(name string) uint32 { return uint32(s.Types.IndByID(name)) },
		price:         s.ShopItemCostNoMerchant50E3D0,
		delayedDelete: delayedDelete,
		addGold:       addGold,
	})
}
