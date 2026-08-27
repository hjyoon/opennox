package server

import (
	"math"
	"runtime"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func questItemServerAddType4F2590(srv *Server, name string, typeInd int) {
	if srv.Types.byID == nil {
		srv.Types.byID = make(map[string]*ObjectType)
	}
	if len(srv.Types.byInd) <= typeInd {
		srv.Types.byInd = append(srv.Types.byInd, make([]*ObjectType, typeInd-len(srv.Types.byInd)+1)...)
	}
	typ := &ObjectType{s: &srv.Types, id: name, ind: uint16(typeInd)}
	srv.Types.byID[strings.ToLower(name)] = typ
	srv.Types.byInd[typeInd] = typ
}

func questItemServerAddModifier4F2590(
	srv *Server,
	storage *[][]byte,
	name string,
	id uint32,
	allowWeapons, allowArmor, allowPos uint32,
) *ModifierEff {
	bytes := append([]byte(name), 0)
	*storage = append(*storage, bytes)
	modifier := &ModifierEff{
		name0:          &bytes[0],
		ind4:           id,
		AllowWeapons28: allowWeapons,
		AllowArmor32:   allowArmor,
		AllowPos36:     allowPos,
	}
	if head := srv.Modif.types[0]; head == nil {
		srv.Modif.types[0] = modifier
	} else {
		for head.next136 != nil {
			head = head.next136
		}
		head.next136 = modifier
		modifier.prev140 = head
	}
	return modifier
}

func questItemServerAddCachedTypes4F2590(srv *Server) {
	for index, name := range questItemTypeNames4F2590 {
		questItemServerAddType4F2590(srv, name, 101+index)
	}
}

func TestQuestItemNativeLayouts4F2590(t *testing.T) {
	wantObjectClass := uintptr(8)
	wantObjectSubclass := uintptr(12)
	wantObjectInitData := uintptr(692)
	wantObjectUseData := uintptr(736)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectClass = 12
		wantObjectSubclass = 16
		wantObjectInitData = 760
		wantObjectUseData = 848
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.TypeInd width", unsafe.Sizeof(Object{}.TypeInd), 2},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantObjectClass},
		{"Object.ObjSubClass", unsafe.Offsetof(Object{}.ObjSubClass), wantObjectSubclass},
		{"Object.InitData", unsafe.Offsetof(Object{}.InitData), wantObjectInitData},
		{"Object.UseData", unsafe.Offsetof(Object{}.UseData), wantObjectUseData},
		{"ModifierInitData.Modifiers", unsafe.Offsetof(ModifierInitData{}.Modifiers), 0},
		{"cache modifier pointer", unsafe.Sizeof(questItemEligibilityCache4F2590[*ModifierEff]{}.defaultReplenishment), unsafe.Sizeof(uintptr(0))},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestQuestItemServerUsesNativeObjectAndModifierPointers4F2590(t *testing.T) {
	srv := new(Server)
	questItemServerAddCachedTypes4F2590(srv)
	var names [][]byte
	replenishments := [4]*ModifierEff{}
	for index, name := range questItemReplenishmentNames4F2590 {
		replenishments[index] = questItemServerAddModifier4F2590(srv, &names, name, uint32(index+1), 0x40, 0x40, 3)
	}
	power := questItemServerAddModifier4F2590(srv, &names, "Power", 10, 0x40, 0, 0)
	material := questItemServerAddModifier4F2590(srv, &names, "Material", 11, 0x40, 0, 0)
	first := questItemServerAddModifier4F2590(srv, &names, "First", 12, 0x40, 0, 1)
	second := questItemServerAddModifier4F2590(srv, &names, "Second", 13, 0x40, 0, 2)

	srv.rewardDefinitions.Objects = [58]rewardObjectDefinition4F0640{
		{Name: "Weapon", TypeInd: 200},
		{},
	}
	srv.rewardDefinitions.WeaponPower = [6]rewardModifierDefinition4F0640{
		{Name: "Power", Modifier: power},
		{},
	}
	srv.rewardDefinitions.Material = [6]rewardModifierDefinition4F0640{
		{Name: "Material", Modifier: material},
		{},
	}
	srv.rewardDefinitions.Enchantments = [57]rewardModifierDefinition4F0640{
		{Name: "First", Modifier: first},
		{Name: "Second", Modifier: second},
		{},
	}
	srv.Weapons.table[0] = weaponRecord{Name: "Weapon", TypeInd: 200, Bit: 0x40}
	data := &ModifierInitData{Modifiers: [4]*ModifierEff{power, material, first, second}}
	item := &Object{
		TypeInd:  200,
		ObjClass: object.ClassWeapon,
		InitData: unsafe.Pointer(data),
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		pointers := []unsafe.Pointer{
			unsafe.Pointer(item), unsafe.Pointer(data), unsafe.Pointer(power), unsafe.Pointer(material),
			unsafe.Pointer(first), unsafe.Pointer(second), unsafe.Pointer(replenishments[0]),
		}
		for index, pointer := range pointers {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("pointer %d does not exercise native high address: %p", index, pointer)
			}
		}
	}
	if got := srv.QuestItemEligible4F2590(item); got != 1 {
		t.Fatalf("native weapon eligibility = %d, want 1", got)
	}
	if srv.questItemEligibility.replenishments != replenishments {
		t.Fatalf("cached native modifier identities = %#v, want %#v", srv.questItemEligibility.replenishments, replenishments)
	}
	if data.Modifiers != [4]*ModifierEff{power, material, first, second} {
		t.Fatalf("native modifier data changed = %#v", data.Modifiers)
	}
	runtime.KeepAlive(names)
}

func TestQuestItemServerInfoBookUseData4F2700(t *testing.T) {
	srv := new(Server)
	questItemServerAddCachedTypes4F2590(srv)

	spellUse := &SpellRewardUseData{Spell: uint8(rewardSpellDefinitions4F09F0[0].SpellID)}
	spell := &Object{
		ObjClass:    object.ClassInfoBook,
		ObjSubClass: object.SubClass(object.BookSpell),
		UseData:     UseDataPtr{Ptr: unsafe.Pointer(spellUse)},
	}
	if got := srv.QuestItemEligible4F2590(spell); got != 1 {
		t.Fatalf("native spell book = %d, want 1", got)
	}

	guideUse := &FieldGuideUseData{}
	guideUse.SetCreature("Urchin")
	guide := &Object{
		ObjClass:    object.ClassInfoBook,
		ObjSubClass: object.SubClass(object.BookFieldGuide),
		UseData:     UseDataPtr{Ptr: unsafe.Pointer(guideUse)},
	}
	if got := srv.QuestItemEligible4F2590(guide); got != 1 {
		t.Fatalf("native field guide = %d, want 1", got)
	}

	abilityUse := &AbilityRewardUseData{Ability: 5}
	ability := &Object{
		ObjClass:    object.ClassInfoBook,
		ObjSubClass: object.SubClass(object.BookAbility),
		UseData:     UseDataPtr{Ptr: unsafe.Pointer(abilityUse)},
	}
	if got := srv.QuestItemEligible4F2590(ability); got != 1 {
		t.Fatalf("native ability book = %d, want 1", got)
	}
	abilityUse.Ability = 6
	if got := srv.QuestItemEligible4F2590(ability); got != 0 {
		t.Fatalf("out-of-range native ability book = %d, want 0", got)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for index, pointer := range []unsafe.Pointer{unsafe.Pointer(spellUse), unsafe.Pointer(guideUse), unsafe.Pointer(abilityUse)} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("use-data pointer %d does not exercise native high address: %p", index, pointer)
			}
		}
	}
}

func TestQuestItemServerDefaultStreetColorModifier4F2B60(t *testing.T) {
	srv := new(Server)
	questItemServerAddCachedTypes4F2590(srv)
	var names [][]byte
	replenishment := questItemServerAddModifier4F2590(srv, &names, "Replenishment1", 1, 0, 0, 3)
	for index := 1; index < len(questItemReplenishmentNames4F2590); index++ {
		questItemServerAddModifier4F2590(srv, &names, questItemReplenishmentNames4F2590[index], uint32(index+1), 0, 0, 3)
	}
	color := questItemServerAddModifier4F2590(srv, &names, "uSeRcOlO1", 10, 0, 0x405, 0)
	srv.Armor.table[0] = armorRecord{Name: "StreetSneakers", TypeInd: 105, Bit: 0x405}
	data := &ModifierInitData{Modifiers: [4]*ModifierEff{color, nil, nil, nil}}
	item := &Object{
		TypeInd:  105,
		ObjClass: object.ClassArmor,
		InitData: unsafe.Pointer(data),
	}
	if got := srv.QuestItemEligible4F2590(item); got != 1 {
		t.Fatalf("native Street color eligibility = %d, want 1", got)
	}
	if srv.questItemEligibility.defaultReplenishment != replenishment {
		t.Fatalf("default cache = %p, want native Replenishment1 %p", srv.questItemEligibility.defaultReplenishment, replenishment)
	}
	bad := questItemServerAddModifier4F2590(srv, &names, "Material1", 11, 0, 0x405, 0)
	data.Modifiers[1] = bad
	if got := srv.QuestItemEligible4F2590(item); got != 0 {
		t.Fatalf("native Street material eligibility = %d, want 0", got)
	}
	runtime.KeepAlive(names)
}

func TestQuestItemServerNilFaultAfterLazyTypeResolution4F2590(t *testing.T) {
	srv := new(Server)
	questItemServerAddCachedTypes4F2590(srv)
	defer func() {
		if recover() == nil {
			t.Fatal("nil native item did not fault")
		}
		for index, typeInd := range srv.questItemEligibility.typeIDs {
			if typeInd != uint32(101+index) {
				t.Fatalf("type cache[%d] = %d, want %d before fault", index, typeInd, 101+index)
			}
		}
	}()
	srv.QuestItemEligible4F2590(nil)
}
