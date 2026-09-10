package server

import (
	"math"
	"reflect"
	"testing"

	"github.com/opennox/libs/types"
)

func TestCharmStart5011F0ConfusePathPreservesNativeIdentityAndNarrowing(t *testing.T) {
	const (
		record = uint64(0x100000101)
		caster = uint64(0x7f031b61e540)
		target = uint64(0x7f031b61e8a0)
	)
	var events []string
	var buffTarget uint64
	var buff, duration int32
	var power int8
	hooks := charmStartHooks5011F0[uint64, uint64, uint64, uint64]{
		loadRecordFlag: func(got uint64) uint32 {
			events = append(events, "flag")
			if got != record {
				t.Fatalf("flag record = %#x", got)
			}
			return 1
		},
		balanceFloat: func(key string) float64 {
			events = append(events, "balance")
			if key != charmConfuseDurationKey5011F0 {
				t.Fatalf("balance key = %q", key)
			}
			return 12345.678901
		},
		floatToInt: func(value float32) int32 {
			events = append(events, "float-int")
			if value != float32(12345.678901) {
				t.Fatalf("float input bits = %#x, want binary32 spill %#x",
					math.Float32bits(value), math.Float32bits(float32(12345.678901)))
			}
			return 0x12345
		},
		loadLevel: func(got uint64) uint32 {
			events = append(events, "level")
			return 0x180
		},
		loadTarget: func(got uint64) uint64 {
			events = append(events, "target")
			return target
		},
		loadCaster: func(got uint64) uint64 {
			events = append(events, "caster")
			return caster
		},
		buffApply: func(gotTarget uint64, gotBuff int32, gotDuration int16, gotPower int8) {
			events = append(events, "buff")
			buffTarget = gotTarget
			buff = gotBuff
			duration = int32(gotDuration)
			power = gotPower
		},
		attribution: func(gotCaster, gotTarget uint64) {
			events = append(events, "attribution")
			if gotCaster != caster || gotTarget != target {
				t.Fatalf("attribution = %#x/%#x", gotCaster, gotTarget)
			}
		},
	}

	if got := charmStart5011F0(record, hooks); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	wantEvents := []string{
		"flag", "balance", "float-int", "level", "target", "buff",
		"target", "caster", "attribution",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %q, want %q", events, wantEvents)
	}
	if buffTarget != target || buff != charmConfusedEnchant5011F0 ||
		duration != int32(int16(0x2345)) || power != -128 {
		t.Fatalf("buff = %#x/%d/%d/%d", buffTarget, buff, duration, power)
	}
}

func TestCharmStart5011F0FallbackUsesRecordDwordWithoutLevelLoad(t *testing.T) {
	const (
		record = uint64(0x1fffffff0)
		caster = uint64(0x7f031b61e540)
		target = uint64(0x7f031b61e8a0)
	)
	position := types.Ptf(12.5, -7.25)
	var events []string
	var storedTarget uint64
	var deadline uint32
	var buffDuration int16
	hooks := charmStartHooks5011F0[uint64, uint64, uint64, uint64]{
		loadRecordFlag: func(uint64) uint32 {
			events = append(events, "flag")
			return 0
		},
		loadCaster: func(uint64) uint64 {
			events = append(events, "caster")
			return caster
		},
		loadSpell: func(uint64) uint32 {
			events = append(events, "spell")
			return 0x123
		},
		spellFlags: func(id uint32) uint32 {
			events = append(events, "spell-flags")
			if id != 0x123 {
				t.Fatalf("spell ID = %#x", id)
			}
			return 0x456
		},
		loadPosition: func(uint64) *types.Pointf {
			events = append(events, "position")
			return &position
		},
		searchTarget: func(pos *types.Pointf, liveCaster uint64, flags uint32, distance float32, mode int32, self uint64) uint64 {
			events = append(events, "search")
			if pos != &position || liveCaster != caster || self != caster || flags != 0x456 ||
				distance != charmSearchRange5011F0 || mode != 0 {
				t.Fatalf("search args = %p/%#x/%#x/%v/%d/%#x", pos, liveCaster, flags, distance, mode, self)
			}
			return target
		},
		storeTarget: func(_ uint64, got uint64) {
			events = append(events, "store-target")
			storedTarget = got
		},
		loadClassLow: func(obj uint64) byte {
			events = append(events, "class")
			if obj == target {
				return charmMonsterClass5011F0
			}
			return 0
		},
		monitored: func(gotCaster, gotTarget uint64) bool {
			events = append(events, "monitored")
			if gotCaster != caster || gotTarget != target {
				t.Fatalf("monitored = %#x/%#x", gotCaster, gotTarget)
			}
			return false
		},
		loadTarget: func(uint64) uint64 {
			events = append(events, "target")
			return target
		},
		loadTypeIndex: func(got uint64) uint16 {
			events = append(events, "type")
			if got != target {
				t.Fatalf("type target = %#x", got)
			}
			return 0x789
		},
		charmable: func(typeIndex uint16) int32 {
			events = append(events, "charmable")
			if typeIndex != 0x789 {
				t.Fatalf("type index = %#x", typeIndex)
			}
			return 7
		},
		guideSize: func(guide int32) int32 {
			events = append(events, "guide-size")
			if guide != 7 {
				t.Fatalf("guide = %d", guide)
			}
			return 0x12340003
		},
		loadLevel: func(uint64) uint32 {
			t.Fatal("fallback loaded spell level")
			return 0
		},
		balanceFloatInd: func(string, int32) float64 {
			t.Fatal("fallback loaded a duration table")
			return 0
		},
		floatToInt: func(float32) int32 {
			t.Fatal("fallback converted a duration table value")
			return 0
		},
		recordDword: func(got uint64) uint32 {
			events = append(events, "record-dword")
			if got != record {
				t.Fatalf("record dword source = %#x", got)
			}
			return uint32(got)
		},
		frame: func() uint32 {
			events = append(events, "frame")
			return 5
		},
		storeDeadline: func(_ uint64, value uint32) {
			events = append(events, "deadline")
			deadline = value
		},
		buffApply: func(gotTarget uint64, buff int32, duration int16, power int8) {
			events = append(events, "buff")
			if gotTarget != target || buff != charmPendingEnchant5011F0 || power != charmPendingPower5011F0 {
				t.Fatalf("pending buff = %#x/%d/%d", gotTarget, buff, power)
			}
			buffDuration = duration
		},
		durationRayStart: func(got uint64) {
			events = append(events, "ray")
			if got != record {
				t.Fatalf("ray record = %#x", got)
			}
		},
	}

	if got := charmStart5011F0(record, hooks); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	wantEvents := []string{
		"flag", "caster", "spell", "spell-flags", "caster", "position", "search",
		"store-target", "class", "caster", "monitored", "target", "type", "charmable",
		"caster", "class", "guide-size", "record-dword", "frame", "target", "deadline",
		"buff", "ray",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %q, want exact oracle order %q", events, wantEvents)
	}
	if storedTarget != target {
		t.Fatalf("stored target = %#x", storedTarget)
	}
	if deadline != 0xfffffff5 || buffDuration != -15 {
		t.Fatalf("deadline/duration = %#x/%d, want %#x/-15", deadline, buffDuration, uint32(0xfffffff5))
	}
}

func TestCharmStart5011F0SupportedSizeUsesWrappedLevelSelector(t *testing.T) {
	const (
		record = uint64(0x100000001)
		caster = uint64(0x100000002)
		target = uint64(0x100000003)
	)
	position := types.Pointf{}
	var selector int32
	var deadline uint32
	var duration int16
	hooks := charmStartHooks5011F0[uint64, uint64, uint64, uint64]{
		loadRecordFlag: func(uint64) uint32 { return 0 },
		loadCaster:     func(uint64) uint64 { return caster },
		loadSpell:      func(uint64) uint32 { return 1 },
		spellFlags:     func(uint32) uint32 { return 0 },
		loadPosition:   func(uint64) *types.Pointf { return &position },
		searchTarget: func(*types.Pointf, uint64, uint32, float32, int32, uint64) uint64 {
			return target
		},
		storeTarget: func(uint64, uint64) {},
		loadClassLow: func(obj uint64) byte {
			if obj == target {
				return charmMonsterClass5011F0
			}
			return 0
		},
		monitored:     func(uint64, uint64) bool { return false },
		loadTarget:    func(uint64) uint64 { return target },
		loadTypeIndex: func(uint64) uint16 { return 1 },
		charmable:     func(uint16) int32 { return 2 },
		guideSize:     func(int32) int32 { return 0x12340004 },
		loadLevel:     func(uint64) uint32 { return 0 },
		balanceFloatInd: func(key string, gotSelector int32) float64 {
			if key != charmLargeDurationKey5011F0 {
				t.Fatalf("duration key = %q", key)
			}
			selector = gotSelector
			return 9.75
		},
		floatToInt: func(value float32) int32 {
			if value != float32(9.75) {
				t.Fatalf("duration input = %v", value)
			}
			return 9
		},
		recordDword: func(uint64) uint32 {
			t.Fatal("supported guide size used record dword")
			return 0
		},
		frame: func() uint32 { return math.MaxUint32 - 4 },
		storeDeadline: func(_ uint64, value uint32) {
			deadline = value
		},
		buffApply: func(_ uint64, _ int32, gotDuration int16, _ int8) {
			duration = gotDuration
		},
		durationRayStart: func(uint64) {},
	}

	if got := charmStart5011F0(record, hooks); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if selector != -1 {
		t.Fatalf("selector = %d, want wrapped -1", selector)
	}
	if deadline != 4 || duration != 10 {
		t.Fatalf("deadline/duration = %#x/%d, want 4/10", deadline, duration)
	}
}

func TestCharmFinish5013E0PlayerQuestSuccessOrderAndNativeIdentity(t *testing.T) {
	const (
		record        = uint64(0x100000101)
		caster        = uint64(0x7f031b61e540)
		target        = uint64(0x7f031b61e8a0)
		monsterUpdate = uint64(0x7f031b620100)
		monsterDef    = uint64(0x7f031b620200)
		playerUpdate  = uint64(0x7f031b620300)
		player        = uint64(0x7f031b620400)
	)
	var events []string
	observe := func(event string) {
		events = append(events, event)
	}
	var status uint32 = 0x12
	var subclass uint32 = 0x20
	var healthMax uint16
	var setHP uint16
	hooks := charmFinishHooks5013E0[uint64, uint64, uint64, uint64, uint64, uint64]{
		loadTarget: func(got uint64) uint64 {
			observe("target")
			if got != record {
				t.Fatalf("target record = %#x", got)
			}
			return target
		},
		loadCaster: func(got uint64) uint64 {
			observe("caster")
			if got != record {
				t.Fatalf("caster record = %#x", got)
			}
			return caster
		},
		loadObjectFlags: func(got uint64) uint32 {
			observe("flags")
			if got != target {
				t.Fatalf("flags target = %#x", got)
			}
			return 0
		},
		distance: func(gotCaster, gotTarget uint64) float64 {
			observe("distance")
			if gotCaster != caster || gotTarget != target {
				t.Fatalf("distance objects = %#x/%#x", gotCaster, gotTarget)
			}
			return 300
		},
		loadDeadline: func(uint64) uint32 {
			observe("deadline")
			return 0
		},
		frame: func() uint32 {
			observe("frame")
			return math.MaxUint32
		},
		loadSubclass: func(got uint64) uint32 {
			observe("subclass")
			if got != target {
				t.Fatalf("subclass target = %#x", got)
			}
			return subclass
		},
		storeSubclass: func(got uint64, value uint32) {
			observe("store-subclass")
			if got != target {
				t.Fatalf("subclass store target = %#x", got)
			}
			subclass = value
		},
		loadClassLow: func(got uint64) byte {
			observe("class")
			if got != caster {
				t.Fatalf("class object = %#x", got)
			}
			return charmPlayerClass5011F0
		},
		loadTypeIndex: func(got uint64) uint16 {
			observe("type")
			if got != target {
				t.Fatalf("type target = %#x", got)
			}
			return 0x3456
		},
		charmable: func(typeIndex uint16) int32 {
			observe("charmable")
			if typeIndex != 0x3456 {
				t.Fatalf("type index = %#x", typeIndex)
			}
			return 7
		},
		checkLimit: func(gotCaster uint64, guide int32) bool {
			observe("limit")
			if gotCaster != caster || guide != 7 {
				t.Fatalf("limit args = %#x/%d", gotCaster, guide)
			}
			return true
		},
		buffOff: func(gotTarget uint64, buff int32) int32 {
			observe("buff-off")
			if gotTarget != target || buff != charmPendingEnchant5011F0 {
				t.Fatalf("buff off = %#x/%d", gotTarget, buff)
			}
			return 0x1234
		},
		findParent: func(got uint64) uint64 {
			observe("parent")
			if got != target {
				t.Fatalf("parent target = %#x", got)
			}
			return uint64(0x7f031b62ffff)
		},
		clearOwner: func(got uint64) {
			observe("clear-owner")
			if got != target {
				t.Fatalf("clear owner target = %#x", got)
			}
		},
		setOwner: func(gotCaster, gotTarget uint64) {
			observe("set-owner")
			if gotCaster != caster || gotTarget != target {
				t.Fatalf("set owner = %#x/%#x", gotCaster, gotTarget)
			}
		},
		hasTeam: func(got uint64) bool {
			observe("has-team")
			if got != target {
				t.Fatalf("team target = %#x", got)
			}
			return true
		},
		loadNetCode: func(got uint64) uint32 {
			observe("net-code")
			if got != target {
				t.Fatalf("net-code target = %#x", got)
			}
			return 0x89abcdef
		},
		changeTeam: func(got uint64, code uint32) {
			observe("change-team")
			if got != target || code != 0x89abcdef {
				t.Fatalf("change team = %#x/%#x", got, code)
			}
		},
		loadMonsterUpdate: func(got uint64) uint64 {
			observe("monster-update")
			if got != target {
				t.Fatalf("monster update target = %#x", got)
			}
			return monsterUpdate
		},
		loadMonsterStatus: func(got uint64) uint32 {
			observe("monster-status")
			if got != monsterUpdate {
				t.Fatalf("monster update = %#x", got)
			}
			return status
		},
		storeMonsterStatus: func(got uint64, value uint32) {
			observe("store-monster-status")
			if got != monsterUpdate {
				t.Fatalf("monster status store = %#x", got)
			}
			status = value
		},
		gameFlag: func(mask uint32) bool {
			observe("game-flag")
			if mask != charmQuestGameFlag5013E0 {
				t.Fatalf("game flag = %#x", mask)
			}
			return true
		},
		typeHealthMax: func(typeIndex uint16) uint16 {
			observe("type-max")
			if typeIndex != 0x3456 {
				t.Fatalf("health type = %#x", typeIndex)
			}
			return 100
		},
		unitHP: func(got uint64) uint16 {
			observe("hp")
			if got != target {
				t.Fatalf("HP target = %#x", got)
			}
			return 90
		},
		loadMonsterDef: func(got uint64) uint64 {
			observe("monster-def")
			if got != monsterUpdate {
				t.Fatalf("definition update = %#x", got)
			}
			return monsterDef
		},
		monsterDefIsNil: func(got uint64) bool {
			observe("def-nil")
			return got == 0
		},
		monsterQuestMax: func(got uint64) uint16 {
			observe("quest-max")
			if got != monsterDef {
				t.Fatalf("quest definition = %#x", got)
			}
			return 70
		},
		storeHealthMax: func(got uint64, maximum uint16) {
			observe("health-max")
			if got != target {
				t.Fatalf("health store target = %#x", got)
			}
			healthMax = maximum
		},
		setHP: func(got uint64, value uint16) {
			observe("set-hp")
			if got != target {
				t.Fatalf("set HP target = %#x", got)
			}
			setHP = value
		},
		loadPlayerUpdate: func(got uint64) uint64 {
			observe("player-update")
			if got != caster {
				t.Fatalf("player update caster = %#x", got)
			}
			return playerUpdate
		},
		loadPlayer: func(got uint64) uint64 {
			observe("player")
			if got != playerUpdate {
				t.Fatalf("player update = %#x", got)
			}
			return player
		},
		loadSummonOrder: func(got uint64) uint32 {
			observe("summon-order")
			if got != player {
				t.Fatalf("order player = %#x", got)
			}
			return 0xfedcba98
		},
		loadPlayerIndex: func(got uint64) byte {
			observe("player-index")
			if got != player {
				t.Fatalf("index player = %#x", got)
			}
			return 17
		},
		orderUnit: func(gotCaster, gotTarget uint64, order uint32) {
			observe("order")
			if gotCaster != caster || gotTarget != target || order != 0xfedcba98 {
				t.Fatalf("order = %#x/%#x/%#x", gotCaster, gotTarget, order)
			}
		},
		reportAcquire: func(index byte, got uint64) {
			observe("acquire")
			if index != 17 || got != target {
				t.Fatalf("acquire = %d/%#x", index, got)
			}
		},
		markMinimap: func(index byte, got uint64, flags uint32) {
			observe("minimap")
			if index != 17 || got != target || flags != 1 {
				t.Fatalf("minimap = %d/%#x/%#x", index, got, flags)
			}
		},
		sendSimpleObject: func(index byte, got uint64) {
			observe("simple")
			if index != 17 || got != target {
				t.Fatalf("simple = %d/%#x", index, got)
			}
		},
		questSpawnCleanup: func(got uint64) {
			observe("cleanup")
			if got != target {
				t.Fatalf("cleanup target = %#x", got)
			}
		},
		loadSpell: func(got uint64) uint32 {
			observe("spell")
			if got != record {
				t.Fatalf("spell record = %#x", got)
			}
			return 0x4321
		},
		spellAudio: func(spell uint32, selector int32) int32 {
			observe("spell-audio")
			if spell != 0x4321 || selector != 1 {
				t.Fatalf("spell audio = %#x/%d", spell, selector)
			}
			return 777
		},
		audio: func(id int32, got uint64) {
			observe("audio")
			if id != 777 || got != caster {
				t.Fatalf("audio = %d/%#x", id, got)
			}
		},
	}

	if got := charmFinish5013E0(record, hooks); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	wantEvents := []string{
		"target", "flags", "caster", "distance", "deadline", "frame",
		"target", "subclass", "caster", "class", "type", "charmable", "caster", "limit",
		"target", "buff-off", "target", "parent", "caster",
		"target", "clear-owner", "target", "caster", "set-owner",
		"target", "has-team", "target", "net-code", "change-team",
		"target", "monster-update", "monster-status", "store-monster-status",
		"game-flag", "target", "type", "type-max", "target", "hp", "monster-def", "def-nil",
		"quest-max", "target", "health-max", "target", "set-hp",
		"caster", "class", "player-update", "player", "summon-order", "target", "order",
		"target", "subclass", "store-subclass", "target", "subclass", "store-subclass",
		"player", "player-index", "target", "acquire",
		"player", "player-index", "target", "minimap",
		"player", "player-index", "target", "simple",
		"game-flag", "target", "cleanup", "caster", "spell", "spell-audio", "audio",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %q, want exact oracle order %q", events, wantEvents)
	}
	if status != 0x92 {
		t.Fatalf("monster status = %#x, want 0x92", status)
	}
	if subclass != 0x1a0 {
		t.Fatalf("subclass = %#x, want 0x1a0", subclass)
	}
	if healthMax != 70 || setHP != 70 {
		t.Fatalf("health max/current = %d/%d, want 70/70", healthMax, setHP)
	}
}

func TestCharmCancel501690DisabledAndActiveReload(t *testing.T) {
	const (
		record      = uint64(0x100000101)
		disabled    = uint64(0x7f031b61e540)
		firstTarget = uint64(0x7f031b61e8a0)
		liveTarget  = uint64(0x7f031b61ef00)
	)
	var events []string
	disabledHooks := charmCancelHooks501690[uint64, uint64]{
		loadTarget: func(uint64) uint64 {
			events = append(events, "target")
			return disabled
		},
		loadObjectFlags: func(got uint64) uint32 {
			events = append(events, "flags")
			if got != disabled {
				t.Fatalf("disabled flags target = %#x", got)
			}
			return charmDisabledFlags5011F0
		},
		objectDword: func(got uint64) int32 {
			events = append(events, "dword")
			return int32(uint32(got))
		},
		buffOff: func(uint64, int32) int32 {
			t.Fatal("disabled target removed a buff")
			return 0
		},
	}
	if got, want := charmCancel501690(record, disabledHooks), int32(uint32(disabled&math.MaxUint32)); got != want {
		t.Fatalf("disabled result = %#x, want %#x", got, want)
	}
	if want := []string{"target", "flags", "dword"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("disabled events = %q, want %q", events, want)
	}

	events = nil
	loads := 0
	activeHooks := charmCancelHooks501690[uint64, uint64]{
		loadTarget: func(uint64) uint64 {
			events = append(events, "target")
			loads++
			if loads == 1 {
				return firstTarget
			}
			return liveTarget
		},
		loadObjectFlags: func(got uint64) uint32 {
			events = append(events, "flags")
			if got != firstTarget {
				t.Fatalf("active flags target = %#x", got)
			}
			return 0
		},
		buffOff: func(got uint64, buff int32) int32 {
			events = append(events, "buff")
			switch buff {
			case charmHeldEnchant501690:
				if got != firstTarget {
					t.Fatalf("held target = %#x", got)
				}
				return 1
			case charmPendingEnchant5011F0:
				if got != liveTarget {
					t.Fatalf("pending target = %#x", got)
				}
				return math.MinInt32
			default:
				t.Fatalf("unexpected buff %d", buff)
				return 0
			}
		},
	}
	if got := charmCancel501690(record, activeHooks); got != math.MinInt32 {
		t.Fatalf("active result = %d, want %d", got, int32(math.MinInt32))
	}
	if want := []string{"target", "flags", "buff", "target", "buff"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("active events = %q, want %q", events, want)
	}
}
