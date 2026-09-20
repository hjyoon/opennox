package server

import (
	"encoding/binary"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

func TestPolypDieNative54CB10CachesCloudUpdateAndPreservesOrder(t *testing.T) {
	source := &Object{PosVec: types.Pointf{X: 17.5, Y: -9.25}}
	update := new(ToxicCloudUpdateData)
	replacement := new(ToxicCloudUpdateData)
	cloud := &Object{UpdateData: unsafe.Pointer(update)}
	var events []string

	PolypDieNative54CB10(source, PolypDeathRuntime54CB10{
		NewObjectByTypeID: func(id string) *Object {
			events = append(events, "new")
			if id != "ToxicCloud" {
				t.Fatalf("cloud type = %q, want ToxicCloud", id)
			}
			return cloud
		},
		CreateObjectAt: func(got, owner *Object, position types.Pointf) {
			events = append(events, "create")
			if got != cloud || owner != nil || position != source.PosVec {
				t.Fatalf("create = %p/%p/%+v, want %p/nil/%+v", got, owner, position, cloud, source.PosVec)
			}
			cloud.UpdateData = unsafe.Pointer(replacement)
		},
		BalanceFloat: func(key string) float64 {
			events = append(events, "balance")
			if key != "ToxicCloudLifetime" {
				t.Fatalf("balance key = %q", key)
			}
			return 2.5
		},
		TickRate: func() uint32 {
			events = append(events, "fps")
			return 3
		},
		Audio: func(id sound.ID, got *Object) {
			events = append(events, "audio")
			if id != sound.SoundPolypExplode || got != source || update.Duration != 8 {
				t.Fatalf("audio/state = %d/%p/%d, want PolypExplode/%p/8", id, got, update.Duration, source)
			}
		},
		DelayedDelete: func(got *Object) {
			events = append(events, "delete")
			if got != source {
				t.Fatalf("deleted = %p, want %p", got, source)
			}
		},
	})

	want := []string{"new", "create", "balance", "fps", "audio", "delete"}
	if !slices.Equal(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if update.Duration != 8 || replacement.Duration != 0 {
		t.Fatalf("cached/replacement duration = %d/%d, want 8/0", update.Duration, replacement.Duration)
	}
}

func TestPolypDieNative54CB10AllocationFailureStillFinishesDeath(t *testing.T) {
	source := new(Object)
	var events []string
	PolypDieNative54CB10(source, PolypDeathRuntime54CB10{
		NewObjectByTypeID: func(string) *Object {
			events = append(events, "new")
			return nil
		},
		CreateObjectAt: func(*Object, *Object, types.Pointf) { t.Fatal("CreateObjectAt called") },
		BalanceFloat:   func(string) float64 { t.Fatal("BalanceFloat called"); return 0 },
		TickRate:       func() uint32 { t.Fatal("TickRate called"); return 0 },
		Audio:          func(sound.ID, *Object) { events = append(events, "audio") },
		DelayedDelete:  func(*Object) { events = append(events, "delete") },
	})
	if want := []string{"new", "audio", "delete"}; !slices.Equal(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestMarkerDieNative54E460ClearsOnlyFirstSlotBeforeEffects(t *testing.T) {
	source := &Object{PosVec: types.Pointf{X: 4, Y: 8}}
	other := new(Object)
	update := &PlayerUpdateData{Field29: [4]*Object{other, source, source, other}}
	player := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	var events []string

	MarkerDieNative54E460(source, MarkerDeathRuntime54E460{
		FindOwnerChainPlayer: func(got *Object) *Object {
			events = append(events, "find")
			if got != source {
				t.Fatalf("owner lookup source = %p, want %p", got, source)
			}
			return player
		},
		PointFX: func(op netmsg.Op, position types.Pointf) {
			events = append(events, "fx")
			if op != netmsg.MSG_FX_SMOKE_BLAST || position != source.PosVec {
				t.Fatalf("effect = %v/%+v", op, position)
			}
			if update.Field29 != [4]*Object{other, nil, source, other} {
				t.Fatalf("marker slots at effect = %v", update.Field29)
			}
		},
		DelayedDelete: func(got *Object) {
			events = append(events, "delete")
			if got != source {
				t.Fatalf("deleted = %p, want %p", got, source)
			}
		},
	})
	if want := []string{"find", "fx", "delete"}; !slices.Equal(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestBoulderDieNative54E4B0CreatesRoundRobinDebris(t *testing.T) {
	source := &Object{PosVec: types.Pointf{X: 100, Y: 200}}
	parts := []*Object{{}, {}, {}}
	var createdTypes []string
	var createdParts []*Object
	var raised []float32
	var forces []float64
	var decay []uint32
	partIndex := uint32(11)
	integerCalls := 0
	deleted := false

	completed := BoulderDieNative54E4B0(source, BoulderDeathRuntime54E4B0{
		RandomInt: func(minimum, maximum int) int {
			integerCalls++
			switch {
			case minimum == 20 && maximum == 30:
				return 3
			case minimum == 45 && maximum == 75:
				return 50
			default:
				t.Fatalf("random integer range = %d..%d", minimum, maximum)
				return 0
			}
		},
		RandomFloat: func(minimum, maximum float32) float32 {
			switch {
			case minimum == 10 && maximum == 70:
				return 40
			case minimum == -2 && maximum == 0:
				return -1
			case minimum == 5 && maximum == 20:
				return 12.5
			default:
				t.Fatalf("random float range = %g..%g", minimum, maximum)
				return 0
			}
		},
		Audio: func(id sound.ID, got *Object) {
			if id != sound.SoundWallDestroyedStone || got != source {
				t.Fatalf("audio = %d/%p", id, got)
			}
		},
		PointFX: func(op netmsg.Op, position types.Pointf) {
			if op != netmsg.MSG_FX_SMOKE_BLAST || position != source.PosVec {
				t.Fatalf("effect = %v/%+v", op, position)
			}
		},
		NewObjectByTypeID: func(id string) *Object {
			createdTypes = append(createdTypes, id)
			return parts[len(createdTypes)-1]
		},
		RandomReachablePoint: func(radius float32, center types.Pointf) types.Pointf {
			if radius != 30 || center != source.PosVec {
				t.Fatalf("reachable point = %g/%+v", radius, center)
			}
			return types.Pointf{X: center.X + float32(len(createdParts)+1), Y: center.Y}
		},
		CreateObjectAt: func(part, owner *Object, position types.Pointf) {
			if owner != nil || position.Y != source.PosVec.Y {
				t.Fatalf("debris create = %p/%+v", owner, position)
			}
			createdParts = append(createdParts, part)
		},
		Raise: func(part *Object, height float32) {
			raised = append(raised, height)
		},
		ApplyForce: func(part *Object, origin types.Pointf, force float64) {
			if origin != source.PosVec {
				t.Fatalf("force origin = %+v, want %+v", origin, source.PosVec)
			}
			forces = append(forces, force)
		},
		DecaySetTime: func(part *Object, frames uint32) {
			decay = append(decay, frames)
		},
		TickRate:     func() uint32 { return 30 },
		PartIndex:    func() uint32 { return partIndex },
		SetPartIndex: func(index uint32) { partIndex = index },
		DelayedDelete: func(got *Object) {
			deleted = true
			if got != source {
				t.Fatalf("deleted = %p, want %p", got, source)
			}
		},
	})

	if !completed || !deleted {
		t.Fatalf("completion/deletion = %t/%t, want true/true", completed, deleted)
	}
	if want := []string{"SmallRock", "BigRock", "MediumRock"}; !slices.Equal(createdTypes, want) {
		t.Fatalf("debris types = %v, want %v", createdTypes, want)
	}
	wantHeights := []float32{5, 1, 3}
	for index, part := range parts {
		if part.Field27 != -1 || !part.ObjFlags.Has(object.FlagBouncy) || math.Float32frombits(part.Field29) != wantHeights[index] {
			t.Fatalf("part %d = field27:%g flags:%#x height:%g", index, part.Field27, part.ObjFlags, math.Float32frombits(part.Field29))
		}
	}
	if !slices.Equal(raised, []float32{40, 40, 40}) || !slices.Equal(forces, []float64{12.5, 12.5, 12.5}) ||
		!slices.Equal(decay, []uint32{1500, 1500, 1500}) {
		t.Fatalf("raise/force/decay = %v/%v/%v", raised, forces, decay)
	}
	if partIndex != 2 || integerCalls != 4 {
		t.Fatalf("part index/integer calls = %d/%d, want 2/4", partIndex, integerCalls)
	}
}

func TestBoulderDieNative54E4B0AllocationFailureKeepsSource(t *testing.T) {
	source := new(Object)
	part := new(Object)
	partIndex := uint32(0)
	factoryCalls := 0
	deleted := false
	completed := BoulderDieNative54E4B0(source, BoulderDeathRuntime54E4B0{
		RandomInt: func(minimum, maximum int) int {
			if minimum == 20 {
				return 2
			}
			return 45
		},
		RandomFloat: func(minimum, maximum float32) float32 { return minimum },
		Audio:       func(sound.ID, *Object) {},
		PointFX:     func(netmsg.Op, types.Pointf) {},
		NewObjectByTypeID: func(string) *Object {
			factoryCalls++
			if factoryCalls == 1 {
				return part
			}
			return nil
		},
		RandomReachablePoint: func(float32, types.Pointf) types.Pointf { return types.Pointf{} },
		CreateObjectAt:       func(*Object, *Object, types.Pointf) {},
		Raise:                func(*Object, float32) {},
		ApplyForce:           func(*Object, types.Pointf, float64) {},
		DecaySetTime:         func(*Object, uint32) {},
		TickRate:             func() uint32 { return 30 },
		PartIndex:            func() uint32 { return partIndex },
		SetPartIndex:         func(index uint32) { partIndex = index },
		DelayedDelete:        func(*Object) { deleted = true },
	})
	if completed || deleted || factoryCalls != 2 || partIndex != 1 {
		t.Fatalf("result = completed:%t deleted:%t factories:%d index:%d", completed, deleted, factoryCalls, partIndex)
	}
}

func TestMonsterGeneratorDieNative54E630PreservesEffectsAndQuestCredit(t *testing.T) {
	update := &MonsterGenUpdateData{Field56: 0x01020304, FuncInd60: 0x10203040}
	attribution := new(Object)
	source := &Object{
		ObjClass:   object.ClassMonsterGenerator,
		PosVec:     types.Pointf{X: 12.5, Y: -3.5},
		Obj130:     attribution,
		UpdateData: unsafe.Pointer(update),
	}
	player := &Player{field4668: 9, field4692: 0x20}
	playerUpdate := &PlayerUpdateData{Player: player}
	playerUnit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(playerUpdate)}
	destroyed := new(Object)
	var events []string
	var fx []byte

	MonsterGeneratorDieNative54E630(source, MonsterGeneratorDeathRuntime54E630{
		Frame: func() uint32 {
			events = append(events, "frame")
			return 77
		},
		SetQuestTimer: func(frame uint32) {
			events = append(events, "timer")
			if frame != 77 {
				t.Fatalf("quest timer = %d, want 77", frame)
			}
		},
		SetQuestMode: func(value int32) {
			events = append(events, "mode")
			if value != 0 {
				t.Fatalf("quest mode = %d, want 0", value)
			}
		},
		ScriptCallback: func(callback *ScriptCallback, caller, trigger *Object, event ScriptEventType) {
			events = append(events, "script")
			wantCallback := (*ScriptCallback)(unsafe.Pointer(&update.Field56))
			if callback != wantCallback || callback.Flags != 0x01020304 || callback.Func != 0x10203040 ||
				caller != attribution || trigger != source || event != NoxEventGeneratorDead {
				t.Fatalf("script callback = %p/%#x/%#x/%p/%p/%v", callback, callback.Flags, callback.Func, caller, trigger, event)
			}
		},
		Audio: func(id sound.ID, got *Object) {
			events = append(events, "audio")
			if id != sound.SoundMonsterGeneratorDie || got != source {
				t.Fatalf("audio = %d/%p", id, got)
			}
		},
		SendFX: func(position types.Pointf, packet []byte) {
			events = append(events, "fx")
			if position != source.PosVec {
				t.Fatalf("FX position = %+v, want %+v", position, source.PosVec)
			}
			fx = append([]byte(nil), packet...)
		},
		QuestMode: func() bool {
			events = append(events, "quest")
			return true
		},
		FindOwnerChainPlayer: func(got *Object) *Object {
			events = append(events, "find")
			if got != attribution {
				t.Fatalf("attribution = %p, want %p", got, attribution)
			}
			return playerUnit
		},
		NewObjectByTypeID: func(id string) *Object {
			events = append(events, "new")
			if id != "DestroyedGenerator" {
				t.Fatalf("destroyed type = %q", id)
			}
			return destroyed
		},
		CreateObjectAt: func(got, owner *Object, position types.Pointf) {
			events = append(events, "create")
			if got != destroyed || owner != nil || position != source.PosVec {
				t.Fatalf("create = %p/%p/%+v", got, owner, position)
			}
		},
		DelayedDelete: func(got *Object) {
			events = append(events, "delete")
			if got != source {
				t.Fatalf("deleted = %p, want %p", got, source)
			}
		},
	})

	wantEvents := []string{"frame", "timer", "mode", "script", "audio", "fx", "quest", "find", "new", "create", "delete"}
	if !slices.Equal(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	wantFX := []byte{0xf0, 25, 12, 0, 0xfc, 0xff, 200}
	if !slices.Equal(fx, wantFX) {
		t.Fatalf("FX packet = %v, want %v", fx, wantFX)
	}
	if player.field4668 != 10 || player.field4692 != 0x28 {
		t.Fatalf("quest credit = %d/%#x, want 10/0x28", player.field4668, player.field4692)
	}
}

func TestRecordMonsterGeneratorDestroyed4D61B0RejectsDestroyedPlayer(t *testing.T) {
	player := &Player{field4668: 4, field4692: 1}
	update := &PlayerUpdateData{Player: player}
	unit := &Object{
		ObjClass:   object.ClassPlayer,
		ObjFlags:   object.FlagDestroyed,
		UpdateData: unsafe.Pointer(update),
	}
	recordMonsterGeneratorDestroyed4D61B0(unit)
	if player.field4668 != 4 || player.field4692 != 1 {
		t.Fatalf("destroyed player credit changed to %d/%#x", player.field4668, player.field4692)
	}
}

func TestMonsterGeneratorBreakFXPacket523200UsesNearestEven(t *testing.T) {
	packet := monsterGeneratorBreakFXPacket523200(types.Pointf{X: 2.5, Y: -1.5}, 0xa5)
	if packet[0] != 0xf0 || packet[1] != 25 || binary.LittleEndian.Uint16(packet[2:4]) != 2 ||
		binary.LittleEndian.Uint16(packet[4:6]) != uint16(0xfffe) || packet[6] != 0xa5 {
		t.Fatalf("packet = %v", packet)
	}
}
