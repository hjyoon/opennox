package server

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestMapTransitionPlayerInitNative4FC6D0LayoutAndWidths(t *testing.T) {
	type layout struct {
		objectUpdateData uintptr
		updatePlayer     uintptr
		updateField138   uintptr
		playerIndex      uintptr
		playerField3680  uintptr
		playerField4792  uintptr
		updateSize       uintptr
		playerSize       uintptr
	}
	want := layout{
		objectUpdateData: 748,
		updatePlayer:     276,
		updateField138:   552,
		playerIndex:      2064,
		playerField3680:  3680,
		playerField4792:  4792,
		updateSize:       556,
		playerSize:       4828,
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		want = layout{
			objectUpdateData: 872,
			updatePlayer:     336,
			updateField138:   648,
			playerIndex:      2068,
			playerField3680:  4976,
			playerField4792:  6096,
			updateSize:       656,
			playerSize:       6160,
		}
	}
	got := layout{
		objectUpdateData: unsafe.Offsetof(Object{}.UpdateData),
		updatePlayer:     unsafe.Offsetof(PlayerUpdateData{}.Player),
		updateField138:   unsafe.Offsetof(PlayerUpdateData{}.Field138),
		playerIndex:      unsafe.Offsetof(Player{}.PlayerInd),
		playerField3680:  unsafe.Offsetof(Player{}.Field3680),
		playerField4792:  unsafe.Offsetof(Player{}.Field4792),
		updateSize:       unsafe.Sizeof(PlayerUpdateData{}),
		playerSize:       unsafe.Sizeof(Player{}),
	}
	if got != want {
		t.Fatalf("native layout = %+v, want %+v", got, want)
	}
	if unsafe.Sizeof(PlayerUpdateData{}.Field138) != 4 ||
		unsafe.Sizeof(Player{}.Field3680) != 4 ||
		unsafe.Sizeof(Player{}.Field4792) != 4 ||
		unsafe.Sizeof(Player{}.PlayerInd) != 1 {
		t.Fatal("map-transition fields lost their original scalar widths")
	}
}

func TestMapTransitionPlayerInitNative4FC6D0ReloadsPlayerAndIndexAndTraversesLiveList(t *testing.T) {
	s := new(Server)
	s.SetMapInitState4FC570(1)
	s.Players.list = make([]Player, 2)
	firstPlayer := &s.Players.list[0]
	secondPlayer := &s.Players.list[1]
	*firstPlayer = Player{Active: 1, PlayerInd: 0, Field4792: 1}
	*secondPlayer = Player{Active: 1, PlayerInd: 1}

	firstUnit := &Object{ObjClass: object.ClassPlayer}
	secondUnit := &Object{ObjClass: object.ClassPlayer}
	firstPlayer.PlayerUnit = firstUnit
	secondPlayer.PlayerUnit = secondUnit
	firstData := &PlayerUpdateData{Player: firstPlayer}
	secondData := &PlayerUpdateData{Player: secondPlayer}
	firstUnit.UpdateData = unsafe.Pointer(firstData)
	secondUnit.UpdateData = unsafe.Pointer(secondData)

	reloadedPlayer := &Player{PlayerInd: 2}
	var events []string
	runtime := MapTransitionPlayerInitRuntime4FC6D0{
		GameFlag: func(mask uint32) int32 {
			if mask == uint32(mapTransitionQuestFlag4FC6D0) {
				return 1
			}
			return 0
		},
		QuestStage:       func() int32 { return 0 },
		RestorePredicate: func() int32 { return 0 },
		RestoreReady:     func() int32 { panic("restore-ready must short-circuit") },
		QueuedRestore:    func() int32 { return 0 },
		SendQuestRestore: func(index uint8, value int32) {
			events = append(events, fmt.Sprintf("quest-restore:%d:%d", index, value))
		},
		MarkQuestReady: func(value int32) {
			events = append(events, fmt.Sprintf("quest-ready:%d", value))
		},
		FinishQuestTransition: func() {
			events = append(events, "quest-finish")
		},
		DataRoot: func() string {
			events = append(events, "data-root")
			return "/data"
		},
		FormatTempSavePath: func(root string) string {
			events = append(events, "format:"+root)
			return root + "/Save/_temp_.dat"
		},
		SavePlayerData: func(path string, index uint8) int32 {
			events = append(events, fmt.Sprintf("save:%s:%d", path, index))
			firstData.Player = reloadedPlayer
			return 1
		},
		PreparePlayerData: func(index uint8) int32 {
			events = append(events, fmt.Sprintf("prepare:%d", index))
			reloadedPlayer.PlayerInd = 3
			return 0
		},
		SendGauntlet: func(index uint8, value int32) {
			events = append(events, fmt.Sprintf("gauntlet:%d:%d", index, value))
			if value == 1 {
				reloadedPlayer.PlayerInd = 4
			} else {
				reloadedPlayer.PlayerInd = 6
			}
		},
		RestorePlayerData: func(path string, index uint8) int32 {
			events = append(events, fmt.Sprintf("restore:%s:%d", path, index))
			reloadedPlayer.PlayerInd = 5
			return 0
		},
		DeleteTempFile: func(path string) {
			events = append(events, "delete:"+path)
			reloadedPlayer.PlayerInd = 7
		},
		FinishPlayerData: func(index uint8) {
			events = append(events, fmt.Sprintf("finish:%d", index))
			if index == 7 {
				firstData.Player = firstPlayer
				firstPlayer.PlayerInd = 0
			}
		},
	}

	s.MapTransitionPlayerInit4FC6D0(runtime)

	want := []string{
		"data-root",
		"format:/data",
		"save:/data/Save/_temp_.dat:0",
		"prepare:2",
		"gauntlet:3:1",
		"restore:/data/Save/_temp_.dat:4",
		"gauntlet:5:0",
		"delete:/data/Save/_temp_.dat",
		"finish:7",
		"finish:1",
		"quest-restore:255:0",
		"quest-ready:1",
		"quest-finish",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events =\n%v\nwant\n%v", events, want)
	}
}

func TestMapTransitionPlayerInitNative4FC6D0OnlineUsesLowBytesAndNativeEnchant(t *testing.T) {
	s := new(Server)
	s.SetMapEntryState4FC580(1)
	s.Players.list = make([]Player, int(mapTransitionHostIndex4FC6D0)+1)
	fields := []struct {
		index     uint8
		field3680 uint32
	}{
		{index: 3, field3680: 0x100},
		{index: 4, field3680: 0x101},
		{index: mapTransitionHostIndex4FC6D0},
	}
	units := make(map[uint8]*Object, len(fields))
	data := make([]PlayerUpdateData, len(fields))
	for i, field := range fields {
		player := &s.Players.list[field.index]
		*player = Player{Active: 1, PlayerInd: field.index, Field3680: field.field3680}
		unit := &Object{ObjClass: object.ClassPlayer}
		units[field.index] = unit
		player.PlayerUnit = unit
		data[i].Player = player
		unit.UpdateData = unsafe.Pointer(&data[i])
	}

	var events []string
	s.MapTransitionPlayerInit4FC6D0(MapTransitionPlayerInitRuntime4FC6D0{
		GameFlag: func(mask uint32) int32 {
			switch mask {
			case uint32(mapTransitionOnlineFlag4FC6D0):
				return 1
			case uint32(mapTransitionChatFlag4FC6D0):
				return 0
			default:
				return 0
			}
		},
		FadeBegin: func(a, b int32) {
			events = append(events, fmt.Sprintf("fade:%d:%d", a, b))
		},
		ApplyEnchant: func(unit *Object, enchant EnchantID, duration, power int32) {
			events = append(events, fmt.Sprintf("enchant:%t:%d:%d:%d", unit == units[3], enchant, duration, power))
		},
	})

	want := []string{"fade:1:1", "enchant:true:23:0:5"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}
