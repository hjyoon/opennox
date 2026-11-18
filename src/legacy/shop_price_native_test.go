package legacy

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/shop_price_original_v1.json.gz
var shopPriceOriginal []byte

type shopPriceNativeCase struct {
	Input struct {
		Name           string
		Mode           int32
		Session        string
		Quest          bool
		Class          object.Class
		Subclass       object.SubClass
		Type           uint16
		Worth          uint32
		Health         []uint16
		Modifiers      [4]*int32
		Ammo           [3]uint8
		Wand           [2]uint8
		Used           bool
		GuideWorth     *uint32 `json:"guide_worth"`
		BuyMultiplier  float32 `json:"buy_multiplier"`
		SellMultiplier float32 `json:"sell_multiplier"`
		Balance        map[string]float32
	}
	Observed struct{ Cost int32 }
}

func TestShopPriceNativeCABI50E3D0(t *testing.T) {
	rows := tradeOfferOriginalRows[shopPriceNativeCase](t, shopPriceOriginal,
		"afb5af96ba0cbff7731eab1d72fb4baa27888fbdde42f2d83029c7f07d746c41", 576)
	world, gems := newShopPriceWorld50E3D0(t, rows[0].Input.Balance)
	game4ConsumerServices(t, game4TradeServices{world: world})
	for route := range 2 {
		for _, row := range rows {
			t.Run(fmt.Sprintf("route%d-%s", route, row.Input.Name), func(t *testing.T) {
				checkShopPriceNativeCase(t, route, row, gems)
			})
		}
	}
}

func newShopPriceWorld50E3D0(t *testing.T, balance map[string]float32) (*server.Server, map[uint16]uint16) {
	t.Helper()
	oldData, flags := datapath.Data(""), noxflags.GetGame()
	t.Cleanup(func() {
		datapath.SetData(oldData)
		noxflags.ResetGame()
		noxflags.SetGame(flags)
	})
	noxflags.ResetGame()
	path := t.TempDir()
	datapath.SetData(path)
	keys := make([]string, 0, len(balance))
	for key := range balance {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var source strings.Builder
	for _, key := range keys {
		fmt.Fprintf(&source, "%s: %.9g\n", key, balance[key])
	}
	source.WriteString("QuestSpellWorthMultiplier: 1\n")
	if err := os.WriteFile(filepath.Join(path, "gamedata.yml"), []byte(source.String()), 0600); err != nil {
		t.Fatal(err)
	}
	const spellID = spell.ID(11)
	spellYAML := fmt.Sprintf("- name: %s\n  price: 74568\n", spellID.String())
	if err := os.WriteFile(filepath.Join(path, "spells.yml"), []byte(spellYAML), 0600); err != nil {
		t.Fatal(err)
	}
	world := server.NewHigh(nil, nil, strman.New())
	t.Cleanup(world.Close)
	world.FreeObjectTypes()
	t.Cleanup(world.FreeObjectTypes)
	if err := world.Balance.Read(); err != nil {
		t.Fatal(err)
	}
	mem, _ := alloc.Make([]byte{}, 4)
	file := binfile.NewMemFile(unsafe.Pointer(&mem[0]), len(mem))
	defer file.Free()
	if err := world.Spells.Read(file, nil, false); err != nil {
		t.Fatal(err)
	}
	if got := world.Spells.Price(spellID); got != 74568 {
		t.Fatalf("fixture spell price = %d", got)
	}
	gems := make(map[uint16]uint16)
	for i, name := range []string{"Diamond", "Ruby", "Emerald"} {
		if err := world.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
			t.Fatal(err)
		}
		gems[uint16(715+i)] = uint16(world.Types.IndByID(name))
	}
	guideWorth := 53
	if err := world.Types.ReadObjectType(&things.Thing{Name: "FixtureGuideSpecies", Price: &guideWorth}); err != nil {
		t.Fatal(err)
	}
	return world, gems
}

func checkShopPriceNativeCase(t *testing.T, route int, row shopPriceNativeCase, gems map[uint16]uint16) {
	t.Helper()
	in := row.Input
	noxflags.UnsetGame(noxflags.GameModeQuest)
	if in.Quest {
		noxflags.SetGame(noxflags.GameModeQuest)
	}
	typ := in.Type
	if remapped, ok := gems[typ]; ok {
		typ = remapped
	}
	item := game4ConsumerValue(t, server.Object{
		TypeInd: typ, ObjClass: in.Class, ObjSubClass: in.Subclass, Worth: in.Worth,
	})
	if in.Health != nil {
		item.HealthData = game4ConsumerValue(t, server.HealthData{Cur: in.Health[0], Max: in.Health[1]})
	}
	initData := game4ConsumerValue(t, server.ModifierInitData{})
	for i, mod := range in.Modifiers {
		if mod != nil {
			initData.Modifiers[i] = game4ConsumerValue(t, server.ModifierEff{Price20: *mod})
		}
	}
	item.InitData = unsafe.Pointer(initData)
	if in.Used {
		item.UpdateData = unsafe.Pointer(game4ConsumerValue(t, server.WeaponArmorUpdateData{Field4: 1}))
	}
	switch {
	case in.Class.Has(object.ClassInfoBook) && in.Subclass&1 != 0:
		item.UseData.Ptr = unsafe.Pointer(game4ConsumerValue(t, server.SpellRewardUseData{Spell: 11}))
	case in.Class.Has(object.ClassInfoBook) && in.Subclass&2 != 0:
		guide := game4ConsumerValue(t, server.FieldGuideUseData{})
		creature := "MissingGuideSpecies"
		if in.GuideWorth != nil {
			creature = "FixtureGuideSpecies"
		}
		guide.SetCreature(creature)
		item.UseData.Ptr = unsafe.Pointer(guide)
	case in.Class.Has(object.ClassWeapon) && in.Subclass&0x82 != 0:
		item.UseData.Ptr = unsafe.Pointer(game4ConsumerValue(t, server.AmmoUseData{
			Charge0: in.Ammo[0], Charge1: in.Ammo[1], Field2: in.Ammo[2],
		}))
	case in.Class.Has(object.ClassWand):
		item.UseData.Ptr = unsafe.Pointer(game4ConsumerValue(t, server.WandUseData{
			Charge: in.Wand[0], MaxCharge: in.Wand[1],
		}))
	}
	var session *server.TradeSession
	if in.Session != "nil" {
		player := game4ConsumerValue(t, server.Object{ObjClass: object.ClassPlayer})
		merchant := game4ConsumerValue(t, server.Object{ObjClass: object.ClassMonster})
		merchant.InitData = unsafe.Pointer(game4ConsumerValue(t, server.ShopkeeperInitData{
			BuyMultiplier: in.BuyMultiplier, SellMultiplier: in.SellMultiplier,
		}))
		session = game4ConsumerValue(t, server.TradeSession{Field8: player, Field12: merchant})
		if in.Session != "empty" {
			session.Field16 = 1
		}
		if in.Session == "first_shop" {
			session.Field8, session.Field12 = merchant, player
		}
	}
	before := *item
	if got := callShopPriceNative50E3D0(route, in.Mode, session, item); got != row.Observed.Cost {
		t.Fatalf("price = %d, original %d", got, row.Observed.Cost)
	}
	if *item != before {
		t.Fatal("pricing changed the object")
	}
}
