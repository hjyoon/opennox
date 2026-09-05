package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestSpellGestureCancelNative4FE680Layouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantObjectClass := uintptr(8)
	wantObjectTeam := uintptr(48)
	wantObjectPosition := uintptr(56)
	wantObjectUpdate := uintptr(748)
	wantCasting := uintptr(188)
	wantCastStart := uintptr(216)
	wantUpdatePlayer := uintptr(276)
	wantPlayerIndex := uintptr(2064)
	wantEntitySize := uintptr(60)
	wantEntityObject := uintptr(4)
	wantEntityNext := uintptr(52)
	wantEntityPrev := uintptr(56)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantObjectClass = 12
		wantObjectTeam = 52
		wantObjectPosition = 60
		wantObjectUpdate = 872
		wantCasting = 240
		wantCastStart = 268
		wantUpdatePlayer = 336
		wantPlayerIndex = 2068
		wantEntitySize = 80
		wantEntityObject = 8
		wantEntityNext = 64
		wantEntityPrev = 72
	}

	for _, check := range []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantObjectClass},
		{"Object.TeamVal", unsafe.Offsetof(Object{}.TeamVal), wantObjectTeam},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantObjectPosition},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantObjectUpdate},
		{"PlayerUpdateData.Field47_0", unsafe.Offsetof(PlayerUpdateData{}.Field47_0), wantCasting},
		{"PlayerUpdateData.SpellCastStart", unsafe.Offsetof(PlayerUpdateData{}.SpellCastStart), wantCastStart},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantUpdatePlayer},
		{"Player.PlayerInd", unsafe.Offsetof(Player{}.PlayerInd), wantPlayerIndex},
		{"MagicEntityClass size", unsafe.Sizeof(MagicEntityClass{}), wantEntitySize},
		{"MagicEntityClass.Obj4", unsafe.Offsetof(MagicEntityClass{}.Obj4), wantEntityObject},
		{"MagicEntityClass.Next52", unsafe.Offsetof(MagicEntityClass{}.Next52), wantEntityNext},
		{"MagicEntityClass.Prev56", unsafe.Offsetof(MagicEntityClass{}.Prev56), wantEntityPrev},
		{"ObjectTeam size", unsafe.Sizeof(ObjectTeam{}), 8},
		{"ObjectTeam.ID", unsafe.Offsetof(ObjectTeam{}.ID), 4},
	} {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}

	unit := new(Object)
	if got, want := spellGestureCancelObjectTeam4FE680(unit), &unit.TeamVal; got != want {
		t.Fatalf("inline team pointer = %p, want %p", got, want)
	}
}

func TestSpellGestureCancelNative4FE680PreservesPointersAndFields(t *testing.T) {
	player := &Player{PlayerInd: 0xe1}
	update := &PlayerUpdateData{
		Field47_0:      0xfe,
		SpellCastStart: 0x89abcdef,
		Player:         player,
	}
	source := &Object{
		TeamVal: ObjectTeam{ID: 3},
		PosVec:  types.Ptf(0, 0),
	}
	target := &Object{
		ObjClass:   object.ClassPlayer | 0x80000000,
		TeamVal:    ObjectTeam{ID: 7},
		PosVec:     types.Ptf(3, 4),
		UpdateData: unsafe.Pointer(update),
	}
	entity := &MagicEntityClass{Obj4: target}
	head := entity

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]uintptr{
			"player": uintptr(unsafe.Pointer(player)),
			"update": uintptr(unsafe.Pointer(update)),
			"source": uintptr(unsafe.Pointer(source)),
			"target": uintptr(unsafe.Pointer(target)),
			"entity": uintptr(unsafe.Pointer(entity)),
		} {
			if pointer <= math.MaxUint32 {
				t.Fatalf("%s pointer %#x does not exercise the native-width path", name, pointer)
			}
		}
	}

	var (
		comparedFirst, comparedSecond *ObjectTeam
		mapSource, mapTarget          *Object
		informedIndex, informedCode   uint8
		informedResult                int32
		audioID, audioKind            int32
		audioObject                   *Object
		audioCode                     uint32
		stateObject                   *Object
		state                         int32
		allocatorLoads                int
		freed                         *MagicEntityClass
	)
	deps := spellGestureCancelNativeDeps4FE680{
		compareTeams: func(first, second *ObjectTeam) int32 {
			comparedFirst, comparedSecond = first, second
			return 0
		},
		mapCheck: func(gotSource, gotTarget *Object) int32 {
			mapSource, mapTarget = gotSource, gotTarget
			return 1
		},
		informResult: func(index, code uint8, result int32) {
			informedIndex, informedCode, informedResult = index, code, result
		},
		audioEvent: func(id int32, object *Object, kind int32, code uint32) {
			audioID, audioObject, audioKind, audioCode = id, object, kind, code
		},
		setPlayerState: func(object *Object, gotState int32) {
			stateObject, state = object, gotState
		},
		loadHead: func() *MagicEntityClass {
			return head
		},
		storeHead: func(value *MagicEntityClass) {
			head = value
		},
		loadAllocator: func() SpellGestureCancelAllocator4FE680 {
			allocatorLoads++
			if head != nil || update.SpellCastStart != 0 || update.Field47_0 != 0 {
				t.Fatalf("allocator load observed head/update = %p/%#x/%#x", head, update.SpellCastStart, update.Field47_0)
			}
			return func(value *MagicEntityClass) {
				freed = value
			}
		},
	}

	spellGestureCancelNative4FE680(source, 6, deps)
	if comparedFirst != &source.TeamVal || comparedSecond != &target.TeamVal {
		t.Fatalf("team pointers = %p/%p, want %p/%p", comparedFirst, comparedSecond, &source.TeamVal, &target.TeamVal)
	}
	if mapSource != source || mapTarget != target {
		t.Fatalf("map pointers = %p/%p, want %p/%p", mapSource, mapTarget, source, target)
	}
	if informedIndex != player.PlayerInd || informedCode != 0 || informedResult != spellGestureCancelResult4FE680 {
		t.Fatalf("inform = %#x/%#x/%d", informedIndex, informedCode, informedResult)
	}
	if audioID != spellGestureCancelFizzle4FE680 || audioObject != target || audioKind != 0 || audioCode != 0 {
		t.Fatalf("audio = %d/%p/%d/%#x", audioID, audioObject, audioKind, audioCode)
	}
	if stateObject != target || state != spellGestureCancelState4FE680 {
		t.Fatalf("state = %p/%d", stateObject, state)
	}
	if update.SpellCastStart != 0 || update.Field47_0 != 0 || update.Player != player {
		t.Fatalf("update = start %#x casting %#x player %p", update.SpellCastStart, update.Field47_0, update.Player)
	}
	if head != nil || allocatorLoads != 1 || freed != entity || entity.Obj4 != target {
		t.Fatalf("queue = head %p allocator loads %d freed %p entity target %p", head, allocatorLoads, freed, entity.Obj4)
	}
}

func TestSpellGestureCancelNative4FE680ReloadsLiveObjectPointers(t *testing.T) {
	player := &Player{PlayerInd: 9}
	update := &PlayerUpdateData{Player: player}
	source := &Object{TeamVal: ObjectTeam{ID: 1}}
	initial := &Object{ObjClass: object.ClassPlayer, TeamVal: ObjectTeam{ID: 2}}
	distance := &Object{PosVec: types.Ptf(3, 4)}
	effect := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	audioTarget := new(Object)
	stateTarget := new(Object)
	entity := &MagicEntityClass{Obj4: initial}
	head := entity

	var mapTarget, audioObject, stateObject *Object
	deps := spellGestureCancelNativeDeps4FE680{
		compareTeams: func(first, second *ObjectTeam) int32 {
			if first != &source.TeamVal || second != &initial.TeamVal {
				t.Fatalf("team pointers = %p/%p", first, second)
			}
			entity.Obj4 = distance
			return 0
		},
		mapCheck: func(_ *Object, target *Object) int32 {
			mapTarget = target
			entity.Obj4 = effect
			return 1
		},
		informResult: func(uint8, uint8, int32) {
			entity.Obj4 = audioTarget
		},
		audioEvent: func(_ int32, target *Object, _ int32, _ uint32) {
			audioObject = target
			entity.Obj4 = stateTarget
		},
		setPlayerState: func(target *Object, _ int32) {
			stateObject = target
		},
		loadHead: func() *MagicEntityClass { return head },
		storeHead: func(value *MagicEntityClass) {
			head = value
		},
		loadAllocator: func() SpellGestureCancelAllocator4FE680 {
			return func(*MagicEntityClass) {}
		},
	}

	spellGestureCancelNative4FE680(source, 6, deps)
	if mapTarget != distance || audioObject != audioTarget || stateObject != stateTarget || head != nil {
		t.Fatalf("live targets/head = map %p audio %p state %p head %p", mapTarget, audioObject, stateObject, head)
	}
}
