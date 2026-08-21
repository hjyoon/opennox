package server

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"
)

func playerMakeDefItemsNativeTestDeps4EF7D0(events *[]string) playerMakeDefItemsNativeDeps4EF7D0 {
	record := func(event string) {
		*events = append(*events, event)
	}
	modifiers := map[uint32]*ModifierEff{
		100: {ind4: 1000},
		200: {ind4: 200},
		300: {ind4: 300},
		400: {ind4: 400},
	}
	for id := uint32(1000); id < 1100; id++ {
		modifiers[id] = &ModifierEff{ind4: id}
	}
	return playerMakeDefItemsNativeDeps4EF7D0{
		removePoison: func(*Object) {
			record("poison")
		},
		setHealthMaximum: func(*Object) {
			record("health-max")
		},
		refreshMana: func(*Object) {
			record("mana-max")
		},
		cancelAbilities: func(*Object) {
			record("cancel")
		},
		resetCamping: func(*Object) {
			record("camping")
		},
		setPlayerState: func(_ *Object, state PlayerState) {
			record(fmt.Sprintf("state:%d", state))
		},
		clearBuffs: func(*Object) {
			record("buffs")
		},
		resetPlayerRuntime: func(*Object) {
			record("runtime")
		},
		gameFlag: func(flag uint32) bool {
			record(fmt.Sprintf("game:%#x", flag))
			return false
		},
		markPlayerObjects: func(index uint8) {
			record(fmt.Sprintf("mark:%d", index))
		},
		reportTotalHealth: func(index uint8, _ *Object) {
			record(fmt.Sprintf("health:%d", index))
		},
		reportTotalMana: func(index uint8, _ *Object) {
			record(fmt.Sprintf("mana:%d", index))
		},
		sendRespawn: func(_ *Object, keep uint8) uint8 {
			record(fmt.Sprintf("send:%d", keep))
			return 0xa5 + keep
		},
		armorEquipFlags: func(*Object) uint32 {
			record("armor-flags")
			return 0
		},
		delayedDelete: func(*Object) {
			record("delete")
		},
		modifierID: func(name string) uint32 {
			record("modifier-id:" + name)
			switch name {
			case "UserColor1":
				return 100
			case "ArmorQuality1":
				return 200
			case "Material1":
				return 300
			case "Replenishment1":
				return 400
			default:
				return 0xff
			}
		},
		modifierDesc: func(id uint32) *ModifierEff {
			record(fmt.Sprintf("modifier-desc:%d", id))
			return modifiers[id]
		},
		respawnItem: func(_ *Object, typeID string, _ *ModifierInitData, _, _ int32) *Object {
			record("item:" + typeID)
			return &Object{}
		},
		questDefaultsReady: func() int32 {
			record("quest-ready")
			return 0
		},
	}
}

func TestPlayerMakeDefItemsNative4EF7D0ResetsCachedNativeState(t *testing.T) {
	player := &Player{PlayerInd: 7}
	player.Info().SetPlayerClass(2)
	health := &HealthData{Cur: 0x1234, Max: 0x5678}
	update := &PlayerUpdateData{
		Player:          player,
		QuestExit:       &Object{},
		QuestWarpGate:   &Object{},
		Field21:         1,
		Field19_1:       2,
		HealthSampleCur: 3,
		Field47_0:       4,
		SpellCastStart:  5,
		TrapSpells:      [5]uint32{6, 7, 8, 9, 10},
		TrapSpellsCnt:   0xaabbccdd,
		HarpoonBolt:     &Object{},
		HarpoonTarg:     &Object{},
		Field67:         11,
	}
	otherUpdate := &PlayerUpdateData{Field21: 0xfeedbeef}
	unit := &Object{
		ObjFlags:   0xffffffff,
		Field541:   0xff,
		HealthData: health,
		Obj130:     &Object{},
		UpdateData: unsafe.Pointer(update),
	}
	var events []string
	deps := playerMakeDefItemsNativeTestDeps4EF7D0(&events)
	baseCancel := deps.cancelAbilities
	deps.cancelAbilities = func(got *Object) {
		baseCancel(got)
		got.UpdateData = unsafe.Pointer(otherUpdate)
	}

	result := playerMakeDefItemsNative4EF7D0(unit, 1, 1, deps)
	if result.kind != playerMakeDefItemsResultByte4EF7D0 || result.value != 0xa5 {
		t.Fatalf("result = %#v, want send byte 0xa5", result)
	}
	wantEvents := []string{
		"poison", "health-max", "mana-max", "cancel", "camping", "state:13", "buffs", "runtime",
		"game:0x2000", "health:7", "mana:7", "send:0",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if update.QuestExit != nil || update.QuestWarpGate != nil || update.Field21 != 0 || update.Field19_1 != 0 ||
		update.HealthSampleCur != 0 || update.Field47_0 != 0 || update.SpellCastStart != 0 || update.TrapSpells != [5]uint32{} ||
		update.TrapSpellsCnt != 0xaabbcc00 || update.HarpoonBolt != nil || update.HarpoonTarg != nil || update.Field67 != 0 {
		t.Fatalf("cached update was not reset exactly: %#v", update)
	}
	for i, sample := range update.HealthSamples {
		if sample != health.Cur {
			t.Fatalf("sample %d = %#x, want %#x", i, sample, health.Cur)
		}
	}
	if unit.Field541 != 0 || unit.Obj130 != nil || uint32(unit.ObjFlags) != playerMakeDefItemsObjectFlagMask4EF7D0 {
		t.Fatalf("unit reset = field541 %d obj130 %p flags %#x", unit.Field541, unit.Obj130, unit.ObjFlags)
	}
	if player.Field4700 != 1 || otherUpdate.Field21 != 0xfeedbeef {
		t.Fatalf("cached/live state = done %d other field %#x", player.Field4700, otherUpdate.Field21)
	}
}

func TestPlayerMakeDefItemsNative4EF7D0UsesNativeModifierPointers(t *testing.T) {
	player := &Player{PlayerInd: 3, ArmorEquip: 0x405}
	player.Info().SetPlayerClass(0)
	update := &PlayerUpdateData{Player: player}
	unit := &Object{HealthData: &HealthData{Cur: 9}, UpdateData: unsafe.Pointer(update)}
	var events []string
	deps := playerMakeDefItemsNativeTestDeps4EF7D0(&events)
	baseGameFlag := deps.gameFlag
	deps.gameFlag = func(flag uint32) bool {
		baseGameFlag(flag)
		return flag == 0x0800
	}
	var attrs *ModifierInitData
	created := &Object{}
	deps.respawnItem = func(_ *Object, typeID string, got *ModifierInitData, a4, a5 int32) *Object {
		events = append(events, fmt.Sprintf("item:%s:%d:%d", typeID, a4, a5))
		attrs = got
		return created
	}

	result := playerMakeDefItemsNative4EF7D0(unit, 0, 0, deps)
	if result.kind != playerMakeDefItemsResultObject4EF7D0 || result.object != created {
		t.Fatalf("result = %#v, want created object", result)
	}
	if attrs == nil || attrs.Modifiers[0] == nil || attrs.Modifiers[0].ind4 != 200 || attrs.Modifiers[1] == nil || attrs.Modifiers[1].ind4 != 300 || attrs.Modifiers[2] != nil || attrs.Modifiers[3] != nil {
		t.Fatalf("modifier attrs = %#v", attrs)
	}
	if got, want := playerMakeDefItemsResultLow4EF7D0(result), uint8(uintptr(unsafe.Pointer(created))); got != want {
		t.Fatalf("low result = %#x, want native pointer byte %#x", got, want)
	}
}

func TestPlayerMakeDefItemsNative4EF7D0CachesInfoReloadsPlayer(t *testing.T) {
	playerA := &Player{PlayerInd: 7, ArmorEquip: 0x405}
	playerA.Info().SetPlayerClass(0)
	playerA.Info().Colors.Shirt1 = 4
	playerA.Info().Colors.Shirt2 = 5
	playerB := &Player{PlayerInd: 9, ArmorEquip: 0x005}
	playerB.Info().SetPlayerClass(2)
	playerB.Info().Colors.Shirt1 = 44
	playerB.Info().Colors.Shirt2 = 55
	update := &PlayerUpdateData{Player: playerA}
	unit := &Object{HealthData: &HealthData{Cur: 1}, UpdateData: unsafe.Pointer(update)}
	var events []string
	deps := playerMakeDefItemsNativeTestDeps4EF7D0(&events)
	deps.gameFlag = func(flag uint32) bool {
		events = append(events, fmt.Sprintf("game:%#x", flag))
		return flag == 0x0a00
	}
	deps.reportTotalHealth = func(index uint8, _ *Object) {
		events = append(events, fmt.Sprintf("health:%d", index))
		update.Player = playerB
	}
	var shirtAttrs *ModifierInitData
	deps.respawnItem = func(_ *Object, typeID string, attrs *ModifierInitData, _, _ int32) *Object {
		events = append(events, "item:"+typeID)
		if typeID == "StreetShirt" {
			shirtAttrs = attrs
		}
		return &Object{}
	}

	result := playerMakeDefItemsNative4EF7D0(unit, 0, 0, deps)
	if result.kind != playerMakeDefItemsResultByte4EF7D0 || result.value != 2 {
		t.Fatalf("result = %#v, want live player class byte 2", result)
	}
	if shirtAttrs == nil || shirtAttrs.Modifiers[1] == nil || shirtAttrs.Modifiers[1].ind4 != 1004 || shirtAttrs.Modifiers[2] == nil || shirtAttrs.Modifiers[2].ind4 != 1005 {
		t.Fatalf("shirt did not use cached player-A info: %#v", shirtAttrs)
	}
	if playerA.Field4700 != 0 || playerB.Field4700 != 1 {
		t.Fatalf("completion did not use live player: A=%d B=%d", playerA.Field4700, playerB.Field4700)
	}
	if !containsPlayerMakeDefItemsEvent4EF7D0(events, "mana:9") {
		t.Fatalf("events do not show live player reload: %v", events)
	}
}

func containsPlayerMakeDefItemsEvent4EF7D0(events []string, want string) bool {
	for _, event := range events {
		if event == want {
			return true
		}
	}
	return false
}

func TestPlayerMakeDefItemsResultLow4EF7D0PreservesSources(t *testing.T) {
	player := &Player{}
	object := &Object{}
	tests := []struct {
		name   string
		result playerMakeDefItemsResult4EF7D0[*Object, *Player]
		want   uint8
	}{
		{name: "nil player", result: playerMakeDefItemsResult4EF7D0[*Object, *Player]{kind: playerMakeDefItemsResultPlayer4EF7D0}, want: 0},
		{name: "player", result: playerMakeDefItemsResult4EF7D0[*Object, *Player]{kind: playerMakeDefItemsResultPlayer4EF7D0, player: player}, want: uint8(uintptr(unsafe.Pointer(player)))},
		{name: "object", result: playerMakeDefItemsResult4EF7D0[*Object, *Player]{kind: playerMakeDefItemsResultObject4EF7D0, object: object}, want: uint8(uintptr(unsafe.Pointer(object)))},
		{name: "byte", result: playerMakeDefItemsResult4EF7D0[*Object, *Player]{kind: playerMakeDefItemsResultByte4EF7D0, value: 0xe7}, want: 0xe7},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := playerMakeDefItemsResultLow4EF7D0(test.result); got != test.want {
				t.Fatalf("result = %#x, want %#x", got, test.want)
			}
		})
	}
}
