package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type ankhTradableDropTestObject4EE370 struct {
	name string
}

type ankhTradableDropTestPoint4EE370 struct {
	name string
}

type ankhTradableDropTestWorld4EE370 struct {
	owner  *ankhTradableDropTestObject4EE370
	item   *ankhTradableDropTestObject4EE370
	point  *ankhTradableDropTestPoint4EE370
	result int32

	events   []string
	faultAt  int
	gotOwner *ankhTradableDropTestObject4EE370
	gotItem  *ankhTradableDropTestObject4EE370
	gotPoint *ankhTradableDropTestPoint4EE370

	onPoint func()
	onItem  func()
	onOwner func()
}

func (w *ankhTradableDropTestWorld4EE370) event(name string) {
	w.events = append(w.events, name)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(name)
	}
}

func (w *ankhTradableDropTestWorld4EE370) hooks() ankhTradableDropHooks4EE370[
	*ankhTradableDropTestObject4EE370,
	*ankhTradableDropTestPoint4EE370,
] {
	return ankhTradableDropHooks4EE370[
		*ankhTradableDropTestObject4EE370,
		*ankhTradableDropTestPoint4EE370,
	]{
		loadPointArg: func() *ankhTradableDropTestPoint4EE370 {
			w.event("point-arg")
			point := w.point
			if w.onPoint != nil {
				w.onPoint()
			}
			return point
		},
		loadItemArg: func() *ankhTradableDropTestObject4EE370 {
			w.event("item-arg")
			item := w.item
			if w.onItem != nil {
				w.onItem()
			}
			return item
		},
		loadOwnerArg: func() *ankhTradableDropTestObject4EE370 {
			w.event("owner-arg")
			owner := w.owner
			if w.onOwner != nil {
				w.onOwner()
			}
			return owner
		},
		defaultDrop: func(owner, item *ankhTradableDropTestObject4EE370, point *ankhTradableDropTestPoint4EE370) int32 {
			w.event("default-drop")
			w.gotOwner = owner
			w.gotItem = item
			w.gotPoint = point
			return w.result
		},
	}
}

func TestAnkhTradableDrop4EE370CachesArgumentsAndPreservesResult(t *testing.T) {
	results := []int32{0, 1, 7, -91, math.MinInt32, math.MaxInt32}
	for _, result := range results {
		t.Run(fmt.Sprintf("result_%d", result), func(t *testing.T) {
			owner := &ankhTradableDropTestObject4EE370{name: "owner"}
			item := &ankhTradableDropTestObject4EE370{name: "item"}
			point := &ankhTradableDropTestPoint4EE370{name: "point"}
			w := &ankhTradableDropTestWorld4EE370{
				owner:  owner,
				item:   item,
				point:  point,
				result: result,
			}
			w.onPoint = func() {
				w.point = &ankhTradableDropTestPoint4EE370{name: "late-point"}
			}
			w.onItem = func() {
				w.item = &ankhTradableDropTestObject4EE370{name: "late-item"}
			}
			w.onOwner = func() {
				w.owner = &ankhTradableDropTestObject4EE370{name: "late-owner"}
			}

			if got := ankhTradableDrop4EE370(w.hooks()); got != result {
				t.Fatalf("result = %d, want %d", got, result)
			}
			if w.gotOwner != owner || w.gotItem != item || w.gotPoint != point {
				t.Fatalf("DefaultDrop args = (%p, %p, %p), want (%p, %p, %p)", w.gotOwner, w.gotItem, w.gotPoint, owner, item, point)
			}
			wantEvents := []string{"point-arg", "item-arg", "owner-arg", "default-drop"}
			if !reflect.DeepEqual(w.events, wantEvents) {
				t.Fatalf("events = %v, want %v", w.events, wantEvents)
			}
		})
	}
}

func TestAnkhTradableDrop4EE370ForwardsNilArguments(t *testing.T) {
	w := &ankhTradableDropTestWorld4EE370{result: -7}
	if got := ankhTradableDrop4EE370(w.hooks()); got != -7 {
		t.Fatalf("result = %d, want -7", got)
	}
	if w.gotOwner != nil || w.gotItem != nil || w.gotPoint != nil {
		t.Fatalf("DefaultDrop args = (%p, %p, %p), want all nil", w.gotOwner, w.gotItem, w.gotPoint)
	}
}

func TestAnkhTradableDrop4EE370FaultPrefixes(t *testing.T) {
	wantEvents := []string{"point-arg", "item-arg", "owner-arg", "default-drop"}
	for faultAt := 1; faultAt <= len(wantEvents); faultAt++ {
		t.Run(fmt.Sprintf("fault_%d", faultAt), func(t *testing.T) {
			w := &ankhTradableDropTestWorld4EE370{
				owner:   &ankhTradableDropTestObject4EE370{name: "owner"},
				item:    &ankhTradableDropTestObject4EE370{name: "item"},
				point:   &ankhTradableDropTestPoint4EE370{name: "point"},
				faultAt: faultAt,
			}
			func() {
				defer func() {
					if recover() == nil {
						t.Fatal("expected injected fault")
					}
				}()
				ankhTradableDrop4EE370(w.hooks())
			}()
			if !reflect.DeepEqual(w.events, wantEvents[:faultAt]) {
				t.Fatalf("events = %v, want %v", w.events, wantEvents[:faultAt])
			}
		})
	}
}
