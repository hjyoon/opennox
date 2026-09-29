package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type equipmentDropTestWorld53AB10 struct {
	pointArg        string
	ownerArg        string
	itemArg         string
	defaultResult   int32
	gameFlags       map[uint32]int32
	serverSubFlags  map[uint32]int32
	fps             uint32
	events          []string
	faultAt         int
	afterDefault    func(*equipmentDropTestWorld53AB10)
	afterSound      func(*equipmentDropTestWorld53AB10)
	afterGameFlag   func(*equipmentDropTestWorld53AB10, uint32)
	afterServerFlag func(*equipmentDropTestWorld53AB10, uint32)
}

func newEquipmentDropTestWorld53AB10() *equipmentDropTestWorld53AB10 {
	return &equipmentDropTestWorld53AB10{
		pointArg:       "point-a",
		ownerArg:       "owner-a",
		itemArg:        "item-a",
		defaultResult:  1,
		gameFlags:      make(map[uint32]int32),
		serverSubFlags: map[uint32]int32{equipmentDropDecayFlag53AB10: 1},
		fps:            30,
	}
}

func (w *equipmentDropTestWorld53AB10) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(value)
	}
}

func (w *equipmentDropTestWorld53AB10) hooks() equipmentDropHooks53AB10[string, string] {
	return equipmentDropHooks53AB10[string, string]{
		loadPointArg: func() string {
			w.event("point-arg:" + w.pointArg)
			return w.pointArg
		},
		loadOwnerArg: func() string {
			w.event("owner-arg:" + w.ownerArg)
			return w.ownerArg
		},
		loadItemArg: func() string {
			w.event("item-arg:" + w.itemArg)
			return w.itemArg
		},
		defaultDrop: func(owner, item, point string) int32 {
			w.event("default:" + owner + ":" + item + ":" + point)
			result := w.defaultResult
			if w.afterDefault != nil {
				w.afterDefault(w)
			}
			return result
		},
		dropSound: func(item string) {
			w.event("sound:" + item)
			if w.afterSound != nil {
				w.afterSound(w)
			}
		},
		gameFlag: func(flag uint32) int32 {
			result := w.gameFlags[flag]
			w.event(fmt.Sprintf("game-flag:%08x=%08x", flag, uint32(result)))
			if w.afterGameFlag != nil {
				w.afterGameFlag(w, flag)
			}
			return result
		},
		serverSubFlag: func(flag uint32) int32 {
			result := w.serverSubFlags[flag]
			w.event(fmt.Sprintf("server-flag:%08x=%08x", flag, uint32(result)))
			if w.afterServerFlag != nil {
				w.afterServerFlag(w, flag)
			}
			return result
		},
		loadGameFPS: func() uint32 {
			w.event(fmt.Sprintf("fps:%08x", w.fps))
			return w.fps
		},
		setDecay: func(item string, delay uint32) {
			w.event(fmt.Sprintf("decay:%s:%08x", item, delay))
		},
	}
}

func equipmentDropSuccessEvents53AB10(fps uint32) []string {
	return []string{
		"point-arg:point-a",
		"owner-arg:owner-a",
		"item-arg:item-a",
		"default:owner-a:item-a:point-a",
		"sound:item-a",
		"game-flag:00000800=00000000",
		"game-flag:00001000=00000000",
		"server-flag:00000002=00000001",
		fmt.Sprintf("fps:%08x", fps),
		fmt.Sprintf("decay:item-a:%08x", fps*equipmentDropSeconds53AB10),
	}
}

func verifyEquipmentDropFaultPrefixes53AB10(
	t *testing.T,
	want []string,
	build func() *equipmentDropTestWorld53AB10,
) {
	t.Helper()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := build()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %v, want %v", w.events, want[:faultAt])
				}
			}()
			weaponDrop53AB10(w.hooks())
		})
	}
}

func TestEquipmentDrop53AB10ExactSuccessTraceAndUint32Wrap(t *testing.T) {
	build := func() *equipmentDropTestWorld53AB10 {
		w := newEquipmentDropTestWorld53AB10()
		w.fps = math.MaxUint32
		return w
	}
	want := equipmentDropSuccessEvents53AB10(math.MaxUint32)
	w := build()
	if got := weaponDrop53AB10(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	if want[len(want)-1] != "decay:item-a:ffffffe7" {
		t.Fatalf("wrapped decay = %q", want[len(want)-1])
	}
	verifyEquipmentDropFaultPrefixes53AB10(t, want, build)
}

func TestEquipmentDrop53AB10DefaultMustEqualExactlyOne(t *testing.T) {
	for _, value := range []int32{0, 1, 2, -1, math.MinInt32} {
		t.Run(fmt.Sprintf("%08x", uint32(value)), func(t *testing.T) {
			w := newEquipmentDropTestWorld53AB10()
			w.defaultResult = value
			got := armorDrop53EB70(w.hooks())
			if value == 1 {
				if got != 1 || !reflect.DeepEqual(w.events, equipmentDropSuccessEvents53AB10(30)) {
					t.Fatalf("result/events = %d/%v", got, w.events)
				}
				return
			}
			want := []string{
				"point-arg:point-a", "owner-arg:owner-a", "item-arg:item-a",
				"default:owner-a:item-a:point-a",
			}
			if got != 0 || !reflect.DeepEqual(w.events, want) {
				t.Fatalf("result/events = %d/%v, want 0/%v", got, w.events, want)
			}
		})
	}
}

func TestEquipmentDrop53AB10ModeAndDecayGatesShortCircuit(t *testing.T) {
	tests := []struct {
		name       string
		configure  func(*equipmentDropTestWorld53AB10)
		wantSuffix []string
	}{
		{
			name: "coop",
			configure: func(w *equipmentDropTestWorld53AB10) {
				w.gameFlags[equipmentDropCoopFlag53AB10] = -1
			},
			wantSuffix: []string{"game-flag:00000800=ffffffff"},
		},
		{
			name: "quest",
			configure: func(w *equipmentDropTestWorld53AB10) {
				w.gameFlags[equipmentDropQuestFlag53AB10] = math.MinInt32
			},
			wantSuffix: []string{
				"game-flag:00000800=00000000",
				"game-flag:00001000=80000000",
			},
		},
		{
			name: "server decay disabled",
			configure: func(w *equipmentDropTestWorld53AB10) {
				w.serverSubFlags[equipmentDropDecayFlag53AB10] = 0
			},
			wantSuffix: []string{
				"game-flag:00000800=00000000",
				"game-flag:00001000=00000000",
				"server-flag:00000002=00000000",
			},
		},
	}
	prefix := []string{
		"point-arg:point-a", "owner-arg:owner-a", "item-arg:item-a",
		"default:owner-a:item-a:point-a", "sound:item-a",
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			build := func() *equipmentDropTestWorld53AB10 {
				w := newEquipmentDropTestWorld53AB10()
				tc.configure(w)
				return w
			}
			w := build()
			if got := equipmentDrop53AB10(w.hooks()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			want := append(append([]string{}, prefix...), tc.wantSuffix...)
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %v, want %v", w.events, want)
			}
			verifyEquipmentDropFaultPrefixes53AB10(t, want, build)
		})
	}
}

func TestEquipmentDrop53AB10CachesArgumentsAndReadsPostSoundState(t *testing.T) {
	w := newEquipmentDropTestWorld53AB10()
	w.gameFlags[equipmentDropCoopFlag53AB10] = 1
	w.gameFlags[equipmentDropQuestFlag53AB10] = 1
	w.serverSubFlags[equipmentDropDecayFlag53AB10] = 0
	w.afterDefault = func(w *equipmentDropTestWorld53AB10) {
		w.pointArg = "point-b"
		w.ownerArg = "owner-b"
		w.itemArg = "item-b"
	}
	w.afterSound = func(w *equipmentDropTestWorld53AB10) {
		w.gameFlags[equipmentDropCoopFlag53AB10] = 0
	}
	w.afterGameFlag = func(w *equipmentDropTestWorld53AB10, flag uint32) {
		switch flag {
		case equipmentDropCoopFlag53AB10:
			w.gameFlags[equipmentDropQuestFlag53AB10] = 0
		case equipmentDropQuestFlag53AB10:
			w.serverSubFlags[equipmentDropDecayFlag53AB10] = -1
		}
	}
	w.afterServerFlag = func(w *equipmentDropTestWorld53AB10, flag uint32) {
		w.fps = 0x80000001
	}
	if got := equipmentDrop53AB10(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := equipmentDropSuccessEvents53AB10(0x80000001)
	want[7] = "server-flag:00000002=ffffffff"
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestWeaponDropSound53AAB0PriorityAndReadOrder(t *testing.T) {
	tests := []struct {
		name     string
		item     int
		class    uint32
		material uint16
		want     []string
	}{
		{name: "nil", want: []string{"item:0"}},
		{
			name: "wand before material", item: 1, class: weaponDropWandClass53AAB0,
			material: weaponDropMetalMaterial53AAB0 | weaponDropWoodMaterial53AAB0,
			want:     []string{"item:1", "class:1=00001000", "audio:831:1:0:00000000"},
		},
		{
			name: "metal before wood", item: 1,
			material: weaponDropMetalMaterial53AAB0 | weaponDropWoodMaterial53AAB0,
			want: []string{
				"item:1", "class:1=00000000", "material:1=0018", "audio:843:1:0:00000000",
			},
		},
		{
			name: "wood", item: 1, material: weaponDropWoodMaterial53AAB0,
			want: []string{
				"item:1", "class:1=00000000", "material:1=0008", "audio:845:1:0:00000000",
			},
		},
		{
			name: "silent material", item: 1, material: 1,
			want: []string{"item:1", "class:1=00000000", "material:1=0001"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			weaponDropSound53AAB0(weaponDropSoundHooks53AAB0[int]{
				loadItemArg: func() int {
					events = append(events, fmt.Sprintf("item:%d", tc.item))
					return tc.item
				},
				loadClass: func(item int) uint32 {
					events = append(events, fmt.Sprintf("class:%d=%08x", item, tc.class))
					return tc.class
				},
				loadMaterial: func(item int) uint16 {
					events = append(events, fmt.Sprintf("material:%d=%04x", item, tc.material))
					return tc.material
				},
				audio: func(id uint32, item int, kind int32, code uint32) {
					events = append(events, fmt.Sprintf("audio:%d:%d:%d:%08x", id, item, kind, code))
				},
			})
			if !reflect.DeepEqual(events, tc.want) {
				t.Fatalf("events = %v, want %v", events, tc.want)
			}
		})
	}
}

func TestArmorDropSound53EAE0PriorityAndShoesBranch(t *testing.T) {
	tests := []struct {
		name     string
		item     int
		material uint16
		subclass uint32
		want     []string
	}{
		{name: "nil", want: []string{"item:0"}},
		{
			name: "metal before all", item: 1,
			material: armorDropMetalMaterial53EAE0 | armorDropWoodMaterial53EAE0 | armorDropLeatherMaterial53EAE0 | armorDropClothMaterial53EAE0,
			want:     []string{"item:1", "material:1=001e", "audio:805:1:0:00000000"},
		},
		{
			name: "wood before leather", item: 1,
			material: armorDropWoodMaterial53EAE0 | armorDropLeatherMaterial53EAE0,
			want:     []string{"item:1", "material:1=000c", "audio:811:1:0:00000000"},
		},
		{
			name: "leather", item: 1, material: armorDropLeatherMaterial53EAE0,
			want: []string{"item:1", "material:1=0004", "audio:808:1:0:00000000"},
		},
		{
			name: "cloth", item: 1, material: armorDropClothMaterial53EAE0,
			want: []string{
				"item:1", "material:1=0002", "subclass:1=00000000", "audio:814:1:0:00000000",
			},
		},
		{
			name: "shoes", item: 1, material: armorDropClothMaterial53EAE0, subclass: armorDropShoesSubclass53EAE0,
			want: []string{
				"item:1", "material:1=0002", "subclass:1=00000020", "audio:817:1:0:00000000",
			},
		},
		{
			name: "silent material", item: 1, material: 1,
			want: []string{"item:1", "material:1=0001"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			armorDropSound53EAE0(armorDropSoundHooks53EAE0[int]{
				loadItemArg: func() int {
					events = append(events, fmt.Sprintf("item:%d", tc.item))
					return tc.item
				},
				loadMaterial: func(item int) uint16 {
					events = append(events, fmt.Sprintf("material:%d=%04x", item, tc.material))
					return tc.material
				},
				loadSubClass: func(item int) uint32 {
					events = append(events, fmt.Sprintf("subclass:%d=%08x", item, tc.subclass))
					return tc.subclass
				},
				audio: func(id uint32, item int, kind int32, code uint32) {
					events = append(events, fmt.Sprintf("audio:%d:%d:%d:%08x", id, item, kind, code))
				},
			})
			if !reflect.DeepEqual(events, tc.want) {
				t.Fatalf("events = %v, want %v", events, tc.want)
			}
		})
	}
}
