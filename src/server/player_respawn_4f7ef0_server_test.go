package server

import (
	"encoding/binary"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestPlayerRespawnNative4F7EF0KeepsNativePointers(t *testing.T) {
	settings := &Settings{}
	binary.LittleEndian.PutUint32(settings.PlayerSkeletons58[:], 1)
	playerUnit := &Object{TeamVal: ObjectTeam{ID: 7}}
	entryPlayer := &Player{PlayerUnit: playerUnit, PlayerInd: 3, Field4700: 99}
	livePlayer := &Player{PlayerInd: 5}
	gate := &Object{PosVec: types.Pointf{X: 9, Y: 10}}
	update := &PlayerUpdateData{Player: entryPlayer, SoulGate: gate}
	unit := &Object{
		UpdateData: unsafe.Pointer(update),
		PosVec:     types.Pointf{X: 11, Y: 12},
		Direction1: Dir16(0xff81),
	}
	team := &Team{IDVal: 7}
	crown := &Object{}

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"settings": unsafe.Pointer(settings),
			"unit":     unsafe.Pointer(unit),
			"update":   unsafe.Pointer(update),
			"player":   unsafe.Pointer(entryPlayer),
			"gate":     unsafe.Pointer(gate),
			"team":     unsafe.Pointer(team),
			"crown":    unsafe.Pointer(crown),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer did not allocate above 32 bits: %p", name, pointer)
			}
		}
	}

	var madeItems bool
	var corpseCenter *types.Pointf
	var corpseDirection int16
	var moved *Object
	var movedTo types.Pointf
	var pickedWho, pickedCrown *Object
	var buffed *Object
	result := playerRespawnNative4F7EF0(unit, playerRespawnNativeDeps4F7EF0{
		loadSettings: func() *Settings { return settings },
		gameFlag: func(flag uint32) int32 {
			switch flag {
			case playerRespawnQuestFlag4F7EF0,
				playerRespawnCrownFlag4F7EF0,
				playerRespawnProtectionFlag4F7EF0:
				return 1
			default:
				return 0
			}
		},
		makeDefaultItems: func(got *Object, stats, keep int32) {
			if got != unit || stats != 1 || keep != 1 {
				t.Fatalf("default items = %p/%d/%d", got, stats, keep)
			}
			madeItems = true
			update.Player = livePlayer
		},
		priorityMessage: func(got *Object, message string, value uint8) {
			if got != unit || message != playerRespawnMessage4F7EF0 || value != 0 {
				t.Fatalf("message = %p/%q/%d", got, message, value)
			}
		},
		audio: func(id uint32, got *Object, first, second int32) {
			if id != playerRespawnQuestSound4F7EF0 || got != unit || first != 0 || second != 0 {
				t.Fatalf("audio = %d/%p/%d/%d", id, got, first, second)
			}
		},
		loadNetworkMode: func() uint32 { return 1 },
		respawnCorpse: func(center *types.Pointf, direction int16) {
			corpseCenter, corpseDirection = center, direction
		},
		soulGatePoint: func(got *Object, output *types.Pointf) {
			if got != gate || *output != unit.PosVec {
				t.Fatalf("soul gate input = %p/%+v", got, *output)
			}
			*output = types.Pointf{X: 21, Y: 22}
		},
		mapPlayerStart: func(*types.Pointf, *Object) {
			t.Fatal("quest SoulGate path called map start")
		},
		move: func(got *Object, destination types.Pointf) {
			moved, movedTo = got, destination
		},
		gameplayFlag: func(flag uint32) int32 {
			if flag != playerRespawnCrownPlayFlag4F7EF0 {
				t.Fatalf("gameplay flag = %#x", flag)
			}
			return 1
		},
		teamByID: func(id uint8) *Team {
			if id != 7 {
				t.Fatalf("team id = %d", id)
			}
			return team
		},
		loadTeamCrown: func(got *Team) *Object {
			if got != team {
				t.Fatalf("team pointer = %p", got)
			}
			return crown
		},
		crownPickup: func(who, gotCrown *Object, first, second int32) {
			if first != 1 || second != 1 {
				t.Fatalf("pickup flags = %d/%d", first, second)
			}
			pickedWho, pickedCrown = who, gotCrown
		},
		loadTickRate: func() uint32 { return 30 },
		applyBuff: func(got *Object, enchant EnchantID, duration uint16, power uint8) {
			if enchant != EnchantID(playerRespawnEnchant4F7EF0) || duration != 150 || power != 5 {
				t.Fatalf("buff = %d/%d/%d", enchant, duration, power)
			}
			buffed = got
		},
	})

	if result.kind != playerRespawnResultScalar4F7EF0 || result.value != 1 {
		t.Fatalf("result = %+v", result)
	}
	if !madeItems || entryPlayer.Field4700 != 0 || update.RespawnMarkers[livePlayer.PlayerInd] != 0xfa {
		t.Fatalf("respawn state: items=%v done=%d marker=%02x", madeItems, entryPlayer.Field4700, update.RespawnMarkers[livePlayer.PlayerInd])
	}
	if corpseCenter != &unit.PosVec || corpseDirection != -127 {
		t.Fatalf("corpse args = %p/%d, want %p/-127", corpseCenter, corpseDirection, &unit.PosVec)
	}
	if moved != unit || movedTo != (types.Pointf{X: 21, Y: 22}) {
		t.Fatalf("move = %p/%+v", moved, movedTo)
	}
	if pickedWho != playerUnit || pickedCrown != crown || buffed != unit {
		t.Fatalf("native pointers changed: pickup=%p/%p buff=%p", pickedWho, pickedCrown, buffed)
	}
}

func TestPlayerRespawnResultLow4F7EF0NarrowsOnlyAtABI(t *testing.T) {
	settings := &Settings{}
	got := playerRespawnResultLow4F7EF0(playerRespawnResult4F7EF0[*Settings]{
		kind:     playerRespawnResultSettings4F7EF0,
		settings: settings,
	})
	want := int16(uint16(uintptr(unsafe.Pointer(settings))))
	if got != want {
		t.Fatalf("settings low short = %#x, want %#x", uint16(got), uint16(want))
	}
	if got := playerRespawnResultLow4F7EF0(playerRespawnResult4F7EF0[*Settings]{value: 0xffff8001}); got != -32767 {
		t.Fatalf("scalar low short = %d", got)
	}
}

func TestSoulGateRespawnPointNative4F80C0UsesLiveNativePosition(t *testing.T) {
	gate := &Object{PosVec: types.Pointf{X: 1, Y: 2}}
	output := types.Pointf{X: -1, Y: -1}
	var centerPointer *types.Pointf
	result := soulGateRespawnPointNative4F80C0(gate, &output, soulGateRespawnPointNativeDeps4F80C0{
		randomPoint: func(radius float32, center, gotOutput *types.Pointf) {
			if radius != 60 || gotOutput != &output {
				t.Fatalf("random args = %v/%p", radius, gotOutput)
			}
			centerPointer = center
			*gotOutput = types.Pointf{X: 3, Y: 4}
		},
		allowTeleport: func(got *types.Pointf) int32 {
			if got != &output {
				t.Fatalf("allow pointer = %p", got)
			}
			return 0
		},
	})
	if result != 0 || centerPointer != &gate.PosVec || output != (types.Pointf{X: 3, Y: 4}) {
		t.Fatalf("result=%d center=%p output=%+v", result, centerPointer, output)
	}
}

func TestRespawnPlayerImplNative53FBC0KeepsCreatedObjectPointers(t *testing.T) {
	center := &types.Pointf{X: 100, Y: 200}
	first := &Object{ObjFlags: object.Flags(0x12340000)}
	second := &Object{ObjFlags: object.Flags(0xabcd0001)}
	objects := []*Object{first, second, nil}
	var created []*Object
	var positions []types.Pointf
	var decayed []*Object
	respawnPlayerImplNative53FBC0(center, -5, respawnPlayerImplNativeDeps53FBC0{
		initialized: func() uint32 { return 1 },
		initialize:  func() { t.Fatal("unexpected initializer") },
		direction: func(direction int16) int32 {
			if direction != -5 {
				t.Fatalf("direction = %d", direction)
			}
			return 4
		},
		typeIndex: func(direction int32, part int) uint32 {
			if direction != 4 {
				t.Fatalf("direction index = %d", direction)
			}
			return uint32(part + 70)
		},
		offset: func(_ int32, part int) types.Pointf {
			return types.Pointf{X: float32(part + 1), Y: float32(part + 2)}
		},
		newObject: func(typeIndex uint32) *Object {
			return objects[int(typeIndex-70)]
		},
		networkMode: func() uint32 { return 1 },
		gameFlag: func(flag uint32) int32 {
			if flag != playerRespawnProtectionFlag4F7EF0 {
				t.Fatalf("game flag = %#x", flag)
			}
			return 1
		},
		createAt: func(object *Object, position types.Pointf) {
			created = append(created, object)
			positions = append(positions, position)
		},
		randomInt: func(minimum, maximum int32) int32 {
			if minimum != 10 || maximum != 20 {
				t.Fatalf("random bounds = %d/%d", minimum, maximum)
			}
			return 12
		},
		tickRate: func() uint32 { return 30 },
		setDecay: func(object *Object, delay uint32) {
			if delay != 360 {
				t.Fatalf("delay = %d", delay)
			}
			decayed = append(decayed, object)
		},
	})
	if len(created) != 2 || created[0] != first || created[1] != second || decayed[0] != first || decayed[1] != second {
		t.Fatalf("native objects changed: created=%p/%p decayed=%p/%p", created[0], created[1], decayed[0], decayed[1])
	}
	if positions[0] != (types.Pointf{X: 101, Y: 202}) || positions[1] != (types.Pointf{X: 102, Y: 203}) {
		t.Fatalf("positions = %+v", positions)
	}
	if uint32(first.ObjFlags) != 0x12340040 || uint32(second.ObjFlags) != 0xabcd0041 {
		t.Fatalf("flags = %#08x/%#08x", first.ObjFlags, second.ObjFlags)
	}
}
