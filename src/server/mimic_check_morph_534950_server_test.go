package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestMimicCheckMorph534950NativeFieldsAndHighPointers(t *testing.T) {
	obj, freeObj := alloc.New(Object{})
	update, freeUpdate := alloc.New(MonsterUpdateData{})
	t.Cleanup(freeObj)
	t.Cleanup(freeUpdate)
	*obj = Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update), PosVec: types.Pointf{X: -0x1p-22}}
	*update = MonsterUpdateData{AIStackInd: 23, StatusFlags: object.MonsterStatus(0x40000), Field137: 0xffffffff}
	head := &update.AIStack[23]
	// The scalar argument consumes exactly its low binary32 word; upper
	// native-width bits in a uintptr slot must not alter the distance.
	upper := ^uintptr(math.MaxUint32)
	*head = AIStackItem{Action: 4, Args: [4]uintptr{upper | uintptr(math.Float32bits(8)), upper, 0xaabbccdd, 0x12345678}, Field5: 9}
	before := *update
	if unsafe.Sizeof(uintptr(0)) > 4 {
		for _, ptr := range []unsafe.Pointer{obj.CObj(), unsafe.Pointer(update), head.C()} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("native pointer=%p, want above 4 GiB", ptr)
			}
		}
	}
	var events []string
	mimicCheckMorphNative534950(obj, mimicCheckMorphNativeDeps534950{
		frame:    func() uint32 { t.Fatal("active branch read frame"); return 0 },
		tickRate: func() uint32 { t.Fatal("active branch read tick rate"); return 0 },
		pushAction: func(got *Object, action ai.ActionType) *AIStackItem {
			events = append(events, fmt.Sprintf("push:%d", action))
			if got != obj {
				t.Fatal("native push identity changed")
			}
			return nil
		},
		audio: func(id sound.ID, got *Object, kind int, code uint32) {
			events = append(events, fmt.Sprintf("audio:%d:%d:%d", id, kind, code))
			if got != obj {
				t.Fatal("native audio identity changed")
			}
		},
	})
	if want := []string{"push:61", "push:34", "audio:460:0:0"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("native events=%v want=%v", events, want)
	}
	if *update != before || obj.UpdateData != unsafe.Pointer(update) {
		t.Fatal("native adapter wrote state or resized an argument slot")
	}
}

func TestMimicCheckMorph534950NativeMalformedStateFaultsBeforeDependencies(t *testing.T) {
	for _, test := range []struct {
		name string
		obj  *Object
	}{
		{"nil object", nil},
		{"nil update", &Object{}},
		{"signed minimum index", &Object{UpdateData: unsafe.Pointer(&MonsterUpdateData{AIStackInd: -128})}},
		{"negative index", &Object{UpdateData: unsafe.Pointer(&MonsterUpdateData{AIStackInd: -1})}},
		{"first index past stack", &Object{UpdateData: unsafe.Pointer(&MonsterUpdateData{AIStackInd: 24})}},
		{"signed maximum index", &Object{UpdateData: unsafe.Pointer(&MonsterUpdateData{AIStackInd: 127})}},
	} {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("malformed native state was normalized or ignored")
				}
			}()
			mimicCheckMorphNative534950(test.obj, mimicCheckMorphNativeDeps534950{
				frame:      func() uint32 { t.Fatal("frame called after invalid state"); return 0 },
				tickRate:   func() uint32 { t.Fatal("tick rate called after invalid state"); return 0 },
				pushAction: func(*Object, ai.ActionType) *AIStackItem { t.Fatal("push called after invalid state"); return nil },
				audio:      func(sound.ID, *Object, int, uint32) { t.Fatal("audio called after invalid state") },
			})
		})
	}
}

func TestMimicCheckMorph534950ProductionStackResetAndAudio(t *testing.T) {
	srv := New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	oldGame, oldEngine := noxflags.GetGame(), noxflags.GetEngine()
	noxflags.ResetGame()
	noxflags.ResetEngine()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldGame)
		noxflags.ResetEngine()
		noxflags.SetEngine(oldEngine)
	})
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	srv.Audio.bySound[sound.SoundMimicMorph].Field12 = 1
	srv.Audio.Reset()
	obj, freeObj := alloc.New(Object{})
	update, freeUpdate := alloc.New(MonsterUpdateData{})
	t.Cleanup(freeObj)
	t.Cleanup(freeUpdate)
	*obj = Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(update), serverHandle: srv.handle, PosVec: types.Pointf{X: 12, Y: -7}}
	srv.SetFrame(131)
	srv.SetTickRate(30)
	var sounds int
	wantTimer, wantChanged := uint32(131), true
	srv.Audio.OnSound(func(id sound.ID, kind int, got *Object, pos types.Pointf) {
		sounds++
		if id != 460 || kind != 0 || got != obj || pos != obj.PosVec {
			t.Fatalf("production sound=%d/%d/%p/%v", id, kind, got, pos)
		}
		if srv.AI.StackChanged != wantChanged || update.Field137 != wantTimer {
			t.Fatalf("audio sees timer/changed=%d/%v want=%d/%v", update.Field137, srv.AI.StackChanged, wantTimer, wantChanged)
		}
	})
	for _, test := range []struct {
		name       string
		action     ai.ActionType
		status     uint32
		index      int8
		wantIndex  int8
		wantAction ai.ActionType
	}{
		{"idle replacement", ai.ACTION_IDLE, 0, 0, 1, ai.ACTION_MORPH_INTO_CHEST},
		{"active append", ai.ACTION_FIGHT, 0x40000, 0, 2, ai.ACTION_MORPH_BACK_TO_SELF},
		{"one remaining slot still sounds", ai.ACTION_FIGHT, 0x40000, 22, 23, ai.DEPENDENCY_UNINTERRUPTABLE},
		{"full stack still sounds", ai.ACTION_FIGHT, 0x40000, 23, 23, ai.ACTION_FIGHT},
	} {
		t.Run(test.name, func(t *testing.T) {
			*update = MonsterUpdateData{AIStackInd: test.index, StatusFlags: object.MonsterStatus(test.status), Field137: 100}
			for i := 0; i <= int(test.index); i++ {
				update.AIStack[i].Action = uint32(test.action)
			}
			srv.AI.StackChanged = false
			beforeSounds := sounds
			// Full stacks do not reset the timer or mark StackChanged; both push
			// attempts still occur and the audio dependency still runs.
			wantTimer, wantChanged = 131, true
			if test.index == 23 {
				wantTimer, wantChanged = 100, false
			}
			srv.MimicCheckMorph534950(obj)
			if sounds != beforeSounds+1 || update.AIStackInd != test.wantIndex || update.AIStackHead().Type() != test.wantAction {
				t.Fatalf("production sounds/index/head=%d/%d/%v", sounds-beforeSounds, update.AIStackInd, update.AIStackHead().Type())
			}
			var events int
			srv.Audio.EachEvent(func(event *AudioEvent) {
				events++
				if event.Sound != 460 || event.Obj != obj || event.Kind != 0 || event.Code != 0 || event.Pos != obj.PosVec {
					t.Fatalf("queued event=%+v", event)
				}
			})
			if events != sounds {
				t.Fatalf("queued event count=%d want=%d", events, sounds)
			}
		})
	}
}
