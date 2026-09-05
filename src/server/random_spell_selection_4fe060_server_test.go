package server

import (
	"math"
	"reflect"
	"testing"

	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
)

func TestRandomSpellSelectionNative4FE060PreservesDwordsAndOrder(t *testing.T) {
	var events []string
	sequence := []spell.ID{133, spell.ID(math.MinInt32), spell.ID(math.MaxInt32), 0}
	next := 1
	got := randomSpellSelectionNative4FE060(0x80000000, 0x40000000, randomSpellSelectionNativeDeps4FE060{
		firstValid: func() spell.ID {
			events = append(events, "first")
			return sequence[0]
		},
		flags: func(gotID spell.ID) things.SpellFlags {
			events = append(events, "flags:"+gotID.String())
			if gotID != spell.ID(math.MinInt32) && gotID != spell.ID(math.MaxInt32) {
				t.Fatalf("flags ID = %d, want signed dword boundary", gotID)
			}
			return things.SpellFlags(0xc0000000)
		},
		nextValid: func(gotID spell.ID) spell.ID {
			events = append(events, "next:"+gotID.String())
			if gotID != sequence[next-1] {
				t.Fatalf("next ID = %d, want %d", gotID, sequence[next-1])
			}
			value := sequence[next]
			next++
			return value
		},
		randomInt: func(minimum, maximum int32) int32 {
			events = append(events, "rng")
			if minimum != 0 || maximum != 1 {
				t.Fatalf("RNG bounds = %d..%d, want 0..1", minimum, maximum)
			}
			return 1
		},
	})
	if got != math.MaxInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MaxInt32))
	}
	want := []string{
		"first",
		"next:" + spell.ID(133).String(),
		"flags:" + spell.ID(math.MinInt32).String(),
		"next:" + spell.ID(math.MinInt32).String(),
		"flags:" + spell.ID(math.MaxInt32).String(),
		"next:" + spell.ID(math.MaxInt32).String(),
		"rng",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestRandomSpellServer4FE060UsesRegistryAllowAllAndLogicRNG(t *testing.T) {
	const seed = 2011
	const (
		firstMask  = uint32(0x80000000)
		secondMask = uint32(0x40000000)
	)
	s := new(Server)
	s.Rand.Logic = prand.New(seed)
	s.Spells.AllowAll = true
	s.Spells.byID = map[spell.ID]*SpellDef{
		1: {ID: 1, Valid: true, Def: things.Spell{Flags: things.SpellFlags(firstMask | secondMask)}},
		3: {ID: 3, Valid: true, Def: things.Spell{Flags: things.SpellFlags(firstMask | secondMask)}},
		4: {ID: 4, Valid: true, Def: things.Spell{Flags: things.SpellFlags(secondMask)}},
		5: {ID: 5, Valid: true, Def: things.Spell{Flags: things.SpellFlags(firstMask)}},
		6: {ID: 6, Valid: true, Def: things.Spell{Flags: things.SpellFlags(firstMask | secondMask)}},
	}

	wantRNG := prand.New(seed)
	wantCandidates := [...]int32{3, 4}
	want := wantCandidates[wantRNG.IntClamp(0, len(wantCandidates)-1)]
	got := s.RandomSpell4FE060(firstMask, secondMask)
	if got != want {
		t.Fatalf("result = %d, want %d", got, want)
	}
	if index := s.Rand.Logic.Index(); index != wantRNG.Index() {
		t.Fatalf("logic RNG index = %d, want %d", index, wantRNG.Index())
	}
}

func TestRandomSpellServer4FE060EmptyRegistryDoesNotDereferenceRNG(t *testing.T) {
	s := new(Server)
	if got := s.RandomSpell4FE060(math.MaxUint32, math.MaxUint32); got != 0 {
		t.Fatalf("empty registry result = %d, want 0", got)
	}
}

func TestRandomSpellExcludedExport4FE100UsesCanonicalResult(t *testing.T) {
	for _, test := range []struct {
		id   int32
		want int32
	}{
		{math.MinInt32, 0},
		{1, 1},
		{133, 1},
		{math.MaxInt32, 0},
	} {
		if got := RandomSpellExcluded4FE100(test.id); got != test.want {
			t.Errorf("ID %d = %d, want %d", test.id, got, test.want)
		}
	}
}
