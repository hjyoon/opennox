package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	playerlib "github.com/opennox/libs/player"
)

func TestPlayerLevelSetNative4EF410BindsFieldsAndServices(t *testing.T) {
	player := &Player{
		Level:              2,
		ProtUnitExperience: 0x11223344,
		ProtPlayerLevel:    0x55667788,
	}
	player.Info().SetPlayerClass(playerlib.Warrior)
	player.SpellLvl[1] = 11
	player.SpellLvl[3] = 33
	player.SpellLvl[5] = 55
	update := &PlayerUpdateData{Player: player}
	unit := &Object{Experience: 1, UpdateData: unsafe.Pointer(update)}
	var events []string
	tableCalls, flagCalls := 0, 0

	playerLevelSetNative4EF410(unit, 4, playerLevelSetNativeDeps4EF410{
		loadXPTable: func(key string, index int32) float64 {
			tableCalls++
			events = append(events, "table")
			if key != "XPTable" || index != 4 {
				t.Fatalf("table args = %q/%d, want XPTable/4", key, index)
			}
			if tableCalls == 1 {
				return 100.25
			}
			player.ProtUnitExperience = 0xaabbccdd
			return 200.5
		},
		protectExperience: func(token uint32, experience float32) {
			events = append(events, "protect-experience")
			if token != 0xaabbccdd || math.Float32bits(experience) != math.Float32bits(200.5) || math.Float32bits(unit.Experience) != math.Float32bits(100.25) {
				t.Fatalf("experience protection args/state = %08x/%08x/%08x", token, math.Float32bits(experience), math.Float32bits(unit.Experience))
			}
			player.ProtPlayerLevel = 0xcafebabe
		},
		reportExperience: func(got *Object) {
			events = append(events, "report")
			if got != unit {
				t.Fatalf("report unit = %p, want %p", got, unit)
			}
			player.ProtPlayerLevel = 0x10203040
		},
		protectLevel: func(token uint32, level uint8) {
			events = append(events, "protect-level")
			if token != 0x10203040 || level != 4 || player.Level != 4 {
				t.Fatalf("level protection args/state = %08x/%d/%d", token, level, player.Level)
			}
		},
		readValues: func(got *Object, value int32) int32 {
			events = append(events, "read-values")
			if got != unit || value != 0 {
				t.Fatalf("read-values args = %p/%d", got, value)
			}
			return -123
		},
		gameFlag: func(flag uint32) int32 {
			flagCalls++
			events = append(events, "flag")
			if flag != 0x800 {
				t.Fatalf("game flag = %#x, want 0x800", flag)
			}
			return 1
		},
		bookAbility: func(kind, ability, index int32) {
			events = append(events, "book")
			if kind != 3 || ability != index+1 || (ability != 1 && ability != 3 && ability != 5) {
				t.Fatalf("book args = %d/%d/%d", kind, ability, index)
			}
		},
		pauseFX: func(got *Object, value int32) {
			events = append(events, "pause")
			if got != unit || value != 0 {
				t.Fatalf("pause args = %p/%d", got, value)
			}
		},
	})

	want := []string{
		"table", "table", "protect-experience", "report", "protect-level",
		"read-values", "flag", "book", "book", "book", "flag", "pause",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
	if tableCalls != 2 || flagCalls != 2 || player.Level != 4 || math.Float32bits(unit.Experience) != math.Float32bits(100.25) {
		t.Fatalf("tables/flags/level/experience = %d/%d/%d/%08x", tableCalls, flagCalls, player.Level, math.Float32bits(unit.Experience))
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update)
	runtime.KeepAlive(player)
}

func TestPlayerLevelSetNative4EF410SignedByteAndNonWarriorBranch(t *testing.T) {
	for _, level := range []uint8{11, 0x80, 0xff} {
		player := &Player{ProtUnitExperience: 1, ProtPlayerLevel: 2}
		player.Info().SetPlayerClass(playerlib.Wizard)
		update := &PlayerUpdateData{Player: player}
		unit := &Object{UpdateData: unsafe.Pointer(update)}
		indexes := make([]int32, 0, 2)
		flags := 0
		playerLevelSetNative4EF410(unit, level, playerLevelSetNativeDeps4EF410{
			loadXPTable: func(_ string, index int32) float64 {
				indexes = append(indexes, index)
				return float64(index)
			},
			protectExperience: func(uint32, float32) {},
			reportExperience:  func(*Object) {},
			protectLevel:      func(uint32, uint8) {},
			readValues:        func(*Object, int32) int32 { return 0 },
			gameFlag: func(uint32) int32 {
				flags++
				return 1
			},
			bookAbility: func(int32, int32, int32) { t.Fatal("book called for non-warrior") },
			pauseFX:     func(*Object, int32) {},
		})
		wantLevel := level
		if int8(wantLevel) > 10 {
			wantLevel = 10
		}
		wantIndexes := []int32{int32(int8(wantLevel)), int32(int8(wantLevel))}
		if player.Level != wantLevel || !reflect.DeepEqual(indexes, wantIndexes) || flags != 2 {
			t.Fatalf("input %02x level/indexes/flags = %02x/%v/%d, want %02x/%v/2", level, player.Level, indexes, flags, wantLevel, wantIndexes)
		}
		runtime.KeepAlive(unit)
		runtime.KeepAlive(update)
		runtime.KeepAlive(player)
	}
}

func TestPlayerLevelSet4EF410NativeLayouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantExperience := uintptr(28)
	wantUpdate := uintptr(748)
	wantUpdateSize := uintptr(556)
	wantPlayer := uintptr(276)
	wantPlayerSize := uintptr(4828)
	wantInfo := uintptr(2185)
	wantClass := uintptr(2251)
	wantLevel := uintptr(3684)
	wantAbilities := uintptr(3696)
	wantExperienceToken := uintptr(4604)
	wantLevelToken := uintptr(4644)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantExperience = 32
		wantUpdate = 872
		wantUpdateSize = 656
		wantPlayer = 336
		wantPlayerSize = 6160
		wantInfo = 2189
		wantClass = 2255
		wantLevel = 4980
		wantAbilities = 4992
		wantExperienceToken = 5908
		wantLevelToken = 5948
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.Experience", unsafe.Offsetof(Object{}.Experience), wantExperience},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantUpdateSize},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayer},
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.info", unsafe.Offsetof(Player{}.info), wantInfo},
		{"Player.info.playerClass", unsafe.Offsetof(Player{}.info) + unsafe.Offsetof(PlayerInfo{}.playerClass), wantClass},
		{"Player.Level", unsafe.Offsetof(Player{}.Level), wantLevel},
		{"Player.SpellLvl", unsafe.Offsetof(Player{}.SpellLvl), wantAbilities},
		{"Player.ProtUnitExperience", unsafe.Offsetof(Player{}.ProtUnitExperience), wantExperienceToken},
		{"Player.ProtPlayerLevel", unsafe.Offsetof(Player{}.ProtPlayerLevel), wantLevelToken},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
	if unsafe.Sizeof(Player{}.SpellLvl[0]) != 4 || len(Player{}.SpellLvl) != 137 {
		t.Fatalf("ability element/count = %d/%d, want 4/137", unsafe.Sizeof(Player{}.SpellLvl[0]), len(Player{}.SpellLvl))
	}
}
