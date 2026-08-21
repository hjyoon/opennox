package server

import (
	"math"
	"reflect"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func TestMonsterInit4F0040NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantObjectType := uintptr(4)
	wantObjectSubclass := uintptr(12)
	wantObjectFlags := uintptr(16)
	wantObjectPosition := uintptr(56)
	wantObjectDirection := uintptr(124)
	wantObjectRadius := uintptr(176)
	wantObjectSpeed := uintptr(548)
	wantObjectHealth := uintptr(556)
	wantObjectUpdate := uintptr(748)
	wantUpdateSize := uintptr(2200)
	wantUpdateDef := uintptr(484)
	wantUpdateStackIndex := uintptr(544)
	wantUpdateStack := uintptr(552)
	wantUpdateAggression := uintptr(1304)
	wantUpdateSight := uintptr(1312)
	wantUpdateField332 := uintptr(1328)
	wantUpdateField333 := uintptr(1332)
	wantUpdateHealthScale := uintptr(1352)
	wantUpdateFlee := uintptr(1356)
	wantUpdateAction := uintptr(1360)
	wantUpdateStatus := uintptr(1440)
	wantActionSize := uintptr(24)
	wantActionArgs := uintptr(4)
	wantActionField5 := uintptr(20)
	wantMonsterDefSize := uintptr(248)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantObjectType = 8
		wantObjectSubclass = 16
		wantObjectFlags = 20
		wantObjectPosition = 60
		wantObjectDirection = 128
		wantObjectRadius = 180
		wantObjectSpeed = 608
		wantObjectHealth = 616
		wantObjectUpdate = 872
		wantUpdateSize = 2824
		wantUpdateDef = 496
		wantUpdateStackIndex = 564
		wantUpdateStack = 576
		wantUpdateAggression = 1912
		wantUpdateSight = 1920
		wantUpdateField332 = 1936
		wantUpdateField333 = 1940
		wantUpdateHealthScale = 1960
		wantUpdateFlee = 1964
		wantUpdateAction = 1968
		wantUpdateStatus = 2048
		wantActionSize = 48
		wantActionArgs = 8
		wantActionField5 = 40
		wantMonsterDefSize = 272
	}

	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), wantObjectType},
		{"Object.ObjSubClass", unsafe.Offsetof(Object{}.ObjSubClass), wantObjectSubclass},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantObjectFlags},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantObjectPosition},
		{"Object.Direction1", unsafe.Offsetof(Object{}.Direction1), wantObjectDirection},
		{"Object.Shape.Circle.R", unsafe.Offsetof(Object{}.Shape) + unsafe.Offsetof(Shape{}.Circle) + unsafe.Offsetof(Shape{}.Circle.R), wantObjectRadius},
		{"Object.SpeedBase", unsafe.Offsetof(Object{}.SpeedBase), wantObjectSpeed},
		{"Object.HealthData", unsafe.Offsetof(Object{}.HealthData), wantObjectHealth},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantObjectUpdate},
		{"MonsterUpdateData size", unsafe.Sizeof(MonsterUpdateData{}), wantUpdateSize},
		{"MonsterUpdateData.Direction94", unsafe.Offsetof(MonsterUpdateData{}.Direction94), 376},
		{"MonsterUpdateData.Pos95", unsafe.Offsetof(MonsterUpdateData{}.Pos95), 380},
		{"MonsterUpdateData.HealthGraph103", unsafe.Offsetof(MonsterUpdateData{}.HealthGraph103), 412},
		{"MonsterUpdateData.MonsterDef", unsafe.Offsetof(MonsterUpdateData{}.MonsterDef), wantUpdateDef},
		{"MonsterUpdateData.AIStackInd", unsafe.Offsetof(MonsterUpdateData{}.AIStackInd), wantUpdateStackIndex},
		{"MonsterUpdateData.AIStack", unsafe.Offsetof(MonsterUpdateData{}.AIStack), wantUpdateStack},
		{"MonsterUpdateData.Aggression", unsafe.Offsetof(MonsterUpdateData{}.Aggression), wantUpdateAggression},
		{"MonsterUpdateData.SightRange", unsafe.Offsetof(MonsterUpdateData{}.SightRange), wantUpdateSight},
		{"MonsterUpdateData.Field332", unsafe.Offsetof(MonsterUpdateData{}.Field332), wantUpdateField332},
		{"MonsterUpdateData.Field333", unsafe.Offsetof(MonsterUpdateData{}.Field333), wantUpdateField333},
		{"MonsterUpdateData.Field338", unsafe.Offsetof(MonsterUpdateData{}.Field338), wantUpdateHealthScale},
		{"MonsterUpdateData.FleeRange", unsafe.Offsetof(MonsterUpdateData{}.FleeRange), wantUpdateFlee},
		{"MonsterUpdateData.AIAction340", unsafe.Offsetof(MonsterUpdateData{}.AIAction340), wantUpdateAction},
		{"MonsterUpdateData.StatusFlags", unsafe.Offsetof(MonsterUpdateData{}.StatusFlags), wantUpdateStatus},
		{"AIStackItem size", unsafe.Sizeof(AIStackItem{}), wantActionSize},
		{"AIStackItem.Action", unsafe.Offsetof(AIStackItem{}.Action), 0},
		{"AIStackItem.Args", unsafe.Offsetof(AIStackItem{}.Args), wantActionArgs},
		{"AIStackItem.Field5", unsafe.Offsetof(AIStackItem{}.Field5), wantActionField5},
		{"MonsterDef size", unsafe.Sizeof(MonsterDef{}), wantMonsterDefSize},
		{"MonsterDef.MeleeAttackRange112", unsafe.Offsetof(MonsterDef{}.MeleeAttackRange112), 112},
		{"HealthData size", unsafe.Sizeof(HealthData{}), 20},
		{"HealthData.Cur", unsafe.Offsetof(HealthData{}.Cur), 0},
		{"HealthData.Field2", unsafe.Offsetof(HealthData{}.Field2), 2},
		{"HealthData.Max", unsafe.Offsetof(HealthData{}.Max), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func newMonsterInitNativeServer4F0040(t *testing.T) *Server {
	t.Helper()
	handle := atomic.AddUintptr(&serverLast, 1)
	s := &Server{handle: handle}
	s.Rand.Logic = prand.New(0)
	s.Types.byID = map[string]*ObjectType{
		"carnivorousplant": {ind: 7, id: "CarnivorousPlant"},
		"fishbig":          {ind: 8, id: "FishBig"},
		"fishsmall":        {ind: 9, id: "FishSmall"},
		"rat":              {ind: 10, id: "Rat"},
		"greenfrog":        {ind: 11, id: "GreenFrog"},
	}
	servers.Store(handle, s)
	t.Cleanup(s.Close)
	return s
}

func TestMonsterInit4F0040NativePlantState(t *testing.T) {
	s := newMonsterInitNativeServer4F0040(t)
	s.SetFrame(0x89abcdef)

	colors := [6]Color3{
		{R: 1, G: 2, B: 3}, {R: 4, G: 5, B: 6}, {R: 7, G: 8, B: 9},
		{R: 10, G: 11, B: 12}, {R: 13, G: 14, B: 15}, {R: 16, G: 17, B: 18},
	}
	update := &MonsterUpdateData{
		MonsterDef: &MonsterDef{MeleeAttackRange112: 12.5},
		AIStackInd: 0,
		Field332:   3,
		Field338:   1.5,
		StatusFlags: object.MonStatusCanCastSpells |
			object.MonStatusHoldYourGround |
			object.MonStatusAlwaysRun,
		Color: colors,
	}
	update.AIStack[0].Action = uint32(ai.ACTION_IDLE)
	health := &HealthData{Cur: 100, Field2: 1, Max: 100}
	unit := &Object{
		TypeInd:     7,
		ObjClass:    object.ClassMonster,
		ObjSubClass: object.SubClass(object.MonsterNPC),
		PosVec: types.Pointf{
			X: math.Float32frombits(0x7fc12345),
			Y: math.Float32frombits(0x80000000),
		},
		Direction1:   Dir16(0x8001),
		HealthData:   health,
		SpeedBase:    9,
		UpdateData:   unsafe.Pointer(update),
		serverHandle: s.handle,
	}
	unit.Shape.Circle.R = 3.25

	var healthCalls int
	s.MonsterInit4F0040(unit, MonsterInitRuntime4F0040{
		SetHealth: func(got *Object, value uint16) {
			healthCalls++
			if got != unit || value != 150 {
				t.Fatalf("SetHealth = %p/%d, want %p/150", got, value, unit)
			}
			got.HealthData.Cur = value
		},
	})

	if healthCalls != 1 || health.Cur != 150 || health.Field2 != 150 {
		t.Fatalf("health calls/state = %d/%+v", healthCalls, health)
	}
	for index, value := range update.HealthGraph103 {
		if value != 150 {
			t.Fatalf("health graph %d = %d", index, value)
		}
	}
	if update.AIAction340 != uint32(ai.ACTION_INVALID) || update.AIStackInd != 0 {
		t.Fatalf("AI action/index = %d/%d", update.AIAction340, update.AIStackInd)
	}
	item := &update.AIStack[0]
	if item.Action != uint32(ai.ACTION_GUARD) || uint32(item.Args[0]) != 0x7fc12345 ||
		uint32(item.Args[1]) != 0x80000000 || uint32(item.Args[2]) != 0xffff8001 {
		t.Fatalf("guard item = action %d args %#x/%#x/%#x", item.Action, item.Args[0], item.Args[1], item.Args[2])
	}
	if math.Float32bits(update.SightRange) != 0x41ce0000 || update.Direction94 != 0xffff8001 ||
		math.Float32bits(update.Pos95.X) != 0x7fc12345 || math.Float32bits(update.Pos95.Y) != 0x80000000 {
		t.Fatalf("sight/direction/position = %#08x/%#08x/%#08x/%#08x",
			math.Float32bits(update.SightRange), update.Direction94,
			math.Float32bits(update.Pos95.X), math.Float32bits(update.Pos95.Y))
	}
	if math.Float32bits(unit.SpeedBase) != 0x404ccccd || math.Float32bits(update.FleeRange) != 0 {
		t.Fatalf("speed/flee = %#08x/%#08x", math.Float32bits(unit.SpeedBase), math.Float32bits(update.FleeRange))
	}
	wantStatus := object.MonStatusCanCastSpells | object.MonStatusHoldYourGround |
		object.MonStatusAlwaysRun | object.MonStatusRunning
	if update.StatusFlags != wantStatus {
		t.Fatalf("status = %#08x, want %#08x", update.StatusFlags, wantStatus)
	}
	if !reflect.DeepEqual(update.Color, colors) {
		t.Fatalf("004F0040 changed NPC colors = %#v, want %#v", update.Color, colors)
	}
}

func TestMonsterInit4F0040NativeUsesSeparatePlantAndExactFishCaches(t *testing.T) {
	s := newMonsterInitNativeServer4F0040(t)
	s.Types.fast.plant = 99
	if got := s.Types.monsterInitPlantID4F0040(); got != 7 || s.Types.fast.monsterInitPlant != 7 || s.Types.fast.plant != 99 {
		t.Fatalf("plant caches = %d/%d/%d", got, s.Types.fast.monsterInitPlant, s.Types.fast.plant)
	}

	unit := &Object{TypeInd: 8}
	s.Types.fast.fishBig = 0
	s.Types.fast.fishSmall = 0
	if !s.Types.monsterInitIsFish4F0040(unit) || s.Types.fast.fishBig != 8 || s.Types.fast.fishSmall != 9 {
		t.Fatalf("initial fish caches = %d/%d", s.Types.fast.fishBig, s.Types.fast.fishSmall)
	}

	// sub_534B10 gates both lookups only on FishSmall. A pre-populated
	// FishSmall cache must therefore leave a missing FishBig cache missing.
	s.Types.fast.fishBig = 0
	s.Types.fast.fishSmall = 9
	if s.Types.monsterInitIsFish4F0040(unit) || s.Types.fast.fishBig != 0 {
		t.Fatalf("asymmetric fish cache = result true, caches %d/%d", s.Types.fast.fishBig, s.Types.fast.fishSmall)
	}
}

func TestMonsterInit4F0040NativeUsesExactLogicRandomSpeed(t *testing.T) {
	s := newMonsterInitNativeServer4F0040(t)
	update := &MonsterUpdateData{AIAction340: uint32(ai.ACTION_INVALID)}
	health := &HealthData{Cur: 9, Max: 10}
	unit := &Object{
		TypeInd:      99,
		ObjClass:     object.ClassMonster,
		SpeedBase:    8,
		HealthData:   health,
		UpdateData:   unsafe.Pointer(update),
		serverHandle: s.handle,
	}
	wantRandom := logicRandomFloat416030(
		prand.New(0),
		math.Float32frombits(monsterInitSpeedMinBits4F0040),
		math.Float32frombits(monsterInitSpeedMaxBits4F0040),
	)
	wantSpeed := monsterInitRandomSpeed4F0040(wantRandom, 8)

	s.MonsterInit4F0040(unit, MonsterInitRuntime4F0040{
		SetHealth: func(*Object, uint16) { t.Fatal("unexpected SetHealth") },
	})
	if index := s.Rand.Logic.Index(); index != 1 {
		t.Fatalf("logic RNG index = %d, want 1", index)
	}
	if math.Float32bits(unit.SpeedBase) != math.Float32bits(wantSpeed) {
		t.Fatalf("speed bits = %#08x, want %#08x", math.Float32bits(unit.SpeedBase), math.Float32bits(wantSpeed))
	}
}

func TestMonsterInit4F0040NativeHasNoEntryPointerGuard(t *testing.T) {
	s := newMonsterInitNativeServer4F0040(t)
	defer func() {
		if recover() == nil {
			t.Fatal("nil unit did not fault")
		}
	}()
	monsterInitNative4F0040(s, nil, MonsterInitRuntime4F0040{})
}
