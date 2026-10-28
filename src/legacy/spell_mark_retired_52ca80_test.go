package legacy

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestMarkCastRetiredCEntry52CA80(t *testing.T) {
	raw, err := os.ReadFile("GAME4_2.c")
	if err != nil {
		t.Fatal(err)
	}
	_, section, found := strings.Cut(string(raw), "//----- (0052CA80) --------------------------------------------------------")
	if !found {
		t.Fatal("original Mark provenance is missing")
	}
	section, _, found = strings.Cut(section, "//----- (0052CBD0) --------------------------------------------------------")
	if !found || !strings.Contains(section, "#if 0\nint sub_52CA80(") {
		t.Fatal("unsafe integer-pointer Mark entry is still enabled")
	}
	_, body, found := strings.Cut(section, "#if 0\n")
	if !found {
		t.Fatal("disabled Mark provenance start is missing")
	}
	body, _, found = strings.Cut(body, "\n#endif")
	if !found || fmt.Sprintf("%x", sha256.Sum256([]byte(strings.TrimSpace(body)))) != "b5e28853aff8dfba8a8b3084e855cb853d26c53cd7122d7c14e2872a4487f482" {
		t.Fatal("retained original Mark C body changed")
	}
	assertRetiredSpellHeaderPrototypes(t)
	entry, err := os.ReadFile("spells.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(entry), "C.sub_52CA80") {
		t.Fatal("public Mark selector still binds the retired C function")
	}
}
