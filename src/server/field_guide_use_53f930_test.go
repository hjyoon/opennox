package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type fieldGuideUseTestData53F930 struct {
	creature string
}

type fieldGuideUseTestPlayer53F930 struct {
	name   string
	class  uint8
	guides map[int32]uint32
}

type fieldGuideUseTestUpdate53F930 struct {
	player *fieldGuideUseTestPlayer53F930
}

type fieldGuideUseTestObject53F930 struct {
	name   string
	class  uint8
	update *fieldGuideUseTestUpdate53F930
	data   *fieldGuideUseTestData53F930
}

type fieldGuideUseTestWorld53F930 struct {
	owner       *fieldGuideUseTestObject53F930
	item        *fieldGuideUseTestObject53F930
	guide       int32
	flags       int32
	awardResult int32
	events      []string
	faultAt     int
	afterFlags  func()
	afterClass  func()
	awardArgs   struct {
		owner  *fieldGuideUseTestObject53F930
		guide  int32
		notify int32
	}
}

func fieldGuideUseTestObjectName53F930(obj *fieldGuideUseTestObject53F930) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func (w *fieldGuideUseTestWorld53F930) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(value)
	}
}

func (w *fieldGuideUseTestWorld53F930) hooks() fieldGuideUseHooks53F930[
	*fieldGuideUseTestObject53F930,
	*fieldGuideUseTestUpdate53F930,
	*fieldGuideUseTestPlayer53F930,
	*fieldGuideUseTestData53F930,
] {
	return fieldGuideUseHooks53F930[
		*fieldGuideUseTestObject53F930,
		*fieldGuideUseTestUpdate53F930,
		*fieldGuideUseTestPlayer53F930,
		*fieldGuideUseTestData53F930,
	]{
		loadOwnerArg: func() *fieldGuideUseTestObject53F930 {
			w.event("owner:" + fieldGuideUseTestObjectName53F930(w.owner))
			return w.owner
		},
		loadClassLow: func(owner *fieldGuideUseTestObject53F930) uint8 {
			w.event(fmt.Sprintf("class:%s=%02x", fieldGuideUseTestObjectName53F930(owner), owner.class))
			return owner.class
		},
		loadItemArg: func() *fieldGuideUseTestObject53F930 {
			w.event("item:" + fieldGuideUseTestObjectName53F930(w.item))
			return w.item
		},
		loadUpdateData: func(owner *fieldGuideUseTestObject53F930) *fieldGuideUseTestUpdate53F930 {
			w.event("update:" + fieldGuideUseTestObjectName53F930(owner))
			return owner.update
		},
		loadUseData: func(item *fieldGuideUseTestObject53F930) *fieldGuideUseTestData53F930 {
			w.event("data:" + fieldGuideUseTestObjectName53F930(item))
			return item.data
		},
		loadCreature: func(data *fieldGuideUseTestData53F930) string {
			w.event("creature:" + data.creature)
			return data.creature
		},
		guideByName: func(name string) int32 {
			w.event(fmt.Sprintf("guide:%s=%08x", name, uint32(w.guide)))
			return w.guide
		},
		gameFlagsCheck: func(mask uint32) int32 {
			w.event(fmt.Sprintf("flags:%08x=%08x", mask, uint32(w.flags)))
			if w.afterFlags != nil {
				w.afterFlags()
			}
			return w.flags
		},
		loadPlayer: func(update *fieldGuideUseTestUpdate53F930) *fieldGuideUseTestPlayer53F930 {
			w.event("player:" + update.player.name)
			return update.player
		},
		loadPlayerClass: func(player *fieldGuideUseTestPlayer53F930) uint8 {
			w.event(fmt.Sprintf("player-class:%s=%02x", player.name, player.class))
			if w.afterClass != nil {
				w.afterClass()
			}
			return player.class
		},
		loadGuideLevel: func(player *fieldGuideUseTestPlayer53F930, guide int32) uint32 {
			level := player.guides[guide]
			w.event(fmt.Sprintf("level:%s:%08x=%08x", player.name, uint32(guide), level))
			return level
		},
		primaryMessage: func(owner *fieldGuideUseTestObject53F930, message string, value uint8) {
			w.event(fmt.Sprintf("primary:%s:%s:%d", fieldGuideUseTestObjectName53F930(owner), message, value))
		},
		awardGuide: func(owner *fieldGuideUseTestObject53F930, guide, notify int32) int32 {
			w.awardArgs.owner = owner
			w.awardArgs.guide = guide
			w.awardArgs.notify = notify
			w.event(fmt.Sprintf("award:%s:%08x:%08x=%08x", fieldGuideUseTestObjectName53F930(owner), uint32(guide), uint32(notify), uint32(w.awardResult)))
			return w.awardResult
		},
		delayedDeleteItem: func(item *fieldGuideUseTestObject53F930) {
			w.event("delete:" + fieldGuideUseTestObjectName53F930(item))
		},
	}
}

func newFieldGuideUseTestWorld53F930() *fieldGuideUseTestWorld53F930 {
	player := &fieldGuideUseTestPlayer53F930{
		name:   "player",
		class:  fieldGuideUseConjurerClass53F930,
		guides: make(map[int32]uint32),
	}
	owner := &fieldGuideUseTestObject53F930{
		name:   "owner",
		class:  0xf4,
		update: &fieldGuideUseTestUpdate53F930{player: player},
	}
	item := &fieldGuideUseTestObject53F930{
		name: "item",
		data: &fieldGuideUseTestData53F930{creature: "CarnivorousPlant"},
	}
	return &fieldGuideUseTestWorld53F930{
		owner:       owner,
		item:        item,
		guide:       24,
		flags:       math.MinInt32,
		awardResult: math.MinInt32,
	}
}

func fieldGuideUseSuccessTrace53F930() []string {
	return []string{
		"owner:owner",
		"class:owner=f4",
		"item:item",
		"update:owner",
		"data:item",
		"creature:CarnivorousPlant",
		"guide:CarnivorousPlant=00000018",
		"flags:00001000=80000000",
		"player:player",
		"player-class:player=02",
		"player:player",
		"level:player:00000018=00000000",
		"award:owner:00000018:00000001=80000000",
		"delete:item",
	}
}

func TestUseFieldGuide53F930ExactSuccessTraceAndFaultPrefixes(t *testing.T) {
	want := fieldGuideUseSuccessTrace53F930()
	build := newFieldGuideUseTestWorld53F930

	w := build()
	if got := useFieldGuide53F930(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if w.awardArgs.owner != w.owner || w.awardArgs.guide != 24 || w.awardArgs.notify != 1 {
		t.Fatalf("award args = %p/%d/%d, want %p/24/1", w.awardArgs.owner, w.awardArgs.guide, w.awardArgs.notify, w.owner)
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%02d", faultAt), func(t *testing.T) {
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
			useFieldGuide53F930(w.hooks())
		})
	}
}

func TestUseFieldGuide53F930GatesAndCanonicalReturns(t *testing.T) {
	t.Run("non-player-does-not-load-item", func(t *testing.T) {
		w := newFieldGuideUseTestWorld53F930()
		w.owner.class = 0xf0
		if got := useFieldGuide53F930(w.hooks()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		want := []string{"owner:owner", "class:owner=f0"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	})

	t.Run("quest-requires-exact-conjurer", func(t *testing.T) {
		w := newFieldGuideUseTestWorld53F930()
		w.owner.update.player.class = 0x82
		if got := useFieldGuide53F930(w.hooks()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		wantTail := []string{
			"flags:00001000=80000000",
			"player:player",
			"player-class:player=82",
			"primary:owner:pickup.c:ObjectEquipClassFail:0",
		}
		if got := w.events[len(w.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
			t.Fatalf("tail = %v, want %v", got, wantTail)
		}
	})

	t.Run("non-quest-skips-class-and-rejects-duplicate", func(t *testing.T) {
		w := newFieldGuideUseTestWorld53F930()
		w.flags = 0
		w.owner.update.player.class = 0xff
		w.owner.update.player.guides[w.guide] = math.MaxUint32
		if got := useFieldGuide53F930(w.hooks()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		wantTail := []string{
			"flags:00001000=00000000",
			"player:player",
			"level:player:00000018=ffffffff",
			"primary:owner:objcoll.c:AlreadyHaveGuide:0",
		}
		if got := w.events[len(w.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
			t.Fatalf("tail = %v, want %v", got, wantTail)
		}
	})
}

func TestUseFieldGuide53F930ReloadsPlayerFromCachedUpdate(t *testing.T) {
	w := newFieldGuideUseTestWorld53F930()
	replacement := &fieldGuideUseTestPlayer53F930{
		name:   "replacement",
		class:  fieldGuideUseConjurerClass53F930,
		guides: map[int32]uint32{w.guide: 7},
	}
	w.afterClass = func() {
		w.owner.update.player = replacement
	}
	if got := useFieldGuide53F930(w.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	wantTail := []string{
		"player-class:player=02",
		"player:replacement",
		"level:replacement:00000018=00000007",
		"primary:owner:objcoll.c:AlreadyHaveGuide:0",
	}
	if got := w.events[len(w.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
		t.Fatalf("tail = %v, want %v", got, wantTail)
	}
}
