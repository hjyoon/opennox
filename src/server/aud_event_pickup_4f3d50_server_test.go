package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/sound"
)

func defaultAudEventPickupNativeDeps4F3D50() audEventPickupNativeDeps4F3D50 {
	rows := [audEventPickupRowStorage4F3D50]audEventPickupSoundRow4F3D50{}
	for row := range rows {
		rows[row].typeInd = audEventPickupSentinel4F3D50
	}
	return audEventPickupNativeDeps4F3D50{
		defaultPickup: func(*Object, *Object, int32, int32) int32 { return 0 },
		loadRowType: func(row int) uint16 {
			return rows[row].typeInd
		},
		loadRowSound: func(row int) uint16 {
			return rows[row].sound
		},
		audio: func(uint32, *Object, int32, uint32) {},
	}
}

func TestAudEventPickup4F3D50NativeLayout(t *testing.T) {
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
		{"sound row size", unsafe.Sizeof(audEventPickupSoundRow4F3D50{}), 4},
		{"sound row type offset", unsafe.Offsetof(audEventPickupSoundRow4F3D50{}.typeInd), 0},
		{"sound row sound offset", unsafe.Offsetof(audEventPickupSoundRow4F3D50{}.sound), 2},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestAudEventPickupSoundTable5367B0ProductionParserKeepsFirstDuplicate(t *testing.T) {
	var typesState serverObjTypes
	first := &ObjectType{s: &typesState, ind: 44}
	second := &ObjectType{s: &typesState, ind: 44}
	parse := pickupParseFuncs["AudEventPickup"]
	if err := parse(first, []string{sound.SoundApplePickup.String()}); err != nil {
		t.Fatal(err)
	}
	if err := parse(second, []string{sound.SoundPotionPickup.String()}); err != nil {
		t.Fatal(err)
	}
	if typesState.pickupSoundTable.initialized != 1 {
		t.Fatalf("init = %d, want 1", typesState.pickupSoundTable.initialized)
	}
	if got := typesState.pickupSoundTable.rows[0]; got != (audEventPickupSoundRow4F3D50{typeInd: 44, sound: uint16(sound.SoundApplePickup)}) {
		t.Fatalf("first row = %+v", got)
	}
	if got := typesState.pickupSoundTable.rows[1]; got != (audEventPickupSoundRow4F3D50{typeInd: 44, sound: uint16(sound.SoundPotionPickup)}) {
		t.Fatalf("second row = %+v", got)
	}
	if got, ok := typesState.pickupSoundTable.first(44); !ok || got != uint16(sound.SoundApplePickup) {
		t.Fatalf("first lookup = %d/%v", got, ok)
	}
}

func TestAudEventPickupSoundTable5367B0ProductionParserCapsAtFifty(t *testing.T) {
	var typesState serverObjTypes
	parse := pickupParseFuncs["AudEventPickup"]
	for i := 0; i < audEventPickupRowCapacity4F3D50+1; i++ {
		objType := &ObjectType{s: &typesState, ind: uint16(i + 1)}
		if err := parse(objType, []string{sound.SoundApplePickup.String()}); err != nil {
			t.Fatal(err)
		}
	}
	for row := 0; row < audEventPickupRowCapacity4F3D50; row++ {
		if got := typesState.pickupSoundTable.rows[row].typeInd; got != uint16(row+1) {
			t.Fatalf("row %d type = %d, want %d", row, got, row+1)
		}
	}
	if got := typesState.pickupSoundTable.rows[audEventPickupRowCapacity4F3D50].typeInd; got != audEventPickupSentinel4F3D50 {
		t.Fatalf("sentinel = %04x, want ffff", got)
	}
}

func TestAudEventPickupSoundTable5367B0EmptyAndUnknownTokensStillInitialize(t *testing.T) {
	tests := [][]string{nil, {}, {""}, {"not-a-real-sound"}}
	for i, args := range tests {
		t.Run(string(rune('a'+i)), func(t *testing.T) {
			var typesState serverObjTypes
			objType := &ObjectType{s: &typesState, ind: 9}
			if err := pickupParseFuncs["AudEventPickup"](objType, args); err != nil {
				t.Fatal(err)
			}
			if typesState.pickupSoundTable.initialized != 1 {
				t.Fatalf("init = %d", typesState.pickupSoundTable.initialized)
			}
			if got := typesState.pickupSoundTable.rows[0]; got != (audEventPickupSoundRow4F3D50{typeInd: audEventPickupSentinel4F3D50}) {
				t.Fatalf("row 0 = %+v", got)
			}
		})
	}
}

func TestAudEventPickupNative4F3D50BindsPointersScalarsLiveTypeAndExactResult(t *testing.T) {
	owner := &Object{}
	item := &Object{TypeInd: 3}
	rows := []audEventPickupSoundRow4F3D50{
		{typeInd: 7, sound: uint16(sound.SoundApplePickup)},
		{typeInd: audEventPickupSentinel4F3D50},
	}
	events := make([]string, 0, 2)
	deps := defaultAudEventPickupNativeDeps4F3D50()
	deps.defaultPickup = func(gotOwner, gotItem *Object, arg3, arg4 int32) int32 {
		events = append(events, "default")
		if gotOwner != owner || gotItem != item || arg3 != math.MinInt32 || arg4 != math.MaxInt32 {
			t.Fatalf("default args = %p/%p/%d/%d", gotOwner, gotItem, arg3, arg4)
		}
		item.TypeInd = 7
		return -91
	}
	deps.loadRowType = func(row int) uint16 { return rows[row].typeInd }
	deps.loadRowSound = func(row int) uint16 { return rows[row].sound }
	deps.audio = func(id uint32, gotOwner *Object, kind int32, code uint32) {
		events = append(events, "audio")
		if id != uint32(sound.SoundApplePickup) || gotOwner != owner || kind != 0 || code != 0 {
			t.Fatalf("audio args = %d/%p/%d/%08x", id, gotOwner, kind, code)
		}
	}
	if got := audEventPickupNative4F3D50(owner, item, math.MinInt32, math.MaxInt32, deps); got != -91 {
		t.Fatalf("result = %d, want -91", got)
	}
	if !reflect.DeepEqual(events, []string{"default", "audio"}) {
		t.Fatalf("events = %v", events)
	}
}

func TestAudEventPickupNative4F3D50NilGuardsSkipDependencies(t *testing.T) {
	tests := []struct {
		name        string
		owner, item *Object
	}{
		{name: "owner", item: &Object{}},
		{name: "item", owner: &Object{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			deps := defaultAudEventPickupNativeDeps4F3D50()
			deps.defaultPickup = func(*Object, *Object, int32, int32) int32 {
				t.Fatal("nil gate called DefaultPickup")
				return 0
			}
			deps.loadRowType = func(int) uint16 { t.Fatal("nil gate read table"); return 0 }
			deps.loadRowSound = func(int) uint16 { t.Fatal("nil gate read sound"); return 0 }
			deps.audio = func(uint32, *Object, int32, uint32) { t.Fatal("nil gate played audio") }
			if got := audEventPickupNative4F3D50(tc.owner, tc.item, 1, 2, deps); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
		})
	}
}

func TestAudEventPickup4F3D50ServerBindingUsesOrderedTableAndAudio(t *testing.T) {
	s := &Server{}
	s.Types.pickupSoundTable.initialized = 1
	for row := range s.Types.pickupSoundTable.rows {
		s.Types.pickupSoundTable.rows[row].typeInd = audEventPickupSentinel4F3D50
	}
	s.Types.pickupSoundTable.rows[0] = audEventPickupSoundRow4F3D50{typeInd: 18, sound: uint16(sound.SoundPotionPickup)}
	owner := &Object{}
	item := &Object{TypeInd: 18}
	runtime := AudEventPickupRuntime4F3D50{
		DefaultPickup: func(gotOwner, gotItem *Object, arg3, arg4 int32) int32 {
			if gotOwner != owner || gotItem != item || arg3 != 7 || arg4 != 11 {
				t.Fatalf("default args = %p/%p/%d/%d", gotOwner, gotItem, arg3, arg4)
			}
			return math.MinInt32
		},
	}
	if got := s.AudEventPickup4F3D50(owner, item, 7, 11, runtime); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	if len(s.Audio.delayedObj) != 1 {
		t.Fatalf("queued audio count = %d, want 1", len(s.Audio.delayedObj))
	}
	audio := s.Audio.delayedObj[0]
	if audio.ID != sound.SoundPotionPickup || audio.Obj != owner || audio.Kind != 0 || audio.Code != 0 {
		t.Fatalf("queued audio = %#v", audio)
	}
}
