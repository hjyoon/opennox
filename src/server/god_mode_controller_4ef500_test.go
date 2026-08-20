package server

import (
	"fmt"
	"reflect"
	"testing"
)

type godModeControllerWorld4EF500 struct {
	coop   int32
	value  uint32
	flags  uint32
	first  int
	next   map[int]int
	events []string
	fault  int
}

func (w *godModeControllerWorld4EF500) record(event string) {
	w.events = append(w.events, event)
	if w.fault != 0 && len(w.events) == w.fault {
		panic(event)
	}
}

func (w *godModeControllerWorld4EF500) hooks() godModeControllerHooks4EF500[int] {
	return godModeControllerHooks4EF500[int]{
		gameFlag: func(flag uint32) int32 {
			w.record(fmt.Sprintf("game:%08x", flag))
			return w.coop
		},
		loadValue: func() uint32 {
			w.record(fmt.Sprintf("value:%08x", w.value))
			return w.value
		},
		loadEngineFlags: func() uint32 {
			w.record(fmt.Sprintf("flags:%08x", w.flags))
			return w.flags
		},
		storeEngineFlags: func(flags uint32) {
			w.record(fmt.Sprintf("store:%08x", flags))
			w.flags = flags
		},
		firstPlayer: func() int {
			w.record(fmt.Sprintf("first:%d", w.first))
			return w.first
		},
		awardScrolls: func(player int) {
			w.record(fmt.Sprintf("scrolls:%d", player))
		},
		awardSpells: func(player int) {
			w.record(fmt.Sprintf("spells:%d", player))
		},
		awardAbilities: func(player int) {
			w.record(fmt.Sprintf("abilities:%d", player))
		},
		nextPlayer: func(player int) int {
			next := w.next[player]
			w.record(fmt.Sprintf("next:%d=%d", player, next))
			return next
		},
	}
}

func TestGodModeController4EF500CoopGateIsFirstAndQuiet(t *testing.T) {
	w := &godModeControllerWorld4EF500{
		value: 0xffffffff,
		flags: 0xa5a5a5a5,
		first: 1,
		next:  map[int]int{1: 0},
	}
	godModeController4EF500(w.hooks())
	want := []string{"game:00000800"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	if w.flags != 0xa5a5a5a5 {
		t.Fatalf("flags = %#08x, want unchanged", w.flags)
	}
}

func TestGodModeController4EF500OnlyWholeDwordOneEnables(t *testing.T) {
	for _, value := range []uint32{0, 1, 2, 0x00000101, 0x80000001, 0xffffffff} {
		t.Run(fmt.Sprintf("%08x", value), func(t *testing.T) {
			const initial = uint32(0xa5a5a5cf)
			w := &godModeControllerWorld4EF500{
				coop:  7,
				value: value,
				flags: initial,
			}
			godModeController4EF500(w.hooks())

			wantFlags := initial &^ godModeControllerEngineMask
			if value == godModeControllerEnable4EF500 {
				wantFlags = initial | godModeControllerEngineMask
			}
			if w.flags != wantFlags {
				t.Fatalf("flags = %#08x, want %#08x", w.flags, wantFlags)
			}
			wantEvents := []string{
				"game:00000800",
				fmt.Sprintf("value:%08x", value),
				fmt.Sprintf("flags:%08x", initial),
				fmt.Sprintf("store:%08x", wantFlags),
				"first:0",
			}
			if !reflect.DeepEqual(w.events, wantEvents) {
				t.Fatalf("events = %v, want %v", w.events, wantEvents)
			}
		})
	}
}

func TestGodModeController4EF500LivePlayerWalkOrder(t *testing.T) {
	w := &godModeControllerWorld4EF500{
		coop:  1,
		value: 1,
		flags: 0x80000004,
		first: 11,
		next:  map[int]int{11: 22, 22: 0},
	}
	godModeController4EF500(w.hooks())
	want := []string{
		"game:00000800",
		"value:00000001",
		"flags:80000004",
		"store:80000034",
		"first:11",
		"scrolls:11",
		"spells:11",
		"abilities:11",
		"next:11=22",
		"scrolls:22",
		"spells:22",
		"abilities:22",
		"next:22=0",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestGodModeController4EF500SuccessorIsReadAfterCallbacks(t *testing.T) {
	w := &godModeControllerWorld4EF500{
		coop:  1,
		value: 1,
		first: 1,
		next:  map[int]int{1: 2, 2: 0, 3: 0},
	}
	hooks := w.hooks()
	hooks.awardAbilities = func(player int) {
		w.record(fmt.Sprintf("abilities:%d", player))
		if player == 1 {
			w.next[1] = 3
		}
	}
	godModeController4EF500(hooks)
	want := []string{
		"game:00000800",
		"value:00000001",
		"flags:00000000",
		"store:00000030",
		"first:1",
		"scrolls:1",
		"spells:1",
		"abilities:1",
		"next:1=3",
		"scrolls:3",
		"spells:3",
		"abilities:3",
		"next:3=0",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestGodModeController4EF500EveryObservationFaultPrefix(t *testing.T) {
	base := func() *godModeControllerWorld4EF500 {
		return &godModeControllerWorld4EF500{
			coop:  1,
			value: 1,
			flags: 0x4,
			first: 1,
			next:  map[int]int{1: 2, 2: 0},
		}
	}
	w := base()
	godModeController4EF500(w.hooks())
	want := append([]string(nil), w.events...)

	for fault := 1; fault <= len(want); fault++ {
		t.Run(fmt.Sprintf("event-%d", fault), func(t *testing.T) {
			w := base()
			w.fault = fault
			defer func() {
				if got := recover(); got != want[fault-1] {
					t.Fatalf("panic = %v, want %q", got, want[fault-1])
				}
				if prefix := want[:fault]; !reflect.DeepEqual(w.events, prefix) {
					t.Fatalf("events = %v, want %v", w.events, prefix)
				}
			}()
			godModeController4EF500(w.hooks())
		})
	}
}

type godModeCommandWorld4EF500 struct {
	quest  bool
	events []string
	fault  int
}

func (w *godModeCommandWorld4EF500) record(event string) {
	w.events = append(w.events, event)
	if w.fault != 0 && len(w.events) == w.fault {
		panic(event)
	}
}

func (w *godModeCommandWorld4EF500) runtime() GodModeCommandRuntime4EF500 {
	return GodModeCommandRuntime4EF500{
		QuestMode: func() bool {
			w.record(fmt.Sprintf("quest:%t", w.quest))
			return w.quest
		},
		SetGod: func(value uint32) {
			w.record(fmt.Sprintf("set:%08x", value))
		},
		LoadString: func(key string) string {
			w.record("load:" + key)
			return "message:" + key
		},
		Print: func(message string) {
			w.record("print:" + message)
		},
	}
}

func TestGodModeCommand4EF500SetAndQuestOrder(t *testing.T) {
	w := &godModeCommandWorld4EF500{}
	GodModeCommand4EF500(true, w.runtime())
	want := []string{
		"quest:false",
		"set:00000001",
		"load:godset",
		"print:message:godset",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}

	quiet := &godModeCommandWorld4EF500{quest: true}
	GodModeCommand4EF500(true, quiet.runtime())
	if want := []string{"quest:true"}; !reflect.DeepEqual(quiet.events, want) {
		t.Fatalf("Quest events = %v, want %v", quiet.events, want)
	}
}

func TestGodModeCommand4EF500UnsetSkipsQuest(t *testing.T) {
	w := &godModeCommandWorld4EF500{quest: true}
	runtime := w.runtime()
	runtime.QuestMode = func() bool {
		t.Fatal("unset queried Quest mode")
		return true
	}
	GodModeCommand4EF500(false, runtime)
	want := []string{
		"set:00000000",
		"load:godunset",
		"print:message:godunset",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestGodModeCommand4EF500FaultPrefixes(t *testing.T) {
	tests := []struct {
		name   string
		enable bool
		want   []string
	}{
		{
			name:   "set",
			enable: true,
			want: []string{
				"quest:false",
				"set:00000001",
				"load:godset",
				"print:message:godset",
			},
		},
		{
			name:   "unset",
			enable: false,
			want: []string{
				"set:00000000",
				"load:godunset",
				"print:message:godunset",
			},
		},
	}

	for _, test := range tests {
		for fault := 1; fault <= len(test.want); fault++ {
			t.Run(fmt.Sprintf("%s/event-%d", test.name, fault), func(t *testing.T) {
				w := &godModeCommandWorld4EF500{fault: fault}
				defer func() {
					if got := recover(); got != test.want[fault-1] {
						t.Fatalf("panic = %v, want %q", got, test.want[fault-1])
					}
					if prefix := test.want[:fault]; !reflect.DeepEqual(w.events, prefix) {
						t.Fatalf("events = %v, want %v", w.events, prefix)
					}
				}()
				GodModeCommand4EF500(test.enable, w.runtime())
			})
		}
	}
}
