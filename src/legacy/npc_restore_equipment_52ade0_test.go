//go:build amd64 || arm64

package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func npcRestoreTestServer52ADE0(t *testing.T) *server.Server {
	t.Helper()
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	oldGetServer := GetServer
	GetServer = func() Server { return &playerAttackLegacyServer538960{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameHost)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})
	for _, name := range []string{
		"LeatherArmor", "ChainTunic", "Breastplate", "LeatherHelm", "SteelHelm", "SteelShield", "WoodenShield",
		"StreetShirt", "StreetPants", "StreetSneakers", "LeatherBoots", "LeatherArmoredBoots", "PlateBoots",
		"MedievalPants", "LeatherLeggings", "ChainLeggings", "PlateLeggings", "MedievalShirt", "WizardRobe",
		"LeatherArmbands", "PlateArms", "MedievalCloak", "ChainCoif", "WizardHelm", "ConjurerHelm", "OrnateHelm",
		"LeatherArmorDestroyed", "BreastplateDestroyed", "SteelHelmDestroyed", "SteelShieldDestroyed", "WoodenShieldDestroyed", "INVALID",
	} {
		if err := srv.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
			t.Fatal(err)
		}
	}
	srv.Nox_xxx_equipArmor_415AB0()
	return srv
}

func npcRestoreTestObject52ADE0(t *testing.T) *server.Object {
	t.Helper()
	obj, free := alloc.New(server.Object{})
	t.Cleanup(free)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 {
		t.Fatalf("equipment fixture object %p must exercise a native high address", obj)
	}
	return obj
}

func TestNPCRestoreEquipped52ADE0RestoresClothingProtectionAndEngage(t *testing.T) {
	srv := npcRestoreTestServer52ADE0(t)
	for _, tc := range []struct {
		name   string
		sub    object.SubClass
		noDrop bool
		armor  float32
	}{
		{"MedievalPants", object.SubClass(object.ArmorPants), true, 0.1},
		{"MedievalShirt", object.SubClass(object.ArmorShirt), true, 0.2},
		{"StreetPants", object.SubClass(object.ArmorPants), true, 0.1},
		{"StreetShirt", object.SubClass(object.ArmorShirt), true, 0.2},
		{"StreetSneakers", object.SubClass(object.ArmorBoots), true, 0.1},
		{"LeatherArmor", object.SubClass(object.ArmorBreastplate), false, 0.3},
		{"MedievalCloak", object.SubClass(object.ArmorBack), false, 0.1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			owner, item := npcRestoreTestObject52ADE0(t), npcRestoreTestObject52ADE0(t)
			update, freeUpdate := alloc.New(server.MonsterUpdateData{})
			attrs, freeAttrs := alloc.New(server.ModifierInitData{})
			modifier, freeModifier := alloc.New(server.ModifierEff{})
			definition, freeDefinition := alloc.New(server.Modifier{})
			for _, free := range []func(){freeUpdate, freeAttrs, freeModifier, freeDefinition} {
				t.Cleanup(free)
			}
			definition.TypeInd = uint32(srv.Types.IndByID(tc.name))
			definition.DamageCoeffOrArmor64 = tc.armor
			srv.Modif.Dword_5d4594_251608 = definition
			t.Cleanup(func() { srv.Modif.Dword_5d4594_251608 = nil })
			modifier.Engage112 = modifierEngagePointerNative4DFBB0(1)
			attrs.Modifiers[2] = modifier
			owner.ObjClass, owner.ObjSubClass = object.ClassMonster, object.SubClass(object.MonsterNPC)
			owner.UpdateData, owner.InvFirstItem = unsafe.Pointer(update), item
			item.TypeInd, item.ObjClass, item.ObjSubClass = uint16(definition.TypeInd), object.ClassArmor, tc.sub
			item.ObjFlags, item.InvHolder, item.InitData = object.FlagEquipped, owner, unsafe.Pointer(attrs)
			update.Field518 = math.Float32bits(0.75)

			npcRestoreEquipped52ADE0(owner)

			if !item.Flags().Has(object.FlagEquipped) || item.Flags().Has(object.FlagNoAutoDrop) != tc.noDrop {
				t.Fatalf("restored %s flags=%#x, want equipped and no-auto-drop=%t", tc.name, item.ObjFlags, tc.noDrop)
			}
			if update.ArmorEquipFlags != srv.Armor.Sub_415DF0(tc.name) || update.Field518 != math.Float32bits(tc.armor) || owner.Field110 != 1 {
				t.Fatalf("restore skipped real equip services: appearance=%#x armor=%g engage=%#x", update.ArmorEquipFlags, math.Float32frombits(update.Field518), owner.Field110)
			}
			if tc.noDrop {
				// The production death drop loop must not dispatch this item.
				drops := 0
				srv.DropAllItems4EDA40(owner, server.DropAllItemsRuntime4EDA40{
					Dispatch: func(*server.Object, *server.Object, *types.Pointf) int32 {
						drops++
						return 0
					},
				})
				if drops != 0 || owner.InvFirstItem != item || item.InvHolder != owner {
					t.Fatal("restored base clothing was lost from the NPC inventory")
				}
			}
		})
	}
}

func TestNPCRestoreEquipped52ADE0LeavesLooseClothingAlone(t *testing.T) {
	srv := npcRestoreTestServer52ADE0(t)
	owner, item := npcRestoreTestObject52ADE0(t), npcRestoreTestObject52ADE0(t)
	update, freeUpdate := alloc.New(server.MonsterUpdateData{})
	t.Cleanup(freeUpdate)
	owner.ObjClass, owner.ObjSubClass, owner.UpdateData = object.ClassMonster, object.SubClass(object.MonsterNPC), unsafe.Pointer(update)
	owner.InvFirstItem = item
	item.TypeInd, item.ObjClass, item.InvHolder = uint16(srv.Types.IndByID("MedievalPants")), object.ClassArmor, owner
	update.Field518 = math.Float32bits(0.25)
	beforeOwner, beforeItem, beforeUpdate := *owner, *item, *update
	npcRestoreEquipped52ADE0(owner)
	if *owner != beforeOwner || *item != beforeItem || *update != beforeUpdate {
		t.Fatal("restoring an unequipped garment changed its flags or the owner's equip state")
	}
}
