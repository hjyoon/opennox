package server

import (
	"fmt"
	"reflect"
	"testing"
)

const (
	remoteAudioUnit501CA0       = uint64(0x100000001)
	remoteAudioUpdate501CA0     = uint64(0x200000002)
	remoteAudioPlayer501CA0     = uint64(0x300000003)
	remoteAudioPlayer2_501CA0   = uint64(0x300000004)
	remoteAudioFollow501CA0     = uint64(0x400000004)
	remoteAudioFollow2_501CA0   = uint64(0x400000005)
	remoteAudioFollowUpd501CA0  = uint64(0x500000005)
	remoteAudioFollowPlr501CA0  = uint64(0x600000006)
	remoteAudioPolygon501CA0    = uint64(0x700000007)
	remoteAudioTeam501CA0       = uint64(0x800000008)
	remoteAudioEvent501CA0      = uint64(0x900000009)
	remoteAudioEvent2_501CA0    = uint64(0x90000000a)
	remoteAudioPosition501CA0   = uint64(0xa0000000a)
	remoteAudioPosition2_501CA0 = uint64(0xa0000000b)
	remoteAudioPlayerPos501CA0  = uint64(0xb0000000b)
)

type remotePlayerAudioWorld501CA0 struct {
	events  []string
	faultAt int

	updates            map[uint64]uint64
	players            map[uint64]uint64
	playerFlags        map[uint64]uint8
	cameraTargets      map[uint64]uint64
	classes            map[uint64]uint8
	positionX          map[uint64]float32
	positionY          map[uint64]float32
	currentPolygon     map[uint64]uint32
	playerZones        map[uint64]uint8
	polygon            uint64
	polygonZone        map[uint64]uint8
	hasTeamValue       bool
	teamIDs            map[uint64]uint32
	teams              map[uint32]uint64
	head               uint64
	eventNext          map[uint64]uint64
	eventKind          map[uint64]int32
	eventCode          map[uint64]uint32
	netCode            map[uint64]uint32
	eventObject        map[uint64]uint64
	eventPositions     map[uint64]uint64
	eventZones         map[uint64]uint8
	phonemeState       map[uint64]uint8
	eventSounds        map[uint64]int32
	playerPositions    map[uint64]uint64
	fadeValue          int32
	soundField20       map[int32]int32
	questMutation      func()
	firstFloatMutation func()
	addMutation        func(uint64, int32)
	directMutation     func(uint64, uint64, int32)
	floatCalls         int
}

func newRemotePlayerAudioWorld501CA0() *remotePlayerAudioWorld501CA0 {
	return &remotePlayerAudioWorld501CA0{
		updates:        map[uint64]uint64{remoteAudioUnit501CA0: remoteAudioUpdate501CA0},
		players:        map[uint64]uint64{remoteAudioUpdate501CA0: remoteAudioPlayer501CA0},
		playerFlags:    make(map[uint64]uint8),
		cameraTargets:  make(map[uint64]uint64),
		classes:        make(map[uint64]uint8),
		positionX:      make(map[uint64]float32),
		positionY:      make(map[uint64]float32),
		currentPolygon: map[uint64]uint32{remoteAudioPlayer501CA0: 1},
		playerZones:    map[uint64]uint8{remoteAudioPlayer501CA0: 7},
		polygonZone:    make(map[uint64]uint8),
		teamIDs:        make(map[uint64]uint32),
		teams:          make(map[uint32]uint64),
		eventNext:      make(map[uint64]uint64),
		eventKind:      make(map[uint64]int32),
		eventCode:      make(map[uint64]uint32),
		netCode:        make(map[uint64]uint32),
		eventObject:    make(map[uint64]uint64),
		eventPositions: make(map[uint64]uint64),
		eventZones:     make(map[uint64]uint8),
		phonemeState:   make(map[uint64]uint8),
		eventSounds:    make(map[uint64]int32),
		playerPositions: map[uint64]uint64{
			remoteAudioPlayer501CA0: remoteAudioPlayerPos501CA0,
		},
		soundField20: make(map[int32]int32),
	}
}

func (w *remotePlayerAudioWorld501CA0) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic("injected fault")
	}
}

func (w *remotePlayerAudioWorld501CA0) hooks() remotePlayerAudioHooks501CA0[
	uint64, uint64, uint64, uint64, uint64, uint64, uint64,
] {
	return remotePlayerAudioHooks501CA0[uint64, uint64, uint64, uint64, uint64, uint64, uint64]{
		loadUpdate: func(object uint64) uint64 {
			w.event(fmt.Sprintf("update:%#x", object))
			return w.updates[object]
		},
		loadPlayer: func(update uint64) uint64 {
			w.event(fmt.Sprintf("player:%#x", update))
			return w.players[update]
		},
		loadPlayerFlagsLow: func(player uint64) uint8 {
			w.event(fmt.Sprintf("flags:%#x", player))
			return w.playerFlags[player]
		},
		loadCameraTarget: func(player uint64) uint64 {
			w.event(fmt.Sprintf("camera:%#x", player))
			return w.cameraTargets[player]
		},
		loadClassLow: func(object uint64) uint8 {
			w.event(fmt.Sprintf("class:%#x", object))
			return w.classes[object]
		},
		loadObjectPositionX: func(object uint64) float32 {
			w.event(fmt.Sprintf("position-x:%#x", object))
			return w.positionX[object]
		},
		floatToInt: func(value float32) int32 {
			w.event(fmt.Sprintf("float-to-int:%g", value))
			w.floatCalls++
			if w.floatCalls == 1 && w.firstFloatMutation != nil {
				w.firstFloatMutation()
			}
			return int32(value)
		},
		loadObjectPositionY: func(object uint64) float32 {
			w.event(fmt.Sprintf("position-y:%#x", object))
			return w.positionY[object]
		},
		polygonAtPoint: func(point [2]int32, previous uint32) uint64 {
			w.event(fmt.Sprintf("polygon-at:%d,%d:%#x", point[0], point[1], previous))
			return w.polygon
		},
		loadPolygonZone: func(polygon uint64) uint8 {
			w.event(fmt.Sprintf("polygon-zone:%#x", polygon))
			return w.polygonZone[polygon]
		},
		loadCurrentPolygonID: func(player uint64) uint32 {
			w.event(fmt.Sprintf("current-polygon:%#x", player))
			return w.currentPolygon[player]
		},
		questCheckSecretArea: func(object uint64) {
			w.event(fmt.Sprintf("quest:%#x", object))
			if w.questMutation != nil {
				w.questMutation()
			}
		},
		loadPlayerAudioZone: func(player uint64) uint8 {
			w.event(fmt.Sprintf("player-zone:%#x", player))
			return w.playerZones[player]
		},
		resetBitmap: func() {
			w.event("reset-bitmap")
		},
		hasTeam: func(object uint64) bool {
			w.event(fmt.Sprintf("has-team:%#x", object))
			return w.hasTeamValue
		},
		loadTeamID: func(object uint64) uint32 {
			w.event(fmt.Sprintf("team-id:%#x", object))
			return w.teamIDs[object]
		},
		teamByID: func(id uint32) uint64 {
			w.event(fmt.Sprintf("team-by-id:%#x", id))
			return w.teams[id]
		},
		firstEvent: func() uint64 {
			w.event("first-event")
			return w.head
		},
		loadEventKind: func(event uint64) int32 {
			w.event(fmt.Sprintf("event-kind:%#x", event))
			return w.eventKind[event]
		},
		loadEventCode: func(event uint64) uint32 {
			w.event(fmt.Sprintf("event-code:%#x", event))
			return w.eventCode[event]
		},
		loadObjectNetCode: func(object uint64) uint32 {
			w.event(fmt.Sprintf("net-code:%#x", object))
			return w.netCode[object]
		},
		loadEventObject: func(event uint64) uint64 {
			w.event(fmt.Sprintf("event-object:%#x", event))
			return w.eventObject[event]
		},
		eventPosition: func(event uint64) uint64 {
			w.event(fmt.Sprintf("event-position:%#x", event))
			return w.eventPositions[event]
		},
		eventZone: func(position, object uint64) uint8 {
			w.event(fmt.Sprintf("event-zone:%#x:%#x", position, object))
			return w.eventZones[position]
		},
		loadPhonemeState: func(update uint64) uint8 {
			w.event(fmt.Sprintf("phoneme-state:%#x", update))
			return w.phonemeState[update]
		},
		loadEventSound: func(event uint64) int32 {
			w.event(fmt.Sprintf("event-sound:%#x", event))
			return w.eventSounds[event]
		},
		playerPosition: func(player uint64) uint64 {
			w.event(fmt.Sprintf("player-position:%#x", player))
			return w.playerPositions[player]
		},
		fade: func(sound int32, eventPosition, listenerPosition uint64) int32 {
			w.event(fmt.Sprintf("fade:%d:%#x:%#x", sound, eventPosition, listenerPosition))
			return w.fadeValue
		},
		loadSoundField20: func(sound int32) int32 {
			w.event(fmt.Sprintf("sound-field20:%d", sound))
			return w.soundField20[sound]
		},
		addAudio: func(event uint64, fade int32) {
			w.event(fmt.Sprintf("add-audio:%#x:%d", event, fade))
			if w.addMutation != nil {
				w.addMutation(event, fade)
			}
		},
		sendDirect: func(object, event uint64, fade int32) {
			w.event(fmt.Sprintf("send-direct:%#x:%#x:%d", object, event, fade))
			if w.directMutation != nil {
				w.directMutation(object, event, fade)
			}
		},
		loadEventNext: func(event uint64) uint64 {
			w.event(fmt.Sprintf("event-next:%#x", event))
			return w.eventNext[event]
		},
		flush: func(object uint64) {
			w.event(fmt.Sprintf("flush:%#x", object))
		},
	}
}

func (w *remotePlayerAudioWorld501CA0) run() {
	remotePlayerAudioUpdate501CA0(remoteAudioUnit501CA0, w.hooks())
}

func checkRemotePlayerAudioEvents501CA0(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("events =\n%v\nwant =\n%v", got, want)
	}
}

func TestRemotePlayerAudio501CA0ReloadsPlayerAfterSecretCallback(t *testing.T) {
	w := newRemotePlayerAudioWorld501CA0()
	w.currentPolygon[remoteAudioPlayer501CA0] = remotePlayerAudioUninitialized501CA0
	w.playerZones[remoteAudioPlayer2_501CA0] = 0xe7
	w.questMutation = func() {
		w.players[remoteAudioUpdate501CA0] = remoteAudioPlayer2_501CA0
	}
	w.run()
	checkRemotePlayerAudioEvents501CA0(t, w.events, []string{
		"update:0x100000001", "player:0x200000002", "flags:0x300000003",
		"current-polygon:0x300000003", "quest:0x100000001",
		"player:0x200000002", "player-zone:0x300000004", "reset-bitmap",
		"has-team:0x100000001", "first-event", "flush:0x100000001",
	})
}

func TestRemotePlayerAudio501CA0FollowPlayerBranch(t *testing.T) {
	w := newRemotePlayerAudioWorld501CA0()
	w.playerFlags[remoteAudioPlayer501CA0] = 3
	w.cameraTargets[remoteAudioPlayer501CA0] = remoteAudioFollow501CA0
	w.classes[remoteAudioFollow501CA0] = remotePlayerAudioPlayerClass501CA0
	w.updates[remoteAudioFollow501CA0] = remoteAudioFollowUpd501CA0
	w.players[remoteAudioFollowUpd501CA0] = remoteAudioFollowPlr501CA0
	w.playerZones[remoteAudioFollowPlr501CA0] = 0xc3
	w.run()
	checkRemotePlayerAudioEvents501CA0(t, w.events, []string{
		"update:0x100000001", "player:0x200000002", "flags:0x300000003",
		"camera:0x300000003", "class:0x400000004", "update:0x400000004",
		"player:0x500000005", "player-zone:0x600000006", "reset-bitmap",
		"has-team:0x100000001", "first-event", "flush:0x100000001",
	})
}

func TestRemotePlayerAudio501CA0ReloadsFollowBetweenCoordinates(t *testing.T) {
	w := newRemotePlayerAudioWorld501CA0()
	w.playerFlags[remoteAudioPlayer501CA0] = 1
	w.cameraTargets[remoteAudioPlayer501CA0] = remoteAudioFollow501CA0
	w.cameraTargets[remoteAudioPlayer2_501CA0] = remoteAudioFollow2_501CA0
	w.positionX[remoteAudioFollow501CA0] = 17
	w.positionY[remoteAudioFollow2_501CA0] = -23
	w.polygon = remoteAudioPolygon501CA0
	w.polygonZone[remoteAudioPolygon501CA0] = 0xa5
	w.firstFloatMutation = func() {
		w.players[remoteAudioUpdate501CA0] = remoteAudioPlayer2_501CA0
	}
	w.run()
	checkRemotePlayerAudioEvents501CA0(t, w.events, []string{
		"update:0x100000001", "player:0x200000002", "flags:0x300000003",
		"camera:0x300000003", "class:0x400000004", "camera:0x300000003",
		"position-x:0x400000004", "float-to-int:17", "player:0x200000002",
		"camera:0x300000004", "position-y:0x400000005", "float-to-int:-23",
		"polygon-at:17,-23:0x0", "polygon-zone:0x700000007", "reset-bitmap",
		"has-team:0x100000001", "first-event", "flush:0x100000001",
	})
}

func configureFullRemotePlayerAudioPath501CA0() *remotePlayerAudioWorld501CA0 {
	w := newRemotePlayerAudioWorld501CA0()
	w.hasTeamValue = true
	w.teamIDs[remoteAudioUnit501CA0] = 9
	w.teams[9] = remoteAudioTeam501CA0
	w.head = remoteAudioEvent501CA0
	w.eventKind[remoteAudioEvent501CA0] = 1
	w.eventCode[remoteAudioEvent501CA0] = 9
	w.eventObject[remoteAudioEvent501CA0] = remoteAudioFollow501CA0
	w.eventPositions[remoteAudioEvent501CA0] = remoteAudioPosition501CA0
	w.eventZones[remoteAudioPosition501CA0] = 7
	w.eventSounds[remoteAudioEvent501CA0] = remotePlayerAudioFirstPhoneme501CA0
	w.fadeValue = 9
	w.soundField20[remotePlayerAudioFirstPhoneme501CA0] = 1
	return w
}

func TestRemotePlayerAudio501CA0EventOrderAndLiveNext(t *testing.T) {
	w := configureFullRemotePlayerAudioPath501CA0()
	w.eventKind[remoteAudioEvent2_501CA0] = 2
	w.netCode[remoteAudioUnit501CA0] = 0x77
	w.eventCode[remoteAudioEvent2_501CA0] = 0x88
	w.addMutation = func(event uint64, fade int32) {
		if event != remoteAudioEvent501CA0 || fade != 4 {
			t.Fatalf("add = %#x/%d, want first event/4", event, fade)
		}
		w.eventNext[remoteAudioEvent501CA0] = remoteAudioEvent2_501CA0
	}
	w.run()
	checkRemotePlayerAudioEvents501CA0(t, w.events, []string{
		"update:0x100000001", "player:0x200000002", "flags:0x300000003",
		"current-polygon:0x300000003", "player:0x200000002", "player-zone:0x300000003",
		"reset-bitmap", "has-team:0x100000001", "team-id:0x100000001", "team-by-id:0x9",
		"first-event", "event-kind:0x900000009", "event-code:0x900000009", "team-by-id:0x9",
		"event-object:0x900000009", "event-position:0x900000009", "event-zone:0xa0000000a:0x400000004",
		"phoneme-state:0x200000002", "event-sound:0x900000009", "event-object:0x900000009",
		"player:0x200000002", "event-sound:0x900000009", "player-position:0x300000003",
		"fade:186:0xa0000000a:0xb0000000b", "event-sound:0x900000009", "sound-field20:186",
		"add-audio:0x900000009:4", "event-next:0x900000009",
		"event-kind:0x90000000a", "net-code:0x100000001", "event-code:0x90000000a",
		"event-next:0x90000000a", "flush:0x100000001",
	})
}

func TestRemotePlayerAudio501CA0FiltersAndDirectSend(t *testing.T) {
	t.Run("team lookup precedes nil-listener rejection", func(t *testing.T) {
		w := newRemotePlayerAudioWorld501CA0()
		w.head = remoteAudioEvent501CA0
		w.eventKind[remoteAudioEvent501CA0] = 1
		w.eventCode[remoteAudioEvent501CA0] = 0x1234
		w.teams[0x1234] = remoteAudioTeam501CA0
		w.run()
		wantTail := []string{
			"first-event", "event-kind:0x900000009", "event-code:0x900000009",
			"team-by-id:0x1234", "event-next:0x900000009", "flush:0x100000001",
		}
		checkRemotePlayerAudioEvents501CA0(t, w.events[len(w.events)-len(wantTail):], wantTail)
	})

	t.Run("zone mismatch stops before phoneme state", func(t *testing.T) {
		w := newRemotePlayerAudioWorld501CA0()
		w.head = remoteAudioEvent501CA0
		w.eventObject[remoteAudioEvent501CA0] = remoteAudioFollow501CA0
		w.eventPositions[remoteAudioEvent501CA0] = remoteAudioPosition501CA0
		w.eventZones[remoteAudioPosition501CA0] = 8
		w.run()
		for _, event := range w.events {
			if event == "phoneme-state:0x200000002" {
				t.Fatal("zone mismatch reached phoneme state")
			}
		}
	})

	t.Run("self phoneme is suppressed after object reload", func(t *testing.T) {
		w := newRemotePlayerAudioWorld501CA0()
		w.head = remoteAudioEvent501CA0
		w.eventObject[remoteAudioEvent501CA0] = remoteAudioUnit501CA0
		w.eventPositions[remoteAudioEvent501CA0] = remoteAudioPosition501CA0
		w.eventZones[remoteAudioPosition501CA0] = 0
		w.eventSounds[remoteAudioEvent501CA0] = remotePlayerAudioLastPhoneme501CA0
		w.run()
		wantTail := []string{
			"event-object:0x900000009", "event-position:0x900000009",
			"event-zone:0xa0000000a:0x100000001", "phoneme-state:0x200000002",
			"event-sound:0x900000009", "event-object:0x900000009",
			"event-next:0x900000009", "flush:0x100000001",
		}
		checkRemotePlayerAudioEvents501CA0(t, w.events[len(w.events)-len(wantTail):], wantTail)
	})

	t.Run("non-bitmap sound is sent directly", func(t *testing.T) {
		w := newRemotePlayerAudioWorld501CA0()
		w.head = remoteAudioEvent501CA0
		w.eventObject[remoteAudioEvent501CA0] = remoteAudioFollow501CA0
		w.eventPositions[remoteAudioEvent501CA0] = remoteAudioPosition501CA0
		w.eventZones[remoteAudioPosition501CA0] = 0
		w.phonemeState[remoteAudioUpdate501CA0] = 1
		w.eventSounds[remoteAudioEvent501CA0] = 42
		w.fadeValue = 11
		var sent bool
		w.directMutation = func(object, event uint64, fade int32) {
			sent = object == remoteAudioUnit501CA0 && event == remoteAudioEvent501CA0 && fade == 5
		}
		w.run()
		if !sent {
			t.Fatal("direct send did not receive native handles and shifted fade")
		}
	})
}

func TestRemotePlayerAudio501CA0UsesArithmeticShift(t *testing.T) {
	if got := remotePlayerAudioHalfFade501CA0(-3); got != -2 {
		t.Fatalf("arithmetic half of -3 = %d, want -2", got)
	}
}

func TestRemotePlayerAudio501CA0EveryObservableFaultPrefix(t *testing.T) {
	worlds := []struct {
		name string
		new  func() *remotePlayerAudioWorld501CA0
	}{
		{
			name: "normal quest",
			new: func() *remotePlayerAudioWorld501CA0 {
				w := newRemotePlayerAudioWorld501CA0()
				w.currentPolygon[remoteAudioPlayer501CA0] = remotePlayerAudioUninitialized501CA0
				return w
			},
		},
		{
			name: "follow player",
			new: func() *remotePlayerAudioWorld501CA0 {
				w := newRemotePlayerAudioWorld501CA0()
				w.playerFlags[remoteAudioPlayer501CA0] = 1
				w.cameraTargets[remoteAudioPlayer501CA0] = remoteAudioFollow501CA0
				w.classes[remoteAudioFollow501CA0] = remotePlayerAudioPlayerClass501CA0
				w.updates[remoteAudioFollow501CA0] = remoteAudioFollowUpd501CA0
				w.players[remoteAudioFollowUpd501CA0] = remoteAudioFollowPlr501CA0
				return w
			},
		},
		{
			name: "follow polygon",
			new: func() *remotePlayerAudioWorld501CA0 {
				w := newRemotePlayerAudioWorld501CA0()
				w.playerFlags[remoteAudioPlayer501CA0] = 1
				w.cameraTargets[remoteAudioPlayer501CA0] = remoteAudioFollow501CA0
				w.positionX[remoteAudioFollow501CA0] = 1
				w.positionY[remoteAudioFollow501CA0] = 2
				w.polygon = remoteAudioPolygon501CA0
				return w
			},
		},
		{name: "event", new: configureFullRemotePlayerAudioPath501CA0},
	}

	for _, test := range worlds {
		t.Run(test.name, func(t *testing.T) {
			base := test.new()
			base.run()
			want := append([]string(nil), base.events...)
			for faultAt := 1; faultAt <= len(want); faultAt++ {
				t.Run(fmt.Sprintf("event_%02d", faultAt), func(t *testing.T) {
					w := test.new()
					w.faultAt = faultAt
					func() {
						defer func() {
							if recover() == nil {
								t.Fatal("expected injected fault")
							}
						}()
						w.run()
					}()
					checkRemotePlayerAudioEvents501CA0(t, w.events, want[:faultAt])
				})
			}
		})
	}
}
