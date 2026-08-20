package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
)

func defaultFoodDropNativeDeps4EDE50() foodDropNativeDeps4EDE50 {
	return foodDropNativeDeps4EDE50{
		defaultDrop: func(*Object, *Object, *types.Pointf) int32 { return 0 },
		gameFlag:    func(uint32) int32 { return 0 },
		loadGameFPS: func() uint32 { return 0 },
		setDecay:    func(*Object, uint32) {},
		audio:       func(uint32, *Object, int32, uint32) {},
	}
}

func TestFoodDrop4EDE50NativeLayout(t *testing.T) {
	wantSubClass := uintptr(12)
	wantFlags := uintptr(16)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantSubClass = 16
		wantFlags = 20
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.ObjSubClass offset", unsafe.Offsetof(Object{}.ObjSubClass), wantSubClass},
		{"Object.ObjFlags offset", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.ObjSubClass size", unsafe.Sizeof(Object{}.ObjSubClass), 4},
		{"Object.ObjFlags size", unsafe.Sizeof(Object{}.ObjFlags), 4},
		{"Pointf size", unsafe.Sizeof(types.Pointf{}), 8},
		{"Pointf.X offset", unsafe.Offsetof(types.Pointf{}.X), 0},
		{"Pointf.Y offset", unsafe.Offsetof(types.Pointf{}.Y), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestFoodDrop4EDE50NativeSoundTableMatchesOracleAndEnums(t *testing.T) {
	want := [...]foodDropSoundRule4EDE50{
		{subClassMask: 0, flagsLowMask: uint16(object.FlagBelow), sound: uint16(sound.SoundMeatDrop)},
		{subClassMask: uint32(object.FoodApple), sound: uint16(sound.SoundAppleDrop)},
		{subClassMask: uint32(object.FoodJug), sound: uint16(sound.SoundPotionDrop)},
		{subClassMask: uint32(object.FoodMushroom), sound: uint16(sound.SoundShroomDrop)},
		{},
	}
	if !reflect.DeepEqual(foodDropSoundRules4EDE50, want) {
		t.Fatalf("sound rules = %#v, want %#v", foodDropSoundRules4EDE50, want)
	}
}

func TestFoodDropNative4EDE50BindsPointersLiveFieldsAndServices(t *testing.T) {
	owner := &Object{}
	food := &Object{}
	point := &types.Pointf{X: 3.5, Y: -9.25}
	events := make([]string, 0, 6)
	deps := defaultFoodDropNativeDeps4EDE50()
	deps.defaultDrop = func(gotOwner, gotFood *Object, gotPoint *types.Pointf) int32 {
		events = append(events, "default")
		if gotOwner != owner || gotFood != food || gotPoint != point {
			t.Fatalf("default args = %p/%p/%p", gotOwner, gotFood, gotPoint)
		}
		return math.MinInt32
	}
	deps.gameFlag = func(flag uint32) int32 {
		events = append(events, fmt.Sprintf("flag:%04x", flag))
		return 0
	}
	deps.loadGameFPS = func() uint32 {
		events = append(events, "fps")
		return math.MaxUint32
	}
	deps.setDecay = func(gotFood *Object, delay uint32) {
		events = append(events, "decay")
		if gotFood != food || delay != 0xffffffe7 {
			t.Fatalf("decay args = %p/%08x", gotFood, delay)
		}
		food.ObjSubClass = object.SubClass(object.FoodMushroom)
	}
	deps.audio = func(id uint32, gotOwner *Object, kind int32, code uint32) {
		events = append(events, "audio")
		if id != uint32(sound.SoundShroomDrop) || gotOwner != owner || kind != 0 || code != 0 {
			t.Fatalf("audio args = %d/%p/%d/%08x", id, gotOwner, kind, code)
		}
	}

	if got := foodDropNative4EDE50(owner, food, point, deps); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	want := []string{"default", "flag:0800", "fps", "decay", "audio"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestFoodDropNative4EDE50ReadsLowObjectFlags(t *testing.T) {
	owner := &Object{}
	food := &Object{ObjFlags: object.FlagBelow | object.FlagMarked}
	point := &types.Pointf{}
	deps := defaultFoodDropNativeDeps4EDE50()
	deps.defaultDrop = func(*Object, *Object, *types.Pointf) int32 { return -1 }
	deps.gameFlag = func(uint32) int32 { return 1 }
	deps.audio = func(id uint32, gotOwner *Object, kind int32, code uint32) {
		if id != uint32(sound.SoundMeatDrop) || gotOwner != owner || kind != 0 || code != 0 {
			t.Fatalf("audio args = %d/%p/%d/%08x", id, gotOwner, kind, code)
		}
	}
	if got := foodDropNative4EDE50(owner, food, point, deps); got != -1 {
		t.Fatalf("result = %d, want -1", got)
	}
}

func TestFoodDropNative4EDE50NilGuardsSkipDependencies(t *testing.T) {
	tests := []struct {
		name        string
		owner, food *Object
		point       *types.Pointf
	}{
		{name: "owner", food: &Object{}, point: &types.Pointf{}},
		{name: "food", owner: &Object{}, point: &types.Pointf{}},
		{name: "point", owner: &Object{}, food: &Object{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			deps := defaultFoodDropNativeDeps4EDE50()
			deps.defaultDrop = func(*Object, *Object, *types.Pointf) int32 {
				t.Fatal("nil guard called DefaultDrop")
				return 0
			}
			deps.gameFlag = func(uint32) int32 { t.Fatal("nil guard read game flag"); return 0 }
			deps.loadGameFPS = func() uint32 { t.Fatal("nil guard read FPS"); return 0 }
			deps.setDecay = func(*Object, uint32) { t.Fatal("nil guard set decay") }
			deps.audio = func(uint32, *Object, int32, uint32) { t.Fatal("nil guard played audio") }
			if got := foodDropNative4EDE50(tc.owner, tc.food, tc.point, deps); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
		})
	}
}

func TestFoodDrop4EDE50ServerBindingUsesFlagsTickRateAudioAndNativeDecay(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	s := &Server{}
	s.SetFrame(100)
	s.SetTickRate(73)
	owner := &Object{}
	food := &Object{ObjSubClass: object.SubClass(object.FoodApple)}
	point := &types.Pointf{X: 7, Y: 11}
	runtime := FoodDropRuntime4EDE50{
		DefaultDrop: func(gotOwner, gotFood *Object, gotPoint *types.Pointf) int32 {
			if gotOwner != owner || gotFood != food || gotPoint != point {
				t.Fatalf("default args = %p/%p/%p", gotOwner, gotFood, gotPoint)
			}
			return math.MinInt32
		},
	}
	if got := s.FoodDrop4EDE50(owner, food, point, runtime); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	if len(s.Audio.delayedObj) != 1 {
		t.Fatalf("queued audio count = %d, want 1", len(s.Audio.delayedObj))
	}
	audio := s.Audio.delayedObj[0]
	if audio.ID != sound.SoundAppleDrop || audio.Obj != owner || audio.Kind != 0 || audio.Code != 0 {
		t.Fatalf("queued audio = %#v", audio)
	}
	wantDelay := uint32(73 * 25)
	if s.decay.head != food || s.decay.next[food] != nil {
		t.Fatalf("decay links = head %p next %p", s.decay.head, s.decay.next[food])
	}
	if food.Field34 != 100+wantDelay || uint32(food.ObjFlags)&decayListedFlag511660 == 0 {
		t.Fatalf("decay state = deadline %d flags %08x", food.Field34, uint32(food.ObjFlags))
	}
}

func TestFoodDropServerDeps4EDE50ObserveLiveCoopFlag(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	deps := foodDropServerDeps4EDE50(&Server{}, FoodDropRuntime4EDE50{})
	if got := deps.gameFlag(foodDropCoopFlag4EDE50); got != 0 {
		t.Fatalf("clear coop = %d, want 0", got)
	}
	noxflags.SetGame(noxflags.GameModeCoop)
	if got := deps.gameFlag(foodDropCoopFlag4EDE50); got != 1 {
		t.Fatalf("set coop = %d, want 1", got)
	}
}
