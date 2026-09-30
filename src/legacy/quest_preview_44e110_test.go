package legacy

import (
	"fmt"
	"reflect"
	"testing"
)

type questPreviewFixture44E110 struct {
	font, fontResult int
	cache            [12]int
	trace            []string
	fail             int
	nilSlot          int
	flags            map[int]uint32
}

func newQuestPreviewFixture44E110() *questPreviewFixture44E110 {
	return &questPreviewFixture44E110{fontResult: 99, fail: -1, nilSlot: -1, flags: make(map[int]uint32)}
}

func (f *questPreviewFixture44E110) step(event string) {
	f.trace = append(f.trace, event)
	if len(f.trace)-1 == f.fail {
		panic("fixture fault")
	}
}

func (f *questPreviewFixture44E110) hooks() questPreviewHooks44E110[int, int] {
	return questPreviewHooks44E110[int, int]{
		loadFont: func() int { f.step("font-load"); return f.font },
		fontByName: func(s string) int {
			f.step("font:" + s)
			return f.fontResult
		},
		storeFont: func(v int) { f.step(fmt.Sprintf("font-store:%d", v)); f.font = v },
		load:      func(slot int) int { f.step(fmt.Sprintf("load:%d", slot)); return f.cache[slot] },
		thingByName: func(s string) int32 {
			f.step("thing:" + s)
			for _, spec := range questPreviewSpecs44E110 {
				if spec.name == s {
					return int32(spec.slot + 101)
				}
			}
			panic("unknown type")
		},
		create: func(typ int32) int {
			f.step(fmt.Sprintf("create:%d", typ))
			if int(typ)-101 == f.nilSlot {
				return 0
			}
			return int(typ)
		},
		store: func(slot, dr int) { f.step(fmt.Sprintf("store:%d:%d", slot, dr)); f.cache[slot] = dr },
		mark: func(dr int) {
			f.step(fmt.Sprintf("mark:%d", dr))
			if dr == 0 {
				panic("nil drawable")
			}
			f.flags[dr] |= 0x1000000
		},
	}
}

func questPreviewGolden44E110() []string {
	out := []string{"font-load", "font:default", "font-store:99"}
	// Independent original order, including exit-before-generator.
	for _, spec := range []questPreviewSpec44E110{
		{1, "GauntletExitB"}, {0, "BeholderGenerator"}, {2, "Ankh"},
		{3, "SoulGate"}, {4, "SilverKey"}, {5, "GoldKey"}, {6, "QuestGoldChest"},
		{7, "QuestGoldPile"}, {8, "DunMirChest4"}, {9, "WarHammer"},
		{10, "HastePotion"}, {11, "ConjurerSpellBook"},
	} {
		out = append(out, fmt.Sprintf("load:%d", spec.slot), "thing:"+spec.name,
			fmt.Sprintf("create:%d", spec.slot+101), fmt.Sprintf("store:%d:%d", spec.slot, spec.slot+101),
			fmt.Sprintf("mark:%d", spec.slot+101))
	}
	return out
}

func TestQuestPreview44E110OrderReuseAndNilFont(t *testing.T) {
	f := newQuestPreviewFixture44E110()
	if got := questPreview44E110(f.hooks()); got != 112 || f.font != 99 {
		t.Fatalf("return=%d font=%d", got, f.font)
	}
	if want := questPreviewGolden44E110(); !reflect.DeepEqual(f.trace, want) {
		t.Fatalf("trace=%v want=%v", f.trace, want)
	}
	f.trace = nil
	for dr := range f.flags {
		f.flags[dr] = 0x80000003
	}
	if got := questPreview44E110(f.hooks()); got != 112 {
		t.Fatalf("cached return=%d", got)
	}
	want := []string{"font-load"}
	for _, slot := range []int{1, 0, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11} {
		want = append(want, fmt.Sprintf("load:%d", slot), fmt.Sprintf("mark:%d", slot+101))
		if f.flags[slot+101] != 0x81000003 {
			t.Fatalf("cached flags[%d]=%#x", slot, f.flags[slot+101])
		}
	}
	if !reflect.DeepEqual(f.trace, want) {
		t.Fatalf("cached trace=%v want=%v", f.trace, want)
	}
	f = newQuestPreviewFixture44E110()
	f.fontResult = 0
	if got := questPreview44E110(f.hooks()); got != 112 || f.font != 0 {
		t.Fatalf("nil font stopped creation: return=%d font=%d", got, f.font)
	}
}

func TestQuestPreview44E110LiveCachesAndLocalCreatedValue(t *testing.T) {
	f := newQuestPreviewFixture44E110()
	h := f.hooks()
	store, mark := h.store, h.mark
	h.store = func(slot, dr int) {
		store(slot, dr)
		if slot == 1 {
			f.cache[slot] = 888 // Flag the cached local, not this new slot value.
		}
	}
	h.mark = func(dr int) {
		mark(dr)
		if dr == 102 {
			f.cache[0] = 777 // Next slot must be read after this callback.
		}
	}
	if got := questPreview44E110(h); got != 112 || f.flags[102] != 0x1000000 || f.flags[777] != 0x1000000 || f.flags[888] != 0 {
		t.Fatalf("return=%d flags=%v", got, f.flags)
	}
	for _, event := range f.trace {
		if event == "thing:BeholderGenerator" || event == "create:101" {
			t.Fatalf("ignored live next cache: %v", f.trace)
		}
	}
}

func TestQuestPreview44E110EveryFaultPrefix(t *testing.T) {
	want := questPreviewGolden44E110()
	for n := range want {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			f := newQuestPreviewFixture44E110()
			f.fail = n
			var caught any
			func() {
				defer func() { caught = recover() }()
				questPreview44E110(f.hooks())
			}()
			if caught != "fixture fault" || !reflect.DeepEqual(f.trace, want[:n+1]) {
				t.Fatalf("fault=%v trace=%v want=%v", caught, f.trace, want[:n+1])
			}
		})
	}
}

func TestQuestPreview44E110PublishesNilBeforeFlagFault(t *testing.T) {
	for _, slot := range []int{1, 0, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11} {
		t.Run(fmt.Sprint(slot), func(t *testing.T) {
			f := newQuestPreviewFixture44E110()
			f.nilSlot = slot
			var caught any
			func() {
				defer func() { caught = recover() }()
				questPreview44E110(f.hooks())
			}()
			n := len(f.trace)
			if caught != "nil drawable" || f.cache[slot] != 0 || n < 2 ||
				f.trace[n-2] != fmt.Sprintf("store:%d:0", slot) || f.trace[n-1] != "mark:0" {
				t.Fatalf("fault=%v cache=%v trace=%v", caught, f.cache, f.trace)
			}
			for _, spec := range questPreviewSpecs44E110 {
				if spec.slot == slot {
					break
				}
				if f.cache[spec.slot] != spec.slot+101 || f.flags[spec.slot+101] != 0x1000000 {
					t.Fatalf("lost prior effects: cache=%v flags=%v", f.cache, f.flags)
				}
			}
		})
	}
}
