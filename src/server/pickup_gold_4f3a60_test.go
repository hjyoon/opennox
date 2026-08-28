package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type pickupGoldTestObject4F3A60 struct {
	name     string
	classLow uint8
	data     *pickupGoldTestData4F3A60
}

type pickupGoldTestData4F3A60 struct {
	name   string
	amount uint32
}

type pickupGoldTestWorld4F3A60 struct {
	arg3, arg4    int32
	defaultResult int32
	events        []string
	faultAt       int
	afterAdd      func(*pickupGoldTestWorld4F3A60)
	afterDelete   func(*pickupGoldTestWorld4F3A60)
	afterDefault  func(*pickupGoldTestWorld4F3A60)
}

func pickupGoldTestObjectName4F3A60(obj *pickupGoldTestObject4F3A60) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func pickupGoldTestDataName4F3A60(data *pickupGoldTestData4F3A60) string {
	if data == nil {
		return "nil"
	}
	return data.name
}

func (w *pickupGoldTestWorld4F3A60) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(value)
	}
}

func (w *pickupGoldTestWorld4F3A60) hooks() pickupGoldHooks4F3A60[
	*pickupGoldTestObject4F3A60,
	*pickupGoldTestData4F3A60,
	string,
] {
	return pickupGoldHooks4F3A60[
		*pickupGoldTestObject4F3A60,
		*pickupGoldTestData4F3A60,
		string,
	]{
		loadOwnerClassLow: func(owner *pickupGoldTestObject4F3A60) uint8 {
			value := owner.classLow
			w.event(fmt.Sprintf("class:%s=%02x", pickupGoldTestObjectName4F3A60(owner), value))
			return value
		},
		loadGoldInitData: func(item *pickupGoldTestObject4F3A60) *pickupGoldTestData4F3A60 {
			data := item.data
			w.event("init:" + pickupGoldTestObjectName4F3A60(item) + "=" + pickupGoldTestDataName4F3A60(data))
			return data
		},
		loadGoldAmount: func(data *pickupGoldTestData4F3A60) uint32 {
			value := data.amount
			w.event(fmt.Sprintf("amount:%s=%08x", pickupGoldTestDataName4F3A60(data), value))
			return value
		},
		addGold: func(owner *pickupGoldTestObject4F3A60, amount uint32) {
			w.event(fmt.Sprintf("add-gold:%s=%08x", pickupGoldTestObjectName4F3A60(owner), amount))
			if w.afterAdd != nil {
				w.afterAdd(w)
			}
		},
		delayedDelete: func(item *pickupGoldTestObject4F3A60) {
			w.event("delete:" + pickupGoldTestObjectName4F3A60(item))
			if w.afterDelete != nil {
				w.afterDelete(w)
			}
		},
		loadString: func(key, path string, line int) string {
			w.event(fmt.Sprintf("string:%s:%s:%d", key, path, line))
			return "localized"
		},
		sendLineMessage: func(owner *pickupGoldTestObject4F3A60, message string, amount uint32) {
			w.event(fmt.Sprintf("line:%s:%s:%08x", pickupGoldTestObjectName4F3A60(owner), message, amount))
		},
		audio: func(id uint32, owner *pickupGoldTestObject4F3A60, kind int32, code uint32) {
			w.event(fmt.Sprintf("audio:%d:%s:%d:%08x", id, pickupGoldTestObjectName4F3A60(owner), kind, code))
		},
		loadArg4: func() int32 {
			w.event(fmt.Sprintf("arg4=%08x", uint32(w.arg4)))
			return w.arg4
		},
		loadArg3: func() int32 {
			w.event(fmt.Sprintf("arg3=%08x", uint32(w.arg3)))
			return w.arg3
		},
		defaultPickup: func(owner, item *pickupGoldTestObject4F3A60, arg3, arg4 int32) int32 {
			w.event(fmt.Sprintf("default:%s:%s:%08x:%08x", pickupGoldTestObjectName4F3A60(owner), pickupGoldTestObjectName4F3A60(item), uint32(arg3), uint32(arg4)))
			if w.afterDefault != nil {
				w.afterDefault(w)
			}
			return w.defaultResult
		},
	}
}

func verifyPickupGoldFaultPrefixes4F3A60(
	t *testing.T,
	want []string,
	build func() (*pickupGoldTestWorld4F3A60, *pickupGoldTestObject4F3A60, *pickupGoldTestObject4F3A60),
) {
	t.Helper()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w, owner, item := build()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %v, want %v", w.events, want[:faultAt])
				}
			}()
			pickupGold4F3A60(owner, item, w.hooks())
		})
	}
}

func pickupGoldPlayerBuild4F3A60() (*pickupGoldTestWorld4F3A60, *pickupGoldTestObject4F3A60, *pickupGoldTestObject4F3A60) {
	cached := &pickupGoldTestData4F3A60{name: "cached", amount: 0x80000001}
	replacement := &pickupGoldTestData4F3A60{name: "replacement", amount: 0x11111111}
	owner := &pickupGoldTestObject4F3A60{name: "owner", classLow: 0x84}
	item := &pickupGoldTestObject4F3A60{name: "item", data: cached}
	w := &pickupGoldTestWorld4F3A60{arg3: 3, arg4: 4, defaultResult: math.MinInt32}
	w.afterAdd = func(*pickupGoldTestWorld4F3A60) {
		cached.amount = 0x10203040
		item.data = replacement
	}
	w.afterDelete = func(*pickupGoldTestWorld4F3A60) {
		cached.amount = 0xfedcba98
		replacement.amount = 0x22222222
	}
	return w, owner, item
}

func pickupGoldPlayerTrace4F3A60() []string {
	return []string{
		"class:owner=84",
		"init:item=cached",
		"amount:cached=80000001",
		"add-gold:owner=80000001",
		"delete:item",
		"amount:cached=fedcba98",
		`string:GoldPickup:C:\NoxPost\src\Server\Object\pickdrop\pickup.c:709`,
		"line:owner:localized:fedcba98",
		"audio:307:owner:0:00000000",
	}
}

func TestPickupGold4F3A60PlayerCachesDataAndReloadsAmount(t *testing.T) {
	w, owner, item := pickupGoldPlayerBuild4F3A60()
	if got := pickupGold4F3A60(owner, item, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := pickupGoldPlayerTrace4F3A60()
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events =\n%v\nwant =\n%v", w.events, want)
	}
	verifyPickupGoldFaultPrefixes4F3A60(t, want, pickupGoldPlayerBuild4F3A60)
}

func pickupGoldNonPlayerBuild4F3A60() (*pickupGoldTestWorld4F3A60, *pickupGoldTestObject4F3A60, *pickupGoldTestObject4F3A60) {
	owner := &pickupGoldTestObject4F3A60{name: "owner", classLow: 0x80}
	item := &pickupGoldTestObject4F3A60{name: "item"}
	w := &pickupGoldTestWorld4F3A60{
		arg3:          math.MinInt32,
		arg4:          math.MaxInt32,
		defaultResult: math.MinInt32,
	}
	return w, owner, item
}

func pickupGoldNonPlayerTrace4F3A60() []string {
	return []string{
		"class:owner=80",
		"arg4=7fffffff",
		"arg3=80000000",
		"default:owner:item:80000000:7fffffff",
		"audio:307:owner:0:00000000",
	}
}

func TestPickupGold4F3A60NonPlayerForwardsFourArgsAndExactResult(t *testing.T) {
	w, owner, item := pickupGoldNonPlayerBuild4F3A60()
	if got := pickupGold4F3A60(owner, item, w.hooks()); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	want := pickupGoldNonPlayerTrace4F3A60()
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	verifyPickupGoldFaultPrefixes4F3A60(t, want, pickupGoldNonPlayerBuild4F3A60)
}

func TestPickupGold4F3A60DefaultZeroSkipsAudio(t *testing.T) {
	owner := &pickupGoldTestObject4F3A60{name: "owner", classLow: 0}
	w := &pickupGoldTestWorld4F3A60{arg3: -3, arg4: -4}
	if got := pickupGold4F3A60(owner, nil, w.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{
		"class:owner=00",
		"arg4=fffffffc",
		"arg3=fffffffd",
		"default:owner:nil:fffffffd:fffffffc",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestPickupGold4F3A60HasNoNilGuards(t *testing.T) {
	t.Run("owner", func(t *testing.T) {
		w := &pickupGoldTestWorld4F3A60{}
		defer func() {
			if recover() == nil {
				t.Fatal("nil owner did not preserve the class-load fault")
			}
			if len(w.events) != 0 {
				t.Fatalf("events before nil owner fault = %v", w.events)
			}
		}()
		pickupGold4F3A60(nil, nil, w.hooks())
	})

	t.Run("player item", func(t *testing.T) {
		owner := &pickupGoldTestObject4F3A60{name: "owner", classLow: 4}
		w := &pickupGoldTestWorld4F3A60{}
		defer func() {
			if recover() == nil {
				t.Fatal("nil Player item did not preserve the InitData-load fault")
			}
			want := []string{"class:owner=04"}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %v, want %v", w.events, want)
			}
		}()
		pickupGold4F3A60(owner, nil, w.hooks())
	})
}
