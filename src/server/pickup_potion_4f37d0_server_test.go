package server

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	playerlib "github.com/opennox/libs/player"
)

func defaultPickupPotionNativeDeps4F37D0() pickupPotionNativeDeps4F37D0 {
	return pickupPotionNativeDeps4F37D0{
		gameFlag:          func(uint32) int32 { return 0 },
		playerClassCanUse: func(*Object, uint8) int32 { return 1 },
		classFailure:      func(*Object, string, uint8) {},
		audio:             func(uint32, *Object, int32, uint32) {},
		playerState:       func(*Object) int32 { return 0 },
		healthMultiplier:  func(uint8) float32 { return 1 },
		adjustHealth:      func(*Object, int32) {},
		manaMultiplier:    func(uint8) float32 { return 1 },
		addMana:           func(*Object, int32) {},
		removePoison:      func(*Object) {},
		spellAudio:        func(int32, int32) uint32 { return 0 },
		delayedDelete:     func(*Object) {},
		decay:             func(*Object) {},
		defaultPickup:     func(*Object, *Object, int32, int32) int32 { return 0 },
	}
}

func TestPickupPotion4F37D0NativeLayouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantType := uintptr(4)
	wantClass := uintptr(8)
	wantSubClass := uintptr(12)
	wantNetCode := uintptr(36)
	wantPoison := uintptr(540)
	wantHealth := uintptr(556)
	wantUseData := uintptr(736)
	wantUpdate := uintptr(748)
	wantUpdateSize := uintptr(556)
	wantUpdatePlayer := uintptr(276)
	wantPlayerSize := uintptr(4828)
	wantPlayerInfo := uintptr(2185)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantType = 8
		wantClass = 12
		wantSubClass = 16
		wantNetCode = 40
		wantPoison = 600
		wantHealth = 616
		wantUseData = 848
		wantUpdate = 872
		wantUpdateSize = 640
		wantUpdatePlayer = 320
		wantPlayerSize = 6160
		wantPlayerInfo = 2189
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), wantType},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjSubClass", unsafe.Offsetof(Object{}.ObjSubClass), wantSubClass},
		{"Object.NetCode", unsafe.Offsetof(Object{}.NetCode), wantNetCode},
		{"Object.Poison540", unsafe.Offsetof(Object{}.Poison540), wantPoison},
		{"Object.HealthData", unsafe.Offsetof(Object{}.HealthData), wantHealth},
		{"Object.UseData", unsafe.Offsetof(Object{}.UseData), wantUseData},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"UseDataPtr size", unsafe.Sizeof(UseDataPtr{}), unsafe.Sizeof(uintptr(0))},
		{"PotionUseData size", unsafe.Sizeof(PotionUseData{}), 4},
		{"PotionUseData.Value", unsafe.Offsetof(PotionUseData{}.Value), 0},
		{"HealthData size", unsafe.Sizeof(HealthData{}), 20},
		{"HealthData.Cur", unsafe.Offsetof(HealthData{}.Cur), 0},
		{"HealthData.Max", unsafe.Offsetof(HealthData{}.Max), 4},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantUpdateSize},
		{"PlayerUpdateData.ManaCur", unsafe.Offsetof(PlayerUpdateData{}.ManaCur), 4},
		{"PlayerUpdateData.ManaMax", unsafe.Offsetof(PlayerUpdateData{}.ManaMax), 8},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantUpdatePlayer},
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.info", unsafe.Offsetof(Player{}.info), wantPlayerInfo},
		{"Player class byte", unsafe.Offsetof(Player{}.info) + unsafe.Offsetof(PlayerInfo{}.playerClass), wantPlayerInfo + 66},
		{"callback result width", unsafe.Sizeof(int32(0)), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestPickupPotionScale4F37D0X87Float32AndRoundToEven(t *testing.T) {
	for _, tc := range []struct {
		name       string
		base       int32
		multiplier float32
		want       int32
	}{
		{name: "positive-half-up-even", base: 3, multiplier: 0.5, want: 2},
		{name: "positive-half-down-even", base: 5, multiplier: 0.5, want: 2},
		{name: "negative-half-away-even", base: -3, multiplier: 0.5, want: -2},
		{name: "negative-half-toward-even", base: -5, multiplier: 0.5, want: -2},
		{name: "float32-spill", base: 16_777_217, multiplier: 1, want: 16_777_216},
		{name: "positive-overflow-indefinite", base: math.MaxInt32, multiplier: 1, want: math.MinInt32},
		{name: "negative-boundary", base: math.MinInt32, multiplier: 1, want: math.MinInt32},
		{name: "positive-infinity", base: 1, multiplier: float32(math.Inf(1)), want: math.MinInt32},
		{name: "negative-infinity", base: 1, multiplier: float32(math.Inf(-1)), want: math.MinInt32},
		{name: "nan", base: 1, multiplier: math.Float32frombits(0x7fc00001), want: math.MinInt32},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := pickupPotionScale4F37D0(tc.base, tc.multiplier); got != tc.want {
				t.Fatalf("scale(%d,%08x) = %d, want %d", tc.base, math.Float32bits(tc.multiplier), got, tc.want)
			}
		})
	}
}

func TestPickupPotionNative4F37D0UsesNativePointersFieldsAndClassScaling(t *testing.T) {
	player := new(Player)
	player.Info().SetPlayerClass(playerlib.Wizard)
	update := &PlayerUpdateData{ManaCur: 3, ManaMax: 100, Player: player}
	health := &HealthData{Cur: 4, Max: 100}
	owner := &Object{
		ObjClass:   object.ClassPlayer,
		Poison540:  1,
		HealthData: health,
		UpdateData: unsafe.Pointer(update),
	}
	use := &PotionUseData{Value: 5}
	potion := &Object{ObjSubClass: object.SubClass(0x70)}
	potion.UseData.SetPtr(unsafe.Pointer(use))
	if unsafe.Sizeof(uintptr(0)) == 8 &&
		(uintptr(unsafe.Pointer(owner)) <= math.MaxUint32 ||
			uintptr(unsafe.Pointer(potion)) <= math.MaxUint32 ||
			uintptr(unsafe.Pointer(update)) <= math.MaxUint32 ||
			uintptr(unsafe.Pointer(player)) <= math.MaxUint32 ||
			uintptr(unsafe.Pointer(use)) <= math.MaxUint32) {
		t.Fatalf("native pointer test did not allocate above 32 bits: owner=%p potion=%p update=%p player=%p use=%p", owner, potion, update, player, use)
	}

	deps := defaultPickupPotionNativeDeps4F37D0()
	var events []string
	deps.gameFlag = func(flag uint32) int32 {
		events = append(events, "flag")
		if flag != pickupPotionClassRestrictionFlag4F37D0 {
			t.Fatalf("first flag = %#x, want class restriction", flag)
		}
		return 0
	}
	deps.playerState = func(got *Object) int32 {
		events = append(events, "state")
		if got != owner {
			t.Fatalf("state owner = %p, want %p", got, owner)
		}
		return 0
	}
	deps.healthMultiplier = func(class uint8) float32 {
		events = append(events, "health-mult")
		if class != uint8(playerlib.Wizard) {
			t.Fatalf("health class = %d, want Wizard", class)
		}
		return 0.5
	}
	deps.adjustHealth = func(got *Object, amount int32) {
		events = append(events, "health")
		if got != owner || amount != 2 {
			t.Fatalf("health args = %p/%d, want owner/2", got, amount)
		}
	}
	deps.manaMultiplier = func(class uint8) float32 {
		events = append(events, "mana-mult")
		if class != uint8(playerlib.Wizard) {
			t.Fatalf("mana class = %d, want Wizard", class)
		}
		return 1.5
	}
	deps.addMana = func(got *Object, amount int32) {
		events = append(events, "mana")
		if got != owner || amount != 8 {
			t.Fatalf("mana args = %p/%d, want owner/8", got, amount)
		}
	}
	deps.audio = func(id uint32, got *Object, kind int32, code uint32) {
		events = append(events, "audio")
		if got != owner || kind != 0 || code != 0 ||
			(id != pickupPotionRestoreHealthSound4F37D0 && id != pickupPotionRestoreManaSound4F37D0 && id != 0x678) {
			t.Fatalf("audio args = %d/%p/%d/%#x", id, got, kind, code)
		}
	}
	deps.removePoison = func(got *Object) {
		events = append(events, "poison")
		if got != owner {
			t.Fatalf("poison owner = %p, want %p", got, owner)
		}
	}
	deps.spellAudio = func(spellID, field int32) uint32 {
		events = append(events, "spell")
		if spellID != 14 || field != 1 {
			t.Fatalf("spell args = %d/%d, want 14/1", spellID, field)
		}
		return 0x678
	}
	deps.delayedDelete = func(got *Object) {
		events = append(events, "delete")
		if got != potion {
			t.Fatalf("delete potion = %p, want %p", got, potion)
		}
	}
	deps.decay = func(*Object) { t.Fatal("cure path must not decay") }
	deps.defaultPickup = func(*Object, *Object, int32, int32) int32 {
		t.Fatal("cure path must not default-pickup")
		return 0
	}

	if got := pickupPotionNative4F37D0(owner, potion, math.MinInt32, math.MaxInt32, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{
		"flag", "state", "health-mult", "health", "audio",
		"mana-mult", "mana", "audio", "poison", "spell", "audio", "delete",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestPickupPotionNative4F37D0FourArgsAndExactResult(t *testing.T) {
	owner := &Object{}
	use := &PotionUseData{Value: -10}
	potion := &Object{}
	potion.UseData.SetPtr(unsafe.Pointer(use))
	for _, result := range []int32{0, 1, 2, -1, math.MinInt32, math.MaxInt32} {
		t.Run(fmt.Sprintf("%08x", uint32(result)), func(t *testing.T) {
			deps := defaultPickupPotionNativeDeps4F37D0()
			var events []string
			deps.playerState = func(*Object) int32 { return 1 }
			deps.decay = func(got *Object) {
				events = append(events, "decay")
				if got != potion {
					t.Fatalf("decay potion = %p, want %p", got, potion)
				}
			}
			deps.defaultPickup = func(gotOwner, gotPotion *Object, arg3, arg4 int32) int32 {
				events = append(events, "default")
				if gotOwner != owner || gotPotion != potion || arg3 != math.MinInt32 || arg4 != math.MaxInt32 {
					t.Fatalf("default args = %p/%p/%d/%d", gotOwner, gotPotion, arg3, arg4)
				}
				return result
			}
			deps.audio = func(id uint32, got *Object, kind int32, code uint32) {
				events = append(events, "audio")
				if id != pickupPotionInventorySound4F37D0 || got != owner || kind != 0 || code != 0 {
					t.Fatalf("pickup audio = %d/%p/%d/%#x", id, got, kind, code)
				}
			}
			if got := pickupPotionNative4F37D0(owner, potion, math.MinInt32, math.MaxInt32, deps); got != result {
				t.Fatalf("result = %d, want %d", got, result)
			}
			want := []string{"decay", "default"}
			if result == 1 {
				want = append(want, "audio")
			}
			if !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %v, want %v", events, want)
			}
		})
	}
}

func TestPickupPotionNative4F37D0PreservesNilPlayerFault(t *testing.T) {
	owner := &Object{
		ObjClass:   object.ClassPlayer,
		UpdateData: unsafe.Pointer(&PlayerUpdateData{}),
	}
	potion := &Object{}
	potion.UseData.SetPtr(unsafe.Pointer(&PotionUseData{Value: 1}))
	deps := defaultPickupPotionNativeDeps4F37D0()
	deps.gameFlag = func(flag uint32) int32 {
		if flag == pickupPotionClassRestrictionFlag4F37D0 {
			return 1
		}
		return 0
	}
	called := false
	deps.playerClassCanUse = func(*Object, uint8) int32 {
		called = true
		return 1
	}
	defer func() {
		if recover() == nil {
			t.Fatal("expected nil Player dereference")
		}
		if called {
			t.Fatal("class-use callback ran after nil Player")
		}
	}()
	pickupPotionNative4F37D0(owner, potion, 1, 2, deps)
}

func TestPickupPotionServerDeps4F37D0ClassMultiplierMapping(t *testing.T) {
	s := new(Server)
	s.Players.Mult.Warrior = ClassStats{Health: 1.25, Mana: 1.5}
	s.Players.Mult.Wizard = ClassStats{Health: 2.25, Mana: 2.5}
	s.Players.Mult.Conjurer = ClassStats{Health: 3.25, Mana: 3.5}
	deps := pickupPotionServerDeps4F37D0(s, PickupPotionRuntime4F37D0{})
	for _, tc := range []struct {
		class        uint8
		health, mana float32
	}{
		{class: uint8(playerlib.Warrior), health: 1.25, mana: 1.5},
		{class: uint8(playerlib.Wizard), health: 2.25, mana: 2.5},
		{class: uint8(playerlib.Conjurer), health: 3.25, mana: 3.5},
	} {
		if got := deps.healthMultiplier(tc.class); got != tc.health {
			t.Errorf("class %d health = %v, want %v", tc.class, got, tc.health)
		}
		if got := deps.manaMultiplier(tc.class); got != tc.mana {
			t.Errorf("class %d mana = %v, want %v", tc.class, got, tc.mana)
		}
	}
}
