package server

import (
	"fmt"
	"reflect"
	"testing"
)

type bearTrapCollideTestObject4EB890 struct {
	name  string
	x     float32
	y     float32
	owner *bearTrapCollideTestObject4EB890
}

type bearTrapCollideTestFixture4EB890 struct {
	events      []string
	allowed     int32
	closed      *bearTrapCollideTestObject4EB890
	onNewObject func()
	onLoadY     func()
	onLoadX     func()
	onLoadOwner func()
}

func bearTrapObjectName4EB890(obj *bearTrapCollideTestObject4EB890) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func (f *bearTrapCollideTestFixture4EB890) event(format string, args ...any) {
	f.events = append(f.events, fmt.Sprintf(format, args...))
}

func (f *bearTrapCollideTestFixture4EB890) hooks() bearTrapCollideHooks4EB890[*bearTrapCollideTestObject4EB890] {
	return bearTrapCollideHooks4EB890[*bearTrapCollideTestObject4EB890]{
		allowed: func(source, target *bearTrapCollideTestObject4EB890) int32 {
			f.event("allowed:%s:%s=%d", bearTrapObjectName4EB890(source), bearTrapObjectName4EB890(target), f.allowed)
			return f.allowed
		},
		newObject: func(name string) *bearTrapCollideTestObject4EB890 {
			f.event("new:%s=%s", name, bearTrapObjectName4EB890(f.closed))
			if f.onNewObject != nil {
				f.onNewObject()
			}
			return f.closed
		},
		loadPosY: func(obj *bearTrapCollideTestObject4EB890) float32 {
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
		loadPosX: func(obj *bearTrapCollideTestObject4EB890) float32 {
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
		loadOwner: func(obj *bearTrapCollideTestObject4EB890) *bearTrapCollideTestObject4EB890 {
			if obj == nil {
				f.event("owner:nil")
				panic("nil source owner")
			}
			value := obj.owner
			f.event("owner:%s=%s", obj.name, bearTrapObjectName4EB890(value))
			if f.onLoadOwner != nil {
				f.onLoadOwner()
			}
			return value
		},
		createAt: func(obj, owner *bearTrapCollideTestObject4EB890, x, y float32, reserved uint32) {
			f.event("create:%s:%s:%g:%g:%d", bearTrapObjectName4EB890(obj), bearTrapObjectName4EB890(owner), x, y, reserved)
		},
		delayedDelete: func(obj *bearTrapCollideTestObject4EB890) {
			f.event("delete:%s", bearTrapObjectName4EB890(obj))
		},
		applyEnchant: func(obj *bearTrapCollideTestObject4EB890, enchant, duration, power uint32) {
			f.event("enchant:%s:%d:%d:%d", bearTrapObjectName4EB890(obj), enchant, duration, power)
		},
		audio: func(id uint32, obj *bearTrapCollideTestObject4EB890, kind int32, code uint32) {
			f.event("audio:%d:%s:%d:%d", id, bearTrapObjectName4EB890(obj), kind, code)
		},
	}
}

func assertBearTrapEvents4EB890(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("event order mismatch\n got: %#v\nwant: %#v", got, want)
	}
}

func TestBearTrapCollide4EB890NilTargetGuardsBeforeSource(t *testing.T) {
	f := &bearTrapCollideTestFixture4EB890{allowed: 1}
	collision := &struct{ guard uint32 }{guard: 0x31415926}
	bearTrapCollide4EB890((*bearTrapCollideTestObject4EB890)(nil), (*bearTrapCollideTestObject4EB890)(nil), collision, f.hooks())
	assertBearTrapEvents4EB890(t, f.events, nil)
	if collision.guard != 0x31415926 {
		t.Fatalf("collision was modified: %#x", collision.guard)
	}
}

func TestBearTrapCollide4EB890GateUsesCTruthiness(t *testing.T) {
	source := &bearTrapCollideTestObject4EB890{name: "source"}
	target := &bearTrapCollideTestObject4EB890{name: "target"}
	for _, allowed := range []int32{0, -7} {
		t.Run(fmt.Sprintf("result_%d", allowed), func(t *testing.T) {
			f := &bearTrapCollideTestFixture4EB890{allowed: allowed}
			bearTrapCollide4EB890(source, target, nil, f.hooks())
			want := []string{fmt.Sprintf("allowed:source:target=%d", allowed)}
			if allowed != 0 {
				want = append(want, "new:ClosedBearTrap=nil")
			}
			assertBearTrapEvents4EB890(t, f.events, want)
		})
	}
}

func TestBearTrapCollide4EB890AllocationFailureSkipsSourceFieldsAndEffects(t *testing.T) {
	source := &bearTrapCollideTestObject4EB890{name: "source", x: 12, y: -8}
	target := &bearTrapCollideTestObject4EB890{name: "target"}
	f := &bearTrapCollideTestFixture4EB890{allowed: 1}
	bearTrapCollide4EB890(source, target, new(int), f.hooks())
	assertBearTrapEvents4EB890(t, f.events, []string{
		"allowed:source:target=1", "new:ClosedBearTrap=nil",
	})
}

func TestBearTrapCollide4EB890SuccessOrderAndLiveFieldLoads(t *testing.T) {
	entryOwner := &bearTrapCollideTestObject4EB890{name: "entry-owner"}
	liveOwner := &bearTrapCollideTestObject4EB890{name: "live-owner"}
	afterOwner := &bearTrapCollideTestObject4EB890{name: "after-owner"}
	source := &bearTrapCollideTestObject4EB890{name: "source", x: 1, y: 2, owner: entryOwner}
	target := &bearTrapCollideTestObject4EB890{name: "target"}
	closed := &bearTrapCollideTestObject4EB890{name: "closed"}
	f := &bearTrapCollideTestFixture4EB890{allowed: 9, closed: closed}
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
		source.x = 44
		source.y = 55
		source.owner = afterOwner
	}

	collision := &struct{ guard uint32 }{guard: 0x27182818}
	bearTrapCollide4EB890(source, target, collision, f.hooks())
	assertBearTrapEvents4EB890(t, f.events, []string{
		"allowed:source:target=9",
		"new:ClosedBearTrap=closed",
		"y:source=22",
		"x:source=33",
		"owner:source=live-owner",
		"create:closed:live-owner:33:22:0",
		"delete:source",
		"enchant:target:5:90:5",
		"enchant:target:14:90:5",
		"audio:846:source:0:0",
	})
	if collision.guard != 0x27182818 {
		t.Fatalf("collision was modified: %#x", collision.guard)
	}
}

func TestBearTrapCollide4EB890NilSourceFaultsOnlyAfterAllocation(t *testing.T) {
	target := &bearTrapCollideTestObject4EB890{name: "target"}
	closed := &bearTrapCollideTestObject4EB890{name: "closed"}
	f := &bearTrapCollideTestFixture4EB890{allowed: 1, closed: closed}
	defer func() {
		if recover() == nil {
			t.Fatal("nil source did not fault")
		}
		assertBearTrapEvents4EB890(t, f.events, []string{
			"allowed:nil:target=1", "new:ClosedBearTrap=closed", "y:nil",
		})
	}()
	bearTrapCollide4EB890((*bearTrapCollideTestObject4EB890)(nil), target, nil, f.hooks())
}
