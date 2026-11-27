package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestRuleListFileTitles4CED40KeepsMapExclusions(t *testing.T) {
	for _, row := range []struct{ file, want string }{
		{"Arena.RUL", ""},
		{"User.rul", ""},
		{".rul", ""},
		{"tour.map", ""},
		{"tour.RuL", "tour"},
		{"cafe\xe9.rul", "cafe\xe9"},
	} {
		if got := listedRuleName("arena", row.file); got != row.want {
			t.Errorf("listed rule %q as %q, want %q", row.file, got, row.want)
		}
	}
	var entry WIN32_FIND_DATAA
	copy(entry.cFileName[:], "Long tour name.rul")
	copy(entry.cAlternateFileName[:], "LONGTO~1.RUL")
	if got := ruleListEntryName(&entry); got != "LONGTO~1.RUL" {
		t.Fatalf("Windows rule-list entry = %q", got)
	}
	clear(entry.cAlternateFileName[:])
	if got := ruleListEntryName(&entry); got != "Long tour name.rul" {
		t.Fatalf("rule-list entry without a short alias = %q", got)
	}
}

func TestRuleListFilename4CF060KeepsNarrowEncoding(t *testing.T) {
	const suffix = 191640
	bytes := memmap.Slice(0x587000, suffix)[:5]
	old := [5]byte(bytes)
	t.Cleanup(func() { copy(bytes, old[:]) })
	copy(bytes, ".rul\x00")
	got, free := ruleListFilename(internWStr("café"), suffix)
	defer free()
	if s := alloc.GoString((*byte)(unsafe.Pointer(got))); s != "caf\xe9.rul" {
		t.Fatalf("legacy filename = %q", s)
	}
}
