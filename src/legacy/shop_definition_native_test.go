package legacy

import (
	"bytes"
	_ "embed"
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"

	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/shop_definition_original_v1.json.gz
var shopDefinitionOriginal []byte

type shopDefinitionRecord struct {
	Type  uint16
	Qty   uint8
	Mark  uint8
	Param uint32
	Mods  [4]*int
}

type shopDefinitionItem struct {
	Type  uint16
	Cls   object.Class
	Sub   object.SubClass
	Byte  uint8
	Guide string
	Mods  [4]*int
}

type shopDefinitionNativeCase struct {
	Input struct {
		Name        string
		Root        string
		Reverse     bool
		Item        *shopDefinitionItem
		Definitions []*shopDefinitionRecord
		NilSession  bool `json:"nil_session"`
	}
	Observed struct {
		Result      int
		Count       uint8
		Definitions []*shopDefinitionRecord
	}
}

func TestShopDefinitionNativeCABI510320(t *testing.T) {
	rows := tradeOfferOriginalRows[shopDefinitionNativeCase](t, shopDefinitionOriginal,
		"28bdcccb52536445c1cbb93041cf2694b067c5e2b4f8db569fc4bb49a2771b46", 282)
	world := server.NewHigh(nil, nil, strman.New())
	t.Cleanup(world.Close)
	world.FreeObjectTypes()
	t.Cleanup(world.FreeObjectTypes)
	for world.Types.Count() < 7 {
		name := fmt.Sprintf("UnusedShopGuide%d", world.Types.Count())
		if err := world.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
			t.Fatal(err)
		}
	}
	for i, name := range []string{"BlueTroll", "Bluetroll", "", "A", "Li", "Lia", "Ab", "Ax"} {
		if err := world.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
			t.Fatal(err)
		}
		if typ := world.Types.ByInd(7 + i); typ == nil || typ.ID() != name {
			t.Fatalf("guide type %d = %v", 7+i, typ)
		}
	}
	game4ConsumerServices(t, game4TradeServices{world: world})
	field, ok := reflect.TypeFor[server.ModifierEff]().FieldByName("ind4")
	if !ok || field.Type.Kind() != reflect.Uint32 {
		t.Fatal("modifier index field")
	}
	for route := range 2 {
		for _, row := range rows {
			t.Run(fmt.Sprintf("route%d-%s", route, row.Input.Name), func(t *testing.T) {
				checkShopDefinitionNativeCase(t, route, row, field.Offset)
			})
		}
	}
}

func writeShopDefinitionRecords(t *testing.T, idata *server.ShopkeeperInitData, records []*shopDefinitionRecord, count uint8) {
	t.Helper()
	idata.Count = count
	for i, record := range records {
		if record == nil {
			continue
		}
		def := &idata.Items[i]
		*def = server.ShopkeeperItemDefinition{TypeInd: uint32(record.Type), Count: record.Qty, Param: record.Param}
		for index, modifier := range record.Mods {
			if modifier != nil {
				slot, ok := server.EncodeShopkeeperModifierID(*modifier)
				if !ok {
					t.Fatalf("invalid fixture modifier ID %d", *modifier)
				}
				def.ModifierSlots[index] = slot
			}
		}
		padding := (*[3]byte)(unsafe.Add(unsafe.Pointer(def), 5))
		*padding = [3]byte{record.Mark ^ 0xaa, record.Mark ^ 0x55, record.Mark}
	}
}

func shopDefinitionFrame(data *server.ShopkeeperInitData) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(data)), unsafe.Sizeof(*data))
}

func checkShopDefinitionNativeCase(t *testing.T, route int, row shopDefinitionNativeCase, modifierIndex uintptr) {
	t.Helper()
	in := row.Input
	stock := game4ConsumerValue(t, server.ShopkeeperInitData{})
	for i := range shopDefinitionFrame(stock) {
		shopDefinitionFrame(stock)[i] = byte(i*7 + 0x73)
	}
	writeShopDefinitionRecords(t, stock, in.Definitions, uint8(len(in.Definitions)))
	expected := game4ConsumerValue(t, server.ShopkeeperInitData{})
	copy(shopDefinitionFrame(expected), shopDefinitionFrame(stock))
	if in.Root == "decrement" {
		writeShopDefinitionRecords(t, expected, row.Observed.Definitions, row.Observed.Count)
	}
	other := game4ConsumerValue(t, server.ShopkeeperInitData{
		Count: 1, Items: [server.ShopkeeperItemDefinitionCount]server.ShopkeeperItemDefinition{
			{TypeInd: 64157, Count: 54},
		},
	})
	otherBefore := bytes.Clone(shopDefinitionFrame(other))
	merchant := game4ConsumerValue(t, server.Object{ObjClass: object.ClassMonster, InitData: unsafe.Pointer(stock)})
	player := game4ConsumerValue(t, server.Object{ObjClass: object.ClassPlayer, InitData: unsafe.Pointer(other)})
	session := game4ConsumerValue(t, server.TradeSession{Field8: merchant, Field12: player})
	if in.Reverse {
		session.Field8, session.Field12 = player, merchant
	}
	beforeSession, beforePlayer, beforeMerchant := *session, *player, *merchant
	var item *server.Object
	var beforeItem server.Object
	if subject := in.Item; subject != nil {
		item = game4ConsumerValue(t, server.Object{TypeInd: subject.Type, ObjClass: subject.Cls, ObjSubClass: subject.Sub})
		attrs := game4ConsumerValue(t, server.ModifierInitData{})
		for i, id := range subject.Mods {
			if id != nil {
				effect := game4ConsumerValue(t, server.ModifierEff{})
				*(*uint32)(unsafe.Add(unsafe.Pointer(effect), modifierIndex)) = uint32(*id)
				attrs.Modifiers[i] = effect
			}
		}
		item.InitData = unsafe.Pointer(attrs)
		if subject.Sub&5 != 0 {
			item.UseData.Ptr = unsafe.Pointer(game4ConsumerValue(t, server.SpellRewardUseData{Spell: subject.Byte}))
		} else {
			guide := game4ConsumerValue(t, server.FieldGuideUseData{})
			guide.SetCreature(subject.Guide)
			item.UseData.Ptr = unsafe.Pointer(guide)
		}
		beforeItem = *item
	}
	var result int
	switch in.Root {
	case "match":
		var def *server.ShopkeeperItemDefinition
		if in.Definitions[0] != nil {
			def = &stock.Items[0]
		}
		result = callShopDefinitionMatch5103F0(route, item, def)
	case "index":
		result = callShopDefinitionIndex5103A0(route, session, item)
	case "decrement":
		arg := session
		if in.NilSession {
			arg = nil
		}
		callShopDefinitionDecrement510320(route, arg, item)
	default:
		t.Fatal("unknown stock entry", in.Root)
	}
	if result != row.Observed.Result {
		t.Fatalf("result = %d, original %d", result, row.Observed.Result)
	}
	for i, word := range shopDefinitionFrame(stock) {
		if want := shopDefinitionFrame(expected)[i]; word != want {
			t.Fatalf("stock owner offset %#x: byte %#02x, original %#02x", i, word, want)
		}
	}
	if !bytes.Equal(otherBefore, shopDefinitionFrame(other)) || *session != beforeSession ||
		*player != beforePlayer || *merchant != beforeMerchant || (item != nil && *item != beforeItem) {
		t.Fatal("entry changed another owner")
	}
}
