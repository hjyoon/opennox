package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func fumbleTestDeps52C060(events *[]string) fumbleCastDeps52C060 {
	return fumbleCastDeps52C060{
		forceDrop:      func(*Object, *Object) int32 { *events = append(*events, "drop"); return -1 },
		dropAll:        func(*Object) { *events = append(*events, "drop-all") },
		loadTypeCache:  func() uint32 { *events = append(*events, "cache"); return 7 },
		lookupType:     func(string) uint32 { *events = append(*events, "lookup"); return 7 },
		storeTypeCache: func(uint32) { *events = append(*events, "store") },
		applyForce: func(_ *Object, _ types.Pointf, force float64) {
			*events = append(*events, fmt.Sprintf("force:%g", force))
		},
		clearOwner:  func(*Object) { *events = append(*events, "clear-owner") },
		ballAudio:   func(*Object) { *events = append(*events, "ball-audio") },
		effectAudio: func(int32, *Object) { *events = append(*events, "effect-audio") },
	}
}

func TestFumbleCastTargetBranches52C060(t *testing.T) {
	for _, class := range []uint32{0, 2, 4, 6, 0x80000000, 0x80000002, 0x80000004, 0xffffffff} {
		for _, sub := range []uint32{0, 0x10, 0x2000, 0x2010, 0x80000000, 0x80002010} {
			t.Run(fmt.Sprintf("%08x/%08x", class, sub), func(t *testing.T) {
				target := &Object{ObjClass: object.Class(class), ObjSubClass: object.SubClass(sub), ObjFlags: object.Flags(0xffffffff)}
				var events []string
				got := fumbleCast52C060(99, &Object{}, &SpellAcceptArg{Obj: target}, fumbleTestDeps52C060(&events))
				want := []string{"effect-audio"}
				if class&4 != 0 || class&2 != 0 && sub&0x10 != 0 {
					want = []string{"cache", "effect-audio"}
				} else if class&2 == 0 || sub&0x2000 == 0 {
					want = []string{"drop-all", "force:50", "effect-audio"}
				}
				if got != 1 || !reflect.DeepEqual(events, want) {
					t.Fatalf("result/events=%d/%v, want 1/%v", got, events, want)
				}
			})
		}
	}
}

func TestFumbleCastEquippedItemPredicate52C060(t *testing.T) {
	for _, flags := range []uint32{0, 0x100, 0x8000, 0x8100, 0xffffffff} {
		for _, class := range []uint32{0, 0x1000, 0x1000000, 0x2000000, 0x4000000, 0x2001000, 0xffffffff} {
			for _, sub := range []uint32{0, 2, 0x102, 0x2000} {
				t.Run(fmt.Sprintf("%08x/%08x/%08x", flags, class, sub), func(t *testing.T) {
					item := &Object{ObjClass: object.Class(class), ObjSubClass: object.SubClass(sub), ObjFlags: object.Flags(flags)}
					target := &Object{ObjClass: object.ClassPlayer, InvFirstItem: item}
					var events []string
					h := fumbleTestDeps52C060(&events)
					drops := 0
					h.forceDrop = func(owner, gotItem *Object) int32 {
						drops++
						if owner != target || gotItem != item {
							t.Fatal("wrong equipment/owner")
						}
						return math.MinInt32 // The original ignores drop success/failure.
					}
					got := fumbleCast52C060(99, nil, &SpellAcceptArg{Obj: target}, h)
					wantDrops := 0
					if flags&0x100 != 0 && (class&0x1001000 != 0 || class&0x2000000 != 0 && sub&2 != 0) {
						wantDrops = 1
					}
					if got != 1 || drops != wantDrops || !reflect.DeepEqual(events, []string{"cache", "effect-audio"}) {
						t.Fatalf("result/drops/events=%d/%d/%v, want drops %d", got, drops, events, wantDrops)
					}
				})
			}
		}
	}
}

func TestFumbleCastNativeInventorySavedNextAndLiveTarget52C060(t *testing.T) {
	objects := make([]*Object, 5)
	for i := range objects {
		obj, free := alloc.New(Object{})
		t.Cleanup(free)
		objects[i] = obj
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 {
			t.Fatalf("Fumble object=%p, want >4 GiB", obj)
		}
	}
	target, other, first, second, skipped := objects[0], objects[1], objects[2], objects[3], objects[4]
	*target = Object{ObjClass: object.ClassPlayer, InvFirstItem: first}
	*first = Object{ObjClass: object.Class(0x1000), ObjFlags: object.Flags(0x100), InvNextItem: second}
	*second = Object{ObjClass: object.Class(0x1000000), ObjFlags: object.Flags(0x100), InvNextItem: skipped}
	*skipped = Object{ObjClass: object.Class(0x1000)} // Unequipped: no drop.
	arg, freeArg := alloc.New(SpellAcceptArg{})
	t.Cleanup(freeArg)
	*arg = SpellAcceptArg{Obj: target, Pos: types.Ptf(-3, 4)}
	var events []string
	h := fumbleTestDeps52C060(&events)
	drops := 0
	h.forceDrop = func(owner, item *Object) int32 {
		drops++
		if drops == 1 {
			if owner != target || item != first {
				t.Fatal("first drop lost native pointer")
			}
			first.InvNextItem = nil
			target.InvFirstItem = skipped
			arg.Obj = other
		} else if drops == 2 {
			if owner != other || item != second {
				t.Fatal("second drop did not use saved next and live target")
			}
			second.InvNextItem = first
		} else {
			t.Fatal("followed mutated inventory links")
		}
		return 0
	}
	h.loadTypeCache = func() uint32 {
		if drops != 2 {
			t.Fatalf("cache loaded before inventory finished: %d", drops)
		}
		return 7
	}
	h.effectAudio = func(id int32, obj *Object) {
		if id != 99 || obj != other {
			t.Fatal("final audio did not reload target")
		}
	}
	if got := fumbleCast52C060(99, nil, arg, h); got != 1 || drops != 2 || arg.Pos != types.Ptf(-3, 4) || len(events) != 0 {
		t.Fatalf("result/drops/arg/events=%d/%d/%+v/%v", got, drops, *arg, events)
	}
}

func TestFumbleCastLiveOwnedTargetAndBallOrder52C060(t *testing.T) {
	marker := &Object{}
	secondBall := &Object{TypeInd: 7}
	ball := &Object{TypeInd: 7, ObjFlags: object.Flags(0xffffffff), Obj130: marker, Field128: secondBall}
	wrong := &Object{TypeInd: 8, Field128: ball}
	target := &Object{ObjClass: object.ClassPlayer}
	ownedTarget := &Object{Field129: wrong, PosVec: types.Ptf(1, 2)}
	audioTarget := &Object{}
	finalTarget := &Object{}
	arg := &SpellAcceptArg{Obj: target}
	var events []string
	h := fumbleTestDeps52C060(&events)
	h.loadTypeCache = func() uint32 { events = append(events, "cache"); return 0 }
	h.lookupType = func(name string) uint32 {
		events = append(events, "lookup:"+name)
		arg.Obj = ownedTarget
		return 7
	}
	h.storeTypeCache = func(value uint32) {
		events = append(events, "store")
		if value != 7 {
			t.Fatal("wrong type store")
		}
		ownedTarget.PosVec = types.Ptf(123.5, -456.25)
	}
	h.applyForce = func(obj *Object, pos types.Pointf, force float64) {
		events = append(events, "force")
		if obj != ball || pos != ownedTarget.PosVec || force != 100 || obj.Obj130 != marker || obj.ObjFlags != object.Flags(0xffffffff) {
			t.Fatal("ball force used wrong target/position or changed unrelated fields")
		}
		ball.Field128 = nil // Match ends the scan; no next-owned dereference.
	}
	h.clearOwner = func(obj *Object) {
		events = append(events, "clear-owner")
		if obj != ball {
			t.Fatal("wrong cleared ball")
		}
		arg.Obj = audioTarget
	}
	h.ballAudio = func(obj *Object) {
		events = append(events, "ball-audio")
		if obj != audioTarget {
			t.Fatal("ball sound used stale target")
		}
		arg.Obj = finalTarget
	}
	h.effectAudio = func(id int32, obj *Object) {
		events = append(events, "effect-audio")
		if id != 99 || obj != finalTarget {
			t.Fatal("effect sound used stale target")
		}
	}
	want := []string{"cache", "lookup:GameBall", "store", "force", "clear-owner", "ball-audio", "effect-audio"}
	if got := fumbleCast52C060(99, nil, arg, h); got != 1 || !reflect.DeepEqual(events, want) || ball.Obj130 != marker || ball.ObjFlags != object.Flags(0xffffffff) {
		t.Fatalf("result/events/ball=%d/%v/%+v", got, events, ball)
	}
}

func TestFumbleCastWholeTypeCacheAndZeroRetry52C060(t *testing.T) {
	for _, tc := range []struct {
		cache, lookup uint32
		kind          uint16
		match, miss   bool
	}{
		{7, 0, 7, true, false},
		{0x10007, 0, 7, false, false},
		{0xffffffff, 0, 0xffff, false, false},
		{0, 0xffff, 0xffff, true, true},
		{0, 0, 0, true, true},
		{0, 0x10007, 7, false, true},
	} {
		t.Run(fmt.Sprintf("%08x/%08x/%04x", tc.cache, tc.lookup, tc.kind), func(t *testing.T) {
			cache := tc.cache
			ball := &Object{TypeInd: tc.kind}
			target := &Object{ObjClass: object.ClassPlayer, Field129: ball}
			var events []string
			h := fumbleTestDeps52C060(&events)
			lookups, stores := 0, 0
			h.loadTypeCache = func() uint32 { return cache }
			h.lookupType = func(name string) uint32 {
				lookups++
				if name != "GameBall" {
					t.Fatal(name)
				}
				return tc.lookup
			}
			h.storeTypeCache = func(value uint32) { stores++; cache = value }
			for i := 0; i < 2; i++ {
				events = nil
				if got := fumbleCast52C060(99, nil, &SpellAcceptArg{Obj: target}, h); got != 1 {
					t.Fatal(got)
				}
				want := []string{"effect-audio"}
				if tc.match {
					want = []string{"force:100", "clear-owner", "ball-audio", "effect-audio"}
				}
				if !reflect.DeepEqual(events, want) {
					t.Fatalf("events=%v, want %v", events, want)
				}
			}
			wantLookups := 0
			if tc.miss {
				wantLookups = 1
				if tc.lookup == 0 {
					wantLookups = 2
				}
			}
			if lookups != wantLookups || stores != wantLookups {
				t.Fatalf("lookup/store=%d/%d, want %d", lookups, stores, wantLookups)
			}
		})
	}
}

func TestFumbleCastOrdinaryTargetReloadAfterDrop52C060(t *testing.T) {
	target := &Object{ObjClass: object.ClassMonster}
	liveTarget, finalTarget := &Object{}, &Object{}
	caster := &Object{PosVec: types.Ptf(1, 2)}
	arg := &SpellAcceptArg{Obj: target, Pos: types.Ptf(-1000, -2000)}
	var events []string
	h := fumbleTestDeps52C060(&events)
	h.dropAll = func(obj *Object) {
		events = append(events, "drop-all")
		if obj != target {
			t.Fatal("drop-all lost entry target")
		}
		arg.Obj = liveTarget
		caster.PosVec = types.Ptf(123.5, -456.25)
	}
	h.applyForce = func(obj *Object, pos types.Pointf, force float64) {
		events = append(events, "force")
		if obj != liveTarget || pos != caster.PosVec || force != 50 {
			t.Fatal("ordinary push used stale/wrong arguments")
		}
		arg.Obj = finalTarget
	}
	h.effectAudio = func(id int32, obj *Object) {
		events = append(events, "effect-audio")
		if id != 99 || obj != finalTarget {
			t.Fatal("ordinary effect audio used stale target")
		}
	}
	if got := fumbleCast52C060(99, caster, arg, h); got != 1 || !reflect.DeepEqual(events, []string{"drop-all", "force", "effect-audio"}) || arg.Pos != types.Ptf(-1000, -2000) {
		t.Fatalf("result/events/arg=%d/%v/%+v", got, events, *arg)
	}
}

func TestFumbleCastNilArgumentsAndOriginalFaultPrefix52C060(t *testing.T) {
	if got := fumbleCast52C060(99, nil, &SpellAcceptArg{}, fumbleCastDeps52C060{}); got != 0 {
		t.Fatal(got)
	}
	for _, tc := range []struct {
		name string
		arg  *SpellAcceptArg
		want []string
	}{
		{"nil-arg", nil, nil},
		{"nil-caster", &SpellAcceptArg{Obj: &Object{ObjClass: object.ClassMonster}}, []string{"drop-all"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			defer func() {
				if recover() == nil || !reflect.DeepEqual(events, tc.want) {
					t.Fatalf("missing original fault/events=%v, want %v", events, tc.want)
				}
			}()
			fumbleCast52C060(99, nil, tc.arg, fumbleTestDeps52C060(&events))
		})
	}
	// The two branches that never use the caster must allow a nil caster.
	for _, target := range []*Object{
		{ObjClass: object.ClassPlayer},
		{ObjClass: object.ClassMonster, ObjSubClass: object.SubClass(0x2000)},
	} {
		var events []string
		if got := fumbleCast52C060(99, nil, &SpellAcceptArg{Obj: target}, fumbleTestDeps52C060(&events)); got != 1 {
			t.Fatal(got)
		}
	}
}

func TestFumbleCastServiceFaultPrefixes52C060(t *testing.T) {
	for _, tc := range []struct {
		name       string
		class, sub uint32
		want       []string
	}{
		{"player-ball", 4, 0, []string{"drop", "cache", "lookup", "store", "force", "clear-owner", "ball-audio", "effect-audio"}},
		{"ordinary-monster", 2, 0, []string{"drop-all", "force", "effect-audio"}},
		{"shopkeeper", 2, 0x2000, []string{"effect-audio"}},
	} {
		for stop := 1; stop <= len(tc.want); stop++ {
			t.Run(fmt.Sprintf("%s/fault-%d", tc.name, stop), func(t *testing.T) {
				item := &Object{ObjClass: object.Class(0x1000), ObjFlags: object.Flags(0x100)}
				ball := &Object{TypeInd: 7}
				target := &Object{ObjClass: object.Class(tc.class), ObjSubClass: object.SubClass(tc.sub), InvFirstItem: item, Field129: ball}
				var events []string
				event := func(name string) {
					events = append(events, name)
					if len(events) == stop {
						panic("fixture service fault")
					}
				}
				h := fumbleCastDeps52C060{
					forceDrop:      func(*Object, *Object) int32 { event("drop"); return 0 },
					dropAll:        func(*Object) { event("drop-all") },
					loadTypeCache:  func() uint32 { event("cache"); return 0 },
					lookupType:     func(string) uint32 { event("lookup"); return 7 },
					storeTypeCache: func(uint32) { event("store") },
					applyForce:     func(*Object, types.Pointf, float64) { event("force") },
					clearOwner:     func(*Object) { event("clear-owner") },
					ballAudio:      func(*Object) { event("ball-audio") },
					effectAudio:    func(int32, *Object) { event("effect-audio") },
				}
				defer func() {
					if got := recover(); got != "fixture service fault" || !reflect.DeepEqual(events, tc.want[:stop]) {
						t.Fatalf("fault/events=%v/%v, want prefix %v", got, events, tc.want[:stop])
					}
				}()
				fumbleCast52C060(99, &Object{}, &SpellAcceptArg{Obj: target}, h)
			})
		}
	}
}

func TestFumbleCastServerNativeCacheOwnerAndEffectAudio52C060(t *testing.T) {
	// Standalone server tests do not load the legacy C data blobs. Extend only
	// the test view when another standalone fixture registered a smaller blob.
	need := int(fumbleCastCacheOffset52C060) + 4
	if blob := memmap.BlobByAddr(fumbleCastCacheBase52C060); blob == nil {
		memmap.RegisterBlobData(fumbleCastCacheBase52C060, "fumble_cast_test", make([]byte, need))
	} else if len(blob.Data) < need {
		previous := *blob
		data := make([]byte, need)
		copy(data, blob.Data)
		blob.Data, blob.Size = data, uintptr(len(data))
		t.Cleanup(func() { *blob = previous })
	}
	cache := memmap.PtrUint32(fumbleCastCacheBase52C060, fumbleCastCacheOffset52C060)
	previous := *cache
	t.Cleanup(func() { *cache = previous })
	*cache = 0
	s := creatureMonitoredTestServer500CC0(t)
	s.Types.byID = map[string]*ObjectType{"gameball": {ind: 0x2468}}
	s.Types.fast.ball, s.Types.fast.winkGameBall4F7DF0 = 11, 22
	s.Spells.byID = map[spell.ID]*SpellDef{spell.SPELL_FUMBLE: {CastSound: sound.SoundFumbleCast, OnSound: sound.SoundFumbleEffect}}
	objects := make([]*Object, 5)
	for i := range objects {
		obj, free := alloc.New(Object{})
		t.Cleanup(free)
		objects[i] = obj
	}
	target, wrong, ball, secondBall, marker := objects[0], objects[1], objects[2], objects[3], objects[4]
	*target = Object{ObjClass: object.ClassPlayer, PosVec: types.Ptf(5, 7), Field129: wrong}
	*wrong = Object{TypeInd: 9, Field128: ball}
	*ball = Object{TypeInd: 0x2468, ObjOwner: target, Field128: secondBall, ObjFlags: object.Flags(0xffffffff), Obj130: marker, serverHandle: s.handle}
	*secondBall = Object{TypeInd: 0x2468, ObjOwner: target}
	arg, freeArg := alloc.New(SpellAcceptArg{})
	t.Cleanup(freeArg)
	*arg = SpellAcceptArg{Obj: target, Pos: types.Ptf(-1000, -2000)}
	before := *arg
	forces := 0
	got := s.CastFumble52C060(int32(spell.SPELL_FUMBLE), nil, marker, nil, arg, math.MinInt32, FumbleCastRuntime52C060{
		ForceDrop: func(*Object, *Object) int32 { t.Fatal("empty inventory caused force-drop"); return 0 },
		DropAll:   func(*Object) { t.Fatal("player caused drop-all") },
		ApplyForce: func(obj *Object, pos types.Pointf, force float64) {
			forces++
			if obj != ball || pos != target.PosVec || force != 100 || len(s.Audio.delayedObj) != 0 || ball.ObjOwner != target {
				t.Fatal("wrong force/owner/audio order")
			}
		},
	})
	if got != 1 || forces != 1 || *arg != before || *cache != 0x2468 || s.Types.fast.ball != 11 || s.Types.fast.winkGameBall4F7DF0 != 22 {
		t.Fatalf("result/forces/arg/cache=%d/%d/%+v/%#x", got, forces, *arg, *cache)
	}
	if target.Field129 != wrong || wrong.Field128 != secondBall || ball.Field128 != secondBall || ball.ObjOwner != nil || secondBall.ObjOwner != target || ball.Obj130 != marker || ball.ObjFlags != object.Flags(0xffffffff) {
		t.Fatal("wrong native owned-list release or unrelated ball mutation")
	}
	if len(s.Audio.delayedObj) != 2 {
		t.Fatalf("audio=%v", s.Audio.delayedObj)
	}
	for i, id := range []sound.ID{926, sound.SoundFumbleEffect} {
		if event := s.Audio.delayedObj[i]; event.ID != id || event.Obj != target || event.Kind != 0 || event.Code != 0 {
			t.Fatalf("audio[%d]=%+v", i, event)
		}
	}
}

func TestFumbleCastNativeObjectAndArgumentLayouts52C060(t *testing.T) {
	want := []uintptr{780, 4, 8, 12, 16, 56, 496, 504, 512, 516, 12, 0, 4}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		want = []uintptr{928, 8, 12, 16, 20, 60, 528, 544, 560, 568, 16, 0, 8}
	}
	got := []uintptr{
		unsafe.Sizeof(Object{}), unsafe.Offsetof(Object{}.TypeInd), unsafe.Offsetof(Object{}.ObjClass), unsafe.Offsetof(Object{}.ObjSubClass),
		unsafe.Offsetof(Object{}.ObjFlags), unsafe.Offsetof(Object{}.PosVec), unsafe.Offsetof(Object{}.InvNextItem), unsafe.Offsetof(Object{}.InvFirstItem),
		unsafe.Offsetof(Object{}.Field128), unsafe.Offsetof(Object{}.Field129), unsafe.Sizeof(SpellAcceptArg{}), unsafe.Offsetof(SpellAcceptArg{}.Obj), unsafe.Offsetof(SpellAcceptArg{}.Pos),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("native layout=%v, want %v", got, want)
	}
}
