package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"slices"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func rewardDefinitionsPutUint32Test4F0640(out *bytes.Buffer, value uint32) {
	var data [4]byte
	binary.LittleEndian.PutUint32(data[:], value)
	out.Write(data[:])
}

func rewardDefinitionsPutStringTest4F0640(out *bytes.Buffer, value string) {
	rewardDefinitionsPutUint32Test4F0640(out, uint32(len(value)))
	out.WriteString(value)
}

func rewardDefinitionsSemanticBytesTest4F0640(tables *rewardDefinitionTables4F0640) []byte {
	var out bytes.Buffer
	for _, definition := range tables.Objects {
		out.WriteByte('O')
		rewardDefinitionsPutUint32Test4F0640(&out, definition.Weight)
		rewardDefinitionsPutStringTest4F0640(&out, definition.Name)
		rewardDefinitionsPutUint32Test4F0640(&out, definition.Kind)
		rewardDefinitionsPutUint32Test4F0640(&out, definition.Slots)
	}
	groups := [][]rewardModifierDefinition4F0640{
		tables.WeaponPower[:],
		tables.ArmorQuality[:],
		tables.Material[:],
		tables.Enchantments[:],
	}
	for groupIndex, group := range groups {
		for _, definition := range group {
			out.WriteByte('M')
			out.WriteByte(byte(groupIndex))
			rewardDefinitionsPutUint32Test4F0640(&out, definition.Group)
			rewardDefinitionsPutStringTest4F0640(&out, definition.Name)
			rewardDefinitionsPutUint32Test4F0640(&out, definition.Slots)
			rewardDefinitionsPutUint32Test4F0640(&out, definition.ExcludeArmor)
			rewardDefinitionsPutUint32Test4F0640(&out, definition.ExcludeWeapon)
		}
	}
	return out.Bytes()
}

func rewardModifierGroupsTest4F0640(tables *rewardDefinitionTables4F0640) [][]rewardModifierDefinition4F0640 {
	return [][]rewardModifierDefinition4F0640{
		tables.WeaponPower[:],
		tables.ArmorQuality[:],
		tables.Material[:],
		tables.Enchantments[:],
	}
}

func TestRewardDefinitionsStaticDataMatchesGAMEEXE4F0640(t *testing.T) {
	var tables rewardDefinitionTables4F0640
	tables.init()
	data := rewardDefinitionsSemanticBytesTest4F0640(&tables)
	if len(data) != 4051 {
		t.Fatalf("semantic data size = %d, want 4051", len(data))
	}
	const want = "b383d4387526223621f56da81121555b4952aa62a4c281077b32f2d0ce452878"
	if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != want {
		t.Fatalf("semantic data SHA-256 = %s, want %s", got, want)
	}
	if unsafe.Sizeof(tables.Objects[0].Weight) != 4 ||
		unsafe.Sizeof(tables.Objects[0].TypeInd) != 4 ||
		unsafe.Sizeof(tables.Objects[0].Kind) != 4 ||
		unsafe.Sizeof(tables.Objects[0].Slots) != 4 {
		t.Fatal("reward object numeric fields must remain exact uint32")
	}
	if unsafe.Sizeof(tables.WeaponPower[0].Modifier) != unsafe.Sizeof((*ModifierEff)(nil)) {
		t.Fatal("resolved modifier must retain the native pointer width")
	}
}

func TestRewardDefinitionsInitOrderAndStores4F0640(t *testing.T) {
	var tables rewardDefinitionTables4F0640
	tables.init()
	objectSentinel := uint32(0xdeadbeef)
	tables.Objects[len(tables.Objects)-1].TypeInd = objectSentinel
	modifierSentinels := [4]ModifierEff{}
	groups := rewardModifierGroupsTest4F0640(&tables)
	for index, group := range groups {
		group[len(group)-1].Modifier = &modifierSentinels[index]
	}

	var events []string
	objectCalls := 0
	modifierIDCalls := 0
	modifierDescCalls := 0
	modifiers := make([]ModifierEff, 71)
	rewardDefinitionsInit4F0640(
		&tables,
		func(name string) int32 {
			events = append(events, "object:"+name)
			value := int32(0x1000 + objectCalls)
			objectCalls++
			return value
		},
		func(name string) int32 {
			events = append(events, "modifier-id:"+name)
			value := int32(0x2000 + modifierIDCalls)
			modifierIDCalls++
			return value
		},
		func(id int32) *ModifierEff {
			events = append(events, fmt.Sprintf("modifier-desc:%d", id))
			want := int32(0x2000 + modifierDescCalls)
			if id != want {
				t.Fatalf("modifier descriptor ID = %d, want %d", id, want)
			}
			result := &modifiers[modifierDescCalls]
			modifierDescCalls++
			return result
		},
	)

	if objectCalls != 57 || modifierIDCalls != 71 || modifierDescCalls != 71 {
		t.Fatalf("lookup counts = object %d, modifier ID %d, descriptor %d; want 57/71/71",
			objectCalls, modifierIDCalls, modifierDescCalls)
	}
	var wantEvents []string
	for _, definition := range tables.Objects {
		if definition.Name == "" {
			break
		}
		wantEvents = append(wantEvents, "object:"+strings.TrimPrefix(definition.Name, "#"))
	}
	modifierIndex := 0
	for _, group := range groups {
		for _, definition := range group {
			if definition.Name == "" {
				break
			}
			wantEvents = append(wantEvents,
				"modifier-id:"+definition.Name,
				fmt.Sprintf("modifier-desc:%d", 0x2000+modifierIndex),
			)
			modifierIndex++
		}
	}
	if !slices.Equal(events, wantEvents) {
		t.Fatalf("events differ:\n got: %q\nwant: %q", events, wantEvents)
	}
	for index := range 57 {
		if got, want := tables.Objects[index].TypeInd, uint32(0x1000+index); got != want {
			t.Fatalf("object row %d type ID = %#x, want %#x", index, got, want)
		}
	}
	if tables.Objects[len(tables.Objects)-1].TypeInd != objectSentinel {
		t.Fatal("object sentinel output was modified")
	}
	modifierIndex = 0
	for groupIndex, group := range groups {
		for rowIndex := 0; rowIndex < len(group)-1; rowIndex++ {
			if got, want := group[rowIndex].Modifier, &modifiers[modifierIndex]; got != want {
				t.Fatalf("modifier group %d row %d = %p, want %p", groupIndex, rowIndex, got, want)
			}
			modifierIndex++
		}
		if group[len(group)-1].Modifier != &modifierSentinels[groupIndex] {
			t.Fatalf("modifier group %d sentinel output was modified", groupIndex)
		}
	}
}

func TestRewardDefinitionsInitReadsNextRowsLive4F0640(t *testing.T) {
	var tables rewardDefinitionTables4F0640
	tables.Objects[0].Name = "#One"
	tables.Objects[1].Name = "Two"
	tables.WeaponPower[0].Name = "WeaponOne"
	tables.WeaponPower[1].Name = "WeaponTwo"
	tables.ArmorQuality[0].Name = "ArmorOne"
	tables.Enchantments[0].Name = "EnchantOne"

	var events []string
	rewardDefinitionsInit4F0640(
		&tables,
		func(name string) int32 {
			events = append(events, "object:"+name)
			tables.Objects[1].Name = ""
			return -1
		},
		func(name string) int32 {
			events = append(events, "modifier-id:"+name)
			if name == "WeaponOne" {
				tables.WeaponPower[1].Name = ""
			}
			return 7
		},
		func(id int32) *ModifierEff {
			events = append(events, fmt.Sprintf("modifier-desc:%d", id))
			if len(events) == 5 {
				tables.Enchantments[0].Name = ""
			}
			return nil
		},
	)
	want := []string{
		"object:One",
		"modifier-id:WeaponOne", "modifier-desc:7",
		"modifier-id:ArmorOne", "modifier-desc:7",
	}
	if !slices.Equal(events, want) {
		t.Fatalf("live-row events = %q, want %q", events, want)
	}
	if tables.Objects[0].TypeInd != ^uint32(0) {
		t.Fatalf("signed object result bits = %#x, want %#x", tables.Objects[0].TypeInd, ^uint32(0))
	}
}

func TestServerRewardDefinitionsInitUsesNativeServices4F0640(t *testing.T) {
	server := &Server{}
	server.Types.byID = map[string]*ObjectType{
		"fanchakram": {id: "FanChakram", ind: 37},
		"forcewand":  {id: "ForceWand", ind: 91},
	}
	name, freeName := alloc.CString("WeaponPower2")
	t.Cleanup(freeName)
	modifier := &ModifierEff{name0: name, ind4: 44}
	server.Modif.types[0] = modifier

	server.RewardDefinitionsInit4F0640()
	if !server.rewardDefinitions.initialized {
		t.Fatal("native reward definitions were not initialized")
	}
	if got := server.rewardDefinitions.Objects[0].TypeInd; got != 37 {
		t.Fatalf("FanChakram type ID = %d, want 37", got)
	}
	if got := server.rewardDefinitions.Objects[8].TypeInd; got != 91 {
		t.Fatalf("ForceWand type ID = %d, want 91", got)
	}
	if got := server.rewardDefinitions.Objects[17].TypeInd; got != 91 {
		t.Fatalf("#ForceWand stripped type ID = %d, want 91", got)
	}
	if got := server.rewardDefinitions.WeaponPower[0].Modifier; got != modifier {
		t.Fatalf("WeaponPower2 modifier = %p, want %p", got, modifier)
	}
}
