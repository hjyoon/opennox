package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type poisonGasTrapCollideTestData4EB910 struct {
	name     string
	duration int32
}

type poisonGasTrapCollideTestObject4EB910 struct {
	name  string
	x     float32
	y     float32
	owner *poisonGasTrapCollideTestObject4EB910
	data  *poisonGasTrapCollideTestData4EB910
}

type poisonGasTrapCollideTestFixture4EB910 struct {
	events           []string
	allowed          int32
	cloud            *poisonGasTrapCollideTestObject4EB910
	lifetime         float32
	fps              uint32
	onNewObject      func()
	onLoadY          func()
	onLoadX          func()
	onLoadOwner      func()
	onCreateAt       func()
	onLoadUpdateData func()
	onLoadLifetime   func()
	multiplyResult   float32
	roundResult      int32
}

func poisonGasTrapObjectName4EB910(obj *poisonGasTrapCollideTestObject4EB910) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func poisonGasTrapDataName4EB910(data *poisonGasTrapCollideTestData4EB910) string {
	if data == nil {
		return "nil"
	}
	return data.name
}

func (f *poisonGasTrapCollideTestFixture4EB910) event(format string, args ...any) {
	f.events = append(f.events, fmt.Sprintf(format, args...))
}

func (f *poisonGasTrapCollideTestFixture4EB910) hooks() poisonGasTrapCollideHooks4EB910[
	*poisonGasTrapCollideTestObject4EB910,
	*poisonGasTrapCollideTestData4EB910,
] {
	return poisonGasTrapCollideHooks4EB910[
		*poisonGasTrapCollideTestObject4EB910,
		*poisonGasTrapCollideTestData4EB910,
	]{
		allowed: func(source, target *poisonGasTrapCollideTestObject4EB910) int32 {
			f.event("allowed:%s:%s=%d", poisonGasTrapObjectName4EB910(source), poisonGasTrapObjectName4EB910(target), f.allowed)
			return f.allowed
		},
		newObject: func(name string) *poisonGasTrapCollideTestObject4EB910 {
			f.event("new:%s=%s", name, poisonGasTrapObjectName4EB910(f.cloud))
			if f.onNewObject != nil {
				f.onNewObject()
			}
			return f.cloud
		},
		loadPosY: func(obj *poisonGasTrapCollideTestObject4EB910) float32 {
			if obj == nil {
				f.event("y:nil")
				panic("nil source Y")
			}
			value := obj.y
			f.event("y:%s=%g", obj.name, value)
			if f.onLoadY != nil {
				f.onLoadY()
			}
			return value
		},
		loadPosX: func(obj *poisonGasTrapCollideTestObject4EB910) float32 {
			if obj == nil {
				f.event("x:nil")
				panic("nil source X")
			}
			value := obj.x
			f.event("x:%s=%g", obj.name, value)
			if f.onLoadX != nil {
				f.onLoadX()
			}
			return value
		},
		loadOwner: func(obj *poisonGasTrapCollideTestObject4EB910) *poisonGasTrapCollideTestObject4EB910 {
			if obj == nil {
				f.event("owner:nil")
				panic("nil source owner")
			}
			value := obj.owner
			f.event("owner:%s=%s", obj.name, poisonGasTrapObjectName4EB910(value))
			if f.onLoadOwner != nil {
				f.onLoadOwner()
			}
			return value
		},
		createAt: func(obj, owner *poisonGasTrapCollideTestObject4EB910, x, y float32, reserved uint32) {
			f.event("create:%s:%s:%g:%g:%d", poisonGasTrapObjectName4EB910(obj), poisonGasTrapObjectName4EB910(owner), x, y, reserved)
			if f.onCreateAt != nil {
				f.onCreateAt()
			}
		},
		loadUpdateData: func(obj *poisonGasTrapCollideTestObject4EB910) *poisonGasTrapCollideTestData4EB910 {
			data := obj.data
			f.event("update:%s=%s", poisonGasTrapObjectName4EB910(obj), poisonGasTrapDataName4EB910(data))
			if f.onLoadUpdateData != nil {
				f.onLoadUpdateData()
			}
			return data
		},
		loadLifetime: func(key string) float32 {
			value := f.lifetime
			f.event("lifetime:%s=%g", key, value)
			if f.onLoadLifetime != nil {
				f.onLoadLifetime()
			}
			return value
		},
		loadFPS: func() uint32 {
			f.event("fps=%#x", f.fps)
			return f.fps
		},
		multiply: func(lifetime float32, fps uint32) float32 {
			f.event("multiply:%g:%#x=%g", lifetime, fps, f.multiplyResult)
			return f.multiplyResult
		},
		floatToInt: func(value float32) int32 {
			f.event("round:%g=%d", value, f.roundResult)
			return f.roundResult
		},
		storeDuration: func(data *poisonGasTrapCollideTestData4EB910, duration int32) {
			f.event("store:%s=%d", poisonGasTrapDataName4EB910(data), duration)
			if data == nil {
				panic("nil cloud UpdateData")
			}
			data.duration = duration
		},
		audio: func(id uint32, obj *poisonGasTrapCollideTestObject4EB910, kind int32, code uint32) {
			f.event("audio:%d:%s:%d:%d", id, poisonGasTrapObjectName4EB910(obj), kind, code)
		},
		delayedDelete: func(obj *poisonGasTrapCollideTestObject4EB910) {
			f.event("delete:%s", poisonGasTrapObjectName4EB910(obj))
		},
	}
}

func assertPoisonGasTrapEvents4EB910(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("event order mismatch\n got: %#v\nwant: %#v", got, want)
	}
}

func TestPoisonGasTrapCollide4EB910NilTargetGuardsBeforeSource(t *testing.T) {
	f := &poisonGasTrapCollideTestFixture4EB910{allowed: 1}
	collision := &struct{ guard uint32 }{guard: 0x31415926}
	poisonGasTrapCollide4EB910((*poisonGasTrapCollideTestObject4EB910)(nil), (*poisonGasTrapCollideTestObject4EB910)(nil), collision, f.hooks())
	assertPoisonGasTrapEvents4EB910(t, f.events, nil)
	if collision.guard != 0x31415926 {
		t.Fatalf("collision was modified: %#x", collision.guard)
	}
}

func TestPoisonGasTrapCollide4EB910GateUsesCTruthiness(t *testing.T) {
	source := &poisonGasTrapCollideTestObject4EB910{name: "source"}
	target := &poisonGasTrapCollideTestObject4EB910{name: "target"}
	for _, allowed := range []int32{0, -7} {
		t.Run(fmt.Sprintf("result_%d", allowed), func(t *testing.T) {
			f := &poisonGasTrapCollideTestFixture4EB910{allowed: allowed}
			poisonGasTrapCollide4EB910(source, target, nil, f.hooks())
			want := []string{fmt.Sprintf("allowed:source:target=%d", allowed)}
			if allowed != 0 {
				want = append(want, "new:ToxicCloud=nil")
			}
			assertPoisonGasTrapEvents4EB910(t, f.events, want)
		})
	}
}

func TestPoisonGasTrapCollide4EB910AllocationFailureSkipsSourceFieldsAndEffects(t *testing.T) {
	source := &poisonGasTrapCollideTestObject4EB910{name: "source", x: 12, y: -8}
	target := &poisonGasTrapCollideTestObject4EB910{name: "target"}
	f := &poisonGasTrapCollideTestFixture4EB910{allowed: 1}
	poisonGasTrapCollide4EB910(source, target, new(int), f.hooks())
	assertPoisonGasTrapEvents4EB910(t, f.events, []string{
		"allowed:source:target=1", "new:ToxicCloud=nil",
	})
}

func TestPoisonGasTrapCollide4EB910SuccessOrderAndLiveLoads(t *testing.T) {
	entryOwner := &poisonGasTrapCollideTestObject4EB910{name: "entry-owner"}
	liveOwner := &poisonGasTrapCollideTestObject4EB910{name: "live-owner"}
	afterOwner := &poisonGasTrapCollideTestObject4EB910{name: "after-owner"}
	entryData := &poisonGasTrapCollideTestData4EB910{name: "entry-data"}
	liveData := &poisonGasTrapCollideTestData4EB910{name: "live-data"}
	afterData := &poisonGasTrapCollideTestData4EB910{name: "after-data"}
	source := &poisonGasTrapCollideTestObject4EB910{name: "source", x: 1, y: 2, owner: entryOwner}
	target := &poisonGasTrapCollideTestObject4EB910{name: "target"}
	cloud := &poisonGasTrapCollideTestObject4EB910{name: "cloud", data: entryData}
	f := &poisonGasTrapCollideTestFixture4EB910{
		allowed: 9, cloud: cloud, lifetime: 12.5, fps: 0x11111111,
		multiplyResult: -9.5, roundResult: -10,
	}
	f.onNewObject = func() {
		source.x = 11
		source.y = 22
	}
	f.onLoadY = func() {
		source.x = 33
	}
	f.onLoadX = func() {
		source.owner = liveOwner
	}
	f.onLoadOwner = func() {
		source.owner = afterOwner
	}
	f.onCreateAt = func() {
		cloud.data = liveData
	}
	f.onLoadUpdateData = func() {
		cloud.data = afterData
		f.lifetime = 17.25
	}
	f.onLoadLifetime = func() {
		f.fps = 0xffffffff
	}

	collision := &struct{ guard uint32 }{guard: 0x27182818}
	poisonGasTrapCollide4EB910(source, target, collision, f.hooks())
	assertPoisonGasTrapEvents4EB910(t, f.events, []string{
		"allowed:source:target=9",
		"new:ToxicCloud=cloud",
		"y:source=22",
		"x:source=33",
		"owner:source=live-owner",
		"create:cloud:live-owner:33:22:0",
		"update:cloud=live-data",
		"lifetime:ToxicCloudLifetime=17.25",
		"fps=0xffffffff",
		"multiply:17.25:0xffffffff=-9.5",
		"round:-9.5=-10",
		"store:live-data=-10",
		"audio:847:source:0:0",
		"delete:source",
	})
	if liveData.duration != -10 || entryData.duration != 0 || afterData.duration != 0 {
		t.Fatalf("duration stores = entry %d, live %d, after %d", entryData.duration, liveData.duration, afterData.duration)
	}
	if collision.guard != 0x27182818 {
		t.Fatalf("collision was modified: %#x", collision.guard)
	}
}

func TestPoisonGasTrapCollide4EB910NilSourceFaultsOnlyAfterAllocation(t *testing.T) {
	target := &poisonGasTrapCollideTestObject4EB910{name: "target"}
	cloud := &poisonGasTrapCollideTestObject4EB910{name: "cloud"}
	f := &poisonGasTrapCollideTestFixture4EB910{allowed: 1, cloud: cloud}
	defer func() {
		if recover() == nil {
			t.Fatal("nil source did not fault")
		}
		assertPoisonGasTrapEvents4EB910(t, f.events, []string{
			"allowed:nil:target=1", "new:ToxicCloud=cloud", "y:nil",
		})
	}()
	poisonGasTrapCollide4EB910((*poisonGasTrapCollideTestObject4EB910)(nil), target, nil, f.hooks())
}

func TestPoisonGasTrapCollide4EB910NilUpdateDataFaultsAtStore(t *testing.T) {
	source := &poisonGasTrapCollideTestObject4EB910{name: "source", x: 1, y: 2}
	target := &poisonGasTrapCollideTestObject4EB910{name: "target"}
	cloud := &poisonGasTrapCollideTestObject4EB910{name: "cloud"}
	f := &poisonGasTrapCollideTestFixture4EB910{
		allowed: 1, cloud: cloud, lifetime: 2.5, fps: 30,
		multiplyResult: 75, roundResult: 75,
	}
	defer func() {
		if recover() == nil {
			t.Fatal("nil UpdateData did not fault")
		}
		assertPoisonGasTrapEvents4EB910(t, f.events, []string{
			"allowed:source:target=1", "new:ToxicCloud=cloud",
			"y:source=2", "x:source=1", "owner:source=nil",
			"create:cloud:nil:1:2:0", "update:cloud=nil",
			"lifetime:ToxicCloudLifetime=2.5", "fps=0x1e",
			"multiply:2.5:0x1e=75", "round:75=75", "store:nil=75",
		})
	}()
	poisonGasTrapCollide4EB910(source, target, nil, f.hooks())
}

func TestPoisonGasTrapMultiply4EB910UsesSignedFPSAndBinary32Spill(t *testing.T) {
	if got := poisonGasTrapMultiply4EB910(1.25, 0xffffffff); got != -1.25 {
		t.Fatalf("signed FPS product = %g, want -1.25", got)
	}
	// 16777217 must enter FIMUL as an exact signed integer. Converting FPS to
	// float32 before multiplication would produce bits 0x4b800001 instead.
	lifetime := math.Float32frombits(0x3f800001)
	if got := math.Float32bits(poisonGasTrapMultiply4EB910(lifetime, 16777217)); got != 0x4b800002 {
		t.Fatalf("binary32 product bits = %#x, want 0x4b800002", got)
	}
}

func TestPoisonGasTrapRound4EB910MatchesFISTP(t *testing.T) {
	tests := []struct {
		value float32
		want  int32
	}{
		{0.5, 0}, {1.5, 2}, {2.5, 2}, {3.5, 4},
		{-0.5, 0}, {-1.5, -2}, {-2.5, -2},
		{math.Float32frombits(0x4effffff), 2147483520},
		{math.Float32frombits(0x4f000000), math.MinInt32},
		{math.Float32frombits(0xcf000000), math.MinInt32},
		{float32(math.Inf(1)), math.MinInt32},
		{float32(math.Inf(-1)), math.MinInt32},
		{float32(math.NaN()), math.MinInt32},
	}
	for _, tc := range tests {
		if got := poisonGasTrapRound4EB910(tc.value); got != tc.want {
			t.Errorf("round(%08x) = %d, want %d", math.Float32bits(tc.value), got, tc.want)
		}
	}
}
