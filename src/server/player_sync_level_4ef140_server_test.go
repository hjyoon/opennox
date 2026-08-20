package server

import (
	"runtime"
	"testing"
	"unsafe"
)

func TestPlayerSyncLevelNative4EF140CooperativeFieldsAndServices(t *testing.T) {
	player := &Player{Level: 3, ProtPlayerLevel: 0x89abcdef}
	update := &PlayerUpdateData{Player: player}
	unit := &Object{Experience: 123.5, UpdateData: unsafe.Pointer(update)}
	flagCalls := 0
	protectCalls := 0
	readCalls := 0
	result := playerSyncLevelNative4EF140(unit, playerSyncLevelNativeDeps4EF140{
		gameFlagsCheck: func(mask uint32) int32 {
			flagCalls++
			if mask != 0x2000 {
				t.Fatalf("flag mask = %#x, want 0x2000", mask)
			}
			return -1
		},
		loadXPTable: func(int32) float64 {
			t.Fatal("XPTable load in cooperative path")
			return 0
		},
		protectLevel: func(token uint32, level uint8) {
			protectCalls++
			if token != 0x89abcdef || level != 10 {
				t.Fatalf("protection args = %#x/%d, want 0x89abcdef/10", token, level)
			}
			if player.Level != 10 {
				t.Fatalf("level at protection = %d, want 10", player.Level)
			}
		},
		readValues: func(got *Object, reward int32) int32 {
			readCalls++
			if got != unit || reward != 0 || player.Level != 10 {
				t.Fatalf("read-values args/state = %p/%d/%d", got, reward, player.Level)
			}
			return -123456789
		},
	})
	if result != -123456789 || flagCalls != 1 || protectCalls != 1 || readCalls != 1 {
		t.Fatalf("result/calls = %d/%d/%d/%d", result, flagCalls, protectCalls, readCalls)
	}
	if unit.Experience != 123.5 {
		t.Fatalf("experience changed to %v", unit.Experience)
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update)
	runtime.KeepAlive(player)
}

func TestPlayerSyncLevelNative4EF140NormalExperienceAndLowByte(t *testing.T) {
	tests := []struct {
		name       string
		experience float32
		wantLevel  uint8
		wantCalls  int
	}{
		{name: "below zero", experience: -1, wantLevel: 0xff, wantCalls: 1},
		{name: "middle", experience: 250, wantLevel: 2, wantCalls: 4},
		{name: "maximum", experience: 2000, wantLevel: 10, wantCalls: 11},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			player := &Player{Level: 44, ProtPlayerLevel: 77}
			update := &PlayerUpdateData{Player: player}
			unit := &Object{Experience: test.experience, UpdateData: unsafe.Pointer(update)}
			tableCalls := 0
			var protected uint8
			result := playerSyncLevelNative4EF140(unit, playerSyncLevelNativeDeps4EF140{
				gameFlagsCheck: func(mask uint32) int32 {
					if mask != 0x2000 {
						t.Fatalf("flag mask = %#x", mask)
					}
					return 0
				},
				loadXPTable: func(index int32) float64 {
					tableCalls++
					return float64(index * 100)
				},
				protectLevel: func(token uint32, level uint8) {
					if token != 77 {
						t.Fatalf("token = %d, want 77", token)
					}
					protected = level
				},
				readValues: func(got *Object, reward int32) int32 {
					if got != unit || reward != 0 {
						t.Fatalf("read-values args = %p/%d", got, reward)
					}
					return 987654321
				},
			})
			if result != 987654321 || player.Level != test.wantLevel || protected != test.wantLevel || tableCalls != test.wantCalls {
				t.Fatalf("result/level/protected/table calls = %d/%#02x/%#02x/%d", result, player.Level, protected, tableCalls)
			}
			runtime.KeepAlive(unit)
			runtime.KeepAlive(update)
			runtime.KeepAlive(player)
		})
	}
}

func TestPlayerSyncLevelNative4EF140HasNoPointerOrClassGates(t *testing.T) {
	t.Run("nil unit", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("nil unit did not fault")
			}
		}()
		playerSyncLevelNative4EF140(nil, playerSyncLevelNativeDeps4EF140{
			gameFlagsCheck: func(uint32) int32 {
				t.Fatal("game flag after nil unit")
				return 0
			},
		})
	})

	t.Run("nil update", func(t *testing.T) {
		unit := &Object{}
		defer func() {
			if recover() == nil {
				t.Fatal("nil update did not fault")
			}
		}()
		playerSyncLevelNative4EF140(unit, playerSyncLevelNativeDeps4EF140{
			gameFlagsCheck: func(uint32) int32 {
				t.Fatal("game flag after nil update")
				return 0
			},
		})
	})

	t.Run("class is not inspected", func(t *testing.T) {
		player := &Player{ProtPlayerLevel: 9}
		update := &PlayerUpdateData{Player: player}
		unit := &Object{ObjClass: 0, UpdateData: unsafe.Pointer(update)}
		if got := playerSyncLevelNative4EF140(unit, playerSyncLevelNativeDeps4EF140{
			gameFlagsCheck: func(uint32) int32 { return 1 },
			protectLevel:   func(uint32, uint8) {},
			readValues:     func(*Object, int32) int32 { return 17 },
		}); got != 17 || player.Level != 10 {
			t.Fatalf("result/level = %d/%d, want 17/10", got, player.Level)
		}
	})
}

func TestPlayerSyncLevel4EF140NativeLayouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantExperience := uintptr(28)
	wantUpdate := uintptr(748)
	wantUpdateSize := uintptr(556)
	wantPlayer := uintptr(276)
	wantPlayerSize := uintptr(4828)
	wantLevel := uintptr(3684)
	wantProtection := uintptr(4644)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantExperience = 32
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
