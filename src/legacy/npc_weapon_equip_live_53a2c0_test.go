//go:build amd64 || arm64

package legacy

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/npc_weapon_equip_live_53a2c0.json
var npcWeaponEquipLiveOracle53A2C0 []byte

// The frozen rows were produced by executing unchanged GAME.EXE 0053A420,
// 0053A2C0, 0053A0F0, 0053A140, 0053A030 and 0053A3D0. Only the type,
// appearance, effect and empty duration-list services were bounded.
// This is callback-order evidence, not a whole original combat/world oracle.
func TestNPCWeaponEquip53A2C0ReloadsSubclassAfterCallbacks(t *testing.T) {
	var oracle struct {
		Scope          string   `json:"scope"`
		OriginalSHA256 string   `json:"original_sha256"`
		Roles          []string `json:"roles"`
		Rows           []struct {
			NPC               bool     `json:"npc"`
			Initial           uint32   `json:"initial"`
			Phase             string   `json:"phase"`
			Mutation          uint32   `json:"mutation"`
			Mutated           bool     `json:"mutated"`
			Result            int      `json:"result"`
			Appearances       []string `json:"appearances"`
			Flags             []uint32 `json:"flags"`
			Subclass          uint32   `json:"subclass"`
			Equipment         uint32   `json:"equipment"`
			CompatibilityRole string   `json:"compatibility_role"`
			CompatibilityWord uint32   `json:"compatibility_word"`
			Animation         uint32   `json:"animation"`
			Effects           uint32   `json:"effects"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(npcWeaponEquipLiveOracle53A2C0, &oracle); err != nil {
		t.Fatal(err)
	}
	if oracle.OriginalSHA256 != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" ||
		oracle.Scope == "" || len(oracle.Rows) != 60 ||
		!reflect.DeepEqual(oracle.Roles, []string{"loose", "quiver", "weapon", "new"}) {
		t.Fatal("frozen original equip oracle provenance or coverage differs")
	}

	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	oldServer, oldLookup, oldFlags := GetServer, objectNPCWeaponEquipFlags, noxflags.GetGame()
	GetServer = func() Server { return &playerAttackLegacyServer538960{srv: srv} }
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameHost)
	t.Cleanup(func() {
		GetServer, objectNPCWeaponEquipFlags = oldServer, oldLookup
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})
	for _, name := range []string{"Apple", "Quiver", "Sword", "Bow"} {
		if err := srv.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
			t.Fatal(err)
		}
	}
	srv.Nox_xxx_equipWeapon_4157C0()

	for _, row := range oracle.Rows {
		t.Run(fmt.Sprintf("npc_%t/type_%d/%s_%d", row.NPC, row.Initial, row.Phase, row.Mutation), func(t *testing.T) {
			owner, freeOwner := alloc.New(server.Object{})
			update, freeUpdate := alloc.New(server.MonsterUpdateData{})
			fire, freeFire := alloc.New(server.ModifierEff{})
			brilliance, freeBrilliance := alloc.New(server.ModifierEff{})
			for _, free := range []func(){freeOwner, freeUpdate, freeFire, freeBrilliance} {
				t.Cleanup(free)
			}
			assertHigh := func(pointer unsafe.Pointer) {
				t.Helper()
				if uintptr(pointer) <= math.MaxUint32 {
					t.Fatalf("C-owned fixture %p must exercise a native high address", pointer)
				}
			}
			for _, pointer := range []unsafe.Pointer{unsafe.Pointer(owner), unsafe.Pointer(update), unsafe.Pointer(fire), unsafe.Pointer(brilliance)} {
				assertHigh(pointer)
			}
			fire.Disengage116 = modifierDisengagePointerNative4DFBB0(1)
			brilliance.Engage112 = modifierEngagePointerNative4DFBB0(8)
			if fire.Disengage116 == nil || brilliance.Engage112 == nil {
				t.Fatal("real C modifier callback is missing")
			}

			owner.ObjClass = object.ClassMonster
			if row.NPC {
				owner.ObjSubClass = object.SubClass(object.MonsterNPC)
			}
			owner.UpdateData, owner.Field110 = unsafe.Pointer(update), 9
			owner.Field37, owner.Field38 = 0x80000015, 0x11223344
			for index := range owner.Field140 {
				owner.Field140[index] = 0xA1000000 | uint32(index<<12) | uint32(index&3)
				if index%2 != 0 {
					owner.Field140[index] |= 0x04000000
				}
			}
			update.WeaponEquipFlags, update.ArmorEquipFlags = 0x102, 0x89abcdef
			update.Field516, update.Field517 = 0x2468ace0, 0x55667788
			update.Field481, update.Field518 = 0xcafebabe, 0x3f400000
			update.Field120_0, update.Field120_1, update.Field120_2, update.Field120_3 = 0x91, 0x82, 0x73, 0x64

			var items [4]*server.Object
			var attrs [4]*server.ModifierInitData
			var beforeItems [4]server.Object
			var beforeAttrs [4]server.ModifierInitData
			names := [4]string{"Apple", "Quiver", "Sword", "Quiver"}
			if row.Initial == 4 {
				names[3] = "Bow"
			} else if row.Initial == 256 {
				names[3] = "Sword"
			}
			initialFlags := [4]uint32{0x20005100, 0x60001100, 0x04003100, 0x80002000}
			subclasses := [4]uint32{0, 2, 256, row.Initial}
			for index := range items {
				item, freeItem := alloc.New(server.Object{})
				attr, freeAttr := alloc.New(server.ModifierInitData{})
				t.Cleanup(freeItem)
				t.Cleanup(freeAttr)
				assertHigh(unsafe.Pointer(item))
				assertHigh(unsafe.Pointer(attr))
				items[index], attrs[index] = item, attr
				item.TypeInd = uint16(srv.Types.IndByID(names[index]))
				if item.TypeInd == 0 {
					t.Fatalf("missing registered %s type", names[index])
				}
				item.ObjClass = object.ClassWeapon
				if index == 0 {
					item.ObjClass = object.ClassFood
				}
				item.ObjSubClass, item.ObjFlags = object.SubClass(subclasses[index]), object.Flags(initialFlags[index])
				item.InvHolder, item.InitData = owner, unsafe.Pointer(attr)
				item.Field37, attr.Field16 = 0xabc00000|uint32(index), 0xdef00000|uint32(index)
				if index == 1 || index == 2 {
					attr.Modifiers[2] = fire
				} else if index == 3 {
					attr.Modifiers[2] = brilliance
				}
				if index > 0 {
					items[index-1].InvNextItem = item
				}
			}
			owner.InvFirstItem = items[0]
			for index := range items {
				beforeItems[index], beforeAttrs[index] = *items[index], *attrs[index]
			}
			beforeOwner, beforeUpdate := *owner, *update
			beforeFire, beforeBrilliance := *fire, *brilliance
			appearances := make([]string, 0)
			mutated, newLookups := false, 0
			objectNPCWeaponEquipFlags = func(item *server.Object) uint32 {
				role := -1
				for index, candidate := range items {
					if item == candidate {
						role = index
						break
					}
				}
				if role < 1 {
					t.Errorf("unexpected NPC weapon lookup %p", item)
					return 0
				}
				appearance := row.NPC
				if role == 3 {
					newLookups++
					appearance = appearance && newLookups == 1
				}
				if appearance {
					if owner.Field38 != math.MaxUint32 {
						t.Error("appearance callback ran before NeedSync")
					}
					if item.Flags().Has(object.FlagEquipped) != (role == 3) {
						t.Error("appearance callback ran before the item's equipped bit changed")
					}
					appearances = append(appearances, oracle.Roles[role])
					if !mutated && row.Phase == oracle.Roles[role] {
						mutated = true
						items[3].ObjSubClass = object.SubClass(row.Mutation)
						update.Field516 = 0x13579bdf
					}
				}
				return srv.Weapons.Nox_xxx_weaponInventoryEquipFlags_415820(item)
			}

			if got := playerEquipWeaponNativeCall53A420(owner, items[3]); got != row.Result {
				t.Errorf("real public C equip result=%d, original=%d", got, row.Result)
			}
			if mutated != row.Mutated || !reflect.DeepEqual(appearances, row.Appearances) {
				t.Errorf("callback order=%v mutation=%t, original=%v mutation=%t", appearances, mutated, row.Appearances, row.Mutated)
			}
			if newLookups != 1+bool2int(row.NPC) {
				t.Errorf("arriving-item lookups=%d, want appearance followed by the post-engage lookup", newLookups)
			}

			wantOwner, wantUpdate := beforeOwner, beforeUpdate
			wantOwner.Field110 = row.Effects
			if row.NPC {
				wantOwner.Field38 = math.MaxUint32
				for index := range wantOwner.Field140 {
					wantOwner.Field140[index] |= 0x400
					if wantOwner.Field37&(uint32(1)<<index) != 0 {
						wantOwner.Field140[index] |= 0x04000000
					}
				}
			}
			wantUpdate.WeaponEquipFlags = row.Equipment
			wantUpdate.Field516, wantUpdate.Field517 = row.CompatibilityWord, row.Animation
			if row.CompatibilityRole == "new" {
				// Original PE32 stores the pointer here. On native 64-bit the
				// existing inventory links/bits remain authoritative; never truncate.
				wantUpdate.Field516 = 0
			} else if row.CompatibilityRole != "word" {
				t.Fatal("unknown frozen compatibility role")
			}
			if *owner != wantOwner {
				t.Errorf("owner differs from the frozen original effects/sync contract: effects=%#x, original=%#x", owner.Field110, row.Effects)
			}
			if *update != wantUpdate {
				t.Errorf("update differs: equip=%#x word=%#x animation=%#x; original=%#x %#x %#x",
					update.WeaponEquipFlags, update.Field516, update.Field517, wantUpdate.WeaponEquipFlags, wantUpdate.Field516, wantUpdate.Field517)
			}
			if len(row.Flags) != len(items) {
				t.Fatal("frozen inventory coverage differs")
			}
			for index, item := range items {
				wantItem := beforeItems[index]
				wantItem.ObjFlags = object.Flags(row.Flags[index])
				if index == 3 {
					wantItem.ObjSubClass = object.SubClass(row.Subclass)
				}
				if *item != wantItem || *attrs[index] != beforeAttrs[index] {
					t.Errorf("%s differs from original flags/subclass or changed inventory/modifier storage", oracle.Roles[index])
				}
			}
			if *fire != beforeFire || *brilliance != beforeBrilliance {
				t.Error("real C effect callback changed the modifier definition")
			}
		})
	}
}
