package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

func defaultAudEventDropNativeDeps4EE2F0() audEventDropNativeDeps4EE2F0 {
	rows := [audEventDropRowStorage4EE2F0]audEventDropSoundRow4EE2F0{}
	for row := range rows {
		rows[row].typeInd = audEventDropSentinel4EE2F0
	}
	return audEventDropNativeDeps4EE2F0{
		defaultDrop: func(*Object, *Object, *types.Pointf) int32 { return 0 },
		loadRowType: func(row int) uint16 {
			return rows[row].typeInd
		},
		loadRowSound: func(row int) uint16 {
			return rows[row].sound
		},
		audio: func(uint32, *Object, int32, uint32) {},
	}
}

func TestAudEventDrop4EE2F0NativeLayout(t *testing.T) {
	wantTypeInd := uintptr(4)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantTypeInd = 8
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.TypeInd offset", unsafe.Offsetof(Object{}.TypeInd), wantTypeInd},
		{"Object.TypeInd size", unsafe.Sizeof(Object{}.TypeInd), 2},
		{"Pointf size", unsafe.Sizeof(types.Pointf{}), 8},
		{"Pointf.X offset", unsafe.Offsetof(types.Pointf{}.X), 0},
		{"Pointf.Y offset", unsafe.Offsetof(types.Pointf{}.Y), 4},
		{"sound row size", unsafe.Sizeof(audEventDropSoundRow4EE2F0{}), 4},
		{"sound row type offset", unsafe.Offsetof(audEventDropSoundRow4EE2F0{}.typeInd), 0},
		{"sound row sound offset", unsafe.Offsetof(audEventDropSoundRow4EE2F0{}.sound), 2},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestAudEventDropSoundTable536AC0ProductionParserKeepsFirstDuplicate(t *testing.T) {
	var typesState serverObjTypes
	first := &ObjectType{s: &typesState, ind: 44}
	second := &ObjectType{s: &typesState, ind: 44}
	parse := dropParseFuncs["AudEventDrop"]
	if err := parse(first, []string{sound.SoundAppleDrop.String()}); err != nil {
		t.Fatal(err)
	}
	if err := parse(second, []string{sound.SoundPotionDrop.String()}); err != nil {
		t.Fatal(err)
	}
	if typesState.dropSoundTable.initialized != 1 {
		t.Fatalf("init = %d, want 1", typesState.dropSoundTable.initialized)
	}
	if got := typesState.dropSoundTable.rows[0]; got != (audEventDropSoundRow4EE2F0{typeInd: 44, sound: uint16(sound.SoundAppleDrop)}) {
		t.Fatalf("first row = %+v", got)
	}
	if got := typesState.dropSoundTable.rows[1]; got != (audEventDropSoundRow4EE2F0{typeInd: 44, sound: uint16(sound.SoundPotionDrop)}) {
		t.Fatalf("second row = %+v", got)
	}
	if got, ok := typesState.dropSoundTable.first(44); !ok || got != uint16(sound.SoundAppleDrop) {
		t.Fatalf("first lookup = %d/%v", got, ok)
	}
}

func TestAudEventDropSoundTable536AC0ProductionParserCapsAtFifty(t *testing.T) {
	var typesState serverObjTypes
	parse := dropParseFuncs["AudEventDrop"]
	for i := 0; i < audEventDropRowCapacity4EE2F0+1; i++ {
		objType := &ObjectType{s: &typesState, ind: uint16(i + 1)}
		if err := parse(objType, []string{sound.SoundAppleDrop.String()}); err != nil {
			t.Fatal(err)
		}
	}
	for row := 0; row < audEventDropRowCapacity4EE2F0; row++ {
		if got := typesState.dropSoundTable.rows[row].typeInd; got != uint16(row+1) {
			t.Fatalf("row %d type = %d, want %d", row, got, row+1)
		}
	}
	if got := typesState.dropSoundTable.rows[audEventDropRowCapacity4EE2F0].typeInd; got != audEventDropSentinel4EE2F0 {
		t.Fatalf("sentinel = %04x, want ffff", got)
	}
}

func TestAudEventDropSoundTable536AC0EmptyAndUnknownTokensStillInitialize(t *testing.T) {
	tests := [][]string{nil, {}, {""}, {"not-a-real-sound"}}
	for i, args := range tests {
		t.Run(string(rune('a'+i)), func(t *testing.T) {
			var typesState serverObjTypes
			objType := &ObjectType{s: &typesState, ind: 9}
			if err := dropParseFuncs["AudEventDrop"](objType, args); err != nil {
				t.Fatal(err)
			}
			if typesState.dropSoundTable.initialized != 1 {
				t.Fatalf("init = %d", typesState.dropSoundTable.initialized)
			}
			if got := typesState.dropSoundTable.rows[0]; got != (audEventDropSoundRow4EE2F0{typeInd: audEventDropSentinel4EE2F0}) {
				t.Fatalf("row 0 = %+v", got)
			}
		})
	}
}

func TestAudEventDropNative4EE2F0BindsPointersLiveTypeAndExactResult(t *testing.T) {
	owner := &Object{}
	item := &Object{TypeInd: 3}
	point := &types.Pointf{X: 1.25, Y: -8.5}
	rows := []audEventDropSoundRow4EE2F0{
		{typeInd: 7, sound: uint16(sound.SoundAppleDrop)},
		{typeInd: audEventDropSentinel4EE2F0},
	}
	events := make([]string, 0, 2)
	deps := defaultAudEventDropNativeDeps4EE2F0()
	deps.defaultDrop = func(gotOwner, gotItem *Object, gotPoint *types.Pointf) int32 {
		events = append(events, "default")
		if gotOwner != owner || gotItem != item || gotPoint != point {
			t.Fatalf("default args = %p/%p/%p", gotOwner, gotItem, gotPoint)
		}
		item.TypeInd = 7
		return math.MinInt32
	}
	deps.loadRowType = func(row int) uint16 { return rows[row].typeInd }
	deps.loadRowSound = func(row int) uint16 { return rows[row].sound }
	deps.audio = func(id uint32, gotOwner *Object, kind int32, code uint32) {
		events = append(events, "audio")
		if id != uint32(sound.SoundAppleDrop) || gotOwner != owner || kind != 0 || code != 0 {
			t.Fatalf("audio args = %d/%p/%d/%08x", id, gotOwner, kind, code)
		}
	}
	if got := audEventDropNative4EE2F0(owner, item, point, deps); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	if !reflect.DeepEqual(events, []string{"default", "audio"}) {
		t.Fatalf("events = %v", events)
	}
}

func TestAudEventDropNative4EE2F0NilGuardsSkipDependencies(t *testing.T) {
	tests := []struct {
		name        string
		owner, item *Object
		point       *types.Pointf
	}{
		{name: "owner", item: &Object{}, point: &types.Pointf{}},
		{name: "item", owner: &Object{}, point: &types.Pointf{}},
		{name: "point", owner: &Object{}, item: &Object{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			deps := defaultAudEventDropNativeDeps4EE2F0()
			deps.defaultDrop = func(*Object, *Object, *types.Pointf) int32 {
				t.Fatal("nil gate called DefaultDrop")
				return 0
			}
			deps.loadRowType = func(int) uint16 { t.Fatal("nil gate read table"); return 0 }
			deps.loadRowSound = func(int) uint16 { t.Fatal("nil gate read sound"); return 0 }
			deps.audio = func(uint32, *Object, int32, uint32) { t.Fatal("nil gate played audio") }
			if got := audEventDropNative4EE2F0(tc.owner, tc.item, tc.point, deps); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
		})
	}
}

func TestAudEventDrop4EE2F0ServerBindingUsesOrderedTableAndAudio(t *testing.T) {
	s := &Server{}
	s.Types.dropSoundTable.initialized = 1
	for row := range s.Types.dropSoundTable.rows {
		s.Types.dropSoundTable.rows[row].typeInd = audEventDropSentinel4EE2F0
	}
	s.Types.dropSoundTable.rows[0] = audEventDropSoundRow4EE2F0{typeInd: 18, sound: uint16(sound.SoundPotionDrop)}
	owner := &Object{}
	item := &Object{TypeInd: 18}
	point := &types.Pointf{X: 7, Y: 11}
	runtime := AudEventDropRuntime4EE2F0{
		DefaultDrop: func(gotOwner, gotItem *Object, gotPoint *types.Pointf) int32 {
			if gotOwner != owner || gotItem != item || gotPoint != point {
				t.Fatalf("default args = %p/%p/%p", gotOwner, gotItem, gotPoint)
			}
			return -91
		},
	}
	if got := s.AudEventDrop4EE2F0(owner, item, point, runtime); got != -91 {
		t.Fatalf("result = %d, want -91", got)
	}
	if len(s.Audio.delayedObj) != 1 {
		t.Fatalf("queued audio count = %d, want 1", len(s.Audio.delayedObj))
	}
	audio := s.Audio.delayedObj[0]
	if audio.ID != sound.SoundPotionDrop || audio.Obj != owner || audio.Kind != 0 || audio.Code != 0 {
		t.Fatalf("queued audio = %#v", audio)
	}
}
