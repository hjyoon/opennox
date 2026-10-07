package server

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func triggerGlyphAlloc52CCD0[T comparable](t *testing.T, value T) *T {
	t.Helper()
	ptr, free := alloc.New(value)
	t.Cleanup(free)
	*ptr = value
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(ptr)) <= math.MaxUint32 {
		t.Fatalf("TriggerGlyph native allocation=%p, want above 4 GiB", ptr)
	}
	return ptr
}

func triggerGlyphTestHooks52CCD0(first *Object, die func(*Object)) triggerGlyphCastHooks52CCD0 {
	return triggerGlyphCastHooks52CCD0{
		first:     func() *Object { return first },
		next:      (*Object).Next,
		hasParent: (*Object).HasOwner,
		typeName:  func(*Object) string { return "Glyph" },
		castSound: func(int32) sound.ID { return 321 },
		audio:     func(sound.ID, *Object, int, uint32) {},
		dieGlyph:  die,
	}
}

func TestTriggerGlyph52CCD0SelectionAndOriginalBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name      string
		positions []types.Pointf
		want      int
	}{
		{"empty", nil, -1},
		{"coincident", []types.Pointf{{}}, 0},
		{"cutoff_equal", []types.Pointf{types.Ptf(10000, 0)}, -1},
		{"cutoff_above", []types.Pointf{types.Ptf(math.Nextafter32(10000, float32(math.Inf(1))), 0)}, -1},
		{"cutoff_below", []types.Pointf{types.Ptf(math.Nextafter32(10000, 0), 0)}, 0},
		{"closer_later", []types.Pointf{types.Ptf(3, 4), types.Ptf(1, 2)}, 1},
		{"tie_first", []types.Pointf{types.Ptf(3, 4), types.Ptf(-3, -4)}, 0},
		{"spill_tie_retains_truly_farther_first", []types.Pointf{types.Ptf(3, math.Float32frombits(0x3f800001)), types.Ptf(3, 1)}, 0},
		{"infinity_does_not_win", []types.Pointf{types.Ptf(float32(math.Inf(1)), 0)}, -1},
		{"unordered_wins", []types.Pointf{types.Ptf(float32(math.NaN()), 0)}, 0},
		{"unordered_cached_distance_admits_outside_cutoff", []types.Pointf{types.Ptf(float32(math.NaN()), 0), types.Ptf(20000, 0)}, 1},
		{"unordered_later_replaces", []types.Pointf{types.Ptf(1, 0), types.Ptf(0, float32(math.NaN()))}, 1},
		{"negative_zero", []types.Pointf{types.Ptf(math.Float32frombits(0x80000000), 0)}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			caster := triggerGlyphAlloc52CCD0(t, Object{})
			var first, prior *Object
			objects := make([]*Object, len(tc.positions))
			for i, position := range tc.positions {
				objects[i] = triggerGlyphAlloc52CCD0(t, Object{ObjOwner: caster, PosVec: position,
					ObjFlags: object.FlagDead | object.FlagDestroyed | object.FlagNoUpdate, ObjClass: 0})
				if prior == nil {
					first = objects[i]
				} else {
					prior.ObjNext = objects[i]
				}
				prior = objects[i]
			}
			calls, audio := 0, 0
			h := triggerGlyphTestHooks52CCD0(first, func(got *Object) {
				calls++
				if tc.want < 0 || got != objects[tc.want] {
					t.Fatalf("selected=%p, want index=%d", got, tc.want)
				}
			})
			h.audio = func(id sound.ID, unit *Object, kind int, code uint32) {
				audio++
				if id != 321 || unit != caster || kind != 0 || code != 0 || calls != 0 {
					t.Fatalf("audio before death=%d/%p/%d/%d death=%d", id, unit, kind, code, calls)
				}
			}
			wantResult := int32(0)
			if tc.want >= 0 {
				wantResult = 1
			}
			if got := triggerGlyphCast52CCD0(math.MinInt32, caster, h); got != wantResult ||
				calls != int(wantResult) || audio != int(wantResult) {
				t.Fatalf("result/death/audio=%d/%d/%d, want %d each", got, calls, audio, wantResult)
			}
		})
	}
}

func TestTriggerGlyph52CCD0TypeNames(t *testing.T) {
	for _, name := range []string{"Glyph", "Glyph\x00suffix", "glyph", "GLYPH", "Glyph2", "Glyph\x00", "", "Glyp", "Glyph "} {
		t.Run(fmt.Sprintf("%q", name), func(t *testing.T) {
			caster := triggerGlyphAlloc52CCD0(t, Object{})
			glyph := triggerGlyphAlloc52CCD0(t, Object{ObjOwner: caster})
			calls := 0
			h := triggerGlyphTestHooks52CCD0(glyph, func(*Object) { calls++ })
			h.typeName = func(*Object) string { return name }
			want := int32(0)
			if name == "Glyph" || name == "Glyph\x00" || name == "Glyph\x00suffix" {
				want = 1
			}
			if got := triggerGlyphCast52CCD0(13, caster, h); got != want || calls != int(want) {
				t.Fatalf("C-name result/death=%d/%d want=%d", got, calls, want)
			}
		})
	}
}

func TestTriggerGlyph52CCD0LiveOrderAndCachedWinner(t *testing.T) {
	caster := triggerGlyphAlloc52CCD0(t, Object{})
	bridge := triggerGlyphAlloc52CCD0(t, Object{ObjOwner: caster})
	last := triggerGlyphAlloc52CCD0(t, Object{ObjOwner: bridge, PosVec: types.Ptf(3, 4)})
	first := triggerGlyphAlloc52CCD0(t, Object{ObjOwner: caster, PosVec: types.Ptf(1, 0)})
	unrelated := triggerGlyphAlloc52CCD0(t, Object{ObjNext: first})
	var events []string
	names := map[*Object]string{unrelated: "unrelated", first: "first", last: "last"}
	h := triggerGlyphTestHooks52CCD0(unrelated, func(obj *Object) {
		events = append(events, "die:"+names[obj])
		if obj != last {
			t.Fatalf("cached winner=%p want=%p", obj, last)
		}
	})
	h.first = func() *Object { events = append(events, "first-world"); return unrelated }
	h.hasParent = func(obj, owner *Object) bool {
		events = append(events, "parent:"+names[obj])
		if owner != caster {
			t.Fatalf("owner=%p want=%p", owner, caster)
		}
		return obj.HasOwner(owner)
	}
	h.typeName = func(obj *Object) string {
		events = append(events, "name:"+names[obj])
		if obj == unrelated {
			t.Fatal("unrelated type was read before ownership rejection")
		}
		if obj == first {
			// Later traversal follows the link modified by the name service,
			// and distances use this live caster position, not an entry copy.
			first.ObjNext = last
			caster.PosVec = types.Ptf(3, 4)
		}
		return "Glyph"
	}
	h.next = func(obj *Object) *Object { events = append(events, "next:"+names[obj]); return obj.ObjNext }
	h.castSound = func(id int32) sound.ID {
		events = append(events, "sound")
		if id != math.MinInt32 {
			t.Fatalf("signed ID=%d", id)
		}
		last.PosVec, last.ObjOwner = types.Ptf(10000, 10000), nil
		return 456
	}
	h.audio = func(id sound.ID, unit *Object, kind int, code uint32) {
		events = append(events, "audio")
		if id != 456 || unit != caster || kind != 0 || code != 0 {
			t.Fatalf("audio=%d/%p/%d/%d", id, unit, kind, code)
		}
	}
	if got := triggerGlyphCast52CCD0(math.MinInt32, caster, h); got != 1 {
		t.Fatalf("result=%d", got)
	}
	want := []string{"first-world", "parent:unrelated", "next:unrelated", "parent:first", "name:first", "next:first",
		"parent:last", "name:last", "next:last", "sound", "audio", "die:last"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("call order=%v, want=%v", events, want)
	}
}

func TestTriggerGlyph52CCD0SelfAndReentrantSelection(t *testing.T) {
	outer := triggerGlyphAlloc52CCD0(t, Object{})
	inner := triggerGlyphAlloc52CCD0(t, Object{})
	var deaths []*Object
	h := triggerGlyphTestHooks52CCD0(outer, func(obj *Object) { deaths = append(deaths, obj) })
	h.audio = func(sound.ID, *Object, int, uint32) {
		inside := triggerGlyphTestHooks52CCD0(inner, func(obj *Object) { deaths = append(deaths, obj) })
		if got := triggerGlyphCast52CCD0(123, inner, inside); got != 1 {
			t.Fatalf("reentrant result=%d", got)
		}
	}
	if got := triggerGlyphCast52CCD0(321, outer, h); got != 1 || !reflect.DeepEqual(deaths, []*Object{inner, outer}) {
		t.Fatalf("self/reentrant result/deaths=%d/%v", got, deaths)
	}
}

func TestTriggerGlyph52CCD0OriginalFaultPrefixes(t *testing.T) {
	for _, stage := range []string{"first", "parent", "name", "next", "sound", "audio", "die"} {
		t.Run(stage, func(t *testing.T) {
			caster := triggerGlyphAlloc52CCD0(t, Object{})
			var events []string
			fault := func(name string) {
				events = append(events, name)
				if name == stage {
					panic(name)
				}
			}
			h := triggerGlyphCastHooks52CCD0{
				first:     func() *Object { fault("first"); return caster },
				hasParent: func(*Object, *Object) bool { fault("parent"); return true },
				typeName:  func(*Object) string { fault("name"); return "Glyph" },
				next:      func(*Object) *Object { fault("next"); return nil },
				castSound: func(int32) sound.ID { fault("sound"); return 321 },
				audio:     func(sound.ID, *Object, int, uint32) { fault("audio") },
				dieGlyph:  func(*Object) { fault("die") },
			}
			defer func() {
				if got := recover(); got != stage {
					t.Fatalf("fault=%v want=%s", got, stage)
				}
				all := []string{"first", "parent", "name", "next", "sound", "audio", "die"}
				for i, name := range all {
					if name == stage && !reflect.DeepEqual(events, all[:i+1]) {
						t.Fatalf("fault prefix=%v want=%v", events, all[:i+1])
					}
				}
			}()
			triggerGlyphCast52CCD0(123, caster, h)
		})
	}
	for _, hasObject := range []bool{false, true} {
		t.Run(fmt.Sprintf("nil_caster_world_%v", hasObject), func(t *testing.T) {
			var first *Object
			if hasObject {
				first = triggerGlyphAlloc52CCD0(t, Object{})
			}
			h := triggerGlyphTestHooks52CCD0(first, func(*Object) { t.Fatal("nil caster triggered a glyph") })
			h.typeName = func(*Object) string { t.Fatal("nil caster performed type lookup"); return "Glyph" }
			if got := triggerGlyphCast52CCD0(123, nil, h); got != 0 {
				t.Fatalf("nil caster=%d", got)
			}
		})
	}
}

func triggerGlyphReference53_52CCD0(a, b types.Pointf) float64 {
	value := func(x float32) *big.Float {
		return new(big.Float).SetPrec(53).SetMode(big.ToZero).SetFloat64(float64(x))
	}
	chop := func() *big.Float { return new(big.Float).SetPrec(53).SetMode(big.ToZero) }
	dx, dy := chop().Sub(value(a.X), value(b.X)), chop().Sub(value(a.Y), value(b.Y))
	xSquare, ySquare := chop().Mul(dx, dx), chop().Mul(dy, dy)
	square, _ := chop().Add(ySquare, xSquare).Float64()
	return square
}

func TestTriggerGlyphDistance52CCD0IndependentChop53Reference(t *testing.T) {
	words := []uint32{0, 0x80000000, 0x3f800001, 0x40400000, 0x461c4000, 0xc61c4001,
		0x007fffff, 0x00800000, 0x1effffff, 0x5e000001, 0x7f7fffff, 0xff7fffff}
	for i, first := range words {
		for j, second := range words {
			t.Run(fmt.Sprintf("%08x_%08x", first, second), func(t *testing.T) {
				a := types.Ptf(math.Float32frombits(first), math.Float32frombits(words[(i+j+1)%len(words)]))
				b := types.Ptf(math.Float32frombits(second), math.Float32frombits(words[(2*i+j+3)%len(words)]))
				want := triggerGlyphReference53_52CCD0(a, b)
				if got := triggerGlyphDistance52CCD0(a, b); math.Float64bits(got) != math.Float64bits(want) {
					t.Fatalf("chop53 distance=%016x/%g want=%016x/%g", math.Float64bits(got), got, math.Float64bits(want), want)
				}
			})
		}
	}
}

func TestTriggerGlyphServer52CCD0UsesNativeWorldTypeAudioAndDeath(t *testing.T) {
	caster := triggerGlyphAlloc52CCD0(t, Object{TypeInd: 2})
	owned := triggerGlyphAlloc52CCD0(t, Object{ObjOwner: caster})
	further := triggerGlyphAlloc52CCD0(t, Object{ObjOwner: caster, TypeInd: 1, PosVec: types.Ptf(100, 0)})
	nearer := triggerGlyphAlloc52CCD0(t, Object{ObjOwner: owned, TypeInd: 1, PosVec: types.Ptf(3, 4), ObjNext: further,
		ObjFlags: object.FlagDestroyed | object.FlagDead})
	other := triggerGlyphAlloc52CCD0(t, Object{TypeInd: 1, ObjNext: nearer})
	s := new(Server)
	s.Objs.SetObjects(other)
	s.Types.byInd = []*ObjectType{nil, {id: "Glyph"}, {id: "NPC"}}
	s.Spells.byID = map[spell.ID]*SpellDef{spell.SPELL_TRIGGER_GLYPH: {CastSound: 321}}
	s.Audio.Init(s)
	t.Cleanup(s.Audio.Free)
	s.Audio.bySound[321].Field12 = 1
	s.Audio.Reset()
	var events []string
	s.Audio.OnSound(func(id sound.ID, kind int, obj *Object, pos types.Pointf) {
		events = append(events, "audio")
		if id != 321 || kind != 0 || obj != caster || pos != caster.PosVec {
			t.Fatal("wrong actual audio event")
		}
	})
	ignored := triggerGlyphAlloc52CCD0(t, SpellAcceptArg{Obj: further, Pos: types.Ptf(5000, 5000)})
	wantArg := *ignored
	got := s.CastTriggerGlyph52CCD0(int32(spell.SPELL_TRIGGER_GLYPH), further, caster, further, ignored, math.MinInt32,
		TriggerGlyphCastRuntime52CCD0{DieGlyph: func(obj *Object) {
			events = append(events, "die")
			if obj != nearer {
				t.Fatalf("native death=%p want=%p", obj, nearer)
			}
		}})
	if got != 1 || *ignored != wantArg || !reflect.DeepEqual(events, []string{"audio", "die"}) {
		t.Fatalf("actual binding result/events=%d/%v", got, events)
	}
	eventCount := 0
	s.Audio.EachEvent(func(ev *AudioEvent) {
		eventCount++
		if ev.Sound != 321 || ev.Obj != caster || ev.Kind != 0 || ev.Code != 0 {
			t.Fatal("wrong actual queued audio")
		}
	})
	if eventCount != 1 {
		t.Fatalf("actual audio count=%d", eventCount)
	}
}
