package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	playerlib "github.com/opennox/libs/player"
)

func defaultPlayerReadValuesNativeDeps4EEDC0() playerReadValuesNativeDeps4EEDC0 {
	return playerReadValuesNativeDeps4EEDC0{
		gameFlagsCheck: func(uint32) int32 { return 0 },
		setHP: func(unit *Object, value uint16) {
			unit.HealthData.Cur = value
		},
		soloMode:       func() int32 { return 0 },
		abilityGiveAll: func(*Object, int8, int32) {},
		protectInt:     func(uint32, uint32) {},
		protectUint16:  func(uint32, uint16) {},
		wideLen: func(info *PlayerInfo) uint32 {
			return uint32(len([]rune(info.Name())))
		},
		protectName: func(*PlayerInfo, uint32, uint32) int32 { return 0 },
	}
}

func TestPlayerReadValues4EEDC0NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantMass := uintptr(120)
	wantCarry := uintptr(490)
	wantFirstItem := uintptr(504)
	wantSpeedBase := uintptr(548)
	wantHealth := uintptr(556)
	wantUpdate := uintptr(748)
	wantUpdateSize := uintptr(556)
	wantUpdatePlayer := uintptr(276)
	wantPlayerSize := uintptr(4828)
	wantInfo := uintptr(2185)
	wantCapacityWord := uintptr(3652)
	wantOverweight := uintptr(3656)
	wantLevel := uintptr(3684)
	wantHPToken := uintptr(4592)
	wantManaToken := uintptr(4600)
	wantSpeedToken := uintptr(4620)
	wantStrengthToken := uintptr(4624)
	wantNameToken := uintptr(4628)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantMass = 124
		wantCarry = 518
		wantFirstItem = 544
		wantSpeedBase = 608
		wantHealth = 616
		wantUpdate = 872
		wantUpdateSize = 640
		wantUpdatePlayer = 320
		wantPlayerSize = 6160
		wantInfo = 2189
		wantCapacityWord = 4948
		wantOverweight = 4952
		wantLevel = 4980
		wantHPToken = 5896
		wantManaToken = 5904
		wantSpeedToken = 5924
		wantStrengthToken = 5928
		wantNameToken = 5932
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.Mass", unsafe.Offsetof(Object{}.Mass), wantMass},
		{"Object.CarryCapacity", unsafe.Offsetof(Object{}.CarryCapacity), wantCarry},
		{"Object.InvFirstItem", unsafe.Offsetof(Object{}.InvFirstItem), wantFirstItem},
		{"Object.SpeedBase", unsafe.Offsetof(Object{}.SpeedBase), wantSpeedBase},
		{"Object.HealthData", unsafe.Offsetof(Object{}.HealthData), wantHealth},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantUpdateSize},
		{"PlayerUpdateData.ManaCur", unsafe.Offsetof(PlayerUpdateData{}.ManaCur), 4},
		{"PlayerUpdateData.ManaMax", unsafe.Offsetof(PlayerUpdateData{}.ManaMax), 8},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantUpdatePlayer},
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.info", unsafe.Offsetof(Player{}.info), wantInfo},
		{"Player.field3652", unsafe.Offsetof(Player{}.field3652), wantCapacityWord},
		{"Player.Field3656", unsafe.Offsetof(Player{}.Field3656), wantOverweight},
		{"Player.Level", unsafe.Offsetof(Player{}.Level), wantLevel},
		{"Player.ProtUnitHPMax", unsafe.Offsetof(Player{}.ProtUnitHPMax), wantHPToken},
		{"Player.ProtUnitManaMax", unsafe.Offsetof(Player{}.ProtUnitManaMax), wantManaToken},
		{"Player.ProtPlayerField2235", unsafe.Offsetof(Player{}.ProtPlayerField2235), wantSpeedToken},
		{"Player.ProtPlayerField2239", unsafe.Offsetof(Player{}.ProtPlayerField2239), wantStrengthToken},
		{"Player.ProtPlayerOrigName", unsafe.Offsetof(Player{}.ProtPlayerOrigName), wantNameToken},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
	player := &Player{}
	base := uintptr(unsafe.Pointer(player))
	initialized := uintptr(unsafe.Pointer(playerReadValuesInitializedPtr4EEDC0(player)))
	if initialized-base != wantInfo-1 {
		t.Errorf("Player initialized byte = %d, want %d", initialized-base, wantInfo-1)
	}
}

func TestPlayerReadValuesNative4EEDC0BindsNativeState(t *testing.T) {
	players := &serverPlayers{}
	players.Stats.Base = ClassStats{Health: 100, Mana: 80, Speed: 400, Strength: 50}
	players.Stats.Warrior = ClassStats{Health: 200, Mana: 100, Speed: 500, Strength: 100}
	players.Stats.Wizard = ClassStats{Health: 190, Mana: 170, Speed: 490, Strength: 95}

	player := &Player{
		Level:               5,
		field3652:           0xabcd0000,
		ProtPlayerField2239: 0x11111111,
		ProtPlayerField2235: 0x22222222,
		ProtUnitManaMax:     0x33333333,
		ProtUnitHPMax:       0x44444444,
		ProtPlayerOrigName:  0x55555555,
	}
	player.Info().SetPlayerClass(playerlib.Wizard)
	player.Info().SetName("Alpha")
	update := &PlayerUpdateData{Player: player}
	health := &HealthData{}
	item2 := &Object{Weight: 250}
	item1 := &Object{Weight: 200, InvNextItem: item2}
	unit := &Object{UpdateData: unsafe.Pointer(update), HealthData: health, InvFirstItem: item1}

	type protectCall struct {
		token uint32
		value uint32
	}
	var (
		abilities []struct {
			unit   *Object
			count  int8
			reward int32
		}
		protects []protectCall
		names    []struct {
			info        *PlayerInfo
			size, token uint32
		}
	)
	deps := defaultPlayerReadValuesNativeDeps4EEDC0()
	deps.abilityGiveAll = func(got *Object, count int8, reward int32) {
		abilities = append(abilities, struct {
			unit   *Object
			count  int8
			reward int32
		}{got, count, reward})
	}
	deps.protectInt = func(token, value uint32) {
		protects = append(protects, protectCall{token, value})
	}
	deps.protectUint16 = func(token uint32, value uint16) {
		protects = append(protects, protectCall{token, uint32(value)})
	}
	deps.protectName = func(info *PlayerInfo, size, token uint32) int32 {
		names = append(names, struct {
			info        *PlayerInfo
			size, token uint32
		}{info, size, token})
		return -0x1234567
	}

	result := playerReadValuesNative4EEDC0(unit, -7, players, deps)
	if result != -0x1234567 {
		t.Fatalf("result = %d, want %d", result, -0x1234567)
	}
	if health.Cur != 140 || health.Max != 140 || update.ManaCur != 120 || update.ManaMax != 120 {
		t.Fatalf("health/mana = %d/%d/%d/%d, want 140/140/120/120", health.Cur, health.Max, update.ManaCur, update.ManaMax)
	}
	if player.Info().Field2239() != 70 || player.Info().Field2235() != 440 {
		t.Fatalf("strength/speed = %d/%d, want 70/440", player.Info().Field2239(), player.Info().Field2235())
	}
	if math.Float32bits(unit.SpeedBase) != math.Float32bits(playerReadValuesScaleSpeedExtended4EEDC0(440)) || math.Float32bits(unit.Mass) != math.Float32bits(24) {
		t.Fatalf("speed base/mass bits = %#x/%#x", math.Float32bits(unit.SpeedBase), math.Float32bits(unit.Mass))
	}
	if player.field3652 != 0xabcd0985 || unit.CarryCapacity != 2437 || player.Field3656 != 0 {
		t.Fatalf("capacity backing/carry/overweight = %#x/%d/%d", player.field3652, unit.CarryCapacity, player.Field3656)
	}
	if got := *playerReadValuesInitializedPtr4EEDC0(player); got != 1 {
		t.Fatalf("initialized byte = %d, want 1", got)
	}
	if len(abilities) != 0 {
		t.Fatalf("wizard abilities = %#v, want none", abilities)
	}
	wantProtects := []protectCall{
		{0x11111111, 70}, {0x22222222, 440}, {0x33333333, 120}, {0x44444444, 140},
	}
	if !reflect.DeepEqual(protects, wantProtects) {
		t.Fatalf("protect calls = %#v, want %#v", protects, wantProtects)
	}
	if len(names) != 1 || names[0].info != player.Info() || names[0].size != 10 || names[0].token != 0x55555555 {
		t.Fatalf("name protection = %#v", names)
	}
}

func TestPlayerReadValuesNative4EEDC0UsesSignedLevelForAbility(t *testing.T) {
	players := &serverPlayers{}
	players.Stats.Base = ClassStats{Health: 1, Mana: 1, Speed: 1, Strength: 1}
	players.Stats.Warrior = ClassStats{Health: 1, Mana: 1, Speed: 1, Strength: 1}
	player := &Player{Level: 200}
	player.Info().SetPlayerClass(playerlib.Warrior)
	player.Info().SetName("W")
	update := &PlayerUpdateData{Player: player}
	unit := &Object{UpdateData: unsafe.Pointer(update), HealthData: &HealthData{}}
	deps := defaultPlayerReadValuesNativeDeps4EEDC0()
	var gotCount int8
	var gotReward int32
	deps.abilityGiveAll = func(got *Object, count int8, reward int32) {
		if got != unit {
			t.Fatalf("ability unit = %p, want %p", got, unit)
		}
		gotCount, gotReward = count, reward
	}
	playerReadValuesNative4EEDC0(unit, 0x76543210, players, deps)
	if gotCount != -56 || gotReward != 0x76543210 {
		t.Fatalf("ability count/reward = %d/%#x, want -56/0x76543210", gotCount, gotReward)
	}
}

func TestPlayerReadValuesNative4EEDC0HasNoEntryPointerGuards(t *testing.T) {
	deps := defaultPlayerReadValuesNativeDeps4EEDC0()
	players := &serverPlayers{}
	nilPlayerUpdate := &PlayerUpdateData{}
	for _, unit := range []*Object{nil, {}, {UpdateData: unsafe.Pointer(nilPlayerUpdate)}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("unit %p did not fault", unit)
				}
			}()
			playerReadValuesNative4EEDC0(unit, 0, players, deps)
		}()
	}
}

func TestPlayerReadValues4EEDC0ConversionHelpers(t *testing.T) {
	for _, tc := range []struct {
		value float32
		want  int32
	}{
		{0.5, 0}, {1.5, 2}, {2.5, 2}, {-1.5, -2},
		{math.Float32frombits(0x4effffff), 2147483520},
		{math.Float32frombits(0x4f000000), math.MinInt32},
		{float32(math.Inf(1)), math.MinInt32},
		{float32(math.NaN()), math.MinInt32},
	} {
		if got := playerReadValuesRound4EEDC0(tc.value); got != tc.want {
			t.Errorf("round(%08x) = %d, want %d", math.Float32bits(tc.value), got, tc.want)
		}
	}
	for _, tc := range []struct {
		value float64
		want  int64
	}{
		{1.9, 1}, {-1.9, -1}, {math.Ldexp(1, 63), math.MinInt64},
		{math.Inf(1), math.MinInt64}, {math.NaN(), math.MinInt64},
	} {
		if got := playerReadValuesTruncInt64_4EEDC0(tc.value); got != tc.want {
			t.Errorf("trunc(%016x) = %d, want %d", math.Float64bits(tc.value), got, tc.want)
		}
	}
}
