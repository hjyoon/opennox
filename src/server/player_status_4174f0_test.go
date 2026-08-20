package server

import (
	"fmt"
	"reflect"
	"testing"
)

type playerStatusTestPlayer4174F0 struct {
	name    string
	flags   uint32
	netCode uint16
}

type playerStatusTestWorld4174F0 struct {
	player       *playerStatusTestPlayer4174F0
	mask         uint32
	host         int32
	chat         int32
	anyPlayers   int32
	timerStatus  int32
	gameFlags    uint32
	modeEnabled  uint8
	setTimer     int32
	reportResult int32
	events       []string
	faultAt      int

	afterNetCode func(*playerStatusTestWorld4174F0, *playerStatusTestPlayer4174F0)
	afterFlags   func(*playerStatusTestWorld4174F0, *playerStatusTestPlayer4174F0)
}

func newPlayerStatusTestWorld4174F0() *playerStatusTestWorld4174F0 {
	return &playerStatusTestWorld4174F0{
		player:       &playerStatusTestPlayer4174F0{name: "player", flags: 0x40000020, netCode: 0xabcd},
		mask:         1,
		host:         1,
		anyPlayers:   1,
		gameFlags:    0x12345678,
		modeEnabled:  1,
		setTimer:     0x22222222,
		reportResult: 0x33333333,
	}
}

func (w *playerStatusTestWorld4174F0) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func playerStatusTestName4174F0(player *playerStatusTestPlayer4174F0) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func (w *playerStatusTestWorld4174F0) needHooks() playerNeedStatusHooks4174F0[*playerStatusTestPlayer4174F0] {
	return playerNeedStatusHooks4174F0[*playerStatusTestPlayer4174F0]{
		loadPlayerArg: func() *playerStatusTestPlayer4174F0 {
			w.record("player:" + playerStatusTestName4174F0(w.player))
			return w.player
		},
		loadMaskArg: func() uint32 {
			w.record(fmt.Sprintf("mask:%08x", w.mask))
			return w.mask
		},
		loadFlags: func(player *playerStatusTestPlayer4174F0) uint32 {
			w.record("flags:" + playerStatusTestName4174F0(player))
			return player.flags
		},
		storeFlags: func(player *playerStatusTestPlayer4174F0, flags uint32) {
			w.record(fmt.Sprintf("store:%s:%08x", playerStatusTestName4174F0(player), flags))
			player.flags = flags
		},
		gameFlag: func(flag uint32) int32 {
			w.record(fmt.Sprintf("game:%x", flag))
			if flag == playerStatusHostFlag4174F0 {
				return w.host
			}
			panic(fmt.Sprintf("unexpected game flag %x", flag))
		},
		reportStatus: func(player *playerStatusTestPlayer4174F0) int32 {
			w.record("report:" + playerStatusTestName4174F0(player))
			return w.reportResult
		},
	}
}

func (w *playerStatusTestWorld4174F0) unsetHooks() playerUnsetStatusHooks417530[*playerStatusTestPlayer4174F0] {
	return playerUnsetStatusHooks417530[*playerStatusTestPlayer4174F0]{
		loadMaskArg: func() uint32 {
			w.record(fmt.Sprintf("mask:%08x", w.mask))
			return w.mask
		},
		loadPlayerArg: func() *playerStatusTestPlayer4174F0 {
			w.record("player:" + playerStatusTestName4174F0(w.player))
			return w.player
		},
		loadFlags: func(player *playerStatusTestPlayer4174F0) uint32 {
			w.record("flags:" + playerStatusTestName4174F0(player))
			return player.flags
		},
		storeFlags: func(player *playerStatusTestPlayer4174F0, flags uint32) {
			w.record(fmt.Sprintf("store:%s:%08x", playerStatusTestName4174F0(player), flags))
			player.flags = flags
		},
		gameFlag: func(flag uint32) int32 {
			w.record(fmt.Sprintf("game:%x", flag))
			switch flag {
			case playerStatusHostFlag4174F0:
				return w.host
			case playerStatusChatFlag417530:
				return w.chat
			default:
				panic(fmt.Sprintf("unexpected game flag %x", flag))
			}
		},
		anyPlayers: func() int32 {
			w.record("any-players")
			return w.anyPlayers
		},
		timerStatus: func() int32 {
			w.record("timer-status")
			return w.timerStatus
		},
		gameFlagsValue: func() uint32 {
			w.record("game-flags")
			return w.gameFlags
		},
		modeEnabled: func(value int16) uint8 {
			w.record(fmt.Sprintf("mode:%04x", uint16(value)))
			return w.modeEnabled
		},
		startTimer: func() {
			w.record("start-timer")
		},
		setTimerStatus: func(value int32) int32 {
			w.record(fmt.Sprintf("set-timer:%d", value))
			return w.setTimer
		},
		reportStatus: func(player *playerStatusTestPlayer4174F0) int32 {
			w.record("report:" + playerStatusTestName4174F0(player))
			return w.reportResult
		},
	}
}

func playerNeedStatusFullEvents4174F0() []string {
	return []string{
		"player:player", "mask:00000001", "flags:player", "store:player:40000021",
		"game:1", "report:player",
	}
}

func playerUnsetStatusFullEvents417530() []string {
	return []string{
		"mask:00000001", "player:player", "flags:player", "store:player:40000020",
		"game:1", "game:80", "any-players", "timer-status", "game-flags", "mode:5678",
		"start-timer", "set-timer:1", "report:player",
	}
}

func TestPlayerNeedStatus4174F0OrderAndReturn(t *testing.T) {
	w := newPlayerStatusTestWorld4174F0()
	if got := playerNeedStatus4174F0(w.needHooks()); got != w.reportResult {
		t.Fatalf("result = %#x, want report result %#x", got, w.reportResult)
	}
	if want := playerNeedStatusFullEvents4174F0(); !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events =\n%q\nwant\n%q", w.events, want)
	}
	if w.player.flags != 0x40000021 {
		t.Fatalf("flags = %#x, want %#x", w.player.flags, uint32(0x40000021))
	}
}

func TestPlayerNeedStatus4174F0UnreportedReturnsHostResult(t *testing.T) {
	w := newPlayerStatusTestWorld4174F0()
	w.mask = 0x80000000
	w.host = 7
	if got := playerNeedStatus4174F0(w.needHooks()); got != 7 {
		t.Fatalf("result = %d, want host result 7", got)
	}
	if w.player.flags != 0xc0000020 {
		t.Fatalf("flags = %#x, want %#x", w.player.flags, uint32(0xc0000020))
	}
}

func TestPlayerUnsetStatus417530FullTransitionOrder(t *testing.T) {
	w := newPlayerStatusTestWorld4174F0()
	w.player.flags |= 1
	if got := playerUnsetStatus417530(w.unsetHooks()); got != w.reportResult {
		t.Fatalf("result = %#x, want report result %#x", got, w.reportResult)
	}
	if want := playerUnsetStatusFullEvents417530(); !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events =\n%q\nwant\n%q", w.events, want)
	}
	if w.player.flags != 0x40000020 {
		t.Fatalf("flags = %#x, want %#x", w.player.flags, uint32(0x40000020))
	}
}

func TestPlayerUnsetStatus417530ServiceGates(t *testing.T) {
	tests := []struct {
		name string
		edit func(*playerStatusTestWorld4174F0)
		want []string
	}{
		{
			name: "not host",
			edit: func(w *playerStatusTestWorld4174F0) { w.host = 0 },
			want: []string{"mask:00000001", "player:player", "flags:player", "store:player:40000020", "game:1"},
		},
		{
			name: "chat mode",
			edit: func(w *playerStatusTestWorld4174F0) { w.chat = 1 },
			want: []string{"mask:00000001", "player:player", "flags:player", "store:player:40000020", "game:1", "game:80", "report:player"},
		},
		{
			name: "no players",
			edit: func(w *playerStatusTestWorld4174F0) { w.anyPlayers = 0 },
			want: []string{"mask:00000001", "player:player", "flags:player", "store:player:40000020", "game:1", "game:80", "any-players", "report:player"},
		},
		{
			name: "timer already active",
			edit: func(w *playerStatusTestWorld4174F0) { w.timerStatus = 1 },
			want: []string{"mask:00000001", "player:player", "flags:player", "store:player:40000020", "game:1", "game:80", "any-players", "timer-status", "report:player"},
		},
		{
			name: "mode disabled",
			edit: func(w *playerStatusTestWorld4174F0) { w.modeEnabled = 0 },
			want: []string{"mask:00000001", "player:player", "flags:player", "store:player:40000020", "game:1", "game:80", "any-players", "timer-status", "game-flags", "mode:5678", "report:player"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := newPlayerStatusTestWorld4174F0()
			w.player.flags |= 1
			test.edit(w)
			playerUnsetStatus417530(w.unsetHooks())
			if !reflect.DeepEqual(w.events, test.want) {
				t.Fatalf("events =\n%q\nwant\n%q", w.events, test.want)
			}
		})
	}
}

func TestPlayerStatus4174F0AllFaultPrefixes(t *testing.T) {
	tests := []struct {
		name string
		want func() []string
		run  func(*playerStatusTestWorld4174F0)
	}{
		{"need", playerNeedStatusFullEvents4174F0, func(w *playerStatusTestWorld4174F0) { playerNeedStatus4174F0(w.needHooks()) }},
		{"unset", playerUnsetStatusFullEvents417530, func(w *playerStatusTestWorld4174F0) { w.player.flags |= 1; playerUnsetStatus417530(w.unsetHooks()) }},
	}
	for _, test := range tests {
		want := test.want()
		for faultAt := 1; faultAt <= len(want); faultAt++ {
			t.Run(fmt.Sprintf("%s-%02d", test.name, faultAt), func(t *testing.T) {
				w := newPlayerStatusTestWorld4174F0()
				w.faultAt = faultAt
				defer func() {
					if got := recover(); got != want[faultAt-1] {
						t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
					}
					if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
						t.Fatalf("events = %q, want %q", w.events, prefix)
					}
				}()
				test.run(w)
			})
		}
	}
}

func TestPlayerStatusReport417630CachesReadsAndPacket(t *testing.T) {
	w := newPlayerStatusTestWorld4174F0()
	w.afterNetCode = func(_ *playerStatusTestWorld4174F0, player *playerStatusTestPlayer4174F0) {
		player.netCode = 0x1111
		player.flags = 0xffffffff
	}
	w.afterFlags = func(_ *playerStatusTestWorld4174F0, player *playerStatusTestPlayer4174F0) {
		player.flags = 0
	}
	got := playerStatusReport417630(playerStatusReportHooks417630[*playerStatusTestPlayer4174F0]{
		loadPlayerArg: func() *playerStatusTestPlayer4174F0 {
			w.record("player:player")
			return w.player
		},
		loadNetCode: func(player *playerStatusTestPlayer4174F0) uint16 {
			w.record("net-code")
			value := player.netCode
			w.afterNetCode(w, player)
			return value
		},
		loadFlags: func(player *playerStatusTestPlayer4174F0) uint32 {
			w.record("flags")
			value := player.flags
			w.afterFlags(w, player)
			return value
		},
		send: func(recipient byte, packet []byte, remove int32) int32 {
			w.record("send")
			if recipient != 255 || remove != 0 {
				t.Fatalf("send args = (%d,%d), want (255,0)", recipient, remove)
			}
			want := []byte{106, 0xcd, 0xab, 0x23, 0x04, 0x00, 0x00}
			if !reflect.DeepEqual(packet, want) {
				t.Fatalf("packet = % x, want % x", packet, want)
			}
			return 0x12345678
		},
	})
	if got != 0x12345678 {
		t.Fatalf("result = %#x, want send result", got)
	}
	if want := []string{"player:player", "net-code", "flags", "send"}; !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}
