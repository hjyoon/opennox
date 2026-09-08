package legacy

import "strings"

// questJournalDeletePatternHooks5007E0 exposes every externally observable
// operation in GAME.EXE 005007E0. Entry is a generic comparable token so the
// semantic contract cannot inherit PE32 pointer width.
type questJournalDeletePatternHooks5007E0[Entry comparable] struct {
	qualify                 func(string) string
	findExact               func(string) Entry
	loadHead                func() Entry
	loadNext                func(Entry) Entry
	equalFoldPrefix         func(Entry, string, int) bool
	firstSubstringRemaining func(Entry, int, string) (int, bool)
	deleteEntry             func(Entry)
}

// questJournalASCIIPrefixEqualFold5007E0 models the C-locale byte comparison
// used by nox_strnicmp. It intentionally does not apply Unicode case folding.
func questJournalASCIIPrefixEqualFold5007E0(left, right string, count int) bool {
	if count < 0 || len(left) < count || len(right) < count {
		return false
	}
	for i := 0; i < count; i++ {
		lc, rc := left[i], right[i]
		if 'A' <= lc && lc <= 'Z' {
			lc += 'a' - 'A'
		}
		if 'A' <= rc && rc <= 'Z' {
			rc += 'a' - 'A'
		}
		if lc != rc {
			return false
		}
	}
	return true
}

// questJournalFirstSubstringRemaining5007E0 returns the strlen value at the
// first strstr result. A later occurrence that reaches the end is irrelevant
// when an earlier occurrence exists.
func questJournalFirstSubstringRemaining5007E0(haystack, needle string) (int, bool) {
	at := strings.Index(haystack, needle)
	if at < 0 {
		return 0, false
	}
	return len(haystack) - at, true
}

// questJournalDeletePatternContract5007E0 preserves GAME.EXE 005007E0's
// branch and access order for valid journal C strings. Qualification and the
// first '*' scan happen before every operation. The no-wildcard path performs
// a second qualified lookup using the original argument and removes at most
// one entry. Wildcard walks save Next before comparing or deleting.
//
// Prefix comparisons are bytewise case-insensitive. Suffix comparisons use
// the first case-sensitive strstr result and require strlen(result) to equal
// the suffix length. The interior branch deliberately derives both the entry
// search offset and suffix from colon+2, rather than from star+1; it also does
// not validate that the first '*' follows the colon.
func questJournalDeletePatternContract5007E0[Entry comparable](
	pattern string,
	h questJournalDeletePatternHooks5007E0[Entry],
) {
	qualified := h.qualify(pattern)
	star := strings.IndexByte(qualified, '*')
	var nilEntry Entry
	if star < 0 {
		entry := h.findExact(pattern)
		if entry != nilEntry {
			h.deleteEntry(entry)
		}
		return
	}

	if qualified == "*:*" {
		for entry := h.loadHead(); entry != nilEntry; {
			next := h.loadNext(entry)
			h.deleteEntry(entry)
			entry = next
		}
		return
	}

	if star == len(qualified)-1 {
		entry := h.loadHead()
		prefixLen := len(qualified) - 1
		for entry != nilEntry {
			next := h.loadNext(entry)
			if h.equalFoldPrefix(entry, qualified, prefixLen) {
				h.deleteEntry(entry)
			}
			entry = next
		}
		return
	}

	if star == 0 {
		entry := h.loadHead()
		suffix := qualified[1:]
		for entry != nilEntry {
			next := h.loadNext(entry)
			remaining, found := h.firstSubstringRemaining(entry, 0, suffix)
			if found && remaining == len(suffix) {
				h.deleteEntry(entry)
			}
			entry = next
		}
		return
	}

	colon := strings.IndexByte(qualified, ':')
	entry := h.loadHead()
	// sub_5009B0 guarantees a colon for valid inputs. If a test hook violates
	// that invariant, expose the original invalid interior path after the head
	// load instead of silently treating it as a non-match.
	if colon < 0 {
		panic("quest journal wildcard pattern has no qualified colon")
	}
	suffixStart := colon + 2
	if suffixStart > len(qualified) {
		panic("quest journal wildcard suffix starts past its terminator")
	}
	suffix := qualified[suffixStart:]
	prefixLen := colon + 1
	for entry != nilEntry {
		next := h.loadNext(entry)
		if h.equalFoldPrefix(entry, qualified, prefixLen) {
			remaining, found := h.firstSubstringRemaining(entry, suffixStart, suffix)
			if found && remaining == len(suffix) {
				h.deleteEntry(entry)
			}
		}
		entry = next
	}
}
