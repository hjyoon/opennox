package server

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
)

func defaultEquipmentDropNativeDeps53AB10() equipmentDropNativeDeps53AB10 {
	return equipmentDropNativeDeps53AB10{
		defaultDrop:   func(*Object, *Object, *types.Pointf) int32 { return 0 },
		serverSubFlag: func(uint32) int32 { return 0 },
		audio:         func(uint32, *Object, int32, uint32) {},
		gameFlag:      func(uint32) int32 { return 0 },
		loadGameFPS:   func() uint32 { return 0 },
		setDecay:      func(*Object, uint32) {},
	}
}

func TestEquipmentDropNative53AB10BindsPointersAndServices(t *testing.T) {
	owner := &Object{}
	item := &Object{ObjClass: object.ClassWand}
	point := &types.Pointf{X: 3.5, Y: -9.25}
	var events []string
	deps := defaultEquipmentDropNativeDeps53AB10()
	deps.defaultDrop = func(gotOwner, gotItem *Object, gotPoint *types.Pointf) int32 {
		events = append(events, "default")
		if gotOwner != owner || gotItem != item || gotPoint != point {
			t.Fatalf("default args = %p/%p/%p, want %p/%p/%p", gotOwner, gotItem, gotPoint, owner, item, point)
		}
		return 1
	}
	deps.audio = func(id uint32, gotItem *Object, kind int32, code uint32) {
		events = append(events, fmt.Sprintf("audio:%d", id))
		if id != uint32(sound.SoundWandDrop) || gotItem != item || kind != 0 || code != 0 {
			t.Fatalf("audio args = %d/%p/%d/%08x", id, gotItem, kind, code)
		}
	}
	deps.gameFlag = func(flag uint32) int32 {
		events = append(events, fmt.Sprintf("game:%04x", flag))
		return 0
	}
	deps.serverSubFlag = func(flag uint32) int32 {
		events = append(events, fmt.Sprintf("server:%02x", flag))
		return 1
	}
	deps.loadGameFPS = func() uint32 {
		events = append(events, "fps")
		return 17
	}
	deps.setDecay = func(gotItem *Object, delay uint32) {
		events = append(events, "decay")
		if gotItem != item || delay != 425 {
			t.Fatalf("decay args = %p/%d", gotItem, delay)
		}
	}

	if got := weaponDropNative53AB10(owner, item, point, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{"default", "audio:831", "game:0800", "game:1000", "server:02", "fps", "decay"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestEquipmentDropNative53AB10RejectsNoncanonicalDefaultWithNilPointers(t *testing.T) {
	deps := defaultEquipmentDropNativeDeps53AB10()
	deps.defaultDrop = func(owner, item *Object, point *types.Pointf) int32 {
		if owner != nil || item != nil || point != nil {
			t.Fatalf("default args = %p/%p/%p", owner, item, point)
		}
		return -1
	}
	deps.audio = func(uint32, *Object, int32, uint32) { t.Fatal("audio called") }
	deps.gameFlag = func(uint32) int32 { t.Fatal("game flag read"); return 0 }
	deps.serverSubFlag = func(uint32) int32 { t.Fatal("server flag read"); return 0 }
	deps.loadGameFPS = func() uint32 { t.Fatal("FPS read"); return 0 }
	deps.setDecay = func(*Object, uint32) { t.Fatal("decay called") }
	if got := armorDropNative53EB70(nil, nil, nil, deps); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
}

func TestEquipmentDropServerBindingsUseObjectFieldsAudioFlagsAndDecay(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	tests := []struct {
		name      string
		item      *Object
		wantSound sound.ID
		run       func(*Server, *Object, *Object, *types.Pointf, func(*Object, *Object, *types.Pointf) int32, func(uint32) int32) int32
	}{
		{
			name:      "weapon wand priority",
			item:      &Object{ObjClass: object.ClassWand, Material: uint16(object.MaterialMetal | object.MaterialWood)},
			wantSound: sound.SoundWandDrop,
			run: func(s *Server, owner, item *Object, point *types.Pointf, defaultDrop func(*Object, *Object, *types.Pointf) int32, serverFlag func(uint32) int32) int32 {
				return s.WeaponDrop53AB10(owner, item, point, WeaponDropRuntime53AB10{
					DefaultDrop: defaultDrop, ServerSubFlag: serverFlag,
				})
			},
		},
		{
			name: "armor shoes",
			item: &Object{
				Material:    uint16(object.MaterialCloth),
				ObjSubClass: object.SubClass(armorDropShoesSubclass53EAE0),
			},
			wantSound: sound.SoundShoesDrop,
			run: func(s *Server, owner, item *Object, point *types.Pointf, defaultDrop func(*Object, *Object, *types.Pointf) int32, serverFlag func(uint32) int32) int32 {
				return s.ArmorDrop53EB70(owner, item, point, ArmorDropRuntime53EB70{
					DefaultDrop: defaultDrop, ServerSubFlag: serverFlag,
				})
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := &Server{}
			s.SetFrame(100)
			s.SetTickRate(73)
			owner := &Object{}
			point := &types.Pointf{X: 7, Y: 11}
			defaultDrop := func(gotOwner, gotItem *Object, gotPoint *types.Pointf) int32 {
				if gotOwner != owner || gotItem != tc.item || gotPoint != point {
					t.Fatalf("default args = %p/%p/%p", gotOwner, gotItem, gotPoint)
				}
				return 1
			}
			serverFlag := func(flag uint32) int32 {
				if flag != equipmentDropDecayFlag53AB10 {
					t.Fatalf("server flag = %#x", flag)
				}
				return 1
			}
			if got := tc.run(s, owner, tc.item, point, defaultDrop, serverFlag); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if len(s.Audio.delayedObj) != 1 {
				t.Fatalf("queued audio count = %d, want 1", len(s.Audio.delayedObj))
			}
			audio := s.Audio.delayedObj[0]
			if audio.ID != tc.wantSound || audio.Obj != tc.item || audio.Kind != 0 || audio.Code != 0 {
				t.Fatalf("queued audio = %#v", audio)
			}
			wantDelay := uint32(73 * equipmentDropSeconds53AB10)
			if s.decay.head != tc.item || s.decay.next[tc.item] != nil {
				t.Fatalf("decay links = head %p next %p", s.decay.head, s.decay.next[tc.item])
			}
			if tc.item.Field34 != 100+wantDelay || uint32(tc.item.ObjFlags)&decayListedFlag511660 == 0 {
				t.Fatalf("decay state = deadline %d flags %08x", tc.item.Field34, uint32(tc.item.ObjFlags))
			}
		})
	}
}

func TestEquipmentDropServerDeps53AB10ObserveLiveGameFlags(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	deps := equipmentDropServerDeps53AB10(&Server{}, nil, nil)
	if got := deps.gameFlag(equipmentDropCoopFlag53AB10); got != 0 {
		t.Fatalf("clear coop = %d, want 0", got)
	}
	noxflags.SetGame(noxflags.GameModeCoop)
	if got := deps.gameFlag(equipmentDropCoopFlag53AB10); got != 1 {
		t.Fatalf("set coop = %d, want 1", got)
	}
	if got := deps.gameFlag(equipmentDropQuestFlag53AB10); got != 0 {
		t.Fatalf("clear quest = %d, want 0", got)
	}
	noxflags.SetGame(noxflags.GameModeQuest)
	if got := deps.gameFlag(equipmentDropQuestFlag53AB10); got != 1 {
		t.Fatalf("set quest = %d, want 1", got)
	}
}
