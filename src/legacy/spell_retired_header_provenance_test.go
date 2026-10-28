package legacy

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

var retiredSpellHeaderPrototypes = []string{
	"int sub_52CA80(int a1, int a2, int a3, int a4);",
	"int sub_52CBD0(int a1, int a2, int a3, int a4);",
	"void sub_52CE60(int a1);",
	"int nox_xxx_castLock_52CE90(int a1, int a2, int a3, int a4);",
	"void sub_52CF90(int a1, int a2);",
	"void sub_52D060(int a1, int a2);",
	"int nox_xxx_castToxicCloud_52DB60(int a1, int a2, int a3, int a4, int a5);",
}

var retiredHeaderComments = regexp.MustCompile(`(?s)/\*.*?\*/|//[^\r\n]*`)

func retiredSpellHeaderPreserved(header string) bool {
	// Preserve the exact historical declarations under test. Freezing the
	// entire shared header also rejected unrelated native pointer ports.
	lines := strings.Split(retiredHeaderComments.ReplaceAllString(header, ""), "\n")
	for _, prototype := range retiredSpellHeaderPrototypes {
		count := 0
		for _, line := range lines {
			if strings.TrimSpace(line) == prototype {
				count++
			}
		}
		if count != 1 {
			return false
		}
	}
	return true
}

func assertRetiredSpellHeaderPrototypes(t *testing.T) {
	t.Helper()
	header, err := os.ReadFile("GAME4_2.h")
	if err != nil {
		t.Fatal(err)
	}
	if !retiredSpellHeaderPreserved(string(header)) {
		t.Fatal("retired original Mark/Lock/ToxicCloud declarations changed")
	}
}

func TestRetiredSpellHeaderRejectsChangedMissingDuplicateAndCommentedDeclarations(t *testing.T) {
	header := strings.Join(retiredSpellHeaderPrototypes, "\n") + "\n"
	if !retiredSpellHeaderPreserved(header + "int nox_xxx_onStartLightning_52F820(void* record);\n") {
		t.Fatal("an unrelated native declaration must not invalidate retired provenance")
	}
	for _, prototype := range retiredSpellHeaderPrototypes {
		for _, replacement := range []string{
			"", prototype + "\n" + prototype,
			strings.Replace(prototype, "int", "int32_t", 1),
			"/* " + prototype + " */",
		} {
			if retiredSpellHeaderPreserved(strings.Replace(header, prototype, replacement, 1)) {
				t.Errorf("modified retired declaration was accepted: %q -> %q", prototype, replacement)
			}
		}
	}
}
