package server

import (
	"fmt"
	"image"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestMeteorUpdateDataLayout53D5A0(t *testing.T) {
	if got := unsafe.Sizeof(MeteorUpdateData{}); got != 4 {
		t.Fatalf("MeteorUpdateData size = %d, want 4", got)
	}
	if got := unsafe.Offsetof(MeteorUpdateData{}.Damage); got != 0 {
		t.Fatalf("MeteorUpdateData.Damage offset = %d, want 0", got)
	}
}

func TestMeteorShowerUpdateNative53D5A0SpawnsMeteor(t *testing.T) {
	sourceData := &MeteorUpdateData{Damage: 41}
	meteorData := new(MeteorUpdateData)
	source := &Object{
		PosVec:     types.Ptf(100, 200),
		Field32:    10,
		Field34:    20,
		UpdateData: unsafe.Pointer(sourceData),
	}
	meteor := &Object{UpdateData: unsafe.Pointer(meteorData)}
	if unsafe.Sizeof(uintptr(0)) == 8 &&
		(uintptr(unsafe.Pointer(source)) <= math.MaxUint32 || uintptr(unsafe.Pointer(meteor)) <= math.MaxUint32) {
		t.Fatalf("object pointers = %p/%p, want addresses above the ABI32 range", source, meteor)
	}

	cosine, sine := SinCosDir(64)
	wantPosition := types.Ptf(
		float32(float64(source.PosVec.X)+25*float64(cosine)),
		float32(float64(source.PosVec.Y)+25*float64(sine)),
	)
	var events []string
	randomCalls := 0
	meteorShowerUpdateNative53D5A0(source, meteorShowerUpdateDeps53D5A0{
		frame: func() uint32 { return 20 },
		fps:   func() uint32 { return 30 },
		randomInt: func(minimum, maximum int) int {
			randomCalls++
			switch randomCalls {
			case 1:
				if minimum != 0 || maximum != 255 {
					t.Fatalf("direction range = %d..%d, want 0..255", minimum, maximum)
				}
				return 64
			case 2:
				if minimum != 4 || maximum != 8 {
					t.Fatalf("delay range = %d..%d, want 4..8", minimum, maximum)
				}
				return 6
			default:
				t.Fatalf("unexpected random integer call %d", randomCalls)
				return 0
			}
		},
		randomFloat: func(minimum, maximum float64) float64 {
			if minimum != 4 || maximum != 12 {
				t.Fatalf("radius range = %v..%v, want 4..12", minimum, maximum)
			}
			return 5
		},
		traceRay: func(from, to types.Pointf) bool {
			events = append(events, "trace")
			if from != source.PosVec || to != wantPosition {
				t.Fatalf("trace = %v -> %v, want %v -> %v", from, to, source.PosVec, wantPosition)
			}
			return true
		},
		newObjectByTypeID: func(typeID string) *Object {
			events = append(events, "new")
			if typeID != "Meteor" {
				t.Fatalf("type ID = %q, want Meteor", typeID)
			}
			return meteor
		},
		createAt: func(got, owner *Object, position types.Pointf) {
			events = append(events, "create")
			if got != meteor || owner != source || position != wantPosition {
				t.Fatalf("create = %p/%p/%v, want %p/%p/%v", got, owner, position, meteor, source, wantPosition)
			}
		},
		raise: func(got *Object, height float32) {
			events = append(events, "raise")
			if got != meteor || height != 255 || got.Field5&meteorFallingXStatus53D5A0 == 0 {
				t.Fatalf("raise = %p/%v xstatus %#x", got, height, got.Field5)
			}
			got.ZVal = height
		},
		delayedDelete: func(*Object) { t.Fatal("active meteor shower was deleted") },
	})

	if got, want := fmt.Sprint(events), "[trace new create raise]"; got != want {
		t.Fatalf("events = %s, want %s", got, want)
	}
	if meteorData.Damage != sourceData.Damage || meteor.Field5&meteorFallingXStatus53D5A0 == 0 ||
		meteor.ZVal != 255 || meteor.Field27 != -8 || source.Field34 != 26 || randomCalls != 2 {
		t.Fatalf("meteor state = damage %d xstatus %#x z %v velocity %v next %d random calls %d",
			meteorData.Damage, meteor.Field5, meteor.ZVal, meteor.Field27, source.Field34, randomCalls)
	}
}

func TestMeteorShowerUpdateNative53D5A0TimingAndBlockedTrace(t *testing.T) {
	source := &Object{Field32: 10, Field34: 20}
	deleted := 0
	meteorShowerUpdateNative53D5A0(source, meteorShowerUpdateDeps53D5A0{
		frame: func() uint32 { return 160 },
		fps:   func() uint32 { return 30 },
		delayedDelete: func(got *Object) {
			if got != source {
				t.Fatalf("deleted = %p, want %p", got, source)
			}
			deleted++
		},
	})
	if deleted != 1 {
		t.Fatalf("deletes at lifetime boundary = %d, want 1", deleted)
	}

	source.Field32 = 20
	source.Field34 = 21
	meteorShowerUpdateNative53D5A0(source, meteorShowerUpdateDeps53D5A0{
		frame:         func() uint32 { return 20 },
		fps:           func() uint32 { return 30 },
		randomInt:     func(int, int) int { t.Fatal("waiting shower used random integer"); return 0 },
		randomFloat:   func(float64, float64) float64 { t.Fatal("waiting shower used random float"); return 0 },
		traceRay:      func(types.Pointf, types.Pointf) bool { t.Fatal("waiting shower traced ray"); return false },
		delayedDelete: func(*Object) { t.Fatal("waiting shower was deleted") },
	})
	if source.Field34 != 21 {
		t.Fatalf("waiting next frame = %d, want 21", source.Field34)
	}

	source.Field34 = 20
	randomCalls := 0
	meteorShowerUpdateNative53D5A0(source, meteorShowerUpdateDeps53D5A0{
		frame: func() uint32 { return 20 },
		fps:   func() uint32 { return 30 },
		randomInt: func(minimum, maximum int) int {
			randomCalls++
			if randomCalls == 1 {
				return 0
			}
			return 4
		},
		randomFloat:       func(float64, float64) float64 { return 4 },
		traceRay:          func(types.Pointf, types.Pointf) bool { return false },
		newObjectByTypeID: func(string) *Object { t.Fatal("blocked trace created meteor"); return nil },
		delayedDelete:     func(*Object) { t.Fatal("blocked trace deleted shower") },
	})
	if source.Field34 != 24 || randomCalls != 2 {
		t.Fatalf("blocked trace scheduling = frame %d, random calls %d; want 24/2", source.Field34, randomCalls)
	}
}

func TestMeteorUpdateNative53D6E0Impact(t *testing.T) {
	update := &MeteorUpdateData{Damage: 37}
	player := &Object{ObjClass: object.ClassPlayer}
	source := &Object{
		PosVec:     types.Ptf(100, 200),
		ZVal:       0,
		ObjOwner:   player,
		UpdateData: unsafe.Pointer(update),
	}
	explosion := new(Object)
	var events []string
	meteorUpdateNative53D6E0(source, meteorUpdateDeps53D6E0{
		audioEvent: func(id uint32, got *Object) {
			events = append(events, "audio")
			if id != 87 || got != source {
				t.Fatalf("audio = %d/%p, want 87/%p", id, got, source)
			}
		},
		makeScorch: func(position types.Pointf, kind int) {
			events = append(events, "scorch")
			if position != source.PosVec || kind != 2 {
				t.Fatalf("scorch = %v/%d, want %v/2", position, kind, source.PosVec)
			}
		},
		earthquake: func(position types.Pointf, magnitude int) {
			events = append(events, "quake")
			if position != source.PosVec || magnitude != 10 {
				t.Fatalf("earthquake = %v/%d, want %v/10", position, magnitude, source.PosVec)
			}
		},
		newObjectByTypeID: func(typeID string) *Object {
			events = append(events, "new")
			if typeID != "MeteorExplode" {
				t.Fatalf("type ID = %q, want MeteorExplode", typeID)
			}
			return explosion
		},
		createAt: func(got, owner *Object, position types.Pointf) {
			events = append(events, "create")
			if got != explosion || owner != nil || position != source.PosVec {
				t.Fatalf("create = %p/%p/%v, want %p/nil/%v", got, owner, position, explosion, source.PosVec)
			}
		},
		findParentChainPlayer: func(got *Object) *Object {
			events = append(events, "parent")
			if got != source {
				t.Fatalf("parent source = %p, want %p", got, source)
			}
			return player
		},
		damageUnitsAround: func(position types.Pointf, outerRadius, innerRadius float32, damage int, damageType object.DamageType, gotPlayer *Object, excluded Obj, damageSource bool) {
			events = append(events, "units")
			if position != source.PosVec || outerRadius != 80 || innerRadius != 30 || damage != 37 ||
				damageType != object.DamageExplosion || gotPlayer != player || excluded != nil || !damageSource {
				t.Fatalf("unit damage = %v/%v/%v/%d/%v/%p/%v/%t",
					position, outerRadius, innerRadius, damage, damageType, gotPlayer, excluded, damageSource)
			}
		},
		damageWalls: func(rect image.Rectangle, position types.Pointf, radius float32, damage int, damageType object.DamageType, got *Object) {
			events = append(events, "walls")
			if rect != image.Rect(0, 5, 7, 12) || position != source.PosVec || radius != 80 || damage != 37 ||
				damageType != object.DamageExplosion || got != source {
				t.Fatalf("wall damage = %v/%v/%v/%d/%v/%p", rect, position, radius, damage, damageType, got)
			}
		},
		damageSource: func() bool {
			events = append(events, "damage-source")
			return true
		},
		delayedDelete: func(got *Object) {
			events = append(events, "delete")
			if got != source {
				t.Fatalf("deleted = %p, want %p", got, source)
			}
		},
	})

	want := "[audio scorch quake new create parent damage-source units walls delete]"
	if got := fmt.Sprint(events); got != want {
		t.Fatalf("events = %s, want %s", got, want)
	}
}

func TestMeteorUpdateNative53D6E0AirborneAndNaNDoNothing(t *testing.T) {
	runtime := meteorUpdateDeps53D6E0{
		audioEvent:        func(uint32, *Object) { t.Fatal("airborne meteor played audio") },
		makeScorch:        func(types.Pointf, int) { t.Fatal("airborne meteor made scorch") },
		earthquake:        func(types.Pointf, int) { t.Fatal("airborne meteor made earthquake") },
		newObjectByTypeID: func(string) *Object { t.Fatal("airborne meteor created explosion"); return nil },
		delayedDelete:     func(*Object) { t.Fatal("airborne meteor was deleted") },
	}
	for _, height := range []float32{1, float32(math.NaN())} {
		meteorUpdateNative53D6E0(&Object{ZVal: height}, runtime)
	}
}

func TestMeteorFloatToInt53D6E0(t *testing.T) {
	tests := []struct {
		value float32
		want  int32
	}{
		{0.5, 0},
		{1.5, 2},
		{-1.5, -2},
		{float32(math.NaN()), math.MinInt32},
		{float32(math.Inf(1)), math.MinInt32},
		{float32(math.Inf(-1)), math.MinInt32},
	}
	for _, test := range tests {
		if got := meteorFloatToInt53D6E0(test.value); got != test.want {
			t.Errorf("meteorFloatToInt53D6E0(%v) = %d, want %d", test.value, got, test.want)
		}
	}
}
