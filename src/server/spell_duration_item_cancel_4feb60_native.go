package server

type itemCancelDurSpellsNativeDeps4FEB60 struct {
	cancel func(int32, *Object)
}

func itemCancelDurSpellsNative4FEB60(
	owner, item *Object,
	deps itemCancelDurSpellsNativeDeps4FEB60,
) {
	ItemCancelDurSpells4FEB60(ItemCancelDurSpellsHooks4FEB60[*Object, *Object]{
		LoadItemArg: func() *Object {
			return item
		},
		LoadClass: func(item *Object) uint32 {
			return uint32(item.ObjClass)
		},
		LoadSubclass: func(item *Object) uint32 {
			return uint32(item.ObjSubClass)
		},
		LoadOwnerArg: func() *Object {
			return owner
		},
		Cancel: deps.cancel,
	})
}

// ItemCancelDurSpells4FEB60 binds GAME.EXE 004FEB60 to native-width Object
// pointers and the restored duration-spell cancellation routine. ObjClass and
// ObjSubClass remain exact unsigned dwords. The adapter deliberately performs
// raw item field loads so a nil item retains the original fault boundary.
//
//go:noinline
func (sp *SpellsDuration) ItemCancelDurSpells4FEB60(owner, item *Object) {
	itemCancelDurSpellsNative4FEB60(owner, item, itemCancelDurSpellsNativeDeps4FEB60{
		cancel: func(spellID int32, caster *Object) {
			sp.SpellCancelDurSpell4FEB10(spellID, caster)
		},
	})
}
