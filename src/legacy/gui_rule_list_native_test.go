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
}

func TestRuleListFilename4CF060KeepsNarrowEncoding(t *testing.T) {
	const suffix = 191640
	bytes := memmap.Slice(0x587000, suffix)[:5]
	old := [5]byte(bytes)
	t.Cleanup(func() { copy(bytes, old[:]) })
	copy(bytes, ".rul\x00")
	got := ruleListFilename(internWStr("café"), suffix)
	defer StrFree(got)
	if s := alloc.GoString((*byte)(unsafe.Pointer(got))); s != "caf\xe9.rul" {
		t.Fatalf("legacy filename = %q", s)
	}
}
