package legacy

import (
	"fmt"
	"reflect"
	"testing"
)

const (
	questJournalDeletePatternEntryA5007E0 = uint64(0x100000101)
	questJournalDeletePatternEntryB5007E0 = uint64(0x200000202)
	questJournalDeletePatternEntryC5007E0 = uint64(0x300000303)
)

type questJournalDeletePatternTestWorld5007E0 struct {
	events    []string
	faultAt   int
	after     map[int]func()
	qualified string
	exact     uint64
	head      uint64
	next      map[uint64]uint64
	names     map[uint64]string
	deleted   []uint64
}

func newQuestJournalDeletePatternTestWorld5007E0(qualified string) *questJournalDeletePatternTestWorld5007E0 {
	return &questJournalDeletePatternTestWorld5007E0{
		after:     make(map[int]func()),
		qualified: qualified,
		next:      make(map[uint64]uint64),
		names:     make(map[uint64]string),
	}
}

func (w *questJournalDeletePatternTestWorld5007E0) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[len(w.events)]; after != nil {
		after()
	}
}

func (w *questJournalDeletePatternTestWorld5007E0) hooks() questJournalDeletePatternHooks5007E0[uint64] {
	return questJournalDeletePatternHooks5007E0[uint64]{
		qualify: func(pattern string) string {
			w.observe(fmt.Sprintf("qualify:%q=%q", pattern, w.qualified))
			return w.qualified
		},
		findExact: func(pattern string) uint64 {
			w.observe(fmt.Sprintf("find-exact:%q=%x", pattern, w.exact))
			return w.exact
		},
		loadHead: func() uint64 {
			w.observe(fmt.Sprintf("load-head=%x", w.head))
			return w.head
		},
		loadNext: func(entry uint64) uint64 {
			next := w.next[entry]
			w.observe(fmt.Sprintf("load-next:%x=%x", entry, next))
			return next
		},
		equalFoldPrefix: func(entry uint64, pattern string, count int) bool {
			matched := questJournalASCIIPrefixEqualFold5007E0(w.names[entry], pattern, count)
			w.observe(fmt.Sprintf("prefix:%x:%q:%d=%t", entry, pattern, count, matched))
			return matched
		},
		firstSubstringRemaining: func(entry uint64, start int, needle string) (int, bool) {
			name := w.names[entry]
			haystack := ""
			if start < len(name) {
				haystack = name[start:]
			}
			remaining, found := questJournalFirstSubstringRemaining5007E0(haystack, needle)
			w.observe(fmt.Sprintf("substring:%x:%d:%q=%d/%t", entry, start, needle, remaining, found))
			return remaining, found
		},
		deleteEntry: func(entry uint64) {
			w.deleted = append(w.deleted, entry)
			w.observe(fmt.Sprintf("delete:%x", entry))
		},
	}
}

func TestQuestJournalDeletePattern5007E0ExactUsesOriginalArgument(t *testing.T) {
	w := newQuestJournalDeletePatternTestWorld5007E0("War01a:Count")
	w.exact = questJournalDeletePatternEntryA5007E0

	questJournalDeletePatternContract5007E0("Count", w.hooks())

	want := []string{
		`qualify:"Count"="War01a:Count"`,
		`find-exact:"Count"=100000101`,
		`delete:100000101`,
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}

	w = newQuestJournalDeletePatternTestWorld5007E0("War01a:Missing")
	questJournalDeletePatternContract5007E0("Missing", w.hooks())
	if len(w.deleted) != 0 {
		t.Fatalf("missing exact lookup deleted %x", w.deleted)
	}
}

func TestQuestJournalDeletePattern5007E0AllUsesQualifiedPatternAndSavedNext(t *testing.T) {
	w := newQuestJournalDeletePatternTestWorld5007E0("*:*")
	w.head = questJournalDeletePatternEntryA5007E0
	w.next[questJournalDeletePatternEntryA5007E0] = questJournalDeletePatternEntryB5007E0
	w.after[4] = func() {
		w.next[questJournalDeletePatternEntryA5007E0] = questJournalDeletePatternEntryC5007E0
		w.head = questJournalDeletePatternEntryC5007E0
	}

	questJournalDeletePatternContract5007E0("qualified-by-hook", w.hooks())

	want := []string{
		`qualify:"qualified-by-hook"="*:*"`,
		`load-head=100000101`,
		`load-next:100000101=200000202`,
		`delete:100000101`,
		`load-next:200000202=0`,
		`delete:200000202`,
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
	if wantDeleted := []uint64{questJournalDeletePatternEntryA5007E0, questJournalDeletePatternEntryB5007E0}; !reflect.DeepEqual(w.deleted, wantDeleted) {
		t.Fatalf("deleted = %x, want %x", w.deleted, wantDeleted)
	}
}

func TestQuestJournalDeletePattern5007E0WildcardBranches(t *testing.T) {
	tests := []struct {
		name      string
		qualified string
		names     []string
		want      []uint64
	}{
		{
			name:      "trailing-star-ascii-prefix",
			qualified: "war01A:*",
			names:     []string{"War01a:Count", "War02a:Count"},
			want:      []uint64{questJournalDeletePatternEntryA5007E0},
		},
		{
			name:      "leading-star-first-substring-only",
			qualified: "*:Count",
			names:     []string{"War01a:Count", "War01a:Count:Count", "War02a:count"},
			want:      []uint64{questJournalDeletePatternEntryA5007E0},
		},
		{
			name:      "interior-star-prefix-and-first-substring",
			qualified: "war01A:*Tail",
			names:     []string{"War01a:UpperTail", "War01a:xTailTail", "War01a:Uppertail"},
			want:      []uint64{questJournalDeletePatternEntryA5007E0},
		},
		{
			name:      "interior-star-does-not-need-to-follow-colon",
			qualified: "Map:ab*cd",
			names:     []string{"mAP:xb*cd", "Map:abZZcd"},
			want:      []uint64{questJournalDeletePatternEntryA5007E0},
		},
	}
	entries := []uint64{
		questJournalDeletePatternEntryA5007E0,
		questJournalDeletePatternEntryB5007E0,
		questJournalDeletePatternEntryC5007E0,
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newQuestJournalDeletePatternTestWorld5007E0(tc.qualified)
			w.head = entries[0]
			for i, name := range tc.names {
				entry := entries[i]
				w.names[entry] = name
				if i+1 < len(tc.names) {
					w.next[entry] = entries[i+1]
				}
			}

			questJournalDeletePatternContract5007E0(tc.qualified, w.hooks())

			if !reflect.DeepEqual(w.deleted, tc.want) {
				t.Fatalf("deleted = %x, want %x; events = %q", w.deleted, tc.want, w.events)
			}
		})
	}
}

func TestQuestJournalDeletePattern5007E0ComparisonPrimitives(t *testing.T) {
	for _, tc := range []struct {
		name        string
		left, right string
		count       int
		want        bool
	}{
		{name: "ascii-fold", left: "War01a:", right: "war01A:", count: 7, want: true},
		{name: "unicode-fold-is-not-used", left: "K", right: "\u212a", count: 1, want: false},
		{name: "short-left", left: "Map", right: "Map:", count: 4, want: false},
		{name: "zero-count", left: "", right: "", count: 0, want: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := questJournalASCIIPrefixEqualFold5007E0(tc.left, tc.right, tc.count); got != tc.want {
				t.Fatalf("match = %t, want %t", got, tc.want)
			}
		})
	}

	for _, tc := range []struct {
		haystack, needle string
		remaining        int
		found            bool
	}{
		{haystack: "TailTail", needle: "Tail", remaining: 8, found: true},
		{haystack: "prefixTail", needle: "Tail", remaining: 4, found: true},
		{haystack: "prefix", needle: "Tail", found: false},
		{haystack: "prefix", needle: "", remaining: 6, found: true},
	} {
		remaining, found := questJournalFirstSubstringRemaining5007E0(tc.haystack, tc.needle)
		if remaining != tc.remaining || found != tc.found {
			t.Fatalf("first substring %q in %q = %d/%t, want %d/%t", tc.needle, tc.haystack, remaining, found, tc.remaining, tc.found)
		}
	}
}

func TestQuestJournalDeletePattern5007E0LoadsNextBeforeMatchAndUsesSavedValue(t *testing.T) {
	w := newQuestJournalDeletePatternTestWorld5007E0("Map:*")
	w.head = questJournalDeletePatternEntryA5007E0
	w.next[questJournalDeletePatternEntryA5007E0] = questJournalDeletePatternEntryB5007E0
	w.names[questJournalDeletePatternEntryA5007E0] = "Map:One"
	w.names[questJournalDeletePatternEntryB5007E0] = "Map:Two"
	w.names[questJournalDeletePatternEntryC5007E0] = "Map:Three"
	w.after[3] = func() {
		w.next[questJournalDeletePatternEntryA5007E0] = questJournalDeletePatternEntryC5007E0
	}

	questJournalDeletePatternContract5007E0("Map:*", w.hooks())

	wantDeleted := []uint64{questJournalDeletePatternEntryA5007E0, questJournalDeletePatternEntryB5007E0}
	if !reflect.DeepEqual(w.deleted, wantDeleted) {
		t.Fatalf("deleted = %x, want saved-next traversal %x", w.deleted, wantDeleted)
	}
	wantPrefix := []string{
		`load-next:100000101=200000202`,
		`prefix:100000101:"Map:*":4=true`,
		`delete:100000101`,
	}
	if got := w.events[2:5]; !reflect.DeepEqual(got, wantPrefix) {
		t.Fatalf("first loop events = %q, want %q", got, wantPrefix)
	}
}

func TestQuestJournalDeletePattern5007E0InvalidInteriorExposesHeadLoad(t *testing.T) {
	w := newQuestJournalDeletePatternTestWorld5007E0("a*b")
	defer func() {
		if recover() == nil {
			t.Fatal("invalid qualified interior wildcard did not fault")
		}
		want := []string{`qualify:"a*b"="a*b"`, `load-head=0`}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	}()
	questJournalDeletePatternContract5007E0("a*b", w.hooks())
}

func TestQuestJournalDeletePattern5007E0FaultPrefixes(t *testing.T) {
	builds := map[string]func() (string, *questJournalDeletePatternTestWorld5007E0){
		"exact": func() (string, *questJournalDeletePatternTestWorld5007E0) {
			w := newQuestJournalDeletePatternTestWorld5007E0("Map:Quest")
			w.exact = questJournalDeletePatternEntryA5007E0
			return "Quest", w
		},
		"all": func() (string, *questJournalDeletePatternTestWorld5007E0) {
			w := newQuestJournalDeletePatternTestWorld5007E0("*:*")
			w.head = questJournalDeletePatternEntryA5007E0
			return "*: *", w
		},
		"prefix": func() (string, *questJournalDeletePatternTestWorld5007E0) {
			w := newQuestJournalDeletePatternTestWorld5007E0("Map:*")
			w.head = questJournalDeletePatternEntryA5007E0
			w.names[w.head] = "map:Quest"
			return "Map:*", w
		},
		"suffix": func() (string, *questJournalDeletePatternTestWorld5007E0) {
			w := newQuestJournalDeletePatternTestWorld5007E0("*:Quest")
			w.head = questJournalDeletePatternEntryA5007E0
			w.names[w.head] = "Map:Quest"
			return "*:Quest", w
		},
		"interior": func() (string, *questJournalDeletePatternTestWorld5007E0) {
			w := newQuestJournalDeletePatternTestWorld5007E0("Map:*Quest")
			w.head = questJournalDeletePatternEntryA5007E0
			w.names[w.head] = "Map:xQuest"
			return "Map:*Quest", w
		},
	}
	for name, build := range builds {
		t.Run(name, func(t *testing.T) {
			pattern, baseline := build()
			questJournalDeletePatternContract5007E0(pattern, baseline.hooks())
			want := append([]string(nil), baseline.events...)

			for faultAt := 1; faultAt <= len(want); faultAt++ {
				t.Run(fmt.Sprintf("event-%d", faultAt), func(t *testing.T) {
					pattern, w := build()
					w.faultAt = faultAt
					var recovered any
					func() {
						defer func() { recovered = recover() }()
						questJournalDeletePatternContract5007E0(pattern, w.hooks())
					}()
					if recovered == nil {
						t.Fatal("fault sentinel was not recovered")
					}
					if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
						t.Fatalf("events = %q, want fault prefix %q", w.events, prefix)
					}
				})
			}
		})
	}
}
