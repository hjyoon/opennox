package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestNetworkGauntletRespawnOrderAndReload51BAD0(t *testing.T) {
	events := make([]string, 0, 7)
	loads := 0
	hooks := networkGauntletHooks51BAD0[string, string, string]{
		loadSubtype: func() uint8 {
			events = append(events, "subtype")
			return networkGauntletRespawn51BAD0
		},
		loadPlayer: func(update string) string {
			events = append(events, "player:"+update)
			return "player"
		},
		loadPlayerUnit: func(player string) string {
			loads++
			events = append(events, "unit:"+player)
			if loads == 1 {
				return "first"
			}
			return "second"
		},
		loadFlags: func(unit string) uint32 {
			events = append(events, "flags:"+unit)
			return networkGauntletDeadFlag51BAD0
		},
		clearField137: func(update string) {
			events = append(events, "clear:"+update)
		},
		respawn: func(unit string) {
			events = append(events, "respawn:"+unit)
		},
		exit: func(string) { t.Fatal("respawn subtype called exit") },
	}
	if got := networkGauntlet51BAD0("packet-unit", "update", hooks); got != 2 {
		t.Fatalf("consumed = %d, want 2", got)
	}
	want := []string{
		"subtype", "player:update", "unit:player", "flags:first",
		"clear:update", "unit:player", "respawn:second",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestNetworkGauntletRespawnGuards51BAD0(t *testing.T) {
	tests := []struct {
		name  string
		unit  string
		flags uint32
		want  []string
	}{
		{name: "missing unit", want: []string{"subtype", "player", "unit"}},
		{name: "not dead", unit: "live", flags: 0x4000, want: []string{"subtype", "player", "unit", "flags"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			events := make([]string, 0, 4)
			hooks := networkGauntletHooks51BAD0[string, int, int]{
				loadSubtype:    func() uint8 { events = append(events, "subtype"); return 3 },
				loadPlayer:     func(int) int { events = append(events, "player"); return 1 },
				loadPlayerUnit: func(int) string { events = append(events, "unit"); return tc.unit },
				loadFlags:      func(string) uint32 { events = append(events, "flags"); return tc.flags },
				clearField137:  func(int) { t.Fatal("guard cleared Field137") },
				respawn:        func(string) { t.Fatal("guard respawned") },
				exit:           func(string) { t.Fatal("guard exited") },
			}
			if got := networkGauntlet51BAD0("packet-unit", 1, hooks); got != 2 {
				t.Fatalf("consumed = %d, want 2", got)
			}
			if !reflect.DeepEqual(events, tc.want) {
				t.Fatalf("events = %v, want %v", events, tc.want)
			}
		})
	}
}

func TestNetworkGauntletExitAndInvalid51BAD0(t *testing.T) {
	for _, tc := range []struct {
		name    string
		subtype uint8
		want    int32
		exits   int
	}{
		{name: "exit", subtype: networkGauntletExit51BAD0, want: 2, exits: 1},
		{name: "invalid", subtype: 26, want: -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			exits := 0
			hooks := networkGauntletHooks51BAD0[string, int, int]{
				loadSubtype: func() uint8 { return tc.subtype },
				loadPlayer:  func(int) int { t.Fatal("non-respawn subtype loaded player"); return 0 },
				loadPlayerUnit: func(int) string {
					t.Fatal("non-respawn subtype loaded unit")
					return ""
				},
				loadFlags:     func(string) uint32 { t.Fatal("non-respawn subtype loaded flags"); return 0 },
				clearField137: func(int) { t.Fatal("non-respawn subtype cleared Field137") },
				respawn:       func(string) { t.Fatal("non-respawn subtype respawned") },
				exit: func(unit string) {
					exits++
					if unit != "packet-unit" {
						t.Fatalf("exit unit = %q", unit)
					}
				},
			}
			if got := networkGauntlet51BAD0("packet-unit", 1, hooks); got != tc.want {
				t.Fatalf("result = %d, want %d", got, tc.want)
			}
			if exits != tc.exits {
				t.Fatalf("exit calls = %d, want %d", exits, tc.exits)
			}
		})
	}
}

func TestNetworkGauntletNativePointers51BAD0(t *testing.T) {
	first := &Object{ObjFlags: object.FlagDead}
	player := &Player{PlayerUnit: first}
	update := &PlayerUpdateData{Player: player, Field137: 0x89abcdef}
	packetUnit := &Object{UpdateData: unsafe.Pointer(update)}
	packet := &[NetworkGauntletPacketSize51BAD0]byte{0xf0, networkGauntletRespawn51BAD0}

	respawns := 0
	exits := 0
	got := (&Server{}).NetworkGauntlet51BAD0(
		packetUnit,
		update,
		packet,
		NetworkGauntletRuntime51BAD0{
			Respawn: func(unit *Object) {
				respawns++
				if unit != first {
					t.Fatalf("respawn unit = %p, want %p", unit, first)
				}
			},
			Exit: func(*Object) { exits++ },
		},
	)
	if got != 2 || update.Field137 != 0 || respawns != 1 || exits != 0 {
		t.Fatalf("result = (%d, field %#x, respawns %d, exits %d), want (2,0,1,0)", got, update.Field137, respawns, exits)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]uintptr{
			"packet unit":  uintptr(unsafe.Pointer(packetUnit)),
			"update":       uintptr(unsafe.Pointer(update)),
			"player":       uintptr(unsafe.Pointer(player)),
			"respawn unit": uintptr(unsafe.Pointer(first)),
		} {
			if pointer <= uintptr(^uint32(0)) {
				t.Fatalf("%s address %#x did not exercise the high native half", name, pointer)
			}
		}
	}
}

func TestNetworkGauntletNativeExitAndInvalid51BAD0(t *testing.T) {
	unit := &Object{}
	update := &PlayerUpdateData{}
	for _, tc := range []struct {
		name    string
		subtype uint8
		want    int32
		exits   int
	}{
		{name: "exit", subtype: networkGauntletExit51BAD0, want: 2, exits: 1},
		{name: "invalid", subtype: 0xff, want: -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			exits := 0
			packet := &[NetworkGauntletPacketSize51BAD0]byte{0xf0, tc.subtype}
			got := (&Server{}).NetworkGauntlet51BAD0(unit, update, packet, NetworkGauntletRuntime51BAD0{
				Respawn: func(*Object) { t.Fatal("non-respawn subtype respawned") },
				Exit: func(got *Object) {
					exits++
					if got != unit {
						t.Fatalf("exit unit = %p, want %p", got, unit)
					}
				},
			})
			if got != tc.want || exits != tc.exits {
				t.Fatalf("result/exits = %d/%d, want %d/%d", got, exits, tc.want, tc.exits)
			}
		})
	}
}
