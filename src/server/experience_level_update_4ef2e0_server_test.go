package server

import (
	"reflect"
	"runtime"
	"testing"
	"unsafe"
)

func TestExperienceLevelUpdateNative4EF2E0BindsFieldsAndSoloServices(t *testing.T) {
	player := &Player{Level: 4, ProtPlayerLevel: 0x89abcdef}
	update := &PlayerUpdateData{Player: player}
	unit := &Object{Experience: 500, NetCode: 0x10203040, UpdateData: unsafe.Pointer(update)}
	var events []string
	experienceLevelUpdateNative4EF2E0(unit, experienceLevelUpdateNativeDeps4EF2E0{
		gameGet: func() int32 {
			events = append(events, "game-get")
			return 0
		},
		gameSubActive: func() bool {
			t.Fatal("game-sub called for game-get zero")
			return false
		},
		loadXPTable: func(key string, index int32) float64 {
			events = append(events, "table")
			if key != "XPTable" || index != 5 {
				t.Fatalf("table args = %q/%d, want XPTable/5", key, index)
			}
			player.Level = 9
			unit.Experience = 600
			return 400
		},
		protectLevel: func(token uint32, level uint8) {
			events = append(events, "protect")
			if token != 0x89abcdef || level != 1 || player.Level != 10 {
				t.Fatalf("protection args/state = %08x/%d/%d", token, level, player.Level)
			}
		},
		readValues: func(got *Object, reward int32) int32 {
			events = append(events, "read-values")
			if got != unit || reward != 1 || player.Level != 10 {
				t.Fatalf("read-values args/state = %p/%d/%d", got, reward, player.Level)
			}
			return -123
		},
		gameFlag: func(flag uint32) int32 {
			events = append(events, "game-flag")
			if flag != 0x800 {
				t.Fatalf("game flag = %#x, want 0x800", flag)
			}
			unit.NetCode = 0x55667788
			return 0
		},
		pauseFX: func(*Object, int32) { t.Fatal("pause called on solo path") },
		audio: func(id uint32, got *Object, kind int32, code uint32) {
			events = append(events, "audio")
			if id != 902 || got != unit || kind != 2 || code != 0x55667788 {
				t.Fatalf("audio args = %d/%p/%d/%08x", id, got, kind, code)
			}
		},
		loadString: func(key, path string, line int) string {
			events = append(events, "string")
			if key != "LevelUP" || path != `C:\NoxPost\src\Server\GameMech\explevel.c` || line != 253 {
				t.Fatalf("string args = %q/%q/%d", key, path, line)
			}
			return "native-level-message"
		},
		sendLineMessage: func(got *Object, message string) bool {
			events = append(events, "line")
			if got != unit || message != "native-level-message" {
				t.Fatalf("line args = %p/%q", got, message)
			}
			return false
		},
	})
	want := []string{"game-get", "table", "protect", "read-values", "game-flag", "audio", "string", "line"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
	if player.Level != 10 || unit.Experience != 600 || unit.NetCode != 0x55667788 {
		t.Fatalf("final fields = level %d, experience %v, net code %08x", player.Level, unit.Experience, unit.NetCode)
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update)
	runtime.KeepAlive(player)
}

func TestExperienceLevelUpdateNative4EF2E0CoopDefersSoloFields(t *testing.T) {
	player := &Player{Level: 1, ProtPlayerLevel: 7}
	update := &PlayerUpdateData{Player: player}
	unit := &Object{Experience: 10, NetCode: 99, UpdateData: unsafe.Pointer(update)}
	paused := 0
	experienceLevelUpdateNative4EF2E0(unit, experienceLevelUpdateNativeDeps4EF2E0{
		gameGet:       func() int32 { return 1 },
		gameSubActive: func() bool { return false },
		loadXPTable:   func(string, int32) float64 { return 10 },
		protectLevel:  func(uint32, uint8) {},
		readValues:    func(*Object, int32) int32 { return 0 },
		gameFlag:      func(uint32) int32 { return -1 },
		pauseFX: func(got *Object, mode int32) {
			paused++
			if got != unit || mode != 0 {
				t.Fatalf("pause args = %p/%d", got, mode)
			}
		},
		audio:           func(uint32, *Object, int32, uint32) { t.Fatal("audio called") },
		loadString:      func(string, string, int) string { t.Fatal("string loaded"); return "" },
		sendLineMessage: func(*Object, string) bool { t.Fatal("line called"); return false },
	})
	if paused != 1 || player.Level != 2 || unit.NetCode != 99 {
		t.Fatalf("pause/level/net code = %d/%d/%d", paused, player.Level, unit.NetCode)
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update)
	runtime.KeepAlive(player)
}

func TestExperienceLevelUpdateNative4EF2E0HasNoClassGate(t *testing.T) {
	player := &Player{Level: 2, ProtPlayerLevel: 3}
	update := &PlayerUpdateData{Player: player}
	unit := &Object{ObjClass: 0, Experience: 0, UpdateData: unsafe.Pointer(update)}
	experienceLevelUpdateNative4EF2E0(unit, experienceLevelUpdateNativeDeps4EF2E0{
		gameGet:         func() int32 { return 1 },
		gameSubActive:   func() bool { return true },
		loadXPTable:     func(string, int32) float64 { t.Fatal("table called"); return 0 },
		protectLevel:    func(uint32, uint8) { t.Fatal("protect called") },
		readValues:      func(*Object, int32) int32 { t.Fatal("read called"); return 0 },
		gameFlag:        func(uint32) int32 { t.Fatal("flag called"); return 0 },
		pauseFX:         func(*Object, int32) { t.Fatal("pause called") },
		audio:           func(uint32, *Object, int32, uint32) { t.Fatal("audio called") },
		loadString:      func(string, string, int) string { t.Fatal("string called"); return "" },
		sendLineMessage: func(*Object, string) bool { t.Fatal("line called"); return false },
	})
	if player.Level != 2 {
		t.Fatalf("early return changed level to %d", player.Level)
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update)
	runtime.KeepAlive(player)
}

func TestExperienceLevelUpdate4EF2E0NativeLayouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantExperience := uintptr(28)
	wantNetCode := uintptr(36)
	wantUpdate := uintptr(748)
	wantUpdateSize := uintptr(556)
	wantPlayer := uintptr(276)
	wantPlayerSize := uintptr(4828)
	wantLevel := uintptr(3684)
	wantProtection := uintptr(4644)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantExperience = 32
		wantNetCode = 40
		wantUpdate = 872
		wantUpdateSize = 640
		wantPlayer = 320
		wantPlayerSize = 6160
		wantLevel = 4980
		wantProtection = 5948
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.Experience", unsafe.Offsetof(Object{}.Experience), wantExperience},
		{"Object.NetCode", unsafe.Offsetof(Object{}.NetCode), wantNetCode},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantUpdateSize},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayer},
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.Level", unsafe.Offsetof(Player{}.Level), wantLevel},
		{"Player.ProtPlayerLevel", unsafe.Offsetof(Player{}.ProtPlayerLevel), wantProtection},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}
