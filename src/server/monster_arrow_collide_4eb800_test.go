package server

import (
	"fmt"
	"reflect"
	"testing"
)

type monsterArrowCollideTestData4EB800 struct {
	name        string
	coopDamage  int32
	otherDamage int32
}

type monsterArrowCollideTestObject4EB800 struct {
	name   string
	data   *monsterArrowCollideTestData4EB800
	parent *monsterArrowCollideTestObject4EB800
	flags  uint32
}

type monsterArrowCollideTestPoint4EB800 struct {
	name string
	x    int32
	y    int32
}

type monsterArrowCollideTestFixture4EB800 struct {
	events       []string
	coop         bool
	damageResult int32
	point        *monsterArrowCollideTestPoint4EB800
	onGameFlag   func()
	onTraceY     func()
	onDamage     func()
}

func monsterArrowObjectName4EB800(obj *monsterArrowCollideTestObject4EB800) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func monsterArrowDataName4EB800(data *monsterArrowCollideTestData4EB800) string {
	if data == nil {
		return "nil"
	}
	return data.name
}

func monsterArrowPointName4EB800(point *monsterArrowCollideTestPoint4EB800) string {
	if point == nil {
		return "nil"
	}
	return point.name
}

func (f *monsterArrowCollideTestFixture4EB800) event(format string, args ...any) {
	f.events = append(f.events, fmt.Sprintf(format, args...))
}

func (f *monsterArrowCollideTestFixture4EB800) hooks() monsterArrowCollideHooks4EB800[
	*monsterArrowCollideTestObject4EB800,
	*monsterArrowCollideTestData4EB800,
	*monsterArrowCollideTestPoint4EB800,
] {
	return monsterArrowCollideHooks4EB800[
		*monsterArrowCollideTestObject4EB800,
		*monsterArrowCollideTestData4EB800,
		*monsterArrowCollideTestPoint4EB800,
	]{
		loadCollideData: func(obj *monsterArrowCollideTestObject4EB800) *monsterArrowCollideTestData4EB800 {
			f.event("data:%s", monsterArrowObjectName4EB800(obj))
			return obj.data
		},
		gameFlag: func(flag uint32) bool {
			f.event("game:%#x=%t", flag, f.coop)
			if f.onGameFlag != nil {
				f.onGameFlag()
			}
			return f.coop
		},
		loadCoopDamage: func(data *monsterArrowCollideTestData4EB800) int32 {
			f.event("coop:%s=%d", monsterArrowDataName4EB800(data), data.coopDamage)
			return data.coopDamage
		},
		loadOtherDamage: func(data *monsterArrowCollideTestData4EB800) int32 {
			f.event("other:%s=%d", monsterArrowDataName4EB800(data), data.otherDamage)
			return data.otherDamage
		},
		loadTargetFlags: func(obj *monsterArrowCollideTestObject4EB800) uint32 {
			f.event("flags:%s=%#x", monsterArrowObjectName4EB800(obj), obj.flags)
			return obj.flags
		},
		findParent: func(obj *monsterArrowCollideTestObject4EB800) *monsterArrowCollideTestObject4EB800 {
			f.event("parent:%s=%s", monsterArrowObjectName4EB800(obj), monsterArrowObjectName4EB800(obj.parent))
			return obj.parent
		},
		targetDamage: func(target, parent, source *monsterArrowCollideTestObject4EB800, damage int32, damageType uint32) int32 {
			f.event("damage:%s:%s:%s:%d:%d", monsterArrowObjectName4EB800(target), monsterArrowObjectName4EB800(parent), monsterArrowObjectName4EB800(source), damage, damageType)
			if f.onDamage != nil {
				f.onDamage()
			}
			return f.damageResult
		},
		tracePoint: func() *monsterArrowCollideTestPoint4EB800 {
			f.event("trace=%s", monsterArrowPointName4EB800(f.point))
			return f.point
		},
		loadTraceY: func(point *monsterArrowCollideTestPoint4EB800) int32 {
			f.event("y:%s=%d", monsterArrowPointName4EB800(point), point.y)
			if f.onTraceY != nil {
				f.onTraceY()
			}
			return point.y
		},
		loadTraceX: func(point *monsterArrowCollideTestPoint4EB800) int32 {
			f.event("x:%s=%d", monsterArrowPointName4EB800(point), point.x)
			return point.x
		},
		damageMap: func(x, y, damage int32, damageType uint32, source *monsterArrowCollideTestObject4EB800) {
			f.event("map:%d:%d:%d:%d:%s", x, y, damage, damageType, monsterArrowObjectName4EB800(source))
		},
		delayedDelete: func(source *monsterArrowCollideTestObject4EB800) {
			f.event("delete:%s", monsterArrowObjectName4EB800(source))
		},
	}
}

func newMonsterArrowCollideFixture4EB800() (
	*monsterArrowCollideTestFixture4EB800,
	*monsterArrowCollideTestObject4EB800,
	*monsterArrowCollideTestObject4EB800,
) {
	owner := &monsterArrowCollideTestObject4EB800{name: "owner"}
	source := &monsterArrowCollideTestObject4EB800{
		name:   "source",
		data:   &monsterArrowCollideTestData4EB800{name: "entry", coopDamage: -17, otherDamage: 29},
		parent: owner,
	}
	target := &monsterArrowCollideTestObject4EB800{name: "target"}
	return &monsterArrowCollideTestFixture4EB800{}, source, target
}

func assertMonsterArrowEvents4EB800(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("event order mismatch\n got: %#v\nwant: %#v", got, want)
	}
}

func TestMonsterArrowCollide4EB800SelectsDamageFromEntryData(t *testing.T) {
	for _, tc := range []struct {
		name     string
		coop     bool
		wantLoad string
	}{
		{name: "other", wantLoad: "other:entry=29"},
		{name: "coop", coop: true, wantLoad: "coop:entry=-17"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f, source, target := newMonsterArrowCollideFixture4EB800()
			f.coop = tc.coop
			target.flags = monsterArrowTargetReject4EB800
			f.onGameFlag = func() {
				source.data = &monsterArrowCollideTestData4EB800{name: "replacement", coopDamage: 100, otherDamage: 200}
			}
			monsterArrowCollide4EB800(source, target, new(int), f.hooks())
			want := []string{"data:source", fmt.Sprintf("game:0x800=%t", tc.coop), tc.wantLoad}
			want = append(want, "flags:target=0x8000")
			assertMonsterArrowEvents4EB800(t, f.events, want)
		})
	}
}

func TestMonsterArrowCollide4EB800RejectedTargetDoesNotDelete(t *testing.T) {
	f, source, target := newMonsterArrowCollideFixture4EB800()
	target.flags = 0xdead8000
	collision := &struct{ guard uint32 }{guard: 0x31415926}
	monsterArrowCollide4EB800(source, target, collision, f.hooks())
	assertMonsterArrowEvents4EB800(t, f.events, []string{
		"data:source", "game:0x800=false", "other:entry=29", "flags:target=0xdead8000",
	})
	if collision.guard != 0x31415926 {
		t.Fatalf("collision was modified: %#x", collision.guard)
	}
}

func TestMonsterArrowCollide4EB800TargetDamageResultIsIgnored(t *testing.T) {
	for _, result := range []int32{0, 0x100} {
		t.Run(fmt.Sprintf("result_%#x", result), func(t *testing.T) {
			f, source, target := newMonsterArrowCollideFixture4EB800()
			f.damageResult = result
			f.onDamage = func() {
				source.data.otherDamage = 777
				source.parent = target
			}
			monsterArrowCollide4EB800(source, target, nil, f.hooks())
			assertMonsterArrowEvents4EB800(t, f.events, []string{
				"data:source", "game:0x800=false", "other:entry=29", "flags:target=0x0",
				"parent:source=owner", "damage:target:owner:source:29:3", "delete:source",
			})
		})
	}
}

func TestMonsterArrowCollide4EB800WallReadsYBeforeLiveX(t *testing.T) {
	f, source, _ := newMonsterArrowCollideFixture4EB800()
	f.coop = true
	f.point = &monsterArrowCollideTestPoint4EB800{name: "impact", x: 12, y: -7}
	f.onTraceY = func() { f.point.x = 44 }
	monsterArrowCollide4EB800(source, (*monsterArrowCollideTestObject4EB800)(nil), &struct{}{}, f.hooks())
	assertMonsterArrowEvents4EB800(t, f.events, []string{
		"data:source", "game:0x800=true", "coop:entry=-17",
		"trace=impact", "y:impact=-7", "x:impact=44", "map:44:-7:-17:11:source", "delete:source",
	})
}

func TestMonsterArrowCollide4EB800WallWithoutTraceStillDeletes(t *testing.T) {
	f, source, _ := newMonsterArrowCollideFixture4EB800()
	monsterArrowCollide4EB800(source, (*monsterArrowCollideTestObject4EB800)(nil), nil, f.hooks())
	assertMonsterArrowEvents4EB800(t, f.events, []string{
		"data:source", "game:0x800=false", "other:entry=29", "trace=nil", "delete:source",
	})
}

func TestMonsterArrowCollide4EB800NilSourceFaultsBeforeMode(t *testing.T) {
	f, _, target := newMonsterArrowCollideFixture4EB800()
	defer func() {
		if recover() == nil {
			t.Fatal("nil source did not fault")
		}
		assertMonsterArrowEvents4EB800(t, f.events, []string{"data:nil"})
	}()
	monsterArrowCollide4EB800((*monsterArrowCollideTestObject4EB800)(nil), target, nil, f.hooks())
}
