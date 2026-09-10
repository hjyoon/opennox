package server

import (
	"fmt"
	"reflect"
	"testing"
)

type audioEventZoneObject501C00 struct{}
type audioEventZonePosition501C00 struct{}

type audioEventZoneWorld501C00 struct {
	events           []string
	faultAt          int
	class            uint8
	playerZone       uint8
	monsterPolygon   uint64
	monsterPolygonID uint32
	fallbackPolygon  uint64
	polygonZones     map[uint64]uint8
	positionX        float32
	positionY        float32
}

func (w *audioEventZoneWorld501C00) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic("injected fault")
	}
}

func (w *audioEventZoneWorld501C00) hooks() audioEventZoneHooks501C00[
	*audioEventZonePosition501C00,
	*audioEventZoneObject501C00,
	uint64,
	uint64,
	uint64,
] {
	return audioEventZoneHooks501C00[
		*audioEventZonePosition501C00,
		*audioEventZoneObject501C00,
		uint64,
		uint64,
		uint64,
	]{
		loadClassLow: func(*audioEventZoneObject501C00) uint8 {
			w.event("class")
			return w.class
		},
		loadUpdate: func(*audioEventZoneObject501C00) uint64 {
			w.event("update")
			return 0x100000001
		},
		loadPlayer: func(update uint64) uint64 {
			w.event(fmt.Sprintf("player:%#x", update))
			return 0x200000002
		},
		loadPlayerZone: func(player uint64) uint8 {
			w.event(fmt.Sprintf("player-zone:%#x", player))
			return w.playerZone
		},
		loadMonsterPolygonID: func(update uint64) uint32 {
			w.event(fmt.Sprintf("monster-polygon-id:%#x", update))
			return w.monsterPolygonID
		},
		polygonByID: func(id uint32) uint64 {
			w.event(fmt.Sprintf("polygon-by-id:%#x", id))
			return w.monsterPolygon
		},
		loadPolygonZone: func(polygon uint64) uint8 {
			w.event(fmt.Sprintf("polygon-zone:%#x", polygon))
			return w.polygonZones[polygon]
		},
		loadPositionX: func(*audioEventZonePosition501C00) float32 {
			w.event("position-x")
			return w.positionX
		},
		floatToInt: func(value float32) int32 {
			w.event(fmt.Sprintf("float-to-int:%g", value))
			return int32(value)
		},
		loadPositionY: func(*audioEventZonePosition501C00) float32 {
			w.event("position-y")
			return w.positionY
		},
		polygonAtPoint: func(point [2]int32, previous uint32) uint64 {
			w.event(fmt.Sprintf("polygon-at-point:%d,%d:%#x", point[0], point[1], previous))
			return w.fallbackPolygon
		},
	}
}

func (w *audioEventZoneWorld501C00) call(object *audioEventZoneObject501C00) uint8 {
	return audioEventZone501C00(&audioEventZonePosition501C00{}, object, w.hooks())
}

func TestAudioEventZone501C00PlayerWinsAndShortCircuitsPosition(t *testing.T) {
	world := &audioEventZoneWorld501C00{
		class:      audioEventZonePlayerClass501C00 | audioEventZoneMonsterClass501C00,
		playerZone: 0xe7,
	}
	got := world.call(&audioEventZoneObject501C00{})
	wantEvents := []string{"class", "update", "player:0x100000001", "player-zone:0x200000002"}
	if got != 0xe7 || !reflect.DeepEqual(world.events, wantEvents) {
		t.Fatalf("result/events = %#x/%v, want 0xe7/%v", got, world.events, wantEvents)
	}
}

func TestAudioEventZone501C00PlayerZeroFallsBackInExactOrder(t *testing.T) {
	world := &audioEventZoneWorld501C00{
		class:           audioEventZonePlayerClass501C00,
		fallbackPolygon: 0x400000004,
		polygonZones:    map[uint64]uint8{0x400000004: 0x91},
		positionX:       17,
		positionY:       -23,
	}
	got := world.call(&audioEventZoneObject501C00{})
	wantEvents := []string{
		"class", "update", "player:0x100000001", "player-zone:0x200000002",
		"position-x", "float-to-int:17", "position-y", "float-to-int:-23",
		"polygon-at-point:17,-23:0x0", "polygon-zone:0x400000004",
	}
	if got != 0x91 || !reflect.DeepEqual(world.events, wantEvents) {
		t.Fatalf("result/events = %#x/%v, want 0x91/%v", got, world.events, wantEvents)
	}
}

func TestAudioEventZone501C00MonsterBranches(t *testing.T) {
	t.Run("nonzero zone short circuits", func(t *testing.T) {
		world := &audioEventZoneWorld501C00{
			class:            audioEventZoneMonsterClass501C00,
			monsterPolygonID: 0x7f,
			monsterPolygon:   0x300000003,
			polygonZones:     map[uint64]uint8{0x300000003: 0xa5},
		}
		got := world.call(&audioEventZoneObject501C00{})
		wantEvents := []string{
			"class", "update", "monster-polygon-id:0x100000001",
			"polygon-by-id:0x7f", "polygon-zone:0x300000003",
		}
		if got != 0xa5 || !reflect.DeepEqual(world.events, wantEvents) {
			t.Fatalf("result/events = %#x/%v, want 0xa5/%v", got, world.events, wantEvents)
		}
	})

	t.Run("missing polygon skips zone load", func(t *testing.T) {
		world := &audioEventZoneWorld501C00{
			class:            audioEventZoneMonsterClass501C00,
			monsterPolygonID: 0xdeadface,
			positionX:        3,
			positionY:        4,
		}
		got := world.call(&audioEventZoneObject501C00{})
		wantEvents := []string{
			"class", "update", "monster-polygon-id:0x100000001",
			"polygon-by-id:0xdeadface", "position-x", "float-to-int:3",
			"position-y", "float-to-int:4", "polygon-at-point:3,4:0x0",
		}
		if got != 0 || !reflect.DeepEqual(world.events, wantEvents) {
			t.Fatalf("result/events = %#x/%v, want 0/%v", got, world.events, wantEvents)
		}
	})
}

func TestAudioEventZone501C00NilObjectStartsAtPosition(t *testing.T) {
	world := &audioEventZoneWorld501C00{positionX: -5, positionY: 6}
	got := world.call(nil)
	wantEvents := []string{
		"position-x", "float-to-int:-5", "position-y", "float-to-int:6",
		"polygon-at-point:-5,6:0x0",
	}
	if got != 0 || !reflect.DeepEqual(world.events, wantEvents) {
		t.Fatalf("result/events = %#x/%v, want 0/%v", got, world.events, wantEvents)
	}
}

func TestAudioEventZone501C00EveryObservableFaultPrefix(t *testing.T) {
	tests := []struct {
		name  string
		world func() *audioEventZoneWorld501C00
	}{
		{
			name: "player zero fallback",
			world: func() *audioEventZoneWorld501C00 {
				return &audioEventZoneWorld501C00{
					class:           audioEventZonePlayerClass501C00,
					fallbackPolygon: 0x400000004,
					polygonZones:    map[uint64]uint8{0x400000004: 0x66},
					positionX:       11,
					positionY:       12,
				}
			},
		},
		{
			name: "monster zero fallback",
			world: func() *audioEventZoneWorld501C00 {
				return &audioEventZoneWorld501C00{
					class:            audioEventZoneMonsterClass501C00,
					monsterPolygonID: 3,
					monsterPolygon:   0x300000003,
					fallbackPolygon:  0x400000004,
					polygonZones: map[uint64]uint8{
						0x300000003: 0,
						0x400000004: 0x77,
					},
					positionX: 13,
					positionY: 14,
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			base := test.world()
			base.call(&audioEventZoneObject501C00{})
			want := append([]string(nil), base.events...)
			for faultAt := 1; faultAt <= len(want); faultAt++ {
				t.Run(fmt.Sprintf("event_%02d", faultAt), func(t *testing.T) {
					world := test.world()
					world.faultAt = faultAt
					func() {
						defer func() {
							if recover() == nil {
								t.Fatal("expected injected fault")
							}
						}()
						world.call(&audioEventZoneObject501C00{})
					}()
					if !reflect.DeepEqual(world.events, want[:faultAt]) {
						t.Fatalf("events = %v, want prefix %v", world.events, want[:faultAt])
					}
				})
			}
		})
	}
}
