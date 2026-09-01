package opennox

import (
	"fmt"
	"reflect"
	"testing"
)

type playerPhonemeBroadcastTestObject4FC960 struct {
	name string
}

type playerPhonemeBroadcastTestEnv4FC960 struct {
	first  *playerPhonemeBroadcastTestObject4FC960
	next   map[*playerPhonemeBroadcastTestObject4FC960]*playerPhonemeBroadcastTestObject4FC960
	codes  map[*playerPhonemeBroadcastTestObject4FC960]uint32
	events []string
}

func playerPhonemeBroadcastTestName4FC960(obj *playerPhonemeBroadcastTestObject4FC960) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func playerPhonemeBroadcastTestHooks4FC960(env *playerPhonemeBroadcastTestEnv4FC960) playerPhonemeBroadcastHooks4FC960[*playerPhonemeBroadcastTestObject4FC960] {
	record := func(format string, args ...any) {
		env.events = append(env.events, fmt.Sprintf(format, args...))
	}
	return playerPhonemeBroadcastHooks4FC960[*playerPhonemeBroadcastTestObject4FC960]{
		firstUnit: func() *playerPhonemeBroadcastTestObject4FC960 {
			record("first:%s", playerPhonemeBroadcastTestName4FC960(env.first))
			return env.first
		},
		nextUnit: func(unit *playerPhonemeBroadcastTestObject4FC960) *playerPhonemeBroadcastTestObject4FC960 {
			next := env.next[unit]
			record("next:%s:%s", playerPhonemeBroadcastTestName4FC960(unit), playerPhonemeBroadcastTestName4FC960(next))
			return next
		},
		loadNetCode: func(unit *playerPhonemeBroadcastTestObject4FC960) uint32 {
			code := env.codes[unit]
			record("net-code:%s:%#x", playerPhonemeBroadcastTestName4FC960(unit), code)
			return code
		},
		spellGetPhoneme: func(code uint32, phoneme int8) int32 {
			record("phoneme:%#x:%d", code, phoneme)
			return -123456789
		},
		audioEvent: func(soundID int32, source *playerPhonemeBroadcastTestObject4FC960, kind int32, listener uint32) {
			record("audio:%d:%s:%d:%#x", soundID, playerPhonemeBroadcastTestName4FC960(source), kind, listener)
		},
	}
}

func TestPlayerPhonemeBroadcast4FC960EmptyListDoesNotObserveArguments(t *testing.T) {
	source := &playerPhonemeBroadcastTestObject4FC960{name: "source"}
	env := &playerPhonemeBroadcastTestEnv4FC960{
		codes: map[*playerPhonemeBroadcastTestObject4FC960]uint32{source: 0x89abcdef},
	}

	got := playerPhonemeBroadcast4FC960(source, -128, playerPhonemeBroadcastTestHooks4FC960(env))
	if got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{"first:nil"}
	if !reflect.DeepEqual(env.events, want) {
		t.Fatalf("events:\n got %v\nwant %v", env.events, want)
	}
}

func TestPlayerPhonemeBroadcast4FC960SkipsSourceAndUsesLiveNext(t *testing.T) {
	source := &playerPhonemeBroadcastTestObject4FC960{name: "source"}
	firstListener := &playerPhonemeBroadcastTestObject4FC960{name: "first-listener"}
	staleListener := &playerPhonemeBroadcastTestObject4FC960{name: "stale-listener"}
	liveListener := &playerPhonemeBroadcastTestObject4FC960{name: "live-listener"}
	env := &playerPhonemeBroadcastTestEnv4FC960{
		first: source,
		next: map[*playerPhonemeBroadcastTestObject4FC960]*playerPhonemeBroadcastTestObject4FC960{
			source:        firstListener,
			firstListener: staleListener,
			staleListener: nil,
			liveListener:  nil,
		},
		codes: map[*playerPhonemeBroadcastTestObject4FC960]uint32{
			source:        0xfedcba98,
			firstListener: 0x89abcdef,
			staleListener: 0x11111111,
			liveListener:  0x76543210,
		},
	}
	hooks := playerPhonemeBroadcastTestHooks4FC960(env)
	originalAudio := hooks.audioEvent
	hooks.audioEvent = func(soundID int32, gotSource *playerPhonemeBroadcastTestObject4FC960, kind int32, listener uint32) {
		originalAudio(soundID, gotSource, kind, listener)
		if listener == env.codes[firstListener] {
			env.next[firstListener] = liveListener
			env.codes[source] = 0x80000001
		}
	}

	got := playerPhonemeBroadcast4FC960(source, -128, hooks)
	if got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{
		"first:source",
		"next:source:first-listener",
		"net-code:first-listener:0x89abcdef",
		"net-code:source:0xfedcba98",
		"phoneme:0xfedcba98:-128",
		"audio:-123456789:source:2:0x89abcdef",
		"next:first-listener:live-listener",
		"net-code:live-listener:0x76543210",
		"net-code:source:0x80000001",
		"phoneme:0x80000001:-128",
		"audio:-123456789:source:2:0x76543210",
		"next:live-listener:nil",
	}
	if !reflect.DeepEqual(env.events, want) {
		t.Fatalf("events:\n got %v\nwant %v", env.events, want)
	}
}

func TestPlayerPhonemeBroadcast4FC960FaultPrefix(t *testing.T) {
	source := &playerPhonemeBroadcastTestObject4FC960{name: "source"}
	listener := &playerPhonemeBroadcastTestObject4FC960{name: "listener"}
	env := &playerPhonemeBroadcastTestEnv4FC960{
		first: listener,
		next:  map[*playerPhonemeBroadcastTestObject4FC960]*playerPhonemeBroadcastTestObject4FC960{listener: nil},
		codes: map[*playerPhonemeBroadcastTestObject4FC960]uint32{
			listener: 0x11111111,
			source:   0x22222222,
		},
	}
	hooks := playerPhonemeBroadcastTestHooks4FC960(env)
	originalLoad := hooks.loadNetCode
	hooks.loadNetCode = func(unit *playerPhonemeBroadcastTestObject4FC960) uint32 {
		code := originalLoad(unit)
		if unit == source {
			panic("source-net-code fault")
		}
		return code
	}

	func() {
		defer func() {
			if got := recover(); got != "source-net-code fault" {
				t.Fatalf("panic = %v, want source-net-code fault", got)
			}
		}()
		playerPhonemeBroadcast4FC960(source, 8, hooks)
	}()

	want := []string{
		"first:listener",
		"net-code:listener:0x11111111",
		"net-code:source:0x22222222",
	}
	if !reflect.DeepEqual(env.events, want) {
		t.Fatalf("events:\n got %v\nwant %v", env.events, want)
	}
}
