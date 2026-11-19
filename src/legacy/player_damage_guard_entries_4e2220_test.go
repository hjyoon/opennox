package legacy

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/original_electric_armor_and_guard_v2.json.gz
var originalPlayerGuard4E2220 []byte

type playerGuardStockInput struct {
	Entry string
	Name  string
	Chain []struct {
		Class       uint32 `json:"cls"`
		Subclass    uint32
		Flags       uint32
		Material    uint16
		Material7   int
		Coefficient uint32
	}
	Special         uint32
	TargetClass     uint32 `json:"target_class"`
	Mask            uint32
	HasSource       bool `json:"has_source"`
	HasWeapon       bool `json:"has_weapon"`
	Amount          uint32
	Kind            int32
	PostItemFlags   *uint32 `json:"post_item_flags"`
	PostTargetClass *uint32 `json:"post_target_class"`
}

type playerGuardStockCall struct {
	Service   string
	Item      string
	Target    string
	Source    string
	Effective string
	Amount    uint32
	Kind      int32
	State     uint32
}

func TestPlayerDamageGuardFrozenOriginal4E2220(t *testing.T) {
	z, err := gzip.NewReader(bytes.NewReader(originalPlayerGuard4E2220))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var stock struct {
		Rows []struct {
			Input    playerGuardStockInput
			Observed struct {
				Value uint32
				Calls []json.RawMessage
			}
		}
		Faults []json.RawMessage
	}
	if err = json.NewDecoder(z).Decode(&stock); err != nil {
		t.Fatal(err)
	}
	if len(stock.Rows) != 514 || len(stock.Faults) != 131 {
		t.Fatalf("unexpected original contexts %d / %d", len(stock.Rows), len(stock.Faults))
	}
	materialName, freeMaterial := alloc.CStringHigh("Material7")
	t.Cleanup(freeMaterial)
	material := server.ModifierEff{}
	// The native ModifierEff record begins with its C name pointer.
	*(**byte)(material.C()) = materialName
	if material.Name() != "Material7" {
		t.Fatal("material fixture does not use the native name slot")
	}
	for _, row := range stock.Rows {
		in := row.Input
		t.Run(in.Name, func(t *testing.T) {
			var target server.Object
			items := make([]server.Object, len(in.Chain))
			initData := make([]server.ModifierInitData, len(in.Chain))
			target.ObjClass = object.Class(in.TargetClass)
			if len(items) != 0 {
				target.InvFirstItem = &items[0]
			}
			for i, value := range in.Chain {
				item := &items[i]
				item.TypeInd = uint16(i)
				item.ObjClass = object.Class(value.Class)
				item.ObjSubClass = object.SubClass(value.Subclass)
				item.ObjFlags = object.Flags(value.Flags)
				item.Material = value.Material
				if value.Material7 != 0 {
					initData[i].Modifiers[1] = &material
					item.InitData = unsafe.Pointer(&initData[i])
				}
				if i+1 < len(items) {
					item.InvNextItem = &items[i+1]
				}
			}
			if in.Entry == "electric-armor-scale" {
				var want []int
				for _, raw := range row.Observed.Calls {
					var call struct {
						Service string
						Item    int
					}
					if err := json.Unmarshal(raw, &call); err != nil {
						t.Fatal(err)
					}
					if call.Service == "coefficient" {
						want = append(want, call.Item)
					}
				}
				var lookups []int
				result := server.PlayerDamageElectricArmorScale4E2220(&target, func() float32 {
					return math.Float32frombits(in.Special)
				}, server.ItemHasMaterial7Modifier4133D0, func(item *server.Object) float32 {
					i := int(item.TypeInd)
					lookups = append(lookups, i)
					return math.Float32frombits(in.Chain[i].Coefficient)
				})
				if bits := math.Float32bits(result); bits != row.Observed.Value || !slices.Equal(lookups, want) {
					t.Fatalf("coefficient %08x and lookups %v, want %08x and %v", bits, lookups, row.Observed.Value, want)
				}
				return
			}
			var source, weapon *server.Object
			if in.HasSource {
				source = &server.Object{}
			}
			if in.HasWeapon {
				weapon = &server.Object{}
			}
			names := map[*server.Object]string{nil: "nil", &target: "target"}
			if source != nil {
				names[source] = "source"
			}
			if weapon != nil {
				names[weapon] = "weapon"
			}
			for i, label := range []string{"item0", "item1", "item2"} {
				names[&items[i]] = label
			}
			armorEntry := in.Entry == "armor-block-wear"
			service := "weapon"
			if armorEntry {
				service = "armor"
			}
			var calls []playerGuardStockCall
			got := playerBlockWear4E22A0(&target, source, weapon, in.Mask, math.Float32frombits(in.Amount),
				object.DamageType(in.Kind), armorEntry, playerBlockWearRuntime4E22A0{
					damage: func(item, owner, src, effective *server.Object, amount float32, typ object.DamageType) bool {
						calls = append(calls, playerGuardStockCall{Service: service, Item: names[item],
							Target: names[owner], Source: names[src], Effective: names[effective],
							Amount: math.Float32bits(amount), Kind: int32(typ)})
						if in.PostItemFlags != nil {
							item.ObjFlags = object.Flags(*in.PostItemFlags)
						}
						if in.PostTargetClass != nil {
							owner.ObjClass = object.Class(*in.PostTargetClass)
						}
						return false
					},
					setPlayerState: func(unit *server.Object, state server.PlayerState) bool {
						if unit != &target {
							t.Fatal("state target changed")
						}
						calls = append(calls, playerGuardStockCall{Service: "state", State: uint32(state)})
						return true
					},
					popMonsterState: func(unit *server.Object) {
						if unit != &target {
							t.Fatal("action target changed")
						}
						calls = append(calls, playerGuardStockCall{Service: "pop"})
					},
				})
			if got != (row.Observed.Value != 0) || len(calls) != len(row.Observed.Calls) {
				t.Fatalf("state response %v and %v, want %d and %d calls", got, calls, row.Observed.Value, len(row.Observed.Calls))
			}
			for i, raw := range row.Observed.Calls {
				var want playerGuardStockCall
				if err := json.Unmarshal(raw, &want); err != nil {
					t.Fatal(err)
				}
				if calls[i] != want {
					t.Fatalf("call %d = %+v, want %+v", i, calls[i], want)
				}
			}
		})
	}
}

func TestPlayerDamageGuardPublicCEntries4E2220(t *testing.T) {
	InitBlobData()
	srv := npcReflectLegacyServer4E17B0(t)
	special := memmap.PtrFloat32(0x587000, 201108)
	previous := *special
	*special = -0.25
	t.Cleanup(func() { *special = previous })
	target, freeTarget := alloc.NewHigh(server.Object{})
	item, freeItem := alloc.NewHigh(server.Object{})
	*item = server.Object{
		ObjClass: object.ClassArmor, ObjSubClass: 0x02000002, ObjFlags: object.FlagEquipped, Material: 0x10,
	}
	source, freeSource := alloc.NewHigh(server.Object{})
	weapon, freeWeapon := alloc.NewHigh(server.Object{})
	attrs, freeAttrs := alloc.NewHigh(server.ModifierInitData{})
	cache, freeCache := alloc.NewHigh(server.ModifierEff{})
	sameName, freeSameName := alloc.NewHigh(server.ModifierEff{})
	name, freeName := alloc.CStringHigh("Material7")
	t.Cleanup(freeTarget)
	t.Cleanup(freeItem)
	t.Cleanup(freeSource)
	t.Cleanup(freeWeapon)
	t.Cleanup(freeAttrs)
	t.Cleanup(freeCache)
	t.Cleanup(freeSameName)
	t.Cleanup(freeName)
	*(**byte)(cache.C()) = name
	*(**byte)(sameName.C()) = name
	cacheCell := srv.Modif.Material7CacheCell4133D0()
	*cacheCell = cache
	t.Cleanup(func() { *cacheCell = nil })
	attrs.Modifiers[1] = sameName
	item.InitData = unsafe.Pointer(attrs)
	for _, ptr := range []*server.Object{target, item, source, weapon} {
		if cgoABIPointerSize == 8 && uintptr(unsafe.Pointer(ptr)) <= math.MaxUint32 {
			t.Fatalf("fixture pointer %p must be above 4 GiB", ptr)
		}
	}
	target.InvFirstItem = item
	before := *item
	scale, blockedWeapon, blockedArmor := playerDamageGuardPublicCalls4E2220(target, source, weapon)
	if scale != 0.75 || blockedWeapon != 0 || blockedArmor != 0 || *item != before {
		t.Fatalf("public C guard result = %g/%d/%d, item mutated=%t", scale, blockedWeapon, blockedArmor, *item != before)
	}
	if scale := playerDamageElectricArmorScale4E2220(srv, target); scale != 1 {
		t.Fatalf("ordinary armor material-name path returned %g, want 1", scale)
	}
	attrs.Modifiers[1] = cache
	if scale, _, _ := playerDamageGuardPublicCalls4E2220(target, source, weapon); scale != 1 {
		t.Fatalf("raw armor material identity path returned %g, want 1", scale)
	}
	attrs.Modifiers[1] = sameName
	item.ObjSubClass, item.TypeInd = 2, 75
	def := &server.Modifier{DamageCoeffOrArmor64: -0.125}
	lookups := 0
	restore := pushArmorCoefficientObserver415BD0(&armorCoefficientObserver415BD0{Lookup: func(index uint16) unsafe.Pointer {
		if index != 75 {
			t.Fatalf("raw armor coefficient received type %d, want 75", index)
		}
		lookups++
		return def.C()
	}})
	defer restore()
	if scale, _, _ := playerDamageGuardPublicCalls4E2220(target, source, weapon); scale != 0.875 || lookups != 1 {
		t.Fatalf("raw armor coefficient result %g with %d lookups", scale, lookups)
	}
}
