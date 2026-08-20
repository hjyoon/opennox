package server

import (
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

func defaultPlayerStatusNativeDeps4174F0() playerStatusNativeDeps4174F0 {
	return playerStatusNativeDeps4174F0{
		gameFlag:       func(uint32) int32 { return 0 },
		reportStatus:   func(*Player) int32 { return 0 },
		anyPlayers:     func() int32 { return 0 },
		timerStatus:    func() int32 { return 0 },
		gameFlagsValue: func() uint32 { return 0 },
		modeEnabled:    func(int16) uint8 { return 0 },
		startTimer:     func() {},
		setTimerStatus: func(int32) int32 { return 0 },
	}
}

func TestPlayerStatus4174F0NativeLayouts(t *testing.T) {
	wantSize := uintptr(4828)
	wantNetCode := uintptr(2060)
	wantFlags := uintptr(3680)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantSize = 6160
		wantNetCode = 2064
		wantFlags = 4976
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Player size", unsafe.Sizeof(Player{}), wantSize},
		{"Player.NetCodeVal", unsafe.Offsetof(Player{}.NetCodeVal), wantNetCode},
		{"Player.Field3680", unsafe.Offsetof(Player{}.Field3680), wantFlags},
		{"Player.NetCodeVal width", unsafe.Sizeof(Player{}.NetCodeVal), 4},
		{"Player.Field3680 width", unsafe.Sizeof(Player{}.Field3680), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestPlayerNeedStatusNative4174F0StoresBeforeReport(t *testing.T) {
	player := &Player{Field3680: 0x20}
	var events []string
	deps := defaultPlayerStatusNativeDeps4174F0()
	deps.gameFlag = func(flag uint32) int32 {
		events = append(events, "host")
		if flag != 1 || player.Field3680 != 0x420 {
			t.Fatalf("host observed flag/state = %#x/%#x, want 1/0x420", flag, player.Field3680)
		}
		return 1
	}
	deps.reportStatus = func(got *Player) int32 {
		events = append(events, "report")
		if got != player || got.Field3680 != 0x420 {
			t.Fatalf("report player/state = %p/%#x, want %p/0x420", got, got.Field3680, player)
		}
		return 77
	}
	if got := playerNeedStatusNative4174F0(player, 0x400, deps); got != 77 {
		t.Fatalf("result = %d, want report result 77", got)
	}
	if !reflect.DeepEqual(events, []string{"host", "report"}) {
		t.Fatalf("events = %q, want [host report]", events)
	}
}

func TestPlayerUnsetStatusNative417530PoisonMaskSkipsObserverServices(t *testing.T) {
	player := &Player{Field3680: 0x420}
	deps := defaultPlayerStatusNativeDeps4174F0()
	deps.gameFlag = func(flag uint32) int32 {
		if flag != 1 || player.Field3680 != 0x20 {
			t.Fatalf("host observed flag/state = %#x/%#x, want 1/0x20", flag, player.Field3680)
		}
		return 1
	}
	deps.anyPlayers = func() int32 { t.Fatal("poison mask reached observer any-player service"); return 0 }
	deps.timerStatus = func() int32 { t.Fatal("poison mask reached observer timer service"); return 0 }
	deps.gameFlagsValue = func() uint32 { t.Fatal("poison mask reached observer game flags"); return 0 }
	deps.modeEnabled = func(int16) uint8 { t.Fatal("poison mask reached observer mode service"); return 0 }
	deps.startTimer = func() { t.Fatal("poison mask reached observer timer start") }
	deps.setTimerStatus = func(int32) int32 { t.Fatal("poison mask reached observer timer state"); return 0 }
	deps.reportStatus = func(got *Player) int32 {
		if got != player || got.Field3680 != 0x20 {
			t.Fatalf("report player/state = %p/%#x, want %p/0x20", got, got.Field3680, player)
		}
		return -9
	}
	if got := playerUnsetStatusNative417530(player, 0x400, deps); got != -9 {
		t.Fatalf("result = %d, want report result -9", got)
	}
}

func TestPlayerUnsetStatusNative417530BindsObserverChain(t *testing.T) {
	player := &Player{Field3680: 1}
	var events []string
	deps := defaultPlayerStatusNativeDeps4174F0()
	deps.gameFlag = func(flag uint32) int32 {
		events = append(events, "game")
		if flag == 1 {
			return 1
		}
		return 0
	}
	deps.anyPlayers = func() int32 { events = append(events, "any"); return 1 }
	deps.timerStatus = func() int32 { events = append(events, "timer"); return 0 }
	deps.gameFlagsValue = func() uint32 { events = append(events, "flags"); return 0xabcd }
	deps.modeEnabled = func(value int16) uint8 {
		events = append(events, "mode")
		if uint16(value) != 0xabcd {
			t.Fatalf("mode value = %#x, want 0xabcd", uint16(value))
		}
		return 1
	}
	deps.startTimer = func() { events = append(events, "start") }
	deps.setTimerStatus = func(value int32) int32 {
		events = append(events, "set")
		if value != 1 {
			t.Fatalf("timer state = %d, want 1", value)
		}
		return 88
	}
	deps.reportStatus = func(*Player) int32 { events = append(events, "report"); return 99 }
	if got := playerUnsetStatusNative417530(player, 1, deps); got != 99 {
		t.Fatalf("result = %d, want report result 99", got)
	}
	want := []string{"game", "game", "any", "timer", "flags", "mode", "start", "set", "report"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
}

func TestPlayerStatusReportNative417630Packet(t *testing.T) {
	player := &Player{NetCodeVal: 0x1234abcd, Field3680: 0xffffffff}
	got := playerStatusReportNative417630(player, func(recipient byte, packet []byte, remove int32) int32 {
		if recipient != 255 || remove != 0 {
			t.Fatalf("send args = (%d,%d), want (255,0)", recipient, remove)
		}
		want := []byte{106, 0xcd, 0xab, 0x23, 0x04, 0x00, 0x00}
		if !reflect.DeepEqual(packet, want) {
			t.Fatalf("packet = % x, want % x", packet, want)
		}
		return -123
	})
	if got != -123 {
		t.Fatalf("result = %d, want send result -123", got)
	}
}

func TestPlayerStatus4174F0ServerMethodsWithoutHost(t *testing.T) {
	previous := noxflags.GetGame()
	noxflags.ResetGame()
	defer func() {
		noxflags.ResetGame()
		noxflags.SetGame(previous)
	}()

	s := new(Server)
	player := &Player{Field3680: 0x20}
	if got := s.NeedPlayerStatus4174F0(player, 0x400); got != 0 {
		t.Fatalf("need result = %d, want 0", got)
	}
	if player.Field3680 != 0x420 {
		t.Fatalf("needed flags = %#x, want 0x420", player.Field3680)
	}
	if got := s.UnsetPlayerStatus417530(player, 0x400, PlayerUnsetStatusRuntime417530{}); got != 0 {
		t.Fatalf("unset result = %d, want 0", got)
	}
	if player.Field3680 != 0x20 {
		t.Fatalf("unset flags = %#x, want 0x20", player.Field3680)
	}
}
